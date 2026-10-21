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
	"fmt"
	"strconv"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Resolucao do endereco de um web service. Porte de LerServicoDeParams
// (ACBrDFe.pas:527) e de TACBrNFe.LerServicoDeParams (ACBrNFe.pas:511).
//
// A busca tem duas entradas que nao sao obvias:
//
//  1. a UF usada NAO e sempre a do emitente. Em contingencia SVC ela vira
//     "SVC-AN" ou "SVC-RS", porque o documento passa a ser autorizado pelo
//     ambiente virtual, nao pela SEFAZ de origem (ACBrNFe.pas:517-523);
//  2. na CONSULTA, a UF e o modelo saem da propria CHAVE do documento, nao
//     da configuracao -- consulta-se onde a nota foi autorizada
//     (ACBrNFeWebServices.pas:2082-2100).

// UFParaBusca devolve a unidade sob a qual o endereco do servico deve ser
// procurado, considerando a contingencia configurada.
//
// Porte do case de ACBrNFe.pas:517-523.
func UFParaBusca(cfg Configuracoes) string {
	switch cfg.FormaEmissao {
	case pcn.TeSVCAN:
		return "SVC-AN"
	case pcn.TeSVCRS:
		return "SVC-RS"
	default:
		return cfg.UF
	}
}

// URLServico devolve o endereco do web service.
//
// `uf` aceita tanto a sigla de um estado quanto "SVC-AN", "SVC-RS" ou "AN";
// use UFParaBusca para obte-la a partir da configuracao.
func URLServico(modelo ModeloDF, uf string, ambiente pcn.TipoAmbiente, servico Servico, versao VersaoDF) (string, error) {
	autorizador, err := autorizadorDe(modelo, uf)
	if err != nil {
		return "", err
	}

	servicos, ok := endpointsPorAutorizador[prefixoDoModelo(modelo)+"/"+autorizador]
	if !ok {
		return "", fmt.Errorf("%w (autorizador %s, modelo %s)",
			ErrSemURL, autorizador, modelo)
	}

	endpoint, ok := servicos[servico]
	if !ok {
		return "", fmt.Errorf("%w (%s nao atende o servico %s)",
			ErrSemURL, autorizador, servico)
	}

	url := endpoint.para(versao, ambiente)
	if url == "" {
		return "", fmt.Errorf("%w (%s nao atende %s na versao %s)",
			ErrSemURL, autorizador, servico, versao)
	}
	return url, nil
}

// autorizadorDe diz qual ambiente atende a UF no modelo informado.
func autorizadorDe(modelo ModeloDF, uf string) (string, error) {
	tabela := autorizadorNFe
	if modelo == MoNFCe {
		tabela = autorizadorNFCe
	}
	if a, ok := tabela[uf]; ok {
		return a, nil
	}
	return "", fmt.Errorf("%w (UF %q desconhecida para o modelo %s)",
		ErrSemURL, uf, modelo)
}

// prefixoDoModelo e o ModeloDFToPrefixo do ACBr: a parte do nome da secao
// do `.ini` que distingue NF-e de NFC-e.
func prefixoDoModelo(modelo ModeloDF) string {
	if modelo == MoNFCe {
		return "NFCe"
	}
	return "NFe"
}

// ModeloDaChave devolve o modelo gravado na chave de acesso, caindo para o
// configurado quando a chave nao e informada.
//
// Porte de TNFeWebService.ModeloDFe (ACBrNFeWebServices.pas:698): a chave
// manda sobre a configuracao, porque consulta-se o documento onde ele
// existe, nao onde o componente esta apontado.
func ModeloDaChave(chave string, padrao ModeloDF) ModeloDF {
	chave = pcn.RemoverLiteralChave(chave)
	if len(chave) < 22 {
		return padrao
	}
	switch chave[20:22] {
	case "55":
		return MoNFe
	case "65":
		return MoNFCe
	default:
		return padrao
	}
}

// UFDaChave devolve a sigla da UF cujo codigo esta nas duas primeiras
// posicoes da chave. Porte de ExtrairUFChaveAcesso.
func UFDaChave(chave string) string {
	chave = pcn.RemoverLiteralChave(chave)
	if len(chave) < 2 {
		return ""
	}
	cUF, err := strconv.Atoi(chave[:2])
	if err != nil {
		return ""
	}
	return pcn.SiglaUF(cUF)
}

// comOverrideDoIni devolve a SoapAction efetiva, aplicando o override do
// `.ini` SO quando ele pode vencer.
//
// Porte do guard `if EstaVazio(FPSoapAction)`, que existe apenas no ramo da
// 4.00 de todo DefinirServicoEAction (ACBrNFeWebServices.pas:820, :2124,
// :2089, :2697). Abaixo da 4.00 o codigo atribui a SoapAction SEM guard, e
// o que o LerServicoDeParams trouxe do `.ini` e simplesmente descartado.
//
// A distincao nao muda nada hoje -- as unicas secoes `[SOAP_V4*]` do `.ini`
// declaram so chaves `_4.00` --, mas e a condicional que governa a
// atribuicao, e e justamente esse tipo de condicional que se perde num
// porte. Materializa-la deixa o dia em que a SEFAZ publicar um override de
// 3.10 ser um caso ja tratado, e nao uma regressao silenciosa.
func comOverrideDoIni(soapAction, doIni string, versao VersaoDF) string {
	if versao >= Ve400 && doIni != "" {
		return doIni
	}
	return soapAction
}

// SoapActionDoIni devolve o override de SoapAction declarado no `.ini`,
// ou vazio quando nao ha.
//
// Porte do efeito de `LerServicoDeParams` (ACBrDFe.pas:546). Quem decide se
// ele VENCE o valor montado pelo codigo e o comOverrideDoIni.
func SoapActionDoIni(modelo ModeloDF, uf string, ambiente pcn.TipoAmbiente, servico Servico, versao VersaoDF) string {
	autorizador, err := autorizadorDe(modelo, uf)
	if err != nil {
		return ""
	}
	servicos, ok := soapActionPorAutorizador[prefixoDoModelo(modelo)+"/"+autorizador]
	if !ok {
		return ""
	}
	return servicos[servico].para(versao, ambiente)
}
