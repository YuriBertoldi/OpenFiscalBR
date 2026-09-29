# Mapeamento de componentes ACBr

Referência carregada sob demanda pela skill `/convert` (PASSO 1 e PASSO 2).

## Componente → diretório dos fontes Delphi

Caminhos relativos ao `DelphiSourcePath` informado (a pasta `Fontes` do ACBr).

| ComponentName | Diretório |
|---|---|
| `ACBrComum` | `<path>/ACBrComum/` |
| `ACBrSPEDFiscal` | `<path>/ACBrTXT/ACBrSPED/ACBrSPEDFiscal/` (mais `ACBrTXTClass.pas` em `<path>/ACBrTXT/`) |
| `PCNComum` | `<path>/PCNComum/` |
| `ACBrDFe` | `<path>/ACBrDFe/` (apenas arquivos da raiz, **não** as subpastas de componentes) |
| `ACBrNFe` | `<path>/ACBrDFe/ACBrNFe/` |
| `ACBrCTe` | `<path>/ACBrDFe/ACBrCTe/` |
| `ACBrMDFe` | `<path>/ACBrDFe/ACBrMDFe/` |
| `ACBrBPe` | `<path>/ACBrDFe/ACBrBPe/` |
| `ACBrNFSeX` | `<path>/ACBrDFe/ACBrNFSeX/` |
| `ACBrNFSe` | `<path>/ACBrDFe/ACBrNFSe/` |
| `ACBrNF3e` | `<path>/ACBrDFe/ACBrNF3e/` |
| `ACBrNFCom` | `<path>/ACBrDFe/ACBrNFCom/` |
| `ACBrNFGas` | `<path>/ACBrDFe/ACBrNFGas/` |
| `ACBrGNRE` | `<path>/ACBrDFe/ACBrGNRE/` |
| `ACBrReinf` | `<path>/ACBrDFe/ACBrReinf/` |
| `ACBreSocial` | `<path>/ACBrDFe/ACBreSocial/` |
| `ACBrBoleto` | `<path>/ACBrBoleto/` |
| `ACBrPIXCD` | `<path>/ACBrPIXCD/` |
| `ACBrSAT` | `<path>/ACBrSAT/` |
| `ACBrTEFD` | `<path>/ACBrTEFD/` |
| `ACBrSerial` | `<path>/ACBrSerial/` |
| `ACBrTCP` | `<path>/ACBrTCP/` |
| `ACBrTXT` | `<path>/ACBrTXT/` |
| `ACBrDiversos` | `<path>/ACBrDiversos/` |
| `ACBrPagFor` | `<path>/ACBrPagFor/` |
| `ACBrBaaS` | `<path>/ACBrBaaS/` |
| `ACBrOpenDelivery` | `<path>/ACBrOpenDelivery/` |

Componente fora da tabela: buscar com Glob `<path>/**/*<ComponentName>*/`. Se ainda assim não
encontrar, **perguntar o caminho exato ao usuário** — nunca chutar o diretório.

## Componente → package Go de destino

| ComponentName | Package Go | Diretório destino |
|---|---|---|
| `ACBrComum` | `comum` | `packages/comum/` |
| `ACBrSPEDFiscal` | `sped` | `packages/sped/` |
| `PCNComum` | `pcn` | `packages/pcn/` |
| `ACBrDFe` | `dfe` | `packages/dfe/` |
| `ACBrNFe` | `nfe` | `packages/nfe/` |
| `ACBrCTe` | `cte` | `packages/cte/` |
| `ACBrMDFe` | `mdfe` | `packages/mdfe/` |
| `ACBrBPe` | `bpe` | `packages/bpe/` |
| `ACBrNFSeX` | `nfsex` | `packages/nfsex/` |
| `ACBrNFSe` | `nfse` | `packages/nfse/` |
| `ACBrNF3e` | `nf3e` | `packages/nf3e/` |
| `ACBrNFCom` | `nfcom` | `packages/nfcom/` |
| `ACBrNFGas` | `nfgas` | `packages/nfgas/` |
| `ACBrBoleto` | `boleto` | `packages/boleto/` |
| `ACBrPIXCD` | `pixcd` | `packages/pixcd/` |
| `ACBrSAT` | `sat` | `packages/sat/` |
| `ACBrGNRE` | `gnre` | `packages/gnre/` |
| `ACBrReinf` | `reinf` | `packages/reinf/` |
| `ACBreSocial` | `esocial` | `packages/esocial/` |
| `ACBrTEFD` | `tefd` | `packages/tefd/` |
| `ACBrPagFor` | `pagfor` | `packages/pagfor/` |
| `ACBrBaaS` | `baas` | `packages/baas/` |

**Regra de fallback para o destino** (componente fora da tabela): package = nome do componente
sem o prefixo `ACBr`, em minúsculas (`ACBrNFGas` → `nfgas`). Componente que vive sob
`<path>/ACBrDFe/<Sub>/` é sempre Layer 3 com dependências `comum`, `pcn`, `dfe`. Registre a
linha nova nesta tabela ao converter — a regra existe para a primeira vez, não para dispensar
a tabela.

### Packages que não vêm de um componente (units compartilhadas)

| Package Go | Origem Delphi | Papel |
|---|---|---|
| `rtc` | `<path>/ACBrDFe/ACBrDFe.RTC.*.pas` | Reforma Tributária (IBS/CBS/IS) — units compartilhadas por NFe, CTe, MDFe, BPe, NF3e, NFCom e NFGas. Convertido como package próprio (Layer 1, depende de `pcn`) para os componentes não duplicarem a árvore. |

## Dependências obrigatórias

A ordem de conversão é imposta por estas dependências — nunca converter um componente cujas
dependências ainda não existam em `packages/`.

| Package | Depende de |
|---|---|
| `comum` | nenhuma |
| `pcn` | `comum` |
| `rtc` | `pcn` |
| `dfe` | `comum`, `pcn` |
| `nfe`, `cte`, `mdfe`, `bpe`, `nfsex`, `nfse`, `nf3e`, `nfcom`, `gnre`, `reinf`, `esocial` | `comum`, `pcn`, `dfe` |
| `nfgas` | `comum`, `pcn`, `rtc` (leitura); a fase de emissão passará a exigir `dfe` |
| `boleto` | `comum` |
| `pixcd` | `comum` |
| `sat` | `comum` |
| `tefd` | `comum` |
| `pagfor` | `comum` |
| `baas` | `comum` |

Dependência considerada satisfeita quando `packages/<dep>/` existe **e** contém arquivos `.go`.
Diretório vazio não conta.

## Arquivos Delphi a excluir do escaneamento

Estes nunca são convertidos — dependem de engines gráficas ou de infraestrutura de design-time
do Delphi, que não têm equivalente em Go:

- Pastas de report visual: `DANFE/`, `DACTE/`, `DAMDFE/`, `DANFSe/`, `DACE/`, `DANF3e/`, `DANFCom/`
- Pastas de engine de report: `Fast/`, `Fortes/`, `LazReport/`, `EscPos/`
- Arquivos `*Reg.pas` — registradores de componentes de design-time

Dentro dos arquivos convertidos, ignorar blocos `{$IFDEF}` que referenciam GUI (`FMX`, `VCL`,
`LCL`) ou plataformas específicas.
