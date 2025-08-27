package order

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"order-back-end/internal/cache"
	"order-back-end/internal/model"
	"order-back-end/internal/repository/mocks"
	"testing"
	"time"
)

func TestOrderService_GetOrderFromDB(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	repo := mock_order.NewMockRepo(ctr)

	ctx := context.Background()
	orderId := "123"
	cache := cache.NewCache(time.Second*10, 10)

	order := &model.OrderInfo{
		OrderUID: orderId,
	}

	repo.EXPECT().GetOrderFromDB(ctx, orderId).Return(order, nil).Times(1)

	service := NewOrderService(repo)
	order, err := service.GetOrderFromDB(ctx, orderId, cache)
	require.NoError(t, err)
	_ = order
}

func TestOrderService_GetOrderFromDB_InCache(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	repo := mock_order.NewMockRepo(ctr)

	ctx := context.Background()
	orderId := "123"
	cache := cache.NewCache(time.Second*10, 10)

	expectedOrder := &model.OrderInfo{OrderUID: orderId}

	// заранее кладём заказ в кэш
	cache.Set(orderId, *expectedOrder)

	// говорим, что репозиторий НЕ должен вызываться
	repo.EXPECT().GetOrderFromDB(gomock.Any(), gomock.Any()).Times(0)

	service := NewOrderService(repo)

	order, err := service.GetOrderFromDB(ctx, orderId, cache)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, expectedOrder.OrderUID, order.OrderUID, "order should come from cache")
}

func TestOrderService_GetOrderFromDB_Error(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()

	repo := mock_order.NewMockRepo(ctr)

	ctx := context.Background()
	orderId := "123"
	cache := cache.NewCache(time.Second*10, 10)

	repoErr := errors.New("db is down")

	repo.EXPECT().GetOrderFromDB(ctx, orderId).Return(nil, repoErr).Times(1)

	service := NewOrderService(repo)
	_, err := service.GetOrderFromDB(ctx, orderId, cache)
	require.Error(t, err)
	require.Equal(t, fmt.Errorf("GetOrderFromDB: %w", repoErr), err)
}
