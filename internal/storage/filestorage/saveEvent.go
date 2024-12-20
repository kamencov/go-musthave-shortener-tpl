package filestorage

import (
	"bufio"
	"encoding/json"
	"os"
)

// IFileStorage - интерфейс для хранения в файле.
type IFileStorage interface {
	SaveURL(shortURL, originalURL, userID string) error
	GetURL(shortURL string) (string, error)
	Close() error
}

// Count - счетчик для уникального идентификатора.
var Count int

// Event - структура для хранения событий.
type Event struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// SaveFile - структура для хранения в файле.
type SaveFile struct {
	file    *os.File
	encoder *json.Encoder
}

// NewSaveFile создает новый SaveFile.
func NewSaveFile(filePath string) (*SaveFile, error) {
	// откройте файл и создайте для него json.Encoder
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	readFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()

	scanner := bufio.NewScanner(readFile)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		Count++
	}

	return &SaveFile{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// WriteSaveModel добавляет Event в файл.
func (s *SaveFile) WriteSaveModel(event *Event) error {
	return s.encoder.Encode(&event)
}

// Close закрывает файл.
func (s *SaveFile) Close() error {
	return s.file.Close()
}

// Ping проверяет соединение с файлом.
func (s *SaveFile) Ping() error {
	return nil
}

// ReadFile - структура для чтения из файла.
type ReadFile struct {
	file    *os.File
	decoder *json.Decoder
}

// ReadEvent читает Event из файл.
func (c *ReadFile) ReadEvent() (*Event, error) {
	// добавьте вызов Decode для чтения и десериализации

	event := &Event{}
	if err := c.decoder.Decode(&event); err != nil {
		return nil, err
	}

	return event, nil
}

// Close закрывает файл.
func (c *ReadFile) Close() error {
	return c.file.Close()
}
