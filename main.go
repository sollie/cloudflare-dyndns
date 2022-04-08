package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

var version = "GIT"

const apiEndpoint = "https://api.cloudflare.com/client/v4/zones"

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

	// Use OpenDNS to look up external IP
	wanip := getIP("whoami.cloudflare", "1.1.1.1")

	for domain, hosts := range viper.Get("zones").(map[string]interface{}) {
		zones := getZoneID(domain)
		if zones.Success {
			switch zones.ResultInfo.Count {
			case 1:
				zoneID := zones.Result[0].ID
				for _, v := range hosts.([]interface{}) {
					host := v.(string)
					id := getID(host, zoneID)

					if id.Success {
						doUpdate(host, wanip, id.Result[0].Content, zoneID, id.Result[0].ID)
					} else {
						for _, e := range id.Errors {
							fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
						}
					}
				}
			case 0:
				fmt.Printf("No data found for domain %s \n", domain)
			default:
				fmt.Printf("More than 1 domain returned when querying for %s\n", domain)
			}
		} else {
			for _, e := range zones.Errors {
				fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
			}
		}
	}
}

func doUpdate(host, wanip, curip, zoneID, recordID string) bool {
	if wanip != curip {
		res := updateHost(host, wanip, zoneID, recordID)
		if res.Success {
			fmt.Printf("%s changed from %s to %s\n", res.Result.Name, curip, res.Result.Content)
			return true
		}

		for _, e := range res.Errors {
			fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
		}
		return false
	}
	fmt.Printf("%s is up to date\n", host)
	return true
}

func getID(hostname, zoneid string) GetResponse {
	url := fmt.Sprintf(apiEndpoint+"/%s/dns_records/?name=%s", zoneid, hostname)

	body, err := doReq(url, "GET", nil)
	if err != nil {
		fmt.Println(err)
	}

	result := GetResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	return result
}

func getZoneID(domain string) ZneResponse {
	url := fmt.Sprintf(apiEndpoint+"?name=%s", domain)

	body, err := doReq(url, "GET", nil)
	if err != nil {
		fmt.Println(err)
	}

	result := ZneResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	return result
}

func updateHost(hostname, ip, zoneid, id string) ChgResponse {
	data := NSRecord{
		"A",
		hostname,
		ip,
		1,
		false,
	}

	url := fmt.Sprintf(apiEndpoint+"/%s/dns_records/%s", zoneid, id)

	nsBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	reqBody := bytes.NewReader(nsBytes)

	body, err := doReq(url, "PUT", reqBody)
	if err != nil {
		fmt.Println(err)
	}

	result := ChgResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	return result
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
		log.Fatal(err)
	}
	if len(r.Answer) == 0 {
		log.Fatal("No results")
	}

	Arec := r.Answer[0].(*dns.TXT)
	return string(Arec.Txt[0])
}

func doReq(url, method string, payload io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Email", viper.GetString("auth-email"))
	req.Header.Set("X-Auth-Key", viper.GetString("auth-key"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return body, nil
}
