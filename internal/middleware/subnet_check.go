package middleware

import (
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"net"
	"net/http"
	"strings"
)

type SubnetCheck struct {
	Subnet string
	log    *logger.Logger
}

func NewSubnetCheck(subnet string, log *logger.Logger) *SubnetCheck {
	return &SubnetCheck{
		Subnet: subnet,
		log:    log,
	}
}

func (s *SubnetCheck) Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// проверяем есть ли доверительный IP
		if s.Subnet == "" {
			s.log.Error("failed, subnet is empty")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// проверяем доверительный IP
		err := checkIP(r, s.Subnet)
		if err != nil {
			s.log.Error("error = ", logger.ErrAttr(err))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// checkIP проверяем IP на валидность
func checkIP(r *http.Request, ts string) error {
	// Извлечение IP из заголовка
	ipStr := r.Header.Get("X-Real-IP")
	if ipStr == "" {
		return errorscustom.ErrIPNotParse // IP отсутствует
	}

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
