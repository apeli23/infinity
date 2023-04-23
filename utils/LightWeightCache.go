package utils

import (
	"sync"
	"time"
)

var CacheInstance *Cache

type cacheEntry struct {
	value      interface{}
	expiration time.Time
}

type Cache struct {
	data sync.Map
}

// This method takes a key as input and returns the corresponding value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	entry, ok := c.data.Load(key)
// If the key does not exist or has expired, the method returns nil and false, respectively.
	if !ok {
		return nil, false
	}

	// Check if the entry has expired
	if entry.(cacheEntry).expiration.Before(time.Now()) {
		c.data.Delete(key)
		return nil, false
	}

	return entry.(cacheEntry).value, true
}

// This method adds a key-value pair to the cache with a given expiration time.
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	entry := cacheEntry{
		value:      value,//string
		//time.Duration value that represents how long the value should be stored in the cache.
		expiration: time.Now().Add(expiration),
	}
	c.data.Store(key, entry)
}

//This method runs in a separate goroutine and purges any entries from the cache whose expiration time has passed.
func (c *Cache) purgeExpiredEntries() {
	for {
		<-time.After(1 * time.Minute) // Adjust the interval as needed
		c.data.Range(func(key, value interface{}) bool {
			if value.(cacheEntry).expiration.Before(time.Now()) {
				c.data.Delete(key)
			}
			return true
		})
	}
}
//This function creates a new instance of the Cache struct and starts the goroutine that purges expired entries. It returns a pointer to the new Cache instance.
func NewCache() *Cache {
	cache := &Cache{}
	go cache.purgeExpiredEntries()
	return cache
}

// This function is called automatically by Go when the package is imported. It creates a global CacheInstance variable that is initialized with a new Cache instanc
func init() {
	CacheInstance = NewCache()
}