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
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitura do formato .ini do ACBr.
// Porte de TNFAgIniReader (ACBrNFAg.IniReader.pas).
//
// Convencao de secao indexada (indices base 1), digitos entre parenteses:
//
//	[gMedNN] (2)        [detNNN] (3)        [gMedicaoNNN] (3)
//	[gTarifNNND] (3+1)  [gProcRefNNN] (3)   [gProcNNNPP] (3+2)
//	[PISNNN] [COFINSNNN] [retTribNNN] [TFSNNN] [TFUNNN] (3)
//	[gHistConsN] (1)    [gConsHNN] (1+2)    [gAnaliseNN] (2)
//	[autXMLNN] (2)
//
// OMISSAO DO ACBr REPLICADA: a VERSAO e lida da secao [infNFGas] -- e o
// literal que esta no TNFAgIniReader.LerIni (copy-paste da NFGas), enquanto
// o IniWriter grava em [infNFAg]. O round-trip do proprio ACBr perde a
// versao (cai no default da configuracao). Testado.

// LerINI le uma NFAg a partir do conteudo (ou caminho) de um arquivo .ini.
func LerINI(iniOuCaminho string, cfg Configuracoes) (*NFAg, error) {
	ini, err := pcn.LerINIArquivoOuString(iniOuCaminho)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIReader le uma NFAg a partir de um io.Reader com conteudo .ini.
func LerINIReader(r io.Reader, cfg Configuracoes) (*NFAg, error) {
	ini, err := pcn.LerINI(r)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIArquivo le uma NFAg de um arquivo .ini.
func LerINIArquivo(caminho string, cfg Configuracoes) (*NFAg, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfag: abrir %q: %w", caminho, err)
	}
	defer f.Close()
	return LerINIReader(f, cfg)
}

// lerINIDocumento e o porte de TNFAgIniReader.LerIni.
func lerINIDocumento(ini *pcn.INI, cfg Configuracoes) (*NFAg, error) {
	if ini == nil {
		return nil, ErrXMLVazio
	}
	n := &NFAg{}

	// secao [infNFGas] de proposito -- ver nota do cabecalho
	n.InfNFAg.Versao = ini.LerFloat("infNFGas", "versao", cfg.VersaoDF.Float())

	lerINIIde(ini, &n.Ide, cfg)
	lerINIEmit(ini, &n.Emit)

	// cUF e cMunFG sao lidos DEPOIS do emitente (default de cMunFG e o
	// municipio do emitente), como no original.
	n.Ide.CUF = ini.LerInteiro("ide", "cUF", 35)
	n.Ide.CMunFG = ini.LerInteiro("ide", "cMunFG", n.Emit.EnderEmit.CMun)

	lerINIDest(ini, &n.Dest)
	lerINILigacao(ini, &n.Ligacao)
	lerINIGSub(ini, &n.GSub)
	lerINIGMed(ini, n)
	lerINIGFatConjunto(ini, &n.GFatConjunto)
	lerINIDet(ini, n)
	lerINITotal(ini, &n.Total)
	rtc.LerINIPgtoVinc(ini, &n.PgtoVinc)
	lerINIGFat(ini, &n.GFat)
	lerINIGAgencia(ini, &n.GAgencia)
	lerINIGQualiAgua(ini, &n.GQualiAgua)
	lerINIAutXML(ini, n)
	lerINIInfPAA(ini, &n.InfPAA)
	lerINIInfAdic(ini, &n.InfAdic)
	lerINIInfRespTec(ini, &n.InfRespTec)

	return n, nil
}

func time0() time.Time { return time.Time{} }

func lerINIIde(ini *pcn.INI, ide *Ide, cfg Configuracoes) {
	const secao = "ide"

	ide.TpAmb, _ = pcn.ParseTipoAmbiente(ini.LerString(secao, "tpAmb", cfg.Ambiente.String()))
	ide.Modelo = ini.LerInteiro(secao, "Modelo", ModeloNFAg)
	ide.Serie = ini.LerInteiro(secao, "Serie", 1)
	ide.NNF = ini.LerInteiro(secao, "nNF", 0)
	ide.CNF = ini.LerInteiro(secao, "cNF", 0)
	ide.DhEmi = ini.LerData(secao, "dhEmi", time0())
	ide.TpEmis, _ = pcn.ParseTipoEmissao(ini.LerString(secao, "tpEmis", cfg.TpEmis.String()))
	ide.NSiteAutoriz, _ = ParseSiteAutorizador(ini.LerString(secao, "nSiteAutoriz", "0"))
	ide.VerProc = ini.LerString(secao, "verProc", "ACBrNFAg")
	ide.FinNFAg, _ = ParseFinalidadeNFAg(ini.LerString(secao, "finNFAg", "0"))
	ide.TpFat, _ = ParseTpFat(ini.LerString(secao, "tpFat", "1"))
	ide.DhCont = ini.LerData(secao, "dhCont", time0())
	ide.XJust = ini.LerString(secao, "xJust", "")
	ide.TpPagAnt, _ = pcn.ParseTpPagAnt(ini.LerString(secao, "tpPagAnt", ""))

	rtc.LerINIGCompraGovReduzido(ini, &ide.GCompraGov)
}

func lerINIEmit(ini *pcn.INI, e *Emit) {
	const secao = "emit"

	e.CNPJ = ini.LerString(secao, "CNPJ", "")
	e.IE = ini.LerString(secao, "IE", "")
	e.XNome = ini.LerString(secao, "xNome", "")
	e.XFant = ini.LerString(secao, "xFant", "")
	e.ISUFEmit = ini.LerString(secao, "ISUFEmit", "")

	lerINIEndereco(ini, secao, &e.EnderEmit)
}

func lerINIEndereco(ini *pcn.INI, secao string, e *Endereco) {
	e.XLgr = ini.LerString(secao, "xLgr", "")
	e.Nro = ini.LerString(secao, "nro", "")
	e.XCpl = ini.LerString(secao, "xCpl", "")
	e.XBairro = ini.LerString(secao, "xBairro", "")
	e.CMun = ini.LerInteiro(secao, "cMun", 0)
	e.XMun = ini.LerString(secao, "xMun", "")
	e.CEP = ini.LerInteiro(secao, "CEP", 0)
	e.UF = ini.LerString(secao, "UF", "")
	e.Fone = ini.LerString(secao, "fone", "")
	e.Email = ini.LerString(secao, "email", "")
}

func lerINIDest(ini *pcn.INI, d *Dest) {
	const secao = "dest"
	if !ini.SecaoExiste(secao) {
		return
	}
	d.XNome = ini.LerString(secao, "xNome", "")
	d.CNPJCPF = ini.LerString(secao, "CNPJCPF", "")
	// idOutros com fallback para a chave idEstrangeiro, como o original
	d.IDOutros = ini.LerString(secao, "idOutros", ini.LerString(secao, "idEstrangeiro", ""))
	d.IE = ini.LerString(secao, "IE", "")
	d.IM = ini.LerString(secao, "IM", "")
	d.CNIS = ini.LerString(secao, "cNIS", "")
	d.NB = ini.LerString(secao, "NB", "")
	d.XNomeAdicional = ini.LerString(secao, "xNomeAdicional", "")

	lerINIEndereco(ini, secao, &d.EnderDest)
}

func lerINILigacao(ini *pcn.INI, l *Ligacao) {
	// a secao e 'Ligacao', com L maiusculo, como no original
	const secao = "Ligacao"
	if !ini.SecaoExiste(secao) {
		return
	}
	l.IDLigacao = ini.LerString(secao, "idLigacao", "")
	l.IDCodCliente = ini.LerString(secao, "idCodCliente", "")
	l.TpLigacao, _ = ParseTpLigacao(ini.LerString(secao, "tpLigacao", ""))
	l.LatGPS = ini.LerString(secao, "latGPS", "")
	l.LongGPS = ini.LerString(secao, "longGPS", "")
	l.CodRoteiroLeitura = ini.LerString(secao, "codRoteiroLeitura", "")
}

func lerINIGSub(ini *pcn.INI, g *GSub) {
	const secao = "gSub"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.ChNFAg = ini.LerString(secao, "chNFAg", "")
	g.MotSub, _ = ParseMotSub(ini.LerString(secao, "motSub", ""))
}

func lerINIGMed(ini *pcn.INI, n *NFAg) {
	n.GMed = nil
	for idx := 1; ; idx++ {
		secao := fmt.Sprintf("gMed%02d", idx)
		v := ini.LerString(secao, "nMed", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		m := GMed{}
		m.NMed = ini.LerInteiro(secao, "nMed", 0)
		m.IDMedidor = ini.LerString(secao, "idMedidor", "")
		m.DMedAnt = ini.LerData(secao, "dMedAnt", time0())
		m.DMedAtu = ini.LerData(secao, "dMedAtu", time0())
		n.GMed = append(n.GMed, m)
	}
}

func lerINIGFatConjunto(ini *pcn.INI, g *GFatConjunto) {
	const secao = "gFatConjunto"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.ChNFAgFat = ini.LerString(secao, "chNFAgFat", "")
}

func lerINIDet(ini *pcn.INI, n *NFAg) {
	n.Det = nil
	for idx := 1; ; idx++ {
		secao := fmt.Sprintf("det%03d", idx)
		v := ini.LerString(secao, "nItem", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		d := Det{}
		d.NItem = ini.LerInteiro(secao, "nItem", 0)
		d.ChNFAgAnt = ini.LerString(secao, "chNFAgAnt", "")
		d.NItemAnt = ini.LerInteiro(secao, "nItemAnt", 0)

		d.Prod.IndOrigemQtd, _ = ParseIndOrigemQtd(ini.LerString(secao, "indOrigemQtd", "1"))
		d.Prod.CProd = ini.LerString(secao, "cProd", "")
		d.Prod.XProd = ini.LerString(secao, "xProd", "")
		// o ACBr le cClass com ReadInteger (o campo e Integer la); aqui a
		// string preserva zeros a esquerda quando o .ini os tiver
		d.Prod.CClass = ini.LerString(secao, "cClass", "")
		d.Prod.TpCategoria, _ = ParseTpCategoria(ini.LerString(secao, "tpCategoria", "1"))
		d.Prod.XCategoria = ini.LerString(secao, "xCategoria", "")
		d.Prod.QEconomias = ini.LerString(secao, "qEconomias", "")
		d.Prod.UMed, _ = ParseUMedFat(ini.LerString(secao, "uMed", ""))
		// o ACBr le qFaturada com ReadInteger (arredonda); aqui aceita
		// fracionado -- mesma divergencia de tipo do XML
		d.Prod.QFaturada = ini.LerFloat(secao, "qFaturada", 0)
		d.Prod.VItem = ini.LerFloat(secao, "vItem", 0)
		d.Prod.FatorPoluicao = ini.LerFloat(secao, "fatorPoluicao", 0)
		d.Prod.VProd = ini.LerFloat(secao, "vProd", 0)
		d.Prod.IndDevolucao, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDevolucao", ""))
		d.InfAdProd = ini.LerString(secao, "infAdProd", "")

		rtc.LerINIGPagAntecipadoProd(ini, &d.Prod.GPagAntecipado, idx, -1)
		lerINIGMedicao(ini, &d.Prod.GMedicao, idx)
		lerINIGTarif(ini, &d, idx)
		lerINIGProcRef(ini, &d.GProcRef, idx)
		rtc.LerINIIBSCBS(ini, &d.Imposto.IBSCBS, idx, -1)
		lerINIPIS(ini, &d.Imposto.PIS, idx)
		lerINICOFINS(ini, &d.Imposto.COFINS, idx)
		lerINIRetTrib(ini, &d.Imposto.RetTrib, idx)
		lerINITFS(ini, &d.Imposto.TFS, idx)
		lerINITFU(ini, &d.Imposto.TFU, idx)

		n.Det = append(n.Det, d)
	}
}

func lerINIGMedicao(ini *pcn.INI, g *GMedicao, det int) {
	secao := fmt.Sprintf("gMedicao%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	g.NMed = ini.LerInteiro(secao, "nMed", 0)
	g.TpMotNaoLeitura, _ = ParseTpMotNaoLeitura(ini.LerString(secao, "tpMotNaoLeitura", ""))

	// os campos do gMedida moram na MESMA secao gMedicaoNNN, como no
	// original
	g.GMedida.TpGrMed, _ = ParseTpGrMed(ini.LerString(secao, "tpGrMed", ""))
	g.GMedida.NUnidConsumo = ini.LerString(secao, "nUnidConsumo", "")
	g.GMedida.VUnidConsumo = ini.LerFloat(secao, "vUnidConsumo", 0)
	g.GMedida.UMed, _ = ParseUMedFat(ini.LerString(secao, "uMed", ""))
	g.GMedida.VMedAnt = ini.LerFloat(secao, "vMedAnt", 0)
	g.GMedida.VMedAtu = ini.LerFloat(secao, "vMedAtu", 0)
	g.GMedida.VConst = ini.LerFloat(secao, "vConst", 0)
	g.GMedida.VMed = ini.LerFloat(secao, "vMed", 0)
}

func lerINIGTarif(ini *pcn.INI, d *Det, det int) {
	d.GTarif = nil
	for idx := 1; ; idx++ {
		secao := fmt.Sprintf("gTarif%03d%d", det, idx)
		v := ini.LerString(secao, "dIniTarif", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		t := GTarif{}
		t.DIniTarif = ini.LerData(secao, "dIniTarif", time0())
		t.DFimTarif = ini.LerData(secao, "dFimTarif", time0())
		t.NAto = ini.LerString(secao, "nAto", "")
		t.AnoAto = ini.LerInteiro(secao, "anoAto", 0)
		t.TpFaixaCons, _ = ParseTpFaixaCons(ini.LerString(secao, "tpFaixaCons", ""))
		d.GTarif = append(d.GTarif, t)
	}
}

func lerINIGProcRef(ini *pcn.INI, g *GProcRef, det int) {
	secao := fmt.Sprintf("gProcRef%03d", det)
	v := ini.LerString(secao, "vItem", "FIM")
	if v == "FIM" || v == "" {
		return
	}
	g.VItem = pcn.StringToFloatDef(v, 0)
	g.QFaturada = ini.LerFloat(secao, "qFaturada", 0)
	g.VProd = ini.LerFloat(secao, "vProd", 0)
	g.IndDevolucao, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDevolucao", ""))

	g.GProc = nil
	for idx := 1; ; idx++ {
		s := fmt.Sprintf("gProc%03d%02d", det, idx)
		nProc := ini.LerString(s, "nProcesso", "FIM")
		if nProc == "FIM" || nProc == "" {
			break
		}
		p := GProc{}
		p.TpProc, _ = ParseTpProc(ini.LerString(s, "tpProc", ""))
		p.NProcesso = nProc
		g.GProc = append(g.GProc, p)
	}
}

func lerINIPIS(ini *pcn.INI, p *PIS, det int) {
	secao := fmt.Sprintf("PIS%03d", det)
	cst := ini.LerString(secao, "CST", "FIM")
	if cst == "FIM" || cst == "" {
		return
	}
	p.CST, _ = pcn.ParseCSTPis(cst)
	p.VBC = ini.LerFloat(secao, "vBC", 0)
	p.PPIS = ini.LerFloat(secao, "pPIS", 0)
	p.VPIS = ini.LerFloat(secao, "vPIS", 0)
}

func lerINICOFINS(ini *pcn.INI, c *COFINS, det int) {
	secao := fmt.Sprintf("COFINS%03d", det)
	cst := ini.LerString(secao, "CST", "FIM")
	if cst == "FIM" || cst == "" {
		return
	}
	c.CST, _ = pcn.ParseCSTCofins(cst)
	c.VBC = ini.LerFloat(secao, "vBC", 0)
	c.PCOFINS = ini.LerFloat(secao, "pCOFINS", 0)
	c.VCOFINS = ini.LerFloat(secao, "vCOFINS", 0)
}

func lerINIRetTrib(ini *pcn.INI, r *RetTrib, det int) {
	secao := fmt.Sprintf("retTrib%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	r.VRetPIS = ini.LerFloat(secao, "vRetPIS", 0)
	r.VRetCOFINS = ini.LerFloat(secao, "vRetCOFINS", 0)
	r.VRetCSLL = ini.LerFloat(secao, "vRetCSLL", 0)
	r.VIRRF = ini.LerFloat(secao, "vIRRF", 0)
}

func lerINITFS(ini *pcn.INI, t *TFS, det int) {
	secao := fmt.Sprintf("TFS%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	t.VBCTFS = ini.LerFloat(secao, "vBCTFS", 0)
	t.PTFS = ini.LerFloat(secao, "pTFS", 0)
	t.VTFS = ini.LerFloat(secao, "vTFS", 0)
}

// lerINITFU le a secao TFUNNN. Atencao: o IniWriter do ACBr GRAVA o TFU na
// secao TFSNNN (bug replicado no ini_writer.go), entao um .ini gerado pelo
// proprio ACBr nunca reali menta este grupo -- so um .ini montado a mao.
func lerINITFU(ini *pcn.INI, t *TFU, det int) {
	secao := fmt.Sprintf("TFU%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	t.VBCTFU = ini.LerFloat(secao, "vBCTFU", 0)
	t.PTFU = ini.LerFloat(secao, "pTFU", 0)
	t.VTFU = ini.LerFloat(secao, "vTFU", 0)
}

func lerINITotal(ini *pcn.INI, t *Total) {
	const secao = "total"

	t.VProd = ini.LerFloat(secao, "vProd", 0)
	t.VRetPIS = ini.LerFloat(secao, "vRetPIS", 0)
	t.VRetCOFINS = ini.LerFloat(secao, "vRetCOFINS", 0)
	t.VRetCSLL = ini.LerFloat(secao, "vRetCSLL", 0)
	t.VIRRF = ini.LerFloat(secao, "vIRRF", 0)
	t.VCOFINS = ini.LerFloat(secao, "vCOFINS", 0)
	t.VPIS = ini.LerFloat(secao, "vPIS", 0)
	t.VTFS = ini.LerFloat(secao, "vTFS", 0)
	t.VTFU = ini.LerFloat(secao, "vTFU", 0)
	t.VNF = ini.LerFloat(secao, "vNF", 0)
	t.VTotDFe = ini.LerFloat(secao, "vTotDFe", 0)

	rtc.LerINIIBSCBSTot(ini, &t.IBSCBSTot)
}

func lerINIGFat(ini *pcn.INI, g *GFat) {
	const secao = "gFat"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.CompetFat = ini.LerData(secao, "CompetFat", time0())
	g.DVencFat = ini.LerData(secao, "dVencFat", time0())
	g.DApresFat = ini.LerData(secao, "dApresFat", time0())
	g.DProxLeitura = ini.LerData(secao, "dProxLeitura", time0())
	g.NFat = ini.LerString(secao, "nFat", "")
	g.CodBarras = ini.LerString(secao, "codBarras", "")
	g.CodDebAuto = ini.LerString(secao, "codDebAuto", "")
	g.CodBanco = ini.LerString(secao, "codBanco", "")
	g.CodAgencia = ini.LerString(secao, "codAgencia", "")

	if ini.SecaoExiste("enderCorresp") {
		lerINIEndereco(ini, "enderCorresp", &g.EnderCorresp)
	}
	if ini.SecaoExiste("gPIX") {
		g.GPIX.URLQRCodePIX = ini.LerString("gPIX", "urlQRCodePIX", "")
	}
}

func lerINIGAgencia(ini *pcn.INI, g *GAgencia) {
	const secao = "gAgencia"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.Econ = ini.LerString(secao, "econ", "")
	g.EconAcumulada = ini.LerString(secao, "econAcumulada", "")
	g.SPrestador = ini.LerString(secao, "sPrestador", "")
	g.DEmissSelo = ini.LerData(secao, "dEmissSelo", time0())
	g.SRegulador = ini.LerString(secao, "sRegulador", "")
	g.NAgenciaAtend = ini.LerString(secao, "nAgenciaAtend", "")
	g.EnderAgenciaAtend = ini.LerString(secao, "enderAgenciaAtend", "")

	g.GHistCons = nil
	for h := 1; ; h++ {
		secaoH := fmt.Sprintf("gHistCons%d", h)
		x := ini.LerString(secaoH, "xHistorico", "FIM")
		if x == "FIM" || x == "" {
			break
		}
		hist := GHistCons{XHistorico: x}
		hist.MedMensal = ini.LerFloat(secaoH, "medMensal", 0)

		for c := 1; ; c++ {
			secaoC := fmt.Sprintf("gCons%d%02d", h, c)
			comp := ini.LerString(secaoC, "CompetFat", "FIM")
			if comp == "FIM" || comp == "" {
				break
			}
			gc := GCons{}
			gc.CompetFat = ini.LerData(secaoC, "CompetFat", time0())
			gc.UMed, _ = ParseUMedFat(ini.LerString(secaoC, "uMed", ""))
			gc.QtdDias = ini.LerString(secaoC, "qtdDias", "")
			gc.MedDiaria = ini.LerFloat(secaoC, "medDiaria", 0)
			gc.Consumo = ini.LerFloat(secaoC, "consumo", 0)
			gc.VolFat = ini.LerFloat(secaoC, "volFat", 0)
			hist.GCons = append(hist.GCons, gc)
		}
		g.GHistCons = append(g.GHistCons, hist)
	}
}

func lerINIGQualiAgua(ini *pcn.INI, g *GQualiAgua) {
	const secao = "gQualiAgua"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.CompetAnalise = ini.LerData(secao, "CompetAnalise", time0())
	g.Conclusao = ini.LerString(secao, "Conclusao", "")
	g.CProcesso = ini.LerString(secao, "cProcesso", "")
	g.SistemaAbast = ini.LerString(secao, "SistemaAbast", "")

	g.GAnalise = nil
	for idx := 1; ; idx++ {
		s := fmt.Sprintf("gAnalise%02d", idx)
		x := ini.LerString(s, "xItemAnalisado", "FIM")
		if x == "FIM" || x == "" {
			break
		}
		a := GAnalise{XItemAnalisado: x}
		a.NAmostraMinima = ini.LerString(s, "nAmostraMinima", "")
		a.NAmostraAnalisada = ini.LerString(s, "nAmostraAnalisada", "")
		a.NAmostraFPadrao = ini.LerString(s, "nAmostraFPadrao", "")
		a.NAmostraDPadrao = ini.LerString(s, "nAmostraDPadrao", "")
		a.NMediaMensal = ini.LerString(s, "nMediaMensal", "")
		a.XValorReferencia = ini.LerString(s, "xValorReferencia", "")
		g.GAnalise = append(g.GAnalise, a)
	}
}

func lerINIAutXML(ini *pcn.INI, n *NFAg) {
	n.AutXML = nil
	for idx := 1; ; idx++ {
		secao := fmt.Sprintf("autXML%02d", idx)
		v := pcn.OnlyCPFCNPJAlphaNum(ini.LerString(secao, "CNPJCPF", ""))
		if v == "" {
			break
		}
		n.AutXML = append(n.AutXML, AutXML{CNPJCPF: v})
	}
}

func lerINIInfPAA(ini *pcn.INI, p *InfPAA) {
	const secao = "infPAA"
	if !ini.SecaoExiste(secao) {
		return
	}
	p.CNPJPAA = ini.LerString(secao, "CNPJPAA", "")
}

func lerINIInfAdic(ini *pcn.INI, i *InfAdic) {
	const secao = "infAdic"
	if !ini.SecaoExiste(secao) {
		return
	}
	i.InfAdFisco = ini.LerString(secao, "infAdFisco", "")
	if v := ini.LerString(secao, "infCpl", ""); v != "" {
		i.InfCpl = []string{v}
	}
}

func lerINIInfRespTec(ini *pcn.INI, r *InfRespTec) {
	const secao = "infRespTec"
	if !ini.SecaoExiste(secao) {
		return
	}
	r.CNPJ = ini.LerString(secao, "CNPJ", "")
	r.XContato = ini.LerString(secao, "xContato", "")
	r.Email = ini.LerString(secao, "email", "")
	r.Fone = ini.LerString(secao, "fone", "")
	r.IDCSRT = ini.LerInteiro(secao, "idCSRT", 0)
	r.HashCSRT = ini.LerString(secao, "hashCSRT", "")
}
