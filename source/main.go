package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/miekg/dns"
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

	wanip, err := getIP("whoami.cloudflare", "1.1.1.1")
	if err != nil {
		slog.Error("Failed to get WAN IP: " + err.Error())
		os.Exit(1)
	}
	slog.Info("WAN IP: " + wanip)

	cfClient, err := cloudflareInit()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	for _, domain := range config.Domains {
		zoneID, err := getZoneID(cfClient, domain.Zone)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		for _, subdomain := range domain.Subdomains {
			timeout := time.Second * 5
			err := handleSubdomains(cfClient, zoneID, subdomain, domain.Zone, wanip, timeout)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
}

func handleSubdomains(api *cloudflare.API, zoneID string, subdomain string, domain string, wanip string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	recordName := subdomain + "." + domain
	record, err := getRecord(ctx, api, zoneID, recordName)
	if err != nil {
		if err.Error() == "record not found" {
			newRecord, err := createDNSRecord(ctx, api, zoneID, "A", recordName, wanip, 300)
			if err != nil {
				return err
			}
			slog.Debug("Created record " + recordName)
			record = newRecord
		} else {
			return err
		}
	}

	if record.Content == wanip {
		slog.Debug("Record " + recordName + " is up to date")
		return nil
	}

	err = updateRecord(ctx, api, zoneID, record.ID, recordName, wanip)
	if err != nil {
		return err
	}

	slog.Debug("Updated record " + recordName + " with IP " + wanip)
	return nil
}

func getIP(target string, server string) (string, error) {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: target + ".", Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}

	c := dns.Client{}
	c.Net = "udp4"

	r, _, err := c.Exchange(m, server+":53")
	if err != nil {
		return "", fmt.Errorf("DNS exchange failed: %w", err)
	}

	if r == nil {
		return "", fmt.Errorf("received nil DNS response")
	}

	if len(r.Answer) == 0 {
		return "", fmt.Errorf("no DNS answer received")
	}

	txtRecord, ok := r.Answer[0].(*dns.TXT)
	if !ok {
		return "", fmt.Errorf("DNS answer is not a TXT record")
	}

	if len(txtRecord.Txt) == 0 {
		return "", fmt.Errorf("TXT record is empty")
	}

	ip := txtRecord.Txt[0]

	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address format: %s", ip)
	}

	return ip, nil
}
