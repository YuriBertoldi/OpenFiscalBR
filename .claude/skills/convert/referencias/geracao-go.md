# Templates de geração Go

Referência carregada sob demanda pela skill `/convert` (PASSO 5). As regras gerais de
mapeamento Delphi → Go estão no `CLAUDE.md` da raiz; aqui ficam os moldes literais.

## Header de licença

Obrigatório no topo de **todo** arquivo `.go` gerado. Substituir `<DATA>` pela data corrente
no formato `YYYY-MM-DD`.

```go
// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em <DATA>.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.
```

O texto do header é reproduzido exatamente como está acima, inclusive sem acentuação — é o
formato já usado nos arquivos existentes (`packages/comum/doc.go`, `packages/sped/doc.go`) e
mudá-lo produziria diff em todo o repositório.

## Imports

Sempre o module path completo:

```go
import "github.com/openfiscalbr/openfiscalbr/packages/comum"
```

## Enums

Para cada enum Delphi `TXxx = (val1, val2, val3)`:

```go
type Xxx int

const (
    Val1 Xxx = iota
    Val2
    Val3
)

var xxxNames = map[Xxx]string{
    Val1: "val1",
    Val2: "val2",
    Val3: "val3",
}

func (x Xxx) String() string {
    if name, ok := xxxNames[x]; ok {
        return name
    }
    return fmt.Sprintf("Xxx(%d)", int(x))
}

func ParseXxx(s string) (Xxx, error) {
    for k, v := range xxxNames {
        if strings.EqualFold(v, s) {
            return k, nil
        }
    }
    return 0, fmt.Errorf("valor Xxx invalido: %s", s)
}
```

Para enums de código fiscal (SPED, CST, CFOP), o `String()` devolve o **código exato da
legislação**, não o nome do identificador — ver `packages/sped/fiscal_types.go`, onde
`CodSit.String()` devolve `"00"`, `"01"`, e assim por diante. Quando o valor é opcional no
layout, criar um membro `...Nenhum` cujo `String()` devolve string vazia.

## Classes → structs

Para cada classe Delphi `TXxx = class(TParent)`:

```go
type Xxx struct {
    Parent  // embedding (heranca)
    // campos de properties (exported)
    CampoA string
    CampoB int
    // campos de eventos
    OnEvento func(sender interface{})
}

func NewXxx() *Xxx {
    x := &Xxx{}
    // logica do constructor Create
    return x
}
```

Só gerar `NewXxx()` quando o `constructor Create` do Delphi tiver lógica de fato (valores
default, alocação de filhos). Struct de dados puro é preenchido por literal — não inventar
construtor vazio.

## Properties com lógica

Se o setter Delphi tem lógica (não é apenas `write FX`), gerar getter/setter em vez de campo
exportado:

```go
func (x *Xxx) Campo() string {
    return x.campo
}

func (x *Xxx) SetCampo(value string) {
    // logica do setter
    x.campo = value
}
```

## Métodos abstratos → interface

Classe com métodos `virtual; abstract;` vira uma interface:

```go
type XxxInterface interface {
    MetodoAbstrato() error
}
```

## XML

Structs serializados em XML (classes de dados fiscais) levam struct tags com o nome **exato**
do schema da SEFAZ. Os nomes fiscais em PT-BR são preservados nos campos Go:

```go
type Ide struct {
    CUF      int    `xml:"cUF"`
    CNF      string `xml:"cNF"`
    NatOp    string `xml:"natOp"`
}
```

## Web services

Cada web service vira um método que monta o envelope SOAP, faz o POST via `net/http` e
parseia a resposta:

```go
func (ws *StatusServico) Executar() (*RetConsStatServ, error) {
    // montar envelope SOAP
    // POST via net/http
    // parsear resposta XML
}
```

## Errors

```go
var (
    ErrComponenteNaoConfigurado = errors.New("componente nao configurado")
)
```

Para erro que carrega contexto, seguir a **convenção oficial** do `CLAUDE.md` (seção
"Convencao de erros"): tipo com campos de contexto + `Unwrap()`, sentinelas testáveis com
`errors.Is`, wrapping com `%w` e nada de panic em biblioteca. Exemplos de referência:
`packages/nfgas/errors.go` (`ErroNFGas`, `ErrosLote`) e `packages/pcn/errors.go`
(`ErroLeitura`). O `ACBrError` de `comum` e o `SPEDFiscalError` de `sped` são legado — não
adotar em package novo.

## Dockerfile da demo

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY packages/ packages/
COPY demos/<pkg>/ cmd/
RUN CGO_ENABLED=0 go build -o server ./cmd/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /server
COPY demos/<pkg>/frontend/ /static/
EXPOSE 8080
CMD ["/server"]
```

`COPY go.mod ./` sem `go.sum`: o módulo hoje não tem dependência externa alguma e o arquivo
`go.sum` não existe — incluí-lo no `COPY` quebra o build. Quando a primeira dependência
externa entrar, passar a `COPY go.mod go.sum ./` e acrescentar `RUN go mod download`.

## docker-compose.yml da demo

```yaml
services:
  openfiscalbr-<pkg>:
    build:
      context: ../../
      dockerfile: demos/<pkg>/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - CERT_PATH=/certs/cert.pfx
      - CERT_PASS=
      - AMBIENTE=homologacao
    volumes:
      - ./certs:/certs
```

As variáveis de certificado só fazem sentido para componentes que assinam documento (DFe,
SAT). Para os demais, omitir.
