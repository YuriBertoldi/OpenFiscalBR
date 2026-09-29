# packages/nfgas

**NFGas — Nota Fiscal de Fornecimento de Gás Canalizado Eletrônica**, modelo **76**,
leiaute **1.00**, namespace `http://www.portalfiscal.inf.br/nfgas`.

Porte do componente Delphi **ACBrNFGas**. **Layer 3** — depende de `comum`, `pcn` e `rtc`.

## Escopo desta fase: leitura

| Capacidade | Estado |
|---|---|
| Ler XML avulso (`NFGas`) e processado (`nfgasProc` + protocolo) | ✔ completo |
| Importação em lote (XMLs concatenados, arquivo, diretório) | ✔ completo |
| Eventos (`procEventoNFGas`, `retEventoNFGas`, `eventoNFGas`) | ✔ completo |
| Retorno de consulta (`retConsSitNFGas`) | ✔ completo |
| Formato `.ini` do ACBr (leitura **e** escrita) | ✔ completo |
| Regras de negócio 226, 227, 247, 252 + chave de acesso (montagem, DV, concatenação) | ✔ completo |
| **Emissão** — gerar XML, assinar, transmitir, cancelar, QR-Code | ✖ stubs com a assinatura definitiva, devolvem `ErrNaoImplementado` |

## Uso — importação em lote

```go
notas, err := nfgas.LerLoteDiretorio(`C:\xmls\entrada`)
if err != nil {
    var lote *nfgas.ErrosLote
    if errors.As(err, &lote) {
        for _, f := range lote.Erros {
            log.Printf("documento com problema: %v", f) // arquivo, indice e chave
        }
    }
}
for _, nota := range notas { // as legiveis SEMPRE voltam
    fmt.Println(nota.ChaveAcesso(), nota.Situacao(), nota.NFGas.Total.VNF)
}
```

Um documento torto **não derruba o lote**: as notas legíveis voltam sempre, e as falhas vêm num
`*ErrosLote` com arquivo, índice e chave de cada uma.

Outros pontos de entrada: `LerXML`/`LerXMLString`/`LerBytes` (um documento),
`LerNota`/`LerNotaArquivo` (com XML original preservado), `LerLote`/`LerLoteBytes`/`LerLoteString`/
`LerLoteArquivo`, `LerEvento*`, `LerRetConsSit*`, `LerINI*`, `GravarINI*`. O tipo `Componente`
(equivalente a `TACBrNFGas`) agrega configuração + notas carregadas e receberá os web services na
fase de emissão.

## Mapa de origem Delphi → Go

| Unit ACBr | Arquivo Go |
|---|---|
| `ACBrNFGas.Classes.pas` | `classes.go` (~45 structs, campos com o nome exato da tag) |
| `ACBrNFGas.Conversao.pas` | `types.go` + `conversao.go` (22 enums com `String()` e `Parse()`) |
| `ACBrNFGas.Consts.pas` | `consts.go` |
| `ACBrNFGas.XmlReader.pas` | `xml_reader.go` (porte 1:1, método a método) |
| `ACBrNFGasNotasFiscais.pas` | `notas_fiscais.go` (lote) |
| `ACBrNFGas.IniReader.pas` / `IniWriter.pas` | `ini_reader.go` / `ini_writer.go` |
| `ACBrNFGas.ValidarRegrasdeNegocio.pas` | `regras_negocio.go` |
| `Servicos\ACBrNFGas.EventoClass.pas` / `RetEnvEvento.pas` | `evento.go` / `evento_reader.go` |
| `Servicos\ACBrNFGas.RetConsSit.pas` | `ret_cons_sit.go` |
| `ACBrNFGas.XmlWriter.pas` / `ACBrNFGasWebServices.pas` | `xml_writer.go` / `web_services.go` (**stubs**) |

## Fidelidade ao ACBr — o que foi replicado de propósito

Cada item tem teste travando o comportamento:

- **`imposto` achatado**: `ICMS00, ICMS10, ICMS20, ICMS40, ICMS41, ICMS51, ICMS60, ICMS70,
  ICMS90` são procurados **nessa ordem** e caem todos na mesma struct `ICMS`; `ICMS41` é
  procurado apesar de não existir no XSD; `pDif`/`vICMSOp`/`vICMSDif` de `ICMS51` não são lidos.
- **`indSemCST`** é lido do nó `imposto`, não do `ICMSxx`, e só quando tem conteúdo.
- **`Dest`**: `idOutros` só sem CNPJ/CPF; `idEstrangeiro` só sem ambos; os dois alimentam o
  **mesmo campo** (`IDEstrangeiro`), como no Delphi. `TagIDOrigem` registra de qual tag veio.
- **`indIEDest`** não é lido do XML — só do `.ini`.
- Precisões: `vItem`/`vProd` De10 em `prod` e De8 em `gProcRef`; `pFCP`/`pPIS`/`pCOFINS`/`pTaxa`
  De4; `qUnidContrat` De6; `vTarifAplic` De8; o resto De2.
- Grafias: tag `vRetCofins` no XML e chave `vRetCOFINS` no `.ini`; tag `gRespTec` para o grupo
  `infRespTec`; `gNF/serie` é string (≠ `ide/serie`, int).
- `CompetEmis`/`CompetApur`/`CompetFat` em `AAAAMM`, zeradas quando fora da faixa.
- `infCpl`: o leiaute admite 5 ocorrências, o ACBr lê **uma** (TODO aberto no fonte) — replicado,
  mas o campo já é `[]string`.
- `qrCodNFGas`: remoção de CDATA **só da primeira ocorrência** (`StringReplace` sem
  `rfReplaceAll`).
- `RetTrib.VBCIRRF` e os grupos vazios de validação (12 famílias) existem e não fazem nada — como
  no fonte.
- `enderCorresp` não lê `cPais`/`xPais`; `enderEmit`/`enderDest` leem.

## Divergências deliberadas em relação ao ACBr

Documentadas no código com `DIVERGENCIA` e cobertas por teste:

| Ponto | ACBr | Aqui | Por quê |
|---|---|---|---|
| `Prod.qFaturada` | `Integer` — **arredonda** a quantidade | `float64` | XSD declara decimal de 4 casas (`TDec_1100_1104`) e o próprio leitor do ACBr lê com `tcDe4`; quantidade de gás arredondada é dado corrompido |
| `Prod.cClass` | `Integer` — perde zeros à esquerda | `string` | XSD declara `xs:string` com pattern `[0-9]{7}`; o zero é significativo |
| `motDesICMS` (XML) | lê como inteiro e trata o **código como ordinal** (código 1→membro 2; 16 e 90 estouram o enum) | parse pelo código | é o que o leiaute define; no `.ini` o próprio ACBr já lê pelo código |
| Erros de leitura | exceção que aborta a nota | `error` com caminho/índice/chave | uma nota torta não pode derrubar a importação do lote |
| `CompetFat` | `StrToDate('01/MM/AAAA')`, dependente de locale | data montada direto | o mesmo XML era lido diferente conforme a máquina |
| Split do lote | busca de string por `</NFGas>`/`</NFGasProc>`/`</procNFGas>` | decoder XML real | mesmo resultado semântico, sem depender de grafia/espaçamento |
| Fuso de `dhEmi` etc. | descartado | preservado (`time.FixedZone`) | hora de parede idêntica campo a campo; a informação deixa de se perder |
| Getter `XMLAssinado` que assina | efeito colateral | não replicado; assinar é método explícito (stub) | getter que assina documento é armadilha |

Acréscimos ao porte (marcados `ACRESCIMO`): leitura do **evento enviado** dentro de
`procEventoNFGas` (única fonte da justificativa de cancelamento) e dos campos
`CNPJDest`/`emailDest`/`cOrgaoAutor` do retorno, que o ACBr declara e não lê.

## Cobertura — declarado × exercitado

- **Leitura XML**: todos os ~45 structs do documento são preenchidos pelo leitor; exceções
  herdadas do ACBr: `RetTrib.VBCIRRF` (nunca lido), `Dest.IndIEDest` (só via `.ini`),
  `Endereco.CPais/XPais` em `enderCorresp`.
- **`.ini`**: todas as seções do `TNFGasIniReader` no round-trip ler↔gravar. Da árvore RTC, o
  `.ini` cobre IBSCBS do item (inclusive `gTransfCred`, `gAjusteCompet`, `gCredPresOper` e
  `gCredPresIBSZFM`), IBSCBSTot, pgtoVinc, gCompraGov e gPagAntecipado; a **monofasia
  (`gIBSCBSMono`) e o Imposto Seletivo só são lidos/escritos em XML**, não em `.ini`.
- **Emissão**: 0% — stubs declarados, honestamente.

## Testes

```bash
go test ./packages/nfgas/... -v
```

76 funções de teste (108 execuções com subtestes), stdlib pura, fixtures em `testdata/`
(inclusive um `.ini` no formato pt-BR do Delphi, com vírgula decimal e data `dd/mm/aaaa`).
