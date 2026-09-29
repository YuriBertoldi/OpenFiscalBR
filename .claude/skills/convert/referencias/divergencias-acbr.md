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
