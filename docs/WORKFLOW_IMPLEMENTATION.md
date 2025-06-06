# Полный Workflow Обработки Фотографий

## Обзор

Этот документ описывает полный workflow обработки фотографий в Stock Photo Automation App, от добавления в очередь до загрузки на стоковые площадки.

## Архитектура

### Основные компоненты

1. **QueueManager** (`services/queue_manager.go`) - управляет очередью обработки
2. **AIService** (`services/ai_service.go`) - интеграция с AI для анализа фотографий
3. **ImageProcessor** (`services/image_processor.go`) - обработка изображений и EXIF
4. **EXIFProcessor** (`services/exif_processor.go`) - специализированная обработка EXIF данных
5. **DatabaseService** (`services/database_service.go`) - работа с базой данных
6. **UploaderManager** (`uploaders/uploader_manager.go`) - загрузка на стоковые площадки

### Модели данных

- **PhotoBatch** - группа фотографий для обработки
- **Photo** - отдельная фотография с метаданными
- **AIResult** - результаты анализа нейросетью
- **ProcessingJob** - активная задача обработки

## Пошаговый Workflow

### 1. Добавление в очередь (ProcessPhotoFolder)

```go
// app.go
func (a *App) ProcessPhotoFolder(folderPath string, description string, photoType string) error
```

**Шаги:**
1. Создается новый `PhotoBatch` с типом (editorial/commercial)
2. Сканируется папка для поиска изображений (`ImageProcessor.ScanFolder`)
3. Устанавливается `ContentType` для всех фотографий
4. Батч добавляется в очередь (`QueueManager.AddBatch`)
5. Запускается обработка очереди если не запущена

### 2. Обработка очереди (QueueManager)

```go
// services/queue_manager.go
func (q *QueueManager) processQueue(settings models.AppSettings)
```

**Основной цикл:**
1. Получение следующего батча из очереди (`getNextBatch`)
2. Проверка лимита одновременных задач
3. Обработка батча (`processBatch`)
4. Обновление статусов в базе данных

### 3. Обработка одного батча

```go
func (q *QueueManager) processBatch(batch models.PhotoBatch, settings models.AppSettings) error
```

**Шаги:**
1. Создание `ProcessingJob` для отслеживания прогресса
2. Обновление статуса батча на "processing"
3. Обработка каждого фото в цикле (`processPhoto`)
4. Обновление прогресса в реальном времени
5. Завершение с статусом "processed" или "failed"

### 4. Обработка одного фото

```go
func (q *QueueManager) processPhoto(photo *models.Photo, batchDescription string, contentType string, settings models.AppSettings) error
```

**Шаги:**
1. **Подготовка для AI** (`ImageProcessor.ProcessPhotoForAI`):
   - Создание миниатюры для экономии токенов
   - Извлечение EXIF данных
   
2. **AI анализ** (`AIService.AnalyzePhoto`):
   - Выбор промпта на основе типа контента (editorial/commercial)
   - Построение контекстного промпта с EXIF данными
   - Отправка запроса в AI API
   - Парсинг JSON ответа
   
3. **Сохранение результатов** (`DatabaseService.UpdatePhotoAIResults`):
   - Сохранение AI результатов в базе данных
   - Обновление статуса фото на "processed"
   
4. **Запись в EXIF** (`ImageProcessor.WriteExifToImage`):
   - Запись метаданных в оригинальный файл
   - Обработка ошибок без прерывания процесса

### 5. Специализированная обработка EXIF

```go
// services/exif_processor.go
func (e *EXIFProcessor) ProcessForEditorial(exifData map[string]string) EditorialEXIFData
func (e *EXIFProcessor) ProcessForCommercial(exifData map[string]string) CommercialEXIFData
```

**Editorial обработка:**
- Извлекает максимум контекстной информации
- GPS координаты, город, страна
- Точные даты и время
- Техническая информация камеры

**Commercial обработка:**
- Извлекает только техническую информацию
- Игнорирует местоположение и даты
- Фокус на универсальности

### 6. Построение контекстного промпта

```go
func (e *EXIFProcessor) BuildContextualPrompt(contentType string, basePrompt string, exifData map[string]string) string
```

**Для Editorial:**
- Добавляет GPS координаты и местоположение
- Включает точную дату и время съемки
- Добавляет техническую информацию

**Для Commercial:**
- Добавляет только техническую информацию
- Исключает локационные данные
- Фокус на универсальных характеристиках

## UI Workflow

### 1. Вкладки обработки (Editorial/Commercial)

**Функциональность:**
- Drag & Drop папок с фотографиями
- Поле описания для всего батча
- Кнопка запуска обработки
- Автоматическое переключение на вкладку Queue

### 2. Вкладка Queue (Очередь обработки)

**Отображение:**
- Список активных батчей
- Прогресс обработки в реальном времени
- Текущее обрабатываемое фото
- Статусы: queued, processing, completed, failed

### 3. Вкладка Review (Просмотр результатов)

**Функциональность:**
- Выбор обработанного батча
- Просмотр результатов AI анализа
- Редактирование метаданных
- Подтверждение/отклонение фотографий
- Повторная генерация с кастомным промптом

### 4. Вкладка History (История)

**Отображение:**
- История всех обработанных батчей
- Детальная информация по каждому батчу
- Статистика обработки

## API Методы

### Основные методы обработки

```go
// Добавление в очередь
ProcessPhotoFolder(folderPath, description, photoType string) error

// Управление очередью
StartQueueProcessing() error
StopQueueProcessing() error
GetQueueStatus() ([]models.BatchStatus, error)

// Просмотр результатов
GetBatchDetails(batchID string) (*models.PhotoBatch, error)
ApprovePhoto(photoID string) error
RejectPhoto(photoID string) error

// Редактирование метаданных
UpdatePhotoMetadata(photoID string, aiResult models.AIResult) error
RegeneratePhotoMetadata(photoID string, customPrompt string) error
```

### Методы загрузки

```go
// Загрузка на стоки (будет реализовано)
UploadApprovedPhotos(batchID string) error
UploadPhoto(photoID string, stockIDs []string) error
```

## База данных

### Таблицы

1. **batches** - информация о батчах
2. **photos** - информация о фотографиях
3. **app_settings** - настройки приложения
4. **stock_configs** - конфигурации стоковых площадок

### Миграции

- Автоматическое добавление новых полей
- Обновление существующих данных
- Поддержка обратной совместимости

## Обработка ошибок

### Уровни ошибок

1. **Критические** - останавливают обработку батча
2. **Предупреждения** - логируются, но не прерывают процесс
3. **Информационные** - для отладки

### Retry логика

- Повторные попытки для AI API
- Обработка временных сбоев сети
- Graceful degradation при ошибках EXIF

## Производительность

### Оптимизации

1. **Миниатюры** - уменьшение размера для AI анализа
2. **Пул воркеров** - ограничение одновременных задач
3. **Кэширование** - результаты EXIF обработки
4. **Lazy loading** - загрузка изображений по требованию

### Мониторинг

- Прогресс обработки в реальном времени
- Статистика производительности
- Логирование всех операций

## Безопасность

### Валидация

- Проверка путей файлов
- Валидация входных данных
- Санитизация пользовательского ввода

### Конфиденциальность

- Шифрование API ключей
- Безопасное хранение паролей
- Логирование без чувствительных данных

## Расширяемость

### Новые AI провайдеры

- Модульная архитектура AI сервиса
- Легкое добавление новых провайдеров
- Унифицированный интерфейс

### Новые стоковые площадки

- Плагинная система загрузчиков
- Шаблоны конфигураций
- Кастомные модули

## Заключение

Реализованный workflow обеспечивает:

1. **Автоматизацию** - минимальное участие пользователя
2. **Гибкость** - поддержка разных типов контента
3. **Надежность** - обработка ошибок и восстановление
4. **Масштабируемость** - обработка больших объемов
5. **Удобство** - интуитивный интерфейс

Система готова для продуктивного использования фотографом для автоматизации рутинных процессов размещения фотографий на стоковых платформах. 