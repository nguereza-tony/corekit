package cache

import "errors"

var (
	// Not found errors
	ErrKeyNotFound = errors.New("key not found")

	// Validation errors
	ErrInvalidTTL = errors.New("invalid TTL")
	ErrInvalidKey = errors.New("invalid key")
	ErrCacheFull  = errors.New("cache is full")

	// Operation errors
	ErrConnectionFailed = errors.New("cache connection failed")
	ErrOperationFailed  = errors.New("cache operation failed")
)
