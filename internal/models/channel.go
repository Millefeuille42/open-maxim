package models

import "maps"

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

func (channel *Channel) Id() string {
	return channel.Name
}

func (channel *Channel) Clone() *Channel {
	channelCopy := *channel
	channelCopy.Members = maps.Clone(channel.Members)
	return &channelCopy
}
