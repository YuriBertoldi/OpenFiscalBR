# OpenFiscalBR - Demo NF-e
# Multi-stage: compila o binario e serve com frontend estatico.

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY packages/ packages/
COPY demos/nfe/ demos/nfe/
RUN CGO_ENABLED=0 go build -o /bin/nfe-demo ./demos/nfe

FROM alpine:3.19
# A transmissao a SEFAZ e TLS: sem o bundle de raizes o handshake falha
# dentro do container, embora funcione no `go run` da maquina.
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/nfe-demo /bin/nfe-demo
COPY demos/nfe/frontend/ /static/
ENV STATIC_DIR=/static
EXPOSE 8080
ENTRYPOINT ["/bin/nfe-demo"]
