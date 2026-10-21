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

package dfe

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Testes da chamada com cabecalho e com varios nomes de resultado.

func envelopeDeTeste(tag, conteudo string) string {
	return `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><` + tag + `>` + conteudo + `</` + tag + `>` +
		`</soap:Body></soap:Envelope>`
}

// TestMontarEnvelopeComCabecalho_HeaderSoQuandoHaConteudo trava a regra do
// DefinirEnvelopeSoap: cabecalho vazio produz envelope SEM header, nao um
// header vazio (ACBrDFeWebService.pas:331).
func TestMontarEnvelopeComCabecalho_HeaderSoQuandoHaConteudo(t *testing.T) {
	const servico = "http://exemplo/wsdl/Servico"

	com := MontarEnvelopeComCabecalho("dadosMsg", servico, "<x/>",
		"cabecMsg", "<cUF>35</cUF>")
	if !strings.Contains(com, "<soap12:Header>") {
		t.Error("com conteudo, o header deveria existir")
	}
	if !strings.Contains(com, `<cabecMsg xmlns="`+servico+`">`) {
		t.Error("o cabecalho deveria declarar o namespace do servico")
	}

	casos := []struct{ tag, conteudo, nome string }{
		{"", "<cUF>35</cUF>", "sem tag"},
		{"cabecMsg", "", "sem conteudo"},
		{"", "", "sem nada"},
	}
	for _, c := range casos {
		env := MontarEnvelopeComCabecalho("dadosMsg", servico, "<x/>", c.tag, c.conteudo)
		if strings.Contains(env, "Header") {
			t.Errorf("%s: nao deveria haver header: %s", c.nome, env)
		}
	}
}

// TestExecutar_TentaOsNomesSemReenviar e o teste central: o laco de
// TagsResultado roda sobre a resposta ja recebida.
//
// Reenviar o documento a cada nome seria errado na consulta e, na
// autorizacao, registraria duplicidade na SEFAZ.
func TestExecutar_TentaOsNomesSemReenviar(t *testing.T) {
	requisicoes := 0
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requisicoes++
		_, _ = io.ReadAll(r.Body)
		// Responde com o ULTIMO nome da lista.
		_, _ = io.WriteString(w, envelopeDeTeste("terceiroNome", "<ok/>"))
	}))
	defer servidor.Close()

	c := &ClienteSOAP{HTTPClient: servidor.Client()}
	resultado, err := c.Executar(context.Background(), Requisicao{
		URL: servidor.URL, Servico: "http://exemplo", SoapAction: "acao",
		TagDadosMsg:   "dadosMsg",
		DadosMsg:      "<x/>",
		TagsResultado: []string{"primeiroNome", "segundoNome", "terceiroNome"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resultado != "<ok/>" {
		t.Errorf("resultado = %q, esperado <ok/>", resultado)
	}
	if requisicoes != 1 {
		t.Errorf("foram %d requisicoes; os nomes devem ser tentados sobre a "+
			"MESMA resposta", requisicoes)
	}
}

// TestExecutar_FaultParaNaHora trava o curto-circuito: um soap:Fault nao
// melhora com outro nome de tag.
func TestExecutar_FaultParaNaHora(t *testing.T) {
	const fault = `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><soap:Fault><soap:Reason><soap:Text>` +
		`Rejeicao: versao invalida` +
		`</soap:Text></soap:Reason></soap:Fault></soap:Body></soap:Envelope>`

	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, fault)
	}))
	defer servidor.Close()

	c := &ClienteSOAP{HTTPClient: servidor.Client()}
	_, err := c.Executar(context.Background(), Requisicao{
		URL: servidor.URL, Servico: "http://exemplo", SoapAction: "acao",
		TagDadosMsg:   "dadosMsg",
		DadosMsg:      "<x/>",
		TagsResultado: []string{"a", "b", "c"},
	})
	if !errors.Is(err, ErrSOAPFault) {
		t.Fatalf("erro = %v; esperado envolver ErrSOAPFault", err)
	}
	// A razao do Fault tem de chegar ao consumidor: sem ela o erro e inutil.
	if !strings.Contains(err.Error(), "versao invalida") {
		t.Errorf("o erro deveria carregar a razao do Fault: %v", err)
	}
}

// TestExecutar_SemTagResultadoEhErroDeUso trava o sentinela.
func TestExecutar_SemTagResultadoEhErroDeUso(t *testing.T) {
	c := &ClienteSOAP{}
	_, err := c.Executar(context.Background(), Requisicao{URL: "http://exemplo"})
	if !errors.Is(err, ErrSemTagResultado) {
		t.Errorf("erro = %v; esperado ErrSemTagResultado", err)
	}
}

// TestExecutar_NenhumNomeCasaDevolveOUltimoErro cobre o caminho em que a
// resposta chega mas nao tem nenhum dos nomes esperados.
func TestExecutar_NenhumNomeCasaDevolveOUltimoErro(t *testing.T) {
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, envelopeDeTeste("nomeInesperado", "<ok/>"))
	}))
	defer servidor.Close()

	c := &ClienteSOAP{HTTPClient: servidor.Client()}
	_, err := c.Executar(context.Background(), Requisicao{
		URL: servidor.URL, Servico: "http://exemplo", SoapAction: "acao",
		TagDadosMsg:   "dadosMsg",
		DadosMsg:      "<x/>",
		TagsResultado: []string{"a", "b"},
	})
	if !errors.Is(err, ErrRespostaSemDados) {
		t.Errorf("erro = %v; esperado envolver ErrRespostaSemDados", err)
	}
}
