package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"imgnation-backend/config"
)

type CorsMiddleware struct {
	Cfg *config.CorsCfg
}

func (c *CorsMiddleware) Middleware(next http.Handler) http.Handler {
	if c.Cfg == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			for _, o := range c.Cfg.AllowedOrigins {
				if o == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		allowed := false
		for _, o := range c.Cfg.AllowedOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Add("Vary", "Origin")

		if c.Cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == http.MethodOptions {
			if len(c.Cfg.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.Cfg.AllowedMethods, ", "))
			}
			if len(c.Cfg.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.Cfg.AllowedHeaders, ", "))
			}
			if c.Cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.Cfg.MaxAge))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
