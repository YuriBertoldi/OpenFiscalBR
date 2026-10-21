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
)

// Testes dos leitores do fluxo assincrono.

const chaveLote1 = "35260199999999000191550010000004561123456791"
const chaveLote2 = "35260199999999000191550010000004571123456792"

// montarRetConsReci monta um retorno de consulta de recibo com os
// protocolos pedidos, na ordem informada.
func montarRetConsReci(cStat int, protocolos ...string) string {
	var b strings.Builder
	b.WriteString(`<retConsReciNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">`)
	b.WriteString(`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic>`)
	b.WriteString(`<nRec>351000000000001</nRec>`)
	b.WriteString(`<cStat>` + itoa(cStat) + `</cStat><xMotivo>motivo</xMotivo>`)
	b.WriteString(`<cUF>35</cUF><cMsg>0</cMsg><xMsg></xMsg>`)
	for i, chave := range protocolos {
		b.WriteString(`<protNFe versao="4.00"><infProt Id="ID` + chave + `">` +
			`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic>` +
			`<chNFe>` + chave + `</chNFe>` +
			`<dhRecbto>2026-01-20T18:45:0` + itoa(i) + `-03:00</dhRecbto>` +
			`<nProt>13526000000000` + itoa(i) + `</nProt>` +
			`<digVal>digest` + itoa(i) + `</digVal>` +
			`<cStat>100</cStat><xMotivo>Autorizado o uso da NF-e</xMotivo>` +
			`</infProt></protNFe>`)
	}
	b.WriteString(`</retConsReciNFe>`)
	return b.String()
}

// TestLerRetEnviNFe_ReciboVemDoInfRec trava a leitura do grupo infRec.
//
// O nRec do retEnviNFe NAO esta na raiz: ele vive dentro de infRec, ao lado
// de dhRecbto e tMed (RetEnvio.pas:139). Um porte que o procure na raiz
// devolve recibo vazio e o fluxo assincrono para ali, sem erro nenhum.
func TestLerRetEnviNFe_ReciboVemDoInfRec(t *testing.T) {
	xml := `<retEnviNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic><cStat>103</cStat>` +
		`<xMotivo>Lote recebido com sucesso</xMotivo><cUF>35</cUF>` +
		`<infRec><nRec>351000000000001</nRec>` +
		`<dhRecbto>2026-01-20T18:45:00-03:00</dhRecbto><tMed>2</tMed></infRec>` +
		`</retEnviNFe>`

	r, err := LerRetEnviNFe(xml)
	if err != nil {
		t.Fatal(err)
	}

	if r.InfRec.NRec != "351000000000001" {
		t.Errorf("nRec = %q, esperado 351000000000001", r.InfRec.NRec)
	}
	if r.InfRec.TMed != 2 {
		t.Errorf("tMed = %d, esperado 2", r.InfRec.TMed)
	}
	if r.InfRec.DhRecbto.IsZero() {
		t.Error("dhRecbto do infRec nao foi lido")
	}
	if !r.LoteRecebido() {
		t.Error("cStat 103 e lote recebido")
	}
}

// TestRetEnviNFe_SoO103EhLoteRecebido trava o criterio de sucesso do envio
// assincrono (ACBrNFeWebServices.pas:1317).
//
// O 104 e tentador porque e sucesso na OUTRA etapa -- a consulta do recibo
// --, e e justamente por isso que ele entra no teste.
func TestRetEnviNFe_SoO103EhLoteRecebido(t *testing.T) {
	for _, c := range []struct {
		cStat    int
		recebido bool
	}{
		{103, true},
		{104, false},
		{108, false},
		{225, false},
		{0, false},
	} {
		r := &RetEnviNFe{CStat: c.cStat}
		if r.LoteRecebido() != c.recebido {
			t.Errorf("cStat %d: LoteRecebido() = %v, esperado %v",
				c.cStat, r.LoteRecebido(), c.recebido)
		}
	}
}

// TestRetConsReciNFe_DistingueProcessadoDeEmProcessamento trava os dois
// codigos que governam o laco de consulta.
//
// 104 encerra; 105 manda consultar de novo. Tratar os dois como sucesso faz
// a aplicacao concluir o envio sem protocolo nenhum.
func TestRetConsReciNFe_DistingueProcessadoDeEmProcessamento(t *testing.T) {
	for _, c := range []struct {
		cStat       int
		processado  bool
		processando bool
	}{
		{104, true, false},
		{105, false, true},
		{103, false, false},
		{217, false, false},
	} {
		r := &RetConsReciNFe{CStat: c.cStat}
		if r.LoteProcessado() != c.processado || r.LoteEmProcessamento() != c.processando {
			t.Errorf("cStat %d: processado=%v processando=%v, esperado %v/%v",
				c.cStat, r.LoteProcessado(), r.LoteEmProcessamento(),
				c.processado, c.processando)
		}
	}
}

// TestLerRetConsReciNFe_LeTodosOsProtocolos trava a leitura da colecao.
//
// O lote assincrono traz UM protNFe por nota, e cada um precisa chegar
// inteiro -- com o XML cru, que e de onde sai o nfeProc de cada nota.
func TestLerRetConsReciNFe_LeTodosOsProtocolos(t *testing.T) {
	r, err := LerRetConsReciNFe(montarRetConsReci(104, chaveLote1, chaveLote2))
	if err != nil {
		t.Fatal(err)
	}

	if len(r.ProtNFe) != 2 {
		t.Fatalf("protocolos lidos = %d, esperado 2", len(r.ProtNFe))
	}
	if r.ProtNFe[0].ChNFe == r.ProtNFe[1].ChNFe {
		t.Fatal("os dois protocolos saíram com a mesma chave")
	}
	for i, p := range r.ProtNFe {
		if p.NProt == "" {
			t.Errorf("protocolo %d sem nProt", i)
		}
		if p.DigVal == "" {
			t.Errorf("protocolo %d sem digVal -- e dele que sai a conferencia do digest", i)
		}
		if !strings.HasPrefix(p.XMLProtNFe, "<protNFe") {
			t.Errorf("protocolo %d: XMLProtNFe nao e o elemento inteiro: %q", i, p.XMLProtNFe)
		}
		if !strings.Contains(p.XMLProtNFe, p.NProt) {
			t.Errorf("protocolo %d: o XML cru nao traz o proprio nProt", i)
		}
	}
}

// TestLerRetConsReciNFe_NaoFiltraProtocoloPorCStat trava a DIVERGENCIA em
// relacao ao retConsSitNFe.
//
// O leitor da consulta de situacao so le o protNFe quando o cStat externo o
// admite; este le sempre (RetConsReciDFe.pas:205, sem guard). A diferenca e
// do ACBr, e replica-la importa: aqui o cStat externo fala do LOTE, e um
// lote processado com nota rejeitada traz protocolo de rejeicao que o
// consumidor precisa ver.
func TestLerRetConsReciNFe_NaoFiltraProtocoloPorCStat(t *testing.T) {
	// 105 -- lote AINDA em processamento -- com protocolo junto.
	r, err := LerRetConsReciNFe(montarRetConsReci(105, chaveLote1))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ProtNFe) != 1 {
		t.Fatalf("o protocolo deveria ser lido mesmo com cStat 105: lidos %d", len(r.ProtNFe))
	}

	// E com um cStat que o retConsSitNFe recusaria.
	r, err = LerRetConsReciNFe(montarRetConsReci(217, chaveLote1))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ProtNFe) != 1 {
		t.Fatalf("este leitor nao tem guard de cStat: lidos %d", len(r.ProtNFe))
	}
}

// TestRetConsReciNFe_ProtocoloDaChave trava a busca por CHAVE.
//
// A SEFAZ nao garante que a ordem do retorno seja a do envio, e o proprio
// ACBr faz um laco duplo comparando chaves (:1645). Casar por posicao
// anexaria o protocolo de uma nota na outra -- e as duas ficariam com
// nfeProc invalido.
func TestRetConsReciNFe_ProtocoloDaChave(t *testing.T) {
	// Protocolos em ordem INVERTIDA em relacao ao que se enviaria.
	r, err := LerRetConsReciNFe(montarRetConsReci(104, chaveLote2, chaveLote1))
	if err != nil {
		t.Fatal(err)
	}

	p := r.ProtocoloDaChave(chaveLote1)
	if p == nil {
		t.Fatal("protocolo da primeira chave nao encontrado")
	}
	if p.ChNFe != chaveLote1 {
		t.Errorf("casou com a chave errada: %q", p.ChNFe)
	}

	// A forma com o literal "NFe" na frente tem de casar tambem: e assim
	// que a chave aparece no infNFe.ID.
	if q := r.ProtocoloDaChave("NFe" + chaveLote1); q == nil || q.ChNFe != chaveLote1 {
		t.Error("a chave com o literal NFe deveria casar")
	}

	if r.ProtocoloDaChave("35260199999999000191550010000009999123456799") != nil {
		t.Error("chave ausente deveria devolver nil")
	}
}
