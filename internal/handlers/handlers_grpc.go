package handlers

import (
	"database/sql"
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/middleware"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	pd "github.com/kamencov/go-musthave-shortener-tpl/internal/proto/proto"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/workers"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"strings"
)

type serviceGRPC interface {
	SaveURL(url, userID string) (string, error)
	SaveSliceOfDB(urls []models.MultipleURL, baseURL, userID string) ([]models.ResultMultipleURL, error)
	GetURL(shortURL string) (string, error)
	GetAllURL(userID, baseURL string) ([]*models.UserURLs, error)
	GetCountURLsAndUsers() (int, int, error)
	Ping() error
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
			h.log.Error("URL deleted", logger.ErrAttr(err))
			return nil, status.Errorf(codes.FailedPrecondition, "URL Deleted")
		}

		h.log.Error("URL not found", logger.ErrAttr(err))
		return nil, status.Errorf(codes.NotFound, "URL not found")
	}

	return &pd.GetURLResponse{OriginalUrl: url}, nil
}

func (h *HandlersRPC) GetPing(context.Context, *pd.Empty) (*pd.GetPingResponse, error) {
	if err := h.service.Ping(); err != nil {
		h.log.Error("Error get ping", logger.ErrAttr(err))
		return nil, status.Errorf(codes.Internal, "Error get ping")
	}

	return &pd.GetPingResponse{Status: "OK"}, nil
}

func (h *HandlersRPC) GetUsersURLs(ctx context.Context, req *pd.GetUsersURLsRequest) (*pd.GetUsersURLsResponse, error) {
	userID := ctx.Value(middleware.UserIDContextKey).(string)

	// Получаем список URL пользователя
	listURLs, err := h.service.GetAllURL(userID, h.baseURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.log.Error("Error no rows")
			return nil, status.Errorf(codes.NotFound, "no found urls")
		}

		h.log.Error("bad request")
		return nil, status.Errorf(codes.InvalidArgument, "bad request")
	}

	if len(listURLs) == 0 {
		h.log.Error("The list is empty")
		return nil, status.Errorf(codes.NotFound, "The list is empty")
	}

	convertedResults := make([]*pd.UserURL, 0, len(listURLs)) // Предварительная аллокация памяти
	for _, v := range listURLs {
		convertedResults = append(convertedResults, &pd.UserURL{
			ShortUrl:    v.ShortURL,
			OriginalUrl: v.OriginalURL,
		})
	}

	return &pd.GetUsersURLsResponse{Urls: convertedResults}, nil
}

func (h *HandlersRPC) DeletionURLs(ctx context.Context, req *pd.DeletionRequest) (*pd.DeletionResponse, error) {
	urls := req.GetUrls()

	if cap(urls) == 0 {
		h.log.Error("Bad request")
		return nil, status.Errorf(codes.InvalidArgument, "Bad request")
	}

	// получаем из контектса userID
	userID := ctx.Value(middleware.UserIDContextKey).(string)

	worker := workers.DeletionRequest{
		User: userID,
		URLs: urls,
	}

	if err := h.worker.SendDeletionRequestToWorker(worker); err != nil {
		h.log.Error("error send to deletion worker request", "error = ", err)
		return nil, status.Errorf(codes.Internal, "problem with deletion")
	}

	return &pd.DeletionResponse{Status: "OK"}, nil
}

func (h *HandlersRPC) GetStatus(ctx context.Context, req *pd.Empty) (*pd.GetStatusResponse, error) {

	// проверяем есть ли доверительный IP
	if h.trustedSubnets == "" {
		h.log.Error("no trusted ip")
		return nil, status.Errorf(codes.PermissionDenied, "no trusted ip")
	}

	// проверяем доверительный IP
	err := checkIPgRPC(ctx, h.trustedSubnets)
	if err != nil {
		h.log.Error("error = ", logger.ErrAttr(err))
		return nil, status.Errorf(codes.PermissionDenied, "no trust in IP")
	}

	// получаем сумму всех urls и users в базе
	countURLs, countUsers, err := h.service.GetCountURLsAndUsers()

	if err != nil {
		h.log.Error("error = ", logger.ErrAttr(err))
		return nil, status.Errorf(codes.Internal, "error service")
	}

	return &pd.GetStatusResponse{Urls: int32(countURLs), Users: int32(countUsers)}, nil
}

func checkIPgRPC(ctx context.Context, ts string) error {
	// Извлечение метаданных из контекста
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errorscustom.ErrIPNotParse // Метаданные отсутствуют
	}

	// Извлечение IP из заголовка X-Real-IP
	ipHeaders := md.Get("x-real-ip")
	if len(ipHeaders) == 0 {
		return errorscustom.ErrIPNotParse // IP отсутствует
	}
	ipStr := ipHeaders[0]

	// Парсим IP
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errorscustom.ErrIPNotParse // Некорректный IP
	}

	// Проверяем, является ли ts подсетью
	if _, ipNet, err := net.ParseCIDR(ts); err == nil {
		if ipNet.Contains(ip) {
			return nil // IP входит в подсеть
		}
	}

	// Если ts не является подсетью, предполагаем, что это список IP
	ipList := strings.Split(ts, ",")
	for _, validIP := range ipList {
		if _, ipNet, err := net.ParseCIDR(validIP); err == nil {
			if ipNet.Contains(ip) {
				return nil // IP входит в подсеть
			}
		}
	}

	// IP не найден в списке
	return errorscustom.ErrIPNotAllowed
}
