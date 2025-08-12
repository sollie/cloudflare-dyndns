package cloudflare

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

type DNSClient interface {
	GetZoneID(domain string) (string, error)
	GetRecord(ctx context.Context, zoneID, recordName string) (cloudflare.DNSRecord, error)
	CreateRecord(ctx context.Context, zoneID, recordType, hostname, content string, ttl int) (cloudflare.DNSRecord, error)
	UpdateRecord(ctx context.Context, zoneID, recordID, hostname, content string) error
}

type Client struct {
	api *cloudflare.API
}

func NewClient(token string) (DNSClient, error) {
	api, err := cloudflare.NewWithAPIToken(token)
	if err != nil {
		return nil, err
	}

	return &Client{api: api}, nil
}

func (c *Client) GetZoneID(domain string) (string, error) {
	zones, err := c.api.ListZones(context.Background(), domain)
	if err != nil {
		return "", err
	}

	for _, zone := range zones {
		if zone.Name == domain {
			return zone.ID, nil
		}
	}

	return "", fmt.Errorf("zone not found: %s", domain)
}

func (c *Client) GetRecord(ctx context.Context, zoneID string, recordName string) (cloudflare.DNSRecord, error) {
	zone := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := c.api.ListDNSRecords(ctx, zone, cloudflare.ListDNSRecordsParams{Name: recordName})
	if err != nil {
		return cloudflare.DNSRecord{}, err
	}

	if len(records) == 0 {
		return cloudflare.DNSRecord{}, fmt.Errorf("record not found")
	}

	return records[0], nil
}

func (c *Client) CreateRecord(ctx context.Context, zoneID, recordType, hostname, content string, ttl int) (cloudflare.DNSRecord, error) {
	createParams := cloudflare.CreateDNSRecordParams{
		ID:      zoneID,
		Type:    recordType,
		Name:    hostname,
		Content: content,
		TTL:     ttl,
	}

	zone := cloudflare.ZoneIdentifier(zoneID)
	record, err := c.api.CreateDNSRecord(ctx, zone, createParams)
	if err != nil {
		return cloudflare.DNSRecord{}, fmt.Errorf("failed to create DNS record: %w", err)
	}

	return record, nil
}

func (c *Client) UpdateRecord(ctx context.Context, zoneID, recordID, hostname, content string) error {
	updateParams := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Name:    hostname,
		Content: content,
	}

	zone := cloudflare.ZoneIdentifier(zoneID)
	_, err := c.api.UpdateDNSRecord(ctx, zone, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}
	return nil
}
