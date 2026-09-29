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

// WriteRegistroFunc is called before/after writing a register line.
// Allows modification of the line content.
type WriteRegistroFunc func(linha *string)

// CheckRegistroFunc is called to validate a register before writing.
// Set abortar to true to skip the register.
type CheckRegistroFunc func(registro interface{}, abortar *bool)

// ErrorFunc is called when an error occurs during processing.
type ErrorFunc func(msg string)
