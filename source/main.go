package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/miekg/dns"
)

var (
	version  = "GIT"
	cfClient *cloudflare.API
)

func main() {
	opts := &slog.HandlerOptions{
		AddSource: false,
		Level:     config.LogLevel,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	slog.Info(os.Args[0] + " version " + version)

	wanip := getIP("whoami.cloudflare", "1.1.1.1")
	slog.Info("WAN IP: " + wanip)

	cfClient, err := cloudflareInit()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	for _, domain := range config.Domains {
		zoneID, err := getZoneID(cfClient, domain.Tld)
		if err != nil {
			slog.Error(err.Error())
			continue
		}

		for _, subdomain := range domain.Subdomains {
			timeout := time.Second * 5
			err := handleSubdomains(cfClient, zoneID, subdomain, domain.Tld, wanip, timeout)
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
			newRecord, err := createDNSRecord(ctx, api, zoneID, "A", recordName, "127.0.0.1", 300)
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

func getIP(target string, server string) string {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: target + ".", Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}

	c := dns.Client{}
	c.Net = "udp4"

	r, _, err := c.Exchange(m, server+":53")
	if err != nil {
		slog.Error(err.Error())
	}
	if len(r.Answer) == 0 {
		slog.Warn("No answer")
	}

	Arec := r.Answer[0].(*dns.TXT)
	return string(Arec.Txt[0])
}
