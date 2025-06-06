package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// EXIFProcessor обрабатывает EXIF данные для разных типов контента
type EXIFProcessor struct{}

// NewEXIFProcessor создает новый обработчик EXIF
func NewEXIFProcessor() *EXIFProcessor {
	return &EXIFProcessor{}
}

// ProcessForEditorial извлекает максимум данных для редакционных фото
func (e *EXIFProcessor) ProcessForEditorial(exifData map[string]string) EditorialEXIFData {
	data := EditorialEXIFData{
		TechnicalInfo: e.extractTechnicalInfo(exifData),
		LocationInfo:  e.extractLocationInfo(exifData),
		DateTimeInfo:  e.extractDateTimeInfo(exifData),
		CameraInfo:    e.extractCameraInfo(exifData),
	}

	log.Printf("Editorial EXIF processing: extracted location=%v, datetime=%v",
		data.LocationInfo.HasLocation, data.DateTimeInfo.HasDateTime)

	return data
}

// ProcessForCommercial извлекает только техническую информацию для коммерческих фото
func (e *EXIFProcessor) ProcessForCommercial(exifData map[string]string) CommercialEXIFData {
	data := CommercialEXIFData{
		TechnicalInfo: e.extractTechnicalInfo(exifData),
		CameraInfo:    e.extractCameraInfo(exifData),
		// НЕ извлекаем местоположение и конкретные даты!
	}

	log.Printf("Commercial EXIF processing: extracted only technical data, location ignored")

	return data
}

// extractLocationInfo извлекает данные о местоположении (только для Editorial)
func (e *EXIFProcessor) extractLocationInfo(exifData map[string]string) LocationInfo {
	info := LocationInfo{}

	// GPS координаты
	if lat, exists := exifData["GPS Latitude"]; exists && lat != "" {
		info.GPSLatitude = lat
		info.HasLocation = true
	}

	if lon, exists := exifData["GPS Longitude"]; exists && lon != "" {
		info.GPSLongitude = lon
		info.HasLocation = true
	}

	// Данные о местоположении
	if city := exifData["GPS City"]; city != "" {
		info.City = city
		info.HasLocation = true
	}

	if country := exifData["GPS Country"]; country != "" {
		info.Country = country
		info.HasLocation = true
	}

	if region := exifData["GPS State"]; region != "" {
		info.Region = region
		info.HasLocation = true
	}

	// Альтернативные поля местоположения
	locationFields := []string{
		"Location", "City", "Country", "State", "Province",
		"GPS Position", "GPS Location", "Location Created",
	}

	for _, field := range locationFields {
		if value, exists := exifData[field]; exists && value != "" {
			if info.Description == "" {
				info.Description = value
			} else {
				info.Description += ", " + value
			}
			info.HasLocation = true
		}
	}

	return info
}

// extractDateTimeInfo извлекает информацию о дате и времени (только для Editorial)
func (e *EXIFProcessor) extractDateTimeInfo(exifData map[string]string) DateTimeInfo {
	info := DateTimeInfo{}

	// Основные поля даты
	dateFields := []string{
		"DateTime Original", "Date/Time Original", "DateTimeOriginal",
		"DateTime", "Date/Time", "Create Date", "Date Created",
	}

	for _, field := range dateFields {
		if dateStr, exists := exifData[field]; exists && dateStr != "" {
			if parsedTime, err := e.parseDateTime(dateStr); err == nil {
				info.DateTime = parsedTime
				info.HasDateTime = true
				info.DateString = dateStr
				break
			}
		}
	}

	// Временная зона
	if tz, exists := exifData["Time Zone"]; exists && tz != "" {
		info.TimeZone = tz
	}

	return info
}

// extractTechnicalInfo извлекает техническую информацию (для обоих типов)
func (e *EXIFProcessor) extractTechnicalInfo(exifData map[string]string) TechnicalInfo {
	info := TechnicalInfo{}

	// Размеры изображения
	if width, exists := exifData["Image Width"]; exists {
		if w, err := strconv.Atoi(width); err == nil {
			info.Width = w
		}
	}

	if height, exists := exifData["Image Height"]; exists {
		if h, err := strconv.Atoi(height); err == nil {
			info.Height = h
		}
	}

	// Технические параметры съемки
	info.ISO = exifData["ISO"]
	info.Aperture = exifData["F Number"]
	info.ShutterSpeed = exifData["Shutter Speed"]
	info.FocalLength = exifData["Focal Length"]
	info.WhiteBalance = exifData["White Balance"]
	info.Flash = exifData["Flash"]

	// Ориентация
	if orientation, exists := exifData["Orientation"]; exists {
		info.Orientation = orientation
	}

	return info
}

// extractCameraInfo извлекает информацию о камере (для обоих типов)
func (e *EXIFProcessor) extractCameraInfo(exifData map[string]string) CameraInfo {
	info := CameraInfo{}

	info.Make = exifData["Camera Make"]
	info.Model = exifData["Camera Model"]
	info.LensModel = exifData["Lens Model"]
	info.Software = exifData["Software"]

	// Альтернативные названия полей
	if info.Make == "" {
		info.Make = exifData["Make"]
	}
	if info.Model == "" {
		info.Model = exifData["Model"]
	}

	return info
}

// parseDateTime пытается распарсить дату из различных форматов
func (e *EXIFProcessor) parseDateTime(dateStr string) (time.Time, error) {
	// Стандартные форматы EXIF
	formats := []string{
		"2006:01:02 15:04:05",
		"2006-01-02 15:04:05",
		"2006:01:02T15:04:05",
		"2006-01-02T15:04:05",
		"2006:01:02",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// BuildContextualPrompt создает контекстный промпт на основе EXIF данных
func (e *EXIFProcessor) BuildContextualPrompt(contentType string, basePrompt string, exifData map[string]string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString(basePrompt)

	if contentType == "editorial" {
		// Для editorial добавляем ВСЮ доступную контекстную информацию
		editorialData := e.ProcessForEditorial(exifData)

		if editorialData.LocationInfo.HasLocation {
			promptBuilder.WriteString("\n\nИНФОРМАЦИЯ О МЕСТОПОЛОЖЕНИИ:")
			if editorialData.LocationInfo.City != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- Город: %s", editorialData.LocationInfo.City))
			}
			if editorialData.LocationInfo.Country != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- Страна: %s", editorialData.LocationInfo.Country))
			}
			if editorialData.LocationInfo.Region != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- Регион: %s", editorialData.LocationInfo.Region))
			}
			if editorialData.LocationInfo.Description != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- Описание локации: %s", editorialData.LocationInfo.Description))
			}
		}

		if editorialData.DateTimeInfo.HasDateTime {
			promptBuilder.WriteString("\n\nИНФОРМАЦИЯ О ВРЕМЕНИ:")
			promptBuilder.WriteString(fmt.Sprintf("\n- Дата съемки: %s", editorialData.DateTimeInfo.DateString))
			if editorialData.DateTimeInfo.TimeZone != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- Временная зона: %s", editorialData.DateTimeInfo.TimeZone))
			}
		}

		promptBuilder.WriteString("\n\nТребования: Используй ВСЮ доступную контекстную информацию для создания точного, фактического описания.")

	} else {
		// Для commercial НЕ добавляем местоположение и даты
		commercialData := e.ProcessForCommercial(exifData)

		promptBuilder.WriteString("\n\nТЕХНИЧЕСКАЯ ИНФОРМАЦИЯ:")
		if commercialData.TechnicalInfo.Width > 0 && commercialData.TechnicalInfo.Height > 0 {
			promptBuilder.WriteString(fmt.Sprintf("\n- Размер: %dx%d",
				commercialData.TechnicalInfo.Width, commercialData.TechnicalInfo.Height))
		}

		if commercialData.CameraInfo.Make != "" || commercialData.CameraInfo.Model != "" {
			promptBuilder.WriteString(fmt.Sprintf("\n- Камера: %s %s",
				commercialData.CameraInfo.Make, commercialData.CameraInfo.Model))
		}

		promptBuilder.WriteString("\n\nТребования: Создай УНИВЕРСАЛЬНОЕ описание БЕЗ конкретных мест, дат и брендов.")
	}

	return promptBuilder.String()
}

// Структуры данных для разных типов обработки

type EditorialEXIFData struct {
	TechnicalInfo TechnicalInfo
	LocationInfo  LocationInfo
	DateTimeInfo  DateTimeInfo
	CameraInfo    CameraInfo
}

type CommercialEXIFData struct {
	TechnicalInfo TechnicalInfo
	CameraInfo    CameraInfo
	// НЕ включаем LocationInfo и DateTimeInfo!
}

type LocationInfo struct {
	HasLocation  bool
	GPSLatitude  string
	GPSLongitude string
	City         string
	Country      string
	Region       string
	Description  string
}

type DateTimeInfo struct {
	HasDateTime bool
	DateTime    time.Time
	DateString  string
	TimeZone    string
}

type TechnicalInfo struct {
	Width        int
	Height       int
	ISO          string
	Aperture     string
	ShutterSpeed string
	FocalLength  string
	WhiteBalance string
	Flash        string
	Orientation  string
}

type CameraInfo struct {
	Make      string
	Model     string
	LensModel string
	Software  string
}
