FROM scratch

LABEL authors="me@aldisti.net"

ADD src/CloudflareDynDNS /

CMD ["/CloudflareDynDNS"]

