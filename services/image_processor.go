package services

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"stock-photo-app/models"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

type ImageProcessor struct {
	tempDir string
	logger  *Logger
}

func NewImageProcessor(tempDir string) *ImageProcessor {
	// Создаем временную папку если не существует
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Failed to create temp directory: %v", err)
	}

	return &ImageProcessor{
		tempDir: tempDir,
		logger:  nil, // для обратной совместимости
	}
}

func NewImageProcessorWithLogger(tempDir string, logger *Logger) *ImageProcessor {
	// Создаем временную папку если не существует
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		if logger != nil {
			logger.LogError("Failed to create temp directory: %v", err)
		} else {
			log.Printf("Failed to create temp directory: %v", err)
		}
	}

	return &ImageProcessor{
		tempDir: tempDir,
		logger:  logger,
	}
}

// ScanFolder сканирует папку и возвращает список изображений
func (p *ImageProcessor) ScanFolder(folderPath string) ([]models.Photo, error) {
	var photos []models.Photo
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".tiff": true,
		".tif":  true,
	}

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !supportedExts[ext] {
			return nil
		}

		// Получаем информацию о файле
		fileInfo, err := d.Info()
		if err != nil {
			log.Printf("Failed to get file info for %s: %v", path, err)
			return nil
		}

		photo := models.Photo{
			ID:           fmt.Sprintf("photo_%d_%s", time.Now().UnixNano(), d.Name()),
			OriginalPath: path,
			FileName:     d.Name(),
			FileSize:     fileInfo.Size(),
			Status:       "pending",
			CreatedAt:    time.Now(),
		}

		photos = append(photos, photo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan folder: %w", err)
	}

	log.Printf("Found %d images in folder %s", len(photos), folderPath)
	return photos, nil
}

// ScanFolderFiles сканирует папку и возвращает список файлов изображений
func (p *ImageProcessor) ScanFolderFiles(folderPath string) ([]models.PhotoFile, error) {
	var files []models.PhotoFile
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".tiff": true,
		".tif":  true,
		".bmp":  true,
		".webp": true,
	}

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		isValid := supportedExts[ext]

		// Получаем информацию о файле
		fileInfo, err := d.Info()
		if err != nil {
			log.Printf("Failed to get file info for %s: %v", path, err)
			return nil
		}

		file := models.PhotoFile{
			Name:      d.Name(),
			Path:      path,
			Size:      fileInfo.Size(),
			Extension: ext,
			IsValid:   isValid,
		}

		files = append(files, file)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan folder: %w", err)
	}

	log.Printf("Found %d files in folder %s (%d valid images)", len(files), folderPath, countValidFiles(files))
	return files, nil
}

// countValidFiles подсчитывает количество валидных изображений
func countValidFiles(files []models.PhotoFile) int {
	count := 0
	for _, file := range files {
		if file.IsValid {
			count++
		}
	}
	return count
}

// CreateThumbnail создает миниатюру изображения
func (p *ImageProcessor) CreateThumbnail(originalPath string, maxSize int) (string, error) {
	// Открываем оригинальное изображение
	src, err := imaging.Open(originalPath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	// Создаем миниатюру с сохранением пропорций
	thumbnail := imaging.Resize(src, maxSize, 0, imaging.Lanczos)

	// Генерируем имя файла для миниатюры
	baseName := strings.TrimSuffix(filepath.Base(originalPath), filepath.Ext(originalPath))
	thumbnailName := fmt.Sprintf("thumb_%s_%d.jpg", baseName, maxSize)
	thumbnailPath := filepath.Join(p.tempDir, thumbnailName)

	// Сохраняем миниатюру
	err = imaging.Save(thumbnail, thumbnailPath, imaging.JPEGQuality(85))
	if err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	log.Printf("Created thumbnail: %s", thumbnailPath)
	return thumbnailPath, nil
}

// ExtractExifData извлекает EXIF данные из изображения
func (p *ImageProcessor) ExtractExifData(imagePath string) (map[string]string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	exifData, err := exif.Decode(file)
	if err != nil {
		// Если EXIF данных нет, возвращаем пустую карту вместо ошибки
		log.Printf("No EXIF data found in %s: %v", imagePath, err)
		return make(map[string]string), nil
	}

	result := make(map[string]string)

	// Список тегов для извлечения
	tags := map[exif.FieldName]string{
		exif.Make:             "Make",
		exif.Model:            "Model",
		exif.DateTime:         "DateTime",
		exif.DateTimeOriginal: "DateTimeOriginal", // Дата съемки (основная)
		exif.FocalLength:      "FocalLength",
		exif.FNumber:          "Aperture",
		exif.ISOSpeedRatings:  "ISO",
		exif.ExposureTime:     "ShutterSpeed",
		exif.Flash:            "Flash",
		exif.WhiteBalance:     "WhiteBalance",
		exif.Orientation:      "Orientation",
		exif.XResolution:      "XResolution",
		exif.YResolution:      "YResolution",
		exif.Software:         "Software",
		exif.Artist:           "Artist",
		exif.Copyright:        "Copyright",
	}

	for fieldName, key := range tags {
		if tag, err := exifData.Get(fieldName); err == nil {
			result[key] = strings.Trim(tag.String(), "\"")
		}
	}

	// Добавляем дополнительные поля даты для лучшего извлечения
	// Приоритизируем DateTimeOriginal как основную дату съемки
	if result["DateTimeOriginal"] != "" {
		// Сохраняем DateTimeOriginal как основную дату съемки
		result["Date/Time Original"] = result["DateTimeOriginal"]
		result["DateTime Original"] = result["DateTimeOriginal"]
	}

	// Добавляем альтернативные названия полей для совместимости
	if result["Make"] != "" {
		result["Camera Make"] = result["Make"]
	}
	if result["Model"] != "" {
		result["Camera Model"] = result["Model"]
	}

	// Добавляем размеры изображения
	if tag, err := exifData.Get(exif.PixelXDimension); err == nil {
		result["Width"] = tag.String()
	}
	if tag, err := exifData.Get(exif.PixelYDimension); err == nil {
		result["Height"] = tag.String()
	}

	// Логируем извлеченные EXIF данные
	fileName := filepath.Base(imagePath)
	if p.logger != nil {
		p.logger.LogEXIFData(fileName, result)
	} else {
		log.Printf("Extracted EXIF data from %s: %d fields", imagePath, len(result))
	}

	return result, nil
}

// EncodeImageToBase64 кодирует изображение в base64
func (p *ImageProcessor) EncodeImageToBase64(imagePath string) (string, error) {
	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(imageBytes)
	return encoded, nil
}

// ProcessPhotoForAI подготавливает фото для отправки в AI
func (p *ImageProcessor) ProcessPhotoForAI(photo *models.Photo, thumbnailSize int) error {
	// Создаем миниатюру
	thumbnailPath, err := p.CreateThumbnail(photo.OriginalPath, thumbnailSize)
	if err != nil {
		return fmt.Errorf("failed to create thumbnail: %w", err)
	}
	photo.ThumbnailPath = thumbnailPath

	// Извлекаем EXIF данные
	exifData, err := p.ExtractExifData(photo.OriginalPath)
	if err != nil {
		log.Printf("Warning: failed to extract EXIF data from %s: %v", photo.OriginalPath, err)
		exifData = make(map[string]string)
	}
	photo.ExifData = exifData

	return nil
}

// WriteExifToImage записывает AI метаданные в EXIF изображения
func (p *ImageProcessor) WriteExifToImage(imagePath string, aiResult models.AIResult) error {
	log.Printf("Writing EXIF data to %s: title='%s', description='%s', keywords=%v, category='%s', quality=%d",
		imagePath, aiResult.Title, aiResult.Description, aiResult.Keywords, aiResult.Category, aiResult.Quality)

	// Используем exiftool для записи EXIF метаданных
	return p.writeExifWithTool(imagePath, aiResult)
}

// writeExifWithTool записывает EXIF данные используя внешний exiftool
func (p *ImageProcessor) writeExifWithTool(imagePath string, aiResult models.AIResult) error {
	// Ищем exiftool
	exifToolPath := p.findExifTool()
	if exifToolPath == "" {
		log.Printf("Warning: exiftool not found, skipping EXIF writing. Install with: brew install exiftool (macOS) or apt-get install libimage-exiftool-perl (Ubuntu)")
		return nil // Не считаем это критической ошибкой
	}

	// Сначала очищаем ВСЕ связанные метаданные в отдельном вызове
	clearArgs := []string{
		"-overwrite_original",
		"-codedcharacterset=utf8",
		"-Title=",
		"-XMP:Title=",
		"-IPTC:ObjectName=",
		"-Description=",
		"-XMP:Description=",
		"-IPTC:Caption-Abstract=",
		"-Keywords=",
		"-IPTC:Keywords=",
		"-XMP:Subject=",
		"-XMP:Category=",
		"-IPTC:Category=",
		"-Rating=",
		"-XMP:Rating=",
		imagePath,
	}

	// Выполняем очистку
	log.Printf("Clearing existing metadata: %s %s", exifToolPath, strings.Join(clearArgs, " "))
	clearCmd := exec.Command(exifToolPath, clearArgs...)
	if clearOutput, err := clearCmd.CombinedOutput(); err != nil {
		log.Printf("Warning: failed to clear existing metadata: %v, output: %s", err, string(clearOutput))
	} else {
		log.Printf("Successfully cleared existing metadata")
	}

	// Подготавливаем аргументы для записи новых данных
	args := []string{
		"-overwrite_original",     // Перезаписываем оригинальный файл
		"-codedcharacterset=utf8", // Используем UTF-8 кодировку
	}

	// Добавляем название
	if aiResult.Title != "" {
		// Экранируем специальные символы
		title := strings.ReplaceAll(aiResult.Title, "\"", "\\\"")
		args = append(args, fmt.Sprintf("-Title=%s", title))
		args = append(args, fmt.Sprintf("-XMP:Title=%s", title))
		args = append(args, fmt.Sprintf("-IPTC:ObjectName=%s", title))
	}

	// Добавляем описание
	if aiResult.Description != "" {
		description := strings.ReplaceAll(aiResult.Description, "\"", "\\\"")
		args = append(args, fmt.Sprintf("-Description=%s", description))
		args = append(args, fmt.Sprintf("-XMP:Description=%s", description))
		args = append(args, fmt.Sprintf("-IPTC:Caption-Abstract=%s", description))
	}

	// Добавляем ключевые слова
	if len(aiResult.Keywords) > 0 {
		// Используем -sep для правильной записи множественных ключевых слов
		keywordsStr := strings.Join(aiResult.Keywords, ", ")
		keywordsStr = strings.ReplaceAll(keywordsStr, "\"", "\\\"")

		// Добавляем флаг -sep для правильного разделения ключевых слов
		args = append(args, "-sep", ", ")

		// Записываем Keywords (автоматически дублируется в IPTC)
		args = append(args, fmt.Sprintf("-Keywords=%s", keywordsStr))

		// Записываем в XMP:Subject как отдельные элементы (после очистки)
		for _, keyword := range aiResult.Keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				keyword = strings.ReplaceAll(keyword, "\"", "\\\"")
				args = append(args, fmt.Sprintf("-XMP:Subject+=%s", keyword))
			}
		}
	}

	// Добавляем категорию если есть
	if aiResult.Category != "" {
		category := strings.ReplaceAll(aiResult.Category, "\"", "\\\"")
		args = append(args, fmt.Sprintf("-XMP:Category=%s", category))
		args = append(args, fmt.Sprintf("-IPTC:Category=%s", category))
	}

	// Добавляем качество (рейтинг) в EXIF
	if aiResult.Quality > 0 {
		args = append(args, fmt.Sprintf("-Rating=%d", aiResult.Quality))
		args = append(args, fmt.Sprintf("-XMP:Rating=%d", aiResult.Quality))
	}

	// Добавляем информацию о создателе/программе
	args = append(args, "-XMP:Creator=Stock Photo App")
	args = append(args, "-Software=Stock Photo App v1.0")

	// Добавляем тип контента
	if aiResult.ContentType != "" {
		args = append(args, fmt.Sprintf("-XMP:ContentType=%s", aiResult.ContentType))
	}

	// Добавляем путь к файлу
	args = append(args, imagePath)

	// Логируем команду для отладки
	log.Printf("Running exiftool command: %s %s", exifToolPath, strings.Join(args, " "))

	// Выполняем команду exiftool
	cmd := exec.Command(exifToolPath, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Error running exiftool: %v, output: %s", err, string(output))
		return fmt.Errorf("failed to write EXIF data: %w", err)
	}

	log.Printf("Successfully wrote EXIF data using exiftool to %s: %s", imagePath, strings.TrimSpace(string(output)))
	return nil
}

// CleanupTempFiles очищает временные файлы старше указанного времени
func (p *ImageProcessor) CleanupTempFiles(olderThan time.Duration) error {
	entries, err := os.ReadDir(p.tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	now := time.Now()
	var deletedCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > olderThan {
			filePath := filepath.Join(p.tempDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to delete temp file %s: %v", filePath, err)
			} else {
				deletedCount++
			}
		}
	}

	log.Printf("Cleaned up %d temporary files", deletedCount)
	return nil
}

// CheckExifToolAvailable проверяет, доступен ли exiftool
func (p *ImageProcessor) CheckExifToolAvailable() bool {
	return p.findExifTool() != ""
}

// findExifTool ищет exiftool в различных локациях
func (p *ImageProcessor) findExifTool() string {
	// Сначала пробуем стандартный поиск в PATH
	if exifPath, err := exec.LookPath("exiftool"); err == nil {
		return exifPath
	}

	// Возможные пути установки ExifTool на macOS
	possiblePaths := []string{
		"/usr/local/bin/exiftool",    // Homebrew
		"/opt/homebrew/bin/exiftool", // Apple Silicon Homebrew
		"/usr/bin/exiftool",          // System installation
		"/opt/local/bin/exiftool",    // MacPorts
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// VerifyExifData проверяет, что EXIF метаданные записались корректно
func (p *ImageProcessor) VerifyExifData(imagePath string, expectedResult models.AIResult) error {
	exifToolPath := p.findExifTool()
	if exifToolPath == "" {
		return fmt.Errorf("exiftool not available")
	}

	// Читаем EXIF данные с помощью exiftool
	cmd := exec.Command(exifToolPath, "-j", "-Title", "-Description", "-Keywords", "-XMP:Title", "-XMP:Description", "-XMP:Keywords", imagePath)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read EXIF data: %w", err)
	}

	log.Printf("EXIF verification for %s: %s", imagePath, strings.TrimSpace(string(output)))
	return nil
}
