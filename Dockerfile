FROM golang:1.23.0 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/neobit ./cmd

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /app/bin/neobit /app/neobit

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/neobit"]
