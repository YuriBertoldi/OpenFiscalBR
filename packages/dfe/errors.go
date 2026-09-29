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
	"errors"
	"fmt"
)

// Erros sentinela do package. Use errors.Is para testar.
var (
	// ErrCertificadoNaoCarregado indica operacao que exige certificado sem
	// um certificado configurado.
	ErrCertificadoNaoCarregado = errors.New("dfe: certificado digital nao carregado")
	// ErrCertificadoInvalido indica PFX/PEM que nao pode ser interpretado.
	ErrCertificadoInvalido = errors.New("dfe: certificado invalido")
	// ErrChaveNaoRSA indica certificado cuja chave privada nao e RSA --
	// a assinatura de DFe (RSA-SHA1) exige RSA, padrao ICP-Brasil.
	ErrChaveNaoRSA = errors.New("dfe: a chave privada do certificado nao e RSA")
	// ErrElementoNaoEncontrado indica que o elemento a assinar/verificar
	// nao existe no XML.
	ErrElementoNaoEncontrado = errors.New("dfe: elemento nao encontrado no XML")
	// ErrAssinaturaInvalida indica digest ou SignatureValue que nao confere.
	ErrAssinaturaInvalida = errors.New("dfe: assinatura invalida")
	// ErrConteudoMisto indica elemento com texto e filhos misturados, que a
	// canonicalizacao deste package nao suporta (nao ocorre em DFe).
	ErrConteudoMisto = errors.New("dfe: conteudo misto nao suportado na canonicalizacao")
	// ErrSOAPFault indica resposta soap:Fault do servidor.
	ErrSOAPFault = errors.New("dfe: o servidor devolveu um SOAP Fault")
	// ErrRespostaSemDados indica envelope de resposta sem o elemento de
	// resultado esperado.
	ErrRespostaSemDados = errors.New("dfe: resposta SOAP sem o elemento de resultado")
)

// ErroTransmissao carrega o contexto de uma falha de comunicacao com a
// SEFAZ: URL, acao e status HTTP.
type ErroTransmissao struct {
	URL        string
	SoapAction string
	StatusHTTP int
	Err        error
}

func (e *ErroTransmissao) Error() string {
	if e.StatusHTTP != 0 {
		return fmt.Sprintf("dfe: transmissao para %s falhou (HTTP %d): %v", e.URL, e.StatusHTTP, e.Err)
	}
	return fmt.Sprintf("dfe: transmissao para %s falhou: %v", e.URL, e.Err)
}

// Unwrap expoe o erro subjacente para errors.Is/errors.As.
func (e *ErroTransmissao) Unwrap() error { return e.Err }
