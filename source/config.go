package main

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Token   string
	Domains []Domain
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
			Subdomains: strings.Split(subdomains, ","),
		})
	}
}