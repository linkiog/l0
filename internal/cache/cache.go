package cache

import (
	"fmt"
	"sync"

	"github.com/linkiog/lo/internal/models"
)

type Cache struct {
	mu    sync.RWMutex
	store map[string]*models.Order
}

func NewCache() *Cache {
	return &Cache{
		store: make(map[string]*models.Order),
	}
}

func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	o, ok := c.store[orderUID]
	if ok {
		fmt.Printf("Order %s found in cache\n", orderUID)
	} else {
		fmt.Printf("Order %s not found in cache\n", orderUID)
	}
	return o, ok
}

func (c *Cache) Set(order *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[order.OrderUID] = order
}
