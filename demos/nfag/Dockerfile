# OpenFiscalBR - Demo NFAg
# Multi-stage: compila o binario e serve com frontend estatico.

FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY packages/ packages/
COPY demos/nfag/ demos/nfag/
RUN CGO_ENABLED=0 go build -o /bin/nfag-demo ./demos/nfag

FROM alpine:3.19
COPY --from=build /bin/nfag-demo /bin/nfag-demo
COPY demos/nfag/frontend/ /static/
ENV STATIC_DIR=/static
EXPOSE 8080
ENTRYPOINT ["/bin/nfag-demo"]
