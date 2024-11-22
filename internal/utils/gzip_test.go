package utils

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCompressWriter(t *testing.T) {

	w := NewCompressWriter(nil)
	if w == nil {
		t.Errorf("ожидался не nil, но получен nil")
	}
}

func TestCompressWriter_Header(t *testing.T) {
	resp := httptest.NewRecorder()

	w := NewCompressWriter(resp)
	h := w.Header()

	if h == nil {
		t.Errorf("ожидался не nil, но получен nil")
	}
}

func TestCompressWriter_Write(t *testing.T) {
	resp := httptest.NewRecorder()

	w := NewCompressWriter(resp)

	if _, err := w.Write([]byte("test")); err != nil {
		t.Errorf("ожидался nil, но получен %v", err)
	}
}

func TestCompressWriter_WriteHeader(t *testing.T) {

	cases := []struct {
		name string
		code int
	}{
		{
			name: "successful",
			code: http.StatusOK,
		},
		{
			name: "error",
			code: http.StatusInternalServerError,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			reps := httptest.NewRecorder()

			w := NewCompressWriter(reps)

			w.WriteHeader(tt.code)
			if reps.Code != tt.code {
				t.Errorf("ожидался %v, но получен %v", tt.code, reps.Code)
			}
		})
	}
}

func TestCompressWriter_Close(t *testing.T) {
	reps := httptest.NewRecorder()

	w := NewCompressWriter(reps)

	if err := w.Close(); err != nil {
		t.Errorf("ожидался nil, но получен %v", err)
	}
}

// mockReader используется для имитации io.ReadCloser.
type mockReader struct {
	data []byte
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), io.EOF
}

func (m *mockReader) Close() error {
	// Имитация закрытия ресурса.
	return nil
}

// mockReaderError используется для имитации io.ReadCloser с ошибкой.
type mockReaderError struct {
	data []byte
}

func (m *mockReaderError) Read(p []byte) (n int, err error) {
	copy(p, m.data)
	return len(m.data), io.EOF
}

func (m *mockReaderError) Close() error {
	return errors.New("gzip close error")
}

func TestNewCompressReader(t *testing.T) {
	t.Run("successful_compression_reader", func(t *testing.T) {
		// Создание тестовых данных
		data := []byte("test data to be compressed")

		// Сжимаем данные с помощью gzip
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err := writer.Write(data)
		if err != nil {
			t.Fatalf("gzip write failed: %v", err)
		}
		writer.Close()

		// Создаем новый reader для сжатых данных
		r := &mockReader{data: buf.Bytes()}
		compressReader, err := NewCompressReader(r)
		if err != nil {
			t.Fatalf("ошибка создания compressReader: %v", err)
		}

		// Проверяем, что compressReader не nil
		if compressReader == nil {
			t.Fatal("ожидался не nil, но получен nil")
		}
	})

	t.Run("reader_error", func(t *testing.T) {
		// Создаем mockReader с некорректными данными
		r := &mockReader{data: []byte("invalid gzip data")}

		// Пытаемся создать compressReader, ожидаем ошибку
		_, err := NewCompressReader(r)
		if err == nil {
			t.Fatal("ожидается ошибку, но получено nil")
		}
	})

}

func TestCompressReader_Read(t *testing.T) {
	// Создание тестовых данных
	data := []byte("test data to be compressed")

	// Сжимаем данные с помощью gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		t.Fatalf("gzip write failed: %v", err)
	}
	writer.Close()

	// Создаем новый reader для сжатых данных
	r := &mockReader{data: buf.Bytes()}
	compressReader, err := NewCompressReader(r)
	if err != nil {
		t.Fatalf("ошибка создания compressReader: %v", err)
	}

	// Проверяем, что compressReader не nil
	if compressReader == nil {
		t.Fatal("ожидался не nil, но получен nil")
	}

	// Пробуем прочитать из compressReader
	var decompressedData []byte
	_, err = compressReader.Read(decompressedData)
	if err != nil && err != io.EOF {
		t.Fatalf("unexpected read error: %v", err)
	}
}

func TestCompressReader_Close(t *testing.T) {
	t.Run("successful_close", func(t *testing.T) {
		// Создание тестовых данных
		data := []byte("test data to be compressed")

		// Сжимаем данные с помощью gzip
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err := writer.Write(data)
		if err != nil {
			t.Fatalf("gzip write failed: %v", err)
		}
		writer.Close()

		// Создаем новый reader для сжатых данных
		r := &mockReader{data: buf.Bytes()}
		compressReader, err := NewCompressReader(r)
		if err != nil {
			t.Fatalf("ошибка создания compressReader: %v", err)
		}

		// Проверяем, что compressReader не nil
		if compressReader == nil {
			t.Fatal("ожидался не nil, но получен nil")
		}

		// Пробуем закрыть compressReader
		if err := compressReader.Close(); err != nil {
			t.Fatalf("ошибка закрытия compressReader: %v", err)
		}
	})

	t.Run("close_error", func(t *testing.T) {
		// Создаем mockGzipReaderWithCloseError
		data := []byte("test data")
		var buf bytes.Buffer
		writer := gzip.NewWriter(&buf)
		_, err := writer.Write(data)
		if err != nil {
			t.Fatalf("gzip write failed: %v", err)
		}
		writer.Close()

		r := &mockReaderError{data: buf.Bytes()}
		compressReader, err := NewCompressReader(r)
		if err != nil {
			t.Fatalf("unexpected error creating compressReader: %v", err)
		}

		if err := compressReader.Close(); err == nil {
			t.Fatal("expected error, but got nil")
		}
	})
}
