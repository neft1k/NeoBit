package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type Config struct {
	Port string
}

func GetConfig() (Config, error) {
	_ = godotenv.Load()
	defaultPort := os.Getenv("PORT")
	port := pflag.StringP("port", "p", defaultPort, "server port")
	pflag.Parse()
	if *port == "" {
		return Config{}, fmt.Errorf("PORT is not defined")
	}
	return Config{
		Port: *port,
	}, nil
}
