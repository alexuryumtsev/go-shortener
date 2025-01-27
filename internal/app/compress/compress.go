package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Обрабатываем сжатые запросы (Content-Encoding: gzip)
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to read compressed request", http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		// Обрабатываем сжатые ответы (Accept-Encoding: gzip)
		ow := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			ow = &conditionalCompressWriter{
				ResponseWriter: w,
				writer:         gzip.NewWriter(w),
			}
			defer ow.(*conditionalCompressWriter).Close()
		}

		next.ServeHTTP(ow, r)
	})
}

// conditionalCompressWriter сжимает только нужные типы контента.
type conditionalCompressWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (cw *conditionalCompressWriter) WriteHeader(statusCode int) {
	contentType := cw.Header().Get("Content-Type")

	// Сжимаем только application/json и text/html.
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html") {
		cw.Header().Set("Content-Encoding", "gzip")
	}

	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *conditionalCompressWriter) Write(p []byte) (int, error) {
	if cw.Header().Get("Content-Encoding") == "gzip" {
		return cw.writer.Write(p)
	}
	return cw.ResponseWriter.Write(p)
}

func (cw *conditionalCompressWriter) Close() error {
	if cw.Header().Get("Content-Encoding") == "gzip" {
		return cw.writer.Close()
	}
	return nil
}

// compressReader распаковывает входящие данные.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	if r == nil {
		return nil, nil // Если тела запроса нет, возвращаем nil.
	}
	zr, err := gzip.NewReader(r)
	if err != nil {
		r.Close()
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (cr *compressReader) Read(p []byte) (int, error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	if err := cr.zr.Close(); err != nil {
		return err
	}
	return cr.r.Close()
}
