package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/bookshop/api/pkg/logger"
)

// Recovery восстанавливает работу после паники
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Логируем информацию о панике
					errMsg := fmt.Sprintf("panic recovered: %v", err)
					log.Error(errMsg, fmt.Errorf("%v\nstack: %s\npath: %s\nmethod: %s",
						err,
						string(debug.Stack()),
						r.URL.Path,
						r.Method))

					// Отправляем ответ с ошибкой
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"error":"internal server error"}`)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
