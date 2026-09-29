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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Escrita do formato .ini do ACBr.
// Porte de TNFGasIniWriter (ACBrNFGas.IniWriter.pas).
//
// Simetrico a ini_reader.go: o que o leitor le, o gravador escreve, na
// mesma secao e na mesma chave.
//
// DIVERGENCIA de formato, em dois pontos, os dois pela mesma razao -- o
// original depende da configuracao regional da maquina e por isso gera
// arquivo diferente em cada instalacao:
//
//   - data: o ACBr usa DateTimeToStr; aqui o formato e ISO
//     ("2026-03-15T10:30:00" com hora, "2026-03-15" sem);
//   - decimal: o ACBr usa o separador do sistema; aqui e sempre ponto.
//
// A leitura aceita as duas formas, entao o .ini gerado pelo Delphi continua
// sendo lido aqui sem ajuste.
//
// ASSIMETRIA no sentido inverso (Go -> Delphi): as competencias
// (CompetEmis, CompetApur, CompetFat) saem como AAAAMM -- a forma do XML,
// que este leitor e o EncodeDataHora aceitam --, mas o StringToDateTime do
// Delphi nao interpreta AAAAMM. Um .ini gerado aqui e lido pelo ACBr perde
// essas tres chaves (ficam em zero la). As demais datas saem em ISO, que o
// Delphi le.

const (
	layoutDataHoraINI = "2006-01-02T15:04:05"
	layoutDataINI     = "2006-01-02"
)

// GravarINI serializa a NFGas no formato .ini do ACBr.
func GravarINI(n *NFGas) (string, error) {
	ini, err := MontarINI(n)
	if err != nil {
		return "", err
	}
	return ini.String(), nil
}

// GravarINIArquivo serializa a NFGas e grava no caminho informado.
func GravarINIArquivo(n *NFGas, caminho string) error {
	ini, err := MontarINI(n)
	if err != nil {
		return err
	}
	return ini.SalvarArquivo(caminho)
}

// MontarINI devolve a NFGas ja montada na estrutura de INI, para quem
// precisar acrescentar secoes proprias antes de serializar.
func MontarINI(n *NFGas) (*pcn.INI, error) {
	if n == nil {
		return nil, ErrXMLVazio
	}
	ini := pcn.NovoINI()

	ini.GravarFloat("infNFGas", "versao", n.InfNFGas.Versao, 2)

	gravarINIIde(ini, n.Ide)
	gravarINIEmit(ini, n.Emit)
	gravarINIDest(ini, n.Dest)
	gravarINIInstalacao(ini, n.Instalacao)
	gravarINIGSub(ini, n.GSub)
	gravarINIGVolContrat(ini, n.GVolContrat)
	gravarINIGMed(ini, n.GMed)
	gravarINIDet(ini, n.Det)
	gravarINITotal(ini, n.Total)
	rtc.GravarINIPgtoVinc(ini, n.PgtoVinc)
	gravarINIGFat(ini, n.GFat)
	gravarINIGAgencia(ini, n.GAgencia)
	gravarINIAutXML(ini, n.AutXML)
	gravarINIInfAdic(ini, n.InfAdic)
	gravarINIInfRespTec(ini, n.InfRespTec)

	return ini, nil
}

func gravarData(ini *pcn.INI, secao, chave string, t time.Time, comHora bool) {
	if t.IsZero() {
		return
	}
	layout := layoutDataINI
	if comHora {
		layout = layoutDataHoraINI
	}
	ini.GravarString(secao, chave, t.Format(layout))
}

func gravarINIIde(ini *pcn.INI, ide Ide) {
	const secao = "ide"

	ini.GravarInteiro(secao, "cUF", ide.CUF)
	ini.GravarString(secao, "tpAmb", ide.TpAmb.String())
	ini.GravarInteiro(secao, "Modelo", ide.Modelo)
	ini.GravarInteiro(secao, "Serie", ide.Serie)
	ini.GravarInteiro(secao, "nNF", ide.NNF)
	ini.GravarInteiro(secao, "cNF", ide.CNF)
	gravarData(ini, secao, "dhEmi", ide.DhEmi, true)
	ini.GravarString(secao, "tpEmis", ide.TpEmis.String())
	ini.GravarString(secao, "nSiteAutoriz", ide.NSiteAutoriz.String())
	ini.GravarInteiro(secao, "cMunFG", ide.CMunFG)
	ini.GravarString(secao, "finNFGas", ide.FinNFGas.String())
	ini.GravarString(secao, "tpFat", ide.TpFat.String())
	ini.GravarString(secao, "verProc", ide.VerProc)
	gravarData(ini, secao, "dhCont", ide.DhCont, true)
	if ide.XJust != "" {
		ini.GravarString(secao, "xJust", ide.XJust)
	}
	if ide.TpPagAnt != pcn.TpaNenhum {
		ini.GravarString(secao, "tpPagAnt", ide.TpPagAnt.String())
	}

	rtc.GravarINIGCompraGovReduzido(ini, ide.GCompraGov)
}

func gravarINIEndereco(ini *pcn.INI, secao string, e Endereco, comPais bool) {
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
	if comPais {
		ini.GravarInteiro(secao, "cPais", e.CPais)
		ini.GravarString(secao, "xPais", e.XPais)
	}
}

func gravarINIEmit(ini *pcn.INI, e Emit) {
	const secao = "emit"
	ini.GravarString(secao, "CNPJ", e.CNPJ)
	ini.GravarString(secao, "IE", e.IE)
	ini.GravarString(secao, "xNome", e.XNome)
	ini.GravarString(secao, "xFant", e.XFant)
	if e.ISUFEmit != "" {
		ini.GravarString(secao, "ISUFEmit", e.ISUFEmit)
	}
	gravarINIEndereco(ini, secao, e.EnderEmit, true)
}

func gravarINIDest(ini *pcn.INI, d Dest) {
	const secao = "dest"
	if d.XNome == "" && d.CNPJCPF == "" && d.IDEstrangeiro == "" {
		return
	}

	ini.GravarString(secao, "xNome", d.XNome)
	ini.GravarString(secao, "CNPJCPF", d.CNPJCPF)
	if d.IDEstrangeiro != "" {
		chave := d.TagIDOrigem
		if chave == "" {
			chave = "idOutros"
		}
		ini.GravarString(secao, chave, d.IDEstrangeiro)
	}
	ini.GravarString(secao, "indIEDest", d.IndIEDest.String())
	ini.GravarString(secao, "IE", d.IE)
	ini.GravarString(secao, "IM", d.IM)
	ini.GravarString(secao, "cNIS", d.CNIS)
	ini.GravarString(secao, "NB", d.NB)
	ini.GravarString(secao, "xNomeAdicional", d.XNomeAdicional)

	gravarINIEndereco(ini, secao, d.EnderDest, true)
}

func gravarINIInstalacao(ini *pcn.INI, i Instalacao) {
	const secao = "Instalacao"
	if i == (Instalacao{}) {
		return
	}
	ini.GravarString(secao, "idInstalacao", i.IDInstalacao)
	ini.GravarString(secao, "idCodCliente", i.IDCodCliente)
	ini.GravarString(secao, "tpInstalacao", i.TpInstalacao.String())
	ini.GravarString(secao, "nContrato", i.NContrato)
	ini.GravarString(secao, "tpClasse", i.TpClasse.String())
	ini.GravarString(secao, "xClasse", i.XClasse)
	ini.GravarString(secao, "latGPS", i.LatGPS)
	ini.GravarString(secao, "longGPS", i.LongGPS)
	ini.GravarString(secao, "codRoteiroLeitura", i.CodRoteiroLeitura)
}

func gravarINIGSub(ini *pcn.INI, g GSub) {
	const secao = "gSub"
	if g.ChNFGas == "" && g.GNF.CNPJ == "" {
		return
	}
	ini.GravarString(secao, "chNFGas", g.ChNFGas)
	ini.GravarString(secao, "motSub", g.MotSub.String())
	ini.GravarString(secao, "CNPJ", g.GNF.CNPJ)
	ini.GravarString(secao, "Serie", g.GNF.Serie)
	ini.GravarInteiro(secao, "nNF", g.GNF.NNF)
	if s := pcn.FormatarCompetencia(g.GNF.CompetEmis); s != "" {
		ini.GravarString(secao, "CompetEmis", s)
	}
	if s := pcn.FormatarCompetencia(g.GNF.CompetApur); s != "" {
		ini.GravarString(secao, "CompetApur", s)
	}
	ini.GravarString(secao, "hash115", g.GNF.Hash115)
}

func gravarINIGVolContrat(ini *pcn.INI, lista []GVolContrat) {
	for i, v := range lista {
		secao := fmt.Sprintf("gVolContrat%02d", i+1)
		ini.GravarInteiro(secao, "nContrat", v.NContrat)
		ini.GravarString(secao, "tpVolContrat", v.TpVolContrat.String())
		ini.GravarFloat(secao, "qUnidContrat", v.QUnidContrat, 6)
	}
}

func gravarINIGMed(ini *pcn.INI, lista []GMed) {
	for i, v := range lista {
		secao := fmt.Sprintf("gMed%02d", i+1)
		ini.GravarInteiro(secao, "nMed", v.NMed)
		ini.GravarString(secao, "idEqp", v.IDEqp)
		gravarData(ini, secao, "dMedAnt", v.DMedAnt, false)
		ini.GravarFloat(secao, "vMedAnt", v.VMedAnt, 4)
		gravarData(ini, secao, "dMedAtu", v.DMedAtu, false)
		ini.GravarFloat(secao, "vMedAtu", v.VMedAtu, 4)
		ini.GravarString(secao, "tpEqp", v.TpEqp.String())
		ini.GravarString(secao, "tpMedidor", v.TpMedidor.String())
	}
}

func gravarINIDet(ini *pcn.INI, lista []Det) {
	for i, d := range lista {
		idx := i + 1
		secao := fmt.Sprintf("det%03d", idx)

		ini.GravarInteiro(secao, "nItem", d.NItem)
		if d.ChNFGasAnt != "" {
			ini.GravarString(secao, "chNFGasAnt", d.ChNFGasAnt)
		}
		if d.NItemAnt != 0 {
			ini.GravarInteiro(secao, "nItemAnt", d.NItemAnt)
		}

		ini.GravarString(secao, "orig", d.GNormal.Imposto.Orig.String())
		ini.GravarString(secao, "indSemCST", d.GNormal.Imposto.IndSemCST.String())

		p := d.GNormal.Prod
		ini.GravarString(secao, "indOrigemQtd", p.IndOrigemQtd.String())
		ini.GravarString(secao, "cProd", p.CProd)
		ini.GravarString(secao, "xProd", p.XProd)
		ini.GravarString(secao, "cClass", p.CClass)
		ini.GravarInteiro(secao, "CFOP", p.CFOP)
		ini.GravarString(secao, "uMed", p.UMed.String())
		ini.GravarFloat(secao, "qFaturada", p.QFaturada, 4)
		ini.GravarFloat(secao, "vItem", p.VItem, 10)
		ini.GravarFloat(secao, "fatorPCS", p.FatorPCS, 4)
		ini.GravarFloat(secao, "fatorPTZ", p.FatorPTZ, 4)
		ini.GravarFloat(secao, "fatorP", p.FatorP, 4)
		ini.GravarFloat(secao, "fatorT", p.FatorT, 4)
		ini.GravarFloat(secao, "vProd", p.VProd, 10)
		ini.GravarString(secao, "indDevolucao", p.IndDevolucao.String())
		if d.GNormal.InfAdProd != "" {
			ini.GravarString(secao, "infAdProd", d.GNormal.InfAdProd)
		}

		gravarINIGMedicao(ini, p.GMedicao, idx)
		rtc.GravarINIGPagAntecipadoProd(ini, p.GPagAntecipado, idx, rtc.SemIndice)

		gravarINIGTarif(ini, d.GNormal.GTarif, idx)
		gravarINIGAgregadora(ini, d.GAgregadora, idx)
		gravarINIGProcRef(ini, d.GNormal.GProcRef, idx)

		gravarINIICMS(ini, d.GNormal.Imposto.ICMS, idx)
		rtc.GravarINIIBSCBS(ini, d.GNormal.Imposto.IBSCBS, idx, rtc.SemIndice)
		gravarINIPIS(ini, d.GNormal.Imposto.PIS, idx)
		gravarINICOFINS(ini, d.GNormal.Imposto.COFINS, idx)
		gravarINIRetTrib(ini, d.GNormal.Imposto.RetTrib, idx)
		gravarINITxReg(ini, d.GNormal.Imposto.TxReg, idx)
	}
}

func gravarINIGMedicao(ini *pcn.INI, g GMedicao, det int) {
	if g == (GMedicao{}) {
		return
	}
	secao := fmt.Sprintf("gMedicao%03d", det)
	ini.GravarInteiro(secao, "nMed", g.NMed)
	ini.GravarInteiro(secao, "nContrat", g.NContrat)
	ini.GravarString(secao, "uMed", g.GMedida.UMed.String())
	ini.GravarFloat(secao, "vMed", g.GMedida.VMed, 4)
	if g.XMotNaoLeitura != "" {
		ini.GravarString(secao, "tpMotNaoLeitura", g.TpMotNaoLeitura.String())
		ini.GravarString(secao, "xMotNaoLeitura", g.XMotNaoLeitura)
	}
}

func gravarINIGTarif(ini *pcn.INI, lista []GTarif, det int) {
	for i, t := range lista {
		secao := fmt.Sprintf("gTarif%03d%d", det, i+1)
		gravarData(ini, secao, "dIniTarif", t.DIniTarif, false)
		gravarData(ini, secao, "dFimTarif", t.DFimTarif, false)
		ini.GravarString(secao, "nAto", t.NAto)
		ini.GravarInteiro(secao, "anoAto", t.AnoAto)
		ini.GravarString(secao, "tpFaixaCons", t.TpFaixaCons.String())
		ini.GravarFloat(secao, "vTarifAplic", t.VTarifAplic, 8)
	}
}

func gravarINIGAgregadora(ini *pcn.INI, g GAgregadora, det int) {
	if g == (GAgregadora{}) {
		return
	}
	secao := fmt.Sprintf("gAgregadora%03d", det)
	ini.GravarString(secao, "cClass", g.CClass)
	ini.GravarFloat(secao, "vTotDFe", g.VTotDFe, 2)
}

func gravarINIGProcRef(ini *pcn.INI, g GProcRef, det int) {
	if g.VItem == 0 && g.VProd == 0 && len(g.GProc) == 0 {
		return
	}
	secao := fmt.Sprintf("gProcRef%03d", det)
	ini.GravarFloat(secao, "vItem", g.VItem, 8)
	ini.GravarFloat(secao, "qFaturada", g.QFaturada, 4)
	ini.GravarFloat(secao, "vProd", g.VProd, 8)
	ini.GravarString(secao, "indDevolucao", g.IndDevolucao.String())

	for i, p := range g.GProc {
		s := fmt.Sprintf("gProc%03d%02d", det, i+1)
		ini.GravarString(s, "tpProc", p.TpProc.String())
		ini.GravarString(s, "nProcesso", p.NProcesso)
	}
}

func gravarINIICMS(ini *pcn.INI, icms ICMS, det int) {
	secao := fmt.Sprintf("ICMS%03d", det)

	ini.GravarString(secao, "indSemCST", icms.IndSemCST.String())
	ini.GravarString(secao, "CST", icms.CST.String())
	ini.GravarString(secao, "modBC", icms.ModBC.String())
	ini.GravarFloat(secao, "pRedBC", icms.PRedBC, 2)
	ini.GravarFloat(secao, "vBC", icms.VBC, 2)
	ini.GravarFloat(secao, "pICMS", icms.PICMS, 2)
	ini.GravarFloat(secao, "vICMS", icms.VICMS, 2)
	ini.GravarFloat(secao, "vICMSDeson", icms.VICMSDeson, 2)
	ini.GravarString(secao, "motDesICMS", icms.MotDesICMS.String())
	ini.GravarString(secao, "indDeduzDeson", icms.IndDeduzDeson.String())
	ini.GravarString(secao, "cBenef", icms.CBenef)
	ini.GravarString(secao, "modBCST", icms.ModBCST.String())
	ini.GravarFloat(secao, "pMVAST", icms.PMVAST, 2)
	ini.GravarFloat(secao, "pRedBCST", icms.PRedBCST, 2)
	ini.GravarFloat(secao, "vBCST", icms.VBCST, 2)
	ini.GravarFloat(secao, "pICMSST", icms.PICMSST, 2)
	ini.GravarFloat(secao, "vICMSST", icms.VICMSST, 2)
	ini.GravarFloat(secao, "vBCFCP", icms.VBCFCP, 2)
	ini.GravarFloat(secao, "pFCPST", icms.PFCPST, 2)
	ini.GravarFloat(secao, "vFCPST", icms.VFCPST, 2)
	ini.GravarFloat(secao, "vBCFCPST", icms.VBCFCPST, 2)
	ini.GravarFloat(secao, "vBCSTRet", icms.VBCSTRet, 2)
	ini.GravarFloat(secao, "pICMSSTRet", icms.PICMSSTRet, 2)
	ini.GravarFloat(secao, "vICMSSubstituto", icms.VICMSSubstituto, 2)
	ini.GravarFloat(secao, "vICMSSTRet", icms.VICMSSTRet, 2)
	ini.GravarFloat(secao, "vBCFCPSTRet", icms.VBCFCPSTRet, 2)
	ini.GravarFloat(secao, "pFCPSTRet", icms.PFCPSTRet, 2)
	ini.GravarFloat(secao, "vFCPSTRet", icms.VFCPSTRet, 2)
	ini.GravarFloat(secao, "pRedBCEfet", icms.PRedBCEfet, 2)
	ini.GravarFloat(secao, "vBCEfet", icms.VBCEfet, 2)
	ini.GravarFloat(secao, "pICMSEfet", icms.PICMSEfet, 2)
	ini.GravarFloat(secao, "vICMSEfet", icms.VICMSEfet, 2)
	ini.GravarFloat(secao, "pFCP", icms.PFCP, 4)
	ini.GravarFloat(secao, "vFCP", icms.VFCP, 2)
}

func gravarINIPIS(ini *pcn.INI, p PIS, det int) {
	if p == (PIS{}) {
		return
	}
	secao := fmt.Sprintf("PIS%03d", det)
	ini.GravarString(secao, "CST", p.CST.String())
	ini.GravarFloat(secao, "vBC", p.VBC, 2)
	ini.GravarFloat(secao, "pPIS", p.PPIS, 4)
	ini.GravarFloat(secao, "vPIS", p.VPIS, 2)
}

func gravarINICOFINS(ini *pcn.INI, c COFINS, det int) {
	if c == (COFINS{}) {
		return
	}
	secao := fmt.Sprintf("COFINS%03d", det)
	ini.GravarString(secao, "CST", c.CST.String())
	ini.GravarFloat(secao, "vBC", c.VBC, 2)
	ini.GravarFloat(secao, "pCOFINS", c.PCOFINS, 4)
	ini.GravarFloat(secao, "vCOFINS", c.VCOFINS, 2)
}

// gravarINIRetTrib grava [retTribNNN]. A chave de COFINS e vRetCOFINS em
// caixa alta -- diferente da tag XML, que e vRetCofins.
func gravarINIRetTrib(ini *pcn.INI, r RetTrib, det int) {
	if r == (RetTrib{}) {
		return
	}
	secao := fmt.Sprintf("retTrib%03d", det)
	ini.GravarFloat(secao, "vRetPIS", r.VRetPIS, 2)
	ini.GravarFloat(secao, "vRetCOFINS", r.VRetCOFINS, 2)
	ini.GravarFloat(secao, "vRetCSLL", r.VRetCSLL, 2)
	ini.GravarFloat(secao, "vIRRF", r.VIRRF, 2)
}

func gravarINITxReg(ini *pcn.INI, t TxReg, det int) {
	if t == (TxReg{}) {
		return
	}
	secao := fmt.Sprintf("TxReg%03d", det)
	ini.GravarFloat(secao, "vBC", t.VBC, 2)
	ini.GravarFloat(secao, "pTaxa", t.PTaxa, 4)
	ini.GravarFloat(secao, "vTaxa", t.VTaxa, 2)
}

// gravarINITotal grava [total], que no .ini ja e achatado -- nao ha
// subsecoes ICMSTot nem vRetTribTot.
//
// vST NAO e gravado: a chave nao existe no .ini do ACBr, embora a tag
// exista no XML. Replicado.
func gravarINITotal(ini *pcn.INI, t Total) {
	const secao = "total"

	ini.GravarFloat(secao, "vProd", t.VProd, 2)
	ini.GravarFloat(secao, "vBC", t.VBC, 2)
	ini.GravarFloat(secao, "vICMS", t.VICMS, 2)
	ini.GravarFloat(secao, "vBCST", t.VBCST, 2)
	ini.GravarFloat(secao, "vFCPST", t.VFCPST, 2)
	ini.GravarFloat(secao, "vICMSDeson", t.VICMSDeson, 2)
	ini.GravarFloat(secao, "vFCP", t.VFCP, 2)
	ini.GravarFloat(secao, "vCOFINS", t.VCOFINS, 2)
	ini.GravarFloat(secao, "vPIS", t.VPIS, 2)
	ini.GravarFloat(secao, "vTxReg", t.VTxReg, 2)
	ini.GravarFloat(secao, "vRetPIS", t.VRetPIS, 2)
	ini.GravarFloat(secao, "vRetCOFINS", t.VRetCOFINS, 2)
	ini.GravarFloat(secao, "vRetCSLL", t.VRetCSLL, 2)
	ini.GravarFloat(secao, "vIRRF", t.VIRRF, 2)
	ini.GravarFloat(secao, "vNF", t.VNF, 2)
	ini.GravarFloat(secao, "vTotDFe", t.VTotDFe, 2)

	rtc.GravarINIIBSCBSTot(ini, t.IBSCBSTot)
}

func gravarINIGFat(ini *pcn.INI, g GFat) {
	const secao = "gFat"
	if g.CompetFat.IsZero() && g.DVencFat.IsZero() && g.NFat == "" && g.CodBarras == "" {
		return
	}

	if s := pcn.FormatarCompetencia(g.CompetFat); s != "" {
		ini.GravarString(secao, "CompetFat", s)
	}
	gravarData(ini, secao, "dVencFat", g.DVencFat, false)
	gravarData(ini, secao, "dApresFat", g.DApresFat, false)
	gravarData(ini, secao, "dProxLeitura", g.DProxLeitura, false)
	ini.GravarString(secao, "nFat", g.NFat)
	ini.GravarString(secao, "codBarras", g.CodBarras)
	ini.GravarString(secao, "codDebAuto", g.CodDebAuto)
	ini.GravarString(secao, "codBanco", g.CodBanco)
	ini.GravarString(secao, "codAgencia", g.CodAgencia)
	if g.InfAdFat != "" {
		ini.GravarString(secao, "infAdFat", g.InfAdFat)
	}

	if g.EnderCorresp != (Endereco{}) {
		gravarINIEndereco(ini, "enderCorresp", g.EnderCorresp, false)
	}
	if g.GPIX.URLQRCodePIX != "" {
		ini.GravarString("gPIX", "urlQRCodePIX", g.GPIX.URLQRCodePIX)
	}
}

func gravarINIGAgencia(ini *pcn.INI, g GAgencia) {
	const secao = "gAgencia"
	if g.NomeAgenciaAtend == "" && len(g.GHistCons) == 0 {
		return
	}

	ini.GravarString(secao, "nomeAgenciaAtend", g.NomeAgenciaAtend)
	ini.GravarString(secao, "enderAgenciaAtend", g.EnderAgenciaAtend)
	ini.GravarString(secao, "sitioAgenciaAtend", g.SitioAgenciaAtend)
	if g.InfAdReg != "" {
		ini.GravarString(secao, "infAdReg", g.InfAdReg)
	}

	for h, hist := range g.GHistCons {
		sh := fmt.Sprintf("gHistCons%d", h+1)
		ini.GravarString(sh, "xHistorico", hist.XHistorico)
		ini.GravarFloat(sh, "medMensal", hist.MedMensal, 4)

		for c, cons := range hist.GCons {
			sc := fmt.Sprintf("gCons%d%02d", h+1, c+1)
			if s := pcn.FormatarCompetencia(cons.CompetFat); s != "" {
				ini.GravarString(sc, "CompetFat", s)
			}
			ini.GravarString(sc, "uMed", cons.UMed.String())
			ini.GravarInteiro(sc, "qtdDias", cons.QtdDias)
			ini.GravarFloat(sc, "medDiaria", cons.MedDiaria, 4)
			ini.GravarFloat(sc, "consumo", cons.Consumo, 4)
			ini.GravarFloat(sc, "vFat", cons.VFat, 4)
		}
	}
}

func gravarINIAutXML(ini *pcn.INI, lista []AutXML) {
	for i, a := range lista {
		ini.GravarString(fmt.Sprintf("autXML%02d", i+1), "CNPJCPF", a.CNPJCPF)
	}
}

func gravarINIInfAdic(ini *pcn.INI, i InfAdic) {
	const secao = "infAdic"
	if i.InfAdFisco == "" && len(i.InfCpl) == 0 {
		return
	}
	ini.GravarString(secao, "infAdFisco", i.InfAdFisco)
	// So a primeira ocorrencia e gravada, espelhando a leitura.
	if len(i.InfCpl) > 0 {
		ini.GravarString(secao, "infCpl", i.InfCpl[0])
	}
}

func gravarINIInfRespTec(ini *pcn.INI, r InfRespTec) {
	const secao = "infRespTec"
	if r == (InfRespTec{}) {
		return
	}
	ini.GravarString(secao, "CNPJ", r.CNPJ)
	ini.GravarString(secao, "xContato", r.XContato)
	ini.GravarString(secao, "email", r.Email)
	ini.GravarString(secao, "fone", r.Fone)
	if r.IDCSRT != 0 {
		ini.GravarInteiro(secao, "idCSRT", r.IDCSRT)
		ini.GravarString(secao, "hashCSRT", r.HashCSRT)
	}
}
