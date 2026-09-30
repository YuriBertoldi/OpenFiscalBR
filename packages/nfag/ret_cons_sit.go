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
	"fmt"
	"io"
	"os"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitura do retorno da consulta de situacao.
// Porte de TRetConsSitNFAg (ACBrNFAg.RetConsSit.pas).

// RetConsSitNFAg e o retorno da consulta de situacao de uma NFAg.
type RetConsSitNFAg struct {
	Versao   string           // atributo versao
	VerAplic string           // verAplic
	TpAmb    pcn.TipoAmbiente // tpAmb
	CUF      int              // cUF
	NRec     string           // nRec
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	DhRecbto time.Time        // dhRecbto
	ChNFAg   string           // chNFAg

	// ProtNFAg e o protocolo, lido apenas quando cStat esta em
	// {100, 101, 104, 150, 151, 155}.
	ProtNFAg pcn.ProcDFe
	// XMLProtNFAg e o trecho exato do elemento protNFAg.
	XMLProtNFAg string

	// ProcEventoNFAg sao os eventos vinculados ao documento.
	ProcEventoNFAg []*RetEventoNFAg

	// XML e o XML de origem.
	XML string
}

// Autorizada informa se a consulta indica documento autorizado.
func (r *RetConsSitNFAg) Autorizada() bool {
	return r != nil && cStatConfirmada[r.CStat]
}

// Cancelada informa se a consulta indica documento cancelado.
func (r *RetConsSitNFAg) Cancelada() bool {
	return r != nil && cStatCancelada[r.CStat]
}

// LerRetConsSit le o retorno da consulta de situacao de um io.Reader.
func LerRetConsSit(r io.Reader) (*RetConsSitNFAg, error) {
	doc, err := pcn.Parse(r)
	if err != nil {
		return nil, err
	}
	return lerRetConsSitDocumento(doc)
}

// LerRetConsSitString le o retorno da consulta a partir de uma string.
func LerRetConsSitString(xml string) (*RetConsSitNFAg, error) {
	return LerRetConsSitBytes([]byte(xml))
}

// LerRetConsSitBytes le o retorno da consulta a partir de bytes.
func LerRetConsSitBytes(dados []byte) (*RetConsSitNFAg, error) {
	doc, err := pcn.ParseBytes(dados)
	if err != nil {
		return nil, err
	}
	return lerRetConsSitDocumento(doc)
}

// LerRetConsSitArquivo le o retorno da consulta de um arquivo.
func LerRetConsSitArquivo(caminho string) (*RetConsSitNFAg, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfag: abrir %q: %w", caminho, err)
	}
	return LerRetConsSitBytes(dados)
}

// lerRetConsSitDocumento e o porte de TRetConsSitNFAg.LerXml.
//
// DIVERGENCIA: o original envolve tudo num try/except que devolve False e
// descarta a causa. Aqui a falha de leitura de um evento vinculado nao
// derruba a leitura do retorno -- o evento problematico e simplesmente
// omitido da lista, e o resto continua acessivel.
func lerRetConsSitDocumento(doc *pcn.Document) (*RetConsSitNFAg, error) {
	if doc == nil || doc.Root == nil {
		return nil, ErrXMLVazio
	}
	node := doc.Root

	r := &RetConsSitNFAg{
		Versao:   node.Attr("versao"),
		VerAplic: pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		CUF:      pcn.ConteudoInt(node.FindAnyNs("cUF")),
		NRec:     pcn.ConteudoStr(node.FindAnyNs("nRec")),
		CStat:    pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:  pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
		DhRecbto: pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto")),
		ChNFAg:   pcn.ConteudoStr(node.FindAnyNs("chNFAg")),
		XML:      doc.XML(),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))

	// O protocolo so e lido nestes cStat -- e a condicao do case do
	// original, replicada.
	if cStatComProtocolo[r.CStat] {
		if prot := node.FindAnyNs("protNFAg"); prot != nil {
			r.XMLProtNFAg = prot.OuterXML()
			pcn.LerInfProt(prot, &r.ProtNFAg, "chNFAg")
		}
	}

	for _, ev := range node.FindAllAnyNs("procEventoNFAg") {
		lido, err := LerEventoString(ev.OuterXML())
		if err != nil {
			continue
		}
		r.ProcEventoNFAg = append(r.ProcEventoNFAg, lido)
	}

	return r, nil
}
