# Armadilhas de um leitor de DFe portado — checklist de auditoria

Generalização das 18 armadilhas encontradas no porte do ACBrNFGas. **Nenhuma delas aparece em
build, `go vet` ou teste ingênuo** — todas exigem comparação com o `.pas` de origem. Para cada
item que se aplicar ao componente: procure no `.pas`, replique ou divirja documentando, e
escreva um teste que trave o comportamento.

## Estrutura e despacho

1. **Grupo achatado por CST/variante** — o leitor procura `Grupo00, Grupo10, ...` numa ordem
   fixa e despeja o primeiro achado numa struct única. A ORDEM é contrato; a lista pode incluir
   membro que não existe no XSD (ICMS41 na NFGas) e omitir campos de um membro (pDif/vICMSOp/
   vICMSDif de ICMS51). Replique lista, ordem e omissões.
2. **Campo lido do nó pai** — ex.: `indSemCST` lido de `imposto`, não do `ICMSxx`; e só quando
   não vazio, DEPOIS do `Exit` de "nenhum grupo achado".
3. **Campos alternativos com precedência e campo compartilhado** — `CNPJ`→`CPF`;
   `idOutros`→`idEstrangeiro` só sem CNPJ/CPF, os dois gravando NO MESMO campo do Delphi
   (properties com o mesmo F-field). Um campo em Go + registro da tag de origem.
4. **Tag ≠ nome do campo** — `gRespTec` (tag) vs `infRespTec` (campo/seção ini); `vRetCofins`
   (XML) vs `vRetCOFINS` (ini). Grafia é contrato, inclusive a caixa.
5. **Grupos achatados na leitura** — `ICMSTot` e `vRetTribTot` caem direto em `Total`;
   `gMedida` cai achatado na seção ini de `gMedicao`. O shape do Go segue a CLASSE, o leitor
   faz o achatamento.
6. **Campo declarado e nunca lido** — `vBCIRRF`, `indIEDest` (só no ini), `cPais/xPais` em
   `enderCorresp`, `infProt/@Id`. Mantenha o campo, replique a omissão, documente no README
   (declarado × exercitado).
7. **Dado em ATRIBUTO de item de coleção** — `@nItem`, `@nMed`, `@nContrat`, `@chNFGasAnt`,
   `@nPag`, `@idTransacao`. Fácil de perder quando só se olham elementos.

## Conversão de valor

8. **Precisão `tcDeN` por campo** (ver `tipos-campo-dfe.md`) — inclusive a MESMA tag com
   precisões diferentes em grupos diferentes.
9. **`tcInt` com `OnlyNumber`** — não é `Atoi`; e CEP/códigos como `int` perdem zero à
   esquerda (comportamento do ACBr; replicado).
10. **Tipo da classe contradiz o XSD** — `qFaturada Integer` vs `TDec_1100_1104`; `cClass
    Integer` vs `xs:string [0-9]{7}`. O leitor lê certo e a property trunca. Siga o XSD,
    marque `DIVERGENCIA`, teste, avise o usuário.
11. **Enum por ordinal** (`tcInt` em property de enum) — corrigir para código (item 3 de
    tipos-campo-dfe.md).
12. **Competência `AAAAMM`** e **`StrToDate` dependente de locale** — `ParseCompetenciaDef`.
13. **Limpeza de CDATA sem `rfReplaceAll`** — só a primeira ocorrência sai (`qrCod*`).
    Determinístico: replique (`pcn.RemoverCDATAPrimeiraOcorrencia`).
14. **Mesmo nome, tipos diferentes em classes distintas** — `serie` int em `ide` e string em
    `gNF`; `qFaturada` int em `Prod` e double em `gProcRef` no próprio ACBr.

## Coleções e estado

15. **`Clear` faltando antes do laço** — releitura acumularia. Reinicialize o slice,
    `DIVERGENCIA`, teste de releitura idempotente.
16. **Coleção indexada pelo contador do laço após `New`** (`refNFe[i]`) — corrompe lista
    pré-existente; não determinístico ⇒ implemente o correto e documente.
17. **Variável não inicializada atribuída a campo** (`chDFePagAnt := sFim` com `sFim` nunca
    escrito, no RTC IniReader) — não replicável; leia da chave que o writer grava.
18. **Split de lote por busca de string** — case-sensitive e dependente de grafia
    (`</NFGas>`/`</NFGasProc>`/`</procNFGas>`). Use decoder de verdade preservando a semântica
    (1 nota por elemento; envelope tem precedência sobre o documento interno).

## Protocolo, eventos e consulta

19. **Protocolo só em certos `cStat`** — a consulta só lê `prot<Comp>` quando `cStat` está numa
    lista fixa (`{100,101,104,150,151,155}` na NFGas). Fora dela, ignora mesmo presente.
20. **`try/except` que devolve `False`** engolindo a causa — vira `error`; falha de um evento
    vinculado não derruba a leitura do retorno.
21. **Parte enviada do evento não é lida** pelo ACBr (`procEvento` → só o retorno) — a
    justificativa de cancelamento (`xJust`) só existe lá. Ler é `ACRESCIMO` legítimo; marque.
22. **Getter com efeito colateral** (`XMLAssinado` que assina) — não replicar; método explícito.

## Ambiente e encoding

23. **BOM UTF-8 e declaração `<?xml?>`** toleradas na entrada; XML original preservado
    byte a byte por nota (`OuterXML` por offset, não reserialização).
24. **Nada de panic em biblioteca** — todo caminho torto vira `error` com contexto
    (arquivo/índice/chave). Teste com corpus de entradas malformadas + `recover`.
