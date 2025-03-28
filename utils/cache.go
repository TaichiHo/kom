package utils

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"k8s.io/klog/v2"
)

func GetOrSetCache[T any](cache *ristretto.Cache[string, any], cacheKey string, ttl time.Duration, queryFunc func() (T, error)) (T, error) {
	var zero T

	// If TTL parameter is not set, no caching is needed, execute query method directly
	if ttl <= 0 {
		return queryFunc()
	}
	// Check if cache hit
	if v, found := cache.Get(cacheKey); found {
		klog.V(5).Infof("cache hit cacheKey= %s", cacheKey)
		return v.(T), nil
	}

	// Cache miss, execute query method
	result, err := queryFunc()
	if err != nil {
		return zero, err
	}

	// Set cache and return result
	cache.SetWithTTL(cacheKey, result, 100, ttl)
	cache.Wait()

	return result, nil
}
