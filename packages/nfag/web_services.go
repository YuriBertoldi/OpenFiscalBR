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

package nfag

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
// ACBrNFAgWebServices.pas sobre o cliente SOAP de packages/dfe.
//
// A recepcao da NFAg e SINCRONA e aceita UM documento por chamada -- nao
// existe lote nem recibo; a resposta ja e o resultado do processamento,
// lido com o mesmo leitor do retorno de consulta (retNFAg e um
// retConsSitNFAg renomeado, e o ACBr faz o mesmo StringReplace).
//
// Todos os metodos recebem context.Context: as chamadas sao de rede e o
// contrato e cancelavel.

// tagDadosMsg e o elemento do Body SOAP (FPBodyElement do ACBr).
const tagDadosMsg = "nfagDadosMsg"

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
		tagDadosMsg, "nfagResultMsg", dadosMsg)
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
	if nota == nil || nota.NFAg == nil {
		return "", ErrXMLVazio
	}
	n := nota.NFAg

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

	// transmissao envia a NFAg SOLTA -- o protocolo so existe depois
	prot := n.ProcNFAg
	n.ProcNFAg = pcn.ProcDFe{}
	xml, err := GerarXML(n)
	n.ProcNFAg = prot
	if err != nil {
		return "", err
	}

	assinado, err := Assinar(c.Configuracoes.Certificado, xml)
	if err != nil {
		return "", err
	}
	nota.XMLAssinado = assinado

	// popula n.Signature com a assinatura recem-criada: e o que permite a
	// GerarXMLProc reembutir o bloco Signature ao montar o nfagProc de
	// distribuicao (no ACBr o proc e montado a partir do proprio
	// XMLAssinado -- ACBrNFAgWebServices.pas:763).
	if doc, err := pcn.ParseString(assinado); err == nil {
		pcn.LerSignature(doc.Root.FindAnyNs("Signature"), &n.Signature)
	}
	return assinado, nil
}

// Enviar transmite UM documento a SEFAZ (recepcao sincrona da NFAg) e
// devolve o resultado do processamento. Se a nota ainda nao tem
// XMLAssinado, gera e assina antes (PrepararEnvio). Quando autorizada
// (cStat 100), o protocolo e anexado a nota (ProcNFAg), pronto para
// GerarXMLProc. Porte de TNFAgRecepcao.
func (c *Componente) Enviar(ctx context.Context, nota *NotaFiscal) (*RetConsSitNFAg, error) {
	if nota == nil || nota.NFAg == nil {
		return nil, ErrXMLVazio
	}
	if nota.XMLAssinado == "" {
		if _, err := c.PrepararEnvio(nota); err != nil {
			return nil, err
		}
	}

	// DadosMsg = <NFAg ...>...</NFAg> comprimido e codificado
	dados, err := dfe.GzipBase64(extrairNFAg(nota.XMLAssinado))
	if err != nil {
		return nil, err
	}
	// limite de 1 MB do TNFAgRecepcao.DefinirDadosMsg
	if len(dados) > 1024*1024 {
		return nil, fmt.Errorf("nfag: XML de dados superior a 1 Mbyte apos compressao (%d Kbytes)", len(dados)/1024)
	}

	resposta, err := c.chamar(ctx, ServicoRecepcao, dados)
	if err != nil {
		return nil, err
	}

	// retNFAg / retConsReciNFAg -> retConsSitNFAg (TratarResposta)
	resposta = reRetNFAg.ReplaceAllString(resposta, "${1}retConsSitNFAg")
	ret, err := LerRetConsSitString(resposta)
	if err != nil {
		return nil, err
	}

	// ValidarDigest do TratarResposta (ACBrNFAgWebServices.pas:747-752):
	// o digVal do protocolo tem que ser o DigestValue do XML transmitido.
	// Mais tolerante que o original: o Delphi compara sempre que o protocolo
	// traz digVal; aqui a comparacao so ocorre com a Signature local
	// preenchida (nota transmitida sem passar por Enviar/Assinar nao tem
	// DigestValue para comparar).
	if ret.ProtNFAg.DigVal != "" && nota.NFAg.Signature.DigestValue != "" &&
		ret.ProtNFAg.DigVal != nota.NFAg.Signature.DigestValue {
		return ret, fmt.Errorf("%w: digVal do protocolo %q difere do DigestValue transmitido %q",
			ErrDigestDivergente, ret.ProtNFAg.DigVal, nota.NFAg.Signature.DigestValue)
	}

	// anexa o protocolo quando autorizado. Na NFAg a condicao do
	// TratarResposta e cStat do RETORNO = 104 ("Lote processado") E
	// CstatProcessado no protocolo -- diferente da NFGas, que checa 100
	// no retorno (ACBrNFAgWebServices.pas, TNFAgRecepcao.TratarResposta).
	if ret.CStat == cStatRetornoSincronoOK && ret.ProtNFAg.NProt != "" &&
		CStatProcessado(ret.ProtNFAg.CStat) {
		nota.NFAg.ProcNFAg = ret.ProtNFAg
	}
	return ret, nil
}

// reRetNFAg troca o nome do elemento de retorno da recepcao pelo nome que
// o leitor de consulta conhece, preservando prefixos de fechamento -- o
// equivalente dos StringReplace com rfIgnoreCase do TratarResposta.
var reRetNFAg = regexp.MustCompile(`(?i)(</?)(?:retNFAg|retConsReciNFAg)\b`)

// Consultar consulta a situacao de um documento pela chave de acesso.
// Porte de TNFAgConsulta.
func (c *Componente) Consultar(ctx context.Context, chave string) (*RetConsSitNFAg, error) {
	chave = pcn.RemoverLiteralChave(strings.TrimSpace(chave))
	if err := pcn.ValidarChaveAcesso(chave); err != nil {
		return nil, err
	}

	// TConsSitNFAg.GerarXML
	dadosMsg := `<consSitNFAg xmlns="` + Namespace + `" versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Configuracoes.Ambiente.String() + "</tpAmb>" +
		"<xServ>CONSULTAR</xServ>" +
		"<chNFAg>" + chave + "</chNFAg>" +
		"</consSitNFAg>"

	resposta, err := c.chamar(ctx, ServicoConsulta, dadosMsg)
	if err != nil {
		return nil, err
	}
	return LerRetConsSitString(resposta)
}

// StatusServico consulta a disponibilidade do servico da SEFAZ.
// Porte de TNFAgStatusServico; a mensagem consStatServNFAg NAO leva cUF
// (TConsStatServ criado com AGerarcUF=False).
func (c *Componente) StatusServico(ctx context.Context) (*RetConsStatServNFAg, error) {
	dadosMsg := `<consStatServNFAg xmlns="` + Namespace + `" versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Configuracoes.Ambiente.String() + "</tpAmb>" +
		"<xServ>STATUS</xServ>" +
		"</consStatServNFAg>"

	resposta, err := c.chamar(ctx, ServicoStatusServico, dadosMsg)
	if err != nil {
		return nil, err
	}
	return LerRetConsStatServ(resposta)
}

// EnviarEvento gera, assina e transmite um evento (sem gzip -- so a
// recepcao de NFAg comprime). Porte de TNFAgEnvEvento. O retorno traz o
// evento enviado e o retEvento do processamento; cStat 135/136/155 e
// evento registrado.
func (c *Componente) EnviarEvento(ctx context.Context, e *EventoNFAg) (*RetEventoNFAg, error) {
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
// Porte de TACBrNFAg.Cancelamento. Diferente do original (ACBrNFAg.pas:
// 436-470), NAO consulta a situacao da nota antes para descobrir o
// protocolo: obter o nProt correto e responsabilidade do chamador (use
// Consultar se nao o tiver).
func (c *Componente) Cancelamento(ctx context.Context, chave, protocolo, justificativa string) (*RetEventoNFAg, error) {
	evento := &EventoNFAg{
		InfEvento: InfEvento{
			ChNFAg:     pcn.RemoverLiteralChave(strings.TrimSpace(chave)),
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

// URLConsultaNFAg devolve a URL publica de consulta do documento.
// Porte de TACBrNFAg.GetURLConsultaNFAg.
func (c *Componente) URLConsultaNFAg(cUF int) (string, error) {
	return URLConsultaNFAg(cUF)
}

// GerarQRCode monta a URL do QR-Code, assinando a chave com o certificado
// configurado quando a emissao e offline. Porte de TACBrNFAg.GetURLQRCode.
func (c *Componente) GerarQRCode(n *NFAg) (string, error) {
	var assinar func(string) (string, error)
	if cert := c.Configuracoes.Certificado; cert != nil {
		assinar = cert.AssinarSHA1Base64
	}
	return GerarQRCode(n, assinar)
}

// extrairNFAg recorta <NFAg ...>...</NFAg> do XML assinado -- o
// RetornarConteudoEntre('<NFAg', '</NFAg>') do original, que descarta um
// eventual envelope nfagProc.
func extrairNFAg(xml string) string {
	ini := strings.Index(xml, "<NFAg")
	fim := strings.LastIndex(xml, "</NFAg>")
	if ini < 0 || fim < 0 || fim < ini {
		return xml
	}
	return xml[ini : fim+len("</NFAg>")]
}
