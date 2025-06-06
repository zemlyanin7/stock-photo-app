package uploaders

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"stock-photo-app/models"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPUploader реализует загрузку через SFTP
type SFTPUploader struct {
	*BaseUploader
	dbService DatabaseService
}

// NewSFTPUploader создает новый SFTP загрузчик
func NewSFTPUploader(dbService DatabaseService) *SFTPUploader {
	info := models.UploaderInfo{
		Name:        "SFTP Uploader",
		Version:     "1.0.0",
		Description: "Загрузка файлов через SFTP протокол",
		Author:      "Stock Photo App",
		Type:        "sftp",
	}

	return &SFTPUploader{
		BaseUploader: NewBaseUploader(info),
		dbService:    dbService,
	}
}

// Upload загружает фото через SFTP
func (u *SFTPUploader) Upload(photo models.Photo, config models.StockConfig) (models.UploadResult, error) {
	// Логируем начало загрузки
	u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "started",
		fmt.Sprintf("Начата загрузка фото %s на %s", photo.FileName, config.Name), "", 0)

	// Подключаемся к SFTP серверу
	sftpClient, sshClient, err := u.connect(config)
	if err != nil {
		u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "failed",
			fmt.Sprintf("Ошибка подключения к SFTP %s", config.Connection.Host), err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка подключения к SFTP: %v", err), false), err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	// Открываем локальный файл
	localFile, err := os.Open(photo.OriginalPath)
	if err != nil {
		u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "failed",
			"Ошибка открытия файла", err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка открытия файла: %v", err), false), err
	}
	defer localFile.Close()

	// Определяем удаленный путь
	remotePath := u.PrepareRemotePath(config, photo)

	// Создаем удаленную папку если не существует
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "/" {
		err = sftpClient.MkdirAll(remoteDir)
		if err != nil {
			log.Printf("Warning: Failed to create remote directory %s: %v", remoteDir, err)
		}
	}

	log.Printf("Uploading %s to %s", photo.FileName, remotePath)

	// Создаем удаленный файл
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "failed",
			fmt.Sprintf("Ошибка создания удаленного файла %s", remotePath), err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка создания удаленного файла: %v", err), false), err
	}
	defer remoteFile.Close()

	// Копируем содержимое файла
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "failed",
			fmt.Sprintf("Ошибка загрузки файла %s", photo.FileName), err.Error(), 0)
		return u.CreateUploadResult(photo.ID, config.ID, fmt.Sprintf("Ошибка загрузки файла: %v", err), false), err
	}

	// Логируем успешную загрузку
	successMessage := fmt.Sprintf("Файл %s успешно загружен на %s", photo.FileName, config.Name)
	u.dbService.LogEvent(photo.BatchID, photo.ID, "sftp_upload", "success", successMessage, "", 100)

	return u.CreateUploadResult(photo.ID, config.ID, successMessage, true), nil
}

// TestConnection тестирует подключение к SFTP серверу
func (u *SFTPUploader) TestConnection(config models.StockConfig) error {
	sftpClient, sshClient, err := u.connect(config)
	if err != nil {
		return err
	}
	defer sftpClient.Close()
	defer sshClient.Close()

	// Проверяем доступность удаленной папки
	if config.Connection.Path != "" && config.Connection.Path != "/" {
		_, err = sftpClient.Stat(config.Connection.Path)
		if err != nil {
			// Пытаемся создать папку
			err = sftpClient.MkdirAll(config.Connection.Path)
			if err != nil {
				return fmt.Errorf("не удается получить доступ к папке %s: %v", config.Connection.Path, err)
			}
		}
	}

	return nil
}

// ValidateConfig проверяет конфигурацию SFTP
func (u *SFTPUploader) ValidateConfig(config models.StockConfig) error {
	requiredFields := []string{"host", "username", "password", "port"}
	return u.ValidateRequiredFields(config, requiredFields)
}

// connect подключается к SFTP серверу
func (u *SFTPUploader) connect(config models.StockConfig) (*sftp.Client, *ssh.Client, error) {
	// Устанавливаем таймаут
	timeout := time.Duration(config.Connection.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Настройка SSH клиента
	sshConfig := &ssh.ClientConfig{
		User: config.Connection.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Connection.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // В продакшене следует использовать проверку ключей
		Timeout:         timeout,
	}

	// Подключаемся к SSH серверу
	addr := fmt.Sprintf("%s:%d", config.Connection.Host, config.Connection.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("не удается подключиться к %s: %v", addr, err)
	}

	// Создаем SFTP клиент
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("не удается создать SFTP клиент: %v", err)
	}

	return sftpClient, sshClient, nil
}
