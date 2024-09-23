package apicache

import "time"

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
