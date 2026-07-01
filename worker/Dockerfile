FROM golang:1.26-alpine AS builder

WORKDIR /app

# cache dep
COPY go.mod go.sum ./
RUN go mod download

# copy src
COPY . .

# buld
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -o worker \
    ./main.go

# final
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/worker .

USER nobody

ENTRYPOINT ["./worker"]

