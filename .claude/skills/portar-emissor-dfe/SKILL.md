---
name: portar-emissor-dfe
description: "Porta a EMISSAO de um componente DFe do ACBr (NFe, CTe, NF3e, NFCom, NFGas, BPe...) para Go: gerador de XML (XmlWriter), assinatura XMLDSig, web services SOAP da SEFAZ, URLs e QR-Code. Cobre a semantica do AddNode, o estado do writer da RTC, o padrao de teste round-trip gerar-ler-gerar, os bugs recorrentes do ACBr na geracao (enum como ordinal, blocos duplicados, tiSim no ordinal zero) e a infraestrutura pronta do packages/dfe. Use ao portar a emissao/transmissao de qualquer DFe, ou ao revisar um writer ja portado. Complementa a /convert e a /portar-leitor-dfe."
argument-hint: "<ComponentName> (ex.: ACBrNF3e)"
---

# portar-emissor-dfe — emissão de DFe, do .pas para Go

Receita destilada do porte da emissão do **ACBrNFGas** (`packages/nfgas` + `packages/dfe`),
que é o exemplar de referência — na dúvida sobre qualquer padrão, abra o arquivo
correspondente lá. Pré-requisito: a LEITURA do componente já portada (`/portar-leitor-dfe`) —
o teste mais forte da geração depende do leitor.

## O que JÁ EXISTE e não se reimplementa

| Infra | Onde | Porte de |
|---|---|---|
| `AddNode` por tipo de campo (`NodeStr/NodeInt/NodeDec/NodeDat`, obrigatório+vazio=tag vazia, opcional+vazio=nil, decimais FIXOS, PadLeft) | `pcn/xml_builder.go` | `TACBrXmlWriter.AddNode` |
| Construtor `Elem` (attrs em ordem de inserção, `Filho(nil)` inofensivo, `TextoBruto` p/ CDATA) | `pcn/xml_builder.go` | escrita de `TACBrXmlNode` |
| `FormatarDataHoraXML(t, uf)` = `DateTimeTodh + GetUTC(UF)` | `pcn/xml_builder.go` | `ACBrUtil.DateTime` |
| Certificado A1 (PFX/PEM), `AssinarXML`, `VerificarAssinatura`, C14N 1.0, `ClienteSOAP` soap12, `GzipBase64`, `HashCSRT` | `packages/dfe` | `TDFeSSL` + `TDFeWebService` |
| Grupos da Reforma Tributária (writer com estado) | `rtc/xml_writer.go` | `TDFeRTCXmlWriter` |

O `dfe` é o único package que importa `net/http` (CLIENTE — exceção documentada no CLAUDE.md
e no check da `/validar-package`).

## PASSO 1 — Ler os fontes de emissão

`Base/<Comp>.XmlWriter.pas` (o contrato), `<Comp>WebServices.pas` (ações/DadosMsg/
TratarResposta), `Base/Servicos/<Comp>.EnvEvento.pas` (gerador de evento), `<Comp>Servicos.ini`
(URLs por UF), `<Comp>.pas` (GetURLQRCode/GetURLConsulta). Anote de cada `Gerar_*`: ordem dos
campos, tipo `tcDeN` POR CAMPO, ocorrência 0/1, e a CONDICIONAL em volta de cada bloco.

**Se existe um IRMÃO da mesma família já portado** (NFGas ↔ NFAg ↔ NF3e ↔ NFCom — utilities
com a mesma arquitetura de fontes), comece pelo diff estruturado unit a unit:
`diff --strip-trailing-cr -w -i irmao.pas novo.pas` (os `.pas` são CRLF; sem `--strip-trailing-cr`
o diff acusa 100% das linhas). O delta é o trabalho real; `urls.go`/`web_services.go`/
`notas_fiscais.go` saem por cópia+ajuste. MAS o modelo de dados é próprio de cada documento
(o det da NFAg é achatado, o imposto não tem ICMS...) — `classes.go` e os writers/readers se
validam struct a struct contra o `.pas` novo, nunca por rename do irmão.

## PASSO 2 — Bugs do ACBr que se repetem em TODO writer (conferir um a um)

1. **Enum cru no AddNode → grava o ORDINAL.** Todo campo onde o `.pas` passa o enum sem
   `XxxToStr` produz XML inválido. Corrigir para `String()` (código do leiaute), marcar
   `DIVERGENCIA`, testar. Na NFGas: finNFGas, indOrigemQtd, tpMotNaoLeitura, tpProc, modBCST,
   motDesICMS (este via tcInt).
2. **`tiSim` é o ordinal ZERO de `TIndicador`** — objeto recém-criado nasce "sim". Campo
   indicador de writer em Go usa `pcn.IndicadorEx` (zero = não informado), nunca
   `pcn.Indicador`; senão regerar uma nota lida descarta grupos em silêncio.
3. **Blocos duplicados** (copy/paste no `.pas`): o ICMS70 da NFGas gera `vICMSDeson` duas
   vezes. Gerar uma; documentar.
4. **`Result := nil` no case-else + `AppendChild` em seguida** = access violation no Delphi.
   Em Go o grupo sai nil (o `Filho` nil-safe absorve) — documentar.
5. **Raiz do proc** pode divergir do XSD (`NFGasProc` × `nfgasProc`). O XSD manda.
6. **`FormatDateTime` de data zero** produz `1899...`. Data zero → tag vazia (divergência
   determinística).
7. **`FormatFloat('00')` ANTES de AddNode opcional** torna o campo incondicional ("00" nunca é
   vazio). REPLICAR — é determinístico (caso nContrat do gMedicao).
8. **`GerarCodigoDFe` sorteia com 8 dígitos**; se a chave do componente tiver campo extra
   (nSiteAutoriz da NFGas), o cNF é de 7 — corrigir o sorteio.

## PASSO 3 — Estado do writer

- O writer do componente carrega `FChave<Comp>` e MUTA o documento (`ID`, `cDV`, `cNF`) —
  replicar os efeitos colaterais.
- O `rtc.Writer` carrega `pRedutor/tpEnteGov/gerarIBSCBSTot`: um Writer NOVO por documento;
  `GerarGCompraGov*` sempre antes dos itens; `GerarIBSCBSTot` só depois dos itens.
- Assinatura pré-existente (`Signature` lida) é reembutida na geração (template
  `TSignature.GerarXML`); assinar de verdade é `dfe.AssinarXML(cert, xml, "inf<Comp>")`.

## PASSO 4 — Web services

- Envelope soap12 e ações: `dfe.MontarEnvelope` + tabela própria (`urls.go` embute o
  `<Comp>Servicos.ini`; seções `Usar=` que apontam para seção inexistente = UF sem serviço →
  erro sentinela, não pânico).
- Recepção pode ser síncrona (NFGas/NF3e/NFCom: 1 documento, `GzipBase64`, limite 1 MB,
  resposta `ret<Comp>` renomeada para o leitor de consulta) ou por lote/recibo (NFe) — copie o
  fluxo do `.pas`, não o de outro componente.
- **Conjuntos de cStat são POR COMPONENTE — copie do `TratarResposta` do `.pas`, nunca do
  irmão.** Sucesso da recepção síncrona: NFAg exige cStat do RETORNO = 104; NFGas usa 100.
  Evento registrado: `[135, 136, 155]` (o 155 — cancelamento fora de prazo — some fácil).
  Confirmada/Processada/Cancelada idem.
- Config de teste: `Configuracoes.URLs map[Servico]string` para apontar ao `httptest`.

## PASSO 4b — IniWriter (onde a revisão do NFAg achou TODOS os defeitos)

O `Gerar_*` do `IniWriter.pas` tem as mesmas armadilhas de condicional do XmlWriter, e o
build/vet não pega nenhuma:

1. **Guard (`Exit`) de cada seção, literal**: `Gerar_gMedicao` só pula com
   `(nMed <= 0) AND (vMed = 0)`; `Gerar_Ligacao` pula com `idLigacao` vazio. Simplificar o
   guard descarta grupo em silêncio.
2. **`DateTimeToIni` = `DateTimeToStr` — data E HORA.** Gravar só a data zera a hora de
   `dhEmi`/`dhCont` no round-trip (formato fixo `AAAA-MM-DDTHH:MM:SS`, não o do locale).
3. **Base do índice writer × reader pode divergir no próprio ACBr** (`gPagAntecipado`:
   writer 0-based, reader 1-based — o round-trip do ACBr perde o grupo). Perde dado →
   corrigir para a base do reader + registrar em `divergencias-acbr.md`.
4. **`raise` de validação no `GravarIni`** (`ValidarChave` → "Chave Inválida") vira `error`.
   Atenção: o IniReader NÃO lê o `ID` — teste de round-trip a partir de `.ini` precisa montar
   o ID antes de regravar.
5. Bugs determinísticos de seção trocada se REPLICAM (versão em `[infNFAg]` lida de
   `[infNFGas]`; TFU gravado em `[TFSNNN]`) — com teste citando o `.pas`.

## PASSO 5 — Testes que provam a geração

1. **Round-trip gerar→ler→gerar byte a byte estável** (o mais forte; pega campo gerado que o
   leitor não lê e vice-versa). Atenção: perda FIEL (ex.: det com gNormal+gAgregadora gera só
   gAgregadora) passa nesse teste — confira as perdas contra o `.pas`.
2. Assinatura self-verify: `PrepararEnvio` → `dfe.VerificarAssinatura` (cert RSA gerado no
   teste).
3. SOAP contra `httptest`: envelope, namespace do DadosMsg, gzip+base64 decodificado e
   assinatura do XML transmitido verificada.
4. **Round-trip do INI** (gravar→ler→comparar): pega guard simplificado, hora perdida e
   índice de seção divergente — foi o que faltou no NFAg e a revisão cobrou.
5. Um teste por divergência aprovada, citando o `.pas`.

## PASSO 6 — Fechar

`/validar-package`, subagente `revisor-go` (só emissão), README (seções de emissão +
divergências), `.openfiscalbr-meta.json` (hashes dos `.pas` novos),
`divergencias-acbr.md` (registro oficial — impede "correção" reversa).
