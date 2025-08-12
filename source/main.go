package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/sollie/cloudflare-dyndns/cloudflare"
	"github.com/sollie/cloudflare-dyndns/dns"
	"github.com/sollie/cloudflare-dyndns/updater"
)

var (
	version = "GIT"
)

func main() {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     config.LogLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info(os.Args[0] + " version " + version)

	resolver := dns.NewResolver("1.1.1.1", "whoami.cloudflare")
	wanip, err := resolver.GetWANIP()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get WAN IP: %v", err))
		os.Exit(1)
	}
	slog.Info("WAN IP: " + wanip)

	cfClient, err := cloudflare.NewClient(config.Token)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to initialize Cloudflare client: %v", err))
		os.Exit(1)
	}

	updateService := updater.NewService(cfClient, config.Timeout, config.TTL)

	for _, domain := range config.Domains {
		zoneID, err := cfClient.GetZoneID(domain.Zone)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to get zone ID for %s: %v", domain.Zone, err))
			continue
		}

		for _, subdomain := range domain.Subdomains {
			recordName := subdomain + "." + domain.Zone
			err := updateService.UpdateSubdomain(zoneID, subdomain, domain.Zone, wanip)
			if err != nil {
				slog.Error(fmt.Sprintf("Failed to update record %s: %v", recordName, err))
			} else {
				slog.Debug(fmt.Sprintf("Successfully processed record %s", recordName))
			}
		}
	}
}
