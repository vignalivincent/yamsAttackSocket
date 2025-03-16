package api

import (
	"net/http"
	"time"

	"github.com/vincentvignali/yamsAttackSocket/internal/logger"
)

func WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info.Printf("Request received: %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next(w, r)
		duration := time.Since(start)
		
		if duration > 500*time.Millisecond {
			logger.Warn.Printf("Request processed: %s %s %s (duration: %v)", 
				r.Method, r.URL.Path, r.RemoteAddr, duration)
		} else {
			logger.Info.Printf("Request processed: %s %s %s (duration: %v)", 
				r.Method, r.URL.Path, r.RemoteAddr, duration)
		}
	}
}
