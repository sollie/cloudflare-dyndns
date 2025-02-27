# cloudflare-dyndns

Set the IP for a hostname to your current WAN IP on a domain hosted on Cloudflare DNS. Resolves whoami.cloudflare from 1.1.1.1 to get your current WAN IP.

## Features

- Automatically updates your DNS records with your current WAN IP.
- Supports multiple zones and subdomains.
- Configurable logging levels to suit your needs.

## Installation

Clone and build in `/source`, or download a docker image from [ghcr.io](https://github.com/sollie/cloudflare-dyndns/pkgs/container/cloudflare-dyndns), using:

```sh
docker pull ghcr.io/sollie/cloudflare-dyndns:latest
```

## Configuration

In v1, the configuration is done using environment variables.

### Required Environment Variables

- `CFDD_TOKEN`: Cloudflare API token with Edit permissions on the zone(s) you want to update.
- `CFDD_ZONE_1`: The zone you want to update.
- `CFDD_SUBDOMAINS_1`: The subdomains you want to update. Separate multiple subdomains with a comma.

### Optional Environment Variables

- `CFDD_LOGLEVEL`: The log level. Default is `Info`.
- `CFDD_ZONE_N`: Additional zones you want to update.
- `CFDD_SUBDOMAINS_N`: Additional subdomains you want to update. Separate multiple subdomains with a comma.

### Example Configuration

```sh
export CFDD_TOKEN="your_cloudflare_api_token"
export CFDD_ZONE_1="example.com"
export CFDD_SUBDOMAINS_1="www,api"
export CFDD_LOGLEVEL="Debug"
```

## Running the Application

To run the application, use the following command:

```sh
docker run -e CFDD_TOKEN -e CFDD_ZONE_1 -e CFDD_SUBDOMAINS_1 -e CFDD_LOGLEVEL ghcr.io/sollie/cloudflare-dyndns:latest
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](https://github.com/sollie/cloudflare-dyndns/blob/main/LICENSE) file for details.

## Contact

For any questions or suggestions, please open an issue.
