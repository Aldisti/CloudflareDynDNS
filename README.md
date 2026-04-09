# CloudflareDynDNS

CloudflareDynDNS keeps Cloudflare A records pointed at the current public IP address. It can run in two modes:

- Poller mode updates a fixed list of domains on a schedule.
- Listener mode exposes an authenticated HTTP endpoint for routers or other clients to trigger updates.

The listener has been tested with a FRITZ!Box 7530 and should work with other Fritz!Box routers that can call a webhook after the WAN IP changes.

## Requirements

- A Cloudflare API token with permission to read zones and edit DNS records.
- A build of the Go binary at `src/CloudflareDynDNS` before creating the Docker image.

## Configuration

Environment variables are shared by both modes unless noted otherwise.

| Variable         | Mode     | Required | Default | Description |
|------------------|----------|----------|---------|-------------|
| **MODE**         | Both     | No       | POLLER  | Execution mode: _POLLER_ or _LISTENER_ |
| **API_TOKEN**    | Both     | Yes      | N/A     | Cloudflare API token |
| **TIMEOUT**      | Both     | No       | 5       | HTTP request timeout in seconds |
| **DOMAIN**       | POLLER   | Yes      | N/A     | Comma-separated list of domain names to keep updated |
| **INTERVAL**     | POLLER   | No       | 60      | Polling interval in seconds |
| **MAX_FAILURES** | POLLER   | No       | -1      | Maximum consecutive failures before stopping (-1 disables the limit) |
| **COOLDOWN**     | POLLER   | No       | -1      | Failure counter reset after this many seconds since last failure (-1 disables the cooldown) |
| **PORT**         | LISTENER | No       | 8080    | HTTP server port |
| **USERNAME**     | LISTENER | Yes      | N/A     | Basic auth username |
| **PASSWORD**     | LISTENER | Yes      | N/A     | Basic auth password |


## How It Works

In Poller mode, the application looks up each domain in Cloudflare, creates missing A records, and then refreshes them on every interval.

In Listener mode, the application serves updates at `/update` over HTTP Basic Auth. The hostname list is passed as repeated `hostname` query parameters, for example `?hostname=foo.example.com&hostname=bar.example.com`.

Both modes resolve the current public IP from `https://api.ipify.org`.

## Build

Build the Linux binary compatible with the distroless image:

```sh
cd src
CGO_ENABLED=0 GOOS=linux go build -o CloudflareDynDNS .
```

Then build the image from the repository's root dir:

```sh
docker build -t cloudflare-dyndns:latest .
```

## Run

Poller mode:

```sh
docker run -d \
    -e MODE=POLLER \
    -e API_TOKEN=<api_token> \
    -e DOMAIN=foo.example.com,bar.example.com \
    cloudflare-dyndns:latest
```

Listener mode:

```sh
docker run -d \
    -p 8080:8080 \
    -e MODE=LISTENER \
    -e API_TOKEN=<api_token> \
    -e USERNAME=something \
    -e PASSWORD=a-very-random-secret \
    cloudflare-dyndns:latest
```

To trigger an update in listener mode:

```sh
curl -u something:a-very-random-secret \
    "http://localhost:8080/update?hostname=foo.example.com&hostname=bar.example.com"
```

## Notes

- Poller mode creates missing A records automatically.
- Listener mode updates existing A records only.

