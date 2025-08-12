package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Token    string
	Domains  []Domain
	LogLevel slog.Level
}

type Domain struct {
	Zone       string
	Subdomains []string
}

var (
	config Config
)

func init() {
	config.Token = os.Getenv("CFDD_TOKEN")
	if config.Token == "" {
		slog.Error("CFDD_TOKEN is required")
		os.Exit(1)
	}

	logLevel := strings.ToLower(os.Getenv("CFDD_LOG_LEVEL"))
	switch logLevel {
	case "debug":
		config.LogLevel = slog.LevelDebug
	case "error":
		config.LogLevel = slog.LevelError
	case "info":
		config.LogLevel = slog.LevelInfo
	default:
		config.LogLevel = slog.LevelWarn
	}

	for i := 1; ; i++ {
		zoneKey := fmt.Sprintf("CFDD_ZONE_%d", i)
		subdomainsKey := fmt.Sprintf("CFDD_SUBDOMAINS_%d", i)

		zone := os.Getenv(zoneKey)
		subdomains := os.Getenv(subdomainsKey)

		if zone == "" || subdomains == "" {
			break
		}

		config.Domains = append(config.Domains, Domain{
			Zone:       zone,
			Subdomains: strings.Split(strings.ReplaceAll(subdomains, " ", ""), ","),
		})
	}
}
