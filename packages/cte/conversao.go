// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package cte

import (
	"errors"
	"fmt"
	"strings"
)

// Enums do leiaute do CT-e. Porte de PCNCTe/pcteConversaoCTe.pas -- ver
// doc.go sobre por que esta unit prevalece sobre Base/ACBrCTe.Conversao.pas.
//
// Cada enum expoe String() com o codigo EXATO do leiaute e ParseXxx com
// erro. A ordem dos membros Go reproduz a ordem de declaracao do Delphi,
// porque e ela que define o ordinal -- e ordinal errado e o defeito que nao
// aparece em build, vet nem teste ingenuo.
//
// Enums compartilhados por todos os DFe (TipoAmbiente, TipoEmissao,
// Indicador, OrigemMercadoria) vivem em packages/pcn e NAO sao
// redeclarados aqui.

// ErrEnumInvalido indica valor de string que nao corresponde a nenhum
// membro do enum.
var ErrEnumInvalido = errors.New("cte: valor invalido para enum")

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %s nao aceita %q", ErrEnumInvalido, enum, valor)
}

// codigoEm devolve o codigo da posicao i, ou vazio se fora de faixa.
// Membro sentinela -1 cai no vazio, como o EnumeradoToStr do Delphi faz com
// TXxx(-1).
func codigoEm(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}

// indiceDe devolve a posicao do codigo na tabela, comparando SEM diferenciar
// maiusculas de minusculas -- e o AnsiSameText de StrToEnumerado
// (ACBrDFe.Conversao.pas). So muda o resultado nos enums cujo codigo nao e
// numerico: TipoDirecao ("N"/"L"/"S"/"O") e TipoPropriedade ("P"/"T").
func indiceDe(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if strings.EqualFold(c, s) {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// indiceDeExato e o mesmo, COM diferenciacao de caixa.
//
// Dois enums do CT-e nao passam pelo StrToEnumerado: ClassDuto e
// TpContratacao tem laco proprio com `=` (pcteConversaoCTe.pas:1109 e
// :1129). Como os codigos deles sao numericos, a diferenca nao muda
// resultado hoje -- mas replicar o mecanismo evita que uma mudanca futura
// de codigo divirja em silencio.
func indiceDeExato(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if c == s {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// ---------------------------------------------------------------------------
// Modelo do documento
// ---------------------------------------------------------------------------

// ModeloCTe e o modelo do documento. Porte de TModeloCTe.
//
// ATENCAO: MoCTe e MoCTeSimp compartilham o codigo "57" -- ver doc.go.
type ModeloCTe int

const (
	MoCTe     ModeloCTe = iota // 57 -- CT-e
	MoGTVe                     // 64 -- GTV-e
	MoCTeOS                    // 67 -- CT-e OS
	MoCTeSimp                  // 57 -- CT-e Simplificado
	// _fimModeloCTe conta os membros declarados; nao e valor do leiaute.
	_fimModeloCTe
)

var modeloCTeCodigos = [...]string{"57", "64", "67", "57"}

// String retorna o codigo do modelo.
func (v ModeloCTe) String() string { return codigoEm(modeloCTeCodigos[:], int(v)) }

// ParseModeloCTe converte o codigo no enum.
//
// A tabela de leitura tem TRES entradas, nao quatro: o "57" sempre devolve
// MoCTe, nunca MoCTeSimp (StrToModeloCTe, :~). A ida e volta pelo codigo do
// modelo NAO fecha, e nao ha como fazer fechar -- a informacao que separa os
// dois nao esta no codigo. Quem precisa distinguir olha o tpCTe, onde o
// Simplificado tem valor proprio (TcCTeSimp).
func ParseModeloCTe(s string) (ModeloCTe, error) {
	switch s {
	case "57":
		return MoCTe, nil
	case "64":
		return MoGTVe, nil
	case "67":
		return MoCTeOS, nil
	}
	return MoCTe, erroEnum("ModeloCTe", s)
}

// ---------------------------------------------------------------------------
// Identificacao da prestacao
// ---------------------------------------------------------------------------

// TipoCTe e o tipo do conhecimento. Porte de TpcteTipoCTe.
type TipoCTe int

const (
	TcNormal        TipoCTe = iota // 0
	TcComplemento                  // 1
	TcAnulacao                     // 2
	TcSubstituto                   // 3
	TcGTVe                         // 4
	TcCTeSimp                      // 5
	TcSubstCTeSimpl                // 6
	// _fimTipoCTe conta os membros declarados; nao e valor do leiaute.
	_fimTipoCTe
)

var tipoCTeCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6"}

// tipoCTeDescricoes e o texto do DACTE, nao o do XML.
//
// O ACBr mantem as duas tabelas (tpCTToStr e tpCTToStrText) e elas DIVERGEM
// das do topo da unit: a ArrayStrings diz "CTeSimp" e a funcao diz
// "CTe Simplificado". Vale a funcao, que e quem o codigo chama.
var tipoCTeDescricoes = [...]string{
	"NORMAL", "COMPLEMENTO", "ANULAÇÃO", "SUBSTITUTO",
	"GTVe", "CTe Simplificado", "CTe Simplificado Substituto",
}

func (v TipoCTe) String() string { return codigoEm(tipoCTeCodigos[:], int(v)) }

// Descricao retorna o texto do DACTE.
func (v TipoCTe) Descricao() string { return codigoEm(tipoCTeDescricoes[:], int(v)) }

func ParseTipoCTe(s string) (TipoCTe, error) {
	i, err := indiceDe(tipoCTeCodigos[:], s, "TipoCTe")
	return TipoCTe(i), err
}

// TipoServico e o tipo da prestacao. Porte de TpcteTipoServico.
//
// A tabela de codigos PULA o "5": o leiaute vai de "4" (vinculado a
// multimodal) para "6" (transporte de pessoas). Preencher a tabela com um
// "5" inexistente deslocaria todos os membros seguintes -- e nada acusaria,
// porque os ordinais continuariam contiguos.
type TipoServico int

const (
	TsNormal         TipoServico = iota // 0
	TsSubcontratacao                    // 1
	TsRedespacho                        // 2
	TsIntermediario                     // 3
	TsMultimodal                        // 4
	TsTranspPessoas                     // 6  <- pula o 5
	TsTranspValores                     // 7
	TsExcessoBagagem                    // 8
	TsGTV                               // 9
	// _fimTipoServico conta os membros declarados; nao e valor do leiaute.
	_fimTipoServico
)

var tipoServicoCodigos = [...]string{"0", "1", "2", "3", "4", "6", "7", "8", "9"}

var tipoServicoDescricoes = [...]string{
	"NORMAL", "SUBCONTRATAÇÃO", "REDESPACHO", "REDESP. INTERMEDIÁRIO",
	"VINC. A MULTIMODAL", "TRANSP. PESSOAS", "TRANSP. VALORES",
	"EXCESSO BAGAGEM", "GTV",
}

func (v TipoServico) String() string { return codigoEm(tipoServicoCodigos[:], int(v)) }

func (v TipoServico) Descricao() string { return codigoEm(tipoServicoDescricoes[:], int(v)) }

func ParseTipoServico(s string) (TipoServico, error) {
	i, err := indiceDe(tipoServicoCodigos[:], s, "TipoServico")
	return TipoServico(i), err
}

// FormaPagamento e a forma de pagamento do frete. Porte de
// TpcteFormaPagamento.
type FormaPagamento int

const (
	FpPago   FormaPagamento = iota // 0
	FpAPagar                       // 1
	FpOutros                       // 2
	// _fimFormaPagamento conta os membros declarados; nao e valor do leiaute.
	_fimFormaPagamento
)

var formaPagamentoCodigos = [...]string{"0", "1", "2"}

var formaPagamentoDescricoes = [...]string{"PAGO", "A PAGAR", "OUTROS"}

func (v FormaPagamento) String() string { return codigoEm(formaPagamentoCodigos[:], int(v)) }

func (v FormaPagamento) Descricao() string {
	return codigoEm(formaPagamentoDescricoes[:], int(v))
}

func ParseFormaPagamento(s string) (FormaPagamento, error) {
	i, err := indiceDe(formaPagamentoCodigos[:], s, "FormaPagamento")
	return FormaPagamento(i), err
}

// Tomador e quem toma o servico. Porte de TpcteTomador.
type Tomador int

const (
	TmRemetente    Tomador = iota // 0
	TmExpedidor                   // 1
	TmRecebedor                   // 2
	TmDestinatario                // 3
	TmOutros                      // 4
	// _fimTomador conta os membros declarados; nao e valor do leiaute.
	_fimTomador
)

var tomadorCodigos = [...]string{"0", "1", "2", "3", "4"}

var tomadorDescricoes = [...]string{
	"REMETENTE", "EXPEDIDOR", "RECEBEDOR", "DESTINATARIO", "OUTROS",
}

func (v Tomador) String() string { return codigoEm(tomadorCodigos[:], int(v)) }

func (v Tomador) Descricao() string { return codigoEm(tomadorDescricoes[:], int(v)) }

func ParseTomador(s string) (Tomador, error) {
	i, err := indiceDe(tomadorCodigos[:], s, "Tomador")
	return Tomador(i), err
}

// Retira indica se a mercadoria sera retirada no local. Porte de
// TpcteRetira.
//
// ATENCAO ao ordinal: RtSim e ZERO e RtNao e UM. E o inverso do que a
// intuicao sugere, e e assim no leiaute -- o campo `retira` responde "o
// recebedor retira?" com 0=Sim e 1=Nao.
type Retira int

const (
	RtSim Retira = iota // 0
	RtNao               // 1
	// _fimRetira conta os membros declarados; nao e valor do leiaute.
	_fimRetira
)

var retiraCodigos = [...]string{"0", "1"}

func (v Retira) String() string { return codigoEm(retiraCodigos[:], int(v)) }

func ParseRetira(s string) (Retira, error) {
	i, err := indiceDe(retiraCodigos[:], s, "Retira")
	return Retira(i), err
}

// Lotacao indica se a prestacao e por lotacao. Porte de TpcteLotacao.
//
// Aqui o ordinal segue a intuicao -- LtNao e zero --, ao contrario de
// Retira. Os dois convivem no mesmo leiaute.
type Lotacao int

const (
	LtNao Lotacao = iota // 0
	LtSim                // 1
	// _fimLotacao conta os membros declarados; nao e valor do leiaute.
	_fimLotacao
)

var lotacaoCodigos = [...]string{"0", "1"}

func (v Lotacao) String() string { return codigoEm(lotacaoCodigos[:], int(v)) }

func ParseLotacao(s string) (Lotacao, error) {
	i, err := indiceDe(lotacaoCodigos[:], s, "Lotacao")
	return Lotacao(i), err
}

// ---------------------------------------------------------------------------
// Documentos
// ---------------------------------------------------------------------------

// ModeloNF e o modelo da nota fiscal referenciada. Porte de TpcteModeloNF.
type ModeloNF int

const (
	MoNF011AAvulsa ModeloNF = iota // 01
	MoNFProdutor                   // 04
	// _fimModeloNF conta os membros declarados; nao e valor do leiaute.
	_fimModeloNF
)

var modeloNFCodigos = [...]string{"01", "04"}

func (v ModeloNF) String() string { return codigoEm(modeloNFCodigos[:], int(v)) }

func ParseModeloNF(s string) (ModeloNF, error) {
	i, err := indiceDe(modeloNFCodigos[:], s, "ModeloNF")
	return ModeloNF(i), err
}

// TipoDocumento e o tipo do documento originario. Porte de
// TpcteTipoDocumento.
//
// Os codigos NAO sao sequenciais: "00", "10", "59", "65", "99" sao os
// modelos de documento fiscal correspondentes (declaracao, dutoviario,
// CF-e SAT, NFC-e e outros).
type TipoDocumento int

const (
	TdDeclaracao TipoDocumento = iota // 00
	TdDutoviario                      // 10
	TdCFeSAT                          // 59
	TdNFCe                            // 65
	TdDocOutros                       // 99
	// _fimTipoDocumento conta os membros declarados; nao e valor do leiaute.
	_fimTipoDocumento
)

var tipoDocumentoCodigos = [...]string{"00", "10", "59", "65", "99"}

func (v TipoDocumento) String() string { return codigoEm(tipoDocumentoCodigos[:], int(v)) }

func ParseTipoDocumento(s string) (TipoDocumento, error) {
	i, err := indiceDe(tipoDocumentoCodigos[:], s, "TipoDocumento")
	return TipoDocumento(i), err
}

// TipoDocumentoAnterior e o tipo do documento de transporte anterior.
// Porte de TpcteTipoDocumentoAnterior.
//
// Os catorze primeiros sao sequenciais ("00" a "13") e o ultimo salta para
// "99" -- o "outros" do leiaute.
type TipoDocumentoAnterior int

const (
	DaCTRC           TipoDocumentoAnterior = iota // 00
	DaCTAC                                        // 01
	DaACT                                         // 02
	DaNF7                                         // 03
	DaNF27                                        // 04
	DaCAN                                         // 05
	DaCTMC                                        // 06
	DaATRE                                        // 07
	DaDTA                                         // 08
	DaCAI                                         // 09
	DaCCPI                                        // 10
	DaCA                                          // 11
	DaTIF                                         // 12
	DaBL                                          // 13
	DaAnteriorOutros                              // 99
	// _fimTipoDocumentoAnterior conta os membros; nao e valor do leiaute.
	_fimTipoDocumentoAnterior
)

var tipoDocumentoAnteriorCodigos = [...]string{
	"00", "01", "02", "03", "04", "05", "06", "07",
	"08", "09", "10", "11", "12", "13", "99",
}

func (v TipoDocumentoAnterior) String() string {
	return codigoEm(tipoDocumentoAnteriorCodigos[:], int(v))
}

func ParseTipoDocumentoAnterior(s string) (TipoDocumentoAnterior, error) {
	i, err := indiceDe(tipoDocumentoAnteriorCodigos[:], s, "TipoDocumentoAnterior")
	return TipoDocumentoAnterior(i), err
}

// IndNegociavel indica se o documento e negociavel. Porte de
// TpcnindNegociavel.
type IndNegociavel int

const (
	InNaoNegociavel IndNegociavel = iota // 0
	InNegociavel                         // 1
	// _fimIndNegociavel conta os membros declarados; nao e valor do leiaute.
	_fimIndNegociavel
)

var indNegociavelCodigos = [...]string{"0", "1"}

func (v IndNegociavel) String() string { return codigoEm(indNegociavelCodigos[:], int(v)) }

func ParseIndNegociavel(s string) (IndNegociavel, error) {
	i, err := indiceDe(indNegociavelCodigos[:], s, "IndNegociavel")
	return IndNegociavel(i), err
}

// ---------------------------------------------------------------------------
// Data, hora e periodo
// ---------------------------------------------------------------------------

// TipoDataPeriodo e a forma de informar a data de entrega. Porte de
// TpcteTipoDataPeriodo.
//
// O ultimo membro tem codigo "-1", e nao vazio: o leiaute usa o literal
// negativo para "nao informado". Tratar como vazio faria a tag sair sem
// conteudo em vez de sair com -1.
type TipoDataPeriodo int

const (
	TdSemData      TipoDataPeriodo = iota // 0
	TdNaData                              // 1
	TdAteData                             // 2
	TdApartirData                         // 3
	TdNoPeriodo                           // 4
	TdNaoInformado                        // -1
	// _fimTipoDataPeriodo conta os membros; nao e valor do leiaute.
	_fimTipoDataPeriodo
)

var tipoDataPeriodoCodigos = [...]string{"0", "1", "2", "3", "4", "-1"}

func (v TipoDataPeriodo) String() string { return codigoEm(tipoDataPeriodoCodigos[:], int(v)) }

func ParseTipoDataPeriodo(s string) (TipoDataPeriodo, error) {
	i, err := indiceDe(tipoDataPeriodoCodigos[:], s, "TipoDataPeriodo")
	return TipoDataPeriodo(i), err
}

// TipoHorarioIntervalo e a forma de informar a hora de entrega. Porte de
// TpcteTipoHorarioIntervalo. Mesma estrutura do TipoDataPeriodo, inclusive
// o "-1" no ultimo membro.
type TipoHorarioIntervalo int

const (
	ThSemHorario       TipoHorarioIntervalo = iota // 0
	ThNoHorario                                    // 1
	ThAteHorario                                   // 2
	ThApartirHorario                               // 3
	ThNoIntervalo                                  // 4
	ThHoraNaoInformado                             // -1
	// _fimTipoHorarioIntervalo conta os membros; nao e valor do leiaute.
	_fimTipoHorarioIntervalo
)

var tipoHorarioIntervaloCodigos = [...]string{"0", "1", "2", "3", "4", "-1"}

func (v TipoHorarioIntervalo) String() string {
	return codigoEm(tipoHorarioIntervaloCodigos[:], int(v))
}

func ParseTipoHorarioIntervalo(s string) (TipoHorarioIntervalo, error) {
	i, err := indiceDe(tipoHorarioIntervaloCodigos[:], s, "TipoHorarioIntervalo")
	return TipoHorarioIntervalo(i), err
}

// ---------------------------------------------------------------------------
// Modal rodoviario e pedagio
// ---------------------------------------------------------------------------

// RspPagPedagio e o responsavel pelo pagamento do pedagio. Porte de
// TpcteRspPagPedagio.
type RspPagPedagio int

const (
	RpEmitente       RspPagPedagio = iota // 0
	RpRemetente                           // 1
	RpExpedidor                           // 2
	RpRecebedor                           // 3
	RpDestinatario                        // 4
	RpTomadorServico                      // 5
	// _fimRspPagPedagio conta os membros; nao e valor do leiaute.
	_fimRspPagPedagio
)

var rspPagPedagioCodigos = [...]string{"0", "1", "2", "3", "4", "5"}

func (v RspPagPedagio) String() string { return codigoEm(rspPagPedagioCodigos[:], int(v)) }

func ParseRspPagPedagio(s string) (RspPagPedagio, error) {
	i, err := indiceDe(rspPagPedagioCodigos[:], s, "RspPagPedagio")
	return RspPagPedagio(i), err
}

// TipoDispositivo e o meio de pagamento do vale-pedagio. Porte de
// TpcteTipoDispositivo. Comeca em "1", nao em "0".
type TipoDispositivo int

const (
	TdCartaoMagnetico TipoDispositivo = iota // 1
	TdTAG                                    // 2
	TdTicket                                 // 3
	// _fimTipoDispositivo conta os membros; nao e valor do leiaute.
	_fimTipoDispositivo
)

var tipoDispositivoCodigos = [...]string{"1", "2", "3"}

func (v TipoDispositivo) String() string { return codigoEm(tipoDispositivoCodigos[:], int(v)) }

func ParseTipoDispositivo(s string) (TipoDispositivo, error) {
	i, err := indiceDe(tipoDispositivoCodigos[:], s, "TipoDispositivo")
	return TipoDispositivo(i), err
}

// TipoPropriedade e a propriedade do veiculo. Porte de
// TpcteTipoPropriedade.
//
// Codigos NAO numericos: "P" (proprio) e "T" (terceiro). E um dos dois
// enums do CT-e em que a comparacao sem caixa do StrToEnumerado muda
// resultado -- "p" tambem e aceito.
type TipoPropriedade int

const (
	TpProprio  TipoPropriedade = iota // P
	TpTerceiro                        // T
	// _fimTipoPropriedade conta os membros; nao e valor do leiaute.
	_fimTipoPropriedade
)

var tipoPropriedadeCodigos = [...]string{"P", "T"}

func (v TipoPropriedade) String() string { return codigoEm(tipoPropriedadeCodigos[:], int(v)) }

func ParseTipoPropriedade(s string) (TipoPropriedade, error) {
	i, err := indiceDe(tipoPropriedadeCodigos[:], s, "TipoPropriedade")
	return TipoPropriedade(i), err
}

// TipoVeiculo distingue tracao de reboque. Porte de TpcteTipoVeiculo.
type TipoVeiculo int

const (
	TvTracao  TipoVeiculo = iota // 0
	TvReboque                    // 1
	// _fimTipoVeiculo conta os membros; nao e valor do leiaute.
	_fimTipoVeiculo
)

var tipoVeiculoCodigos = [...]string{"0", "1"}

func (v TipoVeiculo) String() string { return codigoEm(tipoVeiculoCodigos[:], int(v)) }

func ParseTipoVeiculo(s string) (TipoVeiculo, error) {
	i, err := indiceDe(tipoVeiculoCodigos[:], s, "TipoVeiculo")
	return TipoVeiculo(i), err
}

// TipoFretamento e a modalidade de fretamento do transporte de pessoas.
// Porte de TtpFretamento.
//
// O primeiro membro tem codigo VAZIO -- e o "nenhum" que faz a tag nao
// sair. Os enums do CT-e que tem essa forma sao quatro: este, CRT,
// ClassDuto e TipoContratacao.
//
// Note o nome do terceiro membro no Delphi: `tpContinuo`, com prefixo `tp`
// em vez de `tf` como os irmaos. E erro de digitacao do original, sem
// efeito -- o ordinal e o codigo estao certos.
type TipoFretamento int

const (
	TfNenhum   TipoFretamento = iota // vazio
	TfEventual                       // 1
	TfContinuo                       // 2
	// _fimTipoFretamento conta os membros; nao e valor do leiaute.
	_fimTipoFretamento
)

var tipoFretamentoCodigos = [...]string{"", "1", "2"}

func (v TipoFretamento) String() string { return codigoEm(tipoFretamentoCodigos[:], int(v)) }

func ParseTipoFretamento(s string) (TipoFretamento, error) {
	i, err := indiceDe(tipoFretamentoCodigos[:], s, "TipoFretamento")
	return TipoFretamento(i), err
}

// ---------------------------------------------------------------------------
// Modal ferroviario
// ---------------------------------------------------------------------------

// TipoTrafego e o tipo de trafego ferroviario. Porte de TpcteTipoTrafego.
type TipoTrafego int

const (
	TtProprio           TipoTrafego = iota // 0
	TtMutuo                                // 1
	TtRodoferroviario                      // 2
	TtRodoviarioTrafego                    // 3
	// _fimTipoTrafego conta os membros; nao e valor do leiaute.
	_fimTipoTrafego
)

var tipoTrafegoCodigos = [...]string{"0", "1", "2", "3"}

func (v TipoTrafego) String() string { return codigoEm(tipoTrafegoCodigos[:], int(v)) }

func ParseTipoTrafego(s string) (TipoTrafego, error) {
	i, err := indiceDe(tipoTrafegoCodigos[:], s, "TipoTrafego")
	return TipoTrafego(i), err
}

// TrafegoMutuo indica a ponta do trafego mutuo. Porte de
// TpcteTrafegoMutuo. Comeca em "1".
type TrafegoMutuo int

const (
	TmOrigem  TrafegoMutuo = iota // 1
	TmDestino                     // 2
	// _fimTrafegoMutuo conta os membros; nao e valor do leiaute.
	_fimTrafegoMutuo
)

var trafegoMutuoCodigos = [...]string{"1", "2"}

func (v TrafegoMutuo) String() string { return codigoEm(trafegoMutuoCodigos[:], int(v)) }

func ParseTrafegoMutuo(s string) (TrafegoMutuo, error) {
	i, err := indiceDe(trafegoMutuoCodigos[:], s, "TrafegoMutuo")
	return TrafegoMutuo(i), err
}

// ---------------------------------------------------------------------------
// Modal aquaviario
// ---------------------------------------------------------------------------

// TipoDirecao e o sentido da navegacao. Porte de TpcteDirecao.
//
// Codigos NAO numericos: pontos cardeais em letra. E o segundo dos dois
// enums em que a comparacao sem caixa muda resultado.
type TipoDirecao int

const (
	DrNorte TipoDirecao = iota // N
	DrLeste                    // L
	DrSul                      // S
	DrOeste                    // O
	// _fimTipoDirecao conta os membros; nao e valor do leiaute.
	_fimTipoDirecao
)

var tipoDirecaoCodigos = [...]string{"N", "L", "S", "O"}

func (v TipoDirecao) String() string { return codigoEm(tipoDirecaoCodigos[:], int(v)) }

func ParseTipoDirecao(s string) (TipoDirecao, error) {
	i, err := indiceDe(tipoDirecaoCodigos[:], s, "TipoDirecao")
	return TipoDirecao(i), err
}

// ---------------------------------------------------------------------------
// Modal dutoviario
// ---------------------------------------------------------------------------

// ClassDuto e a classificacao do duto. Porte de TclassDuto.
//
// Um dos dois enums que o ACBr le com laco proprio e comparacao COM caixa
// (:1109), em vez do StrToEnumerado -- ver indiceDeExato.
type ClassDuto int

const (
	TcdNenhum     ClassDuto = iota // vazio
	TcdGasoduto                    // 1
	TcdMineroduto                  // 2
	TcdOleoduto                    // 3
	// _fimClassDuto conta os membros; nao e valor do leiaute.
	_fimClassDuto
)

var classDutoCodigos = [...]string{"", "1", "2", "3"}

func (v ClassDuto) String() string { return codigoEm(classDutoCodigos[:], int(v)) }

func ParseClassDuto(s string) (ClassDuto, error) {
	i, err := indiceDeExato(classDutoCodigos[:], s, "ClassDuto")
	return ClassDuto(i), err
}

// TipoContratacao e a forma de contratacao do transporte dutoviario.
// Porte de TtpContratacao. Tambem lido com comparacao COM caixa (:1129).
type TipoContratacao int

const (
	TcNenhum            TipoContratacao = iota // vazio
	TcPontaPonto                               // 1
	TcCapacidadeEntrada                        // 2
	TcCapacidadeSaida                          // 3
	// _fimTipoContratacao conta os membros; nao e valor do leiaute.
	_fimTipoContratacao
)

var tipoContratacaoCodigos = [...]string{"", "1", "2", "3"}

func (v TipoContratacao) String() string { return codigoEm(tipoContratacaoCodigos[:], int(v)) }

func ParseTipoContratacao(s string) (TipoContratacao, error) {
	i, err := indiceDeExato(tipoContratacaoCodigos[:], s, "TipoContratacao")
	return TipoContratacao(i), err
}

// ---------------------------------------------------------------------------
// Modal aereo
// ---------------------------------------------------------------------------

// TpInfManu e o codigo de informacao de manuseio da carga aerea.
// Porte de TpInfManu.
//
// DUAS TABELAS, e elas NAO sao equivalentes:
//
//	String()    "01".."15", "99"  -- 16 membros, o leiaute
//	StringV2()  "1".."9",   "99"  -- 10 membros, a forma legada
//
// Os seis membros PI965..PI970 (as Packing Instructions da IATA) **nao tem
// codigo na tabela V2**: ela simplesmente nao os alcanca.
//
// O ACBr usa as duas em lados opostos, e isso QUEBRA a ida e volta dele:
//
//	escrita   XmlWriter.pas:3755 e IniWriter.pas:1202 -> TpInfManuToStrV2
//	leitura   XmlReader.pas:203, XmlHandler.pas:3420,
//	          IniReader.pas:1551                      -> StrToTpInfManu
//
// Ou seja: ele grava "1" e procura "01". Um documento gerado pelo ACBr e
// relido pelo ACBr perde o campo, e seis membros nao tem como ser emitidos.
//
// As duas formas ficam expostas aqui porque as duas existem no mundo -- ha
// documento em circulacao com cada uma. QUAL delas o gerador deste porte
// emite e decisao da etapa do gerador, nao desta; o que esta etapa garante
// e que nenhuma das duas se perca.
type TpInfManu int

const (
	ImCEEAV  TpInfManu = iota // 01 / 1
	ImAPCDEA                  // 02 / 2
	ImSAC                     // 03 / 3
	ImAPDENR                  // 04 / 4
	ImAPQI                    // 05 / 5
	ImGSR                     // 06 / 6
	ImNR                      // 07 / 7
	ImAPCC                    // 08 / 8
	ImAAGA                    // 09 / 9
	ImPI965                   // 10 -- sem codigo na tabela V2
	ImPI966                   // 11 -- sem codigo na tabela V2
	ImPI967                   // 12 -- sem codigo na tabela V2
	ImPI968                   // 13 -- sem codigo na tabela V2
	ImPI969                   // 14 -- sem codigo na tabela V2
	ImPI970                   // 15 -- sem codigo na tabela V2
	ImOUTRO                   // 99 / 99
	// _fimTpInfManu conta os membros; nao e valor do leiaute.
	_fimTpInfManu
)

var tpInfManuCodigos = [...]string{
	"01", "02", "03", "04", "05", "06", "07", "08",
	"09", "10", "11", "12", "13", "14", "15", "99",
}

// tpInfManuCodigosV2 cobre SO os dez primeiros membros mais o OUTRO. As
// posicoes de ImPI965 a ImPI970 ficam vazias de proposito -- e a forma de
// dizer "este membro nao existe nesta tabela" sem deslocar os ordinais.
var tpInfManuCodigosV2 = [...]string{
	"1", "2", "3", "4", "5", "6", "7", "8",
	"9", "", "", "", "", "", "", "99",
}

// String retorna o codigo do LEIAUTE, com dois digitos.
func (v TpInfManu) String() string { return codigoEm(tpInfManuCodigos[:], int(v)) }

// StringV2 retorna o codigo da forma LEGADA, com um digito.
//
// Devolve vazio para os seis membros que a tabela legada nao alcanca --
// ImPI965 a ImPI970.
func (v TpInfManu) StringV2() string { return codigoEm(tpInfManuCodigosV2[:], int(v)) }

// ParseTpInfManu le o codigo do LEIAUTE ("01".."15", "99").
func ParseTpInfManu(s string) (TpInfManu, error) {
	i, err := indiceDe(tpInfManuCodigos[:], s, "TpInfManu")
	return TpInfManu(i), err
}

// ParseTpInfManuV2 le o codigo da forma LEGADA ("1".."9", "99").
//
// A busca pula as posicoes vazias: sem isso, uma string vazia casaria com
// ImPI965 e um campo ausente viraria "Packing Instruction 965".
func ParseTpInfManuV2(s string) (TpInfManu, error) {
	if s == "" {
		return ImCEEAV, erroEnum("TpInfManu(V2)", s)
	}
	i, err := indiceDe(tpInfManuCodigosV2[:], s, "TpInfManu(V2)")
	return TpInfManu(i), err
}

// UniMed e a unidade de medida da carga. Porte de TpUniMed.
type UniMed int

const (
	UmKG       UniMed = iota // 1
	UmKGG                    // 2
	UmLitros                 // 3
	UmTI                     // 4
	UmUnidades               // 5
	// _fimUniMed conta os membros; nao e valor do leiaute.
	_fimUniMed
)

var uniMedCodigos = [...]string{"1", "2", "3", "4", "5"}

func (v UniMed) String() string { return codigoEm(uniMedCodigos[:], int(v)) }

func ParseUniMed(s string) (UniMed, error) {
	i, err := indiceDe(uniMedCodigos[:], s, "UniMed")
	return UniMed(i), err
}

// ---------------------------------------------------------------------------
// Seguro, valores e componentes
// ---------------------------------------------------------------------------

// RspSeg e o responsavel pelo seguro da carga. Porte de TpcteRspSeg.
type RspSeg int

const (
	RsRemetente      RspSeg = iota // 0
	RsExpedidor                    // 1
	RsRecebedor                    // 2
	RsDestinatario                 // 3
	RsEmitenteCTe                  // 4
	RsTomadorServico               // 5
	// _fimRspSeg conta os membros; nao e valor do leiaute.
	_fimRspSeg
)

var rspSegCodigos = [...]string{"0", "1", "2", "3", "4", "5"}

var rspSegDescricoes = [...]string{
	"REMETENTE", "EXPEDIDOR", "RECEBEDOR", "DESTINATARIO",
	"EMITENTE", "TOMADOR SERVICO",
}

func (v RspSeg) String() string { return codigoEm(rspSegCodigos[:], int(v)) }

func (v RspSeg) Descricao() string { return codigoEm(rspSegDescricoes[:], int(v)) }

func ParseRspSeg(s string) (RspSeg, error) {
	i, err := indiceDe(rspSegCodigos[:], s, "RspSeg")
	return RspSeg(i), err
}

// Especie e a especie do valor transportado (GTV-e). Porte de TEspecie.
type Especie int

const (
	TeNumerario Especie = iota // 1
	TeCheque                   // 2
	TeMoeda                    // 3
	TeEspOutros                // 4
	// _fimEspecie conta os membros; nao e valor do leiaute.
	_fimEspecie
)

var especieCodigos = [...]string{"1", "2", "3", "4"}

func (v Especie) String() string { return codigoEm(especieCodigos[:], int(v)) }

func ParseEspecie(s string) (Especie, error) {
	i, err := indiceDe(especieCodigos[:], s, "Especie")
	return Especie(i), err
}

// TpNumerario distingue moeda nacional de estrangeira. Porte de
// TtpNumerario.
type TpNumerario int

const (
	TnNacional    TpNumerario = iota // 1
	TnEstrangeiro                    // 2
	// _fimTpNumerario conta os membros; nao e valor do leiaute.
	_fimTpNumerario
)

var tpNumerarioCodigos = [...]string{"1", "2"}

func (v TpNumerario) String() string { return codigoEm(tpNumerarioCodigos[:], int(v)) }

func ParseTpNumerario(s string) (TpNumerario, error) {
	i, err := indiceDe(tpNumerarioCodigos[:], s, "TpNumerario")
	return TpNumerario(i), err
}

// TpComp e o tipo do componente do valor da prestacao. Porte de TtpComp.
type TpComp int

const (
	TcCustodia      TpComp = iota // 1
	TcEmbarque                    // 2
	TcTempoDeEspera               // 3
	TcMalote                      // 4
	TcAdValorem                   // 5
	TcCompOutros                  // 6
	// _fimTpComp conta os membros; nao e valor do leiaute.
	_fimTpComp
)

var tpCompCodigos = [...]string{"1", "2", "3", "4", "5", "6"}

func (v TpComp) String() string { return codigoEm(tpCompCodigos[:], int(v)) }

func ParseTpComp(s string) (TpComp, error) {
	i, err := indiceDe(tpCompCodigos[:], s, "TpComp")
	return TpComp(i), err
}

// TpPrest indica se a prestacao e total ou parcial. Porte de TtpPrest.
type TpPrest int

const (
	TpTotal   TpPrest = iota // 1
	TpParcial                // 2
	// _fimTpPrest conta os membros; nao e valor do leiaute.
	_fimTpPrest
)

var tpPrestCodigos = [...]string{"1", "2"}

func (v TpPrest) String() string { return codigoEm(tpPrestCodigos[:], int(v)) }

func ParseTpPrest(s string) (TpPrest, error) {
	i, err := indiceDe(tpPrestCodigos[:], s, "TpPrest")
	return TpPrest(i), err
}

// ---------------------------------------------------------------------------
// Regime tributario e eventos
// ---------------------------------------------------------------------------

// CRT e o codigo de regime tributario do emitente. Porte de TCRT.
//
// NAO e o mesmo enum do packages/nfe: la o primeiro membro e
// CrtSimplesNacional com codigo "1"; aqui ha um CrtNenhum de codigo VAZIO
// na frente, e todos os demais ficam deslocados em um. Confundir os dois
// declararia o regime errado na prestacao.
type CRT int

const (
	CrtNenhum                CRT = iota // vazio
	CrtSimplesNacional                  // 1
	CrtSimplesExcessoReceita            // 2
	CrtRegimeNormal                     // 3
	CrtSimplesNacionalMEI               // 4
	// _fimCRT conta os membros; nao e valor do leiaute.
	_fimCRT
)

var crtCodigos = [...]string{"", "1", "2", "3", "4"}

func (v CRT) String() string { return codigoEm(crtCodigos[:], int(v)) }

func ParseCRT(s string) (CRT, error) {
	i, err := indiceDe(crtCodigos[:], s, "CRT")
	return CRT(i), err
}

// TpMotivo e o motivo da prestacao em desacordo. Porte de TtpMotivo.
type TpMotivo int

const (
	TmNaoEncontrado TpMotivo = iota // 1
	TmRecusa                        // 2
	TmInexistente                   // 3
	TmMotivoOutro                   // 4
	// _fimTpMotivo conta os membros; nao e valor do leiaute.
	_fimTpMotivo
)

var tpMotivoCodigos = [...]string{"1", "2", "3", "4"}

// tpMotivoDescricoes sao os textos do evento, COM acento -- e o unico
// conjunto de descricoes do CT-e que os leva.
var tpMotivoDescricoes = [...]string{
	"Não Encontrado", "Recusa", "Inexistente", "Outro",
}

func (v TpMotivo) String() string { return codigoEm(tpMotivoCodigos[:], int(v)) }

func (v TpMotivo) Descricao() string { return codigoEm(tpMotivoDescricoes[:], int(v)) }

func ParseTpMotivo(s string) (TpMotivo, error) {
	i, err := indiceDe(tpMotivoCodigos[:], s, "TpMotivo")
	return TpMotivo(i), err
}
