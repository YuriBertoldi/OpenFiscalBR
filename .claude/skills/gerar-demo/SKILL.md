---
name: gerar-demo
description: "Gera a demo de um package do OpenFiscalBR em demos/<pkg>/: servidor HTTP com net/http, handlers, SPA frontend (HTML/JS/CSS puro, sem framework), Dockerfile multi-stage, docker-compose e README. Traz os conjuntos de rotas por tipo de componente (DFe, Boleto, PIXCD, SAT). Use quando o usuario pedir para criar, gerar ou montar uma demo, um servidor de exemplo ou uma API REST de demonstracao para um package do projeto."
argument-hint: "<pkg>"
---

# gerar-demo

Gera a demo de um package em `demos/<pkg>/`, seguindo a Estrutura Padrão de Demo do
`CLAUDE.md`.

## Uso

```
/gerar-demo nfe
/gerar-demo boleto
```

`$1` = nome do package (o mesmo de `packages/<pkg>/`). Se vier vazio, pergunte.

Pré-requisito: `packages/<pkg>/` precisa existir e compilar. Gerar a demo de um package
inexistente produz handlers que não compilam. Se não existir, ofereça rodar `/convert` antes.

---

## Estrutura a gerar

```
demos/<pkg>/
  main.go
  handlers.go
  frontend/index.html
  frontend/app.js
  frontend/style.css
  Dockerfile
  docker-compose.yml
  README.md
```

Use `demos/sped/` como molde de referência — é a demo existente e já segue o padrão.

## main.go

Todo arquivo `.go` leva o header LGPL (ver `CLAUDE.md`). O molde, extraído de
`demos/sped/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("POST /api/gerar", handleGerar)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Serve frontend static files
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenFiscalBR <PKG> Demo rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
```

Pontos do padrão que não devem mudar: `net/http` puro (sem router de terceiro), method-routing
do `ServeMux` do Go 1.22 (`"POST /api/gerar"`), `PORT` e `STATIC_DIR` por variável de ambiente
com default. O projeto não tem dependência externa alguma — introduzir uma só para a demo é
desproporcional.

## Rotas por tipo de componente

**DFe (NFe, CTe, MDFe, BPe, NFSeX, NF3e, NFCom):**
```
POST /api/criar            criar documento
POST /api/validar          validar XML
POST /api/assinar          assinar digitalmente
POST /api/transmitir       transmitir para a SEFAZ
GET  /api/status           status do servico da SEFAZ
POST /api/cancelar         cancelar documento
POST /api/carta-correcao   carta de correcao
POST /api/inutilizar       inutilizar numeracao
GET  /api/consultar/{chave} consultar por chave
```

**Boleto:**
```
POST /api/boleto           gerar boleto
GET  /api/boleto/{id}      consultar boleto
POST /api/remessa          gerar arquivo remessa
POST /api/retorno          processar arquivo retorno
GET  /api/bancos           listar bancos suportados
```

**PIXCD:**
```
POST  /api/cob             criar cobranca
GET   /api/cob/{txid}      consultar cobranca
PATCH /api/cob/{txid}      alterar cobranca
GET   /api/pix/{e2eid}     consultar PIX por e2eid
POST  /api/webhook         configurar webhook
POST  /api/qrcode          gerar QR Code
```

**SAT:**
```
POST /api/venda            enviar venda
POST /api/cancelamento     cancelar venda
GET  /api/status           status do SAT
GET  /api/info             informacoes do SAT
```

**Geradores de arquivo (SPED, PagFor):**
```
POST /api/gerar            gerar o arquivo
GET  /api/status           status do servico
```

Parâmetro de rota do `ServeMux` do Go 1.22 usa chaves e é lido com `r.PathValue("chave")` —
não use o estilo `:chave`, que o `ServeMux` não interpreta.

## handlers.go

Um handler por rota. Cada um monta o componente a partir do JSON do request, chama o método do
package e devolve JSON. Siga `demos/sped/handlers.go`, que já traz o helper `writeError` —
reaproveite-o em vez de criar um segundo formato de erro.

Toda resposta de erro deve sair como JSON com o status HTTP correto; o frontend faz
pretty-print da resposta e um erro em texto puro quebra a exibição.

## frontend/

SPA em HTML, JS e CSS puros, sem framework nem CDN:

- `index.html` — um formulário por endpoint, área de resposta com `<pre>` para o JSON, e um
  indicador de status
- `app.js` — `fetch()` por endpoint e pretty-print via `JSON.stringify(x, null, 2)`
- `style.css` — CSS puro

`<html lang="pt-BR">` e `<meta charset="UTF-8">`.

## Dockerfile e docker-compose.yml

Os moldes estão em `../convert/referencias/geracao-go.md`. Atenção ao `COPY go.mod ./`:
**não inclua `go.sum`** — o módulo não tem dependência externa e o arquivo não existe, então
`COPY go.mod go.sum ./` quebra o build.

As variáveis de certificado no compose só fazem sentido para componentes que assinam documento
(DFe, SAT). Para os demais, omita.

## README.md

Como buildar, como rodar (`go run .` e Docker), a porta, e a tabela de endpoints com exemplo de
request e response em JSON para cada um.

---

## Regras

- **Não invente método do package.** Os handlers só podem chamar o que existe em
  `packages/<pkg>/`. Se uma rota do conjunto acima não tiver método correspondente, omita a
  rota e diga isso no README em vez de gerar um handler que não compila.
- **Sem dependência externa.** O `go.mod` do projeto está em stdlib pura; acrescentar um router
  ou um framework de frontend só pela demo muda o perfil de dependência do repositório inteiro.
- **Valide ao final** com `/validar-package` e confirme que `go build ./demos/<pkg>/...` passa.
