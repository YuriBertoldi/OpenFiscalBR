// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-20.
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

// Modelo de dados da NF-e. Porte de Base/ACBrNFe.Classes.pas.
//
// Os nomes de campo repetem a tag do leiaute da SEFAZ (cUF, nNF, vBC...) em
// PascalCase, com a tag original em comentario. E o contrato com o XSD e com
// quem ja conhece o ACBr; renomear quebraria a rastreabilidade com o .pas.
//
// # Colecoes
//
// Todo TXxxCollection/TXxxCollectionItem do Delphi vira um slice simples. A
// unica excecao e Pag, porque TpagCollection carrega vTroco NA COLECAO
// (Classes.pas:1919-1934) -- virou struct com DetPag e VTroco.
//
// # Defaults: o `default` da property NAO inicializa nada
//
// No Delphi, `property X: T ... default valor` e metadado de streaming de
// design-time; quem inicializa e o `constructor Create`. Os dois DIVERGEM na
// NF-e, e seguir a declaracao da property geraria XML diferente do ACBr:
//
//	property                       declara          Create faz
//	Ide.indPag                     ipPrazo          nada -> ipVista ("0")
//	Ide.tpNF                       tnSaida          nada -> tnEntrada ("0")
//	Ide.tpImp                      tiPaisagem       nada -> tiSemGeracao ("0")
//	Ide.tpAmb                      taHomologacao    nada -> taProducao ("1")
//
// Ou seja, um TIde recem-criado no ACBr esta em PRODUCAO, nao em homologacao.
// O zero value do Go reproduz isso sem esforco, e e por isso que Ide nao tem
// construtor aqui.
//
// Os tres pontos em que o Create do Delphi REALMENTE atribui algo diferente do
// zero estao em NovaNFe, NovoEmit e NovoPagItem -- sem eles o zero value do Go
// divergiria do ACBr. Veja o doc-comment de cada um.

// ===========================================================================
// infNFe e identificacao
// ===========================================================================

// InfNFe e o grupo raiz da nota. Porte de TinfNFe.
//
// Versao e numerica porque o ACBr a compara por faixa (>= 3, >= 4 em
// XmlWriter.pas), mas o XSD exige o literal "4.00": quem gerar precisa
// formatar com DUAS casas (FloatToString '#0.00', XmlWriter.pas:330 e :392),
// nunca com strconv.FormatFloat default.
type InfNFe struct {
	ID     string  // Id (atributo)
	Versao float64 // versao (atributo) -- gravar com 2 casas: "4.00"
}

// RefNF e a referencia a uma NF modelo 1/1A. Porte de TRefNF.
type RefNF struct {
	CUF    int    // cUF
	AAMM   string // AAMM
	CNPJ   string // CNPJ
	Modelo int    // mod
	Serie  int    // serie
	NNF    int    // nNF
}

// RefECF e a referencia a um cupom fiscal. Porte de TRefECF.
type RefECF struct {
	Modelo ECFModRef // mod
	NECF   string    // nECF
	NCOO   string    // nCOO
}

// RefNFP e a referencia a uma NF de produtor rural. Porte de TRefNFP.
//
// Note que modelo aqui e string, enquanto em RefNF e int -- e assim no ACBr
// (Classes.pas:81 x :116). O writer, porem, grava os DOIS com tcInt 02/02
// (XmlWriter.pas:663 e :692), e o XSD exige "01"/"02": o zero a esquerda vem
// do zero-padding do gerador, nao do tipo.
type RefNFP struct {
	CNPJCPF string // CNPJ ou CPF
	CUF     int    // cUF
	AAMM    string // AAMM
	IE      string // IE
	Modelo  string // mod
	Serie   int    // serie
	NNF     int    // nNF
}

// NFref e uma nota referenciada. Porte de TNFrefCollectionItem.
//
// Os cinco grupos sao mutuamente exclusivos no leiaute, mas o ACBr declara
// todos no mesmo item e decide na geracao pelo que estiver preenchido.
type NFref struct {
	RefNFe    string // refNFe
	RefNFeSig string // refNFeSig
	RefCTe    string // refCTe
	RefNF     RefNF  // refNF
	RefECF    RefECF // refECF
	RefNFP    RefNFP // refNFP
}

// Ide e o grupo de identificacao da nota. Porte de TIde.
//
// NAO tem construtor de proposito: TIde.Create so aloca os filhos e nao
// atribui nenhum default, entao o zero value do Go ja reproduz o ACBr. Veja a
// tabela de defaults no topo deste arquivo.
type Ide struct {
	CUF         int                 // cUF
	CNF         int                 // cNF
	NatOp       string              // natOp
	IndPag      IndicadorPagamento  // indPag (saiu do leiaute 4.00)
	Modelo      int                 // mod
	Serie       int                 // serie
	NNF         int                 // nNF
	DEmi        time.Time           // dhEmi (dEmi ate a versao 3.10)
	DSaiEnt     time.Time           // dhSaiEnt (dSaiEnt ate a 3.10)
	HSaiEnt     time.Time           // hSaiEnt (so ate a 3.10)
	TpNF        pcn.TipoNFe         // tpNF
	IdDest      DestinoOperacao     // idDest
	CMunFG      int                 // cMunFG
	NFref       []NFref             // NFref
	TpImp       pcn.TipoImpressao   // tpImp
	TpEmis      pcn.TipoEmissao     // tpEmis
	CDV         int                 // cDV
	TpAmb       pcn.TipoAmbiente    // tpAmb
	FinNFe      FinalidadeNFe       // finNFe
	IndFinal    ConsumidorFinal     // indFinal
	IndPres     PresencaComprador   // indPres
	IndIntermed IndIntermed         // indIntermed
	CIndOp      string              // cIndOp
	ProcEmi     pcn.ProcessoEmissao // procEmi
	VerProc     string              // verProc
	DhCont      time.Time           // dhCont
	XJust       string              // xJust

	// Reforma Tributaria
	DPrevEntrega   time.Time          // dPrevEntrega
	CMunFGIBS      int                // cMunFGIBS
	TpNFDebito     TpNFDebito         // tpNFDebito
	TpNFCredito    TpNFCredito        // tpNFCredito
	GCompraGov     rtc.GCompraGov     // gCompraGov
	GPagAntecipado rtc.GPagAntecipado // gPagAntecipado
}

// ===========================================================================
// Emitente
// ===========================================================================

// EnderEmit e o endereco do emitente. Porte de TenderEmit.
type EnderEmit struct {
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
	CEP     int    // CEP -- gravar com 8 digitos, zero-padded
	CPais   int    // cPais
	XPais   string // xPais
	Fone    string // fone
}

// Emit e o emitente da nota. Porte de TEmit.
//
// Use NovoEmit fora do contexto de NovaNFe: o CRT do ACBr NAO comeca no zero.
type Emit struct {
	CNPJCPF   string    // CNPJ ou CPF
	XNome     string    // xNome
	XFant     string    // xFant
	EnderEmit EnderEmit // enderEmit
	IE        string    // IE
	IEST      string    // IEST
	IM        string    // IM
	CNAE      string    // CNAE
	CRT       CRT       // CRT
	ISUFEmit  string    // ISUFEmit
}

// NovoEmit devolve um emitente com o default real de TEmit.Create
// (Classes.pas:2702): CRT = crtRegimeNormal.
//
// E o unico campo de Emit cujo zero value diverge do ACBr -- o zero de CRT e
// CrtSimplesNacional ("1"), que mudaria o regime tributario declarado na nota.
func NovoEmit() Emit {
	return Emit{CRT: CrtRegimeNormal}
}

// Avulsa e o grupo do fisco emitente de nota avulsa. Porte de TAvulsa.
type Avulsa struct {
	CNPJ    string    // CNPJ
	XOrgao  string    // xOrgao
	Matr    string    // matr
	XAgente string    // xAgente
	Fone    string    // fone
	UF      string    // UF
	NDAR    string    // nDAR
	DEmi    time.Time // dEmi
	VDAR    float64   // vDAR
	RepEmi  string    // repEmi
	DPag    time.Time // dPag
}

// ===========================================================================
// Destinatario, retirada e entrega
// ===========================================================================

// EnderDest e o endereco do destinatario. Porte de TEnderDest.
type EnderDest struct {
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
	CEP     int    // CEP -- gravar com 8 digitos, zero-padded
	CPais   int    // cPais
	XPais   string // xPais
	Fone    string // fone
}

// Dest e o destinatario da nota. Porte de TDest.
type Dest struct {
	CNPJCPF       string        // CNPJ ou CPF
	IDEstrangeiro string        // idEstrangeiro
	XNome         string        // xNome
	EnderDest     EnderDest     // enderDest
	IndIEDest     pcn.IndIEDest // indIEDest
	IE            string        // IE
	ISUF          string        // ISUF
	IM            string        // IM
	Email         string        // email
}

// Retirada e o local de retirada da mercadoria. Porte de TRetirada.
type Retirada struct {
	CNPJCPF string // CNPJ ou CPF
	XNome   string // xNome
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
	CEP     int    // CEP -- gravar com 8 digitos, zero-padded
	CPais   int    // cPais
	XPais   string // xPais
	Fone    string // fone
	Email   string // email
	IE      string // IE
}

// Entrega e o local de entrega da mercadoria. Porte de TEntrega.
//
// Tem exatamente os mesmos campos de Retirada no ACBr (Classes.pas:402 x
// :440). Mantidos como dois tipos, e nao um alias, porque o leiaute pode
// diverge-los e porque e assim que o .pas declara.
type Entrega struct {
	CNPJCPF string // CNPJ ou CPF
	XNome   string // xNome
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
	CEP     int    // CEP -- gravar com 8 digitos, zero-padded
	CPais   int    // cPais
	XPais   string // xPais
	Fone    string // fone
	Email   string // email
	IE      string // IE
}

// AutXML e um CNPJ/CPF autorizado a baixar o XML. Porte de
// TautXMLCollectionItem.
type AutXML struct {
	CNPJCPF string // CNPJ ou CPF
}

// ===========================================================================
// Totais
// ===========================================================================

// ICMSTot e o total dos tributos da nota. Porte de TICMSTot.
type ICMSTot struct {
	VBC            float64 // vBC
	VICMS          float64 // vICMS
	VICMSDeson     float64 // vICMSDeson
	VFCPUFDest     float64 // vFCPUFDest
	VICMSUFDest    float64 // vICMSUFDest
	VICMSUFRemet   float64 // vICMSUFRemet
	VFCP           float64 // vFCP
	VBCST          float64 // vBCST
	VST            float64 // vST
	VFCPST         float64 // vFCPST
	VFCPSTRet      float64 // vFCPSTRet
	VProd          float64 // vProd
	VFrete         float64 // vFrete
	VSeg           float64 // vSeg
	VDesc          float64 // vDesc
	VII            float64 // vII
	VIPI           float64 // vIPI
	VIPIDevol      float64 // vIPIDevol
	VPIS           float64 // vPIS
	VCOFINS        float64 // vCOFINS
	VOutro         float64 // vOutro
	VNF            float64 // vNF
	VTotTrib       float64 // vTotTrib
	VICMSMono      float64 // vICMSMono
	VICMSMonoReten float64 // vICMSMonoReten
	VICMSMonoRet   float64 // vICMSMonoRet
	QBCMono        float64 // qBCMono
	QBCMonoReten   float64 // qBCMonoReten
	QBCMonoRet     float64 // qBCMonoRet
}

// ISSQNTot e o total do ISSQN. Porte de TISSQNtot.
//
// O ACBr mantem dez campos COMENTADOS nesta classe (vINSS, vIR, vCSLL,
// indISSRet, indISS, cServico, cMun, cPais, nProcesso, indIncentivo --
// Classes.pas:1560-1598). Eles pertencem ao grupo ISSQN do ITEM, nao ao
// total, e nao foram portados.
type ISSQNTot struct {
	VServ       float64          // vServ
	VBC         float64          // vBC
	VISS        float64          // vISS
	VPIS        float64          // vPIS
	VCOFINS     float64          // vCOFINS
	DCompet     time.Time        // dCompet
	VDeducao    float64          // vDeducao
	VOutro      float64          // vOutro
	VDescIncond float64          // vDescIncond
	VDescCond   float64          // vDescCond
	VISSRet     float64          // vISSRet
	CRegTrib    pcn.RegTribISSQN // cRegTrib
}

// RetTrib e o total dos tributos retidos. Porte de TretTrib.
type RetTrib struct {
	VRetPIS    float64 // vRetPIS
	VRetCOFINS float64 // vRetCOFINS
	VRetCSLL   float64 // vRetCSLL
	VBCIRRF    float64 // vBCIRRF
	VIRRF      float64 // vIRRF
	VBCRetPrev float64 // vBCRetPrev
	VRetPrev   float64 // vRetPrev
}

// Total e o grupo de totais da nota. Porte de TTotal.
type Total struct {
	ICMSTot  ICMSTot  // ICMSTot
	ISSQNTot ISSQNTot // ISSQNtot
	RetTrib  RetTrib  // retTrib

	// Reforma Tributaria
	ISTot     rtc.ISTot     // ISTot
	IBSCBSTot rtc.IBSCBSTot // IBSCBSTot
	VNFTot    float64       // vNFTot
}

// ===========================================================================
// Transporte
// ===========================================================================

// Transporta e o transportador. Porte de TTransporta.
type Transporta struct {
	CNPJCPF string // CNPJ ou CPF
	XNome   string // xNome
	IE      string // IE
	XEnder  string // xEnder
	XMun    string // xMun
	UF      string // UF
}

// RetTransp e a retencao do ICMS sobre o transporte. Porte de TretTransp.
type RetTransp struct {
	VServ    float64 // vServ
	VBCRet   float64 // vBCRet
	PICMSRet float64 // pICMSRet
	VICMSRet float64 // vICMSRet
	CFOP     string  // CFOP
	CMunFG   int     // cMunFG
}

// VeicTransp e o veiculo do transporte. Porte de TveicTransp.
type VeicTransp struct {
	Placa string // placa
	UF    string // UF
	RNTC  string // RNTC
}

// Reboque e um reboque do veiculo de transporte. Porte de
// TReboqueCollectionItem. Mesmos campos de VeicTransp, como no ACBr.
type Reboque struct {
	Placa string // placa
	UF    string // UF
	RNTC  string // RNTC
}

// Lacre e o numero de um lacre de volume. Porte de TLacresCollectionItem.
type Lacre struct {
	NLacre string // nLacre
}

// Vol e um volume transportado. Porte de TVolCollectionItem.
type Vol struct {
	QVol   int     // qVol
	Esp    string  // esp
	Marca  string  // marca
	NVol   string  // nVol
	PesoL  float64 // pesoL
	PesoB  float64 // pesoB
	Lacres []Lacre // lacres
}

// Transp e o grupo de transporte da nota. Porte de TTransp.
type Transp struct {
	ModFrete   ModalidadeFrete // modFrete
	Transporta Transporta      // transporta
	RetTransp  RetTransp       // retTransp
	VeicTransp VeicTransp      // veicTransp
	Vol        []Vol           // vol
	Reboque    []Reboque       // reboque
	Vagao      string          // vagao
	Balsa      string          // balsa
}

// ===========================================================================
// Cobranca e pagamento
// ===========================================================================

// Fat e a fatura da cobranca. Porte de TFat.
type Fat struct {
	NFat  string  // nFat
	VOrig float64 // vOrig
	VDesc float64 // vDesc
	VLiq  float64 // vLiq
}

// Dup e uma duplicata da cobranca. Porte de TDupCollectionItem.
type Dup struct {
	NDup  string    // nDup
	DVenc time.Time // dVenc
	VDup  float64   // vDup
}

// Cobr e o grupo de cobranca da nota. Porte de TCobr.
type Cobr struct {
	Fat Fat   // fat
	Dup []Dup // dup
}

// PagItem e uma forma de pagamento. Porte de TpagCollectionItem.
//
// Use NovoPagItem: DOIS campos deste item comecam fora do zero no ACBr, e os
// dois saem no XML se ficarem no zero do Go. Veja o doc-comment de NovoPagItem.
type PagItem struct {
	IndPag    IndicadorPagamento // indPag
	TPag      FormaPagamento     // tPag
	XPag      string             // xPag
	VPag      float64            // vPag
	TpIntegra pcn.TpIntegra      // tpIntegra
	CNPJ      string             // CNPJ
	TBand     BandeiraCartao     // tBand
	CAut      string             // cAut
	DPag      time.Time          // dPag
	CNPJPag   string             // CNPJPag
	UFPag     string             // UFPag
	CNPJReceb string             // CNPJReceb
	IDTermPag string             // idTermPag
}

// NovoPagItem devolve uma forma de pagamento com os defaults reais de
// TpagCollectionItem.Create (Classes.pas:4441):
//
//	FindPag := ipNenhum
//	FtBand  := TpcnBandeiraCartao(-1)
//
// Os dois divergem do zero value do Go, e nos dois a diferenca VAZA PARA O
// XML: o zero de IndicadorPagamento e IpVista, que grava "0" em vez de nada,
// e o zero de BandeiraCartao e BcVisa, que grava "01" -- declarando pagamento
// a vista com cartao Visa num pagamento que nao tem nem um nem outro.
//
// E o mesmo defeito que o porte da NFGas encontrou em tiSim: membro "ausente"
// que nao esta no ordinal zero.
func NovoPagItem() PagItem {
	return PagItem{
		IndPag: IpNenhum,
		TBand:  BcNenhum,
	}
}

// Pag e o grupo de pagamento. Porte de TpagCollection.
//
// Nao e um slice simples porque vTroco e property da COLECAO no ACBr
// (Classes.pas:1933), nao do item: o troco e um so para a nota inteira.
//
// O slice se chama DetPag por ser o nome da tag do grupo no leiaute, e nao
// Itens -- a convencao do projeto e preservar o nome do leiaute nos campos.
type Pag struct {
	DetPag []PagItem // detPag
	VTroco float64   // vTroco
}

// InfIntermed e o intermediador da operacao. Porte de TinfIntermed.
type InfIntermed struct {
	CNPJ         string // CNPJ
	IDCadIntTran string // idCadIntTran
}

// ===========================================================================
// Informacoes adicionais
// ===========================================================================

// ObsCont e uma observacao de interesse do contribuinte. Porte de
// TobsContCollectionItem.
type ObsCont struct {
	XCampo string // xCampo (atributo)
	XTexto string // xTexto
}

// ObsFisco e uma observacao de interesse do fisco. Porte de
// TobsFiscoCollectionItem.
type ObsFisco struct {
	XCampo string // xCampo (atributo)
	XTexto string // xTexto
}

// ProcRef e um processo referenciado. Porte de TprocRefCollectionItem.
type ProcRef struct {
	NProc   string            // nProc
	IndProc IndicadorProcesso // indProc
	TpAto   TpAto             // tpAto
}

// InfAdic e o grupo de informacoes adicionais. Porte de TInfAdic.
type InfAdic struct {
	InfAdFisco string     // infAdFisco
	InfCpl     string     // infCpl
	ObsCont    []ObsCont  // obsCont
	ObsFisco   []ObsFisco // obsFisco
	ProcRef    []ProcRef  // procRef
}

// Exporta e o grupo de exportacao. Porte de TExporta.
type Exporta struct {
	UFEmbarq     string // UFEmbarq (ate a versao 3.10)
	XLocEmbarq   string // xLocEmbarq (ate a 3.10)
	UFSaidaPais  string // UFSaidaPais (da 3.10 em diante)
	XLocExporta  string // xLocExporta (da 3.10 em diante)
	XLocDespacho string // xLocDespacho (da 3.10 em diante)
}

// Compra e o grupo de compras publicas. Porte de TCompra.
type Compra struct {
	XNEmp string // xNEmp
	XPed  string // xPed
	XCont string // xCont
}

// ForDia e o fornecimento diario de cana. Porte de TForDiaCollectionItem.
type ForDia struct {
	Dia  int     // dia (atributo)
	Qtde float64 // qtde
}

// Deduc e uma deducao da cana. Porte de TDeducCollectionItem.
type Deduc struct {
	XDed string  // xDed
	VDed float64 // vDed
}

// Cana e o grupo de aquisicao de cana. Porte de Tcana.
type Cana struct {
	Safra   string   // safra
	Ref     string   // ref
	ForDia  []ForDia // forDia
	QTotMes float64  // qTotMes
	QTotAnt float64  // qTotAnt
	QTotGer float64  // qTotGer
	Deduc   []Deduc  // deduc
	VFor    float64  // vFor
	VTotDed float64  // vTotDed
	VLiqFor float64  // vLiqFor
}

// ===========================================================================
// Suplemento, responsavel tecnico, agropecuario e PAA
// ===========================================================================

// InfNFeSupl e o grupo suplementar da NFC-e (QR-Code). Porte de TinfNFeSupl.
type InfNFeSupl struct {
	QRCode   string // qrCode
	URLChave string // urlChave
}

// InfRespTec e o responsavel tecnico pela emissao. Porte de TinfRespTec.
type InfRespTec struct {
	CNPJ     string // CNPJ
	XContato string // xContato
	Email    string // email
	Fone     string // fone
	IDCSRT   int    // idCSRT
	HashCSRT string // hashCSRT
}

// Defensivo e o receituario de defensivo agricola. Porte de
// TdefensivoCollectionItem.
type Defensivo struct {
	NReceituario string // nReceituario
	CPFRespTec   string // CPFRespTec
}

// GuiaTransito e a guia de transito animal ou vegetal. Porte de
// TguiaTransito.
//
// Sem construtor: TguiaTransito.Create (Classes.pas:4669) atribui tpgNenhum,
// que ja e o ordinal zero de TpGuia. A afirmacao depende da ORDEM daquele
// enum, entao esta travada por TestGuiaTransito_ZeroValueNaoDeclaraGuia.
type GuiaTransito struct {
	UFGuia    string // UFGuia
	TpGuia    TpGuia // tpGuia
	SerieGuia string // serieGuia
	NGuia     string // nGuia
}

// Agropecuario e o grupo de operacao agropecuaria. Porte de Tagropecuario.
type Agropecuario struct {
	Defensivo    []Defensivo  // defensivo
	GuiaTransito GuiaTransito // guiaTransito
}

// InfPAA e o grupo do Provedor de Assinatura e Autorizacao. Porte de TinfPAA.
type InfPAA struct {
	CNPJPAA        string // CNPJPAA
	SignatureValue string // SignatureValue
	Modulus        string // Modulus
	Exponent       string // Exponent
}

// ===========================================================================
// Raiz
// ===========================================================================

// NFe e a nota fiscal eletronica completa. Porte de TNFe.
//
// Use NovaNFe para obter os defaults do TNFe.Create.
//
// O campo Det (itens da nota) entra na etapa 2b, junto com o subgrupo det
// inteiro -- produto, impostos e os grupos especificos por tipo de mercadoria.
type NFe struct {
	InfNFe       InfNFe       // infNFe
	Ide          Ide          // ide
	Emit         Emit         // emit
	Avulsa       Avulsa       // avulsa
	Dest         Dest         // dest
	Retirada     Retirada     // retirada
	Entrega      Entrega      // entrega
	AutXML       []AutXML     // autXML
	Total        Total        // total
	Transp       Transp       // transp
	Cobr         Cobr         // cobr
	Pag          Pag          // pag
	InfIntermed  InfIntermed  // infIntermed
	InfAdic      InfAdic      // infAdic
	Exporta      Exporta      // exporta
	Compra       Compra       // compra
	Cana         Cana         // cana
	InfNFeSupl   InfNFeSupl   // infNFeSupl
	InfRespTec   InfRespTec   // infRespTec
	Agropecuario Agropecuario // agropecuario
	InfPAA       InfPAA       // infPAA

	Signature pcn.Signature // Signature
	ProcNFe   pcn.ProcDFe   // nfeProc
}

// Os tres construtores deste arquivo devolvem VALOR, e nao ponteiro como o
// NovoXxx dos packages irmaos: aqui eles produzem structs de dados que a NFe
// embute por valor, e devolver ponteiro obrigaria a desreferenciar em toda
// atribuicao (n.Emit = *NovoEmit()). Nao ha estado nem identidade a preservar.
//
// NovaNFe devolve uma nota com os defaults reais de TNFe.Create
// (Classes.pas:2406-2440), que sao tres e todos saem no XML:
//
//	FinfNFe.Versao := 0
//	FEmit.EnderEmit.xPais := 'BRASIL';  cPais := 1058
//	FDest.EnderDest.xPais := 'BRASIL';  cPais := 1058
//
// Mais o CRT de TEmit.Create, que TNFe.Create herda por construir o emitente.
//
// Quem montar uma NFe por literal (NFe{...}) nao recebe nada disso: o pais do
// emitente e do destinatario sai vazio e o CRT sai como Simples Nacional.
func NovaNFe() NFe {
	var n NFe
	n.InfNFe.Versao = 0
	n.Emit = NovoEmit()
	n.Emit.EnderEmit.XPais = paisPadrao
	n.Emit.EnderEmit.CPais = codigoPaisPadrao
	n.Dest.EnderDest.XPais = paisPadrao
	n.Dest.EnderDest.CPais = codigoPaisPadrao
	return n
}

// Default de pais aplicado por TNFe.Create ao emitente e ao destinatario.
// O codigo e o CODIGO_BRASIL de ACBrDFeConsts.pas:40.
const (
	paisPadrao       = "BRASIL"
	codigoPaisPadrao = 1058
)

// Valores especiais de municipio e UF para endereco no exterior. Porte das
// constantes de ACBrNFe.Classes.pas:2366-2369.
//
// Quem consome: AjustarMunicipioUF (Base/ACBrNFe.XmlWriter.pas:255-257)
// substitui cMun/xMun/xUF por estes literais sempre que cPais nao for o Brasil
// -- e tambem os usa como guarda para nao tentar normalizar um municipio que
// nao existe na tabela do IBGE (:260-263).
//
// E o padrao "valor especial" que o CLAUDE.md destaca: emitir o codigo real do
// municipio no lugar de CMunExterior produz documento rejeitado, e nada no
// build acusa.
const (
	CMunExterior = 9999999    // CMUN_EXTERIOR
	XMunExterior = "EXTERIOR" // XMUN_EXTERIOR
	UFExterior   = "EX"       // UF_EXTERIOR
)
