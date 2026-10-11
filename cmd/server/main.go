package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"maxim/internal/config"
	"maxim/internal/httpServer"
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
	config.CONFIG = config.GetConfig()

	dev := flag.Bool("dev", false, "create dummy and test accounts with password 'password'")
	flag.Parse()

	s := server.NewServer(config.CONFIG.ChatPort)
	if *dev {
		if err := addTestUsers(s); err != nil {
			log.Fatal(fmt.Errorf("create development users: %w", err))
		}
	}

	baseCtx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	ctx, stop := signal.NotifyContext(baseCtx, syscall.SIGTERM, os.Interrupt)
	defer stop()

	go httpServer.RunHttpServer(ctx, cancel, config.CONFIG.HttpPort)
	err := s.Run(ctx)
	if err != nil {
		log.Println(err)
	}

	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
		log.Printf("Fatal shutdown: %v\n", cause)
		return
	}

	log.Println("Server shut down cleanly.")
}
