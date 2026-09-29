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

package sped

import "errors"

var (
	ErrDtIniNaoInformada = errors.New("data inicial (DT_INI) nao informada")
	ErrDtFinNaoInformada = errors.New("data final (DT_FIN) nao informada")
	ErrArquivoNaoGerado  = errors.New("arquivo SPED nao foi gerado")
)

// SPEDFiscalError represents SPED Fiscal specific errors.
type SPEDFiscalError struct {
	Message string
}

func (e *SPEDFiscalError) Error() string { return "SPED Fiscal: " + e.Message }
