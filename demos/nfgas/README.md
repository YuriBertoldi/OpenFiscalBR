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
| `GET /api/exemplos` | — | catálogo de XMLs de exemplo |
| `GET /api/exemplos/{id}` | — | gera o XML do cenário e devolve em JSON |
| `GET /api/exemplos/{id}/download` | — | o mesmo XML como arquivo `.xml` |

O corpo pode ser `{"conteudo": "..."}` em JSON **ou** o documento cru (facilita `curl`):

```bash
curl -s -X POST --data-binary @nota.xml http://localhost:8080/api/ler | jq .resumo
curl -s -X POST --data-binary @pasta_toda.xml http://localhost:8080/api/ler-lote | jq .
```

## XMLs de exemplo

Não precisa ter um XML à mão para experimentar a demo. Escolha um cenário no
seletor **Exemplo** e clique em **Gerar com dados fictícios** — o XML sai
completo (emitente, destinatário, itens, impostos, totais), vai para a caixa
de entrada, a aba adequada é selecionada e o download fica disponível.

| Cenário | O que demonstra |
|---|---|
| `transmissao` | `NFGas` avulsa, sem assinatura nem protocolo — o que se envia à SEFAZ |
| `autorizada` | `nfgasProc` com `protNFGas` (cStat 100) — o que se recebe e arquiva |
| `cancelamento` | `eventoNFGas` com `evCancNFGas`, amarrado à chave do cenário de transmissão |
| `cancelamento-proc` | `procEventoNFGas` — evento enviado + retorno homologado (cStat 135) |
| `completo` | todo grupo opcional do leiaute preenchido (ver ressalva abaixo) |
| `erro-leitura` | quebra em `LerXMLString` com `ErrAtributoVersaoAusente` (HTTP 422) |
| `erro-regras` | lê bem e reprova em `/api/validar`: regras 226, 227, 247 e CNPJ inválido |
| `multi-itens` | 5 itens com CST de ICMS diferentes e total coerente |
| `multi-cfop` | 4 itens com CFOP distintos (5253, 5257, 6253, 5949) |

Os dados são **fixos**: o mesmo cenário gera sempre o mesmo XML, o que permite
repetir um teste. Marque **Variar dados** para sortear número, série e data a
cada clique — útil para importar vários documentos sem colidir chave. Em
**Ajustar dados (opcional)** dá para sobrescrever CNPJ, UF, série, número,
ambiente e quantidade de itens; o que ficar em branco mantém o fictício, e a
chave de acesso é recalculada a partir do que você informou.

Pela API, os mesmos parâmetros vão na query string:

```bash
curl -s http://localhost:8080/api/exemplos | jq '.exemplos[].id'
curl -s http://localhost:8080/api/exemplos/completo | jq -r .xml > nfgas-completa.xml
curl -s 'http://localhost:8080/api/exemplos/transmissao?uf=MG&nnf=4321&variar=1' | jq -r .chave
curl -sOJ http://localhost:8080/api/exemplos/multi-cfop/download
```

> **Sobre os totais:** o ACBr não calcula `vNF` — o campo só é lido, escrito e
> copiado, nunca derivado (`ACBrNFGas.Classes.pas:1735`); preencher o grupo
> `total` é responsabilidade do emitente. Como estes exemplos existem para ser
> importados e conferidos, a demo compõe o `vNF` com as parcelas que acrescem
> ao valor da nota (ST, FCP, FCPST, taxa de regulação e o total dos itens
> agregadores), descontando a desoneração apenas dos itens com
> `indDeduzDeson = 1`. `vTotDFe` acompanha o `vNF`, como na fixture do package.

> **Ressalva do cenário `completo`:** nenhum documento carrega todas as tags
> que o gerador sabe emitir — os grupos `ICMS00`..`ICMS90` se excluem por CST,
> `cNIS` exclui `NB`, `codDebAuto` exclui `codBanco`+`codAgencia`. O cenário
> cobre tudo que pode conviver num documento só, e `exemplos_test.go` mantém
> a lista das exceções com o motivo de cada uma. O protocolo e a assinatura
> são fictícios: serve para **importar**, não para transmitir — para transmitir
> use `transmissao` e assine com certificado de verdade.

Há um documento de exemplo em `../../packages/nfgas/testdata/nfgas_completa.xml`.
