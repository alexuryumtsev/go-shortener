package main

import (
	"context"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alexuryumtsev/go-shortener/config"
	"github.com/alexuryumtsev/go-shortener/internal/app/db"
	grpcserver "github.com/alexuryumtsev/go-shortener/internal/app/grpc"
	"github.com/alexuryumtsev/go-shortener/internal/app/grpc/pb"
	"github.com/alexuryumtsev/go-shortener/internal/app/logger"
	"github.com/alexuryumtsev/go-shortener/internal/app/router"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/file"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/memory"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/pg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Информация о сборке приложения.
// Заполняется при компиляции с помощью флагов линковщика:
// -ldflags "-X main.buildVer=v1.0.0 -X main.buildDt=2023-05-11 -X main.buildCmt=abc123"
var (
	buildVer string // версия сборки
	buildDt  string // дата сборки
	buildCmt string // коммит, на котором собрана версия
)

// Функция для вывода информации о сборке
func printBuildInfo() {
	version := buildVer
	if version == "" {
		version = "N/A"
	}

	date := buildDt
	if date == "" {
		date = "N/A"
	}

	commit := buildCmt
	if commit == "" {
		commit = "N/A"
	}

	log.Printf("Build version: %s\n", version)
	log.Printf("Build date: %s\n", date)
	log.Printf("Build commit: %s\n", commit)
}

func main() {
	// Инициализируем конфигурацию
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Выводим информацию о сборке
	printBuildInfo()

	// Инициализируем логгер
	logger.InitLogger()

	// Подключаемся к базе данных
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var repo storage.URLStorage
	var dbPool *db.Database // Сохраняем ссылку на пул соединений для корректного закрытия

	if cfg.DatabaseDSN() != "" {
		pool, err := db.NewDatabaseConnection(ctx, cfg.DatabaseDSN())
		if err != nil {
			log.Fatalf("Failed connect to db: %v", err)
		}
		dbPool = pool
		repo = pg.NewDatabaseStorage(pool)
	} else if cfg.FileStoragePath() != "" {
		repo = file.NewFileStorage(cfg.FileStoragePath())
	} else {
		repo = memory.NewInMemoryStorage()
	}

	// Инициализируем сервисы
	userService := user.NewUserService("super-secret-key")
	urlService := url.NewURLService(repo, cfg.BaseURL(), cfg.BatchSize())

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    cfg.ServerAddress(),
		Handler: router.ShortenerRouter(cfg, repo, userService, urlService),
	}

	// Создаем gRPC сервер если включен
	var grpcServer *grpc.Server
	var grpcListener net.Listener
	if cfg.EnableGRPC() {
		// Создаем gRPC сервер
		grpcListener, err = net.Listen("tcp", cfg.GRPCAddress())
		if err != nil {
			log.Fatalf("Failed to listen on %s: %v", cfg.GRPCAddress(), err)
		}

		// Создаем логгер для gRPC
		zapLogger, _ := zap.NewProduction()
		sugarLogger := zapLogger.Sugar()

		// Создаем gRPC сервер с интерсепторами
		grpcServer = grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpcserver.RecoveryInterceptor(),
				grpcserver.LoggingInterceptor(sugarLogger),
				grpcserver.AuthInterceptor(userService),
			),
		)

		// Регистрируем сервис
		shortenerServer, err := grpcserver.NewServer(urlService, userService, repo, cfg.TrustedSubnet())
		if err != nil {
			log.Fatalf("Failed to create gRPC server: %v", err)
		}
		pb.RegisterShortenerServer(grpcServer, shortenerServer)

		log.Printf("gRPC server will start at %s\n", cfg.GRPCAddress())
	}

	// Канал для получения сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// WaitGroup для ожидания завершения горутин
	var wg sync.WaitGroup

	// Горутина для запуска сервера
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("Server started at %s\n", cfg.ServerAddress())

		var err error
		if cfg.EnableHTTPS() {
			err = server.ListenAndServeTLS(cfg.CertPath(), cfg.KeyPath())
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Горутина для запуска gRPC сервера
	if cfg.EnableGRPC() && grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("gRPC server started at %s\n", cfg.GRPCAddress())

			if err := grpcServer.Serve(grpcListener); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	// Горутина для обработки сигналов завершения
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Ожидаем сигнал завершения
		sig := <-sigChan
		log.Printf("\nReceived signal: %v. Starting graceful shutdown...\n", sig)

		// Создаем контекст с таймаутом для graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Останавливаем HTTP сервер
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		} else {
			log.Println("HTTP server stopped gracefully")
		}

		// Останавливаем gRPC сервер
		if grpcServer != nil {
			grpcServer.GracefulStop()
			log.Println("gRPC server stopped gracefully")
		}

		// Сохраняем данные в хранилище
		if err := repo.Close(); err != nil {
			log.Printf("Storage close error: %v", err)
		} else {
			log.Println("Storage closed gracefully")
		}

		// Закрываем соединение с базой данных
		if dbPool != nil {
			dbPool.Close()
			log.Println("Database connection closed")
		}

		// Отменяем основной контекст
		cancel()
	}()

	// Ожидаем завершения всех горутин
	wg.Wait()
	log.Println("Application shutdown completed")
}
