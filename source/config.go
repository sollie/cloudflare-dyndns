package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
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

	if err := validateConfig(); err != nil {
		slog.Error("Configuration validation failed: " + err.Error())
		os.Exit(1)
	}
}

func validateConfig() error {
	if len(config.Domains) == 0 {
		return fmt.Errorf("no domains configured - please set CFDD_ZONE_1 and CFDD_SUBDOMAINS_1")
	}

	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	subdomainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

	for i, domain := range config.Domains {
		if domain.Zone == "" {
			return fmt.Errorf("domain %d has empty zone name", i+1)
		}

		if !domainRegex.MatchString(domain.Zone) {
			return fmt.Errorf("domain %d has invalid zone name format: %s", i+1, domain.Zone)
		}

		if len(domain.Subdomains) == 0 {
			return fmt.Errorf("domain %d (%s) has no subdomains configured", i+1, domain.Zone)
		}

		for j, subdomain := range domain.Subdomains {
			if subdomain == "" {
				return fmt.Errorf("domain %d (%s) has empty subdomain at position %d", i+1, domain.Zone, j+1)
			}

			if !subdomainRegex.MatchString(subdomain) {
				return fmt.Errorf("domain %d (%s) has invalid subdomain format: %s", i+1, domain.Zone, subdomain)
			}
		}
	}

	return nil
}
