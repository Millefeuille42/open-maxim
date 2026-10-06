package models

import (
	"maxim/internal/utils"
	"time"
)

type UserDetails struct {
	Gender    string
	FullName  string
	Location  string
	Email     string
	Profile   string
	Signature string
	DOB       time.Time
}

type User struct {
	Username      string
	Password      utils.HashSalt
	Details       UserDetails
	Buddies       map[string]bool
	Ignored       map[string]bool
	ActiveChannel *Channel
}
