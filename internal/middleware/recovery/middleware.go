package recovery

import (
	"net/http"

	"github.com/wroersma/libgo/pkg/logger"
)

// RecoveryMiddleware returns a middleware function that recovers from panics
func RecoveryMiddleware(log logger.Logger) func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request, next func(http.ResponseWriter, *http.Request)) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Error("Panic recovered in HTTP middleware",
					logger.String("method", r.Method),
					logger.String("path", r.URL.Path),
					logger.Any("error", err))

				// Return 500 error
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": 500, "code": "INTERNAL_SERVER_ERROR", "message": "Internal server error"}`))
			}
		}()

		// Call the next handler
		next(w, r)
	}
}
