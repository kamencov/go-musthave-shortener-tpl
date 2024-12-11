package middleware

import (
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubnetCheck_Middleware(t *testing.T) {
	cases := []struct {
		name          string
		trustedSubnet string
		header        string
		expectedCode  int
	}{
		{
			name:          "successful",
			trustedSubnet: "192.168.1.0/24",
			header:        "192.168.1.5",
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			log := logger.NewLogger()
			subnetCheck := NewSubnetCheck(tt.trustedSubnet, log)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-Real-IP", tt.header)
			w := httptest.NewRecorder()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			subnetMidlware := subnetCheck.Middleware(handler)
			subnetMidlware.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected code %v, got %v", tt.expectedCode, w.Code)
			}
		})
	}
}
