package clickup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-coffee-ai-agents/internal/external/interfaces"
)

// Client implements the ClickUp API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string

	// Rate limiting
	rateLimiter *RateLimiter

	// Configuration
	config *Config
}

// Config holds ClickUp client configuration
type Config struct {
	APIKey     string        `yaml:"api_key" json:"api_key"`
	BaseURL    string        `yaml:"base_url" json:"base_url"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout"`
	RetryCount int           `yaml:"retry_count" json:"retry_count"`
	RateLimit  int           `yaml:"rate_limit" json:"rate_limit"`

	// Default workspace/team settings
	TeamID   string `yaml:"team_id" json:"team_id"`
	SpaceID  string `yaml:"space_id" json:"space_id"`
	FolderID string `yaml:"folder_id" json:"folder_id"`
	ListID   string `yaml:"list_id" json:"list_id"`
}

// RateLimiter implements rate limiting for ClickUp API
type RateLimiter struct {
	requestsPerMinute int
	requests          []time.Time
}

// NewClient creates a new ClickUp client
func NewClient(config *Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "https://api.clickup.com/api/v2"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RateLimit == 0 {
		config.RateLimit = 100 // ClickUp default rate limit
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		config:  config,
		rateLimiter: &RateLimiter{
			requestsPerMinute: config.RateLimit,
			requests:          make([]time.Time, 0),
		},
	}
}

// CreateTask creates a new task in ClickUp
func (c *Client) CreateTask(ctx context.Context, req *interfaces.CreateTaskRequest) (*interfaces.Task, error) {
	listID := c.config.ListID
	if req.ProjectID != "" {
		listID = req.ProjectID
	}

	if listID == "" {
		return nil, fmt.Errorf("list ID is required")
	}

	// Convert request to ClickUp format
	clickupReq := &ClickUpCreateTaskRequest{
		Name:         req.Name,
		Description:  req.Description,
		Assignees:    convertAssignees(req.AssigneeIDs),
		Tags:         req.Tags,
		Status:       convertTaskStatus(interfaces.TaskStatusOpen),
		Priority:     convertTaskPriority(req.Priority),
		DueDate:      convertTime(req.DueDate),
		StartDate:    convertTime(req.StartDate),
		TimeEstimate: convertDuration(req.TimeEstimate),
		CustomFields: convertCustomFields(req.CustomFields),
	}

	var response ClickUpTaskResponse
	err := c.makeRequest(ctx, "POST", fmt.Sprintf("/list/%s/task", listID), clickupReq, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return convertClickUpTask(&response), nil
}

// GetTask retrieves a task by ID
func (c *Client) GetTask(ctx context.Context, taskID string) (*interfaces.Task, error) {
	var response ClickUpTaskResponse
	err := c.makeRequest(ctx, "GET", fmt.Sprintf("/task/%s", taskID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return convertClickUpTask(&response), nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(ctx context.Context, taskID string, req *interfaces.UpdateTaskRequest) (*interfaces.Task, error) {
	// Convert request to ClickUp format
	clickupReq := &ClickUpUpdateTaskRequest{}

	if req.Name != nil {
		clickupReq.Name = *req.Name
	}
	if req.Description != nil {
		clickupReq.Description = *req.Description
	}
	if req.Status != nil {
		clickupReq.Status = convertTaskStatus(*req.Status)
	}
	if req.Priority != nil {
		clickupReq.Priority = convertTaskPriority(*req.Priority)
	}
	if req.DueDate != nil {
		clickupReq.DueDate = convertTime(req.DueDate)
	}
	if req.StartDate != nil {
		clickupReq.StartDate = convertTime(req.StartDate)
	}
	if req.Progress != nil {
		// ClickUp doesn't have direct progress field, use custom field
		if clickupReq.CustomFields == nil {
			clickupReq.CustomFields = make(map[string]interface{})
		}
		clickupReq.CustomFields["progress"] = *req.Progress
	}

	var response ClickUpTaskResponse
	err := c.makeRequest(ctx, "PUT", fmt.Sprintf("/task/%s", taskID), clickupReq, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return convertClickUpTask(&response), nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	err := c.makeRequest(ctx, "DELETE", fmt.Sprintf("/task/%s", taskID), nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// ListTasks lists tasks with filtering
func (c *Client) ListTasks(ctx context.Context, req *interfaces.ListTasksRequest) (*interfaces.TaskList, error) {
	listID := c.config.ListID
	if req.ProjectID != "" {
		listID = req.ProjectID
	}

	if listID == "" {
		return nil, fmt.Errorf("list ID is required")
	}

	// Build query parameters
	params := url.Values{}

	if len(req.Status) > 0 {
		statuses := make([]string, len(req.Status))
		for i, status := range req.Status {
			statuses[i] = convertTaskStatus(status)
		}
		params.Set("statuses", strings.Join(statuses, ","))
	}

	if len(req.AssigneeID) > 0 {
		params.Set("assignees", req.AssigneeID)
	}

	if req.DueBefore != nil {
		params.Set("due_date_lt", strconv.FormatInt(req.DueBefore.Unix()*1000, 10))
	}

	if req.DueAfter != nil {
		params.Set("due_date_gt", strconv.FormatInt(req.DueAfter.Unix()*1000, 10))
	}

	if req.OrderBy != "" {
		params.Set("order_by", req.OrderBy)
	}

	if req.Limit > 0 {
		params.Set("page", strconv.Itoa(req.Offset/req.Limit))
	}

	var response ClickUpTaskListResponse
	endpoint := fmt.Sprintf("/list/%s/task", listID)
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	err := c.makeRequest(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Convert response
	tasks := make([]*interfaces.Task, len(response.Tasks))
	for i, task := range response.Tasks {
		tasks[i] = convertClickUpTask(&task)
	}

	return &interfaces.TaskList{
		Tasks:   tasks,
		Total:   len(tasks),
		Limit:   req.Limit,
		Offset:  req.Offset,
		HasMore: len(tasks) == req.Limit,
	}, nil
}

// CreateProject creates a new project (list in ClickUp)
func (c *Client) CreateProject(ctx context.Context, req *interfaces.CreateProjectRequest) (*interfaces.Project, error) {
	folderID := c.config.FolderID
	if folderID == "" {
		return nil, fmt.Errorf("folder ID is required")
	}

	// Convert request to ClickUp format
	clickupReq := &ClickUpCreateListRequest{
		Name:    req.Name,
		Content: req.Description,
	}

	var response ClickUpListResponse
	err := c.makeRequest(ctx, "POST", fmt.Sprintf("/folder/%s/list", folderID), clickupReq, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return convertClickUpList(&response), nil
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(ctx context.Context, projectID string) (*interfaces.Project, error) {
	var response ClickUpListResponse
	err := c.makeRequest(ctx, "GET", fmt.Sprintf("/list/%s", projectID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return convertClickUpList(&response), nil
}

// AddComment adds a comment to a task
func (c *Client) AddComment(ctx context.Context, taskID string, req *interfaces.TaskCommentRequest) (*interfaces.Comment, error) {
	clickupReq := &ClickUpCreateCommentRequest{
		CommentText: req.Content,
		// Remove Assignee field as TaskCommentRequest has Content not Text
	}

	var response ClickUpCommentResponse
	err := c.makeRequest(ctx, "POST", fmt.Sprintf("/task/%s/comment", taskID), clickupReq, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return convertClickUpComment(&response), nil
}

// StartTimeTracking starts time tracking for a task
func (c *Client) StartTimeTracking(ctx context.Context, taskID, userID string) (*interfaces.TimeEntry, error) {
	clickupReq := &ClickUpStartTimeTrackingRequest{
		Description: "Time tracking started",
	}

	var response ClickUpTimeEntryResponse
	err := c.makeRequest(ctx, "POST", fmt.Sprintf("/team/%s/time_entries/start", c.config.TeamID), clickupReq, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to start time tracking: %w", err)
	}

	return convertClickUpTimeEntry(&response), nil
}

// StopTimeTracking stops time tracking
func (c *Client) StopTimeTracking(ctx context.Context, entryID string) (*interfaces.TimeEntry, error) {
	var response ClickUpTimeEntryResponse
	err := c.makeRequest(ctx, "POST", fmt.Sprintf("/team/%s/time_entries/%s/stop", c.config.TeamID, entryID), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to stop time tracking: %w", err)
	}

	return convertClickUpTimeEntry(&response), nil
}

// UpdateProject updates a project
func (c *Client) UpdateProject(ctx context.Context, projectID string, req *interfaces.UpdateProjectRequest) (*interfaces.Project, error) {
	// TODO: Implement project update
	return nil, fmt.Errorf("UpdateProject not implemented")
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	// TODO: Implement project deletion
	return fmt.Errorf("DeleteProject not implemented")
}

// ListProjects lists projects
func (c *Client) ListProjects(ctx context.Context, req *interfaces.ListProjectsRequest) (*interfaces.ProjectList, error) {
	// TODO: Implement project listing
	return nil, fmt.Errorf("ListProjects not implemented")
}

// AssignTask assigns a task to a user
func (c *Client) AssignTask(ctx context.Context, taskID, userID string) error {
	// TODO: Implement task assignment
	return fmt.Errorf("AssignTask not implemented")
}

// UnassignTask unassigns a task from a user
func (c *Client) UnassignTask(ctx context.Context, taskID, userID string) error {
	// TODO: Implement task unassignment
	return fmt.Errorf("UnassignTask not implemented")
}

// GetTimeEntries gets time entries
func (c *Client) GetTimeEntries(ctx context.Context, req *interfaces.TimeEntriesRequest) (*interfaces.TimeEntryList, error) {
	// TODO: Implement time entries retrieval
	return nil, fmt.Errorf("GetTimeEntries not implemented")
}

// GetComments gets comments for a task
func (c *Client) GetComments(ctx context.Context, taskID string) ([]*interfaces.Comment, error) {
	// TODO: Implement comments retrieval
	return nil, fmt.Errorf("GetComments not implemented")
}

// AddAttachment adds an attachment to a task
func (c *Client) AddAttachment(ctx context.Context, taskID string, req *interfaces.AttachmentRequest) (*interfaces.Attachment, error) {
	// TODO: Implement attachment addition
	return nil, fmt.Errorf("AddAttachment not implemented")
}

// RegisterWebhook registers a webhook
func (c *Client) RegisterWebhook(ctx context.Context, req *interfaces.TaskWebhookRequest) (*interfaces.TaskWebhook, error) {
	// TODO: Implement webhook registration
	return nil, fmt.Errorf("RegisterWebhook not implemented")
}

// UnregisterWebhook unregisters a webhook
func (c *Client) UnregisterWebhook(ctx context.Context, webhookID string) error {
	// TODO: Implement webhook unregistration
	return fmt.Errorf("UnregisterWebhook not implemented")
}

// BulkCreateTasks creates multiple tasks
func (c *Client) BulkCreateTasks(ctx context.Context, tasks []*interfaces.CreateTaskRequest) ([]*interfaces.Task, error) {
	// TODO: Implement bulk task creation
	return nil, fmt.Errorf("BulkCreateTasks not implemented")
}

// BulkUpdateTasks updates multiple tasks
func (c *Client) BulkUpdateTasks(ctx context.Context, updates []*interfaces.BulkTaskUpdate) ([]*interfaces.Task, error) {
	// TODO: Implement bulk task updates
	return nil, fmt.Errorf("BulkUpdateTasks not implemented")
}

// SearchTasks searches for tasks
func (c *Client) SearchTasks(ctx context.Context, req *interfaces.TaskSearchRequest) (*interfaces.TaskList, error) {
	// TODO: Implement task search
	return nil, fmt.Errorf("SearchTasks not implemented")
}

// GetProviderInfo returns provider information
func (c *Client) GetProviderInfo() *interfaces.ProviderInfo {
	return &interfaces.ProviderInfo{
		Name:         "ClickUp",
		Version:      "v2",
		Capabilities: []string{"tasks", "projects", "time_tracking", "comments", "attachments"},
		RateLimits: map[string]int{
			"requests_per_minute": c.config.RateLimit,
		},
	}
}

// makeRequest makes an HTTP request to the ClickUp API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	// Prepare request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode >= 400 {
		var errorResp ClickUpErrorResponse
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			return fmt.Errorf("ClickUp API error: %s", errorResp.Err)
		}
		return fmt.Errorf("ClickUp API error: status %d", resp.StatusCode)
	}

	// Parse response
	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Wait implements rate limiting
func (rl *RateLimiter) Wait(ctx context.Context) error {
	now := time.Now()

	// Remove requests older than 1 minute
	cutoff := now.Add(-time.Minute)
	var validRequests []time.Time
	for _, reqTime := range rl.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests = validRequests

	// Check if we can make a request
	if len(rl.requests) >= rl.requestsPerMinute {
		// Wait until the oldest request is more than 1 minute old
		waitTime := rl.requests[0].Add(time.Minute).Sub(now)
		if waitTime > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}
	}

	// Add current request
	rl.requests = append(rl.requests, now)
	return nil
}

// Helper functions for conversion
func convertAssignees(assigneeIDs []string) []int {
	if len(assigneeIDs) == 0 {
		return nil
	}

	assignees := make([]int, 0, len(assigneeIDs))
	for _, id := range assigneeIDs {
		if intID, err := strconv.Atoi(id); err == nil {
			assignees = append(assignees, intID)
		}
	}
	return assignees
}

func convertTaskStatus(status interfaces.TaskStatus) string {
	switch status {
	case interfaces.TaskStatusOpen:
		return "Open"
	case interfaces.TaskStatusInProgress:
		return "in progress"
	case interfaces.TaskStatusReview:
		return "review"
	case interfaces.TaskStatusDone:
		return "complete"
	case interfaces.TaskStatusCancelled:
		return "cancelled"
	default:
		return "Open"
	}
}

func convertTaskPriority(priority interfaces.TaskPriority) int {
	switch priority {
	case interfaces.TaskPriorityUrgent:
		return 1
	case interfaces.TaskPriorityHigh:
		return 2
	case interfaces.TaskPriorityNormal:
		return 3
	case interfaces.TaskPriorityLow:
		return 4
	default:
		return 3
	}
}

func convertTime(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	timestamp := t.Unix() * 1000
	return &timestamp
}

func convertDuration(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	milliseconds := int64(d.Milliseconds())
	return &milliseconds
}

func convertCustomFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}

	// ClickUp custom fields need specific formatting
	converted := make(map[string]interface{})
	for key, value := range fields {
		converted[key] = value
	}
	return converted
}

// ClickUp API request/response types
type ClickUpCreateTaskRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Assignees    []int                  `json:"assignees,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	DueDate      *int64                 `json:"due_date,omitempty"`
	StartDate    *int64                 `json:"start_date,omitempty"`
	TimeEstimate *int64                 `json:"time_estimate,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type ClickUpUpdateTaskRequest struct {
	Name         string                 `json:"name,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	DueDate      *int64                 `json:"due_date,omitempty"`
	StartDate    *int64                 `json:"start_date,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

type ClickUpTaskResponse struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	Description  string               `json:"description"`
	Status       ClickUpStatus        `json:"status"`
	Priority     ClickUpPriority      `json:"priority"`
	Assignees    []ClickUpUser        `json:"assignees"`
	Creator      ClickUpUser          `json:"creator"`
	DateCreated  string               `json:"date_created"`
	DateUpdated  string               `json:"date_updated"`
	DueDate      string               `json:"due_date"`
	StartDate    string               `json:"start_date"`
	TimeEstimate int64                `json:"time_estimate"`
	TimeSpent    int64                `json:"time_spent"`
	Tags         []ClickUpTag         `json:"tags"`
	List         ClickUpList          `json:"list"`
	Project      ClickUpProject       `json:"project"`
	Folder       ClickUpFolder        `json:"folder"`
	Space        ClickUpSpace         `json:"space"`
	URL          string               `json:"url"`
	CustomFields []ClickUpCustomField `json:"custom_fields"`
}

type ClickUpTaskListResponse struct {
	Tasks []ClickUpTaskResponse `json:"tasks"`
}

type ClickUpCreateListRequest struct {
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
}

type ClickUpListResponse struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Content    string        `json:"content"`
	OrderIndex int           `json:"orderindex"`
	Status     string        `json:"status"`
	Priority   string        `json:"priority"`
	Assignee   ClickUpUser   `json:"assignee"`
	TaskCount  int           `json:"task_count"`
	DueDate    string        `json:"due_date"`
	StartDate  string        `json:"start_date"`
	Folder     ClickUpFolder `json:"folder"`
	Space      ClickUpSpace  `json:"space"`
	Archived   bool          `json:"archived"`
}

type ClickUpCreateCommentRequest struct {
	CommentText string `json:"comment_text"`
	Assignee    string `json:"assignee,omitempty"`
}

type ClickUpCommentResponse struct {
	ID          string               `json:"id"`
	Comment     []ClickUpCommentText `json:"comment"`
	CommentText string               `json:"comment_text"`
	User        ClickUpUser          `json:"user"`
	Date        string               `json:"date"`
}

type ClickUpCommentText struct {
	Text string `json:"text"`
}

type ClickUpStartTimeTrackingRequest struct {
	Description string       `json:"description,omitempty"`
	Tags        []ClickUpTag `json:"tags,omitempty"`
	Billable    bool         `json:"billable,omitempty"`
}

type ClickUpTimeEntryResponse struct {
	ID          string       `json:"id"`
	Task        ClickUpTask  `json:"task"`
	User        ClickUpUser  `json:"user"`
	Billable    bool         `json:"billable"`
	Start       string       `json:"start"`
	End         string       `json:"end"`
	Duration    string       `json:"duration"`
	Description string       `json:"description"`
	Tags        []ClickUpTag `json:"tags"`
	Source      string       `json:"source"`
	DateCreated string       `json:"date_created"`
}

type ClickUpErrorResponse struct {
	Err   string `json:"err"`
	ECODE string `json:"ECODE"`
}

// ClickUp data types
type ClickUpStatus struct {
	Status     string `json:"status"`
	Color      string `json:"color"`
	Type       string `json:"type"`
	OrderIndex int    `json:"orderindex"`
}

type ClickUpPriority struct {
	Priority   string `json:"priority"`
	Color      string `json:"color"`
	OrderIndex int    `json:"orderindex"`
}

type ClickUpUser struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Color          string `json:"color"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
}

type ClickUpTag struct {
	Name    string `json:"name"`
	TagFg   string `json:"tag_fg"`
	TagBg   string `json:"tag_bg"`
	Creator int    `json:"creator"`
}

type ClickUpList struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Access bool   `json:"access"`
}

type ClickUpProject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
	Access bool   `json:"access"`
}

type ClickUpFolder struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
	Access bool   `json:"access"`
}

type ClickUpSpace struct {
	ID string `json:"id"`
}

type ClickUpTask struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ClickUpCustomField struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// Conversion functions
func convertClickUpTask(task *ClickUpTaskResponse) *interfaces.Task {
	// Parse timestamps
	createdAt, _ := time.Parse("2006-01-02T15:04:05.000Z", task.DateCreated)
	updatedAt, _ := time.Parse("2006-01-02T15:04:05.000Z", task.DateUpdated)

	var dueDate *time.Time
	if task.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", task.DueDate); err == nil {
			dueDate = &parsed
		}
	}

	var startDate *time.Time
	if task.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", task.StartDate); err == nil {
			startDate = &parsed
		}
	}

	// Convert assignees
	assigneeIDs := make([]string, len(task.Assignees))
	for i, assignee := range task.Assignees {
		assigneeIDs[i] = strconv.Itoa(assignee.ID)
	}

	// Convert tags
	tags := make([]string, len(task.Tags))
	for i, tag := range task.Tags {
		tags[i] = tag.Name
	}

	// Convert custom fields
	customFields := make(map[string]interface{})
	for _, field := range task.CustomFields {
		customFields[field.Name] = field.Value
	}

	// Convert time estimates and spent
	var timeEstimate *time.Duration
	if task.TimeEstimate > 0 {
		duration := time.Duration(task.TimeEstimate) * time.Millisecond
		timeEstimate = &duration
	}

	var timeSpent *time.Duration
	if task.TimeSpent > 0 {
		duration := time.Duration(task.TimeSpent) * time.Millisecond
		timeSpent = &duration
	}

	return &interfaces.Task{
		ID:           task.ID,
		Name:         task.Name,
		Description:  task.Description,
		Status:       convertClickUpTaskStatus(task.Status.Status),
		Priority:     convertClickUpTaskPriority(task.Priority.Priority),
		ProjectID:    task.List.ID,
		AssigneeIDs:  assigneeIDs,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		DueDate:      dueDate,
		StartDate:    startDate,
		Tags:         tags,
		CustomFields: customFields,
		TimeEstimate: timeEstimate,
		TimeSpent:    timeSpent,
		URL:          task.URL,
		ExternalID:   task.ID,
	}
}

func convertClickUpList(list *ClickUpListResponse) *interfaces.Project {
	// Parse timestamps - ClickUp doesn't provide created/updated dates for lists
	now := time.Now()

	var startDate *time.Time
	if list.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", list.StartDate); err == nil {
			startDate = &parsed
		}
	}

	var endDate *time.Time
	if list.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", list.DueDate); err == nil {
			endDate = &parsed
		}
	}

	status := interfaces.ProjectStatusActive
	if list.Archived {
		status = interfaces.ProjectStatusArchived
	}

	return &interfaces.Project{
		ID:          list.ID,
		Name:        list.Name,
		Description: list.Content,
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
		StartDate:   startDate,
		EndDate:     endDate,
		TaskCount:   list.TaskCount,
		URL:         fmt.Sprintf("https://app.clickup.com/t/%s", list.ID),
		ExternalID:  list.ID,
	}
}

func convertClickUpComment(comment *ClickUpCommentResponse) *interfaces.Comment {
	createdAt, _ := time.Parse("2006-01-02T15:04:05.000Z", comment.Date)

	// Extract content from CommentText or Comment array
	content := comment.CommentText
	if content == "" && len(comment.Comment) > 0 {
		// If CommentText is empty, concatenate text from Comment array
		var texts []string
		for _, c := range comment.Comment {
			if c.Text != "" {
				texts = append(texts, c.Text)
			}
		}
		content = strings.Join(texts, " ")
	}

	return &interfaces.Comment{
		ID:        comment.ID,
		TaskID:    "", // Will be set by caller
		UserID:    strconv.Itoa(comment.User.ID),
		Content:   content,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
}

func convertClickUpTimeEntry(entry *ClickUpTimeEntryResponse) *interfaces.TimeEntry {
	createdAt, _ := time.Parse("2006-01-02T15:04:05.000Z", entry.DateCreated)
	startTime, _ := time.Parse("2006-01-02T15:04:05.000Z", entry.Start)

	var endTime *time.Time
	if entry.End != "" {
		if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", entry.End); err == nil {
			endTime = &parsed
		}
	}

	// Parse duration (ClickUp returns duration as string in milliseconds)
	var duration time.Duration
	if durationMs, err := strconv.ParseInt(entry.Duration, 10, 64); err == nil {
		duration = time.Duration(durationMs) * time.Millisecond
	}

	return &interfaces.TimeEntry{
		ID:          entry.ID,
		TaskID:      entry.Task.ID,
		UserID:      strconv.Itoa(entry.User.ID),
		Description: entry.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Duration:    duration,
		Billable:    entry.Billable,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
}

func convertClickUpTaskStatus(status string) interfaces.TaskStatus {
	switch strings.ToLower(status) {
	case "open", "to do":
		return interfaces.TaskStatusOpen
	case "in progress", "in review":
		return interfaces.TaskStatusInProgress
	case "review", "reviewing":
		return interfaces.TaskStatusReview
	case "complete", "done", "closed":
		return interfaces.TaskStatusDone
	case "cancelled", "canceled":
		return interfaces.TaskStatusCancelled
	default:
		return interfaces.TaskStatusOpen
	}
}

func convertClickUpTaskPriority(priority string) interfaces.TaskPriority {
	switch strings.ToLower(priority) {
	case "urgent":
		return interfaces.TaskPriorityUrgent
	case "high":
		return interfaces.TaskPriorityHigh
	case "normal":
		return interfaces.TaskPriorityNormal
	case "low":
		return interfaces.TaskPriorityLow
	default:
		return interfaces.TaskPriorityNormal
	}
}
