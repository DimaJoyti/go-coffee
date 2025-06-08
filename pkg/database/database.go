package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// Database represents a database connection
type Database struct {
	db     *sql.DB
	config *Config
}

// Config represents database configuration
type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// Repository interface for data access
type Repository interface {
	Create(ctx context.Context, entity interface{}) error
	GetByID(ctx context.Context, id string, entity interface{}) error
	Update(ctx context.Context, entity interface{}) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter interface{}, entities interface{}) error
}

// NewDatabase creates a new database connection
func NewDatabase(config *Config) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.MaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		db:     db,
		config: config,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// GetDB returns the underlying sql.DB
func (d *Database) GetDB() *sql.DB {
	return d.db
}

// Health checks database health
func (d *Database) Health(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// UserRepository handles user data operations
type UserRepository struct {
	db *Database
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := r.db.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.CreatedAt, user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1
	`
	
	user := &User{}
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
		FROM users WHERE email = $1
	`
	
	user := &User{}
	err := r.db.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

// UpdateUser updates a user
func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users 
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, updated_at = $6
		WHERE id = $1
	`
	
	_, err := r.db.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	return nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	
	_, err := r.db.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

// OrderRepository handles order data operations
type OrderRepository struct {
	db *Database
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *Database) *OrderRepository {
	return &OrderRepository{db: db}
}

// CreateOrder creates a new order
func (r *OrderRepository) CreateOrder(ctx context.Context, order *Order) error {
	query := `
		INSERT INTO orders (id, user_id, items, total_amount, status, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := r.db.db.ExecContext(ctx, query,
		order.ID, order.UserID, order.Items, order.TotalAmount, order.Status,
		order.PaymentMethod, order.CreatedAt, order.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	
	return nil
}

// GetOrderByID retrieves an order by ID
func (r *OrderRepository) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	query := `
		SELECT id, user_id, items, total_amount, status, payment_method, created_at, updated_at
		FROM orders WHERE id = $1
	`
	
	order := &Order{}
	err := r.db.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.Items, &order.TotalAmount, &order.Status,
		&order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	
	return order, nil
}

// UpdateOrderStatus updates an order status
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE orders SET status = $2, updated_at = $3 WHERE id = $1`
	
	_, err := r.db.db.ExecContext(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	
	return nil
}

// ListOrdersByUser retrieves orders for a user
func (r *OrderRepository) ListOrdersByUser(ctx context.Context, userID string) ([]*Order, error) {
	query := `
		SELECT id, user_id, items, total_amount, status, payment_method, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC
	`
	
	rows, err := r.db.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()
	
	var orders []*Order
	for rows.Next() {
		order := &Order{}
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Items, &order.TotalAmount, &order.Status,
			&order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}
	
	return orders, nil
}

// Data models

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Order represents an order in the system
type Order struct {
	ID            string    `json:"id" db:"id"`
	UserID        string    `json:"user_id" db:"user_id"`
	Items         string    `json:"items" db:"items"` // JSON string
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	Status        string    `json:"status" db:"status"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		Database:     "go_coffee",
		SSLMode:      "disable",
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  time.Hour,
	}
}
