package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"stock-photo-app/models"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService(db *sql.DB) *DatabaseService {
	return &DatabaseService{db: db}
}

// InitializeTables создает необходимые таблицы в БД
func (d *DatabaseService) InitializeTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS batches (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			description TEXT,
			folder_path TEXT,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT (datetime('now')),
			updated_at DATETIME DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS photos (
			id TEXT PRIMARY KEY,
			batch_id TEXT,
			original_path TEXT NOT NULL,
			thumbnail_path TEXT,
			file_name TEXT,
			file_size INTEGER,
			exif_data TEXT, -- JSON
			ai_results TEXT, -- JSON
			upload_status TEXT, -- JSON
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT (datetime('now')),
			FOREIGN KEY (batch_id) REFERENCES batches(id)
		)`,

		`CREATE TABLE IF NOT EXISTS stock_configs (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			supported_types TEXT NOT NULL, -- JSON array
			upload_method TEXT, -- deprecated, use type instead
			connection_config TEXT NOT NULL, -- JSON object
			prompts TEXT, -- JSON object
			settings TEXT, -- JSON object
			module_path TEXT,
			active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS app_settings (
			id TEXT PRIMARY KEY DEFAULT 'main',
			temp_directory TEXT,
			ai_provider TEXT DEFAULT 'openai',
			ai_model TEXT DEFAULT 'gpt-4o',
			ai_api_key TEXT,
			ai_base_url TEXT,
			max_concurrent_jobs INTEGER DEFAULT 3,
			thumbnail_size INTEGER DEFAULT 512,
			language TEXT DEFAULT 'en',
			ai_prompts TEXT, -- JSON
			updated_at DATETIME DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS event_logs (
			id TEXT PRIMARY KEY,
			batch_id TEXT,
			photo_id TEXT,
			event_type TEXT NOT NULL,
			status TEXT NOT NULL,
			message TEXT NOT NULL,
			details TEXT,
			progress INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (batch_id) REFERENCES batches(id) ON DELETE CASCADE,
			FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_photos_batch_id ON photos(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_batches_status ON batches(status)`,
		`CREATE INDEX IF NOT EXISTS idx_photos_status ON photos(status)`,
		`CREATE INDEX IF NOT EXISTS idx_event_logs_batch_id ON event_logs(batch_id)`,
		`CREATE INDEX IF NOT EXISTS idx_event_logs_photo_id ON event_logs(photo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_event_logs_created_at ON event_logs(created_at)`,
	}

	for _, query := range queries {
		_, err := d.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	// Применяем миграции
	if err := d.applyMigrations(); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Создаем дефолтные настройки
	if err := d.createDefaultSettings(); err != nil {
		return fmt.Errorf("failed to create default settings: %w", err)
	}

	// Создаем демо конфигурацию стока (только если нет других активных стоков)
	// if err := d.createDemoStockConfig(); err != nil {
	//	return fmt.Errorf("failed to create demo stock config: %w", err)
	// }

	log.Println("Database tables initialized successfully")
	return nil
}

// SaveBatch сохраняет батч в БД
func (d *DatabaseService) SaveBatch(batch models.PhotoBatch) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Удаляем старые фотографии для этого батча (если батч перезаписывается)
	result, err := tx.Exec("DELETE FROM photos WHERE batch_id = ?", batch.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old photos: %w", err)
	}

	if deletedRows, _ := result.RowsAffected(); deletedRows > 0 {
		log.Printf("DEBUG: Deleted %d old photos for batch %s", deletedRows, batch.ID)
	}

	// Сохраняем батч
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO batches 
		(id, type, description, folder_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		batch.ID, batch.Type, batch.Description, batch.FolderPath,
		batch.Status, batch.CreatedAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save batch: %w", err)
	}

	// Сохраняем фотографии
	for _, photo := range batch.Photos {
		photo.BatchID = batch.ID
		err = d.savePhoto(tx, photo)
		if err != nil {
			return fmt.Errorf("failed to save photo: %w", err)
		}
	}

	return tx.Commit()
}

// savePhoto сохраняет фото в рамках транзакции
func (d *DatabaseService) savePhoto(tx *sql.Tx, photo models.Photo) error {
	exifJSON, _ := json.Marshal(photo.ExifData)
	var aiResultsJSON []byte
	if photo.AIResult != nil {
		aiResultsJSON, _ = json.Marshal(photo.AIResult)
	}
	uploadStatusJSON, _ := json.Marshal(photo.UploadStatus)

	_, err := tx.Exec(`
		INSERT OR REPLACE INTO photos 
		(id, batch_id, content_type, original_path, thumbnail_path, file_name, file_size, 
		 exif_data, ai_results, upload_status, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		photo.ID, photo.BatchID, photo.ContentType, photo.OriginalPath, photo.ThumbnailPath,
		photo.FileName, photo.FileSize, string(exifJSON), string(aiResultsJSON),
		string(uploadStatusJSON), photo.Status, photo.CreatedAt)

	return err
}

// UpdatePhotoAIResults обновляет результаты AI для фото
func (d *DatabaseService) UpdatePhotoAIResults(photoID string, aiResults models.AIResult) error {
	aiResultsJSON, err := json.Marshal(aiResults)
	if err != nil {
		return fmt.Errorf("failed to marshal AI results: %w", err)
	}

	_, err = d.db.Exec(`
		UPDATE photos 
		SET ai_results = ?, status = 'processed', updated_at = datetime('now') 
		WHERE id = ?`,
		string(aiResultsJSON), photoID)

	return err
}

// UpdatePhotoThumbnail обновляет thumbnail path для фото
func (d *DatabaseService) UpdatePhotoThumbnail(photoID string, thumbnailPath string) error {
	_, err := d.db.Exec(`
		UPDATE photos 
		SET thumbnail_path = ?, updated_at = datetime('now') 
		WHERE id = ?`,
		thumbnailPath, photoID)

	if err != nil {
		return fmt.Errorf("failed to update thumbnail path: %w", err)
	}

	return nil
}

// UpdatePhotoUploadStatus обновляет статус загрузки фото
func (d *DatabaseService) UpdatePhotoUploadStatus(photoID string, stockID string, status string) error {
	// Получаем текущий статус
	var uploadStatusJSON string
	err := d.db.QueryRow("SELECT upload_status FROM photos WHERE id = ?", photoID).Scan(&uploadStatusJSON)
	if err != nil {
		return fmt.Errorf("failed to get current upload status: %w", err)
	}

	// Всегда инициализируем карту
	uploadStatus := make(map[string]string)

	// Если есть существующие данные, пытаемся их загрузить
	if uploadStatusJSON != "" && uploadStatusJSON != "null" {
		err = json.Unmarshal([]byte(uploadStatusJSON), &uploadStatus)
		if err != nil {
			log.Printf("Warning: failed to unmarshal upload status for photo %s: %v", photoID, err)
			// Если не удалось разобрать, создаем новую пустую карту
			uploadStatus = make(map[string]string)
		}
	}

	// Проверяем что карта не nil (дополнительная защита)
	if uploadStatus == nil {
		uploadStatus = make(map[string]string)
	}

	// Обновляем статус для конкретного стока
	uploadStatus[stockID] = status

	// Сохраняем обратно
	newStatusJSON, err := json.Marshal(uploadStatus)
	if err != nil {
		return fmt.Errorf("failed to marshal upload status: %w", err)
	}

	_, err = d.db.Exec(`
		UPDATE photos 
		SET upload_status = ?, updated_at = datetime('now') 
		WHERE id = ?`,
		string(newStatusJSON), photoID)

	if err != nil {
		return fmt.Errorf("failed to update upload status in database: %w", err)
	}

	log.Printf("Updated upload status for photo %s, stock %s: %s", photoID, stockID, status)
	return nil
}

// UpdatePhotoUploadQueueStatus обновляет общий статус фотографии в очереди загрузки
func (d *DatabaseService) UpdatePhotoUploadQueueStatus(photoID string, status string) error {
	_, err := d.db.Exec(`
		UPDATE photos 
		SET status = ?, updated_at = datetime('now') 
		WHERE id = ?`,
		status, photoID)

	if err != nil {
		return fmt.Errorf("failed to update photo queue status: %w", err)
	}

	log.Printf("Updated photo %s queue status to: %s", photoID, status)
	return nil
}

// GetBatchHistory возвращает историю обработанных батчей
func (d *DatabaseService) GetBatchHistory(limit int) ([]models.PhotoBatch, error) {
	rows, err := d.db.Query(`
		SELECT id, type, description, folder_path, status, created_at, updated_at
		FROM batches 
		ORDER BY created_at DESC 
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query batches: %w", err)
	}
	defer rows.Close()

	var batches []models.PhotoBatch
	for rows.Next() {
		var batch models.PhotoBatch
		err := rows.Scan(&batch.ID, &batch.Type, &batch.Description,
			&batch.FolderPath, &batch.Status, &batch.CreatedAt, &batch.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batch: %w", err)
		}

		// Загружаем фотографии для каждого батча
		photos, err := d.getPhotosForBatch(batch.ID)
		if err != nil {
			log.Printf("Failed to load photos for batch %s: %v", batch.ID, err)
			continue
		}
		batch.Photos = photos

		batches = append(batches, batch)
	}

	return batches, nil
}

// getPhotosForBatch возвращает фотографии для конкретного батча
func (d *DatabaseService) getPhotosForBatch(batchID string) ([]models.Photo, error) {
	rows, err := d.db.Query(`
		SELECT id, batch_id, content_type, original_path, thumbnail_path, file_name, file_size,
		       exif_data, ai_results, upload_status, status, created_at
		FROM photos 
		WHERE batch_id = ?`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var photo models.Photo
		var exifJSON, aiResultsJSON, uploadStatusJSON string

		err := rows.Scan(&photo.ID, &photo.BatchID, &photo.ContentType, &photo.OriginalPath,
			&photo.ThumbnailPath, &photo.FileName, &photo.FileSize,
			&exifJSON, &aiResultsJSON, &uploadStatusJSON,
			&photo.Status, &photo.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Десериализуем JSON поля
		if exifJSON != "" {
			json.Unmarshal([]byte(exifJSON), &photo.ExifData)
		}
		if aiResultsJSON != "" {
			var aiResult models.AIResult
			err := json.Unmarshal([]byte(aiResultsJSON), &aiResult)
			if err == nil {
				photo.AIResult = &aiResult
			}
		}
		if uploadStatusJSON != "" {
			json.Unmarshal([]byte(uploadStatusJSON), &photo.UploadStatus)
		}

		photos = append(photos, photo)
	}

	return photos, nil
}

// SaveStockConfig сохраняет конфигурацию стока
func (d *DatabaseService) SaveStockConfig(config models.StockConfig) error {
	supportedTypesJSON, _ := json.Marshal(config.SupportedTypes)
	connectionJSON, _ := json.Marshal(config.Connection)
	promptsJSON, _ := json.Marshal(config.Prompts)
	settingsJSON, _ := json.Marshal(config.Settings)

	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO stock_configs 
		(id, name, type, supported_types, upload_method, connection_config, prompts, settings, module_path, active, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		config.ID, config.Name, config.Type, string(supportedTypesJSON), config.UploadMethod,
		string(connectionJSON), string(promptsJSON), string(settingsJSON), config.ModulePath, config.Active, time.Now())

	return err
}

// GetAllStockConfigs возвращает все конфигурации стоков
func (d *DatabaseService) GetAllStockConfigs() ([]models.StockConfig, error) {
	rows, err := d.db.Query(`
		SELECT id, name, type, supported_types, upload_method, connection_config, 
		       prompts, settings, module_path, active, created_at, updated_at
		FROM stock_configs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []models.StockConfig
	for rows.Next() {
		var config models.StockConfig
		var supportedTypesJSON, connectionJSON, promptsJSON, settingsJSON string
		var uploadMethod, modulePath sql.NullString

		err := rows.Scan(&config.ID, &config.Name, &config.Type, &supportedTypesJSON,
			&uploadMethod, &connectionJSON, &promptsJSON,
			&settingsJSON, &modulePath, &config.Active,
			&config.CreatedAt, &config.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Обрабатываем nullable поля
		config.UploadMethod = uploadMethod.String
		config.ModulePath = modulePath.String

		// Десериализуем JSON поля
		json.Unmarshal([]byte(supportedTypesJSON), &config.SupportedTypes)
		json.Unmarshal([]byte(connectionJSON), &config.Connection)
		json.Unmarshal([]byte(promptsJSON), &config.Prompts)
		if settingsJSON != "" {
			json.Unmarshal([]byte(settingsJSON), &config.Settings)
		}

		configs = append(configs, config)
	}

	return configs, nil
}

// GetActiveStockConfigs возвращает активные конфигурации для указанного типа
func (d *DatabaseService) GetActiveStockConfigs(photoType string) ([]models.StockConfig, error) {
	allConfigs, err := d.GetAllStockConfigs()
	if err != nil {
		return nil, err
	}

	var activeConfigs []models.StockConfig
	for _, config := range allConfigs {
		if !config.Active {
			continue
		}

		// Проверяем поддерживается ли тип фото
		for _, supportedType := range config.SupportedTypes {
			if supportedType == photoType {
				activeConfigs = append(activeConfigs, config)
				break
			}
		}
	}

	return activeConfigs, nil
}

// DeleteStockConfig удаляет конфигурацию стока
func (d *DatabaseService) DeleteStockConfig(stockID string) error {
	log.Printf("Attempting to delete stock config with ID: %s", stockID)

	// Проверяем существует ли такая запись
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM stock_configs WHERE id = ?", stockID).Scan(&count)
	if err != nil {
		log.Printf("Error checking if stock exists: %v", err)
		return fmt.Errorf("failed to check if stock exists: %w", err)
	}

	if count == 0 {
		log.Printf("Stock config with ID %s not found", stockID)
		return fmt.Errorf("stock config with ID %s not found", stockID)
	}

	log.Printf("Found stock config, proceeding with deletion...")

	result, err := d.db.Exec("DELETE FROM stock_configs WHERE id = ?", stockID)
	if err != nil {
		log.Printf("Error deleting stock config: %v", err)
		return fmt.Errorf("failed to delete stock config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Successfully deleted stock config. Rows affected: %d", rowsAffected)

	if rowsAffected == 0 {
		return fmt.Errorf("no stock config was deleted (ID: %s)", stockID)
	}

	return nil
}

// GetSettings возвращает настройки приложения
func (d *DatabaseService) GetSettings() (models.AppSettings, error) {
	var settings models.AppSettings
	var promptsJSON string

	err := d.db.QueryRow(`
		SELECT id, temp_directory, ai_provider, ai_model, ai_api_key, ai_base_url,
		       max_concurrent_jobs, ai_timeout, ai_max_tokens, thumbnail_size, language, ai_prompts, updated_at
		FROM app_settings WHERE id = 'main'`).Scan(
		&settings.ID, &settings.TempDirectory, &settings.AIProvider,
		&settings.AIModel, &settings.AIAPIKey, &settings.AIBaseURL,
		&settings.MaxConcurrentJobs, &settings.AITimeout, &settings.AIMaxTokens, &settings.ThumbnailSize, &settings.Language,
		&promptsJSON, &settings.UpdatedAt)

	if err != nil {
		return settings, err
	}

	if promptsJSON != "" {
		json.Unmarshal([]byte(promptsJSON), &settings.AIPrompts)
	}

	return settings, nil
}

// SaveSettings сохраняет настройки приложения
func (d *DatabaseService) SaveSettings(settings models.AppSettings) error {
	promptsJSON, _ := json.Marshal(settings.AIPrompts)

	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO app_settings 
		(id, temp_directory, ai_provider, ai_model, ai_api_key, ai_base_url,
		 max_concurrent_jobs, ai_timeout, ai_max_tokens, thumbnail_size, language, ai_prompts, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"main", settings.TempDirectory, settings.AIProvider, settings.AIModel,
		settings.AIAPIKey, settings.AIBaseURL, settings.MaxConcurrentJobs,
		settings.AITimeout, settings.AIMaxTokens, settings.ThumbnailSize, settings.Language, string(promptsJSON), time.Now())

	return err
}

// UpdateAIPrompt обновляет промпт для определенного типа фото
func (d *DatabaseService) UpdateAIPrompt(photoType string, prompt string) error {
	settings, err := d.GetSettings()
	if err != nil {
		return err
	}

	if settings.AIPrompts == nil {
		settings.AIPrompts = make(map[string]string)
	}

	settings.AIPrompts[photoType] = prompt
	return d.SaveSettings(settings)
}

// createDefaultSettings создает дефолтные настройки
func (d *DatabaseService) createDefaultSettings() error {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM app_settings WHERE id = 'main'").Scan(&count)
	if err != nil {
		return err
	}

	// Если записи настроек нет, создаем новую
	if count == 0 {
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
   - Конкретные имена публичных лиц
   - Точные географические названия 
   - Названия событий и организаций
   - Новостные категории (politics, economy, sports, entertainment)
   - Временные маркеры (2024, recent, current)

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

2. ОПИСАНИЕ (строго до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов
   - ВАЖНО: Описание не должно превышать 200 символов

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

		settings := models.AppSettings{
			ID:                "main",
			TempDirectory:     "./temp",
			AIProvider:        "openai",
			AIModel:           "gpt-4o",
			MaxConcurrentJobs: 3,
			AITimeout:         90,
			AIMaxTokens:       2000,
			ThumbnailSize:     512,
			Language:          "en",
			AIPrompts:         defaultPrompts,
			UpdatedAt:         time.Now(),
		}

		return d.SaveSettings(settings)
	}

	// Если настройки существуют, проверяем наличие промптов
	settings, err := d.GetSettings()
	if err != nil {
		return err
	}

	// Если промпты пустые или отсутствуют, добавляем дефолтные
	if settings.AIPrompts == nil || len(settings.AIPrompts) == 0 {
		return d.updateDefaultPrompts()
	}

	// Проверяем есть ли оба типа промптов
	if _, hasEditorial := settings.AIPrompts["editorial"]; !hasEditorial {
		if settings.AIPrompts == nil {
			settings.AIPrompts = make(map[string]string)
		}
		settings.AIPrompts["editorial"] = `Создай метаданные для редакционного стокового фото:

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
   - Конкретные имена публичных лиц
   - Точные географические названия 
   - Названия событий и организаций
   - Новостные категории (politics, economy, sports, entertainment)
   - Временные маркеры (2024, recent, current)

ВКЛЮЧАЙ: точные названия мест, имена, организации, политические темы, контекст событий

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."]}`
	}

	if _, hasCommercial := settings.AIPrompts["commercial"]; !hasCommercial {
		if settings.AIPrompts == nil {
			settings.AIPrompts = make(map[string]string)
		}
		settings.AIPrompts["commercial"] = `Создай метаданные для коммерческого стокового фото:

ТРЕБОВАНИЯ ДЛЯ COMMERCIAL:
1. НАЗВАНИЕ (до 70 символов):
   - Описательное без конкретных имен и мест
   - Концептуальное (бизнес, семья, технологии)
   - Эмоциональное состояние (счастье, успех)
   - Универсальные формулировки

2. ОПИСАНИЕ (строго до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов
   - ВАЖНО: Описание не должно превышать 200 символов

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

ИЗБЕГАЙ: конкретные имена людей/компаний, бренды, конкретные места, даты

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."]}`
	}

	// Если что-то добавили, сохраняем обновленные настройки
	if _, hasEditorial := settings.AIPrompts["editorial"]; hasEditorial {
		if _, hasCommercial := settings.AIPrompts["commercial"]; hasCommercial {
			return d.SaveSettings(settings)
		}
	}

	return nil
}

// updateDefaultPrompts обновляет дефолтные промпты в существующих настройках
func (d *DatabaseService) updateDefaultPrompts() error {
	// Получаем текущие настройки
	settings, err := d.GetSettings()
	if err != nil {
		// Если настройки не существуют, создаем их
		return d.createDefaultSettings()
	}

	// Если промпты уже есть, не перезаписываем их
	if settings.AIPrompts != nil && len(settings.AIPrompts) > 0 {
		return nil
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

2. ОПИСАНИЕ (строго до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов
   - ВАЖНО: Описание не должно превышать 200 символов

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
	return d.SaveSettings(settings)
}

// applyMigrations применяет миграции
func (d *DatabaseService) applyMigrations() error {
	// Миграция для app_settings
	err := d.migrateAppSettings()
	if err != nil {
		return err
	}

	// Миграция для stock_configs
	err = d.migrateStockConfigs()
	if err != nil {
		return err
	}

	// Миграция для photos
	err = d.migratePhotos()
	if err != nil {
		return err
	}

	return nil
}

// migrateAppSettings применяет миграции для app_settings
func (d *DatabaseService) migrateAppSettings() error {
	// Проверяем, существуют ли поля в таблице app_settings
	rows, err := d.db.Query("PRAGMA table_info(app_settings)")
	if err != nil {
		return fmt.Errorf("failed to get table info: %w", err)
	}
	defer rows.Close()

	hasLanguageField := false
	hasAIModelField := false
	hasAIPromptsField := false
	hasAITimeoutField := false
	hasAIMaxTokensField := false
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, hasDefault int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &hasDefault)
		if err != nil {
			continue
		}

		if name == "language" {
			hasLanguageField = true
		}
		if name == "ai_model" {
			hasAIModelField = true
		}
		if name == "ai_prompts" {
			hasAIPromptsField = true
		}
		if name == "ai_timeout" {
			hasAITimeoutField = true
		}
		if name == "ai_max_tokens" {
			hasAIMaxTokensField = true
		}
	}

	// Если поле language не существует, добавляем его
	if !hasLanguageField {
		_, err = d.db.Exec("ALTER TABLE app_settings ADD COLUMN language TEXT DEFAULT 'en'")
		if err != nil {
			return fmt.Errorf("failed to add language column: %w", err)
		}
		log.Println("Added language column to app_settings table")
	}

	// Если поле ai_model не существует, добавляем его
	if !hasAIModelField {
		_, err = d.db.Exec("ALTER TABLE app_settings ADD COLUMN ai_model TEXT DEFAULT 'gpt-4o'")
		if err != nil {
			return fmt.Errorf("failed to add ai_model column: %w", err)
		}
		log.Println("Added ai_model column to app_settings table")
	}

	// Если поле ai_prompts не существует, добавляем его
	if !hasAIPromptsField {
		_, err = d.db.Exec("ALTER TABLE app_settings ADD COLUMN ai_prompts TEXT")
		if err != nil {
			return fmt.Errorf("failed to add ai_prompts column: %w", err)
		}
		log.Println("Added ai_prompts column to app_settings table")

		// После добавления поля обновляем настройки дефолтными промптами
		err = d.updateDefaultPrompts()
		if err != nil {
			log.Printf("Warning: failed to update default prompts: %v", err)
		}
	}

	// Если поле ai_timeout не существует, добавляем его
	if !hasAITimeoutField {
		_, err = d.db.Exec("ALTER TABLE app_settings ADD COLUMN ai_timeout INTEGER DEFAULT 90")
		if err != nil {
			return fmt.Errorf("failed to add ai_timeout column: %w", err)
		}
		log.Println("Added ai_timeout column to app_settings table")
	}

	// Если поле ai_max_tokens не существует, добавляем его
	if !hasAIMaxTokensField {
		_, err = d.db.Exec("ALTER TABLE app_settings ADD COLUMN ai_max_tokens INTEGER DEFAULT 2000")
		if err != nil {
			return fmt.Errorf("failed to add ai_max_tokens column: %w", err)
		}
		log.Println("Added ai_max_tokens column to app_settings table")
	}

	return nil
}

// migrateStockConfigs применяет миграции для stock_configs
func (d *DatabaseService) migrateStockConfigs() error {
	// Проверяем, существуют ли новые поля в таблице stock_configs
	rows, err := d.db.Query("PRAGMA table_info(stock_configs)")
	if err != nil {
		return fmt.Errorf("failed to get stock_configs table info: %w", err)
	}
	defer rows.Close()

	hasTypeField := false
	hasSettingsField := false
	hasModulePathField := false

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, hasDefault int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &hasDefault)
		if err != nil {
			continue
		}

		switch name {
		case "type":
			hasTypeField = true
		case "settings":
			hasSettingsField = true
		case "module_path":
			hasModulePathField = true
		}
	}

	// Добавляем отсутствующие поля
	if !hasTypeField {
		_, err = d.db.Exec("ALTER TABLE stock_configs ADD COLUMN type TEXT DEFAULT 'ftp'")
		if err != nil {
			return fmt.Errorf("failed to add type column: %w", err)
		}
		log.Println("Added type column to stock_configs table")

		// Обновляем существующие записи: копируем upload_method в type
		_, err = d.db.Exec("UPDATE stock_configs SET type = upload_method WHERE type IS NULL OR type = ''")
		if err != nil {
			log.Printf("Warning: failed to migrate upload_method to type: %v", err)
		}
	}

	if !hasSettingsField {
		_, err = d.db.Exec("ALTER TABLE stock_configs ADD COLUMN settings TEXT")
		if err != nil {
			return fmt.Errorf("failed to add settings column: %w", err)
		}
		log.Println("Added settings column to stock_configs table")
	}

	if !hasModulePathField {
		_, err = d.db.Exec("ALTER TABLE stock_configs ADD COLUMN module_path TEXT")
		if err != nil {
			return fmt.Errorf("failed to add module_path column: %w", err)
		}
		log.Println("Added module_path column to stock_configs table")
	}

	return nil
}

// migratePhotos применяет миграции для photos
func (d *DatabaseService) migratePhotos() error {
	// Проверяем, существуют ли поля в таблице photos
	rows, err := d.db.Query("PRAGMA table_info(photos)")
	if err != nil {
		return fmt.Errorf("failed to get photos table info: %w", err)
	}
	defer rows.Close()

	hasContentTypeField := false
	hasUpdatedAtField := false

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, hasDefault int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &hasDefault)
		if err != nil {
			continue
		}

		switch name {
		case "content_type":
			hasContentTypeField = true
		case "updated_at":
			hasUpdatedAtField = true
		}
	}

	// Если поле content_type не существует, добавляем его
	if !hasContentTypeField {
		_, err = d.db.Exec("ALTER TABLE photos ADD COLUMN content_type TEXT")
		if err != nil {
			return fmt.Errorf("failed to add content_type column: %w", err)
		}
		log.Println("Added content_type column to photos table")

		// Обновляем существующие записи: устанавливаем content_type на основе batch type
		_, err = d.db.Exec(`
			UPDATE photos 
			SET content_type = (
				SELECT type FROM batches WHERE batches.id = photos.batch_id
			) 
			WHERE content_type IS NULL OR content_type = ''`)
		if err != nil {
			log.Printf("Warning: failed to update content_type for existing photos: %v", err)
		} else {
			log.Println("Updated content_type for existing photos based on batch type")
		}
	}

	// Если поле updated_at не существует, добавляем его
	if !hasUpdatedAtField {
		// Сначала добавляем колонку без дефолтного значения
		_, err = d.db.Exec("ALTER TABLE photos ADD COLUMN updated_at DATETIME")
		if err != nil {
			return fmt.Errorf("failed to add updated_at column: %w", err)
		}
		log.Println("Added updated_at column to photos table")

		// Обновляем существующие записи: устанавливаем updated_at = created_at
		_, err = d.db.Exec("UPDATE photos SET updated_at = created_at WHERE updated_at IS NULL")
		if err != nil {
			log.Printf("Warning: failed to update updated_at for existing photos: %v", err)
		} else {
			log.Println("Updated updated_at for existing photos")
		}
	}

	return nil
}

// createDemoStockConfig создает демо конфигурацию стока
func (d *DatabaseService) createDemoStockConfig() error {
	// Проверяем, есть ли уже демо конфигурация
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM stock_configs WHERE id = 'shutterstock-demo'").Scan(&count)
	if err != nil {
		return err
	}

	// Если демо конфигурация уже существует, не создаем новую
	if count > 0 {
		return nil
	}

	// Создаем демо конфигурацию Shutterstock
	config := models.StockConfig{
		ID:             "shutterstock-demo",
		Name:           "Shutterstock (Demo)",
		Type:           "api",
		SupportedTypes: []string{"commercial", "editorial"},
		Connection: models.ConnectionConfig{
			APIUrl:  "https://api.shutterstock.com/v2/images",
			APIKey:  "demo-api-key",
			Timeout: 60,
		},
		Prompts: map[string]string{
			"commercial": "Standard commercial prompt",
			"editorial":  "Standard editorial prompt",
		},
		Settings: map[string]interface{}{
			"demo_mode":    true,
			"max_keywords": 25,
		},
		ModulePath: "",
		Active:     true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return d.SaveStockConfig(config)
}

// LogEvent записывает событие в лог
func (d *DatabaseService) LogEvent(batchID, photoID, eventType, status, message, details string, progress int) error {
	eventID := fmt.Sprintf("event_%d", time.Now().UnixNano())

	_, err := d.db.Exec(`
		INSERT INTO event_logs (id, batch_id, photo_id, event_type, status, message, details, progress, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		eventID, batchID, photoID, eventType, status, message, details, progress)

	if err != nil {
		log.Printf("Failed to log event: %v", err)
		return err
	}

	return nil
}

// GetBatchEvents возвращает события для батча
func (d *DatabaseService) GetBatchEvents(batchID string, limit int) ([]models.EventLog, error) {
	query := `
		SELECT id, batch_id, photo_id, event_type, status, message, details, progress, created_at
		FROM event_logs 
		WHERE batch_id = ?
		ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query batch events: %w", err)
	}
	defer rows.Close()

	var events []models.EventLog
	for rows.Next() {
		var event models.EventLog
		var photoID sql.NullString
		var details sql.NullString

		err := rows.Scan(&event.ID, &event.BatchID, &photoID, &event.EventType,
			&event.Status, &event.Message, &details, &event.Progress, &event.CreatedAt)
		if err != nil {
			continue
		}

		if photoID.Valid {
			event.PhotoID = photoID.String
		}
		if details.Valid {
			event.Details = details.String
		}

		events = append(events, event)
	}

	return events, nil
}

// GetPhotoEvents возвращает события для фотографии
func (d *DatabaseService) GetPhotoEvents(photoID string) ([]models.EventLog, error) {
	rows, err := d.db.Query(`
		SELECT id, batch_id, photo_id, event_type, status, message, details, progress, created_at
		FROM event_logs 
		WHERE photo_id = ?
		ORDER BY created_at DESC`, photoID)
	if err != nil {
		return nil, fmt.Errorf("failed to query photo events: %w", err)
	}
	defer rows.Close()

	var events []models.EventLog
	for rows.Next() {
		var event models.EventLog
		var photoID sql.NullString
		var details sql.NullString

		err := rows.Scan(&event.ID, &event.BatchID, &photoID, &event.EventType,
			&event.Status, &event.Message, &details, &event.Progress, &event.CreatedAt)
		if err != nil {
			continue
		}

		if photoID.Valid {
			event.PhotoID = photoID.String
		}
		if details.Valid {
			event.Details = details.String
		}

		events = append(events, event)
	}

	return events, nil
}

// CleanupOldEvents удаляет старые события (старше указанного количества дней)
func (d *DatabaseService) CleanupOldEvents(olderThanDays int) error {
	_, err := d.db.Exec(`
		DELETE FROM event_logs 
		WHERE created_at < datetime('now', '-' || ? || ' days')`,
		olderThanDays)

	if err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	return nil
}
