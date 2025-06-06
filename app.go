package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"stock-photo-app/models"
	"stock-photo-app/services"
	"stock-photo-app/uploaders"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	db              *sql.DB
	aiService       *services.AIService
	imageProc       *services.ImageProcessor
	dbService       *services.DatabaseService
	uploaderManager *uploaders.UploaderManager
	queueManager    *services.QueueManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// OnStartup is called when the app starts. The context provided
// can be used to perform any startup tasks.
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx

	// Инициализация базы данных
	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	a.db = db

	// Инициализация сервисов
	a.dbService = services.NewDatabaseService(db)
	a.aiService = services.NewAIService()
	a.imageProc = services.NewImageProcessor("./temp")
	a.uploaderManager = uploaders.NewUploaderManager(a.dbService)
	a.queueManager = services.NewQueueManager(db, a.dbService, a.aiService, a.imageProc)

	// Создание таблиц БД
	err = a.dbService.InitializeTables()
	if err != nil {
		log.Fatal("Failed to initialize database tables:", err)
	}

	// Загружаем настройки при старте
	_, err = a.dbService.GetSettings()
	if err != nil {
		log.Printf("Warning: Failed to load settings on startup: %v", err)
	}

	log.Println("App initialized successfully")
}

// OnDomReady is called after front-end resources have been loaded
func (a *App) OnDomReady(ctx context.Context) {
	// Drag & drop уже настроен через опции приложения
	// OnFileDrop будет вызывать callback напрямую из JavaScript
}

// OnBeforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) OnBeforeClose(ctx context.Context) (prevent bool) {
	a.db.Close()
	return false
}

// OnShutdown is called during shutdown after OnBeforeClose
func (a *App) OnShutdown(ctx context.Context) {
	// Perform your teardown here
}

// ProcessPhotoFolder - основной метод для обработки папки с фотографиями
func (a *App) ProcessPhotoFolder(folderPath string, description string, photoType string) error {
	log.Printf("Processing folder: %s, type: %s", folderPath, photoType)

	// Создаем новый батч
	batch := models.PhotoBatch{
		ID:          fmt.Sprintf("batch_%d", getCurrentTimestamp()),
		Type:        photoType,
		Description: description,
		FolderPath:  folderPath,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Сканируем изображения в папке
	photos, err := a.imageProc.ScanFolder(folderPath)
	if err != nil {
		return fmt.Errorf("failed to scan folder: %w", err)
	}

	// Устанавливаем ContentType для всех фотографий
	for i := range photos {
		photos[i].ContentType = photoType
	}

	batch.Photos = photos

	// Добавляем в очередь обработки
	err = a.queueManager.AddBatch(batch)
	if err != nil {
		return fmt.Errorf("failed to add batch to queue: %w", err)
	}

	// Запускаем обработку очереди если она не запущена
	settings, err := a.dbService.GetSettings()
	if err != nil {
		log.Printf("Warning: failed to get settings for queue processing: %v", err)
		// Используем настройки по умолчанию
		settings = models.AppSettings{
			MaxConcurrentJobs: 3,
			ThumbnailSize:     512,
			AIProvider:        "openai",
			AIModel:           "gpt-4o",
			TempDirectory:     "./temp",
		}
	}

	err = a.queueManager.StartProcessing(settings)
	if err != nil {
		log.Printf("Queue processing already running or failed to start: %v", err)
	}

	log.Printf("Batch %s with %d photos added to processing queue", batch.ID, len(photos))
	return nil
}

// GetQueueStatus возвращает статус текущей очереди
func (a *App) GetQueueStatus() ([]models.BatchStatus, error) {
	return a.queueManager.GetQueueStatus()
}

// GetStockConfigs возвращает все конфигурации стоков
func (a *App) GetStockConfigs() ([]models.StockConfig, error) {
	return a.dbService.GetAllStockConfigs()
}

// SaveStockConfig сохраняет конфигурацию стока
func (a *App) SaveStockConfig(config models.StockConfig) error {
	return a.dbService.SaveStockConfig(config)
}

// DeleteStockConfig удаляет конфигурацию стока
func (a *App) DeleteStockConfig(stockID string) error {
	log.Printf("App.DeleteStockConfig called with ID: %s", stockID)
	err := a.dbService.DeleteStockConfig(stockID)
	if err != nil {
		log.Printf("App.DeleteStockConfig failed: %v", err)
		return err
	}
	log.Printf("App.DeleteStockConfig completed successfully for ID: %s", stockID)
	return nil
}

// TestStockConnection тестирует подключение к стоку
func (a *App) TestStockConnection(config models.StockConfig) error {
	return a.uploaderManager.TestConnection(config)
}

// ToggleStockActive переключает активность стока
func (a *App) ToggleStockActive(stockID string) error {
	// Получаем текущую конфигурацию
	stocks, err := a.dbService.GetAllStockConfigs()
	if err != nil {
		return fmt.Errorf("failed to get stock configs: %w", err)
	}

	// Находим сток и переключаем активность
	for _, stock := range stocks {
		if stock.ID == stockID {
			stock.Active = !stock.Active
			return a.dbService.SaveStockConfig(stock)
		}
	}

	return fmt.Errorf("stock with ID %s not found", stockID)
}

// GetSettings возвращает настройки приложения
func (a *App) GetSettings() (models.AppSettings, error) {
	return a.dbService.GetSettings()
}

// SaveSettings сохраняет настройки приложения
func (a *App) SaveSettings(settings models.AppSettings) error {
	return a.dbService.SaveSettings(settings)
}

// UpdateAIPrompt обновляет промпт для AI
func (a *App) UpdateAIPrompt(photoType string, prompt string) error {
	return a.dbService.UpdateAIPrompt(photoType, prompt)
}

// ForceUpdateDefaultPrompts принудительно обновляет дефолтные промпты
func (a *App) ForceUpdateDefaultPrompts() error {
	// Получаем текущие настройки
	settings, err := a.dbService.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	// Устанавливаем дефолтные промпты
	defaultPrompts := map[string]string{
		"editorial": `Создай метаданные для редакционного стокового фото:

ТРЕБОВАНИЯ ДЛЯ EDITORIAL:
1. НАЗВАНИЕ (до 100 символов):
   - Фактическое описание события/сюжета
   - Конкретные имена людей и мест (если применимо)
   - Журналистский стиль без эмоциональной окраски
   - Временной контекст при необходимости

2. ОПИСАНИЕ (до 500 символов):
   - WHO: конкретные имена людей
   - WHAT: точное описание происходящего  
   - WHERE: конкретные места с полными названиями
   - WHEN: даты и время (используй EXIF данные)
   - WHY: контекст и причины события

3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
   - Конкретные имена людей, если узнаваемы
   - Точные названия мест, зданий, организаций
   - События, действия, происходящие на фото
   - Эмоции и настроение людей
   - Политический/социальный контекст
   - Временные маркеры (даты, сезоны, время дня)
   - Объекты, предметы, детали архитектуры
   - Погода, освещение, атмосфера

4. КАТЕГОРИЯ (выбери одну из для Editorial):
   News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment, Healthcare, Education, Crime, Religion, Royalty, Awards & Ceremonies

ВКЛЮЧАЙ: точные названия мест, имена, организации, политические темы, контекст событий

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."], "category": "..."}`,

		"commercial": `Создай метаданные для коммерческого стокового фото:

ТРЕБОВАНИЯ ДЛЯ COMMERCIAL:
1. НАЗВАНИЕ (до 70 символов):
   - Описательное без конкретных имен и мест
   - Концептуальное (бизнес, семья, технологии)
   - Эмоциональное состояние (счастье, успех)
   - Универсальные формулировки

2. ОПИСАНИЕ (до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов

3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
   - Люди: возраст, пол, количество, роли (избегай конкретных имен)
   - Эмоции: какие эмоции выражают люди или передает изображение
   - Концепции: какие идеи, понятия иллюстрирует фото
   - Визуальные характеристики: стиль, цвета, композиция, освещение
   - Локации: тип места без конкретных названий (office, home, etc.)
   - Действия и активности: что происходит на фото
   - Объекты и предметы: что видишь на изображении
   - Настроение и атмосфера: общее впечатление от фото

4. КАТЕГОРИЯ (выбери одну из для Commercial):
   Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture, Music, Art & Design, Objects, Concepts, Beauty, Shopping, Transportation, Home & Garden

ИЗБЕГАЙ: конкретные имена людей/компаний, бренды, конкретные места, даты

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."], "category": "..."}`,
	}

	settings.AIPrompts = defaultPrompts
	err = a.dbService.SaveSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to save updated prompts: %w", err)
	}

	log.Println("Default AI prompts have been updated successfully")
	return nil
}

// GetProcessingHistory возвращает историю обработанных батчей
func (a *App) GetProcessingHistory(limit int) ([]models.PhotoBatch, error) {
	return a.dbService.GetBatchHistory(limit)
}

// StartQueueProcessing запускает обработку очереди
func (a *App) StartQueueProcessing() error {
	settings, err := a.dbService.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	return a.queueManager.StartProcessing(settings)
}

// StopQueueProcessing останавливает обработку очереди
func (a *App) StopQueueProcessing() error {
	a.queueManager.StopProcessing()
	return nil
}

// GetBatchDetails возвращает детали конкретного батча
func (a *App) GetBatchDetails(batchID string) (*models.PhotoBatch, error) {
	batches, err := a.dbService.GetBatchHistory(100)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch history: %w", err)
	}

	for _, batch := range batches {
		if batch.ID == batchID {
			return &batch, nil
		}
	}

	return nil, fmt.Errorf("batch not found: %s", batchID)
}

// ApprovePhoto подтверждает фото для загрузки и записывает EXIF метаданные
func (a *App) ApprovePhoto(photoID string) error {
	log.Printf("ApprovePhoto called for photoID: %s", photoID)

	// Получаем данные фото и его AI результаты
	var originalPath, aiResultJSON string
	err := a.db.QueryRow(`
		SELECT original_path, ai_results 
		FROM photos 
		WHERE id = ?`, photoID).Scan(&originalPath, &aiResultJSON)
	if err != nil {
		log.Printf("Failed to get photo data for %s: %v", photoID, err)
		return fmt.Errorf("failed to get photo data: %w", err)
	}

	log.Printf("Photo data retrieved - originalPath: %s, aiResultJSON length: %d", originalPath, len(aiResultJSON))

	// Обновляем статус на approved
	_, err = a.db.Exec(`
		UPDATE photos 
		SET status = 'approved', updated_at = datetime('now') 
		WHERE id = ?`, photoID)
	if err != nil {
		log.Printf("Failed to update photo status for %s: %v", photoID, err)
		return fmt.Errorf("failed to approve photo: %w", err)
	}

	log.Printf("Photo %s status updated to approved", photoID)

	// Записываем EXIF метаданные в оригинальный файл
	if aiResultJSON != "" && originalPath != "" {
		log.Printf("Starting EXIF write for %s", originalPath)

		var aiResult models.AIResult
		err = json.Unmarshal([]byte(aiResultJSON), &aiResult)
		if err != nil {
			log.Printf("Warning: failed to unmarshal AI result for photo %s: %v", photoID, err)
		} else {
			log.Printf("AI result unmarshaled successfully, writing EXIF...")
			err = a.imageProc.WriteExifToImage(originalPath, aiResult)
			if err != nil {
				log.Printf("Warning: failed to write EXIF to %s: %v", originalPath, err)
			} else {
				log.Printf("EXIF metadata written successfully to %s", originalPath)
			}
		}
	} else {
		log.Printf("Skipping EXIF write - aiResultJSON empty: %t, originalPath empty: %t",
			aiResultJSON == "", originalPath == "")
	}

	log.Printf("Photo %s approved for upload successfully", photoID)
	return nil
}

// RejectPhoto отклоняет фото от загрузки
func (a *App) RejectPhoto(photoID string) error {
	_, err := a.db.Exec(`
		UPDATE photos 
		SET status = 'rejected', updated_at = datetime('now') 
		WHERE id = ?`, photoID)
	if err != nil {
		return fmt.Errorf("failed to reject photo: %w", err)
	}

	log.Printf("Photo %s rejected", photoID)
	return nil
}

// ResetPhotoToProcessed сбрасывает статус фото с approved обратно на processed
// Это нужно для случаев когда пользователь вручную редактирует метаданные
// и хочет снова нажать approve для записи EXIF
func (a *App) ResetPhotoToProcessed(photoID string) error {
	_, err := a.db.Exec(`
		UPDATE photos 
		SET status = 'processed', updated_at = datetime('now') 
		WHERE id = ?`, photoID)
	if err != nil {
		return fmt.Errorf("failed to reset photo to processed: %w", err)
	}

	log.Printf("Photo %s reset to processed status", photoID)
	return nil
}

// UpdatePhotoMetadata обновляет метаданные фото
func (a *App) UpdatePhotoMetadata(photoID string, aiResult models.AIResult) error {
	err := a.dbService.UpdatePhotoAIResults(photoID, aiResult)
	if err != nil {
		return fmt.Errorf("failed to update photo metadata: %w", err)
	}

	log.Printf("Photo %s metadata updated", photoID)
	return nil
}

// RegeneratePhotoMetadata повторно генерирует метаданные для фото
func (a *App) RegeneratePhotoMetadata(photoID string, customPrompt string) error {
	// Получаем данные фото
	var photo models.Photo
	var exifJSON, uploadStatusJSON string

	err := a.db.QueryRow(`
		SELECT id, batch_id, original_path, thumbnail_path, file_name, file_size,
		       exif_data, upload_status, status, created_at
		FROM photos WHERE id = ?`, photoID).Scan(
		&photo.ID, &photo.BatchID, &photo.OriginalPath, &photo.ThumbnailPath,
		&photo.FileName, &photo.FileSize, &exifJSON, &uploadStatusJSON,
		&photo.Status, &photo.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to get photo data: %w", err)
	}

	// Десериализуем EXIF данные
	if exifJSON != "" {
		err = json.Unmarshal([]byte(exifJSON), &photo.ExifData)
		if err != nil {
			photo.ExifData = make(map[string]string)
		}
	}

	// Получаем настройки
	settings, err := a.dbService.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to get settings: %w", err)
	}

	// Получаем тип батча и описание
	var batchType, batchDescription string
	err = a.db.QueryRow("SELECT type, description FROM batches WHERE id = ?", photo.BatchID).Scan(&batchType, &batchDescription)
	if err != nil {
		return fmt.Errorf("failed to get batch info: %w", err)
	}

	photo.ContentType = batchType

	// Проверяем существование thumbnail и пересоздаем если нужно
	if photo.ThumbnailPath == "" || !fileExists(photo.ThumbnailPath) {
		log.Printf("Thumbnail missing for photo %s, recreating...", photo.ID)

		// Пересоздаем thumbnail
		err = a.imageProc.ProcessPhotoForAI(&photo, settings.ThumbnailSize)
		if err != nil {
			return fmt.Errorf("failed to prepare photo for AI: %w", err)
		}

		// Обновляем thumbnail_path в базе данных
		err = a.dbService.UpdatePhotoThumbnail(photo.ID, photo.ThumbnailPath)
		if err != nil {
			log.Printf("Warning: failed to update thumbnail_path in database: %v", err)
		}
	}

	// Используем кастомный промпт если предоставлен
	if customPrompt != "" {
		// Временно заменяем промпт в настройках
		if settings.AIPrompts == nil {
			settings.AIPrompts = make(map[string]string)
		}
		originalPrompt := settings.AIPrompts[batchType]
		settings.AIPrompts[batchType] = customPrompt

		// Анализируем фото
		aiResult, err := a.aiService.AnalyzePhoto(photo, batchDescription, batchType, settings)

		// Восстанавливаем оригинальный промпт
		settings.AIPrompts[batchType] = originalPrompt

		if err != nil {
			return fmt.Errorf("failed to regenerate metadata: %w", err)
		}

		// Сохраняем результаты
		return a.UpdatePhotoMetadata(photoID, *aiResult)
	} else {
		// Используем стандартный промпт
		aiResult, err := a.aiService.AnalyzePhoto(photo, batchDescription, batchType, settings)
		if err != nil {
			return fmt.Errorf("failed to regenerate metadata: %w", err)
		}

		// Сохраняем результаты
		return a.UpdatePhotoMetadata(photoID, *aiResult)
	}
}

// SetPhotoStatus устанавливает статус фотографии
func (a *App) SetPhotoStatus(photoID string, status string) error {
	log.Printf("Setting photo %s status to %s", photoID, status)

	// Проверяем валидный статус
	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"processed":  true,
		"approved":   true,
		"rejected":   true,
		"failed":     true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Обновляем статус в базе данных
	_, err := a.db.Exec("UPDATE photos SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", status, photoID)
	if err != nil {
		return fmt.Errorf("failed to update photo status: %w", err)
	}

	// Логируем событие
	if a.dbService != nil {
		a.dbService.LogEvent("", photoID, "photo_status_changed", status, fmt.Sprintf("Photo %s status changed to %s", photoID, status), "", 0)
	}

	return nil
}

// GetDefaultLanguage возвращает сохраненный язык или "en" по умолчанию
func (a *App) GetDefaultLanguage() string {
	settings, err := a.dbService.GetSettings()
	if err != nil {
		log.Printf("Failed to get language from settings: %v", err)
		return "en"
	}

	if settings.Language == "" {
		return "en"
	}

	return settings.Language
}

// GetAIModels возвращает список доступных моделей для указанного провайдера
func (a *App) GetAIModels(provider string) ([]models.AIModel, error) {
	settings, err := a.dbService.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return a.aiService.GetAvailableModels(provider, settings)
}

// SelectFolder открывает диалог выбора папки
func (a *App) SelectFolder() (string, error) {
	folderPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select photo folder",
	})

	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}

	return folderPath, nil
}

// GetFolderContents возвращает список изображений в папке
func (a *App) GetFolderContents(folderPath string) ([]models.PhotoFile, error) {
	if folderPath == "" {
		return nil, fmt.Errorf("folder path is empty")
	}

	// Используем imageProcessor для сканирования папки
	files, err := a.imageProc.ScanFolderFiles(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan folder: %w", err)
	}

	return files, nil
}

// GetStockTemplates возвращает шаблоны для создания стоков
func (a *App) GetStockTemplates() map[string]models.StockTemplate {
	return a.uploaderManager.GetStockTemplates()
}

// GetAvailableUploaders возвращает список доступных загрузчиков
func (a *App) GetAvailableUploaders() []models.UploaderInfo {
	return a.uploaderManager.GetAvailableUploaders()
}

// ValidateStockConfig валидирует конфигурацию стока
func (a *App) ValidateStockConfig(config models.StockConfig) error {
	return a.uploaderManager.ValidateStockConfig(config)
}

func getCurrentTimestamp() int64 {
	return 1674123456 // Заглушка, используйте time.Now().Unix()
}

// fileExists проверяет существование файла
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetProcessedBatches возвращает батчи готовые к ревью (со статусом processed)
func (a *App) GetProcessedBatches() ([]models.PhotoBatch, error) {
	query := `
		SELECT DISTINCT b.id, b.type, b.description, b.folder_path, b.status,
		       b.created_at, b.updated_at,
		       COUNT(p.id) as total_photos,
		       COUNT(CASE WHEN p.status = 'processed' THEN 1 END) as processed_photos,
		       COUNT(CASE WHEN p.status = 'approved' THEN 1 END) as approved_photos,
		       COUNT(CASE WHEN p.status = 'rejected' THEN 1 END) as rejected_photos
		FROM batches b
		LEFT JOIN photos p ON b.id = p.batch_id
		WHERE b.status = 'processed'
		GROUP BY b.id, b.type, b.description, b.folder_path, b.status, b.created_at, b.updated_at
		ORDER BY b.updated_at DESC`

	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query processed batches: %w", err)
	}
	defer rows.Close()

	var batches []models.PhotoBatch

	for rows.Next() {
		var batch models.PhotoBatch
		var totalPhotos, processedPhotos, approvedPhotos, rejectedPhotos int

		err := rows.Scan(
			&batch.ID, &batch.Type, &batch.Description, &batch.FolderPath,
			&batch.Status, &batch.CreatedAt, &batch.UpdatedAt,
			&totalPhotos, &processedPhotos, &approvedPhotos, &rejectedPhotos,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch row: %w", err)
		}

		// Добавляем статистику фото
		batch.PhotosStats = &models.BatchStats{
			Total:     totalPhotos,
			Processed: processedPhotos,
			Approved:  approvedPhotos,
			Rejected:  rejectedPhotos,
		}

		batches = append(batches, batch)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating batch rows: %w", err)
	}

	log.Printf("Found %d processed batches for review", len(batches))
	return batches, nil
}

// GetPhotoThumbnail возвращает thumbnail фото в виде base64
func (a *App) GetPhotoThumbnail(photoID string) (string, error) {
	// Получаем thumbnailPath из базы данных
	var thumbnailPath string
	err := a.db.QueryRow("SELECT thumbnail_path FROM photos WHERE id = ?", photoID).Scan(&thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("failed to get thumbnail path: %w", err)
	}

	// Если thumbnailPath пустой, возвращаем пустую строку
	if thumbnailPath == "" {
		return "", fmt.Errorf("no thumbnail available for photo %s", photoID)
	}

	// Проверяем существование файла
	if !fileExists(thumbnailPath) {
		return "", fmt.Errorf("thumbnail file not found: %s", thumbnailPath)
	}

	// Читаем файл и кодируем в base64
	imageBytes, err := os.ReadFile(thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("failed to read thumbnail file: %w", err)
	}

	// Возвращаем в формате data URL для прямого использования в img src
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)
	return fmt.Sprintf("data:image/jpeg;base64,%s", base64Data), nil
}

// GetBatchPhotos возвращает все фото из конкретного батча для ревью
func (a *App) GetBatchPhotos(batchID string) ([]models.Photo, error) {
	query := `
		SELECT id, batch_id, original_path, thumbnail_path, file_name, file_size,
		       ai_results, exif_data, upload_status, status, created_at, updated_at
		FROM photos 
		WHERE batch_id = ?
		ORDER BY file_name ASC`

	rows, err := a.db.Query(query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch photos: %w", err)
	}
	defer rows.Close()

	var photos []models.Photo

	for rows.Next() {
		var photo models.Photo
		var aiResultsJSON, exifJSON, uploadStatusJSON string
		var updatedAt sql.NullTime

		err := rows.Scan(
			&photo.ID, &photo.BatchID, &photo.OriginalPath, &photo.ThumbnailPath,
			&photo.FileName, &photo.FileSize, &aiResultsJSON, &exifJSON,
			&uploadStatusJSON, &photo.Status, &photo.CreatedAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan photo row: %w", err)
		}

		// Обработка updated_at поля (может быть NULL)
		if updatedAt.Valid {
			photo.UpdatedAt = updatedAt.Time
		} else {
			photo.UpdatedAt = photo.CreatedAt
		}

		// Десериализуем AI результаты
		if aiResultsJSON != "" {
			err = json.Unmarshal([]byte(aiResultsJSON), &photo.AIResult)
			if err != nil {
				log.Printf("Warning: failed to unmarshal AI results for photo %s: %v", photo.ID, err)
				photo.AIResult = nil
			}
		}

		// Десериализуем EXIF данные
		if exifJSON != "" {
			err = json.Unmarshal([]byte(exifJSON), &photo.ExifData)
			if err != nil {
				log.Printf("Warning: failed to unmarshal EXIF data for photo %s: %v", photo.ID, err)
				photo.ExifData = make(map[string]string)
			}
		} else {
			photo.ExifData = make(map[string]string)
		}

		// Десериализуем статус загрузки
		if uploadStatusJSON != "" {
			err = json.Unmarshal([]byte(uploadStatusJSON), &photo.UploadStatus)
			if err != nil {
				log.Printf("Warning: failed to unmarshal upload status for photo %s: %v", photo.ID, err)
				photo.UploadStatus = make(map[string]string)
			}
		} else {
			photo.UploadStatus = make(map[string]string)
		}

		photos = append(photos, photo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating photo rows: %w", err)
	}

	log.Printf("Found %d photos in batch %s", len(photos), batchID)
	return photos, nil
}

// UploadApprovedPhotos загружает одобренные фото на все активные стоки
func (a *App) UploadApprovedPhotos(batchID string) error {
	// Получаем информацию о батче
	var batchType string
	err := a.db.QueryRow("SELECT type FROM batches WHERE id = ?", batchID).Scan(&batchType)
	if err != nil {
		return fmt.Errorf("failed to get batch type: %w", err)
	}

	// Получаем одобренные фото из батча
	rows, err := a.db.Query(`
		SELECT id, original_path, file_name, ai_results 
		FROM photos 
		WHERE batch_id = ? AND status = 'approved'`, batchID)
	if err != nil {
		return fmt.Errorf("failed to get approved photos: %w", err)
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var photo models.Photo
		var aiResultsJSON string

		err := rows.Scan(&photo.ID, &photo.OriginalPath, &photo.FileName, &aiResultsJSON)
		if err != nil {
			continue
		}

		// Десериализуем AI результаты
		if aiResultsJSON != "" {
			var aiResult models.AIResult
			if json.Unmarshal([]byte(aiResultsJSON), &aiResult) == nil {
				photo.AIResult = &aiResult
			}
		}

		photo.BatchID = batchID
		photo.ContentType = batchType
		photos = append(photos, photo)
	}

	if len(photos) == 0 {
		return fmt.Errorf("no approved photos found in batch")
	}

	// Получаем активные стоковые конфигурации для данного типа
	stockConfigs, err := a.dbService.GetActiveStockConfigs(batchType)
	if err != nil {
		return fmt.Errorf("failed to get stock configs: %w", err)
	}

	if len(stockConfigs) == 0 {
		return fmt.Errorf("no active stock configurations found for type: %s", batchType)
	}

	log.Printf("Starting upload of %d photos to %d stocks", len(photos), len(stockConfigs))

	// Загружаем на каждый сток
	for _, stockConfig := range stockConfigs {
		for _, photo := range photos {
			go func(stock models.StockConfig, p models.Photo) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Panic in upload goroutine for photo %s to stock %s: %v", p.FileName, stock.Name, r)
						if err := a.dbService.UpdatePhotoUploadStatus(p.ID, stock.ID, "failed"); err != nil {
							log.Printf("Failed to update upload status after panic: %v", err)
						}
					}
				}()

				err := a.uploadPhotoToStock(p, stock)
				if err != nil {
					log.Printf("Failed to upload photo %s to stock %s: %v", p.FileName, stock.Name, err)
					if updateErr := a.dbService.UpdatePhotoUploadStatus(p.ID, stock.ID, "failed"); updateErr != nil {
						log.Printf("Failed to update upload status to failed: %v", updateErr)
					}
				} else {
					log.Printf("Successfully uploaded photo %s to stock %s", p.FileName, stock.Name)
					if updateErr := a.dbService.UpdatePhotoUploadStatus(p.ID, stock.ID, "uploaded"); updateErr != nil {
						log.Printf("Failed to update upload status to uploaded: %v", updateErr)
					}
				}
			}(stockConfig, photo)
		}
	}

	return nil
}

// uploadPhotoToStock загружает одно фото на конкретный сток
func (a *App) uploadPhotoToStock(photo models.Photo, stockConfig models.StockConfig) error {
	// Используем uploaderManager для загрузки
	uploader, err := a.uploaderManager.GetUploader(stockConfig.Type)
	if err != nil {
		return fmt.Errorf("failed to get uploader for %s: %w", stockConfig.Type, err)
	}

	// Обновляем статус на "uploading"
	a.dbService.UpdatePhotoUploadStatus(photo.ID, stockConfig.ID, "uploading")

	// Выполняем загрузку
	result, err := uploader.Upload(photo, stockConfig)
	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	log.Printf("Upload result for %s to %s: %+v", photo.FileName, stockConfig.Name, result)

	return nil
}

// DeleteBatch удаляет батч и все связанные фото
func (a *App) DeleteBatch(batchID string) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем все фото батча
	_, err = tx.Exec("DELETE FROM photos WHERE batch_id = ?", batchID)
	if err != nil {
		return fmt.Errorf("failed to delete photos: %w", err)
	}

	// Удаляем сам батч
	_, err = tx.Exec("DELETE FROM batches WHERE id = ?", batchID)
	if err != nil {
		return fmt.Errorf("failed to delete batch: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Batch %s deleted successfully", batchID)
	return nil
}

// GetUploadProgress возвращает прогресс загрузки для батча
func (a *App) GetUploadProgress(batchID string) (map[string]interface{}, error) {
	// Получаем все фото из батча
	rows, err := a.db.Query(`
		SELECT id, file_name, upload_status 
		FROM photos 
		WHERE batch_id = ? AND status = 'approved'`, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos: %w", err)
	}
	defer rows.Close()

	type PhotoProgress struct {
		ID       string            `json:"id"`
		FileName string            `json:"fileName"`
		Stocks   map[string]string `json:"stocks"`
	}

	var photos []PhotoProgress
	totalPhotos := 0
	uploadingCount := 0
	uploadedCount := 0
	failedCount := 0

	for rows.Next() {
		var photo PhotoProgress
		var uploadStatusJSON string

		err := rows.Scan(&photo.ID, &photo.FileName, &uploadStatusJSON)
		if err != nil {
			continue
		}

		// Десериализуем статус загрузки
		photo.Stocks = make(map[string]string)
		if uploadStatusJSON != "" && uploadStatusJSON != "null" {
			json.Unmarshal([]byte(uploadStatusJSON), &photo.Stocks)
		}

		// Подсчитываем статистику
		totalPhotos++
		for _, status := range photo.Stocks {
			switch status {
			case "uploading":
				uploadingCount++
			case "uploaded":
				uploadedCount++
			case "failed":
				failedCount++
			}
		}

		photos = append(photos, photo)
	}

	return map[string]interface{}{
		"photos":         photos,
		"totalPhotos":    totalPhotos,
		"uploadingCount": uploadingCount,
		"uploadedCount":  uploadedCount,
		"failedCount":    failedCount,
	}, nil
}

// GetBatchEvents возвращает события для батча
func (a *App) GetBatchEvents(batchID string, limit int) ([]models.EventLog, error) {
	log.Printf("Getting events for batch %s (limit: %d)", batchID, limit)
	return a.dbService.GetBatchEvents(batchID, limit)
}

// GetPhotoEvents возвращает события для фотографии
func (a *App) GetPhotoEvents(photoID string) ([]models.EventLog, error) {
	log.Printf("Getting events for photo %s", photoID)
	return a.dbService.GetPhotoEvents(photoID)
}

// GetProcessingProgress возвращает текущий прогресс обработки батча
func (a *App) GetProcessingProgress(batchID string) (models.ProcessingProgress, error) {
	log.Printf("Getting processing progress for batch %s", batchID)

	// Получаем последние события
	events, err := a.dbService.GetBatchEvents(batchID, 10)
	if err != nil {
		log.Printf("Warning: failed to get events: %v", err)
		events = []models.EventLog{}
	}

	// Получаем статистику батча из базы данных
	var status string
	var totalPhotos int
	err = a.db.QueryRow(`
		SELECT status, 
		       (SELECT COUNT(*) FROM photos WHERE batch_id = ?)
		FROM batches WHERE id = ?`, batchID, batchID).Scan(&status, &totalPhotos)
	if err != nil {
		return models.ProcessingProgress{}, fmt.Errorf("failed to get batch info: %w", err)
	}

	// Формируем ответ
	progress := models.ProcessingProgress{
		BatchID:      batchID,
		TotalPhotos:  totalPhotos,
		RecentEvents: events,
		Status:       status,
	}

	// Определяем прогресс на основе статуса
	switch status {
	case "pending":
		progress.OverallProgress = 0
		progress.CurrentStep = "waiting"
	case "processing":
		progress.CurrentStep = "ai_processing"
		// Подсчитываем прогресс по обработанным фото
		var processedCount int
		a.db.QueryRow(`SELECT COUNT(*) FROM photos WHERE batch_id = ? AND status IN ('processed', 'failed')`, batchID).Scan(&processedCount)
		if totalPhotos > 0 {
			progress.OverallProgress = (processedCount * 100) / totalPhotos
		}
	case "processed":
		progress.Status = "completed"
		progress.OverallProgress = 100
		progress.CurrentStep = "completed"
	default:
		progress.Status = status
		progress.CurrentStep = "unknown"
	}

	return progress, nil
}

// CheckExifToolStatus проверяет статус ExifTool
func (a *App) CheckExifToolStatus() map[string]interface{} {
	available := a.imageProc.CheckExifToolAvailable()

	result := map[string]interface{}{
		"available": available,
		"message":   "",
	}

	if !available {
		result["message"] = "ExifTool не установлен. Метаданные не будут записываться в EXIF файлов. Установите ExifTool для полной функциональности."
	} else {
		result["message"] = "ExifTool доступен. Метаданные будут автоматически записываться в EXIF файлов."
	}

	return result
}
