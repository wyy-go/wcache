package main

import (
	"github.com/gin-gonic/gin"
	"github.com/wyy-go/wcache"
	"github.com/wyy-go/wcache/persist/memory"
	"time"
)

func main() {
	app := gin.New()

	app.GET("/hello/:a/:b", custom())
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}

func custom() gin.HandlerFunc {
	f := wcache.CacheByRequestURI(
		wcache.WithCacheStore(memory.NewMemoryStore(time.Minute)),
		wcache.WithExpire(5*time.Second),
		wcache.WithGenerateCacheKey(func(c *gin.Context) (string, bool) {
			return c.GetString("cache_key"), true
		}),
		wcache.WithHandle(func(c *gin.Context) {
			c.String(200, "hello world")
		}),
	)

	return func(c *gin.Context) {
		a := c.Param("a")
		b := c.Param("b")
		c.Set("cache_key", wcache.CacheKeyWithPrefix(wcache.PageCachePrefix, a+":"+b))
		f(c)
	}
}
