package wcache

import (
	"encoding"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type Pool interface {
	Get() *ResponseCache
	Put(*ResponseCache)
}

type cachePool struct {
	pool *sync.Pool
}

// NewPool new pool for responseCache
func NewPool() Pool {
	return &cachePool{
		&sync.Pool{
			New: func() interface{} { return &ResponseCache{Header: make(http.Header)} },
		},
	}
}

// Get implement Pool interface
func (p *cachePool) Get() *ResponseCache {
	return p.pool.Get().(*ResponseCache)
}

// Put implement Pool interface
func (p *cachePool) Put(c *ResponseCache) {
	c.Data = c.Data[:0]
	c.Header = make(http.Header)
	c.encode = nil
	p.pool.Put(c)
}

type ResponseCache struct {
	Status int
	Header http.Header
	Data   []byte
	encode Encoding
}

var _ encoding.BinaryMarshaler = (*ResponseCache)(nil)
var _ encoding.BinaryUnmarshaler = (*ResponseCache)(nil)

func (c *ResponseCache) MarshalBinary() ([]byte, error) {
	return c.encode.Marshal(c)
}

func (c *ResponseCache) UnmarshalBinary(data []byte) error {
	return c.encode.Unmarshal(data, c)
}

func getCacheFromWriter(cacheWriter *responseCacheWriter, encode Encoding) *ResponseCache {
	return &ResponseCache{
		cacheWriter.Status(),
		cacheWriter.Header().Clone(),
		cacheWriter.body.Bytes(),
		encode,
	}
}

func responseWithCache(c *gin.Context, options *Options, respCache *ResponseCache) {
	c.Writer.WriteHeader(respCache.Status)

	for key, values := range respCache.Header {
		for _, val := range values {
			c.Writer.Header().Set(key, val)
		}
	}

	if _, err := c.Writer.Write(respCache.Data); err != nil {
		options.logger.Errorf("write response error: %s", err)
	}

	// abort handler chain and return directly
	c.Abort()
}
