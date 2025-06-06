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

// ShutterstockAPIUploader - пример интеграции с Shutterstock API
// ВНИМАНИЕ: Это ПРИМЕР для разработчиков, не готовый к использованию код!
type ShutterstockAPIUploader struct {
	*uploaders.BaseUploader
}

// NewShutterstockAPIUploader создает новый загрузчик для Shutterstock
func NewShutterstockAPIUploader() *ShutterstockAPIUploader {
	info := models.UploaderInfo{
		Name:        "Shutterstock API Example",
		Version:     "1.0.0",
		Description: "Пример интеграции с Shutterstock API",
		Author:      "Stock Photo App Developer",
		Type:        "api",
		Website:     "https://developers.shutterstock.com/",
	}

	return &ShutterstockAPIUploader{
		BaseUploader: uploaders.NewBaseUploader(info),
	}
}

// ShutterstockUploadRequest структура запроса для Shutterstock API
type ShutterstockUploadRequest struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Keywords         []string `json:"keywords"`
	Categories       []int    `json:"categories"`
	ModelReleased    bool     `json:"model_released"`
	PropertyReleased bool     `json:"property_released"`
	EditorialUse     bool     `json:"editorial_use"`
}

// ShutterstockUploadResponse ответ от Shutterstock API
type ShutterstockUploadResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	UploadURL string `json:"upload_url"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	ErrorCode int    `json:"error_code,omitempty"`
}

// Upload загружает фото в Shutterstock
func (u *ShutterstockAPIUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Подготавливаем метаданные для Shutterstock
	uploadRequest := ShutterstockUploadRequest{
		Title:            photo.AIResult.Title,
		Description:      photo.AIResult.Description,
		Keywords:         photo.AIResult.Keywords,
		Categories:       []int{1}, // Категория по умолчанию, должна определяться из AI анализа
		ModelReleased:    false,    // Определяется из EXIF или настроек
		PropertyReleased: false,
		EditorialUse:     strings.Contains(config.Type, "editorial"),
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
		apiURL = "https://api.shutterstock.com/v2/images"
	}

	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания запроса: %v", err), false), err
	}

	// Устанавливаем заголовки для Shutterstock
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	// Добавляем дополнительные заголовки
	for key, value := range config.Connection.Headers {
		req.Header.Set(key, value)
	}

	// Создаем HTTP клиент с таймаутом
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second // Shutterstock может требовать больше времени
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
	var shutterstockResponse ShutterstockUploadResponse
	if err := json.Unmarshal(responseBody, &shutterstockResponse); err != nil {
		// Если не удалось распарсить JSON, возвращаем как есть
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return u.CreateUploadResult(photo.ID, config.ID,
					fmt.Sprintf("Shutterstock API ошибка (код %d): %s",
						resp.StatusCode, string(responseBody)), false),
				fmt.Errorf("API вернул код %d", resp.StatusCode)
		}
	}

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := shutterstockResponse.Error
		if errorMsg == "" {
			errorMsg = string(responseBody)
		}
		return u.CreateUploadResult(photo.ID, config.ID,
				fmt.Sprintf("Shutterstock API ошибка (код %d): %s",
					resp.StatusCode, errorMsg), false),
			fmt.Errorf("API вернул код %d", resp.StatusCode)
	}

	// Успешная загрузка
	result := u.CreateUploadResult(photo.ID, config.ID,
		"Файл успешно загружен в Shutterstock", true)
	if shutterstockResponse.UploadURL != "" {
		result.UploadURL = shutterstockResponse.UploadURL
	}

	return result, nil
}

// TestConnection тестирует подключение к Shutterstock API
func (u *ShutterstockAPIUploader) TestConnection(config models.StockConfig) error {
	// Проверяем доступность API через запрос к информации о пользователе
	testURL := "https://api.shutterstock.com/v2/user"
	if config.Connection.APIUrl != "" {
		// Используем базовый URL из конфигурации
		testURL = strings.Replace(config.Connection.APIUrl, "/images", "/user", 1)
	}

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания тестового запроса: %v", err)
	}

	// Устанавливаем заголовки авторизации
	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	// Создаем HTTP клиент
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Shutterstock API: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	switch resp.StatusCode {
	case 200:
		return nil // Успешное подключение
	case 401:
		return fmt.Errorf("ошибка авторизации: проверьте API ключ Shutterstock")
	case 403:
		return fmt.Errorf("доступ запрещен: недостаточно прав для API ключа")
	case 429:
		return fmt.Errorf("превышен лимит запросов к Shutterstock API")
	default:
		return fmt.Errorf("Shutterstock API вернул код %d", resp.StatusCode)
	}
}

// ValidateConfig проверяет конфигурацию для Shutterstock
func (u *ShutterstockAPIUploader) ValidateConfig(config models.StockConfig) error {
	// Проверяем обязательные поля
	if config.Connection.APIKey == "" {
		return fmt.Errorf("требуется API ключ Shutterstock")
	}

	// Проверяем формат API ключа (примерная валидация)
	if len(config.Connection.APIKey) < 32 {
		return fmt.Errorf("API ключ Shutterstock слишком короткий")
	}

	// Проверяем URL если указан
	if config.Connection.APIUrl != "" {
		if !strings.Contains(config.Connection.APIUrl, "shutterstock.com") {
			return fmt.Errorf("API URL должен содержать 'shutterstock.com'")
		}
	}

	return nil
}
