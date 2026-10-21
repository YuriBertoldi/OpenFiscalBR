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

// RetInutNFe e o retorno do pedido de inutilizacao.
// Porte de TRetInutNFe (ACBrNFe.RetInut.pas).
type RetInutNFe struct {
	Versao   string           // atributo versao
	ID       string           // atributo Id de infInut
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	XJust    string           // xJust
	CUF      int              // cUF
	Ano      int              // ano
	CNPJ     string           // CNPJ ou CPF
	Modelo   int              // mod
	Serie    int              // serie
	NNFIni   int              // nNFIni
	NNFFin   int              // nNFFin
	DhRecbto time.Time        // dhRecbto
	NProt    string           // nProt

	// XMLProcInutNFe e o documento de inutilizacao pronto para arquivo,
	// preenchido pelo Inutilizar quando o cStat o admite. Nao vem do XML
	// de retorno: e montado colando o pedido assinado com este retorno.
	XMLProcInutNFe string
}

// cStatInutilizada e o 102, "Inutilizacao de numero homologado".
const cStatInutilizada = 102

// cStatInutilizacaoDuplicada e o 563, "Ja existe pedido de inutilizacao com
// a mesma faixa". O ACBr NAO o considera sucesso, mas gera o procInutNFe
// mesmo assim (ACBrNFeWebServices.pas:2778) -- o numero ja esta inutilizado,
// logo ha documento a arquivar.
const cStatInutilizacaoDuplicada = 563

// Homologada informa se a inutilizacao foi aceita.
//
// Porte do resultado de TNFeInutilizacao.TratarResposta: SO o 102. O 563
// nao entra, apesar de gerar o procInutNFe.
func (r *RetInutNFe) Homologada() bool {
	return r != nil && r.CStat == cStatInutilizada
}

// TemProcInutNFe informa se cabe montar o documento de inutilizacao.
//
// Porte da condicao de :2777: dois codigos, nao um. O 563 diz que a faixa
// ja estava inutilizada -- o pedido falhou, mas a inutilizacao existe.
func (r *RetInutNFe) TemProcInutNFe() bool {
	return r != nil &&
		(r.CStat == cStatInutilizada || r.CStat == cStatInutilizacaoDuplicada)
}

// LerRetInutNFe le o XML de retInutNFe.
//
// Aceita as DUAS formas que o ACBr aceita (ACBrNFe.RetInut.pas:124-133): a
// raiz sendo o proprio retInutNFe, ou um envelope que o contenha -- que e o
// caso do procInutNFe arquivado.
func LerRetInutNFe(xml string) (*RetInutNFe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetInutNFe{Versao: node.Attr("versao")}

	inf := node.FindAnyNs("infInut")
	if ret := node.FindAnyNs("retInutNFe"); ret != nil {
		inf = ret.FindAnyNs("infInut")
	}
	if inf == nil {
		return r, nil
	}

	r.ID = inf.Attr("Id")
	r.VerAplic = pcn.ConteudoStr(inf.FindAnyNs("verAplic"))
	r.CStat = pcn.ConteudoInt(inf.FindAnyNs("cStat"))
	r.XMotivo = pcn.ConteudoStr(inf.FindAnyNs("xMotivo"))
	r.CUF = pcn.ConteudoInt(inf.FindAnyNs("cUF"))
	r.Ano = pcn.ConteudoInt(inf.FindAnyNs("ano"))
	r.CNPJ = pcn.ConteudoCNPJCPF(inf)
	r.Modelo = pcn.ConteudoInt(inf.FindAnyNs("mod"))
	r.Serie = pcn.ConteudoInt(inf.FindAnyNs("serie"))
	r.NNFIni = pcn.ConteudoInt(inf.FindAnyNs("nNFIni"))
	r.NNFFin = pcn.ConteudoInt(inf.FindAnyNs("nNFFin"))
	r.DhRecbto = pcn.ConteudoDataDef(inf.FindAnyNs("dhRecbto"))
	r.NProt = pcn.ConteudoStr(inf.FindAnyNs("nProt"))
	r.XJust = pcn.ConteudoStr(inf.FindAnyNs("xJust"))
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))

	return r, nil
}

// GerarXMLProcInutNFe monta o documento de inutilizacao para arquivo.
//
// Porte de TNFeInutilizacao.TratarResposta (ACBrNFeWebServices.pas:2779):
// o envelope cola o PEDIDO assinado com o RETORNO, os dois como vieram.
// Nenhum dos dois e regerado -- o pedido traz assinatura, e regerar
// invalidaria o digest.
//
// Repare na caixa do elemento: "ProcInutNFe" com P maiusculo, que e o que
// o ACBr emite. O XSD da SEFAZ nomeia "ProcInutNFe" tambem.
//
// Duas diferencas deliberadas em relacao ao original:
//
//   - o prologo <?xml?> NAO entra. O ACBr o acrescenta aqui porque grava o
//     resultado em disco na mesma linha (:2784); nesta biblioteca gravar e
//     do consumidor, e o GerarXML da nota tambem devolve sem prologo;
//   - os dois pedacos passam por RemoverDeclaracaoXML. No ACBr eles nunca
//     tem prologo, entao e defesa, nao mudanca de comportamento -- mas o
//     consumidor desta funcao pode estar relendo um arquivo de disco.
func GerarXMLProcInutNFe(pedidoAssinado, retorno, versao string) string {
	return `<ProcInutNFe versao="` + versao + `" xmlns="` + namespaceNFe + `">` +
		pcn.RemoverDeclaracaoXML(pedidoAssinado) +
		pcn.RemoverDeclaracaoXML(retorno) +
		`</ProcInutNFe>`
}
