package models

import (
	"maxim/internal/utils"
	"time"
)

type User struct {
	Username      string
	Password      utils.HashSalt
	Gender        string
	FullName      string
	Location      string
	Email         string
	Profile       string
	Signature     string
	Buddies       map[string]bool
	Ignored       map[string]bool
	ActiveChannel *Channel
	DOB           time.Time
}
