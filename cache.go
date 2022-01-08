package wcache

import (
	"crypto/sha1"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"
)

// PageCachePrefix default page cache key prefix
var PageCachePrefix = "wcache.page.cache:"

// Cache user must pass getCacheKey to describe the way to generate cache key
func Cache(opts ...Option) gin.HandlerFunc {
	options := &Options{
		logger:                    NewDiscard(),
		hitCacheCallback:          defaultHitCacheCallback,
		shareSingleFlightCallback: defaultShareSingleFlightCallback,
		group:                     new(singleflight.Group),
		store:                     nil,
		expire:                    10 * time.Minute,
		handle:                    defaultHandle,
		generateCacheKey:          GenerateCacheKeyByURI,
		pool:                      NewPool(),
		encode:                    JSONEncoding{},
		rand:                      defaultRand,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.store == nil {
		panic("you must set a cache store!")
	}

	return func(c *gin.Context) {
		cacheKey, shouldCache := options.generateCacheKey(c)
		if !shouldCache {
			options.handle(c)
			return
		}

		// read cache first
		respCache := options.pool.Get()
		defer options.pool.Put(respCache)
		respCache.encode = options.encode

		err := options.store.Get(cacheKey, respCache)
		if err == nil {
			responseWithCache(c, options, respCache)
			options.hitCacheCallback(c)
			return
		} else {
			options.logger.Errorf("get cache error: %s, cache key: %s", err, cacheKey)
		}

		inFlight := false
		rawRespCache, _, shared := options.group.Do(cacheKey, func() (interface{}, error) {
			if options.singleFlightForgetTimeout > 0 {
				forgetTimer := time.AfterFunc(options.singleFlightForgetTimeout, func() {
					options.group.Forget(cacheKey)
				})
				defer forgetTimer.Stop()
			}

			// use responseCacheWriter in order to record the response
			cacheWriter := &responseCacheWriter{ResponseWriter: c.Writer}
			c.Writer = cacheWriter
			options.handle(c)

			inFlight = true
			respCache := getCacheFromWriter(cacheWriter, options.encode)

			// only cache 2xx response
			if !c.IsAborted() && cacheWriter.Status() < 300 && cacheWriter.Status() >= 200 {
				if err := options.store.Set(cacheKey, respCache, options.expire+options.rand()); err != nil {
					options.logger.Errorf("set cache key error: %s, cache key: %s", err, cacheKey)
				}
			}

			return respCache, nil
		})

		if !inFlight && shared {
			responseWithCache(c, options, rawRespCache.(*ResponseCache))
			options.shareSingleFlightCallback(c)
		}
	}
}

// CacheByRequestURI a shortcut function for caching response by uri
func CacheByRequestURI(opts ...Option) gin.HandlerFunc {
	return Cache(opts...)
}

// CacheByRequestPath a shortcut function for caching response by url path, means will discard the query params
func CacheByRequestPath(opts ...Option) gin.HandlerFunc {
	return Cache(append(opts, WithGenerateCacheKey(GenerateCacheKeyByPath))...)
}

func CacheKeyWithPrefix(prefix, key string) string {
	if len(key) > 200 {
		d := sha1.Sum([]byte(key))
		return prefix + string(d[:])
	}
	return prefix + key
}

// GenerateCacheKeyByURI generate key with PageCachePrefix and request uri
func GenerateCacheKeyByURI(c *gin.Context) (string, bool) {
	return CacheKeyWithPrefix(PageCachePrefix, url.QueryEscape(c.Request.RequestURI)), true
}

// GenerateCacheKeyByPath generate key with PageCachePrefix and request Path
func GenerateCacheKeyByPath(c *gin.Context) (string, bool) {
	return CacheKeyWithPrefix(PageCachePrefix, url.QueryEscape(c.Request.URL.Path)), true
}
