# packages/nfag

**NFAg — Nota Fiscal de Fornecimento de Água Canalizada Eletrônica**, modelo **75**,
leiaute **1.00**, namespace `http://www.portalfiscal.inf.br/nfag`.

Porte do componente Delphi **ACBrNFAg**. **Layer 3** — depende de `comum`, `pcn`, `rtc`
e `dfe`. Nasceu como irmão do `packages/nfgas` (mesma família de notas de utilities) e
compartilha com ele toda a infraestrutura de assinatura/SOAP; o **modelo de dados é próprio**
da água — não é um rename.

## Escopo

| Capacidade | Estado |
|---|---|
| Ler XML avulso (`NFAg`) e processado (`NFAgProc`/`nfagProc` + protocolo) | ✔ completo |
| Importação em lote (XMLs concatenados, arquivo, diretório) | ✔ completo |
| Eventos (`procEventoNFAg`, `retEventoNFAg`, `eventoNFAg`) | ✔ completo |
| Retorno de consulta (`retConsSitNFAg`) e de status (`retConsStatServNFAg`) | ✔ completo |
| Formato `.ini` do ACBr (leitura **e** escrita) | ✔ completo |
| Regras de negócio 226, 227, 247, 252 + chave de acesso | ✔ completo |
| **Geração de XML** (`GerarXML`, `GerarXMLProc`, `GerarXMLEvento`, `GerarXMLProcEvento`) | ✔ completo |
| **Assinatura** XMLDSig (`Assinar`, `AssinarEvento`, via `packages/dfe`) | ✔ completo |
| **Transmissão à SEFAZ** — recepção síncrona, consulta, status, cancelamento | ✔ completo |
| QR-Code e URL de consulta pública | ✔ completo |

## Uso — emissão

```go
c := nfag.NovoComponente()
c.Configuracoes.UF = "SP" // toda UF exceto MA/PA usa o SVRS
c.Configuracoes.Certificado, _ = dfe.CarregarPFXArquivo("emitente.pfx", "senha")

nota := &nfag.NotaFiscal{NFAg: montarNota()}
ret, err := c.Enviar(ctx, nota) // gera, assina e transmite (síncrono, 1 nota)
if err == nil && nota.NFAg.Processada() {
    xmlProc, _ := nfag.GerarXMLProc(nota.NFAg)
}
```

**Atenção ao critério de sucesso da recepção**: na NFAg o retorno síncrono é considerado
autorizado com **cStat do retorno = 104** ("Lote processado") + protocolo processado —
diferente da NFGas, que usa 100. É o que o `TNFAgRecepcao.TratarResposta` faz, replicado
e testado.

## O que a NFAg tem de diferente (é do leiaute, não do porte)

- `det` **achatado**: `gTarif`/`prod`/`imposto`/`gProcRef` direto no item (sem
  `gNormal`/`gAgregadora`).
- `imposto` **sem ICMS** (e sem `orig`/`indSemCST`): só IBSCBS, PIS, COFINS, retTrib e as
  taxas **TFS** e **TFU**.
- Grupos próprios de água: `ligacao`, `gFatConjunto`, `gQualiAgua` (análises de qualidade)
  e `infPAA`.
- `gMed` tem só identificação e datas; as leituras (vMedAnt/vMedAtu/vConst/vMed) ficam no
  `gMedida` do item. `gTarif` não tem `vTarifAplic`. `gCons` usa `qtdDias` **string** e
  `volFat`.
- `ligacao`, `gFat` e `gAgencia` são gerados **sempre** (sem os guards da NFGas); em
  homologação o `xNome` do destinatário é trocado pelo texto fixo de homologação.
- O Id de `infNFAg` leva o literal **NFAG maiúsculo** (pattern do XSD) e as ações SOAP têm
  grafia própria (`nfagStatusServicoNF`, `nfagConsultaNF`...), com URLs em
  `nfag.svrs.rs.gov.br/ws/...` (path minúsculo).
- Quatro tipos de evento no leiaute de retorno (cancelamento 110111, autorização de
  substituição 240140, ajuste 240150, liberação de prazo 240170); o **gerador** só
  implementa o cancelamento, como o ACBr.

## Mapa de origem Delphi → Go

| Unit ACBr | Arquivo Go |
|---|---|
| `ACBrNFAg.Classes.pas` | `classes.go` |
| `ACBrNFAg.Conversao.pas` | `types.go` + `conversao.go` |
| `ACBrNFAg.Consts.pas` | `consts.go` |
| `ACBrNFAg.XmlReader.pas` | `xml_reader.go` (porte 1:1) |
| `ACBrNFAg.XmlWriter.pas` | `xml_writer.go` (porte 1:1) |
| `ACBrNFAgNotasFiscais.pas` | `notas_fiscais.go` |
| `ACBrNFAg.IniReader.pas` / `IniWriter.pas` | `ini_reader.go` / `ini_writer.go` |
| `ACBrNFAg.ValidarRegrasdeNegocio.pas` | `regras_negocio.go` |
| `Servicos\ACBrNFAg.EventoClass.pas` / `RetEnvEvento.pas` / `EnvEvento.pas` | `evento.go` / `evento_reader.go` / `evento_writer.go` |
| `Servicos\ACBrNFAg.RetConsSit.pas` | `ret_cons_sit.go` |
| `ACBrDFeComum.RetConsStatServ.pas` | `ret_status.go` |
| `ACBrNFAgWebServices.pas` | `web_services.go` |
| `ACBrNFAgServicos.ini` | `urls.go` (tabela embutida; MA/PA → SVAN sem URL = `ErrSemURL`) |

## Divergências deliberadas em relação ao ACBr

Registradas em `.claude/skills/convert/referencias/divergencias-acbr.md`, cada uma com
teste:

| Ponto | ACBr | Aqui | Por quê |
|---|---|---|---|
| Raiz do proc na GERAÇÃO | `NFAgProc` | `nfagProc` | é o que o XSD (`procNFAg_v1.00.xsd`) define; o leitor aceita as três grafias |
| `retTrib` do item | grava `vRetCOFINS` | grava `vRetCofins` | grafia do XSD (`nfagTiposBasico_v1.00.xsd:925`); a maiúscula é inválida e o valor se perderia na releitura |
| `Prod.cClass` / `Prod.qFaturada` | `Integer` | `string` / `float64` | tipo do XSD; o ACBr perde zeros à esquerda / arredonda |
| `IndDevolucao` | `TIndicador` (tiSim = ordinal ZERO) | `pcn.IndicadorEx` | zero-value seguro (mesma decisão da NFGas) |
| Competência com data zero | `189912` | tag vazia | determinístico e claramente lixo |
| `cNF` aleatório | 8 dígitos | 7 dígitos | a posição 36 da chave é o `nSiteAutoriz` |
| `Ler_Dest` sem guard | access violation com `<dest>` ausente | leitura tolerante | biblioteca não estoura |
| `CompetFat`/`CompetAnalise` | `StrToDate` dependente de locale | data montada direto | mesmo XML lido igual em qualquer máquina |
| `.ini`, `gPagAntecipadoNNN` do item | writer 0-based × reader 1-based — round-trip do ACBr perde o grupo | writer 1-based, igual ao reader | perde dado |
| `.ini`, `dhEmi`/`dhCont` | `DateTimeToStr` (formato do locale) | `AAAA-MM-DDTHH:MM:SS` fixo, hora preservada | mesmo `.ini` lido igual em qualquer máquina |

Notas de estilo do `.ini` gerado (round-trip equivalente, mas **não** byte-comparável ao do
ACBr): chaves de valor vazio são omitidas onde o ACBr as grava vazias (`dhCont`, `xJust`,
`idOutros`, `chNFAgAnt`...). O `TEventoNFAg.LerFromIni` (seções `[EVENTOnnn]`,
`EnvEvento.pas:386-431`) **não foi portado** — evento entra pela API (`Cancelamento`) ou por
XML. O `Cancelamento` não faz a consulta prévia da nota que o `TACBrNFAg.Cancelamento` faz
para descobrir o protocolo — o `nProt` é responsabilidade do chamador.

### Bugs do ACBr REPLICADOS (determinísticos — não "corrigir")

- **INI**: a versão é gravada em `[infNFAg]` e lida de `[infNFGas]` (copy-paste da NFGas) —
  o round-trip do próprio ACBr perde a versão; o **TFU é gravado na seção `[TFSNNN]`** e
  lido de `[TFUNNN]` — TFU gravado nunca é relido. Os dois cobertos por
  `TestINI_RoundTripComBugsDoACBr`.
- `tpFat` **não é lido do XML** (o `Ler_Ide` não tem a tag); só entra via `.ini`.
- `vBCIRRF` é **gerado** e nunca **lido**; assimetrias de precisão writer×reader do próprio
  ACBr replicadas (prod `vItem`/`vProd`: leitura De10, geração De2; `pPIS`/`pCOFINS`:
  leitura De4, geração De2; `medMensal`/`consumo`: leitura De4, geração De2).
- `DescricaoTipoEvento` devolve `"CANCELAMENTO DE NF3-e"` (typo do ACBr; só descritivo).
- `IdentificaSchema` do ACBr procura `<infNFGas` (typo inofensivo — o default já é
  `schNFAg`); aqui a busca é pela raiz, com o mesmo resultado.

### Acréscimos ao porte

Marcados `ACRESCIMO` no código: leitura do **evento enviado** dentro de
`procEventoNFAg` (única fonte da justificativa de cancelamento) e dos campos
`CNPJDest`/`emailDest`/`cOrgaoAutor` do retorno, que o ACBr declara e não lê; e
`GerarXMLProcEvento`, que **escreve** o `procEventoNFAg` — o `TRetEventoNFAg` do
ACBr só tem leitura, e sem isso não há como produzir o envelope que o
contribuinte arquiva.

## Testes

```bash
go test ./packages/nfag/... -v
```

Stdlib pura e sem rede (SEFAZ simulada com `httptest`, certificado RSA gerado em teste).
Round-trip gerar→ler→gerar **byte a byte estável** — foi esse teste que flagrou a grafia
inválida `vRetCOFINS` do gerador do ACBr. Fixtures em `testdata/` (inclusive `.ini` pt-BR
com vírgula decimal e o cenário dos dois bugs de INI).
