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
	"fmt"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Geracao do XML de eventoNFGas -- porte de TEventoNFGas.GerarXml +
// Gerar_InfEvento/Gerar_DetEvento/Gerar_Evento_Cancelamento
// (ACBrNFGas.EnvEvento.pas). O unico tipo de evento do leiaute e o
// cancelamento; qualquer outro devolve erro, como o raise do original.

// GerarXMLEvento gera o XML de eventoNFGas (sem assinatura -- use
// AssinarEvento sobre o resultado). Como no original, tem efeito colateral:
// monta e grava InfEvento.ID ("ID" + tpEvento + chave + nSeq com 2 digitos).
func GerarXMLEvento(e *EventoNFGas) (string, error) {
	evento, err := gerarElemEvento(e, true)
	if err != nil {
		return "", err
	}
	return evento.XML(), nil
}

// gerarElemEvento monta o elemento eventoNFGas. O namespace so e declarado
// aqui quando o evento vai avulso; dentro de um procEventoNFGas ele ja foi
// declarado na raiz, e repeti-lo seria redundante.
func gerarElemEvento(e *EventoNFGas, comNamespace bool) (*pcn.Elem, error) {
	if e == nil {
		return nil, ErrXMLVazio
	}
	if e.InfEvento.TpEvento != TeCancelamento {
		return nil, fmt.Errorf("nfgas: tipo de evento nao implementado para NFGas: %s",
			e.InfEvento.TpEvento.String())
	}
	if e.Versao == "" {
		e.Versao = "1.00"
	}

	chave := pcn.RemoverLiteralChave(e.InfEvento.ChNFGas)
	e.InfEvento.ID = "ID" + e.InfEvento.TpEvento.String() + chave +
		fmt.Sprintf("%02d", e.InfEvento.NSeqEvento)

	// CNPJ vazio e extraido da chave de acesso (posicoes 7..20), como o
	// ExtrairCNPJCPFChaveAcesso do original.
	cnpj := pcn.OnlyCPFCNPJAlphaNum(e.InfEvento.CNPJ)
	if cnpj == "" && len(chave) >= 20 {
		cnpj = chave[6:20]
	}

	cOrgao := e.InfEvento.COrgaoEfetivo()
	uf := pcn.SiglaUF(cOrgao)

	evento := pcn.NovoElem("eventoNFGas")
	if comNamespace {
		evento.Attr("xmlns", Namespace)
	}
	evento.
		Attr("versao", e.Versao).
		Filho(pcn.NovoElem("infEvento").
			Attr("Id", e.InfEvento.ID).
			Filho(pcn.NodeInt("cOrgao", cOrgao, 1, true)).
			Filho(pcn.NodeStr("tpAmb", e.InfEvento.TpAmb.String(), true)).
			Filho(pcn.NodeStrSemFiltro("CNPJ", cnpj, true)).
			Filho(pcn.NodeStrSemFiltro("chNFGas", e.InfEvento.ChNFGas, true)).
			Filho(pcn.NodeStr("dhEvento", pcn.FormatarDataHoraXML(e.InfEvento.DhEvento, uf), true)).
			Filho(pcn.NodeStrSemFiltro("tpEvento", e.InfEvento.TpEvento.String(), true)).
			Filho(pcn.NodeInt("nSeqEvento", e.InfEvento.NSeqEvento, 1, true)).
			Filho(pcn.NovoElem("detEvento").
				Attr("versaoEvento", e.Versao).
				Filho(pcn.NovoElem("evCancNFGas").
					// descEvento e SEMPRE a descricao canonica do tipo,
					// como a property DescEvento do original -- texto do
					// usuario nao e transmitido (o XSD fixa o valor)
					Filho(pcn.NodeStr("descEvento", e.InfEvento.DescEvento(), true)).
					Filho(pcn.NodeStrSemFiltro("nProt", e.InfEvento.DetEvento.NProt, true)).
					Filho(pcn.NodeStr("xJust", e.InfEvento.DetEvento.XJust, true)))))

	if e.Signature.Assinada() {
		evento.Filho(gerarSignature(&e.Signature))
	}
	return evento, nil
}

// GerarXMLProcEvento gera o XML de procEventoNFGas -- o envelope com o evento
// ENVIADO e o retorno da SEFAZ, que e o que o contribuinte arquiva e o que um
// importador recebe. Simetrico ao GerarXMLProc do documento.
//
// ACRESCIMO em relacao ao ACBr, que le o procEventoNFGas mas nao o escreve
// (TRetEventoNFGas so tem leitura). Sem a parte enviada nao ha xJust, entao
// retorno sem evento e recusado em vez de gerar envelope pela metade.
func GerarXMLProcEvento(r *RetEventoNFGas) (string, error) {
	if r == nil {
		return "", ErrXMLVazio
	}
	if !r.TemEvento {
		return "", ErrEventoAusente
	}
	if r.RetInfEvento.NProt == "" {
		return "", ErrProtocoloAusente
	}

	versao := r.Versao
	if versao == "" {
		versao = "1.00"
	}
	if r.Evento.Versao == "" {
		r.Evento.Versao = versao
	}

	evento, err := gerarElemEvento(&r.Evento, false)
	if err != nil {
		return "", err
	}

	return pcn.NovoElem("procEventoNFGas").
		Attr("versao", versao).
		Attr("xmlns", Namespace).
		Filho(evento).
		Filho(gerarRetEvento(r, versao)).XML(), nil
}

// gerarRetEvento monta o retEventoNFGas a partir do retorno lido/preenchido.
// Como o Gerar_ProcNFGas do documento, grava as tags na ordem do XSD; as
// opcionais (CNPJDest, emailDest, cOrgaoAutor) so saem quando preenchidas.
func gerarRetEvento(r *RetEventoNFGas, versao string) *pcn.Elem {
	ret := r.RetInfEvento
	uf := pcn.SiglaUF(ret.COrgao)

	// Campos nao preenchidos no retorno caem para o evento ENVIADO, que o
	// envelope obriga a existir. Sem isso, um retorno parcialmente montado
	// (so nProt e cStat, o caso comum de quem preenche a mao) emitiria
	// tpEvento "-99999" -- o String() do TeNaoMapeado, que e o zero value --
	// dentro de um envelope que passa por todas as guardas acima.
	//
	// TpAmb fica de fora de proposito: o zero value de pcn.TipoAmbiente e
	// TaProducao, valor legitimo, entao "nao informado" e indistinguivel de
	// "producao" e qualquer fallback aqui seria chute.
	id := ret.ID
	if id == "" {
		id = r.Evento.InfEvento.ID
	}
	chave := ret.ChNFGas
	if chave == "" {
		chave = r.Evento.InfEvento.ChNFGas
	}
	if ret.TpEvento == TeNaoMapeado {
		ret.TpEvento = r.Evento.InfEvento.TpEvento
	}
	if ret.NSeqEvento == 0 {
		ret.NSeqEvento = r.Evento.InfEvento.NSeqEvento
	}
	if ret.COrgao == 0 {
		ret.COrgao = r.Evento.InfEvento.COrgaoEfetivo()
		uf = pcn.SiglaUF(ret.COrgao)
	}
	if ret.XEvento == "" {
		ret.XEvento = r.Evento.InfEvento.DescEvento()
	}

	inf := pcn.NovoElem("infEvento").Attr("Id", id).
		Filho(pcn.NodeStr("tpAmb", ret.TpAmb.String(), true)).
		Filho(pcn.NodeStr("verAplic", ret.VerAplic, true)).
		Filho(pcn.NodeInt("cOrgao", ret.COrgao, 1, true)).
		Filho(pcn.NodeInt("cStat", ret.CStat, 1, true)).
		Filho(pcn.NodeStr("xMotivo", ret.XMotivo, true)).
		Filho(pcn.NodeStrSemFiltro("chNFGas", chave, true)).
		Filho(pcn.NodeStrSemFiltro("tpEvento", ret.TpEvento.String(), true)).
		Filho(pcn.NodeStr("xEvento", ret.XEvento, true)).
		Filho(pcn.NodeInt("nSeqEvento", ret.NSeqEvento, 1, true))

	if ret.CNPJDest != "" {
		inf.Filho(pcn.NodeStrSemFiltro("CNPJDest", ret.CNPJDest, false))
	}
	if ret.EmailDest != "" {
		inf.Filho(pcn.NodeStr("emailDest", ret.EmailDest, false))
	}
	if ret.COrgaoAutor != 0 {
		inf.Filho(pcn.NodeInt("cOrgaoAutor", ret.COrgaoAutor, 1, false))
	}

	inf.Filho(pcn.NodeStr("dhRegEvento", pcn.FormatarDataHoraXML(ret.DhRegEvento, uf), true)).
		Filho(pcn.NodeStrSemFiltro("nProt", ret.NProt, true))

	elem := pcn.NovoElem("retEventoNFGas").Attr("versao", versao).Filho(inf)
	if r.Signature.Assinada() {
		elem.Filho(gerarSignature(&r.Signature))
	}
	return elem
}
