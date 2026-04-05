# CloudflareDynDNS
Simple Go script to automatically update a DNS record in Cloudflare

## Build

### Local

If you don't need the docker image, then use the following commands:

```sh
git clone https://github.com/Aldisti/CloudflareDynDNS.git
cd CloudflareDynDNS
go build -C src .
```

### Docker Image

Build the Docker image using the following commands:

```sh
git clone https://github.com/Aldisti/CloudflareDynDNS.git
cd CloudflareDynDNS
CGO_ENABLED=0 GOOS=linux go build -C src -a -installsuffix cgo .
docker build -t cloudflare-dyndns .
```

## Run

Environment variables:
 - **ZONE_ID**: your Cloudflare DNS Zone ID
 - **API_TOKEN**: your Cloudflare API Token
 - **DOMAIN**: the full domain (e.g. myapp.example.com)
 - **INTERVAL**: how often update the DNS record, in seconds
 - **MAX_FAILURES**: maximum number of failures before stopping the program
 - **TIMEOUT**: http requests timeout, in seconds

Of these variables, only the last two (INTERVAL and MAX_FAILURES) are optional, the others are all mandatory.

### Docker

```sh
docker run aldisti/cloudflare-dyndns:latest \
    -e API_TOKEN <api_token> \
    -e ZONE_ID <zone_id> \
    -e DOMAIN <your.full.domain>
```
