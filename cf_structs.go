package main

import (
	"time"
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
