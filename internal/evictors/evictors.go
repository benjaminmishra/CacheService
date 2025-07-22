package evictors

// Evictor defines the interface for cache eviction policies.
type Evictor interface {
	OnSet(key string)

	OnGet(key string)

	OnDelete(key string)

	Evict(count int) []string
}
