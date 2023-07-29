FROM golang:1.19.1-alpine3.16 as builder

WORKDIR /app

COPY . .

RUN go build -o main cmd/main.go

FROM alpine:3.16

WORKDIR /app

COPY --from=builder /app/main .
COPY www ./www

CMD [ "/app/main" ]