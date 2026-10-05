package main

import (
	"context"
	"log"
	"maxim/internal/models"
	"maxim/internal/server"
	"os"
	"os/signal"
	"syscall"
)

var dummyUser = models.User{
	Username: "dummy",
	Buddies:  make(map[string]bool),
	Ignored:  make(map[string]bool),
}

var testUser = models.User{
	Username: "test",
	Buddies:  make(map[string]bool),
	Ignored:  make(map[string]bool),
}

func addTestUsers(s *server.Server) {
	hashSalt, err := s.Hash.GenerateHash([]byte("password"), []byte(s.Salt))
	if err != nil {
		log.Fatal(err)
	}

	dummyUser.Hash = hashSalt.Hash
	dummyUser.Salt = hashSalt.Salt
	testUser.Hash = hashSalt.Hash
	testUser.Salt = hashSalt.Salt

	s.Users.Add(dummyUser.Username, &dummyUser)
	s.Users.Add(testUser.Username, &testUser)
}

func main() {
	// TODO add argument parsing?
	salt := os.Getenv("MAXIM_SALT")
	if len(salt) == 0 {
		log.Fatal("MAXIM_SALT environment variable not set")
	}
	s := server.NewServer(2002, salt)
	ctx, cancel := context.WithCancel(context.Background())
	// TODO make this happen in dev mode only
	addTestUsers(s)

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sig
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel()
	}()

	s.Run(ctx)
}
