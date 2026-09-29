# packages/rtc

Grupos da **Reforma Tributária sobre o Consumo** (IBS, CBS, Imposto Seletivo, compra
governamental, pagamento vinculado) compartilhados por todos os documentos fiscais eletrônicos.

**Layer 1.** Depende de `packages/pcn`.

## Componente Delphi correspondente

| Unit ACBr | Arquivo Go |
|---|---|
| `ACBrDFe/ACBrDFe.RTC.Classes.pas` (1.709 linhas) | `classes.go` |
| `ACBrDFe/ACBrDFe.RTC.XmlReader.pas` (739 linhas) | `xml_reader.go` |
| `ACBrDFe/ACBrDFe.RTC.IniReader.pas` / `IniWriter.pas` (parcial) | `ini.go` — IBSCBS do item (com gTransfCred/gAjusteCompet/gCredPresOper/gCredPresIBSZFM), IBSCBSTot, pgtoVinc, gCompraGov, gPagAntecipado; **fora do `.ini`**: monofasia (`gIBSCBSMono`) e Imposto Seletivo |
| enums de RTC em `ACBrDFe.Conversao.pas` | `conversao.go` |

Existe como package próprio porque a mesma árvore é consumida por **NFe, CTe, MDFe, BPe, NF3e,
NFCom e NFGas** — no ACBr ela também é uma unit compartilhada, não parte de um componente.

## Uso

```go
doc, _ := pcn.ParseString(xml)
imposto := doc.Root.FindAnyNs("infNFGas").FindAnyNs("det").
    FindAnyNs("gNormal").FindAnyNs("imposto")

var ibscbs rtc.IBSCBS
rtc.LerIBSCBS(imposto.Find("IBSCBS"), &ibscbs)

fmt.Println(ibscbs.CST, ibscbs.GIBSCBS.GIBSUF.VIBSUF)
```

Todos os grupos são **structs por valor**, não ponteiros: no Delphi os construtores sempre
instanciam os subgrupos, então eles nunca são `nil`. O valor zero já é navegável.

## Enums

| Enum | Membros | Observação |
|---|---|---|
| `TpEnteGov` | 7 | `TcgNenhum` = `""` |
| `TpOperGov` | 5 | `TogNenhum` = `""` |
| `CSTIBSCBS` | 19 | códigos de 3 dígitos |
| `TpALCZFMCBS` | 2 | **não tem membro vazio** — começa em `"1"` |
| `CCredPres` | 14 | `CpNenhum` = `""` |
| `TpCredPresIBSZFM` | 6 | **`TcpSemCredito` vale `"0"`**, não `"1"` — há um membro "nenhum" antes |

## Fidelidade ao Delphi

Regras do original preservadas de propósito, todas com teste:

| Ponto | Comportamento |
|---|---|
| Busca de elemento | sempre `Find` (nome qualificado), **nunca** `FindAnyNs` — exceto `LerGPagAntecipadoProd`, o único que usa `FindAnyNs` |
| `node == nil` | sai sem tocar na struct (`if not Assigned(ANode) then Exit`) |
| `gpBioDiferenca` em ad valorem | **não é lido** por `LerGIBSMonoAdValorem` nem `LerGCBSMonoAdValorem`, embora o campo exista — omissão do ACBr, replicada |
| `vIBSMonoRet` / `vCBSMonoRet` | `De4` no item, `De2` no total — mesma tag, precisões diferentes |
| `gIBSUF` / `gIBSMun` no total | mesmas tags do item, **structs diferentes** (`GIBSUFTot` ≠ `GIBSUFValores`) |

### Divergências deliberadas

| Ponto | ACBr | Aqui | Por quê |
|---|---|---|---|
| `pgto/@nPag` ausente | `StrToInt` levanta exceção | vale `0` | atributo ausente não pode derrubar a importação de um lote |
| `gPagAntecipado` | não limpa a lista e indexa `refNFe[i]` pelo índice do laço | lista reinicializada, itens acrescentados em ordem | o original só é definido em leitura limpa |
| `competApur` (`AAAA-MM`) | concatena `"-01"` e usa `StringToDateTime` dependente de locale | monta a data direto | resultado idêntico, sem depender da máquina |

## Cobertura

| Área | Structs declarados | Exercitados por leitor |
|---|---|---|
| IBS/CBS do item | 17 | 17 |
| Monofasia | 15 | 15 |
| Totais | 7 | 7 |
| Compra governamental / pagamento antecipado | 6 | 6 |
| Imposto Seletivo | 2 | 2 |
| Pagamento vinculado | 2 | 2 |

**48 structs declarados, 48 alcançados por algum `Ler*`.** Os dois campos `gpBioDiferenca` dos
grupos ad valorem são a única exceção — declarados e nunca preenchidos, por omissão do ACBr,
com teste travando o comportamento.

Nada aqui gera XML — a geração entra junto com a emissão.

## Testes

```bash
go test ./packages/rtc/... -v
```

## Geração de XML (`xml_writer.go`)

Porte 1:1 de `TDFeRTCXmlWriter` (`ACBrDFe.RTC.XmlWriter.pas`). O `Writer` carrega **estado**
entre chamadas, como o original — use um `Writer` novo por documento:

- `GerarGCompraGovReduzido`/`GerarGCompraGov` capturam `pRedutor`/`tpEnteGov` ANTES de decidir
  se o grupo sai — o estado força `gRed` nos itens e habilita `gTribCompraGov`;
- `GerarIBSCBS` liga o flag interno que autoriza `GerarIBSCBSTot` no total do documento;
- o corpo do grupo IBSCBS varia por `ModeloDFe` (na NFGas, só CST 000 gera `gIBSCBS`).

Round-trip writer→reader coberto em `xml_writer_test.go`. Divergência defensiva: os
`AppendChild` que o Delphi executa sobre `Result=nil` (access violation em potencial) aqui são
no-ops.
