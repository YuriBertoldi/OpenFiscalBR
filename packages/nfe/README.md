# packages/nfe

Nota Fiscal Eletrônica (modelo 55) e Nota Fiscal de Consumidor Eletrônica
(modelo 65) — porte do componente **ACBrNFe** do [Projeto ACBr](http://projetoacbr.com.br).

> **Estado:** em construção. Esta etapa entrega apenas os **tipos e a conversão
> de enums**. O modelo de dados, o leitor, o gerador, os web services, os
> eventos e o `.ini` chegam nas etapas seguintes — veja [Roteiro](#roteiro).

## Origem dos fontes

| Unit Delphi | Arquivo Go |
|---|---|
| `PCNNFe/pcnConversaoNFe.pas` | `types.go` + `conversao.go` |

### Por que `pcnConversaoNFe.pas`, e não `ACBrNFe.Conversao.pas`

O ACBrNFe declara os **mesmos 51 enums em duas units**, e elas divergem:

- `Base/ACBrNFe.Conversao.pas`
- `PCNNFe/pcnConversaoNFe.pas`

Vale a segunda, por três evidências independentes:

1. **Resolução do `uses`.** Todas as units vivas — `Base/ACBrNFe.XmlWriter.pas`,
   `.XmlReader.pas`, `.IniWriter.pas`, `.IniReader.pas` e `ACBrNFe.Classes.pas`
   — listam `pcnConversaoNFe`, e o Delphi resolve identificador duplicado pela
   **última** unit listada. Em `ACBrNFe.Classes.pas`, a única que ainda lista as
   duas, `pcnConversaoNFe` vem depois.
2. **Os próprios autores do ACBr desativaram a unit `Base/`.** Em
   `ACBrNFe.ValidarRegrasdeNegocio.pas:45` e em
   `Base/Servicos/ACBrNFe.EventoClass.pas:48` a linha `ACBrNFe.Conversao` está
   **comentada**, logo abaixo do `pcnConversaoNFe` ativo.
3. **O gerador não compila sem ela:** `Base/ACBrNFe.XmlWriter.pas:546` usa
   `fnDebito`, membro que só existe em `PCNNFe/pcnConversaoNFe.pas:303`.

Tomar a unit `Base/` por fonte da verdade perderia, em silêncio:

| Perda | Onde |
|---|---|
| `fnCredito` (`"5"`) e `fnDebito` (`"6"`) em `finNFe` | `pcnConversaoNFe.pas:303` |
| `fpPagInstantaneoPIXAutomatico` (`"23"`) e `fpTEFBookTransfer` (`"24"`) | `pcnConversaoNFe.pas:156` |

Travado por `TestFormaPagamento_MembrosExclusivosDaUnitPCN` e
`TestFinalidadeNFe_CreditoEDebitoExistem`.

## Semântica das tabelas de código

As duas funções de conversão do ACBr (`ACBrDFe.Conversao.pas`) têm
comportamento que não se adivinha pela assinatura, e os dois estão replicados:

| Função | Comportamento | Consequência portada |
|---|---|---|
| `EnumeradoToStr` | **não interrompe** o laço — a última ocorrência do membro vence | `CrtRegimeNormal` grava `"3"`, não `""`, apesar de aparecer na posição 0 da tabela |
| `StrToEnumerado` | interrompe na **primeira** ocorrência; sem casar, devolve `AEnumerados[0]` (não `-1`) | `ParseCRT("")` devolve `CrtRegimeNormal` sem erro; o valor devolvido junto do erro é o fallback do ACBr |
| `StrToEnumerado` | compara com `AnsiSameText` (**ignora caixa**) | `ParseIndEscala("s")` → `IeRelevante` |
| laço próprio com `=` | compara **com** caixa, e levanta `EACBrException` em vez de devolver fallback | `ParseIndicadorPagamento`, `ParseTpGuia`, `ParseTpNFDebito`, `ParseTpNFCredito`, `ParseCSTIS` |

> **`indPag` não vem do par deprecado.** `IndpagToStr`/`StrToIndpag` estão
> marcados `deprecated` em `pcnConversaoNFe.pas:63-64`; quem lê e grava `indPag`
> de verdade é `IndpagToStrEX`/`StrToIndpagEX` (`:554-572`) — `XmlReader.pas:238`
> e `:1497`, `IniReader.pas:204` e `:1510`, `IniWriter.pas:225` e `:1348`,
> `JSONReader`/`JSONWriter`. Por isso `IndicadorPagamento` está no grupo da
> comparação exata, e o erro de `ParseIndicadorPagamento` **tem de ser tratado**:
> ali o ACBr aborta, não há fallback com o mesmo significado.

## Tipos implementados

### `conversao.go` — enums do leiaute (40)

Cada um expõe `String()` com o código exato do leiaute e `ParseXxx(string) (T, error)`.

`IndicadorPagamento`, `DeterminacaoBaseIcms`, `DeterminacaoBaseIcmsST`,
`MotivoDesoneracaoICMS`, `CstIpi`, `IndicadorProcesso`, `CRT`,
`IndicadorTotal`, `ECFModRef`, `ISSQNcSitTrib`, `DestinoOperacao`,
`ConsumidorFinal`, `PresencaComprador`, `FormaPagamento`, `BandeiraCartao`,
`TipoViaTransp`, `TipoIntermedio`, `IndISSRet`, `IndISS`, `TipoAutor`,
`IndOperacao`, `IndSomaPISST`, `IndSomaCOFINSST`, `FinalidadeNFe`,
`IndicadorNFe`, `TipoOperacao`, `CondicaoVeiculo`, `TipoArma`, `IndEscala`,
`ModalidadeFrete`, `Autorizacao`, `IndIntermed`, `TpAto`, `IndImport`,
`MotRedAdRem`, `TpMotivo`, `TpGuia`, `TpNFDebito`, `TpNFCredito`, `CSTIS`.

### `types.go` — controle do componente (7)

| Tipo | Papel |
|---|---|
| `StatusNFe` | estado interno do componente (sem representação no leiaute) |
| `VersaoDF` | versão do leiaute (`2.00`..`4.00`), com `Float()` e `ParseVersaoDFFloat` **por faixa** |
| `VersaoQrCode` | versão do QR-Code da NFC-e, com `AjustarVersaoQrCode` |
| `ModeloDF` | `55`/`65`, com `Prefixo()` |
| `SchemaNFe` | 50 schemas XSD, com `SchemaEvento()` |
| `LayOut` | 20 serviços da SEFAZ, com `Servico()` e `Schema()` |
| `ForcarGeracaoTag` | opção de configuração do gerador |

### Enums compartilhados — **não** redeclarados aqui

`TipoAmbiente`, `TipoEmissao`, `Indicador`, `IndicadorEx`, `OrigemMercadoria`,
`CSTIcms`, `CSTPis`, `CSTCofins` e `TpPagAnt` vivem em
[`packages/pcn`](../pcn) (porte de `ACBrDFe.Conversao.pas`).

## Fidelidade

Replicado de propósito, mesmo parecendo errado:

- **`SchemaNFe.String()` vem da RTTI, não do array.** `SchemaNFeToStr` usa
  `GetEnumName` + `Copy` e é a única chamada por `ACBrNFe.pas`;
  `TSchemaNFeArrayStrings` não é referenciada por nenhuma unit da NFe. As duas
  discordam em `InsucessoEntregaNFe` e `CancInsucessoEntregaNFe`.
- **`LayOut.Servico()` é vazio nos três layouts de URL.** `LayOutToServico` só
  mapeia os 17 primeiros; `TLayOutArrayStrings`, que tem nomes para os outros
  três, está morto.
- **`ParseLayOutServico("RecepcaoEvento")` sempre devolve `LayNFeCCe`**, porque
  o `StrToEnumerado` para na primeira ocorrência. Os três layouts de evento
  compartilham serviço e schema, então a volta não perde nada de útil.
- **`ParseVersaoDFFloat` nunca falha.** As faixas de `DblToVersaoDF` são
  abertas nas duas pontas, o que torna o ramo de erro do Delphi inalcançável.
- **Ordens que não acompanham o código**, mantidas como declaradas:
  `TipoAutor` (`taOutros`=`"9"` antes de `taEmpresaSucessora`=`"8"`),
  `CstIpi` (`00`, `49`, `50`, `99`, só então `01`), `FormaPagamento`
  (`19` → `90`, `98`, `99` → `20`), `TipoOperacao` (`toOutros`=`"0"` por
  último) e `MotivoDesoneracaoICMS` (`12` → `16` → `90`).
- **`IndicadorTotal` nasce "compõe o total".** `itSomaTotalNFe` é o ordinal
  zero, então um item recém-criado grava `"1"`.

### Divergências deliberadas

A escolha de `pcnConversaoNFe.pas` **não** é divergência: é a leitura correta de
qual unit o compilador Delphi usa. As divergências desta etapa são duas, ambas
da política "não replicar comportamento indefinido" do `CLAUDE.md`:

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `ParseCSTIS` | `TryStrToCSTIS` (`:1789-1803`) **não atribui `Value`** quando não encontra — o chamador fica com variável não inicializada; `StrToCSTIS` levanta exceção | devolve `CstisNenhum` + `error` |
| 2 | `ParseTpGuia`, `ParseTpNFDebito`, `ParseTpNFCredito`, `ParseIndicadorPagamento` | `raise EACBrException` | devolve o membro "Nenhum"/zero + `error`, pela convenção de erros do projeto (biblioteca não entra em pânico) |

Em ambos o `error` carrega o valor inválido e é testável com `errors.Is`.

### Duplicação de `TpNFDebito` com `packages/rtc`

O ACBr declara `TtpNFDebito` **duas vezes** — em `ACBrDFe.Conversao.pas` (origem
de `rtc.TpNFDebito`) e em `pcnConversaoNFe.pas` (origem do daqui) — com os
mesmos 9 membros. Em Go são dois tipos distintos, e cada um faz uma coisa:

- `rtc.TpNFDebito` é a chave que decide a emissão do grupo RTC
  (`rtc/xml_writer.go:170`) e **não tem** `String()`/`Parse`.
- `nfe.TpNFDebito` carrega o código do leiaute.

Mantidos separados para não editar package de outro território, e **travados por
`TestTpNFDebito_OrdinaisAlinhadosComRTC`**: a conversão entre os dois, que a
etapa do gerador vai precisar, é por ordinal, e um membro inserido no meio de
qualquer das duas listas mapearia o valor errado sem erro de compilação.
Revisável na etapa 4, se valer a pena unificar.

### Omissões

**Fora de escopo** — apresentação do DANFE, que depende de engine gráfica sem
equivalente em Go:

- `*ToStrTagPosText`, `*ToDescricao`, `*ToDescStr`, `indProcToDescrStr`,
  `modFreteToDesStr`, `ArmaTipoStr` e a família `Veiculos*Str`.
- Os enums usados **unicamente** pelo DANFE: `TpcnPercentualTributos`,
  `TpcnImprimeDescAcrescItem`, `TImprimirUnidQtdeValor` e
  `TpcnInformacoesDePagamento`.

**Adiada** — código vivo que depende de tipo ainda não portado:

- `StrToTpEventoNFe` (`pcnConversaoNFe.pas:1145-1178`) — a tabela dos 51 códigos
  de evento da NF-e, registrada no `initialization` da unit (`:1812`,
  `RegisterStrToTpEventoDFe`). Depende de `TpcnTpEvento`, que ainda não existe em
  [`packages/pcn`](../pcn). Vai com a **etapa 7** (eventos).

**Código morto no ACBr** — sem nenhum chamador em todo o `Fontes/`:

- `CRTTocRegTrib` (`:788-794`) e `StrToSchemaNFe` (`:1241-1264`).

### Declarado × exercitado

| Tipo | Situação |
|---|---|
| `IndicadorNFe` | nenhuma unit do componente o lê ou grava — a consulta de NF-e destinadas foi desativada pela SEFAZ |
| `CSTIS` | **não tipar campo com ele.** O próprio ACBr desistiu do enum até a publicação da tabela oficial e deixou o campo como `string`: `ACBrDFe.RTC.Classes.pas:784-785` e `:796-797` têm a versão `TCSTIS` comentada e a `string` ativa, idem `JSONReader.pas:1816` e `JSONWriter.pas:2587`. [`packages/rtc`](../rtc) já segue isso (`ISel.CSTIS` é `string`, `classes.go:397`) e lê valores fora desta tabela — `"001"` em `rtc_test.go:596`. Tipar o campo com `CSTIS` faria o leitor recusar conteúdo que o ACBr aceita hoje |

Os dois são mantidos para o inventário da unit de origem ficar completo.

## Como rodar os testes

```bash
go test ./packages/nfe/... -v
```

Sem rede, sem certificado e sem dependência externa.

### O que os testes travam

| Teste | Defeito que pega |
|---|---|
| `TestEnums_TabelaCobreTodoMembroDeclarado` | **tabela mais curta que o enum** — constante acrescentada sem o código correspondente, que passaria a gravar string vazia em silêncio. Usa o sentinela `_fimXxx`, que acompanha o bloco `const` |
| `TestEnums_CodigoDeCadaOrdinalVoltaAoMesmoOrdinal` | código **repetido** e desalinhamento entre `String()` e `Parse`, nos 43 enums de uma vez |
| `TestEnums_TabelaLiteralDosNaoSequenciais` | dois códigos **trocados de posição** nos 14 enums fora de ordem numérica — é o único defeito da família que sobrevive ao round-trip, e só se pega com uma transcrição independente do `.pas` |
| `TestTpNFDebito_OrdinaisAlinhadosComRTC` | divergência de ordinal entre o `TpNFDebito` daqui e o de `packages/rtc` |
| `TestEnums_ParseRecusaCodigoInexistente` | erro não reconhecível por `errors.Is` |
| `TestCRT_RegimeNormalGravaTresEVazioEhLidoComoRegimeNormal` | "simplificar" a tabela do CRT para uma ocorrência só |
| `TestIndicadorTotal_ValorZeroCompoeOTotal` | inverter a ordem dos membros e mudar o default de todo item |
| `TestEnums_CodigosForaDeOrdemNumerica` | "arrumar" por ordem numérica os enums fora de ordem |
| `TestTabelasDeControle_CobremTodoMembroDeclarado` | tabela curta em `StatusNFe`, `SchemaNFe`, `LayOut` e `VersaoDF` |
| `TestFormaPagamento_MembrosExclusivosDaUnitPCN` e `TestFinalidadeNFe_CreditoEDebitoExistem` | repontar o port para `Base/ACBrNFe.Conversao.pas` |
| `TestEnums_SentinelaNegativoGravaVazio` | perder o sentinela `TXxx(-1)` |
| `TestParse_SensibilidadeACaixaSegueAFuncaoDeOrigem` | uniformizar a comparação de caixa entre os dois conversores |
| `TestSchemaNFe_NomeVemDaRTTIENaoDoArrayMorto` | trocar a RTTI pelo array morto e procurar o XSD errado |
| `TestLayOut_*` | preencher os três serviços de URL que o ACBr deixa vazios |
| `TestParseVersaoDFFloat_ResolvePorFaixaENaoPorIgualdade` | trocar faixa por igualdade e recusar `3.01`/`4.01` |
| `TestAjustarVersaoQrCode_NaoExisteQRCodeAteALeiaute310` | gerar QR-Code em leiaute que não o admite |

A correspondência ordinal→código foi conferida **mecanicamente** contra
`pcnConversaoNFe.pas` nos 43 enums, aplicando a semântica de última ocorrência
do `EnumeradoToStr`.

## Roteiro

| # | Etapa | Estado |
|---|---|---|
| 1 | `types.go` + `conversao.go` | **nesta entrega** |
| 2a | `classes.go` — cabeçalho, totais, transporte, cobrança | pendente |
| 2b | `classes.go` — `det` (produto e impostos) | pendente |
| 3 | `xml_reader.go` | pendente |
| 4a/4b | `xml_writer.go` | pendente |
| 5 | regras de negócio, chave de acesso, QR-Code | pendente |
| 6a | status, consulta e `urls.go` por UF | pendente |
| 6b | autorização, inutilização, consulta cadastro, CSC | pendente |
| 7 | eventos (cancelamento, CC-e, manifestação) + `StrToTpEventoNFe` | pendente |
| 8 | Distribuição DF-e | pendente |
| 9 | `ini_reader.go` + `ini_writer.go` | pendente |
| 10 | demo em `demos/nfe` | pendente |
