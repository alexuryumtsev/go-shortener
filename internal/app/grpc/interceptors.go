package grpc

import (
	"context"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor создает интерсептор для аутентификации
func AuthInterceptor(userService user.UserService) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Пропускаем аутентификацию для некоторых методов
		if info.FullMethod == "/shortener.Shortener/Ping" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Проверяем наличие токена
		tokens := md.Get("auth_token")
		if len(tokens) == 0 {
			// Генерируем новый токен
			token, err := userService.GenerateUserToken()
			if err != nil {
				return nil, err
			}

			// Добавляем токен в исходящие метаданные
			header := metadata.Pairs("auth_token", token)
			grpc.SendHeader(ctx, header)

			// Добавляем токен во входящие метаданные для использования в обработчике
			md.Append("auth_token", token)
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		return handler(ctx, req)
	}
}

// LoggingInterceptor создает интерсептор для логирования
func LoggingInterceptor(logger *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Вызываем обработчик
		resp, err := handler(ctx, req)

		// Логируем результат
		duration := time.Since(start)
		if err != nil {
			logger.Errorw("gRPC request failed",
				"method", info.FullMethod,
				"duration", duration,
				"error", err,
			)
		} else {
			logger.Infow("gRPC request completed",
				"method", info.FullMethod,
				"duration", duration,
			)
		}

		return resp, err
	}
}

// RecoveryInterceptor создает интерсептор для обработки паник
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = grpc.Errorf(grpc.Code(err), "panic recovered: %v", r)
			}
		}()

		return handler(ctx, req)
	}
}
