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
	Tld        string
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

	logLevel := os.Getenv("CFDD_LOG_LEVEL")
	switch logLevel {
	case "Debug":
		config.LogLevel = slog.LevelDebug
	case "Error":
		config.LogLevel = slog.LevelError
	case "Info":
		config.LogLevel = slog.LevelInfo
	default:
		config.LogLevel = slog.LevelWarn
	}

	for i := 1; ; i++ {
		tldKey := fmt.Sprintf("CFDD_TLD_%d", i)
		subdomainsKey := fmt.Sprintf("CFDD_SUBDOMAINS_%d", i)

		tld := os.Getenv(tldKey)
		subdomains := os.Getenv(subdomainsKey)

		if tld == "" || subdomains == "" {
			break
		}

		config.Domains = append(config.Domains, Domain{
			Tld:        tld,
			Subdomains: strings.Split(strings.ReplaceAll(subdomains, " ", ""), ","),
		})
	}
}
