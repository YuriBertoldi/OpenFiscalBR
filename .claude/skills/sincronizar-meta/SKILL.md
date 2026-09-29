---
name: sincronizar-meta
description: "Sincroniza o .openfiscalbr-meta.json e os READMEs do OpenFiscalBR com o estado real do codigo: registra os hashes SHA256 dos fontes Delphi de cada componente convertido, confere a tabela de status dos packages no README da raiz contra o que existe de fato, e verifica se cada package tem README.md e teste. Use quando o usuario pedir para atualizar os metadados, sincronizar o status dos packages, conferir se o README esta correto, ou apos concluir uma conversao."
---

# sincronizar-meta

Põe `.openfiscalbr-meta.json` e os READMEs de acordo com o que existe de fato no repositório.

## Uso

```
/sincronizar-meta
```

Sem argumentos — sempre varre o projeto inteiro.

**Estado inicial conhecido:** `.openfiscalbr-meta.json` tem `"components": {}`. Nem `comum`
nem `sped` foram registrados, embora ambos existam. Enquanto isso não for corrigido, o modo de
atualização da `/convert` (PASSO 7) não tem hash de referência para comparar e sempre trata
tudo como conversão nova.

---

## Etapa 1 — Inventariar o que existe

1. Liste `packages/*/` e, para cada um, os arquivos `.go`.
2. Liste `demos/*/`.
3. Leia `.openfiscalbr-meta.json` e compare com o inventário: packages presentes no disco e
   ausentes do meta, e o inverso.

## Etapa 2 — Registrar os hashes

Para cada package presente no disco mas ausente do meta, é preciso saber de qual componente
ACBr ele veio e onde estão os fontes Delphi. O mapeamento componente → package está em
`../convert/referencias/componentes.md`.

Se o caminho dos fontes Delphi não estiver no meta nem tiver sido informado, **pergunte**. Sem
ele não há como calcular hash, e registrar a entrada sem `delphiFileHashes` dá a falsa
impressão de que o modo de atualização vai funcionar.

Calcule o SHA256 de cada `.pas` do componente (aplicando as exclusões documentadas em
`../convert/referencias/componentes.md`) e grave:

```json
{
  "components": {
    "ACBrComum": {
      "goPackage": "comum",
      "delphiSourcePath": "<caminho>",
      "lastConvertedAt": "<ISO 8601>",
      "delphiFileHashes": { "ACBrBase.pas": "sha256:<hash>" },
      "goFiles": ["base.go", "doc.go", "errors.go", "txt_class.go", "util.go"]
    }
  }
}
```

Preserve as chaves de topo já existentes (`projectName`, `license`, `derivedFrom`).

## Etapa 3 — Conferir o status no README da raiz

Compare a tabela "Status dos Packages" do `README.md` com a realidade. Um package só é
**Completo** quando todo tipo declarado está de fato exercitado pelo código — não basta o
diretório existir.

Para o `packages/sped`, o critério concreto é a razão entre structs de registro declarados e
writers implementados:

```bash
grep -c "^type Registro.* struct" packages/sped/bloco_*.go | awk -F: '{s+=$2} END {print "structs:", s}'
grep -ch "riteRegistro.*(" packages/sped/write_*.go | awk '{s+=$1} END {print "writers:", s}'
```

Um struct sem writer é um registro que o consumidor consegue preencher e que **nunca aparece
no arquivo gerado** — silenciosamente. Por isso a divergência importa mais do que uma
imprecisão de documentação.

Estado conhecido na última verificação: `packages/sped` está marcado "Completo" no README da
raiz, mas tem cerca de 198 structs de registro para cerca de 59 writers. O status honesto é
**Parcial**, com a proporção explicitada.

Reporte a divergência e proponha a correção antes de editar — o usuário pode preferir
completar o código a rebaixar o status.

## Etapa 4 — Conferir a estrutura obrigatória de cada package

O `CLAUDE.md` exige, para todo package convertido:

- `README.md` com descrição, componente Delphi correspondente, tipos principais, tabela de
  registros/tipos implementados e como rodar os testes
- `<pkg>_test.go` com testes reais, sem dependência externa
- header de licença LGPL em todo `.go`

```bash
for d in packages/*/; do
  p=$(basename "$d")
  [ -f "$d/README.md" ] || echo "SEM README: $p"
  ls "$d"/*_test.go >/dev/null 2>&1 || echo "SEM TESTE: $p"
done
```

Para o header de licença e o restante da higiene, use `/validar-package` — não duplique a
verificação aqui.

---

## Formato do relatório

```
## Sincronização de metadados

### .openfiscalbr-meta.json
- Componentes registrados: N (antes: M)
- Sem caminho de fontes Delphi conhecido: <lista>

### Status dos packages (README da raiz)
| Package | README diz | Realidade | Ação |
|---|---|---|---|

### Estrutura obrigatória
(pendências por package, ou "tudo conforme")
```

## Regras

- **Proponha antes de editar o README.** Rebaixar um status de "Completo" para "Parcial" é uma
  afirmação sobre a maturidade do projeto — quem decide é o usuário.
- **Nunca registre um componente no meta sem os hashes.** Uma entrada sem `delphiFileHashes`
  faz a `/convert` acreditar que tem base de comparação quando não tem, e o modo de
  atualização passa a reportar "nenhuma alteração" para qualquer coisa.
- **Não invente o `delphiSourcePath`.** Se não souber de onde veio o package, pergunte; um
  caminho errado gera hashes de fontes que não são os de origem, o que é pior do que não ter
  hash nenhum.
