package main

import (
	"context"
	"log"
	"maxim/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// TODO add argument parsing?
	salt := os.Getenv("MAXIM_SALT")
	if len(salt) == 0 {
		log.Fatal("MAXIM_SALT environment variable not set")
	}
	s := server.NewServer(2002, salt)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sig
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel()
	}()

	s.Run(ctx)
}
