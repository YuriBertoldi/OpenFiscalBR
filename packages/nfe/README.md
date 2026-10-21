# packages/nfe

Nota Fiscal Eletrônica (modelo 55) e Nota Fiscal de Consumidor Eletrônica
(modelo 65) — porte do componente **ACBrNFe** do [Projeto ACBr](http://projetoacbr.com.br).

> **Estado:** em construção. Estão prontos os **tipos/enums**, o **modelo de
> dados completo**, o **leitor de XML** e o **gerador do cabeçalho**. O gerador
> do item, os web services, os eventos e o `.ini` chegam nas etapas seguintes —
> veja [Roteiro](#roteiro).

## Origem dos fontes

| Unit Delphi | Arquivo Go |
|---|---|
| `PCNNFe/pcnConversaoNFe.pas` | `types.go` + `conversao.go` |
| `Base/ACBrNFe.Classes.pas` | `classes.go` (nota) + `classes_det.go` (item) |
| `Base/ACBrNFe.XmlReader.pas` | `xml_reader.go` |
| `Base/ACBrNFe.XmlWriter.pas` | `xml_writer.go` + `chave_acesso.go` |

> A unit de classes vem em **dois arquivos** porque o subgrupo `det` sozinho
> tem 29 das 75 classes. `classes.go` traz a nota (identificação, emitente,
> destinatário, totais, transporte, cobrança, informações adicionais) e
> `classes_det.go` traz o item (produto, os dez grupos de imposto e os grupos
> por tipo de mercadoria regulada).

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

Vivem em [`packages/pcn`](../pcn), que é o porte de `ACBrDFe.Conversao.pas`:
`TipoAmbiente`, `TipoEmissao`, `Indicador`, `IndicadorEx`, `OrigemMercadoria`,
`CSTIcms`, `CSTPis`, `CSTCofins`, `TpPagAnt` e — acrescentados pela etapa 2a —
`TipoNFe`, `TipoImpressao`, `ProcessoEmissao`, `IndIEDest`, `RegTribISSQN`,
`TpIntegra` e — pela etapa 2b — `CSOSNIcms` e `IndIncentivo`.

Os oito novos foram para o `pcn` em vez de serem declarados aqui porque são
enums de **todo DFe**, não da NF-e: CT-e e MDF-e usam os mesmos. Declará-los no
`nfe` garantiria a triplicação, que é exatamente o problema já registrado em
[`TpNFDebito`](#duplicação-de-tpnfdebito-com-packagesrtc).

## Modelo de dados

Porte de `Base/ACBrNFe.Classes.pas`. **75 classes / 651 campos**, conferidos
nome a nome e tipo a tipo contra o `.pas` (490 campos primitivos, zero
divergência de tipo).

| Arquivo | Escopo | Classes |
|---|---|---|
| `classes.go` | a nota | 46 |
| `classes_det.go` | o item (`det`) | 30 |

### Coleções

Todo `TXxxCollection`/`TXxxCollectionItem` virou slice simples. **Exceção
única:** `TpagCollection` carrega `vTroco` na própria coleção
(`Classes.pas:1933`) — o troco é um só para a nota inteira, não por forma de
pagamento. Virou `type Pag struct { DetPag []PagItem; VTroco float64 }`.

### O `default` da property NÃO inicializa nada

É a armadilha central do modelo de dados. No Delphi, `property X: T ... default valor`
é metadado de streaming de design-time; quem inicializa é o `constructor
Create`. Na NF-e os dois **divergem**:

| property | declara | `TIde.Create` faz | valor efetivo |
|---|---|---|---|
| `Ide.indPag` | `ipPrazo` | nada | `ipVista` → `"0"` |
| `Ide.tpNF` | `tnSaida` | nada | `tnEntrada` → `"0"` |
| `Ide.tpImp` | `tiPaisagem` | nada | `tiSemGeracao` → `"0"` |
| `Ide.tpAmb` | `taHomologacao` | nada | `taProducao` → `"1"` |

Ou seja: **um `TIde` recém-criado no ACBr está em produção, não em
homologação.** O zero value do Go reproduz isso sem esforço, e é por isso que
`Ide` **não tem construtor** aqui. `TestIde_ZeroValueReproduzOACBr` trava essa
ausência.

### Construtores que existem porque o zero value divergiria

| Construtor | Origem | O que atribui | Por que importa |
|---|---|---|---|
| `NovoPagItem` | `Classes.pas:4441` | `IndPag = IpNenhum`, `TBand = BcNenhum` | o zero do Go grava `"0"` e `"01"` — declararia pagamento à vista com cartão Visa |
| `NovoEmit` | `Classes.pas:2702` | `CRT = CrtRegimeNormal` | o zero do Go grava `"1"` (Simples Nacional) |
| `NovaNFe` | `Classes.pas:2406` | `xPais`/`cPais` = `BRASIL`/`1058` no emitente e no destinatário, mais o `CRT` herdado | sem isso o país sai vazio nos dois endereços |
| `NovoProd` | `Classes.pas:2747` | `IndEscala = IeNenhum` (ordinal **2**) | o zero do Go grava `"S"` — todo item declararia produção em escala relevante |
| `NovoDetItem` | `Classes.pas:2537` | herda o `Prod` de `NovoProd` | idem, via item |

`NovoPagItem` é o mesmo defeito que o porte da NFGas encontrou em `tiSim`: o
membro que significa "ausente" não está no ordinal zero.

Os três devolvem **valor**, e não ponteiro como o `NovoXxx` dos packages
irmãos: produzem structs de dados que a `NFe` embute por valor, sem estado nem
identidade a preservar, e devolver ponteiro obrigaria a desreferenciar em toda
atribuição.

### Valores especiais de endereço no exterior

`CMunExterior` (`9999999`), `XMunExterior` (`"EXTERIOR"`) e `UFExterior`
(`"EX"`), porte de `Classes.pas:2366-2369`. `AjustarMunicipioUF`
(`XmlWriter.pas:255-257`) substitui `cMun`/`xMun`/`xUF` por eles sempre que
`cPais` não for o Brasil. Emitir o código real do município no lugar deles
produz documento rejeitado, e nada no build acusa.

### Contrato de zero-padding — para a etapa do gerador

Três campos são numéricos no modelo mas exigem zero à esquerda no XML. O
`AddNode` do ACBr faz o preenchimento; quem gerar **não pode** usar
`strconv.Itoa` direto:

| Campo | XSD exige | Como o ACBr grava |
|---|---|---|
| `CEP` | `[0-9]{8}` | `tcInt, 08, 08` (`XmlWriter.pas:787`, `929`, `974`, `1024`) |
| `RefNF.Modelo` e `RefNFP.Modelo` | `"01"`/`"02"` | `tcInt, 02, 02` (`XmlWriter.pas:663` e `:692`) |
| `InfNFe.Versao` | `"4.00"` | `FloatToString '#0.00'` (`XmlWriter.pas:330`, `392`) |
| `Comb.CProdANP` | `[0-9]{9}` | `tcInt, 09, 09` (`XmlWriter.pas:1657`) |
| `ISSQN.CMunFG` | 7 dígitos | `tcInt, 07, 07` (`XmlWriter.pas:3372`) |
| `OrigComb.CUFOrig` | 2 dígitos | `tcInt, 2, 2` (`XmlWriter.pas:1791`) |

`Comb.CProdANP` é o mais exposto: qualquer código ANP abaixo de 100.000.000
sai com menos de nove dígitos e o documento é rejeitado. O preenchimento vem de
`PadLeft` em `ACBrXmlWriter.pas:410-411`, não do tipo.

### Condicional por versão de leiaute — `med`

Único grupo do modelo com comportamento que muda pela **versão**, e não por
conteúdo (`XmlWriter.pas:1593-1620`):

| `infNFe.Versao` | Campos emitidos | `MaxMed` |
|---|---|---|
| `>= 4` | `cProdANVISA`, `xMotivoIsencao` | 1 |
| `< 4` | `nLote`, `qLote`, `dFab`, `dVal` | 500 |

Os quatro campos antigos viraram o grupo `rastro` na 4.00. Emitir os dois
conjuntos juntos produz XML rejeitado, e o limite de ocorrências também muda —
é a dimensão que o `CLAUDE.md` aponta como a mais fácil de perder no writer.

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
qual unit o compilador Delphi usa.

**Dos enums e do modelo**, ambas da política "não replicar comportamento
indefinido" do `CLAUDE.md`:

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `ParseCSTIS` | `TryStrToCSTIS` (`:1789-1803`) **não atribui `Value`** quando não encontra — o chamador fica com variável não inicializada; `StrToCSTIS` levanta exceção | devolve `CstisNenhum` + `error` |
| 2 | `ParseTpGuia`, `ParseTpNFDebito`, `ParseTpNFCredito`, `ParseIndicadorPagamento` | `raise EACBrException` | devolve o membro "Nenhum"/zero + `error`, pela convenção de erros do projeto (biblioteca não entra em pânico) |

Em ambos o `error` carrega o valor inválido e é testável com `errors.Is`.

**Do leitor de XML** — seis, todas com comentário `DIVERGENCIA` no código,
teste próprio e registro em
[`divergencias-acbr.md`](../../.claude/skills/convert/referencias/divergencias-acbr.md):

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | pagamento na versão 3.xx | reaproveita `AuxNode` e procura `<pag>` **dentro de `<cobr>`** — nenhum pagamento é lido | resolve o nó de pagamento separadamente |
| 2 | `cana/forDia/@dia` | `StrToInt` sem `Def` nem checagem → access violation | vale 0 |
| 3 | `infAdic/obsCont/@xCampo` | AV quando há atributo que não é `xCampo` | vale vazio |
| 4 | `refNFP/mod` | lê com `tcInt` numa property string: perde o zero à esquerda de `"04"` | lê como texto |
| 5 | envelope sem `<NFe>` | devolve sucesso com a nota vazia | devolve `ErrNFeNaoEncontrada` |
| 6 | `DecodeHTMLEntities` | decodifica entidades uma segunda vez | não aplica: o `encoding/xml` do Go já decodificou |

**Do gerador de XML** — quatro, registradas em `divergencias-acbr.md`:

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | opção de campos da fatura | `CamposFatObrigatorios`, **ligada** pelo `Create` | `FatCamposOpcionais` — invertida, para o zero value do Go reproduzir o default do ACBr |
| 2 | `protNFe/infProt` | grava sem filtrar | `NodeStrSemFiltro`, para o `xMotivo` voltar igual ao recebido |
| 3 | documento nulo na geração | — | sentinela próprio `ErrDocumentoNulo` |
| 4 | dígito da chave | último caractere | idem, mas devolve **erro** se não for dígito |

#### A ordem dos campos é contrato — e o round-trip é cego para ela

O leitor acha tag por nome, então gerar→ler→gerar fica estável **mesmo com os
campos fora de sequência**. O XSD, porém, é `xs:sequence`: campo fora de ordem
é documento rejeitado pela SEFAZ.

Foi assim que uma inversão de `tpGuia`/`UFGuia` no `guiaTransito` passou por
todos os outros testes desta etapa e só apareceu na conferência com o `.pas`.
Desde então há `TestGerar_OrdemDosCamposSegueOXSD`, que afirma **ordem**, não
presença — e que foi verificado reintroduzindo o defeito.

> Atenção ao portar CT-e e MDF-e: os IDs do leiaute (`ZF05`, `ZF06`) **não**
> seguem a ordem do XSD. Seguir os IDs produz documento rejeitado.

#### Como a fidelidade do leitor foi verificada

Os dois eixos silenciosos foram conferidos **tag a tag, mecanicamente**, por
[`ferramentas/conferir-leitor-dfe.py`](../../ferramentas/conferir-leitor-dfe.py):

| Eixo | Tags conferidas | Divergências |
|---|---|---|
| precisão `tcDeN` → `ConteudoDeN` | 356 | 0 |
| `Find` × `FindAnyNs` | 487 | 0 |

O script existe porque **nenhum teste de leitura alcança a precisão**: o
parâmetro de casas do `pcn.ConteudoDec` só tem efeito quando o node está
marcado com `FloatIsIntString`, flag que o parser nunca liga — no XML da SEFAZ
o valor já vem com ponto decimal. `ConteudoDe2` e `ConteudoDe10` devolvem o
mesmo resultado, e a precisão é **contrato declarativo**, herdado do mesmo
mecanismo do ACBr. Só se verifica contra o `.pas`.

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
- `TveicProd.CombDescricao` (`getCombDescricao`, `Classes.pas:3698`) — property
  somente-leitura que devolve texto de impressão (`"01 - ALCOOL"`). Se o DANFE
  for portado um dia, note que o `else` da função devolve `tpComb` cru.
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
| `ISSQN.IndISSRet` | declarado na classe; o leitor de XML do ACBr **nunca o lê** (só o de `.ini`). Replicado |
| `COFINS.VBCProd` | declarado na classe; nenhum dos quatro subgrupos de COFINS o lê. Replicado |
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
| `TestNovoPagItem_IndPagETBandNaoComecamNoZero` | perder o construtor e gravar `tPag`/`tBand` falsos em toda forma de pagamento |
| `TestNovoEmit_CRTComecaEmRegimeNormal` | emitente nascer no Simples Nacional |
| `TestNovaNFe_AplicaOsDefaultsDoConstrutorDelphi` | país do emitente/destinatário sair vazio |
| `TestIde_ZeroValueReproduzOACBr` | "corrigir" o `Ide` com os defaults que a property declara e trocar produção por homologação |
| `TestGuiaTransito_ZeroValueNaoDeclaraGuia` | reordenar `TpGuia` e fazer toda nota nascer declarando uma guia de trânsito |
| `TestNovoProd_IndEscalaNaoComecaNoZero` | perder o construtor e gravar `indEscala="S"` em todo item |
| `TestICMS_ZeroValueReproduzOACBr` | "corrigir" o ICMS com `CST00` por causa do `default cst00` da property, fazendo todo item sem ICMS declarar tributação integral |
| `TestICMS_MotDesICMSNasceTaxi` | "consertar" para `MdiNenhum` e divergir do ACBr |
| `TestArma_TpArmaZeroEhUsoPermitido` | reordenar `TipoArma` |
| `TestImposto_DefaultsDeCSTCoincidemComOZero` | reordenar `CstIpi`/`CSTPis`/`CSTCofins` e mudar o CST default de todo item |
| `TestParseVersaoDFFloat_ResolvePorFaixaENaoPorIgualdade` | trocar faixa por igualdade e recusar `3.01`/`4.01` |
| `TestAjustarVersaoQrCode_NaoExisteQRCodeAteALeiaute310` | gerar QR-Code em leiaute que não o admite |

A correspondência ordinal→código foi conferida **mecanicamente** contra
`pcnConversaoNFe.pas` nos 43 enums, aplicando a semântica de última ocorrência
do `EnumeradoToStr`.


## Regras de negócio, QR-Code e URLs

### As 155 rejeições

`regras_negocio.go`, `regras_negocio_itens.go` e `regras_negocio_totais.go` portam
`ACBrNFe.ValidarRegrasdeNegocio.pas`. São as rejeições que a SEFAZ aplicaria e que o
componente antecipa localmente.

Três coisas mudam como elas se leem:

1. **Rodam sobre o documento pronto**, depois do gerador, e várias comparam o XML com a
   **configuração** — o ambiente da nota contra o do web service (regra 252), a UF do
   emitente contra a UF autorizadora (226). Sem `Configuracoes`, metade delas não existe.
2. **O "agora" entra por parâmetro**, não de `time.Now()`. Oito regras comparam datas com o
   momento da transmissão, e validação que muda de resultado conforme o relógio não é
   testável.
3. **Cada rejeição é um valor** (`ErroRegraNegocio`, com código, mensagem e `nItem`
   separados), não uma string concatenada como no ACBr — é o que o consumidor precisa para
   decidir o que fazer com cada uma.

```go
erros := nfe.ValidarRegrasNegocio(&nota, cfg, time.Now())
for _, e := range erros {
    fmt.Println(e) // 777-Rejeicao: Obrigatoria a informacao do NCM completo [nItem: 2]
}
```

Quatorze regras estão **comentadas** no ACBr e por isso não foram portadas; a lista e o
motivo de cada uma estão no doc-comment de `ValidarRegrasNegocio`.

### A parte difícil: os três níveis de acumulação

O laço que confere cada item é o mesmo que **acumula** os somatórios com que os totais serão
comparados. A divisão tem três níveis, e é ela que torna as regras de total fáceis de errar
em silêncio:

| Nível | O que entra | Condição |
|---|---|---|
| 1 | frete, seguro, desconto, outros, II, IPI, IPI devolvido, ICMS desonerado | **sempre**, mesmo em item fora do total |
| 2 | PIS e COFINS | num de **dois** somatórios, conforme o item seja mercadoria ou serviço |
| 3 | base e valor do ICMS, ST, FCP, valor do produto | só com `indTot = soma`; e o valor do produto ainda tem a exceção da nota de ajuste |

### QR-Code

`urls.go` porta as três versões (`ACBrNFe.pas:535-790`). A versão efetiva passa por
`AjustarVersaoQrCode`, que **rebaixa** a pedida conforme a versão do leiaute.

A versão 3 é a única que depende do **certificado**, e só em contingência offline: ela fecha
com uma assinatura RSA-SHA1, não com um hash. Para não arrastar o certificado para dentro do
leiaute, a assinatura entra como função:

```go
qr, err := nfe.GerarQRCode(&nota, cfg, func(texto string) (string, error) {
    return certificado.AssinarSHA1Base64(texto)
})
```

Quem decide se a nota está em contingência é o **tpEmis lido da chave**, não o do grupo
`ide` — é a chave que o consumidor tem em mãos ao conferir o documento.

### Tabela de URLs

`urls_tabela.go` traz 135 entradas do `ACBrNFeServicos.ini`, limitadas às duas famílias que o
QR-Code e o `urlChave` consomem. Os três passos de resolução do ACBr — chave versionada
exata, maior versão `<=` a pedida, chave sem versão — e o redirecionamento `Usar=` entre
seções já estão **resolvidos** nos valores: são codificação do `.ini`, não regra fiscal.

`ferramentas/conferir-urls-dfe.py` reimplementa a resolução e compara entrada a entrada.

### Validadores que passaram a existir no `pcn`

As regras dependiam de quatro validadores que o projeto não tinha. Entraram em
`packages/pcn` porque são Layer 1 e CTe, MDFe e NFSe vão reusá-los:

| Função | Origem |
|---|---|
| `ValidarIE` | `ACBrValidador.pas:1101-1580` — 27 algoritmos, um por UF |
| `ValidarMunicipio` | `ACBrDFeUtil.pas:893` |
| `ValidarGTIN` / `ValidarPrefixoGTIN` | `ACBrValidador.pas:1924` e `:1690` |
| `ValidarISUF` | `ACBrDFeUtil.pas:950` |

O `ValidarIE` mantém o arranjo **table-driven** do original: cada UF é uma linha de tabela e
um motor único roda sobre ela. Vinte e sete funções separadas esconderiam a diferença entre
as UFs no meio do código repetido — e é a diferença que importa.

Ele é verificado pelos **71 vetores da própria suíte de testes do ACBr** (37 válidos e 34
inválidos), não por exemplos inventados. Fica o registro para CTe e MDFe: o ACBr **tem**
suíte de testes, em `Testes/FPCUnit/`, e ela é fonte melhor do que exemplo construído à mão.

## Roteiro

| # | Etapa | Estado |
|---|---|---|
| 1 | `types.go` + `conversao.go` | concluída |
| 2a | `classes.go` — cabeçalho, totais, transporte, cobrança | concluída |
| 2b | `classes_det.go` — `det` (produto e impostos) + `NFe.Det` | concluída |
| 3 | `xml_reader.go` | concluída |
| 4a | `xml_writer.go` — cabeçalho, totais, transporte, cobrança | concluída |
| 4b | `xml_writer_det.go` — `det` (produto e impostos) | concluída |
| 5 | regras de negócio, QR-Code e URLs públicas | **nesta entrega** |
| 6a | status e consulta + endpoints dos web services | pendente |
| 6b | autorização, inutilização, consulta cadastro, CSC | pendente |
| 7 | eventos (cancelamento, CC-e, manifestação) + `StrToTpEventoNFe` | pendente |
| 8 | Distribuição DF-e | pendente |
| 9 | `ini_reader.go` + `ini_writer.go` | pendente |
| 10 | demo em `demos/nfe` | pendente |
