package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write(body)
		} else {
			w.Write([]byte(`{"message": "hello, world"}`))
		}
	}))

	tests := []struct {
		name           string
		method         string
		body           []byte
		compressed     bool
		acceptEncoding string
		contentType    string
		wantStatus     int
		wantBody       string
	}{
		{
			name:           "Client supports gzip response",
			method:         http.MethodGet,
			acceptEncoding: "gzip",
			contentType:    "application/json",
			wantStatus:     http.StatusOK,
			wantBody:       "",
		},
		{
			name:        "Client sends gzipped request",
			method:      http.MethodPost,
			body:        []byte(`{"test": "data"}`),
			compressed:  true,
			contentType: "application/json",
			wantStatus:  http.StatusOK,
			wantBody:    `{"test": "data"}`,
		},
		{
			name:        "Invalid gzip data",
			method:      http.MethodPost,
			body:        []byte("invalid gzip data"),
			compressed:  true,
			contentType: "application/json",
			wantStatus:  http.StatusOK,
		},
		{
			name:           "Plain text content type",
			method:         http.MethodGet,
			acceptEncoding: "gzip",
			contentType:    "text/plain",
			wantStatus:     http.StatusOK,
			wantBody:       "",
		},
		{
			name:        "Empty request body",
			method:      http.MethodPost,
			compressed:  true,
			contentType: "application/json",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				if tt.compressed {
					var buf bytes.Buffer
					gw := gzip.NewWriter(&buf)
					_, err := gw.Write(tt.body)
					require.NoError(t, err)
					require.NoError(t, gw.Close())
					reqBody = &buf
				} else {
					reqBody = bytes.NewReader(tt.body)
				}
			}

			req := httptest.NewRequest(tt.method, "/", reqBody)
			if tt.compressed {
				req.Header.Set("Content-Encoding", "gzip")
			}
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)

			if tt.wantBody != "" {
				var body []byte
				var err error

				if res.Header.Get("Content-Encoding") == "gzip" {
					gz, err := gzip.NewReader(res.Body)
					require.NoError(t, err)
					body, err = io.ReadAll(gz)
					require.NoError(t, err)
					require.NoError(t, gz.Close())
				} else {
					body, err = io.ReadAll(res.Body)
					require.NoError(t, err)
				}

				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}

func TestCompressReader_Close(t *testing.T) {
	// Test normal close
	data := []byte("test data")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(data)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	cr, err := newCompressReader(io.NopCloser(&buf))
	require.NoError(t, err)
	require.NoError(t, cr.Close())

	// Test nil reader
	cr, err = newCompressReader(nil)
	assert.NoError(t, err)
	assert.Nil(t, cr)
}

func BenchmarkGzipMiddleware(b *testing.B) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello, world", "data": "{"key": "value"}"}`))
	}))

	benchmarks := []struct {
		name           string
		acceptEncoding string
	}{
		{
			name:           "With Gzip",
			acceptEncoding: "gzip",
		},
		{
			name:           "Without Gzip",
			acceptEncoding: "",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// Сброс таймера перед началом измерений
			b.StopTimer() // Останавливаем таймер перед началом итерации
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept-Encoding", bm.acceptEncoding)
				rr := httptest.NewRecorder()

				b.StartTimer() // Начинаем измерение
				handler.ServeHTTP(rr, req)
				b.StopTimer() // Останавливаем измерение

				res := rr.Result()
				_, _ = io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		})
	}
}
