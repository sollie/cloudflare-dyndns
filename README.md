# cloudflare-dyndns

*WIP*
Set the IP for a hostname to your current WAN ip on a domain hosted on Cloudflare DNS.

This progrem looks up "myip.opendns.com" on resolver1.opendns.com to get your WAN ip and uses this to update the configured hostnames using the Cloudflare API.

## Installation

Install using `go get github.com/sollie/cloudflare-dyndns`.

## Configuration

We look for `cloudflare-dyndns.yaml` in one of the following locations:
* /etc/cloudflare-dyndns/
* $HOME/.cloudflare-dyndns
* $PWD

Copy `cloudflare-dyndns.yaml.dist` to `cloudflare-dyndns.yaml` in your desired location and edit to your liking.

* auth-email
 * The email address you use to log in to Cloudflare.
* auth-key
 * When logged in to Cloudflare, select your profile in the top-right corner. Find the API keys. You want the Global API Key.
* zoneid
 * When logged in to Cloudflare, select your domain and go to Overview. It should be under "Domain Summary".
* hostnames
 * One or more hostnames in the domain referenced by zoneid to update.
