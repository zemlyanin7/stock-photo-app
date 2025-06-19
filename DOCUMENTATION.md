# Stock Photo Automation App - Полная Документация

## 📖 Содержание

1. [Обзор проекта](#обзор-проекта)
2. [Установка и настройка](#установка-и-настройка) 
3. [Архитектура системы](#архитектура-системы)
4. [Рабочий процесс (Workflow)](#рабочий-процесс-workflow)
5. [AI промпты и анализ](#ai-промпты-и-анализ)
6. [EXIF обработка и метаданные](#exif-обработка-и-метаданные)
7. [Система загрузки на стоки](#система-загрузки-на-стоки)
8. [UI компоненты](#ui-компоненты)
9. [База данных](#база-данных)
10. [API справочник](#api-справочник)
11. [Безопасность](#безопасность)
12. [Разработка](#разработка)

---

## Обзор проекта

### Описание

Stock Photo Automation App — это desktop приложение для автоматизации работы фотографа со стоковыми площадками. Приложение использует AI для создания метаданных (названий, описаний, ключевых слов) и автоматически загружает фотографии на стоковые платформы.

### Основные возможности

- **🤖 AI анализ фотографий** - автоматическое создание названий, описаний и ключевых слов
- **📂 Batch обработка** - массовая обработка папок с фотографиями  
- **📝 Два типа контента** - Editorial (новостные) и Commercial (коммерческие) фотографии
- **📊 Система очередей** - параллельная обработка и загрузка с контролем ресурсов
- **🔧 EXIF обработка** - запись метаданных в файлы изображений
- **📤 Множественные загрузчики** - поддержка FTP, SFTP и API загрузок
- **🎯 Управление выбором** - точный контроль какие файлы загружать

### Технический стек

- **Backend**: Go 1.19+ с Wails v2.10.1
- **Frontend**: Vanilla JavaScript + Vite + Tailwind CSS
- **База данных**: SQLite 
- **AI провайдеры**: OpenAI GPT-4, (Claude - в разработке)
- **Протоколы загрузки**: FTP/FTPS, SFTP, REST API

---

## Установка и настройка

### Системные требования

- **Go**: 1.19 или выше
- **Node.js**: 16 или выше  
- **ExifTool**: для записи метаданных в EXIF
- **Wails**: для сборки desktop приложения

### Установка ExifTool

ExifTool необходим для записи AI-сгенерированных метаданных в EXIF данные изображений.

**macOS:**
```bash
brew install exiftool
```

**Ubuntu/Debian:**
```bash
sudo apt-get install libimage-exiftool-perl
```

**Windows:**
1. Скачайте с [exiftool.org](https://exiftool.org/)
2. Распакуйте и добавьте в PATH
3. Переименуйте `exiftool(-k).exe` в `exiftool.exe`

**Проверка:**
```bash
exiftool -ver
```

### Установка для разработки

```bash
# 1. Клонирование репозитория
git clone <repository-url>
cd stock-photo-app

# 2. Установка зависимостей Go
go mod download

# 3. Установка зависимостей фронтенда
cd frontend && npm install && cd ..

# 4. Запуск в режиме разработки
wails dev
```

### Первая настройка

При первом запуске:

1. **База данных**: Автоматически создается `app.db`
2. **API ключи**: Настройте в Settings → AI
3. **Стоки**: Добавьте настройки стоковых площадок в Settings → Stocks

---

## Архитектура системы

### Основные компоненты

```
app.go                    # Главный контроллер приложения
├── services/
│   ├── queue_manager.go       # Управление очередью обработки AI
│   ├── upload_queue_manager.go # Управление очередью загрузки на стоки
│   ├── ai_service.go          # Интеграция с AI API
│   ├── image_processor.go     # Обработка изображений и EXIF
│   ├── exif_processor.go      # Специализированная EXIF обработка
│   ├── database_service.go    # Работа с базой данных
│   └── logger.go              # Система логирования
├── uploaders/
│   ├── manager.go             # Менеджер загрузчиков
│   ├── ftp_uploader.go        # FTP/FTPS загрузка
│   ├── sftp_uploader.go       # SFTP загрузка
│   └── api_uploader.go        # API загрузка
├── models/
│   └── models.go              # Модели данных
└── frontend/
    ├── src/main.js            # Основная логика UI
    ├── src/upload-manager.js  # Управление загрузкой
    └── src/i18n.js           # Интернационализация
```

### Потоки данных

```
1. Пользователь → Drag&Drop папки → app.ProcessPhotoFolder()
2. QueueManager → Обработка изображений + AI анализ  
3. DatabaseService → Сохранение результатов + EXIF запись
4. UI Review → Редактирование + одобрение фотографий
5. UploadQueueManager → Загрузка на стоковые площадки
```

---

## Рабочий процесс (Workflow)

### 1. Добавление в очередь обработки

**Метод**: `app.ProcessPhotoFolder(folderPath, description, photoType)`

**Этапы**:
1. Создание `PhotoBatch` с типом (editorial/commercial)
2. Сканирование папки для поиска изображений
3. Установка `ContentType` для всех фотографий  
4. Добавление батча в очередь обработки
5. Запуск `QueueManager` если не активен

### 2. Обработка очереди (QueueManager)

**Параллельная обработка** с worker pool:

```go
// Настройка воркеров (по умолчанию 3)
numWorkers := settings.MaxConcurrentJobs

// Обработка каждого фото
for each photo:
    1. Создание миниатюры (512px) для экономии AI токенов
    2. Извлечение EXIF данных
    3. AI анализ с контекстным промптом  
    4. Сохранение результатов в базу данных
    5. Запись метаданных в EXIF оригинального файла
```

### 3. AI анализ фотографии

**Этапы AI обработки**:

1. **Подготовка изображения**:
   - Создание оптимизированной миниатюры
   - Извлечение EXIF метаданных
   - Определение типа контента

2. **Выбор промпта**:
   - Editorial: промпт для новостных фото  
   - Commercial: промпт для коммерческих фото

3. **Построение контекстного промпта**:
   - Базовый промпт + EXIF данные + описание пользователя
   - Адаптация под тип контента

4. **Отправка в AI API**:
   - Кодирование изображения в base64
   - Отправка запроса с промптом
   - Получение структурированного JSON ответа

5. **Обработка результата**:
   - Парсинг JSON ответа
   - Валидация длины полей (200 символов для commercial description)
   - Сохранение в базу данных

### 4. Просмотр и редактирование (Review)

**UI компоненты**:
- Список обработанных батчей
- Превью фотографий с миниатюрами
- Редактируемые поля метаданных
- Система одобрения/отклонения
- Регенерация метаданных с кастомными промптами

### 5. Управление загрузкой

**Новая система выбора фотографий**:
- ✅ Чекбоксы для каждой фотографии
- ✅ Массовые операции: "Select All", "Clear All"
- ✅ Кнопка "Upload Selected" для загрузки выбранных
- ✅ Ограничение параллельности (максимум 2 загрузки одновременно)

### 6. Очередь загрузки (UploadQueueManager)

**Функциональность**:
- Worker pool с ограничением (2 воркера)
- Отслеживание статуса каждого файла
- Загрузка на несколько стоков параллельно  
- Детальное логирование всех операций
- UI отображение прогресса в реальном времени

---

## AI промпты и анализ

### Определение типа контента

Система автоматически определяет тип контента, но также позволяет пользователю выбрать тип при добавлении папки.

### Промпт для Commercial фотографий

```markdown
Создай метаданные для коммерческого стокового фото:

СОЗДАЙ:
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

3. КЛЮЧЕВЫЕ СЛОВА (15-20 слов):
   - Демографические: woman, man, family, adult
   - Эмоциональные: happy, successful, confident
   - Концептуальные: business, lifestyle, technology
   - Визуальные: modern, bright, professional

ИЗБЕГАЙ:
- Конкретные имена людей или компаний
- Названия брендов и торговых марок  
- Конкретные географические названия
- Даты и временные привязки

ФОРМАТ ОТВЕТА (JSON):
{
  "title": "название",
  "description": "описание", 
  "keywords": ["слово1", "слово2", ...],
  "category": "Business"
}
```

### Промпт для Editorial фотографий

```markdown
Создай метаданные для редакционного фото:

СОЗДАЙ:
1. НАЗВАНИЕ (до 100 символов):
   - Фактическое описание события
   - Конкретные имена людей и мест
   - Журналистский стиль изложения

2. ОПИСАНИЕ (до 500 символов):
   - WHO: конкретные имена людей
   - WHAT: точное описание происходящего
   - WHERE: конкретные места с полными названиями
   - WHEN: даты и время (если применимо)
   - WHY: контекст и причины события

3. КЛЮЧЕВЫЕ СЛОВА (15-25 слов):
   - Конкретные имена публичных лиц
   - Географические названия (города, страны)
   - Названия событий и организаций
   - Новостные категории

ВКЛЮЧАЙ:
- Точные названия мест и организаций
- Имена публичных лиц и знаменитостей
- Политические и общественные темы
- Контекст новостных событий

ФОРМАТ ОТВЕТА (JSON):
{
  "title": "название",
  "description": "описание",
  "keywords": ["слово1", "слово2", ...], 
  "category": "News"
}
```

### Контекстные промпты с EXIF

Система добавляет к базовому промпту информацию из EXIF:

**Для Editorial**:
```
EXIF КОНТЕКСТ:
Дата съемки: 2024-01-15 14:30:22
Камера: Canon EOS R5  
Локация: 52.5200° N, 13.4050° E (Berlin, Germany)
```

**Для Commercial**:
```
ТЕХНИЧЕСКАЯ ИНФОРМАЦИЯ:
Камера: Canon EOS R5
Объектив: 24-70mm f/2.8
ISO: 800, f/4.0, 1/125s
```

### Валидация результатов

Система автоматически проверяет и корректирует AI результаты:

```go
// Валидация длины описания
if contentType == "commercial" && len(result.Description) > 200 {
    log.Printf("Commercial description exceeds 200 chars, truncating")
    result.Description = result.Description[:197] + "..."
}

if contentType == "editorial" && len(result.Description) > 500 {
    log.Printf("Editorial description exceeds 500 chars, truncating") 
    result.Description = result.Description[:497] + "..."
}
```

---

## EXIF обработка и метаданные

### Когда записываются EXIF данные

EXIF метаданные записываются в оригинальные файлы при нажатии кнопки **"Approve"** в процессе просмотра результатов.

### Поля метаданных

**Стандартные поля EXIF/IPTC/XMP**:

- **Title** → `EXIF:Title`, `XMP:Title`, `IPTC:ObjectName`
- **Description** → `EXIF:Description`, `XMP:Description`, `IPTC:Caption-Abstract`
- **Keywords** → `IPTC:Keywords`, `XMP:Subject` (как отдельные элементы)
- **Category** → `XMP:Category`, `IPTC:Category`  
- **Quality Rating** → `EXIF:Rating`, `XMP:Rating` (1-100)
- **Content Type** → `XMP:ContentType` (editorial/commercial)

### Категории стоков

**Editorial категории**:
News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment

**Commercial категории**:  
Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture

### Техническая реализация

**Использует ExifTool**:
```bash
# Двухэтапная запись для надежности
exiftool -overwrite_original -all:all= photo.jpg  # Очистка
exiftool -overwrite_original -Title="AI Title" -Description="AI Description" photo.jpg
```

**Особенности**:
- ✅ Правильная запись множественных ключевых слов как отдельных элементов  
- ✅ Поддержка всех форматов (JPEG, TIFF, PNG)
- ✅ Graceful fallback если ExifTool недоступен
- ✅ Полная перезапись старых AI-метаданных при повторном approve
- ✅ Сохранение технических EXIF данных камеры

---

## Система загрузки на стоки

### Архитектура загрузчиков

**Модульная система** с поддержкой различных протоколов:

```
UploaderManager
├── FTPUploader    # FTP/FTPS загрузка файлов
├── SFTPUploader   # SFTP загрузка файлов  
├── APIUploader    # REST API загрузка + метаданные
└── CustomUploader # Расширяемость для новых протоколов
```

### Интерфейс StockUploader

Каждый загрузчик реализует единый интерфейс:

```go
type StockUploader interface {
    Upload(photo Photo, config StockConfig) (UploadResult, error)
    TestConnection(config StockConfig) error
    GetInfo() UploaderInfo
    ValidateConfig(config StockConfig) error
}
```

### Типы подключений

**FTP/FTPS**:
- Порт по умолчанию: 21
- Поддержка пассивного режима
- SSL/TLS шифрование

**SFTP**:
- Порт по умолчанию: 22  
- SSH ключи и пароли
- Автоматическое создание папок

**API**:
- Multipart upload файлов
- JSON метаданные
- Настраиваемые headers и параметры

### Конфигурация стоков

```go
type StockConfig struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`  
    Type        string                 `json:"type"`        // "ftp", "sftp", "api"
    IsActive    bool                   `json:"isActive"`
    Connection  ConnectionConfig       `json:"connection"`
    Settings    map[string]interface{} `json:"settings"`
}
```

### Очередь загрузки (UploadQueueManager)

**Новая функциональность управления загрузкой**:

- **🔢 Ограничение параллельности**: максимум 2 файла одновременно
- **📊 Детальное отслеживание**: статус каждого файла на каждом стоке  
- **✅ Система выбора**: чекбоксы для отметки файлов для загрузки
- **🎛️ Массовые операции**: "Select All", "Clear All", "Upload Selected"
- **📈 UI отслеживание**: прогресс очереди в реальном времени

**Worker Pool архитектура**:
```go
// 2 воркера обрабатывают очередь параллельно
for i := 0; i < maxConcurrentUploads; i++ {
    go uploadWorker(i)
}
```

**Состояния файлов**:
- `pending` → `queued` → `uploading` → `uploaded`/`failed`/`partially_uploaded`

---

## UI компоненты

### Главные вкладки

**Editorial/Commercial вкладки**:
- Drag & Drop зоны для папок
- Поле описания для батча
- Кнопки запуска обработки

**Queue (Очередь)**:
- Список активных батчей  
- Прогресс обработки в реальном времени
- Детали по каждой фотографии

**Review (Просмотр)**:  
- Выбор обработанного батча
- Превью фотографий с метаданными
- Редактирование AI результатов
- Система одобрения/отклонения
- Регенерация с кастомными промптами

**Settings (Настройки)**:
- AI провайдер и API ключи
- Настройки стоковых площадок  
- Кастомизация промптов
- Языковые настройки

### Upload Manager

**Новые элементы управления загрузкой**:

```html
<!-- Чекбоксы для каждой фотографии -->
<input type="checkbox" class="photo-select-checkbox" data-photo-id="...">

<!-- Панель массового управления -->
<button id="selectAllForUploadBtn">Select All</button>
<button id="clearSelectionBtn">Clear All</button>  
<button id="uploadSelectedBtn">Upload Selected (0)</button>

<!-- Статус очереди загрузки -->
<div id="uploadQueueStatus">
  Active: <span id="activeUploadsCount">0</span> / 
  Queue: <span id="queueLengthCount">0</span> /
  Max: <span id="maxConcurrentCount">2</span>
</div>
```

**JavaScript функциональность**:
```javascript
// Переключение выбора фотографии
uploadManager.togglePhotoSelection(photoId, selected)

// Массовые операции  
uploadManager.selectAllPhotos()
uploadManager.clearAllSelection()
uploadManager.uploadSelectedPhotos()

// Отслеживание прогресса
uploadManager.startUploadQueueTracking()
```

### Интернационализация

Поддержка языков через `src/i18n.js`:
- 🇷🇺 Русский (по умолчанию)
- 🇺🇸 English
- Легко расширяемо для новых языков

---

## База данных

### Схема таблиц

**batches** - информация о батчах:
```sql
CREATE TABLE batches (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,           -- 'editorial' или 'commercial'  
    description TEXT,             -- описание от пользователя
    folder_path TEXT,
    status TEXT,                  -- 'pending', 'processing', 'processed'
    created_at DATETIME,
    updated_at DATETIME
);
```

**photos** - информация о фотографиях:
```sql  
CREATE TABLE photos (
    id TEXT PRIMARY KEY,
    batch_id TEXT,
    file_name TEXT,
    original_path TEXT,
    thumbnail_path TEXT,
    content_type TEXT,            -- 'editorial' или 'commercial'
    status TEXT,                  -- 'pending', 'processing', 'processed', 'approved', 'rejected'
    ai_results TEXT,              -- JSON с результатами AI
    upload_status TEXT,           -- статус загрузки
    exif_json TEXT,              -- EXIF данные как JSON
    created_at DATETIME,
    updated_at DATETIME
);
```

**app_settings** - настройки приложения:
```sql
CREATE TABLE app_settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    created_at DATETIME,
    updated_at DATETIME  
);
```

**stock_configs** - конфигурации стоков:
```sql
CREATE TABLE stock_configs (
    id TEXT PRIMARY KEY,
    name TEXT,
    type TEXT,                    -- 'ftp', 'sftp', 'api'
    is_active BOOLEAN,
    connection_config TEXT,       -- JSON с настройками подключения
    settings TEXT,                -- дополнительные настройки
    created_at DATETIME,
    updated_at DATETIME
);
```

**event_logs** - логи событий:
```sql
CREATE TABLE event_logs (
    id INTEGER PRIMARY KEY,
    batch_id TEXT,
    photo_id TEXT,
    event_type TEXT,              -- 'batch_start', 'ai_processing', 'stock_upload'
    status TEXT,                  -- 'started', 'progress', 'success', 'failed'
    message TEXT,
    details TEXT,
    progress INTEGER,             -- 0-100
    created_at DATETIME
);
```

### Миграции

Система автоматических миграций при запуске приложения:
- Создание таблиц если не существуют
- Добавление новых столбцов
- Обновление схемы под новые версии

---

## API справочник

### Основные методы обработки

```go
// Добавление папки в очередь обработки
ProcessPhotoFolder(folderPath, description, photoType string) error

// Управление очередью обработки
StartQueueProcessing() error  
StopQueueProcessing() error
GetQueueStatus() ([]models.BatchStatus, error)

// Получение результатов обработки
GetProcessedBatches() ([]models.PhotoBatch, error)
GetBatchDetails(batchID string) (*models.PhotoBatch, error)
GetBatchPhotos(batchID string) ([]models.Photo, error)
```

### Методы просмотра и редактирования

```go
// Управление статусами фотографий
ApprovePhoto(photoID string) error
RejectPhoto(photoID string) error
ResetPhotoToProcessed(photoID string) error

// Редактирование метаданных
UpdatePhotoMetadata(photoID string, aiResult models.AIResult) error
RegeneratePhotoMetadata(photoID string, customPrompt string) error
```

### Методы загрузки на стоки

```go
// Управление выбором для загрузки  
SetPhotoSelectedForUpload(photoID string, selected bool) error
SelectAllPhotosForUpload(batchID string) error
ClearAllPhotoSelection(batchID string) error

// Загрузка фотографий
UploadSelectedPhotos(batchID string, photoIDs []string) error
UploadApprovedPhotos(batchID string) error

// Управление очередью загрузки
GetUploadQueueStatus() map[string]interface{}
StopUploadQueue() error
```

### Методы настроек

```go
// Настройки приложения
GetSettings() (models.AppSettings, error)
SaveSettings(settings models.AppSettings) error

// Управление стоками
GetStockConfigs() ([]models.StockConfig, error)
SaveStockConfig(config models.StockConfig) error
DeleteStockConfig(stockID string) error
TestStockConnection(config models.StockConfig) error

// AI промпты  
UpdateAIPrompt(photoType string, prompt string) error
ForceUpdateDefaultPrompts() error
```

### Frontend API (JavaScript)

```javascript
// Основные операции через Wails binding
await window.go.main.App.ProcessPhotoFolder(folderPath, description, type)
await window.go.main.App.GetQueueStatus()
await window.go.main.App.UploadSelectedPhotos(batchId, photoIds)

// Upload Manager
const uploadManager = new UploadManager(app)
uploadManager.selectAllPhotos()
uploadManager.clearAllSelection()
uploadManager.uploadSelectedPhotos()
```

---

## Безопасность

### Защита API ключей

**Хранение в базе данных**:
- API ключи хранятся в зашифрованном виде
- Пароли шифруются перед сохранением
- Никогда не логируются в открытом виде

**Файлы исключенные из Git**:
```gitignore
# База данных с чувствительными данными
*.db
*.sqlite
*.sqlite3

# Временные файлы
temp/
logs/ 
*.log

# Конфигурации с API ключами
.env
secrets.json
credentials.json
```

### Безопасность сети

**FTP/SFTP**:
- Таймауты подключения
- Валидация параметров
- Поддержка SSL/TLS

**API**:
- Проверка SSL сертификатов
- Валидация API ключей
- Ограничения размера файлов

### Валидация файлов

**Безопасность путей**:
```go
// Защита от directory traversal
filepath.Clean(userPath)
filepath.Join(safeDir, relativePath)
```

**Поддерживаемые форматы**:
- JPEG, PNG, TIFF, WEBP
- Проверка магических байтов файлов
- Ограничения размера файлов

---

## Разработка

### Локальная разработка

```bash
# Запуск в режиме разработки
wails dev

# Доступ к dev серверу
# http://localhost:34115 - тестирование Go методов в браузере
```

### Сборка релиза

```bash
# Сборка для текущей платформы
wails build

# Сборка для всех платформ
wails build -platform windows/amd64,darwin/amd64,darwin/arm64,linux/amd64
```

### Добавление нового загрузчика

**1. Создать загрузчик**:
```go
// uploaders/my_uploader.go
type MyUploader struct {
    *BaseUploader
}

func (u *MyUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
    // Реализация загрузки
}
```

**2. Зарегистрировать в менеджере**:
```go
// uploaders/manager.go  
manager.RegisterUploader("my_type", NewMyUploader())
```

**3. Добавить шаблон конфигурации**:
```go
"my_type": {
    Type: "my_type",
    Name: "My Custom Uploader",
    Fields: []models.TemplateField{
        // поля настроек
    },
}
```

### Система логирования

**Отдельные лог файлы**:
- `logs/ai_YYYY-MM-DD.log` - AI операции
- `logs/exif_YYYY-MM-DD.log` - EXIF операции  
- `logs/debug_YYYY-MM-DD.log` - отладочная информация
- `logs/error_YYYY-MM-DD.log` - ошибки

**Автоочистка старых логов**:
```go
// Очистка логов старше 7 дней
app.CleanOldLogs(7)
```

### Тестирование

**Mock режим**:
- Автоматическое определение если Wails недоступен
- Имитация AI API ответов
- Тестовые данные для разработки UI

**Проверка подключений**:
```go
// Тест AI API
app.TestAIConnection(settings)

// Тест стокового подключения  
app.TestStockConnection(config)
```

### Расширение AI провайдеров

**Добавление нового провайдера**:
```go
// services/ai_service.go
case "new_provider":
    return s.analyzeWithNewProvider(photo, description, prompt, contentType, settings)
```

**Обновление настроек**:
```javascript
// frontend - добавить в AI providers список
{ value: "new_provider", text: "New AI Provider" }
```

---

## Примеры использования

### Типичный workflow

1. **Подготовка фотографий**:
   ```
   photos/
   ├── commercial/
   │   ├── business-meeting-01.jpg
   │   └── family-kitchen-02.jpg  
   └── editorial/
       ├── protest-berlin-03.jpg
       └── conference-tech-04.jpg
   ```

2. **Обработка через приложение**:
   - Drag & Drop папки на Editorial/Commercial вкладки
   - Добавление описания: "Business photos from corporate event"
   - Запуск обработки и ожидание AI анализа

3. **Просмотр и редактирование**:
   - Переход на вкладку Review
   - Выбор обработанного батча
   - Редактирование AI результатов при необходимости
   - Approve нужных фотографий

4. **Загрузка на стоки**:
   - Выбор фотографий через чекбоксы
   - Нажатие "Upload Selected"
   - Отслеживание прогресса загрузки

### Пример AI результата

**Commercial фото**:
```json
{
  "title": "Business team collaboration meeting in modern office",
  "description": "Diverse group of professionals working together on laptops in contemporary workplace, representing teamwork and corporate success",
  "keywords": ["business", "team", "meeting", "office", "professional", "collaboration", "laptop", "modern", "workplace", "corporate", "success", "teamwork", "diverse", "working"],
  "category": "Business"
}
```

**Editorial фото**:
```json  
{
  "title": "Environmental activists protest outside Parliament building",
  "description": "Demonstrators holding signs and banners gather in front of Parliament building to protest government environmental policies and demand immediate climate action legislation",
  "keywords": ["protest", "Parliament", "environment", "climate change", "activism", "demonstration", "government", "politics", "crowd", "signs", "banners", "legislation", "policy"],
  "category": "Politics"
}
```

---

## Заключение

Stock Photo Automation App представляет собой комплексное решение для автоматизации работы с стоковыми фотографиями. Архитектура приложения обеспечивает:

- **🔄 Масштабируемость**: модульная система загрузчиков и очередей
- **🛡️ Надежность**: error handling и graceful fallbacks
- **⚡ Производительность**: параллельная обработка и оптимизация ресурсов  
- **🎛️ Гибкость**: кастомизируемые промпты и настройки
- **👥 Удобство**: интуитивный UI и автоматизация рутинных процессов

Документация будет обновляться по мере развития проекта и добавления новых функций.