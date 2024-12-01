package handlers

import (
	"context"
	"database/sql"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	logger2 "github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/proto/proto"
	service2 "github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

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
				protoUrl := &proto.MultipleURL{
					CorrelationId: url.CorrelationID,
					OriginalUrl:   url.OriginalURL,
				}
				req.Urls = append(req.Urls, protoUrl)
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
