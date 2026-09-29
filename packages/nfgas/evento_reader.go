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
	"io"
	"os"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitura do XML de evento. Porte de TRetEventoNFGas.LerXml
// (ACBrNFGas.RetEnvEvento.pas).
//
// Raizes aceitas:
//
//	<procEventoNFGas>  envelope com o evento enviado e o retorno
//	<retEventoNFGas>   so o retorno
//	<eventoNFGas>      so o evento enviado (ACRESCIMO, ver abaixo)
//
// DIVERGENCIA: o original envolve a leitura inteira num try/except que
// devolve False, engolindo a causa. Aqui a falha vira error.

// LerEvento le um evento de um io.Reader.
func LerEvento(r io.Reader) (*RetEventoNFGas, error) {
	doc, err := pcn.Parse(r)
	if err != nil {
		return nil, err
	}
	return lerEventoDocumento(doc)
}

// LerEventoString le um evento a partir de uma string.
func LerEventoString(xml string) (*RetEventoNFGas, error) {
	return LerEventoBytes([]byte(xml))
}

// LerEventoBytes le um evento a partir de bytes.
func LerEventoBytes(dados []byte) (*RetEventoNFGas, error) {
	doc, err := pcn.ParseBytes(dados)
	if err != nil {
		return nil, err
	}
	return lerEventoDocumento(doc)
}

// LerEventoArquivo le um evento de um arquivo.
func LerEventoArquivo(caminho string) (*RetEventoNFGas, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfgas: abrir %q: %w", caminho, err)
	}
	ev, err := LerEventoBytes(dados)
	if err != nil {
		return nil, err
	}
	ev.RetInfEvento.NomeArquivo = caminho
	return ev, nil
}

func lerEventoDocumento(doc *pcn.Document) (*RetEventoNFGas, error) {
	if doc == nil || doc.Root == nil {
		return nil, ErrXMLVazio
	}
	root := doc.Root

	ev := &RetEventoNFGas{
		Versao: root.Attr("versao"),
		XML:    doc.XML(),
	}

	var signatureNode *pcn.Node

	switch {
	case strings.EqualFold(root.Nome, "procEventoNFGas"):
		// ACRESCIMO: o ACBr ignora a parte enviada. Sem ela nao ha xJust.
		if evNode := root.FindAnyNs("eventoNFGas"); evNode != nil {
			lerEventoEnviado(evNode, &ev.Evento)
			ev.TemEvento = true
		}

		if retNode := root.FindAnyNs("retEventoNFGas"); retNode != nil {
			if ev.Versao == "" {
				ev.Versao = retNode.Attr("versao")
			}
			signatureNode = retNode.FindAnyNs("Signature")
			lerRetInfEvento(retNode.FindAnyNs("infEvento"), &ev.RetInfEvento)
		}

	case strings.EqualFold(root.Nome, "retEventoNFGas"):
		signatureNode = root.FindAnyNs("Signature")
		lerRetInfEvento(root.FindAnyNs("infEvento"), &ev.RetInfEvento)

	case strings.EqualFold(root.Nome, "eventoNFGas"):
		// ACRESCIMO: o ACBr nao le o evento avulso aqui.
		lerEventoEnviado(root, &ev.Evento)
		ev.TemEvento = true

	default:
		return nil, ErrXMLIncorreto
	}

	if signatureNode != nil {
		pcn.LerSignature(signatureNode, &ev.Signature)
	}

	return ev, nil
}

// lerRetInfEvento le o retorno do evento. Porte de Ler_RetEvento.
//
// ACRESCIMO: CNPJDest, emailDest e cOrgaoAutor estao declarados em
// TRetInfEvento e nao sao lidos pelo ACBr. Aqui sao lidos -- sao dados do
// documento e nada se perde em te-los.
func lerRetInfEvento(node *pcn.Node, r *RetInfEvento) {
	if node == nil || r == nil {
		return
	}

	r.XML = node.OuterXML()
	r.ID = node.Attr("Id")
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))
	r.VerAplic = pcn.ConteudoStr(node.FindAnyNs("verAplic"))
	r.COrgao = pcn.ConteudoInt(node.FindAnyNs("cOrgao"))
	r.CStat = pcn.ConteudoInt(node.FindAnyNs("cStat"))
	r.XMotivo = pcn.ConteudoStr(node.FindAnyNs("xMotivo"))
	r.ChNFGas = pcn.ConteudoStr(node.FindAnyNs("chNFGas"))
	r.TpEvento, _ = ParseTipoEvento(pcn.ConteudoStr(node.FindAnyNs("tpEvento")))
	r.XEvento = pcn.ConteudoStr(node.FindAnyNs("xEvento"))
	r.NSeqEvento = pcn.ConteudoInt(node.FindAnyNs("nSeqEvento"))
	r.DhRegEvento = pcn.ConteudoDataDef(node.FindAnyNs("dhRegEvento"))
	r.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))

	// Acrescimos ao porte.
	r.CNPJDest = pcn.ConteudoStr(node.FindAnyNs("CNPJDest"))
	r.EmailDest = pcn.ConteudoStr(node.FindAnyNs("emailDest"))
	r.COrgaoAutor = pcn.ConteudoInt(node.FindAnyNs("cOrgaoAutor"))
}

// lerEventoEnviado le o evento enviado a SEFAZ.
//
// ACRESCIMO ao porte: o ACBr gera esta parte (Gerar_InfEvento em
// ACBrNFGas.EnvEvento.pas) mas nao a le de volta. Os nomes das tags saem
// dali.
func lerEventoEnviado(node *pcn.Node, e *EventoNFGas) {
	if node == nil || e == nil {
		return
	}
	e.Versao = node.Attr("versao")

	inf := node.FindAnyNs("infEvento")
	if inf == nil {
		return
	}

	e.InfEvento.ID = inf.Attr("Id")
	e.InfEvento.COrgao = pcn.ConteudoInt(inf.FindAnyNs("cOrgao"))
	e.InfEvento.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
	e.InfEvento.CNPJ = pcn.ConteudoStr(inf.FindAnyNs("CNPJ"))
	e.InfEvento.ChNFGas = pcn.ConteudoStr(inf.FindAnyNs("chNFGas"))
	e.InfEvento.DhEvento = pcn.ConteudoDataDef(inf.FindAnyNs("dhEvento"))
	e.InfEvento.TpEvento, _ = ParseTipoEvento(pcn.ConteudoStr(inf.FindAnyNs("tpEvento")))
	e.InfEvento.NSeqEvento = pcn.ConteudoInt(inf.FindAnyNs("nSeqEvento"))

	if det := inf.FindAnyNs("detEvento"); det != nil {
		e.InfEvento.DetEvento.Versao = det.Attr("versao")
		e.InfEvento.DetEvento.DescEvento = pcn.ConteudoStr(det.FindAnyNs("descEvento"))
		e.InfEvento.DetEvento.NProt = pcn.ConteudoStr(det.FindAnyNs("nProt"))
		e.InfEvento.DetEvento.XJust = pcn.ConteudoStr(det.FindAnyNs("xJust"))
		e.InfEvento.DetEvento.IDPedidoCancelado = pcn.ConteudoStr(det.FindAnyNs("idPedidoCancelado"))
	}

	pcn.LerSignature(node.FindAnyNs("Signature"), &e.Signature)
}
