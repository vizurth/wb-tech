package cache

import (
	"github.com/stretchr/testify/require"
	"order-back-end/internal/model"
	"testing"
	"time"
)

func TestOrderCacheSetAndGet(t *testing.T) {
	c := NewCache(1*time.Second, 10)

	order := model.OrderInfo{OrderUID: "123"}

	c.Set("123", order)

	got, ok := c.Get("123")
	require.True(t, ok, "expected order to be in cache")
	require.Equal(t, "123", got.OrderUID, "expected order uid to be 123")
}

func TestOrderCacheSetExpired(t *testing.T) {
	c := NewCache(1*time.Second, 1)

	order1 := model.OrderInfo{OrderUID: "111"}
	order2 := model.OrderInfo{OrderUID: "222"}

	c.Set("111", order1)
	c.Set("222", order2)
	_, ok1 := c.Get("111")
	require.False(t, ok1, "expected order 111 to be expired")

	_, ok2 := c.Get("222")
	require.True(t, ok2, "expected order 222 to be in cache")
}

func TestOrderCacheDelete(t *testing.T) {
	c := NewCache(1*time.Second, 10)

	order := model.OrderInfo{OrderUID: "123"}
	c.Set("123", order)

	c.Delete("123")
	_, ok := c.Get("123")
	require.False(t, ok, "expected order to be in cache")
}

func TestOrderCacheCleanupExpired(t *testing.T) {
	// TTL очень маленький, чтобы тест был быстрым
	c := NewCache(1*time.Millisecond, 10)

	order := model.OrderInfo{OrderUID: "123"}
	c.Set("123", order)

	// проверяем, что элемент изначально есть
	got, ok := c.Get("123")
	require.True(t, ok)
	require.Equal(t, "123", got.OrderUID)

	// ждём немного больше TTL, чтобы элемент устарел
	time.Sleep(2 * time.Millisecond)

	// элемент должен быть удалён автоматически фоновым cleanup
	_, ok = c.Get("123")
	require.False(t, ok, "expected order to be expired and cleaned up")
}
