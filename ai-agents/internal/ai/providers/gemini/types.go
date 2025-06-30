package gemini

// Gemini API request and response types

// GeminiGenerateRequest represents a content generation request to Gemini
type GeminiGenerateRequest struct {
	Contents         []GeminiContent          `json:"contents"`
	Tools            []GeminiTool             `json:"tools,omitempty"`
	SafetySettings   []GeminiSafetySetting    `json:"safetySettings,omitempty"`
	GenerationConfig *GeminiGenerationConfig  `json:"generationConfig,omitempty"`
}

// GeminiGenerateResponse represents a content generation response from Gemini
type GeminiGenerateResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// GeminiContent represents content in a Gemini request/response
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text         string                `json:"text,omitempty"`
	InlineData   *GeminiInlineData     `json:"inlineData,omitempty"`
	FunctionCall *GeminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

// GeminiInlineData represents inline data (e.g., images)
type GeminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // Base64 encoded
}

// GeminiFunctionCall represents a function call
type GeminiFunctionCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// GeminiFunctionResponse represents a function response
type GeminiFunctionResponse struct {
	Name     string                 `json:"name"`
	Response map[string]interface{} `json:"response"`
}

// GeminiCandidate represents a response candidate
type GeminiCandidate struct {
	Content       GeminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason,omitempty"`
	Index         int                  `json:"index,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiSafetyRating represents a safety rating
type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

// GeminiPromptFeedback represents feedback about the prompt
type GeminiPromptFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"`
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// GeminiUsageMetadata represents usage metadata
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiGenerationConfig represents generation configuration
type GeminiGenerationConfig struct {
	StopSequences   []string `json:"stopSequences,omitempty"`
	Temperature     float64  `json:"temperature,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
}

// GeminiTool represents a tool that can be used
type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations,omitempty"`
}

// GeminiFunctionDeclaration represents a function declaration
type GeminiFunctionDeclaration struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// GeminiSafetySetting represents a safety setting
type GeminiSafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

// Gemini embedding types

// GeminiEmbedRequest represents an embedding request to Gemini
type GeminiEmbedRequest struct {
	Model   string        `json:"model"`
	Content GeminiContent `json:"content"`
	TaskType string       `json:"taskType,omitempty"`
	Title    string       `json:"title,omitempty"`
}

// GeminiEmbedResponse represents an embedding response from Gemini
type GeminiEmbedResponse struct {
	Embedding GeminiEmbedding `json:"embedding"`
}

// GeminiEmbedding represents an embedding
type GeminiEmbedding struct {
	Values []float64 `json:"values"`
}

// Gemini batch embedding types

// GeminiBatchEmbedRequest represents a batch embedding request
type GeminiBatchEmbedRequest struct {
	Requests []GeminiEmbedRequest `json:"requests"`
}

// GeminiBatchEmbedResponse represents a batch embedding response
type GeminiBatchEmbedResponse struct {
	Embeddings []GeminiEmbedding `json:"embeddings"`
}

// Gemini model types

// GeminiModelsResponse represents the models list response
type GeminiModelsResponse struct {
	Models []GeminiModel `json:"models"`
}

// GeminiModel represents a Gemini model
type GeminiModel struct {
	Name                 string   `json:"name"`
	BaseModelId          string   `json:"baseModelId,omitempty"`
	Version              string   `json:"version"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description"`
	InputTokenLimit      int      `json:"inputTokenLimit"`
	OutputTokenLimit     int      `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	Temperature          float64  `json:"temperature,omitempty"`
	TopP                 float64  `json:"topP,omitempty"`
	TopK                 int      `json:"topK,omitempty"`
}

// Gemini error types

// GeminiErrorResponse represents an error response from Gemini
type GeminiErrorResponse struct {
	Error GeminiError `json:"error"`
}

// GeminiError represents an error from Gemini
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Gemini streaming types

// GeminiStreamResponse represents a streaming response chunk
type GeminiStreamResponse struct {
	Candidates     []GeminiCandidate     `json:"candidates,omitempty"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
	UsageMetadata  *GeminiUsageMetadata  `json:"usageMetadata,omitempty"`
}

// Gemini tuning types

// GeminiTuningRequest represents a model tuning request
type GeminiTuningRequest struct {
	DisplayName     string                    `json:"displayName"`
	BaseModel       string                    `json:"baseModel"`
	TuningTask      GeminiTuningTask          `json:"tuningTask"`
	Description     string                    `json:"description,omitempty"`
}

// GeminiTuningTask represents a tuning task
type GeminiTuningTask struct {
	StartTime        string                   `json:"startTime,omitempty"`
	CompleteTime     string                   `json:"completeTime,omitempty"`
	Snapshots        []GeminiTuningSnapshot   `json:"snapshots,omitempty"`
	TrainingData     GeminiDataset            `json:"trainingData"`
	Hyperparameters  *GeminiHyperparameters   `json:"hyperparameters,omitempty"`
}

// GeminiTuningSnapshot represents a tuning snapshot
type GeminiTuningSnapshot struct {
	Step         int                    `json:"step"`
	Epoch        int                    `json:"epoch"`
	MeanLoss     float64                `json:"meanLoss"`
	ComputeTime  string                 `json:"computeTime"`
}

// GeminiDataset represents a dataset
type GeminiDataset struct {
	Examples []GeminiTuningExample `json:"examples"`
}

// GeminiTuningExample represents a tuning example
type GeminiTuningExample struct {
	TextInput  string `json:"textInput"`
	Output     string `json:"output"`
}

// GeminiHyperparameters represents hyperparameters
type GeminiHyperparameters struct {
	LearningRate         float64 `json:"learningRate,omitempty"`
	LearningRateMultiplier float64 `json:"learningRateMultiplier,omitempty"`
	EpochCount           int     `json:"epochCount,omitempty"`
	BatchSize            int     `json:"batchSize,omitempty"`
}

// Gemini operation types

// GeminiOperation represents a long-running operation
type GeminiOperation struct {
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Done     bool                   `json:"done"`
	Error    *GeminiError           `json:"error,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

// Gemini file types

// GeminiFile represents a file
type GeminiFile struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName,omitempty"`
	MimeType    string                 `json:"mimeType"`
	SizeBytes   string                 `json:"sizeBytes"`
	CreateTime  string                 `json:"createTime"`
	UpdateTime  string                 `json:"updateTime"`
	ExpirationTime string              `json:"expirationTime,omitempty"`
	Sha256Hash  string                 `json:"sha256Hash"`
	Uri         string                 `json:"uri"`
	State       string                 `json:"state"`
	Error       *GeminiError           `json:"error,omitempty"`
	VideoMetadata *GeminiVideoMetadata `json:"videoMetadata,omitempty"`
}

// GeminiVideoMetadata represents video metadata
type GeminiVideoMetadata struct {
	VideoDuration string `json:"videoDuration"`
}

// GeminiUploadFileRequest represents a file upload request
type GeminiUploadFileRequest struct {
	File GeminiFile `json:"file"`
}

// GeminiUploadFileResponse represents a file upload response
type GeminiUploadFileResponse struct {
	File GeminiFile `json:"file"`
}

// Gemini constants for safety categories
const (
	GeminiHarmCategoryUnspecified         = "HARM_CATEGORY_UNSPECIFIED"
	GeminiHarmCategoryDerogatory          = "HARM_CATEGORY_DEROGATORY"
	GeminiHarmCategoryToxicity            = "HARM_CATEGORY_TOXICITY"
	GeminiHarmCategoryViolence            = "HARM_CATEGORY_VIOLENCE"
	GeminiHarmCategorySexual              = "HARM_CATEGORY_SEXUAL"
	GeminiHarmCategoryMedical             = "HARM_CATEGORY_MEDICAL"
	GeminiHarmCategoryDangerous           = "HARM_CATEGORY_DANGEROUS"
	GeminiHarmCategoryHarassment          = "HARM_CATEGORY_HARASSMENT"
	GeminiHarmCategoryHateSpeech          = "HARM_CATEGORY_HATE_SPEECH"
	GeminiHarmCategorySexuallyExplicit    = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	GeminiHarmCategoryDangerousContent    = "HARM_CATEGORY_DANGEROUS_CONTENT"
)

// Gemini constants for safety thresholds
const (
	GeminiHarmBlockThresholdUnspecified   = "HARM_BLOCK_THRESHOLD_UNSPECIFIED"
	GeminiHarmBlockThresholdBlockLowAndAbove = "BLOCK_LOW_AND_ABOVE"
	GeminiHarmBlockThresholdBlockMediumAndAbove = "BLOCK_MEDIUM_AND_ABOVE"
	GeminiHarmBlockThresholdBlockOnlyHigh = "BLOCK_ONLY_HIGH"
	GeminiHarmBlockThresholdBlockNone     = "BLOCK_NONE"
)

// Gemini constants for finish reasons
const (
	GeminiFinishReasonUnspecified = "FINISH_REASON_UNSPECIFIED"
	GeminiFinishReasonStop        = "STOP"
	GeminiFinishReasonMaxTokens   = "MAX_TOKENS"
	GeminiFinishReasonSafety      = "SAFETY"
	GeminiFinishReasonRecitation  = "RECITATION"
	GeminiFinishReasonOther       = "OTHER"
)

// Gemini constants for task types
const (
	GeminiTaskTypeUnspecified           = "TASK_TYPE_UNSPECIFIED"
	GeminiTaskTypeRetrievalQuery        = "RETRIEVAL_QUERY"
	GeminiTaskTypeRetrievalDocument     = "RETRIEVAL_DOCUMENT"
	GeminiTaskTypeSemanticSimilarity    = "SEMANTIC_SIMILARITY"
	GeminiTaskTypeClassification        = "CLASSIFICATION"
	GeminiTaskTypeClustering            = "CLUSTERING"
)
