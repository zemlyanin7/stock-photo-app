package uploaders

import (
	"fmt"
	"log"
	"os"
	"stock-photo-app/models"
	"strings"
	"time"

	"crypto/tls"

	"github.com/jlaffaye/ftp"
)

// FTPUploader реализует загрузку через FTP
type FTPUploader struct {
	*BaseUploader
	dbService DatabaseService
}

// DatabaseService интерфейс для логирования
type DatabaseService interface {
	LogEvent(batchID, photoID, eventType, status, message, details string, progress int) error
}

// NewFTPUploader создает новый FTP загрузчик
func NewFTPUploader(dbService DatabaseService) *FTPUploader {
	info := models.UploaderInfo{
		Name:        "FTP Uploader",
		Version:     "1.0.0",
		Description: "Загрузка файлов через FTP/FTPS протокол с поддержкой шифрования",
		Author:      "Stock Photo App",
		Type:        "ftp",
	}

	return &FTPUploader{
		BaseUploader: NewBaseUploader(info),
		dbService:    dbService,
	}
}

// Upload загружает фото через FTP с retry логикой
func (u *FTPUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Логируем попытку подключения
		if attempt > 1 {
			log.Printf("FTP: Retry attempt %d/%d for %s", attempt, maxRetries, photo.FileName)
			u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "progress",
				fmt.Sprintf("Попытка загрузки %d/%d для %s", attempt, maxRetries, photo.FileName), "", (attempt-1)*25)

			// Экспоненциальная задержка между попытками
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		// Подключаемся к FTP серверу
		conn, err := u.connect(config)
		if err != nil {
			lastErr = err
			if attempt == maxRetries {
				// Логируем ошибку подключения после всех попыток
				u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "failed",
					fmt.Sprintf("Ошибка подключения к FTP %s после %d попыток", config.Connection.Host, maxRetries), err.Error(), 0)
				return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка подключения к FTP %s после %d попыток: %v", config.Connection.Host, maxRetries, err), false), err
			}
			log.Printf("FTP: Connection attempt %d failed: %v", attempt, err)
			continue
		}

		// Попытка загрузки
		result, err := u.uploadFile(conn, photo, config)
		conn.Quit()

		if err == nil {
			return result, nil
		}

		lastErr = err
		if attempt == maxRetries {
			// Логируем финальную ошибку
			u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "failed",
				fmt.Sprintf("Ошибка загрузки файла %s после %d попыток", photo.FileName, maxRetries), err.Error(), 0)
			return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка загрузки после %d попыток: %v", maxRetries, err), false), err
		}

		log.Printf("FTP: Upload attempt %d failed: %v", attempt, err)
	}

	return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Все попытки загрузки неудачны: %v", lastErr), false), lastErr
}

// uploadFile выполняет загрузку файла через установленное соединение
func (u *FTPUploader) uploadFile(conn *ftp.ServerConn, photo models.Photo, config models.StockConfig) (models.UploadResult, error) {

	// Логируем начало загрузки
	u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "started",
		fmt.Sprintf("Начата загрузка фото %s на %s", photo.FileName, config.Name), "", 0)

	// Переходим в нужную папку
	if config.Connection.Path != "" && config.Connection.Path != "/" {
		log.Printf("FTP: Changing directory to %s", config.Connection.Path)
		err := conn.ChangeDir(config.Connection.Path)
		if err != nil {
			// Логируем ошибку смены директории
			u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "failed",
				"Ошибка смены директории FTP", err.Error(), 0)
			return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка смены директории: %v", err), false), err
		}
	}

	// Открываем файл
	file, err := os.Open(photo.OriginalPath)
	if err != nil {
		// Логируем ошибку открытия файла
		u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "failed",
			"Ошибка открытия файла", err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка открытия файла: %v", err), false), err
	}
	defer file.Close()

	// Загружаем файл
	log.Printf("FTP: Uploading file %s", photo.FileName)
	err = conn.Stor(photo.FileName, file)
	if err != nil {
		// Логируем ошибку загрузки
		u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "failed",
			fmt.Sprintf("Ошибка загрузки файла %s", photo.FileName), err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка загрузки: %v", err), false), err
	}

	// Логируем успешную загрузку
	successMessage := fmt.Sprintf("Файл %s успешно загружен на %s", photo.FileName, config.Name)
	u.dbService.LogEvent(photo.BatchID, photo.ID, "ftp_upload", "success",
		successMessage, "", 100)

	log.Printf("FTP: File %s uploaded successfully", photo.FileName)
	return u.CreateUploadResult(photo.ID, config.ID, successMessage, true), nil
}

// TestConnection тестирует подключение к FTP серверу
func (u *FTPUploader) TestConnection(config models.StockConfig) error {
	conn, err := u.connect(config)
	if err != nil {
		return err
	}
	defer conn.Quit()

	// Проверяем доступность удаленной папки
	if config.Connection.Path != "" && config.Connection.Path != "/" {
		err = conn.ChangeDir(config.Connection.Path)
		if err != nil {
			return fmt.Errorf("не удается получить доступ к папке %s: %v", config.Connection.Path, err)
		}
	}

	return nil
}

// ValidateConfig проверяет конфигурацию FTP
func (u *FTPUploader) ValidateConfig(config models.StockConfig) error {
	requiredFields := []string{"host", "username", "password", "port"}
	return u.ValidateRequiredFields(config, requiredFields)
}

// connect подключается к FTP серверу
func (u *FTPUploader) connect(config models.StockConfig) (*ftp.ServerConn, error) {
	// Устанавливаем таймаут
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second // Увеличиваем с 30 до 120 секунд
	}

	// Определяем режим шифрования
	encryption := config.Connection.Encryption
	if encryption == "" {
		encryption = "none"
	}

	log.Printf("FTP: [DEBUG] Stock Config ID: %s, Name: %s", config.ID, config.Name)
	log.Printf("FTP: [DEBUG] Connection config: %+v", config.Connection)
	log.Printf("FTP: Connecting to %s:%d with encryption=%s, timeout=%v",
		config.Connection.Host, config.Connection.Port, encryption, timeout)

	// Настройка TLS конфигурации
	var tlsConfig *tls.Config
	if encryption != "none" {
		tlsConfig = &tls.Config{
			ServerName:         config.Connection.Host,
			InsecureSkipVerify: !config.Connection.VerifyCert,
			MinVersion:         tls.VersionTLS12, // Минимум TLS 1.2
			MaxVersion:         tls.VersionTLS13, // Максимум TLS 1.3
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			},
		}
		log.Printf("FTP: TLS Config - ServerName: %s, VerifyCert: %t",
			tlsConfig.ServerName, !tlsConfig.InsecureSkipVerify)
	}

	var conn *ftp.ServerConn
	var err error
	addr := fmt.Sprintf("%s:%d", config.Connection.Host, config.Connection.Port)

	// Выбираем режим подключения в зависимости от типа шифрования
	switch encryption {
	case "none":
		// Обычный FTP без шифрования
		dialOptions := []ftp.DialOption{ftp.DialWithTimeout(timeout)}
		if config.Connection.Passive {
			dialOptions = append(dialOptions, ftp.DialWithDisabledEPSV(true))
		}
		conn, err = ftp.Dial(addr, dialOptions...)

	case "auto":
		// Пробуем FTPS, если не получается - обычный FTP
		dialOptionsSecure := []ftp.DialOption{
			ftp.DialWithTimeout(timeout),
			ftp.DialWithExplicitTLS(tlsConfig),
		}
		if config.Connection.Passive {
			dialOptionsSecure = append(dialOptionsSecure, ftp.DialWithDisabledEPSV(true))
		}

		conn, err = ftp.Dial(addr, dialOptionsSecure...)
		if err != nil {
			// Если FTPS не работает, пробуем обычный FTP
			dialOptionsPlain := []ftp.DialOption{ftp.DialWithTimeout(timeout)}
			if config.Connection.Passive {
				dialOptionsPlain = append(dialOptionsPlain, ftp.DialWithDisabledEPSV(true))
			}
			conn, err = ftp.Dial(addr, dialOptionsPlain...)
		}

	case "explicit":
		// Явный FTPS (FTPS explicit) - подключение по обычному порту с последующим переходом на TLS
		dialOptions := []ftp.DialOption{
			ftp.DialWithTimeout(timeout),
			ftp.DialWithExplicitTLS(tlsConfig),
		}

		// Добавляем дополнительные опции для пассивного режима
		if config.Connection.Passive {
			dialOptions = append(dialOptions, ftp.DialWithDisabledEPSV(true))
		}

		conn, err = ftp.Dial(addr, dialOptions...)

	case "implicit":
		// Неявный FTPS (FTPS implicit) - подключение сразу через TLS, обычно порт 990
		if config.Connection.Port == 21 {
			// Автоматически меняем порт для implicit FTPS
			addr = fmt.Sprintf("%s:990", config.Connection.Host)
		}
		// Используем DialWithTLS для implicit FTPS (подключение сразу через TLS)
		conn, err = ftp.Dial(addr, ftp.DialWithTimeout(timeout), ftp.DialWithTLS(tlsConfig))

	default:
		return nil, fmt.Errorf("неподдерживаемый тип шифрования: %s", encryption)
	}

	if err != nil {
		// Детальная диагностика ошибки
		errMsg := fmt.Sprintf("не удается подключиться к %s (шифрование: %s)", addr, encryption)
		if encryption == "none" && config.Connection.Port != 21 {
			errMsg += fmt.Sprintf(" - проверьте порт %d для обычного FTP", config.Connection.Port)
		} else if encryption == "implicit" && config.Connection.Port == 21 {
			errMsg += " - для implicit FTPS обычно используется порт 990"
		}

		// Проверяем тип ошибки
		if strings.Contains(err.Error(), "timeout") {
			errMsg += " - таймаут подключения (проверьте сеть/файервол)"
		} else if strings.Contains(err.Error(), "connection refused") {
			errMsg += " - отказ подключения (проверьте хост/порт)"
		} else if strings.Contains(err.Error(), "no such host") {
			errMsg += " - хост не найден (проверьте адрес сервера)"
		}

		return nil, fmt.Errorf("%s: %v", errMsg, err)
	}

	// Авторизуемся
	log.Printf("FTP: Attempting login for user %s", config.Connection.Username)
	err = conn.Login(config.Connection.Username, config.Connection.Password)
	if err != nil {
		conn.Quit()
		return nil, fmt.Errorf("ошибка авторизации: %v", err)
	}

	log.Printf("FTP: Successfully connected and logged in to %s", config.Connection.Host)

	// Устанавливаем пассивный режим если указан (по умолчанию используется пассивный)
	if !config.Connection.Passive {
		log.Printf("FTP: Warning - Active mode not explicitly supported by library")
	}

	// Тестируем соединение простой командой
	_, err = conn.CurrentDir()
	if err != nil {
		conn.Quit()
		return nil, fmt.Errorf("не удается получить текущую директорию (проблема с режимом FTP): %v", err)
	}

	return conn, nil
}
