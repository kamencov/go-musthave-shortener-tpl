package mapstorage

import (
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapStorage_SaveURL(t *testing.T) {
	t.Run("successful_saving", func(t *testing.T) {
		s := NewMapURL()
		err := s.SaveURL("test", "", "")
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("URL is empty"), err)
		err = s.SaveURL("test", "https://example.com", "")
		assert.Nil(t, err)
	})
}

func TestMapStorage_GetURL(t *testing.T) {
	t.Run("successful_getting", func(t *testing.T) {
		s := NewMapURL()
		err := s.SaveURL("test", "https://example.com", "")
		assert.Nil(t, err)
		_, err = s.GetURL("")
		assert.NotNil(t, err)
		assert.Equal(t, errors.New("URL not found"), err)
		url, err := s.GetURL("test")
		assert.Nil(t, err)
		assert.Equal(t, "https://example.com", url)
	})
}

func TestMapStorage_Plags(t *testing.T) {
	s := NewMapURL()

	t.Run("ping", func(t *testing.T) {
		err := s.Ping()
		if err != nil {
			t.Errorf("ожидался статус %v, но получен %v", nil, err)
		}
	})

	t.Run("get", func(t *testing.T) {
		_, err := s.GetAllURL("test", "https://example.com")
		if err == nil {
			t.Error("ожидали ошибку, но получили nil")
		}

		_, err = s.GetCountURLs()
		if err == nil {
			t.Error("ожидали ошибку, но получили nil")
		}

		_, err = s.GetCountUsers()
		if err == nil {
			t.Error("ожидали ошибку, но получили nil")
		}
	})

	t.Run("save", func(t *testing.T) {
		_, err := s.SaveSlice([]models.MultipleURL{}, "", "")
		if err != nil {
			t.Error("ожидали nil, но получили ошибку")
		}

	})

	t.Run("close", func(t *testing.T) {
		err := s.Close()
		if err != nil {
			t.Error("ожидали nil, но получили ошибку")
		}
	})
}
