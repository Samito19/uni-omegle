# syntax=docker/dockerfile:1

FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go . 

RUN go build -o backend

EXPOSE 8080

CMD ["./backend"]

# CMD ["go", "run", "*.go"]









