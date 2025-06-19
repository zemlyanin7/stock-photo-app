// Импортируем i18n
import './i18n.js';
import { UploadManager } from './upload-manager.js';
import { EventsOn, OnFileDrop } from '../wailsjs/runtime/runtime.js';

// main.js file loaded

// Main application JavaScript
class StockPhotoApp {
    constructor(isRealWailsApp = false) {
        this.currentTab = 'editorial';
        this.selectedFolder = null;
        this.settings = null;
        this.queueUpdateInterval = null;
        this.isWailsMode = isRealWailsApp;
        this.uploadManager = null;
        
        this.init();
    }

    async init() {
        this.updateLoadingMessage('Loading translations...');
        
        // Инициализируем i18n
        await window.i18n.init();
        
        this.updateLoadingMessage('Loading settings...');
        
        // Загружаем настройки (включая язык) за один раз
        await this.loadSettings();
        
        this.updateLoadingMessage('Setting up interface...');
        
        this.setupEventListeners();
        this.setupAIModelSelector();
        this.setupDragAndDrop('editorialDropZone', 'editorial');
        this.setupDragAndDrop('commercialDropZone', 'commercial');
        this.startQueueUpdates();
        
        this.updateLoadingMessage('Configuring drag & drop...');
        
        // Показываем уведомление о режиме работы
        if (this.isWailsMode) {
            // Настраиваем drag & drop с помощью Wails API
            try {
                OnFileDrop((x, y, paths) => {
                    this.handleWailsFileDrop(x, y, paths);
                }, false); // false = обрабатываем drop в любом месте окна
                

            } catch (error) {
                console.error('Error initializing drag & drop:', error);
            }
        }
        
        // Обновляем индикатор режима
        this.updateModeIndicator();
        
        // Инициализируем менеджер загрузки после основной инициализации
        this.updateLoadingMessage('Initializing upload manager...');
        try {
            // Добавляем небольшую задержку, чтобы убедиться что DOM полностью готов
            setTimeout(() => {
                try {
                    this.uploadManager = new UploadManager(this);

                } catch (error) {
                    console.error('Error initializing upload manager:', error);
                }
            }, 100);
        } catch (error) {
            console.error('Error setting up upload manager initialization:', error);
        }
        
        // Скрываем экран загрузки и показываем приложение
        this.hideLoadingScreen();
        
        // Показываем уведомление о готовности
        if (this.isWailsMode) {
            this.showNotification(window.i18n.t('app.ready'), 'success');
        } else {
            this.showNotification(window.i18n.t('notifications.demoMode'), 'warning');
        }
    }

    updateLoadingMessage(message) {
        const loadingMessage = document.getElementById('loadingMessage');
        if (loadingMessage) {
            loadingMessage.textContent = message;
        }
    }

    hideLoadingScreen() {
        const loadingScreen = document.getElementById('loadingScreen');
        if (loadingScreen) {
            loadingScreen.style.opacity = '0';
            loadingScreen.style.transition = 'opacity 0.3s ease';
            setTimeout(() => {
                loadingScreen.remove();
            }, 300);
        }
    }

    setupEventListeners() {

        
        // Tab switching
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchTab(tab);
            });
        });

        // Settings tabs
        document.querySelectorAll('.settings-tab').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.target.dataset.tab;
                this.switchSettingsTab(tab);
            });
        });

        // Language switcher
        const languageSelector = document.getElementById('languageSelector');
        if (languageSelector) {
            languageSelector.addEventListener('change', (e) => {
                this.changeLanguage(e.target.value);
            });
        } else {
            console.error('Language selector not found');
        }

        // Folder selection
        const selectFolderBtn = document.getElementById('selectFolderBtn');
        if (selectFolderBtn) {
            selectFolderBtn.addEventListener('click', () => {
                this.selectFolder();
            });
        }

        // Show files lists
        const showEditorialFilesBtn = document.getElementById('showEditorialFilesBtn');
        if (showEditorialFilesBtn) {
            showEditorialFilesBtn.addEventListener('click', () => {
                this.showFilesList('editorial');
            });
        }
        
        const showCommercialFilesBtn = document.getElementById('showCommercialFilesBtn');
        if (showCommercialFilesBtn) {
            showCommercialFilesBtn.addEventListener('click', () => {
                this.showFilesList('commercial');
            });
        }

        // Settings modal
        const settingsBtn = document.getElementById('settingsBtn');
        if (settingsBtn) {
            settingsBtn.addEventListener('click', () => {
                this.openSettings();
            });
        }
        
        const closeSettingsBtn = document.getElementById('closeSettingsBtn');
        if (closeSettingsBtn) {
            closeSettingsBtn.addEventListener('click', () => {
                this.closeSettings();
            });
        }

        // AI provider change
        const aiProvider = document.getElementById('aiProvider');
        if (aiProvider) {
            aiProvider.addEventListener('change', () => {
                this.populateSettingsForm();
            });
        }

        // Process buttons
        document.getElementById('processEditorialBtn').addEventListener('click', () => {
            this.processPhotos('editorial');
        });
        document.getElementById('processCommercialBtn').addEventListener('click', () => {
            this.processPhotos('commercial');
        });

        // Save settings
        document.getElementById('saveSettingsBtn').addEventListener('click', () => {
            this.saveSettings();
        });

        // Queue refresh
        document.getElementById('refreshQueueBtn').addEventListener('click', () => {
            this.updateQueue();
        });

        // History refresh
        document.getElementById('refreshHistoryBtn').addEventListener('click', () => {
            this.updateHistory();
        });

        // Review refresh
        document.getElementById('refreshReviewBtn').addEventListener('click', () => {
            this.updateReview();
        });

        // Batch selector change
        document.getElementById('batchSelector').addEventListener('change', (e) => {
            this.loadBatchForReview(e.target.value);
        });

        // Batch actions - используем делегирование событий
        document.addEventListener('click', (e) => {
            if (e.target.id === 'uploadToStocksBtn' || e.target.closest('#uploadToStocksBtn')) {
                e.preventDefault();
                this.uploadToStocks();
            }
            if (e.target.id === 'deleteBatchBtn' || e.target.closest('#deleteBatchBtn')) {
                e.preventDefault();
                this.deleteBatch();
            }
            if (e.target.id === 'approveAllBtn' || e.target.closest('#approveAllBtn')) {
                e.preventDefault();
                this.approveAllPhotos();
            }
            if (e.target.id === 'rejectAllBtn' || e.target.closest('#rejectAllBtn')) {
                e.preventDefault();
                this.rejectAllPhotos();
            }
            if (e.target.id === 'regenerateAllBtn' || e.target.closest('#regenerateAllBtn')) {
                e.preventDefault();
                this.regenerateAllPhotos();
            }
        });

        // Stock management
        document.getElementById('addStockBtn').addEventListener('click', () => {
            this.openAddStockModal();
        });
        document.getElementById('closeAddStockBtn').addEventListener('click', () => {
            this.closeAddStockModal();
        });
        document.getElementById('cancelAddStockBtn').addEventListener('click', () => {
            this.closeAddStockModal();
        });
        
        // Add stock form
        document.getElementById('addStockForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.addStock(e);
        });

        // Stock type change for dynamic fields
        document.getElementById('stockType').addEventListener('change', (e) => {
            this.onStockTypeChange(e.target.value);
        });

        // Test AI connection
        document.getElementById('testAiConnectionBtn').addEventListener('click', () => {
            this.testAIConnection();
        });

        // Test stock connection in add modal
        document.getElementById('testStockConnectionBtn').addEventListener('click', () => {
            this.testStockConnectionInModal();
        });

        // Reset prompt buttons
        document.getElementById('resetEditorialPromptBtn').addEventListener('click', () => {
            this.resetPromptToDefault('editorial');
        });
        
        document.getElementById('resetCommercialPromptBtn').addEventListener('click', () => {
            this.resetPromptToDefault('commercial');
        });

        // Force update prompts button
        document.getElementById('forceUpdatePromptsBtn').addEventListener('click', () => {
            this.forceUpdatePrompts();
        });

        // Подписываемся на изменения языка
        window.i18n.subscribe((language) => {
            this.updateLanguageSelectors(language);
            this.updateDynamicContent();
        });
    }

    updateLanguageSelectors(language) {
        // Обновляем селекторы языка
        document.getElementById('languageSelector').value = language;
        document.getElementById('settingsLanguage').value = language;
    }

    updateDynamicContent() {
        // Обновляем динамический контент, который не обновляется автоматически
        this.updateDropZones();
        this.updateButtonTexts();
    }

    updateDropZones() {
        // Обновляем зоны перетаскивания если папка не выбрана
        if (!this.selectedFolder || this.selectedFolder.type !== 'editorial') {
            this.resetDropZone('editorial');
        }
        if (!this.selectedFolder || this.selectedFolder.type !== 'commercial') {
            this.resetDropZone('commercial');
        }
    }

    updateButtonTexts() {
        // Обновляем тексты кнопок обработки
        const editorialBtn = document.getElementById('processEditorialBtn');
        const commercialBtn = document.getElementById('processCommercialBtn');
        
        if (!editorialBtn.disabled) {
            editorialBtn.querySelector('span').textContent = window.i18n.t('editorial.process');
        }
        if (!commercialBtn.disabled) {
            commercialBtn.querySelector('span').textContent = window.i18n.t('commercial.process');
        }
    }

    resetDropZone(photoType) {
        const dropZone = document.getElementById(photoType + 'DropZone');
        dropZone.innerHTML = `
            <div class="space-y-2">
                <i class="fas fa-cloud-upload-alt text-4xl text-gray-400"></i>
                <div>
                    <p class="text-lg text-gray-600">${window.i18n.t(photoType + '.dropZone.drag')}</p>
                    <p class="text-sm text-gray-500">${window.i18n.t(photoType + '.dropZone.browse')}</p>
                </div>
            </div>
        `;
    }

    async changeLanguage(language) {
        try {
            await window.i18n.switchLanguage(language);
            
            // Автоматически сохраняем язык в настройках
            if (this.settings) {
                this.settings.language = language;
                await window.go.main.App.SaveSettings(this.settings);
            } else {
                // Если настройки еще не загружены, создаем базовые настройки с языком
                const basicSettings = {
                    tempDirectory: './temp',
                    aiProvider: 'openai',
                    thumbnailSize: 512,
                    maxConcurrentJobs: 3,
                    language: language,
                    aiPrompts: {
                        editorial: '',
                        commercial: ''
                    }
                };
                await window.go.main.App.SaveSettings(basicSettings);
                this.settings = basicSettings;
            }
        } catch (error) {
            console.error('Error changing language:', error);
            this.showNotification(window.i18n.t('app.error'), 'error');
        }
    }

    setupDragAndDrop(elementId, photoType) {
        const dropZone = document.getElementById(elementId);
        
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('drag-over');
        });

        dropZone.addEventListener('dragleave', (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
        });

        dropZone.addEventListener('drop', async (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
            
            if (!this.isWailsMode) {
                this.showNotification(window.i18n.t('notifications.dragNotSupported'), 'warning');
                return;
            }
            
            const items = Array.from(e.dataTransfer.items);
            const files = Array.from(e.dataTransfer.files);
            
            // Пытаемся определить папку
            let folderPath = null;
            
            // Способ 1: Проверяем items для папок (современные браузеры)
            for (const item of items) {
                if (item.kind === 'file') {
                    const entry = item.webkitGetAsEntry ? item.webkitGetAsEntry() : null;
                    if (entry && entry.isDirectory) {
                        // Это папка, но в Wails нам нужен реальный путь файловой системы
                        // Показываем уведомление что нужно использовать кнопку "Browse"
                        this.showNotification(window.i18n.t('notifications.useBrowseButton'), 'info');
                        return;
                    }
                }
            }
            
            // Способ 2: Получаем путь из файлов (если перетащили файлы из папки)
            if (files.length > 0 && !folderPath) {
                const firstFile = files[0];
                if (firstFile.path) {
                    // В Electron/Wails файлы имеют свойство path
                    folderPath = firstFile.path.substring(0, firstFile.path.lastIndexOf('/'));
                } else if (firstFile.webkitRelativePath) {
                    // Для input[type="file"] с webkitdirectory
                    const relativePath = firstFile.webkitRelativePath;
                    folderPath = relativePath.substring(0, relativePath.lastIndexOf('/'));
                } else {
                    // Файлы без информации о пути - показываем уведомление
                    this.showNotification(window.i18n.t('notifications.cannotDetermineFolder'), 'warning');
                    return;
                }
            }
            
            // Если нет файлов или папок - показываем уведомление
            if (!folderPath && files.length === 0) {
                this.showNotification(window.i18n.t('notifications.dragFilesHint'), 'info');
                return;
            }
            
            // Если не удалось определить путь к папке
            if (!folderPath) {
                this.showNotification(window.i18n.t('notifications.folderSelectionError'), 'warning');
                return;
            }
            
            // Обрабатываем найденную папку
            await this.selectFolder(folderPath, photoType);
        });

        // Click to browse
        dropZone.addEventListener('click', (e) => {
            // Проверяем что клик не был по кнопке просмотра файлов
            if (e.target.classList.contains('view-files-btn')) {
                return; // Не открываем диалог, если кликнули на кнопку просмотра
            }
            this.browseFolder(photoType);
        });
    }

    async selectFolder(folderPath, photoType) {
        try {
            // Загружаем содержимое папки
            const files = await window.go.main.App.GetFolderContents(folderPath);
            
            const validFiles = files.filter(file => file.isValid);
            if (validFiles.length === 0) {
                this.showNotification(window.i18n.t('notifications.noValidImages'), 'warning');
                return;
            }
            
            this.selectedFolder = { 
                path: folderPath, 
                type: photoType, 
                files: files,
                validCount: validFiles.length 
            };
            
            const dropZone = document.getElementById(photoType + 'DropZone');
            dropZone.innerHTML = `
                <div class="space-y-3">
                    <i class="fas fa-folder-open text-4xl text-green-500"></i>
                    <div>
                        <p class="text-lg text-gray-900">${window.i18n.t(photoType + '.dropZone.selected')}</p>
                        <p class="text-sm text-gray-500">${folderPath}</p>
                        <p class="text-sm text-blue-600">${validFiles.length} ${window.i18n.t('notifications.validImages')}</p>
                    </div>
                    <button class="view-files-btn text-sm text-blue-600 hover:text-blue-800 underline" data-photo-type="${photoType}">
                        ${window.i18n.t('notifications.viewFiles')}
                    </button>
                </div>
            `;
            
            // Добавляем обработчик события для кнопки просмотра файлов
            const viewFilesBtn = dropZone.querySelector('.view-files-btn');
            if (viewFilesBtn) {
                viewFilesBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    this.showFilesList(photoType);
                });
            }
            
            this.showNotification(`${window.i18n.t('notifications.folderSelected')}: ${validFiles.length} ${window.i18n.t('notifications.validImages')}`, 'success');
        } catch (error) {
            console.error('Error loading folder contents:', error);
            this.showNotification(window.i18n.t('notifications.errorProcessing') + ': ' + error.message, 'error');
        }
    }

    async browseFolder(photoType) {
        if (!this.isWailsMode) {
            this.showNotification(window.i18n.t('notifications.desktopAppOnly'), 'warning');
            return;
        }
        
        try {
            const folderPath = await window.go.main.App.SelectFolder();
            if (folderPath) {
                await this.selectFolder(folderPath, photoType);
            }
        } catch (error) {
            console.error('Error selecting folder:', error);
            this.showNotification(window.i18n.t('notifications.errorProcessing') + ': ' + error.message, 'error');
        }
    }

    async processPhotos(photoType) {
        if (!this.selectedFolder || this.selectedFolder.type !== photoType) {
            this.showNotification(window.i18n.t('notifications.selectFolder'), 'error');
            return;
        }

        const description = document.getElementById(photoType + 'Description').value.trim();

        const processBtn = document.getElementById('process' + photoType.charAt(0).toUpperCase() + photoType.slice(1) + 'Btn');
        processBtn.disabled = true;
        processBtn.innerHTML = `<i class="fas fa-spinner fa-spin mr-2"></i><span>${window.i18n.t(photoType + '.processing')}</span>`;

        try {
            // Вызываем Go метод, передаем описание как есть (может быть пустым)
            await window.go.main.App.ProcessPhotoFolder(this.selectedFolder.path, description, photoType);
            
            this.showNotification(window.i18n.t('notifications.photosQueued'), 'success');
            this.switchTab('queue');
            this.updateQueue();
            
            // Сбрасываем форму
            this.resetForm(photoType);
        } catch (error) {
            console.error('Error processing photos:', error);
            this.showNotification(window.i18n.t('notifications.errorProcessing') + ': ' + error.message, 'error');
        } finally {
            processBtn.disabled = false;
            processBtn.innerHTML = `<i class="fas fa-play mr-2"></i><span>${window.i18n.t(photoType + '.process')}</span>`;
        }
    }

    resetForm(photoType) {
        this.selectedFolder = null;
        document.getElementById(photoType + 'Description').value = '';
        this.resetDropZone(photoType);
    }

    switchTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active', 'border-blue-500', 'text-blue-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });
        
        document.querySelector(`[data-tab="${tabName}"]`).classList.remove('border-transparent', 'text-gray-500');
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active', 'border-blue-500', 'text-blue-600');

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(tabName).classList.add('active');

        this.currentTab = tabName;

        // Load content for specific tabs
        if (tabName === 'queue') {
            this.updateQueue();
        } else if (tabName === 'history') {
            this.updateHistory();
        } else if (tabName === 'review') {
            this.updateReview();
        } else if (tabName === 'logs') {
            this.showLogs();
        }
    }

    switchSettingsTab(tabName) {
        // Update settings tab buttons
        document.querySelectorAll('.settings-tab').forEach(btn => {
            btn.classList.remove('active', 'border-blue-500', 'text-blue-600');
            btn.classList.add('border-transparent', 'text-gray-500');
        });
        
        document.querySelector(`[data-tab="${tabName}"]`).classList.remove('border-transparent', 'text-gray-500');
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active', 'border-blue-500', 'text-blue-600');

        // Update settings content
        document.querySelectorAll('.settings-content').forEach(content => {
            content.classList.remove('active');
            content.style.display = 'none';
        });
        const activeContent = document.getElementById(tabName + '-settings');
        activeContent.classList.add('active');
        activeContent.style.display = 'block';

        // Load specific content
        if (tabName === 'stocks') {
            this.loadStockConfigs();
        }
    }

    async openSettings() {
        await this.loadSettings();
        this.populateSettingsForm();
        
        const modal = document.getElementById('settingsModal');
        if (modal) {
            modal.classList.remove('hidden');
        }
        
        // Проверяем статус ExifTool
        this.checkExifToolStatus();
    }

    closeSettings() {
        document.getElementById('settingsModal').classList.add('hidden');
    }

    async checkExifToolStatus() {
        const statusEl = document.getElementById('exifToolStatus');
        const iconEl = document.getElementById('exifToolIcon');
        const messageEl = document.getElementById('exifToolMessage');
        
        if (!statusEl || !iconEl || !messageEl) return;
        
        try {
            if (this.isWailsMode) {
                const status = await window.go.main.App.CheckExifToolStatus();
                
                if (status.available) {
                    // ExifTool доступен
                    statusEl.className = 'mt-1 p-3 rounded-md border border-green-200 bg-green-50';
                    iconEl.className = 'fas fa-check-circle text-green-500 mr-2';
                    messageEl.textContent = status.message;
                } else {
                    // ExifTool недоступен
                    statusEl.className = 'mt-1 p-3 rounded-md border border-yellow-200 bg-yellow-50';
                    iconEl.className = 'fas fa-exclamation-triangle text-yellow-500 mr-2';
                    messageEl.textContent = status.message;
                }
            } else {
                // Mock режим - показываем предупреждение
                statusEl.className = 'mt-1 p-3 rounded-md border border-gray-200 bg-gray-50';
                iconEl.className = 'fas fa-info-circle text-gray-500 mr-2';
                messageEl.textContent = 'Демо режим - статус ExifTool недоступен';
            }
        } catch (error) {
            console.error('Error checking ExifTool status:', error);
            statusEl.className = 'mt-1 p-3 rounded-md border border-red-200 bg-red-50';
            iconEl.className = 'fas fa-times-circle text-red-500 mr-2';
            messageEl.textContent = 'Ошибка при проверке статуса ExifTool';
        }
    }

    async loadSettings() {
        try {
    
            if (this.isWailsMode) {
    
                this.settings = await window.go.main.App.GetSettings();
    
            } else {
    
                // В mock режиме создаем базовые настройки сразу
                this.settings = {
                    tempDirectory: './temp',
                    aiProvider: 'openai',
                    aiModel: 'gpt-4o',
                    thumbnailSize: 512,
                    maxConcurrentJobs: 3,
                    language: 'en',
                    aiPrompts: {
                        editorial: 'Default editorial prompt...',
                        commercial: 'Default commercial prompt...'
                    }
                };
            }
            
            // Устанавливаем язык из настроек и обновляем селекторы
            if (this.settings.language && this.settings.language !== window.i18n.getCurrentLanguage()) {
                await window.i18n.switchLanguage(this.settings.language);
            }
            
            // Принудительно обновляем селекторы языка
            this.updateLanguageSelectors(this.settings.language || window.i18n.getCurrentLanguage());
            
        } catch (error) {
            console.error('Error loading settings:', error);
            
            // Создаем fallback настройки
            this.settings = {
                tempDirectory: './temp',
                aiProvider: 'openai',
                aiModel: 'gpt-4o',
                thumbnailSize: 512,
                maxConcurrentJobs: 3,
                language: 'en',
                aiPrompts: {
                    editorial: 'Default editorial prompt...',
                    commercial: 'Default commercial prompt...'
                }
            };
            
            // Если настройки не загрузились, устанавливаем английский и обновляем селекторы
            if (this.isWailsMode) {
                // В Wails режиме показываем ошибку только если i18n уже инициализирован
                try {
                    this.showNotification(window.i18n.t('notifications.errorLoading'), 'error');
                } catch (i18nError) {
                    console.error('i18n error:', i18nError);
                    this.showNotification('Error loading settings', 'error');
                }
            }
            
            // Устанавливаем селекторы на английский
            this.updateLanguageSelectors('en');
        }
    }

    populateSettingsForm() {
        if (!this.settings) return;

        document.getElementById('tempDirectory').value = this.settings.tempDirectory || './temp';
        document.getElementById('thumbnailSize').value = this.settings.thumbnailSize || 512;
        document.getElementById('maxConcurrentJobs').value = this.settings.maxConcurrentJobs || 3;
        document.getElementById('aiProvider').value = this.settings.aiProvider || 'openai';
        document.getElementById('aiApiKey').value = this.settings.aiApiKey || '';
        document.getElementById('aiBaseUrl').value = this.settings.aiBaseUrl || '';
        document.getElementById('aiTimeout').value = this.settings.aiTimeout || 90;
        document.getElementById('aiMaxTokens').value = this.settings.aiMaxTokens || 2000;
        
        // Устанавливаем язык в селекторе
        const language = this.settings.language || window.i18n.getCurrentLanguage();
        document.getElementById('settingsLanguage').value = language;
        
        if (this.settings.aiPrompts) {
            document.getElementById('editorialPrompt').value = this.settings.aiPrompts.editorial || '';
            document.getElementById('commercialPrompt').value = this.settings.aiPrompts.commercial || '';
        }

        // Загружаем модели для текущего провайдера
        this.loadAIModels();
    }

    async onAIProviderChange(provider) {
        // Очищаем кастомный селектор при смене провайдера
        const input = document.getElementById('aiModelInput');
        input.value = '';
        
        this.currentModels = [];
        this.selectedModel = null;
        
        // Скрываем dropdown и очищаем описание модели
        this.hideAIModelDropdown();
        document.getElementById('modelDescription').textContent = '';
        
        // Автоматически загружаем модели для нового провайдера
        await this.loadAIModels();
    }

    onAIModelChange(modelId) {
        // Находим информацию о выбранной модели и показываем описание
        if (this.currentModels && this.currentModels.length > 0) {
            const model = this.currentModels.find(m => m.id === modelId);
            if (model) {
                const description = `${model.description} (Max tokens: ${model.maxTokens ? model.maxTokens.toLocaleString() : 'N/A'})`;
                document.getElementById('modelDescription').textContent = description;
            }
        } else if (this.selectedModel && this.selectedModel.isCustom) {
            document.getElementById('modelDescription').textContent = window.i18n.t('settings.ai.modelCustom');
        }
    }

    async loadAIModels() {
        if (!this.isWailsMode) {
            // В режиме демо загружаем mock модели
            this.loadMockModels();
            return;
        }

        const provider = document.getElementById('aiProvider').value;
        if (!provider) return;

        const loadBtn = document.getElementById('loadModelsBtn');
        const originalText = loadBtn.textContent;
        
        loadBtn.disabled = true;
        loadBtn.textContent = window.i18n.t('settings.ai.loadingModels');

        try {
            const models = await window.go.main.App.GetAIModels(provider);
            this.currentModels = models;
            
            // Отображаем модели в кастомном селекторе
            this.renderAIModelList(models);
            
            // Если есть сохраненная модель, выбираем её
            if (this.settings && this.settings.aiModel) {
                const savedModel = models.find(m => m.id === this.settings.aiModel);
                if (savedModel) {
                    this.selectAIModel(savedModel);
                } else {
                    // Если сохраненная модель не найдена, это может быть кастомная модель
                    this.selectAIModel({ 
                        id: this.settings.aiModel, 
                        name: this.settings.aiModel, 
                        description: window.i18n.t('settings.ai.modelCustom'), 
                        isCustom: true 
                    });
                }
            } else {
                // Выбираем первую модель по умолчанию
                if (models.length > 0) {
                    this.selectAIModel(models[0]);
                }
            }

            this.showNotification(window.i18n.t('settings.ai.modelsLoaded'), 'success');
        } catch (error) {
            console.error('Error loading AI models:', error);
            this.showNotification(window.i18n.t('settings.ai.modelsLoadFailed') + ': ' + error.message, 'error');
        } finally {
            loadBtn.disabled = false;
            loadBtn.textContent = originalText;
        }
    }

    // Загружает mock модели для демо режима
    loadMockModels() {
        const provider = document.getElementById('aiProvider').value;
        
        if (provider === 'openai') {
            this.currentModels = [
                {
                    id: 'o1',
                    name: 'o1',
                    description: 'Most advanced reasoning model for complex tasks',
                    maxTokens: 100000
                },
                {
                    id: 'o1-mini',
                    name: 'o1-mini',
                    description: 'Faster reasoning model for coding and math',
                    maxTokens: 65536
                },
                {
                    id: 'o1-preview',
                    name: 'o1-preview',
                    description: 'Preview of advanced reasoning capabilities',
                    maxTokens: 32768
                },
                {
                    id: 'o3-mini',
                    name: 'o3-mini (January 2025)',
                    description: 'Advanced reasoning model, successor to o1-mini',
                    maxTokens: 65536
                },
                {
                    id: 'gpt-4o',
                    name: 'GPT-4o',
                    description: 'High-intelligence flagship model for complex tasks',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4o-2024-11-20',
                    name: 'GPT-4o (November 2024)',
                    description: 'GPT-4o model with vision capabilities and enhanced performance',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4o-mini',
                    name: 'GPT-4o mini',
                    description: 'Affordable and intelligent small model for fast tasks',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4-turbo',
                    name: 'GPT-4 Turbo',
                    description: 'GPT-4 Turbo with enhanced capabilities and vision',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4.1',
                    name: 'GPT-4.1 (April 2025)',
                    description: 'Next-generation GPT-4 model with enhanced features',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4.5-preview',
                    name: 'GPT-4.5 (February 2025)',
                    description: 'Enhanced GPT-4 model with improved capabilities',
                    maxTokens: 128000
                },
                {
                    id: 'gpt-4',
                    name: 'GPT-4',
                    description: 'Advanced GPT-4 model with multimodal capabilities',
                    maxTokens: 8192
                }
            ];
        } else if (provider === 'claude') {
            this.currentModels = [
                {
                    id: 'claude-opus-4-20250514',
                    name: 'Claude Opus 4',
                    description: 'Most capable and intelligent model with superior reasoning capabilities',
                    maxTokens: 200000
                },
                {
                    id: 'claude-sonnet-4-20250514',
                    name: 'Claude Sonnet 4',
                    description: 'High-performance model with exceptional reasoning and efficiency',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-7-sonnet-20250219',
                    name: 'Claude 3.7 Sonnet',
                    description: 'High-performance model with early extended thinking capabilities',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-5-sonnet-20241022',
                    name: 'Claude 3.5 Sonnet v2',
                    description: 'Most intelligent model with enhanced vision capabilities (Latest)',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-5-sonnet-20240620',
                    name: 'Claude 3.5 Sonnet v1',
                    description: 'Original Claude 3.5 Sonnet with advanced capabilities',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-5-haiku-20241022',
                    name: 'Claude 3.5 Haiku',
                    description: 'Fastest model with vision capabilities and high intelligence',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-opus-20240229',
                    name: 'Claude 3 Opus',
                    description: 'Most powerful model for complex tasks with exceptional reasoning',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-sonnet-20240229',
                    name: 'Claude 3 Sonnet',
                    description: 'Balanced model for general tasks with strong performance',
                    maxTokens: 200000
                },
                {
                    id: 'claude-3-haiku-20240307',
                    name: 'Claude 3 Haiku',
                    description: 'Fast and cost-effective model for quick responses',
                    maxTokens: 200000
                }
            ];
        }
        
        this.renderAIModelList(this.currentModels);
        
        if (this.currentModels.length > 0) {
            this.selectAIModel(this.currentModels[0]);
        }
    }

    async saveSettings() {
        // Получаем выбранную модель из кастомного селектора
        const selectedModelId = this.selectedModel ? this.selectedModel.id : '';
        
        const settings = {
            tempDirectory: document.getElementById('tempDirectory').value,
            thumbnailSize: parseInt(document.getElementById('thumbnailSize').value),
            maxConcurrentJobs: parseInt(document.getElementById('maxConcurrentJobs').value),
            aiProvider: document.getElementById('aiProvider').value,
            aiModel: selectedModelId,
            aiApiKey: document.getElementById('aiApiKey').value,
            aiBaseUrl: document.getElementById('aiBaseUrl').value,
            aiTimeout: parseInt(document.getElementById('aiTimeout').value),
            aiMaxTokens: parseInt(document.getElementById('aiMaxTokens').value),
            language: document.getElementById('settingsLanguage').value,
            aiPrompts: {
                editorial: document.getElementById('editorialPrompt').value,
                commercial: document.getElementById('commercialPrompt').value
            }
        };

        try {
            await window.go.main.App.SaveSettings(settings);
            this.settings = settings;
            this.showNotification(window.i18n.t('settings.saved'), 'success');
            this.closeSettings();
        } catch (error) {
            console.error('Error saving settings:', error);
            this.showNotification(window.i18n.t('notifications.errorSaving') + ': ' + error.message, 'error');
        }
    }

    async testAIConnection() {
        const btn = document.getElementById('testAiConnectionBtn');
        btn.disabled = true;
        btn.textContent = window.i18n.t('settings.ai.testing');

        try {
            // Получаем текущие настройки из формы
            const settings = {
                aiProvider: document.getElementById('aiProvider').value,
                aiApiKey: document.getElementById('aiApiKey').value,
                aiBaseUrl: document.getElementById('aiBaseUrl').value
            };

            // В реальном приложении здесь был бы вызов TestAIConnection
            // await window.go.main.App.TestAIConnection(settings);
            
            // Временная заглушка
            await new Promise(resolve => setTimeout(resolve, 1000));
            
            this.showNotification(window.i18n.t('settings.ai.testSuccess'), 'success');
        } catch (error) {
            console.error('AI connection test failed:', error);
            this.showNotification(window.i18n.t('settings.ai.testFailed') + ': ' + error.message, 'error');
        } finally {
            btn.disabled = false;
            btn.textContent = window.i18n.t('settings.ai.testConnection');
        }
    }

    // resetPromptToDefault сбрасывает промпт к дефолтному значению
    resetPromptToDefault(promptType) {
        const defaultPrompts = {
            editorial: `Создай метаданные для редакционного стокового фото:

ТРЕБОВАНИЯ ДЛЯ EDITORIAL:
1. НАЗВАНИЕ (до 100 символов):
   - Фактическое описание события/сюжета
   - Конкретные имена людей и мест (если применимо)
   - Журналистский стиль без эмоциональной окраски
   - Временной контекст при необходимости

2. ОПИСАНИЕ (до 500 символов):
   - WHO: конкретные имена людей
   - WHAT: точное описание происходящего  
   - WHERE: конкретные места с полными названиями
   - WHEN: даты и время (используй EXIF данные)
   - WHY: контекст и причины события

3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
   - Конкретные имена публичных лиц (только если четко идентифицируемы)
   - Точные географические названия (если возможно определить)
   - Названия событий и организаций (если видны на фото)
   - Тематические категории: politics, economy, sports, entertainment, news
   - Временные маркеры: указывай конкретные даты если есть в EXIF
   - Контекст события: что происходит, кто участвует, где проходит
   - Визуальные элементы: что реально видно на изображении

4. КАТЕГОРИЯ (выбери одну из для Editorial):
   News, Politics, Current Events, Documentary, Entertainment, Celebrity, Sports Events, Business & Finance, Social Issues, War & Conflict, Disasters, Environment, Healthcare, Education, Crime, Religion, Royalty, Awards & Ceremonies

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ВКЛЮЧАЙ: точные названия мест, имена, организации, политические темы, контекст событий
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."], "category": "..."}`,

            commercial: `Создай метаданные для коммерческого стокового фото:

ТРЕБОВАНИЯ ДЛЯ COMMERCIAL:
1. НАЗВАНИЕ (до 70 символов):
   - Описательное без конкретных имен и мест
   - Концептуальное (бизнес, семья, технологии)
   - Эмоциональное состояние (счастье, успех)
   - Универсальные формулировки

2. ОПИСАНИЕ (строго до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов
   - ВАЖНО: Описание не должно превышать 200 символов

3. КЛЮЧЕВЫЕ СЛОВА (48-55 слов):
   АНАЛИЗИРУЙ ИЗОБРАЖЕНИЕ и создавай ключевые слова на основе того, что РЕАЛЬНО видишь:
   - Люди: возраст, пол, количество, роли (избегай конкретных имен)
   - Эмоции: какие эмоции выражают люди или передает изображение
   - Концепции: какие идеи, понятия иллюстрирует фото
   - Визуальные характеристики: стиль, цвета, композиция, освещение
   - Локации: тип места без конкретных названий (office, home, etc.)
   - Действия и активности: что происходит на фото
   - Объекты и предметы: что видишь на изображении
   - Настроение и атмосфера: общее впечатление от фото

4. КАТЕГОРИЯ (выбери одну из для Commercial):
   Business, Lifestyle, Nature, Technology, People, Family, Food & Drink, Fashion, Travel, Health & Wellness, Education, Sport & Fitness, Animals, Architecture, Music, Art & Design, Objects, Concepts, Beauty, Shopping, Transportation, Home & Garden

ВАЖНО: Все метаданные должны быть НА АНГЛИЙСКОМ ЯЗЫКЕ. Title, description и keywords - только английский!
ИЗБЕГАЙ: конкретные имена людей/компаний, бренды, конкретные места, даты
ФОРМАТ JSON: {"title": "...", "description": "...", "keywords": ["...", "..."], "category": "..."}`,
        };

        const promptElement = document.getElementById(promptType + 'Prompt');
        if (promptElement && defaultPrompts[promptType]) {
            promptElement.value = defaultPrompts[promptType];
            this.showNotification(`Промпт для ${promptType === 'editorial' ? 'редакционных' : 'коммерческих'} фото сброшен к умолчанию`, 'success');
        }
    }

    // forceUpdatePrompts принудительно обновляет дефолтные промпты в базе данных
    async forceUpdatePrompts() {
        if (!this.isWailsMode) {
            this.showNotification('Функция доступна только в полной версии приложения', 'warning');
            return;
        }

        const btn = document.getElementById('forceUpdatePromptsBtn');
        const originalText = btn.textContent;
        
        btn.disabled = true;
        btn.textContent = window.i18n.t('settings.ai.updatingPrompts');

        try {
            // Вызываем метод принудительного обновления промптов
            await window.go.main.App.ForceUpdateDefaultPrompts();
            
            // Перезагружаем настройки чтобы обновить UI
            await this.loadSettings();
            this.populateSettingsForm();
            
            this.showNotification(window.i18n.t('settings.ai.promptsUpdated'), 'success');
        } catch (error) {
            console.error('Error updating prompts:', error);
            this.showNotification('Ошибка обновления промптов: ' + error.message, 'error');
        } finally {
            btn.disabled = false;
            btn.textContent = originalText;
        }
    }

    async updateQueue() {
        try {
            // Сохраняем состояние открытых журналов перед обновлением
            const openLogs = this.getOpenLogsState();
            
            const queueStatus = await window.go.main.App.GetQueueStatus();
            this.renderQueue(queueStatus);
            
            // Восстанавливаем состояние открытых журналов после обновления
            this.restoreOpenLogsState(openLogs);
        } catch (error) {
            console.error('Error updating queue:', error);
            this.showNotification(window.i18n.t('notifications.errorQueue'), 'error');
        }
    }

    getOpenLogsState() {
        const openLogs = [];
        const logContainers = document.querySelectorAll('[id^="queue-logs-"]');
        
        logContainers.forEach(container => {
            if (!container.classList.contains('hidden')) {
                const batchId = container.id.replace('queue-logs-', '');
                openLogs.push(batchId);
            }
        });
        
        return openLogs;
    }

    restoreOpenLogsState(openLogs) {
        openLogs.forEach(batchId => {
            const logsContainer = document.getElementById(`queue-logs-${batchId}`);
            if (logsContainer) {
                logsContainer.classList.remove('hidden');
                // Загружаем логи для восстановленного контейнера
                this.loadQueueLogs(batchId);
            }
        });
    }

    renderQueue(queueStatus) {
        const container = document.getElementById('queueContainer');
        
        if (!queueStatus || queueStatus.length === 0) {
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-tasks text-3xl mb-2"></i>
                    <p>${window.i18n.t('queue.empty')}</p>
                </div>
            `;
            return;
        }

        const queueHTML = queueStatus.map(batch => `
            <div class="border border-gray-200 rounded-lg p-4 mb-4">
                <div class="flex justify-between items-start mb-2">
                    <div class="flex-1">
                        <h4 class="font-medium text-gray-900">${batch.description}</h4>
                        <p class="text-sm text-gray-500">${window.i18n.t('queue.type')}: ${batch.type} • ${window.i18n.t('queue.photos')}: ${batch.processedPhotos}/${batch.totalPhotos}</p>
                        ${batch.currentStep ? `<p class="text-sm text-blue-600">${window.i18n.t('queue.currentStep')}: ${batch.currentStep}</p>` : ''}
                        ${batch.photos && batch.photos.length > 0 ? (() => {
                            const stats = this.calculateBatchStats(batch.photos);
                            return `
                                <div class="flex flex-wrap gap-1 mt-1">
                                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                                        Всего: ${stats.total}
                                    </span>
                                    <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                        Готово: ${stats.completed}
                                    </span>
                                    ${stats.processing > 0 ? `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">В процессе: ${stats.processing}</span>` : ''}
                                    ${stats.waiting > 0 ? `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">Ожидает: ${stats.waiting}</span>` : ''}
                                    ${stats.failed > 0 ? `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">Ошибки: ${stats.failed}</span>` : ''}
                                </div>
                            `;
                        })() : ''}
                    </div>
                    <div class="flex items-center space-x-2">
                        <button onclick="app.toggleQueueLogs('${batch.batchId}')" 
                                class="text-gray-500 hover:text-blue-600 p-1 rounded" 
                                title="${window.i18n.t('queue.viewLogs') || 'Показать логи'}">
                            <i class="fas fa-list-alt"></i>
                        </button>
                        <span class="px-2 py-1 text-xs rounded-full ${this.getStatusBadgeClass(batch.status)}">
                            ${batch.status}
                        </span>
                    </div>
                </div>
                
                <div class="w-full bg-gray-200 rounded-full h-2 mb-2">
                    <div class="bg-blue-600 h-2 rounded-full progress-bar" style="width: ${batch.progress}%"></div>
                </div>
                
                <div class="text-sm text-gray-600 mb-3">
                    ${window.i18n.t('queue.progress')}: ${batch.progress}%
                    ${batch.currentPhoto ? `• ${window.i18n.t('queue.current')}: ${batch.currentPhoto}` : ''}
                    ${batch.error ? `• ${window.i18n.t('queue.error')}: ${batch.error}` : ''}
                </div>

                <!-- Детальные этапы -->
                ${this.renderQueueSteps(batch)}

                <!-- Детальная информация по фотографиям -->
                ${batch.photos && batch.photos.length > 0 ? this.renderPhotoProgress(batch.photos) : ''}

                <!-- Контейнер для логов -->
                <div id="queue-logs-${batch.batchId}" class="hidden mt-3 border-t pt-3">
                    <div class="flex items-center justify-between mb-2">
                        <h5 class="font-medium text-gray-800">${window.i18n.t('queue.detailedLogs') || 'Детальные логи'}</h5>
                        <button onclick="app.refreshQueueLogs('${batch.batchId}')" 
                                class="text-xs text-blue-600 hover:text-blue-800">
                            <i class="fas fa-sync mr-1"></i>${window.i18n.t('logs.refresh') || 'Обновить'}
                        </button>
                    </div>
                    <div id="queue-logs-content-${batch.batchId}" class="space-y-2 max-h-64 overflow-y-auto">
                        <!-- Логи будут загружены сюда -->
                    </div>
                </div>
            </div>
        `).join('');

        container.innerHTML = queueHTML;
    }

    renderQueueSteps(batch) {
        const steps = [
            { name: 'initialization', label: window.i18n.t('queue.steps.initialization') || 'Инициализация', icon: 'fa-play' },
            { name: 'ai_processing', label: window.i18n.t('queue.steps.aiProcessing') || 'AI обработка', icon: 'fa-brain' },
            { name: 'uploading', label: window.i18n.t('queue.steps.uploading') || 'Загрузка', icon: 'fa-upload' },
            { name: 'completed', label: window.i18n.t('queue.steps.completed') || 'Завершено', icon: 'fa-check' }
        ];

        let currentStepIndex = 0;
        
        // Определяем текущий шаг на основе статуса batch и currentStep
        if (batch.status === 'completed') {
            currentStepIndex = 3; // Завершено
        } else if (batch.status === 'processing') {
            // Определяем более точно на основе currentStep
            if (batch.currentStep === 'ai_processing' || batch.currentStep === 'ai_analysis' || batch.currentStep === 'preparation') {
                currentStepIndex = 1; // AI обработка
            } else if (batch.currentStep === 'uploading') {
                currentStepIndex = 2; // Загрузка
            } else {
                currentStepIndex = 1; // По умолчанию AI обработка
            }
        } else if (batch.status === 'pending' || batch.status === 'waiting') {
            currentStepIndex = 0; // Инициализация
        }

        // Debug: показываем информацию о текущем этапе в консоли только при разработке
        if (window.location.hostname === 'localhost') {
            console.log('Batch step debug:', { 
                batchId: batch.batchId, 
                status: batch.status, 
                currentStep: batch.currentStep, 
                calculatedIndex: currentStepIndex 
            });
        }

        const stepsHTML = steps.map((step, index) => {
            let stepClass = 'text-gray-400';
            let iconClass = 'text-gray-400';
            let bgClass = 'bg-gray-100';
            
            if (index < currentStepIndex) {
                // Завершенные этапы
                stepClass = 'text-green-600';
                iconClass = 'text-green-600';
                bgClass = 'bg-green-100';
            } else if (index === currentStepIndex) {
                // Текущий этап
                stepClass = 'text-blue-600';
                iconClass = 'text-blue-600';
                bgClass = 'bg-blue-100';
            }

            return `
                <div class="flex items-center space-x-1 px-2 py-1 rounded ${bgClass} ${stepClass}">
                    <i class="fas ${step.icon} ${iconClass}"></i>
                    <span class="text-xs font-medium">${step.label}</span>
                </div>
                ${index < steps.length - 1 ? '<i class="fas fa-chevron-right text-gray-400 mx-1"></i>' : ''}
            `;
        }).join('');

        return `
            <div class="mb-3 p-2 bg-gray-50 rounded border">
                <div class="flex items-center space-x-1 text-xs">
                    ${stepsHTML}
                </div>
            </div>
        `;
    }

    renderPhotoProgress(photos) {
        if (!photos || photos.length === 0) return '';

        const photosHTML = photos.map(photo => {
            const statusIcon = this.getPhotoProgressIcon(photo.status);
            const statusColor = this.getPhotoProgressColor(photo.status);
            const stepName = this.getStepDisplayName(photo.step);

            return `
                <div class="flex items-center justify-between py-1 px-2 text-xs border-b border-gray-100 last:border-b-0">
                    <div class="flex items-center space-x-2 flex-1 min-w-0">
                        <span class="text-sm">${statusIcon}</span>
                        <span class="truncate font-medium">${photo.fileName}</span>
                        <span class="text-gray-500">${stepName}</span>
                    </div>
                    <div class="flex items-center space-x-2">
                        <div class="w-16 bg-gray-200 rounded-full h-1">
                            <div class="bg-blue-600 h-1 rounded-full transition-all duration-300" style="width: ${photo.progress}%"></div>
                        </div>
                        <span class="text-xs text-gray-500 w-8 text-right">${photo.progress}%</span>
                        ${photo.error ? `<i class="fas fa-exclamation-triangle text-red-500" title="${photo.error}"></i>` : ''}
                    </div>
                </div>
            `;
        }).join('');

        return `
            <div class="mt-3 border border-gray-200 rounded">
                <div class="bg-gray-50 px-3 py-2 border-b border-gray-200">
                    <h5 class="text-sm font-medium text-gray-700">Прогресс по фотографиям</h5>
                </div>
                <div class="max-h-32 overflow-y-auto">
                    ${photosHTML}
                </div>
            </div>
        `;
    }

    getPhotoProgressIcon(status) {
        switch (status) {
            case 'completed': return '✅';
            case 'failed': return '❌';
            case 'processing': return '🔄';
            case 'pending': return '⏳';
            default: return '📝';
        }
    }

    getPhotoProgressColor(status) {
        switch (status) {
            case 'completed': return 'text-green-600';
            case 'failed': return 'text-red-600';
            case 'processing': return 'text-blue-600';
            case 'pending': return 'text-gray-500';
            default: return 'text-gray-600';
        }
    }

    getStepDisplayName(step) {
        const stepNames = {
            'waiting': 'Ожидание',
            'preparation': 'Подготовка',
            'ai_analysis': 'AI анализ',
            'saving': 'Сохранение',
            'exif_writing': 'Запись EXIF',
            'completed': 'Завершено'
        };
        return stepNames[step] || step;
    }

    calculateBatchStats(photos) {
        const stats = {
            total: photos.length,
            completed: 0,
            processing: 0,
            waiting: 0,
            failed: 0
        };

        photos.forEach(photo => {
            switch (photo.status) {
                case 'completed':
                    stats.completed++;
                    break;
                case 'processing':
                    stats.processing++;
                    break;
                case 'failed':
                    stats.failed++;
                    break;
                default:
                    stats.waiting++;
                    break;
            }
        });

        return stats;
    }

    async toggleQueueLogs(batchId) {
        const logsContainer = document.getElementById(`queue-logs-${batchId}`);
        
        if (logsContainer.classList.contains('hidden')) {
            logsContainer.classList.remove('hidden');
            await this.loadQueueLogs(batchId);
        } else {
            logsContainer.classList.add('hidden');
        }
    }

    async loadQueueLogs(batchId) {
        const logsContent = document.getElementById(`queue-logs-content-${batchId}`);
        
        try {
            logsContent.innerHTML = '<div class="text-center py-2 text-gray-500"><i class="fas fa-spinner fa-spin mr-2"></i>Загрузка логов...</div>';
            
            const [events, progress] = await Promise.all([
                window.go.main.App.GetBatchEvents(batchId, 20),
                window.go.main.App.GetProcessingProgress(batchId)
            ]);

            if (events.length === 0) {
                logsContent.innerHTML = '<div class="text-center py-2 text-gray-500">Нет доступных логов</div>';
                return;
            }

            const logsHTML = events.slice(-10).map(event => {
                const timeStr = new Date(event.createdAt).toLocaleTimeString();
                const statusIcon = this.getEventStatusIcon(event.status);
                const statusColor = this.getEventStatusColor(event.status);
                
                return `
                    <div class="flex items-start space-x-2 p-2 bg-white rounded border text-xs">
                        <span class="text-sm">${statusIcon}</span>
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center justify-between">
                                <span class="font-medium ${statusColor}">${event.message}</span>
                                <span class="text-gray-400 text-xs">${timeStr}</span>
                            </div>
                            ${event.details ? `
                                <div class="mt-1 text-gray-600 text-xs truncate" title="${event.details}">
                                    ${event.details.length > 80 ? event.details.substring(0, 80) + '...' : event.details}
                                </div>
                            ` : ''}
                        </div>
                    </div>
                `;
            }).join('');

            logsContent.innerHTML = logsHTML;
        } catch (error) {
            console.error('Error loading queue logs:', error);
            logsContent.innerHTML = '<div class="text-center py-2 text-red-500">Ошибка загрузки логов</div>';
        }
    }

    async refreshQueueLogs(batchId) {
        await this.loadQueueLogs(batchId);
    }

    getEventStatusIcon(status) {
        switch (status) {
            case 'success': return '✅';
            case 'failed': return '❌';
            case 'started': return '🚀';
            case 'progress': return '⏳';
            default: return '📝';
        }
    }

    getEventStatusColor(status) {
        switch (status) {
            case 'success': return 'text-green-600';
            case 'failed': return 'text-red-600';
            case 'started': return 'text-blue-600';
            case 'progress': return 'text-yellow-600';
            default: return 'text-gray-600';
        }
    }

    async updateHistory() {
        try {
            const history = await window.go.main.App.GetProcessingHistory(20);
            this.renderHistory(history);
        } catch (error) {
            console.error('Error updating history:', error);
            this.showNotification(window.i18n.t('notifications.errorHistory'), 'error');
        }
    }

    renderHistory(history) {
        const container = document.getElementById('historyContainer');
        
        if (!history || history.length === 0) {
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-history text-3xl mb-2"></i>
                    <p>${window.i18n.t('history.empty')}</p>
                </div>
            `;
            return;
        }

        const historyHTML = history.map(batch => `
            <div class="border border-gray-200 rounded-lg p-4 mb-4">
                <div class="flex justify-between items-start">
                    <div>
                        <h4 class="font-medium text-gray-900">${batch.description}</h4>
                        <p class="text-sm text-gray-500">
                            ${window.i18n.t('queue.type')}: ${batch.type} • ${batch.photos.length} ${window.i18n.t('queue.photos')} • 
                            ${new Date(batch.createdAt).toLocaleDateString()}
                        </p>
                    </div>
                    <span class="px-2 py-1 text-xs rounded-full ${this.getStatusBadgeClass(batch.status)}">
                        ${batch.status}
                    </span>
                </div>
                
                <div class="mt-2">
                    <button class="text-blue-600 text-sm hover:text-blue-800" onclick="app.toggleBatchDetails('${batch.id}')">
                        <i class="fas fa-chevron-down mr-1"></i>${window.i18n.t('history.viewDetails') || 'View Details'}
                    </button>
                    <div id="details-${batch.id}" class="hidden mt-2 text-sm text-gray-600">
                        <p>Folder: ${batch.folderPath}</p>
                        <!-- Дополнительные детали можно добавить здесь -->
                    </div>
                </div>
            </div>
        `).join('');

        container.innerHTML = historyHTML;
    }

    // Review methods
    async updateReview() {
        try {
            const batches = await window.go.main.App.GetProcessedBatches();
            this.renderBatchSelector(batches);
        } catch (error) {
            console.error('Error loading processed batches:', error);
            this.showNotification(window.i18n.t('notifications.errorLoading'), 'error');
        }
    }

    renderBatchSelector(batches) {
        const selector = document.getElementById('batchSelector');
        
        // Очищаем селектор
        selector.innerHTML = `<option value="" data-i18n="review.selectBatchPlaceholder">Select a processed batch...</option>`;
        
        if (!batches || batches.length === 0) {
            // Показываем сообщение что нет батчей для ревью
            const container = document.getElementById('reviewContainer');
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-images text-3xl mb-2"></i>
                    <p data-i18n="review.noBatches">No processed batches available for review</p>
                </div>
            `;
            return;
        }

        // Добавляем батчи в селектор
        batches.forEach(batch => {
            const option = document.createElement('option');
            option.value = batch.id;
            
            const stats = batch.photosStats || { total: 0, processed: 0, approved: 0, rejected: 0 };
            const statusText = `${batch.type.toUpperCase()} - ${stats.total} photos (${stats.approved} approved, ${stats.rejected} rejected)`;
            
            option.textContent = `${batch.description || 'Batch ' + batch.id} - ${statusText}`;
            selector.appendChild(option);
        });

        // Обновляем переводы
        window.i18n.updateInterface();
        
        // Показываем кнопки управления если есть батчи
        this.updateBatchActions(batches.length > 0 ? batches[0].id : '');
    }

    async loadBatchForReview(batchId) {
        if (!batchId) {
            // Очищаем контейнер если батч не выбран
            const container = document.getElementById('reviewContainer');
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-images text-3xl mb-2"></i>
                    <p data-i18n="review.empty">Select a batch to review results</p>
                </div>
            `;
            window.i18n.updateInterface();
            return;
        }

        try {
            const photos = await window.go.main.App.GetBatchPhotos(batchId);
            this.renderPhotosForReview(photos);
            this.updateBatchActionsIfExists(batchId);
        } catch (error) {
            console.error('Error loading batch photos:', error);
            this.showNotification(window.i18n.t('notifications.errorLoading'), 'error');
        }
    }

    renderPhotosForReview(photos) {
        const container = document.getElementById('reviewContainer');
        
        if (!photos || photos.length === 0) {
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-images text-3xl mb-2"></i>
                    <p data-i18n="review.noPhotos">No photos found in this batch</p>
                </div>
            `;
            window.i18n.updateInterface();
            return;
        }

        const photosHTML = photos.map(photo => {
            const aiResult = photo.aiResult;
            const hasAI = aiResult && aiResult.title;
            
            // Определяем статус фото
            const statusClass = this.getPhotoStatusClass(photo.status);
            const statusIcon = this.getPhotoStatusIcon(photo.status);
            
            return `
                <div class="bg-white rounded-lg shadow-md overflow-hidden">
                    <!-- Изображение -->
                    <div class="relative">
                        <img id="thumb-${photo.id}" 
                             alt="${photo.fileName}" 
                             class="w-full h-48 object-cover"
                             src="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjNmNGY2Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzZiNzI4MCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk5vIEltYWdlPC90ZXh0Pjwvc3ZnPg==">
                        
                        <!-- Статус фото -->
                        <div class="absolute top-2 right-2">
                            <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusClass}">
                                <i class="${statusIcon} mr-1"></i>
                                ${photo.status}
                            </span>
                        </div>
                        
                        <!-- Чекбокс для выбора фотографии -->
                        <div class="absolute top-2 left-2">
                            <label class="flex items-center">
                                <input type="checkbox" 
                                       class="photo-select-checkbox w-4 h-4 text-blue-600 bg-white border-gray-300 rounded focus:ring-blue-500" 
                                       data-photo-id="${photo.id}"
                                       ${photo.selectedForUpload ? 'checked' : ''}
                                       onchange="window.app.togglePhotoSelection('${photo.id}', this.checked)">
                                <span class="ml-1 text-white text-xs bg-black bg-opacity-50 px-1 rounded">Select</span>
                            </label>
                        </div>
                    </div>

                    <!-- Информация о фото -->
                    <div class="p-4">
                        <h3 class="font-medium text-gray-900 mb-2">${photo.fileName}</h3>
                        
                        ${hasAI ? `
                            <!-- AI результаты -->
                            <div class="space-y-3">
                                <!-- Заголовок -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
                                    <input type="text" 
                                           value="${this.escapeHtml(aiResult.title)}" 
                                           class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                           onchange="window.app.updatePhotoTitle('${photo.id}', this.value)">
                                </div>

                                <!-- Описание -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
                                    <textarea rows="3" 
                                              class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                              onchange="window.app.updatePhotoDescription('${photo.id}', this.value)">${this.escapeHtml(aiResult.description)}</textarea>
                                </div>

                                <!-- Ключевые слова -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Keywords</label>
                                    <input type="text" 
                                           value="${aiResult.keywords ? aiResult.keywords.join(', ') : ''}" 
                                           class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                           placeholder="keyword1, keyword2, keyword3"
                                           onchange="window.app.updatePhotoKeywords('${photo.id}', this.value)">
                                </div>

                                <!-- Категория и качество -->
                                <div class="grid grid-cols-2 gap-3">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
                                        <select class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                                onchange="window.app.updatePhotoCategory('${photo.id}', this.value)">
                                            <option value="">Select category...</option>
                                            ${this.getCategoryOptions(aiResult.category || '', photo.contentType)}
                                        </select>
                                    </div>
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 mb-1">Quality</label>
                                        <input type="number" 
                                               value="${aiResult.quality || 0}" 
                                               min="0" max="10"
                                               class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                               onchange="window.app.updatePhotoQuality('${photo.id}', this.value)">
                                    </div>
                                </div>
                            </div>
                        ` : `
                            <div class="text-center py-4 text-gray-500">
                                <i class="fas fa-robot text-2xl mb-2"></i>
                                <p class="text-sm">No AI analysis available</p>
                            </div>
                        `}

                        <!-- Действия -->
                        <div class="mt-4 flex justify-between items-center">
                            <div class="flex space-x-2">
                                <button onclick="window.app.approvePhoto('${photo.id}')" 
                                        class="px-3 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700 ${photo.status === 'approved' ? 'opacity-50' : ''}"
                                        ${photo.status === 'approved' ? 'disabled' : ''}>
                                    <i class="fas fa-check mr-1"></i>Approve
                                </button>
                                <button onclick="window.app.rejectPhoto('${photo.id}')" 
                                        class="px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700 ${photo.status === 'rejected' ? 'opacity-50' : ''}"
                                        ${photo.status === 'rejected' ? 'disabled' : ''}>
                                    <i class="fas fa-times mr-1"></i>Reject
                                </button>
                            </div>
                            <div class="flex space-x-2">
                                <button id="regenerateBtn-${photo.id}" 
                                        class="px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed">
                                    <i id="regenerateIcon-${photo.id}" class="fas fa-sync-alt mr-1"></i>Regenerate
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            `;
        }).join('');

        container.innerHTML = `
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                ${photosHTML}
            </div>
        `;
        
        // Загружаем thumbnails асинхронно
        this.loadThumbnails(photos);
        
        // Добавляем event listeners для кнопок Regenerate
        this.setupRegenerateButtons(photos);
    }

    async loadThumbnails(photos) {
        for (const photo of photos) {
            try {
                const thumbImg = document.getElementById(`thumb-${photo.id}`);
                if (!thumbImg) continue;
                
                // Получаем thumbnail через API
                if (this.isWailsMode) {
                    const thumbnailData = await window.go.main.App.GetPhotoThumbnail(photo.id);
                    thumbImg.src = thumbnailData;
                } else {
                    // Mock - используем заглушку
                    const mockImg = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZTVlN2ViIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxMiIgZmlsbD0iIzZiNzI4MCIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk1vY2sgSW1hZ2U8L3RleHQ+PC9zdmc+';
                    thumbImg.src = mockImg;
                }
            } catch (error) {
                console.error(`Error loading thumbnail for ${photo.id}:`, error);
                // Thumbnail остается как заглушка
            }
        }
    }

    getPhotoStatusClass(status) {
        const classes = {
            'pending': 'bg-yellow-100 text-yellow-800',
            'processing': 'bg-blue-100 text-blue-800',
            'processed': 'bg-purple-100 text-purple-800',
            'approved': 'bg-green-100 text-green-800',
            'rejected': 'bg-red-100 text-red-800',
            'failed': 'bg-red-100 text-red-800'
        };
        return classes[status] || 'bg-gray-100 text-gray-800';
    }

    getPhotoStatusIcon(status) {
        const icons = {
            'pending': 'fas fa-clock',
            'processing': 'fas fa-spinner fa-spin',
            'processed': 'fas fa-check-circle',
            'approved': 'fas fa-thumbs-up',
            'rejected': 'fas fa-thumbs-down',
            'failed': 'fas fa-exclamation-triangle'
        };
        return icons[status] || 'fas fa-question-circle';
    }

    escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Получить опции категорий для выпадающего списка
    getCategoryOptions(selectedCategory, contentType) {
        // Категории для коммерческих фото (для рекламы, маркетинга, иллюстраций)
        const commercialCategories = [
            'Business', 'Lifestyle', 'Nature', 'Technology', 'People', 'Family',
            'Food & Drink', 'Fashion', 'Travel', 'Health & Wellness', 'Education',
            'Sport & Fitness', 'Animals', 'Architecture', 'Music', 'Art & Design',
            'Objects', 'Concepts', 'Beauty', 'Shopping', 'Transportation', 'Home & Garden'
        ];
        
        // Категории для редакционных фото (для новостей, документалистики)
        const editorialCategories = [
            'News', 'Politics', 'Current Events', 'Documentary', 'Entertainment',
            'Celebrity', 'Sports Events', 'Business & Finance', 'Social Issues',
            'War & Conflict', 'Disasters', 'Environment', 'Healthcare',
            'Education', 'Crime', 'Religion', 'Royalty', 'Awards & Ceremonies'
        ];
        
        let categories = contentType === 'editorial' ? editorialCategories : commercialCategories;
        
        return categories.map(category => 
            `<option value="${category}" ${category === selectedCategory ? 'selected' : ''}>${category}</option>`
        ).join('');
    }

    // Photo action methods
    async approvePhoto(photoId) {
        const approveBtn = document.querySelector(`button[onclick*="approvePhoto('${photoId}')"]`);
        const originalText = approveBtn ? approveBtn.innerHTML : '';
        const originalClasses = approveBtn ? approveBtn.className : '';
        
        try {
            if (approveBtn) {
                this.setActionButtonState(approveBtn, 'loading', 'Approving...');
            }
            
            // Показываем прогресс
            this.showNotification(window.i18n.t('review.approvingPhoto') || 'Approving photo and writing EXIF metadata...', 'info');
            
            await window.go.main.App.ApprovePhoto(photoId);
            
            if (approveBtn) {
                this.setActionButtonState(approveBtn, 'success', 'Approved!');
            }
            
            this.showNotification(window.i18n.t('review.photoApprovedWithExif') || 'Photo approved and EXIF metadata written to original file', 'success');
            
            await this.delay(600);
            
            // Обновляем текущий батч
            const batchId = document.getElementById('batchSelector').value;
            if (batchId) {
                this.loadBatchForReview(batchId);
            }
        } catch (error) {
            console.error('Error approving photo:', error);
            
            if (approveBtn) {
                this.setActionButtonState(approveBtn, 'error', 'Failed');
                await this.delay(1000);
                approveBtn.disabled = false;
                approveBtn.className = originalClasses;
                approveBtn.innerHTML = originalText;
            }
            
            this.showNotification('Error approving photo: ' + error.message, 'error');
        }
    }

    async rejectPhoto(photoId) {
        const rejectBtn = document.querySelector(`button[onclick*="rejectPhoto('${photoId}')"]`);
        const originalText = rejectBtn ? rejectBtn.innerHTML : '';
        const originalClasses = rejectBtn ? rejectBtn.className : '';
        
        try {
            if (rejectBtn) {
                this.setActionButtonState(rejectBtn, 'loading', 'Rejecting...');
            }
            
            await window.go.main.App.RejectPhoto(photoId);
            
            if (rejectBtn) {
                this.setActionButtonState(rejectBtn, 'success', 'Rejected!');
            }
            
            this.showNotification('Photo rejected', 'success');
            
            await this.delay(600);
            
            // Обновляем текущий батч
            const batchId = document.getElementById('batchSelector').value;
            if (batchId) {
                this.loadBatchForReview(batchId);
            }
        } catch (error) {
            console.error('Error rejecting photo:', error);
            
            if (rejectBtn) {
                this.setActionButtonState(rejectBtn, 'error', 'Failed');
                await this.delay(1000);
                rejectBtn.disabled = false;
                rejectBtn.className = originalClasses;
                rejectBtn.innerHTML = originalText;
            }
            
            this.showNotification('Error rejecting photo: ' + error.message, 'error');
        }
    }

    async updatePhotoTitle(photoId, title) {
        try {
            // Получаем текущие AI результаты и обновляем title
            await this.updatePhotoAIField(photoId, 'title', title);
        } catch (error) {
            console.error('Error updating photo title:', error);
            this.showNotification('Error updating title: ' + error.message, 'error');
        }
    }

    async updatePhotoDescription(photoId, description) {
        try {
            await this.updatePhotoAIField(photoId, 'description', description);
        } catch (error) {
            console.error('Error updating photo description:', error);
            this.showNotification('Error updating description: ' + error.message, 'error');
        }
    }

    async updatePhotoKeywords(photoId, keywordsString) {
        try {
            const keywords = keywordsString.split(',').map(k => k.trim()).filter(k => k.length > 0);
            await this.updatePhotoAIField(photoId, 'keywords', keywords);
        } catch (error) {
            console.error('Error updating photo keywords:', error);
            this.showNotification('Error updating keywords: ' + error.message, 'error');
        }
    }

    async updatePhotoCategory(photoId, category) {
        try {
            await this.updatePhotoAIField(photoId, 'category', category);
        } catch (error) {
            console.error('Error updating photo category:', error);
            this.showNotification('Error updating category: ' + error.message, 'error');
        }
    }

    async updatePhotoQuality(photoId, quality) {
        try {
            await this.updatePhotoAIField(photoId, 'quality', parseInt(quality));
        } catch (error) {
            console.error('Error updating photo quality:', error);
            this.showNotification('Error updating quality: ' + error.message, 'error');
        }
    }

    // Переключение выбора фотографии для загрузки (делегируем в UploadManager)
    async togglePhotoSelection(photoId, selected) {
        if (this.uploadManager) {
            await this.uploadManager.togglePhotoSelection(photoId, selected);
        }
    }

    async updatePhotoAIField(photoId, field, value) {
        // Получаем текущие фото из батча
        const batchId = document.getElementById('batchSelector').value;
        if (!batchId) return;

        const photos = await window.go.main.App.GetBatchPhotos(batchId);
        const photo = photos.find(p => p.id === photoId);
        if (!photo || !photo.aiResult) return;

        // Обновляем поле
        const updatedAIResult = { ...photo.aiResult };
        updatedAIResult[field] = value;

        // Сохраняем обновленные результаты
        await window.go.main.App.UpdatePhotoMetadata(photoId, updatedAIResult);
        
        // Если фото было approved, переводим его обратно в processed
        // чтобы пользователь мог снова нажать approve для записи EXIF
        if (photo.status === 'approved') {
            await this.setPhotoStatus(photoId, 'processed');
            
            // Перезагружаем батч для обновления UI
            this.loadBatchForReview(batchId);
            
            this.showNotification(window.i18n.t('review.metadataUpdatedReapprove') || 'Metadata updated. Click Approve again to write EXIF to file.', 'info');
        } else {
            // Для фото в других статусах просто показываем успешное сохранение
            this.showNotification(window.i18n.t('review.metadataSaved') || 'Metadata saved successfully', 'success');
        }
    }

    // Вспомогательная функция для изменения статуса фото
    async setPhotoStatus(photoId, status) {
        try {
            if (this.isWailsMode) {
                await window.go.main.App.SetPhotoStatus(photoId, status);
            } else {
                // Mock режим
                console.log(`Mock: Setting photo ${photoId} status to ${status}`);
            }
        } catch (error) {
            console.error('Error setting photo status:', error);
            throw error;
        }
    }

    async regeneratePhotoMetadata(photoId) {
        const regenerateBtn = document.getElementById(`regenerateBtn-${photoId}`);
        const regenerateIcon = document.getElementById(`regenerateIcon-${photoId}`);
        
        if (!regenerateBtn) {
            console.error('Regenerate button not found for photo:', photoId);
            alert('Regenerate button not found! Check console for details.');
            return;
        }
        
        if (regenerateBtn.disabled) {
            return; // Уже выполняется
        }

        // Сохраняем оригинальное состояние кнопки
        const originalText = regenerateBtn.innerHTML;
        const originalClasses = regenerateBtn.className;

        // Получаем текущие данные фотографии
        const photoData = await this.getPhotoData(photoId);
        
        // Показываем диалог с текущими данными и возможностью добавить комментарий
        this.showRegenerateDialog(photoData, async (correctionComment) => {
                try {
                    // Этап 1: Начинаем процесс
                    this.setRegenerateButtonState(regenerateBtn, regenerateIcon, 'loading', 'Preparing...');
                    await this.delay(300); // Небольшая задержка для плавности

                    // Этап 2: Анализируем изображение
                    this.setRegenerateButtonState(regenerateBtn, regenerateIcon, 'processing', 'Analyzing...');
                    
                    if (this.isWailsMode) {
                        await window.go.main.App.RegeneratePhotoMetadata(photoId, correctionComment || '');
                    } else {
                        // Mock режим с имитацией этапов
                        await this.delay(1000);
                        this.setRegenerateButtonState(regenerateBtn, regenerateIcon, 'processing', 'Generating...');
                        await this.delay(1500);
                    }
                    
                    // Этап 3: Успешное завершение
                    this.setRegenerateButtonState(regenerateBtn, regenerateIcon, 'success', 'Complete!');
                    this.showNotification('Metadata regenerated successfully', 'success');
                    
                    // Ждем немного, чтобы показать успех
                    await this.delay(800);
                    
                    // Обновляем текущий батч
                    const batchId = document.getElementById('batchSelector').value;
                    if (batchId) {
                        this.loadBatchForReview(batchId);
                    }
                    
                } catch (error) {
                    console.error('Error regenerating metadata:', error);
                    
                    // Этап 3: Показываем ошибку
                    this.setRegenerateButtonState(regenerateBtn, regenerateIcon, 'error', 'Failed');
                    this.showNotification('Error regenerating metadata: ' + error.message, 'error');
                    
                    // Ждем немного, чтобы показать ошибку
                    await this.delay(1500);
                    
                } finally {
                    // Восстанавливаем кнопку (проверяем, что элементы еще существуют)
                    const currentRegenerateBtn = document.getElementById(`regenerateBtn-${photoId}`);
                    const currentRegenerateIcon = document.getElementById(`regenerateIcon-${photoId}`);
                    
                    if (currentRegenerateBtn) {
                        currentRegenerateBtn.disabled = false;
                        currentRegenerateBtn.className = originalClasses;
                        currentRegenerateBtn.innerHTML = originalText;
                    }
                }
            },
            () => {
                // Пользователь отменил
                this.showNotification('Metadata regeneration cancelled', 'info');
            }
        );
    }

    // Вспомогательная функция для установки состояния кнопки Regenerate
    setRegenerateButtonState(button, icon, state, text) {
        if (!button || !icon) {
            console.error('Button or icon not found in setRegenerateButtonState');
            return;
        }
        
        // Удаляем все классы состояний
        button.className = button.className.replace(/bg-\w+-\d+/g, '').replace(/hover:bg-\w+-\d+/g, '');
        button.classList.remove('animate-pulse', 'animate-bounce', 'transform', 'scale-105');
        
        switch (state) {
            case 'loading':
                button.className += ' bg-yellow-500 hover:bg-yellow-600 animate-pulse';
                icon.className = 'fas fa-spinner fa-spin mr-1';
                button.innerHTML = `<i class="fas fa-spinner fa-spin mr-1"></i>${text}`;
                break;
                
            case 'processing':
                button.className += ' bg-blue-500 hover:bg-blue-600 animate-pulse';
                icon.className = 'fas fa-cog fa-spin mr-1';
                button.innerHTML = `<i class="fas fa-cog fa-spin mr-1"></i>${text}`;
                break;
                
            case 'success':
                button.className += ' bg-green-500 hover:bg-green-600 transform scale-105';
                icon.className = 'fas fa-check mr-1';
                button.innerHTML = `<i class="fas fa-check mr-1"></i>${text}`;
                // Добавляем анимацию успеха
                setTimeout(() => {
                    if (button) {
                        button.classList.add('animate-bounce');
                    }
                }, 100);
                break;
                
            case 'error':
                button.className += ' bg-red-500 hover:bg-red-600 animate-pulse';
                icon.className = 'fas fa-exclamation-triangle mr-1';
                button.innerHTML = `<i class="fas fa-exclamation-triangle mr-1"></i>${text}`;
                break;
        }
        
        button.disabled = true;
    }

    // Вспомогательная функция для установки состояния кнопок действий (Approve/Reject)
    setActionButtonState(button, state, text) {
        if (!button) return;
        
        // Сохраняем базовые классы кнопки
        const baseClasses = button.className.split(' ').filter(cls => 
            !cls.startsWith('bg-') && !cls.startsWith('hover:bg-') && 
            !cls.includes('animate-') && !cls.includes('transform') && !cls.includes('scale-')
        ).join(' ');
        
        button.classList.remove('animate-pulse', 'animate-bounce', 'transform', 'scale-105');
        
        switch (state) {
            case 'loading':
                button.className = baseClasses + ' bg-gray-500 hover:bg-gray-600 animate-pulse';
                button.innerHTML = `<i class="fas fa-spinner fa-spin mr-1"></i>${text}`;
                break;
                
            case 'success':
                if (text.includes('Approved')) {
                    button.className = baseClasses + ' bg-green-600 hover:bg-green-700 transform scale-105';
                } else {
                    button.className = baseClasses + ' bg-orange-600 hover:bg-orange-700 transform scale-105';
                }
                button.innerHTML = `<i class="fas fa-check mr-1"></i>${text}`;
                setTimeout(() => {
                    if (button) {
                        button.classList.add('animate-bounce');
                    }
                }, 100);
                break;
                
            case 'error':
                button.className = baseClasses + ' bg-red-600 hover:bg-red-700 animate-pulse';
                button.innerHTML = `<i class="fas fa-exclamation-triangle mr-1"></i>${text}`;
                break;
        }
        
        button.disabled = true;
    }

    // Настройка event listeners для кнопок Regenerate
    setupRegenerateButtons(photos) {
        photos.forEach(photo => {
            const regenerateBtn = document.getElementById(`regenerateBtn-${photo.id}`);
            if (regenerateBtn) {
                console.log('Setting up event listener for regenerate button:', photo.id);
                
                // Удаляем старые обработчики
                regenerateBtn.onclick = null;
                
                // Добавляем новый обработчик
                regenerateBtn.addEventListener('click', (e) => {
                    e.preventDefault();
                    console.log('Event listener triggered for photo:', photo.id);
                    this.regeneratePhotoMetadata(photo.id);
                });
            } else {
                console.warn('Regenerate button not found for photo:', photo.id);
            }
        });
    }

    // Вспомогательная функция для задержки
    delay(ms) {
        console.log(`Starting delay for ${ms}ms`);
        return new Promise(resolve => {
            setTimeout(() => {
                console.log(`Delay of ${ms}ms completed`);
                resolve();
            }, ms);
        });
    }

    // Получение данных фотографии для показа в диалоге
    async getPhotoData(photoId) {
        try {
            if (this.isWailsMode) {
                const photo = await window.go.main.App.GetPhoto(photoId);
                console.log('Got photo data from API:', photo);
                return photo;
            } else {
                // Mock данные для демо режима
                return {
                    id: photoId,
                    aiResult: {
                        title: 'Modern office building with glass facade',
                        description: 'Contemporary architectural photography showing a modern office building with reflective glass windows during daylight hours. The building features clean geometric lines and represents modern urban development.',
                        keywords: ['architecture', 'office building', 'modern', 'glass facade', 'urban', 'contemporary', 'business', 'corporate', 'daylight'],
                        category: 'Architecture',
                        quality: 8
                    }
                };
            }
        } catch (error) {
            console.error('Error getting photo data:', error);
            
            // В случае ошибки, пытаемся получить данные из DOM
            const photoCard = document.querySelector(`[data-photo-id="${photoId}"]`);
            if (photoCard) {
                const titleEl = photoCard.querySelector('.photo-title');
                const descEl = photoCard.querySelector('.photo-description');
                const keywordsEl = photoCard.querySelector('.photo-keywords');
                const categoryEl = photoCard.querySelector('.photo-category');
                
                return {
                    id: photoId,
                    aiResult: {
                        title: titleEl ? titleEl.textContent || titleEl.value : 'Error loading title',
                        description: descEl ? descEl.textContent || descEl.value : 'Error loading description',
                        keywords: keywordsEl ? (keywordsEl.textContent || keywordsEl.value || 'error, loading').split(',').map(k => k.trim()) : ['error', 'loading'],
                        category: categoryEl ? categoryEl.textContent || categoryEl.value : 'Error loading category'
                    }
                };
            }
            
            return {
                id: photoId,
                aiResult: {
                    title: 'Error loading title',
                    description: 'Error loading description', 
                    keywords: ['error', 'loading'],
                    category: 'Error loading category'
                }
            };
        }
    }

    // Показ диалога для регенерации с текущими данными
    showRegenerateDialog(photoData, onConfirm, onCancel) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50';
        modal.id = 'regenerateModal';
        
        modal.innerHTML = `
            <div class="relative top-10 mx-auto p-6 border w-4/5 max-w-4xl shadow-lg rounded-md bg-white">
                <div class="mt-3">
                    <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 mb-4">
                        <i class="fas fa-sync-alt text-blue-600"></i>
                    </div>
                    <h3 class="text-xl font-medium text-gray-900 mb-6 text-center">${window.i18n.t('review.regenerateDialog.title')}</h3>
                    
                    <div class="mb-6">
                        <h4 class="text-lg font-medium text-gray-800 mb-3">${window.i18n.t('review.regenerateDialog.currentData')}</h4>
                        
                        <div class="bg-gray-50 rounded-lg p-4 mb-4">
                            <div class="mb-3">
                                <label class="block text-sm font-medium text-gray-700 mb-1">${window.i18n.t('review.regenerateDialog.title_field')}</label>
                                <div class="p-2 bg-white rounded border text-sm">${this.escapeHtml(photoData.aiResult?.title || window.i18n.t('review.regenerateDialog.noTitle'))}</div>
                            </div>
                            
                            <div class="mb-3">
                                <label class="block text-sm font-medium text-gray-700 mb-1">${window.i18n.t('review.regenerateDialog.description_field')}</label>
                                <div class="p-2 bg-white rounded border text-sm max-h-20 overflow-y-auto">${this.escapeHtml(photoData.aiResult?.description || window.i18n.t('review.regenerateDialog.noDescription'))}</div>
                            </div>
                            
                            <div class="mb-3">
                                <label class="block text-sm font-medium text-gray-700 mb-1">${window.i18n.t('review.regenerateDialog.keywords_field')}</label>
                                <div class="p-2 bg-white rounded border text-sm">${this.escapeHtml(this.formatKeywords(photoData.aiResult?.keywords) || window.i18n.t('review.regenerateDialog.noKeywords'))}</div>
                            </div>
                            
                            <div>
                                <label class="block text-sm font-medium text-gray-700 mb-1">${window.i18n.t('review.regenerateDialog.category_field')}</label>
                                <div class="p-2 bg-white rounded border text-sm">${this.escapeHtml(photoData.aiResult?.category || window.i18n.t('review.regenerateDialog.noCategory'))}</div>
                            </div>
                        </div>
                    </div>
                    
                    <div class="mb-6">
                        <label class="block text-sm font-medium text-gray-700 mb-2">
                            ${window.i18n.t('review.regenerateDialog.feedbackLabel')}
                        </label>
                        <textarea 
                            id="correctionInput" 
                            placeholder="${window.i18n.t('review.regenerateDialog.feedbackPlaceholder')}"
                            class="w-full p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                            rows="4"
                        ></textarea>
                        <p class="text-xs text-gray-500 mt-2">
                            <i class="fas fa-lightbulb mr-1"></i>
                            ${window.i18n.t('review.regenerateDialog.feedbackTip')}
                        </p>
                    </div>
                    
                    <div class="flex justify-center space-x-4">
                        <button id="regenerateConfirmBtn" class="px-6 py-2 bg-blue-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-300">
                            <i class="fas fa-sync-alt mr-2"></i>${window.i18n.t('review.regenerateDialog.regenerateBtn')}
                        </button>
                        <button id="regenerateCancelBtn" class="px-6 py-2 bg-gray-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300">
                            <i class="fas fa-times mr-2"></i>${window.i18n.t('review.regenerateDialog.cancelBtn')}
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        const input = document.getElementById('correctionInput');
        input.focus();
        
        const handleConfirm = () => {
            const correctionComment = input.value.trim();
            document.body.removeChild(modal);
            if (onConfirm) onConfirm(correctionComment);
        };
        
        const handleCancel = () => {
            document.body.removeChild(modal);
            if (onCancel) onCancel();
        };
        
        document.getElementById('regenerateConfirmBtn').onclick = handleConfirm;
        document.getElementById('regenerateCancelBtn').onclick = handleCancel;
        
        // Ctrl+Enter для подтверждения, Escape для отмены
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && e.ctrlKey) {
                e.preventDefault();
                handleConfirm();
            } else if (e.key === 'Escape') {
                e.preventDefault();
                handleCancel();
            }
        });
        
        modal.onclick = (e) => {
            if (e.target === modal) {
                handleCancel();
            }
        };
    }

    // Функция для показа диалога ввода текста
    showCustomPrompt(message, placeholder = '', onConfirm, onCancel) {
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50';
        modal.id = 'customPromptModal';
        
        const messageHtml = message.replace(/\n/g, '<br>');
        
        modal.innerHTML = `
            <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
                <div class="mt-3">
                    <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100">
                        <i class="fas fa-edit text-blue-600"></i>
                    </div>
                    <h3 class="text-lg font-medium text-gray-900 mb-4 text-center mt-4">Custom AI Prompt</h3>
                    <div class="mt-2 px-7 py-3">
                        <p class="text-sm text-gray-500 mb-4">${messageHtml}</p>
                        <textarea 
                            id="promptInput" 
                            placeholder="${placeholder}"
                            class="w-full p-3 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
                            rows="4"
                        ></textarea>
                        <p class="text-xs text-gray-400 mt-2">Tip: Leave empty to use default AI prompt. Press Ctrl+Enter to confirm.</p>
                    </div>
                    <div class="items-center px-4 py-3 flex justify-center space-x-4">
                        <button id="promptConfirmBtn" class="px-4 py-2 bg-blue-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-300">
                            <i class="fas fa-magic mr-2"></i>Generate
                        </button>
                        <button id="promptCancelBtn" class="px-4 py-2 bg-gray-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300">
                            Cancel
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        const input = document.getElementById('promptInput');
        input.focus();
        
        const handleConfirm = () => {
            const value = input.value.trim();
            document.body.removeChild(modal);
            if (onConfirm) onConfirm(value);
        };
        
        const handleCancel = () => {
            document.body.removeChild(modal);
            if (onCancel) onCancel();
        };
        
        document.getElementById('promptConfirmBtn').onclick = handleConfirm;
        document.getElementById('promptCancelBtn').onclick = handleCancel;
        
        // Enter для подтверждения, Escape для отмены
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && e.ctrlKey) {
                e.preventDefault();
                handleConfirm();
            } else if (e.key === 'Escape') {
                e.preventDefault();
                handleCancel();
            }
        });
        
        modal.onclick = (e) => {
            if (e.target === modal) {
                handleCancel();
            }
        };
    }

    // Массовые действия для всех фотографий в батче
    async approveAllPhotos() {
        const batchId = document.getElementById('batchSelector').value;
        if (!batchId) {
            this.showNotification('No batch selected', 'error');
            return;
        }

        const approveAllBtn = document.getElementById('approveAllBtn');
        const originalText = approveAllBtn.innerHTML;
        
        try {
            approveAllBtn.disabled = true;
            approveAllBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i>Approving...';
            
            if (this.isWailsMode) {
                const photos = await window.go.main.App.GetBatchPhotos(batchId);
                const promises = photos.map(photo => 
                    window.go.main.App.SetPhotoStatus(photo.id, 'approved')
                );
                await Promise.all(promises);
            } else {
                // Mock режим
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
            
            this.showNotification('All photos approved', 'success');
            this.loadBatchForReview(batchId);
            
        } catch (error) {
            console.error('Error approving all photos:', error);
            this.showNotification('Error approving photos: ' + error.message, 'error');
        } finally {
            approveAllBtn.disabled = false;
            approveAllBtn.innerHTML = originalText;
        }
    }

    // Функция для показа уведомлений
    showNotification(message, type = 'info') {
        const container = document.getElementById('notificationContainer');
        if (!container) {
            // Если контейнера нет, создаем простое уведомление
            const notification = document.createElement('div');
            notification.className = `fixed top-4 right-4 p-4 rounded-md shadow-lg z-50 max-w-sm`;
            
            switch (type) {
                case 'success':
                    notification.className += ' bg-green-500 text-white';
                    break;
                case 'error':
                    notification.className += ' bg-red-500 text-white';
                    break;
                case 'warning':
                    notification.className += ' bg-yellow-500 text-white';
                    break;
                default:
                    notification.className += ' bg-blue-500 text-white';
            }
            
            notification.textContent = message;
            document.body.appendChild(notification);
            
            setTimeout(() => {
                if (notification.parentNode) {
                    notification.parentNode.removeChild(notification);
                }
            }, 3000);
            return;
        }

        const id = 'notification-' + Date.now();
        
        const typeClasses = {
            'success': 'bg-green-50 border-green-200 text-green-800',
            'error': 'bg-red-50 border-red-200 text-red-800',
            'warning': 'bg-yellow-50 border-yellow-200 text-yellow-800',
            'info': 'bg-blue-50 border-blue-200 text-blue-800'
        };

        const iconClasses = {
            'success': 'fas fa-check-circle text-green-400',
            'error': 'fas fa-exclamation-circle text-red-400',
            'warning': 'fas fa-exclamation-triangle text-yellow-400',
            'info': 'fas fa-info-circle text-blue-400'
        };

        const notification = document.createElement('div');
        notification.id = id;
        notification.className = `border rounded-md p-4 mb-4 ${typeClasses[type]}`;
        notification.innerHTML = `
            <div class="flex">
                <div class="flex-shrink-0">
                    <i class="${iconClasses[type]}"></i>
                </div>
                <div class="ml-3 flex-1">
                    <p class="text-sm font-medium">${message}</p>
                </div>
                <div class="ml-4 flex-shrink-0">
                    <button onclick="document.getElementById('${id}').remove()" 
                            class="inline-flex text-gray-400 hover:text-gray-600">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            </div>
        `;

        container.appendChild(notification);

        // Автоматически удаляем уведомление через 5 секунд
        setTimeout(() => {
            if (document.getElementById(id)) {
                document.getElementById(id).remove();
            }
        }, 5000);
    }

    // Настройка AI Model Selector
    setupAIModelSelector() {
        console.log('Setting up AI model selector...');
        
        const input = document.getElementById('aiModelInput');
        const toggle = document.getElementById('aiModelToggle');
        const dropdown = document.getElementById('aiModelDropdown');
        
        if (!input || !toggle || !dropdown) {
            console.log('AI model selector elements not found, skipping setup');
            return;
        }

        this.currentModels = [];
        this.selectedModel = null;
        this.isDropdownOpen = false;

        // Обработчик клика по кнопке toggle
        toggle.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            this.toggleAIModelDropdown();
        });

        // Обработчик ввода в поле поиска
        input.addEventListener('input', (e) => {
            const searchTerm = e.target.value;
            this.filterAIModels(searchTerm);
            if (!this.isDropdownOpen) {
                this.showAIModelDropdown();
            }
        });

        // Обработчик фокуса на поле ввода
        input.addEventListener('focus', () => {
            this.showAIModelDropdown();
        });

        // Закрытие dropdown при клике вне элемента
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.ai-model-selector')) {
                this.hideAIModelDropdown();
            }
        });
    }

    showAIModelDropdown() {
        const dropdown = document.getElementById('aiModelDropdown');
        const toggle = document.getElementById('aiModelToggle');
        
        if (dropdown && toggle) {
            dropdown.classList.remove('hidden');
            const icon = toggle.querySelector('i');
            if (icon) {
                icon.classList.remove('fa-chevron-down');
                icon.classList.add('fa-chevron-up');
            }
            this.isDropdownOpen = true;
        }
    }

    hideAIModelDropdown() {
        const dropdown = document.getElementById('aiModelDropdown');
        const toggle = document.getElementById('aiModelToggle');
        
        if (dropdown && toggle) {
            dropdown.classList.add('hidden');
            const icon = toggle.querySelector('i');
            if (icon) {
                icon.classList.remove('fa-chevron-up');
                icon.classList.add('fa-chevron-down');
            }
            this.isDropdownOpen = false;
        }
    }

    toggleAIModelDropdown() {
        if (this.isDropdownOpen) {
            this.hideAIModelDropdown();
        } else {
            this.showAIModelDropdown();
        }
    }

    filterAIModels(searchTerm) {
        // Базовая реализация фильтрации
        console.log('Filtering AI models with term:', searchTerm);
    }

    // Показ списка файлов
    showFilesList(photoType) {
        console.log(`Showing files list for ${photoType}`);
        
        if (!this.selectedFolder || this.selectedFolder.type !== photoType) {
            this.showNotification('No folder selected for this photo type', 'warning');
            return;
        }
        
        this.showNotification(`Files list for ${photoType} - feature available in full version`, 'info');
    }

    // Обновление индикатора режима
    updateModeIndicator() {
        const indicator = document.getElementById('modeIndicator');
        if (!indicator) {
            console.log('Mode indicator not found');
            return;
        }
        
        if (this.isWailsMode) {
            indicator.textContent = '✓ Desktop App';
            indicator.className = 'px-2 py-1 text-xs rounded-full bg-green-100 text-green-800';
        } else {
            indicator.textContent = '⚠ Demo Mode';
            indicator.className = 'px-2 py-1 text-xs rounded-full bg-yellow-100 text-yellow-800';
        }
        indicator.classList.remove('hidden');
    }

    // Запуск периодического обновления очереди
    startQueueUpdates() {
        console.log('Starting queue updates...');
        // Обновляем очередь каждые 5 секунд если мы на вкладке queue
        this.queueUpdateInterval = setInterval(() => {
            if (this.currentTab === 'queue') {
                this.updateQueue();
            }
        }, 5000);
    }

    // Форматирование ключевых слов для отображения
    formatKeywords(keywords) {
        if (!keywords) return '';
        if (Array.isArray(keywords)) {
            return keywords.join(', ');
        }
        return keywords;
    }

    // Обработка Wails file drop
    async handleWailsFileDrop(x, y, paths) {
        console.log(`Files dropped at ${x},${y}:`, paths);
        this.showNotification('File drop feature available in desktop app', 'info');
    }

    // Обновляет действия батча, если функция существует (для совместимости)
    updateBatchActionsIfExists(batchId) {
        console.log('updateBatchActionsIfExists called for batch:', batchId);
        // Функция-заглушка для совместимости
        // В будущем здесь можно добавить логику обновления действий батча
        // например, показ/скрытие кнопок "Upload All", "Select All" и т.д.
        
        // Пример базовой реализации:
        const batchActionsContainer = document.getElementById('batchActions');
        if (batchActionsContainer) {
            // Показываем контейнер с действиями, если он скрыт
            batchActionsContainer.style.display = 'block';
            console.log('Batch actions container made visible');
        }
    }

}

window.addEventListener("DOMContentLoaded", async () => {
    // Определяем, запущено ли приложение в Wails среде
    const isWailsApp = typeof window.go !== 'undefined' && window.go.main && window.go.main.App;
    console.log('Wails environment detected:', isWailsApp);
    
    window.app = new StockPhotoApp(isWailsApp);
});