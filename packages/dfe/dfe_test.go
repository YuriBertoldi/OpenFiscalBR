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

package dfe

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Canonicalizacao (C14N 1.0)
// ---------------------------------------------------------------------------

func TestC14NElementoVazioEDeclNamespaceHerdado(t *testing.T) {
	// C14N nunca usa self-closing, e o namespace default herdado do
	// documento e renderizado no elemento raiz do fragmento.
	got, err := CanonicalizarFragmento("<ide><cUF>35</cUF><serie/></ide>",
		map[string]string{"": "http://www.portalfiscal.inf.br/nfgas"}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := `<ide xmlns="http://www.portalfiscal.inf.br/nfgas"><cUF>35</cUF><serie></serie></ide>`
	if got != want {
		t.Fatalf("c14n = %q, esperado %q", got, want)
	}
}

func TestC14NOrdenaAtributos(t *testing.T) {
	got, err := CanonicalizarFragmento(`<a versao="1.00" Id="X"/>`, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	// ordenacao por local name: "Id" < "versao" (maiusculas antes)
	want := `<a Id="X" versao="1.00"></a>`
	if got != want {
		t.Fatalf("c14n = %q, esperado %q", got, want)
	}
}

func TestC14NNaoRedeclaraNamespaceJaRenderizado(t *testing.T) {
	// xmlns repetido no filho com o MESMO valor nao e renderizado de novo.
	frag := `<NFGas xmlns="http://ns"><infNFGas xmlns="http://ns"><x>1</x></infNFGas></NFGas>`
	got, err := CanonicalizarFragmento(frag, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	want := `<NFGas xmlns="http://ns"><infNFGas><x>1</x></infNFGas></NFGas>`
	if got != want {
		t.Fatalf("c14n = %q, esperado %q", got, want)
	}
}

func TestC14NEscapes(t *testing.T) {
	got, err := CanonicalizarFragmento("<a v=\"x\tq\">a&amp;b<b/></a>", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	want := "<a v=\"x&#x9;q\">a&amp;b<b></b></a>"
	if got != want {
		t.Fatalf("c14n = %q, esperado %q", got, want)
	}
}

func TestC14NPreservaTextoMistoEIdentacao(t *testing.T) {
	// Whitespace entre elementos FAZ PARTE do digest -- o canonicalizador
	// trabalha sobre os bytes originais, entao preserva.
	frag := "<a>\n  <b>1</b>\n</a>"
	got, err := CanonicalizarFragmento(frag, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "<a>\n  <b>1</b>\n</a>" {
		t.Fatalf("c14n = %q", got)
	}
}

func TestC14NRemoveSignatureEnveloped(t *testing.T) {
	frag := `<root><dado>1</dado><Signature xmlns="http://www.w3.org/2000/09/xmldsig#"><SignedInfo></SignedInfo></Signature></root>`
	got, err := CanonicalizarFragmento(frag, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	want := `<root><dado>1</dado></root>`
	if got != want {
		t.Fatalf("c14n = %q, esperado %q", got, want)
	}
}

func TestC14NAtributoComPrefixoNaoSuportadoDaErro(t *testing.T) {
	_, err := CanonicalizarFragmento(`<a xmlns:x="http://x" x:attr="1"/>`, nil, false)
	if err == nil {
		t.Fatal("esperado erro para atributo com prefixo")
	}
}

// ---------------------------------------------------------------------------
// Certificado e assinatura
// ---------------------------------------------------------------------------

func gerarCertificadoTeste(t *testing.T) *Certificado {
	t.Helper()
	chave, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(0xABCD),
		Subject:      pkix.Name{CommonName: "EMPRESA TESTE:11222333000181"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &chave.PublicKey, chave)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &Certificado{Cert: cert, Chave: chave}
}

const xmlParaAssinar = `<NFGas xmlns="http://www.portalfiscal.inf.br/nfgas">` +
	`<infNFGas Id="NFGas35260311222333000181760010000000011000000018" versao="1.00">` +
	`<ide><cUF>35</cUF><serie>1</serie></ide></infNFGas></NFGas>`

func TestAssinarEVerificarRoundTrip(t *testing.T) {
	cert := gerarCertificadoTeste(t)
	assinado, err := AssinarXML(cert, xmlParaAssinar, "infNFGas")
	if err != nil {
		t.Fatalf("AssinarXML: %v", err)
	}
	for _, marca := range []string{
		"<Signature xmlns=\"http://www.w3.org/2000/09/xmldsig#\">",
		"URI=\"#NFGas35260311222333000181760010000000011000000018\"",
		"<X509Certificate>",
	} {
		if !strings.Contains(assinado, marca) {
			t.Fatalf("XML assinado sem %q:\n%s", marca, assinado)
		}
	}
	if !strings.HasSuffix(assinado, "</Signature></NFGas>") {
		t.Fatalf("Signature deveria ser o ultimo filho de NFGas:\n%s", assinado)
	}
	if err := VerificarAssinatura(assinado); err != nil {
		t.Fatalf("VerificarAssinatura: %v", err)
	}
}

func TestVerificarAssinaturaDetectaAdulteracao(t *testing.T) {
	cert := gerarCertificadoTeste(t)
	assinado, err := AssinarXML(cert, xmlParaAssinar, "infNFGas")
	if err != nil {
		t.Fatal(err)
	}
	adulterado := strings.Replace(assinado, "<cUF>35</cUF>", "<cUF>43</cUF>", 1)
	err = VerificarAssinatura(adulterado)
	if !errors.Is(err, ErrAssinaturaInvalida) {
		t.Fatalf("esperado ErrAssinaturaInvalida, veio %v", err)
	}
}

func TestAssinarSubstituiAssinaturaAnterior(t *testing.T) {
	cert := gerarCertificadoTeste(t)
	assinado, err := AssinarXML(cert, xmlParaAssinar, "infNFGas")
	if err != nil {
		t.Fatal(err)
	}
	reassinado, err := AssinarXML(cert, assinado, "infNFGas")
	if err != nil {
		t.Fatalf("reassinar: %v", err)
	}
	if n := strings.Count(reassinado, "<Signature "); n != 1 {
		t.Fatalf("esperada 1 Signature apos reassinar, ha %d", n)
	}
	if err := VerificarAssinatura(reassinado); err != nil {
		t.Fatalf("VerificarAssinatura apos reassinar: %v", err)
	}
}

func TestAssinarSemCertificado(t *testing.T) {
	_, err := AssinarXML(nil, xmlParaAssinar, "infNFGas")
	if !errors.Is(err, ErrCertificadoNaoCarregado) {
		t.Fatalf("esperado ErrCertificadoNaoCarregado, veio %v", err)
	}
}

func TestAssinarElementoInexistente(t *testing.T) {
	cert := gerarCertificadoTeste(t)
	_, err := AssinarXML(cert, xmlParaAssinar, "infQualquer")
	if !errors.Is(err, ErrElementoNaoEncontrado) {
		t.Fatalf("esperado ErrElementoNaoEncontrado, veio %v", err)
	}
}

func TestCarregarPEM(t *testing.T) {
	base := gerarCertificadoTeste(t)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: base.Cert.Raw})
	chavePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(base.Chave)})

	cert, err := CarregarPEM(certPEM, chavePEM)
	if err != nil {
		t.Fatalf("CarregarPEM: %v", err)
	}
	if cert.RazaoSocial() != "EMPRESA TESTE:11222333000181" {
		t.Fatalf("RazaoSocial = %q", cert.RazaoSocial())
	}
	if cert.NumeroSerie() != "ABCD" {
		t.Fatalf("NumeroSerie = %q", cert.NumeroSerie())
	}
	if cert.Vencido() {
		t.Fatal("certificado de teste nao deveria estar vencido")
	}
}

func TestHashCSRT(t *testing.T) {
	// Formula de CalcularHashCSRT (ACBrDFeUtil.pas):
	// base64(sha1(CSRT + chave)). Valor esperado validado por implementacao
	// independente: printf '%s' 'G8063...23' | openssl dgst -sha1 -binary | openssl base64
	got := HashCSRT("G8063VRTNDMO886SFNK5LDUDEI24XJ22YIPO",
		"41180678005645000172550010000000021800267423")
	want := "1ANB/jmfFAA86Wvhh5GmzZwGyjQ="
	if got != want {
		t.Fatalf("HashCSRT = %q, esperado %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// gzip + base64
// ---------------------------------------------------------------------------

func TestGzipBase64RoundTrip(t *testing.T) {
	original := xmlParaAssinar
	comprimido, err := GzipBase64(original)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(comprimido, "<>") {
		t.Fatal("saida deveria ser base64 puro")
	}
	volta, err := Base64Gunzip(comprimido)
	if err != nil {
		t.Fatal(err)
	}
	if volta != original {
		t.Fatalf("round-trip perdeu dados: %q", volta)
	}
}

// ---------------------------------------------------------------------------
// SOAP
// ---------------------------------------------------------------------------

func TestMontarEnvelope(t *testing.T) {
	got := MontarEnvelope("nfgasDadosMsg",
		"http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasStatusServico",
		"<consStatServNFGas/>")
	want := `<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"` +
		` xmlns:xsd="http://www.w3.org/2001/XMLSchema"` +
		` xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap12:Body><nfgasDadosMsg xmlns="http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasStatusServico">` +
		`<consStatServNFGas/></nfgasDadosMsg></soap12:Body></soap12:Envelope>`
	if got != want {
		t.Fatalf("envelope divergente:\n%s\n---\n%s", got, want)
	}
}

func TestExtrairResultadoSOAP(t *testing.T) {
	corpo := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><nfgasResultMsg xmlns="http://svc">` +
		`<retConsStatServNFGas versao="1.00"><cStat>107</cStat></retConsStatServNFGas>` +
		`</nfgasResultMsg></soap:Body></soap:Envelope>`
	got, err := ExtrairResultadoSOAP(corpo, "nfgasResultMsg")
	if err != nil {
		t.Fatal(err)
	}
	want := `<retConsStatServNFGas versao="1.00"><cStat>107</cStat></retConsStatServNFGas>`
	if got != want {
		t.Fatalf("resultado = %q", got)
	}
}

func TestExtrairResultadoSOAPCaseInsensitive(t *testing.T) {
	// A SEFAZ ora responde nfgasResultMsg, ora NFGasResultMsg -- o ACBr
	// tenta as duas grafias (TratarResposta).
	corpo := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><NFGasResultMsg xmlns="http://svc"><ret>ok</ret></NFGasResultMsg></soap:Body></soap:Envelope>`
	got, err := ExtrairResultadoSOAP(corpo, "nfgasResultMsg")
	if err != nil {
		t.Fatal(err)
	}
	if got != "<ret>ok</ret>" {
		t.Fatalf("resultado = %q", got)
	}
}

func TestExtrairResultadoSOAPFault(t *testing.T) {
	corpo := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><soap:Fault><soap:Reason><soap:Text xml:lang="pt">Falha de schema</soap:Text></soap:Reason></soap:Fault></soap:Body></soap:Envelope>`
	_, err := ExtrairResultadoSOAP(corpo, "nfgasResultMsg")
	if !errors.Is(err, ErrSOAPFault) {
		t.Fatalf("esperado ErrSOAPFault, veio %v", err)
	}
	if !strings.Contains(err.Error(), "Falha de schema") {
		t.Fatalf("erro sem o texto do Fault: %v", err)
	}
}

func TestClienteSOAPChamar(t *testing.T) {
	var recebido string
	var contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		recebido = string(corpo)
		contentType = r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = w.Write([]byte(`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
			`<soap:Body><nfgasResultMsg xmlns="http://svc"><cStat>107</cStat></nfgasResultMsg></soap:Body></soap:Envelope>`))
	}))
	defer srv.Close()

	c := &ClienteSOAP{HTTPClient: srv.Client()}
	got, err := c.Chamar(context.Background(), srv.URL,
		"http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasStatusServico",
		"http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasStatusServico/NFGasStatusServicoNF",
		"nfgasDadosMsg", "nfgasResultMsg", "<consStatServNFGas/>")
	if err != nil {
		t.Fatalf("Chamar: %v", err)
	}
	if got != "<cStat>107</cStat>" {
		t.Fatalf("resultado = %q", got)
	}
	if !strings.Contains(recebido, `<nfgasDadosMsg xmlns="http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasStatusServico"><consStatServNFGas/></nfgasDadosMsg>`) {
		t.Fatalf("envelope recebido inesperado: %s", recebido)
	}
	if !strings.Contains(contentType, "application/soap+xml") || !strings.Contains(contentType, "action=") {
		t.Fatalf("Content-Type inesperado: %s", contentType)
	}
}

func TestClienteSOAPErroHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "indisponivel", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := &ClienteSOAP{HTTPClient: srv.Client()}
	_, err := c.Chamar(context.Background(), srv.URL, "http://svc", "acao", "nfgasDadosMsg", "nfgasResultMsg", "<x/>")
	var et *ErroTransmissao
	if !errors.As(err, &et) {
		t.Fatalf("esperado *ErroTransmissao, veio %v", err)
	}
	if et.StatusHTTP != http.StatusServiceUnavailable {
		t.Fatalf("StatusHTTP = %d", et.StatusHTTP)
	}
}

func TestAssinarDocumentoEnvelopadoInsereNoDocElement(t *testing.T) {
	// assinar um XML ja envelopado (nfgasProc) tem que por a Signature como
	// ultimo filho do NFGas (pai do infNFGas), nunca fora dele no proc.
	cert := gerarCertificadoTeste(t)
	// monta o proc com a NFGas sem redeclarar xmlns (herda do proc)
	nota := strings.Replace(xmlParaAssinar, ` xmlns="http://www.portalfiscal.inf.br/nfgas"`, "", 1)
	proc := `<nfgasProc versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfgas">` + nota +
		`<protNFGas versao="1.00"><infProt><nProt>1</nProt></infProt></protNFGas></nfgasProc>`

	assinado, err := AssinarXML(cert, proc, "infNFGas")
	if err != nil {
		t.Fatalf("AssinarXML no proc: %v", err)
	}
	if !strings.Contains(assinado, "</Signature></NFGas>") {
		t.Fatalf("Signature deveria fechar junto do NFGas (ultimo filho), nao fora dele:\n%s", assinado)
	}
	if err := VerificarAssinatura(assinado); err != nil {
		t.Fatalf("VerificarAssinatura no proc: %v", err)
	}
}
