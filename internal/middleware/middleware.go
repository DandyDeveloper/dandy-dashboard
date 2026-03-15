package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const headerRequestID = "X-Request-Id"

// responseWriter wraps http.ResponseWriter to capture the status code for
// logging. It also proxies http.Flusher so SSE handlers work through the chain.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Flush proxies to the underlying writer so SSE / chunked responses work.
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Chain applies middlewares right-to-left so the first argument is outermost.
func Chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Logger logs method, path, status, and latency for every request.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", r.Header.Get(headerRequestID),
			)
		})
	}
}

// Recover catches panics and returns a 500 without crashing the server.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						"error", err,
						"path", r.URL.Path,
						"request_id", r.Header.Get(headerRequestID),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID generates a unique ID per request, sets it on both the response
// header and the request header so downstream handlers can read it.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := generateID()
			w.Header().Set(headerRequestID, id)
			// Clone the request so we can add the header for downstream handlers.
			r = r.Clone(r.Context())
			r.Header.Set(headerRequestID, id)
			next.ServeHTTP(w, r)
		})
	}
}

// CORS handles preflight OPTIONS requests and sets access-control headers.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				a := strings.TrimSpace(allowed)
				if a == "*" {
					// Explicit wildcard: set header to * (not the request origin)
					// so the browser enforces the restriction properly.
					w.Header().Set("Access-Control-Allow-Origin", "*")
					break
				}
				if a == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					break
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Dashboard-Key")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKey rejects requests to /api/ that are missing the configured key header.
// A blank key disables the check entirely.
func APIKey(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" &&
				strings.HasPrefix(r.URL.Path, "/api/") &&
				r.Header.Get("X-Dashboard-Key") != key {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%x", b)
}
