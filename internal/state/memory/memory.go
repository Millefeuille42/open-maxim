package memory

type Engine[T any] struct {
	items map[string]T
}

func NewEngine[T any]() *Engine[T] {
	return &Engine[T]{
		items: make(map[string]T),
	}
}

func (m *Engine[T]) Add(key string, item T) {
	m.items[key] = item
}

func (m *Engine[T]) Get(key string) (T, bool) {
	item, exists := m.items[key]
	return item, exists
}

func (m *Engine[T]) Remove(key string) {
	delete(m.items, key)
}
