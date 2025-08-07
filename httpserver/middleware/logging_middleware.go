package middleware

import (
	"net/http"

	"github.com/rs/zerolog/hlog"
)

// Custom ResponseWriter to capture the status code
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := newStatusRecorder(w)
			next.ServeHTTP(rec, r)

			// Extract the logger from the request's context
			log := hlog.FromRequest(r)

			// Log the method, URL, and status code
			log.Info().
				Int("status", rec.statusCode).
				Send()
		})
	}
}
