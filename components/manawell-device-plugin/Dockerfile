FROM golang:1.24.5 AS builder
WORKDIR /src

ARG FULL_VERSION="dev"

# go.work is annoying, use dockerignore to make it lean
COPY . .

ENV GOWORK=auto
WORKDIR /src/components/manawell-device-plugin
RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
      -ldflags="-X 'main.Version=${FULL_VERSION:-dev}'" \
      -o /out/plugin ./

FROM alpine AS runtime
ARG FULL_VERSION="dev"

LABEL name="manawell-device-plugin"
LABEL org.opencontainers.image.source="https://github.com/fukaraca/runesmith"
LABEL org.opencontainers.image.authors="@fukaraca"
LABEL org.opencontainers.image.version="${FULL_VERSION}"

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app/manawell-device-plugin
COPY --from=builder /out/plugin ./

RUN mkdir -p /app/manawell-device-plugin/configs
RUN mkdir -p /var/lib/kubelet/device-plugins

EXPOSE 8080

ENTRYPOINT ["/app/manawell-device-plugin/plugin"]
