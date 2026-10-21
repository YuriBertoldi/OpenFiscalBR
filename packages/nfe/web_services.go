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
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Transmissao dos documentos para a SEFAZ. Porte de TNFeWebService e das
// classes que o especializam (ACBrNFeWebServices.pas).
//
// Cada servico resolve tres coisas antes de transmitir, e todas as tres
// dependem de dados que NAO estao no documento:
//
//  1. o ENDERECO, que sai da tabela por autorizador (urls_servicos.go);
//  2. o NOME DO SERVICO e a SoapAction, que mudam com a versao do leiaute
//     e, numa UF, com a propria UF -- a Bahia usa uma notacao propria;
//  3. a versaoDados do cabecalho SOAP, que e a versao do SERVICO, nao a do
//     documento.
//
// A quarta coisa, o envelope, e do packages/dfe: aqui so entra o que e
// especifico da NF-e.

// namespaceNFeServicos e o namespace que vai no dadosMsg de todo servico.
const namespaceNFeServicos = `xmlns="` + namespaceNFe + `"`

// Componente transmite documentos para a SEFAZ.
//
// Porte parcial de TACBrNFe: so a parte de web services. A geracao do XML,
// as regras de negocio e o QR-Code sao funcoes do package e nao precisam
// dele.
type Componente struct {
	Config Configuracoes

	// cliente e criado sob demanda, na primeira chamada.
	cliente *dfe.ClienteSOAP
}

// NovoComponente devolve um componente com a configuracao informada.
func NovoComponente(cfg Configuracoes) *Componente {
	return &Componente{Config: cfg}
}

// clienteSOAP devolve o cliente, criando-o na primeira chamada.
func (c *Componente) clienteSOAP() (*dfe.ClienteSOAP, error) {
	if c.cliente != nil {
		return c.cliente, nil
	}
	if c.Config.Certificado == nil {
		return nil, ErrCertificadoNaoConfigurado
	}
	c.cliente = dfe.NovoClienteSOAP(c.Config.Certificado, c.Config.Timeout)
	return c.cliente, nil
}

// versaoServico e a versao que vai no <versaoDados> do cabecalho SOAP.
//
// Porte de GerarVersaoDadosSoap (ACBrNFeWebServices.pas:740): e a versao do
// SERVICO, formatada com duas casas.
func (c *Componente) versaoServico() string {
	return pcn.FormatarVersaoXML(c.Config.VersaoDF.Float())
}

// enderecoWsd e o GetUrlWsd do ACBr: a raiz dos nomes de servico SOAP.
const enderecoWsd = "http://www.portalfiscal.inf.br/nfe/wsdl/"

// StatusServico consulta a disponibilidade do web service.
// Porte de TNFeStatusServico.
//
// O `agora` serve a correcao de horario do SVRS; passe o instante da
// chamada.
func (c *Componente) StatusServico(ctx context.Context, agora time.Time) (*RetConsStatServ, error) {
	url, err := URLServico(c.Config.ModeloDF, UFParaBusca(c.Config),
		c.Config.Ambiente, ServicoStatus, c.Config.VersaoDF)
	if err != nil {
		return nil, err
	}
	return c.statusServicoNaURL(ctx, url, agora)
}

// statusServicoNaURL e o StatusServico com o endereco JA RESOLVIDO.
//
// A separacao existe para o teste poder apontar a chamada a um servidor
// local sem precisar de uma entrada falsa na tabela de endpoints -- e, de
// quebra, deixa claro que a escolha do endereco e decidida antes de
// qualquer conexao.
func (c *Componente) statusServicoNaURL(ctx context.Context, url string, agora time.Time) (*RetConsStatServ, error) {
	servico, soapAction := c.servicoEActionDoStatus()
	if doIni := SoapActionDoIni(c.Config.ModeloDF, UFParaBusca(c.Config),
		c.Config.Ambiente, ServicoStatus, c.Config.VersaoDF); doIni != "" {
		soapAction = doIni
	}

	dadosMsg := "<consStatServ " + namespaceNFeServicos +
		` versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Config.Ambiente.String() + "</tpAmb>" +
		"<cUF>" + strconv.Itoa(c.Config.CodigoUF) + "</cUF>" +
		"<xServ>STATUS</xServ>" +
		"</consStatServ>"

	resposta, err := c.chamar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		DadosMsg:  dadosMsg,
		Cabecalho: c.cabecalhoSOAP(c.Config.CodigoUF),
		TagsResultado: []string{"nfeStatusServicoNF2Result",
			"NfeStatusServicoNFResult", "nfeResultMsg"},
	})
	if err != nil {
		return nil, err
	}

	ret, err := LerRetConsStatServ(resposta)
	if err != nil {
		return nil, err
	}
	ret.DhRecbto = AjustarHoraDoSVRS(ret.DhRecbto, url, agora)
	return ret, nil
}

// servicoEActionDoStatus devolve o nome do servico SOAP e a SoapAction.
// Porte de TNFeStatusServico.DefinirServicoEAction
// (ACBrNFeWebServices.pas:810).
//
// Tres formas, e a do meio existe por causa de UMA UF.
func (c *Componente) servicoEActionDoStatus() (servico, soapAction string) {
	if c.Config.VersaoDF >= Ve400 {
		servico = enderecoWsd + "NFeStatusServico4"
		return servico, servico + "/nfeStatusServicoNF"
	}

	// A Bahia usa uma notacao de servicos diferente das demais UFs -- so
	// para a NF-e, so na 3.10 e so fora de contingencia.
	if c.Config.CodigoUF == 29 && c.Config.ModeloDF == MoNFe &&
		c.Config.VersaoDF == Ve310 && c.Config.FormaEmissao == pcn.TeNormal {
		servico = enderecoWsd + "NfeStatusServico"
		return servico, servico + "/NfeStatusServicoNF"
	}

	// Abaixo da 4.00 a SoapAction e o PROPRIO nome do servico, sem sufixo.
	servico = enderecoWsd + "NfeStatusServico2"
	return servico, servico
}

// Consultar consulta a situacao de uma NF-e pela chave de acesso.
// Porte de TNFeConsulta.
//
// A UF e o modelo saem da CHAVE, nao da configuracao: consulta-se onde a
// nota foi autorizada (ACBrNFeWebServices.pas:2082-2100). A excecao e a
// contingencia SVC, que desvia a consulta para o ambiente virtual.
func (c *Componente) Consultar(ctx context.Context, chave string) (*RetConsSitNFe, error) {
	chave = pcn.RemoverLiteralChave(strings.TrimSpace(chave))
	if err := pcn.ValidarChaveAcesso(chave); err != nil {
		return nil, fmt.Errorf("nfe: chave %q invalida: %w", chave, err)
	}

	modelo := ModeloDaChave(chave, c.Config.ModeloDF)

	uf := UFDaChave(chave)
	switch c.Config.FormaEmissao {
	case pcn.TeSVCAN:
		uf = "SVC-AN"
	case pcn.TeSVCRS:
		uf = "SVC-RS"
	}

	url, err := URLServico(modelo, uf, c.Config.Ambiente, ServicoConsultaProtocolo, c.Config.VersaoDF)
	if err != nil {
		return nil, err
	}
	return c.consultarNaURL(ctx, url, chave)
}

// consultarNaURL e o Consultar com o endereco JA RESOLVIDO e a chave JA
// VALIDADA. Ver a nota em statusServicoNaURL.
func (c *Componente) consultarNaURL(ctx context.Context, url, chave string) (*RetConsSitNFe, error) {
	servico, soapAction := c.servicoEActionDaConsulta()
	cUF := pcn.CodigoUF(UFDaChave(chave))

	// O override do `.ini` vence o que o DefinirServicoEAction montaria.
	// E por ele que a consulta da Bahia na 4.00 usa "NFeConsulta4" e nao
	// "NFeConsultaProtocolo4" -- ver soapActionPorAutorizador.
	if doIni := SoapActionDoIni(ModeloDaChave(chave, c.Config.ModeloDF),
		UFDaChave(chave), c.Config.Ambiente, ServicoConsultaProtocolo,
		c.Config.VersaoDF); doIni != "" {
		soapAction = doIni
	}

	// PENDENTE: o ACBr usa o tpAmb da NOTA CARREGADA quando ha uma, e so
	// cai para a configuracao quando nao ha (ACBrNFeWebServices.pas:2086).
	// O Componente ainda nao carrega notas; quando a etapa de autorizacao
	// trouxer isso, a precedencia tem de voltar.
	dadosMsg := "<consSitNFe " + namespaceNFeServicos +
		` versao="` + c.versaoServico() + `">` +
		"<tpAmb>" + c.Config.Ambiente.String() + "</tpAmb>" +
		"<xServ>CONSULTAR</xServ>" +
		"<chNFe>" + chave + "</chNFe>" +
		"</consSitNFe>"

	resposta, err := c.chamar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		DadosMsg: dadosMsg,
		// O cUF do cabecalho vem da CHAVE, nao da configuracao
		// (TNFeConsulta.GerarUFSoap, ACBrNFeWebServices.pas:2169).
		Cabecalho: c.cabecalhoSOAP(cUF),
		TagsResultado: []string{"NfeConsultaNF2Result",
			"NfeConsultaNFResult", "nfeResultMsg"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetConsSitNFe(resposta)
}

// servicoEActionDaConsulta e o par do servicoEActionDoStatus.
// Porte de TNFeConsulta.DefinirServicoEAction (ACBrNFeWebServices.pas:2115).
func (c *Componente) servicoEActionDaConsulta() (servico, soapAction string) {
	if c.Config.VersaoDF >= Ve400 {
		servico = enderecoWsd + "NFeConsultaProtocolo4"
		return servico, servico + "/nfeConsultaNF"
	}

	if c.Config.CodigoUF == 29 && c.Config.ModeloDF == MoNFe &&
		c.Config.VersaoDF == Ve310 && c.Config.FormaEmissao == pcn.TeNormal {
		servico = enderecoWsd + "NfeConsulta"
	} else {
		servico = enderecoWsd + "NfeConsulta2"
	}
	return servico, servico
}

// tagCabecalhoNFe e o elemento do <soap12:Header>, usado so abaixo da 4.00.
const tagCabecalhoNFe = "nfeCabecMsg"

// cabecalhoSOAP monta o conteudo do header: cUF e versaoDados.
//
// Da versao 4.00 em diante a NF-e NAO manda header nenhum
// (ACBrNFeWebServices.pas:692) -- devolver vazio aqui faz o envelope sair
// sem ele.
//
// O cUF e parametro, e nao sai da configuracao, porque a CONSULTA usa o da
// chave do documento (TNFeConsulta.GerarUFSoap, :2169).
func (c *Componente) cabecalhoSOAP(cUF int) string {
	if c.Config.VersaoDF >= Ve400 {
		return ""
	}
	return "<cUF>" + strconv.Itoa(cUF) + "</cUF>" +
		"<versaoDados>" + c.versaoServico() + "</versaoDados>"
}

// chamar transmite o dadosMsg e extrai o conteudo da resposta.
//
// `tagsResultado` sao os nomes que a SEFAZ pode usar para envolver o
// retorno -- eles mudam por versao de servico e por UF. O cliente tenta
// todos sobre a MESMA resposta, como o SeparaDadosArray do ACBr.
func (c *Componente) chamar(ctx context.Context, req dfe.Requisicao) (string, error) {
	cliente, err := c.clienteSOAP()
	if err != nil {
		return "", err
	}
	req.TagDadosMsg = "nfeDadosMsg"
	req.TagCabecalho = tagCabecalhoNFe
	return cliente.Executar(ctx, req)
}
