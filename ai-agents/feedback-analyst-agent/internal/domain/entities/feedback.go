package entities

import (
	"time"

	"github.com/google/uuid"
)

// Feedback represents customer feedback in the domain
type Feedback struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	CustomerID        *uuid.UUID             `json:"customer_id,omitempty" redis:"customer_id"`
	CustomerName      string                 `json:"customer_name" redis:"customer_name"`
	CustomerEmail     string                 `json:"customer_email" redis:"customer_email"`
	Source            FeedbackSource         `json:"source" redis:"source"`
	Type              FeedbackType           `json:"type" redis:"type"`
	Category          FeedbackCategory       `json:"category" redis:"category"`
	Priority          FeedbackPriority       `json:"priority" redis:"priority"`
	Status            FeedbackStatus         `json:"status" redis:"status"`
	Subject           string                 `json:"subject" redis:"subject"`
	Content           string                 `json:"content" redis:"content"`
	Rating            *int                   `json:"rating,omitempty" redis:"rating"`
	SentimentAnalysis *SentimentAnalysis     `json:"sentiment_analysis,omitempty"`
	Topics            []string               `json:"topics" redis:"topics"`
	Keywords          []string               `json:"keywords" redis:"keywords"`
	Tags              []string               `json:"tags" redis:"tags"`
	Location          *CustomerLocation      `json:"location,omitempty"`
	Device            *DeviceInfo            `json:"device,omitempty"`
	OrderID           *uuid.UUID             `json:"order_id,omitempty" redis:"order_id"`
	ProductIDs        []uuid.UUID            `json:"product_ids" redis:"product_ids"`
	StoreID           *uuid.UUID             `json:"store_id,omitempty" redis:"store_id"`
	AssignedTo        *uuid.UUID             `json:"assigned_to,omitempty" redis:"assigned_to"`
	ResponseSuggestions []*ResponseSuggestion `json:"response_suggestions,omitempty"`
	Responses         []*FeedbackResponse    `json:"responses,omitempty"`
	Attachments       []*FeedbackAttachment  `json:"attachments,omitempty"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	IsAnonymous       bool                   `json:"is_anonymous" redis:"is_anonymous"`
	IsVerified        bool                   `json:"is_verified" redis:"is_verified"`
	IsPublic          bool                   `json:"is_public" redis:"is_public"`
	IsResolved        bool                   `json:"is_resolved" redis:"is_resolved"`
	ResolutionTime    *time.Duration         `json:"resolution_time,omitempty" redis:"resolution_time"`
	SatisfactionScore *float64               `json:"satisfaction_score,omitempty" redis:"satisfaction_score"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	ResolvedAt        *time.Time             `json:"resolved_at,omitempty" redis:"resolved_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// FeedbackSource defines the source of feedback
type FeedbackSource string

const (
	SourceWebsite     FeedbackSource = "website"
	SourceMobile      FeedbackSource = "mobile"
	SourceEmail       FeedbackSource = "email"
	SourcePhone       FeedbackSource = "phone"
	SourceSocialMedia FeedbackSource = "social_media"
	SourceInStore     FeedbackSource = "in_store"
	SourceSurvey      FeedbackSource = "survey"
	SourceReview      FeedbackSource = "review"
	SourceChat        FeedbackSource = "chat"
	SourceAPI         FeedbackSource = "api"
)

// FeedbackType defines the type of feedback
type FeedbackType string

const (
	TypeComplaint    FeedbackType = "complaint"
	TypeCompliment   FeedbackType = "compliment"
	TypeSuggestion   FeedbackType = "suggestion"
	TypeQuestion     FeedbackType = "question"
	TypeBugReport    FeedbackType = "bug_report"
	TypeFeatureRequest FeedbackType = "feature_request"
	TypeReview       FeedbackType = "review"
	TypeTestimonial  FeedbackType = "testimonial"
	TypeGeneral      FeedbackType = "general"
)

// FeedbackCategory defines the category of feedback
type FeedbackCategory string

const (
	CategoryProduct     FeedbackCategory = "product"
	CategoryService     FeedbackCategory = "service"
	CategoryDelivery    FeedbackCategory = "delivery"
	CategoryPricing     FeedbackCategory = "pricing"
	CategoryWebsite     FeedbackCategory = "website"
	CategoryApp         FeedbackCategory = "app"
	CategoryStaff       FeedbackCategory = "staff"
	CategoryCleanliness FeedbackCategory = "cleanliness"
	CategoryAmbiance    FeedbackCategory = "ambiance"
	CategoryWaitTime    FeedbackCategory = "wait_time"
	CategoryOther       FeedbackCategory = "other"
)

// FeedbackPriority defines the priority of feedback
type FeedbackPriority string

const (
	PriorityLow      FeedbackPriority = "low"
	PriorityMedium   FeedbackPriority = "medium"
	PriorityHigh     FeedbackPriority = "high"
	PriorityCritical FeedbackPriority = "critical"
	PriorityUrgent   FeedbackPriority = "urgent"
)

// FeedbackStatus defines the status of feedback
type FeedbackStatus string

const (
	StatusNew        FeedbackStatus = "new"
	StatusAssigned   FeedbackStatus = "assigned"
	StatusInProgress FeedbackStatus = "in_progress"
	StatusPending    FeedbackStatus = "pending"
	StatusResolved   FeedbackStatus = "resolved"
	StatusClosed     FeedbackStatus = "closed"
	StatusEscalated  FeedbackStatus = "escalated"
	StatusArchived   FeedbackStatus = "archived"
)

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Sentiment    Sentiment              `json:"sentiment" redis:"sentiment"`
	Score        float64                `json:"score" redis:"score"`
	Confidence   float64                `json:"confidence" redis:"confidence"`
	Emotions     map[string]float64     `json:"emotions" redis:"emotions"`
	Aspects      []*AspectSentiment     `json:"aspects,omitempty"`
	Language     string                 `json:"language" redis:"language"`
	Subjectivity float64                `json:"subjectivity" redis:"subjectivity"`
	Polarity     float64                `json:"polarity" redis:"polarity"`
	AnalyzedAt   time.Time              `json:"analyzed_at" redis:"analyzed_at"`
	Model        string                 `json:"model" redis:"model"`
}

// Sentiment defines sentiment types
type Sentiment string

const (
	SentimentPositive Sentiment = "positive"
	SentimentNegative Sentiment = "negative"
	SentimentNeutral  Sentiment = "neutral"
	SentimentMixed    Sentiment = "mixed"
)

// AspectSentiment represents sentiment for specific aspects
type AspectSentiment struct {
	Aspect     string    `json:"aspect" redis:"aspect"`
	Sentiment  Sentiment `json:"sentiment" redis:"sentiment"`
	Score      float64   `json:"score" redis:"score"`
	Confidence float64   `json:"confidence" redis:"confidence"`
	Mentions   []string  `json:"mentions" redis:"mentions"`
}

// CustomerLocation represents customer location information
type CustomerLocation struct {
	Country   string  `json:"country" redis:"country"`
	State     string  `json:"state" redis:"state"`
	City      string  `json:"city" redis:"city"`
	Latitude  float64 `json:"latitude" redis:"latitude"`
	Longitude float64 `json:"longitude" redis:"longitude"`
	IPAddress string  `json:"ip_address" redis:"ip_address"`
}

// DeviceInfo represents device information
type DeviceInfo struct {
	Type         string `json:"type" redis:"type"`
	OS           string `json:"os" redis:"os"`
	Browser      string `json:"browser" redis:"browser"`
	Version      string `json:"version" redis:"version"`
	UserAgent    string `json:"user_agent" redis:"user_agent"`
	ScreenSize   string `json:"screen_size" redis:"screen_size"`
	IsMobile     bool   `json:"is_mobile" redis:"is_mobile"`
}

// ResponseSuggestion represents AI-generated response suggestions
type ResponseSuggestion struct {
	ID          uuid.UUID              `json:"id" redis:"id"`
	FeedbackID  uuid.UUID              `json:"feedback_id" redis:"feedback_id"`
	Type        ResponseType           `json:"type" redis:"type"`
	Content     string                 `json:"content" redis:"content"`
	Tone        ResponseTone           `json:"tone" redis:"tone"`
	Confidence  float64                `json:"confidence" redis:"confidence"`
	Reasoning   string                 `json:"reasoning" redis:"reasoning"`
	Templates   []string               `json:"templates" redis:"templates"`
	Actions     []*SuggestedAction     `json:"actions,omitempty"`
	IsSelected  bool                   `json:"is_selected" redis:"is_selected"`
	IsUsed      bool                   `json:"is_used" redis:"is_used"`
	CreatedAt   time.Time              `json:"created_at" redis:"created_at"`
	GeneratedBy string                 `json:"generated_by" redis:"generated_by"`
}

// ResponseType defines types of responses
type ResponseType string

const (
	ResponseTypeApology      ResponseType = "apology"
	ResponseTypeThankYou     ResponseType = "thank_you"
	ResponseTypeExplanation  ResponseType = "explanation"
	ResponseTypeSolution     ResponseType = "solution"
	ResponseTypeEscalation   ResponseType = "escalation"
	ResponseTypeFollowUp     ResponseType = "follow_up"
	ResponseTypeInformation  ResponseType = "information"
	ResponseTypeAcknowledgment ResponseType = "acknowledgment"
)

// ResponseTone defines the tone of responses
type ResponseTone string

const (
	ToneFormal     ResponseTone = "formal"
	ToneFriendly   ResponseTone = "friendly"
	ToneEmpathetic ResponseTone = "empathetic"
	ToneProfessional ResponseTone = "professional"
	ToneApologetic ResponseTone = "apologetic"
	ToneGrateful   ResponseTone = "grateful"
	ToneHelpful    ResponseTone = "helpful"
)

// SuggestedAction represents suggested actions for feedback resolution
type SuggestedAction struct {
	Type        ActionType `json:"type" redis:"type"`
	Description string     `json:"description" redis:"description"`
	Priority    int        `json:"priority" redis:"priority"`
	Assignee    *uuid.UUID `json:"assignee,omitempty" redis:"assignee"`
	DueDate     *time.Time `json:"due_date,omitempty" redis:"due_date"`
	Metadata    map[string]interface{} `json:"metadata" redis:"metadata"`
}

// ActionType defines types of suggested actions
type ActionType string

const (
	ActionRefund        ActionType = "refund"
	ActionReplacement   ActionType = "replacement"
	ActionDiscount      ActionType = "discount"
	ActionFollowUp      ActionType = "follow_up"
	ActionEscalate      ActionType = "escalate"
	ActionTraining      ActionType = "training"
	ActionProcessChange ActionType = "process_change"
	ActionInvestigate   ActionType = "investigate"
	ActionNotify        ActionType = "notify"
)

// FeedbackResponse represents responses to feedback
type FeedbackResponse struct {
	ID         uuid.UUID     `json:"id" redis:"id"`
	FeedbackID uuid.UUID     `json:"feedback_id" redis:"feedback_id"`
	Content    string        `json:"content" redis:"content"`
	Type       ResponseType  `json:"type" redis:"type"`
	Tone       ResponseTone  `json:"tone" redis:"tone"`
	Channel    string        `json:"channel" redis:"channel"`
	IsPublic   bool          `json:"is_public" redis:"is_public"`
	IsAI       bool          `json:"is_ai" redis:"is_ai"`
	CreatedAt  time.Time     `json:"created_at" redis:"created_at"`
	CreatedBy  uuid.UUID     `json:"created_by" redis:"created_by"`
	SentAt     *time.Time    `json:"sent_at,omitempty" redis:"sent_at"`
	ReadAt     *time.Time    `json:"read_at,omitempty" redis:"read_at"`
}

// FeedbackAttachment represents file attachments
type FeedbackAttachment struct {
	ID         uuid.UUID `json:"id" redis:"id"`
	FeedbackID uuid.UUID `json:"feedback_id" redis:"feedback_id"`
	FileName   string    `json:"file_name" redis:"file_name"`
	FileType   string    `json:"file_type" redis:"file_type"`
	FileSize   int64     `json:"file_size" redis:"file_size"`
	URL        string    `json:"url" redis:"url"`
	IsImage    bool      `json:"is_image" redis:"is_image"`
	IsVideo    bool      `json:"is_video" redis:"is_video"`
	IsAudio    bool      `json:"is_audio" redis:"is_audio"`
	CreatedAt  time.Time `json:"created_at" redis:"created_at"`
}

// FeedbackTrend represents feedback trends and analytics
type FeedbackTrend struct {
	Period           string                           `json:"period"`
	TotalFeedback    int                              `json:"total_feedback"`
	SentimentBreakdown map[Sentiment]int              `json:"sentiment_breakdown"`
	CategoryBreakdown  map[FeedbackCategory]int       `json:"category_breakdown"`
	SourceBreakdown    map[FeedbackSource]int         `json:"source_breakdown"`
	TypeBreakdown      map[FeedbackType]int           `json:"type_breakdown"`
	AverageRating      float64                        `json:"average_rating"`
	AverageResolutionTime time.Duration               `json:"average_resolution_time"`
	SatisfactionScore  float64                        `json:"satisfaction_score"`
	TopKeywords        []KeywordFrequency             `json:"top_keywords"`
	TopTopics          []TopicFrequency               `json:"top_topics"`
	TrendDirection     TrendDirection                 `json:"trend_direction"`
	GrowthRate         float64                        `json:"growth_rate"`
	GeneratedAt        time.Time                      `json:"generated_at"`
}

// KeywordFrequency represents keyword frequency data
type KeywordFrequency struct {
	Keyword   string  `json:"keyword"`
	Count     int     `json:"count"`
	Sentiment Sentiment `json:"sentiment"`
	Change    float64 `json:"change"`
}

// TopicFrequency represents topic frequency data
type TopicFrequency struct {
	Topic     string    `json:"topic"`
	Count     int       `json:"count"`
	Sentiment Sentiment `json:"sentiment"`
	Change    float64   `json:"change"`
}

// TrendDirection defines trend directions
type TrendDirection string

const (
	TrendUp    TrendDirection = "up"
	TrendDown  TrendDirection = "down"
	TrendFlat  TrendDirection = "flat"
)

// NewFeedback creates a new feedback instance
func NewFeedback(customerName, customerEmail, subject, content string, source FeedbackSource, createdBy uuid.UUID) *Feedback {
	now := time.Now()
	return &Feedback{
		ID:            uuid.New(),
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Source:        source,
		Type:          TypeGeneral,
		Category:      CategoryOther,
		Priority:      PriorityMedium,
		Status:        StatusNew,
		Subject:       subject,
		Content:       content,
		Topics:        []string{},
		Keywords:      []string{},
		Tags:          []string{},
		ProductIDs:    []uuid.UUID{},
		Metadata:      make(map[string]interface{}),
		IsAnonymous:   false,
		IsVerified:    false,
		IsPublic:      false,
		IsResolved:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
		CreatedBy:     createdBy,
		UpdatedBy:     createdBy,
		Version:       1,
	}
}

// UpdateStatus updates the feedback status
func (f *Feedback) UpdateStatus(newStatus FeedbackStatus, updatedBy uuid.UUID) {
	f.Status = newStatus
	f.UpdatedBy = updatedBy
	f.UpdatedAt = time.Now()
	f.Version++

	// Handle status-specific logic
	if newStatus == StatusResolved || newStatus == StatusClosed {
		if !f.IsResolved {
			now := time.Now()
			f.ResolvedAt = &now
			f.IsResolved = true
			
			// Calculate resolution time
			resolutionTime := now.Sub(f.CreatedAt)
			f.ResolutionTime = &resolutionTime
		}
	}
}

// AddResponse adds a response to the feedback
func (f *Feedback) AddResponse(response *FeedbackResponse) {
	response.FeedbackID = f.ID
	f.Responses = append(f.Responses, response)
	f.UpdatedAt = time.Now()
	f.Version++
}

// AddSuggestion adds a response suggestion
func (f *Feedback) AddSuggestion(suggestion *ResponseSuggestion) {
	suggestion.FeedbackID = f.ID
	f.ResponseSuggestions = append(f.ResponseSuggestions, suggestion)
	f.UpdatedAt = time.Now()
	f.Version++
}

// SetSentimentAnalysis sets the sentiment analysis results
func (f *Feedback) SetSentimentAnalysis(analysis *SentimentAnalysis) {
	f.SentimentAnalysis = analysis
	f.UpdatedAt = time.Now()
	f.Version++
}

// AssignTo assigns the feedback to a user
func (f *Feedback) AssignTo(userID uuid.UUID, updatedBy uuid.UUID) {
	f.AssignedTo = &userID
	f.Status = StatusAssigned
	f.UpdatedBy = updatedBy
	f.UpdatedAt = time.Now()
	f.Version++
}

// IsHighPriority checks if feedback is high priority
func (f *Feedback) IsHighPriority() bool {
	return f.Priority == PriorityHigh || f.Priority == PriorityCritical || f.Priority == PriorityUrgent
}

// IsNegative checks if feedback has negative sentiment
func (f *Feedback) IsNegative() bool {
	return f.SentimentAnalysis != nil && f.SentimentAnalysis.Sentiment == SentimentNegative
}

// RequiresImmediateAttention checks if feedback needs immediate attention
func (f *Feedback) RequiresImmediateAttention() bool {
	return f.IsHighPriority() && f.IsNegative() && f.Status == StatusNew
}

// GetResolutionTime returns the resolution time if resolved
func (f *Feedback) GetResolutionTime() *time.Duration {
	if f.IsResolved && f.ResolvedAt != nil {
		duration := f.ResolvedAt.Sub(f.CreatedAt)
		return &duration
	}
	return nil
}

// CalculateSatisfactionScore calculates satisfaction score based on various factors
func (f *Feedback) CalculateSatisfactionScore() float64 {
	score := 0.0
	
	// Base score from rating
	if f.Rating != nil {
		score = float64(*f.Rating) * 20 // Convert 1-5 to 0-100
	}
	
	// Adjust based on sentiment
	if f.SentimentAnalysis != nil {
		switch f.SentimentAnalysis.Sentiment {
		case SentimentPositive:
			score += 20
		case SentimentNegative:
			score -= 20
		case SentimentNeutral:
			// No adjustment
		}
	}
	
	// Adjust based on type
	switch f.Type {
	case TypeCompliment, TypeTestimonial:
		score += 10
	case TypeComplaint:
		score -= 10
	}
	
	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	
	f.SatisfactionScore = &score
	return score
}
