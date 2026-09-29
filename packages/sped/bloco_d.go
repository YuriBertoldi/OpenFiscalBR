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

// NewRegistroD001 cria um novo RegistroD001 com IndDad=1 (sem dados).
func NewRegistroD001() *RegistroD001 {
	return &RegistroD001{OpenBlocos: OpenBlocos{IndDad: 1}}
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

// RegistroD110 - Itens do Documento - Nota Fiscal de Servicos de Transporte (codigo 07)
type RegistroD110 struct {
	NumItem      string
	CodItem      string
	VlServ       float64
	VlOutro      float64
	RegistroD120 []*RegistroD120
}

// RegistroD120 - Complemento da Nota Fiscal de Servicos de Transporte (codigo 07)
type RegistroD120 struct {
	CodMunOrig string
	CodMunDest string
	VeicID     string
	UfID       string
}

// RegistroD130 - Complemento do Conhecimento Rodoviario de Cargas (codigo 08)
// e Conhecimento Rodoviario de Cargas Avulso (codigo 8B)
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
	UfID         string
}

// RegistroD140 - Complemento do Conhecimento Aquaviario de Cargas (codigo 09)
type RegistroD140 struct {
	CodPartConsg string
	CodMunOrig   string
	CodMunDest   string
	IndVeic      TipoVeiculo
	VeicID       string
	IndNav       TipoNavegacao
	Viagem       string
	VlFrtLiq     float64
	VlDesp       float64
	VlFrtBrt     float64
	VlBrt        float64
}

// RegistroD150 - Complemento do Conhecimento Aereo de Cargas (codigo 10)
type RegistroD150 struct {
	CodMunOrig   string
	CodMunDest   string
	VeicID       string
	ViagNac      string
	VlLiqFrt     float64
	VlDesp       float64
	VlTot        float64
	IndNFCarga   string
	VlPesoBrt    float64
	VlPesoLiq    float64
}

// RegistroD160 - Complemento do CT-e (codigo 57), CT-e Avulso e
// CT-e de Substituicao
type RegistroD160 struct {
	DESPACHO     string
	CNPJCPFRem   string
	IERem        string
	CodMunOrig   string
	CNPJCPFDest  string
	IEDest       string
	CodMunDest   string
	RegistroD161 []*RegistroD161
	RegistroD162 []*RegistroD162
}

// RegistroD161 - Local da Coleta e Entrega (CT-e)
type RegistroD161 struct {
	IndCarga     string
	CNPJCPFCol   string
	IECol        string
	CodMunCol    string
	CNPJCPFEntg  string
	IEEntg       string
	CodMunEntg   string
}

// RegistroD162 - Identificacao dos Documentos Fiscais (CT-e)
type RegistroD162 struct {
	CodMod   string
	Ser      string
	NumDoc   string
	DtDoc    time.Time
	VlDoc    float64
	VlDocFrt float64
}

// RegistroD170 - Complemento do Conhecimento Multimodal de Cargas (codigo 26)
type RegistroD170 struct {
	CodPartConsg string
	CodPartRed   string
	CodMunOrig   string
	CodMunDest   string
	OtmCar       string
	VlLiqFrt     float64
	VlGris       float64
	VlPedg       float64
	VlOut        float64
	VlFrt        float64
	VeicID       string
	UfID         string
}

// RegistroD180 - Modais (CT-e)
type RegistroD180 struct {
	NumSeq  string
	IndEmit IndEmit
	CNPJCPFEmit string
	UFEmit  string
	IEEmit  string
	CodMunEmit string
	CNPJCPF string
	UF      string
	IE      string
	CodMun  string
	CodMod  string
	Ser     string
	Sub     string
	NumDoc  string
	DtDoc   time.Time
	VlDoc   float64
}

// RegistroD190 - Registro Analitico dos Documentos (codigo 07, 08, 8B, 09, 10,
// 11, 26, 27, 57, 67)
type RegistroD190 struct {
	CstICMS    CstIcms
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlRedBC    float64
	CodObs     string
}

// RegistroD195 - Observacoes do Lancamento Fiscal (codigo 07, 08, 8B, 09, 10,
// 11, 26, 27, 57, 67)
type RegistroD195 struct {
	CodObs       string
	TxtCompl     string
	RegistroD197 []*RegistroD197
}

// RegistroD197 - Outras Obrigacoes Tributarias, Ajustes e Informacoes de
// Valores Provenientes de Documento Fiscal
type RegistroD197 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroD300 - Nota Fiscal de Servico de Transporte (codigo 07) para
// Aquaviario (codigo 09), Aereo (codigo 10), Ferroviario (codigo 11)
// consolidados
type RegistroD300 struct {
	CodMod     string
	Ser        string
	Sub        string
	NumDocIni  string
	NumDocFin  string
	CstICMS    CstIcms
	CFOP       string
	AliqICMS   float64
	DtDoc      time.Time
	VlOpr      float64
	VlDesc     float64
	VlServ     float64
	VlBcICMS   float64
	VlICMS     float64
}

// RegistroD350 - Equipamento ECF (codigos 2E, 13, 14, 15, 16)
type RegistroD350 struct {
	CodMod string
	EcfMod string
	EcfFab string
	EcfCx  string
}

// RegistroD400 - Resumo de Movimento Diario (codigo 18)
type RegistroD400 struct {
	CodPart    string
	CodMod     string
	CodSit     CodSit
	Ser        string
	Sub        string
	NumDoc     string
	DtDoc      time.Time
	VlDoc      float64
	VlDesc     float64
	VlServ     float64
	VlBcICMS   float64
	VlICMS     float64
	VlPIS      float64
	VlCOFINS   float64
	CodCta     string
}

// RegistroD500 - Nota Fiscal de Servico de Comunicacao (codigo 21),
// Nota Fiscal de Servico de Telecomunicacao (codigo 22)
type RegistroD500 struct {
	IndOper    IndOper
	IndEmit    IndEmit
	CodPart    string
	CodMod     string
	CodSit     CodSit
	Ser        string
	Sub        string
	NumDoc     string
	DtDoc      time.Time
	DtAP       time.Time
	VlDoc      float64
	VlDesc     float64
	VlServ     float64
	VlServNT   float64
	VlTerc     float64
	VlDa       float64
	VlBcICMS   float64
	VlICMS     float64
	CodInf     string
	VlPIS      float64
	VlCOFINS   float64
	CodCta     string
	TpAssinante TpAssinante
}

// RegistroD600 - Consolidacao da Prestacao de Servicos - Notas Fiscais de
// Servico de Comunicacao (codigo 21) e de Servico de Telecomunicacao (codigo 22)
type RegistroD600 struct {
	CodMod     string
	CodMun     string
	Ser        string
	Sub        string
	CodCons    ClasseConsumo
	QtdCons    int
	QtdCanc    int
	DtDoc      time.Time
	VlDoc      float64
	VlDesc     float64
	VlServ     float64
	VlServNT   float64
	VlTerc     float64
	VlDa       float64
	VlBcICMS   float64
	VlICMS     float64
}

// RegistroD695 - Consolidacao da Prestacao de Servicos - Notas Fiscais de
// Servico de Comunicacao (codigo 21) e de Telecomunicacao (codigo 22)
type RegistroD695 struct {
	CodMod       string
	Ser          string
	NroOrdIni    string
	NroOrdFin    string
	DtDocIni     time.Time
	DtDocFin     time.Time
	NomMest      string
	ChvCodDig    string
	RegistroD696 []*RegistroD696
}

// RegistroD696 - Registro Analitico dos Documentos (codigos 21 e 22)
type RegistroD696 struct {
	CstICMS      CstIcms
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlRedBC      float64
	CodObs       string
	RegistroD697 []*RegistroD697
}

// RegistroD697 - Registro de Informacoes de Outras UFs, relativamente aos
// servicos nao medidos de TV por assinatura, provimento de acesso a Internet
type RegistroD697 struct {
	UF       string
	VlBcICMS float64
	VlICMS   float64
}

// RegistroD700 - Nota Fiscal Fatura Eletronica de Servicos de Comunicacao - NFCom
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
	VlPIS        float64
	VlCOFINS     float64
	CodInf       string
	ChvNFCom     string
	FinNFCom     FinEmissaoFaturaEletronica
	TpFat        TipoFaturamentoDocEletronico
	CodMod_DocRef string
	ChvNFComRef  string
	IndDest      IndDestinatarioAcessante
	CodMunDest   string
	RegistroD730 []*RegistroD730
	RegistroD735 []*RegistroD735
}

// RegistroD730 - Registro Analitico da NFCom (codigo 62)
type RegistroD730 struct {
	CstICMS    CstIcms
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlRedBC    float64
	CodObs     string
	RegistroD731 []*RegistroD731
}

// RegistroD731 - Informacoes do Fundo de Combate a Pobreza - FCP (NFCom)
type RegistroD731 struct {
	VlFcpOp float64
}

// RegistroD735 - Observacoes do Lancamento Fiscal (NFCom)
type RegistroD735 struct {
	CodObs       string
	TxtCompl     string
	RegistroD737 []*RegistroD737
}

// RegistroD737 - Outras Obrigacoes Tributarias, Ajustes e Informacoes de
// Valores Provenientes de Documento Fiscal
type RegistroD737 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroD750 - Escrituracao Consolidada da NFCom (codigo 62)
type RegistroD750 struct {
	CodMod       string
	Ser          string
	DtDoc        time.Time
	QtdCons      int
	IndPrepago   IndFormaPagto
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
	RegistroD760 []*RegistroD760
}

// RegistroD760 - Registro Analitico da Consolidacao NFCom (codigo 62)
type RegistroD760 struct {
	CstICMS      CstIcms
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlRedBC      float64
	CodObs       string
	RegistroD761 []*RegistroD761
}

// RegistroD761 - Informacoes do Fundo de Combate a Pobreza - FCP (Consolidacao NFCom)
type RegistroD761 struct {
	VlFcpOp float64
}

// RegistroD990 - Encerramento do Bloco D
type RegistroD990 struct {
	QtdLinD int
}
