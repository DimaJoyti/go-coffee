package entities

import (
	"time"

	"github.com/google/uuid"
)

// Brand represents a comprehensive brand entity for social media management
type Brand struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Name              string                 `json:"name" redis:"name"`
	DisplayName       string                 `json:"display_name" redis:"display_name"`
	Description       string                 `json:"description" redis:"description"`
	Industry          string                 `json:"industry" redis:"industry"`
	Website           string                 `json:"website" redis:"website"`
	Logo              string                 `json:"logo" redis:"logo"`
	Colors            *BrandColors           `json:"colors,omitempty"`
	Typography        *BrandTypography       `json:"typography,omitempty"`
	Voice             *BrandVoice            `json:"voice,omitempty"`
	Guidelines        *BrandGuidelines       `json:"guidelines,omitempty"`
	SocialProfiles    []*SocialProfile       `json:"social_profiles,omitempty"`
	ContactInfo       *ContactInfo           `json:"contact_info,omitempty"`
	ComplianceRules   []*ComplianceRule      `json:"compliance_rules,omitempty"`
	ContentTemplates  []*ContentTemplate     `json:"content_templates,omitempty"`
	HashtagSets       []*HashtagSet          `json:"hashtag_sets,omitempty"`
	Keywords          []string               `json:"keywords" redis:"keywords"`
	Tags              []string               `json:"tags" redis:"tags"`
	Languages         []string               `json:"languages" redis:"languages"`
	TimeZone          string                 `json:"time_zone" redis:"time_zone"`
	Status            BrandStatus            `json:"status" redis:"status"`
	CustomFields      map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	Metadata          map[string]interface{} `json:"metadata" redis:"metadata"`
	ExternalIDs       map[string]string      `json:"external_ids" redis:"external_ids"`
	Active            bool                   `json:"is_active" redis:"is_active"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy         uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// BrandStatus defines the status of a brand
type BrandStatus string

const (
	BrandStatusActive   BrandStatus = "active"
	BrandStatusInactive BrandStatus = "inactive"
	BrandStatusSuspended BrandStatus = "suspended"
	BrandStatusArchived BrandStatus = "archived"
)

// BrandColors represents brand color palette
type BrandColors struct {
	Primary     string   `json:"primary" redis:"primary"`
	Secondary   string   `json:"secondary" redis:"secondary"`
	Accent      string   `json:"accent" redis:"accent"`
	Background  string   `json:"background" redis:"background"`
	Text        string   `json:"text" redis:"text"`
	Success     string   `json:"success" redis:"success"`
	Warning     string   `json:"warning" redis:"warning"`
	Error       string   `json:"error" redis:"error"`
	Palette     []string `json:"palette" redis:"palette"`
}

// BrandTypography represents brand typography settings
type BrandTypography struct {
	PrimaryFont   string `json:"primary_font" redis:"primary_font"`
	SecondaryFont string `json:"secondary_font" redis:"secondary_font"`
	HeadingFont   string `json:"heading_font" redis:"heading_font"`
	BodyFont      string `json:"body_font" redis:"body_font"`
	FontSizes     map[string]string `json:"font_sizes" redis:"font_sizes"`
	LineHeights   map[string]string `json:"line_heights" redis:"line_heights"`
}

// BrandVoice represents brand voice and tone guidelines
type BrandVoice struct {
	Personality   []string               `json:"personality" redis:"personality"`
	Tone          ContentTone            `json:"tone" redis:"tone"`
	Style         VoiceStyle             `json:"style" redis:"style"`
	DosList       []string               `json:"dos_list" redis:"dos_list"`
	DontsList     []string               `json:"donts_list" redis:"donts_list"`
	KeyPhrases    []string               `json:"key_phrases" redis:"key_phrases"`
	AvoidPhrases  []string               `json:"avoid_phrases" redis:"avoid_phrases"`
	Examples      []*VoiceExample        `json:"examples,omitempty"`
	Guidelines    string                 `json:"guidelines" redis:"guidelines"`
}

// VoiceStyle defines the style of brand voice
type VoiceStyle string

const (
	VoiceStyleFormal     VoiceStyle = "formal"
	VoiceStyleCasual     VoiceStyle = "casual"
	VoiceStylePlayful    VoiceStyle = "playful"
	VoiceStyleProfessional VoiceStyle = "professional"
	VoiceStyleFriendly   VoiceStyle = "friendly"
	VoiceStyleAuthoritative VoiceStyle = "authoritative"
	VoiceStyleConversational VoiceStyle = "conversational"
)

// VoiceExample represents an example of brand voice usage
type VoiceExample struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	Title       string    `json:"title" redis:"title"`
	Context     string    `json:"context" redis:"context"`
	GoodExample string    `json:"good_example" redis:"good_example"`
	BadExample  string    `json:"bad_example" redis:"bad_example"`
	Explanation string    `json:"explanation" redis:"explanation"`
	Category    string    `json:"category" redis:"category"`
}

// BrandGuidelines represents comprehensive brand guidelines
type BrandGuidelines struct {
	LogoUsage        *LogoGuidelines        `json:"logo_usage,omitempty"`
	ColorUsage       *ColorGuidelines       `json:"color_usage,omitempty"`
	TypographyUsage  *TypographyGuidelines  `json:"typography_usage,omitempty"`
	ImageryStyle     *ImageryGuidelines     `json:"imagery_style,omitempty"`
	ContentGuidelines *ContentGuidelines    `json:"content_guidelines,omitempty"`
	SocialMediaRules *SocialMediaGuidelines `json:"social_media_rules,omitempty"`
	ComplianceRules  *ComplianceGuidelines  `json:"compliance_rules,omitempty"`
	LastUpdated      time.Time              `json:"last_updated" redis:"last_updated"`
	UpdatedBy        uuid.UUID              `json:"updated_by" redis:"updated_by"`
}

// LogoGuidelines represents logo usage guidelines
type LogoGuidelines struct {
	MinSize         string   `json:"min_size" redis:"min_size"`
	ClearSpace      string   `json:"clear_space" redis:"clear_space"`
	AllowedFormats  []string `json:"allowed_formats" redis:"allowed_formats"`
	ProhibitedUses  []string `json:"prohibited_uses" redis:"prohibited_uses"`
	BackgroundRules []string `json:"background_rules" redis:"background_rules"`
	Variations      []string `json:"variations" redis:"variations"`
}

// ColorGuidelines represents color usage guidelines
type ColorGuidelines struct {
	PrimaryUsage    string            `json:"primary_usage" redis:"primary_usage"`
	SecondaryUsage  string            `json:"secondary_usage" redis:"secondary_usage"`
	AccentUsage     string            `json:"accent_usage" redis:"accent_usage"`
	Combinations    map[string]string `json:"combinations" redis:"combinations"`
	Accessibility   []string          `json:"accessibility" redis:"accessibility"`
	ProhibitedUses  []string          `json:"prohibited_uses" redis:"prohibited_uses"`
}

// TypographyGuidelines represents typography usage guidelines
type TypographyGuidelines struct {
	HierarchyRules  []string          `json:"hierarchy_rules" redis:"hierarchy_rules"`
	SizeGuidelines  map[string]string `json:"size_guidelines" redis:"size_guidelines"`
	SpacingRules    []string          `json:"spacing_rules" redis:"spacing_rules"`
	AlignmentRules  []string          `json:"alignment_rules" redis:"alignment_rules"`
	ProhibitedFonts []string          `json:"prohibited_fonts" redis:"prohibited_fonts"`
}

// ImageryGuidelines represents imagery style guidelines
type ImageryGuidelines struct {
	Style           string   `json:"style" redis:"style"`
	ColorTreatment  string   `json:"color_treatment" redis:"color_treatment"`
	Composition     []string `json:"composition" redis:"composition"`
	SubjectMatter   []string `json:"subject_matter" redis:"subject_matter"`
	Filters         []string `json:"filters" redis:"filters"`
	ProhibitedTypes []string `json:"prohibited_types" redis:"prohibited_types"`
	QualityStandards []string `json:"quality_standards" redis:"quality_standards"`
}

// ContentGuidelines represents content creation guidelines
type ContentGuidelines struct {
	ToneGuidelines    []string `json:"tone_guidelines" redis:"tone_guidelines"`
	StyleGuidelines   []string `json:"style_guidelines" redis:"style_guidelines"`
	LengthGuidelines  map[string]string `json:"length_guidelines" redis:"length_guidelines"`
	HashtagRules      []string `json:"hashtag_rules" redis:"hashtag_rules"`
	MentionRules      []string `json:"mention_rules" redis:"mention_rules"`
	LinkRules         []string `json:"link_rules" redis:"link_rules"`
	ProhibitedContent []string `json:"prohibited_content" redis:"prohibited_content"`
}

// SocialMediaGuidelines represents social media specific guidelines
type SocialMediaGuidelines struct {
	PostingFrequency  map[PlatformType]string `json:"posting_frequency" redis:"posting_frequency"`
	OptimalTimes      map[PlatformType][]string `json:"optimal_times" redis:"optimal_times"`
	PlatformVoice     map[PlatformType]string `json:"platform_voice" redis:"platform_voice"`
	HashtagLimits     map[PlatformType]int `json:"hashtag_limits" redis:"hashtag_limits"`
	ContentTypes      map[PlatformType][]string `json:"content_types" redis:"content_types"`
	EngagementRules   []string `json:"engagement_rules" redis:"engagement_rules"`
	CrisisProtocol    []string `json:"crisis_protocol" redis:"crisis_protocol"`
}

// ComplianceGuidelines represents compliance guidelines
type ComplianceGuidelines struct {
	LegalRequirements []string `json:"legal_requirements" redis:"legal_requirements"`
	DisclosureRules   []string `json:"disclosure_rules" redis:"disclosure_rules"`
	PrivacyRules      []string `json:"privacy_rules" redis:"privacy_rules"`
	AccessibilityRules []string `json:"accessibility_rules" redis:"accessibility_rules"`
	IndustryStandards []string `json:"industry_standards" redis:"industry_standards"`
	ReviewProcess     []string `json:"review_process" redis:"review_process"`
}

// SocialProfile represents a social media profile for the brand
type SocialProfile struct {
	ID           uuid.UUID    `json:"id" redis:"id"`
	BrandID      uuid.UUID    `json:"brand_id" redis:"brand_id"`
	Platform     PlatformType `json:"platform" redis:"platform"`
	Username     string       `json:"username" redis:"username"`
	DisplayName  string       `json:"display_name" redis:"display_name"`
	Bio          string       `json:"bio" redis:"bio"`
	URL          string       `json:"url" redis:"url"`
	ProfileImage string       `json:"profile_image" redis:"profile_image"`
	CoverImage   string       `json:"cover_image" redis:"cover_image"`
	Verified     bool         `json:"verified" redis:"verified"`
	Followers    int64        `json:"followers" redis:"followers"`
	Following    int64        `json:"following" redis:"following"`
	Posts        int64        `json:"posts" redis:"posts"`
	AccessToken  string       `json:"access_token" redis:"access_token"`
	RefreshToken string       `json:"refresh_token" redis:"refresh_token"`
	TokenExpiry  *time.Time   `json:"token_expiry,omitempty" redis:"token_expiry"`
	IsActive     bool         `json:"is_active" redis:"is_active"`
	LastSync     *time.Time   `json:"last_sync,omitempty" redis:"last_sync"`
	CreatedAt    time.Time    `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" redis:"updated_at"`
}

// ContactInfo represents brand contact information
type ContactInfo struct {
	Email       string `json:"email" redis:"email"`
	Phone       string `json:"phone" redis:"phone"`
	Address     string `json:"address" redis:"address"`
	City        string `json:"city" redis:"city"`
	State       string `json:"state" redis:"state"`
	Country     string `json:"country" redis:"country"`
	PostalCode  string `json:"postal_code" redis:"postal_code"`
	SupportEmail string `json:"support_email" redis:"support_email"`
	SupportPhone string `json:"support_phone" redis:"support_phone"`
}

// ComplianceRule represents a compliance rule for the brand
type ComplianceRule struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	BrandID     uuid.UUID `json:"brand_id" redis:"brand_id"`
	Name        string    `json:"name" redis:"name"`
	Description string    `json:"description" redis:"description"`
	Type        string    `json:"type" redis:"type"`
	Rule        string    `json:"rule" redis:"rule"`
	Severity    string    `json:"severity" redis:"severity"`
	IsActive    bool      `json:"is_active" redis:"is_active"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
}

// ContentTemplate represents a content template for the brand
type ContentTemplate struct {
	ID          uuid.UUID     `json:"id" redis:"id"`
	BrandID     uuid.UUID     `json:"brand_id" redis:"brand_id"`
	Name        string        `json:"name" redis:"name"`
	Description string        `json:"description" redis:"description"`
	Type        ContentType   `json:"type" redis:"type"`
	Category    ContentCategory `json:"category" redis:"category"`
	Template    string        `json:"template" redis:"template"`
	Variables   []string      `json:"variables" redis:"variables"`
	Platforms   []PlatformType `json:"platforms" redis:"platforms"`
	Tags        []string      `json:"tags" redis:"tags"`
	IsActive    bool          `json:"is_active" redis:"is_active"`
	UsageCount  int           `json:"usage_count" redis:"usage_count"`
	CreatedAt   time.Time     `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" redis:"updated_at"`
	CreatedBy   uuid.UUID     `json:"created_by" redis:"created_by"`
}

// HashtagSet represents a set of hashtags for the brand
type HashtagSet struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	BrandID     uuid.UUID `json:"brand_id" redis:"brand_id"`
	Name        string    `json:"name" redis:"name"`
	Description string    `json:"description" redis:"description"`
	Hashtags    []string  `json:"hashtags" redis:"hashtags"`
	Category    string    `json:"category" redis:"category"`
	Platform    PlatformType `json:"platform" redis:"platform"`
	IsActive    bool      `json:"is_active" redis:"is_active"`
	UsageCount  int       `json:"usage_count" redis:"usage_count"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
	CreatedBy   uuid.UUID `json:"created_by" redis:"created_by"`
}

// User represents a user in the social media management system
type User struct {
	ID           uuid.UUID              `json:"id" redis:"id"`
	Username     string                 `json:"username" redis:"username"`
	Email        string                 `json:"email" redis:"email"`
	FirstName    string                 `json:"first_name" redis:"first_name"`
	LastName     string                 `json:"last_name" redis:"last_name"`
	DisplayName  string                 `json:"display_name" redis:"display_name"`
	Avatar       string                 `json:"avatar" redis:"avatar"`
	Role         UserRole               `json:"role" redis:"role"`
	Permissions  []string               `json:"permissions" redis:"permissions"`
	BrandAccess  []uuid.UUID            `json:"brand_access" redis:"brand_access"`
	Status       UserStatus             `json:"status" redis:"status"`
	TimeZone     string                 `json:"time_zone" redis:"time_zone"`
	Language     string                 `json:"language" redis:"language"`
	Preferences  *UserPreferences       `json:"preferences,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields" redis:"custom_fields"`
	LastLoginAt  *time.Time             `json:"last_login_at,omitempty" redis:"last_login_at"`
	IsActive     bool                   `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy    uuid.UUID              `json:"created_by" redis:"created_by"`
	UpdatedBy    uuid.UUID              `json:"updated_by" redis:"updated_by"`
	Version      int64                  `json:"version" redis:"version"`
}

// UserRole defines the role of a user
type UserRole string

const (
	UserRoleAdmin        UserRole = "admin"
	UserRoleBrandManager UserRole = "brand_manager"
	UserRoleContentManager UserRole = "content_manager"
	UserRoleCreator      UserRole = "creator"
	UserRoleEditor       UserRole = "editor"
	UserRoleReviewer     UserRole = "reviewer"
	UserRoleApprover     UserRole = "approver"
	UserRoleAnalyst      UserRole = "analyst"
	UserRoleViewer       UserRole = "viewer"
)

// UserStatus defines the status of a user
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// UserPreferences represents user preferences
type UserPreferences struct {
	Theme               string            `json:"theme" redis:"theme"`
	Language            string            `json:"language" redis:"language"`
	TimeZone            string            `json:"time_zone" redis:"time_zone"`
	DateFormat          string            `json:"date_format" redis:"date_format"`
	TimeFormat          string            `json:"time_format" redis:"time_format"`
	NotificationSettings *NotificationSettings `json:"notification_settings,omitempty"`
	DefaultView         string            `json:"default_view" redis:"default_view"`
	AutoSave            bool              `json:"auto_save" redis:"auto_save"`
	ShowTutorials       bool              `json:"show_tutorials" redis:"show_tutorials"`
	CustomSettings      map[string]interface{} `json:"custom_settings" redis:"custom_settings"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	EmailNotifications    bool `json:"email_notifications" redis:"email_notifications"`
	PushNotifications     bool `json:"push_notifications" redis:"push_notifications"`
	SlackNotifications    bool `json:"slack_notifications" redis:"slack_notifications"`
	ContentApproval       bool `json:"content_approval" redis:"content_approval"`
	CampaignUpdates       bool `json:"campaign_updates" redis:"campaign_updates"`
	PerformanceAlerts     bool `json:"performance_alerts" redis:"performance_alerts"`
	ScheduleReminders     bool `json:"schedule_reminders" redis:"schedule_reminders"`
	MentionAlerts         bool `json:"mention_alerts" redis:"mention_alerts"`
	WeeklyReports         bool `json:"weekly_reports" redis:"weekly_reports"`
	MonthlyReports        bool `json:"monthly_reports" redis:"monthly_reports"`
}

// NewBrand creates a new brand with default values
func NewBrand(name, displayName, description string, createdBy uuid.UUID) *Brand {
	now := time.Now()
	return &Brand{
		ID:           uuid.New(),
		Name:         name,
		DisplayName:  displayName,
		Description:  description,
		Keywords:     []string{},
		Tags:         []string{},
		Languages:    []string{"en"},
		TimeZone:     "UTC",
		Status:       BrandStatusActive,
		CustomFields: make(map[string]interface{}),
		Metadata:     make(map[string]interface{}),
		ExternalIDs:  make(map[string]string),
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// NewUser creates a new user with default values
func NewUser(username, email, firstName, lastName string, role UserRole, createdBy uuid.UUID) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		DisplayName:  firstName + " " + lastName,
		Role:         role,
		Permissions:  []string{},
		BrandAccess:  []uuid.UUID{},
		Status:       UserStatusActive,
		TimeZone:     "UTC",
		Language:     "en",
		CustomFields: make(map[string]interface{}),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
		Version:      1,
	}
}

// AddSocialProfile adds a social profile to the brand
func (b *Brand) AddSocialProfile(profile *SocialProfile) {
	profile.BrandID = b.ID
	b.SocialProfiles = append(b.SocialProfiles, profile)
	b.UpdatedAt = time.Now()
	b.Version++
}

// GetSocialProfile gets a social profile by platform
func (b *Brand) GetSocialProfile(platform PlatformType) *SocialProfile {
	for _, profile := range b.SocialProfiles {
		if profile.Platform == platform && profile.IsActive {
			return profile
		}
	}
	return nil
}

// IsActive checks if the brand is active
func (b *Brand) IsActive() bool {
	return b.Status == BrandStatusActive && b.Active
}

// HasAccess checks if a user has access to the brand
func (u *User) HasAccess(brandID uuid.UUID) bool {
	if u.Role == UserRoleAdmin {
		return true
	}
	
	for _, id := range u.BrandAccess {
		if id == brandID {
			return true
		}
	}
	return false
}

// GrantBrandAccess grants access to a brand
func (u *User) GrantBrandAccess(brandID uuid.UUID) {
	if !u.HasAccess(brandID) {
		u.BrandAccess = append(u.BrandAccess, brandID)
		u.UpdatedAt = time.Now()
		u.Version++
	}
}

// RevokeBrandAccess revokes access to a brand
func (u *User) RevokeBrandAccess(brandID uuid.UUID) {
	for i, id := range u.BrandAccess {
		if id == brandID {
			u.BrandAccess = append(u.BrandAccess[:i], u.BrandAccess[i+1:]...)
			u.UpdatedAt = time.Now()
			u.Version++
			return
		}
	}
}
