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

package rtc

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Estruturas de dados da Reforma Tributaria sobre o Consumo.
// Porte de ACBrDFe.RTC.Classes.pas.
//
// O nome de cada campo e o nome da tag XML correspondente.

// ---------------------------------------------------------------------------
// Pagamento antecipado e documentos referenciados
// ---------------------------------------------------------------------------

// GPagAntecipadoProd e o grupo de pagamento antecipado no nivel do item.
// Porte de TgPagAntecipadoProd.
type GPagAntecipadoProd struct {
	ChDFePagAnt string // chDFePagAnt
	NItemPagAnt int    // nItemPagAnt
}

// RefDFePagAnt e a chave de um documento de pagamento antecipado.
// Porte de TrefDFePagAntCollectionItem.
type RefDFePagAnt struct {
	RefDFeChave string // conteudo de chDFePagAnt ou refNFe
}

// GPagAntecipado e o grupo de pagamento antecipado no nivel da ide.
// Porte de TgPagAntecipado.
type GPagAntecipado struct {
	RefNFe []RefDFePagAnt // chDFePagAnt (ide) ou refNFe (NFe)
}

// DFeReferenciado referencia um documento fiscal e um item dele.
// Porte de TDFeReferenciado.
type DFeReferenciado struct {
	ChaveAcesso string // chaveAcesso
	NItem       int    // nItem
}

// RefDFeAnt e a chave de um documento anterior referenciado.
// Porte de TrefDFeAntCollectionItem.
type RefDFeAnt struct {
	RefDFeAnt string // refDFeAnt
}

// DFeRef e a chave de um documento referenciado na compra governamental da
// NFe. Porte de TDFErefCollectionItem.
type DFeRef struct {
	RefDFeChave string // refDFeAnt
}

// ---------------------------------------------------------------------------
// Compra governamental
// ---------------------------------------------------------------------------

// GCompraGovReduzido e o grupo de compra governamental usado por BPe, CTe,
// NF3e, NFAg, NFCom e NFGas. Porte de TgCompraGovReduzido.
type GCompraGovReduzido struct {
	TpEnteGov TpEnteGov   // tpEnteGov
	PRedutor  float64     // pRedutor
	TpOperGov TpOperGov   // tpOperGov
	RefDFe    []RefDFeAnt // refDFeAnt (0..N)
}

// GCompraGov e o grupo de compra governamental usado pela NFe.
// Porte de TgCompraGov.
type GCompraGov struct {
	TpEnteGov TpEnteGov // tpEnteGov
	PRedutor  float64   // pRedutor
	TpOperGov TpOperGov // tpOperGov
	RefDFeAnt []DFeRef  // refDFeAnt (0..N)
}

// ---------------------------------------------------------------------------
// Blocos elementares de IBS/CBS
// ---------------------------------------------------------------------------

// GALCZFMCBS e o grupo da Zona Franca de Manaus / Areas de Livre Comercio
// para a CBS. Porte de TgALCZFMCBS.
type GALCZFMCBS struct {
	TpALCZFMCBS     TpALCZFMCBS // tpALCZFMCBS
	NProcSuframa    string      // nProcSuframa
	PAliqEfetRegCBS float64     // pAliqEfetRegCBS
	VTribRegCBS     float64     // vTribRegCBS
}

// GDif e o grupo de diferimento. Porte de TgDif.
type GDif struct {
	PDif float64 // pDif
	VDif float64 // vDif
}

// GDevTrib e o grupo de devolucao de tributo. Porte de TgDevTrib.
type GDevTrib struct {
	PDevTrib float64 // pDevTrib
	VDevTrib float64 // vDevTrib
}

// GRed e o grupo de reducao de aliquota. Porte de TgRed.
type GRed struct {
	PRedAliq  float64 // pRedAliq
	PAliqEfet float64 // pAliqEfet
}

// GIBSUFValores e o grupo de valores do IBS da UF. Porte de TgIBSUFValores.
type GIBSUFValores struct {
	PIBSUF   float64  // pIBSUF
	GDif     GDif     // gDif
	GDevTrib GDevTrib // gDevTrib
	GRed     GRed     // gRed
	VIBSUF   float64  // vIBSUF
}

// GIBSMunValores e o grupo de valores do IBS municipal.
// Porte de TgIBSMunValores.
type GIBSMunValores struct {
	PIBSMun  float64  // pIBSMun
	GDif     GDif     // gDif
	GDevTrib GDevTrib // gDevTrib
	GRed     GRed     // gRed
	VIBSMun  float64  // vIBSMun
}

// GCBSValores e o grupo de valores da CBS. Porte de TgCBSValores.
type GCBSValores struct {
	PCBS       float64    // pCBS
	GDif       GDif       // gDif
	GDevTrib   GDevTrib   // gDevTrib
	GRed       GRed       // gRed
	GALCZFMCBS GALCZFMCBS // gALCZFMCBS
	VCBS       float64    // vCBS
}

// GTribRegular e o grupo de tributacao regular. Porte de TgTribRegular.
type GTribRegular struct {
	CSTReg             CSTIBSCBS // CSTReg
	CClassTribReg      string    // cClassTribReg
	PAliqEfetRegIBSUF  float64   // pAliqEfetRegIBSUF
	VTribRegIBSUF      float64   // vTribRegIBSUF
	PAliqEfetRegIBSMun float64   // pAliqEfetRegIBSMun
	VTribRegIBSMun     float64   // vTribRegIBSMun
	PAliqEfetRegCBS    float64   // pAliqEfetRegCBS
	VTribRegCBS        float64   // vTribRegCBS
}

// GTribCompraGov e o grupo de tributacao da compra governamental.
// Porte de TgTribCompraGov.
type GTribCompraGov struct {
	PAliqIBSUF  float64 // pAliqIBSUF
	VTribIBSUF  float64 // vTribIBSUF
	PAliqIBSMun float64 // pAliqIBSMun
	VTribIBSMun float64 // vTribIBSMun
	PAliqCBS    float64 // pAliqCBS
	VTribCBS    float64 // vTribCBS
}

// GIBSCBS e o grupo de apuracao do IBS e da CBS no item.
// Porte de TgIBSCBS.
type GIBSCBS struct {
	VBC            float64        // vBC
	VIBS           float64        // vIBS
	GIBSUF         GIBSUFValores  // gIBSUF
	GIBSMun        GIBSMunValores // gIBSMun
	GCBS           GCBSValores    // gCBS
	GTribRegular   GTribRegular   // gTribRegular
	GTribCompraGov GTribCompraGov // gTribCompraGov
}

// ---------------------------------------------------------------------------
// Monofasia
// ---------------------------------------------------------------------------

// GMonoPadraoIBSQtde e a monofasia padrao do IBS por quantidade.
// Porte de TgMonoPadraoIBSQtde.
type GMonoPadraoIBSQtde struct {
	QBCMono  float64 // qBCMono
	AdRemIBS float64 // adRemIBS
	VIBSMono float64 // vIBSMono
}

// GMonoRetenIBSQtde e a retencao monofasica do IBS por quantidade.
// Porte de TgMonoRetenIBSQtde.
type GMonoRetenIBSQtde struct {
	QBCMonoReten  float64 // qBCMonoReten
	AdRemIBSReten float64 // adRemIBSReten
	VIBSMonoReten float64 // vIBSMonoReten
}

// GpBioDiferencaIBS e a diferenca de biocombustivel no IBS.
// Porte de TgpBioDiferencaIBS.
type GpBioDiferencaIBS struct {
	QBCBioComb    float64 // qBCBioComb
	VIBSDiferenca float64 // vIBSDiferenca
}

// GMonoRetIBS e o IBS monofasico retido anteriormente.
// Porte de TgMonoRetIBS.
type GMonoRetIBS struct {
	VIBSMonoRet float64 // vIBSMonoRet
}

// GIBSMonoAdRem e a monofasia do IBS ad rem. Porte de TgIBSMonoAdRem.
type GIBSMonoAdRem struct {
	GMonoPadrao    GMonoPadraoIBSQtde // gMonoPadrao
	GMonoReten     GMonoRetenIBSQtde  // gMonoReten
	GMonoRet       GMonoRetIBS        // gMonoRet
	GpBioDiferenca GpBioDiferencaIBS  // gpBioDiferenca
}

// GMonoPadraoIBSAliq e a monofasia padrao do IBS por aliquota.
// Porte de TgMonoPadraoIBSAliq.
type GMonoPadraoIBSAliq struct {
	VBCMono      float64 // vBCMono
	PAliqMonoUF  float64 // pAliqMonoUF
	VIBSMonoUF   float64 // vIBSMonoUF
	PAliqMonoMun float64 // pAliqMonoMun
	VIBSMonoMun  float64 // vIBSMonoMun
	VIBSMono     float64 // vIBSMono
}

// GMonoRetenIBSAliq e a retencao monofasica do IBS por aliquota.
// Porte de TgMonoRetenIBSAliq.
type GMonoRetenIBSAliq struct {
	VBCMonoReten   float64 // vBCMonoReten
	PAliqMonoReten float64 // pAliqMonoReten
	VIBSMonoReten  float64 // vIBSMonoReten
}

// GIBSMonoAdValorem e a monofasia do IBS ad valorem.
// Porte de TgIBSMonoAdValorem.
type GIBSMonoAdValorem struct {
	GMonoPadrao    GMonoPadraoIBSAliq // gMonoPadrao
	GMonoReten     GMonoRetenIBSAliq  // gMonoReten
	GMonoRet       GMonoRetIBS        // gMonoRet
	GpBioDiferenca GpBioDiferencaIBS  // gpBioDiferenca -- declarado, nao lido (ver xml_reader.go)
}

// GMonoPadraoCBSQtde e a monofasia padrao da CBS por quantidade.
// Porte de TgMonoPadraoCBSQtde.
type GMonoPadraoCBSQtde struct {
	QBCMono  float64 // qBCMono
	AdRemCBS float64 // adRemCBS
	VCBSMono float64 // vCBSMono
}

// GMonoRetenCBSQtde e a retencao monofasica da CBS por quantidade.
// Porte de TgMonoRetenCBSQtde.
type GMonoRetenCBSQtde struct {
	QBCMonoReten  float64 // qBCMonoReten
	AdRemCBSReten float64 // adRemCBSReten
	VCBSMonoReten float64 // vCBSMonoReten
}

// GpBioDiferencaCBS e a diferenca de biocombustivel na CBS.
// Porte de TgpBioDiferencaCBS.
type GpBioDiferencaCBS struct {
	QBCBioComb    float64 // qBCBioComb
	VCBSDiferenca float64 // vCBSDiferenca
}

// GMonoRetCBS e a CBS monofasica retida anteriormente.
// Porte de TgMonoRetCBS.
type GMonoRetCBS struct {
	VCBSMonoRet float64 // vCBSMonoRet
}

// GCBSMonoAdRem e a monofasia da CBS ad rem. Porte de TgCBSMonoAdRem.
type GCBSMonoAdRem struct {
	GMonoPadrao    GMonoPadraoCBSQtde // gMonoPadrao
	GMonoReten     GMonoRetenCBSQtde  // gMonoReten
	GMonoRet       GMonoRetCBS        // gMonoRet
	GpBioDiferenca GpBioDiferencaCBS  // gpBioDiferenca
}

// GMonoPadraoCBSAliq e a monofasia padrao da CBS por aliquota.
// Porte de TgMonoPadraoCBSAliq.
type GMonoPadraoCBSAliq struct {
	VBCMono      float64 // vBCMono
	PAliqMonoCBS float64 // pAliqMonoCBS
	VCBSMono     float64 // vCBSMono
}

// GMonoRetenCBSAliq e a retencao monofasica da CBS por aliquota.
// Porte de TgMonoRetenCBSAliq.
type GMonoRetenCBSAliq struct {
	VBCMonoReten   float64 // vBCMonoReten
	PAliqMonoReten float64 // pAliqMonoReten
	VCBSMonoReten  float64 // vCBSMonoReten
}

// GCBSMonoAdValorem e a monofasia da CBS ad valorem.
// Porte de TgCBSMonoAdValorem.
type GCBSMonoAdValorem struct {
	GMonoPadrao    GMonoPadraoCBSAliq // gMonoPadrao
	GMonoReten     GMonoRetenCBSAliq  // gMonoReten
	GMonoRet       GMonoRetCBS        // gMonoRet
	GpBioDiferenca GpBioDiferencaCBS  // gpBioDiferenca -- declarado, nao lido (ver xml_reader.go)
}

// GIBSCBSMono agrupa a monofasia de IBS e CBS do item.
// Porte de TgIBSCBSMono.
type GIBSCBSMono struct {
	GIBSMonoAdRem     GIBSMonoAdRem     // gIBSMonoAdRem
	GIBSMonoAdValorem GIBSMonoAdValorem // gIBSMonoAdValorem
	GCBSMonoAdRem     GCBSMonoAdRem     // gCBSMonoAdRem
	GCBSMonoAdValorem GCBSMonoAdValorem // gCBSMonoAdValorem
	VTotIBSMonoItem   float64           // vTotIBSMonoItem
	VTotCBSMonoItem   float64           // vTotCBSMonoItem
}

// ---------------------------------------------------------------------------
// Creditos e ajustes
// ---------------------------------------------------------------------------

// GTransfCred e a transferencia de credito. Porte de TgTransfCred.
type GTransfCred struct {
	VIBS float64 // vIBS
	VCBS float64 // vCBS
}

// GAjusteCompet e o ajuste de competencia. Porte de TgAjusteCompet.
type GAjusteCompet struct {
	CompetApur time.Time // competApur, no XML no formato AAAA-MM
	VIBS       float64   // vIBS
	VCBS       float64   // vCBS
}

// GEstornoCred e o estorno de credito. Porte de TgEstornoCred.
type GEstornoCred struct {
	VIBSEstCred float64 // vIBSEstCred
	VCBSEstCred float64 // vCBSEstCred
}

// GIBSCBSCredPres e o credito presumido de IBS ou de CBS.
// Porte de TgIBSCBSCredPres.
type GIBSCBSCredPres struct {
	PCredPres        float64 // pCredPres
	VCredPres        float64 // vCredPres
	VCredPresCondSus float64 // vCredPresCondSus
}

// GCredPresOper e o credito presumido da operacao.
// Porte de TgCredPresOper.
type GCredPresOper struct {
	VBCCredPres  float64         // vBCCredPres
	CCredPres    CCredPres       // cCredPres
	GIBSCredPres GIBSCBSCredPres // gIBSCredPres
	GCBSCredPres GIBSCBSCredPres // gCBSCredPres
}

// CredPresIBSZFM e o credito presumido do IBS na Zona Franca de Manaus.
// Porte de TCredPresIBSZFM.
type CredPresIBSZFM struct {
	CompetApur       time.Time        // competApur, no XML no formato AAAA-MM
	TpCredPresIBSZFM TpCredPresIBSZFM // tpCredPresIBSZFM
	VCredPresIBSZFM  float64          // vCredPresIBSZFM
}

// IBSCBS e o grupo de IBS/CBS do item. Porte de TIBSCBS.
type IBSCBS struct {
	CST             CSTIBSCBS       // CST
	CClassTrib      string          // cClassTrib
	IndDoacao       pcn.IndicadorEx // indDoacao
	GIBSCBS         GIBSCBS         // gIBSCBS
	GIBSCBSMono     GIBSCBSMono     // gIBSCBSMono
	GTransfCred     GTransfCred     // gTransfCred
	GAjusteCompet   GAjusteCompet   // gAjusteCompet
	GEstornoCred    GEstornoCred    // gEstornoCred
	GCredPresOper   GCredPresOper   // gCredPresOper
	GCredPresIBSZFM CredPresIBSZFM  // gCredPresIBSZFM
}

// ---------------------------------------------------------------------------
// Imposto Seletivo
// ---------------------------------------------------------------------------

// GIS e o grupo do Imposto Seletivo no item. Porte de TgIS.
//
// CSTIS e string, e nao enum: o ACBr deixou assim ate a publicacao de uma
// tabela oficial de CST do IS.
type GIS struct {
	CSTIS        string  // CSTIS
	CClassTribIS string  // cClassTribIS
	VBCIS        float64 // vBCIS
	PIS          float64 // pIS
	AdRemIS      float64 // adRemIS
	UTrib        string  // uTrib
	QTrib        float64 // qTrib
	VIS          float64 // vIS
}

// ISTot e o total do Imposto Seletivo. Porte de TISTot.
type ISTot struct {
	VIS float64 // vIS
}

// ---------------------------------------------------------------------------
// Totais
// ---------------------------------------------------------------------------

// GIBSUFTot e o total do IBS da UF. Porte de TgIBSUFTot.
type GIBSUFTot struct {
	VDif     float64 // vDif
	VDevTrib float64 // vDevTrib
	VIBSUF   float64 // vIBSUF
}

// GIBSMunTot e o total do IBS municipal. Porte de TgIBSMunTot.
type GIBSMunTot struct {
	VDif     float64 // vDif
	VDevTrib float64 // vDevTrib
	VIBSMun  float64 // vIBSMun
}

// GIBS e o total do IBS. Porte de TgIBS.
type GIBS struct {
	GIBSUFTot        GIBSUFTot  // gIBSUF
	GIBSMunTot       GIBSMunTot // gIBSMun
	VIBS             float64    // vIBS
	VCredPres        float64    // vCredPres
	VCredPresCondSus float64    // vCredPresCondSus
}

// GCBS e o total da CBS. Porte de TgCBS.
type GCBS struct {
	VDif             float64 // vDif
	VDevTrib         float64 // vDevTrib
	VCBS             float64 // vCBS
	VCredPres        float64 // vCredPres
	VCredPresCondSus float64 // vCredPresCondSus
}

// GMono e o total da monofasia. Porte de TgMono.
type GMono struct {
	VIBSMono      float64 // vIBSMono
	VCBSMono      float64 // vCBSMono
	VIBSMonoReten float64 // vIBSMonoReten
	VCBSMonoReten float64 // vCBSMonoReten
	VIBSMonoRet   float64 // vIBSMonoRet
	VCBSMonoRet   float64 // vCBSMonoRet
}

// IBSCBSTot e o total de IBS e CBS do documento. Porte de TIBSCBSTot.
type IBSCBSTot struct {
	VBCIBSCBS    float64      // vBCIBSCBS
	GIBS         GIBS         // gIBS
	GCBS         GCBS         // gCBS
	GMono        GMono        // gMono
	GEstornoCred GEstornoCred // gEstornoCred
}

// ---------------------------------------------------------------------------
// Pagamento vinculado
// ---------------------------------------------------------------------------

// Pgto e um pagamento vinculado ao documento.
// Porte de TpgtoCollectionItem.
type Pgto struct {
	NPag        int    // atributo @nPag
	IDTransacao string // atributo @idTransacao
	TpMeioPgto  string // tpMeioPgto
	CNPJReceb   string // CNPJReceb
	CNPJBasePSP string // CNPJBasePSP
}

// PgtoVinc e o grupo de pagamentos vinculados. Porte de TpgtoVinc.
type PgtoVinc struct {
	Pgto []Pgto // pgto (0..N)
}
