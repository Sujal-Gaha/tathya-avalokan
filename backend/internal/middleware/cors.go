package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// NewCORSMiddleware configures Chi CORS middleware with supported methods and headers.
// When allowedOrigins is ["*"], it uses AllowOriginFunc to reflect the request origin,
// which adheres to W3C/WHATWG specs when AllowCredentials is true.
func NewCORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	opts := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		opts.AllowOriginFunc = func(r *http.Request, origin string) bool {
			return true
		}
	} else {
		opts.AllowedOrigins = allowedOrigins
	}

	return cors.Handler(opts)
}
