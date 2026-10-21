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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// RetConsStatServ e o retorno da consulta de status do servico.
// Porte de TRetConsStatServ (ACBrDFeComum.RetConsStatServ.pas).
type RetConsStatServ struct {
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

// EmOperacao informa se o servico respondeu 107, o unico cStat que o ACBr
// trata como sucesso (ACBrNFeWebServices.pas:905).
func (r *RetConsStatServ) EmOperacao() bool {
	return r != nil && r.CStat == 107
}

// LerRetConsStatServ le o XML de retConsStatServ.
func LerRetConsStatServ(xml string) (*RetConsStatServ, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetConsStatServ{
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

// AjustarHoraDoSVRS corrige o dhRecbto que o ambiente do Rio Grande do Sul
// devolve adiantado em uma hora.
//
// Porte do remendo de ACBrNFeWebServices.pas:885-893. O comentario do ACBr
// explica: "WebService do RS retorna horario de verao mesmo pros estados
// que nao adotam esse horario, ao utilizar esta hora para basear a emissao
// da nota acontece o erro".
//
// Tres condicoes, todas necessarias: a URL ser do SVRS, a diferenca para o
// instante de referencia passar de cinquenta minutos, e a UF nao estar em
// horario de verao.
//
// A terceira e hoje SEMPRE verdadeira: o horario de verao foi extinto pelo
// Decreto 9.772/2019, e o porte nao tem IsHorarioDeVerao justamente por
// isso. Fica declarada no nome da funcao e neste comentario, para o dia em
// que voltar.
//
// `agora` entra por parametro, como nas regras de negocio: uma correcao que
// depende do relogio nao seria testavel.
func AjustarHoraDoSVRS(dhRecbto time.Time, url string, agora time.Time) time.Time {
	if dhRecbto.IsZero() || !strings.Contains(url, "svrs.rs.gov.br") {
		return dhRecbto
	}
	// MinutesBetween do Delphi TRUNCA para minutos inteiros, entao o ajuste
	// so entra a partir de 51 minutos cheios. Sem o Truncate, a janela entre
	// 50 e 51 minutos divergiria do original.
	if agora.Sub(dhRecbto).Abs().Truncate(time.Minute) <= 50*time.Minute {
		return dhRecbto
	}
	return dhRecbto.Add(-time.Hour)
}
