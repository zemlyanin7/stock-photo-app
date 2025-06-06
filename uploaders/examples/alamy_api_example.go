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

// AlamyAPIUploader - пример интеграции с Alamy API
// ВНИМАНИЕ: Это ПРИМЕР для разработчиков, не готовый к использованию код!
type AlamyAPIUploader struct {
	*uploaders.BaseUploader
}

// NewAlamyAPIUploader создает новый загрузчик для Alamy
func NewAlamyAPIUploader() *AlamyAPIUploader {
	info := models.UploaderInfo{
		Name:        "Alamy API Example",
		Version:     "1.0.0",
		Description: "Пример интеграции с Alamy API",
		Author:      "Stock Photo App Developer",
		Type:        "api",
		Website:     "https://www.alamy.com/contributor/",
	}

	return &AlamyAPIUploader{
		BaseUploader: uploaders.NewBaseUploader(info),
	}
}

// AlamyUploadRequest структура запроса для Alamy API
type AlamyUploadRequest struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Keywords         []string `json:"keywords"`
	Category         string   `json:"category"`
	ImageType        string   `json:"image_type"` // "editorial", "creative"
	ModelReleased    bool     `json:"model_released"`
	PropertyReleased bool     `json:"property_released"`
	Location         string   `json:"location,omitempty"`
	DateTaken        string   `json:"date_taken,omitempty"`
}

// AlamyUploadResponse ответ от Alamy API
type AlamyUploadResponse struct {
	ImageID    string `json:"image_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	PreviewURL string `json:"preview_url,omitempty"`
	Error      string `json:"error,omitempty"`
}

// Upload загружает фото в Alamy
func (u *AlamyAPIUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Определяем тип изображения
	imageType := "creative"
	for _, supportedType := range config.SupportedTypes {
		if supportedType == "editorial" {
			imageType = "editorial"
			break
		}
	}

	// Подготавливаем метаданные для Alamy
	uploadRequest := AlamyUploadRequest{
		Title:            photo.AIResult.Title,
		Description:      photo.AIResult.Description,
		Keywords:         photo.AIResult.Keywords,
		Category:         photo.AIResult.Category,
		ImageType:        imageType,
		ModelReleased:    false, // Определяется из EXIF или настроек
		PropertyReleased: false,
		Location:         "", // Извлекается из EXIF GPS данных
		DateTaken:        "", // Извлекается из EXIF
	}

	// Извлекаем дополнительную информацию из EXIF
	if dateTaken, exists := photo.ExifData["DateTime"]; exists {
		uploadRequest.DateTaken = dateTaken
	}
	if location, exists := photo.ExifData["GPSLocation"]; exists {
		uploadRequest.Location = location
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
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания формы: %v", err), false), err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка копирования файла: %v", err), false), err
	}

	// Добавляем каждое поле метаданных отдельно
	writer.WriteField("title", uploadRequest.Title)
	writer.WriteField("description", uploadRequest.Description)
	writer.WriteField("image_type", uploadRequest.ImageType)
	writer.WriteField("keywords", strings.Join(uploadRequest.Keywords, ", "))

	if uploadRequest.Category != "" {
		writer.WriteField("category", uploadRequest.Category)
	}
	if uploadRequest.Location != "" {
		writer.WriteField("location", uploadRequest.Location)
	}
	if uploadRequest.DateTaken != "" {
		writer.WriteField("date_taken", uploadRequest.DateTaken)
	}

	writer.Close()

	// Создаем HTTP запрос
	apiURL := config.Connection.APIUrl
	if apiURL == "" {
		apiURL = "https://api.alamy.com/v1/images/upload" // Примерный endpoint
	}

	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return u.CreateUploadResult(photo.ID, config.ID,
			fmt.Sprintf("Ошибка создания запроса: %v", err), false), err
	}

	// Устанавливаем заголовки для Alamy
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)
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
	var alamyResponse AlamyUploadResponse
	json.Unmarshal(responseBody, &alamyResponse)

	// Проверяем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errorMsg := alamyResponse.Error
		if errorMsg == "" {
			errorMsg = string(responseBody)
		}
		return u.CreateUploadResult(photo.ID, config.ID,
				fmt.Sprintf("Alamy API ошибка (код %d): %s",
					resp.StatusCode, errorMsg), false),
			fmt.Errorf("API вернул код %d", resp.StatusCode)
	}

	// Успешная загрузка
	message := "Файл успешно загружен в Alamy"
	if alamyResponse.ImageID != "" {
		message += fmt.Sprintf(" (ID: %s)", alamyResponse.ImageID)
	}

	result := u.CreateUploadResult(photo.ID, config.ID, message, true)
	if alamyResponse.PreviewURL != "" {
		result.UploadURL = alamyResponse.PreviewURL
	}

	return result, nil
}

// TestConnection тестирует подключение к Alamy API
func (u *AlamyAPIUploader) TestConnection(config models.StockConfig) error {
	// Проверяем доступность API через запрос к профилю
	testURL := "https://api.alamy.com/v1/user/profile" // Примерный endpoint

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return fmt.Errorf("ошибка создания тестового запроса: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+config.Connection.APIKey)
	req.Header.Set("User-Agent", "Stock Photo App v1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Alamy API: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return nil
	case 401:
		return fmt.Errorf("ошибка авторизации: проверьте API ключ Alamy")
	case 403:
		return fmt.Errorf("доступ запрещен: недостаточно прав")
	default:
		return fmt.Errorf("Alamy API вернул код %d", resp.StatusCode)
	}
}

// ValidateConfig проверяет конфигурацию для Alamy
func (u *AlamyAPIUploader) ValidateConfig(config models.StockConfig) error {
	if config.Connection.APIKey == "" {
		return fmt.Errorf("требуется API ключ Alamy")
	}

	if config.Connection.APIUrl != "" {
		if !strings.Contains(config.Connection.APIUrl, "alamy.com") {
			return fmt.Errorf("API URL должен содержать 'alamy.com'")
		}
	}

	return nil
}
