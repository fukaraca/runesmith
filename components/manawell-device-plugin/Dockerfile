FROM golang:1.24.5 AS builder
WORKDIR /src

ARG FULL_VERSION="dev"

# go.work is annoying, use dockerignore to make it lean
COPY . .
RUN mkdir -p /out

ENV GOWORK=auto
WORKDIR /src/components/manawell-device-plugin
RUN go build -v -ldflags="-X 'main.Version=${FULL_VERSION:-dev}'" -o /out/plugin ./

FROM alpine AS runtime
ARG FULL_VERSION="dev"

LABEL name="manawell-device-plugin"
LABEL org.opencontainers.image.source="https://github.com/fukaraca/runesmith"
LABEL org.opencontainers.image.authors="@fukaraca"
LABEL org.opencontainers.image.version="${FULL_VERSION}"

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app/manawell-device-plugin
COPY --from=builder /out/plugin ./plugin

RUN mkdir -p /app/manawell-device-plugin/configs
RUN mkdir -p /var/lib/kubelet/device-plugins

EXPOSE 8080

ENTRYPOINT ["/app/manawell-device-plugin/plugin"]
