package store

import (
	"errors"
	"sync"
)

var ErrOutOfStock = errors.New("out of stock")
var ErrSKUNotFound = errors.New("sku not found")

// InventoryStore is an in-memory, thread-safe stock ledger keyed by SKU.
// Swapping this for a real database later only means writing a new type
// that satisfies the same method set — handlers never change.
type InventoryStore struct {
	mu    sync.Mutex
	stock map[string]int
}

func NewInventoryStore() *InventoryStore {
	return &InventoryStore{
		stock: map[string]int{
			"SKU-1001": 50,
			"SKU-1002": 20,
			"SKU-1003": 0,
		},
	}
}

func (s *InventoryStore) Get(sku string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	qty, ok := s.stock[sku]
	if !ok {
		return 0, ErrSKUNotFound
	}
	return qty, nil
}

// Reserve decrements stock for sku by qty if enough is available, returning
// the remaining quantity so callers can update a metric without a second
// locked lookup (which would otherwise race against a concurrent reservation).
func (s *InventoryStore) Reserve(sku string, qty int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.stock[sku]
	if !ok {
		return 0, ErrSKUNotFound
	}
	if current < qty {
		return current, ErrOutOfStock
	}
	s.stock[sku] = current - qty
	return s.stock[sku], nil
}

// All returns a snapshot of every SKU's current stock, used to seed the
// stock-level gauge at startup.
func (s *InventoryStore) All() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := make(map[string]int, len(s.stock))
	for sku, qty := range s.stock {
		snapshot[sku] = qty
	}
	return snapshot
}
