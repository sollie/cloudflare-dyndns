package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

var version = "GIT"

func main() {
	fmt.Printf("%s version %s\n", os.Args[0], version)
	viper.SetConfigType("yaml")
	viper.SetConfigName("cloudflare-dyndns")
	viper.AddConfigPath("/etc/cloudflare-dyndns/")
	viper.AddConfigPath("$HOME/.cloudflare-dyndns")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Fatal error config file: %s \n", err)
		os.Exit(1)
	}

	wanip := getIP("whoami.cloudflare", "1.1.1.1")

	for domain, hosts := range viper.Get("zones").(map[string]interface{}) {
		processDomain(domain, hosts.([]interface{}))
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
		slog.Fatal(err)
	}
	if len(r.Answer) == 0 {
		slog.Fatal("No results")
	}

	Arec := r.Answer[0].(*dns.TXT)
	return string(Arec.Txt[0])
}
