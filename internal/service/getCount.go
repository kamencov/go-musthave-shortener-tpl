package service

import "golang.org/x/sync/errgroup"

// GetCountURLsAndUsers возвращает количество URL и пользователей в хранилище.
// Считает данные параллельно с помощью errgroup.
func (s *Service) GetCountURLsAndUsers() (int, int, error) {
	g := errgroup.Group{}

	var (
		countURLs  int
		countUsers int
	)

	g.Go(func() error {
		count, err := s.storage.GetCountURLs()
		if err != nil {
			return err
		}
		countURLs = count
		return nil
	})

	g.Go(func() error {
		count, err := s.storage.GetCountUsers()
		if err != nil {
			return err
		}
		countUsers = count
		return nil
	})
	if err := g.Wait(); err != nil {
		return 0, 0, err
	}
	return countURLs, countUsers, nil
}
