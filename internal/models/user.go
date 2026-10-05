package models

import "time"

type User struct {
	Username      string
	Hash, Salt    []byte
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
