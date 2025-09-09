package cache

import (
	"order-back-end/internal/model"
	"sync"
	"time"
)

type Cache interface {
	Set(id string, o model.OrderInfo)
	Get(orderUID string) (model.OrderInfo, bool)
	Delete(orderUID string)
}

type cacheItem struct {
	value      model.OrderInfo
	expiration int64
}

type OrderCache struct {
	mu      sync.RWMutex
	orders  map[string]cacheItem
	ttl     time.Duration
	maxSize int
}

var _ Cache = (*OrderCache)(nil) // На этапе компиляции будет проверка удовлетворяет ли OrderCache интерфейсу

// NewCache создаём кэш с TTL и ограничением по размеру
func NewCache(ttl time.Duration, maxSize int) Cache {
	c := &OrderCache{
		orders:  make(map[string]cacheItem),
		ttl:     ttl,
		maxSize: maxSize,
	}

	// запускаем фоновую очистку
	go c.сleanup()

	return c
}

// Set добавляет или обновляет заказ
func (c *OrderCache) Set(id string, o model.OrderInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// если кэш переполнен → удалим случайный элемент
	if c.maxSize > 0 && len(c.orders) >= c.maxSize {
		for k := range c.orders {
			delete(c.orders, k)
			break
		}
	}

	c.orders[id] = cacheItem{
		value:      o,
		expiration: time.Now().Add(c.ttl).UnixNano(),
	}
}

// Get возвращает заказ, если он ещё валиден
func (c *OrderCache) Get(orderUID string) (model.OrderInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.orders[orderUID]
	if !ok || (item.expiration > 0 && time.Now().UnixNano() > item.expiration) {
		return model.OrderInfo{}, false
	}
	return item.value, true
}

// Delete удаляет заказ по ключу
func (c *OrderCache) Delete(orderUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.orders, orderUID)
}

func (c *OrderCache) cleanupExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, v := range c.orders {
		if now > v.expiration {
			delete(c.orders, k)
		}
	}
	c.mu.Unlock()
}

// Сleanup периодически удаляет устаревшие элементы
func (c *OrderCache) сleanup() {
	ticker := time.NewTicker(c.ttl)
	for range ticker.C {
		c.cleanupExpired()
	}
}
