package apicache

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

// CacheMiddleware is a middleware that caches GET requests
func CacheMiddleware(apiCache *ApiCache, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				// Do not cache responses other than GET requests
				next.ServeHTTP(w, r)
				return
			}

			cachedResponse, err := apiCache.GetCache(r)
			if err == nil && cachedResponse != nil {
				// Send cached response to the client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if bytesResponse, ok := cachedResponse.([]byte); ok {
					w.Write(bytesResponse)
				} else {
					log.Printf("Failed to cast cached response to []byte: %v", cachedResponse)
				}
				return
			}

			// Capture the response for caching
			cw := &cacheWriter{ResponseWriter: w, buf: &bytes.Buffer{}}
			next.ServeHTTP(cw, r)

			success, err := apiCache.SetCache(r, cw.buf.Bytes())
			if err != nil {
				log.Printf("Failed setting cache: %v", err)
			} else if !success {
				log.Printf("Cache set operation not successful")
			}
		})
	}
}

type cacheWriter struct {
	http.ResponseWriter
	buf *bytes.Buffer
}

func (cw *cacheWriter) Write(b []byte) (int, error) {
	cw.buf.Write(b)
	return cw.ResponseWriter.Write(b)
}
