package db

import (
	"context"
	"database/sql"
	"fmt"

	errors2 "github.com/kamencov/go-musthave-shortener-tpl/internal/errorscustom"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
)

// GetURL возвращает оригинальную ссылку по короткой ссылке.
func (p *PstStorage) GetURL(shortURL string) (string, error) {
	var originalURL string
	var deletedURL bool
	//var storage []*models.Storage
	db := p.storage
	// создаем запрос
	query := "SELECT original_url, is_deleted FROM urls WHERE short_url = $1"
	// делаем запрос
	row := db.QueryRowContext(context.Background(), query, shortURL)

	if row == nil {
		return "", sql.ErrNoRows
	}

	if err := row.Scan(&originalURL, &deletedURL); err != nil {
		return "", err
	}

	if deletedURL {
		return "", errors2.ErrDeletedURL
	}

	return originalURL, nil
}

// GetAllURL возвращает все сохраненные ссылки пользователя.
func (p *PstStorage) GetAllURL(userID, baseURL string) ([]*models.UserURLs, error) {
	var userURLs []*models.UserURLs
	tx, err := p.storage.Begin()

	if err != nil {
		return nil, err
	}
	// создаем запрос
	query := "SELECT short_url, original_url FROM urls WHERE user_id = $1"

	// делаем запрос
	rows, err := tx.QueryContext(context.Background(), query, userID)
	if err != nil {
		return nil, sql.ErrNoRows
	}
	defer rows.Close()

	//собираем все сохраненные ссылки от пользователя
	for rows.Next() {
		var userURL models.UserURLs
		if err = rows.Scan(&userURL.ShortURL, &userURL.OriginalURL); err != nil {
			return nil, err
		}
		userURL.ShortURL = fmt.Sprintf("%s/%s", baseURL, userURL.ShortURL)
		userURLs = append(userURLs, &userURL)

	}

	if err = rows.Err(); err != nil {
		tx.Rollback()
		return nil, err
	}

	// завершаем транзакцию
	tx.Commit()
	return userURLs, nil
}
