package accesslog

import (
	"log/slog"
	"net/http"
)

func NewMiddleware(logger *slog.Logger, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = NewResponseWriter(w)

		defer func() {
			logger.With(
				"method", r.Method,
				"path", r.URL.Path,
				"status", w.(*ResponseWriter).Status(),
				"bytesIn", r.ContentLength,
				"bytesOut", w.(*ResponseWriter).BytesOut(),
			).DebugContext(r.Context(), "handled request")
		}()

		handler.ServeHTTP(w, r)
	}
}
