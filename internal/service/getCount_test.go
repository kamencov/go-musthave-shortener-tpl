package service

import (
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"testing"
)

func TestService_GetCountURLsAndUsers(t *testing.T) {
	cases := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "successful",
			expectedErr: nil,
		},

		{
			name:        "error",
			expectedErr: sql.ErrNoRows,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storageMock := mocks.NewMockStorage(ctrl)
			storageMock.EXPECT().GetCountUsers().Return(1, tt.expectedErr)
			storageMock.EXPECT().GetCountURLs().Return(1, tt.expectedErr)

			service := NewService(storageMock, logger.NewLogger(logger.WithLevel("info")))

			_, _, err := service.GetCountURLsAndUsers()

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Ожидали ошибку %v, пришла ошибка %v", tt.expectedErr, err)
			}
		})
	}
}
