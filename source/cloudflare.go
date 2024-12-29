package main

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

func cloudflareInit() (*cloudflare.API, error) {
	api, err := cloudflare.NewWithAPIToken(config.Token)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func getZoneID(api *cloudflare.API, domain string) (string, error) {
	zones, err := api.ListZones(context.Background(), domain)
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

func getRecordID(ctx context.Context, api *cloudflare.API, zoneID string, recordName string) (string, error) {
	zone := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := api.ListDNSRecords(ctx, zone, cloudflare.ListDNSRecordsParams{Name: recordName})
	if err != nil {
		return "", err
	}

	if len(records) == 0 {
		return "", fmt.Errorf("record not found")
	}

	return records[0].ID, nil
}

func createDNSRecord(ctx context.Context, client *cloudflare.API, zoneID, recordType, hostname, content string, ttl int) (cloudflare.DNSRecord, error) {
	createParams := cloudflare.CreateDNSRecordParams{
		ID:      zoneID,
		Type:    recordType,
		Name:    hostname,
		Content: content,
		TTL:     ttl,
	}

	zone := cloudflare.ZoneIdentifier(zoneID)
	record, err := client.CreateDNSRecord(ctx, zone, createParams)
	if err != nil {
		return cloudflare.DNSRecord{}, fmt.Errorf("failed to create DNS record: %w", err)
	}

	return record, nil
}

func updateRecord(ctx context.Context, client *cloudflare.API, zoneID, recordID, hostname, content string) error {
	updateParams := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Name:    hostname,
		Content: content,
	}

	zone := cloudflare.ZoneIdentifier(zoneID)
	_, err := client.UpdateDNSRecord(ctx, zone, updateParams)
	if err != nil {
		return fmt.Errorf("failed to update DNS record: %w", err)
	}
	return nil
}
