package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int
	}{
		{
			name: "No panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantStatus: http.StatusOK,
		},
		{
			name: "With panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			}),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := ErrorMiddleware(tt.handler)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			result := rec.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.wantStatus, result.StatusCode)
		})
	}
}

func TestProcessError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		shortenedURL   string
		responseString bool
		wantStatus     int
		wantBody       string
		wantJSON       bool
	}{
		{
			name: "Unique violation error with string response",
			err: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
			shortenedURL:   "http://short.url/abc",
			responseString: true,
			wantStatus:     http.StatusConflict,
			wantBody:       "http://short.url/abc",
			wantJSON:       false,
		},
		{
			name: "Unique violation error with JSON response",
			err: &pgconn.PgError{
				Code: pgerrcode.UniqueViolation,
			},
			shortenedURL:   "http://short.url/abc",
			responseString: false,
			wantStatus:     http.StatusConflict,
			wantJSON:       true,
		},
		{
			name:         "Generic error",
			err:          errors.New("test error"),
			shortenedURL: "http://short.url/abc",
			wantStatus:   http.StatusBadRequest,
			wantBody:     "test error\n",
			wantJSON:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			ProcessError(rec, tt.err, tt.shortenedURL, tt.responseString)

			result := rec.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.wantStatus, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if tt.wantJSON {
				var response models.ResponseBody
				err = json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Equal(t, tt.shortenedURL, response.ShortURL)
			} else {
				assert.Equal(t, tt.wantBody, string(body))
			}
		})
	}
}
