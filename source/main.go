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
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			recordName := subdomain + "." + domain.Tld
			recordID, err := getRecordID(ctx, cfClient, zoneID, recordName)
			if err != nil {
				if err.Error() == "Record not found" {
					record, err := createDNSRecord(ctx, cfClient, zoneID, "A", recordName, "127.0.0.1", 300)
					if err != nil {
						slog.Error(err.Error())
						continue
					}
					recordID = record.ID
				} else {
					slog.Error(err.Error())
					continue
				}

				err = updateRecord(ctx, cfClient, zoneID, recordID, recordName, wanip)
				if err != nil {
					slog.Error(err.Error())
					continue
				}
			}
		}
	}
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
