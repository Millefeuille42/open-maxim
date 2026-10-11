package config

import (
	"log"
	"os"
	"strconv"
)

// TODO use viper maybe someday

type AdvertConfig struct {
	TextPath  string
	ImagePath string
}

type Config struct {
	ChatPort   int
	HttpPort   int
	ChatAddr   string
	AdvertAddr string
	Banner     AdvertConfig
	Big        AdvertConfig
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

func GetConfig() Config {
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
		advertAddr = "http://localhost/advert?"
	}

	bannerTextPath := os.Getenv("MAXIM_BANNER_TEXT_PATH")
	if bannerTextPath == "" {
		bannerTextPath = "./assets/banner.txt"
	}

	bannerImagePath := os.Getenv("MAXIM_BANNER_IMAGE_PATH")
	if bannerImagePath == "" {
		bannerImagePath = "./assets/banner.png"
	}

	bigTextPath := os.Getenv("MAXIM_BIG_TEXT_PATH")
	if bigTextPath == "" {
		bigTextPath = "./assets/big.txt"
	}

	bigImagePath := os.Getenv("MAXIM_BIG_IMAGE_PATH")
	if bigImagePath == "" {
		bigImagePath = "./assets/big.png"
	}

	return Config{
		ChatPort:   validatePort(chatPort),
		HttpPort:   validatePort(httpPort),
		ChatAddr:   chatAddr,
		AdvertAddr: advertAddr,
		Banner: AdvertConfig{
			TextPath:  bannerTextPath,
			ImagePath: bannerImagePath,
		},
		Big: AdvertConfig{
			TextPath:  bigTextPath,
			ImagePath: bigImagePath,
		},
	}
}

var CONFIG = Config{}
