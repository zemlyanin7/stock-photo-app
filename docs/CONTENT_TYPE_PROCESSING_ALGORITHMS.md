# Алгоритмы обработки Editorial vs Commercial фотографий

## Обзор

В приложении есть две отдельные вкладки для разных типов контента. Пользователь самостоятельно выбирает тип контента при загрузке. Каждый тип требует специфического алгоритма обработки данных.

## 🗞️ EDITORIAL - Редакционные фотографии

### Характеристики
- Документальные снимки реальных событий
- Новостная и общественная значимость  
- Требования к фактической точности
- Конкретные люди, места и события

### Алгоритм обработки EXIF данных

#### ✅ ЧТО ИЗВЛЕКАЕМ:
```go
EditorialEXIFData {
    LocationInfo: {
        GPSLatitude:  "55.7558° N"     // GPS координаты - ВАЖНО!
        GPSLongitude: "37.6176° E"     // GPS координаты - ВАЖНО!
        City:        "Moscow"          // Город - КРИТИЧНО!
        Country:     "Russia"          // Страна - КРИТИЧНО!
        Region:      "Moscow Oblast"   // Регион/область
        Description: "Red Square, Kremlin" // Конкретное место
    }
    DateTimeInfo: {
        DateTime:    "2024-03-15 14:30:00" // Точная дата - ВАЖНО!
        DateString:  "2024:03:15 14:30:00" // Оригинальная строка
        TimeZone:    "UTC+3"               // Временная зона
    }
    TechnicalInfo: { /* базовые технические данные */ }
    CameraInfo:    { /* информация о камере */ }
}
```

#### 🎯 КАК ИСПОЛЬЗУЕМ В AI ПРОМПТЕ:
```markdown
ИНФОРМАЦИЯ О МЕСТОПОЛОЖЕНИИ:
- Город: Moscow
- Страна: Russia  
- Регион: Moscow Oblast
- Описание локации: Red Square, Kremlin

ИНФОРМАЦИЯ О ВРЕМЕНИ:
- Дата съемки: 2024:03:15 14:30:00
- Временная зона: UTC+3

Требования: Используй ВСЮ доступную контекстную информацию для создания точного, фактического описания.
```

#### 📝 РЕЗУЛЬТАТ AI АНАЛИЗА:
```json
{
  "title": "Political rally at Red Square in Moscow, March 15, 2024",
  "description": "Large crowd gathers at Red Square near Kremlin in Moscow, Russia during political demonstration on March 15, 2024. Participants holding banners and flags in front of historic St. Basil's Cathedral.",
  "keywords": ["Moscow", "Red Square", "Kremlin", "Russia", "political rally", "demonstration", "March 2024", "crowd", "St Basil Cathedral", "news event", "politics", "government", "protest", "public gathering", "Russian politics"]
}
```

### 🔍 Валидация Editorial контента:
- ✅ Проверяем наличие конкретных мест
- ✅ Проверяем указание дат
- ✅ Проверяем фактическую информацию
- ✅ Проверяем новостную ценность

---

## 💼 COMMERCIAL - Коммерческие фотографии

### Характеристики
- Постановочные сцены для рекламы
- Универсальность применения
- Отсутствие конкретных привязок
- Концептуальность и эмоциональность

### Алгоритм обработки EXIF данных

#### ✅ ЧТО ИЗВЛЕКАЕМ:
```go
CommercialEXIFData {
    TechnicalInfo: {
        Width:        "4000"        // Размер изображения
        Height:       "3000"        // Размер изображения  
        ISO:          "100"         // Технические параметры
        Aperture:     "f/2.8"       // Технические параметры
        ShutterSpeed: "1/250"       // Технические параметры
        FocalLength:  "85mm"        // Технические параметры
        Orientation:  "horizontal"  // Ориентация
    }
    CameraInfo: {
        Make:      "Canon"          // Марка камеры (без брендинга в контенте!)
        Model:     "EOS R5"         // Модель камеры
        LensModel: "RF 85mm f/1.2"  // Объектив
    }
    // НЕ ВКЛЮЧАЕМ:
    // ❌ LocationInfo - местоположение ИГНОРИРУЕМ!
    // ❌ DateTimeInfo - конкретные даты ИГНОРИРУЕМ!
}
```

#### ⚠️ ЧТО НЕ ИЗВЛЕКАЕМ:
```go
// ❌ ИГНОРИРУЕМ ЭТИ ДАННЫЕ:
LocationInfo {
    GPSLatitude:  // НЕ используем! 
    GPSLongitude: // НЕ используем!
    City:        // НЕ используем!
    Country:     // НЕ используем!
    // Конкретные места мешают универсальности!
}

DateTimeInfo {
    DateTime:    // НЕ используем!
    DateString:  // НЕ используем!
    // Конкретные даты мешают универсальности!
}
```

#### 🎯 КАК ИСПОЛЬЗУЕМ В AI ПРОМПТЕ:
```markdown
ТЕХНИЧЕСКАЯ ИНФОРМАЦИЯ:
- Размер: 4000x3000
- Камера: Canon EOS R5

Требования: Создай УНИВЕРСАЛЬНОЕ описание БЕЗ конкретных мест, дат и брендов.
```

#### 📝 РЕЗУЛЬТАТ AI АНАЛИЗА:
```json
{
  "title": "Happy family cooking together in modern kitchen",
  "description": "Diverse family enjoying cooking time in contemporary home kitchen setting, representing togetherness and lifestyle",
  "keywords": ["family", "cooking", "kitchen", "home", "lifestyle", "togetherness", "modern", "happy", "diverse", "contemporary", "domestic", "food preparation", "bonding", "household", "indoor"]
}
```

### 🔍 Валидация Commercial контента:
- ❌ Проверяем ОТСУТСТВИЕ конкретных мест
- ❌ Проверяем ОТСУТСТВИЕ конкретных дат  
- ❌ Проверяем ОТСУТСТВИЕ брендов
- ✅ Проверяем универсальность терминов
- ✅ Проверяем коммерческую применимость

---

## 📊 Сравнительная таблица

| Аспект | Editorial | Commercial |
|--------|-----------|------------|
| **GPS координаты** | ✅ КРИТИЧНО использовать | ❌ ИГНОРИРОВАТЬ |
| **Конкретные места** | ✅ "Red Square, Moscow" | ❌ "modern office" |
| **Даты съемки** | ✅ "March 15, 2024" | ❌ Не упоминать |
| **Имена людей** | ✅ "President Biden" | ❌ "businessman" |
| **Бренды/логотипы** | ✅ Допустимо если новостное | ❌ ЗАПРЕЩЕНО |
| **Эмоциональность** | ❌ Избегать | ✅ "happy, successful" |
| **Универсальность** | ❌ Конкретность важнее | ✅ КРИТИЧНО |

## 🔧 Практическая реализация

### Настройка в AI сервисе:

```go
func (s *AIService) AnalyzePhoto(photo models.Photo, contentType string) {
    if contentType == "editorial" {
        // Извлекаем ВСЕ данные из EXIF
        editorialData := s.exifProcessor.ProcessForEditorial(photo.ExifData)
        prompt := s.buildEditorialPrompt(editorialData)
        
        // Фокус на фактах и конкретности
        result := s.generateWithEditorialRules(photo, prompt)
        
    } else { // commercial
        // Извлекаем ТОЛЬКО техническую информацию  
        commercialData := s.exifProcessor.ProcessForCommercial(photo.ExifData)
        prompt := s.buildCommercialPrompt(commercialData) 
        
        // Фокус на универсальности и концепциях
        result := s.generateWithCommercialRules(photo, prompt)
    }
}
```

### Настройка промптов:

#### Editorial промпт:
```markdown
Используй данные GPS и местоположения для точного указания места съемки.
Включи конкретные даты и время из EXIF данных.
Создай фактическое, журналистское описание события.
```

#### Commercial промпт:  
```markdown
ИГНОРИРУЙ данные GPS и конкретное местоположение.
НЕ указывай конкретные даты съемки.
Создай универсальное описание подходящее для рекламы.
```

## 🚨 Важные выводы

### Для Editorial:
1. **GPS данные критично важны** - они дают контекст места
2. **Даты из EXIF необходимы** - они создают временной контекст  
3. **Конкретность > Универсальность** - точные факты важнее применимости

### Для Commercial:
1. **GPS данные вредят** - они ограничивают применимость
2. **Конкретные даты мешают** - они привязывают к времени
3. **Универсальность > Конкретность** - применимость важнее точности

### Технический вывод:
**Да, нужно два разных алгоритма обработки EXIF данных!**
- Editorial: максимальное извлечение контекстной информации
- Commercial: игнорирование локационных и временных данных 