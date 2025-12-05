package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"orders/pkg/models"

	_ "github.com/go-sql-driver/mysql"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(dbUser, dbPassword, dbHost, dbName string) (*ProductRepository, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Retry connection
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Retrying database connection... (%v)", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect after retries: %v", err)
	}

	// Connection pool
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to database")

	return &ProductRepository{db: db}, nil
}

func (r *ProductRepository) Close() error {
	return r.db.Close()
}

// BeginTx starts a new transaction
func (r *ProductRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

// GetProductWithLock retrieves product with FOR UPDATE lock
func (r *ProductRepository) GetProductWithLock(tx *sql.Tx, productID int) (*models.Product, error) {
	var product models.Product
	err := tx.QueryRow(
		"SELECT id, price, stock FROM products WHERE id = ? FOR UPDATE",
		productID,
	).Scan(&product.ID, &product.Price, &product.Stock)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	} else if err != nil {
		return nil, err
	}

	return &product, nil
}

// GetProduct retrieves product without locking (for optimistic strategy)
func (r *ProductRepository) GetProduct(productID int) (*models.Product, error) {
	var product models.Product
	err := r.db.QueryRow(
		"SELECT id, price, stock FROM products WHERE id = ?",
		productID,
	).Scan(&product.ID, &product.Price, &product.Stock)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("product not found")
	} else if err != nil {
		return nil, err
	}

	return &product, nil
}

// DecrementStock decrements product stock within a transaction
func (r *ProductRepository) DecrementStock(tx *sql.Tx, productID, quantity int) error {
	_, err := tx.Exec(
		"UPDATE products SET stock = stock - ? WHERE id = ?",
		quantity, productID,
	)
	return err
}

// DecrementStockOptimistic uses compare-and-swap for optimistic concurrency
func (r *ProductRepository) DecrementStockOptimistic(productID, quantity, expectedStock int) (bool, error) {
	result, err := r.db.Exec(
		"UPDATE products SET stock = stock - ? WHERE id = ? AND stock = ?",
		quantity, productID, expectedStock,
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil // True if update succeeded
}
