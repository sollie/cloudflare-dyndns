package updater

import (
	"context"
	"time"

	"github.com/sollie/cloudflare-dyndns/cloudflare"
)

type DNSUpdater interface {
	UpdateSubdomain(zoneID, subdomain, domain, wanIP string) error
}

type Service struct {
	client  cloudflare.DNSClient
	timeout time.Duration
	ttl     int
}

func NewService(client cloudflare.DNSClient, timeout time.Duration, ttl int) DNSUpdater {
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
			record = newRecord
		} else {
			return err
		}
	}

	if record.Content == wanIP {
		return nil
	}

	err = s.client.UpdateRecord(ctx, zoneID, record.ID, recordName, wanIP)
	if err != nil {
		return err
	}

	return nil
}
