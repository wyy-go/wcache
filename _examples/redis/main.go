package main

import (
	redisStore "github.com/wyy-go/wcache/persist/redis"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/wyy-go/wcache"
)

func main() {
	app := gin.New()

	store := redisStore.NewRedisStore(redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
		DB:      0,
	}))

	app.GET("/hello",
		wcache.CacheByRequestURI(
			wcache.WithCacheStore(store),
			wcache.WithExpire(2*time.Second),
			wcache.WithHandle(func(c *gin.Context) {
				c.String(200, "hello world")
			})),
	)
	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
