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
	if e == nil {
		return "", ErrXMLVazio
	}
	if e.InfEvento.TpEvento != TeCancelamento {
		return "", fmt.Errorf("nfgas: tipo de evento nao implementado para NFGas: %s",
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

	evento := pcn.NovoElem("eventoNFGas").
		Attr("xmlns", Namespace).
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
	return evento.XML(), nil
}
