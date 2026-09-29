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
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Mini-DOM: parse e navegacao
// ---------------------------------------------------------------------------

const xmlExemplo = `<?xml version="1.0" encoding="UTF-8"?>
<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas">
  <infNFGas versao="1.00" Id="NFGas35260311222333000181760010000000011100000012">
    <ide>
      <cUF>35</cUF>
      <mod>76</mod>
      <dhEmi>2026-03-15T10:30:00-03:00</dhEmi>
    </ide>
    <det nItem="1" chNFGasAnt="123" nItemAnt="7">
      <prod><vProd>1234.56</vProd></prod>
    </det>
    <det nItem="2">
      <prod><vProd>10.00</vProd></prod>
    </det>
  </infNFGas>
  <Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
    <SignedInfo>
      <Reference URI="#NFGas352603112223330001817600100000000111000000">
        <DigestValue>ABC123</DigestValue>
      </Reference>
    </SignedInfo>
    <SignatureValue>SIG==</SignatureValue>
    <KeyInfo><X509Data><X509Certificate>CERT==</X509Certificate></X509Data></KeyInfo>
  </Signature>
</NFGas>`

func parsear(t *testing.T, xml string) *Document {
	t.Helper()
	doc, err := ParseString(xml)
	if err != nil {
		t.Fatalf("ParseString: %v", err)
	}
	return doc
}

func TestParse_RaizEHierarquia(t *testing.T) {
	doc := parsear(t, xmlExemplo)

	if doc.Root.Nome != "NFGas" {
		t.Errorf("raiz = %q, esperado NFGas", doc.Root.Nome)
	}
	if doc.Root.Space != "http://www.portalfiscal.inf.br/nfgas" {
		t.Errorf("namespace da raiz = %q", doc.Root.Space)
	}

	inf := doc.Root.FindAnyNs("infNFGas")
	if inf == nil {
		t.Fatal("infNFGas nao encontrado")
	}
	if got := inf.Attr("versao"); got != "1.00" {
		t.Errorf("versao = %q, esperado 1.00", got)
	}
	if got := ConteudoInt(inf.FindAnyNs("ide").FindAnyNs("cUF")); got != 35 {
		t.Errorf("cUF = %d, esperado 35", got)
	}
}

func TestParse_DeclaracaoXMLNaoViraElemento(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	if doc.Root.Nome != "NFGas" {
		t.Fatalf("o prologo <?xml?> virou elemento: raiz = %q", doc.Root.Nome)
	}
}

func TestFind_ComparaNomeQualificado(t *testing.T) {
	// TACBrXmlNodeList.Find compara Node.Name -- o nome COM prefixo. Find
	// acha <Signature> sem prefixo; FindAnyNs acha nos dois casos.
	doc := parsear(t, xmlExemplo)

	if doc.Root.Find("Signature") == nil {
		t.Error("Find(Signature) deveria achar o elemento sem prefixo")
	}

	comPrefixo := `<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas"` +
		` xmlns:ds="http://www.w3.org/2000/09/xmldsig#">` +
		`<ds:Signature><ds:SignatureValue>X</ds:SignatureValue></ds:Signature></NFGas>`
	doc2 := parsear(t, comPrefixo)

	if doc2.Root.Find("Signature") != nil {
		t.Error("Find(Signature) nao deveria achar <ds:Signature> -- o ACBr tambem nao acha")
	}
	if doc2.Root.FindAnyNs("Signature") == nil {
		t.Error("FindAnyNs(Signature) deveria achar <ds:Signature>")
	}
	if got := doc2.Root.FindAnyNs("Signature").NomeQualificado; got != "ds:Signature" {
		t.Errorf("NomeQualificado = %q, esperado ds:Signature", got)
	}
}

func TestFindAllAnyNs_ELeituraDeAtributoDeItem(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	dets := doc.Root.FindAnyNs("infNFGas").FindAllAnyNs("det")
	if len(dets) != 2 {
		t.Fatalf("len(det) = %d, esperado 2", len(dets))
	}
	if got := AtributoInt(dets[0], "nItem"); got != 1 {
		t.Errorf("det[0]/@nItem = %d, esperado 1", got)
	}
	if got := dets[0].Attr("chNFGasAnt"); got != "123" {
		t.Errorf("det[0]/@chNFGasAnt = %q", got)
	}
	// Atributo ausente devolve vazio/zero, nao erro.
	if got := dets[1].Attr("chNFGasAnt"); got != "" {
		t.Errorf("det[1]/@chNFGasAnt = %q, esperado vazio", got)
	}
	if got := AtributoInt(dets[1], "nItemAnt"); got != 0 {
		t.Errorf("det[1]/@nItemAnt = %d, esperado 0", got)
	}
}

func TestNavegacao_SeguraEmNodeNil(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	// Caminho inteiro inexistente: no Delphi isso e access violation.
	n := doc.Root.FindAnyNs("naoExiste").FindAnyNs("tambemNao").FindAnyNs("nem")
	if n != nil {
		t.Fatal("cadeia inexistente deveria devolver nil")
	}
	if got := ConteudoStr(n); got != "" {
		t.Errorf("ConteudoStr(nil) = %q", got)
	}
	if got := ConteudoInt(n); got != 0 {
		t.Errorf("ConteudoInt(nil) = %d", got)
	}
	if got := ConteudoDe2(n); got != 0 {
		t.Errorf("ConteudoDe2(nil) = %v", got)
	}
	if got := ConteudoDataDef(n); !got.IsZero() {
		t.Errorf("ConteudoDataDef(nil) = %v", got)
	}
	if n.Existe() {
		t.Error("Existe() em nil deveria ser false")
	}
}

func TestPrimeiroDe_RespeitaOrdemDosNomes(t *testing.T) {
	// Grupo achatado por CST: vale a ordem dos NOMES, nao a do documento.
	xml := `<imposto><ICMS60><CST>60</CST></ICMS60><ICMS00><CST>00</CST></ICMS00></imposto>`
	doc := parsear(t, xml)

	n := doc.Root.PrimeiroDe("ICMS00", "ICMS10", "ICMS60")
	if n == nil || n.Nome != "ICMS00" {
		t.Fatalf("PrimeiroDe devolveu %v, esperado ICMS00 (primeiro NOME da lista)", n)
	}

	n = doc.Root.PrimeiroDe("ICMS10", "ICMS60", "ICMS00")
	if n == nil || n.Nome != "ICMS60" {
		t.Fatalf("PrimeiroDe devolveu %v, esperado ICMS60", n)
	}

	if doc.Root.PrimeiroDe("ICMS20", "ICMS40") != nil {
		t.Error("PrimeiroDe sem acerto deveria devolver nil")
	}
}

func TestOuterXML_ByteExato(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	ide := doc.Root.FindAnyNs("infNFGas").FindAnyNs("ide")

	got := ide.OuterXML()
	if !strings.HasPrefix(got, "<ide>") || !strings.HasSuffix(got, "</ide>") {
		t.Fatalf("OuterXML nao delimita o elemento: %q", got)
	}
	if !strings.Contains(got, "<cUF>35</cUF>") {
		t.Errorf("OuterXML perdeu conteudo: %q", got)
	}
	// Precisa ser reparseavel -- e o que permite reaproveitar protNFGas e
	// procEventoNFGas noutra leitura.
	if _, err := ParseString(got); err != nil {
		t.Errorf("OuterXML nao e reparseavel: %v", err)
	}
}

func TestOuterXML_ElementoAutoFechado(t *testing.T) {
	doc := parsear(t, `<a><b attr="1"/></a>`)
	if got := doc.Root.FindAnyNs("b").OuterXML(); got != `<b attr="1"/>` {
		t.Errorf("OuterXML = %q, esperado <b attr=\"1\"/>", got)
	}
}

func TestCaminho(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	n := doc.Root.FindAnyNs("infNFGas").FindAnyNs("ide").FindAnyNs("dhEmi")
	if got := n.Caminho(); got != "NFGas/infNFGas/ide/dhEmi" {
		t.Errorf("Caminho = %q", got)
	}
}

func TestParse_Erros(t *testing.T) {
	casos := []struct {
		nome string
		xml  string
		err  error
	}{
		{"vazio", "", ErrXMLVazio},
		{"so espaco", "   \n\t ", ErrXMLVazio},
		{"tag nao fechada", "<a><b></a>", ErrXMLInvalido},
		{"lixo", "nao sou xml <<<", ErrXMLInvalido},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, err := ParseString(c.xml)
			if !errors.Is(err, c.err) {
				t.Errorf("erro = %v, esperado %v", err, c.err)
			}
		})
	}
}

func TestParse_ToleraBOM(t *testing.T) {
	doc, err := ParseString(bomUTF8 + `<a><b>1</b></a>`)
	if err != nil {
		t.Fatalf("BOM deveria ser tolerado: %v", err)
	}
	if doc.Root.Nome != "a" {
		t.Errorf("raiz = %q", doc.Root.Nome)
	}
}

// ---------------------------------------------------------------------------
// Tipos de campo
// ---------------------------------------------------------------------------

func nodeDe(t *testing.T, conteudo string) *Node {
	t.Helper()
	doc := parsear(t, "<v>"+conteudo+"</v>")
	return doc.Root
}

func TestConteudoInt_AplicaOnlyNumber(t *testing.T) {
	// tcInt passa o conteudo por OnlyNumber ANTES de converter. Nao e Atoi.
	casos := []struct {
		conteudo string
		esperado int
	}{
		{"35", 35},
		{" 35 ", 35},
		{"3-5", 35},
		{"n35", 35},
		{"1.234", 1234},
		{"", 0},
		{"abc", 0},
	}
	for _, c := range casos {
		if got := ConteudoInt(nodeDe(t, c.conteudo)); got != c.esperado {
			t.Errorf("ConteudoInt(%q) = %d, esperado %d", c.conteudo, got, c.esperado)
		}
	}
}

func TestConteudoDec_PontoDecimal(t *testing.T) {
	casos := []struct {
		conteudo string
		casas    int
		esperado float64
	}{
		{"1234.56", 2, 1234.56},
		{"1234,56", 2, 1234.56},
		{"0.0001", 4, 0.0001},
		{"123.4567891", 10, 123.4567891},
		{"", 2, 0},
		{"abc", 2, 0},
	}
	for _, c := range casos {
		if got := ConteudoDec(nodeDe(t, c.conteudo), c.casas); got != c.esperado {
			t.Errorf("ConteudoDec(%q, %d) = %v, esperado %v", c.conteudo, c.casas, got, c.esperado)
		}
	}
}

func TestConteudoDec_FloatIsIntString(t *testing.T) {
	// Com o flag ligado, o conteudo e inteiro com decimais implicitas.
	n := nodeDe(t, "1234")
	n.FloatIsIntString = true

	if got := ConteudoDec(n, 2); got != 12.34 {
		t.Errorf("ConteudoDec com FloatIsIntString = %v, esperado 12.34", got)
	}
	if got := ConteudoDec(n, 4); got != 0.1234 {
		t.Errorf("ConteudoDec com FloatIsIntString e 4 casas = %v, esperado 0.1234", got)
	}

	n.FloatIsIntString = false
	if got := ConteudoDec(n, 2); got != 1234 {
		t.Errorf("ConteudoDec sem o flag = %v, esperado 1234", got)
	}
}

func TestConteudoBool(t *testing.T) {
	casos := map[string]bool{"true": true, "TRUE": true, "True": true, "false": false, "": false, "1": false}
	for conteudo, esperado := range casos {
		if got := ConteudoBool(nodeDe(t, conteudo)); got != esperado {
			t.Errorf("ConteudoBool(%q) = %v, esperado %v", conteudo, got, esperado)
		}
	}
}

func TestConteudoCNPJCPF_PrefereCNPJ(t *testing.T) {
	doc := parsear(t, `<emit><CNPJ>11222333000181</CNPJ><CPF>52998224725</CPF></emit>`)
	if got := ConteudoCNPJCPF(doc.Root); got != "11222333000181" {
		t.Errorf("com CNPJ e CPF devolveu %q, esperado o CNPJ", got)
	}

	doc = parsear(t, `<emit><CNPJ></CNPJ><CPF>52998224725</CPF></emit>`)
	if got := ConteudoCNPJCPF(doc.Root); got != "52998224725" {
		t.Errorf("com CNPJ vazio devolveu %q, esperado o CPF", got)
	}

	doc = parsear(t, `<emit></emit>`)
	if got := ConteudoCNPJCPF(doc.Root); got != "" {
		t.Errorf("sem nenhum devolveu %q", got)
	}
}

func TestConteudoHora(t *testing.T) {
	if got := ConteudoHora(nodeDe(t, "10:30:45")); got != 10*time.Hour+30*time.Minute+45*time.Second {
		t.Errorf("ConteudoHora = %v", got)
	}
	if got := ConteudoHora(nodeDe(t, "")); got != 0 {
		t.Errorf("ConteudoHora vazio = %v", got)
	}
}

// ---------------------------------------------------------------------------
// Data e hora
// ---------------------------------------------------------------------------

func TestEncodeDataHora_Formatos(t *testing.T) {
	casos := []struct {
		nome                          string
		entrada                       string
		ano, mes, dia, hora, min, seg int
		offsetSeg                     int
	}{
		{"data ISO", "2026-03-15", 2026, 3, 15, 0, 0, 0, 0},
		{"data compacta", "20260315", 2026, 3, 15, 0, 0, 0, 0},
		{"competencia", "202603", 2026, 3, 1, 0, 0, 0, 0},
		{"data e hora", "2026-03-15T10:30:45", 2026, 3, 15, 10, 30, 45, 0},
		{"com offset", "2026-03-15T10:30:45-03:00", 2026, 3, 15, 10, 30, 45, -3 * 3600},
		{"com offset positivo", "2026-03-15T10:30:45+02:00", 2026, 3, 15, 10, 30, 45, 2 * 3600},
		{"com Z", "2026-03-15T10:30:45Z", 2026, 3, 15, 10, 30, 45, 0},
		{"com fracao", "2026-03-15T10:30:45.123", 2026, 3, 15, 10, 30, 45, 0},
		{"separado por espaco", "2026-03-15 10:30:45", 2026, 3, 15, 10, 30, 45, 0},
		{"brasileiro", "15/03/2026", 2026, 3, 15, 0, 0, 0, 0},
		{"bissexto", "2024-02-29", 2024, 2, 29, 0, 0, 0, 0},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got, err := EncodeDataHora(c.entrada)
			if err != nil {
				t.Fatalf("EncodeDataHora(%q): %v", c.entrada, err)
			}
			if got.Year() != c.ano || int(got.Month()) != c.mes || got.Day() != c.dia {
				t.Errorf("data = %04d-%02d-%02d, esperado %04d-%02d-%02d",
					got.Year(), got.Month(), got.Day(), c.ano, c.mes, c.dia)
			}
			if got.Hour() != c.hora || got.Minute() != c.min || got.Second() != c.seg {
				t.Errorf("hora = %02d:%02d:%02d, esperado %02d:%02d:%02d",
					got.Hour(), got.Minute(), got.Second(), c.hora, c.min, c.seg)
			}
			_, off := got.Zone()
			if off != c.offsetSeg {
				t.Errorf("offset = %ds, esperado %ds", off, c.offsetSeg)
			}
		})
	}
}

func TestEncodeDataHora_PreservaHoraDeParede(t *testing.T) {
	// O ACBr descarta o fuso e guarda a hora de parede. Aqui o fuso e
	// preservado, mas a hora de parede continua a mesma -- e o que garante
	// que nenhum campo lido muda de valor em relacao ao Delphi.
	got, err := EncodeDataHora("2026-03-15T23:59:59-03:00")
	if err != nil {
		t.Fatal(err)
	}
	if got.Hour() != 23 || got.Day() != 15 {
		t.Errorf("hora de parede deslocada: %v", got)
	}
}

func TestEncodeDataHora_VazioNaoEErro(t *testing.T) {
	got, err := EncodeDataHora("")
	if err != nil {
		t.Fatalf("vazio nao deveria dar erro: %v", err)
	}
	if !got.IsZero() {
		t.Errorf("vazio deveria dar tempo zero, deu %v", got)
	}
}

func TestParseCompetencia(t *testing.T) {
	got, err := ParseCompetencia("202603")
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2026 || got.Month() != time.March || got.Day() != 1 {
		t.Errorf("ParseCompetencia = %v, esperado 2026-03-01", got)
	}

	invalidos := []string{"20260", "202613", "202600", "abcdef", "000003"}
	for _, s := range invalidos {
		if _, err := ParseCompetencia(s); err == nil {
			t.Errorf("ParseCompetencia(%q) deveria falhar", s)
		}
		if got := ParseCompetenciaDef(s); !got.IsZero() {
			t.Errorf("ParseCompetenciaDef(%q) = %v, esperado tempo zero", s, got)
		}
	}

	if got, err := ParseCompetencia(""); err != nil || !got.IsZero() {
		t.Errorf("competencia vazia = (%v, %v), esperado (zero, nil)", got, err)
	}
}

func TestFormatarCompetencia(t *testing.T) {
	if got := FormatarCompetencia(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)); got != "202603" {
		t.Errorf("FormatarCompetencia = %q", got)
	}
	if got := FormatarCompetencia(time.Time{}); got != "" {
		t.Errorf("FormatarCompetencia(zero) = %q, esperado vazio", got)
	}
}

// ---------------------------------------------------------------------------
// Conversoes numericas e de texto
// ---------------------------------------------------------------------------

func TestStringDecimalToFloat(t *testing.T) {
	casos := []struct {
		valor    string
		casas    int
		esperado float64
	}{
		{"10000", 2, 100},
		{"123", 2, 1.23},
		{"1", 2, 0.01},
		{"1", 4, 0.0001},
		{"123456", 0, 123456},
		{"-123", 2, -1.23},
	}
	for _, c := range casos {
		got, err := StringDecimalToFloat(c.valor, c.casas)
		if err != nil {
			t.Fatalf("StringDecimalToFloat(%q, %d): %v", c.valor, c.casas, err)
		}
		if got != c.esperado {
			t.Errorf("StringDecimalToFloat(%q, %d) = %v, esperado %v", c.valor, c.casas, got, c.esperado)
		}
	}
}

func TestStringToFloat_SeparadorDecimalEOUltimo(t *testing.T) {
	casos := map[string]float64{
		"1234.56":  1234.56,
		"1234,56":  1234.56,
		"1.234,56": 1234.56,
		"1,234.56": 1234.56,
		"1234":     1234,
		"-10.5":    -10.5,
	}
	for entrada, esperado := range casos {
		got, err := StringToFloat(entrada)
		if err != nil {
			t.Fatalf("StringToFloat(%q): %v", entrada, err)
		}
		if got != esperado {
			t.Errorf("StringToFloat(%q) = %v, esperado %v", entrada, got, esperado)
		}
	}
	if got := StringToFloatDef("xyz", 7); got != 7 {
		t.Errorf("StringToFloatDef invalido = %v, esperado o padrao 7", got)
	}
}

func TestRemoverCDATA_TodasVsPrimeira(t *testing.T) {
	entrada := "<![CDATA[a]]><![CDATA[b]]>"

	if got := RemoverCDATA(entrada); got != "ab" {
		t.Errorf("RemoverCDATA = %q, esperado ab", got)
	}
	// O ACBr usa StringReplace SEM rfReplaceAll ao limpar qrCodNFGas: so a
	// primeira ocorrencia de cada marcacao sai. Comportamento deliberado.
	if got := RemoverCDATAPrimeiraOcorrencia(entrada); got != "a<![CDATA[b]]>" {
		t.Errorf("RemoverCDATAPrimeiraOcorrencia = %q, esperado a<![CDATA[b]]>", got)
	}
}

func TestRemoverDeclaracaoXML(t *testing.T) {
	if got := RemoverDeclaracaoXML(`<?xml version="1.0"?><a/>`); got != "<a/>" {
		t.Errorf("RemoverDeclaracaoXML = %q", got)
	}
	if got := RemoverDeclaracaoXML("<a/>"); got != "<a/>" {
		t.Errorf("sem prologo deveria ficar igual, deu %q", got)
	}
}

func TestRemoverBOM(t *testing.T) {
	if got := string(RemoverBOM([]byte(bomUTF8 + "abc"))); got != "abc" {
		t.Errorf("RemoverBOM = %q", got)
	}
	if got := string(RemoverBOM([]byte("abc"))); got != "abc" {
		t.Errorf("sem BOM deveria ficar igual, deu %q", got)
	}
}

// ---------------------------------------------------------------------------
// Enums compartilhados
// ---------------------------------------------------------------------------

func TestEnums_CodigoDoLeiaute(t *testing.T) {
	if got := TaProducao.String(); got != "1" {
		t.Errorf("TaProducao = %q", got)
	}
	if got := TaHomologacao.String(); got != "2" {
		t.Errorf("TaHomologacao = %q", got)
	}
	if got := TeOffLine.String(); got != "9" {
		t.Errorf("TeOffLine = %q", got)
	}
	if got := TiSim.String(); got != "1" {
		t.Errorf("TiSim = %q", got)
	}
	if got := TiNao.String(); got != "0" {
		t.Errorf("TiNao = %q", got)
	}
	if got := TieNenhum.String(); got != "" {
		t.Errorf("TieNenhum = %q, esperado vazio", got)
	}
	if got := OeVazio.String(); got != "" {
		t.Errorf("OeVazio = %q, esperado vazio", got)
	}
	if got := OeReservadoParaUsoFuturo.String(); got != "9" {
		t.Errorf("OeReservadoParaUsoFuturo = %q", got)
	}
	if got := TpaNenhum.String(); got != "" {
		t.Errorf("TpaNenhum = %q, esperado vazio", got)
	}
	if got := Pis99.String(); got != "99" {
		t.Errorf("Pis99 = %q", got)
	}
	if got := Cof49.String(); got != "49" {
		t.Errorf("Cof49 = %q", got)
	}
}

func TestEnums_ParseIdaEVolta(t *testing.T) {
	for i := TeNormal; i <= TeOffLine; i++ {
		v, err := ParseTipoEmissao(i.String())
		if err != nil || v != i {
			t.Errorf("TipoEmissao %d nao sobreviveu ao round-trip: %v %v", i, v, err)
		}
	}
	for i := Pis01; i <= Pis99; i++ {
		v, err := ParseCSTPis(i.String())
		if err != nil || v != i {
			t.Errorf("CSTPis %d nao sobreviveu ao round-trip: %v %v", i, v, err)
		}
	}
	for i := Cof01; i <= Cof99; i++ {
		v, err := ParseCSTCofins(i.String())
		if err != nil || v != i {
			t.Errorf("CSTCofins %d nao sobreviveu ao round-trip: %v %v", i, v, err)
		}
	}
}

func TestEnums_ParseInvalidoDevolveErro(t *testing.T) {
	if _, err := ParseTipoAmbiente("3"); !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("erro = %v, esperado ErrEnumInvalido", err)
	}
	if _, err := ParseCSTPis("00"); !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("erro = %v, esperado ErrEnumInvalido", err)
	}
	// IndicadorEx e TpPagAnt aceitam string vazia como membro valido.
	if v, err := ParseIndicadorEx(""); err != nil || v != TieNenhum {
		t.Errorf("ParseIndicadorEx(\"\") = (%v, %v), esperado (TieNenhum, nil)", v, err)
	}
	if v, err := ParseTpPagAnt(""); err != nil || v != TpaNenhum {
		t.Errorf("ParseTpPagAnt(\"\") = (%v, %v), esperado (TpaNenhum, nil)", v, err)
	}
}

func TestCSTIcms_TabelasEntradaESaidaDivergem(t *testing.T) {
	// Divergencia deliberada do ACBr: cinco membros tem codigo diferente na
	// entrada e na saida. String() e ParseCSTIcms nao sao inversas.
	casos := []struct {
		membro  CSTIcms
		entrada string
		saida   string
	}{
		{CSTPart10, "10part", "10"},
		{CSTPart90, "90part", "90"},
		{CSTPart20, "20part", "20"},
		{CSTRep41, "41rep", "41"},
		{CSTRep60, "60rep", "60"},
		{CSTICMSOutraUF, "91", "90"},
		{CST00, "00", "00"},
		{CSTICMSSN, "SN", "SN"},
		{CSTVazio, "", ""},
	}
	for _, c := range casos {
		if got := c.membro.CodigoEntrada(); got != c.entrada {
			t.Errorf("CodigoEntrada(%d) = %q, esperado %q", c.membro, got, c.entrada)
		}
		if got := c.membro.String(); got != c.saida {
			t.Errorf("String(%d) = %q, esperado %q", c.membro, got, c.saida)
		}
		v, err := ParseCSTIcms(c.entrada)
		if err != nil {
			t.Fatalf("ParseCSTIcms(%q): %v", c.entrada, err)
		}
		if v != c.membro {
			t.Errorf("ParseCSTIcms(%q) = %d, esperado %d", c.entrada, v, c.membro)
		}
	}
}

// ---------------------------------------------------------------------------
// Validacao de documentos
// ---------------------------------------------------------------------------

func TestValidarCPF(t *testing.T) {
	if err := ValidarCPF("52998224725"); err != nil {
		t.Errorf("CPF valido recusado: %v", err)
	}
	if err := ValidarCPF("529.982.247-25"); err != nil {
		t.Errorf("CPF formatado recusado: %v", err)
	}
	for _, cpf := range []string{"52998224724", "11111111111", "123", ""} {
		if err := ValidarCPF(cpf); !errors.Is(err, ErrDocumentoInvalido) {
			t.Errorf("CPF %q deveria ser invalido, erro = %v", cpf, err)
		}
	}
}

func TestValidarCNPJ_Numerico(t *testing.T) {
	if err := ValidarCNPJ("11222333000181"); err != nil {
		t.Errorf("CNPJ valido recusado: %v", err)
	}
	if err := ValidarCNPJ("11.222.333/0001-81"); err != nil {
		t.Errorf("CNPJ formatado recusado: %v", err)
	}
	for _, cnpj := range []string{"11222333000182", "00000000000000", "123"} {
		if err := ValidarCNPJ(cnpj); !errors.Is(err, ErrDocumentoInvalido) {
			t.Errorf("CNPJ %q deveria ser invalido, erro = %v", cnpj, err)
		}
	}
}

func TestValidarCNPJ_Alfanumerico(t *testing.T) {
	// Exemplo oficial do CNPJ alfanumerico: 12.ABC.345/01DE-35.
	if err := ValidarCNPJ("12ABC34501DE35"); err != nil {
		t.Errorf("CNPJ alfanumerico valido recusado: %v", err)
	}
	if err := ValidarCNPJ("12.ABC.345/01DE-35"); err != nil {
		t.Errorf("CNPJ alfanumerico formatado recusado: %v", err)
	}
	if err := ValidarCNPJ("12ABC34501DE36"); !errors.Is(err, ErrDocumentoInvalido) {
		t.Error("CNPJ alfanumerico com DV errado deveria ser recusado")
	}
	// Os dois digitos verificadores continuam sendo sempre numericos.
	if err := ValidarCNPJ("12ABC34501DEAB"); !errors.Is(err, ErrDocumentoInvalido) {
		t.Error("DV nao numerico deveria ser recusado")
	}
}

func TestValidarCNPJouCPF(t *testing.T) {
	if err := ValidarCNPJouCPF(""); err != nil {
		t.Errorf("documento vazio nao deveria ser erro: %v", err)
	}
	if err := ValidarCNPJouCPF("52998224725"); err != nil {
		t.Errorf("CPF recusado: %v", err)
	}
	if err := ValidarCNPJouCPF("11222333000181"); err != nil {
		t.Errorf("CNPJ recusado: %v", err)
	}
	if err := ValidarCNPJouCPF("1234567"); !errors.Is(err, ErrDocumentoInvalido) {
		t.Error("documento de 7 digitos deveria ser invalido")
	}
}

func TestCodigoUF(t *testing.T) {
	if !ValidarCodigoUF(35) || SiglaUF(35) != "SP" {
		t.Error("35 deveria ser SP")
	}
	if ValidarCodigoUF(99) {
		t.Error("99 nao deveria ser UF valida")
	}
	if got := CodigoUF("rs"); got != 43 {
		t.Errorf("CodigoUF(rs) = %d, esperado 43", got)
	}
	if got := CodigoUF("XX"); got != 0 {
		t.Errorf("CodigoUF(XX) = %d, esperado 0", got)
	}
}

// ---------------------------------------------------------------------------
// Chave de acesso
// ---------------------------------------------------------------------------

func TestRemoverLiteralChave(t *testing.T) {
	casos := map[string]string{
		"NFGas35260311222333000181760010000000011100000012": "35260311222333000181760010000000011100000012",
		"NFe35260311222333000181550010000000011100000012":   "35260311222333000181550010000000011100000012",
		"35260311222333000181760010000000011100000012":      "35260311222333000181760010000000011100000012",
		"":          "",
		"SEMDIGITO": "",
	}
	for entrada, esperado := range casos {
		if got := RemoverLiteralChave(entrada); got != esperado {
			t.Errorf("RemoverLiteralChave(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}

func TestDigitoChaveAcesso_CasosCalculaveisAMao(t *testing.T) {
	// Pesos: "4329876543298765432987654329876543298765432".
	casos := []struct {
		nome     string
		base43   string
		esperado int
	}{
		// soma 0 -> resto 0 -> digito 0
		{"tudo zero", strings.Repeat("0", 43), 0},
		// 1 na posicao 1 (peso 4) -> soma 4 -> resto 4 -> 11-4 = 7
		{"um na primeira posicao", "1" + strings.Repeat("0", 42), 7},
		// 1 na posicao 43 (peso 2) -> soma 2 -> resto 2 -> 11-2 = 9
		{"um na ultima posicao", strings.Repeat("0", 42) + "1", 9},
		// 3 na posicao 1 (peso 4) -> soma 12 -> resto 1 -> digito 0
		{"resto 1 vira zero", "3" + strings.Repeat("0", 42), 0},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got, err := DigitoChaveAcesso(c.base43)
			if err != nil {
				t.Fatalf("DigitoChaveAcesso: %v", err)
			}
			if got != c.esperado {
				t.Errorf("digito = %d, esperado %d", got, c.esperado)
			}
		})
	}
}

func TestDigitoChaveAcesso_TamanhoErrado(t *testing.T) {
	if _, err := DigitoChaveAcesso("123"); !errors.Is(err, ErrChaveInvalida) {
		t.Errorf("erro = %v, esperado ErrChaveInvalida", err)
	}
}

func TestValidarChaveAcesso(t *testing.T) {
	base := "3526" + "03" + "11222333000181" + "76" + "001" + "000000001" + "1" + "1" + "0000001"
	if len(base) != 43 {
		t.Fatalf("base de teste tem %d caracteres, deveria ter 43", len(base))
	}
	dv, err := DigitoChaveAcesso(base)
	if err != nil {
		t.Fatal(err)
	}
	chave := base + string(rune('0'+dv))

	if err := ValidarChaveAcesso(chave); err != nil {
		t.Errorf("chave valida recusada: %v", err)
	}
	if err := ValidarChaveAcesso("NFGas" + chave); err != nil {
		t.Errorf("chave com literal recusada: %v", err)
	}

	// Digito verificador trocado.
	ruim := base + string(rune('0'+(dv+1)%10))
	if err := ValidarChaveAcesso(ruim); !errors.Is(err, ErrChaveInvalida) {
		t.Error("digito verificador errado deveria ser recusado")
	}
	// Tamanho errado.
	if err := ValidarChaveAcesso(base); !errors.Is(err, ErrChaveInvalida) {
		t.Error("chave de 43 caracteres deveria ser recusada")
	}
	// UF inexistente.
	if err := ValidarChaveAcesso("99" + chave[2:]); !errors.Is(err, ErrChaveInvalida) {
		t.Error("codigo de UF 99 deveria ser recusado")
	}
}

func TestValidarAAMM(t *testing.T) {
	if err := ValidarAAMM("2603"); err != nil {
		t.Errorf("2603 deveria ser valido: %v", err)
	}
	for _, s := range []string{"2613", "2600", "260", "abcd"} {
		if err := ValidarAAMM(s); !errors.Is(err, ErrChaveInvalida) {
			t.Errorf("AAMM %q deveria ser invalido", s)
		}
	}
}

func TestPreencherZerosEsquerda(t *testing.T) {
	casos := []struct {
		entrada  string
		tamanho  int
		esperado string
	}{
		{"1", 3, "001"},
		{"123", 3, "123"},
		{"12345", 3, "345"},
		{"", 2, "00"},
	}
	for _, c := range casos {
		if got := PreencherZerosEsquerda(c.entrada, c.tamanho); got != c.esperado {
			t.Errorf("PreencherZerosEsquerda(%q, %d) = %q, esperado %q", c.entrada, c.tamanho, got, c.esperado)
		}
	}
	if got := FormatarInteiroZeros(7, 3); got != "007" {
		t.Errorf("FormatarInteiroZeros = %q", got)
	}
}

// ---------------------------------------------------------------------------
// Signature e ProcDFe
// ---------------------------------------------------------------------------

func TestLerSignature(t *testing.T) {
	doc := parsear(t, xmlExemplo)
	var sig Signature
	LerSignature(doc.Root.Find("Signature"), &sig)

	if sig.URI != "#NFGas352603112223330001817600100000000111000000" {
		t.Errorf("URI = %q", sig.URI)
	}
	if sig.DigestValue != "ABC123" {
		t.Errorf("DigestValue = %q", sig.DigestValue)
	}
	if sig.SignatureValue != "SIG==" {
		t.Errorf("SignatureValue = %q", sig.SignatureValue)
	}
	if sig.X509Certificate != "CERT==" {
		t.Errorf("X509Certificate = %q", sig.X509Certificate)
	}
	if sig.Vazia() {
		t.Error("assinatura lida nao deveria estar vazia")
	}
}

func TestLerSignature_SemAssinaturaNaoEstoura(t *testing.T) {
	doc := parsear(t, `<NFGas><infNFGas/></NFGas>`)
	var sig Signature
	LerSignature(doc.Root.Find("Signature"), &sig) // node nil
	if !sig.Vazia() {
		t.Error("sem elemento Signature a struct deveria ficar vazia")
	}
}

func TestLerInfProt(t *testing.T) {
	xml := `<protNFGas versao="1.00"><infProt Id="ID123">
		<tpAmb>2</tpAmb><verAplic>SP_1.0</verAplic>
		<chNFGas>35260311222333000181760010000000011100000012</chNFGas>
		<dhRecbto>2026-03-15T11:00:00-03:00</dhRecbto>
		<nProt>135260000000001</nProt><digVal>DIG==</digVal>
		<cStat>100</cStat><xMotivo>Autorizado o uso</xMotivo>
		<cMsg>1</cMsg><xMsg>obs</xMsg>
	</infProt></protNFGas>`
	doc := parsear(t, xml)

	proc := NovoProcDFe("1.00", "http://www.portalfiscal.inf.br/nfgas", "NFGasProc", "NFGas")
	LerInfProt(doc.Root, proc, "chNFGas")

	if proc.TpAmb != TaHomologacao {
		t.Errorf("tpAmb = %v", proc.TpAmb)
	}
	if proc.VerAplic != "SP_1.0" {
		t.Errorf("verAplic = %q", proc.VerAplic)
	}
	if proc.ChDFe != "35260311222333000181760010000000011100000012" {
		t.Errorf("chDFe = %q", proc.ChDFe)
	}
	if proc.CStat != 100 || proc.XMotivo != "Autorizado o uso" {
		t.Errorf("cStat/xMotivo = %d/%q", proc.CStat, proc.XMotivo)
	}
	if proc.CMsg != 1 || proc.XMsg != "obs" {
		t.Errorf("cMsg/xMsg = %d/%q", proc.CMsg, proc.XMsg)
	}
	if proc.DhRecbto.Hour() != 11 {
		t.Errorf("dhRecbto = %v", proc.DhRecbto)
	}
	if proc.Vazio() {
		t.Error("protocolo lido nao deveria estar vazio")
	}
	// Campos que o ACBr deliberadamente NAO le.
	if proc.ID != "" {
		t.Errorf("infProt/@Id nao e lido pelo ACBr, mas veio %q", proc.ID)
	}
}

// ---------------------------------------------------------------------------
// INI
// ---------------------------------------------------------------------------

const iniExemplo = "[ide]\r\ncUF=35\r\nnNF=123\r\nvProd=1234.56\r\ndhEmi=2026-03-15T10:30:00-03:00\r\n\r\n" +
	"[emit]\r\nCNPJ=11222333000181\r\nxNome=Empresa Teste\r\n"

func TestINI_LeituraBasica(t *testing.T) {
	ini, err := LerINIString(iniExemplo)
	if err != nil {
		t.Fatalf("LerINIString: %v", err)
	}

	if !ini.SecaoExiste("ide") || !ini.SecaoExiste("emit") {
		t.Error("secoes ide e emit deveriam existir")
	}
	if ini.SecaoExiste("dest") {
		t.Error("secao dest nao deveria existir")
	}
	if got := ini.LerInteiro("ide", "cUF", 0); got != 35 {
		t.Errorf("cUF = %d", got)
	}
	if got := ini.LerFloat("ide", "vProd", 0); got != 1234.56 {
		t.Errorf("vProd = %v", got)
	}
	if got := ini.LerString("emit", "xNome", ""); got != "Empresa Teste" {
		t.Errorf("xNome = %q", got)
	}
	if got := ini.LerData("ide", "dhEmi", time.Time{}); got.Year() != 2026 || got.Hour() != 10 {
		t.Errorf("dhEmi = %v", got)
	}
}

func TestINI_PadraoQuandoAusente(t *testing.T) {
	ini, err := LerINIString(iniExemplo)
	if err != nil {
		t.Fatal(err)
	}
	if got := ini.LerInteiro("ide", "naoExiste", 99); got != 99 {
		t.Errorf("chave ausente = %d, esperado o padrao 99", got)
	}
	if got := ini.LerString("naoExiste", "nada", "pad"); got != "pad" {
		t.Errorf("secao ausente = %q, esperado o padrao", got)
	}
}

func TestINI_NomesSemDistinguirCaixa(t *testing.T) {
	// TMemIniFile nao distingue maiusculas de minusculas.
	ini, err := LerINIString(iniExemplo)
	if err != nil {
		t.Fatal(err)
	}
	if !ini.SecaoExiste("IDE") {
		t.Error("SecaoExiste deveria ignorar a caixa")
	}
	if got := ini.LerInteiro("IDE", "CUF", 0); got != 35 {
		t.Errorf("leitura com caixa diferente = %d, esperado 35", got)
	}
}

func TestINI_RoundTrip(t *testing.T) {
	ini := NovoINI()
	ini.GravarInteiro("ide", "cUF", 35)
	ini.GravarString("ide", "verProc", "OpenFiscalBR")
	ini.GravarFloat("ide", "vProd", 1234.5, 2)
	ini.GravarBool("ide", "teste", true)
	ini.GravarData("ide", "dhEmi", time.Date(2026, 3, 15, 10, 30, 0, 0, time.UTC), "2006-01-02T15:04:05")

	texto := ini.String()
	lido, err := LerINIString(texto)
	if err != nil {
		t.Fatalf("reler o INI gerado: %v", err)
	}

	if got := lido.LerInteiro("ide", "cUF", 0); got != 35 {
		t.Errorf("cUF apos round-trip = %d", got)
	}
	if got := lido.LerFloat("ide", "vProd", 0); got != 1234.5 {
		t.Errorf("vProd apos round-trip = %v", got)
	}
	if !lido.LerBool("ide", "teste", false) {
		t.Error("bool nao sobreviveu ao round-trip")
	}
	if got := lido.LerData("ide", "dhEmi", time.Time{}); got.Day() != 15 {
		t.Errorf("dhEmi apos round-trip = %v", got)
	}
	if got := lido.LerString("ide", "verProc", ""); got != "OpenFiscalBR" {
		t.Errorf("verProc apos round-trip = %q", got)
	}
}

func TestINI_GravarFloatUsaPonto(t *testing.T) {
	ini := NovoINI()
	ini.GravarFloat("s", "v", 1234.5, 2)
	if got := ini.LerString("s", "v", ""); got != "1234.50" {
		t.Errorf("GravarFloat gravou %q, esperado 1234.50 com ponto", got)
	}
}

func TestINI_OrdemPreservada(t *testing.T) {
	ini := NovoINI()
	ini.GravarString("b", "x", "1")
	ini.GravarString("a", "y", "2")
	ini.GravarString("b", "z", "3")

	if got := ini.Secoes(); len(got) != 2 || got[0] != "b" || got[1] != "a" {
		t.Errorf("Secoes = %v, esperado [b a]", got)
	}
	if got := ini.Chaves("b"); len(got) != 2 || got[0] != "x" || got[1] != "z" {
		t.Errorf("Chaves(b) = %v, esperado [x z]", got)
	}
}

func TestINI_SobrescreveChaveExistente(t *testing.T) {
	ini := NovoINI()
	ini.GravarString("s", "k", "1")
	ini.GravarString("s", "K", "2")
	if got := ini.LerString("s", "k", ""); got != "2" {
		t.Errorf("valor = %q, esperado 2 (mesma chave, caixa diferente)", got)
	}
	if got := ini.Chaves("s"); len(got) != 1 {
		t.Errorf("Chaves = %v, esperado uma so", got)
	}
}

func TestINI_LerArquivoOuString(t *testing.T) {
	dir := t.TempDir()
	caminho := filepath.Join(dir, "nota.ini")
	if err := os.WriteFile(caminho, []byte(iniExemplo), 0o644); err != nil {
		t.Fatal(err)
	}

	// Caminho de arquivo existente.
	ini, err := LerINIArquivoOuString(caminho)
	if err != nil {
		t.Fatalf("por caminho: %v", err)
	}
	if got := ini.LerInteiro("ide", "cUF", 0); got != 35 {
		t.Errorf("por caminho, cUF = %d", got)
	}

	// Conteudo direto.
	ini, err = LerINIArquivoOuString(iniExemplo)
	if err != nil {
		t.Fatalf("por conteudo: %v", err)
	}
	if got := ini.LerInteiro("ide", "cUF", 0); got != 35 {
		t.Errorf("por conteudo, cUF = %d", got)
	}
}

func TestINI_SalvarArquivo(t *testing.T) {
	ini := NovoINI()
	ini.GravarString("ide", "cUF", "35")

	caminho := filepath.Join(t.TempDir(), "saida.ini")
	if err := ini.SalvarArquivo(caminho); err != nil {
		t.Fatalf("SalvarArquivo: %v", err)
	}
	dados, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dados), "[ide]") || !strings.Contains(string(dados), "cUF=35") {
		t.Errorf("arquivo gravado = %q", dados)
	}
	if !strings.Contains(string(dados), "\r\n") {
		t.Error("o INI deveria ser gravado com CRLF")
	}
}

func TestINI_IgnoraComentarios(t *testing.T) {
	ini, err := LerINIString("; comentario\n# outro\n[s]\nk=1\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := ini.LerInteiro("s", "k", 0); got != 1 {
		t.Errorf("k = %d", got)
	}
	if len(ini.Secoes()) != 1 {
		t.Errorf("Secoes = %v, esperado so [s]", ini.Secoes())
	}
}

// ---------------------------------------------------------------------------
// Erros
// ---------------------------------------------------------------------------

func TestErroLeitura_CarregaCaminhoEDesembrulha(t *testing.T) {
	err := NovoErroLeitura("infNFGas/ide/dhEmi", "2026-99-99", ErrDataInvalida)

	if !errors.Is(err, ErrDataInvalida) {
		t.Error("errors.Is deveria alcancar o erro subjacente")
	}
	msg := err.Error()
	if !strings.Contains(msg, "infNFGas/ide/dhEmi") {
		t.Errorf("mensagem sem o caminho: %q", msg)
	}
	if !strings.Contains(msg, "2026-99-99") {
		t.Errorf("mensagem sem o valor: %q", msg)
	}

	var alvo *ErroLeitura
	if !errors.As(err, &alvo) {
		t.Fatal("errors.As deveria reconhecer *ErroLeitura")
	}
	if alvo.Caminho != "infNFGas/ide/dhEmi" {
		t.Errorf("Caminho = %q", alvo.Caminho)
	}
}
