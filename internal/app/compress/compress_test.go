package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "hello, world"}`))
	}))

	tests := []struct {
		name               string
		acceptEncoding     string
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Client supports gzip",
			acceptEncoding:     "gzip",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message": "hello, world"}`,
		},
		{
			name:               "Client does not support gzip",
			acceptEncoding:     "",
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"message": "hello, world"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatusCode {
				t.Errorf("expected status %d, got %d", tt.expectedStatusCode, res.StatusCode)
			}

			var body []byte
			var err error

			if strings.Contains(res.Header.Get("Content-Encoding"), "gzip") {
				gz, err := gzip.NewReader(res.Body)
				require.NoError(t, err)

				body, err = io.ReadAll(gz)
				require.NoError(t, err)

				defer gz.Close()
			} else {
				body, err = io.ReadAll(res.Body)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedBody, string(body))
		})
	}
}

// ...existing code...

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
			// Reset timer before each benchmark iteration
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				req.Header.Set("Accept-Encoding", bm.acceptEncoding)
				rr := httptest.NewRecorder()

				handler.ServeHTTP(rr, req)

				res := rr.Result()
				_, _ = io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		})
	}
}
