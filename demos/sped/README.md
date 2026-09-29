# OpenFiscalBR - Demo SPED Fiscal

Demo interativa para geracao de arquivos SPED Fiscal (EFD-ICMS/IPI).

## Executar localmente

```bash
cd demos/sped
go run .
```

Acesse http://localhost:8080

## Executar com Docker

```bash
# Na raiz do projeto
docker compose -f demos/sped/docker-compose.yml up --build
```

Acesse http://localhost:8080

## Endpoints da API

| Metodo | Rota          | Descricao                        |
|--------|---------------|----------------------------------|
| GET    | /api/status   | Status do servico                |
| POST   | /api/gerar    | Gera arquivo SPED Fiscal (TXT)   |

### POST /api/gerar

Corpo JSON com campos do Registro 0000 e opcionalmente do Registro 0005.

Exemplo:
```json
{
  "cod_ver": 15,
  "cod_fin": 0,
  "dt_ini": "2026-09-01",
  "dt_fin": "2026-09-30",
  "nome": "EMPRESA TESTE LTDA",
  "cnpj": "12345678000199",
  "uf": "SP",
  "ie": "123456789",
  "cod_mun": 3550308,
  "ind_perfil": 0,
  "ind_ativ": 1
}
```

Resposta:
```json
{
  "sucesso": true,
  "mensagem": "Arquivo SPED Fiscal gerado com sucesso",
  "arquivo": "SPED_12345678000199_092026.txt",
  "conteudo": "|0000|015|0|01092026|..."
}
```
