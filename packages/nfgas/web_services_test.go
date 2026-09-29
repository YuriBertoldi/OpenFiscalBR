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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
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

// servidorSOAP monta um httptest que devolve resultado dentro de
// nfgasResultMsg e captura o corpo recebido.
func servidorSOAP(t *testing.T, resultado string, corpoRecebido *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		if corpoRecebido != nil {
			*corpoRecebido = string(corpo)
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
			`<soap:Body><nfgasResultMsg xmlns="http://svc">` + resultado +
			`</nfgasResultMsg></soap:Body></soap:Envelope>`))
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

// ---------------------------------------------------------------------------
// PrepararEnvio -- geracao + assinatura verificavel
// ---------------------------------------------------------------------------

func TestPrepararEnvioAssinaVerificavel(t *testing.T) {
	c := NovoComponente()
	c.Configuracoes.Certificado = certificadoTeste(t)

	nota := &NotaFiscal{NFGas: notaMinima()}
	assinado, err := c.PrepararEnvio(nota)
	if err != nil {
		t.Fatalf("PrepararEnvio: %v", err)
	}
	if nota.XMLAssinado != assinado || assinado == "" {
		t.Fatal("XMLAssinado nao foi preenchido")
	}
	if err := dfe.VerificarAssinatura(assinado); err != nil {
		t.Fatalf("assinatura gerada nao verifica: %v", err)
	}
	if !strings.Contains(assinado, `URI="#NFGas`+nota.ChaveAcesso()+`"`) {
		t.Fatalf("Reference URI inesperada: %.400s", assinado)
	}
}

func TestPrepararEnvioCalculaHashCSRT(t *testing.T) {
	c := NovoComponente()
	c.Configuracoes.Certificado = certificadoTeste(t)
	c.Configuracoes.IDCSRT = 1
	c.Configuracoes.CSRT = "G8063VRTNDMO886SFNK5LDUDEI24XJ22YIPO"

	n := notaMinima()
	n.InfRespTec = InfRespTec{CNPJ: "11222333000181", XContato: "Contato",
		Email: "resp@teste.com", Fone: "1130001000"}
	nota := &NotaFiscal{NFGas: n}

	assinado, err := c.PrepararEnvio(nota)
	if err != nil {
		t.Fatal(err)
	}
	want := dfe.HashCSRT(c.Configuracoes.CSRT, n.ChaveAcesso())
	if !strings.Contains(assinado, "<idCSRT>001</idCSRT><hashCSRT>"+want+"</hashCSRT>") {
		t.Fatalf("gRespTec sem idCSRT/hashCSRT esperados: %s", assinado)
	}
}

// ---------------------------------------------------------------------------
// Enviar -- recepcao sincrona
// ---------------------------------------------------------------------------

func TestEnviarRecepcaoSincrona(t *testing.T) {
	// a SEFAZ responde retNFGas, que e um retConsSitNFGas renomeado; o
	// digVal do protocolo ecoa o DigestValue do XML recebido, como a SEFAZ
	// real faz (e como o ValidarDigest exige)
	var corpo string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		corpo = string(b)
		digVal := ""
		if xmlEnviado, err := dfe.Base64Gunzip(dadosMsg(corpo)); err == nil {
			digVal = trecho(xmlEnviado, "DigestValue")
		}
		retorno := `<retNFGas xmlns="` + Namespace + `" versao="1.00">` +
			`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>100</cStat>` +
			`<xMotivo>Autorizado o uso da NFGas</xMotivo><cUF>35</cUF>` +
			`<protNFGas versao="1.00"><infProt><tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic>` +
			`<chNFGas>` + chaveTeste + `</chNFGas><dhRecbto>2026-03-05T10:05:00-03:00</dhRecbto>` +
			`<nProt>335260000000001</nProt><digVal>` + digVal + `</digVal><cStat>100</cStat>` +
			`<xMotivo>Autorizado o uso da NFGas</xMotivo></infProt></protNFGas></retNFGas>`
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
			`<soap:Body><nfgasResultMsg xmlns="http://svc">` + retorno +
			`</nfgasResultMsg></soap:Body></soap:Envelope>`))
	}))
	defer srv.Close()
	c := componenteTeste(t, srv)

	nota := &NotaFiscal{NFGas: notaMinima()}
	ret, err := c.Enviar(testContext(), nota)
	if err != nil {
		t.Fatalf("Enviar: %v", err)
	}
	if ret.CStat != 100 || ret.ProtNFGas.NProt != "335260000000001" {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	// protocolo anexado a nota, pronto para GerarXMLProc
	if nota.NFGas.ProcNFGas.NProt != "335260000000001" {
		t.Fatalf("protocolo nao anexado a nota: %+v", nota.NFGas.ProcNFGas)
	}

	// o DadosMsg e gzip+base64 do <NFGas> assinado, dentro do envelope
	ini := strings.Index(corpo, ">") // fim da tag Envelope
	if !strings.Contains(corpo, `<nfgasDadosMsg xmlns="`+Namespace+`/wsdl/NFGasRecepcao">`) {
		t.Fatalf("envelope sem nfgasDadosMsg no namespace do servico: %.300s", corpo[ini:])
	}
	b64 := dadosMsg(corpo)
	xmlEnviado, err := dfe.Base64Gunzip(b64)
	if err != nil {
		t.Fatalf("DadosMsg nao e gzip+base64: %v", err)
	}
	if !strings.HasPrefix(xmlEnviado, "<NFGas") || !strings.HasSuffix(xmlEnviado, "</NFGas>") {
		t.Fatalf("DadosMsg nao e a NFGas solta: %.200s", xmlEnviado)
	}
	if err := dfe.VerificarAssinatura(xmlEnviado); err != nil {
		t.Fatalf("XML transmitido sem assinatura valida: %v", err)
	}

	// o fluxo documentado continua: GerarXMLProc apos a autorizacao tem que
	// reembutir a Signature (no ACBr o proc e montado do XMLAssinado)
	proc, err := GerarXMLProc(nota.NFGas)
	if err != nil {
		t.Fatalf("GerarXMLProc apos Enviar: %v", err)
	}
	if !strings.Contains(proc, "<Signature ") {
		t.Fatalf("nfgasProc sem Signature -- documento de distribuicao invalido:\n%.400s", proc)
	}
	if err := dfe.VerificarAssinatura(proc); err != nil {
		t.Fatalf("assinatura do nfgasProc nao verifica: %v", err)
	}
	if !strings.Contains(proc, "<protNFGas") {
		t.Fatal("nfgasProc sem protocolo")
	}
}

func TestEnviarDigestDivergente(t *testing.T) {
	// ValidarDigest do TratarResposta: digVal do protocolo diferente do
	// DigestValue transmitido e erro, e o protocolo NAO e anexado como bom.
	retorno := `<retNFGas xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>100</cStat>` +
		`<xMotivo>Autorizado</xMotivo><cUF>35</cUF>` +
		`<protNFGas versao="1.00"><infProt><tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic>` +
		`<chNFGas>` + chaveTeste + `</chNFGas><dhRecbto>2026-03-05T10:05:00-03:00</dhRecbto>` +
		`<nProt>335260000000001</nProt><digVal>DIGEST-ERRADO=</digVal><cStat>100</cStat>` +
		`<xMotivo>Autorizado</xMotivo></infProt></protNFGas></retNFGas>`

	srv := servidorSOAP(t, retorno, nil)
	defer srv.Close()
	c := componenteTeste(t, srv)

	nota := &NotaFiscal{NFGas: notaMinima()}
	_, err := c.Enviar(testContext(), nota)
	if !errors.Is(err, ErrDigestDivergente) {
		t.Fatalf("esperado ErrDigestDivergente, veio %v", err)
	}
}

// ---------------------------------------------------------------------------
// StatusServico e Consultar
// ---------------------------------------------------------------------------

func TestStatusServico(t *testing.T) {
	retorno := `<retConsStatServNFGas xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>107</cStat>` +
		`<xMotivo>Servico em Operacao</xMotivo><cUF>43</cUF>` +
		`<dhRecbto>2026-03-05T10:00:00-03:00</dhRecbto><tMed>1</tMed></retConsStatServNFGas>`

	var corpo string
	srv := servidorSOAP(t, retorno, &corpo)
	defer srv.Close()
	c := componenteTeste(t, srv)

	ret, err := c.StatusServico(testContext())
	if err != nil {
		t.Fatalf("StatusServico: %v", err)
	}
	if !ret.EmOperacao() || ret.TMed != 1 || ret.CUF != 43 {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	// consStatServNFGas NAO leva cUF (TConsStatServ com AGerarcUF=False)
	msg := dadosMsg(corpo)
	if !strings.Contains(msg, "<xServ>STATUS</xServ>") || strings.Contains(msg, "<cUF>") {
		t.Fatalf("consStatServNFGas inesperado: %s", msg)
	}
}

func TestConsultar(t *testing.T) {
	retorno := `<retConsSitNFGas xmlns="` + Namespace + `" versao="1.00">` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cStat>100</cStat>` +
		`<xMotivo>Autorizado o uso da NFGas</xMotivo><cUF>35</cUF>` +
		`<chNFGas>` + chaveTeste + `</chNFGas></retConsSitNFGas>`

	var corpo string
	srv := servidorSOAP(t, retorno, &corpo)
	defer srv.Close()
	c := componenteTeste(t, srv)

	ret, err := c.Consultar(testContext(), "NFGas"+chaveTeste)
	if err != nil {
		t.Fatalf("Consultar: %v", err)
	}
	if ret.CStat != 100 || ret.ChNFGas != chaveTeste {
		t.Fatalf("retorno inesperado: %+v", ret)
	}
	msg := dadosMsg(corpo)
	for _, m := range []string{"<xServ>CONSULTAR</xServ>", "<chNFGas>" + chaveTeste + "</chNFGas>"} {
		if !strings.Contains(msg, m) {
			t.Fatalf("consSitNFGas sem %q: %s", m, msg)
		}
	}

	// chave invalida falha antes de qualquer rede
	if _, err := c.Consultar(testContext(), "123"); err == nil {
		t.Fatal("chave invalida deveria falhar cedo")
	}
}

// ---------------------------------------------------------------------------
// Cancelamento (evento)
// ---------------------------------------------------------------------------

func TestCancelamento(t *testing.T) {
	retorno := `<retEventoNFGas xmlns="` + Namespace + `" versao="1.00"><infEvento>` +
		`<tpAmb>2</tpAmb><verAplic>SVRS1.0</verAplic><cOrgao>35</cOrgao>` +
		`<cStat>135</cStat><xMotivo>Evento registrado e vinculado a NFGas</xMotivo>` +
		`<chNFGas>` + chaveTeste + `</chNFGas><tpEvento>110111</tpEvento>` +
		`<nSeqEvento>1</nSeqEvento><dhRegEvento>2026-03-06T09:00:00-03:00</dhRegEvento>` +
		`<nProt>335260000000002</nProt></infEvento></retEventoNFGas>`

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

	// o DadosMsg e o eventoNFGas ASSINADO, sem gzip
	msg := dadosMsg(corpo)
	if !strings.Contains(msg, "<evCancNFGas>") || !strings.Contains(msg, "<nProt>335260000000001</nProt>") {
		t.Fatalf("evento de cancelamento inesperado: %s", msg)
	}
	if !strings.Contains(msg, "<SignatureValue>") {
		t.Fatal("evento deveria ir assinado")
	}
	if err := dfe.VerificarAssinatura(msg); err != nil {
		t.Fatalf("assinatura do evento nao verifica: %v", err)
	}
}

// ---------------------------------------------------------------------------
// QR-Code via Componente (offline assina com o certificado configurado)
// ---------------------------------------------------------------------------

func TestComponenteGerarQRCodeOffline(t *testing.T) {
	c := NovoComponente()
	c.Configuracoes.Certificado = certificadoTeste(t)

	n := notaMinima()
	n.Ide.TpEmis = pcn.TeOffLine
	if _, err := GerarXML(n); err != nil {
		t.Fatal(err)
	}
	url, err := c.GerarQRCode(n)
	if err != nil {
		t.Fatalf("GerarQRCode: %v", err)
	}
	if !strings.Contains(url, "&sign=") {
		t.Fatalf("QR-Code offline sem sign: %q", url)
	}
}

// dadosMsg extrai o conteudo de <nfgasDadosMsg ...> do envelope capturado.
func dadosMsg(corpo string) string {
	ini := strings.Index(corpo, "<nfgasDadosMsg")
	if ini < 0 {
		return ""
	}
	ini = strings.Index(corpo[ini:], ">") + ini + 1
	fim := strings.Index(corpo, "</nfgasDadosMsg>")
	if fim < 0 || fim < ini {
		return ""
	}
	return corpo[ini:fim]
}
