package uploaders

import (
	"fmt"
	"path/filepath"
	"stock-photo-app/models"
	"strings"
)

// BaseUploader базовая структура для всех загрузчиков
type BaseUploader struct {
	info models.UploaderInfo
}

// NewBaseUploader создает новый базовый загрузчик
func NewBaseUploader(info models.UploaderInfo) *BaseUploader {
	return &BaseUploader{info: info}
}

// GetInfo возвращает информацию о загрузчике
func (b *BaseUploader) GetInfo() models.UploaderInfo {
	return b.info
}

// ValidateRequiredFields проверяет обязательные поля в конфигурации
func (b *BaseUploader) ValidateRequiredFields(config models.StockConfig, requiredFields []string) error {
	for _, field := range requiredFields {
		switch field {
		case "host":
			if config.Connection.Host == "" {
				return fmt.Errorf("поле Host обязательно для заполнения")
			}
		case "username":
			if config.Connection.Username == "" {
				return fmt.Errorf("поле Username обязательно для заполнения")
			}
		case "password":
			if config.Connection.Password == "" {
				return fmt.Errorf("поле Password обязательно для заполнения")
			}
		case "port":
			if config.Connection.Port <= 0 {
				return fmt.Errorf("поле Port должно быть больше 0")
			}
		case "apiKey":
			if config.Connection.APIKey == "" {
				return fmt.Errorf("поле API Key обязательно для заполнения")
			}
		case "apiUrl":
			if config.Connection.APIUrl == "" {
				return fmt.Errorf("поле API URL обязательно для заполнения")
			}
		}
	}
	return nil
}

// PrepareRemotePath подготавливает удаленный путь для загрузки
func (b *BaseUploader) PrepareRemotePath(config models.StockConfig, photo models.Photo) string {
	remotePath := config.Connection.Path
	if !strings.HasSuffix(remotePath, "/") {
		remotePath += "/"
	}

	// Добавляем имя файла
	filename := filepath.Base(photo.OriginalPath)
	return remotePath + filename
}

// CreateUploadResult создает результат загрузки
func (b *BaseUploader) CreateUploadResult(photoID, stockID, message string, success bool) models.UploadResult {
	return models.UploadResult{
		PhotoID: photoID,
		StockID: stockID,
		Success: success,
		Message: message,
	}
}

// GetStockTemplates возвращает шаблоны для разных типов стоков
func GetStockTemplates() map[string]models.StockTemplate {
	return map[string]models.StockTemplate{
		"ftp": {
			Type:        "ftp",
			Name:        "FTP Upload",
			Description: "Загрузка файлов через FTP протокол",
			Fields: []models.TemplateField{
				{Name: "host", Type: "text", Label: "FTP Сервер", Required: true, Placeholder: "ftp.example.com"},
				{Name: "port", Type: "number", Label: "Порт", Required: true, Default: 21, Placeholder: "21 для FTP, 990 для implicit FTPS"},
				{Name: "username", Type: "text", Label: "Имя пользователя", Required: true},
				{Name: "password", Type: "password", Label: "Пароль", Required: true},
				{Name: "path", Type: "text", Label: "Удаленная папка", Default: "/", Placeholder: "/uploads/"},
				{Name: "encryption", Type: "select", Label: "Шифрование", Default: "none", Options: []string{"none", "auto", "explicit", "implicit"}},
				{Name: "verifyCert", Type: "checkbox", Label: "Проверять SSL сертификаты", Default: true},
				{Name: "passive", Type: "checkbox", Label: "Пассивный режим", Default: true},
				{Name: "timeout", Type: "number", Label: "Таймаут (сек)", Default: 30},
			},
			Defaults: map[string]interface{}{
				"port":       21,
				"path":       "/",
				"encryption": "none",
				"verifyCert": true,
				"passive":    true,
				"timeout":    30,
			},
		},
		"sftp": {
			Type:        "sftp",
			Name:        "SFTP Upload",
			Description: "Загрузка файлов через SFTP протокол",
			Fields: []models.TemplateField{
				{Name: "host", Type: "text", Label: "SFTP Сервер", Required: true, Placeholder: "sftp.example.com"},
				{Name: "port", Type: "number", Label: "Порт", Required: true, Default: 22},
				{Name: "username", Type: "text", Label: "Имя пользователя", Required: true},
				{Name: "password", Type: "password", Label: "Пароль", Required: true},
				{Name: "path", Type: "text", Label: "Удаленная папка", Default: "/", Placeholder: "/uploads/"},
				{Name: "timeout", Type: "number", Label: "Таймаут (сек)", Default: 30},
			},
			Defaults: map[string]interface{}{
				"port":    22,
				"path":    "/",
				"timeout": 30,
			},
		},
	}
}
