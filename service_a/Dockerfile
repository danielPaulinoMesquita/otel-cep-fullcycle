FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY ../go.mod ./
COPY ../go.sum ./

RUN go mod download

COPY /service_a .

RUN go build -o service_a main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /build/service_a /app/

EXPOSE 8080

CMD ["./service_a"]