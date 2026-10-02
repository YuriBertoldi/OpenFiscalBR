// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-30.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfag

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// ---------------------------------------------------------------------------
// GerarXML -- round-trip e fidelidade ao TNFAgXmlWriter
// ---------------------------------------------------------------------------

// TestGerarXML_RoundTripEstavel: o XML gerado da fixture completa, relido e
// regerado, tem que ser byte a byte igual.
func TestGerarXML_RoundTripEstavel(t *testing.T) {
	n := lerCompleta(t)

	xml1, err := GerarXML(n)
	if err != nil {
		t.Fatalf("GerarXML (1a): %v", err)
	}
	relido, err := LerBytes([]byte(xml1))
	if err != nil {
		t.Fatalf("LerBytes do XML gerado: %v", err)
	}
	xml2, err := GerarXML(relido)
	if err != nil {
		t.Fatalf("GerarXML (2a): %v", err)
	}
	if xml1 != xml2 {
		t.Fatalf("round-trip instavel:\n1a: %s\n2a: %s", xml1, xml2)
	}
}

func TestGerarXML_IdComLiteralMaiusculo(t *testing.T) {
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	// o XSD exige NFAG (maiusculo) na frente da chave
	if !strings.Contains(xml, `<infNFAg Id="NFAG`+chaveTeste+`" versao="1.00">`) {
		t.Fatalf("Id sem o literal NFAG maiusculo: %.200s", xml)
	}
	if !strings.HasPrefix(xml, `<NFAg xmlns="`+Namespace+`">`) {
		t.Fatalf("raiz inesperada: %.100s", xml)
	}
	if n.Ide.CDV != 6 {
		t.Fatalf("cDV = %d, esperado 6", n.Ide.CDV)
	}
}

func TestGerarXML_OrdemDaIde(t *testing.T) {
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	ide := xml[strings.Index(xml, "<ide>"):strings.Index(xml, "</ide>")]
	ordem := []string{"<cUF>", "<tpAmb>", "<mod>75</mod>", "<serie>", "<nNF>",
		"<cNF>", "<cDV>", "<dhEmi>", "<tpEmis>", "<nSiteAutoriz>", "<cMunFG>",
		"<finNFAg>", "<tpFat>", "<verProc>", "<gCompraGov>", "<tpPagAnt>"}
	pos := -1
	for _, tag := range ordem {
		p := strings.Index(ide, tag)
		if p < 0 {
			t.Fatalf("ide sem %s: %s", tag, ide)
		}
		if p < pos {
			t.Fatalf("%s fora de ordem em: %s", tag, ide)
		}
		pos = p
	}
}

func TestGerarXML_HomologacaoTrocaXNomeDest(t *testing.T) {
	// GerarDest: em homologacao o xNome do destinatario e substituido pelo
	// texto fixo (que diz "NF-E", copiado da NFe -- literal replicado)
	n := notaMinima()
	n.Ide.TpAmb = pcn.TaHomologacao
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<xNome>NF-E EMITIDA EM AMBIENTE DE HOMOLOGACAO - SEM VALOR FISCAL</xNome>") {
		t.Fatalf("homologacao deveria trocar o xNome do dest: %s", trecho(xml, "dest"))
	}

	n2 := notaMinima()
	n2.Ide.TpAmb = pcn.TaProducao
	xml2, err := GerarXML(n2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml2, "<xNome>Cliente Teste</xNome>") {
		t.Fatalf("producao deveria manter o xNome real: %s", trecho(xml2, "dest"))
	}
}

func TestGerarXML_CClassComPadDeZeros(t *testing.T) {
	// cClass e Integer no ACBr gravado com tcInt min 7 -- o pad de zeros
	// reproduz o PadLeft
	n := notaMinima()
	n.Det[0].Prod.CClass = "100101"
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<cClass>0100101</cClass>") {
		t.Fatalf("cClass sem pad de 7: %s", trecho(xml, "cClass"))
	}
}

func TestGerarXML_QFaturadaIntVsDe4(t *testing.T) {
	// Frac() > 0 -> tcDe4; inteira -> tcInt (no ACBr o campo e Integer e o
	// ramo De4 e codigo morto; aqui os dois vivem)
	n := notaMinima()
	n.Det[0].Prod.QFaturada = 10
	xml, _ := GerarXML(n)
	if !strings.Contains(xml, "<qFaturada>10</qFaturada>") {
		t.Fatalf("qFaturada inteira deveria sair sem casas: %s", trecho(xml, "qFaturada"))
	}

	n2 := notaMinima()
	n2.Det[0].Prod.QFaturada = 10.5
	xml2, _ := GerarXML(n2)
	if !strings.Contains(xml2, "<qFaturada>10.5000</qFaturada>") {
		t.Fatalf("qFaturada fracionada deveria sair De4: %s", trecho(xml2, "qFaturada"))
	}
}

func TestGerarXML_GMedicaoRamos(t *testing.T) {
	// gMedida quando ha leitura anterior OU atual; senao tpMotNaoLeitura
	n := notaMinima()
	n.Det[0].Prod.GMedicao = GMedicao{NMed: 1, GMedida: GMedida{VMedAnt: 100, VMedAtu: 150, VMed: 50, TpGrMed: TgmAguaTratada, UMed: UmM3}}
	xml, _ := GerarXML(n)
	if !strings.Contains(xml, "<gMedida>") || strings.Contains(xml, "tpMotNaoLeitura") {
		t.Fatalf("com leitura deveria sair gMedida: %s", trecho(xml, "gMedicao"))
	}

	n2 := notaMinima()
	n2.Det[0].Prod.GMedicao = GMedicao{NMed: 1, TpMotNaoLeitura: TmDistribuidora}
	xml2, _ := GerarXML(n2)
	if strings.Contains(xml2, "<gMedida>") || !strings.Contains(xml2, "<tpMotNaoLeitura>2</tpMotNaoLeitura>") {
		t.Fatalf("sem leitura deveria sair tpMotNaoLeitura: %s", trecho(xml2, "gMedicao"))
	}
}

func TestGerarXML_ImpostoSempreEGruposCondicionais(t *testing.T) {
	// o elemento imposto e SEMPRE gerado; PIS/COFINS/retTrib/TFS/TFU so
	// quando tem valor
	n := notaMinima()
	n.Det[0].Imposto = Imposto{}
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<imposto/>") {
		t.Fatalf("imposto vazio deveria sair como elemento vazio: %s", trecho(xml, "det"))
	}

	n2 := notaMinima()
	n2.Det[0].Imposto.TFU = TFU{VBCTFU: 100, PTFU: 0.25, VTFU: 0.25}
	xml2, _ := GerarXML(n2)
	if !strings.Contains(xml2, "<TFU><vBCTFU>100.00</vBCTFU><pTFU>0.25</pTFU><vTFU>0.25</vTFU></TFU>") {
		t.Fatalf("TFU: %s", trecho(xml2, "imposto"))
	}
}

func TestGerarXML_RetTribGrafiaDoXSDEVBCIRRF(t *testing.T) {
	// DIVERGENCIA: o ACBr grava vRetCOFINS (maiusculo) no item -- o XSD e o
	// leitor usam vRetCofins; a grafia foi corrigida para a do XSD (senao o
	// valor se perde na releitura). vBCIRRF E gerado (e nunca lido).
	n := notaMinima()
	n.Det[0].Imposto.RetTrib = RetTrib{VRetPIS: 1, VRetCOFINS: 2, VRetCSLL: 3, VBCIRRF: 100, VIRRF: 4}
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	ret := trecho(xml, "retTrib")
	for _, m := range []string{"<vRetCofins>2.00</vRetCofins>", "<vBCIRRF>100.00</vBCIRRF>"} {
		if !strings.Contains(ret, m) {
			t.Fatalf("retTrib sem %q: %s", m, ret)
		}
	}
	if strings.Contains(ret, "vRetCOFINS") {
		t.Fatal("a grafia maiuscula do ACBr e invalida contra o XSD")
	}
}

func TestGerarXML_GProcRefSoComVItem(t *testing.T) {
	n := notaMinima()
	n.Det[0].GProcRef = GProcRef{QFaturada: 5, VProd: 10} // vItem zero
	xml, _ := GerarXML(n)
	if strings.Contains(xml, "<gProcRef>") {
		t.Fatal("gProcRef sem vItem nao deveria sair (condicao vItem > 0)")
	}

	n2 := notaMinima()
	n2.Det[0].GProcRef = GProcRef{VItem: 2.5, QFaturada: 5, VProd: 12.5,
		GProc: []GProc{{TpProc: TpProcon, NProcesso: "P1"}}}
	xml2, _ := GerarXML(n2)
	g := trecho(xml2, "gProcRef")
	// vItem/vProd De2 na geracao (o leitor le De8 -- assimetria replicada)
	for _, m := range []string{"<vItem>2.50</vItem>", "<qFaturada>5</qFaturada>",
		"<vProd>12.50</vProd>", "<tpProc>5</tpProc>"} {
		if !strings.Contains(g, m) {
			t.Fatalf("gProcRef sem %q: %s", m, g)
		}
	}
}

func TestGerarXML_GFatSempreECodBancoObrigatorioNoElse(t *testing.T) {
	// gFat e SEMPRE gerado na NFAg; sem codDebAuto, codBanco/codAgencia
	// saem obrigatorios (tags vazias se nao informados)
	n := notaMinima()
	n.GFat = GFat{} // tudo zero
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	g := xml[strings.Index(xml, "<gFat>"):]
	for _, m := range []string{"<CompetFat/>", "<dVencFat/>", "<codBarras/>",
		"<codBanco/>", "<codAgencia/>"} {
		if !strings.Contains(g, m) {
			t.Fatalf("gFat vazio deveria ter %q (obrigatorios do original): %.300s", m, g)
		}
	}
	if strings.Contains(g, "189912") {
		t.Fatal("competencia zero vazou como 189912")
	}
}

func TestGerarXML_LigacaoEGAgenciaSempre(t *testing.T) {
	n := notaMinima()
	n.Ligacao = Ligacao{}
	n.GAgencia = GAgencia{}
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	// ligacao sempre sai (sem o guard da NFGas), com os obrigatorios vazios
	if !strings.Contains(xml, "<ligacao>") || !strings.Contains(xml, "<idLigacao/>") {
		t.Fatalf("ligacao deveria sair sempre: %s", trecho(xml, "ligacao"))
	}
	// gAgencia sempre sai, com nAgenciaAtend/enderAgenciaAtend obrigatorios
	if !strings.Contains(xml, "<gAgencia>") || !strings.Contains(xml, "<nAgenciaAtend/>") {
		t.Fatalf("gAgencia deveria sair sempre: %s", xml[strings.Index(xml, "<gAgencia"):])
	}
}

func TestGerarXML_InfAdicSoComConteudo(t *testing.T) {
	// diferente da NFGas: infAdic so sai quando tem conteudo
	n := notaMinima()
	n.InfAdic = InfAdic{}
	xml, _ := GerarXML(n)
	if strings.Contains(xml, "infAdic") {
		t.Fatal("infAdic vazio nao deveria sair na NFAg")
	}
}

func TestGerarXMLProc_RaizConformeXSD(t *testing.T) {
	n, err := LerBytes(carregarFixture(t, "nfag_proc.xml"))
	if err != nil {
		t.Fatal(err)
	}
	xml, err := GerarXMLProc(n)
	if err != nil {
		t.Fatalf("GerarXMLProc: %v", err)
	}
	// DIVERGENCIA: raiz nfagProc conforme o XSD (o ACBr gera e le NFAgProc)
	if !strings.HasPrefix(xml, `<nfagProc versao="1.00" xmlns="`+Namespace+`">`) {
		t.Fatalf("raiz do proc inesperada: %.120s", xml)
	}
	for _, tag := range []string{"<protNFAg versao=", "<infProt>", "<nProt>335260000000101</nProt>"} {
		if !strings.Contains(xml, tag) {
			t.Fatalf("proc sem %s:\n%s", tag, xml)
		}
	}
	// relido, protocolo se preserva
	relido, err := LerBytes([]byte(xml))
	if err != nil {
		t.Fatalf("reler proc: %v", err)
	}
	if relido.ProcNFAg.NProt != n.ProcNFAg.NProt {
		t.Fatalf("protocolo perdido no round-trip do proc")
	}

	semProt := notaMinima()
	if _, err := GerarXMLProc(semProt); !errors.Is(err, ErrProtocoloAusente) {
		t.Fatalf("esperado ErrProtocoloAusente, veio %v", err)
	}
}

func TestGerarXML_CNFZeroSorteiaCodigoValido(t *testing.T) {
	n := notaMinima()
	n.Ide.CNF = 0
	if _, err := GerarXML(n); err != nil {
		t.Fatal(err)
	}
	if n.Ide.CNF == 0 || n.Ide.CNF == n.Ide.NNF || n.Ide.CNF > 9999999 {
		t.Fatalf("cNF sorteado invalido: %d", n.Ide.CNF)
	}
}

// ---------------------------------------------------------------------------
// Evento
// ---------------------------------------------------------------------------

func TestGerarXMLEvento_Cancelamento(t *testing.T) {
	e := &EventoNFAg{
		InfEvento: InfEvento{
			TpAmb:      pcn.TaHomologacao,
			ChNFAg:     chaveTeste,
			DhEvento:   time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC),
			TpEvento:   TeCancelamento,
			NSeqEvento: 1,
			DetEvento:  DetEvento{NProt: "335260000000101", XJust: "cancelamento em teste unitario"},
		},
	}
	xml, err := GerarXMLEvento(e)
	if err != nil {
		t.Fatal(err)
	}
	wantID := "ID" + TeCancelamento.String() + chaveTeste + "01"
	if e.InfEvento.ID != wantID {
		t.Fatalf("Id do evento = %q, esperado %q", e.InfEvento.ID, wantID)
	}
	for _, m := range []string{
		`<eventoNFAg xmlns="` + Namespace + `" versao="1.00">`,
		"<cOrgao>35</cOrgao>",
		"<CNPJ>11222333000181</CNPJ>",
		"<tpEvento>110111</tpEvento>",
		`<detEvento versaoEvento="1.00"><evCancNFAg>`,
		"<descEvento>Cancelamento</descEvento>",
	} {
		if !strings.Contains(xml, m) {
			t.Fatalf("evento sem %q:\n%s", m, xml)
		}
	}
}

func TestGerarXMLEvento_SoCancelamento(t *testing.T) {
	// os outros 3 tipos existem no leiaute de RETORNO, mas o gerador so
	// implementa cancelamento (como o ACBr)
	for _, tp := range []TipoEvento{TeAutorizadoSubstituicao, TeAutorizadoAjuste, TeLiberacaoPrazoCancelado} {
		e := &EventoNFAg{InfEvento: InfEvento{TpEvento: tp, ChNFAg: chaveTeste}}
		if _, err := GerarXMLEvento(e); err == nil {
			t.Errorf("GerarXMLEvento(%s) deveria falhar: gerador so implementa cancelamento", tp.Descricao())
		}
	}
}

// procEventoTeste monta um retorno de cancelamento completo, nos moldes de
// testdata/proc_evento_cancelamento.xml.
func procEventoTeste() *RetEventoNFAg {
	return &RetEventoNFAg{
		Versao:    "1.00",
		TemEvento: true,
		Evento: EventoNFAg{
			InfEvento: InfEvento{
				TpAmb:      pcn.TaHomologacao,
				ChNFAg:     chaveTeste,
				DhEvento:   time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
				TpEvento:   TeCancelamento,
				NSeqEvento: 1,
				DetEvento:  DetEvento{NProt: "335260000000001", XJust: "cancelamento em teste unitario"},
			},
		},
		RetInfEvento: RetInfEvento{
			TpAmb:       pcn.TaHomologacao,
			VerAplic:    "SP_NFAG_1.0.0",
			COrgao:      35,
			CStat:       135,
			XMotivo:     "Evento registrado e vinculado a NFAg",
			ChNFAg:      chaveTeste,
			TpEvento:    TeCancelamento,
			XEvento:     "Cancelamento",
			NSeqEvento:  1,
			CNPJDest:    "99888777000166",
			EmailDest:   "dest@teste.com.br",
			COrgaoAutor: 35,
			DhRegEvento: time.Date(2026, 3, 10, 9, 5, 0, 0, time.UTC),
			NProt:       "135260000000099",
		},
	}
}

func TestGerarXMLProcEvento(t *testing.T) {
	r := procEventoTeste()
	xml, err := GerarXMLProcEvento(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{
		`<procEventoNFAg versao="1.00" xmlns="` + Namespace + `">`,
		`<eventoNFAg versao="1.00">`, // namespace so na raiz, nao repetido aqui
		"<evCancNFAg>",
		`<retEventoNFAg versao="1.00">`,
		"<cStat>135</cStat>",
		"<nProt>135260000000099</nProt>",
		"<CNPJDest>99888777000166</CNPJDest>",
	} {
		if !strings.Contains(xml, m) {
			t.Fatalf("procEvento sem %q:\n%s", m, xml)
		}
	}

	// round-trip: o leitor tem que achar as DUAS partes do envelope
	lido, err := LerEventoString(xml)
	if err != nil {
		t.Fatalf("reler procEvento: %v", err)
	}
	if !lido.TemEvento {
		t.Fatal("round-trip perdeu a parte enviada do evento")
	}
	if got := lido.Justificativa(); got != "cancelamento em teste unitario" {
		t.Fatalf("xJust = %q", got)
	}
	if !lido.RetInfEvento.Registrado() || lido.RetInfEvento.NProt != "135260000000099" {
		t.Fatalf("retorno do evento nao sobreviveu: %+v", lido.RetInfEvento)
	}
	if lido.ChaveAcesso() != chaveTeste {
		t.Fatalf("chave = %q", lido.ChaveAcesso())
	}
}

// TestGerarXMLProcEvento_OrdemDosCampos trava a ORDEM das tags do
// retEventoNFAg. O leitor usa FindAnyNs campo a campo e e insensivel a
// ordem, entao o round-trip nao protege nada aqui: trocar cStat de lugar com
// xMotivo passaria em build, vet e em todos os outros testes, e geraria XML
// que o XSD da SEFAZ rejeita. A ordem esperada vem de
// testdata/proc_evento_cancelamento.xml.
func TestGerarXMLProcEvento_OrdemDosCampos(t *testing.T) {
	r := procEventoTeste()
	xml, err := GerarXMLProcEvento(r)
	if err != nil {
		t.Fatal(err)
	}

	ret := trecho(xml, "retEventoNFAg")
	if ret == "" {
		// a tag tem atributo, entao trecho() nao a encontra pelo nome puro
		ini := strings.Index(xml, "<retEventoNFAg")
		if ini < 0 {
			t.Fatal("retEventoNFAg ausente")
		}
		ret = xml[ini:]
	}

	ordem := []string{
		"<tpAmb>", "<verAplic>", "<cOrgao>", "<cStat>", "<xMotivo>",
		"<chNFAg>", "<tpEvento>", "<xEvento>", "<nSeqEvento>",
		"<CNPJDest>", "<emailDest>", "<cOrgaoAutor>", "<dhRegEvento>",
		"<nProt>",
	}
	anterior := -1
	for _, tag := range ordem {
		pos := strings.Index(ret, tag)
		if pos < 0 {
			t.Fatalf("retEventoNFAg sem %s:\n%s", tag, ret)
		}
		if pos <= anterior {
			t.Fatalf("%s fora de ordem no retEventoNFAg (posicao %d, anterior %d):\n%s",
				tag, pos, anterior, ret)
		}
		anterior = pos
	}
}

// TestGerarXMLProcEvento_RetornoParcial cobre o fallback para o evento
// enviado: retorno so com nProt e cStat -- o caso de quem preenche a mao --
// nao pode produzir tpEvento "-99999" (o String() do zero value).
func TestGerarXMLProcEvento_RetornoParcial(t *testing.T) {
	r := procEventoTeste()
	r.RetInfEvento = RetInfEvento{
		TpAmb:   pcn.TaHomologacao,
		CStat:   135,
		XMotivo: "Evento registrado e vinculado a NFAg",
		NProt:   "135260000000099",
	}

	xml, err := GerarXMLProcEvento(r)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(xml, "-99999") {
		t.Fatalf("tpEvento caiu no zero value do enum:\n%s", xml)
	}
	for _, m := range []string{
		"<tpEvento>110111</tpEvento>",
		"<nSeqEvento>1</nSeqEvento>",
		"<cOrgao>35</cOrgao>",
		"<xEvento>Cancelamento</xEvento>",
		"<chNFAg>" + chaveTeste + "</chNFAg>",
	} {
		if !strings.Contains(xml, m) {
			t.Errorf("retorno parcial nao herdou %q do evento enviado", m)
		}
	}
}

func TestGerarXMLProcEvento_Incompleto(t *testing.T) {
	if _, err := GerarXMLProcEvento(nil); !errors.Is(err, ErrXMLVazio) {
		t.Fatalf("nil deveria dar ErrXMLVazio, veio %v", err)
	}
	// retorno sem a parte enviada: sem ela nao ha xJust, e o envelope sairia
	// pela metade -- recusa explicita em vez de documento incompleto
	semEvento := &RetEventoNFAg{RetInfEvento: RetInfEvento{NProt: "1"}}
	if _, err := GerarXMLProcEvento(semEvento); !errors.Is(err, ErrEventoAusente) {
		t.Fatalf("sem evento deveria dar ErrEventoAusente, veio %v", err)
	}
	semProt := &RetEventoNFAg{
		TemEvento: true,
		Evento: EventoNFAg{InfEvento: InfEvento{
			TpEvento: TeCancelamento, ChNFAg: chaveTeste, NSeqEvento: 1,
		}},
	}
	if _, err := GerarXMLProcEvento(semProt); !errors.Is(err, ErrProtocoloAusente) {
		t.Fatalf("sem protocolo deveria dar ErrProtocoloAusente, veio %v", err)
	}
}

// ---------------------------------------------------------------------------
// QR-Code e URLs
// ---------------------------------------------------------------------------

func TestGerarQRCode(t *testing.T) {
	n := notaMinima()
	if _, err := GerarXML(n); err != nil {
		t.Fatal(err)
	}
	url, err := GerarQRCode(n, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://dfe-portal.svrs.rs.gov.br/nfag/qrCode?chNFAg=" + n.ChaveAcesso() + "&tpAmb=2"
	if url != want {
		t.Fatalf("QRCode = %q, esperado %q", url, want)
	}
}

func TestURLServico(t *testing.T) {
	url, err := URLServico("SP", pcn.TaHomologacao, ServicoStatusServico)
	if err != nil {
		t.Fatal(err)
	}
	// host nfag e path "ws" MINUSCULO (diferente da NFGas)
	if url != "https://nfag-homologacao.svrs.rs.gov.br/ws/NFAgStatusServico/NFAgStatusServico.asmx" {
		t.Fatalf("URL homologacao = %q", url)
	}
	for _, uf := range []string{"MA", "PA"} {
		if _, err := URLServico(uf, pcn.TaProducao, ServicoRecepcao); !errors.Is(err, ErrSemURL) {
			t.Fatalf("UF %s deveria dar ErrSemURL, veio %v", uf, err)
		}
	}
	// acoes do NFAg tem grafia propria em minusculo no metodo
	if SoapAction(ServicoStatusServico) != Namespace+"/wsdl/NFAgStatusServico/nfagStatusServicoNF" {
		t.Fatalf("SoapAction status = %q", SoapAction(ServicoStatusServico))
	}
	if SoapAction(ServicoConsulta) != Namespace+"/wsdl/NFAgConsulta/nfagConsultaNF" {
		t.Fatalf("SoapAction consulta = %q", SoapAction(ServicoConsulta))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// notaMinima monta uma NFAg valida minima para testes de geracao.
func notaMinima() *NFAg {
	return &NFAg{
		InfNFAg: InfNFAg{Versao: 1.00},
		Ide: Ide{
			CUF: 35, TpAmb: pcn.TaHomologacao, Modelo: 75, Serie: 1, NNF: 1,
			CNF: 18, DhEmi: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
			TpEmis: pcn.TeNormal, NSiteAutoriz: Sa1, CMunFG: 3550308,
			FinNFAg: FnNormal, TpFat: TfNormal, VerProc: "OpenFiscalBR",
		},
		Emit: Emit{
			CNPJ: "11222333000181", IE: "111111111111", XNome: "Saneamento Teste",
			EnderEmit: Endereco{XLgr: "Rua A", Nro: "1", XBairro: "Centro",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1000000, UF: "SP"},
		},
		Dest: Dest{
			XNome: "Cliente Teste", CNPJCPF: "52998224725",
			EnderDest: Endereco{XLgr: "Rua B", Nro: "2", XBairro: "Centro",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1000000, UF: "SP"},
		},
		Ligacao: Ligacao{IDLigacao: "LIG-1", TpLigacao: TlAgua, LatGPS: "-23.5", LongGPS: "-46.6"},
		Det: []Det{{
			NItem: 1,
			Prod: Prod{
				IndOrigemQtd: IoMedia, CProd: "AGUA", XProd: "Agua canalizada",
				CClass: "0100101", TpCategoria: TcResidencial,
				UMed: UmM3, QFaturada: 10, VItem: 2.5, VProd: 25,
			},
		}},
		Total: Total{VProd: 25, VNF: 25},
		GFat: GFat{
			CompetFat:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			DVencFat:     time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			DProxLeitura: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
			CodBarras:    "84670000000000250011222333000181000000000001",
			CodBanco:     "001", CodAgencia: "1234",
		},
		GAgencia: GAgencia{NAgenciaAtend: "AG-1", EnderAgenciaAtend: "Praca Central, 1"},
	}
}

// trecho devolve o conteudo da primeira ocorrencia da tag.
func trecho(xml, tag string) string {
	ini := strings.Index(xml, "<"+tag+">")
	if ini < 0 {
		return ""
	}
	ini += len(tag) + 2
	fim := strings.Index(xml[ini:], "</"+tag+">")
	if fim < 0 {
		return ""
	}
	return xml[ini : ini+fim]
}
