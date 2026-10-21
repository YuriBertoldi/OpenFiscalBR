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
	"errors"
	"fmt"
	"strings"
)

// Enums do leiaute da NF-e. Porte de PCNNFe/pcnConversaoNFe.pas -- ver doc.go
// sobre por que esta unit prevalece sobre Base/ACBrNFe.Conversao.pas.
//
// Cada enum expoe String() com o codigo EXATO do leiaute e ParseXxx com erro.
// A ordem dos membros Go reproduz a ordem de declaracao do Delphi, porque e
// ela que define o ordinal -- e ordinal errado e o defeito que nao aparece em
// build, vet nem teste ingenuo.
//
// Enums compartilhados por todos os DFe (TipoAmbiente, TipoEmissao,
// Indicador, OrigemMercadoria, CSTIcms, CSTPis, CSTCofins) vivem em
// packages/pcn e NAO sao redeclarados aqui.

// ErrEnumInvalido indica valor de string que nao corresponde a nenhum
// membro do enum.
var ErrEnumInvalido = errors.New("nfe: valor invalido para enum")

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %s nao aceita %q", ErrEnumInvalido, enum, valor)
}

// codigoEm devolve o codigo da posicao i, ou vazio se fora de faixa.
// Membro sentinela -1 cai no vazio, como o EnumeradoToStr do Delphi faz com
// TXxx(-1).
func codigoEm(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}

// indiceDe devolve a posicao do codigo na tabela, comparando SEM diferenciar
// maiusculas de minusculas -- e o AnsiSameText de StrToEnumerado
// (ACBrDFe.Conversao.pas:834). So muda o resultado nos enums cujo codigo nao
// e numerico: IndEscala ("S"/"N"), ISSQNcSitTrib ("N"/"R"/"S"/"I") e
// ECFModRef ("2B"/"2C"/"2D").
func indiceDe(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if strings.EqualFold(c, s) {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// indiceDeExato compara diferenciando maiusculas de minusculas, para os enums
// que o ACBr converte com um laco proprio usando "=" em vez de StrToEnumerado:
// TpcnIndicadorPagamento, TtpGuia, TtpNFDebito, TtpNFCredito e TCSTIS.
//
// Nesses cinco o Delphi levanta EACBrException em vez de devolver um fallback,
// entao o chamador PRECISA tratar o erro -- ignora-lo grava o membro zero numa
// nota em que o ACBr teria abortado a leitura.
func indiceDeExato(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if c == s {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}

// ===========================================================================
// IndicadorPagamento (TpcnIndicadorPagamento)
// ===========================================================================

// IndicadorPagamento e o indicador da forma de pagamento (tag indPag), campo
// que saiu do leiaute 4.00 e sobrevive para gerar as versoes anteriores.
//
// A conversao segue IndpagToStrEX/StrToIndpagEX (pcnConversaoNFe.pas:554-572),
// e NAO o par IndpagToStr/StrToIndpag, que esta marcado deprecated na propria
// unit (:63-64). A diferenca importa: o par EX usa laco proprio com "=" (caixa
// respeitada) e levanta EACBrException quando nada casa, em vez do fallback
// silencioso do StrToEnumerado. Todo o codigo vivo usa o EX -- XmlReader.pas:238
// e :1497, IniReader.pas:204 e :1510, IniWriter.pas:225 e :1348, JSONReader e
// JSONWriter.
type IndicadorPagamento int

const (
	IpVista  IndicadorPagamento = iota // 0
	IpPrazo                            // 1
	IpOutras                           // 2
	IpNenhum                           // vazio
	// _fimIndicadorPagamento conta os membros declarados; nao e valor do leiaute.
	_fimIndicadorPagamento
)

var indicadorPagamentoCodigos = [...]string{"0", "1", "2", ""}

// String retorna o codigo do leiaute.
func (v IndicadorPagamento) String() string {
	return codigoEm(indicadorPagamentoCodigos[:], int(v))
}

// ParseIndicadorPagamento converte o codigo do leiaute no enum. Comparacao
// exata, como o laco proprio de StrToIndpagEX. O erro TEM de ser tratado: o
// ACBr levanta excecao neste ponto, entao nao existe fallback com o mesmo
// significado do Delphi.
func ParseIndicadorPagamento(s string) (IndicadorPagamento, error) {
	i, err := indiceDeExato(indicadorPagamentoCodigos[:], s, "IndicadorPagamento")
	return IndicadorPagamento(i), err
}

// ===========================================================================
// DeterminacaoBaseIcms (TpcnDeterminacaoBaseIcms) -- tag modBC
// ===========================================================================

// DeterminacaoBaseIcms e a modalidade de determinacao da base de calculo do
// ICMS (campo N13 do leiaute).
type DeterminacaoBaseIcms int

const (
	DbiMargemValorAgregado DeterminacaoBaseIcms = iota // 0
	DbiPauta                                           // 1
	DbiPrecoTabelado                                   // 2
	DbiValorOperacao                                   // 3
	DbiNenhum                                          // vazio
	// _fimDeterminacaoBaseIcms conta os membros declarados; nao e valor do leiaute.
	_fimDeterminacaoBaseIcms
)

var determinacaoBaseIcmsCodigos = [...]string{"0", "1", "2", "3", ""}

// String retorna o codigo do leiaute.
func (v DeterminacaoBaseIcms) String() string {
	return codigoEm(determinacaoBaseIcmsCodigos[:], int(v))
}

// ParseDeterminacaoBaseIcms converte o codigo do leiaute no enum.
func ParseDeterminacaoBaseIcms(s string) (DeterminacaoBaseIcms, error) {
	i, err := indiceDe(determinacaoBaseIcmsCodigos[:], s, "DeterminacaoBaseIcms")
	return DeterminacaoBaseIcms(i), err
}

// ===========================================================================
// DeterminacaoBaseIcmsST (TpcnDeterminacaoBaseIcmsST) -- tag modBCST
// ===========================================================================

// DeterminacaoBaseIcmsST e a modalidade de determinacao da base de calculo do
// ICMS ST (campo N18 do leiaute).
type DeterminacaoBaseIcmsST int

const (
	DbisPrecoTabelado       DeterminacaoBaseIcmsST = iota // 0
	DbisListaNegativa                                     // 1
	DbisListaPositiva                                     // 2
	DbisListaNeutra                                       // 3
	DbisMargemValorAgregado                               // 4
	DbisPauta                                             // 5
	DbisValordaOperacao                                   // 6
	// _fimDeterminacaoBaseIcmsST conta os membros declarados; nao e valor do leiaute.
	_fimDeterminacaoBaseIcmsST

	// DbisNenhum e o sentinela TpcnDeterminacaoBaseIcmsST(-1) do Delphi:
	// nao e membro declarado do enum, so o valor que representa "ausente".
	DbisNenhum DeterminacaoBaseIcmsST = -1
)

var determinacaoBaseIcmsSTCodigos = [...]string{"0", "1", "2", "3", "4", "5", "6"}

// String retorna o codigo do leiaute, ou vazio para DbisNenhum.
func (v DeterminacaoBaseIcmsST) String() string {
	return codigoEm(determinacaoBaseIcmsSTCodigos[:], int(v))
}

// ParseDeterminacaoBaseIcmsST converte o codigo do leiaute no enum.
// String vazia resolve para DbisNenhum, como a ultima posicao da tabela do
// Delphi.
func ParseDeterminacaoBaseIcmsST(s string) (DeterminacaoBaseIcmsST, error) {
	if s == "" {
		return DbisNenhum, nil
	}
	i, err := indiceDe(determinacaoBaseIcmsSTCodigos[:], s, "DeterminacaoBaseIcmsST")
	if err != nil {
		return DbisPrecoTabelado, err
	}
	return DeterminacaoBaseIcmsST(i), nil
}

// ===========================================================================
// MotivoDesoneracaoICMS (TpcnMotivoDesoneracaoICMS) -- tag motDesICMS
// ===========================================================================

// MotivoDesoneracaoICMS e o motivo da desoneracao do ICMS (campo N28).
//
// Atencao ao ordinal: os codigos NAO sao sequenciais a partir de 12 --
// mdiOlimpiadaRio2016 vale "16" e mdiSolicitadoFisco vale "90". Gravar o
// ordinal no lugar do codigo produz XML invalido, e foi exatamente o defeito
// encontrado no leitor da NFGas.
type MotivoDesoneracaoICMS int

const (
	MdiTaxi                  MotivoDesoneracaoICMS = iota // 1
	MdiDeficienteFisico                                   // 2
	MdiProdutorAgropecuario                               // 3
	MdiFrotistaLocadora                                   // 4
	MdiDiplomaticoConsular                                // 5
	MdiAmazoniaLivreComercio                              // 6
	MdiSuframa                                            // 7
	MdiVendaOrgaosPublicos                                // 8
	MdiOutros                                             // 9
	MdiDeficienteCondutor                                 // 10
	MdiDeficienteNaoCondutor                              // 11
	MdiOrgaoFomento                                       // 12
	MdiOlimpiadaRio2016                                   // 16
	MdiSolicitadoFisco                                    // 90
	// _fimMotivoDesoneracaoICMS conta os membros declarados; nao e valor do leiaute.
	_fimMotivoDesoneracaoICMS

	// MdiNenhum e o sentinela TpcnMotivoDesoneracaoICMS(-1) do Delphi.
	MdiNenhum MotivoDesoneracaoICMS = -1
)

var motivoDesoneracaoICMSCodigos = [...]string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "16", "90",
}

// String retorna o codigo do leiaute, ou vazio para MdiNenhum.
func (v MotivoDesoneracaoICMS) String() string {
	return codigoEm(motivoDesoneracaoICMSCodigos[:], int(v))
}

// ParseMotivoDesoneracaoICMS converte o codigo do leiaute no enum.
// String vazia resolve para MdiNenhum.
func ParseMotivoDesoneracaoICMS(s string) (MotivoDesoneracaoICMS, error) {
	if s == "" {
		return MdiNenhum, nil
	}
	i, err := indiceDe(motivoDesoneracaoICMSCodigos[:], s, "MotivoDesoneracaoICMS")
	if err != nil {
		return MdiTaxi, err
	}
	return MotivoDesoneracaoICMS(i), nil
}

// ===========================================================================
// CstIpi (TpcnCstIpi)
// ===========================================================================

// CstIpi e o codigo de situacao tributaria do IPI.
//
// A ordem de declaracao do ACBr NAO e a ordem numerica dos codigos: comeca
// em 00, 49, 50, 99 e so depois retoma em 01. O ordinal nao tem relacao com
// o codigo.
type CstIpi int

const (
	Ipi00 CstIpi = iota // 00 - Entrada com Recuperacao de Credito
	Ipi49               // 49 - Outras Entradas
	Ipi50               // 50 - Saida Tributada
	Ipi99               // 99 - Outras Saidas
	Ipi01               // 01 - Entrada Tributavel com Aliquota Zero
	Ipi02               // 02 - Entrada Isenta
	Ipi03               // 03 - Entrada Nao-Tributada
	Ipi04               // 04 - Entrada Imune
	Ipi05               // 05 - Entrada com Suspensao
	Ipi51               // 51 - Saida Tributavel com Aliquota Zero
	Ipi52               // 52 - Saida Isenta
	Ipi53               // 53 - Saida Nao-Tributada
	Ipi54               // 54 - Saida Imune
	Ipi55               // 55 - Saida com Suspensao
	// _fimCstIpi conta os membros declarados; nao e valor do leiaute.
	_fimCstIpi
)

var cstIpiCodigos = [...]string{
	"00", "49", "50", "99", "01", "02", "03", "04", "05",
	"51", "52", "53", "54", "55",
}

// String retorna o codigo do leiaute.
func (v CstIpi) String() string { return codigoEm(cstIpiCodigos[:], int(v)) }

// ParseCstIpi converte o codigo do leiaute no enum.
func ParseCstIpi(s string) (CstIpi, error) {
	i, err := indiceDe(cstIpiCodigos[:], s, "CstIpi")
	return CstIpi(i), err
}

// ===========================================================================
// IndicadorProcesso (TpcnIndicadorProcesso) -- tag indProc
// ===========================================================================

// IndicadorProcesso e a origem do processo que autoriza a operacao
// (campo 401i do leiaute).
type IndicadorProcesso int

const (
	IpSEFAZ           IndicadorProcesso = iota // 0
	IpJusticaFederal                           // 1
	IpJusticaEstadual                          // 2
	IpSecexRFB                                 // 3
	IpCONFAZ                                   // 4
	IpOutros                                   // 9
	// _fimIndicadorProcesso conta os membros declarados; nao e valor do leiaute.
	_fimIndicadorProcesso
)

var indicadorProcessoCodigos = [...]string{"0", "1", "2", "3", "4", "9"}

// String retorna o codigo do leiaute.
func (v IndicadorProcesso) String() string {
	return codigoEm(indicadorProcessoCodigos[:], int(v))
}

// ParseIndicadorProcesso converte o codigo do leiaute no enum.
func ParseIndicadorProcesso(s string) (IndicadorProcesso, error) {
	i, err := indiceDe(indicadorProcessoCodigos[:], s, "IndicadorProcesso")
	return IndicadorProcesso(i), err
}

// ===========================================================================
// CRT (TpcnCRT) -- tag CRT
// ===========================================================================

// CRT e o codigo de regime tributario do emitente (campo 49a do leiaute).
//
// O mapeamento do ACBr lista crtRegimeNormal DUAS vezes -- na posicao do
// codigo vazio e na do codigo "3" (pcnConversaoNFe.pas:776). Como
// EnumeradoToStr nao interrompe o laco, vale a ULTIMA ocorrencia, e
// crtRegimeNormal grava "3". Na leitura, porem, StrToEnumerado interrompe na
// PRIMEIRA, entao a string vazia tambem resolve para crtRegimeNormal -- sem
// erro. As duas pontas estao replicadas abaixo.
type CRT int

const (
	CrtSimplesNacional       CRT = iota // 1
	CrtSimplesExcessoReceita            // 2
	CrtRegimeNormal                     // 3 (e tambem o destino da string vazia)
	CrtMEI                              // 4
	// _fimCRT conta os membros declarados; nao e valor do leiaute.
	_fimCRT
)

var crtCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v CRT) String() string { return codigoEm(crtCodigos[:], int(v)) }

// ParseCRT converte o codigo do leiaute no enum. String vazia e valida e
// resolve para CrtRegimeNormal, e esse tambem e o valor devolvido junto do
// erro -- os dois comportamentos vem de StrToEnumerado, cujo fallback de
// falha e AEnumerados[0], aqui crtRegimeNormal.
func ParseCRT(s string) (CRT, error) {
	if s == "" {
		return CrtRegimeNormal, nil
	}
	i, err := indiceDe(crtCodigos[:], s, "CRT")
	if err != nil {
		return CrtRegimeNormal, err
	}
	return CRT(i), nil
}

// ===========================================================================
// IndicadorTotal (TpcnIndicadorTotal) -- tag indTot
// ===========================================================================

// IndicadorTotal diz se o valor do item compoe o total da NF-e (campo 117b).
//
// Cuidado com o valor zero: o ACBr declara itSomaTotalNFe PRIMEIRO, entao um
// item recem-criado ja nasce "compoe o total" e grava "1". Trocar a ordem dos
// membros aqui inverteria o default de todo item.
type IndicadorTotal int

const (
	ItSomaTotalNFe    IndicadorTotal = iota // 1
	ItNaoSomaTotalNFe                       // 0
	// _fimIndicadorTotal conta os membros declarados; nao e valor do leiaute.
	_fimIndicadorTotal
)

var indicadorTotalCodigos = [...]string{"1", "0"}

// String retorna o codigo do leiaute.
func (v IndicadorTotal) String() string {
	return codigoEm(indicadorTotalCodigos[:], int(v))
}

// ParseIndicadorTotal converte o codigo do leiaute no enum. O valor devolvido
// junto do erro e ItNaoSomaTotalNFe, primeiro da tabela de mapeamento do
// ACBr e portanto o fallback de StrToEnumerado.
func ParseIndicadorTotal(s string) (IndicadorTotal, error) {
	i, err := indiceDe(indicadorTotalCodigos[:], s, "IndicadorTotal")
	if err != nil {
		return ItNaoSomaTotalNFe, err
	}
	return IndicadorTotal(i), nil
}

// ===========================================================================
// ECFModRef (TpcnECFModRef) -- tag mod do grupo refECF
// ===========================================================================

// ECFModRef e o modelo do ECF referenciado (campo B20k do leiaute).
type ECFModRef int

const (
	ECFModRefVazio ECFModRef = iota // vazio
	ECFModRef2B                     // 2B
	ECFModRef2C                     // 2C
	ECFModRef2D                     // 2D
	// _fimECFModRef conta os membros declarados; nao e valor do leiaute.
	_fimECFModRef
)

var ecfModRefCodigos = [...]string{"", "2B", "2C", "2D"}

// String retorna o codigo do leiaute.
func (v ECFModRef) String() string { return codigoEm(ecfModRefCodigos[:], int(v)) }

// ParseECFModRef converte o codigo do leiaute no enum.
func ParseECFModRef(s string) (ECFModRef, error) {
	i, err := indiceDe(ecfModRefCodigos[:], s, "ECFModRef")
	return ECFModRef(i), err
}

// ===========================================================================
// ISSQNcSitTrib (TpcnISSQNcSitTrib) -- tag cSitTrib do ISSQN
// ===========================================================================

// ISSQNcSitTrib e a situacao tributaria do ISSQN.
type ISSQNcSitTrib int

const (
	ISSQNcSitTribVazio      ISSQNcSitTrib = iota // vazio
	ISSQNcSitTribNormal                          // N
	ISSQNcSitTribRetida                          // R
	ISSQNcSitTribSubstituta                      // S
	ISSQNcSitTribIsenta                          // I
	// _fimISSQNcSitTrib conta os membros declarados; nao e valor do leiaute.
	_fimISSQNcSitTrib
)

var issqncSitTribCodigos = [...]string{"", "N", "R", "S", "I"}

// String retorna o codigo do leiaute.
func (v ISSQNcSitTrib) String() string { return codigoEm(issqncSitTribCodigos[:], int(v)) }

// ParseISSQNcSitTrib converte o codigo do leiaute no enum.
func ParseISSQNcSitTrib(s string) (ISSQNcSitTrib, error) {
	i, err := indiceDe(issqncSitTribCodigos[:], s, "ISSQNcSitTrib")
	return ISSQNcSitTrib(i), err
}

// ===========================================================================
// DestinoOperacao (TpcnDestinoOperacao) -- tag idDest
// ===========================================================================

// DestinoOperacao e o local de destino da operacao.
type DestinoOperacao int

const (
	DoInterna       DestinoOperacao = iota // 1
	DoInterestadual                        // 2
	DoExterior                             // 3
	// _fimDestinoOperacao conta os membros declarados; nao e valor do leiaute.
	_fimDestinoOperacao
)

var destinoOperacaoCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v DestinoOperacao) String() string { return codigoEm(destinoOperacaoCodigos[:], int(v)) }

// ParseDestinoOperacao converte o codigo do leiaute no enum.
func ParseDestinoOperacao(s string) (DestinoOperacao, error) {
	i, err := indiceDe(destinoOperacaoCodigos[:], s, "DestinoOperacao")
	return DestinoOperacao(i), err
}

// ===========================================================================
// ConsumidorFinal (TpcnConsumidorFinal) -- tag indFinal
// ===========================================================================

// ConsumidorFinal indica operacao com consumidor final.
type ConsumidorFinal int

const (
	CfNao             ConsumidorFinal = iota // 0
	CfConsumidorFinal                        // 1
	// _fimConsumidorFinal conta os membros declarados; nao e valor do leiaute.
	_fimConsumidorFinal
)

var consumidorFinalCodigos = [...]string{"0", "1"}

// String retorna o codigo do leiaute.
func (v ConsumidorFinal) String() string { return codigoEm(consumidorFinalCodigos[:], int(v)) }

// ParseConsumidorFinal converte o codigo do leiaute no enum.
func ParseConsumidorFinal(s string) (ConsumidorFinal, error) {
	i, err := indiceDe(consumidorFinalCodigos[:], s, "ConsumidorFinal")
	return ConsumidorFinal(i), err
}

// ===========================================================================
// PresencaComprador (TpcnPresencaComprador) -- tag indPres
// ===========================================================================

// PresencaComprador indica a presenca do comprador no estabelecimento.
type PresencaComprador int

const (
	PcNao                           PresencaComprador = iota // 0
	PcPresencial                                             // 1
	PcInternet                                               // 2
	PcTeleatendimento                                        // 3
	PcEntregaDomicilio                                       // 4
	PcPresencialForaEstabelecimento                          // 5
	PcOutros                                                 // 9
	// _fimPresencaComprador conta os membros declarados; nao e valor do leiaute.
	_fimPresencaComprador
)

var presencaCompradorCodigos = [...]string{"0", "1", "2", "3", "4", "5", "9"}

// String retorna o codigo do leiaute.
func (v PresencaComprador) String() string {
	return codigoEm(presencaCompradorCodigos[:], int(v))
}

// ParsePresencaComprador converte o codigo do leiaute no enum.
func ParsePresencaComprador(s string) (PresencaComprador, error) {
	i, err := indiceDe(presencaCompradorCodigos[:], s, "PresencaComprador")
	return PresencaComprador(i), err
}

// ===========================================================================
// FormaPagamento (TpcnFormaPagamento) -- tag tPag
// ===========================================================================

// FormaPagamento e a forma de pagamento do grupo pag.
//
// Como no CstIpi, a ordem de declaracao nao acompanha a ordem numerica: os
// codigos saltam de "19" para "90", "98", "99" e so entao voltam para "20".
// Os quatro ultimos membros sao acrescimos recentes do leiaute e existem
// APENAS em PCNNFe/pcnConversaoNFe.pas.
type FormaPagamento int

const (
	FpDinheiro                     FormaPagamento = iota // 01
	FpCheque                                             // 02
	FpCartaoCredito                                      // 03
	FpCartaoDebito                                       // 04
	FpCreditoLoja                                        // 05
	FpValeAlimentacao                                    // 10
	FpValeRefeicao                                       // 11
	FpValePresente                                       // 12
	FpValeCombustivel                                    // 13
	FpDuplicataMercantil                                 // 14
	FpBoletoBancario                                     // 15
	FpDepositoBancario                                   // 16
	FpPagamentoInstantaneo                               // 17 - PIX dinamico
	FpTransfBancario                                     // 18
	FpProgramaFidelidade                                 // 19
	FpSemPagamento                                       // 90
	FpRegimeEspecial                                     // 98 - Regime Especial NFF
	FpOutro                                              // 99
	FpPagamentoInstantaneoEstatico                       // 20 - PIX estatico
	FpCreditoEmLojaPorDevolucao                          // 21
	FpFalhaHardware                                      // 22
	FpPagamentoPosterior                                 // 91
	FpPagInstantaneoPIXAutomatico                        // 23
	FpTEFBookTransfer                                    // 24
	// _fimFormaPagamento conta os membros declarados; nao e valor do leiaute.
	_fimFormaPagamento
)

var formaPagamentoCodigos = [...]string{
	"01", "02", "03", "04", "05", "10", "11", "12", "13", "14", "15", "16",
	"17", "18", "19", "90", "98", "99", "20", "21", "22", "91", "23", "24",
}

// String retorna o codigo do leiaute.
func (v FormaPagamento) String() string { return codigoEm(formaPagamentoCodigos[:], int(v)) }

// ParseFormaPagamento converte o codigo do leiaute no enum.
func ParseFormaPagamento(s string) (FormaPagamento, error) {
	i, err := indiceDe(formaPagamentoCodigos[:], s, "FormaPagamento")
	return FormaPagamento(i), err
}

// ===========================================================================
// BandeiraCartao (TpcnBandeiraCartao) -- tag tBand
// ===========================================================================

// BandeiraCartao e a bandeira da operadora de cartao.
type BandeiraCartao int

const (
	BcVisa            BandeiraCartao = iota // 01
	BcMasterCard                            // 02
	BcAmericanExpress                       // 03
	BcSorocred                              // 04
	BcDinersClub                            // 05
	BcElo                                   // 06
	BcHipercard                             // 07
	BcAura                                  // 08
	BcCabal                                 // 09
	BcAlelo                                 // 10
	BcBanesCard                             // 11
	BcCalCard                               // 12
	BcCredz                                 // 13
	BcDiscover                              // 14
	BcGoodCard                              // 15
	BcGreenCard                             // 16
	BcHiper                                 // 17
	BcJcB                                   // 18
	BcMais                                  // 19
	BcMaxVan                                // 20
	BcPolicard                              // 21
	BcRedeCompras                           // 22
	BcSodexo                                // 23
	BcValeCard                              // 24
	BcVerocheque                            // 25
	BcVR                                    // 26
	BcTicket                                // 27
	BcOutros                                // 99
	// _fimBandeiraCartao conta os membros declarados; nao e valor do leiaute.
	_fimBandeiraCartao

	// BcNenhum e o sentinela TpcnBandeiraCartao(-1) do Delphi.
	BcNenhum BandeiraCartao = -1
)

var bandeiraCartaoCodigos = [...]string{
	"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
	"13", "14", "15", "16", "17", "18", "19", "20", "21", "22", "23", "24",
	"25", "26", "27", "99",
}

// String retorna o codigo do leiaute, ou vazio para BcNenhum.
func (v BandeiraCartao) String() string { return codigoEm(bandeiraCartaoCodigos[:], int(v)) }

// ParseBandeiraCartao converte o codigo do leiaute no enum.
// String vazia resolve para BcNenhum.
func ParseBandeiraCartao(s string) (BandeiraCartao, error) {
	if s == "" {
		return BcNenhum, nil
	}
	i, err := indiceDe(bandeiraCartaoCodigos[:], s, "BandeiraCartao")
	if err != nil {
		return BcVisa, err
	}
	return BandeiraCartao(i), nil
}

// ===========================================================================
// TipoViaTransp (TpcnTipoViaTransp) -- tag tpViaTransp
// ===========================================================================

// TipoViaTransp e a via de transporte internacional da exportacao.
type TipoViaTransp int

const (
	TvMaritima          TipoViaTransp = iota // 1
	TvFluvial                                // 2
	TvLacustre                               // 3
	TvAerea                                  // 4
	TvPostal                                 // 5
	TvFerroviaria                            // 6
	TvRodoviaria                             // 7
	TvConduto                                // 8
	TvMeiosProprios                          // 9
	TvEntradaSaidaFicta                      // 10
	TvCourier                                // 11
	TvEmMaos                                 // 12
	TvPorReboque                             // 13
	// _fimTipoViaTransp conta os membros declarados; nao e valor do leiaute.
	_fimTipoViaTransp
)

var tipoViaTranspCodigos = [...]string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13",
}

// String retorna o codigo do leiaute.
func (v TipoViaTransp) String() string { return codigoEm(tipoViaTranspCodigos[:], int(v)) }

// ParseTipoViaTransp converte o codigo do leiaute no enum.
func ParseTipoViaTransp(s string) (TipoViaTransp, error) {
	i, err := indiceDe(tipoViaTranspCodigos[:], s, "TipoViaTransp")
	return TipoViaTransp(i), err
}

// ===========================================================================
// TipoIntermedio (TpcnTipoIntermedio) -- tag tpIntermedio
// ===========================================================================

// TipoIntermedio e a forma de importacao quanto a intermediacao.
type TipoIntermedio int

const (
	TiContaPropria TipoIntermedio = iota // 1
	TiContaOrdem                         // 2
	TiEncomenda                          // 3
	// _fimTipoIntermedio conta os membros declarados; nao e valor do leiaute.
	_fimTipoIntermedio
)

var tipoIntermedioCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v TipoIntermedio) String() string { return codigoEm(tipoIntermedioCodigos[:], int(v)) }

// ParseTipoIntermedio converte o codigo do leiaute no enum.
func ParseTipoIntermedio(s string) (TipoIntermedio, error) {
	i, err := indiceDe(tipoIntermedioCodigos[:], s, "TipoIntermedio")
	return TipoIntermedio(i), err
}

// ===========================================================================
// IndISSRet (TpcnindISSRet) -- tag indISSRet
// ===========================================================================

// IndISSRet indica se o ISS foi retido na fonte.
//
// Atencao: o "nao" vale "2", nao "0" -- diferente do pcn.Indicador comum.
type IndISSRet int

const (
	IirSim IndISSRet = iota // 1
	IirNao                  // 2
	// _fimIndISSRet conta os membros declarados; nao e valor do leiaute.
	_fimIndISSRet
)

var indISSRetCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v IndISSRet) String() string { return codigoEm(indISSRetCodigos[:], int(v)) }

// ParseIndISSRet converte o codigo do leiaute no enum.
func ParseIndISSRet(s string) (IndISSRet, error) {
	i, err := indiceDe(indISSRetCodigos[:], s, "IndISSRet")
	return IndISSRet(i), err
}

// ===========================================================================
// IndISS (TpcnindISS) -- tag indISS
// ===========================================================================

// IndISS e a exigibilidade do ISS.
type IndISS int

const (
	IiExigivel                IndISS = iota // 1
	IiNaoIncidencia                         // 2
	IiIsencao                               // 3
	IiExportacao                            // 4
	IiImunidade                             // 5
	IiExigSuspDecisaoJudicial               // 6
	IiExigSuspProcessoAdm                   // 7
	// _fimIndISS conta os membros declarados; nao e valor do leiaute.
	_fimIndISS
)

var indISSCodigos = [...]string{"1", "2", "3", "4", "5", "6", "7"}

// String retorna o codigo do leiaute.
func (v IndISS) String() string { return codigoEm(indISSCodigos[:], int(v)) }

// ParseIndISS converte o codigo do leiaute no enum.
func ParseIndISS(s string) (IndISS, error) {
	i, err := indiceDe(indISSCodigos[:], s, "IndISS")
	return IndISS(i), err
}

// ===========================================================================
// TipoAutor (TpcnTipoAutor) -- tag tpAutor
// ===========================================================================

// TipoAutor e o autor do evento.
//
// Os dois ultimos codigos estao FORA de ordem na tabela do ACBr: taOutros
// vale "9" e taEmpresaSucessora, declarado depois, vale "8"
// (pcnConversaoNFe.pas:1097). Ordenar a tabela por codigo trocaria os dois.
type TipoAutor int

const (
	TaEmpresaEmitente     TipoAutor = iota // 1
	TaEmpresaDestinataria                  // 2
	TaEmpresa                              // 3
	TaFisco                                // 5
	TaRFB                                  // 6
	TaOutros                               // 9
	TaEmpresaSucessora                     // 8
	// _fimTipoAutor conta os membros declarados; nao e valor do leiaute.
	_fimTipoAutor
)

var tipoAutorCodigos = [...]string{"1", "2", "3", "5", "6", "9", "8"}

// String retorna o codigo do leiaute.
func (v TipoAutor) String() string { return codigoEm(tipoAutorCodigos[:], int(v)) }

// ParseTipoAutor converte o codigo do leiaute no enum.
func ParseTipoAutor(s string) (TipoAutor, error) {
	i, err := indiceDe(tipoAutorCodigos[:], s, "TipoAutor")
	return TipoAutor(i), err
}

// ===========================================================================
// IndOperacao (TpcnIndOperacao) -- tag indOp do admCscNFCe
// ===========================================================================

// IndOperacao e a operacao solicitada na administracao do CSC da NFC-e.
type IndOperacao int

const (
	IoConsultaCSC IndOperacao = iota // 1
	IoNovoCSC                        // 2
	IoRevogaCSC                      // 3
	// _fimIndOperacao conta os membros declarados; nao e valor do leiaute.
	_fimIndOperacao
)

var indOperacaoCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v IndOperacao) String() string { return codigoEm(indOperacaoCodigos[:], int(v)) }

// ParseIndOperacao converte o codigo do leiaute no enum.
func ParseIndOperacao(s string) (IndOperacao, error) {
	i, err := indiceDe(indOperacaoCodigos[:], s, "IndOperacao")
	return IndOperacao(i), err
}

// ===========================================================================
// IndSomaPISST (TIndSomaPISST) -- tag indSomaPISST
// ===========================================================================

// IndSomaPISST indica se o PIS ST compoe o total da NF-e.
type IndSomaPISST int

const (
	IspNenhum         IndSomaPISST = iota // vazio
	IspPISSTNaoCompoe                     // 0
	IspPISSTCompoe                        // 1
	// _fimIndSomaPISST conta os membros declarados; nao e valor do leiaute.
	_fimIndSomaPISST
)

var indSomaPISSTCodigos = [...]string{"", "0", "1"}

// String retorna o codigo do leiaute.
func (v IndSomaPISST) String() string { return codigoEm(indSomaPISSTCodigos[:], int(v)) }

// ParseIndSomaPISST converte o codigo do leiaute no enum.
func ParseIndSomaPISST(s string) (IndSomaPISST, error) {
	i, err := indiceDe(indSomaPISSTCodigos[:], s, "IndSomaPISST")
	return IndSomaPISST(i), err
}

// ===========================================================================
// IndSomaCOFINSST (TIndSomaCOFINSST) -- tag indSomaCOFINSST
// ===========================================================================

// IndSomaCOFINSST indica se a COFINS ST compoe o total da NF-e.
type IndSomaCOFINSST int

const (
	IscNenhum            IndSomaCOFINSST = iota // vazio
	IscCOFINSSTNaoCompoe                        // 0
	IscCOFINSSTCompoe                           // 1
	// _fimIndSomaCOFINSST conta os membros declarados; nao e valor do leiaute.
	_fimIndSomaCOFINSST
)

var indSomaCOFINSSTCodigos = [...]string{"", "0", "1"}

// String retorna o codigo do leiaute.
func (v IndSomaCOFINSST) String() string {
	return codigoEm(indSomaCOFINSSTCodigos[:], int(v))
}

// ParseIndSomaCOFINSST converte o codigo do leiaute no enum.
func ParseIndSomaCOFINSST(s string) (IndSomaCOFINSST, error) {
	i, err := indiceDe(indSomaCOFINSSTCodigos[:], s, "IndSomaCOFINSST")
	return IndSomaCOFINSST(i), err
}

// ===========================================================================
// FinalidadeNFe (TpcnFinalidadeNFe) -- tag finNFe
// ===========================================================================

// FinalidadeNFe e a finalidade de emissao da NF-e (campo B25 do leiaute).
//
// FnCredito e FnDebito sao da Reforma Tributaria e existem APENAS em
// PCNNFe/pcnConversaoNFe.pas -- Base/ACBrNFe.Conversao.pas para em fnDevolucao.
// O gerador depende deles (Base/ACBrNFe.XmlWriter.pas:546), e e a prova de
// qual das duas units vale.
type FinalidadeNFe int

const (
	FnNormal       FinalidadeNFe = iota // 1
	FnComplementar                      // 2
	FnAjuste                            // 3
	FnDevolucao                         // 4
	FnCredito                           // 5
	FnDebito                            // 6
	// _fimFinalidadeNFe conta os membros declarados; nao e valor do leiaute.
	_fimFinalidadeNFe
)

var finalidadeNFeCodigos = [...]string{"1", "2", "3", "4", "5", "6"}

// String retorna o codigo do leiaute.
func (v FinalidadeNFe) String() string { return codigoEm(finalidadeNFeCodigos[:], int(v)) }

// ParseFinalidadeNFe converte o codigo do leiaute no enum.
func ParseFinalidadeNFe(s string) (FinalidadeNFe, error) {
	i, err := indiceDe(finalidadeNFeCodigos[:], s, "FinalidadeNFe")
	return FinalidadeNFe(i), err
}

// ===========================================================================
// IndicadorNFe (TpcnIndicadorNFe) -- tag indNFe do consNFeDest
// ===========================================================================

// IndicadorNFe filtra a consulta de NF-e destinadas pelo estado de
// manifestacao.
//
// Declarado e nunca exercitado: nenhuma unit do componente le ou grava este
// enum -- a consulta de NF-e destinadas foi desativada pela SEFAZ. Portado
// para manter o inventario completo da unit de origem.
type IndicadorNFe int

const (
	InTodas                     IndicadorNFe = iota // 0
	InSemManifestacaoComCiencia                     // 1
	InSemManifestacaoSemCiencia                     // 2
	// _fimIndicadorNFe conta os membros declarados; nao e valor do leiaute.
	_fimIndicadorNFe
)

var indicadorNFeCodigos = [...]string{"0", "1", "2"}

// String retorna o codigo do leiaute.
func (v IndicadorNFe) String() string { return codigoEm(indicadorNFeCodigos[:], int(v)) }

// ParseIndicadorNFe converte o codigo do leiaute no enum.
func ParseIndicadorNFe(s string) (IndicadorNFe, error) {
	i, err := indiceDe(indicadorNFeCodigos[:], s, "IndicadorNFe")
	return IndicadorNFe(i), err
}

// ===========================================================================
// TipoOperacao (TpcnTipoOperacao) -- tag tpOp do grupo veicProd
// ===========================================================================

// TipoOperacao e o tipo da operacao com veiculo novo (campo J02).
//
// O codigo de toOutros e "0", declarado por ULTIMO -- o ordinal 3 grava "0".
type TipoOperacao int

const (
	ToVendaConcessionaria TipoOperacao = iota // 1
	ToFaturamentoDireto                       // 2
	ToVendaDireta                             // 3
	ToOutros                                  // 0
	// _fimTipoOperacao conta os membros declarados; nao e valor do leiaute.
	_fimTipoOperacao
)

var tipoOperacaoCodigos = [...]string{"1", "2", "3", "0"}

// String retorna o codigo do leiaute.
func (v TipoOperacao) String() string { return codigoEm(tipoOperacaoCodigos[:], int(v)) }

// ParseTipoOperacao converte o codigo do leiaute no enum.
func ParseTipoOperacao(s string) (TipoOperacao, error) {
	i, err := indiceDe(tipoOperacaoCodigos[:], s, "TipoOperacao")
	return TipoOperacao(i), err
}

// ===========================================================================
// CondicaoVeiculo (TpcnCondicaoVeiculo) -- tag condVeic
// ===========================================================================

// CondicaoVeiculo e a condicao do veiculo novo (campo J22).
type CondicaoVeiculo int

const (
	CvAcabado     CondicaoVeiculo = iota // 1
	CvInacabado                          // 2
	CvSemiAcabado                        // 3
	// _fimCondicaoVeiculo conta os membros declarados; nao e valor do leiaute.
	_fimCondicaoVeiculo
)

var condicaoVeiculoCodigos = [...]string{"1", "2", "3"}

// String retorna o codigo do leiaute.
func (v CondicaoVeiculo) String() string { return codigoEm(condicaoVeiculoCodigos[:], int(v)) }

// ParseCondicaoVeiculo converte o codigo do leiaute no enum.
func ParseCondicaoVeiculo(s string) (CondicaoVeiculo, error) {
	i, err := indiceDe(condicaoVeiculoCodigos[:], s, "CondicaoVeiculo")
	return CondicaoVeiculo(i), err
}

// ===========================================================================
// TipoArma (TpcnTipoArma) -- tag tpArma
// ===========================================================================

// TipoArma e o indicador do tipo de arma de fogo (campo L02).
type TipoArma int

const (
	TaUsoPermitido TipoArma = iota // 0
	TaUsoRestrito                  // 1
	// _fimTipoArma conta os membros declarados; nao e valor do leiaute.
	_fimTipoArma
)

var tipoArmaCodigos = [...]string{"0", "1"}

// String retorna o codigo do leiaute.
func (v TipoArma) String() string { return codigoEm(tipoArmaCodigos[:], int(v)) }

// ParseTipoArma converte o codigo do leiaute no enum.
func ParseTipoArma(s string) (TipoArma, error) {
	i, err := indiceDe(tipoArmaCodigos[:], s, "TipoArma")
	return TipoArma(i), err
}

// ===========================================================================
// IndEscala (TpcnIndEscala) -- tag indEscala
// ===========================================================================

// IndEscala indica se o produto e fabricado em escala relevante.
type IndEscala int

const (
	IeRelevante    IndEscala = iota // S
	IeNaoRelevante                  // N
	IeNenhum                        // vazio
	// _fimIndEscala conta os membros declarados; nao e valor do leiaute.
	_fimIndEscala
)

var indEscalaCodigos = [...]string{"S", "N", ""}

// String retorna o codigo do leiaute.
func (v IndEscala) String() string { return codigoEm(indEscalaCodigos[:], int(v)) }

// ParseIndEscala converte o codigo do leiaute no enum. A comparacao ignora
// maiusculas e minusculas, como o AnsiSameText do ACBr -- "s" resolve para
// IeRelevante.
func ParseIndEscala(s string) (IndEscala, error) {
	i, err := indiceDe(indEscalaCodigos[:], s, "IndEscala")
	return IndEscala(i), err
}

// ===========================================================================
// ModalidadeFrete (TpcnModalidadeFrete) -- tag modFrete
// ===========================================================================

// ModalidadeFrete e a modalidade do frete do grupo transp.
type ModalidadeFrete int

const (
	MfContaEmitente       ModalidadeFrete = iota // 0
	MfContaDestinatario                          // 1
	MfContaTerceiros                             // 2
	MfProprioRemetente                           // 3
	MfProprioDestinatario                        // 4
	MfSemFrete                                   // 9
	// _fimModalidadeFrete conta os membros declarados; nao e valor do leiaute.
	_fimModalidadeFrete
)

var modalidadeFreteCodigos = [...]string{"0", "1", "2", "3", "4", "9"}

// String retorna o codigo do leiaute.
func (v ModalidadeFrete) String() string { return codigoEm(modalidadeFreteCodigos[:], int(v)) }

// ParseModalidadeFrete converte o codigo do leiaute no enum.
func ParseModalidadeFrete(s string) (ModalidadeFrete, error) {
	i, err := indiceDe(modalidadeFreteCodigos[:], s, "ModalidadeFrete")
	return ModalidadeFrete(i), err
}

// ===========================================================================
// Autorizacao (TAutorizacao) -- tag indAutoriz
// ===========================================================================

// Autorizacao indica se o destinatario autoriza a operacao.
type Autorizacao int

const (
	TaNaoPermite  Autorizacao = iota // 0
	TaPermite                        // 1
	TaNaoInformar                    // vazio
	// _fimAutorizacao conta os membros declarados; nao e valor do leiaute.
	_fimAutorizacao
)

var autorizacaoCodigos = [...]string{"0", "1", ""}

// String retorna o codigo do leiaute.
func (v Autorizacao) String() string { return codigoEm(autorizacaoCodigos[:], int(v)) }

// ParseAutorizacao converte o codigo do leiaute no enum.
func ParseAutorizacao(s string) (Autorizacao, error) {
	i, err := indiceDe(autorizacaoCodigos[:], s, "Autorizacao")
	return Autorizacao(i), err
}

// ===========================================================================
// IndIntermed (TindIntermed) -- tag indIntermed
// ===========================================================================

// IndIntermed indica a existencia de intermediador na operacao.
type IndIntermed int

const (
	IiSemOperacao              IndIntermed = iota // vazio
	IiOperacaoSemIntermediador                    // 0
	IiOperacaoComIntermediador                    // 1
	// _fimIndIntermed conta os membros declarados; nao e valor do leiaute.
	_fimIndIntermed
)

var indIntermedCodigos = [...]string{"", "0", "1"}

// String retorna o codigo do leiaute.
func (v IndIntermed) String() string { return codigoEm(indIntermedCodigos[:], int(v)) }

// ParseIndIntermed converte o codigo do leiaute no enum.
func ParseIndIntermed(s string) (IndIntermed, error) {
	i, err := indiceDe(indIntermedCodigos[:], s, "IndIntermed")
	return IndIntermed(i), err
}

// ===========================================================================
// TpAto (TtpAto) -- tag tpAto
// ===========================================================================

// TpAto e o tipo de ato concessorio que fundamenta o beneficio fiscal.
type TpAto int

const (
	TaNenhum                TpAto = iota // vazio
	TaTermoAcordo                        // 08
	TaRegimeEspecial                     // 10
	TaAutorizacaoEspecifica              // 12
	TaAjusteSNIEF                        // 14
	TaConvenioICMS                       // 15
	// _fimTpAto conta os membros declarados; nao e valor do leiaute.
	_fimTpAto
)

var tpAtoCodigos = [...]string{"", "08", "10", "12", "14", "15"}

// String retorna o codigo do leiaute.
func (v TpAto) String() string { return codigoEm(tpAtoCodigos[:], int(v)) }

// ParseTpAto converte o codigo do leiaute no enum.
func ParseTpAto(s string) (TpAto, error) {
	i, err := indiceDe(tpAtoCodigos[:], s, "TpAto")
	return TpAto(i), err
}

// ===========================================================================
// IndImport (TindImport) -- tag indImport
// ===========================================================================

// IndImport indica se o bem ou servico e importado.
type IndImport int

const (
	IiNacional  IndImport = iota // 0
	IiImportado                  // 1
	// _fimIndImport conta os membros declarados; nao e valor do leiaute.
	_fimIndImport
)

var indImportCodigos = [...]string{"0", "1"}

// String retorna o codigo do leiaute.
func (v IndImport) String() string { return codigoEm(indImportCodigos[:], int(v)) }

// ParseIndImport converte o codigo do leiaute no enum.
func ParseIndImport(s string) (IndImport, error) {
	i, err := indiceDe(indImportCodigos[:], s, "IndImport")
	return IndImport(i), err
}

// ===========================================================================
// MotRedAdRem (TmotRedAdRem) -- tag motRedAdRem
// ===========================================================================

// MotRedAdRem e o motivo da reducao da aliquota ad rem do IBS/CBS.
type MotRedAdRem int

const (
	MotTranspColetivo MotRedAdRem = iota // 1
	MotOutros                            // 9
	// _fimMotRedAdRem conta os membros declarados; nao e valor do leiaute.
	_fimMotRedAdRem
)

var motRedAdRemCodigos = [...]string{"1", "9"}

// String retorna o codigo do leiaute.
func (v MotRedAdRem) String() string { return codigoEm(motRedAdRemCodigos[:], int(v)) }

// ParseMotRedAdRem converte o codigo do leiaute no enum.
func ParseMotRedAdRem(s string) (MotRedAdRem, error) {
	i, err := indiceDe(motRedAdRemCodigos[:], s, "MotRedAdRem")
	return MotRedAdRem(i), err
}

// ===========================================================================
// TpMotivo (TtpMotivo) -- tag tpMotivo do evento de insucesso de entrega
// ===========================================================================

// TpMotivo e o motivo do insucesso na entrega da NF-e.
type TpMotivo int

const (
	TmNaoEncontrado TpMotivo = iota // 1
	TmRecusa                        // 2
	TmInexistente                   // 3
	TmOutro                         // 4
	// _fimTpMotivo conta os membros declarados; nao e valor do leiaute.
	_fimTpMotivo
)

var tpMotivoCodigos = [...]string{"1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v TpMotivo) String() string { return codigoEm(tpMotivoCodigos[:], int(v)) }

// ParseTpMotivo converte o codigo do leiaute no enum.
func ParseTpMotivo(s string) (TpMotivo, error) {
	i, err := indiceDe(tpMotivoCodigos[:], s, "TpMotivo")
	return TpMotivo(i), err
}

// ===========================================================================
// TpGuia (TtpGuia) -- tag tpGuia do grupo guiaTransito
// ===========================================================================

// TpGuia e o tipo de guia de transito do grupo agropecuario.
type TpGuia int

const (
	TpgNenhum        TpGuia = iota // vazio
	TpgGTA                         // 1
	TpgTTA                         // 2
	TpgDTA                         // 3
	TpgATV                         // 4
	TpgPTV                         // 5
	TpgGTV                         // 6
	TpgGuiaFlorestal               // 7
	// _fimTpGuia conta os membros declarados; nao e valor do leiaute.
	_fimTpGuia
)

var tpGuiaCodigos = [...]string{"", "1", "2", "3", "4", "5", "6", "7"}

// String retorna o codigo do leiaute.
func (v TpGuia) String() string { return codigoEm(tpGuiaCodigos[:], int(v)) }

// ParseTpGuia converte o codigo do leiaute no enum. A comparacao diferencia
// maiusculas de minusculas: o ACBr converte este enum com um laco proprio
// usando "=", nao com StrToEnumerado.
func ParseTpGuia(s string) (TpGuia, error) {
	i, err := indiceDeExato(tpGuiaCodigos[:], s, "TpGuia")
	return TpGuia(i), err
}

// ===========================================================================
// TpNFDebito (TtpNFDebito) -- Reforma Tributaria
// ===========================================================================

// TpNFDebito e o tipo de nota de debito do IBS/CBS.
type TpNFDebito int

const (
	TdNenhum                          TpNFDebito = iota // vazio
	TdTransferenciaCreditoCooperativa                   // 01
	TdAnulacao                                          // 02
	TdDebitosNaoProcessadas                             // 03
	TdMultaJuros                                        // 04
	TdTransferenciaCreditoSucessao                      // 05
	TdPagamentoAntecipado                               // 06
	TdPerdaEmEstoque                                    // 07
	TdDesenquadramentodoSN                              // 08
	// _fimTpNFDebito conta os membros declarados; nao e valor do leiaute.
	_fimTpNFDebito
)

var tpNFDebitoCodigos = [...]string{"", "01", "02", "03", "04", "05", "06", "07", "08"}

// String retorna o codigo do leiaute.
func (v TpNFDebito) String() string { return codigoEm(tpNFDebitoCodigos[:], int(v)) }

// ParseTpNFDebito converte o codigo do leiaute no enum. Comparacao exata,
// como o laco proprio do ACBr.
func ParseTpNFDebito(s string) (TpNFDebito, error) {
	i, err := indiceDeExato(tpNFDebitoCodigos[:], s, "TpNFDebito")
	return TpNFDebito(i), err
}

// ===========================================================================
// TpNFCredito (TtpNFCredito) -- Reforma Tributaria
// ===========================================================================

// TpNFCredito e o tipo de nota de credito do IBS/CBS.
type TpNFCredito int

const (
	TcNenhum                       TpNFCredito = iota // vazio
	TcMultaJuros                                      // 01
	TcApropriacaoCreditoPresumido                     // 02
	TcRetorno                                         // 03
	TcReducaoValores                                  // 04
	TcTransferenciaCreditoSucessao                    // 05
	TcRetornoRecusaParcial                            // 06
	// _fimTpNFCredito conta os membros declarados; nao e valor do leiaute.
	_fimTpNFCredito
)

var tpNFCreditoCodigos = [...]string{"", "01", "02", "03", "04", "05", "06"}

// String retorna o codigo do leiaute.
func (v TpNFCredito) String() string { return codigoEm(tpNFCreditoCodigos[:], int(v)) }

// ParseTpNFCredito converte o codigo do leiaute no enum. Comparacao exata,
// como o laco proprio do ACBr.
func ParseTpNFCredito(s string) (TpNFCredito, error) {
	i, err := indiceDeExato(tpNFCreditoCodigos[:], s, "TpNFCredito")
	return TpNFCredito(i), err
}

// ===========================================================================
// CSTIS (TCSTIS) -- Imposto Seletivo da Reforma Tributaria
// ===========================================================================

// CSTIS e o codigo de situacao tributaria do Imposto Seletivo. Os codigos tem
// TRES digitos e os zeros a esquerda sao significativos ("000", nao "0").
//
// DECLARADO E NAO EXERCITADO. O campo correspondente NAO deve ser tipado com
// este enum: o proprio ACBr desistiu dele ate a publicacao da tabela oficial e
// deixou o campo como string -- ACBrDFe.RTC.Classes.pas:784-785 e :796-797 tem
// a versao TCSTIS comentada e a string ativa, e o mesmo vale em
// ACBrNFe.JSONReader.pas:1816 e JSONWriter.pas:2587.
//
// packages/rtc ja segue isso (ISel.CSTIS e string, rtc/classes.go:397), e le
// valores fora desta tabela -- "001" em rtc_test.go:596. Tipar o campo com
// CSTIS faria o leitor recusar conteudo que o ACBr aceita hoje.
//
// Mantido para o inventario da unit de origem ficar completo.
type CSTIS int

const (
	CstisNenhum CSTIS = iota // vazio
	Cstis000                 // 000
	Cstis100                 // 100
	Cstis200                 // 200
	Cstis300                 // 300
	Cstis400                 // 400
	Cstis500                 // 500
	Cstis600                 // 600
	// _fimCSTIS conta os membros declarados; nao e valor do leiaute.
	_fimCSTIS
)

var cstISCodigos = [...]string{"", "000", "100", "200", "300", "400", "500", "600"}

// String retorna o codigo do leiaute.
func (v CSTIS) String() string { return codigoEm(cstISCodigos[:], int(v)) }

// ParseCSTIS converte o codigo do leiaute no enum. Comparacao exata, como o
// TryStrToCSTIS do ACBr.
func ParseCSTIS(s string) (CSTIS, error) {
	i, err := indiceDeExato(cstISCodigos[:], s, "CSTIS")
	return CSTIS(i), err
}

// ===========================================================================
// IndAceitacao (TIndAceitacao) -- tag indAceitacao dos eventos da Reforma
// ===========================================================================

// IndAceitacao e a manifestacao do contribuinte sobre um debito ou sobre um
// pedido de transferencia de credito.
//
// Vem de ACBrDFe.Conversao.pas:609, e os dois eventos que o usam o emitem
// de formas DIFERENTES no ACBr: o aceite de debito passa o enum direto a um
// no `tcInt` (:2857), e a manifestacao de sucessao passa pelo
// indAceitacaoToStr (:744). Como o Delphi converte o enum pelo ordinal, os
// dois produzem o mesmo "0"/"1" -- a divergencia e so de caminho.
type IndAceitacao int

const (
	IaNaoAceite IndAceitacao = iota // 0
	IaAceite                        // 1
	// _fimIndAceitacao conta os membros declarados; nao e valor do leiaute.
	_fimIndAceitacao
)

var indAceitacaoCodigos = [...]string{"0", "1"}

// String retorna o codigo do leiaute.
func (v IndAceitacao) String() string { return codigoEm(indAceitacaoCodigos[:], int(v)) }

// ParseIndAceitacao converte o codigo do leiaute no enum.
func ParseIndAceitacao(s string) (IndAceitacao, error) {
	i, err := indiceDe(indAceitacaoCodigos[:], s, "IndAceitacao")
	return IndAceitacao(i), err
}
