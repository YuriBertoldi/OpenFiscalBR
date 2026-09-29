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

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/comum"
)

// SPED is the base struct for all SPED components.
// Ported from TACBrSPED (ACBrSped.pas)
type SPED struct {
	comum.TXTClass
	DtIni   time.Time
	DtFin   time.Time
	Gravado bool
}

// NewSPED creates a new SPED instance with default settings.
func NewSPED() *SPED {
	s := &SPED{}
	s.Delimitador = "|"
	s.TrimString = true
	return s
}

// LimpaRegistros clears all registers. Override in sub-components.
func (s *SPED) LimpaRegistros() {
	s.Conteudo = nil
}
