# Mapeamento de campos â€” ACBr Delphi â†’ Go

Tabela gerada, **nao editar a mao**. Para regerar um registro:

```bash
py ferramentas/comparar-campos.py <bloco> <registro> --doc
```

Cada linha e uma posicao da linha do arquivo SPED: a coluna `ACBr` traz o argumento passado ao
`LFill`/`DFill` no `.pas` de origem, e a coluna `Go` o argumento passado ao `LFillStr`/`DFill`
correspondente. Serve para conferir o port sem reabrir o Delphi, e para que uma renomeacao de
campo apareca como diff.

Como ler as diferencas que aparecem aqui:

- **Nome proprio de um lado so** (`ChaveEletronicaCTe` no `D100`, `strIND_FRT`) â€” o ACBr montou
  o valor numa variavel local antes de concatenar, em vez de usar o campo direto. Nao e
  divergencia; a equivalencia esta registrada em `ALIAS_GO`, no topo do comparador.
- **Marcado com "(ramos por versao)"** â€” o writer do ACBr tem mais de um `Add`, um por versao
  de leiaute, e o extrator le so o primeiro. O lado Go fatora o trecho comum e ramifica so a
  diferenca, entao aparece mais longo. Sao 14 registros: `0220`, `1310`, `C170`, `C177`,
  `C405`, `C460`, `D510`, `D700`, `E113`, `E116`, `E240`, `E250`, `E310` e `G110`.
- **`D750`** nao aparece: nao tem writer no Go de proposito, porque no ACBr ele monta a linha e
  nunca a grava. Veja a tabela de defeitos reproduzidos no `README.md`.

Nos 255 registros restantes a quantidade e o nome batem posicao a posicao.

## Bloco 0

### 0000
| # | ACBr | Go |
|---|------|-----|
| 0 | 0000 | 0000 |
| 1 | CodVerToStr(COD_VER) | CodVer.String() |
| 2 | Integer(COD_FIN) | CodFin |
| 3 | DT_INI | DtIni |
| 4 | DT_FIN | DtFin |
| 5 | NOME | Nome |
| 6 | CNPJ | CNPJ |
| 7 | CPF | CPF |
| 8 | UF | UF |
| 9 | IE | IE |
| 10 | COD_MUN | CodMun |
| 11 | IM | IM |
| 12 | SUFRAMA | Suframa |
| 13 | strIND_PERFIL | IndPerfil.String() |
| 14 | Integer(IND_ATIV) | IndAtiv |

### 0001
| # | ACBr | Go |
|---|------|-----|
| 0 | 0001 | 0001 |
| 1 | Integer(FRegistro0001.IND_MOV) | Registro0001.IndMov |

### 0005
| # | ACBr | Go |
|---|------|-----|
| 0 | 0005 | 0005 |
| 1 | FANTASIA | Fantasia |
| 2 | CEP | CEP |
| 3 | ENDERECO | Endereco |
| 4 | NUM | Num |
| 5 | COMPL | Compl |
| 6 | BAIRRO | Bairro |
| 7 | FONE | Fone |
| 8 | FAX | Fax |
| 9 | EMAIL | Email |

### 0015
| # | ACBr | Go |
|---|------|-----|
| 0 | 0015 | 0015 |
| 1 | UF_ST | UfST |
| 2 | IE_ST | IeST |

### 0100
| # | ACBr | Go |
|---|------|-----|
| 0 | 0100 | 0100 |
| 1 | NOME | Nome |
| 2 | CPF | CPF |
| 3 | CRC | CRC |
| 4 | CNPJ | CNPJ |
| 5 | CEP | CEP |
| 6 | ENDERECO | Endereco |
| 7 | NUM | Num |
| 8 | COMPL | Compl |
| 9 | BAIRRO | Bairro |
| 10 | FONE | Fone |
| 11 | FAX | Fax |
| 12 | EMAIL | Email |
| 13 | COD_MUN | CodMun |

### 0150
| # | ACBr | Go |
|---|------|-----|
| 0 | 0150 | 0150 |
| 1 | COD_PART | CodPart |
| 2 | NOME | Nome |
| 3 | COD_PAIS | CodPais |
| 4 | CNPJ | CNPJ |
| 5 | CPF | CPF |
| 6 | IE | IE |
| 7 | 9999999 | 9999999 |
| 8 | COD_MUN | CodMun |
| 9 | SUFRAMA | Suframa |
| 10 | ENDERECO | Endereco |
| 11 | NUM | Num |
| 12 | COMPL | Compl |
| 13 | BAIRRO | Bairro |

### 0175
| # | ACBr | Go |
|---|------|-----|
| 0 | 0175 | 0175 |
| 1 | DT_ALT | DtAlt |
| 2 | NR_CAMPO | NrCampo |
| 3 | CONT_ANT | ContAnt |

### 0190
| # | ACBr | Go |
|---|------|-----|
| 0 | 0190 | 0190 |
| 1 | UNID | Unid |
| 2 | DESCR | Descr |

### 0200
| # | ACBr | Go |
|---|------|-----|
| 0 | 0200 | 0200 |
| 1 | COD_ITEM | CodItem |
| 2 | DESCR_ITEM | DescrItem |
| 3 | COD_BARRA | CodBarra |
| 4 | COD_ANT_ITEM | CodAntItem |
| 5 | UNID_INV | UnidInv |
| 6 | strTIPO_ITEM | TipoItem.String() |
| 7 | COD_NCM | CodNCM |
| 8 | EX_IPI | ExIPI |
| 9 | COD_GEN | CodGen |
| 10 | COD_LST | CodLst |
| 11 | ALIQ_ICMS | AliqICMS |
| 12 | CEST | CEST |

### 0205
| # | ACBr | Go |
|---|------|-----|
| 0 | 0205 | 0205 |
| 1 | DESCR_ANT_ITEM | DescrAntItem |
| 2 | DT_INI | DtIni |
| 3 | DT_FIN | DtFin |
| 4 | COD_ANT_ITEM | CodAntItem |

### 0206
| # | ACBr | Go |
|---|------|-----|
| 0 | 0206 | 0206 |
| 1 | COD_COMB | CodComb |

### 0210
| # | ACBr | Go |
|---|------|-----|
| 0 | 0210 | 0210 |
| 1 | COD_ITEM_COMP | CodItemComp |
| 2 | QTD_COMP | QtdComp |
| 3 | PERDA | Perda |

### 0220 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | 0220 | 0220 |
| 1 | UmReg0220.UNID_CONV | UnidConv |
| 2 | UmReg0220.FAT_CONV | FatConv |
| 3 | — | CodBarra |

### 0221
| # | ACBr | Go |
|---|------|-----|
| 0 | 0221 | 0221 |
| 1 | wReg0221.COD_ITEM_ATOMICO | CodItemAtomico |
| 2 | wReg0221.QTDE_CONTIDA | QtdeContida |

### 0300
| # | ACBr | Go |
|---|------|-----|
| 0 | 0300 | 0300 |
| 1 | COD_IND_BEM | CodIndBem |
| 2 | IDENT_MERC | IdentMerc |
| 3 | DESCR_ITEM | DescrItem |
| 4 | COD_PRNC | CodPrnc |
| 5 | COD_CTA | CodCta |
| 6 | NR_PARC | NrParc |

### 0305
| # | ACBr | Go |
|---|------|-----|
| 0 | 0305 | 0305 |
| 1 | COD_CCUS | CodCcus |
| 2 | FUNC | Func |
| 3 | VIDA_UTIL | VidaUtil |

### 0400
| # | ACBr | Go |
|---|------|-----|
| 0 | 0400 | 0400 |
| 1 | COD_NAT | CodNat |
| 2 | DESCR_NAT | DescrNat |

### 0450
| # | ACBr | Go |
|---|------|-----|
| 0 | 0450 | 0450 |
| 1 | COD_INF | CodInf |
| 2 | TXT | Txt |

### 0460
| # | ACBr | Go |
|---|------|-----|
| 0 | 0460 | 0460 |
| 1 | COD_OBS | CodObs |
| 2 | TXT | Txt |

### 0500
| # | ACBr | Go |
|---|------|-----|
| 0 | 0500 | 0500 |
| 1 | DT_ALT | DtAlt |
| 2 | COD_NAT_CC | CodNatCC |
| 3 | IND_CTA | IndCta |
| 4 | NIVEL | Nivel |
| 5 | COD_CTA | CodCta |
| 6 | NOME_CTA | NomeCta |

### 0600
| # | ACBr | Go |
|---|------|-----|
| 0 | 0600 | 0600 |
| 1 | DT_ALT | DtAlt |
| 2 | COD_CCUS | CodCcus |
| 3 | CCUS | Ccus |

### 0990
| # | ACBr | Go |
|---|------|-----|
| 0 | 0990 | 0990 |
| 1 | QTD_LIN_0 | Registro0990.QtdLin0 |

## Bloco B

### B001
| # | ACBr | Go |
|---|------|-----|
| 0 | B001 | B001 |
| 1 | Integer(IND_MOV) | RegistroB001.IndMov |

### B020
| # | ACBr | Go |
|---|------|-----|
| 0 | B020 | B020 |
| 1 | Integer(IND_OPER) | IndOper |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | CodSitToStr(COD_SIT) | CodSit.String() |
| 6 | SER | Ser |
| 7 | NUM_DOC | NumDoc |
| 8 | CHV_NFE | ChvNFe |
| 9 | DT_DOC | DtDoc |
| 10 | COD_MUN_SERV | CodMunServ |
| 11 | VL_CONT | VlCont |
| 12 | VL_MAT_TERC | VlMatTerc |
| 13 | VL_SUB | VlSub |
| 14 | VL_ISNT_ISS | VlIsntIss |
| 15 | VL_DED_BC | VlDedBC |
| 16 | VL_BC_ISS | VlBcIss |
| 17 | VL_BC_ISS_RT | VlBcIssRT |
| 18 | VL_ISS_RT | VlIssRT |
| 19 | VL_ISS | VlIss |
| 20 | COD_INF_OBS | CodInfObs |

### B025
| # | ACBr | Go |
|---|------|-----|
| 0 | B025 | B025 |
| 1 | VL_CONT_P | VlContP |
| 2 | VL_BC_ISS_P | VlBcIssP |
| 3 | ALIQ_ISS | AliqIss |
| 4 | VL_ISS_P | VlIssP |
| 5 | VL_ISNT_ISS_P | VlIsntIssP |
| 6 | COD_SERV | CodServ |

### B030
| # | ACBr | Go |
|---|------|-----|
| 0 | B030 | B030 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | NUM_DOC_INI | NumDocIni |
| 4 | NUM_DOC_FIN | NumDocFin |
| 5 | DT_DOC | DtDoc |
| 6 | QTD_CANC | QtdCanc |
| 7 | VL_CONT | VlCont |
| 8 | VL_ISNT_ISS | VlIsntIss |
| 9 | VL_BC_ISS | VlBcIss |
| 10 | VL_ISS | VlIss |
| 11 | COD_INF_OBS | CodInfObs |

### B035
| # | ACBr | Go |
|---|------|-----|
| 0 | B035 | B035 |
| 1 | VL_CONT_P | VlContP |
| 2 | VL_BC_ISS_P | VlBcIssP |
| 3 | ALIQ_ISS | AliqIss |
| 4 | VL_ISS_P | VlIssP |
| 5 | VL_ISNT_ISS_P | VlIsntIssP |
| 6 | COD_SERV | CodServ |

### B350
| # | ACBr | Go |
|---|------|-----|
| 0 | B035 | B035 |
| 1 | COD_CTD | CodCtd |
| 2 | CTA_ISS | CtaIss |
| 3 | CTA_COSIF | CtaCosif |
| 4 | QTD_OCOR | QtdOcor |
| 5 | COD_SERV | CodServ |
| 6 | VL_CONT | VlCont |
| 7 | VL_BC_ISS | VlBcIss |
| 8 | ALIQ_ISS | AliqIss |
| 9 | VL_ISS | VlIss |
| 10 | COD_INF_OBS | CodInfObs |

### B420
| # | ACBr | Go |
|---|------|-----|
| 0 | B420 | B420 |
| 1 | VL_CONT | VlCont |
| 2 | VL_BC_ISS | VlBcIss |
| 3 | ALIQ_ISS | AliqIss |
| 4 | VL_ISNT_ISS | VlIsntIss |
| 5 | VL_ISS | VlIss |
| 6 | COD_SERV | CodServ |

### B440
| # | ACBr | Go |
|---|------|-----|
| 0 | B440 | B440 |
| 1 | IndOperToStr(IND_OPER) | IndOper |
| 2 | COD_PART | CodPart |
| 3 | VL_CONT_RT | VlContRT |
| 4 | VL_BC_ISS_RT | VlBcIssRT |
| 5 | VL_ISS_RT | VlIssRT |

### B460
| # | ACBr | Go |
|---|------|-----|
| 0 | B460 | B460 |
| 1 | IndicadorDeducaoToStr(IND_DED) | IndDed.String() |
| 2 | VL_DED | VlDed |
| 3 | NUM_PROC | NumProc |
| 4 | IndicadorProcessoToStr(IND_PROC) | IndProc.String() |
| 5 | PROC | Proc |
| 6 | COD_INF_OBS | CodInfObs |
| 7 | IndicadorObrigacaoToStr(IND_OBR) | IndObr.String() |

### B470
| # | ACBr | Go |
|---|------|-----|
| 0 | B470 | B470 |
| 1 | VL_CONT | VlCont |
| 2 | VL_MAT_TERC | VlMatTerc |
| 3 | VL_MAT_PROP | VlMatProp |
| 4 | VL_SUB | VlSub |
| 5 | VL_ISNT | VlIsnt |
| 6 | VL_DED_BC | VlDedBC |
| 7 | VL_BC_ISS | VlBcIss |
| 8 | VL_BC_ISS_RT | VlBcIssRT |
| 9 | VL_ISS | VlIss |
| 10 | VL_ISS_RT | VlIssRT |
| 11 | VL_DED | VlDed |
| 12 | VL_ISS_REC | VlIssRec |
| 13 | VL_ISS_ST | VlIssST |
| 14 | VL_ISS_REC_UNI | VlIssRecUni |

### B500
| # | ACBr | Go |
|---|------|-----|
| 0 | B500 | B500 |
| 1 | VL_REC | VlRec |
| 2 | QTD_PROF | QtdProf |
| 3 | VL_OR | VlOR |

### B510
| # | ACBr | Go |
|---|------|-----|
| 0 | B510 | B510 |
| 1 | IND_PROF | IndProf |
| 2 | IND_ESC | IndEsc |
| 3 | IND_SOC | IndSoc |
| 4 | CPF | CPF |
| 5 | NOME | Nome |

### B990
| # | ACBr | Go |
|---|------|-----|
| 0 | B990 | B990 |
| 1 | QTD_LIN_B | RegistroB990.QtdLinB |

## Bloco C

### C001
| # | ACBr | Go |
|---|------|-----|
| 0 | C001 | C001 |
| 1 | Integer(IND_MOV) | RegistroC001.IndMov |

### C100
| # | ACBr | Go |
|---|------|-----|
| 0 | C100 | C100 |
| 1 | Integer(IND_OPER) | IndOper |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | strCOD_SIT | CodSit.String() |
| 6 | SER | Ser |
| 7 | NUM_DOC | NumDoc |
| 8 | CHV_NFE | ChvNFe |
| 9 | DT_DOC | dtDoc |
| 10 | DT_E_S | dtES |
| 11 | VL_DOC | VlDoc |
| 12 | strIND_PGTO | indPgto.StringEm(b.DtIni) |
| 13 | VL_DESC | VlDesc |
| 14 | VL_ABAT_NT | VlAbatNT |
| 15 | VL_MERC | VlMerc |
| 16 | strIND_FRT | indFrt.StringEm(b.DtIni) |
| 17 | VL_FRT | VlFrt |
| 18 | VL_SEG | VlSeg |
| 19 | VL_OUT_DA | VlOutDa |
| 20 | VL_BC_ICMS | VlBcICMS |
| 21 | VL_ICMS | VlICMS |
| 22 | VL_BC_ICMS_ST | VlBcICMSST |
| 23 | VL_ICMS_ST | VlICMSST |
| 24 | VL_IPI | VlIPI |
| 25 | VL_PIS | VlPIS |
| 26 | VL_COFINS | VlCOFINS |
| 27 | VL_PIS_ST | VlPISST |
| 28 | VL_COFINS_ST | VlCOFINSST |

### C101
| # | ACBr | Go |
|---|------|-----|
| 0 | C101 | C101 |
| 1 | VL_FCP_UF_DEST | VlFcpUFDest |
| 2 | VL_ICMS_UF_DEST | VlICMSUFDest |
| 3 | VL_ICMS_UF_REM | VlICMSUFRem |

### C105
| # | ACBr | Go |
|---|------|-----|
| 0 | C105 | C105 |
| 1 | strOPER | Oper |
| 2 | vRegC105.UF | UF |

### C110
| # | ACBr | Go |
|---|------|-----|
| 0 | C110 | C110 |
| 1 | COD_INF | CodInf |
| 2 | TXT_COMPL | TxtCompl |

### C111
| # | ACBr | Go |
|---|------|-----|
| 0 | C111 | C111 |
| 1 | NUM_PROC | NumProc |
| 2 | intIND_PROC | IndProc.String() |

### C112
| # | ACBr | Go |
|---|------|-----|
| 0 | C112 | C112 |
| 1 | Integer(COD_DA) | CodDa |
| 2 | UF | UF |
| 3 | NUM_DA | NumDa |
| 4 | COD_AUT | CodAut |
| 5 | VL_DA | VlDa |
| 6 | DT_VCTO | DtVcto |
| 7 | DT_PGTO | DtPgto |

### C113
| # | ACBr | Go |
|---|------|-----|
| 0 | C113 | C113 |
| 1 | Integer(IND_OPER) | IndOper |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | SER | Ser |
| 6 | SUB | Sub |
| 7 | NUM_DOC | NumDoc |
| 8 | DT_DOC | DtDoc |
| 9 | CHV_DOCe | ChvDocE |

### C114
| # | ACBr | Go |
|---|------|-----|
| 0 | C114 | C114 |
| 1 | COD_MOD | CodMod |
| 2 | ECF_FAB | EcfFab |
| 3 | ECF_CX | EcfCx |
| 4 | NUM_DOC | NumDoc |
| 5 | DT_DOC | DtDoc |

### C115
| # | ACBr | Go |
|---|------|-----|
| 0 | C115 | C115 |
| 1 | intIND_CARGA | IndCarga |
| 2 | CNPJ_COL | CNPJCol |
| 3 | IE_COL | IECol |
| 4 | CPF_COL | CPFCol |
| 5 | COD_MUN_COL | CodMunCol |
| 6 | CNPJ_ENTG | CNPJEntg |
| 7 | IE_ENTG | IEEntg |
| 8 | CPF_ENTG | CPFEntg |
| 9 | COD_MUN_ENTG | CodMunEntg |

### C116
| # | ACBr | Go |
|---|------|-----|
| 0 | C116 | C116 |
| 1 | COD_MOD | CodMod |
| 2 | NR_SAT | NrSat |
| 3 | CHV_CFE | ChvCFe |
| 4 | NUM_CFE | NumCFe |
| 5 | DT_DOC | DtDoc |

### C120
| # | ACBr | Go |
|---|------|-----|
| 0 | C120 | C120 |
| 1 | Integer(COD_DOC_IMP) | CodDocImp |
| 2 | NUM_DOC__IMP | NumDocImp |
| 3 | PIS_IMP | PisImp |
| 4 | COFINS_IMP | CofinsImp |
| 5 | NUM_ACDRAW | NumACDraw |

### C130
| # | ACBr | Go |
|---|------|-----|
| 0 | C130 | C130 |
| 1 | VL_SERV_NT | VlServNT |
| 2 | VL_BC_ISSQN | VlBcISSQN |
| 3 | VL_ISSQN | VlISSQN |
| 4 | VL_BC_IRRF | VlBcIRRF |
| 5 | VL_IRRF | VlIRRF |
| 6 | VL_BC_PREV | VlBcPrev |
| 7 | VL_PREV | VlPrev |

### C140
| # | ACBr | Go |
|---|------|-----|
| 0 | C140 | C140 |
| 1 | Integer(IND_EMIT) | IndEmit |
| 2 | strIND_TIT | IndTit.String() |
| 3 | DESC_TIT | DescTit |
| 4 | NUM_TIT | NumTit |
| 5 | QTD_PARC | QtdParc |
| 6 | VL_TIT | VlTit |

### C141
| # | ACBr | Go |
|---|------|-----|
| 0 | C141 | C141 |
| 1 | NUM_PARC | NumParc |
| 2 | DT_VCTO | DtVcto |
| 3 | VL_PARC | VlParc |

### C160
| # | ACBr | Go |
|---|------|-----|
| 0 | C160 | C160 |
| 1 | COD_PART | CodPart |
| 2 | VEIC_ID | VeicID |
| 3 | QTD_VOL | QtdVol |
| 4 | PESO_BRT | PesoBrt |
| 5 | PESO_LIQ | PesoLiq |
| 6 | UF_ID | UFID |

### C165
| # | ACBr | Go |
|---|------|-----|
| 0 | C165 | C165 |
| 1 | COD_PART | CodPart |
| 2 | VEIC_ID | VeicID |
| 3 | COD_AUT | CodAut |
| 4 | NR_PASSE | NrPasse |
| 5 | HORA | Hora |
| 6 | TEMPER | Temper |
| 7 | QTD_VOL | QtdVol |
| 8 | PESO_BRT | PesoBrt |
| 9 | PESO_LIQ | PesoLiq |
| 10 | NOM_MOT | NomMot |
| 11 | CPF | CPF |
| 12 | UF_ID | UFID |

### C170 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | C170 | C170 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | DESCR_COMPL | DescrCompl |
| 4 | UNID | Qtd |
| 5 | VL_ITEM | Unid |
| 6 | VL_DESC | VlItem |
| 7 | Integer(IND_MOV) | VlDesc |
| 8 | CST_ICMS | IndMov |
| 9 | CFOP | CstICMS.String() |
| 10 | COD_NAT | CFOP |
| 11 | VL_BC_ICMS | CodNat |
| 12 | ALIQ_ICMS | VlBcICMS |
| 13 | VL_ICMS | AliqICMS |
| 14 | VL_BC_ICMS_ST | VlICMS |
| 15 | ALIQ_ST | VlBcICMSST |
| 16 | VL_ICMS_ST | AliqST |
| 17 | strIND_APUR | VlICMSST |
| 18 | CST_IPI | IndApur.String() |
| 19 | COD_ENQ | CstIPI.String() |
| 20 | VL_BC_IPI | CodEnq |
| 21 | ALIQ_IPI | VlBcIPI |
| 22 | VL_IPI | AliqIPI |
| 23 | CST_PIS | VlIPI |
| 24 | VL_BC_PIS | CstPIS.String() |
| 25 | ALIQ_PIS_PERC | VlBcPIS |
| 26 | QUANT_BC_PIS | AliqPIS |
| 27 | ALIQ_PIS_R | QtdBcPIS |
| 28 | VL_PIS | AliqPISReais |
| 29 | CST_COFINS | VlPIS |
| 30 | VL_BC_COFINS | CstCOFINS.String() |
| 31 | ALIQ_COFINS_PERC | VlBcCOFINS |
| 32 | QUANT_BC_COFINS | AliqCOFINS |
| 33 | ALIQ_COFINS_R | QtdBcCOFINS |
| 34 | VL_COFINS | AliqCOFINSReais |
| 35 | COD_CTA | VlCOFINS |
| 36 | VL_ABAT_NT | CodCta |
| 37 | — | VlAbatNT |

### C171
| # | ACBr | Go |
|---|------|-----|
| 0 | C171 | C171 |
| 1 | NUM_TANQUE | NumTanque |
| 2 | QTDE | Qtde |

### C172
| # | ACBr | Go |
|---|------|-----|
| 0 | C172 | C172 |
| 1 | VL_BC_ISSQN | VlBcISSQN |
| 2 | ALIQ_ISSQN | AliqISSQN |
| 3 | VL_ISSQN | VlISSQN |

### C173
| # | ACBr | Go |
|---|------|-----|
| 0 | C173 | C173 |
| 1 | LOTE_MED | LoteMed |
| 2 | QTD_ITEM | QtdItem |
| 3 | DT_FAB | DtFab |
| 4 | DT_VAL | DtVal |
| 5 | Integer(IND_MED) | IndMed |
| 6 | Integer(TP_PROD) | TpProd |
| 7 | VL_TAB_MAX | VlTabMax |

### C174
| # | ACBr | Go |
|---|------|-----|
| 0 | C174 | C174 |
| 1 | Integer(IND_ARM) | IndArm |
| 2 | NUM_ARM | NumArm |
| 3 | DESCR_COMPL | DescrCompl |

### C175
| # | ACBr | Go |
|---|------|-----|
| 0 | C175 | C175 |
| 1 | intIND_VEIC_OPER | indVeicOperInt(r.IndVeicOper) |
| 2 | CNPJ | CNPJ |
| 3 | UF | UF |
| 4 | CHASSI_VEIC | ChassiVeic |

### C176
| # | ACBr | Go |
|---|------|-----|
| 0 | C176 | C176 |
| 1 | COD_MOD_ULT_E | CodModUltE |
| 2 | NUM_DOC_ULT_E | NumDocUltE |
| 3 | SER_ULT_E | SerUltE |
| 4 | DT_ULT_E | DtUltE |
| 5 | COD_PART_ULT_E | CodPartUltE |
| 6 | QUANT_ULT_E | QuantUltE |
| 7 | VL_UNIT_ULT_E | VlUnitUltE |
| 8 | VL_UNIT_BC_ST | VlUnitBcST |
| 9 | CHAVE_NFE_ULT_E | ChaveNfeUltE |
| 10 | NUM_ITEM_ULT_E | NumItemUltE |
| 11 | VL_UNIT_BC_ICMS_ULT_E | VlUnitBcICMSUltE |
| 12 | ALIQ_ICMS_ULT_E | AliqICMSUltE |
| 13 | VL_UNIT_LIMITE_BC_ICMS_ULT_E | VlUnitLimiteBcICMSUltE |
| 14 | VL_UNIT_ICMS_ULT_E | VlUnitICMSUltE |
| 15 | ALIQ_ST_ULT_E | AliqSTUltE |
| 16 | VL_UNIT_RES | VlUnitRes |
| 17 | COD_RESP_RET | CodRespRet |
| 18 | MotivoRessarcimentoToStr(COD_MOT_RES) | CodMotRes.String() |
| 19 | CHAVE_NFE_RET | ChaveNfeRet |
| 20 | COD_PART_NFE_RET | CodPartNfeRet |
| 21 | SER_NFE_RET | SerNfeRet |
| 22 | NUM_NFE_RET | NumNfeRet |
| 23 | ITEM_NFE_RET | ItemNfeRet |
| 24 | COD_DA | CodDA |
| 25 | NUM_DA | NumDA |
| 26 | VL_UNIT_RES_FCP_ST | VlUnitResFcpST |

### C177 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | — | C177 |
| 1 | — | CodSeloIPI |
| 2 | — | QtSeloIPI |
| 3 | — | CodInfItem |

### C178
| # | ACBr | Go |
|---|------|-----|
| 0 | C178 | C178 |
| 1 | CL_ENQ | ClEnq |
| 2 | VL_UNID | VlUnid |
| 3 | QUANT_PAD | QuantPad |

### C179
| # | ACBr | Go |
|---|------|-----|
| 0 | C179 | C179 |
| 1 | BC_ST_ORIG_DEST | BcSTOrigDest |
| 2 | ICMS_ST_REP | ICMSSTRep |
| 3 | ICMS_ST_COMPL | ICMSSTCompl |
| 4 | BC_RET | BcRet |
| 5 | ICMS_RET | ICMSRet |

### C180
| # | ACBr | Go |
|---|------|-----|
| 0 | C180 | C180 |
| 1 | COD_RESP_RET | CodRespRet |
| 2 | QUANT_CONV | QuantConv |
| 3 | UNID | Unid |
| 4 | VL_UNIT_CONV | VlUnitConv |
| 5 | VL_UNIT_ICMS_OP_CONV | VlUnitICMSOpConv |
| 6 | VL_UNIT_BC_ICMS_ST_CONV | VlUnitBcICMSSTConv |
| 7 | VL_UNIT_ICMS_ST_CONV | VlUnitICMSSTConv |
| 8 | VL_UNIT_FCP_ST_CONV | VlUnitFcpSTConv |
| 9 | COD_DA | CodDA |
| 10 | NUM_DA | NumDA |

### C181
| # | ACBr | Go |
|---|------|-----|
| 0 | C181 | C181 |
| 1 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 2 | QUANT_CONV | QuantConv |
| 3 | UNID | Unid |
| 4 | COD_MOD_SAIDA | CodModSaida |
| 5 | SERIE_SAIDA | SerieSaida |
| 6 | ECF_FAB_SAIDA | EcfFabSaida |
| 7 | NUM_DOC_SAIDA | NumDocSaida |
| 8 | CHV_DFE_SAIDA | ChvDfeSaida |
| 9 | DT_DOC_SAIDA | DtDocSaida |
| 10 | NUM_ITEM_SAIDA | NumItemSaida |
| 11 | VL_UNIT_CONV_SAIDA | VlUnitConvSaida |
| 12 | VL_UNIT_ICMS_OP_ESTOQUE_CONV_SAIDA | VlUnitICMSOpEstoqueConvSaida |
| 13 | VL_UNIT_ICMS_ST_ESTOQUE_CONV_SAIDA | VlUnitICMSSTEstoqueConvSaida |
| 14 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV_SAIDA | VlUnitFcpICMSSTEstoqueConvSaida |
| 15 | VL_UNIT_ICMS_NA_OPERACAO_CONV_SAIDA | VlUnitICMSNaOperacaoConvSaida |
| 16 | VL_UNIT_ICMS_OP_CONV_SAIDA | VlUnitICMSOpConvSaida |
| 17 | VL_UNIT_ICMS_ST_CONV_REST | VlUnitICMSSTConvRest |
| 18 | VL_UNIT_FCP_ST_CONV_REST | VlUnitFcpSTConvRest |
| 19 | VL_UNIT_ICMS_ST_CONV_COMPL | VlUnitICMSSTConvCompl |
| 20 | VL_UNIT_FCP_ST_CONV_COMPL | VlUnitFcpSTConvCompl |

### C185
| # | ACBr | Go |
|---|------|-----|
| 0 | C185 | C185 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | CST_ICMS | CstICMS |
| 4 | CFOP | CFOP |
| 5 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 6 | QUANT_CONV | QuantConv |
| 7 | UNID | Unid |
| 8 | VL_UNIT_CONV | VlUnitConv |
| 9 | VL_UNIT_ICMS_NA_OPERACAO_CONV | VlUnitICMSNaOperacaoConv |
| 10 | VL_UNIT_ICMS_OP_CONV | VlUnitICMSOpConv |
| 11 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | VlUnitICMSOpEstoqueConv |
| 12 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | VlUnitICMSSTEstoqueConv |
| 13 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | VlUnitFcpICMSSTEstoqueConv |
| 14 | VL_UNIT_ICMS_ST_CONV_REST | VlUnitICMSSTConvRest |
| 15 | VL_UNIT_FCP_ST_CONV_REST | VlUnitFcpSTConvRest |
| 16 | VL_UNIT_ICMS_ST_CONV_COMPL | VlUnitICMSSTConvCompl |
| 17 | VL_UNIT_FCP_ST_CONV_COMPL | VlUnitFcpSTConvCompl |

### C186
| # | ACBr | Go |
|---|------|-----|
| 0 | C186 | C186 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | CST_ICMS | CstICMS |
| 4 | CFOP | CFOP |
| 5 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 6 | QUANT_CONV | QuantConv |
| 7 | UNID | Unid |
| 8 | COD_MOD_ENTRADA | CodModEntrada |
| 9 | SERIE_ENTRADA | SerieEntrada |
| 10 | NUM_DOC_ENTRADA | NumDocEntrada |
| 11 | CHV_DFE_ENTRADA | ChvDfeEntrada |
| 12 | DT_DOC_ENTRADA | DtDocEntrada |
| 13 | NUM_ITEM_ENTRADA | NumItemEntrada |
| 14 | VL_UNIT_CONV_ENTRADA | VlUnitConvEntrada |
| 15 | VL_UNIT_ICMS_OP_CONV_ENTRADA | VlUnitICMSOpConvEntrada |
| 16 | VL_UNIT_BC_ICMS_ST_CONV_ENTRADA | VlUnitBcICMSSTConvEntrada |
| 17 | VL_UNIT_ICMS_ST_CONV_ENTRADA | VlUnitICMSSTConvEntrada |
| 18 | VL_UNIT_FCP_ST_CONV_ENTRADA | VlUnitFcpSTConvEntrada |

### C190
| # | ACBr | Go |
|---|------|-----|
| 0 | C190 | C190 |
| 1 | CST_ICMS | CstICMS.String() |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_ST | VlBcICMSST |
| 8 | VL_ICMS_ST | VlICMSST |
| 9 | VL_RED_BC | VlRedBC |
| 10 | VL_IPI | VlIPI |
| 11 | COD_OBS | CodObs |

### C191
| # | ACBr | Go |
|---|------|-----|
| 0 | C191 | C191 |
| 1 | VL_FCP_OP | VlFcpOp |
| 2 | VL_FCP_ST | VlFcpST |
| 3 | VL_FCP_RET | VlFcpRet |

### C195
| # | ACBr | Go |
|---|------|-----|
| 0 | C195 | C195 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### C197
| # | ACBr | Go |
|---|------|-----|
| 0 | C197 | C197 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

### C300
| # | ACBr | Go |
|---|------|-----|
| 0 | C300 | C300 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | SUB | Sub |
| 4 | NUM_DOC_INI | NumDocIni |
| 5 | NUM_DOC_FIN | NumDocFin |
| 6 | DT_DOC | DtDoc |
| 7 | VL_DOC | VlDoc |
| 8 | VL_PIS | VlPIS |
| 9 | VL_COFINS | VlCOFINS |
| 10 | COD_CTA | CodCta |

### C310
| # | ACBr | Go |
|---|------|-----|
| 0 | C310 | C310 |
| 1 | NUM_DOC_CANC | NumDocCanc |

### C320
| # | ACBr | Go |
|---|------|-----|
| 0 | C320 | C320 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | COD_OBS | CodObs |

### C321
| # | ACBr | Go |
|---|------|-----|
| 0 | C321 | C321 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |
| 3 | UNID | Unid |
| 4 | VL_ITEM | VlItem |
| 5 | VL_DESC | VlDesc |
| 6 | VL_BC_ICMS | VlBcICMS |
| 7 | VL_ICMS | VlICMS |
| 8 | VL_PIS | VlPIS |
| 9 | VL_COFINS | VlCOFINS |

### C330
| # | ACBr | Go |
|---|------|-----|
| 0 | C330 | C330 |
| 1 | COD_MOT_REST_COMPL | codMotRestCompl |
| 2 | QUANT_CONV | quantConv |
| 3 | UNID | unid |
| 4 | VL_UNIT_CONV | vlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | vlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | vlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | vlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | vlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | vlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | vlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | vlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | vlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | vlUnitFcpSTConvCompl |

### C350
| # | ACBr | Go |
|---|------|-----|
| 0 | C350 | C350 |
| 1 | SER | Ser |
| 2 | SUB_SER | SubSer |
| 3 | NUM_DOC | NumDoc |
| 4 | DT_DOC | DtDoc |
| 5 | CNPJ_CPF | CNPJCPF |
| 6 | VL_MERC | VlMerc |
| 7 | VL_DOC | VlDoc |
| 8 | VL_DESC | VlDesc |
| 9 | VL_PIS | VlPIS |
| 10 | VL_COFINS | VlCOFINS |
| 11 | COD_CTA | CodCta |

### C370
| # | ACBr | Go |
|---|------|-----|
| 0 | C370 | C370 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | UNID | Unid |
| 5 | VL_ITEM | VlItem |
| 6 | VL_DESC | VlDesc |

### C380
| # | ACBr | Go |
|---|------|-----|
| 0 | C380 | C380 |
| 1 | COD_MOT_REST_COMPL | codMotRestCompl |
| 2 | QUANT_CONV | quantConv |
| 3 | UNID | unid |
| 4 | VL_UNIT_CONV | vlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | vlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | vlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | vlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | vlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | vlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | vlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | vlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | vlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | vlUnitFcpSTConvCompl |
| 14 | CST_ICMS | CstICMS |
| 15 | CFOP | CFOP |

### C390
| # | ACBr | Go |
|---|------|-----|
| 0 | C390 | C390 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | COD_OBS | CodObs |

### C400
| # | ACBr | Go |
|---|------|-----|
| 0 | C400 | C400 |
| 1 | COD_MOD | CodMod |
| 2 | ECF_MOD | EcfMod |
| 3 | ECF_FAB | EcfFab |
| 4 | ECF_CX | EcfCx |

### C405 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | C405 | C405 |
| 1 | DT_DOC | DtDoc |
| 2 | CRO | Cro |
| 3 | CRZ | Crz |
| 4 | NUM_COO_FIN | NumCooFin |
| 5 | NUM_COO_FIN | GtFin |
| 6 | GT_FIN | VlBrt |
| 7 | VL_BRT | — |

### C410
| # | ACBr | Go |
|---|------|-----|
| 0 | C410 | C410 |
| 1 | VL_PIS | VlPIS |
| 2 | VL_COFINS | VlCOFINS |

### C420
| # | ACBr | Go |
|---|------|-----|
| 0 | C420 | C420 |
| 1 | COD_TOT_PAR | CodTotPar |
| 2 | VLR_ACUM_TOT | VlrAcumTot |
| 3 | NR_TOT | NrTot |
| 4 | DESCR_NR_TOT | DescrNrTot |

### C425
| # | ACBr | Go |
|---|------|-----|
| 0 | C425 | C425 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |
| 3 | UNID | Unid |
| 4 | VL_ITEM | VlItem |
| 5 | VL_PIS | VlPIS |
| 6 | VL_COFINS | VlCOFINS |

### C430
| # | ACBr | Go |
|---|------|-----|
| 0 | C430 | C430 |
| 1 | COD_MOT_REST_COMPL | codMotRestCompl |
| 2 | QUANT_CONV | quantConv |
| 3 | UNID | unid |
| 4 | VL_UNIT_CONV | vlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | vlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | vlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | vlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | vlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | vlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | vlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | vlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | vlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | vlUnitFcpSTConvCompl |
| 14 | CST_ICMS | CstICMS |
| 15 | CFOP | CFOP |

### C460 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | C460 | C460 |
| 1 | COD_MOD | CodMod |
| 2 | strCOD_SIT | CodSit.String() |
| 3 | NUM_DOC | NumDoc |
| 4 | NUM_DOC | DtDoc |
| 5 | DT_DOC | VlDoc |
| 6 | VL_DOC | VlPIS |
| 7 | VL_PIS | VlCOFINS |
| 8 | VL_COFINS | CPFCNPJ |
| 9 | CPF_CNPJ | NomAdq |
| 10 | NOM_ADQ | — |

### C465
| # | ACBr | Go |
|---|------|-----|
| 0 | C465 | C465 |
| 1 | CHV_CFE | ChvCFe |
| 2 | NUM_CCF | NumCCF |

### C470
| # | ACBr | Go |
|---|------|-----|
| 0 | C470 | C470 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |
| 3 | QTD_CANC | QtdCanc |
| 4 | UNID | Unid |
| 5 | VL_ITEM | VlItem |
| 6 | CST_ICMS | CstICMS |
| 7 | CFOP | CFOP |
| 8 | ALIQ_ICMS | AliqICMS |
| 9 | VL_PIS | VlPIS |
| 10 | VL_COFINS | VlCOFINS |

### C480
| # | ACBr | Go |
|---|------|-----|
| 0 | C480 | C480 |
| 1 | COD_MOT_REST_COMPL | codMotRestCompl |
| 2 | QUANT_CONV | quantConv |
| 3 | UNID | unid |
| 4 | VL_UNIT_CONV | vlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | vlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | vlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | vlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | vlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | vlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | vlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | vlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | vlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | vlUnitFcpSTConvCompl |
| 14 | CST_ICMS | CstICMS |
| 15 | CFOP | CFOP |

### C490
| # | ACBr | Go |
|---|------|-----|
| 0 | C490 | C490 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | COD_OBS | CodObs |

### C495
| # | ACBr | Go |
|---|------|-----|
| 0 | C495 | C495 |
| 1 | ALIQ_ICMS | AliqICMS |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | QTD_CANC | QtdCanc |
| 5 | UNID | Unid |
| 6 | VL_ITEM | VlItem |
| 7 | VL_DESC | VlDesc |
| 8 | VL_CANC | VlCanc |
| 9 | VL_ACMO | VlAcmo |
| 10 | VL_BC_ICMS | VlBcICMS |
| 11 | VL_ICMS | VlICMS |
| 12 | VL_ISEN | VlIsen |
| 13 | VL_NT | VlNT |
| 14 | VL_ICMS_ST | VlICMSST |

### C500
| # | ACBr | Go |
|---|------|-----|
| 0 | C500 | C500 |
| 1 | Integer(UmRegC500.IND_OPER) | IndOper |
| 2 | Integer(UmRegC500.IND_EMIT) | IndEmit |
| 3 | UmRegC500.COD_PART | CodPart |
| 4 | UmRegC500.COD_MOD | CodMod |
| 5 | strCOD_SIT | CodSit.String() |
| 6 | UmRegC500.SER | Ser |
| 7 | UmRegC500.SUB | Sub |
| 8 | UmRegC500.COD_CONS | CodCons |
| 9 | UmRegC500.NUM_DOC | NumDoc |
| 10 | UmRegC500.DT_DOC | DtDoc |
| 11 | UmRegC500.DT_E_S | DtES |
| 12 | UmRegC500.VL_DOC | VlDoc |
| 13 | UmRegC500.VL_DESC | VlDesc |
| 14 | UmRegC500.VL_FORN | VlForn |
| 15 | UmRegC500.VL_SERV_NT | VlServNT |
| 16 | UmRegC500.VL_TERC | VlTerc |
| 17 | UmRegC500.VL_DA | VlDa |
| 18 | UmRegC500.VL_BC_ICMS | VlBcICMS |
| 19 | UmRegC500.VL_ICMS | VlICMS |
| 20 | UmRegC500.VL_BC_ICMS_ST | VlBcICMSST |
| 21 | UmRegC500.VL_ICMS_ST | VlICMSST |
| 22 | UmRegC500.COD_INF | CodInf |
| 23 | UmRegC500.VL_PIS | VlPIS |
| 24 | UmRegC500.VL_COFINS | VlCOFINS |
| 25 | intTP_LIGACAO | tpLigacaoIntC(r.TpLigacao) |
| 26 | strCOD_GRUPO_TENSAO | CodGrupoTensao.String() |
| 27 | UmRegC500.CHV_DOCe | ChvDOCe |
| 28 | vFin_DOCe | fin |
| 29 | UmRegC500.CHV_DOCe_REF | ChvDOCeRef |
| 30 | intIND_DEST | indDest |
| 31 | UmRegC500.COD_MUN_DEST | CodMunDest |
| 32 | UmRegC500.COD_CTA | CodCta |
| 33 | UmRegC500.COD_MOD_DOC_REF | CodModDocRef |
| 34 | UmRegC500.HASH_DOC_REF | HashDocRef |
| 35 | UmRegC500.SER_DOC_REF | SerDocRef |
| 36 | UmRegC500.NUM_DOC_REF | NumDocRef |
| 37 | UmRegC500.MES_DOC_REF | MesDocRef |
| 38 | UmRegC500.ENER_INJET | EnerInjet |
| 39 | UmRegC500.OUTRAS_DED | OutrasDed |

### C510
| # | ACBr | Go |
|---|------|-----|
| 0 | C510 | C510 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | COD_CLASS | CodClass |
| 4 | QTD | Qtd |
| 5 | UNID | Unid |
| 6 | VL_ITEM | VlItem |
| 7 | VL_DESC | VlDesc |
| 8 | CST_ICMS | CstICMS |
| 9 | CFOP | CFOP |
| 10 | VL_BC_ICMS | VlBcICMS |
| 11 | ALIQ_ICMS | AliqICMS |
| 12 | VL_ICMS | VlICMS |
| 13 | VL_BC_ICMS_ST | VlBcICMSST |
| 14 | ALIQ_ST | AliqST |
| 15 | VL_ICMS_ST | VlICMSST |
| 16 | Integer(IND_REC) | IndRec |
| 17 | COD_PART | CodPart |
| 18 | VL_PIS | VlPIS |
| 19 | VL_COFINS | VlCOFINS |
| 20 | COD_CTA | CodCta |

### C590
| # | ACBr | Go |
|---|------|-----|
| 0 | C590 | C590 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_ST | VlBcICMSST |
| 8 | VL_ICMS_ST | VlICMSST |
| 9 | VL_RED_BC | VlRedBC |
| 10 | COD_OBS | CodObs |

### C591
| # | ACBr | Go |
|---|------|-----|
| 0 | C591 | C591 |
| 1 | VL_FCP_OP | VlFcpOp |
| 2 | VL_FCP_ST | VlFcpST |

### C595
| # | ACBr | Go |
|---|------|-----|
| 0 | C595 | C595 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### C597
| # | ACBr | Go |
|---|------|-----|
| 0 | C597 | C597 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

### C600
| # | ACBr | Go |
|---|------|-----|
| 0 | C600 | C600 |
| 1 | COD_MOD | CodMod |
| 2 | COD_MUN | CodMun |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | COD_CONS | CodCons |
| 6 | QTD_CONS | QtdCons |
| 7 | QTD_CANC | QtdCanc |
| 8 | DT_DOC | DtDoc |
| 9 | VL_DOC | VlDoc |
| 10 | VL_DESC | VlDesc |
| 11 | CONS | Cons |
| 12 | VL_FORN | VlForn |
| 13 | VL_SERV_NT | VlServNT |
| 14 | VL_TERC | VlTerc |
| 15 | VL_DA | VlDa |
| 16 | VL_BC_ICMS | VlBcICMS |
| 17 | VL_ICMS | VlICMS |
| 18 | VL_BC_ICMS_ST | VlBcICMSST |
| 19 | VL_ICMS_ST | VlICMSST |
| 20 | VL_PIS | VlPIS |
| 21 | VL_COFINS | VlCOFINS |

### C601
| # | ACBr | Go |
|---|------|-----|
| 0 | C601 | C601 |
| 1 | NUM_DOC_CANC | NumDocCanc |

### C610
| # | ACBr | Go |
|---|------|-----|
| 0 | C610 | C610 |
| 1 | COD_CLASS | CodClass |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | UNID | Unid |
| 5 | VL_ITEM | VlItem |
| 6 | VL_DESC | VlDesc |
| 7 | CST_ICMS | CstICMS |
| 8 | CFOP | CFOP |
| 9 | ALIQ_ICMS | AliqICMS |
| 10 | VL_BC_ICMS | VlBcICMS |
| 11 | VL_ICMS | VlICMS |
| 12 | VL_BC_ICMS_ST | VlBcICMSST |
| 13 | VL_ICMS_ST | VlICMSST |
| 14 | VL_PIS | VlPIS |
| 15 | VL_COFINS | VlCOFINS |
| 16 | COD_CTA | CodCta |

### C690
| # | ACBr | Go |
|---|------|-----|
| 0 | C690 | C690 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | VL_RED_BC | VlRedBC |
| 9 | VL_BC_ICMS_ST | VlBcICMSST |
| 10 | VL_ICMS_ST | VlICMSST |

### C700
| # | ACBr | Go |
|---|------|-----|
| 0 | C700 | C700 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | NRO_ORD_INI | NroOrdIni |
| 4 | NRO_ORD_FIN | NroOrdFin |
| 5 | DT_DOC_INI | DtDocIni |
| 6 | DT_DOC_FIN | DtDocFin |
| 7 | NOM_MEST | NomMest |
| 8 | CHV_COD_DIG | ChvCodDig |

### C790
| # | ACBr | Go |
|---|------|-----|
| 0 | C790 | C790 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_ST | VlBcICMSST |
| 8 | VL_ICMS_ST | VlICMSST |
| 9 | VL_RED_BC | VlRedBC |
| 10 | COD_OBS | CodObs |

### C791
| # | ACBr | Go |
|---|------|-----|
| 0 | C791 | C791 |
| 1 | UF | UF |
| 2 | VL_BC_ICMS_ST | VlBcICMSST |
| 3 | VL_ICMS_ST | VlICMSST |

### C800
| # | ACBr | Go |
|---|------|-----|
| 0 | C800 | C800 |
| 1 | COD_MOD | CodMod |
| 2 | strCOD_SIT | CodSit.String() |
| 3 | NUM_CFE | NumCFe |
| 4 | DT_DOC | DtDoc |
| 5 | VL_CFE | VlCFe |
| 6 | VL_PIS | pis |
| 7 | VL_COFINS | cofins |
| 8 | CNPJ_CPF | CNPJCPF |
| 9 | NR_SAT | NrSat |
| 10 | CHV_CFE | ChvCFe |
| 11 | VL_DESC | VlDesc |
| 12 | VL_MERC | VlMerc |
| 13 | VL_OUT_DA | VlOutDa |
| 14 | VL_ICMS | VlICMS |
| 15 | VL_PIS_ST | pisST |
| 16 | VL_COFINS_ST | cofinsST |

### C810
| # | ACBr | Go |
|---|------|-----|
| 0 | C810 | C810 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | UNID | Unid |
| 5 | VL_ITEM | VlItem |
| 6 | CST_ICMS | CstICMS |
| 7 | CFOP | CFOP |

### C815
| # | ACBr | Go |
|---|------|-----|
| 0 | C815 | C815 |
| 1 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 2 | QUANT_CONV | QuantConv |
| 3 | UNID | Unid |
| 4 | VL_UNIT_CONV | VlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | VlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | VlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | VlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | VlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | VlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | VlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | VlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | VlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | VlUnitFcpSTConvCompl |

### C850
| # | ACBr | Go |
|---|------|-----|
| 0 | C850 | C850 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | COD_OBS | CodObs |

### C855
| # | ACBr | Go |
|---|------|-----|
| 0 | C855 | C855 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### C857
| # | ACBr | Go |
|---|------|-----|
| 0 | C857 | C857 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

### C860
| # | ACBr | Go |
|---|------|-----|
| 0 | C860 | C860 |
| 1 | COD_MOD | CodMod |
| 2 | NR_SAT | NrSat |
| 3 | DT_DOC | DtDoc |
| 4 | DOC_INI | DocIni |
| 5 | DOC_FIN | DocFin |

### C870
| # | ACBr | Go |
|---|------|-----|
| 0 | C870 | C870 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |
| 3 | UNID | Unid |
| 4 | CST_ICMS | CstICMS |
| 5 | CFOP | CFOP |

### C880
| # | ACBr | Go |
|---|------|-----|
| 0 | C880 | C880 |
| 1 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 2 | QUANT_CONV | QuantConv |
| 3 | UNID | Unid |
| 4 | VL_UNIT_CONV | VlUnitConv |
| 5 | VL_UNIT_ICMS_NA_OPERACAO_CONV | VlUnitICMSNaOperacaoConv |
| 6 | VL_UNIT_ICMS_OP_CONV | VlUnitICMSOpConv |
| 7 | VL_UNIT_ICMS_OP_ESTOQUE_CONV | VlUnitICMSOpEstoqueConv |
| 8 | VL_UNIT_ICMS_ST_ESTOQUE_CONV | VlUnitICMSSTEstoqueConv |
| 9 | VL_UNIT_FCP_ICMS_ST_ESTOQUE_CONV | VlUnitFcpICMSSTEstoqueConv |
| 10 | VL_UNIT_ICMS_ST_CONV_REST | VlUnitICMSSTConvRest |
| 11 | VL_UNIT_FCP_ST_CONV_REST | VlUnitFcpSTConvRest |
| 12 | VL_UNIT_ICMS_ST_CONV_COMPL | VlUnitICMSSTConvCompl |
| 13 | VL_UNIT_FCP_ST_CONV_COMPL | VlUnitFcpSTConvCompl |

### C890
| # | ACBr | Go |
|---|------|-----|
| 0 | C890 | C890 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | COD_OBS | CodObs |

### C895
| # | ACBr | Go |
|---|------|-----|
| 0 | C895 | C895 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### C897
| # | ACBr | Go |
|---|------|-----|
| 0 | C897 | C897 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

### C990
| # | ACBr | Go |
|---|------|-----|
| 0 | C990 | C990 |
| 1 | QTD_LIN_C | RegistroC990.QtdLinC |

## Bloco D

### D001
| # | ACBr | Go |
|---|------|-----|
| 0 | D001 | D001 |
| 1 | Integer(IND_MOV) | RegistroD001.IndMov |

### D100
| # | ACBr | Go |
|---|------|-----|
| 0 | D100 | D100 |
| 1 | Integer(IND_OPER) | IndOper |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | strCOD_SIT | CodSit.String() |
| 6 | SER | Ser |
| 7 | SUB | Sub |
| 8 | NUM_DOC | NumDoc |
| 9 | ChaveEletronicaCTe | chvCTe |
| 10 | DT_DOC | DtDoc |
| 11 | DT_A_P | DtAP |
| 12 | TP_CT_e | TpCTe |
| 13 | CHV_CTE_REF | ChvCTeRef |
| 14 | VL_DOC | VlDoc |
| 15 | VL_DESC | VlDesc |
| 16 | strIND_FRT | IndFrt.StringEmD100(b.DtIni) |
| 17 | VL_SERV | VlServ |
| 18 | VL_BC_ICMS | VlBcICMS |
| 19 | VL_ICMS | VlICMS |
| 20 | VL_NT | VlNT |
| 21 | COD_INF | CodInf |
| 22 | COD_CTA | CodCta |
| 23 | COD_MUN_ORIG | CodMunOrig |
| 24 | COD_MUN_DEST | CodMunDest |

### D101
| # | ACBr | Go |
|---|------|-----|
| 0 | D101 | D101 |
| 1 | VL_FCP_UF_DEST | VlFcpUFDest |
| 2 | VL_ICMS_UF_DEST | VlICMSUFDest |
| 3 | VL_ICMS_UF_REM | VlICMSUFRem |

### D110
| # | ACBr | Go |
|---|------|-----|
| 0 | D110 | D110 |
| 1 | NUN_ITEM | NunItem |
| 2 | COD_ITEM | CodItem |
| 3 | VL_SERV | VlServ |
| 4 | VL_OUT | VlOut |

### D120
| # | ACBr | Go |
|---|------|-----|
| 0 | D120 | D120 |
| 1 | COD_MUN_ORIG | CodMunOrig |
| 2 | COD_MUN_DEST | CodMunDest |
| 3 | VEIC_ID | VeicID |
| 4 | UF_ID | UFID |

### D130
| # | ACBr | Go |
|---|------|-----|
| 0 | D130 | D130 |
| 1 | COD_PART_CONSG | CodPartConsg |
| 2 | COD_PART_RED | CodPartRed |
| 3 | strIND_FRT_RED | IndFrtRed.String() |
| 4 | COD_MUN_ORIG | CodMunOrig |
| 5 | COD_MUN_DEST | CodMunDest |
| 6 | VEIC_ID | VeicID |
| 7 | VL_LIQ_FRT | VlLiqFrt |
| 8 | VL_SEC_CAT | VlSecCat |
| 9 | VL_DESP | VlDesp |
| 10 | VL_PEDG | VlPedg |
| 11 | VL_OUT | VlOut |
| 12 | VL_FRT | VlFrt |
| 13 | UF_ID | UFID |

### D140
| # | ACBr | Go |
|---|------|-----|
| 0 | D140 | D140 |
| 1 | COD_PART_CONSG | CodPartConsg |
| 2 | COD_MUN_ORIG | CodMunOrig |
| 3 | COD_MUN_DEST | CodMunDest |
| 4 | Integer(IND_VEIC) | IndVeic |
| 5 | VEIC_ID | VeicID |
| 6 | Integer(IND_NAV) | IndNav |
| 7 | VIAGEM | Viagem |
| 8 | VL_FRT_LIQ | VlFrtLiq |
| 9 | VL_DESP_PORT | VlDespPort |
| 10 | VL_DESP_CAR_DESC | VlDespCarDesc |
| 11 | VL_OUT | VlOut |
| 12 | VL_FRT_BRT | VlFrtBrt |
| 13 | VL_FRT_MM | VlFrtMM |

### D150
| # | ACBr | Go |
|---|------|-----|
| 0 | D150 | D150 |
| 1 | COD_MUN_ORIG | CodMunOrig |
| 2 | COD_MUN_DEST | CodMunDest |
| 3 | VEIC_ID | VeicID |
| 4 | VIAGEM | Viagem |
| 5 | intIND_TFA | IndTFA |
| 6 | VL_PESO_TX | VlPesoTx |
| 7 | VL_TX_TERR | VlTxTerr |
| 8 | VL_TX_RED | VlTxRed |
| 9 | VL_OUT | VlOut |
| 10 | VL_TX_ADV | VlTxAdv |

### D160
| # | ACBr | Go |
|---|------|-----|
| 0 | D160 | D160 |
| 1 | DESPACHO | Despacho |
| 2 | CNPJ_CPF_REM | CNPJCPFRem |
| 3 | IE_REM | IERem |
| 4 | COD_MUN_ORI | CodMunOri |
| 5 | CNPJ_CPF_DEST | CNPJCPFDest |
| 6 | IE_DEST | IEDest |
| 7 | COD_MUN_DEST | CodMunDest |

### D161
| # | ACBr | Go |
|---|------|-----|
| 0 | D161 | D161 |
| 1 | intIND_CARGA | IndCarga |
| 2 | CNPJ_COL | CNPJCol |
| 3 | IE_COL | IECol |
| 4 | COD_MUN_COL | CodMunCol |
| 5 | CNPJ_ENTG | CNPJEntg |
| 6 | IE_ENTG | IEEntg |
| 7 | COD_MUN_ENTG | CodMunEntg |

### D162
| # | ACBr | Go |
|---|------|-----|
| 0 | D162 | D162 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | NUM_DOC | NumDoc |
| 4 | DT_DOC | DtDoc |
| 5 | VL_DOC | VlDoc |
| 6 | VL_MERC | VlMerc |
| 7 | QTD_VOL | QtdVol |
| 8 | PESO_BRT | PesoBrt |
| 9 | PESO_LIQ | PesoLiq |

### D170
| # | ACBr | Go |
|---|------|-----|
| 0 | D170 | D170 |
| 1 | COD_PART_CONSG | CodPartConsg |
| 2 | COD_PART_RED | CodPartRed |
| 3 | COD_MUN_ORIG | CodMunOrig |
| 4 | COD_MUN_DEST | CodMunDest |
| 5 | OTM | Otm |
| 6 | Integer(IND_NAT_FRT) | IndNatFrt |
| 7 | VL_LIQ_FRT | VlLiqFrt |
| 8 | VL_GRIS | VlGris |
| 9 | VL_PDG | VlPdg |
| 10 | VL_OUT | VlOut |
| 11 | VL_FRT | VlFrt |
| 12 | VL_FRT | VlFrt |
| 13 | UF_ID | UFID |

### D180
| # | ACBr | Go |
|---|------|-----|
| 0 | D180 | D180 |
| 1 | NUM_SEQ | NumSeq |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | CNPJ_EMIT | CNPJEmit |
| 4 | UF_EMIT | UFEmit |
| 5 | IE_EMIT | IEEmit |
| 6 | COD_MUN_ORIG | CodMunOrig |
| 7 | CNPJ_CPF_TOM | CNPJCPFTom |
| 8 | UF_TOM | UFTom |
| 9 | IE_TOM | IETom |
| 10 | COD_MUN_DEST | CodMunDest |
| 11 | COD_MOD | CodMod |
| 12 | SER | Ser |
| 13 | SUB | Sub |
| 14 | NUM_DOC | NumDoc |
| 15 | DT_DOC | DtDoc |
| 16 | VL_DOC | VlDoc |

### D190
| # | ACBr | Go |
|---|------|-----|
| 0 | D190 | D190 |
| 1 | CST_ICMS | CstICMS.String() |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | COD_OBS | CodObs |

### D195
| # | ACBr | Go |
|---|------|-----|
| 0 | D195 | D195 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### D197
| # | ACBr | Go |
|---|------|-----|
| 0 | D197 | D197 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

### D300
| # | ACBr | Go |
|---|------|-----|
| 0 | D300 | D300 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | SUB | Sub |
| 4 | NUM_DOC_INI | NumDocIni |
| 5 | NUM_DOC_FIN | NumDocFin |
| 6 | CST_ICMS | CstICMS |
| 7 | CFOP | CFOP |
| 8 | ALIQ_ICMS | AliqICMS |
| 9 | DT_DOC | DtDoc |
| 10 | VL_OPR | VlOpr |
| 11 | VL_DESC | VlDesc |
| 12 | VL_SERV | VlServ |
| 13 | VL_SEG | VlSeg |
| 14 | VL_OUT_DESP | VlOutDesp |
| 15 | VL_BC_ICMS | VlBcICMS |
| 16 | VL_ICMS | VlICMS |
| 17 | VL_RED_BC | VlRedBC |
| 18 | COD_OBS | CodObs |
| 19 | COD_CTA | CodCta |

### D301
| # | ACBr | Go |
|---|------|-----|
| 0 | D301 | D301 |
| 1 | NUM_DOC_CANC | NumDocCanc |

### D310
| # | ACBr | Go |
|---|------|-----|
| 0 | D310 | D310 |
| 1 | COD_MUN_ORIG | CodMunOrig |
| 2 | VL_SERV | VlServ |
| 3 | VL_BC_ICMS | VlBcICMS |
| 4 | VL_ICMS | VlICMS |

### D350
| # | ACBr | Go |
|---|------|-----|
| 0 | D350 | D350 |
| 1 | COD_MOD | CodMod |
| 2 | ECF_MOD | EcfMod |
| 3 | ECF_FAB | EcfFab |
| 4 | ECF_CX | EcfCx |

### D355
| # | ACBr | Go |
|---|------|-----|
| 0 | D355 | D355 |
| 1 | DT_DOC | DtDoc |
| 2 | CRO | Cro |
| 3 | CRZ | Crz |
| 4 | NUM_COO_FIN | NumCooFin |
| 5 | GT_FIN | GtFin |
| 6 | VL_BRT | VlBrt |

### D360
| # | ACBr | Go |
|---|------|-----|
| 0 | D360 | D360 |
| 1 | VL_PIS | VlPIS |
| 2 | VL_COFINS | VlCOFINS |

### D365
| # | ACBr | Go |
|---|------|-----|
| 0 | D365 | D365 |
| 1 | COD_TOT_PAR | CodTotPar |
| 2 | VLR_ACUM_TOT | VlrAcumTot |
| 3 | NR_TOT | NrTot |
| 4 | DESCR_NR_TOT | DescrNrTot |

### D370
| # | ACBr | Go |
|---|------|-----|
| 0 | D370 | D370 |
| 1 | COD_MUN_ORIG | CodMunOrig |
| 2 | VL_SERV | VlServ |
| 3 | QTD_BILH | QtdBilh |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | VL_ICMS | VlICMS |

### D390
| # | ACBr | Go |
|---|------|-----|
| 0 | D390 | D390 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ISSQN | VlBcISSQN |
| 6 | ALIQ_ISSQN | AliqISSQN |
| 7 | VL_ISSQN | VlISSQN |
| 8 | VL_BC_ICMS | VlBcICMS |
| 9 | VL_ICMS | VlICMS |
| 10 | COD_OBS | CodObs |

### D400
| # | ACBr | Go |
|---|------|-----|
| 0 | D400 | D400 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | strCOD_SIT | CodSit.String() |
| 4 | SER | Ser |
| 5 | SUB | Sub |
| 6 | NUM_DOC | NumDoc |
| 7 | DT_DOC | DtDoc |
| 8 | VL_DOC | VlDoc |
| 9 | VL_DESC | VlDesc |
| 10 | VL_SERV | VlServ |
| 11 | VL_BC_ICMS | VlBcICMS |
| 12 | VL_ICMS | VlICMS |
| 13 | VL_PIS | VlPIS |
| 14 | VL_COFINS | VlCOFINS |
| 15 | COD_CTA | CodCta |

### D410
| # | ACBr | Go |
|---|------|-----|
| 0 | D410 | D410 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | SUB | Sub |
| 4 | NUM_DOC_INI | NumDocIni |
| 5 | NUM_DOC_FIN | NumDocFin |
| 6 | DT_DOC | DtDoc |
| 7 | CST_ICMS | CstICMS |
| 8 | CFOP | CFOP |
| 9 | ALIQ_ICMS | AliqICMS |
| 10 | VL_OPR | VlOpr |
| 11 | VL_DESC | VlDesc |
| 12 | VL_SERV | VlServ |
| 13 | VL_BC_ICMS | VlBcICMS |
| 14 | VL_ICMS | VlICMS |

### D411
| # | ACBr | Go |
|---|------|-----|
| 0 | D411 | D411 |
| 1 | NUM_DOC_CANC | NumDocCanc |

### D420
| # | ACBr | Go |
|---|------|-----|
| 0 | D420 | D420 |
| 1 | COD_MUN_ORIG | CodMunOrig |
| 2 | VL_SERV | VlServ |
| 3 | VL_BC_ICMS | VlBcICMS |
| 4 | VL_ICMS | VlICMS |

### D500
| # | ACBr | Go |
|---|------|-----|
| 0 | D500 | D500 |
| 1 | Integer(IND_OPER) | IndOper |
| 2 | Integer(IND_EMIT) | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | strCOD_SIT | CodSit.String() |
| 6 | SER | Ser |
| 7 | SUB | Sub |
| 8 | NUM_DOC | NumDoc |
| 9 | DT_DOC | DtDoc |
| 10 | DT_A_P | DtAP |
| 11 | VL_DOC | VlDoc |
| 12 | VL_DESC | VlDesc |
| 13 | VL_SERV | VlServ |
| 14 | VL_SERV_NT | VlServNT |
| 15 | VL_TERC | VlTerc |
| 16 | VL_DA | VlDa |
| 17 | VL_BC_ICMS | VlBcICMS |
| 18 | VL_ICMS | VlICMS |
| 19 | COD_INF | CodInf |
| 20 | VL_PIS | VlPIS |
| 21 | VL_COFINS | VlCOFINS |
| 22 | COD_CTA | CodCta |
| 23 | intTP_ASSINANTE | tpAssinanteInt(r.TpAssinante) |

### D510 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | D510 | D510 |
| 1 | — | NumItem |
| 2 | — | CodItem |
| 3 | — | CodClass |
| 4 | — | Qtd |
| 5 | — | Unid |
| 6 | — | VlItem |
| 7 | — | VlDesc |
| 8 | — | CstICMS |
| 9 | — | CFOP |
| 10 | — | VlBcICMS |
| 11 | — | AliqICMS |
| 12 | — | VlICMS |
| 13 | — | VlBcICMSUF |
| 14 | — | VlICMSUF |
| 15 | — | IndRec |
| 16 | — | CodPart |
| 17 | — | VlPIS |
| 18 | — | VlCOFINS |
| 19 | — | CodCta |

### D530
| # | ACBr | Go |
|---|------|-----|
| 0 | D530 | D530 |
| 1 | intIND_SERV | IndServ |
| 2 | DT_INI_SERV | DtIniServ |
| 3 | DT_FIN_SERV | DtFinServ |
| 4 | PER_FISCAL | PerFiscal |
| 5 | COD_AREA | CodArea |
| 6 | TERMINAL | Terminal |

### D590
| # | ACBr | Go |
|---|------|-----|
| 0 | D590 | D590 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_UF | VlBcICMSUF |
| 8 | VL_ICMS_UF | VlICMSUF |
| 9 | VL_RED_BC | VlRedBC |
| 10 | COD_OBS | CodObs |

### D600
| # | ACBr | Go |
|---|------|-----|
| 0 | D600 | D600 |
| 1 | COD_MOD | CodMod |
| 2 | COD_MUN | CodMun |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | COD_CONS | CodCons |
| 6 | QTD_CONS | QtdCons |
| 7 | DT_DOC | DtDoc |
| 8 | VL_DOC | VlDoc |
| 9 | VL_DESC | VlDesc |
| 10 | VL_SERV | VlServ |
| 11 | VL_SERV_NT | VlServNT |
| 12 | VL_TERC | VlTerc |
| 13 | VL_DA | VlDa |
| 14 | VL_BC_ICMS | VlBcICMS |
| 15 | VL_ICMS | VlICMS |
| 16 | VL_PIS | VlPIS |
| 17 | VL_COFINS | VlCOFINS |

### D610
| # | ACBr | Go |
|---|------|-----|
| 0 | D610 | D610 |
| 1 | COD_CLASS | CodClass |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | UNID | Unid |
| 5 | VL_ITEM | VlItem |

### D690
| # | ACBr | Go |
|---|------|-----|
| 0 | D690 | D690 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_UF | VlBcICMSUF |
| 8 | VL_ICMS_UF | VlICMSUF |
| 9 | VL_RED_BC | VlRedBC |
| 10 | COD_OBS | CodObs |

### D695
| # | ACBr | Go |
|---|------|-----|
| 0 | D695 | D695 |
| 1 | COD_MOD | CodMod |
| 2 | SER | Ser |
| 3 | NRO_ORD_INI | NroOrdIni |
| 4 | NRO_ORD_FIN | NroOrdFin |
| 5 | DT_DOC_INI | DtDocIni |
| 6 | DT_DOC_FIN | DtDocFin |
| 7 | NOM_MEST | NomMest |
| 8 | CHV_COD_DIG | ChvCodDig |

### D696
| # | ACBr | Go |
|---|------|-----|
| 0 | D696 | D696 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_BC_ICMS_UF | VlBcICMSUF |
| 8 | VL_ICMS_UF | VlICMSUF |
| 9 | VL_RED_BC | VlRedBC |
| 10 | COD_OBS | CodObs |

### D697
| # | ACBr | Go |
|---|------|-----|
| 0 | D697 | D697 |
| 1 | UF | UF |
| 2 | VL_BC_ICMS | VlBcICMS |
| 3 | VL_ICMS | VlICMS |

### D700 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | D700 | D700 |
| 1 | IndOperToStr(wRegD700.IND_OPER) | IndOper |
| 2 | IndEmitToStr(wRegD700.IND_EMIT) | IndEmit |
| 3 | wRegD700.COD_PART | CodPart |
| 4 | wRegD700.COD_MOD | CodMod |
| 5 | wCodSit | cod |
| 6 | wRegD700.SER | Ser |
| 7 | wRegD700.NUM_DOC | NumDoc |
| 8 | wRegD700.DT_DOC | DtDoc |
| 9 | wRegD700.DT_E_S | DtES |
| 10 | wRegD700.VL_DOC | VlDoc |
| 11 | wRegD700.VL_DESC | VlDesc |
| 12 | wRegD700.VL_SERV | VlServ |
| 13 | wRegD700.VL_SERV_NT | VlServNT |
| 14 | wRegD700.VL_TERC | VlTerc |
| 15 | wRegD700.VL_DA | VlDa |
| 16 | wRegD700.VL_BC_ICMS | VlBcICMS |
| 17 | wRegD700.VL_ICMS | VlICMS |
| 18 | wRegD700.COD_INF | CodInf |
| 19 | wRegD700.VL_PIS | VlPIS |
| 20 | wRegD700.VL_COFINS | VlCOFINS |
| 21 | wRegD700.CHV_DOCe | ChvDOCe |
| 22 | — | FinDOCe.String() |
| 23 | — | TipFat.String() |
| 24 | — | FinDOCe.String() |
| 25 | — | TipFat.String() |
| 26 | — | CodModDocRef |
| 27 | — | ChvDOCeRef |
| 28 | — | HashDocRef |
| 29 | — | SerDocRef |
| 30 | — | NumDocRef |
| 31 | — | MesDocRef |
| 32 | — | CodMunDest |
| 33 | — | Ded |

### D730
| # | ACBr | Go |
|---|------|-----|
| 0 | D730 | D730 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | COD_OBS | CodObs |

### D731
| # | ACBr | Go |
|---|------|-----|
| 0 | D731 | D731 |
| 1 | VL_FCP_OP | VlFcpOp |

### D735
| # | ACBr | Go |
|---|------|-----|
| 0 | D735 | D735 |
| 1 | COD_OBS | CodObs |
| 2 | TXT_COMPL | TxtCompl |

### D737
| # | ACBr | Go |
|---|------|-----|
| 0 | D737 | D737 |
| 1 | COD_AJ | CodAj |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | COD_ITEM | CodItem |
| 4 | VL_BC_ICMS | VlBcICMS |
| 5 | ALIQ_ICMS | AliqICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_OUTROS | VlOutros |

--   D750: sem writer no Go (ACBr tem 16 campos)
### D760
| # | ACBr | Go |
|---|------|-----|
| 0 | D760 | D760 |
| 1 | CST_ICMS | CstICMS |
| 2 | CFOP | CFOP |
| 3 | ALIQ_ICMS | AliqICMS |
| 4 | VL_OPR | VlOpr |
| 5 | VL_BC_ICMS | VlBcICMS |
| 6 | VL_ICMS | VlICMS |
| 7 | VL_RED_BC | VlRedBC |
| 8 | COD_OBS | CodObs |

### D761
| # | ACBr | Go |
|---|------|-----|
| 0 | D761 | D761 |
| 1 | VL_FCP_OP | VlFcpOp |

### D990
| # | ACBr | Go |
|---|------|-----|
| 0 | D990 | D990 |
| 1 | QTD_LIN_D | RegistroD990.QtdLinD |

## Bloco E

### E001
| # | ACBr | Go |
|---|------|-----|
| 0 | E001 | E001 |
| 1 | Integer(IND_MOV) | RegistroE001.IndMov |

### E100
| # | ACBr | Go |
|---|------|-----|
| 0 | E100 | E100 |
| 1 | DT_INI | DtIni |
| 2 | DT_FIN | DtFin |

### E110
| # | ACBr | Go |
|---|------|-----|
| 0 | E110 | E110 |
| 1 | VL_TOT_DEBITOS | VlTotDebitos |
| 2 | VL_AJ_DEBITOS | VlAjDebitos |
| 3 | VL_TOT_AJ_DEBITOS | VlTotAjDebitos |
| 4 | VL_ESTORNOS_CRED | VlEstornosCred |
| 5 | VL_TOT_CREDITOS | VlTotCreditos |
| 6 | VL_AJ_CREDITOS | VlAjCreditos |
| 7 | VL_TOT_AJ_CREDITOS | VlTotAjCreditos |
| 8 | VL_ESTORNOS_DEB | VlEstornosDeb |
| 9 | VL_SLD_CREDOR_ANT | VlSldCredorAnt |
| 10 | VL_SLD_APURADO | VlSldApurado |
| 11 | VL_TOT_DED | VlTotDed |
| 12 | VL_ICMS_RECOLHER | VlICMSRecolher |
| 13 | VL_SLD_CREDOR_TRANSPORTAR | VlSldCredorTransportar |
| 14 | DEB_ESP | DebEsp |

### E111
| # | ACBr | Go |
|---|------|-----|
| 0 | E111 | E111 |
| 1 | COD_AJ_APUR | CodAjApur |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | VL_AJ_APUR | VlAjApur |

### E112
| # | ACBr | Go |
|---|------|-----|
| 0 | E112 | E112 |
| 1 | NUM_DA | NumDA |
| 2 | NUM_PROC | NumProc |
| 3 | strIND_PROC | IndProc.String() |
| 4 | PROC | Proc |
| 5 | TXT_COMPL | TxtCompl |

### E113 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | E113 | E113 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC | NumDoc |
| 6 | DT_DOC | DtDoc |
| 7 | CHV_NFE | ChvNFe |
| 8 | COD_ITEM | CodItem |
| 9 | VL_AJ_ITEM | VlAjItem |
| 10 | E113 | CodItem |
| 11 | COD_PART | VlAjItem |
| 12 | COD_MOD | ChvNFe |
| 13 | SER | — |
| 14 | SUB | — |
| 15 | NUM_DOC | — |
| 16 | DT_DOC | — |
| 17 | COD_ITEM | — |
| 18 | VL_AJ_ITEM | — |
| 19 | CHV_NFE | — |

### E115
| # | ACBr | Go |
|---|------|-----|
| 0 | E115 | E115 |
| 1 | COD_INF_ADIC | CodInfAdic |
| 2 | VL_INF_ADIC | VlInfAdic |
| 3 | DESCR_COMPL_AJ | DescrComplAj |

### E116 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | E116 | E116 |
| 1 | COD_OR | CodOR |
| 2 | VL_OR | VlOR |
| 3 | DT_VCTO | DtVcto |
| 4 | COD_REC | CodRec |
| 5 | NUM_PROC | NumProc |
| 6 | strIND_PROC | IndProc.String() |
| 7 | PROC | Proc |
| 8 | TXT_COMPL | TxtCompl |
| 9 | — | MesRef |

### E200
| # | ACBr | Go |
|---|------|-----|
| 0 | E200 | E200 |
| 1 | UF | UF |
| 2 | DT_INI | DtIni |
| 3 | DT_FIN | DtFin |

### E210
| # | ACBr | Go |
|---|------|-----|
| 0 | E210 | E210 |
| 1 | Integer(IND_MOV_ST) | IndMovST |
| 2 | VL_SLD_CRED_ANT_ST | VlSldCredAntST |
| 3 | VL_DEVOL_ST | VlDevolST |
| 4 | VL_RESSARC_ST | VlRessarcST |
| 5 | VL_OUT_CRED_ST | VlOutCredST |
| 6 | VL_AJ_CREDITOS_ST | VlAjCreditosST |
| 7 | VL_RETENCAO_ST | VlRetencaoST |
| 8 | VL_OUT_DEB_ST | VlOutDebST |
| 9 | VL_AJ_DEBITOS_ST | VlAjDebitosST |
| 10 | VL_SLD_DEV_ANT_ST | VlSldDevAntST |
| 11 | VL_DEDUCOES_ST | VlDeducoesST |
| 12 | VL_ICMS_RECOL_ST | VlICMSRecolST |
| 13 | VL_SLD_CRED_ST_TRANSPORTAR | VlSldCredSTTransportar |
| 14 | DEB_ESP_ST | DebEspST |

### E220
| # | ACBr | Go |
|---|------|-----|
| 0 | E220 | E220 |
| 1 | COD_AJ_APUR | CodAjApur |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | VL_AJ_APUR | VlAjApur |

### E230
| # | ACBr | Go |
|---|------|-----|
| 0 | E230 | E230 |
| 1 | NUM_DA | NumDA |
| 2 | NUM_PROC | NumProc |
| 3 | intIND_PROC | indProcInt(r.IndProc) |
| 4 | PROC | Proc |
| 5 | TXT_COMPL | TxtCompl |

### E240 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | E240 | E240 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC | NumDoc |
| 6 | DT_DOC | DtDoc |
| 7 | COD_ITEM | CodItem |
| 8 | VL_AJ_ITEM | VlAjItem |
| 9 | — | ChvNFe |

### E250 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | E250 | E250 |
| 1 | COD_OR | CodOR |
| 2 | VL_OR | VlOR |
| 3 | DT_VCTO | DtVcto |
| 4 | COD_REC | CodRec |
| 5 | NUM_PROC | NumProc |
| 6 | strIND_PROC | IndProc.String() |
| 7 | PROC | Proc |
| 8 | TXT_COMPL | TxtCompl |
| 9 | — | MesRef |

### E300
| # | ACBr | Go |
|---|------|-----|
| 0 | E300 | E300 |
| 1 | UF | UF |
| 2 | DT_INI | DtIni |
| 3 | DT_FIN | DtFin |

### E310 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | E310 | E310 |
| 1 | Integer(IND_MOV_DIFAL) | IndMovDIFAL |
| 2 | VL_SLD_CRED_ANT_DIF | VlSldCredAntDIFAL |
| 3 | VL_TOT_DEBITOS_DIFAL | VlTotDebitosDIFAL |
| 4 | VL_OUT_DEB_DIFAL | VlOutDebDIFAL |
| 5 | VL_TOT_DEB_FCP | VlTotDebFCP |
| 6 | VL_TOT_CREDITOS_DIFAL | VlTotCreditosDIFAL |
| 7 | VL_TOT_CRED_FCP | VlTotCredFCP |
| 8 | VL_OUT_CRED_DIFAL | VlOutCredDIFAL |
| 9 | VL_SLD_DEV_ANT_DIFAL | VlSldDevAntDIFAL |
| 10 | VL_DEDUCOES_DIFAL | VlDeducoesDIFAL |
| 11 | VL_RECOL | VlRecol |
| 12 | VL_SLD_CRED_TRANSPORTAR | VlSldCredTransportar |
| 13 | DEB_ESP_DIFAL | DebEspDIFAL |
| 14 | — | VlTotCreditosDIFAL |
| 15 | — | VlOutCredDIFAL |
| 16 | — | VlSldDevAntDIFAL |
| 17 | — | VlDeducoesDIFAL |
| 18 | — | VlRecolDIFAL |
| 19 | — | VlSldCredTranspDIFAL |
| 20 | — | DebEspDIFAL |
| 21 | — | VlSldCredAntFCP |
| 22 | — | VlTotDebFCP |
| 23 | — | VlOutDebFCP |
| 24 | — | VlTotCredFCP |
| 25 | — | VlOutCredFCP |
| 26 | — | VlSldDevAntFCP |
| 27 | — | VlDeducoesFCP |
| 28 | — | VlRecolFCP |
| 29 | — | VlSldCredTranspFCP |
| 30 | — | DebEspFCP |

### E311
| # | ACBr | Go |
|---|------|-----|
| 0 | E311 | E311 |
| 1 | COD_AJ_APUR | CodAjApur |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | VL_AJ_APUR | VlAjApur |

### E312
| # | ACBr | Go |
|---|------|-----|
| 0 | E312 | E312 |
| 1 | NUM_DA | NumDA |
| 2 | NUM_PROC | NumProc |
| 3 | intIND_PROC | indProcInt(r.IndProc) |
| 4 | PROC | Proc |
| 5 | TXT_COMPL | TxtCompl |

### E313
| # | ACBr | Go |
|---|------|-----|
| 0 | E313 | E313 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC | NumDoc |
| 6 | CHV_DOCe | ChvDOCe |
| 7 | DT_DOC | DtDoc |
| 8 | COD_ITEM | CodItem |
| 9 | VL_AJ_ITEM | VlAjItem |

### E316
| # | ACBr | Go |
|---|------|-----|
| 0 | E316 | E316 |
| 1 | COD_OR | CodOR |
| 2 | VL_OR | VlOR |
| 3 | DT_VCTO | DtVcto |
| 4 | COD_REC | CodRec |
| 5 | NUM_PROC | NumProc |
| 6 | strIND_PROC | IndProc.String() |
| 7 | PROC | Proc |
| 8 | TXT_COMPL | TxtCompl |
| 9 | MES_REF | MesRef |

### E500
| # | ACBr | Go |
|---|------|-----|
| 0 | E500 | E500 |
| 1 | Integer(IND_APUR) | ordinal |
| 2 | DT_INI | DtIni |
| 3 | DT_FIN | DtFin |

### E510
| # | ACBr | Go |
|---|------|-----|
| 0 | E510 | E510 |
| 1 | CFOP | CFOP |
| 2 | CST_IPI | CstIPI |
| 3 | VL_CONT_IPI | VlContIPI |
| 4 | VL_BC_IPI | VlBcIPI |
| 5 | VL_IPI | VlIPI |

### E520
| # | ACBr | Go |
|---|------|-----|
| 0 | E520 | E520 |
| 1 | VL_SD_ANT_IPI | VlSdAntIPI |
| 2 | VL_DEB_IPI | VlDebIPI |
| 3 | VL_CRED_IPI | VlCredIPI |
| 4 | VL_OD_IPI | VlOdIPI |
| 5 | VL_OC_IPI | VlOcIPI |
| 6 | VL_SC_IPI | VlScIPI |
| 7 | VL_SD_IPI | VlSdIPI |

### E530
| # | ACBr | Go |
|---|------|-----|
| 0 | E530 | E530 |
| 1 | Integer(IND_AJ) | IndAj |
| 2 | VL_AJ | VlAj |
| 3 | COD_AJ | CodAj |
| 4 | intIND_DOC | indDocInt(r.IndDoc) |
| 5 | NUM_DOC | NumDoc |
| 6 | DESCR_AJ | DescrAj |

### E531
| # | ACBr | Go |
|---|------|-----|
| 0 | E531 | E531 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC | NumDoc |
| 6 | DT_DOC | DtDoc |
| 7 | COD_ITEM | CodItem |
| 8 | VL_AJ_ITEM | VlAjItem |
| 9 | CHV_NFE | ChvNFe |

### E990
| # | ACBr | Go |
|---|------|-----|
| 0 | E990 | E990 |
| 1 | QTD_LIN_E | RegistroE990.QtdLinE |

## Bloco G

### G001
| # | ACBr | Go |
|---|------|-----|
| 0 | G001 | G001 |
| 1 | Integer(IND_MOV) | RegistroG001.IndMov |

### G110 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | G110 | G110 |
| 1 | DT_INI | DtIni |
| 2 | DT_FIN | DtFin |
| 3 | MODO_CIAP | ModoCiap |
| 4 | SALDO_IN_ICMS | SaldoInICMS |
| 5 | SALDO_FN_ICMS | SaldoFnICMS |
| 6 | SOM_PARC | SaldoInICMS |
| 7 | VL_TRIB_EXP | SomParc |
| 8 | VL_TOTAL | VlTribExp |
| 9 | IND_PER_SAI | VlTotal |
| 10 | ICMS_APROP | IndPerSai |
| 11 | SOM_ICMS_OC | ICMSAprop |
| 12 | — | SomICMSOC |

### G125
| # | ACBr | Go |
|---|------|-----|
| 0 | G125 | G125 |
| 1 | COD_IND_BEM | CodIndBem |
| 2 | DT_MOV | DtMov |
| 3 | strTIPO_MOV | TipoMov.String() |
| 4 | VL_IMOB_ICMS_OP | VlImobICMSOp |
| 5 | VL_IMOB_ICMS_ST | VlImobICMSST |
| 6 | VL_IMOB_ICMS_FRT | VlImobICMSFrt |
| 7 | VL_IMOB_ICMS_DIF | VlImobICMSDif |
| 8 | NUM_PARC | NumParc |
| 9 | VL_PARC_PASS | VlParcPass |
| 10 | VL_PARC_APROP | VlParcAprop |

### G126
| # | ACBr | Go |
|---|------|-----|
| 0 | G126 | G126 |
| 1 | DT_INI | DtIni |
| 2 | DT_FIN | DtFin |
| 3 | NUM_PARC | NumParc |
| 4 | VL_PARC_PASS | VlParcPass |
| 5 | VL_TRIB_OC | VlTribOC |
| 6 | VL_TOTAL | VlTotal |
| 7 | IND_PER_SAI | IndPerSai |
| 8 | VL_PARC_APROP | VlParcAprop |

### G130
| # | ACBr | Go |
|---|------|-----|
| 0 | G130 | G130 |
| 1 | Integer(IND_EMIT) | IndEmit |
| 2 | COD_PART | CodPart |
| 3 | COD_MOD | CodMod |
| 4 | SERIE | Serie |
| 5 | NUM_DOC | NumDoc |
| 6 | CHV_NFE_CTE | ChvNFeCTe |
| 7 | DT_DOC | DtDoc |
| 8 | NUM_DA | NumDA |

### G140
| # | ACBr | Go |
|---|------|-----|
| 0 | G140 | G140 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | QTDE | Qtde |
| 4 | UNID | Unid |
| 5 | VL_ICMS_OP_APLICADO | VlICMSOpAplicado |
| 6 | VL_ICMS_ST_APLICADO | VlICMSSTAplicado |
| 7 | VL_ICMS_FRT_APLICADO | VlICMSFrtAplicado |
| 8 | VL_ICMS_DIF_APLICADO | VlICMSDifAplicado |

### G990
| # | ACBr | Go |
|---|------|-----|
| 0 | G990 | G990 |
| 1 | QTD_LIN_G | RegistroG990.QtdLinG |

## Bloco H

### H001
| # | ACBr | Go |
|---|------|-----|
| 0 | H001 | H001 |
| 1 | Integer(IND_MOV) | RegistroH001.IndMov |

### H005
| # | ACBr | Go |
|---|------|-----|
| 0 | H005 | H005 |
| 1 | DT_INV | DtInv |
| 2 | VL_INV | VlInv |
| 3 | strMotInv | MotInv.String() |

### H010
| # | ACBr | Go |
|---|------|-----|
| 0 | H010 | H010 |
| 1 | COD_ITEM | CodItem |
| 2 | UNID | Unid |
| 3 | QTD | Qtd |
| 4 | VL_UNIT | VlUnit |
| 5 | VL_ITEM | VlItem |
| 6 | Integer(IND_PROP) | IndProp.String() |
| 7 | COD_PART | CodPart |
| 8 | TXT_COMPL | TxtCompl |
| 9 | COD_CTA | CodCta |
| 10 | VL_ITEM_IR | VlItemIR |

### H011
| # | ACBr | Go |
|---|------|-----|
| 0 | H011 | H011 |
| 1 | CNPJ | CNPJ |

### H020
| # | ACBr | Go |
|---|------|-----|
| 0 | H020 | H020 |
| 1 | CST_ICMS | CstICMS.String() |
| 2 | BC_ICMS | BcICMS |
| 3 | VL_ICMS | VlICMS |

### H030
| # | ACBr | Go |
|---|------|-----|
| 0 | H030 | H030 |
| 1 | VL_ICMS_OP | VlICMSOp |
| 2 | VL_BC_ICMS_ST | VlBcICMSST |
| 3 | VL_ICMS_ST | VlICMSST |
| 4 | VL_FCP | VlFCP |

### H990
| # | ACBr | Go |
|---|------|-----|
| 0 | H990 | H990 |
| 1 | QTD_LIN_H | RegistroH990.QtdLinH |

## Bloco K

### K001
| # | ACBr | Go |
|---|------|-----|
| 0 | K001 | K001 |
| 1 | Integer(IND_MOV) | RegistroK001.IndMov |

### K010
| # | ACBr | Go |
|---|------|-----|
| 0 | K010 | K010 |
| 1 | Integer(IND_TIPO_LEIAUTE) | IndTipoLeiaute |

### K100
| # | ACBr | Go |
|---|------|-----|
| 0 | K100 | K100 |
| 1 | DT_INI | DtIni |
| 2 | DT_FIN | DtFin |

### K200
| # | ACBr | Go |
|---|------|-----|
| 0 | K200 | K200 |
| 1 | DT_EST | DtEst |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | Integer(IND_EST) | IndEst |
| 5 | COD_PART | CodPart |

### K210
| # | ACBr | Go |
|---|------|-----|
| 0 | K210 | K210 |
| 1 | DT_INI_OS | DtIniOS |
| 2 | DT_FIN_OS | DtFinOS |
| 3 | COD_DOC_OS | CodDocOS |
| 4 | COD_ITEM_ORI | CodItemOri |
| 5 | QTD_ORI | QtdOri |

### K215
| # | ACBr | Go |
|---|------|-----|
| 0 | K215 | K215 |
| 1 | COD_ITEM_DES | CodItemDes |
| 2 | QTD_DES | QtdDes |

### K220
| # | ACBr | Go |
|---|------|-----|
| 0 | K220 | K220 |
| 1 | DT_MOV | DtMov |
| 2 | COD_ITEM_ORI | CodItemOri |
| 3 | COD_ITEM_DEST | CodItemDest |
| 4 | QTD | Qtd |
| 5 | QTD_DEST | QtdDest |

### K230
| # | ACBr | Go |
|---|------|-----|
| 0 | K230 | K230 |
| 1 | DT_INI_OP | DtIniOP |
| 2 | DT_FIN_OP | DtFinOP |
| 3 | COD_DOC_OP | CodDocOP |
| 4 | COD_ITEM | CodItem |
| 5 | QTD_ENC | QtdEnc |

### K235
| # | ACBr | Go |
|---|------|-----|
| 0 | K235 | K235 |
| 1 | RegK230.RegistroK235.Items[intFor].DT_SAIDA | DtSaida |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | COD_INS_SUBST | CodInsSubst |

### K250
| # | ACBr | Go |
|---|------|-----|
| 0 | K250 | K250 |
| 1 | DT_PROD | DtProd |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |

### K255
| # | ACBr | Go |
|---|------|-----|
| 0 | K255 | K255 |
| 1 | DT_CONS | DtCons |
| 2 | COD_ITEM | CodItem |
| 3 | QTD | Qtd |
| 4 | COD_INS_SUBST | CodInsSubst |

### K260
| # | ACBr | Go |
|---|------|-----|
| 0 | K260 | K260 |
| 1 | COD_OP_OS | CodOpOS |
| 2 | COD_ITEM | CodItem |
| 3 | DT_SAIDA | DtSaida |
| 4 | QTD_SAIDA | QtdSaida |
| 5 | DT_RET | DtRet |
| 6 | QTD_RET | QtdRet |

### K265
| # | ACBr | Go |
|---|------|-----|
| 0 | K265 | K265 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD_CONS | QtdCons |
| 3 | QTD_RET | QtdRet |

### K270
| # | ACBr | Go |
|---|------|-----|
| 0 | K270 | K270 |
| 1 | DT_INI_AP | DtIniAP |
| 2 | DT_FIN_AP | DtFinAP |
| 3 | COD_OP_OS | CodOpOS |
| 4 | COD_ITEM | CodItem |
| 5 | QTD_COR_POS | QtdCorPos |
| 6 | QTD_COR_NEG | QtdCorNeg |
| 7 | ORIGEM | Origem |

### K275
| # | ACBr | Go |
|---|------|-----|
| 0 | K275 | K275 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD_COR_POS | QtdCorPos |
| 3 | QTD_COR_NEG | QtdCorNeg |
| 4 | COD_INS_SUBST | CodInsSubst |

### K280
| # | ACBr | Go |
|---|------|-----|
| 0 | K280 | K280 |
| 1 | DT_EST | DtEst |
| 2 | COD_ITEM | CodItem |
| 3 | QTD_COR_POS | QtdCorPos |
| 4 | QTD_COR_NEG | QtdCorNeg |
| 5 | Integer(IND_EST) | IndEst |
| 6 | COD_PART | CodPart |

### K290
| # | ACBr | Go |
|---|------|-----|
| 0 | K290 | K290 |
| 1 | DT_INI_OP | DtIniOP |
| 2 | DT_FIN_OP | DtFinOP |
| 3 | COD_DOC_OP | CodDocOP |

### K291
| # | ACBr | Go |
|---|------|-----|
| 0 | K291 | K291 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |

### K292
| # | ACBr | Go |
|---|------|-----|
| 0 | K292 | K292 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |

### K300
| # | ACBr | Go |
|---|------|-----|
| 0 | K300 | K300 |
| 1 | DT_PROD | DtProd |

### K301
| # | ACBr | Go |
|---|------|-----|
| 0 | K301 | K301 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |

### K302
| # | ACBr | Go |
|---|------|-----|
| 0 | K302 | K302 |
| 1 | COD_ITEM | CodItem |
| 2 | QTD | Qtd |

### K990
| # | ACBr | Go |
|---|------|-----|
| 0 | K990 | K990 |
| 1 | QTD_LIN_K | RegistroK990.QtdLinK |

## Bloco 1

### 1001
| # | ACBr | Go |
|---|------|-----|
| 0 | 1001 | 1001 |
| 1 | Integer(IND_MOV) | Registro1001.IndMov |

### 1010
| # | ACBr | Go |
|---|------|-----|
| 0 | 1010 | 1010 |
| 1 | IND_EXP | IndExp |
| 2 | IND_CCRF | IndCCRF |
| 3 | IND_COMB | IndComb |
| 4 | IND_USINA | IndUsina |
| 5 | IND_VA | IndVA |
| 6 | IND_EE | IndEE |
| 7 | IND_CART | IndCart |
| 8 | IND_FORM | IndForm |
| 9 | IND_AER | IndAer |
| 10 | IND_GIAF1 | IndGIAF1 |
| 11 | IND_GIAF3 | IndGIAF3 |
| 12 | IND_GIAF4 | IndGIAF4 |
| 13 | IND_REST_RESSARC_COMPL_ICMS | IndRestRessarcComplICMS |

### 1100
| # | ACBr | Go |
|---|------|-----|
| 0 | 1100 | 1100 |
| 1 | Integer(IND_DOC) | IndDoc |
| 2 | NRO_DE | NroDE |
| 3 | DT_DE | DtDE |
| 4 | Integer(NAT_EXP) | NatExp |
| 5 | NRO_RE | NroRE |
| 6 | DT_RE | DtRE |
| 7 | CHC_EMB | ChcEmb |
| 8 | DT_CHC | DtChc |
| 9 | DT_AVB | DtAvb |
| 10 | strTP_CHC | TpChc.String() |
| 11 | PAIS | Pais |

### 1105
| # | ACBr | Go |
|---|------|-----|
| 0 | 1105 | 1105 |
| 1 | COD_MOD | CodMod |
| 2 | SERIE | Serie |
| 3 | NUM_DOC | NumDoc |
| 4 | CHV_NFE | ChvNFe |
| 5 | DT_DOC | DtDoc |
| 6 | COD_ITEM | CodItem |

### 1110
| # | ACBr | Go |
|---|------|-----|
| 0 | 1110 | 1110 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | NUM_DOC | NumDoc |
| 5 | DT_DOC | DtDoc |
| 6 | CHV_NFE | ChvNFe |
| 7 | NR_MEMO | NrMemo |
| 8 | QTD | Qtd |
| 9 | UNID | Unid |

### 1200
| # | ACBr | Go |
|---|------|-----|
| 0 | 1200 | 1200 |
| 1 | COD_AJ_APUR | CodAjApur |
| 2 | SLD_CRED | SldCred |
| 3 | CRED_APR | CredApr |
| 4 | CRED_RECEB | CredReceb |
| 5 | CRED_UTIL | CredUtil |
| 6 | SLD_CRED_FIM | SldCredFim |

### 1210
| # | ACBr | Go |
|---|------|-----|
| 0 | 1210 | 1210 |
| 1 | TIPO_UTIL | TipoUtil |
| 2 | NR_DOC | NrDoc |
| 3 | VL_CRED_UTIL | VlCredUtil |
| 4 | CHV_DOCe | ChvDOCe |

### 1250
| # | ACBr | Go |
|---|------|-----|
| 0 | 1250 | 1250 |
| 1 | VL_CREDITO_ICMS_OP | VlCreditoICMSOp |
| 2 | VL_ICMS_ST_REST | VlICMSSTRest |
| 3 | VL_FCP_ST_REST | VlFCPSTRest |
| 4 | VL_ICMS_ST_COMPL | VlICMSSTCompl |
| 5 | VL_FCP_ST_COMPL | VlFCPSTCompl |

### 1255
| # | ACBr | Go |
|---|------|-----|
| 0 | 1255 | 1255 |
| 1 | COD_MOT_REST_COMPL | CodMotRestCompl |
| 2 | VL_CREDITO_ICMS_OP_MOT | VlCreditoICMSOpMot |
| 3 | VL_ICMS_ST_REST_MOT | VlICMSSTRestMot |
| 4 | VL_FCP_ST_REST_MOT | VlFCPSTRestMot |
| 5 | VL_ICMS_ST_COMPL_MOT | VlICMSSTComplMot |
| 6 | VL_FCP_ST_COMPL_MOT | VlFCPSTComplMot |

### 1300
| # | ACBr | Go |
|---|------|-----|
| 0 | 1300 | 1300 |
| 1 | COD_ITEM | CodItem |
| 2 | DT_FECH | DtFech |
| 3 | ESTQ_ABERT | EstqAbert |
| 4 | VOL_ENTR | VolEntr |
| 5 | VOL_DISP | VolDisp |
| 6 | VOL_SAIDAS | VolSaidas |
| 7 | ESTQ_ESCR | EstqEscr |
| 8 | VAL_AJ_PERDA | ValAjPerda |
| 9 | VAL_AJ_GANHO | ValAjGanho |
| 10 | FECH_FISICO | FechFisico |

### 1310 (ramos por versao, veja a nota do topo)
| # | ACBr | Go |
|---|------|-----|
| 0 | 1310 | 1310 |
| 1 | NUM_TANQUE | NumTanque |
| 2 | ESTQ_ABERT | EstqAbert |
| 3 | VOL_ENTR | VolEntr |
| 4 | VOL_DISP | VolDisp |
| 5 | VOL_SAIDAS | VolSaidas |
| 6 | ESTQ_ESCR | EstqEscr |
| 7 | VAL_AJ_PERDA | ValAjPerda |
| 8 | VAL_AJ_GANHO | ValAjGanho |
| 9 | FECH_FISICO | FechFisico |
| 10 | — | CapTanque |

### 1320
| # | ACBr | Go |
|---|------|-----|
| 0 | 1320 | 1320 |
| 1 | NUM_BICO | NumBico |
| 2 | NR_INTERV | NrInterv |
| 3 | MOT_INTERV | MotInterv |
| 4 | NOM_INTERV | NomInterv |
| 5 | CNPJ_INTERV | CnpjInterv |
| 6 | CPF_INTERV | CpfInterv |
| 7 | VAL_FECHA | ValFecha |
| 8 | VAL_ABERT | ValAbert |
| 9 | VOL_AFERI | VolAferi |
| 10 | VOL_VENDAS | VolVendas |

### 1350
| # | ACBr | Go |
|---|------|-----|
| 0 | 1350 | 1350 |
| 1 | SERIE | Serie |
| 2 | FABRICANTE | Fabricante |
| 3 | MODELO | Modelo |
| 4 | Integer(TIPO_MEDICAO) | TipoMedicao |

### 1360
| # | ACBr | Go |
|---|------|-----|
| 0 | 1360 | 1360 |
| 1 | NUM_LACRE | NumLacre |
| 2 | DT_APLICACAO | DtAplicacao |

### 1370
| # | ACBr | Go |
|---|------|-----|
| 0 | 1370 | 1370 |
| 1 | NUM_BICO | NumBico |
| 2 | COD_ITEM | CodItem |
| 3 | NUM_TANQUE | NumTanque |

### 1390
| # | ACBr | Go |
|---|------|-----|
| 0 | 1390 | 1390 |
| 1 | COD_PROD | CodProd |

### 1391
| # | ACBr | Go |
|---|------|-----|
| 0 | 1391 | 1391 |
| 1 | vReg1391.DT_REGISTRO | DtRegistro |
| 2 | vReg1391.QTD_MOID | QtdMoid |
| 3 | vReg1391.ESTQ_INI | EstqIni |
| 4 | vReg1391.QTD_PRODUZ | QtdProduz |
| 5 | vReg1391.ENT_ANID_HID | EntAnidHid |
| 6 | vReg1391.OUTR_ENTR | OutrEntr |
| 7 | vReg1391.PERDA | Perda |
| 8 | vReg1391.CONS | Cons |
| 9 | vReg1391.SAI_ANI_HID | SaiAniHid |
| 10 | vReg1391.SAIDAS | Saidas |
| 11 | vReg1391.ESTQ_FIN | EstqFin |
| 12 | vReg1391.ESTQ_INI_MEL | EstqIniMel |
| 13 | vReg1391.PROD_DIA_MEL | ProdDiaMel |
| 14 | vReg1391.UTIL_MEL | UtilMel |
| 15 | vReg1391.PROD_ALC_MEL | ProdAlcMel |
| 16 | vReg1391.OBS | Obs |
| 17 | vReg1391.COD_ITEM | CodItem |
| 18 | vReg1391.TP_RESIDUO | tpResiduoComoData(r.TpResiduo) |
| 19 | vReg1391.QTD_RESIDUO | QtdResiduo |
| 20 | vReg1391.QTD_RESIDUO_DDG | QtdResiduoDDG |
| 21 | vReg1391.QTD_RESIDUO_WDG | QtdResiduoWDG |
| 22 | vReg1391.QTD_RESIDUO_CANA | QtdResiduoCana |

### 1400
| # | ACBr | Go |
|---|------|-----|
| 0 | 1400 | 1400 |
| 1 | vCodItem | codItem |
| 2 | Trim(vReg1400.MUN) | strings.TrimSpace(r.Mun) |
| 3 | vReg1400.VALOR | Valor |

### 1500
| # | ACBr | Go |
|---|------|-----|
| 0 | 1500 | 1500 |
| 1 | IND_OPER | IndOper |
| 2 | IND_EMIT | IndEmit |
| 3 | COD_PART | CodPart |
| 4 | COD_MOD | CodMod |
| 5 | strCOD_SIT | CodSit.String() |
| 6 | SER | Ser |
| 7 | SUB | Sub |
| 8 | strCOD_CONS | CodCons.String() |
| 9 | NUM_DOC | NumDoc |
| 10 | DT_DOC | DtDoc |
| 11 | DT_E_S | DtES |
| 12 | VL_DESC | VlDesc |
| 13 | VL_DESC | VlDesc |
| 14 | VL_FORN | VlForn |
| 15 | VL_SERV_NT | VlServNT |
| 16 | VL_TERC | VlTerc |
| 17 | VL_DA | VlDa |
| 18 | VL_BC_ICMS | VlBcICMS |
| 19 | VL_ICMS | VlICMS |
| 20 | VL_BC_ICMS_ST | VlBcICMSST |
| 21 | VL_ICMS_ST | VlICMSST |
| 22 | COD_INF | CodInf |
| 23 | VL_PIS | VlPIS |
| 24 | VL_COFINS | VlCOFINS |
| 25 | intTP_LIGACAO | tpLigacaoInt(r.TpLigacao) |
| 26 | strCOD_GRUPO_TENSAO | CodGrupoTensao.String() |

### 1510
| # | ACBr | Go |
|---|------|-----|
| 0 | 1510 | 1510 |
| 1 | NUM_ITEM | NumItem |
| 2 | COD_ITEM | CodItem |
| 3 | COD_CLASS | CodClass |
| 4 | QTD | Qtd |
| 5 | UNID | Unid |
| 6 | VL_ITEM | VlItem |
| 7 | VL_DESC | VlDesc |
| 8 | CST_ICMS | CstICMS |
| 9 | CFOP | CFOP |
| 10 | VL_BC_ICMS | VlBcICMS |
| 11 | ALIQ_ICMS | AliqICMS |
| 12 | VL_ICMS | VlICMS |
| 13 | VL_BC_ICMS_ST | VlBcICMSST |
| 14 | ALIQ_ST | AliqST |
| 15 | VL_ICMS_ST | VlICMSST |
| 16 | Integer(IND_REC) | IndRec |
| 17 | COD_PART | CodPart |
| 18 | VL_PIS | VlPIS |
| 19 | VL_COFINS | VlCOFINS |
| 20 | COD_CTA | CodCta |

### 1600
| # | ACBr | Go |
|---|------|-----|
| 0 | 1600 | 1600 |
| 1 | COD_PART | CodPart |
| 2 | TOT_CREDITO | TotCredito |
| 3 | TOT_DEBITO | TotDebito |

### 1601
| # | ACBr | Go |
|---|------|-----|
| 0 | 1601 | 1601 |
| 1 | COD_PART_IP | CodPartIP |
| 2 | COD_PART_IT | CodPartIT |
| 3 | TOT_VS | TotVS |
| 4 | TOT_ISS | TotISS |
| 5 | TOT_OUTROS | TotOutros |

### 1700
| # | ACBr | Go |
|---|------|-----|
| 0 | 1700 | 1700 |
| 1 | strCOD_DISP | CodDisp.String() |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC_INI | NumDocIni |
| 6 | NUM_DOC_FIN | NumDocFin |
| 7 | NUM_AUT | NumAut |

### 1710
| # | ACBr | Go |
|---|------|-----|
| 0 | 1710 | 1710 |
| 1 | NUM_DOC_INI | NumDocIni |
| 2 | NUM_DOC_FIN | NumDocFin |

### 1800
| # | ACBr | Go |
|---|------|-----|
| 0 | 1800 | 1800 |
| 1 | VL_CARGA | VlCarga |
| 2 | VL_PASS | VlPass |
| 3 | VL_FAT | VlFat |
| 4 | IND_RAT | IndRat |
| 5 | VL_ICMS_ANT | VlICMSAnt |
| 6 | VL_BC_ICMS | VlBcICMS |
| 7 | VL_ICMS_APUR | VlICMSApur |
| 8 | VL_BC_ICMS_APUR | VlBcICMSApur |
| 9 | VL_DIF | VlDif |

### 1900
| # | ACBr | Go |
|---|------|-----|
| 0 | 1900 | 1900 |
| 1 | IND_APUR_ICMS | IndApurICMS |
| 2 | DESCR_COMPL_OUT_APUR | DescrComplOutApur |

### 1910
| # | ACBr | Go |
|---|------|-----|
| 0 | 1910 | 1910 |
| 1 | DT_INI | DtIni |
| 2 | DT_FIN | DtFin |

### 1920
| # | ACBr | Go |
|---|------|-----|
| 0 | 1920 | 1920 |
| 1 | VL_TOT_TRANSF_DEBITOS_OA | VlTotTransfDebitosOA |
| 2 | VL_TOT_AJ_DEBITOS_OA | VlTotAjDebitosOA |
| 3 | VL_ESTORNOS_CRED_OA | VlEstornosCredOA |
| 4 | VL_TOT_TRANSF_CREDITOS_OA | VlTotTransfCreditosOA |
| 5 | VL_TOT_AJ_CREDITOS_OA | VlTotAjCreditosOA |
| 6 | VL_ESTORNOS_DEB_OA | VlEstornosDebOA |
| 7 | VL_SLD_CREDOR_ANT_OA | VlSldCredorAntOA |
| 8 | VL_SLD_APURADO_OA | VlSldApuradoOA |
| 9 | VL_TOT_DED | VlTotDed |
| 10 | VL_ICMS_RECOLHER_OA | VlICMSRecolherOA |
| 11 | VL_SLD_CREDOR_TRANSP_OA | VlSldCredorTranspOA |
| 12 | DEB_ESP_OA | DebEspOA |

### 1921
| # | ACBr | Go |
|---|------|-----|
| 0 | 1921 | 1921 |
| 1 | COD_AJ_APUR | CodAjApur |
| 2 | DESCR_COMPL_AJ | DescrComplAj |
| 3 | VL_AJ_APUR | VlAjApur |

### 1922
| # | ACBr | Go |
|---|------|-----|
| 0 | 1922 | 1922 |
| 1 | NUM_DA | NumDA |
| 2 | NUM_PROC | NumProc |
| 3 | IND_PROC | IndProc |
| 4 | PROC | Proc |
| 5 | TXT_COMPL | TxtCompl |

### 1923
| # | ACBr | Go |
|---|------|-----|
| 0 | 1923 | 1923 |
| 1 | COD_PART | CodPart |
| 2 | COD_MOD | CodMod |
| 3 | SER | Ser |
| 4 | SUB | Sub |
| 5 | NUM_DOC | NumDoc |
| 6 | DT_DOC | DtDoc |
| 7 | COD_ITEM | CodItem |
| 8 | VL_AJ_ITEM | VlAjItem |
| 9 | CHV_DOCe | ChvDOCe |

### 1925
| # | ACBr | Go |
|---|------|-----|
| 0 | 1925 | 1925 |
| 1 | COD_INF_ADIC | CodInfAdic |
| 2 | VL_INF_ADIC | VlInfAdic |
| 3 | DESCR_COMPL_AJ | DescrComplAj |

### 1926
| # | ACBr | Go |
|---|------|-----|
| 0 | 1926 | 1926 |
| 1 | COD_OR | CodOR |
| 2 | VL_OR | VlOR |
| 3 | DT_VCTO | DtVcto |
| 4 | COD_REC | CodRec |
| 5 | NUM_PROC | NumProc |
| 6 | IND_PROC | IndProc |
| 7 | PROC | Proc |
| 8 | TXT_COMPL | TxtCompl |
| 9 | MES_REF | MesRef |

### 1960
| # | ACBr | Go |
|---|------|-----|
| 0 | 1960 | 1960 |
| 1 | IND_AP | IndAp |
| 2 | G1_01 | G1_01 |
| 3 | G1_02 | G1_02 |
| 4 | G1_03 | G1_03 |
| 5 | G1_04 | G1_04 |
| 6 | G1_05 | G1_05 |
| 7 | G1_06 | G1_06 |
| 8 | G1_07 | G1_07 |
| 9 | G1_08 | G1_08 |
| 10 | G1_09 | G1_09 |
| 11 | G1_10 | G1_10 |
| 12 | G1_11 | G1_11 |

### 1970
| # | ACBr | Go |
|---|------|-----|
| 0 | 1970 | 1970 |
| 1 | IND_AP | IndAp |
| 2 | G3_01 | G3_01 |
| 3 | G3_02 | G3_02 |
| 4 | G3_03 | G3_03 |
| 5 | G3_04 | G3_04 |
| 6 | G3_05 | G3_05 |
| 7 | G3_06 | G3_06 |
| 8 | G3_07 | G3_07 |
| 9 | G3_T | G3_T |
| 10 | G3_08 | G3_08 |
| 11 | G3_09 | G3_09 |

### 1975
| # | ACBr | Go |
|---|------|-----|
| 0 | 1975 | 1975 |
| 1 | ALIQ_IMP_BASE | AliqImpBase |
| 2 | G3_10 | G3_10 |
| 3 | G3_11 | G3_11 |
| 4 | G3_12 | G3_12 |

### 1980
| # | ACBr | Go |
|---|------|-----|
| 0 | 1980 | 1980 |
| 1 | IND_AP | IndAp |
| 2 | G4_01 | G4_01 |
| 3 | G4_02 | G4_02 |
| 4 | G4_03 | G4_03 |
| 5 | G4_04 | G4_04 |
| 6 | G4_05 | G4_05 |
| 7 | G4_06 | G4_06 |
| 8 | G4_07 | G4_07 |
| 9 | G4_08 | G4_08 |
| 10 | G4_09 | G4_09 |
| 11 | G4_10 | G4_10 |
| 12 | G4_11 | G4_11 |
| 13 | G4_12 | G4_12 |

### 1990
| # | ACBr | Go |
|---|------|-----|
| 0 | 1990 | 1990 |
| 1 | QTD_LIN_1 | Registro1990.QtdLin1 |

## Bloco 9

### 9001
| # | ACBr | Go |
|---|------|-----|
| 0 | 9001 | 9001 |
| 1 | Integer(IND_MOV) | Registro9001.IndMov |

### 9900
| # | ACBr | Go |
|---|------|-----|
| 0 | 9900 | 9900 |
| 1 | REG_BLC | RegBlc |
| 2 | QTD_REG_BLC | QtdRegBlc |

### 9990
| # | ACBr | Go |
|---|------|-----|
| 0 | 9990 | 9990 |
| 1 | QTD_LIN_9 | Registro9990.QtdLin9 |

### 9999
| # | ACBr | Go |
|---|------|-----|
| 0 | 9999 | 9999 |
| 1 | QTD_LIN | Registro9999.QtdLin |

