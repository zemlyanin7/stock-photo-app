package services

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger сервис для записи логов в файлы
type Logger struct {
	logDir      string
	debugFile   *os.File
	errorFile   *os.File
	aiFile      *os.File
	exifFile    *os.File
	debugLogger *log.Logger
	errorLogger *log.Logger
	aiLogger    *log.Logger
	exifLogger  *log.Logger
}

// NewLogger создает новый логгер с записью в файлы
func NewLogger(logDir string) (*Logger, error) {
	// Создаем папку logs если не существует
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	logger := &Logger{
		logDir: logDir,
	}

	// Инициализируем файлы логов
	if err := logger.initLogFiles(); err != nil {
		return nil, fmt.Errorf("failed to initialize log files: %w", err)
	}

	return logger, nil
}

// initLogFiles создает и открывает файлы логов
func (l *Logger) initLogFiles() error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// Создаем файлы для разных типов логов
	var err error

	// Debug лог (общие отладочные сообщения)
	debugPath := filepath.Join(l.logDir, fmt.Sprintf("debug_%s.log", dateStr))
	l.debugFile, err = os.OpenFile(debugPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open debug log file: %w", err)
	}

	// Error лог (ошибки)
	errorPath := filepath.Join(l.logDir, fmt.Sprintf("error_%s.log", dateStr))
	l.errorFile, err = os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %w", err)
	}

	// AI лог (специально для AI промптов и ответов)
	aiPath := filepath.Join(l.logDir, fmt.Sprintf("ai_%s.log", dateStr))
	l.aiFile, err = os.OpenFile(aiPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open ai log file: %w", err)
	}

	// EXIF лог (для EXIF данных)
	exifPath := filepath.Join(l.logDir, fmt.Sprintf("exif_%s.log", dateStr))
	l.exifFile, err = os.OpenFile(exifPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open exif log file: %w", err)
	}

	// Создаем логгеры с записью в файлы И в stdout
	l.debugLogger = log.New(io.MultiWriter(l.debugFile, os.Stdout), "[DEBUG] ", log.LstdFlags|log.Lshortfile)
	l.errorLogger = log.New(io.MultiWriter(l.errorFile, os.Stderr), "[ERROR] ", log.LstdFlags|log.Lshortfile)
	l.aiLogger = log.New(io.MultiWriter(l.aiFile, os.Stdout), "[AI] ", log.LstdFlags)
	l.exifLogger = log.New(io.MultiWriter(l.exifFile, os.Stdout), "[EXIF] ", log.LstdFlags)

	return nil
}

// LogDebug записывает отладочное сообщение
func (l *Logger) LogDebug(format string, args ...interface{}) {
	l.debugLogger.Printf(format, args...)
}

// LogError записывает сообщение об ошибке
func (l *Logger) LogError(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// LogAI записывает AI-related логи (промпты, ответы)
func (l *Logger) LogAI(format string, args ...interface{}) {
	l.aiLogger.Printf(format, args...)
}

// LogEXIF записывает EXIF-related логи
func (l *Logger) LogEXIF(format string, args ...interface{}) {
	l.exifLogger.Printf(format, args...)
}

// LogAIPrompt записывает полный промпт для AI с разделителями
func (l *Logger) LogAIPrompt(photoFileName string, prompt string, description string) {
	l.aiLogger.Println("================================")
	l.aiLogger.Printf("AI PROMPT for photo: %s", photoFileName)
	l.aiLogger.Println("================================")
	if description != "" {
		l.aiLogger.Printf("Batch Description: %s", description)
		l.aiLogger.Println("--------------------------------")
	}
	l.aiLogger.Printf("Full Prompt:\n%s", prompt)
	l.aiLogger.Println("================================")
}

// LogAIResponse записывает ответ от AI
func (l *Logger) LogAIResponse(photoFileName string, response string) {
	l.aiLogger.Println("--------------------------------")
	l.aiLogger.Printf("AI RESPONSE for photo: %s", photoFileName)
	l.aiLogger.Println("--------------------------------")
	l.aiLogger.Printf("Response:\n%s", response)
	l.aiLogger.Println("--------------------------------")
}

// LogEXIFData записывает извлеченные EXIF данные
func (l *Logger) LogEXIFData(photoFileName string, exifData map[string]string) {
	l.exifLogger.Println("================================")
	l.exifLogger.Printf("EXIF DATA for photo: %s", photoFileName)
	l.exifLogger.Println("================================")
	l.exifLogger.Printf("Total EXIF fields: %d", len(exifData))

	// Группируем по типам полей
	dateFields := make(map[string]string)
	cameraFields := make(map[string]string)
	locationFields := make(map[string]string)
	otherFields := make(map[string]string)

	for key, value := range exifData {
		if value == "" {
			continue
		}

		// Классифицируем поля
		switch key {
		case "DateTimeOriginal", "DateTime Original", "Date/Time Original",
			"DateTime", "Date/Time", "Create Date", "Date Created":
			dateFields[key] = value
		case "Make", "Model", "Camera Make", "Camera Model", "Lens Model", "Software":
			cameraFields[key] = value
		case "GPS Latitude", "GPS Longitude", "GPS Position", "Location", "City", "Country":
			locationFields[key] = value
		default:
			otherFields[key] = value
		}
	}

	// Выводим по группам
	if len(dateFields) > 0 {
		l.exifLogger.Println("Date Fields:")
		for key, value := range dateFields {
			l.exifLogger.Printf("  %s: %s", key, value)
		}
	}

	if len(cameraFields) > 0 {
		l.exifLogger.Println("Camera Fields:")
		for key, value := range cameraFields {
			l.exifLogger.Printf("  %s: %s", key, value)
		}
	}

	if len(locationFields) > 0 {
		l.exifLogger.Println("Location Fields:")
		for key, value := range locationFields {
			l.exifLogger.Printf("  %s: %s", key, value)
		}
	}

	if len(otherFields) > 0 {
		l.exifLogger.Println("Other Fields:")
		for key, value := range otherFields {
			l.exifLogger.Printf("  %s: %s", key, value)
		}
	}

	l.exifLogger.Println("================================")
}

// Close закрывает все файлы логов
func (l *Logger) Close() error {
	var lastErr error

	if l.debugFile != nil {
		if err := l.debugFile.Close(); err != nil {
			lastErr = err
		}
	}

	if l.errorFile != nil {
		if err := l.errorFile.Close(); err != nil {
			lastErr = err
		}
	}

	if l.aiFile != nil {
		if err := l.aiFile.Close(); err != nil {
			lastErr = err
		}
	}

	if l.exifFile != nil {
		if err := l.exifFile.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// CleanOldLogs удаляет старые лог файлы (старше указанного количества дней)
func (l *Logger) CleanOldLogs(daysToKeep int) error {
	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("failed to read logs directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Проверяем только наши лог файлы
		if !isLogFile(entry.Name()) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			logPath := filepath.Join(l.logDir, entry.Name())
			if err := os.Remove(logPath); err != nil {
				l.LogError("Failed to remove old log file %s: %v", logPath, err)
			} else {
				l.LogDebug("Removed old log file: %s", logPath)
			}
		}
	}

	return nil
}

// isLogFile проверяет является ли файл нашим лог файлом
func isLogFile(filename string) bool {
	prefixes := []string{"debug_", "error_", "ai_", "exif_"}
	for _, prefix := range prefixes {
		if len(filename) > len(prefix) && filename[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
