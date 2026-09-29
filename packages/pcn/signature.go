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

package pcn

// Signature carrega os dados da assinatura digital XMLDSig do documento.
// Porte de TSignature (ACBrXmlBase.pas).
type Signature struct {
	URI              string // atributo URI de SignedInfo/Reference
	DigestValue      string // SignedInfo/Reference/DigestValue
	SignatureValue   string // Signature/SignatureValue
	X509Certificate  string // KeyInfo/X509Data/X509Certificate
	IDSignature      string // preenchido apenas na geracao
	IDSignatureValue string // preenchido apenas na geracao
}

// Assinada informa se ha uma assinatura COMPLETA (DigestValue,
// SignatureValue e X509Certificate presentes) -- a condicao que o
// taSomenteSeAssinada do ACBr exige para reembutir o bloco Signature na
// geracao (ACBrNFGas.XmlWriter.pas:268-280). Um struct so com URI, por
// exemplo, esta "nao vazio" mas nao esta assinado.
func (s *Signature) Assinada() bool {
	return s != nil && s.DigestValue != "" && s.SignatureValue != "" && s.X509Certificate != ""
}

// Vazia informa se nenhum dado de assinatura foi lido.
func (s *Signature) Vazia() bool {
	if s == nil {
		return true
	}
	return s.URI == "" && s.DigestValue == "" && s.SignatureValue == "" && s.X509Certificate == ""
}

// Limpar zera todos os campos.
func (s *Signature) Limpar() {
	if s == nil {
		return
	}
	*s = Signature{}
}

// LerSignature preenche a assinatura a partir do elemento <Signature>.
// Porte de LerSignature (ACBrXmlBase.pas).
//
// Node nil nao e erro: documento sem assinatura simplesmente deixa a
// estrutura zerada, como no original.
func LerSignature(node *Node, sig *Signature) {
	if node == nil || sig == nil {
		return
	}

	reference := node.FindAnyNs("SignedInfo").FindAnyNs("Reference")
	x509Data := node.FindAnyNs("KeyInfo").FindAnyNs("X509Data")

	sig.URI = reference.Attr("URI")
	sig.DigestValue = ConteudoStr(reference.FindAnyNs("DigestValue"))
	sig.SignatureValue = ConteudoStr(node.FindAnyNs("SignatureValue"))
	sig.X509Certificate = ConteudoStr(x509Data.FindAnyNs("X509Certificate"))
}
