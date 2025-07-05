package http

import (
	"net/http"
)

// InternalTokenMiddleware проверяет внутренний токен для сервисных запросов
func InternalTokenMiddleware(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Internal-Token")
			if token == "" || token != expectedToken {
				http.Error(w, "Forbidden: invalid internal token", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
