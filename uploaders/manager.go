package uploaders

import (
	"fmt"
	"stock-photo-app/models"
	"sync"
)

// UploaderManager управляет всеми загрузчиками
type UploaderManager struct {
	uploaders map[string]models.StockUploader
	mu        sync.RWMutex
}

// NewUploaderManager создает новый менеджер загрузчиков
func NewUploaderManager(dbService DatabaseService) *UploaderManager {
	manager := &UploaderManager{
		uploaders: make(map[string]models.StockUploader),
	}

	// Регистрируем встроенные загрузчики
	manager.RegisterUploader("ftp", NewFTPUploader(dbService))
	manager.RegisterUploader("sftp", NewSFTPUploader(dbService))
	manager.RegisterUploader("api", NewAPIUploader())

	return manager
}

// RegisterUploader регистрирует новый загрузчик
func (m *UploaderManager) RegisterUploader(uploaderType string, uploader models.StockUploader) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploaders[uploaderType] = uploader
}

// GetUploader возвращает загрузчик по типу
func (m *UploaderManager) GetUploader(uploaderType string) (models.StockUploader, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	uploader, exists := m.uploaders[uploaderType]
	if !exists {
		return nil, fmt.Errorf("загрузчик типа '%s' не найден", uploaderType)
	}

	return uploader, nil
}

// GetAvailableUploaders возвращает список доступных загрузчиков
func (m *UploaderManager) GetAvailableUploaders() []models.UploaderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var uploaders []models.UploaderInfo
	for _, uploader := range m.uploaders {
		uploaders = append(uploaders, uploader.GetInfo())
	}

	return uploaders
}

// UploadPhoto загружает фото используя соответствующий загрузчик
func (m *UploaderManager) UploadPhoto(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Определяем тип загрузчика
	uploaderType := config.Type
	if uploaderType == "" {
		// Для обратной совместимости используем UploadMethod
		uploaderType = config.UploadMethod
	}

	// Получаем загрузчик
	uploader, err := m.GetUploader(uploaderType)
	if err != nil {
		return models.UploadResult{
			PhotoID: photo.ID,
			StockID: config.ID,
			Success: false,
			Message: fmt.Sprintf("Загрузчик не найден: %v", err),
		}, err
	}

	// Валидируем конфигурацию
	if err := uploader.ValidateConfig(config); err != nil {
		return models.UploadResult{
			PhotoID: photo.ID,
			StockID: config.ID,
			Success: false,
			Message: fmt.Sprintf("Ошибка конфигурации: %v", err),
		}, err
	}

	// Выполняем загрузку
	return uploader.Upload(photo, config)
}

// TestConnection тестирует подключение к стоку
func (m *UploaderManager) TestConnection(config models.StockConfig) error {
	// Определяем тип загрузчика
	uploaderType := config.Type
	if uploaderType == "" {
		uploaderType = config.UploadMethod
	}

	// Получаем загрузчик
	uploader, err := m.GetUploader(uploaderType)
	if err != nil {
		return err
	}

	// Валидируем конфигурацию
	if err := uploader.ValidateConfig(config); err != nil {
		return fmt.Errorf("ошибка конфигурации: %v", err)
	}

	// Тестируем подключение
	return uploader.TestConnection(config)
}

// ValidateStockConfig валидирует конфигурацию стока
func (m *UploaderManager) ValidateStockConfig(config models.StockConfig) error {
	// Определяем тип загрузчика
	uploaderType := config.Type
	if uploaderType == "" {
		uploaderType = config.UploadMethod
	}

	// Получаем загрузчик
	uploader, err := m.GetUploader(uploaderType)
	if err != nil {
		return err
	}

	// Валидируем конфигурацию
	return uploader.ValidateConfig(config)
}

// GetStockTemplates возвращает шаблоны для создания стоков
func (m *UploaderManager) GetStockTemplates() map[string]models.StockTemplate {
	return GetStockTemplates()
}

// GetSupportedTypes возвращает поддерживаемые типы загрузчиков
func (m *UploaderManager) GetSupportedTypes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var types []string
	for uploaderType := range m.uploaders {
		types = append(types, uploaderType)
	}

	return types
}
