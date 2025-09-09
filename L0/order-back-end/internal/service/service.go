package order

import (
	"context"
	"fmt"
	"order-back-end/internal/cache"
	"order-back-end/internal/model"
	order "order-back-end/internal/repository"
)

// OrderService часть слоистой архитектуры
type OrderService struct {
	repository order.Repo
}

// NewOrderService создаем экземпляр класса
func NewOrderService(repository order.Repo) *OrderService {
	return &OrderService{
		repository: repository,
	}
}

func (s *OrderService) GetOrderFromDB(ctx context.Context, orderID string, cache cache.Cache) (*model.OrderInfo, error) {
	fmt.Println("GetOrderFromDB: ", orderID)
	if cachedOrder, ok := cache.Get(orderID); ok {
		return &cachedOrder, nil
	}

	dbOrder, err := s.repository.GetOrderFromDB(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("GetOrderFromDB: %w", err)
	}

	cache.Set(orderID, *dbOrder)
	return dbOrder, nil
}
