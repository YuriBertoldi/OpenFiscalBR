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
// Bloco D - Documentos Fiscais II - Servicos (ICMS)
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroD001 - Abertura do Bloco D
type RegistroD001 struct {
	OpenBlocos
	RegistroD100 []*RegistroD100
	RegistroD300 []*RegistroD300
	RegistroD350 []*RegistroD350
	RegistroD400 []*RegistroD400
	RegistroD500 []*RegistroD500
	RegistroD600 []*RegistroD600
	RegistroD695 []*RegistroD695
	RegistroD700 []*RegistroD700
	RegistroD750 []*RegistroD750
}

// NewRegistroD001 cria um novo RegistroD001 com IndMov=1 (sem dados).
func NewRegistroD001() *RegistroD001 {
	return &RegistroD001{OpenBlocos: OpenBlocos{IndMov: 1}}
}

// RegistroD100 - Nota Fiscal de Servico de Transporte (codigo 07),
// Conhecimentos de Transporte Rodoviario de Cargas (codigo 08),
// Conhecimentos de Transporte de Cargas Avulso (codigo 8B),
// Aquaviario de Cargas (codigo 09), Aereo (codigo 10),
// Ferroviario de Cargas (codigo 11), Multimodal (codigo 26),
// Nota Fiscal de Transporte Ferroviario de Carga (codigo 27),
// CT-e (codigo 57) e CT-e OS (codigo 67)
type RegistroD100 struct {
	IndOper      IndOper
	IndEmit      IndEmit
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	Sub          string
	NumDoc       string
	ChvCTe       string
	DtDoc        time.Time
	DtAP         time.Time
	TpCTe        string
	ChvCTeRef    string
	VlDoc        float64
	VlDesc       float64
	IndFrt       IndFrt
	VlServ       float64
	VlBcICMS     float64
	VlICMS       float64
	VlNT         float64
	CodInf       string
	CodCta       string
	CodMunOrig   string
	CodMunDest   string
	RegistroD101 []*RegistroD101
	RegistroD110 []*RegistroD110
	RegistroD130 []*RegistroD130
	RegistroD140 []*RegistroD140
	RegistroD150 []*RegistroD150
	RegistroD160 []*RegistroD160
	RegistroD170 []*RegistroD170
	RegistroD180 []*RegistroD180
	RegistroD190 []*RegistroD190
	RegistroD195 []*RegistroD195
}

// RegistroD101 - Informacao Complementar dos Documentos Fiscais quando das
// operacoes interestaduais destinadas a consumidor final nao contribuinte (EC 87/15)
type RegistroD101 struct {
	VlFcpUFDest  float64
	VlICMSUFDest float64
	VlICMSUFRem  float64
}

// RegistroD110 - Itens do Documento - Nota Fiscal de Servico de Transporte
type RegistroD110 struct {
	NunItem      int
	CodItem      string
	VlServ       float64
	VlOut        float64
	RegistroD120 []*RegistroD120
}

// RegistroD120 - Complemento da Nota Fiscal de Servico de Transporte
type RegistroD120 struct {
	CodMunOrig string
	CodMunDest string
	VeicID     string
	UFID       string
}

// RegistroD130 - Complemento do Conhecimento Rodoviario de Cargas
type RegistroD130 struct {
	CodPartConsg string
	CodPartRed   string
	IndFrtRed    TipoFreteRedespacho
	CodMunOrig   string
	CodMunDest   string
	VeicID       string
	VlLiqFrt     float64
	VlSecCat     float64
	VlDesp       float64
	VlPedg       float64
	VlOut        float64
	VlFrt        float64
	UFID         string
}

// RegistroD140 - Complemento do Conhecimento Aquaviario de Cargas
type RegistroD140 struct {
	CodPartConsg  string
	CodMunOrig    string
	CodMunDest    string
	IndVeic       TipoVeiculo
	VeicID        string
	IndNav        TipoNavegacao
	Viagem        string
	VlFrtLiq      float64
	VlDespPort    float64
	VlDespCarDesc float64
	VlOut         float64
	VlFrtBrt      float64
	VlFrtMM       float64
}

// RegistroD150 - Complemento do Conhecimento Aereo
type RegistroD150 struct {
	CodMunOrig string
	CodMunDest string
	VeicID     string
	Viagem     string
	IndTFA     TipoTarifa
	VlPesoTx   float64
	VlTxTerr   float64
	VlTxRed    float64
	VlOut      float64
	VlTxAdv    float64
}

// RegistroD160 - Carga Transportada
type RegistroD160 struct {
	Despacho     string
	CNPJCPFRem   string
	IERem        string
	CodMunOri    string
	CNPJCPFDest  string
	IEDest       string
	CodMunDest   string
	RegistroD161 []*RegistroD161
	RegistroD162 []*RegistroD162
}

// RegistroD161 - Local de Coleta e Entrega
type RegistroD161 struct {
	IndCarga   TipoTransporte
	CNPJCol    string
	IECol      string
	CodMunCol  string
	CNPJEntg   string
	IEEntg     string
	CodMunEntg string
}

// RegistroD162 - Identificacao dos Documentos Fiscais
type RegistroD162 struct {
	CodMod  string
	Ser     string
	NumDoc  string
	DtDoc   time.Time
	VlDoc   float64
	VlMerc  float64
	QtdVol  int
	PesoBrt float64
	PesoLiq float64
}

// RegistroD170 - Complemento do Conhecimento Multimodal de Cargas
type RegistroD170 struct {
	CodPartConsg string
	CodPartRed   string
	CodMunOrig   string
	CodMunDest   string
	Otm          string
	IndNatFrt    NaturezaFrete
	VlLiqFrt     float64
	VlGris       float64
	VlPdg        float64
	VlOut        float64
	VlFrt        float64
	VeicID       string
	UFID         string
}

// RegistroD180 - Modais
type RegistroD180 struct {
	NumSeq     string
	IndEmit    IndEmit
	CNPJEmit   string
	UFEmit     string
	IEEmit     string
	CodMunOrig string
	CNPJCPFTom string
	UFTom      string
	IETom      string
	CodMunDest string
	CodMod     string
	Ser        string
	Sub        string
	NumDoc     string
	DtDoc      time.Time
	VlDoc      float64
}

// RegistroD190 - Registro Analitico dos Documentos
type RegistroD190 struct {
	CstICMS  CstIcms
	CFOP     string
	AliqICMS float64
	VlOpr    float64
	VlBcICMS float64
	VlICMS   float64
	VlRedBC  float64
	CodObs   string
}

// RegistroD195 - Observacoes do Lancamento Fiscal
type RegistroD195 struct {
	CodObs       string
	TxtCompl     string
	RegistroD197 []*RegistroD197
}

// RegistroD197 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroD197 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroD300 - Registro Analitico dos Bilhetes Consolidados
type RegistroD300 struct {
	CodMod       string
	Ser          string
	Sub          string
	NumDocIni    string
	NumDocFin    string
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	DtDoc        time.Time
	VlOpr        float64
	VlDesc       float64
	VlServ       float64
	VlSeg        float64
	VlOutDesp    float64
	VlBcICMS     float64
	VlICMS       float64
	VlRedBC      float64
	CodObs       string
	CodCta       string
	RegistroD301 []*RegistroD301
	RegistroD310 []*RegistroD310
}

// RegistroD301 - Documentos Cancelados dos Bilhetes
type RegistroD301 struct {
	NumDocCanc string
}

// RegistroD310 - Complemento dos Bilhetes
type RegistroD310 struct {
	CodMunOrig string
	VlServ     float64
	VlBcICMS   float64
	VlICMS     float64
}

// RegistroD350 - Equipamento ECF
type RegistroD350 struct {
	CodMod       string
	EcfMod       string
	EcfFab       string
	EcfCx        string
	RegistroD355 []*RegistroD355
}

// RegistroD355 - Reducao Z
type RegistroD355 struct {
	DtDoc        time.Time
	Cro          int
	Crz          int
	NumCooFin    int
	GtFin        float64
	VlBrt        float64
	RegistroD360 []*RegistroD360
	RegistroD365 []*RegistroD365
	RegistroD390 []*RegistroD390
}

// RegistroD360 - PIS e COFINS Totalizados no Dia
type RegistroD360 struct {
	VlPIS    float64
	VlCOFINS float64
}

// RegistroD365 - Registro dos Totalizadores Parciais da Reducao Z
type RegistroD365 struct {
	CodTotPar    string
	VlrAcumTot   float64
	NrTot        string
	DescrNrTot   string
	RegistroD370 []*RegistroD370
}

// RegistroD370 - Complemento dos Documentos Informados
type RegistroD370 struct {
	CodMunOrig string
	VlServ     float64
	QtdBilh    int
	VlBcICMS   float64
	VlICMS     float64
}

// RegistroD390 - Registro Analitico do Movimento Diario
type RegistroD390 struct {
	CstICMS   string
	CFOP      string
	AliqICMS  float64
	VlOpr     float64
	VlBcISSQN float64
	AliqISSQN float64
	VlISSQN   float64
	VlBcICMS  float64
	VlICMS    float64
	CodObs    string
}

// RegistroD400 - Resumo do Movimento Diario
type RegistroD400 struct {
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	Sub          string
	NumDoc       string
	DtDoc        time.Time
	VlDoc        float64
	VlDesc       float64
	VlServ       float64
	VlBcICMS     float64
	VlICMS       float64
	VlPIS        float64
	VlCOFINS     float64
	CodCta       string
	RegistroD410 []*RegistroD410
	RegistroD420 []*RegistroD420
}

// RegistroD410 - Documentos Informados
type RegistroD410 struct {
	CodMod       string
	Ser          string
	Sub          string
	NumDocIni    string
	NumDocFin    string
	DtDoc        time.Time
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlDesc       float64
	VlServ       float64
	VlBcICMS     float64
	VlICMS       float64
	RegistroD411 []*RegistroD411
}

// RegistroD411 - Documentos Cancelados dos Documentos Informados
type RegistroD411 struct {
	NumDocCanc string
}

// RegistroD420 - Complemento dos Documentos Informados
type RegistroD420 struct {
	CodMunOrig string
	VlServ     float64
	VlBcICMS   float64
	VlICMS     float64
}

// RegistroD500 - Nota Fiscal de Servico de Comunicacao (codigo 21) e
// de Telecomunicacao (codigo 22)
type RegistroD500 struct {
	IndOper      IndOper
	IndEmit      IndEmit
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	Sub          string
	NumDoc       string
	DtDoc        time.Time
	DtAP         time.Time
	VlDoc        float64
	VlDesc       float64
	VlServ       float64
	VlServNT     float64
	VlTerc       float64
	VlDa         float64
	VlBcICMS     float64
	VlICMS       float64
	CodInf       string
	VlPIS        float64
	VlCOFINS     float64
	CodCta       string
	TpAssinante  TpAssinante
	RegistroD510 []*RegistroD510
	RegistroD530 []*RegistroD530
	RegistroD590 []*RegistroD590
}

// RegistroD510 - Itens do Documento - Servico de Comunicacao e Telecomunicacao
type RegistroD510 struct {
	NumItem    string
	CodItem    string
	CodClass   string
	Qtd        float64
	Unid       string
	VlItem     float64
	VlDesc     float64
	CstICMS    string
	CFOP       string
	VlBcICMS   float64
	AliqICMS   float64
	VlICMS     float64
	VlBcICMSUF float64
	VlICMSUF   float64
	IndRec     IndReceita
	CodPart    string
	VlPIS      float64
	VlCOFINS   float64
	CodCta     string
}

// RegistroD530 - Terminal Faturado
type RegistroD530 struct {
	IndServ   ServicoPrestado
	DtIniServ time.Time
	DtFinServ time.Time
	PerFiscal string
	CodArea   string
	Terminal  string
}

// RegistroD590 - Registro Analitico do Documento
type RegistroD590 struct {
	CstICMS    string
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlBcICMSUF float64
	VlICMSUF   float64
	VlRedBC    float64
	CodObs     string
}

// RegistroD600 - Consolidacao da Prestacao de Servicos - Notas de Servico de
// Comunicacao e de Telecomunicacao
type RegistroD600 struct {
	CodMod       string
	CodMun       string
	Ser          string
	Sub          int
	CodCons      int
	QtdCons      int
	DtDoc        time.Time
	VlDoc        float64
	VlDesc       float64
	VlServ       float64
	VlServNT     float64
	VlTerc       float64
	VlDa         float64
	VlBcICMS     float64
	VlICMS       float64
	VlPIS        float64
	VlCOFINS     float64
	RegistroD610 []*RegistroD610
	RegistroD690 []*RegistroD690
}

// RegistroD610 - Itens do Documento Consolidado
type RegistroD610 struct {
	CodClass int
	CodItem  string
	Qtd      float64
	Unid     string
	VlItem   float64
}

// RegistroD690 - Registro Analitico dos Documentos Consolidados
type RegistroD690 struct {
	CstICMS    string
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlBcICMSUF float64
	VlICMSUF   float64
	VlRedBC    float64
	CodObs     string
}

// RegistroD695 - Consolidacao dos Documentos - Nota Fiscal de Servico de
// Comunicacao e Telecomunicacao
type RegistroD695 struct {
	CodMod       string
	Ser          string
	NroOrdIni    int
	NroOrdFin    int
	DtDocIni     time.Time
	DtDocFin     time.Time
	NomMest      string
	ChvCodDig    string
	RegistroD696 []*RegistroD696
}

// RegistroD696 - Registro Analitico dos Documentos Consolidados
type RegistroD696 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSUF   float64
	VlICMSUF     float64
	VlRedBC      float64
	CodObs       string
	RegistroD697 []*RegistroD697
}

// RegistroD697 - Registro de Informacoes de Outras UFs
type RegistroD697 struct {
	UF       string
	VlBcICMS float64
	VlICMS   float64
}

// RegistroD700 - Nota Fiscal Fatura Eletronica de Servicos de Comunicacao
// (codigo 62)
type RegistroD700 struct {
	IndOper      IndOper
	IndEmit      IndEmit
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	NumDoc       string
	DtDoc        time.Time
	DtES         time.Time
	VlDoc        float64
	VlDesc       float64
	VlServ       float64
	VlServNT     float64
	VlTerc       float64
	VlDa         float64
	VlBcICMS     float64
	VlICMS       float64
	CodInf       string
	VlPIS        float64
	VlCOFINS     float64
	ChvDOCe      string
	FinDOCe      FinEmissaoFaturaEletronica
	TipFat       TipoFaturamentoDocEletronico
	CodModDocRef string
	ChvDOCeRef   string
	HashDocRef   string
	SerDocRef    string
	NumDocRef    string
	MesDocRef    string
	CodMunDest   string
	Ded          float64
	RegistroD730 []*RegistroD730
	RegistroD735 []*RegistroD735
}

// RegistroD730 - Registro Analitico do Documento
type RegistroD730 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlRedBC      float64
	CodObs       string
	RegistroD731 []*RegistroD731
}

// RegistroD731 - Informacao Complementar do FCP
type RegistroD731 struct {
	VlFcpOp float64
}

// RegistroD735 - Observacoes do Lancamento Fiscal
type RegistroD735 struct {
	CodObs       string
	TxtCompl     string
	RegistroD737 []*RegistroD737
}

// RegistroD737 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroD737 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroD750 - Escrituracao Consolidada da Nota Fiscal Fatura Eletronica de
// Servicos de Comunicacao (codigo 62)
type RegistroD750 struct {
	CodMod       string
	Ser          string
	DtDoc        time.Time
	QtdCons      float64
	IndPrePago   IndFormaPagto
	VlDoc        float64
	VlServ       float64
	VlServNT     float64
	VlTerc       float64
	VlDesc       float64
	VlDa         float64
	VlBcICMS     float64
	VlICMS       float64
	VlPIS        float64
	VlCOFINS     float64
	Ded          float64
	RegistroD760 []*RegistroD760
}

// RegistroD760 - Registro Analitico do Documento Consolidado
type RegistroD760 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlRedBC      float64
	CodObs       string
	RegistroD761 []*RegistroD761
}

// RegistroD761 - Informacao Complementar do FCP
type RegistroD761 struct {
	VlFcpOp float64
}

// RegistroD990 - Encerramento do Bloco D
type RegistroD990 struct {
	QtdLinD int
}
