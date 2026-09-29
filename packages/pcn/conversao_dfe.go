// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-29.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package pcn

import (
	"errors"
	"fmt"
)

// Enums compartilhados por todos os documentos fiscais eletronicos.
// Porte de ACBrDFe.Conversao.pas.
//
// Cada enum expoe String() com o codigo exato do leiaute e ParseXxx com
// erro -- os StrToXxx do Delphi levantam excecao, que em Go vira error.

// ErrEnumInvalido indica valor de string que nao corresponde a nenhum
// membro do enum.
var ErrEnumInvalido = errors.New("pcn: valor invalido para enum")

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %s nao aceita %q", ErrEnumInvalido, enum, valor)
}

// ===========================================================================
// TipoAmbiente (TACBrTipoAmbiente)
// ===========================================================================

// TipoAmbiente identifica o ambiente de autorizacao do documento.
type TipoAmbiente int

const (
	TaProducao    TipoAmbiente = iota // 1
	TaHomologacao                     // 2
)

// String retorna o codigo do leiaute ("1" ou "2").
func (v TipoAmbiente) String() string {
	switch v {
	case TaProducao:
		return "1"
	case TaHomologacao:
		return "2"
	default:
		return ""
	}
}

// ParseTipoAmbiente converte o codigo do leiaute no enum.
func ParseTipoAmbiente(s string) (TipoAmbiente, error) {
	switch s {
	case "1":
		return TaProducao, nil
	case "2":
		return TaHomologacao, nil
	default:
		return TaProducao, erroEnum("TipoAmbiente", s)
	}
}

// ===========================================================================
// TipoEmissao (TACBrTipoEmissao)
// ===========================================================================

// TipoEmissao identifica a forma de emissao do documento.
type TipoEmissao int

const (
	TeNormal       TipoEmissao = iota // 1
	TeContingencia                    // 2
	TeSCAN                            // 3
	TeDPEC                            // 4
	TeFSDA                            // 5
	TeSVCAN                           // 6
	TeSVCRS                           // 7
	TeSVCSP                           // 8
	TeOffLine                         // 9
)

var tipoEmissaoCodigos = [...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

// String retorna o codigo do leiaute ("1".."9").
func (v TipoEmissao) String() string {
	if int(v) < 0 || int(v) >= len(tipoEmissaoCodigos) {
		return ""
	}
	return tipoEmissaoCodigos[v]
}

// ParseTipoEmissao converte o codigo do leiaute no enum.
func ParseTipoEmissao(s string) (TipoEmissao, error) {
	for i, c := range tipoEmissaoCodigos {
		if c == s {
			return TipoEmissao(i), nil
		}
	}
	return TeNormal, erroEnum("TipoEmissao", s)
}

// ===========================================================================
// Indicador (TIndicador) e IndicadorEx (TIndicadorEx)
// ===========================================================================

// Indicador e o par sim/nao do leiaute, onde sim vale "1" e nao vale "0".
type Indicador int

const (
	TiSim Indicador = iota // 1
	TiNao                  // 0
)

// String retorna "1" para sim e "0" para nao.
func (v Indicador) String() string {
	switch v {
	case TiSim:
		return "1"
	case TiNao:
		return "0"
	default:
		return ""
	}
}

// ParseIndicador converte o codigo do leiaute no enum.
func ParseIndicador(s string) (Indicador, error) {
	switch s {
	case "1":
		return TiSim, nil
	case "0":
		return TiNao, nil
	default:
		return TiNao, erroEnum("Indicador", s)
	}
}

// IndicadorEx e o indicador com um terceiro estado para campo ausente.
type IndicadorEx int

const (
	TieNenhum IndicadorEx = iota // vazio
	TieSim                       // 1
	TieNao                       // 0
)

// String retorna "", "1" ou "0".
func (v IndicadorEx) String() string {
	switch v {
	case TieSim:
		return "1"
	case TieNao:
		return "0"
	default:
		return ""
	}
}

// ParseIndicadorEx converte o codigo do leiaute no enum. String vazia e
// valida e resulta em TieNenhum.
func ParseIndicadorEx(s string) (IndicadorEx, error) {
	switch s {
	case "":
		return TieNenhum, nil
	case "1":
		return TieSim, nil
	case "0":
		return TieNao, nil
	default:
		return TieNenhum, erroEnum("IndicadorEx", s)
	}
}

// ===========================================================================
// OrigemMercadoria (TOrigemMercadoria)
// ===========================================================================

// OrigemMercadoria e a origem da mercadoria para fins de ICMS.
type OrigemMercadoria int

const (
	OeNacional                                  OrigemMercadoria = iota // 0
	OeEstrangeiraImportacaoDireta                                       // 1
	OeEstrangeiraAdquiridaBrasil                                        // 2
	OeNacionalConteudoImportacaoSuperior40                              // 3
	OeNacionalProcessosBasicos                                          // 4
	OeNacionalConteudoImportacaoInferiorIgual40                         // 5
	OeEstrangeiraImportacaoDiretaSemSimilar                             // 6
	OeEstrangeiraAdquiridaBrasilSemSimilar                              // 7
	OeNacionalConteudoImportacaoSuperior70                              // 8
	OeReservadoParaUsoFuturo                                            // 9
	OeVazio                                                             // vazio
)

var origemMercadoriaCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ""}

// String retorna o codigo do leiaute ("0".."9" ou "").
func (v OrigemMercadoria) String() string {
	if int(v) < 0 || int(v) >= len(origemMercadoriaCodigos) {
		return ""
	}
	return origemMercadoriaCodigos[v]
}

// ParseOrigemMercadoria converte o codigo do leiaute no enum.
func ParseOrigemMercadoria(s string) (OrigemMercadoria, error) {
	for i, c := range origemMercadoriaCodigos {
		if c == s {
			return OrigemMercadoria(i), nil
		}
	}
	return OeVazio, erroEnum("OrigemMercadoria", s)
}

// ===========================================================================
// CSTIcms (TCSTIcms)
// ===========================================================================

// CSTIcms e o codigo de situacao tributaria do ICMS.
//
// O ACBr mantem DUAS tabelas de codigo para o mesmo enum: a de entrada,
// usada na leitura, e a de saida, usada na geracao. Elas divergem em cinco
// membros -- cstICMSOutraUF, cstPart10, cstPart90, cstRep41 e cstRep60 --
// que na entrada tem sufixo ("10part", "41rep") e na saida nao. Por isso
// String() e ParseCSTIcms NAO sao inversas uma da outra, e e assim no
// original: ver TCSTIcmsArrayStringsEnt e TCSTIcmsArrayStringsSai em
// ACBrDFe.Conversao.pas.
type CSTIcms int

const (
	CSTVazio       CSTIcms = iota // "" / ""
	CST00                         // 00
	CST10                         // 10
	CST20                         // 20
	CST30                         // 30
	CST40                         // 40
	CST41                         // 41
	CST45                         // 45
	CST50                         // 50
	CST51                         // 51
	CST60                         // 60
	CST70                         // 70
	CST80                         // 80 (apenas CTe)
	CST81                         // 81 (apenas CTe)
	CST90                         // 90
	CSTICMSOutraUF                // 91 na entrada, 90 na saida
	CSTICMSSN                     // SN
	CSTPart10                     // 10part na entrada, 10 na saida
	CSTPart90                     // 90part na entrada, 90 na saida
	CSTRep41                      // 41rep na entrada, 41 na saida
	CSTRep60                      // 60rep na entrada, 60 na saida
	CST02                         // 02
	CST15                         // 15
	CST53                         // 53
	CST61                         // 61
	CST01                         // 01
	CST12                         // 12
	CST13                         // 13
	CST14                         // 14
	CST21                         // 21
	CST72                         // 72
	CST73                         // 73
	CST74                         // 74
	CSTPart20                     // 20part na entrada, 20 na saida
)

var cstIcmsEntrada = [...]string{
	"", "00", "10", "20", "30", "40", "41", "45", "50", "51", "60", "70",
	"80", "81", "90", "91", "SN", "10part", "90part", "41rep", "60rep",
	"02", "15", "53", "61", "01", "12", "13", "14", "21", "72", "73", "74", "20part",
}

var cstIcmsSaida = [...]string{
	"", "00", "10", "20", "30", "40", "41", "45", "50", "51", "60", "70",
	"80", "81", "90", "90", "SN", "10", "90", "41", "60",
	"02", "15", "53", "61", "01", "12", "13", "14", "21", "72", "73", "74", "20",
}

// String retorna o codigo de SAIDA, usado na geracao do XML.
// Porte de CSTICMSToStr.
func (v CSTIcms) String() string {
	if int(v) < 0 || int(v) >= len(cstIcmsSaida) {
		return ""
	}
	return cstIcmsSaida[v]
}

// CodigoEntrada retorna o codigo de ENTRADA, o mesmo que ParseCSTIcms aceita.
func (v CSTIcms) CodigoEntrada() string {
	if int(v) < 0 || int(v) >= len(cstIcmsEntrada) {
		return ""
	}
	return cstIcmsEntrada[v]
}

// ParseCSTIcms converte o codigo do leiaute no enum, usando a tabela de
// ENTRADA -- exatamente como TryStrToCSTICMS.
func ParseCSTIcms(s string) (CSTIcms, error) {
	for i, c := range cstIcmsEntrada {
		if c == s {
			return CSTIcms(i), nil
		}
	}
	return CSTVazio, erroEnum("CSTIcms", s)
}

// ===========================================================================
// CSTPis (TCSTPis) e CSTCofins (TCSTCofins)
// ===========================================================================

// cstPisCofinsCodigos e a tabela comum a PIS e COFINS -- os dois enums do
// ACBr tem exatamente os mesmos 33 codigos, na mesma ordem.
var cstPisCofinsCodigos = [...]string{
	"01", "02", "03", "04", "05", "06", "07", "08", "09",
	"49", "50", "51", "52", "53", "54", "55", "56",
	"60", "61", "62", "63", "64", "65", "66", "67",
	"70", "71", "72", "73", "74", "75", "98", "99",
}

// CSTPis e o codigo de situacao tributaria do PIS.
type CSTPis int

// Membros de CSTPis, na ordem de TCSTPis.
const (
	Pis01 CSTPis = iota
	Pis02
	Pis03
	Pis04
	Pis05
	Pis06
	Pis07
	Pis08
	Pis09
	Pis49
	Pis50
	Pis51
	Pis52
	Pis53
	Pis54
	Pis55
	Pis56
	Pis60
	Pis61
	Pis62
	Pis63
	Pis64
	Pis65
	Pis66
	Pis67
	Pis70
	Pis71
	Pis72
	Pis73
	Pis74
	Pis75
	Pis98
	Pis99
)

// String retorna o codigo de dois digitos do leiaute.
func (v CSTPis) String() string {
	if int(v) < 0 || int(v) >= len(cstPisCofinsCodigos) {
		return ""
	}
	return cstPisCofinsCodigos[v]
}

// ParseCSTPis converte o codigo do leiaute no enum.
func ParseCSTPis(s string) (CSTPis, error) {
	for i, c := range cstPisCofinsCodigos {
		if c == s {
			return CSTPis(i), nil
		}
	}
	return Pis01, erroEnum("CSTPis", s)
}

// CSTCofins e o codigo de situacao tributaria da COFINS.
type CSTCofins int

// Membros de CSTCofins, na ordem de TCSTCofins.
const (
	Cof01 CSTCofins = iota
	Cof02
	Cof03
	Cof04
	Cof05
	Cof06
	Cof07
	Cof08
	Cof09
	Cof49
	Cof50
	Cof51
	Cof52
	Cof53
	Cof54
	Cof55
	Cof56
	Cof60
	Cof61
	Cof62
	Cof63
	Cof64
	Cof65
	Cof66
	Cof67
	Cof70
	Cof71
	Cof72
	Cof73
	Cof74
	Cof75
	Cof98
	Cof99
)

// String retorna o codigo de dois digitos do leiaute.
func (v CSTCofins) String() string {
	if int(v) < 0 || int(v) >= len(cstPisCofinsCodigos) {
		return ""
	}
	return cstPisCofinsCodigos[v]
}

// ParseCSTCofins converte o codigo do leiaute no enum.
func ParseCSTCofins(s string) (CSTCofins, error) {
	for i, c := range cstPisCofinsCodigos {
		if c == s {
			return CSTCofins(i), nil
		}
	}
	return Cof01, erroEnum("CSTCofins", s)
}

// ===========================================================================
// TpPagAnt (TtpPagAnt)
// ===========================================================================

// TpPagAnt indica o tipo de pagamento antecipado.
type TpPagAnt int

const (
	TpaNenhum                                   TpPagAnt = iota // vazio
	TpaPagServicoNaoContinuado                                  // 1
	TpaPagServicoContinuado                                     // 2
	TpaFornecimentoComPagRealizadoAnteriormente                 // 3
)

var tpPagAntCodigos = [...]string{"", "1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpPagAnt) String() string {
	if int(v) < 0 || int(v) >= len(tpPagAntCodigos) {
		return ""
	}
	return tpPagAntCodigos[v]
}

// ParseTpPagAnt converte o codigo do leiaute no enum. String vazia e valida.
func ParseTpPagAnt(s string) (TpPagAnt, error) {
	for i, c := range tpPagAntCodigos {
		if c == s {
			return TpPagAnt(i), nil
		}
	}
	return TpaNenhum, erroEnum("TpPagAnt", s)
}
