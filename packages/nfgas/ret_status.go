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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// RetConsStatServNFGas e o retorno da consulta de status do servico.
// Porte de TRetConsStatServ (ACBrDFeComum.RetConsStatServ.pas).
type RetConsStatServNFGas struct {
	Versao    string           // atributo versao
	TpAmb     pcn.TipoAmbiente // tpAmb
	VerAplic  string           // verAplic
	CStat     int              // cStat
	XMotivo   string           // xMotivo
	CUF       int              // cUF
	DhRecbto  time.Time        // dhRecbto
	TMed      int              // tMed
	DhRetorno time.Time        // dhRetorno
	XObs      string           // xObs
}

// EmOperacao informa se o servico respondeu 107 (em operacao).
func (r *RetConsStatServNFGas) EmOperacao() bool {
	return r != nil && r.CStat == 107
}

// LerRetConsStatServ le o XML de retConsStatServNFGas.
// Porte 1:1 de TRetConsStatServ.LerXml.
func LerRetConsStatServ(xml string) (*RetConsStatServNFGas, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root
	r := &RetConsStatServNFGas{
		Versao:    node.Attr("versao"),
		VerAplic:  pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		CStat:     pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:   pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
		CUF:       pcn.ConteudoInt(node.FindAnyNs("cUF")),
		DhRecbto:  pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto")),
		TMed:      pcn.ConteudoInt(node.FindAnyNs("tMed")),
		DhRetorno: pcn.ConteudoDataDef(node.FindAnyNs("dhRetorno")),
		XObs:      pcn.ConteudoStr(node.FindAnyNs("xObs")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))
	return r, nil
}
