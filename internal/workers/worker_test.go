package workers

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/logger"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/mocks"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/service"
	"testing"
)

type TestWorkerDeleted struct {
	name        string
	userID      string
	urls        []string
	expectedErr error
}

// TestNewWorkerDeleted - тестируем работоспособность воркера.
func TestNewWorkerDeleted(t *testing.T) {
	testCase := []TestWorkerDeleted{
		{
			name:        "successful",
			userID:      "test",
			urls:        []string{"www, vvv"},
			expectedErr: nil,
		},
	}

	for _, tt := range testCase {
		t.Run(tt.name, func(t *testing.T) {
			// делаем заглушку базы.
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockStorage := mocks.NewMockStorage(ctrl)

			// имитируем действие метода DeletedURLs
			mockStorage.EXPECT().DeletedURLs(gomock.Any(), gomock.Any()).Return(tt.expectedErr).AnyTimes()

			// создаем сервис.
			serviceTest := service.NewService(mockStorage, logger.NewLogger())

			// создаем воркер
			workTest := NewWorkerDeleted(serviceTest)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go workTest.StartWorkerDeletion(ctx)

			// Отправка запроса на удаление
			err := workTest.SendDeletionRequestToWorker(DeletionRequest{User: "test", URLs: []string{"example.com"}})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
