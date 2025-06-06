package uploaders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"stock-photo-app/models"
	"time"
)

// APIUploader реализует загрузку через REST API
type APIUploader struct {
	*BaseUploader
}

// NewAPIUploader создает новый API загрузчик
func NewAPIUploader() *APIUploader {
	info := models.UploaderInfo{
		Name:        "API Uploader",
		Version:     "1.0.0",
		Description: "Загрузка файлов через REST API",
		Author:      "Stock Photo App",
		Type:        "api",
	}

	return &APIUploader{
		BaseUploader: NewBaseUploader(info),
	}
}

// Upload загружает фото через API
func (u *APIUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Если это тестовый URL, имитируем загрузку
	if config.Connection.APIUrl == "https://api.shutterstock.com/v2/images" ||
		config.Connection.APIUrl == "https://demo.example.com/upload" ||
		config.Connection.APIUrl == "" {
		log.Printf("Demo mode: simulating upload of %s to %s", photo.FileName, config.Name)

		// Имитируем время загрузки
		time.Sleep(5 * time.Second)

		// Имитируем успешную загрузку
		result := u.CreateUploadResult(photo.ID, config.ID, "Файл успешно загружен (демо режим)", true)
		result.UploadURL = fmt.Sprintf("https://demo.example.com/photos/%s", photo.ID)
		return result, nil
	}

	// Открываем файл
	file, err := os.Open(photo.OriginalPath)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка открытия файла: %v", err), false), err
	}
	defer file.Close()

	// Создаем multipart форму
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем файл
	filename := filepath.Base(photo.OriginalPath)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка создания формы: %v", err), false), err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка копирования файла: %v", err), false), err
	}

	// Добавляем метаданные если есть
	if photo.AIResult != nil && photo.AIResult.Title != "" {
		writer.WriteField("title", photo.AIResult.Title)
	}
	if photo.AIResult != nil && photo.AIResult.Description != "" {
		writer.WriteField("description", photo.AIResult.Description)
	}
	if photo.AIResult != nil && len(photo.AIResult.Keywords) > 0 {
		keywordsJSON, _ := json.Marshal(photo.AIResult.Keywords)
		writer.WriteField("keywords", string(keywordsJSON))
	}

	writer.Close()

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", config.Connection.APIUrl, &requestBody)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка создания запроса: %v", err), false), err
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)

	// Добавляем дополнительные заголовки из конфигурации
	for key, value := range config.Connection.Headers {
		req.Header.Set(key, value)
	}

	// Создаем HTTP клиент с таймаутом
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка выполнения запроса: %v", err), false), err
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка чтения ответа: %v", err), false), err
	}

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return u.CreateUploadResult(photo.ID, config.ID,
				fmt.Sprintf("Ошибка API (код %d): %s", resp.StatusCode, string(responseBody)), false),
			fmt.Errorf("API вернул код %d", resp.StatusCode)
	}

	// Пытаемся распарсить ответ для получения URL
	var apiResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		URL     string `json:"url"`
		ID      string `json:"id"`
		Error   string `json:"error"`
	}

	json.Unmarshal(responseBody, &apiResponse)

	result := u.CreateUploadResult(photo.ID, config.ID, "Файл успешно загружен через API", true)
	if apiResponse.URL != "" {
		result.UploadURL = apiResponse.URL
	}

	return result, nil
}

// TestConnection тестирует подключение к API
func (u *APIUploader) TestConnection(config models.StockConfig) error {
	// Создаем тестовый запрос
	req, err := http.NewRequest("GET", config.Connection.APIUrl, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания тестового запроса: %v", err)
	}

	// Устанавливаем заголовки авторизации
	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)

	// Добавляем дополнительные заголовки
	for key, value := range config.Connection.Headers {
		req.Header.Set(key, value)
	}

	// Создаем HTTP клиент
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка подключения к API: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем, что API отвечает (даже если возвращает ошибку авторизации)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("ошибка авторизации: проверьте API ключ")
	}

	return nil
}

// ValidateConfig проверяет конфигурацию API
func (u *APIUploader) ValidateConfig(config models.StockConfig) error {
	requiredFields := []string{"apiUrl", "apiKey"}
	return u.ValidateRequiredFields(config, requiredFields)
}
