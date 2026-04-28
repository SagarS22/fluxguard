FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go/ratelimiter go.mod ./
RUN go build -o /bin/fluxguard ./go/ratelimiter

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/fluxguard /bin/fluxguard

EXPOSE 8080

CMD ["fluxguard"]