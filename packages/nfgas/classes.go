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

package nfgas

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Modelo de dados da NFGas. Porte de ACBrNFGas.Classes.pas.
//
// O comentario de cada campo traz o nome EXATO da tag XML, que e o unico
// contrato real com o schema da SEFAZ -- os structs nao levam tag `xml:`
// porque a leitura e imperativa (ver xml_reader.go).
//
// Grupos 1:1 sao structs por valor; grupos 0..N sao slices.

// NFGas e o documento fiscal. Porte de TNFGas.
type NFGas struct {
	InfNFGas     InfNFGas      // infNFGas
	Ide          Ide           // ide
	Emit         Emit          // emit
	Dest         Dest          // dest
	Instalacao   Instalacao    // instalacao
	GVolContrat  []GVolContrat // gVolContrat (0..20)
	GSub         GSub          // gSub
	GMed         []GMed        // gMed (0..99)
	Det          []Det         // det (1..990)
	Total        Total         // total
	PgtoVinc     rtc.PgtoVinc  // pgtoVinc
	GFat         GFat          // gFat
	GAgencia     GAgencia      // gAgencia
	AutXML       []AutXML      // autXML (0..N)
	InfAdic      InfAdic       // infAdic
	InfRespTec   InfRespTec    // gRespTec -- a tag NAO se chama infRespTec
	InfNFGasSupl InfNFGasSupl  // infNFGasSupl
	ProcNFGas    pcn.ProcDFe   // protNFGas/infProt, quando o XML e nfgasProc
	Signature    pcn.Signature // Signature
}

// InfNFGas carrega os atributos do elemento infNFGas.
// Porte de TinfNFGas.
type InfNFGas struct {
	ID     string  // atributo Id, com o literal "NFGas" na frente da chave
	Versao float64 // atributo versao
}

// Ide e o grupo de identificacao do documento. Porte de TIde.
type Ide struct {
	CUF          int                    // cUF
	TpAmb        pcn.TipoAmbiente       // tpAmb
	Modelo       int                    // mod -- sempre 76
	Serie        int                    // serie
	NNF          int                    // nNF
	CNF          int                    // cNF
	CDV          int                    // cDV
	DhEmi        time.Time              // dhEmi
	TpEmis       pcn.TipoEmissao        // tpEmis
	NSiteAutoriz SiteAutorizador        // nSiteAutoriz
	CMunFG       int                    // cMunFG
	FinNFGas     FinalidadeNFGas        // finNFGas
	TpFat        TpFat                  // tpFat
	VerProc      string                 // verProc
	DhCont       time.Time              // dhCont
	XJust        string                 // xJust
	GCompraGov   rtc.GCompraGovReduzido // gCompraGov
	TpPagAnt     pcn.TpPagAnt           // tpPagAnt
}

// Endereco e o endereco usado por emit, dest e enderCorresp.
// Porte de TEndereco.
//
// CPais e XPais existem na classe do ACBr e sao lidos de enderEmit e
// enderDest, mas NAO existem no tipo TEndeEmi do XSD nem sao lidos de
// enderCorresp -- ver xml_reader.go.
type Endereco struct {
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	CEP     int    // CEP
	UF      string // UF
	Fone    string // fone
	Email   string // email
	CPais   int    // cPais
	XPais   string // xPais
}

// Emit e o grupo do emitente. Porte de TEmit.
type Emit struct {
	CNPJ      string   // CNPJ (ou CPF, via ObterCNPJCPF)
	IE        string   // IE
	XNome     string   // xNome
	XFant     string   // xFant
	EnderEmit Endereco // enderEmit
	ISUFEmit  string   // ISUFEmit
}

// Dest e o grupo do destinatario. Porte de TDest.
//
// IDEstrangeiro guarda tanto idOutros quanto idEstrangeiro: no Delphi as
// duas properties leem e escrevem o MESMO campo FidEstrangeiro, e o leitor
// preenche por precedencia (CNPJ/CPF, depois idOutros, depois
// idEstrangeiro). TagIDOrigem registra de qual tag o valor veio -- isso o
// ACBr nao guarda.
type Dest struct {
	XNome          string    // xNome
	CNPJCPF        string    // CNPJ ou CPF
	IDEstrangeiro  string    // idOutros / idEstrangeiro (mesmo campo no ACBr)
	TagIDOrigem    string    // "idOutros" ou "idEstrangeiro"; vazio se nenhum
	IndIEDest      IndIEDest // indIEDest -- so no .ini, NAO lido do XML
	IE             string    // IE
	IM             string    // IM
	CNIS           string    // cNIS
	NB             string    // NB
	XNomeAdicional string    // xNomeAdicional
	EnderDest      Endereco  // enderDest
}

// IDOutros devolve o identificador alternativo. Espelha a property
// TDest.idOutros do Delphi, que compartilha campo com idEstrangeiro.
func (d *Dest) IDOutros() string { return d.IDEstrangeiro }

// SetIDOutros grava o identificador alternativo.
func (d *Dest) SetIDOutros(v string) { d.IDEstrangeiro = v }

// Instalacao e o grupo da instalacao consumidora. Porte de TInstalacao.
type Instalacao struct {
	IDInstalacao      string       // idInstalacao
	IDCodCliente      string       // idCodCliente
	TpInstalacao      TpInstalacao // tpInstalacao
	NContrato         string       // nContrato
	TpClasse          TpClasse     // tpClasse
	XClasse           string       // xClasse
	LatGPS            string       // latGPS
	LongGPS           string       // longGPS
	CodRoteiroLeitura string       // codRoteiroLeitura
}

// GSub e o grupo de substituicao de documento. Porte de TgSub.
type GSub struct {
	ChNFGas string // chNFGas
	GNF     GNF    // gNF
	MotSub  MotSub // motSub
}

// GNF referencia a nota fiscal substituida. Porte de TgNF.
//
// Serie aqui e string, diferente de Ide.Serie, que e int. E assim no ACBr.
type GNF struct {
	CNPJ       string    // CNPJ
	Serie      string    // serie -- string, nao int
	NNF        int       // nNF
	CompetEmis time.Time // CompetEmis, no XML no formato AAAAMM
	CompetApur time.Time // CompetApur, no XML no formato AAAAMM
	Hash115    string    // hash115
}

// GVolContrat e um volume contratado. Porte de TgVolContratCollectionItem.
type GVolContrat struct {
	NContrat     int        // atributo @nContrat
	TpVolContrat VolContrat // tpVolContrat
	QUnidContrat float64    // qUnidContrat
}

// GMed e uma medicao de equipamento. Porte de TgMedCollectionItem.
type GMed struct {
	NMed      int       // atributo @nMed
	IDEqp     string    // idEqp
	DMedAnt   time.Time // dMedAnt
	VMedAnt   float64   // vMedAnt
	DMedAtu   time.Time // dMedAtu
	VMedAtu   float64   // vMedAtu
	TpEqp     TpEqp     // tpEqp
	TpMedidor TpMedidor // tpMedidor
}

// GTarif e uma faixa tarifaria do item.
// Porte de TgTarifCollectionItem.
type GTarif struct {
	DIniTarif   time.Time   // dIniTarif
	DFimTarif   time.Time   // dFimTarif
	NAto        string      // nAto
	AnoAto      int         // anoAto
	TpFaixaCons TpFaixaCons // tpFaixaCons
	VTarifAplic float64     // vTarifAplic
}

// GMedida e a medida registrada. Porte de TgMedida.
type GMedida struct {
	UMed UMed    // uMed
	VMed float64 // vMed
}

// GMedicao e o grupo de medicao do produto. Porte de TgMedicao.
type GMedicao struct {
	NMed            int             // nMed
	NContrat        int             // nContrat
	GMedida         GMedida         // gMedida
	TpMotNaoLeitura TpMotNaoLeitura // tpMotNaoLeitura
	XMotNaoLeitura  string          // xMotNaoLeitura
}

// Prod e o produto do item. Porte de TProd.
//
// DIVERGENCIA DE TIPO em dois campos, ambas contra o ACBr e a favor do XSD:
//
//   - CClass e string. O XSD declara xs:string com pattern [0-9]{7}, e o
//     leitor do ACBr le como tcStr -- mas a classe declara Integer, o que
//     descarta os zeros a esquerda de um codigo de 7 posicoes.
//   - QFaturada e float64. O XSD declara TDec_1100_1104 (ate 4 casas
//     decimais) e o leitor le com tcDe4 -- mas a classe declara Integer, o
//     que ARREDONDA a quantidade faturada de gas.
//
// Nos dois casos o comportamento do ACBr e deterministico, mas perde dado
// do documento; como o objetivo aqui e leitura fiel do XML, o tipo segue o
// schema. Ha teste cobrindo os dois.
type Prod struct {
	IndOrigemQtd   IndOrigemQtd           // indOrigemQtd
	GMedicao       GMedicao               // gMedicao
	CProd          string                 // cProd
	XProd          string                 // xProd
	CClass         string                 // cClass -- ver nota acima
	CFOP           int                    // CFOP
	UMed           UMedItem               // uMed
	QFaturada      float64                // qFaturada -- ver nota acima
	VItem          float64                // vItem
	FatorPCS       float64                // fatorPCS
	FatorPTZ       float64                // fatorPTZ
	FatorP         float64                // fatorP
	FatorT         float64                // fatorT
	VProd          float64                // vProd
	IndDevolucao   pcn.Indicador          // indDevolucao
	GPagAntecipado rtc.GPagAntecipadoProd // gPagAntecipado
}

// ICMS e o grupo de ICMS do item. Porte de TICMS.
//
// O leitor ACHATA aqui os grupos ICMS00..ICMS90: qualquer que seja o grupo
// presente no XML, os campos caem todos nesta struct, e o CST e quem
// diferencia. Ver lerICMS em xml_reader.go.
type ICMS struct {
	CST             pcn.CSTIcms            // CST
	VBC             float64                // vBC
	PICMS           float64                // pICMS
	VICMS           float64                // vICMS
	VBCFCP          float64                // vBCFCP
	PFCP            float64                // pFCP
	VFCP            float64                // vFCP
	VBCST           float64                // vBCST
	PICMSST         float64                // pICMSST
	VICMSST         float64                // vICMSST
	PFCPST          float64                // pFCPST
	VFCPST          float64                // vFCPST
	PRedBC          float64                // pRedBC
	VICMSDeson      float64                // vICMSDeson
	CBenef          string                 // cBenef
	IndSemCST       pcn.Indicador          // indSemCST -- lido do no imposto, nao do ICMSxx
	VBCSTRet        float64                // vBCSTRet
	PICMSSTRet      float64                // pICMSSTRet
	VICMSSubstituto float64                // vICMSSubstituto
	VICMSSTRet      float64                // vICMSSTRet
	VBCFCPSTRet     float64                // vBCFCPSTRet
	PFCPSTRet       float64                // pFCPSTRet
	VFCPSTRet       float64                // vFCPSTRet
	PRedBCEfet      float64                // pRedBCEfet
	VBCEfet         float64                // vBCEfet
	PICMSEfet       float64                // pICMSEfet
	VICMSEfet       float64                // vICMSEfet
	ModBC           DeterminacaoBaseIcms   // modBC
	ModBCST         DeterminacaoBaseIcmsST // modBCST
	MotDesICMS      MotivoDesoneracaoICMS  // motDesICMS
	PMVAST          float64                // pMVAST
	PRedBCST        float64                // pRedBCST
	VBCFCPST        float64                // vBCFCPST
	IndDeduzDeson   pcn.IndicadorEx        // indDeduzDeson
}

// PIS e o grupo de PIS do item. Porte de TPIS.
type PIS struct {
	CST  pcn.CSTPis // CST
	VBC  float64    // vBC
	PPIS float64    // pPIS
	VPIS float64    // vPIS
}

// COFINS e o grupo de COFINS do item. Porte de TCOFINS.
type COFINS struct {
	CST     pcn.CSTCofins // CST
	VBC     float64       // vBC
	PCOFINS float64       // pCOFINS
	VCOFINS float64       // vCOFINS
}

// RetTrib e o grupo de tributos retidos do item. Porte de TretTrib.
//
// Atencao a grafia da tag de COFINS: no XML e vRetCofins, nao vRetCOFINS.
type RetTrib struct {
	VRetPIS    float64 // vRetPIS
	VRetCOFINS float64 // vRetCofins
	VRetCSLL   float64 // vRetCSLL
	VBCIRRF    float64 // vBCIRRF
	VIRRF      float64 // vIRRF
}

// TxReg e a taxa de regulacao do item. Porte de TTxReg.
type TxReg struct {
	VBC   float64 // vBC
	PTaxa float64 // pTaxa
	VTaxa float64 // vTaxa
}

// Imposto agrupa os tributos do item. Porte de TImposto.
type Imposto struct {
	Orig      pcn.OrigemMercadoria // orig
	ICMS      ICMS                 // ICMS00..ICMS90, achatados
	IndSemCST pcn.Indicador        // indSemCST
	PIS       PIS                  // PIS
	COFINS    COFINS               // COFINS
	RetTrib   RetTrib              // retTrib
	TxReg     TxReg                // TxReg
	IBSCBS    rtc.IBSCBS           // IBSCBS
}

// GProc referencia um processo administrativo ou judicial.
// Porte de TgProcCollectionItem.
type GProc struct {
	TpProc    TpProc // tpProc
	NProcesso string // nProcesso
}

// GProcRef e o grupo de item referente a processo. Porte de TgProcRef.
//
// QFaturada aqui e Double no proprio ACBr, diferente de Prod.QFaturada.
type GProcRef struct {
	VItem        float64       // vItem
	QFaturada    float64       // qFaturada
	VProd        float64       // vProd
	IndDevolucao pcn.Indicador // indDevolucao
	GProc        []GProc       // gProc (0..N)
}

// GNormal e o grupo de item normal. Porte de TgNormal.
type GNormal struct {
	GTarif    []GTarif // gTarif (0..6)
	Prod      Prod     // prod
	Imposto   Imposto  // imposto
	GProcRef  GProcRef // gProcRef
	InfAdProd string   // infAdProd
}

// GAgregadora e o grupo de item agregador. Porte de TgAgregadora.
type GAgregadora struct {
	CClass  string  // cClass
	VTotDFe float64 // vTotDFe
}

// Det e um item do documento. Porte de TDetCollectionItem.
type Det struct {
	NItem       int         // atributo @nItem
	ChNFGasAnt  string      // atributo @chNFGasAnt
	NItemAnt    int         // atributo @nItemAnt
	GNormal     GNormal     // gNormal
	GAgregadora GAgregadora // gAgregadora
}

// Total e o grupo de totais do documento. Porte de TTotal.
//
// Os campos de ICMSTot e de vRetTribTot sao ACHATADOS aqui, como no ACBr.
type Total struct {
	VProd      float64       // vProd
	VBC        float64       // ICMSTot/vBC
	VICMS      float64       // ICMSTot/vICMS
	VICMSDeson float64       // ICMSTot/vICMSDeson
	VFCP       float64       // ICMSTot/vFCP
	VBCST      float64       // ICMSTot/vBCST
	VST        float64       // ICMSTot/vST
	VFCPST     float64       // ICMSTot/vFCPST
	VCOFINS    float64       // vCOFINS
	VPIS       float64       // vPIS
	VNF        float64       // vNF
	VRetPIS    float64       // vRetTribTot/vRetPIS
	VRetCOFINS float64       // vRetTribTot/vRetCofins
	VRetCSLL   float64       // vRetTribTot/vRetCSLL
	VIRRF      float64       // vRetTribTot/vIRRF
	VTxReg     float64       // vTxReg
	IBSCBSTot  rtc.IBSCBSTot // IBSCBSTot
	VTotDFe    float64       // vTotDFe
}

// GPIX e o grupo de cobranca por PIX. Porte de TgPIX.
type GPIX struct {
	URLQRCodePIX string // urlQRCodePIX
}

// GFat e o grupo de faturamento. Porte de TgFat.
type GFat struct {
	CompetFat    time.Time // CompetFat, no XML no formato AAAAMM
	DVencFat     time.Time // dVencFat
	CodBarras    string    // codBarras
	CodDebAuto   string    // codDebAuto
	CodBanco     string    // codBanco
	CodAgencia   string    // codAgencia
	EnderCorresp Endereco  // enderCorresp
	GPIX         GPIX      // gPIX
	DApresFat    time.Time // dApresFat
	DProxLeitura time.Time // dProxLeitura
	NFat         string    // nFat
	InfAdFat     string    // infAdFat
}

// GCons e um mes do historico de consumo.
// Porte de TgConsCollectionItem.
type GCons struct {
	CompetFat time.Time // CompetFat, no XML no formato AAAAMM
	UMed      UMed      // uMed
	QtdDias   int       // qtdDias
	MedDiaria float64   // medDiaria
	Consumo   float64   // consumo
	VFat      float64   // vFat
}

// GHistCons e um bloco do historico de consumo.
// Porte de TgHistConsCollectionItem.
type GHistCons struct {
	XHistorico string  // xHistorico
	MedMensal  float64 // medMensal
	GCons      []GCons // gCons (0..N)
}

// GAgencia e o grupo da agencia de atendimento. Porte de TgAgencia.
type GAgencia struct {
	NomeAgenciaAtend  string      // nomeAgenciaAtend
	EnderAgenciaAtend string      // enderAgenciaAtend
	SitioAgenciaAtend string      // sitioAgenciaAtend
	InfAdReg          string      // infAdReg
	GHistCons         []GHistCons // gHistCons (0..N)
}

// AutXML e um autorizado a baixar o XML.
// Porte de TautXMLCollectionItem.
type AutXML struct {
	CNPJCPF string // CNPJ ou CPF
}

// InfAdic e o grupo de informacoes adicionais. Porte de TInfAdic.
//
// InfCpl e slice porque o leiaute admite ate 5 ocorrencias. O leitor do
// ACBr le apenas a PRIMEIRA -- ha um TODO aberto no proprio fonte --, e o
// porte reproduz isso; a forma de slice fica pronta para quando a leitura
// das 5 for corrigida, sem quebrar quem ja consome.
type InfAdic struct {
	InfAdFisco string   // infAdFisco
	InfCpl     []string // infCpl (0..5, apenas a primeira e lida)
}

// InfRespTec e o grupo do responsavel tecnico. Porte de TinfRespTec.
//
// A TAG no XML e gRespTec; infRespTec e o nome do campo e da secao no .ini.
type InfRespTec struct {
	CNPJ     string // CNPJ
	XContato string // xContato
	Email    string // email
	Fone     string // fone
	IDCSRT   int    // idCSRT
	HashCSRT string // hashCSRT
}

// InfNFGasSupl carrega as informacoes suplementares do documento.
// Porte de TinfNFGasSupl.
type InfNFGasSupl struct {
	QrCodNFGas string // qrCodNFGas
}

// ---------------------------------------------------------------------------
// Acesso de conveniencia
// ---------------------------------------------------------------------------

// ChaveAcesso devolve a chave de 44 posicoes, sem o literal "NFGas" que
// precede o atributo Id. Porte de TNotaFiscal.NumID.
func (n *NFGas) ChaveAcesso() string {
	if n == nil {
		return ""
	}
	return pcn.RemoverLiteralChave(n.InfNFGas.ID)
}

// Confirmada informa se o documento foi confirmado pela SEFAZ, a partir do
// cStat do protocolo. Porte de TACBrNFGas.CstatConfirmada.
func (n *NFGas) Confirmada() bool {
	return n != nil && cStatConfirmada[n.ProcNFGas.CStat]
}

// Processada informa se o documento foi processado pela SEFAZ.
// Porte de TACBrNFGas.CstatProcessado.
func (n *NFGas) Processada() bool {
	return n != nil && cStatProcessado[n.ProcNFGas.CStat]
}

// Cancelada informa se o documento foi cancelado.
// Porte de TACBrNFGas.CstatCancelada.
func (n *NFGas) Cancelada() bool {
	return n != nil && cStatCancelada[n.ProcNFGas.CStat]
}
