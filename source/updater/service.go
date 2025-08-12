package updater

import (
	"context"
	"log/slog"
	"time"

	"github.com/sollie/cloudflare-dyndns/cloudflare"
)

type Service struct {
	client  *cloudflare.Client
	timeout time.Duration
	ttl     int
}

func NewService(client *cloudflare.Client, timeout time.Duration, ttl int) *Service {
	return &Service{
		client:  client,
		timeout: timeout,
		ttl:     ttl,
	}
}

func (s *Service) UpdateSubdomain(zoneID, subdomain, domain, wanIP string) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	recordName := subdomain + "." + domain
	record, err := s.client.GetRecord(ctx, zoneID, recordName)
	if err != nil {
		if err.Error() == "record not found" {
			newRecord, err := s.client.CreateRecord(ctx, zoneID, "A", recordName, wanIP, s.ttl)
			if err != nil {
				return err
			}
			slog.Debug("Created record " + recordName)
			record = newRecord
		} else {
			return err
		}
	}

	if record.Content == wanIP {
		slog.Debug("Record " + recordName + " is up to date")
		return nil
	}

	err = s.client.UpdateRecord(ctx, zoneID, record.ID, recordName, wanIP)
	if err != nil {
		return err
	}

	slog.Debug("Updated record " + recordName + " with IP " + wanIP)
	return nil
}
