package main

import (
	"sync"
	"time"
)

type Item struct {
	Value      string
	Expiration int64
}

type Cache struct {
	items map[string]*Item
	mu    sync.RWMutex
}

func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]*Item),
	}

	go cache.StartEvictionTimer()

	return cache
}

func (c *Cache) Set(key string, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value:      value,
		Expiration: time.Now().Add(ttl).UnixNano(),
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found || item.Expiration < time.Now().UnixNano() {
		return "", false
	}

	return item.Value, true
}

func (c *Cache) StartEvictionTimer() {
	for {
		time.Sleep(time.Minute)
		c.evictItems()
	}
}

func (c *Cache) DeleteItem(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache) evictItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if time.Now().UnixNano() > item.Expiration {
			delete(c.items, key)
		}
	}
}
