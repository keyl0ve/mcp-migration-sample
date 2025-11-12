package middleware

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func HTTPLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s from %s -> %d (%s)", r.Method, r.URL.Path, r.RemoteAddr, lrw.status, duration)
	})
}
