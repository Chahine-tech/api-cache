package apicache

import "time"

// Config holds the configuration for ApiCache
type Config struct {
	Expiration time.Duration
	Prefix     string
}

// Default configuration
var defaultConfig = Config{
	Expiration: 24 * time.Hour,
	Prefix:     "",
}
