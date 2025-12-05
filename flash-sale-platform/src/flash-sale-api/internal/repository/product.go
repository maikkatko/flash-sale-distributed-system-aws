package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, p *Product) error {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)",
		p.Name, p.Description, p.Price, p.Stock,
	)
	if err != nil {
		return fmt.Errorf("insert product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}
	p.ID = int(id)

	return r.db.QueryRowContext(ctx,
		"SELECT created_at, updated_at FROM products WHERE id = ?", id,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (*Product, error) {
	var p Product
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = ?",
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get product: %w", err)
	}
	return &p, nil
}

func (r *ProductRepository) GetByIDs(ctx context.Context, ids []int) ([]Product, error) {
	if len(ids) == 0 {
		return []Product{}, nil
	}

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	query := "SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id IN (?" +
		strings.Repeat(",?", len(ids)-1) + ")"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) GetAll(ctx context.Context, limit int) ([]Product, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, name, description, price, stock, created_at, updated_at FROM products LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, p *Product) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE products SET name = ?, description = ?, price = ?, stock = ? WHERE id = ?",
		p.Name, p.Description, p.Price, p.Stock, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetPrice(ctx context.Context, id int) (float64, error) {
	var price float64
	err := r.db.QueryRowContext(ctx, "SELECT price FROM products WHERE id = ?", id).Scan(&price)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("product not found")
	}
	if err != nil {
		return 0, fmt.Errorf("get price: %w", err)
	}
	return price, nil
}

func (r *ProductRepository) GetAllStocks(ctx context.Context) (map[int]int, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, stock FROM products")
	if err != nil {
		return nil, fmt.Errorf("get stocks: %w", err)
	}
	defer rows.Close()

	stocks := make(map[int]int)
	for rows.Next() {
		var id, stock int
		if err := rows.Scan(&id, &stock); err != nil {
			return nil, fmt.Errorf("scan stock: %w", err)
		}
		stocks[id] = stock
	}
	return stocks, nil
}
