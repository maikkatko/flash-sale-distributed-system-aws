package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Order struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ProductID  int       `json:"product_id"`
	Quantity   int       `json:"quantity"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type OrderHandler struct {
	db *sql.DB
}

func NewOrderHandler(db *sql.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

// GetByID retrieves an order by ID
func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	var order Order
	err := h.db.QueryRowContext(c.Request.Context(),
		"SELECT id, user_id, product_id, quantity, total_price, status, created_at FROM orders WHERE id = ?",
		id,
	).Scan(&order.ID, &order.UserID, &order.ProductID, &order.Quantity, &order.TotalPrice, &order.Status, &order.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetByUser retrieves all orders for a user
func (h *OrderHandler) GetByUser(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter required"})
		return
	}

	rows, err := h.db.QueryContext(c.Request.Context(),
		"SELECT id, user_id, product_id, quantity, total_price, status, created_at FROM orders WHERE user_id = ? ORDER BY created_at DESC LIMIT 100",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	orders := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.TotalPrice, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		orders = append(orders, o)
	}

	c.JSON(http.StatusOK, orders)
}
