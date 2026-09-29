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
// Bloco G - Controle do Credito de ICMS do Ativo Permanente (CIAP)
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroG001 - Abertura do Bloco G
type RegistroG001 struct {
	OpenBlocos
	RegistroG110 []*RegistroG110
}

// NewRegistroG001 cria um novo RegistroG001 com IndDad=1 (sem dados).
func NewRegistroG001() *RegistroG001 {
	return &RegistroG001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroG110 - ICMS - Ativo Permanente - CIAP
type RegistroG110 struct {
	DtIni        time.Time
	DtFin        time.Time
	SldInICMS    float64
	SomParcl     float64
	VlTribExp    float64
	VlTotal      float64
	IndPerSai    int
	ICMSAprop    float64
	SomICMSOC    float64
	RegistroG125 []*RegistroG125
}

// RegistroG125 - Movimentacao de Bem ou Componente do Ativo Imobilizado
type RegistroG125 struct {
	CodItem      string
	DtMoviment   time.Time
	TipoMov      string
	VlImobICMS   float64
	VlICMSOp     float64
	VlICMSST     float64
	VlFCP        float64
	NumParcl     int
	VlParcl      float64
	RegistroG126 []*RegistroG126
	RegistroG130 []*RegistroG130
	RegistroG140 []*RegistroG140
}

// RegistroG126 - Outros Creditos CIAP
type RegistroG126 struct {
	DtIni    time.Time
	DtFin    time.Time
	NumParcl int
	VlParcl  float64
	VlTribOC float64
	VlTotOC  float64
	IndPerOC int
	VlICMSOC float64
}

// RegistroG130 - Identificacao do Documento Fiscal
type RegistroG130 struct {
	IndEmit IndEmit
	CodPart string
	CodMod  string
	Ser     string
	NumDoc  string
	ChvNFe  string
	DtDoc   time.Time
	NumDA   string
}

// RegistroG140 - Identificacao do Item do Documento Fiscal
type RegistroG140 struct {
	NumItem   string
	CodItem   string
	Qtd       float64
	Unid      string
	VlICMSOPr float64
	VlBcICMSST float64
	VlICMSST  float64
	VlFCP     float64
}

// RegistroG990 - Encerramento do Bloco G
type RegistroG990 struct {
	QtdLinG int
}
