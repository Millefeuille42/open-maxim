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
	for _, user := range []*models.User{&dummyUser, &testUser} {
		hash, err := s.Hash.GenerateHash([]byte("password"))
		if err != nil {
			log.Fatal(err)
		}
		user.Password = *hash
		s.Users.Add(user.Username, user)
	}
}

func main() {
	// TODO add argument parsing?
	s := server.NewServer(2002)
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
