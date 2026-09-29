// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-28.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package sped

import "fmt"

// ---------------------------------------------------------------------------
// CstIcms - Codigo da Situacao Tributaria referente ao ICMS
// Portado de: TACBrCstIcms (ACBrEFDBlocos.pas)
// ---------------------------------------------------------------------------

// CstIcms representa o Codigo da Situacao Tributaria do ICMS.
type CstIcms int

const (
	CstIcmsNenhum                                                              CstIcms = iota // Nenhum (vazio)
	CstIcmsTributadaIntegralmente                                                             // 000 - Tributada integralmente
	CstIcmsTributadaComCobracaPorST                                                           // 010 - Tributada e com cobranca do ICMS por substituicao tributaria
	CstIcmsComReducao                                                                         // 020 - Com reducao de base de calculo
	CstIcmsIsentaComCobracaPorST                                                              // 030 - Isenta ou nao tributada e com cobranca do ICMS por substituicao tributaria
	CstIcmsIsenta                                                                             // 040 - Isenta
	CstIcmsNaoTributada                                                                       // 041 - Nao tributada
	CstIcmsSuspensao                                                                          // 050 - Suspensao
	CstIcmsDiferimento                                                                        // 051 - Diferimento
	CstIcmsCobradoAnteriormentePorST                                                          // 060 - ICMS cobrado anteriormente por substituicao tributaria
	CstIcmsComReducaoPorST                                                                    // 070 - Com reducao de base de calculo e cobranca do ICMS por substituicao tributaria
	CstIcmsOutros                                                                             // 090 - Outros
	CstIcmsEstrangeiraImportacaoDiretaTributadaIntegralmente                                  // 100 - Estrangeira - Importacao direta - Tributada integralmente
	CstIcmsEstrangeiraImportacaoDiretaTributadaComCobracaPorST                                // 110 - Estrangeira - Importacao direta - Tributada e com cobranca do ICMS por ST
	CstIcmsEstrangeiraImportacaoDiretaComReducao                                              // 120 - Estrangeira - Importacao direta - Com reducao de base de calculo
	CstIcmsEstrangeiraImportacaoDiretaIsentaComCobracaPorST                                   // 130 - Estrangeira - Importacao direta - Isenta ou nao tributada e com cobranca do ICMS por ST
	CstIcmsEstrangeiraImportacaoDiretaIsenta                                                  // 140 - Estrangeira - Importacao direta - Isenta
	CstIcmsEstrangeiraImportacaoDiretaNaoTributada                                            // 141 - Estrangeira - Importacao direta - Nao tributada
	CstIcmsEstrangeiraImportacaoDiretaSuspensao                                               // 150 - Estrangeira - Importacao direta - Suspensao
	CstIcmsEstrangeiraImportacaoDiretaDiferimento                                             // 151 - Estrangeira - Importacao direta - Diferimento
	CstIcmsEstrangeiraImportacaoDiretaCobradoAnteriormentePorST                               // 160 - Estrangeira - Importacao direta - ICMS cobrado anteriormente por ST
	CstIcmsEstrangeiraImportacaoDiretaComReducaoPorST                                         // 170 - Estrangeira - Importacao direta - Com reducao de base de calculo e cobranca do ICMS por ST
	CstIcmsEstrangeiraImportacaoDiretaOutros                                                  // 190 - Estrangeira - Importacao direta - Outras
	CstIcmsEstrangeiraAdqMercIntTributadaIntegralmente                                        // 200 - Estrangeira - Adquirida no mercado interno - Tributada integralmente
	CstIcmsEstrangeiraAdqMercIntTributadaComCobracaPorST                                      // 210 - Estrangeira - Adquirida no mercado interno - Tributada e com cobranca do ICMS por ST
	CstIcmsEstrangeiraAdqMercIntComReducao                                                    // 220 - Estrangeira - Adquirida no mercado interno - Com reducao de base de calculo
	CstIcmsEstrangeiraAdqMercIntIsentaComCobracaPorST                                         // 230 - Estrangeira - Adquirida no mercado interno - Isenta ou nao tributada e com cobranca do ICMS por ST
	CstIcmsEstrangeiraAdqMercIntIsenta                                                        // 240 - Estrangeira - Adquirida no mercado interno - Isenta
	CstIcmsEstrangeiraAdqMercIntNaoTributada                                                  // 241 - Estrangeira - Adquirida no mercado interno - Nao tributada
	CstIcmsEstrangeiraAdqMercIntSuspensao                                                     // 250 - Estrangeira - Adquirida no mercado interno - Suspensao
	CstIcmsEstrangeiraAdqMercIntDiferimento                                                   // 251 - Estrangeira - Adquirida no mercado interno - Diferimento
	CstIcmsEstrangeiraAdqMercIntCobradoAnteriormentePorST                                     // 260 - Estrangeira - Adquirida no mercado interno - ICMS cobrado anteriormente por ST
	CstIcmsEstrangeiraAdqMercIntComReducaoPorST                                               // 270 - Estrangeira - Adquirida no mercado interno - Com reducao de base de calculo e cobranca do ICMS por ST
	CstIcmsEstrangeiraAdqMercIntOutros                                                        // 290 - Estrangeira - Adquirida no mercado interno - Outras
	CstIcms300                                                                                // 300
	CstIcms310                                                                                // 310
	CstIcms320                                                                                // 320
	CstIcms330                                                                                // 330
	CstIcms340                                                                                // 340
	CstIcms341                                                                                // 341
	CstIcms350                                                                                // 350
	CstIcms351                                                                                // 351
	CstIcms360                                                                                // 360
	CstIcms370                                                                                // 370
	CstIcms390                                                                                // 390
	CstIcms400                                                                                // 400
	CstIcms410                                                                                // 410
	CstIcms420                                                                                // 420
	CstIcms430                                                                                // 430
	CstIcms440                                                                                // 440
	CstIcms441                                                                                // 441
	CstIcms450                                                                                // 450
	CstIcms451                                                                                // 451
	CstIcms460                                                                                // 460
	CstIcms470                                                                                // 470
	CstIcms490                                                                                // 490
	CstIcms500                                                                                // 500
	CstIcms510                                                                                // 510
	CstIcms520                                                                                // 520
	CstIcms530                                                                                // 530
	CstIcms540                                                                                // 540
	CstIcms541                                                                                // 541
	CstIcms550                                                                                // 550
	CstIcms551                                                                                // 551
	CstIcms560                                                                                // 560
	CstIcms570                                                                                // 570
	CstIcms590                                                                                // 590
	CstIcms600                                                                                // 600
	CstIcms610                                                                                // 610
	CstIcms620                                                                                // 620
	CstIcms630                                                                                // 630
	CstIcms640                                                                                // 640
	CstIcms641                                                                                // 641
	CstIcms650                                                                                // 650
	CstIcms651                                                                                // 651
	CstIcms660                                                                                // 660
	CstIcms670                                                                                // 670
	CstIcms690                                                                                // 690
	CstIcms700                                                                                // 700
	CstIcms710                                                                                // 710
	CstIcms720                                                                                // 720
	CstIcms730                                                                                // 730
	CstIcms740                                                                                // 740
	CstIcms741                                                                                // 741
	CstIcms750                                                                                // 750
	CstIcms751                                                                                // 751
	CstIcms760                                                                                // 760
	CstIcms770                                                                                // 770
	CstIcms790                                                                                // 790
	CstIcms800                                                                                // 800
	CstIcms810                                                                                // 810
	CstIcms820                                                                                // 820
	CstIcms830                                                                                // 830
	CstIcms840                                                                                // 840
	CstIcms841                                                                                // 841
	CstIcms850                                                                                // 850
	CstIcms851                                                                                // 851
	CstIcms860                                                                                // 860
	CstIcms870                                                                                // 870
	CstIcms890                                                                                // 890
	CstIcmsSimplesNacionalTributadaComPermissaoCredito                                        // 101 - Simples Nacional - Tributada com permissao de credito
	CstIcmsSimplesNacionalTributadaSemPermissaoCredito                                        // 102 - Simples Nacional - Tributada sem permissao de credito
	CstIcmsSimplesNacionalIsencaoPorFaixaReceitaBruta                                         // 103 - Simples Nacional - Isencao do ICMS para faixa de receita bruta
	CstIcmsSimplesNacionalTributadaComPermissaoCreditoComST                                   // 201 - Simples Nacional - Tributada com permissao de credito e com cobranca do ICMS por ST
	CstIcmsSimplesNacionalTributadaSemPermissaoCreditoComST                                   // 202 - Simples Nacional - Tributada sem permissao de credito e com cobranca do ICMS por ST
	CstIcmsSimplesNacionalIsencaoPorFaixaReceitaBrutaComST                                    // 203 - Simples Nacional - Isencao para faixa de receita bruta e com cobranca do ICMS por ST
	CstIcmsSimplesNacionalImune                                                               // 300 - Simples Nacional - Imune
	CstIcmsSimplesNacionalNaoTributada                                                        // 400 - Simples Nacional - Nao tributada
	CstIcmsSimplesNacionalCobradoAnteriormentePorST                                           // 500 - Simples Nacional - ICMS cobrado anteriormente por ST ou por antecipacao
	CstIcmsSimplesNacionalOutros                                                              // 900 - Simples Nacional - Outros
	CstIcmsTributacaoMonofasicaPropriaCombustiveis                                            // 002 - Tributacao Monofasica Propria do ICMS nas operacoes com combustiveis
	CstIcmsTributacaoMonofasicaPropriaComRetencaoCombustiveis                                 // 015 - Tributacao Monofasica Propria e com responsabilidade pela retencao do ICMS nas operacoes com combustiveis
	CstIcmsTributacaoMonofasicaRecolhimentoDiferidoCombustiveis                               // 053 - Tributacao Monofasica com recolhimento diferido do ICMS nas operacoes com combustiveis
	CstIcmsTributacaoMonofasicaCombustiveisCobradoAnteriormente                               // 061 - Tributacao Monofasica sobre combustiveis com ICMS cobrado anteriormente
	CstIcmsEstrangeiraImpDiretaTribMonofasicaPropriaCombustiveis                              // 102 - Estrangeira - Tributacao Monofasica Propria do ICMS nas operacoes com combustiveis
	CstIcmsEstrangeiraImpDiretaTribMonofasicaPropriaComRetencaoCombustiveis                   // 115 - Estrangeira - Tributacao Monofasica Propria e com responsabilidade pela retencao do ICMS
	CstIcmsEstrangeiraImpDiretaTribMonofasicaRecolhimentoDiferidoCombustiveis                 // 153 - Estrangeira - Tributacao Monofasica com recolhimento diferido do ICMS
	CstIcmsEstrangeiraImpDiretaTribMonofasicaCombustiveisCobradoAnteriormente                 // 161 - Estrangeira - Tributacao Monofasica sobre combustiveis com ICMS cobrado anteriormente
	CstIcmsEstrangeiraAdqMercIntTribMonofasicaPropriaCombustiveis                             // 202 - Estrangeira - Adq. mercado interno - Tributacao Monofasica Propria do ICMS
	CstIcmsEstrangeiraAdqMercIntTribMonofasicaPropriaComRetencaoCombustiveis                  // 215 - Estrangeira - Adq. mercado interno - Tributacao Monofasica Propria e com retencao
	CstIcmsEstrangeiraAdqMercIntTribMonofasicaRecolhimentoDiferidoCombustiveis                // 253 - Estrangeira - Adq. mercado interno - Tributacao Monofasica com recolhimento diferido
	CstIcmsEstrangeiraAdqMercIntTribMonofasicaCombustiveisCobradoAnteriormente                // 261 - Estrangeira - Adq. mercado interno - Tributacao Monofasica com ICMS cobrado anteriormente
)

// cstIcmsStrings mapeia o indice ordinal do enum para o codigo de texto SPED.
// A ordem corresponde exatamente ao iota acima (indice 0..121).
var cstIcmsStrings = [...]string{
	"",    // CstIcmsNenhum
	"000", // CstIcmsTributadaIntegralmente
	"010", // CstIcmsTributadaComCobracaPorST
	"020", // CstIcmsComReducao
	"030", // CstIcmsIsentaComCobracaPorST
	"040", // CstIcmsIsenta
	"041", // CstIcmsNaoTributada
	"050", // CstIcmsSuspensao
	"051", // CstIcmsDiferimento
	"060", // CstIcmsCobradoAnteriormentePorST
	"070", // CstIcmsComReducaoPorST
	"090", // CstIcmsOutros
	"100", // CstIcmsEstrangeiraImportacaoDiretaTributadaIntegralmente
	"110", // CstIcmsEstrangeiraImportacaoDiretaTributadaComCobracaPorST
	"120", // CstIcmsEstrangeiraImportacaoDiretaComReducao
	"130", // CstIcmsEstrangeiraImportacaoDiretaIsentaComCobracaPorST
	"140", // CstIcmsEstrangeiraImportacaoDiretaIsenta
	"141", // CstIcmsEstrangeiraImportacaoDiretaNaoTributada
	"150", // CstIcmsEstrangeiraImportacaoDiretaSuspensao
	"151", // CstIcmsEstrangeiraImportacaoDiretaDiferimento
	"160", // CstIcmsEstrangeiraImportacaoDiretaCobradoAnteriormentePorST
	"170", // CstIcmsEstrangeiraImportacaoDiretaComReducaoPorST
	"190", // CstIcmsEstrangeiraImportacaoDiretaOutros
	"200", // CstIcmsEstrangeiraAdqMercIntTributadaIntegralmente
	"210", // CstIcmsEstrangeiraAdqMercIntTributadaComCobracaPorST
	"220", // CstIcmsEstrangeiraAdqMercIntComReducao
	"230", // CstIcmsEstrangeiraAdqMercIntIsentaComCobracaPorST
	"240", // CstIcmsEstrangeiraAdqMercIntIsenta
	"241", // CstIcmsEstrangeiraAdqMercIntNaoTributada
	"250", // CstIcmsEstrangeiraAdqMercIntSuspensao
	"251", // CstIcmsEstrangeiraAdqMercIntDiferimento
	"260", // CstIcmsEstrangeiraAdqMercIntCobradoAnteriormentePorST
	"270", // CstIcmsEstrangeiraAdqMercIntComReducaoPorST
	"290", // CstIcmsEstrangeiraAdqMercIntOutros
	"300", // CstIcms300
	"310", // CstIcms310
	"320", // CstIcms320
	"330", // CstIcms330
	"340", // CstIcms340
	"341", // CstIcms341
	"350", // CstIcms350
	"351", // CstIcms351
	"360", // CstIcms360
	"370", // CstIcms370
	"390", // CstIcms390
	"400", // CstIcms400
	"410", // CstIcms410
	"420", // CstIcms420
	"430", // CstIcms430
	"440", // CstIcms440
	"441", // CstIcms441
	"450", // CstIcms450
	"451", // CstIcms451
	"460", // CstIcms460
	"470", // CstIcms470
	"490", // CstIcms490
	"500", // CstIcms500
	"510", // CstIcms510
	"520", // CstIcms520
	"530", // CstIcms530
	"540", // CstIcms540
	"541", // CstIcms541
	"550", // CstIcms550
	"551", // CstIcms551
	"560", // CstIcms560
	"570", // CstIcms570
	"590", // CstIcms590
	"600", // CstIcms600
	"610", // CstIcms610
	"620", // CstIcms620
	"630", // CstIcms630
	"640", // CstIcms640
	"641", // CstIcms641
	"650", // CstIcms650
	"651", // CstIcms651
	"660", // CstIcms660
	"670", // CstIcms670
	"690", // CstIcms690
	"700", // CstIcms700
	"710", // CstIcms710
	"720", // CstIcms720
	"730", // CstIcms730
	"740", // CstIcms740
	"741", // CstIcms741
	"750", // CstIcms750
	"751", // CstIcms751
	"760", // CstIcms760
	"770", // CstIcms770
	"790", // CstIcms790
	"800", // CstIcms800
	"810", // CstIcms810
	"820", // CstIcms820
	"830", // CstIcms830
	"840", // CstIcms840
	"841", // CstIcms841
	"850", // CstIcms850
	"851", // CstIcms851
	"860", // CstIcms860
	"870", // CstIcms870
	"890", // CstIcms890
	"101", // CstIcmsSimplesNacionalTributadaComPermissaoCredito
	"102", // CstIcmsSimplesNacionalTributadaSemPermissaoCredito
	"103", // CstIcmsSimplesNacionalIsencaoPorFaixaReceitaBruta
	"201", // CstIcmsSimplesNacionalTributadaComPermissaoCreditoComST
	"202", // CstIcmsSimplesNacionalTributadaSemPermissaoCreditoComST
	"203", // CstIcmsSimplesNacionalIsencaoPorFaixaReceitaBrutaComST
	"300", // CstIcmsSimplesNacionalImune
	"400", // CstIcmsSimplesNacionalNaoTributada
	"500", // CstIcmsSimplesNacionalCobradoAnteriormentePorST
	"900", // CstIcmsSimplesNacionalOutros
	"002", // CstIcmsTributacaoMonofasicaPropriaCombustiveis
	"015", // CstIcmsTributacaoMonofasicaPropriaComRetencaoCombustiveis
	"053", // CstIcmsTributacaoMonofasicaRecolhimentoDiferidoCombustiveis
	"061", // CstIcmsTributacaoMonofasicaCombustiveisCobradoAnteriormente
	"102", // CstIcmsEstrangeiraImpDiretaTribMonofasicaPropriaCombustiveis
	"115", // CstIcmsEstrangeiraImpDiretaTribMonofasicaPropriaComRetencaoCombustiveis
	"153", // CstIcmsEstrangeiraImpDiretaTribMonofasicaRecolhimentoDiferidoCombustiveis
	"161", // CstIcmsEstrangeiraImpDiretaTribMonofasicaCombustiveisCobradoAnteriormente
	"202", // CstIcmsEstrangeiraAdqMercIntTribMonofasicaPropriaCombustiveis
	"215", // CstIcmsEstrangeiraAdqMercIntTribMonofasicaPropriaComRetencaoCombustiveis
	"253", // CstIcmsEstrangeiraAdqMercIntTribMonofasicaRecolhimentoDiferidoCombustiveis
	"261", // CstIcmsEstrangeiraAdqMercIntTribMonofasicaCombustiveisCobradoAnteriormente
}

// String retorna o codigo SPED do CstIcms (ex: "000", "010", "090", "101").
func (c CstIcms) String() string {
	if int(c) >= 0 && int(c) < len(cstIcmsStrings) {
		return cstIcmsStrings[c]
	}
	return fmt.Sprintf("CstIcms(%d)", int(c))
}

// ---------------------------------------------------------------------------
// CstPisCofins - Codigo da Situacao Tributaria referente ao PIS/COFINS
// (enum compacto usado no SPED Fiscal)
// Portado de: TACBrCstPisCofins (ACBrEFDBlocos.pas / ACBrEPCBlocos.pas)
// ---------------------------------------------------------------------------

// CstPisCofins representa o CST compacto do PIS/COFINS.
type CstPisCofins int

const (
	CstPisCofinsOperTribComAliqBasica         CstPisCofins = iota // 01 - Operacao Tributavel com Aliquota Basica
	CstPisCofinsOperTribAliqZero                                  // 06 - Operacao Tributavel a Aliquota Zero
	CstPisCofinsOperIsentaContribuicao                            // 07 - Operacao Isenta da Contribuicao
	CstPisCofinsOperSemIncidenciaContribuicao                     // 08 - Operacao sem Incidencia da Contribuicao
	CstPisCofinsOperComSuspensaoContribuicao                      // 09 - Operacao com Suspensao da Contribuicao
	CstPisCofinsOutrasOperacoesSaida                              // 49 - Outras Operacoes de Saida
	CstPisCofinsOutrasDespesas                                    // 99 - Outras Operacoes
	CstPisCofinsNenhum                                            // Nenhum (vazio)
)

// cstPisCofinsStrings mapeia o indice ordinal do enum para o codigo de texto SPED.
var cstPisCofinsStrings = [...]string{
	"01", // CstPisCofinsOperTribComAliqBasica
	"06", // CstPisCofinsOperTribAliqZero
	"07", // CstPisCofinsOperIsentaContribuicao
	"08", // CstPisCofinsOperSemIncidenciaContribuicao
	"09", // CstPisCofinsOperComSuspensaoContribuicao
	"49", // CstPisCofinsOutrasOperacoesSaida
	"99", // CstPisCofinsOutrasDespesas
	"",   // CstPisCofinsNenhum
}

// String retorna o codigo SPED do CstPisCofins (ex: "01", "06", "99").
func (c CstPisCofins) String() string {
	if int(c) >= 0 && int(c) < len(cstPisCofinsStrings) {
		return cstPisCofinsStrings[c]
	}
	return fmt.Sprintf("CstPisCofins(%d)", int(c))
}

// ---------------------------------------------------------------------------
// CstIpi - Codigo da Situacao Tributaria referente ao IPI
// Portado de: TACBrCstIpi (ACBrEFDBlocos.pas / ACBrEPCBlocos.pas)
// ---------------------------------------------------------------------------

// CstIpi representa o Codigo da Situacao Tributaria do IPI.
type CstIpi int

const (
	CstIpiEntradaRecuperacaoCredito CstIpi = iota // 00 - Entrada com recuperacao de credito
	CstIpiEntradaTributadaZero                    // 01 - Entrada tributada com aliquota zero
	CstIpiEntradaIsenta                           // 02 - Entrada isenta
	CstIpiEntradaNaoTributada                     // 03 - Entrada nao-tributada
	CstIpiEntradaImune                            // 04 - Entrada imune
	CstIpiEntradaComSuspensao                     // 05 - Entrada com suspensao
	CstIpiOutrasEntradas                          // 49 - Outras entradas
	CstIpiSaidaTributada                          // 50 - Saida tributada
	CstIpiSaidaTributadaZero                      // 51 - Saida tributada com aliquota zero
	CstIpiSaidaIsenta                             // 52 - Saida isenta
	CstIpiSaidaNaoTributada                       // 53 - Saida nao-tributada
	CstIpiSaidaImune                              // 54 - Saida imune
	CstIpiSaidaComSuspensao                       // 55 - Saida com suspensao
	CstIpiOutrasSaidas                            // 99 - Outras saidas
	CstIpiVazio                                   // Vazio (sem informacao)
)

// cstIpiStrings mapeia o indice ordinal do enum para o codigo de texto SPED.
var cstIpiStrings = [...]string{
	"00", // CstIpiEntradaRecuperacaoCredito
	"01", // CstIpiEntradaTributadaZero
	"02", // CstIpiEntradaIsenta
	"03", // CstIpiEntradaNaoTributada
	"04", // CstIpiEntradaImune
	"05", // CstIpiEntradaComSuspensao
	"49", // CstIpiOutrasEntradas
	"50", // CstIpiSaidaTributada
	"51", // CstIpiSaidaTributadaZero
	"52", // CstIpiSaidaIsenta
	"53", // CstIpiSaidaNaoTributada
	"54", // CstIpiSaidaImune
	"55", // CstIpiSaidaComSuspensao
	"99", // CstIpiOutrasSaidas
	"",   // CstIpiVazio
}

// String retorna o codigo SPED do CstIpi (ex: "00", "01", "50", "99").
func (c CstIpi) String() string {
	if int(c) >= 0 && int(c) < len(cstIpiStrings) {
		return cstIpiStrings[c]
	}
	return fmt.Sprintf("CstIpi(%d)", int(c))
}

// ---------------------------------------------------------------------------
// CstPis - Codigo da Situacao Tributaria referente ao PIS
// Portado de: TACBrCstPis (ACBrEFDBlocos.pas / ACBrEPCBlocos.pas)
// ---------------------------------------------------------------------------

// CstPis representa o Codigo da Situacao Tributaria do PIS.
type CstPis int

const (
	CstPisValorAliquotaNormal                           CstPis = iota // 01 - Operacao Tributavel com Aliquota Basica (valor da operacao aliquota normal cumulativo/nao cumulativo)
	CstPisValorAliquotaDiferenciada                                   // 02 - Operacao Tributavel com Aliquota Diferenciada
	CstPisQtdeAliquotaUnidade                                         // 03 - Operacao Tributavel com Aliquota por Unidade de Medida de Produto
	CstPisMonofaticaAliquotaZero                                      // 04 - Operacao Tributavel Monofasica - Revenda a Aliquota Zero
	CstPisValorAliquotaPorST                                          // 05 - Operacao Tributavel por Substituicao Tributaria
	CstPisAliquotaZero                                                // 06 - Operacao Tributavel a Aliquota Zero
	CstPisIsentaContribuicao                                          // 07 - Operacao Isenta da Contribuicao
	CstPisSemIncidenciaContribuicao                                   // 08 - Operacao sem Incidencia da Contribuicao
	CstPisSuspensaoContribuicao                                       // 09 - Operacao com Suspensao da Contribuicao
	CstPisOutrasOperacoesSaida                                        // 49 - Outras Operacoes de Saida
	CstPisOperCredExcRecTribMercInt                                   // 50 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita Tributada no Mercado Interno
	CstPisOperCredExcRecNaoTribMercInt                                // 51 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita Nao Tributada no Mercado Interno
	CstPisOperCredExcRecExportacao                                    // 52 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita de Exportacao
	CstPisOperCredRecTribNaoTribMercInt                               // 53 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno
	CstPisOperCredRecTribMercIntEExportacao                           // 54 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas no Mercado Interno e de Exportacao
	CstPisOperCredRecNaoTribMercIntEExportacao                        // 55 - Operacao com Direito a Credito - Vinculada a Receitas Nao-Tributadas no Mercado Interno e de Exportacao
	CstPisOperCredRecTribENaoTribMercIntEExportacao                   // 56 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno e de Exportacao
	CstPisCredPresAquiExcRecTribMercInt                               // 60 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita Tributada no Mercado Interno
	CstPisCredPresAquiExcRecNaoTribMercInt                            // 61 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita Nao-Tributada no Mercado Interno
	CstPisCredPresAquiExcRecExportacao                                // 62 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita de Exportacao
	CstPisCredPresAquiRecTribNaoTribMercInt                           // 63 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno
	CstPisCredPresAquiRecTribMercIntEExportacao                       // 64 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas no Mercado Interno e de Exportacao
	CstPisCredPresAquiRecNaoTribMercIntEExportacao                    // 65 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Nao-Tributadas no Mercado Interno e de Exportacao
	CstPisCredPresAquiRecTribENaoTribMercIntEExportacao               // 66 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno e de Exportacao
	CstPisOutrasOperacoesCredPresumido                                // 67 - Credito Presumido - Outras Operacoes
	CstPisOperAquiSemDirCredito                                       // 70 - Operacao de Aquisicao sem Direito a Credito
	CstPisOperAquiComIsencao                                          // 71 - Operacao de Aquisicao com Isencao
	CstPisOperAquiComSuspensao                                        // 72 - Operacao de Aquisicao com Suspensao
	CstPisOperAquiAliquotaZero                                        // 73 - Operacao de Aquisicao a Aliquota Zero
	CstPisOperAquiSemIncidenciaContribuicao                           // 74 - Operacao de Aquisicao sem Incidencia da Contribuicao
	CstPisOperAquiPorST                                               // 75 - Operacao de Aquisicao por Substituicao Tributaria
	CstPisOutrasOperacoesEntrada                                      // 98 - Outras Operacoes de Entrada
	CstPisOutrasOperacoes                                             // 99 - Outras Operacoes
	CstPisNenhum                                                      // Nenhum (vazio)
)

// cstPisStrings mapeia o indice ordinal do enum para o codigo de texto SPED.
var cstPisStrings = [...]string{
	"01", // CstPisValorAliquotaNormal
	"02", // CstPisValorAliquotaDiferenciada
	"03", // CstPisQtdeAliquotaUnidade
	"04", // CstPisMonofaticaAliquotaZero
	"05", // CstPisValorAliquotaPorST
	"06", // CstPisAliquotaZero
	"07", // CstPisIsentaContribuicao
	"08", // CstPisSemIncidenciaContribuicao
	"09", // CstPisSuspensaoContribuicao
	"49", // CstPisOutrasOperacoesSaida
	"50", // CstPisOperCredExcRecTribMercInt
	"51", // CstPisOperCredExcRecNaoTribMercInt
	"52", // CstPisOperCredExcRecExportacao
	"53", // CstPisOperCredRecTribNaoTribMercInt
	"54", // CstPisOperCredRecTribMercIntEExportacao
	"55", // CstPisOperCredRecNaoTribMercIntEExportacao
	"56", // CstPisOperCredRecTribENaoTribMercIntEExportacao
	"60", // CstPisCredPresAquiExcRecTribMercInt
	"61", // CstPisCredPresAquiExcRecNaoTribMercInt
	"62", // CstPisCredPresAquiExcRecExportacao
	"63", // CstPisCredPresAquiRecTribNaoTribMercInt
	"64", // CstPisCredPresAquiRecTribMercIntEExportacao
	"65", // CstPisCredPresAquiRecNaoTribMercIntEExportacao
	"66", // CstPisCredPresAquiRecTribENaoTribMercIntEExportacao
	"67", // CstPisOutrasOperacoesCredPresumido
	"70", // CstPisOperAquiSemDirCredito
	"71", // CstPisOperAquiComIsencao
	"72", // CstPisOperAquiComSuspensao
	"73", // CstPisOperAquiAliquotaZero
	"74", // CstPisOperAquiSemIncidenciaContribuicao
	"75", // CstPisOperAquiPorST
	"98", // CstPisOutrasOperacoesEntrada
	"99", // CstPisOutrasOperacoes
	"",   // CstPisNenhum
}

// String retorna o codigo SPED do CstPis (ex: "01", "02", "49", "99").
func (c CstPis) String() string {
	if int(c) >= 0 && int(c) < len(cstPisStrings) {
		return cstPisStrings[c]
	}
	return fmt.Sprintf("CstPis(%d)", int(c))
}

// ---------------------------------------------------------------------------
// CstCofins - Codigo da Situacao Tributaria referente ao COFINS
// Portado de: TACBrCstCofins (ACBrEFDBlocos.pas / ACBrEPCBlocos.pas)
// ---------------------------------------------------------------------------

// CstCofins representa o Codigo da Situacao Tributaria do COFINS.
type CstCofins int

const (
	CstCofinsValorAliquotaNormal                           CstCofins = iota // 01 - Operacao Tributavel com Aliquota Basica (valor da operacao aliquota normal cumulativo/nao cumulativo)
	CstCofinsValorAliquotaDiferenciada                                      // 02 - Operacao Tributavel com Aliquota Diferenciada
	CstCofinsQtdeAliquotaUnidade                                            // 03 - Operacao Tributavel com Aliquota por Unidade de Medida de Produto
	CstCofinsMonofaticaAliquotaZero                                         // 04 - Operacao Tributavel Monofasica - Revenda a Aliquota Zero
	CstCofinsValorAliquotaPorST                                             // 05 - Operacao Tributavel por Substituicao Tributaria
	CstCofinsAliquotaZero                                                   // 06 - Operacao Tributavel a Aliquota Zero
	CstCofinsIsentaContribuicao                                             // 07 - Operacao Isenta da Contribuicao
	CstCofinsSemIncidenciaContribuicao                                      // 08 - Operacao sem Incidencia da Contribuicao
	CstCofinsSuspensaoContribuicao                                          // 09 - Operacao com Suspensao da Contribuicao
	CstCofinsOutrasOperacoesSaida                                           // 49 - Outras Operacoes de Saida
	CstCofinsOperCredExcRecTribMercInt                                      // 50 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita Tributada no Mercado Interno
	CstCofinsOperCredExcRecNaoTribMercInt                                   // 51 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita Nao-Tributada no Mercado Interno
	CstCofinsOperCredExcRecExportacao                                       // 52 - Operacao com Direito a Credito - Vinculada Exclusivamente a Receita de Exportacao
	CstCofinsOperCredRecTribNaoTribMercInt                                  // 53 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno
	CstCofinsOperCredRecTribMercIntEExportacao                              // 54 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas no Mercado Interno e de Exportacao
	CstCofinsOperCredRecNaoTribMercIntEExportacao                           // 55 - Operacao com Direito a Credito - Vinculada a Receitas Nao Tributadas no Mercado Interno e de Exportacao
	CstCofinsOperCredRecTribENaoTribMercIntEExportacao                      // 56 - Operacao com Direito a Credito - Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno e de Exportacao
	CstCofinsCredPresAquiExcRecTribMercInt                                  // 60 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita Tributada no Mercado Interno
	CstCofinsCredPresAquiExcRecNaoTribMercInt                               // 61 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita Nao-Tributada no Mercado Interno
	CstCofinsCredPresAquiExcRecExportacao                                   // 62 - Credito Presumido - Operacao de Aquisicao Vinculada Exclusivamente a Receita de Exportacao
	CstCofinsCredPresAquiRecTribNaoTribMercInt                              // 63 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno
	CstCofinsCredPresAquiRecTribMercIntEExportacao                          // 64 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas no Mercado Interno e de Exportacao
	CstCofinsCredPresAquiRecNaoTribMercIntEExportacao                       // 65 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Nao-Tributadas no Mercado Interno e de Exportacao
	CstCofinsCredPresAquiRecTribENaoTribMercIntEExportacao                  // 66 - Credito Presumido - Operacao de Aquisicao Vinculada a Receitas Tributadas e Nao-Tributadas no Mercado Interno e de Exportacao
	CstCofinsOutrasOperacoesCredPresumido                                   // 67 - Credito Presumido - Outras Operacoes
	CstCofinsOperAquiSemDirCredito                                          // 70 - Operacao de Aquisicao sem Direito a Credito
	CstCofinsOperAquiComIsencao                                             // 71 - Operacao de Aquisicao com Isencao
	CstCofinsOperAquiComSuspensao                                           // 72 - Operacao de Aquisicao com Suspensao
	CstCofinsOperAquiAliquotaZero                                           // 73 - Operacao de Aquisicao a Aliquota Zero
	CstCofinsOperAquiSemIncidenciaContribuicao                              // 74 - Operacao de Aquisicao sem Incidencia da Contribuicao
	CstCofinsOperAquiPorST                                                  // 75 - Operacao de Aquisicao por Substituicao Tributaria
	CstCofinsOutrasOperacoesEntrada                                         // 98 - Outras Operacoes de Entrada
	CstCofinsOutrasOperacoes                                                // 99 - Outras Operacoes
	CstCofinsNenhum                                                         // Nenhum (vazio)
)

// cstCofinsStrings mapeia o indice ordinal do enum para o codigo de texto SPED.
var cstCofinsStrings = [...]string{
	"01", // CstCofinsValorAliquotaNormal
	"02", // CstCofinsValorAliquotaDiferenciada
	"03", // CstCofinsQtdeAliquotaUnidade
	"04", // CstCofinsMonofaticaAliquotaZero
	"05", // CstCofinsValorAliquotaPorST
	"06", // CstCofinsAliquotaZero
	"07", // CstCofinsIsentaContribuicao
	"08", // CstCofinsSemIncidenciaContribuicao
	"09", // CstCofinsSuspensaoContribuicao
	"49", // CstCofinsOutrasOperacoesSaida
	"50", // CstCofinsOperCredExcRecTribMercInt
	"51", // CstCofinsOperCredExcRecNaoTribMercInt
	"52", // CstCofinsOperCredExcRecExportacao
	"53", // CstCofinsOperCredRecTribNaoTribMercInt
	"54", // CstCofinsOperCredRecTribMercIntEExportacao
	"55", // CstCofinsOperCredRecNaoTribMercIntEExportacao
	"56", // CstCofinsOperCredRecTribENaoTribMercIntEExportacao
	"60", // CstCofinsCredPresAquiExcRecTribMercInt
	"61", // CstCofinsCredPresAquiExcRecNaoTribMercInt
	"62", // CstCofinsCredPresAquiExcRecExportacao
	"63", // CstCofinsCredPresAquiRecTribNaoTribMercInt
	"64", // CstCofinsCredPresAquiRecTribMercIntEExportacao
	"65", // CstCofinsCredPresAquiRecNaoTribMercIntEExportacao
	"66", // CstCofinsCredPresAquiRecTribENaoTribMercIntEExportacao
	"67", // CstCofinsOutrasOperacoesCredPresumido
	"70", // CstCofinsOperAquiSemDirCredito
	"71", // CstCofinsOperAquiComIsencao
	"72", // CstCofinsOperAquiComSuspensao
	"73", // CstCofinsOperAquiAliquotaZero
	"74", // CstCofinsOperAquiSemIncidenciaContribuicao
	"75", // CstCofinsOperAquiPorST
	"98", // CstCofinsOutrasOperacoesEntrada
	"99", // CstCofinsOutrasOperacoes
	"",   // CstCofinsNenhum
}

// String retorna o codigo SPED do CstCofins (ex: "01", "02", "49", "99").
func (c CstCofins) String() string {
	if int(c) >= 0 && int(c) < len(cstCofinsStrings) {
		return cstCofinsStrings[c]
	}
	return fmt.Sprintf("CstCofins(%d)", int(c))
}
