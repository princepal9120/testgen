package llm

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
)

// Cache provides semantic caching for LLM responses
type Cache struct {
	entries map[string]*cacheEntry
	maxSize int
	mu      sync.RWMutex
	hits    int
	misses  int
}

type cacheEntry struct {
	response *CompletionResponse
	key      string
}

// NewCache creates a new cache with the given maximum size
func NewCache(maxSize int) *Cache {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &Cache{
		entries: make(map[string]*cacheEntry),
		maxSize: maxSize,
	}
}

// GenerateKey creates a cache key from the request parameters.
func (c *Cache) GenerateKey(prompt string, systemRole string, model string) string {
	return c.GenerateKeyParts(prompt, systemRole, model)
}

// GenerateKeyParts creates a cache key from normalized string parts.
func (c *Cache) GenerateKeyParts(parts ...string) string {
	hasher := sha256.New()
	for idx, part := range parts {
		if idx > 0 {
			hasher.Write([]byte("|"))
		}
		hasher.Write([]byte(strings.TrimSpace(part)))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// Get retrieves a cached response
func (c *Cache) Get(key string) (*CompletionResponse, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if exists {
		c.hits++
		// Clone response to prevent mutation
		respCopy := *entry.response
		respCopy.Cached = true
		return &respCopy, true
	}

	c.misses++
	return nil, false
}

// Set stores a response in the cache
func (c *Cache) Set(key string, response *CompletionResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: if at capacity, remove oldest (first found)
	if len(c.entries) >= c.maxSize {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[key] = &cacheEntry{
		response: response,
		key:      key,
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
	c.hits = 0
	c.misses = 0
}

// Stats returns cache statistics
func (c *Cache) Stats() (size int, hits int, misses int, hitRate float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	size = len(c.entries)
	hits = c.hits
	misses = c.misses

	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return
}

// Counts returns raw cache hit/miss counters.
func (c *Cache) Counts() (hits int, misses int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// CachedProvider wraps a Provider with caching
type CachedProvider struct {
	provider Provider
	cache    *Cache
}

// NewCachedProvider creates a provider wrapper with caching
func NewCachedProvider(provider Provider, cache *Cache) *CachedProvider {
	if cache == nil {
		cache = NewCache(10000)
	}
	return &CachedProvider{
		provider: provider,
		cache:    cache,
	}
}

// GetCache returns the underlying cache
func (p *CachedProvider) GetCache() *Cache {
	return p.cache
}

// GetProvider returns the underlying provider
func (p *CachedProvider) GetProvider() Provider {
	return p.provider
}
