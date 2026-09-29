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

package pcn

import "time"

// ProcDFe carrega os dados do protocolo de autorizacao de um documento
// fiscal eletronico. Porte de TProcDFe (ACBrDFeComum.Proc.pas).
//
// Os campos de caminho de arquivo e os de montagem do XML do original
// (PathDFe, XML_DFe, XML_prot) ficam de fora: pertencem a emissao, que
// ainda nao foi portada.
type ProcDFe struct {
	ID       string       // atributo Id de infProt
	TpAmb    TipoAmbiente // tpAmb
	VerAplic string       // verAplic
	ChDFe    string       // chNFGas / chNFe / chCTe...
	DhRecbto time.Time    // dhRecbto
	NProt    string       // nProt
	DigVal   string       // digVal
	CStat    int          // cStat
	XMotivo  string       // xMotivo
	CMsg     int          // cMsg
	XMsg     string       // xMsg

	// Versao, NameSpace, TagGrupo e TagDFe sao os parametros que o
	// construtor do Delphi recebe e guarda para gerar o XML do processo.
	// Ficam aqui para que a geracao futura nao precise reabrir a struct.
	Versao    string
	NameSpace string
	TagGrupo  string
	TagDFe    string
}

// NovoProcDFe cria um ProcDFe com os metadados de geracao preenchidos.
// Porte de TProcDFe.Create.
func NovoProcDFe(versao, nameSpace, tagGrupo, tagDFe string) *ProcDFe {
	return &ProcDFe{
		Versao:    versao,
		NameSpace: nameSpace,
		TagGrupo:  tagGrupo,
		TagDFe:    tagDFe,
	}
}

// Vazio informa se nenhum dado de protocolo foi lido.
func (p *ProcDFe) Vazio() bool {
	if p == nil {
		return true
	}
	return p.NProt == "" && p.CStat == 0 && p.ChDFe == ""
}

// LerInfProt preenche o protocolo a partir de um elemento que contenha
// infProt -- tipicamente <protNFGas>, <protNFe> ou equivalente.
//
// tagChave e o nome da tag que carrega a chave de acesso, que varia por
// documento ("chNFGas", "chNFe", "chCTe"); ela alimenta ChDFe.
//
// Le exatamente os campos que o ACBr le. Nao sao lidos, por omissao do
// original: infProt/@Id, infProt/infFisco e a assinatura do protocolo.
func LerInfProt(protNode *Node, proc *ProcDFe, tagChave string) {
	if protNode == nil || proc == nil {
		return
	}
	infProt := protNode.FindAnyNs("infProt")
	if infProt == nil {
		return
	}

	proc.TpAmb, _ = ParseTipoAmbiente(ConteudoStr(infProt.FindAnyNs("tpAmb")))
	if tagChave != "" {
		proc.ChDFe = ConteudoStr(infProt.FindAnyNs(tagChave))
	}
	proc.VerAplic = ConteudoStr(infProt.FindAnyNs("verAplic"))
	proc.DhRecbto = ConteudoDataDef(infProt.FindAnyNs("dhRecbto"))
	proc.NProt = ConteudoStr(infProt.FindAnyNs("nProt"))
	proc.DigVal = ConteudoStr(infProt.FindAnyNs("digVal"))
	proc.CStat = ConteudoInt(infProt.FindAnyNs("cStat"))
	proc.XMotivo = ConteudoStr(infProt.FindAnyNs("xMotivo"))
	proc.CMsg = ConteudoInt(infProt.FindAnyNs("cMsg"))
	proc.XMsg = ConteudoStr(infProt.FindAnyNs("xMsg"))
}
