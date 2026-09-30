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
	"os"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Escrita do formato .ini do ACBr.
// Porte de TNFAgIniWriter (ACBrNFAg.IniWriter.pas).
//
// DOIS bugs deterministicos do original, REPLICADOS e testados:
//
//  1. a versao e GRAVADA em [infNFAg], mas o IniReader a LE de [infNFGas]
//     (copy-paste da NFGas) -- o round-trip do proprio ACBr perde a versao;
//  2. o grupo TFU e gravado na secao [TFSNNN] (Gerar_TFU usa o literal
//     'TFS'), enquanto o reader le TFU de [TFUNNN] -- TFU gravado nunca e
//     relido.

// GravarINI serializa a NFAg no formato .ini do ACBr.
// Como o original (GravarIni, IniWriter.pas:148-149), recusa nota com chave
// invalida -- o raise "Chave Invalida" vira erro.
func GravarINI(n *NFAg) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}
	if err := pcn.ValidarChaveAcesso(n.InfNFAg.ID); err != nil {
		return "", fmt.Errorf("nfag: NFAg inconsistente para gerar INI: %w", err)
	}
	ini := pcn.NovoINI()

	// bug 1 replicado: grava em [infNFAg]; o reader le de [infNFGas]
	ini.GravarString("infNFAg", "ID", n.InfNFAg.ID)
	v, _ := ParseVersaoNFAgFloat(n.InfNFAg.Versao)
	ini.GravarString("infNFAg", "versao", v.String())

	gravarINIIde(ini, n.Ide)
	gravarINIEmit(ini, n.Emit)
	gravarINIDest(ini, n.Dest)
	gravarINILigacao(ini, n.Ligacao)
	gravarINIGSub(ini, n.GSub)
	gravarINIGMed(ini, n.GMed)
	gravarINIGFatConjunto(ini, n.GFatConjunto)
	gravarINIDet(ini, n.Det)
	gravarINITotal(ini, n.Total)
	rtc.GravarINIPgtoVinc(ini, n.PgtoVinc)
	gravarINIGFat(ini, n.GFat)
	gravarINIGAgencia(ini, n.GAgencia)
	gravarINIGQualiAgua(ini, n.GQualiAgua)
	gravarINIAutXML(ini, n.AutXML)
	gravarINIInfPAA(ini, n.InfPAA)
	gravarINIInfAdic(ini, n.InfAdic)
	gravarINIInfRespTec(ini, n.InfRespTec)

	return ini.String(), nil
}

// GravarINIArquivo serializa e grava em disco.
func GravarINIArquivo(n *NFAg, caminho string) error {
	s, err := GravarINI(n)
	if err != nil {
		return err
	}
	if err := os.WriteFile(caminho, []byte(s), 0o644); err != nil {
		return fmt.Errorf("nfag: gravar %q: %w", caminho, err)
	}
	return nil
}

func gravarINIIde(ini *pcn.INI, ide Ide) {
	const secao = "ide"
	ini.GravarInteiro(secao, "cUF", ide.CUF)
	ini.GravarString(secao, "tpAmb", ide.TpAmb.String())
	ini.GravarInteiro(secao, "Modelo", ide.Modelo)
	ini.GravarInteiro(secao, "Serie", ide.Serie)
	ini.GravarInteiro(secao, "nNF", ide.NNF)
	ini.GravarInteiro(secao, "cNF", ide.CNF)
	// DateTimeToIni do original usa DateTimeToStr (data E hora, formato de
	// locale); aqui o formato e fixo, como no nfgas, e a hora e preservada.
	ini.GravarData(secao, "dhEmi", ide.DhEmi, "2006-01-02T15:04:05")
	ini.GravarString(secao, "tpEmis", ide.TpEmis.String())
	ini.GravarString(secao, "nSiteAutoriz", ide.NSiteAutoriz.String())
	ini.GravarInteiro(secao, "cMunFG", ide.CMunFG)
	ini.GravarString(secao, "finNFAg", ide.FinNFAg.String())
	ini.GravarString(secao, "tpFat", ide.TpFat.String())
	ini.GravarString(secao, "verProc", ide.VerProc)
	if !ide.DhCont.IsZero() {
		ini.GravarData(secao, "dhCont", ide.DhCont, "2006-01-02T15:04:05")
	}
	if ide.XJust != "" {
		ini.GravarString(secao, "xJust", ide.XJust)
	}
	if ide.TpPagAnt != pcn.TpaNenhum {
		ini.GravarString(secao, "tpPagAnt", ide.TpPagAnt.String())
	}
	rtc.GravarINIGCompraGovReduzido(ini, ide.GCompraGov)
}

func gravarINIEmit(ini *pcn.INI, e Emit) {
	const secao = "emit"
	ini.GravarString(secao, "CNPJ", e.CNPJ)
	ini.GravarString(secao, "IE", e.IE)
	ini.GravarString(secao, "xNome", e.XNome)
	ini.GravarString(secao, "xFant", e.XFant)
	ini.GravarString(secao, "ISUFEmit", e.ISUFEmit)
	gravarINIEndereco(ini, secao, e.EnderEmit)
}

func gravarINIEndereco(ini *pcn.INI, secao string, e Endereco) {
	ini.GravarString(secao, "xLgr", e.XLgr)
	ini.GravarString(secao, "nro", e.Nro)
	ini.GravarString(secao, "xCpl", e.XCpl)
	ini.GravarString(secao, "xBairro", e.XBairro)
	ini.GravarInteiro(secao, "cMun", e.CMun)
	ini.GravarString(secao, "xMun", e.XMun)
	ini.GravarInteiro(secao, "CEP", e.CEP)
	ini.GravarString(secao, "UF", e.UF)
	ini.GravarString(secao, "fone", e.Fone)
	ini.GravarString(secao, "email", e.Email)
}

func gravarINIDest(ini *pcn.INI, d Dest) {
	const secao = "dest"
	ini.GravarString(secao, "xNome", d.XNome)
	ini.GravarString(secao, "CNPJCPF", d.CNPJCPF)
	if d.IDOutros != "" {
		ini.GravarString(secao, "idOutros", d.IDOutros)
	}
	ini.GravarString(secao, "IE", d.IE)
	ini.GravarString(secao, "IM", d.IM)
	ini.GravarString(secao, "cNIS", d.CNIS)
	ini.GravarString(secao, "NB", d.NB)
	ini.GravarString(secao, "xNomeAdicional", d.XNomeAdicional)
	gravarINIEndereco(ini, secao, d.EnderDest)
}

func gravarINILigacao(ini *pcn.INI, l Ligacao) {
	const secao = "Ligacao"
	// Gerar_Ligacao (IniWriter.pas:269-270): sem idLigacao, a secao nao sai.
	if strings.TrimSpace(l.IDLigacao) == "" {
		return
	}
	ini.GravarString(secao, "idLigacao", l.IDLigacao)
	ini.GravarString(secao, "idCodCliente", l.IDCodCliente)
	ini.GravarString(secao, "tpLigacao", l.TpLigacao.String())
	ini.GravarString(secao, "latGPS", l.LatGPS)
	ini.GravarString(secao, "longGPS", l.LongGPS)
	ini.GravarString(secao, "codRoteiroLeitura", l.CodRoteiroLeitura)
}

func gravarINIGSub(ini *pcn.INI, g GSub) {
	if g.ChNFAg == "" {
		return
	}
	ini.GravarString("gSub", "chNFAg", g.ChNFAg)
	ini.GravarString("gSub", "motSub", g.MotSub.String())
}

func gravarINIGMed(ini *pcn.INI, gMed []GMed) {
	for i, m := range gMed {
		secao := fmt.Sprintf("gMed%02d", i+1)
		ini.GravarInteiro(secao, "nMed", m.NMed)
		ini.GravarString(secao, "idMedidor", m.IDMedidor)
		ini.GravarData(secao, "dMedAnt", m.DMedAnt, "02/01/2006")
		ini.GravarData(secao, "dMedAtu", m.DMedAtu, "02/01/2006")
	}
}

func gravarINIGFatConjunto(ini *pcn.INI, g GFatConjunto) {
	if g.ChNFAgFat == "" {
		return
	}
	ini.GravarString("gFatConjunto", "chNFAgFat", g.ChNFAgFat)
}

func gravarINIDet(ini *pcn.INI, dets []Det) {
	for i, d := range dets {
		idx := i + 1
		secao := fmt.Sprintf("det%03d", idx)
		ini.GravarInteiro(secao, "nItem", d.NItem)
		if d.ChNFAgAnt != "" {
			ini.GravarString(secao, "chNFAgAnt", d.ChNFAgAnt)
		}
		if d.NItemAnt > 0 {
			ini.GravarInteiro(secao, "nItemAnt", d.NItemAnt)
		}
		ini.GravarString(secao, "indOrigemQtd", d.Prod.IndOrigemQtd.String())
		ini.GravarString(secao, "cProd", d.Prod.CProd)
		ini.GravarString(secao, "xProd", d.Prod.XProd)
		ini.GravarString(secao, "cClass", d.Prod.CClass)
		ini.GravarString(secao, "tpCategoria", d.Prod.TpCategoria.String())
		ini.GravarString(secao, "xCategoria", d.Prod.XCategoria)
		ini.GravarString(secao, "qEconomias", d.Prod.QEconomias)
		ini.GravarString(secao, "uMed", d.Prod.UMed.String())
		ini.GravarFloat(secao, "qFaturada", d.Prod.QFaturada, 4)
		ini.GravarFloat(secao, "vItem", d.Prod.VItem, 10)
		ini.GravarFloat(secao, "fatorPoluicao", d.Prod.FatorPoluicao, 4)
		ini.GravarFloat(secao, "vProd", d.Prod.VProd, 10)
		if d.Prod.IndDevolucao != pcn.TieNenhum {
			ini.GravarString(secao, "indDevolucao", d.Prod.IndDevolucao.String())
		}
		if d.InfAdProd != "" {
			ini.GravarString(secao, "infAdProd", d.InfAdProd)
		}

		// DIVERGENCIA: o Gerar_gPagAntecipadoProd do original recebe Index
		// 0-based (IniWriter.pas:346) enquanto o reader le 1-based
		// (IniReader.pas:357) -- o round-trip do ACBr perde o grupo. Aqui o
		// writer usa o mesmo indice 1-based do reader.
		rtc.GravarINIGPagAntecipadoProd(ini, d.Prod.GPagAntecipado, idx, -1)
		gravarINIGMedicao(ini, d.Prod.GMedicao, idx)
		gravarINIGTarif(ini, d.GTarif, idx)
		gravarINIGProcRef(ini, d.GProcRef, idx)
		rtc.GravarINIIBSCBS(ini, d.Imposto.IBSCBS, idx, -1)
		gravarINIPIS(ini, d.Imposto.PIS, idx)
		gravarINICOFINS(ini, d.Imposto.COFINS, idx)
		gravarINIRetTrib(ini, d.Imposto.RetTrib, idx)
		gravarINITFS(ini, d.Imposto.TFS, idx)
		gravarINITFU(ini, d.Imposto.TFU, idx)
	}
}

func gravarINIGMedicao(ini *pcn.INI, g GMedicao, det int) {
	// Gerar_gMedicao (IniWriter.pas:363-364): so pula com nMed<=0 E vMed=0.
	if g.NMed <= 0 && g.GMedida.VMed == 0 {
		return
	}
	secao := fmt.Sprintf("gMedicao%03d", det)
	ini.GravarInteiro(secao, "nMed", g.NMed)
	ini.GravarString(secao, "tpMotNaoLeitura", g.TpMotNaoLeitura.String())
	ini.GravarString(secao, "tpGrMed", g.GMedida.TpGrMed.String())
	ini.GravarString(secao, "nUnidConsumo", g.GMedida.NUnidConsumo)
	ini.GravarFloat(secao, "vUnidConsumo", g.GMedida.VUnidConsumo, 2)
	ini.GravarString(secao, "uMed", g.GMedida.UMed.String())
	ini.GravarFloat(secao, "vMedAnt", g.GMedida.VMedAnt, 2)
	ini.GravarFloat(secao, "vMedAtu", g.GMedida.VMedAtu, 2)
	ini.GravarFloat(secao, "vConst", g.GMedida.VConst, 2)
	ini.GravarFloat(secao, "vMed", g.GMedida.VMed, 2)
}

func gravarINIGTarif(ini *pcn.INI, tarifas []GTarif, det int) {
	for i, t := range tarifas {
		secao := fmt.Sprintf("gTarif%03d%d", det, i+1)
		ini.GravarData(secao, "dIniTarif", t.DIniTarif, "02/01/2006")
		if !t.DFimTarif.IsZero() {
			ini.GravarData(secao, "dFimTarif", t.DFimTarif, "02/01/2006")
		}
		ini.GravarString(secao, "nAto", t.NAto)
		ini.GravarInteiro(secao, "anoAto", t.AnoAto)
		ini.GravarString(secao, "tpFaixaCons", t.TpFaixaCons.String())
	}
}

func gravarINIGProcRef(ini *pcn.INI, g GProcRef, det int) {
	if g.VItem == 0 && g.QFaturada == 0 && g.VProd == 0 && len(g.GProc) == 0 {
		return
	}
	secao := fmt.Sprintf("gProcRef%03d", det)
	ini.GravarFloat(secao, "vItem", g.VItem, 8)
	ini.GravarFloat(secao, "qFaturada", g.QFaturada, 4)
	ini.GravarFloat(secao, "vProd", g.VProd, 8)
	if g.IndDevolucao != pcn.TieNenhum {
		ini.GravarString(secao, "indDevolucao", g.IndDevolucao.String())
	}
	for i, p := range g.GProc {
		s := fmt.Sprintf("gProc%03d%02d", det, i+1)
		ini.GravarString(s, "tpProc", p.TpProc.String())
		ini.GravarString(s, "nProcesso", p.NProcesso)
	}
}

func gravarINIPIS(ini *pcn.INI, p PIS, det int) {
	if p.VBC == 0 && p.PPIS == 0 && p.VPIS == 0 {
		return
	}
	secao := fmt.Sprintf("PIS%03d", det)
	ini.GravarString(secao, "CST", p.CST.String())
	ini.GravarFloat(secao, "vBC", p.VBC, 2)
	ini.GravarFloat(secao, "pPIS", p.PPIS, 4)
	ini.GravarFloat(secao, "vPIS", p.VPIS, 2)
}

func gravarINICOFINS(ini *pcn.INI, c COFINS, det int) {
	if c.VBC == 0 && c.PCOFINS == 0 && c.VCOFINS == 0 {
		return
	}
	secao := fmt.Sprintf("COFINS%03d", det)
	ini.GravarString(secao, "CST", c.CST.String())
	ini.GravarFloat(secao, "vBC", c.VBC, 2)
	ini.GravarFloat(secao, "pCOFINS", c.PCOFINS, 4)
	ini.GravarFloat(secao, "vCOFINS", c.VCOFINS, 2)
}

func gravarINIRetTrib(ini *pcn.INI, r RetTrib, det int) {
	if r.VRetPIS == 0 && r.VRetCOFINS == 0 && r.VRetCSLL == 0 && r.VIRRF == 0 {
		return
	}
	secao := fmt.Sprintf("retTrib%03d", det)
	ini.GravarFloat(secao, "vRetPIS", r.VRetPIS, 2)
	ini.GravarFloat(secao, "vRetCOFINS", r.VRetCOFINS, 2)
	ini.GravarFloat(secao, "vRetCSLL", r.VRetCSLL, 2)
	ini.GravarFloat(secao, "vIRRF", r.VIRRF, 2)
}

func gravarINITFS(ini *pcn.INI, t TFS, det int) {
	if t.VBCTFS == 0 && t.PTFS == 0 && t.VTFS == 0 {
		return
	}
	secao := fmt.Sprintf("TFS%03d", det)
	ini.GravarFloat(secao, "vBCTFS", t.VBCTFS, 2)
	ini.GravarFloat(secao, "pTFS", t.PTFS, 2)
	ini.GravarFloat(secao, "vTFS", t.VTFS, 2)
}

// gravarINITFU: BUG DO ACBr REPLICADO (bug 2 do cabecalho) -- o Gerar_TFU
// grava na secao 'TFS'+NNN, e o reader le TFU de 'TFU'+NNN.
func gravarINITFU(ini *pcn.INI, t TFU, det int) {
	if t.VBCTFU == 0 && t.PTFU == 0 && t.VTFU == 0 {
		return
	}
	secao := fmt.Sprintf("TFS%03d", det)
	ini.GravarFloat(secao, "vBCTFU", t.VBCTFU, 2)
	ini.GravarFloat(secao, "pTFU", t.PTFU, 2)
	ini.GravarFloat(secao, "vTFU", t.VTFU, 2)
}

func gravarINITotal(ini *pcn.INI, t Total) {
	const secao = "total"
	ini.GravarFloat(secao, "vProd", t.VProd, 2)
	ini.GravarFloat(secao, "vRetPIS", t.VRetPIS, 2)
	ini.GravarFloat(secao, "vRetCOFINS", t.VRetCOFINS, 2)
	ini.GravarFloat(secao, "vRetCSLL", t.VRetCSLL, 2)
	ini.GravarFloat(secao, "vIRRF", t.VIRRF, 2)
	ini.GravarFloat(secao, "vCOFINS", t.VCOFINS, 2)
	ini.GravarFloat(secao, "vPIS", t.VPIS, 2)
	ini.GravarFloat(secao, "vTFS", t.VTFS, 2)
	ini.GravarFloat(secao, "vTFU", t.VTFU, 2)
	ini.GravarFloat(secao, "vNF", t.VNF, 2)
	ini.GravarFloat(secao, "vTotDFe", t.VTotDFe, 2)

	rtc.GravarINIIBSCBSTot(ini, t.IBSCBSTot)
}

func gravarINIGFat(ini *pcn.INI, g GFat) {
	const secao = "gFat"
	ini.GravarData(secao, "CompetFat", g.CompetFat, "02/01/2006")
	ini.GravarData(secao, "dVencFat", g.DVencFat, "02/01/2006")
	ini.GravarData(secao, "dApresFat", g.DApresFat, "02/01/2006")
	ini.GravarData(secao, "dProxLeitura", g.DProxLeitura, "02/01/2006")
	ini.GravarString(secao, "nFat", g.NFat)
	ini.GravarString(secao, "codBarras", g.CodBarras)
	ini.GravarString(secao, "codDebAuto", g.CodDebAuto)
	ini.GravarString(secao, "codBanco", g.CodBanco)
	ini.GravarString(secao, "codAgencia", g.CodAgencia)

	if g.EnderCorresp.XLgr != "" {
		gravarINIEndereco(ini, "enderCorresp", g.EnderCorresp)
	}
	if g.GPIX.URLQRCodePIX != "" {
		ini.GravarString("gPIX", "urlQRCodePIX", g.GPIX.URLQRCodePIX)
	}
}

func gravarINIGAgencia(ini *pcn.INI, g GAgencia) {
	const secao = "gAgencia"
	ini.GravarString(secao, "econ", g.Econ)
	ini.GravarString(secao, "econAcumulada", g.EconAcumulada)
	ini.GravarString(secao, "sPrestador", g.SPrestador)
	if !g.DEmissSelo.IsZero() {
		ini.GravarData(secao, "dEmissSelo", g.DEmissSelo, "02/01/2006")
	}
	ini.GravarString(secao, "sRegulador", g.SRegulador)
	ini.GravarString(secao, "nAgenciaAtend", g.NAgenciaAtend)
	ini.GravarString(secao, "enderAgenciaAtend", g.EnderAgenciaAtend)

	for h, hist := range g.GHistCons {
		secaoH := fmt.Sprintf("gHistCons%d", h+1)
		ini.GravarString(secaoH, "xHistorico", hist.XHistorico)
		ini.GravarFloat(secaoH, "medMensal", hist.MedMensal, 4)
		for c, gc := range hist.GCons {
			secaoC := fmt.Sprintf("gCons%d%02d", h+1, c+1)
			ini.GravarData(secaoC, "CompetFat", gc.CompetFat, "02/01/2006")
			ini.GravarString(secaoC, "uMed", gc.UMed.String())
			ini.GravarString(secaoC, "qtdDias", gc.QtdDias)
			ini.GravarFloat(secaoC, "medDiaria", gc.MedDiaria, 4)
			ini.GravarFloat(secaoC, "consumo", gc.Consumo, 4)
			ini.GravarFloat(secaoC, "volFat", gc.VolFat, 2)
		}
	}
}

func gravarINIGQualiAgua(ini *pcn.INI, g GQualiAgua) {
	if g.CompetAnalise.IsZero() {
		return
	}
	const secao = "gQualiAgua"
	ini.GravarData(secao, "CompetAnalise", g.CompetAnalise, "02/01/2006")
	ini.GravarString(secao, "Conclusao", g.Conclusao)
	ini.GravarString(secao, "cProcesso", g.CProcesso)
	ini.GravarString(secao, "SistemaAbast", g.SistemaAbast)

	for i, a := range g.GAnalise {
		s := fmt.Sprintf("gAnalise%02d", i+1)
		ini.GravarString(s, "xItemAnalisado", a.XItemAnalisado)
		ini.GravarString(s, "nAmostraMinima", a.NAmostraMinima)
		ini.GravarString(s, "nAmostraAnalisada", a.NAmostraAnalisada)
		ini.GravarString(s, "nAmostraFPadrao", a.NAmostraFPadrao)
		ini.GravarString(s, "nAmostraDPadrao", a.NAmostraDPadrao)
		ini.GravarString(s, "nMediaMensal", a.NMediaMensal)
		ini.GravarString(s, "xValorReferencia", a.XValorReferencia)
	}
}

func gravarINIAutXML(ini *pcn.INI, aut []AutXML) {
	for i, a := range aut {
		secao := fmt.Sprintf("autXML%02d", i+1)
		ini.GravarString(secao, "CNPJCPF", a.CNPJCPF)
	}
}

func gravarINIInfPAA(ini *pcn.INI, p InfPAA) {
	if p.CNPJPAA == "" {
		return
	}
	ini.GravarString("infPAA", "CNPJPAA", p.CNPJPAA)
}

func gravarINIInfAdic(ini *pcn.INI, i InfAdic) {
	infCpl := ""
	if len(i.InfCpl) > 0 {
		infCpl = i.InfCpl[0]
	}
	if i.InfAdFisco == "" && infCpl == "" {
		return
	}
	ini.GravarString("infAdic", "infAdFisco", i.InfAdFisco)
	ini.GravarString("infAdic", "infCpl", infCpl)
}

func gravarINIInfRespTec(ini *pcn.INI, r InfRespTec) {
	if r.CNPJ == "" {
		return
	}
	const secao = "infRespTec"
	ini.GravarString(secao, "CNPJ", r.CNPJ)
	ini.GravarString(secao, "xContato", r.XContato)
	ini.GravarString(secao, "email", r.Email)
	ini.GravarString(secao, "fone", r.Fone)
	if r.IDCSRT != 0 {
		ini.GravarInteiro(secao, "idCSRT", r.IDCSRT)
	}
	if r.HashCSRT != "" {
		ini.GravarString(secao, "hashCSRT", r.HashCSRT)
	}
}
