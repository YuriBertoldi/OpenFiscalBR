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
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// chaveTeste e a chave usada nas fixtures. Layout da NFGas:
// cUF(2) AAMM(4) CNPJ(14) mod(2) serie(3) nNF(9) tpEmis(1) nSiteAutoriz(1)
// cNF(7) DV(1).
const chaveTeste = "35260311222333000181760010000000011000000018"

func carregarFixture(t *testing.T, nome string) []byte {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join("testdata", nome))
	if err != nil {
		t.Fatalf("abrir fixture %s: %v", nome, err)
	}
	return dados
}

func lerCompleta(t *testing.T) *NFGas {
	t.Helper()
	n, err := LerBytes(carregarFixture(t, "nfgas_completa.xml"))
	if err != nil {
		t.Fatalf("LerBytes: %v", err)
	}
	return n
}

// ---------------------------------------------------------------------------
// Enums -- codigo exato do leiaute
// ---------------------------------------------------------------------------

func TestEnums_CodigoDoLeiaute(t *testing.T) {
	casos := []struct {
		nome     string
		got      string
		esperado string
	}{
		{"Ve100", Ve100.String(), "1.00"},
		{"InContribuinte", InContribuinte.String(), "1"},
		{"InIsento", InIsento.String(), "2"},
		{"InNaoContribuinte", InNaoContribuinte.String(), "9"},
		{"Sa0", Sa0.String(), "0"},
		{"Sa9", Sa9.String(), "9"},
		{"IoMedia", IoMedia.String(), "1"},
		{"IoSemQuantidade", IoSemQuantidade.String(), "6"},
		{"Umm3", Umm3.String(), "1"},
		{"Umim3", Umim3.String(), "1"},
		{"UmiUnidade", UmiUnidade.String(), "2"},
		{"TmConsumidor", TmConsumidor.String(), "1"},
		{"TmIndependente", TmIndependente.String(), "3"},
		{"FnNormal", FnNormal.String(), "0"},
		{"FnSubstituicao", FnSubstituicao.String(), "3"},
		{"TiCativo", TiCativo.String(), "1"},
		{"TiParcialmenteLivre", TiParcialmenteLivre.String(), "3"},
		{"TcComercial", TcComercial.String(), "01"},
		{"TcRefinaria", TcRefinaria.String(), "10"},
		{"TcOutros", TcOutros.String(), "99"},
		{"MsErroLeitura", MsErroLeitura.String(), "01"},
		{"MsDecisaoReguladora", MsDecisaoReguladora.String(), "06"},
		{"VcDemandaMinima", VcDemandaMinima.String(), "1"},
		{"VcVolumeContratado", VcVolumeContratado.String(), "4"},
		{"TeMedidor", TeMedidor.String(), "1"},
		{"TeConversor", TeConversor.String(), "2"},
		{"TmTurbina", TmTurbina.String(), "1"},
		{"TmUltrasonico", TmUltrasonico.String(), "4"},
		{"TfFixa", TfFixa.String(), "1"},
		{"TfMedia", TfMedia.String(), "2"},
		{"TpProcAdmEstadual", TpProcAdmEstadual.String(), "0"},
		{"TpProcon", TpProcon.String(), "5"},
		{"TfNormal", TfNormal.String(), "1"},
		{"TfAgregador", TfAgregador.String(), "3"},
		{"Veqr000", Veqr000.String(), "0"},
		{"Veqr200", Veqr200.String(), "2"},
		{"TeCancelamento", TeCancelamento.String(), "110111"},
		{"TeNaoMapeado", TeNaoMapeado.String(), "-99999"},
		{"TeLiberacaoPrazoCancelado", TeLiberacaoPrazoCancelado.String(), "240170"},
	}
	for _, c := range casos {
		if c.got != c.esperado {
			t.Errorf("%s = %q, esperado %q", c.nome, c.got, c.esperado)
		}
	}
}

func TestFinalidade_NaoTemCodigos1e2(t *testing.T) {
	// O leiaute salta de 0 para 3. Aceitar "1" seria inventar codigo.
	if _, err := ParseFinalidadeNFGas("1"); !errors.Is(err, ErrEnumInvalido) {
		t.Error("finNFGas nao deveria aceitar 1")
	}
	if _, err := ParseFinalidadeNFGas("2"); !errors.Is(err, ErrEnumInvalido) {
		t.Error("finNFGas nao deveria aceitar 2")
	}
}

func TestEnums_RoundTrip(t *testing.T) {
	testar := func(nome string, total int, str func(int) string, parse func(string) (int, error)) {
		t.Helper()
		for i := 0; i < total; i++ {
			s := str(i)
			v, err := parse(s)
			if err != nil {
				t.Errorf("%s[%d]: Parse(%q): %v", nome, i, s, err)
				continue
			}
			if v != i {
				t.Errorf("%s[%d]: Parse(%q) = %d", nome, i, s, v)
			}
		}
	}

	testar("IndIEDest", 3,
		func(i int) string { return IndIEDest(i).String() },
		func(s string) (int, error) { v, e := ParseIndIEDest(s); return int(v), e })
	testar("SiteAutorizador", 10,
		func(i int) string { return SiteAutorizador(i).String() },
		func(s string) (int, error) { v, e := ParseSiteAutorizador(s); return int(v), e })
	testar("IndOrigemQtd", 6,
		func(i int) string { return IndOrigemQtd(i).String() },
		func(s string) (int, error) { v, e := ParseIndOrigemQtd(s); return int(v), e })
	testar("UMedItem", 2,
		func(i int) string { return UMedItem(i).String() },
		func(s string) (int, error) { v, e := ParseUMedItem(s); return int(v), e })
	testar("TpClasse", 11,
		func(i int) string { return TpClasse(i).String() },
		func(s string) (int, error) { v, e := ParseTpClasse(s); return int(v), e })
	testar("MotSub", 6,
		func(i int) string { return MotSub(i).String() },
		func(s string) (int, error) { v, e := ParseMotSub(s); return int(v), e })
	testar("VolContrat", 4,
		func(i int) string { return VolContrat(i).String() },
		func(s string) (int, error) { v, e := ParseVolContrat(s); return int(v), e })
	testar("TpMedidor", 4,
		func(i int) string { return TpMedidor(i).String() },
		func(s string) (int, error) { v, e := ParseTpMedidor(s); return int(v), e })
	testar("TpProc", 6,
		func(i int) string { return TpProc(i).String() },
		func(s string) (int, error) { v, e := ParseTpProc(s); return int(v), e })
	testar("TpFat", 3,
		func(i int) string { return TpFat(i).String() },
		func(s string) (int, error) { v, e := ParseTpFat(s); return int(v), e })
	testar("TipoEvento", 5,
		func(i int) string { return TipoEvento(i).String() },
		func(s string) (int, error) { v, e := ParseTipoEvento(s); return int(v), e })
}

func TestUMed_TipoDistintoDeUMedItem(t *testing.T) {
	// UMed (medicao) so admite m3; UMedItem (produto) admite tambem Unidade.
	if _, err := ParseUMed("2"); !errors.Is(err, ErrEnumInvalido) {
		t.Error("UMed nao deveria aceitar o codigo 2")
	}
	if v, err := ParseUMedItem("2"); err != nil || v != UmiUnidade {
		t.Errorf("UMedItem deveria aceitar 2: %v %v", v, err)
	}
	if Umm3.Descricao() != "m3" || UmiUnidade.Descricao() != "Unidade" {
		t.Error("descricoes de unidade incorretas")
	}
}

func TestModBC_ModBCST_MotDesICMS(t *testing.T) {
	// modBC tem membro nomeado para vazio; modBCST e motDesICMS usam
	// sentinela -1, reproduzindo TDeterminacaoBaseIcmsST(-1) do ACBr.
	if DbiNenhum.String() != "" {
		t.Errorf("DbiNenhum = %q, esperado vazio", DbiNenhum.String())
	}
	if v, err := ParseDeterminacaoBaseIcms(""); err != nil || v != DbiNenhum {
		t.Errorf("modBC vazio = (%v, %v)", v, err)
	}
	if v, err := ParseDeterminacaoBaseIcmsST(""); err != nil || v != DbisNenhum {
		t.Errorf("modBCST vazio = (%v, %v), esperado DbisNenhum", v, err)
	}
	if v, err := ParseMotivoDesoneracaoICMS(""); err != nil || v != MdiNenhum {
		t.Errorf("motDesICMS vazio = (%v, %v), esperado MdiNenhum", v, err)
	}
	if DbisNenhum != -1 || MdiNenhum != -1 {
		t.Error("os sentinelas de vazio deveriam valer -1")
	}

	// Codigos nao sequenciais de motDesICMS.
	casos := map[string]MotivoDesoneracaoICMS{
		"1": MdiTaxi, "9": MdiOutros, "12": MdiOrgaoFomento,
		"16": MdiOlimpiadaRio2016, "90": MdiSolicitadoFisco,
	}
	for codigo, esperado := range casos {
		v, err := ParseMotivoDesoneracaoICMS(codigo)
		if err != nil || v != esperado {
			t.Errorf("motDesICMS %q = (%v, %v), esperado %v", codigo, v, err, esperado)
		}
		if v.String() != codigo {
			t.Errorf("String(%v) = %q, esperado %q", v, v.String(), codigo)
		}
	}
	// 13, 14 e 15 nao existem na tabela.
	for _, codigo := range []string{"13", "14", "15", "0"} {
		if _, err := ParseMotivoDesoneracaoICMS(codigo); !errors.Is(err, ErrEnumInvalido) {
			t.Errorf("motDesICMS %q deveria ser invalido", codigo)
		}
	}
}

func TestSchemaNFGas(t *testing.T) {
	// SchemaNFGasToStr remove o prefixo "sch" do nome do membro.
	if SchNFGas.String() != "NFGas" {
		t.Errorf("SchNFGas = %q", SchNFGas.String())
	}
	if SchconsSitNFGas.String() != "consSitNFGas" {
		t.Errorf("SchconsSitNFGas = %q", SchconsSitNFGas.String())
	}
	// SchemaEventoToStr so tem string para o cancelamento.
	if SchevCancNFGas.SchemaEvento() != "evCancNFGas" {
		t.Errorf("SchemaEvento(SchevCancNFGas) = %q", SchevCancNFGas.SchemaEvento())
	}
	if SchNFGas.SchemaEvento() != "" {
		t.Errorf("SchemaEvento(SchNFGas) = %q, esperado vazio", SchNFGas.SchemaEvento())
	}
	// Parse aceita com e sem prefixo, e corta em "_".
	for _, s := range []string{"NFGas", "schNFGas", "schNFGas_v1.00"} {
		if v, err := ParseSchemaNFGas(s); err != nil || v != SchNFGas {
			t.Errorf("ParseSchemaNFGas(%q) = (%v, %v)", s, v, err)
		}
	}
}

func TestLayOutNFGas_Schema(t *testing.T) {
	casos := map[LayOutNFGas]SchemaNFGas{
		LayNFGasStatusServico: SchconsStatServNFGas,
		LayNFGasRecepcao:      SchNFGas,
		LayNFGasConsulta:      SchconsSitNFGas,
		LayNFGasRetRecepcao:   SchretNFGas,
		LayNFGasEvento:        SchEventoNFGas,
		// Os servicos de URL caem no "else" do case original.
		LayNFGasURLQRCode:   SchErroNFGas,
		LayURLConsultaNFGas: SchErroNFGas,
	}
	for lay, esperado := range casos {
		if got := lay.Schema(); got != esperado {
			t.Errorf("%v.Schema() = %v, esperado %v", lay, got, esperado)
		}
	}
	if LayNFGasEvento.String() != "NFGasRecepcaoEvento" {
		t.Errorf("LayNFGasEvento = %q", LayNFGasEvento.String())
	}
}

// ---------------------------------------------------------------------------
// Leitura do XML -- documento completo
// ---------------------------------------------------------------------------

func TestLerXML_Identificacao(t *testing.T) {
	n := lerCompleta(t)

	if n.InfNFGas.ID != "NFGas"+chaveTeste {
		t.Errorf("Id = %q", n.InfNFGas.ID)
	}
	if n.InfNFGas.Versao != 1.00 {
		t.Errorf("versao = %v", n.InfNFGas.Versao)
	}
	if n.ChaveAcesso() != chaveTeste {
		t.Errorf("ChaveAcesso = %q", n.ChaveAcesso())
	}

	ide := n.Ide
	if ide.CUF != 35 || ide.Modelo != ModeloNFGas || ide.Serie != 1 || ide.NNF != 1 {
		t.Errorf("ide = %+v", ide)
	}
	if ide.CNF != 1 {
		t.Errorf("cNF = %d, esperado 1 (zeros a esquerda descartados)", ide.CNF)
	}
	if ide.CDV != 8 {
		t.Errorf("cDV = %d", ide.CDV)
	}
	if ide.TpAmb != pcn.TaHomologacao {
		t.Errorf("tpAmb = %v", ide.TpAmb)
	}
	if ide.TpEmis != pcn.TeNormal || ide.NSiteAutoriz != Sa0 {
		t.Errorf("tpEmis/nSiteAutoriz = %v/%v", ide.TpEmis, ide.NSiteAutoriz)
	}
	if ide.CMunFG != 3550308 || ide.FinNFGas != FnNormal || ide.TpFat != TfNormal {
		t.Errorf("cMunFG/finNFGas/tpFat = %d/%v/%v", ide.CMunFG, ide.FinNFGas, ide.TpFat)
	}
	if ide.TpPagAnt != pcn.TpaPagServicoContinuado {
		t.Errorf("tpPagAnt = %v", ide.TpPagAnt)
	}
	if ide.VerProc != "OpenFiscalBR 1.0" {
		t.Errorf("verProc = %q", ide.VerProc)
	}
}

func TestLerXML_DhEmiPreservaHoraEFuso(t *testing.T) {
	n := lerCompleta(t)
	d := n.Ide.DhEmi

	if d.Year() != 2026 || d.Month() != time.March || d.Day() != 15 {
		t.Errorf("data = %v", d)
	}
	if d.Hour() != 10 || d.Minute() != 30 {
		t.Errorf("hora de parede = %02d:%02d, esperado 10:30", d.Hour(), d.Minute())
	}
	if _, off := d.Zone(); off != -3*3600 {
		t.Errorf("offset = %ds, esperado -10800", off)
	}
}

func TestLerXML_GCompraGovNaIde(t *testing.T) {
	n := lerCompleta(t)
	g := n.Ide.GCompraGov

	if g.TpEnteGov != rtc.TcgEstados || g.TpOperGov != rtc.TogFornecimento {
		t.Errorf("gCompraGov enums = %v/%v", g.TpEnteGov, g.TpOperGov)
	}
	if g.PRedutor != 12.3456 {
		t.Errorf("pRedutor = %v", g.PRedutor)
	}
	if len(g.RefDFe) != 2 || g.RefDFe[1].RefDFeAnt != "CHAVEGOV2" {
		t.Errorf("refDFeAnt = %+v", g.RefDFe)
	}
}

func TestLerXML_EmitEDest(t *testing.T) {
	n := lerCompleta(t)

	if n.Emit.CNPJ != "11222333000181" {
		t.Errorf("emit/CNPJ = %q", n.Emit.CNPJ)
	}
	if n.Emit.ISUFEmit != "ISUF123" {
		t.Errorf("ISUFEmit = %q", n.Emit.ISUFEmit)
	}
	if n.Emit.EnderEmit.CEP != 1310100 {
		t.Errorf("CEP = %d (o zero a esquerda some, CEP e int)", n.Emit.EnderEmit.CEP)
	}
	if n.Emit.EnderEmit.CPais != 1058 || n.Emit.EnderEmit.XPais != "BRASIL" {
		t.Errorf("enderEmit le cPais/xPais: %+v", n.Emit.EnderEmit)
	}

	// Sem CNPJ, ObterCNPJCPF cai para CPF.
	if n.Dest.CNPJCPF != "52998224725" {
		t.Errorf("dest/CNPJCPF = %q, esperado o CPF", n.Dest.CNPJCPF)
	}
	if n.Dest.CNIS != "12345678901" || n.Dest.NB != "9876543210" {
		t.Errorf("cNIS/NB = %q/%q", n.Dest.CNIS, n.Dest.NB)
	}
	if n.Dest.EnderDest.CPais != 1058 {
		t.Errorf("enderDest deveria ler cPais: %d", n.Dest.EnderDest.CPais)
	}
}

func TestLerXML_IndIEDestNaoEhLidoDoXML(t *testing.T) {
	// O ACBr le indIEDest apenas do .ini. Replicado: mesmo presente no XML,
	// o campo fica no valor zero.
	xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `">
		<dest><xNome>X</xNome><indIEDest>9</indIEDest></dest></infNFGas></NFGas>`
	n, err := LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if n.Dest.IndIEDest != InContribuinte {
		t.Errorf("indIEDest = %v; o ACBr nao le esse campo do XML", n.Dest.IndIEDest)
	}
}

func TestLerXML_DestPrecedenciaDoIdentificador(t *testing.T) {
	base := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><dest>%s</dest></infNFGas></NFGas>`

	casos := []struct {
		nome      string
		corpo     string
		cnpjcpf   string
		idEstr    string
		tagOrigem string
	}{
		{
			"com CNPJ, ignora os alternativos",
			`<CNPJ>11222333000181</CNPJ><idOutros>OUT</idOutros><idEstrangeiro>EST</idEstrangeiro>`,
			"11222333000181", "", "",
		},
		{
			"sem CNPJ, usa idOutros",
			`<idOutros>OUT</idOutros><idEstrangeiro>EST</idEstrangeiro>`,
			"", "OUT", "idOutros",
		},
		{
			"sem CNPJ nem idOutros, usa idEstrangeiro",
			`<idEstrangeiro>EST</idEstrangeiro>`,
			"", "EST", "idEstrangeiro",
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			n, err := LerXMLString(strings.Replace(base, "%s", c.corpo, 1))
			if err != nil {
				t.Fatal(err)
			}
			if n.Dest.CNPJCPF != c.cnpjcpf {
				t.Errorf("CNPJCPF = %q, esperado %q", n.Dest.CNPJCPF, c.cnpjcpf)
			}
			if n.Dest.IDEstrangeiro != c.idEstr {
				t.Errorf("IDEstrangeiro = %q, esperado %q", n.Dest.IDEstrangeiro, c.idEstr)
			}
			if n.Dest.TagIDOrigem != c.tagOrigem {
				t.Errorf("TagIDOrigem = %q, esperado %q", n.Dest.TagIDOrigem, c.tagOrigem)
			}
			// idOutros e idEstrangeiro compartilham campo, como no Delphi.
			if n.Dest.IDOutros() != n.Dest.IDEstrangeiro {
				t.Error("IDOutros e IDEstrangeiro deveriam ser o mesmo campo")
			}
		})
	}
}

func TestLerXML_GSubECompetencias(t *testing.T) {
	n := lerCompleta(t)
	g := n.GSub

	if g.MotSub != MsErroLeitura {
		t.Errorf("motSub = %v", g.MotSub)
	}
	if g.GNF.Serie != "001" {
		t.Errorf("gNF/serie = %q -- e string, nao int, e o zero a esquerda importa", g.GNF.Serie)
	}
	if g.GNF.NNF != 5 {
		t.Errorf("gNF/nNF = %d", g.GNF.NNF)
	}
	esperado := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !g.GNF.CompetEmis.Equal(esperado) {
		t.Errorf("CompetEmis = %v, esperado %v", g.GNF.CompetEmis, esperado)
	}
	if g.GNF.CompetApur.Month() != time.January {
		t.Errorf("CompetApur = %v", g.GNF.CompetApur)
	}
}

func TestLerXML_CompetenciaInvalidaFicaZerada(t *testing.T) {
	// O ACBr so aceita AAAAMM com 6 digitos, ano > 0 e mes de 1 a 12;
	// fora disso deixa zerado, sem erro.
	for _, valor := range []string{"20261", "202613", "202600", "abcdef", ""} {
		xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `">` +
			`<gSub><gNF><CompetEmis>` + valor + `</CompetEmis></gNF></gSub></infNFGas></NFGas>`
		n, err := LerXMLString(xml)
		if err != nil {
			t.Fatalf("%q: %v", valor, err)
		}
		if !n.GSub.GNF.CompetEmis.IsZero() {
			t.Errorf("CompetEmis(%q) = %v, esperado tempo zero", valor, n.GSub.GNF.CompetEmis)
		}
	}
}

func TestLerXML_ColecoesComDadoEmAtributo(t *testing.T) {
	n := lerCompleta(t)

	if len(n.GVolContrat) != 2 {
		t.Fatalf("gVolContrat = %d itens", len(n.GVolContrat))
	}
	if n.GVolContrat[0].NContrat != 1 || n.GVolContrat[1].NContrat != 2 {
		t.Errorf("@nContrat nao foi lido do atributo: %+v", n.GVolContrat)
	}
	if n.GVolContrat[0].QUnidContrat != 1500.123456 {
		t.Errorf("qUnidContrat = %v, esperado 6 casas", n.GVolContrat[0].QUnidContrat)
	}
	if n.GVolContrat[0].TpVolContrat != VcVolumeContratado {
		t.Errorf("tpVolContrat = %v", n.GVolContrat[0].TpVolContrat)
	}

	if len(n.GMed) != 1 {
		t.Fatalf("gMed = %d itens", len(n.GMed))
	}
	m := n.GMed[0]
	if m.NMed != 1 || m.IDEqp != "MED-0001" {
		t.Errorf("gMed = %+v", m)
	}
	if m.VMedAnt != 1000.1234 || m.VMedAtu != 1250.5678 {
		t.Errorf("vMedAnt/vMedAtu = %v/%v", m.VMedAnt, m.VMedAtu)
	}
	if m.TpEqp != TeMedidor || m.TpMedidor != TmDiafragma {
		t.Errorf("tpEqp/tpMedidor = %v/%v", m.TpEqp, m.TpMedidor)
	}
	if m.DMedAnt.Day() != 15 || m.DMedAnt.Month() != time.February {
		t.Errorf("dMedAnt = %v", m.DMedAnt)
	}
}

func TestLerXML_DetEAtributosDoItem(t *testing.T) {
	n := lerCompleta(t)

	if len(n.Det) != 2 {
		t.Fatalf("det = %d itens", len(n.Det))
	}
	if n.Det[0].NItem != 1 || n.Det[1].NItem != 2 {
		t.Errorf("@nItem = %d/%d", n.Det[0].NItem, n.Det[1].NItem)
	}
	if n.Det[0].ChNFGasAnt != "" || n.Det[0].NItemAnt != 0 {
		t.Errorf("item 1 nao deveria ter referencia anterior: %+v", n.Det[0])
	}
	if n.Det[1].ChNFGasAnt == "" || n.Det[1].NItemAnt != 1 {
		t.Errorf("@chNFGasAnt/@nItemAnt do item 2 = %q/%d", n.Det[1].ChNFGasAnt, n.Det[1].NItemAnt)
	}
}

func TestLerXML_ProdPrecisaoETipos(t *testing.T) {
	n := lerCompleta(t)
	p := n.Det[0].GNormal.Prod

	if p.IndOrigemQtd != IoMedido {
		t.Errorf("indOrigemQtd = %v", p.IndOrigemQtd)
	}
	if p.CProd != "GAS001" || p.XProd != "Gas Natural Canalizado" {
		t.Errorf("cProd/xProd = %q/%q", p.CProd, p.XProd)
	}
	// cClass e string: o XSD exige [0-9]{7} e o zero a esquerda e
	// significativo. O ACBr declara Integer e perderia o zero.
	if p.CClass != "0100101" {
		t.Errorf("cClass = %q, esperado 0100101 com o zero a esquerda", p.CClass)
	}
	if p.CFOP != 5253 || p.UMed != Umim3 {
		t.Errorf("CFOP/uMed = %d/%v", p.CFOP, p.UMed)
	}
	// qFaturada e decimal: o ACBr declara Integer e arredondaria para 250.
	if p.QFaturada != 250.4444 {
		t.Errorf("qFaturada = %v, esperado 250.4444 com as 4 casas", p.QFaturada)
	}
	// vItem e vProd sao De10 em prod.
	if p.VItem != 3.1234567891 {
		t.Errorf("vItem = %v, esperado 10 casas", p.VItem)
	}
	if p.VProd != 782.1234567891 {
		t.Errorf("vProd = %v, esperado 10 casas", p.VProd)
	}
	if p.FatorPCS != 1.01 || p.FatorPTZ != 0.99 {
		t.Errorf("fatores = %v/%v", p.FatorPCS, p.FatorPTZ)
	}
	if p.IndDevolucao != pcn.TieNao {
		t.Errorf("indDevolucao = %v", p.IndDevolucao)
	}
	if p.GPagAntecipado.NItemPagAnt != 1 || p.GPagAntecipado.ChDFePagAnt == "" {
		t.Errorf("gPagAntecipado = %+v", p.GPagAntecipado)
	}
}

func TestLerXML_GMedicao(t *testing.T) {
	n := lerCompleta(t)
	g := n.Det[0].GNormal.Prod.GMedicao

	if g.NMed != 1 || g.NContrat != 1 {
		t.Errorf("gMedicao = %+v", g)
	}
	if g.GMedida.UMed != Umm3 || g.GMedida.VMed != 250.4444 {
		t.Errorf("gMedida = %+v", g.GMedida)
	}
	if g.TpMotNaoLeitura != TmDistribuidora {
		t.Errorf("tpMotNaoLeitura = %v", g.TpMotNaoLeitura)
	}
	if g.XMotNaoLeitura != "Medidor inacessivel" {
		t.Errorf("xMotNaoLeitura = %q", g.XMotNaoLeitura)
	}
}

func TestLerXML_GTarif(t *testing.T) {
	n := lerCompleta(t)
	lista := n.Det[0].GNormal.GTarif

	if len(lista) != 2 {
		t.Fatalf("gTarif = %d itens", len(lista))
	}
	if lista[0].NAto != "ATO-123" || lista[0].AnoAto != 2026 {
		t.Errorf("gTarif[0] = %+v", lista[0])
	}
	// vTarifAplic e De8.
	if lista[0].VTarifAplic != 3.12345678 {
		t.Errorf("vTarifAplic = %v, esperado 8 casas", lista[0].VTarifAplic)
	}
	if lista[0].TpFaixaCons != TfFixa || lista[1].TpFaixaCons != TfMedia {
		t.Errorf("tpFaixaCons = %v/%v", lista[0].TpFaixaCons, lista[1].TpFaixaCons)
	}
	// dFimTarif ausente no segundo item.
	if !lista[1].DFimTarif.IsZero() {
		t.Errorf("dFimTarif ausente deveria ser zero, veio %v", lista[1].DFimTarif)
	}
}

func TestLerXML_GTarifNaoAcumulaEmReleitura(t *testing.T) {
	// O ACBr NAO limpa gTarif antes do laco. Aqui o slice e reinicializado:
	// ler duas vezes o mesmo XML nao duplica as faixas.
	dados := carregarFixture(t, "nfgas_completa.xml")

	n1, err := LerBytes(dados)
	if err != nil {
		t.Fatal(err)
	}
	n2, err := LerBytes(dados)
	if err != nil {
		t.Fatal(err)
	}
	if len(n1.Det[0].GNormal.GTarif) != len(n2.Det[0].GNormal.GTarif) {
		t.Errorf("releitura mudou a quantidade de gTarif: %d vs %d",
			len(n1.Det[0].GNormal.GTarif), len(n2.Det[0].GNormal.GTarif))
	}
}

// ---------------------------------------------------------------------------
// ICMS achatado
// ---------------------------------------------------------------------------

func TestLerXML_ICMSAchatado(t *testing.T) {
	n := lerCompleta(t)
	icms := n.Det[0].GNormal.Imposto.ICMS

	if icms.CST != pcn.CST20 {
		t.Errorf("CST = %v", icms.CST)
	}
	if icms.ModBC != DbiValorOperacao {
		t.Errorf("modBC = %v", icms.ModBC)
	}
	if icms.VBC != 521.41 || icms.PICMS != 18 || icms.VICMS != 93.85 {
		t.Errorf("ICMS = %+v", icms)
	}
	// pFCP e De4; os demais percentuais sao De2.
	if icms.PFCP != 2.0 || icms.VFCP != 10.43 {
		t.Errorf("pFCP/vFCP = %v/%v", icms.PFCP, icms.VFCP)
	}
	if icms.CBenef != "SP000001" {
		t.Errorf("cBenef = %q", icms.CBenef)
	}
	if icms.IndDeduzDeson != pcn.TieSim {
		t.Errorf("indDeduzDeson = %v", icms.IndDeduzDeson)
	}
	// motDesICMS e lido pelo CODIGO. O ACBr le como inteiro e trataria "9"
	// como ordinal 9, que e mdiDeficienteCondutor.
	if icms.MotDesICMS != MdiOutros {
		t.Errorf("motDesICMS = %v (%q), esperado MdiOutros", icms.MotDesICMS, icms.MotDesICMS.String())
	}
	// indSemCST vem do no imposto, nao do ICMSxx.
	if icms.IndSemCST != pcn.TieNao {
		t.Errorf("indSemCST = %v", icms.IndSemCST)
	}
}

func TestLerXML_ICMSOrdemDeBusca(t *testing.T) {
	// A ordem e a do .pas: ICMS00, 10, 20, 40, 41, 51, 60, 70, 90. Com mais
	// de um grupo presente, vence o PRIMEIRO DA ORDEM, nao o do documento.
	xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><det nItem="1"><gNormal><imposto>
		<ICMS90><CST>90</CST><vBC>90.00</vBC></ICMS90>
		<ICMS10><CST>10</CST><vBC>10.00</vBC></ICMS10>
	</imposto></gNormal></det></infNFGas></NFGas>`

	n, err := LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	icms := n.Det[0].GNormal.Imposto.ICMS
	if icms.VBC != 10 {
		t.Errorf("vBC = %v, esperado 10 (ICMS10 vem antes de ICMS90 na ordem de busca)", icms.VBC)
	}
}

func TestLerXML_ICMSTodosOsGruposAchatam(t *testing.T) {
	for _, grupo := range []string{"ICMS00", "ICMS10", "ICMS20", "ICMS40", "ICMS41", "ICMS51", "ICMS60", "ICMS70", "ICMS90"} {
		t.Run(grupo, func(t *testing.T) {
			xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><det nItem="1"><gNormal><imposto>` +
				`<` + grupo + `><CST>00</CST><vBC>123.45</vBC><vICMS>22.22</vICMS></` + grupo + `>` +
				`</imposto></gNormal></det></infNFGas></NFGas>`

			n, err := LerXMLString(xml)
			if err != nil {
				t.Fatal(err)
			}
			icms := n.Det[0].GNormal.Imposto.ICMS
			if icms.VBC != 123.45 || icms.VICMS != 22.22 {
				t.Errorf("%s nao foi achatado: %+v", grupo, icms)
			}
		})
	}
}

func TestLerXML_ICMS41EhProcuradoMesmoSemExistirNoXSD(t *testing.T) {
	// O ACBr procura ICMS41, que nao existe no XSD da NFGas. Replicado: se
	// aparecer num XML de terceiro, e lido.
	if gruposICMS[4] != "ICMS41" {
		t.Errorf("a 5a posicao da ordem de busca deveria ser ICMS41, e %q", gruposICMS[4])
	}
	if len(gruposICMS) != 9 {
		t.Errorf("a ordem de busca tem %d grupos, esperado 9", len(gruposICMS))
	}
}

func TestLerXML_SemGrupoICMSNaoAlteraNada(t *testing.T) {
	// Sem nenhum grupo ICMSxx, o Ler_ICMS do ACBr sai no Exit ANTES de ler
	// indSemCST (XmlReader.pas:619-620,656) -- ou seja, o indSemCST do XML e
	// DESCARTADO nesse caso, tanto no ICMS quanto no Imposto (que so recebe
	// o campo via .ini). Replicado.
	//
	// O valor "0" (TiNao) e distinguivel do zero do enum (TiSim): se algum
	// caminho passasse a ler a tag, este teste quebraria.
	xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><det nItem="1"><gNormal><imposto>
		<orig>0</orig><indSemCST>0</indSemCST>
	</imposto></gNormal></det></infNFGas></NFGas>`

	n, err := LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	icms := n.Det[0].GNormal.Imposto.ICMS
	if icms != (ICMS{}) {
		t.Errorf("sem ICMSxx a struct deveria ficar zerada, veio %+v", icms)
	}
	if n.Det[0].GNormal.Imposto.IndSemCST != pcn.TieNenhum {
		t.Errorf("indSemCST deveria ser descartado (TieNenhum), veio %v",
			n.Det[0].GNormal.Imposto.IndSemCST)
	}

	// COM grupo ICMSxx, o mesmo indSemCST do no imposto E lido para o ICMS.
	xmlCom := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><det nItem="1"><gNormal><imposto>
		<ICMS00><CST>00</CST></ICMS00><indSemCST>0</indSemCST>
	</imposto></gNormal></det></infNFGas></NFGas>`
	n2, err := LerXMLString(xmlCom)
	if err != nil {
		t.Fatal(err)
	}
	if n2.Det[0].GNormal.Imposto.ICMS.IndSemCST != pcn.TieNao {
		t.Errorf("com ICMSxx presente, indSemCST deveria ser lido: %v",
			n2.Det[0].GNormal.Imposto.ICMS.IndSemCST)
	}
}

func TestLerXML_PISCOFINSRetTribTxReg(t *testing.T) {
	n := lerCompleta(t)
	imp := n.Det[0].GNormal.Imposto

	if imp.Orig != pcn.OeNacional {
		t.Errorf("orig = %v", imp.Orig)
	}
	if imp.PIS.CST != pcn.Pis01 || imp.PIS.PPIS != 1.65 || imp.PIS.VPIS != 12.90 {
		t.Errorf("PIS = %+v", imp.PIS)
	}
	if imp.COFINS.CST != pcn.Cof01 || imp.COFINS.PCOFINS != 7.60 {
		t.Errorf("COFINS = %+v", imp.COFINS)
	}
	// A tag e vRetCofins, nao vRetCOFINS.
	if imp.RetTrib.VRetCOFINS != 2.22 {
		t.Errorf("vRetCofins = %v", imp.RetTrib.VRetCOFINS)
	}
	if imp.RetTrib.VRetPIS != 1.11 || imp.RetTrib.VRetCSLL != 3.33 || imp.RetTrib.VIRRF != 4.44 {
		t.Errorf("retTrib = %+v", imp.RetTrib)
	}
	// vBCIRRF existe na struct e NAO e lido pelo ACBr.
	if imp.RetTrib.VBCIRRF != 0 {
		t.Errorf("vBCIRRF = %v; o ACBr nao le esse campo", imp.RetTrib.VBCIRRF)
	}
	if imp.TxReg.PTaxa != 0.5 || imp.TxReg.VTaxa != 3.91 {
		t.Errorf("TxReg = %+v", imp.TxReg)
	}
}

func TestLerXML_IBSCBSDoItem(t *testing.T) {
	n := lerCompleta(t)
	ibs := n.Det[0].GNormal.Imposto.IBSCBS

	if ibs.CST != rtc.CST200 || ibs.CClassTrib != "000001" {
		t.Errorf("IBSCBS = CST %v, cClassTrib %q", ibs.CST, ibs.CClassTrib)
	}
	if ibs.GIBSCBS.VBC != 782.12 || ibs.GIBSCBS.VIBS != 66.48 {
		t.Errorf("gIBSCBS = %+v", ibs.GIBSCBS)
	}
	if ibs.GIBSCBS.GIBSUF.PIBSUF != 0.1 || ibs.GIBSCBS.GIBSUF.VIBSUF != 7.82 {
		t.Errorf("gIBSUF = %+v", ibs.GIBSCBS.GIBSUF)
	}
	if ibs.GIBSCBS.GIBSUF.GDif.PDif != 1.0 || ibs.GIBSCBS.GIBSUF.GRed.PAliqEfet != 4.0 {
		t.Errorf("subgrupos de gIBSUF = %+v", ibs.GIBSCBS.GIBSUF)
	}
	if ibs.GIBSCBS.GCBS.GALCZFMCBS.NProcSuframa != "SUF-001" {
		t.Errorf("gALCZFMCBS = %+v", ibs.GIBSCBS.GCBS.GALCZFMCBS)
	}
	if ibs.GEstornoCred.VIBSEstCred != 5.0 || ibs.GEstornoCred.VCBSEstCred != 6.0 {
		t.Errorf("gEstornoCred = %+v", ibs.GEstornoCred)
	}
}

func TestLerXML_GProcRefPrecisaoDiferenteDeProd(t *testing.T) {
	n := lerCompleta(t)
	g := n.Det[0].GNormal.GProcRef

	// vItem e vProd sao De8 aqui e De10 em prod.
	if g.VItem != 1.23456789 {
		t.Errorf("gProcRef/vItem = %v, esperado 8 casas", g.VItem)
	}
	if g.VProd != 12.96296285 {
		t.Errorf("gProcRef/vProd = %v, esperado 8 casas", g.VProd)
	}
	if g.QFaturada != 10.5 {
		t.Errorf("gProcRef/qFaturada = %v", g.QFaturada)
	}
	if g.IndDevolucao != pcn.TieSim {
		t.Errorf("indDevolucao = %v", g.IndDevolucao)
	}
	if len(g.GProc) != 2 {
		t.Fatalf("gProc = %d itens", len(g.GProc))
	}
	if g.GProc[0].TpProc != TpJusticaEstadual || g.GProc[1].TpProc != TpProcon {
		t.Errorf("tpProc = %v/%v", g.GProc[0].TpProc, g.GProc[1].TpProc)
	}
}

func TestLerXML_TotalAchatado(t *testing.T) {
	n := lerCompleta(t)
	tot := n.Total

	if tot.VProd != 832.12 || tot.VNF != 948.72 || tot.VTotDFe != 948.72 {
		t.Errorf("total = %+v", tot)
	}
	// Campos de ICMSTot achatados.
	if tot.VBC != 571.41 || tot.VICMS != 102.85 || tot.VICMSDeson != 10 || tot.VFCP != 10.43 {
		t.Errorf("ICMSTot achatado = vBC %v, vICMS %v", tot.VBC, tot.VICMS)
	}
	// Campos de vRetTribTot achatados; a tag e vRetCofins.
	if tot.VRetPIS != 1.11 || tot.VRetCOFINS != 2.22 || tot.VRetCSLL != 3.33 || tot.VIRRF != 4.44 {
		t.Errorf("vRetTribTot achatado = %+v", tot)
	}
	if tot.VCOFINS != 59.44 || tot.VPIS != 12.90 || tot.VTxReg != 3.91 {
		t.Errorf("totais de contribuicao = %+v", tot)
	}
}

func TestLerXML_IBSCBSTot(t *testing.T) {
	n := lerCompleta(t)
	tot := n.Total.IBSCBSTot

	if tot.VBCIBSCBS != 832.12 {
		t.Errorf("vBCIBSCBS = %v", tot.VBCIBSCBS)
	}
	if tot.GIBS.GIBSUFTot.VIBSUF != 7.82 || tot.GIBS.GIBSMunTot.VIBSMun != 15.64 {
		t.Errorf("gIBS do total = %+v", tot.GIBS)
	}
	if tot.GCBS.VCBS != 70.39 || tot.GCBS.VCredPresCondSus != 8 {
		t.Errorf("gCBS do total = %+v", tot.GCBS)
	}
	if tot.GMono.VIBSMono != 9 || tot.GMono.VCBSMonoRet != 14 {
		t.Errorf("gMono = %+v", tot.GMono)
	}
	if tot.GEstornoCred.VIBSEstCred != 15 {
		t.Errorf("gEstornoCred do total = %+v", tot.GEstornoCred)
	}
}

func TestLerXML_PgtoVinc(t *testing.T) {
	n := lerCompleta(t)
	if len(n.PgtoVinc.Pgto) != 2 {
		t.Fatalf("pgto = %d itens", len(n.PgtoVinc.Pgto))
	}
	if n.PgtoVinc.Pgto[0].NPag != 1 || n.PgtoVinc.Pgto[0].IDTransacao != "TX-0001" {
		t.Errorf("pgto[0] = %+v", n.PgtoVinc.Pgto[0])
	}
	if n.PgtoVinc.Pgto[1].TpMeioPgto != "17" {
		t.Errorf("pgto[1]/tpMeioPgto = %q", n.PgtoVinc.Pgto[1].TpMeioPgto)
	}
}

func TestLerXML_GFatEEnderCorresp(t *testing.T) {
	n := lerCompleta(t)
	g := n.GFat

	esperado := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !g.CompetFat.Equal(esperado) {
		t.Errorf("CompetFat = %v, esperado %v", g.CompetFat, esperado)
	}
	if g.DVencFat.Day() != 10 || g.DVencFat.Month() != time.April {
		t.Errorf("dVencFat = %v", g.DVencFat)
	}
	if g.NFat != "FAT-2026-03-0001" || g.CodBanco != "001" {
		t.Errorf("gFat = %+v", g)
	}
	if g.GPIX.URLQRCodePIX != "https://pix.exemplo.com.br/qr/0001" {
		t.Errorf("gPIX = %+v", g.GPIX)
	}
	if g.EnderCorresp.XLgr != "Rua da Correspondencia" {
		t.Errorf("enderCorresp = %+v", g.EnderCorresp)
	}
	// enderCorresp NAO le cPais/xPais, mesmo presentes no XML -- e assim no
	// ACBr, e o tipo do XSD nem tem esses campos.
	if g.EnderCorresp.CPais != 0 || g.EnderCorresp.XPais != "" {
		t.Errorf("enderCorresp nao deveria ler cPais/xPais: %d/%q",
			g.EnderCorresp.CPais, g.EnderCorresp.XPais)
	}
}

func TestLerXML_GAgenciaEHistoricoDeConsumo(t *testing.T) {
	n := lerCompleta(t)
	g := n.GAgencia

	if g.NomeAgenciaAtend != "Agencia Central" || g.InfAdReg != "Regulado pela ARSESP" {
		t.Errorf("gAgencia = %+v", g)
	}
	if len(g.GHistCons) != 1 {
		t.Fatalf("gHistCons = %d itens", len(g.GHistCons))
	}
	h := g.GHistCons[0]
	if h.MedMensal != 245 {
		t.Errorf("medMensal = %v", h.MedMensal)
	}
	if len(h.GCons) != 2 {
		t.Fatalf("gCons = %d itens", len(h.GCons))
	}
	if h.GCons[0].QtdDias != 28 || h.GCons[0].Consumo != 250 || h.GCons[0].VFat != 900 {
		t.Errorf("gCons[0] = %+v", h.GCons[0])
	}
	if h.GCons[0].CompetFat.Month() != time.February {
		t.Errorf("gCons[0]/CompetFat = %v", h.GCons[0].CompetFat)
	}
	if h.GCons[1].UMed != Umm3 {
		t.Errorf("gCons[1]/uMed = %v", h.GCons[1].UMed)
	}
}

func TestLerXML_AutXMLInfAdicRespTec(t *testing.T) {
	n := lerCompleta(t)

	if len(n.AutXML) != 2 {
		t.Fatalf("autXML = %d itens", len(n.AutXML))
	}
	if n.AutXML[0].CNPJCPF != "99888777000166" {
		t.Errorf("autXML[0] = %+v", n.AutXML[0])
	}
	if n.AutXML[1].CNPJCPF != "52998224725" {
		t.Errorf("autXML[1] deveria cair para CPF: %+v", n.AutXML[1])
	}

	if n.InfAdic.InfAdFisco != "Documento emitido para teste" {
		t.Errorf("infAdFisco = %q", n.InfAdic.InfAdFisco)
	}
	// O leiaute admite 5 infCpl, o ACBr le so a primeira. Replicado.
	if len(n.InfAdic.InfCpl) != 1 {
		t.Errorf("infCpl = %d ocorrencias, esperado 1 (o ACBr le so a primeira)", len(n.InfAdic.InfCpl))
	}
	if len(n.InfAdic.InfCpl) > 0 && n.InfAdic.InfCpl[0] != "Informacao complementar 1" {
		t.Errorf("infCpl[0] = %q", n.InfAdic.InfCpl[0])
	}

	// A TAG e gRespTec; o campo e infRespTec.
	if n.InfRespTec.CNPJ != "99888777000166" || n.InfRespTec.IDCSRT != 1 {
		t.Errorf("infRespTec (tag gRespTec) = %+v", n.InfRespTec)
	}
}

func TestLerXML_QrCodeRemoveCDATAUmaVez(t *testing.T) {
	n := lerCompleta(t)
	qr := n.InfNFGasSupl.QrCodNFGas

	if strings.Contains(qr, "CDATA") || strings.Contains(qr, "]]>") {
		t.Errorf("CDATA nao foi removido: %q", qr)
	}
	if !strings.HasPrefix(qr, "https://") {
		t.Errorf("qrCodNFGas = %q", qr)
	}

	// O ACBr usa StringReplace SEM rfReplaceAll: apenas a primeira marcacao
	// de cada tipo sai. Deliberado.
	xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"></infNFGas>` +
		`<infNFGasSupl><qrCodNFGas>&lt;![CDATA[a]]&gt;&lt;![CDATA[b]]&gt;</qrCodNFGas></infNFGasSupl></NFGas>`
	n2, err := LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if n2.InfNFGasSupl.QrCodNFGas != "a<![CDATA[b]]>" {
		t.Errorf("qrCodNFGas = %q, esperado a<![CDATA[b]]> (so a primeira ocorrencia sai)",
			n2.InfNFGasSupl.QrCodNFGas)
	}
}

func TestLerXML_Signature(t *testing.T) {
	n := lerCompleta(t)
	if n.Signature.Vazia() {
		t.Fatal("assinatura nao foi lida")
	}
	if n.Signature.DigestValue != "RGlnZXN0VmFsdWVUZXN0ZQ==" {
		t.Errorf("DigestValue = %q", n.Signature.DigestValue)
	}
	if n.Signature.X509Certificate != "Q2VydGlmaWNhZG9UZXN0ZQ==" {
		t.Errorf("X509Certificate = %q", n.Signature.X509Certificate)
	}
	if n.Signature.URI != "#NFGas"+chaveTeste {
		t.Errorf("URI = %q", n.Signature.URI)
	}
}

// ---------------------------------------------------------------------------
// nfgasProc e protocolo
// ---------------------------------------------------------------------------

func TestLerXML_NFGasProcComProtocolo(t *testing.T) {
	n, err := LerBytes(carregarFixture(t, "nfgas_proc.xml"))
	if err != nil {
		t.Fatalf("LerBytes: %v", err)
	}

	if n.ChaveAcesso() != chaveTeste {
		t.Errorf("chave = %q", n.ChaveAcesso())
	}
	p := n.ProcNFGas
	if p.CStat != 100 || p.NProt != "135260000000001" {
		t.Errorf("protocolo = %+v", p)
	}
	if p.ChDFe != chaveTeste {
		t.Errorf("chDFe = %q", p.ChDFe)
	}
	if p.VerAplic != "SP_NFGAS_1.0.0" || p.XMotivo != "Autorizado o uso da NFGas" {
		t.Errorf("protocolo = %+v", p)
	}
	if p.CMsg != 1 || p.XMsg != "Observacao do fisco" {
		t.Errorf("cMsg/xMsg = %d/%q", p.CMsg, p.XMsg)
	}
	if p.DhRecbto.Hour() != 11 {
		t.Errorf("dhRecbto = %v", p.DhRecbto)
	}
	// infProt/@Id NAO e lido pelo ACBr.
	if p.ID != "" {
		t.Errorf("infProt/@Id = %q; o ACBr nao le esse atributo", p.ID)
	}

	if !n.Confirmada() || !n.Processada() || n.Cancelada() {
		t.Errorf("situacao: confirmada=%v processada=%v cancelada=%v",
			n.Confirmada(), n.Processada(), n.Cancelada())
	}
}

func TestCStat_Classificacao(t *testing.T) {
	// Tabela de TACBrNFGas.CstatConfirmada/Processado/Cancelada
	// (ACBrNFGas.pas:230-256): Confirmada == Processado == {100,150};
	// Cancelada inclui o 135. Os codigos 110/301/302 sao do NFe e NAO
	// existem na tabela da NFGas.
	casos := []struct {
		cStat                             int
		confirmada, processada, cancelada bool
	}{
		{100, true, true, false},
		{150, true, true, false},
		{110, false, false, false},
		{101, false, false, true},
		{135, false, false, true},
		{151, false, false, true},
		{155, false, false, true},
		{301, false, false, false},
		{999, false, false, false},
	}
	for _, c := range casos {
		if got := CStatConfirmada(c.cStat); got != c.confirmada {
			t.Errorf("CStatConfirmada(%d) = %v", c.cStat, got)
		}
		if got := CStatProcessado(c.cStat); got != c.processada {
			t.Errorf("CStatProcessado(%d) = %v", c.cStat, got)
		}
		if got := CStatCancelada(c.cStat); got != c.cancelada {
			t.Errorf("CStatCancelada(%d) = %v", c.cStat, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Leitura em lote
// ---------------------------------------------------------------------------

func TestLerLote_Misto(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "lote_misto.xml"))
	if err != nil {
		t.Fatalf("LerLoteBytes: %v", err)
	}
	if len(notas) != 3 {
		t.Fatalf("len(notas) = %d, esperado 3", len(notas))
	}

	if notas[0].NFGas.Ide.NNF != 1 || notas[1].NFGas.Ide.NNF != 2 || notas[2].NFGas.Ide.NNF != 3 {
		t.Errorf("ordem das notas: %d %d %d",
			notas[0].NFGas.Ide.NNF, notas[1].NFGas.Ide.NNF, notas[2].NFGas.Ide.NNF)
	}
	// So a do meio veio dentro de nfgasProc.
	if notas[0].CStat() != 0 || notas[1].CStat() != 100 || notas[2].CStat() != 0 {
		t.Errorf("cStat = %d %d %d", notas[0].CStat(), notas[1].CStat(), notas[2].CStat())
	}
	if notas[1].Situacao() != "confirmada" {
		t.Errorf("situacao da nota 2 = %q", notas[1].Situacao())
	}
	if notas[0].Situacao() != "sem protocolo" {
		t.Errorf("situacao da nota 1 = %q", notas[0].Situacao())
	}
}

func TestLerLote_NaoContaNFGasDentroDeNFGasProcDuasVezes(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "nfgas_proc.xml"))
	if err != nil {
		t.Fatalf("LerLoteBytes: %v", err)
	}
	if len(notas) != 1 {
		t.Fatalf("len(notas) = %d, esperado 1 -- a NFGas interna nao pode contar de novo", len(notas))
	}
	if notas[0].CStat() != 100 {
		t.Errorf("a nota deveria trazer o protocolo: cStat = %d", notas[0].CStat())
	}
}

func TestLerLote_PreservaXMLOriginalReparseavel(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "lote_misto.xml"))
	if err != nil {
		t.Fatal(err)
	}
	for i, nota := range notas {
		if nota.XMLOriginal == "" {
			t.Fatalf("nota %d sem XMLOriginal", i)
		}
		relido, err := LerXMLString(nota.XMLOriginal)
		if err != nil {
			t.Errorf("nota %d: XMLOriginal nao e reparseavel: %v", i, err)
			continue
		}
		if relido.ChaveAcesso() != nota.ChaveAcesso() {
			t.Errorf("nota %d: chave mudou no reparse", i)
		}
	}
}

func TestLerLote_NotaTortaNaoDerrubaOLote(t *testing.T) {
	// Uma nota sem o atributo Id no meio de tres.
	lote := `<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas"><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><ide><nNF>1</nNF></ide></infNFGas></NFGas>` +
		`<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas"><infNFGas versao="1.00"><ide><nNF>2</nNF></ide></infNFGas></NFGas>` +
		`<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas"><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><ide><nNF>3</nNF></ide></infNFGas></NFGas>`

	notas, err := LerLoteString(lote)
	if len(notas) != 2 {
		t.Fatalf("len(notas) = %d, esperado 2 lidas apesar da falha", len(notas))
	}
	if err == nil {
		t.Fatal("a falha da nota 2 deveria ser reportada")
	}

	var lotErr *ErrosLote
	if !errors.As(err, &lotErr) {
		t.Fatalf("erro = %T, esperado *ErrosLote", err)
	}
	if len(lotErr.Erros) != 1 {
		t.Fatalf("erros = %d, esperado 1", len(lotErr.Erros))
	}
	if lotErr.Erros[0].Indice != 1 {
		t.Errorf("indice do erro = %d, esperado 1", lotErr.Erros[0].Indice)
	}
	if !errors.Is(err, ErrAtributoIDAusente) {
		t.Errorf("errors.Is deveria alcancar ErrAtributoIDAusente, erro = %v", err)
	}
}

func TestLerLote_Vazio(t *testing.T) {
	if _, err := LerLoteString(""); !errors.Is(err, ErrXMLVazio) {
		t.Errorf("lote vazio: erro = %v", err)
	}
	if _, err := LerLoteString(`<outro><coisa/></outro>`); !errors.Is(err, ErrLoteVazio) {
		t.Errorf("lote sem NFGas: erro = %v", err)
	}
}

func TestLerLoteArquivo_EDiretorio(t *testing.T) {
	dir := t.TempDir()
	dados := carregarFixture(t, "nfgas_completa.xml")

	for _, nome := range []string{"a.xml", "b.xml", "ignorar.txt"} {
		if err := os.WriteFile(filepath.Join(dir, nome), dados, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	notas, err := LerLoteArquivo(filepath.Join(dir, "a.xml"))
	if err != nil {
		t.Fatalf("LerLoteArquivo: %v", err)
	}
	if len(notas) != 1 || notas[0].NomeArq == "" {
		t.Errorf("LerLoteArquivo nao preencheu NomeArq: %+v", notas)
	}

	todas, err := LerLoteDiretorio(dir)
	if err != nil {
		t.Fatalf("LerLoteDiretorio: %v", err)
	}
	if len(todas) != 2 {
		t.Errorf("LerLoteDiretorio = %d notas, esperado 2 (o .txt e ignorado)", len(todas))
	}
}

func TestNotaFiscal_Conveniencias(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "nfgas_proc.xml"))
	if err != nil {
		t.Fatal(err)
	}
	nota := notas[0]

	if nota.CalcularNomeArquivo() != chaveTeste+"-NFGas.xml" {
		t.Errorf("CalcularNomeArquivo = %q", nota.CalcularNomeArquivo())
	}
	if nota.Msg() != "Autorizado o uso da NFGas" {
		t.Errorf("Msg = %q", nota.Msg())
	}
	if nota.Assinado() {
		t.Error("esta fixture nao tem assinatura")
	}
}

func TestLerXML_ToleraBOMEDeclaracao(t *testing.T) {
	dados := carregarFixture(t, "nfgas_completa.xml")
	comBOM := append([]byte("\xef\xbb\xbf"), dados...)

	n, err := LerBytes(comBOM)
	if err != nil {
		t.Fatalf("BOM deveria ser tolerado: %v", err)
	}
	if n.ChaveAcesso() != chaveTeste {
		t.Errorf("chave = %q", n.ChaveAcesso())
	}
}

// ---------------------------------------------------------------------------
// Erros de leitura
// ---------------------------------------------------------------------------

func TestLerXML_Erros(t *testing.T) {
	casos := []struct {
		nome string
		xml  string
		err  error
	}{
		{"vazio", "", pcn.ErrXMLVazio},
		{"malformado", "<NFGas><infNFGas>", pcn.ErrXMLInvalido},
		{"sem infNFGas", `<NFGas><outro/></NFGas>`, ErrXMLIncorreto},
		{"sem Id", `<NFGas><infNFGas versao="1.00"/></NFGas>`, ErrAtributoIDAusente},
		{"sem versao", `<NFGas><infNFGas Id="NFGas123"/></NFGas>`, ErrAtributoVersaoAusente},
		{"proc sem NFGas", `<nfgasProc><protNFGas/></nfgasProc>`, ErrNFGasNaoEncontrada},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, err := LerXMLString(c.xml)
			if !errors.Is(err, c.err) {
				t.Errorf("erro = %v, esperado %v", err, c.err)
			}
		})
	}
}

func TestLerXML_NaoEntraEmPanico(t *testing.T) {
	// Um leitor que estoura o processo e inaceitavel em importacao de lote.
	entradas := []string{
		"", "   ", "<", "<a", "<NFGas/>", "<NFGas></NFGas>",
		`<NFGas><infNFGas versao="x" Id=""/></NFGas>`,
		`<NFGas><infNFGas versao="1.00" Id="NFGas1"><det nItem="x"><gNormal><imposto><ICMS00><CST>zz</CST><vBC>abc</vBC></ICMS00></imposto></gNormal></det></infNFGas></NFGas>`,
	}
	for _, e := range entradas {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panico com entrada %q: %v", e, r)
				}
			}()
			_, _ = LerXMLString(e)
			_, _ = LerLoteString(e)
			_, _ = LerEventoString(e)
			_, _ = LerRetConsSitString(e)
		}()
	}
}

func TestLerXML_ValorInvalidoNaoAborta(t *testing.T) {
	// Enum invalido nao derruba a leitura: o campo fica no valor zero e o
	// resto do documento continua sendo lido. E o comportamento util para
	// importacao -- o ACBr levantaria excecao e perderia a nota inteira.
	xml := `<NFGas><infNFGas versao="1.00" Id="NFGas` + chaveTeste + `"><ide>
		<cUF>35</cUF><nNF>42</nNF><tpFat>ZZZ</tpFat><finNFGas>9</finNFGas>
	</ide></infNFGas></NFGas>`

	n, err := LerXMLString(xml)
	if err != nil {
		t.Fatalf("enum invalido nao deveria abortar: %v", err)
	}
	if n.Ide.NNF != 42 {
		t.Errorf("os demais campos deveriam ser lidos: nNF = %d", n.Ide.NNF)
	}
}

// ---------------------------------------------------------------------------
// Eventos
// ---------------------------------------------------------------------------

func TestRetInfEventoRegistrado(t *testing.T) {
	// conjunto de aceite do ACBrNFGasWebServices.pas:1571: [135, 136, 155]
	for _, cStat := range []int{135, 136, 155} {
		if !(&RetInfEvento{CStat: cStat}).Registrado() {
			t.Errorf("cStat %d deveria ser evento registrado", cStat)
		}
	}
	for _, cStat := range []int{0, 100, 128, 573} {
		if (&RetInfEvento{CStat: cStat}).Registrado() {
			t.Errorf("cStat %d nao deveria ser evento registrado", cStat)
		}
	}
}

func TestLerEvento_ProcEventoCancelamento(t *testing.T) {
	ev, err := LerEventoBytes(carregarFixture(t, "proc_evento_cancelamento.xml"))
	if err != nil {
		t.Fatalf("LerEventoBytes: %v", err)
	}

	if ev.Versao != "1.00" {
		t.Errorf("versao = %q", ev.Versao)
	}
	r := ev.RetInfEvento
	if r.CStat != 135 || !r.Registrado() {
		t.Errorf("retorno = cStat %d", r.CStat)
	}
	if r.TpEvento != TeCancelamento || r.NSeqEvento != 1 {
		t.Errorf("tpEvento/nSeqEvento = %v/%d", r.TpEvento, r.NSeqEvento)
	}
	if r.ChNFGas != chaveTeste || r.NProt != "135260000000099" {
		t.Errorf("chNFGas/nProt = %q/%q", r.ChNFGas, r.NProt)
	}
	if r.DhRegEvento.Hour() != 9 || r.DhRegEvento.Minute() != 5 {
		t.Errorf("dhRegEvento = %v", r.DhRegEvento)
	}
	// Acrescimos ao porte -- o ACBr declara e nao le.
	if r.CNPJDest != "99888777000166" || r.EmailDest != "dest@teste.com.br" || r.COrgaoAutor != 35 {
		t.Errorf("campos acrescentados ao porte = %q/%q/%d", r.CNPJDest, r.EmailDest, r.COrgaoAutor)
	}
	if r.XML == "" || !strings.Contains(r.XML, "<infEvento") {
		t.Errorf("XML do infEvento nao foi preservado: %q", r.XML)
	}

	// Parte enviada -- ACRESCIMO: e a unica fonte da justificativa.
	if !ev.TemEvento {
		t.Fatal("a parte enviada deveria ter sido lida")
	}
	if ev.Justificativa() != "Erro na medicao do consumo do periodo faturado" {
		t.Errorf("xJust = %q", ev.Justificativa())
	}
	if ev.Evento.InfEvento.DetEvento.NProt != "135260000000001" {
		t.Errorf("detEvento/nProt = %q", ev.Evento.InfEvento.DetEvento.NProt)
	}
	if ev.Evento.InfEvento.CNPJ != "11222333000181" {
		t.Errorf("infEvento/CNPJ = %q", ev.Evento.InfEvento.CNPJ)
	}
	if ev.Evento.InfEvento.DhEvento.Day() != 16 {
		t.Errorf("dhEvento = %v", ev.Evento.InfEvento.DhEvento)
	}

	if !ev.Cancelamento() {
		t.Error("Cancelamento() deveria ser true")
	}
	if ev.ChaveAcesso() != chaveTeste {
		t.Errorf("ChaveAcesso = %q", ev.ChaveAcesso())
	}
	if ev.Signature.Vazia() {
		t.Error("a assinatura do retorno deveria ter sido lida")
	}
}

func TestLerEvento_SoRetorno(t *testing.T) {
	xml := `<retEventoNFGas versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfgas">
		<infEvento Id="ID1"><tpAmb>1</tpAmb><cStat>136</cStat>
		<chNFGas>` + chaveTeste + `</chNFGas><tpEvento>110111</tpEvento><nSeqEvento>2</nSeqEvento>
		</infEvento></retEventoNFGas>`

	ev, err := LerEventoString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if ev.TemEvento {
		t.Error("nao havia parte enviada neste XML")
	}
	if ev.Justificativa() != "" {
		t.Errorf("sem a parte enviada nao ha justificativa, veio %q", ev.Justificativa())
	}
	if !ev.RetInfEvento.Registrado() {
		t.Errorf("cStat 136 deveria contar como registrado")
	}
	if ev.RetInfEvento.NSeqEvento != 2 {
		t.Errorf("nSeqEvento = %d", ev.RetInfEvento.NSeqEvento)
	}
}

func TestLerEvento_SoEnviado(t *testing.T) {
	xml := `<eventoNFGas versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfgas">
		<infEvento Id="ID2"><cOrgao>35</cOrgao><tpAmb>2</tpAmb>
		<chNFGas>` + chaveTeste + `</chNFGas><tpEvento>110111</tpEvento><nSeqEvento>1</nSeqEvento>
		<detEvento versao="1.00"><descEvento>Cancelamento</descEvento><xJust>Justificativa X</xJust></detEvento>
		</infEvento></eventoNFGas>`

	ev, err := LerEventoString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.TemEvento || ev.Justificativa() != "Justificativa X" {
		t.Errorf("evento enviado = %+v", ev.Evento.InfEvento)
	}
	if ev.ChaveAcesso() != chaveTeste {
		t.Errorf("ChaveAcesso deveria cair para a do enviado: %q", ev.ChaveAcesso())
	}
}

func TestInfEvento_COrgaoEfetivoEDescricao(t *testing.T) {
	// Sem cOrgao, vale os dois primeiros digitos da chave.
	i := InfEvento{ChNFGas: chaveTeste}
	if got := i.COrgaoEfetivo(); got != 35 {
		t.Errorf("COrgaoEfetivo = %d, esperado 35", got)
	}
	i.COrgao = 41
	if got := i.COrgaoEfetivo(); got != 41 {
		t.Errorf("COrgaoEfetivo = %d, esperado 41", got)
	}

	i.TpEvento = TeCancelamento
	if i.DescEvento() != "Cancelamento" {
		t.Errorf("DescEvento = %q", i.DescEvento())
	}
	if DescricaoTipoEvento(TeCancelamento) != "CANCELAMENTO DE NFGas" {
		t.Errorf("DescricaoTipoEvento = %q", DescricaoTipoEvento(TeCancelamento))
	}
	if DescricaoTipoEvento(TeAutorizadoAjuste) != "Nao Definido" {
		t.Errorf("o ACBr so mapeia o cancelamento: %q", DescricaoTipoEvento(TeAutorizadoAjuste))
	}
}

func TestLerEvento_RaizDesconhecida(t *testing.T) {
	if _, err := LerEventoString(`<outraCoisa/>`); !errors.Is(err, ErrXMLIncorreto) {
		t.Errorf("erro = %v, esperado ErrXMLIncorreto", err)
	}
}

// ---------------------------------------------------------------------------
// Retorno de consulta
// ---------------------------------------------------------------------------

func TestLerRetConsSit(t *testing.T) {
	r, err := LerRetConsSitBytes(carregarFixture(t, "ret_cons_sit.xml"))
	if err != nil {
		t.Fatalf("LerRetConsSitBytes: %v", err)
	}

	if r.CStat != 101 || r.XMotivo != "Cancelamento de NFGas homologado" {
		t.Errorf("retorno = cStat %d, %q", r.CStat, r.XMotivo)
	}
	if r.CUF != 35 || r.NRec != "351260000000123" {
		t.Errorf("cUF/nRec = %d/%q", r.CUF, r.NRec)
	}
	if r.ChNFGas != chaveTeste {
		t.Errorf("chNFGas = %q", r.ChNFGas)
	}
	if !r.Cancelada() || r.Autorizada() {
		t.Errorf("situacao: cancelada=%v autorizada=%v", r.Cancelada(), r.Autorizada())
	}

	// cStat 101 esta na lista que traz protocolo.
	if r.ProtNFGas.CStat != 100 || r.ProtNFGas.NProt != "135260000000001" {
		t.Errorf("protNFGas = %+v", r.ProtNFGas)
	}
	if !strings.Contains(r.XMLProtNFGas, "<protNFGas") {
		t.Errorf("XMLProtNFGas = %q", r.XMLProtNFGas)
	}

	if len(r.ProcEventoNFGas) != 1 {
		t.Fatalf("eventos vinculados = %d, esperado 1", len(r.ProcEventoNFGas))
	}
	if r.ProcEventoNFGas[0].RetInfEvento.CStat != 135 {
		t.Errorf("evento vinculado = %+v", r.ProcEventoNFGas[0].RetInfEvento)
	}
}

func TestLerRetConsSit_ProtocoloSoNosCStatPrevistos(t *testing.T) {
	// A lista do ACBr e {100, 101, 104, 150, 151, 155}. Fora dela, o grupo
	// protNFGas e ignorado mesmo estando presente.
	monta := func(cStat int) string {
		return `<retConsSitNFGas versao="1.00"><cStat>` +
			pcn.FormatarInteiroZeros(cStat, 3) +
			`</cStat><protNFGas><infProt><nProt>999</nProt><cStat>100</cStat></infProt></protNFGas></retConsSitNFGas>`
	}

	for _, cStat := range []int{100, 101, 104, 150, 151, 155} {
		r, err := LerRetConsSitString(monta(cStat))
		if err != nil {
			t.Fatal(err)
		}
		if r.ProtNFGas.NProt != "999" {
			t.Errorf("cStat %d deveria ler o protocolo", cStat)
		}
	}
	for _, cStat := range []int{110, 217, 999} {
		r, err := LerRetConsSitString(monta(cStat))
		if err != nil {
			t.Fatal(err)
		}
		if r.ProtNFGas.NProt != "" {
			t.Errorf("cStat %d nao deveria ler o protocolo, veio %q", cStat, r.ProtNFGas.NProt)
		}
	}
}

// ---------------------------------------------------------------------------
// Regras de negocio e chave de acesso
// ---------------------------------------------------------------------------

func TestValidarConcatChave(t *testing.T) {
	n := lerCompleta(t)

	if !ValidarConcatChave(n) {
		t.Fatalf("a chave da fixture deveria bater com os campos; chave = %q", n.ChaveAcesso())
	}

	// Qualquer campo da concatenacao que mude invalida.
	alteracoes := []struct {
		nome string
		muda func(*NFGas)
	}{
		{"cUF", func(x *NFGas) { x.Ide.CUF = 41 }},
		{"ano", func(x *NFGas) { x.Ide.DhEmi = x.Ide.DhEmi.AddDate(1, 0, 0) }},
		{"mes", func(x *NFGas) { x.Ide.DhEmi = x.Ide.DhEmi.AddDate(0, 1, 0) }},
		{"CNPJ", func(x *NFGas) { x.Emit.CNPJ = "99888777000166" }},
		{"modelo", func(x *NFGas) { x.Ide.Modelo = 55 }},
		{"serie", func(x *NFGas) { x.Ide.Serie = 2 }},
		{"nNF", func(x *NFGas) { x.Ide.NNF = 2 }},
		{"tpEmis", func(x *NFGas) { x.Ide.TpEmis = pcn.TeContingencia }},
		{"nSiteAutoriz", func(x *NFGas) { x.Ide.NSiteAutoriz = Sa1 }},
		{"cNF", func(x *NFGas) { x.Ide.CNF = 2 }},
	}
	for _, a := range alteracoes {
		t.Run(a.nome, func(t *testing.T) {
			copia := lerCompleta(t)
			a.muda(copia)
			if ValidarConcatChave(copia) {
				t.Errorf("mudar %s deveria invalidar a chave", a.nome)
			}
		})
	}
}

func TestMontarChaveAcesso(t *testing.T) {
	n := lerCompleta(t)

	chave, err := MontarChaveAcesso(n)
	if err != nil {
		t.Fatalf("MontarChaveAcesso: %v", err)
	}
	if chave != chaveTeste {
		t.Errorf("chave montada = %q, esperado %q", chave, chaveTeste)
	}
	if len(chave) != TamanhoChaveAcesso {
		t.Errorf("chave tem %d posicoes", len(chave))
	}
	// A chave montada precisa passar na validacao completa.
	if err := pcn.ValidarChaveAcesso(chave); err != nil {
		t.Errorf("chave montada nao valida: %v", err)
	}
}

func TestValidarChaveAcesso_DaFixture(t *testing.T) {
	n := lerCompleta(t)
	if err := ValidarChaveAcesso(n); err != nil {
		t.Errorf("a chave da fixture nao passou: %v", err)
	}
	// cDV declarado tem que bater com o digito da chave.
	dv, err := pcn.DigitoChaveAcesso(chaveTeste[:43])
	if err != nil {
		t.Fatal(err)
	}
	if dv != n.Ide.CDV {
		t.Errorf("cDV declarado = %d, digito calculado = %d", n.Ide.CDV, dv)
	}
}

func TestValidarRegrasNegocio(t *testing.T) {
	n := lerCompleta(t)
	cfg := Configuracoes{
		Ambiente: pcn.TaHomologacao,
		UF:       "SP",
		CodigoUF: 35,
	}

	if erros := ValidarRegrasNegocio(n, cfg); len(erros) != 0 {
		for _, e := range erros {
			t.Errorf("rejeicao inesperada: %v", e)
		}
	}
}

func TestValidarRegrasNegocio_Rejeicoes(t *testing.T) {
	cfgOK := Configuracoes{Ambiente: pcn.TaHomologacao, UF: "SP", CodigoUF: 35}

	casos := []struct {
		nome   string
		muda   func(*NFGas, *Configuracoes)
		codigo int
	}{
		{
			"226 - municipio do emitente de outra UF",
			func(n *NFGas, c *Configuracoes) { n.Emit.EnderEmit.CMun = 4106902 },
			226,
		},
		{
			"227 - chave divergente dos campos",
			func(n *NFGas, c *Configuracoes) { n.Ide.NNF = 999 },
			227,
		},
		{
			"247 - UF do emitente diferente da do servico",
			func(n *NFGas, c *Configuracoes) { n.Emit.EnderEmit.UF = "RJ" },
			247,
		},
		{
			"252 - ambiente diferente do configurado",
			func(n *NFGas, c *Configuracoes) { c.Ambiente = pcn.TaProducao },
			252,
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			n := lerCompleta(t)
			cfg := cfgOK
			c.muda(n, &cfg)

			erros := ValidarRegrasNegocio(n, cfg)
			achou := false
			for _, e := range erros {
				if e.Codigo == c.codigo {
					achou = true
				}
			}
			if !achou {
				t.Errorf("esperava a rejeicao %d, veio %v", c.codigo, erros)
			}
		})
	}
}

func TestErroRegraNegocio_Formato(t *testing.T) {
	e := &ErroRegraNegocio{Codigo: 226, Mensagem: "Rejeicao: teste"}
	if e.Error() != "226-Rejeicao: teste" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestValidarRegrasNegocio_SemConfiguracaoNaoRejeita(t *testing.T) {
	// Quem so importa XML nao tem UF autorizadora configurada; as regras que
	// dependem dela ficam em silencio em vez de rejeitar tudo.
	n := lerCompleta(t)
	cfg := Configuracoes{Ambiente: n.Ide.TpAmb}

	for _, e := range ValidarRegrasNegocio(n, cfg) {
		if e.Codigo == 226 || e.Codigo == 247 {
			t.Errorf("sem UF configurada a regra %d nao deveria disparar", e.Codigo)
		}
	}
}

// ---------------------------------------------------------------------------
// Formato .ini
// ---------------------------------------------------------------------------

func TestLerINI_FormatoDoACBr(t *testing.T) {
	// Arquivo com datas dd/mm/aaaa e decimais com virgula, como o ACBr gera
	// numa maquina pt-BR.
	caminho := filepath.Join("testdata", "nfgas_acbr.ini")
	n, err := LerINIArquivo(caminho, Configuracoes{VersaoDF: Ve100, Ambiente: pcn.TaHomologacao})
	if err != nil {
		t.Fatalf("LerINIArquivo: %v", err)
	}

	if n.InfNFGas.Versao != 1.00 {
		t.Errorf("versao = %v", n.InfNFGas.Versao)
	}
	if n.Ide.CUF != 35 || n.Ide.Modelo != 76 || n.Ide.NNF != 1 {
		t.Errorf("ide = %+v", n.Ide)
	}
	if n.Ide.DhEmi.Day() != 15 || n.Ide.DhEmi.Month() != time.March || n.Ide.DhEmi.Hour() != 10 {
		t.Errorf("dhEmi = %v", n.Ide.DhEmi)
	}
	if n.Ide.TpPagAnt != pcn.TpaPagServicoContinuado {
		t.Errorf("tpPagAnt = %v", n.Ide.TpPagAnt)
	}

	if n.Emit.CNPJ != "11222333000181" || n.Emit.EnderEmit.CMun != 3550308 {
		t.Errorf("emit = %+v", n.Emit)
	}
	// No .ini indIEDest E lido -- diferente do XML.
	if n.Dest.IndIEDest != InIsento {
		t.Errorf("indIEDest = %v, esperado InIsento", n.Dest.IndIEDest)
	}
	if n.Dest.CNPJCPF != "52998224725" {
		t.Errorf("dest/CNPJCPF = %q", n.Dest.CNPJCPF)
	}

	if n.Instalacao.TpClasse != TcResidencial {
		t.Errorf("tpClasse = %v", n.Instalacao.TpClasse)
	}
	if n.GSub.GNF.Serie != "001" || n.GSub.MotSub != MsErroLeitura {
		t.Errorf("gSub = %+v", n.GSub)
	}
	if n.GSub.GNF.CompetEmis.Month() != time.February {
		t.Errorf("CompetEmis = %v", n.GSub.GNF.CompetEmis)
	}

	if len(n.GVolContrat) != 1 || n.GVolContrat[0].QUnidContrat != 1500.123456 {
		t.Errorf("gVolContrat = %+v (decimal com virgula)", n.GVolContrat)
	}
	if len(n.GMed) != 1 || n.GMed[0].VMedAtu != 1250.5678 {
		t.Errorf("gMed = %+v", n.GMed)
	}
	if n.GMed[0].DMedAnt.Month() != time.February {
		t.Errorf("dMedAnt = %v", n.GMed[0].DMedAnt)
	}
}

func TestLerINI_ItemComSubsecoes(t *testing.T) {
	n, err := LerINIArquivo(filepath.Join("testdata", "nfgas_acbr.ini"), Configuracoes{})
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Det) != 1 {
		t.Fatalf("det = %d itens", len(n.Det))
	}
	d := n.Det[0]

	if d.GNormal.Prod.CClass != "0100101" {
		t.Errorf("cClass = %q -- o zero a esquerda tem que sobreviver", d.GNormal.Prod.CClass)
	}
	if d.GNormal.Prod.QFaturada != 250.4444 {
		t.Errorf("qFaturada = %v", d.GNormal.Prod.QFaturada)
	}
	if d.GNormal.Prod.GMedicao.GMedida.VMed != 250.4444 {
		t.Errorf("gMedicao/gMedida achatado = %+v", d.GNormal.Prod.GMedicao)
	}
	if len(d.GNormal.GTarif) != 1 || d.GNormal.GTarif[0].NAto != "ATO-123" {
		t.Errorf("gTarif = %+v", d.GNormal.GTarif)
	}
	if d.GAgregadora.CClass != "0100101" {
		t.Errorf("gAgregadora = %+v", d.GAgregadora)
	}
	if d.GNormal.Imposto.ICMS.CST != pcn.CST20 || d.GNormal.Imposto.ICMS.VBC != 521.41 {
		t.Errorf("ICMS = %+v", d.GNormal.Imposto.ICMS)
	}
	// No .ini motDesICMS ja e lido pelo codigo, tambem no ACBr.
	if d.GNormal.Imposto.ICMS.MotDesICMS != MdiOutros {
		t.Errorf("motDesICMS = %v", d.GNormal.Imposto.ICMS.MotDesICMS)
	}
	if d.GNormal.Imposto.PIS.CST != pcn.Pis01 || d.GNormal.Imposto.COFINS.VCOFINS != 59.44 {
		t.Errorf("PIS/COFINS = %+v / %+v", d.GNormal.Imposto.PIS, d.GNormal.Imposto.COFINS)
	}
	// No .ini a chave e vRetCOFINS em caixa alta.
	if d.GNormal.Imposto.RetTrib.VRetCOFINS != 2.22 {
		t.Errorf("retTrib/vRetCOFINS = %v", d.GNormal.Imposto.RetTrib.VRetCOFINS)
	}
	if d.GNormal.Imposto.TxReg.PTaxa != 0.5 {
		t.Errorf("TxReg = %+v", d.GNormal.Imposto.TxReg)
	}
	if len(d.GNormal.GProcRef.GProc) != 1 {
		t.Errorf("gProc = %+v", d.GNormal.GProcRef.GProc)
	}
	if d.GNormal.Imposto.IBSCBS.CST != rtc.CST200 {
		t.Errorf("IBSCBS = %+v", d.GNormal.Imposto.IBSCBS)
	}
	if d.GNormal.Imposto.IBSCBS.GIBSCBS.GIBSUF.PIBSUF != 0.1 {
		t.Errorf("gIBSUF do .ini = %+v", d.GNormal.Imposto.IBSCBS.GIBSCBS.GIBSUF)
	}

	// gPagAntecipado: no ACBr o campo chDFePagAnt vem de uma variavel nunca
	// inicializada (sempre vazio); aqui e lido da chave homonima.
	if d.GNormal.Prod.GPagAntecipado.ChDFePagAnt != "35260111222333000181760010000000001100000003" ||
		d.GNormal.Prod.GPagAntecipado.NItemPagAnt != 1 {
		t.Errorf("gPagAntecipado do .ini = %+v", d.GNormal.Prod.GPagAntecipado)
	}

	// Grupos de credito/ajuste da Reforma no .ini.
	ibs := d.GNormal.Imposto.IBSCBS
	if ibs.GTransfCred.VIBS != 61 || ibs.GTransfCred.VCBS != 62 {
		t.Errorf("gTransfCred do .ini = %+v", ibs.GTransfCred)
	}
	if ibs.GAjusteCompet.VIBS != 71 || ibs.GAjusteCompet.CompetApur.Month() != time.March {
		t.Errorf("gAjusteCompet do .ini = %+v", ibs.GAjusteCompet)
	}
	if ibs.GCredPresOper.CCredPres != rtc.Cp07 || ibs.GCredPresOper.GIBSCredPres.PCredPres != 1.2345 ||
		ibs.GCredPresOper.GCBSCredPres.VCredPresCondSus != 95 {
		t.Errorf("gCredPresOper do .ini = %+v", ibs.GCredPresOper)
	}
	if ibs.GCredPresIBSZFM.TpCredPresIBSZFM != rtc.TcpBensCapital || ibs.GCredPresIBSZFM.VCredPresIBSZFM != 96 {
		t.Errorf("gCredPresIBSZFM do .ini = %+v", ibs.GCredPresIBSZFM)
	}
}

func TestLerINI_GruposFinais(t *testing.T) {
	n, err := LerINIArquivo(filepath.Join("testdata", "nfgas_acbr.ini"), Configuracoes{})
	if err != nil {
		t.Fatal(err)
	}

	if n.Total.VNF != 948.72 || n.Total.VBC != 571.41 {
		t.Errorf("total = %+v", n.Total)
	}
	if len(n.PgtoVinc.Pgto) != 1 || n.PgtoVinc.Pgto[0].IDTransacao != "TX-0001" {
		t.Errorf("pgtoVinc = %+v", n.PgtoVinc)
	}
	if n.GFat.CompetFat.Month() != time.March || n.GFat.DVencFat.Day() != 10 {
		t.Errorf("gFat = %+v", n.GFat)
	}
	if n.GFat.EnderCorresp.XLgr != "Rua da Correspondencia" {
		t.Errorf("enderCorresp = %+v", n.GFat.EnderCorresp)
	}
	if n.GFat.GPIX.URLQRCodePIX == "" {
		t.Errorf("gPIX = %+v", n.GFat.GPIX)
	}
	if len(n.GAgencia.GHistCons) != 1 || len(n.GAgencia.GHistCons[0].GCons) != 2 {
		t.Errorf("gHistCons/gCons = %+v", n.GAgencia.GHistCons)
	}
	if n.GAgencia.GHistCons[0].GCons[1].VFat != 870 {
		t.Errorf("gCons[1] = %+v", n.GAgencia.GHistCons[0].GCons[1])
	}
	if len(n.AutXML) != 1 || n.AutXML[0].CNPJCPF != "99888777000166" {
		t.Errorf("autXML = %+v", n.AutXML)
	}
	if len(n.InfAdic.InfCpl) != 1 {
		t.Errorf("infCpl = %+v", n.InfAdic.InfCpl)
	}
	if n.InfRespTec.IDCSRT != 1 {
		t.Errorf("infRespTec = %+v", n.InfRespTec)
	}
}

func TestINI_RoundTrip(t *testing.T) {
	original := lerCompleta(t)

	texto, err := GravarINI(original)
	if err != nil {
		t.Fatalf("GravarINI: %v", err)
	}
	relido, err := LerINI(texto, Configuracoes{VersaoDF: Ve100, Ambiente: original.Ide.TpAmb})
	if err != nil {
		t.Fatalf("reler o .ini gerado: %v", err)
	}

	// Identificacao
	if relido.Ide.CUF != original.Ide.CUF || relido.Ide.NNF != original.Ide.NNF {
		t.Errorf("ide: %+v vs %+v", relido.Ide, original.Ide)
	}
	if !relido.Ide.DhEmi.Equal(original.Ide.DhEmi.UTC()) &&
		relido.Ide.DhEmi.Hour() != original.Ide.DhEmi.Hour() {
		t.Errorf("dhEmi: %v vs %v", relido.Ide.DhEmi, original.Ide.DhEmi)
	}
	if relido.Ide.TpPagAnt != original.Ide.TpPagAnt {
		t.Errorf("tpPagAnt: %v vs %v", relido.Ide.TpPagAnt, original.Ide.TpPagAnt)
	}

	// Emitente e destinatario
	if relido.Emit.CNPJ != original.Emit.CNPJ || relido.Emit.ISUFEmit != original.Emit.ISUFEmit {
		t.Errorf("emit: %+v", relido.Emit)
	}
	if relido.Dest.CNPJCPF != original.Dest.CNPJCPF || relido.Dest.CNIS != original.Dest.CNIS {
		t.Errorf("dest: %+v", relido.Dest)
	}

	// Colecoes
	if len(relido.GVolContrat) != len(original.GVolContrat) {
		t.Errorf("gVolContrat: %d vs %d", len(relido.GVolContrat), len(original.GVolContrat))
	}
	if len(relido.GMed) != len(original.GMed) {
		t.Errorf("gMed: %d vs %d", len(relido.GMed), len(original.GMed))
	}
	if len(relido.Det) != len(original.Det) {
		t.Fatalf("det: %d vs %d", len(relido.Det), len(original.Det))
	}
	if len(relido.AutXML) != len(original.AutXML) {
		t.Errorf("autXML: %d vs %d", len(relido.AutXML), len(original.AutXML))
	}

	// Item: valores que passam por formatacao decimal.
	po, pr := original.Det[0].GNormal.Prod, relido.Det[0].GNormal.Prod
	if pr.CClass != po.CClass {
		t.Errorf("cClass: %q vs %q", pr.CClass, po.CClass)
	}
	if pr.QFaturada != po.QFaturada {
		t.Errorf("qFaturada: %v vs %v", pr.QFaturada, po.QFaturada)
	}
	if pr.VProd != po.VProd {
		t.Errorf("vProd: %v vs %v", pr.VProd, po.VProd)
	}
	if pr.VItem != po.VItem {
		t.Errorf("vItem: %v vs %v", pr.VItem, po.VItem)
	}

	io, ir := original.Det[0].GNormal.Imposto, relido.Det[0].GNormal.Imposto
	if ir.ICMS.CST != io.ICMS.CST || ir.ICMS.VBC != io.ICMS.VBC || ir.ICMS.PFCP != io.ICMS.PFCP {
		t.Errorf("ICMS: %+v vs %+v", ir.ICMS, io.ICMS)
	}
	if ir.ICMS.MotDesICMS != io.ICMS.MotDesICMS {
		t.Errorf("motDesICMS: %v vs %v", ir.ICMS.MotDesICMS, io.ICMS.MotDesICMS)
	}
	if ir.RetTrib.VRetCOFINS != io.RetTrib.VRetCOFINS {
		t.Errorf("vRetCOFINS: %v vs %v", ir.RetTrib.VRetCOFINS, io.RetTrib.VRetCOFINS)
	}
	if ir.IBSCBS.CST != io.IBSCBS.CST || ir.IBSCBS.GIBSCBS.VBC != io.IBSCBS.GIBSCBS.VBC {
		t.Errorf("IBSCBS: %+v vs %+v", ir.IBSCBS, io.IBSCBS)
	}
	if pr.GPagAntecipado != po.GPagAntecipado {
		t.Errorf("gPagAntecipado: %+v vs %+v", pr.GPagAntecipado, po.GPagAntecipado)
	}
	if ir.IBSCBS.GEstornoCred != io.IBSCBS.GEstornoCred {
		t.Errorf("gEstornoCred: %+v vs %+v", ir.IBSCBS.GEstornoCred, io.IBSCBS.GEstornoCred)
	}

	// Totais e grupos finais
	if relido.Total.VNF != original.Total.VNF || relido.Total.VBC != original.Total.VBC {
		t.Errorf("total: %+v", relido.Total)
	}
	if relido.Total.IBSCBSTot.VBCIBSCBS != original.Total.IBSCBSTot.VBCIBSCBS {
		t.Errorf("IBSCBSTot: %+v", relido.Total.IBSCBSTot)
	}
	if len(relido.GAgencia.GHistCons) != len(original.GAgencia.GHistCons) {
		t.Errorf("gHistCons: %d vs %d", len(relido.GAgencia.GHistCons), len(original.GAgencia.GHistCons))
	}
	if len(relido.GAgencia.GHistCons) > 0 &&
		len(relido.GAgencia.GHistCons[0].GCons) != len(original.GAgencia.GHistCons[0].GCons) {
		t.Errorf("gCons: %d vs %d",
			len(relido.GAgencia.GHistCons[0].GCons), len(original.GAgencia.GHistCons[0].GCons))
	}
	if !relido.GFat.CompetFat.Equal(original.GFat.CompetFat) {
		t.Errorf("CompetFat: %v vs %v", relido.GFat.CompetFat, original.GFat.CompetFat)
	}
}

func TestGravarINI_UsaPontoDecimal(t *testing.T) {
	n := lerCompleta(t)
	texto, err := GravarINI(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(texto, "vNF=948.72") {
		t.Errorf("o gravador deveria usar ponto decimal; trecho de [total]:\n%s",
			recortar(texto, "[total]", 400))
	}
	if strings.Contains(texto, "948,72") {
		t.Error("nenhum decimal deveria sair com virgula")
	}
}

func TestGravarINIArquivo(t *testing.T) {
	n := lerCompleta(t)
	caminho := filepath.Join(t.TempDir(), "saida.ini")

	if err := GravarINIArquivo(n, caminho); err != nil {
		t.Fatalf("GravarINIArquivo: %v", err)
	}
	relido, err := LerINIArquivo(caminho, Configuracoes{})
	if err != nil {
		t.Fatalf("reler o arquivo gravado: %v", err)
	}
	if relido.Ide.NNF != n.Ide.NNF {
		t.Errorf("nNF apos o round-trip em disco = %d", relido.Ide.NNF)
	}
}

func recortar(texto, marca string, tamanho int) string {
	p := strings.Index(texto, marca)
	if p < 0 {
		return texto
	}
	fim := p + tamanho
	if fim > len(texto) {
		fim = len(texto)
	}
	return texto[p:fim]
}

// ---------------------------------------------------------------------------
// Componente
// ---------------------------------------------------------------------------

func TestComponente_CarregarEValidar(t *testing.T) {
	c := NovoComponente()
	c.Configuracoes.Ambiente = pcn.TaHomologacao
	c.Configuracoes.UF = "SP"
	c.Configuracoes.CodigoUF = 35

	caminho := filepath.Join(t.TempDir(), "nota.xml")
	if err := os.WriteFile(caminho, carregarFixture(t, "nfgas_completa.xml"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := c.CarregarArquivo(caminho); err != nil {
		t.Fatalf("CarregarArquivo: %v", err)
	}
	if len(c.NotasFiscais) != 1 {
		t.Fatalf("NotasFiscais = %d", len(c.NotasFiscais))
	}

	rej := c.ValidarRegrasDeNegocio()
	if len(rej) != 1 {
		t.Fatalf("resultado da validacao = %d listas", len(rej))
	}
	if len(rej[0]) != 0 {
		t.Errorf("rejeicoes inesperadas: %v", rej[0])
	}

	c.Limpar()
	if len(c.NotasFiscais) != 0 {
		t.Error("Limpar nao esvaziou a lista")
	}
}

func TestComponente_ValorZeroFunciona(t *testing.T) {
	var c Componente
	if err := c.CarregarString(string(carregarFixture(t, "nfgas_completa.xml"))); err != nil {
		t.Fatalf("o valor zero do Componente deveria funcionar: %v", err)
	}
	if len(c.NotasFiscais) != 1 {
		t.Errorf("NotasFiscais = %d", len(c.NotasFiscais))
	}
	if c.NomeModelo() != "NFGas" || c.NamespaceURI() != Namespace {
		t.Errorf("modelo/namespace = %q/%q", c.NomeModelo(), c.NamespaceURI())
	}
}

func TestComponente_CarregarINI(t *testing.T) {
	c := NovoComponente()
	if err := c.CarregarINI(filepath.Join("testdata", "nfgas_acbr.ini")); err != nil {
		t.Fatalf("CarregarINI: %v", err)
	}
	if len(c.NotasFiscais) != 1 || c.NotasFiscais[0].NFGas.Ide.NNF != 1 {
		t.Errorf("NotasFiscais = %+v", c.NotasFiscais)
	}
}

func TestIdentificarSchema(t *testing.T) {
	casos := map[string]SchemaNFGas{
		`<NFGas/>`:           SchNFGas,
		`<nfgasProc/>`:       SchNFGas,
		`<retConsSitNFGas/>`: SchconsSitNFGas,
		`<procEventoNFGas/>`: SchEventoNFGas,
		`<retEventoNFGas/>`:  SchEventoNFGas,
		`<evCancNFGas/>`:     SchevCancNFGas,
		`<retNFGas/>`:        SchretNFGas,
	}
	for xml, esperado := range casos {
		got, err := IdentificarSchema(xml)
		if err != nil {
			t.Errorf("IdentificarSchema(%q): %v", xml, err)
			continue
		}
		if got != esperado {
			t.Errorf("IdentificarSchema(%q) = %v, esperado %v", xml, got, esperado)
		}
	}
	if _, err := IdentificarSchema(`<qualquerCoisa/>`); !errors.Is(err, ErrXMLIncorreto) {
		t.Errorf("raiz desconhecida deveria dar ErrXMLIncorreto, veio %v", err)
	}
}

// ---------------------------------------------------------------------------
// Emissao -- contratos sem rede (a bateria completa esta em
// xml_writer_test.go e web_services_test.go)
// ---------------------------------------------------------------------------

func TestEmissao_ContratosSemRede(t *testing.T) {
	c := NovoComponente()
	ctx := testContext()

	// transmissao sem certificado configurado falha cedo, sem rede
	if _, err := c.Consultar(ctx, chaveTeste); !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Errorf("Consultar sem certificado = %v", err)
	}
	if _, err := c.StatusServico(ctx); !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Errorf("StatusServico sem certificado = %v", err)
	}
	if _, err := c.Cancelamento(ctx, chaveTeste, "191000000000001", "justificativa de teste"); !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Errorf("Cancelamento sem certificado = %v", err)
	}

	// URL publica de consulta nao depende de certificado
	if url, err := c.URLConsultaNFGas(35); err != nil || url == "" {
		t.Errorf("URLConsultaNFGas(35) = %q, %v", url, err)
	}
	// MA e PA apontam para o SVAN, sem URL publicada (lacuna do ACBr)
	if _, err := c.URLConsultaNFGas(21); !errors.Is(err, ErrSemURL) {
		t.Errorf("URLConsultaNFGas(MA) deveria dar ErrSemURL, veio %v", err)
	}
}

// ---------------------------------------------------------------------------
// Erros
// ---------------------------------------------------------------------------

func TestErroNFGas_Contexto(t *testing.T) {
	e := &ErroNFGas{Indice: 2, Chave: chaveTeste, Arquivo: "lote.xml", Err: ErrXMLIncorreto}
	msg := e.Error()

	if !strings.Contains(msg, "lote.xml") || !strings.Contains(msg, "documento 3") {
		t.Errorf("mensagem sem contexto: %q", msg)
	}
	if !strings.Contains(msg, chaveTeste) {
		t.Errorf("mensagem sem a chave: %q", msg)
	}
	if !errors.Is(e, ErrXMLIncorreto) {
		t.Error("errors.Is deveria alcancar o erro subjacente")
	}
}

func testContext() context.Context { return context.Background() }
