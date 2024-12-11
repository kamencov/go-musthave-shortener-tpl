package handlers

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	logger2 "github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/proto/proto"
	service2 "github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"testing"
)

func TestNewHandlersRPC(t *testing.T) {
	// Моки зависимостей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	log := logger2.NewLogger()
	mockService := service2.NewService(mockStorage, log)
	mockWorker := workers.NewMockWorker(ctrl)

	// Тестовые данные
	baseURL := "http://localhost:8080"
	trustedSubnets := "192.168.1.0/24"

	// Вызываем тестируемую функцию
	handlersRPC := NewHandlersRPC(mockService, baseURL, log, mockWorker, trustedSubnets)

	// Проверяем результат
	if handlersRPC == nil {
		t.Fatal("Expected non-nil handlersRPC")
	}
}

func TestHandlersRPC_PostJSON(t *testing.T) {
	cases := []struct {
		name             string
		url              string
		shortURL         string
		expevtedCheckErr error
		expectedSaveErr  error
		expectedCode     codes.Code
	}{
		{
			name:             "successful",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  nil,
			expectedCode:     codes.OK,
		},
		{
			name:         "bad_url",
			url:          "",
			expectedCode: codes.InvalidArgument,
		},
		{
			name:             "url_already_exists",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: errorscustom.ErrConflict,
			expectedSaveErr:  errorscustom.ErrConflict,
			expectedCode:     codes.AlreadyExists,
		},
		{
			name:             "error_server",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  sql.ErrNoRows,
			expectedCode:     codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}

			storageMock.EXPECT().CheckURL(cc.url).Return(cc.shortURL, cc.expevtedCheckErr).AnyTimes()
			storageMock.EXPECT().SaveURL(gomock.Any(), cc.url, "test").Return(cc.expectedSaveErr).AnyTimes()

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			req := &proto.PostJSONRequest{Url: cc.url}
			_, err := handler.PostJSON(ctx, req)
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}

			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_PostURL(t *testing.T) {
	cases := []struct {
		name             string
		url              string
		shortURL         string
		expevtedCheckErr error
		expectedSaveErr  error
		expectedCode     codes.Code
	}{
		{
			name:             "successful",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  nil,
			expectedCode:     codes.OK,
		},
		{
			name:         "bad_url",
			url:          "",
			expectedCode: codes.InvalidArgument,
		},
		{
			name:             "url_already_exists",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: errorscustom.ErrConflict,
			expectedSaveErr:  errorscustom.ErrConflict,
			expectedCode:     codes.AlreadyExists,
		},
		{
			name:             "error_server",
			url:              "https://ya.ru",
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  sql.ErrNoRows,
			expectedCode:     codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}

			storageMock.EXPECT().CheckURL(cc.url).Return(cc.shortURL, cc.expevtedCheckErr).AnyTimes()
			storageMock.EXPECT().SaveURL(gomock.Any(), cc.url, "test").Return(cc.expectedSaveErr).AnyTimes()

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			req := &proto.PostURLRequest{Url: cc.url}
			_, err := handler.PostURL(ctx, req)
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}

			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_PostBatchDB(t *testing.T) {
	cases := []struct {
		name                 string
		urls                 []models.MultipleURL
		resultMultipleURL    []models.ResultMultipleURL
		expectedSaveSliceErr error
		expectedCode         codes.Code
	}{
		{
			name: "successful",
			urls: []models.MultipleURL{
				{
					CorrelationID: "test",
					OriginalURL:   "https://ya.ru",
				},
			},
			resultMultipleURL: []models.ResultMultipleURL{
				{
					CorrelationID: "test",
					ShortURL:      "127.0.0.1/test",
				},
			},
			expectedSaveSliceErr: nil,
			expectedCode:         codes.OK,
		},
		{
			name:         "bad_request",
			urls:         []models.MultipleURL{},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "error_server",
			urls: []models.MultipleURL{
				{
					CorrelationID: "test",
					OriginalURL:   "https://ya.ru",
				},
			},
			expectedSaveSliceErr: sql.ErrNoRows,
			expectedCode:         codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}
			storageMock.EXPECT().
				SaveSlice(gomock.Any(), "127.0.0.1", "test").
				Return(cc.resultMultipleURL, cc.expectedSaveSliceErr).
				AnyTimes()

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			req := &proto.PostBatchDBRequest{Urls: []*proto.MultipleURL{}}
			for _, url := range cc.urls {
				// Convert models.MultipleURL to proto.MultipleURL
				protoURL := &proto.MultipleURL{
					CorrelationId: url.CorrelationID,
					OriginalUrl:   url.OriginalURL,
				}
				req.Urls = append(req.Urls, protoURL)
			}

			_, err := handler.PostBatchDB(ctx, req)
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_GetURL(t *testing.T) {
	cases := []struct {
		name         string
		shortURL     string
		url          string
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name:         "successful",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  nil,
			expectedCode: codes.OK,
		},
		{
			name:         "bad_request",
			shortURL:     "",
			expectedCode: codes.Unimplemented,
		},
		{
			name:         "url_deleted",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  errorscustom.ErrDeletedURL,
			expectedCode: codes.FailedPrecondition,
		},
		{
			name:         "url_not_found",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  sql.ErrNoRows,
			expectedCode: codes.NotFound,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}

			storageMock.EXPECT().GetURL(cc.shortURL).
				Return(cc.url, cc.expectedErr).
				AnyTimes()

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			req := &proto.GetURLRequest{ShortUrl: cc.shortURL}

			_, err := handler.GetURL(ctx, req)

			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_GetPing(t *testing.T) {
	cases := []struct {
		name         string
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name:         "successful",
			expectedErr:  nil,
			expectedCode: codes.OK,
		},
		{
			name:         "successful",
			expectedErr:  errorscustom.ErrConflict,
			expectedCode: codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}

			storageMock.EXPECT().
				Ping().
				Return(cc.expectedErr).
				AnyTimes()

			_, err := handler.GetPing(context.Background(), &proto.Empty{})
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}

}

func TestHandlersRPC_GetUsersURLs(t *testing.T) {
	cases := []struct {
		name         string
		user         string
		result       []*models.UserURLs
		expectedErr  error
		expectedCode codes.Code
	}{
		{
			name: "successful",
			user: "test",
			result: []*models.UserURLs{
				{
					ShortURL:    "test",
					OriginalURL: "www.test.ru",
				},
			},
			expectedErr:  nil,
			expectedCode: codes.OK,
		},
		{
			name: "no_rows",
			user: "test",
			result: []*models.UserURLs{
				{
					ShortURL:    "test",
					OriginalURL: "www.test.ru",
				},
			},
			expectedErr:  sql.ErrNoRows,
			expectedCode: codes.NotFound,
		},
		{
			name: "bad_request",
			user: "test",
			result: []*models.UserURLs{
				{
					ShortURL:    "test",
					OriginalURL: "www.test.ru",
				},
			},
			expectedErr:  errorscustom.ErrConflict,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "bad_request",
			user:         "test",
			result:       []*models.UserURLs{},
			expectedErr:  nil,
			expectedCode: codes.NotFound,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: "",
			}

			storageMock.EXPECT().
				GetAllURL(cc.user, gomock.Any()).
				Return(cc.result, cc.expectedErr).
				AnyTimes()

			req := &proto.GetUsersURLsRequest{
				UserId: cc.user,
			}

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			_, err := handler.GetUsersURLs(ctx, req)

			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_DeletionURLs(t *testing.T) {
	cases := []struct {
		name              string
		urls              []string
		expectedWorkerErr error
		expectedCode      codes.Code
	}{
		{
			name:              "successful",
			urls:              []string{"http://example.com", "http://example2.com"},
			expectedWorkerErr: nil,
			expectedCode:      codes.OK,
		},
		{
			name:         "bad_request",
			urls:         []string{},
			expectedCode: codes.InvalidArgument,
		},
		{
			name:              "problem_worker",
			urls:              []string{"http://example.com", "http://example2.com"},
			expectedWorkerErr: errors.New("invalid_worker"),
			expectedCode:      codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storageMock := mocks.NewMockStorage(ctrl)
			workerMock := workers.NewMockWorker(ctrl)

			logger := logger2.NewLogger()

			service := service2.NewService(storageMock, logger)

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         workerMock,
				trustedSubnets: "",
			}

			workerMock.EXPECT().
				SendDeletionRequestToWorker(gomock.Any()).
				Return(cc.expectedWorkerErr).
				AnyTimes()

			req := &proto.DeletionRequest{
				UserId: "",
				Urls:   cc.urls,
			}

			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")

			_, err := handler.DeletionURLs(ctx, req)
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}

func TestHandlersRPC_GetStatus(t *testing.T) {
	cases := []struct {
		name                     string
		trustedSubnet            string
		ctx                      bool
		header                   string
		count                    int
		expectedErrGetCountURLs  error
		expectedErrGetCountUsers error
		expectedCode             codes.Code
	}{
		{
			name:          "successful",
			trustedSubnet: "192.168.1.0/24",
			header:        "192.168.1.5",
			count:         5,
			expectedCode:  codes.OK,
		},
		{
			name:          "not_use_trusted_subnet",
			trustedSubnet: "",
			expectedCode:  codes.PermissionDenied,
		},
		{
			name:          "bad context",
			trustedSubnet: "192.168.1.0/24",
			ctx:           true,
			expectedCode:  codes.PermissionDenied,
		},
		{
			name:          "not_have_header",
			trustedSubnet: "192.168.1.0/24",
			expectedCode:  codes.PermissionDenied,
		},
		{
			name:          "use_not_correct_header",
			trustedSubnet: "192.168.1.0/24",
			header:        "192.168.",
			expectedCode:  codes.PermissionDenied,
		},
		{
			name:          "use_not_correct_trusted_subnet",
			trustedSubnet: "192.168.1/24",
			header:        "192.168.1.5",
			expectedCode:  codes.PermissionDenied,
		},
		{
			name:          "use_two_trusted_subnets",
			trustedSubnet: "192.168.1.0/24,192.168.2.0/24",
			header:        "192.168.1.5",
			count:         5,
			expectedCode:  codes.OK,
		},
		{
			name:                     "bad_get_count_urls_and_users",
			trustedSubnet:            "192.168.1.0/24",
			header:                   "192.168.1.5",
			count:                    5,
			expectedErrGetCountUsers: errorscustom.ErrConflict,
			expectedCode:             codes.Internal,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storageMock := mocks.NewMockStorage(ctrl)
			logger := logger2.NewLogger()
			service := service2.NewService(storageMock, logger)
			storageMock.EXPECT().GetCountURLs().Return(cc.count, cc.expectedErrGetCountURLs).AnyTimes()
			storageMock.EXPECT().GetCountUsers().Return(cc.count, cc.expectedErrGetCountUsers).AnyTimes()

			handler := &HandlersRPC{
				service:        service,
				baseURL:        "127.0.0.1",
				log:            logger,
				worker:         nil,
				trustedSubnets: cc.trustedSubnet,
			}

			ctx := context.Background()

			if !cc.ctx {
				md := metadata.Pairs("x-real-ip", cc.header)
				ctx = metadata.NewIncomingContext(context.Background(), md)
			}

			_, err := handler.GetStatus(ctx, nil)
			if err != nil {
				code, ok := status.FromError(err)
				if !ok {
					t.Errorf("unexpected error type: %v", err)
				}
				if code.Code() != cc.expectedCode {
					t.Errorf("unexpected error code: got %v, want %v", code.Code(), cc.expectedCode)
				}
			} else if cc.expectedCode != codes.OK {
				t.Errorf("expected error code %v, got none", cc.expectedCode)
			}
		})
	}
}
