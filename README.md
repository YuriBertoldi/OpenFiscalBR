# OpenFiscalBR

Port nativo em **Go** do [Projeto ACBr](http://projetoacbr.com.br) — a principal biblioteca open-source de automacao fiscal brasileira, originalmente escrita em Delphi/Lazarus.

Licenciado sob **LGPL 2.1+** (obra derivada do Projeto ACBr).

## Status dos Packages

| Package | Descricao | Status |
|---------|-----------|--------|
| `packages/comum` | Base: TXTClass, utils, erros, FormatFloatBR | Completo |
| `packages/sped` | SPED Fiscal (EFD-ICMS/IPI): tipos, registros, blocos, geracao TXT | Completo |
| `packages/pcn` | Gerador/Leitor XML (PCNComum) | Pendente |
| `packages/dfe` | DFe base, SSL, config | Pendente |
| `packages/nfe` | NFe (Nota Fiscal Eletronica) | Pendente |
| `packages/cte` | CTe (Conhecimento de Transporte Eletronico) | Pendente |
| `packages/mdfe` | MDFe (Manifesto Eletronico de Documentos Fiscais) | Pendente |
| `packages/boleto` | Boleto bancario | Pendente |
| `packages/pixcd` | PIX (pagamentos instantaneos) | Pendente |
| `packages/sat` | SAT Fiscal (CF-e) | Pendente |
| `packages/nfsex` | NFSe (Nota Fiscal de Servicos Eletronica) | Pendente |

## Status das Demos

| Demo | Descricao | Status |
|------|-----------|--------|
| `demos/sped` | API REST + Frontend para geracao de arquivo SPED Fiscal | Completo |

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
  comum/         Layer 0: base, TXTClass, validadores, utils
  sped/          Layer 0: SPED Fiscal (EFD-ICMS/IPI)
  pcn/           Layer 1: gerador/leitor XML (PCNComum)
  dfe/           Layer 2: DFe base, SSL, config
  nfe/           Layer 3: NFe
  ...
demos/
  sped/          Demo SPED Fiscal (API REST + Frontend + Docker)
```

## Convertendo novos componentes

Use a skill `/convert` para portar componentes Delphi para Go:

```
/convert ACBrNFe C:\MeusACBr\Fontes
```

Veja `.claude/skills/convert.md` para detalhes do workflow.

## Ordem de Dependencias

A conversao deve respeitar esta ordem — nunca converter um componente sem que suas dependencias existam:

1. **Layer 0**: `comum` (sem dependencias) e `sped` (depende de `comum`)
2. **Layer 1**: `pcn` (depende de `comum`)
3. **Layer 2**: `dfe` (depende de `comum`, `pcn`)
4. **Layer 3**: componentes finais (`nfe`, `cte`, `mdfe`, `boleto`, `pixcd`, `sat`, `nfsex`)

## Licenca

LGPL 2.1 ou posterior. Veja [LICENSE.TXT](LICENSE.TXT).

Este projeto e uma obra derivada do [Projeto ACBr](http://projetoacbr.com.br), conforme permitido pela Secao 2 da LGPL 2.1.

## Creditos

- **Projeto ACBr** — Daniel Simoes de Almeida e colaboradores
- Veja [NOTICE.md](NOTICE.md) para creditos completos
