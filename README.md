# CloudflareDynDNS

CloudflareDynDNS is a lightweight Go application that keeps a Cloudflare DNS record updated with your current public IP.

## Requirements

- Go 1.24+
- A Cloudflare API token with permission to edit DNS records

## Build

### Build locally

```sh
git clone https://github.com/Aldisti/CloudflareDynDNS.git
cd CloudflareDynDNS
go build -C src -o CloudflareDynDNS .
```

This creates the binary at `src/CloudflareDynDNS`.

### Build Docker image

```sh
git clone https://github.com/Aldisti/CloudflareDynDNS.git
cd CloudflareDynDNS
CGO_ENABLED=0 GOOS=linux go build -C src -a -installsuffix cgo -o CloudflareDynDNS .
docker build -t cloudflare-dyndns:latest .
```

## Configuration

Environment variables:

- `API_TOKEN` (required): Cloudflare API token
- `DOMAIN` (required): full domain name to update (for example myapp.example.com)
- `INTERVAL` (optional): update interval in seconds
- `MAX_FAILURES` (optional): maximum consecutive failures before exit
- `TIMEOUT` (optional): HTTP timeout in seconds

Only `API_TOKEN` and `DOMAIN` are mandatory.

## Run

### Run binary

```sh
API_TOKEN=<api_token> DOMAIN=<your.full.domain> ./src/CloudflareDynDNS
```

### Run with Docker

```sh
docker run --rm \
    -e API_TOKEN=<api_token> \
    -e DOMAIN=<your.full.domain> \
    cloudflare-dyndns:latest
```
