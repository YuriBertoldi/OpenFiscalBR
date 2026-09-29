---
name: convert
description: Converte um componente ACBR Delphi para package Go nativo (OpenFiscalBR)
arguments:
  - name: ComponentName
    description: "Nome do componente ACBR (ex: ACBrNFe, ACBrBoleto, ACBrComum, PCNComum, ACBrDFe, ACBrPIXCD, ACBrSAT)"
    required: true
  - name: DelphiSourcePath
    description: "Caminho raiz da pasta Fontes dos fontes Delphi (ex: C:\\MeusACBr\\Fontes)"
    required: true
---

# Skill: /convert - Conversor ACBR Delphi -> Go (OpenFiscalBR)

Voce e um agente especializado em converter componentes ACBR de Delphi para Go nativo.
Siga EXATAMENTE este workflow passo a passo. Leia o CLAUDE.md do projeto para as regras de mapeamento.

## PASSO 1: Validar parametros e localizar fontes Delphi

1. Receba `$ARGUMENTS.ComponentName` e `$ARGUMENTS.DelphiSourcePath`.
2. Valide que `DelphiSourcePath` existe (use Glob para verificar).
3. Localize o diretorio do componente dentro de `DelphiSourcePath`:
   - Mapeamento de componentes para diretorios:
     - `ACBrComum` -> `<path>/ACBrComum/`
     - `PCNComum` -> `<path>/PCNComum/`
     - `ACBrDFe` -> `<path>/ACBrDFe/` (apenas arquivos raiz, NAO subpastas de componentes)
     - `ACBrNFe` -> `<path>/ACBrDFe/ACBrNFe/`
     - `ACBrCTe` -> `<path>/ACBrDFe/ACBrCTe/`
     - `ACBrMDFe` -> `<path>/ACBrDFe/ACBrMDFe/`
     - `ACBrBPe` -> `<path>/ACBrDFe/ACBrBPe/`
     - `ACBrNFSeX` -> `<path>/ACBrDFe/ACBrNFSeX/`
     - `ACBrNFSe` -> `<path>/ACBrDFe/ACBrNFSe/`
     - `ACBrNF3e` -> `<path>/ACBrDFe/ACBrNF3e/`
     - `ACBrNFCom` -> `<path>/ACBrDFe/ACBrNFCom/`
     - `ACBrGNRE` -> `<path>/ACBrDFe/ACBrGNRE/`
     - `ACBrReinf` -> `<path>/ACBrDFe/ACBrReinf/`
     - `ACBreSocial` -> `<path>/ACBrDFe/ACBreSocial/`
     - `ACBrBoleto` -> `<path>/ACBrBoleto/`
     - `ACBrPIXCD` -> `<path>/ACBrPIXCD/`
     - `ACBrSAT` -> `<path>/ACBrSAT/`
     - `ACBrTEFD` -> `<path>/ACBrTEFD/`
     - `ACBrSerial` -> `<path>/ACBrSerial/`
     - `ACBrTCP` -> `<path>/ACBrTCP/`
     - `ACBrTXT` -> `<path>/ACBrTXT/`
     - `ACBrDiversos` -> `<path>/ACBrDiversos/`
     - `ACBrPagFor` -> `<path>/ACBrPagFor/`
     - `ACBrBaaS` -> `<path>/ACBrBaaS/`
     - `ACBrOpenDelivery` -> `<path>/ACBrOpenDelivery/`
   - Se o componente nao estiver na lista, busque com Glob: `<path>/**/*<ComponentName>*/`
   - Se nao encontrar, pergunte ao usuario o caminho exato.
4. Verifique que existem arquivos `.pas` no diretorio encontrado.

## PASSO 2: Resolver dependencias

1. Determine o package Go destino usando esta tabela:

| ComponentName | Package Go | Diretorio destino |
|---|---|---|
| `ACBrComum` | `comum` | `packages/comum/` |
| `PCNComum` | `pcn` | `packages/pcn/` |
| `ACBrDFe` | `dfe` | `packages/dfe/` |
| `ACBrNFe` | `nfe` | `packages/nfe/` |
| `ACBrCTe` | `cte` | `packages/cte/` |
| `ACBrMDFe` | `mdfe` | `packages/mdfe/` |
| `ACBrBPe` | `bpe` | `packages/bpe/` |
| `ACBrNFSeX` | `nfsex` | `packages/nfsex/` |
| `ACBrNFSe` | `nfse` | `packages/nfse/` |
| `ACBrNF3e` | `nf3e` | `packages/nf3e/` |
| `ACBrNFCom` | `nfcom` | `packages/nfcom/` |
| `ACBrBoleto` | `boleto` | `packages/boleto/` |
| `ACBrPIXCD` | `pixcd` | `packages/pixcd/` |
| `ACBrSAT` | `sat` | `packages/sat/` |
| `ACBrGNRE` | `gnre` | `packages/gnre/` |
| `ACBrReinf` | `reinf` | `packages/reinf/` |
| `ACBreSocial` | `esocial` | `packages/esocial/` |
| `ACBrTEFD` | `tefd` | `packages/tefd/` |
| `ACBrPagFor` | `pagfor` | `packages/pagfor/` |
| `ACBrBaaS` | `baas` | `packages/baas/` |

2. Verifique as dependencias obrigatorias:

| Package | Depende de |
|---|---|
| `comum` | nenhuma |
| `pcn` | `comum` |
| `dfe` | `comum`, `pcn` |
| `nfe`, `cte`, `mdfe`, `bpe`, `nfsex`, `nfse`, `nf3e`, `nfcom`, `gnre`, `reinf`, `esocial` | `comum`, `pcn`, `dfe` |
| `boleto` | `comum` |
| `pixcd` | `comum` |
| `sat` | `comum` |
| `tefd` | `comum` |
| `pagfor` | `comum` |
| `baas` | `comum` |

3. Para cada dependencia, verifique se o diretorio `packages/<dep>/` existe E contem arquivos `.go`.
4. Se alguma dependencia nao existir, PERGUNTE ao usuario:
   > "O componente X depende de [dep1, dep2] que ainda nao foram convertidos. Deseja que eu converta essas dependencias primeiro?"
   Se sim, execute este mesmo workflow recursivamente para cada dependencia, na ordem correta.

## PASSO 3: Verificar se package Go ja existe

1. Verifique se `packages/<pkg>/` existe e contem arquivos `.go`.
2. Se NAO existe -> va para PASSO 4 (criacao completa).
3. Se existe -> va para PASSO 7 (modo atualizacao).

## PASSO 4: Escanear fontes Delphi

1. Liste todos os arquivos `.pas` recursivamente no diretorio do componente (use Glob `**/*.pas`).
2. EXCLUA arquivos de:
   - Pastas `DANFE/`, `DACTE/`, `DAMDFE/`, `DANFSe/`, `DACE/`, `DANF3e/`, `DANFCom/` (reports visuais, nao se aplicam a Go)
   - Arquivos `*Reg.pas` (registradores de componentes Delphi design-time)
   - Arquivos de pastas `Fast/`, `Fortes/`, `LazReport/`, `EscPos/` (report engines especificos)
3. Para cada arquivo `.pas`, leia o conteudo e extraia:
   - **Nome da unit** (linha `unit XXX;`)
   - **Clausula `uses`** (secoes `interface uses` e `implementation uses`)
   - **Declaracoes de tipo na secao `type`**:
     - Classes: `TXxx = class(TParent)` com properties, methods, fields
     - Enums: `TXxx = (val1, val2, ...)`
     - Records: `TXxx = record`
     - Arrays: `TXxxArray = array of TXxx`
   - **Hierarquia** (classe pai)
   - **Properties**: nome, tipo, read accessor, write accessor, se tem logica no setter
   - **Metodos**: nome, parametros (nome:tipo), tipo retorno (function), visibilidade (public/protected/private/published), modificadores (virtual/override/abstract/static/class)
   - **Constantes** (secao `const`)
   - **Tipos de eventos** (`TOnXxx = procedure(Sender: TObject; ...) of object`)
4. Monte um inventario completo de todos os tipos, seus campos e metodos.

## PASSO 5: Gerar package Go

1. Crie o diretorio `packages/<pkg>/`.
2. Gere os arquivos Go seguindo a estrutura padrao (veja CLAUDE.md).
3. REGRAS DE GERACAO:

### Header de licenca
Todo arquivo `.go` DEVE comecar com o header LGPL (veja CLAUDE.md).
Substitua `<DATA>` pela data atual no formato `YYYY-MM-DD`.

### Package declaration
```go
package <pkg>
```

### Imports
Use o module path completo: `github.com/openfiscalbr/openfiscalbr/packages/comum`

### Enums
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

### Classes -> Structs
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

### Properties com logica
Se o setter Delphi tem logica (nao e apenas `write FX`), crie getter/setter:
```go
func (x *Xxx) Campo() string {
    return x.campo
}

func (x *Xxx) SetCampo(value string) {
    // logica do setter
    x.campo = value
}
```

### Metodos abstratos -> Interface
Se a classe tem metodos `virtual; abstract;`, crie uma interface:
```go
type XxxInterface interface {
    MetodoAbstrato() error
}
```

### XML
Para structs que sao serializados em XML (classes de dados fiscais):
```go
type Ide struct {
    CUF      int    `xml:"cUF"`
    CNF      string `xml:"cNF"`
    NatOp    string `xml:"natOp"`
    // ...
}
```

### Web Services
Para cada web service, crie um metodo que faz HTTP request:
```go
func (ws *StatusServico) Executar() (*RetConsStatServ, error) {
    // montar envelope SOAP
    // fazer POST via net/http
    // parsear resposta XML
}
```

### Errors
```go
var (
    ErrComponenteNaoConfigurado = errors.New("componente nao configurado")
    // ...
)
```

4. Gere testes unitarios reais em `<pkg>_test.go` cobrindo:
   - **Criacao e defaults**: NewXxx() retorna campos com valores corretos
   - **Enums**: String() retorna o codigo SPED exato conforme legislacao (ex: CstIcms "000", CodSit "02")
   - **Formato de saida**: linhas pipe-delimited comecam e terminam com |
   - **Formatacao de campos**: datas ddmmaaaa, valores monetarios com virgula, inteiros com zeros a esquerda
   - **Validacoes**: campos obrigatorios, regras de negocio (DT_INI primeiro dia do mes, etc.)
   - **Integracao**: gerar arquivo completo e validar estrutura, ordem dos blocos, contagem de linhas
   - **Casos de borda**: ano bissexto, valores zero com nulo=true/false, campos vazios

5. Gere um `README.md` no diretorio do package com:
   - Descricao do package e qual componente Delphi corresponde
   - Tipos e structs principais com exemplos de uso
   - Como rodar os testes
   - Tabela de registros/tipos implementados

## PASSO 6: Gerar demo

1. Crie o diretorio `demos/<pkg>/`.
2. Gere os arquivos:

### main.go
Servidor HTTP com rotas baseadas no tipo de componente:

**DFe (NFe, CTe, MDFe, BPe, NFSeX, etc.):**
```
POST /api/criar           -- Criar documento
POST /api/validar         -- Validar XML
POST /api/assinar         -- Assinar digitalmente
POST /api/transmitir      -- Transmitir para SEFAZ
GET  /api/status          -- Status do servico SEFAZ
POST /api/cancelar        -- Cancelar documento
POST /api/carta-correcao  -- Carta de correcao
POST /api/inutilizar      -- Inutilizar numeracao
GET  /api/consultar/:chave -- Consultar por chave
```

**Boleto:**
```
POST /api/boleto          -- Gerar boleto
GET  /api/boleto/:id      -- Consultar boleto
POST /api/remessa         -- Gerar arquivo remessa
POST /api/retorno         -- Processar arquivo retorno
GET  /api/bancos          -- Listar bancos suportados
```

**PIXCD:**
```
POST /api/cob             -- Criar cobranca
GET  /api/cob/:txid       -- Consultar cobranca
PATCH /api/cob/:txid      -- Alterar cobranca
POST /api/pix             -- Consultar PIX
GET  /api/pix/:e2eid      -- Consultar PIX por e2eid
POST /api/webhook         -- Configurar webhook
POST /api/qrcode          -- Gerar QR Code
```

**SAT:**
```
POST /api/venda           -- Enviar venda
POST /api/cancelamento    -- Cancelar venda
GET  /api/status          -- Status do SAT
GET  /api/info            -- Informacoes do SAT
```

### handlers.go
Handlers que instanciam o componente e chamam os metodos.

### frontend/index.html
SPA HTML com:
- Formularios para cada endpoint
- Botoes de acao
- Area de resposta (JSON pretty-print)
- Indicador de status
- Estilizacao moderna (CSS puro, sem framework)

```html
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenFiscalBR - Demo <PKG></title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <header>
        <h1>OpenFiscalBR - <PKG></h1>
        <div id="status-indicator"></div>
    </header>
    <main>
        <!-- forms para cada endpoint -->
    </main>
    <section id="response">
        <h2>Resposta</h2>
        <pre id="response-body"></pre>
    </section>
    <script src="app.js"></script>
</body>
</html>
```

### frontend/app.js
```javascript
// Funcoes para chamar cada endpoint via fetch()
// Pretty-print das respostas JSON
// Indicador de status do servico
```

### frontend/style.css
Estilizacao limpa e funcional.

### Dockerfile
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
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

### docker-compose.yml
```yaml
version: '3.8'
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

### README.md
Instrucoes de uso: como buildar, como rodar com Docker, como acessar o frontend.

## PASSO 7: Modo atualizacao (quando package Go ja existe)

1. Leia o arquivo `.openfiscalbr-meta.json` para obter os hashes anteriores do componente.
2. Compute os hashes SHA256 atuais de todos os `.pas` do componente Delphi.
3. Compare com os hashes armazenados:
   - Se TODOS os hashes sao iguais -> informe "Nenhuma alteracao detectada nos fontes Delphi." e encerre.
   - Se ha diferencas -> continue.
4. Para cada arquivo `.pas` que mudou:
   - Leia o conteudo atual do `.pas`
   - Leia o arquivo `.go` correspondente
   - Identifique as diferencas:
     - Novos tipos/classes adicionados no Delphi
     - Tipos/classes removidos do Delphi
     - Campos/properties adicionados ou removidos em classes existentes
     - Novos valores em enums
     - Metodos adicionados ou removidos
     - Mudancas de assinatura (tipos de parametros)
5. Reporte as diferencas ao usuario em formato claro:
   ```
   ## Alteracoes detectadas em ACBrNFe:

   ### Arquivos Delphi modificados:
   - ACBrNFe.Classes.pas (hash mudou)
   - ACBrNFe.Conversao.pas (hash mudou)

   ### Tipos novos:
   - TNewType em ACBrNFe.Classes.pas

   ### Campos novos:
   - TCampo.NovoField: String em ACBrNFe.Classes.pas

   ### Enum values novos:
   - TipoEmissao: teNovoTipo em ACBrNFe.Conversao.pas
   ```
6. Pergunte ao usuario se deseja aplicar as atualizacoes.
7. Se sim, aplique as mudancas nos arquivos Go:
   - PRESERVE qualquer bloco marcado com `// CUSTOM:` nos arquivos Go
   - Adicione novos tipos/campos/enum values
   - Marque tipos removidos com `// DEPRECATED: removido do Delphi em <data>`
8. Atualize `.openfiscalbr-meta.json` com os novos hashes.

## PASSO 8: Verificar compilacao

1. Execute `go build ./packages/<pkg>/...` e verifique se compila sem erros.
2. Execute `go vet ./packages/<pkg>/...` para verificar problemas.
3. Se houver erros de compilacao, corrija iterativamente:
   - Imports faltando -> adicione
   - Tipos nao encontrados -> verifique se a dependencia foi convertida
   - Metodos com assinatura errada -> corrija
4. Atualize `.openfiscalbr-meta.json` com os dados da conversao:
   ```json
   {
     "components": {
       "<ComponentName>": {
         "goPackage": "<pkg>",
         "delphiSourcePath": "<DelphiSourcePath>",
         "lastConvertedAt": "<timestamp ISO 8601>",
         "delphiFileHashes": { "<arquivo.pas>": "sha256:<hash>" },
         "goFiles": ["<arquivo1.go>", "<arquivo2.go>"]
       }
     }
   }
   ```

## NOTAS IMPORTANTES

- NAO converta arquivos de reports visuais (DANFE, DACTE, etc.) -- eles dependem de engines graficas Delphi
- NAO converta arquivos `*Reg.pas` -- sao registradores de design-time
- IGNORE blocos `{$IFDEF}` que referenciam GUI (FMX, VCL, LCL) ou plataformas especificas
- PRESERVE campos com nomes fiscais em PT-BR nos structs Go (mapeiam para XML da SEFAZ)
- Use `encoding/xml` struct tags para garantir compatibilidade com os schemas XML oficiais
- Para componentes com muitas implementacoes (Boleto: 70+ bancos, PIXCD: 20+ PSPs), pergunte ao usuario quais implementacoes deseja converter ou se quer todas
