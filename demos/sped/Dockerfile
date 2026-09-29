FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY packages/ packages/
COPY demos/sped/ cmd/
RUN go build -o server ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
COPY demos/sped/frontend/ /static/
ENV STATIC_DIR=/static
EXPOSE 8080
CMD ["/server"]
