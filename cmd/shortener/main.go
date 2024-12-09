package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	middleware2 "github.com/go-chi/chi/v5/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/handlers"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	pd "github.com/kamencov/go-musthave-shortener-tpl/internal/proto/proto"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service/auth"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"
	"google.golang.org/grpc"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/swaggo/http-swagger/example/go-chi/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

//@title URL Shortener API
//@version 1.0
//@description API Server for shortener

//@host localhost:8080
//@BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// BuildVersion = определяет версию приложения
// BuildDate = определяет дату сборки
// BuildCommit = определяет коммит сборки
var (
	BuildVersion = "N/A"
	BuildDate    = "N/A"
	BuildCommit  = "N/A"
)

func main() {
	// выводит глобальную информацию
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildDate)
	fmt.Printf("Build commit: %s\n", BuildCommit)

	// инициализируем конфиг.
	configs := NewConfigs()
	configs.Parse()

	// инициализируем logger.
	logs := logger.NewLogger(logger.WithLevel(configs.LogLevel))
	logs.Info("Start logger")

	// инициализируем хранилище.
	repo := initDB(configs.AddrDB, configs.PathFile)
	logs.Info("Connecting DB")
	defer func(repo service.Storage) {
		err := repo.Close()
		if err != nil {
			logs.Error("Fatal", logger.ErrAttr(err))
		}
	}(repo)

	// инициализируем сервис.
	urlService := service.NewService(
		repo,
		logs,
	)
	logs.Info("Service created")

	// инициализируем проверку авторизацию.
	serviceAuth := auth.NewServiceAuth(repo)
	authorization := middleware.NewAuthMiddleware(serviceAuth, logs)

	// инициализируем worker.
	worker := workers.NewWorkerDeleted(urlService)

	// инициализируем проверку подсети.
	subnetMiddleware := middleware.NewSubnetCheck(configs.TrustedSubnet, logs)

	// передаем в хенлер сервис и baseURL.
	shortHandlers := handlers.NewHandlers(urlService, configs.BaseURL, logs, worker)
	logs.Info(fmt.Sprintf("Handlers created PORT: %s", configs.AddrServer))

	// Создаём gRPC-хендлер
	grpcHandlers := handlers.NewHandlersRPC(urlService, configs.BaseURL, logs, worker, configs.TrustedSubnet)

	// Создаем gRPC сервер с Interceptor
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authorization.UnaryInterceptor))

	// Регистрируем gRPC-хендлер
	pd.RegisterShortenerServer(grpcServer, grpcHandlers)

	// Запускаем gRPC-сервер
	go func() {
		// Запуск сервера
		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			logs.Error("Failed to listen on port 50051", logger.ErrAttr(err))
		}
		logs.Info("gRPC server is running on port 50051...")
		if err = grpcServer.Serve(listener); err != nil {
			logs.Error("Failed to serve gRPC server", logger.ErrAttr(err))
		}
	}()

	// инициализировали роутер и создали Post и Get.
	r := chi.NewRouter()
	r.Use(middleware.WithLogging)

	// Swagger route.
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// Добавляем pprof маршруты вручную.
	// Ограничим доступ с помощью контроля ip.
	r.Group(func(r chi.Router) {
		r.Use(middleware.PprofMiddleware)
		r.Mount("/debug", middleware2.Profiler())
	})

	r.Route("/", func(r chi.Router) {
		r.Use(middleware.GZipMiddleware)
		r.Use(authorization.AuthMiddleware)
		r.Post("/", shortHandlers.PostURL)
		r.Post("/api/shorten", shortHandlers.PostJSON)
		r.Post("/api/shorten/batch", shortHandlers.PostBatchDB)
	})

	r.Get("/{id}", shortHandlers.GetURL)
	r.Get("/ping", shortHandlers.GetPing)

	r.Route("/api/internal", func(r chi.Router) {
		r.Use(subnetMiddleware.Middleware)
		r.Get("/stats", shortHandlers.GetStatus)
	})

	r.Route("/api/user/urls", func(r chi.Router) {
		r.Use(authorization.CheckAuthMiddleware)
		r.Get("/", shortHandlers.GetUsersURLs)
		r.Delete("/", shortHandlers.DeletionURLs)
	})

	// Создаем HTTP-сервер с поддержкой graceful shutdown
	server := &http.Server{
		Addr:    configs.AddrServer,
		Handler: r,
	}

	// Настройка контекста для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	go worker.StartWorkerDeletion(ctx)

	// Запускаем сервер в горутине
	go func() {
		if *configs.HTTPS {
			logs.Info("Starting HTTPS server with self-signed certificate")
			if err := server.ListenAndServeTLS("./cert.pem", "./key.pem"); !errors.Is(err, http.ErrServerClosed) {
				logs.Error("Failed to start HTTPS server:", logger.ErrAttr(err))
			}
		} else {
			logs.Info("Starting HTTP server")
			if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				logs.Error("Failed to start HTTP server:", logger.ErrAttr(err))
			}
		}
	}()

	// Обработка системных сигналов для graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logs.Info("Shutting down gracefully...")

	// Останавливаем сервер и ожидаем завершения текущих запросов
	if err := server.Shutdown(ctx); err != nil {
		logs.Error("Failed to gracefully shutdown server:", logger.ErrAttr(err))
	}

	// Отсанавливаем gRPC-сервер
	grpcServer.GracefulStop()

	cancel() // Завершаем контекст для worker

	logs.Info("Shutdown complete")
}
