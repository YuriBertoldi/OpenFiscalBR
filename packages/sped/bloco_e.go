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
// Bloco E - Apuracao do ICMS e do IPI
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroE001 - Abertura do Bloco E
type RegistroE001 struct {
	OpenBlocos
	RegistroE100 []*RegistroE100
	RegistroE200 []*RegistroE200
	RegistroE300 []*RegistroE300
	RegistroE500 []*RegistroE500
}

// NewRegistroE001 cria um novo RegistroE001 com IndDad=1 (sem dados).
func NewRegistroE001() *RegistroE001 {
	return &RegistroE001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroE100 - Periodo da Apuracao do ICMS
type RegistroE100 struct {
	DtIni        time.Time
	DtFin        time.Time
	RegistroE110 []*RegistroE110
}

// RegistroE110 - Apuracao do ICMS - Operacoes Proprias
type RegistroE110 struct {
	VlTotDebitos      float64
	VlAjDebitos       float64
	VlTotAjDebitos    float64
	VlEstornosCred    float64
	VlTotCreditos     float64
	VlAjCreditos      float64
	VlTotAjCreditos   float64
	VlEstornosDeb     float64
	VlSldCredorAnt    float64
	VlSldApurado      float64
	VlTotDed          float64
	VlICMSRecolher    float64
	VlSldCredorTransp float64
	DebEsp            float64
	RegistroE111      []*RegistroE111
	RegistroE112      []*RegistroE112
	RegistroE113      []*RegistroE113
	RegistroE115      []*RegistroE115
	RegistroE116      []*RegistroE116
}

// RegistroE111 - Ajuste/Beneficio/Incentivo da Apuracao do ICMS
type RegistroE111 struct {
	CodAjApur    string
	DescrComplAj string
	VlAjApur     float64
}

// RegistroE112 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS
type RegistroE112 struct {
	NumDA    string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
}

// RegistroE113 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS -
// Identificacao dos Documentos Fiscais
type RegistroE113 struct {
	CodPart  string
	CodMod   string
	Ser      string
	Sub      string
	NumDoc   string
	DtDoc    time.Time
	CodItem  string
	VlAjItem float64
	ChvDOCe  string
}

// RegistroE115 - Informacoes Adicionais da Apuracao do ICMS -
// Valores Declaratorios
type RegistroE115 struct {
	CodInfAdic   string
	VlInfAdic    float64
	DescrComplAj string
}

// RegistroE116 - Obrigacoes do ICMS Recolhido ou a Recolher -
// Operacoes Proprias
type RegistroE116 struct {
	CodOR    string
	VlOR     float64
	DtVcto   time.Time
	CodRec   string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
	MesRef   string
}

// RegistroE200 - Periodo da Apuracao do ICMS - Substituicao Tributaria
type RegistroE200 struct {
	UF           string
	DtIni        time.Time
	DtFin        time.Time
	RegistroE210 []*RegistroE210
}

// RegistroE210 - Apuracao do ICMS - Substituicao Tributaria
type RegistroE210 struct {
	IndMovST          MovimentoST
	VlSldCredAntST    float64
	VlDevSTAnt        float64
	VlRessarcSTAnt    float64
	VlOutCredST       float64
	VlAjCreditosST    float64
	VlRetencaoST      float64
	VlOutDebST        float64
	VlAjDebitosST     float64
	VlSldDevAntST     float64
	VlDeducoesST      float64
	VlICMSRecST       float64
	VlSldCredSTTransp float64
	DebEspST          float64
	RegistroE220      []*RegistroE220
	RegistroE230      []*RegistroE230
	RegistroE240      []*RegistroE240
	RegistroE250      []*RegistroE250
}

// RegistroE220 - Ajuste/Beneficio/Incentivo da Apuracao do ICMS ST
type RegistroE220 struct {
	CodAjApur    string
	DescrComplAj string
	VlAjApur     float64
}

// RegistroE230 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS ST
type RegistroE230 struct {
	NumDA    string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
}

// RegistroE240 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS ST -
// Identificacao dos Documentos Fiscais
type RegistroE240 struct {
	CodPart  string
	CodMod   string
	Ser      string
	Sub      string
	NumDoc   string
	DtDoc    time.Time
	CodItem  string
	VlAjItem float64
	ChvDOCe  string
}

// RegistroE250 - Obrigacoes do ICMS Recolhido ou a Recolher - ST
type RegistroE250 struct {
	CodOR    string
	VlOR     float64
	DtVcto   time.Time
	CodRec   string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
	MesRef   string
}

// RegistroE300 - Periodo da Apuracao do ICMS Diferencial de Aliquota - UF
// Origem/Destino EC 87/15
type RegistroE300 struct {
	UF           string
	DtIni        time.Time
	DtFin        time.Time
	RegistroE310 []*RegistroE310
}

// RegistroE310 - Apuracao do ICMS Diferencial de Aliquota - UF Origem/Destino
// EC 87/15
type RegistroE310 struct {
	IndMovDIFAL          MovimentoDIFAL
	VlSldCredAntDIFAL    float64
	VlTotDebitosDIFAL    float64
	VlOutDebDIFAL        float64
	VlTotCreditosDIFAL   float64
	VlOutCredDIFAL       float64
	VlSldDevAntDIFAL     float64
	VlDeducoesDIFAL      float64
	VlRecolDIFAL         float64
	VlSldCredTranspDIFAL float64
	DebEspDIFAL          float64
	VlSldCredAntFCP      float64
	VlTotDebFCP          float64
	VlOutDebFCP          float64
	VlTotCredFCP         float64
	VlOutCredFCP         float64
	VlSldDevAntFCP       float64
	VlDeducoesFCP        float64
	VlRecolFCP           float64
	VlSldCredTranspFCP   float64
	DebEspFCP            float64
	RegistroE311         []*RegistroE311
	RegistroE312         []*RegistroE312
	RegistroE313         []*RegistroE313
	RegistroE316         []*RegistroE316
}

// RegistroE311 - Ajuste/Beneficio/Incentivo da Apuracao do ICMS Diferencial de
// Aliquota UF Origem/Destino EC 87/15
type RegistroE311 struct {
	CodAjApur    string
	DescrComplAj string
	VlAjApur     float64
}

// RegistroE312 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS
// Diferencial de Aliquota UF Origem/Destino EC 87/15
type RegistroE312 struct {
	NumDA    string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
}

// RegistroE313 - Informacoes Adicionais dos Ajustes da Apuracao do ICMS
// Diferencial de Aliquota - Identificacao dos Documentos Fiscais
type RegistroE313 struct {
	CodPart  string
	CodMod   string
	Ser      string
	Sub      string
	NumDoc   string
	DtDoc    time.Time
	CodItem  string
	VlAjItem float64
	ChvDOCe  string
}

// RegistroE316 - Obrigacoes do ICMS Recolhido ou a Recolher - Diferencial de
// Aliquota UF Origem/Destino EC 87/15
type RegistroE316 struct {
	CodOR    string
	VlOR     float64
	DtVcto   time.Time
	CodRec   string
	NumProc  string
	IndProc  OrigemProcesso
	Proc     string
	TxtCompl string
	MesRef   string
}

// RegistroE500 - Periodo de Apuracao do IPI
type RegistroE500 struct {
	IndApur      ApuracaoIPI
	DtIni        time.Time
	DtFin        time.Time
	RegistroE510 []*RegistroE510
	RegistroE520 []*RegistroE520
	RegistroE530 []*RegistroE530
}

// RegistroE510 - Consolidacao dos Valores do IPI
type RegistroE510 struct {
	CFOP      string
	CstIPI    CstIpi
	VlContIPI float64
	VlBcIPI   float64
	VlIPI     float64
}

// RegistroE520 - Apuracao do IPI
type RegistroE520 struct {
	VlSdAnteriorIPI float64
	VlDebitosIPI    float64
	VlCreditosIPI   float64
	VlOdIPI         float64
	VlOcIPI         float64
	VlScIPI         float64
	VlSdIPI         float64
}

// RegistroE530 - Ajustes da Apuracao do IPI
type RegistroE530 struct {
	IndAj        TipoAjuste
	VlAj         float64
	CodAj        string
	IndDoc       OrigemDocto
	NumDoc       string
	DescrAj      string
	RegistroE531 []*RegistroE531
}

// RegistroE531 - Informacoes Adicionais dos Ajustes da Apuracao do IPI -
// Identificacao dos Documentos Fiscais
type RegistroE531 struct {
	CodPart  string
	CodMod   string
	Ser      string
	Sub      string
	NumDoc   string
	DtDoc    time.Time
	CodItem  string
	VlAjItem float64
	ChvNFe   string
}

// RegistroE990 - Encerramento do Bloco E
type RegistroE990 struct {
	QtdLinE int
}
