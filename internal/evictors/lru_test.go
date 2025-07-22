package evictors

import "testing"

func TestLRUEvictor(t *testing.T) {
	lruEvictor := NewLRUEvictor()

	lruEvictor.OnSet("a")
	lruEvictor.OnSet("b")
	lruEvictor.OnGet("a")
	lruEvictor.OnSet("c")

	itemsToBeEvicted := lruEvictor.Evict(1)

	if itemsToBeEvicted == nil || len(itemsToBeEvicted) != 1 || itemsToBeEvicted[0] != "b" {
		t.Fatalf("expected b to be evicted, got %v", itemsToBeEvicted)
	}

	lruEvictor.OnDelete("a")

	if e := lruEvictor.Evict(1); len(e) == 0 || e[0] != "c" {
		t.Fatalf("expected c to be evicted")
	}
}

func TestLRUEvictorEmpty(t *testing.T) {
	lruEvictor := NewLRUEvictor()

	if itemsToEvict := lruEvictor.Evict(1); len(itemsToEvict) != 0 {
		t.Fatalf("expected no items to evict, got %d", len(itemsToEvict))
	}
}
