package main

import (
	"encoding/json"
	"os"
	"testing"
)

// TestParse - тестирует парсинг конфигурационной строки.
func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		addrServ string
		baseURL  string
		logLevel string
		pathFile string
		addrDB   string
	}{
		{
			name:     "successful",
			addrServ: ":8080",
			baseURL:  "http://localhost:8080",
			logLevel: "info",
			pathFile: "",
			addrDB:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfigs()
			cfg.AddrServer = tt.addrServ
			cfg.BaseURL = tt.baseURL
			cfg.LogLevel = tt.logLevel
			cfg.PathFile = tt.pathFile
			cfg.AddrDB = tt.addrDB
			cfg.Parse()

			if cfg.AddrServer != tt.addrServ {
				t.Errorf("Ожидали %v, пришли %v", tt.addrServ, cfg.AddrServer)
			}
			if cfg.BaseURL != tt.baseURL {
				t.Errorf("Ожидали %v, пришли %v", tt.baseURL, cfg.BaseURL)
			}
			if cfg.LogLevel != tt.logLevel {
				t.Errorf("Ожидали %v, пришли %v", tt.logLevel, cfg.LogLevel)
			}
			if cfg.PathFile != tt.pathFile {
				t.Errorf("Ожидали %v, пришли %v", tt.pathFile, cfg.PathFile)
			}
			if cfg.AddrDB != tt.addrDB {
				t.Errorf("Ожидали %v, пришли %v", tt.addrDB, cfg.AddrDB)
			}
		})
	}
}

func TestNewConfigs(t *testing.T) {
	cfg := NewConfigs()
	if cfg == nil {
		t.Errorf("Ожидали %v, пришли %v", nil, cfg)
	}

}

func TestLoadConfig_Successful(t *testing.T) {
	// Создаём временный JSON-файл для теста
	configData := Configs{
		AddrServer: ":9000",
		BaseURL:    "http://example.com",
		LogLevel:   "debug",
		PathFile:   "/tmp/storage",
		AddrDB:     "postgres://user:pass@localhost:5432/db",
		HTTPS:      new(bool),
	}

	*configData.HTTPS = true // устанавливаем значение для HTTPS в true

	file, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatalf("Не удалось создать временный файл: %v", err)
	}
	defer os.Remove(file.Name())

	// Сохраняем конфигурационные данные в JSON-файл
	if err := json.NewEncoder(file).Encode(configData); err != nil {
		t.Fatalf("Не удалось записать данные в файл: %v", err)
	}

	// Инициализируем тестируемую структуру Configs и устанавливаем путь к файлу конфигурации
	cfg := NewConfigs()
	cfg.ConfigFile = file.Name()
	err = cfg.loadConfig()
	if err != nil {
		t.Errorf("Ожидали %v, пришли %v", nil, err)
	}
}
