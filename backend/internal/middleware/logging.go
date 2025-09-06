package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/cbalite/backend/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func NewLoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()
			
			w.Header().Set("X-Request-ID", requestID)
			
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			log.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remote_ip":  r.RemoteAddr,
			}).Info("Request started")

			next.ServeHTTP(wrapped, r)

			log.WithFields(map[string]interface{}{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     wrapped.status,
				"size":       wrapped.size,
				"duration":   time.Since(start).String(),
			}).Info("Request completed")
		})
	}
}