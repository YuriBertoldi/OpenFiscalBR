# packages/cte

Porte do componente **ACBrCTe** — Conhecimento de Transporte Eletrônico.

| | |
|---|---|
| Componente Delphi | `ACBrDFe/ACBrCTe/` |
| Modelos | 57 (CT-e), 64 (GTV-e), 67 (CT-e OS), 57 (CT-e Simplificado) |
| Leiaute | 2.00, 3.00, 4.00 |
| Estado | etapa 1 de 11 — tipos e conversão |

## Quatro documentos, um componente

A diferença estrutural em relação ao `packages/nfe` é que o CT-e não é um documento com dois
modelos: são **quatro leiautes distintos** atendidos pela mesma árvore de classes.

O **CT-e Simplificado compartilha o código 57** com o CT-e normal. O comentário do próprio
ACBr (`ACBrCTe.Conversao.pas:106`) explica: *"O código correto do modelo CTe Simplificado é 57,
mas foi alterado para 58"* — internamente há um 58 para distingui-los, e o que sai no documento
é 57.

A consequência aparece na conversão: `ModeloCTeToStr` tem quatro entradas e `StrToModeloCTe`
tem **três**. Ler `"57"` devolve sempre `MoCTe`, nunca `MoCTeSimp` — e não há como fazer a ida
e volta fechar, porque a informação que separa os dois não está no código. Quem precisa
distinguir olha o `tpCTe`, onde o Simplificado tem valor próprio (`TcCTeSimp`).

## Seis modais

`rodoviário`, `aéreo`, `aquaviário`, `ferroviário`, `dutoviário` e `multimodal` — árvores de
classe separadas e mutuamente exclusivas. É o análogo do grupo `det` da NF-e em peso, mas com
seis variantes em vez de uma. Entram nas etapas 2b, 4b e 9b.

## Origem dos enums: `PCNCTe/pcteConversaoCTe.pas`

O ACBrCTe declara os **mesmos 37 enums em duas units**, e elas divergem:

- `Base/ACBrCTe.Conversao.pas`
- `PCNCTe/pcteConversaoCTe.pas`

Vale a segunda, por duas evidências independentes:

1. **Resolução do `uses`.** As seis units vivas — `XmlWriter`, `XmlReader`, `IniWriter`,
   `IniReader`, `Classes` e `XmlHandler` — listam `pcteConversaoCTe`. A única menção a
   `ACBrCTe.Conversao` está em `XmlWriter.pas:46` e está **comentada**.
2. **A unit `Base/` perde membros.** O `TSchemaCTe` de lá não tem `schevVincPgto` nem
   `schevCancVincPgto` — o evento de Vínculo de Pagamento e o cancelamento dele, os dois mais
   recentes.

Travado por `TestSchemaCTe_MembrosExclusivosDaUnitPCN`.

## Armadilhas dos enums

As que o teste trava, porque nenhuma aparece em build, `vet` ou round-trip ingênuo:

| enum | o que surpreende |
|---|---|
| `TipoServico` | a tabela **pula o código `5`**: vai de `4` (multimodal) para `6` (transporte de pessoas). Preencher a lacuna desloca os quatro membros seguintes |
| `Retira` | `RtSim` é **zero** e `RtNao` é um. O campo responde *"o recebedor retira?"* com 0=Sim |
| `Lotacao` | o contrário — `LtNao` é zero. Os dois convivem no mesmo leiaute |
| `TipoDataPeriodo`, `TipoHorarioIntervalo` | o último código é `"-1"`, não vazio |
| `TipoFretamento`, `CRT`, `ClassDuto`, `TipoContratacao` | o **primeiro** membro tem código vazio |
| `CRT` | **não é o mesmo do `packages/nfe`**: lá o primeiro membro é `CrtSimplesNacional` com código `"1"`; aqui há um `CrtNenhum` vazio na frente e tudo desloca em um |
| `TipoDirecao`, `TipoPropriedade` | códigos em letra (`N`/`L`/`S`/`O`, `P`/`T`) — os dois únicos em que a comparação sem caixa do `StrToEnumerado` muda resultado |
| `ClassDuto`, `TipoContratacao` | o ACBr os lê com laço próprio e comparação **com** caixa, não pelo `StrToEnumerado` |
| `SchemaCTe` | o nome vem do **próprio identificador Pascal** (`GetEnumName` menos o prefixo `sch`), não de tabela. A `TSchemaCTeArrayStrings` tem 18 entradas vazias e é letra morta |

### `TpInfManu`: duas tabelas que não se falam

O caso mais grave, e é defeito do original:

| | códigos | membros cobertos |
|---|---|---|
| `String()` | `01`..`15`, `99` | 16 — inclui as Packing Instructions `ImPI965`..`ImPI970` |
| `StringV2()` | `1`..`9`, `99` | 10 — **os seis `ImPI*` não têm código** |

O ACBr usa as duas em lados opostos:

```
escrita   XmlWriter.pas:3755, IniWriter.pas:1202   -> TpInfManuToStrV2   ("1")
leitura   XmlReader.pas:203, XmlHandler.pas:3420,
          IniReader.pas:1551                       -> StrToTpInfManu    ("01")
```

Ele **grava `1` e procura `01`**: um documento gerado pelo ACBr e relido pelo ACBr perde o
campo, e seis membros não têm como ser emitidos.

As duas formas ficam expostas aqui porque as duas existem no mundo. **Qual delas o gerador
emite é decisão da etapa do gerador**, não desta; o que esta etapa garante é que nenhuma se
perca — e `ParseTpInfManuV2` recusa string vazia, que de outro modo casaria com `ImPI965`.

## Divergências deliberadas

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `StrToSchemaCTe` | devolve `TSchemaCTe(-1)` para nome desconhecido **e ainda marca `ok := True`** — um enum fora de faixa que se propaga em silêncio | devolve `SchErro` mais um `error`. Mesma decisão do `pcn.ParseSchemaDFe` na Distribuição DF-e da NF-e |
| 2 | `StrToXxx` em geral | `raise EACBrException` | devolve o membro zero + `error`, pela convenção de erros do projeto |

## Como rodar os testes

```bash
go test ./packages/cte/ -v
```

Sem dependência externa e sem rede.

### Conferência mecânica

```bash
py ferramentas/conferir-enums-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrCTe/PCNCTe/pcteConversaoCTe.pas" \
   packages/cte/conversao.go packages/cte/types.go \
   --par TpcteTipoCTe=tipoCTeCodigos --par TpInfManu=tpInfManuCodigos
```

**29 enums conferidos, zero divergência.** O único aviso que resta é o `TpInfManuV2` descrito
acima — tabela de código divergindo de tabela de código dentro do próprio ACBr.

**13 cortes deliberados**, todos pegos por algum teste.

## Roteiro

| # | Etapa | Estado |
|---|---|---|
| 1 | `types.go` + `conversao.go` | **nesta entrega** |
| 2a | `classes.go` — cabeçalho, prestação, documentos | pendente |
| 2b | `classes_modais.go` — os seis modais | pendente |
| 3 | `xml_reader.go` | pendente |
| 4a | `xml_writer.go` — cabeçalho e prestação | pendente |
| 4b | `xml_writer_modais.go` | pendente |
| 5 | regras de negócio e URLs | pendente |
| 6a | web services — status e consulta | pendente |
| 6b | web services — recepção, inutilização, cadastro | pendente |
| 7 | eventos | pendente |
| 8 | formato `.ini` — cabeçalho e prestação | pendente |
| 9 | formato `.ini` — modais | pendente |
| 10 | demo em `demos/cte` | pendente |
