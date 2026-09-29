# Package comum

Package base do OpenFiscalBR, equivalente ao `ACBrComum` do Projeto ACBr.
Fornece tipos fundamentais, utilitarios e o gerador de arquivos texto (TXTClass) usado por todos os demais packages.

## Componentes

### TXTClass

Gerador de arquivos texto pipe-delimited usado pelo SPED e outros componentes fiscais.

```go
txt := comum.NewTXTClass()
txt.NomeArquivo = "saida.txt"
txt.Delimitador = "|"

// Montar linha SPED
linha := txt.LFillStr("0000", 0, false, '0') +
    txt.LFillStr("016", 0, false, '0') +
    txt.LFillInt(3550308, 7, false, '0') +
    txt.LFillDate(time.Now(), "02012006", false) +
    txt.DFill(1234.56, 2, false)

txt.Add(linha, true)  // adiciona "|" final
txt.SaveToFile()
```

**Metodos de formatacao:**

| Metodo | Uso | Exemplo |
|--------|-----|---------|
| `LFillStr(value, size, nulo, char)` | Campo texto com pad a esquerda | `LFillStr("SP", 0, false, '0')` → `\|SP` |
| `LFillInt(value, size, nulo, char)` | Campo inteiro com zeros a esquerda | `LFillInt(42, 7, false, '0')` → `\|0000042` |
| `LFillFloat(value, size, dec, nulo, char, mask)` | Campo monetario | `LFillFloat(100.5, 0, 2, false, '0', "")` → `\|100,50` |
| `LFillDate(value, mask, nulo)` | Data no formato SPED (ddmmaaaa) | `LFillDate(dt, "02012006", false)` → `\|01092026` |
| `RFill(value, size, char)` | Campo texto com pad a direita | `RFill("AB", 5, ' ')` → `\|AB   ` |
| `DFill(value, decimal, nulo)` | Valor decimal simplificado | `DFill(100.5, 2, false)` → `\|100,50` |

### FormatFloatBR

Formata float64 com separador decimal virgula (convencao brasileira):

```go
comum.FormatFloatBR(1234.56, "0.00") // "1234,56"
```

### ACBrError

Tipo de erro padrao da biblioteca:

```go
err := comum.NewACBrError("mensagem de erro")
```

## Testes

```bash
go test ./packages/comum/... -v
```
