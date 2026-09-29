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
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// ---------------------------------------------------------------------------
// Writer -- gCompraGov e propagacao de estado
// ---------------------------------------------------------------------------

func TestWriterGCompraGovReduzido(t *testing.T) {
	w := NovoWriter(ModeloNFGas)
	e := w.GerarGCompraGovReduzido(GCompraGovReduzido{
		TpEnteGov: TcgUniao,
		PRedutor:  10.5,
		TpOperGov: TogFornecimento,
		RefDFe:    []RefDFeAnt{{RefDFeAnt: strings.Repeat("1", 44)}},
	})
	got := e.XML()
	want := "<gCompraGov><tpEnteGov>1</tpEnteGov><pRedutor>10.5000</pRedutor>" +
		"<tpOperGov>1</tpOperGov><refDFeAnt>" + strings.Repeat("1", 44) + "</refDFeAnt></gCompraGov>"
	if got != want {
		t.Fatalf("gCompraGov = %q, esperado %q", got, want)
	}
}

func TestWriterGCompraGovSemRedutorNaoGeraMasCapturaEstado(t *testing.T) {
	// Gerar_gCompraGovReduzido seta FpRedutor/FtpEnteGov ANTES do teste --
	// o estado sobrevive mesmo sem gerar o grupo.
	w := NovoWriter(ModeloNFGas)
	if e := w.GerarGCompraGovReduzido(GCompraGovReduzido{TpEnteGov: TcgEstados, PRedutor: 0}); e != nil {
		t.Fatalf("pRedutor=0 nao deveria gerar gCompraGov: %s", e.XML())
	}
	if w.tpEnteGov != TcgEstados {
		t.Fatal("tpEnteGov deveria ter sido capturado mesmo sem gerar o grupo")
	}
}

func TestWriterPRedutorForcaGRedNosItens(t *testing.T) {
	// Com FpRedutor > 0, gRed e gerado no item MESMO com pRedAliq e
	// pAliqEfet zerados (condicao do Gerar_gIBSUF).
	w := NovoWriter(ModeloNFGas)
	w.GerarGCompraGovReduzido(GCompraGovReduzido{TpEnteGov: TcgUniao, PRedutor: 20})

	e := w.gerarGIBSUF(GIBSUFValores{PIBSUF: 0.1, VIBSUF: 1})
	if !strings.Contains(e.XML(), "<gRed><pRedAliq>0.0000</pRedAliq><pAliqEfet>0.00</pAliqEfet></gRed>") {
		t.Fatalf("gRed deveria ser forcado por pRedutor>0: %s", e.XML())
	}

	// sem o redutor, gRed zerado nao aparece
	w2 := NovoWriter(ModeloNFGas)
	if strings.Contains(w2.gerarGIBSUF(GIBSUFValores{PIBSUF: 0.1, VIBSUF: 1}).XML(), "gRed") {
		t.Fatal("gRed nao deveria aparecer sem redutor e sem valores")
	}
}

// ---------------------------------------------------------------------------
// Writer -- IBSCBS por modelo
// ---------------------------------------------------------------------------

func ibscbsCST000() IBSCBS {
	return IBSCBS{
		CST:        CST000,
		CClassTrib: "000001",
		GIBSCBS: GIBSCBS{
			VBC:     100,
			VIBS:    17.7,
			GIBSUF:  GIBSUFValores{PIBSUF: 17.7, VIBSUF: 17.7},
			GIBSMun: GIBSMunValores{PIBSMun: 0, VIBSMun: 0},
			GCBS:    GCBSValores{PCBS: 8.8, VCBS: 8.8},
		},
	}
}

func TestWriterIBSCBSNFGasCST000(t *testing.T) {
	w := NovoWriter(ModeloNFGas)
	got := w.GerarIBSCBS(ibscbsCST000()).XML()
	want := "<IBSCBS><CST>000</CST><cClassTrib>000001</cClassTrib>" +
		"<gIBSCBS><vBC>100.00</vBC>" +
		"<gIBSUF><pIBSUF>17.7000</pIBSUF><vIBSUF>17.70</vIBSUF></gIBSUF>" +
		"<gIBSMun><pIBSMun>0.0000</pIBSMun><vIBSMun>0.00</vIBSMun></gIBSMun>" +
		"<vIBS>17.70</vIBS>" +
		"<gCBS><pCBS>8.8000</pCBS><vCBS>8.80</vCBS></gCBS>" +
		"</gIBSCBS></IBSCBS>"
	if got != want {
		t.Fatalf("IBSCBS NFGas =\n%s\nesperado\n%s", got, want)
	}
	if !w.gerarIBSCBSTot {
		t.Fatal("gerarIBSCBSTot deveria ter sido ligado")
	}
}

func TestWriterIBSCBSNFGasCST200SoGeraCabecalho(t *testing.T) {
	// Gerar_IBSCBSNFGas: so cst000 gera gIBSCBS -- CST 200 fica so com
	// CST + cClassTrib.
	ib := ibscbsCST000()
	ib.CST = CST200
	got := NovoWriter(ModeloNFGas).GerarIBSCBS(ib).XML()
	want := "<IBSCBS><CST>200</CST><cClassTrib>000001</cClassTrib></IBSCBS>"
	if got != want {
		t.Fatalf("IBSCBS CST200 = %q, esperado %q", got, want)
	}
}

func TestWriterIBSCBSSemCSTNaoGera(t *testing.T) {
	w := NovoWriter(ModeloNFGas)
	if e := w.GerarIBSCBS(IBSCBS{}); e != nil {
		t.Fatalf("IBSCBS vazio nao deveria gerar: %s", e.XML())
	}
	if w.gerarIBSCBSTot {
		t.Fatal("gerarIBSCBSTot nao deveria ligar sem grupo IBSCBS")
	}
}

func TestWriterIBSCBSIndDoacao(t *testing.T) {
	ib := ibscbsCST000()
	ib.IndDoacao = pcn.TieSim
	got := NovoWriter(ModeloNFGas).GerarIBSCBS(ib).XML()
	if !strings.Contains(got, "<indDoacao>1</indDoacao>") {
		t.Fatalf("indDoacao=1 deveria aparecer: %s", got)
	}
	// TieNao NAO gera a tag (o original so escreve quando tieSim)
	ib.IndDoacao = pcn.TieNao
	got = NovoWriter(ModeloNFGas).GerarIBSCBS(ib).XML()
	if strings.Contains(got, "indDoacao") {
		t.Fatalf("indDoacao nao deveria aparecer com tieNao: %s", got)
	}
}

func TestWriterGTribCompraGovExigeEnteGov(t *testing.T) {
	// Condicao do Gerar_gIBSCBS: pAliqIBSUF > 0 E FtpEnteGov <> tcgNenhum.
	ib := ibscbsCST000()
	ib.GIBSCBS.GTribCompraGov = GTribCompraGov{PAliqIBSUF: 17.7, VTribIBSUF: 17.7}

	// sem compra governamental no documento: nao gera
	got := NovoWriter(ModeloNFGas).GerarIBSCBS(ib).XML()
	if strings.Contains(got, "gTribCompraGov") {
		t.Fatalf("gTribCompraGov exige tpEnteGov definido: %s", got)
	}

	// com o estado capturado do gCompraGov: gera
	w := NovoWriter(ModeloNFGas)
	w.GerarGCompraGovReduzido(GCompraGovReduzido{TpEnteGov: TcgUniao, PRedutor: 10})
	got = w.GerarIBSCBS(ib).XML()
	if !strings.Contains(got, "<gTribCompraGov><pAliqIBSUF>17.7000</pAliqIBSUF>") {
		t.Fatalf("gTribCompraGov deveria ser gerado: %s", got)
	}
}

// ---------------------------------------------------------------------------
// Writer -- IBSCBSTot
// ---------------------------------------------------------------------------

func TestWriterIBSCBSTotSoAposItemComIBSCBS(t *testing.T) {
	tot := IBSCBSTot{
		VBCIBSCBS: 100,
		GIBS:      GIBS{VIBS: 17.7, GIBSUFTot: GIBSUFTot{VIBSUF: 17.7}},
		GCBS:      GCBS{VCBS: 8.8},
	}

	w := NovoWriter(ModeloNFGas)
	if e := w.GerarIBSCBSTot(tot); e != nil {
		t.Fatalf("IBSCBSTot sem item com IBSCBS nao deveria gerar: %s", e.XML())
	}

	w.GerarIBSCBS(ibscbsCST000())
	got := w.GerarIBSCBSTot(tot).XML()
	want := "<IBSCBSTot><vBCIBSCBS>100.00</vBCIBSCBS>" +
		"<gIBS><gIBSUF><vDif>0.00</vDif><vDevTrib>0.00</vDevTrib><vIBSUF>17.70</vIBSUF></gIBSUF>" +
		"<gIBSMun><vDif>0.00</vDif><vDevTrib>0.00</vDevTrib><vIBSMun>0.00</vIBSMun></gIBSMun>" +
		"<vIBS>17.70</vIBS><vCredPres>0.00</vCredPres><vCredPresCondSus>0.00</vCredPresCondSus></gIBS>" +
		"<gCBS><vDif>0.00</vDif><vDevTrib>0.00</vDevTrib><vCBS>8.80</vCBS>" +
		"<vCredPres>0.00</vCredPres><vCredPresCondSus>0.00</vCredPresCondSus></gCBS>" +
		"</IBSCBSTot>"
	if got != want {
		t.Fatalf("IBSCBSTot =\n%s\nesperado\n%s", got, want)
	}
}

// ---------------------------------------------------------------------------
// Writer -- pgtoVinc
// ---------------------------------------------------------------------------

func TestWriterPgtoVinc(t *testing.T) {
	w := NovoWriter(ModeloNFGas)
	if e := w.GerarPgtoVinc(PgtoVinc{}); e != nil {
		t.Fatal("pgtoVinc vazio nao deveria gerar")
	}
	got := w.GerarPgtoVinc(PgtoVinc{Pgto: []Pgto{{
		NPag: 1, IDTransacao: "ABC", TpMeioPgto: "01",
		CNPJReceb: "11222333000181", CNPJBasePSP: "11222333",
	}}}).XML()
	want := `<pgtoVinc><pgto nPag="1" idTransacao="ABC">` +
		"<tpMeioPgto>01</tpMeioPgto><CNPJReceb>11222333000181</CNPJReceb>" +
		"<CNPJBasePSP>11222333</CNPJBasePSP></pgto></pgtoVinc>"
	if got != want {
		t.Fatalf("pgtoVinc = %q, esperado %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Round-trip: o que o writer gera, o reader do rtc le de volta
// ---------------------------------------------------------------------------

func TestWriterReaderRoundTripIBSCBS(t *testing.T) {
	original := ibscbsCST000()
	original.GIBSCBS.GIBSUF.GDif = GDif{PDif: 50, VDif: 8.85}
	original.GIBSCBS.GCBS.GDevTrib = GDevTrib{VDevTrib: 1.23}
	original.GIBSCBS.GTribRegular = GTribRegular{
		CSTReg: CST200, CClassTribReg: "200001",
		PAliqEfetRegIBSUF: 1.5, VTribRegIBSUF: 1.5,
	}

	xmlStr := NovoWriter(ModeloNFGas).GerarIBSCBS(original).XML()
	doc, err := pcn.ParseString(xmlStr)
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}

	var lido IBSCBS
	LerIBSCBS(doc.Root, &lido)

	if lido.CST != original.CST || lido.CClassTrib != original.CClassTrib {
		t.Fatalf("cabecalho divergente: %+v", lido)
	}
	if lido.GIBSCBS.VBC != 100 || lido.GIBSCBS.VIBS != 17.7 {
		t.Fatalf("gIBSCBS divergente: %+v", lido.GIBSCBS)
	}
	if lido.GIBSCBS.GIBSUF.GDif != original.GIBSCBS.GIBSUF.GDif {
		t.Fatalf("gDif divergente: %+v", lido.GIBSCBS.GIBSUF.GDif)
	}
	if lido.GIBSCBS.GCBS.GDevTrib != original.GIBSCBS.GCBS.GDevTrib {
		t.Fatalf("gDevTrib divergente: %+v", lido.GIBSCBS.GCBS.GDevTrib)
	}
	if lido.GIBSCBS.GTribRegular != original.GIBSCBS.GTribRegular {
		t.Fatalf("gTribRegular divergente: %+v", lido.GIBSCBS.GTribRegular)
	}
}
