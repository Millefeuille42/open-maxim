package state

type Engine[T any] interface {
	Add(key string, item T)
	Get(key string) (T, bool)
	GetAll() map[string]T
	Remove(key string)
}
