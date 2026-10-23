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

// ===========================================================================
// TipoNFe (TTipoNFe) -- tag tpNF
// ===========================================================================

// TipoNFe indica se a operacao e de entrada ou de saida.
//
// Atencao ao valor zero: TnEntrada e o primeiro membro, entao um ide recem
// criado grava "0" (entrada). A property do ACBr declara `default tnSaida`,
// mas o `default` do Delphi e metadado de streaming e NAO inicializa nada --
// TIde.Create nao atribui tpNF.
type TipoNFe int

const (
	TnEntrada TipoNFe = iota // 0
	TnSaida                  // 1
)

var tipoNFeCodigos = [...]string{"0", "1"}

// String retorna o codigo do leiaute ("0" ou "1").
func (v TipoNFe) String() string {
	if int(v) < 0 || int(v) >= len(tipoNFeCodigos) {
		return ""
	}
	return tipoNFeCodigos[v]
}

// ParseTipoNFe converte o codigo do leiaute no enum.
func ParseTipoNFe(s string) (TipoNFe, error) {
	for i, c := range tipoNFeCodigos {
		if c == s {
			return TipoNFe(i), nil
		}
	}
	return TnEntrada, erroEnum("TipoNFe", s)
}

// ===========================================================================
// TipoImpressao (TACBrTipoImpressao) -- tag tpImp
// ===========================================================================

// TipoImpressao e o formato de impressao do DANFE.
type TipoImpressao int

const (
	TiSemGeracao        TipoImpressao = iota // 0
	TiRetrato                                // 1
	TiPaisagem                               // 2
	TiSimplificado                           // 3
	TiNFCe                                   // 4
	TiMsgEletronica                          // 5
	TiSimplificadoTipo2                      // 6
)

var tipoImpressaoCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6"}

// String retorna o codigo do leiaute ("0".."6").
func (v TipoImpressao) String() string {
	if int(v) < 0 || int(v) >= len(tipoImpressaoCodigos) {
		return ""
	}
	return tipoImpressaoCodigos[v]
}

// ParseTipoImpressao converte o codigo do leiaute no enum.
func ParseTipoImpressao(s string) (TipoImpressao, error) {
	for i, c := range tipoImpressaoCodigos {
		if c == s {
			return TipoImpressao(i), nil
		}
	}
	return TiSemGeracao, erroEnum("TipoImpressao", s)
}

// ===========================================================================
// ProcessoEmissao (TACBrProcessoEmissao) -- tag procEmi
// ===========================================================================

// ProcessoEmissao identifica o processo que emitiu o documento.
type ProcessoEmissao int

const (
	PeAplicativoContribuinte        ProcessoEmissao = iota // 0
	PeAvulsaFisco                                          // 1
	PeAvulsaContribuinte                                   // 2
	PeContribuinteAplicativoFisco                          // 3
	PeProvedorAssinaturaAutorizacao                        // 4
)

var processoEmissaoCodigos = [...]string{"0", "1", "2", "3", "4"}

// String retorna o codigo do leiaute ("0".."4").
func (v ProcessoEmissao) String() string {
	if int(v) < 0 || int(v) >= len(processoEmissaoCodigos) {
		return ""
	}
	return processoEmissaoCodigos[v]
}

// ParseProcessoEmissao converte o codigo do leiaute no enum.
func ParseProcessoEmissao(s string) (ProcessoEmissao, error) {
	for i, c := range processoEmissaoCodigos {
		if c == s {
			return ProcessoEmissao(i), nil
		}
	}
	return PeAplicativoContribuinte, erroEnum("ProcessoEmissao", s)
}

// ===========================================================================
// IndIEDest (TindIEDest) -- tag indIEDest
// ===========================================================================

// IndIEDest indica a contribuicao do destinatario para o ICMS.
//
// Os codigos nao sao sequenciais: o terceiro membro vale "9", nao "3".
type IndIEDest int

const (
	InContribuinte    IndIEDest = iota // 1
	InIsento                           // 2
	InNaoContribuinte                  // 9
)

var indIEDestCodigos = [...]string{"1", "2", "9"}

// String retorna o codigo do leiaute.
func (v IndIEDest) String() string {
	if int(v) < 0 || int(v) >= len(indIEDestCodigos) {
		return ""
	}
	return indIEDestCodigos[v]
}

// ParseIndIEDest converte o codigo do leiaute no enum.
func ParseIndIEDest(s string) (IndIEDest, error) {
	for i, c := range indIEDestCodigos {
		if c == s {
			return IndIEDest(i), nil
		}
	}
	return InContribuinte, erroEnum("IndIEDest", s)
}

// ===========================================================================
// RegTribISSQN (TRegTribISSQN) -- tag cRegTrib do ISSQNtot
// ===========================================================================

// RegTribISSQN e o regime especial de tributacao do ISSQN.
//
// Atencao: RTISSNenhum e o ULTIMO membro e vale "0", nao vazio. O valor zero
// do tipo e RTISSMicroempresaMunicipal ("1").
type RegTribISSQN int

const (
	RTISSMicroempresaMunicipal  RegTribISSQN = iota // 1
	RTISSEstimativa                                 // 2
	RTISSSociedadeProfissionais                     // 3
	RTISSCooperativa                                // 4
	RTISSMEI                                        // 5
	RTISSMEEPP                                      // 6
	RTISSNenhum                                     // 0
)

var regTribISSQNCodigos = [...]string{"1", "2", "3", "4", "5", "6", "0"}

// String retorna o codigo do leiaute.
func (v RegTribISSQN) String() string {
	if int(v) < 0 || int(v) >= len(regTribISSQNCodigos) {
		return ""
	}
	return regTribISSQNCodigos[v]
}

// ParseRegTribISSQN converte o codigo do leiaute no enum.
func ParseRegTribISSQN(s string) (RegTribISSQN, error) {
	for i, c := range regTribISSQNCodigos {
		if c == s {
			return RegTribISSQN(i), nil
		}
	}
	return RTISSMicroempresaMunicipal, erroEnum("RegTribISSQN", s)
}

// ===========================================================================
// TpIntegra (TtpIntegra) -- tag tpIntegra do grupo card
// ===========================================================================

// TpIntegra indica se o processo de pagamento e integrado ao sistema de
// automacao do contribuinte.
type TpIntegra int

const (
	TiNaoInformado    TpIntegra = iota // vazio
	TiPagIntegrado                     // 1
	TiPagNaoIntegrado                  // 2
)

var tpIntegraCodigos = [...]string{"", "1", "2"}

// String retorna o codigo do leiaute.
func (v TpIntegra) String() string {
	if int(v) < 0 || int(v) >= len(tpIntegraCodigos) {
		return ""
	}
	return tpIntegraCodigos[v]
}

// ParseTpIntegra converte o codigo do leiaute no enum. String vazia e valida.
func ParseTpIntegra(s string) (TpIntegra, error) {
	for i, c := range tpIntegraCodigos {
		if c == s {
			return TpIntegra(i), nil
		}
	}
	return TiNaoInformado, erroEnum("TpIntegra", s)
}

// ===========================================================================
// CSOSNIcms (TCSOSNIcms) -- tag CSOSN
// ===========================================================================

// CSOSNIcms e o codigo de situacao da operacao no Simples Nacional, usado no
// lugar do CST quando o emitente e optante.
//
// Como em CSTIcms, o primeiro membro e o VAZIO -- um ICMS recem-criado nao
// declara CSOSN nenhum.
type CSOSNIcms int

const (
	CsosnVazio CSOSNIcms = iota // vazio
	Csosn101                    // 101
	Csosn102                    // 102
	Csosn103                    // 103
	Csosn201                    // 201
	Csosn202                    // 202
	Csosn203                    // 203
	Csosn300                    // 300
	Csosn400                    // 400
	Csosn500                    // 500
	Csosn900                    // 900
)

var csosnIcmsCodigos = [...]string{
	"", "101", "102", "103", "201", "202", "203", "300", "400", "500", "900",
}

// String retorna o codigo do leiaute.
func (v CSOSNIcms) String() string {
	if int(v) < 0 || int(v) >= len(csosnIcmsCodigos) {
		return ""
	}
	return csosnIcmsCodigos[v]
}

// ParseCSOSNIcms converte o codigo do leiaute no enum. String vazia e valida.
func ParseCSOSNIcms(s string) (CSOSNIcms, error) {
	for i, c := range csosnIcmsCodigos {
		if c == s {
			return CSOSNIcms(i), nil
		}
	}
	return CsosnVazio, erroEnum("CSOSNIcms", s)
}

// ===========================================================================
// IndIncentivo (TindIncentivo) -- tag indIncentivo do ISSQN
// ===========================================================================

// IndIncentivo indica se o servico e incentivado.
//
// Atencao: o "nao" vale "2", nao "0" -- como em IndISSRet, e diferente do
// Indicador comum.
type IndIncentivo int

const (
	IiSim IndIncentivo = iota // 1
	IiNao                     // 2
)

var indIncentivoCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v IndIncentivo) String() string {
	if int(v) < 0 || int(v) >= len(indIncentivoCodigos) {
		return ""
	}
	return indIncentivoCodigos[v]
}

// ParseIndIncentivo converte o codigo do leiaute no enum.
func ParseIndIncentivo(s string) (IndIncentivo, error) {
	for i, c := range indIncentivoCodigos {
		if c == s {
			return IndIncentivo(i), nil
		}
	}
	return IiSim, erroEnum("IndIncentivo", s)
}

// ===========================================================================
// SituacaoDFe (TSituacaoDFe) -- tag cSitNFe do resumo da Distribuicao DF-e
// ===========================================================================

// SituacaoDFe e a situacao de um documento no resumo que a Distribuicao
// DF-e devolve. Porte de TSituacaoDFe (ACBrDFe.Conversao.pas:214).
//
// NAO e o cStat: e um codigo proprio de quatro valores, e no schema
// `procNFe` ele nem vem pronto -- sai de um `case` sobre o cStat do
// protocolo (RetDistDFeInt.pas:620).
type SituacaoDFe int

const (
	SnAutorizado SituacaoDFe = iota // 1
	SnDenegado                      // 2
	SnCancelado                     // 3
	SnEncerrado                     // 4
	// _fimSituacaoDFe conta os membros declarados; nao e valor do leiaute.
	_fimSituacaoDFe
)

var situacaoDFeCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v SituacaoDFe) String() string {
	if int(v) < 0 || int(v) >= len(situacaoDFeCodigos) {
		return ""
	}
	return situacaoDFeCodigos[v]
}

// ParseSituacaoDFe converte o codigo do leiaute no enum.
func ParseSituacaoDFe(s string) (SituacaoDFe, error) {
	for i, c := range situacaoDFeCodigos {
		if c == s {
			return SituacaoDFe(i), nil
		}
	}
	return SnAutorizado, erroEnum("SituacaoDFe", s)
}

// ===========================================================================
// Modal (TModal) -- tag modal do CT-e e do MDF-e
// ===========================================================================

// Modal e o meio de transporte da prestacao.
// Porte de TModal (ACBrDFe.Conversao.pas:188).
//
// Vive aqui, e nao no packages/cte, porque e da unit COMPARTILHADA do ACBr
// e DOIS componentes o usam: o CT-e e o MDF-e. Declara-lo no CT-e
// obrigaria o MDF-e a redeclarar, e enum duplicado entre packages e
// exatamente o que o alinhamento por ordinal do TpNFDebito mostrou custar
// caro.
//
// ATENCAO: o ACBr mantem um TpcteModal ANTIGO em PCNComum/pcnConversao.pas,
// marcado `deprecated` com a mensagem "Use o metodo da Unit
// ACBrDFe.Conversao.pas". A declaracao viva e a de la, e e esta.
type Modal int

const (
	MdRodoviario  Modal = iota // 01
	MdAereo                    // 02
	MdAquaviario               // 03
	MdFerroviario              // 04
	MdDutoviario               // 05
	MdMultimodal               // 06
	// _fimModal conta os membros declarados; nao e valor do leiaute.
	_fimModal
)

var modalCodigos = [...]string{"01", "02", "03", "04", "05", "06"}

// modalDescricoes e o texto de impressao do DACTE/DAMDFE, nao o do XML.
var modalDescricoes = [...]string{
	"RODOVIÁRIO", "AÉREO", "AQUAVIÁRIO", "FERROVIÁRIO",
	"DUTOVIÁRIO", "MULTIMODAL",
}

// String retorna o codigo do leiaute, com DOIS digitos.
//
// A largura importa: o schema declara o campo com dois caracteres, e
// emitir "1" em vez de "01" e rejeicao.
func (v Modal) String() string {
	if int(v) < 0 || int(v) >= len(modalCodigos) {
		return ""
	}
	return modalCodigos[v]
}

// Descricao retorna o texto de impressao.
func (v Modal) Descricao() string {
	if int(v) < 0 || int(v) >= len(modalDescricoes) {
		return ""
	}
	return modalDescricoes[v]
}

// ParseModal converte o codigo do leiaute no enum.
func ParseModal(s string) (Modal, error) {
	for i, c := range modalCodigos {
		if c == s {
			return Modal(i), nil
		}
	}
	return MdRodoviario, erroEnum("Modal", s)
}

// ===========================================================================
// UnidMed, UnidTransp e UnidCarga -- compartilhados por CT-e e MDF-e
// ===========================================================================
//
// Os tres vem de ACBrDFe.Conversao.pas, e o proprio ACBr marca dois deles
// com o comentario `// CTe e MDFe` (:459 e :465). Vivem aqui pela mesma
// razao do Modal: declara-los no packages/cte obrigaria o MDF-e a
// redeclarar.

// UnidMed e a unidade de medida da carga. Porte de TUnidMed (:474).
//
// ATENCAO a largura: os codigos tem DOIS digitos e comecam em "00". Um
// enum de seis membros numerados de zero convida a tratar como inteiro, e
// o schema recusa "0" no lugar de "00".
type UnidMed int

const (
	UM3      UnidMed = iota // 00
	UKG                     // 01
	UTON                    // 02
	UUnidade                // 03
	ULitros                 // 04
	UMMBTU                  // 05
	// _fimUnidMed conta os membros declarados; nao e valor do leiaute.
	_fimUnidMed
)

var unidMedCodigos = [...]string{"00", "01", "02", "03", "04", "05"}

// unidMedDescricoes e o texto de impressao do DACTE/DAMDFE, copiado
// LITERAL de TUnidMedDescricaoArrayStrings (ACBrDFe.Conversao.pas:480).
//
// Duas sao abreviacoes: "UND" e "LT", nao "UNIDADE" e "LITROS". Escrever
// por extenso parece melhoria e nao e -- o DACTE tem campo de largura fixa,
// e o que cabe e a abreviacao.
var unidMedDescricoes = [...]string{
	"M3", "KG", "TON", "UND", "LT", "MMBTU",
}

// String retorna o codigo do leiaute, com dois digitos.
func (v UnidMed) String() string { return codigoEmDFe(unidMedCodigos[:], int(v)) }

// Descricao retorna o texto de impressao.
func (v UnidMed) Descricao() string { return codigoEmDFe(unidMedDescricoes[:], int(v)) }

// ParseUnidMed converte o codigo do leiaute no enum.
func ParseUnidMed(s string) (UnidMed, error) {
	for i, c := range unidMedCodigos {
		if c == s {
			return UnidMed(i), nil
		}
	}
	return UM3, erroEnum("UnidMed", s)
}

// UnidTransp e o tipo da unidade de transporte. Porte de TUnidTransp (:451).
//
// Diferente do UnidMed, aqui os codigos tem UM digito e comecam em "1".
// As duas formas convivem no mesmo documento.
type UnidTransp int

const (
	UtRodoTracao  UnidTransp = iota // 1
	UtRodoReboque                   // 2
	UtNavio                         // 3
	UtBalsa                         // 4
	UtAeronave                      // 5
	UtVagao                         // 6
	UtOutros                        // 7
	// _fimUnidTransp conta os membros declarados; nao e valor do leiaute.
	_fimUnidTransp
)

var unidTranspCodigos = [...]string{"1", "2", "3", "4", "5", "6", "7"}

// String retorna o codigo do leiaute.
func (v UnidTransp) String() string { return codigoEmDFe(unidTranspCodigos[:], int(v)) }

// ParseUnidTransp converte o codigo do leiaute no enum.
func ParseUnidTransp(s string) (UnidTransp, error) {
	for i, c := range unidTranspCodigos {
		if c == s {
			return UnidTransp(i), nil
		}
	}
	return UtRodoTracao, erroEnum("UnidTransp", s)
}

// UnidCarga e o tipo da unidade de carga. Porte de TUnidCarga (:460).
type UnidCarga int

const (
	UcContainer UnidCarga = iota // 1
	UcULD                        // 2
	UcPallet                     // 3
	UcOutros                     // 4
	// _fimUnidCarga conta os membros declarados; nao e valor do leiaute.
	_fimUnidCarga
)

var unidCargaCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v UnidCarga) String() string { return codigoEmDFe(unidCargaCodigos[:], int(v)) }

// ParseUnidCarga converte o codigo do leiaute no enum.
func ParseUnidCarga(s string) (UnidCarga, error) {
	for i, c := range unidCargaCodigos {
		if c == s {
			return UnidCarga(i), nil
		}
	}
	return UcContainer, erroEnum("UnidCarga", s)
}

// codigoEmDFe devolve o codigo da posicao i, ou vazio se fora de faixa.
func codigoEmDFe(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}
