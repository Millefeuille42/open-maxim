package models

type Identifiable interface {
	Id() string
}

type Clonable[T any] interface {
	Identifiable
	Clone() T
}
