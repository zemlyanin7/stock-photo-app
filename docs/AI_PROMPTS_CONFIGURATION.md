# Настройка AI промптов для Editorial и Commercial фотографий

## Обзор

Этот документ содержит практические рекомендации по настройке системных промптов в приложении для корректной генерации метаданных для разных типов фотографий.

## Структура системных промптов

### 1. Промпт для анализа типа контента

```markdown
Проанализируй изображение и определи его тип:

COMMERCIAL (коммерческие стоковые фото):
- Постановочные сцены
- Модели с подписанными релизами  
- Концептуальные изображения для рекламы
- Без конкретных брендов и логотипов
- Студийная или контролируемая съемка

EDITORIAL (редакционные фото):
- Документальные снимки реальных событий
- Новостные и общественно значимые сюжеты
- Естественные моменты без постановки
- Конкретные люди, места и события
- Репортажная съемка

Ответь только: "COMMERCIAL" или "EDITORIAL"
```

### 2. Промпт для COMMERCIAL фотографий

```markdown
Создай метаданные для коммерческого стокового фото:

АНАЛИЗ ИЗОБРАЖЕНИЯ: {image_description}

СОЗДАЙ:
1. НАЗВАНИЕ (до 70 символов):
   - Описательное без конкретных имен и мест
   - Концептуальное (бизнес, семья, технологии)
   - Эмоциональное состояние (счастье, успех)
   - Универсальные формулировки

2. ОПИСАНИЕ (до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Избегание имен и брендов
   - Универсальность для разных контекстов

3. КЛЮЧЕВЫЕ СЛОВА (15-20 слов):
   - Демографические: woman, man, family, adult, senior
   - Эмоциональные: happy, successful, confident, relaxed
   - Концептуальные: business, lifestyle, technology, health
   - Визуальные: modern, bright, professional, casual
   - Общие локации: office, home, outdoors, studio

ИЗБЕГАЙ:
- Конкретные имена людей или компаний
- Названия брендов и торговых марок
- Конкретные географические названия
- Даты и временные привязки
- Политические или спорные темы

ФОРМАТ ОТВЕТА:
Title: [название]
Description: [описание]  
Keywords: [слово1, слово2, слово3, ...]

Дополнительные метаданные:
- Digital Source Type: digitalCapture
- Intellectual Genre: lifestyle
- Model Release: required
- Commercial Use: yes
```

### 3. Промпт для EDITORIAL фотографий

```markdown
Создай метаданные для редакционного фото:

АНАЛИЗ ИЗОБРАЖЕНИЯ: {image_description}

СОЗДАЙ:
1. НАЗВАНИЕ (до 100 символов):
   - Фактическое описание события
   - Конкретные имена людей и мест
   - Журналистский стиль изложения
   - Временной контекст при необходимости

2. ОПИСАНИЕ (до 500 символов):
   - WHO: конкретные имена людей
   - WHAT: точное описание происходящего  
   - WHERE: конкретные места с полными названиями
   - WHEN: даты и время (если применимо)
   - WHY: контекст и причины события

3. КЛЮЧЕВЫЕ СЛОВА (15-25 слов):
   - Конкретные имена публичных лиц
   - Географические названия (города, страны, здания)
   - Названия событий и организаций
   - Новостные категории (politics, economy, sports)
   - Временные маркеры (2024, recent, current)

ВКЛЮЧАЙ:
- Точные названия мест и организаций
- Имена публичных лиц и знаменитостей
- Названия событий и мероприятий
- Политические и общественные темы
- Контекст новостных событий

ФОРМАТ ОТВЕТА:
Title: [название]
Description: [описание]
Keywords: [слово1, слово2, слово3, ...]

Дополнительные метаданные:
- Digital Source Type: digitalCapture
- Intellectual Genre: news
- Model Release: not applicable
- Editorial Use Only: yes
- Event: [название события если применимо]
```

## Примеры настроенных промптов

### Пример 1: Commercial Business Photo

**Входные данные:**
```
Изображение: Группа людей в деловых костюмах сидит за столом с ноутбуками в современном офисе
```

**Промпт результат:**
```
Title: Business team collaboration meeting in modern office
Description: Diverse group of professionals working together on laptops in contemporary workplace setting, representing teamwork and corporate success
Keywords: business, team, meeting, office, professional, collaboration, laptop, modern, workplace, corporate, success, teamwork, diverse, working, contemporary

Digital Source Type: digitalCapture
Intellectual Genre: lifestyle  
Model Release: required
Commercial Use: yes
```

### Пример 2: Editorial News Photo

**Входные данные:**
```
Изображение: Толпа людей с плакатами возле здания парламента
```

**Промпт результат:**
```
Title: Environmental activists protest outside Parliament building demanding climate action
Description: Demonstrators holding signs and banners gather in front of Parliament building to protest government environmental policies and demand immediate climate action legislation
Keywords: protest, Parliament, environment, climate change, activism, demonstration, government, politics, crowd, signs, banners, legislation, policy, activism, environmental protection

Digital Source Type: digitalCapture
Intellectual Genre: news
Model Release: not applicable
Editorial Use Only: yes
Event: Climate Action Protest 2024
```

## Настройка в коде приложения

### 1. Определение промптов в конфигурации

```go
// models/ai_prompts.go
type AIPrompts struct {
    ContentTypeAnalysis string `json:"content_type_analysis"`
    CommercialPrompt    string `json:"commercial_prompt"`
    EditorialPrompt     string `json:"editorial_prompt"`
}

var DefaultPrompts = AIPrompts{
    ContentTypeAnalysis: `Проанализируй изображение и определи его тип...`,
    CommercialPrompt:    `Создай метаданные для коммерческого стокового фото...`,
    EditorialPrompt:     `Создай метаданные для редакционного фото...`,
}
```

### 2. Логика выбора промпта

```go
// services/ai_service.go
func (s *AIService) GenerateMetadata(imageData []byte, userPromptSettings *AIPrompts) (*ImageMetadata, error) {
    // 1. Определяем тип контента
    contentType, err := s.analyzeContentType(imageData)
    if err != nil {
        return nil, err
    }
    
    // 2. Выбираем подходящий промпт
    var prompt string
    switch contentType {
    case "COMMERCIAL":
        prompt = userPromptSettings.CommercialPrompt
    case "EDITORIAL":
        prompt = userPromptSettings.EditorialPrompt
    default:
        prompt = userPromptSettings.CommercialPrompt // по умолчанию
    }
    
    // 3. Генерируем метаданные
    return s.generateWithPrompt(imageData, prompt)
}
```

### 3. Пользовательский интерфейс настроек

```html
<!-- frontend/prompts_settings.html -->
<div class="prompts-configuration">
    <h3>Настройка AI промптов</h3>
    
    <div class="prompt-section">
        <label>Промпт для анализа типа контента:</label>
        <textarea id="content-type-prompt" rows="5">{{.ContentTypeAnalysis}}</textarea>
    </div>
    
    <div class="prompt-section">
        <label>Промпт для коммерческих фото:</label>
        <textarea id="commercial-prompt" rows="10">{{.CommercialPrompt}}</textarea>
    </div>
    
    <div class="prompt-section">
        <label>Промпт для редакционных фото:</label>
        <textarea id="editorial-prompt" rows="10">{{.EditorialPrompt}}</textarea>
    </div>
    
    <button onclick="savePrompts()">Сохранить настройки</button>
    <button onclick="resetToDefaults()">Сбросить к умолчанию</button>
</div>
```

## Валидация результатов

### Проверка для Commercial:
```go
func validateCommercialMetadata(metadata *ImageMetadata) []string {
    var issues []string
    
    // Проверяем отсутствие конкретных имен
    if containsProperNames(metadata.Title) {
        issues = append(issues, "Название содержит конкретные имена")
    }
    
    // Проверяем отсутствие брендов
    if containsBrands(metadata.Keywords) {
        issues = append(issues, "Ключевые слова содержат названия брендов")
    }
    
    return issues
}
```

### Проверка для Editorial:
```go
func validateEditorialMetadata(metadata *ImageMetadata) []string {
    var issues []string
    
    // Проверяем наличие конкретной информации
    if len(metadata.Description) < 100 {
        issues = append(issues, "Описание слишком короткое для editorial")
    }
    
    // Проверяем новостную ценность
    if !hasNewsValue(metadata.Keywords) {
        issues = append(issues, "Отсутствуют новостные ключевые слова")
    }
    
    return issues
}
```

## Рекомендации по использованию

### 1. Тестирование промптов
- Используйте разнообразные изображения для тестирования
- Проверяйте результаты на соответствие требованиям стоков
- Настраивайте промпты на основе обратной связи

### 2. Мониторинг качества
- Ведите статистику одобрения/отклонения на стоках
- Анализируйте причины отклонений
- Корректируйте промпты на основе статистики

### 3. Обновление промптов
- Регулярно обновляйте базовые промпты
- Следите за изменениями требований стоков
- Добавляйте новые тренды и концепции

## Заключение

Правильная настройка AI промптов критично важна для:
- Повышения качества метаданных
- Соответствия требованиям стоковых площадок
- Увеличения процента одобрения фотографий
- Экономии времени на ручную корректировку

Система должна быть гибкой и позволять пользователю настраивать промпты под свои потребности и специфику контента. 