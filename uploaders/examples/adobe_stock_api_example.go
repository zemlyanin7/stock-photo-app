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

// AdobeStockAPIUploader - пример интеграции с Adobe Stock API
// ВНИМАНИЕ: Это ПРИМЕР для разработчиков, не готовый к использованию код!
type AdobeStockAPIUploader struct {
	*uploaders.BaseUploader
}

// NewAdobeStockAPIUploader создает новый загрузчик для Adobe Stock
func NewAdobeStockAPIUploader() *AdobeStockAPIUploader {
	info := models.UploaderInfo{
		Name:        "Adobe Stock API Example",
		Version:     "1.0.0",
		Description: "Пример интеграции с Adobe Stock API",
		Author:      "Stock Photo App Developer",
		Type:        "api",
		Website:     "https://developer.adobe.com/stock/",
	}

	return &AdobeStockAPIUploader{
		BaseUploader: uploaders.NewBaseUploader(info),
	}
}

// AdobeStockUploadRequest структура запроса для Adobe Stock API
type AdobeStockUploadRequest struct {
	Title        string   `json:"title"`
	Keywords     []string `json:"keywords"`
	Description  string   `json:"description"`
	Category     int      `json:"category"`
	ContentType  string   `json:"content_type"`  // "photo", "vector", "video"
	ContentLevel string   `json:"content_level"` // "1" = general, "2" = moderate, "3" = adult
	ReleaseInfo  []string `json:"release_info"`  // model/property releases
	EditorialUse bool     `json:"editorial_use"`
	CreativeUse  bool     `json:"creative_use"`
}

// AdobeStockUploadResponse ответ от Adobe Stock API
type AdobeStockUploadResponse struct {
	SubmissionID string `json:"submission_id"`
	Status       string `json:"status"`
	StatusDetail string `json:"status_detail"`
	PreviewURL   string `json:"preview_url"`
	Error        string `json:"error,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
}

// Upload загружает фото в Adobe Stock
func (u *AdobeStockAPIUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Определяем тип контента и уровень на основе настроек
	isEditorial := false
	for _, supportedType := range config.SupportedTypes {
		if supportedType == "editorial" {
			isEditorial = true
			break
		}
	}

	// Подготавливаем метаданные для Adobe Stock
	uploadRequest := AdobeStockUploadRequest{
		Title:        photo.AIResult.Title,
		Keywords:     photo.AIResult.Keywords,
		Description:  photo.AIResult.Description,
		Category:     1, // Определяется из AI анализа или настроек
		ContentType:  "photo",
		ContentLevel: "1",        // General content
		ReleaseInfo:  []string{}, // Должно заполняться из EXIF или настроек
		EditorialUse: isEditorial,
		CreativeUse:  !isEditorial,
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

	// Добавляем файл изображения
	filename := filepath.Base(photo.OriginalPath)
	filePart, err := writer.CreateFormFile("content", filename)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания формы для файла: %v", err), false), err
	}

	_, err = io.Copy(filePart, file)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка копирования файла: %v", err), false), err
	}

	// Добавляем метаданные как отдельные поля
	writer.WriteField("title", uploadRequest.Title)
	writer.WriteField("description", uploadRequest.Description)
	writer.WriteField("content_type", uploadRequest.ContentType)
	writer.WriteField("content_level", uploadRequest.ContentLevel)

	// Добавляем ключевые слова как строку через запятую
	keywordsStr := strings.Join(uploadRequest.Keywords, ", ")
	writer.WriteField("keywords", keywordsStr)

	// Добавляем категорию
	writer.WriteField("category", fmt.Sprintf("%d", uploadRequest.Category))

	// Добавляем флаги использования
	if uploadRequest.EditorialUse {
		writer.WriteField("editorial_use", "true")
	}
	if uploadRequest.CreativeUse {
		writer.WriteField("creative_use", "true")
	}

	writer.Close()

	// Создаем HTTP запрос
	apiURL := config.Connection.APIUrl
	if apiURL == "" {
		apiURL = "https://stock-stage.adobe.io/Rest/Media/1/Files"
	}

	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания запроса: %v", err), false), err
	}

	// Устанавливаем заголовки для Adobe Stock
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	// Adobe Stock может требовать дополнительных заголовков
	if bearerToken, exists := config.Connection.Headers["Authorization"]; exists {
		req.Header.Set("Authorization", bearerToken)
	}

	// Добавляем дополнительные заголовки
	for key, value := range config.Connection.Headers {
		if key != "Authorization" { // Уже добавлено выше
			req.Header.Set(key, value)
		}
	}

	// Создаем HTTP клиент с увеличенным таймаутом
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 180 * time.Second // Adobe Stock может требовать еще больше времени
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
	var adobeResponse AdobeStockUploadResponse
	if err := json.Unmarshal(responseBody, &adobeResponse); err != nil {
		// Если не удалось распарсить JSON, проверяем статус
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return u.CreateUploadResult(photo.ID, config.ID,
					fmt.Sprintf("Adobe Stock API ошибка (код %d): %s",
						resp.StatusCode, string(responseBody)), false),
				fmt.Errorf("API вернул код %d", resp.StatusCode)
		}
	}

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := adobeResponse.Error
		if errorMsg == "" {
			errorMsg = string(responseBody)
		}
		return u.CreateUploadResult(photo.ID, config.ID,
				fmt.Sprintf("Adobe Stock API ошибка (код %d): %s",
					resp.StatusCode, errorMsg), false),
			fmt.Errorf("API вернул код %d", resp.StatusCode)
	}

	// Успешная загрузка
	message := "Файл успешно загружен в Adobe Stock"
	if adobeResponse.SubmissionID != "" {
		message += fmt.Sprintf(" (ID: %s)", adobeResponse.SubmissionID)
	}

	result := u.CreateUploadResult(photo.ID, config.ID, message, true)
	if adobeResponse.PreviewURL != "" {
		result.UploadURL = adobeResponse.PreviewURL
	}

	return result, nil
}

// TestConnection тестирует подключение к Adobe Stock API
func (u *AdobeStockAPIUploader) TestConnection(config models.StockConfig) error {
	// Проверяем доступность API через запрос к профилю
	testURL := "https://stock-stage.adobe.io/Rest/Libraries/1/Member/Profile"
	if config.Connection.APIUrl != "" {
		// Используем базовый URL из конфигурации для построения тестового URL
		baseURL := strings.Replace(config.Connection.APIUrl, "/Files", "", 1)
		testURL = baseURL + "/Member/Profile"
	}

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания тестового запроса: %v", err)
	}

	// Устанавливаем заголовки авторизации для Adobe Stock
	req.Header.Set("X-API-Key", config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	// Добавляем Authorization заголовок если есть
	if bearerToken, exists := config.Connection.Headers["Authorization"]; exists {
		req.Header.Set("Authorization", bearerToken)
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
		return fmt.Errorf("ошибка подключения к Adobe Stock API: %v", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	switch resp.StatusCode {
	case 200:
		return nil // Успешное подключение
	case 401:
		return fmt.Errorf("ошибка авторизации: проверьте API ключ и токен Adobe Stock")
	case 403:
		return fmt.Errorf("доступ запрещен: недостаточно прав для загрузки в Adobe Stock")
	case 429:
		return fmt.Errorf("превышен лимит запросов к Adobe Stock API")
	default:
		// Читаем тело ответа для дополнительной информации
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Adobe Stock API вернул код %d: %s", resp.StatusCode, string(body))
	}
}

// ValidateConfig проверяет конфигурацию для Adobe Stock
func (u *AdobeStockAPIUploader) ValidateConfig(config models.StockConfig) error {
	// Проверяем обязательные поля
	if config.Connection.APIKey == "" {
		return fmt.Errorf("требуется API ключ Adobe Stock (X-API-Key)")
	}

	// Проверяем наличие токена авторизации
	if _, exists := config.Connection.Headers["Authorization"]; !exists {
		return fmt.Errorf("требуется токен авторизации Adobe Stock (Bearer token)")
	}

	// Проверяем формат API ключа
	if len(config.Connection.APIKey) < 32 {
		return fmt.Errorf("API ключ Adobe Stock слишком короткий")
	}

	// Проверяем URL если указан
	if config.Connection.APIUrl != "" {
		if !strings.Contains(config.Connection.APIUrl, "adobe.io") &&
			!strings.Contains(config.Connection.APIUrl, "adobe.com") {
			return fmt.Errorf("API URL должен быть Adobe Stock API endpoint")
		}
	}

	return nil
}
