package services

import (
	"encoding/json"
	"fmt"
	"log"
	"stock-photo-app/models"
	"stock-photo-app/uploaders"
	"sync"
	"time"
)

// UploadQueueManager управляет очередью загрузки файлов на стоки
type UploadQueueManager struct {
	uploaderManager      *uploaders.UploaderManager
	dbService            *DatabaseService
	activeUploads        map[string]*UploadJob
	uploadsMutex         sync.RWMutex
	maxConcurrentUploads int
	isProcessing         bool
	processingMutex      sync.Mutex
	jobChannel           chan *UploadJob
	stopChannel          chan bool
}

// UploadJob представляет задачу загрузки файла
type UploadJob struct {
	PhotoID      string               `json:"photoId"`
	BatchID      string               `json:"batchId"`
	FileName     string               `json:"fileName"`
	Photo        models.Photo         `json:"photo"`
	StockConfigs []models.StockConfig `json:"stockConfigs"`
	StartTime    time.Time            `json:"startTime"`
	Status       string               `json:"status"`   // "pending", "uploading", "completed", "failed"
	Progress     map[string]string    `json:"progress"` // stockID -> status
}

// NewUploadQueueManager создает новый менеджер очереди загрузки
func NewUploadQueueManager(uploaderManager *uploaders.UploaderManager, dbService *DatabaseService) *UploadQueueManager {
	return &UploadQueueManager{
		uploaderManager:      uploaderManager,
		dbService:            dbService,
		activeUploads:        make(map[string]*UploadJob),
		maxConcurrentUploads: 2, // Ограничение до 2 файлов параллельно
		jobChannel:           make(chan *UploadJob, 100),
		stopChannel:          make(chan bool),
	}
}

// StartUploadQueue запускает обработку очереди загрузки
func (q *UploadQueueManager) StartUploadQueue() {
	q.processingMutex.Lock()
	defer q.processingMutex.Unlock()

	if q.isProcessing {
		log.Printf("Upload queue is already running")
		return
	}

	q.isProcessing = true
	log.Printf("Starting upload queue with max %d concurrent uploads", q.maxConcurrentUploads)

	// Запускаем worker'ов для обработки загрузок
	for i := 0; i < q.maxConcurrentUploads; i++ {
		go q.uploadWorker(i)
	}
}

// StopUploadQueue останавливает обработку очереди загрузки
func (q *UploadQueueManager) StopUploadQueue() {
	q.processingMutex.Lock()
	defer q.processingMutex.Unlock()

	if !q.isProcessing {
		return
	}

	q.isProcessing = false

	// Останавливаем всех worker'ов
	for i := 0; i < q.maxConcurrentUploads; i++ {
		q.stopChannel <- true
	}

	log.Printf("Upload queue stopped")
}

// QueuePhotosForUpload добавляет фотографии в очередь загрузки
func (q *UploadQueueManager) QueuePhotosForUpload(batchID string, photoIDs []string) error {
	// Получаем информацию о батче для определения типа
	var batchType string
	err := q.dbService.db.QueryRow("SELECT type FROM batches WHERE id = ?", batchID).Scan(&batchType)
	if err != nil {
		return fmt.Errorf("failed to get batch type: %w", err)
	}

	// Получаем активные стоковые конфигурации для данного типа
	stockConfigs, err := q.dbService.GetActiveStockConfigs(batchType)
	if err != nil {
		return fmt.Errorf("failed to get stock configs: %w", err)
	}

	if len(stockConfigs) == 0 {
		return fmt.Errorf("no active stock configurations found for type: %s", batchType)
	}

	// Получаем данные фотографий
	for _, photoID := range photoIDs {
		photo, err := q.getPhotoData(photoID)
		if err != nil {
			log.Printf("Warning: failed to get photo data for %s: %v", photoID, err)
			continue
		}

		// Создаем задачу загрузки
		job := &UploadJob{
			PhotoID:      photoID,
			BatchID:      batchID,
			FileName:     photo.FileName,
			Photo:        photo,
			StockConfigs: stockConfigs,
			Status:       "pending",
			Progress:     make(map[string]string),
		}

		// Инициализируем прогресс для каждого стока
		for _, config := range stockConfigs {
			job.Progress[config.ID] = "pending"
		}

		// Обновляем статус фотографии как "queued_for_upload"
		q.dbService.UpdatePhotoUploadQueueStatus(photoID, "queued")

		// Добавляем в очередь
		select {
		case q.jobChannel <- job:
			log.Printf("Photo %s queued for upload to %d stocks", photo.FileName, len(stockConfigs))
		default:
			log.Printf("Upload queue is full, skipping photo %s", photo.FileName)
		}
	}

	return nil
}

// uploadWorker обрабатывает задачи загрузки
func (q *UploadQueueManager) uploadWorker(workerID int) {
	log.Printf("Upload worker %d started", workerID)

	for {
		select {
		case job := <-q.jobChannel:
			q.processUploadJob(workerID, job)
		case <-q.stopChannel:
			log.Printf("Upload worker %d stopped", workerID)
			return
		}
	}
}

// processUploadJob обрабатывает одну задачу загрузки
func (q *UploadQueueManager) processUploadJob(workerID int, job *UploadJob) {
	log.Printf("Worker %d: Starting upload of %s", workerID, job.FileName)

	// Добавляем в активные загрузки
	q.uploadsMutex.Lock()
	job.Status = "uploading"
	job.StartTime = time.Now()
	q.activeUploads[job.PhotoID] = job
	q.uploadsMutex.Unlock()

	defer func() {
		// Удаляем из активных загрузок
		q.uploadsMutex.Lock()
		delete(q.activeUploads, job.PhotoID)
		q.uploadsMutex.Unlock()
	}()

	// Обновляем статус фотографии
	q.dbService.UpdatePhotoUploadQueueStatus(job.PhotoID, "uploading")

	successCount := 0
	failedCount := 0

	// Загружаем на каждый сток последовательно
	for _, stockConfig := range job.StockConfigs {
		log.Printf("Worker %d: Uploading %s to %s", workerID, job.FileName, stockConfig.Name)

		// Обновляем прогресс
		q.uploadsMutex.Lock()
		job.Progress[stockConfig.ID] = "uploading"
		q.uploadsMutex.Unlock()

		// Обновляем статус в базе данных
		q.dbService.UpdatePhotoUploadStatus(job.PhotoID, stockConfig.ID, "uploading")

		// Логируем начало загрузки на конкретный сток
		q.dbService.LogEvent(job.BatchID, job.PhotoID, "stock_upload", "started",
			fmt.Sprintf("Начата загрузка %s на %s (worker %d)", job.FileName, stockConfig.Name, workerID), "", 0)

		// Выполняем загрузку
		result, err := q.uploaderManager.UploadPhoto(job.Photo, stockConfig)

		if err != nil || !result.Success {
			log.Printf("Worker %d: Failed to upload %s to %s: %v", workerID, job.FileName, stockConfig.Name, err)

			// Обновляем прогресс
			q.uploadsMutex.Lock()
			job.Progress[stockConfig.ID] = "failed"
			q.uploadsMutex.Unlock()

			// Обновляем статус в базе данных
			q.dbService.UpdatePhotoUploadStatus(job.PhotoID, stockConfig.ID, "failed")

			// Логируем ошибку
			errorMsg := result.Message
			if err != nil {
				errorMsg = err.Error()
			}
			q.dbService.LogEvent(job.BatchID, job.PhotoID, "stock_upload", "failed",
				fmt.Sprintf("Ошибка загрузки %s на %s", job.FileName, stockConfig.Name), errorMsg, 0)

			failedCount++
		} else {
			log.Printf("Worker %d: Successfully uploaded %s to %s", workerID, job.FileName, stockConfig.Name)

			// Обновляем прогресс
			q.uploadsMutex.Lock()
			job.Progress[stockConfig.ID] = "uploaded"
			q.uploadsMutex.Unlock()

			// Обновляем статус в базе данных
			q.dbService.UpdatePhotoUploadStatus(job.PhotoID, stockConfig.ID, "uploaded")

			// Логируем успех
			q.dbService.LogEvent(job.BatchID, job.PhotoID, "stock_upload", "success",
				fmt.Sprintf("Файл %s успешно загружен на %s", job.FileName, stockConfig.Name), "", 100)

			successCount++
		}
	}

	// Определяем финальный статус
	var finalStatus string
	if successCount > 0 && failedCount == 0 {
		finalStatus = "uploaded"
		job.Status = "completed"
	} else if successCount == 0 && failedCount > 0 {
		finalStatus = "upload_failed"
		job.Status = "failed"
	} else {
		finalStatus = "partially_uploaded"
		job.Status = "completed"
	}

	// Обновляем финальный статус фотографии
	q.dbService.UpdatePhotoUploadQueueStatus(job.PhotoID, finalStatus)

	// Логируем завершение
	q.dbService.LogEvent(job.BatchID, job.PhotoID, "stock_upload", "completed",
		fmt.Sprintf("Загрузка %s завершена. Успешно: %d, Ошибок: %d", job.FileName, successCount, failedCount), "", 100)

	log.Printf("Worker %d: Finished uploading %s. Success: %d, Failed: %d", workerID, job.FileName, successCount, failedCount)
}

// GetUploadStatus возвращает статус загрузки
func (q *UploadQueueManager) GetUploadStatus() map[string]interface{} {
	q.uploadsMutex.RLock()
	defer q.uploadsMutex.RUnlock()

	activeJobs := make([]map[string]interface{}, 0, len(q.activeUploads))
	for _, job := range q.activeUploads {
		activeJobs = append(activeJobs, map[string]interface{}{
			"photoId":   job.PhotoID,
			"fileName":  job.FileName,
			"status":    job.Status,
			"progress":  job.Progress,
			"startTime": job.StartTime,
		})
	}

	return map[string]interface{}{
		"isProcessing":  q.isProcessing,
		"activeUploads": len(q.activeUploads),
		"maxConcurrent": q.maxConcurrentUploads,
		"activeJobs":    activeJobs,
		"queueLength":   len(q.jobChannel),
	}
}

// getPhotoData получает данные фотографии из базы данных
func (q *UploadQueueManager) getPhotoData(photoID string) (models.Photo, error) {
	var photo models.Photo
	var exifJSON, aiResultsJSON, uploadStatusJSON string

	err := q.dbService.db.QueryRow(`
		SELECT id, batch_id, content_type, original_path, thumbnail_path, file_name, file_size,
		       exif_data, ai_results, upload_status, status, created_at
		FROM photos 
		WHERE id = ?`, photoID).Scan(
		&photo.ID, &photo.BatchID, &photo.ContentType, &photo.OriginalPath,
		&photo.ThumbnailPath, &photo.FileName, &photo.FileSize,
		&exifJSON, &aiResultsJSON, &uploadStatusJSON,
		&photo.Status, &photo.CreatedAt)

	if err != nil {
		return photo, fmt.Errorf("failed to get photo data: %w", err)
	}

	// Десериализация JSON полей если они не пустые
	if exifJSON != "" {
		json.Unmarshal([]byte(exifJSON), &photo.ExifData)
	}
	if aiResultsJSON != "" {
		var aiResult models.AIResult
		if json.Unmarshal([]byte(aiResultsJSON), &aiResult) == nil {
			photo.AIResult = &aiResult
		}
	}
	if uploadStatusJSON != "" {
		json.Unmarshal([]byte(uploadStatusJSON), &photo.UploadStatus)
	}

	// Инициализируем карты если они nil
	if photo.ExifData == nil {
		photo.ExifData = make(map[string]string)
	}
	if photo.UploadStatus == nil {
		photo.UploadStatus = make(map[string]string)
	}

	return photo, nil
}

// GetStatus возвращает текущий статус очереди загрузки
func (uqm *UploadQueueManager) GetStatus() map[string]interface{} {
	uqm.uploadsMutex.RLock()
	defer uqm.uploadsMutex.RUnlock()

	activeJobs := make([]map[string]interface{}, 0)
	for _, job := range uqm.activeUploads {
		jobInfo := map[string]interface{}{
			"photoId":  job.PhotoID,
			"fileName": job.FileName,
			"status":   job.Status,
			"progress": job.Progress,
		}
		activeJobs = append(activeJobs, jobInfo)
	}

	return map[string]interface{}{
		"isProcessing":  uqm.isProcessing,
		"activeUploads": len(uqm.activeUploads),
		"queueLength":   len(uqm.jobChannel),
		"maxConcurrent": uqm.maxConcurrentUploads,
		"activeJobs":    activeJobs,
	}
}

// Stop останавливает очередь загрузки
func (uqm *UploadQueueManager) Stop() {
	uqm.processingMutex.Lock()
	defer uqm.processingMutex.Unlock()

	if !uqm.isProcessing {
		return
	}

	log.Println("Stopping upload queue...")

	// Отправляем сигнал остановки
	close(uqm.stopChannel)

	// Ждем завершения активных загрузок
	uqm.uploadsMutex.Lock()
	for len(uqm.activeUploads) > 0 {
		uqm.uploadsMutex.Unlock()
		time.Sleep(100 * time.Millisecond)
		uqm.uploadsMutex.Lock()
	}
	uqm.uploadsMutex.Unlock()

	uqm.isProcessing = false

	// Создаем новый канал остановки для следующего запуска
	uqm.stopChannel = make(chan bool)

	log.Println("Upload queue stopped")
}
