# cloudflare-dyndns

*WIP*

Set the IP for a hostname to your current WAN ip on a domain hosted on Cloudflare DNS.

This program looks up "myip.opendns.com" on resolver1.opendns.com to get your WAN ip and uses this to update the configured hostnames using the Cloudflare API.

## Installation

Install using `go get github.com/sollie/cloudflare-dyndns`.

## Configuration

We look for `cloudflare-dyndns.yaml` in one of the following locations:
* /etc/cloudflare-dyndns/
* $HOME/.cloudflare-dyndns
* $PWD

Copy `cloudflare-dyndns.yaml.dist` to `cloudflare-dyndns.yaml` in your desired location and edit to your liking.

```yaml
auth-email: user@example.com
auth-key: Global API Key from CF profile
zones:
  domain.tld:
    - myhost.domain.tld
    - anotherhost.domain.tld
  otherdomain.tld:
    - site.otherdomain.tld
```
