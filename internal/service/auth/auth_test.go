package auth

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"testing"
)

// NewServiceAuth - тестирует корректное создание сервиса авторизации.
func TestNewServiceAuth(t *testing.T) {
	t.Run("create_service", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storage := mocks.NewMockStorage(ctrl)
		authServ := NewServiceAuth(storage)

		if authServ == nil {
			t.Errorf("не удалось создать сервис авторизации")
		}

	})
}

// CreatTokenForUser - тестирует корректное создание токена.
func TestServiceAuth_CreatTokenForUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		expectedErr error
	}{
		{
			name:   "successful",
			userID: "testID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServ := ServiceAuth{}

			token, err := authServ.CreatTokenForUser(tt.userID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Ожидали ошибку %v, пришла ошибка %v", tt.expectedErr, err)
			}

			if token == "" {
				t.Errorf("Ожидали токен, пришел пустой")
			}

		})
	}
}

// VerifyUser - тестируем верификацию по токену.
func TestServiceAuth_VerifyUser(t *testing.T) {
	authServ := ServiceAuth{}
	token, err := authServ.CreatTokenForUser("testID")

	if err != nil {
		t.Errorf("Ожидали ошибку = nil, пришла ошибка %v", err)
	}

	tests := []struct {
		name        string
		token       string
		expectedErr error
	}{
		{
			name:        "successful",
			token:       token,
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err = authServ.VerifyUser(token)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Ожидали ошибку %v, пришла ошибка %v", tt.expectedErr, err)

			}
		})
	}
}
