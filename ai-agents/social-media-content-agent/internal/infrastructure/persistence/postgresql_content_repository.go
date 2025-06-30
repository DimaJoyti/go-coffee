package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/entities"
	"go-coffee-ai-agents/social-media-content-agent/internal/domain/repositories"
)

// PostgreSQLContentRepository implements the ContentRepository interface
type PostgreSQLContentRepository struct {
	db    *sql.DB
	cache *redis.Client
}

// NewPostgreSQLContentRepository creates a new PostgreSQL content repository
func NewPostgreSQLContentRepository(db *sql.DB, cache *redis.Client) *PostgreSQLContentRepository {
	return &PostgreSQLContentRepository{
		db:    db,
		cache: cache,
	}
}

// Create creates a new content record
func (r *PostgreSQLContentRepository) Create(ctx context.Context, content *entities.Content) error {
	query := `
		INSERT INTO content (
			id, title, body, type, format, status, priority, category,
			brand_id, campaign_id, creator_id, approver_id, platforms,
			hashtags, mentions, tags, keywords, tone, language,
			scheduled_at, published_at, expires_at, target_audience,
			custom_fields, metadata, external_ids, is_template,
			template_id, ai_generated, ai_prompt, is_archived,
			created_at, updated_at, created_by, updated_by, version
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36
		)`

	// Convert slices to PostgreSQL arrays
	platforms := pq.Array(content.Platforms)
	hashtags := pq.Array(content.Hashtags)
	mentions := pq.Array(content.Mentions)
	tags := pq.Array(content.Tags)
	keywords := pq.Array(content.Keywords)

	// Convert maps to JSON
	customFields, _ := json.Marshal(content.CustomFields)
	metadata, _ := json.Marshal(content.Metadata)
	externalIDs, _ := json.Marshal(content.ExternalIDs)
	targetAudience, _ := json.Marshal(content.TargetAudience)

	_, err := r.db.ExecContext(ctx, query,
		content.ID, content.Title, content.Body, content.Type, content.Format,
		content.Status, content.Priority, content.Category, content.BrandID,
		content.CampaignID, content.CreatorID, content.ApproverID, platforms,
		hashtags, mentions, tags, keywords, content.Tone, content.Language,
		content.ScheduledAt, content.PublishedAt, content.ExpiresAt, targetAudience,
		customFields, metadata, externalIDs, content.IsTemplate, content.TemplateID,
		content.AIGenerated, content.AIPrompt, content.IsArchived,
		content.CreatedAt, content.UpdatedAt, content.CreatedBy, content.UpdatedBy, content.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to create content: %w", err)
	}

	// Cache the content
	r.cacheContent(ctx, content)

	return nil
}

// GetByID retrieves content by ID
func (r *PostgreSQLContentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Content, error) {
	// Try cache first
	if content := r.getCachedContent(ctx, id); content != nil {
		return content, nil
	}

	query := `
		SELECT id, title, body, type, format, status, priority, category,
			   brand_id, campaign_id, creator_id, approver_id, platforms,
			   hashtags, mentions, tags, keywords, tone, language,
			   scheduled_at, published_at, expires_at, target_audience,
			   custom_fields, metadata, external_ids, is_template,
			   template_id, ai_generated, ai_prompt, is_archived,
			   created_at, updated_at, created_by, updated_by, version
		FROM content WHERE id = $1 AND is_archived = false`

	row := r.db.QueryRowContext(ctx, query, id)
	content, err := r.scanContent(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	// Cache the result
	r.cacheContent(ctx, content)

	return content, nil
}

// Update updates an existing content record
func (r *PostgreSQLContentRepository) Update(ctx context.Context, content *entities.Content) error {
	query := `
		UPDATE content SET
			title = $2, body = $3, type = $4, format = $5, status = $6,
			priority = $7, category = $8, campaign_id = $9, approver_id = $10,
			platforms = $11, hashtags = $12, mentions = $13, tags = $14,
			keywords = $15, tone = $16, language = $17, scheduled_at = $18,
			published_at = $19, expires_at = $20, target_audience = $21,
			custom_fields = $22, metadata = $23, external_ids = $24,
			is_template = $25, template_id = $26, ai_generated = $27,
			ai_prompt = $28, is_archived = $29, updated_at = $30,
			updated_by = $31, version = $32
		WHERE id = $1 AND version = $33`

	// Convert data for PostgreSQL
	platforms := pq.Array(content.Platforms)
	hashtags := pq.Array(content.Hashtags)
	mentions := pq.Array(content.Mentions)
	tags := pq.Array(content.Tags)
	keywords := pq.Array(content.Keywords)

	customFields, _ := json.Marshal(content.CustomFields)
	metadata, _ := json.Marshal(content.Metadata)
	externalIDs, _ := json.Marshal(content.ExternalIDs)
	targetAudience, _ := json.Marshal(content.TargetAudience)

	oldVersion := content.Version - 1
	result, err := r.db.ExecContext(ctx, query,
		content.ID, content.Title, content.Body, content.Type, content.Format,
		content.Status, content.Priority, content.Category, content.CampaignID,
		content.ApproverID, platforms, hashtags, mentions, tags, keywords,
		content.Tone, content.Language, content.ScheduledAt, content.PublishedAt,
		content.ExpiresAt, targetAudience, customFields, metadata, externalIDs,
		content.IsTemplate, content.TemplateID, content.AIGenerated, content.AIPrompt,
		content.IsArchived, content.UpdatedAt, content.UpdatedBy, content.Version,
		oldVersion,
	)

	if err != nil {
		return fmt.Errorf("failed to update content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("content not found or version conflict: %s", content.ID)
	}

	// Update cache
	r.cacheContent(ctx, content)

	return nil
}

// Delete soft deletes a content record
func (r *PostgreSQLContentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE content SET is_archived = true, updated_at = $2 WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete content: %w", err)
	}

	// Remove from cache
	r.removeCachedContent(ctx, id)

	return nil
}

// List retrieves content with filtering
func (r *PostgreSQLContentRepository) List(ctx context.Context, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	query, args := r.buildListQuery(filter)
	
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// ListByBrand retrieves content by brand ID
func (r *PostgreSQLContentRepository) ListByBrand(ctx context.Context, brandID uuid.UUID, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	if filter == nil {
		filter = &repositories.ContentFilter{}
	}
	filter.BrandIDs = []uuid.UUID{brandID}
	return r.List(ctx, filter)
}

// ListByCampaign retrieves content by campaign ID
func (r *PostgreSQLContentRepository) ListByCampaign(ctx context.Context, campaignID uuid.UUID, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	if filter == nil {
		filter = &repositories.ContentFilter{}
	}
	filter.CampaignIDs = []uuid.UUID{campaignID}
	return r.List(ctx, filter)
}

// ListByCreator retrieves content by creator ID
func (r *PostgreSQLContentRepository) ListByCreator(ctx context.Context, creatorID uuid.UUID, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	if filter == nil {
		filter = &repositories.ContentFilter{}
	}
	filter.CreatorIDs = []uuid.UUID{creatorID}
	return r.List(ctx, filter)
}

// ListByStatus retrieves content by status
func (r *PostgreSQLContentRepository) ListByStatus(ctx context.Context, status entities.ContentStatus, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	if filter == nil {
		filter = &repositories.ContentFilter{}
	}
	filter.Statuses = []entities.ContentStatus{status}
	return r.List(ctx, filter)
}

// ListByPlatform retrieves content by platform
func (r *PostgreSQLContentRepository) ListByPlatform(ctx context.Context, platform entities.PlatformType, filter *repositories.ContentFilter) ([]*entities.Content, error) {
	if filter == nil {
		filter = &repositories.ContentFilter{}
	}
	filter.Platforms = []entities.PlatformType{platform}
	return r.List(ctx, filter)
}

// GetScheduledContent retrieves content scheduled within a time range
func (r *PostgreSQLContentRepository) GetScheduledContent(ctx context.Context, from, to time.Time) ([]*entities.Content, error) {
	query := `
		SELECT id, title, body, type, format, status, priority, category,
			   brand_id, campaign_id, creator_id, approver_id, platforms,
			   hashtags, mentions, tags, keywords, tone, language,
			   scheduled_at, published_at, expires_at, target_audience,
			   custom_fields, metadata, external_ids, is_template,
			   template_id, ai_generated, ai_prompt, is_archived,
			   created_at, updated_at, created_by, updated_by, version
		FROM content 
		WHERE status = 'scheduled' 
		  AND scheduled_at BETWEEN $1 AND $2 
		  AND is_archived = false
		ORDER BY scheduled_at ASC`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// GetContentDueForPublishing retrieves content that should be published now
func (r *PostgreSQLContentRepository) GetContentDueForPublishing(ctx context.Context) ([]*entities.Content, error) {
	now := time.Now()
	return r.GetScheduledContent(ctx, time.Time{}, now)
}

// GetExpiredContent retrieves expired content
func (r *PostgreSQLContentRepository) GetExpiredContent(ctx context.Context) ([]*entities.Content, error) {
	query := `
		SELECT id, title, body, type, format, status, priority, category,
			   brand_id, campaign_id, creator_id, approver_id, platforms,
			   hashtags, mentions, tags, keywords, tone, language,
			   scheduled_at, published_at, expires_at, target_audience,
			   custom_fields, metadata, external_ids, is_template,
			   template_id, ai_generated, ai_prompt, is_archived,
			   created_at, updated_at, created_by, updated_by, version
		FROM content 
		WHERE expires_at < $1 
		  AND is_archived = false
		ORDER BY expires_at ASC`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired content: %w", err)
	}
	defer rows.Close()

	var contents []*entities.Content
	for rows.Next() {
		content, err := r.scanContent(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content: %w", err)
		}
		contents = append(contents, content)
	}

	return contents, nil
}

// Helper methods

func (r *PostgreSQLContentRepository) buildListQuery(filter *repositories.ContentFilter) (string, []interface{}) {
	query := `
		SELECT id, title, body, type, format, status, priority, category,
			   brand_id, campaign_id, creator_id, approver_id, platforms,
			   hashtags, mentions, tags, keywords, tone, language,
			   scheduled_at, published_at, expires_at, target_audience,
			   custom_fields, metadata, external_ids, is_template,
			   template_id, ai_generated, ai_prompt, is_archived,
			   created_at, updated_at, created_by, updated_by, version
		FROM content WHERE is_archived = false`

	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter != nil {
		if len(filter.BrandIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("brand_id = ANY($%d)", argIndex))
			args = append(args, pq.Array(filter.BrandIDs))
			argIndex++
		}

		if len(filter.CampaignIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("campaign_id = ANY($%d)", argIndex))
			args = append(args, pq.Array(filter.CampaignIDs))
			argIndex++
		}

		if len(filter.CreatorIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("creator_id = ANY($%d)", argIndex))
			args = append(args, pq.Array(filter.CreatorIDs))
			argIndex++
		}

		if len(filter.Statuses) > 0 {
			conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argIndex))
			args = append(args, pq.Array(filter.Statuses))
			argIndex++
		}

		if len(filter.Types) > 0 {
			conditions = append(conditions, fmt.Sprintf("type = ANY($%d)", argIndex))
			args = append(args, pq.Array(filter.Types))
			argIndex++
		}

		if filter.CreatedAfter != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
			args = append(args, *filter.CreatedAfter)
			argIndex++
		}

		if filter.CreatedBefore != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
			args = append(args, *filter.CreatedBefore)
			argIndex++
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Add ordering
	if filter != nil && filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder == "desc" {
			order = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, order)
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add pagination
	if filter != nil {
		if filter.Limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, filter.Limit)
			argIndex++
		}

		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
			argIndex++
		}
	}

	return query, args
}

// scanContent scans a database row into a Content entity
func (r *PostgreSQLContentRepository) scanContent(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Content, error) {
	var content entities.Content
	var platforms, hashtags, mentions, tags, keywords pq.StringArray
	var customFields, metadata, externalIDs, targetAudience []byte

	err := scanner.Scan(
		&content.ID, &content.Title, &content.Body, &content.Type, &content.Format,
		&content.Status, &content.Priority, &content.Category, &content.BrandID,
		&content.CampaignID, &content.CreatorID, &content.ApproverID, &platforms,
		&hashtags, &mentions, &tags, &keywords, &content.Tone, &content.Language,
		&content.ScheduledAt, &content.PublishedAt, &content.ExpiresAt, &targetAudience,
		&customFields, &metadata, &externalIDs, &content.IsTemplate, &content.TemplateID,
		&content.AIGenerated, &content.AIPrompt, &content.IsArchived,
		&content.CreatedAt, &content.UpdatedAt, &content.CreatedBy, &content.UpdatedBy, &content.Version,
	)

	if err != nil {
		return nil, err
	}

	// Convert PostgreSQL arrays to Go slices
	content.Platforms = make([]entities.PlatformType, len(platforms))
	for i, p := range platforms {
		content.Platforms[i] = entities.PlatformType(p)
	}

	content.Hashtags = []string(hashtags)
	content.Mentions = []string(mentions)
	content.Tags = []string(tags)
	content.Keywords = []string(keywords)

	// Unmarshal JSON fields
	if len(customFields) > 0 {
		json.Unmarshal(customFields, &content.CustomFields)
	}
	if len(metadata) > 0 {
		json.Unmarshal(metadata, &content.Metadata)
	}
	if len(externalIDs) > 0 {
		json.Unmarshal(externalIDs, &content.ExternalIDs)
	}
	if len(targetAudience) > 0 {
		json.Unmarshal(targetAudience, &content.TargetAudience)
	}

	return &content, nil
}

// Cache operations

func (r *PostgreSQLContentRepository) cacheContent(ctx context.Context, content *entities.Content) {
	if r.cache == nil {
		return
	}

	key := fmt.Sprintf("content:%s", content.ID)
	data, err := json.Marshal(content)
	if err != nil {
		return
	}

	r.cache.Set(ctx, key, data, 15*time.Minute)
}

func (r *PostgreSQLContentRepository) getCachedContent(ctx context.Context, id uuid.UUID) *entities.Content {
	if r.cache == nil {
		return nil
	}

	key := fmt.Sprintf("content:%s", id)
	data, err := r.cache.Get(ctx, key).Result()
	if err != nil {
		return nil
	}

	var content entities.Content
	if err := json.Unmarshal([]byte(data), &content); err != nil {
		return nil
	}

	return &content
}

func (r *PostgreSQLContentRepository) removeCachedContent(ctx context.Context, id uuid.UUID) {
	if r.cache == nil {
		return
	}

	key := fmt.Sprintf("content:%s", id)
	r.cache.Del(ctx, key)
}
