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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitura do formato .ini do ACBr.
// Porte de TNFGasIniReader (ACBrNFGas.IniReader.pas).
//
// Convencao de secao indexada, com o numero de digitos entre parenteses:
//
//	[gVolContratNN] (2)   [gMedNN] (2)       [detNNN] (3)
//	[gMedicaoNNN] (3)     [gTarifNNND] (3+1) [gAgregadoraNNN] (3)
//	[gProcRefNNN] (3)     [gProcNNNPP] (3+2)
//	[ICMSNNN] [PISNNN] [COFINSNNN] [retTribNNN] [TxRegNNN] (3)
//	[gHistConsN] (1)      [gConsHNN] (1+2)   [autXMLNN] (2)
//
// Os indices sao base 1. Um grupo repetido termina quando a chave sentinela
// da secao seguinte nao existe -- e o "FIM" do original.
//
// DIVERGENCIA: as datas sao lidas por pcn.EncodeDataHora, que aceita tanto
// o formato ISO quanto dd/mm/aaaa. O original usa StringToDateTime, que
// depende do formato de data do sistema operacional.

// LerINI le uma NFGas a partir do conteudo de um arquivo .ini.
//
// O parametro aceita tanto o caminho de um arquivo quanto o proprio
// conteudo, como LerIniArquivoOuString do ACBr.
func LerINI(iniOuCaminho string, cfg Configuracoes) (*NFGas, error) {
	ini, err := pcn.LerINIArquivoOuString(iniOuCaminho)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIReader le uma NFGas a partir de um io.Reader com conteudo .ini.
func LerINIReader(r io.Reader, cfg Configuracoes) (*NFGas, error) {
	ini, err := pcn.LerINI(r)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIArquivo le uma NFGas de um arquivo .ini.
func LerINIArquivo(caminho string, cfg Configuracoes) (*NFGas, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfgas: abrir %q: %w", caminho, err)
	}
	defer f.Close()
	return LerINIReader(f, cfg)
}

// lerINIDocumento e o porte de TNFGasIniReader.LerIni.
func lerINIDocumento(ini *pcn.INI, cfg Configuracoes) (*NFGas, error) {
	if ini == nil {
		return nil, ErrXMLVazio
	}
	n := &NFGas{}

	n.InfNFGas.Versao = ini.LerFloat("infNFGas", "versao", cfg.VersaoDF.Float())

	lerINIIde(ini, &n.Ide, cfg)
	lerINIEmit(ini, &n.Emit)

	// cUF e cMunFG sao lidos DEPOIS do emitente, porque o default de cMunFG
	// e o municipio do emitente. Ordem preservada do original.
	n.Ide.CUF = ini.LerInteiro("ide", "cUF", 35)
	n.Ide.CMunFG = ini.LerInteiro("ide", "cMunFG", n.Emit.EnderEmit.CMun)

	lerINIDest(ini, &n.Dest)
	lerINIInstalacao(ini, &n.Instalacao)
	lerINIGSub(ini, &n.GSub)
	lerINIGVolContrat(ini, n)
	lerINIGMed(ini, n)
	lerINIDet(ini, n)
	lerINITotal(ini, &n.Total)
	rtc.LerINIPgtoVinc(ini, &n.PgtoVinc)
	lerINIGFat(ini, &n.GFat)
	lerINIGAgencia(ini, &n.GAgencia)
	lerINIAutXML(ini, n)
	lerINIInfAdic(ini, &n.InfAdic)
	lerINIInfRespTec(ini, &n.InfRespTec)

	return n, nil
}

func lerINIIde(ini *pcn.INI, ide *Ide, cfg Configuracoes) {
	const secao = "ide"

	ide.TpAmb, _ = pcn.ParseTipoAmbiente(ini.LerString(secao, "tpAmb", cfg.Ambiente.String()))
	ide.Modelo = ini.LerInteiro(secao, "Modelo", ModeloNFGas)
	ide.Serie = ini.LerInteiro(secao, "Serie", 1)
	ide.NNF = ini.LerInteiro(secao, "nNF", 0)
	ide.CNF = ini.LerInteiro(secao, "cNF", 0)
	ide.DhEmi = ini.LerData(secao, "dhEmi", time0())
	ide.TpEmis, _ = pcn.ParseTipoEmissao(ini.LerString(secao, "tpEmis", cfg.TpEmis.String()))
	ide.NSiteAutoriz, _ = ParseSiteAutorizador(ini.LerString(secao, "nSiteAutoriz", "0"))
	ide.VerProc = ini.LerString(secao, "verProc", "ACBrNFGas")
	ide.FinNFGas, _ = ParseFinalidadeNFGas(ini.LerString(secao, "finNFGas", "0"))
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

	lerINIEndereco(ini, secao, &e.EnderEmit, true)
}

// lerINIEndereco le os campos de endereco, que no .ini ficam ACHATADOS na
// mesma secao do grupo pai (emit, dest) ou em secao propria
// (enderCorresp).
func lerINIEndereco(ini *pcn.INI, secao string, end *Endereco, comPais bool) {
	end.XLgr = ini.LerString(secao, "xLgr", "")
	end.Nro = ini.LerString(secao, "nro", "")
	end.XCpl = ini.LerString(secao, "xCpl", "")
	end.XBairro = ini.LerString(secao, "xBairro", "")
	end.CMun = ini.LerInteiro(secao, "cMun", 0)
	end.XMun = ini.LerString(secao, "xMun", "")
	end.CEP = ini.LerInteiro(secao, "CEP", 0)
	end.UF = ini.LerString(secao, "UF", "")
	end.Fone = ini.LerString(secao, "fone", "")
	end.Email = ini.LerString(secao, "email", "")
	if comPais {
		end.CPais = ini.LerInteiro(secao, "cPais", 0)
		end.XPais = ini.LerString(secao, "xPais", "")
	}
}

// lerINIDest le a secao [dest], que e OPCIONAL -- se nao existir, o grupo
// fica zerado.
//
// Diferente do XML, aqui indIEDest E lido, e idOutros cai para
// idEstrangeiro quando ausente.
func lerINIDest(ini *pcn.INI, d *Dest) {
	const secao = "dest"
	if !ini.SecaoExiste(secao) {
		return
	}

	d.XNome = ini.LerString(secao, "xNome", "")
	d.CNPJCPF = ini.LerString(secao, "CNPJCPF", "")

	if v := ini.LerString(secao, "idOutros", ""); v != "" {
		d.IDEstrangeiro = v
		d.TagIDOrigem = "idOutros"
	} else if v := ini.LerString(secao, "idEstrangeiro", ""); v != "" {
		d.IDEstrangeiro = v
		d.TagIDOrigem = "idEstrangeiro"
	}

	d.IndIEDest, _ = ParseIndIEDest(ini.LerString(secao, "indIEDest", "1"))
	d.IE = ini.LerString(secao, "IE", "")
	d.IM = ini.LerString(secao, "IM", "")
	d.CNIS = ini.LerString(secao, "cNIS", "")
	d.NB = ini.LerString(secao, "NB", "")
	d.XNomeAdicional = ini.LerString(secao, "xNomeAdicional", "")

	lerINIEndereco(ini, secao, &d.EnderDest, true)
}

func lerINIInstalacao(ini *pcn.INI, i *Instalacao) {
	const secao = "Instalacao"
	if !ini.SecaoExiste(secao) {
		return
	}
	i.IDInstalacao = ini.LerString(secao, "idInstalacao", "")
	i.IDCodCliente = ini.LerString(secao, "idCodCliente", "")
	i.TpInstalacao, _ = ParseTpInstalacao(ini.LerString(secao, "tpInstalacao", ""))
	i.NContrato = ini.LerString(secao, "nContrato", "")
	i.TpClasse, _ = ParseTpClasse(ini.LerString(secao, "tpClasse", ""))
	i.XClasse = ini.LerString(secao, "xClasse", "")
	i.LatGPS = ini.LerString(secao, "latGPS", "")
	i.LongGPS = ini.LerString(secao, "longGPS", "")
	i.CodRoteiroLeitura = ini.LerString(secao, "codRoteiroLeitura", "")
}

// lerINIGSub le a secao [gSub]. Os campos de gNF ficam ACHATADOS na mesma
// secao, como no original.
func lerINIGSub(ini *pcn.INI, g *GSub) {
	const secao = "gSub"
	if !ini.SecaoExiste(secao) {
		return
	}
	g.ChNFGas = ini.LerString(secao, "chNFGas", "")
	g.MotSub, _ = ParseMotSub(ini.LerString(secao, "motSub", ""))
	g.GNF.CNPJ = ini.LerString(secao, "CNPJ", "")
	g.GNF.Serie = ini.LerString(secao, "Serie", "")
	g.GNF.NNF = ini.LerInteiro(secao, "nNF", 0)
	g.GNF.CompetEmis = ini.LerData(secao, "CompetEmis", time0())
	g.GNF.CompetApur = ini.LerData(secao, "CompetApur", time0())
	g.GNF.Hash115 = ini.LerString(secao, "hash115", "")
}

func lerINIGVolContrat(ini *pcn.INI, n *NFGas) {
	n.GVolContrat = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("gVolContrat%02d", i)
		v := ini.LerString(secao, "nContrat", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		item := GVolContrat{NContrat: ini.LerInteiro(secao, "nContrat", 0)}
		item.TpVolContrat, _ = ParseVolContrat(ini.LerString(secao, "tpVolContrat", ""))
		item.QUnidContrat = ini.LerFloat(secao, "qUnidContrat", 0)
		n.GVolContrat = append(n.GVolContrat, item)
	}
}

func lerINIGMed(ini *pcn.INI, n *NFGas) {
	n.GMed = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("gMed%02d", i)
		v := ini.LerString(secao, "nMed", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		item := GMed{NMed: ini.LerInteiro(secao, "nMed", 0)}
		// idEqp cai para idMedidor, nome antigo da chave.
		item.IDEqp = ini.LerString(secao, "idEqp", ini.LerString(secao, "idMedidor", ""))
		item.DMedAnt = ini.LerData(secao, "dMedAnt", time0())
		item.VMedAnt = ini.LerFloat(secao, "vMedAnt", 0)
		item.DMedAtu = ini.LerData(secao, "dMedAtu", time0())
		item.VMedAtu = ini.LerFloat(secao, "vMedAtu", 0)
		item.TpEqp, _ = ParseTpEqp(ini.LerString(secao, "tpEqp", ""))
		item.TpMedidor, _ = ParseTpMedidor(ini.LerString(secao, "tpMedidor", ""))
		n.GMed = append(n.GMed, item)
	}
}

// lerINIDet le as secoes [detNNN] e todas as secoes filhas do item.
//
// Note que orig, indSemCST, os campos de prod e infAdProd ficam na propria
// secao [detNNN]; so os grupos maiores tem secao propria.
func lerINIDet(ini *pcn.INI, n *NFGas) {
	n.Det = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("det%03d", i)
		v := ini.LerString(secao, "nItem", "FIM")
		if v == "FIM" || v == "" {
			break
		}

		var d Det
		d.NItem = ini.LerInteiro(secao, "nItem", 0)
		d.ChNFGasAnt = ini.LerString(secao, "chNFGasAnt", "")
		d.NItemAnt = ini.LerInteiro(secao, "nItemAnt", 0)

		d.GNormal.Imposto.Orig, _ = pcn.ParseOrigemMercadoria(ini.LerString(secao, "orig", "0"))
		d.GNormal.Imposto.IndSemCST, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indSemCST", ""))

		p := &d.GNormal.Prod
		p.IndOrigemQtd, _ = ParseIndOrigemQtd(ini.LerString(secao, "indOrigemQtd", "1"))
		p.CProd = ini.LerString(secao, "cProd", "")
		p.XProd = ini.LerString(secao, "xProd", "")
		p.CClass = ini.LerString(secao, "cClass", "")
		p.CFOP = ini.LerInteiro(secao, "CFOP", 0)
		p.UMed, _ = ParseUMedItem(ini.LerString(secao, "uMed", ""))
		p.QFaturada = ini.LerFloat(secao, "qFaturada", 0)
		p.VItem = ini.LerFloat(secao, "vItem", 0)
		p.FatorPCS = ini.LerFloat(secao, "fatorPCS", 0)
		p.FatorPTZ = ini.LerFloat(secao, "fatorPTZ", 0)
		p.FatorP = ini.LerFloat(secao, "fatorP", 0)
		p.FatorT = ini.LerFloat(secao, "fatorT", 0)
		p.VProd = ini.LerFloat(secao, "vProd", 0)
		p.IndDevolucao, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDevolucao", ""))
		d.GNormal.InfAdProd = ini.LerString(secao, "infAdProd", "")

		lerINIGMedicao(ini, &p.GMedicao, i)
		rtc.LerINIGPagAntecipadoProd(ini, &p.GPagAntecipado, i, rtc.SemIndice)

		lerINIGTarif(ini, &d.GNormal.GTarif, i)
		lerINIGAgregadora(ini, &d.GAgregadora, i)
		lerINIGProcRef(ini, &d.GNormal.GProcRef, i)

		lerINIICMS(ini, &d.GNormal.Imposto.ICMS, i)
		rtc.LerINIIBSCBS(ini, &d.GNormal.Imposto.IBSCBS, i, rtc.SemIndice)
		lerINIPIS(ini, &d.GNormal.Imposto.PIS, i)
		lerINICOFINS(ini, &d.GNormal.Imposto.COFINS, i)
		lerINIRetTrib(ini, &d.GNormal.Imposto.RetTrib, i)
		lerINITxReg(ini, &d.GNormal.Imposto.TxReg, i)

		n.Det = append(n.Det, d)
	}
}

func lerINIGMedicao(ini *pcn.INI, g *GMedicao, det int) {
	secao := fmt.Sprintf("gMedicao%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	g.NMed = ini.LerInteiro(secao, "nMed", 0)
	g.NContrat = ini.LerInteiro(secao, "nContrat", 0)
	// Os campos de gMedida ficam ACHATADOS nesta secao.
	g.GMedida.UMed, _ = ParseUMed(ini.LerString(secao, "uMed", ""))
	g.GMedida.VMed = ini.LerFloat(secao, "vMed", 0)
	g.TpMotNaoLeitura, _ = ParseTpMotNaoLeitura(ini.LerString(secao, "tpMotNaoLeitura", ""))
	g.XMotNaoLeitura = ini.LerString(secao, "xMotNaoLeitura", "")
}

func lerINIGTarif(ini *pcn.INI, lista *[]GTarif, det int) {
	*lista = nil
	for t := 1; ; t++ {
		secao := fmt.Sprintf("gTarif%03d%d", det, t)
		v := ini.LerString(secao, "dIniTarif", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		item := GTarif{
			DIniTarif: pcn.EncodeDataHoraDef(v),
			DFimTarif: ini.LerData(secao, "dFimTarif", time0()),
			NAto:      ini.LerString(secao, "nAto", ""),
			AnoAto:    ini.LerInteiro(secao, "anoAto", 0),
		}
		item.TpFaixaCons, _ = ParseTpFaixaCons(ini.LerString(secao, "tpFaixaCons", ""))
		item.VTarifAplic = ini.LerFloat(secao, "vTarifAplic", 0)
		*lista = append(*lista, item)
	}
}

func lerINIGAgregadora(ini *pcn.INI, g *GAgregadora, det int) {
	secao := fmt.Sprintf("gAgregadora%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	g.CClass = ini.LerString(secao, "cClass", "")
	g.VTotDFe = ini.LerFloat(secao, "vTotDFe", 0)
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
	for p := 1; ; p++ {
		s := fmt.Sprintf("gProc%03d%02d", det, p)
		nProcesso := ini.LerString(s, "nProcesso", "FIM")
		if nProcesso == "FIM" || nProcesso == "" {
			break
		}
		item := GProc{NProcesso: nProcesso}
		item.TpProc, _ = ParseTpProc(ini.LerString(s, "tpProc", ""))
		g.GProc = append(g.GProc, item)
	}
}

func lerINIICMS(ini *pcn.INI, icms *ICMS, det int) {
	secao := fmt.Sprintf("ICMS%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	icms.IndSemCST, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indSemCST", ""))
	icms.CST, _ = pcn.ParseCSTIcms(ini.LerString(secao, "CST", ""))
	icms.ModBC, _ = ParseDeterminacaoBaseIcms(ini.LerString(secao, "modBC", ""))
	icms.PRedBC = ini.LerFloat(secao, "pRedBC", 0)
	icms.VBC = ini.LerFloat(secao, "vBC", 0)
	icms.PICMS = ini.LerFloat(secao, "pICMS", 0)
	icms.VICMS = ini.LerFloat(secao, "vICMS", 0)
	icms.VICMSDeson = ini.LerFloat(secao, "vICMSDeson", 0)
	// No .ini motDesICMS e lido pelo CODIGO -- diferente do XML, onde o
	// ACBr le como inteiro e trata o codigo como ordinal.
	icms.MotDesICMS, _ = ParseMotivoDesoneracaoICMS(ini.LerString(secao, "motDesICMS", ""))
	icms.IndDeduzDeson, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDeduzDeson", ""))
	icms.CBenef = ini.LerString(secao, "cBenef", "")
	icms.ModBCST, _ = ParseDeterminacaoBaseIcmsST(ini.LerString(secao, "modBCST", ""))
	icms.PMVAST = ini.LerFloat(secao, "pMVAST", 0)
	icms.PRedBCST = ini.LerFloat(secao, "pRedBCST", 0)
	icms.VBCST = ini.LerFloat(secao, "vBCST", 0)
	icms.PICMSST = ini.LerFloat(secao, "pICMSST", 0)
	icms.VICMSST = ini.LerFloat(secao, "vICMSST", 0)
	icms.VBCFCP = ini.LerFloat(secao, "vBCFCP", 0)
	icms.PFCPST = ini.LerFloat(secao, "pFCPST", 0)
	icms.VFCPST = ini.LerFloat(secao, "vFCPST", 0)
	icms.VBCFCPST = ini.LerFloat(secao, "vBCFCPST", 0)
	icms.VBCSTRet = ini.LerFloat(secao, "vBCSTRet", 0)
	icms.PICMSSTRet = ini.LerFloat(secao, "pICMSSTRet", 0)
	icms.VICMSSubstituto = ini.LerFloat(secao, "vICMSSubstituto", 0)
	icms.VICMSSTRet = ini.LerFloat(secao, "vICMSSTRet", 0)
	icms.VBCFCPSTRet = ini.LerFloat(secao, "vBCFCPSTRet", 0)
	icms.PFCPSTRet = ini.LerFloat(secao, "pFCPSTRet", 0)
	icms.VFCPSTRet = ini.LerFloat(secao, "vFCPSTRet", 0)
	icms.PRedBCEfet = ini.LerFloat(secao, "pRedBCEfet", 0)
	icms.VBCEfet = ini.LerFloat(secao, "vBCEfet", 0)
	icms.PICMSEfet = ini.LerFloat(secao, "pICMSEfet", 0)
	icms.VICMSEfet = ini.LerFloat(secao, "vICMSEfet", 0)
	icms.PFCP = ini.LerFloat(secao, "pFCP", 0)
	icms.VFCP = ini.LerFloat(secao, "vFCP", 0)
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

// lerINIRetTrib le [retTribNNN].
//
// Atencao: no .ini a chave e vRetCOFINS em CAIXA ALTA, enquanto no XML a
// tag e vRetCofins. E assim no ACBr.
// vBCIRRF nao existe no .ini, como tambem nao e lido do XML.
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

func lerINITxReg(ini *pcn.INI, t *TxReg, det int) {
	secao := fmt.Sprintf("TxReg%03d", det)
	if !ini.SecaoExiste(secao) {
		return
	}
	t.VBC = ini.LerFloat(secao, "vBC", 0)
	t.PTaxa = ini.LerFloat(secao, "pTaxa", 0)
	t.VTaxa = ini.LerFloat(secao, "vTaxa", 0)
}

// lerINITotal le a secao [total], que no .ini ja vem achatada -- nao ha
// subsecoes ICMSTot nem vRetTribTot.
func lerINITotal(ini *pcn.INI, t *Total) {
	const secao = "total"

	t.VProd = ini.LerFloat(secao, "vProd", 0)
	t.VBC = ini.LerFloat(secao, "vBC", 0)
	t.VICMS = ini.LerFloat(secao, "vICMS", 0)
	t.VBCST = ini.LerFloat(secao, "vBCST", 0)
	t.VFCPST = ini.LerFloat(secao, "vFCPST", 0)
	t.VICMSDeson = ini.LerFloat(secao, "vICMSDeson", 0)
	t.VFCP = ini.LerFloat(secao, "vFCP", 0)
	t.VCOFINS = ini.LerFloat(secao, "vCOFINS", 0)
	t.VPIS = ini.LerFloat(secao, "vPIS", 0)
	t.VTxReg = ini.LerFloat(secao, "vTxReg", 0)
	t.VRetPIS = ini.LerFloat(secao, "vRetPIS", 0)
	t.VRetCOFINS = ini.LerFloat(secao, "vRetCOFINS", 0)
	t.VRetCSLL = ini.LerFloat(secao, "vRetCSLL", 0)
	t.VIRRF = ini.LerFloat(secao, "vIRRF", 0)
	t.VNF = ini.LerFloat(secao, "vNF", 0)
	t.VTotDFe = ini.LerFloat(secao, "vTotDFe", 0)

	// vST nao existe no .ini do ACBr, embora exista no XML.

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
	g.InfAdFat = ini.LerString(secao, "infAdFat", "")

	if ini.SecaoExiste("enderCorresp") {
		lerINIEndereco(ini, "enderCorresp", &g.EnderCorresp, false)
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
	g.NomeAgenciaAtend = ini.LerString(secao, "nomeAgenciaAtend", "")
	g.EnderAgenciaAtend = ini.LerString(secao, "enderAgenciaAtend", "")
	g.SitioAgenciaAtend = ini.LerString(secao, "sitioAgenciaAtend", "")
	g.InfAdReg = ini.LerString(secao, "infAdReg", "")

	g.GHistCons = nil
	for h := 1; ; h++ {
		sh := fmt.Sprintf("gHistCons%d", h)
		xHistorico := ini.LerString(sh, "xHistorico", "FIM")
		if xHistorico == "FIM" || xHistorico == "" {
			break
		}
		item := GHistCons{
			XHistorico: xHistorico,
			MedMensal:  ini.LerFloat(sh, "medMensal", 0),
		}

		for c := 1; ; c++ {
			sc := fmt.Sprintf("gCons%d%02d", h, c)
			competFat := ini.LerString(sc, "CompetFat", "FIM")
			if competFat == "FIM" || competFat == "" {
				break
			}
			cons := GCons{
				CompetFat: pcn.EncodeDataHoraDef(competFat),
				QtdDias:   ini.LerInteiro(sc, "qtdDias", 0),
				MedDiaria: ini.LerFloat(sc, "medDiaria", 0),
				Consumo:   ini.LerFloat(sc, "consumo", 0),
				VFat:      ini.LerFloat(sc, "vFat", 0),
			}
			cons.UMed, _ = ParseUMed(ini.LerString(sc, "uMed", ""))
			item.GCons = append(item.GCons, cons)
		}

		g.GHistCons = append(g.GHistCons, item)
	}
}

func lerINIAutXML(ini *pcn.INI, n *NFGas) {
	n.AutXML = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("autXML%02d", i)
		doc := pcn.OnlyCPFCNPJAlphaNum(ini.LerString(secao, "CNPJCPF", ""))
		if doc == "" {
			break
		}
		n.AutXML = append(n.AutXML, AutXML{CNPJCPF: doc})
	}
}

func lerINIInfAdic(ini *pcn.INI, i *InfAdic) {
	const secao = "infAdic"
	if !ini.SecaoExiste(secao) {
		return
	}
	i.InfAdFisco = ini.LerString(secao, "infAdFisco", "")
	i.InfCpl = nil
	if v := ini.LerString(secao, "infCpl", ""); v != "" {
		i.InfCpl = append(i.InfCpl, v)
	}
}

// lerINIInfRespTec le a secao [infRespTec] -- no .ini o nome da secao e
// infRespTec, e nao gRespTec como a tag do XML.
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

// time0 devolve o tempo zero, usado como default de data no .ini.
func time0() time.Time { return time.Time{} }
