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
// Bloco 1 - Outras Informacoes
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// Registro1001 - Abertura do Bloco 1
type Registro1001 struct {
	OpenBlocos
	Registro1010 []*Registro1010
	Registro1100 []*Registro1100
	Registro1200 []*Registro1200
	Registro1250 []*Registro1250
	Registro1300 []*Registro1300
	Registro1350 []*Registro1350
	Registro1390 []*Registro1390
	Registro1400 []*Registro1400
	Registro1500 []*Registro1500
	Registro1600 []*Registro1600
	Registro1601 []*Registro1601
	Registro1700 []*Registro1700
	Registro1800 []*Registro1800
	Registro1900 []*Registro1900
	Registro1960 []*Registro1960
	Registro1970 []*Registro1970
	Registro1980 []*Registro1980
}

// NewRegistro1001 cria um novo Registro1001 com IndMov=1 (sem dados).
func NewRegistro1001() *Registro1001 {
	return &Registro1001{OpenBlocos: OpenBlocos{IndMov: 1}}
}

// Registro1010 - Obrigatoriedade de Registros do Bloco 1
type Registro1010 struct {
	IndExp                  string
	IndCCRF                 string
	IndComb                 string
	IndUsina                string
	IndVA                   string
	IndEE                   string
	IndCart                 string
	IndForm                 string
	IndAer                  string
	IndGIAF1                string
	IndGIAF3                string
	IndGIAF4                string
	IndRestRessarcComplICMS string
}

// Registro1100 - Registro de Informacoes sobre Exportacao
type Registro1100 struct {
	IndDoc       TipoDoctoExport
	NroDE        string
	DtDE         time.Time
	NatExp       Exportacao
	NroRE        string
	DtRE         time.Time
	ChcEmb       string
	DtChc        time.Time
	DtAvb        time.Time
	TpChc        ConhecEmbarque
	Pais         string
	Registro1105 []*Registro1105
}

// Registro1105 - Documentos Fiscais de Exportacao
type Registro1105 struct {
	CodMod       string
	Serie        string
	NumDoc       string
	ChvNFe       string
	DtDoc        time.Time
	CodItem      string
	Registro1110 []*Registro1110
}

// Registro1110 - Operacoes de Exportacao Indireta - Mercadorias de Terceiros
type Registro1110 struct {
	CodPart string
	CodMod  string
	Ser     string
	NumDoc  string
	DtDoc   time.Time
	ChvNFe  string
	NrMemo  string
	Qtd     float64
	Unid    string
}

// Registro1200 - Controle de Creditos Fiscais - ICMS
type Registro1200 struct {
	CodAjApur    string
	SldCred      float64
	CredApr      float64
	CredReceb    float64
	CredUtil     float64
	SldCredFim   float64
	Registro1210 []*Registro1210
}

// Registro1210 - Utilizacao de Creditos Fiscais - ICMS
type Registro1210 struct {
	TipoUtil   string
	NrDoc      string
	VlCredUtil float64
	ChvDOCe    string
}

// Registro1250 - Informacoes Consolidadas de Saldos de Restituicao,
// Ressarcimento e Complementacao do ICMS
type Registro1250 struct {
	VlCreditoICMSOp float64
	VlICMSSTRest    float64
	VlFCPSTRest     float64
	VlICMSSTCompl   float64
	VlFCPSTCompl    float64
	Registro1255    []*Registro1255
}

// Registro1255 - Informacoes Consolidadas de Saldos de Restituicao,
// Ressarcimento e Complementacao do ICMS por Motivo
type Registro1255 struct {
	CodMotRestCompl    string
	VlCreditoICMSOpMot float64
	VlICMSSTRestMot    float64
	VlFCPSTRestMot     float64
	VlICMSSTComplMot   float64
	VlFCPSTComplMot    float64
}

// Registro1300 - Movimentacao Diaria de Combustiveis
type Registro1300 struct {
	CodItem      string
	DtFech       time.Time
	EstqAbert    float64
	VolEntr      float64
	VolDisp      float64
	VolSaidas    float64
	EstqEscr     float64
	ValAjPerda   float64
	ValAjGanho   float64
	FechFisico   float64
	Registro1310 []*Registro1310
}

// Registro1310 - Movimentacao Diaria de Combustiveis por Tanque
type Registro1310 struct {
	NumTanque    string
	EstqAbert    float64
	VolEntr      float64
	VolDisp      float64
	VolSaidas    float64
	EstqEscr     float64
	ValAjPerda   float64
	ValAjGanho   float64
	FechFisico   float64
	CapTanque    int
	Registro1320 []*Registro1320
}

// Registro1320 - Volume de Vendas
type Registro1320 struct {
	NumBico    string
	NrInterv   string
	MotInterv  string
	NomInterv  string
	CnpjInterv string
	CpfInterv  string
	ValFecha   float64
	ValAbert   float64
	VolAferi   float64
	VolVendas  float64
}

// Registro1350 - Bombas
type Registro1350 struct {
	Serie        string
	Fabricante   string
	Modelo       string
	TipoMedicao  Medicao
	Registro1360 []*Registro1360
	Registro1370 []*Registro1370
}

// Registro1360 - Lacres das Bombas
type Registro1360 struct {
	NumLacre    string
	DtAplicacao time.Time
}

// Registro1370 - Bicos da Bomba
type Registro1370 struct {
	NumBico   string
	CodItem   string
	NumTanque string
}

// Registro1390 - Controle de Producao de Usina
type Registro1390 struct {
	CodProd      string
	Registro1391 []*Registro1391
}

// Registro1391 - Producao Diaria da Usina
type Registro1391 struct {
	DtRegistro     time.Time
	QtdMoid        float64
	EstqIni        float64
	QtdProduz      float64
	EntAnidHid     float64
	OutrEntr       float64
	Perda          float64
	Cons           float64
	SaiAniHid      float64
	Saidas         float64
	EstqFin        float64
	EstqIniMel     float64
	ProdDiaMel     float64
	UtilMel        float64
	ProdAlcMel     float64
	Obs            string
	CodItem        string
	TpResiduo      int
	QtdResiduo     float64
	QtdResiduoDDG  float64
	QtdResiduoWDG  float64
	QtdResiduoCana float64
}

// Registro1400 - Informacao sobre Valores Agregados
type Registro1400 struct {
	CodItemIPM string
	CodItem    string
	Mun        string
	Valor      float64
}

// Registro1500 - Nota Fiscal/Conta de Energia Eletrica (codigo 06) -
// Operacoes Interestaduais
type Registro1500 struct {
	IndOper        string
	IndEmit        string
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
	Registro1510   []*Registro1510
}

// Registro1510 - Itens do Documento Nota Fiscal/Conta de Energia Eletrica
type Registro1510 struct {
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

// Registro1600 - Total das Operacoes com Cartao de Credito e/ou Debito,
// Loja (Private Label) e Demais Instrumentos de Pagamento Eletronico
type Registro1600 struct {
	CodPart    string
	TotCredito float64
	TotDebito  float64
}

// Registro1601 - Complemento da Operacao - Instrumentos de Pagamento Eletronico
type Registro1601 struct {
	CodPartIP string
	CodPartIT string
	TotVS     float64
	TotISS    float64
	TotOutros float64
}

// Registro1700 - Documentos Fiscais Utilizados
type Registro1700 struct {
	CodDisp      Dispositivo
	CodMod       string
	Ser          string
	Sub          string
	NumDocIni    string
	NumDocFin    string
	NumAut       string
	Registro1710 []*Registro1710
}

// Registro1710 - Documentos Fiscais Cancelados/Inutilizados
type Registro1710 struct {
	NumDocIni string
	NumDocFin string
}

// Registro1800 - Credito Presumido sobre Prestacoes de Servicos de Transporte
type Registro1800 struct {
	VlCarga      float64
	VlPass       float64
	VlFat        float64
	IndRat       float64
	VlICMSAnt    float64
	VlBcICMS     float64
	VlICMSApur   float64
	VlBcICMSApur float64
	VlDif        float64
}

// Registro1900 - Indicador de sub-apuracao do ICMS
type Registro1900 struct {
	IndApurICMS       string
	DescrComplOutApur string
	Registro1910      []*Registro1910
}

// Registro1910 - Periodo da sub-apuracao do ICMS
type Registro1910 struct {
	DtIni        time.Time
	DtFin        time.Time
	Registro1920 []*Registro1920
}

// Registro1920 - Sub-apuracao do ICMS
type Registro1920 struct {
	VlTotTransfDebitosOA  float64
	VlTotAjDebitosOA      float64
	VlEstornosCredOA      float64
	VlTotTransfCreditosOA float64
	VlTotAjCreditosOA     float64
	VlEstornosDebOA       float64
	VlSldCredorAntOA      float64
	VlSldApuradoOA        float64
	VlTotDed              float64
	VlICMSRecolherOA      float64
	VlSldCredorTranspOA   float64
	DebEspOA              float64
	Registro1921          []*Registro1921
	Registro1925          []*Registro1925
	Registro1926          []*Registro1926
}

// Registro1921 - Ajuste/Beneficio/Incentivo da sub-apuracao do ICMS
type Registro1921 struct {
	CodAjApur    string
	DescrComplAj string
	VlAjApur     float64
	Registro1922 []*Registro1922
	Registro1923 []*Registro1923
}

// Registro1922 - Informacoes Adicionais dos Ajustes da sub-apuracao do ICMS
type Registro1922 struct {
	NumDA    string
	NumProc  string
	IndProc  string
	Proc     string
	TxtCompl string
}

// Registro1923 - Informacoes Adicionais dos Ajustes da sub-apuracao do ICMS -
// Identificacao dos documentos fiscais
type Registro1923 struct {
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

// Registro1925 - Informacoes Adicionais da sub-apuracao - Valores Declaratorios
type Registro1925 struct {
	CodInfAdic   string
	VlInfAdic    float64
	DescrComplAj string
}

// Registro1926 - Obrigacoes do ICMS a recolher - Operacoes referentes a
// sub-apuracao
type Registro1926 struct {
	CodOR    string
	VlOR     float64
	DtVcto   time.Time
	CodRec   string
	NumProc  string
	IndProc  string
	Proc     string
	TxtCompl string
	MesRef   string
}

// Registro1960 - GIAF1
type Registro1960 struct {
	IndAp string
	G1_01 float64
	G1_02 float64
	G1_03 float64
	G1_04 float64
	G1_05 float64
	G1_06 float64
	G1_07 float64
	G1_08 float64
	G1_09 float64
	G1_10 float64
	G1_11 float64
}

// Registro1970 - GIAF3
type Registro1970 struct {
	IndAp        string
	G3_01        float64
	G3_02        float64
	G3_03        float64
	G3_04        float64
	G3_05        float64
	G3_06        float64
	G3_07        float64
	G3_T         float64
	G3_08        float64
	G3_09        float64
	Registro1975 []*Registro1975
}

// Registro1975 - GIAF3 - Por aliquota do imposto de base
type Registro1975 struct {
	AliqImpBase float64
	G3_10       float64
	G3_11       float64
	G3_12       float64
}

// Registro1980 - GIAF4
type Registro1980 struct {
	IndAp string
	G4_01 float64
	G4_02 float64
	G4_03 float64
	G4_04 float64
	G4_05 float64
	G4_06 float64
	G4_07 float64
	G4_08 float64
	G4_09 float64
	G4_10 float64
	G4_11 float64
	G4_12 float64
}

// Registro1990 - Encerramento do Bloco 1
type Registro1990 struct {
	QtdLin1 int
}
