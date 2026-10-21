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

// Consulta do cadastro de contribuintes. Porte de TConsCad
// (ACBrDFeComum.ConsCad.pas) e TRetConsCad (ACBrDFeComum.RetConsCad.pas).
//
// E o unico servico da NF-e que nao fala de documento fiscal nenhum: ele
// pergunta a SEFAZ se uma inscricao estadual existe e esta ativa. Por isso
// a UF nao sai da configuracao do emissor -- ela e o ALVO da consulta.

// InfCad e um cadastro devolvido pela consulta.
//
// A SEFAZ pode devolver MAIS DE UM: um contribuinte com varias inscricoes
// na mesma UF aparece uma vez por inscricao.
type InfCad struct {
	IE         string    // IE
	CNPJ       string    // CNPJ
	CPF        string    // CPF
	UF         string    // UF
	CSit       int       // cSit -- 0 nao habilitado, 1 habilitado
	IndCredNFe int       // indCredNFe
	IndCredCTe int       // indCredCTe
	XNome      string    // xNome
	XFant      string    // xFant
	XRegApur   string    // xRegApur
	CNAE       int       // CNAE
	DIniAtiv   time.Time // dIniAtiv
	DUltSit    time.Time // dUltSit
	DBaixa     time.Time // dBaixa
	IEUnica    string    // IEUnica
	IEAtual    string    // IEAtual

	// Endereco e o grupo ender, ACHATADO no item como no ACBr: o
	// TInfCadCollectionItem guarda xLgr..CEP direto, sem struct intermedia.
	XLgr    string // ender/xLgr
	Nro     string // ender/nro
	XCpl    string // ender/xCpl
	XBairro string // ender/xBairro
	CMun    int    // ender/cMun
	XMun    string // ender/xMun
	CEP     string // ender/CEP
}

// Habilitado informa se a inscricao esta ativa. cSit 1 e habilitado.
func (i InfCad) Habilitado() bool { return i.CSit == 1 }

// RetConsCad e o retorno da consulta de cadastro.
type RetConsCad struct {
	Versao   string    // atributo versao
	VerAplic string    // infCons/verAplic
	CStat    int       // infCons/cStat
	XMotivo  string    // infCons/xMotivo
	UF       string    // infCons/UF
	IE       string    // infCons/IE
	CNPJ     string    // infCons/CNPJ
	CPF      string    // infCons/CPF
	DhCons   time.Time // infCons/dhCons
	CUF      int       // infCons/cUF

	InfCad []InfCad // infCons/infCad
}

const (
	// cStatCadastroUmaOcorrencia e o 111: "Consulta cadastro com uma
	// ocorrencia".
	cStatCadastroUmaOcorrencia = 111
	// cStatCadastroVariasOcorrencias e o 112: mais de uma inscricao.
	cStatCadastroVariasOcorrencias = 112
)

// Encontrado informa se a consulta achou cadastro.
//
// Porte do resultado de TNFeConsultaCadastro.TratarResposta
// (ACBrNFeWebServices.pas:2991): dois codigos, e so eles. Um cadastro
// INEXISTENTE volta com outro cStat e lista vazia.
func (r *RetConsCad) Encontrado() bool {
	return r != nil &&
		(r.CStat == cStatCadastroUmaOcorrencia ||
			r.CStat == cStatCadastroVariasOcorrencias)
}

// GerarXMLConsCad monta o ConsCad.
//
// Porte de TConsCad.GerarXML. A precedencia dos documentos e IE, depois
// CNPJ, depois CPF -- e EXCLUSIVA: so um deles sai no XML. O componente
// reforca isso nos setters (SetCNPJ/SetCPF/SetIE, :2865-2896), que limpam
// os outros dois; aqui a escolha acontece na geracao, que e o lugar onde
// ela de fato importa.
func GerarXMLConsCad(uf, ie, cnpj, cpf, versao string) string {
	documento := pcn.NovoElem("CPF").Texto(cpf)
	switch {
	case ie != "":
		documento = pcn.NovoElem("IE").Texto(ie)
	case cnpj != "":
		documento = pcn.NovoElem("CNPJ").Texto(cnpj)
	}

	return pcn.NovoElem("ConsCad").
		Attr("xmlns", namespaceNFe).
		Attr("versao", versao).
		Filho(pcn.NovoElem("infCons").
			Filho(pcn.NovoElem("xServ").Texto("CONS-CAD")).
			Filho(pcn.NovoElem("UF").Texto(uf)).
			Filho(documento)).
		XML()
}

// LerRetConsCad le o XML de retConsCad.
func LerRetConsCad(xml string) (*RetConsCad, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetConsCad{Versao: node.Attr("versao")}

	inf := node.FindAnyNs("infCons")
	if inf == nil {
		return r, nil
	}

	// O preenchimento com zeros a esquerda nao e cosmetico: ha web service
	// que devolve o CNPJ sem eles, e o consumidor que comparar o retorno
	// com o cadastro proprio erraria (RetConsCad.pas:210-219).
	r.CNPJ = preencherDocumento(pcn.ConteudoStr(inf.FindAnyNs("CNPJ")), 14)
	r.CPF = preencherDocumento(pcn.ConteudoStr(inf.FindAnyNs("CPF")), 11)
	r.VerAplic = pcn.ConteudoStr(inf.FindAnyNs("verAplic"))
	r.CStat = pcn.ConteudoInt(inf.FindAnyNs("cStat"))
	r.XMotivo = pcn.ConteudoStr(inf.FindAnyNs("xMotivo"))
	r.UF = pcn.ConteudoStr(inf.FindAnyNs("UF"))
	r.IE = pcn.ConteudoStr(inf.FindAnyNs("IE"))
	r.DhCons = pcn.ConteudoDataDef(inf.FindAnyNs("dhCons"))
	r.CUF = pcn.ConteudoInt(inf.FindAnyNs("cUF"))

	for _, cad := range inf.FindAllAnyNs("infCad") {
		item := InfCad{
			CNPJ:       preencherDocumento(pcn.ConteudoStr(cad.FindAnyNs("CNPJ")), 14),
			CPF:        preencherDocumento(pcn.ConteudoStr(cad.FindAnyNs("CPF")), 11),
			IE:         pcn.ConteudoStr(cad.FindAnyNs("IE")),
			UF:         pcn.ConteudoStr(cad.FindAnyNs("UF")),
			CSit:       pcn.ConteudoInt(cad.FindAnyNs("cSit")),
			IndCredNFe: pcn.ConteudoInt(cad.FindAnyNs("indCredNFe")),
			IndCredCTe: pcn.ConteudoInt(cad.FindAnyNs("indCredCTe")),
			XNome:      pcn.ConteudoStr(cad.FindAnyNs("xNome")),
			XFant:      pcn.ConteudoStr(cad.FindAnyNs("xFant")),
			XRegApur:   pcn.ConteudoStr(cad.FindAnyNs("xRegApur")),
			CNAE:       pcn.ConteudoInt(cad.FindAnyNs("CNAE")),
			DIniAtiv:   pcn.ConteudoDataDef(cad.FindAnyNs("dIniAtiv")),
			DUltSit:    pcn.ConteudoDataDef(cad.FindAnyNs("dUltSit")),
			DBaixa:     pcn.ConteudoDataDef(cad.FindAnyNs("dBaixa")),
			IEUnica:    pcn.ConteudoStr(cad.FindAnyNs("IEUnica")),
			IEAtual:    pcn.ConteudoStr(cad.FindAnyNs("IEAtual")),
		}
		if ender := cad.FindAnyNs("ender"); ender != nil {
			item.XLgr = pcn.ConteudoStr(ender.FindAnyNs("xLgr"))
			item.Nro = pcn.ConteudoStr(ender.FindAnyNs("nro"))
			item.XCpl = pcn.ConteudoStr(ender.FindAnyNs("xCpl"))
			item.XBairro = pcn.ConteudoStr(ender.FindAnyNs("xBairro"))
			item.CMun = pcn.ConteudoInt(ender.FindAnyNs("cMun"))
			item.XMun = pcn.ConteudoStr(ender.FindAnyNs("xMun"))
			item.CEP = pcn.ConteudoStr(ender.FindAnyNs("CEP"))
		}
		r.InfCad = append(r.InfCad, item)
	}

	return r, nil
}

// preencherDocumento completa o documento com zeros a esquerda, e so
// quando ele veio preenchido e curto -- vazio continua vazio.
func preencherDocumento(documento string, tamanho int) string {
	if documento == "" || len(documento) >= tamanho {
		return documento
	}
	return pcn.PreencherZerosEsquerda(documento, tamanho)
}
