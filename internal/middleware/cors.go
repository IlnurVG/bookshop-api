package middleware

import (
	"net/http"
)

// CORS добавляет заголовки CORS в ответ
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем origin из заголовка запроса
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		// Устанавливаем заголовки CORS
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Обрабатываем preflight запросы
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
