package middleware

import (
	"net/http"
	"time"

	"github.com/DennisPing/cs6650-twinder-a3/lib/logger"
	"github.com/rs/zerolog/hlog"
)

var zlog = logger.GetLogger()

func LoggingMiddleware(next http.Handler) http.Handler {
	h := hlog.NewHandler(zlog)

	accessHandler := hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("code", status).
				Dur("response_time", duration).
				Msg("request")
		})

	return h(accessHandler(next))
}
