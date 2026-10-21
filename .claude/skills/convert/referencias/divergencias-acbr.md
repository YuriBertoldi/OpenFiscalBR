# Registro de divergências deliberadas em relação ao ACBr

Referência para quem for portar, revisar ou **desenvolver em cima** dos packages do
OpenFiscalBR. Lista os pontos em que o Go diverge do Delphi **de propósito**, com a evidência
no `.pas` de origem. Toda entrada aqui tem comentário `DIVERGENCIA` no código e teste travando
o comportamento — antes de "corrigir" qualquer um destes, leia a justificativa.

Regra de origem (`CLAUDE.md`, seção Fidelidade): comportamento determinístico do ACBr é
replicado mesmo quando parece errado; **não** se replica o que perde dado do documento, depende
de variável não inicializada ou de configuração da máquina. As divergências abaixo são desse
segundo grupo, aprovadas pelo usuário em 2026-09-29.

## Perda de dado no ACBr — corrigidas no Go (aprovadas)

| # | Onde | ACBr | OpenFiscalBR | Evidência |
|---|---|---|---|---|
| 1 | `nfgas` — `prod/qFaturada` | property `Integer` (**arredonda a quantidade de gás**) apesar de o leitor ler `tcDe4` | `float64` com as 4 casas | `ACBrNFGas.Classes.pas:413` vs `ACBrNFGas.XmlReader.pas:535` e XSD `TDec_1100_1104` |
| 2 | `nfgas` — `prod/cClass` | property `Integer` (perde o zero à esquerda de código `[0-9]{7}`) apesar de o leitor ler `tcStr` | `string` | `ACBrNFGas.Classes.pas:410` vs `ACBrNFGas.XmlReader.pas:532` e XSD `xs:string pattern [0-9]{7}` |
| 3 | `nfgas` — `motDesICMS` (XML) | lê a tag com `tcInt` e atribui o inteiro ao enum: o **código do leiaute vira ordinal** (código "1"→2º membro; 16 e 90 estouram o enum) | parse pelo código, como o próprio ACBr faz no `.ini` | `ACBrNFGas.XmlReader.pas:629` vs `ACBrNFGas.IniReader.pas:533` |

Consequência para quem desenvolve em cima: **não** compare valores desses três campos com a
saída do ACBrNFGas Delphi esperando igualdade — o lado errado é o Delphi. Se um dia o ACBr
corrigir, o modo atualização da `/convert` acusa o `.pas` mudado e esta tabela é o contexto.

## Demais divergências deliberadas (mesma política)

Documentadas em detalhe no `README.md` de cada package e em
`.claude/skills/portar-leitor-dfe/referencias/armadilhas-leitor-dfe.md`:

- **Fuso horário** (`dhEmi`, `dhRecbto`...): o ACBr descarta o offset; o Go preserva como
  `time.FixedZone`. Hora de parede idêntica campo a campo.
- **Locale**: `StrToDate`/`DateTimeToStr`/separador decimal do sistema → formatos fixos
  (competência `AAAAMM`, data ISO, ponto decimal). O mesmo arquivo passa a ser lido igual em
  qualquer máquina.
- **Exceção → `error`**: leitura nunca entra em pânico nem aborta o lote; nota torta vira
  entrada em `*ErrosLote` com arquivo/índice/chave.
- **Split de lote** por decoder XML em vez de busca de string case-sensitive
  (`</NFGas>`/`</NFGasProc>`/`</procNFGas>`).
- **`pgto/@nPag` ausente** vale 0 (ACBr: `StrToInt` levanta exceção).
- **`chDFePagAnt` no `.ini` da RTC**: o ACBr lê de variável jamais inicializada (sempre vazio,
  `ACBrDFe.RTC.IniReader.pas:163`); o Go lê da chave homônima.
- **Laço sem `Clear`** (`gTarif`, `gPagAntecipado`): slices reinicializados — o acúmulo do
  original só é observável em reuso de objeto, comportamento indefinido.
- **Getter com efeito colateral** (`XMLAssinado` que assina): não replicado; assinar será
  método explícito na fase de emissão.

## O que foi REPLICADO mesmo parecendo errado (não "corrigir")

Determinístico e sem perda de dado — mudar isso quebra a compatibilidade com o ACBr:

- `ICMS41` é procurado no achatamento do ICMS apesar de não existir no XSD da NFGas; `pDif`/
  `vICMSOp`/`vICMSDif` de `ICMS51` não são lidos.
- `indIEDest` não é lido do XML (só do `.ini`); `vBCIRRF` nunca é lido; `infProt/@Id` idem.
- `infCpl`: leiaute admite 5 ocorrências, lê-se só a primeira (TODO aberto no ACBr).
- `qrCodNFGas`: CDATA removido só na primeira ocorrência (`StringReplace` sem `rfReplaceAll`).
- `enderCorresp` não lê `cPais`/`xPais`; `gpBioDiferenca` não é lido nos grupos ad valorem da
  monofasia; `pDevTrib` não é lido em `gIBSMun`/`gCBS` no `.ini`.
- Protocolo da consulta só é lido com `cStat` em `{100, 101, 104, 150, 151, 155}`.

## Divergências da GERAÇÃO/EMISSÃO da NFGas (aprovadas em 2026-09-29)

Bugs determinísticos do `ACBrNFGas.XmlWriter.pas` corrigidos no porte — todos com teste em
`packages/nfgas/xml_writer_test.go` citando a divergência:

| # | Onde | O ACBr faz | O porte faz | Fonte |
|---|---|---|---|---|
| 1 | raiz do documento processado | gera `NFGasProc` | gera `nfgasProc`, como o XSD (`nfgasProc_v1.00.xsd`) e o próprio LEITOR do ACBr esperam | `GerarXml`, linha 248 |
| 2 | `finNFGas`, `indOrigemQtd`, `tpMotNaoLeitura`, `tpProc`, `modBCST`, `motDesICMS` | passa o enum cru ao `AddNode` → grava o **ordinal** | grava o código do leiaute (`String()`) | `Gerar_Ide` etc. |
| 3 | ICMS70 | gera o bloco `vICMSDeson` **duas vezes** (a 2ª com `cBenef`) | só o primeiro bloco (`vICMSDeson`+`motDesICMS`+`indDeduzDeson`); `cBenef` não sai no ICMS70 | `ICMS70`, linhas 917–930 |
| 4 | CST de ICMS fora do leiaute | `Result := nil` + `AppendChild` em seguida → **access violation** | o grupo `imposto` sai nil (item sem imposto), sem pânico | `Gerar_det_imposto`, linha 976 |
| 5 | competência/dhCont com data zero | `FormatDateTime` da data zero → `189912` (em `Gerar_gFat` e `Gerar_gCons`) e `1899-12-30T...` (em `dhCont` quando só `xJust` veio preenchido) | tag vazia | `Gerar_gFat`:1194, `Gerar_gCons`:1337, `Gerar_Ide`:393. Obs.: no `Gerar_gNF` o **próprio ACBr** já guarda a data zero (`if CompetEmis > 0`, :531-542) — ali não há divergência |
| 6 | `cNF` aleatório | `GerarCodigoDFe` sorteia com 8 dígitos — a posição 36 da chave da NFGas é o `nSiteAutoriz`, então código de 8 dígitos corrompe a chave | sorteio com 7 dígitos, mesma lista de códigos proibidos | `ACBrDFeUtil.pas` |

### Divergência de TIPO: `TIndicador` tem `tiSim` como ordinal ZERO

`TIndicador = (tiSim, tiNao)` (`ACBrDFe.Conversao.pas:199`). Um `TImposto`/`TProd` recém-criado
no Delphi tem `indSemCST`/`indDevolucao` = `tiSim` — regerar uma nota lida de XML sem essas tags
faria o ACBr **descartar o grupo ICMS inteiro** e marcar devolução em todo item. No porte, os
campos `Imposto.IndSemCST`, `ICMS.IndSemCST`, `Prod.IndDevolucao` e `GProcRef.IndDevolucao` são
`pcn.IndicadorEx` (zero = `TieNenhum` = não informado). Round-trip ler→gerar coberto por
`TestGerarXML_RoundTripEstavel`.

### Replicado mesmo parecendo errado (geração)

- `gMedicao/nContrat` sai SEMPRE (`FormatFloat('00')` antes do `AddNode` opcional → `"00"`
  nunca é vazio).
- `pgto/@nPag` e `pgto/@idTransacao` saem sempre, mesmo vazios (`SetAttribute` incondicional).
- `infAdic` é gerado mesmo sem conteúdo.
- `Det` com `gNormal` E `gAgregadora` preenchidos: só `gAgregadora` sai (`Gerar_det`).
- `gProcRef/qFaturada` inteira sai como `tcInt` (sem casas); fracionada como `tcDe4`.
- **`tpMotNaoLeitura`/`xMotNaoLeitura` são RAMO MORTO na NFGas.** `Gerar_det_prod` só chama
  `Gerar_det_prod_gMedicao` quando `nMed > 0` **e** `gMedida.vMed > 0`
  (`ACBrNFGas.XmlWriter.pas:687`), mas dentro dele essas duas tags só saem no ramo
  `vMed = 0` (`:720-728`). Nenhum XML gerado pelo ACBr pode contê-las, e o porte replica a
  guarda (`packages/nfgas/xml_writer.go:377`, marcado `OMISSAO DO ACBr REPLICADA`). A **NFAg
  não tem o defeito**: lá a guarda exige apenas `nMed > 0`, e o ramo é alcançável.
- URLs: MA e PA apontam para seções `NFGas_SVAN_*` que **não existem** no
  `ACBrNFGasServicos.ini` → `ErrSemURL` (lacuna herdada, documentada).

### Omissões deliberadas (geração)

- `ListaDeAlertas`/`wAlerta` não portada — validação efetiva é o XSD da SEFAZ +
  `regras_negocio.go`.
- `NormatizarMunicipios` (lookup em arquivo de municípios) não portada.

### Acréscimos ao porte (existem no Go, não existem no ACBr)

Valem para `nfgas` e `nfag`, com teste em cada package:

| Acréscimo | Por quê |
|---|---|
| Leitura do **evento enviado** dentro de `procEvento*` (`RetEvento*.Evento`/`TemEvento`) | é a única fonte do `xJust`; o `TRetEvento*` do ACBr só lê a parte de retorno, e sem isso uma importação de cancelamentos não tem a justificativa |
| Leitura de `CNPJDest`, `emailDest` e `cOrgaoAutor` do retorno | declarados em `TRetInfEvento` e nunca lidos pelo ACBr; são dados do documento |
| **`GerarXMLProcEvento`** (`evento_writer.go`) | o `TRetEvento*` do ACBr só tem leitura — não há como produzir o envelope `procEvento*` que o contribuinte arquiva. Simétrico ao `GerarXMLProc` do documento; recusa com `ErrEventoAusente` quando falta a parte enviada, e herda do evento enviado os campos que o retorno não trouxer (senão `tpEvento` sairia como `-99999`, o `String()` do zero value) |

## Divergências do ACBrNFAg (aprovadas em 2026-09-30)

O `packages/nfag` (água, modelo 75) herda a política do NFGas. Específicas dele, todas com
teste em `packages/nfag/`:

| # | Onde | O ACBr faz | O porte faz | Fonte |
|---|---|---|---|---|
| 1 | raiz do proc na geração | `NFAgProc` (gera E lê) | gera `nfagProc` (XSD `procNFAg_v1.00.xsd`); o leitor aceita `NFAgProc`, `nfagProc` e `procNFAg` | `GerarXml`:271 / `LerXml`:145 |
| 2 | `retTrib` do item | grava **`vRetCOFINS`** | grava `vRetCofins` | XSD `nfagTiposBasico_v1.00.xsd:925` e o próprio leitor (`Ler_RetTrib`:536) usam a grafia minúscula — o valor gravado pelo ACBr se perde na releitura. Flagrado pelo teste de round-trip byte a byte |
| 3 | `Ler_Dest` | **sem guard** `if not Assigned` → AV com `<dest>` ausente | leitura tolerante | `XmlReader.pas:307` |
| 4 | `Ler_gMed` @nMed | `StrToInt` sem Def → exceção com atributo ausente | vale 0 | `XmlReader.pas:369` |
| 5 | `.ini`, `gPagAntecipadoNNN` do item | writer grava com índice **0-based** (`IniWriter.pas:346`) e o reader lê **1-based** (`IniReader.pas:357`) — o round-trip do ACBr perde o grupo | writer usa o mesmo 1-based do reader | flagrado por teste de round-trip; perde dado, então corrigido (não replicado) |

### Bugs determinísticos REPLICADOS (não "corrigir")

- **INI, versão**: gravada em `[infNFAg]` (`IniWriter.pas:154`), lida de `[infNFGas]`
  (`IniReader.pas:144`, copy-paste da NFGas) — o round-trip do ACBr perde a versão.
- **INI, TFU**: `Gerar_TFU` grava na seção `'TFS'+NNN` (`IniWriter.pas:498`); o reader lê de
  `[TFUNNN]` — TFU gravado nunca é relido.
- `tpFat` não é lido do XML (`Ler_Ide` não tem a tag); `vBCIRRF` é gerado e nunca lido.
- Assimetrias de precisão writer×reader do próprio ACBr (replicadas): prod
  `vItem`/`vProd` leitura tcDe10 × geração tcDe2; `pPIS`/`pCOFINS` leitura De4 × geração
  De2; `medMensal`/`consumo` leitura De4 × geração De2; gProcRef `vItem`/`vProd` leitura
  De8 × geração De2.
- `DescricaoTipoEvento` = "CANCELAMENTO DE NF3-e" (typo, só descritivo);
  `IdentificaSchema` procura `<infNFGas` (typo inofensivo, default já é schNFAg).
- Recepção síncrona: sucesso exige **cStat do retorno = 104** (não 100 como na NFGas) —
  não é bug, é o contrato do `TNFAgRecepcao.TratarResposta`, replicado e testado.

## Decisões de origem do ACBrNFe (2026-10-20)

Não são divergências — são leituras de qual código do ACBr está vivo. Ficam aqui porque
a leitura ingênua leva ao arquivo errado, e o erro é silencioso.

### A unit de conversão da NFe é `PCNNFe/pcnConversaoNFe.pas`

O ACBrNFe declara os **mesmos 51 enums em duas units**, e elas divergem:
`Base/ACBrNFe.Conversao.pas` e `PCNNFe/pcnConversaoNFe.pas`. **Vale a segunda**, por três
evidências independentes:

1. Todas as units vivas (`Base/ACBrNFe.XmlWriter.pas`, `.XmlReader.pas`, `.IniWriter.pas`,
   `.IniReader.pas`, `ACBrNFe.Classes.pas`) listam `pcnConversaoNFe`, e o Delphi resolve
   identificador duplicado pela **última** unit listada. Em `ACBrNFe.Classes.pas` — a única
   que ainda lista as duas — `pcnConversaoNFe` vem depois.
2. Os próprios autores do ACBr **comentaram** `ACBrNFe.Conversao` em
   `ACBrNFe.ValidarRegrasdeNegocio.pas:45` e em `Base/Servicos/ACBrNFe.EventoClass.pas:48`,
   logo abaixo do `pcnConversaoNFe` ativo.
3. `Base/ACBrNFe.XmlWriter.pas:546` usa `fnDebito`, que só existe em
   `pcnConversaoNFe.pas:303` — o gerador não compilaria contra a unit `Base/`.

O que se perderia ao portar a unit `Base/`:

| Perda | Onde |
|---|---|
| `fnCredito` (`"5"`) e `fnDebito` (`"6"`) em `finNFe` | `pcnConversaoNFe.pas:303` |
| `fpPagInstantaneoPIXAutomatico` (`"23"`) e `fpTEFBookTransfer` (`"24"`) | `pcnConversaoNFe.pas:156` |

Travado por `TestFormaPagamento_MembrosExclusivosDaUnitPCN` e
`TestFinalidadeNFe_CreditoEDebitoExistem` em `packages/nfe`.

> O mesmo par de units duplicadas existe no **CTe** e no **MDFe**. Confira o `uses` das
> units `Base/` antes de portar os enums de cada um — não presuma que a resposta é igual.

### `EnumeradoToStr` fica com a ÚLTIMA ocorrência

`ACBrDFe.Conversao.pas:843` não interrompe o laço. Em tabela que repete um membro, vence
a última posição. É o que faz `crtRegimeNormal` gravar `"3"` e não `""`, apesar de
aparecer na posição 0 da tabela do `CRTToStr` (`pcnConversaoNFe.pas:776`).

O par `StrToEnumerado` (`:826`) faz o oposto — para na **primeira** — e, quando nada casa,
devolve `AEnumerados[0]`, **não** `-1`. Por isso todo `ParseXxx` de `packages/nfe` devolve
o fallback do ACBr junto do `error`: o leitor que ignora o erro obtém o mesmo valor do
Delphi.

### `indPag` da NFe: o par vivo é `StrToIndpagEX`, não `StrToIndpag`

`IndpagToStr`/`StrToIndpag` estão marcados **`deprecated`** na própria unit
(`pcnConversaoNFe.pas:63-64`). Todo o código vivo usa o par `EX` (`:554-572`) —
`XmlReader.pas:238` e `:1497`, `IniReader.pas:204` e `:1510`, `IniWriter.pas:225` e
`:1348`, `JSONReader.pas:260`/`:1535`, `JSONWriter.pas:245`/`:2185`.

Não é detalhe de estilo: o par `EX` usa laço próprio com `=` (**caixa respeitada**) e
**levanta `EACBrException`** quando nada casa, em vez do fallback silencioso de
`StrToEnumerado`. Portar o par deprecado planta a armadilha de o leitor gravar `indPag`
zerado numa nota em que o ACBr teria abortado.

> Vale a varredura equivalente no CTe e no MDFe: procure `deprecated` na unit de conversão
> antes de portar qualquer `StrToXxx`.

### Divergências deliberadas de `packages/nfe` (2026-10-20)

Política do `CLAUDE.md`: não replicar comportamento indefinido, e biblioteca não entra em
pânico. Ambas com teste e `errors.Is`.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `ParseCSTIS` | `TryStrToCSTIS` (`pcnConversaoNFe.pas:1789-1803`) **não atribui `Value`** quando não encontra — o chamador fica com variável não inicializada; `StrToCSTIS` levanta exceção | devolve `CstisNenhum` + `error` |
| 2 | `ParseTpGuia`, `ParseTpNFDebito`, `ParseTpNFCredito`, `ParseIndicadorPagamento` | `raise EACBrException` | devolve o membro "Nenhum"/zero + `error` |

### `TCSTIS` é string no ACBr vivo — não tipe campo com o enum

O enum existe na unit, mas o ACBr **desistiu dele** até a publicação da tabela oficial:
`ACBrDFe.RTC.Classes.pas:784-785` e `:796-797` têm a versão `TCSTIS` comentada e a
`string` ativa; idem `ACBrNFe.JSONReader.pas:1816` e `JSONWriter.pas:2587`.
`packages/rtc` já segue (`ISel.CSTIS` é `string`) e lê `"001"`, valor fora da tabela do
enum. `packages/nfe` porta o enum só para o inventário ficar completo, com o aviso no
doc-comment.

### Nome de schema da NFe vem da RTTI, não do array

`SchemaNFeToStr` (`pcnConversaoNFe.pas:1237`) usa `GetEnumName` + `Copy` e é a única
chamada por `ACBrNFe.pas`. O `TSchemaNFeArrayStrings` **não é referenciado por nenhuma
unit da NFe** e discorda em dois membros: diz `InsucessoEntrega`/`CancInsucessoEntrega`
onde a RTTI diz `InsucessoEntregaNFe`/`CancInsucessoEntregaNFe`. O nome decide qual XSD o
componente procura em disco. Mesmo caso em `TLayOutArrayStrings`, também morto.
