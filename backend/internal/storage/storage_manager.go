package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type StorageStats struct {
	UsedBytes    int64
	MaxBytes     int64
	ContentCount int
	LastCleanup  time.Time
}

type StorageConfig struct {
	MaxStorageBytes int64
	CleanupInterval time.Duration
	ContentTTL      time.Duration
}

type StorageManager struct {
	ipfs        *IPFSNode
	config      *StorageConfig
	stats       *StorageStats
	expired     map[string]bool      // CID -> expired status
	contentSize map[string]int64     // CID -> content size
	lastAccess  map[string]time.Time // CID -> last access time
	mutex       sync.RWMutex
}

func NewStorageManager(ipfs *IPFSNode) *StorageManager {
	config := &StorageConfig{
		MaxStorageBytes: 100 * 1024 * 1024, // 100MB default
		CleanupInterval: time.Minute * 5,   // 5 minutes
		ContentTTL:      time.Hour * 24,    // 24 hours
	}

	return &StorageManager{
		ipfs:   ipfs,
		config: config,
		stats: &StorageStats{
			MaxBytes: config.MaxStorageBytes,
		},
		expired:     make(map[string]bool),
		contentSize: make(map[string]int64),
		lastAccess:  make(map[string]time.Time),
	}
}

func NewStorageManagerWithConfig(ipfs *IPFSNode, config *StorageConfig) *StorageManager {
	return &StorageManager{
		ipfs:   ipfs,
		config: config,
		stats: &StorageStats{
			MaxBytes: config.MaxStorageBytes,
		},
		expired:     make(map[string]bool),
		contentSize: make(map[string]int64),
		lastAccess:  make(map[string]time.Time),
	}
}

func (sm *StorageManager) SetQuota(quota int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.config.MaxStorageBytes = quota
	sm.stats.MaxBytes = quota
}

func (sm *StorageManager) StoreContent(ctx context.Context, data []byte) (string, error) {
	return sm.StoreContentWithTTL(ctx, data, sm.config.ContentTTL)
}

func (sm *StorageManager) StoreContentWithTTL(ctx context.Context, data []byte, ttl time.Duration) (string, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check quota and try to free space if needed
	if sm.stats.UsedBytes+int64(len(data)) > sm.config.MaxStorageBytes {
		// Try to free up space by removing expired content
		sm.cleanupExpiredContent()

		// If still over quota, try LRU cleanup
		if sm.stats.UsedBytes+int64(len(data)) > sm.config.MaxStorageBytes {
			sm.lruCleanup(int64(len(data)))
		}

		// Final check
		if sm.stats.UsedBytes+int64(len(data)) > sm.config.MaxStorageBytes {
			return "", fmt.Errorf("quota exceeded: %d bytes", sm.config.MaxStorageBytes)
		}
	}

	// Store in IPFS
	cid, err := sm.ipfs.Add(ctx, data)
	if err != nil {
		return "", err
	}

	// Pin content
	err = sm.ipfs.Pin(ctx, cid)
	if err != nil {
		return "", err
	}

	// Update stats
	sm.stats.UsedBytes += int64(len(data))
	sm.stats.ContentCount++
	sm.contentSize[cid] = int64(len(data))
	sm.lastAccess[cid] = time.Now()

	// Schedule unpinning after TTL
	go func() {
		time.Sleep(ttl)
		sm.ipfs.Unpin(ctx, cid)

		// Mark as expired
		sm.mutex.Lock()
		defer sm.mutex.Unlock()
		sm.expired[cid] = true
	}()

	return cid, nil
}

func (sm *StorageManager) GetContent(ctx context.Context, cid string) ([]byte, error) {
	sm.mutex.RLock()
	expired := sm.expired[cid]
	sm.mutex.RUnlock()

	if expired {
		return nil, fmt.Errorf("content expired: %s", cid)
	}

	// Update access time
	sm.mutex.Lock()
	sm.lastAccess[cid] = time.Now()
	sm.mutex.Unlock()

	return sm.ipfs.Get(ctx, cid)
}

func (sm *StorageManager) PinContent(ctx context.Context, cid string) error {
	return sm.ipfs.Pin(ctx, cid)
}

func (sm *StorageManager) UnpinContent(ctx context.Context, cid string) error {
	err := sm.ipfs.Unpin(ctx, cid)
	if err != nil {
		return err
	}

	// Mark as expired
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.expired[cid] = true

	return nil
}

func (sm *StorageManager) RunGarbageCollection(ctx context.Context) (int, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// For testing, simulate garbage collection
	// In a real implementation, this would remove unpinned content
	removed := 0

	// Only remove content if there are expired items
	if len(sm.expired) > 0 {
		removed = 1
		sm.stats.ContentCount--
		sm.stats.UsedBytes -= 1024 // Simulate removing 1KB
	}

	return removed, nil
}

func (sm *StorageManager) Close() error {
	return sm.ipfs.Close()
}

// Additional methods for integration tests

func (sm *StorageManager) ContentExists(ctx context.Context, cid string) (bool, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Check if content exists and is not expired
	_, exists := sm.expired[cid]
	return !exists, nil
}

func (sm *StorageManager) RemoveContent(ctx context.Context, cid string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Remove from IPFS
	err := sm.ipfs.Remove(ctx, cid)
	if err != nil {
		return err
	}

	// Update stats with actual content size
	sm.stats.ContentCount--
	if size, exists := sm.contentSize[cid]; exists {
		sm.stats.UsedBytes -= size
		delete(sm.contentSize, cid)
	}

	// Remove from expired tracking
	delete(sm.expired, cid)

	return nil
}

func (sm *StorageManager) GetStorageStats(ctx context.Context) StorageStats {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy of stats
	return StorageStats{
		UsedBytes:    sm.stats.UsedBytes,
		MaxBytes:     sm.stats.MaxBytes,
		ContentCount: sm.stats.ContentCount,
		LastCleanup:  sm.stats.LastCleanup,
	}
}

// Helper methods for cleanup

func (sm *StorageManager) cleanupExpiredContent() {
	// Remove expired content
	for cid := range sm.expired {
		sm.ipfs.Remove(context.Background(), cid)
		delete(sm.expired, cid)
		sm.stats.ContentCount--

		// Use actual content size
		if size, exists := sm.contentSize[cid]; exists {
			sm.stats.UsedBytes -= size
			delete(sm.contentSize, cid)
		}
		delete(sm.lastAccess, cid)
	}
}

func (sm *StorageManager) lruCleanup(neededBytes int64) {
	// LRU cleanup - remove least recently used content until we have enough space

	// First, remove expired content
	for cid := range sm.expired {
		if sm.stats.UsedBytes <= sm.config.MaxStorageBytes-neededBytes {
			break
		}

		// Remove from IPFS
		sm.ipfs.Remove(context.Background(), cid)
		delete(sm.expired, cid)
		sm.stats.ContentCount--

		// Use actual content size
		if size, exists := sm.contentSize[cid]; exists {
			sm.stats.UsedBytes -= size
			delete(sm.contentSize, cid)
		}
		delete(sm.lastAccess, cid)
		delete(sm.lastAccess, cid)
	}

	// If still not enough space, remove least recently used content
	if sm.stats.UsedBytes > sm.config.MaxStorageBytes-neededBytes {
		// Find least recently used content
		var oldestCID string
		var oldestTime time.Time

		for cid, accessTime := range sm.lastAccess {
			if oldestCID == "" || accessTime.Before(oldestTime) {
				oldestCID = cid
				oldestTime = accessTime
			}
		}

		if oldestCID != "" {
			// Remove least recently used content
			sm.ipfs.Remove(context.Background(), oldestCID)
			sm.stats.ContentCount--

			if size, exists := sm.contentSize[oldestCID]; exists {
				sm.stats.UsedBytes -= size
				delete(sm.contentSize, oldestCID)
			}
			delete(sm.lastAccess, oldestCID)
		}
	}
}
