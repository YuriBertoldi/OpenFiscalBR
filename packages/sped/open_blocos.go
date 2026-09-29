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

// OpenBlocos is the base struct for all SPED blocks.
// IND_DAD indicates if the block has data (0) or not (1).
type OpenBlocos struct {
	IndMov int // 0=com dados, 1=sem dados (default 1)
}

// NewOpenBlocos creates a new OpenBlocos with IndMov defaulting to 1 (sem dados).
func NewOpenBlocos() *OpenBlocos {
	return &OpenBlocos{IndMov: 1}
}
