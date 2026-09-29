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

// NewRegistroC001 cria um novo RegistroC001 com IndMov=1 (sem dados).
func NewRegistroC001() *RegistroC001 {
	return &RegistroC001{OpenBlocos: OpenBlocos{IndMov: 1}}
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
	RegistroC130 []*RegistroC130
	RegistroC140 []*RegistroC140
	RegistroC160 []*RegistroC160
	RegistroC165 []*RegistroC165
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

// RegistroC105 - Operacoes com ICMS ST recolhido para UF diversa
type RegistroC105 struct {
	Oper IndTipoOperacaoST
	UF   string
}

// RegistroC110 - Complemento do Documento - Informacoes Complementares
type RegistroC110 struct {
	CodInf       string
	TxtCompl     string
	RegistroC111 []*RegistroC111
	RegistroC112 []*RegistroC112
	RegistroC113 []*RegistroC113
	RegistroC114 []*RegistroC114
	RegistroC115 []*RegistroC115
	RegistroC116 []*RegistroC116
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

// RegistroC120 - Complemento de Documento - Operacoes de Importacao
type RegistroC120 struct {
	CodDocImp DoctoImporta
	NumDocImp string
	PisImp    float64
	CofinsImp float64
	NumACDraw string
}

// RegistroC115 - Local de Coleta e/ou Entrega
type RegistroC115 struct {
	IndCarga   TipoTransporte
	CNPJCol    string
	IECol      string
	CPFCol     string
	CodMunCol  string
	CNPJEntg   string
	IEEntg     string
	CPFEntg    string
	CodMunEntg string
}

// RegistroC116 - Cupom Fiscal Eletronico Referenciado
type RegistroC116 struct {
	CodMod string
	NrSat  string
	ChvCFe string
	NumCFe string
	DtDoc  time.Time
}

// RegistroC130 - ISSQN, IRRF e Previdencia Social
type RegistroC130 struct {
	VlServNT  float64
	VlBcISSQN float64
	VlISSQN   float64
	VlBcIRRF  float64
	VlIRRF    float64
	VlBcPrev  float64
	VlPrev    float64
}

// RegistroC140 - Fatura (codigo 01)
type RegistroC140 struct {
	IndEmit      IndEmit
	IndTit       TipoTitulo
	DescTit      string
	NumTit       string
	QtdParc      int
	VlTit        float64
	RegistroC141 []*RegistroC141
}

// RegistroC141 - Vencimento da Fatura (codigo 01)
type RegistroC141 struct {
	NumParc string
	DtVcto  time.Time
	VlParc  float64
}

// RegistroC160 - Volumes Transportados (codigo 01 e 04), exceto Combustiveis
type RegistroC160 struct {
	CodPart string
	VeicID  string
	QtdVol  int
	PesoBrt float64
	PesoLiq float64
	UFID    string
}

// RegistroC165 - Operacoes com Combustiveis (codigo 01)
type RegistroC165 struct {
	CodPart string
	VeicID  string
	CodAut  string
	NrPasse string
	Hora    string
	Temper  string
	QtdVol  int
	PesoBrt float64
	PesoLiq float64
	NomMot  string
	CPF     string
	UFID    string
}

// RegistroC170 - Itens do Documento
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
	RegistroC171    []*RegistroC171
	RegistroC172    []*RegistroC172
	RegistroC173    []*RegistroC173
	RegistroC174    []*RegistroC174
	RegistroC175    []*RegistroC175
	RegistroC176    []*RegistroC176
	RegistroC177    []*RegistroC177
	RegistroC178    []*RegistroC178
	RegistroC179    []*RegistroC179
	RegistroC180    []*RegistroC180
	RegistroC181    []*RegistroC181
}

// RegistroC171 - Armazenamento de Combustiveis
type RegistroC171 struct {
	NumTanque string
	Qtde      float64
}

// RegistroC172 - Operacoes com ISSQN
type RegistroC172 struct {
	VlBcISSQN float64
	AliqISSQN float64
	VlISSQN   float64
}

// RegistroC173 - Operacoes com Medicamentos
type RegistroC173 struct {
	LoteMed  string
	QtdItem  float64
	DtFab    time.Time
	DtVal    time.Time
	IndMed   TipoBaseMedicamento
	TpProd   TipoProduto
	VlTabMax float64
}

// RegistroC174 - Operacoes com Armas de Fogo
type RegistroC174 struct {
	IndArm     TipoArmaFogo
	NumArm     string
	DescrCompl string
}

// RegistroC175 - Operacoes com Veiculos Novos
type RegistroC175 struct {
	IndVeicOper IndVeicOper
	CNPJ        string
	UF          string
	ChassiVeic  string
}

// RegistroC176 - Ressarcimento de ICMS em operacoes com Substituicao Tributaria
type RegistroC176 struct {
	CodModUltE             string
	NumDocUltE             string
	SerUltE                string
	DtUltE                 time.Time
	CodPartUltE            string
	QuantUltE              float64
	VlUnitUltE             float64
	VlUnitBcST             float64
	ChaveNfeUltE           string
	NumItemUltE            string
	VlUnitBcICMSUltE       float64
	AliqICMSUltE           float64
	VlUnitLimiteBcICMSUltE float64
	VlUnitICMSUltE         float64
	AliqSTUltE             float64
	VlUnitRes              float64
	CodRespRet             string
	CodMotRes              MotivoRessarcimento
	ChaveNfeRet            string
	CodPartNfeRet          string
	SerNfeRet              string
	NumNfeRet              string
	ItemNfeRet             string
	CodDA                  string
	NumDA                  string
	VlUnitResFcpST         float64
}

// RegistroC177 - Operacoes com Produtos Sujeitos a Selo de Controle IPI
type RegistroC177 struct {
	CodSeloIPI string
	QtSeloIPI  float64
	CodInfItem string
}

// RegistroC178 - Operacoes com Produtos Sujeitos a Tributacao de IPI por
// Unidade ou Quantidade de Produto
type RegistroC178 struct {
	ClEnq    string
	VlUnid   float64
	QuantPad float64
}

// RegistroC179 - Informacoes Complementares ST
type RegistroC179 struct {
	BcSTOrigDest float64
	ICMSSTRep    float64
	ICMSSTCompl  float64
	BcRet        float64
	ICMSRet      float64
}

// RegistroC180 - Informacoes Complementares das Operacoes de Entrada de
// Mercadorias Sujeitas a Substituicao Tributaria
type RegistroC180 struct {
	CodRespRet         string
	QuantConv          float64
	Unid               string
	VlUnitConv         float64
	VlUnitICMSOpConv   float64
	VlUnitBcICMSSTConv float64
	VlUnitICMSSTConv   float64
	VlUnitFcpSTConv    float64
	CodDA              string
	NumDA              string
}

// RegistroC181 - Informacoes Complementares das Operacoes de Saida de
// Mercadorias Sujeitas a Substituicao Tributaria
type RegistroC181 struct {
	CodMotRestCompl                 string
	QuantConv                       float64
	Unid                            string
	CodModSaida                     string
	SerieSaida                      string
	EcfFabSaida                     string
	NumDocSaida                     string
	ChvDfeSaida                     string
	DtDocSaida                      time.Time
	NumItemSaida                    string
	VlUnitConvSaida                 float64
	VlUnitICMSOpEstoqueConvSaida    *float64
	VlUnitICMSSTEstoqueConvSaida    *float64
	VlUnitFcpICMSSTEstoqueConvSaida *float64
	VlUnitICMSNaOperacaoConvSaida   *float64
	VlUnitICMSOpConvSaida           *float64
	VlUnitICMSSTConvRest            *float64
	VlUnitFcpSTConvRest             *float64
	VlUnitICMSSTConvCompl           *float64
	VlUnitFcpSTConvCompl            *float64
}

// RegistroC185 - Informacoes Complementares das Operacoes de Saida de
// Mercadorias Sujeitas a Substituicao Tributaria
type RegistroC185 struct {
	NumItem                    string
	CodItem                    string
	CstICMS                    string
	CFOP                       string
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   *float64
	VlUnitICMSOpConv           *float64
	VlUnitICMSOpEstoqueConv    *float64
	VlUnitICMSSTEstoqueConv    *float64
	VlUnitFcpICMSSTEstoqueConv *float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
}

// RegistroC186 - Informacoes Complementares das Operacoes de Saida de
// Mercadorias Sujeitas a Substituicao Tributaria - Documento de Entrada
type RegistroC186 struct {
	NumItem                   string
	CodItem                   string
	CstICMS                   string
	CFOP                      string
	CodMotRestCompl           string
	QuantConv                 float64
	Unid                      string
	CodModEntrada             string
	SerieEntrada              string
	NumDocEntrada             string
	ChvDfeEntrada             string
	DtDocEntrada              time.Time
	NumItemEntrada            string
	VlUnitConvEntrada         float64
	VlUnitICMSOpConvEntrada   float64
	VlUnitBcICMSSTConvEntrada float64
	VlUnitICMSSTConvEntrada   float64
	VlUnitFcpSTConvEntrada    float64
}

// RegistroC190 - Registro Analitico do Documento
type RegistroC190 struct {
	CstICMS      CstIcms
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlRedBC      float64
	VlIPI        float64
	CodObs       string
	RegistroC191 []*RegistroC191
}

// RegistroC191 - Informacoes do FCP na NF-e
type RegistroC191 struct {
	VlFcpOp  float64
	VlFcpST  float64
	VlFcpRet float64
}

// RegistroC195 - Observacoes do Lancamento Fiscal
type RegistroC195 struct {
	CodObs       string
	TxtCompl     string
	RegistroC197 []*RegistroC197
}

// RegistroC197 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroC197 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroC300 - Resumo Diario de Notas Fiscais de Venda a Consumidor
type RegistroC300 struct {
	CodMod       string
	Ser          string
	Sub          string
	NumDocIni    string
	NumDocFin    string
	DtDoc        time.Time
	VlDoc        float64
	VlPIS        float64
	VlCOFINS     float64
	CodCta       string
	RegistroC310 []*RegistroC310
	RegistroC320 []*RegistroC320
}

// RegistroC310 - Documentos Cancelados de Notas Fiscais de Venda a Consumidor
type RegistroC310 struct {
	NumDocCanc string
}

// RegistroC320 - Registro Analitico do Resumo Diario
type RegistroC320 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlRedBC      float64
	CodObs       string
	RegistroC321 []*RegistroC321
}

// RegistroC321 - Itens do Resumo Diario
type RegistroC321 struct {
	CodItem      string
	Qtd          float64
	Unid         string
	VlItem       float64
	VlDesc       float64
	VlBcICMS     float64
	VlICMS       float64
	VlPIS        float64
	VlCOFINS     float64
	RegistroC330 []*RegistroC330
}

// RegistroC330 - Complemento dos Itens do Resumo Diario - Ressarcimento
type RegistroC330 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   float64
	VlUnitICMSOpConv           float64
	VlUnitICMSOpEstoqueConv    float64
	VlUnitICMSSTEstoqueConv    float64
	VlUnitFcpICMSSTEstoqueConv float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
}

// RegistroC350 - Nota Fiscal de Venda a Consumidor (codigo 02)
type RegistroC350 struct {
	Ser          string
	SubSer       string
	NumDoc       string
	DtDoc        time.Time
	CNPJCPF      string
	VlMerc       float64
	VlDoc        float64
	VlDesc       float64
	VlPIS        float64
	VlCOFINS     float64
	CodCta       string
	RegistroC370 []*RegistroC370
	RegistroC390 []*RegistroC390
}

// RegistroC370 - Itens da Nota Fiscal de Venda a Consumidor
type RegistroC370 struct {
	NumItem      string
	CodItem      string
	Qtd          float64
	Unid         string
	VlItem       float64
	VlDesc       float64
	RegistroC380 []*RegistroC380
}

// RegistroC380 - Complemento dos Itens - Ressarcimento
type RegistroC380 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   float64
	VlUnitICMSOpConv           float64
	VlUnitICMSOpEstoqueConv    float64
	VlUnitICMSSTEstoqueConv    float64
	VlUnitFcpICMSSTEstoqueConv float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
	CstICMS                    string
	CFOP                       string
}

// RegistroC390 - Registro Analitico das Notas Fiscais de Venda a Consumidor
type RegistroC390 struct {
	CstICMS  string
	CFOP     string
	AliqICMS float64
	VlOpr    float64
	VlBcICMS float64
	VlICMS   float64
	VlRedBC  float64
	CodObs   string
}

// RegistroC400 - Equipamento ECF
type RegistroC400 struct {
	CodMod       string
	EcfMod       string
	EcfFab       string
	EcfCx        string
	RegistroC405 []*RegistroC405
}

// RegistroC405 - Reducao Z
type RegistroC405 struct {
	DtDoc        time.Time
	Cro          int
	Crz          int
	NumCooFin    int
	GtFin        float64
	VlBrt        float64
	RegistroC410 []*RegistroC410
	RegistroC420 []*RegistroC420
	RegistroC460 []*RegistroC460
	RegistroC490 []*RegistroC490
}

// RegistroC410 - PIS e COFINS Totalizados no Dia
type RegistroC410 struct {
	VlPIS    float64
	VlCOFINS float64
}

// RegistroC420 - Registro dos Totalizadores Parciais da Reducao Z
type RegistroC420 struct {
	CodTotPar    string
	VlrAcumTot   float64
	NrTot        int
	DescrNrTot   string
	RegistroC425 []*RegistroC425
}

// RegistroC425 - Resumo de Itens do Movimento Diario
type RegistroC425 struct {
	CodItem      string
	Qtd          float64
	Unid         string
	VlItem       float64
	VlPIS        float64
	VlCOFINS     float64
	RegistroC430 []*RegistroC430
}

// RegistroC430 - Complemento dos Itens - Ressarcimento
type RegistroC430 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   float64
	VlUnitICMSOpConv           float64
	VlUnitICMSOpEstoqueConv    float64
	VlUnitICMSSTEstoqueConv    float64
	VlUnitFcpICMSSTEstoqueConv float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
	CstICMS                    string
	CFOP                       string
}

// RegistroC460 - Documento Fiscal Emitido por ECF
type RegistroC460 struct {
	CodMod       string
	CodSit       CodSit
	NumDoc       string
	DtDoc        time.Time
	VlDoc        float64
	VlPIS        float64
	VlCOFINS     float64
	CPFCNPJ      string
	NomAdq       string
	RegistroC465 []*RegistroC465
	RegistroC470 []*RegistroC470
}

// RegistroC465 - Complemento do Cupom Fiscal Eletronico
type RegistroC465 struct {
	ChvCFe string
	NumCCF string
}

// RegistroC470 - Itens do Documento Fiscal Emitido por ECF
type RegistroC470 struct {
	CodItem      string
	Qtd          float64
	QtdCanc      float64
	Unid         string
	VlItem       float64
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlPIS        float64
	VlCOFINS     float64
	RegistroC480 []*RegistroC480
}

// RegistroC480 - Complemento dos Itens - Ressarcimento
type RegistroC480 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   float64
	VlUnitICMSOpConv           float64
	VlUnitICMSOpEstoqueConv    float64
	VlUnitICMSSTEstoqueConv    float64
	VlUnitFcpICMSSTEstoqueConv float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
	CstICMS                    string
	CFOP                       string
}

// RegistroC490 - Registro Analitico do Movimento Diario
type RegistroC490 struct {
	CstICMS  string
	CFOP     string
	AliqICMS float64
	VlOpr    float64
	VlBcICMS float64
	VlICMS   float64
	CodObs   string
}

// RegistroC495 - Resumo Mensal de Itens do ECF por Estabelecimento
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
	VlIsen   float64
	VlNT     float64
	VlICMSST float64
}

// RegistroC500 - Nota Fiscal/Conta de Energia Eletrica, Agua, Gas
type RegistroC500 struct {
	IndOper        IndOper
	IndEmit        IndEmit
	CodPart        string
	CodMod         string
	CodSit         CodSit
	Ser            string
	Sub            string
	CodCons        string
	NumDoc         string
	DtDoc          time.Time
	DtES           time.Time
	VlDoc          float64
	VlDesc         float64
	VlForn         float64
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
	CodCta         string
	CodModDocRef   string
	HashDocRef     string
	SerDocRef      string
	NumDocRef      string
	MesDocRef      string
	EnerInjet      *float64
	OutrasDed      *float64
	RegistroC510   []*RegistroC510
	RegistroC590   []*RegistroC590
	RegistroC595   []*RegistroC595
}

// RegistroC510 - Itens do Documento - Energia Eletrica, Agua, Gas
type RegistroC510 struct {
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
	VlBcICMSST float64
	AliqST     float64
	VlICMSST   float64
	IndRec     IndRec
	CodPart    string
	VlPIS      float64
	VlCOFINS   float64
	CodCta     string
}

// RegistroC590 - Registro Analitico do Documento
type RegistroC590 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlRedBC      float64
	CodObs       string
	RegistroC591 []*RegistroC591
}

// RegistroC591 - Informacoes do FCP
type RegistroC591 struct {
	VlFcpOp float64
	VlFcpST float64
}

// RegistroC595 - Observacoes do Lancamento Fiscal
type RegistroC595 struct {
	CodObs       string
	TxtCompl     string
	RegistroC597 []*RegistroC597
}

// RegistroC597 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroC597 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroC600 - Consolidacao Diaria - Energia Eletrica, Agua, Gas
type RegistroC600 struct {
	CodMod       string
	CodMun       string
	Ser          string
	Sub          string
	CodCons      string
	QtdCons      int
	QtdCanc      int
	DtDoc        time.Time
	VlDoc        float64
	VlDesc       float64
	Cons         int
	VlForn       float64
	VlServNT     float64
	VlTerc       float64
	VlDa         float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlPIS        float64
	VlCOFINS     float64
	RegistroC601 []*RegistroC601
	RegistroC610 []*RegistroC610
	RegistroC690 []*RegistroC690
}

// RegistroC601 - Documentos Cancelados da Consolidacao Diaria
type RegistroC601 struct {
	NumDocCanc string
}

// RegistroC610 - Itens do Documento Consolidado
type RegistroC610 struct {
	CodClass   string
	CodItem    string
	Qtd        float64
	Unid       string
	VlItem     float64
	VlDesc     float64
	CstICMS    string
	CFOP       string
	AliqICMS   float64
	VlBcICMS   float64
	VlICMS     float64
	VlBcICMSST float64
	VlICMSST   float64
	VlPIS      float64
	VlCOFINS   float64
	CodCta     string
}

// RegistroC690 - Registro Analitico dos Documentos Consolidados
type RegistroC690 struct {
	CstICMS    string
	CFOP       string
	AliqICMS   float64
	VlOpr      float64
	VlBcICMS   float64
	VlICMS     float64
	VlRedBC    float64
	VlBcICMSST float64
	VlICMSST   float64
	CodObs     string
}

// RegistroC700 - Consolidacao dos Documentos NF/Conta Energia Eletrica (codigo 06)
type RegistroC700 struct {
	CodMod       string
	Ser          string
	NroOrdIni    int
	NroOrdFin    int
	DtDocIni     time.Time
	DtDocFin     time.Time
	NomMest      string
	ChvCodDig    string
	RegistroC790 []*RegistroC790
}

// RegistroC790 - Registro Analitico dos Documentos Consolidados
type RegistroC790 struct {
	CstICMS      string
	CFOP         string
	AliqICMS     float64
	VlOpr        float64
	VlBcICMS     float64
	VlICMS       float64
	VlBcICMSST   float64
	VlICMSST     float64
	VlRedBC      float64
	CodObs       string
	RegistroC791 []*RegistroC791
}

// RegistroC791 - Registro de Informacoes de ICMS ST por UF
type RegistroC791 struct {
	UF         string
	VlBcICMSST float64
	VlICMSST   float64
}

// RegistroC800 - Cupom Fiscal Eletronico - CF-e-SAT (codigo 59)
type RegistroC800 struct {
	CodMod       string
	CodSit       CodSit
	NumCFe       string
	DtDoc        time.Time
	VlCFe        float64
	VlPIS        *float64
	VlCOFINS     *float64
	CNPJCPF      string
	NrSat        string
	ChvCFe       string
	VlDesc       float64
	VlMerc       float64
	VlOutDa      float64
	VlICMS       float64
	VlPISST      *float64
	VlCOFINSST   *float64
	RegistroC810 []*RegistroC810
	RegistroC850 []*RegistroC850
	RegistroC855 []*RegistroC855
}

// RegistroC810 - Itens do Cupom Fiscal Eletronico
type RegistroC810 struct {
	NumItem      string
	CodItem      string
	Qtd          float64
	Unid         string
	VlItem       float64
	CstICMS      string
	CFOP         string
	RegistroC815 *RegistroC815
}

// RegistroC815 - Complemento dos Itens - Ressarcimento
type RegistroC815 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   *float64
	VlUnitICMSOpConv           *float64
	VlUnitICMSOpEstoqueConv    *float64
	VlUnitICMSSTEstoqueConv    *float64
	VlUnitFcpICMSSTEstoqueConv *float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
}

// RegistroC850 - Registro Analitico do CF-e-SAT
type RegistroC850 struct {
	CstICMS  string
	CFOP     string
	AliqICMS float64
	VlOpr    float64
	VlBcICMS float64
	VlICMS   float64
	CodObs   string
}

// RegistroC855 - Observacoes do Lancamento Fiscal
type RegistroC855 struct {
	CodObs       string
	TxtCompl     string
	RegistroC857 []*RegistroC857
}

// RegistroC857 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroC857 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroC860 - Identificacao do Equipamento SAT-CF-e
type RegistroC860 struct {
	CodMod       string
	NrSat        string
	DtDoc        time.Time
	DocIni       string
	DocFin       string
	RegistroC870 []*RegistroC870
	RegistroC890 []*RegistroC890
	RegistroC895 []*RegistroC895
}

// RegistroC870 - Itens do Resumo Diario do CF-e-SAT
type RegistroC870 struct {
	CodItem      string
	Qtd          float64
	Unid         string
	CstICMS      string
	CFOP         string
	RegistroC880 *RegistroC880
}

// RegistroC880 - Complemento dos Itens - Ressarcimento
type RegistroC880 struct {
	CodMotRestCompl            string
	QuantConv                  float64
	Unid                       string
	VlUnitConv                 float64
	VlUnitICMSNaOperacaoConv   *float64
	VlUnitICMSOpConv           *float64
	VlUnitICMSOpEstoqueConv    *float64
	VlUnitICMSSTEstoqueConv    *float64
	VlUnitFcpICMSSTEstoqueConv *float64
	VlUnitICMSSTConvRest       *float64
	VlUnitFcpSTConvRest        *float64
	VlUnitICMSSTConvCompl      *float64
	VlUnitFcpSTConvCompl       *float64
}

// RegistroC890 - Registro Analitico do Resumo Diario do CF-e-SAT
type RegistroC890 struct {
	CstICMS  string
	CFOP     string
	AliqICMS float64
	VlOpr    float64
	VlBcICMS float64
	VlICMS   float64
	CodObs   string
}

// RegistroC895 - Observacoes do Lancamento Fiscal
type RegistroC895 struct {
	CodObs       string
	TxtCompl     string
	RegistroC897 []*RegistroC897
}

// RegistroC897 - Outras Obrigacoes Tributarias, Ajustes e Informacoes
type RegistroC897 struct {
	CodAj        string
	DescrComplAj string
	CodItem      string
	VlBcICMS     float64
	AliqICMS     float64
	VlICMS       float64
	VlOutros     float64
}

// RegistroC990 - Encerramento do Bloco C
type RegistroC990 struct {
	QtdLinC int
}
