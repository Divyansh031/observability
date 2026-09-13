package store

import (
	"errors"
	"strconv"
	"sync"
)

var ErrOrderNotFound = errors.New("order not found")

type Order struct {
	ID            string  `json:"id"`
	SKU           string  `json:"sku"`
	Qty           int     `json:"qty"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id,omitempty"`
}

type OrderStore struct {
	mu     sync.Mutex
	orders map[string]Order
	nextID int
}

func NewOrderStore() *OrderStore {
	return &OrderStore{orders: make(map[string]Order)}
}

// Create assigns a new ID and stores the order.
func (s *OrderStore) Create(o Order) Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	o.ID = "order-" + strconv.Itoa(s.nextID)
	s.orders[o.ID] = o
	return o
}

// Update overwrites an existing order in place, keyed by its existing ID.
func (s *OrderStore) Update(o Order) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.orders[o.ID] = o
}

func (s *OrderStore) Get(id string) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	o, ok := s.orders[id]
	if !ok {
		return Order{}, ErrOrderNotFound
	}
	return o, nil
}
