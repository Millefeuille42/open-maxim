package main

import (
	"log"
	"maxim/internal/server"
	"os"
)

// TODO: add context, cancellation etc

func main() {
	// TODO add argument parsing?
	salt := os.Getenv("MAXIM_SALT")
	if len(salt) == 0 {
		log.Fatal("MAXIM_SALT environment variable not set")
	}
	s := server.NewServer(2002, salt)
	s.Run()
}
