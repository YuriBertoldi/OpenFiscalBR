# OpenFiscalBR

Port nativo em **Go** do [Projeto ACBr](http://projetoacbr.com.br) — a principal biblioteca open-source de automacao fiscal brasileira, originalmente escrita em Delphi/Lazarus.

Licenciado sob **LGPL 2.1+** (obra derivada do Projeto ACBr).

## Status dos Packages

| Package | Descricao | Status |
|---------|-----------|--------|
| `packages/comum` | Base: TXTClass, utils, erros, FormatFloatBR | Completo |
| `packages/sped` | SPED Fiscal (EFD-ICMS/IPI): tipos, registros, blocos, geracao TXT | Completo (270 registros com writer) |
| `packages/pcn` | Base de XML de DFe: mini-DOM, tipos de campo, enums DFe, validador CNPJ/CPF, chave de acesso, ini, construtor de XML (AddNode) | Completo (leitura e geracao) |
| `packages/rtc` | Reforma Tributaria (IBS/CBS/IS) compartilhada pelos DFe | Classes + leitor + gerador XML completos; ini parcial (sem monofasia) |
| `packages/dfe` | DFe base: certificado A1 (PFX/PEM), assinatura XMLDSig (C14N 1.0), cliente SOAP 1.2 com TLS mutuo | Completo |
| `packages/nfe` | NFe (Nota Fiscal Eletronica) | Pendente |
| `packages/cte` | CTe (Conhecimento de Transporte Eletronico) | Pendente |
| `packages/mdfe` | MDFe (Manifesto Eletronico de Documentos Fiscais) | Pendente |
| `packages/boleto` | Boleto bancario | Pendente |
| `packages/pixcd` | PIX (pagamentos instantaneos) | Pendente |
| `packages/sat` | SAT Fiscal (CF-e) | Pendente |
| `packages/nfsex` | NFSe (Nota Fiscal de Servicos Eletronica) | Pendente |
| `packages/nfgas` | NFGas (NF de Gas Canalizado, modelo 76) | Completo: leitura, geracao, assinatura, transmissao a SEFAZ (recepcao sincrona, consulta, status, cancelamento), QR-Code |

## Status das Demos

| Demo | Descricao | Status |
|------|-----------|--------|
| `demos/sped` | API REST + Frontend para geracao de arquivo SPED Fiscal | Completo |
| `demos/nfgas` | API REST + Frontend para NFGas: leitura, importacao em lote, geracao, assinatura e transmissao (certificado via env) | Completo |

## Inicio Rapido

### Pre-requisitos

- Go 1.22+
- Docker (opcional, para demos)

### Usando o package SPED programaticamente

```go
package main

import (
    "fmt"
    "time"

    "github.com/openfiscalbr/openfiscalbr/packages/sped"
)

func main() {
    fiscal := sped.NewSPEDFiscal()
    fiscal.Path = "."
    fiscal.Arquivo = "SpedFiscal.txt"
    fiscal.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
    fiscal.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

    r := fiscal.Bloco0.Registro0000
    r.CodVer = sped.VlVersao115
    r.CodFin = sped.CodFinOriginal
    r.DtIni = fiscal.DtIni
    r.DtFin = fiscal.DtFin
    r.Nome = "MINHA EMPRESA LTDA"
    r.CNPJ = "12345678000199"
    r.UF = "SP"
    r.IE = "123456789"
    r.CodMun = 3550308
    r.IndPerfil = sped.PerfilA
    r.IndAtiv = sped.AtivOutros

    fiscal.Bloco0.Registro0001.IndDad = 0

    if err := fiscal.SaveFileTXT(); err != nil {
        fmt.Println("Erro:", err)
        return
    }
    fmt.Println("Arquivo SPED gerado com sucesso!")
}
```

### Lendo NFGas em lote programaticamente

```go
package main

import (
    "fmt"

    "github.com/openfiscalbr/openfiscalbr/packages/nfgas"
)

func main() {
    notas, err := nfgas.LerLoteDiretorio(`C:\xmls\nfgas`)
    if err != nil {
        fmt.Println("Documentos com problema:", err) // as notas legiveis vem mesmo assim
    }
    for _, n := range notas {
        fmt.Println(n.ChaveAcesso(), n.Situacao(), n.NFGas.Total.VNF)
    }
}
```

### Rodando a demo SPED

```bash
cd demos/sped
go run .
# Acesse http://localhost:8080
```

Com Docker:

```bash
docker compose -f demos/sped/docker-compose.yml up --build
```

## Rodando os testes

```bash
go test ./packages/... -v
```

## Estrutura do Projeto

```
packages/
  comum/         Layer 0: base, TXTClass, utils
  sped/          Layer 0: SPED Fiscal (EFD-ICMS/IPI)
  pcn/           Layer 1: base de XML de DFe (mini-DOM, tipos de campo, validador, chave, ini)
  rtc/           Layer 1: Reforma Tributaria (IBS/CBS) compartilhada pelos DFe
  dfe/           Layer 2: certificado A1, assinatura XMLDSig, cliente SOAP SEFAZ
  nfe/           Layer 3: NFe
  nfgas/         Layer 3: NFGas (leitura e emissao)
  ...
demos/
  sped/          Demo SPED Fiscal (API REST + Frontend + Docker)
  nfgas/         Demo NFGas: leitura, lote, geracao, assinatura e transmissao (API REST + Frontend + Docker)
ferramentas/     Scripts de apoio ao port (nao fazem parte do modulo Go)
```

## Skills do projeto

O repositorio traz skills do Claude Code em `.claude/skills/`, carregadas automaticamente ao
abrir uma sessao na raiz do projeto:

| Skill | Para que serve |
|-------|----------------|
| `/convert <Componente> <FontesDelphi>` | Porta um componente ACBr de Delphi para Go |
| `/validar-package [pkg]` | gofmt, build, vet, testes, header de licenca, module path |
| `/adicionar-registro-sped <registro>` | Acrescenta um registro ao `packages/sped` |
| `/gerar-demo <pkg>` | Gera a demo em `demos/<pkg>/` |
| `/sincronizar-meta` | Atualiza `.openfiscalbr-meta.json` e o status dos packages |
| `/portar-leitor-dfe <Componente>` | Porta o leitor de XML de um DFe com fidelidade auditavel |

Ha tambem o subagente `revisor-go` (`.claude/agents/revisor-go.md`), que revisa codigo portado
quanto a fidelidade ao Delphi de origem e as convencoes do projeto.

Exemplo:

```
/convert ACBrNFe C:\MeusACBr\Fontes
```

Veja `.claude/skills/convert/SKILL.md` para detalhes do workflow.

## Ordem de Dependencias

A conversao deve respeitar esta ordem — nunca converter um componente sem que suas dependencias existam:

1. **Layer 0**: `comum` (sem dependencias) e `sped` (depende de `comum`)
2. **Layer 1**: `pcn` (depende de `comum`) e `rtc` (depende de `pcn`)
3. **Layer 2**: `dfe` (depende de `comum`, `pcn`)
4. **Layer 3**: componentes finais (`nfe`, `cte`, `mdfe`, `boleto`, `pixcd`, `sat`, `nfsex`,
   `nfgas` — depende de `comum`, `pcn`, `rtc` e `dfe`)

## Licenca

LGPL 2.1 ou posterior. Veja [LICENSE.TXT](LICENSE.TXT).

Este projeto e uma obra derivada do [Projeto ACBr](http://projetoacbr.com.br), conforme permitido pela Secao 2 da LGPL 2.1.

## Creditos

- **Projeto ACBr** — Daniel Simoes de Almeida e colaboradores
- Veja [NOTICE.md](NOTICE.md) para creditos completos
