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
5. **Nome de tag MONTADO no Delphi some da conta.** O leitor da Distribuição
   DF-e é o caso extremo: ele é genérico e monta as tags a partir do tipo de
   documento (`'ch' + FptpDFe`, `'inf' + FptpDFe`, `'prot' + FptpDFe`). O
   porte as escreve literalmente — `chNFe`, `infNFe`, `protNFe` — e elas
   aparecem como "só no .go".

### A lista de "sem contraparte" precisa ser lida, não ignorada

Rodando contra o `ACBrDFeComum.RetDistDFeInt.pas` o script dá **40 tags
conferidas, nenhuma divergência** — e 22 diferenças de cobertura, todas
explicáveis:

| lado | quantas | o quê |
|---|---|---|
| só no `.pas` | 16 | `CNPJ`/`CPF` (lidos por helper) e catorze do **CT-e**: grupo `CTe` do `detEvento`, `vPrest`/`vTPrest`, `itensAverbados` |
| só no `.go` | 6 | nomes que o Delphi **monta** a partir de `FptpDFe` |

Nenhuma delas é campo perdido — mas isso só se sabe depois de conferir uma a
uma. É exatamente para isso que o relatório lista os dois conjuntos.

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

## `conferir-regras-dfe.py`

Confere as regras de negócio portadas contra o `ValidarRegrasdeNegocio` do ACBr.

```bash
py ferramentas/conferir-regras-dfe.py <ValidarRegrasdeNegocio.pas> <arquivo.go>...

py ferramentas/conferir-regras-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.ValidarRegrasdeNegocio.pas" \
   packages/nfe/regras_negocio*.go
```

O risco que ele cobre não é **errar** uma regra — é **esquecer** uma. Regra que
some não quebra build, não quebra `vet`, não quebra teste e não aparece no
round-trip; só aparece quando a SEFAZ recusa o documento em produção.

Separa as regras **ativas** das **comentadas**. O ACBr comenta regra que a
SEFAZ revogou ou que ainda não entrou em vigor, e isso é informação: o porte
não deve implementá-las, mas quem lê precisa saber que existem. Na NFe são 14
(502, 508, 529, 571, 600, 629, 630, 719, 720, 806, 856, 858, 883, 888).

### Limites

1. **Reconhece o código por POSIÇÃO SINTÁTICA**, não por contexto: um literal
   de três dígitos cercado por `(`, `{` ou `,` de um lado e `,` ou `)` do
   outro. Cobre as quatro formas que o porte usa — chamada direta, chamada com
   `item`, tabela de confrontos e código passado como parâmetro — sem casar com
   números de conta, como o `100` de uma divisão por cem.
2. **Não confere a CONDIÇÃO da regra**, só a presença. Regra portada com o
   `and` trocado por `or` passa. Para isso são os testes Go e a revisão.
3. **A lista "só no Go" não reprova** — pode ser falso positivo da extração.
   Precisa ser lida.

## `conferir-urls-dfe.py`

Confere a tabela de URLs portada contra o `<Componente>Servicos.ini` do ACBr.

```bash
py ferramentas/conferir-urls-dfe.py <servicos.ini> <urls_tabela.go>

py ferramentas/conferir-urls-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrNFe/ACBrNFeServicos.ini" \
   packages/nfe/urls_tabela.go
```

A tabela de URLs é **dado**: mais de cem entradas transcritas de um `.ini` de
1388 linhas. Nenhum teste Go consegue dizer se uma delas foi digitada errada —
um teste só pode repetir o valor que está no Go, o que não prova nada. A única
fonte de verdade é o `.ini`.

O script reimplementa a resolução do ACBr (`ACBrDFe.pas:456-607`), que tem três
etapas — chave versionada exata, maior versão `<=` a pedida, chave sem versão —
mais o redirecionamento `Usar=` entre seções, com proteção contra ciclo.

Além de comparar entrada a entrada, confere as **invariantes** que o porte
assume para poder simplificar a tabela (por exemplo: "a URL do QR-Code não
varia com a versão"). Se o ACBr publicar um `.ini` em que isso deixe de valer,
o script acusa antes de a simplificação virar defeito.

### Limites

1. **Resolve o `Usar=` em tempo de PORTE**, e o Go guarda o resultado. Isso é
   deliberado: o redirecionamento é codificação do `.ini`, não regra fiscal.
   Em troca, o Go não reproduz a estrutura de seções do original.
2. **Tabela parcial precisa ser declarada** em `PARCIAIS`. São as que listam só
   as divergências em relação a outra (`urlsQRCodeModelo55` tem duas UFs, não
   27). Sem a declaração, as 25 ausentes virariam "FALTA".
3. **Acusa entrada REDUNDANTE** numa tabela de divergências — UF declarada que
   repete o valor da tabela base. É ruído que esconde o fato de a divergência
   existir em poucos lugares.

## `conferir-endpoints-dfe.py`

Confere a tabela de **endereços de transmissão** contra o `Servicos.ini`.
Complementa o `conferir-urls-dfe.py`, que cuida das URLs públicas.

```bash
py ferramentas/conferir-endpoints-dfe.py <servicos.ini> <tabela.go>

py ferramentas/conferir-endpoints-dfe.py \
   "$ACBR_FONTES/ACBrDFe/ACBrNFe/ACBrNFeServicos.ini" \
   packages/nfe/urls_servicos_tabela.go
```

Os endpoints têm forma diferente das URLs públicas, e a diferença é
**estrutural**: nem toda UF opera o próprio ambiente de autorização. A maioria
delega — dezesseis para o SVRS no modelo 55, dezenove no 65 — e o `.ini`
expressa isso com `Usar=` de seção para seção. Quem decide o endereço não é a
UF do emitente, é o **autorizador** que atende aquela UF naquele modelo.

O porte guarda as duas coisas separadas (mapa UF→autorizador e endereços por
autorizador), e o script confere **cinco** coisas:

1. o mapa UF→autorizador contra o destino real do `Usar=`;
2. os endereços, combinação a combinação, com a resolução completa do `.ini`;
3. UF que **mistura** definição local com delegação;
4. os overrides de `SoapAction` declarados pela chave `SoapAction=`;
5. a **versão da mensagem**, quando ela não acompanha a do leiaute da nota.

A terceira mudou de significado na etapa 6b. A premissa original — "o
redirecionamento é sempre de seção inteira" — estava **errada**: o ACBr resolve
serviço a serviço (`ACBrDFe.pas:550-571`), e o MA delega ao SVAN mantendo o
próprio `NfeConsultaCadastro_2.00`. A tabela continua válida porque o MA é a
única UF do grupo SVAN, e o script passou a emitir **AVISO** nesse caso e a
**falhar** se o grupo crescer.

A quinta é o terceiro eixo do `.ini`, e foi ela que achou sozinha que o
`consCad` sai em 2.00, o `admCscNFCe` em 1.00 e os dois serviços de evento
também em 1.00 — mesmo num componente 4.00. Ela precisa da pasta de schemas
para decidir:

```bash
py ferramentas/conferir-endpoints-dfe.py    "$ACBR_FONTES/ACBrDFe/ACBrNFe/ACBrNFeServicos.ini"    packages/nfe/urls_servicos_tabela.go    "$ACBR/Exemplos/ACBrDFe/Schemas/NFe"
```

Sem a pasta ela apenas lista os casos em que a chave do `.ini` ficou abaixo da
versão pedida, como AVISO.

### Limites

1. **Só conhece os serviços declarados em `SERVICOS`.** Acrescente a entrada
   quando a etapa que transmite o serviço novo entrar.
2. **Não valida o endereço.** Confere que o Go e o `.ini` dizem a mesma coisa;
   se o `.ini` do ACBr estiver desatualizado, os dois estarão errados juntos.
3. **A queda de versão da URL não é conferida.** A tabela Go guarda o valor já
   resolvido para cada versão; a lógica de "maior versão `<=` a pedida" é
   exercitada na geração. A queda da *mensagem*, essa sim, é a verificação 5.

## `conferir-eventos-dfe.py`

Confere os **geradores de evento** contra o `ACBrNFe.EnvEvento.pas`.

```bash
py ferramentas/conferir-eventos-dfe.py <EnvEvento.pas> <dir Go>

py ferramentas/conferir-eventos-dfe.py    "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/Servicos/ACBrNFe.EnvEvento.pas"    packages/nfe
```

Os eventos são o lugar do porte onde **mais nome se repete**. Três elementos
distintos saem com a mesma tag — `<gConsumo>` vem do consumo pessoal e da
ALC/ZFM, `<gPerecimento>` do transporte do adquirente e do fornecedor — e o
`<gControleEstoque>` tem **sete formas**, cada uma com um par de tags próprio.

Nada disso aparece num round-trip: o leitor acha a tag pelo nome, e o nome do
elemento **pai** é o mesmo nos dois lados da troca.

O script confere três coisas, por evento:

1. o **conjunto** de tags emitidas — faltando ou sobrando;
2. a **ordem** delas, que é `xs:sequence` no XSD;
3. o **literal fixo**, quando o Delphi passa uma constante no lugar do campo
   (o `tpAutor := '1'` de seis eventos e o `indQuitacao := '1'`).

E uma quarta, separada: as **sete formas do `gControleEstoque`**, resolvidas
pelo tipo do parâmetro no Delphi e comparadas com o Go.

Ele **segue os helpers dos dois lados** — os `Gerar_gXxx` do Delphi e as
funções locais do Go. Um script que parasse no corpo do evento acusaria "sobra"
em metade dos casos e não conferiria nada do que importa.

### Limites

1. **Lê os dois lados por expressão regular, não por parser.** Uma tag montada
   por concatenação, de qualquer lado, escapa.
2. **Não sabe nada sobre obrigatoriedade nem sobre tipo de campo.** Um `tcDe2`
   portado como `tcDe4` passa aqui — quem cuida disso é o
   `conferir-gerador-dfe.py`.
3. **A correspondência evento↔`gControleEstoque` é afirmada à mão**, na tabela
   `CONTROLE_ESTOQUE`. É o único lugar do script onde a ligação entre os dois
   lados não é descoberta sozinha, e é deliberado: as sete sobrecargas do
   Delphi diferem **só** pelo tipo do parâmetro, e nenhuma regex resolve isso.
4. **A comparação de ordem só olha as tags que os dois lados têm.** Uma tag
   faltando já foi acusada antes; compará-la de novo na ordem só produziria
   ruído.
