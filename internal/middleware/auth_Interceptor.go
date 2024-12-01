package middleware

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"time"
)

// UnaryInterceptor - гRPC интерсептор для логирования и проверки авторизации.
func (a *AuthMiddleware) UnaryInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

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
	} else {
		// Логируем успешный ответ
		a.log.Info(
			"gRPC request completed",
			"method", info.FullMethod,
			"response", resp,
			"duration", duration,
		)
	}

	// Возвращаем ответ и ошибку
	return resp, err
}

// UnaryAuthInterceptor - interceptor для проверки токена в unary RPC запросах.
func (a *AuthMiddleware) UnaryAuthInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

	// Извлекаем токен из метаданных
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	var accessToken string
	if authHeader, exists := md["authorization"]; exists && len(authHeader) > 0 {
		accessToken = authHeader[0]
	} else if cookieHeader, exists := md["cookie"]; exists && len(cookieHeader) > 0 {
		// Пример чтения токена из cookie
		cookies := parseCookies(cookieHeader[0])
		accessToken = cookies[string(UserIDContextKey)]
	}

	// Проверяем токен
	userID, err := a.authService.VerifyUser(accessToken)
	if err != nil {
		// Генерируем новый токен и пользователя
		userID = uuid.New().String()
		token, err := a.authService.CreatTokenForUser(userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to generate auth token")
		}

		// Добавляем новый токен в контекст
		md.Append("set-cookie", fmt.Sprintf("%s=%s; HttpOnly; Path=/", string(UserIDContextKey), token))
		md.Append("authorization", token)
		if err = grpc.SendHeader(ctx, md); err != nil {
			return nil, status.Error(codes.Internal, "failed to save header")
		}
	}

	// Передаем userID в контекст
	ctxWithUser := context.WithValue(ctx, UserIDContextKey, userID)

	// Передаем управление следующему хендлеру
	return handler(ctxWithUser, req)
}

// parseCookies парсит строку куки в map
func parseCookies(cookieHeader string) map[string]string {
	cookies := make(map[string]string)
	pairs := strings.Split(cookieHeader, "; ")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}
	return cookies
}

func (a *AuthMiddleware) UnaryCheckAuthInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {

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
