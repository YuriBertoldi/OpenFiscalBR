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
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// URLs publicas e QR-Code da NFC-e.
//
// Porte de TACBrNFe.GetURLConsultaNFCe e das tres versoes de
// GetURLQRCode (ACBrNFe.pas:535-790). A tabela de URLs mora em
// urls_tabela.go.
//
// O QR-Code e o unico artefato do package que depende do CERTIFICADO, e so
// na versao 3: as versoes 1 e 2 fecham com um SHA-1 do texto mais o CSC, e a
// 3 troca isso por uma ASSINATURA RSA. Para nao arrastar o certificado para
// dentro do leiaute, a assinatura entra como funcao -- mesmo arranjo do
// nfgas.GerarQRCode.

// urlPorAmbiente guarda as duas URLs de um servico numa UF.
type urlPorAmbiente struct {
	Producao    string
	Homologacao string
}

// para devolve a URL do ambiente pedido.
func (u urlPorAmbiente) para(ambiente pcn.TipoAmbiente) string {
	if ambiente == pcn.TaProducao {
		return u.Producao
	}
	return u.Homologacao
}

// URLQRCode devolve a URL base do QR-Code para o modelo, a UF e o ambiente.
//
// O modelo escolhe a tabela como no ACBr: o 55 lê a secao NFe_<UF> e o 65 a
// NFCe_<UF> (ACBrNFe.pas:656-659). As duas so divergem em BA e PR.
func URLQRCode(modelo int, cUF int, ambiente pcn.TipoAmbiente) (string, error) {
	uf := pcn.SiglaUF(cUF)
	if uf == "" {
		return "", fmt.Errorf("%w (URL-QRCode, codigo de UF %d)", ErrSemURL, cUF)
	}

	if modelo == 55 {
		if u, ok := urlsQRCodeModelo55[uf]; ok {
			return u.para(ambiente), nil
		}
	}

	u, ok := urlsQRCode[uf]
	if !ok {
		return "", fmt.Errorf("%w (URL-QRCode, UF %s)", ErrSemURL, uf)
	}
	return u.para(ambiente), nil
}

// URLConsulta devolve a URL do portal de consulta publica, que vai no
// urlChave do grupo infNFeSupl a partir do leiaute 4.00.
//
// Porte de GetURLConsultaNFCe (ACBrNFe.pas:535). Duas sutilezas do original:
//
//   - quem escolhe a tabela e o ModeloDF da CONFIGURACAO, nao o modelo do
//     documento;
//   - a versao usada na busca e a do QR-CODE, nao a do leiaute -- e e por
//     isso que a URL da NFC-e muda entre a versao 1 e as demais.
//
// A versao recebida ja deve vir AJUSTADA (por AjustarVersaoQrCode); o
// original a ajusta internamente (ACBrNFe.pas:543), e GerarQRCode faz o
// mesmo antes de chamar aqui.
func URLConsulta(modelo ModeloDF, cUF int, ambiente pcn.TipoAmbiente, versaoQRCode VersaoQrCode) (string, error) {
	uf := pcn.SiglaUF(cUF)
	if uf == "" {
		return "", fmt.Errorf("%w (URL-Consulta, codigo de UF %d)", ErrSemURL, cUF)
	}

	tabela := urlsConsultaNFCe
	servico := "URL-ConsultaNFCe"
	switch {
	case modelo == MoNFe:
		tabela, servico = urlsConsultaNFe, "URL-ConsultaNFe"
	case versaoQRCode == Veqr100:
		tabela = urlsConsultaNFCeQRCodeV1
	}

	u, ok := tabela[uf]
	if !ok {
		return "", fmt.Errorf("%w (%s, UF %s)", ErrSemURL, servico, uf)
	}
	return u.para(ambiente), nil
}

// AssinarSHA1Base64 e a funcao que a versao 3 do QR-Code precisa: recebe o
// texto e devolve a assinatura RSA-SHA1 dele em base64.
//
// E o CalcHash(texto, dgstSHA1, outBase64, Assina=True) do ACBr
// (ACBrNFe.pas:777) -- note o ultimo argumento: NAO e um hash, e uma
// assinatura com a chave privada do certificado. Quem tem o certificado e o
// consumidor da biblioteca, entao ele passa a funcao.
type AssinarSHA1Base64 func(texto string) (string, error)

// GerarQRCode monta a URL do QR-Code do documento e devolve o texto que vai
// na tag qrCode. Porte de TACBrNFe.GetURLQRCode (ACBrNFe.pas:633).
//
// A versao efetiva passa por AjustarVersaoQrCode, que rebaixa a pedida
// conforme a versao do leiaute -- uma nota 3.10 nunca gera QR-Code versao 3,
// por mais que a configuracao peca.
//
// assinar so e usado na versao 3; nas versoes 1 e 2 pode ser nil.
func GerarQRCode(n *NFe, cfg Configuracoes, assinar AssinarSHA1Base64) (string, error) {
	if n == nil {
		return "", ErrDocumentoNulo
	}

	switch AjustarVersaoQrCode(cfg.VersaoQRCode, ParseVersaoDFFloat(n.InfNFe.Versao)) {
	case Veqr300:
		return gerarQRCodeV3(n, assinar)
	case Veqr200:
		return gerarQRCodeV2(n, cfg)
	default:
		return gerarQRCodeV1(n, cfg)
	}
}

// gerarQRCodeV1 monta o QR-Code em formato de query string, da NT 2015/002.
// Porte de GetURLQRCodeV1 (ACBrNFe.pas:651).
//
// E a unica versao que carrega o valor do ICMS e a data de emissao no proprio
// QR-Code, os dois em hexadecimal.
func gerarQRCodeV1(n *NFe, cfg Configuracoes) (string, error) {
	urlUF, err := URLQRCode(n.Ide.Modelo, n.Ide.CUF, n.Ide.TpAmb)
	if err != nil {
		return "", err
	}

	chave := pcn.RemoverLiteralChave(n.InfNFe.ID)
	uf := pcn.SiglaUF(n.Ide.CUF)

	dhEmiHex := asciiParaHex(pcn.FormatarDataHoraXML(n.Ide.DEmi, uf))
	digValHex := asciiParaHex(n.Signature.DigestValue)
	if ufDeHexMinusculo(n.Ide.CUF) {
		dhEmiHex = strings.ToLower(dhEmiHex)
		digValHex = strings.ToLower(digValHex)
	}

	// Na versao 1 o identificador do CSC vai com seis posicoes.
	idCSC := pcn.PreencherZerosEsquerda(strconv.Itoa(inteiroOuZero(cfg.IDCSC)), 6)
	csc := cscOuDerivado(cfg.CSC, chave, idCSC)

	entrada := "chNFe=" + chave +
		"&nVersao=100" +
		"&tpAmb=" + n.Ide.TpAmb.String() +
		destinatarioNaQuery(n) +
		"&dhEmi=" + dhEmiHex +
		"&vNF=" + pcn.FormatarDecimalXML(n.Total.ICMSTot.VNF, 2) +
		"&vICMS=" + pcn.FormatarDecimalXML(n.Total.ICMSTot.VICMS, 2) +
		"&digVal=" + digValHex +
		"&cIdToken="

	hash := asciiParaHex(string(sha1Bytes(entrada + idCSC + csc)))

	separador := "?"
	if strings.Contains(urlUF, "?") {
		separador = "&"
	}
	return urlUF + separador + entrada + idCSC + "&cHashQRCode=" + hash, nil
}

// gerarQRCodeV2 monta o QR-Code em formato de campos separados por barra
// vertical, da NT 2017/001. Porte de GetURLQRCodeV2 (ACBrNFe.pas:704).
//
// O bloco de contingencia (dia, valor e digest) SO entra quando a nota foi
// emitida offline -- e quem diz isso e o tpEmis lido da CHAVE, nao do ide.
func gerarQRCodeV2(n *NFe, cfg Configuracoes) (string, error) {
	urlUF, err := URLQRCode(n.Ide.Modelo, n.Ide.CUF, n.Ide.TpAmb)
	if err != nil {
		return "", err
	}

	chave := pcn.RemoverLiteralChave(n.InfNFe.ID)

	digValHex := asciiParaHex(n.Signature.DigestValue)
	if ufDeHexMinusculo(n.Ide.CUF) {
		digValHex = strings.ToLower(digValHex)
	}

	// Da versao 2 em diante o identificador do CSC vai SEM zeros a esquerda.
	idCSC := strconv.Itoa(inteiroOuZero(cfg.IDCSC))
	csc := cscOuDerivado(cfg.CSC, chave, idCSC)

	entrada := chave + "|2|" + n.Ide.TpAmb.String() + "|"
	if TipoEmissaoDaChave(chave) == 9 {
		entrada += fmt.Sprintf("%02d", n.Ide.DEmi.Day()) + "|" +
			pcn.FormatarDecimalXML(n.Total.ICMSTot.VNF, 2) + "|" +
			digValHex + "|"
	}

	hash := asciiParaHex(string(sha1Bytes(entrada + idCSC + csc)))

	return acrescentarSufixoP(urlUF) + entrada + idCSC + "|" + hash, nil
}

// gerarQRCodeV3 monta o QR-Code da NT 2020/006. Porte de GetURLQRCodeV3
// (ACBrNFe.pas:750).
//
// Diferente das duas anteriores em tres pontos: nao usa CSC nenhum, fecha com
// uma ASSINATURA do certificado em vez de um hash, e so acrescenta qualquer
// coisa alem do cabecalho quando a emissao foi offline.
func gerarQRCodeV3(n *NFe, assinar AssinarSHA1Base64) (string, error) {
	urlUF, err := URLQRCode(n.Ide.Modelo, n.Ide.CUF, n.Ide.TpAmb)
	if err != nil {
		return "", err
	}

	chave := pcn.RemoverLiteralChave(n.InfNFe.ID)
	entrada := chave + "|3|" + n.Ide.TpAmb.String()

	if TipoEmissaoDaChave(chave) == 9 {
		entrada += "|" + fmt.Sprintf("%02d", n.Ide.DEmi.Day()) +
			"|" + pcn.FormatarDecimalXML(n.Total.ICMSTot.VNF, 2) +
			"|" + tipoDestinatario(n) +
			"|" + destinatarioDoQRCode(n)

		if assinar == nil {
			return "", ErrCertificadoObrigatorio
		}
		assinatura, err := assinar(entrada)
		if err != nil {
			return "", err
		}
		entrada += "|" + assinatura
	}

	return acrescentarSufixoP(urlUF) + entrada, nil
}

// ufsComHexMinusculo sao SP, PR e MS, as tres UFs que exigem o hexadecimal
// do QR-Code em caixa baixa (ACBrNFe.pas:574 e :718).
var ufsComHexMinusculo = map[int]bool{35: true, 41: true, 50: true}

func ufDeHexMinusculo(cUF int) bool { return ufsComHexMinusculo[cUF] }

// TipoEmissaoDaChave devolve o digito de tipo de emissao gravado na chave de
// acesso -- a 35a posicao. Porte de ExtrairTipoEmissaoChaveAcesso
// (ACBrDFeUtil.pas:1152), sem o desvio do modelo 59 (SAT), que nao existe na
// NF-e.
//
// O QR-Code consulta a CHAVE, nao o ide.tpEmis, porque e a chave que o
// consumidor do documento tem em maos ao conferir.
func TipoEmissaoDaChave(chave string) int {
	chave = pcn.RemoverLiteralChave(chave)
	if len(chave) < 35 {
		return 0
	}
	d, err := strconv.Atoi(chave[34:35])
	if err != nil {
		return 0
	}
	return d
}

// cscOuDerivado devolve o CSC configurado ou, quando ele esta vazio, o
// codigo DERIVADO da propria chave que o ACBr monta (ACBrNFe.pas:680).
//
// Replicado apesar de estranho: um CSC inventado a partir da chave nao
// valida na SEFAZ. O ACBr faz isso para o DANFE de teste nao quebrar, e o
// comportamento e deterministico.
func cscOuDerivado(csc, chave, idCSC string) string {
	if csc != "" {
		return csc
	}
	// Copy(idNFe, 7, 8) + '20' + Copy(idNFe, 3, 2) + Copy(cIdCSC, 3, 4),
	// com os indices do Delphi (base 1) convertidos.
	return fatiar(chave, 6, 14) + "20" + fatiar(chave, 2, 4) + fatiar(idCSC, 2, 6)
}

// fatiar recorta [inicio, fim) sem estourar, como o Copy do Delphi.
func fatiar(s string, inicio, fim int) string {
	if inicio >= len(s) {
		return ""
	}
	if fim > len(s) {
		fim = len(s)
	}
	return s[inicio:fim]
}

// destinatarioNaQuery devolve o trecho "&cDest=..." da versao 1, ou vazio
// quando nao ha destinatario identificado.
func destinatarioNaQuery(n *NFe) string {
	if d := destinatarioDoQRCode(n); d != "" {
		return "&cDest=" + d
	}
	return ""
}

// destinatarioDoQRCode devolve o identificador do destinatario: o documento
// estrangeiro tem precedencia sobre o CNPJ/CPF (ACBrNFe.pas:662).
//
// O teste de vazio do idEstrangeiro e SEM trim, e so o resultado final e
// trimado -- e e isso que faz um idEstrangeiro composto so de espacos
// VENCER o CNPJ e produzir destinatario vazio. Replicado.
func destinatarioDoQRCode(n *NFe) string {
	if n.Dest.IDEstrangeiro != "" {
		return strings.TrimSpace(n.Dest.IDEstrangeiro)
	}
	return strings.TrimSpace(pcn.OnlyCPFCNPJAlphaNum(n.Dest.CNPJCPF))
}

// tipoDestinatario e o campo TipoDest da versao 3 (ACBrNFe.pas:760-769).
//
// ATENCAO a ordem: o ACBr marca 3 para estrangeiro e DEPOIS sobrescreve com
// 2 ou 1 se houver CNPJ/CPF com tamanho valido. Um documento que carregue os
// dois sai como CPF/CNPJ, nao como estrangeiro. Replicado.
func tipoDestinatario(n *NFe) string {
	tipo := ""
	if strings.TrimSpace(n.Dest.IDEstrangeiro) != "" {
		tipo = "3"
	}
	switch len(pcn.OnlyCPFCNPJAlphaNum(n.Dest.CNPJCPF)) {
	case 11:
		tipo = "2"
	case 14:
		tipo = "1"
	}
	return tipo
}

// acrescentarSufixoP garante o "?p=" que separa a URL do conteudo nas
// versoes 2 e 3. Algumas UFs (SC) ja trazem o sufixo na propria URL.
func acrescentarSufixoP(url string) string {
	if strings.Contains(url, "?p=") {
		return url
	}
	return url + "?p="
}

// asciiParaHex e o AsciiToHex do ACBr (ACBrUtil.Math.pas:524): cada byte em
// dois digitos hexadecimais MAIUSCULOS.
//
// A caixa importa -- o hash entra na URL e a SEFAZ o confere literalmente.
// E e justamente por ela que SP, PR e MS precisam do ToLower depois.
func asciiParaHex(s string) string {
	const digitos = "0123456789ABCDEF"
	var b strings.Builder
	b.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		b.WriteByte(digitos[s[i]>>4])
		b.WriteByte(digitos[s[i]&0x0F])
	}
	return b.String()
}

// sha1Bytes devolve o digest BINARIO, que e o que o SHA1 do Delphi entrega
// antes de passar pelo AsciiToHex.
func sha1Bytes(s string) []byte {
	h := sha1.Sum([]byte(s))
	return h[:]
}

// inteiroOuZero converte o IDCSC, que o ACBr guarda como string e converte
// com StrToIntDef(..., 0).
func inteiroOuZero(s string) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return v
}
