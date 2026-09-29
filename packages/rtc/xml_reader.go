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

package rtc

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitores dos grupos da Reforma Tributaria.
// Porte de TDFeRTCXmlReader (ACBrDFe.RTC.XmlReader.pas), metodo a metodo.
//
// Regras herdadas do original, preservadas aqui de proposito:
//
//   - node nil sai sem fazer nada, deixando a struct no valor anterior
//     ("if not Assigned(ANode) then Exit");
//   - a busca e sempre Find (nome qualificado), NUNCA FindAnyNs -- a unica
//     excecao e Ler_gPagAntecipadoProd, que usa FindAnyNs;
//   - a precisao de cada campo (De2 x De4) e a declarada no .pas, e varia
//     para o MESMO nome de tag em grupos diferentes.
//
// Divergencias deliberadas estao marcadas com "DIVERGENCIA" no comentario.

// LerGCompraGovReduzido le o grupo gCompraGov dos DFe que usam a forma
// reduzida (BPe, CTe, NF3e, NFAg, NFCom, NFGas).
// Porte de Ler_gCompraGovReduzido.
func LerGCompraGovReduzido(node *pcn.Node, g *GCompraGovReduzido) {
	if node == nil || g == nil {
		return
	}

	g.TpEnteGov, _ = ParseTpEnteGov(pcn.ConteudoStr(node.Find("tpEnteGov")))
	g.PRedutor = pcn.ConteudoDe4(node.Find("pRedutor"))
	g.TpOperGov, _ = ParseTpOperGov(pcn.ConteudoStr(node.Find("tpOperGov")))

	g.RefDFe = nil
	for _, n := range node.FindAll("refDFeAnt") {
		g.RefDFe = append(g.RefDFe, RefDFeAnt{RefDFeAnt: pcn.ConteudoStr(n)})
	}
}

// LerGPagAntecipadoProd le o grupo gPagAntecipado do item.
// Porte de Ler_gPagAntecipadoProd.
//
// Unico leitor da RTC que usa FindAnyNs em vez de Find.
func LerGPagAntecipadoProd(node *pcn.Node, g *GPagAntecipadoProd) {
	if node == nil || g == nil {
		return
	}
	g.ChDFePagAnt = pcn.ConteudoStr(node.FindAnyNs("chDFePagAnt"))
	g.NItemPagAnt = pcn.ConteudoInt(node.FindAnyNs("nItemPagAnt"))
}

// LerGPagAntecipadoIde le o grupo gPagAntecipado da ide, onde as chaves vem
// em elementos chDFePagAnt. Porte de Ler_gPagAntecipadoIde.
//
// DIVERGENCIA: o original nao limpa a lista antes do laco e ainda indexa
// refNFe[i] com o indice do laco, o que corrompe itens preexistentes numa
// releitura. Aqui a lista e reinicializada e os itens sao acrescentados na
// ordem -- o resultado e igual numa leitura limpa, que e o unico caso em
// que o comportamento do original e definido.
func LerGPagAntecipadoIde(node *pcn.Node, g *GPagAntecipado) {
	if node == nil || g == nil {
		return
	}
	g.RefNFe = nil
	for _, n := range node.FindAll("chDFePagAnt") {
		g.RefNFe = append(g.RefNFe, RefDFePagAnt{RefDFeChave: pcn.ConteudoStr(n)})
	}
}

// LerIBSCBS le o grupo IBSCBS do item. Porte de Ler_IBSCBS.
func LerIBSCBS(node *pcn.Node, v *IBSCBS) {
	if node == nil || v == nil {
		return
	}

	v.CST, _ = ParseCSTIBSCBS(pcn.ConteudoStr(node.Find("CST")))
	v.CClassTrib = pcn.ConteudoStr(node.Find("cClassTrib"))
	v.IndDoacao, _ = pcn.ParseIndicadorEx(pcn.ConteudoStr(node.Find("indDoacao")))

	LerGIBSCBS(node.Find("gIBSCBS"), &v.GIBSCBS)
	LerGIBSCBSMono(node.Find("gIBSCBSMono"), &v.GIBSCBSMono)
	LerGTransfCred(node.Find("gTransfCred"), &v.GTransfCred)
	LerGAjusteCompet(node.Find("gAjusteCompet"), &v.GAjusteCompet)
	LerGEstornoCred(node.Find("gEstornoCred"), &v.GEstornoCred)
	LerGCredPresOper(node.Find("gCredPresOper"), &v.GCredPresOper)
	LerGCredPresIBSZFM(node.Find("gCredPresIBSZFM"), &v.GCredPresIBSZFM)
}

// LerGIBSCBS le o grupo gIBSCBS. Porte de Ler_gIBSCBS.
func LerGIBSCBS(node *pcn.Node, v *GIBSCBS) {
	if node == nil || v == nil {
		return
	}

	v.VBC = pcn.ConteudoDe2(node.Find("vBC"))
	v.VIBS = pcn.ConteudoDe2(node.Find("vIBS"))

	LerGIBSUF(node.Find("gIBSUF"), &v.GIBSUF)
	LerGIBSMun(node.Find("gIBSMun"), &v.GIBSMun)
	LerGCBS(node.Find("gCBS"), &v.GCBS)
	LerGTribRegular(node.Find("gTribRegular"), &v.GTribRegular)
	LerGTribCompraGov(node.Find("gTribCompraGov"), &v.GTribCompraGov)
}

// LerGIBSUF le o grupo gIBSUF. Porte de Ler_gIBSUF.
func LerGIBSUF(node *pcn.Node, v *GIBSUFValores) {
	if node == nil || v == nil {
		return
	}
	v.PIBSUF = pcn.ConteudoDe4(node.Find("pIBSUF"))
	LerGDif(node.Find("gDif"), &v.GDif)
	LerGDevTrib(node.Find("gDevTrib"), &v.GDevTrib)
	LerGRed(node.Find("gRed"), &v.GRed)
	v.VIBSUF = pcn.ConteudoDe2(node.Find("vIBSUF"))
}

// LerGIBSMun le o grupo gIBSMun. Porte de Ler_gIBSMun.
func LerGIBSMun(node *pcn.Node, v *GIBSMunValores) {
	if node == nil || v == nil {
		return
	}
	v.PIBSMun = pcn.ConteudoDe4(node.Find("pIBSMun"))
	LerGDif(node.Find("gDif"), &v.GDif)
	LerGDevTrib(node.Find("gDevTrib"), &v.GDevTrib)
	LerGRed(node.Find("gRed"), &v.GRed)
	v.VIBSMun = pcn.ConteudoDe2(node.Find("vIBSMun"))
}

// LerGCBS le o grupo gCBS do item. Porte de Ler_gCBS.
func LerGCBS(node *pcn.Node, v *GCBSValores) {
	if node == nil || v == nil {
		return
	}
	v.PCBS = pcn.ConteudoDe4(node.Find("pCBS"))
	LerGDif(node.Find("gDif"), &v.GDif)
	LerGDevTrib(node.Find("gDevTrib"), &v.GDevTrib)
	LerGRed(node.Find("gRed"), &v.GRed)
	LerGALCZFMCBS(node.Find("gALCZFMCBS"), &v.GALCZFMCBS)
	v.VCBS = pcn.ConteudoDe2(node.Find("vCBS"))
}

// LerGDif le o grupo gDif. Porte de Ler_gDif.
func LerGDif(node *pcn.Node, v *GDif) {
	if node == nil || v == nil {
		return
	}
	v.PDif = pcn.ConteudoDe4(node.Find("pDif"))
	v.VDif = pcn.ConteudoDe2(node.Find("vDif"))
}

// LerGDevTrib le o grupo gDevTrib. Porte de Ler_gDevTrib.
func LerGDevTrib(node *pcn.Node, v *GDevTrib) {
	if node == nil || v == nil {
		return
	}
	v.PDevTrib = pcn.ConteudoDe4(node.Find("pDevTrib"))
	v.VDevTrib = pcn.ConteudoDe2(node.Find("vDevTrib"))
}

// LerGRed le o grupo gRed. Porte de Ler_gRed.
func LerGRed(node *pcn.Node, v *GRed) {
	if node == nil || v == nil {
		return
	}
	v.PRedAliq = pcn.ConteudoDe4(node.Find("pRedAliq"))
	v.PAliqEfet = pcn.ConteudoDe2(node.Find("pAliqEfet"))
}

// LerGTribRegular le o grupo gTribRegular. Porte de Ler_gTribRegular.
func LerGTribRegular(node *pcn.Node, v *GTribRegular) {
	if node == nil || v == nil {
		return
	}
	v.CSTReg, _ = ParseCSTIBSCBS(pcn.ConteudoStr(node.Find("CSTReg")))
	v.CClassTribReg = pcn.ConteudoStr(node.Find("cClassTribReg"))
	v.PAliqEfetRegIBSUF = pcn.ConteudoDe4(node.Find("pAliqEfetRegIBSUF"))
	v.VTribRegIBSUF = pcn.ConteudoDe2(node.Find("vTribRegIBSUF"))
	v.PAliqEfetRegIBSMun = pcn.ConteudoDe4(node.Find("pAliqEfetRegIBSMun"))
	v.VTribRegIBSMun = pcn.ConteudoDe2(node.Find("vTribRegIBSMun"))
	v.PAliqEfetRegCBS = pcn.ConteudoDe4(node.Find("pAliqEfetRegCBS"))
	v.VTribRegCBS = pcn.ConteudoDe2(node.Find("vTribRegCBS"))
}

// LerGTribCompraGov le o grupo gTribCompraGov.
// Porte de Ler_gTribCompraGov.
func LerGTribCompraGov(node *pcn.Node, v *GTribCompraGov) {
	if node == nil || v == nil {
		return
	}
	v.PAliqIBSUF = pcn.ConteudoDe4(node.Find("pAliqIBSUF"))
	v.VTribIBSUF = pcn.ConteudoDe2(node.Find("vTribIBSUF"))
	v.PAliqIBSMun = pcn.ConteudoDe4(node.Find("pAliqIBSMun"))
	v.VTribIBSMun = pcn.ConteudoDe2(node.Find("vTribIBSMun"))
	v.PAliqCBS = pcn.ConteudoDe4(node.Find("pAliqCBS"))
	v.VTribCBS = pcn.ConteudoDe2(node.Find("vTribCBS"))
}

// LerGEstornoCred le o grupo gEstornoCred do item.
// Porte de Ler_gEstornoCred.
func LerGEstornoCred(node *pcn.Node, v *GEstornoCred) {
	if node == nil || v == nil {
		return
	}
	v.VIBSEstCred = pcn.ConteudoDe2(node.Find("vIBSEstCred"))
	v.VCBSEstCred = pcn.ConteudoDe2(node.Find("vCBSEstCred"))
}

// LerGALCZFMCBS le o grupo gALCZFMCBS. Porte de Ler_gALCZFMCBS.
func LerGALCZFMCBS(node *pcn.Node, v *GALCZFMCBS) {
	if node == nil || v == nil {
		return
	}
	v.TpALCZFMCBS, _ = ParseTpALCZFMCBS(pcn.ConteudoStr(node.Find("tpALCZFMCBS")))
	v.NProcSuframa = pcn.ConteudoStr(node.Find("nProcSuframa"))
	v.PAliqEfetRegCBS = pcn.ConteudoDe4(node.Find("pAliqEfetRegCBS"))
	v.VTribRegCBS = pcn.ConteudoDe2(node.Find("vTribRegCBS"))
}

// ---------------------------------------------------------------------------
// Totais
// ---------------------------------------------------------------------------

// LerIBSCBSTot le o grupo IBSCBSTot. Porte de Ler_IBSCBSTot.
func LerIBSCBSTot(node *pcn.Node, v *IBSCBSTot) {
	if node == nil || v == nil {
		return
	}
	v.VBCIBSCBS = pcn.ConteudoDe2(node.Find("vBCIBSCBS"))
	LerGIBSTot(node.Find("gIBS"), &v.GIBS)
	LerGCBSTot(node.Find("gCBS"), &v.GCBS)
	LerGMonoTot(node.Find("gMono"), &v.GMono)
	LerGEstornoCredTot(node.Find("gEstornoCred"), &v.GEstornoCred)
}

// LerGIBSTot le o grupo gIBS do total. Porte de Ler_gIBSTot.
//
// Note que os subgrupos se chamam gIBSUF e gIBSMun no XML, mas alimentam
// GIBSUFTot e GIBSMunTot -- que sao structs diferentes das homonimas do
// item, com campos diferentes.
func LerGIBSTot(node *pcn.Node, v *GIBS) {
	if node == nil || v == nil {
		return
	}
	LerGIBSUFTot(node.Find("gIBSUF"), &v.GIBSUFTot)
	LerGIBSMunTot(node.Find("gIBSMun"), &v.GIBSMunTot)

	v.VIBS = pcn.ConteudoDe2(node.Find("vIBS"))
	v.VCredPres = pcn.ConteudoDe2(node.Find("vCredPres"))
	v.VCredPresCondSus = pcn.ConteudoDe2(node.Find("vCredPresCondSus"))
}

// LerGIBSUFTot le o grupo gIBSUF do total. Porte de Ler_gIBSUFTot.
func LerGIBSUFTot(node *pcn.Node, v *GIBSUFTot) {
	if node == nil || v == nil {
		return
	}
	v.VDif = pcn.ConteudoDe2(node.Find("vDif"))
	v.VDevTrib = pcn.ConteudoDe2(node.Find("vDevTrib"))
	v.VIBSUF = pcn.ConteudoDe2(node.Find("vIBSUF"))
}

// LerGIBSMunTot le o grupo gIBSMun do total. Porte de Ler_gIBSMunTot.
func LerGIBSMunTot(node *pcn.Node, v *GIBSMunTot) {
	if node == nil || v == nil {
		return
	}
	v.VDif = pcn.ConteudoDe2(node.Find("vDif"))
	v.VDevTrib = pcn.ConteudoDe2(node.Find("vDevTrib"))
	v.VIBSMun = pcn.ConteudoDe2(node.Find("vIBSMun"))
}

// LerGCBSTot le o grupo gCBS do total. Porte de Ler_gCBSTot.
func LerGCBSTot(node *pcn.Node, v *GCBS) {
	if node == nil || v == nil {
		return
	}
	v.VDif = pcn.ConteudoDe2(node.Find("vDif"))
	v.VDevTrib = pcn.ConteudoDe2(node.Find("vDevTrib"))
	v.VCBS = pcn.ConteudoDe2(node.Find("vCBS"))
	v.VCredPres = pcn.ConteudoDe2(node.Find("vCredPres"))
	v.VCredPresCondSus = pcn.ConteudoDe2(node.Find("vCredPresCondSus"))
}

// LerGEstornoCredTot le o grupo gEstornoCred do total.
// Porte de Ler_gEstornoCredTot -- identico a Ler_gEstornoCred no original.
func LerGEstornoCredTot(node *pcn.Node, v *GEstornoCred) {
	LerGEstornoCred(node, v)
}

// LerGMonoTot le o grupo gMono do total. Porte de Ler_gMonoTot.
//
// Atencao a precisao: aqui vIBSMonoRet e vCBSMonoRet sao De2, enquanto nos
// grupos do item (LerGMonoRetIBS, LerGMonoRetCBS) os mesmos nomes sao De4.
func LerGMonoTot(node *pcn.Node, v *GMono) {
	if node == nil || v == nil {
		return
	}
	v.VIBSMono = pcn.ConteudoDe2(node.Find("vIBSMono"))
	v.VCBSMono = pcn.ConteudoDe2(node.Find("vCBSMono"))
	v.VIBSMonoReten = pcn.ConteudoDe2(node.Find("vIBSMonoReten"))
	v.VCBSMonoReten = pcn.ConteudoDe2(node.Find("vCBSMonoReten"))
	v.VIBSMonoRet = pcn.ConteudoDe2(node.Find("vIBSMonoRet"))
	v.VCBSMonoRet = pcn.ConteudoDe2(node.Find("vCBSMonoRet"))
}

// ---------------------------------------------------------------------------
// Pagamento vinculado
// ---------------------------------------------------------------------------

// LerPgtoVinc le o grupo pgtoVinc. Porte de Ler_pgtoVinc.
func LerPgtoVinc(node *pcn.Node, v *PgtoVinc) {
	if node == nil || v == nil {
		return
	}
	for _, n := range node.FindAll("pgto") {
		LerPgto(n, &v.Pgto)
	}
}

// LerPgto acrescenta um pagamento a lista. Porte de Ler_pgto.
//
// DIVERGENCIA: o original faz StrToInt sobre o atributo nPag, o que levanta
// excecao se o atributo faltar ou nao for numerico. Aqui vale 0 -- um
// atributo ausente nao pode derrubar a importacao de um lote inteiro.
func LerPgto(node *pcn.Node, lista *[]Pgto) {
	if node == nil || lista == nil {
		return
	}
	item := Pgto{
		NPag:        pcn.AtributoInt(node, "nPag"),
		IDTransacao: node.Attr("idTransacao"),
		TpMeioPgto:  pcn.ConteudoStr(node.Find("tpMeioPgto")),
		CNPJReceb:   pcn.ConteudoStr(node.Find("CNPJReceb")),
		CNPJBasePSP: pcn.ConteudoStr(node.Find("CNPJBasePSP")),
	}
	*lista = append(*lista, item)
}

// ---------------------------------------------------------------------------
// Usado pela NF-e
// ---------------------------------------------------------------------------

// LerGCompraGov le o grupo gCompraGov na forma completa da NFe.
// Porte de Ler_gCompraGov.
func LerGCompraGov(node *pcn.Node, g *GCompraGov) {
	if node == nil || g == nil {
		return
	}
	g.TpEnteGov, _ = ParseTpEnteGov(pcn.ConteudoStr(node.Find("tpEnteGov")))
	g.PRedutor = pcn.ConteudoDe4(node.Find("pRedutor"))
	g.TpOperGov, _ = ParseTpOperGov(pcn.ConteudoStr(node.Find("tpOperGov")))

	g.RefDFeAnt = nil
	for _, n := range node.FindAll("refDFeAnt") {
		g.RefDFeAnt = append(g.RefDFeAnt, DFeRef{RefDFeChave: pcn.ConteudoStr(n)})
	}
}

// LerGPagAntecipado le o grupo gPagAntecipado da NFe, onde as chaves vem em
// elementos refNFe. Porte de Ler_gPagAntecipado.
//
// DIVERGENCIA: mesma do LerGPagAntecipadoIde -- o original nao limpa a
// lista antes do laco.
func LerGPagAntecipado(node *pcn.Node, g *GPagAntecipado) {
	if node == nil || g == nil {
		return
	}
	g.RefNFe = nil
	for _, n := range node.FindAll("refNFe") {
		g.RefNFe = append(g.RefNFe, RefDFePagAnt{RefDFeChave: pcn.ConteudoStr(n)})
	}
}

// LerISel le o grupo do Imposto Seletivo. Porte de Ler_ISel.
func LerISel(node *pcn.Node, v *GIS) {
	if node == nil || v == nil {
		return
	}
	v.CSTIS = pcn.ConteudoStr(node.Find("CSTIS"))
	v.CClassTribIS = pcn.ConteudoStr(node.Find("cClassTribIS"))
	v.VBCIS = pcn.ConteudoDe2(node.Find("vBCIS"))
	v.PIS = pcn.ConteudoDe2(node.Find("pIS"))
	v.AdRemIS = pcn.ConteudoDe4(node.Find("adRemIS"))
	v.UTrib = pcn.ConteudoStr(node.Find("uTrib"))
	v.QTrib = pcn.ConteudoDe4(node.Find("qTrib"))
	v.VIS = pcn.ConteudoDe2(node.Find("vIS"))
}

// LerISTot le o total do Imposto Seletivo. Porte de Ler_ISTot.
func LerISTot(node *pcn.Node, v *ISTot) {
	if node == nil || v == nil {
		return
	}
	v.VIS = pcn.ConteudoDe2(node.Find("vIS"))
}

// LerDFeReferenciado le um documento referenciado.
// Porte de Ler_DFeReferenciado.
func LerDFeReferenciado(node *pcn.Node, v *DFeReferenciado) {
	if node == nil || v == nil {
		return
	}
	v.ChaveAcesso = pcn.ConteudoStr(node.Find("chaveAcesso"))
	v.NItem = pcn.ConteudoInt(node.Find("nItem"))
}

// ---------------------------------------------------------------------------
// Monofasia
// ---------------------------------------------------------------------------

// LerGIBSCBSMono le o grupo gIBSCBSMono. Porte de Ler_gIBSCBSMono.
func LerGIBSCBSMono(node *pcn.Node, v *GIBSCBSMono) {
	if node == nil || v == nil {
		return
	}
	LerGIBSMonoAdRem(node.Find("gIBSMonoAdRem"), &v.GIBSMonoAdRem)
	LerGIBSMonoAdValorem(node.Find("gIBSMonoAdValorem"), &v.GIBSMonoAdValorem)
	LerGCBSMonoAdRem(node.Find("gCBSMonoAdRem"), &v.GCBSMonoAdRem)
	LerGCBSMonoAdValorem(node.Find("gCBSMonoAdValorem"), &v.GCBSMonoAdValorem)

	v.VTotIBSMonoItem = pcn.ConteudoDe2(node.Find("vTotIBSMonoItem"))
	v.VTotCBSMonoItem = pcn.ConteudoDe2(node.Find("vTotCBSMonoItem"))
}

// LerGIBSMonoAdRem le o grupo gIBSMonoAdRem. Porte de Ler_gIBSMonoAdRem.
func LerGIBSMonoAdRem(node *pcn.Node, v *GIBSMonoAdRem) {
	if node == nil || v == nil {
		return
	}
	LerGMonoPadraoIBSQtde(node.Find("gMonoPadrao"), &v.GMonoPadrao)
	LerGMonoRetenIBSQtde(node.Find("gMonoReten"), &v.GMonoReten)
	LerGMonoRetIBS(node.Find("gMonoRet"), &v.GMonoRet)
	LerGpBioDiferencaIBS(node.Find("gpBioDiferenca"), &v.GpBioDiferenca)
}

// LerGMonoPadraoIBSQtde le gMonoPadrao do grupo ad rem do IBS.
// Porte de Ler_gMonoPadraoIBSQtde.
func LerGMonoPadraoIBSQtde(node *pcn.Node, v *GMonoPadraoIBSQtde) {
	if node == nil || v == nil {
		return
	}
	v.QBCMono = pcn.ConteudoDe4(node.Find("qBCMono"))
	v.AdRemIBS = pcn.ConteudoDe4(node.Find("adRemIBS"))
	v.VIBSMono = pcn.ConteudoDe2(node.Find("vIBSMono"))
}

// LerGMonoRetenIBSQtde le gMonoReten do grupo ad rem do IBS.
// Porte de Ler_gMonoRetenIBSQtde.
func LerGMonoRetenIBSQtde(node *pcn.Node, v *GMonoRetenIBSQtde) {
	if node == nil || v == nil {
		return
	}
	v.QBCMonoReten = pcn.ConteudoDe4(node.Find("qBCMonoReten"))
	v.AdRemIBSReten = pcn.ConteudoDe4(node.Find("adRemIBSReten"))
	v.VIBSMonoReten = pcn.ConteudoDe2(node.Find("vIBSMonoReten"))
}

// LerGMonoRetIBS le gMonoRet do IBS. Porte de Ler_gMonoRetIBS.
//
// vIBSMonoRet e De4 aqui e De2 em LerGMonoTot -- mesma tag, precisoes
// diferentes, como no original.
func LerGMonoRetIBS(node *pcn.Node, v *GMonoRetIBS) {
	if node == nil || v == nil {
		return
	}
	v.VIBSMonoRet = pcn.ConteudoDe4(node.Find("vIBSMonoRet"))
}

// LerGpBioDiferencaIBS le gpBioDiferenca do IBS.
// Porte de Ler_gpBioDiferencaIBS.
func LerGpBioDiferencaIBS(node *pcn.Node, v *GpBioDiferencaIBS) {
	if node == nil || v == nil {
		return
	}
	v.QBCBioComb = pcn.ConteudoDe4(node.Find("qBCBioComb"))
	v.VIBSDiferenca = pcn.ConteudoDe2(node.Find("vIBSDiferenca"))
}

// LerGIBSMonoAdValorem le o grupo gIBSMonoAdValorem.
// Porte de Ler_gIBSMonoAdValorem.
//
// OMISSAO DO ACBr REPLICADA: gpBioDiferenca existe na struct e NAO e lido
// aqui, ao contrario do que acontece em Ler_gIBSMonoAdRem. E deterministico
// e esta assim no .pas; nao foi "corrigido" para nao divergir em silencio.
func LerGIBSMonoAdValorem(node *pcn.Node, v *GIBSMonoAdValorem) {
	if node == nil || v == nil {
		return
	}
	LerGMonoPadraoIBSAliq(node.Find("gMonoPadrao"), &v.GMonoPadrao)
	LerGMonoRetenIBSAliq(node.Find("gMonoReten"), &v.GMonoReten)
	LerGMonoRetIBS(node.Find("gMonoRet"), &v.GMonoRet)
}

// LerGMonoPadraoIBSAliq le gMonoPadrao do grupo ad valorem do IBS.
// Porte de Ler_gMonoPadraoIBSAliq.
func LerGMonoPadraoIBSAliq(node *pcn.Node, v *GMonoPadraoIBSAliq) {
	if node == nil || v == nil {
		return
	}
	v.VBCMono = pcn.ConteudoDe4(node.Find("vBCMono"))
	v.PAliqMonoUF = pcn.ConteudoDe4(node.Find("pAliqMonoUF"))
	v.VIBSMonoUF = pcn.ConteudoDe4(node.Find("vIBSMonoUF"))
	v.PAliqMonoMun = pcn.ConteudoDe4(node.Find("pAliqMonoMun"))
	v.VIBSMonoMun = pcn.ConteudoDe4(node.Find("vIBSMonoMun"))
	v.VIBSMono = pcn.ConteudoDe2(node.Find("vIBSMono"))
}

// LerGMonoRetenIBSAliq le gMonoReten do grupo ad valorem do IBS.
// Porte de Ler_gMonoRetenIBSAliq.
func LerGMonoRetenIBSAliq(node *pcn.Node, v *GMonoRetenIBSAliq) {
	if node == nil || v == nil {
		return
	}
	v.VBCMonoReten = pcn.ConteudoDe2(node.Find("vBCMonoReten"))
	v.PAliqMonoReten = pcn.ConteudoDe4(node.Find("pAliqMonoReten"))
	v.VIBSMonoReten = pcn.ConteudoDe2(node.Find("vIBSMonoReten"))
}

// LerGCBSMonoAdRem le o grupo gCBSMonoAdRem. Porte de Ler_gCBSMonoAdRem.
func LerGCBSMonoAdRem(node *pcn.Node, v *GCBSMonoAdRem) {
	if node == nil || v == nil {
		return
	}
	LerGMonoPadraoCBSQtde(node.Find("gMonoPadrao"), &v.GMonoPadrao)
	LerGMonoRetenCBSQtde(node.Find("gMonoReten"), &v.GMonoReten)
	LerGMonoRetCBS(node.Find("gMonoRet"), &v.GMonoRet)
	LerGpBioDiferencaCBS(node.Find("gpBioDiferenca"), &v.GpBioDiferenca)
}

// LerGMonoPadraoCBSQtde le gMonoPadrao do grupo ad rem da CBS.
// Porte de Ler_gMonoPadraoCBSQtde.
func LerGMonoPadraoCBSQtde(node *pcn.Node, v *GMonoPadraoCBSQtde) {
	if node == nil || v == nil {
		return
	}
	v.QBCMono = pcn.ConteudoDe4(node.Find("qBCMono"))
	v.AdRemCBS = pcn.ConteudoDe4(node.Find("adRemCBS"))
	v.VCBSMono = pcn.ConteudoDe2(node.Find("vCBSMono"))
}

// LerGMonoRetenCBSQtde le gMonoReten do grupo ad rem da CBS.
// Porte de Ler_gMonoRetenCBSQtde.
func LerGMonoRetenCBSQtde(node *pcn.Node, v *GMonoRetenCBSQtde) {
	if node == nil || v == nil {
		return
	}
	v.QBCMonoReten = pcn.ConteudoDe4(node.Find("qBCMonoReten"))
	v.AdRemCBSReten = pcn.ConteudoDe4(node.Find("adRemCBSReten"))
	v.VCBSMonoReten = pcn.ConteudoDe2(node.Find("vCBSMonoReten"))
}

// LerGMonoRetCBS le gMonoRet da CBS. Porte de Ler_gMonoRetCBS.
//
// vCBSMonoRet e De4 aqui e De2 em LerGMonoTot, como no original.
func LerGMonoRetCBS(node *pcn.Node, v *GMonoRetCBS) {
	if node == nil || v == nil {
		return
	}
	v.VCBSMonoRet = pcn.ConteudoDe4(node.Find("vCBSMonoRet"))
}

// LerGpBioDiferencaCBS le gpBioDiferenca da CBS.
// Porte de Ler_gpBioDiferencaCBS.
func LerGpBioDiferencaCBS(node *pcn.Node, v *GpBioDiferencaCBS) {
	if node == nil || v == nil {
		return
	}
	v.QBCBioComb = pcn.ConteudoDe4(node.Find("qBCBioComb"))
	v.VCBSDiferenca = pcn.ConteudoDe2(node.Find("vCBSDiferenca"))
}

// LerGCBSMonoAdValorem le o grupo gCBSMonoAdValorem.
// Porte de Ler_gCBSMonoAdValorem.
//
// OMISSAO DO ACBr REPLICADA: gpBioDiferenca existe na struct e nao e lido,
// espelhando Ler_gIBSMonoAdValorem.
func LerGCBSMonoAdValorem(node *pcn.Node, v *GCBSMonoAdValorem) {
	if node == nil || v == nil {
		return
	}
	LerGMonoPadraoCBSAliq(node.Find("gMonoPadrao"), &v.GMonoPadrao)
	LerGMonoRetenCBSAliq(node.Find("gMonoReten"), &v.GMonoReten)
	LerGMonoRetCBS(node.Find("gMonoRet"), &v.GMonoRet)
}

// LerGMonoPadraoCBSAliq le gMonoPadrao do grupo ad valorem da CBS.
// Porte de Ler_gMonoPadraoCBSAliq.
func LerGMonoPadraoCBSAliq(node *pcn.Node, v *GMonoPadraoCBSAliq) {
	if node == nil || v == nil {
		return
	}
	v.VBCMono = pcn.ConteudoDe4(node.Find("vBCMono"))
	v.PAliqMonoCBS = pcn.ConteudoDe4(node.Find("pAliqMonoCBS"))
	v.VCBSMono = pcn.ConteudoDe2(node.Find("vCBSMono"))
}

// LerGMonoRetenCBSAliq le gMonoReten do grupo ad valorem da CBS.
// Porte de Ler_gMonoRetenCBSAliq.
func LerGMonoRetenCBSAliq(node *pcn.Node, v *GMonoRetenCBSAliq) {
	if node == nil || v == nil {
		return
	}
	v.VBCMonoReten = pcn.ConteudoDe2(node.Find("vBCMonoReten"))
	v.PAliqMonoReten = pcn.ConteudoDe4(node.Find("pAliqMonoReten"))
	v.VCBSMonoReten = pcn.ConteudoDe2(node.Find("vCBSMonoReten"))
}

// ---------------------------------------------------------------------------
// Creditos e ajustes
// ---------------------------------------------------------------------------

// LerGTransfCred le o grupo gTransfCred. Porte de Ler_gTransfCred.
func LerGTransfCred(node *pcn.Node, v *GTransfCred) {
	if node == nil || v == nil {
		return
	}
	v.VIBS = pcn.ConteudoDe2(node.Find("vIBS"))
	v.VCBS = pcn.ConteudoDe2(node.Find("vCBS"))
}

// LerGCredPresIBSZFM le o grupo gCredPresIBSZFM.
// Porte de Ler_gCredPresIBSZFM.
//
// competApur vem como AAAA-MM; o original concatena "-01" e converte com
// StringToDateTime(aData, 'YYYY-MM-DD') -- formato explicito, sem depender
// de locale. Aqui a competencia e montada direto; o resultado e o mesmo.
func LerGCredPresIBSZFM(node *pcn.Node, v *CredPresIBSZFM) {
	if node == nil || v == nil {
		return
	}
	v.CompetApur = parseCompetenciaHifen(pcn.ConteudoStr(node.Find("competApur")))
	v.TpCredPresIBSZFM, _ = ParseTpCredPresIBSZFM(pcn.ConteudoStr(node.Find("tpCredPresIBSZFM")))
	v.VCredPresIBSZFM = pcn.ConteudoDe2(node.Find("vCredPresIBSZFM"))
}

// LerGAjusteCompet le o grupo gAjusteCompet. Porte de Ler_gAjusteCompet.
func LerGAjusteCompet(node *pcn.Node, v *GAjusteCompet) {
	if node == nil || v == nil {
		return
	}
	v.CompetApur = parseCompetenciaHifen(pcn.ConteudoStr(node.Find("competApur")))
	v.VIBS = pcn.ConteudoDe2(node.Find("vIBS"))
	v.VCBS = pcn.ConteudoDe2(node.Find("vCBS"))
}

// LerGCredPresOper le o grupo gCredPresOper. Porte de Ler_gCredPresOper.
func LerGCredPresOper(node *pcn.Node, v *GCredPresOper) {
	if node == nil || v == nil {
		return
	}
	v.VBCCredPres = pcn.ConteudoDe2(node.Find("vBCCredPres"))
	v.CCredPres, _ = ParseCCredPres(pcn.ConteudoStr(node.Find("cCredPres")))
	LerGIBSCredPres(node.Find("gIBSCredPres"), &v.GIBSCredPres)
	LerGCBSCredPres(node.Find("gCBSCredPres"), &v.GCBSCredPres)
}

// LerGIBSCredPres le o grupo gIBSCredPres. Porte de Ler_gIBSCredPres.
func LerGIBSCredPres(node *pcn.Node, v *GIBSCBSCredPres) {
	lerCredPres(node, v)
}

// LerGCBSCredPres le o grupo gCBSCredPres. Porte de Ler_gCBSCredPres --
// identico ao do IBS no original.
func LerGCBSCredPres(node *pcn.Node, v *GIBSCBSCredPres) {
	lerCredPres(node, v)
}

func lerCredPres(node *pcn.Node, v *GIBSCBSCredPres) {
	if node == nil || v == nil {
		return
	}
	v.PCredPres = pcn.ConteudoDe4(node.Find("pCredPres"))
	v.VCredPres = pcn.ConteudoDe2(node.Find("vCredPres"))
	v.VCredPresCondSus = pcn.ConteudoDe2(node.Find("vCredPresCondSus"))
}

// parseCompetenciaHifen interpreta a competencia no formato AAAA-MM usado
// pelos grupos de credito e ajuste. Conteudo invalido devolve tempo zero,
// como no original.
func parseCompetenciaHifen(s string) time.Time {
	return pcn.ParseCompetenciaDef(pcn.OnlyNumber(s))
}
