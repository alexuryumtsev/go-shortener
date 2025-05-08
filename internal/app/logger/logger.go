// Package logger содержит функции для инициализации и использования логгирования в приложении.
// Он использует библиотеку zap для создания логов и middleware для обработки HTTP-запросов.
package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugarLogger *zap.SugaredLogger

// InitLogger инициализирует логгер с использованием zap.
func InitLogger() {
	logger, err := zap.NewProduction()
	if err != nil {
		// В случае ошибки инициализации используем стандартный логгер
		log.Printf("Failed to initialize zap logger: %v", err)
		return
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			// Логируем ошибку Sync(), но не прерываем работу,
			// так как логгер уже инициализирован
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	sugarLogger = logger.Sugar()
}

// Middleware создает middleware для логирования HTTP-запросов и ответов.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{w, http.StatusOK, 0}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		sugarLogger.Infow("HTTP request",
			"method", r.Method,
			"uri", r.RequestURI,
			"status", ww.status,
			"size", ww.size,
			"duration", duration,
		)
	})
}

// responseWriter оборачивает http.ResponseWriter для захвата статуса и размера ответа.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// WriteHeader записывает статус ответа и сохраняет его.
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write записывает данные в ответ и сохраняет размер ответа.
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
