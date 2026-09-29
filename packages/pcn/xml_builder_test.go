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

package pcn

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Elem -- construtor de XML de geracao
// ---------------------------------------------------------------------------

func TestElemXMLCompacto(t *testing.T) {
	e := NovoElem("ide").
		Filho(NovoElem("cUF").Texto("35")).
		Filho(NovoElem("serie").Texto("1"))
	got := e.XML()
	want := "<ide><cUF>35</cUF><serie>1</serie></ide>"
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
}

func TestElemAtributosNaOrdemDeInsercao(t *testing.T) {
	// A ordem Id -> versao replica o Gerar_InfNFGas do ACBr (SetAttribute
	// em sequencia) e coincide com a ordem canonica do C14N.
	e := NovoElem("infNFGas").Attr("Id", "NFGas123").Attr("versao", "1.00")
	got := e.XML()
	want := `<infNFGas Id="NFGas123" versao="1.00"/>`
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
}

func TestElemFilhoNilIgnorado(t *testing.T) {
	// AppendChild(nil) do ACBr quando o grupo opcional esta vazio.
	e := NovoElem("total").Filho(nil).Filhos(nil, NovoElem("vNF").Texto("10.00"), nil)
	got := e.XML()
	want := "<total><vNF>10.00</vNF></total>"
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
	if !e.TemFilhos() {
		t.Fatal("TemFilhos() deveria ser true")
	}
}

func TestElemEscapeDeTexto(t *testing.T) {
	got := NovoElem("xNome").Texto(`P&G <Gas> "SA"`).XML()
	want := `<xNome>P&amp;G &lt;Gas&gt; "SA"</xNome>`
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
}

func TestElemTextoBrutoNaoEscapa(t *testing.T) {
	// qrCodNFGas carrega CDATA literal.
	got := NovoElem("qrCodNFGas").TextoBruto("<![CDATA[https://x?a=1&b=2]]>").XML()
	want := "<qrCodNFGas><![CDATA[https://x?a=1&b=2]]></qrCodNFGas>"
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
}

func TestElemEscapeDeAtributo(t *testing.T) {
	got := NovoElem("a").Attr("v", `x"y'<&`).XML()
	want := `<a v="x&quot;y&#39;&lt;&amp;"/>`
	if got != want {
		t.Fatalf("XML() = %q, esperado %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Formatadores de geracao
// ---------------------------------------------------------------------------

func TestFormatarDecimalXML(t *testing.T) {
	// AddNode tcDeN: casas FIXAS, ponto decimal, sem agrupamento.
	casos := []struct {
		v     float64
		casas int
		want  string
	}{
		{1234.5, 2, "1234.50"},
		{0, 2, "0.00"},
		{0.1234567891, 10, "0.1234567891"},
		{10, 4, "10.0000"},
		{1.005, 2, "1.00"}, // FormatFloat arredonda half-even do float
	}
	for _, c := range casos {
		if got := FormatarDecimalXML(c.v, c.casas); got != c.want {
			t.Errorf("FormatarDecimalXML(%v, %d) = %q, esperado %q", c.v, c.casas, got, c.want)
		}
	}
}

func TestFormatarVersaoXML(t *testing.T) {
	if got := FormatarVersaoXML(1.0); got != "1.00" {
		t.Fatalf("FormatarVersaoXML(1.0) = %q, esperado \"1.00\"", got)
	}
}

func TestFormatarDataXML(t *testing.T) {
	d := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	if got := FormatarDataXML(d); got != "2026-03-07" {
		t.Fatalf("FormatarDataXML = %q", got)
	}
	if got := FormatarDataXML(time.Time{}); got != "" {
		t.Fatalf("FormatarDataXML(zero) = %q, esperado vazio", got)
	}
}

func TestFormatarDataHoraXMLComFusoDaUF(t *testing.T) {
	// DateTimeTodh + GetUTC(UF): sem fuso no time, usa o oficial da UF.
	d := time.Date(2026, 3, 7, 14, 30, 5, 0, time.UTC)
	casos := []struct {
		uf   string
		want string
	}{
		{"SP", "2026-03-07T14:30:05-03:00"},
		{"AC", "2026-03-07T14:30:05-05:00"},
		{"AM", "2026-03-07T14:30:05-04:00"},
		{"MT", "2026-03-07T14:30:05-04:00"},
		{"", "2026-03-07T14:30:05-03:00"},
	}
	for _, c := range casos {
		if got := FormatarDataHoraXML(d, c.uf); got != c.want {
			t.Errorf("FormatarDataHoraXML(UF=%q) = %q, esperado %q", c.uf, got, c.want)
		}
	}
}

func TestFormatarDataHoraXMLComFusoProprio(t *testing.T) {
	// time.Time ja localizado: o offset do proprio valor prevalece.
	loc := time.FixedZone("-04", -4*3600)
	d := time.Date(2026, 3, 7, 14, 30, 5, 0, loc)
	if got := FormatarDataHoraXML(d, "SP"); got != "2026-03-07T14:30:05-04:00" {
		t.Fatalf("FormatarDataHoraXML = %q", got)
	}
}

func TestFiltrarTextoXML(t *testing.T) {
	// FiltrarTextoXML default do TACBrXmlWriter: TiraAcentos + colapso de
	// espacos + quebras de linha viram ';'.
	got := FiltrarTextoXML("Fornecimento  de   gás\r\ncanalizado\nSão Paulo")
	want := "Fornecimento de gas;canalizado;Sao Paulo"
	if got != want {
		t.Fatalf("FiltrarTextoXML = %q, esperado %q", got, want)
	}
}

func TestTiraAcentos(t *testing.T) {
	got := TiraAcentos("ÁGUA çedilha ãõ ÊÎÔÛ nº 1°")
	want := "AGUA cedilha ao EIOU no 1o"
	if got != want {
		t.Fatalf("TiraAcentos = %q, esperado %q", got, want)
	}
}

func TestPadLeftZeros(t *testing.T) {
	if got := PadLeftZeros("7", 3); got != "007" {
		t.Fatalf("PadLeftZeros = %q", got)
	}
	if got := PadLeftZeros("1234", 3); got != "1234" {
		t.Fatalf("PadLeftZeros nao deve truncar: %q", got)
	}
}

func TestOffsetUFCobreTodasAsFaixas(t *testing.T) {
	if OffsetUF("ac") != "-05:00" {
		t.Fatal("AC deveria ser -05:00 (caso-insensitivo)")
	}
	for _, uf := range []string{"AM", "RR", "RO", "MT", "MS"} {
		if OffsetUF(uf) != "-04:00" {
			t.Fatalf("%s deveria ser -04:00", uf)
		}
	}
	for _, uf := range []string{"SP", "RS", "DF", "MA", "PA", ""} {
		if OffsetUF(uf) != "-03:00" {
			t.Fatalf("%s deveria ser -03:00", uf)
		}
	}
}

func TestElemXMLRoundTripComParser(t *testing.T) {
	// O que o builder gera o mini-DOM le de volta.
	x := NovoElem("NFGas").Attr("xmlns", "http://www.portalfiscal.inf.br/nfgas").
		Filho(NovoElem("infNFGas").Attr("Id", "NFGas1").Attr("versao", "1.00").
			Filho(NovoElem("ide").Filho(NovoElem("cUF").Texto("35")))).XML()
	doc, err := ParseString(x)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	cuf := doc.Root.FindAnyNs("infNFGas").FindAnyNs("ide").FindAnyNs("cUF")
	if ConteudoStr(cuf) != "35" {
		t.Fatalf("round-trip perdeu cUF: %q", ConteudoStr(cuf))
	}
	if !strings.Contains(x, `Id="NFGas1" versao="1.00"`) {
		t.Fatalf("ordem de atributos inesperada: %s", x)
	}
}
