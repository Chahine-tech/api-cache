package apicache

import (
	"bytes"
	"log"
	"net/http"
	"time"
)

// CacheMiddleware is a middleware that caches responses
func CacheMiddleware(apiCache *ApiCache, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cachedResponse, err := apiCache.GetCache(r)
			if err == nil && cachedResponse != nil {
				// Send cached response to the client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write(cachedResponse.([]byte))
				return
			}

			// Capture the response for caching
			cw := &cacheWriter{ResponseWriter: w, buf: &bytes.Buffer{}}
			next.ServeHTTP(cw, r)

			success, err := apiCache.SetCache(r, cw.buf.Bytes())
			if err != nil || !success {
				log.Printf("failed setting cache: %v", err)
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
