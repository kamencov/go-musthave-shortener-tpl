package service

import (
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
)

func TestService_SaveURL(t *testing.T) {
	cases := []struct {
		name        string
		url         string
		userID      string
		errCheckURL error
		errSaveURL  error
		expectedErr error
	}{
		{
			name:        "successful",
			url:         "https://ya.ru",
			userID:      "testID",
			errCheckURL: nil,
			errSaveURL:  nil,
			expectedErr: nil,
		},
		{
			name:        "err_check_url",
			url:         "https://ya.ru",
			userID:      "testID",
			errCheckURL: errorscustom.ErrConflict,
			errSaveURL:  nil,
			expectedErr: errorscustom.ErrConflict,
		},
		{
			name:        "err_save_url",
			url:         "https://ya.ru",
			userID:      "testID",
			errCheckURL: nil,
			errSaveURL:  errorscustom.ErrConflict,
			expectedErr: errorscustom.ErrConflict,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storageMock := mocks.NewMockStorage(ctrl)
			serv := NewService(storageMock, logger.NewLogger())

			storageMock.EXPECT().
				CheckURL(gomock.Any()).
				Return(tt.url, tt.errCheckURL).
				AnyTimes()
			storageMock.EXPECT().
				SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.errSaveURL).
				AnyTimes()

			_, err := serv.SaveURL(tt.url, tt.userID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})

	}
}

func BenchmarkService_SaveURL(b *testing.B) {
	cntl := gomock.NewController(b)
	defer cntl.Finish()
	mockStorage := mocks.NewMockStorage(cntl)
	mockStorage.EXPECT().CheckURL(gomock.Any()).Return("https://example.com", nil).AnyTimes()
	mockStorage.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	service := NewService(mockStorage, logger.NewLogger(logger.WithLevel("info")))

	for i := 0; i < b.N; i++ {
		service.SaveURL("https://example.com", "")
	}
}
