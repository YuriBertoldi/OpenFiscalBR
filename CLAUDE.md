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

## Skills Disponiveis

Definidas em `.claude/skills/<nome>/SKILL.md`:

- `/convert <ComponentName> <DelphiSourcePath>` -- Converte um componente Delphi para Go
- `/validar-package [pkg]` -- gofmt, build, vet, testes, header de licenca, module path
- `/adicionar-registro-sped <registro>` -- Acrescenta um registro ao `packages/sped`
- `/gerar-demo <pkg>` -- Gera a demo em `demos/<pkg>/`
- `/sincronizar-meta` -- Atualiza `.openfiscalbr-meta.json` e o status dos packages

Subagente `revisor-go` (`.claude/agents/revisor-go.md`) -- revisa codigo portado quanto a
fidelidade ao Delphi de origem e as convencoes deste arquivo. So reporta, nao corrige.

> Skills sao carregadas na inicializacao da sessao. Ao criar uma skill nova, ela so fica
> disponivel a partir da proxima sessao aberta na raiz do projeto.

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

## Fidelidade ao Delphi -- o defeito dominante do port

Um port que compila, passa nos testes e gera documento invalido e o caso comum aqui. O que se
perde nao e a estrutura: e a **condicional em volta dela**. Na auditoria do `packages/sped`,
100% das divergencias encontradas eram disso, e nenhuma aparecia em build, `vet` ou teste.

Ao portar ou revisar qualquer writer, procure no `.pas` de origem:

- `ifthen` / `IfThen` / `EncodeDate` -- campo que so existe a partir de certa data
- `COD_VER` / `vlVersaoNNN` -- campo que depende da **versao do leiaute** (dimensao diferente
  da data; as duas convivem)
- `if` envolvendo a escrita inteira -- bloco ou registro que nao existe no periodo
- `case` de enum dentro de `if DT_INI < ...` -- tabela de codigos **remanejada** por epoca
- `Check(` -- validacao obrigatoria
- `if Assigned(On...)` -- callback a expor

Regras derivadas disso:

- **Contagem de campos nao e auditoria.** Struct copiado de outro registro tem a quantidade
  certa e o conteudo errado. Confira nome a nome e confira a hierarquia.
- **Propague o estado do componente para as partes.** Writer que decide layout por vigencia
  precisa receber a data; se nao receber, falha em silencio.
- **Quando o ACBr estiver errado, replique se for deterministico** (com teste e comentario
  dizendo que e intencional). Nao replique o que depende de ordem de iteracao ou variavel nao
  inicializada -- documente a divergencia e leve ao usuario.
- **Separe "tipos declarados" de "tipos exercitados"** ao reportar cobertura. Struct sem writer
  e preenchido pelo consumidor e nunca aparece na saida.

Ferramenta de apoio: `ferramentas/comparar-campos.py` (veja `ferramentas/README.md` para os
limites dela).

## Fronteira entre package e demo

`packages/<pkg>` e **biblioteca**: gera o documento fiscal e devolve dado. Nao expoe URL,
nao serve HTTP, nao fala JSON. Todo transporte vive em `demos/<pkg>`.

Na pratica isso significa que **nenhum arquivo em `packages/` importa `net/http`,
`encoding/json`, `net` ou `html/template`**. A dependencia e numa direcao so: o demo importa
o package, nunca o contrario.

Excecao unica: `packages/dfe/soap.go` importa `net/http` como **CLIENTE** SOAP dos web
services da SEFAZ. Transmitir o documento (recepcao, consulta, evento) e a propria funcao
fiscal — e o papel do `TDFeWebService` do ACBr — e nao exposicao de transporte. Continua
proibido em qualquer package, `dfe` incluido: servidor HTTP, handler, `encoding/json` e
template. Cliente HTTP fora do `dfe` tambem e achado.

Por que a regra existe:

- Quem consome a biblioteca (um ERP, um job, um CLI) nao quer subir servidor para gerar um
  arquivo. O exemplo do `README.md` da raiz roda sem HTTP nenhum.
- Formato de transporte muda por consumidor (REST, gRPC, fila, CLI); leiaute fiscal nao.
  Misturar os dois faria a mudanca de um arrastar o outro.
- O demo e descartavel e serve para demonstrar; o package e o produto.

`/validar-package` verifica isso automaticamente.

## Convencao de erros (OFICIAL)

O padrao de erro do projeto e o **idiomatico do Go**, estabelecido nos packages `pcn`, `rtc` e
`nfgas`:

- **Sentinelas** com `errors.New`, prefixadas pelo package (`nfgas: XML da NFGas nao
  carregado`), testaveis com `errors.Is`.
- **Tipos com contexto** implementando `Unwrap()` (e `Unwrap() []error` em agregadores como
  `nfgas.ErrosLote`), para `errors.Is`/`errors.As` alcancarem a causa. O contexto carrega o que
  torna o erro acionavel: caminho da tag (`infNFGas/ide/dhEmi`), indice no lote, arquivo, chave
  de acesso.
- **Wrapping** com `fmt.Errorf("...: %w", err)`.
- **`raise EACBrException` do Delphi vira `error`**, nunca panic. Biblioteca nao entra em
  panico com entrada torta; numa importacao em lote, um documento invalido vira entrada de erro
  e os demais seguem.
- Fase futura declarada e nao implementada devolve o sentinela `ErrNaoImplementado` do package.

O `comum.ACBrError` e o `Check`/panic do `TXTClass` sao **legado do `sped`** (herdados do
TACBrTXTClass): permanecem la por compatibilidade, mas NAO devem ser adotados em package novo.
Exemplo de referencia: `packages/nfgas/errors.go`.

## Higiene do Repositorio

- Arquivos `.go` sempre em **LF** (ver `.gitattributes`). Fontes Delphi `.pas`/`.dfm` em CRLF.
- `go build ./... && go vet ./... && go test ./packages/...` verdes antes de qualquer commit.
- Para conferir formatacao, use `/validar-package` -- `gofmt -l` sozinho da falso positivo em
  100% dos arquivos quando o checkout esta com `core.autocrlf=true`.

## Convencoes do Package sped

- Structs de registro em `bloco_<x>.go`; writers em `write_bloco_<x>.go` (padrao alvo) ou na
  secao do bloco em `write_blocos.go`.
- Writer de registro de dados e **nao exportado** (`writeRegistroC170`); so `X001`, `X990` e
  `0000` sao exportados.
- Todo writer carrega doc-comment `// Formato: |REG|CAMPO1|...|` -- e o unico contrato sobre a
  ordem dos campos, que nada mais valida.
- Todo writer incrementa `QtdLinX` (alimenta o 9999) **e** `RegistroXXXCount` (alimenta o 9900).
- A contagem do 9900 usa sempre o contador, nunca `len(slice)` -- `len()` nao alcanca registro
  neto (o 0175 pertence a cada 0150).
- Para acrescentar um registro, use `/adicionar-registro-sped`.

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
