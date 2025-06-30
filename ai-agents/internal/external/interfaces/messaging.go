package interfaces

import (
	"context"
	"time"
)

// MessagingProvider defines the interface for messaging platforms
type MessagingProvider interface {
	// Message operations
	SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error)
	UpdateMessage(ctx context.Context, messageID string, req *UpdateMessageRequest) (*Message, error)
	DeleteMessage(ctx context.Context, messageID string) error
	GetMessage(ctx context.Context, messageID string) (*Message, error)
	
	// Channel operations
	CreateChannel(ctx context.Context, req *CreateChannelRequest) (*Channel, error)
	GetChannel(ctx context.Context, channelID string) (*Channel, error)
	UpdateChannel(ctx context.Context, channelID string, req *UpdateChannelRequest) (*Channel, error)
	DeleteChannel(ctx context.Context, channelID string) error
	ListChannels(ctx context.Context, req *ListChannelsRequest) (*ChannelList, error)
	
	// User operations
	GetUser(ctx context.Context, userID string) (*MessagingUser, error)
	ListUsers(ctx context.Context, req *ListUsersRequest) (*MessagingUserList, error)
	
	// File operations
	UploadFile(ctx context.Context, req *FileUploadRequest) (*File, error)
	GetFile(ctx context.Context, fileID string) (*File, error)
	DeleteFile(ctx context.Context, fileID string) error
	
	// Bot operations
	SetBotPresence(ctx context.Context, presence BotPresence) error
	GetBotInfo(ctx context.Context) (*BotInfo, error)
	
	// Interactive components
	SendInteractiveMessage(ctx context.Context, req *InteractiveMessageRequest) (*Message, error)
	HandleInteraction(ctx context.Context, interaction *Interaction) (*InteractionResponse, error)
	
	// Webhooks and events
	RegisterWebhook(ctx context.Context, req *MessagingWebhookRequest) (*MessagingWebhook, error)
	UnregisterWebhook(ctx context.Context, webhookID string) error
	HandleWebhookEvent(ctx context.Context, event *WebhookEvent) error
	
	// Bulk operations
	BulkSendMessages(ctx context.Context, messages []*SendMessageRequest) ([]*Message, error)
	
	// Search
	SearchMessages(ctx context.Context, req *MessagingSearchRequest) (*MessageSearchResult, error)
	
	// Provider info
	GetProviderInfo() *ProviderInfo
}

// Message represents a message in the messaging platform
type Message struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	UserID      string                 `json:"user_id"`
	Text        string                 `json:"text"`
	
	// Message formatting
	Blocks      []MessageBlock         `json:"blocks,omitempty"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
	
	// Message metadata
	Type        MessageType            `json:"type"`
	Subtype     string                 `json:"subtype,omitempty"`
	ThreadID    string                 `json:"thread_id,omitempty"`
	
	// Timestamps
	Timestamp   time.Time              `json:"timestamp"`
	EditedAt    *time.Time             `json:"edited_at,omitempty"`
	
	// Reactions and interactions
	Reactions   []Reaction             `json:"reactions,omitempty"`
	ReplyCount  int                    `json:"reply_count"`
	
	// Files and media
	Files       []File                 `json:"files,omitempty"`
	
	// Bot information
	BotID       string                 `json:"bot_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	
	// External references
	ExternalID  string                 `json:"external_id,omitempty"`
	URL         string                 `json:"url,omitempty"`
}

// Channel represents a channel in the messaging platform
type Channel struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Topic       string                 `json:"topic,omitempty"`
	Purpose     string                 `json:"purpose,omitempty"`
	
	// Channel properties
	Type        ChannelType            `json:"type"`
	IsPrivate   bool                   `json:"is_private"`
	IsArchived  bool                   `json:"is_archived"`
	
	// Membership
	MemberCount int                    `json:"member_count"`
	Members     []string               `json:"members,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	
	// Creator and admin info
	CreatorID   string                 `json:"creator_id"`
	AdminIDs    []string               `json:"admin_ids,omitempty"`
	
	// External references
	ExternalID  string                 `json:"external_id,omitempty"`
	URL         string                 `json:"url,omitempty"`
}

// User represents a user in the messaging platform
type MessagingUser struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	RealName    string                 `json:"real_name,omitempty"`
	DisplayName string                 `json:"display_name,omitempty"`
	Email       string                 `json:"email,omitempty"`
	
	// User properties
	IsBot       bool                   `json:"is_bot"`
	IsAdmin     bool                   `json:"is_admin"`
	IsOwner     bool                   `json:"is_owner"`
	IsActive    bool                   `json:"is_active"`
	
	// Profile information
	Avatar      string                 `json:"avatar,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Phone       string                 `json:"phone,omitempty"`
	Timezone    string                 `json:"timezone,omitempty"`
	
	// Status
	Status      UserStatus             `json:"status"`
	StatusText  string                 `json:"status_text,omitempty"`
	StatusEmoji string                 `json:"status_emoji,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastSeen    *time.Time             `json:"last_seen,omitempty"`
}

// File represents a file in the messaging platform
type File struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Title       string                 `json:"title,omitempty"`
	MimeType    string                 `json:"mime_type"`
	Size        int64                  `json:"size"`
	
	// File URLs
	URL         string                 `json:"url"`
	DownloadURL string                 `json:"download_url,omitempty"`
	ThumbnailURL string                `json:"thumbnail_url,omitempty"`
	
	// File metadata
	IsPublic    bool                   `json:"is_public"`
	IsExternal  bool                   `json:"is_external"`
	
	// Upload information
	UploaderID  string                 `json:"uploader_id"`
	ChannelID   string                 `json:"channel_id,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// BotInfo represents bot information
type BotInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AppID       string                 `json:"app_id"`
	UserID      string                 `json:"user_id,omitempty"`
	
	// Bot properties
	IsActive    bool                   `json:"is_active"`
	Avatar      string                 `json:"avatar,omitempty"`
	
	// Permissions
	Scopes      []string               `json:"scopes,omitempty"`
	
	// Timestamps
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// MessageBlock represents a block in a structured message
type MessageBlock struct {
	Type        string                 `json:"type"`
	Text        *TextBlock             `json:"text,omitempty"`
	Elements    []BlockElement         `json:"elements,omitempty"`
	Accessory   *BlockElement          `json:"accessory,omitempty"`
	Fields      []TextBlock            `json:"fields,omitempty"`
	BlockID     string                 `json:"block_id,omitempty"`
}

// TextBlock represents a text block
type TextBlock struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text"`
	Emoji       bool                   `json:"emoji,omitempty"`
	Verbatim    bool                   `json:"verbatim,omitempty"`
}

// BlockElement represents an interactive element
type BlockElement struct {
	Type        string                 `json:"type"`
	Text        *TextBlock             `json:"text,omitempty"`
	ActionID    string                 `json:"action_id,omitempty"`
	Value       string                 `json:"value,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Style       string                 `json:"style,omitempty"`
	Options     []Option               `json:"options,omitempty"`
	Confirm     *ConfirmDialog         `json:"confirm,omitempty"`
}

// Option represents an option in a select element
type Option struct {
	Text        *TextBlock             `json:"text"`
	Value       string                 `json:"value"`
	Description *TextBlock             `json:"description,omitempty"`
	URL         string                 `json:"url,omitempty"`
}

// ConfirmDialog represents a confirmation dialog
type ConfirmDialog struct {
	Title       *TextBlock             `json:"title"`
	Text        *TextBlock             `json:"text"`
	Confirm     *TextBlock             `json:"confirm"`
	Deny        *TextBlock             `json:"deny"`
	Style       string                 `json:"style,omitempty"`
}

// MessageAttachment represents a message attachment
type MessageAttachment struct {
	Color       string                 `json:"color,omitempty"`
	Fallback    string                 `json:"fallback,omitempty"`
	Title       string                 `json:"title,omitempty"`
	TitleLink   string                 `json:"title_link,omitempty"`
	Text        string                 `json:"text,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	ThumbURL    string                 `json:"thumb_url,omitempty"`
	Footer      string                 `json:"footer,omitempty"`
	FooterIcon  string                 `json:"footer_icon,omitempty"`
	Timestamp   *time.Time             `json:"ts,omitempty"`
	Fields      []AttachmentField      `json:"fields,omitempty"`
	Actions     []AttachmentAction     `json:"actions,omitempty"`
}

// AttachmentField represents a field in an attachment
type AttachmentField struct {
	Title       string                 `json:"title"`
	Value       string                 `json:"value"`
	Short       bool                   `json:"short"`
}

// AttachmentAction represents an action in an attachment
type AttachmentAction struct {
	Name        string                 `json:"name"`
	Text        string                 `json:"text"`
	Type        string                 `json:"type"`
	Value       string                 `json:"value,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Style       string                 `json:"style,omitempty"`
	Confirm     *ConfirmDialog         `json:"confirm,omitempty"`
}

// Reaction represents a reaction to a message
type Reaction struct {
	Name        string                 `json:"name"`
	Count       int                    `json:"count"`
	Users       []string               `json:"users,omitempty"`
}

// Interaction represents an interactive component interaction
type Interaction struct {
	Type        string                 `json:"type"`
	Token       string                 `json:"token"`
	ActionID    string                 `json:"action_id,omitempty"`
	BlockID     string                 `json:"block_id,omitempty"`
	Value       string                 `json:"value,omitempty"`
	UserID      string                 `json:"user_id"`
	ChannelID   string                 `json:"channel_id"`
	MessageID   string                 `json:"message_id,omitempty"`
	TriggerID   string                 `json:"trigger_id,omitempty"`
	ResponseURL string                 `json:"response_url,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// InteractionResponse represents a response to an interaction
type InteractionResponse struct {
	ResponseType string                 `json:"response_type"`
	Text         string                 `json:"text,omitempty"`
	Blocks       []MessageBlock         `json:"blocks,omitempty"`
	Attachments  []MessageAttachment    `json:"attachments,omitempty"`
	ReplaceOriginal bool                `json:"replace_original,omitempty"`
	DeleteOriginal  bool                `json:"delete_original,omitempty"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	Type        string                 `json:"type"`
	EventID     string                 `json:"event_id"`
	EventTime   time.Time              `json:"event_time"`
	Data        map[string]interface{} `json:"data"`
	Signature   string                 `json:"signature,omitempty"`
}

// Request types
type SendMessageRequest struct {
	ChannelID   string                 `json:"channel_id"`
	Text        string                 `json:"text,omitempty"`
	Blocks      []MessageBlock         `json:"blocks,omitempty"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
	ThreadID    string                 `json:"thread_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	IconEmoji   string                 `json:"icon_emoji,omitempty"`
	IconURL     string                 `json:"icon_url,omitempty"`
	LinkNames   bool                   `json:"link_names,omitempty"`
	Parse       string                 `json:"parse,omitempty"`
	ReplyBroadcast bool                `json:"reply_broadcast,omitempty"`
	UnfurlLinks bool                   `json:"unfurl_links,omitempty"`
	UnfurlMedia bool                   `json:"unfurl_media,omitempty"`
}

type UpdateMessageRequest struct {
	Text        *string                `json:"text,omitempty"`
	Blocks      []MessageBlock         `json:"blocks,omitempty"`
	Attachments []MessageAttachment    `json:"attachments,omitempty"`
}

type CreateChannelRequest struct {
	Name        string                 `json:"name"`
	IsPrivate   bool                   `json:"is_private,omitempty"`
	Topic       string                 `json:"topic,omitempty"`
	Purpose     string                 `json:"purpose,omitempty"`
	UserIDs     []string               `json:"user_ids,omitempty"`
}

type UpdateChannelRequest struct {
	Name        *string                `json:"name,omitempty"`
	Topic       *string                `json:"topic,omitempty"`
	Purpose     *string                `json:"purpose,omitempty"`
}

type ListChannelsRequest struct {
	Types       []ChannelType          `json:"types,omitempty"`
	ExcludeArchived bool               `json:"exclude_archived,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Cursor      string                 `json:"cursor,omitempty"`
}

type ListUsersRequest struct {
	IncludeLocale bool                 `json:"include_locale,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Cursor      string                 `json:"cursor,omitempty"`
}

type FileUploadRequest struct {
	Filename    string                 `json:"filename"`
	Content     []byte                 `json:"content"`
	MimeType    string                 `json:"mime_type,omitempty"`
	Title       string                 `json:"title,omitempty"`
	ChannelID   string                 `json:"channel_id,omitempty"`
	ThreadID    string                 `json:"thread_id,omitempty"`
	InitialComment string              `json:"initial_comment,omitempty"`
}

type InteractiveMessageRequest struct {
	ChannelID   string                 `json:"channel_id"`
	Text        string                 `json:"text,omitempty"`
	Blocks      []MessageBlock         `json:"blocks"`
	ThreadID    string                 `json:"thread_id,omitempty"`
}

type MessagingWebhookRequest struct {
	URL         string                 `json:"url"`
	Events      []string               `json:"events"`
	Secret      string                 `json:"secret,omitempty"`
}

type MessagingSearchRequest struct {
	Query       string                 `json:"query"`
	Sort        string                 `json:"sort,omitempty"`
	SortDir     string                 `json:"sort_dir,omitempty"`
	Highlight   bool                   `json:"highlight,omitempty"`
	Count       int                    `json:"count,omitempty"`
	Page        int                    `json:"page,omitempty"`
}

// Response types
type ChannelList struct {
	Channels    []*Channel             `json:"channels"`
	NextCursor  string                 `json:"next_cursor,omitempty"`
	HasMore     bool                   `json:"has_more"`
}

type MessagingUserList struct {
	Users       []*MessagingUser       `json:"users"`
	NextCursor  string                 `json:"next_cursor,omitempty"`
	HasMore     bool                   `json:"has_more"`
}

type MessageSearchResult struct {
	Messages    []*Message             `json:"messages"`
	Total       int                    `json:"total"`
	Page        int                    `json:"page"`
	PerPage     int                    `json:"per_page"`
	HasMore     bool                   `json:"has_more"`
}

// Enums
type MessageType string

const (
	MessageTypeMessage     MessageType = "message"
	MessageTypeChannelJoin MessageType = "channel_join"
	MessageTypeChannelLeave MessageType = "channel_leave"
	MessageTypeChannelTopic MessageType = "channel_topic"
	MessageTypeChannelPurpose MessageType = "channel_purpose"
	MessageTypeChannelName MessageType = "channel_name"
	MessageTypeChannelArchive MessageType = "channel_archive"
	MessageTypeChannelUnarchive MessageType = "channel_unarchive"
)

type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeDM      ChannelType = "dm"
	ChannelTypeGroup   ChannelType = "group"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusAway   UserStatus = "away"
	UserStatusDND    UserStatus = "dnd"
)

type BotPresence string

const (
	BotPresenceAuto BotPresence = "auto"
	BotPresenceAway BotPresence = "away"
)

// Webhook represents a webhook configuration
type MessagingWebhook struct {
	ID          string                 `json:"id"`
	URL         string                 `json:"url"`
	Events      []string               `json:"events"`
	Secret      string                 `json:"secret,omitempty"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
