# wcache

![GitHub Repo stars](https://img.shields.io/github/stars/wyy-go/wcache?style=social)
![](https://img.shields.io/badge/license-MIT-green)
![GitHub](https://img.shields.io/github/license/wyy-go/wcache)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/wyy-go/wcache)
![GitHub CI Status](https://img.shields.io/github/workflow/status/wyy-go/wcache/ci?label=CI)
[![Go Report Card](https://goreportcard.com/badge/github.com/wyy-go/wcache)](https://goreportcard.com/report/github.com/wyy-go/wcache)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/wyy-go/wcache?tab=doc)
[![codecov](https://codecov.io/gh/wyy-go/wcache/branch/main/graph/badge.svg)](https://codecov.io/gh/wyy-go/wcache)



A high performance gin middleware to cache http response. Compared to gin cache. It has a huge performance improvement.


# Feature

* Has a huge performance improvement compared to gin-contrib/cache.
* Cache http response in local memory or Redis.
* Offer a way to custom the cache strategy by per request.
* Use singleflight to avoid cache breakdown problem.

# How To Use

## Install
```
go get -u github.com/wyy-go/wcache
```

## Example

### Cache In Local Memory

```go
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
```

### Cache In Redis

```go
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
```



# Benchmark

```
wrk -c 500 -d 1m -t 5 http://127.0.0.1:8080/hello
```

## MemoryStore

![MemoryStore QPS](https://www.cyhone.com/img/gin-cache/memory_cache_qps.png)

## RedisStore

![RedisStore QPS](https://www.cyhone.com/img/gin-cache/redis_cache_qps.png)
