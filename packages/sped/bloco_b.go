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
// Bloco B - Escrituracao e Apuracao do ISS
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroB001 - Abertura do Bloco B
type RegistroB001 struct {
	OpenBlocos
	RegistroB020 []*RegistroB020
	RegistroB030 []*RegistroB030
	RegistroB350 []*RegistroB350
	RegistroB420 []*RegistroB420
	RegistroB440 []*RegistroB440
	RegistroB460 []*RegistroB460
	RegistroB470 []*RegistroB470
	RegistroB500 []*RegistroB500
}

// NewRegistroB001 cria um novo RegistroB001 com IndMov=1 (sem dados).
func NewRegistroB001() *RegistroB001 {
	return &RegistroB001{OpenBlocos: OpenBlocos{IndMov: 1}}
}

// RegistroB020 - Nota Fiscal (codigo 01), NF-e (codigo 55),
// NF Avulsa (codigo 1B) e NFC-e (codigo 65)
type RegistroB020 struct {
	IndOper      IndOper
	IndEmit      IndEmit
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	NumDoc       string
	ChvNFe       string
	DtDoc        time.Time
	CodMunServ   string
	VlCont       float64
	VlMatTerc    float64
	VlSub        float64
	VlIsntIss    float64
	VlDedBC      float64
	VlBcIss      float64
	VlBcIssRT    float64
	VlIssRT      float64
	VlIss        float64
	CodInfObs    string
	RegistroB025 []*RegistroB025
}

// RegistroB025 - Detalhamento por Combinacao de Aliquota e Item da Lista
// de Servicos da LC 116/2003
type RegistroB025 struct {
	VlContP    float64
	VlBcIssP   float64
	AliqIss    float64
	VlIssP     float64
	VlIsntIssP float64
	CodServ    string
}

// RegistroB030 - Nota Fiscal de Servicos Simplificada (codigo 3B)
type RegistroB030 struct {
	CodMod       string
	Ser          string
	NumDocIni    string
	NumDocFin    string
	DtDoc        time.Time
	QtdCanc      int
	VlCont       float64
	VlIsntIss    float64
	VlBcIss      float64
	VlIss        float64
	CodInfObs    string
	RegistroB035 []*RegistroB035
}

// RegistroB035 - Detalhamento por Combinacao de Aliquota e Item da Lista
// de Servicos, da Nota Fiscal de Servicos Simplificada
type RegistroB035 struct {
	VlContP    float64
	VlBcIssP   float64
	AliqIss    float64
	VlIssP     float64
	VlIsntIssP float64
	CodServ    string
}

// RegistroB350 - Servicos Prestados por Instituicoes Financeiras
type RegistroB350 struct {
	CodCtd    string
	CtaIss    string
	CtaCosif  string
	QtdOcor   int
	CodServ   string
	VlCont    float64
	VlBcIss   float64
	AliqIss   float64
	VlIss     float64
	CodInfObs string
}

// RegistroB420 - Totalizacao dos Valores dos Servicos Prestados
type RegistroB420 struct {
	VlCont    float64
	VlBcIss   float64
	AliqIss   float64
	VlIsntIss float64
	VlIss     float64
	CodServ   string
}

// RegistroB440 - Totalizacao dos Valores Retidos
type RegistroB440 struct {
	IndOper   IndOper
	CodPart   string
	VlContRT  float64
	VlBcIssRT float64
	VlIssRT   float64
}

// RegistroB460 - Deducao, Incentivo e Beneficio Fiscal
type RegistroB460 struct {
	IndDed    IndicadorDeducao
	VlDed     float64
	NumProc   string
	IndProc   OrigemProcesso
	Proc      string
	CodInfObs string
	IndObr    IndicadorObrigacao
}

// RegistroB470 - Apuracao do ISS
type RegistroB470 struct {
	VlCont      float64
	VlMatTerc   float64
	VlMatProp   float64
	VlSub       float64
	VlIsnt      float64
	VlDedBC     float64
	VlBcIss     float64
	VlBcIssRT   float64
	VlIss       float64
	VlIssRT     float64
	VlDed       float64
	VlIssRec    float64
	VlIssST     float64
	VlIssRecUni float64
}

// RegistroB500 - Informacoes Complementares da Apuracao - Sociedade
// Uniprofissional
type RegistroB500 struct {
	VlRec        float64
	QtdProf      int
	VlOR         float64
	RegistroB510 []*RegistroB510
}

// RegistroB510 - Profissionais Habilitados - Sociedade Uniprofissional
type RegistroB510 struct {
	IndProf string
	IndEsc  string
	IndSoc  string
	CPF     string
	Nome    string
}

// RegistroB990 - Encerramento do Bloco B
type RegistroB990 struct {
	QtdLinB int
}
