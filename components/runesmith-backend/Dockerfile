FROM golang:1.24.5 AS builder
WORKDIR /src

ARG FULL_VERSION="dev"

COPY . .

ENV GOWORK=auto
WORKDIR /src/components/runesmith-backend
RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
      -ldflags="-X 'main.Version=${FULL_VERSION:-dev}'" \
      -o /out/server ./cmd/server

FROM alpine AS runtime
ARG FULL_VERSION="dev"

LABEL name="runesmith-backend"
LABEL org.opencontainers.image.source="https://github.com/fukaraca/runesmith"
LABEL org.opencontainers.image.authors="@fukaraca"
LABEL org.opencontainers.image.version="${FULL_VERSION}"

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app/runesmith-backend
COPY --from=builder /out/server ./

RUN mkdir -p /app/runesmith-backend/configs

EXPOSE 8080

ENTRYPOINT ["/app/runesmith-backend/server"]