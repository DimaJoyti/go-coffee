package marketplace

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/DimaJoyti/go-coffee/developer-dao/pkg/logger"
	"go.uber.org/zap"
)

// Repository interfaces

// SolutionRepository handles solution data operations
type SolutionRepository interface {
	Create(ctx context.Context, solution *Solution) error
	GetByID(ctx context.Context, solutionID uint64) (*Solution, error)
	List(ctx context.Context, filter *SolutionFilter) ([]*Solution, int, error)
	Update(ctx context.Context, solution *Solution) error
	Delete(ctx context.Context, solutionID uint64) error
	GetByDeveloper(ctx context.Context, developerAddress string, limit, offset int) ([]*Solution, error)
	GetByCategory(ctx context.Context, category SolutionCategory, limit, offset int) ([]*Solution, error)
	GetPopular(ctx context.Context, limit, offset int) ([]*Solution, error)
	GetTrending(ctx context.Context, limit, offset int) ([]*Solution, error)
	GetRecentlyUpdated(ctx context.Context, since time.Duration, limit, offset int) ([]*Solution, error)
	IncrementInstallCount(ctx context.Context, solutionID uint64) error
}

// ReviewRepository handles review data operations
type ReviewRepository interface {
	Create(ctx context.Context, review *Review) error
	GetByID(ctx context.Context, id int64) (*Review, error)
	GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Review, error)
	GetByReviewer(ctx context.Context, reviewerAddress string, limit, offset int) ([]*Review, error)
	HasReviewed(ctx context.Context, solutionID uint64, reviewerAddress string) (bool, error)
	Update(ctx context.Context, review *Review) error
	Delete(ctx context.Context, id int64) error
}

// InstallationRepository handles installation data operations
type InstallationRepository interface {
	Create(ctx context.Context, installation *Installation) error
	GetByID(ctx context.Context, id int64) (*Installation, error)
	GetBySolutionID(ctx context.Context, solutionID uint64) ([]*Installation, error)
	GetByInstaller(ctx context.Context, installerAddress string, limit, offset int) ([]*Installation, error)
	Update(ctx context.Context, installation *Installation) error
	Delete(ctx context.Context, id int64) error
}

// CategoryRepository handles category data operations
type CategoryRepository interface {
	GetAll(ctx context.Context) ([]*Category, error)
	GetByCategory(ctx context.Context, category SolutionCategory) (*Category, error)
	UpdateSolutionCount(ctx context.Context, category SolutionCategory, count int) error
}

// solutionRepository implements SolutionRepository
type solutionRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewSolutionRepository creates a new solution repository
func NewSolutionRepository(db *sql.DB, logger *logger.Logger) SolutionRepository {
	return &solutionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *solutionRepository) Create(ctx context.Context, solution *Solution) error {
	query := `
		INSERT INTO solutions (
			solution_id, name, description, category, version, developer_address,
			status, repository_url, documentation_url, demo_url, tags,
			created_at, updated_at, transaction_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		solution.SolutionID,
		solution.Name,
		solution.Description,
		int(solution.Category),
		solution.Version,
		solution.DeveloperAddress,
		int(solution.Status),
		solution.RepositoryURL,
		solution.DocumentationURL,
		solution.DemoURL,
		strings.Join(solution.Tags, ","),
		solution.CreatedAt,
		solution.UpdatedAt,
		solution.TransactionHash,
	).Scan(&solution.ID)

	if err != nil {
		r.logger.Error("Failed to create solution", zap.Error(err))
		return fmt.Errorf("failed to create solution: %w", err)
	}

	r.logger.Info("Solution created", zap.Uint64("solutionID", solution.SolutionID))
	return nil
}

func (r *solutionRepository) GetByID(ctx context.Context, solutionID uint64) (*Solution, error) {
	query := `
		SELECT id, solution_id, name, description, category, version, developer_address,
			   status, repository_url, documentation_url, demo_url, tags,
			   install_count, average_rating, review_count, created_at, updated_at,
			   approved_at, approver_address, transaction_hash, block_number
		FROM solutions 
		WHERE solution_id = $1`

	solution := &Solution{}
	var tagsStr string
	err := r.db.QueryRowContext(ctx, query, solutionID).Scan(
		&solution.ID,
		&solution.SolutionID,
		&solution.Name,
		&solution.Description,
		&solution.Category,
		&solution.Version,
		&solution.DeveloperAddress,
		&solution.Status,
		&solution.RepositoryURL,
		&solution.DocumentationURL,
		&solution.DemoURL,
		&tagsStr,
		&solution.InstallCount,
		&solution.AverageRating,
		&solution.ReviewCount,
		&solution.CreatedAt,
		&solution.UpdatedAt,
		&solution.ApprovedAt,
		&solution.ApproverAddress,
		&solution.TransactionHash,
		&solution.BlockNumber,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("solution not found: %d", solutionID)
		}
		r.logger.Error("Failed to get solution", zap.Error(err))
		return nil, fmt.Errorf("failed to get solution: %w", err)
	}

	// Parse tags
	if tagsStr != "" {
		solution.Tags = strings.Split(tagsStr, ",")
	}

	return solution, nil
}

func (r *solutionRepository) List(ctx context.Context, filter *SolutionFilter) ([]*Solution, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Build WHERE conditions
	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, int(*filter.Category))
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, int(*filter.Status))
		argIndex++
	}

	if filter.DeveloperAddress != "" {
		conditions = append(conditions, fmt.Sprintf("developer_address = $%d", argIndex))
		args = append(args, filter.DeveloperAddress)
		argIndex++
	}

	if filter.MinRating > 0 {
		conditions = append(conditions, fmt.Sprintf("average_rating >= $%d", argIndex))
		args = append(args, filter.MinRating)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM solutions %s", whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count solutions: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "DESC"
		if filter.SortOrder == "asc" {
			direction = "ASC"
		}
		orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
	}

	// Get solutions with pagination
	query := fmt.Sprintf(`
		SELECT id, solution_id, name, description, category, version, developer_address,
			   status, repository_url, documentation_url, demo_url, tags,
			   install_count, average_rating, review_count, created_at, updated_at,
			   approved_at, approver_address, transaction_hash, block_number
		FROM solutions %s
		%s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list solutions: %w", err)
	}
	defer rows.Close()

	var solutions []*Solution
	for rows.Next() {
		solution := &Solution{}
		var tagsStr string
		err := rows.Scan(
			&solution.ID,
			&solution.SolutionID,
			&solution.Name,
			&solution.Description,
			&solution.Category,
			&solution.Version,
			&solution.DeveloperAddress,
			&solution.Status,
			&solution.RepositoryURL,
			&solution.DocumentationURL,
			&solution.DemoURL,
			&tagsStr,
			&solution.InstallCount,
			&solution.AverageRating,
			&solution.ReviewCount,
			&solution.CreatedAt,
			&solution.UpdatedAt,
			&solution.ApprovedAt,
			&solution.ApproverAddress,
			&solution.TransactionHash,
			&solution.BlockNumber,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan solution: %w", err)
		}

		// Parse tags
		if tagsStr != "" {
			solution.Tags = strings.Split(tagsStr, ",")
		}

		solutions = append(solutions, solution)
	}

	return solutions, total, nil
}

func (r *solutionRepository) Update(ctx context.Context, solution *Solution) error {
	query := `
		UPDATE solutions 
		SET status = $1, approver_address = $2, updated_at = $3, approved_at = $4,
			block_number = $5, install_count = $6, average_rating = $7, review_count = $8,
			description = $9, repository_url = $10, documentation_url = $11, demo_url = $12,
			tags = $13
		WHERE solution_id = $14`

	solution.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		int(solution.Status),
		solution.ApproverAddress,
		solution.UpdatedAt,
		solution.ApprovedAt,
		solution.BlockNumber,
		solution.InstallCount,
		solution.AverageRating,
		solution.ReviewCount,
		solution.Description,
		solution.RepositoryURL,
		solution.DocumentationURL,
		solution.DemoURL,
		strings.Join(solution.Tags, ","),
		solution.SolutionID,
	)

	if err != nil {
		r.logger.Error("Failed to update solution", zap.Error(err))
		return fmt.Errorf("failed to update solution: %w", err)
	}

	return nil
}

func (r *solutionRepository) Delete(ctx context.Context, solutionID uint64) error {
	query := `DELETE FROM solutions WHERE solution_id = $1`

	result, err := r.db.ExecContext(ctx, query, solutionID)
	if err != nil {
		return fmt.Errorf("failed to delete solution: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("solution not found: %d", solutionID)
	}

	return nil
}

func (r *solutionRepository) GetByDeveloper(ctx context.Context, developerAddress string, limit, offset int) ([]*Solution, error) {
	// Implementation similar to List but with developer filter
	return []*Solution{}, nil // Simplified for now
}

func (r *solutionRepository) GetByCategory(ctx context.Context, category SolutionCategory, limit, offset int) ([]*Solution, error) {
	// Implementation similar to List but with category filter
	return []*Solution{}, nil // Simplified for now
}

func (r *solutionRepository) GetPopular(ctx context.Context, limit, offset int) ([]*Solution, error) {
	// Implementation would order by install_count DESC
	return []*Solution{}, nil // Simplified for now
}

func (r *solutionRepository) GetTrending(ctx context.Context, limit, offset int) ([]*Solution, error) {
	// Implementation would order by recent install growth
	return []*Solution{}, nil // Simplified for now
}

func (r *solutionRepository) GetRecentlyUpdated(ctx context.Context, since time.Duration, limit, offset int) ([]*Solution, error) {
	cutoff := time.Now().Add(-since)
	query := `
		SELECT id, solution_id, name, description, category, version, developer_address,
			   status, repository_url, documentation_url, demo_url, tags,
			   install_count, average_rating, review_count, created_at, updated_at,
			   approved_at, approver_address, transaction_hash, block_number
		FROM solutions 
		WHERE updated_at >= $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, cutoff, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently updated solutions: %w", err)
	}
	defer rows.Close()

	var solutions []*Solution
	for rows.Next() {
		solution := &Solution{}
		var tagsStr string
		err := rows.Scan(
			&solution.ID,
			&solution.SolutionID,
			&solution.Name,
			&solution.Description,
			&solution.Category,
			&solution.Version,
			&solution.DeveloperAddress,
			&solution.Status,
			&solution.RepositoryURL,
			&solution.DocumentationURL,
			&solution.DemoURL,
			&tagsStr,
			&solution.InstallCount,
			&solution.AverageRating,
			&solution.ReviewCount,
			&solution.CreatedAt,
			&solution.UpdatedAt,
			&solution.ApprovedAt,
			&solution.ApproverAddress,
			&solution.TransactionHash,
			&solution.BlockNumber,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan solution: %w", err)
		}

		// Parse tags
		if tagsStr != "" {
			solution.Tags = strings.Split(tagsStr, ",")
		}

		solutions = append(solutions, solution)
	}

	return solutions, nil
}

func (r *solutionRepository) IncrementInstallCount(ctx context.Context, solutionID uint64) error {
	query := `UPDATE solutions SET install_count = install_count + 1 WHERE solution_id = $1`

	_, err := r.db.ExecContext(ctx, query, solutionID)
	if err != nil {
		return fmt.Errorf("failed to increment install count: %w", err)
	}

	return nil
}
