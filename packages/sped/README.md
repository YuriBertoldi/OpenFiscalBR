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

Um registro so e efetivamente gerado quando tem **writer**. Structs declarados sem writer
podem ser preenchidos pelo consumidor e **nunca aparecem no arquivo de saida** — por isso a
tabela abaixo separa as duas colunas.

| Bloco | Structs | Writers | Situacao |
|-------|---------|---------|----------|
| 0 | 23 | 22 | Completo (falta apenas o 0002, hoje escrito inline dentro do writer do 0001) |
| B | 13 | 13 | Completo |
| C | 27 | 11 | Parcial |
| D | 33 | 4 | Parcial |
| E | 26 | 6 | Parcial |
| G | 7 | 7 | Completo |
| H | 7 | 7 | Completo |
| K | 23 | 23 | Completo |
| 1 | 35 | 3 | Apenas abertura, 1010 e encerramento |
| 9 | 4 | 4 | Completo |
| **Total** | **198** | **100** | **98 structs sem writer** |

Registros com writer, por bloco:

- **Bloco 0**: 0000, 0001, 0005, 0015, 0100, 0150, 0175, 0190, 0200, 0205, 0206, 0210, 0220, 0221, 0300, 0305, 0400, 0450, 0460, 0500, 0600, 0990
- **Bloco B**: B001, B020, B025, B030, B035, B350, B420, B440, B460, B470, B500, B510, B990
- **Bloco C**: C001, C100, C101, C110, C111, C112, C113, C114, C170, C190, C990
- **Bloco D**: D001, D100, D190, D990
- **Bloco E**: E001, E100, E110, E111, E116, E990
- **Bloco G**: G001, G110, G125, G126, G130, G140, G990
- **Bloco H**: H001, H005, H010, H011, H020, H030, H990
- **Bloco K**: K001, K010, K100, K200, K210, K215, K220, K230, K235, K250, K255, K260, K265, K270, K275, K280, K290, K291, K292, K300, K301, K302, K990
- **Bloco 1**: 1001, 1010, 1990
- **Bloco 9**: 9001, 9900 (contagem por tipo de registro), 9990, 9999

Alem dos 98 structs sem writer, faltam structs que nem existem: `C130`, `C140`, `C160` e
`C165`, que no ACBr sao filhos de `TRegistroC100`.

Para implementar um registro, use a skill `/adicionar-registro-sped`.

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
- Vigencia dos blocos B, G e K (periodos anteriores nao geram o bloco)
- Contagem do Registro 9900 a partir dos contadores, incluindo registro neto (0175)

## Divergencias conhecidas vs ACBr Delphi

Rastreamento contra `ACBrSPEDFiscal` (Delphi XE3). Corrigir antes de considerar o package
pronto para producao.

| # | Divergencia | Local Go | Referencia ACBr |
|---|-------------|----------|-----------------|
| 1 | `CheckRegistroFunc` exposto so no C100 (`OnCheckRegistroC100`); os demais registros nao tem callback | `bloco_writers.go` | `ACBrEFDBloco_C_Events.pas` |
| 2 | Sem construtores `RegistroXXXNew` (ponto unico de criacao e validacao de hierarquia) | — | `ACBrEFDBloco_C_Class.pas:242-318` |
| 3 | 98 structs sem writer; faltam tambem os structs C130, C140, C160 e C165 | ver tabela de cobertura acima | — |
| 4 | Campo morto `Registro9001.Registro9900` (zero referencias) | `bloco_9.go` | — |
| 5 | Registro 0002 escrito inline dentro do writer do 0001, em vez de writer proprio — por isso nao aparece no 9900 | `write_bloco_0.go` (`WriteRegistro0001`) | — |
| 6 | Writers de dados concentrados em `write_blocos.go`; o ACBr usa uma unit por bloco | `write_blocos.go` | `ACBrEFDBloco_<X>_Class.pas` |

### Auditoria campo a campo

Os writers dos blocos **0, B, G, H e K** estao completos e conferidos campo a campo contra o
`.pas` de origem. Os blocos **C, D e E** tem os writers existentes conferidos, mas ainda com
registros por implementar. O bloco **1** so tem 1001, 1010 e 1990.

> **Nota sobre os blocos G e H.** Mesmo problema do Bloco B: os structs eram em boa parte
> inventados. No G, `RegistroG110` nao tinha `MODO_CIAP` nem `SALDO_FN_ICMS`, o `RegistroG125`
> usava nomes de outro registro, e `G126`/`G140` tinham campos que nao existem no layout. No H,
> `RegistroH011` era copia do `H020` -- o registro real tem um unico campo, `CNPJ`. Todos
> reescritos a partir do `.pas` e cobertos por teste.

> **Nota sobre o Bloco B.** O bloco foi reescrito para ficar identico ao ACBr: os registros
> B020, B350, B420, B440, B460, B470 e B500 tinham campos divergentes ou trocados, e B030,
> B035 e B510 nao existiam. A hierarquia tambem estava errada -- o B500 e filho do B001, nao
> do B470.
>
> Um ponto merece destaque: no ACBr o `WriteRegistroB350` emite a linha com o literal
> **`B035`**, e nao `B350`. Aparentemente e um erro de copia no Delphi, ja que o registro B035
> tem outro conjunto de campos. O comportamento foi mantido por fidelidade ao original e esta
> coberto por teste (`TestWriteRegistroB350_EmiteLiteralB035ComoNoACBr`), para que a escolha
> seja explicita e nao se perca numa refatoracao futura.

Ja corrigidas, todas cobertas por teste de regressao:

- Vigencia dos blocos B, G e K — fora do periodo o bloco nao e escrito.
- Registro 9900 alimentado pelos contadores, inclusive registro neto.
- Propagacao de `DT_INI`/`DT_FIN` aos blocos (o Registro 0002 nunca era emitido quando o
  consumidor atribuia `DtIni` direto, como faz o exemplo do README).
- `Checkf` em `comum`, validacao de CHV_NFE obrigatoria para NF-e modelo 55 e validacao de
  COD_NAT_CC / IND_CTA no registro 0500 — as tres unicas validacoes que o ACBr tem ativas
  para os registros ja portados.
- `IND_FRT` e `IND_PGTO` com as tabelas por vigencia (`StringEm`).
- C100 com tratamento de documento cancelado/denegado/inutilizado e variante NFC-e (modelo 65).
- Campos condicionais por vigencia ou versao de leiaute, todos perdidos no port original:
  `COD_MUN = 9999999` para participante do exterior (0150), `CEST` so a partir de 2017 (0200),
  `COD_BARRA` so acima da versao 114 (0220), `VL_ABAT_NT` so acima da versao 111 (C170),
  `COD_MUN_ORIG`/`COD_MUN_DEST` so a partir de 2018 (D100) e `VL_ITEM_IR` so a partir de
  2015 (H010).

### Divergencias deliberadas em relacao ao ACBr

Dois pontos em que o port **nao** reproduz o Delphi, por serem defeitos do original:

1. `IND_FRT`/`IND_PGTO` — no ACBr os `case` de cada faixa de vigencia nao cobrem todos os
   membros, e a variavel `strIND_FRT`/`strIND_PGTO`, declarada fora do laco, conserva o valor
   da iteracao anterior, emitindo o codigo de outro documento. Aqui, valor inaplicavel na
   vigencia sai vazio.
2. Callback de validacao — no ACBr a flag `booAborta` e inicializada uma unica vez antes do
   laco, entao vetar um documento derruba todos os seguintes. Aqui a flag e reavaliada a cada
   registro.
