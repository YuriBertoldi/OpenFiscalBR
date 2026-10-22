// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package cte

import (
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Structs de dados do CT-e. Porte de ACBrCTe.Classes.pas.
//
// Os SEIS MODAIS -- rodoviario, aereo, aquaviario, ferroviario, dutoviario
// e multimodal -- vivem em classes_modais.go e sao preenchidos na etapa 2b.
//
// Os grupos da Reforma Tributaria NAO sao redeclarados aqui: o ACBr os
// declara no topo de ACBrCTe.Classes.pas, mas sao os MESMOS tipos que o
// packages/rtc ja porta do ACBrDFe.RTC. Catorze classes reaproveitadas
// (TgIBSCBS, TIBSCBS, TgCompraGovReduzido e companhia).
//
// Enums compartilhados com o MDF-e -- Modal, UnidMed, UnidTransp e
// UnidCarga -- vivem em packages/pcn, porque e de la que o ACBr os declara
// (ACBrDFe.Conversao.pas) e porque dois componentes os usam.
//
// Convencao de nome: campo fiscal mantem o nome do leiaute (cUF, nCT, vBC),
// porque e ele que mapeia para a tag. O comentario ao lado diz a tag quando
// o nome Go precisou divergir.

// ---------------------------------------------------------------------------
// Identificacao
// ---------------------------------------------------------------------------

// InfCTe e o grupo raiz do documento. Porte de TInfCTe.
type InfCTe struct {
	ID     string  // atributo Id
	Versao float64 // atributo versao
}

// Endereco e o endereco de um participante. Porte de TEndereco.
//
// Tem cPais e xPais, que o EnderEmit NAO tem -- o emitente e sempre
// nacional. Sao dois tipos no ACBr pela mesma razao, e mante-los separados
// evita emitir pais no grupo que nao o aceita.
type Endereco struct {
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	CEP     int    // CEP
	UF      string // UF
	CPais   int    // cPais
	XPais   string // xPais
}

// Toma03 identifica o tomador quando ele e um dos participantes.
// Porte de TToma03.
//
// O nome vem do leiaute: o grupo `toma03` aceita os valores 0 a 3 --
// remetente, expedidor, recebedor e destinatario. Quando o tomador e
// OUTRO, o grupo usado e o toma4.
type Toma03 struct {
	Toma Tomador // toma
}

// Toma4 identifica o tomador quando ele nao e nenhum dos participantes.
// Porte de TToma4.
type Toma4 struct {
	Toma      Tomador  // toma -- sempre TmOutros neste grupo
	CNPJCPF   string   // CNPJ ou CPF
	IE        string   // IE
	XNome     string   // xNome
	XFant     string   // xFant
	Fone      string   // fone
	EnderToma Endereco // enderToma
	Email     string   // email
}

// InfPercurso e uma UF do percurso. Porte de TinfPercursoCollectionItem.
type InfPercurso struct {
	UFPer string // UFPer
}

// Ide e o grupo de identificacao do CT-e. Porte de TIde.
//
// E o grupo que decide quase tudo do documento: o MODAL escolhe qual das
// seis arvores de infModal sai, o TPSERV separa CT-e de GTV-e, e o par
// toma03/toma4 e MUTUAMENTE EXCLUSIVO -- preencher os dois gera documento
// que a SEFAZ recusa.
type Ide struct {
	CUF            int                    // cUF
	CCT            int                    // cCT -- codigo numerico
	CFOP           int                    // CFOP
	NatOp          string                 // natOp
	ForPag         FormaPagamento         // forPag
	Modelo         int                    // mod
	Serie          int                    // serie
	NCT            int                    // nCT
	DhEmi          time.Time              // dhEmi
	TpImp          pcn.TipoImpressao      // tpImp
	TpEmis         pcn.TipoEmissao        // tpEmis
	CDV            int                    // cDV
	TpAmb          pcn.TipoAmbiente       // tpAmb
	TpCTe          TipoCTe                // tpCTe
	ProcEmi        pcn.ProcessoEmissao    // procEmi
	VerProc        string                 // verProc
	IndGlobalizado pcn.Indicador          // indGlobalizado
	RefCTe         string                 // refCTe
	CMunEnv        int                    // cMunEnv
	XMunEnv        string                 // xMunEnv
	UFEnv          string                 // UFEnv
	Modal          pcn.Modal              // modal
	TpServ         TipoServico            // tpServ
	CMunIni        int                    // cMunIni
	XMunIni        string                 // xMunIni
	UFIni          string                 // UFIni
	CMunFim        int                    // cMunFim
	XMunFim        string                 // xMunFim
	UFFim          string                 // UFFim
	Retira         Retira                 // retira
	XDetRetira     string                 // xDetRetira
	IndIEToma      pcn.IndIEDest          // indIEToma
	DhSaidaOrig    time.Time              // dhSaidaOrig
	DhChegadaDest  time.Time              // dhChegadaDest
	Toma03         Toma03                 // toma03
	Toma4          Toma4                  // toma4
	InfPercurso    []InfPercurso          // infPercurso
	DhCont         time.Time              // dhCont
	XJust          string                 // xJust
	GCompraGov     rtc.GCompraGovReduzido // gCompraGov
}

// ---------------------------------------------------------------------------
// Complemento
// ---------------------------------------------------------------------------

// Pass e uma passagem do fluxo de carga. Porte de TPassCollectionItem.
type Pass struct {
	XPass string // xPass
}

// Fluxo e o fluxo de carga. Porte de TFluxo.
type Fluxo struct {
	XOrig string // xOrig
	Pass  []Pass // pass
	XDest string // xDest
	XRota string // xRota
}

// SemData, ComData e NoPeriodo sao as TRES formas de informar a data de
// entrega, mutuamente exclusivas. Portes de TSemData, TComData e
// TNoPeriodo.
//
// Cada uma repete o tpPer. Nao e redundancia do porte: no leiaute o tipo
// vive DENTRO do grupo escolhido, e e o que diz a quem le qual grupo
// esperar.
type SemData struct {
	TpPer TipoDataPeriodo // tpPer
}

// ComData informa uma data unica.
type ComData struct {
	TpPer TipoDataPeriodo // tpPer
	DProg time.Time       // dProg
}

// NoPeriodo informa um intervalo de datas.
type NoPeriodo struct {
	TpPer TipoDataPeriodo // tpPer
	DIni  time.Time       // dIni
	DFim  time.Time       // dFim
}

// SemHora, ComHora e NoInter sao as tres formas de informar a hora de
// entrega. Portes de TSemHora, TComHora e TNoInter.
type SemHora struct {
	TpHor TipoHorarioIntervalo // tpHor
}

// ComHora informa uma hora unica.
type ComHora struct {
	TpHor TipoHorarioIntervalo // tpHor
	HProg time.Time            // hProg
}

// NoInter informa um intervalo de horas.
type NoInter struct {
	TpHor TipoHorarioIntervalo // tpHor
	HIni  time.Time            // hIni
	HFim  time.Time            // hFim
}

// Entrega e a previsao de entrega. Porte de TEntrega.
//
// Carrega os SEIS subgrupos mais dois campos que dizem QUAL deles vale --
// e assim no ACBr. Os campos TipoData e TipoHora nao sao tags: sao o
// seletor que o gerador consulta para decidir o que emitir.
type Entrega struct {
	// TipoData e TipoHora escolhem o subgrupo. Nao saem no XML -- sao
	// properties do ACBr com esse mesmo nome, nao tags do leiaute.
	TipoData TipoDataPeriodo      // TipoData -- seletor, nao e tag
	TipoHora TipoHorarioIntervalo // TipoHora -- seletor, nao e tag

	SemData   SemData   // semData
	ComData   ComData   // comData
	NoPeriodo NoPeriodo // noPeriodo
	SemHora   SemHora   // semHora
	ComHora   ComHora   // comHora
	NoInter   NoInter   // noInter
}

// ObsCont e uma observacao de interesse do contribuinte.
// Porte de TObsContCollectionItem.
type ObsCont struct {
	XCampo string // xCampo (atributo)
	XTexto string // xTexto
}

// Compl e o grupo de informacoes complementares. Porte de TCompl.
//
// ObsFisco e do MESMO tipo do ObsCont: no ACBr, `TObsFiscoCollection` e um
// ALIAS de `TObsContCollection` (Classes.pas:621), nao um tipo proprio. O
// que separa os dois e so a tag em que saem.
type Compl struct {
	XCaracAd  string    // xCaracAd
	XCaracSer string    // xCaracSer
	XEmi      string    // xEmi
	Fluxo     Fluxo     // fluxo
	Entrega   Entrega   // Entrega
	OrigCalc  string    // origCalc
	DestCalc  string    // destCalc
	XObs      string    // xObs
	ObsCont   []ObsCont // ObsCont
	ObsFisco  []ObsCont // ObsFisco
}

// ---------------------------------------------------------------------------
// Participantes
// ---------------------------------------------------------------------------

// EnderEmit e o endereco do emitente. Porte de TEnderEmit.
//
// NAO tem cPais nem xPais: o emitente de CT-e e sempre nacional. Tem fone,
// que o Endereco generico nao tem.
type EnderEmit struct {
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	CEP     int    // CEP
	UF      string // UF
	Fone    string // fone
}

// Emit e o emitente do CT-e. Porte de TEmit.
//
// So aceita CNPJ -- nao ha CNPJCPF aqui, ao contrario dos demais
// participantes. Transportador e sempre pessoa juridica.
type Emit struct {
	CNPJ      string    // CNPJ
	IE        string    // IE
	IEST      string    // IEST
	XNome     string    // xNome
	XFant     string    // xFant
	EnderEmit EnderEmit // enderEmit
	CRT       CRT       // CRT
	ISUFEmit  string    // ISUFEmit
}

// Toma e o tomador do servico no CT-e OS. Porte de TToma.
//
// Nao confundir com Toma03/Toma4 do Ide: aqueles IDENTIFICAM qual
// participante e o tomador no CT-e normal; este e o grupo completo do
// tomador, usado quando o documento e CT-e OS.
type Toma struct {
	CNPJCPF   string        // CNPJ ou CPF
	IE        string        // IE
	XNome     string        // xNome
	XFant     string        // xFant
	Fone      string        // fone
	EnderToma Endereco      // enderToma
	Email     string        // email
	Toma      Tomador       // toma
	IndIEToma pcn.IndIEDest // indIEToma
	ISUF      string        // ISUF
}

// LocColeta e o local de coleta da carga. Porte de TLocColeta.
//
// ESTE GRUPO NAO SAI EM XML. Nem o XmlWriter nem o XmlReader do CT-e
// mencionam `locColeta`, e ele nao existe em XSD nenhum -- vive so no .ini
// (IniWriter.pas:572). E grupo morto do ACBr, mantido aqui por fidelidade
// e porque a etapa do .ini vai precisar dele.
//
// E um endereco REDUZIDO -- sem CEP, sem pais e sem fone. Por isso e tipo
// proprio e nao reusa o Endereco generico.
type LocColeta struct {
	CNPJCPF string // CNPJ ou CPF
	XNome   string // xNome
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
}

// Rem e o remetente da carga. Porte de TRem.
type Rem struct {
	CNPJCPF   string    // CNPJ ou CPF
	IE        string    // IE
	XNome     string    // xNome
	XFant     string    // xFant
	Fone      string    // fone
	EnderReme Endereco  // enderReme
	Email     string    // email
	LocColeta LocColeta // locColeta -- so no .ini; nao existe em XSD nenhum
}

// Exped e o expedidor da carga. Porte de TExped.
//
// Nao tem xFant nem locColeta, ao contrario do remetente.
type Exped struct {
	CNPJCPF    string   // CNPJ ou CPF
	IE         string   // IE
	XNome      string   // xNome
	Fone       string   // fone
	EnderExped Endereco // enderExped
	Email      string   // email
}

// Receb e o recebedor da carga. Porte de TReceb.
type Receb struct {
	CNPJCPF    string   // CNPJ ou CPF
	IE         string   // IE
	XNome      string   // xNome
	Fone       string   // fone
	EnderReceb Endereco // enderReceb
	Email      string   // email
}

// LocEnt e o local de entrega da carga. Porte de TLocEnt.
//
// Mesma forma reduzida do LocColeta, e o MESMO destino: nao sai em XML,
// so no .ini (IniWriter.pas:580).
type LocEnt struct {
	CNPJCPF string // CNPJ ou CPF
	XNome   string // xNome
	XLgr    string // xLgr
	Nro     string // nro
	XCpl    string // xCpl
	XBairro string // xBairro
	CMun    int    // cMun
	XMun    string // xMun
	UF      string // UF
}

// Dest e o destinatario da carga. Porte de TDest.
//
// Tem ISUF (inscricao na SUFRAMA), que remetente, expedidor e recebedor
// nao tem.
type Dest struct {
	CNPJCPF   string   // CNPJ ou CPF
	IE        string   // IE
	XNome     string   // xNome
	Fone      string   // fone
	ISUF      string   // ISUF
	EnderDest Endereco // enderDest
	Email     string   // email
	LocEnt    LocEnt   // locEnt -- so no .ini; nao existe em XSD nenhum
}

// ---------------------------------------------------------------------------
// Valores da prestacao
// ---------------------------------------------------------------------------

// Comp e um componente do valor da prestacao.
// Porte de TCompCollectionItem.
type Comp struct {
	XNome string  // xNome
	VComp float64 // vComp
}

// VPrest e o grupo de valores da prestacao. Porte de TVPrest.
type VPrest struct {
	VTPrest float64 // vTPrest
	VRec    float64 // vRec
	Comp    []Comp  // Comp
}

// ---------------------------------------------------------------------------
// ICMS
// ---------------------------------------------------------------------------
//
// O CT-e tem SETE grupos de ICMS mutuamente exclusivos, e o ACBr declara
// um tipo para cada -- a estrutura e diferente do `det` da NF-e, onde um
// struct unico carrega todos os campos de todos os CST.
//
// Qual sai e decidido pelo campo SituTrib do grupo ICMS.

// CST00 e o ICMS tributado integralmente. Porte de TCST00.
type CST00 struct {
	CST   pcn.CSTIcms // CST
	VBC   float64     // vBC
	PICMS float64     // pICMS
	VICMS float64     // vICMS
}

// CST20 e o ICMS com reducao de base. Porte de TCST20.
type CST20 struct {
	CST        pcn.CSTIcms // CST
	PRedBC     float64     // pRedBC
	VBC        float64     // vBC
	PICMS      float64     // pICMS
	VICMS      float64     // vICMS
	VICMSDeson float64     // vICMSDeson
	CBenef     string      // cBenef
}

// CST45 e o ICMS isento, nao tributado ou diferido. Porte de TCST45.
//
// O numero do grupo e 45, mas o CST que ele carrega pode ser 40, 41 ou 51:
// o leiaute agrupou tres situacoes num grupo so.
type CST45 struct {
	CST        pcn.CSTIcms // CST
	VICMSDeson float64     // vICMSDeson
	CBenef     string      // cBenef
}

// CST60 e o ICMS cobrado por substituicao tributaria. Porte de TCST60.
type CST60 struct {
	CST        pcn.CSTIcms // CST
	VBCSTRet   float64     // vBCSTRet
	VICMSSTRet float64     // vICMSSTRet
	PICMSSTRet float64     // pICMSSTRet
	VCred      float64     // vCred
	VICMSDeson float64     // vICMSDeson
	CBenef     string      // cBenef
}

// CST90 e o ICMS outros. Porte de TCST90.
type CST90 struct {
	CST        pcn.CSTIcms // CST
	PRedBC     float64     // pRedBC
	VBC        float64     // vBC
	PICMS      float64     // pICMS
	VICMS      float64     // vICMS
	VCred      float64     // vCred
	VICMSDeson float64     // vICMSDeson
	CBenef     string      // cBenef
}

// ICMSOutraUF e o ICMS devido a outra UF. Porte de TICMSOutraUF.
//
// Os nomes dos campos levam o sufixo OutraUF no PROPRIO leiaute --
// `vBCOutraUF`, `pICMSOutraUF`. Nao e redundancia a eliminar.
type ICMSOutraUF struct {
	CST           pcn.CSTIcms // CST
	PRedBCOutraUF float64     // pRedBCOutraUF
	VBCOutraUF    float64     // vBCOutraUF
	PICMSOutraUF  float64     // pICMSOutraUF
	VICMSOutraUF  float64     // vICMSOutraUF
	VICMSDeson    float64     // vICMSDeson
	CBenef        string      // cBenef
}

// ICMSSN e o ICMS do Simples Nacional. Porte de TICMSSN.
//
// O IndSN e INTEIRO, nao enum -- o leiaute o declara assim e o unico valor
// previsto e 1.
type ICMSSN struct {
	CST   pcn.CSTIcms // CST
	IndSN int         // indSN
}

// ICMS reune os sete grupos. Porte de TICMS.
//
// SituTrib decide qual sai, e os sete sao MUTUAMENTE EXCLUSIVOS. Preencher
// mais de um nao e erro de compilacao -- e o gerador que escolhe.
type ICMS struct {
	// SituTrib escolhe o grupo. Nao e tag propria -- o CST sai DENTRO do
	// grupo escolhido -- e e property do ACBr com esse mesmo nome.
	SituTrib pcn.CSTIcms // SituTrib -- seletor, nao e tag

	ICMS00      CST00       // ICMS00
	ICMS20      CST20       // ICMS20
	ICMS45      CST45       // ICMS45
	ICMS60      CST60       // ICMS60
	ICMS90      CST90       // ICMS90
	ICMSOutraUF ICMSOutraUF // ICMSOutraUF
	ICMSSN      ICMSSN      // ICMSSN
}

// ICMSUFFim e o ICMS da partilha para a UF de termino. Porte de
// TICMSUFFim.
type ICMSUFFim struct {
	VBCUFFim   float64 // vBCUFFim
	PFCPUFFim  float64 // pFCPUFFim
	PICMSUFFim float64 // pICMSUFFim
	PICMSInter float64 // pICMSInter
	// PICMSInterPart so existe no leiaute 2.00: foi removido na 3.00, e o
	// Gerar_ICMSUFFim do ACBr (XmlWriter.pas:1665) nunca o emite. Sai so
	// no .ini.
	PICMSInterPart float64 // pICMSInterPart -- so no .ini e no leiaute 2.00
	VFCPUFFim      float64 // vFCPUFFim
	VICMSUFFim     float64 // vICMSUFFim
	VICMSUFIni     float64 // vICMSUFIni
}

// InfTribFed sao os tributos federais. Porte de TinfTribFed.
type InfTribFed struct {
	VPIS    float64 // vPIS
	VCOFINS float64 // vCOFINS
	VIR     float64 // vIR
	VINSS   float64 // vINSS
	VCSLL   float64 // vCSLL
}

// Imp e o grupo de impostos. Porte de TImp.
type Imp struct {
	ICMS       ICMS       // ICMS
	VTotTrib   float64    // vTotTrib
	InfAdFisco string     // infAdFisco
	ICMSUFFim  ICMSUFFim  // ICMSUFFim
	InfTribFed InfTribFed // infTribFed

	// Reforma Tributaria.
	//
	// O rtc.IBSCBS e SUPERCONJUNTO do TIBSCBS do CT-e: o original tem
	// cinco properties (CST, cClassTrib, indDoacao, gIBSCBS,
	// gEstornoCred) e o tipo do rtc traz mais cinco grupos --
	// GIBSCBSMono, GTransfCred, GAjusteCompet, GCredPresOper e
	// GCredPresIBSZFM -- que o CT-e NAO aceita e que o gerador vai
	// ignorar. Reusar ainda e melhor que redeclarar, mas preencher um
	// dos cinco aqui nao produz efeito nem aviso.
	IBSCBS  rtc.IBSCBS // IBSCBS
	VTotDFe float64    // vTotDFe
}

// ---------------------------------------------------------------------------
// Carga e unidades
// ---------------------------------------------------------------------------

// InfQ e uma quantidade da carga. Porte de TInfQCollectionItem.
type InfQ struct {
	CUnid  pcn.UnidMed // cUnid
	TpMed  string      // tpMed
	QCarga float64     // qCarga
}

// InfCarga e o grupo de informacoes da carga. Porte de TInfCarga.
type InfCarga struct {
	VCarga      float64 // vCarga
	ProPred     string  // proPred
	XOutCat     string  // xOutCat
	InfQ        []InfQ  // infQ
	VCargaAverb float64 // vCargaAverb
}

// Lacre e o numero de um lacre. Porte de TLacreCollectionItem.
//
// O mesmo tipo serve a lacUnidCarga e a lacUnidTransp: no ACBr os dois sao
// ALIAS de TLacreCollection (Classes.pas:1156 e :1158).
type Lacre struct {
	NLacre string // nLacre
}

// InfUnidCarga e uma unidade de carga. Porte de
// TinfUnidCargaCollectionItem.
type InfUnidCarga struct {
	TpUnidCarga  pcn.UnidCarga // tpUnidCarga
	IDUnidCarga  string        // idUnidCarga
	LacUnidCarga []Lacre       // lacUnidCarga
	QtdRat       float64       // qtdRat
}

// InfUnidTransp e uma unidade de transporte. Porte de
// TinfUnidTranspCollectionItem.
//
// Contem unidades de CARGA dentro: um caminhao leva containers, e cada
// container leva lacres proprios.
type InfUnidTransp struct {
	TpUnidTransp  pcn.UnidTransp // tpUnidTransp
	IDUnidTransp  string         // idUnidTransp
	LacUnidTransp []Lacre        // lacUnidTransp
	InfUnidCarga  []InfUnidCarga // infUnidCarga
	QtdRat        float64        // qtdRat
}

// ---------------------------------------------------------------------------
// Documentos transportados
// ---------------------------------------------------------------------------

// InfNF e uma nota fiscal em papel transportada.
// Porte de TInfNFCollectionItem.
type InfNF struct {
	NRoma         string          // nRoma
	NPed          string          // nPed
	Modelo        ModeloNF        // mod
	Serie         string          // serie
	NDoc          string          // nDoc
	DEmi          time.Time       // dEmi
	VBC           float64         // vBC
	VICMS         float64         // vICMS
	VBCST         float64         // vBCST
	VST           float64         // vST
	VProd         float64         // vProd
	VNF           float64         // vNF
	NCFOP         int             // nCFOP
	NPeso         float64         // nPeso
	PIN           string          // PIN
	DPrev         time.Time       // dPrev
	InfUnidTransp []InfUnidTransp // infUnidTransp
	InfUnidCarga  []InfUnidCarga  // infUnidCarga
}

// InfNFe e uma NF-e transportada. Porte de TInfNFeCollectionItem.
type InfNFe struct {
	Chave         string          // chave
	PIN           string          // PIN
	DPrev         time.Time       // dPrev
	InfUnidTransp []InfUnidTransp // infUnidTransp
	InfUnidCarga  []InfUnidCarga  // infUnidCarga
}

// InfOutros e outro documento transportado.
// Porte de TInfOutrosCollectionItem.
type InfOutros struct {
	TpDoc         TipoDocumento   // tpDoc
	DescOutros    string          // descOutros
	NDoc          string          // nDoc
	DEmi          time.Time       // dEmi
	VDocFisc      float64         // vDocFisc
	DPrev         time.Time       // dPrev
	InfUnidTransp []InfUnidTransp // infUnidTransp
	InfUnidCarga  []InfUnidCarga  // infUnidCarga
}

// InfDCe e uma Declaracao de Conteudo eletronica transportada.
// Porte de TInfDCeCollectionItem.
type InfDCe struct {
	Chave string // chave
}

// InfDoc reune os documentos transportados. Porte de TInfDoc.
type InfDoc struct {
	InfNF     []InfNF     // infNF
	InfNFe    []InfNFe    // infNFe
	InfOutros []InfOutros // infOutros
	InfDCe    []InfDCe    // infDCe
}

// Peri e um produto perigoso transportado.
// Porte de TPeriCollectionItem (Classes.pas:1788).
//
// # Dois donos, decididos pela VERSAO do leiaute
//
// O grupo `peri` nao pende sempre do mesmo lugar, e o ACBr troca o dono
// conforme a versao:
//
//	ve200            infCTeNorm.peri   (XmlWriter.pas:1792 -- `if VersaoDF = ve200`)
//	ve300 em diante  aereo.peri        (XmlWriter.pas:3727 -- `if VersaoDF >= ve300`)
//
// Por isso o tipo vive aqui, no classes.go, e nao junto dos modais: os dois
// donos precisam dele. Mas a etapa 2b TEM de declarar `Peri []Peri` dentro
// do Aereo -- sem isso o CT-e aereo 4.00 sai sem produto perigoso nenhum, e
// nada acusa: nao quebra build, nao quebra vet, nao quebra teste.
//
// # A estrutura emitida tambem muda por versao
//
// De ve300 em diante o ACBr emite nONU, qTotEmb e um grupo infTotAP com
// qTotProd e uniAP; na ve200 emite os sete campos legados
// (XmlWriter.pas:2459-2505). Os campos sao os mesmos; o que muda e o
// aninhamento.
//
// # Por que as quantidades sao string
//
// qTotProd, qVolTipo e qTotEmb sao declarados `string` no original e
// escritos com tcStr (XmlWriter.pas:2510). O porte mantem, por fidelidade
// -- mas a razao NAO e "texto livre sempre": na 2.00 o schema aceita texto
// (TString), e na 4.00 o qTotProd e TDec_1104 dentro do infTotAP. Da 3.00
// em diante o valor precisa chegar aqui ja formatado como decimal.
type Peri struct {
	NONU        string // nONU
	XNomeAE     string // xNomeAE
	XClaRisco   string // xClaRisco
	GrEmb       string // grEmb
	QTotProd    string // qTotProd
	QVolTipo    string // qVolTipo
	PontoFulgor string // pontoFulgor
	QTotEmb     string // qTotEmb
	UniAP       UniMed // uniAP
}

// ---------------------------------------------------------------------------
// Documentos anteriores
// ---------------------------------------------------------------------------

// IDDocAntPap e um documento anterior em PAPEL.
// Porte de TIdDocAntPapCollectionItem.
type IDDocAntPap struct {
	TpDoc  TipoDocumentoAnterior // tpDoc
	Serie  string                // serie
	Subser string                // subser
	NDoc   string                // nDoc
	DEmi   time.Time             // dEmi
}

// IDDocAntEle e um documento anterior ELETRONICO.
// Porte de TIdDocAntEleCollectionItem.
//
// Tem DOIS campos de chave -- `chave` e `chCTe` -- que o ACBr mantem
// separados. Nao sao sinonimos: o leiaute mudou de nome entre versoes e o
// componente aceita os dois.
type IDDocAntEle struct {
	Chave string // chave
	ChCTe string // chCTe
}

// IDDocAnt agrupa os documentos anteriores de um emitente.
// Porte de TIdDocAntCollectionItem.
type IDDocAnt struct {
	IDDocAntPap []IDDocAntPap // idDocAntPap
	IDDocAntEle []IDDocAntEle // idDocAntEle
}

// EmiDocAnt e o emitente dos documentos anteriores.
// Porte de TEmiDocAntCollectionItem.
type EmiDocAnt struct {
	CNPJCPF  string     // CNPJ ou CPF
	IE       string     // IE
	UF       string     // UF
	XNome    string     // xNome
	IDDocAnt []IDDocAnt // idDocAnt
}

// DocAnt e o grupo de documentos anteriores. Porte de TDocAnt.
type DocAnt struct {
	EmiDocAnt []EmiDocAnt // emiDocAnt
}

// ---------------------------------------------------------------------------
// Seguro e ordem de coleta
// ---------------------------------------------------------------------------

// Seg e uma apolice de seguro da carga. Porte de TSegCollectionItem.
type Seg struct {
	RespSeg RspSeg  // respSeg
	XSeg    string  // xSeg
	NApol   string  // nApol
	NAver   string  // nAver
	VCarga  float64 // vCarga
}

// EmiOCC e o emitente da ordem de coleta. Porte de TEmiOCC.
type EmiOCC struct {
	CNPJ string // CNPJ
	CInt string // cInt
	IE   string // IE
	UF   string // UF
	Fone string // fone
}

// Occ e uma ordem de coleta. Porte de TOccCollectionItem.
type Occ struct {
	Serie  string    // serie
	NOcc   int       // nOcc
	DEmi   time.Time // dEmi
	EmiOcc EmiOCC    // emiOcc
}

// ---------------------------------------------------------------------------
// Cobranca
// ---------------------------------------------------------------------------

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

// Cobr e o grupo de cobranca. Porte de TCobr.
type Cobr struct {
	Fat Fat   // fat
	Dup []Dup // dup
}

// ---------------------------------------------------------------------------
// Substituicao e complementares
// ---------------------------------------------------------------------------

// RefNF e a nota fiscal referenciada na substituicao. Porte de TRefNF.
//
// O `modelo` e STRING e `serie`/`subserie`/`nro` sao INTEIROS -- a mistura
// e do leiaute, nao do porte.
type RefNF struct {
	CNPJCPF  string    // CNPJ ou CPF
	Modelo   string    // mod
	Serie    int       // serie
	Subserie int       // subserie
	Nro      int       // nro
	Valor    float64   // valor
	DEmi     time.Time // dEmi
}

// TomaICMS e a referencia do tomador contribuinte. Porte de TTomaICMS.
//
// Os tres campos sao MUTUAMENTE EXCLUSIVOS: o documento substituido e uma
// NF-e, uma NF em papel ou um CT-e.
type TomaICMS struct {
	RefNFe string // refNFe
	RefNF  RefNF  // refNF
	RefCte string // refCte
}

// TomaNaoICMS e a referencia do tomador nao contribuinte.
// Porte de TTomaNaoICMS.
type TomaNaoICMS struct {
	RefCteAnu string // refCteAnu
}

// InfCteSub e o grupo de substituicao. Porte de TInfCteSub.
type InfCteSub struct {
	ChCte         string        // chCte
	RefCteAnu     string        // refCteAnu
	TomaICMS      TomaICMS      // tomaICMS
	TomaNaoICMS   TomaNaoICMS   // tomaNaoICMS
	IndAlteraToma pcn.Indicador // indAlteraToma
}

// InfGlobalizado e a observacao do CT-e globalizado.
// Porte de TInfGlobalizado.
type InfGlobalizado struct {
	XObs string // xObs
}

// InfCTeMultimodal e um CT-e vinculado ao multimodal.
// Porte de TInfCTeMultimodalCollectionItem.
type InfCTeMultimodal struct {
	ChCTeMultimodal string // chCTeMultimodal
}

// InfServVinc reune os CT-e vinculados ao servico. Porte de TInfServVinc.
type InfServVinc struct {
	InfCTeMultimodal []InfCTeMultimodal // infCTeMultimodal
}

// InfServico descreve o servico prestado no CT-e OS.
// Porte de TInfServico.
type InfServico struct {
	XDescServ string  // xDescServ
	QCarga    float64 // qCarga
}

// InfDocRef e um documento referenciado no CT-e OS.
// Porte de TinfDocRefCollectionItem.
type InfDocRef struct {
	NDoc     string    // nDoc
	Serie    string    // serie
	Subserie string    // subserie
	DEmi     time.Time // dEmi
	VDoc     float64   // vDoc
	ChBPe    string    // chBPe
}

// InfGTVeComp e um componente do valor da GTV-e.
// Porte de TinfGTVeCompCollectionItem.
type InfGTVeComp struct {
	TpComp TpComp  // tpComp
	VComp  float64 // vComp
	XComp  string  // xComp
}

// InfGTVe e uma GTV-e referenciada. Porte de TinfGTVeCollectionItem.
type InfGTVe struct {
	ChCTe string        // chCTe
	Comp  []InfGTVeComp // Comp
}

// VeicNovos e um veiculo novo transportado.
// Porte de TVeicNovosCollectionItem.
type VeicNovos struct {
	Chassi string  // chassi
	CCor   string  // cCor
	XCor   string  // xCor
	CMod   string  // cMod
	VUnit  float64 // vUnit
	VFrete float64 // vFrete
}

// ---------------------------------------------------------------------------
// GTV-e
// ---------------------------------------------------------------------------

// InfEspecie e uma especie de valor transportado na GTV-e.
// Porte de TinfEspecieCollectionItem.
type InfEspecie struct {
	TpEspecie   Especie     // tpEspecie
	VEspecie    float64     // vEspecie
	TpNumerario TpNumerario // tpNumerario
	XMoedaEstr  string      // xMoedaEstr
}

// InfVeiculo e um veiculo da GTV-e.
// Porte de TinfVeiculoCollectionItem.
type InfVeiculo struct {
	Placa string // placa
	UF    string // UF
	RNTRC string // RNTRC
}

// DetGTV e o detalhamento da GTV-e. Porte de TdetGTV.
type DetGTV struct {
	InfEspecie []InfEspecie // infEspecie
	QCarga     float64      // qCarga
	InfVeiculo []InfVeiculo // infVeiculo
}

// ---------------------------------------------------------------------------
// CT-e Simplificado
// ---------------------------------------------------------------------------

// InfNFeTranspParcial e uma NF-e com transporte parcial.
// Porte de TinfNFeTranspParcialCollectionItem.
type InfNFeTranspParcial struct {
	ChNFe string // chNFe
}

// InfDocAntDet e um documento anterior do detalhamento simplificado.
// Porte de TinfDocAntCollectionItem.
//
// O nome Go leva o sufixo Det porque o ACBr tem DUAS colecoes com nomes
// quase iguais: `TIdDocAntCollection` (documentos anteriores do grupo
// docAnt) e `TinfDocAntCollection` (do detalhamento do Simplificado).
// Mante-las com o mesmo nome aqui seria colisao.
type InfDocAntDet struct {
	ChCTe               string                // chCTe
	TpPrest             TpPrest               // tpPrest
	InfNFeTranspParcial []InfNFeTranspParcial // infNFeTranspParcial
}

// Det e um item do detalhamento do CT-e Simplificado.
// Porte de TdetCollectionItem.
type Det struct {
	CMunIni   int            // cMunIni
	XMunIni   string         // xMunIni
	CMunFim   int            // cMunFim
	XMunFim   string         // xMunFim
	VPrest    float64        // vPrest
	VRec      float64        // vRec
	Comp      []Comp         // Comp
	InfNFe    []InfNFe       // infNFe
	InfDocAnt []InfDocAntDet // infdocAnt
}

// Total e o total do CT-e Simplificado. Porte de Ttotal.
type Total struct {
	VTPrest float64 // vTPrest
	VTRec   float64 // vTRec
}

// ---------------------------------------------------------------------------
// Pagamento vinculado
// ---------------------------------------------------------------------------

// Pgto e um pagamento vinculado. Porte de TpgtoCollectionItem.
type Pgto struct {
	TpMeioPgto  string // tpMeioPgto
	CNPJReceb   string // CNPJReceb
	CNPJBasePSP string // CNPJBasePSP
	NPag        int    // nPag
	IDTransacao string // idTransacao
}

// PgtoVinc e o grupo de pagamentos vinculados. Porte de TpgtoVinc.
type PgtoVinc struct {
	Pgto []Pgto // pgto
}

// ---------------------------------------------------------------------------
// Grupos complementares e de anulacao
// ---------------------------------------------------------------------------

// InfCteComp e a referencia do CT-e complementado, chave UNICA.
// Porte de TInfCteComp.
//
// Vale ate a 3.00 INCLUSIVE. O ACBr decide com `if VersaoDF <= ve300`
// (XmlWriter.pas:493 e :554): ate la sai este grupo; da 4.00 em diante sai
// o InfCteComp10, que e colecao.
//
// Dentro dele a TAG muda de nome entre as duas versoes que o usam: `chave`
// na ve200 e `chCTe` da ve300 em diante (XmlWriter.pas:2657-2680). O
// elemento externo e `infCteComp` nos dois casos.
type InfCteComp struct {
	Chave string // chave
}

// InfCteCompItem e a referencia do CT-e complementado, em colecao.
// Porte de TInfCteCompCollectionItem.
//
// Vale da 4.00 em diante. Emitir este grupo num documento 3.00 e rejeicao:
// o schema da 3.00 nao o conhece.
type InfCteCompItem struct {
	ChCTe string // chCTe
}

// InfCteAnu e a referencia do CT-e anulado. Porte de TInfCteAnu.
//
// Sai no MESMO `if VersaoDF <= ve300` do InfCteComp: na 4.00 o ACBr nao o
// emite. O campo continua aqui porque o grupo existe nos leiautes antigos,
// mas o gerador da etapa 4 nao deve escreve-lo na 4.00.
type InfCteAnu struct {
	ChCTe string    // chCTe
	DEmi  time.Time // dEmi
}

// AutXML e um CNPJ/CPF autorizado a baixar o XML.
// Porte de TautXMLCollectionItem.
type AutXML struct {
	CNPJCPF string // CNPJ ou CPF
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

// InfCTeSupl e o grupo suplementar com o QR-Code. Porte de TinfCTeSupl.
type InfCTeSupl struct {
	QRCodCTe string // qrCodCTe
}

// ---------------------------------------------------------------------------
// Raiz
// ---------------------------------------------------------------------------

// InfCTeNorm e o grupo do CT-e NORMAL. Porte de TInfCTeNorm.
//
// E onde os seis modais entram, e eles sao MUTUAMENTE EXCLUSIVOS: o campo
// `modal` do Ide decide qual sai. Os seis tipos vivem em classes_modais.go.
type InfCTeNorm struct {
	InfCarga InfCarga // infCarga
	InfDoc   InfDoc   // infDoc
	DocAnt   DocAnt   // docAnt
	Seg      []Seg    // seg

	// Os seis modais, mutuamente exclusivos. O rodoOS e a variante do
	// rodoviario usada no CT-e OS.
	Rodo       Rodo       // rodo
	RodoOS     RodoOS     // rodoOS
	Aereo      Aereo      // aereo
	Aquav      Aquav      // aquav
	Ferrov     Ferrov     // ferrov
	Duto       Duto       // duto
	Multimodal Multimodal // multimodal

	Peri           []Peri         // peri
	VeicNovos      []VeicNovos    // veicNovos
	Cobr           Cobr           // cobr
	InfCteSub      InfCteSub      // infCteSub
	InfGlobalizado InfGlobalizado // infGlobalizado
	InfServVinc    InfServVinc    // infServVinc
	InfServico     InfServico     // infServico
	InfDocRef      []InfDocRef    // infDocRef
	RefCTeCanc     string         // refCTeCanc
	InfGTVe        []InfGTVe      // infGTVe
}

// InfModal e o grupo de modal do CT-e Simplificado. Porte de TinfModal.
//
// Tem TRES modais, nao seis: o Simplificado so admite rodoviario, aereo e
// aquaviario. Nao e omissao do porte -- e o que o ACBr declara
// (Classes.pas:2661).
type InfModal struct {
	Rodo  Rodo  // rodo
	Aereo Aereo // aereo
	Aquav Aquav // aquav
}

// CTe e o Conhecimento de Transporte Eletronico completo.
// Porte de TCTe (Classes.pas:2724).
//
// Um unico struct atende os QUATRO modelos, e os grupos que nao pertencem
// ao modelo em uso ficam zerados:
//
//	CT-e (57)           ide + compl + participantes + vPrest + imp + infCTeNorm
//	CT-e OS (67)        idem, com `toma` no lugar de toma03/toma4, e infServico
//	GTV-e (64)          origem, destino e detGTV
//	CT-e Simplificado   det, infModal, total -- em vez de infCTeNorm
type CTe struct {
	InfCTe InfCTe // infCTe
	Ide    Ide    // ide
	Compl  Compl  // compl
	Emit   Emit   // emit
	Toma   Toma   // toma -- CT-e OS
	Rem    Rem    // rem
	Exped  Exped  // exped
	Receb  Receb  // receb
	Dest   Dest   // dest
	VPrest VPrest // vPrest
	Imp    Imp    // imp

	InfCTeNorm   InfCTeNorm       // infCTeNorm
	InfCteComp   InfCteComp       // infCteComp -- forma antiga
	InfCteComp10 []InfCteCompItem // infCteComp -- forma da 3.00 em diante
	InfCteAnu    InfCteAnu        // infCteAnu
	AutXML       []AutXML         // autXML
	InfRespTec   InfRespTec       // infRespTec
	InfCTeSupl   InfCTeSupl       // infCTeSupl

	// GTV-e: origem e destino usam o EnderEmit, nao o Endereco -- e assim
	// no ACBr (Classes.pas:2744).
	Origem  EnderEmit // origem
	Destino EnderEmit // destino
	DetGTV  DetGTV    // detGTV

	// CT-e Simplificado
	InfCarga  InfCarga  // infCarga
	Det       []Det     // det
	InfModal  InfModal  // infModal
	Cobr      Cobr      // cobr
	InfCteSub InfCteSub // infCteSub
	Total     Total     // total

	PgtoVinc PgtoVinc // pgtoVinc

	// ProcCTe nao cobre tudo que o TProcCTe tem: faltam XML_CTe e
	// XML_prot (pcteProcCTe.pas:47-91, marcados "Usando na Montagem do
	// cteProc") e os tres caminhos de arquivo. Divida da etapa de web
	// services.
	ProcCTe   pcn.ProcDFe   // procCTe
	Signature pcn.Signature // Signature
}
