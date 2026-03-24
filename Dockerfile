FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /app ./cmd/main.go

FROM alpine:3.20

COPY --from=builder /app /app

EXPOSE 8080
CMD ["/app"]
