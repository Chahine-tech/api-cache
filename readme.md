# api-cache

api-cache is a reusable Go package for managing Redis cache for HTTP routes.

## Installation

```sh
go get github.com/Chahine-tech/api-cache
```

## Usage

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
    "github.com/Chahine-tech/api-cache"
)

func main() {
    // Create a new Redis client
    redisClient := apicache.NewRedisClient("localhost:6379", "", 0)

    // Create an ApiCache instance
    cache := apicache.NewApiCache(redisClient)

    // Example HTTP request to build the cache key
    req, _ := http.NewRequest("GET", "http://example.com/resource?id=123", nil)

    // Set a value in the cache
    err := cache.SetCache(req, map[string]interface{}{
        "data": "example data",
        "id":   123,
    })
    if err != nil {
        log.Fatalf("Failed to set cache: %v", err)
    }

    // Get a value from the cache
    cachedData, err := cache.GetCache(req)
    if err != nil {
        log.Fatalf("Failed to get cache: %v", err)
    }

    if cachedData != nil {
        fmt.Printf("Cached data: %v\n", cachedData)
    } else {
        fmt.Println("No cached data for this key.")
    }

    // Set up an HTTP server and use CacheMiddleware
    mux := http.NewServeMux()

    // Use the CacheMiddleware for the /resource route
    mux.Handle("/resource", apicache.CacheMiddleware(cache, 24*time.Hour)(http.HandlerFunc(resourceHandler)))

    // Start the HTTP server
    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
    // Handle the request and send a response
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"message": "Hello, World!"}`))
}
```