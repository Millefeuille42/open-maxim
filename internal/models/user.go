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
	ActiveChannel *Channel
	DOB           time.Time
}
