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

import (
	"strconv"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Estruturas de evento da NFGas.
// Porte de ACBrNFGas.EventoClass.pas e ACBrNFGas.RetEnvEvento.pas.

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
	ChNFGas    string           // chNFGas
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
	if len(i.ChNFGas) < 2 {
		return 0
	}
	v, err := strconv.Atoi(i.ChNFGas[:2])
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
// Porte de TInfEvento.DescricaoTipoEvento.
func DescricaoTipoEvento(t TipoEvento) string {
	if t == TeCancelamento {
		return "CANCELAMENTO DE NFGas"
	}
	return "Nao Definido"
}

// EventoNFGas e o evento enviado a SEFAZ. Porte do item de TEventoNFGas.
type EventoNFGas struct {
	Versao    string        // atributo versao de eventoNFGas
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
	ChNFGas     string           // chNFGas
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
// Os cStat de aceite de evento sao 135 (registrado e vinculado) e 136
// (registrado, nao vinculado).
func (r *RetInfEvento) Registrado() bool {
	if r == nil {
		return false
	}
	return r.CStat == 135 || r.CStat == 136
}

// RetEventoNFGas e o retorno de um evento, com o evento enviado quando o
// XML e um procEventoNFGas. Porte de TRetEventoNFGas.
type RetEventoNFGas struct {
	Versao       string        // atributo versao
	RetInfEvento RetInfEvento  // retEventoNFGas/infEvento
	Signature    pcn.Signature // Signature do retorno
	XML          string        // XML de origem

	// Evento e o evento ENVIADO, lido de procEventoNFGas/eventoNFGas.
	//
	// ACRESCIMO em relacao ao ACBr: TRetEventoNFGas so le a parte de
	// retorno. Sem esta parte, uma importacao de eventos de cancelamento
	// nao tem acesso a justificativa (xJust), que so existe no enviado.
	Evento EventoNFGas
	// TemEvento informa se a parte enviada estava presente no XML.
	TemEvento bool
}

// ChaveAcesso devolve a chave do documento a que o evento se refere,
// preferindo a do retorno e caindo para a do evento enviado.
func (r *RetEventoNFGas) ChaveAcesso() string {
	if r == nil {
		return ""
	}
	if r.RetInfEvento.ChNFGas != "" {
		return pcn.RemoverLiteralChave(r.RetInfEvento.ChNFGas)
	}
	return pcn.RemoverLiteralChave(r.Evento.InfEvento.ChNFGas)
}

// Cancelamento informa se o evento e um cancelamento da NFGas.
func (r *RetEventoNFGas) Cancelamento() bool {
	if r == nil {
		return false
	}
	return r.RetInfEvento.TpEvento == TeCancelamento ||
		r.Evento.InfEvento.TpEvento == TeCancelamento
}

// Justificativa devolve a justificativa informada no evento enviado.
// Vazia quando o XML so traz a parte de retorno.
func (r *RetEventoNFGas) Justificativa() string {
	if r == nil {
		return ""
	}
	return r.Evento.InfEvento.DetEvento.XJust
}
