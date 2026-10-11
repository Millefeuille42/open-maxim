package httpServer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maxim/internal/config"
	"net/http"
	"os"
	"time"
)

type httpHandler func(http.ResponseWriter, *http.Request)

func handleServeFile(path string, fileType string) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fileType)

		//nolint:gosec // function should not be used against user variables
		data, err := os.ReadFile(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println(err)
		}
	}
}

func handleAdvert(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	body := fmt.Sprintf(""+
		"big_advert=%s\r\n"+
		"banner_advert=%s\r\n",
		"http://192.168.1.157/big.txt", "http://192.168.1.157/banner.txt",
	)
	_, err := w.Write([]byte(body))
	if err != nil {
		fmt.Printf("error writing response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleConfig(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	body := fmt.Sprintf(""+
		"chat_server=%s\r\n"+
		"chat_port=%d\r\n"+
		"advert_url=%s\r\n",
		config.CONFIG.ChatAddr, config.CONFIG.ChatPort, config.CONFIG.AdvertAddr,
	)
	_, err := w.Write([]byte(body))
	if err != nil {
		fmt.Printf("error writing response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func defaultHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:gosec // limited to url-encoded character
		log.Printf("[HTTP] %s %s", r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func RunHttpServer(ctx context.Context, cancel context.CancelCauseFunc, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)
	mux.HandleFunc("/maxim_settings_us.asp", handleConfig)
	mux.HandleFunc("/advert", handleAdvert)
	mux.HandleFunc("/banner.txt", handleServeFile(config.CONFIG.Banner.TextPath, "text/plain"))
	mux.HandleFunc("/banner.png", handleServeFile(config.CONFIG.Banner.ImagePath, "image/png"))
	mux.HandleFunc("/big.txt", handleServeFile(config.CONFIG.Big.TextPath, "text/plain"))
	mux.HandleFunc("/big.png", handleServeFile(config.CONFIG.Big.ImagePath, "image/png"))

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           loggerMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		log.Println("[HTTP] Shutting down server...")

		shutdownCtx, cancelHandle := context.WithTimeout(ctx, 5*time.Second)
		defer cancelHandle()

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
