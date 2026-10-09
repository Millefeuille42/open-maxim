package models

import (
	"maps"
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
	Username string
	Password utils.HashSalt
	Details  UserDetails
	Buddies  map[string]bool
	Ignored  map[string]bool
}

func (user *User) Id() string {
	return user.Username
}

func (user *User) Clone() *User {
	userCopy := *user
	userCopy.Password.Hash = append([]byte(nil), user.Password.Hash...)
	userCopy.Password.Salt = append([]byte(nil), user.Password.Salt...)
	userCopy.Buddies = maps.Clone(user.Buddies)
	userCopy.Ignored = maps.Clone(user.Ignored)
	if userCopy.Buddies == nil {
		userCopy.Buddies = make(map[string]bool)
	}
	if userCopy.Ignored == nil {
		userCopy.Ignored = make(map[string]bool)
	}
	return &userCopy
}
