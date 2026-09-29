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
- URLs: MA e PA apontam para seções `NFGas_SVAN_*` que **não existem** no
  `ACBrNFGasServicos.ini` → `ErrSemURL` (lacuna herdada, documentada).

### Omissões deliberadas (geração)

- `ListaDeAlertas`/`wAlerta` não portada — validação efetiva é o XSD da SEFAZ +
  `regras_negocio.go`.
- `NormatizarMunicipios` (lookup em arquivo de municípios) não portada.
