package service

import (
	"errors"
	"github.com/kamencov/go-musthave-shortener-tpl/internal/utils"
	"log"
)

func (s *Service) SaveURL(url string) (string, error) {
	encodeURL, err := utils.EncodeURL(url)

	if err != nil {
		log.Println(err)
		return "", errors.New("URL is empty")
	}

	err = s.storage.SaveURL(encodeURL, url)
	if err != nil {
		log.Println(err)
		return "", err
	}

	log.Println("URL encoded successfully")

	return encodeURL, nil
}