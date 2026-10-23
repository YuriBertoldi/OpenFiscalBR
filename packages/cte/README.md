# packages/cte

Porte do componente **ACBrCTe** — Conhecimento de Transporte Eletrônico.

| | |
|---|---|
| Componente Delphi | `ACBrDFe/ACBrCTe/` |
| Modelos | 57 (CT-e), 64 (GTV-e), 67 (CT-e OS), 57 (CT-e Simplificado) |
| Leiaute | 2.00, 3.00, 4.00 |
| Estado | etapa 2a de 11 — classes de dados |

## Quatro documentos, um componente

A diferença estrutural em relação ao `packages/nfe` é que o CT-e não é um documento com dois
modelos: são **quatro leiautes distintos** atendidos pela mesma árvore de classes.

O **CT-e Simplificado compartilha o código 57** com o CT-e normal. O comentário do próprio
ACBr (`ACBrCTe.Conversao.pas:106`) explica: *"O código correto do modelo CTe Simplificado é 57,
mas foi alterado para 58"* — internamente há um 58 para distingui-los, e o que sai no documento
é 57.

A consequência aparece na conversão: `ModeloCTeToStr` tem quatro entradas e `StrToModeloCTe`
tem **três**. Ler `"57"` devolve sempre `MoCTe`, nunca `MoCTeSimp` — e não há como fazer a ida
e volta fechar, porque a informação que separa os dois não está no código. Quem precisa
distinguir olha o `tpCTe`, onde o Simplificado tem valor próprio (`TcCTeSimp`).

## Seis modais

`rodoviário`, `aéreo`, `aquaviário`, `ferroviário`, `dutoviário` e `multimodal` — árvores de
classe separadas e mutuamente exclusivas. É o análogo do grupo `det` da NF-e em peso, mas com
seis variantes em vez de uma. Entram nas etapas 2b, 4b e 9.

Na etapa 2a eles estão **declarados vazios** em `classes_modais.go`, junto com o `RodoOS` (a
variante do rodoviário usada no CT-e OS). A alternativa seria omitir os campos `rodo`, `aereo`
e companhia do `InfCTeNorm` até a 2b — mas isso faria a raiz portada divergir do `TCTe` por uma
etapa inteira, e struct raiz incompleto é o tipo de coisa que não quebra build nem teste e
produz documento sem modal nenhum.

## Reaproveitamento: o que *não* está neste package

| O que | Onde vive | Por quê |
|---|---|---|
| Os **14 grupos da Reforma Tributária** (`gIBSCBS`, `IBSCBS`, `gCompraGovReduzido`, `gRed`, `gDif`, `gEstornoCred`, …) | `packages/rtc` | o ACBr os redeclara no topo do `ACBrCTe.Classes.pas`, mas são os mesmos tipos do `ACBrDFe.RTC`. Os 14 foram conferidos um a um antes de decidir. **O `rtc.IBSCBS` é superconjunto**: traz cinco grupos que o CT-e não aceita, anotados no campo `Imp.IBSCBS` |
| `Modal`, `UnidMed`, `UnidTransp`, `UnidCarga` | `packages/pcn` | vêm de `ACBrDFe.Conversao.pas`, a unit **compartilhada**, e o próprio ACBr marca dois deles com `// CTe e MDFe`. Declará-los aqui obrigaria o MDF-e a redeclarar |
| `TipoAmbiente`, `TipoEmissao`, `TipoImpressao`, `ProcessoEmissao`, `IndIEDest`, `CSTIcms`, `Indicador`, `ProcDFe`, `Signature` | `packages/pcn` | idem — base comum a todos os DFe |

## Quatro aliases que não são tipo

Só se acham lendo o `.pas`; portá-los como tipo próprio cria structs que o original não tem:

| No ACBr | Linha | No porte |
|---|---|---|
| `TObsFiscoCollection = TObsContCollection` | `:621` | `Compl.ObsFisco` é `[]ObsCont` |
| `TlacUnidTranspCollection` e `TlacUnidCargaCollection` = `TLacreCollection` | `:1156`, `:1158` | os dois são `[]Lacre` |
| `TinfUnidCargaNFCollection = TinfUnidCargaCollection` | `:1214` | mesmo `[]InfUnidCarga` de sempre |
| `TLacRodoCollection = TLacreCollection` | `:1630` | `[]Lacre` — **da etapa 2b** |

O primeiro foi achado **pela conferência mecânica**, depois de já estar portado errado. Os quatro
estão travados por `TestClasses_AliasesDoACBr` — um alias não aparece em lista de classe nenhuma,
nem a conferência o alcança, então o que segura a decisão é a atribuição deixar de compilar.

## Armadilhas das classes

| Onde | O que surpreende |
|---|---|
| `Peri` | tem **dois donos**, decididos pela versão: `infCTeNorm.peri` só na 2.00 (`XmlWriter.pas:1792`), `aereo.peri` da 3.00 em diante (`:3727`). O tipo mora no `classes.go` porque os dois precisam dele — e a **2b tem de declarar `Peri []Peri` dentro do `Aereo`** |
| `Peri` | `qTotProd`, `qVolTipo` e `qTotEmb` são **string** porque o ACBr assim declara e escreve com `tcStr`. Na 2.00 o schema aceita texto livre; da 3.00 em diante o `qTotProd` é `TDec_1104` dentro do `infTotAP` e precisa chegar formatado |
| `CTe.infCteComp` | duas properties que saem com a **mesma tag** `infCteComp`. O singular vale até a **3.00 inclusive**, a coleção da **4.00** (`if VersaoDF <= ve300`, `XmlWriter.pas:493`). Dentro do singular a tag interna muda: `chave` na 2.00, `chCTe` na 3.00 |
| `CTe.infCteAnu` | sai no **mesmo `if`** do `infCteComp`: na 4.00 o ACBr não o emite |
| `LocColeta`, `LocEnt`, `ICMSUFFim.pICMSInterPart` | **não saem em XML nenhum** — zero ocorrências no `XmlWriter`/`XmlReader` e ausência de qualquer XSD. Vivem só no `.ini`. Mantidos por fidelidade; a etapa do `.ini` vai precisar deles |
| `Ide`, `InfNF`, `RefNF` | a property chama `modelo` e a tag é `mod` |
| `Endereco` × `EnderEmit` | o segundo **não tem** `cPais`/`xPais` (emitente é sempre nacional) e **tem** `fone` |
| `LocColeta`, `LocEnt` | além disso, são endereço reduzido: sem CEP, sem país, sem fone |
| `Emit` | só aceita **CNPJ**, não `CNPJCPF` — transportador é sempre pessoa jurídica |
| `ICMS` | **sete** grupos mutuamente exclusivos em tipos separados, não um struct achatado como o `det` da NF-e. `SituTrib` escolhe |
| `CST45` | o grupo chama 45, mas carrega CST **40, 41 ou 51** — o leiaute juntou três situações |
| `ICMSSN.indSN` | é **inteiro**, não enum |
| `Entrega` | seis subgrupos mais `TipoData`/`TipoHora`, que **não são tags** — são o seletor que o gerador consulta |
| `InfModal` (Simplificado) | tem **três** modais, não seis: só rodoviário, aéreo e aquaviário (`:2661`) |
| `IDDocAntEle` | dois campos de chave, `chave` e `chCTe`, que o ACBr mantém separados — o leiaute mudou de nome entre versões |
| `TCTe` | está em `:2724` com **um** espaço de indentação, não dois. Foi o que fez a raiz escapar da primeira varredura |

## Origem dos enums: `PCNCTe/pcteConversaoCTe.pas`

O ACBrCTe declara os **mesmos 37 enums em duas units**, e elas divergem:

- `Base/ACBrCTe.Conversao.pas`
- `PCNCTe/pcteConversaoCTe.pas`

Vale a segunda, por duas evidências independentes:

1. **Resolução do `uses`.** As seis units vivas — `XmlWriter`, `XmlReader`, `IniWriter`,
   `IniReader`, `Classes` e `XmlHandler` — listam `pcteConversaoCTe`. A única menção a
   `ACBrCTe.Conversao` está em `XmlWriter.pas:46` e está **comentada**.
2. **A unit `Base/` perde membros.** O `TSchemaCTe` de lá não tem `schevVincPgto` nem
   `schevCancVincPgto` — o evento de Vínculo de Pagamento e o cancelamento dele, os dois mais
   recentes.

Travado por `TestSchemaCTe_MembrosExclusivosDaUnitPCN`.

## Armadilhas dos enums

As que o teste trava, porque nenhuma aparece em build, `vet` ou round-trip ingênuo:

| enum | o que surpreende |
|---|---|
| `TipoServico` | a tabela **pula o código `5`**: vai de `4` (multimodal) para `6` (transporte de pessoas). Preencher a lacuna desloca os quatro membros seguintes |
| `Retira` | `RtSim` é **zero** e `RtNao` é um. O campo responde *"o recebedor retira?"* com 0=Sim |
| `Lotacao` | o contrário — `LtNao` é zero. Os dois convivem no mesmo leiaute |
| `TipoDataPeriodo`, `TipoHorarioIntervalo` | o último código é `"-1"`, não vazio |
| `TipoFretamento`, `CRT`, `ClassDuto`, `TipoContratacao` | o **primeiro** membro tem código vazio |
| `CRT` | **não é o mesmo do `packages/nfe`**: lá o primeiro membro é `CrtSimplesNacional` com código `"1"`; aqui há um `CrtNenhum` vazio na frente e tudo desloca em um |
| `TipoDirecao`, `TipoPropriedade` | códigos em letra (`N`/`L`/`S`/`O`, `P`/`T`) — os dois únicos em que a comparação sem caixa do `StrToEnumerado` muda resultado |
| `ClassDuto`, `TipoContratacao` | o ACBr os lê com laço próprio e comparação **com** caixa, não pelo `StrToEnumerado` |
| `SchemaCTe` | o nome vem do **próprio identificador Pascal** (`GetEnumName` menos o prefixo `sch`), não de tabela. A `TSchemaCTeArrayStrings` tem 18 entradas vazias e é letra morta |

### `TpInfManu`: duas tabelas que não se falam

O caso mais grave, e é defeito do original:

| | códigos | membros cobertos |
|---|---|---|
| `String()` | `01`..`15`, `99` | 16 — inclui as Packing Instructions `ImPI965`..`ImPI970` |
| `StringV2()` | `1`..`9`, `99` | 10 — **os seis `ImPI*` não têm código** |

O ACBr usa as duas em lados opostos:

```
escrita   XmlWriter.pas:3755, IniWriter.pas:1202   -> TpInfManuToStrV2   ("1")
leitura   XmlReader.pas:203, XmlHandler.pas:3420,
          IniReader.pas:1551                       -> StrToTpInfManu    ("01")
```

Ele **grava `1` e procura `01`**: um documento gerado pelo ACBr e relido pelo ACBr perde o
campo, e seis membros não têm como ser emitidos.

As duas formas ficam expostas aqui porque as duas existem no mundo. **Qual delas o gerador
emite é decisão da etapa do gerador**, não desta; o que esta etapa garante é que nenhuma se
perca — e `ParseTpInfManuV2` recusa string vazia, que de outro modo casaria com `ImPI965`.

## Divergências deliberadas

| # | Onde | O ACBr faz | O porte faz |
|---|---|---|---|
| 1 | `StrToSchemaCTe` | devolve `TSchemaCTe(-1)` para nome desconhecido **e ainda marca `ok := True`** — um enum fora de faixa que se propaga em silêncio | devolve `SchErro` mais um `error`. Mesma decisão do `pcn.ParseSchemaDFe` na Distribuição DF-e da NF-e |
| 2 | `StrToXxx` em geral | `raise EACBrException` | devolve o membro zero + `error`, pela convenção de erros do projeto |

## Como rodar os testes

```bash
go test ./packages/cte/ -v
```

Sem dependência externa e sem rede.

### Conferência mecânica

```bash
py ferramentas/conferir-enums-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrCTe/PCNCTe/pcteConversaoCTe.pas" \
   packages/cte/conversao.go packages/cte/types.go \
   --par TpcteTipoCTe=tipoCTeCodigos --par TpInfManu=tpInfManuCodigos
```

**29 enums conferidos, zero divergência.** O único aviso que resta é o `TpInfManuV2` descrito
acima — tabela de código divergindo de tabela de código dentro do próprio ACBr.

**13 cortes deliberados**, todos pegos por algum teste.

As classes da etapa 2a:

```bash
py ferramentas/conferir-classes-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrCTe/Base/ACBrCTe.Classes.pas" \
   packages/cte/classes.go packages/cte/classes_modais.go \
   --par TinfDocAntCollectionItem=InfDocAntDet \
   --par TInfCteCompCollectionItem=InfCteCompItem
```

**98 classes casadas, zero divergência de campo**, e **nenhuma struct no Go sem classe no
Delphi** — nada inventado.

As 7 divergências que o relatório mostra são exatamente os modais declarados vazios, e as 33
classes "sem struct no Go" são 14 da Reforma Tributária (que vivem no `packages/rtc`) mais 19
subclasses dos modais. A dívida da 2b fica visível no próprio relatório em vez de escondida —
quando a 2b entrar, os dois números zeram.

**9 cortes deliberados**, todos acusados — campo removido, campo inventado, struct inteira
removida, `ObsFisco` de volta como tipo próprio, grupo pendurado na struct errada, escalar
trocado, **alvo de slice trocado** (`[]InfNFe` → `[]InfNF`), aridade trocada, e struct tag
acrescentada sem cegar o leitor.

> O script confere nome, tipo escalar, aridade e alvo de slice, mas **não confere ordem** nem
> alcança alias. Veja `ferramentas/README.md` para os limites.

## Roteiro

| # | Etapa | Estado |
|---|---|---|
| 1 | `types.go` + `conversao.go` | pronta |
| 2a | `classes.go` — cabeçalho, prestação, documentos | **nesta entrega** |
| 2b | `classes_modais.go` — os seis modais | pendente |
| 3 | `xml_reader.go` | pendente |
| 4a | `xml_writer.go` — cabeçalho e prestação | pendente |
| 4b | `xml_writer_modais.go` | pendente |
| 5 | regras de negócio e URLs | pendente |
| 6a | web services — status e consulta | pendente |
| 6b | web services — recepção, inutilização, cadastro | pendente |
| 7 | eventos | pendente |
| 8 | formato `.ini` — cabeçalho e prestação | pendente |
| 9 | formato `.ini` — modais | pendente |
| 10 | demo em `demos/cte` | pendente |
