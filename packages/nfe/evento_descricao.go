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

package nfe

import (
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Descricao de evento -- o conteudo da tag <descEvento>.
//
// Porte de TInfEvento.getDescEvento (ACBrNFe.EventoClass.pas:715). Nao e o
// mesmo texto do TipoEvento.String() do pcn: aquele e o nome CURTO, que o
// ACBr usa para montar nome de arquivo ("CCe"); este e o texto que vai NO
// XML ("Carta de Correcao"), e a SEFAZ o valida.
//
// Dois detalhes que um porte perde:
//
//   - os eventos da Reforma Tributaria levam ACENTOS. O ACBr desliga o
//     RetirarAcentos so para eles (:947-957), com o comentario "Nao funciona
//     sem acentos". Os classicos vao sem;
//   - varios tipos COMPARTILHAM descricao: os dois PedProrrog, os dois
//     CanPedProrrog, os quatro EventoFisco. Um mapa resolve; um switch com
//     um ramo por tipo convida a divergir.

// descricaoDoEvento devolve o texto do descEvento.
//
// Quando o documento ja traz DescEvento preenchido, ele PREVALECE -- o
// ACBr nao oferece essa saida, mas um evento novo publicado pela SEFAZ
// precisa de alguma forma de sair sem esperar o porte.
func descricaoDoEvento(inf InfEvento) string {
	if d := strings.TrimSpace(inf.DetEvento.DescEvento); d != "" {
		return d
	}
	return descricaoPorEvento[inf.TpEvento]
}

// DescricaoDoTipoEvento e a versao publica, para quem monta o detEvento por
// fora ou quer rotular um evento lido.
func DescricaoDoTipoEvento(t pcn.TipoEvento) string {
	return descricaoPorEvento[t]
}

// EventoLevaAcentos informa se o descEvento deste tipo e emitido COM
// acentos.
//
// Porte de DescricaoEventoPrecisaDeAcentos (ACBrNFe.EnvEvento.pas:947).
// A lista e dos eventos da Reforma Tributaria mais o ator interessado e o
// cancelamento da conciliacao financeira -- e nao coincide com "ser da
// Reforma": o teCancGenerico, que e da Reforma, NAO esta nela.
func EventoLevaAcentos(t pcn.TipoEvento) bool {
	return eventoComAcentos[t]
}

var eventoComAcentos = map[pcn.TipoEvento]bool{
	pcn.TeAtorInteressadoNFe:                      true,
	pcn.TeCancConcFinanceira:                      true,
	pcn.TePagIntegLibCredPresAdq:                  true,
	pcn.TeImporALCZFM:                             true,
	pcn.TePerecPerdaRouboFurtoTranspContratFornec: true,
	pcn.TeFornecNaoRealizPagAntec:                 true,
	pcn.TeSolicApropCredPres:                      true,
	pcn.TeDestItemConsPessoal:                     true,
	pcn.TePerecPerdaRouboFurtoTranspContratAqu:    true,
	pcn.TeAceiteDebitoApuracaoNotaCredito:         true,
	pcn.TeImobilizacaoItem:                        true,
	pcn.TeSolicApropCredCombustivel:               true,
	pcn.TeSolicApropCredBensServicos:              true,
	pcn.TeAtualizacaoDataPrevisaoEntrega:          true,
	pcn.TeManifPedTransfCredIBSSucessao:           true,
	pcn.TeManifPedTransfCredCBSSucessao:           true,
}

var descricaoPorEvento = map[pcn.TipoEvento]string{
	pcn.TeCCe:                                     "Carta de Correcao",
	pcn.TeCancelamento:                            "Cancelamento",
	pcn.TeCancSubst:                               "Cancelamento por substituicao",
	pcn.TeManifDestConfirmacao:                    "Confirmacao da Operacao",
	pcn.TeManifDestCiencia:                        "Ciencia da Operacao",
	pcn.TeManifDestDesconhecimento:                "Desconhecimento da Operacao",
	pcn.TeManifDestOperNaoRealizada:               "Operacao nao Realizada",
	pcn.TeEPECNFe:                                 "EPEC",
	pcn.TeEPEC:                                    "EPEC",
	pcn.TeMultiModal:                              "Registro Multimodal",
	pcn.TeRegistroPassagem:                        "Registro de Passagem",
	pcn.TeRegistroPassagemNFe:                     "Registro de Passagem NF-e",
	pcn.TeRegistroPassagemBRId:                    "Registro de Passagem BRId",
	pcn.TeEncerramento:                            "Encerramento",
	pcn.TeInclusaoCondutor:                        "Inclusao Condutor",
	pcn.TeRegistroCTe:                             "CT-e Autorizado para NF-e",
	pcn.TeRegistroPassagemNFeCancelado:            "Registro de Passagem para NF-e Cancelado",
	pcn.TeRegistroPassagemNFeRFID:                 "Registro de Passagem para NF-e RFID",
	pcn.TeCTeAutorizado:                           "CT-e Autorizado",
	pcn.TeCTeCancelado:                            "CT-e Cancelado",
	pcn.TeMDFeAutorizado:                          "MDF-e Autorizado",
	pcn.TeMDFeAutorizado2:                         "MDF-e Autorizado",
	pcn.TeMDFeCancelado:                           "MDF-e Cancelado",
	pcn.TeMDFeCancelado2:                          "MDF-e Cancelado",
	pcn.TeVistoriaSuframa:                         "Vistoria SUFRAMA",
	pcn.TePedProrrog1:                             "Pedido de Prorrogacao",
	pcn.TePedProrrog2:                             "Pedido de Prorrogacao",
	pcn.TeCanPedProrrog1:                          "Cancelamento de Pedido de Prorrogacao",
	pcn.TeCanPedProrrog2:                          "Cancelamento de Pedido de Prorrogacao",
	pcn.TeEventoFiscoPP1:                          "Evento Fisco",
	pcn.TeEventoFiscoPP2:                          "Evento Fisco",
	pcn.TeEventoFiscoCPP1:                         "Evento Fisco",
	pcn.TeEventoFiscoCPP2:                         "Evento Fisco",
	pcn.TeConfInternalizacao:                      "Confirmacao de Internalizacao da Mercadoria na SUFRAMA",
	pcn.TeComprEntrega:                            "Comprovante de Entrega do CT-e",
	pcn.TeComprEntregaNFe:                         "Comprovante de Entrega da NF-e",
	pcn.TeCancComprEntregaNFe:                     "Cancelamento Comprovante de Entrega da NF-e",
	pcn.TeAtorInteressadoNFe:                      "Ator interessado na NF-e",
	pcn.TeInsucessoEntregaNFe:                     "Insucesso na Entrega da NF-e",
	pcn.TeCancInsucessoEntregaNFe:                 "Cancelamento Insucesso na Entrega da NF-e",
	pcn.TeConcFinanceira:                          "ECONF",
	pcn.TeCancConcFinanceira:                      "Cancelamento Conciliação Financeira",
	pcn.TeCancGenerico:                            "Cancelamento de Evento",
	pcn.TePagIntegLibCredPresAdq:                  "Informação de efetivo pagamento integral para liberar crédito presumido do adquirente",
	pcn.TeImporALCZFM:                             "Importação em ALC/ZFM não convertida em isenção",
	pcn.TePerecPerdaRouboFurtoTranspContratFornec: "Perecimento, perda, roubo ou furto durante o transporte contratado pelo fornecedor",
	pcn.TeFornecNaoRealizPagAntec:                 "Fornecimento não realizado com pagamento antecipado",
	pcn.TeSolicApropCredPres:                      "Solicitação de Apropriação de crédito presumido",
	pcn.TeDestItemConsPessoal:                     "Destinação de item para consumo pessoal",
	pcn.TePerecPerdaRouboFurtoTranspContratAqu:    "Perecimento, perda, roubo ou furto durante o transporte contratado pelo adquirente",
	pcn.TeAceiteDebitoApuracaoNotaCredito:         "Aceite de débito na apuração por emissão de nota de crédito",
	pcn.TeImobilizacaoItem:                        "Imobilização de Item",
	pcn.TeSolicApropCredCombustivel:               "Solicitação de Apropriação de Crédito de Combustível",
	pcn.TeSolicApropCredBensServicos:              "Solicitação de Apropriação de Crédito para bens e serviços que dependem de atividade do adquirente",
	pcn.TeManifPedTransfCredIBSSucessao:           "Manifestação sobre Pedido de Transferência de Crédito de IBS em Operação de Sucessão",
	pcn.TeManifPedTransfCredCBSSucessao:           "Manifestação sobre Pedido de Transferência de Crédito de CBS em Operação de Sucessão",
	pcn.TeAtualizacaoDataPrevisaoEntrega:          "Atualização da Data de Previsão de Entrega",
}
