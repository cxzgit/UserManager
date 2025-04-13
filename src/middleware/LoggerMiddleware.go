package middleware

import (
	"log"
	"net/http"
	"time"
)

// 日志
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("[%s] %s %s %s", r.Method, r.URL.Path, r.RemoteAddr, duration)
	})
}
