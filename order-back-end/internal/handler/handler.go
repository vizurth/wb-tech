package order

import (
	"net/http"
	"order-back-end/internal/cache"
	order "order-back-end/internal/service"

	"github.com/gin-gonic/gin"
)

// OrderHandler структура содержащая router, cache, service - которые соответствует слоистой архитектуре
type OrderHandler struct {
	service *order.OrderService
	router  *gin.Engine
	cache   *cache.Cache
}

// NewHandler создает экземпляр OrderHandler
func NewHandler(service *order.OrderService, router *gin.Engine, cache *cache.Cache) *OrderHandler {
	return &OrderHandler{
		service: service,
		router:  router,
		cache:   cache,
	}
}

// GetOrder handler который реализует ручку GET /order/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	ctx := c.Request.Context()

	// парсим id
	orderIdStr := c.Param("id")

	// получаем orderInfo из service
	orderInfo, err := h.service.GetOrderFromDB(ctx, orderIdStr, h.cache)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orderInfo)
}

// RegisterRoutes регистрируем все ручки
func (h *OrderHandler) RegisterRoutes() {
	orderR := h.router.Group("/order")

	orderR.GET("/:id", h.GetOrder)
}
