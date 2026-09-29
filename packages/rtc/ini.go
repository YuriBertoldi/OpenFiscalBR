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
	"fmt"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitura e escrita dos grupos da Reforma Tributaria no formato .ini do
// ACBr. Porte de ACBrDFe.RTC.IniReader.pas e ACBrDFe.RTC.IniWriter.pas.
//
// ESCOPO: estao portados os grupos que a NFGas usa -- compra governamental,
// pagamento antecipado do item, IBSCBS do item (arvore de gIBSCBS,
// gEstornoCred, gTransfCred, gAjusteCompet, gCredPresOper e
// gCredPresIBSZFM), IBSCBSTot e pagamento vinculado. Ficam FORA do .ini,
// lidos e escritos apenas em XML: a monofasia (gIBSCBSMono) e o Imposto
// Seletivo (gIS/ISTot). Quando outro componente precisar deles no .ini, e
// aqui que entram.
//
// Nome de secao: o ACBr usa tres formas, conforme os indices.
//
//	Idx1 == -1                -> "gIBSUF"
//	Idx1 >= 0 e Idx2 == -1    -> "gIBSUF" + 3 digitos de Idx1
//	Idx1 >= 0 e Idx2 >= 0     -> "gIBSUF" + 2 digitos de Idx1 + 3 de Idx2
//
// SemIndice e o valor a passar quando o grupo nao e indexado.
const SemIndice = -1

// NomeSecao monta o nome da secao conforme a convencao do ACBr.
func NomeSecao(nome string, idx1, idx2 int) string {
	if idx1 == SemIndice {
		return nome
	}
	if idx2 == SemIndice {
		return fmt.Sprintf("%s%03d", nome, idx1)
	}
	return fmt.Sprintf("%s%02d%03d", nome, idx1, idx2)
}

// ---------------------------------------------------------------------------
// Leitura
// ---------------------------------------------------------------------------

// LerINIGCompraGovReduzido le o grupo gCompraGov do .ini.
// Porte de Ler_gCompraGovReduzido.
func LerINIGCompraGovReduzido(ini *pcn.INI, g *GCompraGovReduzido) {
	if ini == nil || g == nil {
		return
	}
	g.TpEnteGov, _ = ParseTpEnteGov(ini.LerString("gCompraGov", "tpEnteGov", ""))
	g.PRedutor = ini.LerFloat("gCompraGov", "pRedutor", 0)
	g.TpOperGov, _ = ParseTpOperGov(ini.LerString("gCompraGov", "tpOperGov", ""))

	g.RefDFe = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("refDFeAnt%03d", i)
		v := ini.LerString(secao, "refDFeAnt", "FIM")
		if v == "FIM" {
			break
		}
		g.RefDFe = append(g.RefDFe, RefDFeAnt{RefDFeAnt: v})
	}
}

// LerINIGPagAntecipadoProd le o grupo gPagAntecipado do item.
// Porte de Ler_gPagAntecipadoProd.
//
// DIVERGENCIA: o original atribui chDFePagAnt a partir de uma variavel
// (sFim) que NUNCA e inicializada -- o campo sempre sai vazio, qualquer que
// seja o conteudo do .ini. Como o valor depende de variavel nao
// inicializada, o comportamento nao e replicavel com seguranca; aqui a
// chave e lida da chave 'chDFePagAnt', que e o que o gravador escreve.
func LerINIGPagAntecipadoProd(ini *pcn.INI, g *GPagAntecipadoProd, idx1, idx2 int) {
	if ini == nil || g == nil {
		return
	}
	secao := NomeSecao("gPagAntecipado", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	g.ChDFePagAnt = ini.LerString(secao, "chDFePagAnt", "")
	g.NItemPagAnt = ini.LerInteiro(secao, "nItemPagAnt", 0)
}

// LerINIIBSCBS le o grupo IBSCBS do item. Porte de Ler_IBSCBS.
func LerINIIBSCBS(ini *pcn.INI, v *IBSCBS, idx1, idx2 int) {
	if ini == nil || v == nil {
		return
	}
	secao := NomeSecao("IBSCBS", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}

	v.CST, _ = ParseCSTIBSCBS(ini.LerString(secao, "CST", ""))
	v.CClassTrib = ini.LerString(secao, "cClassTrib", "")
	v.IndDoacao, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDoacao", ""))

	lerINIGIBSCBS(ini, &v.GIBSCBS, idx1, idx2)
	lerINIGEstornoCred(ini, &v.GEstornoCred, idx1, idx2)

	// Estes quatro grupos usam apenas o primeiro indice, com 3 digitos --
	// e assim no TDFeRTCIniReader (Ler_gTransfCred etc. recebem so Idx).
	lerINIGTransfCred(ini, &v.GTransfCred, idx1)
	lerINIGAjusteCompet(ini, &v.GAjusteCompet, idx1)
	lerINIGCredPresOper(ini, &v.GCredPresOper, idx1)
	lerINIGCredPresIBSZFM(ini, &v.GCredPresIBSZFM, idx1)
}

// secaoIdx3 monta o nome de secao dos grupos que indexam so pelo primeiro
// indice: nome + 3 digitos, ou o nome puro quando nao ha indice.
func secaoIdx3(nome string, idx int) string {
	if idx == SemIndice {
		return nome
	}
	return fmt.Sprintf("%s%03d", nome, idx)
}

// lerINIGTransfCred le [gTransfCredNNN]. Porte de Ler_gTransfCred.
func lerINIGTransfCred(ini *pcn.INI, v *GTransfCred, idx int) {
	secao := secaoIdx3("gTransfCred", idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.VIBS = ini.LerFloat(secao, "vIBS", 0)
	v.VCBS = ini.LerFloat(secao, "vCBS", 0)
}

// lerINIGAjusteCompet le [gAjusteCompetNNN]. Porte de Ler_gAjusteCompet.
func lerINIGAjusteCompet(ini *pcn.INI, v *GAjusteCompet, idx int) {
	secao := secaoIdx3("gAjusteCompet", idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.CompetApur = ini.LerData(secao, "competApur", v.CompetApur)
	v.VIBS = ini.LerFloat(secao, "vIBS", 0)
	v.VCBS = ini.LerFloat(secao, "vCBS", 0)
}

// lerINIGCredPresOper le [gCredPresOperNNN] e os subgrupos
// [gIBSCredPresNNN]/[gCBSCredPresNNN]. Porte de Ler_gCredPresOper.
func lerINIGCredPresOper(ini *pcn.INI, v *GCredPresOper, idx int) {
	secao := secaoIdx3("gCredPresOper", idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.VBCCredPres = ini.LerFloat(secao, "vBCCredPres", 0)
	v.CCredPres, _ = ParseCCredPres(ini.LerString(secao, "cCredPres", ""))

	lerINIGIBSCBSCredPres(ini, &v.GIBSCredPres, "gIBSCredPres", idx)
	lerINIGIBSCBSCredPres(ini, &v.GCBSCredPres, "gCBSCredPres", idx)
}

// lerINIGIBSCBSCredPres le [gIBSCredPresNNN] ou [gCBSCredPresNNN].
// Porte de Ler_gIBSCBSCredPres.
func lerINIGIBSCBSCredPres(ini *pcn.INI, v *GIBSCBSCredPres, grupo string, idx int) {
	secao := secaoIdx3(grupo, idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.PCredPres = ini.LerFloat(secao, "pCredPres", 0)
	v.VCredPres = ini.LerFloat(secao, "vCredPres", 0)
	v.VCredPresCondSus = ini.LerFloat(secao, "vCredPresCondSus", 0)
}

// lerINIGCredPresIBSZFM le [gCredPresIBSZFMNNN].
// Porte de Ler_gCredPresIBSZFM.
func lerINIGCredPresIBSZFM(ini *pcn.INI, v *CredPresIBSZFM, idx int) {
	secao := secaoIdx3("gCredPresIBSZFM", idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.CompetApur = ini.LerData(secao, "competApur", v.CompetApur)
	v.TpCredPresIBSZFM, _ = ParseTpCredPresIBSZFM(ini.LerString(secao, "tpCredPresIBSZFM", ""))
	v.VCredPresIBSZFM = ini.LerFloat(secao, "vCredPresIBSZFM", 0)
}

func lerINIGIBSCBS(ini *pcn.INI, v *GIBSCBS, idx1, idx2 int) {
	secao := NomeSecao("gIBSCBS", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.VBC = ini.LerFloat(secao, "vBC", 0)
	v.VIBS = ini.LerFloat(secao, "vIBS", 0)

	lerINIGIBSUF(ini, &v.GIBSUF, idx1, idx2)
	lerINIGIBSMun(ini, &v.GIBSMun, idx1, idx2)
	lerINIGCBS(ini, &v.GCBS, idx1, idx2)
	lerINIGALCZFMCBS(ini, &v.GCBS.GALCZFMCBS, idx1, idx2)
	lerINIGTribRegular(ini, &v.GTribRegular, idx1, idx2)
	lerINIGTribCompraGov(ini, &v.GTribCompraGov, idx1, idx2)
}

func lerINIGIBSUF(ini *pcn.INI, v *GIBSUFValores, idx1, idx2 int) {
	secao := NomeSecao("gIBSUF", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.PIBSUF = ini.LerFloat(secao, "pIBSUF", 0)
	v.VIBSUF = ini.LerFloat(secao, "vIBSUF", 0)
	v.GDif.PDif = ini.LerFloat(secao, "pDif", 0)
	v.GDif.VDif = ini.LerFloat(secao, "vDif", 0)
	v.GDevTrib.PDevTrib = ini.LerFloat(secao, "pDevTrib", 0)
	v.GDevTrib.VDevTrib = ini.LerFloat(secao, "vDevTrib", 0)
	v.GRed.PRedAliq = ini.LerFloat(secao, "pRedAliq", 0)
	v.GRed.PAliqEfet = ini.LerFloat(secao, "pAliqEfet", 0)
}

// lerINIGIBSMun le gIBSMun do .ini.
//
// OMISSAO DO ACBr REPLICADA: pDevTrib NAO e lido aqui, embora seja lido em
// gIBSUF e exista na struct. Vale o mesmo para gCBS.
func lerINIGIBSMun(ini *pcn.INI, v *GIBSMunValores, idx1, idx2 int) {
	secao := NomeSecao("gIBSMun", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.PIBSMun = ini.LerFloat(secao, "pIBSMun", 0)
	v.VIBSMun = ini.LerFloat(secao, "vIBSMun", 0)
	v.GDif.PDif = ini.LerFloat(secao, "pDif", 0)
	v.GDif.VDif = ini.LerFloat(secao, "vDif", 0)
	v.GDevTrib.VDevTrib = ini.LerFloat(secao, "vDevTrib", 0)
	v.GRed.PRedAliq = ini.LerFloat(secao, "pRedAliq", 0)
	v.GRed.PAliqEfet = ini.LerFloat(secao, "pAliqEfet", 0)
}

func lerINIGCBS(ini *pcn.INI, v *GCBSValores, idx1, idx2 int) {
	secao := NomeSecao("gCBS", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.PCBS = ini.LerFloat(secao, "pCBS", 0)
	v.VCBS = ini.LerFloat(secao, "vCBS", 0)
	v.GDif.PDif = ini.LerFloat(secao, "pDif", 0)
	v.GDif.VDif = ini.LerFloat(secao, "vDif", 0)
	v.GDevTrib.VDevTrib = ini.LerFloat(secao, "vDevTrib", 0)
	v.GRed.PRedAliq = ini.LerFloat(secao, "pRedAliq", 0)
	v.GRed.PAliqEfet = ini.LerFloat(secao, "pAliqEfet", 0)
}

func lerINIGALCZFMCBS(ini *pcn.INI, v *GALCZFMCBS, idx1, idx2 int) {
	secao := NomeSecao("gALCZFMCBS", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.NProcSuframa = ini.LerString(secao, "nProcSuframa", "")
	v.PAliqEfetRegCBS = ini.LerFloat(secao, "pAliqEfetRegCBS", 0)
	// tpALCZFMCBS so e atribuido quando a chave tem conteudo -- o enum nao
	// tem membro para vazio.
	if s := ini.LerString(secao, "tpALCZFMCBS", ""); s != "" {
		v.TpALCZFMCBS, _ = ParseTpALCZFMCBS(s)
	}
	v.VTribRegCBS = ini.LerFloat(secao, "vTribRegCBS", 0)
}

func lerINIGTribRegular(ini *pcn.INI, v *GTribRegular, idx1, idx2 int) {
	secao := NomeSecao("gTribRegular", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.CSTReg, _ = ParseCSTIBSCBS(ini.LerString(secao, "CSTReg", ""))
	v.CClassTribReg = ini.LerString(secao, "cClassTribReg", "")
	v.PAliqEfetRegIBSUF = ini.LerFloat(secao, "pAliqEfetRegIBSUF", 0)
	v.VTribRegIBSUF = ini.LerFloat(secao, "vTribRegIBSUF", 0)
	v.PAliqEfetRegIBSMun = ini.LerFloat(secao, "pAliqEfetRegIBSMun", 0)
	v.VTribRegIBSMun = ini.LerFloat(secao, "vTribRegIBSMun", 0)
	v.PAliqEfetRegCBS = ini.LerFloat(secao, "pAliqEfetRegCBS", 0)
	v.VTribRegCBS = ini.LerFloat(secao, "vTribRegCBS", 0)
}

func lerINIGTribCompraGov(ini *pcn.INI, v *GTribCompraGov, idx1, idx2 int) {
	secao := NomeSecao("gTribCompraGov", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.PAliqIBSUF = ini.LerFloat(secao, "pAliqIBSUF", 0)
	v.VTribIBSUF = ini.LerFloat(secao, "vTribIBSUF", 0)
	v.PAliqIBSMun = ini.LerFloat(secao, "pAliqIBSMun", 0)
	v.VTribIBSMun = ini.LerFloat(secao, "vTribIBSMun", 0)
	v.PAliqCBS = ini.LerFloat(secao, "pAliqCBS", 0)
	v.VTribCBS = ini.LerFloat(secao, "vTribCBS", 0)
}

func lerINIGEstornoCred(ini *pcn.INI, v *GEstornoCred, idx1, idx2 int) {
	secao := NomeSecao("gEstornoCred", idx1, idx2)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.VIBSEstCred = ini.LerFloat(secao, "vIBSEstCred", 0)
	v.VCBSEstCred = ini.LerFloat(secao, "vCBSEstCred", 0)
}

// LerINIIBSCBSTot le o grupo IBSCBSTot do .ini. Porte de Ler_IBSCBSTot.
//
// Note que as secoes do total tem sufixo Tot -- gIBSUFTot, gIBSMunTot,
// gCBSTot, gEstornoCredTot --, diferente das do item.
func LerINIIBSCBSTot(ini *pcn.INI, v *IBSCBSTot) {
	if ini == nil || v == nil {
		return
	}
	if !ini.SecaoExiste("IBSCBSTot") {
		return
	}
	v.VBCIBSCBS = ini.LerFloat("IBSCBSTot", "vBCIBSCBS", 0)

	if ini.SecaoExiste("gIBS") {
		v.GIBS.VIBS = ini.LerFloat("gIBS", "vIBS", 0)
		v.GIBS.VCredPres = ini.LerFloat("gIBS", "vCredPres", 0)
		v.GIBS.VCredPresCondSus = ini.LerFloat("gIBS", "vCredPresCondSus", 0)

		if ini.SecaoExiste("gIBSUFTot") {
			v.GIBS.GIBSUFTot.VDif = ini.LerFloat("gIBSUFTot", "vDif", 0)
			v.GIBS.GIBSUFTot.VDevTrib = ini.LerFloat("gIBSUFTot", "vDevTrib", 0)
			v.GIBS.GIBSUFTot.VIBSUF = ini.LerFloat("gIBSUFTot", "vIBSUF", 0)
		}
		if ini.SecaoExiste("gIBSMunTot") {
			v.GIBS.GIBSMunTot.VDif = ini.LerFloat("gIBSMunTot", "vDif", 0)
			v.GIBS.GIBSMunTot.VDevTrib = ini.LerFloat("gIBSMunTot", "vDevTrib", 0)
			v.GIBS.GIBSMunTot.VIBSMun = ini.LerFloat("gIBSMunTot", "vIBSMun", 0)
		}
	}

	if ini.SecaoExiste("gCBSTot") {
		v.GCBS.VDif = ini.LerFloat("gCBSTot", "vDif", 0)
		v.GCBS.VDevTrib = ini.LerFloat("gCBSTot", "vDevTrib", 0)
		v.GCBS.VCBS = ini.LerFloat("gCBSTot", "vCBS", 0)
		v.GCBS.VCredPres = ini.LerFloat("gCBSTot", "vCredPres", 0)
		v.GCBS.VCredPresCondSus = ini.LerFloat("gCBSTot", "vCredPresCondSus", 0)
	}

	if ini.SecaoExiste("gEstornoCredTot") {
		v.GEstornoCred.VIBSEstCred = ini.LerFloat("gEstornoCredTot", "vIBSEstCred", 0)
		v.GEstornoCred.VCBSEstCred = ini.LerFloat("gEstornoCredTot", "vCBSEstCred", 0)
	}
}

// LerINIPgtoVinc le as secoes [pgtoVincNN]. Porte de Ler_pgtoVinc.
func LerINIPgtoVinc(ini *pcn.INI, v *PgtoVinc) {
	if ini == nil || v == nil {
		return
	}
	v.Pgto = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("pgtoVinc%02d", i)
		nPag := ini.LerString(secao, "nPag", "FIM")
		if nPag == "FIM" || nPag == "" {
			break
		}
		v.Pgto = append(v.Pgto, Pgto{
			NPag:        ini.LerInteiro(secao, "nPag", 0),
			IDTransacao: ini.LerString(secao, "idTransacao", ""),
			TpMeioPgto:  ini.LerString(secao, "tpMeioPgto", ""),
			CNPJReceb:   ini.LerString(secao, "CNPJReceb", ""),
			CNPJBasePSP: ini.LerString(secao, "CNPJBasePSP", ""),
		})
	}
}

// ---------------------------------------------------------------------------
// Escrita
// ---------------------------------------------------------------------------

// GravarINIGCompraGovReduzido grava o grupo gCompraGov no .ini.
func GravarINIGCompraGovReduzido(ini *pcn.INI, g GCompraGovReduzido) {
	if ini == nil {
		return
	}
	if g.TpEnteGov == TcgNenhum && g.TpOperGov == TogNenhum &&
		g.PRedutor == 0 && len(g.RefDFe) == 0 {
		return
	}
	ini.GravarString("gCompraGov", "tpEnteGov", g.TpEnteGov.String())
	ini.GravarFloat("gCompraGov", "pRedutor", g.PRedutor, 4)
	ini.GravarString("gCompraGov", "tpOperGov", g.TpOperGov.String())

	for i, r := range g.RefDFe {
		ini.GravarString(fmt.Sprintf("refDFeAnt%03d", i+1), "refDFeAnt", r.RefDFeAnt)
	}
}

// GravarINIGPagAntecipadoProd grava o grupo gPagAntecipado do item.
func GravarINIGPagAntecipadoProd(ini *pcn.INI, g GPagAntecipadoProd, idx1, idx2 int) {
	if ini == nil || g == (GPagAntecipadoProd{}) {
		return
	}
	secao := NomeSecao("gPagAntecipado", idx1, idx2)
	ini.GravarString(secao, "chDFePagAnt", g.ChDFePagAnt)
	ini.GravarInteiro(secao, "nItemPagAnt", g.NItemPagAnt)
}

// GravarINIIBSCBS grava o grupo IBSCBS do item.
func GravarINIIBSCBS(ini *pcn.INI, v IBSCBS, idx1, idx2 int) {
	if ini == nil {
		return
	}
	if v.CST == CSTNenhum && v.CClassTrib == "" && v.GIBSCBS == (GIBSCBS{}) {
		return
	}

	secao := NomeSecao("IBSCBS", idx1, idx2)
	ini.GravarString(secao, "CST", v.CST.String())
	ini.GravarString(secao, "cClassTrib", v.CClassTrib)
	ini.GravarString(secao, "indDoacao", v.IndDoacao.String())

	g := v.GIBSCBS
	if g != (GIBSCBS{}) {
		s := NomeSecao("gIBSCBS", idx1, idx2)
		ini.GravarFloat(s, "vBC", g.VBC, 2)
		ini.GravarFloat(s, "vIBS", g.VIBS, 2)

		if g.GIBSUF != (GIBSUFValores{}) {
			s := NomeSecao("gIBSUF", idx1, idx2)
			ini.GravarFloat(s, "pIBSUF", g.GIBSUF.PIBSUF, 4)
			ini.GravarFloat(s, "vIBSUF", g.GIBSUF.VIBSUF, 2)
			ini.GravarFloat(s, "pDif", g.GIBSUF.GDif.PDif, 4)
			ini.GravarFloat(s, "vDif", g.GIBSUF.GDif.VDif, 2)
			ini.GravarFloat(s, "pDevTrib", g.GIBSUF.GDevTrib.PDevTrib, 4)
			ini.GravarFloat(s, "vDevTrib", g.GIBSUF.GDevTrib.VDevTrib, 2)
			ini.GravarFloat(s, "pRedAliq", g.GIBSUF.GRed.PRedAliq, 4)
			ini.GravarFloat(s, "pAliqEfet", g.GIBSUF.GRed.PAliqEfet, 2)
		}
		if g.GIBSMun != (GIBSMunValores{}) {
			s := NomeSecao("gIBSMun", idx1, idx2)
			ini.GravarFloat(s, "pIBSMun", g.GIBSMun.PIBSMun, 4)
			ini.GravarFloat(s, "vIBSMun", g.GIBSMun.VIBSMun, 2)
			ini.GravarFloat(s, "pDif", g.GIBSMun.GDif.PDif, 4)
			ini.GravarFloat(s, "vDif", g.GIBSMun.GDif.VDif, 2)
			ini.GravarFloat(s, "vDevTrib", g.GIBSMun.GDevTrib.VDevTrib, 2)
			ini.GravarFloat(s, "pRedAliq", g.GIBSMun.GRed.PRedAliq, 4)
			ini.GravarFloat(s, "pAliqEfet", g.GIBSMun.GRed.PAliqEfet, 2)
		}
		if g.GCBS != (GCBSValores{}) {
			s := NomeSecao("gCBS", idx1, idx2)
			ini.GravarFloat(s, "pCBS", g.GCBS.PCBS, 4)
			ini.GravarFloat(s, "vCBS", g.GCBS.VCBS, 2)
			ini.GravarFloat(s, "pDif", g.GCBS.GDif.PDif, 4)
			ini.GravarFloat(s, "vDif", g.GCBS.GDif.VDif, 2)
			ini.GravarFloat(s, "vDevTrib", g.GCBS.GDevTrib.VDevTrib, 2)
			ini.GravarFloat(s, "pRedAliq", g.GCBS.GRed.PRedAliq, 4)
			ini.GravarFloat(s, "pAliqEfet", g.GCBS.GRed.PAliqEfet, 2)

			if g.GCBS.GALCZFMCBS != (GALCZFMCBS{}) {
				s := NomeSecao("gALCZFMCBS", idx1, idx2)
				ini.GravarString(s, "tpALCZFMCBS", g.GCBS.GALCZFMCBS.TpALCZFMCBS.String())
				ini.GravarString(s, "nProcSuframa", g.GCBS.GALCZFMCBS.NProcSuframa)
				ini.GravarFloat(s, "pAliqEfetRegCBS", g.GCBS.GALCZFMCBS.PAliqEfetRegCBS, 4)
				ini.GravarFloat(s, "vTribRegCBS", g.GCBS.GALCZFMCBS.VTribRegCBS, 2)
			}
		}
		if g.GTribRegular != (GTribRegular{}) {
			s := NomeSecao("gTribRegular", idx1, idx2)
			ini.GravarString(s, "CSTReg", g.GTribRegular.CSTReg.String())
			ini.GravarString(s, "cClassTribReg", g.GTribRegular.CClassTribReg)
			ini.GravarFloat(s, "pAliqEfetRegIBSUF", g.GTribRegular.PAliqEfetRegIBSUF, 4)
			ini.GravarFloat(s, "vTribRegIBSUF", g.GTribRegular.VTribRegIBSUF, 2)
			ini.GravarFloat(s, "pAliqEfetRegIBSMun", g.GTribRegular.PAliqEfetRegIBSMun, 4)
			ini.GravarFloat(s, "vTribRegIBSMun", g.GTribRegular.VTribRegIBSMun, 2)
			ini.GravarFloat(s, "pAliqEfetRegCBS", g.GTribRegular.PAliqEfetRegCBS, 4)
			ini.GravarFloat(s, "vTribRegCBS", g.GTribRegular.VTribRegCBS, 2)
		}
		if g.GTribCompraGov != (GTribCompraGov{}) {
			s := NomeSecao("gTribCompraGov", idx1, idx2)
			ini.GravarFloat(s, "pAliqIBSUF", g.GTribCompraGov.PAliqIBSUF, 4)
			ini.GravarFloat(s, "vTribIBSUF", g.GTribCompraGov.VTribIBSUF, 2)
			ini.GravarFloat(s, "pAliqIBSMun", g.GTribCompraGov.PAliqIBSMun, 4)
			ini.GravarFloat(s, "vTribIBSMun", g.GTribCompraGov.VTribIBSMun, 2)
			ini.GravarFloat(s, "pAliqCBS", g.GTribCompraGov.PAliqCBS, 4)
			ini.GravarFloat(s, "vTribCBS", g.GTribCompraGov.VTribCBS, 2)
		}
	}

	if v.GEstornoCred != (GEstornoCred{}) {
		s := NomeSecao("gEstornoCred", idx1, idx2)
		ini.GravarFloat(s, "vIBSEstCred", v.GEstornoCred.VIBSEstCred, 2)
		ini.GravarFloat(s, "vCBSEstCred", v.GEstornoCred.VCBSEstCred, 2)
	}

	if v.GTransfCred != (GTransfCred{}) {
		s := secaoIdx3("gTransfCred", idx1)
		ini.GravarFloat(s, "vIBS", v.GTransfCred.VIBS, 2)
		ini.GravarFloat(s, "vCBS", v.GTransfCred.VCBS, 2)
	}

	if v.GAjusteCompet != (GAjusteCompet{}) {
		s := secaoIdx3("gAjusteCompet", idx1)
		if !v.GAjusteCompet.CompetApur.IsZero() {
			ini.GravarString(s, "competApur", v.GAjusteCompet.CompetApur.Format("2006-01-02"))
		}
		ini.GravarFloat(s, "vIBS", v.GAjusteCompet.VIBS, 2)
		ini.GravarFloat(s, "vCBS", v.GAjusteCompet.VCBS, 2)
	}

	if v.GCredPresOper != (GCredPresOper{}) {
		s := secaoIdx3("gCredPresOper", idx1)
		ini.GravarFloat(s, "vBCCredPres", v.GCredPresOper.VBCCredPres, 2)
		ini.GravarString(s, "cCredPres", v.GCredPresOper.CCredPres.String())

		gravarINIGIBSCBSCredPres(ini, v.GCredPresOper.GIBSCredPres, "gIBSCredPres", idx1)
		gravarINIGIBSCBSCredPres(ini, v.GCredPresOper.GCBSCredPres, "gCBSCredPres", idx1)
	}

	if v.GCredPresIBSZFM != (CredPresIBSZFM{}) {
		s := secaoIdx3("gCredPresIBSZFM", idx1)
		if !v.GCredPresIBSZFM.CompetApur.IsZero() {
			ini.GravarString(s, "competApur", v.GCredPresIBSZFM.CompetApur.Format("2006-01-02"))
		}
		ini.GravarString(s, "tpCredPresIBSZFM", v.GCredPresIBSZFM.TpCredPresIBSZFM.String())
		ini.GravarFloat(s, "vCredPresIBSZFM", v.GCredPresIBSZFM.VCredPresIBSZFM, 2)
	}
}

func gravarINIGIBSCBSCredPres(ini *pcn.INI, v GIBSCBSCredPres, grupo string, idx int) {
	if v == (GIBSCBSCredPres{}) {
		return
	}
	s := secaoIdx3(grupo, idx)
	ini.GravarFloat(s, "pCredPres", v.PCredPres, 4)
	ini.GravarFloat(s, "vCredPres", v.VCredPres, 2)
	ini.GravarFloat(s, "vCredPresCondSus", v.VCredPresCondSus, 2)
}

// GravarINIIBSCBSTot grava o grupo IBSCBSTot no .ini.
func GravarINIIBSCBSTot(ini *pcn.INI, v IBSCBSTot) {
	if ini == nil {
		return
	}
	if v.VBCIBSCBS == 0 && v.GIBS == (GIBS{}) && v.GCBS == (GCBS{}) &&
		v.GEstornoCred == (GEstornoCred{}) {
		return
	}

	ini.GravarFloat("IBSCBSTot", "vBCIBSCBS", v.VBCIBSCBS, 2)

	if v.GIBS != (GIBS{}) {
		ini.GravarFloat("gIBS", "vIBS", v.GIBS.VIBS, 2)
		ini.GravarFloat("gIBS", "vCredPres", v.GIBS.VCredPres, 2)
		ini.GravarFloat("gIBS", "vCredPresCondSus", v.GIBS.VCredPresCondSus, 2)

		if v.GIBS.GIBSUFTot != (GIBSUFTot{}) {
			ini.GravarFloat("gIBSUFTot", "vDif", v.GIBS.GIBSUFTot.VDif, 2)
			ini.GravarFloat("gIBSUFTot", "vDevTrib", v.GIBS.GIBSUFTot.VDevTrib, 2)
			ini.GravarFloat("gIBSUFTot", "vIBSUF", v.GIBS.GIBSUFTot.VIBSUF, 2)
		}
		if v.GIBS.GIBSMunTot != (GIBSMunTot{}) {
			ini.GravarFloat("gIBSMunTot", "vDif", v.GIBS.GIBSMunTot.VDif, 2)
			ini.GravarFloat("gIBSMunTot", "vDevTrib", v.GIBS.GIBSMunTot.VDevTrib, 2)
			ini.GravarFloat("gIBSMunTot", "vIBSMun", v.GIBS.GIBSMunTot.VIBSMun, 2)
		}
	}

	if v.GCBS != (GCBS{}) {
		ini.GravarFloat("gCBSTot", "vDif", v.GCBS.VDif, 2)
		ini.GravarFloat("gCBSTot", "vDevTrib", v.GCBS.VDevTrib, 2)
		ini.GravarFloat("gCBSTot", "vCBS", v.GCBS.VCBS, 2)
		ini.GravarFloat("gCBSTot", "vCredPres", v.GCBS.VCredPres, 2)
		ini.GravarFloat("gCBSTot", "vCredPresCondSus", v.GCBS.VCredPresCondSus, 2)
	}

	if v.GEstornoCred != (GEstornoCred{}) {
		ini.GravarFloat("gEstornoCredTot", "vIBSEstCred", v.GEstornoCred.VIBSEstCred, 2)
		ini.GravarFloat("gEstornoCredTot", "vCBSEstCred", v.GEstornoCred.VCBSEstCred, 2)
	}
}

// GravarINIPgtoVinc grava as secoes [pgtoVincNN].
func GravarINIPgtoVinc(ini *pcn.INI, v PgtoVinc) {
	if ini == nil {
		return
	}
	for i, p := range v.Pgto {
		secao := fmt.Sprintf("pgtoVinc%02d", i+1)
		ini.GravarInteiro(secao, "nPag", p.NPag)
		ini.GravarString(secao, "idTransacao", p.IDTransacao)
		ini.GravarString(secao, "tpMeioPgto", p.TpMeioPgto)
		ini.GravarString(secao, "CNPJReceb", p.CNPJReceb)
		ini.GravarString(secao, "CNPJBasePSP", p.CNPJBasePSP)
	}
}
