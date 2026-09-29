# packages/pcn

Base de leitura de XML compartilhada por todos os documentos fiscais eletrônicos do
OpenFiscalBR (NFe, CTe, MDFe, NF3e, NFCom, NFGas...).

**Layer 1.** Depende apenas de `packages/comum` e da biblioteca padrão do Go — zero dependência
externa.

## Componente Delphi correspondente

| Unit ACBr | O que veio para cá |
|---|---|
| `ACBrLibXML2/ACBrXmlDocument.pas` | `Document`, `Node`, `Find`, `FindAnyNs`, `FindAll`, `FindAllAnyNs`, `OuterXML` |
| `ACBrDFe/ACBrXmlBase.pas` | `ConteudoStr/Int/Dec/Data/Bool`, `ConteudoCNPJCPF`, `Signature`, `LerSignature`, `NormatizarBoolean`, `RemoverCDATA` |
| `ACBrDFe/ACBrDFe.Conversao.pas` | enums compartilhados: `TipoAmbiente`, `TipoEmissao`, `Indicador`, `IndicadorEx`, `OrigemMercadoria`, `CSTIcms`, `CSTPis`, `CSTCofins`, `TpPagAnt` |
| `ACBrDFe/Comum/ACBrDFeComum.Proc.pas` | `ProcDFe`, `LerInfProt` |
| `ACBrDFe/ACBrDFeUtil.pas` | `DigitoChaveAcesso`, `ValidarChaveAcesso`, `RemoverLiteralChave`, `ValidarAAMM`, `ValidarCodigoUF` |
| `ACBrDiversos/ACBrValidador.pas` | `ValidarCPF`, `ValidarCNPJ` (inclusive alfanumérico), `ValidarCNPJouCPF` |
| `ACBrComum/ACBrUtil.*` | `OnlyNumber`, `StringToFloat`, `StringDecimalToFloat`, `EncodeDataHora`, `ParseCompetencia` |
| `TMemIniFile` (RTL) | `INI` — leitor/gravador `.ini` mínimo, sem dependência externa |

## Por que um mini-DOM e não `xml.Unmarshal`

Os leitores do ACBr fazem coisas que struct tag não expressa: achatam `ICMS00..ICMS90` num
único struct, leem um campo do nó pai, aplicam precedência entre campos alternativos, trazem
dados em atributo de item de coleção. Um mini-DOM sobre `encoding/xml` permite portar cada
`Ler_*` do `.pas` linha a linha — que é o que torna a auditoria de fidelidade possível.

Toda navegação é segura em `nil`: encadear por um elemento inexistente devolve `nil`, e ler um
`nil` devolve o valor zero. No Delphi, o mesmo caminho é access violation.

```go
doc, err := pcn.ParseString(xml)
if err != nil {
    return err
}
ide := doc.Root.FindAnyNs("infNFGas").FindAnyNs("ide")

cUF   := pcn.ConteudoInt(ide.FindAnyNs("cUF"))          // OnlyNumber antes de converter
vProd := pcn.ConteudoDe2(ide.FindAnyNs("vProd"))        // precisão do leiaute
dhEmi := pcn.ConteudoDataDef(ide.FindAnyNs("dhEmi"))    // zero em vez de erro, como o ACBr
```

### `Find` × `FindAnyNs`

| Método | Compara | Acha `<Signature>` | Acha `<ds:Signature>` |
|---|---|---|---|
| `Find` | nome **qualificado** (`TACBrXmlNode.Name`) | sim | não |
| `FindAnyNs` | *local name* (`TACBrXmlNode.LocalName`) | sim | sim |

A diferença é fiel ao ACBr e importa: alguns pontos do leitor da NFGas usam `Find`, e num XML
com prefixo declarado eles realmente não encontram o campo.

## Tipos principais

| Tipo | Papel |
|---|---|
| `Document`, `Node`, `Atributo` | árvore XML somente leitura |
| `Signature` | dados da assinatura XMLDSig |
| `ProcDFe` | protocolo de autorização |
| `INI` | arquivo `.ini` em memória |
| `ErroLeitura` | erro com o caminho da tag (`infNFGas/ide/dhEmi`), compatível com `errors.Is`/`As` |

## Divergências deliberadas em relação ao Delphi

Documentadas no código, com teste cobrindo cada uma:

| Ponto | ACBr | Aqui | Por quê |
|---|---|---|---|
| Fuso horário em `dhEmi`/`dhRecbto` | descarta o offset | preserva como `time.FixedZone` | hora de parede idêntica campo a campo; deixa de perder a informação de fuso |
| Separador decimal | o do sistema operacional | sempre o último `.` ou `,` da string | mesmo resultado para XML de DFe, e para de variar conforme a máquina |
| `GravarFloat` no `.ini` | separador do sistema | sempre ponto (leitura aceita os dois) | idem |
| Navegação por elemento ausente | access violation | `nil` propagado | leitura em lote não pode derrubar o processo |

## Precisão dos campos decimais

`ConteudoDec(n, casas)` e os atalhos `ConteudoDe1`..`ConteudoDe10` correspondem a
`tcDe1`..`tcDe10`. No XML da SEFAZ o valor já vem com ponto decimal e `casas` documenta a
precisão do leiaute; o parâmetro só muda o resultado quando o node está marcado com
`FloatIsIntString` (conteúdo inteiro com decimais implícitas).

**Usar a precisão errada não quebra build, `vet` nem teste ingênuo.** Confira sempre contra o
`.pas` de origem.

## Cobertura

| Área | Declarado | Exercitado por teste |
|---|---|---|
| Mini-DOM (parse, navegação, `OuterXML`, `Caminho`) | 100% | 100% |
| Tipos de campo (`tcStr`/`Int`/`DeN`/`Dat`/`Hor`/`Bool`/`CNPJCPF`) | 100% | 100% |
| Enums compartilhados | 9 enums | 9 enums, com round-trip `String()`/`Parse()` |
| Validador CNPJ/CPF | numérico e alfanumérico | ambos |
| Chave de acesso | DV, validação, literal | DV conferido com casos calculáveis à mão |
| `INI` | leitura, escrita, round-trip | 100% |

Nada aqui gera XML — `pcn` é, nesta fase, exclusivamente de leitura. A geração entra junto com
a emissão.

## Testes

```bash
go test ./packages/pcn/... -v
```

Sem dependência externa e sem arquivo fora de `t.TempDir()`.

## Geração de XML (`xml_builder.go`)

Fase de emissão: o construtor `Elem` (porte da escrita de `TACBrXmlNode`) e os equivalentes do
`TACBrXmlWriter.AddNode` por tipo de campo:

| Função | Semântica do AddNode |
|---|---|
| `NodeStr` / `NodeStrSemFiltro` | `tcStr` com/sem `FiltrarTextoXML` (acentos, espaços, quebras→`;`) |
| `NodeInt` | `tcInt` — zero é vazio quando opcional; `PadLeft` de zeros até o mínimo |
| `NodeDec` | `tcDeN` — casas decimais **fixas**, ponto; zero é vazio quando opcional |
| `NodeDat` | `tcDat` — `AAAA-MM-DD`; tempo zero é vazio |
| `FormatarDataHoraXML` + `OffsetUF` | `DateTimeTodh` + `GetUTC(UF)` (AC −05; AM/RR/RO/MT/MS −04; resto −03) |

Regra herdada do ACBr: campo **obrigatório** com valor vazio gera a **tag vazia**; opcional com
valor vazio não gera nada. A `ListaDeAlertas` (wAlerta) não foi portada — omissão deliberada,
a validação efetiva é o XSD da SEFAZ.
