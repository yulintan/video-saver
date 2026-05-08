FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY *.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o video-saver .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/video-saver /video-saver
EXPOSE 8080
CMD ["/video-saver"]
