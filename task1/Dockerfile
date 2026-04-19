FROM golang:1.25 AS builder

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /custom-app ./cmd/custom-app

FROM alpine:3.22

WORKDIR /app

RUN mkdir -p /app/logs /app/config

COPY --from=builder /custom-app /app/custom-app

EXPOSE 8080

CMD ["/app/custom-app"]
