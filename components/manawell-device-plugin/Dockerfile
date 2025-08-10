FROM golang:1.24.5 AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the device plugin binary
# Adjust the path ./cmd/device-plugin if your main.go lives elsewhere
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" \
    -o /out/plugin ./

FROM alpine as runtime
RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app/manawell-device-plugin
COPY --from=builder /out/plugin ./plugin

RUN mkdir -p /app/manawell-device-plugin/configs
RUN mkdir -p /var/lib/kubelet/device-plugins

EXPOSE 8080

ENTRYPOINT ["/app/manawell-device-plugin/plugin"]
