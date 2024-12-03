package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"
)

// Структура для имитации ошибки чтения
type errorReader struct {
}

// Метод Read возвращает ошибку
func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("эмулированная ошибка чтения")
}

// Закрытие errorReader
func (e *errorReader) Close() error {
	return nil
}

// Предполагаем, что функция EncodeURL и переменная MapStorage уже определены в вашем пакете
func TestNewHandlers(t *testing.T) {
	// Моки зависимостей
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStorage := mocks.NewMockStorage(ctrl)
	log := logger.NewLogger()
	mockService := service.NewService(mockStorage, log)
	mockWorker := workers.NewMockWorker(ctrl)

	// Тестовые данные
	baseURL := "http://localhost:8080"
	trustedSubnets := "192.168.1.0/24"

	// Вызываем тестируемую функцию
	handlersRPC := NewHandlers(mockService, baseURL, log, mockWorker, trustedSubnets)

	// Проверяем результат
	if handlersRPC == nil {
		t.Fatal("Expected non-nil handlersRPC")
	}
}

func TestHandlers_PostJSON(t *testing.T) {
	cases := []struct {
		name             string
		body             io.Reader
		shortURL         string
		expevtedCheckErr error
		expectedSaveErr  error
		expectedCode     int
	}{
		{
			name:             "successful",
			body:             bytes.NewBuffer([]byte(`{"url": "https://ya.ru"}`)),
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  nil,
			expectedCode:     http.StatusCreated,
		},
		{
			name:         "bad_url",
			body:         io.NopCloser(&errorReader{}),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "no_body",
			body:         bytes.NewBuffer(nil),
			expectedCode: http.StatusNotFound,
		},
		{
			name:             "url_already_exists",
			body:             bytes.NewBuffer([]byte(`{"url": "https://ya.ru"}`)),
			shortURL:         "test",
			expevtedCheckErr: errorscustom.ErrConflict,
			expectedSaveErr:  errorscustom.ErrConflict,
			expectedCode:     http.StatusConflict,
		},
		{
			name:             "error_server",
			body:             bytes.NewBuffer([]byte(`{"url": "https://ya.ru"}`)),
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  sql.ErrNoRows,
			expectedCode:     http.StatusInternalServerError,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			logs := logger.NewLogger(logger.WithLevel("info"))
			storageMock := mocks.NewMockStorage(ctrl)
			storageMock.EXPECT().CheckURL(gomock.Any()).Return(cc.shortURL, cc.expevtedCheckErr).AnyTimes()
			storageMock.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(cc.expectedSaveErr).AnyTimes()

			serv := service.NewService(storageMock, logs)
			handlers := NewHandlers(serv, "http://localhost:8080", logs, nil, "")

			req := httptest.NewRequest("POST", "/", cc.body)
			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			handlers.PostJSON(w, req)

			if w.Code != cc.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", cc.expectedCode, w.Code)
			}

		})

	}
}

func TestHandlers_PostURL(t *testing.T) {
	cases := []struct {
		name             string
		url              io.Reader
		shortURL         string
		expevtedCheckErr error
		expectedSaveErr  error
		expectedCode     int
	}{
		{
			name:             "successful",
			url:              bytes.NewBuffer([]byte("https://ya.ru")),
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  nil,
			expectedCode:     http.StatusCreated,
		},
		{
			name:         "bad_url",
			url:          io.NopCloser(&errorReader{}),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "no_body",
			url:          bytes.NewBuffer(nil),
			expectedCode: http.StatusNotFound,
		},
		{
			name:             "url_already_exists",
			url:              bytes.NewBuffer([]byte("https://ya.ru")),
			shortURL:         "test",
			expevtedCheckErr: errorscustom.ErrConflict,
			expectedSaveErr:  errorscustom.ErrConflict,
			expectedCode:     http.StatusConflict,
		},
		{
			name:             "error_server",
			url:              bytes.NewBuffer([]byte("https://ya.ru")),
			shortURL:         "test",
			expevtedCheckErr: nil,
			expectedSaveErr:  sql.ErrNoRows,
			expectedCode:     http.StatusInternalServerError,
		},
	}
	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			log := logger.NewLogger()

			serv := service.NewService(storageMock, log)

			storageMock.EXPECT().CheckURL(gomock.Any()).Return(cc.shortURL, cc.expevtedCheckErr).AnyTimes()
			storageMock.EXPECT().SaveURL(gomock.Any(), gomock.Any(), "test").Return(cc.expectedSaveErr).AnyTimes()

			handlers := NewHandlers(serv, "http://localhost:8080", log, nil, "")

			req := httptest.NewRequest("POST", "/", cc.url)
			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handlers.PostURL(w, req)

			if w.Code != cc.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", cc.expectedCode, w.Code)
			}
		})
	}
}

func TestHandlers_GetURL(t *testing.T) {
	cases := []struct {
		name         string
		shortURL     string
		url          string
		expectedErr  error
		expectedCode int
	}{
		{
			name:         "successful",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  nil,
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "bad_request",
			shortURL:     "",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "url_deleted",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  errorscustom.ErrDeletedURL,
			expectedCode: http.StatusGone,
		},
		{
			name:         "url_not_found",
			shortURL:     "test",
			url:          "test.ru",
			expectedErr:  sql.ErrNoRows,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			log := logger.NewLogger()

			serv := service.NewService(storageMock, log)
			storageMock.EXPECT().GetURL(cc.shortURL).
				Return(cc.url, cc.expectedErr).
				AnyTimes()

			handler := NewHandlers(serv, "http://localhost:8080", log, nil, "")

			req := httptest.NewRequest("GET", "/", nil)
			ctx := context.WithValue(context.Background(), middleware.UserIDContextKey, "test")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			chiCtx := chi.NewRouteContext()
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
			chiCtx.URLParams.Add("id", cc.shortURL)

			handler.GetURL(w, req)

			if w.Code != cc.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", cc.expectedCode, w.Code)
			}
		})
	}
}

func TestHandlers_GetPing(t *testing.T) {
	cases := []struct {
		name         string
		expectedErr  error
		expectedCode int
	}{
		{
			name:         "successful",
			expectedErr:  nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "successful",
			expectedErr:  errorscustom.ErrConflict,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			storageMock := mocks.NewMockStorage(ctrl)
			log := logger.NewLogger()

			serv := service.NewService(storageMock, log)

			handler := NewHandlers(serv, "http://localhost:8080", log, nil, "")

			storageMock.EXPECT().
				Ping().
				Return(cc.expectedErr).
				AnyTimes()

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			handler.GetPing(w, req)
			if w.Code != cc.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", cc.expectedCode, w.Code)
			}
		})
	}
}

func TestPostBatchDB_Success(t *testing.T) {
	testURL := "https://youtube.com"

	multipleURL := models.MultipleURL{
		CorrelationID: "1",
		OriginalURL:   testURL,
	}

	reseltMultip := []models.ResultMultipleURL{}

	marsh, err := json.Marshal([]models.MultipleURL{multipleURL})
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewReader(marsh)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgre := mocks.NewMockStorage(ctrl)
	mockPostgre.EXPECT().SaveSlice(gomock.Any(), gomock.Any(), gomock.Any()).Return(reseltMultip, nil)
	logger := logger.NewLogger(logger.WithLevel("info"))

	service := service.NewService(mockPostgre, logger)
	handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

	req, err := http.NewRequest("POST", "/api/shorten/batch", buf)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PostBatchDB(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("ожидался статус %d, но получен %d", http.StatusCreated, w.Code)
	}
}

func TestPostBatchDB_InCorrectRequest(t *testing.T) {

	buf := bytes.NewReader([]byte("lololo"))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgre := mocks.NewMockStorage(ctrl)

	logger := logger.NewLogger(logger.WithLevel("info"))

	service := service.NewService(mockPostgre, logger)
	handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

	req, err := http.NewRequest("POST", "/api/shorten/batch", buf)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PostBatchDB(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("ожидался статус %d, но получен %d", http.StatusBadRequest, w.Code)
	}
}

func TestPostBatchDB_StorageError(t *testing.T) {
	testURL := "https://youtube.com"

	reseltMultip := []models.ResultMultipleURL{}

	multipleURL := models.MultipleURL{
		CorrelationID: "1",
		OriginalURL:   testURL,
	}

	marsh, err := json.Marshal([]models.MultipleURL{multipleURL})
	if err != nil {
		t.Fatal(err)
	}
	buf := bytes.NewReader(marsh)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgre := mocks.NewMockStorage(ctrl)
	mockErr := errors.New("Some error")
	mockPostgre.EXPECT().SaveSlice(gomock.Any(), gomock.Any(), gomock.Any()).Return(reseltMultip, mockErr)

	logger := logger.NewLogger(logger.WithLevel("info"))

	service := service.NewService(mockPostgre, logger)
	handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

	req, err := http.NewRequest("POST", "/api/shorten/batch", buf)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PostBatchDB(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("ожидался статус %d, но получен %d", http.StatusInternalServerError, w.Code)
	}
}

func TestPostBatchDB_EmptyRequest(t *testing.T) {

	buf := bytes.NewReader([]byte(""))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgre := mocks.NewMockStorage(ctrl)

	logger := logger.NewLogger(logger.WithLevel("info"))

	service := service.NewService(mockPostgre, logger)
	handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

	req, err := http.NewRequest("POST", "/api/shorten/batch", buf)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handlers.PostBatchDB(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("ожидался статус %d, но получен %d", http.StatusNotFound, w.Code)
	}
}

func TestHandlers_GetUsersURLs(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode int
		userURLs     []*models.UserURLs
		expectedErr  error
		ctx          bool
	}{
		{
			name: "successful_get",
			userURLs: []*models.UserURLs{
				{
					ShortURL:    "test",
					OriginalURL: "original_test",
				},
			},
			expectedCode: 200,
		},
		{
			name:         "bad_request_context",
			ctx:          true,
			expectedCode: 401,
		},
		{
			name:         "no_found_urls",
			expectedErr:  sql.ErrNoRows,
			expectedCode: 204,
		},
		{
			name:         "bad_get_all_url",
			expectedErr:  errors.New("error_not_sql"),
			expectedCode: 400,
		},
		{
			name:         "len_0",
			expectedCode: 204,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPostgre := mocks.NewMockStorage(ctrl)

			logger := logger.NewLogger(logger.WithLevel("info"))

			service := service.NewService(mockPostgre, logger)

			handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

			req := httptest.NewRequest("GET", "/", nil)

			if !tt.ctx {
				ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
				req = req.WithContext(ctx)
			}

			mockPostgre.EXPECT().GetAllURL(gomock.Any(), gomock.Any()).Return(tt.userURLs, tt.expectedErr).AnyTimes()

			resp := httptest.NewRecorder()

			handlers.GetUsersURLs(resp, req)

			if resp.Code != tt.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", tt.expectedCode, resp.Code)
			}

		})
	}
}

func TestHandlers_DeletionURLs(t *testing.T) {
	tests := []struct {
		name              string
		body              string
		expectedCode      int
		expectedWorkerErr error
		ctx               bool
	}{
		{
			name:         "successful",
			body:         `["http://example.com", "http://example2.com"]`,
			expectedCode: 202,
		},
		{
			name:         "invalid_body",
			expectedCode: 500,
		},
		{
			name:              "invalid_worker",
			body:              `["http://example.com", "http://example2.com"]`,
			expectedWorkerErr: errors.New("invalid_worker"),
			expectedCode:      500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// создаем logger.
			loger := logger.NewLogger()

			// создаем заглушку базы.
			dbMock := mocks.NewMockStorage(ctrl)

			// создаем сервис.
			service := service.NewService(dbMock, loger)

			// создаем заглушку worker.
			workerMock := workers.NewMockWorker(ctrl)
			workerMock.EXPECT().SendDeletionRequestToWorker(gomock.Any()).Return(tt.expectedWorkerErr).AnyTimes()

			// создаем запрос.
			req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewBuffer([]byte(tt.body)))

			resp := httptest.NewRecorder()

			// создаем контекст авторизации.
			if !tt.ctx {
				ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
				req = req.WithContext(ctx)
			}

			// собираем handler.
			handler := &Handlers{
				service: service,
				baseURL: "http://localhost:8080/",
				logger:  loger,
				worker:  workerMock,
			}

			handler.DeletionURLs(resp, req)

			if resp.Code != tt.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", tt.expectedCode, resp.Code)
			}
		})
	}
}

func TestHandlers_GetStatus(t *testing.T) {
	cases := []struct {
		name                     string
		trustedSubnet            string
		header                   string
		count                    int
		expectedErrGetCountURLs  error
		expectedErrGetCountUsers error
		expectedCode             int
	}{
		{
			name:          "successful",
			trustedSubnet: "192.168.1.0/24",
			header:        "192.168.1.5",
			count:         5,
			expectedCode:  http.StatusOK,
		},
		{
			name:          "not_use_trusted_subnet",
			trustedSubnet: "",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "not_have_header",
			trustedSubnet: "192.168.1.0/24",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "use_not_correct_header",
			trustedSubnet: "192.168.1.0/24",
			header:        "192.168.",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "use_not_correct_trusted_subnet",
			trustedSubnet: "192.168.1/24",
			header:        "192.168.1.5",
			expectedCode:  http.StatusForbidden,
		},
		{
			name:          "use_two_trusted_subnets",
			trustedSubnet: "192.168.1.0/24,192.168.2.0/24",
			header:        "192.168.1.5",
			count:         5,
			expectedCode:  http.StatusOK,
		},
		{
			name:                     "bad_get_count_urls_and_users",
			trustedSubnet:            "192.168.1.0/24",
			header:                   "192.168.1.5",
			count:                    5,
			expectedErrGetCountUsers: errorscustom.ErrConflict,
			expectedCode:             http.StatusInternalServerError,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPostgre := mocks.NewMockStorage(ctrl)
			mockPostgre.EXPECT().GetCountURLs().Return(tt.count, tt.expectedErrGetCountURLs).AnyTimes()
			mockPostgre.EXPECT().GetCountUsers().Return(tt.count, tt.expectedErrGetCountUsers).AnyTimes()
			newLogger := logger.NewLogger(logger.WithLevel("info"))
			newService := service.NewService(mockPostgre, newLogger)
			handlers := NewHandlers(newService, "http://localhost:8080/", newLogger, nil, tt.trustedSubnet)

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Real-IP", tt.header)
			resp := httptest.NewRecorder()

			handlers.GetStatus(resp, req)

			if resp.Code != tt.expectedCode {
				t.Errorf("ожидался статус %d, но получен %d", tt.expectedCode, resp.Code)
			}
		})
	}
}
