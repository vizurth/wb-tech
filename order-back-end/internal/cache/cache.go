package cache

import (
	"order-back-end/internal/model"
	"sync"
)

type Cache struct {
	Mu     sync.RWMutex
	Orders map[string]model.OrderInfo
}

// NewCache создаем экземпляр Cache
func NewCache() *Cache {
	return &Cache{
		Orders: make(map[string]model.OrderInfo),
	}
}

// Set добавляем наш model.OrderInfo в Cache
func (c *Cache) Set(id string, o model.OrderInfo) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.Orders[id] = o
}

// Get получаем model.OrderInfo из Cache
func (c *Cache) Get(orderUID string) (model.OrderInfo, bool) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	o, ok := c.Orders[orderUID]
	return o, ok
}
