package dns

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

type WanIPResolver interface {
	GetWANIP() (string, error)
}

type Resolver struct {
	server string
	target string
}

func NewResolver(server, target string) WanIPResolver {
	return &Resolver{
		server: server,
		target: target,
	}
}

func (r *Resolver) GetWANIP() (string, error) {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, 1)
	m.Question[0] = dns.Question{Name: r.target + ".", Qtype: dns.TypeTXT, Qclass: dns.ClassCHAOS}

	c := dns.Client{}
	c.Net = "udp4"

	resp, _, err := c.Exchange(m, r.server+":53")
	if err != nil {
		return "", fmt.Errorf("DNS exchange failed: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("received nil DNS response")
	}

	if len(resp.Answer) == 0 {
		return "", fmt.Errorf("no DNS answer received")
	}

	txtRecord, ok := resp.Answer[0].(*dns.TXT)
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
