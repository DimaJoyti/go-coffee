package accounts

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// Repository defines the interface for account data operations
type Repository interface {
	// Account operations
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	GetAccountByUserID(ctx context.Context, userID string) (*Account, error)
	GetAccountByResetToken(ctx context.Context, token string) (*Account, error)
	UpdateAccount(ctx context.Context, account *Account) error
	DeleteAccount(ctx context.Context, id string) error
	ListAccounts(ctx context.Context, req *AccountListRequest) ([]Account, int, error)

	// Session operations
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	DeleteSession(ctx context.Context, id string) error
	DeleteExpiredSessions(ctx context.Context) error
	InvalidateAllSessions(ctx context.Context, accountID string) error

	// Security operations
	CreateSecurityEvent(ctx context.Context, event *SecurityEvent) error
	GetSecurityEvents(ctx context.Context, accountID string, limit int) ([]SecurityEvent, error)

	// KYC operations
	CreateKYCDocument(ctx context.Context, doc *KYCDocument) error
	GetKYCDocuments(ctx context.Context, accountID string) ([]KYCDocument, error)
	UpdateKYCDocument(ctx context.Context, doc *KYCDocument) error
	DeleteKYCDocument(ctx context.Context, id string) error
}

// PostgreSQLRepository implements Repository using PostgreSQL
type PostgreSQLRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(db *sqlx.DB) Repository {
	return &PostgreSQLRepository{db: db}
}

// CreateAccount creates a new account
func (r *PostgreSQLRepository) CreateAccount(ctx context.Context, account *Account) error {
	query := `
		INSERT INTO accounts (
			id, user_id, email, phone, first_name, last_name, date_of_birth,
			nationality, country, state, city, address, postal_code,
			account_type, account_status, kyc_status, kyc_level, risk_score,
			compliance_flags, two_factor_enabled, two_factor_method,
			account_limits, notification_preferences, metadata, created_at, updated_at
		) VALUES (
			:id, :user_id, :email, :phone, :first_name, :last_name, :date_of_birth,
			:nationality, :country, :state, :city, :address, :postal_code,
			:account_type, :account_status, :kyc_status, :kyc_level, :risk_score,
			:compliance_flags, :two_factor_enabled, :two_factor_method,
			:account_limits, :notification_preferences, :metadata, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccountByID retrieves an account by ID
func (r *PostgreSQLRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	var account Account
	query := `
		SELECT id, user_id, email, phone, first_name, last_name, date_of_birth,
			   nationality, country, state, city, address, postal_code,
			   account_type, account_status, kyc_status, kyc_level, risk_score,
			   compliance_flags, two_factor_enabled, two_factor_method,
			   last_login_at, last_login_ip, failed_login_count,
			   account_limits, notification_preferences, metadata,
			   created_at, updated_at, deleted_at
		FROM accounts
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetAccountByEmail retrieves an account by email
func (r *PostgreSQLRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	var account Account
	query := `
		SELECT id, user_id, email, phone, first_name, last_name, date_of_birth,
			   nationality, country, state, city, address, postal_code,
			   account_type, account_status, kyc_status, kyc_level, risk_score,
			   compliance_flags, two_factor_enabled, two_factor_method,
			   last_login_at, last_login_ip, failed_login_count,
			   account_limits, notification_preferences, metadata,
			   created_at, updated_at, deleted_at
		FROM accounts
		WHERE email = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetAccountByUserID retrieves an account by user ID
func (r *PostgreSQLRepository) GetAccountByUserID(ctx context.Context, userID string) (*Account, error) {
	var account Account
	query := `
		SELECT id, user_id, email, phone, first_name, last_name, date_of_birth,
			   nationality, country, state, city, address, postal_code,
			   account_type, account_status, kyc_status, kyc_level, risk_score,
			   compliance_flags, two_factor_enabled, two_factor_method,
			   last_login_at, last_login_ip, failed_login_count,
			   account_limits, notification_preferences, metadata,
			   created_at, updated_at, deleted_at
		FROM accounts
		WHERE user_id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetAccountByResetToken retrieves an account by password reset token
func (r *PostgreSQLRepository) GetAccountByResetToken(ctx context.Context, token string) (*Account, error) {
	var account Account
	query := `
		SELECT id, user_id, email, phone, first_name, last_name, date_of_birth,
			   nationality, country, state, city, address, postal_code,
			   account_type, account_status, kyc_status, kyc_level, risk_score,
			   compliance_flags, two_factor_enabled, two_factor_method,
			   last_login_at, last_login_ip, failed_login_count,
			   account_limits, notification_preferences, metadata,
			   created_at, updated_at, deleted_at
		FROM accounts
		WHERE metadata->>'password_reset_token' = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &account, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// UpdateAccount updates an existing account
func (r *PostgreSQLRepository) UpdateAccount(ctx context.Context, account *Account) error {
	account.UpdatedAt = time.Now()

	query := `
		UPDATE accounts SET
			email = :email, phone = :phone, first_name = :first_name, last_name = :last_name,
			date_of_birth = :date_of_birth, nationality = :nationality, country = :country,
			state = :state, city = :city, address = :address, postal_code = :postal_code,
			account_status = :account_status, kyc_status = :kyc_status, kyc_level = :kyc_level,
			risk_score = :risk_score, compliance_flags = :compliance_flags,
			two_factor_enabled = :two_factor_enabled, two_factor_method = :two_factor_method,
			last_login_at = :last_login_at, last_login_ip = :last_login_ip,
			failed_login_count = :failed_login_count, account_limits = :account_limits,
			notification_preferences = :notification_preferences, metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, account)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// DeleteAccount soft deletes an account
func (r *PostgreSQLRepository) DeleteAccount(ctx context.Context, id string) error {
	query := `UPDATE accounts SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

// ListAccounts retrieves a list of accounts with pagination and filtering
func (r *PostgreSQLRepository) ListAccounts(ctx context.Context, req *AccountListRequest) ([]Account, int, error) {
	var accounts []Account
	var total int

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if req.Status != "" {
		conditions = append(conditions, fmt.Sprintf("account_status = $%d", argIndex))
		args = append(args, req.Status)
		argIndex++
	}

	if req.KYCStatus != "" {
		conditions = append(conditions, fmt.Sprintf("kyc_status = $%d", argIndex))
		args = append(args, req.KYCStatus)
		argIndex++
	}

	if req.Country != "" {
		conditions = append(conditions, fmt.Sprintf("country = $%d", argIndex))
		args = append(args, req.Country)
		argIndex++
	}

	if req.SearchTerm != "" {
		searchCondition := fmt.Sprintf("(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex+1, argIndex+2)
		conditions = append(conditions, searchCondition)
		searchTerm := "%" + req.SearchTerm + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
		argIndex += 3
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM accounts WHERE %s", whereClause)
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count accounts: %w", err)
	}

	// Get paginated results
	offset := (req.Page - 1) * req.Limit
	query := fmt.Sprintf(`
		SELECT id, user_id, email, phone, first_name, last_name, date_of_birth,
			   nationality, country, state, city, address, postal_code,
			   account_type, account_status, kyc_status, kyc_level, risk_score,
			   compliance_flags, two_factor_enabled, two_factor_method,
			   last_login_at, last_login_ip, failed_login_count,
			   account_limits, notification_preferences, metadata,
			   created_at, updated_at, deleted_at
		FROM accounts
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	err = r.db.SelectContext(ctx, &accounts, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list accounts: %w", err)
	}

	return accounts, total, nil
}

// CreateSession creates a new session
func (r *PostgreSQLRepository) CreateSession(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (
			id, account_id, device_id, ip_address, user_agent, location,
			session_token, refresh_token, expires_at, is_active, metadata,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :device_id, :ip_address, :user_agent, :location,
			:session_token, :refresh_token, :expires_at, :is_active, :metadata,
			:created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByToken retrieves a session by token
func (r *PostgreSQLRepository) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	var session Session
	query := `
		SELECT id, account_id, device_id, ip_address, user_agent, location,
			   session_token, refresh_token, expires_at, is_active, metadata,
			   created_at, updated_at
		FROM sessions
		WHERE session_token = $1 AND is_active = true AND expires_at > NOW()`

	err := r.db.GetContext(ctx, &session, query, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
func (r *PostgreSQLRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	var session Session
	query := `
		SELECT id, account_id, device_id, ip_address, user_agent, location,
			   session_token, refresh_token, expires_at, is_active, metadata,
			   created_at, updated_at
		FROM sessions
		WHERE refresh_token = $1 AND is_active = true AND expires_at > NOW()`

	err := r.db.GetContext(ctx, &session, query, refreshToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (r *PostgreSQLRepository) UpdateSession(ctx context.Context, session *Session) error {
	session.UpdatedAt = time.Now()

	query := `
		UPDATE sessions SET
			ip_address = :ip_address, user_agent = :user_agent, location = :location,
			expires_at = :expires_at, is_active = :is_active, metadata = :metadata,
			updated_at = :updated_at
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session
func (r *PostgreSQLRepository) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteExpiredSessions deletes expired sessions
func (r *PostgreSQLRepository) DeleteExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW() OR is_active = false`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}

// InvalidateAllSessions invalidates all sessions for an account
func (r *PostgreSQLRepository) InvalidateAllSessions(ctx context.Context, accountID string) error {
	query := `UPDATE sessions SET is_active = false WHERE account_id = $1`

	_, err := r.db.ExecContext(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	return nil
}

// CreateSecurityEvent creates a new security event
func (r *PostgreSQLRepository) CreateSecurityEvent(ctx context.Context, event *SecurityEvent) error {
	query := `
		INSERT INTO security_events (
			id, account_id, event_type, severity, description, ip_address,
			user_agent, location, resolved, resolved_at, metadata, created_at
		) VALUES (
			:id, :account_id, :event_type, :severity, :description, :ip_address,
			:user_agent, :location, :resolved, :resolved_at, :metadata, :created_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, event)
	if err != nil {
		return fmt.Errorf("failed to create security event: %w", err)
	}

	return nil
}

// GetSecurityEvents retrieves security events for an account
func (r *PostgreSQLRepository) GetSecurityEvents(ctx context.Context, accountID string, limit int) ([]SecurityEvent, error) {
	var events []SecurityEvent
	query := `
		SELECT id, account_id, event_type, severity, description, ip_address,
			   user_agent, location, resolved, resolved_at, metadata, created_at
		FROM security_events
		WHERE account_id = $1
		ORDER BY created_at DESC
		LIMIT $2`

	err := r.db.SelectContext(ctx, &events, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %w", err)
	}

	return events, nil
}

// CreateKYCDocument creates a new KYC document
func (r *PostgreSQLRepository) CreateKYCDocument(ctx context.Context, doc *KYCDocument) error {
	query := `
		INSERT INTO kyc_documents (
			id, account_id, document_type, document_url, status,
			uploaded_at, verified_at, expires_at, metadata
		) VALUES (
			:id, :account_id, :document_type, :document_url, :status,
			:uploaded_at, :verified_at, :expires_at, :metadata
		)`

	_, err := r.db.NamedExecContext(ctx, query, doc)
	if err != nil {
		return fmt.Errorf("failed to create KYC document: %w", err)
	}

	return nil
}

// GetKYCDocuments retrieves KYC documents for an account
func (r *PostgreSQLRepository) GetKYCDocuments(ctx context.Context, accountID string) ([]KYCDocument, error) {
	var docs []KYCDocument
	query := `
		SELECT id, account_id, document_type, document_url, status,
			   uploaded_at, verified_at, expires_at, metadata
		FROM kyc_documents
		WHERE account_id = $1
		ORDER BY uploaded_at DESC`

	err := r.db.SelectContext(ctx, &docs, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get KYC documents: %w", err)
	}

	return docs, nil
}

// UpdateKYCDocument updates an existing KYC document
func (r *PostgreSQLRepository) UpdateKYCDocument(ctx context.Context, doc *KYCDocument) error {
	query := `
		UPDATE kyc_documents SET
			status = :status, verified_at = :verified_at, expires_at = :expires_at,
			metadata = :metadata
		WHERE id = :id`

	_, err := r.db.NamedExecContext(ctx, query, doc)
	if err != nil {
		return fmt.Errorf("failed to update KYC document: %w", err)
	}

	return nil
}

// DeleteKYCDocument deletes a KYC document
func (r *PostgreSQLRepository) DeleteKYCDocument(ctx context.Context, id string) error {
	query := `DELETE FROM kyc_documents WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete KYC document: %w", err)
	}

	return nil
}
