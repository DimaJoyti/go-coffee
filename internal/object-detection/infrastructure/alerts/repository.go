package alerts

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/DimaJoyti/go-coffee/internal/object-detection/domain"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// Repository implements the alert repository interface
type Repository struct {
	logger *zap.Logger
	db     *sql.DB
}

// NewRepository creates a new alert repository
func NewRepository(logger *zap.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger.With(zap.String("component", "alert_repository")),
		db:     db,
	}
}

// CreateAlert creates a new alert
func (r *Repository) CreateAlert(alert *domain.Alert) error {
	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var positionJSON, boundingBoxJSON []byte
	if alert.Position != nil {
		positionJSON, err = json.Marshal(alert.Position)
		if err != nil {
			return fmt.Errorf("failed to marshal position: %w", err)
		}
	}

	if alert.BoundingBox != nil {
		boundingBoxJSON, err = json.Marshal(alert.BoundingBox)
		if err != nil {
			return fmt.Errorf("failed to marshal bounding box: %w", err)
		}
	}

	query := `
		INSERT INTO alerts (
			id, type, severity, title, message, stream_id, zone_id, object_id, object_class,
			confidence, position, bounding_box, metadata, status, created_at, updated_at,
			resolved_at, resolved_by, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
	`

	_, err = r.db.Exec(query,
		alert.ID,
		string(alert.Type),
		string(alert.Severity),
		alert.Title,
		alert.Message,
		alert.StreamID,
		alert.ZoneID,
		alert.ObjectID,
		alert.ObjectClass,
		alert.Confidence,
		string(positionJSON),
		string(boundingBoxJSON),
		string(metadataJSON),
		string(alert.Status),
		alert.CreatedAt,
		alert.UpdatedAt,
		alert.ResolvedAt,
		alert.ResolvedBy,
		pq.Array(alert.Tags),
	)

	if err != nil {
		return fmt.Errorf("failed to insert alert: %w", err)
	}

	r.logger.Info("Alert created in database",
		zap.String("alert_id", alert.ID),
		zap.String("type", string(alert.Type)))

	return nil
}

// GetAlert retrieves an alert by ID
func (r *Repository) GetAlert(id string) (*domain.Alert, error) {
	query := `
		SELECT id, type, severity, title, message, stream_id, zone_id, object_id, object_class,
			   confidence, position, bounding_box, metadata, status, created_at, updated_at,
			   resolved_at, resolved_by, tags
		FROM alerts
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var alert domain.Alert
	var alertType, severity, status string
	var positionJSON, boundingBoxJSON, metadataJSON string
	var tags pq.StringArray

	err := row.Scan(
		&alert.ID,
		&alertType,
		&severity,
		&alert.Title,
		&alert.Message,
		&alert.StreamID,
		&alert.ZoneID,
		&alert.ObjectID,
		&alert.ObjectClass,
		&alert.Confidence,
		&positionJSON,
		&boundingBoxJSON,
		&metadataJSON,
		&status,
		&alert.CreatedAt,
		&alert.UpdatedAt,
		&alert.ResolvedAt,
		&alert.ResolvedBy,
		&tags,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan alert: %w", err)
	}

	// Parse enum fields
	alert.Type = domain.AlertType(alertType)
	alert.Severity = domain.AlertSeverity(severity)
	alert.Status = domain.AlertStatus(status)
	alert.Tags = []string(tags)

	// Parse JSON fields
	if positionJSON != "" {
		if err := json.Unmarshal([]byte(positionJSON), &alert.Position); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position: %w", err)
		}
	}

	if boundingBoxJSON != "" {
		if err := json.Unmarshal([]byte(boundingBoxJSON), &alert.BoundingBox); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bounding box: %w", err)
		}
	}

	if err := json.Unmarshal([]byte(metadataJSON), &alert.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &alert, nil
}

// GetAlerts retrieves alerts with filters
func (r *Repository) GetAlerts(filters domain.AlertFilters) ([]*domain.Alert, error) {
	query := `
		SELECT id, type, severity, title, message, stream_id, zone_id, object_id, object_class,
			   confidence, position, bounding_box, metadata, status, created_at, updated_at,
			   resolved_at, resolved_by, tags
		FROM alerts
		WHERE 1=1
	`

	var args []interface{}
	argIndex := 1

	// Build WHERE clause based on filters
	if filters.StreamID != "" {
		query += fmt.Sprintf(" AND stream_id = $%d", argIndex)
		args = append(args, filters.StreamID)
		argIndex++
	}

	if filters.ZoneID != "" {
		query += fmt.Sprintf(" AND zone_id = $%d", argIndex)
		args = append(args, filters.ZoneID)
		argIndex++
	}

	if filters.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, string(filters.Type))
		argIndex++
	}

	if filters.Severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argIndex)
		args = append(args, string(filters.Severity))
		argIndex++
	}

	if filters.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, string(filters.Status))
		argIndex++
	}

	if filters.ObjectClass != "" {
		query += fmt.Sprintf(" AND object_class = $%d", argIndex)
		args = append(args, filters.ObjectClass)
		argIndex++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.StartTime)
		argIndex++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.EndTime)
		argIndex++
	}

	if len(filters.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argIndex)
		args = append(args, pq.Array(filters.Tags))
		argIndex++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filters.Offset)
		argIndex++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.Alert

	for rows.Next() {
		var alert domain.Alert
		var alertType, severity, status string
		var positionJSON, boundingBoxJSON, metadataJSON string
		var tags pq.StringArray

		err := rows.Scan(
			&alert.ID,
			&alertType,
			&severity,
			&alert.Title,
			&alert.Message,
			&alert.StreamID,
			&alert.ZoneID,
			&alert.ObjectID,
			&alert.ObjectClass,
			&alert.Confidence,
			&positionJSON,
			&boundingBoxJSON,
			&metadataJSON,
			&status,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&alert.ResolvedAt,
			&alert.ResolvedBy,
			&tags,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		// Parse enum fields
		alert.Type = domain.AlertType(alertType)
		alert.Severity = domain.AlertSeverity(severity)
		alert.Status = domain.AlertStatus(status)
		alert.Tags = []string(tags)

		// Parse JSON fields
		if positionJSON != "" {
			if err := json.Unmarshal([]byte(positionJSON), &alert.Position); err != nil {
				return nil, fmt.Errorf("failed to unmarshal position: %w", err)
			}
		}

		if boundingBoxJSON != "" {
			if err := json.Unmarshal([]byte(boundingBoxJSON), &alert.BoundingBox); err != nil {
				return nil, fmt.Errorf("failed to unmarshal bounding box: %w", err)
			}
		}

		if err := json.Unmarshal([]byte(metadataJSON), &alert.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate alerts: %w", err)
	}

	return alerts, nil
}

// UpdateAlert updates an existing alert
func (r *Repository) UpdateAlert(alert *domain.Alert) error {
	metadataJSON, err := json.Marshal(alert.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var positionJSON, boundingBoxJSON []byte
	if alert.Position != nil {
		positionJSON, err = json.Marshal(alert.Position)
		if err != nil {
			return fmt.Errorf("failed to marshal position: %w", err)
		}
	}

	if alert.BoundingBox != nil {
		boundingBoxJSON, err = json.Marshal(alert.BoundingBox)
		if err != nil {
			return fmt.Errorf("failed to marshal bounding box: %w", err)
		}
	}

	query := `
		UPDATE alerts
		SET type = $2, severity = $3, title = $4, message = $5, stream_id = $6, zone_id = $7,
			object_id = $8, object_class = $9, confidence = $10, position = $11, bounding_box = $12,
			metadata = $13, status = $14, updated_at = $15, resolved_at = $16, resolved_by = $17, tags = $18
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		alert.ID,
		string(alert.Type),
		string(alert.Severity),
		alert.Title,
		alert.Message,
		alert.StreamID,
		alert.ZoneID,
		alert.ObjectID,
		alert.ObjectClass,
		alert.Confidence,
		string(positionJSON),
		string(boundingBoxJSON),
		string(metadataJSON),
		string(alert.Status),
		alert.UpdatedAt,
		alert.ResolvedAt,
		alert.ResolvedBy,
		pq.Array(alert.Tags),
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert not found: %s", alert.ID)
	}

	r.logger.Info("Alert updated in database", zap.String("alert_id", alert.ID))
	return nil
}

// DeleteAlert deletes an alert
func (r *Repository) DeleteAlert(id string) error {
	// Start transaction to delete alert and related data
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete notifications first (foreign key constraint)
	_, err = tx.Exec("DELETE FROM notifications WHERE alert_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete notifications: %w", err)
	}

	// Delete the alert
	result, err := tx.Exec("DELETE FROM alerts WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("alert not found: %s", id)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Alert deleted from database", zap.String("alert_id", id))
	return nil
}

// CreateAlertRule creates a new alert rule
func (r *Repository) CreateAlertRule(rule *domain.AlertRule) error {
	conditionsJSON, err := json.Marshal(rule.Conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal conditions: %w", err)
	}

	actionsJSON, err := json.Marshal(rule.Actions)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	query := `
		INSERT INTO alert_rules (
			id, name, description, stream_id, zone_id, type, severity, conditions, actions,
			cooldown, is_active, created_at, updated_at, last_triggered, trigger_count, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = r.db.Exec(query,
		rule.ID,
		rule.Name,
		rule.Description,
		rule.StreamID,
		rule.ZoneID,
		string(rule.Type),
		string(rule.Severity),
		string(conditionsJSON),
		string(actionsJSON),
		rule.Cooldown.Nanoseconds(),
		rule.IsActive,
		rule.CreatedAt,
		rule.UpdatedAt,
		rule.LastTriggered,
		rule.TriggerCount,
		pq.Array(rule.Tags),
	)

	if err != nil {
		return fmt.Errorf("failed to insert alert rule: %w", err)
	}

	r.logger.Info("Alert rule created in database",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name))

	return nil
}

// GetAlertRule retrieves an alert rule by ID
func (r *Repository) GetAlertRule(id string) (*domain.AlertRule, error) {
	query := `
		SELECT id, name, description, stream_id, zone_id, type, severity, conditions, actions,
			   cooldown, is_active, created_at, updated_at, last_triggered, trigger_count, tags
		FROM alert_rules
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)

	var rule domain.AlertRule
	var ruleType, severity string
	var conditionsJSON, actionsJSON string
	var cooldownNanos int64
	var tags pq.StringArray

	err := row.Scan(
		&rule.ID,
		&rule.Name,
		&rule.Description,
		&rule.StreamID,
		&rule.ZoneID,
		&ruleType,
		&severity,
		&conditionsJSON,
		&actionsJSON,
		&cooldownNanos,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
		&rule.LastTriggered,
		&rule.TriggerCount,
		&tags,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("alert rule not found: %s", id)
		}
		return nil, fmt.Errorf("failed to scan alert rule: %w", err)
	}

	// Parse enum fields
	rule.Type = domain.AlertType(ruleType)
	rule.Severity = domain.AlertSeverity(severity)
	rule.Tags = []string(tags)
	rule.Cooldown = time.Duration(cooldownNanos)

	// Parse JSON fields
	if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
	}

	if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}

	return &rule, nil
}

// GetAlertRules retrieves alert rules for a stream
func (r *Repository) GetAlertRules(streamID string) ([]*domain.AlertRule, error) {
	query := `
		SELECT id, name, description, stream_id, zone_id, type, severity, conditions, actions,
			   cooldown, is_active, created_at, updated_at, last_triggered, trigger_count, tags
		FROM alert_rules
		WHERE stream_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AlertRule

	for rows.Next() {
		var rule domain.AlertRule
		var ruleType, severity string
		var conditionsJSON, actionsJSON string
		var cooldownNanos int64
		var tags pq.StringArray

		err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Description,
			&rule.StreamID,
			&rule.ZoneID,
			&ruleType,
			&severity,
			&conditionsJSON,
			&actionsJSON,
			&cooldownNanos,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&rule.LastTriggered,
			&rule.TriggerCount,
			&tags,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}

		// Parse enum fields
		rule.Type = domain.AlertType(ruleType)
		rule.Severity = domain.AlertSeverity(severity)
		rule.Tags = []string(tags)
		rule.Cooldown = time.Duration(cooldownNanos)

		// Parse JSON fields
		if err := json.Unmarshal([]byte(conditionsJSON), &rule.Conditions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal conditions: %w", err)
		}

		if err := json.Unmarshal([]byte(actionsJSON), &rule.Actions); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}

		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate alert rules: %w", err)
	}

	return rules, nil
}