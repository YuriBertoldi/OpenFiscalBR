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
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Comunicacao com os web services da SEFAZ. Porte de
// ACBrNFGasWebServices.pas sobre o cliente SOAP de packages/dfe.
//
// A recepcao da NFGas e SINCRONA e aceita UM documento por chamada -- nao
// existe lote nem recibo; a resposta ja e o resultado do processamento,
// lido com o mesmo leitor do retorno de consulta (retNFGas e um
// retConsSitNFGas renomeado, e o ACBr faz o mesmo StringReplace).
//
// Todos os metodos recebem context.Context: as chamadas sao de rede e o
// contrato e cancelavel.

// tagDadosMsg e o elemento do Body SOAP (FPBodyElement do ACBr).
const tagDadosMsg = "nfgasDadosMsg"

// clienteSOAP monta o cliente com o certificado e o timeout configurados.
func (c *Componente) clienteSOAP() (*dfe.ClienteSOAP, error) {
	if c.Configuracoes.Certificado == nil {
		return nil, ErrCertificadoObrigatorio
	}
	return dfe.NovoClienteSOAP(c.Configuracoes.Certificado, c.Configuracoes.Timeout), nil
}

func (c *Componente) chamar(ctx context.Context, servico Servico, dadosMsg string) (string, error) {
	cliente, err := c.clienteSOAP()
	if err != nil {
		return "", err
	}
	url := c.Configuracoes.URLs[servico]
	if url == "" {
		url, err = URLServico(c.Configuracoes.UF, c.Configuracoes.Ambiente, servico)
		if err != nil {
			return "", err
		}
	}
	return cliente.Chamar(ctx, url, NamespaceServico(servico), SoapAction(servico),
		tagDadosMsg, "nfgasResultMsg", dadosMsg)
}

// versaoServico devolve a versao do leiaute como string ("1.00").
func (c *Componente) versaoServico() string {
	return c.Configuracoes.VersaoDF.String()
}

// PrepararEnvio gera, prepara o gRespTec (hashCSRT) e assina o XML de uma
// nota, gravando em nota.XMLAssinado. E o trecho de TNotasFiscais.Assinar
// que antecede a transmissao; util tambem para inspecionar o XML antes de
// enviar.
func (c *Componente) PrepararEnvio(nota *NotaFiscal) (string, error) {
	if nota == nil || nota.NFGas == nil {
		return "", ErrXMLVazio
	}
	n := nota.NFGas

	// hashCSRT vem da configuracao (RespTec), como no ACBr; a chave e
	// necessaria, entao monta antes.
	if c.Configuracoes.IDCSRT != 0 && c.Configuracoes.CSRT != "" && n.InfRespTec.CNPJ != "" {
		n.InfRespTec.IDCSRT = c.Configuracoes.IDCSRT
		chave, err := MontarChaveAcesso(n)
		if err != nil {
			return "", err
		}
		n.InfRespTec.HashCSRT = dfe.HashCSRT(c.Configuracoes.CSRT, chave)
	}

	// transmissao envia a NFGas SOLTA -- o protocolo so existe depois
	prot := n.ProcNFGas
	n.ProcNFGas = pcn.ProcDFe{}
	xml, err := GerarXML(n)
	n.ProcNFGas = prot
	if err != nil {
		return "", err
	}

	assinado, err := Assinar(c.Configuracoes.Certificado, xml)
	if err != nil {
		return "", err
	}
	nota.XMLAssinado = assinado

	// popula n.Signature com a assinatura recem-criada: e o que permite a
	// GerarXMLProc reembutir o bloco Signature ao montar o nfgasProc de
	// distribuicao (no ACBr o proc e montado a partir do proprio
	// XMLAssinado -- ACBrNFGasWebServices.pas:763).
	if doc, err := pcn.ParseString(assinado); err == nil {
		pcn.LerSignature(doc.Root.FindAnyNs("Signature"), &n.Signature)
	}
	return assinado, nil
}

// Enviar transmite UM documento a SEFAZ (recepcao sincrona da NFGas) e
// devolve o resultado do processamento. Se a nota ainda nao tem
// XMLAssinado, gera e assina antes (PrepararEnvio). Quando autorizada
// (cStat 100), o protocolo e anexado a nota (ProcNFGas), pronto para
// GerarXMLProc. Porte de TNFGasRecepcao.
func (c *Componente) Enviar(ctx context.Context, nota *NotaFiscal) (*RetConsSitNFGas, error) {
	if nota == nil || nota.NFGas == nil {
		return nil, ErrXMLVazio
	}
	if nota.XMLAssinado == "" {
		if _, err := c.PrepararEnvio(nota); err != nil {
			return nil, err
		}
	}

	// DadosMsg = <NFGas ...>...</NFGas> comprimido e codificado
	dados, err := dfe.GzipBase64(extrairNFGas(nota.XMLAssinado))
	if err != nil {
		return nil, err
	}
	// limite de 1 MB do TNFGasRecepcao.DefinirDadosMsg
	if len(dados) > 1024*1024 {
		return nil, fmt.Errorf("nfgas: XML de dados superior a 1 Mbyte apos compressao (%d Kbytes)", len(dados)/1024)
	}

	resposta, err := c.chamar(ctx, ServicoRecepcao, dados)
	if err != nil {
		return nil, err
	}

	// retNFGas / retConsReciNFGas -> retConsSitNFGas (TratarResposta)
	resposta = reRetNFGas.ReplaceAllString(resposta, "${1}retConsSitNFGas")
	ret, err := LerRetConsSitString(resposta)
	if err != nil {
		return nil, err
	}

	// ValidarDigest do TratarResposta (ACBrNFGasWebServices.pas:747-752):
	// o digVal do protocolo tem que ser o DigestValue do XML transmitido.
	if ret.ProtNFGas.DigVal != "" && nota.NFGas.Signature.DigestValue != "" &&
		ret.ProtNFGas.DigVal != nota.NFGas.Signature.DigestValue {
		return ret, fmt.Errorf("%w: digVal do protocolo %q difere do DigestValue transmitido %q",
			ErrDigestDivergente, ret.ProtNFGas.DigVal, nota.NFGas.Signature.DigestValue)
	}

	// anexa o protocolo quando confirmado. No ACBr a condicao e
	// cStat=100 + CstatProcessado; na NFGas CstatProcessado == Confirmada
	// == {100, 150}, entao o efeito e o mesmo.
	if ret.ProtNFGas.NProt != "" && CStatConfirmada(ret.ProtNFGas.CStat) {
		nota.NFGas.ProcNFGas = ret.ProtNFGas
	}
	return ret, nil
}

// reRetNFGas troca o nome do elemento de retorno da recepcao pelo nome que
// o leitor de consulta conhece, preservando prefixos de fechamento -- o
// equivalente dos StringReplace com rfIgnoreCase do TratarResposta.
var reRetNFGas = regexp.MustCompile(`(?i)(</?)(?:retNFGas|retConsReciNFGas)\b`)

// Consultar consulta a situacao de um documento pela chave de acesso.
// Porte de TNFGasConsulta.
func (c *Componente) Consultar(ctx context.Context, chave string) (*RetConsSitNFGas, error) {
	chave = pcn.RemoverLiteralChave(strings.TrimSpace(chave))
	if err := pcn.ValidarChaveAcesso(chave); err != nil {
		return nil, err
	}

	// TConsSitNFGas.GerarXML
	dadosMsg := `<consSitNFGas xmlns="` + Namespace + `" versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Configuracoes.Ambiente.String() + "</tpAmb>" +
		"<xServ>CONSULTAR</xServ>" +
		"<chNFGas>" + chave + "</chNFGas>" +
		"</consSitNFGas>"

	resposta, err := c.chamar(ctx, ServicoConsulta, dadosMsg)
	if err != nil {
		return nil, err
	}
	return LerRetConsSitString(resposta)
}

// StatusServico consulta a disponibilidade do servico da SEFAZ.
// Porte de TNFGasStatusServico; a mensagem consStatServNFGas NAO leva cUF
// (TConsStatServ criado com AGerarcUF=False).
func (c *Componente) StatusServico(ctx context.Context) (*RetConsStatServNFGas, error) {
	dadosMsg := `<consStatServNFGas xmlns="` + Namespace + `" versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Configuracoes.Ambiente.String() + "</tpAmb>" +
		"<xServ>STATUS</xServ>" +
		"</consStatServNFGas>"

	resposta, err := c.chamar(ctx, ServicoStatusServico, dadosMsg)
	if err != nil {
		return nil, err
	}
	return LerRetConsStatServ(resposta)
}

// EnviarEvento gera, assina e transmite um evento (sem gzip -- so a
// recepcao de NFGas comprime). Porte de TNFGasEnvEvento. O retorno traz o
// evento enviado e o retEvento do processamento; cStat 135/136/155 e
// evento registrado.
func (c *Componente) EnviarEvento(ctx context.Context, e *EventoNFGas) (*RetEventoNFGas, error) {
	if e == nil {
		return nil, ErrXMLVazio
	}
	if e.InfEvento.TpAmb != c.Configuracoes.Ambiente {
		e.InfEvento.TpAmb = c.Configuracoes.Ambiente
	}
	if e.Versao == "" {
		e.Versao = c.versaoServico()
	}

	xml, err := GerarXMLEvento(e)
	if err != nil {
		return nil, err
	}
	if c.Configuracoes.Certificado == nil {
		return nil, ErrCertificadoObrigatorio
	}
	assinado, err := AssinarEvento(c.Configuracoes.Certificado, xml)
	if err != nil {
		return nil, err
	}

	resposta, err := c.chamar(ctx, ServicoRecepcaoEvento, assinado)
	if err != nil {
		return nil, err
	}
	return LerEventoString(resposta)
}

// Cancelamento monta e envia o evento de cancelamento de um documento.
// Porte de TACBrNFGas.Cancelamento.
func (c *Componente) Cancelamento(ctx context.Context, chave, protocolo, justificativa string) (*RetEventoNFGas, error) {
	evento := &EventoNFGas{
		InfEvento: InfEvento{
			ChNFGas:    pcn.RemoverLiteralChave(strings.TrimSpace(chave)),
			TpAmb:      c.Configuracoes.Ambiente,
			DhEvento:   time.Now(),
			TpEvento:   TeCancelamento,
			NSeqEvento: 1,
			DetEvento: DetEvento{
				NProt: protocolo,
				XJust: justificativa,
			},
		},
	}
	return c.EnviarEvento(ctx, evento)
}

// URLConsultaNFGas devolve a URL publica de consulta do documento.
// Porte de TACBrNFGas.GetURLConsultaNFGas.
func (c *Componente) URLConsultaNFGas(cUF int) (string, error) {
	return URLConsultaNFGas(cUF)
}

// GerarQRCode monta a URL do QR-Code, assinando a chave com o certificado
// configurado quando a emissao e offline. Porte de TACBrNFGas.GetURLQRCode.
func (c *Componente) GerarQRCode(n *NFGas) (string, error) {
	var assinar func(string) (string, error)
	if cert := c.Configuracoes.Certificado; cert != nil {
		assinar = cert.AssinarSHA1Base64
	}
	return GerarQRCode(n, assinar)
}

// extrairNFGas recorta <NFGas ...>...</NFGas> do XML assinado -- o
// RetornarConteudoEntre('<NFGas', '</NFGas>') do original, que descarta um
// eventual envelope nfgasProc.
func extrairNFGas(xml string) string {
	ini := strings.Index(xml, "<NFGas")
	fim := strings.LastIndex(xml, "</NFGas>")
	if ini < 0 || fim < 0 || fim < ini {
		return xml
	}
	return xml[ini : fim+len("</NFGas>")]
}
