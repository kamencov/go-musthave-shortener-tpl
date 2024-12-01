package handlers

import (
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	pd "github.com/kamencov/go-musthave-shortener-tpl/internal/proto/proto"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serviceGRPC interface {
	SaveURL(url, userID string) (string, error)
	SaveSliceOfDB(urls []models.MultipleURL, baseURL, userID string) ([]models.ResultMultipleURL, error)
	GetURL(shortURL string) (string, error)
}

type HandlersRPC struct {
	pd.UnimplementedShortenerServer
	service        serviceGRPC
	baseURL        string
	log            *logger.Logger
	worker         workers.Worker
	trustedSubnets string
}

func NewHandlersRPC(service serviceGRPC, baseURL string,
	log *logger.Logger, worker workers.Worker,
	trustedSubnets string) *HandlersRPC {
	return &HandlersRPC{
		service:        service,
		baseURL:        baseURL,
		log:            log,
		worker:         worker,
		trustedSubnets: trustedSubnets,
	}
}

func (h *HandlersRPC) PostJSON(ctx context.Context, req *pd.PostJSONRequest) (*pd.PostJSONResponse, error) {
	// проверяем есть ли ссылка
	if req.GetUrl() == "" {
		h.log.Error("Received empty URL")
		return nil, status.Errorf(codes.InvalidArgument, "URL cannot be empty")
	}

	userID := ctx.Value(middleware.UserIDContextKey).(string)

	shortURL, err := h.service.SaveURL(req.GetUrl(), userID)
	if err != nil {
		if errors.Is(err, errorscustom.ErrConflict) {
			h.log.Error("URL already exists", logger.StringAttr("url", req.GetUrl()))
			return &pd.PostJSONResponse{ShortUrl: h.resultBody(shortURL)}, status.Errorf(codes.AlreadyExists, "URL already exists")
		}
		h.log.Error("Internal server error", logger.ErrAttr(err))
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Возвращаем ответ с короткой ссылкой
	return &pd.PostJSONResponse{ShortUrl: h.resultBody(shortURL)}, nil
}

// ResultBody собирает ссылку для возврата в body ответа.
func (h *HandlersRPC) resultBody(res string) string {
	return h.baseURL + "/" + res
}

func (h *HandlersRPC) PostURL(ctx context.Context, req *pd.PostURLRequest) (*pd.PostURLResponse, error) {
	// проверяем есть ли ссылка
	if req.GetUrl() == "" {
		h.log.Error("Received empty URL")
		return nil, status.Errorf(codes.InvalidArgument, "URL cannot be empty")
	}

	userID := ctx.Value(middleware.UserIDContextKey).(string)

	// создаем короткую ссылку
	encodeURL, err := h.service.SaveURL(req.GetUrl(), userID)
	if err != nil {
		if errors.Is(err, errorscustom.ErrConflict) {
			h.log.Error("URL already exists", logger.StringAttr("url", req.GetUrl()))
			return &pd.PostURLResponse{ShortUrl: encodeURL}, status.Errorf(codes.AlreadyExists, "URL already exists")
		}

		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	// Возвращаем ответ с короткой ссылкой
	return &pd.PostURLResponse{ShortUrl: h.resultBody(encodeURL)}, nil
}

func (h *HandlersRPC) PostBatchDB(ctx context.Context, req *pd.PostBatchDBRequest) (*pd.PostBatchDBResponse, error) {
	var multipleURL []models.MultipleURL

	body := req.GetUrls()
	// проверяем есть ли ссылка
	if cap(body) == 0 {
		h.log.Error("Received empty URL")
		return nil, status.Errorf(codes.InvalidArgument, "URL cannot be empty")
	}

	userID := ctx.Value(middleware.UserIDContextKey).(string)

	// Записываем в пустую структуру полученный запрос
	for _, url := range body {
		multipleURL = append(multipleURL, models.MultipleURL{CorrelationID: url.CorrelationId, OriginalURL: url.OriginalUrl})
	}

	resultMultipleURL, err := h.service.SaveSliceOfDB(multipleURL, h.baseURL, userID)
	if err != nil {
		h.log.Error("Error shorten URL = ", logger.ErrAttr(err))
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	convertedResults := make([]*pd.ResultMultipleURL, 0, len(resultMultipleURL)) // Предварительная аллокация памяти
	for _, v := range resultMultipleURL {
		convertedResults = append(convertedResults, &pd.ResultMultipleURL{
			CorrelationId: v.CorrelationID,
			ShortUrl:      v.ShortURL,
		})
	}

	return &pd.PostBatchDBResponse{Results: convertedResults}, nil
}

func (h *HandlersRPC) GetURL(ctx context.Context, req *pd.GetURLRequest) (*pd.GetURLResponse, error) {
	shortURL := req.GetShortUrl()

	//проверяем на пустой запрос
	if shortURL == "" {
		h.log.Error("Not have short URL")
		return nil, status.Errorf(codes.Unimplemented, "Please provide correct short URL")
	}

	//ищем в мапе сохраненный url
	url, err := h.service.GetURL(shortURL)

	if err != nil {
		if errors.Is(err, errorscustom.ErrDeletedURL) {
			h.log.Error("URL deleted")
			return nil, status.Errorf(codes.FailedPrecondition, "URL Deleted")
		}

		h.log.Error("URL not found")
		return nil, status.Errorf(codes.NotFound, "URL not found")
	}

	return &pd.GetURLResponse{OriginalUrl: url}, nil
}
