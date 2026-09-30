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

import (
	"strconv"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Estruturas de evento da NFAg.
// Porte de ACBrNFAg.EventoClass.pas e ACBrNFAg.RetEnvEvento.pas.

// DetEvento e o detalhe do evento enviado. Porte de TDetEvento.
type DetEvento struct {
	Versao            string // atributo versao de detEvento
	DescEvento        string // descEvento
	NProt             string // nProt
	XJust             string // xJust
	IDPedidoCancelado string // idPedidoCancelado
}

// InfEvento sao os dados do evento enviado. Porte de TInfEvento.
type InfEvento struct {
	ID         string           // atributo Id
	COrgao     int              // cOrgao
	TpAmb      pcn.TipoAmbiente // tpAmb
	CNPJ       string           // CNPJ
	ChNFAg     string           // chNFAg
	DhEvento   time.Time        // dhEvento
	TpEvento   TipoEvento       // tpEvento
	NSeqEvento int              // nSeqEvento
	DetEvento  DetEvento        // detEvento
}

// COrgaoEfetivo devolve cOrgao ou, quando ele nao veio preenchido, os dois
// primeiros digitos da chave de acesso. Porte de TInfEvento.getcOrgao.
func (i *InfEvento) COrgaoEfetivo() int {
	if i == nil {
		return 0
	}
	if i.COrgao != 0 {
		return i.COrgao
	}
	if len(i.ChNFAg) < 2 {
		return 0
	}
	v, err := strconv.Atoi(i.ChNFAg[:2])
	if err != nil {
		return 0
	}
	return v
}

// DescEvento devolve a descricao curta do tipo de evento.
// Porte de TInfEvento.getDescEvento, que so mapeia o cancelamento.
func (i *InfEvento) DescEvento() string {
	if i == nil || i.TpEvento != TeCancelamento {
		return ""
	}
	return "Cancelamento"
}

// DescricaoTipoEvento devolve a descricao longa do tipo de evento.
// Porte de TInfEvento.DescricaoTipoEvento -- o literal do ACBr diz
// "CANCELAMENTO DE NF3-e" (copy-paste da NF3e), REPLICADO: o texto e so
// descritivo (log/nome de arquivo) e nao entra no XML.
func DescricaoTipoEvento(t TipoEvento) string {
	if t == TeCancelamento {
		return "CANCELAMENTO DE NF3-e"
	}
	return "Nao Definido"
}

// EventoNFAg e o evento enviado a SEFAZ. Porte do item de TEventoNFAg.
type EventoNFAg struct {
	Versao    string        // atributo versao de eventoNFAg
	InfEvento InfEvento     // infEvento
	Signature pcn.Signature // Signature
}

// RetInfEvento e o retorno do processamento de um evento.
// Porte de TRetInfEvento.
//
// CNPJDest, EmailDest e COrgaoAutor existem na classe do ACBr e NAO sao
// lidos pelo leitor dele -- ficam aqui pelo mesmo motivo, para nao mudar a
// forma da struct, e sao preenchidos por esta implementacao (ver
// evento_reader.go).
type RetInfEvento struct {
	ID          string           // atributo Id
	TpAmb       pcn.TipoAmbiente // tpAmb
	VerAplic    string           // verAplic
	COrgao      int              // cOrgao
	CStat       int              // cStat
	XMotivo     string           // xMotivo
	ChNFAg      string           // chNFAg
	TpEvento    TipoEvento       // tpEvento
	XEvento     string           // xEvento
	NSeqEvento  int              // nSeqEvento
	CNPJDest    string           // CNPJDest
	EmailDest   string           // emailDest
	COrgaoAutor int              // cOrgaoAutor
	DhRegEvento time.Time        // dhRegEvento
	NProt       string           // nProt
	XML         string           // OuterXml de infEvento
	NomeArquivo string           // preenchido na leitura de arquivo
}

// Registrado informa se o evento foi aceito pela SEFAZ.
// Os cStat de aceite sao 135 (registrado e vinculado), 136 (registrado,
// nao vinculado) e 155 (cancelamento homologado fora de prazo) -- o mesmo
// conjunto do ACBrNFAgWebServices.pas:1573.
func (r *RetInfEvento) Registrado() bool {
	if r == nil {
		return false
	}
	return r.CStat == 135 || r.CStat == 136 || r.CStat == 155
}

// RetEventoNFAg e o retorno de um evento, com o evento enviado quando o
// XML e um procEventoNFAg. Porte de TRetEventoNFAg.
type RetEventoNFAg struct {
	Versao       string        // atributo versao
	RetInfEvento RetInfEvento  // retEventoNFAg/infEvento
	Signature    pcn.Signature // Signature do retorno
	XML          string        // XML de origem

	// Evento e o evento ENVIADO, lido de procEventoNFAg/eventoNFAg.
	//
	// ACRESCIMO em relacao ao ACBr: TRetEventoNFAg so le a parte de
	// retorno. Sem esta parte, uma importacao de eventos de cancelamento
	// nao tem acesso a justificativa (xJust), que so existe no enviado.
	Evento EventoNFAg
	// TemEvento informa se a parte enviada estava presente no XML.
	TemEvento bool
}

// ChaveAcesso devolve a chave do documento a que o evento se refere,
// preferindo a do retorno e caindo para a do evento enviado.
func (r *RetEventoNFAg) ChaveAcesso() string {
	if r == nil {
		return ""
	}
	if r.RetInfEvento.ChNFAg != "" {
		return pcn.RemoverLiteralChave(r.RetInfEvento.ChNFAg)
	}
	return pcn.RemoverLiteralChave(r.Evento.InfEvento.ChNFAg)
}

// Cancelamento informa se o evento e um cancelamento da NFAg.
func (r *RetEventoNFAg) Cancelamento() bool {
	if r == nil {
		return false
	}
	return r.RetInfEvento.TpEvento == TeCancelamento ||
		r.Evento.InfEvento.TpEvento == TeCancelamento
}

// Justificativa devolve a justificativa informada no evento enviado.
// Vazia quando o XML so traz a parte de retorno.
func (r *RetEventoNFAg) Justificativa() string {
	if r == nil {
		return ""
	}
	return r.Evento.InfEvento.DetEvento.XJust
}
