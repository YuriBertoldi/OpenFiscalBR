---
name: adicionar-registro-sped
description: "Acrescenta um registro novo ao package packages/sped (SPED Fiscal EFD-ICMS/IPI): struct do registro, campo no registro pai, writer da linha pipe-delimited, contadores, entrada no 9900 e teste. Cobre os 12 pontos de acoplamento do package, as pegadinhas de formatacao (parametro nulo, ordem de campos sem trava) e a fidelidade ao writer ACBr Delphi de origem. Use quando o usuario pedir para implementar, adicionar, portar ou completar um registro SPED (K200, 1010, H020, C110...), ou quando pedir para preencher os blocos incompletos."
argument-hint: "<codigo do registro>"
---

# adicionar-registro-sped

Implementa um registro do SPED Fiscal em `packages/sped`, do struct ao teste.

## Uso

```
/adicionar-registro-sped K200
/adicionar-registro-sped 1010
```

`$1` = código do registro (ex.: `K200`, `C110`, `1010`). Se vier vazio, pergunte qual registro.

**Contexto que justifica esta skill:** o package tem cerca de 198 structs de registro
declarados e apenas ~59 writers. São por volta de **139 registros que existem como struct e
nunca são escritos** — nos blocos G, H, K e 1 só saem as linhas de abertura (`X001`) e
encerramento (`X990`). Para esses, boa parte da checklist abaixo já está pronta: confira
antes de escrever, não duplique.

Há também registros que nem struct têm — `C130`, `C140`, `C160` e `C165` são filhos de
`TRegistroC100` no ACBr e não existem no `RegistroC100` do Go. Nesses, o fluxo é completo.

---

## Etapa 1 — Levantar o layout oficial

Fontes da verdade, nesta ordem de precedência:

1. **O writer Delphi de origem** — `ACBrEFDBloco_<X>_Class.pas`, procedure
   `TBloco_<X>.WriteRegistro<COD>`. É o que este projeto está portando; a sequência de
   chamadas `LFill`/`DFill` nele **é** a ordem dos campos.
2. **A classe do registro** — `ACBrEFDBloco_<X>.pas`, para os tipos e a hierarquia.
3. **O Guia Prático da EFD ICMS/IPI** — tamanho, decimais e obrigatoriedade.

Ao ler o writer Delphi, extraia também o que não é layout, porque é justamente o que se
perde num port apressado:

- todo `Check(...)` — validação obrigatória
- todo `case ... of` que monta string a partir de enum — vira o `String()` do enum
- **toda comparação com `DT_INI` / `EncodeDate(...)`** — regra de vigência; ver Etapa 9
- as flags `Nulo` de cada campo, inclusive quando são variáveis e não literais
- ramos `if COD_MOD = 'XX'` — variantes de layout por modelo de documento

Se o `.pas` de origem não estiver acessível, **pare e peça o caminho**. Inventar a ordem ou a
obrigatoriedade dos campos gera um arquivo que o validador da SEFAZ rejeita, e o erro não
aparece em build nem em teste — ver a Etapa 5.

## Etapa 2 — Conferir o que já existe

Para o registro `Xnnn`, verifique em `packages/sped`:

```bash
grep -rn "RegistroXnnn" packages/sped/
```

Classifique cada um dos pontos de acoplamento como **pronto** ou **a fazer**. Nos registros
dos blocos C, G, H, K e 1 que já existem sem writer, os pontos 1, 2, 7 e 8 tipicamente já
estão feitos.

**Struct existente não é struct correto.** Confira campo a campo contra o `.pas` antes de
escrever o writer em cima dele. Casos reais encontrados no package:

- `RegistroB030` era uma **cópia literal** do `RegistroB025` — nomes plausíveis, layout de
  outro registro.
- `Registro1010` tinha 5 campos inventados no lugar dos 13 do layout.
- `RegistroB500` e `RegistroB510` eram structs de outros registros.
- `RegistroB440` tinha `CodMunServ` onde o layout pede `COD_PART` — a contagem batia.

Confira também a **hierarquia**: o `RegistroB500` estava declarado como filho do `B470`
quando no ACBr é filho do `B001`.

## Etapa 3 — Os 12 pontos de acoplamento

| # | Arquivo | O quê |
|---|---|---|
| 1 | `bloco_<x>.go` | `type RegistroXnnn struct` + doc-comment `// RegistroXnnn - <descrição do Guia Prático>` |
| 2 | `bloco_<x>.go` | campo `RegistroXnnn []*RegistroXnnn` no struct do **pai** |
| 3 | `fiscal_types.go` | enums de domínio fechado novos (`type E int` + consts + `String()`) |
| 4 | `write_blocos.go` | `func (b *BlocoX) writeRegistroXnnn(parent *RegistroXppp)` |
| 5 | writer do pai | chamar `b.writeRegistroXnnn(r)` dentro do `for`, na posição do layout |
| 6 | writer novo | `b.Add(linha, true)` + `b.RegistroX990.QtdLinX++` |
| 7 | `bloco_writers.go` | declarar `RegistroXnnnCount int` no struct `BlocoX` |
| 8 | `bloco_writers.go` | zerar `b.RegistroXnnnCount = 0` em `func (b *BlocoX) LimpaRegistros()` |
| 9 | writer novo | `b.RegistroXnnnCount++` após o `Add` |
| 10 | `fiscal.go` | entrada em `populateRegistro9900()` |
| 11 | `sped_test.go` | teste de formato da linha |
| 12 | `packages/sped/README.md` | atualizar a lista de registros implementados |

**Não precisa mexer** em `WriteBlocoX`, `SaveFileTXT`, `allBlocoSPEDs`, `inicializaBloco` nem
`IniciaGeracao` — esses só mudam quando se acrescenta um **bloco** inteiro, não um registro.

## Etapa 4 — Escrever o struct (pontos 1 a 3)

Structs ficam em `bloco_<letra>.go`, um arquivo por bloco. São structs de dados puros: sem
tags, sem métodos, sem interface.

Mapa de tipos, seguido sem exceção no package:

| Campo do layout | Tipo Go |
|---|---|
| Valor monetário, alíquota, quantidade | `float64` |
| Data | `time.Time` |
| Numérico inteiro contável (`COD_MUN`, `CRO`, `CRZ`) | `int` |
| Código alfanumérico ou texto | `string` |
| Código de domínio fechado | enum próprio em `fiscal_types.go` |

**`CFOP`, `NUM_DOC` e `NUM_ITEM` são `string`, não `int`** — precisam preservar zeros à
esquerda. O mesmo vale para qualquer código numérico que não seja usado em aritmética.

Registro filho entra no pai como `[]*RegistroXnnn`, com o campo nomeado igual ao tipo.
Cardinalidade 1:1 usa ponteiro simples (`Registro0100 *Registro0100`) em vez de slice.

Enum novo segue o padrão de `fiscal_types.go`: `type E int` + consts + `String()` devolvendo o
**código literal da legislação**. Se o campo for opcional no layout, inclua um membro
`...Nenhum` cujo `String()` devolve `""`.

## Etapa 5 — Escrever o writer (pontos 4 a 6, 9)

O receiver é o **Bloco**, não o registro. Writers de dados são **não exportados**
(`writeRegistroC100`); só `X001` e `X990` são exportados. Filho recebe o pai por parâmetro.

Molde real, de `write_blocos.go`:

```go
// writeRegistroC170 gera as linhas do registro C170.
// Formato: |C170|NUM_ITEM|COD_ITEM|DESCR_COMPL|QTD|UNID|VL_ITEM|VL_DESC|...|
func (b *BlocoC) writeRegistroC170(parent *RegistroC100) {
	for _, r := range parent.RegistroC170 {
		linha := b.LFillStr("C170", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, true) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC170Count++
	}
}
```

### As quatro funções de formatação

Use apenas estas, de `packages/comum/txt_class.go`:

| Função | Assinatura | Para quê |
|---|---|---|
| `LFillStr` | `(value string, size int, nulo bool, char byte) string` | literal do REG, todo `string`, e **todo enum via `.String()`** |
| `LFillInt` | `(value int64, size int, nulo bool, char byte) string` | campos `int` — exige cast `int64(...)` |
| `LFillDate` | `(value time.Time, mask string, nulo bool) string` | sempre com a máscara `"02012006"` (ddmmaaaa no layout de referência do Go) |
| `DFill` | `(value float64, decimal int, nulo bool) string` | todo valor monetário, alíquota e quantidade |

`LFillFloat`, `RFill` e `Check` existem em `comum` mas têm **zero uso** no `sped`. Não
introduza — divergir do padrão aqui espalha dois jeitos de formatar o mesmo tipo de campo.

Cada função devolve `"|" + valor`. O `|` final da linha vem do `Add(linha, true)`.

### Campo opcional: o parâmetro `nulo`

`nulo = true` significa "se vazio ou zero, emite só o delimitador":

```go
b.DFill(r.VlItem, 2, false)   // obrigatório: zero vira |0,00
b.DFill(r.VlDesc, 2, true)    // opcional:    zero vira | (vazio)
```

**O `nulo` só esvazia valor ZERO — não apaga valor preenchido.** `DFill(1000, 2, true)` emite
`1000,00`, não vazio. É o mesmo comportamento do `LFill` do ACBr. Isso muda como se lê o
tratamento de documento cancelado: o efeito de `booNFCancelada` é duplo — datas e indicadores
são zerados **na origem** (e por isso saem vazios sempre), enquanto os campos monetários
apenas passam a aceitar vazio quando valem zero, onde num documento normal sairiam `0,00`.

**Pegadinha:** em `LFillStr` e `LFillInt` com `size == 0`, o `nulo` é irrelevante — o loop de
padding não roda e string vazia devolve `"|"` de qualquer forma. É por isso que o package
inteiro chama `LFillStr(..., 0, false, '0')`. O `nulo` só tem efeito quando há `size > 0`,
como em `LFillStr(r.CEP, 8, false, '0')`.

Para enum opcional, o mecanismo não é o `nulo` e sim o membro `...Nenhum` com `String()` vazio.

### A ordem dos campos não tem trava nenhuma

A ordem da linha é **exclusivamente** a sequência das concatenações `+`. Trocar duas linhas de
lugar não quebra compilação, não quebra `go vet` e não quebra teste — gera silenciosamente um
arquivo que a SEFAZ rejeita.

Os dois mecanismos que protegem contra isso são obrigatórios:

1. O doc-comment `// Formato: |REG|CAMPO1|...|` acima do writer, conferido 1:1 contra as
   concatenações. É o padrão já seguido em todo o `write_bloco_0.go`; em `write_blocos.go` só
   os writers exportados têm, e os de dados não — ao acrescentar um registro, escreva o
   comentário mesmo que os vizinhos não tenham.
2. O teste da Etapa 7 asseverando a **linha inteira literal**.

## Etapa 6 — Contadores e o 9900 (pontos 7, 8, 10)

São dois contadores, com destinos diferentes:

- **`QtdLinX++`** no registro `X990` do bloco — alimenta o total do registro 9999.
- **`RegistroXnnnCount++`** — alimenta o registro 9900.

Ambos são obrigatórios em todo writer.

O registro **9900 é populado à parte**, em `populateRegistro9900()` de `fiscal.go`. Duas
coisas a saber sobre o estado atual dessa função:

1. Ela só cobre o Bloco 0. Os blocos B a 1 têm apenas os hardcodes de abertura e encerramento
   (`f.addRegistro9900("C001", 1)` / `("C990", 1)`), então C100, C170, C190, B020, D100 e
   companhia **não aparecem no 9900**.
2. Onde ela conta, usa `len(slice)` — e não os contadores. Os `RegistroXnnnCount` estão
   declarados e sendo incrementados, mas hoje ninguém os lê.

**Use o contador, não `len(slice)`.** É o mecanismo do ACBr original
(`ACBrSpedFiscal.pas:832-841`), e é o único que funciona para netos: o registro 0175 pertence
a cada 0150, então não existe um slice único cujo `len()` dê a contagem correta — o ACBr usa
`Registro0175Count` justamente por isso (`ACBrSpedFiscal.pas:864-869`). O `len()` do Bloco 0
é um desvio do port Go, não o padrão a seguir.

```go
	if f.BlocoC.RegistroC170Count > 0 {
		f.addRegistro9900("C170", f.BlocoC.RegistroC170Count)
	}
```

A ordem já favorece isso: `populateRegistro9900()` roda depois dos writers, então os
contadores estão preenchidos quando ela executa.

Por isso o reset em `LimpaRegistros()` (ponto 8) não é burocracia: sem ele, o 9900 acumula
entre duas gerações no mesmo processo.

## Etapa 7 — Teste (ponto 11)

Testes ficam em `packages/sped/sped_test.go`, package interno `sped`, `testing` puro (o
projeto não usa testify). O teste chama o writer direto e lê `Bloco.Conteudo` em memória —
não grava arquivo.

Prefira asseverar a **linha inteira literal**, que é o que trava a ordem dos campos:

```go
func TestWriteRegistroC170_OutputFormat(t *testing.T) {
	f := NewSPEDFiscal()
	// ... montar o pai e o registro ...

	f.BlocoC.writeRegistroC100()

	want := "|C170|1|ITEM01|10,000|UN|100,00||"
	if f.BlocoC.Conteudo[1] != want {
		t.Errorf("linha C170:\ngot:  %q\nwant: %q", f.BlocoC.Conteudo[1], want)
	}
}
```

Quando for mais legível conferir campo a campo, use o padrão por índice já presente no
arquivo. Lembre que `strings.Split(linha, "|")` devolve `fields[0]` vazio, então **`fields[1]`
é o REG** e o N-ésimo campo do layout é `fields[N+1]`:

```go
	fields := strings.Split(line, "|")
	if fields[1] != "C170" {
		t.Errorf("REG = %q", fields[1])
	}
	if fields[4] != "10,000" {
		t.Errorf("QTD = %q, want 10,000", fields[4])
	}
```

Use `t.Fatalf` para pré-condição (número de linhas) e `t.Errorf` para cada campo, para não
abortar na primeira falha. Nas mensagens, nomeie o campo como no Guia Prático (`QTD`,
`VL_ITEM`), não como no Go.

Cubra também os casos de borda que a seção "Testes Obrigatórios" do `CLAUDE.md` pede: valor
zero com `nulo` true e false, campo vazio, data em ano bissexto.

## Etapa 8 — Fechar

1. Atualize a lista de registros implementados em `packages/sped/README.md` (ponto 12).
2. Rode `/validar-package sped`.
3. Se o README da raiz marcar `packages/sped` como "Completo" e ainda houver registros sem
   writer, rode `/sincronizar-meta` para acertar o status.

## Etapa 8b — Auditar o que você escreveu

```bash
py ferramentas/comparar-campos.py <bloco> <registro>
```

Primeira passada apenas. A ferramenta compara **quantidade** de campos, não nomes — leia
`ferramentas/README.md` para os limites antes de confiar num `OK`. A conferência nome a nome
contra o `.pas` continua sendo manual.

Registros marcados com `[condicional: ...]` têm `ifthen`/`EncodeDate`/`DT_INI` no writer
Delphi e **sempre** exigem leitura manual: é onde as regras se perdem.

## Etapa 9 — Regras de vigência

Vários registros e blocos só existem a partir de uma data. No ACBr isso é uma guarda em volta
da escrita, e o bloco inteiro é **omitido** fora da vigência —
`ACBrSpedFiscal.pas:630-647`:

```pascal
   if DT_INI >= EncodeDate(2019,01,01) then
   begin
     WriteRegistroB001;
     WriteRegistroB990;
     Bloco_B.WriteBuffer;
   end;
```

Há o mesmo padrão para o bloco G (a partir de 2011-01-01) e K (2016-01-01).

Enums também têm vigência: o `IND_FRT` do C100 tem três tabelas de código distintas conforme
`DT_INI` (antes de 2012, antes de 2018, e de 2018 em diante), e o `IND_PGTO` perde o valor
`"9"` a partir de 2012-07-01. Repare que os códigos foram **remanejados**, não acrescentados:
`"0"` significa "por conta de terceiros" até 2011 e "por conta do emitente" depois. Um
`String()` estático não dá conta — quando o `.pas` mostrar o `case` dependendo de `DT_INI`,
implemente `StringEm(dtIni time.Time) string` e chame essa forma no writer.

### Vigência por data e por versão de leiaute são dimensões diferentes

Nem toda condicional olha `DT_INI`. Boa parte olha `COD_VER` do registro 0000. Os dois
mecanismos convivem, e confundi-los gera campo no período errado:

| Condicional | Governada por | Campo |
|---|---|---|
| `DtIni >= 2017-01-01` | data | `CEST` no 0200 |
| `DtIni >= 2018-01-01` | data | `COD_MUN_ORIG`/`COD_MUN_DEST` no D100 |
| `DtIni >= 2015-01-01` | data | `VL_ITEM_IR` no H010 |
| `CodVer > VlVersao114` | versão | `COD_BARRA` no 0220 |
| `CodVer > VlVersao111` | versão | `VL_ABAT_NT` no C170 |
| `CodVer >= VlVersao110` | versão | `CHV_DOCe` no C113 |
| `CodVer >= VlVersao112` / `>= 113` | versão | `IND_GIAF*` / `IND_REST_RESSARC` no 1010 |
| `CodVer == VlVersao101` | versão | E116 **não é escrito** |

Padrão do package: campo condicional sai para uma função auxiliar
`func nomeCampoNNN(b *BlocoX, r *RegistroNNN) string` que devolve `""` quando não se aplica —
**string vazia, não delimitador vazio**: o campo some da linha inteiramente, como no `ifthen`
do ACBr. O comparador de campos sabe seguir essas funções.

Quando a condicional afeta mais de um campo ou a linha toda, use `if` dentro do writer com
`linha += ...` (é o caso de `MES_REF` no E116 e dos `IND_GIAF*` no 1010).

Se o registro que você está acrescentando tiver vigência, escreva um teste por faixa —
de `DtIni` ou de `CodVer`, conforme o caso. Uma regra de vigência perdida produz arquivo
aceito pelo build e rejeitado pelo PVA.

---

## Regras

- **Nunca invente a ordem, o tamanho ou a obrigatoriedade de um campo.** Não há validação que
  pegue isso; o erro só aparece na recepção da SEFAZ, muito depois. Sem o layout em mãos,
  pare e peça.
- **Sempre escreva o doc-comment `// Formato:`**, mesmo que os writers vizinhos do mesmo
  arquivo não tenham. É o único contrato existente sobre a ordem dos campos.
- **Sempre inclua os dois contadores.** `QtdLinX++` alimenta o 9999 e `RegistroXnnnCount++`
  alimenta o 9900. Esquecer qualquer um não quebra nada visivelmente — o arquivo sai com
  contagem errada e é rejeitado na recepção.
- **Não descarte `Check` nem regra de vigência do Delphi** por parecerem acessórios ao
  layout. São a parte do writer original que um port mecânico perde, e a ausência delas não
  aparece em nenhum teste de formato.
- **Um registro por vez.** Ao completar um bloco inteiro, faça e valide um registro antes de
  passar ao próximo; uma leva de writers sem teste intermediário é impossível de depurar
  quando a contagem sai errada.
- **Quando o ACBr estiver errado, replique mesmo assim — se for determinístico.** O
  `WriteRegistroB350` do ACBr emite a linha com o literal `B035`. Está portado assim, com teste
  e comentário dizendo que é intencional. Divergir em silêncio do original é pior do que
  carregar o defeito dele, porque quebra a compatibilidade sem ninguém perceber.
  A exceção é comportamento **não determinístico** — no ACBr, as variáveis `strIND_FRT` e
  `booAborta` são declaradas fora do laço e conservam o valor do documento anterior quando o
  `case` não cobre o membro. Replicar isso faria o mesmo dado gerar arquivos diferentes conforme
  a ordem dos documentos. Nesses casos, implemente o correto, registre na seção
  "Divergências deliberadas" do `packages/sped/README.md` e **avise o usuário** — é decisão dele.

## Armadilhas conhecidas

| Sintoma | Causa |
|---|---|
| 9999 com contagem errada | faltou `QtdLinX++` |
| registro ausente do 9900 | faltou a entrada em `populateRegistro9900` |
| 9900 inflando a cada geração no mesmo processo | faltou `Count = 0` em `LimpaRegistros` |
| neto contado a menos no 9900 | uso de `len(slice)` em vez do contador |
| campo com `0,00` onde devia sair vazio | `nulo=false` onde o `.pas` usa `true` |
| CFOP ou NUM_DOC perdendo zero à esquerda | campo declarado `int` em vez de `string` |
| campos trocados de posição, tudo verde | a ordem só existe nas concatenações — faltou o teste de linha inteira |
| bloco emitido em período que não deveria | guarda de vigência por `DtIni` ausente |
