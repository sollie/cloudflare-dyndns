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
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
)

var (
	version = "GIT"
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

// ZneResponse Used to fetch zoneid. We only care about a couple of fields, so this stays ugly
type ZneResponse struct {
	Result []struct {
		ID                  string    `json:"id"`
		Name                string    `json:"name"`
		Status              string    `json:"status"`
		Paused              bool      `json:"-,omitempty"`
		Type                string    `json:"-,omitempty"`
		DevelopmentMode     int       `json:"-,omitempty"`
		NameServers         string    `json:"-,omitempty"`
		OriginalNameServers string    `json:"-,omitempty"`
		OriginalRegistrar   string    `json:"-,omitempty"`
		OriginalDnshost     string    `json:"-,omitempty"`
		ModifiedOn          time.Time `json:"modified_on"`
		CreatedOn           time.Time `json:"created_on"`
		ActivatedOn         time.Time `json:"activated_on"`
		Meta                string    `json:"-,omitempty"`
		Owner               string    `json:"-,omitempty"`
		Account             string    `json:"-,omitempty"`
		Permissions         string    `json:"-,omitempty"`
		Plan                string    `json:"-,omitempty"`
	} `json:"result"`
	ResultInfo ResultInfo `json:"result_info,omitempty"`
	Success    bool       `json:"success"`
	Errors     []Errors   `json:"errors,omitempty"`
	Messages   string     `json:"-,omitempty"`
}

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
	var wanip = getIP("myip.opendns.com", "resolver1.opendns.com")

	for domain, hosts := range viper.Get("zones").(map[string]interface{}) {
		var curip string

		zones := getZoneID(domain)
		if zones.Success == true {
			switch zones.ResultInfo.Count {
			case 1:
				zoneid := zones.Result[0].ID
				for _, v := range hosts.([]interface{}) {
					host := v.(string)
					id := getID(host, zoneid)

					if id.Success == true {
						curip = id.Result[0].Content

						if wanip != id.Result[0].Content {
							res := updateHost(host, wanip, zoneid, id.Result[0].ID)
							if res.Success == true {
								fmt.Printf("%s changed from %s to %s\n", res.Result.Name, curip, res.Result.Content)
							} else {
								for _, e := range res.Errors {
									fmt.Printf("Code: %d\nError: %s\n", e.Code, e.Message)
								}
							}
						} else {
							fmt.Printf("%s is up to date\n", host)
						}
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

func getID(hostname, zoneid string) GetResponse {

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/?name=%s", zoneid, hostname)

	body, err := doReq(url, "GET", nil)

	result := GetResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

	return result
}

func getZoneID(domain string) ZneResponse {

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones?name=%s", domain)

	body, err := doReq(url, "GET", nil)

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

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", zoneid, id)

	nsBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	reqBody := bytes.NewReader(nsBytes)

	body, err := doReq(url, "PUT", reqBody)

	result := ChgResponse{}

	if err = json.Unmarshal(body, &result); err != nil {
		panic(err)
	}

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

func doReq(url string, method string, payload io.Reader) ([]byte, error) {
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
