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
// Bloco C - Documentos Fiscais I - Mercadorias (ICMS/IPI)
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroC001 - Abertura do Bloco C
type RegistroC001 struct {
	OpenBlocos
	RegistroC100 []*RegistroC100
	RegistroC300 []*RegistroC300
	RegistroC350 []*RegistroC350
	RegistroC400 []*RegistroC400
	RegistroC495 []*RegistroC495
	RegistroC500 []*RegistroC500
	RegistroC600 []*RegistroC600
	RegistroC700 []*RegistroC700
	RegistroC800 []*RegistroC800
	RegistroC860 []*RegistroC860
}

// NewRegistroC001 cria um novo RegistroC001 com IndDad=1 (sem dados).
func NewRegistroC001() *RegistroC001 {
	return &RegistroC001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroC100 - Nota Fiscal (codigo 01), NF Avulsa (codigo 1B),
// NF Produtor (codigo 04), NF-e (codigo 55) e NFC-e (codigo 65)
type RegistroC100 struct {
	IndOper      IndOper
	IndEmit      IndEmit
	CodPart      string
	CodMod       string
	CodSit       CodSit
	Ser          string
	NumDoc       string
	ChvNFe       string
	DtDoc        time.Time
	DtES         time.Time
	VlDoc        float64
	IndPgto      IndPgto
	VlDesc       float64
	VlAbatNT     float64
	VlMerc       float64
	IndFrt       IndFrt
	VlFrt        float64
	VlSeg        float64
	VlOutDa      float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlIPI        float64
	VlPIS        float64
	VlCOFINS     float64
	VlPISST      float64
	VlCOFINSST   float64
	RegistroC101 []*RegistroC101
	RegistroC105 []*RegistroC105
	RegistroC110 []*RegistroC110
	RegistroC120 []*RegistroC120
	RegistroC170 []*RegistroC170
	RegistroC185 []*RegistroC185
	RegistroC186 []*RegistroC186
	RegistroC190 []*RegistroC190
	RegistroC195 []*RegistroC195
}

// RegistroC101 - Informacao Complementar dos Documentos Fiscais quando das
// operacoes interestaduais destinadas a consumidor final nao contribuinte (EC 87/15)
type RegistroC101 struct {
	VlFcpUFDest  float64
	VlICMSUFDest float64
	VlICMSUFRem  float64
}

// RegistroC105 - Operacoes com ICMS ST Recolhido para UF Diversa
type RegistroC105 struct {
	Oper IndTipoOperacaoST
	UF   string
}

// RegistroC110 - Informacao Complementar da Nota Fiscal (codigo 01, 1B, 04, 55)
type RegistroC110 struct {
	CodInf       string
	TxtCompl     string
	RegistroC111 []*RegistroC111
	RegistroC112 []*RegistroC112
	RegistroC113 []*RegistroC113
	RegistroC114 []*RegistroC114
}

// RegistroC111 - Processo Referenciado
type RegistroC111 struct {
	NumProc string
	IndProc OrigemProcesso
}

// RegistroC112 - Documento de Arrecadacao Referenciado
type RegistroC112 struct {
	CodDa  DoctoArrecada
	UF     string
	NumDa  string
	CodAut string
	VlDa   float64
	DtVcto time.Time
	DtPgto time.Time
}

// RegistroC113 - Documento Fiscal Referenciado
type RegistroC113 struct {
	IndOper IndOper
	IndEmit IndEmit
	CodPart string
	CodMod  string
	Ser     string
	Sub     string
	NumDoc  string
	DtDoc   time.Time
	ChvDocE string
}

// RegistroC114 - Cupom Fiscal Referenciado
type RegistroC114 struct {
	CodMod string
	EcfFab string
	EcfCx  string
	NumDoc string
	DtDoc  time.Time
}

// RegistroC120 - Documento de Importacao (codigo 01)
type RegistroC120 struct {
	CodDocImp DoctoImporta
	NumDocImp string
	PisImp    float64
	CofinsImp float64
	NumAcdraw string
}

// RegistroC170 - Itens do Documento (codigo 01, 1B, 04, 55)
type RegistroC170 struct {
	NumItem         string
	CodItem         string
	DescrCompl      string
	Qtd             float64
	Unid            string
	VlItem          float64
	VlDesc          float64
	IndMov          IndMovFisica
	CstICMS         CstIcms
	CFOP            string
	CodNat          string
	VlBcICMS        float64
	AliqICMS        float64
	VlICMS          float64
	VlBcICMSST      float64
	AliqST          float64
	VlICMSST        float64
	IndApur         ApuracaoIPI
	CstIPI          CstIpi
	CodEnq          string
	VlBcIPI         float64
	AliqIPI         float64
	VlIPI           float64
	CstPIS          CstPis
	VlBcPIS         float64
	AliqPIS         float64
	QtdBcPIS        float64
	AliqPISReais    float64
	VlPIS           float64
	CstCOFINS       CstCofins
	VlBcCOFINS      float64
	AliqCOFINS      float64
	QtdBcCOFINS     float64
	AliqCOFINSReais float64
	VlCOFINS        float64
	CodCta          string
	VlAbatNT        float64
}

// RegistroC185 - Informacoes complementares das operacoes de saida de
// mercadorias sujeitas a substituicao tributaria (codigo 55)
type RegistroC185 struct {
	NumOp      string
	CodItem    string
	CstICMS    CstIcms
	CFOP       string
	CodMot     MotivoRessarcimento
	VlOpr      float64
	VlBcICMS   float64
	AliqICMS   float64
	VlICMS     float64
	VlBcICMSST float64
	VlICMSST   float64
	VlFCP      float64
	VlFCPST    float64
}

// RegistroC186 - Informacoes complementares das operacoes de entrada de
// mercadorias sujeitas a substituicao tributaria (codigo 55)
type RegistroC186 struct {
	NumOp      string
	CodItem    string
	CstICMS    CstIcms
	CFOP       string
	CodMot     MotivoRessarcimento
	VlOpr      float64
	VlBcICMS   float64
	AliqICMS   float64
	VlICMS     float64
	VlBcICMSST float64
	VlICMSST   float64
	VlFCP      float64
	VlFCPST    float64
}

// RegistroC190 - Registro Analitico do Documento (codigo 01, 1B, 04, 55, 65)
type RegistroC190 struct {
	CstICMS    CstIcms
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlBcICMSST float64
	VlICMSST   float64
	VlRedBC    float64
	VlIPI      float64
	CodObs     string
}

// RegistroC195 - Observacoes do Lancamento Fiscal (codigo 01, 1B, 55)
type RegistroC195 struct {
	CodObs       string
	TxtCompl     string
	RegistroC197 []*RegistroC197
}

// RegistroC197 - Outras Obrigacoes Tributarias, Ajustes e Informacoes de
// Valores Provenientes de Documento Fiscal
type RegistroC197 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// ---------------------------------------------------------------------------
// Registros C300..C860 - Documentos Fiscais Diversos
// ---------------------------------------------------------------------------

// RegistroC300 - Resumo Diario das Notas Fiscais de Venda a Consumidor (codigo 02)
type RegistroC300 struct {
	CodMod    string
	Ser       string
	Sub       string
	NumDocIni string
	NumDocFin string
	DtDoc     time.Time
	VlDoc     float64
	VlPIS     float64
	VlCOFINS  float64
	CodCta    string
}

// RegistroC350 - Nota Fiscal de Venda a Consumidor (codigo 02)
type RegistroC350 struct {
	Ser      string
	Sub      string
	NumDoc   string
	DtDoc    time.Time
	CNPJCPF  string
	VlMerc   float64
	VlDoc    float64
	VlDesc   float64
	VlPIS    float64
	VlCOFINS float64
	CodCta   string
}

// RegistroC400 - Equipamento ECF (codigo 02, 2D)
type RegistroC400 struct {
	CodMod       string
	EcfMod       string
	EcfFab       string
	EcfCx        string
	RegistroC405 []*RegistroC405
}

// RegistroC405 - Reducao Z (codigo 02, 2D)
type RegistroC405 struct {
	DtDoc     time.Time
	Cro       int
	Crz       int
	NumCooFin int
	GtFin     float64
	VlBrt     float64
}

// RegistroC495 - Resumo Mensal de Itens do ECF por Estabelecimento (codigo 02, 2D)
type RegistroC495 struct {
	AliqICMS float64
	CodItem  string
	Qtd      float64
	QtdCanc  float64
	Unid     string
	VlItem   float64
	VlDesc   float64
	VlCanc   float64
	VlAcmo   float64
	VlBcICMS float64
	VlICMS   float64
	VlISEN   float64
	VlNT     float64
	VlICMSST float64
}

// RegistroC500 - Nota Fiscal/Conta de Energia Eletrica (codigo 06),
// Nota Fiscal/Conta de Fornecimento d'Agua Canalizada (codigo 29),
// Nota Fiscal Consumo de Gas (codigo 28) e NF3e (codigo 66)
type RegistroC500 struct {
	IndOper        IndOper
	IndEmit        IndEmit
	CodPart        string
	CodMod         string
	CodSit         CodSit
	Ser            string
	Sub            string
	CodCons        ClasseConsumo
	NumDoc         string
	DtDoc          time.Time
	DtES           time.Time
	VlDoc          float64
	VlDesc         float64
	VlFornEC       float64
	VlServNT       float64
	VlTerc         float64
	VlDa           float64
	VlBcICMS       float64
	VlICMS         float64
	VlBcICMSST     float64
	VlICMSST       float64
	CodInf         string
	VlPIS          float64
	VlCOFINS       float64
	TpLigacao      TpLigacao
	CodGrupoTensao GrupoTensao
	ChvDOCe        string
	FinDOCe        FinalidadeEmissaoDocEletronico
	ChvDOCeRef     string
	IndDest        IndDestinatarioAcessante
	CodMunDest     string
	CodCContab     string
}

// RegistroC600 - Consolidacao Diaria de Notas Fiscais/Contas de Energia Eletrica
// (codigo 06), Nota Fiscal/Conta de Fornecimento d'Agua (codigo 29) e
// Nota Fiscal/Conta de Gas (codigo 28)
type RegistroC600 struct {
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
	VlFornEC   float64
	VlServNT   float64
	VlTerc     float64
	VlDa       float64
	VlBcICMS   float64
	VlICMS     float64
	VlBcICMSST float64
	VlICMSST   float64
	VlPIS      float64
	VlCOFINS   float64
}

// RegistroC700 - Consolidacao dos Documentos NF/Conta de Energia Eletrica
// (codigo 06), NF/Conta de Fornecimento d'Agua (codigo 29) e
// NF/Conta de Fornecimento de Gas (codigo 28) - documentos de saida
type RegistroC700 struct {
	CodMod    string
	Ser       string
	NroOrdIni string
	NroOrdFin string
	DtDocIni  time.Time
	DtDocFin  time.Time
	NomMest   string
	ChvCodDig string
}

// RegistroC800 - Cupom Fiscal Eletronico - SAT (CF-e-SAT) (codigo 59)
type RegistroC800 struct {
	CodMod     string
	CodSit     CodSit
	NumCFe     string
	DtDoc      time.Time
	VlCFe      float64
	VlPIS      float64
	VlCOFINS   float64
	CNPJCPFOp  string
	VlDesc     float64
	VlMerc     float64
	VlOutDa    float64
	VlICMS     float64
	VlPISST    float64
	VlCOFINSST float64
}

// RegistroC860 - Identificacao do Equipamento SAT-CF-e
type RegistroC860 struct {
	CodMod string
	NrSat  string
	DtDoc  time.Time
	DocIni string
	DocFin string
}

// RegistroC990 - Encerramento do Bloco C
type RegistroC990 struct {
	QtdLinC int
}
