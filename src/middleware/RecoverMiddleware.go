package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// 全局异常捕获
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[Recover] panic: %v\n%s", err, debug.Stack())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"code":500,"msg":"服务器内部错误，请联系管理员"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
