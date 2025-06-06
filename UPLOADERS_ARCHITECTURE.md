# Архитектура Модульных Загрузчиков Стоков

## Обзор

Новая архитектура поддерживает разные типы подключений к стоковым сайтам через модульную систему загрузчиков. Каждый тип подключения (FTP, SFTP, API) имеет свой собственный модуль с возможностью легкого добавления новых типов.

## Структура

### Основные компоненты

1. **BaseUploader** (`uploaders/base.go`) - базовый класс для всех загрузчиков
2. **UploaderManager** (`uploaders/manager.go`) - менеджер для управления загрузчиками
3. **Конкретные загрузчики**:
   - FTPUploader (`uploaders/ftp_uploader.go`)
   - SFTPUploader (`uploaders/sftp_uploader.go`) 
   - APIUploader (`uploaders/api_uploader.go`)

### Интерфейс StockUploader

Каждый загрузчик должен реализовать интерфейс:

```go
type StockUploader interface {
    Upload(photo Photo, config StockConfig) (UploadResult, error)
    TestConnection(config StockConfig) error
    GetInfo() UploaderInfo
    ValidateConfig(config StockConfig) error
}
```

## Поддерживаемые Типы Подключений

### FTP
- **Порт по умолчанию**: 21
- **Настройки**: host, port, username, password, path, passive, timeout
- **Особенности**: поддержка пассивного режима

### SFTP
- **Порт по умолчанию**: 22
- **Настройки**: host, port, username, password, path, timeout
- **Особенности**: SSH-ключи, автоматическое создание папок

### API
- **Настройки**: apiUrl, apiKey, timeout, headers, params
- **Особенности**: multipart upload, передача метаданных

## Конфигурация Стоков

### Новые поля в StockConfig

```go
type StockConfig struct {
    Type           string                 // "ftp", "sftp", "api", "custom"
    Settings       map[string]interface{} // дополнительные настройки
    ModulePath     string                 // путь к модулю для кастомных типов
    // ... существующие поля
}
```

### Миграция базы данных

Автоматически добавляются новые поля:
- `type` - тип загрузчика
- `settings` - JSON с дополнительными настройками
- `module_path` - путь к кастомному модулю

## Шаблоны Конфигурации

Каждый тип имеет шаблон с описанием полей:

```go
type StockTemplate struct {
    Type        string
    Name        string
    Description string
    Fields      []TemplateField
    Defaults    map[string]interface{}
}
```

### Типы полей

- `text` - текстовое поле
- `password` - поле пароля
- `number` - числовое поле
- `url` - URL поле
- `checkbox` - чекбокс
- `select` - выпадающий список

## Фронтенд

### Динамическое создание полей

При выборе типа подключения автоматически создаются соответствующие поля формы на основе шаблона.

### Новые переводы

Добавлены переводы для всех типов подключений и полей в `frontend/src/locales/`.

## API Методы

### Backend (app.go)

- `GetStockTemplates()` - получить шаблоны типов
- `GetAvailableUploaders()` - список загрузчиков
- `ValidateStockConfig()` - валидация конфигурации
- `TestStockConnection()` - тест подключения

### Frontend (main.js)

- `onStockTypeChange()` - обработка смены типа
- `createDynamicField()` - создание динамических полей
- `getStockTemplates()` - получение шаблонов

## Добавление Нового Типа Загрузчика

### 1. Создать новый модуль

```go
// uploaders/custom_uploader.go
type CustomUploader struct {
    *BaseUploader
}

func NewCustomUploader() *CustomUploader {
    info := models.UploaderInfo{
        Name:        "Custom Uploader",
        Type:        "custom",
        // ...
    }
    return &CustomUploader{
        BaseUploader: NewBaseUploader(info),
    }
}

// Реализовать методы интерфейса
func (u *CustomUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
    // Логика загрузки
}
```

### 2. Зарегистрировать в менеджере

```go
// uploaders/manager.go
func NewUploaderManager() *UploaderManager {
    manager := &UploaderManager{
        uploaders: make(map[string]models.StockUploader),
    }
    
    // ... существующие
    manager.RegisterUploader("custom", NewCustomUploader())
    
    return manager
}
```

### 3. Добавить шаблон

```go
// uploaders/base.go
func GetStockTemplates() map[string]models.StockTemplate {
    return map[string]models.StockTemplate{
        // ... существующие
        "custom": {
            Type: "custom",
            Name: "Custom Upload",
            Fields: []models.TemplateField{
                // определить поля
            },
        },
    }
}
```

### 4. Обновить переводы

Добавить переводы в `frontend/src/locales/`.

## Обратная Совместимость

Существующие конфигурации продолжат работать:
- Поле `uploadMethod` автоматически копируется в `type` при миграции
- Если `type` не указан, используется `uploadMethod`

## Безопасность

### FTP/SFTP
- Таймауты подключения
- Валидация параметров подключения

### API
- Проверка SSL сертификатов
- Валидация API ключей
- Ограничения на размер файлов

## Логирование

Все загрузчики логируют:
- Успешные подключения
- Ошибки загрузки
- Время выполнения операций

## Тестирование

Каждый загрузчик поддерживает тестирование подключения через метод `TestConnection()`.

## Производительность

- Пулы соединений для FTP/SFTP
- Параллельная загрузка файлов
- Настраиваемые таймауты

---

Эта архитектура обеспечивает гибкость для добавления новых типов стоков и легкость в поддержке существующих. 