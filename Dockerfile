FROM golang:1.15.3-alpine3.12 AS builder
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on
WORKDIR /app
COPY go.mod go.sum slagobot.go ./
RUN go build -o slagobot -v -x slagobot.go && chmod +x slagobot

FROM alpine:3.12
RUN apk update \
    && apk add ca-certificates \
    && rm -rf /var/chache/apk/* \
    && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /app/slagobot .
ENTRYPOINT ["./slagobot"]
