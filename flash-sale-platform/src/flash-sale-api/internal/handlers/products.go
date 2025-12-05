package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"flash-sale-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productRepo *repository.ProductRepository
	redisRepo   *repository.RedisRepository
}

func NewProductHandler(productRepo *repository.ProductRepository, redisRepo *repository.RedisRepository) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
		redisRepo:   redisRepo,
	}
}

// Create adds a new product
func (h *ProductHandler) Create(c *gin.Context) {
	var p repository.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	if err := h.productRepo.Create(ctx, &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	// Sync inventory to Redis
	h.redisRepo.SetInventory(ctx, p.ID, p.Stock)

	c.JSON(http.StatusCreated, p)
}

// GetByID retrieves a single product
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	p, err := h.productRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// GetAll retrieves products (optionally filtered by IDs)
func (h *ProductHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	idsStr := c.Query("ids")

	var products []repository.Product
	var err error

	if idsStr != "" {
		// Parse comma-separated IDs
		idStrs := strings.Split(idsStr, ",")
		ids := make([]int, 0, len(idStrs))
		for _, s := range idStrs {
			if id, e := strconv.Atoi(strings.TrimSpace(s)); e == nil {
				ids = append(ids, id)
			}
		}
		products, err = h.productRepo.GetByIDs(ctx, ids)
	} else {
		products, err = h.productRepo.GetAll(ctx, 100)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Update modifies an existing product
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var p repository.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p.ID = id
	ctx := c.Request.Context()

	if err := h.productRepo.Update(ctx, &p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	// Sync inventory to Redis
	h.redisRepo.SetInventory(ctx, id, p.Stock)

	c.Status(http.StatusNoContent)
}

// Delete removes a product
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	ctx := c.Request.Context()

	if err := h.productRepo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	// Remove from Redis
	h.redisRepo.DeleteInventory(ctx, id)

	c.Status(http.StatusNoContent)
}

// SyncInventory syncs all product inventory from DB to Redis
func (h *ProductHandler) SyncInventory(c *gin.Context) {
	ctx := c.Request.Context()

	stocks, err := h.productRepo.GetAllStocks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	count := 0
	for id, stock := range stocks {
		if err := h.redisRepo.SetInventory(ctx, id, stock); err == nil {
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{"synced": count})
}
