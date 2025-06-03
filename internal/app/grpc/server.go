// Package grpc содержит реализацию gRPC сервера
package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alexuryumtsev/go-shortener/internal/app/grpc/pb"
	"github.com/alexuryumtsev/go-shortener/internal/app/models"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/url"
	"github.com/alexuryumtsev/go-shortener/internal/app/service/user"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage"
	"github.com/alexuryumtsev/go-shortener/internal/app/storage/pg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server представляет gRPC сервер
type Server struct {
	pb.UnimplementedShortenerServer
	urlService  url.URLService
	userService user.UserService
	storage     storage.URLStorage
	trustedNet  *net.IPNet
}

// NewServer создает новый gRPC сервер
func NewServer(urlService url.URLService, userService user.UserService, storage storage.URLStorage, trustedSubnet string) (*Server, error) {
	s := &Server{
		urlService:  urlService,
		userService: userService,
		storage:     storage,
	}

	// Парсим доверенную подсеть, если она задана
	if trustedSubnet != "" {
		_, subnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted subnet: %w", err)
		}
		s.trustedNet = subnet
	}

	return s, nil
}

// getUserIDFromContext извлекает ID пользователя из контекста
func (s *Server) getUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata found")
	}

	tokens := md.Get("auth_token")
	if len(tokens) == 0 {
		return "", errors.New("no auth token found")
	}

	userID, err := s.userService.VerifyUserToken(tokens[0])
	if err != nil {
		// Если токен невалидный, генерируем новый
		token, err := s.userService.GenerateUserToken()
		if err != nil {
			return "", err
		}
		userID, _ = s.userService.VerifyUserToken(token)
	}

	return userID, nil
}

// CreateShortURL создает короткий URL
func (s *Server) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	if strings.TrimSpace(req.Url) == "" {
		return nil, status.Error(codes.InvalidArgument, "empty URL")
	}

	shortURL, err := s.urlService.ShortenerURL(ctx, req.Url, userID)
	if err != nil {
		// Проверяем, является ли это конфликтом (URL уже существует)
		if strings.Contains(err.Error(), "conflict") {
			// Возвращаем успех с существующим URL
			return &pb.CreateShortURLResponse{ShortUrl: shortURL}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateShortURLResponse{ShortUrl: shortURL}, nil
}

// CreateShortURLJSON создает короткий URL (JSON версия)
func (s *Server) CreateShortURLJSON(ctx context.Context, req *pb.CreateShortURLJSONRequest) (*pb.CreateShortURLJSONResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	if strings.TrimSpace(req.Url) == "" {
		return nil, status.Error(codes.InvalidArgument, "empty URL")
	}

	shortURL, err := s.urlService.ShortenerURL(ctx, req.Url, userID)
	if err != nil {
		// Проверяем, является ли это конфликтом (URL уже существует)
		if strings.Contains(err.Error(), "conflict") {
			// Возвращаем успех с существующим URL
			return &pb.CreateShortURLJSONResponse{Result: shortURL}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateShortURLJSONResponse{Result: shortURL}, nil
}

// CreateShortURLBatch создает несколько коротких URL
func (s *Server) CreateShortURLBatch(ctx context.Context, req *pb.CreateShortURLBatchRequest) (*pb.CreateShortURLBatchResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	if len(req.Urls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty batch")
	}

	// Преобразуем proto модели в модели сервиса
	var batchModels []models.URLBatchModel
	for _, item := range req.Urls {
		batchModels = append(batchModels, models.URLBatchModel{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		})
	}

	responseModels, err := s.urlService.SaveBatchShortenerURL(ctx, batchModels, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Преобразуем ответ в proto модели
	var response pb.CreateShortURLBatchResponse
	for _, model := range responseModels {
		response.Urls = append(response.Urls, &pb.BatchResponseItem{
			CorrelationId: model.CorrelationID,
			ShortUrl:      model.ShortURL,
		})
	}

	return &response, nil
}

// GetOriginalURL получает оригинальный URL по короткому
func (s *Server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "URL ID is required")
	}

	originalURL, exists, err := s.urlService.GetURLByID(ctx, req.Id)
	if !exists {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "This URL is no longer available as it has been deleted by the owner")
	}

	return &pb.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

// GetUserURLs получает все URL пользователя
func (s *Server) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	userURLs, err := s.urlService.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get user URLs")
	}

	// Преобразуем в proto модели
	var response pb.GetUserURLsResponse
	for _, url := range userURLs {
		response.Urls = append(response.Urls, &pb.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &response, nil
}

// DeleteUserURLs удаляет URL пользователя
func (s *Server) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	userID, err := s.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication failed")
	}

	if len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no URLs to delete")
	}

	// Запускаем асинхронное удаление
	go func() {
		deleteCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.urlService.DeleteUserURLsBatch(deleteCtx, userID, req.ShortUrls)
	}()

	return &pb.DeleteUserURLsResponse{}, nil
}

// Ping проверяет соединение с БД
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	if dbRepo, ok := s.storage.(*pg.DatabaseStorage); ok {
		if err := dbRepo.Ping(ctx); err != nil {
			return nil, status.Error(codes.Unavailable, "Database connection error")
		}
	}
	return &pb.PingResponse{}, nil
}

// GetStats возвращает статистику
func (s *Server) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	// Проверяем доступ по IP
	if s.trustedNet == nil {
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}

	// Получаем IP клиента из метаданных
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}

	// Получаем IP из заголовка X-Real-IP
	realIPs := md.Get("x-real-ip")
	if len(realIPs) == 0 {
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}

	// Парсим IP адрес
	clientIP := net.ParseIP(realIPs[0])
	if clientIP == nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid client IP")
	}

	// Проверяем, что IP в доверенной подсети
	if !s.trustedNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "Forbidden")
	}

	// Получаем статистику
	stats, err := s.storage.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to get statistics")
	}

	return &pb.GetStatsResponse{
		Urls:  int32(stats.URLs),
		Users: int32(stats.Users),
	}, nil
}
