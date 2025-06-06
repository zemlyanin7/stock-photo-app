// Импортируем i18n
import './i18n.js';
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
                
                console.log('Drag & drop initialized successfully');
            } catch (error) {
                console.error('Error initializing drag & drop:', error);
            }
        }
        
        // Обновляем индикатор режима
        this.updateModeIndicator();
        
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
        console.log('Setting up event listeners...');
        
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
            
            // Если настройки не загрузились, устанавливаем английский и обновляем селекторы
            if (this.isWailsMode) {
            this.showNotification(window.i18n.t('notifications.errorLoading'), 'error');
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

2. ОПИСАНИЕ (до 200 символов):
   - Общее описание без конкретики
   - Фокус на эмоциях и концепциях
   - Универсальность для разных контекстов

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
        window.i18n.updateContent();
        
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
            window.i18n.updateContent();
            return;
        }

        try {
            const photos = await window.go.main.App.GetBatchPhotos(batchId);
            this.renderPhotosForReview(photos);
            this.updateBatchActions(batchId);
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
            window.i18n.updateContent();
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
                                           onchange="app.updatePhotoTitle('${photo.id}', this.value)">
                                </div>

                                <!-- Описание -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
                                    <textarea rows="3" 
                                              class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                              onchange="app.updatePhotoDescription('${photo.id}', this.value)">${this.escapeHtml(aiResult.description)}</textarea>
                                </div>

                                <!-- Ключевые слова -->
                                <div>
                                    <label class="block text-sm font-medium text-gray-700 mb-1">Keywords</label>
                                    <input type="text" 
                                           value="${aiResult.keywords ? aiResult.keywords.join(', ') : ''}" 
                                           class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                           placeholder="keyword1, keyword2, keyword3"
                                           onchange="app.updatePhotoKeywords('${photo.id}', this.value)">
                                </div>

                                <!-- Категория и качество -->
                                <div class="grid grid-cols-2 gap-3">
                                    <div>
                                        <label class="block text-sm font-medium text-gray-700 mb-1">Category</label>
                                        <select class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                                                onchange="app.updatePhotoCategory('${photo.id}', this.value)">
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
                                               onchange="app.updatePhotoQuality('${photo.id}', this.value)">
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
                                <button onclick="app.approvePhoto('${photo.id}')" 
                                        class="px-3 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700 ${photo.status === 'approved' ? 'opacity-50' : ''}"
                                        ${photo.status === 'approved' ? 'disabled' : ''}>
                                    <i class="fas fa-check mr-1"></i>Approve
                                </button>
                                <button onclick="app.rejectPhoto('${photo.id}')" 
                                        class="px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700 ${photo.status === 'rejected' ? 'opacity-50' : ''}"
                                        ${photo.status === 'rejected' ? 'disabled' : ''}>
                                    <i class="fas fa-times mr-1"></i>Reject
                                </button>
                            </div>
                            <div class="flex space-x-2">
                                <button id="regenerateBtn-${photo.id}" onclick="app.regeneratePhotoMetadata('${photo.id}')" 
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
        try {
            // Показываем прогресс
            this.showNotification(window.i18n.t('review.approvingPhoto') || 'Approving photo and writing EXIF metadata...', 'info');
            
            await window.go.main.App.ApprovePhoto(photoId);
            this.showNotification(window.i18n.t('review.photoApprovedWithExif') || 'Photo approved and EXIF metadata written to original file', 'success');
            
            // Обновляем текущий батч
            const batchId = document.getElementById('batchSelector').value;
            if (batchId) {
                this.loadBatchForReview(batchId);
            }
        } catch (error) {
            console.error('Error approving photo:', error);
            this.showNotification('Error approving photo: ' + error.message, 'error');
        }
    }

    async rejectPhoto(photoId) {
        try {
            await window.go.main.App.RejectPhoto(photoId);
            this.showNotification('Photo rejected', 'success');
            // Обновляем текущий батч
            const batchId = document.getElementById('batchSelector').value;
            if (batchId) {
                this.loadBatchForReview(batchId);
            }
        } catch (error) {
            console.error('Error rejecting photo:', error);
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
        
        if (!regenerateBtn || regenerateBtn.disabled) {
            return; // Уже выполняется
        }

        try {
            // Показываем спиннер и блокируем кнопку
            regenerateBtn.disabled = true;
            regenerateIcon.className = 'fas fa-spinner fa-spin mr-1';
            
            const customPrompt = prompt('Enter custom prompt (leave empty for default):');
            if (customPrompt === null) {
                // Пользователь отменил
                return;
            }
            
            if (this.isWailsMode) {
                await window.go.main.App.RegeneratePhotoMetadata(photoId, customPrompt || '');
            } else {
                // Mock режим
                await new Promise(resolve => setTimeout(resolve, 2000));
            }
            
            this.showNotification('Metadata regenerated', 'success');
            
            // Обновляем текущий батч
            const batchId = document.getElementById('batchSelector').value;
            if (batchId) {
                this.loadBatchForReview(batchId);
            }
        } catch (error) {
            console.error('Error regenerating metadata:', error);
            this.showNotification('Error regenerating metadata: ' + error.message, 'error');
        } finally {
            // Восстанавливаем кнопку
            if (regenerateBtn) {
                regenerateBtn.disabled = false;
                regenerateIcon.className = 'fas fa-sync-alt mr-1';
            }
        }
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

    async rejectAllPhotos() {
        const batchId = document.getElementById('batchSelector').value;
        if (!batchId) {
            this.showNotification('No batch selected', 'error');
            return;
        }

        const rejectAllBtn = document.getElementById('rejectAllBtn');
        const originalText = rejectAllBtn.innerHTML;
        
        try {
            rejectAllBtn.disabled = true;
            rejectAllBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i>Rejecting...';
            
            if (this.isWailsMode) {
                const photos = await window.go.main.App.GetBatchPhotos(batchId);
                const promises = photos.map(photo => 
                    window.go.main.App.SetPhotoStatus(photo.id, 'rejected')
                );
                await Promise.all(promises);
            } else {
                // Mock режим
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
            
            this.showNotification('All photos rejected', 'success');
            this.loadBatchForReview(batchId);
            
        } catch (error) {
            console.error('Error rejecting all photos:', error);
            this.showNotification('Error rejecting photos: ' + error.message, 'error');
        } finally {
            rejectAllBtn.disabled = false;
            rejectAllBtn.innerHTML = originalText;
        }
    }

    async regenerateAllPhotos() {
        const batchId = document.getElementById('batchSelector').value;
        if (!batchId) {
            this.showNotification('No batch selected', 'error');
            return;
        }

        const customPrompt = prompt('Enter custom prompt for all photos (leave empty for default):');
        if (customPrompt === null) {
            return; // Пользователь отменил
        }

        const regenerateAllBtn = document.getElementById('regenerateAllBtn');
        const originalText = regenerateAllBtn.innerHTML;
        
        try {
            regenerateAllBtn.disabled = true;
            regenerateAllBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i>Regenerating...';
            
            if (this.isWailsMode) {
                const photos = await window.go.main.App.GetBatchPhotos(batchId);
                
                // Показываем прогресс
                let completed = 0;
                const total = photos.length;
                
                for (const photo of photos) {
                    try {
                        await window.go.main.App.RegeneratePhotoMetadata(photo.id, customPrompt || '');
                        completed++;
                        
                        // Обновляем прогресс в кнопке
                        regenerateAllBtn.innerHTML = `<i class="fas fa-spinner fa-spin mr-1"></i>Regenerating... (${completed}/${total})`;
                        
                    } catch (error) {
                        console.error(`Error regenerating photo ${photo.id}:`, error);
                    }
                }
            } else {
                // Mock режим
                await new Promise(resolve => setTimeout(resolve, 2000));
            }
            
            this.showNotification('All photos regenerated', 'success');
            this.loadBatchForReview(batchId);
            
        } catch (error) {
            console.error('Error regenerating all photos:', error);
            this.showNotification('Error regenerating photos: ' + error.message, 'error');
        } finally {
            regenerateAllBtn.disabled = false;
            regenerateAllBtn.innerHTML = originalText;
        }
    }



    // Управление кнопками батча
    updateBatchActions(batchId) {
        const actionsDiv = document.getElementById('batchActions');
        const uploadBtn = document.getElementById('uploadToStocksBtn');
        
        if (batchId) {
            actionsDiv.classList.remove('hidden');
            this.currentBatchId = batchId;
            
            // Проверяем есть ли одобренные фото для включения кнопки загрузки
            this.checkApprovedPhotos(batchId);
        } else {
            actionsDiv.classList.add('hidden');
            this.currentBatchId = null;
        }
    }

    async checkApprovedPhotos(batchId) {
        try {
            if (this.isWailsMode) {
                const photos = await window.go.main.App.GetBatchPhotos(batchId);
                const approvedCount = photos.filter(p => p.status === 'approved').length;
                
                const uploadBtn = document.getElementById('uploadToStocksBtn');
                uploadBtn.disabled = approvedCount === 0;
                
                if (approvedCount === 0) {
                    uploadBtn.title = window.i18n.t('review.noApprovedPhotos') || 'No approved photos to upload';
                } else {
                    uploadBtn.title = window.i18n.t('review.uploadApprovedPhotos', { count: approvedCount }) || `Upload ${approvedCount} approved photos to stocks`;
                }
            }
        } catch (error) {
            console.error('Error checking approved photos:', error);
        }
    }

    async uploadToStocks() {
        if (!this.currentBatchId) {
            this.showNotification(window.i18n.t('review.noBatchSelected') || 'No batch selected', 'error');
            return;
        }

        const uploadBtn = document.getElementById('uploadToStocksBtn');
        const originalText = uploadBtn.innerHTML;
        
        try {
            // Показываем прогресс
            uploadBtn.disabled = true;
            uploadBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i> Uploading...';
            
            if (this.isWailsMode) {
                await window.go.main.App.UploadApprovedPhotos(this.currentBatchId);
                
                // Начинаем отслеживание прогресса
                this.startUploadProgressTracking(this.currentBatchId);
            } else {
                // Mock режим
                console.log('Mock: Uploading approved photos from batch', this.currentBatchId);
                await new Promise(resolve => setTimeout(resolve, 2000));
            }
            
            this.showNotification(window.i18n.t('review.uploadStarted') || 'Upload to stocks started successfully', 'success');
            
        } catch (error) {
            console.error('Error uploading to stocks:', error);
            
            // Показываем более информативное сообщение об ошибке
            let errorMessage = window.i18n.t('review.uploadError') || 'Failed to start upload to stocks';
            if (error.message) {
                if (error.message.includes('no active stock configurations')) {
                    errorMessage = window.i18n.t('review.noActiveStocks') || 'No active stock configurations found for this photo type. Please configure stocks in Settings > Stocks.';
                } else if (error.message.includes('no approved photos')) {
                    errorMessage = window.i18n.t('review.noApprovedPhotos') || 'No approved photos found for upload';
                } else {
                    errorMessage = `${window.i18n.t('review.uploadError') || 'Upload error'}: ${error.message}`;
                }
            }
            this.showNotification(errorMessage, 'error');
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.innerHTML = originalText;
        }
    }

    deleteBatch() {
        if (!this.currentBatchId) {
            this.showNotification(window.i18n.t('review.noBatchSelected') || 'No batch selected', 'error');
            return;
        }

        // Подтверждение удаления через кастомный диалог
        const confirmMsg = window.i18n.t('review.confirmDelete') || 'Are you sure you want to delete this batch and all its photos? This action cannot be undone.';
        
        this.showCustomConfirm(confirmMsg, () => {
            console.log('User confirmed batch deletion, proceeding...');
            this.performBatchDeletion();
        }, () => {
            console.log('User cancelled batch deletion');
        });
    }

    async performBatchDeletion() {
        const deleteBtn = document.getElementById('deleteBatchBtn');
        const originalText = deleteBtn.innerHTML;
        
        try {
            // Показываем прогресс
            deleteBtn.disabled = true;
            deleteBtn.innerHTML = '<i class="fas fa-spinner fa-spin mr-1"></i> Удаление...';
            
            if (this.isWailsMode) {
                await window.go.main.App.DeleteBatch(this.currentBatchId);
            } else {
                // Mock режим
                console.log('Mock: Deleting batch', this.currentBatchId);
                await new Promise(resolve => setTimeout(resolve, 1000));
            }
            
            this.showNotification(window.i18n.t('review.batchDeleted') || 'Batch deleted successfully', 'success');
            
            // Обновляем интерфейс
            await this.updateReview();
            
        } catch (error) {
            console.error('Error deleting batch:', error);
            this.showNotification(window.i18n.t('review.deleteError') || 'Failed to delete batch', 'error');
        } finally {
            deleteBtn.disabled = false;
            deleteBtn.innerHTML = originalText;
        }
    }

    // Отслеживание прогресса загрузки
    startUploadProgressTracking(batchId) {
        // Создаем индикатор прогресса
        this.showUploadProgress(batchId);
        
        // Проверяем прогресс каждые 2 секунды
        this.uploadProgressInterval = setInterval(async () => {
            try {
                const progress = await window.go.main.App.GetUploadProgress(batchId);
                this.updateUploadProgress(progress);
                
                // Если все загрузки завершены, останавливаем отслеживание
                if (progress.uploadingCount === 0) {
                    this.stopUploadProgressTracking();
                    
                    // Обновляем интерфейс Review
                    setTimeout(() => {
                        this.loadBatchForReview(this.currentBatchId);
                    }, 1000);
                }
            } catch (error) {
                console.error('Error tracking upload progress:', error);
                this.stopUploadProgressTracking();
            }
        }, 2000);
    }

    stopUploadProgressTracking() {
        if (this.uploadProgressInterval) {
            clearInterval(this.uploadProgressInterval);
            this.uploadProgressInterval = null;
        }
        
        // Скрываем индикатор прогресса
        this.hideUploadProgress();
    }

    showUploadProgress(batchId) {
        // Создаем модальное окно с прогрессом
        const progressModal = document.createElement('div');
        progressModal.id = 'uploadProgressModal';
        progressModal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50';
        
        progressModal.innerHTML = `
            <div class="bg-white rounded-lg p-6 max-w-md w-full mx-4">
                <div class="flex items-center justify-between mb-4">
                    <h3 class="text-lg font-medium text-gray-900">Загрузка на стоки</h3>
                    <button id="closeUploadProgress" class="text-gray-400 hover:text-gray-600">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div id="uploadProgressContent">
                    <div class="flex items-center justify-center py-4">
                        <i class="fas fa-spinner fa-spin text-2xl text-blue-500 mr-2"></i>
                        <span>Начинаем загрузку...</span>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(progressModal);
        
        // Закрытие по клику на кнопку
        document.getElementById('closeUploadProgress').addEventListener('click', () => {
            this.stopUploadProgressTracking();
        });
    }

    updateUploadProgress(progress) {
        const content = document.getElementById('uploadProgressContent');
        if (!content) return;
        
        const { photos, totalPhotos, uploadingCount, uploadedCount, failedCount } = progress;
        
        content.innerHTML = `
            <div class="space-y-4">
                <!-- Общий прогресс -->
                <div>
                    <div class="flex justify-between text-sm text-gray-600 mb-1">
                        <span>Прогресс</span>
                        <span>${uploadedCount + failedCount}/${totalPhotos} фото</span>
                    </div>
                    <div class="w-full bg-gray-200 rounded-full h-2">
                        <div class="bg-blue-600 h-2 rounded-full transition-all duration-300" 
                             style="width: ${totalPhotos > 0 ? ((uploadedCount + failedCount) / totalPhotos * 100) : 0}%"></div>
                    </div>
                </div>
                
                <!-- Статистика -->
                <div class="grid grid-cols-3 gap-2 text-center text-sm">
                    <div class="p-2 bg-blue-50 rounded">
                        <div class="font-medium text-blue-600">${uploadingCount}</div>
                        <div class="text-blue-500">Загружается</div>
                    </div>
                    <div class="p-2 bg-green-50 rounded">
                        <div class="font-medium text-green-600">${uploadedCount}</div>
                        <div class="text-green-500">Загружено</div>
                    </div>
                    <div class="p-2 bg-red-50 rounded">
                        <div class="font-medium text-red-600">${failedCount}</div>
                        <div class="text-red-500">Ошибка</div>
                    </div>
                </div>
                
                <!-- Детали по фото -->
                <div class="max-h-32 overflow-y-auto">
                    ${photos.map(photo => {
                        const stockStatuses = Object.entries(photo.stocks).map(([stock, status]) => {
                            const statusClass = status === 'uploaded' ? 'text-green-600' : 
                                              status === 'failed' ? 'text-red-600' : 'text-blue-600';
                            const statusIcon = status === 'uploaded' ? 'fa-check' : 
                                             status === 'failed' ? 'fa-times' : 'fa-spinner fa-spin';
                            return `<span class="${statusClass}"><i class="fas ${statusIcon} mr-1"></i>${stock}</span>`;
                        }).join(', ');
                        
                        return `
                            <div class="flex justify-between items-center py-1 text-sm">
                                <span class="truncate mr-2">${photo.fileName}</span>
                                <div class="text-right">${stockStatuses}</div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    }

    hideUploadProgress() {
        const modal = document.getElementById('uploadProgressModal');
        if (modal) {
            modal.remove();
        }
    }

    toggleBatchDetails(batchId) {
        const details = document.getElementById(`details-${batchId}`);
        const icon = details.previousElementSibling.querySelector('i');
        
        if (details.classList.contains('hidden')) {
            details.classList.remove('hidden');
            icon.classList.remove('fa-chevron-down');
            icon.classList.add('fa-chevron-up');
        } else {
            details.classList.add('hidden');
            icon.classList.remove('fa-chevron-up');
            icon.classList.add('fa-chevron-down');
        }
    }

    getStatusBadgeClass(status) {
        const classes = {
            'pending': 'bg-yellow-100 text-yellow-800',
            'processing': 'bg-blue-100 text-blue-800',
            'completed': 'bg-green-100 text-green-800',
            'failed': 'bg-red-100 text-red-800'
        };
        return classes[status] || 'bg-gray-100 text-gray-800';
    }

    async loadStockConfigs() {
        try {
            console.log('Loading stock configs...');
            const stocks = await window.go.main.App.GetStockConfigs();
            console.log('Loaded stocks:', stocks);
            this.renderStockConfigs(stocks);
        } catch (error) {
            console.error('Error loading stock configs:', error);
            this.showNotification(window.i18n.t('notifications.errorLoading'), 'error');
        }
    }

    renderStockConfigs(stocks) {
        console.log('Rendering stock configs:', stocks);
        const container = document.getElementById('stocksContainer');
        
        if (!stocks || stocks.length === 0) {
            console.log('No stocks to render - showing empty state');
            container.innerHTML = `
                <div class="text-center py-8 text-gray-500">
                    <i class="fas fa-store text-3xl mb-2"></i>
                    <p>${window.i18n.t('settings.stocks.empty')}</p>
                </div>
            `;
            return;
        }

        const stocksHTML = stocks.map(stock => {
            // Используем новое поле type, но если его нет, то uploadMethod для обратной совместимости
            const stockType = stock.type || stock.uploadMethod || 'unknown';
            const hostInfo = stock.connection?.host || stock.connection?.apiUrl || window.i18n.t('notifications.notSpecified') || 'not specified';
            
            return `
            <div class="border border-gray-200 rounded-lg p-4 mb-3">
                <div class="flex justify-between items-start">
                    <div>
                        <h5 class="font-medium text-gray-900">${stock.name}</h5>
                        <p class="text-sm text-gray-500">
                            ID: ${stock.id} • ${window.i18n.t('addStock.fields.type')}: ${stockType.toUpperCase()} • 
                            ${window.i18n.t('addStock.fields.supportedTypes')}: ${stock.supportedTypes?.join(', ') || window.i18n.t('notifications.notSpecified') || 'not specified'}
                        </p>
                        <p class="text-sm text-gray-500">
                            ${stockType === 'api' ? 'API URL' : window.i18n.t('addStock.fields.host')}: ${hostInfo}
                        </p>
                    </div>
                    <div class="flex items-center space-x-2">
                        <label class="flex items-center">
                            <input type="checkbox" ${stock.active ? 'checked' : ''} 
                                   onchange="window.app.toggleStockActive('${stock.id}')"
                                   class="mr-1">
                            <span class="text-sm">${window.i18n.t('settings.stocks.active')}</span>
                        </label>
                        <button onclick="window.app.testStockConnection('${stock.id}')"
                                class="text-blue-600 hover:text-blue-800 text-sm" 
                                title="${window.i18n.t('settings.stocks.test')}">
                            <i class="fas fa-plug"></i>
                        </button>
                        <button onclick="window.app.editStock('${stock.id}')"
                                class="text-gray-600 hover:text-gray-800 text-sm" 
                                title="${window.i18n.t('settings.stocks.edit')}">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button onclick="event.preventDefault(); event.stopPropagation(); window.app.deleteStock('${stock.id}');"
                                class="text-red-600 hover:text-red-800 text-sm" 
                                title="${window.i18n.t('settings.stocks.delete')}">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        `}).join('');

        container.innerHTML = stocksHTML;
    }

    async toggleStockActive(stockId) {
        try {
            if (this.isWailsMode) {
                await window.go.main.App.ToggleStockActive(stockId);
                this.showNotification('Stock activity toggled successfully', 'success');
            } else {
                console.log('Mock: Toggling stock activity for', stockId);
                this.showNotification(`Mock: Stock ${stockId} activity toggled`, 'success');
            }
            this.loadStockConfigs();
        } catch (error) {
            console.error('Error toggling stock activity:', error);
            this.showNotification(window.i18n.t('notifications.errorSaving'), 'error');
        }
    }

    async testStockConnection(stockId) {
        try {
            // Получаем конфигурацию стока и тестируем
            const stocks = await window.go.main.App.GetStockConfigs();
            const stock = stocks.find(s => s.id === stockId);
            
            if (stock) {
                await window.go.main.App.TestStockConnection(stock);
                this.showNotification(`${window.i18n.t('addStock.connectionSuccess')} ${stock.name}`, 'success');
            }
        } catch (error) {
            console.error('Stock connection test failed:', error);
            this.showNotification(window.i18n.t('addStock.connectionFailed') + ': ' + error.message, 'error');
        }
    }

    async editStock(stockId) {
        try {
            // Получаем список всех стоков
            const stocks = await window.go.main.App.GetStockConfigs();
            const stock = stocks.find(s => s.id === stockId);
            
            if (!stock) {
                this.showNotification('Stock configuration not found', 'error');
                return;
            }

            // Загружаем шаблоны стоков
            try {
                this.stockTemplates = await this.getStockTemplates();
            } catch (error) {
                console.error('Error loading stock templates:', error);
                this.stockTemplates = {};
            }

            // Устанавливаем режим редактирования
            this.editingStockId = stockId;
            this.editingStock = stock;

            // Открываем модальное окно
            const modal = document.getElementById('addStockModal');
            const modalTitle = modal.querySelector('h3');
            const submitButton = modal.querySelector('button[type="submit"]');
            
            // Меняем заголовок и кнопку
            modalTitle.textContent = window.i18n.t('addStock.editTitle') || 'Edit Stock Site';
            submitButton.textContent = window.i18n.t('addStock.save') || 'Save Changes';
            
            modal.classList.remove('hidden');

            // Заполняем форму данными стока
            this.populateStockForm(stock);

        } catch (error) {
            console.error('Error opening edit stock modal:', error);
            this.showNotification('Error loading stock for editing', 'error');
        }
    }

    deleteStock(stockId) {
        console.log('Delete stock called with ID:', stockId);
        
        const confirmMessage = window.i18n.t('addStock.deleteConfirm') || 'Вы уверены, что хотите удалить эту конфигурацию стока?';
        
        // Создаем кастомный диалог подтверждения
        this.showCustomConfirm(confirmMessage, () => {
            console.log('User confirmed deletion, proceeding...');
            this.performStockDeletion(stockId);
        }, () => {
            console.log('User cancelled deletion');
        });
    }

    showCustomConfirm(message, onConfirm, onCancel) {
        // Создаем модальный диалог
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50';
        modal.id = 'customConfirmModal';
        
        const confirmText = window.i18n.t('settings.stocks.delete') || 'Удалить';
        const cancelText = window.i18n.t('addStock.cancel') || 'Отмена';
        const titleText = 'Подтверждение удаления';
        
        modal.innerHTML = `
            <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
                <div class="mt-3 text-center">
                    <div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-red-100">
                        <i class="fas fa-exclamation-triangle text-red-600"></i>
                    </div>
                    <h3 class="text-lg font-medium text-gray-900 mt-4">${titleText}</h3>
                    <div class="mt-2 px-7 py-3">
                        <p class="text-sm text-gray-500">${message}</p>
                    </div>
                    <div class="items-center px-4 py-3">
                        <button id="confirmBtn" class="px-4 py-2 bg-red-600 text-white text-base font-medium rounded-md w-24 mr-2 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-300">
                            ${confirmText}
                        </button>
                        <button id="cancelBtn" class="px-4 py-2 bg-gray-500 text-white text-base font-medium rounded-md w-24 hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300">
                            ${cancelText}
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
        
        // Обработчики событий
        document.getElementById('confirmBtn').onclick = () => {
            document.body.removeChild(modal);
            onConfirm();
        };
        
        document.getElementById('cancelBtn').onclick = () => {
            document.body.removeChild(modal);
            if (onCancel) onCancel();
        };
        
        // Закрытие по клику вне диалога
        modal.onclick = (e) => {
            if (e.target === modal) {
                document.body.removeChild(modal);
                if (onCancel) onCancel();
            }
        };
        
        // Закрытие по Escape
        const handleKeyDown = (e) => {
            if (e.key === 'Escape') {
                document.body.removeChild(modal);
                document.removeEventListener('keydown', handleKeyDown);
                if (onCancel) onCancel();
            }
        };
        document.addEventListener('keydown', handleKeyDown);
    }

    async performStockDeletion(stockId) {
        try {
            if (this.isWailsMode) {
                await window.go.main.App.DeleteStockConfig(stockId);
            } else {
                console.log('Mock: Deleting stock', stockId);
            }
            
            const successText = window.i18n.t('addStock.deleted') || 'Конфигурация стока удалена';
            this.showNotification(successText, 'success');
            await this.loadStockConfigs();
            
        } catch (error) {
            console.error('Error deleting stock:', error);
            const errorText = window.i18n.t('addStock.error') || 'Ошибка удаления конфигурации стока';
            this.showNotification(errorText + ': ' + error.message, 'error');
        }
    }

    async openAddStockModal() {
        // Загружаем шаблоны стоков
        try {
            this.stockTemplates = await this.getStockTemplates();
        } catch (error) {
            console.error('Error loading stock templates:', error);
            this.stockTemplates = {};
        }
        
        document.getElementById('addStockModal').classList.remove('hidden');
        // Очищаем форму и динамические поля
        document.getElementById('addStockForm').reset();
        document.getElementById('dynamicFields').innerHTML = '';
        document.getElementById('stockTypeDescription').textContent = '';
        
        // Скрываем контейнер тестирования и очищаем результат
        document.getElementById('testConnectionContainer').classList.add('hidden');
        document.getElementById('testConnectionResult').classList.add('hidden');
        document.getElementById('testConnectionResult').innerHTML = '';
    }

    closeAddStockModal() {
        document.getElementById('addStockModal').classList.add('hidden');
        document.getElementById('addStockForm').reset();
        document.getElementById('dynamicFields').innerHTML = '';
        
        // Скрываем и очищаем результат тестирования
        document.getElementById('testConnectionContainer').classList.add('hidden');
        document.getElementById('testConnectionResult').classList.add('hidden');
        document.getElementById('testConnectionResult').innerHTML = '';

        // Сбрасываем режим редактирования
        this.editingStockId = null;
        this.editingStock = null;
        
        // Возвращаем оригинальные тексты
        const modal = document.getElementById('addStockModal');
        const modalTitle = modal.querySelector('h3');
        const submitButton = modal.querySelector('button[type="submit"]');
        modalTitle.textContent = window.i18n.t('addStock.title') || 'Add Stock Site';
        submitButton.textContent = window.i18n.t('addStock.add') || 'Add Stock Site';
    }

    populateStockForm(stock) {
        // Заполняем основные поля
        document.querySelector('input[name="id"]').value = stock.id;
        document.querySelector('input[name="name"]').value = stock.name;
        document.getElementById('stockType').value = stock.type;

        // Заполняем поддерживаемые типы
        const supportedTypesCheckboxes = document.querySelectorAll('input[name="supportedTypes"]');
        supportedTypesCheckboxes.forEach(checkbox => {
            checkbox.checked = stock.supportedTypes && stock.supportedTypes.includes(checkbox.value);
        });

        // Генерируем динамические поля для выбранного типа
        this.onStockTypeChange(stock.type);

        // Ждем генерации полей, затем заполняем их
        setTimeout(() => {
            if (stock.connection) {
                Object.keys(stock.connection).forEach(key => {
                    const field = document.getElementById(key);
                    if (field) {
                        if (field.type === 'checkbox') {
                            field.checked = stock.connection[key];
                        } else {
                            field.value = stock.connection[key];
                        }
                    }
                });
            }
        }, 100);
    }

    async getStockTemplates() {
        if (this.isWailsMode) {
            try {
            return await window.go.main.App.GetStockTemplates();
            } catch (error) {
                console.error('Error getting stock templates:', error);
            }
        }

        // Mock данные для развития без Wails
            return {
                "ftp": {
                    "type": "ftp",
                    "name": "FTP Upload",
                    "description": window.i18n.t('addStock.types.ftp.description'),
                    "fields": [
                        {"name": "host", "type": "text", "label": window.i18n.t('addStock.fieldLabels.ftpServer'), "required": true, "placeholder": "ftp.example.com"},
                    {"name": "port", "type": "number", "label": window.i18n.t('addStock.fields.port'), "required": true, "default": 21, "placeholder": "21 для FTP, 990 для implicit FTPS"},
                        {"name": "username", "type": "text", "label": window.i18n.t('addStock.fieldLabels.username'), "required": true},
                        {"name": "password", "type": "password", "label": window.i18n.t('addStock.fieldLabels.password'), "required": true},
                        {"name": "path", "type": "text", "label": window.i18n.t('addStock.fieldLabels.remotePath'), "default": "/", "placeholder": "/uploads/"},
                    {"name": "encryption", "type": "select", "label": window.i18n.t('addStock.fieldLabels.encryption'), "default": "none", "options": ["none", "auto", "explicit", "implicit"]},
                    {"name": "verifyCert", "type": "checkbox", "label": window.i18n.t('addStock.fieldLabels.verifyCert'), "default": true},
                        {"name": "passive", "type": "checkbox", "label": window.i18n.t('addStock.fieldLabels.passiveMode'), "default": true},
                        {"name": "timeout", "type": "number", "label": window.i18n.t('addStock.fieldLabels.timeout'), "default": 30}
                    ]
                },
                "sftp": {
                    "type": "sftp",
                    "name": "SFTP Upload", 
                    "description": window.i18n.t('addStock.types.sftp.description'),
                    "fields": [
                        {"name": "host", "type": "text", "label": window.i18n.t('addStock.fieldLabels.sftpServer'), "required": true, "placeholder": "sftp.example.com"},
                        {"name": "port", "type": "number", "label": window.i18n.t('addStock.fields.port'), "required": true, "default": 22},
                        {"name": "username", "type": "text", "label": window.i18n.t('addStock.fieldLabels.username'), "required": true},
                        {"name": "password", "type": "password", "label": window.i18n.t('addStock.fieldLabels.password'), "required": true},
                        {"name": "path", "type": "text", "label": window.i18n.t('addStock.fieldLabels.remotePath'), "default": "/", "placeholder": "/uploads/"},
                        {"name": "timeout", "type": "number", "label": window.i18n.t('addStock.fieldLabels.timeout'), "default": 30}
                    ]
            }
        };
    }

    onStockTypeChange(stockType) {
        const description = document.getElementById('stockTypeDescription');
        const dynamicFields = document.getElementById('dynamicFields');
        const testContainer = document.getElementById('testConnectionContainer');
        
        // Очищаем предыдущие поля
        dynamicFields.innerHTML = '';
        
        if (!stockType || !this.stockTemplates || !this.stockTemplates[stockType]) {
            description.textContent = '';
            testContainer.classList.add('hidden');
            return;
        }

        const template = this.stockTemplates[stockType];
        
        // Обновляем описание
        const typeKey = `addStock.types.${stockType}.description`;
        description.textContent = window.i18n.t(typeKey) || template.description;
        
        // Создаем динамические поля
        if (template.fields) {
            template.fields.forEach(field => {
                const fieldDiv = this.createDynamicField(field);
                dynamicFields.appendChild(fieldDiv);
                
                // Добавляем слушатель изменений для валидации
                const input = fieldDiv.querySelector('input, select, textarea');
                if (input) {
                    input.addEventListener('input', () => {
                        this.validateStockConnectionFields();
                    });
                }
            });
        }
        
        // Показываем кнопку тестирования для FTP и SFTP
        if (stockType === 'ftp' || stockType === 'sftp') {
            testContainer.classList.remove('hidden');
            this.validateStockConnectionFields(); // Первоначальная валидация
        } else {
            testContainer.classList.add('hidden');
        }
    }

    createDynamicField(field) {
        const div = document.createElement('div');
        
        const label = document.createElement('label');
        label.className = 'block text-sm font-medium text-gray-700';
        label.textContent = field.label;
        label.setAttribute('for', field.name);
        
        let input;
        
        switch (field.type) {
            case 'text':
            case 'password':
            case 'url':
            case 'number':
                input = document.createElement('input');
                input.type = field.type;
                input.name = field.name;
                input.id = field.name;
                input.className = 'mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500';
                if (field.placeholder) input.placeholder = field.placeholder;
                if (field.default !== undefined) input.value = field.default;
                if (field.required) input.required = true;
                break;
                
            case 'select':
                input = document.createElement('select');
                input.name = field.name;
                input.id = field.name;
                input.className = 'mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500';
                if (field.required) input.required = true;
                
                // Добавляем опции
                if (field.options) {
                    field.options.forEach(optionValue => {
                        const option = document.createElement('option');
                        option.value = optionValue;
                        
                        // Используем переводы для опций шифрования
                        if (field.name === 'encryption') {
                            option.textContent = window.i18n.t(`addStock.encryption.${optionValue}`) || optionValue;
                        } else {
                            option.textContent = optionValue;
                        }
                        
                        // Устанавливаем значение по умолчанию
                        if (field.default === optionValue) {
                            option.selected = true;
                        }
                        
                        input.appendChild(option);
                    });
                }
                break;
                
            case 'checkbox':
                input = document.createElement('div');
                input.className = 'mt-2';
                const checkboxLabel = document.createElement('label');
                checkboxLabel.className = 'inline-flex items-center';
                
                const checkbox = document.createElement('input');
                checkbox.type = 'checkbox';
                checkbox.name = field.name;
                checkbox.id = field.name;
                checkbox.className = 'rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50';
                if (field.default === true) checkbox.checked = true;
                
                const span = document.createElement('span');
                span.className = 'ml-2';
                span.textContent = field.label;
                
                checkboxLabel.appendChild(checkbox);
                checkboxLabel.appendChild(span);
                input.appendChild(checkboxLabel);
                
                // Для чекбокса не нужен отдельный label
                div.appendChild(input);
                return div;
                
            default:
                input = document.createElement('input');
                input.type = 'text';
                input.name = field.name;
                input.id = field.name;
                input.className = 'mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500';
        }
        
        div.appendChild(label);
        div.appendChild(input);
        
        return div;
    }

    async addStock(event) {
        event.preventDefault();
        
        const formData = new FormData(event.target);
        const supportedTypes = formData.getAll('supportedTypes');
        const stockType = formData.get('stockType');
        
        // Собираем все поля подключения
        const connection = {};
        const template = this.stockTemplates && this.stockTemplates[stockType];
        
        if (template && template.fields) {
            template.fields.forEach(field => {
                const value = formData.get(field.name);
                if (value !== null) {
                    if (field.type === 'number') {
                        connection[field.name] = parseInt(value) || field.default || 0;
                    } else if (field.type === 'checkbox') {
                        connection[field.name] = formData.has(field.name);
                    } else {
                        connection[field.name] = value || field.default || '';
                    }
                }
            });
        }
        
        const stockConfig = {
            id: formData.get('id'),
            name: formData.get('name'),
            type: stockType,
            supportedTypes: supportedTypes,
            connection: connection,
            prompts: this.editingStock ? this.editingStock.prompts : {},
            settings: this.editingStock ? this.editingStock.settings : {},
            active: this.editingStock ? this.editingStock.active : true
        };

        try {
            if (this.isWailsMode) {
                await window.go.main.App.SaveStockConfig(stockConfig);
            } else {
                console.log('Mock: SaveStockConfig', stockConfig);
            }
            
            const successMessage = this.editingStockId ? 
                (window.i18n.t('addStock.updated') || 'Stock configuration updated successfully') :
                window.i18n.t('addStock.success');
            
            this.showNotification(successMessage, 'success');
            this.closeAddStockModal();
            this.loadStockConfigs();
        } catch (error) {
            console.error('Error saving stock:', error);
            const errorMessage = this.editingStockId ?
                'Error updating stock configuration' :
                (window.i18n.t('addStock.error') + ': ' + error.message);
            this.showNotification(errorMessage, 'error');
        }
    }

    // Валидация полей подключения для активации кнопки тестирования
    validateStockConnectionFields() {
        const testBtn = document.getElementById('testStockConnectionBtn');
        const stockType = document.getElementById('stockType').value;
        
        if (!stockType || (stockType !== 'ftp' && stockType !== 'sftp')) {
            testBtn.disabled = true;
            return;
        }
        
        const template = this.stockTemplates && this.stockTemplates[stockType];
        if (!template || !template.fields) {
            testBtn.disabled = true;
            return;
        }
        
        // Проверяем обязательные поля
        let allRequiredFieldsFilled = true;
        template.fields.forEach(field => {
            if (field.required) {
                const input = document.querySelector(`[name="${field.name}"]`);
                if (!input || !input.value.trim()) {
                    allRequiredFieldsFilled = false;
                }
            }
        });
        
        testBtn.disabled = !allRequiredFieldsFilled;
    }

    // Тестирование соединения в модальном окне
    async testStockConnectionInModal() {
        const testBtn = document.getElementById('testStockConnectionBtn');
        const testResult = document.getElementById('testConnectionResult');
        const stockType = document.getElementById('stockType').value;
        
        // Показываем состояние загрузки
        const originalText = testBtn.querySelector('span').textContent;
        testBtn.disabled = true;
        testBtn.querySelector('span').textContent = window.i18n.t('addStock.testingConnection');
        testBtn.querySelector('i').className = 'fas fa-spinner fa-spin mr-2';
        
        // Очищаем предыдущий результат
        testResult.classList.add('hidden');
        testResult.innerHTML = '';
        
        try {
            // Собираем данные из формы
            const formData = new FormData(document.getElementById('addStockForm'));
            const connection = {};
            const template = this.stockTemplates && this.stockTemplates[stockType];
            
            if (template && template.fields) {
                template.fields.forEach(field => {
                    const value = formData.get(field.name);
                    if (value !== null) {
                        if (field.type === 'number') {
                            connection[field.name] = parseInt(value) || field.default || 0;
                        } else if (field.type === 'checkbox') {
                            connection[field.name] = formData.has(field.name);
                        } else {
                            connection[field.name] = value || field.default || '';
                        }
                    }
                });
            }
            
            const testConfig = {
                type: stockType,
                connection: connection
            };
            
            // Вызываем метод тестирования
            if (this.isWailsMode) {
                await window.go.main.App.TestStockConnection(testConfig);
            } else {
                // Mock тестирование
                await new Promise(resolve => setTimeout(resolve, 2000)); // Имитация задержки
                console.log('Mock: TestStockConnection', testConfig);
            }
            
            // Успешное тестирование
            testResult.innerHTML = `
                <div class="flex items-center p-3 bg-green-50 border border-green-200 rounded-md">
                    <i class="fas fa-check-circle text-green-400 mr-2"></i>
                    <span class="text-green-800">${window.i18n.t('addStock.connectionSuccess')}</span>
                </div>
            `;
            testResult.classList.remove('hidden');
            
        } catch (error) {
            console.error('Stock connection test failed:', error);
            
            // Показываем ошибку
            testResult.innerHTML = `
                <div class="flex items-center p-3 bg-red-50 border border-red-200 rounded-md">
                    <i class="fas fa-exclamation-circle text-red-400 mr-2"></i>
                    <div class="text-red-800">
                        <div class="font-medium">${window.i18n.t('addStock.connectionFailed')}</div>
                        <div class="text-sm mt-1">${error.message || 'Unknown error'}</div>
                    </div>
                </div>
            `;
            testResult.classList.remove('hidden');
        } finally {
            // Восстанавливаем кнопку
            testBtn.disabled = false;
            testBtn.querySelector('span').textContent = originalText;
            testBtn.querySelector('i').className = 'fas fa-plug mr-2';
            
            // Повторная валидация для состояния кнопки
            this.validateStockConnectionFields();
        }
    }

    startQueueUpdates() {
        // Обновляем очередь каждые 5 секунд если мы на вкладке queue
        this.queueUpdateInterval = setInterval(() => {
            if (this.currentTab === 'queue') {
                this.updateQueue();
            }
        }, 5000);
    }

    showNotification(message, type = 'info') {
        const container = document.getElementById('notificationContainer');
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

    showFilesList(photoType) {
        if (!this.selectedFolder || this.selectedFolder.type !== photoType) {
            return;
        }
        
        const files = this.selectedFolder.files;
        const modal = document.createElement('div');
        modal.className = 'fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50';
        modal.id = 'filesListModal';
        
        const validFiles = files.filter(f => f.isValid);
        const invalidFiles = files.filter(f => !f.isValid);
        
        modal.innerHTML = `
            <div class="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-md bg-white">
                <div class="mt-3">
                    <div class="flex justify-between items-center mb-4">
                        <h3 class="text-lg font-medium text-gray-900">${window.i18n.t('notifications.filesInFolder')}</h3>
                        <button onclick="document.getElementById('filesListModal').remove()" 
                                class="text-gray-400 hover:text-gray-600">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    
                    <div class="mb-4">
                        <p class="text-sm text-gray-600">
                            ${window.i18n.t('notifications.totalFiles')}: ${files.length} | 
                            ${window.i18n.t('notifications.validImages')}: ${validFiles.length} | 
                            ${window.i18n.t('notifications.invalidFiles')}: ${invalidFiles.length}
                        </p>
                    </div>
                    
                    <div class="max-h-96 overflow-y-auto">
                        ${validFiles.length > 0 ? `
                            <div class="mb-4">
                                <h4 class="font-medium text-green-600 mb-2">${window.i18n.t('notifications.validImages')} (${validFiles.length})</h4>
                                <div class="space-y-1">
                                    ${validFiles.map(file => `
                                        <div class="flex justify-between items-center p-2 bg-green-50 rounded text-sm">
                                            <span class="text-gray-900">${file.name}</span>
                                            <span class="text-gray-500">${this.formatFileSize(file.size)}</span>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        ` : ''}
                        
                        ${invalidFiles.length > 0 ? `
                            <div>
                                <h4 class="font-medium text-red-600 mb-2">${window.i18n.t('notifications.invalidFiles')} (${invalidFiles.length})</h4>
                                <div class="space-y-1">
                                    ${invalidFiles.map(file => `
                                        <div class="flex justify-between items-center p-2 bg-red-50 rounded text-sm">
                                            <span class="text-gray-900">${file.name}</span>
                                            <span class="text-gray-500">${this.formatFileSize(file.size)}</span>
                                        </div>
                                    `).join('')}
                                </div>
                            </div>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
        
        document.body.appendChild(modal);
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    updateModeIndicator() {
        const indicator = document.getElementById('modeIndicator');
        if (this.isWailsMode) {
            indicator.textContent = '✓ Desktop App';
            indicator.className = 'px-2 py-1 text-xs rounded-full bg-green-100 text-green-800';
        } else {
            indicator.textContent = '⚠ Demo Mode';
            indicator.className = 'px-2 py-1 text-xs rounded-full bg-yellow-100 text-yellow-800';
        }
        indicator.classList.remove('hidden');
    }

    async handleWailsFileDrop(x, y, paths) {
        console.log(`Files dropped at ${x},${y}:`, paths);
        
        // Определяем тип фото на основе координат drop zone
        const photoType = this.determinePhotoTypeFromCoordinates(x, y);
        
        if (!photoType) {
            this.showNotification(window.i18n.t('notifications.dragToAreas'), 'warning');
            return;
        }
        
        // Если перетащили несколько элементов, берем первый
        let folderPath = paths[0];
        
        // Проверяем - это папка или файл
        try {
            // Пытаемся определить, это папка или файл
            // Если это файл, получаем путь к его папке
            const isFile = folderPath.includes('.') && !folderPath.endsWith('/');
            
            if (isFile) {
                // Это файл - получаем путь к папке
                const lastSlash = folderPath.lastIndexOf('/');
                const lastBackslash = folderPath.lastIndexOf('\\');
                const separator = Math.max(lastSlash, lastBackslash);
                
                if (separator !== -1) {
                    folderPath = folderPath.substring(0, separator);
                } else {
                    this.showNotification(window.i18n.t('notifications.cannotDetermineFromFile'), 'error');
                    return;
                }
            }
            
            // Обрабатываем папку
            await this.selectFolder(folderPath, photoType);
            
        } catch (error) {
            console.error('Error processing dropped files:', error);
            this.showNotification(window.i18n.t('notifications.dragProcessingError', {error: error.message}), 'error');
        }
    }

    determinePhotoTypeFromCoordinates(x, y) {
        // Получаем drop zones
        const editorialDropZone = document.getElementById('editorialDropZone');
        const commercialDropZone = document.getElementById('commercialDropZone');
        
        if (!editorialDropZone || !commercialDropZone) {
            return null;
        }
        
        // Получаем границы областей
        const editorialRect = editorialDropZone.getBoundingClientRect();
        const commercialRect = commercialDropZone.getBoundingClientRect();
        
        // Проверяем, попадают ли координаты в editorial zone
        if (x >= editorialRect.left && x <= editorialRect.right && 
            y >= editorialRect.top && y <= editorialRect.bottom) {
            return 'editorial';
        }
        
        // Проверяем, попадают ли координаты в commercial zone
        if (x >= commercialRect.left && x <= commercialRect.right && 
            y >= commercialRect.top && y <= commercialRect.bottom) {
            return 'commercial';
        }
        
        return null;
    }

    // Настройка кастомного селектора модели AI
    setupAIModelSelector() {
        const input = document.getElementById('aiModelInput');
        const toggle = document.getElementById('aiModelToggle');
        const dropdown = document.getElementById('aiModelDropdown');
        const modelList = document.getElementById('aiModelList');
        const emptyMessage = document.getElementById('aiModelEmpty');

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

        // Обработчик Enter в поле ввода
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                const value = input.value.trim();
                if (value) {
                    this.selectAIModel({ 
                        id: value, 
                        name: value, 
                        description: window.i18n.t('settings.ai.modelCustom'), 
                        isCustom: true 
                    });
                }
            } else if (e.key === 'Escape') {
                this.hideAIModelDropdown();
            } else if (e.key === 'ArrowDown') {
                e.preventDefault();
                this.navigateAIModelList('down');
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                this.navigateAIModelList('up');
            }
        });

        // Закрытие dropdown при клике вне элемента
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.ai-model-selector')) {
                this.hideAIModelDropdown();
            }
        });
    }

    // Показать dropdown моделей
    showAIModelDropdown() {
        const dropdown = document.getElementById('aiModelDropdown');
        const toggle = document.getElementById('aiModelToggle');
        
        dropdown.classList.remove('hidden');
        toggle.querySelector('i').classList.remove('fa-chevron-down');
        toggle.querySelector('i').classList.add('fa-chevron-up');
        this.isDropdownOpen = true;
    }

    // Скрыть dropdown моделей
    hideAIModelDropdown() {
        const dropdown = document.getElementById('aiModelDropdown');
        const toggle = document.getElementById('aiModelToggle');
        
        dropdown.classList.add('hidden');
        toggle.querySelector('i').classList.remove('fa-chevron-up');
        toggle.querySelector('i').classList.add('fa-chevron-down');
        this.isDropdownOpen = false;
    }

    // Переключить состояние dropdown
    toggleAIModelDropdown() {
        if (this.isDropdownOpen) {
            this.hideAIModelDropdown();
        } else {
            this.showAIModelDropdown();
        }
    }

    // Фильтрация моделей по поисковому запросу
    filterAIModels(searchTerm) {
        const modelList = document.getElementById('aiModelList');
        const emptyMessage = document.getElementById('aiModelEmpty');
        
        if (!this.currentModels || this.currentModels.length === 0) {
            modelList.innerHTML = '';
            emptyMessage.classList.remove('hidden');
            return;
        }

        const filteredModels = this.currentModels.filter(model => 
            model.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            model.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
            (model.description && model.description.toLowerCase().includes(searchTerm.toLowerCase()))
        );

        this.renderAIModelList(filteredModels);
        
        if (filteredModels.length === 0) {
            emptyMessage.classList.remove('hidden');
        } else {
            emptyMessage.classList.add('hidden');
        }
    }

    // Отображение списка моделей
    renderAIModelList(models) {
        const modelList = document.getElementById('aiModelList');
        
        modelList.innerHTML = models.map(model => `
            <button type="button" class="ai-model-option" data-model-id="${model.id}">
                <span class="model-name">${model.name}</span>
                <span class="model-description">${model.description || ''}</span>
                <span class="model-tokens">${model.maxTokens ? `${model.maxTokens.toLocaleString()} tokens` : ''}</span>
            </button>
        `).join('');

        // Добавляем обработчики событий для опций
        modelList.querySelectorAll('.ai-model-option').forEach(option => {
            option.addEventListener('click', (e) => {
                e.preventDefault();
                const modelId = option.dataset.modelId;
                const model = models.find(m => m.id === modelId);
                if (model) {
                    this.selectAIModel(model);
                }
            });
        });
    }

    // Выбор модели
    selectAIModel(model) {
        const input = document.getElementById('aiModelInput');
        
        this.selectedModel = model;
        input.value = model.name || model.id;
        
        this.hideAIModelDropdown();
        this.onAIModelChange(model.id);
        
        // Визуально выделяем выбранную модель
        const options = document.querySelectorAll('.ai-model-option');
        options.forEach(opt => opt.classList.remove('selected'));
        
        const selectedOption = document.querySelector(`[data-model-id="${model.id}"]`);
        if (selectedOption) {
            selectedOption.classList.add('selected');
        }
    }

    // Навигация по списку моделей клавишами
    navigateAIModelList(direction) {
        const options = document.querySelectorAll('.ai-model-option');
        if (options.length === 0) return;

        let currentIndex = -1;
        const selectedOption = document.querySelector('.ai-model-option.selected');
        if (selectedOption) {
            currentIndex = Array.from(options).indexOf(selectedOption);
        }

        let newIndex;
        if (direction === 'down') {
            newIndex = currentIndex < options.length - 1 ? currentIndex + 1 : 0;
        } else {
            newIndex = currentIndex > 0 ? currentIndex - 1 : options.length - 1;
        }

        // Убираем выделение с текущей опции
        options.forEach(opt => opt.classList.remove('selected'));
        
        // Выделяем новую опцию
        options[newIndex].classList.add('selected');
        options[newIndex].scrollIntoView({ block: 'nearest' });
    }

    showTab(tabName) {
        // Hide all tabs
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.add('hidden');
        });

        // Show the selected tab
        document.getElementById(tabName).classList.remove('hidden');
    }

    showLogs() {
        this.showTab('logs');
        this.updateLogs();
    }

    async updateLogs() {
        try {
            // Пытаемся получить все обработанные батчи, и если не получается - берем из истории
            let batches = [];
            try {
                console.log('Trying to load processed batches...');
                batches = await window.go.main.App.GetProcessedBatches();
                console.log('GetProcessedBatches result:', batches);
            } catch (error) {
                console.log('GetProcessedBatches failed, trying GetProcessingHistory:', error);
                try {
                    batches = await window.go.main.App.GetProcessingHistory(20);
                    console.log('GetProcessingHistory result:', batches);
                } catch (historyError) {
                    console.error('Both methods failed:', historyError);
                    batches = [];
                }
            }
            
            console.log('Final batches for logs:', batches);
            this.renderLogViewer(batches);
        } catch (error) {
            console.error('Error loading logs:', error);
            this.showNotification(window.i18n.t('logs.loadError') || 'Failed to load logs', 'error');
        }
    }

    renderLogViewer(batches) {
        const container = document.getElementById('logs-content');
        
        if (batches.length === 0) {
            container.innerHTML = `
                <div class="bg-white rounded-lg shadow-md p-6">
                    <div class="text-center text-gray-500 py-12">
                        <div class="mb-4">
                            <i class="fas fa-inbox text-6xl text-gray-300"></i>
                        </div>
                        <h3 class="text-lg font-medium text-gray-700 mb-2">
                            ${window.i18n.t('logs.noBatches') || 'Нет доступных батчей'}
                        </h3>
                        <p class="text-gray-500 mb-4">
                            Обработайте фотографии в разделе "Редакционные Фото" или "Коммерческие Фото",<br>
                            чтобы увидеть здесь журнал событий.
                        </p>
                        <button onclick="app.switchTab('editorial')" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md mr-2">
                            Редакционные Фото
                        </button>
                        <button onclick="app.switchTab('commercial')" class="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded-md">
                            Коммерческие Фото
                        </button>
                    </div>
                </div>
            `;
            return;
        }

        container.innerHTML = `
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="mb-6">
                    <h3 class="text-lg font-semibold text-gray-900 mb-2">
                        ${window.i18n.t('logs.title') || 'Журнал событий'}
                    </h3>
                    <p class="text-gray-600 text-sm mb-4">
                        Выберите батч из списка ниже, чтобы просмотреть детальные логи обработки
                    </p>
                    
                    <!-- Селектор батчей -->
                    <div class="bg-gray-50 rounded-lg p-4 mb-4">
                        <label for="batchLogSelector" class="block text-sm font-medium text-gray-700 mb-2">
                            📋 Выберите батч для просмотра логов:
                        </label>
                        <div class="flex items-center space-x-3">
                            <select id="batchLogSelector" class="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
                                <option value="">${window.i18n.t('logs.selectBatch') || '-- Выберите батч --'}</option>
                                ${batches.map(batch => `
                                    <option value="${batch.id}">${batch.description || 'Batch ' + batch.id} (${new Date(batch.createdAt).toLocaleDateString()})</option>
                                `).join('')}
                            </select>
                            <button id="refreshLogsBtn" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium">
                                <i class="fas fa-sync mr-1"></i>
                                ${window.i18n.t('logs.refresh') || 'Обновить'}
                            </button>
                        </div>
                        <p class="text-xs text-gray-500 mt-2">
                            Доступно батчей: ${batches.length}
                        </p>
                    </div>
                </div>
                
                <div id="logsContainer" class="space-y-2">
                    <div class="text-center text-gray-500 py-12 border-2 border-dashed border-gray-200 rounded-lg">
                        <div class="mb-4">
                            <i class="fas fa-arrow-up text-4xl text-gray-300"></i>
                        </div>
                        <h4 class="text-lg font-medium text-gray-600 mb-2">
                            ${window.i18n.t('logs.selectBatchToView') || 'Выберите батч выше'}
                        </h4>
                        <p class="text-gray-500">
                            Используйте селектор батчей выше,<br>
                            чтобы просмотреть детальные события обработки
                        </p>
                    </div>
                </div>
            </div>
        `;

        // Обработчики событий
        document.getElementById('batchLogSelector').addEventListener('change', (e) => {
            if (e.target.value) {
                this.loadBatchLogs(e.target.value);
            } else {
                document.getElementById('logsContainer').innerHTML = `
                    <div class="text-center text-gray-500 py-12 border-2 border-dashed border-gray-200 rounded-lg">
                        <div class="mb-4">
                            <i class="fas fa-arrow-up text-4xl text-gray-300"></i>
                        </div>
                        <h4 class="text-lg font-medium text-gray-600 mb-2">
                            ${window.i18n.t('logs.selectBatchToView') || 'Выберите батч выше'}
                        </h4>
                        <p class="text-gray-500">
                            Используйте селектор батчей выше,<br>
                            чтобы просмотреть детальные события обработки
                        </p>
                    </div>
                `;
            }
        });

        document.getElementById('refreshLogsBtn').addEventListener('click', () => {
            const selectedBatch = document.getElementById('batchLogSelector').value;
            if (selectedBatch) {
                this.loadBatchLogs(selectedBatch);
            }
        });
    }

    async loadBatchLogs(batchID) {
        try {
            const [events, progress] = await Promise.all([
                window.go.main.App.GetBatchEvents(batchID, 50),
                window.go.main.App.GetProcessingProgress ? window.go.main.App.GetProcessingProgress(batchID) : Promise.resolve(null)
            ]);

            this.renderLogs(events, progress);
        } catch (error) {
            console.error('Error loading batch logs:', error);
            document.getElementById('logsContainer').innerHTML = `
                <div class="text-center text-red-500 py-8">
                    <p>${window.i18n.t('logs.loadError') || 'Failed to load logs'}: ${error.message}</p>
                </div>
            `;
        }
    }

    renderLogs(events, progress) {
        const container = document.getElementById('logsContainer');
        
        if (events.length === 0) {
            container.innerHTML = `
                <div class="text-center text-gray-500 py-8">
                    <p>${window.i18n.t('logs.noEvents') || 'No events found'}</p>
                </div>
            `;
            return;
        }

        const eventsByType = {
            'batch_start': [],
            'ai_processing': [],
            'ftp_upload': [],
            'batch_complete': [],
            'error': []
        };

        events.forEach(event => {
            if (eventsByType[event.eventType]) {
                eventsByType[event.eventType].push(event);
            } else {
                if (!eventsByType['other']) eventsByType['other'] = [];
                eventsByType['other'].push(event);
            }
        });

        const getEventTypeLabel = (type) => {
            const labels = {
                'batch_start': window.i18n.t('logs.eventTypes.batchStart') || 'Batch Start',
                'ai_processing': window.i18n.t('logs.eventTypes.aiProcessing') || 'AI Processing',
                'ftp_upload': window.i18n.t('logs.eventTypes.ftpUpload') || 'FTP Upload',
                'batch_complete': window.i18n.t('logs.eventTypes.batchComplete') || 'Batch Complete',
                'error': window.i18n.t('logs.eventTypes.error') || 'Error',
                'other': window.i18n.t('logs.eventTypes.other') || 'Other'
            };
            return labels[type] || type;
        };

        const getStatusIcon = (status) => {
            switch (status) {
                case 'success': return '✅';
                case 'failed': return '❌';
                case 'started': return '🚀';
                case 'progress': return '⏳';
                default: return '📝';
            }
        };

        const getStatusColor = (status) => {
            switch (status) {
                case 'success': return 'text-green-600 bg-green-50';
                case 'failed': return 'text-red-600 bg-red-50';
                case 'started': return 'text-blue-600 bg-blue-50';
                case 'progress': return 'text-yellow-600 bg-yellow-50';
                default: return 'text-gray-600 bg-gray-50';
            }
        };

        let html = `
            <div class="mb-6 p-4 bg-gray-50 rounded-lg">
                <h4 class="font-semibold text-gray-900 mb-2">${window.i18n.t('logs.batchSummary') || 'Batch Summary'}</h4>
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div>
                        <span class="text-gray-500">${window.i18n.t('logs.status') || 'Status'}:</span>
                        <span class="ml-1 font-medium">${progress.status}</span>
                    </div>
                    <div>
                        <span class="text-gray-500">${window.i18n.t('logs.progress') || 'Progress'}:</span>
                        <span class="ml-1 font-medium">${progress.overallProgress}%</span>
                    </div>
                    <div>
                        <span class="text-gray-500">${window.i18n.t('logs.totalPhotos') || 'Total Photos'}:</span>
                        <span class="ml-1 font-medium">${progress.totalPhotos}</span>
                    </div>
                    <div>
                        <span class="text-gray-500">${window.i18n.t('logs.currentStep') || 'Current Step'}:</span>
                        <span class="ml-1 font-medium">${progress.currentStep}</span>
                    </div>
                </div>
            </div>
        `;

        Object.keys(eventsByType).forEach(eventType => {
            const typeEvents = eventsByType[eventType];
            if (typeEvents.length === 0) return;

            html += `
                <div class="mb-6">
                    <h4 class="font-semibold text-gray-900 mb-3 flex items-center">
                        <span class="mr-2">${getEventTypeLabel(eventType)}</span>
                        <span class="bg-gray-200 text-gray-600 text-xs px-2 py-1 rounded-full">${typeEvents.length}</span>
                    </h4>
                    <div class="space-y-2">
            `;

            typeEvents.forEach(event => {
                const timeStr = new Date(event.createdAt).toLocaleTimeString();
                html += `
                    <div class="flex items-start space-x-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50">
                        <div class="text-lg">${getStatusIcon(event.status)}</div>
                        <div class="flex-1 min-w-0">
                            <div class="flex items-center justify-between">
                                <p class="text-sm font-medium text-gray-900">${event.message}</p>
                                <div class="flex items-center space-x-2">
                                    <span class="text-xs px-2 py-1 rounded-full ${getStatusColor(event.status)}">${event.status}</span>
                                    <span class="text-xs text-gray-500">${timeStr}</span>
                                </div>
                            </div>
                            ${event.details ? `
                                <details class="mt-2">
                                    <summary class="text-xs text-gray-500 cursor-pointer hover:text-gray-700">
                                        ${window.i18n.t('logs.showDetails') || 'Show details'}
                                    </summary>
                                    <div class="mt-1 p-2 bg-gray-100 rounded text-xs text-gray-700 whitespace-pre-wrap">${event.details}</div>
                                </details>
                            ` : ''}
                            ${event.progress > 0 ? `
                                <div class="mt-2">
                                    <div class="flex justify-between text-xs text-gray-500 mb-1">
                                        <span>${window.i18n.t('logs.progress') || 'Progress'}</span>
                                        <span>${event.progress}%</span>
                                    </div>
                                    <div class="w-full bg-gray-200 rounded-full h-1">
                                        <div class="bg-blue-600 h-1 rounded-full" style="width: ${event.progress}%"></div>
                                    </div>
                                </div>
                            ` : ''}
                        </div>
                    </div>
                `;
            });

            html += `
                    </div>
                </div>
            `;
        });

        container.innerHTML = html;
    }
}

// Инициализация приложения когда Wails готов
window.addEventListener('DOMContentLoaded', async () => {
    console.log('🚀 DOM Content Loaded, initializing app...');
    
    // Проверяем наличие Wails быстро для mock режима
    const quickWailsCheck = () => {
        return !!(window.go && window.go.main && window.go.main.App && 
                 typeof window.go.main.App.GetSettings === 'function');
    };
    
    let wailsAvailable = quickWailsCheck();
    
    if (!wailsAvailable) {
    // Пробуем дождаться события context ready от Wails
    const waitForWailsEvent = () => {
        return new Promise((resolve) => {
            let eventReceived = false;
            
            // Слушаем различные события готовности Wails
            const events = ['wails:ready', 'contextready', 'domready', 'wails:init'];
            
            events.forEach(eventName => {
                window.addEventListener(eventName, () => {
                    console.log(`📡 Received ${eventName} event`);
                    if (!eventReceived) {
                        eventReceived = true;
                        resolve(true);
                    }
                });
            });
            
                // Уменьшенный таймаут для быстрой инициализации mock режима
            setTimeout(() => {
                if (!eventReceived) {
                        console.log('⏰ No Wails events received, checking for mock mode');
                    resolve(false);
                }
                }, 1000); // 1 секунда вместо 2
        });
    };
    
    // Сначала ждем события Wails
    await waitForWailsEvent();
    
        // Проверяем еще раз после события
        wailsAvailable = quickWailsCheck();
        
        if (!wailsAvailable) {
            // Ждем инициализации Wails runtime с уменьшенным таймаутом
    let attempts = 0;
            const maxAttempts = 50; // 5 секунд ожидания вместо 10

    const waitForWails = () => {
        return new Promise((resolve) => {
            const checkWails = () => {
                attempts++;
                
                console.log(`Попытка ${attempts}/${maxAttempts}: Проверяем Wails runtime...`);
                
                const hasGoMethods = window.go && window.go.main && window.go.main.App;
                const hasRuntime = window.runtime;
                const hasWailsContext = window.wails || window.Wails;
                
                // Также проверяем наличие конкретных методов
                const hasSelectFolder = hasGoMethods && typeof window.go.main.App.SelectFolder === 'function';
                
                // Пробуем вызвать простой метод для проверки
                let methodWorks = false;
                if (hasGoMethods) {
        try {
                        // Пробуем вызвать GetDefaultLanguage - он должен работать без параметров
                        window.go.main.App.GetDefaultLanguage().then(() => {
                            console.log('🎯 Go methods are working!');
                        }).catch(() => {
                            console.log('⚠️ Go methods exist but not working yet');
                        });
                        methodWorks = true;
                    } catch (e) {
                        console.log('⚠️ Error calling Go method:', e.message);
        }
                }
                
                if (hasGoMethods && (hasRuntime || hasWailsContext || methodWorks)) {
                    console.log('✅ Wails runtime loaded successfully');
                    resolve(true);
                } else if (attempts >= maxAttempts) {
                            console.log('❌ Wails runtime not available after 5 seconds, running in mock mode');
                    resolve(false);
             } else {
                    setTimeout(checkWails, 100);
                }
            };
            checkWails();
        });
    };
    
            wailsAvailable = await waitForWails();
        }
    }
    
    if (wailsAvailable) {
        // Реальное Wails приложение
        window.app = new StockPhotoApp(true);
        window.app = window.app; // Делаем app доступным глобально для onclick handlers
        console.log('Stock Photo App initialized with Wails backend');
    } else {
        // Mock режим для демонстрации
        console.log('Running in mock mode - creating mock Go methods');
        
        // Создаем заглушки для Wails API
        window.go = {
            main: {
                App: {
                    GetSettings: async () => {
                        console.log('Mock: GetSettings');
                        return Promise.resolve({
                            tempDirectory: './temp',
                            aiProvider: 'openai',
                            aiModel: 'gpt-4o',
                            aiApiKey: '',
                            aiBaseUrl: '',
                            thumbnailSize: 512,
                            maxConcurrentJobs: 3,
                            language: 'en',
                            aiPrompts: {
                                editorial: 'Mock editorial prompt...',
                                commercial: 'Mock commercial prompt...'
                            }
                        });
                    },
                    SaveSettings: async (settings) => {
                        console.log('Mock: SaveSettings', settings);
                        return Promise.resolve();
                    },
                    ProcessPhotoFolder: async (folderPath, description, photoType) => {
                        console.log('Mock: ProcessPhotoFolder', folderPath, description, photoType);
                        
                        // Имитируем ошибку в mock режиме для демонстрации
                        if (!folderPath || folderPath.trim() === '') {
                            throw new Error(window.i18n.t('notifications.folderSelectionDesktopOnly'));
                        }
                        
                        // Если все проверки прошли успешно
                        return Promise.resolve();
                    },
                    GetProcessingHistory: async (limit) => {
                        console.log('Mock: GetProcessingHistory', limit);
                        return Promise.resolve([]);
                    },
                    GetDefaultLanguage: async () => {
                        console.log('Mock: GetDefaultLanguage');
                        return Promise.resolve('en');
                    },
                    SelectFolder: async () => {
                        console.log('Mock: SelectFolder - ' + window.i18n.t('notifications.appModeOnly'));
                        throw new Error(window.i18n.t('notifications.folderSelectionDesktopOnly'));
                    },
                    GetFolderContents: async (folderPath) => {
                        console.log('Mock: GetFolderContents', folderPath);
                        return Promise.resolve([
                            {
                                name: 'photo1.jpg',
                                path: folderPath + '/photo1.jpg',
                                size: 2048576,
                                extension: '.jpg',
                                isValid: true
                            },
                            {
                                name: 'photo2.png',
                                path: folderPath + '/photo2.png',
                                size: 1024768,
                                extension: '.png',
                                isValid: true
                            },
                            {
                                name: 'document.txt',
                                path: folderPath + '/document.txt',
                                size: 1024,
                                extension: '.txt',
                                isValid: false
                            }
                        ]);
                    },
                    GetAIModels: async (provider) => {
                        console.log('Mock: GetAIModels', provider);
                        if (provider === 'openai') {
                            return Promise.resolve([
                                {
                                    id: 'o1',
                                    name: 'o1',
                                    description: 'Most advanced reasoning model for complex tasks',
                                    maxTokens: 100000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'o1-mini',
                                    name: 'o1-mini',
                                    description: 'Faster reasoning model for coding and math',
                                    maxTokens: 65536,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'o1-preview',
                                    name: 'o1-preview',
                                    description: 'Preview of advanced reasoning capabilities',
                                    maxTokens: 32768,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'o3-mini',
                                    name: 'o3-mini (January 2025)',
                                    description: 'Advanced reasoning model, successor to o1-mini',
                                    maxTokens: 65536,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4o',
                                    name: 'GPT-4o',
                                    description: 'High-intelligence flagship model for complex tasks',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4o-2024-11-20',
                                    name: 'GPT-4o (November 2024)',
                                    description: 'GPT-4o model with vision capabilities and enhanced performance',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4o-mini',
                                    name: 'GPT-4o mini',
                                    description: 'Affordable and intelligent small model for fast tasks',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4-turbo',
                                    name: 'GPT-4 Turbo',
                                    description: 'GPT-4 Turbo with enhanced capabilities and vision',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4.1',
                                    name: 'GPT-4.1 (April 2025)',
                                    description: 'Next-generation GPT-4 model with enhanced features',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4.5-preview',
                                    name: 'GPT-4.5 (February 2025)',
                                    description: 'Enhanced GPT-4 model with improved capabilities',
                                    maxTokens: 128000,
                                    supportsVision: true,
                                    provider: 'openai'
                                },
                                {
                                    id: 'gpt-4',
                                    name: 'GPT-4',
                                    description: 'Advanced GPT-4 model with multimodal capabilities',
                                    maxTokens: 8192
                                }
                            ]);
                        } else if (provider === 'claude') {
                            return Promise.resolve([
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
                            ]);
                        }
                        return Promise.resolve([]);
                    }
                }
            }
        };
        
        window.app = new StockPhotoApp(false);
        window.app = window.app; // Делаем app доступным глобально для onclick handlers
        console.log('Stock Photo App initialized with mock backend');
    }
});