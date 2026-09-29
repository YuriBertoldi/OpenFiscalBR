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
	RegistroB350 []*RegistroB350
	RegistroB420 []*RegistroB420
	RegistroB440 []*RegistroB440
	RegistroB460 []*RegistroB460
	RegistroB470 []*RegistroB470
	RegistroB500 []*RegistroB500
}

// NewRegistroB001 cria um novo RegistroB001 com IndDad=1 (sem dados).
func NewRegistroB001() *RegistroB001 {
	return &RegistroB001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroB020 - Nota Fiscal (codigo 01), NF-e (codigo 55),
// NF Avulsa (codigo 1B) e NFC-e (codigo 65)
type RegistroB020 struct {
	IndOper       IndOper
	IndEmit       IndEmit
	CodPart       string
	CodMod        string
	CodSit        CodSit
	Ser           string
	NumDoc        string
	ChvNFe        string
	DtDoc         time.Time
	CodMunServ    string
	VlContIss     float64
	VlBcIss       float64
	VlIssRT       float64
	VlDed         float64
	VlIss         float64
	CodInfObs     string
	RegistroB025  []*RegistroB025
	RegistroB030  []*RegistroB030
	RegistroB035  []*RegistroB035
}

// RegistroB025 - Detalhamento por Combinacao de Aliquota e Item da Lista
// de Servicos da LC 116/2003
type RegistroB025 struct {
	VlContP   float64
	VlBcIssP  float64
	AliqIss   float64
	VlIssP    float64
	VlIssPRet float64
	CodServ   string
	CodInfObs string
}

// RegistroB030 - Nota Fiscal de Servicos Simplificada - Detalhamento por
// combinacao de aliquota e item (municipios que nao usam NFS-e)
type RegistroB030 struct {
	VlContP   float64
	VlBcIssP  float64
	AliqIss   float64
	VlIssP    float64
	VlIssPRet float64
	CodServ   string
	CodInfObs string
}

// RegistroB035 - Nota Fiscal de Servicos Simplificada - Detalhamento por
// Municipio
type RegistroB035 struct {
	VlContP   float64
	VlBcIssP  float64
	AliqIss   float64
	VlIssP    float64
	VlIssPRet float64
	CodServ   string
	CodInfObs string
}

// RegistroB350 - Servicos Prestados por Instituicoes Financeiras
type RegistroB350 struct {
	CodCtaISS string
	CodInfObs string
	VlCont    float64
	VlBcIss   float64
	AliqIss   float64
	VlIss     float64
	CodServ   string
}

// RegistroB420 - Totaliza Valores por Codigo de Tributacao ISS
type RegistroB420 struct {
	VlCont    float64
	VlBcIss   float64
	AliqIss   float64
	VlIssaPag float64
	VlDed     float64
	VlIss     float64
	CodServ   string
}

// RegistroB440 - Totaliza Valores por Municipio - ISS Retido pelo Tomador
type RegistroB440 struct {
	IndOper    IndOper
	CodMunServ string
	VlContRT  float64
	VlBcIssRT float64
	VlIssRT   float64
}

// RegistroB460 - Deducao do ISS
type RegistroB460 struct {
	IndDed   IndicadorDeducao
	VlDed    float64
	NumProc  string
	IndProc  OrigemProcesso
}

// RegistroB470 - Apuracao do ISS
type RegistroB470 struct {
	VlCont       float64
	VlMatTerc    float64
	VlMatProp    float64
	VlSub        float64
	VlIsntIss    float64
	VlDedBC      float64
	VlBcIss      float64
	VlBcIssRT    float64
	VlIssRT      float64
	VlDedIss     float64
	VlIss        float64
	RegistroB500 []*RegistroB500
}

// RegistroB500 - Apuracao do ISS - Obrigacoes Recolhidas ou a Recolher
type RegistroB500 struct {
	VlBcIss      float64
	AliqIss      float64
	VlIssaPag    float64
	VlDed        float64
	VlIss        float64
	IndObrISS    IndicadorObrigacao
	CodServ      string
	RegistroB510 []*RegistroB510
}

// RegistroB510 - Informacoes Complementares do ISS
type RegistroB510 struct {
	CodPart  string
	CodItem  string
	VlContP  float64
	VlBcIssP float64
	AliqIss  float64
	VlIssP   float64
}

// RegistroB990 - Encerramento do Bloco B
type RegistroB990 struct {
	QtdLinB int
}
