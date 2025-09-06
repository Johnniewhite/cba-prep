package cache

import "errors"

var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheInvalidType = errors.New("invalid cache type")
	ErrCacheConnectionFailed = errors.New("cache connection failed")
)