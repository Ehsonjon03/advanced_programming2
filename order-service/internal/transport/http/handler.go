package http

import (
	"net/http"
	"order-service/internal/domain"
	"order-service/internal/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

// GET /orders/:id
func (h *OrderHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.GetByID(id)
	// ПРОВЕРКА: если ошибка или заказ не найден (nil)
	if err != nil || order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var ord domain.Order
	if err := c.ShouldBindJSON(&ord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	res, err := h.uc.CreateOrder(ord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": res})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	minStr := c.Query("min_amount")
	maxStr := c.Query("max_amount")

	if minStr == "" || maxStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both min_amount and max_amount are required"})
		return
	}

	min, _ := strconv.ParseInt(minStr, 10, 64)
	max, _ := strconv.ParseInt(maxStr, 10, 64)

	orders, err := h.uc.GetFilteredOrders(min, max)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
