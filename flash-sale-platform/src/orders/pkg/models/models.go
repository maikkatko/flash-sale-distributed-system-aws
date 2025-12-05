package models

// PurchaseRequest from client
type PurchaseRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	ProductID int    `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// Product represents product data
type Product struct {
	ID    int
	Price float64
	Stock int
}

// OrderMessage published to SQS
type OrderMessage struct {
	UserID     string  `json:"user_id"`
	ProductID  int     `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Timestamp  string  `json:"timestamp"`
}

// PurchaseResponse returned to client
type PurchaseResponse struct {
	UserID     string  `json:"user_id"`
	ProductID  int     `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Message    string  `json:"message"`
}
