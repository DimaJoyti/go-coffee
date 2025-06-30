package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go-coffee-ai-agents/internal/common"
	"go-coffee-ai-agents/internal/observability"
)

// Beverage represents a beverage entity with database fields
type Beverage struct {
	ID           uuid.UUID                `json:"id"`
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Theme        string                   `json:"theme"`
	Ingredients  []common.BeverageIngredient `json:"ingredients"`
	Instructions []string                 `json:"instructions,omitempty"`
	PrepTime     int                      `json:"prep_time_minutes,omitempty"`
	Servings     int                      `json:"servings,omitempty"`
	Difficulty   string                   `json:"difficulty,omitempty"`
	Tags         []string                 `json:"tags,omitempty"`
	Nutrition    *common.NutritionInfo    `json:"nutrition,omitempty"`
	Cost         *common.CostEstimate     `json:"cost,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
	CreatedBy    string                   `json:"created_by"`
	Rating       float64                  `json:"rating"`
	ViewCount    int64                    `json:"view_count"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

// BeverageRepository defines the interface for beverage data operations
type BeverageRepository interface {
	Repository
	
	// Beverage-specific operations
	FindByTheme(ctx context.Context, theme string) ([]*Beverage, error)
	FindByCreator(ctx context.Context, createdBy string) ([]*Beverage, error)
	FindByIngredient(ctx context.Context, ingredient string) ([]*Beverage, error)
	FindRecent(ctx context.Context, limit int) ([]*Beverage, error)
	FindPopular(ctx context.Context, limit int) ([]*Beverage, error)
	UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	
	// Search operations
	Search(ctx context.Context, query string, filters BeverageFilters) ([]*Beverage, error)
	GetStatistics(ctx context.Context) (*BeverageStatistics, error)
}

// BeverageFilters defines filters for beverage search
type BeverageFilters struct {
	Theme         string    `json:"theme,omitempty"`
	CreatedBy     string    `json:"created_by,omitempty"`
	MinRating     *float64  `json:"min_rating,omitempty"`
	MaxCost       *float64  `json:"max_cost,omitempty"`
	MaxCalories   *int      `json:"max_calories,omitempty"`
	MaxPrepTime   *int      `json:"max_prep_time,omitempty"`
	AllergenFree  []string  `json:"allergen_free,omitempty"`
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	Limit         int       `json:"limit,omitempty"`
	Offset        int       `json:"offset,omitempty"`
}

// BeverageStatistics contains beverage statistics
type BeverageStatistics struct {
	TotalBeverages    int64            `json:"total_beverages"`
	TotalByTheme      map[string]int64 `json:"total_by_theme"`
	TotalByCreator    map[string]int64 `json:"total_by_creator"`
	AverageRating     float64          `json:"average_rating"`
	AverageCost       float64          `json:"average_cost"`
	AverageCalories   float64          `json:"average_calories"`
	AveragePrepTime   float64          `json:"average_prep_time"`
	PopularIngredients []string        `json:"popular_ingredients"`
	CreatedToday      int64            `json:"created_today"`
	CreatedThisWeek   int64            `json:"created_this_week"`
	CreatedThisMonth  int64            `json:"created_this_month"`
}

// PostgresBeverageRepository implements BeverageRepository for PostgreSQL
type PostgresBeverageRepository struct {
	*BaseRepository
}

// NewPostgresBeverageRepository creates a new PostgreSQL beverage repository
func NewPostgresBeverageRepository(
	db *DB,
	logger *observability.StructuredLogger,
	metrics *observability.MetricsCollector,
	tracing *observability.TracingHelper,
) BeverageRepository {
	baseRepo := NewBaseRepository(db, "beverages", logger, metrics, tracing)
	return &PostgresBeverageRepository{
		BaseRepository: baseRepo,
	}
}

// WithTx returns a new repository instance with transaction
func (r *PostgresBeverageRepository) WithTx(tx *Tx) Repository {
	baseRepo := r.BaseRepository.WithTx(tx).(*BaseRepository)
	return &PostgresBeverageRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new beverage
func (r *PostgresBeverageRepository) Create(ctx context.Context, entity interface{}) error {
	beverage, ok := entity.(*Beverage)
	if !ok {
		return fmt.Errorf("entity must be a *Beverage")
	}

	ctx, span := r.tracing.StartDatabaseSpan(ctx, "INSERT", "beverages")
	defer span.End()

	start := time.Now()

	// Serialize ingredients and metadata to JSON
	ingredientsJSON, err := json.Marshal(beverage.Ingredients)
	if err != nil {
		r.tracing.RecordError(span, err, "Failed to marshal ingredients")
		return fmt.Errorf("failed to marshal ingredients: %w", err)
	}

	metadataJSON, err := json.Marshal(beverage.Metadata)
	if err != nil {
		r.tracing.RecordError(span, err, "Failed to marshal metadata")
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO beverages (
			id, name, description, theme, ingredients, metadata, 
			created_by, rating, view_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	args := []interface{}{
		beverage.ID,
		beverage.Name,
		beverage.Description,
		beverage.Theme,
		ingredientsJSON,
		metadataJSON,
		beverage.CreatedBy,
		beverage.Rating,
		beverage.ViewCount,
		beverage.CreatedAt,
		beverage.UpdatedAt,
	}

	executor := r.getExecutor()
	_, err = executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to create beverage")
		r.logger.ErrorContext(ctx, "Failed to create beverage", err,
			"beverage_id", beverage.ID,
			"name", beverage.Name,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to create beverage: %w", err)
	}

	r.tracing.RecordSuccess(span, "Beverage created successfully")
	r.logger.InfoContext(ctx, "Beverage created successfully",
		"beverage_id", beverage.ID,
		"name", beverage.Name,
		"theme", beverage.Theme,
		"created_by", beverage.CreatedBy,
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetByID retrieves a beverage by its ID
func (r *PostgresBeverageRepository) GetByID(ctx context.Context, id interface{}, dest interface{}) error {
	beverageID, ok := id.(uuid.UUID)
	if !ok {
		return fmt.Errorf("id must be a uuid.UUID")
	}

	beverage, ok := dest.(*Beverage)
	if !ok {
		return fmt.Errorf("dest must be a *Beverage")
	}

	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages WHERE id = $1`

	executor := r.getExecutor()
	row := executor.QueryRowContext(ctx, query, beverageID)

	var ingredientsJSON, metadataJSON []byte
	err := row.Scan(
		&beverage.ID,
		&beverage.Name,
		&beverage.Description,
		&beverage.Theme,
		&ingredientsJSON,
		&metadataJSON,
		&beverage.CreatedBy,
		&beverage.Rating,
		&beverage.ViewCount,
		&beverage.CreatedAt,
		&beverage.UpdatedAt,
	)

	duration := time.Since(start)

	if err != nil {
		if err == sql.ErrNoRows {
			r.tracing.RecordError(span, err, "Beverage not found")
			r.logger.DebugContext(ctx, "Beverage not found",
				"beverage_id", beverageID,
				"duration_ms", duration.Milliseconds())
			return ErrNotFound
		}

		r.tracing.RecordError(span, err, "Failed to get beverage")
		r.logger.ErrorContext(ctx, "Failed to get beverage", err,
			"beverage_id", beverageID,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to get beverage: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(ingredientsJSON, &beverage.Ingredients); err != nil {
		r.tracing.RecordError(span, err, "Failed to unmarshal ingredients")
		return fmt.Errorf("failed to unmarshal ingredients: %w", err)
	}

	if err := json.Unmarshal(metadataJSON, &beverage.Metadata); err != nil {
		r.tracing.RecordError(span, err, "Failed to unmarshal metadata")
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	r.tracing.RecordSuccess(span, "Beverage retrieved successfully")
	r.logger.DebugContext(ctx, "Beverage retrieved successfully",
		"beverage_id", beverageID,
		"name", beverage.Name,
		"duration_ms", duration.Milliseconds())

	return nil
}

// FindByTheme finds beverages by theme
func (r *PostgresBeverageRepository) FindByTheme(ctx context.Context, theme string) ([]*Beverage, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages WHERE theme = $1 ORDER BY created_at DESC`

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, theme)
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to find beverages by theme")
		r.logger.ErrorContext(ctx, "Failed to find beverages by theme", err,
			"theme", theme,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to find beverages by theme: %w", err)
	}
	defer rows.Close()

	beverages, err := r.scanBeverages(rows)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan beverages")
		return nil, fmt.Errorf("failed to scan beverages: %w", err)
	}

	r.tracing.RecordSuccess(span, "Beverages found by theme")
	r.logger.DebugContext(ctx, "Beverages found by theme",
		"theme", theme,
		"count", len(beverages),
		"duration_ms", duration.Milliseconds())

	return beverages, nil
}

// FindByCreator finds beverages by creator
func (r *PostgresBeverageRepository) FindByCreator(ctx context.Context, createdBy string) ([]*Beverage, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages WHERE created_by = $1 ORDER BY created_at DESC`

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, createdBy)
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to find beverages by creator")
		r.logger.ErrorContext(ctx, "Failed to find beverages by creator", err,
			"created_by", createdBy,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to find beverages by creator: %w", err)
	}
	defer rows.Close()

	beverages, err := r.scanBeverages(rows)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan beverages")
		return nil, fmt.Errorf("failed to scan beverages: %w", err)
	}

	r.tracing.RecordSuccess(span, "Beverages found by creator")
	r.logger.DebugContext(ctx, "Beverages found by creator",
		"created_by", createdBy,
		"count", len(beverages),
		"duration_ms", duration.Milliseconds())

	return beverages, nil
}

// FindByIngredient finds beverages containing a specific ingredient
func (r *PostgresBeverageRepository) FindByIngredient(ctx context.Context, ingredient string) ([]*Beverage, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	// Use JSON operations to search within ingredients
	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages 
		WHERE ingredients::text ILIKE $1 
		ORDER BY created_at DESC`

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, "%"+ingredient+"%")
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to find beverages by ingredient")
		r.logger.ErrorContext(ctx, "Failed to find beverages by ingredient", err,
			"ingredient", ingredient,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to find beverages by ingredient: %w", err)
	}
	defer rows.Close()

	beverages, err := r.scanBeverages(rows)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan beverages")
		return nil, fmt.Errorf("failed to scan beverages: %w", err)
	}

	r.tracing.RecordSuccess(span, "Beverages found by ingredient")
	r.logger.DebugContext(ctx, "Beverages found by ingredient",
		"ingredient", ingredient,
		"count", len(beverages),
		"duration_ms", duration.Milliseconds())

	return beverages, nil
}

// FindRecent finds recently created beverages
func (r *PostgresBeverageRepository) FindRecent(ctx context.Context, limit int) ([]*Beverage, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages 
		ORDER BY created_at DESC 
		LIMIT $1`

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, limit)
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to find recent beverages")
		r.logger.ErrorContext(ctx, "Failed to find recent beverages", err,
			"limit", limit,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to find recent beverages: %w", err)
	}
	defer rows.Close()

	beverages, err := r.scanBeverages(rows)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan beverages")
		return nil, fmt.Errorf("failed to scan beverages: %w", err)
	}

	r.tracing.RecordSuccess(span, "Recent beverages found")
	r.logger.DebugContext(ctx, "Recent beverages found",
		"limit", limit,
		"count", len(beverages),
		"duration_ms", duration.Milliseconds())

	return beverages, nil
}

// FindPopular finds popular beverages by rating and view count
func (r *PostgresBeverageRepository) FindPopular(ctx context.Context, limit int) ([]*Beverage, error) {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "SELECT", "beverages")
	defer span.End()

	start := time.Now()

	query := `
		SELECT id, name, description, theme, ingredients, metadata, 
		       created_by, rating, view_count, created_at, updated_at
		FROM beverages 
		ORDER BY (rating * 0.7 + view_count * 0.3) DESC, created_at DESC
		LIMIT $1`

	executor := r.getExecutor()
	rows, err := executor.QueryContext(ctx, query, limit)
	if err != nil {
		duration := time.Since(start)
		r.tracing.RecordError(span, err, "Failed to find popular beverages")
		r.logger.ErrorContext(ctx, "Failed to find popular beverages", err,
			"limit", limit,
			"duration_ms", duration.Milliseconds())
		return nil, fmt.Errorf("failed to find popular beverages: %w", err)
	}
	defer rows.Close()

	beverages, err := r.scanBeverages(rows)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to scan beverages")
		return nil, fmt.Errorf("failed to scan beverages: %w", err)
	}

	r.tracing.RecordSuccess(span, "Popular beverages found")
	r.logger.DebugContext(ctx, "Popular beverages found",
		"limit", limit,
		"count", len(beverages),
		"duration_ms", duration.Milliseconds())

	return beverages, nil
}

// UpdateRating updates the rating of a beverage
func (r *PostgresBeverageRepository) UpdateRating(ctx context.Context, id uuid.UUID, rating float64) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "UPDATE", "beverages")
	defer span.End()

	start := time.Now()

	query := `UPDATE beverages SET rating = $1, updated_at = $2 WHERE id = $3`

	executor := r.getExecutor()
	result, err := executor.ExecContext(ctx, query, rating, time.Now(), id)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to update beverage rating")
		r.logger.ErrorContext(ctx, "Failed to update beverage rating", err,
			"beverage_id", id,
			"rating", rating,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to update beverage rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "Could not get rows affected", "error", err.Error())
	} else if rowsAffected == 0 {
		r.tracing.RecordError(span, ErrNotFound, "Beverage not found for rating update")
		return ErrNotFound
	}

	r.tracing.RecordSuccess(span, "Beverage rating updated")
	r.logger.DebugContext(ctx, "Beverage rating updated",
		"beverage_id", id,
		"rating", rating,
		"duration_ms", duration.Milliseconds())

	return nil
}

// IncrementViewCount increments the view count of a beverage
func (r *PostgresBeverageRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracing.StartDatabaseSpan(ctx, "UPDATE", "beverages")
	defer span.End()

	start := time.Now()

	query := `UPDATE beverages SET view_count = view_count + 1, updated_at = $1 WHERE id = $2`

	executor := r.getExecutor()
	result, err := executor.ExecContext(ctx, query, time.Now(), id)
	duration := time.Since(start)

	if err != nil {
		r.tracing.RecordError(span, err, "Failed to increment view count")
		r.logger.ErrorContext(ctx, "Failed to increment view count", err,
			"beverage_id", id,
			"duration_ms", duration.Milliseconds())
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "Could not get rows affected", "error", err.Error())
	} else if rowsAffected == 0 {
		r.tracing.RecordError(span, ErrNotFound, "Beverage not found for view count increment")
		return ErrNotFound
	}

	r.tracing.RecordSuccess(span, "View count incremented")
	r.logger.DebugContext(ctx, "View count incremented",
		"beverage_id", id,
		"duration_ms", duration.Milliseconds())

	return nil
}

// scanBeverages scans multiple beverage rows
func (r *PostgresBeverageRepository) scanBeverages(rows *sql.Rows) ([]*Beverage, error) {
	var beverages []*Beverage

	for rows.Next() {
		beverage := &Beverage{}
		var ingredientsJSON, metadataJSON []byte

		err := rows.Scan(
			&beverage.ID,
			&beverage.Name,
			&beverage.Description,
			&beverage.Theme,
			&ingredientsJSON,
			&metadataJSON,
			&beverage.CreatedBy,
			&beverage.Rating,
			&beverage.ViewCount,
			&beverage.CreatedAt,
			&beverage.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan beverage row: %w", err)
		}

		// Deserialize JSON fields
		if err := json.Unmarshal(ingredientsJSON, &beverage.Ingredients); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ingredients: %w", err)
		}

		if err := json.Unmarshal(metadataJSON, &beverage.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		beverages = append(beverages, beverage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return beverages, nil
}

// Search searches beverages with filters (placeholder implementation)
func (r *PostgresBeverageRepository) Search(ctx context.Context, query string, filters BeverageFilters) ([]*Beverage, error) {
	// This would implement full-text search with filters
	// For now, return a simple name-based search
	return r.FindByTheme(ctx, query)
}

// GetStatistics returns beverage statistics (placeholder implementation)
func (r *PostgresBeverageRepository) GetStatistics(ctx context.Context) (*BeverageStatistics, error) {
	// This would implement comprehensive statistics gathering
	// For now, return basic statistics
	count, err := r.Count(ctx, "", nil)
	if err != nil {
		return nil, err
	}

	return &BeverageStatistics{
		TotalBeverages: count,
	}, nil
}
