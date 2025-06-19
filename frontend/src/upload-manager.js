// Модуль управления загрузкой фотографий
export class UploadManager {
    constructor(app) {
        this.app = app;
        this.uploadQueueInterval = null;
        this.selectedPhotos = new Set();
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        // Функция для проверки готовности DOM
        const waitForElements = () => {
            const selectAllBtn = document.getElementById('selectAllForUploadBtn');
            const clearSelectionBtn = document.getElementById('clearSelectionBtn');
            const uploadSelectedBtn = document.getElementById('uploadSelectedBtn');
            const stopUploadQueueBtn = document.getElementById('stopUploadQueueBtn');
            
            return selectAllBtn && clearSelectionBtn && uploadSelectedBtn && stopUploadQueueBtn;
        };
        
        // Пытаемся инициализировать сразу, если элементы готовы
        if (waitForElements()) {
            this.setupEventListeners();
            return;
        }
        
        // Если элементы не готовы, ждем с интервалами
        let retryCount = 0;
        const maxRetries = 50; // максимум 5 секунд ожидания
        
        const retryInterval = setInterval(() => {
            retryCount++;
            
            if (waitForElements()) {
                clearInterval(retryInterval);
                this.setupEventListeners();
            } else if (retryCount >= maxRetries) {
                clearInterval(retryInterval);
                console.warn('Upload manager initialization timeout - some buttons may not work');
                // Пытаемся инициализировать то что можем
                this.setupEventListeners();
            }
        }, 100);
    }
    
    setupEventListeners() {
        try {
            // Кнопки массового выбора
            const selectAllBtn = document.getElementById('selectAllForUploadBtn');
            if (selectAllBtn) {
                selectAllBtn.addEventListener('click', () => {
                    this.selectAllPhotos();
                });
            } else {
                console.warn('selectAllForUploadBtn not found');
            }

            const clearSelectionBtn = document.getElementById('clearSelectionBtn');
            if (clearSelectionBtn) {
                clearSelectionBtn.addEventListener('click', () => {
                    this.clearAllSelection();
                });
            } else {
                console.warn('clearSelectionBtn not found');
            }

            const uploadSelectedBtn = document.getElementById('uploadSelectedBtn');
            if (uploadSelectedBtn) {
                uploadSelectedBtn.addEventListener('click', () => {
                    this.uploadSelectedPhotos();
                });
            } else {
                console.warn('uploadSelectedBtn not found');
            }

            const stopUploadQueueBtn = document.getElementById('stopUploadQueueBtn');
            if (stopUploadQueueBtn) {
                stopUploadQueueBtn.addEventListener('click', () => {
                    this.stopUploadQueue();
                });
            } else {
                console.warn('stopUploadQueueBtn not found');
            }
        } catch (error) {
            console.error('Error initializing upload manager event listeners:', error);
        }
    }

    // Переключение выбора фотографии
    async togglePhotoSelection(photoId, selected) {
        try {
            await window.go.main.App.SetPhotoSelectedForUpload(photoId, selected);
            
            if (selected) {
                this.selectedPhotos.add(photoId);
            } else {
                this.selectedPhotos.delete(photoId);
            }
            
            this.updateSelectionUI();
        } catch (error) {
            console.error('Error toggling photo selection:', error);
            this.app.showNotification('Error updating photo selection: ' + error.message, 'error');
        }
    }

    // Выбрать все фотографии
    async selectAllPhotos() {
        const batchSelector = document.getElementById('batchSelector');
        if (!batchSelector) {
            console.error('batchSelector element not found');
            this.app.showNotification('Batch selector not found', 'error');
            return;
        }
        
        const batchId = batchSelector.value;
        
        if (!batchId || batchId.trim() === '') {
            this.app.showNotification('No batch selected', 'error');
            return;
        }

        try {
            await window.go.main.App.SelectAllPhotosForUpload(batchId);
            this.app.showNotification('All photos selected for upload', 'success');
            
            // Обновляем интерфейс
            this.app.loadBatchForReview(batchId);
        } catch (error) {
            console.error('Error selecting all photos:', error);
            this.app.showNotification('Error selecting all photos: ' + (error.message || error), 'error');
        }
    }

    // Очистить весь выбор
    async clearAllSelection() {
        const batchSelector = document.getElementById('batchSelector');
        if (!batchSelector) {
            console.error('batchSelector element not found');
            this.app.showNotification('Batch selector not found', 'error');
            return;
        }
        
        const batchId = batchSelector.value;
        
        if (!batchId || batchId.trim() === '') {
            this.app.showNotification('No batch selected', 'error');
            return;
        }

        try {
            await window.go.main.App.ClearAllPhotoSelection(batchId);
            this.app.showNotification('All photo selections cleared', 'success');
            
            // Обновляем интерфейс
            this.selectedPhotos.clear();
            this.app.loadBatchForReview(batchId);
        } catch (error) {
            console.error('Error clearing selection:', error);
            this.app.showNotification('Error clearing selection: ' + (error.message || error), 'error');
        }
    }

    // Загрузить выбранные фотографии
    async uploadSelectedPhotos() {
        const batchSelector = document.getElementById('batchSelector');
        if (!batchSelector) {
            console.error('batchSelector element not found');
            this.app.showNotification('Batch selector not found', 'error');
            return;
        }
        
        const batchId = batchSelector.value;
        
        if (!batchId || batchId.trim() === '') {
            this.app.showNotification('No batch selected', 'error');
            return;
        }

        // Собираем ID выбранных фотографий
        const selectedPhotoIds = Array.from(document.querySelectorAll('.photo-select-checkbox:checked'))
            .map(checkbox => checkbox.dataset.photoId);

        if (selectedPhotoIds.length === 0) {
            this.app.showNotification('No photos selected for upload', 'error');
            return;
        }

        const uploadBtn = document.getElementById('uploadSelectedBtn');
        const originalText = uploadBtn.innerHTML;
        
        try {
            // Показываем прогресс
            uploadBtn.disabled = true;
            uploadBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i> Uploading...';
            

            await window.go.main.App.UploadSelectedPhotos(batchId, selectedPhotoIds);
            
            // Начинаем отслеживание прогресса
            this.startUploadQueueTracking();
            
            this.app.showNotification(`Started upload of ${selectedPhotoIds.length} selected photos`, 'success');
            
        } catch (error) {
            console.error('Error uploading selected photos:', error);
            this.app.showNotification('Error starting upload: ' + (error.message || error), 'error');
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.innerHTML = originalText;
        }
    }

    // Остановить очередь загрузки
    async stopUploadQueue() {
        try {
            await window.go.main.App.StopUploadQueue();
            this.app.showNotification('Upload queue stopped', 'success');
            this.stopUploadQueueTracking();
        } catch (error) {
            console.error('Error stopping upload queue:', error);
            this.app.showNotification('Error stopping upload queue: ' + error.message, 'error');
        }
    }

    // Отслеживание статуса очереди загрузки
    startUploadQueueTracking() {
        // Показываем секцию статуса
        document.getElementById('uploadQueueStatus').classList.remove('hidden');
        
        // Проверяем статус каждые 2 секунды
        this.uploadQueueInterval = setInterval(async () => {
            try {
                const status = await window.go.main.App.GetUploadQueueStatus();
                this.updateUploadQueueStatus(status);
                
                // Если очередь пуста и нет активных загрузок, останавливаем отслеживание
                if (!status.isProcessing || (status.activeUploads === 0 && status.queueLength === 0)) {
                    this.stopUploadQueueTracking();
                    
                    // Обновляем интерфейс Review через некоторое время
                    setTimeout(() => {
                        const batchId = document.getElementById('batchSelector').value;
                        if (batchId) {
                            this.app.loadBatchForReview(batchId);
                        }
                    }, 1000);
                }
            } catch (error) {
                console.error('Error tracking upload queue:', error);
                this.stopUploadQueueTracking();
            }
        }, 2000);
    }

    // Остановить отслеживание очереди
    stopUploadQueueTracking() {
        if (this.uploadQueueInterval) {
            clearInterval(this.uploadQueueInterval);
            this.uploadQueueInterval = null;
        }
        
        // Скрываем секцию статуса через 3 секунды
        setTimeout(() => {
            document.getElementById('uploadQueueStatus').classList.add('hidden');
        }, 3000);
    }

    // Обновить статус очереди загрузки в UI
    updateUploadQueueStatus(status) {
        document.getElementById('activeUploadsCount').textContent = status.activeUploads || 0;
        document.getElementById('queueLengthCount').textContent = status.queueLength || 0;
        document.getElementById('maxConcurrentCount').textContent = status.maxConcurrent || 2;
        
        const statusText = status.isProcessing ? 'Processing' : 'Idle';
        document.getElementById('queueStatusText').textContent = statusText;
        
        // Обновляем список активных заданий
        this.renderActiveJobs(status.activeJobs || []);
    }

    // Отображение активных заданий
    renderActiveJobs(activeJobs) {
        const container = document.getElementById('activeJobsList');
        
        if (!activeJobs || activeJobs.length === 0) {
            container.innerHTML = '';
            return;
        }

        const jobsHTML = activeJobs.map(job => {
            const progressStocks = Object.entries(job.progress || {}).map(([stockId, status]) => {
                const statusClass = status === 'uploaded' ? 'text-green-600' : 
                                  status === 'failed' ? 'text-red-600' : 'text-blue-600';
                const statusIcon = status === 'uploaded' ? 'fa-check' : 
                                 status === 'failed' ? 'fa-times' : 'fa-spinner fa-spin';
                return `<span class="${statusClass}"><i class="fas ${statusIcon} mr-1"></i>${stockId.substr(0, 8)}</span>`;
            }).join(' ');

            return `
                <div class="bg-white rounded p-3 border border-gray-200">
                    <div class="flex justify-between items-center">
                        <div class="flex-1">
                            <div class="font-medium text-sm">${job.fileName}</div>
                            <div class="text-xs text-gray-500">Status: ${job.status}</div>
                        </div>
                        <div class="text-right text-xs">
                            ${progressStocks}
                        </div>
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = jobsHTML;
    }

    // Обновить UI выбора
    updateSelectionUI() {
        const selectedCount = document.querySelectorAll('.photo-select-checkbox:checked').length;
        const uploadBtn = document.getElementById('uploadSelectedBtn');
        
        if (uploadBtn) {
            uploadBtn.disabled = selectedCount === 0;
            const span = uploadBtn.querySelector('span');
            if (span) {
                span.textContent = selectedCount > 0 ? `Upload Selected (${selectedCount})` : 'Upload Selected';
            }
        }
    }
}

// Экспортируем для глобального доступа
window.UploadManager = UploadManager; 