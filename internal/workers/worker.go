package workers

import (
	"context"
	"fmt"
	"sync"

	"github.com/kamencov/go-musthave-shortener-tpl/internal/service"
)

// Worker - интерфейс воркера.
//
//go:generate mockgen -source=worker.go -destination=mock_worker.go -package=workers
type Worker interface {
	SendDeletionRequestToWorker(req DeletionRequest) error
}

// DeletionRequest - запрос на удаление URL из хранилища.
type DeletionRequest struct {
	User string
	URLs []string
}

// WorkerDeleted - воркер для удаления URL из хранилища.
var deleteQueue = make(chan DeletionRequest, 10)

// WorkerDeleted - воркер для удаления URL из хранилища.
type WorkerDeleted struct {
	storage      *service.Service
	errorChannel chan error
	wg           *sync.WaitGroup
}

// NewWorkerDeleted - конструктор воркера.
func NewWorkerDeleted(storage *service.Service) *WorkerDeleted {
	return &WorkerDeleted{
		storage: storage,
		wg:      &sync.WaitGroup{},
	}
}

// StartWorkerDeletion стартует воркер для удаления URL из хранилища.
func (w *WorkerDeleted) StartWorkerDeletion(ctx context.Context) {
	// Запуск worker'а
	for {
		select {
		case req, ok := <-deleteQueue:
			if !ok {
				// Если deleteQueue закрыт, выходим из цикла
				w.wg.Wait()
				return
			}
			w.wg.Add(1)
			go func() {
				defer w.wg.Done()
				w.processDeletion(ctx, req)
			}()
		case <-ctx.Done():
			w.wg.Wait()
			return
		}
	}
}

// processDeletion обрабатывает удаление URL из хранилища.
func (w *WorkerDeleted) processDeletion(ctx context.Context, req DeletionRequest) {
	if err := w.storage.DeletedURLs(req.URLs, req.User); err != nil {
		select {
		case w.errorChannel <- err:
		case <-ctx.Done():
			fmt.Println("Operation canceled, skipping error reporting.")
		}
	}
}

// SendDeletionRequestToWorker отправляет запрос на удаление URL из хранилища.
func (w *WorkerDeleted) SendDeletionRequestToWorker(req DeletionRequest) error {
	select {
	case deleteQueue <- req:
		return nil
	default:
		return fmt.Errorf("the deletion request queue is currently full, please try again later")
	}
}
