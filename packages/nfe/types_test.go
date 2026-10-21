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
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// TestTabelasDeControle_CobremTodoMembroDeclarado estende a conferencia de
// tabela curta (ver TestEnums_TabelaCobreTodoMembroDeclarado) aos tipos de
// controle, que nao entram na lista generica de todosOsEnums.
func TestTabelasDeControle_CobremTodoMembroDeclarado(t *testing.T) {
	casos := []struct {
		nome     string
		declarou int
		tabela   int
	}{
		{"StatusNFe", int(_fimStatusNFe), len(statusNFeNomes)},
		{"SchemaNFe", int(_fimSchemaNFe), len(schemaNFeNomes)},
		{"LayOut", int(_fimLayOut), len(layOutServicos)},
		{"VersaoDF (codigos)", int(_fimVersaoDF), len(versaoDFCodigos)},
		{"VersaoDF (valores)", int(_fimVersaoDF), len(versaoDFValores)},
	}
	for _, c := range casos {
		if c.declarou != c.tabela {
			t.Errorf("%s: %d membros declarados no const, %d entradas na tabela",
				c.nome, c.declarou, c.tabela)
		}
	}
}

// TestTpNFDebito_OrdinaisAlinhadosComRTC trava a duplicacao do enum entre este
// package e packages/rtc.
//
// O ACBr declara TtpNFDebito duas vezes -- em ACBrDFe.Conversao.pas (de onde
// veio rtc.TpNFDebito) e em pcnConversaoNFe.pas (de onde veio o daqui) -- com
// os mesmos 9 membros. Em Go sao dois TIPOS distintos: o do rtc e so a chave
// que decide a emissao do grupo RTC (rtc/xml_writer.go:170) e nao tem
// String()/Parse; o daqui carrega o codigo do leiaute.
//
// A etapa do gerador vai precisar converter um no outro, e a conversao sera
// por ordinal. Se um dos dois ganhar um membro no meio da lista, a conversao
// passa a mapear o valor errado SEM erro de compilacao -- o defeito-tipo deste
// projeto, so que na fronteira entre packages. Este teste e o que o impede.
func TestTpNFDebito_OrdinaisAlinhadosComRTC(t *testing.T) {
	pares := []struct {
		nome  string
		aqui  TpNFDebito
		noRTC rtc.TpNFDebito
	}{
		{"TdNenhum", TdNenhum, rtc.TdNenhum},
		{"TdTransferenciaCreditoCooperativa", TdTransferenciaCreditoCooperativa, rtc.TdTransferenciaCreditoCooperativa},
		{"TdAnulacao", TdAnulacao, rtc.TdAnulacao},
		{"TdDebitosNaoProcessadas", TdDebitosNaoProcessadas, rtc.TdDebitosNaoProcessadas},
		{"TdMultaJuros", TdMultaJuros, rtc.TdMultaJuros},
		{"TdTransferenciaCreditoSucessao", TdTransferenciaCreditoSucessao, rtc.TdTransferenciaCreditoSucessao},
		{"TdPagamentoAntecipado", TdPagamentoAntecipado, rtc.TdPagamentoAntecipado},
		{"TdPerdaEmEstoque", TdPerdaEmEstoque, rtc.TdPerdaEmEstoque},
		{"TdDesenquadramentodoSN", TdDesenquadramentodoSN, rtc.TdDesenquadramentodoSN},
	}

	if len(pares) != int(_fimTpNFDebito) {
		t.Fatalf("o enum daqui tem %d membros, mas o teste compara %d",
			int(_fimTpNFDebito), len(pares))
	}
	for _, p := range pares {
		if int(p.aqui) != int(p.noRTC) {
			t.Errorf("%s: ordinal %d aqui e %d em packages/rtc",
				p.nome, int(p.aqui), int(p.noRTC))
		}
	}
}

// TestSchemaNFe_NomeVemDaRTTIENaoDoArrayMorto trava a escolha entre as duas
// tabelas de nome de schema que o ACBr mantem, e que divergem.
//
// SchemaNFeToStr usa RTTI (GetEnumName + Copy, pcnConversaoNFe.pas:1237), e e
// a unica chamada por ACBrNFe.pas. TSchemaNFeArrayStrings nao e referenciada
// por nenhuma unit da NFe. Nos dois membros abaixo as tabelas discordam, e o
// nome do schema decide qual XSD o componente procura em disco.
func TestSchemaNFe_NomeVemDaRTTIENaoDoArrayMorto(t *testing.T) {
	casos := []struct {
		schema          SchemaNFe
		esperado        string
		arrayMortoDizia string
	}{
		{SchInsucessoEntregaNFe, "InsucessoEntregaNFe", "InsucessoEntrega"},
		{SchCancInsucessoEntregaNFe, "CancInsucessoEntregaNFe", "CancInsucessoEntrega"},
	}

	for _, c := range casos {
		got := c.schema.String()
		if got == c.arrayMortoDizia {
			t.Errorf("schema resolveu para %q, nome do TSchemaNFeArrayStrings (morto); "+
				"SchemaNFeToStr devolve %q", got, c.esperado)
			continue
		}
		if got != c.esperado {
			t.Errorf("schema resolveu para %q, esperado %q", got, c.esperado)
		}
	}
}

// TestSchemaNFe_TodoMembroTemNomeDistinto pega tabela mais curta que o enum
// (os ultimos schemas resolveriam para vazio e o componente procuraria um
// arquivo XSD sem nome) e nome repetido por copia-e-cola.
func TestSchemaNFe_TodoMembroTemNomeDistinto(t *testing.T) {
	vistos := make(map[string]SchemaNFe)
	for s := SchErro; s <= SchAtualizacaoDataPrevisaoEntrega; s++ {
		nome := s.String()
		if nome == "" {
			t.Errorf("schema de ordinal %d nao tem nome: tabela mais curta que o enum", int(s))
			continue
		}
		if anterior, repetido := vistos[nome]; repetido {
			t.Errorf("schemas de ordinal %d e %d compartilham o nome %q",
				int(anterior), int(s), nome)
		}
		vistos[nome] = s
	}
}

// TestSchemaNFe_SchemaEventoSoRespondeParaEvento confirma que o mapa de
// codigos de evento cobre os eventos e devolve vazio para o resto, como o
// EnumeradoToStr de SchemaEventoToStr.
func TestSchemaNFe_SchemaEventoSoRespondeParaEvento(t *testing.T) {
	comEvento := map[SchemaNFe]string{
		SchcancNFe:                        "e110111",
		SchEnvCCe:                         "e110110",
		SchCancSubst:                      "e110112",
		SchEnvEPEC:                        "e110140",
		SchCancGenerico:                   "e110001",
		SchManifDestConfirmacao:           "e210200",
		SchAtualizacaoDataPrevisaoEntrega: "e112150",
	}
	for schema, esperado := range comEvento {
		if got := schema.SchemaEvento(); got != esperado {
			t.Errorf("%v.SchemaEvento() = %q, esperado %q", schema, got, esperado)
		}
	}

	semEvento := []SchemaNFe{SchErro, SchNfe, SchInutNFe, SchconsStatServ, SchdistDFeInt}
	for _, schema := range semEvento {
		if got := schema.SchemaEvento(); got != "" {
			t.Errorf("%v nao e schema de evento, mas devolveu %q", schema, got)
		}
	}
}

// TestLayOut_ServicoDosTresLayoutsDeURLEhVazio trava uma omissao do ACBr que
// e facil "consertar" por engano.
//
// LayOutToServico (pcnConversaoNFe.pas:1180) mapeia so os 17 primeiros
// layouts. Os tres de URL ficam de fora e resolvem para vazio. Existe um
// TLayOutArrayStrings com nomes para eles ("URL-QRCode" etc.), mas ele esta
// morto -- a URL de QR-Code e de consulta e resolvida por GetURLQRCode e
// GetURLConsulta, nao pelo nome de servico.
func TestLayOut_ServicoDosTresLayoutsDeURLEhVazio(t *testing.T) {
	for _, l := range []LayOut{LayNFeURLQRCode, LayURLConsultaNFe, LayURLConsultaNFCe} {
		if got := l.Servico(); got != "" {
			t.Errorf("layout de ordinal %d devolveu servico %q; LayOutToServico nao o mapeia",
				int(l), got)
		}
	}

	// Os demais precisam ter nome, senao a secao do ACBrNFeServicos.ini
	// nao e encontrada e o servico fica sem URL.
	for l := LayNfeRecepcao; l <= LayNFCeEPEC; l++ {
		if l.Servico() == "" {
			t.Errorf("layout de ordinal %d ficou sem nome de servico", int(l))
		}
	}
}

// TestLayOut_RecepcaoEventoResolveSempreParaCCe cobre a assimetria dos tres
// layouts de evento, que compartilham o servico "RecepcaoEvento".
//
// StrToEnumerado para na PRIMEIRA ocorrencia, entao o caminho de volta nunca
// chega a LayNFeEvento nem LayNFeEventoAN. E perda de informacao do ACBr,
// replicada: os tres apontam para o mesmo servico e para o mesmo schema, e
// por isso a ida e o que importa.
func TestLayOut_RecepcaoEventoResolveSempreParaCCe(t *testing.T) {
	for _, l := range []LayOut{LayNFeCCe, LayNFeEvento, LayNFeEventoAN} {
		if got := l.Servico(); got != "RecepcaoEvento" {
			t.Errorf("layout de ordinal %d = %q, esperado \"RecepcaoEvento\"", int(l), got)
		}
		if got := l.Schema(); got != SchenvEvento {
			t.Errorf("layout de ordinal %d resolve para o schema %v, esperado SchenvEvento",
				int(l), got)
		}
	}

	v, err := ParseLayOutServico("RecepcaoEvento")
	if err != nil {
		t.Fatalf("ParseLayOutServico(\"RecepcaoEvento\") devolveu erro: %v", err)
	}
	if v != LayNFeCCe {
		t.Errorf("ParseLayOutServico(\"RecepcaoEvento\") = %d, esperado LayNFeCCe (primeira ocorrencia)",
			int(v))
	}
}

// TestLayOut_SchemaDeLayoutForaDaTabelaEhErro confirma o case-else de
// LayOutToSchema. Os tres layouts de URL nao validam mensagem nenhuma.
func TestLayOut_SchemaDeLayoutForaDaTabelaEhErro(t *testing.T) {
	for _, l := range []LayOut{LayNFeURLQRCode, LayURLConsultaNFe, LayURLConsultaNFCe} {
		if got := l.Schema(); got != SchErro {
			t.Errorf("layout de ordinal %d resolve para o schema %v, esperado SchErro",
				int(l), got)
		}
	}
}

// TestParseVersaoDFFloat_ResolvePorFaixaENaoPorIgualdade cobre o porte de
// DblToVersaoDF, que classifica por intervalo aberto nas duas pontas.
//
// Uma implementacao "obvia" por igualdade (d == 3.10) recusaria 3.01 e 4.01,
// que no ACBr sao validos.
func TestParseVersaoDFFloat_ResolvePorFaixaENaoPorIgualdade(t *testing.T) {
	casos := []struct {
		entrada  float64
		esperado VersaoDF
	}{
		{0, Ve200},    // abaixo de 3.00 cai em Ve200, inclusive zero
		{1.5, Ve200},  // idem
		{2.00, Ve200}, // o ramo explicito d = 2.0
		{2.99, Ve200},
		{3.00, Ve300},
		{3.09, Ve300}, // ainda abaixo de 3.10
		{3.10, Ve310},
		{3.99, Ve310},
		{4.00, Ve400},
		{5.00, Ve400}, // versao futura cai em Ve400, nao em erro
	}
	for _, c := range casos {
		if got := ParseVersaoDFFloat(c.entrada); got != c.esperado {
			t.Errorf("ParseVersaoDFFloat(%v) = %v, esperado %v", c.entrada, got, c.esperado)
		}
	}
}

// TestVersaoDF_FloatEhOInversoDeParsePorFaixa confirma que o numero gravado
// por Float volta a mesma versao -- o par usado para ler e gravar a versao do
// servico no ACBrNFeServicos.ini.
func TestVersaoDF_FloatEhOInversoDeParsePorFaixa(t *testing.T) {
	for v := Ve200; v <= Ve400; v++ {
		if volta := ParseVersaoDFFloat(v.Float()); volta != v {
			t.Errorf("%v.Float() = %v, que volta como %v", v, v.Float(), volta)
		}
	}
}

// TestAjustarVersaoQrCode_NaoExisteQRCodeAteALeiaute310 cobre a regra de
// AjustarVersaoQRCode (ACBrNFe.pas:497): o QR-Code so existe a partir da
// versao 4.00, e dali em diante a versao minima e 1.
func TestAjustarVersaoQrCode_NaoExisteQRCodeAteALeiaute310(t *testing.T) {
	casos := []struct {
		versaoQrCode VersaoQrCode
		versaoXML    VersaoDF
		esperado     VersaoQrCode
	}{
		// Ate a 3.10 o QR-Code e zerado, mesmo pedido explicitamente.
		{Veqr300, Ve200, Veqr000},
		{Veqr300, Ve300, Veqr000},
		{Veqr300, Ve310, Veqr000},

		// Da 4.00 em diante, piso em Veqr100.
		{Veqr000, Ve400, Veqr100},
		{Veqr100, Ve400, Veqr100},
		{Veqr200, Ve400, Veqr200},
		{Veqr300, Ve400, Veqr300},
	}
	for _, c := range casos {
		got := AjustarVersaoQrCode(c.versaoQrCode, c.versaoXML)
		if got != c.esperado {
			t.Errorf("AjustarVersaoQrCode(%v, %v) = %v, esperado %v",
				c.versaoQrCode, c.versaoXML, got, c.esperado)
		}
	}
}

// TestModeloDF_PrefixoDistingueNFeDeNFCe confirma ModeloDFToPrefixo, que
// decide o nome do arquivo gerado e o prefixo de tag.
func TestModeloDF_PrefixoDistingueNFeDeNFCe(t *testing.T) {
	if got := MoNFe.Prefixo(); got != "NFe" {
		t.Errorf("MoNFe.Prefixo() = %q, esperado \"NFe\"", got)
	}
	if got := MoNFCe.Prefixo(); got != "NFCe" {
		t.Errorf("MoNFCe.Prefixo() = %q, esperado \"NFCe\"", got)
	}
}
