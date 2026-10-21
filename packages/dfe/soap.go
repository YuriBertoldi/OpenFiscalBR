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
	"errors"
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
	return c.Executar(ctx, Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:   tagDadosMsg,
		DadosMsg:      dadosMsg,
		TagsResultado: []string{tagResultMsg},
	})
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

// Chamada de web service com CABECALHO SOAP e com varios nomes possiveis
// para o elemento de resultado.
//
// O Chamar (soap.go) cobre o caso simples, que e o da NFGas: sem header e
// com um nome de resultado so. A NF-e precisa dos dois eixos a mais:
//
//   - abaixo da versao 4.00 ela manda um <soap12:Header> com o cUF e o
//     versaoDados; da 4.00 em diante, NENHUM header
//     (ACBrNFeWebServices.pas:692);
//   - o elemento de retorno muda de nome por versao de servico e por UF, e
//     o ACBr tenta os nomes em ordem (SeparaDadosArray).

// Requisicao descreve uma chamada a um web service da SEFAZ.
//
// Existe como struct, e nao como lista de parametros, porque sao sete
// eixos e a maioria das chamadas deixa dois deles vazios -- uma assinatura
// posicional de sete argumentos seria ilegivel no ponto de uso.
type Requisicao struct {
	URL        string
	Servico    string // namespace do servico, usado no envelope
	SoapAction string

	// TagDadosMsg e o elemento que envolve a mensagem no Body
	// ("nfeDadosMsg", "nfgasDadosMsg").
	TagDadosMsg string
	DadosMsg    string

	// TagCabecalho e Cabecalho montam o <soap12:Header>. Vazios produzem
	// envelope SEM header nenhum -- nao um header vazio.
	TagCabecalho string
	Cabecalho    string

	// TagsResultado sao os nomes possiveis do elemento de retorno, tentados
	// em ordem sobre a MESMA resposta.
	TagsResultado []string
}

// Executar transmite a requisicao e devolve o conteudo do elemento de
// resultado.
//
// Os nomes de TagsResultado sao tentados sobre a resposta JA RECEBIDA.
// Reenviar o documento a cada tentativa seria errado e, nos servicos de
// autorizacao, perigoso: a SEFAZ registraria duplicidade.
func (c *ClienteSOAP) Executar(ctx context.Context, req Requisicao) (string, error) {
	if len(req.TagsResultado) == 0 {
		return "", ErrSemTagResultado
	}

	envelope := MontarEnvelopeComCabecalho(req.TagDadosMsg, req.Servico,
		req.DadosMsg, req.TagCabecalho, req.Cabecalho)

	corpo, status, err := c.transmitir(ctx, req.URL, req.SoapAction, envelope)
	if err != nil {
		return "", err
	}

	var ultimoErro error
	for _, tag := range req.TagsResultado {
		resultado, err := ExtrairResultadoSOAP(corpo, tag)
		if err == nil {
			if status != http.StatusOK {
				return "", &ErroTransmissao{URL: req.URL, SoapAction: req.SoapAction,
					StatusHTTP: status, Err: fmt.Errorf("status HTTP inesperado")}
			}
			return resultado, nil
		}
		// Um soap:Fault nao melhora com outro nome de tag: para na hora.
		if errors.Is(err, ErrSOAPFault) {
			return "", &ErroTransmissao{URL: req.URL, SoapAction: req.SoapAction,
				StatusHTTP: status, Err: err}
		}
		ultimoErro = err
	}
	return "", &ErroTransmissao{URL: req.URL, SoapAction: req.SoapAction,
		StatusHTTP: status, Err: ultimoErro}
}

// MontarEnvelopeComCabecalho e o MontarEnvelope com um <soap12:Header>
// opcional. Porte de TDFeWebService.DefinirEnvelopeSoap
// (ACBrDFeWebService.pas:320-345).
func MontarEnvelopeComCabecalho(tagDadosMsg, servico, dadosMsg, tagCabecalho, cabecalho string) string {
	if tagCabecalho == "" || cabecalho == "" {
		return MontarEnvelope(tagDadosMsg, servico, dadosMsg)
	}

	return `<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"` +
		` xmlns:xsd="http://www.w3.org/2001/XMLSchema"` +
		` xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap12:Header><` + tagCabecalho + ` xmlns="` + servico + `">` +
		cabecalho +
		`</` + tagCabecalho + `></soap12:Header>` +
		`<soap12:Body><` + tagDadosMsg + ` xmlns="` + servico + `">` +
		dadosMsg +
		`</` + tagDadosMsg + `></soap12:Body></soap12:Envelope>`
}

// transmitir envia o envelope ja montado e devolve o corpo bruto da
// resposta, junto do status HTTP.
func (c *ClienteSOAP) transmitir(ctx context.Context, url, soapAction, envelope string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(envelope))
	if err != nil {
		return "", 0, &ErroTransmissao{URL: url, SoapAction: soapAction, Err: err}
	}
	req.Header.Set("Content-Type", `application/soap+xml; charset=utf-8; action="`+soapAction+`"`)
	req.Header.Set("SOAPAction", soapAction)

	cliente := c.HTTPClient
	if cliente == nil {
		cliente = &http.Client{Timeout: TimeoutPadrao}
	}
	resp, err := cliente.Do(req)
	if err != nil {
		return "", 0, &ErroTransmissao{URL: url, SoapAction: soapAction, Err: err}
	}
	defer resp.Body.Close()

	corpo, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, &ErroTransmissao{URL: url, SoapAction: soapAction,
			StatusHTTP: resp.StatusCode, Err: err}
	}
	return string(corpo), resp.StatusCode, nil
}
