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

package nfag

import "fmt"

// Enums do leiaute da NFAg, com o CODIGO exato que vai no XML.
// Porte de ACBrNFAg.Conversao.pas. Cada StrToXxx do Delphi (que da raise)
// vira ParseXxx devolvendo (T, error).

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %q nao e um %s do leiaute", ErrEnumInvalido, valor, enum)
}

func codigoEm(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}

func indiceDe(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if c == s {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// ===========================================================================
// SiteAutorizador (TSiteAutorizador)
// ===========================================================================

// SiteAutorizador e o site autorizador da chave de acesso (posicao 36).
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

// String retorna o digito do site autorizador.
func (v SiteAutorizador) String() string { return codigoEm(siteAutorizadorCodigos[:], int(v)) }

// ParseSiteAutorizador converte o digito no enum.
func ParseSiteAutorizador(s string) (SiteAutorizador, error) {
	i, err := indiceDe(siteAutorizadorCodigos[:], s, "SiteAutorizador")
	return SiteAutorizador(i), err
}

// ===========================================================================
// FinalidadeNFAg (TFinalidadeNFAg)
// ===========================================================================

// FinalidadeNFAg e a finalidade de emissao (tag finNFAg).
type FinalidadeNFAg int

const (
	FnNormal       FinalidadeNFAg = iota // 0
	FnSubstituicao                       // 3
)

var finalidadeCodigos = [...]string{"0", "3"}

// String retorna o codigo do leiaute.
func (v FinalidadeNFAg) String() string { return codigoEm(finalidadeCodigos[:], int(v)) }

// ParseFinalidadeNFAg converte o codigo no enum.
func ParseFinalidadeNFAg(s string) (FinalidadeNFAg, error) {
	i, err := indiceDe(finalidadeCodigos[:], s, "FinalidadeNFAg")
	return FinalidadeNFAg(i), err
}

// ===========================================================================
// TpFat (TtpFat)
// ===========================================================================

// TpFat e o tipo de faturamento (tag tpFat).
type TpFat int

const (
	TfNormal   TpFat = iota // 1
	TfTerceiro              // 2
	TfConjunto              // 3
)

var tpFatCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpFat) String() string { return codigoEm(tpFatCodigos[:], int(v)) }

// ParseTpFat converte o codigo no enum.
func ParseTpFat(s string) (TpFat, error) {
	i, err := indiceDe(tpFatCodigos[:], s, "TpFat")
	return TpFat(i), err
}

// ===========================================================================
// TpFaixaCons (TtpFaixaCons)
// ===========================================================================

// TpFaixaCons e a faixa prevista de consumo (tag tpFaixaCons).
type TpFaixaCons int

const (
	TfcMinimo TpFaixaCons = iota // 1
	TfcMedio                     // 2
	TfcMaximo                    // 3
)

var tpFaixaConsCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpFaixaCons) String() string { return codigoEm(tpFaixaConsCodigos[:], int(v)) }

// ParseTpFaixaCons converte o codigo no enum.
func ParseTpFaixaCons(s string) (TpFaixaCons, error) {
	i, err := indiceDe(tpFaixaConsCodigos[:], s, "TpFaixaCons")
	return TpFaixaCons(i), err
}

// ===========================================================================
// IndIEDest (TindIEDest)
// ===========================================================================

// IndIEDest e o indicador da IE do destinatario. So aparece no .ini.
type IndIEDest int

const (
	InContribuinte    IndIEDest = iota // 1
	InIsento                           // 2
	InNaoContribuinte                  // 9
)

var indIEDestCodigos = [...]string{"1", "2", "9"}

// String retorna o codigo do leiaute.
func (v IndIEDest) String() string { return codigoEm(indIEDestCodigos[:], int(v)) }

// ParseIndIEDest converte o codigo no enum.
func ParseIndIEDest(s string) (IndIEDest, error) {
	i, err := indiceDe(indIEDestCodigos[:], s, "IndIEDest")
	return IndIEDest(i), err
}

// ===========================================================================
// TpLigacao (TtpLigacao)
// ===========================================================================

// TpLigacao e o tipo da ligacao (tag tpLigacao): agua, esgoto ou ambos.
type TpLigacao int

const (
	TlAgua       TpLigacao = iota // 1
	TlEsgoto                      // 2
	TlAguaEsgoto                  // 3
)

var tpLigacaoCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpLigacao) String() string { return codigoEm(tpLigacaoCodigos[:], int(v)) }

// ParseTpLigacao converte o codigo no enum.
func ParseTpLigacao(s string) (TpLigacao, error) {
	i, err := indiceDe(tpLigacaoCodigos[:], s, "TpLigacao")
	return TpLigacao(i), err
}

// ===========================================================================
// MotSub (TmotSub)
// ===========================================================================

// MotSub e o motivo da substituicao (tag motSub).
type MotSub int

const (
	MsErroLeitura     MotSub = iota // 01
	MsErroPreco                     // 02
	MsDecisaoJudicial               // 03
	MsErroCadastral                 // 04
	MsErroTributacao                // 05
)

var motSubCodigos = [...]string{"01", "02", "03", "04", "05"}

// String retorna o codigo do leiaute.
func (v MotSub) String() string { return codigoEm(motSubCodigos[:], int(v)) }

// ParseMotSub converte o codigo no enum.
func ParseMotSub(s string) (MotSub, error) {
	i, err := indiceDe(motSubCodigos[:], s, "MotSub")
	return MotSub(i), err
}

// ===========================================================================
// UMed (TuMed)
// ===========================================================================

// UMed e a unidade de medida generica declarada em TuMed. ATENCAO: os
// membros do ACBr sao umkW/umkWh (copiados da NF3e) com codigos "1"/"2";
// o leiaute da NFAg nao usa esse enum em nenhuma tag -- ele existe na
// unit de conversao e nada o referencia. Mantido por fidelidade.
type UMed int

const (
	UmkW  UMed = iota // 1
	UmkWh             // 2
)

var uMedCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v UMed) String() string { return codigoEm(uMedCodigos[:], int(v)) }

// ParseUMed converte o codigo no enum.
func ParseUMed(s string) (UMed, error) {
	i, err := indiceDe(uMedCodigos[:], s, "UMed")
	return UMed(i), err
}

// ===========================================================================
// MotDifTarif (TmotDifTarif)
// ===========================================================================

// MotDifTarif e o motivo de tarifa diferente. Declarado em
// ACBrNFAg.Conversao.pas e nao referenciado por leitor/gerador -- mantido
// por fidelidade a unit de origem.
type MotDifTarif int

const (
	MdtDecisaoJudicial      MotDifTarif = iota // 01
	MdtDecisaoDistribuidora                    // 02
	MdtDesconto                                // 03
	MdtAlteracao                               // 04
)

var motDifTarifCodigos = [...]string{"01", "02", "03", "04"}

// String retorna o codigo do leiaute.
func (v MotDifTarif) String() string { return codigoEm(motDifTarifCodigos[:], int(v)) }

// ParseMotDifTarif converte o codigo no enum.
func ParseMotDifTarif(s string) (MotDifTarif, error) {
	i, err := indiceDe(motDifTarifCodigos[:], s, "MotDifTarif")
	return MotDifTarif(i), err
}

// ===========================================================================
// IndOrigemQtd (TindOrigemQtd)
// ===========================================================================

// IndOrigemQtd e o indicador da origem da quantidade faturada
// (tag indOrigemQtd).
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

// ParseIndOrigemQtd converte o codigo no enum.
func ParseIndOrigemQtd(s string) (IndOrigemQtd, error) {
	i, err := indiceDe(indOrigemQtdCodigos[:], s, "IndOrigemQtd")
	return IndOrigemQtd(i), err
}

// ===========================================================================
// TpMotNaoLeitura (TtpMotNaoLeitura)
// ===========================================================================

// TpMotNaoLeitura e o motivo da nao leitura do medidor
// (tag tpMotNaoLeitura).
type TpMotNaoLeitura int

const (
	TmConsumidor    TpMotNaoLeitura = iota // 1
	TmDistribuidora                        // 2
	TmIndependente                         // 3
)

var tpMotNaoLeituraCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TpMotNaoLeitura) String() string { return codigoEm(tpMotNaoLeituraCodigos[:], int(v)) }

// ParseTpMotNaoLeitura converte o codigo no enum.
func ParseTpMotNaoLeitura(s string) (TpMotNaoLeitura, error) {
	i, err := indiceDe(tpMotNaoLeituraCodigos[:], s, "TpMotNaoLeitura")
	return TpMotNaoLeitura(i), err
}

// ===========================================================================
// TpGrMed (TtpGrMed)
// ===========================================================================

// TpGrMed e o tipo de grandeza medida (tag tpGrMed).
type TpGrMed int

const (
	TgmAguaTratada                  TpGrMed = iota // 01
	TgmEsgotoTratado                               // 02
	TgmEsgotoColetado                              // 03
	TgmResidoSolidos                               // 04
	TgmAguaReuso                                   // 05
	TgmEsgotoEstatico                              // 06
	TgmCaptacaoAguaBruta                           // 07
	TgmRecebimentoTratamentoChorume                // 08
	TgmOutros                                      // 99
)

var tpGrMedCodigos = [...]string{"01", "02", "03", "04", "05", "06", "07", "08", "99"}

// String retorna o codigo do leiaute.
func (v TpGrMed) String() string { return codigoEm(tpGrMedCodigos[:], int(v)) }

// ParseTpGrMed converte o codigo no enum.
func ParseTpGrMed(s string) (TpGrMed, error) {
	i, err := indiceDe(tpGrMedCodigos[:], s, "TpGrMed")
	return TpGrMed(i), err
}

// ===========================================================================
// UMedFat (TuMedFat)
// ===========================================================================

// UMedFat e a unidade de medida de faturamento (tags uMed do prod, gMedida
// e gCons).
type UMedFat int

const (
	UmM3      UMedFat = iota // 1
	UmLitros                 // 2
	UmUnidade                // 5
	UmTon                    // 6
)

var uMedFatCodigos = [...]string{"1", "2", "5", "6"}
var uMedFatDescricoes = [...]string{"M3", "Litros", "Unidade", "Tonelada"}

// String retorna o codigo do leiaute -- atencao: a tabela salta de "2"
// para "5" (nao ha codigos 3 e 4).
func (v UMedFat) String() string { return codigoEm(uMedFatCodigos[:], int(v)) }

// Descricao retorna a descricao legivel. Porte de uMedFatToDesc.
func (v UMedFat) Descricao() string { return codigoEm(uMedFatDescricoes[:], int(v)) }

// ParseUMedFat converte o codigo no enum.
func ParseUMedFat(s string) (UMedFat, error) {
	i, err := indiceDe(uMedFatCodigos[:], s, "UMedFat")
	return UMedFat(i), err
}

// ===========================================================================
// TpCategoria (TtpCategoria)
// ===========================================================================

// TpCategoria e a categoria de consumo (tag tpCategoria).
type TpCategoria int

const (
	TcComercial       TpCategoria = iota // 01
	TcConsumoProprio                     // 02
	TcIndustrial                         // 04
	TcResidencial                        // 06
	TcRural                              // 07
	TcServicoPublico                     // 08
	TcSocial                             // 09
	TcMista                              // 10
	TcResidComercial                     // 11
	TcResidIndustrial                    // 12
	TcOutros                             // 99
)

var tpCategoriaCodigos = [...]string{
	"01", "02", "04", "06", "07", "08", "09", "10", "11", "12", "99",
}

// String retorna o codigo do leiaute -- a tabela tem SALTOS (nao ha 03 nem
// 05), copiados de tpCategoriaArrayStrings.
func (v TpCategoria) String() string { return codigoEm(tpCategoriaCodigos[:], int(v)) }

// ParseTpCategoria converte o codigo no enum.
func ParseTpCategoria(s string) (TpCategoria, error) {
	i, err := indiceDe(tpCategoriaCodigos[:], s, "TpCategoria")
	return TpCategoria(i), err
}

// ===========================================================================
// TpProc (TtpProc)
// ===========================================================================

// TpProc e o tipo do processo referenciado (tag tpProc). ATENCAO: na NFAg
// os codigos comecam em "0" (processo administrativo ESTADUAL), diferente
// da NFGas.
type TpProc int

const (
	TpProcAdmEstadual TpProc = iota // 0
	TpJusticaFederal                // 1
	TpJusticaEstadual               // 2
	TpProcAdmMunicial               // 3
	TpProcAdmFederal                // 4
	TpProcon                        // 5
)

var tpProcCodigos = [...]string{"0", "1", "2", "3", "4", "5"}

// String retorna o codigo do leiaute.
func (v TpProc) String() string { return codigoEm(tpProcCodigos[:], int(v)) }

// ParseTpProc converte o codigo no enum.
func ParseTpProc(s string) (TpProc, error) {
	i, err := indiceDe(tpProcCodigos[:], s, "TpProc")
	return TpProc(i), err
}

// ===========================================================================
// TipoEvento (TACBrTipoEvento, subconjunto da NFAg)
// ===========================================================================

// TipoEvento e o tipo do evento (tag tpEvento). A NFAg mapeia QUATRO
// eventos em StrToTpEventoNFAg -- cancelamento, autorizacao de
// substituicao, autorizacao de ajuste e liberacao de prazo de cancelamento
// -- mas o GERADOR do ACBr (e o leiaute de envio) so implementa o
// cancelamento.
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
	"Nao Mapeado", "Cancelamento", "Autorizado Substituicao",
	"Autorizado Ajuste", "Liberacao de Prazo de Cancelamento",
}

// String retorna o codigo do leiaute.
func (v TipoEvento) String() string { return codigoEm(tipoEventoCodigos[:], int(v)) }

// Descricao retorna a descricao curta do evento.
func (v TipoEvento) Descricao() string { return codigoEm(tipoEventoDescricoes[:], int(v)) }

// ParseTipoEvento converte o codigo no enum. Porte de StrToTpEventoNFAg.
func ParseTipoEvento(s string) (TipoEvento, error) {
	i, err := indiceDe(tipoEventoCodigos[:], s, "TipoEvento")
	return TipoEvento(i), err
}
