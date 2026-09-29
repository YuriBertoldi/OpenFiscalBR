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

// Package comum provides base types and utilities shared across all
// OpenFiscalBR components. It is a Go port of the core ACBr units
// ACBrBase.pas, ACBrTXTClass.pas, and related utility code.
//
// Key types:
//
//   - Component: base component (TACBrComponent)
//   - TXTClass:  text file generator used by SPED (TACBrTXTClass)
//   - ACBrError: standard error type (EACBrException)
//
// Key functions:
//
//   - FormatFloatBR: formats a float for SPED output
//   - FloatMask:     generates a format mask string
//   - ACBrStr:       identity function (no encoding conversion in Go)
package comum
