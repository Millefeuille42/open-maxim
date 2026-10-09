package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"maxim/internal/models"
	"maxim/internal/server"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type config struct {
	chatPort   int
	httpPort   int
	chatAddr   string
	advertAddr string
}

var Config = config{}

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

func handleConfig(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	body := fmt.Sprintf(""+
		"chat_server=%s\n"+
		"chat_port=%d\n"+
		"advert_url=%s\n",
		Config.chatAddr, Config.chatPort, Config.advertAddr,
	)
	_, err := w.Write([]byte(body))
	if err != nil {
		fmt.Printf("error writing response: %s\n", err)
		return
	}
}

func runHttpServer(ctx context.Context, cancel context.CancelCauseFunc, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/maxim_settings_us.asp", handleConfig)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("[HTTP] Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			log.Printf("[HTTP] Graceful shutdown error: %v", err)
		}
	}()

	log.Println("[HTTP] Listening on port", port)
	err := s.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
		cancel(err)
	}
}

func validatePort(srcPort string) int {
	port, err := strconv.Atoi(srcPort)
	if err != nil {
		log.Fatalf("Invalid port number")
	}
	if port < 1 || port > 65535 {
		log.Fatal("port must be between 1 and 65535")
	}
	return port
}

func getConfig() config {
	chatPort := os.Getenv("MAXIM_CHAT_PORT")
	if chatPort == "" {
		chatPort = "2002"
	}

	httpPort := os.Getenv("MAXIM_HTTP_PORT")
	if httpPort == "" {
		httpPort = "80"
	}

	chatAddr := os.Getenv("MAXIM_CHAT_ADDR")
	if chatAddr == "" {
		chatAddr = "localhost"
	}

	advertAddr := os.Getenv("MAXIM_ADVERT_ADDR")
	if advertAddr == "" {
		advertAddr = "localhost"
	}

	return config{
		chatPort:   validatePort(chatPort),
		httpPort:   validatePort(httpPort),
		chatAddr:   chatAddr,
		advertAddr: advertAddr,
	}
}

func main() {
	Config = getConfig()

	dev := flag.Bool("dev", false, "create dummy and test accounts with password 'password'")
	flag.Parse()

	s := server.NewServer(Config.chatPort)
	if *dev {
		if err := addTestUsers(s); err != nil {
			log.Fatal(fmt.Errorf("create development users: %w", err))
		}
	}

	baseCtx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	ctx, stop := signal.NotifyContext(baseCtx, syscall.SIGTERM, os.Interrupt)
	defer stop()

	go runHttpServer(ctx, cancel, Config.httpPort)
	err := s.Run(ctx)
	if err != nil {
		log.Println(err)
	}

	if cause := context.Cause(ctx); cause != nil && !errors.Is(cause, context.Canceled) {
		log.Fatalf("Fatal shutdown: %v", cause)
	}

	log.Println("Server shut down cleanly.")
}
