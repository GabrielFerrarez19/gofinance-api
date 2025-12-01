package cache

type CacheMiddleware struct {
	cache *CacheService
}

func NewCacheMiddleware(cache *CacheService) *CacheMiddleware {
	return &CacheMiddleware{
		cache: cache,
	}
}
