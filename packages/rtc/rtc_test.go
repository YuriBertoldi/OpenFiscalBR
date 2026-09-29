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
	"errors"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

func parsear(t *testing.T, xml string) *pcn.Node {
	t.Helper()
	doc, err := pcn.ParseString(xml)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	return doc.Root
}

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

func TestEnums_CodigoDoLeiaute(t *testing.T) {
	casos := []struct {
		nome     string
		got      string
		esperado string
	}{
		{"TcgNenhum", TcgNenhum.String(), ""},
		{"TcgUniao", TcgUniao.String(), "1"},
		{"TcgComiteGestorIBS", TcgComiteGestorIBS.String(), "6"},
		{"TogNenhum", TogNenhum.String(), ""},
		{"TogRecebimentoPagFornecPosterior", TogRecebimentoPagFornecPosterior.String(), "4"},
		{"CSTNenhum", CSTNenhum.String(), ""},
		{"CST000", CST000.String(), "000"},
		{"CST830", CST830.String(), "830"},
		{"TpALCZFMCBSnOpInd", TpALCZFMCBSnOpInd.String(), "1"},
		{"TpALCZFMCBSOpInd", TpALCZFMCBSOpInd.String(), "2"},
		{"CpNenhum", CpNenhum.String(), ""},
		{"Cp13", Cp13.String(), "13"},
		{"TcpNenhum", TcpNenhum.String(), ""},
		{"TcpSemCredito", TcpSemCredito.String(), "0"},
		{"TcpBensInformaticaOutros", TcpBensInformaticaOutros.String(), "4"},
	}
	for _, c := range casos {
		if c.got != c.esperado {
			t.Errorf("%s = %q, esperado %q", c.nome, c.got, c.esperado)
		}
	}
}

func TestEnums_RoundTrip(t *testing.T) {
	for i := TcgNenhum; i <= TcgComiteGestorIBS; i++ {
		if v, err := ParseTpEnteGov(i.String()); err != nil || v != i {
			t.Errorf("TpEnteGov %d: %v %v", i, v, err)
		}
	}
	for i := TogNenhum; i <= TogRecebimentoPagFornecPosterior; i++ {
		if v, err := ParseTpOperGov(i.String()); err != nil || v != i {
			t.Errorf("TpOperGov %d: %v %v", i, v, err)
		}
	}
	for i := CSTNenhum; i <= CST830; i++ {
		if v, err := ParseCSTIBSCBS(i.String()); err != nil || v != i {
			t.Errorf("CSTIBSCBS %d: %v %v", i, v, err)
		}
	}
	for i := CpNenhum; i <= Cp13; i++ {
		if v, err := ParseCCredPres(i.String()); err != nil || v != i {
			t.Errorf("CCredPres %d: %v %v", i, v, err)
		}
	}
	for i := TcpNenhum; i <= TcpBensInformaticaOutros; i++ {
		if v, err := ParseTpCredPresIBSZFM(i.String()); err != nil || v != i {
			t.Errorf("TpCredPresIBSZFM %d: %v %v", i, v, err)
		}
	}
	for i := TpALCZFMCBSnOpInd; i <= TpALCZFMCBSOpInd; i++ {
		if v, err := ParseTpALCZFMCBS(i.String()); err != nil || v != i {
			t.Errorf("TpALCZFMCBS %d: %v %v", i, v, err)
		}
	}
}

func TestTpCredPresIBSZFM_DeslocamentoDoZero(t *testing.T) {
	// O membro "sem credito" vale "0", nao "1": ha um membro "nenhum" antes
	// dele, que vale string vazia. Confundir os dois desloca a tabela toda.
	v, err := ParseTpCredPresIBSZFM("0")
	if err != nil {
		t.Fatal(err)
	}
	if v != TcpSemCredito {
		t.Errorf("Parse(\"0\") = %d, esperado TcpSemCredito (%d)", v, TcpSemCredito)
	}
	if v, _ := ParseTpCredPresIBSZFM(""); v != TcpNenhum {
		t.Errorf("Parse(\"\") = %d, esperado TcpNenhum", v)
	}
}

func TestTpALCZFMCBS_NaoTemMembroVazio(t *testing.T) {
	// Diferente dos demais enums da RTC, este comeca em "1".
	if _, err := ParseTpALCZFMCBS(""); !errors.Is(err, ErrEnumInvalido) {
		t.Error("TpALCZFMCBS nao deveria aceitar string vazia")
	}
}

func TestEnums_ParseInvalido(t *testing.T) {
	if _, err := ParseCSTIBSCBS("999"); !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("erro = %v, esperado ErrEnumInvalido", err)
	}
	if _, err := ParseCCredPres("99"); !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("erro = %v, esperado ErrEnumInvalido", err)
	}
}

// ---------------------------------------------------------------------------
// Compra governamental
// ---------------------------------------------------------------------------

func TestLerGCompraGovReduzido(t *testing.T) {
	node := parsear(t, `<gCompraGov>
		<tpEnteGov>2</tpEnteGov>
		<pRedutor>12.3456</pRedutor>
		<tpOperGov>3</tpOperGov>
		<refDFeAnt>CHAVE1</refDFeAnt>
		<refDFeAnt>CHAVE2</refDFeAnt>
	</gCompraGov>`)

	var g GCompraGovReduzido
	LerGCompraGovReduzido(node, &g)

	if g.TpEnteGov != TcgEstados {
		t.Errorf("tpEnteGov = %d", g.TpEnteGov)
	}
	if g.PRedutor != 12.3456 {
		t.Errorf("pRedutor = %v", g.PRedutor)
	}
	if g.TpOperGov != TogFornecimentoPagRealizado {
		t.Errorf("tpOperGov = %d", g.TpOperGov)
	}
	if len(g.RefDFe) != 2 || g.RefDFe[0].RefDFeAnt != "CHAVE1" || g.RefDFe[1].RefDFeAnt != "CHAVE2" {
		t.Errorf("refDFe = %+v", g.RefDFe)
	}
}

func TestLerGCompraGovReduzido_LimpaListaAntesDoLaco(t *testing.T) {
	node := parsear(t, `<gCompraGov><refDFeAnt>NOVA</refDFeAnt></gCompraGov>`)

	g := GCompraGovReduzido{RefDFe: []RefDFeAnt{{RefDFeAnt: "ANTIGA"}}}
	LerGCompraGovReduzido(node, &g)

	if len(g.RefDFe) != 1 || g.RefDFe[0].RefDFeAnt != "NOVA" {
		t.Errorf("a lista deveria ser reinicializada, ficou %+v", g.RefDFe)
	}
}

func TestLerGCompraGov_FormaCompletaNFe(t *testing.T) {
	node := parsear(t, `<gCompraGov>
		<tpEnteGov>1</tpEnteGov><pRedutor>5.5</pRedutor><tpOperGov>1</tpOperGov>
		<refDFeAnt>A</refDFeAnt>
	</gCompraGov>`)

	var g GCompraGov
	LerGCompraGov(node, &g)

	if g.TpEnteGov != TcgUniao || g.TpOperGov != TogFornecimento {
		t.Errorf("enums = %d/%d", g.TpEnteGov, g.TpOperGov)
	}
	if len(g.RefDFeAnt) != 1 || g.RefDFeAnt[0].RefDFeChave != "A" {
		t.Errorf("refDFeAnt = %+v", g.RefDFeAnt)
	}
}

// ---------------------------------------------------------------------------
// Pagamento antecipado
// ---------------------------------------------------------------------------

func TestLerGPagAntecipadoProd_UsaFindAnyNs(t *testing.T) {
	// Unico leitor da RTC que usa FindAnyNs -- funciona com prefixo.
	node := parsear(t, `<gPagAntecipado xmlns:x="urn:teste">
		<x:chDFePagAnt>CHAVE</x:chDFePagAnt><x:nItemPagAnt>7</x:nItemPagAnt>
	</gPagAntecipado>`)

	var g GPagAntecipadoProd
	LerGPagAntecipadoProd(node, &g)

	if g.ChDFePagAnt != "CHAVE" || g.NItemPagAnt != 7 {
		t.Errorf("FindAnyNs deveria achar mesmo com prefixo: %+v", g)
	}
}

func TestLerGPagAntecipadoIde_ELerGPagAntecipado(t *testing.T) {
	node := parsear(t, `<gPagAntecipado><chDFePagAnt>A</chDFePagAnt><chDFePagAnt>B</chDFePagAnt></gPagAntecipado>`)
	var g GPagAntecipado
	LerGPagAntecipadoIde(node, &g)
	if len(g.RefNFe) != 2 || g.RefNFe[1].RefDFeChave != "B" {
		t.Errorf("ide: refNFe = %+v", g.RefNFe)
	}

	// A variante da NFe le refNFe em vez de chDFePagAnt.
	node = parsear(t, `<gPagAntecipado><refNFe>X</refNFe></gPagAntecipado>`)
	var g2 GPagAntecipado
	LerGPagAntecipado(node, &g2)
	if len(g2.RefNFe) != 1 || g2.RefNFe[0].RefDFeChave != "X" {
		t.Errorf("NFe: refNFe = %+v", g2.RefNFe)
	}
}

// ---------------------------------------------------------------------------
// IBS/CBS do item
// ---------------------------------------------------------------------------

const xmlIBSCBS = `<IBSCBS>
	<CST>200</CST>
	<cClassTrib>000001</cClassTrib>
	<indDoacao>1</indDoacao>
	<gIBSCBS>
		<vBC>1000.00</vBC>
		<vIBS>85.00</vIBS>
		<gIBSUF>
			<pIBSUF>0.1234</pIBSUF>
			<gDif><pDif>1.2345</pDif><vDif>10.00</vDif></gDif>
			<gDevTrib><pDevTrib>2.3456</pDevTrib><vDevTrib>20.00</vDevTrib></gDevTrib>
			<gRed><pRedAliq>3.4567</pRedAliq><pAliqEfet>4.56</pAliqEfet></gRed>
			<vIBSUF>30.00</vIBSUF>
		</gIBSUF>
		<gIBSMun>
			<pIBSMun>0.5678</pIBSMun>
			<gDif><pDif>5.6789</pDif><vDif>11.00</vDif></gDif>
			<gDevTrib><pDevTrib>6.7890</pDevTrib><vDevTrib>21.00</vDevTrib></gDevTrib>
			<gRed><pRedAliq>7.8901</pRedAliq><pAliqEfet>5.67</pAliqEfet></gRed>
			<vIBSMun>31.00</vIBSMun>
		</gIBSMun>
		<gCBS>
			<pCBS>0.9012</pCBS>
			<gDif><pDif>8.9012</pDif><vDif>12.00</vDif></gDif>
			<gDevTrib><pDevTrib>9.0123</pDevTrib><vDevTrib>22.00</vDevTrib></gDevTrib>
			<gRed><pRedAliq>1.1111</pRedAliq><pAliqEfet>6.78</pAliqEfet></gRed>
			<gALCZFMCBS>
				<tpALCZFMCBS>2</tpALCZFMCBS>
				<nProcSuframa>PROC123</nProcSuframa>
				<pAliqEfetRegCBS>2.2222</pAliqEfetRegCBS>
				<vTribRegCBS>33.00</vTribRegCBS>
			</gALCZFMCBS>
			<vCBS>32.00</vCBS>
		</gCBS>
		<gTribRegular>
			<CSTReg>000</CSTReg>
			<cClassTribReg>000002</cClassTribReg>
			<pAliqEfetRegIBSUF>1.1234</pAliqEfetRegIBSUF>
			<vTribRegIBSUF>41.00</vTribRegIBSUF>
			<pAliqEfetRegIBSMun>2.1234</pAliqEfetRegIBSMun>
			<vTribRegIBSMun>42.00</vTribRegIBSMun>
			<pAliqEfetRegCBS>3.1234</pAliqEfetRegCBS>
			<vTribRegCBS>43.00</vTribRegCBS>
		</gTribRegular>
		<gTribCompraGov>
			<pAliqIBSUF>1.9999</pAliqIBSUF><vTribIBSUF>51.00</vTribIBSUF>
			<pAliqIBSMun>2.9999</pAliqIBSMun><vTribIBSMun>52.00</vTribIBSMun>
			<pAliqCBS>3.9999</pAliqCBS><vTribCBS>53.00</vTribCBS>
		</gTribCompraGov>
	</gIBSCBS>
	<gTransfCred><vIBS>61.00</vIBS><vCBS>62.00</vCBS></gTransfCred>
	<gAjusteCompet><competApur>2026-03</competApur><vIBS>71.00</vIBS><vCBS>72.00</vCBS></gAjusteCompet>
	<gEstornoCred><vIBSEstCred>81.00</vIBSEstCred><vCBSEstCred>82.00</vCBSEstCred></gEstornoCred>
	<gCredPresOper>
		<vBCCredPres>91.00</vBCCredPres>
		<cCredPres>07</cCredPres>
		<gIBSCredPres><pCredPres>1.2345</pCredPres><vCredPres>92.00</vCredPres><vCredPresCondSus>93.00</vCredPresCondSus></gIBSCredPres>
		<gCBSCredPres><pCredPres>2.2345</pCredPres><vCredPres>94.00</vCredPres><vCredPresCondSus>95.00</vCredPresCondSus></gCBSCredPres>
	</gCredPresOper>
	<gCredPresIBSZFM>
		<competApur>2026-04</competApur>
		<tpCredPresIBSZFM>2</tpCredPresIBSZFM>
		<vCredPresIBSZFM>96.00</vCredPresIBSZFM>
	</gCredPresIBSZFM>
</IBSCBS>`

func TestLerIBSCBS_ArvoreCompleta(t *testing.T) {
	var v IBSCBS
	LerIBSCBS(parsear(t, xmlIBSCBS), &v)

	if v.CST != CST200 {
		t.Errorf("CST = %d, esperado CST200", v.CST)
	}
	if v.CClassTrib != "000001" {
		t.Errorf("cClassTrib = %q", v.CClassTrib)
	}
	if v.IndDoacao != pcn.TieSim {
		t.Errorf("indDoacao = %d", v.IndDoacao)
	}

	g := v.GIBSCBS
	if g.VBC != 1000 || g.VIBS != 85 {
		t.Errorf("vBC/vIBS = %v/%v", g.VBC, g.VIBS)
	}
	if g.GIBSUF.PIBSUF != 0.1234 || g.GIBSUF.VIBSUF != 30 {
		t.Errorf("gIBSUF = %+v", g.GIBSUF)
	}
	if g.GIBSUF.GDif.PDif != 1.2345 || g.GIBSUF.GDif.VDif != 10 {
		t.Errorf("gIBSUF/gDif = %+v", g.GIBSUF.GDif)
	}
	if g.GIBSUF.GDevTrib.PDevTrib != 2.3456 || g.GIBSUF.GDevTrib.VDevTrib != 20 {
		t.Errorf("gIBSUF/gDevTrib = %+v", g.GIBSUF.GDevTrib)
	}
	if g.GIBSUF.GRed.PRedAliq != 3.4567 || g.GIBSUF.GRed.PAliqEfet != 4.56 {
		t.Errorf("gIBSUF/gRed = %+v", g.GIBSUF.GRed)
	}
	if g.GIBSMun.PIBSMun != 0.5678 || g.GIBSMun.VIBSMun != 31 {
		t.Errorf("gIBSMun = %+v", g.GIBSMun)
	}
	if g.GCBS.PCBS != 0.9012 || g.GCBS.VCBS != 32 {
		t.Errorf("gCBS = %+v", g.GCBS)
	}
	if g.GCBS.GALCZFMCBS.TpALCZFMCBS != TpALCZFMCBSOpInd || g.GCBS.GALCZFMCBS.NProcSuframa != "PROC123" {
		t.Errorf("gALCZFMCBS = %+v", g.GCBS.GALCZFMCBS)
	}
	if g.GTribRegular.CSTReg != CST000 || g.GTribRegular.VTribRegCBS != 43 {
		t.Errorf("gTribRegular = %+v", g.GTribRegular)
	}
	if g.GTribCompraGov.PAliqCBS != 3.9999 || g.GTribCompraGov.VTribCBS != 53 {
		t.Errorf("gTribCompraGov = %+v", g.GTribCompraGov)
	}

	if v.GTransfCred.VIBS != 61 || v.GTransfCred.VCBS != 62 {
		t.Errorf("gTransfCred = %+v", v.GTransfCred)
	}
	if v.GEstornoCred.VIBSEstCred != 81 || v.GEstornoCred.VCBSEstCred != 82 {
		t.Errorf("gEstornoCred = %+v", v.GEstornoCred)
	}
	if v.GCredPresOper.CCredPres != Cp07 || v.GCredPresOper.VBCCredPres != 91 {
		t.Errorf("gCredPresOper = %+v", v.GCredPresOper)
	}
	if v.GCredPresOper.GIBSCredPres.PCredPres != 1.2345 ||
		v.GCredPresOper.GCBSCredPres.VCredPresCondSus != 95 {
		t.Errorf("credito presumido = %+v / %+v",
			v.GCredPresOper.GIBSCredPres, v.GCredPresOper.GCBSCredPres)
	}
}

func TestLerIBSCBS_CompetenciaAAAAMM(t *testing.T) {
	var v IBSCBS
	LerIBSCBS(parsear(t, xmlIBSCBS), &v)

	esperadoAjuste := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !v.GAjusteCompet.CompetApur.Equal(esperadoAjuste) {
		t.Errorf("gAjusteCompet/competApur = %v, esperado %v", v.GAjusteCompet.CompetApur, esperadoAjuste)
	}

	esperadoZFM := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	if !v.GCredPresIBSZFM.CompetApur.Equal(esperadoZFM) {
		t.Errorf("gCredPresIBSZFM/competApur = %v, esperado %v", v.GCredPresIBSZFM.CompetApur, esperadoZFM)
	}
	if v.GCredPresIBSZFM.TpCredPresIBSZFM != TcpBensCapital {
		t.Errorf("tpCredPresIBSZFM = %d, esperado TcpBensCapital", v.GCredPresIBSZFM.TpCredPresIBSZFM)
	}
}

func TestLerIBSCBS_CompetenciaVaziaDaTempoZero(t *testing.T) {
	node := parsear(t, `<IBSCBS><gAjusteCompet><competApur></competApur><vIBS>1.00</vIBS></gAjusteCompet></IBSCBS>`)
	var v IBSCBS
	LerIBSCBS(node, &v)
	if !v.GAjusteCompet.CompetApur.IsZero() {
		t.Errorf("competencia vazia deveria dar tempo zero, deu %v", v.GAjusteCompet.CompetApur)
	}
	if v.GAjusteCompet.VIBS != 1 {
		t.Errorf("os demais campos deveriam continuar sendo lidos: vIBS = %v", v.GAjusteCompet.VIBS)
	}
}

func TestLerIBSCBS_NodeNilNaoAltera(t *testing.T) {
	v := IBSCBS{CClassTrib: "preexistente"}
	LerIBSCBS(nil, &v)
	if v.CClassTrib != "preexistente" {
		t.Error("node nil deveria sair sem tocar na struct")
	}
}

// ---------------------------------------------------------------------------
// Monofasia
// ---------------------------------------------------------------------------

const xmlMono = `<gIBSCBSMono>
	<gIBSMonoAdRem>
		<gMonoPadrao><qBCMono>1.1111</qBCMono><adRemIBS>2.2222</adRemIBS><vIBSMono>3.33</vIBSMono></gMonoPadrao>
		<gMonoReten><qBCMonoReten>4.4444</qBCMonoReten><adRemIBSReten>5.5555</adRemIBSReten><vIBSMonoReten>6.66</vIBSMonoReten></gMonoReten>
		<gMonoRet><vIBSMonoRet>7.7777</vIBSMonoRet></gMonoRet>
		<gpBioDiferenca><qBCBioComb>8.8888</qBCBioComb><vIBSDiferenca>9.99</vIBSDiferenca></gpBioDiferenca>
	</gIBSMonoAdRem>
	<gIBSMonoAdValorem>
		<gMonoPadrao><vBCMono>1.0001</vBCMono><pAliqMonoUF>2.0002</pAliqMonoUF><vIBSMonoUF>3.0003</vIBSMonoUF><pAliqMonoMun>4.0004</pAliqMonoMun><vIBSMonoMun>5.0005</vIBSMonoMun><vIBSMono>6.06</vIBSMono></gMonoPadrao>
		<gMonoReten><vBCMonoReten>7.07</vBCMonoReten><pAliqMonoReten>8.0008</pAliqMonoReten><vIBSMonoReten>9.09</vIBSMonoReten></gMonoReten>
		<gMonoRet><vIBSMonoRet>1.2345</vIBSMonoRet></gMonoRet>
		<gpBioDiferenca><qBCBioComb>99.9999</qBCBioComb><vIBSDiferenca>88.88</vIBSDiferenca></gpBioDiferenca>
	</gIBSMonoAdValorem>
	<gCBSMonoAdRem>
		<gMonoPadrao><qBCMono>1.1111</qBCMono><adRemCBS>2.2222</adRemCBS><vCBSMono>3.33</vCBSMono></gMonoPadrao>
		<gMonoReten><qBCMonoReten>4.4444</qBCMonoReten><adRemCBSReten>5.5555</adRemCBSReten><vCBSMonoReten>6.66</vCBSMonoReten></gMonoReten>
		<gMonoRet><vCBSMonoRet>7.7777</vCBSMonoRet></gMonoRet>
		<gpBioDiferenca><qBCBioComb>8.8888</qBCBioComb><vCBSDiferenca>9.99</vCBSDiferenca></gpBioDiferenca>
	</gCBSMonoAdRem>
	<gCBSMonoAdValorem>
		<gMonoPadrao><vBCMono>1.0001</vBCMono><pAliqMonoCBS>2.0002</pAliqMonoCBS><vCBSMono>3.03</vCBSMono></gMonoPadrao>
		<gMonoReten><vBCMonoReten>4.04</vBCMonoReten><pAliqMonoReten>5.0005</pAliqMonoReten><vCBSMonoReten>6.06</vCBSMonoReten></gMonoReten>
		<gMonoRet><vCBSMonoRet>7.8901</vCBSMonoRet></gMonoRet>
		<gpBioDiferenca><qBCBioComb>77.7777</qBCBioComb><vCBSDiferenca>66.66</vCBSDiferenca></gpBioDiferenca>
	</gCBSMonoAdValorem>
	<vTotIBSMonoItem>100.00</vTotIBSMonoItem>
	<vTotCBSMonoItem>200.00</vTotCBSMonoItem>
</gIBSCBSMono>`

func TestLerGIBSCBSMono(t *testing.T) {
	var v GIBSCBSMono
	LerGIBSCBSMono(parsear(t, xmlMono), &v)

	if v.VTotIBSMonoItem != 100 || v.VTotCBSMonoItem != 200 {
		t.Errorf("totais = %v/%v", v.VTotIBSMonoItem, v.VTotCBSMonoItem)
	}

	adRem := v.GIBSMonoAdRem
	if adRem.GMonoPadrao.QBCMono != 1.1111 || adRem.GMonoPadrao.VIBSMono != 3.33 {
		t.Errorf("gIBSMonoAdRem/gMonoPadrao = %+v", adRem.GMonoPadrao)
	}
	if adRem.GMonoReten.AdRemIBSReten != 5.5555 {
		t.Errorf("gIBSMonoAdRem/gMonoReten = %+v", adRem.GMonoReten)
	}
	if adRem.GMonoRet.VIBSMonoRet != 7.7777 {
		t.Errorf("gIBSMonoAdRem/gMonoRet = %+v", adRem.GMonoRet)
	}
	if adRem.GpBioDiferenca.QBCBioComb != 8.8888 || adRem.GpBioDiferenca.VIBSDiferenca != 9.99 {
		t.Errorf("gIBSMonoAdRem/gpBioDiferenca = %+v", adRem.GpBioDiferenca)
	}

	adVal := v.GIBSMonoAdValorem
	if adVal.GMonoPadrao.VIBSMonoMun != 5.0005 || adVal.GMonoPadrao.VIBSMono != 6.06 {
		t.Errorf("gIBSMonoAdValorem/gMonoPadrao = %+v", adVal.GMonoPadrao)
	}

	cbsAdRem := v.GCBSMonoAdRem
	if cbsAdRem.GMonoPadrao.AdRemCBS != 2.2222 || cbsAdRem.GMonoRet.VCBSMonoRet != 7.7777 {
		t.Errorf("gCBSMonoAdRem = %+v", cbsAdRem)
	}
	if cbsAdRem.GpBioDiferenca.VCBSDiferenca != 9.99 {
		t.Errorf("gCBSMonoAdRem/gpBioDiferenca = %+v", cbsAdRem.GpBioDiferenca)
	}

	cbsAdVal := v.GCBSMonoAdValorem
	if cbsAdVal.GMonoPadrao.PAliqMonoCBS != 2.0002 || cbsAdVal.GMonoRet.VCBSMonoRet != 7.8901 {
		t.Errorf("gCBSMonoAdValorem = %+v", cbsAdVal)
	}
}

func TestMonoAdValorem_NaoLeGpBioDiferenca(t *testing.T) {
	// OMISSAO DO ACBr, replicada de proposito: Ler_gIBSMonoAdValorem e
	// Ler_gCBSMonoAdValorem nao leem gpBioDiferenca, embora o grupo exista
	// no XML e o campo exista na struct. Ler_gIBSMonoAdRem le.
	//
	// Se algum dia o ACBr corrigir isso, este teste quebra -- e e para
	// quebrar, porque a correcao precisa ser deliberada.
	var v GIBSCBSMono
	LerGIBSCBSMono(parsear(t, xmlMono), &v)

	if v.GIBSMonoAdValorem.GpBioDiferenca != (GpBioDiferencaIBS{}) {
		t.Errorf("gIBSMonoAdValorem/gpBioDiferenca deveria ficar zerado (omissao do ACBr), veio %+v",
			v.GIBSMonoAdValorem.GpBioDiferenca)
	}
	if v.GCBSMonoAdValorem.GpBioDiferenca != (GpBioDiferencaCBS{}) {
		t.Errorf("gCBSMonoAdValorem/gpBioDiferenca deveria ficar zerado (omissao do ACBr), veio %+v",
			v.GCBSMonoAdValorem.GpBioDiferenca)
	}
}

// ---------------------------------------------------------------------------
// Totais
// ---------------------------------------------------------------------------

const xmlIBSCBSTot = `<IBSCBSTot>
	<vBCIBSCBS>1000.00</vBCIBSCBS>
	<gIBS>
		<gIBSUF><vDif>1.00</vDif><vDevTrib>2.00</vDevTrib><vIBSUF>3.00</vIBSUF></gIBSUF>
		<gIBSMun><vDif>4.00</vDif><vDevTrib>5.00</vDevTrib><vIBSMun>6.00</vIBSMun></gIBSMun>
		<vIBS>7.00</vIBS><vCredPres>8.00</vCredPres><vCredPresCondSus>9.00</vCredPresCondSus>
	</gIBS>
	<gCBS>
		<vDif>10.00</vDif><vDevTrib>11.00</vDevTrib><vCBS>12.00</vCBS>
		<vCredPres>13.00</vCredPres><vCredPresCondSus>14.00</vCredPresCondSus>
	</gCBS>
	<gMono>
		<vIBSMono>15.00</vIBSMono><vCBSMono>16.00</vCBSMono>
		<vIBSMonoReten>17.00</vIBSMonoReten><vCBSMonoReten>18.00</vCBSMonoReten>
		<vIBSMonoRet>19.00</vIBSMonoRet><vCBSMonoRet>20.00</vCBSMonoRet>
	</gMono>
	<gEstornoCred><vIBSEstCred>21.00</vIBSEstCred><vCBSEstCred>22.00</vCBSEstCred></gEstornoCred>
</IBSCBSTot>`

func TestLerIBSCBSTot(t *testing.T) {
	var v IBSCBSTot
	LerIBSCBSTot(parsear(t, xmlIBSCBSTot), &v)

	if v.VBCIBSCBS != 1000 {
		t.Errorf("vBCIBSCBS = %v", v.VBCIBSCBS)
	}
	if v.GIBS.GIBSUFTot.VIBSUF != 3 || v.GIBS.GIBSMunTot.VIBSMun != 6 {
		t.Errorf("gIBS = %+v", v.GIBS)
	}
	if v.GIBS.VIBS != 7 || v.GIBS.VCredPresCondSus != 9 {
		t.Errorf("gIBS totais = %+v", v.GIBS)
	}
	if v.GCBS.VCBS != 12 || v.GCBS.VCredPresCondSus != 14 {
		t.Errorf("gCBS = %+v", v.GCBS)
	}
	if v.GMono.VIBSMono != 15 || v.GMono.VCBSMonoRet != 20 {
		t.Errorf("gMono = %+v", v.GMono)
	}
	if v.GEstornoCred.VIBSEstCred != 21 || v.GEstornoCred.VCBSEstCred != 22 {
		t.Errorf("gEstornoCred = %+v", v.GEstornoCred)
	}
}

func TestGIBSTot_SubgruposTemNomeDeTagDoItem(t *testing.T) {
	// No total, as tags continuam sendo gIBSUF e gIBSMun, mas os campos sao
	// outros: vDif/vDevTrib/vIBSUF, e nao pIBSUF/gDif/gRed. Trocar as
	// structs passa despercebido no build.
	var v GIBS
	LerGIBSTot(parsear(t, `<gIBS><gIBSUF><vIBSUF>99.00</vIBSUF></gIBSUF></gIBS>`), &v)
	if v.GIBSUFTot.VIBSUF != 99 {
		t.Errorf("gIBSUF do total = %+v", v.GIBSUFTot)
	}
}

// ---------------------------------------------------------------------------
// Pagamento vinculado
// ---------------------------------------------------------------------------

func TestLerPgtoVinc(t *testing.T) {
	node := parsear(t, `<pgtoVinc>
		<pgto nPag="1" idTransacao="TX1">
			<tpMeioPgto>03</tpMeioPgto><CNPJReceb>11222333000181</CNPJReceb><CNPJBasePSP>11222333</CNPJBasePSP>
		</pgto>
		<pgto nPag="2" idTransacao="TX2"><tpMeioPgto>04</tpMeioPgto></pgto>
	</pgtoVinc>`)

	var v PgtoVinc
	LerPgtoVinc(node, &v)

	if len(v.Pgto) != 2 {
		t.Fatalf("len(pgto) = %d, esperado 2", len(v.Pgto))
	}
	if v.Pgto[0].NPag != 1 || v.Pgto[0].IDTransacao != "TX1" {
		t.Errorf("pgto[0] atributos = %+v", v.Pgto[0])
	}
	if v.Pgto[0].TpMeioPgto != "03" || v.Pgto[0].CNPJReceb != "11222333000181" {
		t.Errorf("pgto[0] = %+v", v.Pgto[0])
	}
	if v.Pgto[1].NPag != 2 || v.Pgto[1].TpMeioPgto != "04" {
		t.Errorf("pgto[1] = %+v", v.Pgto[1])
	}
}

func TestLerPgto_SemAtributoNPagNaoEstoura(t *testing.T) {
	// DIVERGENCIA: o original faz StrToInt e levanta excecao. Aqui vale 0 --
	// um atributo ausente nao pode derrubar a importacao do lote inteiro.
	node := parsear(t, `<pgtoVinc><pgto><tpMeioPgto>03</tpMeioPgto></pgto></pgtoVinc>`)
	var v PgtoVinc
	LerPgtoVinc(node, &v)

	if len(v.Pgto) != 1 {
		t.Fatalf("len(pgto) = %d", len(v.Pgto))
	}
	if v.Pgto[0].NPag != 0 {
		t.Errorf("nPag = %d, esperado 0", v.Pgto[0].NPag)
	}
	if v.Pgto[0].TpMeioPgto != "03" {
		t.Errorf("os demais campos deveriam ser lidos: %+v", v.Pgto[0])
	}
}

// ---------------------------------------------------------------------------
// Imposto Seletivo e documento referenciado
// ---------------------------------------------------------------------------

func TestLerISel(t *testing.T) {
	node := parsear(t, `<IS>
		<CSTIS>001</CSTIS><cClassTribIS>000003</cClassTribIS>
		<vBCIS>100.00</vBCIS><pIS>1.23</pIS><adRemIS>2.3456</adRemIS>
		<uTrib>UN</uTrib><qTrib>3.4567</qTrib><vIS>4.56</vIS>
	</IS>`)

	var v GIS
	LerISel(node, &v)

	if v.CSTIS != "001" || v.CClassTribIS != "000003" {
		t.Errorf("CSTIS/cClassTribIS = %q/%q", v.CSTIS, v.CClassTribIS)
	}
	if v.VBCIS != 100 || v.PIS != 1.23 || v.AdRemIS != 2.3456 {
		t.Errorf("valores = %+v", v)
	}
	if v.UTrib != "UN" || v.QTrib != 3.4567 || v.VIS != 4.56 {
		t.Errorf("tributacao = %+v", v)
	}
}

func TestLerISTot(t *testing.T) {
	var v ISTot
	LerISTot(parsear(t, `<ISTot><vIS>77.00</vIS></ISTot>`), &v)
	if v.VIS != 77 {
		t.Errorf("vIS = %v", v.VIS)
	}
}

func TestLerDFeReferenciado(t *testing.T) {
	var v DFeReferenciado
	LerDFeReferenciado(parsear(t, `<ref><chaveAcesso>CHAVE</chaveAcesso><nItem>5</nItem></ref>`), &v)
	if v.ChaveAcesso != "CHAVE" || v.NItem != 5 {
		t.Errorf("DFeReferenciado = %+v", v)
	}
}

// ---------------------------------------------------------------------------
// Precisao por campo
// ---------------------------------------------------------------------------

func TestPrecisao_MesmaTagPrecisoesDiferentes(t *testing.T) {
	// vIBSMonoRet e De4 no item e De2 no total. A diferenca so aparece com
	// FloatIsIntString ligado, mas o teste documenta a assimetria e trava a
	// escolha de casas feita em cada leitor.
	nodeItem := parsear(t, `<gMonoRet><vIBSMonoRet>1234</vIBSMonoRet></gMonoRet>`)
	nodeItem.FindAnyNs("vIBSMonoRet").FloatIsIntString = true
	var item GMonoRetIBS
	LerGMonoRetIBS(nodeItem, &item)
	if item.VIBSMonoRet != 0.1234 {
		t.Errorf("item: vIBSMonoRet = %v, esperado 0.1234 (De4)", item.VIBSMonoRet)
	}

	nodeTot := parsear(t, `<gMono><vIBSMonoRet>1234</vIBSMonoRet></gMono>`)
	nodeTot.FindAnyNs("vIBSMonoRet").FloatIsIntString = true
	var tot GMono
	LerGMonoTot(nodeTot, &tot)
	if tot.VIBSMonoRet != 12.34 {
		t.Errorf("total: vIBSMonoRet = %v, esperado 12.34 (De2)", tot.VIBSMonoRet)
	}
}

// ---------------------------------------------------------------------------
// Robustez
// ---------------------------------------------------------------------------

func TestLeitores_ToleramNodeNilEPonteiroNil(t *testing.T) {
	// Nenhum leitor pode estourar com entrada ausente -- e o caminho normal
	// num documento que nao traz o grupo.
	LerGCompraGovReduzido(nil, nil)
	LerGCompraGov(nil, nil)
	LerGPagAntecipadoProd(nil, nil)
	LerGPagAntecipadoIde(nil, nil)
	LerGPagAntecipado(nil, nil)
	LerIBSCBS(nil, nil)
	LerGIBSCBS(nil, nil)
	LerGIBSUF(nil, nil)
	LerGIBSMun(nil, nil)
	LerGCBS(nil, nil)
	LerGDif(nil, nil)
	LerGDevTrib(nil, nil)
	LerGRed(nil, nil)
	LerGTribRegular(nil, nil)
	LerGTribCompraGov(nil, nil)
	LerGEstornoCred(nil, nil)
	LerGALCZFMCBS(nil, nil)
	LerIBSCBSTot(nil, nil)
	LerGIBSTot(nil, nil)
	LerGIBSUFTot(nil, nil)
	LerGIBSMunTot(nil, nil)
	LerGCBSTot(nil, nil)
	LerGEstornoCredTot(nil, nil)
	LerGMonoTot(nil, nil)
	LerPgtoVinc(nil, nil)
	LerPgto(nil, nil)
	LerISel(nil, nil)
	LerISTot(nil, nil)
	LerDFeReferenciado(nil, nil)
	LerGIBSCBSMono(nil, nil)
	LerGIBSMonoAdRem(nil, nil)
	LerGMonoPadraoIBSQtde(nil, nil)
	LerGMonoRetenIBSQtde(nil, nil)
	LerGMonoRetIBS(nil, nil)
	LerGpBioDiferencaIBS(nil, nil)
	LerGIBSMonoAdValorem(nil, nil)
	LerGMonoPadraoIBSAliq(nil, nil)
	LerGMonoRetenIBSAliq(nil, nil)
	LerGCBSMonoAdRem(nil, nil)
	LerGMonoPadraoCBSQtde(nil, nil)
	LerGMonoRetenCBSQtde(nil, nil)
	LerGMonoRetCBS(nil, nil)
	LerGpBioDiferencaCBS(nil, nil)
	LerGCBSMonoAdValorem(nil, nil)
	LerGMonoPadraoCBSAliq(nil, nil)
	LerGMonoRetenCBSAliq(nil, nil)
	LerGTransfCred(nil, nil)
	LerGCredPresIBSZFM(nil, nil)
	LerGAjusteCompet(nil, nil)
	LerGCredPresOper(nil, nil)
	LerGIBSCredPres(nil, nil)
	LerGCBSCredPres(nil, nil)
}

func TestValorZero_JaEUtilizavel(t *testing.T) {
	// Structs por valor: o zero value nao tem ponteiro nil para desreferenciar.
	var v IBSCBS
	if v.GIBSCBS.GIBSUF.GDif.PDif != 0 {
		t.Error("a arvore deveria ser navegavel sem inicializacao")
	}
	var tot IBSCBSTot
	if tot.GIBS.GIBSUFTot.VIBSUF != 0 {
		t.Error("a arvore de totais deveria ser navegavel sem inicializacao")
	}
}

// ---------------------------------------------------------------------------
// INI -- round-trip do IBSCBS completo
// ---------------------------------------------------------------------------

func TestINI_IBSCBSRoundTripCompleto(t *testing.T) {
	// Exercita TODOS os grupos do IBSCBS que o .ini cobre, inclusive os de
	// credito/ajuste (gTransfCred, gAjusteCompet, gCredPresOper,
	// gCredPresIBSZFM) -- lacuna apontada na revisao de fidelidade.
	original := IBSCBS{
		CST:        CST200,
		CClassTrib: "000001",
		IndDoacao:  pcn.TieSim,
		GIBSCBS: GIBSCBS{
			VBC:  1000,
			VIBS: 85,
			GIBSUF: GIBSUFValores{
				PIBSUF:   0.1234,
				VIBSUF:   7.82,
				GDif:     GDif{PDif: 1.2345, VDif: 10},
				GDevTrib: GDevTrib{PDevTrib: 2.3456, VDevTrib: 20},
				GRed:     GRed{PRedAliq: 3.4567, PAliqEfet: 4.56},
			},
			GIBSMun: GIBSMunValores{PIBSMun: 0.5678, VIBSMun: 15.64},
			GCBS: GCBSValores{
				PCBS: 0.9, VCBS: 70.39,
				GALCZFMCBS: GALCZFMCBS{
					TpALCZFMCBS: TpALCZFMCBSOpInd, NProcSuframa: "SUF-1",
					PAliqEfetRegCBS: 1.5, VTribRegCBS: 11.73,
				},
			},
			GTribRegular:   GTribRegular{CSTReg: CST000, CClassTribReg: "000002", VTribRegCBS: 43},
			GTribCompraGov: GTribCompraGov{PAliqCBS: 3.9999, VTribCBS: 53},
		},
		GEstornoCred:  GEstornoCred{VIBSEstCred: 5, VCBSEstCred: 6},
		GTransfCred:   GTransfCred{VIBS: 61, VCBS: 62},
		GAjusteCompet: GAjusteCompet{CompetApur: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), VIBS: 71, VCBS: 72},
		GCredPresOper: GCredPresOper{
			VBCCredPres:  91,
			CCredPres:    Cp07,
			GIBSCredPres: GIBSCBSCredPres{PCredPres: 1.2345, VCredPres: 92, VCredPresCondSus: 93},
			GCBSCredPres: GIBSCBSCredPres{PCredPres: 2.2345, VCredPres: 94, VCredPresCondSus: 95},
		},
		GCredPresIBSZFM: CredPresIBSZFM{
			CompetApur:       time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			TpCredPresIBSZFM: TcpBensCapital,
			VCredPresIBSZFM:  96,
		},
	}

	ini := pcn.NovoINI()
	GravarINIIBSCBS(ini, original, 1, SemIndice)

	var relido IBSCBS
	LerINIIBSCBS(ini, &relido, 1, SemIndice)

	if relido.CST != original.CST || relido.CClassTrib != original.CClassTrib {
		t.Errorf("cabecalho: %+v", relido)
	}
	if relido.GIBSCBS.GIBSUF != original.GIBSCBS.GIBSUF {
		t.Errorf("gIBSUF: %+v vs %+v", relido.GIBSCBS.GIBSUF, original.GIBSCBS.GIBSUF)
	}
	if relido.GIBSCBS.GCBS.GALCZFMCBS != original.GIBSCBS.GCBS.GALCZFMCBS {
		t.Errorf("gALCZFMCBS: %+v", relido.GIBSCBS.GCBS.GALCZFMCBS)
	}
	if relido.GEstornoCred != original.GEstornoCred {
		t.Errorf("gEstornoCred: %+v", relido.GEstornoCred)
	}
	if relido.GTransfCred != original.GTransfCred {
		t.Errorf("gTransfCred: %+v", relido.GTransfCred)
	}
	if relido.GAjusteCompet.VIBS != original.GAjusteCompet.VIBS ||
		!relido.GAjusteCompet.CompetApur.Equal(original.GAjusteCompet.CompetApur) {
		t.Errorf("gAjusteCompet: %+v vs %+v", relido.GAjusteCompet, original.GAjusteCompet)
	}
	if relido.GCredPresOper.CCredPres != original.GCredPresOper.CCredPres ||
		relido.GCredPresOper.GIBSCredPres != original.GCredPresOper.GIBSCredPres ||
		relido.GCredPresOper.GCBSCredPres != original.GCredPresOper.GCBSCredPres {
		t.Errorf("gCredPresOper: %+v", relido.GCredPresOper)
	}
	if relido.GCredPresIBSZFM.TpCredPresIBSZFM != original.GCredPresIBSZFM.TpCredPresIBSZFM ||
		relido.GCredPresIBSZFM.VCredPresIBSZFM != original.GCredPresIBSZFM.VCredPresIBSZFM ||
		!relido.GCredPresIBSZFM.CompetApur.Equal(original.GCredPresIBSZFM.CompetApur) {
		t.Errorf("gCredPresIBSZFM: %+v", relido.GCredPresIBSZFM)
	}
}

func TestINI_GPagAntecipadoProdRoundTrip(t *testing.T) {
	// DIVERGENCIA testada: o ACBr le chDFePagAnt de variavel nunca
	// inicializada (sempre vazio); aqui a chave homonima e lida de volta.
	original := GPagAntecipadoProd{ChDFePagAnt: "CHAVE-ANTECIPADO", NItemPagAnt: 7}

	ini := pcn.NovoINI()
	GravarINIGPagAntecipadoProd(ini, original, 1, SemIndice)

	var relido GPagAntecipadoProd
	LerINIGPagAntecipadoProd(ini, &relido, 1, SemIndice)

	if relido != original {
		t.Errorf("round-trip = %+v, esperado %+v", relido, original)
	}
}
