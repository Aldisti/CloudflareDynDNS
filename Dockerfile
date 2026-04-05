FROM gcr.io/distroless/static-debian13

LABEL authors="me@aldisti.net"

ADD ./src/CloudflareDynDNS /app/CloudflareDynDNS

ENTRYPOINT ["/app/CloudflareDynDNS"]
