# Обновление AI Промптов и Категорий - Декабрь 2024

## 🎯 Основные улучшения

### 1. Увеличено количество ключевых слов
- **Было**: 15-25 ключевых слов для Editorial, 15-20 для Commercial
- **Стало**: 48-55 ключевых слов для ВСЕХ типов контента
- **Эффект**: Значительно улучшенная находимость фото в поисковых системах стоков

### 2. Автоматический выбор категорий
- **Новое**: AI автоматически выбирает категорию из стандартного списка стоков
- **ВАЖНО**: Категории теперь РАЗНЫЕ для Editorial и Commercial контента!

**Editorial категории (для новостей и документалистики)**:
News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment, Healthcare, Education, Crime, Religion, Royalty, Awards & Ceremonies

**Commercial категории (для рекламы и маркетинга)**:
Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture, Music, Art & Design, Objects, Concepts, Beauty, Shopping, Transportation, Home & Garden

- **Интерфейс**: Заменено текстовое поле на dropdown с предустановленными опциями, список меняется в зависимости от типа контента

### 3. Расширенная EXIF запись
- **Категории**: Теперь записываются в XMP:Category и IPTC:Category
- **Рейтинг качества**: Записывается в EXIF:Rating и XMP:Rating
- **Полная очистка**: При approve очищаются все связанные поля включая рейтинг

## 🔧 Технические изменения

### Backend (Go)
- `services/database_service.go`: Обновлены промпты по умолчанию
- `app.go`: Обновлены промпты в функции ForceUpdateDefaultPrompts
- `services/ai_service.go`: Расширен JSON Schema, добавлена категория как обязательное поле
- `services/image_processor.go`: Добавлена запись рейтинга и категории в EXIF

### Frontend (JavaScript)
- `frontend/src/main.js`: 
  - Добавлена функция `getCategoryOptions()` 
  - Заменен input на select для категорий
  - Обновлены промпты по умолчанию
  - Динамический список категорий в зависимости от типа контента

### JSON Schema Updates
```json
{
  "required": ["title", "description", "keywords", "category"],
  "properties": {
    "keywords": {
      "description": "Массив ключевых слов (48-55 слов)"
    },
    "category": {
      "type": "string",
      "description": "Категория фотографии из списка стандартных категорий стоков"
    }
  }
}
```

## 📈 Ожидаемые результаты

### Для фотографов
- **Лучшая видимость**: В 2-3 раза больше ключевых слов = больше показов в поиске
- **Профессиональность**: Стандартизированные категории ускоряют модерацию
- **Автоматизация**: Меньше ручной работы при выборе категорий

### Для стоковых площадок
- **Соответствие стандартам**: Использование официальных категорий
- **Лучшая индексация**: Больше релевантных ключевых слов
- **Правильные EXIF**: Категории и рейтинги записываются в метаданные

## 🚀 Как использовать

1. **Обновите промпты**: Новые промпты автоматически применятся при следующем запуске
2. **Проверьте категории**: В интерфейсе Review теперь dropdown вместо текстового поля
3. **AI анализ**: Новые фото будут анализироваться с 48-55 ключевыми словами
4. **EXIF запись**: При approve будут записываться категории и рейтинги

## 🔍 Примеры улучшений

### Commercial Prompt
```
3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   - Демографические: woman, man, family, adult, senior, child, teenager, couple, group
   - Эмоциональные: happy, successful, confident, relaxed, joyful, satisfied, peaceful, excited
   - Концептуальные: business, lifestyle, technology, health, education, finance, teamwork, communication
   - Визуальные: modern, bright, professional, casual, elegant, minimalist, colorful, clean
   - Общие локации: office, home, outdoors, studio, urban, indoor, workplace, public
   - Действия: working, meeting, talking, smiling, thinking, planning, creating, learning

4. КАТЕГОРИЯ (выбери одну из для Commercial):
   Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture, Music, Art & Design, Objects, Concepts, Beauty, Shopping, Transportation, Home & Garden
```

### Editorial Prompt
```
3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   - Конкретные имена публичных лиц
   - Точные географические названия 
   - Названия событий и организаций
   - Новостные категории (politics, economy, sports, entertainment)
   - Временные маркеры (2024, recent, current)

4. КАТЕГОРИЯ (выбери одну из для Editorial):
   News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment, Healthcare, Education, Crime, Religion, Royalty, Awards & Ceremonies
```

## 🔧 Исправлена проблема с шаблонными ключевыми словами

**Проблема**: AI копировал примеры ключевых слов из промпта вместо анализа изображения
**Пример**: На фото архитектуры без людей AI добавлял "woman, man, happy, successful..."

**Решение**: Убраны конкретные списки примеров, добавлены инструкции для анализа:
```
АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
- Люди: возраст, пол, количество, роли (избегай конкретных имен)
- Эмоции: какие эмоции выражают люди или передает изображение
- Концепции: какие идеи, понятия иллюстрирует фото
- Визуальные характеристики: стиль, цвета, композиция, освещение
...
```

## ✅ Готово к тестированию

Все изменения внесены и готовы к тестированию. Рекомендуется:

1. Запустить `wails dev` для проверки работы
2. Протестировать анализ нескольких фото разных типов
3. Проверить работу dropdown категорий
4. Убедиться в записи EXIF при approve
5. **Проверить качество ключевых слов** - теперь должны соответствовать содержанию изображения

Обновление значительно улучшает качество AI анализа и соответствие стандартам стоковых площадок!

# История изменений Stock Photo App

## 2024-12-XX - Исправление AI промптов

### Проблема 1: Неправильные категории для Editorial vs Commercial
**Проблема**: Приложение использовало одинаковые категории для Editorial и Commercial фото, хотя они должны быть разными.

**Решение**: Разделил категории по типам контента:
- **Editorial**: News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment, Healthcare, Education, Crime, Religion, Royalty, Awards & Ceremonies
- **Commercial**: Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture, Music, Art & Design, Objects, Concepts, Beauty, Shopping, Transportation, Home & Garden

**Измененные файлы**:
- `frontend/src/main.js` - функция `getCategoryOptions()`
- `services/database_service.go` - все 4 AI промпта
- `app.go` - функция `ForceUpdateDefaultPrompts`

### Проблема 2: AI копирует шаблонные ключевые слова
**Проблема**: AI копировал примеры ключевых слов из промптов вместо анализа изображений.

**Пример**: для архитектурных фото без людей AI добавлял "woman, man, adult, senior, child, teenager, group, happy, successful, confident, relaxed..."

**Старые промпты содержали**:
```
- Демографические: woman, man, family, adult, senior, child, teenager, couple, group
- Эмоциональные: happy, successful, confident, relaxed, joyful, satisfied, peaceful, excited
```

**Новые промпты**:
```
АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
- Люди: возраст, пол, количество, роли (избегай конкретных имен)
- Эмоции: какие эмоции выражают люди или передает изображение
```

**Измененные файлы**:
- `services/database_service.go` - все 4 промпта обновлены
- `app.go` - функция `ForceUpdateDefaultPrompts`

### Проблема 3: Остатки старых промптов во фронтенде
**Проблема**: Во frontend коде остались старые примеры ключевых слов в функции `resetPromptToDefault()`.

**Решение**: Обновил промпты во фронтенде, убрав все конкретные примеры слов и заменив их инструкциями для анализа.

**Измененные файлы**:
- `frontend/src/main.js` - функция `resetPromptToDefault()` для обоих типов промптов

### Проблема 4: Отсутствие требования английского языка во фронтенде
**Проблема**: При исправлении промптов во фронтенде пропало важное требование о том, что все метаданные должны быть на английском языке.

**Решение**: Добавил обратно требование английского языка в оба промпта:
```
ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
```

**Измененные файлы**:
- `frontend/src/main.js` - добавлено требование в Editorial и Commercial промпты

### Проблема 5: Отсутствие массовых действий и улучшений UX
**Проблема**: Пользователю нужны массовые действия для управления фотографиями в batch и улучшенный UX для regenerate.

**Решение**: Добавлены массовые кнопки и улучшен интерфейс:

**Новые массовые кнопки**:
- **Approve All** - одобрить все фотографии в batch
- **Reject All** - отклонить все фотографии в batch  
- **Regenerate All** - перегенерировать метаданные для всех фотографий

**Улучшения UX для regenerate**:
- Спиннер/крутелка при работе regenerate
- Блокировка кнопки во время выполнения
- Счетчик прогресса для "Regenerate All" (1/10, 2/10 и т.д.)

**Измененные файлы**:
- `frontend/index.html` - добавлены массовые кнопки в batchActions
- `frontend/src/main.js` - добавлены функции `approveAllPhotos()`, `rejectAllPhotos()`, `regenerateAllPhotos()`, улучшена `regeneratePhotoMetadata()`
- `app.go` - добавлен backend метод `SetPhotoStatus()`

### Результат
Теперь AI будет:
1. Использовать правильные категории для Editorial vs Commercial
2. Анализировать реальное содержимое изображений
3. Генерировать релевантные ключевые слова без копирования шаблонов

И пользователи получили:
4. **Массовые действия** для быстрого управления всеми фотографиями
5. **Улучшенный UX** с индикаторами прогресса и блокировкой кнопок

Все изменения готовы для тестирования с `wails dev`. 