package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Order represents an order in the database
type Order struct {
	ID            string
	UserID        string
	ProductID     int
	Quantity      int
	TotalPrice    float64
	Status        string
	CorrelationID string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OrderRepository handles database operations for orders
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new order repository with PostgreSQL connection
func NewOrderRepository() (*OrderRepository, error) {
	host := os.Getenv("DB_HOST")
	port := getEnvOrDefault("DB_PORT", "3306")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Retry connection
	var pingErr error
	for i := 0; i < 10; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			break
		}
		log.Printf("Failed to ping database, retrying in 5s... (%v)", pingErr)
		time.Sleep(5 * time.Second)
	}
	if pingErr != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", pingErr)
	}

	log.Println("Order Repository: Successfully connected to MySQL")

	repo := &OrderRepository{db: db}

	// Initialize schema
	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initSchema creates tables if they don't exist
func (r *OrderRepository) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			initial_stock INT NOT NULL,
			current_stock INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB;

		CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			product_id INT,
			quantity INT NOT NULL DEFAULT 1,
			total_price DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			correlation_id VARCHAR(36),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products(id),
			INDEX idx_orders_user_id (user_id),
			INDEX idx_orders_product_id (product_id),
			INDEX idx_orders_status (status)
		) ENGINE=InnoDB;

		CREATE TABLE IF NOT EXISTS audit_log (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_type VARCHAR(50) NOT NULL,
			product_id INT,
			order_id VARCHAR(36),
			stock_before INT,
			stock_after INT,
			details JSON,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_audit_product_id (product_id)
		) ENGINE=InnoDB;
	`

	_, err := r.db.Exec(schema)
	if err != nil {
		return err
	}

	log.Println("Database schema initialized")
	return nil
}

// CreateOrder inserts a new order into the database
func (r *OrderRepository) CreateOrder(ctx context.Context, order Order) error {
	query := `
		INSERT INTO orders (id, user_id, product_id, quantity, total_price, status, correlation_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE id=id
	`

	_, err := r.db.ExecContext(ctx, query,
		order.ID,
		order.UserID,
		order.ProductID,
		order.Quantity,
		order.TotalPrice,
		order.Status,
		order.CorrelationID,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	log.Printf("Order %s created in database", order.ID)
	return nil
}

// DecrementStock reduces product stock in the database
func (r *OrderRepository) DecrementStock(ctx context.Context, productID int, quantity int) error {
	// Use a transaction with optimistic locking
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current stock for audit
	var stockBefore int
	err = tx.QueryRowContext(ctx,
		"SELECT current_stock FROM products WHERE id = ? FOR UPDATE",
		productID,
	).Scan(&stockBefore)
	if err != nil {
		return fmt.Errorf("failed to get current stock: %w", err)
	}

	// Decrement stock
	result, err := tx.ExecContext(ctx,
		"UPDATE products SET current_stock = current_stock - ? WHERE id = ? AND current_stock >= ?",
		quantity, productID, quantity,
	)
	if err != nil {
		return fmt.Errorf("failed to decrement stock: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient stock or product not found")
	}

	// Create audit log
	stockAfter := stockBefore - quantity
	_, err = tx.ExecContext(ctx,
		`INSERT INTO audit_log (event_type, product_id, stock_before, stock_after, details)
		 VALUES ('stock_decrement', ?, ?, ?, ?)`,
		productID, stockBefore, stockAfter,
		fmt.Sprintf(`{"quantity": %d}`, quantity),
	)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
		// Don't fail the transaction for audit log errors
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Stock decremented for product %d: %d -> %d", productID, stockBefore, stockAfter)
	return nil
}

// GetOrder retrieves an order by ID
func (r *OrderRepository) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	var order Order
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, product_id, quantity, total_price, status, correlation_id, created_at, updated_at
		 FROM orders WHERE id = ?`,
		orderID,
	).Scan(
		&order.ID, &order.UserID, &order.ProductID, &order.Quantity,
		&order.TotalPrice, &order.Status, &order.CorrelationID,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// UpdateOrderStatus updates the status of an order
func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?",
		status, orderID,
	)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

// Close closes the database connection
func (r *OrderRepository) Close() error {
	return r.db.Close()
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
