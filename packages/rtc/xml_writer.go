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
	"strconv"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Gerador dos grupos da Reforma Tributaria (IBS/CBS/IS) -- porte 1:1 de
// TDFeRTCXmlWriter (ACBrDFe.RTC.XmlWriter.pas). Cada metodo GerarXxx
// corresponde a um Gerar_Xxx do original, com a mesma ordem de campos e as
// mesmas condicionais; devolve nil quando o Delphi devolve nil.
//
// O writer carrega ESTADO entre chamadas, como o original:
//   - pRedutor/tpEnteGov sao capturados por GerarGCompraGovReduzido (ou
//     GerarGCompraGov) e mudam as condicionais de gRed e gTribCompraGov nos
//     itens -- por isso o MESMO Writer precisa gerar o documento inteiro;
//   - gerarIBSCBSTot e ligado quando algum item gera o grupo IBSCBS, e e o
//     que autoriza GerarIBSCBSTot no total do documento.

// ModeloDFe identifica o documento que esta sendo gerado, porque o grupo
// IBSCBS muda por modelo. Porte de TModelosDFe (ACBrDFe.Conversao.pas).
type ModeloDFe int

const (
	ModeloBPe ModeloDFe = iota
	ModeloBPeTM
	ModeloBPeTA
	ModeloCTe
	ModeloCTeOS
	ModeloCTeSimp
	ModeloGTVe
	ModeloNF3e
	ModeloNFAg
	ModeloNFCom
	ModeloNFe
	ModeloNFCe
	ModeloNFGas
)

// TpNFDebito e o tipo de NF-e de debito. Porte de TtpNFDebito
// (ACBrDFe.Conversao.pas); os codigos sao os de TtpNFDebitoArrayStrings.
type TpNFDebito int

const (
	TdNenhum                          TpNFDebito = iota // vazio
	TdTransferenciaCreditoCooperativa                   // 01
	TdAnulacao                                          // 02
	TdDebitosNaoProcessadas                             // 03
	TdMultaJuros                                        // 04
	TdTransferenciaCreditoSucessao                      // 05
	TdPagamentoAntecipado                               // 06
	TdPerdaEmEstoque                                    // 07
	TdDesenquadramentodoSN                              // 08
)

// Writer gera os grupos XML da Reforma Tributaria.
// Porte de TDFeRTCXmlWriter.
type Writer struct {
	// Modelo e o DFe em geracao (FModelosDFe).
	Modelo ModeloDFe
	// TpNFDebito so importa para NFe (FtpNFDebito).
	TpNFDebito TpNFDebito

	pRedutor       float64   // FpRedutor
	tpEnteGov      TpEnteGov // FtpEnteGov
	gerarIBSCBSTot bool      // FpGerarGrupoIBSCBSTot
}

// NovoWriter cria o writer para um documento. Use um Writer novo por
// documento gerado: o estado interno nao pode vazar entre documentos.
func NovoWriter(modelo ModeloDFe) *Writer {
	return &Writer{Modelo: modelo}
}

// ---------------------------------------------------------------------------
// Usado pela maioria dos DF-e
// ---------------------------------------------------------------------------

// GerarGCompraGovReduzido -- porte de Gerar_gCompraGovReduzido.
// CAPTURA pRedutor/tpEnteGov no estado do writer ANTES do teste, exatamente
// como o original: mesmo sem gerar o grupo, o estado influencia gRed e
// gTribCompraGov dos itens.
func (w *Writer) GerarGCompraGovReduzido(g GCompraGovReduzido) *pcn.Elem {
	w.pRedutor = g.PRedutor
	w.tpEnteGov = g.TpEnteGov

	if w.pRedutor <= 0 {
		return nil
	}
	e := pcn.NovoElem("gCompraGov").
		Filho(pcn.NodeStr("tpEnteGov", g.TpEnteGov.String(), true)).
		Filho(pcn.NodeDec("pRedutor", g.PRedutor, 4, true)).
		Filho(pcn.NodeStr("tpOperGov", g.TpOperGov.String(), true))
	for _, ref := range g.RefDFe {
		e.Filho(pcn.NodeStrSemFiltro("refDFeAnt", ref.RefDFeAnt, true))
	}
	return e
}

// GerarGPagAntecipadoProd -- porte de Gerar_gPagAntecipadoProd (nivel item).
func (w *Writer) GerarGPagAntecipadoProd(g GPagAntecipadoProd) *pcn.Elem {
	return pcn.NovoElem("gPagAntecipado").
		Filho(pcn.NodeStrSemFiltro("chDFePagAnt", g.ChDFePagAnt, true)).
		Filho(pcn.NodeInt("nItemPagAnt", g.NItemPagAnt, 1, false))
}

// GerarGPagAntecipadoIde -- porte de Gerar_gPagAntecipadoIde (nivel ide).
func (w *Writer) GerarGPagAntecipadoIde(g GPagAntecipado) *pcn.Elem {
	if len(g.RefNFe) == 0 {
		return nil
	}
	e := pcn.NovoElem("gPagAntecipado")
	for _, ref := range g.RefNFe {
		e.Filho(pcn.NodeStrSemFiltro("chDFePagAnt", ref.RefDFeChave, true))
	}
	return e
}

// GerarIBSCBS -- porte de Gerar_IBSCBS. O corpo por modelo e delegado como
// no original; o case-else do Delphi faz AppendChild(nil), que aqui e o
// Filho(nil) inofensivo.
func (w *Writer) GerarIBSCBS(ibscbs IBSCBS) *pcn.Elem {
	var e *pcn.Elem

	if ibscbs.CST != CSTNenhum && ibscbs.CClassTrib != "" {
		w.gerarIBSCBSTot = true
		e = pcn.NovoElem("IBSCBS").
			Filho(pcn.NodeStr("CST", ibscbs.CST.String(), true)).
			Filho(pcn.NodeStr("cClassTrib", ibscbs.CClassTrib, true))

		if ibscbs.IndDoacao == pcn.TieSim {
			e.Filho(pcn.NodeStr("indDoacao", "1", false))
		}

		switch w.Modelo {
		case ModeloBPe, ModeloBPeTM, ModeloBPeTA:
			e.Filho(w.gerarIBSCBSBPe(ibscbs))
		case ModeloCTe, ModeloCTeOS, ModeloCTeSimp, ModeloGTVe:
			e.Filho(w.gerarIBSCBSCTe(ibscbs))
		case ModeloNF3e:
			e.Filho(w.gerarIBSCBSNF3e(ibscbs))
		case ModeloNFAg:
			e.Filho(w.gerarIBSCBSNFAg(ibscbs))
		case ModeloNFCom:
			e.Filho(w.gerarIBSCBSNFCom(ibscbs))
		case ModeloNFe, ModeloNFCe:
			e.Filho(w.gerarIBSCBSNFe(ibscbs))
		case ModeloNFGas:
			e.Filho(w.gerarIBSCBSNFGas(ibscbs))
		}
	}

	// DIVERGENCIA (defensiva): no Delphi estes AppendChild rodam mesmo com
	// Result=nil e estourariam com access violation se a condicao fosse
	// verdadeira sem o grupo principal; aqui Filho sobre nil e no-op.
	if ibscbs.GEstornoCred.VIBSEstCred > 0 || ibscbs.GEstornoCred.VCBSEstCred > 0 ||
		(w.Modelo == ModeloNFe && w.TpNFDebito == TdPerdaEmEstoque) {
		e.Filho(w.gerarGEstornoCred(ibscbs.GEstornoCred))
	}

	if w.Modelo == ModeloNFe {
		if ibscbs.GCredPresOper.CCredPres != CpNenhum {
			e.Filho(w.gerarGCredPresOper(ibscbs.GCredPresOper))
		} else if ibscbs.GCredPresIBSZFM.TpCredPresIBSZFM != TcpNenhum {
			e.Filho(w.gerarGCredPresIBSZFM(ibscbs.GCredPresIBSZFM))
		}
	}
	return e
}

func (w *Writer) gerarIBSCBSBPe(ibscbs IBSCBS) *pcn.Elem {
	switch w.Modelo {
	case ModeloBPe, ModeloBPeTA:
		switch ibscbs.CST {
		case CST000, CST200, CST222:
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		}
	case ModeloBPeTM:
		if ibscbs.CST == CST200 {
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		}
	}
	return nil
}

func (w *Writer) gerarIBSCBSCTe(ibscbs IBSCBS) *pcn.Elem {
	switch w.Modelo {
	case ModeloCTe, ModeloCTeSimp:
		switch ibscbs.CST {
		case CST000, CST200:
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		}
	case ModeloCTeOS:
		switch ibscbs.CST {
		case CST000, CST200, CST222:
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		}
	}
	return nil
}

func (w *Writer) gerarIBSCBSNF3e(ibscbs IBSCBS) *pcn.Elem {
	switch ibscbs.CST {
	case CST000, CST200, CST510, CST830:
		return w.gerarGIBSCBS(ibscbs.GIBSCBS)
	}
	return nil
}

func (w *Writer) gerarIBSCBSNFAg(ibscbs IBSCBS) *pcn.Elem {
	if ibscbs.CST == CST000 {
		return w.gerarGIBSCBS(ibscbs.GIBSCBS)
	}
	return nil
}

func (w *Writer) gerarIBSCBSNFCom(ibscbs IBSCBS) *pcn.Elem {
	switch ibscbs.CST {
	case CST000, CST200:
		return w.gerarGIBSCBS(ibscbs.GIBSCBS)
	}
	return nil
}

func (w *Writer) gerarIBSCBSNFe(ibscbs IBSCBS) *pcn.Elem {
	switch w.Modelo {
	case ModeloNFe:
		switch ibscbs.CST {
		case CST000, CST200, CST510, CST515, CST550, CST830:
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		case CST620:
			return w.gerarGIBSCBSMono(ibscbs.GIBSCBSMono)
		case CST800:
			return w.gerarGTransfCred(ibscbs.GTransfCred)
		case CST811:
			return w.gerarGAjusteCompet(ibscbs.GAjusteCompet)
		}
	case ModeloNFCe:
		switch ibscbs.CST {
		case CST000, CST200:
			return w.gerarGIBSCBS(ibscbs.GIBSCBS)
		case CST620:
			return w.gerarGIBSCBSMono(ibscbs.GIBSCBSMono)
		}
	}
	return nil
}

func (w *Writer) gerarIBSCBSNFGas(ibscbs IBSCBS) *pcn.Elem {
	if ibscbs.CST == CST000 {
		return w.gerarGIBSCBS(ibscbs.GIBSCBS)
	}
	return nil
}

func (w *Writer) gerarGIBSCBS(g GIBSCBS) *pcn.Elem {
	e := pcn.NovoElem("gIBSCBS").
		Filho(pcn.NodeDec("vBC", g.VBC, 2, true)).
		Filho(w.gerarGIBSUF(g.GIBSUF)).
		Filho(w.gerarGIBSMun(g.GIBSMun)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(w.gerarGCBS(g.GCBS))

	if g.GTribRegular.CSTReg != CSTNenhum {
		e.Filho(w.gerarGTribRegular(g.GTribRegular))
	}
	if g.GTribCompraGov.PAliqIBSUF > 0 && w.tpEnteGov != TcgNenhum {
		e.Filho(w.gerarGTribCompraGov(g.GTribCompraGov))
	}
	return e
}

func (w *Writer) gerarGIBSUF(g GIBSUFValores) *pcn.Elem {
	e := pcn.NovoElem("gIBSUF").
		Filho(pcn.NodeDec("pIBSUF", g.PIBSUF, 4, true))
	if g.GDif.PDif > 0 {
		e.Filho(w.gerarGDif(g.GDif))
	}
	if g.GDevTrib.VDevTrib > 0 {
		e.Filho(w.gerarGDevTrib(g.GDevTrib))
	}
	if g.GRed.PRedAliq > 0 || g.GRed.PAliqEfet > 0 || w.pRedutor > 0 {
		e.Filho(w.gerarGRed(g.GRed))
	}
	return e.Filho(pcn.NodeDec("vIBSUF", g.VIBSUF, 2, true))
}

func (w *Writer) gerarGIBSMun(g GIBSMunValores) *pcn.Elem {
	e := pcn.NovoElem("gIBSMun").
		Filho(pcn.NodeDec("pIBSMun", g.PIBSMun, 4, true))
	if g.GDif.PDif > 0 {
		e.Filho(w.gerarGDif(g.GDif))
	}
	if g.GDevTrib.VDevTrib > 0 {
		e.Filho(w.gerarGDevTrib(g.GDevTrib))
	}
	if g.GRed.PRedAliq > 0 || g.GRed.PAliqEfet > 0 || w.pRedutor > 0 {
		e.Filho(w.gerarGRed(g.GRed))
	}
	return e.Filho(pcn.NodeDec("vIBSMun", g.VIBSMun, 2, true))
}

func (w *Writer) gerarGCBS(g GCBSValores) *pcn.Elem {
	e := pcn.NovoElem("gCBS").
		Filho(pcn.NodeDec("pCBS", g.PCBS, 4, true))
	if g.GDif.PDif > 0 {
		e.Filho(w.gerarGDif(g.GDif))
	}
	if g.GDevTrib.VDevTrib > 0 {
		e.Filho(w.gerarGDevTrib(g.GDevTrib))
	}
	if g.GRed.PRedAliq > 0 || g.GRed.PAliqEfet > 0 || w.pRedutor > 0 {
		e.Filho(w.gerarGRed(g.GRed))
	}
	if g.GALCZFMCBS.PAliqEfetRegCBS > 0 || g.GALCZFMCBS.VTribRegCBS > 0 {
		e.Filho(w.gerarGALCZFMCBS(g.GALCZFMCBS))
	}
	return e.Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true))
}

func (w *Writer) gerarGDif(g GDif) *pcn.Elem {
	return pcn.NovoElem("gDif").
		Filho(pcn.NodeDec("pDif", g.PDif, 4, true)).
		Filho(pcn.NodeDec("vDif", g.VDif, 2, true))
}

func (w *Writer) gerarGDevTrib(g GDevTrib) *pcn.Elem {
	return pcn.NovoElem("gDevTrib").
		Filho(pcn.NodeDec("pDevTrib", g.PDevTrib, 2, false)).
		Filho(pcn.NodeDec("vDevTrib", g.VDevTrib, 2, true))
}

func (w *Writer) gerarGRed(g GRed) *pcn.Elem {
	// pAliqEfet e tcDe2 no original (nao De4), diferente do pRedAliq.
	return pcn.NovoElem("gRed").
		Filho(pcn.NodeDec("pRedAliq", g.PRedAliq, 4, true)).
		Filho(pcn.NodeDec("pAliqEfet", g.PAliqEfet, 2, true))
}

func (w *Writer) gerarGTribRegular(g GTribRegular) *pcn.Elem {
	return pcn.NovoElem("gTribRegular").
		Filho(pcn.NodeStr("CSTReg", g.CSTReg.String(), true)).
		Filho(pcn.NodeStr("cClassTribReg", g.CClassTribReg, true)).
		Filho(pcn.NodeDec("pAliqEfetRegIBSUF", g.PAliqEfetRegIBSUF, 4, true)).
		Filho(pcn.NodeDec("vTribRegIBSUF", g.VTribRegIBSUF, 2, true)).
		Filho(pcn.NodeDec("pAliqEfetRegIBSMun", g.PAliqEfetRegIBSMun, 4, true)).
		Filho(pcn.NodeDec("vTribRegIBSMun", g.VTribRegIBSMun, 2, true)).
		Filho(pcn.NodeDec("pAliqEfetRegCBS", g.PAliqEfetRegCBS, 4, true)).
		Filho(pcn.NodeDec("vTribRegCBS", g.VTribRegCBS, 2, true))
}

func (w *Writer) gerarGTribCompraGov(g GTribCompraGov) *pcn.Elem {
	return pcn.NovoElem("gTribCompraGov").
		Filho(pcn.NodeDec("pAliqIBSUF", g.PAliqIBSUF, 4, true)).
		Filho(pcn.NodeDec("vTribIBSUF", g.VTribIBSUF, 2, true)).
		Filho(pcn.NodeDec("pAliqIBSMun", g.PAliqIBSMun, 4, true)).
		Filho(pcn.NodeDec("vTribIBSMun", g.VTribIBSMun, 2, true)).
		Filho(pcn.NodeDec("pAliqCBS", g.PAliqCBS, 4, true)).
		Filho(pcn.NodeDec("vTribCBS", g.VTribCBS, 2, true))
}

func (w *Writer) gerarGEstornoCred(g GEstornoCred) *pcn.Elem {
	return pcn.NovoElem("gEstornoCred").
		Filho(pcn.NodeDec("vIBSEstCred", g.VIBSEstCred, 2, true)).
		Filho(pcn.NodeDec("vCBSEstCred", g.VCBSEstCred, 2, true))
}

func (w *Writer) gerarGALCZFMCBS(g GALCZFMCBS) *pcn.Elem {
	return pcn.NovoElem("gALCZFMCBS").
		Filho(pcn.NodeStr("tpALCZFMCBS", g.TpALCZFMCBS.String(), true)).
		Filho(pcn.NodeStr("nProcSuframa", g.NProcSuframa, false)).
		Filho(pcn.NodeDec("pAliqEfetRegCBS", g.PAliqEfetRegCBS, 4, true)).
		Filho(pcn.NodeDec("vTribRegCBS", g.VTribRegCBS, 2, true))
}

// GerarIBSCBSTot -- porte de Gerar_IBSCBSTot. So gera se algum item gerou o
// grupo IBSCBS (estado interno ligado por GerarIBSCBS).
func (w *Writer) GerarIBSCBSTot(tot IBSCBSTot) *pcn.Elem {
	if !w.gerarIBSCBSTot {
		return nil
	}
	e := pcn.NovoElem("IBSCBSTot").
		Filho(pcn.NodeDec("vBCIBSCBS", tot.VBCIBSCBS, 2, true))

	if tot.GIBS.VIBS > 0 ||
		tot.GIBS.GIBSUFTot.VDif > 0 || tot.GIBS.GIBSMunTot.VDif > 0 ||
		tot.GIBS.GIBSUFTot.VDevTrib > 0 || tot.GIBS.GIBSMunTot.VDevTrib > 0 ||
		tot.GIBS.VCredPres > 0 || tot.GIBS.VCredPresCondSus > 0 {
		e.Filho(w.gerarGIBSTot(tot.GIBS))
	}

	if tot.GCBS.VCBS > 0 || tot.GCBS.VDif > 0 || tot.GCBS.VDevTrib > 0 ||
		tot.GCBS.VCredPres > 0 || tot.GCBS.VCredPresCondSus > 0 {
		e.Filho(w.gerarGCBSTot(tot.GCBS))
	}

	if tot.GMono.VIBSMono > 0 || tot.GMono.VCBSMono > 0 ||
		tot.GMono.VIBSMonoReten > 0 || tot.GMono.VCBSMonoReten > 0 ||
		tot.GMono.VIBSMonoRet > 0 || tot.GMono.VCBSMonoRet > 0 {
		e.Filho(w.gerarGMonoTot(tot.GMono))
	}

	if tot.GEstornoCred.VIBSEstCred > 0 || tot.GEstornoCred.VCBSEstCred > 0 {
		e.Filho(w.gerarGEstornoCredTot(tot.GEstornoCred))
	}
	return e
}

func (w *Writer) gerarGIBSTot(g GIBS) *pcn.Elem {
	return pcn.NovoElem("gIBS").
		Filho(w.gerarGIBSUFTot(g.GIBSUFTot)).
		Filho(w.gerarGIBSMunTot(g.GIBSMunTot)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCredPres", g.VCredPres, 2, true)).
		Filho(pcn.NodeDec("vCredPresCondSus", g.VCredPresCondSus, 2, true))
}

func (w *Writer) gerarGIBSUFTot(g GIBSUFTot) *pcn.Elem {
	return pcn.NovoElem("gIBSUF").
		Filho(pcn.NodeDec("vDif", g.VDif, 2, true)).
		Filho(pcn.NodeDec("vDevTrib", g.VDevTrib, 2, true)).
		Filho(pcn.NodeDec("vIBSUF", g.VIBSUF, 2, true))
}

func (w *Writer) gerarGIBSMunTot(g GIBSMunTot) *pcn.Elem {
	return pcn.NovoElem("gIBSMun").
		Filho(pcn.NodeDec("vDif", g.VDif, 2, true)).
		Filho(pcn.NodeDec("vDevTrib", g.VDevTrib, 2, true)).
		Filho(pcn.NodeDec("vIBSMun", g.VIBSMun, 2, true))
}

func (w *Writer) gerarGCBSTot(g GCBS) *pcn.Elem {
	return pcn.NovoElem("gCBS").
		Filho(pcn.NodeDec("vDif", g.VDif, 2, true)).
		Filho(pcn.NodeDec("vDevTrib", g.VDevTrib, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NodeDec("vCredPres", g.VCredPres, 2, true)).
		Filho(pcn.NodeDec("vCredPresCondSus", g.VCredPresCondSus, 2, true))
}

func (w *Writer) gerarGEstornoCredTot(g GEstornoCred) *pcn.Elem {
	return pcn.NovoElem("gEstornoCred").
		Filho(pcn.NodeDec("vIBSEstCred", g.VIBSEstCred, 2, true)).
		Filho(pcn.NodeDec("vCBSEstCred", g.VCBSEstCred, 2, true))
}

// GerarPgtoVinc -- porte de Gerar_pgtoVinc + Gerar_pgto.
func (w *Writer) GerarPgtoVinc(pv PgtoVinc) *pcn.Elem {
	if len(pv.Pgto) == 0 {
		return nil
	}
	e := pcn.NovoElem("pgtoVinc")
	for _, p := range pv.Pgto {
		// os DOIS atributos sao incondicionais no original (SetAttribute
		// sempre), mesmo com valor vazio -- replicado.
		e.Filho(pcn.NovoElem("pgto").
			Attr("nPag", strconv.Itoa(p.NPag)).
			Attr("idTransacao", p.IDTransacao).
			Filho(pcn.NodeStr("tpMeioPgto", p.TpMeioPgto, true)).
			Filho(pcn.NodeStr("CNPJReceb", p.CNPJReceb, true)).
			Filho(pcn.NodeStr("CNPJBasePSP", p.CNPJBasePSP, true)))
	}
	return e
}

// ---------------------------------------------------------------------------
// Usado pela NF-e
// ---------------------------------------------------------------------------

// GerarGCompraGov -- porte de Gerar_gCompraGov (variante da NFe, com a
// colecao refDFeAnt propria). Tambem captura o estado pRedutor/tpEnteGov.
func (w *Writer) GerarGCompraGov(g GCompraGov) *pcn.Elem {
	w.pRedutor = g.PRedutor
	w.tpEnteGov = g.TpEnteGov

	if w.pRedutor <= 0 {
		return nil
	}
	e := pcn.NovoElem("gCompraGov").
		Filho(pcn.NodeStr("tpEnteGov", g.TpEnteGov.String(), true)).
		Filho(pcn.NodeDec("pRedutor", g.PRedutor, 4, true)).
		Filho(pcn.NodeStr("tpOperGov", g.TpOperGov.String(), true))
	for _, ref := range g.RefDFeAnt {
		e.Filho(pcn.NodeStrSemFiltro("refDFeAnt", ref.RefDFeChave, true))
	}
	return e
}

// GerarGPagAntecipado -- porte de Gerar_gPagAntecipado (NFe, tag refNFe).
func (w *Writer) GerarGPagAntecipado(g GPagAntecipado) *pcn.Elem {
	if len(g.RefNFe) == 0 {
		return nil
	}
	e := pcn.NovoElem("gPagAntecipado")
	for _, ref := range g.RefNFe {
		e.Filho(pcn.NodeStrSemFiltro("refNFe", ref.RefDFeChave, true))
	}
	return e
}

// GerarISel -- porte de Gerar_ISel (Imposto Seletivo). CSTIS vai como string
// crua ate a publicacao da tabela oficial, como no original.
func (w *Writer) GerarISel(is GIS) *pcn.Elem {
	e := pcn.NovoElem("IS").
		Filho(pcn.NodeStr("CSTIS", is.CSTIS, true)).
		Filho(pcn.NodeStr("cClassTribIS", is.CClassTribIS, true))

	if is.VBCIS > 0 || is.PIS > 0 || is.UTrib != "" || is.QTrib > 0 || is.VIS > 0 {
		e.Filho(pcn.NodeDec("vBCIS", is.VBCIS, 2, true)).
			Filho(pcn.NodeDec("pIS", is.PIS, 2, true)).
			Filho(pcn.NodeDec("adRemIS", is.AdRemIS, 4, false))
		if is.UTrib != "" || is.QTrib > 0 {
			e.Filho(pcn.NodeStr("uTrib", is.UTrib, true)).
				Filho(pcn.NodeDec("qTrib", is.QTrib, 4, false))
		}
		e.Filho(pcn.NodeDec("vIS", is.VIS, 2, true))
	}
	return e
}

func (w *Writer) gerarGIBSCBSMono(g GIBSCBSMono) *pcn.Elem {
	gerarIBSAdRem := g.GIBSMonoAdRem.GMonoPadrao.QBCMono > 0 ||
		g.GIBSMonoAdRem.GMonoReten.QBCMonoReten > 0 ||
		g.GIBSMonoAdRem.GMonoRet.VIBSMonoRet > 0
	gerarIBSAdValorem := g.GIBSMonoAdValorem.GMonoPadrao.VBCMono > 0 ||
		g.GIBSMonoAdValorem.GMonoReten.VBCMonoReten > 0 ||
		g.GIBSMonoAdValorem.GMonoRet.VIBSMonoRet > 0
	gerarCBSAdRem := g.GCBSMonoAdRem.GMonoPadrao.QBCMono > 0 ||
		g.GCBSMonoAdRem.GMonoReten.QBCMonoReten > 0 ||
		g.GCBSMonoAdRem.GMonoRet.VCBSMonoRet > 0
	gerarCBSAdValorem := g.GCBSMonoAdValorem.GMonoPadrao.VBCMono > 0 ||
		g.GCBSMonoAdValorem.GMonoReten.VBCMonoReten > 0 ||
		g.GCBSMonoAdValorem.GMonoRet.VCBSMonoRet > 0

	if !gerarIBSAdRem && !gerarIBSAdValorem && !gerarCBSAdRem && !gerarCBSAdValorem {
		return nil
	}

	e := pcn.NovoElem("gIBSCBSMono")
	if gerarIBSAdRem {
		e.Filho(w.gerarGIBSMonoAdRem(g.GIBSMonoAdRem))
	} else if gerarIBSAdValorem {
		e.Filho(w.gerarGIBSMonoAdValorem(g.GIBSMonoAdValorem))
	}
	if gerarCBSAdRem {
		e.Filho(w.gerarGCBSMonoAdRem(g.GCBSMonoAdRem))
	} else if gerarCBSAdValorem {
		e.Filho(w.gerarGCBSMonoAdValorem(g.GCBSMonoAdValorem))
	}
	return e.
		Filho(pcn.NodeDec("vTotIBSMonoItem", g.VTotIBSMonoItem, 2, true)).
		Filho(pcn.NodeDec("vTotCBSMonoItem", g.VTotCBSMonoItem, 2, true))
}

func (w *Writer) gerarGIBSMonoAdRem(g GIBSMonoAdRem) *pcn.Elem {
	e := pcn.NovoElem("gIBSMonoAdRem")
	if g.GMonoPadrao.QBCMono > 0 {
		e.Filho(pcn.NovoElem("gMonoPadrao").
			Filho(pcn.NodeDec("qBCMono", g.GMonoPadrao.QBCMono, 4, false)).
			Filho(pcn.NodeDec("adRemIBS", g.GMonoPadrao.AdRemIBS, 4, true)).
			Filho(pcn.NodeDec("vIBSMono", g.GMonoPadrao.VIBSMono, 2, true)))
	}
	if g.GMonoReten.QBCMonoReten > 0 {
		e.Filho(pcn.NovoElem("gMonoReten").
			Filho(pcn.NodeDec("qBCMonoReten", g.GMonoReten.QBCMonoReten, 4, false)).
			Filho(pcn.NodeDec("adRemIBSReten", g.GMonoReten.AdRemIBSReten, 4, true)).
			Filho(pcn.NodeDec("vIBSMonoReten", g.GMonoReten.VIBSMonoReten, 2, true)))
	}
	if g.GMonoRet.VIBSMonoRet > 0 {
		e.Filho(w.gerarGMonoRetIBS(g.GMonoRet))
	}
	if g.GpBioDiferenca.QBCBioComb > 0 {
		e.Filho(w.gerarGpBioDiferencaIBS(g.GpBioDiferenca))
	}
	return e
}

func (w *Writer) gerarGMonoRetIBS(g GMonoRetIBS) *pcn.Elem {
	return pcn.NovoElem("gMonoRet").
		Filho(pcn.NodeDec("vIBSMonoRet", g.VIBSMonoRet, 2, true))
}

func (w *Writer) gerarGpBioDiferencaIBS(g GpBioDiferencaIBS) *pcn.Elem {
	return pcn.NovoElem("gpBioDiferenca").
		Filho(pcn.NodeDec("qBCBioComb", g.QBCBioComb, 4, true)).
		Filho(pcn.NodeDec("vIBSDiferenca", g.VIBSDiferenca, 2, true))
}

func (w *Writer) gerarGIBSMonoAdValorem(g GIBSMonoAdValorem) *pcn.Elem {
	e := pcn.NovoElem("gIBSMonoAdValorem")
	if g.GMonoPadrao.VBCMono > 0 {
		e.Filho(pcn.NovoElem("gMonoPadrao").
			Filho(pcn.NodeDec("vBCMono", g.GMonoPadrao.VBCMono, 2, false)).
			Filho(pcn.NodeDec("pAliqMonoUF", g.GMonoPadrao.PAliqMonoUF, 4, true)).
			Filho(pcn.NodeDec("vIBSMonoUF", g.GMonoPadrao.VIBSMonoUF, 2, true)).
			Filho(pcn.NodeDec("pAliqMonoMun", g.GMonoPadrao.PAliqMonoMun, 4, true)).
			Filho(pcn.NodeDec("vIBSMonoMun", g.GMonoPadrao.VIBSMonoMun, 2, true)).
			Filho(pcn.NodeDec("vIBSMono", g.GMonoPadrao.VIBSMono, 2, true)))
	}
	if g.GMonoReten.VBCMonoReten > 0 {
		e.Filho(pcn.NovoElem("gMonoReten").
			Filho(pcn.NodeDec("vBCMonoReten", g.GMonoReten.VBCMonoReten, 2, false)).
			Filho(pcn.NodeDec("pAliqMonoReten", g.GMonoReten.PAliqMonoReten, 4, true)).
			Filho(pcn.NodeDec("vIBSMonoReten", g.GMonoReten.VIBSMonoReten, 2, true)))
	}
	if g.GMonoRet.VIBSMonoRet > 0 {
		e.Filho(w.gerarGMonoRetIBS(g.GMonoRet))
	}
	if g.GpBioDiferenca.QBCBioComb > 0 {
		e.Filho(w.gerarGpBioDiferencaIBS(g.GpBioDiferenca))
	}
	return e
}

func (w *Writer) gerarGCBSMonoAdRem(g GCBSMonoAdRem) *pcn.Elem {
	e := pcn.NovoElem("gCBSMonoAdRem")
	if g.GMonoPadrao.QBCMono > 0 {
		e.Filho(pcn.NovoElem("gMonoPadrao").
			Filho(pcn.NodeDec("qBCMono", g.GMonoPadrao.QBCMono, 4, false)).
			Filho(pcn.NodeDec("adRemCBS", g.GMonoPadrao.AdRemCBS, 4, true)).
			Filho(pcn.NodeDec("vCBSMono", g.GMonoPadrao.VCBSMono, 2, true)))
	}
	if g.GMonoReten.QBCMonoReten > 0 {
		e.Filho(pcn.NovoElem("gMonoReten").
			Filho(pcn.NodeDec("qBCMonoReten", g.GMonoReten.QBCMonoReten, 4, false)).
			Filho(pcn.NodeDec("adRemCBSReten", g.GMonoReten.AdRemCBSReten, 4, true)).
			Filho(pcn.NodeDec("vCBSMonoReten", g.GMonoReten.VCBSMonoReten, 2, true)))
	}
	if g.GMonoRet.VCBSMonoRet > 0 {
		e.Filho(w.gerarGMonoRetCBS(g.GMonoRet))
	}
	if g.GpBioDiferenca.QBCBioComb > 0 {
		e.Filho(w.gerarGpBioDiferencaCBS(g.GpBioDiferenca))
	}
	return e
}

func (w *Writer) gerarGMonoRetCBS(g GMonoRetCBS) *pcn.Elem {
	return pcn.NovoElem("gMonoRet").
		Filho(pcn.NodeDec("vCBSMonoRet", g.VCBSMonoRet, 2, true))
}

func (w *Writer) gerarGpBioDiferencaCBS(g GpBioDiferencaCBS) *pcn.Elem {
	return pcn.NovoElem("gpBioDiferenca").
		Filho(pcn.NodeDec("qBCBioComb", g.QBCBioComb, 4, true)).
		Filho(pcn.NodeDec("vCBSDiferenca", g.VCBSDiferenca, 2, true))
}

func (w *Writer) gerarGCBSMonoAdValorem(g GCBSMonoAdValorem) *pcn.Elem {
	e := pcn.NovoElem("gCBSMonoAdValorem")
	if g.GMonoPadrao.VBCMono > 0 {
		e.Filho(pcn.NovoElem("gMonoPadrao").
			Filho(pcn.NodeDec("vBCMono", g.GMonoPadrao.VBCMono, 2, false)).
			Filho(pcn.NodeDec("pAliqMonoCBS", g.GMonoPadrao.PAliqMonoCBS, 4, true)).
			Filho(pcn.NodeDec("vCBSMono", g.GMonoPadrao.VCBSMono, 2, true)))
	}
	if g.GMonoReten.VBCMonoReten > 0 {
		e.Filho(pcn.NovoElem("gMonoReten").
			Filho(pcn.NodeDec("vBCMonoReten", g.GMonoReten.VBCMonoReten, 2, false)).
			Filho(pcn.NodeDec("pAliqMonoReten", g.GMonoReten.PAliqMonoReten, 4, true)).
			Filho(pcn.NodeDec("vCBSMonoReten", g.GMonoReten.VCBSMonoReten, 2, true)))
	}
	if g.GMonoRet.VCBSMonoRet > 0 {
		e.Filho(w.gerarGMonoRetCBS(g.GMonoRet))
	}
	if g.GpBioDiferenca.QBCBioComb > 0 {
		e.Filho(w.gerarGpBioDiferencaCBS(g.GpBioDiferenca))
	}
	return e
}

func (w *Writer) gerarGTransfCred(g GTransfCred) *pcn.Elem {
	return pcn.NovoElem("gTransfCred").
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true))
}

func (w *Writer) gerarGCredPresIBSZFM(g CredPresIBSZFM) *pcn.Elem {
	return pcn.NovoElem("gCredPresIBSZFM").
		Filho(pcn.NodeStr("competApur", formatarCompetenciaXML(g.CompetApur), true)).
		Filho(pcn.NodeStr("tpCredPresIBSZFM", g.TpCredPresIBSZFM.String(), true)).
		Filho(pcn.NodeDec("vCredPresIBSZFM", g.VCredPresIBSZFM, 2, true))
}

func (w *Writer) gerarGAjusteCompet(g GAjusteCompet) *pcn.Elem {
	return pcn.NovoElem("gAjusteCompet").
		Filho(pcn.NodeStr("competApur", formatarCompetenciaXML(g.CompetApur), true)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true))
}

func (w *Writer) gerarGCredPresOper(g GCredPresOper) *pcn.Elem {
	e := pcn.NovoElem("gCredPresOper").
		Filho(pcn.NodeDec("vBCCredPres", g.VBCCredPres, 2, true)).
		Filho(pcn.NodeStr("cCredPres", g.CCredPres.String(), true))
	if g.GIBSCredPres.PCredPres > 0 {
		e.Filho(w.gerarGIBSCBSCredPres(g.GIBSCredPres, "gIBSCredPres"))
	}
	if g.GCBSCredPres.PCredPres > 0 {
		e.Filho(w.gerarGIBSCBSCredPres(g.GCBSCredPres, "gCBSCredPres"))
	}
	return e
}

func (w *Writer) gerarGIBSCBSCredPres(g GIBSCBSCredPres, grupo string) *pcn.Elem {
	e := pcn.NovoElem(grupo).
		Filho(pcn.NodeDec("pCredPres", g.PCredPres, 4, true))
	if g.VCredPresCondSus > 0 {
		e.Filho(pcn.NodeDec("vCredPresCondSus", g.VCredPresCondSus, 2, true))
	} else {
		e.Filho(pcn.NodeDec("vCredPres", g.VCredPres, 2, true))
	}
	return e
}

// GerarDFeReferenciado -- porte de Gerar_DFeReferenciado. O alerta de chave
// invalida do original (wAlerta) nao e portado; a chave vai como esta.
func (w *Writer) GerarDFeReferenciado(g DFeReferenciado) *pcn.Elem {
	if g.ChaveAcesso == "" {
		return nil
	}
	return pcn.NovoElem("DFeReferenciado").
		Filho(pcn.NodeStrSemFiltro("chaveAcesso", g.ChaveAcesso, true)).
		Filho(pcn.NodeInt("nItem", g.NItem, 1, false))
}

// GerarISTot -- porte de Gerar_ISTot.
func (w *Writer) GerarISTot(tot ISTot) *pcn.Elem {
	if tot.VIS <= 0 {
		return nil
	}
	return pcn.NovoElem("ISTot").
		Filho(pcn.NodeDec("vIS", tot.VIS, 2, true))
}

func (w *Writer) gerarGMonoTot(g GMono) *pcn.Elem {
	return pcn.NovoElem("gMono").
		Filho(pcn.NodeDec("vIBSMono", g.VIBSMono, 2, true)).
		Filho(pcn.NodeDec("vCBSMono", g.VCBSMono, 2, true)).
		Filho(pcn.NodeDec("vIBSMonoReten", g.VIBSMonoReten, 2, true)).
		Filho(pcn.NodeDec("vCBSMonoReten", g.VCBSMonoReten, 2, true)).
		Filho(pcn.NodeDec("vIBSMonoRet", g.VIBSMonoRet, 2, true)).
		Filho(pcn.NodeDec("vCBSMonoRet", g.VCBSMonoRet, 2, true))
}

// formatarCompetenciaXML formata AAAA-MM (FormatDateTime('yyyy-mm')).
// DIVERGENCIA determinística: data zero devolve vazio em vez do "1899-12"
// que o Delphi produziria com a data zero dele.
func formatarCompetenciaXML(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01")
}
