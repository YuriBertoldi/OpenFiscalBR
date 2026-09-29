# Package sped

Package SPED Fiscal (EFD-ICMS/IPI) do OpenFiscalBR, equivalente ao `ACBrSPEDFiscal` do Projeto ACBr.
Gera arquivos digitais no formato texto pipe-delimited conforme o **Guia Pratico da EFD-ICMS/IPI** publicado pela Receita Federal do Brasil.

## Arquitetura

```
SPEDFiscal (orquestrador)
  ├── Bloco0  (Abertura, Identificacao e Referencias)
  ├── BlocoB  (Escrituracao e Apuracao do ISS) — a partir de 2019
  ├── BlocoC  (Documentos Fiscais I - Mercadorias)
  ├── BlocoD  (Documentos Fiscais II - Servicos)
  ├── BlocoE  (Apuracao do ICMS e do IPI)
  ├── BlocoG  (Controle do Credito de ICMS do Ativo Permanente) — a partir de 2011
  ├── BlocoH  (Inventario Fisico)
  ├── BlocoK  (Controle da Producao e do Estoque) — a partir de 2016
  ├── Bloco1  (Outras Informacoes)
  └── Bloco9  (Controle e Encerramento do Arquivo Digital)
```

## Uso Basico

```go
fiscal := sped.NewSPEDFiscal()
fiscal.Path = "/tmp"
fiscal.Arquivo = "SpedFiscal.txt"
fiscal.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
fiscal.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

// Preencher Registro 0000 (obrigatorio)
r := fiscal.Bloco0.Registro0000
r.CodVer = sped.VlVersao115
r.CodFin = sped.CodFinOriginal
r.DtIni = fiscal.DtIni
r.DtFin = fiscal.DtFin
r.Nome = "EMPRESA LTDA"
r.CNPJ = "12345678000199"
r.UF = "SP"
r.IE = "123456789"
r.CodMun = 3550308
r.IndPerfil = sped.PerfilA
r.IndAtiv = sped.AtivOutros

// Marcar bloco 0 com dados
fiscal.Bloco0.Registro0001.IndDad = 0

// Gerar arquivo
err := fiscal.SaveFileTXT()
```

## Tipos Enumerados

Todos os enums implementam `String()` retornando o codigo SPED esperado:

| Tipo | Exemplo | Saida |
|------|---------|-------|
| `VersaoLeiauteFiscal` | `VlVersao115.String()` | `"016"` |
| `CodFin` | `CodFinOriginal.String()` | `"0"` |
| `IndPerfil` | `PerfilA.String()` | `"A"` |
| `CodSit` | `SitCancelado.String()` | `"02"` |
| `CstIcms` | `CstIcmsTributadaIntegralmente.String()` | `"000"` |
| `CstPis` | `CstPisValorAliquotaNormal.String()` | `"01"` |
| `CstIpi` | `CstIpiSaidaTributada.String()` | `"50"` |

## Registros Implementados

### Bloco 0 (22 registros)
0000, 0001, 0002, 0005, 0015, 0100, 0150, 0175, 0190, 0200, 0205, 0206, 0210, 0220, 0221, 0300, 0305, 0400, 0450, 0460, 0500, 0600, 0990

### Bloco B
B001, B020, B025, B350, B420, B440, B460, B470, B500, B990

### Bloco C
C001, C100, C170, C190, C990

### Bloco D
D001, D100, D190, D990

### Bloco E
E001, E100, E110, E990

### Blocos G, H, K, 1
Registros de abertura e encerramento + registros de dados principais

### Bloco 9
9001, 9900 (contagem por tipo de registro), 9990, 9999

## Validacoes per Legislacao

- `DT_INI` deve ser o primeiro dia do mes
- `DT_FIN` deve ser o ultimo dia do mes
- `DT_FIN >= DT_INI`
- Blocos escritos na ordem obrigatoria: 0 → B → C → D → E → G → H → K → 1 → 9
- Arquivo usa CRLF como terminador de linha
- Todos os campos delimitados por `|` (pipe)
- Registro 9999 contem a contagem total de linhas do arquivo
- Valores monetarios usam virgula como separador decimal

## Testes

```bash
go test ./packages/sped/... -v
```

Os testes validam:
- Formato pipe-delimited de cada linha
- Codigos dos enums conforme tabelas da legislacao (CST ICMS, PIS, COFINS, IPI)
- Ordem dos blocos no arquivo gerado
- Contagem de linhas no Registro 9999
- Terminadores CRLF
- Registros obrigatorios presentes
- Validacao de datas (primeiro/ultimo dia do mes, ano bissexto)
