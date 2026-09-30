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

package nfag

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Modelo de dados da NFAg. Porte de ACBrNFAg.Classes.pas, campo a campo,
// com os nomes exatos das tags do leiaute nos comentarios.
//
// Diferencas estruturais em relacao a NFGas (sao do proprio leiaute):
// o det e ACHATADO (gTarif/prod/imposto/gProcRef direto, sem gNormal nem
// gAgregadora); o imposto NAO tem ICMS (so IBSCBS/PIS/COFINS/retTrib e as
// taxas TFS/TFU); e ha grupos proprios de agua: ligacao, gFatConjunto,
// gQualiAgua e infPAA.

// NFAg e o documento fiscal. Porte de TNFAg.
type NFAg struct {
	InfNFAg      InfNFAg       // infNFAg
	Ide          Ide           // ide
	Emit         Emit          // emit
	Dest         Dest          // dest
	Ligacao      Ligacao       // ligacao
	GSub         GSub          // gSub
	GMed         []GMed        // gMed (0..99)
	GFatConjunto GFatConjunto  // gFatConjunto
	Det          []Det         // det (1..990)
	Total        Total         // total
	PgtoVinc     rtc.PgtoVinc  // pgtoVinc
	GFat         GFat          // gFat
	GAgencia     GAgencia      // gAgencia
	GQualiAgua   GQualiAgua    // gQualiAgua
	AutXML       []AutXML      // autXML (0..N)
	InfAdic      InfAdic       // infAdic
	InfPAA       InfPAA        // infPAA
	InfRespTec   InfRespTec    // gRespTec -- a tag NAO se chama infRespTec
	InfNFAgSupl  InfNFAgSupl   // infNFAgSupl
	ProcNFAg     pcn.ProcDFe   // protNFAg/infProt, quando o XML e processado
	Signature    pcn.Signature // Signature
}

// InfNFAg carrega os atributos do elemento infNFAg. Porte de TinfNFAg.
type InfNFAg struct {
	ID     string  // atributo Id, com o literal "NFAG" na frente da chave
	Versao float64 // atributo versao
}

// Ide e o grupo de identificacao do documento. Porte de TIde.
type Ide struct {
	CUF          int                    // cUF
	TpAmb        pcn.TipoAmbiente       // tpAmb
	Modelo       int                    // mod -- sempre 75
	Serie        int                    // serie
	NNF          int                    // nNF
	CNF          int                    // cNF
	CDV          int                    // cDV
	DhEmi        time.Time              // dhEmi
	TpEmis       pcn.TipoEmissao        // tpEmis
	NSiteAutoriz SiteAutorizador        // nSiteAutoriz
	CMunFG       int                    // cMunFG
	FinNFAg      FinalidadeNFAg         // finNFAg
	TpFat        TpFat                  // tpFat
	VerProc      string                 // verProc
	DhCont       time.Time              // dhCont
	XJust        string                 // xJust
	GCompraGov   rtc.GCompraGovReduzido // gCompraGov
	TpPagAnt     pcn.TpPagAnt           // tpPagAnt
}

// Endereco e o endereco usado por emit, dest e enderCorresp.
// Porte de TEndereco (a NFAg NAO tem cPais/xPais, nem na classe).
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

// Dest e o grupo do destinatario. Porte de TDest. Diferente da NFGas,
// idOutros e um campo proprio (nao ha idEstrangeiro).
type Dest struct {
	XNome          string    // xNome
	CNPJCPF        string    // CNPJ ou CPF
	IDOutros       string    // idOutros -- so lido quando CNPJCPF vazio
	IndIEDest      IndIEDest // indIEDest -- so no .ini, NAO existe no XML
	IE             string    // IE
	IM             string    // IM
	CNIS           string    // cNIS
	NB             string    // NB
	XNomeAdicional string    // xNomeAdicional
	EnderDest      Endereco  // enderDest
}

// Ligacao e o grupo da ligacao de agua/esgoto. Porte de Tligacao.
type Ligacao struct {
	IDLigacao         string    // idLigacao
	IDCodCliente      string    // idCodCliente
	TpLigacao         TpLigacao // tpLigacao
	LatGPS            string    // latGPS
	LongGPS           string    // longGPS
	CodRoteiroLeitura string    // codRoteiroLeitura
}

// GSub e o grupo de substituicao. Porte de TgSub -- a NFAg NAO tem o gNF
// da nota substituida; so a chave e o motivo.
type GSub struct {
	ChNFAg string // chNFAg
	MotSub MotSub // motSub
}

// GMed e um medidor. Porte de TgMedCollectionItem -- diferente da NFGas,
// aqui so ha identificacao e datas (as leituras ficam no gMedida do item).
type GMed struct {
	NMed      int       // atributo @nMed
	IDMedidor string    // idMedidor
	DMedAnt   time.Time // dMedAnt
	DMedAtu   time.Time // dMedAtu
}

// GFatConjunto referencia a NFAg de faturamento em conjunto.
// Porte de TgFatConjunto.
type GFatConjunto struct {
	ChNFAgFat string // chNFAgFat
}

// GTarif e uma faixa tarifaria do item. Porte de TgTarifCollectionItem --
// a NFAg NAO tem vTarifAplic.
type GTarif struct {
	DIniTarif   time.Time   // dIniTarif
	DFimTarif   time.Time   // dFimTarif
	NAto        string      // nAto
	AnoAto      int         // anoAto
	TpFaixaCons TpFaixaCons // tpFaixaCons
}

// GMedida e a medida registrada. Porte de TgMedida -- na NFAg as leituras
// anterior/atual, a constante e o valor medido ficam AQUI, nao no gMed.
type GMedida struct {
	TpGrMed      TpGrMed // tpGrMed
	NUnidConsumo string  // nUnidConsumo
	VUnidConsumo float64 // vUnidConsumo
	UMed         UMedFat // uMed
	VMedAnt      float64 // vMedAnt
	VMedAtu      float64 // vMedAtu
	VConst       float64 // vConst
	VMed         float64 // vMed
}

// GMedicao e o grupo de medicao do produto. Porte de TgMedicao -- sem o
// nContrat e o xMotNaoLeitura da NFGas.
type GMedicao struct {
	NMed            int             // nMed
	GMedida         GMedida         // gMedida
	TpMotNaoLeitura TpMotNaoLeitura // tpMotNaoLeitura
}

// Prod e o produto do item. Porte de TProd.
//
// DIVERGENCIA DE TIPO em dois campos, ambas contra o ACBr e a favor do XSD
// (mesma decisao tomada na NFGas, com teste):
//
//   - CClass e string: o leitor do ACBr le com tcStr, mas a classe declara
//     Integer e o gerador grava com tcInt -- os zeros a esquerda do codigo
//     de 7 posicoes so sobrevivem porque o tcInt faz PadLeft(7).
//   - QFaturada e float64: o leitor le com tcDe4, mas a classe declara
//     Integer, ARREDONDANDO a quantidade faturada.
type Prod struct {
	IndOrigemQtd   IndOrigemQtd           // indOrigemQtd
	GMedicao       GMedicao               // gMedicao
	CProd          string                 // cProd
	XProd          string                 // xProd
	CClass         string                 // cClass -- ver nota acima
	TpCategoria    TpCategoria            // tpCategoria
	XCategoria     string                 // xCategoria
	QEconomias     string                 // qEconomias
	UMed           UMedFat                // uMed
	QFaturada      float64                // qFaturada -- ver nota acima
	VItem          float64                // vItem
	FatorPoluicao  float64                // fatorPoluicao
	VProd          float64                // vProd
	IndDevolucao   pcn.IndicadorEx        // indDevolucao -- IndicadorEx: tiSim e o ordinal ZERO de TIndicador (ver nfgas)
	GPagAntecipado rtc.GPagAntecipadoProd // gPagAntecipado
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
// Atencao a grafia da tag de COFINS na LEITURA: vRetCofins; na GERACAO o
// ACBr grava vRetCOFINS (assimetria do proprio ACBr -- ver xml_writer.go).
type RetTrib struct {
	VRetPIS    float64 // vRetPIS
	VRetCOFINS float64 // vRetCofins / vRetCOFINS
	VRetCSLL   float64 // vRetCSLL
	VBCIRRF    float64 // vBCIRRF -- GERADO pelo ACBr, mas nunca LIDO
	VIRRF      float64 // vIRRF
}

// TFS e a Taxa de Fiscalizacao e Servicos. Porte de TTFS.
type TFS struct {
	VBCTFS float64 // vBCTFS
	PTFS   float64 // pTFS
	VTFS   float64 // vTFS
}

// TFU e a Taxa de Fiscalizacao e Uso. Porte de TTFU.
type TFU struct {
	VBCTFU float64 // vBCTFU
	PTFU   float64 // pTFU
	VTFU   float64 // vTFU
}

// Imposto agrupa os tributos do item. Porte de TImposto -- a NFAg NAO tem
// ICMS nem orig/indSemCST.
type Imposto struct {
	IBSCBS  rtc.IBSCBS // IBSCBS
	PIS     PIS        // PIS
	COFINS  COFINS     // COFINS
	RetTrib RetTrib    // retTrib
	TFS     TFS        // TFS
	TFU     TFU        // TFU
}

// GProc referencia um processo administrativo ou judicial.
// Porte de TgProcCollectionItem.
type GProc struct {
	TpProc    TpProc // tpProc
	NProcesso string // nProcesso
}

// GProcRef e o grupo de item referente a processo. Porte de TgProcRef.
type GProcRef struct {
	VItem        float64         // vItem
	QFaturada    float64         // qFaturada
	VProd        float64         // vProd
	IndDevolucao pcn.IndicadorEx // indDevolucao
	GProc        []GProc         // gProc (0..N)
}

// Det e um item do documento. Porte de TDetCollectionItem -- o corpo e
// ACHATADO (sem gNormal/gAgregadora).
type Det struct {
	NItem     int      // atributo @nItem
	ChNFAgAnt string   // atributo @chNFAgAnt
	NItemAnt  int      // atributo @nItemAnt
	GTarif    []GTarif // gTarif (0..6)
	Prod      Prod     // prod
	Imposto   Imposto  // imposto
	GProcRef  GProcRef // gProcRef
	InfAdProd string   // infAdProd
}

// Total e o grupo de totais do documento. Porte de TTotal -- sem ICMSTot;
// os campos de vRetTribTot sao achatados aqui, como no ACBr.
type Total struct {
	VProd      float64       // vProd
	VRetPIS    float64       // vRetTribTot/vRetPIS
	VRetCOFINS float64       // vRetTribTot/vRetCofins
	VRetCSLL   float64       // vRetTribTot/vRetCSLL
	VIRRF      float64       // vRetTribTot/vIRRF
	VCOFINS    float64       // vCOFINS
	VPIS       float64       // vPIS
	VTFS       float64       // vTFS
	VTFU       float64       // vTFU
	VNF        float64       // vNF
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
	DApresFat    time.Time // dApresFat
	DProxLeitura time.Time // dProxLeitura
	NFat         string    // nFat
	CodBarras    string    // codBarras
	CodDebAuto   string    // codDebAuto
	CodBanco     string    // codBanco
	CodAgencia   string    // codAgencia
	EnderCorresp Endereco  // enderCorresp
	GPIX         GPIX      // gPIX
}

// GCons e um mes do historico de consumo. Porte de TgConsCollectionItem --
// qtdDias e STRING (o leitor le com tcStr) e o valor faturado e volFat.
type GCons struct {
	CompetFat time.Time // CompetFat, no XML no formato AAAAMM
	UMed      UMedFat   // uMed
	QtdDias   string    // qtdDias
	MedDiaria float64   // medDiaria
	Consumo   float64   // consumo
	VolFat    float64   // volFat
}

// GHistCons e um bloco do historico de consumo.
// Porte de TgHistConsCollectionItem.
type GHistCons struct {
	XHistorico string  // xHistorico
	MedMensal  float64 // medMensal
	GCons      []GCons // gCons (0..N)
}

// GAgencia e o grupo da agencia reguladora. Porte de TgAgencia -- os campos
// diferem por completo da NFGas (selos e indicadores de qualidade).
type GAgencia struct {
	Econ              string      // econ
	EconAcumulada     string      // econAcumulada
	SPrestador        string      // sPrestador
	DEmissSelo        time.Time   // dEmissSelo
	SRegulador        string      // sRegulador
	NAgenciaAtend     string      // nAgenciaAtend
	EnderAgenciaAtend string      // enderAgenciaAtend
	GHistCons         []GHistCons // gHistCons (0..5)
}

// GAnalise e um item analisado da qualidade da agua.
// Porte de TgAnaliseCollectionItem (todos os campos sao string no leiaute).
type GAnalise struct {
	XItemAnalisado    string // xItemAnalisado
	NAmostraMinima    string // nAmostraMinima
	NAmostraAnalisada string // nAmostraAnalisada
	NAmostraFPadrao   string // nAmostraFPadrao
	NAmostraDPadrao   string // nAmostraDPadrao
	NMediaMensal      string // nMediaMensal
	XValorReferencia  string // xValorReferencia
}

// GQualiAgua e o grupo de qualidade da agua. Porte de TgQualiAgua.
type GQualiAgua struct {
	CompetAnalise time.Time  // CompetAnalise, no XML no formato AAAAMM
	GAnalise      []GAnalise // gAnalise (0..10)
	Conclusao     string     // Conclusao
	CProcesso     string     // cProcesso
	SistemaAbast  string     // SistemaAbast
}

// AutXML e um autorizado a baixar o XML. Porte de TautXMLCollectionItem.
type AutXML struct {
	CNPJCPF string // CNPJ ou CPF
}

// InfAdic e o grupo de informacoes adicionais. Porte de TInfAdic.
//
// InfCpl e slice porque o leiaute admite ate 5 ocorrencias; o leitor do
// ACBr le apenas a PRIMEIRA (TODO aberto no proprio fonte), e o porte
// reproduz isso.
type InfAdic struct {
	InfAdFisco string   // infAdFisco
	InfCpl     []string // infCpl (0..5, apenas a primeira e lida)
}

// InfPAA identifica o Prestador de Apoio ao Abastecimento.
// Porte de TInfPAA.
type InfPAA struct {
	CNPJPAA string // CNPJPAA
}

// InfRespTec e o grupo do responsavel tecnico. Porte de TinfRespTec.
// A TAG no XML e gRespTec; infRespTec e o nome do campo e da secao no .ini.
type InfRespTec struct {
	CNPJ     string // CNPJ
	XContato string // xContato
	Email    string // email
	Fone     string // fone
	IDCSRT   int    // idCSRT
	HashCSRT string // hashCSRT
}

// InfNFAgSupl carrega as informacoes suplementares do documento.
// Porte de TinfNFAgSupl.
type InfNFAgSupl struct {
	QrCodNFAg string // qrCodNFAg
}

// ---------------------------------------------------------------------------
// Acesso de conveniencia
// ---------------------------------------------------------------------------

// ChaveAcesso devolve a chave de 44 posicoes, sem o literal "NFAG" que
// precede o atributo Id. Porte de TNotaFiscal.NumID.
func (n *NFAg) ChaveAcesso() string {
	if n == nil {
		return ""
	}
	return pcn.RemoverLiteralChave(n.InfNFAg.ID)
}

// Confirmada informa se o protocolo indica documento confirmado.
// Porte de TNotaFiscal.Confirmada.
func (n *NFAg) Confirmada() bool {
	return n != nil && cStatConfirmada[n.ProcNFAg.CStat]
}

// Processada informa se o protocolo indica documento processado.
// Porte de TNotaFiscal.Processada.
func (n *NFAg) Processada() bool {
	return n != nil && cStatProcessado[n.ProcNFAg.CStat]
}

// Cancelada informa se o protocolo indica documento cancelado.
// Porte de TNotaFiscal.Cancelada.
func (n *NFAg) Cancelada() bool {
	return n != nil && cStatCancelada[n.ProcNFAg.CStat]
}
