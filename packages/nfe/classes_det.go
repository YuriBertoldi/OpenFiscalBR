// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Item da nota: produto, impostos e grupos por tipo de mercadoria.
//
// Porte de Base/ACBrNFe.Classes.pas:478-1480 -- a MESMA unit de classes.go,
// separada em dois arquivos porque o subgrupo det sozinho tem 29 classes e
// dez grupos de imposto. As convencoes do topo de classes.go valem aqui:
// nome de campo igual a tag do leiaute, colecao vira slice, e o `default` da
// property do Delphi NAO inicializa nada.
//
// E a metade do modelo onde o risco de fidelidade se concentra.

// ===========================================================================
// Declaracao de Importacao, exportacao e rastreabilidade
// ===========================================================================

// Adi e uma adicao da Declaracao de Importacao. Porte de TAdiCollectionItem.
type Adi struct {
	NAdicao     int     // nAdicao
	NSeqAdi     int     // nSeqAdi
	CFabricante string  // cFabricante
	VDescDI     float64 // vDescDI
	NDraw       string  // nDraw
}

// DI e uma Declaracao de Importacao do item. Porte de TDICollectionItem.
type DI struct {
	NDi          string         // nDI
	DDi          time.Time      // dDI
	XLocDesemb   string         // xLocDesemb
	UFDesemb     string         // UFDesemb
	DDesemb      time.Time      // dDesemb
	TpViaTransp  TipoViaTransp  // tpViaTransp
	VAFRMM       float64        // vAFRMM
	TpIntermedio TipoIntermedio // tpIntermedio
	CNPJ         string         // CNPJ
	UFTerceiro   string         // UFTerceiro
	CExportador  string         // cExportador
	Adi          []Adi          // adi
}

// DetExport e o detalhamento de exportacao do item. Porte de
// TdetExportCollectionItem.
type DetExport struct {
	NDraw   string  // nDraw
	NRE     string  // nRE
	ChNFe   string  // chNFe
	QExport float64 // qExport
}

// Rastro e o grupo de rastreabilidade do item. Porte de
// TRastroCollectionItem.
type Rastro struct {
	NLote  string    // nLote
	QLote  float64   // qLote
	DFab   time.Time // dFab
	DVal   time.Time // dVal
	CAgreg string    // cAgreg
}

// ===========================================================================
// Mercadorias reguladas: veiculo, medicamento, arma e combustivel
// ===========================================================================

// VeicProd e o grupo de veiculo novo. Porte de TveicProd.
//
// Varios campos numericos sao string aqui (Pot, Cilin, PesoL, PesoB, CMT,
// Dist): e assim no ACBr, e o XSD os declara como string com pattern -- tipar
// como numero perderia o zero a esquerda.
//
// A property somente-leitura CombDescricao (getCombDescricao,
// Classes.pas:3698) NAO foi portada: devolve texto de impressao do DANFE
// ("01 - ALCOOL"), que e apresentacao e esta fora do escopo do port.
type VeicProd struct {
	TpOP         TipoOperacao    // tpOp
	Chassi       string          // chassi
	CCor         string          // cCor
	XCor         string          // xCor
	Pot          string          // pot
	Cilin        string          // cilin
	PesoL        string          // pesoL
	PesoB        string          // pesoB
	NSerie       string          // nSerie
	TpComb       string          // tpComb
	NMotor       string          // nMotor
	CMT          string          // CMT
	Dist         string          // dist
	AnoMod       int             // anoMod
	AnoFab       int             // anoFab
	TpPint       string          // tpPint
	TpVeic       int             // tpVeic
	EspVeic      int             // espVeic
	VIN          string          // VIN
	CondVeic     CondicaoVeiculo // condVeic
	CMod         string          // cMod
	CCorDENATRAN string          // cCorDENATRAN
	Lota         int             // lota
	TpRest       int             // tpRest
}

// Med e o grupo de medicamento. Porte de TMedCollectionItem.
//
// ATENCAO, etapa do gerador: este grupo tem condicional POR VERSAO DE LEIAUTE,
// nao por data -- o criterio do ACBr e `NFe.infNFe.Versao >= 4`, e sao TRES
// consequencias (ACBrNFe.XmlWriter.pas:1593-1620):
//
//	Versao >= 4 : saem cProdANVISA e xMotivoIsencao; MaxMed = 1
//	Versao <  4 : saem nLote, qLote, dFab e dVal;    MaxMed = 500
//
// Os quatro campos antigos viraram o grupo rastro na 4.00 e permanecem
// declarados para gerar as versoes anteriores. Emitir os dois conjuntos juntos
// produz XML rejeitado, e o limite de ocorrencias tambem muda.
type Med struct {
	CProdANVISA    string    // cProdANVISA (so Versao >= 4)
	XMotivoIsencao string    // xMotivoIsencao (so Versao >= 4)
	NLote          string    // nLote (so Versao < 4)
	QLote          float64   // qLote (so Versao < 4)
	DFab           time.Time // dFab (so Versao < 4)
	DVal           time.Time // dVal (so Versao < 4)
	VPMC           float64   // vPMC
}

// Arma e o grupo de armamento. Porte de TArmaCollectionItem.
//
// Sem construtor: a property declara `default taUsoPermitido`, que e o
// ordinal zero de TipoArma, e a classe nao tem Create proprio.
type Arma struct {
	TpArma TipoArma // tpArma
	NSerie string   // nSerie
	NCano  string   // nCano
	Descr  string   // descr
}

// CIDE e a contribuicao de intervencao no dominio economico do combustivel.
// Porte de TCIDE.
type CIDE struct {
	QBCProd   float64 // qBCProd
	VAliqProd float64 // vAliqProd
	VCIDE     float64 // vCIDE
}

// ICMSComb e o ICMS do combustivel. Porte de TICMSComb.
type ICMSComb struct {
	VBCICMS   float64 // vBCICMS
	VICMS     float64 // vICMS
	VBCICMSST float64 // vBCICMSST
	VICMSST   float64 // vICMSST
}

// ICMSInter e o ICMS interestadual do combustivel. Porte de TICMSInter.
type ICMSInter struct {
	VBCICMSSTDest float64 // vBCICMSSTDest
	VICMSSTDest   float64 // vICMSSTDest
}

// ICMSCons e o ICMS do combustivel na UF de consumo. Porte de TICMSCons.
type ICMSCons struct {
	VBCICMSSTCons float64 // vBCICMSSTCons
	VICMSSTCons   float64 // vICMSSTCons
	UFCons        string  // UFcons
}

// Encerrante e o grupo de encerrante do combustivel. Porte de Tencerrante.
type Encerrante struct {
	NBico   int     // nBico
	NBomba  int     // nBomba
	NTanque int     // nTanque
	VEncIni float64 // vEncIni
	VEncFin float64 // vEncFin
}

// OrigComb e a origem do combustivel. Porte de TorigCombCollectionItem.
type OrigComb struct {
	IndImport IndImport // indImport
	CUFOrig   int       // cUFOrig -- gravar com 2 digitos, zero-padded
	POrig     float64   // pOrig
}

// Comb e o grupo de combustivel. Porte de TComb.
type Comb struct {
	CProdANP   int        // cProdANP -- gravar com 9 digitos, zero-padded
	PMixGN     float64    // pMixGN
	DescANP    string     // descANP
	PGLP       float64    // pGLP
	PGNn       float64    // pGNn
	PGNi       float64    // pGNi
	VPart      float64    // vPart
	CODIF      string     // CODIF
	QTemp      float64    // qTemp
	UFCons     string     // UFCons
	CIDE       CIDE       // CIDE
	ICMS       ICMSComb   // ICMS
	ICMSInter  ICMSInter  // ICMSInter
	ICMSCons   ICMSCons   // ICMSCons
	Encerrante Encerrante // encerrante
	PBio       float64    // pBio
	OrigComb   []OrigComb // origComb
}

// ===========================================================================
// Produto
// ===========================================================================

// NVE e a Nomenclatura de Valor Aduaneiro e Estatistica. Porte de
// TNVECollectionItem.
type NVE struct {
	NVE string // NVE
}

// CredPresumido e o credito presumido do item. Porte de
// TCredPresumidoCollectionItem.
type CredPresumido struct {
	CCredPresumido string  // cCredPresumido
	PCredPresumido float64 // pCredPresumido
	VCredPresumido float64 // vCredPresumido
}

// Prod e o produto do item. Porte de TProd.
//
// Use NovoProd: IndEscala NAO comeca no zero. Veja o doc-comment do
// construtor.
//
// O campo genero (Fgenero) do ACBr esta comentado desde a versao 3.10 do
// leiaute (Classes.pas:951 e :1010) e nao foi portado.
type Prod struct {
	CProd         string          // cProd
	NItem         int             // nItem (atributo do det)
	CEAN          string          // cEAN
	CBarra        string          // cBarra
	XProd         string          // xProd
	NCM           string          // NCM
	NVE           []NVE           // NVE
	EXTIPI        string          // EXTIPI
	CFOP          string          // CFOP
	UCom          string          // uCom
	QCom          float64         // qCom
	VUnCom        float64         // vUnCom
	VProd         float64         // vProd
	CEANTrib      string          // cEANTrib
	CBarraTrib    string          // cBarraTrib
	UTrib         string          // uTrib
	QTrib         float64         // qTrib
	VUnTrib       float64         // vUnTrib
	VFrete        float64         // vFrete
	VSeg          float64         // vSeg
	VDesc         float64         // vDesc
	VOutro        float64         // vOutro
	IndTot        IndicadorTotal  // indTot
	DI            []DI            // DI
	XPed          string          // xPed
	NItemPed      string          // nItemPed
	DetExport     []DetExport     // detExport
	Rastro        []Rastro        // rastro
	VeicProd      VeicProd        // veicProd
	Med           []Med           // med
	Arma          []Arma          // arma
	Comb          Comb            // comb
	NRECOPI       string          // nRECOPI
	NFCI          string          // nFCI
	CEST          string          // CEST
	IndEscala     IndEscala       // indEscala
	CNPJFab       string          // CNPJFab
	CBenef        string          // cBenef
	CredPresumido []CredPresumido // credPresumido

	// Reforma Tributaria
	IndBemMovelUsado pcn.IndicadorEx      // indBemMovelUsado
	TpCredPresIBSZFM rtc.TpCredPresIBSZFM // tpCredPresIBSZFM
}

// NovoProd devolve um produto com o default real de TProd.Create
// (Classes.pas:2747): FindEscala := ieNenhum.
//
// ieNenhum e o ordinal DOIS de TpcnIndEscala = (ieRelevante, ieNaoRelevante,
// ieNenhum). O zero value do Go seria IeRelevante, que grava "S" -- todo item
// sairia declarando producao em escala relevante, campo que o leiaute so
// admite junto do CEST.
//
// Os demais defaults de TProd coincidem com o zero: IndTot nasce
// ItSomaTotalNFe ("1", o item compoe o total da nota) e IndBemMovelUsado
// nasce TieNenhum (vazio).
func NovoProd() Prod {
	return Prod{IndEscala: IeNenhum}
}

// ===========================================================================
// Impostos do item
// ===========================================================================

// ICMS e o grupo de ICMS do item. Porte de TICMS.
//
// E o maior grupo do leiaute: o ACBr achata numa UNICA classe todos os CST e
// CSOSN, e o gerador escolhe quais campos emitir pelo CST. Os comentarios
// N11..N32 do .pas sao a numeracao do campo no leiaute e foram mantidos.
//
// Sem construtor: TICMS.Create (Classes.pas:4170) so atribui
// FindDeduzDeson := tieNenhum, que ja e o ordinal zero de pcn.IndicadorEx. As
// properties declaram ainda `default oeNacional`, `default cst00`,
// `default dbiMargemValorAgregado` e `default dbisPrecoTabelado`, mas --
// como em Ide -- `default` nao inicializa nada. Os tres primeiros coincidem
// com o ordinal zero; quanto ao CST, o zero de pcn.CSTIcms e CSTVazio, e o
// Delphi tambem declara cstVazio como PRIMEIRO membro
// (ACBrDFe.Conversao.pas:256) -- entao o zero value reproduz o ACBr.
type ICMS struct {
	Orig            pcn.OrigemMercadoria   // orig        N11
	CST             pcn.CSTIcms            // CST         N12
	CSOSN           pcn.CSOSNIcms          // CSOSN       N12a
	ModBC           DeterminacaoBaseIcms   // modBC       N13
	PRedBC          float64                // pRedBC      N14
	VBC             float64                // vBC         N15
	PICMS           float64                // pICMS       N16
	VICMS           float64                // vICMS       N17
	ModBCST         DeterminacaoBaseIcmsST // modBCST     N18
	PMVAST          float64                // pMVAST      N19
	PRedBCST        float64                // pRedBCST    N20
	VBCST           float64                // vBCST       N21
	PICMSST         float64                // pICMSST     N22
	VICMSST         float64                // vICMSST     N23
	UFST            string                 // UFST        N24
	PBCOp           float64                // pBCOp       N25
	VBCSTRet        float64                // vBCSTRet    N26
	VICMSSTRet      float64                // vICMSSTRet  N27
	MotDesICMS      MotivoDesoneracaoICMS  // motDesICMS  N28
	PCredSN         float64                // pCredSN     N29
	VCredICMSSN     float64                // vCredICMSSN N30
	VBCSTDest       float64                // vBCSTDest   N31
	VICMSSTDest     float64                // vICMSSTDest N32
	VICMSDeson      float64                // vICMSDeson
	VICMSOp         float64                // vICMSOp
	PDif            float64                // pDif
	VICMSDif        float64                // vICMSDif
	VBCFCP          float64                // vBCFCP
	PFCP            float64                // pFCP
	VFCP            float64                // vFCP
	VBCFCPST        float64                // vBCFCPST
	PFCPST          float64                // pFCPST
	VFCPST          float64                // vFCPST
	VBCFCPSTRet     float64                // vBCFCPSTRet
	PFCPSTRet       float64                // pFCPSTRet
	VFCPSTRet       float64                // vFCPSTRet
	PST             float64                // pST
	PRedBCEfet      float64                // pRedBCEfet
	VBCEfet         float64                // vBCEfet
	PICMSEfet       float64                // pICMSEfet
	VICMSEfet       float64                // vICMSEfet
	VICMSSubstituto float64                // vICMSSubstituto
	VICMSSTDeson    float64                // vICMSSTDeson
	MotDesICMSST    MotivoDesoneracaoICMS  // motDesICMSST
	VFCPDif         float64                // vFCPDif
	VFCPEfet        float64                // vFCPEfet
	PFCPDif         float64                // pFCPDif

	// Monofasia -- CST 02 e 15
	AdRemICMS float64 // adRemICMS
	VICMSMono float64 // vICMSMono
	QBCMono   float64 // qBCMono

	// CST 15
	AdRemICMSReten float64     // adRemICMSReten
	VICMSMonoReten float64     // vICMSMonoReten
	QBCMonoReten   float64     // qBCMonoReten
	PRedAdRem      float64     // pRedAdRem
	MotRedAdRem    MotRedAdRem // motRedAdRem

	// CST 53
	VICMSMonoOp  float64 // vICMSMonoOp
	QBCMonoDif   float64 // qBCMonoDif
	AdRemICMSDif float64 // adRemICMSDif
	VICMSMonoDif float64 // vICMSMonoDif

	// CST 61
	AdRemICMSRet float64 // adRemICMSRet
	VICMSMonoRet float64 // vICMSMonoRet
	QBCMonoRet   float64 // qBCMonoRet

	IndDeduzDeson pcn.IndicadorEx // indDeduzDeson
	CBenefRBC     string          // cBenefRBC
}

// IPI e o grupo de IPI do item. Porte de TIPI.
//
// Sem construtor: a property declara `default ipi00`, que coincide com o
// ordinal zero de CstIpi. Cuidado ao mexer naquele enum -- a ordem dele NAO
// segue a sequencia numerica (00, 49, 50, 99, 01...), e o zero value grava
// "00", nao vazio.
type IPI struct {
	ClEnq    string  // clEnq
	CNPJProd string  // CNPJProd
	CSelo    string  // cSelo
	QSelo    int     // qSelo
	CEnq     string  // cEnq
	CST      CstIpi  // CST
	VBC      float64 // vBC
	QUnid    float64 // qUnid
	VUnid    float64 // vUnid
	PIPI     float64 // pIPI
	VIPI     float64 // vIPI
}

// II e o grupo de Imposto de Importacao do item. Porte de TII.
type II struct {
	VBC      float64 // vBC
	VDespAdu float64 // vDespAdu
	VII      float64 // vII
	VIOF     float64 // vIOF
}

// PIS e o grupo de PIS do item. Porte de TPIS.
//
// Sem construtor: a property declara `default pis01`, que coincide com o
// ordinal zero de pcn.CSTPis. O zero value grava "01", nao vazio.
type PIS struct {
	CST       pcn.CSTPis // CST
	VBC       float64    // vBC
	PPIS      float64    // pPIS
	VPIS      float64    // vPIS
	QBCProd   float64    // qBCProd
	VAliqProd float64    // vAliqProd
}

// PISST e o grupo de PIS substituicao tributaria. Porte de TPISST.
type PISST struct {
	VBC          float64      // vBC
	PPIS         float64      // pPIS
	QBCProd      float64      // qBCProd
	VAliqProd    float64      // vAliqProd
	VPIS         float64      // vPIS
	IndSomaPISST IndSomaPISST // indSomaPISST
}

// COFINS e o grupo de COFINS do item. Porte de TCOFINS.
//
// Sem construtor: a property declara `default cof01`, que coincide com o
// ordinal zero de pcn.CSTCofins. O zero value grava "01", nao vazio.
type COFINS struct {
	CST       pcn.CSTCofins // CST
	VBC       float64       // vBC
	PCOFINS   float64       // pCOFINS
	VCOFINS   float64       // vCOFINS
	VBCProd   float64       // vBCProd
	VAliqProd float64       // vAliqProd
	QBCProd   float64       // qBCProd
}

// COFINSST e o grupo de COFINS substituicao tributaria. Porte de TCOFINSST.
type COFINSST struct {
	VBC             float64         // vBC
	PCOFINS         float64         // pCOFINS
	QBCProd         float64         // qBCProd
	VAliqProd       float64         // vAliqProd
	VCOFINS         float64         // vCOFINS
	IndSomaCOFINSST IndSomaCOFINSST // indSomaCOFINSST
}

// ISSQN e o grupo de ISSQN do item. Porte de TISSQN.
//
// Contem os dez campos que TISSQNtot mantem COMENTADOS: eles pertencem ao
// item, nao ao total da nota.
//
// Sem construtor: a property declara `default ISSQNcSitTribVazio`, que
// coincide com o ordinal zero de ISSQNcSitTrib.
type ISSQN struct {
	VBC          float64          // vBC
	VAliq        float64          // vAliq
	VISSQN       float64          // vISSQN
	CMunFG       int              // cMunFG -- gravar com 7 digitos, zero-padded
	CListServ    string           // cListServ
	CSitTrib     ISSQNcSitTrib    // cSitTrib
	VDeducao     float64          // vDeducao
	VOutro       float64          // vOutro
	VDescIncond  float64          // vDescIncond
	VDescCond    float64          // vDescCond
	IndISSRet    IndISSRet        // indISSRet
	VISSRet      float64          // vISSRet
	IndISS       IndISS           // indISS
	CServico     string           // cServico
	CMun         int              // cMun
	CPais        int              // cPais
	NProcesso    string           // nProcesso
	IndIncentivo pcn.IndIncentivo // indIncentivo
}

// ICMSUFDest e o grupo de partilha do ICMS para a UF de destino (EC 87/2015).
// Porte de TICMSUFDest.
type ICMSUFDest struct {
	VBCUFDest      float64 // vBCUFDest
	VBCFCPUFDest   float64 // vBCFCPUFDest
	PFCPUFDest     float64 // pFCPUFDest
	PICMSUFDest    float64 // pICMSUFDest
	PICMSInter     float64 // pICMSInter
	PICMSInterPart float64 // pICMSInterPart
	VFCPUFDest     float64 // vFCPUFDest
	VICMSUFDest    float64 // vICMSUFDest
	VICMSUFRemet   float64 // vICMSUFRemet
}

// Imposto e o grupo de impostos do item. Porte de TImposto.
type Imposto struct {
	VTotTrib   float64    // vTotTrib
	ICMS       ICMS       // ICMS
	IPI        IPI        // IPI
	II         II         // II
	PIS        PIS        // PIS
	PISST      PISST      // PISST
	COFINS     COFINS     // COFINS
	COFINSST   COFINSST   // COFINSST
	ISSQN      ISSQN      // ISSQN
	ICMSUFDest ICMSUFDest // ICMSUFDest

	// Reforma Tributaria
	ISel   rtc.GIS    // IS
	IBSCBS rtc.IBSCBS // IBSCBS
}

// ===========================================================================
// Item
// ===========================================================================

// ObsItem e uma observacao de item. Porte de TobsItem.
//
// O mesmo tipo serve a obsCont e a obsFisco do item -- no ACBr sao duas
// instancias de TobsItem, nao duas classes.
type ObsItem struct {
	XCampo string // xCampo (atributo)
	XTexto string // xTexto
}

// DetItem e um item da nota. Porte de TDetCollectionItem.
//
// Use NovoDetItem: o Prod embutido precisa do default de IndEscala.
type DetItem struct {
	Prod      Prod    // prod
	Imposto   Imposto // imposto
	PDevol    float64 // pDevol
	VIPIDevol float64 // vIPIDevol
	InfAdProd string  // infAdProd
	ObsCont   ObsItem // obsCont
	ObsFisco  ObsItem // obsFisco

	// Reforma Tributaria
	VItem           float64             // vItem
	DFeReferenciado rtc.DFeReferenciado // DFeReferenciado
}

// NovoDetItem devolve um item com os defaults de TDetCollectionItem.Create
// (Classes.pas:2537), que alem de alocar os filhos herda o default de
// TProd.Create.
func NovoDetItem() DetItem {
	return DetItem{Prod: NovoProd()}
}
