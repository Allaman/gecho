FROM alpine:latest
COPY gecho /
ENTRYPOINT ["/gecho"]
