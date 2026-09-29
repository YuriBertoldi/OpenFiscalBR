# Tipos de campo do leitor de DFe (TACBrTipoCampo → Go)

Fonte da verdade no Delphi: `ObterConteudoTag` em `ACBrDFe/ACBrXmlBase.pas`. Equivalentes Go
em `packages/pcn/tipos_campo.go`. **Precisão errada não quebra build, `vet` nem teste ingênuo**
— é a fonte de erro silencioso número um de um porte de leitor.

## Tabela de correspondência

| Delphi | Go (`packages/pcn`) | Semântica exata |
|---|---|---|
| `tcStr`, `tcEsp`, `tcStrOrig`, `tcNumStr` | `ConteudoStr(n)` | `Trim` do texto; node nil ⇒ `""` |
| `tcInt` | `ConteudoInt(n)` | **`OnlyNumber` ANTES de converter**: `"3-5"` ⇒ 35, `"n35"` ⇒ 35; vazio/ilegível ⇒ 0. Não é `Atoi`. |
| `tcInt64` | `ConteudoInt64(n)` | idem, 64 bits |
| `tcDe1`..`tcDe8`, `tcDe10` | `ConteudoDe1(n)`..`ConteudoDe10(n)` (ou `ConteudoDec(n, casas)`) | decimal com ponto (`StringToFloatDef`, 0 no erro). O nº de casas só muda o resultado com `Node.FloatIsIntString` ligado ("1234" com De2 ⇒ 12,34) — falso em todo XML da SEFAZ, mas o flag existe e os ports devem declarar a precisão correta. |
| `tcDat`, `tcDatHor` | `ConteudoDataDef(n)` (zero no erro) ou `ConteudoData(n)` (com erro) | aceita `AAAA-MM-DD`, `AAAAMMDD`, `AAAAMM` (⇒ dia 1), `dd/mm/aaaa`, `...Thh:mm:ss±hh:mm`, sufixo `Z`, fração de segundo. **Fuso preservado como `time.FixedZone`** (o ACBr descarta — hora de parede fica idêntica). |
| `tcHor` | `ConteudoHora(n)` | `hh:mm[:ss]` ⇒ `time.Duration` desde a meia-noite |
| `tcBool`, `tcBoolStr` | `ConteudoBool(n)` | só `"true"` (qualquer caixa) é verdadeiro; `"1"` é FALSO |
| `ObterConteudoTagCNPJCPF` | `ConteudoCNPJCPF(n)` | lê `CNPJ`; vazio ⇒ lê `CPF`. Usa `Find` estrito, como o original (`ConteudoCNPJCPFAnyNs` é a variante tolerante). |
| atributo via `Attributes.Items['x']` | `n.Attr("x")` / `AtributoInt(n, "x")` | vazio/0 quando ausente. O ACBr às vezes faz `StrToInt` direto no atributo — isso levanta exceção lá; aqui é 0 com `DIVERGENCIA` documentada. |

## Competência `AAAAMM`

Dois padrões no ACBr, os dois cobertos por `pcn`:

- `Copy`+`EncodeDate` com validação (`CompetEmis`/`CompetApur` da NFGas): use
  `pcn.ParseCompetenciaDef(s)` — fora da faixa (`len≠6`, ano 0, mês fora de 1..12) ⇒ tempo zero,
  sem erro.
- `'01/'+MM+'/'+AAAA` + `StrToDate` (`CompetFat`): **dependente de locale no Delphi** — use o
  mesmo `ParseCompetenciaDef` e marque `DIVERGENCIA`.

## Enum a partir de tag

- `StrToXxx(ObterConteudo(..., tcStr))` ⇒ `ParseXxx(pcn.ConteudoStr(...))`; o `raise` do Delphi
  vira `(T, error)`, e o leitor usa `v, _ =` para não abortar a nota (documente).
- `StrToXxx(out ok, ...)` (padrão `EnumeradoToStr`) ⇒ idem; o sentinela `TXxx(-1)` do Delphi
  vira membro `-1` explícito no Go (`DbisNenhum`, `MdiNenhum`).
- **`ObterConteudo(..., tcInt)` atribuído a property de enum é BUG do ACBr**: o código do
  leiaute é tratado como ordinal (código "1" vira o 2º membro; códigos altos como 16/90 estouram
  o enum). Parse pelo código, `DIVERGENCIA`, teste. Caso real: `motDesICMS` na NFGas.

## `Find` × `FindAnyNs` (TACBrXmlNodeList)

| Delphi | Compara | Acha `<Signature>` | Acha `<ds:Signature>` |
|---|---|---|---|
| `Find` | `Node.Name` — nome **com prefixo** | sim | **não** |
| `FindAnyNs` | `Node.LocalName` | sim | sim |

O leitor de cada componente mistura os dois (na NFGas: raiz/`ide`/`ISUFEmit`/`qrCodNFGas`/
`Signature`/`gTarif`/`gHistCons`/`gCons` usam `Find`; o resto, `FindAnyNs`). Copie a escolha
por campo — trocar muda o comportamento com XML prefixado.
