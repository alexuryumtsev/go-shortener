package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServerAddress(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		wantErr bool
	}{
		{
			name:    "Valid address with port only",
			addr:    ":8080",
			wantErr: false,
		},
		{
			name:    "Valid address with host and port",
			addr:    "localhost:8080",
			wantErr: false,
		},
		{
			name:    "Valid address with IP and port",
			addr:    "127.0.0.1:8080",
			wantErr: false,
		},
		{
			name:    "Invalid: missing port",
			addr:    "localhost",
			wantErr: true,
		},
		{
			name:    "Invalid: wrong port format",
			addr:    ":abc",
			wantErr: true,
		},
		{
			name:    "Invalid: empty address",
			addr:    "",
			wantErr: true,
		},
		{
			name:    "Invalid: wrong separator",
			addr:    "localhost-8080",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerAddress(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid HTTP URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "Valid HTTPS URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "Valid URL with port",
			url:     "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "Valid URL with path",
			url:     "http://example.com/path",
			wantErr: false,
		},
		{
			name:    "Invalid: empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "Invalid: missing scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "Invalid: malformed URL",
			url:     "http://[invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBaseURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
