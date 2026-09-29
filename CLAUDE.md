# OpenFiscalBR - Automacao Fiscal Brasileira em Go

Port nativo em Go do [Projeto ACBr](http://projetoacbr.com.br) (Delphi/Lazarus).
Licenciado sob LGPL 2.1+ (obra derivada do Projeto ACBr).

## Estrutura do Projeto

```
packages/          -- Packages Go (destino da conversao)
  comum/           -- Layer 0: base, utils, validador (de ACBrComum)
  pcn/             -- Layer 1: gerador/leitor XML (de PCNComum)
  dfe/             -- Layer 2: DFe base, SSL, config (de ACBrDFe)
  nfe/             -- Layer 3: NFe (de ACBrNFe)
  cte/             -- Layer 3: CTe
  mdfe/            -- Layer 3: MDFe
  boleto/          -- Layer 3: Boleto
  pixcd/           -- Layer 3: PIX
  sat/             -- Layer 3: SAT
  nfsex/           -- Layer 3: NFSeX
demos/             -- Demos (API REST + Frontend + Docker)
```

Module: `github.com/openfiscalbr/openfiscalbr`

## Skill Disponivel

- `/convert <ComponentName> <DelphiSourcePath>` -- Converte um componente Delphi para Go

## Regras de Mapeamento Delphi -> Go

| Delphi | Go |
|---|---|
| `TXxx = class(TParent)` | `type Xxx struct { Parent; ... }` (embedding) |
| `property X: T read FX write SetX` | Campo exportado `X T` ou getter/setter se houver logica no setter |
| `TEnum = (val1, val2)` | `type Enum int` + `const ( Val1 Enum = iota; Val2 )` + `String()` + `Parse()` |
| `TObjectList<T>` | `[]T` (slice) |
| `constructor Create` | `func NewXxx() *Xxx` |
| `destructor Destroy` | desnecessario (GC do Go) |
| `procedure virtual; abstract;` | metodo em Go `interface` |
| `procedure override;` | metodo no struct que embute (shadow) |
| `TNotifyEvent = procedure(Sender: TObject) of object` | `type NotifyFunc func(sender interface{})` |
| `raise EException.Create(msg)` | `return fmt.Errorf(msg)` ou custom error type |
| `TStringList` | `[]string` |
| `TMemIniFile` / `TIniFile` | `gopkg.in/ini.v1` |
| XML gerador (`pcnGerador`) | `encoding/xml` com struct tags |
| XML leitor (`pcnLeitor`) | `encoding/xml` decoder |
| SOAP web services | `net/http` + XML marshal/unmarshal |
| Certificado digital PFX | `crypto/tls`, `crypto/x509`, `golang.org/x/crypto/pkcs12` |
| `TDateTime` | `time.Time` |
| `Currency` | `float64` ou `github.com/shopspring/decimal` |
| `record` | `struct` (sem metodos ou com metodos simples) |
| `array of T` | `[]T` |
| `set of TEnum` | `map[Enum]bool` ou bitfield |

## Convencoes de Nomes

- **Package names**: lowercase sem underscore (`comum`, `nfe`, `boleto`, `pixcd`)
- **Arquivos**: snake_case (`nfe_classes.go`, `web_services.go`, `conversao.go`)
- **Structs**: PascalCase sem prefixo `T` (`TACBrNFe` -> `NFe`, `TACBrBoleto` -> `Boleto`)
- **Campos fiscais**: manter nomes PT-BR originais (`cUF`, `CNPJ`, `nNF`, `vBC`) pois mapeiam para schemas XML da SEFAZ
- **Enums**: PascalCase (`TaProducao`, `TaHomologacao`)
- **Constantes**: PascalCase ou UPPER_SNAKE conforme contexto

## Ordem de Dependencias (OBRIGATORIA)

A conversao deve respeitar esta ordem -- nunca converter um componente sem que suas dependencias existam:

1. **Layer 0**: `comum` (ACBrComum) -- sem dependencias
2. **Layer 1**: `pcn` (PCNComum) -- depende de `comum`
3. **Layer 2**: `dfe` (ACBrDFe base) -- depende de `comum`, `pcn`
4. **Layer 3**: componentes finais -- dependem de layers 0-2:
   - `nfe` (ACBrNFe) -> `comum`, `pcn`, `dfe`
   - `cte` (ACBrCTe) -> `comum`, `pcn`, `dfe`
   - `mdfe` (ACBrMDFe) -> `comum`, `pcn`, `dfe`
   - `boleto` (ACBrBoleto) -> `comum`
   - `pixcd` (ACBrPIXCD) -> `comum`
   - `sat` (ACBrSAT) -> `comum`
   - `nfsex` (ACBrNFSeX) -> `comum`, `pcn`, `dfe`

## Header de Licenca Obrigatorio

TODO arquivo `.go` gerado DEVE comecar com:

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

## Estrutura Padrao de Package Go

Cada componente convertido segue esta estrutura de arquivos:

```
packages/<pkg>/
  doc.go              -- documentacao do package + header licenca
  types.go            -- enums, constantes, tipos simples
  conversao.go        -- enum<->string, Parse funcs
  classes.go          -- structs de dados (mapeados das classes Delphi)
  configuracoes.go    -- structs de configuracao
  web_services.go     -- logica HTTP/SOAP
  errors.go           -- error types customizados
  <pkg>.go            -- struct principal + metodos
  <pkg>_test.go       -- testes unitarios reais (veja Testes Obrigatorios)
  README.md           -- documentacao do package
```

## Testes Obrigatorios

Cada package convertido DEVE ter testes unitarios reais que validem:

- **Enums**: String() retorna o codigo SPED exato conforme tabelas da legislacao
- **Formato de saida**: linhas pipe-delimited comecam e terminam com |
- **Campos**: datas ddmmaaaa, monetarios com virgula, inteiros com zeros a esquerda
- **Validacoes de negocio**: regras da legislacao (DT_INI primeiro dia do mes, etc.)
- **Integracao**: gerar arquivo completo, validar estrutura e contagem de linhas
- **Casos de borda**: ano bissexto, valores zero com nulo true/false, campos vazios

Os testes devem rodar com `go test ./packages/<pkg>/... -v` sem dependencias externas.

## README por Package

Cada package DEVE ter um `README.md` documentando:
- Descricao e componente Delphi correspondente
- Tipos e structs principais com exemplos de uso
- Tabela de registros/tipos implementados
- Como rodar os testes

## README Raiz

O `README.md` na raiz do projeto deve ser mantido atualizado com:
- Status de cada package (Completo/Pendente)
- Status de cada demo
- Exemplo de uso rapido
- Instrucoes de build e testes

## Estrutura Padrao de Demo

Cada demo segue esta estrutura:

```
demos/<pkg>/
  main.go             -- servidor HTTP (net/http)
  handlers.go         -- route handlers
  frontend/
    index.html        -- SPA HTML/JS
    app.js            -- logica frontend (fetch API)
    style.css         -- estilizacao
  Dockerfile          -- multi-stage build (Go + static frontend)
  docker-compose.yml  -- compose para subir
  README.md           -- instrucoes de uso
```

## Metadados de Conversao

Arquivo `.openfiscalbr-meta.json` na raiz rastreia o estado de cada conversao (hashes dos fontes Delphi, data, arquivos Go gerados). Usado pelo modo de atualizacao da skill `/convert`.
