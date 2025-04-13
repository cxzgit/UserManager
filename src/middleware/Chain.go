package middleware

import "net/http"

// Chain 将多个中间件按给定顺序依次包裹到 handler 上
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
