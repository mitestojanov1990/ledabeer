package storage

import (
	"context"
	"sync"
	"time"
)

type LRUConfig struct {
	MaxSize         int
	TTL             time.Duration
	CleanupInterval time.Duration
}

type LRUStats struct {
	Hits      int
	Misses    int
	Size      int
	MaxSize   int
	Evictions int
}

type CacheEntry struct {
	Data        []byte
	Timestamp   time.Time
	AccessCount int
}

type LRUCache struct {
	config      *LRUConfig
	entries     map[string]*CacheEntry
	accessOrder []string // Most recently used first
	stats       *LRUStats
	mutex       sync.RWMutex
	stopChan    chan struct{}
}

func NewLRUCache(config *LRUConfig) *LRUCache {
	cache := &LRUCache{
		config:      config,
		entries:     make(map[string]*CacheEntry),
		accessOrder: make([]string, 0),
		stats:       &LRUStats{MaxSize: config.MaxSize},
		stopChan:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

func (c *LRUCache) Get(ctx context.Context, cid string) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if entry exists and is not expired
	entry, exists := c.entries[cid]
	if exists && !c.isExpired(entry) {
		// Update access order
		c.updateAccessOrder(cid)
		entry.AccessCount++
		c.stats.Hits++
		return entry.Data, nil
	}

	// Cache miss - remove expired entry if exists
	if exists {
		delete(c.entries, cid)
		c.removeFromAccessOrder(cid)
	}

	c.stats.Misses++
	return nil, nil // Cache miss - caller should fetch from IPFS
}

func (c *LRUCache) Set(ctx context.Context, cid string, data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict
	if len(c.entries) >= c.config.MaxSize && c.entries[cid] == nil {
		c.evictLRU()
	}

	// Add or update entry
	c.entries[cid] = &CacheEntry{
		Data:        data,
		Timestamp:   time.Now(),
		AccessCount: 1,
	}

	// Update access order
	c.updateAccessOrder(cid)

	return nil
}

func (c *LRUCache) updateAccessOrder(cid string) {
	// Remove from current position
	c.removeFromAccessOrder(cid)

	// Add to front (most recently used)
	c.accessOrder = append([]string{cid}, c.accessOrder...)
}

func (c *LRUCache) removeFromAccessOrder(cid string) {
	for i, id := range c.accessOrder {
		if id == cid {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			break
		}
	}
}

func (c *LRUCache) evictLRU() {
	if len(c.accessOrder) == 0 {
		return
	}

	// Remove least recently used (last in access order)
	lruCid := c.accessOrder[len(c.accessOrder)-1]
	delete(c.entries, lruCid)
	c.accessOrder = c.accessOrder[:len(c.accessOrder)-1]
	c.stats.Evictions++
}

func (c *LRUCache) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.Timestamp) > c.config.TTL
}

func (c *LRUCache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

func (c *LRUCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	// Find expired entries
	for cid, entry := range c.entries {
		if now.Sub(entry.Timestamp) > c.config.TTL {
			toRemove = append(toRemove, cid)
		}
	}

	// Remove expired entries
	for _, cid := range toRemove {
		delete(c.entries, cid)
		c.removeFromAccessOrder(cid)
	}
}

func (c *LRUCache) GetStats() LRUStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return LRUStats{
		Hits:      c.stats.Hits,
		Misses:    c.stats.Misses,
		Size:      len(c.entries),
		MaxSize:   c.stats.MaxSize,
		Evictions: c.stats.Evictions,
	}
}

func (c *LRUCache) Close() {
	close(c.stopChan)
}
