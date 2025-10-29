package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"kln-test/internal/config"
)

type Middleware func(http.Handler) http.Handler

// Chain combines multiple middleware into a single middleware
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Recovery middleware recovers from panics and returns a 500 error
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("panic: %v\n%s", err, debug.Stack())
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Auth middleware validates the API key
func Auth(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authConfig := cfg.GetAuthConfig()
			username, password, ok := r.BasicAuth()

			if !ok || username != authConfig.Username || password != authConfig.Password {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateHeaders middleware ensures required headers are present
func ValidateHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				http.Error(w, "Authorization header is required", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ConfigReload middleware checks if config needs to be reloaded
// For similicity, it reloads config on every request in this example
// In a real-world scenario, file watchers would be used to detect changes
func ConfigReload(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := cfg.Reload(); err != nil {
				log.Printf("Failed to reload config: %v", err)
				http.Error(w, "Failed to reload config", http.StatusInternalServerError)
				return
			}
			log.Println("Configuration reloaded successfully")

			next.ServeHTTP(w, r)
		})
	}
}
