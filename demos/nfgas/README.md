# Demo NFGas — leitura e importação em lote

Servidor HTTP + frontend para exercitar o `packages/nfgas`: leitura de XML avulso, importação
em lote, validação (chave de acesso + regras de negócio), eventos, retorno de consulta e o
formato `.ini` do ACBr.

**Escopo desta fase: somente leitura.** As rotas de emissão (`/api/assinar`, `/api/transmitir`,
`/api/cancelar`, `/api/inutilizar`) **não existem** porque o package ainda não tem método por
trás delas — a emissão do documento fica para uma fase futura, e os métodos correspondentes no
package devolvem `ErrNaoImplementado`.

## Rodando

```bash
# direto
go run ./demos/nfgas
# ou com Docker
cd demos/nfgas && docker compose up --build
```

Abra <http://localhost:8080>. A porta muda com `PORT`, o diretório do frontend com `STATIC_DIR`.

## Rotas

| Rota | Corpo | O que faz |
|---|---|---|
| `POST /api/ler` | XML de uma `NFGas` ou `nfgasProc` | devolve resumo + árvore completa em JSON |
| `POST /api/ler-lote` | XMLs concatenados (misturando avulsas e processadas) | devolve resumo por nota; nota torta não derruba o lote |
| `POST /api/validar` | XML | dígito verificador, concatenação da chave (regra 227) e regras de negócio 226/247/252 |
| `POST /api/ler-evento` | `procEventoNFGas`, `retEventoNFGas` ou `eventoNFGas` | tipo, cStat, protocolo e justificativa do cancelamento |
| `POST /api/ler-consulta` | `retConsSitNFGas` | situação, protocolo e eventos vinculados |
| `POST /api/ler-ini` | `.ini` no formato ACBr | devolve o documento em JSON |
| `GET /api/status` | — | ping com modelo, leiaute e rotas |

O corpo pode ser `{"conteudo": "..."}` em JSON **ou** o documento cru (facilita `curl`):

```bash
curl -s -X POST --data-binary @nota.xml http://localhost:8080/api/ler | jq .resumo
curl -s -X POST --data-binary @pasta_toda.xml http://localhost:8080/api/ler-lote | jq .
```

Há um documento de exemplo em `../../packages/nfgas/testdata/nfgas_completa.xml`.
