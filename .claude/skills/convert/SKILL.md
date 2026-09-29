---
name: convert
description: "Converte um componente ACBr Delphi para package Go nativo do OpenFiscalBR. Escaneia os fontes .pas, resolve a ordem de dependencias entre layers, gera o package em packages/<pkg>/ com testes e README, gera a demo em demos/<pkg>/ e registra os hashes no .openfiscalbr-meta.json. Quando o package ja existe, entra em modo atualizacao e reporta as diferencas antes de aplicar. Use quando o usuario pedir para converter, portar ou migrar um componente ACBr (ACBrNFe, ACBrBoleto, ACBrPIXCD, PCNComum, ACBrDFe, ACBrSAT...) de Delphi para Go."
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

## PASSO 7 — Modo atualização

1. Leia os hashes anteriores do componente em `.openfiscalbr-meta.json`.
2. Calcule o SHA256 atual de cada `.pas` do componente.
3. Todos iguais → informe "Nenhuma alteração detectada nos fontes Delphi." e encerre.
4. Para cada `.pas` alterado, compare com o `.go` correspondente e identifique: tipos novos ou
   removidos, campos/properties novos ou removidos, valores novos em enums, métodos novos ou
   removidos, mudanças de assinatura.
5. Reporte antes de aplicar:

   ```
   ## Alterações detectadas em ACBrNFe

   ### Arquivos Delphi modificados
   - ACBrNFe.Classes.pas
   - ACBrNFe.Conversao.pas

   ### Tipos novos
   - TNewType em ACBrNFe.Classes.pas

   ### Campos novos
   - TCampo.NovoField: String em ACBrNFe.Classes.pas

   ### Valores novos em enum
   - TipoEmissao: teNovoTipo em ACBrNFe.Conversao.pas
   ```

6. Pergunte se deve aplicar. Ao aplicar:
   - **preserve** todo bloco marcado com `// CUSTOM:` — é código escrito à mão que não tem
     origem no Delphi e seria perdido
   - acrescente os tipos, campos e valores de enum novos
   - marque os removidos com `// DEPRECATED: removido do Delphi em <data>` em vez de apagar,
     para não quebrar quem já importa o símbolo

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
4. Para o `sped`, use `ferramentas/comparar-campos.py` como primeira passada; leia o
   `ferramentas/README.md` para os limites dela antes de confiar no `OK`.
5. Chame o subagente `revisor-go` sobre o package.

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
