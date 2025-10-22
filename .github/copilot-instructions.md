# Copilot Instructions for cloudflare-dyndns

## Project Overview

This is a **Cloudflare Dynamic DNS updater** written in Go that automatically updates DNS records with your current WAN IP address. The application queries Cloudflare's 1.1.1.1 DNS resolver to get the current WAN IP and then updates specified DNS records on domains hosted on Cloudflare.

**Repository Stats:**
- Language: Go (Go 1.23.0+, toolchain 1.24.1+)
- Size: Small (~400 lines of Go code across 5 files)
- Type: CLI application packaged as Docker container
- Runtime: Designed to run as a single-shot execution or in containers

## Project Structure

```
/
├── .github/
│   ├── dependabot.yaml           # Dependency updates for gomod, docker, github-actions
│   └── workflows/
│       └── docker-publish.yml    # CI/CD workflow for Docker image builds
├── source/                       # All Go source code
│   ├── main.go                   # Application entry point (61 lines)
│   ├── config.go                 # Configuration and environment variable handling (130 lines)
│   ├── go.mod                    # Go module definition
│   ├── go.sum                    # Go dependency checksums
│   ├── cloudflare/
│   │   └── client.go            # Cloudflare API client wrapper (91 lines)
│   ├── dns/
│   │   └── resolver.go          # DNS resolver for getting WAN IP (66 lines)
│   └── updater/
│       └── service.go           # DNS record update service (57 lines)
├── Dockerfile                    # Multi-stage Docker build
├── .gitignore                    # Git ignore rules (source/cloudflare-dyndns)
├── LICENSE                       # Apache License 2.0
└── README.md                     # Project documentation
```

## Architecture

The application follows a clean, modular architecture:

1. **main.go**: Entry point that orchestrates the flow
   - Initializes logging (JSON format to stdout)
   - Loads configuration from environment variables
   - Gets WAN IP via DNS query to 1.1.1.1
   - Iterates through configured zones/subdomains and updates them

2. **config.go**: Configuration management
   - Loads environment variables on init()
   - Validates domain and subdomain format with regex
   - Supports multiple zones (CFDD_ZONE_1, CFDD_ZONE_2, etc.)
   - Fails fast if required config is missing

3. **cloudflare/client.go**: Cloudflare API wrapper
   - Implements DNSClient interface
   - Methods: GetZoneID, GetRecord, CreateRecord, UpdateRecord
   - Uses official cloudflare-go SDK v0.116.0

4. **dns/resolver.go**: WAN IP resolution
   - Queries whoami.cloudflare via 1.1.1.1 DNS
   - Uses TXT record in CHAOS class
   - Validates returned IP format

5. **updater/service.go**: DNS update logic
   - Creates record if it doesn't exist
   - Updates only if IP changed (no-op if already correct)
   - Uses context with configurable timeout

## Building and Testing

### Prerequisites
- Go 1.23.0 or later (tested with 1.24.7)
- Docker (for container builds)

### Build Steps

**ALWAYS run builds from the `/source` directory.**

#### Standard Build
```bash
cd source
go build -o cloudflare-dyndns
```
Build time: ~5-10 seconds
Output: 12MB binary

#### Build with CGO Disabled (matching Dockerfile)
```bash
cd source
CGO_ENABLED=0 go build -o cloudflare-dyndns
```
This produces a statically-linked binary suitable for scratch containers.

#### Clean Build
```bash
cd source
rm -f cloudflare-dyndns
go clean
go build -o cloudflare-dyndns
```

### Dependency Management

**Download dependencies:**
```bash
cd source
go mod download
```
Time: ~5-10 seconds (first time)

**Tidy dependencies:**
```bash
cd source
go mod tidy
```
Note: This may add indirect dependencies like testify, go-cmp, yaml.v3 to go.sum.

### Testing

**Run tests:**
```bash
cd source
go test ./...
```
Current status: No test files exist. Output will be:
```
?   	github.com/sollie/cloudflare-dyndns	[no test files]
?   	github.com/sollie/cloudflare-dyndns/cloudflare	[no test files]
?   	github.com/sollie/cloudflare-dyndns/dns	[no test files]
?   	github.com/sollie/cloudflare-dyndns/updater	[no test files]
```

### Linting and Code Quality

**Format code:**
```bash
cd source
go fmt ./...
```
This should produce no output if code is already formatted.

**Vet code:**
```bash
cd source
go vet ./...
```
This should produce no output if there are no issues.

**Note:** There is no golangci-lint or other linter configuration in this repository.

### Running the Application

The application requires environment variables to run:

```bash
export CFDD_TOKEN="your_cloudflare_api_token"
export CFDD_ZONE_1="example.com"
export CFDD_SUBDOMAINS_1="www,api"
cd source
./cloudflare-dyndns
```

Without environment variables, it will exit with error:
```
ERROR CFDD_TOKEN is required
```

## Docker Build

The Dockerfile uses a multi-stage build:
1. **builder**: golang:1.25.3-alpine - compiles Go binary
2. **upx**: ghcr.io/sollie/docker-upx:v5.0.1 - compresses binary
3. **final**: scratch - minimal runtime image

**Build Docker image:**
```bash
docker build -t cloudflare-dyndns .
```

**Note:** Docker builds may fail in restricted network environments due to Alpine package manager requiring internet access to dl-cdn.alpinelinux.org. The error will look like:
```
WARNING: updating and opening https://dl-cdn.alpinelinux.org/alpine/v3.22/main: Permission denied
```
This is expected in sandboxed CI environments and does not indicate a code issue.

## GitHub Workflows

### docker-publish.yml

Triggers on:
- Push to `main` branch
- Push of tags matching `v*.*.*`
- Pull requests to `main` branch

Steps:
1. Checkout repository
2. Install cosign (for signing, skipped on PRs)
3. Set up Docker Buildx
4. Login to GitHub Container Registry (ghcr.io)
5. Extract Docker metadata
6. Build and push Docker image (push only on main/tags)
7. Sign the published image with cosign

**For PRs:** Build is tested but not pushed to registry.

## Configuration

### Required Environment Variables
- `CFDD_TOKEN`: Cloudflare API token with DNS edit permissions

### Zone Configuration (at least one required)
- `CFDD_ZONE_1`: First zone (domain name)
- `CFDD_SUBDOMAINS_1`: Comma-separated subdomains for zone 1
- `CFDD_ZONE_N`: Additional zones (N=2,3,4...)
- `CFDD_SUBDOMAINS_N`: Subdomains for zone N

### Optional Configuration
- `CFDD_LOG_LEVEL`: Log level (debug, info, error, default: warn)
- `CFDD_TIMEOUT_SECONDS`: API timeout in seconds (default: 5)
- `CFDD_TTL`: DNS record TTL in seconds (default: 300)

### Configuration Validation
The application validates:
- Domain format: `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
- Subdomain format: `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`
- At least one zone configured
- No empty zones or subdomains

Validation failures cause immediate exit with descriptive error messages.

## Dependencies

### Direct Dependencies (go.mod)
- `github.com/cloudflare/cloudflare-go` v0.116.0 - Official Cloudflare API client
- `github.com/miekg/dns` v1.1.68 - DNS library for WAN IP resolution

### Key Indirect Dependencies
- `golang.org/x/net` v0.40.0
- `golang.org/x/sys` v0.33.0
- `golang.org/x/time` v0.9.0

## Key Behaviors and Patterns

1. **Logging**: Uses structured JSON logging via `log/slog` to stdout
2. **Error Handling**: Fails fast on critical errors (config, auth, network)
3. **Idempotency**: Only updates DNS if IP has changed
4. **Record Creation**: Automatically creates A records if they don't exist
5. **Context Usage**: All API calls use context with timeout
6. **Exit Codes**: Non-zero exit on any error

## Common Development Tasks

### Adding a New Package
1. Add import in relevant .go file
2. Run `go mod tidy` from source/ directory
3. Verify build still works

### Modifying Configuration
1. Edit `config.go` for new environment variables
2. Update `validateConfig()` if adding validation rules
3. Update README.md configuration section
4. Test with various invalid inputs to ensure validation works

### Adding API Methods
1. Add method to DNSClient interface in `cloudflare/client.go`
2. Implement method in Client struct
3. Use in `updater/service.go` or `main.go`

## Important Notes for Coding Agents

1. **Build Location**: ALWAYS run `go` commands from the `/source` directory, not the repository root.

2. **Binary Exclusion**: The built binary `source/cloudflare-dyndns` is in .gitignore and should NOT be committed.

3. **No Tests**: This repository has no test files. When adding new functionality, tests are not required but would be welcome following Go testing conventions.

4. **Docker Build Issues**: Docker builds may fail in restricted environments. This is expected and not a code issue. Focus on Go build validation instead.

5. **Configuration is Required**: The application cannot run without valid CFDD_TOKEN and at least one zone configured. Don't expect successful runs without proper environment variables.

6. **Code Style**: Use `go fmt` for formatting. The codebase uses minimal comments; follow this pattern unless complex logic requires explanation.

7. **Dependencies**: Only add dependencies if absolutely necessary. This is intentionally a small, focused application.

8. **Trust These Instructions**: These instructions are accurate as of the last repository state. Only search for additional information if you find these instructions incomplete or encounter unexpected behavior.

## Coding Standards for New Code

When adding new features or modifying existing code, follow these standards to maintain consistency with the existing codebase:

### Adding New Features

1. **Keep it minimal**: Only add what's necessary. This is intentionally a small, focused application.
2. **Follow existing patterns**: Review similar code in the repository before implementing new functionality.
3. **Use structured logging**: All logging must use `log/slog` with appropriate levels (Debug, Info, Warn, Error).
4. **Fail fast**: Exit with non-zero status on critical errors (config, auth, network failures).
5. **Use contexts**: All API calls and operations with timeouts should use `context.Context`.

### Error Handling Pattern

Always use structured logging with `slog.Error` and include descriptive context. Follow this pattern:

```go
result, err := someOperation()
if err != nil {
    slog.Error("Failed to perform operation", "error", err)
    os.Exit(1) // For critical errors
    // OR
    continue // For non-critical errors that can be skipped
}
```

**Key points:**
- Use `fmt.Sprintf` to format error messages with context
- Include what failed and the error details
- Critical errors (config, auth) should call `os.Exit(1)`
- Non-critical errors (individual DNS updates) can continue processing

### Security Guidelines

1. **Never log sensitive data**: API tokens, credentials, and secrets must NEVER be logged.
   - ✅ Good: `slog.Error("CFDD_TOKEN is required")`
   - ❌ Bad: `slog.Error(fmt.Sprintf("Invalid token: %s", token))`

2. **Validate all inputs**: Use regex patterns or validation functions for user-supplied data.
   - Domain names: `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`
   - Subdomains: `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`

3. **Sanitize error messages**: Ensure error messages don't leak sensitive information about the system or configuration.

### Dependencies Guidelines

1. **Minimize dependencies**: Only add new dependencies if absolutely necessary. Each dependency increases attack surface and maintenance burden.

2. **Use official libraries**: Prefer official SDKs (like `cloudflare-go`) over third-party alternatives.

3. **Run `go mod tidy` from the correct directory**:
   ```bash
   cd source
   go mod tidy
   ```
   ALWAYS run from `/source` directory, not the repository root.

4. **Verify builds after dependency changes**:
   ```bash
   cd source
   go build -o cloudflare-dyndns
   ```

5. **Check for vulnerabilities**: Before adding dependencies, ensure they don't have known security issues.
