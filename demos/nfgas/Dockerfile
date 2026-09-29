# OpenFiscalBR - Demo NFGas
# Multi-stage: compila o binario e serve com frontend estatico.

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY packages/ packages/
COPY demos/nfgas/ demos/nfgas/
RUN CGO_ENABLED=0 go build -o /bin/nfgas-demo ./demos/nfgas

FROM alpine:3.19
COPY --from=build /bin/nfgas-demo /bin/nfgas-demo
COPY demos/nfgas/frontend/ /static/
ENV STATIC_DIR=/static
EXPOSE 8080
ENTRYPOINT ["/bin/nfgas-demo"]
