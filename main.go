package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

// NSRecord Contains output from getIP
type NSRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// Errors Contains errors from CF response
type Errors struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	ErrorChain []struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"errorchain,omitempty"`
}

// ResultInfo Contains info from CF response
type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	Count      int `json:"count"`
	TotalCount int `json:"total_count"`
}

// Result Contains result in CF response
type Result struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Content    string    `json:"content"`
	Proxiable  bool      `json:"proxiable"`
	Proxied    bool      `json:"proxied"`
	TTL        int       `json:"ttl"`
	Locked     bool      `json:"locked"`
	ZoneID     string    `json:"zone_id"`
	ZoneName   string    `json:"zone_name"`
	ModifiedOn time.Time `json:"modified_on"`
	CreatedOn  time.Time `json:"created_on"`
	Meta       string    `json:"-,omitempty"`
}

// GetResponse Output from CF API lookup
type GetResponse struct {
	Result     []Result   `json:"result,omitempty"`
	ResultInfo ResultInfo `json:"result_info,omitempty"`
	Success    bool       `json:"success"`
	Errors     []Errors   `json:"errors,omitempty"`
	Messages   string     `json:"-,omitempty"`
}

// ChgResponse Output from CF API change
type ChgResponse struct {
	Result     Result     `json:"result,omitempty"`
	ResultInfo ResultInfo `json:"result_info,omitempty"`
	Success    bool       `json:"success"`
	Errors     []Errors   `json:"errors,omitempty"`
	Messages   string     `json:"-,omitempty"`
}

func main() {
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
	var wanip = getIP("myip.opendns.com", "resolver1.opendns.com")
	fmt.Printf("Our WAN IP is: %s\n", wanip)

	for _, v := range viper.Get("hostnames").([]interface{}) {
		host := v.(string)
		id := getID(host)

		if id.Success == true {
			fmt.Printf("Registered ip for %s is %s\n", host, id.Result[0].Content)

			if wanip != id.Result[0].Content {
				res := updateHost(host, wanip, id.Result[0].ID)
				if res.Success == true {
					fmt.Printf("%s updated to %s\n", res.Result.Name, res.Result.Content)
				} else {
					for _, e := range res.Errors {
						fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
					}
					os.Exit(1)
				}
			} else {
				fmt.Println("IP already up to date.")
				fmt.Println("Exiting...")
			}
		} else {
			for _, e := range id.Errors {
				fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
			}
			os.Exit(1)
		}
	}
}

func getID(hostname string) GetResponse {

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/?name=%s", viper.GetString("zoneid"), hostname)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("X-Auth-Email", viper.GetString("auth-email"))
	req.Header.Set("X-Auth-Key", viper.GetString("auth-key"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	result := GetResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	return result
}

func updateHost(hostname string, ip string, id string) ChgResponse {
	data := NSRecord{
		"A",
		hostname,
		ip,
		1,
		false,
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", viper.GetString("zoneid"), id)

	nsBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	reqBody := bytes.NewReader(nsBytes)

	req, err := http.NewRequest("PUT", url, reqBody)
	if err != nil {
		panic(err)
	}

	req.Header.Set("X-Auth-Email", viper.GetString("auth-email"))
	req.Header.Set("X-Auth-Key", viper.GetString("auth-key"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	result := ChgResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	return result
}

func getIP(target string, server string) string {
	c := dns.Client{}
	m := dns.Msg{}
	m.SetQuestion(target+".", dns.TypeA)
	r, _, err := c.Exchange(&m, server+":53")
	if err != nil {
		log.Fatal(err)
	}
	if len(r.Answer) == 0 {
		log.Fatal("No results")
	}

	Arec := r.Answer[0].(*dns.A)
	return fmt.Sprintf("%s", Arec.A)
}
