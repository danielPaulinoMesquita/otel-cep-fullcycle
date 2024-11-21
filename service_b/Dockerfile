FROM golang:1.22

WORKDIR /app

COPY ../go.mod ./
COPY ../go.sum ./

RUN go mod download

COPY /service_b .

RUN go build -o /app/service_b main.go

EXPOSE 8081

CMD ["/app/service_b"]