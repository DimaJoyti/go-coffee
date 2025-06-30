package openai

// OpenAI API request and response types

// OpenAICompletionRequest represents a completion request to OpenAI
type OpenAICompletionRequest struct {
	Model            string    `json:"model"`
	Prompt           string    `json:"prompt"`
	MaxTokens        int       `json:"max_tokens,omitempty"`
	Temperature      float64   `json:"temperature,omitempty"`
	TopP             float64   `json:"top_p,omitempty"`
	N                int       `json:"n,omitempty"`
	Stream           bool      `json:"stream,omitempty"`
	Logprobs         int       `json:"logprobs,omitempty"`
	Echo             bool      `json:"echo,omitempty"`
	Stop             []string  `json:"stop,omitempty"`
	PresencePenalty  float64   `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64   `json:"frequency_penalty,omitempty"`
	BestOf           int       `json:"best_of,omitempty"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	User             string    `json:"user,omitempty"`
}

// OpenAICompletionResponse represents a completion response from OpenAI
type OpenAICompletionResponse struct {
	ID      string                    `json:"id"`
	Object  string                    `json:"object"`
	Created int64                     `json:"created"`
	Model   string                    `json:"model"`
	Choices []OpenAICompletionChoice  `json:"choices"`
	Usage   OpenAIUsage               `json:"usage"`
}

// OpenAICompletionChoice represents a completion choice
type OpenAICompletionChoice struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	Logprobs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

// OpenAIChatRequest represents a chat completion request to OpenAI
type OpenAIChatRequest struct {
	Model            string              `json:"model"`
	Messages         []OpenAIChatMessage `json:"messages"`
	MaxTokens        int                 `json:"max_tokens,omitempty"`
	Temperature      float64             `json:"temperature,omitempty"`
	TopP             float64             `json:"top_p,omitempty"`
	N                int                 `json:"n,omitempty"`
	Stream           bool                `json:"stream,omitempty"`
	Stop             []string            `json:"stop,omitempty"`
	PresencePenalty  float64             `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64             `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64  `json:"logit_bias,omitempty"`
	User             string              `json:"user,omitempty"`
	Functions        []OpenAIFunction    `json:"functions,omitempty"`
	FunctionCall     interface{}         `json:"function_call,omitempty"`
}

// OpenAIChatResponse represents a chat completion response from OpenAI
type OpenAIChatResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []OpenAIChatChoice  `json:"choices"`
	Usage   OpenAIUsage         `json:"usage"`
}

// OpenAIChatChoice represents a chat completion choice
type OpenAIChatChoice struct {
	Index        int                 `json:"index"`
	Message      OpenAIChatMessage   `json:"message"`
	FinishReason string              `json:"finish_reason"`
	Delta        *OpenAIChatMessage  `json:"delta,omitempty"`
}

// OpenAIChatMessage represents a chat message
type OpenAIChatMessage struct {
	Role         string                 `json:"role"`
	Content      string                 `json:"content"`
	Name         string                 `json:"name,omitempty"`
	FunctionCall *OpenAIFunctionCall    `json:"function_call,omitempty"`
}

// OpenAIFunction represents a function definition
type OpenAIFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// OpenAIFunctionCall represents a function call
type OpenAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OpenAIEmbeddingRequest represents an embedding request to OpenAI
type OpenAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	User  string   `json:"user,omitempty"`
}

// OpenAIEmbeddingResponse represents an embedding response from OpenAI
type OpenAIEmbeddingResponse struct {
	Object string                `json:"object"`
	Data   []OpenAIEmbeddingData `json:"data"`
	Model  string                `json:"model"`
	Usage  OpenAIUsage           `json:"usage"`
}

// OpenAIEmbeddingData represents embedding data
type OpenAIEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// OpenAIUsage represents token usage information
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIErrorResponse represents an error response from OpenAI
type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

// OpenAIError represents an error from OpenAI
type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    interface{} `json:"code"`
}

// OpenAIModelsResponse represents the models list response
type OpenAIModelsResponse struct {
	Object string          `json:"object"`
	Data   []OpenAIModel   `json:"data"`
}

// OpenAIModel represents a model from OpenAI
type OpenAIModel struct {
	ID         string                 `json:"id"`
	Object     string                 `json:"object"`
	Created    int64                  `json:"created"`
	OwnedBy    string                 `json:"owned_by"`
	Permission []OpenAIModelPermission `json:"permission"`
	Root       string                 `json:"root"`
	Parent     interface{}            `json:"parent"`
}

// OpenAIModelPermission represents model permissions
type OpenAIModelPermission struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int64       `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
}

// OpenAI streaming response types

// OpenAIStreamResponse represents a streaming response chunk
type OpenAIStreamResponse struct {
	ID      string                   `json:"id"`
	Object  string                   `json:"object"`
	Created int64                    `json:"created"`
	Model   string                   `json:"model"`
	Choices []OpenAIStreamChoice     `json:"choices"`
}

// OpenAIStreamChoice represents a streaming choice
type OpenAIStreamChoice struct {
	Index        int                `json:"index"`
	Delta        OpenAIChatMessage  `json:"delta"`
	FinishReason *string            `json:"finish_reason"`
}

// OpenAI fine-tuning types

// OpenAIFineTuneRequest represents a fine-tuning request
type OpenAIFineTuneRequest struct {
	TrainingFile                 string   `json:"training_file"`
	ValidationFile               string   `json:"validation_file,omitempty"`
	Model                        string   `json:"model,omitempty"`
	NEpochs                      int      `json:"n_epochs,omitempty"`
	BatchSize                    int      `json:"batch_size,omitempty"`
	LearningRateMultiplier       float64  `json:"learning_rate_multiplier,omitempty"`
	PromptLossWeight             float64  `json:"prompt_loss_weight,omitempty"`
	ComputeClassificationMetrics bool     `json:"compute_classification_metrics,omitempty"`
	ClassificationNClasses       int      `json:"classification_n_classes,omitempty"`
	ClassificationPositiveClass  string   `json:"classification_positive_class,omitempty"`
	ClassificationBetas          []float64 `json:"classification_betas,omitempty"`
	Suffix                       string   `json:"suffix,omitempty"`
}

// OpenAIFineTuneResponse represents a fine-tuning response
type OpenAIFineTuneResponse struct {
	ID             string                    `json:"id"`
	Object         string                    `json:"object"`
	Model          string                    `json:"model"`
	CreatedAt      int64                     `json:"created_at"`
	Events         []OpenAIFineTuneEvent     `json:"events"`
	FineTunedModel string                    `json:"fine_tuned_model"`
	Hyperparams    OpenAIFineTuneHyperparams `json:"hyperparams"`
	OrganizationID string                    `json:"organization_id"`
	ResultFiles    []OpenAIFile              `json:"result_files"`
	Status         string                    `json:"status"`
	ValidationFiles []OpenAIFile             `json:"validation_files"`
	TrainingFiles  []OpenAIFile              `json:"training_files"`
	UpdatedAt      int64                     `json:"updated_at"`
}

// OpenAIFineTuneEvent represents a fine-tuning event
type OpenAIFineTuneEvent struct {
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// OpenAIFineTuneHyperparams represents fine-tuning hyperparameters
type OpenAIFineTuneHyperparams struct {
	BatchSize              int     `json:"batch_size"`
	LearningRateMultiplier float64 `json:"learning_rate_multiplier"`
	NEpochs                int     `json:"n_epochs"`
	PromptLossWeight       float64 `json:"prompt_loss_weight"`
}

// OpenAIFile represents a file object
type OpenAIFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

// OpenAI moderation types

// OpenAIModerationRequest represents a moderation request
type OpenAIModerationRequest struct {
	Input string `json:"input"`
	Model string `json:"model,omitempty"`
}

// OpenAIModerationResponse represents a moderation response
type OpenAIModerationResponse struct {
	ID      string                     `json:"id"`
	Model   string                     `json:"model"`
	Results []OpenAIModerationResult   `json:"results"`
}

// OpenAIModerationResult represents moderation results
type OpenAIModerationResult struct {
	Categories     OpenAIModerationCategories     `json:"categories"`
	CategoryScores OpenAIModerationCategoryScores `json:"category_scores"`
	Flagged        bool                           `json:"flagged"`
}

// OpenAIModerationCategories represents moderation categories
type OpenAIModerationCategories struct {
	Hate            bool `json:"hate"`
	HateThreatening bool `json:"hate/threatening"`
	SelfHarm        bool `json:"self-harm"`
	Sexual          bool `json:"sexual"`
	SexualMinors    bool `json:"sexual/minors"`
	Violence        bool `json:"violence"`
	ViolenceGraphic bool `json:"violence/graphic"`
}

// OpenAIModerationCategoryScores represents moderation category scores
type OpenAIModerationCategoryScores struct {
	Hate            float64 `json:"hate"`
	HateThreatening float64 `json:"hate/threatening"`
	SelfHarm        float64 `json:"self-harm"`
	Sexual          float64 `json:"sexual"`
	SexualMinors    float64 `json:"sexual/minors"`
	Violence        float64 `json:"violence"`
	ViolenceGraphic float64 `json:"violence/graphic"`
}
