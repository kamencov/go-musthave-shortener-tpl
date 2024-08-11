package db

import (
	"context"
	"database/sql"
	"errors"
	errors2 "github.com/kamencov/go-musthave-shortener-tpl/internal/errors"
)

func (p *PstStorage) CheckURL(originalURL string) (string, error) {
	var shortURL string

	err := p.storage.QueryRowContext(context.Background(),
		"SELECT short_url FROM urls WHERE original_url = $1",
		originalURL).Scan(&shortURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return shortURL, errors2.ErrConflict
}
