package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var sugarLogger *zap.SugaredLogger

func InitLogger() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugarLogger = logger.Sugar()
}

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

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
