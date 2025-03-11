package middleware

import (
	"net/http"
	"time"

	"github.com/bookshop/api/pkg/logger"
)

// Logging добавляет логирование запросов
func Logging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Создаем ResponseWriter, который может отслеживать статус ответа
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Обрабатываем запрос
			next.ServeHTTP(rw, r)

			// Логируем информацию о запросе
			duration := time.Since(start)
			log.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWriter оборачивает http.ResponseWriter для отслеживания статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader перехватывает статус ответа
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
