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

// Testes do leitor da consulta de situacao.
//
// O foco sao os GUARDS de cStat: o ACBr so le o protNFe e o retCancNFe
// quando o cStat da CONSULTA os admite (RetConsSit.pas:196 e :226). Ler
// incondicionalmente e o defeito dominante do port na forma classica --
// a sequencia de campos copiada e a condicional que a governava, perdida.

// montarRetConsSit monta um retorno com os blocos pedidos.
func montarRetConsSit(cStat int, comProtocolo, comCancelamento bool, eventos int) string {
	var b strings.Builder
	b.WriteString(`<retConsSitNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">`)
	b.WriteString(`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic>`)
	b.WriteString(`<cStat>` + itoa(cStat) + `</cStat><xMotivo>motivo</xMotivo>`)
	b.WriteString(`<cUF>35</cUF><chNFe>35260199999999000191550010000004561123456791</chNFe>`)

	if comProtocolo {
		b.WriteString(`<protNFe versao="4.00"><infProt><tpAmb>2</tpAmb>` +
			`<verAplic>SP_v4</verAplic>` +
			`<chNFe>35260199999999000191550010000004561123456791</chNFe>` +
			`<dhRecbto>2026-01-20T18:45:00-03:00</dhRecbto>` +
			`<nProt>135260000000001</nProt><digVal>abc</digVal>` +
			`<cStat>100</cStat><xMotivo>Autorizado o uso da NF-e</xMotivo>` +
			`</infProt></protNFe>`)
	}
	if comCancelamento {
		b.WriteString(`<retCancNFe versao="3.10"><infCanc><tpAmb>2</tpAmb>` +
			`<verAplic>SP_v4</verAplic><cStat>101</cStat>` +
			`<xMotivo>Cancelamento homologado</xMotivo><cUF>35</cUF>` +
			`<chNFe>35260199999999000191550010000004561123456791</chNFe>` +
			`<dhRecbto>2026-01-21T09:00:00-03:00</dhRecbto>` +
			`<nProt>135260000000002</nProt></infCanc></retCancNFe>`)
	}
	for i := 0; i < eventos; i++ {
		b.WriteString(`<procEventoNFe versao="1.00"><evento><infEvento>` +
			`<tpEvento>11011` + itoa(i) + `</tpEvento>` +
			`</infEvento></evento></procEventoNFe>`)
	}
	b.WriteString(`</retConsSitNFe>`)
	return b.String()
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var d []byte
	for v > 0 {
		d = append([]byte{byte('0' + v%10)}, d...)
		v /= 10
	}
	return string(d)
}

// TestLerRetConsSitNFe_GuardsDeCStat trava as duas listas que governam a
// leitura dos blocos.
//
// O caso que importa e o ultimo: uma REJEICAO que venha acompanhada de
// protNFe -- a SEFAZ-PR ja produziu isso, e o proprio ACBr comenta o caso
// (ACBrNFeWebServices.pas:2417). Sem o guard, Autorizada() responderia
// sobre um protocolo que o ACBr teria descartado.
func TestLerRetConsSitNFe_GuardsDeCStat(t *testing.T) {
	casos := []struct {
		cStat          int
		leProtocolo    bool
		leCancelamento bool
		descricao      string
	}{
		{100, true, false, "autorizada: le protocolo, nao le cancelamento"},
		{101, true, true, "cancelada: le os dois"},
		{104, true, false, "lote processado: le protocolo"},
		{150, true, false, "autorizada fora do prazo"},
		{151, true, true, "cancelada fora do prazo: le os dois"},
		{155, true, true, "cancelada por substituicao: le os dois"},
		{217, false, false, "REJEICAO: nao le nenhum dos dois"},
		{999, false, false, "codigo desconhecido: nao le nenhum"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			// O XML traz SEMPRE os dois blocos; quem decide le-los e o cStat.
			r, err := LerRetConsSitNFe(montarRetConsSit(c.cStat, true, true, 0))
			if err != nil {
				t.Fatal(err)
			}

			if (r.ProtNFe != nil) != c.leProtocolo {
				t.Errorf("protNFe lido=%v, esperado %v", r.ProtNFe != nil, c.leProtocolo)
			}
			if (r.RetCancNFe != nil) != c.leCancelamento {
				t.Errorf("retCancNFe lido=%v, esperado %v",
					r.RetCancNFe != nil, c.leCancelamento)
			}
		})
	}
}

// TestLerRetConsSitNFe_GuardaOXMLDoProtocolo trava a preservacao do XML
// cru, que e o que permite montar o nfeProc depois.
func TestLerRetConsSitNFe_GuardaOXMLDoProtocolo(t *testing.T) {
	r, err := LerRetConsSitNFe(montarRetConsSit(100, true, false, 0))
	if err != nil {
		t.Fatal(err)
	}

	if r.XMLProtNFe == "" {
		t.Fatal("o XML cru do protNFe deveria ser preservado: e dele que sai o nfeProc")
	}
	if !strings.HasPrefix(r.XMLProtNFe, "<protNFe") {
		t.Errorf("XMLProtNFe deveria comecar em <protNFe: %q", r.XMLProtNFe)
	}
	if !strings.Contains(r.XMLProtNFe, "135260000000001") {
		t.Error("o XML cru deveria conter o protocolo")
	}

	// Sem protocolo lido, nao ha XML cru.
	r, err = LerRetConsSitNFe(montarRetConsSit(217, true, false, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.XMLProtNFe != "" {
		t.Error("numa rejeicao o protNFe nao e lido, e o XML cru tampouco")
	}
}

// TestLerRetConsSitNFe_EventosVemComoXMLCru trava a decisao de devolver os
// eventos sem parse.
func TestLerRetConsSitNFe_EventosVemComoXMLCru(t *testing.T) {
	r, err := LerRetConsSitNFe(montarRetConsSit(100, true, false, 3))
	if err != nil {
		t.Fatal(err)
	}

	if len(r.ProcEventos) != 3 {
		t.Fatalf("eventos lidos = %d, esperado 3", len(r.ProcEventos))
	}
	for i, ev := range r.ProcEventos {
		if !strings.HasPrefix(ev, "<procEventoNFe") {
			t.Errorf("evento %d nao e o elemento inteiro: %q", i, ev)
		}
	}
	// Cada evento e distinto -- um porte que devolvesse o mesmo tres vezes
	// passaria na contagem.
	if r.ProcEventos[0] == r.ProcEventos[1] {
		t.Error("os eventos deveriam ser distintos")
	}
}

// TestCStatProcessadoECancelada trava os dois conjuntos canonicos do ACBr
// (ACBrNFe.pas:295-311).
//
// Os codigos que mais escapam de um porte apressado sao o 120 (autorizacao
// por DPEC) e o 135 e 155 do cancelamento.
func TestCStatProcessadoECancelada(t *testing.T) {
	for _, c := range []int{100, 120, 150} {
		if !CStatProcessado(c) {
			t.Errorf("CStatProcessado(%d) = false, esperado true", c)
		}
	}
	for _, c := range []int{101, 135, 151, 155} {
		if !CStatCancelada(c) {
			t.Errorf("CStatCancelada(%d) = false, esperado true", c)
		}
	}
	// Os dois conjuntos nao se cruzam.
	for _, c := range []int{100, 120, 150} {
		if CStatCancelada(c) {
			t.Errorf("%d esta nos dois conjuntos", c)
		}
	}
	for _, c := range []int{104, 217, 0, 999} {
		if CStatProcessado(c) || CStatCancelada(c) {
			t.Errorf("%d nao deveria estar em nenhum dos conjuntos", c)
		}
	}
}

// TestRetConsSitNFe_HelpersUsamOCStatCerto trava a distincao entre o cStat
// da CONSULTA e o do PROTOCOLO.
//
// Processada() olha o externo (ACBrNFeWebServices.pas:2397); Autorizada()
// olha o do protNFe, porque uma nota cancelada volta com cStat externo de
// cancelamento e ainda traz o protocolo da autorizacao original (:2419).
func TestRetConsSitNFe_HelpersUsamOCStatCerto(t *testing.T) {
	// Cancelada: o externo diz 101, o protocolo diz 100.
	r, err := LerRetConsSitNFe(montarRetConsSit(101, true, true, 0))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Processada() {
		t.Error("cStat 101 esta em CStatCancelada, logo a consulta foi processada")
	}
	if !r.Autorizada() {
		t.Error("o protNFe traz cStat 100: a nota FOI autorizada, e depois cancelada")
	}
	if !r.Cancelada() {
		t.Error("o retCancNFe traz cStat 101")
	}

	// Rejeicao: nem processada, nem autorizada, nem cancelada.
	r, err = LerRetConsSitNFe(montarRetConsSit(217, true, true, 0))
	if err != nil {
		t.Fatal(err)
	}
	if r.Processada() || r.Autorizada() || r.Cancelada() {
		t.Errorf("rejeicao 217 nao e nenhuma das tres: %v %v %v",
			r.Processada(), r.Autorizada(), r.Cancelada())
	}
}
