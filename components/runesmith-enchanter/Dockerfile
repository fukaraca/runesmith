FROM golang:1.24.5 AS builder
WORKDIR /src

ARG FULL_VERSION="dev"

COPY . .

ENV GOWORK=auto
WORKDIR /src/components/runesmith-enchanter
RUN go build -v -ldflags="-X 'main.Version=${FULL_VERSION}'" -o /src/enchanter ./

FROM alpine AS runtime
ARG FULL_VERSION="dev"

LABEL name="runesmith-enchanter"
LABEL org.opencontainers.image.source="https://github.com/fukaraca/runesmith"
LABEL org.opencontainers.image.authors="@fukaraca"
LABEL org.opencontainers.image.version="${FULL_VERSION}"

RUN apk add --no-cache ca-certificates && update-ca-certificates

WORKDIR /app/runesmith-enchanter
COPY --from=builder /src/enchanter ./

EXPOSE 8080

ENTRYPOINT ["/app/runesmith-enchanter/enchanter"]
