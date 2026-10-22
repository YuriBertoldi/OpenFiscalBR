# Demo — NF-e / NFC-e

API REST e página web para experimentar o `packages/nfe`: leitura, validação, conversão entre
XML e `.ini`, e emissão com transmissão à SEFAZ.

A demo é **descartável e serve para demonstrar**; o package é o produto. Todo o transporte HTTP
vive aqui — `packages/nfe` não expõe URL, não serve HTTP e não fala JSON.

## Rodar

```bash
# Direto, da raiz do repositório
go run ./demos/nfe

# Ou com Docker
cd demos/nfe && docker compose up --build
```

Abre em http://localhost:8080.

Sem certificado configurado, tudo que **não** fala com a SEFAZ funciona: ler, validar,
converter entre os dois formatos e gerar XML. As rotas de transmissão respondem `412` dizendo o
que falta configurar.

## Configuração

Tudo por variável de ambiente, com default seguro:

| variável | default | para quê |
|---|---|---|
| `PORT` | `8080` | porta do servidor |
| `STATIC_DIR` | `frontend` | pasta do frontend |
| `UF` | `SP` | UF do emitente — decide o endereço do web service |
| `MODELO` | `55` | `55` (NF-e) ou `65` (NFC-e) |
| `AMBIENTE` | `homologacao` | `producao` ou `homologacao` |
| `CERT_PATH` | — | caminho do `.pfx` (certificado A1) |
| `CERT_PASS` | — | senha do `.pfx` |
| `CSC` / `IDCSC` | — | Código de Segurança do Contribuinte; só a NFC-e usa |
| `TIMEOUT_SEFAZ` | `90` | segundos de espera pela SEFAZ |

> **O ambiente padrão é homologação, e é preciso dizer `producao` explicitamente.** As rotas de
> transmissão enviam documento de verdade: `/api/transmitir`, `/api/cancelar`,
> `/api/carta-correcao` e `/api/inutilizar` produzem efeito fiscal real em produção.

## Endpoints

### Leitura e conversão — sem certificado, sem rede

| rota | o que faz |
|---|---|
| `POST /api/ler` | lê uma NF-e em XML **ou** `.ini` e devolve a nota estruturada |
| `POST /api/ler-lote` | lê vários documentos de uma vez; um ilegível não derruba os demais |
| `POST /api/validar` | confere a chave de acesso **e** as regras de negócio |
| `POST /api/ini-para-xml` | converte o `.ini` do ACBr em XML |
| `POST /api/xml-para-ini` | converte o XML em `.ini` |
| `POST /api/ler-evento` | lê o retorno de um lote de eventos (`retEnvEvento`) |
| `POST /api/ler-distribuicao` | lê um `retDistDFeInt`, descomprimindo cada `docZip` |
| `GET /api/status` | status da API e catálogo de rotas |

### Transmissão — exigem certificado

| rota | o que faz |
|---|---|
| `POST /api/gerar` | gera o XML a partir de XML ou `.ini` |
| `POST /api/assinar` | gera, assina e — na NFC-e — calcula o QR-Code |
| `POST /api/transmitir` | envia a nota à SEFAZ |
| `GET /api/status-sefaz` | consulta o status do serviço |
| `GET /api/consultar/{chave}` | consulta uma nota pela chave |
| `POST /api/cancelar` | cancela uma nota autorizada |
| `POST /api/carta-correcao` | envia carta de correção |
| `POST /api/inutilizar` | inutiliza faixa de numeração |
| `POST /api/distribuicao` | consulta os documentos emitidos **contra** o contribuinte |

### Exemplos

| rota | o que faz |
|---|---|
| `GET /api/exemplos` | lista os exemplos do catálogo |
| `GET /api/exemplos/{id}` | devolve o conteúdo de um exemplo |

Os exemplos são **montados em Go e serializados na hora**, não guardados como texto fixo: assim
não envelhecem em relação ao package — se um campo mudar de nome, a demo deixa de compilar em
vez de passar a gerar documento inválido em silêncio.

**Todos os dados são fictícios.** O CNPJ é o que o próprio ACBr usa nos exemplos dele, o
ambiente sai sempre como homologação, e há teste travando as duas coisas.

## Exemplos de uso

Ler uma nota (aceita JSON ou o corpo cru, para facilitar o `curl`):

```bash
curl -X POST localhost:8080/api/ler --data-binary @nota.xml
curl -X POST localhost:8080/api/ler -d '{"conteudo":"<nfeProc>...</nfeProc>"}'
```

```json
{
  "resumo": {
    "chave": "35260399999999000191550010000010011123456780",
    "modelo": 55, "serie": 1, "nNF": 1001,
    "emitente": "EMPRESA EXEMPLO LTDA",
    "destinatario": "CLIENTE EXEMPLO SA",
    "uf": "SP", "vNF": 2060, "itens": 2
  },
  "nfe": { "InfNFe": { "...": "" } }
}
```

Validar:

```bash
curl -X POST localhost:8080/api/validar --data-binary @nota.xml
```

```json
{
  "chave": "35260399999999000191550010000010011123456780",
  "chaveValida": true,
  "rejeicoes": [
    { "codigo": 610, "mensagem": "Total da NF difere do somatorio dos itens", "item": 0 }
  ],
  "valida": false
}
```

> A chave e as regras de negócio são validadas **separadamente** porque respondem perguntas
> diferentes: a chave confere a coerência interna do documento, as regras conferem o conteúdo
> contra a legislação. Juntá-las numa resposta só faria o consumidor não saber o que corrigir.

Converter `.ini` em XML — o caminho que a maioria dos ERP usa:

```bash
curl -X POST localhost:8080/api/ini-para-xml --data-binary @nota.ini
```

```json
{ "chave": "35260399999999000191550010000010011123456780", "xml": "<NFe>...</NFe>" }
```

Cancelar (o protocolo é obrigatório — o cancelamento referencia a **autorização**, não a nota):

```bash
curl -X POST localhost:8080/api/cancelar -d '{
  "chave": "35260399999999000191550010000010011123456780",
  "protocolo": "135260000000001",
  "justificativa": "Erro de digitacao no valor dos produtos",
  "idLote": 1
}'
```

Distribuição DF-e, varrendo o lote inteiro:

```bash
curl -X POST localhost:8080/api/distribuicao -d '{
  "cUFAutor": 35, "cnpjCpf": "99999999000191", "ultNSU": "0", "varrer": true
}'
```

> A RFB devolve no máximo **cinquenta** documentos por chamada, então `varrer: false` quase
> nunca traz tudo. O critério de parada é `ultNSU == maxNSU`, e não "o lote veio vazio".

## Testes

```bash
go test ./demos/nfe/
```

Os testes travam o **contrato entre a demo e o package**: cada exemplo é lido de volta pelo
próprio `packages/nfe`, a chave de acesso é conferida, e a lista de rotas anunciada em
`/api/status` é comparada com o que o `main.go` registra de fato — lendo o fonte, para não
haver uma segunda lista a esquecer de atualizar.

Um detalhe que o teste da chave documenta: o `.ini` **não** restaura o `infNFe/ID`. É uma das
oito chaves que o ACBr grava e não lê — a chave é derivada dos campos, não transportada. O
teste remonta a chave a partir do que foi lido e confere que bate com a que o exemplo gerou;
divergir significaria que algum dos oito campos que a compõem se perdeu na ida e volta.
