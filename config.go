package apicache

import (
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for ApiCache
type Config struct {
	Expiration time.Duration
	Prefix     string
	TTLs       []EndpointTTL
}

// EndpointTTL holds the Time-To-Live configuration for specific endpoints
type EndpointTTL struct {
	Path   string
	Method string
	TTL    time.Duration
}

// Default configuration
var defaultConfig = Config{
	Expiration: 24 * time.Hour,
	Prefix:     "",
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() Config {
	expStr := os.Getenv("APICACHE_EXPIRATION") // APICACHE_EXPIRATION=3600 for example
	prefix := os.Getenv("APICACHE_PREFIX")     // APICACHE_PREFIX="myprefix:" for example

	expiration, err := strconv.Atoi(expStr)
	if err != nil {
		expiration = 86400 // Default to 24 hours in seconds
	}

	return Config{
		Expiration: time.Duration(expiration) * time.Second,
		Prefix:     prefix,
	}
}
