// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import (
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do EPEC e da conciliacao financeira.
//
// Sao as duas estruturas mais RAMIFICADAS do gerador de eventos, e as duas
// unicas cujo erro produz grupo vazio ou campo no nivel errado -- o defeito
// dominante do projeto. O conferir-eventos-dfe.py nao as alcanca: ele
// compara conjunto e ordem de tags, e por construcao nao distingue os ramos
// de uma condicional.

// eventoEPEC monta um EPEC com a chave informada.
func eventoEPEC(chave string) *EventoNFe {
	e := eventoBase(pcn.TeEPECNFe)
	e.Evento[0].InfEvento.ChNFe = chave
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.DhEmi = time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)
	d.TpNF = pcn.TnSaida
	d.IE = "1234567890"
	d.VNF, d.VICMS, d.VST = 1000, 180, 50
	d.Dest = Destinatario{UF: "MG", CNPJCPF: "99999999000191", IE: "0987654321"}
	return e
}

// TestGerarDetEventoEPEC_NFeTemVSTDentroDoDest trava o ramo do modelo 55.
//
// Na NF-e os tres valores vivem DENTRO do grupo dest, e o vST so existe
// aqui (Gerar_DestNFe, :372-425).
func TestGerarDetEventoEPEC_NFeTemVSTDentroDoDest(t *testing.T) {
	xml := gerarEventoDeTeste(t, eventoEPEC(chaveDeMinas))

	dest := recortarEntre(t, xml, "<dest>", "</dest>")
	for _, tag := range []string{"<UF>MG</UF>", "<CNPJ>99999999000191</CNPJ>",
		"<IE>0987654321</IE>", "<vNF>1000.00</vNF>", "<vICMS>180.00</vICMS>",
		"<vST>50.00</vST>"} {
		if !strings.Contains(dest, tag) {
			t.Errorf("faltou %s dentro do dest: %s", tag, dest)
		}
	}

	conferirOrdem(t, xml, "<descEvento>", "<cOrgaoAutor>", "<tpAutor>",
		"<verAplic>", "<dhEmi>", "<tpNF>", "<IE>", "<dest>")
}

// TestGerarDetEventoEPEC_NFCeTemValoresFORADoDestESemVST trava o ramo do
// modelo 65, que e o oposto do anterior em DUAS coisas.
//
// O comentario do ACBr diz tudo (:650): "No EPEC da NFC-e segundo o schema
// as TAGs vNF e vICMS estao fora do grupo dest e nao tem a TAG vST".
func TestGerarDetEventoEPEC_NFCeTemValoresFORADoDestESemVST(t *testing.T) {
	xml := gerarEventoDeTeste(t, eventoEPEC(chaveNFCeDeMinas))

	if strings.Contains(xml, "<vST>") {
		t.Errorf("o EPEC de NFC-e NAO tem vST: %s", xml)
	}

	dest := recortarEntre(t, xml, "<dest>", "</dest>")
	if strings.Contains(dest, "<vNF>") || strings.Contains(dest, "<vICMS>") {
		t.Errorf("na NFC-e os valores ficam FORA do dest: %s", dest)
	}
	// O dest da NFC-e tambem nao leva IE.
	if strings.Contains(dest, "<IE>") {
		t.Errorf("o dest da NFC-e nao tem IE: %s", dest)
	}

	// E eles estao no detEvento, depois do dest.
	if !strings.Contains(xml, "<vNF>1000.00</vNF>") ||
		!strings.Contains(xml, "<vICMS>180.00</vICMS>") {
		t.Errorf("os valores sumiram do detEvento: %s", xml)
	}
	conferirOrdem(t, xml, "<IE>", "<dest>", "</dest>", "<vNF>", "<vICMS>")
}

// TestGerarDetEventoEPEC_DestDaNFCeEhOPCIONAL trava a condicao COMPOSTA de
// :332.
//
// O grupo inteiro so sai quando ha UF E (documento OU idEstrangeiro).
// Perder a composicao produz um <dest> vazio no lugar de nenhum -- e o
// schema recusa os dois jeitos errados.
func TestGerarDetEventoEPEC_DestDaNFCeEhOPCIONAL(t *testing.T) {
	casos := []struct {
		dest      Destinatario
		temDest   bool
		descricao string
	}{
		{Destinatario{UF: "MG", CNPJCPF: "99999999000191"}, true, "UF e CNPJ"},
		{Destinatario{UF: "EX", IDEstrangeiro: "AB123"}, true, "UF e estrangeiro"},
		{Destinatario{UF: "MG"}, false, "so UF: grupo nao sai"},
		{Destinatario{CNPJCPF: "99999999000191"}, false, "so documento: grupo nao sai"},
		{Destinatario{}, false, "nada: grupo nao sai"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			e := eventoEPEC(chaveNFCeDeMinas)
			e.Evento[0].InfEvento.DetEvento.Dest = c.dest

			xml := gerarEventoDeTeste(t, e)
			if got := strings.Contains(xml, "<dest>"); got != c.temDest {
				t.Errorf("dest presente = %v, esperado %v: %s", got, c.temDest, xml)
			}
			// Na NF-e o grupo e obrigatorio e sai sempre -- o contraste.
			e = eventoEPEC(chaveDeMinas)
			e.Evento[0].InfEvento.DetEvento.Dest = c.dest
			if xml := gerarEventoDeTeste(t, e); !strings.Contains(xml, "<dest>") {
				t.Errorf("na NF-e o dest e obrigatorio: %s", xml)
			}
		})
	}
}

// TestDocumentoDoDestinatario_EstrangeiroPorIDOuPorUF trava a condicao de
// :342, que tem DUAS portas de entrada.
func TestDocumentoDoDestinatario_EstrangeiroPorIDOuPorUF(t *testing.T) {
	casos := []struct {
		dest     Destinatario
		esperado string
	}{
		{Destinatario{UF: "MG", CNPJCPF: "99999999000191"},
			"<CNPJ>99999999000191</CNPJ>"},
		{Destinatario{UF: "MG", CNPJCPF: "11122233396"},
			"<CPF>11122233396</CPF>"},
		{Destinatario{UF: "MG", IDEstrangeiro: "AB123"},
			"<idEstrangeiro>AB123</idEstrangeiro>"},
		// UF "EX" basta, mesmo sem idEstrangeiro -- e ai o elemento sai
		// vazio, como no ACBr.
		{Destinatario{UF: "EX", CNPJCPF: "99999999000191"},
			"<idEstrangeiro/>"},
	}
	for _, c := range casos {
		if got := documentoDoDestinatario(c.dest).XML(); got != c.esperado {
			t.Errorf("%+v: %q, esperado %q", c.dest, got, c.esperado)
		}
	}
}

// TestGerarDetPagEvento_DoisBlocosINDEPENDENTES trava :892 e :907.
//
// O bloco do pagador sai quando ha CNPJPag OU UFPag; o do recebedor, quando
// ha CNPJReceb OU UFReceb. Sao condicoes separadas, e emitir os dois sempre
// produz tags vazias que o schema recusa.
func TestGerarDetPagEvento_DoisBlocosINDEPENDENTES(t *testing.T) {
	base := DetPagEvento{
		TPag: FpDinheiro,
		VPag: 100,
		DPag: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
	}

	casos := []struct {
		ajustar      func(*DetPagEvento)
		temPagador   bool
		temRecebedor bool
		descricao    string
	}{
		{func(p *DetPagEvento) {}, false, false, "nenhum dos dois"},
		{func(p *DetPagEvento) { p.CNPJPag = "99999999000191" }, true, false,
			"so CNPJPag ja abre o bloco do pagador"},
		{func(p *DetPagEvento) { p.UFPag = "MG" }, true, false,
			"so UFPag ja abre o bloco do pagador"},
		{func(p *DetPagEvento) { p.CNPJReceb = "99999999000191" }, false, true,
			"so CNPJReceb abre o do recebedor, e nao o do pagador"},
		{func(p *DetPagEvento) { p.UFReceb = "SP" }, false, true,
			"so UFReceb abre o do recebedor"},
		{func(p *DetPagEvento) { p.CNPJPag = "1"; p.UFReceb = "SP" }, true, true,
			"os dois, por campos diferentes"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			p := base
			c.ajustar(&p)
			xml := gerarDetPagEvento(p).XML()

			if got := strings.Contains(xml, "<UFPag"); got != c.temPagador {
				t.Errorf("bloco do pagador = %v, esperado %v: %s",
					got, c.temPagador, xml)
			}
			if got := strings.Contains(xml, "<UFReceb"); got != c.temRecebedor {
				t.Errorf("bloco do recebedor = %v, esperado %v: %s",
					got, c.temRecebedor, xml)
			}
			// Os campos do inicio saem sempre.
			if !strings.Contains(xml, "<vPag>100.00</vPag>") {
				t.Errorf("vPag deveria sair sempre: %s", xml)
			}
		})
	}
}

// TestGerarDetPagEvento_OrdemDosCampos trava a sequencia de :866-918.
func TestGerarDetPagEvento_OrdemDosCampos(t *testing.T) {
	p := DetPagEvento{
		IndPag:    IpVista,
		TPag:      FpCartaoCredito,
		XPag:      "pagamento",
		VPag:      100,
		DPag:      time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
		CNPJPag:   "99999999000191",
		UFPag:     "MG",
		CNPJIF:    "11111111000191",
		TBand:     BcVisa,
		CAut:      "123456",
		CNPJReceb: "22222222000191",
		UFReceb:   "SP",
	}
	conferirOrdem(t, gerarDetPagEvento(p).XML(),
		"<indPag>", "<tPag>", "<xPag>", "<vPag>", "<dPag>",
		"<CNPJPag>", "<UFPag>", "<CNPJIF>", "<tBand>", "<cAut>",
		"<CNPJReceb>", "<UFReceb>")
}

// TestLerDetEvento_EPECdeNFCeRecuperaVNFeVICMS trava o ACRESCIMO ao ACBr.
//
// O Ler_DetEvento do original so le os tres valores DENTRO do Ler_Dest
// (:195), entao o EPEC de NFC-e -- onde eles ficam fora, e onde o dest
// pode nem existir -- perde os dois no round-trip.
func TestLerDetEvento_EPECdeNFCeRecuperaVNFeVICMS(t *testing.T) {
	e := eventoEPEC(chaveNFCeDeMinas)
	e.Evento[0].InfEvento.DetEvento.Dest = Destinatario{} // sem dest
	xml := gerarEventoDeTeste(t, e)

	if strings.Contains(xml, "<dest>") {
		t.Fatalf("o caso de teste exige um EPEC SEM dest: %s", xml)
	}

	r, err := LerRetEventoNFe(xml)
	if err != nil {
		t.Fatal(err)
	}
	d := r.InfEvento.DetEvento
	if d.VNF != 1000 || d.VICMS != 180 {
		t.Errorf("vNF/vICMS lidos = %v / %v, esperado 1000 / 180", d.VNF, d.VICMS)
	}
}

// TestLerDetEvento_DestVenceONivelDeCima trava a precedencia.
//
// Na NF-e os valores vivem dentro do dest; o leitor le os dois niveis, e o
// de dentro tem de vencer -- senao um documento que trouxesse os dois
// (nenhum traz, mas o leitor nao sabe) devolveria o valor errado.
func TestLerDetEvento_DestVenceONivelDeCima(t *testing.T) {
	xml := `<evento versao="1.00" xmlns="` + namespaceNFe + `">` +
		`<infEvento Id="X"><cOrgao>31</cOrgao><tpAmb>2</tpAmb>` +
		`<CNPJ>99999999000191</CNPJ><chNFe>` + chaveDeMinas + `</chNFe>` +
		`<dhEvento>2026-01-20T18:45:00-03:00</dhEvento>` +
		`<tpEvento>110140</tpEvento><nSeqEvento>1</nSeqEvento>` +
		`<verEvento>1.00</verEvento><detEvento versao="1.00">` +
		`<descEvento>EPEC</descEvento>` +
		`<vNF>9.99</vNF>` + // o de FORA, que nao deveria vencer
		`<dest><UF>MG</UF><CNPJ>99999999000191</CNPJ>` +
		`<vNF>1000.00</vNF><vICMS>180.00</vICMS><vST>50.00</vST></dest>` +
		`</detEvento></infEvento></evento>`

	r, err := LerRetEventoNFe(xml)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.InfEvento.DetEvento.VNF; got != 1000 {
		t.Errorf("vNF = %v; o valor de DENTRO do dest tem de vencer", got)
	}
}
