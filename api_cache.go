package apicache

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

// ApiCache structure
type ApiCache struct {
	redisClient *redis.Client
	config      Config
}

// NewApiCache creates a new ApiCache instance
func NewApiCache(redisClient *redis.Client, config ...Config) *ApiCache {
	cfg := defaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	return &ApiCache{
		redisClient: redisClient,
		config:      cfg,
	}
}

// GetCache retrieves data from the cache for a given request
func (c *ApiCache) GetCache(req *http.Request) (interface{}, error) {
	key := buildKey(req, c.config.Prefix)
	rawData, err := c.redisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	} else if err != nil {
		return nil, err
	}

	data, err := base64.StdEncoding.DecodeString(rawData)
	if err != nil {
		return nil, err
	}

	decompressedData, err := decompress(data)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := msgpack.Unmarshal(decompressedData, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SetCache stores data in the cache for a given request
func (c *ApiCache) SetCache(req *http.Request, data interface{}) (bool, error) {
	key := buildKey(req, c.config.Prefix)
	packedData, err := msgpack.Marshal(data)
	if err != nil {
		return false, err
	}

	compressedData, err := compress(packedData)
	if err != nil {
		return false, err
	}

	encodedData := base64.StdEncoding.EncodeToString(compressedData)
	ttl := c.getTTL(req)
	status, err := c.redisClient.Set(context.Background(), key, encodedData, ttl).Result()
	if err != nil {
		return false, err
	}

	return status == "OK", nil
}

// InvalidateCache invalidates the cache for a given request
func (c *ApiCache) InvalidateCache(req *http.Request) (bool, error) {
	key := buildKey(req, c.config.Prefix)
	count, err := c.redisClient.Del(context.Background(), key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *ApiCache) getTTL(req *http.Request) time.Duration {
	for _, ttl := range c.config.TTLs {
		if ttl.Path == req.URL.Path && ttl.Method == req.Method {
			return ttl.TTL
		}
	}
	return c.config.Expiration
}

// buildKey generates a cache key based on the request
func buildKey(req *http.Request, prefix string) string {
	query := req.URL.Query().Encode()
	method := req.Method
	path := strings.TrimPrefix(req.URL.Path, "/")

	return strings.ToLower(
		prefix + method + "__" + path + "__" + query,
	)
}
