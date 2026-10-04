package models

type Channel struct {
	Name    string
	Members map[string]bool
}

func NewChannel(name string) *Channel {
	return &Channel{
		Name:    name,
		Members: make(map[string]bool),
	}
}
