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
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// ---------------------------------------------------------------------------
// GerarXML -- round-trip e fidelidade ao TNFGasXmlWriter
// ---------------------------------------------------------------------------

// TestGerarXML_RoundTripEstavel e o teste mais forte da geracao: o XML
// gerado a partir da fixture completa, relido e regerado, tem que ser
// BYTE A BYTE igual ao da primeira geracao. Qualquer campo gerado que o
// leitor nao le (ou lido em no errado) quebra aqui.
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

func TestGerarXML_ChaveEId(t *testing.T) {
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	// a fixture ja tem cNF != 0, entao a chave regerada e a mesma
	if n.InfNFGas.ID != "NFGas"+chaveTeste {
		t.Fatalf("ID = %q, esperado NFGas%s", n.InfNFGas.ID, chaveTeste)
	}
	if !strings.Contains(xml, `<infNFGas Id="NFGas`+chaveTeste+`" versao="1.00">`) {
		t.Fatalf("infNFGas sem Id/versao esperados: %.200s", xml)
	}
	if !strings.HasPrefix(xml, `<NFGas xmlns="`+Namespace+`">`) {
		t.Fatalf("raiz inesperada: %.100s", xml)
	}
	if n.Ide.CDV != 8 {
		t.Fatalf("cDV = %d, esperado 8", n.Ide.CDV)
	}
}

func TestGerarXML_OrdemDaIde(t *testing.T) {
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	ini := strings.Index(xml, "<ide>")
	fim := strings.Index(xml, "</ide>")
	ide := xml[ini : fim+6]
	// ordem exata do Gerar_Ide; finNFGas e tpFat por CODIGO, nao ordinal
	ordem := []string{"<cUF>", "<tpAmb>", "<mod>76</mod>", "<serie>", "<nNF>",
		"<cNF>", "<cDV>", "<dhEmi>", "<tpEmis>", "<nSiteAutoriz>", "<cMunFG>",
		"<finNFGas>", "<tpFat>", "<verProc>"}
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

func TestGerarXML_CodigosNaoOrdinais(t *testing.T) {
	// DIVERGENCIA 2 do writer: o ACBr grava ordinais nesses campos; aqui
	// tem que sair o CODIGO do leiaute.
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<finNFGas>"+n.Ide.FinNFGas.String()+"</finNFGas>") {
		t.Fatalf("finNFGas nao esta em codigo do leiaute: %s", trecho(xml, "finNFGas"))
	}
	if !strings.Contains(xml, "<indOrigemQtd>") {
		t.Fatal("indOrigemQtd ausente")
	}
	// nenhum enum sai vazio (ordinal invalido produziria vazio no String())
	for _, tag := range []string{"<finNFGas></finNFGas>", "<tpFat></tpFat>", "<indOrigemQtd></indOrigemQtd>"} {
		if strings.Contains(xml, tag) {
			t.Fatalf("codigo de enum vazio: %s", tag)
		}
	}
}

func TestGerarXML_DecimaisComCasasFixas(t *testing.T) {
	n := lerCompleta(t)
	// o det 1 da fixture tem gNormal E gAgregadora preenchidos; na GERACAO
	// o gAgregadora vence e o gNormal (com gTarif) nao sai -- fidelidade ao
	// Gerar_det. Para exercitar as precisoes do gNormal, limpa o agregador.
	n.Det[0].GAgregadora = GAgregadora{}
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	// qUnidContrat tcDe6, vTarifAplic tcDe8, qFaturada tcDe4, vItem/vProd
	// tcDe8 (precisoes POR CAMPO do XmlWriter.pas)
	casos := map[string]int{
		"qUnidContrat": 6,
		"vTarifAplic":  8,
		"qFaturada":    4,
		"vItem":        8,
		"vProd":        8,
		"vNF":          2,
	}
	for tag, casas := range casos {
		v := trecho(xml, tag)
		if v == "" {
			t.Errorf("tag %s ausente", tag)
			continue
		}
		p := strings.Index(v, ".")
		if p < 0 || len(v)-p-1 != casas {
			t.Errorf("%s = %q, esperadas %d casas decimais", tag, v, casas)
		}
	}
}

func TestGerarXML_GruposCondicionais(t *testing.T) {
	n := lerCompleta(t)
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	// infAdic e SEMPRE gerado, mesmo sem conteudo relevante
	if !strings.Contains(xml, "<infAdic>") && !strings.Contains(xml, "<infAdic/>") {
		t.Fatal("infAdic deveria ser gerado sempre")
	}
	// nContrat do gMedicao sai sempre com 2 digitos (FormatFloat '00'
	// antes do AddNode opcional -- omissao do ACBr replicada)
	if strings.Contains(xml, "<gMedicao>") && !strings.Contains(xml, "<nContrat>") {
		t.Fatal("nContrat do gMedicao deveria sair sempre (\"00\" nunca e vazio)")
	}
}

func TestGerarXML_ICMS70SemDuplicataDeDesoneracao(t *testing.T) {
	n := notaMinima()
	det := &n.Det[0].GNormal
	det.Imposto.ICMS = ICMS{
		CST: pcn.CST70, ModBC: 0, PRedBC: 10, VBC: 100, PICMS: 17, VICMS: 15.3,
		ModBCST: 0, VBCST: 0, PICMSST: 0, VICMSST: 0,
		VICMSDeson: 5.5, MotDesICMS: MdiOutros, IndDeduzDeson: pcn.TieSim,
	}
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	bloco := xml[strings.Index(xml, "<ICMS70>"):strings.Index(xml, "</ICMS70>")]
	if got := strings.Count(bloco, "<vICMSDeson>"); got != 1 {
		t.Fatalf("vICMSDeson deveria aparecer 1 vez no ICMS70 (ACBr duplica), apareceu %d:\n%s", got, bloco)
	}
	if !strings.Contains(xml, "<motDesICMS>"+MdiOutros.String()+"</motDesICMS>") {
		t.Fatalf("motDesICMS deveria sair pelo CODIGO (ACBr grava ordinal): %s", trecho(xml, "motDesICMS"))
	}
	if !strings.Contains(xml, "<indDeduzDeson>1</indDeduzDeson>") {
		t.Fatalf("indDeduzDeson ausente: %s", xml)
	}
}

func TestGerarXML_ImpostoIndSemCST(t *testing.T) {
	n := notaMinima()
	n.Det[0].GNormal.Imposto.IndSemCST = pcn.TieSim
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<imposto><indSemCST>1</indSemCST>") {
		t.Fatalf("indSemCST=1 deveria substituir orig+ICMS: %s", trecho(xml, "imposto"))
	}
	if strings.Contains(xml, "<orig>") || strings.Contains(xml, "<ICMS00>") {
		t.Fatal("com indSemCST nao pode haver orig/ICMSxx")
	}
}

func TestGerarXML_CSTForaDoLeiauteDerrubaImposto(t *testing.T) {
	// DIVERGENCIA 4: o Delphi estouraria com AV; aqui o grupo imposto sai
	// nil e o item fica sem imposto.
	n := notaMinima()
	n.Det[0].GNormal.Imposto.ICMS.CST = pcn.CST30
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(xml, "<imposto>") {
		t.Fatalf("CST 30 nao e do leiaute; imposto nao deveria ser gerado: %s", trecho(xml, "imposto"))
	}
}

func TestGerarXMLProc_RaizConformeXSD(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "nfgas_proc.xml"))
	if err != nil || len(notas) == 0 {
		t.Fatalf("ler nfgas_proc.xml: %v", err)
	}
	n := notas[0].NFGas

	xml, err := GerarXMLProc(n)
	if err != nil {
		t.Fatalf("GerarXMLProc: %v", err)
	}
	// DIVERGENCIA 1: raiz nfgasProc (o ACBr gera NFGasProc, contra o XSD)
	if !strings.HasPrefix(xml, `<nfgasProc versao="1.00" xmlns="`+Namespace+`">`) {
		t.Fatalf("raiz do proc inesperada: %.120s", xml)
	}
	for _, tag := range []string{"<protNFGas versao=", "<infProt>", "<nProt>", "<cStat>", "<dhRecbto>"} {
		if !strings.Contains(xml, tag) {
			t.Fatalf("proc sem %s:\n%s", tag, xml)
		}
	}
	// round-trip do proc: relido, chave e protocolo se preservam
	relido, err := LerBytes([]byte(xml))
	if err != nil {
		t.Fatalf("reler proc: %v", err)
	}
	if relido.ProcNFGas.NProt != n.ProcNFGas.NProt {
		t.Fatalf("protocolo perdido: %q != %q", relido.ProcNFGas.NProt, n.ProcNFGas.NProt)
	}

	// sem protocolo, GerarXMLProc recusa
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
	if n.Ide.CNF == 0 || n.Ide.CNF == n.Ide.NNF {
		t.Fatalf("cNF sorteado invalido: %d", n.Ide.CNF)
	}
	if n.Ide.CNF > 9999999 {
		t.Fatalf("cNF com mais de 7 digitos: %d (posicao 36 da chave e o nSiteAutoriz)", n.Ide.CNF)
	}
}

// ---------------------------------------------------------------------------
// Evento
// ---------------------------------------------------------------------------

func TestGerarXMLEvento_Cancelamento(t *testing.T) {
	e := &EventoNFGas{
		InfEvento: InfEvento{
			TpAmb:      pcn.TaHomologacao,
			ChNFGas:    chaveTeste,
			DhEvento:   time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
			TpEvento:   TeCancelamento,
			NSeqEvento: 1,
			DetEvento:  DetEvento{NProt: "335260000000001", XJust: "cancelamento em teste unitario"},
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
		`<eventoNFGas xmlns="` + Namespace + `" versao="1.00">`,
		`<infEvento Id="` + wantID + `">`,
		"<cOrgao>35</cOrgao>",         // extraido da chave (getcOrgao)
		"<CNPJ>11222333000181</CNPJ>", // extraido da chave quando vazio
		"<tpEvento>110111</tpEvento>",
		"<nSeqEvento>1</nSeqEvento>",
		`<detEvento versaoEvento="1.00"><evCancNFGas>`,
		"<descEvento>Cancelamento</descEvento>",
		"<nProt>335260000000001</nProt>",
	} {
		if !strings.Contains(xml, m) {
			t.Fatalf("evento sem %q:\n%s", m, xml)
		}
	}
	// round-trip com o leitor de evento existente
	ret, err := LerEventoString(xml)
	if err != nil {
		t.Fatalf("reler evento: %v", err)
	}
	if ret.Evento.InfEvento.ChNFGas != chaveTeste || ret.Evento.InfEvento.DetEvento.XJust == "" {
		t.Fatalf("round-trip do evento perdeu dados: %+v", ret.Evento.InfEvento)
	}
}

func TestGerarXMLEvento_TipoNaoSuportado(t *testing.T) {
	e := &EventoNFGas{}
	if _, err := GerarXMLEvento(e); err == nil {
		t.Fatal("tipo de evento nao implementado deveria dar erro (raise do original)")
	}
}

// procEventoTeste monta um retorno de cancelamento completo, nos moldes de
// testdata/proc_evento_cancelamento.xml.
func procEventoTeste() *RetEventoNFGas {
	return &RetEventoNFGas{
		Versao:    "1.00",
		TemEvento: true,
		Evento: EventoNFGas{
			InfEvento: InfEvento{
				TpAmb:      pcn.TaHomologacao,
				ChNFGas:    chaveTeste,
				DhEvento:   time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC),
				TpEvento:   TeCancelamento,
				NSeqEvento: 1,
				DetEvento:  DetEvento{NProt: "335260000000001", XJust: "cancelamento em teste unitario"},
			},
		},
		RetInfEvento: RetInfEvento{
			TpAmb:       pcn.TaHomologacao,
			VerAplic:    "SP_NFGAS_1.0.0",
			COrgao:      35,
			CStat:       135,
			XMotivo:     "Evento registrado e vinculado a NFGas",
			ChNFGas:     chaveTeste,
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
		`<procEventoNFGas versao="1.00" xmlns="` + Namespace + `">`,
		`<eventoNFGas versao="1.00">`, // namespace so na raiz, nao repetido aqui
		"<evCancNFGas>",
		`<retEventoNFGas versao="1.00">`,
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
// retEventoNFGas. O leitor usa FindAnyNs campo a campo e e insensivel a
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

	ret := trecho(xml, "retEventoNFGas")
	if ret == "" {
		// a tag tem atributo, entao trecho() nao a encontra pelo nome puro
		ini := strings.Index(xml, "<retEventoNFGas")
		if ini < 0 {
			t.Fatal("retEventoNFGas ausente")
		}
		ret = xml[ini:]
	}

	ordem := []string{
		"<tpAmb>", "<verAplic>", "<cOrgao>", "<cStat>", "<xMotivo>",
		"<chNFGas>", "<tpEvento>", "<xEvento>", "<nSeqEvento>",
		"<CNPJDest>", "<emailDest>", "<cOrgaoAutor>", "<dhRegEvento>",
		"<nProt>",
	}
	anterior := -1
	for _, tag := range ordem {
		pos := strings.Index(ret, tag)
		if pos < 0 {
			t.Fatalf("retEventoNFGas sem %s:\n%s", tag, ret)
		}
		if pos <= anterior {
			t.Fatalf("%s fora de ordem no retEventoNFGas (posicao %d, anterior %d):\n%s",
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
		XMotivo: "Evento registrado e vinculado a NFGas",
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
		"<chNFGas>" + chaveTeste + "</chNFGas>",
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
	semEvento := &RetEventoNFGas{RetInfEvento: RetInfEvento{NProt: "1"}}
	if _, err := GerarXMLProcEvento(semEvento); !errors.Is(err, ErrEventoAusente) {
		t.Fatalf("sem evento deveria dar ErrEventoAusente, veio %v", err)
	}
	semProt := &RetEventoNFGas{
		TemEvento: true,
		Evento: EventoNFGas{InfEvento: InfEvento{
			TpEvento: TeCancelamento, ChNFGas: chaveTeste, NSeqEvento: 1,
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
	want := "https://dfe-portal.svrs.rs.gov.br/nfgas/qrCode?chNFGas=" + n.ChaveAcesso() + "&tpAmb=2"
	if url != want {
		t.Fatalf("QRCode = %q, esperado %q", url, want)
	}

	// emissao offline exige o sign
	n.Ide.TpEmis = pcn.TeOffLine
	url, err = GerarQRCode(n, func(chave string) (string, error) { return "ASSINATURA", nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(url, "&sign=ASSINATURA") {
		t.Fatalf("QRCode offline sem sign: %q", url)
	}
	if _, err := GerarQRCode(n, nil); !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Fatalf("offline sem certificado deveria falhar, veio %v", err)
	}
}

func TestURLServico(t *testing.T) {
	url, err := URLServico("SP", pcn.TaHomologacao, ServicoStatusServico)
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://nfgas-homologacao.svrs.rs.gov.br/WS/NFGasStatusServico/NFGasStatusServico.asmx" {
		t.Fatalf("URL homologacao = %q", url)
	}
	url, _ = URLServico("SP", pcn.TaProducao, ServicoRecepcao)
	if url != "https://nfgas.svrs.rs.gov.br/WS/NFGasRecepcao/NFGasRecepcao.asmx" {
		t.Fatalf("URL producao = %q", url)
	}
	// MA e PA: SVAN sem URL (lacuna herdada do ACBrNFGasServicos.ini)
	for _, uf := range []string{"MA", "PA"} {
		if _, err := URLServico(uf, pcn.TaProducao, ServicoRecepcao); !errors.Is(err, ErrSemURL) {
			t.Fatalf("UF %s deveria dar ErrSemURL, veio %v", uf, err)
		}
	}
	if SoapAction(ServicoRecepcao) != Namespace+"/wsdl/NFGasRecepcao/nfgasRecepcao" {
		t.Fatalf("SoapAction recepcao = %q", SoapAction(ServicoRecepcao))
	}
	if SoapAction(ServicoStatusServico) != Namespace+"/wsdl/NFGasStatusServico/NFGasStatusServicoNF" {
		t.Fatalf("SoapAction status = %q", SoapAction(ServicoStatusServico))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// notaMinima monta uma NFGas valida minima para testes de geracao.
func notaMinima() *NFGas {
	return &NFGas{
		InfNFGas: InfNFGas{Versao: 1.00},
		Ide: Ide{
			CUF: 35, TpAmb: pcn.TaHomologacao, Modelo: 76, Serie: 1, NNF: 1,
			CNF: 18, DhEmi: time.Date(2026, 3, 5, 10, 0, 0, 0, time.UTC),
			TpEmis: pcn.TeNormal, NSiteAutoriz: Sa1, CMunFG: 3550308,
			FinNFGas: FnNormal, TpFat: TfNormal, VerProc: "OpenFiscalBR",
		},
		Emit: Emit{
			CNPJ: "11222333000181", IE: "111111111111", XNome: "Distribuidora Teste",
			EnderEmit: Endereco{XLgr: "Rua A", Nro: "1", XBairro: "Centro",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1000000, UF: "SP"},
		},
		Dest: Dest{
			XNome: "Cliente Teste", CNPJCPF: "52998224725",
			EnderDest: Endereco{XLgr: "Rua B", Nro: "2", XBairro: "Centro",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1000000, UF: "SP"},
		},
		Det: []Det{{
			NItem: 1,
			GNormal: GNormal{
				Prod: Prod{
					IndOrigemQtd: IoMedia, CProd: "GAS", XProd: "Gas canalizado",
					UMed: Umim3, QFaturada: 10, VItem: 2.5, VProd: 25,
				},
				Imposto: Imposto{
					ICMS: ICMS{CST: pcn.CST00, VBC: 25, PICMS: 18, VICMS: 4.5},
				},
			},
		}},
		Total: Total{VProd: 25, VBC: 25, VICMS: 4.5, VNF: 25},
	}
}

// trecho devolve o conteudo da primeira ocorrencia da tag, para mensagens
// de erro e asserts de formato.
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

func TestGerarXML_CompetenciaEDhContZerosViramTagVazia(t *testing.T) {
	// DIVERGENCIA 5: FormatDateTime de data zero no Delphi produz "189912"
	// (gFat/gCons) e "1899-12-30T..." (dhCont com xJust preenchido); aqui a
	// data zero vira tag vazia.
	n := notaMinima()
	n.GFat = GFat{
		DVencFat:     time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		DProxLeitura: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC),
		CodBarras:    "84670000000000000011222333000181000000000001",
		// CompetFat fica zero de proposito
	}
	n.Ide.XJust = "justificativa de contingencia com quinze+" // dhCont zero
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<CompetFat/>") {
		t.Fatalf("CompetFat zero deveria gerar tag vazia: %.200s", xml[strings.Index(xml, "<gFat>"):])
	}
	if !strings.Contains(xml, "<dhCont/>") {
		t.Fatalf("dhCont zero (com xJust) deveria gerar tag vazia: %s", trecho(xml, "ide"))
	}
	if strings.Contains(xml, "1899") {
		t.Fatalf("data zero vazou como 1899: %s", xml)
	}
}

func TestGerarXML_SignatureParcialNaoEmbute(t *testing.T) {
	// taSomenteSeAssinada exige DigestValue+SignatureValue+X509Certificate;
	// um struct so com URI nao pode virar bloco Signature quebrado.
	n := notaMinima()
	n.Signature.URI = "#NFGas" + chaveTeste
	xml, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(xml, "<Signature") {
		t.Fatalf("Signature parcial nao deveria ser embutida: %s", xml)
	}
}
