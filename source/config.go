package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Token    string
	Domains  []Domain
	LogLevel slog.Level
	Timeout  time.Duration
	TTL      int
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

	timeoutStr := os.Getenv("CFDD_TIMEOUT_SECONDS")
	if timeoutStr == "" {
		config.Timeout = 5 * time.Second
	} else {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err != nil {
			slog.Warn("Invalid CFDD_TIMEOUT_SECONDS, using default 5 seconds")
			config.Timeout = 5 * time.Second
		} else {
			config.Timeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	ttlStr := os.Getenv("CFDD_TTL")
	if ttlStr == "" {
		config.TTL = 300
	} else {
		if ttl, err := strconv.Atoi(ttlStr); err != nil {
			slog.Warn("Invalid CFDD_TTL, using default 300")
			config.TTL = 300
		} else {
			config.TTL = ttl
		}
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
