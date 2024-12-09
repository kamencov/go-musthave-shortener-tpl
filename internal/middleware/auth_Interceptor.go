package middleware

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

// UnaryInterceptor - гRPC интерсептор для логирования и проверки авторизации.
func (a *AuthMiddleware) UnaryInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	// Логируем запрос с использованием вашего логгера
	start := time.Now()
	a.log.Info(
		"gRPC request started",
		"method", info.FullMethod,
		"request", req,
	)

	// Методы, которые требуют CheckAuthInterceptor
	methodsRequiringCheckAuth := map[string]bool{
		"/shortener.ShortenerService/GetUsersURLs": true,
		"/shortener.ShortenerService/DeletionURLs": true,
		"/shortener.ShortenerService/GetStatus":    true,
	}

	// Если метод в списке, применяем CheckAuthInterceptor
	if methodsRequiringCheckAuth[info.FullMethod] {
		resp, err = a.UnaryCheckAuthInterceptor(ctx, req, info, handler)
	} else {
		// Для всех остальных методов применяем AuthInterceptor
		resp, err = a.UnaryAuthInterceptor(ctx, req, info, handler)
	}

	// Логируем завершение запроса с использованием логгера
	duration := time.Since(start)
	if err != nil {
		// Логируем ошибку
		a.log.Error(
			"gRPC request failed",
			"method", info.FullMethod,
			"error", err,
			"duration", duration,
		)
		return resp, err
	}
	// Логируем успешный ответ
	a.log.Info(
		"gRPC request completed",
		"method", info.FullMethod,
		"response", resp,
		"duration", duration,
	)
	// Возвращаем ответ и ошибку
	return resp, nil
}

// UnaryAuthInterceptor - interceptor для проверки токена в unary RPC запросах.
func (a *AuthMiddleware) UnaryAuthInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	// Извлекаем метаданные из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil) // Инициализируем пустые метаданные, если их нет
	}

	var accessToken string
	if authHeader, exists := md["authorization"]; exists && len(authHeader) > 0 {
		accessToken = authHeader[0]
	}

	// Проверяем токен
	userID, err := a.authService.VerifyUser(accessToken)

	if err != nil {
		// Генерируем новый токен и пользователя
		userID = uuid.New().String()
		newToken, err := a.authService.CreatTokenForUser(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to generate auth token")
		}

		// Добавляем новый токен в исходящие метаданные
		newMD := metadata.Pairs("authorization", newToken)
		ctx = metadata.NewOutgoingContext(ctx, newMD)
		a.log.Info("Generated new user and token")
	} else {
		a.log.Info("Verified user")
	}

	// Передаем userID в контекст
	ctxWithUser := context.WithValue(ctx, UserIDContextKey, userID)

	// Передаем управление следующему хендлеру
	return handler(ctxWithUser, req)
}

func (a *AuthMiddleware) UnaryCheckAuthInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	var accessToken string

	// Читаем токен из метаданных (например, "authorization")
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		authHeaders := md.Get("authorization")
		if len(authHeaders) > 0 {
			accessToken = authHeaders[0]
		}
	}

	// Если токен отсутствует, возвращаем ошибку
	if accessToken == "" {
		return nil, status.Errorf(codes.Unauthenticated, "access token is missing")
	}

	// Проверяем токен с помощью AuthService
	userID, err := a.authService.VerifyUser(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	// Добавляем userID в контекст
	ctxWithUser := context.WithValue(ctx, UserIDContextKey, userID)

	// Передаем управление следующему interceptor или хендлеру
	return handler(ctxWithUser, req)
}
