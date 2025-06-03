package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/alexuryumtsev/go-shortener/internal/app/grpc/pb"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (pb.ShortenerClient, func()) {
	// Создаем моки
	mockStorage := storage.NewMockStorage()
	mockURLService := url.NewMockURLService("http://localhost:8080", nil)
	mockUserService := user.NewMockUserService("test-token")

	// Создаем bufconn.Listener для каждого теста отдельно
	lis := bufconn.Listen(bufSize)

	// Создаем сервер
	server := grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor(mockUserService)),
	)

	shortenerServer, err := NewServer(mockURLService, mockUserService, mockStorage, "")
	require.NoError(t, err)

	pb.RegisterShortenerServer(server, shortenerServer)

	go func() {
		if err := server.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Создаем клиента
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewShortenerClient(conn)

	cleanup := func() {
		conn.Close()
		server.Stop()
		lis.Close()
	}

	return client, cleanup
}

func TestCreateShortURL(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Valid URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "Empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

			resp, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
				Url: tt.url,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.ShortUrl)
				assert.Contains(t, resp.ShortUrl, "http://localhost:8080")
			}
		})
	}
}

func TestCreateShortURLBatch(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

	req := &pb.CreateShortURLBatchRequest{
		Urls: []*pb.URLBatchItem{
			{
				CorrelationId: "1",
				OriginalUrl:   "https://example1.com",
			},
			{
				CorrelationId: "2",
				OriginalUrl:   "https://example2.com",
			},
		},
	}

	resp, err := client.CreateShortURLBatch(ctx, req)
	require.NoError(t, err)
	assert.Len(t, resp.Urls, 2)

	for i, item := range resp.Urls {
		assert.Equal(t, req.Urls[i].CorrelationId, item.CorrelationId)
		assert.Contains(t, item.ShortUrl, "http://localhost:8080")
	}
}

func TestGetUserURLs(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

	// Создаем несколько URL
	urls := []string{"https://example1.com", "https://example2.com", "https://example3.com"}
	for _, url := range urls {
		_, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
			Url: url,
		})
		require.NoError(t, err)
	}

	// Получаем все URL пользователя
	resp, err := client.GetUserURLs(ctx, &pb.GetUserURLsRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Urls), 3)
}

func TestPing(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()
	_, err := client.Ping(ctx, &pb.PingRequest{})
	assert.NoError(t, err)
}

func TestGetStats_Forbidden(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(),
		"auth_token", "test-token",
		"x-real-ip", "192.168.1.1",
	)

	_, err := client.GetStats(ctx, &pb.GetStatsRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Forbidden")
}

func TestDeleteUserURLs(t *testing.T) {
	client, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

	// Создаем URL
	createResp, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
		Url: "https://example.com",
	})
	require.NoError(t, err)

	// Удаляем URL
	_, err = client.DeleteUserURLs(ctx, &pb.DeleteUserURLsRequest{
		ShortUrls: []string{createResp.ShortUrl},
	})
	assert.NoError(t, err)
}
