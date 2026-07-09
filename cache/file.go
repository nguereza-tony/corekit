package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nguereza-tony/corekit/config"
)

type fileItem struct {
	Value      string    `json:"value"`
	Expiration time.Time `json:"expiration"`
	IsCounter  bool      `json:"is_counter"`
}

type FileCache struct {
	dir             string
	mu              sync.RWMutex
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewFileCache creates a new file-based cache
func NewFileCache(cfg *config.CacheConfig) (*FileCache, error) {
	if err := os.MkdirAll(cfg.File.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	fc := &FileCache{
		dir:             cfg.File.Path,
		cleanupInterval: time.Duration(cfg.File.CleanupInterval) * time.Second,
		stopCleanup:     make(chan struct{}),
	}

	go fc.cleanupLoop()
	return fc, nil
}

// Get retrieves a value from cache
func (fc *FileCache) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	fc.mu.RLock()
	defer fc.mu.RUnlock()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	item, err := fc.readItem(key)
	if err != nil {
		return "", err
	}

	if item.IsCounter {
		return "", errors.New("key is a counter, use Increment/Decrement")
	}

	return item.Value, nil
}

// Set stores a value in cache with expiration
func (fc *FileCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	expiration := time.Time{}
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	item := &fileItem{
		Value:      value,
		Expiration: expiration,
		IsCounter:  false,
	}

	return fc.writeItem(key, item)
}

// Delete removes a key from cache
func (fc *FileCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := fc.getFilePath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Exists checks if a key exists in cache
func (fc *FileCache) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	fc.mu.RLock()
	defer fc.mu.RUnlock()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	_, err := fc.readItem(key)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Expire sets expiration on an existing key
func (fc *FileCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	item, err := fc.readItem(key)
	if err != nil {
		return err
	}

	if ttl > 0 {
		item.Expiration = time.Now().Add(ttl)
	} else {
		item.Expiration = time.Time{}
	}

	return fc.writeItem(key, item)
}

// Increment increments a counter
func (fc *FileCache) Increment(ctx context.Context, key string) (int64, error) {
	return fc.IncrementBy(ctx, key, 1)
}

// IncrementBy increments a counter by a specific amount
func (fc *FileCache) IncrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	if key == "" {
		return 0, ErrInvalidKey
	}

	fc.mu.Lock()
	defer fc.mu.Unlock()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	item, err := fc.readItem(key)
	if err != nil && !errors.Is(err, ErrKeyNotFound) {
		return 0, err
	}

	if item == nil {
		// Create new counter
		item = &fileItem{
			Value:      fmt.Sprintf("%d", delta),
			Expiration: time.Time{},
			IsCounter:  true,
		}
	} else if !item.IsCounter {
		return 0, errors.New("key exists but is not a counter")
	} else {
		// Parse current value
		var current int64
		if _, err := fmt.Sscan(item.Value, &current); err != nil {
			return 0, fmt.Errorf("invalid counter value: %w", err)
		}
		current += delta
		item.Value = fmt.Sprintf("%d", current)
	}

	if err := fc.writeItem(key, item); err != nil {
		return 0, err
	}

	var result int64
	fmt.Sscan(item.Value, &result)
	return result, nil
}

// Decrement decrements a counter
func (fc *FileCache) Decrement(ctx context.Context, key string) (int64, error) {
	return fc.DecrementBy(ctx, key, 1)
}

// DecrementBy decrements a counter by a specific amount
func (fc *FileCache) DecrementBy(ctx context.Context, key string, delta int64) (int64, error) {
	return fc.IncrementBy(ctx, key, -delta)
}

// Clear removes all keys from cache
func (fc *FileCache) Clear(ctx context.Context) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	entries, err := os.ReadDir(fc.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			if err := os.Remove(filepath.Join(fc.dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// Ping checks if the cache is reachable
func (fc *FileCache) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if directory exists and is writable
	info, err := os.Stat(fc.dir)
	if err != nil {
		return fmt.Errorf("cache directory not accessible: %w", err)
	}
	if !info.IsDir() {
		return errors.New("cache path is not a directory")
	}

	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(fc.dir, ".ping_test")
	if err := os.WriteFile(testFile, []byte("ping"), 0644); err != nil {
		return fmt.Errorf("cache directory not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}

// Close closes the cache connection
func (fc *FileCache) Close() error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if fc.stopCleanup != nil {
		close(fc.stopCleanup)
		fc.stopCleanup = nil
	}
	return nil
}

func (fc *FileCache) getFilePath(key string) string {
	// Sanitize key to prevent directory traversal
	safeKey := filepath.Base(key)
	return filepath.Join(fc.dir, safeKey+".json")
}

func (fc *FileCache) readItem(key string) (*fileItem, error) {
	filePath := fc.getFilePath(key)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}

	// Check if expired
	if !item.Expiration.IsZero() && time.Now().After(item.Expiration) {
		os.Remove(filePath)
		return nil, ErrKeyNotFound
	}

	return &item, nil
}

func (fc *FileCache) writeItem(key string, item *fileItem) error {
	filePath := fc.getFilePath(key)
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// cleanupLoop periodically removes expired items
func (fc *FileCache) cleanupLoop() {
	ticker := time.NewTicker(fc.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fc.cleanup()
		case <-fc.stopCleanup:
			return
		}
	}
}

func (fc *FileCache) cleanup() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	entries, err := os.ReadDir(fc.dir)
	if err != nil {
		return
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(fc.dir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var item fileItem
		if err := json.Unmarshal(data, &item); err != nil {
			continue
		}

		if !item.Expiration.IsZero() && now.After(item.Expiration) {
			os.Remove(filePath)
		}
	}
}
