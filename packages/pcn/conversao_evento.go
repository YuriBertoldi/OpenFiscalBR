// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package pcn

import "fmt"

// ===========================================================================
// TipoEvento (TACBrTipoEvento) -- tag tpEvento dos eventos de DFe
// ===========================================================================

// TipoEvento e o catalogo de eventos de TODOS os documentos fiscais
// eletronicos, nao so da NF-e. Porte de TACBrTipoEvento
// (ACBrDFe.Conversao.pas:93), que o ACBr tambem mantem unico.
//
// Ele vive aqui, e nao em packages/nfe, porque CTe e MDFe compartilham a
// lista -- varios eventos de MDFe chegam numa consulta de NF-e, e o leitor
// precisa saber nomea-los.
//
// A conversao tem um DESEQUILIBRIO que governa o desenho:
//
//	enum -> codigo   total e sem ambiguidade (Codigo)
//	codigo -> enum   AMBIGUO no catalogo inteiro
//
// Sete codigos se repetem entre documentos diferentes -- o 110112 e
// "Encerramento" no MDFe e "Cancelamento por substituicao" na NF-e, e o
// 240140 e "CTe de substituicao" no CTe e "Autorizado substituicao" no
// MDFe. Por isso NAO existe um ParseTipoEvento global: cada documento tem a
// sua tabela, como o StrToTpEventoNFe, o StrToTpEventoCTe e o
// StrToTpEventoMDFe do ACBr.
type TipoEvento int

const (
	TeNaoMapeado                              TipoEvento = iota // -99999 -- NaoMapeado
	TeCCe                                                       // 110110 -- CCe
	TeCancelamento                                              // 110111 -- Cancelamento
	TeManifDestConfirmacao                                      // 210200 -- ManifDestConfirmacao
	TeManifDestCiencia                                          // 210210 -- ManifDestCiencia
	TeManifDestDesconhecimento                                  // 210220 -- ManifDestDesconhecimento
	TeManifDestOperNaoRealizada                                 // 210240 -- ManifDestOperNaoRealizada
	TeEncerramento                                              // 110112 -- Encerramento
	TeEPEC                                                      // 110113 -- EPEC
	TeInclusaoCondutor                                          // 110114 -- InclusaoCondutor
	TeMultiModal                                                // 110160 -- MultiModal
	TeRegistroPassagem                                          // 310620 -- RegistroPassagem
	TeRegistroPassagemBRId                                      // 510620 -- RegistroPassagemBRId
	TeEPECNFe                                                   // 110140 -- EPECNFe
	TeRegistroCTe                                               // 610600 -- RegistroCTe
	TeRegistroPassagemNFeCancelado                              // 610501 -- RegistroPassagemNFeCancelado
	TeRegistroPassagemNFeRFID                                   // 610550 -- RegistroPassagemNFeRFID
	TeCTeCancelado                                              // 610601 -- CTeCancelado
	TeMDFeCancelado                                             // 610611 -- MDFeCancelado
	TeVistoriaSuframa                                           // 990900 -- VistoriaSuframa
	TePedProrrog1                                               // 111500 -- PedProrrog1
	TePedProrrog2                                               // 111501 -- PedProrrog2
	TeCanPedProrrog1                                            // 111502 -- CanPedProrrog1
	TeCanPedProrrog2                                            // 111503 -- CanPedProrrog2
	TeEventoFiscoPP1                                            // 411500 -- EventoFiscoPP1
	TeEventoFiscoPP2                                            // 411501 -- EventoFiscoPP2
	TeEventoFiscoCPP1                                           // 411502 -- EventoFiscoCPP1
	TeEventoFiscoCPP2                                           // 411503 -- EventoFiscoCPP2
	TeRegistroPassagemNFe                                       // 610500 -- RegistroPassagemNFe
	TeConfInternalizacao                                        // 990910 -- ConfInternalizacao
	TeCTeAutorizado                                             // 000000 -- CTeAutorizado
	TeMDFeAutorizado                                            // 610610 -- MDFeAutorizado
	TePrestDesacordo                                            // 610110 -- PrestDesacordo
	TeGTV                                                       // 110170 -- GTV
	TeMDFeAutorizado2                                           // 310610 -- MDFeAutorizado2
	TeNaoEmbarque                                               // 110115 -- NaoEmbarque
	TeMDFeCancelado2                                            // 310611 -- MDFeCancelado2
	TeMDFeAutorizadoComCTe                                      // 610614 -- MDFeAutorizadoComCTe
	TeRegPasNfeProMDFe                                          // 610510 -- RegPasNfeProMDFe
	TeRegPasNfeProMDFeCte                                       // 610514 -- RegPasNfeProMDFeCte
	TeRegPasAutMDFeComCte                                       // 610554 -- RegPasAutMDFeComCte
	TeCancelamentoMDFeAutComCTe                                 // 610615 -- CancelamentoMDFeAutComCTe
	TeAverbacaoExportacao                                       // 790700 -- AverbacaoExportacao
	TeAutCteComplementar                                        // 240130 -- AutCteComplementar
	TeCancCteComplementar                                       // 240131 -- CancCteComplementar
	TeCTeSubstituicao                                           // 240140 -- CTeSubstituicao
	TeCTeAnulacao                                               // 240150 -- CTeAnulacao
	TeLiberacaoEPEC                                             // 240160 -- LiberacaoEPEC
	TeLiberacaoPrazoCanc                                        // 240170 -- LiberacaoPrazoCanc
	TeAutorizadoRedespacho                                      // 440130 -- AutorizadoRedespacho
	TeAutorizadoRedespIntermed                                  // 440140 -- AutorizadoRedespIntermed
	TeAutorizadoSubcontratacao                                  // 440150 -- AutorizadoSubcontratacao
	TeAutorizadoServMultimodal                                  // 440160 -- AutorizadoServMultimodal
	TeCancSubst                                                 // 110112 -- CancelamentoPorSubstituicao
	TeAlteracaoPoltrona                                         // 110116 -- AlteracaoPoltrona
	TeComprEntrega                                              // 110180 -- ComprEntrega
	TeCancComprEntrega                                          // 110181 -- CancComprEntrega
	TeInclusaoDFe                                               // 110115 -- InclusaoDFe
	TeAutorizadoSubstituicao                                    // 240140 -- AutorizadoSubstituicao
	TeAutorizadoAjuste                                          // 240150 -- AutorizadoAjuste
	TeLiberacaoPrazoCancelado                                   // 240170 -- LiberacaoPrazoCancelado
	TePagamentoOperacao                                         // 110116 -- PagamentoOperacao
	TeExcessoBagagem                                            // 110117 -- ExcessoBagagem
	TeEncerramentoFisco                                         // 310112 -- EncerramentoFisco
	TeComprEntregaNFe                                           // 110130 -- ComprEntregaNFe
	TeCancComprEntregaNFe                                       // 110131 -- CancComprEntregaNFe
	TeAtorInteressadoNFe                                        // 110150 -- AtorInteressadoNFe
	TeComprEntregaCTe                                           // 610130 -- ComprEntregaCTe
	TeCancComprEntregaCTe                                       // 610131 -- CancComprEntregaCTe
	TeConfirmaServMDFe                                          // 110117 -- ConfirmaServMDFe
	TeAlteracaoPagtoServMDFe                                    // 110118 -- AlteracaoPagtoServMDFe
	TeCancPrestDesacordo                                        // 610111 -- CancPrestDesacordo
	TeInsucessoEntregaCTe                                       // 110190 -- InsucessoEntregaCTe
	TeCancInsucessoEntregaCTe                                   // 110191 -- CancInsucessoEntregaCTe
	TeInsucessoEntregaNFe                                       // 110192 -- InsucessoEntregaNFe
	TeCancInsucessoEntregaNFe                                   // 110193 -- CancInsucessoEntregaNFe
	TeConcFinanceira                                            // 110750 -- ConcFinanceira
	TeCancConcFinanceira                                        // 110751 -- CancConcFinanceira
	TeRegistroPassagemMDFe                                      // 510630 -- RegistroPassagemMDFe
	TeCancGenerico                                              // 110001 -- CancGenerico
	TePagIntegLibCredPresAdq                                    // 112110 -- PagIntegLibCredPresAdq
	TeImporALCZFM                                               // 112120 -- ImporALCZFM
	TePerecPerdaRouboFurtoTranspContratFornec                   // 112130 -- PerecPerdaRouboFurtoTranspContratFornec
	TeFornecNaoRealizPagAntec                                   // 112140 -- FornecNaoRealizPagAntec
	TeSolicApropCredPres                                        // 211110 -- SolicApropCredPres
	TeDestItemConsPessoal                                       // 211120 -- DestItemConsPessoal
	TePerecPerdaRouboFurtoTranspContratAqu                      // 211124 -- PerecPerdaRouboFurtoTranspContratAqu
	TeAceiteDebitoApuracaoNotaCredito                           // 211128 -- AceiteDebitoApuracaoNotaCredito
	TeImobilizacaoItem                                          // 211130 -- ImobilizacaoItem
	TeSolicApropCredCombustivel                                 // 211140 -- SolicApropCredCombustivel
	TeSolicApropCredBensServicos                                // 211150 -- SolicApropCredBensServicos
	TeManifPedTransfCredIBSSucessao                             // 212110 -- ManifPedTransfCredIBSSucessao
	TeManifPedTransfCredCBSSucessao                             // 212120 -- ManifPedTransfCredCBSSucessao
	TeAtualizacaoDataPrevisaoEntrega                            // 112150 -- AtualizacaoDataPrevisaoEntrega
	TeVinculoPgto                                               // 110300 -- VinculoPagamento
	TeCancVinculoPgto                                           // 110301 -- CancVinculoPgto
	// _fimTipoEvento conta os membros declarados; nao e valor do leiaute.
	_fimTipoEvento
)

var tipoEventoCodigos = [...]string{
	"-99999",
	"110110",
	"110111",
	"210200",
	"210210",
	"210220",
	"210240",
	"110112",
	"110113",
	"110114",
	"110160",
	"310620",
	"510620",
	"110140",
	"610600",
	"610501",
	"610550",
	"610601",
	"610611",
	"990900",
	"111500",
	"111501",
	"111502",
	"111503",
	"411500",
	"411501",
	"411502",
	"411503",
	"610500",
	"990910",
	"000000",
	"610610",
	"610110",
	"110170",
	"310610",
	"110115",
	"310611",
	"610614",
	"610510",
	"610514",
	"610554",
	"610615",
	"790700",
	"240130",
	"240131",
	"240140",
	"240150",
	"240160",
	"240170",
	"440130",
	"440140",
	"440150",
	"440160",
	"110112",
	"110116",
	"110180",
	"110181",
	"110115",
	"240140",
	"240150",
	"240170",
	"110116",
	"110117",
	"310112",
	"110130",
	"110131",
	"110150",
	"610130",
	"610131",
	"110117",
	"110118",
	"610111",
	"110190",
	"110191",
	"110192",
	"110193",
	"110750",
	"110751",
	"510630",
	"110001",
	"112110",
	"112120",
	"112130",
	"112140",
	"211110",
	"211120",
	"211124",
	"211128",
	"211130",
	"211140",
	"211150",
	"212110",
	"212120",
	"112150",
	"110300",
	"110301",
}

var tipoEventoDescricoes = [...]string{
	"NaoMapeado",
	"CCe",
	"Cancelamento",
	"ManifDestConfirmacao",
	"ManifDestCiencia",
	"ManifDestDesconhecimento",
	"ManifDestOperNaoRealizada",
	"Encerramento",
	"EPEC",
	"InclusaoCondutor",
	"MultiModal",
	"RegistroPassagem",
	"RegistroPassagemBRId",
	"EPECNFe",
	"RegistroCTe",
	"RegistroPassagemNFeCancelado",
	"RegistroPassagemNFeRFID",
	"CTeCancelado",
	"MDFeCancelado",
	"VistoriaSuframa",
	"PedProrrog1",
	"PedProrrog2",
	"CanPedProrrog1",
	"CanPedProrrog2",
	"EventoFiscoPP1",
	"EventoFiscoPP2",
	"EventoFiscoCPP1",
	"EventoFiscoCPP2",
	"RegistroPassagemNFe",
	"ConfInternalizacao",
	"CTeAutorizado",
	"MDFeAutorizado",
	"PrestDesacordo",
	"GTV",
	"MDFeAutorizado2",
	"NaoEmbarque",
	"MDFeCancelado2",
	"MDFeAutorizadoComCTe",
	"RegPasNfeProMDFe",
	"RegPasNfeProMDFeCte",
	"RegPasAutMDFeComCte",
	"CancelamentoMDFeAutComCTe",
	"AverbacaoExportacao",
	"AutCteComplementar",
	"CancCteComplementar",
	"CTeSubstituicao",
	"CTeAnulacao",
	"LiberacaoEPEC",
	"LiberacaoPrazoCanc",
	"AutorizadoRedespacho",
	"AutorizadoRedespIntermed",
	"AutorizadoSubcontratacao",
	"AutorizadoServMultimodal",
	"CancelamentoPorSubstituicao",
	"AlteracaoPoltrona",
	"ComprEntrega",
	"CancComprEntrega",
	"InclusaoDFe",
	"AutorizadoSubstituicao",
	"AutorizadoAjuste",
	"LiberacaoPrazoCancelado",
	"PagamentoOperacao",
	"ExcessoBagagem",
	"EncerramentoFisco",
	"ComprEntregaNFe",
	"CancComprEntregaNFe",
	"AtorInteressadoNFe",
	"ComprEntregaCTe",
	"CancComprEntregaCTe",
	"ConfirmaServMDFe",
	"AlteracaoPagtoServMDFe",
	"CancPrestDesacordo",
	"InsucessoEntregaCTe",
	"CancInsucessoEntregaCTe",
	"InsucessoEntregaNFe",
	"CancInsucessoEntregaNFe",
	"ConcFinanceira",
	"CancConcFinanceira",
	"RegistroPassagemMDFe",
	"CancGenerico",
	"PagIntegLibCredPresAdq",
	"ImporALCZFM",
	"PerecPerdaRouboFurtoTranspContratFornec",
	"FornecNaoRealizPagAntec",
	"SolicApropCredPres",
	"DestItemConsPessoal",
	"PerecPerdaRouboFurtoTranspContratAqu",
	"AceiteDebitoApuracaoNotaCredito",
	"ImobilizacaoItem",
	"SolicApropCredCombustivel",
	"SolicApropCredBensServicos",
	"ManifPedTransfCredIBSSucessao",
	"ManifPedTransfCredCBSSucessao",
	"AtualizacaoDataPrevisaoEntrega",
	"VinculoPagamento",
	"CancVinculoPgto",
}

// Codigo devolve o codigo de seis digitos do leiaute.
// Porte de TipoEventoToStr.
func (v TipoEvento) Codigo() string {
	if int(v) < 0 || int(v) >= len(tipoEventoCodigos) {
		return ""
	}
	return tipoEventoCodigos[v]
}

// String devolve o nome curto do evento, que e o que o ACBr usa para
// montar nome de arquivo (TipoEventoToDescricaoStr).
//
// NAO e o codigo -- quem quiser o codigo chama Codigo(). A distincao
// importa porque os dois sao string e trocar um pelo outro compila.
func (v TipoEvento) String() string {
	if int(v) < 0 || int(v) >= len(tipoEventoDescricoes) {
		return ""
	}
	return tipoEventoDescricoes[v]
}

// ParseTipoEventoNFe converte o codigo do leiaute no evento, usando a
// tabela da NF-e. Porte de StrToTpEventoNFe
// (ACBrNFe.Conversao.pas:1128).
//
// A tabela da NF-e e mais larga que a lista de eventos que ela GERA: ela
// inclui registros de passagem, averbacao de exportacao e eventos de MDFe
// e CTe que chegam amarrados a chave da nota. O leitor precisa deles; o
// gerador, nao.
// DIVERGENCIA: o `-99999` do sentinela fica de FORA da tabela, e por isso
// devolve erro. O StrToTpEventoNFe o inclui e devolve teNaoMapeado com
// ok=True -- ou seja, trata "nao mapeado" como entrada valida. A SEFAZ
// nunca o devolve; aceitar o sentinela so criaria um caminho em que um
// evento sem tipo passa como se tivesse um.
func ParseTipoEventoNFe(s string) (TipoEvento, error) {
	if t, ok := tipoEventoNFePorCodigo[s]; ok {
		return t, nil
	}
	return TeNaoMapeado, fmt.Errorf("pcn: TipoEvento invalido para a NF-e: %q", s)
}

var tipoEventoNFePorCodigo = map[string]TipoEvento{
	"110110": TeCCe,
	"110111": TeCancelamento,
	"110112": TeCancSubst,
	"110140": TeEPECNFe,
	"111500": TePedProrrog1,
	"111501": TePedProrrog2,
	"111502": TeCanPedProrrog1,
	"111503": TeCanPedProrrog2,
	"210200": TeManifDestConfirmacao,
	"210210": TeManifDestCiencia,
	"210220": TeManifDestDesconhecimento,
	"210240": TeManifDestOperNaoRealizada,
	"610600": TeRegistroCTe,
	"610614": TeMDFeAutorizadoComCTe,
	"790700": TeAverbacaoExportacao,
	"990900": TeVistoriaSuframa,
	"990910": TeConfInternalizacao,
	"110180": TeComprEntrega,
	"610554": TeRegPasAutMDFeComCte,
	"610510": TeRegPasNfeProMDFe,
	"610615": TeCancelamentoMDFeAutComCTe,
	"610610": TeMDFeAutorizado,
	"110130": TeComprEntregaNFe,
	"110131": TeCancComprEntregaNFe,
	"110150": TeAtorInteressadoNFe,
	"610130": TeComprEntregaCTe,
	"610131": TeCancComprEntregaCTe,
	"610601": TeCTeCancelado,
	"110192": TeInsucessoEntregaNFe,
	"110193": TeCancInsucessoEntregaNFe,
	"610514": TeRegPasNfeProMDFeCte,
	"610500": TeRegistroPassagemNFe,
	"110750": TeConcFinanceira,
	"110751": TeCancConcFinanceira,
	"510630": TeRegistroPassagemMDFe,
	"110001": TeCancGenerico,
	"112110": TePagIntegLibCredPresAdq,
	"112120": TeImporALCZFM,
	"112130": TePerecPerdaRouboFurtoTranspContratFornec,
	"112140": TeFornecNaoRealizPagAntec,
	"211110": TeSolicApropCredPres,
	"211120": TeDestItemConsPessoal,
	"211124": TePerecPerdaRouboFurtoTranspContratAqu,
	"211128": TeAceiteDebitoApuracaoNotaCredito,
	"211130": TeImobilizacaoItem,
	"211140": TeSolicApropCredCombustivel,
	"211150": TeSolicApropCredBensServicos,
	"212110": TeManifPedTransfCredIBSSucessao,
	"212120": TeManifPedTransfCredCBSSucessao,
	"112150": TeAtualizacaoDataPrevisaoEntrega,
}
