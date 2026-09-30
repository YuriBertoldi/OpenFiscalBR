---
name: portar-leitor-dfe
description: "Porta o leitor de XML de um componente DFe do ACBr (NFe, CTe, NF3e, NFCom, NFGas, BPe...) para Go, com fidelidade auditavel campo a campo. Cobre o padrao leitor-imperativo sobre o mini-DOM do packages/pcn, a tabela de tipos de campo tcStr/tcInt/tcDeN, a leitura de eventos e retornos de consulta, e o checklist de armadilhas que build/vet/teste nao pegam. Use ao portar a LEITURA de qualquer DFe, ou ao revisar um leitor ja portado. Complementa a /convert: ela rege o processo do componente inteiro; esta skill rege a parte do XmlReader."
argument-hint: "<ComponentName> (ex.: ACBrNF3e)"
---

# portar-leitor-dfe — leitor de XML de DFe, do .pas para Go

Todo componente DFe do ACBr tem a mesma anatomia de leitura, e o porte dela tem sempre os
mesmos defeitos silenciosos. Esta skill é a receita destilada do porte do **ACBrNFGas**
(`packages/nfgas`), que é o exemplar de referência — na dúvida sobre qualquer padrão, abra o
arquivo correspondente lá.

## Pré-requisitos

- `packages/pcn` existe (mini-DOM, tipos de campo, validador, chave de acesso, INI).
- Se o componente tiver grupos da Reforma Tributária (`gCompraGov`, `IBSCBS`, `IBSCBSTot`,
  `pgtoVinc`, `gPagAntecipado`), **não os reimplemente**: use `packages/rtc`.
- Leia `referencias/tipos-campo-dfe.md` e `referencias/armadilhas-leitor-dfe.md` **antes** do
  primeiro `.go`.
- Divergência que você encontrar ou criar em relação ao ACBr: registre em
  `.claude/skills/convert/referencias/divergencias-acbr.md` — é o registro oficial que impede
  alguém de "corrigir" de volta uma divergência aprovada.

## PASSO 1 — Mapear as units de leitura do componente

> Componente irmão de um já portado (NFGas ↔ NFAg ↔ NF3e ↔ NFCom)? Comece pelo diff
> estruturado dos `.pas` (`diff --strip-trailing-cr -w -i`) — a técnica está no PASSO 1 da
> `/portar-emissor-dfe`. O delta é o trabalho; o modelo de dados se valida struct a struct.

Em `<Fontes>/ACBrDFe/<Componente>/`:

| Unit | O que é | Vai para |
|---|---|---|
| `Base/<Comp>.XmlReader.pas` | o leitor — o CONTRATO do porte | `xml_reader.go` |
| `Base/<Comp>.Classes.pas` | modelo de dados | `classes.go` |
| `Base/<Comp>.Conversao.pas` | enums e arrays de código | `types.go` + `conversao.go` |
| `<Comp>NotasFiscais.pas` (ou equivalente) | carga em lote, split, encoding | `notas_fiscais.go` |
| `Base/Servicos/<Comp>.RetEnvEvento.pas` + `EventoClass.pas` | eventos | `evento.go` + `evento_reader.go` |
| `Base/Servicos/<Comp>.RetConsSit.pas` | retorno de consulta | `ret_cons_sit.go` |
| `Base/<Comp>.ValidarRegrasdeNegocio.pas` | regras (muitas vazias — porte o estado real) | `regras_negocio.go` |

Confira também o XSD em `Exemplos/ACBrDFe/Schemas/<Comp>/` — ele decide empates entre a classe
e o leitor (ex.: campo declarado `Integer` na classe mas `xs:string` com pattern no XSD).

## PASSO 2 — Portar o leitor 1:1, método a método

Um `lerXxx` em Go por `Ler_Xxx` do Delphi, na mesma ordem, com o mesmo guard
(`if node == nil || destino == nil { return }` ≡ `if not Assigned(ANode) then Exit`).

Regras inegociáveis:

1. **`Find` × `FindAnyNs` é decisão POR CAMPO**, copiada do `.pas`. `Find` compara o nome
   qualificado (com prefixo); `FindAnyNs`, o local name. Não "padronize".
2. **Precisão `tcDeN` é POR CAMPO** e varia para a mesma tag em grupos diferentes (na NFGas,
   `vItem` é De10 em `prod` e De8 em `gProcRef`). Copie do `.pas`, nunca deduza do nome.
3. **Enum lido com `tcStr` + `StrToXxx`** vira `ParseXxx(pcn.ConteudoStr(...))`. **Enum lido
   com `tcInt` é BUG do ACBr** (trata código como ordinal) — corrija para parse por código,
   documente com `DIVERGENCIA` e teste (caso `motDesICMS` da NFGas).
4. Atribuição condicionada a conteúdo (`if Lvalor <> '' then`) permanece condicionada.
5. Grupo achatado por CST (`ICMSNode := Find('ICMS00'); if not Assigned então ICMS10...`) vira
   `node.PrimeiroDe("ICMS00", "ICMS10", ...)` com a MESMA lista, na MESMA ordem — inclusive
   membros que não existem no XSD, se o `.pas` os procurar.
6. Campo lido do nó PAI (ex.: `indSemCST` lido de `imposto`, não de `ICMSxx`) permanece no pai.
7. Laço sem `Clear` antes no Delphi: reinicialize o slice mesmo assim e marque `DIVERGENCIA` —
   acúmulo em releitura só é observável em reuso de objeto, indefinido no original.
8. Exceção do Delphi vira `error` sentinela (`errors.Is`-compatível). Leitor **não entra em
   pânico** com XML torto: importação em lote não pode cair por causa de uma nota.

## PASSO 3 — Carga em lote

Não copie o split por string do ACBr (`Pos('</NFe>')` etc., sensível a caixa e espaçamento).
Use o padrão de `nfgas/notas_fiscais.go`: `xml.Decoder` + `Skip()` capturando o range de bytes
de cada elemento raiz (preferindo o envelope `<x>Proc` e pulando o documento interno para não
contar duas vezes). Nota ilegível entra em `*ErrosLote` com índice/arquivo/chave; as legíveis
retornam sempre.

## PASSO 4 — Testes que provam fidelidade

Fixtures em `testdata/` escritas a partir do XSD (o ACBr não distribui XML de exemplo):
documento completo com TODOS os grupos, envelope com protocolo, lote misto, um caso por grupo
de CST achatado, evento, retorno de consulta e — se houver INI — um `.ini` no formato pt-BR do
Delphi (vírgula decimal, data `dd/mm/aaaa`).

Cada item do checklist de `referencias/armadilhas-leitor-dfe.md` que se aplicar ganha um teste
com comentário citando o `.pas`. Chave de acesso da fixture: monte-a válida (DV correto) para
os testes de regra 227 e de validação de chave passarem por mérito, não por acaso.

## PASSO 5 — Auditar e reportar

1. Releia o `.pas` lado a lado com o `.go`, campo a campo — contagem não é auditoria.
2. Rode o subagente `revisor-go` sobre o package.
3. `/validar-package <pkg>`.
4. No `README.md`: tabela unit→arquivo, seção "Fidelidade" (o que foi replicado de propósito),
   seção "Divergências deliberadas" (com o porquê de cada uma) e cobertura declarado×exercitado.
5. Divergência não determinística ou perda de dado do ACBr (tipo `Integer` para campo decimal,
   ordinal por código): corrigir, documentar, testar e **avisar o usuário** — é decisão dele
   manter.
