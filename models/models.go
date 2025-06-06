package models

import "time"

// PhotoBatch представляет группу фотографий для обработки
type PhotoBatch struct {
	ID          string      `json:"id" db:"id"`
	Type        string      `json:"type" db:"type"` // "editorial" or "commercial"
	Description string      `json:"description" db:"description"`
	FolderPath  string      `json:"folderPath" db:"folder_path"`
	Photos      []Photo     `json:"photos"`
	PhotosStats *BatchStats `json:"photosStats,omitempty"` // статистика фото в батче
	Status      string      `json:"status" db:"status"`    // "pending", "processing", "completed", "failed"
	CreatedAt   time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time   `json:"updatedAt" db:"updated_at"`
}

// BatchStats содержит статистику по фото в батче
type BatchStats struct {
	Total     int `json:"total"`     // общее количество фото
	Processed int `json:"processed"` // обработано нейросетью
	Approved  int `json:"approved"`  // одобрено к загрузке
	Rejected  int `json:"rejected"`  // отклонено
}

// Photo представляет отдельную фотографию
type Photo struct {
	ID            string            `json:"id" db:"id"`
	BatchID       string            `json:"batchId" db:"batch_id"`
	ContentType   string            `json:"contentType" db:"content_type"` // "editorial" or "commercial"
	OriginalPath  string            `json:"originalPath" db:"original_path"`
	ThumbnailPath string            `json:"thumbnailPath" db:"thumbnail_path"`
	FileName      string            `json:"fileName" db:"file_name"`
	FileSize      int64             `json:"fileSize" db:"file_size"`
	ExifData      map[string]string `json:"exifData"`
	AIResult      *AIResult         `json:"aiResult,omitempty"` // результаты AI анализа
	UploadStatus  map[string]string `json:"uploadStatus"`       // stock_id -> status
	Status        string            `json:"status" db:"status"` // "pending", "processing", "completed", "failed"
	CreatedAt     time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time         `json:"updatedAt,omitempty"` // время последнего обновления
}

// AIResult содержит результаты анализа нейросетью
type AIResult struct {
	ContentType string   `json:"contentType"` // "editorial" or "commercial" - определенный AI тип контента
	Title       string   `json:"title"`
	Keywords    []string `json:"keywords"`
	Quality     int      `json:"quality"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Processed   bool     `json:"processed"`
	Error       string   `json:"error,omitempty"`
}

// StockConfig представляет конфигурацию стока
type StockConfig struct {
	ID             string                 `json:"id" db:"id"`
	Name           string                 `json:"name" db:"name"`
	Type           string                 `json:"type" db:"type"`                  // "ftp", "sftp", "api", "custom"
	SupportedTypes []string               `json:"supportedTypes"`                  // ["editorial", "commercial"]
	UploadMethod   string                 `json:"uploadMethod" db:"upload_method"` // deprecated, use Type instead
	Connection     ConnectionConfig       `json:"connection"`
	Prompts        map[string]string      `json:"prompts"`  // "editorial" -> prompt, "commercial" -> prompt
	Settings       map[string]interface{} `json:"settings"` // дополнительные настройки для конкретного типа стока
	Active         bool                   `json:"active" db:"active"`
	ModulePath     string                 `json:"modulePath" db:"module_path"` // путь к файлу модуля загрузчика
	CreatedAt      time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time              `json:"updatedAt" db:"updated_at"`
}

// ConnectionConfig содержит настройки подключения к стоку
type ConnectionConfig struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Path       string `json:"path"`
	APIKey     string `json:"apiKey,omitempty"`
	APIUrl     string `json:"apiUrl,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`    // таймаут соединения в секундах
	UseSSL     bool   `json:"useSSL,omitempty"`     // использовать SSL/TLS
	Passive    bool   `json:"passive,omitempty"`    // пассивный режим для FTP
	Encryption string `json:"encryption,omitempty"` // тип шифрования: "none", "auto", "explicit", "implicit"
	VerifyCert bool   `json:"verifyCert,omitempty"` // проверять SSL сертификаты
	// Дополнительные параметры для API подключений
	Headers map[string]string `json:"headers,omitempty"`
	Params  map[string]string `json:"params,omitempty"`
}

// AppSettings содержит глобальные настройки приложения
type AppSettings struct {
	ID                string            `json:"id" db:"id"`
	TempDirectory     string            `json:"tempDirectory" db:"temp_directory"`
	AIProvider        string            `json:"aiProvider" db:"ai_provider"` // "openai", "claude"
	AIModel           string            `json:"aiModel" db:"ai_model"`       // "gpt-4-vision-preview", "claude-3-opus", etc.
	AIAPIKey          string            `json:"aiApiKey" db:"ai_api_key"`
	AIBaseURL         string            `json:"aiBaseUrl" db:"ai_base_url"`
	MaxConcurrentJobs int               `json:"maxConcurrentJobs" db:"max_concurrent_jobs"`
	AITimeout         int               `json:"aiTimeout" db:"ai_timeout"`      // таймаут AI запросов в секундах
	AIMaxTokens       int               `json:"aiMaxTokens" db:"ai_max_tokens"` // максимальное количество токенов в ответе
	ThumbnailSize     int               `json:"thumbnailSize" db:"thumbnail_size"`
	Language          string            `json:"language" db:"language"` // "en", "ru", etc.
	AIPrompts         map[string]string `json:"aiPrompts"`              // "editorial" -> prompt, "commercial" -> prompt
	UpdatedAt         time.Time         `json:"updatedAt" db:"updated_at"`
}

// BatchStatus представляет статус обработки батча
type BatchStatus struct {
	BatchID         string             `json:"batchId"`
	Type            string             `json:"type"`
	Description     string             `json:"description"`
	TotalPhotos     int                `json:"totalPhotos"`
	ProcessedPhotos int                `json:"processedPhotos"`
	Status          string             `json:"status"`
	Progress        int                `json:"progress"` // 0-100
	CurrentPhoto    string             `json:"currentPhoto,omitempty"`
	CurrentStep     string             `json:"currentStep,omitempty"` // текущий этап обработки
	Photos          []PhotoProcessInfo `json:"photos,omitempty"`      // детальная информация по фотографиям
	Error           string             `json:"error,omitempty"`
}

// PhotoProcessInfo содержит информацию о процессе обработки фотографии
type PhotoProcessInfo struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Status   string `json:"status"`   // "pending", "processing", "completed", "failed"
	Progress int    `json:"progress"` // 0-100
	Step     string `json:"step"`     // "preparation", "ai_analysis", "saving", "exif_writing", "completed"
	Error    string `json:"error,omitempty"`
}

// AIRequest представляет запрос к AI API
type AIRequest struct {
	Image       string            `json:"image"` // base64 encoded
	ExifData    map[string]string `json:"exifData"`
	Description string            `json:"description"`
	Prompt      string            `json:"prompt"`
	PhotoType   string            `json:"photoType"`
}

// AIResponse представляет ответ от AI API
type AIResponse struct {
	Title       string   `json:"title"`
	Keywords    []string `json:"keywords"`
	Quality     int      `json:"quality"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Error       string   `json:"error,omitempty"`
}

// UploadJob представляет задачу загрузки
type UploadJob struct {
	PhotoID     string      `json:"photoId"`
	BatchID     string      `json:"batchId"`
	StockConfig StockConfig `json:"stockConfig"`
	Priority    int         `json:"priority"`
}

// UploadResult представляет результат загрузки
type UploadResult struct {
	PhotoID   string `json:"photoId"`
	StockID   string `json:"stockId"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	UploadURL string `json:"uploadUrl,omitempty"`
}

// QueueInfo представляет информацию о очереди
type QueueInfo struct {
	TotalJobs      int `json:"totalJobs"`
	ProcessingJobs int `json:"processingJobs"`
	CompletedJobs  int `json:"completedJobs"`
	FailedJobs     int `json:"failedJobs"`
}

// PhotoFile представляет файл фотографии при выборе папки
type PhotoFile struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	IsValid   bool   `json:"isValid"`
}

// AIModel представляет информацию о модели ИИ
type AIModel struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	MaxTokens      int    `json:"maxTokens"`
	SupportsVision bool   `json:"supportsVision"`
	Provider       string `json:"provider"`
}

// AIModelsResponse представляет ответ со списком моделей
type AIModelsResponse struct {
	Models []AIModel `json:"models"`
	Error  string    `json:"error,omitempty"`
}

// StockUploader интерфейс для модульных загрузчиков
type StockUploader interface {
	// Загрузка файла
	Upload(photo Photo, config StockConfig) (UploadResult, error)
	// Тестирование соединения
	TestConnection(config StockConfig) error
	// Получение информации о загрузчике
	GetInfo() UploaderInfo
	// Валидация конфигурации
	ValidateConfig(config StockConfig) error
}

// UploaderInfo содержит информацию о загрузчике
type UploaderInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Type        string `json:"type"` // "ftp", "sftp", "api", "custom"
	Website     string `json:"website,omitempty"`
}

// StockTemplate содержит шаблон конфигурации для нового стока
type StockTemplate struct {
	Type        string                 `json:"type"` // "ftp", "sftp", "api"
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Fields      []TemplateField        `json:"fields"`   // поля для настройки
	Defaults    map[string]interface{} `json:"defaults"` // значения по умолчанию
	Examples    map[string]string      `json:"examples"` // примеры заполнения
}

// TemplateField описывает поле в шаблоне
type TemplateField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "text", "number", "password", "select", "checkbox", "url"
	Label       string      `json:"label"`
	Required    bool        `json:"required"`
	Options     []string    `json:"options,omitempty"` // для select
	Placeholder string      `json:"placeholder,omitempty"`
	Help        string      `json:"help,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Validation  string      `json:"validation,omitempty"` // regex для валидации
}

// EventLog представляет запись в логе событий
type EventLog struct {
	ID        string    `json:"id" db:"id"`
	BatchID   string    `json:"batchId" db:"batch_id"`
	PhotoID   string    `json:"photoId,omitempty" db:"photo_id"`  // опционально для событий по конкретному фото
	EventType string    `json:"eventType" db:"event_type"`        // "ai_processing", "ftp_upload", "batch_start", "batch_complete", "error"
	Status    string    `json:"status" db:"status"`               // "started", "success", "failed", "progress"
	Message   string    `json:"message" db:"message"`             // описание события
	Details   string    `json:"details,omitempty" db:"details"`   // дополнительные детали (JSON или текст)
	Progress  int       `json:"progress,omitempty" db:"progress"` // прогресс 0-100
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// ProcessingProgress представляет прогресс обработки
type ProcessingProgress struct {
	BatchID         string     `json:"batchId"`
	TotalPhotos     int        `json:"totalPhotos"`
	CurrentStep     string     `json:"currentStep"`
	CurrentPhoto    string     `json:"currentPhoto"`
	PhotoProgress   int        `json:"photoProgress"`   // прогресс текущего фото 0-100
	OverallProgress int        `json:"overallProgress"` // общий прогресс 0-100
	Status          string     `json:"status"`          // "processing", "completed", "failed"
	RecentEvents    []EventLog `json:"recentEvents"`    // последние события
}

// UploadProgress представляет прогресс загрузки
type UploadProgress struct {
	BatchID        string     `json:"batchId"`
	TotalPhotos    int        `json:"totalPhotos"`
	UploadedPhotos int        `json:"uploadedPhotos"`
	FailedPhotos   int        `json:"failedPhotos"`
	CurrentPhoto   string     `json:"currentPhoto"`
	Status         string     `json:"status"`
	RecentEvents   []EventLog `json:"recentEvents"`
}
