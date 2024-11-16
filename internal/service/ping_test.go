package service

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"testing"
)

func TestService_Ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mocks.NewMockStorage(ctrl)
	service := NewService(storage, logger.NewLogger(logger.WithLevel("info")))
	t.Run("ping", func(t *testing.T) {
		storage.EXPECT().Ping().Return(nil)
		err := service.Ping()
		if !errors.Is(err, nil) {
			t.Errorf("ожидался статус %v, но получен %v", nil, err)
		}
	})
}
