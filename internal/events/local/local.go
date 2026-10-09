package local

import (
	"errors"
	"maxim/internal/events"
)

var ErrEngineFull = errors.New("local event engine is at capacity")

type Engine struct {
	messages chan events.ClientMessage
}

func NewLocalEngine(bufferSize int) *Engine {
	return &Engine{
		messages: make(chan events.ClientMessage, bufferSize),
	}
}

func (e *Engine) Send(msg events.ClientMessage) error {
	select {
	case e.messages <- msg:
		return nil
	default:
		return ErrEngineFull
	}
}

func (e *Engine) Receive() <-chan events.ClientMessage {
	return e.messages
}
