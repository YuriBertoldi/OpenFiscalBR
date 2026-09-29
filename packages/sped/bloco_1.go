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

// NewRegistro1001 cria um novo Registro1001 com IndDad=1 (sem dados).
func NewRegistro1001() *Registro1001 {
	return &Registro1001{OpenBlocos: OpenBlocos{IndDad: 1}}
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
	ChvNFe       string
	DtEmb        time.Time
	CodMunOrig   string
	Registro1105 []*Registro1105
}

// Registro1105 - Documentos Fiscais de Exportacao
type Registro1105 struct {
	CodMod       string
	Ser          string
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
	NroMemo string
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
	VlCredICMSOp    float64
	VlICMSST        float64
	VlFCPST         float64
	VlCredICMSPfcp  float64
	VlGlosaOp       float64
	VlGlosaST       float64
	VlGlosaFcp      float64
	VlCredICMSCompl float64
	Registro1255    []*Registro1255
}

// Registro1255 - Informacoes Consolidadas de Saldos de Restituicao,
// Ressarcimento e Complementacao do ICMS por Motivo
type Registro1255 struct {
	CodMot          MotivoRessarcimento
	VlCredICMSOp    float64
	VlICMSST        float64
	VlFCPST         float64
	VlCredICMSPfcp  float64
	VlGlosaOp       float64
	VlGlosaST       float64
	VlGlosaFcp      float64
	VlCredICMSCompl float64
}

// Registro1300 - Movimentacao Diaria de Combustiveis
type Registro1300 struct {
	CodItem      string
	DtFech       time.Time
	EstqAbert    float64
	VolEntr      float64
	VolDisp      float64
	VolSaidas    float64
	EstqFech     float64
	Registro1310 []*Registro1310
}

// Registro1310 - Movimentacao Diaria de Combustiveis por Tanque
type Registro1310 struct {
	NumTanque    string
	EstqAbert    float64
	VolEntr      float64
	VolDisp      float64
	VolSaidas    float64
	EstqFech     float64
	Registro1320 []*Registro1320
}

// Registro1320 - Volume de Vendas
type Registro1320 struct {
	NumBicoBba string
	NrInterv   int
	MotInterv  string
	NomInterv  string
	CnpjInterv string
	CpfInterv  string
	VlVendas   float64
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
	NumLacre string
	DtAplic  time.Time
}

// Registro1370 - Bicos da Bomba
type Registro1370 struct {
	NumBico string
	CodItem string
	CodComb string
}

// Registro1390 - Controle de Producao de Usina
type Registro1390 struct {
	CodProd      TipoResiduo
	Registro1391 []*Registro1391
}

// Registro1391 - Producao Diaria da Usina
type Registro1391 struct {
	DtRegistro   time.Time
	QtdMoidaDia  float64
	EstqIni      float64
	QtdProduzDia float64
	IndTipoEstq  string
	QtdSaidaDia  float64
	EstqFin      float64
	QtdPerdaDia  float64
	QtdEntDia    float64
	QtdDevolDia  float64
}

// Registro1400 - Informacao sobre Valores Agregados
type Registro1400 struct {
	CodItem string
	MunOrig string
	VlItem  float64
	IndApur string
}

// Registro1500 - Nota Fiscal/Conta de Energia Eletrica (codigo 06) -
// Operacoes Interestaduais
type Registro1500 struct {
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

// Registro1800 - DCTA - Demonstrativo de Credito do ICMS sobre Transporte Aereo
type Registro1800 struct {
	VlCargaTrib float64
	VlCargaNT   float64
	VlReceitas  float64
	IndRatio    float64
	VlCREDICMS  float64
}

// Registro1900 - Indicador de Sub-Apuracao do ICMS
type Registro1900 struct {
	IndApurICMS  string
	DescrComplAj string
	Registro1910 []*Registro1910
}

// Registro1910 - Periodo da Sub-Apuracao do ICMS
type Registro1910 struct {
	DtIni        time.Time
	DtFin        time.Time
	Registro1920 []*Registro1920
}

// Registro1920 - Sub-Apuracao do ICMS
type Registro1920 struct {
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
	Registro1921      []*Registro1921
	Registro1925      []*Registro1925
	Registro1926      []*Registro1926
}

// Registro1921 - Ajuste/Beneficio/Incentivo da Sub-Apuracao do ICMS
type Registro1921 struct {
	CodAjApur    string
	DescrComplAj string
	VlAjApur     float64
}

// Registro1925 - Informacoes Adicionais da Sub-Apuracao do ICMS -
// Valores Declaratorios
type Registro1925 struct {
	CodInfAdic   string
	VlInfAdic    float64
	DescrComplAj string
}

// Registro1926 - Obrigacoes do ICMS Recolhido ou a Recolher -
// Sub-Apuracao
type Registro1926 struct {
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

// Registro1960 - GIAF 1 - Guia de Informacao e Apuracao do ICMS -
// ICMS Diferido
type Registro1960 struct {
	IndAp string
	G1_01 string
	G1_02 string
	G1_03 string
	G1_04 string
	G1_05 float64
	G1_06 float64
	G1_07 float64
	G1_08 float64
	G1_09 float64
	G1_10 float64
	G1_11 float64
}

// Registro1970 - GIAF 3 - Guia de Informacao e Apuracao do ICMS -
// ICMS Incentivado (Pernambuco)
type Registro1970 struct {
	IndAp        string
	G3_01        string
	G3_02        string
	G3_03        string
	G3_04        string
	G3_05        string
	G3_T         float64
	G3_08        float64
	G3_09        float64
	Registro1975 []*Registro1975
}

// Registro1975 - GIAF 3 - Guia de Informacao e Apuracao do ICMS -
// ICMS Incentivado - Detalhamento
type Registro1975 struct {
	AliqImpBase float64
	G3_05       float64
	G3_06       float64
	G3_07       float64
}

// Registro1980 - GIAF 4 - Guia de Informacao e Apuracao do ICMS -
// ICMS a Recuperar PRODEPE
type Registro1980 struct {
	IndAp string
	G4_01 string
	G4_02 string
	G4_03 string
	G4_04 string
	G4_05 float64
	G4_06 float64
	G4_07 float64
}

// Registro1990 - Encerramento do Bloco 1
type Registro1990 struct {
	QtdLin1 int
}
