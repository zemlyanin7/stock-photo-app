package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"stock-photo-app/models"
	"stock-photo-app/uploaders"
	"strings"
	"time"
)

// GettyImagesAPIUploader - пример интеграции с Getty Images API
// ВНИМАНИЕ: Это ПРИМЕР для разработчиков, не готовый к использованию код!
type GettyImagesAPIUploader struct {
	*uploaders.BaseUploader
}

// NewGettyImagesAPIUploader создает новый загрузчик для Getty Images
func NewGettyImagesAPIUploader() *GettyImagesAPIUploader {
	info := models.UploaderInfo{
		Name:        "Getty Images API Example",
		Version:     "1.0.0",
		Description: "Пример интеграции с Getty Images API",
		Author:      "Stock Photo App Developer",
		Type:        "api",
		Website:     "https://developers.gettyimages.com/",
	}

	return &GettyImagesAPIUploader{
		BaseUploader: uploaders.NewBaseUploader(info),
	}
}

// GettyImagesUploadRequest структура запроса для Getty Images API
type GettyImagesUploadRequest struct {
	Title      string   `json:"title"`
	Caption    string   `json:"caption"`
	Keywords   []string `json:"keywords"`
	Category   string   `json:"category"`
	Collection string   `json:"collection"`
	Editorial  bool     `json:"editorial"`
	Exclusive  bool     `json:"exclusive"`
	Rights     []string `json:"rights"`
}

// GettyImagesUploadResponse ответ от Getty Images API
type GettyImagesUploadResponse struct {
	AssetID    string `json:"asset_id"`
	Status     string `json:"status"`
	UploadURL  string `json:"upload_url"`
	PreviewURL string `json:"preview_url"`
	Error      string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

// Upload загружает фото в Getty Images
func (u *GettyImagesAPIUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Определяем тип контента
	isEditorial := false
	for _, supportedType := range config.SupportedTypes {
		if supportedType == "editorial" {
			isEditorial = true
			break
		}
	}

	// Подготавливаем метаданные для Getty Images
	uploadRequest := GettyImagesUploadRequest{
		Title:      photo.AIResult.Title,
		Caption:    photo.AIResult.Description,
		Keywords:   photo.AIResult.Keywords,
		Category:   photo.AIResult.Category,
		Collection: "iStock", // или другая коллекция
		Editorial:  isEditorial,
		Exclusive:  false,      // Определяется из настроек
		Rights:     []string{}, // Модельные и имущественные релизы
	}

	// Открываем файл
	file, err := os.Open(photo.OriginalPath)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка открытия файла: %v", err), false), err
	}
	defer file.Close()

	// Создаем multipart форму
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Добавляем файл
	filename := filepath.Base(photo.OriginalPath)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания формы: %v", err), false), err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка копирования файла: %v", err), false), err
	}

	// Добавляем метаданные как JSON
	metadataJSON, _ := json.Marshal(uploadRequest)
	writer.WriteField("metadata", string(metadataJSON))

	writer.Close()

	// Создаем HTTP запрос
	apiURL := config.Connection.APIUrl
	if apiURL == "" {
		apiURL = "https://api.gettyimages.com/v3/boards" // Примерный endpoint
	}

	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания запроса: %v", err), false), err
	}

	// Устанавливаем заголовки для Getty Images
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Api-Key", config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	// Добавляем дополнительные заголовки
	for key, value := range config.Connection.Headers {
		req.Header.Set(key, value)
	}

	// Создаем HTTP клиент
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка выполнения запроса: %v", err), false), err
	}
	defer resp.Body.Close()

	// Читаем ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка чтения ответа: %v", err), false), err
	}

	// Парсим ответ
	var gettyResponse GettyImagesUploadResponse
	json.Unmarshal(responseBody, &gettyResponse)

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := gettyResponse.Error
		if errorMsg == "" {
			errorMsg = string(responseBody)
		}
		return u.CreateUploadResult(photo.ID, config.ID,
				fmt.Sprintf("Getty Images API ошибка (код %d): %s",
					resp.StatusCode, errorMsg), false),
			fmt.Errorf("API вернул код %d", resp.StatusCode)
	}

	// Успешная загрузка
	result := u.CreateUploadResult(photo.ID, config.ID,
		"Файл успешно загружен в Getty Images", true)
	if gettyResponse.UploadURL != "" {
		result.UploadURL = gettyResponse.UploadURL
	}

	return result, nil
}

// TestConnection тестирует подключение к Getty Images API
func (u *GettyImagesAPIUploader) TestConnection(config models.StockConfig) error {
	// Простая проверка доступности API
	testURL := "https://api.gettyimages.com/v3/customers/current"

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания тестового запроса: %v", err)
	}

	req.Header.Set("Api-Key", config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Getty Images API: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("ошибка авторизации: проверьте API ключ Getty Images")
	case 403:
		return fmt.Errorf("доступ запрещен: недостаточно прав")
	default:
		return fmt.Errorf("Getty Images API вернул код %d", resp.StatusCode)
	}
}

// ValidateConfig проверяет конфигурацию для Getty Images
func (u *GettyImagesAPIUploader) ValidateConfig(config models.StockConfig) error {
	if config.Connection.APIKey == "" {
		return fmt.Errorf("требуется API ключ Getty Images")
	}

	if config.Connection.APIUrl != "" {
		if !strings.Contains(config.Connection.APIUrl, "gettyimages.com") {
			return fmt.Errorf("API URL должен содержать 'gettyimages.com'")
		}
	}

	return nil
}
