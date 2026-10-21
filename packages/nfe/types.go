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

// Tipos de controle do componente: estado, versao de leiaute, schema e
// layout de servico. Porte das declaracoes correspondentes de
// PCNNFe/pcnConversaoNFe.pas (ver doc.go sobre a escolha dessa unit).
//
// Os membros destes enums NAO tem codigo no leiaute, com excecao de VersaoDF,
// VersaoQrCode e ModeloDF -- esses sao gravados no XML e por isso expoem
// String() com o codigo exato.

// ===========================================================================
// StatusNFe (TStatusACBrNFe)
// ===========================================================================

// StatusNFe e o estado interno do componente durante uma operacao.
// Nao tem representacao no leiaute.
type StatusNFe int

const (
	StIdle StatusNFe = iota
	StNFeStatusServico
	StNFeRecepcao
	StNFeRetRecepcao
	StNFeConsulta
	StNFeCancelamento
	StNFeInutilizacao
	StNFeRecibo
	StNFeCadastro
	StNFeEmail
	StNFeCCe
	StNFeEvento
	StConsNFeDest
	StDownloadNFe
	StAdmCSCNFCe
	StDistDFeInt
	StEnvioWebService
	// _fimStatusNFe conta os membros declarados; nao e valor do leiaute.
	_fimStatusNFe
)

var statusNFeNomes = [...]string{
	"Idle", "NFeStatusServico", "NFeRecepcao", "NFeRetRecepcao",
	"NFeConsulta", "NFeCancelamento", "NFeInutilizacao", "NFeRecibo",
	"NFeCadastro", "NFeEmail", "NFeCCe", "NFeEvento", "ConsNFeDest",
	"DownloadNFe", "AdmCSCNFCe", "DistDFeInt", "EnvioWebService",
}

// String retorna o nome do estado, para log e diagnostico.
func (v StatusNFe) String() string {
	if int(v) < 0 || int(v) >= len(statusNFeNomes) {
		return "Desconhecido"
	}
	return statusNFeNomes[v]
}

// ===========================================================================
// VersaoDF (TpcnVersaoDF)
// ===========================================================================

// VersaoDF e a versao do leiaute da NF-e. Diferente da NFGas, que tem versao
// unica, a NF-e convive com quatro, e o gerador decide campo a campo por ela.
type VersaoDF int

const (
	Ve200 VersaoDF = iota // 2.00
	Ve300                 // 3.00
	Ve310                 // 3.10
	Ve400                 // 4.00
	// _fimVersaoDF conta os membros declarados; nao e valor do leiaute.
	_fimVersaoDF
)

var versaoDFCodigos = [...]string{"2.00", "3.00", "3.10", "4.00"}
var versaoDFValores = [...]float64{2.00, 3.00, 3.10, 4.00}

// String retorna o codigo do leiaute ("2.00".."4.00").
func (v VersaoDF) String() string { return codigoEm(versaoDFCodigos[:], int(v)) }

// Float retorna a versao como numero, equivalente a VersaoDFToDbl.
// Fora de faixa devolve 0, como o case-else do Delphi.
func (v VersaoDF) Float() float64 {
	if int(v) < 0 || int(v) >= len(versaoDFValores) {
		return 0
	}
	return versaoDFValores[v]
}

// ParseVersaoDF converte o codigo do leiaute no enum.
func ParseVersaoDF(s string) (VersaoDF, error) {
	i, err := indiceDe(versaoDFCodigos[:], s, "VersaoDF")
	return VersaoDF(i), err
}

// ParseVersaoDFFloat converte um numero na versao, por FAIXA e nao por
// igualdade -- porte literal de DblToVersaoDF (pcnConversaoNFe.pas:1344).
//
// A faixa e aberta nas duas pontas: qualquer valor abaixo de 3.00 vira Ve200 e
// qualquer valor a partir de 4.00 vira Ve400. Para qualquer valor NUMERICO o
// ramo de erro do Delphi e portanto inalcancavel, e esta funcao nunca devolve
// erro -- replicado de proposito, com teste.
//
// NaN e a unica entrada que separa os dois: no Delphi todas as comparacoes
// falham e o else devolve ve310; aqui o default devolve Ve400. Nao tratado,
// porque a entrada vem de leitura de numero do .ini e nunca e NaN.
func ParseVersaoDFFloat(d float64) VersaoDF {
	switch {
	case d == 2.0 || d < 3.0:
		return Ve200
	case d >= 3.0 && d < 3.1:
		return Ve300
	case d >= 3.10 && d < 4:
		return Ve310
	default: // d >= 4
		return Ve400
	}
}

// ===========================================================================
// VersaoQrCode (TpcnVersaoQrCode)
// ===========================================================================

// VersaoQrCode e a versao do QR-Code da NFC-e.
type VersaoQrCode int

const (
	Veqr000 VersaoQrCode = iota // 0 - sem QR-Code
	Veqr100                     // 1
	Veqr200                     // 2
	Veqr300                     // 3
	// _fimVersaoQrCode conta os membros declarados; nao e valor do leiaute.
	_fimVersaoQrCode
)

var versaoQrCodeCodigos = [...]string{"0", "1", "2", "3"}

// String retorna o codigo do leiaute ("0".."3").
func (v VersaoQrCode) String() string { return codigoEm(versaoQrCodeCodigos[:], int(v)) }

// Float retorna a versao como numero, equivalente a VersaoQrCodeToDbl.
// Fora de faixa devolve 0, como o case-else do Delphi.
func (v VersaoQrCode) Float() float64 {
	if int(v) < 0 || int(v) >= len(versaoQrCodeCodigos) {
		return 0
	}
	return float64(v)
}

// ParseVersaoQrCode converte o codigo do leiaute no enum.
func ParseVersaoQrCode(s string) (VersaoQrCode, error) {
	i, err := indiceDe(versaoQrCodeCodigos[:], s, "VersaoQrCode")
	return VersaoQrCode(i), err
}

// AjustarVersaoQrCode aplica a regra de AjustarVersaoQRCode
// (ACBrNFe.pas:502): ate a versao 3.10 do leiaute nao existe QR-Code; da 4.00
// em diante a versao do QR-Code e no minimo 1.
func AjustarVersaoQrCode(versaoQrCode VersaoQrCode, versaoXML VersaoDF) VersaoQrCode {
	if versaoXML <= Ve310 {
		return Veqr000
	}
	if versaoQrCode < Veqr100 {
		return Veqr100
	}
	return versaoQrCode
}

// ===========================================================================
// ModeloDF (TpcnModeloDF)
// ===========================================================================

// ModeloDF distingue a NF-e (modelo 55) da NFC-e (modelo 65).
type ModeloDF int

const (
	MoNFe  ModeloDF = iota // 55
	MoNFCe                 // 65
	// _fimModeloDF conta os membros declarados; nao e valor do leiaute.
	_fimModeloDF
)

var modeloDFCodigos = [...]string{"55", "65"}

// String retorna o codigo do leiaute ("55" ou "65").
func (v ModeloDF) String() string { return codigoEm(modeloDFCodigos[:], int(v)) }

// Prefixo retorna o prefixo usado em nome de arquivo e de tag
// ("NFe" ou "NFCe"), porte de ModeloDFToPrefixo.
func (v ModeloDF) Prefixo() string {
	if v == MoNFCe {
		return "NFCe"
	}
	return "NFe"
}

// ParseModeloDF converte o codigo do leiaute no enum.
func ParseModeloDF(s string) (ModeloDF, error) {
	i, err := indiceDe(modeloDFCodigos[:], s, "ModeloDF")
	return ModeloDF(i), err
}

// ===========================================================================
// ForcarGeracaoTag (TForcarGeracaoTag)
// ===========================================================================

// ForcarGeracaoTag controla, na configuracao do componente, se uma tag
// opcional deve sair mesmo sem conteudo. Nao tem codigo no leiaute.
type ForcarGeracaoTag int

const (
	FgtNunca ForcarGeracaoTag = iota
	FgtSomenteProducao
	FgtSomenteHomologacao
	FgtSempre
	// _fimForcarGeracaoTag conta os membros declarados; nao e valor do leiaute.
	_fimForcarGeracaoTag
)

// ===========================================================================
// SchemaNFe (TSchemaNFe)
// ===========================================================================

// SchemaNFe identifica o schema XSD usado para validar cada mensagem.
type SchemaNFe int

const (
	SchErro SchemaNFe = iota
	SchNfe
	SchcancNFe
	SchInutNFe
	SchEnvCCe
	SchEnvEventoCancNFe
	SchEnvConfRecebto
	SchEnvEPEC
	SchconsReciNFe
	SchconsSitNFe
	SchconsStatServ
	SchconsCad
	SchenvEvento
	SchconsNFeDest
	SchdownloadNFe
	SchretEnviNFe
	SchadmCscNFCe
	SchdistDFeInt
	ScheventoEPEC
	SchCancSubst
	SchPedProrrog1
	SchPedProrrog2
	SchCanPedProrrog1
	SchCanPedProrrog2
	SchManifDestConfirmacao
	SchManifDestCiencia
	SchManifDestDesconhecimento
	SchManifDestOperNaoRealizada
	SchCompEntrega
	SchCancCompEntrega
	SchAtorInteressadoNFe
	SchInsucessoEntregaNFe
	SchCancInsucessoEntregaNFe
	SchConcFinanceira
	SchCancConcFinanceira
	SchCancGenerico
	SchPagIntegLibCredPresAdq
	SchImporALCZFM
	SchPerecPerdaRouboFurtoTranspContratFornec
	SchFornecNaoRealizPagAntec
	SchSolicApropCredPres
	SchDestItemConsPessoal
	SchPerecPerdaRouboFurtoTranspContratAqu
	SchAceiteDebitoApuracaoNotaCredito
	SchImobilizacaoItem
	SchSolicApropCredCombustivel
	SchSolicApropCredBensServicos
	SchManifPedTransfCredIBSSucessao
	SchManifPedTransfCredCBSSucessao
	SchAtualizacaoDataPrevisaoEntrega
	// _fimSchemaNFe conta os membros declarados; nao e valor do leiaute.
	_fimSchemaNFe
)

// schemaNFeNomes e o nome do IDENTIFICADOR Delphi sem o prefixo "sch", que e
// o que SchemaNFeToStr devolve (ela usa RTTI: GetEnumName + Copy).
//
// Atencao: NAO e o conteudo de TSchemaNFeArrayStrings. As duas tabelas
// divergem em dois membros -- o array diz "InsucessoEntrega" e
// "CancInsucessoEntrega", a RTTI diz "InsucessoEntregaNFe" e
// "CancInsucessoEntregaNFe". Vale a RTTI: TSchemaNFeArrayStrings esta MORTA no
// componente (nenhuma unit da NFe a referencia), enquanto SchemaNFeToStr e
// chamada por ACBrNFe.pas.
var schemaNFeNomes = [...]string{
	"Erro", "Nfe", "cancNFe", "InutNFe", "EnvCCe", "EnvEventoCancNFe",
	"EnvConfRecebto", "EnvEPEC", "consReciNFe", "consSitNFe", "consStatServ",
	"consCad", "envEvento", "consNFeDest", "downloadNFe", "retEnviNFe",
	"admCscNFCe", "distDFeInt", "eventoEPEC", "CancSubst", "PedProrrog1",
	"PedProrrog2", "CanPedProrrog1", "CanPedProrrog2", "ManifDestConfirmacao",
	"ManifDestCiencia", "ManifDestDesconhecimento", "ManifDestOperNaoRealizada",
	"CompEntrega", "CancCompEntrega", "AtorInteressadoNFe",
	"InsucessoEntregaNFe", "CancInsucessoEntregaNFe", "ConcFinanceira",
	"CancConcFinanceira", "CancGenerico", "PagIntegLibCredPresAdq",
	"ImporALCZFM", "PerecPerdaRouboFurtoTranspContratFornec",
	"FornecNaoRealizPagAntec", "SolicApropCredPres", "DestItemConsPessoal",
	"PerecPerdaRouboFurtoTranspContratAqu", "AceiteDebitoApuracaoNotaCredito",
	"ImobilizacaoItem", "SolicApropCredCombustivel",
	"SolicApropCredBensServicos", "ManifPedTransfCredIBSSucessao",
	"ManifPedTransfCredCBSSucessao", "AtualizacaoDataPrevisaoEntrega",
}

// String retorna o nome do schema sem o prefixo "sch", porte de
// SchemaNFeToStr.
func (v SchemaNFe) String() string { return codigoEm(schemaNFeNomes[:], int(v)) }

// schemaEventoCodigos mapeia o schema no codigo do evento ("eNNNNNN"), porte
// de SchemaEventoToStr. Schema que nao e de evento nao aparece aqui e devolve
// string vazia, como o EnumeradoToStr do Delphi.
var schemaEventoCodigos = map[SchemaNFe]string{
	SchCancGenerico:              "e110001",
	SchEnvCCe:                    "e110110",
	SchcancNFe:                   "e110111",
	SchCancSubst:                 "e110112",
	SchEnvEPEC:                   "e110140",
	SchPedProrrog1:               "e111500",
	SchPedProrrog2:               "e111501",
	SchCanPedProrrog1:            "e111502",
	SchCanPedProrrog2:            "e111503",
	SchManifDestConfirmacao:      "e210200",
	SchManifDestCiencia:          "e210210",
	SchManifDestDesconhecimento:  "e210220",
	SchManifDestOperNaoRealizada: "e210240",
	SchCompEntrega:               "e110130",
	SchCancCompEntrega:           "e110131",
	SchAtorInteressadoNFe:        "e110150",
	SchInsucessoEntregaNFe:       "e110192",
	SchCancInsucessoEntregaNFe:   "e110193",
	SchConcFinanceira:            "e110750",
	SchCancConcFinanceira:        "e110751",

	// Reforma Tributaria
	SchPagIntegLibCredPresAdq:                  "e112110",
	SchImporALCZFM:                             "e112120",
	SchPerecPerdaRouboFurtoTranspContratFornec: "e112130",
	SchFornecNaoRealizPagAntec:                 "e112140",
	SchSolicApropCredPres:                      "e211110",
	SchDestItemConsPessoal:                     "e211120",
	SchPerecPerdaRouboFurtoTranspContratAqu:    "e211124",
	SchAceiteDebitoApuracaoNotaCredito:         "e211128",
	SchImobilizacaoItem:                        "e211130",
	SchSolicApropCredCombustivel:               "e211140",
	SchSolicApropCredBensServicos:              "e211150",
	SchManifPedTransfCredIBSSucessao:           "e212110",
	SchManifPedTransfCredCBSSucessao:           "e212120",
	SchAtualizacaoDataPrevisaoEntrega:          "e112150",
}

// SchemaEvento retorna o codigo do schema de evento ("e110111"), ou vazio
// quando o schema nao e de evento. Porte de SchemaEventoToStr.
func (v SchemaNFe) SchemaEvento() string { return schemaEventoCodigos[v] }

// ===========================================================================
// LayOut (TLayOut)
// ===========================================================================

// LayOut identifica o servico da SEFAZ, e e a chave de secao usada para
// resolver URL, versao e SOAPAction em ACBrNFeServicos.ini.
type LayOut int

const (
	LayNfeRecepcao LayOut = iota
	LayNfeRetRecepcao
	LayNfeCancelamento
	LayNfeInutilizacao
	LayNfeConsulta
	LayNfeStatusServico
	LayNfeCadastro
	LayNFeCCe
	LayNFeEvento
	LayNFeEventoAN
	LayNFeConsNFeDest
	LayNFeDownloadNFe
	LayNfeAutorizacao
	LayNfeRetAutorizacao
	LayAdministrarCSCNFCe
	LayDistDFeInt
	LayNFCeEPEC
	LayNFeURLQRCode
	LayURLConsultaNFe
	LayURLConsultaNFCe
	// _fimLayOut conta os membros declarados; nao e valor do leiaute.
	_fimLayOut
)

// layOutServicos e o nome do servico por layout, porte de LayOutToServico.
//
// Os tres ultimos membros (LayNFeURLQRCode, LayURLConsultaNFe,
// LayURLConsultaNFCe) NAO estao na tabela de LayOutToServico e por isso
// resolvem para vazio -- replicado. Eles existem para enderecar a URL de
// QR-Code e de consulta, que sao lidas por outro caminho (GetURLQRCode e
// GetURLConsulta, ACBrNFe.pas), nao pelo nome de servico.
//
// LayNFeCCe, LayNFeEvento e LayNFeEventoAN compartilham o mesmo servico
// ("RecepcaoEvento"), de proposito.
var layOutServicos = [...]string{
	"NfeRecepcao", "NfeRetRecepcao", "NfeCancelamento", "NfeInutilizacao",
	"NfeConsultaProtocolo", "NfeStatusServico", "NfeConsultaCadastro",
	"RecepcaoEvento", "RecepcaoEvento", "RecepcaoEvento", "NfeConsultaDest",
	"NfeDownloadNF", "NfeAutorizacao", "NfeRetAutorizacao",
	"AdministrarCSCNFCe", "NFeDistribuicaoDFe", "EventoEPEC",
	"", "", "",
}

// Servico retorna o nome do servico do layout, porte de LayOutToServico.
func (v LayOut) Servico() string { return codigoEm(layOutServicos[:], int(v)) }

// ParseLayOutServico resolve o layout pelo nome do servico, porte de
// ServicoToLayOut.
//
// Como o Delphi para na PRIMEIRA ocorrencia, "RecepcaoEvento" resolve sempre
// para LayNFeCCe -- nunca para LayNFeEvento nem LayNFeEventoAN. Replicado.
func ParseLayOutServico(s string) (LayOut, error) {
	for i, nome := range layOutServicos {
		// a tabela do Delphi so tem os 17 primeiros; o vazio dos tres
		// ultimos nao e um nome de servico valido
		if nome != "" && nome == s {
			return LayOut(i), nil
		}
	}
	return LayNfeRecepcao, erroEnum("LayOut", s)
}

// layOutSchemas mapeia o layout no schema de validacao, porte de
// LayOutToSchema. Layout ausente da tabela resolve para SchErro, como o
// case-else do Delphi.
var layOutSchemas = map[LayOut]SchemaNFe{
	LayNfeRecepcao:        SchNfe,
	LayNfeRetRecepcao:     SchconsReciNFe,
	LayNfeCancelamento:    SchcancNFe,
	LayNfeInutilizacao:    SchInutNFe,
	LayNfeConsulta:        SchconsSitNFe,
	LayNfeStatusServico:   SchconsStatServ,
	LayNfeCadastro:        SchconsCad,
	LayNFeCCe:             SchenvEvento,
	LayNFeEvento:          SchenvEvento,
	LayNFeEventoAN:        SchenvEvento,
	LayNFeConsNFeDest:     SchconsNFeDest,
	LayNFeDownloadNFe:     SchdownloadNFe,
	LayNfeAutorizacao:     SchNfe,
	LayNfeRetAutorizacao:  SchretEnviNFe,
	LayAdministrarCSCNFCe: SchadmCscNFCe,
	LayDistDFeInt:         SchdistDFeInt,
	LayNFCeEPEC:           ScheventoEPEC,
}

// Schema retorna o schema de validacao do layout, porte de LayOutToSchema.
func (v LayOut) Schema() SchemaNFe {
	if s, ok := layOutSchemas[v]; ok {
		return s
	}
	return SchErro
}
