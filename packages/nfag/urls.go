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
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Tabela de web services da NFAg -- porte de ACBrNFAgServicos.ini
// (embutido: o ACBr le o .ini em tempo de execucao; aqui a tabela e fonte).
//
// No .ini do ACBr, TODAS as UFs apontam para o SVRS, exceto MA e PA, que
// apontam para secoes NFAg_SVAN_P/H que NAO EXISTEM no arquivo -- lacuna
// herdada do ACBr. Para MA e PA toda operacao devolve ErrSemURL ate a SEFAZ
// (e o ACBr) publicarem os enderecos do SVAN.

// Servico identifica um web service da NFAg. Os literais sao os nomes das
// chaves do .ini e tambem o sufixo do namespace/acao SOAP.
type Servico string

const (
	ServicoRecepcao       Servico = "NFAgRecepcao"
	ServicoRecepcaoEvento Servico = "NFAgRecepcaoEvento"
	ServicoConsulta       Servico = "NFAgConsulta"
	ServicoStatusServico  Servico = "NFAgStatusServico"
)

// urlWsdBase e o prefixo do namespace/acao SOAP (GetUrlWsd do
// TDFeWebService: NameSpaceURI + "/wsdl/").
const urlWsdBase = Namespace + "/wsdl/"

// NamespaceServico devolve o namespace do elemento nfagDadosMsg
// (FPServico do ACBr).
func NamespaceServico(s Servico) string { return urlWsdBase + string(s) }

// SoapAction devolve a acao SOAP completa do servico
// (FPSoapAction do ACBr).
func SoapAction(s Servico) string {
	switch s {
	case ServicoStatusServico:
		return urlWsdBase + "NFAgStatusServico/nfagStatusServicoNF"
	case ServicoRecepcao:
		return urlWsdBase + "NFAgRecepcao/nfagRecepcao"
	case ServicoConsulta:
		return urlWsdBase + "NFAgConsulta/nfagConsultaNF"
	case ServicoRecepcaoEvento:
		return urlWsdBase + "NFAgRecepcaoEvento/nfagRecepcaoEvento"
	}
	return ""
}

var urlsSVRS = map[pcn.TipoAmbiente]map[Servico]string{
	pcn.TaProducao: {
		ServicoRecepcao:       "https://nfag.svrs.rs.gov.br/ws/NFAgRecepcao/NFAgRecepcao.asmx",
		ServicoRecepcaoEvento: "https://nfag.svrs.rs.gov.br/ws/NFAgRecepcaoEvento/NFAgRecepcaoEvento.asmx",
		ServicoConsulta:       "https://nfag.svrs.rs.gov.br/ws/NFAgConsulta/NFAgConsulta.asmx",
		ServicoStatusServico:  "https://nfag.svrs.rs.gov.br/ws/NFAgStatusServico/NFAgStatusServico.asmx",
	},
	pcn.TaHomologacao: {
		ServicoRecepcao:       "https://nfag-homologacao.svrs.rs.gov.br/ws/NFAgRecepcao/NFAgRecepcao.asmx",
		ServicoRecepcaoEvento: "https://nfag-homologacao.svrs.rs.gov.br/ws/NFAgRecepcaoEvento/NFAgRecepcaoEvento.asmx",
		ServicoConsulta:       "https://nfag-homologacao.svrs.rs.gov.br/ws/NFAgConsulta/NFAgConsulta.asmx",
		ServicoStatusServico:  "https://nfag-homologacao.svrs.rs.gov.br/ws/NFAgStatusServico/NFAgStatusServico.asmx",
	},
}

// as duas URLs de portal valem para producao e homologacao (mesmo valor nas
// secoes NFAg_SVRS_P e NFAg_SVRS_H do .ini)
const (
	urlQRCodeSVRS   = "https://dfe-portal.svrs.rs.gov.br/nfag/qrCode"
	urlConsultaSVRS = "https://dfe-portal.svrs.rs.gov.br/nfag/Consulta"
)

// ufSemURL sao as UFs que o .ini aponta para o SVAN, sem URLs definidas.
var ufSemURL = map[string]bool{"MA": true, "PA": true}

// URLServico devolve a URL do web service para a UF e o ambiente.
// Porte de LerServicoDeParams sobre o ACBrNFAgServicos.ini.
func URLServico(uf string, ambiente pcn.TipoAmbiente, servico Servico) (string, error) {
	if ufSemURL[uf] {
		return "", fmt.Errorf("%w (UF %s aponta para o SVAN)", ErrSemURL, uf)
	}
	if pcn.CodigoUF(uf) == 0 {
		return "", fmt.Errorf("%w (UF %q desconhecida)", ErrSemURL, uf)
	}
	url := urlsSVRS[ambiente][servico]
	if url == "" {
		return "", fmt.Errorf("%w (servico %s)", ErrSemURL, servico)
	}
	return url, nil
}

// URLConsultaNFAg devolve a URL do portal de consulta publica.
// Porte de TACBrNFAg.GetURLConsultaNFAg.
func URLConsultaNFAg(cUF int) (string, error) {
	uf := pcn.SiglaUF(cUF)
	if ufSemURL[uf] || uf == "" {
		return "", fmt.Errorf("%w (URL-ConsultaNFAg, UF %q)", ErrSemURL, uf)
	}
	return urlConsultaSVRS, nil
}

// GerarQRCode monta a URL do QR-Code do documento.
// Porte de TACBrNFAg.GetURLQRCode: url + "?chNFAg=" + chave +
// "&tpAmb=" + ambiente e, em emissao offline (teOffLine), "&sign=" com o
// SHA-1 da chave assinado pelo certificado (RSA-SHA1 em base64).
//
// assinarChave so e usado no caso offline; passe nil na emissao normal.
func GerarQRCode(n *NFAg, assinarChave func(chave string) (string, error)) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}
	uf := pcn.SiglaUF(n.Ide.CUF)
	if ufSemURL[uf] || uf == "" {
		return "", fmt.Errorf("%w (URL-QRCode, UF %q)", ErrSemURL, uf)
	}

	chave := n.ChaveAcesso()
	url := urlQRCodeSVRS + "?chNFAg=" + chave + "&tpAmb=" + n.Ide.TpAmb.String()

	if n.Ide.TpEmis == pcn.TeOffLine {
		if assinarChave == nil {
			return "", ErrCertificadoObrigatorio
		}
		sign, err := assinarChave(chave)
		if err != nil {
			return "", err
		}
		url += "&sign=" + sign
	}
	return url, nil
}

// HashSHA1Base64 e o CalcHash(dgstSHA1, outBase64) SEM assinatura RSA --
// mantido para conferencia; o sign do QR-Code offline usa AssinarSHA1RSA
// do certificado (CalcHash com Assinar=True).
func HashSHA1Base64(s string) string {
	h := sha1.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(h[:])
}
