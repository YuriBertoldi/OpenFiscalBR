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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
)

func certificadoTeste(t *testing.T) *dfe.Certificado {
	t.Helper()
	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "EMISSOR TESTE:11222333000181"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &chave.PublicKey, chave)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	return &dfe.Certificado{Cert: cert, Chave: chave}
}

func servidorSOAP(t *testing.T, resultado string, corpoRecebido *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		if corpoRecebido != nil {
			*corpoRecebido = string(corpo)
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
			`<soap:Body><nfagResultMsg xmlns="http://svc">` + resultado +
			`</nfagResultMsg></soap:Body></soap:Envelope>`))
	}))
}

func componenteTeste(t *testing.T, srv *httptest.Server) *Componente {
	t.Helper()
	c := NovoComponente()
	c.Configuracoes.UF = "SP"
	c.Configuracoes.Certificado = certificadoTeste(t)
	c.Configuracoes.URLs = map[Servico]string{
		ServicoRecepcao:       srv.URL,
		ServicoRecepcaoEvento: srv.URL,
		ServicoConsulta:       srv.URL,
		ServicoStatusServico:  srv.URL,
	}
	return c
}

// dadosMsg extrai o conteudo de <nfagDadosMsg ...> do envelope capturado.
func dadosMsg(corpo string) string {
	ini := strings.Index(corpo, "<nfagDadosMsg")
	if ini < 0 {
		return ""
	}
	ini = strings.Index(corpo[ini:], ">") + ini + 1
	fim := strings.Index(corpo, "</nfagDadosMsg>")
	if fim < 0 || fim < ini {
		return ""
	}
	return corpo[ini:fim]
}

// ---------------------------------------------------------------------------
// PrepararEnvio -- geracao + assinatura verificavel
// ---------------------------------------------------------------------------

func TestPrepararEnvioAssinaVerificavel(t *testing.T) {
	c := NovoComponente()
	c.Configuracoes.Certificado = certificadoTeste(t)

	nota := &NotaFiscal{NFAg: notaMinima()}
	assinado, err := c.PrepararEnvio(nota)
	if err != nil {
		t.Fatalf("PrepararEnvio: %v", err)
	}
	if err := dfe.VerificarAssinatura(assinado); err != nil {
		t.Fatalf("assinatura gerada nao verifica: %v", err)
	}
	// a Reference usa o Id com o literal NFAG maiusculo
	if !strings.Contains(assinado, `URI="#NFAG`+nota.ChaveAcesso()+`"`) {
		t.Fatalf("Reference URI inesperada: %.400s", assinado)
	}
}

// ---------------------------------------------------------------------------
// Enviar -- recepcao sincrona (cStat 104 no retorno!)
// ---------------------------------------------------------------------------

func retornoRecepcao(cStatRetorno string, digVal string) string {
	return `<retNFAg xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>` + cStatRetorno + `</cStat>` +
		`<xMotivo>Lote processado</xMotivo><cUF>35</cUF>` +
		`<protNFAg versao="1.00"><infProt><tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic>` +
		`<chNFAg>` + chaveTeste + `</chNFAg><dhRecbto>2026-03-05T10:05:00-03:00</dhRecbto>` +
		`<nProt>335260000000001</nProt><digVal>` + digVal + `</digVal><cStat>100</cStat>` +
		`<xMotivo>Autorizado o uso da NFAg</xMotivo></infProt></protNFAg></retNFAg>`
}

func TestEnviarRecepcaoSincrona(t *testing.T) {
	// a NFAg considera o envio OK com cStat do RETORNO = 104 (nao 100!);
	// o digVal do protocolo ecoa o DigestValue transmitido
	var corpo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		corpo = string(b)
		digVal := ""
		if xmlEnviado, err := dfe.Base64Gunzip(dadosMsg(corpo)); err == nil {
			digVal = trecho(xmlEnviado, "DigestValue")
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
			`<soap:Body><nfagResultMsg xmlns="http://svc">` + retornoRecepcao("104", digVal) +
			`</nfagResultMsg></soap:Body></soap:Envelope>`))
	}))
	defer srv.Close()
	c := componenteTeste(t, srv)

	nota := &NotaFiscal{NFAg: notaMinima()}
	ret, err := c.Enviar(testContext(), nota)
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if ret.CStat != 104 || ret.ProtNFAg.CStat != 100 {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	// protocolo anexado com cStat do retorno = 104 + protocolo processado
	if nota.NFAg.ProcNFAg.NProt != "335260000000001" {
		t.Fatalf("protocolo nao anexado: %+v", nota.NFAg.ProcNFAg)
	}

	// envelope: nfagDadosMsg no namespace do servico; DadosMsg gzip+base64
	if !strings.Contains(corpo, `<nfagDadosMsg xmlns="`+Namespace+`/wsdl/NFAgRecepcao">`) {
		t.Fatalf("envelope inesperado: %.300s", corpo)
	}
	xmlEnviado, err := dfe.Base64Gunzip(dadosMsg(corpo))
	if err != nil {
		t.Fatalf("DadosMsg nao e gzip+base64: %v", err)
	}
	if !strings.HasPrefix(xmlEnviado, "<NFAg") || !strings.HasSuffix(xmlEnviado, "</NFAg>") {
		t.Fatalf("DadosMsg nao e a NFAg solta: %.200s", xmlEnviado)
	}
	if err := dfe.VerificarAssinatura(xmlEnviado); err != nil {
		t.Fatalf("XML transmitido sem assinatura valida: %v", err)
	}

	// GerarXMLProc apos a autorizacao reembute a Signature
	proc, err := GerarXMLProc(nota.NFAg)
	if err != nil {
		t.Fatalf("GerarXMLProc: %v", err)
	}
	if !strings.Contains(proc, "<Signature ") {
		t.Fatal("nfagProc sem Signature")
	}
	if err := dfe.VerificarAssinatura(proc); err != nil {
		t.Fatalf("assinatura do nfagProc nao verifica: %v", err)
	}
}

func TestEnviarRetorno100NaoAnexaProtocolo(t *testing.T) {
	// fidelidade ao TratarResposta da NFAg: retorno com cStat 100 (que na
	// NFGas seria sucesso) NAO anexa o protocolo -- a condicao exige 104
	srv := servidorSOAP(t, retornoRecepcao("100", ""), nil)
	defer srv.Close()
	c := componenteTeste(t, srv)

	nota := &NotaFiscal{NFAg: notaMinima()}
	ret, err := c.Enviar(testContext(), nota)
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if ret.CStat != 100 {
		t.Fatalf("cStat do retorno = %d", ret.CStat)
	}
	if nota.NFAg.ProcNFAg.NProt != "" {
		t.Fatal("protocolo NAO deveria ser anexado com retorno cStat 100 (condicao da NFAg e 104)")
	}
}

// ---------------------------------------------------------------------------
// StatusServico, Consultar e Cancelamento
// ---------------------------------------------------------------------------

func TestStatusServico(t *testing.T) {
	retorno := `<retConsStatServNFAg xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>107</cStat>` +
		`<xMotivo>Servico em Operacao</xMotivo><cUF>43</cUF>` +
		`<dhRecbto>2026-03-05T10:00:00-03:00</dhRecbto><tMed>1</tMed></retConsStatServNFAg>`

	var corpo string
	srv := servidorSOAP(t, retorno, &corpo)
	defer srv.Close()
	c := componenteTeste(t, srv)

	ret, err := c.StatusServico(testContext())
	if err != nil {
		t.Fatalf("StatusServico: %v", err)
	}
	if !ret.EmOperacao() || ret.CUF != 43 {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	msg := dadosMsg(corpo)
	if !strings.Contains(msg, "<consStatServNFAg") || !strings.Contains(msg, "<xServ>STATUS</xServ>") ||
		strings.Contains(msg, "<cUF>") {
		t.Fatalf("consStatServNFAg inesperado (sem cUF, como o original): %s", msg)
	}
}

func TestConsultar(t *testing.T) {
	retorno := `<retConsSitNFAg xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>100</cStat>` +
		`<xMotivo>Autorizado o uso da NFAg</xMotivo><cUF>35</cUF>` +
		`<chNFAg>` + chaveTeste + `</chNFAg></retConsSitNFAg>`

	var corpo string
	srv := servidorSOAP(t, retorno, &corpo)
	defer srv.Close()
	c := componenteTeste(t, srv)

	ret, err := c.Consultar(testContext(), "NFAG"+chaveTeste)
	if err != nil {
		t.Fatalf("Consultar: %v", err)
	}
	if ret.CStat != 100 || ret.ChNFAg != chaveTeste {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	msg := dadosMsg(corpo)
	for _, m := range []string{"<consSitNFAg", "<xServ>CONSULTAR</xServ>", "<chNFAg>" + chaveTeste + "</chNFAg>"} {
		if !strings.Contains(msg, m) {
			t.Fatalf("consSitNFAg sem %q: %s", m, msg)
		}
	}
}

func TestCancelamento(t *testing.T) {
	retorno := `<retEventoNFAg xmlns="` + Namespace + `" versao="1.00"><infEvento>` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cOrgao>35</cOrgao>` +
		`<cStat>135</cStat><xMotivo>Evento registrado e vinculado a NFAg</xMotivo>` +
		`<chNFAg>` + chaveTeste + `</chNFAg><tpEvento>110111</tpEvento>` +
		`<nSeqEvento>1</nSeqEvento><dhRegEvento>2026-03-06T09:00:00-03:00</dhRegEvento>` +
		`<nProt>335260000000002</nProt></infEvento></retEventoNFAg>`

	var corpo string
	srv := servidorSOAP(t, retorno, &corpo)
	defer srv.Close()
	c := componenteTeste(t, srv)

	ret, err := c.Cancelamento(testContext(), chaveTeste, "335260000000001", "cancelamento em homologacao")
	if err != nil {
		t.Fatalf("Cancelamento: %v", err)
	}
	if ret.RetInfEvento.CStat != 135 || !ret.RetInfEvento.Registrado() {
		t.Fatalf("retorno inesperado: %+v", ret.RetInfEvento)
	}

	// evento assinado, sem gzip
	msg := dadosMsg(corpo)
	if !strings.Contains(msg, "<evCancNFAg>") || !strings.Contains(msg, "<SignatureValue>") {
		t.Fatalf("evento inesperado: %.300s", msg)
	}
	if err := dfe.VerificarAssinatura(msg); err != nil {
		t.Fatalf("assinatura do evento nao verifica: %v", err)
	}
}
