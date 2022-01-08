package wcache

import (
	"github.com/wyy-go/wcache/persist"
	"golang.org/x/sync/singleflight"
	"time"

	"github.com/gin-gonic/gin"
)

// Options contains all options
type Options struct {
	logger                    Logger
	group                     *singleflight.Group
	store                     persist.CacheStore
	expire                    time.Duration
	generateCacheKey          GenerateCacheKey
	hitCacheCallback          OnHitCacheCallback
	singleFlightForgetTimeout time.Duration
	shareSingleFlightCallback OnShareSingleFlightCallback
	handle                    gin.HandlerFunc
	pool                      Pool
	encode                    Encoding
	rand                      Rand
}

// Option represents the optional function.
type Option func(c *Options)

// OnHitCacheCallback define the callback when use cache
type OnHitCacheCallback func(c *gin.Context)

// OnShareSingleFlightCallback define the callback when share the singleflight result
type OnShareSingleFlightCallback func(c *gin.Context)

type GenerateCacheKey func(c *gin.Context) (string, bool)
type Rand func() time.Duration

var defaultRand = func() time.Duration { return 0 }
var defaultHitCacheCallback = func(c *gin.Context) {}
var defaultHandle = func(c *gin.Context) {}
var defaultShareSingleFlightCallback = func(c *gin.Context) {}

// WithLogger set the custom logger
func WithLogger(l Logger) Option {
	return func(c *Options) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithOnHitCache will be called when cache hit.
func WithOnHitCache(cb OnHitCacheCallback) Option {
	return func(c *Options) {
		if cb != nil {
			c.hitCacheCallback = cb
		}
	}
}

// WithOnShareSingleFlight will be called when share the singleflight result
func WithOnShareSingleFlight(cb OnShareSingleFlightCallback) Option {
	return func(c *Options) {
		if cb != nil {
			c.shareSingleFlightCallback = cb
		}
	}
}

// WithSingleFlightForgetTimeout to reduce the impact of long tail requests. when request in the singleflight,
// after the forget timeout, singleflight.Forget will be called
func WithSingleFlightForgetTimeout(forgetTimeout time.Duration) Option {
	return func(c *Options) {
		if forgetTimeout > 0 {
			c.singleFlightForgetTimeout = forgetTimeout
		}
	}
}

func WithSingleflight(group *singleflight.Group) Option {
	return func(c *Options) {
		if group != nil {
			c.group = group
		}
	}
}

func WithCacheStore(store persist.CacheStore) Option {
	return func(c *Options) {
		if store != nil {
			c.store = store
		}
	}
}

func WithExpire(expire time.Duration) Option {
	return func(c *Options) {
		c.expire = expire
	}
}

func WithHandle(handle gin.HandlerFunc) Option {
	return func(c *Options) {
		if handle != nil {
			c.handle = handle
		}
	}
}

func WithGenerateCacheKey(f GenerateCacheKey) Option {
	return func(c *Options) {
		if f != nil {
			c.generateCacheKey = f
		}
	}
}

func WithEncoding(encode Encoding) Option {
	return func(c *Options) {
		if encode != nil {
			c.encode = encode
		}
	}
}

func WithPool(pool Pool) Option {
	return func(c *Options) {
		if pool != nil {
			c.pool = pool
		}
	}
}

func WithRand(rand Rand) Option {
	return func(c *Options) {
		if rand != nil {
			c.rand = rand
		}
	}
}
