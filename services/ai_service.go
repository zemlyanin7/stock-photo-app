package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"stock-photo-app/models"
	"strings"
	"time"
)

type AIService struct {
	httpClient    *http.Client
	exifProcessor *EXIFProcessor
}

func NewAIService() *AIService {
	return &AIService{
		httpClient:    &http.Client{}, // таймаут будет устанавливаться динамически
		exifProcessor: NewEXIFProcessor(),
	}
}

// OpenAI API структуры для Structured Outputs
type ResponseFormat struct {
	Type       string     `json:"type"`
	JSONSchema JSONSchema `json:"json_schema"`
}

type JSONSchema struct {
	Name   string `json:"name"`
	Schema Schema `json:"schema"`
	Strict bool   `json:"strict"`
}

type Schema struct {
	Type                 string              `json:"type"`
	Properties           map[string]Property `json:"properties"`
	Required             []string            `json:"required"`
	AdditionalProperties bool                `json:"additionalProperties"`
}

type Property struct {
	Type        string    `json:"type"`
	Description string    `json:"description,omitempty"`
	Items       *Property `json:"items,omitempty"`
}

type OpenAIRequest struct {
	Model               string          `json:"model"`
	Messages            []Message       `json:"messages"`
	MaxCompletionTokens int             `json:"max_completion_tokens"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// AnalyzePhoto отправляет фото на анализ в AI с учетом типа контента
func (s *AIService) AnalyzePhoto(photo models.Photo, description string, contentType string, settings models.AppSettings) (*models.AIResult, error) {
	// Выбираем промпт на основе типа контента
	prompt := ""
	if settings.AIPrompts != nil {
		if p, exists := settings.AIPrompts[contentType]; exists {
			prompt = p
		}
	}

	// Используем fallback промпт если не найден
	if prompt == "" {
		if contentType == "editorial" {
			prompt = "Analyze this editorial photograph and create factual metadata with specific location and time information."
		} else {
			prompt = "Analyze this commercial photograph and create universal metadata suitable for advertising use."
		}
	}

	switch settings.AIProvider {
	case "openai":
		return s.analyzeWithOpenAI(photo, description, prompt, contentType, settings)
	case "claude":
		return s.analyzeWithClaude(photo, description, prompt, contentType, settings)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", settings.AIProvider)
	}
}

// analyzeWithOpenAI анализирует изображение через OpenAI API с retry логикой
func (s *AIService) analyzeWithOpenAI(photo models.Photo, description string, prompt string, contentType string, settings models.AppSettings) (*models.AIResult, error) {
	const maxRetries = 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := s.analyzePhotoAttempt(photo, description, prompt, contentType, settings)
		if err == nil {
			return result, nil
		}

		log.Printf("AI analysis attempt %d/%d failed for photo %s: %v", attempt, maxRetries, photo.FileName, err)

		// Если это последняя попытка или критическая ошибка, возвращаем ошибку
		if attempt == maxRetries || !s.isRetryableError(err) {
			return nil, err
		}

		// Пауза между попытками
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	return nil, fmt.Errorf("all retry attempts failed")
}

// analyzePhotoAttempt выполняет одну попытку анализа фото
func (s *AIService) analyzePhotoAttempt(photo models.Photo, description string, prompt string, contentType string, settings models.AppSettings) (*models.AIResult, error) {
	// Кодируем изображение в base64
	imageProcessor := NewImageProcessor(settings.TempDirectory)
	base64Image, err := imageProcessor.EncodeImageToBase64(photo.ThumbnailPath)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	// Формируем полный промпт с учетом типа контента
	fullPrompt := s.exifProcessor.BuildContextualPrompt(contentType, prompt, photo.ExifData)

	// Создаем запрос
	model := settings.AIModel
	if model == "" {
		model = "gpt-4o" // fallback
	}

	// Создаем JSON Schema для структурированного вывода
	responseFormat := &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchema{
			Name:   "photo_metadata",
			Strict: true,
			Schema: Schema{
				Type:                 "object",
				AdditionalProperties: false,
				Required:             []string{"title", "description", "keywords", "category"},
				Properties: map[string]Property{
					"title": {
						Type:        "string",
						Description: "Название фотографии (до 100 символов)",
					},
					"description": {
						Type:        "string",
						Description: "Описание фотографии (до 500 символов)",
					},
					"keywords": {
						Type:        "array",
						Description: "Массив ключевых слов (48-55 слов)",
						Items: &Property{
							Type: "string",
						},
					},
					"category": {
						Type:        "string",
						Description: "Категория фотографии из списка стандартных категорий стоков",
					},
				},
			},
		},
	}

	// Устанавливаем максимальное количество токенов из настроек
	maxTokens := 2000 // значение по умолчанию
	if settings.AIMaxTokens > 0 {
		maxTokens = settings.AIMaxTokens
	}

	request := OpenAIRequest{
		Model:               model,
		MaxCompletionTokens: maxTokens,
		ResponseFormat:      responseFormat,
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: fullPrompt,
					},
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
						},
					},
				},
			},
		},
	}

	// Отправляем запрос
	response, err := s.sendOpenAIRequest(request, settings)
	if err != nil {
		return nil, err
	}

	// Парсим ответ
	result, err := s.parseAIResponse(response.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}

	// Устанавливаем тип контента в результате
	result.ContentType = contentType
	return result, nil
}

// analyzeWithClaude анализирует изображение через Claude API
func (s *AIService) analyzeWithClaude(photo models.Photo, description string, prompt string, contentType string, settings models.AppSettings) (*models.AIResult, error) {
	// TODO: Реализовать интеграцию с Claude API
	// Пока что заглушка
	log.Printf("Claude API integration not implemented yet")

	// Возвращаем тестовый результат
	return &models.AIResult{
		ContentType: contentType,
		Title:       "Test Photo Title",
		Keywords:    []string{"test", "photo", "stock"},
		Quality:     85,
		Description: "Test AI analysis result",
		Category:    "general",
		Processed:   true,
	}, nil
}

// buildFullPrompt строит полный промпт для AI
func (s *AIService) buildFullPrompt(basePrompt string, description string, exifData map[string]string) string {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(basePrompt)
	promptBuilder.WriteString("\n\nPhoto Description: ")
	promptBuilder.WriteString(description)

	if len(exifData) > 0 {
		promptBuilder.WriteString("\n\nEXIF Data:")
		for key, value := range exifData {
			if value != "" {
				promptBuilder.WriteString(fmt.Sprintf("\n- %s: %s", key, value))
			}
		}
	}

	promptBuilder.WriteString("\n\nPlease analyze this image and return a JSON response with the following structure:")
	promptBuilder.WriteString("\n{")
	promptBuilder.WriteString("\n  \"title\": \"Brief, descriptive title for stock photo\",")
	promptBuilder.WriteString("\n  \"keywords\": [\"keyword1\", \"keyword2\", \"keyword3\", \"...\"],")
	promptBuilder.WriteString("\n  \"quality\": 85,")
	promptBuilder.WriteString("\n  \"description\": \"Detailed description for stock site\",")
	promptBuilder.WriteString("\n  \"category\": \"appropriate category from the list\"")
	promptBuilder.WriteString("\n}")
	promptBuilder.WriteString("\n\nIMPORTANT: Generate 48-55 keywords and select category from the provided list in the prompt.")

	return promptBuilder.String()
}

// sendOpenAIRequest отправляет запрос к OpenAI API
func (s *AIService) sendOpenAIRequest(request OpenAIRequest, settings models.AppSettings) (*OpenAIResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	apiURL := settings.AIBaseURL
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	// Устанавливаем таймаут из настроек
	timeout := 90 * time.Second // значение по умолчанию
	if settings.AITimeout > 0 {
		timeout = time.Duration(settings.AITimeout) * time.Second
	}

	// Создаем клиент с таймаутом для этого запроса
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.AIAPIKey)

	// Retry логика для AI запросов
	var resp *http.Response
	var body []byte
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Создаем новый request для каждой попытки (так как body может быть прочитан)
		reqCopy, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request copy: %w", err)
		}
		reqCopy.Header.Set("Content-Type", "application/json")
		reqCopy.Header.Set("Authorization", "Bearer "+settings.AIAPIKey)

		resp, err = client.Do(reqCopy)
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to send request after %d attempts: %w", maxRetries, err)
			}
			log.Printf("AI request attempt %d failed: %v, retrying...", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second) // экспоненциальная задержка
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("failed to read response after %d attempts: %w", maxRetries, err)
			}
			log.Printf("Failed to read response on attempt %d: %v, retrying...", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		// Проверяем код ответа
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			if attempt == maxRetries {
				return nil, fmt.Errorf("server error after %d attempts: HTTP %d: %s", maxRetries, resp.StatusCode, string(body))
			}
			log.Printf("Server error (HTTP %d) on attempt %d, retrying...", resp.StatusCode, attempt)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		if resp.StatusCode == 429 { // Rate limit
			if attempt == maxRetries {
				return nil, fmt.Errorf("rate limit exceeded after %d attempts: %s", maxRetries, string(body))
			}
			log.Printf("Rate limit hit on attempt %d, retrying...", attempt)
			time.Sleep(time.Duration(attempt*2) * time.Second) // более долгая задержка для rate limit
			continue
		}

		// Если получили не server error и не rate limit, выходим из цикла
		break
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("API error: %s", response.Error.Message)
	}

	return &response, nil
}

// parseAIResponse парсит ответ от AI и извлекает JSON
func (s *AIService) parseAIResponse(content string) (*models.AIResult, error) {
	log.Printf("DEBUG: AI response content: %s", content)

	// Проверяем пустой ответ
	if strings.TrimSpace(content) == "" {
		log.Printf("ERROR: AI returned empty response")
		return nil, fmt.Errorf("AI returned empty response - please check your API key and quota")
	}

	// С Structured Outputs ответ должен быть чистым JSON
	var aiResponse models.AIResponse
	err := json.Unmarshal([]byte(content), &aiResponse)
	if err != nil {
		// Если прямой парсинг не удался, попробуем найти JSON в тексте (fallback для старых моделей)
		startIdx := strings.Index(content, "{")
		endIdx := strings.LastIndex(content, "}")

		if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
			log.Printf("ERROR: Failed to parse AI response. Content: %s", content)
			// Возвращаем более информативную ошибку
			if len(content) > 100 {
				return nil, fmt.Errorf("no valid JSON found in AI response (first 100 chars): %s...", content[:100])
			}
			return nil, fmt.Errorf("no valid JSON found in AI response: %s", content)
		}

		jsonStr := content[startIdx : endIdx+1]
		err = json.Unmarshal([]byte(jsonStr), &aiResponse)
		if err != nil {
			log.Printf("ERROR: Failed to parse extracted JSON. JSON: %s, Error: %v", jsonStr, err)
			return nil, fmt.Errorf("failed to parse AI response JSON: %w", err)
		}
	}

	result := &models.AIResult{
		Title:       aiResponse.Title,
		Keywords:    aiResponse.Keywords,
		Quality:     aiResponse.Quality,
		Description: aiResponse.Description,
		Category:    aiResponse.Category,
		Processed:   true,
	}

	if aiResponse.Error != "" {
		result.Error = aiResponse.Error
		result.Processed = false
	}

	return result, nil
}

// TestConnection тестирует подключение к AI API
func (s *AIService) TestConnection(settings models.AppSettings) error {
	switch settings.AIProvider {
	case "openai":
		return s.testOpenAIConnection(settings)
	case "claude":
		return s.testClaudeConnection(settings)
	default:
		return fmt.Errorf("unsupported AI provider: %s", settings.AIProvider)
	}
}

// testOpenAIConnection тестирует подключение к OpenAI
func (s *AIService) testOpenAIConnection(settings models.AppSettings) error {
	apiURL := settings.AIBaseURL
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/models"
	} else {
		apiURL = strings.TrimSuffix(apiURL, "/chat/completions") + "/models"
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+settings.AIAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to AI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// testClaudeConnection тестирует подключение к Claude
func (s *AIService) testClaudeConnection(settings models.AppSettings) error {
	// TODO: Реализовать тестирование Claude API
	log.Printf("Claude API connection test not implemented yet")
	return nil
}

// isRetryableError проверяет является ли ошибка подходящей для повтора
func (s *AIService) isRetryableError(err error) bool {
	errorMsg := strings.ToLower(err.Error())

	// Ошибки которые стоит повторить
	retryableErrors := []string{
		"empty response",
		"no valid json",
		"failed to parse",
		"connection",
		"timeout",
		"rate limit",
		"server error",
		"service unavailable",
		"internal error",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errorMsg, retryable) {
			return true
		}
	}

	return false
}

// GetAvailableModels возвращает список доступных моделей для указанного провайдера
func (s *AIService) GetAvailableModels(provider string, settings models.AppSettings) ([]models.AIModel, error) {
	switch provider {
	case "openai":
		return s.getOpenAIModels(settings)
	case "claude":
		return s.getClaudeModels(settings)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", provider)
	}
}

// getOpenAIModels получает список моделей OpenAI
func (s *AIService) getOpenAIModels(settings models.AppSettings) ([]models.AIModel, error) {
	// Если есть API ключ, сначала пытаемся получить актуальный список с API
	if settings.AIAPIKey != "" {
		actualModels, err := s.fetchOpenAIModels(settings)
		if err == nil && len(actualModels) > 0 {
			log.Printf("Successfully fetched %d models from OpenAI API", len(actualModels))
			return actualModels, nil
		}
		log.Printf("Failed to fetch OpenAI models from API: %v, using static fallback", err)
	} else {
		log.Printf("No API key provided, using static model list")
	}

	// Fallback: статический список актуальных моделей
	staticModels := []models.AIModel{
		// O1 Series (новейшие модели рассуждений)
		{
			ID:             "o1",
			Name:           "o1",
			Description:    "Most advanced reasoning model for complex tasks",
			MaxTokens:      100000,
			SupportsVision: true,
			Provider:       "openai",
		},
		{
			ID:             "o1-mini",
			Name:           "o1-mini",
			Description:    "Faster reasoning model for coding and math",
			MaxTokens:      65536,
			SupportsVision: true,
			Provider:       "openai",
		},
		{
			ID:             "o1-preview",
			Name:           "o1-preview",
			Description:    "Preview of advanced reasoning capabilities",
			MaxTokens:      32768,
			SupportsVision: true,
			Provider:       "openai",
		},
		// GPT-4o Series (latest flagship models)
		{
			ID:             "gpt-4o",
			Name:           "GPT-4o",
			Description:    "High-intelligence flagship model for complex tasks",
			MaxTokens:      128000,
			SupportsVision: true,
			Provider:       "openai",
		},
		{
			ID:             "gpt-4o-2024-11-20",
			Name:           "GPT-4o (November 2024)",
			Description:    "Latest GPT-4o model with improved capabilities",
			MaxTokens:      128000,
			SupportsVision: true,
			Provider:       "openai",
		},
		{
			ID:             "gpt-4o-mini",
			Name:           "GPT-4o mini",
			Description:    "Affordable and intelligent small model for fast tasks",
			MaxTokens:      128000,
			SupportsVision: true,
			Provider:       "openai",
		},
		{
			ID:             "gpt-4-turbo",
			Name:           "GPT-4 Turbo",
			Description:    "Latest GPT-4 Turbo model with vision",
			MaxTokens:      128000,
			SupportsVision: true,
			Provider:       "openai",
		},
	}

	return staticModels, nil
}

// fetchOpenAIModels получает актуальный список моделей через API OpenAI
func (s *AIService) fetchOpenAIModels(settings models.AppSettings) ([]models.AIModel, error) {
	apiURL := settings.AIBaseURL
	if apiURL == "" {
		apiURL = "https://api.openai.com/v1/models"
	} else {
		apiURL = strings.TrimSuffix(apiURL, "/chat/completions") + "/models"
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+settings.AIAPIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var aiModels []models.AIModel
	for _, model := range response.Data {
		// Фильтруем модели, подходящие для анализа изображений
		if s.isVisionCapableModel(model.ID) {
			aiModel := s.convertToAIModel(model.ID)
			aiModels = append(aiModels, aiModel)
		}
	}

	log.Printf("Fetched %d vision-capable models from OpenAI API", len(aiModels))
	return aiModels, nil
}

// isVisionCapableModel проверяет, поддерживает ли модель анализ изображений
func (s *AIService) isVisionCapableModel(modelID string) bool {
	// Исключаем модели, которые точно не подходят для анализа изображений
	excludePatterns := []string{
		"dall-e",         // Image generation models
		"tts-",           // Text-to-speech models
		"whisper",        // Speech-to-text models
		"text-embedding", // Embedding models
		"babbage",        // Legacy completion models
		"davinci",        // Legacy completion models
		"moderation",     // Moderation models
		"instruct",       // Instruction following models (не vision)
		"realtime",       // Realtime models (audio)
		"audio",          // Audio models
		"transcribe",     // Transcription models
		"computer-use",   // Computer use models
		"search",         // Search models
	}

	for _, pattern := range excludePatterns {
		if strings.Contains(modelID, pattern) {
			return false
		}
	}

	// Включаем модели, которые поддерживают vision

	// O-series (reasoning models with vision) - все новые O модели
	if strings.HasPrefix(modelID, "o1") || strings.HasPrefix(modelID, "o3") ||
		strings.HasPrefix(modelID, "o4") {
		return true
	}

	// GPT-4.x series (новые модели)
	if strings.HasPrefix(modelID, "gpt-4.") {
		return true
	}

	// GPT-4o series (все варианты поддерживают vision)
	if strings.Contains(modelID, "gpt-4o") && !strings.Contains(modelID, "audio") &&
		!strings.Contains(modelID, "realtime") && !strings.Contains(modelID, "transcribe") &&
		!strings.Contains(modelID, "search") {
		return true
	}

	// GPT-4 Turbo models (с vision)
	if strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "turbo") {
		return true
	}

	// GPT-4 Vision models (explicit vision models)
	if strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "vision") {
		return true
	}

	// GPT-4 dated models (2024 models обычно поддерживают vision)
	if strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "2024") {
		return true
	}

	// Классические GPT-4 models (базовые модели с vision)
	if modelID == "gpt-4" || modelID == "gpt-4-0613" {
		return true
	}

	// ChatGPT-4o (специальная модель)
	if strings.Contains(modelID, "chatgpt-4o") {
		return true
	}

	return false
}

// convertToAIModel преобразует ID модели в структуру AIModel с правильными метаданными
func (s *AIService) convertToAIModel(modelID string) models.AIModel {
	model := models.AIModel{
		ID:             modelID,
		Name:           s.getModelDisplayName(modelID),
		Description:    s.getModelDescription(modelID),
		MaxTokens:      s.getModelMaxTokens(modelID),
		SupportsVision: true,
		Provider:       "openai",
	}
	return model
}

// getModelDisplayName возвращает красивое название модели
func (s *AIService) getModelDisplayName(modelID string) string {
	switch {
	// O-series models
	case modelID == "o1":
		return "o1"
	case modelID == "o1-mini":
		return "o1-mini"
	case modelID == "o1-preview":
		return "o1-preview"
	case strings.HasPrefix(modelID, "o1-pro"):
		return "o1-pro (" + s.extractDateFromModel(modelID) + ")"
	case strings.HasPrefix(modelID, "o1-"):
		return "o1 (" + s.extractDateFromModel(modelID) + ")"
	case strings.HasPrefix(modelID, "o3-mini"):
		return "o3-mini (" + s.extractDateFromModel(modelID) + ")"
	case strings.HasPrefix(modelID, "o4-mini"):
		return "o4-mini (" + s.extractDateFromModel(modelID) + ")"

	// GPT-4.x series (новые модели)
	case strings.HasPrefix(modelID, "gpt-4.5"):
		return "GPT-4.5 (" + s.extractDateFromModel(modelID) + ")"
	case strings.HasPrefix(modelID, "gpt-4.1"):
		if strings.Contains(modelID, "mini") {
			return "GPT-4.1 mini (" + s.extractDateFromModel(modelID) + ")"
		} else if strings.Contains(modelID, "nano") {
			return "GPT-4.1 nano (" + s.extractDateFromModel(modelID) + ")"
		}
		return "GPT-4.1 (" + s.extractDateFromModel(modelID) + ")"

	// GPT-4o series
	case modelID == "gpt-4o":
		return "GPT-4o"
	case strings.Contains(modelID, "gpt-4o-mini") && !strings.Contains(modelID, "2024"):
		return "GPT-4o mini"
	case strings.Contains(modelID, "gpt-4o-mini"):
		return "GPT-4o mini (" + s.extractDateFromModel(modelID) + ")"
	case strings.Contains(modelID, "chatgpt-4o"):
		return "ChatGPT-4o (" + s.extractDateFromModel(modelID) + ")"
	case strings.Contains(modelID, "gpt-4o"):
		return "GPT-4o (" + s.extractDateFromModel(modelID) + ")"

	// GPT-4 Turbo series
	case strings.Contains(modelID, "gpt-4-turbo") && !strings.Contains(modelID, "2024"):
		return "GPT-4 Turbo"
	case strings.Contains(modelID, "gpt-4-turbo"):
		return "GPT-4 Turbo (" + s.extractDateFromModel(modelID) + ")"
	case strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "preview"):
		return "GPT-4 Turbo Preview (" + s.extractDateFromModel(modelID) + ")"

	// GPT-4 Vision series
	case strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "vision"):
		return "GPT-4 Vision (" + s.extractDateFromModel(modelID) + ")"

	// Classic GPT-4
	case modelID == "gpt-4":
		return "GPT-4"
	case modelID == "gpt-4-0613":
		return "GPT-4 (June 2023)"

	// Special models
	case strings.Contains(modelID, "gpt-image"):
		return "GPT Image (" + s.extractDateFromModel(modelID) + ")"
	case strings.Contains(modelID, "codex"):
		return "Codex (" + s.extractDateFromModel(modelID) + ")"

	// GPT-3.5 (включаем только если прошли фильтрацию)
	case strings.Contains(modelID, "gpt-3.5-turbo") && !strings.Contains(modelID, "instruct"):
		return "GPT-3.5 Turbo (" + s.extractDateFromModel(modelID) + ")"

	default:
		return strings.ToUpper(modelID)
	}
}

// getModelDescription возвращает описание модели
func (s *AIService) getModelDescription(modelID string) string {
	switch {
	// O-series models
	case modelID == "o1":
		return "Most advanced reasoning model for complex tasks"
	case modelID == "o1-mini":
		return "Faster reasoning model for coding and math"
	case modelID == "o1-preview":
		return "Preview of advanced reasoning capabilities"
	case strings.HasPrefix(modelID, "o1-pro"):
		return "Professional-grade reasoning model with enhanced capabilities"
	case strings.HasPrefix(modelID, "o3-mini"):
		return "Advanced reasoning model, successor to o1-mini"
	case strings.HasPrefix(modelID, "o4-mini"):
		return "Next-generation reasoning model with improved performance"

	// GPT-4.x series
	case strings.HasPrefix(modelID, "gpt-4.5"):
		return "Enhanced GPT-4 model with improved capabilities"
	case strings.Contains(modelID, "gpt-4.1-nano"):
		return "Ultra-lightweight GPT-4.1 model for fast tasks"
	case strings.Contains(modelID, "gpt-4.1-mini"):
		return "Compact GPT-4.1 model for efficient processing"
	case strings.HasPrefix(modelID, "gpt-4.1"):
		return "Next-generation GPT-4 model with enhanced features"

	// GPT-4o series
	case modelID == "gpt-4o":
		return "High-intelligence flagship model for complex tasks"
	case strings.Contains(modelID, "gpt-4o-mini"):
		return "Affordable and intelligent small model for fast tasks"
	case strings.Contains(modelID, "chatgpt-4o"):
		return "ChatGPT-optimized version of GPT-4o"
	case strings.Contains(modelID, "gpt-4o"):
		return "GPT-4o model with vision capabilities and enhanced performance"

	// GPT-4 Turbo series
	case strings.Contains(modelID, "gpt-4-turbo"):
		return "GPT-4 Turbo with enhanced capabilities and vision"
	case strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "preview"):
		return "Preview version of GPT-4 Turbo with latest improvements"

	// GPT-4 Vision series
	case strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "vision"):
		return "GPT-4 model with vision capabilities"

	// Classic GPT-4
	case modelID == "gpt-4" || modelID == "gpt-4-0613":
		return "Advanced GPT-4 model with multimodal capabilities"

	// Special models
	case strings.Contains(modelID, "gpt-image"):
		return "Specialized model for image understanding and generation"
	case strings.Contains(modelID, "codex"):
		return "Code-specialized model for programming tasks"

	// GPT-3.5
	case strings.Contains(modelID, "gpt-3.5-turbo"):
		return "Fast and efficient model for general tasks"

	default:
		return fmt.Sprintf("OpenAI %s model with vision support", modelID)
	}
}

// getModelMaxTokens возвращает максимальное количество токенов для модели
func (s *AIService) getModelMaxTokens(modelID string) int {
	switch {
	// O-series models
	case modelID == "o1":
		return 100000
	case modelID == "o1-mini":
		return 65536
	case modelID == "o1-preview":
		return 32768
	case strings.HasPrefix(modelID, "o1-pro"):
		return 128000
	case strings.HasPrefix(modelID, "o3-mini"):
		return 65536
	case strings.HasPrefix(modelID, "o4-mini"):
		return 65536

	// GPT-4.x series
	case strings.Contains(modelID, "gpt-4.1-nano"):
		return 32768
	case strings.Contains(modelID, "gpt-4.1-mini"):
		return 65536
	case strings.HasPrefix(modelID, "gpt-4.1") || strings.HasPrefix(modelID, "gpt-4.5"):
		return 128000

	// GPT-4o series - все имеют 128K контекст
	case strings.Contains(modelID, "gpt-4o"):
		return 128000

	// GPT-4 Turbo series - 128K контекст
	case strings.Contains(modelID, "gpt-4-turbo") ||
		(strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "preview")):
		return 128000

	// GPT-4 Vision models - зависит от версии
	case strings.Contains(modelID, "gpt-4") && strings.Contains(modelID, "vision"):
		if strings.Contains(modelID, "preview") {
			return 4096 // Ранние vision модели
		}
		return 128000

	// Classic GPT-4
	case modelID == "gpt-4":
		return 8192
	case modelID == "gpt-4-0613":
		return 8192

	// Special models
	case strings.Contains(modelID, "gpt-image"):
		return 32768
	case strings.Contains(modelID, "codex"):
		return 8192

	// GPT-3.5 series
	case strings.Contains(modelID, "gpt-3.5-turbo"):
		if strings.Contains(modelID, "16k") {
			return 16384
		}
		return 4096

	default:
		return 4096 // Консервативная оценка
	}
}

// extractDateFromModel извлекает дату из ID модели для отображения
func (s *AIService) extractDateFromModel(modelID string) string {
	// 2025 dates
	if strings.Contains(modelID, "2025-04") {
		return "April 2025"
	} else if strings.Contains(modelID, "2025-03") {
		return "March 2025"
	} else if strings.Contains(modelID, "2025-02") {
		return "February 2025"
	} else if strings.Contains(modelID, "2025-01") {
		return "January 2025"
	}

	// 2024 dates
	if strings.Contains(modelID, "2024-12") {
		return "December 2024"
	} else if strings.Contains(modelID, "2024-11") {
		return "November 2024"
	} else if strings.Contains(modelID, "2024-10") {
		return "October 2024"
	} else if strings.Contains(modelID, "2024-09") {
		return "September 2024"
	} else if strings.Contains(modelID, "2024-08") {
		return "August 2024"
	} else if strings.Contains(modelID, "2024-07") {
		return "July 2024"
	} else if strings.Contains(modelID, "2024-06") {
		return "June 2024"
	} else if strings.Contains(modelID, "2024-05") {
		return "May 2024"
	} else if strings.Contains(modelID, "2024-04") {
		return "April 2024"
	} else if strings.Contains(modelID, "2024-03") {
		return "March 2024"
	} else if strings.Contains(modelID, "2024-02") {
		return "February 2024"
	} else if strings.Contains(modelID, "2024-01") {
		return "January 2024"
	}

	// 2023 dates
	if strings.Contains(modelID, "1106") {
		return "November 2023"
	} else if strings.Contains(modelID, "0613") {
		return "June 2023"
	} else if strings.Contains(modelID, "0125") {
		return "January 2024"
	}

	// Special handling for specific model patterns
	if strings.Contains(modelID, "latest") {
		return "latest"
	}

	return "latest"
}

// getClaudeModels получает список моделей Claude
func (s *AIService) getClaudeModels(settings models.AppSettings) ([]models.AIModel, error) {
	// Если есть API ключ, сначала пытаемся получить актуальный список с API
	if settings.AIAPIKey != "" {
		actualModels, err := s.fetchClaudeModels(settings)
		if err == nil && len(actualModels) > 0 {
			log.Printf("Successfully fetched %d models from Claude API", len(actualModels))
			return actualModels, nil
		}
		log.Printf("Failed to fetch Claude models from API: %v, using static fallback", err)
	} else {
		log.Printf("No API key provided, using static Claude model list")
	}

	// Fallback: статический список актуальных моделей
	staticModels := []models.AIModel{
		// Claude 4 Series (новейшие модели)
		{
			ID:             "claude-opus-4-20250514",
			Name:           "Claude Opus 4",
			Description:    "Most capable and intelligent model with superior reasoning capabilities",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
		{
			ID:             "claude-sonnet-4-20250514",
			Name:           "Claude Sonnet 4",
			Description:    "High-performance model with exceptional reasoning and efficiency",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},

		// Claude 3.7 Series
		{
			ID:             "claude-3-7-sonnet-20250219",
			Name:           "Claude 3.7 Sonnet",
			Description:    "High-performance model with early extended thinking capabilities",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},

		// Claude 3.5 Series (актуальные версии)
		{
			ID:             "claude-3-5-sonnet-20241022",
			Name:           "Claude 3.5 Sonnet v2",
			Description:    "Most intelligent model with enhanced vision capabilities (Latest)",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
		{
			ID:             "claude-3-5-sonnet-20240620",
			Name:           "Claude 3.5 Sonnet v1",
			Description:    "Original Claude 3.5 Sonnet with advanced capabilities",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
		{
			ID:             "claude-3-5-haiku-20241022",
			Name:           "Claude 3.5 Haiku",
			Description:    "Fastest model with vision capabilities and high intelligence",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},

		// Claude 3 Series (legacy но все еще актуальные)
		{
			ID:             "claude-3-opus-20240229",
			Name:           "Claude 3 Opus",
			Description:    "Most powerful model for complex tasks with exceptional reasoning",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
		{
			ID:             "claude-3-sonnet-20240229",
			Name:           "Claude 3 Sonnet",
			Description:    "Balanced model for general tasks with strong performance",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
		{
			ID:             "claude-3-haiku-20240307",
			Name:           "Claude 3 Haiku",
			Description:    "Fast and cost-effective model for quick responses",
			MaxTokens:      200000,
			SupportsVision: true,
			Provider:       "claude",
		},
	}

	return staticModels, nil
}

// fetchClaudeModels получает актуальный список моделей через API Claude
func (s *AIService) fetchClaudeModels(settings models.AppSettings) ([]models.AIModel, error) {
	apiURL := settings.AIBaseURL
	if apiURL == "" {
		apiURL = "https://api.anthropic.com/v1/models"
	} else {
		apiURL = strings.TrimSuffix(apiURL, "/v1/messages") + "/v1/models"
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", settings.AIAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Data []struct {
			ID          string `json:"id"`
			DisplayName string `json:"display_name"`
			CreatedAt   string `json:"created_at"`
			Type        string `json:"type"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var aiModels []models.AIModel
	for _, model := range response.Data {
		// Все модели Claude поддерживают vision и text
		aiModel := s.convertClaudeToAIModel(model.ID, model.DisplayName)
		aiModels = append(aiModels, aiModel)
	}

	log.Printf("Fetched %d models from Claude API", len(aiModels))
	return aiModels, nil
}

// convertClaudeToAIModel преобразует Claude модель в структуру AIModel
func (s *AIService) convertClaudeToAIModel(modelID, displayName string) models.AIModel {
	model := models.AIModel{
		ID:             modelID,
		Name:           s.getClaudeDisplayName(modelID, displayName),
		Description:    s.getClaudeDescription(modelID),
		MaxTokens:      s.getClaudeMaxTokens(modelID),
		SupportsVision: true, // Все современные модели Claude поддерживают vision
		Provider:       "claude",
	}
	return model
}

// getClaudeDisplayName возвращает красивое название модели Claude
func (s *AIService) getClaudeDisplayName(modelID, displayName string) string {
	// Если есть display_name из API, используем его
	if displayName != "" && displayName != modelID {
		return displayName
	}

	// Иначе генерируем на основе ID
	switch {
	case strings.Contains(modelID, "claude-opus-4"):
		return "Claude Opus 4"
	case strings.Contains(modelID, "claude-sonnet-4"):
		return "Claude Sonnet 4"
	case strings.Contains(modelID, "claude-3-7-sonnet"):
		return "Claude 3.7 Sonnet"
	case strings.Contains(modelID, "claude-3-5-sonnet"):
		if strings.Contains(modelID, "20241022") {
			return "Claude 3.5 Sonnet v2"
		}
		return "Claude 3.5 Sonnet"
	case strings.Contains(modelID, "claude-3-5-haiku"):
		return "Claude 3.5 Haiku"
	case strings.Contains(modelID, "claude-3-opus"):
		return "Claude 3 Opus"
	case strings.Contains(modelID, "claude-3-sonnet"):
		return "Claude 3 Sonnet"
	case strings.Contains(modelID, "claude-3-haiku"):
		return "Claude 3 Haiku"
	default:
		return strings.ToUpper(modelID)
	}
}

// getClaudeDescription возвращает описание модели Claude
func (s *AIService) getClaudeDescription(modelID string) string {
	switch {
	case strings.Contains(modelID, "claude-opus-4"):
		return "Most capable and intelligent model with superior reasoning capabilities"
	case strings.Contains(modelID, "claude-sonnet-4"):
		return "High-performance model with exceptional reasoning and efficiency"
	case strings.Contains(modelID, "claude-3-7-sonnet"):
		return "High-performance model with early extended thinking capabilities"
	case strings.Contains(modelID, "claude-3-5-sonnet"):
		if strings.Contains(modelID, "20241022") {
			return "Most intelligent model with enhanced vision capabilities (Latest)"
		}
		return "Advanced model with superior intelligence and vision capabilities"
	case strings.Contains(modelID, "claude-3-5-haiku"):
		return "Fastest model with vision capabilities and high intelligence"
	case strings.Contains(modelID, "claude-3-opus"):
		return "Most powerful model for complex tasks with exceptional reasoning"
	case strings.Contains(modelID, "claude-3-sonnet"):
		return "Balanced model for general tasks with strong performance"
	case strings.Contains(modelID, "claude-3-haiku"):
		return "Fast and cost-effective model for quick responses"
	default:
		return fmt.Sprintf("Claude %s model with vision support", modelID)
	}
}

// getClaudeMaxTokens возвращает максимальное количество токенов для модели Claude
func (s *AIService) getClaudeMaxTokens(modelID string) int {
	switch {
	// Claude 4 models
	case strings.Contains(modelID, "claude-opus-4"):
		return 200000
	case strings.Contains(modelID, "claude-sonnet-4"):
		return 200000

	// Claude 3.7 models
	case strings.Contains(modelID, "claude-3-7-sonnet"):
		return 200000

	// Claude 3.5 models
	case strings.Contains(modelID, "claude-3-5"):
		return 200000

	// Claude 3 models
	case strings.Contains(modelID, "claude-3"):
		return 200000

	default:
		return 200000 // По умолчанию 200K для всех современных моделей Claude
	}
}
