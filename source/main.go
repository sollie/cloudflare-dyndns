package main

import (
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
		slog.Error("Failed to get WAN IP: " + err.Error())
		os.Exit(1)
	}
	slog.Info("WAN IP: " + wanip)

	cfClient, err := cloudflare.NewClient(config.Token)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	updateService := updater.NewService(cfClient, config.Timeout, config.TTL)

	for _, domain := range config.Domains {
		zoneID, err := cfClient.GetZoneID(domain.Zone)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		for _, subdomain := range domain.Subdomains {
			err := updateService.UpdateSubdomain(zoneID, subdomain, domain.Zone, wanip)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
}
