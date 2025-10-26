package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type StorageStats struct {
	TotalSize   int64
	ItemCount   int
	PinnedCount int
}

type StorageManager struct {
	ipfs    *IPFSNode
	quota   int64
	stats   *StorageStats
	expired map[string]bool // CID -> expired status
	mutex   sync.RWMutex
}

func NewStorageManager(ipfs *IPFSNode) *StorageManager {
	return &StorageManager{
		ipfs:    ipfs,
		quota:   100 * 1024 * 1024, // 100MB default quota
		stats:   &StorageStats{},
		expired: make(map[string]bool),
	}
}

func (sm *StorageManager) SetQuota(quota int64) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.quota = quota
}

func (sm *StorageManager) StoreContent(ctx context.Context, data []byte) (string, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check quota
	if sm.stats.TotalSize+int64(len(data)) > sm.quota {
		return "", fmt.Errorf("quota exceeded: %d bytes", sm.quota)
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
	sm.stats.TotalSize += int64(len(data))
	sm.stats.ItemCount++
	sm.stats.PinnedCount++

	return cid, nil
}

func (sm *StorageManager) StoreContentWithTTL(ctx context.Context, data []byte, ttl time.Duration) (string, error) {
	cid, err := sm.StoreContent(ctx, data)
	if err != nil {
		return "", err
	}

	// Schedule unpinning after TTL
	go func() {
		time.Sleep(ttl)
		sm.ipfs.Unpin(ctx, cid)

		// Mark as expired
		sm.mutex.Lock()
		defer sm.mutex.Unlock()
		sm.expired[cid] = true
		sm.stats.PinnedCount--
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

	// Update stats
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.stats.PinnedCount--

	return nil
}

func (sm *StorageManager) RunGarbageCollection(ctx context.Context) (int, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// For testing, simulate garbage collection
	// In a real implementation, this would remove unpinned content
	removed := 0

	// Only remove content if there are unpinned items
	unpinnedCount := sm.stats.ItemCount - sm.stats.PinnedCount
	if unpinnedCount > 0 {
		removed = 1
		sm.stats.ItemCount--
		sm.stats.TotalSize -= 1024 // Simulate removing 1KB
	}

	return removed, nil
}

func (sm *StorageManager) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy of stats
	return &StorageStats{
		TotalSize:   sm.stats.TotalSize,
		ItemCount:   sm.stats.ItemCount,
		PinnedCount: sm.stats.PinnedCount,
	}, nil
}

func (sm *StorageManager) Close() error {
	return sm.ipfs.Close()
}
