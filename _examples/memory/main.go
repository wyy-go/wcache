package main

import (
	"github.com/wyy-go/wcache/persist/memory"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wyy-go/wcache"
)

func main() {
	app := gin.New()

	memoryStore := memory.NewMemoryStore(1 * time.Minute)

	app.GET("/hello",
		wcache.CacheByRequestURI(
			wcache.WithCacheStore(memoryStore),
			wcache.WithExpire(2*time.Second),
			wcache.WithHandle(func(c *gin.Context) {
				c.String(200, "hello world")
			})),
	)

	if err := app.Run(":8080"); err != nil {
		panic(err)
	}
}
