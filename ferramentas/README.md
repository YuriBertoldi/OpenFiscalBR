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

- `OK` — mesma quantidade de campos dos dois lados
- `DIF` — quantidades divergentes, com as duas listas lado a lado
- `--` — existe no ACBr e não tem writer no Go
- `??` — não encontrado no ACBr

O marcador `[condicional: ...]` sinaliza que o writer Delphi tem `ifthen`,
`EncodeDate` ou teste de `DT_INI` na montagem da linha. **Todo registro marcado
assim precisa de conferência manual** — é exatamente onde os campos condicionais
foram perdidos no port original.

O script segue chamadas a funções auxiliares do padrão `nomeHelper(b, r)`, então
campos condicionais extraídos para função própria continuam sendo contados.

### Limites — leia antes de confiar no resultado

1. **Compara quantidade, não nomes.** O registro `B440` passava como `OK` com 5
   campos dos dois lados e mesmo assim emitia `COD_MUN_SERV` onde o layout pede
   `COD_PART`. `OK` significa "vale investigar menos", não "está correto".
2. **Lê apenas um ramo quando o ACBr tem dois `Add`.** Registros como `0220`,
   `E116` e `1010` têm blocos alternativos por versão de leiaute; o extrator pega
   o primeiro.
3. **Não captura `LFILL` em caixa alta**, que o ACBr usa em pontos isolados (o
   campo `QTD` do `C170`, por exemplo).

Nos três casos a diferença aparece como `DIF` ou `OK` enganoso, e a conferência
é manual contra o `.pas`.

### Caminho dos fontes

O caminho do ACBr está fixo no topo do script (`ACBR`). Ajuste se os fontes
Delphi estiverem em outro lugar na sua máquina.
