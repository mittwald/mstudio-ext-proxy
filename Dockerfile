FROM golang:1.23 AS builder

WORKDIR /app

ENV CGO_ENABLED=0

# Copy go modules manifests and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and build the binary
COPY . .
RUN go build -o /mstudio-ext-proxy

FROM scratch

ENV PORT=8000

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /mstudio-ext-proxy /mstudio-ext-proxy

# Set entrypoint
ENTRYPOINT ["/mstudio-ext-proxy"]
