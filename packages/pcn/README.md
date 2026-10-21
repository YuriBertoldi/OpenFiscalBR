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
| `ACBrDFe/ACBrDFe.Conversao.pas` | enums compartilhados: `TipoAmbiente`, `TipoEmissao`, `Indicador`, `IndicadorEx`, `OrigemMercadoria`, `CSTIcms`, `CSTPis`, `CSTCofins`, `TpPagAnt`, `TipoNFe`, `TipoImpressao`, `ProcessoEmissao`, `IndIEDest`, `RegTribISSQN`, `TpIntegra`, `CSOSNIcms`, `IndIncentivo` |

| `ACBrDFe/Comum/ACBrDFeComum.Proc.pas` | `ProcDFe`, `LerInfProt` |
| `ACBrDFe/ACBrDFeUtil.pas` | `DigitoChaveAcesso`, `ValidarChaveAcesso`, `RemoverLiteralChave`, `ValidarAAMM`, `ValidarCodigoUF` |
| `ACBrDiversos/ACBrValidador.pas` | `ValidarCPF`, `ValidarCNPJ` (inclusive alfanumérico), `ValidarCNPJouCPF` |
| `ACBrComum/ACBrUtil.*` | `OnlyNumber`, `StringToFloat`, `StringDecimalToFloat`, `EncodeDataHora`, `ParseCompetencia` |
| `TMemIniFile` (RTL) | `INI` — leitor/gravador `.ini` mínimo, sem dependência externa |

> Os oito ultimos entraram com as etapas 2a e 2b do `packages/nfe`. Moram aqui, e não no
> package do documento, porque são de **todo DFe** — CT-e e MDF-e usam os
> mesmos. Dois têm código fora de ordem e merecem atenção ao portar:
> `IndIEDest` (o terceiro membro vale `"9"`, não `"3"`) e `RegTribISSQN`
> (`RTISSNenhum` é o **último** membro e vale `"0"`, não vazio). `IndIncentivo`
> repete a armadilha do `IndISSRet`: o "não" vale `"2"`, não `"0"`.

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

## Validadores de documento

Além do CPF/CNPJ e da chave de acesso, o package traz os validadores que as **regras de
negócio** dos DFe consomem:

| Função | Arquivo | Origem no ACBr |
|---|---|---|
| `ValidarIE(ie, uf)` | `validador_ie.go` | `ACBrValidador.pas:1101-1580` |
| `ValidarMunicipio(codigo)` | `validador_dfe.go` | `ACBrDFeUtil.pas:893` |
| `ValidarGTIN(numero)` | `validador_dfe.go` | `ACBrValidador.pas:1924` |
| `ValidarPrefixoGTIN(numero)` | `validador_dfe.go` | `ACBrValidador.pas:1690` |
| `ValidarISUF(isuf)` | `validador_dfe.go` | `ACBrDFeUtil.pas:950` |
| `CompararValor(a, b, tolerância)` | `validador_dfe.go` | `ComparaValor` (ACBrUtil.Math) |

### `ValidarIE` é table-driven, como o original

Cada UF tem seu próprio algoritmo, mas o ACBr não os escreve um a um: descreve cada UF como
uma **linha de tabela** — tamanho, o que cada posição aceita, linha de pesos, módulo, ajustes
e um segundo dígito opcional — e roda um motor único sobre ela.

O porte mantém o arranjo, porque é o que torna a conferência com o original possível. A
tabela (`validador_ie_tabela.go`) segue a mesma ordem do `.pas`, para leitura lado a lado.

Dois detalhes que o motor precisa acertar e que são fáceis de inverter:

- **as posições são listadas da direita para a esquerda**: `digitos[0]` é o último caractere;
- **as colunas de pesos também**: a coluna 1 multiplica o último caractere do documento
  alinhado em 14 posições.

Algumas UFs **mutam** o documento antes de validar (BA e MT completam com zeros), outras
escolhem a regra pelo **tamanho** (AC, PE, RN, RO, SP, TO) e duas pelo **valor** (AP e GO
comparam o documento como texto contra faixas).

`ISENTO` é aceito: é o conteúdo que o leiaute manda informar quando o contribuinte não tem
inscrição.

### Como foram verificados

Os testes usam os **vetores da própria suíte do ACBr**
(`Testes/FPCUnit/ACBrDiversos/acbrvalidadortest.pas`): 37 inscrições válidas e 34 inválidas,
cobrindo as 27 UFs e as variantes por tamanho. Usar os vetores do original, e não exemplos
gerados aqui, é o que torna o teste capaz de reprovar o porte — uma IE calculada pelo próprio
motor validaria contra si mesma mesmo com o algoritmo errado.

Além deles, há três guardas que os vetores sozinhos não dariam:

- cada IE válida tem o **último dígito trocado** pelos outros nove e precisa reprovar;
- a matriz de pesos é **transcrita de novo** no teste, a partir do `.pas`, e comparada célula
  a célula;
- a tabela de prefixos GTIN é conferida quanto a faixa invertida e sobreposição.


## Catálogo de eventos de DFe (`conversao_evento.go`)

`TipoEvento` porta `TACBrTipoEvento` (`ACBrDFe.Conversao.pas:93`) — **96 membros** cobrindo
NF-e, CTe e MDFe. Ele vive aqui, e não em `packages/nfe`, porque o ACBr também o mantém único:
vários eventos de MDFe chegam amarrados à chave de uma NF-e, e o leitor precisa nomeá-los.

A conversão tem um desequilíbrio que governa o desenho:

| direção | |
|---|---|
| enum → código | total e sem ambiguidade (`Codigo()`) |
| código → enum | **ambíguo** no catálogo inteiro |

Sete códigos se repetem entre documentos diferentes:

| código | significados |
|---|---|
| `110112` | Encerramento (MDFe) / Cancelamento por substituição (NF-e) |
| `110115` | Não embarque / Inclusão de DF-e |
| `110116` | Alteração de poltrona / Pagamento de operação |
| `110117` | Excesso de bagagem / Confirmação de serviço do MDFe |
| `240140` | CT-e de substituição / Autorizado substituição |
| `240150` | CT-e de anulação / Autorizado ajuste |
| `240170` | Liberação de prazo de cancelamento (×2) |

Por isso **não existe** um `ParseTipoEvento` global. Cada documento tem a sua tabela —
`ParseTipoEventoNFe` é o porte do `StrToTpEventoNFe`, e CTe e MDFe ganharão as suas.

### `Codigo()` e `String()` não são a mesma coisa

Os dois devolvem `string`, e trocar um pelo outro **compila**:

```go
TeCancelamento.Codigo()  // "110111" -- vai no XML
TeCancelamento.String()  // "Cancelamento" -- nome curto, para log e arquivo
```

Um gerador que emitisse `<tpEvento>Cancelamento</tpEvento>` passaria em qualquer teste de
estrutura. O `descEvento` do XML é um terceiro texto ("Carta de Correcao"), que vive em
`packages/nfe/evento_descricao.go` porque é específico da NF-e.

### A tabela da NF-e lê mais do que gera

São **50 códigos** na leitura contra 31 na geração: a NF-e precisa nomear registro de
passagem, averbação de exportação e eventos de MDFe e CTe que chegam amarrados à chave da
nota. Estreitar a tabela para "o que a NF-e gera" é o atalho que faz a consulta devolver evento
sem nome.
