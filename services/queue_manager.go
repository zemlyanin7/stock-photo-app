package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"stock-photo-app/models"
	"sync"
	"time"
)

// QueueManager управляет очередью обработки фотографий
type QueueManager struct {
	db                *sql.DB
	dbService         *DatabaseService
	aiService         *AIService
	imageProcessor    *ImageProcessor
	exifProcessor     *EXIFProcessor
	activeJobs        map[string]*ProcessingJob
	jobsMutex         sync.RWMutex
	maxConcurrentJobs int
	isProcessing      bool
	processingMutex   sync.Mutex
}

// ProcessingJob представляет активную задачу обработки
type ProcessingJob struct {
	BatchID       string
	CurrentPhoto  string
	CurrentStep   string
	Progress      int
	Status        string
	Error         string
	StartTime     time.Time
	PhotoProgress map[string]models.PhotoProcessInfo // фото ID -> прогресс
}

type photoResult struct {
	photo models.Photo
	err   error
}

// NewQueueManager создает новый менеджер очередей
func NewQueueManager(db *sql.DB, dbService *DatabaseService, aiService *AIService, imageProcessor *ImageProcessor) *QueueManager {
	return &QueueManager{
		db:             db,
		dbService:      dbService,
		aiService:      aiService,
		imageProcessor: imageProcessor,
		exifProcessor:  NewEXIFProcessor(),
		activeJobs:     make(map[string]*ProcessingJob),
		jobsMutex:      sync.RWMutex{},
	}
}

// StartProcessing запускает обработку очереди
func (q *QueueManager) StartProcessing(settings models.AppSettings) error {
	q.processingMutex.Lock()
	defer q.processingMutex.Unlock()

	if q.isProcessing {
		return fmt.Errorf("processing already running")
	}

	q.maxConcurrentJobs = settings.MaxConcurrentJobs
	if q.maxConcurrentJobs <= 0 {
		q.maxConcurrentJobs = 3
	}

	q.isProcessing = true
	go q.processQueue(settings)

	log.Printf("Queue processing started with %d concurrent jobs", q.maxConcurrentJobs)
	return nil
}

// StopProcessing останавливает обработку очереди
func (q *QueueManager) StopProcessing() {
	q.processingMutex.Lock()
	defer q.processingMutex.Unlock()

	q.isProcessing = false
	log.Println("Queue processing stopped")
}

// AddBatch добавляет батч в очередь обработки
func (q *QueueManager) AddBatch(batch models.PhotoBatch) error {
	// Обновляем статус батча на "queued"
	batch.Status = "queued"

	log.Printf("DEBUG: Adding batch %s with %d photos to queue", batch.ID, len(batch.Photos))

	// Устанавливаем ContentType для всех фотографий
	for i := range batch.Photos {
		batch.Photos[i].ContentType = batch.Type
		batch.Photos[i].Status = "pending"
		log.Printf("DEBUG: Photo %d: ID=%s, FileName=%s, ContentType=%s", i+1, batch.Photos[i].ID, batch.Photos[i].FileName, batch.Photos[i].ContentType)
	}

	err := q.dbService.SaveBatch(batch)
	if err != nil {
		return fmt.Errorf("failed to save batch to queue: %w", err)
	}

	// Проверяем сколько фотографий реально сохранилось в базе
	savedPhotos, err := q.getPhotosForBatch(batch.ID)
	if err != nil {
		log.Printf("WARNING: Failed to verify saved photos count: %v", err)
	} else {
		log.Printf("DEBUG: Verified %d photos saved in database for batch %s", len(savedPhotos), batch.ID)
		for i, photo := range savedPhotos {
			log.Printf("DEBUG: Saved photo %d: ID=%s, FileName=%s, ContentType=%s, Status=%s", i+1, photo.ID, photo.FileName, photo.ContentType, photo.Status)
		}
	}

	log.Printf("Batch %s added to queue with %d photos", batch.ID, len(batch.Photos))
	return nil
}

// processQueue основной цикл обработки очереди
func (q *QueueManager) processQueue(settings models.AppSettings) {
	for q.isProcessing {
		// Получаем следующий батч для обработки
		batch, err := q.getNextBatch()
		if err != nil {
			log.Printf("Error getting next batch: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		if batch == nil {
			// Нет батчей для обработки, ждем
			time.Sleep(5 * time.Second)
			continue
		}

		// Обрабатываем батч
		err = q.processBatch(*batch, settings)
		if err != nil {
			log.Printf("Error processing batch %s: %v", batch.ID, err)
			q.updateBatchStatus(batch.ID, "failed", err.Error())
		}
	}
}

// getNextBatch получает следующий батч для обработки
func (q *QueueManager) getNextBatch() (*models.PhotoBatch, error) {
	// Проверяем количество активных задач
	q.jobsMutex.RLock()
	activeCount := len(q.activeJobs)
	q.jobsMutex.RUnlock()

	if activeCount >= q.maxConcurrentJobs {
		return nil, nil // Достигнут лимит одновременных задач
	}

	// Ищем батч в статусе "queued"
	rows, err := q.db.Query(`
		SELECT id, type, description, folder_path, status, created_at, updated_at
		FROM batches 
		WHERE status = 'queued' 
		ORDER BY created_at ASC 
		LIMIT 1`)
	if err != nil {
		return nil, fmt.Errorf("failed to query queued batches: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil // Нет батчей в очереди
	}

	var batch models.PhotoBatch
	err = rows.Scan(&batch.ID, &batch.Type, &batch.Description,
		&batch.FolderPath, &batch.Status, &batch.CreatedAt, &batch.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan batch: %w", err)
	}

	// Загружаем фотографии для батча
	photos, err := q.getPhotosForBatch(batch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get photos for batch: %w", err)
	}
	batch.Photos = photos

	return &batch, nil
}

// processBatch обрабатывает один батч
func (q *QueueManager) processBatch(batch models.PhotoBatch, settings models.AppSettings) error {
	log.Printf("Starting processing batch %s with %d photos", batch.ID, len(batch.Photos))
	log.Printf("DEBUG: Photos in batch %s:", batch.ID)
	for i, photo := range batch.Photos {
		log.Printf("DEBUG: Photo %d: ID=%s, FileName=%s, ContentType=%s, Status=%s", i+1, photo.ID, photo.FileName, photo.ContentType, photo.Status)
	}

	// Создаем задачу
	job := &ProcessingJob{
		BatchID:       batch.ID,
		Progress:      0,
		Status:        "processing",
		CurrentStep:   "initialization",
		StartTime:     time.Now(),
		PhotoProgress: make(map[string]models.PhotoProcessInfo),
	}

	// Инициализируем прогресс для каждой фотографии
	for _, photo := range batch.Photos {
		job.PhotoProgress[photo.ID] = models.PhotoProcessInfo{
			ID:       photo.ID,
			FileName: photo.FileName,
			Status:   "pending",
			Progress: 0,
			Step:     "waiting",
		}
	}

	q.jobsMutex.Lock()
	q.activeJobs[batch.ID] = job
	q.jobsMutex.Unlock()

	defer func() {
		q.jobsMutex.Lock()
		delete(q.activeJobs, batch.ID)
		q.jobsMutex.Unlock()
	}()

	// Обновляем статус батча
	q.updateBatchStatus(batch.ID, "processing", "")

	// Устанавливаем начальный этап
	job.CurrentStep = "ai_processing"

	// Логируем начало обработки батча
	q.dbService.LogEvent(batch.ID, "", "batch_start", "started",
		fmt.Sprintf("Начата обработка батча с %d фотографиями", len(batch.Photos)), "", 0)

	processedCount := 0

	// Параллельная обработка фотографий с worker pool
	numWorkers := settings.MaxConcurrentJobs
	if numWorkers <= 0 {
		numWorkers = 3 // значение по умолчанию
	}

	log.Printf("Starting parallel processing with %d workers for %d photos", numWorkers, len(batch.Photos))

	photoChannel := make(chan models.Photo, len(batch.Photos))
	resultChannel := make(chan photoResult, len(batch.Photos))

	// Заполняем канал фотографиями
	for _, photo := range batch.Photos {
		photoChannel <- photo
	}
	close(photoChannel)

	// Запускаем worker'ов
	for i := 0; i < numWorkers; i++ {
		go q.photoWorker(i, photoChannel, resultChannel, batch.ID, batch.Description, batch.Type, settings, job)
	}

	// Собираем результаты
	for i := 0; i < len(batch.Photos); i++ {
		// Проверяем не остановлена ли обработка
		q.processingMutex.Lock()
		if !q.isProcessing {
			q.processingMutex.Unlock()
			log.Printf("Processing stopped, batch %s interrupted", batch.ID)
			q.dbService.LogEvent(batch.ID, "", "batch_interrupted", "failed",
				"Обработка прервана пользователем", "", processedCount*100/len(batch.Photos))
			return fmt.Errorf("processing was stopped")
		}
		q.processingMutex.Unlock()

		result := <-resultChannel
		processedCount++

		if result.err != nil {
			log.Printf("Failed to process photo %s: %v", result.photo.FileName, result.err)

			// Логируем ошибку
			q.dbService.LogEvent(batch.ID, result.photo.ID, "ai_processing", "failed",
				fmt.Sprintf("Ошибка AI обработки фото %s", result.photo.FileName), result.err.Error(), 0)

			q.updatePhotoStatus(result.photo.ID, "failed", result.err.Error())

			// Обновляем статус в job
			if photoInfo, exists := job.PhotoProgress[result.photo.ID]; exists {
				photoInfo.Status = "failed"
				photoInfo.Error = result.err.Error()
				job.PhotoProgress[result.photo.ID] = photoInfo
			}
		} else {
			// Логируем успех
			q.dbService.LogEvent(batch.ID, result.photo.ID, "ai_processing", "success",
				fmt.Sprintf("AI обработка фото %s завершена успешно", result.photo.FileName), "", 100)

			// Помечаем фото как processed только при успехе
			q.updatePhotoStatus(result.photo.ID, "processed", "")

			// Обновляем статус в job
			if photoInfo, exists := job.PhotoProgress[result.photo.ID]; exists {
				photoInfo.Status = "completed"
				photoInfo.Progress = 100
				photoInfo.Step = "completed"
				job.PhotoProgress[result.photo.ID] = photoInfo
			}
		}

		// Обновляем общий прогресс
		if job, exists := q.activeJobs[batch.ID]; exists {
			job.Progress = (processedCount * 100) / len(batch.Photos)
			// Обновляем текущее фото (показываем последнее обработанное)
			job.CurrentPhoto = result.photo.FileName
		}

		log.Printf("Photo %s marked as processed. Total processed: %d/%d", result.photo.FileName, processedCount, len(batch.Photos))
	}

	// Завершаем обработку батча
	job.Progress = 100
	job.Status = "completed"
	job.CurrentStep = "completed"
	q.updateBatchStatus(batch.ID, "processed", "")

	// Логируем завершение обработки батча
	q.dbService.LogEvent(batch.ID, "", "batch_complete", "success",
		fmt.Sprintf("Обработка батча завершена. Обработано: %d/%d фотографий", processedCount, len(batch.Photos)), "", 100)

	log.Printf("Batch %s processing completed. Successfully processed: %d/%d photos", batch.ID, processedCount, len(batch.Photos))
	return nil
}

// processPhoto обрабатывает одно фото
func (q *QueueManager) processPhoto(photo *models.Photo, batchDescription string, contentType string, settings models.AppSettings, job *ProcessingJob) error {
	log.Printf("Starting to process photo %s (content type: %s)", photo.FileName, contentType)

	// Шаг 1: Подготавливаем фото для AI (создаем миниатюру, извлекаем EXIF)
	log.Printf("Step 1: Preparing photo %s for AI", photo.FileName)
	q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "progress",
		fmt.Sprintf("Подготовка фото %s для AI анализа", photo.FileName), "", 10)

	// Обновляем прогресс в job
	if photoInfo, exists := job.PhotoProgress[photo.ID]; exists {
		photoInfo.Step = "preparation"
		photoInfo.Progress = 10
		job.PhotoProgress[photo.ID] = photoInfo
	}

	err := q.imageProcessor.ProcessPhotoForAI(photo, settings.ThumbnailSize)
	if err != nil {
		log.Printf("Failed to prepare photo %s for AI: %v", photo.FileName, err)
		q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "failed",
			fmt.Sprintf("Ошибка подготовки фото %s", photo.FileName), err.Error(), 0)
		return fmt.Errorf("failed to prepare photo for AI: %w", err)
	}
	log.Printf("Photo %s prepared for AI successfully", photo.FileName)

	// Сохраняем thumbnail path в базе данных
	if photo.ThumbnailPath != "" {
		err = q.dbService.UpdatePhotoThumbnail(photo.ID, photo.ThumbnailPath)
		if err != nil {
			log.Printf("Warning: failed to save thumbnail path for %s: %v", photo.FileName, err)
		}
	}

	// Шаг 2: Отправляем в AI для анализа
	log.Printf("Step 2: Analyzing photo %s with AI", photo.FileName)
	q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "progress",
		fmt.Sprintf("Отправка фото %s на AI анализ", photo.FileName), "", 30)

	// Обновляем прогресс в job
	if photoInfo, exists := job.PhotoProgress[photo.ID]; exists {
		photoInfo.Step = "ai_analysis"
		photoInfo.Progress = 30
		job.PhotoProgress[photo.ID] = photoInfo
	}

	aiResult, err := q.aiService.AnalyzePhoto(*photo, batchDescription, contentType, settings)
	if err != nil {
		log.Printf("Failed to analyze photo %s with AI: %v", photo.FileName, err)
		q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "failed",
			fmt.Sprintf("Ошибка AI анализа фото %s", photo.FileName), err.Error(), 30)
		return fmt.Errorf("failed to analyze photo with AI: %w", err)
	}
	log.Printf("Photo %s analyzed successfully, got title: %s", photo.FileName, aiResult.Title)

	// Шаг 3: Сохраняем результаты AI
	log.Printf("Step 3: Saving AI results for photo %s", photo.FileName)
	q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "progress",
		fmt.Sprintf("Сохранение результатов AI для фото %s", photo.FileName), "", 70)

	// Обновляем прогресс в job
	if photoInfo, exists := job.PhotoProgress[photo.ID]; exists {
		photoInfo.Step = "saving"
		photoInfo.Progress = 70
		job.PhotoProgress[photo.ID] = photoInfo
	}

	err = q.dbService.UpdatePhotoAIResults(photo.ID, *aiResult)
	if err != nil {
		log.Printf("Failed to save AI results for photo %s: %v", photo.FileName, err)
		q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "failed",
			fmt.Sprintf("Ошибка сохранения AI результатов для фото %s", photo.FileName), err.Error(), 70)
		return fmt.Errorf("failed to save AI results: %w", err)
	}
	log.Printf("AI results saved for photo %s", photo.FileName)

	// Шаг 4: Записываем метаданные в EXIF оригинального файла
	log.Printf("Step 4: Writing EXIF data to photo %s", photo.FileName)
	q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "progress",
		fmt.Sprintf("Запись EXIF данных в фото %s", photo.FileName), "", 90)

	// Обновляем прогресс в job
	if photoInfo, exists := job.PhotoProgress[photo.ID]; exists {
		photoInfo.Step = "exif_writing"
		photoInfo.Progress = 90
		job.PhotoProgress[photo.ID] = photoInfo
	}

	err = q.imageProcessor.WriteExifToImage(photo.OriginalPath, *aiResult)
	if err != nil {
		log.Printf("Warning: failed to write EXIF to %s: %v", photo.OriginalPath, err)
		q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "warning",
			fmt.Sprintf("Предупреждение при записи EXIF в фото %s", photo.FileName), err.Error(), 90)
		// Не фейлим весь процесс из-за EXIF ошибки
	} else {
		log.Printf("EXIF data written successfully to photo %s", photo.FileName)
		q.dbService.LogEvent(photo.BatchID, photo.ID, "ai_processing", "success",
			fmt.Sprintf("AI обработка фото %s завершена успешно", photo.FileName),
			fmt.Sprintf("Название: %s, Ключевых слов: %d", aiResult.Title, len(aiResult.Keywords)), 100)
	}

	log.Printf("Photo %s processing completed successfully", photo.FileName)
	return nil
}

// GetQueueStatus возвращает текущий статус очереди
func (q *QueueManager) GetQueueStatus() ([]models.BatchStatus, error) {
	var statuses []models.BatchStatus

	// Получаем активные задачи
	q.jobsMutex.RLock()
	activeJobs := make(map[string]*ProcessingJob)
	for k, v := range q.activeJobs {
		activeJobs[k] = v
	}
	q.jobsMutex.RUnlock()

	// Получаем батчи в очереди и в процессе
	rows, err := q.db.Query(`
		SELECT id, type, description, status, created_at,
		       (SELECT COUNT(*) FROM photos WHERE batch_id = batches.id) as total_photos,
		       (SELECT COUNT(*) FROM photos WHERE batch_id = batches.id AND status = 'processed') as processed_photos
		FROM batches 
		WHERE status IN ('queued', 'processing', 'processed')
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query queue status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var batchID, batchType, description, status string
		var createdAt time.Time
		var totalPhotos, processedPhotos int

		err := rows.Scan(&batchID, &batchType, &description, &status, &createdAt, &totalPhotos, &processedPhotos)
		if err != nil {
			continue
		}

		batchStatus := models.BatchStatus{
			BatchID:         batchID,
			Type:            batchType,
			Description:     description,
			TotalPhotos:     totalPhotos,
			ProcessedPhotos: processedPhotos,
			Status:          status,
		}

		// Если задача активна, берем более точный прогресс
		if job, exists := activeJobs[batchID]; exists {
			batchStatus.Progress = job.Progress
			batchStatus.CurrentPhoto = job.CurrentPhoto
			batchStatus.CurrentStep = job.CurrentStep
			batchStatus.Status = job.Status
			if job.Error != "" {
				batchStatus.Error = job.Error
			}

			// Добавляем детальную информацию по фотографиям
			photos := make([]models.PhotoProcessInfo, 0, len(job.PhotoProgress))
			for _, photoInfo := range job.PhotoProgress {
				photos = append(photos, photoInfo)
			}
			batchStatus.Photos = photos
		} else {
			// Вычисляем прогресс на основе обработанных фото
			if totalPhotos > 0 {
				batchStatus.Progress = int(float64(processedPhotos) / float64(totalPhotos) * 100)
			}
		}

		statuses = append(statuses, batchStatus)
	}

	return statuses, nil
}

// updateBatchStatus обновляет статус батча
func (q *QueueManager) updateBatchStatus(batchID string, status string, errorMsg string) {
	_, err := q.db.Exec(`
		UPDATE batches 
		SET status = ?, updated_at = datetime('now') 
		WHERE id = ?`,
		status, batchID)
	if err != nil {
		log.Printf("Failed to update batch status: %v", err)
	}
}

// updatePhotoStatus обновляет статус фото
func (q *QueueManager) updatePhotoStatus(photoID string, status string, errorMsg string) {
	_, err := q.db.Exec(`
		UPDATE photos 
		SET status = ?, updated_at = datetime('now') 
		WHERE id = ?`,
		status, photoID)
	if err != nil {
		log.Printf("Failed to update photo status: %v", err)
	}
}

// getPhotosForBatch получает фотографии для батча
func (q *QueueManager) getPhotosForBatch(batchID string) ([]models.Photo, error) {
	rows, err := q.db.Query(`
		SELECT id, batch_id, content_type, original_path, thumbnail_path, file_name, file_size,
		       exif_data, ai_results, upload_status, status, created_at
		FROM photos 
		WHERE batch_id = ? AND status IN ('pending', 'processing', 'processed', 'failed')`, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var photo models.Photo
		var exifJSON, aiResultsJSON, uploadStatusJSON string

		err := rows.Scan(&photo.ID, &photo.BatchID, &photo.ContentType, &photo.OriginalPath,
			&photo.ThumbnailPath, &photo.FileName, &photo.FileSize,
			&exifJSON, &aiResultsJSON, &uploadStatusJSON,
			&photo.Status, &photo.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Десериализуем JSON поля (если не пустые)
		if exifJSON != "" {
			json.Unmarshal([]byte(exifJSON), &photo.ExifData)
		}
		if aiResultsJSON != "" {
			var aiResult models.AIResult
			err := json.Unmarshal([]byte(aiResultsJSON), &aiResult)
			if err == nil {
				photo.AIResult = &aiResult
			}
		}
		if uploadStatusJSON != "" {
			json.Unmarshal([]byte(uploadStatusJSON), &photo.UploadStatus)
		}

		photos = append(photos, photo)
	}

	return photos, nil
}

// photoWorker обрабатывает фотографии из канала
func (q *QueueManager) photoWorker(workerID int, photoChannel <-chan models.Photo, resultChannel chan<- photoResult, batchID, batchDescription, contentType string, settings models.AppSettings, job *ProcessingJob) {
	log.Printf("Worker %d started", workerID)

	for photo := range photoChannel {
		log.Printf("Worker %d processing photo: %s", workerID, photo.FileName)

		// Обновляем статус фотографии в job
		if photoInfo, exists := job.PhotoProgress[photo.ID]; exists {
			photoInfo.Status = "processing"
			photoInfo.Step = "ai_processing"
			job.PhotoProgress[photo.ID] = photoInfo
		}

		// Логируем начало обработки фото
		q.dbService.LogEvent(batchID, photo.ID, "ai_processing", "started",
			fmt.Sprintf("Начата AI обработка фото %s (worker %d)", photo.FileName, workerID), "", 0)

		// Обрабатываем фото
		err := q.processPhoto(&photo, batchDescription, contentType, settings, job)

		// Отправляем результат
		resultChannel <- photoResult{photo: photo, err: err}

		log.Printf("Worker %d finished processing photo: %s", workerID, photo.FileName)
	}

	log.Printf("Worker %d finished", workerID)
}
