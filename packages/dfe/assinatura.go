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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Assinatura XMLDSig enveloped dos DFe: RSA-SHA1 com Canonical XML 1.0,
// referencia por Id ("#NFGas<chave>"). Porte do papel de TDFeSSL.Assinar +
// do template TSignature.GerarXML (pcnSignature). SHA1 e uma exigencia dos
// leiautes da SEFAZ, nao uma escolha deste port.

// Algoritmos e namespaces do XMLDSig usados pelos DFe.
const (
	nsXMLDSig    = "http://www.w3.org/2000/09/xmldsig#"
	algC14N      = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"
	algRSASHA1   = "http://www.w3.org/2000/09/xmldsig#rsa-sha1"
	algEnveloped = "http://www.w3.org/2000/09/xmldsig#enveloped-signature"
	algSHA1      = "http://www.w3.org/2000/09/xmldsig#sha1"
)

// AssinarXML assina o documento, referenciando o elemento infElement (ex.:
// "infNFGas", "infEvento") pelo atributo Id, e insere o bloco <Signature>
// como ultimo filho do elemento raiz -- exatamente onde o ACBr insere.
// Assinatura preexistente e descartada e refeita, como no AssinarXML do
// TDFeSSL. Devolve o XML assinado.
func AssinarXML(cert *Certificado, xmlStr, infElement string) (string, error) {
	if cert == nil || cert.Chave == nil {
		return "", ErrCertificadoNaoCarregado
	}

	doc, err := pcn.ParseString(xmlStr)
	if err != nil {
		return "", fmt.Errorf("dfe: assinar: %w", err)
	}

	// remove assinatura anterior, se houver
	if sig := buscarProfundo(doc.Root, "Signature", nsXMLDSig); sig != nil {
		xmlStr = strings.Replace(xmlStr, sig.OuterXML(), "", 1)
		doc, err = pcn.ParseString(xmlStr)
		if err != nil {
			return "", fmt.Errorf("dfe: assinar: %w", err)
		}
	}

	inf := doc.Root
	if inf.Nome != infElement {
		inf = buscarProfundo(doc.Root, infElement, "")
	}
	if inf == nil {
		return "", fmt.Errorf("%w: <%s>", ErrElementoNaoEncontrado, infElement)
	}
	id := inf.Attr("Id")
	if id == "" {
		return "", fmt.Errorf("dfe: assinar: <%s> sem atributo Id", infElement)
	}

	canonico, err := CanonicalizarFragmento(inf.OuterXML(), map[string]string{"": inf.Space}, true)
	if err != nil {
		return "", err
	}
	digest := sha1.Sum([]byte(canonico))
	digestB64 := base64.StdEncoding.EncodeToString(digest[:])

	signedInfo := montarSignedInfo(id, digestB64, true)
	hashSignedInfo := sha1.Sum([]byte(signedInfo))
	assinatura, err := rsa.SignPKCS1v15(rand.Reader, cert.Chave, crypto.SHA1, hashSignedInfo[:])
	if err != nil {
		return "", fmt.Errorf("dfe: assinar: %w", err)
	}

	bloco := "<Signature xmlns=\"" + nsXMLDSig + "\">" +
		montarSignedInfo(id, digestB64, false) +
		"<SignatureValue>" + base64.StdEncoding.EncodeToString(assinatura) + "</SignatureValue>" +
		"<KeyInfo><X509Data><X509Certificate>" +
		base64.StdEncoding.EncodeToString(cert.Cert.Raw) +
		"</X509Certificate></X509Data></KeyInfo></Signature>"

	// insere como ULTIMO FILHO do docElement -- o PAI do elemento
	// referenciado (NFGas, eventoNFGas...). Trabalhar relativo ao pai, e
	// nao ao fim do documento, evita que assinar um XML ja envelopado
	// (nfgasProc) deixe a Signature fora do documento.
	docElem := inf.Pai()
	if docElem == nil {
		docElem = inf
	}
	outer := docElem.OuterXML()
	pos := strings.LastIndex(outer, "</")
	if pos < 0 {
		return "", fmt.Errorf("dfe: assinar: XML sem elemento de fechamento")
	}
	return strings.Replace(xmlStr, outer, outer[:pos]+bloco+outer[pos:], 1), nil
}

// montarSignedInfo monta o SignedInfo. comXmlns=true produz a forma CANONICA
// (usada para calcular a assinatura, com o xmlns herdado do Signature
// renderizado); false produz a forma inserida no documento, onde o xmlns ja
// esta declarado no <Signature> pai. Fora o xmlns, as duas formas sao
// byte a byte identicas -- e por construcao ja canonicas (sem self-closing,
// atributos unicos), entao a SEFAZ recalcula o mesmo hash.
func montarSignedInfo(id, digestB64 string, comXmlns bool) string {
	xmlns := ""
	if comXmlns {
		xmlns = " xmlns=\"" + nsXMLDSig + "\""
	}
	return "<SignedInfo" + xmlns + ">" +
		"<CanonicalizationMethod Algorithm=\"" + algC14N + "\"></CanonicalizationMethod>" +
		"<SignatureMethod Algorithm=\"" + algRSASHA1 + "\"></SignatureMethod>" +
		"<Reference URI=\"#" + id + "\">" +
		"<Transforms>" +
		"<Transform Algorithm=\"" + algEnveloped + "\"></Transform>" +
		"<Transform Algorithm=\"" + algC14N + "\"></Transform>" +
		"</Transforms>" +
		"<DigestMethod Algorithm=\"" + algSHA1 + "\"></DigestMethod>" +
		"<DigestValue>" + digestB64 + "</DigestValue>" +
		"</Reference></SignedInfo>"
}

// VerificarAssinatura confere o digest do elemento referenciado e a
// assinatura RSA do SignedInfo contra o certificado embutido no KeyInfo.
// Porte de TDFeSSL.VerificarAssinatura. Nao valida a cadeia ICP-Brasil --
// isso exige o repositorio de ACs e fica a cargo do consumidor.
func VerificarAssinatura(xmlStr string) error {
	doc, err := pcn.ParseString(xmlStr)
	if err != nil {
		return fmt.Errorf("dfe: verificar assinatura: %w", err)
	}
	sig := buscarProfundo(doc.Root, "Signature", nsXMLDSig)
	if sig == nil {
		return fmt.Errorf("%w: <Signature>", ErrElementoNaoEncontrado)
	}
	signedInfo := sig.FindAnyNs("SignedInfo")
	if signedInfo == nil {
		return fmt.Errorf("%w: <SignedInfo>", ErrElementoNaoEncontrado)
	}
	ref := signedInfo.FindAnyNs("Reference")
	uri := ref.Attr("URI")
	if !strings.HasPrefix(uri, "#") {
		return fmt.Errorf("%w: Reference URI %q nao suportada (esperado \"#Id\")", ErrAssinaturaInvalida, uri)
	}
	id := uri[1:]

	alvo := buscarPorId(doc.Root, id)
	if alvo == nil {
		return fmt.Errorf("%w: elemento com Id=%q", ErrElementoNaoEncontrado, id)
	}

	// 1) digest do elemento referenciado
	canonico, err := CanonicalizarFragmento(alvo.OuterXML(), map[string]string{"": alvo.Space}, true)
	if err != nil {
		return err
	}
	digest := sha1.Sum([]byte(canonico))
	digestDoc := strings.TrimSpace(pcn.ConteudoStr(ref.FindAnyNs("DigestValue")))
	if base64.StdEncoding.EncodeToString(digest[:]) != digestDoc {
		return fmt.Errorf("%w: DigestValue nao confere para Id=%q", ErrAssinaturaInvalida, id)
	}

	// 2) assinatura RSA do SignedInfo canonico
	herdado := map[string]string{"": nsXMLDSig}
	if qn := sig.NomeQualificado; strings.Contains(qn, ":") {
		herdado = map[string]string{qn[:strings.Index(qn, ":")]: nsXMLDSig}
	}
	siCanonico, err := CanonicalizarFragmento(signedInfo.OuterXML(), herdado, false)
	if err != nil {
		return err
	}
	hashSI := sha1.Sum([]byte(siCanonico))

	certB64 := strings.TrimSpace(pcn.ConteudoStr(sig.FindAnyNs("KeyInfo").FindAnyNs("X509Data").FindAnyNs("X509Certificate")))
	if certB64 == "" {
		return fmt.Errorf("%w: <X509Certificate>", ErrElementoNaoEncontrado)
	}
	certDER, err := base64.StdEncoding.DecodeString(retirarEspacosB64(certB64))
	if err != nil {
		return fmt.Errorf("%w: X509Certificate ilegivel: %s", ErrAssinaturaInvalida, err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCertificadoInvalido, err)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return ErrChaveNaoRSA
	}

	sigB64 := strings.TrimSpace(pcn.ConteudoStr(sig.FindAnyNs("SignatureValue")))
	sigBytes, err := base64.StdEncoding.DecodeString(retirarEspacosB64(sigB64))
	if err != nil {
		return fmt.Errorf("%w: SignatureValue ilegivel: %s", ErrAssinaturaInvalida, err)
	}
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA1, hashSI[:], sigBytes); err != nil {
		return fmt.Errorf("%w: SignatureValue nao confere", ErrAssinaturaInvalida)
	}
	return nil
}

// HashCSRT calcula o hash do CSRT para o QR-Code / infRespTec:
// base64(sha1(CSRT + chave de acesso)). Porte de CalcularHashCSRT
// (ACBrDFeUtil.pas).
func HashCSRT(csrt, chave string) string {
	h := sha1.Sum([]byte(csrt + chave))
	return base64.StdEncoding.EncodeToString(h[:])
}

// AssinarSHA1Base64 assina o SHA-1 de dados com a chave RSA do certificado
// e devolve em base64 -- o CalcHash(dgstSHA1, outBase64, Assinar=True) do
// TDFeSSL, usado no parametro sign do QR-Code em emissao offline.
func (c *Certificado) AssinarSHA1Base64(dados string) (string, error) {
	if c == nil || c.Chave == nil {
		return "", ErrCertificadoNaoCarregado
	}
	h := sha1.Sum([]byte(dados))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.Chave, crypto.SHA1, h[:])
	if err != nil {
		return "", fmt.Errorf("dfe: assinar hash: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// buscarProfundo procura, em profundidade, o primeiro elemento com o local
// name informado (e, se space != "", tambem o namespace).
func buscarProfundo(n *pcn.Node, local, space string) *pcn.Node {
	if n == nil {
		return nil
	}
	if n.Nome == local && (space == "" || n.Space == space) {
		return n
	}
	for _, f := range n.Filhos {
		if achado := buscarProfundo(f, local, space); achado != nil {
			return achado
		}
	}
	return nil
}

// buscarPorId procura, em profundidade, o elemento com atributo Id igual.
func buscarPorId(n *pcn.Node, id string) *pcn.Node {
	if n == nil {
		return nil
	}
	if n.Attr("Id") == id {
		return n
	}
	for _, f := range n.Filhos {
		if achado := buscarPorId(f, id); achado != nil {
			return achado
		}
	}
	return nil
}

// retirarEspacosB64 remove quebras e espacos de um base64 que pode vir
// formatado em multiplas linhas.
func retirarEspacosB64(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n':
			return -1
		}
		return r
	}, s)
}
