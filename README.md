# cloudflare-dyndns

Set the IP for a hostname to your current WAN ip on a domain hosted on Cloudflare DNS.
Resolves whoami.cloudflare from 1.1.1.1 to get your current WAN IP.

## Installation

Clone and build in `/source`, or download a docker image from [ghcr.io](https://github.com/sollie/cloudflare-dyndns/pkgs/container/cloudflare-dyndns),
using `docker pull ghcr.io/sollie/cloudflare-dyndns:latest`.

## Configuration

In v1 the configuration is done using environment variables.

### Required:
* CFDD_TOKEN: Cloudflare API token with Edit permissions on the zone(s) you want to update.
* CFDD_ZONE_1: The zone you want to update.
* CFDD_SUBDOMAINS_1: The subdomains you want to update. Separate multiple subdomains with a comma.

### Optional:
* CFDD_LOGLEVEL: The log level. Default is `Info`.
* CFDD_ZONE_N: Additional zones you want to update.
* CFDD_SUBDOMAINS_N: Additional subdomains you want to update. Separate multiple subdomains with a comma.
