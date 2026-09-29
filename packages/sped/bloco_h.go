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

package sped

import "time"

// ---------------------------------------------------------------------------
// Bloco H - Inventario Fisico
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroH001 - Abertura do Bloco H
type RegistroH001 struct {
	OpenBlocos
	RegistroH005 []*RegistroH005
}

// NewRegistroH001 cria um novo RegistroH001 com IndDad=1 (sem dados).
func NewRegistroH001() *RegistroH001 {
	return &RegistroH001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroH005 - Totais do Inventario
type RegistroH005 struct {
	DtInv        time.Time
	VlInv        float64
	MotInv       MotInv
	RegistroH010 []*RegistroH010
	RegistroH020 []*RegistroH020
}

// RegistroH010 - Inventario
type RegistroH010 struct {
	CodItem      string
	Unid         string
	Qtd          float64
	VlUnit       float64
	VlItem       float64
	IndProp      IndProp
	CodPart      string
	TxtCompl     string
	CodCta       string
	VlItemIR     float64
	RegistroH011 []*RegistroH011
	RegistroH020 []*RegistroH020
	RegistroH030 []*RegistroH030
}

// RegistroH011 - Proprietario do estoque, quando diferente do informante
type RegistroH011 struct {
	CNPJ string
}

// RegistroH020 - Informacao Complementar do Inventario
type RegistroH020 struct {
	CstICMS CstIcms
	BcICMS  float64
	VlICMS  float64
}

// RegistroH030 - Informacoes Complementares do Inventario das
// mercadorias sujeitas ao regime de Substituicao Tributaria
type RegistroH030 struct {
	VlICMSOp   float64
	VlBcICMSST float64
	VlICMSST   float64
	VlFCP      float64
}

// RegistroH990 - Encerramento do Bloco H
type RegistroH990 struct {
	QtdLinH int
}
