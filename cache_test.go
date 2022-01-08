package wcache

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"
	"github.com/wyy-go/wcache/persist"
	"github.com/wyy-go/wcache/persist/memory"
	redisStore "github.com/wyy-go/wcache/persist/redis"
	"golang.org/x/sync/singleflight"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sony/sonyflake"
	"github.com/stretchr/testify/assert"
)

var (
	sf          *sonyflake.Sonyflake
	longUri     = "/" + strings.Repeat("1234567890123456789012345", 8)
	redisHost   = "127.0.0.1:6379"
	password    = ""
	enableRedis = false
)

func init() {
	var st sonyflake.Settings
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}

	gin.SetMode(gin.TestMode)
}

var newStore = func(defaultExpiration time.Duration) persist.CacheStore {
	if enableRedis {

		client := redis.NewClient(&redis.Options{
			Addr:     redisHost,
			DB:       10,
			Password: password,
		})

		if err := client.Ping(context.Background()).Err(); err != nil {
			panic(err)
		}

		return redisStore.NewRedisStore(client)
	}

	return memory.NewMemoryStore(defaultExpiration)
}

func performRequest(target string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func generateID() string {
	id, _ := sf.NextID()
	return fmt.Sprint(id)
}

func TestCache(t *testing.T) {
	store := newStore(time.Second * 60)
	r := gin.New()
	r.GET("/cache/ping",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
		),
	)

	w1 := performRequest("/cache/ping", r)
	w2 := performRequest("/cache/ping", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCacheNoNeedCache(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache/ping",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
			WithGenerateCacheKey(func(c *gin.Context) (string, bool) {
				return "", false
			}),
			WithPool(NewPool()),
			WithRand(func() time.Duration {
				return time.Duration(rand.Intn(5)) * time.Second
			}),
			WithSingleflight(&singleflight.Group{}),
			WithLogger(NewDiscard()),
			WithEncoding(JSONEncoding{}),
		),
	)

	w1 := performRequest("/cache/ping", r)
	w2 := performRequest("/cache/ping", r)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCacheExpire(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache/ping",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
		),
	)

	w1 := performRequest("/cache/ping", r)
	time.Sleep(time.Second * 3)
	w2 := performRequest("/cache/ping", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCacheHtmlFile(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.LoadHTMLFiles("testdata/template.html")

	r.GET("/cache/html",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.HTML(http.StatusOK, "template.html", gin.H{"value": generateID()})
			}),
		),
	)

	w1 := performRequest("/cache/html", r)
	w2 := performRequest("/cache/html", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCacheHtmlFileExpire(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.LoadHTMLFiles("testdata/template.html")
	r.GET("/cache/html",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second),
			WithHandle(func(c *gin.Context) {
				c.HTML(http.StatusOK, "template.html", gin.H{"value": generateID()})
			}),
		),
	)

	w1 := performRequest("/cache/html", r)
	time.Sleep(time.Second * 3)
	w2 := performRequest("/cache/html", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCacheAborted(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache/aborted",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.AbortWithStatusJSON(http.StatusOK, map[string]string{"time": generateID()})
			}),
		),
	)

	w1 := performRequest("/cache/aborted", r)
	time.Sleep(time.Millisecond * 500)
	w2 := performRequest("/cache/aborted", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCacheStatus400(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache/400",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusBadRequest, generateID())
			}),
		),
	)

	w1 := performRequest("/cache/400", r)
	time.Sleep(time.Millisecond * 500)
	w2 := performRequest("/cache/400", r)

	assert.Equal(t, http.StatusBadRequest, w1.Code)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.NotEqual(t, w1.Body.String(), w2.Body.String())
}

func TestCacheStatus207(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache/207",
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusMultiStatus, generateID())
			}),
		),
	)

	w1 := performRequest("/cache/207", r)
	time.Sleep(time.Millisecond * 500)
	w2 := performRequest("/cache/207", r)

	assert.Equal(t, http.StatusMultiStatus, w1.Code)
	assert.Equal(t, http.StatusMultiStatus, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCacheLongURI(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET(longUri,
		Cache(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
		),
	)

	w1 := performRequest(longUri, r)
	w2 := performRequest(longUri, r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCacheWithRequestPath(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache_by_path",
		CacheByRequestPath(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
		),
	)

	w1 := performRequest("/cache_by_path?foo=1", r)
	w2 := performRequest("/cache_by_path?foo=2", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
}

func TestCacheWithRequestURI(t *testing.T) {
	store := newStore(time.Second * 60)

	r := gin.New()
	r.GET("/cache_by_uri",
		CacheByRequestURI(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, generateID())
			}),
		),
	)

	w1 := performRequest("/cache_by_uri?foo=1", r)
	w2 := performRequest("/cache_by_uri?foo=1", r)
	w3 := performRequest("/cache_by_uri?foo=2", r)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, http.StatusOK, w3.Code)
	assert.Equal(t, w1.Body.String(), w2.Body.String())
	assert.NotEqual(t, w2.Body.String(), w3.Body.String())
}

type memoryDelayStore struct {
	*memory.MemoryStore
}

func newDelayStore(expire time.Duration) *memoryDelayStore {
	return &memoryDelayStore{memory.NewMemoryStore(expire)}
}

func (c *memoryDelayStore) Set(key string, value interface{}, expires time.Duration) error {
	time.Sleep(time.Millisecond * 50)
	return c.Cache.SetWithTTL(key, value, expires)
}

func TestCacheInSingleflight(t *testing.T) {
	store := newDelayStore(60 * time.Second)

	r := gin.New()
	r.GET("/singleflight",
		CacheByRequestURI(
			WithCacheStore(store),
			WithExpire(time.Second*3),
			WithHandle(func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			}),
		),
	)
	outp := make(chan string, 10)

	for i := 0; i < 5; i++ {
		go func() {
			resp := performRequest("/singleflight", r)
			outp <- resp.Body.String()
		}()
	}
	time.Sleep(time.Millisecond * 500)
	for i := 0; i < 5; i++ {
		go func() {
			resp := performRequest("/singleflight", r)
			outp <- resp.Body.String()
		}()
	}
	time.Sleep(time.Millisecond * 500)

	for i := 0; i < 10; i++ {
		v := <-outp
		assert.Equal(t, "OK", v)
	}
}

func TestBodyWrite(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	writer := &responseCacheWriter{c.Writer, bytes.Buffer{}}
	c.Writer = writer

	c.Writer.WriteHeader(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
	c.Writer.WriteString("foo") // nolint: errcheck
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, "foo", w.Body.String())
	assert.Equal(t, "foo", writer.body.String())
	assert.True(t, c.Writer.Written())
	c.Writer.WriteString("bar") // nolint: errcheck
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, "foobar", w.Body.String())
	assert.Equal(t, "foobar", writer.body.String())
	assert.True(t, c.Writer.Written())
}

func TestDiscard(_ *testing.T) {
	l := NewDiscard()
	l.Debugf("")
	l.Infof("")
	l.Errorf("")
	l.Warnf("")
	l.DPanicf("")
	l.Fatalf("")
}

func TestJSONEncoding(t *testing.T) {
	want := ResponseCache{
		Status: 2,
		Header: nil,
		Data:   []byte{1, 20, 3, 90},
		encode: nil,
	}

	encode := JSONEncoding{}

	data, err := encode.Marshal(want)
	require.NoError(t, err)

	got := ResponseCache{}
	err = encode.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestJSONGzipEncoding(t *testing.T) {
	want := ResponseCache{
		Status: 2,
		Header: nil,
		Data:   []byte{1, 20, 3, 90},
		encode: nil,
	}

	encode := JSONGzipEncoding{}

	data, err := encode.Marshal(want)
	require.NoError(t, err)

	got := ResponseCache{}
	err = encode.Unmarshal(data, &got)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
