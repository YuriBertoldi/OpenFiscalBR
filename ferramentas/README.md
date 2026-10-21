# Ferramentas de apoio ao port

Scripts de desenvolvimento. Não fazem parte do módulo Go e não são compilados.

## Configuração

Os dois scripts leem os fontes Delphi do ACBr, que ficam fora deste repositório.
Informe onde eles estão pela variável `ACBR_FONTES` — sem ela o script avisa e
sai, em vez de falhar com "arquivo não encontrado" lá adiante:

```bash
# Windows (cmd)
set ACBR_FONTES=C:\caminho\ACBr\Fontes\ACBrTXT\ACBrSPED\ACBrSPEDFiscal

# Windows (PowerShell)
$env:ACBR_FONTES = "C:\caminho\ACBr\Fontes\ACBrTXT\ACBrSPED\ACBrSPEDFiscal"

# bash
export ACBR_FONTES=/caminho/ACBr/Fontes/ACBrTXT/ACBrSPED/ACBrSPEDFiscal
```

`GO_SPED` aponta para o `packages/sped` e tem default relativo ao próprio
repositório — só precisa ser definida para comparar com outra cópia.

## `comparar-campos.py`

Compara a sequência de campos de um writer do ACBr Delphi com a do writer Go
correspondente, para auditar fidelidade do port.

```bash
py ferramentas/comparar-campos.py <bloco> <registro> [<registro>...]

py ferramentas/comparar-campos.py 0 0000 0150 0200
py ferramentas/comparar-campos.py C C100 C170 C190
py ferramentas/comparar-campos.py B B020 B350 B470
```

Saída por registro:

- `OK` — mesma quantidade de campos **e** mesmos nomes, posição a posição
- `NOME` — quantidade bate, mas algum campo tem nome divergente (imprime quais)
- `DIF` — quantidades divergentes, com as duas listas lado a lado
- `--` — existe no ACBr e não tem writer no Go
- `??` — não encontrado no ACBr

### Modo `--doc` — tabela de nomes

Com `--doc` o script troca o diagnóstico por uma tabela markdown por registro,
alinhando os nomes dos dois lados:

```bash
py ferramentas/comparar-campos.py D D100 --doc
```

```
| # | ACBr | Go |
|---|------|-----|
| 0 | D100 | D100 |
| 1 | IND_OPER | IndOper |
```

É o que gera `packages/sped/MAPEAMENTO-CAMPOS.md`.

O marcador `[condicional: ...]` sinaliza que o writer Delphi tem `ifthen`,
`EncodeDate` ou teste de `DT_INI` na montagem da linha. **Todo registro marcado
assim precisa de conferência manual** — é exatamente onde os campos condicionais
foram perdidos no port original.

O script segue chamadas a funções auxiliares do padrão `nomeHelper(b, r)`, então
campos condicionais extraídos para função própria continuam sendo contados.

### Limites — leia antes de confiar no resultado

1. **Lê apenas um ramo quando o ACBr tem dois `Add`.** Registros como `0220`,
   `E116`, `G110` e `1010` têm blocos alternativos por versão de leiaute; o
   extrator pega o primeiro. O lado Go, que fatora o trecho comum e ramifica só
   a diferença, aparece então com campo a mais — `DIF` que é falso positivo.
2. **Não captura `LFILL` em caixa alta**, que o ACBr usa em pontos isolados (o
   campo `QTD` do `C170`, por exemplo).
3. **Compara nome, não semântica.** Nome igual não garante que o valor escrito
   seja o certo — a condicional em volta do campo continua sendo conferência
   manual contra o `.pas`, guiada pelo marcador `[condicional: ...]`.
4. **Nome divergente nem sempre é defeito.** Quando o ACBr monta o valor numa
   variável local com nome próprio (`ChaveEletronicaCTe` no `D100`), os dois
   lados divergem legitimamente. Esses casos estão em `ALIAS_GO`, no topo do
   script — acrescente ali em vez de abreviar o nome do campo Go.

Contrapartida do item 4: **helper do lado Go deve repetir o nome do campo nos
parâmetros**. Abreviar (`icmsOp` em vez de `vlUnitICMSOpConv`) cega a auditoria
nos quatro registros que compartilham o helper.

### O que a auditoria de nomes já pegou

A comparação de nomes foi acrescentada depois da auditoria de quantidade e,
mesmo com 100% dos registros passando em quantidade, achou:

- `D100` — a chave do CT-e não era zerada no documento inutilizado, o `nulo` dos
  seis monetários estava arbitrário e o `IND_FRT` ignorava o remapeamento por
  vigência (que no bloco D tem corte diferente do bloco C)
- `G110` — a linha saía nas versões 100/101, quando o ACBr não emite nenhuma
- quatro campos com nome abreviado, renomeados para bater com o ACBr

O `go vet` somado a isso também revelou `E112`, `E113` e `E115`: writers prontos
que ninguém chamava.

### Caminho dos fontes

O caminho do ACBr está fixo no topo do script (`ACBR`). Ajuste se os fontes
Delphi estiverem em outro lugar na sua máquina.

## `conferir-leitor-dfe.py`

Confere um leitor de DFe portado contra o `XmlReader` Delphi de origem, tag a
tag, nos dois eixos que o build, o `go vet` e o teste de leitura **não**
alcançam:

1. a **precisão** de cada campo (`tcDe2` × `ConteudoDe2`, `tcDe10` ×
   `ConteudoDe10`, ...);
2. a escolha entre **`Find`** (nome qualificado) e **`FindAnyNs`** (local name).

```bash
py ferramentas/conferir-leitor-dfe.py <arquivo.pas> <arquivo.go>

py ferramentas/conferir-leitor-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.XmlReader.pas" \
   packages/nfe/xml_reader.go
```

Sai com código 1 se houver divergência, para encaixar em CI.

### Por que um script, e não um teste Go

Porque **nenhum teste de leitura consegue distinguir `ConteudoDe2` de
`ConteudoDe10`**. O parâmetro de casas do `pcn.ConteudoDec` só tem efeito
quando o node está marcado com `FloatIsIntString`, flag que o parser nunca
liga — no XML da SEFAZ o valor já vem com ponto decimal. Os dois devolvem o
mesmo resultado.

A precisão é um **contrato declarativo**, herdado do mesmo mecanismo do ACBr,
e só se verifica comparando com o `.pas`. Um teste que afirme travá-la dá
falsa confiança, que é pior do que não ter teste.

### Limites — leia antes de confiar no resultado

1. **Casa por NOME DE TAG, não por caminho.** A mesma tag em grupos diferentes
   vira uma entrada só: se o `.pas` lê `vItem` como `tcDe10` em `prod` e
   `tcDe8` em `gProcRef`, o script aceita qualquer um dos dois no Go. Tag com
   precisão divergente entre grupos precisa de conferência manual — o relatório
   mostra o conjunto de tipos encontrados de cada lado justamente para isso.
2. **Só enxerga `ObterConteudo(...Find('tag'), tc...)` literal.** Campo lido por
   variável, por helper próprio ou com o nó guardado antes some da conta. Na
   NFe isso acontece com `tpViaTransp` e `tpIntermedio`, que o ACBr lê com
   `tcInt` e converte depois — eles aparecem na lista de "sem contraparte", que
   **precisa ser lida item a item**, não ignorada.
3. **Não confere ordem, condicional nem guard.** Condicional por versão de
   leiaute, `if` em volta da leitura e default aplicado na ausência do nó
   continuam sendo conferência manual contra o `.pas`.
4. **Linha comentada no Delphi é descartada** (`//`), para não contar campo
   desativado — `RENAVAM` do `veicProd` é o caso na NFe.

## `conferir-gerador-dfe.py`

O par do anterior, para o lado da **escrita**. Confere um gerador de DFe portado
contra o `XmlWriter` Delphi de origem, tag a tag, nos dois eixos que o build, o
`go vet` e o **round-trip** não alcançam:

1. a **precisão** de cada campo (`tcDe2` × `NodeDec(..., 2, ...)`,
   `FormatoValor4ou2` × `nodeAliquota`, ...);
2. a **obrigatoriedade** — o `nOcorrencias` do `AddNode`, que vale `1` (sai mesmo
   zerado) ou `0` (some quando vazio);
3. no modo `--ordem`, a **sequência** das tags de um grupo — o eixo mais caro de
   errar, porque o XSD é `xs:sequence` e campo fora de ordem é documento
   rejeitado pela SEFAZ.

```bash
py ferramentas/conferir-gerador-dfe.py <arquivo.pas> <arquivo.go> [--linhas INICIO:FIM]

py ferramentas/conferir-gerador-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.XmlWriter.pas" \
   packages/nfe/xml_writer_det.go --linhas 1118:3507
```

O recorte por linha existe porque o writer Delphi é um arquivo só, de milhares
de linhas, enquanto o porte Go está dividido em arquivos. Sem ele a comparação
mistura grupos e produz ruído. Para auditar o gerador inteiro, concatene os
arquivos Go num temporário e rode sobre a faixa toda.

Sai com código 1 se houver divergência, para encaixar em CI.

### Modo `--ordem`

```bash
py ferramentas/conferir-gerador-dfe.py <pas> <go> --ordem INICIO:FIM <funcaoGo> [prefixoDoCase]

py ferramentas/conferir-gerador-dfe.py "$PAS" packages/nfe/xml_writer_det.go \
   --ordem 2097:2199 gerarICMSRegimeNormal pcn.CST10
```

Imprime as duas sequências lado a lado e sai com código 1 **só** quando a ordem
relativa das tags comuns diverge — que é o que o XSD cobra. É **interativo de
propósito**: o mapeamento entre faixa do `.pas` e função Go muda a cada etapa do
porte e envelheceria guardado num arquivo.

Tag presente em um lado só **não** reprova; vai para uma lista à parte. É
esperada em três casos, e o relatório os separa justamente para serem lidos:

- sub-writer que o porte **inlinou** aparece só do lado Go — confira a POSIÇÃO em
  que ele entra (foi assim com `NVE` e `gCred` dentro do `prod`);
- sub-writer que o porte manteve em função própria **some dos dois lados**;
- campo que o Delphi escreve com dois `AddNode` literais, um por ramo do `if`,
  aparece **duplicado só do lado `.pas`** (`vAFRMM`, `pGLP`, `cListServ`).

### Por que um script, e não um teste Go

Pelo mesmo motivo do leitor, do outro lado do ciclo: **o round-trip
(gerar→ler→gerar) é cego para os dois eixos**. A precisão some porque o leitor
converte `"0.00"` e `"0.0000"` no mesmo `float64`; a obrigatoriedade some porque
um campo zerado, lido de volta, continua zerado — as duas gerações concordam em
qualquer uma das duas escolhas.

### Limites — leia antes de confiar no resultado

1. **Casa por NOME DE TAG, não por caminho**, igual ao do leitor. `vBC` aparece
   em dez grupos e vira uma entrada só; o relatório mostra o conjunto de tipos
   de cada lado justamente para isso ser conferido à mão quando divergir.
2. **A ordem só é conferida quando você pede, grupo a grupo.** O modo `--ordem`
   exige a faixa do `.pas` e o nome da função Go porque o porte reagrupa os
   `case` do Delphi em funções próprias — descobrir esse mapeamento sozinho daria
   mais falso positivo do que achado. Como guard permanente ficam os testes Go de
   sequência (`xml_writer_ordem_test.go`, `xml_writer_det_test.go`), e é por onde
   já passaram dois defeitos reais (`guiaTransito` e `pBCOp`/`UFST`).
3. **Campo decidido por EXPRESSÃO no Go vira pendência, não falha.** Onde o
   Delphi usa duas chamadas literais (`AddNode(..., 1, ...)` e
   `AddNode(..., 0, ...)`) e o porte usa uma chamada condicional, o script não
   consegue decidir e lista o campo em "decididos por expressão". **Essa lista
   precisa ser lida item a item.** Reprovar ali deixaria a ferramenta
   permanentemente vermelha, e ferramenta vermelha para de ser lida.
4. **Helper do porte precisa ser declarado.** Construtor de node próprio
   (`nodeAliquota`, `nodeHora`, `nodeNumStr`) entra em `CONSTRUTORES_GO`; helper
   que esconde a decisão de precisão dentro dele (`nodePercentualGLP`) entra em
   `HELPERS_GO`, com o conjunto de pares copiado do `.pas`. Helper não declarado
   some da conta e a tag cai em "sem construtor de node no Go" — lista que
   também **precisa ser lida**, não ignorada.
5. **Comentário do Delphi é descartado** (`//`, `{...}`, `(*...*)`), para não
   contar campo desativado. Na NFe isso acontece com `indISSRet`, `RENAVAM` e as
   duas tags da NT 2016.002 que o ACBr comentou e reescreveu logo abaixo.
