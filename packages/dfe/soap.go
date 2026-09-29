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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Cliente SOAP 1.2 dos web services da SEFAZ, com TLS mutuo pelo certificado
// A1. Porte do papel de TDFeWebService (EnviarDados/Envelope) + TDFeSSL
// (transporte HTTP).

// TimeoutPadrao e o timeout do ACBr (TimeOut = 5000ms por requisicao no
// componente, mas na pratica os servicos da SEFAZ exigem folga; o ACBrNFGas
// usa o default do WebServices que o usuario ajusta). 90s cobre o pior caso.
const TimeoutPadrao = 90 * time.Second

// ClienteSOAP envia envelopes SOAP 1.2 para a SEFAZ.
type ClienteSOAP struct {
	// HTTPClient pode ser substituido em teste. NovoClienteSOAP ja o
	// configura com o certificado do cliente.
	HTTPClient *http.Client
}

// NovoClienteSOAP cria o cliente com TLS mutuo usando o certificado A1.
// cert nil cria um cliente sem certificado de cliente (so serve para
// servicos que nao o exigem e para testes locais).
func NovoClienteSOAP(cert *Certificado, timeout time.Duration) *ClienteSOAP {
	if timeout <= 0 {
		timeout = TimeoutPadrao
	}
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
	if cert != nil {
		tlsCfg.Certificates = []tls.Certificate{cert.TLSCertificate()}
	}
	return &ClienteSOAP{
		HTTPClient: &http.Client{
			Timeout:   timeout,
			Transport: &http.Transport{TLSClientConfig: tlsCfg},
		},
	}
}

// MontarEnvelope monta o envelope SOAP 1.2 no formato exato do ACBr
// (TDFeWebService.Executar): Body com um unico elemento <tagDadosMsg>
// no namespace do servico, contendo a mensagem XML.
//
//	tagDadosMsg: "nfgasDadosMsg"
//	servico:     "http://www.portalfiscal.inf.br/nfgas/wsdl/NFGasRecepcao"
func MontarEnvelope(tagDadosMsg, servico, dadosMsg string) string {
	return `<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"` +
		` xmlns:xsd="http://www.w3.org/2001/XMLSchema"` +
		` xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap12:Body><` + tagDadosMsg + ` xmlns="` + servico + `">` +
		dadosMsg +
		`</` + tagDadosMsg + `></soap12:Body></soap12:Envelope>`
}

// Chamar envia dadosMsg para a URL do servico e devolve o CONTEUDO do
// elemento de resultado (tagResultMsg, ex.: "nfgasResultMsg") -- ja sem o
// envelope, equivalente ao SeparaDados + RemoverNameSpace do ACBr.
// soapAction e a action completa do metodo (vai no Content-Type, como o
// SOAP 1.2 pede, e tambem no header SOAPAction por compatibilidade).
func (c *ClienteSOAP) Chamar(ctx context.Context, url, servico, soapAction, tagDadosMsg, tagResultMsg, dadosMsg string) (string, error) {
	envelope := MontarEnvelope(tagDadosMsg, servico, dadosMsg)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(envelope))
	if err != nil {
		return "", &ErroTransmissao{URL: url, SoapAction: soapAction, Err: err}
	}
	req.Header.Set("Content-Type", `application/soap+xml; charset=utf-8; action="`+soapAction+`"`)
	req.Header.Set("SOAPAction", soapAction)

	cliente := c.HTTPClient
	if cliente == nil {
		cliente = &http.Client{Timeout: TimeoutPadrao}
	}
	resp, err := cliente.Do(req)
	if err != nil {
		return "", &ErroTransmissao{URL: url, SoapAction: soapAction, Err: err}
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &ErroTransmissao{URL: url, SoapAction: soapAction, StatusHTTP: resp.StatusCode, Err: err}
	}

	resultado, err := ExtrairResultadoSOAP(string(corpo), tagResultMsg)
	if err != nil {
		// mesmo com Fault o servidor costuma responder 500 -- devolve o
		// contexto HTTP junto
		return "", &ErroTransmissao{URL: url, SoapAction: soapAction, StatusHTTP: resp.StatusCode, Err: err}
	}
	if resp.StatusCode != http.StatusOK {
		return "", &ErroTransmissao{URL: url, SoapAction: soapAction, StatusHTTP: resp.StatusCode,
			Err: fmt.Errorf("status HTTP inesperado")}
	}
	return resultado, nil
}

// ExtrairResultadoSOAP localiza o elemento de resultado no envelope de
// resposta e devolve seu conteudo interno (os filhos serializados, ou o
// texto). A comparacao do nome e por local name, caso-insensitivo, porque a
// SEFAZ ora responde "nfgasResultMsg", ora "NFGasResultMsg" -- e o motivo de
// o ACBr chamar SeparaDados duas vezes. soap:Fault vira erro com o texto do
// Reason/faultstring.
func ExtrairResultadoSOAP(corpo, tagResultMsg string) (string, error) {
	doc, err := pcn.ParseString(corpo)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrRespostaSemDados, err)
	}

	if fault := buscarProfundo(doc.Root, "Fault", ""); fault != nil {
		razao := pcn.ConteudoStr(buscarProfundo(fault, "Text", ""))
		if razao == "" {
			razao = pcn.ConteudoStr(buscarProfundo(fault, "faultstring", ""))
		}
		if razao == "" {
			razao = pcn.ConteudoStr(buscarProfundo(fault, "Reason", ""))
		}
		return "", fmt.Errorf("%w: %s", ErrSOAPFault, razao)
	}

	no := buscarLocalFold(doc.Root, tagResultMsg)
	if no == nil {
		return "", fmt.Errorf("%w: <%s>", ErrRespostaSemDados, tagResultMsg)
	}
	return innerXML(no), nil
}

// buscarLocalFold procura em profundidade por local name, caso-insensitivo.
func buscarLocalFold(n *pcn.Node, local string) *pcn.Node {
	if n == nil {
		return nil
	}
	if strings.EqualFold(n.Nome, local) {
		return n
	}
	for _, f := range n.Filhos {
		if achado := buscarLocalFold(f, local); achado != nil {
			return achado
		}
	}
	return nil
}

// innerXML devolve o conteudo interno de um elemento: os filhos
// serializados byte a byte (OuterXML) ou, sem filhos, o texto.
func innerXML(n *pcn.Node) string {
	if n == nil {
		return ""
	}
	if len(n.Filhos) == 0 {
		return strings.TrimSpace(n.Texto)
	}
	var b strings.Builder
	for _, f := range n.Filhos {
		b.WriteString(f.OuterXML())
	}
	return b.String()
}
