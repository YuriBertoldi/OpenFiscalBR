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
	"fmt"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Administracao do CSC da NFC-e. Porte de TAdmCSCNFCe (ACBrNFe.AdmCSC.pas)
// e TRetAdmCSCNFCe (ACBrNFe.RetAdmCSC.pas).
//
// O CSC -- Codigo de Seguranca do Contribuinte -- e o segredo que fecha o
// hash do QR-Code das versoes 1 e 2. Este servico consulta, pede um novo ou
// revoga um existente, e e o unico da NFC-e que opera por RAIZ DE CNPJ: um
// CSC vale para todos os estabelecimentos do mesmo grupo na UF.

// PedidoAdmCSC e a mensagem admCscNFCe.
type PedidoAdmCSC struct {
	// IndOp diz qual das tres operacoes se quer.
	IndOp IndOperacao

	// RaizCNPJ sao os OITO primeiros digitos do CNPJ -- o grupo, nao o
	// estabelecimento.
	RaizCNPJ string

	// IDCSC e CodigoCSC so vao no XML na REVOGACAO; nas outras duas
	// operacoes o grupo dadosCsc nem e emitido
	// (ACBrNFe.AdmCSC.pas:88-95).
	IDCSC     int
	CodigoCSC string
}

// DadosCSC e um par identificador/codigo devolvido pelo servico.
type DadosCSC struct {
	IDCSC     int    // idCsc
	CodigoCSC string // codigoCsc
}

// RetAdmCSCNFCe e o retorno da administracao do CSC.
type RetAdmCSCNFCe struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	IndOp    IndOperacao      // indOp
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	DadosCSC []DadosCSC       // dadosCsc
}

// Faixa de cStat que indica operacao bem-sucedida, por operacao
// (ACBrNFeWebServices.pas:3916-3925):
//
//	150  consulta devolveu os CSC ativos
//	151  consulta nao achou CSC ativo
//	152  requisicao devolveu um CSC novo
//	153  revogacao confirmada
//
// Repare que o 151 -- "nao ha CSC ativo" -- CONTA como sucesso: a consulta
// funcionou, a resposta e que nao ha nada. Tratar esse codigo como falha e
// o erro natural de quem porta olhando so a descricao.
const (
	cStatCSCMinimo = 150
	cStatCSCMaximo = 153
)

// Atendido informa se o servico processou o pedido.
func (r *RetAdmCSCNFCe) Atendido() bool {
	return r != nil && r.CStat >= cStatCSCMinimo && r.CStat <= cStatCSCMaximo
}

// GerarXMLAdmCSC monta o admCscNFCe.
//
// Porte de TAdmCSCNFCe.GerarXML. O grupo dadosCsc so sai na REVOGACAO, e o
// idCsc vai com seis posicoes zero-preenchidas (FormatFloat '000000').
func GerarXMLAdmCSC(p PedidoAdmCSC, ambiente pcn.TipoAmbiente, versao string) string {
	raiz := pcn.NovoElem("admCscNFCe").
		Attr("xmlns", namespaceNFe).
		Attr("versao", versao).
		Filho(pcn.NovoElem("tpAmb").Texto(ambiente.String())).
		Filho(pcn.NovoElem("indOp").Texto(p.IndOp.String())).
		Filho(pcn.NovoElem("raizCNPJ").Texto(p.RaizCNPJ))

	if p.IndOp == IoRevogaCSC {
		raiz.Filho(pcn.NovoElem("dadosCsc").
			Filho(pcn.NovoElem("idCsc").Texto(fmt.Sprintf("%06d", p.IDCSC))).
			Filho(pcn.NovoElem("codigoCsc").Texto(p.CodigoCSC)))
	}

	return raiz.XML()
}

// LerRetAdmCSCNFCe le o XML de retAdmCscNFCe.
func LerRetAdmCSCNFCe(xml string) (*RetAdmCSCNFCe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetAdmCSCNFCe{
		Versao:  node.Attr("versao"),
		CStat:   pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo: pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))
	r.IndOp, _ = ParseIndOperacao(pcn.ConteudoStr(node.FindAnyNs("indOp")))

	for _, d := range node.FindAllAnyNs("dadosCsc") {
		r.DadosCSC = append(r.DadosCSC, DadosCSC{
			IDCSC:     pcn.ConteudoInt(d.FindAnyNs("idCsc")),
			CodigoCSC: pcn.ConteudoStr(d.FindAnyNs("codigoCsc")),
		})
	}

	return r, nil
}
