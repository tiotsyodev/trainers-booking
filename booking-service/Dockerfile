FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.work go.work.sum ./
COPY gen/ ./gen/
COPY booking-service/ ./booking-service/

WORKDIR /src/booking-service
RUN CGO_ENABLED=0 GOOS=linux go build -o /booking-service ./cmd/main.go

FROM alpine:3.20
COPY --from=builder /booking-service /booking-service
EXPOSE 50051
ENTRYPOINT ["/booking-service"]