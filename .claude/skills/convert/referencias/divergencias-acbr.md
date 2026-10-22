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

### Divergências do LEITOR de XML da NFe (2026-10-21)

Porte de `Base/ACBrNFe.XmlReader.pas`. Todas com comentário `DIVERGENCIA` no
código e teste em `packages/nfe/xml_reader_test.go`.

| # | Onde | O ACBr faz | O porte faz | Fonte |
|---|---|---|---|---|
| 1 | `lerCobr`, pagamento na versão 3.xx | reaproveita a variável `AuxNode`: ela recebe `<cobr>` e **só** é reapontada para `<pag>` quando `Versao >= 4`. Na 3.xx o laço procura `<pag>` **dentro de `<cobr>`** — onde ele não existe — e **nenhum pagamento é lido**. Na 2.00 `tagPag` fica não inicializada | resolve o nó de pagamento separadamente do de cobrança; a 3.xx lê os `<pag>` do infNFe | `:1442-1519` |
| 2 | `lerCana`, `forDia/@dia` | `StrToInt(Attributes.Items['dia'].Content)` sem `Def` e sem checar nil → **access violation** com o atributo ausente | vale 0 | `:1614` |
| 3 | `lerInfAdic`, `obsCont`/`obsFisco` `@xCampo` | `if Attributes.Count > 0 then Items['xCampo'].Content` → AV quando há atributo que **não** é `xCampo` | vale vazio | `:1537`, `:1547` |
| 4 | `lerIdeNFref`, `refNFP/mod` | lê com `tcInt` e atribui a property **string** (`Classes.pas:116`): o zero à esquerda de `"01"`/`"04"` se perde | lê como texto, como o XSD declara | `:334` |
| 5 | `lerDocumento`, envelope sem `<NFe>` | `if NFeNode <> nil` sem `else`, e devolve `True`: o chamador fica com uma `TNFe` vazia sem saber | devolve `ErrNFeNaoEncontrada` | `:135-161` |
| 6 | `xNome` de dest/retirada/entrega/transporta | aplica `DecodeHTMLEntities` sobre o texto já lido | não aplica: o `encoding/xml` do Go já resolve as referências de entidade ao produzir o `CharData`, e uma segunda passada reduziria `&amp;amp;` a `&` | `:452`, `:498`, `:519`, `:1373` |

Replicado mesmo parecendo errado (**não** "corrigir**"**):

- **`cNF == 0` vira `-2`** (`:235`) — sentinela para o gerador sortear um código novo.
- **`dest` ausente marca `indIEDest = inNaoContribuinte`** (`:441-445`) — é o único
  guard do leitor que grava valor quando o grupo não existe.
- **Default de país só em `enderEmit`/`enderDest`** (`:383-389`, `:482-488`);
  retirada e entrega ficam com `cPais` 0 e `xPais` vazio.
- **`nItem` vem do contador, não do atributo `@nItem`** (`:543`) — nota com
  `@nItem` fora de ordem é renumerada.
- **ICMS (21 grupos), PIS (4) e COFINS (4) são lidos em SEQUÊNCIA**, e vale o
  **último** encontrado — semântica oposta à de um `PrimeiroDe`.
- **`motRedAdRem` só é lido com `pRedAdRem != 0`** (`:964`); sem o guard todo item
  monofásico declararia "transporte coletivo", que é o ordinal zero do enum.
- **`IPI.CST` ganha default por `tpNF`** antes de ler os subgrupos (`:1062-1066`),
  e só dentro do bloco `<IPI>`.
- **`qrCode`: CDATA removido só na primeira ocorrência** (`StringReplace` sem
  `rfReplaceAll`, `:1645-1646`).
- **Declarados e nunca lidos do XML:** `ISSQN.indISSRet` e `COFINS.vBCProd`.

> A precisão `tcDeN` e a escolha `Find`/`FindAnyNs` foram conferidas tag a tag por
> `ferramentas/conferir-leitor-dfe.py` — 356 e 487 tags, zero divergências. Esse
> script existe porque **nenhum teste de leitura consegue** distinguir `ConteudoDe2`
> de `ConteudoDe10`: o parâmetro de casas do `pcn.ConteudoDec` só tem efeito com
> `FloatIsIntString`, que o parser nunca liga. A precisão é contrato declarativo,
> e só se verifica contra o `.pas`.

### Divergências do GERADOR de XML da NFe — cabeçalho (2026-10-21)

Porte de `Base/ACBrNFe.XmlWriter.pas` (o grupo `det` fica para a etapa 4b). Todas com
teste em `packages/nfe/xml_writer_test.go` ou `xml_writer_ordem_test.go`.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | opção `CamposFatObrigatorios` | o `Create` a liga (`:220`); o campo se chama `CamposFatObrigatorios` | campo **invertido**: `Opcoes.FatCamposOpcionais`. Em Go o zero value de `bool` é `false`, e manter o nome original faria `Opcoes{}` gerar **diferente** do ACBr. Invertido, o zero value reproduz o default |
| 2 | `protNFe/infProt` | grava `Content` direto, sem `FiltrarTextoXML` | `NodeStrSemFiltro`, para o `xMotivo` voltar igual ao que a SEFAZ devolveu, acento incluso |
| 3 | `ErrXMLVazio` × documento nulo | — | sentinela próprio `ErrDocumentoNulo`: o do leitor significa "XML ausente", não "`*NFe` nil" |
| 4 | extração do dígito da chave | `ExtrairDigitoChaveAcesso` pega o ÚLTIMO caractere | idem — e devolve **erro** se não for dígito, em vez de silenciar uma chave malformada |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **`cPais` do emitente é a constante `CODIGO_BRASIL`**, não o campo do modelo (`:789`).
- **`cMunFG` do `retTransp` é gravado com `tcStr`**, sem zero à esquerda, enquanto todos os
  outros usam `tcInt` (`:3826`).
- **`cRegTrib` é omitido quando vale `RTISSNenhum`**, apesar de esse membro valer `"0"` no
  leiaute (`:3598`).
- **`agropecuario` é else-if**: havendo `defensivo`, a `guiaTransito` não sai nem preenchida
  (`:1071-1082`).
- **`qrCode` só é embrulhado em CDATA quando NÃO contém `"?p="`** (`:345`).
- **`cNF` sentinela:** `0`/`-1` sorteiam; qualquer outro negativo vira `0`. O `-2` que o
  leitor grava, portanto, **não** dispara sorteio — é o que torna o round-trip estável.
- **A chave usa o CNPJ do fisco em nota avulsa**, não o do contribuinte (`:311-314`).

> **A ordem dos campos é contrato, e o round-trip é cego para ela.** O leitor acha tag por
> nome, então gerar→ler→gerar fica estável mesmo com os campos fora de sequência — e o XSD é
> `xs:sequence`. Foi assim que uma inversão de `tpGuia`/`UFGuia` no `guiaTransito` passou por
> todos os outros testes da etapa 4a e só apareceu na conferência com o `.pas`. Desde então há
> `TestGerar_OrdemDosCamposSegueOXSD`, que afirma ORDEM e não presença.
>
> Atenção ao portar CTe e MDFe: os IDs do leiaute (`ZF05`, `ZF06`) **não** seguem a ordem do
> XSD, e seguir os IDs produz documento rejeitado.

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

### Divergências do GERADOR de XML da NFe — item `det` (2026-10-21)

Porte de `GerarDet` e dos 24 métodos que ele encadeia (`ACBrNFe.XmlWriter.pas:1118-3507`).
Testes em `packages/nfe/xml_writer_det_test.go`.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | opção `GerarTagIPIparaNaoTributado` | o `Create` a **liga** (`:214`) | campo **invertido**: `Opcoes.OmitirTagIPIParaNaoTributado`. Mesma razão do `FatCamposOpcionais` — o zero value do `bool` precisa reproduzir o default do original |
| 2 | `AddNodeCNPJCPF` com CNPJ obrigatório e vazio | preenche **quatorze zeros** (`ACBrXmlWriter.pas:179`: a condição é `obrigatorio OR tamanho > 0`) | idem — era um defeito do `gerarCNPJCPF` da etapa 4a, achado ao reler o helper para a DI |
| 3 | `tcHor` com hora `00:00:00` | trata como **vazio**: omite quando opcional e emite elemento **vazio** quando obrigatório (`ACBrXmlWriter.pas:324` olha só o relógio, não a data do `TDateTime`; `:442-446` dá a regra do obrigatório) | `nodeHora`, idem. Era outro defeito de 4a, achado por `conferir-gerador-dfe.py` |
| 4 | limpeza do CNPJ/CPF | `AddNodeCNPJCPF` usa `OnlyAlphaNum`, que **não** muda a caixa (`:161`); só `AddNodeCNPJ` usa a variante que maiusculiza (`:203`) | idem — `gerarCNPJCPF` passou a usar `pcn.OnlyAlphaNum`. A distinção é inerte hoje e passa a valer com o **CNPJ alfanumérico de 2026** |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **O `det` MUTA o documento em quatro campos**: `cEAN`/`cEANTrib` vazios viram `"SEM GTIN"`,
  o `xProd` do primeiro item vira o aviso de homologação, e `cEnq` vazio vira `"999"`
  (`:1163`, `:1183`, `:1237`, `:2990`). Não dá para evitar sem mudar a saída — o próprio
  ACBr relê `cEAN` logo depois para decidir se valida o GTIN.
- **`PISST` e `COFINSST`: o segundo `if` SOBRESCREVE o `Result` do primeiro** (`:3150-3177`
  e `:3322-3351`). Preenchendo `vBC`/`pPIS` **e** `qBCProd`/`vAliqProd`, o grupo por valor é
  perdido. O ACBr emite alerta e segue; é determinístico, então foi replicado — e há teste
  justamente para impedir que alguém "conserte" sem perceber a divergência.
- **`tpNFDebito = 06` (pagamento antecipado) não gera imposto nenhum.** O comentário da
  NT 2025.002 logo acima da condicional (`:1862`) lista o código entre as **exceções** que
  devem detalhar, mas a lista externa do código só excetua `tdPerdaEmEstoque` (`:1871`). A
  condicional interna (`:1875`), que existiria para dar a ele PIS/COFINS sem ICMS, fica
  quase morta: só é alcançável com `tpNFCredito` de exceção **e** `tpNFDebito` antecipado ao
  mesmo tempo. Comentário e código discordam no ACBr; o porte segue o **código**.
- **Dois CSTs/CSOSN caem em tag SEM sufixo**, pelo mesmo motivo — `case` sem `else`, `Result`
  de string fica vazio:
  - **`cstPart20` foi esquecido no `BuscaTag`** (`XmlWriter.pas:1916-1941`), embora exista no
    enum (`ACBrDFe.Conversao.pas:259`) e seja tratado no `case` do writer (`:2201`). Resultado:
    `<ICMS><ICMS>` aninhado.
  - **`csosnVazio` não é coberto pelo `CSOSNToStrTagPos`** (`ACBrDFe.Conversao.pas:1282`),
    gerando `<ICMSSN>`.

  Os dois são replicados e travados por `TestGerarDet_CSTSemEntradaNoMapaGeraTagAninhada` —
  que existe justamente para impedir que alguém "conserte" sem perceber a divergência.
- **`pMVAST` do grupo 90 sai sempre com duas casas** (`:2630`), enquanto todos os outros
  percentuais do writer seguem o `FormatoValor4ou2`.
- **`vBCFCPSTRet`/`pFCPSTRet`/`vFCPSTRet` são opcionais no `ICMSST` e obrigatórios no
  `ICMS60`** (`:2713` × `:2443`), e o `vFCPSTRet` muda de ID entre os dois (`N27c` × `N27d`).
- **`vBCFCPUFDest` e `pFCPUFDest` são marcados obrigatórios** contra a NT 2016.002 v1.50, que
  os tornou opcionais — justificativa no próprio código: SEFAZ que não implantaram o schema
  novo (`:3472`).
- **O bloco de ICMS do combustível (`ICMSComb`/`ICMSInter`/`ICMSCons`) é inalcançável a partir
  da 2.00** (`:1704`), e o comentário "Manter Disponível para versão 3.0" deixou o código
  morto no lugar.
- **`indISSRet` está comentado** (`:3409`) e não é gerado, embora a classe carregue o campo.

> A conferência tag a tag de **tipo** e **obrigatoriedade** é mecânica:
> `ferramentas/conferir-gerador-dfe.py` — 374 tags do gerador inteiro (4a + 4b), zero
> divergências. O script existe pelo mesmo motivo do `conferir-leitor-dfe.py`: o round-trip
> é cego para as duas coisas, porque o leitor converte `"0.00"` e `"0.0000"` no mesmo float
> e um campo zerado relido continua zerado.
>
> A **ordem** tem o modo `--ordem` do mesmo script, usado grupo a grupo: conferidos os 9 ramos
> de CST, os 5 de CSOSN e os 8 demais grupos do item, todos batendo. Como guard permanente
> ficam os testes Go de sequência, e o `det` acrescentou `TestGerarDet_OrdemDosCamposSegueOXSD`
> e `TestGerarDet_PBCOpVemAntesDeUFST`. Esse segundo é a mesma armadilha do `guiaTransito` da
> etapa 4a: os IDs do leiaute dizem `N24 = UFST`, `N25 = pBCOp`, e o XSD manda `pBCOp` primeiro.
>
> **Cuidado com teste de condicional por versão que busca no documento inteiro.** O
> `vICMSDeson` existe no item *e* no `ICMSTot`, e o do total sai a partir da 3.10 — um
> `strings.Contains` no XML completo faz o teste passar mesmo com o item mudo. Os testes de
> corte por versão recortam o grupo `imposto` antes de procurar.

### Divergências das REGRAS DE NEGÓCIO, QR-Code e URLs da NFe (2026-10-21)

Porte de `ACBrNFe.ValidarRegrasdeNegocio.pas` (155 rejeições ativas), do QR-Code em três
versões (`ACBrNFe.pas:535-790`), da tabela de URLs do `ACBrNFeServicos.ini` e dos validadores
que faltavam em `packages/pcn` (IE, município, GTIN, ISUF).

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `ValidarISUF` | confere o tamanho sobre `OnlyNumber(ISUF)` mas **indexa a string crua** (`ACBrDFeUtil.pas:957-965`); uma inscrição com máscara somaria pontos e levantaria exceção | as duas coisas usam o valor limpo |
| 2 | resultado das validações | devolve `string` vazia (ok) ou mensagem de erro | `error` com sentinela (`pcn.ErrIEInvalida`) ou `bool`, conforme o consumidor |
| 3 | URL não encontrada | devolve string vazia **em silêncio**, e o QR-Code sai sem o endereço da SEFAZ | `nfe.ErrSemURL` |
| 4 | QR-Code v3 sem certificado | o `SSL.CalcHash` falharia internamente | `nfe.ErrCertificadoObrigatorio`, e só no caso offline — fora dele a v3 não assina |
| 5 | erros das regras | acumula numa **string** e devolve um booleano | `[]*ErroRegraNegocio`, com código, mensagem e `nItem` separados |
| 6 | `destinatarioDoQRCode` com `idEstrangeiro` só de espaços | testa `<> ''` **sem trim** e só trima o resultado: o estrangeiro vence o CNPJ e o `cDest` sai vazio (`ACBrNFe.pas:662`) | idem |
| 7 | regra 515 (município do local de entrega) | compara `Entrega.UF = 'EX'` **literal**, enquanto a 513 usa `SameText` na retirada (`:608` × `:585`) | idem — a assimetria é do original e está replicada |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **O `ValidarIE` soma o próprio dígito verificador** no ISUF: o laço vai de 1 a **9** com peso
  `10-i`, então o nono termo tem peso 1. Parece engano de leitura e é o que faz o resto fechar.
- **CSC vazio vira um código DERIVADO da chave** (`ACBrNFe.pas:680`): `Copy(chave,7,8) + '20' +
  Copy(chave,3,2) + Copy(cIdCSC,3,4)`. Não valida na SEFAZ; o ACBr faz para o DANFE de teste
  não quebrar. Determinístico, logo replicado.
- **Abaixo da 4.00 o `AjustarVersaoQrCode` devolve `veqr000`** ("sem QR-Code"), mas o `case` do
  ACBr cai no `else` e **gera a versão 1** (`:642-648`). Ele não deixa de gerar, rebaixa.
- **O `TipoDest` da v3 tem precedência invertida** (`:760-769`): marca 3 para estrangeiro e
  **depois** sobrescreve com 2 ou 1 se houver CNPJ/CPF. Um documento com os dois sai com tipo
  de CPF e **documento do estrangeiro** — os dois campos discordam entre si.
- **A regra 785 (entrega a domicílio proibida pela UF) tem a lista VAZIA**: o ACBr compara
  contra o literal `'XX'` e o próprio comentário pergunta "Qual estado não permite entrega a
  domicílio?" (`:368`). Nunca dispara. Replicada com a lista vazia.
- **A regra 707 é a única sem o prefixo "Rejeição:"** na mensagem (`:338`).
- **O `cUFsValidas` inclui `EX`**, que passa na validação de UF e reprova na de IE por não ter
  regra de tamanho (`ACBrValidador.pas:45`).
- **A regra 851 compara duas vezes com "menor que", em OU** (`:311`): soma de parcelas MAIOR
  que a fatura **não** é rejeitada, apesar da mensagem dizer "difere".
- **A regra 610 tem tolerância DUPLA** (`:1125`): falhando o milésimo, aceita 0.009 contra o
  total com **ou sem** a desoneração do ICMS — é como o ACBr convive com as duas leituras que
  as UFs deram ao `vICMSDeson`.
- **A regra 321 varre os itens com `repeat..until`** e por isso lê `Det[0]` mesmo com a lista
  vazia (`:1237`); o porte usa um `for` comum, que nesse caso não entra.
- **A tabela de prefixos GTIN não está ordenada**: 490-499 vem antes de 460-469. A busca é
  varredura linear, então não importa — mas um teste que exigisse ordem reprovaria à toa.
- **GO é a única UF com regra de PREFIXO** além do vetor de posições: o prefixo de dois
  dígitos tem de ser 10, 11 ou 20 a 29 (`ACBrValidador.pas:1261-1264`). No original a
  verificação apenas **grava a mensagem e segue** — o dígito é calculado do mesmo jeito e o
  documento reprova no fim. O efeito observável é o mesmo de reprovar na hora.
- **A regra 725 (CFOP da NFC-e) é testada por SUBSTRING** no ACBr (`pos(CFOP, 'XXXX,5101,…')`),
  não por pertinência ao conjunto (`:719`). Para CFOP de 4 dígitos os dois são equivalentes —
  nenhum trecho de 4 dígitos da lista cruza vírgula. **Divergem** para CFOP de outro tamanho:
  `"510"` o ACBr aceitaria e o porte rejeita. Portado como conjunto, que é o que o teste
  pretendia.

> **Três auditorias mecânicas novas**, todas pela mesma razão: o dado é grande demais para
> conferência a olho e nenhum teste Go consegue provar que uma transcrição está certa.
>
> - `ferramentas/conferir-regras-dfe.py` — 155/155 regras ativas com contraparte, e lista as
>   14 que o ACBr mantém **comentadas** (502, 508, 529, 571, 600, 629, 630, 719, 720, 806,
>   856, 858, 883, 888). O risco que ele cobre não é errar uma regra, é **esquecer** uma:
>   regra que some não quebra build, teste nem round-trip.
> - `ferramentas/conferir-urls-dfe.py` — 135 entradas idênticas à resolução do `.ini`, mais
>   três invariantes que o porte assume (a URL do QR-Code não varia por versão; a consulta da
>   NF-e também não; a da NFC-e é igual nas versões 2 e 3).
> - O `ValidarIE` é conferido pelos **71 vetores da própria suíte de testes do ACBr**
>   (`Testes/FPCUnit/ACBrDiversos/acbrvalidadortest.pas`), 37 válidos e 34 inválidos. Vale
>   lembrar para CTe e MDFe: o ACBr **tem** suíte de testes, e ela é fonte melhor do que
>   exemplo inventado — os meus estavam errados em 12 das 27 UFs.
>
> **Dois defeitos reais que só a revisão pegou**, e que nenhum dos 71 vetores do ACBr alcança:
>
> - a **regra de prefixo da IE de GO** tinha sumido do porte. O vetor negativo de GO do ACBr
>   (`"12.312.312-3"`) reprova pelo **dígito verificador**, então a ausência da regra ficava
>   invisível; prefixos 12 a 19 com dígito válido passavam.
> - o `ComparaValor` do ACBr **arredonda a diferença a nove casas** antes de compará-la com a
>   tolerância (`SimpleRoundTo(..., -9)`, `ACBrUtil.Math.pas:236`). Sem isso, uma diferença
>   que vale exatamente a tolerância pode sair do ponto flutuante como `0.0010000000000331966`
>   e reprovar documento válido — e essa função é o motor dos dezessete confrontos de total,
>   da 610, da 851 e da aritmética do troco.
>
> **Armadilha de teste encontrada três vezes nesta etapa:** asserção por `strings.Contains`
> num formato de campos separados. No QR-Code v3 o separador da versão é literalmente `|3|` e
> o ambiente de produção é `|1|`, então procurar o `TipoDest` por `Contains` casa com o campo
> errado e passa sempre. Confira por POSIÇÃO. O mesmo vale para o confronto de totais quando
> a fixture não exercita o campo: zero comparado com zero passa com a acumulação quebrada —
> a fixture da NFe não tem frete nem desconto, e o teste da tolerância dupla da 610 pulava
> **sempre** porque ela também não tem ICMS desonerado. Teste que depende de fixture precisa
> **falhar** quando o campo vem zerado, não passar em silêncio.

### Divergências dos WEB SERVICES da NFe — status e consulta (2026-10-21)

Porte de `TNFeStatusServico` e `TNFeConsulta` (`ACBrNFeWebServices.pas`), mais a tabela de
endpoints do `ACBrNFeServicos.ini`.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | tabela de endpoints | guarda por SEÇÃO do `.ini`, com redirecionamento `Usar=` em tempo de execução | guarda **por autorizador**, com o redirecionamento já resolvido. As 58 combinações de UF e ambiente colapsam em 15 autorizadores no modelo 55 e 12 no 65 |
| 2 | versão "achada" na busca | quando a versão pedida não existe, cai para a inferior **e devolve a versão achada**, que vira o `versaoDados` (`ACBrDFe.pas:505`) | a queda está materializada na **URL**; o `versaoDados` usa sempre a versão configurada. Dispensar a dimensão é deliberado: o `AchaArquivoSchema` (`:581-590`) desfaz a queda quando o `.xsd` da versão pedida existe, e ele existe na instalação padrão |
| 3 | `TratarResposta` da consulta | grava eventos em disco, atualiza as notas carregadas e ajusta `AguardarConsultaRet` | só lê a resposta. Gravar arquivo é papel de quem consome a biblioteca |
| 4 | `MinutesBetween > 50` no ajuste do SVRS | trunca para minutos inteiros, logo ajusta a partir de **51** minutos cheios | idem, com `Truncate(time.Minute)` |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **A SoapAction da Bahia na 4.00 vem do `.ini`, não do código.** `[NFe_BA_P]` declara
  `SoapAction=SOAP_V4_BA`, e naquela seção a action da consulta de protocolo é
  `.../NFeConsulta4/nfeConsultaNF` — **não** o `NFeConsultaProtocolo4` que o
  `DefinirServicoEAction` montaria. O `FPServico` (namespace do envelope) continua o
  montado, porque nenhuma seção usa a chave `WSDL=`; só a action muda. É o efeito do
  `EstaVazio(FPServico)` de `:2123`, e sem ele a consulta da Bahia não funciona.
- **O ajuste de horário do SVRS** (`:885-893`) subtrai uma hora do `dhRecbto` quando a URL é
  do SVRS e a diferença passa de 50 minutos. A terceira condição do original
  (`not IsHorarioDeVerao`) é hoje **sempre verdadeira**: o horário de verão foi extinto pelo
  Decreto 9.772/2019.
- **A leitura do `protNFe` e do `retCancNFe` é condicionada ao `cStat` da CONSULTA**
  (`RetConsSit.pas:196` e `:226`), com listas **diferentes**: `[100,101,104,150,151,155]`
  para o protocolo e `[101,151,155]` para o cancelamento. Uma rejeição que venha acompanhada
  de `protNFe` — caso que a SEFAZ-PR já produziu — tem o bloco descartado.
- **Os conjuntos de `cStat` são do ACBr, não inventados**: `CstatProcessado` é
  `[100, 120, 150]` e `CstatCancelada` é `[101, 135, 151, 155]` (`ACBrNFe.pas:295-311`). O
  120 é autorização por DPEC e é o que mais escapa.
- **`Processada()` olha o `cStat` externo; `Autorizada()` olha o do `protNFe`.** O ACBr usa
  os dois em momentos diferentes (`:2397` e `:2419`), porque uma nota cancelada volta com
  `cStat` externo de cancelamento e ainda traz o protocolo da autorização original.

> **Como o bloqueante foi achado — e por que a auditoria não pegou.** O
> `conferir-endpoints-dfe.py` comparava só as URLs; as chaves `WSDL=` e `SoapAction=` do
> `.ini` ficavam fora. Uma delas existe e muda o comportamento numa UF em produção. Fica o
> aviso para CTe e MDFe: o `.ini` do ACBr tem **três** eixos de configuração por seção, não
> um — endereço, namespace do serviço e SoapAction — e os dois últimos são declarados por
> redirecionamento para outra seção.

### Divergências dos WEB SERVICES da NFe — autorização, inutilização, cadastro e CSC (2026-10-21)

Porte de `TNFeRecepcao`, `TNFeRetRecepcao`, `TNFeRecibo`, `TNFeInutilizacao`,
`TNFeConsultaCadastro` e `TAdministrarCSCNFCe` (`ACBrNFeWebServices.pas`).

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `TNFeRetRecepcao.Executar` | consulta o recibo em **laço**, com `Tentativas` e `IntervaloTentativas` da configuração (`:1475-1497`) | `ConsultarRecibo` faz **uma** chamada. Repetir é política de quem consome, e o `context.Context` já dá o cancelamento; quem quiser o original chama em laço enquanto `LoteEmProcessamento()` for verdadeiro |
| 2 | `TratarResposta` de todos os seis | grava XML em disco (`GravarXML`, `Gravar`), atualiza as notas carregadas e monta nome de arquivo | só lê a resposta e, no caso síncrono, anexa o protocolo à nota em memória. Gravar é papel do consumidor |
| 3 | montagem do `nfeProc` | cola o `XMLAssinado` **como texto** com o `XMLprotNFe` (`:1551`) | o mesmo para o `ProcInutNFe` (`GerarXMLProcInutNFe`); para a nota, o `GerarXML` regera a partir do struct quando `ProcNFe.NProt` está preenchido, reembutindo a `Signature` lida de volta pelo `PrepararEnvio` — decisão já tomada no `nfag` |
| 4 | `ValidarDigest` | compara o `digVal` do protocolo com o `DigestValue` da nota e **levanta exceção** (`:1247`) | não implementado nesta etapa; o `digVal` é devolvido em `ProtDFe.DigVal` e a conferência fica com o consumidor |
| 5 | `ProcInutNFe` | prefixa `<?xml ... ?>` porque grava em disco na mesma linha (`:2779`) | devolve sem prólogo, como o `GerarXML` da nota |
| 6 | consulta de cadastro sem documento | monta `infCons` com `<CPF></CPF>` vazio e transmite | devolve `ErrDocumentoConsultaAusente` sem gastar a chamada |
| 7 | `DefinirDadosIntegrador` | alimenta o Integrador Fiscal do Ceará | fora do escopo do porte |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **`EhAutorizacao` não é só uma comparação de versão.** `ACBrNFe.pas:313` tem uma exceção por
  UF: a **NFC-e 3.10 no Amazonas** (cUF 13) permanece no par antigo
  `NfeRecepcao`/`NfeRetRecepcao`. Portar só `versao >= ve310` manda a NFC-e amazonense para um
  endereço que não existe.
- **A inutilização NÃO desvia para SVC.** `TNFeInutilizacao.DefinirURL` (`:2672`) usa
  `Configuracoes.WebServices.UF` cru — sem o `case FormaEmissao` que a autorização tem
  (`:1066`). Padronizar os dois é a "correção" que um revisor faria, e ela mandaria o pedido
  para o ambiente virtual, que não o atende. Travado em
  `TestEnviar_EInutilizar_EscolhemUFsDiferentesEmContingencia`.
- **No cabeçalho SOAP do envio, a versão vem do DOCUMENTO e o cUF vem da CONFIGURAÇÃO.** São
  eixos separados: `InicializarServico` (`:1026`) tira a versão do `infNFe.Versao` da primeira
  nota, mas `TNFeRecepcao` **não** sobrescreve `GerarUFSoap`, então o cUF cai no do
  `TDFeWebService`, que lê `Configuracoes.WebServices.UFCodigo` (`ACBrDFeWebService.pas:349`).
  A URL, essa sim, sai do `Ide.cUF`. Unificar os três é a simplificação que quebra o SVC.
- **O `indSinc` tem condição composta de três alternativas** (`:1146`): serviço de autorização
  **ou** modelo 65 **ou** versão ≥ 3.10. O único caso em que as três falham — NF-e 2.00 no lote
  antigo — sai sem a tag, e é a única configuração em que a presença dela causaria rejeição de
  schema.
- **O sufixo da SoapAction do lote zipado muda com a versão** (`:1113`): `ZIP` da 4.00 em
  diante, `LoteZip` antes. E o limite de 500 KB é medido **depois** da compressão (`:1159`) —
  medir antes recusaria lote que caberia.
- **O ramo síncrono do `TratarResposta` tem condição composta** (`:1191`):
  `(modelo 65 OU versão ≥ 3.10) E Sincrono`. Uma NF-e 2.00 pedida como síncrona ainda volta
  como recibo.
- **O protocolo síncrono só é anexado com DUAS condições** (`:1233`): `cStat` externo 104 **e**
  `CstatProcessado` no protocolo. E a nota é escolhida pela **chave**, não pela posição
  (`:1245`).
- **O 563 da inutilização não é sucesso, mas gera o `ProcInutNFe`.** `TratarResposta` devolve
  `True` só no 102 (`:2765`) e monta o documento em `[102, 563]` (`:2777`). O 563 diz que a
  faixa já estava inutilizada: o pedido falhou, a inutilização existe.
- **O Mato Grosso do Sul (cUF 51) aceita CPF na inutilização, com duas metades contraditórias**
  (`ACBrNFe.Inut.pas:128-145`): no **ID** o CPF entra com três zeros à esquerda, para ocupar as
  catorze posições do CNPJ; no **corpo** ele vai limpo, dentro de `<CPF>`. O original recalcula
  o documento depois de montar o ID justamente para desfazer o preenchimento.
- **O `<ano>` do corpo e o do ID saem por caminhos diferentes.** O ID usa
  `Copy(IntToStrZero(ano,4),3,2)` e sempre tem duas posições; o corpo usa
  `if ano > 2000 then ano := ano - 2000` e sai **sem** zero à esquerda. Para 2009 o ID traz
  `09` e o corpo traz `9`. O limite é `> 2000`, não `>= 2000`: o ano 2000 sairia inteiro no
  corpo — incoerência sem efeito prático, replicada de propósito.
- **O 151 do CSC conta como sucesso.** A faixa é `[150..153]` (`:3925`), e o 151 é "não há CSC
  ativo": a consulta funcionou, a resposta é que não há nada. Lendo só a descrição, o porte
  natural o trata como falha.
- **O `dadosCsc` só é emitido na REVOGAÇÃO** (`ACBrNFe.AdmCSC.pas:88`). Emiti-lo sempre faria a
  consulta mandar um CSC que ela não deveria mandar — e o código é segredo.
- **`TAdministrarCSCNFCe` manda `<soap12:Header>` em QUALQUER versão** (`:3940`), ao contrário
  de todos os outros serviços da NF-e, que o suprimem da 4.00 em diante (`:692`).
- **O Mato Grosso embrulha a consulta de cadastro duas vezes, da 4.00 em diante**: a mensagem
  entra num `<nfeDadosMsg>` extra (`:2958`) **e** o elemento do Body vira `<consultaCadastro>`
  (`:3019`). São duas mudanças na mesma condição.
- **A exceção de notação de serviços da Bahia na inutilização tem quatro fatores** (`:2697`):
  UF 29, modelo 55, versão 3.10 e emissão normal. Afrouxar qualquer um manda outra UF para um
  endereço que só a Bahia expõe. E, nesse ramo, a `SoapAction` é atribuída **fora** do guard
  `EstaVazio` — o guard protege só o `FPServico`.

**A "versão achada" volta — e desta vez ela importa.** A etapa 6a dispensou essa dimensão
(divergência #2 de lá) com o argumento de que o `AchaArquivoSchema` (`ACBrDFe.pas:580-590`)
desfaz a queda de versão quando o `.xsd` da versão pedida existe. O argumento vale para
`consStatServ`, `consSitNFe`, `inutNFe`, `nfe` e `retEnviNFe` — todos publicados em 3.10 e
4.00. Mas **dois serviços desta etapa não têm esse resgate**:

| Serviço | Chave no `.ini` | Schema instalado | Versão que o ACBr emite |
|---|---|---|---|
| `NfeConsultaCadastro` | `CadConsultaCadastro_2.00` | **só** `consCad_v2.00.xsd` | `2.00` |
| `AdministrarCSCNFCe` | `AdministrarCSCNFCe_1.00` | **nenhum** | `1.00` |

Emitir `versao="4.00"` nesses dois é rejeição certa: a SEFAZ valida a mensagem contra o XSD do
**serviço**, não contra o da NF-e. O porte materializou os dois em `versaoFixaPorServico`
(`urls_servicos.go`), e o `versaoDados` do cabeçalho acompanha. A supressão do cabeçalho
continua governada pela versão **configurada** (`:692`) — na consulta de cadastro as duas
divergem: um componente na 4.00 não manda cabeçalho, embora a mensagem seja 2.00.

Duas correções de etapas anteriores entraram junto:

- **O override de `SoapAction` do `.ini` só vence na 4.00.** O guard `if EstaVazio(FPSoapAction)`
  existe apenas no ramo `>= ve400` de todo `DefinirServicoEAction`; abaixo dele o código
  atribui sem guard e descarta o que o `LerServicoDeParams` trouxe. As etapas 5 e 6a aplicavam
  o override em qualquer versão. Não muda nada hoje — as seções `[SOAP_V4*]` só declaram chaves
  `_4.00` —, mas é a condicional que governa a atribuição. Materializada em `comOverrideDoIni`.
- **O redirecionamento `Usar=` NÃO é de seção inteira.** `ACBrDFe.pas:550-571` resolve
  **serviço a serviço**: tenta a seção local e só cai no `Usar=` quando aquela chave está
  vazia. Uma UF faz as duas coisas — o **MA**, que no modelo 55 delega ao SVAN mas mantém o
  próprio `NfeConsultaCadastro_2.00`. A tabela por autorizador continua válida porque o MA é a
  **única** UF do grupo SVAN; o `conferir-endpoints-dfe.py` passou a emitir AVISO nesse caso e
  a FALHAR se o grupo crescer. A premissa registrada na etapa 6a estava errada e foi corrigida.

Achados do `revisor-go` nesta etapa, e o que saiu deles:

- **bloqueante — versão do `admCscNFCe` e do `ConsCad`**: já descrito acima; o revisor chegou
  à mesma conclusão por caminho independente (cadeia `DefinirURL` → `LerServicoDeParams` →
  `AchaArquivoSchema`). Corrigido com `versaoFixaPorServico`.
- **importante — o override de `SoapAction` era procurado na seção errada**: a URL saía da
  seção do DOCUMENTO e o override, da seção da CONFIGURAÇÃO. No ACBr os três (URL,
  `FPServico`, `FPSoapAction`) voltam de **uma** chamada a `LerServicoDeParams`
  (`:1079-1087`), logo da mesma seção. Corrigido: `DestinoEnvio` e `ReciboDoLote` carregam o
  modelo e a UF usados na busca, e o override sai dali. Hoje o efeito é nulo — a Bahia é a
  única UF com `SoapAction=` e a action dela para autorização **coincide** com a que o código
  monta —, e é por isso que o teste precisa injetar um override fictício para ser capaz de
  falhar; está dito no comentário dele.
- **importante — a derivação de `Enviar` não tinha teste**: versão, UF, modelo e par de
  serviços saem da primeira nota, e os testes só exercitavam `enviarNaURL`, que recebe tudo
  pronto. Extraído `enderecoDoEnvio`, com cinco testes — inclusive o que fecha o laço da
  exceção do Amazonas, provando que ela chega ao ENDEREÇO e não só à função pura.
- **menor — `ConsultarRecibo` não anexava protocolo**: `TNFeRetRecepcao.TratarRespostaFinal`
  (`:1667-1679`) grava o `procNFe` **sempre** que a chave bate, inclusive em REJEIÇÃO — é o
  `cStat` e o `xMotivo` do protocolo que explicam o motivo ao usuário. Só a montagem do
  `nfeProc` é condicionada a `[100,120,150]`. A regra é **diferente** da do ramo síncrono, que
  exige `cStat` 104 do lote E `CstatProcessado` do protocolo. Exportado `AnexarProtocolos`
  com a regra assíncrona e teste que trava a assimetria.
- **menor — `ConsultarRecibo` usava o modelo da configuração**: `TNFeRecibo.DefinirURL`
  (`:1888-1899`) tira o modelo de `Items[0]`. Virou campo de `ReciboDoLote`, que o
  `ResultadoEnvio` já devolve preenchido.
- **menor — `rfIgnoreCase` do `StringReplace`**: conscientemente não replicado. O `pos` que o
  ACBr faz ANTES de trocar é sensível a caixa, então o ignore-case só teria efeito numa
  resposta que misturasse caixas no mesmo documento. Registrado no doc-comment da função.

> **Chave do `.ini` que não é lida:** as três seções `[SOAP_V4*]` declaram
> `CadConsultaCadastro_4.00`, mas o `LerServicoDeParams` procura pelo nome do serviço, que é
> `NfeConsultaCadastro`. A chave **nunca é encontrada** e nenhum override de SoapAction se
> aplica à consulta de cadastro — nem na Bahia. O porte replica isso porque o ACBr também não a
> acha; a tabela `soapActionPorAutorizador` não tem a entrada.

### Divergências dos EVENTOS da NFe (2026-10-21)

Porte de `TEventoNFe` (`ACBrNFe.EnvEvento.pas`, 3089 linhas), `TRetEventoNFe`
(`ACBrNFe.RetEnvEvento.pas`) e `TNFeEnvEvento` (`ACBrNFeWebServices.pas:3030`), mais o catálogo
`TACBrTipoEvento` (`ACBrDFe.Conversao.pas:93`), que foi para o `pcn` por ser compartilhado com
CTe e MDFe.

São **31 tipos de evento** no dispatch do gerador e **50 códigos** na tabela de leitura — a
tabela de leitura é mais larga porque a NF-e precisa nomear registro de passagem, averbação de
exportação e eventos de MDFe e CTe que chegam amarrados à chave da nota.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `chNFePend` (`RetEnvEvento.pas:620`) | procura uma tag `chNFePend` **dentro** do próprio elemento `chNFePend`, que no XSD é um `TChNFe` — tipo simples, com a chave como texto. **Nunca acha nada**; e grava em `Items[Idx]` com o índice do laço de `chNFePend`, que não é o do `retEvento` | lê o conteúdo do próprio nó e anexa ao item correto. Replicar um defeito que só produz vazio seria pior que corrigi-lo — a lista de EPEC pendentes é a única informação útil desse retorno |
| 2 | `numItem` do `itemPedido` | **grava** como atributo (`:1100`) e **lê** como tag (`:679`) — o round-trip dele perde o campo | lê o atributo, com retrocesso para a tag |
| 3 | `LerXml` (`:728-748`) | compara o nome da raiz contra quatro literais e sai calado quando nenhum casa | caminho genérico (procura `<evento>` e `<retEvento>` em paralelo, que são mutuamente inofensivos), com ramo próprio só para o `NFeDFe`, onde o `procEventoNFe` aparece **duas vezes aninhado** |
| 4 | `xCondUso` da CC-e | o setter `setxCondUso` **sobrescreve sempre** o texto legal | só preenche quando vazio — senão não haveria como corrigir o texto se a legislação mudar antes do porte |
| 5 | tipo de evento sem gerador | o `case` fica sem ramo e o `infEvento` sai **sem `detEvento`**; a SEFAZ rejeita sem dizer o porquê | devolve `ErrEventoNaoSuportado` |
| 6 | laço de tentativas, gravação em disco, `DefinirDadosIntegrador`, validação contra XSD | faz | fora do escopo |
| 7 | `ManifestarDestinatario` | o CNPJ do manifestante sai do componente | é **parâmetro obrigatório**: quem manifesta é o destinatário, e a chave carrega o documento do *emitente* |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **O `gControleEstoque` tem SETE formas, todas com o mesmo nome de elemento.** No Delphi a
  distinção é por **sobrecarga** — sete `Gerar_gControleEstoque` que diferem apenas no tipo do
  parâmetro (`:2378`, `:2390`, `:2406`, `:2646`, `:2724`, `:2735`, `:3076`):

  | contexto | tags |
  |---|---|
  | consumo pessoal | `qConsumo` / `uConsumo` |
  | importação em ALC/ZFM | `qtde` / `unidade` |
  | perecimento (adquirente) | `qPerecimento` / `uPerecimento` |
  | perecimento (fornecedor) | `qPerecimento` / `uPerecimento` / **`vIBS`** / **`vCBS`** |
  | imobilização | `qImobilizado` / `uImobilizado` |
  | combustível | `qComb` / `uComb` |
  | item não fornecido | `qNaoFornecida` / `uNaoFornecida` |

  Um gerador genérico emitiria `qConsumo` nos sete casos e **nenhum round-trip pegaria** — o
  elemento pai tem o mesmo nome nos dois lados da troca.

- **Três elementos saem com nome repetido em contextos diferentes**: `<gConsumo>` vem do
  consumo pessoal *e* da ALC/ZFM; `<gPerecimento>` vem do transporte do adquirente *e* do
  fornecedor; `<gControleEstoque>` dos sete acima. O que decide a coleção, na leitura, é o
  `tpEvento` — e é por isso que o `Ler_DetEvento` tem um `case` (`:508`) ao lado da leitura
  achatada dos campos comuns.
- **`gCredPres` é o nome da CLASSE; a tag é `<gCredPresOper>`** (`:2543`).
- **`gCredito` é o único grupo da Reforma sem `gControleEstoque`**, e os valores têm nome
  próprio: `vCredIBS` e `vCredCBS`, não `vIBS`/`vCBS` (`:2670`).
- **O `tpAutor` é fixo `"1"` em seis eventos e vem do campo nos demais.** O ACBr comenta o
  porquê em cada um: "Conforme o Schema o único valor aceito é 1". Uniformizar os dois grupos
  — em qualquer direção — produz rejeição numa metade. O `indQuitacao` do 112110 é igual
  (`:2503`), e nem campo tem.
- **As quatro manifestações do destinatário diferem por UM parâmetro.** Confirmação e ciência
  só têm `descEvento`; desconhecimento tem `xJust` **opcional**; operação não realizada tem
  `xJust` **obrigatório** (`:696-739`).
- **O `cOrgao` zero cai para o `cUF` da chave** (`getcOrgao`, `EventoClass.pas:705`). É o que
  faz o evento ir para a SEFAZ da nota sem o consumidor informar nada — e é o que torna a
  manifestação do destinatário um caso especial, porque ela precisa do **91** explícito.
- **O autor do evento tem três condicionais empilhadas** (`:982-1013`): CNPJ vazio cai para o
  documento da chave; um documento de catorze dígitos numa nota de **série 910..969** é um CPF
  com três zeros à frente e tem os três descartados; e essa segunda regra **não** vale para a
  manifestação, porque ali o autor é o destinatário e não o emitente da chave.
- **O `Id` do evento tem 54 posições** e o `nSeqEvento` vai com **duas** (`%.2d`, `:286`).
- **A lista de eventos que levam ACENTO** (`:947`) não é derivável: o cancelamento genérico é
  da Reforma e está **fora**; o ator interessado não é e está **dentro** — e a descrição dele
  é ASCII puro. A implicação só vale numa direção: toda descrição acentuada precisa estar na
  lista.

**O roteamento é a condicional mais rica de todo o porte da NF-e.** Eventos do mesmo documento
vão para SEFAZ **diferentes**, e a ordem das atribuições em `DefinirURL` (`:3076-3160`) é o que
decide:

| evento | destino |
|---|---|
| CC-e, cancelamento, cancelamento por substituição, prorrogação (4) | SEFAZ da UF da **chave** (ou SVC em contingência) |
| todo o resto, inclusive a manifestação do destinatário | Ambiente Nacional (`AN`) |
| insucesso na entrega e seu cancelamento | `SVRS`, **sempre** |
| conciliação financeira | `SVRS` na NF-e; UF da chave na NFC-e — **olha o modelo da CONFIGURAÇÃO**, não o da chave |
| os 15 da Reforma Tributária | `SVRS`, **sempre** |
| EPEC de NFC-e **em SP** | serviço **próprio** (`EventoEPEC`); fora de SP não existe |

Os três últimos grupos são desvios **posteriores** ao desvio para o AN, e por isso vencem.
Inverter a ordem deixaria o insucesso na entrega no Ambiente Nacional. E a contingência SVC é
escolhida **antes** de tudo, então ela alcança o cancelamento mas **não** a manifestação.

> **A versão do evento é 1.00, sempre.** A SEFAZ publica um único `envEvento_v1.00.xsd`, e o
> `FPVersaoServico` do ACBr sai dele (`:3619`) — inclusive para os eventos da Reforma. Emitir
> "4.00" porque o componente está na 4.00 é rejeição de schema, e é o erro natural: todos os
> outros serviços acompanham a configuração. Entrou em `versaoFixaPorServico` junto com o
> `EventoEPEC`, e o `conferir-endpoints-dfe.py` achou os dois sozinho.

Achados do `revisor-go` nesta etapa, e o que saiu deles:

- **bloqueante — o `descEvento` dos eventos acentuados saía SEM acento.** A lista
  `EventoLevaAcentos` foi portada e testada, mas **nenhum gerador a consultava**: todo
  `descEvento` passava por `pcn.NodeStr`, que aplica `FiltrarTextoXML` → `TiraAcentos`. O
  evento 112120 saía com `Importacao em ALC/ZFM nao convertida em isencao` quando o ACBr emite
  o texto acentuado. Corrigido com `pcn.NodeStrComAcentos` e `nodeDescEvento`, e travado por
  teste que olha o **XML gerado** — afirmar sobre a lista ou sobre o mapa não bastava, e foi
  por isso que a lacuna passou.

  Junto veio uma **divergência nova**: no ACBr a flag `RetirarAcentos` é **global e não é
  restaurada** depois de ligada, então o primeiro evento acentuado de um lote contamina todos
  os seguintes. Isso depende da ordem de iteração, e o `CLAUDE.md` manda não replicar — aqui a
  decisão é por evento, com teste.
- **importante — `vNF`/`vICMS` do EPEC de NFC-e nunca eram lidos.** Na NFC-e os dois ficam
  **fora** do grupo `dest`, e o `dest` pode nem existir; o `Ler_DetEvento` do ACBr só os lê
  dentro do `Ler_Dest` (`:195`) e perde os dois no round-trip. O porte passou a lê-los também
  no nível do `detEvento`, com o `dest` **vencendo** quando existe. É **acréscimo** ao ACBr.
- **importante — EPEC e `detPag` sem nenhum teste.** São as duas estruturas mais ramificadas
  do gerador, e o `conferir-eventos-dfe.py` não as alcança (ele compara conjunto e ordem de
  tags, e não distingue ramos de condicional). Entraram nove testes: os dois ramos do EPEC por
  modelo, a condição composta do `dest` opcional da NFC-e, as duas portas do `idEstrangeiro`,
  e as quatro combinações dos dois blocos independentes do `detPag`.
- **importante — troca de `Find` por `FindAnyNs` em onze tags**, sem registro nem teste. O
  ACBr usa `FindAll` (com prefixo) para `retEvento`, `itemPedido`, `detPag` e os oito grupos da
  Reforma, e `FindAnyNs` nas demais — inconsistência dele, já que nada no leiaute justifica
  tratar `detPag` diferente de `autXML`. A troca é uniformemente mais permissiva; agora está
  registrada no doc-comment e travada por teste com documento prefixado.
- **menor — o comentário dizia "seis eventos" fixam `tpAutor`; são quatro** (`:490`, `:541`,
  `:586`, `:633`). O código estava certo; só a especificação do comentário estava errada, e é
  ela que o próximo porte vai usar.
- **menor — o EPEC decidia o modelo pelo complemento da condição.** O ACBr escreve
  `if sModelo = '55'`; o porte escrevia `== "65"`. Equivalente para 55 e 65, divergente para
  chave curta ou modelo desconhecido — onde o ACBr cai no ramo da NFC-e. Invertido para
  replicar o literal, com teste.
- **menor — `autXML` é `maxOccurs=1` no schema do 110150.** O ACBr gera todos e apenas alerta
  (`:322`); a SEFAZ rejeita. O porte emite só o **primeiro**.
- **menor — `idLote` lido como `int`** onde o Delphi usa `tcInt64` e o campo tem quinze
  dígitos. Corrigido com `pcn.ConteudoInt64`, e criado `pcn.NodeInt64` para o lado da escrita.
- **menor — `DetEvento.Versao` recebia o `verEvento`**, que é outro campo do leiaute. Separados:
  `InfEvento.VersaoEvento` é a tag, `DetEvento.Versao` é o atributo do `detEvento`.
- **menor — `ParseTipoEventoNFe("-99999")` devolve erro**, enquanto o ACBr o considera válido.
  Deliberado: aceitar o sentinela criaria um caminho em que evento sem tipo passa como se
  tivesse um. Registrado no doc-comment e no teste.

Mais três divergências que faltavam no registro:

- **`descricaoDoEvento` dá precedência ao `DetEvento.DescEvento` informado.** No ACBr o
  `DescEvento` é getter **somente leitura** (`EventoClass.pas:620`) e o campo do consumidor é
  ignorado. Aqui ele vence — é o que permite emitir um evento novo antes de o porte conhecê-lo.
- **Lote acima de 20 eventos vira erro.** O ACBr só emite `wAlerta` e **gera assim mesmo**
  (`:2522`).
- **O `nSeqEvento` do helper `Cancelamento` sai 1, não 0.** O `TACBrNFe.Cancelamento`
  (`ACBrNFe.pas:812`) deixa o campo em zero, e o `Id` termina em `"00"` — que a SEFAZ não
  aceita, porque a sequência começa em 1. É **correção de defeito determinístico** do ACBr.

> **Ferramenta nova:** `ferramentas/conferir-eventos-dfe.py` confere, por evento, o conjunto e
> a **ordem** das tags do `detEvento`, os literais fixos e as sete formas do `gControleEstoque`
> — seguindo os helpers dos dois lados. Na entrega: 31 eventos, 5 literais e 7 formas,
> idênticos ao Delphi.

### Divergências da DISTRIBUIÇÃO DF-e da NFe (2026-10-21)

Porte de `TDistribuicaoDFe` (`ACBrNFeWebServices.pas:3953`), `TDistDFeInt`
(`ACBrDFeComum.DistDFeInt.pas`) e `TRetDistDFeInt` (`ACBrDFeComum.RetDistDFeInt.pas`), mais o
enum `TSchemaDFe`, que foi para o `pcn` por ser compartilhado — o mesmo lote mistura documentos
de NF-e, CT-e, MDF-e, BP-e e GTV-e.

É o serviço pelo qual o contribuinte descobre os documentos emitidos **contra** ele. Nenhum
outro serviço da NF-e faz isso: todos os demais partem de um documento que o próprio emissor
já tem.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `StrToSchemaDFe` (`ACBrDFe.Conversao.pas:1156`) | **levanta exceção** quando o nome não está no enum. Num lote de cinquenta documentos isso derruba a leitura inteira porque um deles trouxe schema novo — e a RFB acrescenta schema sem avisar | `ParseSchemaDFe` devolve `SchemaDesconhecido` **e** um erro, mas o `lerDocZip` **descarta o erro** de propósito: o atributo `schema` não decide nada (quem decide é a raiz), então um valor novo não é problema. O item fica com `SchemaDesconhecido` no campo — é ali que se diagnostica — e os outros quarenta e nove seguem |
| 2 | `docZip` corrompido | o `LerXml` tem um `except` que devolve `False` e **descarta o lote inteiro** | o item ganha `ErroLeitura` e os vizinhos continuam lidos |
| 3 | varredura | não existe. O ACBr expõe só a chamada única | `VarrerDistribuicaoDFe` percorre até o fim. A RFB devolve no máximo cinquenta `docZip` por chamada, então a chamada única quase nunca traz tudo |
| 4 | `cOrgao` do `procEventoNFe` | o campo existe no struct e **não é lido** (`:781-812`) | lido. A omissão parece descuido, não decisão |
| 5 | gravação em disco, `GerarPathDistribuicao`, `GerarNomeArquivo` | monta caminho e grava | fora do escopo — gravar é do consumidor |
| 6 | nível do `detEvento` no evento processado | procura o grupo como filho **direto** da raiz `procEventoNFe` (`:850`), e o `FindAnyNs` só percorre filhos diretos. O caminho real do leiaute é `procEventoNFe/evento/infEvento/detEvento` — ou seja, o grupo **nunca** é preenchido para a NF-e | lido do nível certo, com a busca na raiz mantida como retrocesso |
| 7 | `XML` do item com raiz desconhecida | só é atribuído **dentro** de cada `Ler*` (`:513, :537, :634, :791`), então raiz desconhecida fica com `XML` vazio e o `TratarResposta` pula o item (`:4072`) | guardado sempre. É o que permite arquivar um documento de tipo que o porte ainda não sabe ler |
| 8 | base64 do `docZip` com quebra de linha | o `Decode4to3Ex` (`synacode.pas:609-613`) testa `y in [33..127]` e por isso **ignora** espaço, CR e LF em qualquer posição | o `base64` do Go recusa qualquer byte fora do alfabeto. O `dfe.Base64Gunzip` descarta brancos antes de decodificar, replicando o original. Sem isso, base64 quebrado em colunas — o que sai de pretty-printer e de arquivo gravado por outra ferramenta, que é o caso do `LerXMLFromFile` — perderia o documento em silêncio |
| 9 | chave no `<consChNFe>` | emite a chave **crua**, como veio do consumidor | passa por `pcn.RemoverLiteralChave`. Uma chave copiada do atributo `Id` vem com o prefixo `NFe`, e a RFB rejeita o pedido inteiro. Nenhum caso em que o literal deva ir para a SEFAZ |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **A precedência das três formas de consulta está em dois `if` aninhados** (`:154-171`):
  o **NSU** vence tudo; faltando ele, a **Chave** vence o **ultNSU**. Perdê-la faz um pedido
  com NSU e ultNSU preenchidos varrer o lote inteiro no lugar de buscar o documento pedido.
- **A escolha entre `<CNPJ>` e `<CPF>` é pelo TAMANHO depois do preenchimento**, e o
  preenchimento tem dois alvos (`:133-152`): até 11 dígitos completa até 11 e sai como CPF;
  12, 13 ou qualquer outro completa até 14 e sai como CNPJ. Documento **vazio** cai no ramo do
  CPF e sai `<CPF></CPF>`, porque o tamanho continua zero e nenhum dos dois `if` dispara.
- **O NSU passa por `StrToIntDef(x, 0)` antes de ser preenchido com quinze zeros.** Um NSU não
  numérico vira **zero** em vez de ser recusado — numa varredura, isso faz o lote recomeçar do
  início. Determinístico, e o consumidor controla a entrada.
- **O tipo de cada item é decidido pelo NOME DA RAIZ do documento descomprimido** (`:817-855`),
  e **não** pelo atributo `schema` do `docZip`. Os dois concordam na prática, mas é a raiz que
  manda — e o atributo continua sendo lido e guardado.
- **O `resNFe` e o `procNFe` preenchem o MESMO struct por caminhos diferentes.** No resumo os
  campos vêm prontos; no documento completo saem de quatro grupos aninhados (`ide`, `emit`,
  `total/ICMSTot`, `protNFe/infProt`), e a chave sai do **atributo `Id`** pelas 44 posições
  finais — `RightStr(Id, 44)`, não remoção do prefixo "NFe".
- **A situação do documento não vem pronta no `procNFe`**: sai de um `case` sobre o `cStat` do
  protocolo (`:620`), e o `case` **não tem `else`** — código fora de `[100, 101, 110, 132]`
  deixa o campo no valor anterior.
- **O `dhEmi` cai para `dEmi` quando vazio** (`:563`): na versão 2.00 a data de emissão ainda
  não tinha hora. Sem o retrocesso, toda nota antiga vem com data zerada.
- **O 137 CONTA como sucesso** (`:4120`). "Nenhum documento localizado" é a resposta mais
  comum de uma varredura diária; tratá-la como falha faz o laço nunca terminar.

> **A mensagem tem DUAS camadas, e só uma vem do gerador.** O elemento do Body é
> `nfeDistDFeInteresse` (`:3983`); dentro dele vem um `nfeDadosMsg` que o próprio
> `TDistDFeInt` emite, porque o `tagGrupoMsg` do construtor é esse (`:4029`). Nenhum outro
> serviço da NF-e tem as duas. E não há `<soap12:Header>` em versão nenhuma — o `Clear` zera o
> `FPHeaderElement` (`:3984`).

> **O endereço não varia por UF.** A primeira linha do `DefinirURL` é `UF := 'AN'` e a
> configuração nem é consultada (`:4001`): o serviço é tratado diretamente pela RFB. É o único
> serviço da NF-e em que a contingência SVC não muda nada. O **modelo**, esse sim, sai da chave
> quando ela é informada.

> **A versão da mensagem é 1.01**, mesmo num componente 4.00 — quinto serviço a entrar em
> `versaoFixaPorServico`, e o `conferir-endpoints-dfe.py` o achou sozinho.

**Acréscimos sem equivalente no ACBr.** Nenhum deles muda o XML gerado nem o que se lê do
retorno; todos existem porque, sem eles, o consumidor reescreveria a mesma regra:

| o que | por quê |
|---|---|
| `RetDistDFeInt.Atendida()` | o critério de sucesso é `cStat ∈ {137, 138}` e está embutido no `TratarResposta` (`:4120`). O **137** — "nenhum documento localizado" — conta como sucesso, e quem o tratar como falha faz o laço nunca terminar |
| `RetDistDFeInt.VarreduraConcluida()` | o fim é `ultNSU == maxNSU`, **não** "o lote veio vazio". A comparação é numérica, pelo mesmo motivo do laço |
| `ChaveDoDocZip(d)` | os três grupos guardam a chave em campos diferentes, e qual está preenchido depende do schema |
| `SchemaDFe.EDocumento()` | distingue resumo de documento completo. O ACBr faz essa pergunta com `in [schprocNFe, schprocCTe, ...]` espalhado, não com uma função |

`VarrerDistribuicaoDFe` bater no limite de chamadas **não é erro** e não tem sentinela: o
retorno é o mesmo de uma varredura completa. Quem precisa distinguir olha o `RetDistDFeInt`
devolvido — `VarreduraConcluida()` falso com `err == nil` significa que ainda há documentos, e
o `UltNSU` desse retorno é de onde a próxima varredura recomeça.

**Conferência mecânica do leitor.** O `conferir-leitor-dfe.py` rodou contra o
`ACBrDFeComum.RetDistDFeInt.pas`: **40 tags conferidas, nenhuma divergência** de precisão nem
de `Find`/`FindAnyNs`. As 22 diferenças de cobertura que ele lista são todas explicáveis, e
foram conferidas uma a uma:

- **16 só no `.pas`** — `CNPJ` e `CPF`, que o porte lê pelo helper
  `pcn.ConteudoCNPJCPFAnyNs` (limite 2 do script), e catorze do **CT-e**: o grupo `CTe` do
  `detEvento`, o `vPrest`/`vTPrest` (valor total da prestação, equivalente do `vNF` lá) e os
  `itensAverbados` da averbação de exportação. O `LerGrupo_detEvento_CTe`,
  `LerGrupo_detEvento_emit` e `LerGrupo_detEvento_itensAverbados` **não foram portados** —
  pertencem ao CT-e e entram com ele.

  > O `LerGrupo_detEvento_emit` é **código morto no próprio ACBr**: o chamador já resolve o nó
  > com `ANode.Childrens.FindAnyNs('emit')` (`:748`) e a procedure procura `'emit'` **de novo**
  > lá dentro (`:701`). Só preencheria com um `detEvento/emit/emit`, que não existe em leiaute
  > nenhum. Quando o CT-e entrar, o grupo deve ser lido do nível certo — e isso é correção, não
  > replicação de defeito, porque o defeito é determinístico mas o campo é dado fiscal perdido;
- **6 só no `.go`** — `NFe`, `chNFe`, `infNFe`, `protNFe`, `eventoNFe` e `retEventoNFe`. No
  Delphi esses nomes são **montados** a partir do parâmetro `FptpDFe` (`'ch' + FptpDFe`,
  `'inf' + FptpDFe`, `'prot' + FptpDFe`), e o script só enxerga literal. O porte os escreve
  literalmente porque é o leitor da NF-e; o do CT-e escreverá `chCTe`, `infCte` e `protCTe`.

**Achado do próprio teste:** a guarda contra laço infinito comparava o NSU como **texto**, e
`"7"` ≠ `"000000000000007"`. A RFB devolve o NSU preenchido com quinze zeros e o consumidor
costuma guardar o número cru, então toda varredura que começasse de um NSU não preenchido
gastava uma chamada à toa. Corrigido para comparação numérica, nos dois lugares que comparam
NSU (`VarreduraConcluida` e o laço). O teste pegou porque o helper duplicava o laço — e a
duplicação virou achado também: o laço foi extraído para `varrer`, que recebe quem faz a
chamada, justamente para que o teste exercite o laço REAL e não uma cópia dele.


### Divergências do FORMATO .ini da NFe (2026-10-22)

Porte de `TNFeIniReader` e `TNFeIniWriter` (`ACBrNFe.IniReader.pas`, 1782 linhas, e
`ACBrNFe.IniWriter.pas`, 1543), mais a extensão de `packages/rtc/ini.go` com os grupos da
Reforma que a NF-e usa e a NFGas não usava.

O `.ini` é como a maioria dos ERP alimenta o ACBr. **Não tem schema publicado** — o contrato é
o próprio código —, e é por isso que o porte dele é especialmente sujeito a erro silencioso:
uma chave com nome trocado não quebra build, não quebra `vet`, e só aparece quando o cliente
reclama que o campo sumiu.

Três convenções do formato que escapam de quem o porta pela primeira vez:

| convenção | o que é | o que acontece se for ignorada |
|---|---|---|
| **alias de seção** | `[Identificacao]` ou `[ide]`, `[Emitente]` ou `[emit]`, `[Produto]` ou `[det]`, escolhidos pela seção que EXISTE no arquivo | metade dos arquivos em circulação é recusada |
| **alias de chave** | `cNF`/`Codigo`, `natOp`/`NaturezaOperacao`, `xLgr`/`Logradouro`, `cMun`/`CidadeCod` — cadeia aninhada, com ORDEM | o nome antigo vence o do leiaute, ou o campo some |
| **sentinela de fim** | não há contador: o grupo repetido termina quando a chave sentinela daquele grupo falta | o laço não termina, ou termina cedo demais |

A sentinela quase nunca é a primeira chave, e nem sempre é a mesma coisa:

- `det` — `Codigo` ou `cProd`, e **só o literal "FIM" encerra**; vazio NÃO encerra;
- `ICMS` — `CST`, caindo para `CSOSN` (o Simples Nacional não tem CST);
- `PISST` e `COFINSST` — a **concatenação** de duas leituras com defaults `"F"` e `"IM"`,
  comparada com `"FIM"`. É o teste "alguma destas duas chaves existe?" escrito de um jeito que
  só faz sentido olhando os defaults;
- `ISSQN` — a base de cálculo, e ainda tem de ser **positiva**;
- `detExport` — `nRE`, caindo para `chNFe`, caindo para `nDraw`;
- `volume` — **não tem sentinela**: o laço para quando os SEIS campos estão vazios. Um volume
  que traga só o peso continua sendo volume.

A largura do índice **varia por grupo** e errá-la não dá erro — a seção simplesmente não é
encontrada e o grupo sai vazio:

| forma | exemplo |
|---|---|
| 3 dígitos | `[det001]`, `[NFref001]`, `[ICMS001]` |
| 3 + 3 | `[NVE001001]`, `[rastro001001]`, `[Lacre001001]` |
| 3 + 1 | `[gCred0011]` — **um dígito só** |
| 3 + 2 | `[origComb00101]` |
| 2 | `[autXML01]`, `[defensivo01]`, `[gPagAntecipado01]` |
| 3 + 3 + 3 | `[LADI001001001]` |

`autXML` é o único grupo que aceita **duas larguras**: o leitor tenta três dígitos e, não
achando, dois.

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | modelo do ECF na referenciada | o gravador escreve a chave `modelo` (`:308`), o leitor procura `mod` ou `ModECF` (`:293`). O modelo do cupom **se perde** numa ida e volta | grava `mod`, que os dois leitores encontram. `modelo` não é lido por ninguém |
| 2 | `[Retirada]` | o gravador emite OITO campos e o leitor lê QUINZE (`:387-399`): xNome, CEP, cPais, xPais, Fone, Email e IE do local de retirada se perdem na ida e volta | grava os quinze. São campos do leiaute e o leitor do próprio ACBr os recupera |
| 3 | `gMono` do `IBSCBSTot` | o `Ler_IBSCBSTot` LÊ o grupo e o `Gerar_IBSCBSTot` **não o grava** — o total da monofasia inteiro se perde | gravado. O leitor do ACBr o recupera sem alteração do lado dele |
| 4 | `vTroco` do pagamento | gravado FORA do laço, na variável de seção que sobrou da última volta (`:1356`). Com a coleção VAZIA essa variável nunca foi atribuída e o Delphi grava numa seção de nome indefinido | vai na ÚLTIMA seção de pagamento — o que o original faz no caso definido — e não sai quando não há pagamento nenhum |
| 5 | data e decimal | `DateTimeToStr` e `WriteFloat` dependem da configuração regional da máquina: o mesmo documento gera arquivo diferente em cada instalação | data em ISO, decimal com ponto. A leitura aceita as duas formas |
| 6 | `Gerar_IBSCBS` do `.ini` | condiciona o bloco inteiro a `gIBSCBS.vBC > 0`. Um item monofásico (CST 620) tem esse vBC zerado — a base dele vive no `gIBSCBSMono` — e **nada** seria gravado: nem o CST, nem a monofasia | vale a condição do gerador de XML do próprio ACBr (CST e cClassTrib), para que o `.ini` e o XML do mesmo documento contenham o mesmo grupo |
| 7 | `gpBioDiferencaIBS`/`CBS` | UMA seção por tributo, gravada e lida pelos DOIS ramos (ad rem e ad valorem). Quem grava por último vence, e o outro ramo volta com o conteúdo dele | replicado — é determinístico. A seção deveria carregar o sufixo do ramo, como `gMonoRetIBSQtde` × `gMonoRetIBSAliq` carregam |

Replicado mesmo parecendo errado (**não** "corrigir"):

- **O modelo padrão é 65 (NFC-e), não 55.** Faltando o `mod`, toda nota sai como NFC-e.
- **`[Retirada]` e `[Entrega]` são condicionadas ao LOGRADOURO na leitura e ao CNPJCPF na
  escrita.** O local de retirada pode ser um endereço sem documento: é lido e não é gravado de
  volta.
- **`[Compra]` é condicionada só ao `xNEmp` na escrita**, embora o leitor aceite qualquer um
  dos três campos. Nota com pedido e sem empenho é lida e não é gravada de volta.
- **`[detExport]` só é gravado com `nDraw`**, embora o leitor aceite `nRE` ou `chNFe`.
- **O IPI depende do `cEnq`, não do CST.** Item com IPI tributado e sem código de
  enquadramento não grava o grupo.
- **`NVE` e `rastro` PARAM no primeiro item vazio** (`else Break`), em vez de pulá-lo: um
  vazio no meio da lista descarta os seguintes. Leitor e gravador concordam nisso.
- **O `nSeqAdi` cai para o PRÓPRIO índice do laço** quando ausente (`:709`) — único lugar do
  formato em que um campo do documento assume o número de ordem da seção.
- **Dois aliases do original têm erro de digitação** e estão preservados, porque arquivo
  antigo os traz: `ValorCONFINSISS` (COFINS do ISSQN, com o N trocado de lugar) e
  `ValorBaseFCPSTRes` (com RES no fim, não RET).
- **`[ICMSUFDEST]` é gravado com DEST em maiúsculas** e lido como `ICMSUFDest`. A comparação
  de seção ignora caixa nos dois lados, então fecha; a grafia está preservada para que o
  arquivo saia idêntico ao do ACBr.

> **A única condicional POR VERSÃO DE LEIAUTE do `.ini` da NF-e está no medicamento.** A 4.00
> trocou o grupo inteiro: saiu o controle de lote (`nLote`, `qLote`, `dFab`, `dVal`), que
> passou para o grupo `rastro`, e entraram `cProdANVISA` e `xMotivoIsencao`. O `vPMC` sai nas
> duas. Gravar os dois conjuntos sempre produz arquivo que o validador recusa — e é por isso
> que o gravador precisa receber a versão da nota, que vive em `InfNFe` e não no item.

> **PIS e COFINS escolhem os campos pelo CST, e o `case` não tem `else`.** O 01 e o 02 saem
> por alíquota (`vBC`, `pPIS`), o 03 por quantidade (`qBCProd`, `vAliqProd`), o 99 pelos dois;
> qualquer outro CST grava **só o CST**. `vBC` e `qBCProd` são excludentes no leiaute, então
> um gravador que emita os dois sempre produz `.ini` que a SEFAZ recusa — e o round-trip não
> acusa, porque o leitor aceita os dois.

**Extensão do `packages/rtc`.** O `TNFeIniReader` herda de `TDFeRTCIniReader`, e o `.ini` da
NF-e usa cinco grupos da Reforma que o porte só tinha pela metade. Entraram:

- **`gIBSCBSMono`** — a árvore inteira da monofasia, em `packages/rtc/ini_mono.go`. São quatro
  ramos com a mesma forma (ad rem e ad valorem, para IBS e para CBS), e "ad rem" cobra por
  QUANTIDADE enquanto "ad valorem" cobra por VALOR: `gMonoPadraoIBSQtde` contra
  `gMonoPadraoIBSAliq`, seções distintas com campos distintos. O `Ler_IBSCBS` chama
  `Ler_gIBSCBSMono` **incondicionalmente**, então o porte anterior perdia a monofasia em
  silêncio em qualquer documento que a trouxesse;
- **`ISel`/`ISTot`/`gMonoTot`** — o Imposto Seletivo. O `CSTIS` é lido como **string**, e não
  como enum: o ACBr deixou a conversão comentada esperando a tabela oficial de CSTs do IS;
- **`gCompraGov` completo** (o portado era o *Reduzido*) e **`gPagAntecipado` do `ide`**;
- **`DFeReferenciado`** do item.

> **O gravador do `IBSCBS` passou a receber o MODELO do documento.** O `Gerar_IBSCBS` do ACBr
> despacha por `ModelosDFe` e depois pelo CST — a mesma tabela que o gerador de XML já tinha —,
> e o porte do `.ini` a ignorava: gravava o que estivesse preenchido. A tabela foi extraída
> para `rtc.GrupoDoCST`, **compartilhada pelos dois geradores**, porque duas cópias divergem em
> silêncio: o build passa, o teste de cada lado passa, e o `.ini` deixa de casar com o XML do
> mesmo documento. As sete funções por modelo do `xml_writer.go` foram substituídas pela
> consulta à tabela.
>
> Efeito colateral em `packages/nfgas` e `packages/nfag`: com o despacho correto, os dois
> passam a gravar o cabeçalho do IBSCBS e **não** o `gIBSCBS` quando o CST não é 000 — o
> `Gerar_IBSCBSNFGas` e o `Gerar_IBSCBSNFAg` só emitem o grupo com esse CST. O arquivo de
> exemplo do próprio ACBr (`testdata/nfgas_acbr.ini`) usa CST 200, e a LEITURA não tem essa
> restrição: o grupo é lido inteiro e descartado na escrita. O teste de round-trip do `nfgas`
> afirmava o contrário e foi corrigido — ele codificava o contrato do porte anterior, não o do
> Delphi. O `nfag` não mudou de comportamento na prática, porque a fixture dele já usa CST 000;
> os dois modelos entraram na tabela de `TestINI_IBSCBSSubgrupoDependeDoCSTeDoModelo`.

**Conferência mecânica.** O `conferir-ini-dfe.py` extrai as CHAVES que cada lado lê e grava e
compara nos dois sentidos: **485 chaves lidas, idênticas**, e **380 gravadas**, com a única
divergência (`mod`, do ECF) declarada no comando com `--esperado`.

O script também aponta a assimetria DENTRO do próprio ACBr, que é onde o round-trip perde dado:
**113 chaves que ele lê e não grava** — quase todas os aliases em português, que são entrada e
não saída — e **8 que ele grava e não lê**: `ID`, `nItem` e os seis do `[procNFe]`. Os dois
primeiros são derivados na leitura (a chave vem dos campos, o número do item vem do índice do
laço); o `[procNFe]` é o protocolo de autorização, que o `.ini` emite para exportação e nunca
relê. Preservado nos dois casos — o `.ini` é um formato de ENTRADA para emissão, e ler o
protocolo de volta deixaria um valor obsoleto vencer o que a SEFAZ devolveu.

O que o script **não** confere, e por isso precisou de teste: nome e largura de seção (montados
em variável nos dois lados) e a condicional em volta de cada grupo. Foram **44 cortes
deliberados**, todos pegos por algum teste.

**O que a revisão de fidelidade pegou, e o round-trip não pegaria.** Um defeito **bloqueante** e
três regras sem rede:

- **o `detExport` se perdia inteiro.** A sentinela do grupo é uma **cascata** de três `if`
  (`:726-736`), e cada degrau testa também o COMPRIMENTO — não é a cadeia aninhada de
  `ReadString` que o resto do formato usa. O porte tinha achatado as duas formas no mesmo
  helper, que para na primeira chave PRESENTE, mesmo vazia. E o próprio gravador emite `nRe=` e
  `chNFe=` vazios sempre que há drawback, que é o caso mais comum do grupo: **um `.ini` gerado
  aqui deixava de ser lido aqui**;
- **`lerInteiroAlias` e `lerFloatAlias` tinham a semântica trocada.** No Delphi o inteiro
  encadeia no DEFAULT do `ReadInteger` (`StrToIntDef(ReadString(...,''), Default)`), então chave
  presente e vazia **cede para o alias**; o decimal encadeia nos `ReadString` dentro de um
  `StringToFloatDef`, então chave presente e vazia entrega **zero** sem consultar o alias. O
  porte fazia o oposto nos dois;
- **a sentinela concatenada do PISST/COFINSST estava parafraseada**, e a paráfrase perdia um
  caso: chave presente-e-vazia com a outra ausente dá `"" + "IM" = "IM"`, que não é `"FIM"` e
  tem comprimento maior que zero — o grupo É lido. Reescrita literalmente como a concatenação
  do original, que é mais curta E mais fiel;
- **a árvore da monofasia tinha metade dos campos zerados na fixture.** Dez campos cujo nome de
  chave nunca era exercitado, nos quatro ramos de nomes quase iguais (`adRemIBSReten` ×
  `adRemCBSReten`, `gMonoRetIBSQtde` × `gMonoRetIBSAliq`). Os quatro ramos estavam corretos,
  mas nada os segurava: agora a fixture preenche tudo e o round-trip compara os quatro.

**Mais três divergências, encontradas na revisão de fidelidade:**

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 8 | `gpBioDiferencaCBS` | o gravador emite a chave com o nome do **IBS** (`RTC.IniWriter.pas:896`: `WriteFloat(sSecao, 'vIBSDiferenca', gpBioDiferenca.vCBSDiferenca)`) e o leitor procura `vCBSDiferenca`. É cópia-e-cola do grupo do IBS, e o ACBr perde o campo na própria ida e volta | grava `vCBSDiferenca` nos dois lados. **A consequência precisa ser sabida**: um `.ini` gerado aqui e lido pelo ACBr devolve o campo, mas um gerado pelo ACBr e lido aqui não — a chave que ele gravou tem outro nome. Corrigir o lado que lê seria pior: passaria a aceitar a chave errada como contrato |
| 9 | `refDFeAnt` do `gCompraGov` | o `Ler_gCompraGov` **lê** a seção `[refDFeAntNNN]`, e o `Gerar_DFeRef` que a gravaria **não é chamado por ninguém** — nem pelo `Gerar_gCompraGov`, nem pelo gerador da NF-e | gravado junto com o `gCompraGov`, que é de onde a lista pertence |
| 10 | `Gerar_DFeReferenciado` | grava a seção **sem condição**: todo item de uma NF-e produz `[DFeReferenciadoNNN]` com `chaveAcesso=` e `nItem=0` | guarda de grupo vazio, como o `GravarINIISel`. O `.ini` sai menor que o do ACBr, nunca diferente no conteúdo |

E duas que o original faz e o porte não replica por decisão do `CLAUDE.md`:

- **`StrToInt` cru em três campos** — `ISSQN.cMunFG` (`:1239`), `cana.dia` (`:1681`) e
  `adi.nAdicao` (`:705`). O Delphi **levanta `EConvertError`** com chave ausente ou não
  numérica; aqui o valor cai para zero. Biblioteca não entra em pânico com entrada torta: numa
  importação em lote, um `.ini` torto viraria exceção e derrubaria os demais.
- **O `vTroco` fora do laço de pagamento** (`:1356`) usa a variável de seção que sobrou da
  última volta. Com a coleção vazia essa variável nunca foi atribuída — comportamento que
  depende de memória não inicializada, e por isso não replicado.

**Achado do próprio porte, fora do script.** A sentinela do item distingue chave **ausente** de
chave **vazia**, e a primeira versão tratava as duas igual:

```pascal
sProdID := ReadString(sSecao, 'Codigo', ReadString(sSecao, 'cProd', 'FIM'))
if sProdID = 'FIM' then break
```

Faltando as duas chaves o resultado é `'FIM'` e o laço acaba — e é assim que ele acaba, porque
seção inexistente cai nesse caso. Mas um `cProd=` **presente e vazio** devolve string vazia,
que não é `'FIM'`: o item entra na nota, e é a validação que o recusa depois.

O helper `lerStringAlias` devolve `""` nos dois casos, então a cadeia precisou ser escrita à
mão neste laço. A versão anterior compensava com um teste de `SecaoExiste` que criava um item
vazio a mais antes de encerrar. Os dois cortes — tratar vazio como fim, e voltar ao helper —
agora falham em teste.
