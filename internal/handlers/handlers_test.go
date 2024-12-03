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
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/storage/mapstorage"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"

	"github.com/stretchr/testify/assert"
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

func TestHandlersPostJSON(t *testing.T) {
	logs := logger.NewLogger(logger.WithLevel("info"))
	storage := mapstorage.NewMapURL()

	urlService := service.NewService(storage, logs)
	shortHandlers := NewHandlers(urlService, "http://localhost:8080", logs, nil, "")

	t.Run("test_post_JSON", func(t *testing.T) {
		payload := "{\"url\": \"https://practicum.yandex.ru\"}"
		param := strings.NewReader(payload)
		rRequest := httptest.NewRequest("POST", "/", param)
		ctx := context.WithValue(rRequest.Context(), middleware.UserIDContextKey, "userID")
		rRequest = rRequest.WithContext(ctx)
		wResonse := httptest.NewRecorder()

		shortHandlers.PostJSON(wResonse, rRequest)

		// Проверяем, что статус ответа - 201 Created
		assert.Equal(t, http.StatusCreated, wResonse.Code)

	})

	t.Run("test_post_JSON_noBody", func(t *testing.T) {
		payload := ""
		param := strings.NewReader(payload)
		rRequest := httptest.NewRequest("POST", "/", param)
		ctx := context.WithValue(rRequest.Context(), middleware.UserIDContextKey, "userID")
		rRequest = rRequest.WithContext(ctx)
		wResonse := httptest.NewRecorder()

		shortHandlers.PostJSON(wResonse, rRequest)

		//проверяем пустое тело
		assert.Equal(t, http.StatusNotFound, wResonse.Code)
	})
}

func TestGetURL(t *testing.T) {
	// Тест на успешное декодирование URL
	logs := logger.NewLogger(logger.WithLevel("info"))

	storage := mapstorage.NewMapURL()

	urlService := service.NewService(storage, logs)
	shortHandlers := NewHandlers(urlService, "http://localhost:8080", logs, nil, "")
	t.Run("test_get_URL", func(t *testing.T) {

		payload := []byte("http://example.com")
		rRequest := httptest.NewRequest("POST", "/body", bytes.NewBuffer(payload))
		ctx := context.WithValue(rRequest.Context(), middleware.UserIDContextKey, "userID")
		rRequest = rRequest.WithContext(ctx)
		wResonse := httptest.NewRecorder()

		shortHandlers.PostURL(wResonse, rRequest)

		responseURL := wResonse.Body.String()
		encodedURL := strings.TrimPrefix(responseURL, "http://localhost:8080/")
		rRequest = httptest.NewRequest("GET", "http://localhost:8080/", nil)
		wResonse = httptest.NewRecorder()

		chiCtx := chi.NewRouteContext()
		rRequest = rRequest.WithContext(context.WithValue(rRequest.Context(), chi.RouteCtxKey, chiCtx))
		chiCtx.URLParams.Add("id", encodedURL)

		shortHandlers.GetURL(wResonse, rRequest)

		// Проверяем, что статус ответа - 200 OK
		assert.Equal(t, http.StatusTemporaryRedirect, wResonse.Code)

		// Проверяем, что в MapStorage добавлен новый URL
		originalURL, err := storage.GetURL(encodedURL)
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com", originalURL)

		// Проверяем, что в MapStorage нет URL
		rRequest = httptest.NewRequest("GET", "http://localhost:8080/", nil)
		chiCtx = chi.NewRouteContext()
		rRequest = rRequest.WithContext(context.WithValue(rRequest.Context(), chi.RouteCtxKey, chiCtx))
		chiCtx.URLParams.Add("id", "Nourl")
		wResonse = httptest.NewRecorder()
		shortHandlers.GetURL(wResonse, rRequest)
		assert.Equal(t, http.StatusNotFound, wResonse.Code)
	})
}

func TestGetPing(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostgre := mocks.NewMockStorage(ctrl)
	mockPostgre.EXPECT().Ping().Return(nil)

	logger := logger.NewLogger(logger.WithLevel("info"))

	service := service.NewService(mockPostgre, logger)
	handlers := &Handlers{service: service, baseURL: "http://localhost:8080/", logger: logger}

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	handlers.GetPing(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ожидался статус %d, но получен %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("ожидался заголовок %s, но получен %s", "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
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

			//создаем logger.
			loger := logger.NewLogger()

			//создаем заглушку базы.
			dbMock := mocks.NewMockStorage(ctrl)

			//создаем сервис.
			service := service.NewService(dbMock, loger)

			//создаем заглушку worker.
			workerMock := workers.NewMockWorker(ctrl)
			workerMock.EXPECT().SendDeletionRequestToWorker(gomock.Any()).Return(tt.expectedWorkerErr).AnyTimes()

			//создаем запрос.
			req := httptest.NewRequest(http.MethodDelete, "/", bytes.NewBuffer([]byte(tt.body)))

			resp := httptest.NewRecorder()

			//создаем контекст авторизации.
			if !tt.ctx {
				ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, "userID")
				req = req.WithContext(ctx)
			}

			//собираем handler.
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
