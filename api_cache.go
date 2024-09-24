package apicache

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmihailenco/msgpack/v5"
)

// ApiCache structure
type ApiCache struct {
	redisClient *redis.Client
	config      Config
	mu          sync.Mutex
}

// NewApiCache creates a new ApiCache instance with optional environment-based configuration
func NewApiCache(redisClient *redis.Client, useEnv bool, config ...Config) *ApiCache {
	cfg := defaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if useEnv {
		envConfig := LoadConfigFromEnv()
		cfg.Expiration = envConfig.Expiration
		cfg.Prefix = envConfig.Prefix
	}
	return &ApiCache{
		redisClient: redisClient,
		config:      cfg,
	}
}

// GetCache retrieves data from the cache for a given request
func (c *ApiCache) GetCache(req *http.Request) (interface{}, error) {
	// By default, only cache GET requests
	if req.Method != http.MethodGet {
		return nil, nil
	}

	key := buildKey(req, c.config.Prefix)

	// Mutex Lock to ensure thread-safety
	c.mu.Lock()
	defer c.mu.Unlock()

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
	// By default, only cache GET requests
	if req.Method != http.MethodGet {
		return false, nil
	}

	// Apply rate-limiting before setting cache
	if !c.rateLimit(req) {
		return false, nil
	}

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

	// Mutex Lock to ensure thread-safety
	c.mu.Lock()
	defer c.mu.Unlock()

	pipeliner := c.redisClient.Pipeline()
	pipeliner.Set(context.Background(), key, encodedData, ttl)
	_, err = pipeliner.Exec(context.Background())
	if err != nil {
		return false, err
	}

	return true, nil
}

// InvalidateCache invalidates the cache for a given request
func (c *ApiCache) InvalidateCache(req *http.Request) (bool, error) {
	// By default, only invalidate cache for GET requests
	if req.Method == http.MethodGet {
		key := buildKey(req, c.config.Prefix)

		// Mutex Lock to ensure thread-safety
		c.mu.Lock()
		defer c.mu.Unlock()

		count, err := c.redisClient.Del(context.Background(), key).Result()
		if err != nil {
			return false, err
		}
		return count > 0, nil
	}
	return false, nil
}

// rateLimit applies a simple rate-limiting mechanism
func (c *ApiCache) rateLimit(req *http.Request) bool {
	key := buildKey(req, c.config.Prefix) + ":rate-limit"
	limit := 10 // Allow max 10 requests per minute
	duration := time.Minute

	current, err := c.redisClient.Incr(context.Background(), key).Result()
	if err != nil {
		return false
	}

	if current == 1 {
		c.redisClient.Expire(context.Background(), key, duration)
	}

	if current > int64(limit) {
		return false
	}

	return true
}

// getTTL returns the TTL for a specific request based on the config
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
