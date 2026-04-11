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

// GET /orders?min_amount=X&max_amount=Y
func (h *OrderHandler) GetOrders(c *gin.Context) {
	// 1. Получаем данные из Query параметров
	minStr := c.Query("min_amount")
	maxStr := c.Query("max_amount")

	// 2. ПРОВЕРКА: Если данные отсутствуют или пустые
	if minStr == "" || maxStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both min_amount and max_amount are required"})
		return
	}

	// Конвертируем строки в числа (int64, так как в domain у нас int64)
	min, errMin := strconv.ParseInt(minStr, 10, 64)
	max, errMax := strconv.ParseInt(maxStr, 10, 64)

	if errMin != nil || errMax != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameters must be valid integers"})
		return
	}

	// 3. ПРОВЕРКА: Если min меньше нуля
	if min < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount cannot be less than 0"})
		return
	}

	// 4. ПРОВЕРКА: Если max больше миллиона
	if max > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max_amount cannot exceed 1,000,000"})
		return
	}

	// Дополнительная логика: если мин больше макс
	if min > max {
		c.JSON(http.StatusBadRequest, gin.H{"error": "min_amount cannot be greater than max_amount"})
		return
	}

	// 5. Вызов UseCase, если все проверки пройдены
	orders, err := h.uc.GetFilteredOrders(min, max)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	// 6. Отдаем результат
	c.JSON(http.StatusOK, orders)
}
