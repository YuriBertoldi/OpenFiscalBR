# Ferramentas de apoio ao port

Scripts de desenvolvimento. Não fazem parte do módulo Go e não são compilados.

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
