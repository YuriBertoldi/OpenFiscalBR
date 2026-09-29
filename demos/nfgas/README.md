# Demo NFGas — leitura, importação em lote e emissão

Servidor HTTP + frontend para exercitar o `packages/nfgas`: leitura de XML avulso, importação
em lote, validação (chave de acesso + regras de negócio), eventos, retorno de consulta, o
formato `.ini` do ACBr **e a emissão** — geração do XML, assinatura XMLDSig e transmissão à
SEFAZ (recepção síncrona, consulta, status e cancelamento).

## Rodando

```bash
# direto
go run ./demos/nfgas
# ou com Docker
cd demos/nfgas && docker compose up --build
```

Abra <http://localhost:8080>. A porta muda com `PORT`, o diretório do frontend com `STATIC_DIR`.

Para as rotas que assinam/transmitem, configure o certificado A1 e o contexto pelo ambiente:

| Variável | Uso |
|---|---|
| `CERT_PATH` / `CERT_PASS` | caminho e senha do `.pfx` |
| `UF` | UF autorizadora (default `SP`; MA e PA não têm web service — lacuna do próprio ACBr) |
| `AMBIENTE` | `homologacao` (default) ou `producao` |

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
| `POST /api/gerar` | XML ou `.ini` da nota | gera o XML (chave, cDV e cNF calculados) + URL do QR-Code |
| `POST /api/assinar` | XML, ou `.ini` para gerar antes | assina `infNFGas` com o certificado do ambiente |
| `POST /api/transmitir` | XML ou `.ini` | gera, assina e envia à SEFAZ (recepção síncrona de 1 nota); devolve o retorno + XML assinado |
| `GET /api/status-sefaz` | — | `consStatServNFGas` no web service configurado |
| `POST /api/consultar-sefaz` | `{"chave": "..."}` | `consSitNFGas` |
| `POST /api/cancelar` | `{"chave","protocolo","justificativa"}` | evento de cancelamento assinado |

O corpo pode ser `{"conteudo": "..."}` em JSON **ou** o documento cru (facilita `curl`):

```bash
curl -s -X POST --data-binary @nota.xml http://localhost:8080/api/ler | jq .resumo
curl -s -X POST --data-binary @pasta_toda.xml http://localhost:8080/api/ler-lote | jq .
```

Há um documento de exemplo em `../../packages/nfgas/testdata/nfgas_completa.xml`.
