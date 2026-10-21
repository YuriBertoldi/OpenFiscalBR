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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitores dos dois retornos do envio ASSINCRONO.
//
// O fluxo tem duas etapas e cada uma tem o seu documento:
//
//	retEnviNFe      resposta ao envio -- traz o RECIBO, nao o protocolo
//	retConsReciNFe  resposta a consulta do recibo -- traz os protocolos
//
// Porte de TretEnvDFe (ACBrDFeComum.RetEnvio.pas) e de TRetConsReciDFe
// (ACBrDFeComum.RetConsReciDFe.pas). As duas classes sao COMUNS a todos os
// DFe: o que muda entre eles e so o sufixo das tags (protNFe/chNFe contra
// protCTe/chCTe), que o ACBr passa como parametro no construtor.

// InfRec e o grupo infRec do retEnviNFe: o comprovante de que o lote entrou
// na fila. Nao diz nada sobre o resultado das notas.
type InfRec struct {
	NRec     string    // nRec -- o recibo, que alimenta a consulta seguinte
	DhRecbto time.Time // dhRecbto
	TMed     int       // tMed, tempo medio de resposta em segundos
}

// RetEnviNFe e o retorno do envio de lote.
// Porte de TretEnvDFe.
//
// O cStat 103 ("Lote recebido com sucesso") e o unico que indica envio
// aceito -- e e o que o TNFeRecepcao.TratarResposta usa como resultado
// (ACBrNFeWebServices.pas:1317). Qualquer outro codigo significa que o lote
// NAO entrou na fila e nao ha recibo para consultar.
type RetEnviNFe struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	CUF      int              // cUF
	VerAplic string           // verAplic
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	InfRec   InfRec           // infRec
}

// LoteRecebido informa se o lote entrou na fila de processamento.
//
// Porte do resultado de TNFeRecepcao.TratarResposta no ramo ASSINCRONO
// (ACBrNFeWebServices.pas:1317).
func (r *RetEnviNFe) LoteRecebido() bool {
	return r != nil && r.CStat == cStatLoteRecebido
}

// cStatLoteRecebido e o 103 do leiaute: "Lote recebido com sucesso".
const cStatLoteRecebido = 103

// LerRetEnviNFe le o XML de retEnviNFe.
func LerRetEnviNFe(xml string) (*RetEnviNFe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetEnviNFe{
		Versao:   node.Attr("versao"),
		CUF:      pcn.ConteudoInt(node.FindAnyNs("cUF")),
		VerAplic: pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		CStat:    pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:  pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))

	if inf := node.FindAnyNs("infRec"); inf != nil {
		r.InfRec = InfRec{
			NRec:     pcn.ConteudoStr(inf.FindAnyNs("nRec")),
			DhRecbto: pcn.ConteudoDataDef(inf.FindAnyNs("dhRecbto")),
			TMed:     pcn.ConteudoInt(inf.FindAnyNs("tMed")),
		}
	}
	return r, nil
}

// ProtDFe e um protocolo devolvido na consulta do recibo.
//
// Tem os mesmos campos do ProtNFe da consulta de situacao, mais o Id e o
// XMLProtNFe. Nao os compartilha com ele porque os dois leitores sao
// classes distintas no ACBr e evoluem separados -- unificar aqui criaria um
// acoplamento que o original nao tem.
type ProtDFe struct {
	ID       string           // atributo Id de infProt
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	ChNFe    string           // chNFe
	DhRecbto time.Time        // dhRecbto
	NProt    string           // nProt
	DigVal   string           // digVal
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	CMsg     int              // cMsg
	XMsg     string           // xMsg

	// XMLProtNFe e o <protNFe> COMO VEIO. E dele que sai o nfeProc:
	// ACBrNFeWebServices.pas:1551 cola este trecho com o XML assinado da
	// nota.
	XMLProtNFe string
}

// RetConsReciNFe e o retorno da consulta do recibo.
// Porte de TRetConsReciDFe.
type RetConsReciNFe struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	NRec     string           // nRec
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	CUF      int              // cUF
	CMsg     int              // cMsg
	XMsg     string           // xMsg

	// ProtNFe traz um protocolo por nota do lote. A ordem e a do retorno,
	// que NAO e necessariamente a do envio: a correspondencia se faz pela
	// chave (ACBrNFeWebServices.pas:1649).
	ProtNFe []ProtDFe
}

// LoteProcessado informa se a SEFAZ terminou de processar o lote.
//
// 104 e "Lote processado"; 105 e "Lote em processamento", e e o unico
// codigo que manda consultar de novo (ACBrNFeWebServices.pas:1631 e :1498).
func (r *RetConsReciNFe) LoteProcessado() bool {
	return r != nil && r.CStat == cStatLoteProcessado
}

// LoteEmProcessamento informa se vale consultar o recibo outra vez.
func (r *RetConsReciNFe) LoteEmProcessamento() bool {
	return r != nil && r.CStat == cStatLoteEmProcessamento
}

const (
	// cStatLoteProcessado e o 104: a SEFAZ terminou; os protocolos estao
	// no proprio retorno.
	cStatLoteProcessado = 104
	// cStatLoteEmProcessamento e o 105: ainda na fila, consulte de novo.
	cStatLoteEmProcessamento = 105
)

// ProtocoloDaChave devolve o protocolo da nota cuja chave bate, ou nil.
//
// A busca e pela chave e NAO pela posicao porque a SEFAZ nao garante ordem
// -- o proprio ACBr faz um laco duplo comparando chaves
// (ACBrNFeWebServices.pas:1645-1652).
func (r *RetConsReciNFe) ProtocoloDaChave(chave string) *ProtDFe {
	if r == nil {
		return nil
	}
	chave = pcn.RemoverLiteralChave(chave)
	for i := range r.ProtNFe {
		if pcn.RemoverLiteralChave(r.ProtNFe[i].ChNFe) == chave {
			return &r.ProtNFe[i]
		}
	}
	return nil
}

// LerRetConsReciNFe le o XML de retConsReciNFe.
func LerRetConsReciNFe(xml string) (*RetConsReciNFe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetConsReciNFe{
		Versao:   node.Attr("versao"),
		VerAplic: pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		NRec:     pcn.ConteudoStr(node.FindAnyNs("nRec")),
		CStat:    pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:  pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
		CUF:      pcn.ConteudoInt(node.FindAnyNs("cUF")),
		CMsg:     pcn.ConteudoInt(node.FindAnyNs("cMsg")),
		XMsg:     pcn.ConteudoStr(node.FindAnyNs("xMsg")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))

	// Sem guard de cStat, ao contrario do retConsSitNFe: o Ler_ProtDFe le
	// os protocolos sempre que existirem (RetConsReciDFe.pas:205). A
	// diferenca e do ORIGINAL, nao uma omissao daqui -- aqui o cStat
	// externo fala do LOTE, e a situacao de cada nota ja vem no cStat do
	// proprio protocolo.
	for _, prot := range node.FindAllAnyNs("protNFe") {
		p := ProtDFe{XMLProtNFe: prot.OuterXML()}
		if inf := prot.FindAnyNs("infProt"); inf != nil {
			p.ID = inf.Attr("Id")
			p.VerAplic = pcn.ConteudoStr(inf.FindAnyNs("verAplic"))
			p.ChNFe = pcn.ConteudoStr(inf.FindAnyNs("chNFe"))
			p.DhRecbto = pcn.ConteudoDataDef(inf.FindAnyNs("dhRecbto"))
			p.NProt = pcn.ConteudoStr(inf.FindAnyNs("nProt"))
			p.DigVal = pcn.ConteudoStr(inf.FindAnyNs("digVal"))
			p.CStat = pcn.ConteudoInt(inf.FindAnyNs("cStat"))
			p.XMotivo = pcn.ConteudoStr(inf.FindAnyNs("xMotivo"))
			p.CMsg = pcn.ConteudoInt(inf.FindAnyNs("cMsg"))
			p.XMsg = pcn.ConteudoStr(inf.FindAnyNs("xMsg"))
			p.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
		}
		r.ProtNFe = append(r.ProtNFe, p)
	}

	return r, nil
}
