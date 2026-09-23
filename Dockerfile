FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o scheduler ./main.go


FROM ubuntu:latest

WORKDIR /app

ENV TODO_PORT=7540

ENV TODO_DBFILE=/app/scheduler.db

ENV TODO_PASSWORD="12345"

EXPOSE 7540

COPY --from=builder /app/scheduler .

COPY web/ ./web/

CMD ["./scheduler"]
