// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-28.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package comum

import "errors"

var (
	// ErrArquivoNaoEspecificado is returned when NomeArquivo is empty.
	ErrArquivoNaoEspecificado = errors.New("NomeArquivo nao especificado")

	// ErrValorNaoNumerico is returned when a non-numeric value is passed
	// where a number is expected.
	ErrValorNaoNumerico = errors.New("parametro Value nao possui valor numerico")
)
