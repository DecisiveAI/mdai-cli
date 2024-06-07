FROM golang:1.22.1-alpine3.19 AS builder
WORKDIR /opt/mdai-cli
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /mdai-cli main.go

FROM alpine:latest
WORKDIR /
RUN apk update && apk upgrade && apk add --no-cache ca-certificates docker && update-ca-certificates
COPY --from=builder /mdai-cli /mdai-cli
ENTRYPOINT ["/mdai-cli"]
