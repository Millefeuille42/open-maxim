package state

import (
	"errors"
	"maxim/internal/models"
)

var (
	ErrAlreadyExists = errors.New("another entity with this id already exists")
	ErrNotFound      = errors.New("entity not found")
)

type Engine[T any] interface {
	Create(*T) error
	Get(string) (*T, bool)
	Update(*T) error
	List() ([]string, error)
}

type UserEngine interface {
	Engine[models.User]
}

type ChannelEngine interface {
	Engine[models.Channel]
}
