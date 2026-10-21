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

// RetConsSitNFe e o retorno da consulta da situacao da NF-e.
// Porte de TRetConsSitNFe (ACBrNFe.RetConsSit.pas).
//
// O documento tem tres blocos independentes, e nenhum deles e obrigatorio:
//
//	protNFe       o protocolo de autorizacao, quando a nota foi autorizada
//	retCancNFe    o cancelamento pelo metodo ANTIGO, anterior ao evento
//	procEventoNFe os eventos registrados contra a chave
//
// O cStat do grupo externo diz o resultado da CONSULTA; quem diz a situacao
// da nota e o cStat do protNFe.
type RetConsSitNFe struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	CUF      int              // cUF
	NRec     string           // nRec
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	DhRecbto time.Time        // dhRecbto
	ChNFe    string           // chNFe

	ProtNFe    *ProtNFe    // protNFe/infProt
	RetCancNFe *RetCancNFe // retCancNFe/infCanc

	// ProcEventos traz os eventos COMO XML CRU, um por entrada.
	//
	// A leitura estruturada depende do leitor de eventos, que entra na
	// etapa propria. Devolver o XML permite ao consumidor guardar o
	// documento -- que e o que a maioria faz com evento vindo da consulta
	// -- sem que este package finja ter um leitor que nao tem.
	ProcEventos []string

	// XMLProtNFe e o protNFe COMO VEIO, sem passar pelo parse.
	//
	// E dele que sai o nfeProc: o ACBr monta o documento processado colando
	// o XML original da nota com este trecho (ACBrNFeWebServices.pas:2465).
	// Os campos ja lidos nao reconstroem byte a byte o elemento que a SEFAZ
	// assinou, entao sem o XML cru nao ha como arquivar a nota processada.
	XMLProtNFe string
}

// CStatProcessado diz se o codigo indica nota PROCESSADA com sucesso.
// Porte de TACBrNFe.CstatProcessado (ACBrNFe.pas:295).
//
// O 120 e autorizacao por DPEC, e e facil esquecer dele.
func CStatProcessado(cStat int) bool {
	switch cStat {
	case 100, 120, 150:
		return true
	}
	return false
}

// CStatCancelada diz se o codigo indica nota CANCELADA.
// Porte de TACBrNFe.CstatCancelada (ACBrNFe.pas:304).
func CStatCancelada(cStat int) bool {
	switch cStat {
	case 101, 135, 151, 155:
		return true
	}
	return false
}

// ProtNFe e o protocolo de autorizacao devolvido na consulta.
type ProtNFe struct {
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
}

// RetCancNFe e o cancelamento pelo metodo antigo, anterior ao evento de
// cancelamento. Permanece no leiaute para notas canceladas antes da
// mudanca.
type RetCancNFe struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	CUF      int              // cUF
	ChNFe    string           // chNFe
	DhRecbto time.Time        // dhRecbto
	NProt    string           // nProt
}

// Processada diz se a CONSULTA achou a nota -- autorizada ou cancelada.
//
// Porte do resultado de TNFeConsulta.TratarResposta
// (ACBrNFeWebServices.pas:2397): o criterio e o cStat EXTERNO, o da
// consulta, nao o do protocolo.
func (r *RetConsSitNFe) Processada() bool {
	return r != nil && (CStatProcessado(r.CStat) || CStatCancelada(r.CStat))
}

// Autorizada informa se a nota tem protocolo de AUTORIZACAO.
//
// Olha o cStat do protNFe, nao o da consulta, porque uma nota cancelada
// volta com cStat externo de cancelamento e ainda assim traz o protNFe da
// autorizacao original -- o proprio ACBr trata esse caso
// (ACBrNFeWebServices.pas:2417-2420).
func (r *RetConsSitNFe) Autorizada() bool {
	return r != nil && r.ProtNFe != nil && CStatProcessado(r.ProtNFe.CStat)
}

// Cancelada informa se a nota foi cancelada pelo metodo antigo.
//
// NAO cobre o cancelamento por EVENTO, que e o caminho atual: para esse, o
// consumidor precisa olhar ProcEventos.
func (r *RetConsSitNFe) Cancelada() bool {
	return r != nil && r.RetCancNFe != nil && CStatCancelada(r.RetCancNFe.CStat)
}

// LerRetConsSitNFe le o XML de retConsSitNFe.
func LerRetConsSitNFe(xml string) (*RetConsSitNFe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetConsSitNFe{
		Versao:   node.Attr("versao"),
		VerAplic: pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		CUF:      pcn.ConteudoInt(node.FindAnyNs("cUF")),
		NRec:     pcn.ConteudoStr(node.FindAnyNs("nRec")),
		CStat:    pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:  pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
		DhRecbto: pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto")),
		ChNFe:    pcn.ConteudoStr(node.FindAnyNs("chNFe")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))

	// Os dois grupos abaixo so sao lidos quando o cStat da CONSULTA os
	// admite (RetConsSit.pas:196 e :226). Ler incondicionalmente faria uma
	// rejeicao que viesse acompanhada de protNFe -- caso que a SEFAZ-PR ja
	// produziu -- ser tratada como autorizacao.
	if cStatAdmiteProtocolo(r.CStat) {
		if prot := node.FindAnyNs("protNFe"); prot != nil {
			if inf := prot.FindAnyNs("infProt"); inf != nil {
				p := &ProtNFe{
					VerAplic: pcn.ConteudoStr(inf.FindAnyNs("verAplic")),
					ChNFe:    pcn.ConteudoStr(inf.FindAnyNs("chNFe")),
					DhRecbto: pcn.ConteudoDataDef(inf.FindAnyNs("dhRecbto")),
					NProt:    pcn.ConteudoStr(inf.FindAnyNs("nProt")),
					DigVal:   pcn.ConteudoStr(inf.FindAnyNs("digVal")),
					CStat:    pcn.ConteudoInt(inf.FindAnyNs("cStat")),
					XMotivo:  pcn.ConteudoStr(inf.FindAnyNs("xMotivo")),
					CMsg:     pcn.ConteudoInt(inf.FindAnyNs("cMsg")),
					XMsg:     pcn.ConteudoStr(inf.FindAnyNs("xMsg")),
				}
				p.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
				r.ProtNFe = p
			}
			r.XMLProtNFe = prot.OuterXML()
		}
	}

	if cStatAdmiteCancelamento(r.CStat) {
		if canc := node.FindAnyNs("retCancNFe"); canc != nil {
			if inf := canc.FindAnyNs("infCanc"); inf != nil {
				c := &RetCancNFe{
					Versao:   canc.Attr("versao"),
					VerAplic: pcn.ConteudoStr(inf.FindAnyNs("verAplic")),
					CStat:    pcn.ConteudoInt(inf.FindAnyNs("cStat")),
					XMotivo:  pcn.ConteudoStr(inf.FindAnyNs("xMotivo")),
					CUF:      pcn.ConteudoInt(inf.FindAnyNs("cUF")),
					ChNFe:    pcn.ConteudoStr(inf.FindAnyNs("chNFe")),
					DhRecbto: pcn.ConteudoDataDef(inf.FindAnyNs("dhRecbto")),
					NProt:    pcn.ConteudoStr(inf.FindAnyNs("nProt")),
				}
				c.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
				r.RetCancNFe = c
			}
		}
	}

	for _, ev := range node.FindAllAnyNs("procEventoNFe") {
		r.ProcEventos = append(r.ProcEventos, ev.OuterXML())
	}

	return r, nil
}

// cStatAdmiteProtocolo e a lista de RetConsSit.pas:196.
func cStatAdmiteProtocolo(cStat int) bool {
	switch cStat {
	case 100, 101, 104, 150, 151, 155:
		return true
	}
	return false
}

// cStatAdmiteCancelamento e a lista de RetConsSit.pas:226. E mais estreita
// que a anterior: so os tres codigos de cancelamento.
func cStatAdmiteCancelamento(cStat int) bool {
	switch cStat {
	case 101, 151, 155:
		return true
	}
	return false
}
