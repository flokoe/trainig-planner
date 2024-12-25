package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware wraps an http.HandlerFunc and logs request details
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next(w, r)

		// Log the request details
		log.Printf(
			"%s %s %s %s",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			time.Since(start),
		)
	}
}
