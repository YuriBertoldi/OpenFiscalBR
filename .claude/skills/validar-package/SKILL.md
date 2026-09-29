---
name: validar-package
description: "Valida um package Go do OpenFiscalBR antes de dar o trabalho por concluido: gofmt (imune a falso positivo de CRLF no Windows), go build, go vet, go test, header de licenca LGPL em todo .go, module path nos imports, fronteira entre package e demo (packages/ nao expoe HTTP nem JSON) e integridade das contagens do arquivo gerado. Use quando o usuario pedir para validar, verificar, conferir ou checar a qualidade de um package, quando terminar de gerar ou editar codigo Go no projeto, ou antes de commitar."
argument-hint: "[<pkg>]"
---

# validar-package — portão de qualidade

Roda a bateria de verificação do projeto sobre um package (ou sobre tudo, se nenhum for
informado).

## Uso

```
/validar-package            # valida packages/ e demos/ inteiros
/validar-package sped       # valida apenas packages/sped/
```

`$1` = nome do package (opcional). Sem argumento, o alvo é `./...`.

Execute as seis verificações **na ordem**, e reporte todas as falhas encontradas ao final —
não pare na primeira. Erro de compilação (verificação 2) é a única exceção: se o pacote não
compila, as verificações 3 e 4 não têm como rodar, então pule-as e diga isso no relatório.

---

## 1 — gofmt

**Não use `gofmt -l`.** O repositório tem `core.autocrlf=true`; a árvore de trabalho fica com
CRLF, o `gofmt` normaliza para LF, e o resultado é que `gofmt -l` acusa **100% dos arquivos**.
Isso torna a verificação inútil e faz um problema real passar despercebido no meio do ruído.

Compare com o CR removido dos dois lados:

```bash
for f in $(git ls-files '*.go'); do
  if ! diff -q <(tr -d '\r' < "$f") <(gofmt < "$f" | tr -d '\r') >/dev/null 2>&1; then
    echo "DESFORMATADO: $f"
  fi
done
```

Para ver a diferença de um arquivo específico:

```bash
diff <(tr -d '\r' < <arquivo>) <(gofmt < <arquivo> | tr -d '\r')
```

O caso mais comum no projeto é desalinhamento de campos de struct — `gofmt` alinha os tipos
em coluna e um campo acrescentado depois costuma quebrar o alinhamento dos vizinhos.

Ao corrigir, rode `gofmt -w <arquivo>`: isso reescreve o arquivo em LF, o que é o
comportamento correto (o `.gitattributes`/`autocrlf` cuida da conversão no commit).

## 2 — Compilação

```bash
go build ./...                    # ou ./packages/<pkg>/...
```

Erros típicos e o que significam:

| Erro | Causa provável |
|---|---|
| `undefined: X` | dependência ainda não convertida, ou símbolo não exportado (minúsculo) |
| `cannot find module` | import usando caminho relativo em vez do module path |
| `imported and not used` | import deixado para trás após edição |

## 3 — Análise estática

```bash
go vet ./...                      # ou ./packages/<pkg>/...
```

## 4 — Testes

```bash
go test ./packages/... -v         # ou ./packages/<pkg>/... -v
```

Os testes devem rodar **sem dependência externa** — é regra do `CLAUDE.md`. Se algum teste
precisar de rede, banco ou arquivo fora do `t.TempDir()`, isso é um achado a reportar.

## 5 — Header de licença

Todo arquivo `.go` deve começar com o header LGPL definido no `CLAUDE.md`. Verificação rápida
dos que não têm:

```bash
for f in $(git ls-files '*.go'); do
  head -1 "$f" | grep -q "OpenFiscalBR - Automacao Fiscal Brasileira em Go" || echo "SEM HEADER: $f"
done
```

O header é obrigatório porque o projeto é obra derivada do ACBr sob LGPL 2.1+ — a atribuição
em cada arquivo não é formalidade, é o que mantém a conformidade com a licença de origem.

## 6 — Module path nos imports

Imports internos devem usar `github.com/openfiscalbr/openfiscalbr/packages/<pkg>`, nunca
caminho relativo:

```bash
grep -rn '"\./\|"\.\./' --include='*.go' . || echo "OK: nenhum import relativo"
```

## 7 — Fronteira entre package e demo

`packages/` e biblioteca e **nao expoe URL**: nada de servidor HTTP, JSON ou template. Todo
transporte de EXPOSICAO vive em `demos/`.

Excecao unica: `packages/dfe/soap.go` importa `net/http` como **CLIENTE** dos web services da
SEFAZ — transmitir o documento e funcao fiscal (o TDFeWebService do ACBr), nao exposicao.
A regra continua valendo para servidor, handler e JSON, inclusive no `dfe`.

Arquivos `_test.go` tambem ficam fora do check: o `httptest` (que importa `net/http`) e a
forma correta de testar o cliente SOAP sem rede, e teste nao entra no binario do consumidor.

```bash
grep -rn '"net/http"\|"encoding/json"\|"html/template"\|^\s*"net"$' \
  --include='*.go' --exclude='*_test.go' packages/ | grep -v '^packages/dfe/soap.go:' \
  || echo "OK: nenhum transporte em packages/"
```

Qualquer outra ocorrencia e um achado: o consumidor da biblioteca (um ERP, um job, um CLI)
precisa gerar o documento sem subir servidor. Se um handler HTTP foi parar no package, mova-o
para `demos/<pkg>/handlers.go` -- a dependencia so pode apontar do demo para o package.
`http.HandleFunc`, `http.ListenAndServe` e `json.Marshal` sao achado em QUALQUER package,
`dfe` incluido.

## 8 — Integridade do arquivo gerado

Verificacoes de layout nao pegam **valor calculado**. Um registro pode ter a sequencia de
campos perfeita e declarar contagem errada -- foi o caso do `9990` do SPED, que contava uma
linha a menos porque nao somava a propria linha do `9999`.

Quando o package gera arquivo, gere um de verdade e confira:

- o registro de totalizacao declara exatamente o numero de linhas do arquivo
- o registro de totalizacao por bloco bate com as linhas daquele bloco
- todo tipo de registro presente no arquivo aparece no registro de contagem por tipo
- toda linha comeca e termina com o delimitador

Para o `sped`, isso esta coberto por `TestSaveFileTXT_ContagensDoBloco9` e
`TestPopulateRegistro9900_ContaRegistrosDeDados`.

## 9 — Writer orfao

Writer que ninguem chama nao quebra build, `vet` nem teste: o registro simplesmente nunca sai
no arquivo. No `sped` isso escondeu tres registros (`E112`, `E113`, `E115`) por toda a
conversao, e uma auditoria campo a campo de 270 registros passou por cima deles -- ela so olha
writers que existem, nao se alguem os aciona.

```bash
grep -ho "func (b \*Bloco[0-9A-Za-z]*) [wW]riteRegistro[0-9A-Z]*(" packages/sped/write_*.go \
  | sed 's/.*) //; s/(//' | sort -u > /tmp/decl.txt
for w in $(cat /tmp/decl.txt); do
  [ "$(grep -rho "\.$w(" packages/sped/*.go | wc -l)" -eq 0 ] && echo "NUNCA CHAMADO: $w"
done
```

Saida vazia e o esperado. Reporte qualquer achado como **Bloqueante**.

O caminho mais barato para o mesmo defeito e ler o `unusedfunc` dos diagnosticos do editor --
foi ele que apontou os tres. Nao descarte diagnostico informativo sem olhar.

---

## Formato do relatório

```
## Validação — <pkg ou "projeto inteiro">

| Verificação | Resultado |
|---|---|
| gofmt          | OK / N arquivos desformatados |
| go build       | OK / falhou |
| go vet         | OK / N avisos |
| go test        | OK (N testes) / N falhas |
| header LGPL    | OK / N arquivos sem header |
| module path    | OK / N imports relativos |
| fronteira demo | OK / N arquivos com transporte em packages/ |
| integridade    | OK / divergencia nas contagens |

### Achados
(um bloco por achado, com arquivo:linha e o que fazer)
```

## Regras

- **Reporte números, não impressões.** "14 arquivos desformatados" é acionável; "alguns
  arquivos precisam de ajuste" não é.
- **Não corrija sozinho o que for além de `gofmt -w`.** Falha de teste ou erro de compilação
  pode ter mais de uma correção possível — reporte e deixe a decisão com quem pediu, a menos
  que o usuário tenha pedido para corrigir.
- **Não altere um teste para fazê-lo passar.** Se o teste falha, ou o código está errado, ou a
  expectativa do teste está errada; decidir qual dos dois é uma questão de regra fiscal, não de
  conveniência. Na dúvida, pare e pergunte.
