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
// ESCOPO: a compra governamental nas duas formas (completa e reduzida), o
// pagamento antecipado nos dois niveis (ide e item), o IBSCBS do item com a
// arvore de gIBSCBS, gEstornoCred, gTransfCred, gAjusteCompet,
// gCredPresOper e gCredPresIBSZFM, o IBSCBSTot e o pagamento vinculado.
//
// A monofasia (gIBSCBSMono) e o Imposto Seletivo (gIS/ISTot/gMono) ficam em
// ini_mono.go -- entraram com a NF-e, que e o primeiro documento do porte a
// emiti-los no .ini.
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

// LerINIGCompraGov le o grupo gCompraGov COMPLETO -- o da NF-e.
// Porte de Ler_gCompraGov.
//
// Nao confundir com LerINIGCompraGovReduzido: os dois leem a MESMA secao
// [gCompraGov] e os MESMOS tres campos, e diferem apenas no tipo da lista
// de documentos anteriores. O ACBr tambem mantem os dois (TgCompraGov x
// TgCompraGovReduzido) -- BP-e, CT-e, NF3e, NFAg, NFCom e NFGas usam o
// reduzido; a NF-e usa este.
//
// A lista vem de [refDFeAntNNN], com a chave na propria chave `refDFeAnt`.
func LerINIGCompraGov(ini *pcn.INI, g *GCompraGov) {
	if ini == nil || g == nil {
		return
	}
	g.TpEnteGov, _ = ParseTpEnteGov(ini.LerString("gCompraGov", "tpEnteGov", ""))
	g.PRedutor = ini.LerFloat("gCompraGov", "pRedutor", 0)
	g.TpOperGov, _ = ParseTpOperGov(ini.LerString("gCompraGov", "tpOperGov", ""))

	g.RefDFeAnt = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("refDFeAnt%03d", i)
		v := ini.LerString(secao, "refDFeAnt", "FIM")
		if v == "FIM" {
			break
		}
		g.RefDFeAnt = append(g.RefDFeAnt, DFeRef{RefDFeChave: v})
	}
}

// LerINIGPagAntecipado le o gPagAntecipado do IDE -- a lista de notas que
// pagaram a operacao antecipadamente. Porte de Ler_gPagAntecipado.
//
// A secao indexa com DOIS digitos ([gPagAntecipado01]), e nao tres como a
// maioria dos grupos repetidos. A chave e `refNFe`.
//
// Nao confundir com LerINIGPagAntecipadoProd, que e o grupo do ITEM: aquele
// tem secao propria por item e carrega chave MAIS numero do item.
func LerINIGPagAntecipado(ini *pcn.INI, g *GPagAntecipado) {
	if ini == nil || g == nil {
		return
	}
	g.RefNFe = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("gPagAntecipado%02d", i)
		v := ini.LerString(secao, "refNFe", "FIM")
		if v == "FIM" || v == "" {
			break
		}
		g.RefNFe = append(g.RefNFe, RefDFePagAnt{RefDFeChave: v})
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

	// A monofasia e lida SEMPRE, qualquer que seja o CST -- o `case` que
	// escolhe entre gIBSCBS e gIBSCBSMono existe so na ESCRITA
	// (Gerar_IBSCBSNFe). Na leitura os dois grupos convivem, e e a
	// presenca da secao que decide.
	LerINIGIBSCBSMono(ini, &v.GIBSCBSMono, idx1)

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

	LerINIGMonoTot(ini, &v.GMono)

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

// GravarINIGCompraGov grava o gCompraGov COMPLETO -- o da NF-e.
// Porte de Gerar_gCompraGov.
//
// So sai com pRedutor POSITIVO, como no reduzido.
//
// ACRESCIMO: a lista de documentos anteriores tambem sai daqui. No ACBr o
// Gerar_DFeRef e uma procedure separada que NAO E CHAMADA POR NINGUEM --
// nem pelo Gerar_gCompraGov, nem pelo gerador da NF-e --, embora o
// Ler_gCompraGov LEIA a secao [refDFeAntNNN]. E mais uma assimetria do
// original: o grupo e lido e nunca gravado de volta.
func GravarINIGCompraGov(ini *pcn.INI, g GCompraGov) {
	if ini == nil || g.PRedutor <= 0 {
		return
	}
	ini.GravarString("gCompraGov", "tpEnteGov", g.TpEnteGov.String())
	ini.GravarFloat("gCompraGov", "pRedutor", g.PRedutor, 4)
	ini.GravarString("gCompraGov", "tpOperGov", g.TpOperGov.String())

	GravarINIDFeRef(ini, g.RefDFeAnt)
}

// GravarINIDFeRef grava [refDFeAntNNN]. Porte de Gerar_DFeRef.
func GravarINIDFeRef(ini *pcn.INI, refs []DFeRef) {
	if ini == nil {
		return
	}
	for i, r := range refs {
		ini.GravarString(fmt.Sprintf("refDFeAnt%03d", i+1), "refDFeAnt", r.RefDFeChave)
	}
}

// GravarINIGPagAntecipado grava o gPagAntecipado do IDE.
// Porte de Gerar_gPagAntecipado.
//
// A secao indexa com DOIS digitos e a chave e `refNFe` -- as duas coisas
// diferem do grupo homonimo do item.
func GravarINIGPagAntecipado(ini *pcn.INI, g GPagAntecipado) {
	if ini == nil {
		return
	}
	for i, r := range g.RefNFe {
		ini.GravarString(fmt.Sprintf("gPagAntecipado%02d", i+1), "refNFe", r.RefDFeChave)
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

// OpcoesIBSCBS carrega o que o Gerar_IBSCBS do ACBr le de PROPRIEDADES do
// writer, e nao do grupo que esta gravando.
//
// Sem isso o gravador nao tem como decidir o que sai: o subgrupo depende do
// MODELO do documento, e o gEstornoCred de uma NF-e depende do tpNFDebito.
type OpcoesIBSCBS struct {
	// Modelo e o documento em geracao (FModelosDFe).
	Modelo ModeloDFe

	// TpNFDebito so e consultado quando Modelo e ModeloNFe.
	TpNFDebito TpNFDebito
}

// GravarINIIBSCBS grava o grupo IBSCBS do item. Porte de Gerar_IBSCBS.
//
// Tres decisoes do original governam a escrita, e as tres dependem de algo
// FORA do grupo:
//
//  1. qual subgrupo sai -- gIBSCBS, gIBSCBSMono, gTransfCred ou
//     gAjusteCompet -- depende do modelo E do CST. A tabela esta em
//     GrupoDoCST, compartilhada com o gerador de XML;
//  2. o gEstornoCred sai com valor positivo OU, numa NF-e, quando o
//     tpNFDebito e "perda em estoque" -- mesmo zerado;
//  3. gCredPresOper e gCredPresIBSZFM so existem na NF-e, e sao
//     EXCLUDENTES: havendo o primeiro, o segundo nao sai.
//
// DIVERGENCIA: o ACBr condiciona o bloco inteiro -- cabecalho e subgrupo --
// a `IBSCBS.gIBSCBS.vBC > 0`. Um item monofasico (CST 620) costuma ter esse
// vBC zerado, porque a base dele vive no gIBSCBSMono, e com a condicao do
// original NADA seria gravado: nem o CST, nem a monofasia. O proprio ACBr
// nao usa essa condicao no gerador de XML, que testa CST e cClassTrib
// (Gerar_IBSCBS do XmlWriter) -- os dois lados dele discordam entre si. Aqui
// vale a condicao do XML, para que o .ini e o XML do mesmo documento
// contenham o mesmo grupo.
func GravarINIIBSCBS(ini *pcn.INI, v IBSCBS, idx1, idx2 int, op OpcoesIBSCBS) {
	if ini == nil {
		return
	}

	if v.CST != CSTNenhum || v.CClassTrib != "" {
		secao := NomeSecao("IBSCBS", idx1, idx2)
		ini.GravarString(secao, "CST", v.CST.String())
		ini.GravarString(secao, "cClassTrib", v.CClassTrib)
		ini.GravarString(secao, "indDoacao", v.IndDoacao.String())

		switch GrupoDoCST(op.Modelo, v.CST) {
		case GrupoIBSCBSNormal:
			gravarINIGIBSCBS(ini, v.GIBSCBS, idx1, idx2)
		case GrupoIBSCBSMono:
			GravarINIGIBSCBSMono(ini, v.GIBSCBSMono, idx1)
		case GrupoIBSCBSTransfCred:
			gravarINIGTransfCred(ini, v.GTransfCred, idx1)
		case GrupoIBSCBSAjusteCompet:
			gravarINIGAjusteCompet(ini, v.GAjusteCompet, idx1)
		}
	}

	if v.GEstornoCred.VIBSEstCred > 0 || v.GEstornoCred.VCBSEstCred > 0 ||
		(op.Modelo == ModeloNFe && op.TpNFDebito == TdPerdaEmEstoque) {
		s := NomeSecao("gEstornoCred", idx1, idx2)
		ini.GravarFloat(s, "vIBSEstCred", v.GEstornoCred.VIBSEstCred, 2)
		ini.GravarFloat(s, "vCBSEstCred", v.GEstornoCred.VCBSEstCred, 2)
	}

	if op.Modelo == ModeloNFe {
		if v.GCredPresOper.CCredPres != CpNenhum {
			gravarINIGCredPresOper(ini, v.GCredPresOper, idx1)
		} else if v.GCredPresIBSZFM.TpCredPresIBSZFM != TcpNenhum {
			gravarINIGCredPresIBSZFM(ini, v.GCredPresIBSZFM, idx1)
		}
	}
}

// gravarINIGIBSCBS grava o gIBSCBS e os grupos aninhados nele.
// Porte de Gerar_gIBSCBS e dos Gerar_gIBSUF/gIBSMun/gCBS que ele chama.
func gravarINIGIBSCBS(ini *pcn.INI, g GIBSCBS, idx1, idx2 int) {
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

// gravarINIGTransfCred grava [gTransfCredNNN]. Porte de Gerar_gTransfCred.
func gravarINIGTransfCred(ini *pcn.INI, g GTransfCred, idx int) {
	s := secaoIdx3("gTransfCred", idx)
	ini.GravarFloat(s, "vIBS", g.VIBS, 2)
	ini.GravarFloat(s, "vCBS", g.VCBS, 2)
}

// gravarINIGAjusteCompet grava [gAjusteCompetNNN].
// Porte de Gerar_gAjusteCompet.
func gravarINIGAjusteCompet(ini *pcn.INI, g GAjusteCompet, idx int) {
	s := secaoIdx3("gAjusteCompet", idx)
	if !g.CompetApur.IsZero() {
		ini.GravarString(s, "competApur", g.CompetApur.Format("2006-01-02"))
	}
	ini.GravarFloat(s, "vIBS", g.VIBS, 2)
	ini.GravarFloat(s, "vCBS", g.VCBS, 2)
}

// gravarINIGCredPresOper grava [gCredPresOperNNN] e os dois subgrupos.
// Porte de Gerar_gCredPresOper.
func gravarINIGCredPresOper(ini *pcn.INI, g GCredPresOper, idx int) {
	s := secaoIdx3("gCredPresOper", idx)
	ini.GravarFloat(s, "vBCCredPres", g.VBCCredPres, 2)
	ini.GravarString(s, "cCredPres", g.CCredPres.String())

	gravarINIGIBSCBSCredPres(ini, g.GIBSCredPres, "gIBSCredPres", idx)
	gravarINIGIBSCBSCredPres(ini, g.GCBSCredPres, "gCBSCredPres", idx)
}

// gravarINIGCredPresIBSZFM grava [gCredPresIBSZFMNNN].
// Porte de Gerar_gCredPresIBSZFM.
func gravarINIGCredPresIBSZFM(ini *pcn.INI, g CredPresIBSZFM, idx int) {
	s := secaoIdx3("gCredPresIBSZFM", idx)
	if !g.CompetApur.IsZero() {
		ini.GravarString(s, "competApur", g.CompetApur.Format("2006-01-02"))
	}
	ini.GravarString(s, "tpCredPresIBSZFM", g.TpCredPresIBSZFM.String())
	ini.GravarFloat(s, "vCredPresIBSZFM", g.VCredPresIBSZFM, 2)
}

// gravarINIGIBSCBSCredPres grava [gIBSCredPresNNN] ou [gCBSCredPresNNN] --
// o mesmo conjunto de campos, em secao diferente por tributo.
// Porte de Gerar_gIBSCBSCredPres.
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
		v.GMono == (GMono{}) && v.GEstornoCred == (GEstornoCred{}) {
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

	// ACRESCIMO: o Gerar_IBSCBSTot do ACBr NAO grava o gMono, embora o
	// Ler_IBSCBSTot o leia (:~). Uma ida e volta pelo .ini perderia o total
	// da monofasia inteiro. Gravado aqui -- o leitor do proprio ACBr
	// recupera a secao sem alteracao nenhuma do lado dele.
	GravarINIGMonoTot(ini, v.GMono)

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
