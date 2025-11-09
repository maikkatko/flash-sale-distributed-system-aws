//go:build products

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// Product represents the main product entity.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// initDB connects to the database and creates the schema if it doesn't exist.
func initDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	// Retry connection to handle slow database startup
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Failed to ping database, retrying in 5s... (%v)", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	// Optimize connection pool for performance
	db.SetMaxOpenConns(100) // Max number of open connections to the database
	db.SetMaxIdleConns(10)  // Max number of connections in the idle connection pool
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to the database.")

	// Schema for products table
	createProductsTable := `
		CREATE TABLE IF NOT EXISTS products (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_name (name)
		) ENGINE=InnoDB;`

	if _, err := db.Exec(createProductsTable); err != nil {
		log.Fatalf("Failed to create products table: %v", err)
	}
	log.Println("Database 'products' table is ready.")
}

func main() {
	initDB()
	defer db.Close()

	router := gin.Default()
	router.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Define product routes
	router.POST("/products", createProduct)
	router.GET("/products", getProducts) // For bulk fetching by query ?ids=...
	router.GET("/products/:id", getProductByID)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port for products service
	}
	router.Run(":" + port)
}

// createProduct adds a new product to the database.
func createProduct(c *gin.Context) {
	var newProduct Product
	if err := c.ShouldBindJSON(&newProduct); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data: " + err.Error()})
		return
	}

	// Start a new transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	result, err := tx.Exec(
		"INSERT INTO products (name, description, price, stock) VALUES (?, ?, ?, ?)",
		newProduct.Name, newProduct.Description, newProduct.Price, newProduct.Stock,
	)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve last insert ID"})
		return
	}

	// Read the newly created product back from the database to get all generated fields
	var createdProduct Product
	err = tx.QueryRow("SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = ?", id).Scan(&createdProduct.ID, &createdProduct.Name, &createdProduct.Description, &createdProduct.Price, &createdProduct.Stock, &createdProduct.CreatedAt, &createdProduct.UpdatedAt)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created product"})
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, createdProduct)
}

// getProductByID retrieves a single product by its ID.
func getProductByID(c *gin.Context) {
	id := c.Param("id")
	var product Product

	err := db.QueryRow(
		"SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = ?",
		id,
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// getProducts retrieves multiple products by their IDs, passed as a query parameter.
// Example: GET /products?ids=1,2,3
func getProducts(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'ids' is required"})
		return
	}

	idSlice := strings.Split(idsStr, ",")
	args := make([]interface{}, len(idSlice))
	for i, id := range idSlice {
		args[i] = id
	}

	// Build the query with the correct number of placeholders
	query := "SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id IN (?" + strings.Repeat(",?", len(args)-1) + ")"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error fetching products"})
		return
	}
	defer rows.Close()

	products := []Product{}
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan product"})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, products)
}

// updateProduct modifies an existing product.
func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
		return
	}

	_, err := db.Exec(
		"UPDATE products SET name = ?, description = ?, price = ?, stock = ? WHERE id = ?",
		product.Name, product.Description, product.Price, product.Stock, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.Status(http.StatusNoContent)
}

// deleteProduct removes a product from the database.
func deleteProduct(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM products WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}
	c.Status(http.StatusNoContent)
}
