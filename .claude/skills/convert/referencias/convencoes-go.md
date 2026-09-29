# Convenções Go do OpenFiscalBR (além do CLAUDE.md)

Referência da `/convert` e da `/portar-leitor-dfe`. Consolida decisões tomadas nos portes de
`sped` e `nfgas` que não cabem na tabela de mapeamento do `CLAUDE.md`.

## Organização

- `packages/` é a convenção do repo (não `internal/`/`pkg/`). Dentro dela, Go idiomático:
  um package por diretório, nome = pasta, lowercase, sem underscore, sem *stutter*
  (`pcn.Node`, não `pcn.PCNNode`).
- **Sem subpackage prematuro.** Eventos, consulta e ini de um DFe ficam no mesmo package do
  documento — separar criaria ciclo de import ou um package `types` anêmico.
- Arquivos por coesão, `snake_case`, cada um espelhando uma unit `.pas` de origem (serve à
  auditoria de fidelidade). `testdata/` para fixtures (nome reservado do toolchain).
- **Zero dependência externa**: `go.mod` sem `require`, não existe `go.sum` (o Dockerfile das
  demos e a `/validar-package` assumem isso). Precisa de `.ini`? Use `pcn.INI`. Precisa de
  decimal? `float64` com a precisão do leiaute documentada.

## API

- **Aceitar interface, devolver struct**: `LerLote(r io.Reader) ([]*NotaFiscal, error)`, com
  conveniências `LerXXXString/Bytes/Arquivo` por cima.
- **Zero-value útil**; construtor `NovoXxx()` só quando o `Create` do Delphi aplica default real.
- Métodos de rede futuros já nascem com `context.Context` no contrato (mesmo como stub).
- Superfície exportada mínima: leitores auxiliares (`lerIde`, `lerICMS`) não exportados.
- Stub de fase futura: assinatura definitiva + `return ErrNaoImplementado`, nunca panic nem
  silêncio.

## Erros

**Convenção OFICIAL do projeto** (seção "Convencao de erros" do `CLAUDE.md`; exemplo de
referência: `packages/nfgas/errors.go`). O `comum.ACBrError` e o `Check`/panic são legado do
`sped` — não adotar em package novo.

- Sentinelas com `errors.New` + tipos com `Unwrap()` (e `Unwrap() []error` em agregadores),
  para `errors.Is`/`As` funcionarem de ponta a ponta.
- Erro de leitura carrega CONTEXTO: caminho da tag (`infNFGas/ide/dhEmi`), índice no lote,
  arquivo, chave de acesso. Numa importação de milhares de XMLs é isso que torna o erro
  acionável.
- **Nada de panic em código de biblioteca.** O `Check`/panic do `comum.TXTClass` é legado do
  sped; leitor devolve `error` sempre.
- Documento torto não derruba o lote: as notas legíveis retornam + `*ErrosLote` com as falhas.

## Enums

- `String()` com o código EXATO do leiaute + `ParseXxx(s) (T, error)` (o `raise` do Delphi).
- Membro "nenhum" mapeando `""` quando o array do Delphi o tiver; sentinela `TXxx(-1)` do
  Delphi vira membro `-1` explícito.
- Prefixos que colidem no Delphi (`tm*`, `tf*`) não colidem em Go porque os tipos são
  distintos — mantenha um tipo por enum, nunca funda dois.
- Tabelas assimétricas (CSTIcms entrada×saída) ganham os dois lados (`String()` = saída,
  `CodigoEntrada()`, `Parse` pela entrada) e teste provando que NÃO são inversas.

## Trabalho concorrente no mesmo checkout

Quando mais de um agente/sessão trabalha neste repositório ao mesmo tempo:

1. **Divida territórios por diretório** e anuncie (SendMessage/combinado com o usuário) quais
   pastas são suas. Zona alheia é intocável mesmo para "consertar rapidinho" — inclusive quando
   `go build ./...` do repo inteiro quebra por causa dela: registre e siga; valide o seu com
   `go build ./packages/<seus>/...`.
2. **Arquivos compartilhados** (`README.md` da raiz, `.openfiscalbr-meta.json`, `CLAUDE.md`,
   tabelas de `referencias/`) ficam para o FIM do trabalho, em edição isolada: confira
   `git status` + mtime antes, edite acrescentando linhas (nunca sobrescrevendo o arquivo), e
   se estiver quente (mtime de minutos), espere ou avise o outro lado.
3. Helper que "naturalmente moraria" num package do outro território nasce no seu (ou num
   package novo) e a observação é levada ao usuário — não se edita package alheio por
   conveniência.
4. Arquivo seu sumiu ou mudou sem você editar? Não é erro seu de contexto: cheque com o outro
   agente antes de "corrigir". (Já aconteceu: um agente apagou `packages/pcn/` do outro por
   presumir resíduo.)
