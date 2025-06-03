// Пример клиента для gRPC сервера
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	var (
		serverAddr = flag.String("server", "localhost:50051", "gRPC server address")
		authToken  = flag.String("token", "", "Authentication token")
	)
	flag.Parse()

	// Устанавливаем соединение с сервером
	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Создаем клиента
	client := pb.NewShortenerClient(conn)

	// Создаем контекст с токеном аутентификации
	ctx := context.Background()
	if *authToken != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "auth_token", *authToken)
	}

	// Пример 1: Создание короткого URL
	log.Println("=== Creating short URL ===")
	createResp, err := client.CreateShortURL(ctx, &pb.CreateShortURLRequest{
		Url: "https://github.com/golang/go",
	})
	if err != nil {
		log.Printf("Failed to create short URL: %v", err)
	} else {
		log.Printf("Short URL created: %s", createResp.ShortUrl)
	}

	// Пример 2: Создание нескольких коротких URL
	log.Println("\n=== Creating batch of short URLs ===")
	batchResp, err := client.CreateShortURLBatch(ctx, &pb.CreateShortURLBatchRequest{
		Urls: []*pb.URLBatchItem{
			{CorrelationId: "1", OriginalUrl: "https://golang.org"},
			{CorrelationId: "2", OriginalUrl: "https://go.dev"},
			{CorrelationId: "3", OriginalUrl: "https://pkg.go.dev"},
		},
	})
	if err != nil {
		log.Printf("Failed to create batch: %v", err)
	} else {
		for _, item := range batchResp.Urls {
			log.Printf("Batch item %s: %s", item.CorrelationId, item.ShortUrl)
		}
	}

	// Пример 3: Получение всех URL пользователя
	log.Println("\n=== Getting user URLs ===")
	userURLsResp, err := client.GetUserURLs(ctx, &pb.GetUserURLsRequest{})
	if err != nil {
		log.Printf("Failed to get user URLs: %v", err)
	} else {
		for _, url := range userURLsResp.Urls {
			log.Printf("URL: %s -> %s", url.ShortUrl, url.OriginalUrl)
		}
	}

	// Пример 4: Проверка соединения с БД
	log.Println("\n=== Pinging database ===")
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err = client.Ping(pingCtx, &pb.PingRequest{})
	if err != nil {
		log.Printf("Ping failed: %v", err)
	} else {
		log.Println("Ping successful")
	}

	// Пример 5: Получение статистики (требует доверенный IP)
	log.Println("\n=== Getting stats ===")
	statsCtx := metadata.AppendToOutgoingContext(ctx, "x-real-ip", "192.168.1.100")
	statsResp, err := client.GetStats(statsCtx, &pb.GetStatsRequest{})
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		log.Printf("Stats: URLs=%d, Users=%d", statsResp.Urls, statsResp.Users)
	}
}
