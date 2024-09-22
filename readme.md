# api-cache

api-cache is a reusable Go package for managing Redis cache for HTTP routes.

## Installation

Ensure you replace `path/to/your/module/apicache` with your actual import path.

## Usage

```go
import (
    "fmt"
    "log"
    "net/http"
    "time"
    "path/to/your/module/apicache" // Update with the actual path of your module
)

func main() {
    redisClient := apicache.NewRedisClient("localhost:6379", "", 0)
    cache := apicache.NewApiCache(redisClient)

    req, _ := http.NewRequest("GET", "http://example.com/resource?id=123", nil)

    err := cache.SetCache(req, map[string]interface{}{
        "data": "example data",
        "id":   123,
    }, 24*time.Hour)

    if err != nil {
        log.Fatalf("Failed to set cache: %v", err)
    }

    cachedData, err := cache.GetCache(req)
    if err != nil {
        log.Fatalf("Failed to get cache: %v", err)
    }

    if cachedData != nil {
        fmt.Printf("Cached data: %v\n", cachedData)
    } else {
        fmt.Println("No cached data for this key.")
    }
}

```