FROM golang:1.22-alpine AS builder
WORKDIR /app
# O modulo nao tem dependencia externa, entao nao existe go.sum. Incluir o
# arquivo no COPY quebra o build. Quando a primeira dependencia externa entrar,
# voltar a "COPY go.mod go.sum ./" e acrescentar "RUN go mod download".
COPY go.mod ./
COPY packages/comum/ packages/comum/
COPY packages/sped/ packages/sped/
COPY demos/sped/ cmd/
RUN CGO_ENABLED=0 go build -o server ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
COPY demos/sped/frontend/ /static/
ENV STATIC_DIR=/static
EXPOSE 8080
CMD ["/server"]
