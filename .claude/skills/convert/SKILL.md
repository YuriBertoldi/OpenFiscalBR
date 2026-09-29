---
name: convert
description: "Converte um componente ACBr Delphi para package Go nativo do OpenFiscalBR, e atualiza um package ja existente quando o ACBr publica versao nova. Escaneia os fontes .pas, resolve a ordem de dependencias entre layers, gera o package em packages/<pkg>/ com testes e README, gera a demo em demos/<pkg>/ e registra os hashes no .openfiscalbr-meta.json. No modo atualizacao compara os hashes gravados, procura campo novo sob condicional de versao ou data, tabela de codigos remanejada, registro novo ou revogado e validacao nova, e reporta antes de aplicar. Use quando o usuario pedir para converter, portar ou migrar um componente ACBr (ACBrNFe, ACBrBoleto, ACBrPIXCD, PCNComum, ACBrDFe, ACBrSAT...) de Delphi para Go, ou para trazer para o Go uma atualizacao do ACBr."
argument-hint: "<ComponentName> <DelphiSourcePath>"
---

# convert — conversor ACBr Delphi → Go

Converte um componente ACBr de Delphi para Go nativo, respeitando a ordem de dependências
entre layers do projeto.

## Uso

```
/convert <ComponentName> <DelphiSourcePath>

/convert ACBrNFe C:\MeusACBr\Fontes
/convert PCNComum C:\MeusACBr\Fontes
```

- `$1` = `ComponentName` — nome do componente ACBr (ex.: `ACBrNFe`, `ACBrBoleto`, `PCNComum`)
- `$2` = `DelphiSourcePath` — caminho da pasta `Fontes` dos fontes Delphi

Se qualquer um dos dois vier vazio, **pergunte ao usuário antes de prosseguir**. Não adivinhe
o componente nem o caminho dos fontes — converter o componente errado gera um package inteiro
que precisa ser descartado.

Antes de começar, leia o `CLAUDE.md` da raiz: ele traz a tabela de mapeamento Delphi → Go, as
convenções de nome e a estrutura padrão de package.

---

## PASSO 1 — Validar parâmetros e localizar os fontes

1. Confirme que `$2` existe (Glob).
2. Localize o diretório do componente usando a tabela **Componente → diretório dos fontes
   Delphi** de `referencias/componentes.md`.
3. Confirme que há arquivos `.pas` no diretório encontrado. Se não houver, pare e reporte —
   provavelmente o caminho aponta para a raiz errada.

## PASSO 2 — Resolver dependências

1. Determine o package Go destino pela tabela **Componente → package Go** de
   `referencias/componentes.md`.
2. Confira as dependências obrigatórias na tabela do mesmo arquivo. Uma dependência só conta
   como satisfeita se `packages/<dep>/` existir **e** contiver arquivos `.go`.
3. Se faltar alguma, pergunte:

   > O componente X depende de [dep1, dep2], que ainda não foram convertidos. Quer que eu
   > converta essas dependências primeiro?

   Se sim, execute este mesmo workflow recursivamente para cada uma, na ordem de layers.

## PASSO 3 — Verificar se o package já existe

- `packages/<pkg>/` não existe ou está sem `.go` → PASSO 4 (criação completa).
- Já existe com `.go` → PASSO 7 (modo atualização). Não sobrescreva um package existente sem
  passar pelo modo atualização: ele é o que preserva os blocos `// CUSTOM:`.

## PASSO 4 — Escanear os fontes Delphi

1. Liste os `.pas` recursivamente (Glob `**/*.pas`), aplicando as exclusões da seção
   **Arquivos Delphi a excluir do escaneamento** de `referencias/componentes.md`.
2. Para cada arquivo, extraia:
   - nome da unit (`unit XXX;`)
   - cláusulas `uses` (seções `interface` e `implementation`)
   - declarações da seção `type`: classes (`TXxx = class(TParent)`), enums, records, arrays
   - hierarquia (classe pai)
   - properties: nome, tipo, accessors, e **se o setter tem lógica** (decide campo exportado
     vs. getter/setter)
   - métodos: nome, parâmetros, retorno, visibilidade, modificadores (`virtual`, `override`,
     `abstract`, `class`)
   - constantes e tipos de evento (`TOnXxx = procedure(...) of object`)
3. Monte o inventário completo antes de gerar qualquer arquivo. Gerar arquivo por arquivo sem
   inventário produz referências a tipos que ainda não existem.

## PASSO 5 — Gerar o package Go

Crie `packages/<pkg>/` seguindo a **Estrutura Padrão de Package Go** do `CLAUDE.md`.

Os moldes literais de cada construção — header de licença, enums, classes → structs,
properties com lógica, métodos abstratos → interface, XML, web services, errors — estão em
`referencias/geracao-go.md`. Leia esse arquivo antes de escrever o primeiro `.go`.

Gere também:

- **Testes reais** em `<pkg>_test.go`, cobrindo o que a seção "Testes Obrigatórios" do
  `CLAUDE.md` exige: defaults dos construtores, `String()` dos enums devolvendo o código exato
  da legislação, formato de saída, formatação de campo (data `ddmmaaaa`, monetário com vírgula,
  inteiro com zeros à esquerda), validações de negócio, integração e casos de borda. Sem
  dependência externa — o projeto usa `testing` puro, sem testify.
- **`README.md`** do package: descrição, componente Delphi correspondente, tipos principais com
  exemplo de uso, tabela de registros/tipos implementados, como rodar os testes.

## PASSO 6 — Gerar a demo

Crie `demos/<pkg>/` seguindo a **Estrutura Padrão de Demo** do `CLAUDE.md`. Use `demos/sped/`
como molde de referência — é a demo existente e já segue o padrão.

Ou invoque `/gerar-demo <pkg>`, que faz exatamente este passo de forma isolada e tem os
conjuntos de rotas por tipo de componente (DFe, Boleto, PIXCD, SAT).

## PASSO 7 — Modo atualização (o ACBr mudou)

O ACBr é atualizado a cada nova versão de leiaute. Este passo traz essas mudanças para um
package **que já existe**, sem reconverter do zero e sem perder o que já foi corrigido à mão.

### 7.1 — Ter uma linha de base

```bash
py ferramentas/comparar-campos.py <bloco> <registro>   # so funciona com fontes acessiveis
```

O modo de atualização depende de `.openfiscalbr-meta.json` ter os hashes da conversão
anterior. **Se `components` estiver vazio para o componente, não há base de comparação** e
qualquer resposta de "nenhuma alteração" é falsa. Nesse caso, rode `/sincronizar-meta` contra
a cópia antiga dos fontes antes de mais nada, ou trate como auditoria completa (7.4).

### 7.2 — O que mudou nos fontes

1. Calcule o SHA256 de cada `.pas` do componente e compare com `delphiFileHashes`.
2. Todos iguais → "Nenhuma alteração detectada nos fontes Delphi." e encerre.
3. Para cada `.pas` alterado, faça o diff contra a versão anterior se ela existir. Sem ela,
   compare o `.pas` atual com o `.go` correspondente.

### 7.3 — O que procurar, em ordem de risco

Atualização de ACBr quase nunca é "campo novo no fim". Procure, nesta ordem:

| Mudança | Como aparece no diff | Risco se passar |
|---|---|---|
| **Campo novo sob condicional de versão** | `if COD_VER >= vlVersaoNNN` novo, ou constante de versão nova no enum | linha curta no período novo |
| **Campo novo sob condicional de data** | `ifthen(DT_INI >= EncodeDate(...))` novo | idem |
| **Tabela de códigos remanejada** | `case` novo dentro de `if DT_INI < ...` | código válido com significado errado |
| **Registro novo** | `procedure WriteRegistroXxx` nova + classe nova | registro simplesmente ausente |
| **Registro que deixou de existir** | `Exit` novo no topo, ou guarda de versão | registro emitido fora de vigência |
| **Mudança de ordem** | posição de um `LFill` trocada | arquivo rejeitado, e nada acusa |
| **Validação nova** | `Check(` novo | documento inválido gerado em silêncio |
| **Tamanho/decimais alterados** | 2º e 3º argumentos do `LFill` | campo fora de layout |

A constante de versão nova é o gatilho mais confiável: quando o ACBr ganha `vlVersaoNNN`,
**todo `if COD_VER` do componente merece releitura**, não só os que apareceram no diff.

### 7.4 — Aplicar

Reporte antes de mexer:

```
## Alteracoes detectadas em ACBrNFe (v1.2.3 -> v1.3.0)

### Fontes alterados
- ACBrNFe.Classes.pas, ACBrNFe.Conversao.pas

### Registros afetados
- C170: campo VL_XXX novo a partir da versao 120
- C500: tabela de IND_YYY remanejada para DT_INI >= 2027-01-01
- C999: registro novo (sem equivalente em Go)

### Regras de vigencia novas
- 3 ocorrencias de COD_VER >= vlVersao120
```

Ao aplicar:

- **Preserve todo bloco `// CUSTOM:`** — é código sem origem no Delphi e seria perdido.
- **Preserve as divergências deliberadas já documentadas.** O package registra, no README,
  comportamentos do ACBr reproduzidos de propósito e defeitos que decidimos não replicar.
  Uma atualização não pode desfazê-los por acidente: releia essa seção antes de aplicar.
- Campo novo sob condicional → helper próprio devolvendo `""` fora da vigência, ou `linha +=`
  dentro de `if`, conforme o padrão já usado no package.
- Removido do Delphi → `// DEPRECATED: removido do Delphi em <data>`, não apague; quem
  importa o símbolo continua compilando.
- Constante de versão nova → acrescente ao enum `VersaoLeiauteFiscal` **na ordem**, já que a
  comparação é por `iota`.

### 7.5 — Fechar a atualização

Nesta ordem, sem pular:

1. `py ferramentas/comparar-campos.py <bloco> <registros afetados>` — confira que a contagem
   voltou a bater. Leia `ferramentas/README.md`: `OK` não prova correção, e a ferramenta não
   compara nomes.
2. **Teste por faixa** para cada vigência nova — um caso antes do corte e um depois. Sem isso
   a regra nova não está coberta e some na próxima refatoração.
3. `/validar-package <pkg>` — inclui a verificação de integridade das contagens do arquivo.
4. Subagente `revisor-go` sobre os arquivos tocados.
5. `/sincronizar-meta` para gravar os hashes novos. **Sem este passo a próxima atualização
   fica cega**, porque volta a não ter linha de base.

## PASSO 7b — Auditoria de fidelidade

**Não pule este passo.** Um package pode compilar, passar em todos os testes e mesmo assim
gerar documento inválido, porque o que se perde num port não é a estrutura — é a lógica em
volta dela.

Na conversão do `sped`, **todas** as divergências encontradas foram da mesma natureza: o port
copiou a sequência de campos e descartou a condicional que a governava. Nenhuma foi detectada
por build, `vet` ou teste; todas exigiram comparação com o `.pas`.

### O que um port perde, em ordem de frequência

| Categoria | Como aparece no Delphi | Consequência de perder |
|---|---|---|
| **Vigência de campo** | `ifthen(DT_INI >= EncodeDate(2017,1,1), LFill(CEST), '')` | campo a mais ou a menos para o período declarado |
| **Vigência de bloco/registro** | `if DT_INI >= EncodeDate(2019,1,1) then begin ... end` | bloco inteiro emitido fora do período em que existe |
| **Versão de leiaute** | `if COD_VER <= vlVersao114 then` / `>= vlVersao110` | mesma coisa, governada pela versão e não pela data |
| **Tabela de códigos por período** | `if DT_INI < EncodeDate(2012,1,1) then case IND_FRT of ...` | código **remapeado**: o mesmo valor significa coisas diferentes por época |
| **Estado do documento** | `if Pos(strCOD_SIT,'02, 03, 04, 05') > 0 then` | cancelado/denegado sai como documento normal |
| **Variante por modelo** | `if COD_MOD = '65' then` | campos que o modelo não admite saem preenchidos |
| **Valor especial** | `IfThen(booExterior, LFill('9999999'), LFill(COD_MUN, 7))` | código inválido no lugar do literal exigido |
| **Validação** | `Check(condicao, 'mensagem', [args])` | documento inválido gerado em silêncio |
| **Callback** | `if Assigned(OnCheckRegistroX) then ...` | ponto de extensão declarado e nunca chamado |

### Procedimento

1. Para cada tipo/classe convertido, abra o `.pas` de origem lado a lado e confira a sequência
   de campos **um a um**, não por contagem.
2. Procure no `.pas`, especificamente: `ifthen`, `IfThen`, `EncodeDate`, `COD_VER`, `DT_INI`,
   `Check(`, `case ... of` que monte string, e qualquer `if` em volta da montagem da linha.
   Cada ocorrência é uma regra que precisa existir no Go.
3. Confira também **hierarquia** (quem é filho de quem) e **nomes de campo** — não só a
   quantidade. Um struct copiado de outro registro tem a contagem certa e o conteúdo errado.
4. Confira a **cadeia de chamadas**: todo writer convertido é chamado por alguém, e pelo mesmo
   pai que o chama no `.pas`. Writer órfão não quebra build nem teste — a linha simplesmente
   não sai. No `sped` isso escondeu três registros (`E112`, `E113`, `E115`), e o que os revelou
   foi o `unusedfunc` do `go vet`/diagnósticos, não a auditoria de campos.
5. Para o `sped`, use `ferramentas/comparar-campos.py` como primeira passada; leia o
   `ferramentas/README.md` para os limites dela antes de confiar no `OK`.
6. Chame o subagente `revisor-go` sobre o package.

### Nomeie os campos como o Delphi nomeia

A auditoria automática de nomes só funciona se o lado Go não abreviar. Duas regras práticas,
ambas aprendidas corrigindo falso positivo:

- **Campo de struct** repete o nome do layout (`VL_SLD_CREDOR_TRANSPORTAR` → `VlSldCredorTransportar`,
  não `VlSldCredorTransp`). Abreviação economiza dez caracteres e custa uma auditoria.
- **Parâmetro de helper compartilhado** repete o nome do campo que recebe. Um helper com
  parâmetros abreviados cega a auditoria em todos os registros que o usam de uma vez.

Quando o Delphi batiza uma variável local com nome próprio (`ChaveEletronicaCTe` no `D100`),
registre o par em `ALIAS_GO`, no topo do comparador — não renomeie o campo Go para o nome da
variável.

### Quando o Delphi de origem estiver errado

Acontece. No `sped`, o `WriteRegistroB350` do ACBr emite uma linha com o literal `B035`, e as
tabelas de `IND_FRT`/`IND_PGTO` deixam a variável conservar o valor do documento anterior
quando o `case` não cobre o membro.

Regra: **replique o comportamento determinístico**, mesmo que pareça errado — é o que mantém
a compatibilidade com o original, e corrigir por conta própria diverge silenciosamente. Cubra
com teste e comentário explicando que é intencional, para ninguém "consertar" depois.

**Não replique comportamento não determinístico** — o que depende de ordem de iteração,
memória de laço ou variável não inicializada. Nesses casos, implemente o comportamento
correto, documente a divergência e **avise o usuário**, porque é decisão dele.

## PASSO 8 — Validar e registrar

1. Rode `/validar-package <pkg>` (build, vet, testes, gofmt, header de licença, module path).
   Corrija iterativamente: import faltando, tipo não encontrado (verifique se a dependência foi
   convertida), assinatura divergente.
2. Rode `/sincronizar-meta` para gravar os hashes e os arquivos gerados em
   `.openfiscalbr-meta.json` e atualizar o status no `README.md` da raiz.
3. **Reporte a cobertura com honestidade.** No `README.md` do package, separe "tipos
   declarados" de "tipos efetivamente exercitados pelo código". Um struct que o consumidor
   consegue preencher e que nunca aparece na saída é pior do que um tipo ausente: o ausente
   quebra a compilação, o mudo passa despercebido até a rejeição pelo fisco. O `sped` ficou
   marcado como "Completo" no README por meses tendo 59 writers para 198 structs.

Formato da entrada no meta:

```json
{
  "components": {
    "<ComponentName>": {
      "goPackage": "<pkg>",
      "delphiSourcePath": "<DelphiSourcePath>",
      "lastConvertedAt": "<timestamp ISO 8601>",
      "delphiFileHashes": { "<arquivo.pas>": "sha256:<hash>" },
      "goFiles": ["<arquivo1.go>", "<arquivo2.go>"]
    }
  }
}
```

---

## Regras

- **Preserve os nomes fiscais em PT-BR** nos campos dos structs (`cUF`, `CNPJ`, `nNF`, `vBC`).
  Eles mapeiam 1:1 para os schemas XML da SEFAZ; "corrigir" para inglês quebra a serialização.
- **Não converta reports visuais nem `*Reg.pas`.** Dependem de engine gráfica e de
  design-time do Delphi, sem equivalente em Go.
- **Nunca pule a ordem de layers.** Converter um Layer 3 antes do `comum`/`pcn`/`dfe` gera um
  package que não compila e cujo erro só aparece dezenas de arquivos depois.
- **Componentes com muitas implementações** (Boleto: 70+ bancos; PIXCD: 20+ PSPs) — pergunte
  quais implementações converter antes de começar. Converter todas sem perguntar produz
  milhares de linhas que o usuário talvez não queira revisar.
- **Na dúvida sobre um mapeamento Delphi → Go que não esteja na tabela do `CLAUDE.md`, pare e
  pergunte.** Um mapeamento inventado se propaga por todo o package e fica caro de desfazer.
- **Não confie em contagem de campos como prova de fidelidade.** Registro com a quantidade
  certa e o campo errado existe e é comum — é o caso do `B440`, que passava em toda verificação
  automática emitindo `COD_MUN_SERV` onde o layout pede `COD_PART`.
- **Propague o estado do componente para as partes.** No `sped`, `DtIni` não chegava aos
  blocos, e todo writer que decidia layout por vigência lia data zerada. O sintoma foi um
  registro que simplesmente nunca era emitido, sem erro algum.
- **Contador declarado precisa ser lido em algum lugar.** Se você gerar um campo de contagem,
  ligue-o ao consumidor na mesma conversão. No `sped` havia 130 contadores incrementados que
  nada lia, enquanto a contagem real usava outro mecanismo, incompleto.
- **O package não expõe URL.** Nada de `net/http`, `encoding/json` ou template em
  `packages/`; todo transporte vai para `demos/<pkg>`. O consumidor da biblioteca precisa
  gerar o documento sem subir servidor.
- **Confira as contagens gerando um arquivo de verdade.** Registro de totalização é valor
  calculado, e nenhuma conferência de layout o alcança. O `9990` do `sped` declarava uma
  linha a menos e passou por toda a auditoria campo a campo.
