package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"maxim/internal/models"
	"maxim/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func addTestUsers(s *server.Server) error {
	for _, username := range []string{"dummy", "test"} {
		hash, err := s.Hash.GenerateHash([]byte("password"))
		if err != nil {
			return err
		}
		err = s.Users.Create(&models.User{
			Username: username,
			Password: *hash,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	port := flag.Int("port", 2002, "TCP port")
	dev := flag.Bool("dev", false, "create dummy and test accounts with password 'password'")
	flag.Parse()
	if *port < 1 || *port > 65535 {
		log.Fatal("port must be between 1 and 65535")
	}
	s := server.NewServer(*port)
	if *dev {
		if err := addTestUsers(s); err != nil {
			log.Fatal(fmt.Errorf("create development users: %w", err))
		}
	}
	err := s.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
