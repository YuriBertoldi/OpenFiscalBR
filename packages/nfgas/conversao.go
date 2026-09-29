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

package nfgas

import "fmt"

// Enums de dominio da NFGas. Porte de ACBrNFGas.Conversao.pas.
//
// Cada enum expoe String() com o codigo exato do leiaute e ParseXxx que
// devolve error -- os StrToXxx do Delphi levantam EACBrException.

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %s nao aceita %q", ErrEnumInvalido, enum, valor)
}

// codigoEm devolve o codigo da posicao i, ou vazio se fora de faixa.
func codigoEm(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}

// indiceDe devolve a posicao do codigo na tabela.
func indiceDe(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if c == s {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// ===========================================================================
// IndIEDest (TindIEDest)
// ===========================================================================

// IndIEDest indica o tipo de inscricao estadual do destinatario.
//
// Atencao: este campo existe no formato .ini e NAO e lido do XML pelo ACBr
// -- ver xml_reader.go, lerDest.
type IndIEDest int

const (
	InContribuinte    IndIEDest = iota // 1
	InIsento                           // 2
	InNaoContribuinte                  // 9
)

var indIEDestCodigos = [...]string{"1", "2", "9"}

// String retorna o codigo do leiaute.
func (v IndIEDest) String() string { return codigoEm(indIEDestCodigos[:], int(v)) }

// ParseIndIEDest converte o codigo do leiaute no enum.
func ParseIndIEDest(s string) (IndIEDest, error) {
	i, err := indiceDe(indIEDestCodigos[:], s, "IndIEDest")
	return IndIEDest(i), err
}

// ===========================================================================
// SiteAutorizador (TSiteAutorizador)
// ===========================================================================

// SiteAutorizador identifica o site autorizador, de 0 a 9.
// Ocupa a posicao 36 da chave de acesso da NFGas.
type SiteAutorizador int

const (
	Sa0 SiteAutorizador = iota
	Sa1
	Sa2
	Sa3
	Sa4
	Sa5
	Sa6
	Sa7
	Sa8
	Sa9
)

var siteAutorizadorCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}

// String retorna o codigo do leiaute.
func (v SiteAutorizador) String() string { return codigoEm(siteAutorizadorCodigos[:], int(v)) }

// ParseSiteAutorizador converte o codigo do leiaute no enum.
// Porte de StrToSiteAutorizator (o nome tem typo no original).
func ParseSiteAutorizador(s string) (SiteAutorizador, error) {
	i, err := indiceDe(siteAutorizadorCodigos[:], s, "SiteAutorizador")
	return SiteAutorizador(i), err
}

// ===========================================================================
// IndOrigemQtd (TindOrigemQtd)
// ===========================================================================

// IndOrigemQtd indica a origem da quantidade faturada.
type IndOrigemQtd int

const (
	IoMedia         IndOrigemQtd = iota // 1
	IoMedido                            // 2
	IoContatada                         // 3
	IoCalculada                         // 4
	IoCusto                             // 5
	IoSemQuantidade                     // 6
)

var indOrigemQtdCodigos = [...]string{"1", "2", "3", "4", "5", "6"}

// String retorna o codigo do leiaute.
func (v IndOrigemQtd) String() string { return codigoEm(indOrigemQtdCodigos[:], int(v)) }

// ParseIndOrigemQtd converte o codigo do leiaute no enum.
func ParseIndOrigemQtd(s string) (IndOrigemQtd, error) {
	i, err := indiceDe(indOrigemQtdCodigos[:], s, "IndOrigemQtd")
	return IndOrigemQtd(i), err
}

// ===========================================================================
// UMed (TuMed) e UMedItem (TuMedItem)
// ===========================================================================

// UMed e a unidade de medida da medicao e do historico de consumo.
// Tem um unico membro. NAO confundir com UMedItem, usada no produto.
type UMed int

const (
	Umm3 UMed = iota // 1 = m3
)

var uMedCodigos = [...]string{"1"}
var uMedDescricoes = [...]string{"m3"}

// String retorna o codigo do leiaute.
func (v UMed) String() string { return codigoEm(uMedCodigos[:], int(v)) }

// Descricao retorna a unidade por extenso. Porte de uMedToDesc.
func (v UMed) Descricao() string { return codigoEm(uMedDescricoes[:], int(v)) }

// ParseUMed converte o codigo do leiaute no enum.
func ParseUMed(s string) (UMed, error) {
	i, err := indiceDe(uMedCodigos[:], s, "UMed")
	return UMed(i), err
}

// UMedItem e a unidade de medida do produto. Tipo distinto de UMed: admite
// tambem "Unidade", que a medicao nao admite.
type UMedItem int

const (
	Umim3      UMedItem = iota // 1 = m3
	UmiUnidade                 // 2 = Unidade
)

var uMedItemCodigos = [...]string{"1", "2"}
var uMedItemDescricoes = [...]string{"m3", "Unidade"}

// String retorna o codigo do leiaute.
func (v UMedItem) String() string { return codigoEm(uMedItemCodigos[:], int(v)) }

// Descricao retorna a unidade por extenso. Porte de uMedItemToDesc.
func (v UMedItem) Descricao() string { return codigoEm(uMedItemDescricoes[:], int(v)) }

// ParseUMedItem converte o codigo do leiaute no enum.
func ParseUMedItem(s string) (UMedItem, error) {
	i, err := indiceDe(uMedItemCodigos[:], s, "UMedItem")
	return UMedItem(i), err
}

// ===========================================================================
// TpMotNaoLeitura (TtpMotNaoLeitura)
// ===========================================================================

// TpMotNaoLeitura e o motivo da nao leitura do medidor.
//
// Cuidado: no Delphi os membros tem prefixo tm*, que colide com os de
// TtpMedidor. Aqui os tipos sao distintos e o compilador separa.
type TpMotNaoLeitura int

const (
	TmConsumidor    TpMotNaoLeitura = iota // 1
	TmDistribuidora                        // 2
	TmIndependente                         // 3
)

var tpMotNaoLeituraCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpMotNaoLeitura) String() string { return codigoEm(tpMotNaoLeituraCodigos[:], int(v)) }

// ParseTpMotNaoLeitura converte o codigo do leiaute no enum.
func ParseTpMotNaoLeitura(s string) (TpMotNaoLeitura, error) {
	i, err := indiceDe(tpMotNaoLeituraCodigos[:], s, "TpMotNaoLeitura")
	return TpMotNaoLeitura(i), err
}

// ===========================================================================
// FinalidadeNFGas (TFinalidadeNFGas)
// ===========================================================================

// FinalidadeNFGas e a finalidade de emissao do documento.
//
// Nao ha codigos "1" e "2": o leiaute salta de 0 para 3.
type FinalidadeNFGas int

const (
	FnNormal       FinalidadeNFGas = iota // 0
	FnSubstituicao                        // 3
)

var finalidadeCodigos = [...]string{"0", "3"}

// String retorna o codigo do leiaute.
func (v FinalidadeNFGas) String() string { return codigoEm(finalidadeCodigos[:], int(v)) }

// ParseFinalidadeNFGas converte o codigo do leiaute no enum.
func ParseFinalidadeNFGas(s string) (FinalidadeNFGas, error) {
	i, err := indiceDe(finalidadeCodigos[:], s, "FinalidadeNFGas")
	return FinalidadeNFGas(i), err
}

// ===========================================================================
// TpInstalacao (TtpInstalacao)
// ===========================================================================

// TpInstalacao e o tipo de instalacao do consumidor.
type TpInstalacao int

const (
	TiCativo            TpInstalacao = iota // 1
	TiLivre                                 // 2
	TiParcialmenteLivre                     // 3
)

var tpInstalacaoCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpInstalacao) String() string { return codigoEm(tpInstalacaoCodigos[:], int(v)) }

// ParseTpInstalacao converte o codigo do leiaute no enum.
// Porte de StrToInstalacao.
func ParseTpInstalacao(s string) (TpInstalacao, error) {
	i, err := indiceDe(tpInstalacaoCodigos[:], s, "TpInstalacao")
	return TpInstalacao(i), err
}

// ===========================================================================
// TpClasse (TtpClasse)
// ===========================================================================

// TpClasse e a classe de consumo da instalacao.
type TpClasse int

const (
	TcComercial     TpClasse = iota // 01
	TcIndustrial                    // 02
	TcResidencial                   // 03
	TcTermico                       // 04
	TcVeicularPosto                 // 05
	TcVeicularFrota                 // 06
	TcGNC                           // 07
	TcGNL                           // 08
	TcCogeracao                     // 09
	TcRefinaria                     // 10
	TcOutros                        // 99
)

var tpClasseCodigos = [...]string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "99"}

// String retorna o codigo de dois digitos do leiaute.
func (v TpClasse) String() string { return codigoEm(tpClasseCodigos[:], int(v)) }

// ParseTpClasse converte o codigo do leiaute no enum. Porte de StrToClasse.
func ParseTpClasse(s string) (TpClasse, error) {
	i, err := indiceDe(tpClasseCodigos[:], s, "TpClasse")
	return TpClasse(i), err
}

// ===========================================================================
// MotSub (TmotSub)
// ===========================================================================

// MotSub e o motivo da substituicao do documento.
type MotSub int

const (
	MsErroLeitura       MotSub = iota // 01
	MsErroPreco                       // 02
	MsDecisaoJudicial                 // 03
	MsErroCadastral                   // 04
	MsErroTributacao                  // 05
	MsDecisaoReguladora               // 06
)

var motSubCodigos = [...]string{"01", "02", "03", "04", "05", "06"}

// String retorna o codigo de dois digitos do leiaute.
func (v MotSub) String() string { return codigoEm(motSubCodigos[:], int(v)) }

// ParseMotSub converte o codigo do leiaute no enum.
func ParseMotSub(s string) (MotSub, error) {
	i, err := indiceDe(motSubCodigos[:], s, "MotSub")
	return MotSub(i), err
}

// ===========================================================================
// VolContrat (TVolContrat)
// ===========================================================================

// VolContrat e o tipo de volume contratado.
type VolContrat int

const (
	VcDemandaMinima     VolContrat = iota // 1
	VcMontanteUso                         // 2
	VcEncargoCapacidade                   // 3
	VcVolumeContratado                    // 4
)

var volContratCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v VolContrat) String() string { return codigoEm(volContratCodigos[:], int(v)) }

// ParseVolContrat converte o codigo do leiaute no enum.
func ParseVolContrat(s string) (VolContrat, error) {
	i, err := indiceDe(volContratCodigos[:], s, "VolContrat")
	return VolContrat(i), err
}

// ===========================================================================
// TpEqp (TtpEqp) e TpMedidor (TtpMedidor)
// ===========================================================================

// TpEqp e o tipo do equipamento de medicao.
type TpEqp int

const (
	TeMedidor   TpEqp = iota // 1
	TeConversor              // 2
)

var tpEqpCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v TpEqp) String() string { return codigoEm(tpEqpCodigos[:], int(v)) }

// ParseTpEqp converte o codigo do leiaute no enum.
func ParseTpEqp(s string) (TpEqp, error) {
	i, err := indiceDe(tpEqpCodigos[:], s, "TpEqp")
	return TpEqp(i), err
}

// TpMedidor e a tecnologia do medidor.
type TpMedidor int

const (
	TmTurbina     TpMedidor = iota // 1
	TmRotativo                     // 2
	TmDiafragma                    // 3
	TmUltrasonico                  // 4
)

var tpMedidorCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v TpMedidor) String() string { return codigoEm(tpMedidorCodigos[:], int(v)) }

// ParseTpMedidor converte o codigo do leiaute no enum.
func ParseTpMedidor(s string) (TpMedidor, error) {
	i, err := indiceDe(tpMedidorCodigos[:], s, "TpMedidor")
	return TpMedidor(i), err
}

// ===========================================================================
// TpFaixaCons (TtpFaixaCons)
// ===========================================================================

// TpFaixaCons e o tipo da faixa de consumo da tarifa.
//
// Cuidado: no Delphi os membros tem prefixo tf*, que colide com os de
// TtpFat. Aqui os tipos sao distintos.
type TpFaixaCons int

const (
	TfFixa  TpFaixaCons = iota // 1
	TfMedia                    // 2
)

var tpFaixaConsCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v TpFaixaCons) String() string { return codigoEm(tpFaixaConsCodigos[:], int(v)) }

// ParseTpFaixaCons converte o codigo do leiaute no enum.
func ParseTpFaixaCons(s string) (TpFaixaCons, error) {
	i, err := indiceDe(tpFaixaConsCodigos[:], s, "TpFaixaCons")
	return TpFaixaCons(i), err
}

// ===========================================================================
// TpProc (TtpProc)
// ===========================================================================

// TpProc e a origem do processo referenciado.
type TpProc int

const (
	TpProcAdmEstadual TpProc = iota // 0
	TpJusticaFederal                // 1
	TpJusticaEstadual               // 2
	TpProcAdmMunicial               // 3 -- "Municial" e o nome no ACBr
	TpProcAdmFederal                // 4
	TpProcon                        // 5
)

var tpProcCodigos = [...]string{"0", "1", "2", "3", "4", "5"}

// String retorna o codigo do leiaute.
func (v TpProc) String() string { return codigoEm(tpProcCodigos[:], int(v)) }

// ParseTpProc converte o codigo do leiaute no enum.
func ParseTpProc(s string) (TpProc, error) {
	i, err := indiceDe(tpProcCodigos[:], s, "TpProc")
	return TpProc(i), err
}

// ===========================================================================
// TpFat (TtpFat)
// ===========================================================================

// TpFat e o tipo de faturamento do documento.
type TpFat int

const (
	TfNormal    TpFat = iota // 1
	TfAgregado               // 2
	TfAgregador              // 3
)

var tpFatCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpFat) String() string { return codigoEm(tpFatCodigos[:], int(v)) }

// ParseTpFat converte o codigo do leiaute no enum.
func ParseTpFat(s string) (TpFat, error) {
	i, err := indiceDe(tpFatCodigos[:], s, "TpFat")
	return TpFat(i), err
}

// ===========================================================================
// DeterminacaoBaseIcms (TDeterminacaoBaseIcms)
// ===========================================================================

// DeterminacaoBaseIcms e a modalidade de determinacao da base de calculo do
// ICMS.
type DeterminacaoBaseIcms int

const (
	DbiMargemValorAgregado DeterminacaoBaseIcms = iota // 0
	DbiPauta                                           // 1
	DbiPrecoTabelado                                   // 2
	DbiValorOperacao                                   // 3
	DbiNenhum                                          // vazio
)

var modBCCodigos = [...]string{"0", "1", "2", "3", ""}

var modBCDescricoes = [...]string{
	"0 - Margem Valor Agregado (%)",
	"1 - Pauta (Valor)",
	"2 - Preco Tabelado Max. (valor)",
	"3 - Valor da operacao",
	"",
}

// String retorna o codigo do leiaute. Porte de modBCToStr.
func (v DeterminacaoBaseIcms) String() string { return codigoEm(modBCCodigos[:], int(v)) }

// Descricao retorna o texto de exibicao. Porte de modBCToStrTagPosText.
func (v DeterminacaoBaseIcms) Descricao() string { return codigoEm(modBCDescricoes[:], int(v)) }

// ParseDeterminacaoBaseIcms converte o codigo do leiaute no enum.
// Vazio e valido e resulta em DbiNenhum. Porte de StrTomodBC.
func ParseDeterminacaoBaseIcms(s string) (DeterminacaoBaseIcms, error) {
	i, err := indiceDe(modBCCodigos[:], s, "DeterminacaoBaseIcms")
	if err != nil {
		return DbiNenhum, err
	}
	return DeterminacaoBaseIcms(i), nil
}

// ===========================================================================
// DeterminacaoBaseIcmsST (TDeterminacaoBaseIcmsST)
// ===========================================================================

// DeterminacaoBaseIcmsST e a modalidade de determinacao da base de calculo
// do ICMS por substituicao tributaria.
//
// DbisNenhum vale -1 para reproduzir o sentinela TDeterminacaoBaseIcmsST(-1)
// que o ACBr usa para representar o campo ausente -- ele nao e um membro
// declarado do enum no Delphi.
type DeterminacaoBaseIcmsST int

const (
	DbisPrecoTabelado       DeterminacaoBaseIcmsST = iota // 0
	DbisListaNegativa                                     // 1
	DbisListaPositiva                                     // 2
	DbisListaNeutra                                       // 3
	DbisMargemValorAgregado                               // 4
	DbisPauta                                             // 5
	DbisValordaOperacao                                   // 6

	DbisNenhum DeterminacaoBaseIcmsST = -1 // vazio
)

var modBCSTCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6"}

var modBCSTDescricoes = [...]string{
	"0 - Preco tabelado ou maximo sugerido",
	"1 - Lista Negativa (valor)",
	"2 - Lista Positiva (valor)",
	"3 - Lista Neutra (valor)",
	"4 - Margem Valor Agregado (%)",
	"5 - Pauta (valor)",
	"6 - Valor da Operacao",
}

// String retorna o codigo do leiaute. Porte de modBCSTToStr.
func (v DeterminacaoBaseIcmsST) String() string { return codigoEm(modBCSTCodigos[:], int(v)) }

// Descricao retorna o texto de exibicao. Porte de modBCSTToStrTagPosText.
func (v DeterminacaoBaseIcmsST) Descricao() string { return codigoEm(modBCSTDescricoes[:], int(v)) }

// ParseDeterminacaoBaseIcmsST converte o codigo do leiaute no enum.
// Vazio e valido e resulta em DbisNenhum. Porte de StrTomodBCST.
func ParseDeterminacaoBaseIcmsST(s string) (DeterminacaoBaseIcmsST, error) {
	if s == "" {
		return DbisNenhum, nil
	}
	i, err := indiceDe(modBCSTCodigos[:], s, "DeterminacaoBaseIcmsST")
	if err != nil {
		return DbisNenhum, err
	}
	return DeterminacaoBaseIcmsST(i), nil
}

// ===========================================================================
// MotivoDesoneracaoICMS (TMotivoDesoneracaoICMS)
// ===========================================================================

// MotivoDesoneracaoICMS e o motivo da desoneracao do ICMS.
//
// Os codigos nao sao sequenciais: saltam de 12 para 16 e de 16 para 90.
// MdiNenhum vale -1, reproduzindo o sentinela TMotivoDesoneracaoICMS(-1).
type MotivoDesoneracaoICMS int

const (
	MdiTaxi                  MotivoDesoneracaoICMS = iota // 1
	MdiDeficienteFisico                                   // 2
	MdiProdutorAgropecuario                               // 3
	MdiFrotistaLocadora                                   // 4
	MdiDiplomaticoConsular                                // 5
	MdiAmazoniaLivreComercio                              // 6
	MdiSuframa                                            // 7
	MdiVendaOrgaosPublicos                                // 8
	MdiOutros                                             // 9
	MdiDeficienteCondutor                                 // 10
	MdiDeficienteNaoCondutor                              // 11
	MdiOrgaoFomento                                       // 12
	MdiOlimpiadaRio2016                                   // 16
	MdiSolicitadoFisco                                    // 90

	MdiNenhum MotivoDesoneracaoICMS = -1 // vazio
)

var motDesICMSCodigos = [...]string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "16", "90",
}

var motDesICMSDescricoes = [...]string{
	"1 - Taxi",
	"2 - Deficiente Fisico",
	"3 - Produtor Agropecuario",
	"4 - Frotista/Locadora",
	"5 - Diplomatico/Consular",
	"6 - Utilit./Motos da Am./Areas Livre Com.",
	"7 - SUFRAMA",
	"8 - Venda a Orgaos Publicos",
	"9 - Outros",
	"10 - Deficiente Condutor",
	"11 - Deficiente nao Condutor",
	"12 - Orgao Fomento",
	"16 - Olimpiadas Rio 2016",
	"90 - Solicitado pelo Fisco",
}

// String retorna o codigo do leiaute. Porte de motDesICMSToStr.
func (v MotivoDesoneracaoICMS) String() string { return codigoEm(motDesICMSCodigos[:], int(v)) }

// Descricao retorna o texto de exibicao. Porte de motDesICMSToStrTagPosText.
func (v MotivoDesoneracaoICMS) Descricao() string { return codigoEm(motDesICMSDescricoes[:], int(v)) }

// ParseMotivoDesoneracaoICMS converte o codigo do leiaute no enum.
// Vazio e valido e resulta em MdiNenhum. Porte de StrTomotDesICMS.
func ParseMotivoDesoneracaoICMS(s string) (MotivoDesoneracaoICMS, error) {
	if s == "" {
		return MdiNenhum, nil
	}
	i, err := indiceDe(motDesICMSCodigos[:], s, "MotivoDesoneracaoICMS")
	if err != nil {
		return MdiNenhum, err
	}
	return MotivoDesoneracaoICMS(i), nil
}

// ===========================================================================
// TipoEvento (TACBrTipoEvento, restrito aos eventos da NFGas)
// ===========================================================================

// TipoEvento e o tipo de evento aceito pela NFGas.
//
// A NFGas usa 110111 para cancelamento -- e NAO 110111/110110 como a NFe.
// Porte de StrToTpEventoNFGas, que o ACBr registra em tempo de execucao
// via RegisterStrToTpEventoDFe(..., 'NFGas').
type TipoEvento int

const (
	TeNaoMapeado              TipoEvento = iota // -99999
	TeCancelamento                              // 110111
	TeAutorizadoSubstituicao                    // 240140
	TeAutorizadoAjuste                          // 240150
	TeLiberacaoPrazoCancelado                   // 240170
)

var tipoEventoCodigos = [...]string{"-99999", "110111", "240140", "240150", "240170"}

var tipoEventoDescricoes = [...]string{
	"Nao Mapeado",
	"Cancelamento",
	"Autorizado Substituicao",
	"Autorizado Ajuste",
	"Liberacao Prazo Cancelado",
}

// String retorna o codigo do evento.
func (v TipoEvento) String() string { return codigoEm(tipoEventoCodigos[:], int(v)) }

// Descricao retorna a descricao do evento.
func (v TipoEvento) Descricao() string { return codigoEm(tipoEventoDescricoes[:], int(v)) }

// ParseTipoEvento converte o codigo do evento no enum.
func ParseTipoEvento(s string) (TipoEvento, error) {
	i, err := indiceDe(tipoEventoCodigos[:], s, "TipoEvento")
	return TipoEvento(i), err
}
