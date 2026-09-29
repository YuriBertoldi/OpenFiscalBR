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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"os"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"
)

// Certificado e um certificado digital A1 carregado em memoria, com a chave
// privada RSA e a cadeia. Porte do papel de TDFeSSL quanto a certificado
// (CarregarCertificado, CertCNPJ, CertDataVenc, CertNumeroSerie...).
//
// DIVERGENCIA do CLAUDE.md (que indica golang.org/x/crypto/pkcs12): aquela
// biblioteca esta congelada e nao decodifica PFX modernos da ICP-Brasil
// (PBES2/AES-256); usamos software.sslmate.com/src/go-pkcs12, que decodifica.
type Certificado struct {
	// Cert e o certificado do titular.
	Cert *x509.Certificate
	// Chave e a chave privada RSA correspondente.
	Chave *rsa.PrivateKey
	// Cadeia sao os certificados intermediarios/raiz presentes no PFX,
	// quando houver.
	Cadeia []*x509.Certificate
}

// CarregarPFX carrega um certificado A1 no formato PFX/PKCS#12.
func CarregarPFX(dados []byte, senha string) (*Certificado, error) {
	chave, cert, cadeia, err := pkcs12.DecodeChain(dados, senha)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCertificadoInvalido, err)
	}
	rsaChave, ok := chave.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w (tipo %T)", ErrChaveNaoRSA, chave)
	}
	return &Certificado{Cert: cert, Chave: rsaChave, Cadeia: cadeia}, nil
}

// CarregarPFXArquivo carrega um PFX a partir do caminho em disco -- o
// equivalente do ArquivoPFX + Senha da configuracao do ACBr.
func CarregarPFXArquivo(caminho, senha string) (*Certificado, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("dfe: ler PFX %s: %w", caminho, err)
	}
	return CarregarPFX(dados, senha)
}

// CarregarPEM carrega certificado e chave privada em PEM (par cert/key),
// como os gerados por openssl. Aceita chave PKCS#1 e PKCS#8.
func CarregarPEM(certPEM, chavePEM []byte) (*Certificado, error) {
	par, err := tls.X509KeyPair(certPEM, chavePEM)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCertificadoInvalido, err)
	}
	cert, err := x509.ParseCertificate(par.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrCertificadoInvalido, err)
	}
	rsaChave, ok := par.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("%w (tipo %T)", ErrChaveNaoRSA, par.PrivateKey)
	}
	var cadeia []*x509.Certificate
	for _, der := range par.Certificate[1:] {
		c, err := x509.ParseCertificate(der)
		if err == nil {
			cadeia = append(cadeia, c)
		}
	}
	return &Certificado{Cert: cert, Chave: rsaChave, Cadeia: cadeia}, nil
}

// TLSCertificate devolve o par no formato exigido pelo tls.Config
// (Certificates) para o TLS mutuo com a SEFAZ.
func (c *Certificado) TLSCertificate() tls.Certificate {
	cert := tls.Certificate{PrivateKey: c.Chave}
	if c.Cert != nil {
		cert.Certificate = append(cert.Certificate, c.Cert.Raw)
	}
	for _, ca := range c.Cadeia {
		cert.Certificate = append(cert.Certificate, ca.Raw)
	}
	return cert
}

// Vencimento devolve a data de expiracao (CertDataVenc do ACBr).
func (c *Certificado) Vencimento() time.Time {
	if c == nil || c.Cert == nil {
		return time.Time{}
	}
	return c.Cert.NotAfter
}

// Vencido informa se o certificado ja expirou.
func (c *Certificado) Vencido() bool {
	return c != nil && c.Cert != nil && time.Now().After(c.Cert.NotAfter)
}

// RazaoSocial devolve o CN do titular (CertRazaoSocial do ACBr).
func (c *Certificado) RazaoSocial() string {
	if c == nil || c.Cert == nil {
		return ""
	}
	return c.Cert.Subject.CommonName
}

// NumeroSerie devolve o numero de serie em hexadecimal (CertNumeroSerie).
func (c *Certificado) NumeroSerie() string {
	if c == nil || c.Cert == nil {
		return ""
	}
	return fmt.Sprintf("%X", c.Cert.SerialNumber)
}

// CNPJ devolve o CNPJ do titular extraido da extensao ICP-Brasil
// (otherName OID 2.16.76.1.3.3 no subjectAltName), como o CertCNPJ do ACBr.
// Devolve vazio quando a extensao nao existe (ex.: e-CPF).
func (c *Certificado) CNPJ() string {
	if c == nil || c.Cert == nil {
		return ""
	}
	return extrairOtherNameICP(c.Cert, asn1.ObjectIdentifier{2, 16, 76, 1, 3, 3})
}

var oidSubjectAltName = asn1.ObjectIdentifier{2, 5, 29, 17}

// extrairOtherNameICP varre o subjectAltName procurando um otherName com o
// OID informado e devolve o valor textual (padrao das extensoes ICP-Brasil).
func extrairOtherNameICP(cert *x509.Certificate, oid asn1.ObjectIdentifier) string {
	for _, ext := range cert.Extensions {
		if !ext.Id.Equal(oidSubjectAltName) {
			continue
		}
		var seq asn1.RawValue
		if _, err := asn1.Unmarshal(ext.Value, &seq); err != nil {
			return ""
		}
		resto := seq.Bytes
		for len(resto) > 0 {
			var gn asn1.RawValue
			var err error
			resto, err = asn1.Unmarshal(resto, &gn)
			if err != nil {
				return ""
			}
			// otherName e o GeneralName de tag [0]
			if gn.Tag != 0 || gn.Class != asn1.ClassContextSpecific {
				continue
			}
			var onOID asn1.ObjectIdentifier
			corpo, err := asn1.Unmarshal(gn.Bytes, &onOID)
			if err != nil || !onOID.Equal(oid) {
				continue
			}
			// value e [0] EXPLICIT, contendo uma string ASN.1
			var valor asn1.RawValue
			if _, err := asn1.Unmarshal(corpo, &valor); err != nil {
				continue
			}
			var s asn1.RawValue
			if _, err := asn1.Unmarshal(valor.Bytes, &s); err != nil {
				// alguns emissores gravam a string direto, sem wrapper
				return string(valor.Bytes)
			}
			return string(s.Bytes)
		}
	}
	return ""
}
