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
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do pedido da Distribuicao DF-e.

// TestConsultaDaDistribuicao_PrecedenciaDasTresFormas trava o aninhamento
// de :154-171.
//
// Sao TRES formas mutuamente exclusivas, e a precedencia esta em dois `if`
// aninhados: o NSU vence tudo; faltando ele, a Chave vence o ultNSU.
// Perde-la faz um pedido com NSU e ultNSU preenchidos varrer o lote inteiro
// no lugar de buscar o documento pedido.
func TestConsultaDaDistribuicao_PrecedenciaDasTresFormas(t *testing.T) {
	casos := []struct {
		pedido    PedidoDistribuicao
		esperado  string
		descricao string
	}{
		{PedidoDistribuicao{UltNSU: "7"},
			"<distNSU><ultNSU>000000000000007</ultNSU></distNSU>",
			"so ultNSU: varredura"},
		{PedidoDistribuicao{Chave: chaveDeMinas},
			"<consChNFe><chNFe>" + chaveDeMinas + "</chNFe></consChNFe>",
			"so chave"},
		{PedidoDistribuicao{NSU: "42"},
			"<consNSU><NSU>000000000000042</NSU></consNSU>",
			"so NSU"},
		{PedidoDistribuicao{NSU: "42", Chave: chaveDeMinas, UltNSU: "7"},
			"<consNSU><NSU>000000000000042</NSU></consNSU>",
			"os tres: o NSU vence"},
		{PedidoDistribuicao{Chave: chaveDeMinas, UltNSU: "7"},
			"<consChNFe><chNFe>" + chaveDeMinas + "</chNFe></consChNFe>",
			"chave e ultNSU: a chave vence"},
		{PedidoDistribuicao{},
			"<distNSU><ultNSU>000000000000000</ultNSU></distNSU>",
			"nada: varredura do inicio"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			if got := consultaDaDistribuicao(c.pedido); got != c.esperado {
				t.Errorf("= %q, esperado %q", got, c.esperado)
			}
		})
	}
}

// TestNSUComZeros_QuinzePosicoes trava o Poem_Zeros(..., 15).
//
// O ACBr converte para INTEIRO antes de preencher, e por isso um NSU nao
// numerico vira ZERO em vez de ser recusado -- numa varredura, isso faz o
// lote recomecar do inicio. Replicado, e o teste registra o porque.
func TestNSUComZeros_QuinzePosicoes(t *testing.T) {
	casos := []struct{ entrada, esperado string }{
		{"", "000000000000000"},
		{"0", "000000000000000"},
		{"7", "000000000000007"},
		{"000000000000042", "000000000000042"},
		{" 42 ", "000000000000042"},
		{"999999999999999", "999999999999999"},
		// Nao numerico vira zero -- e o StrToIntDef(x, 0) do original.
		{"abc", "000000000000000"},
	}
	for _, c := range casos {
		got := nsuComZeros(c.entrada)
		if got != c.esperado {
			t.Errorf("nsuComZeros(%q) = %q, esperado %q", c.entrada, got, c.esperado)
		}
		if len(got) != tamanhoNSU {
			t.Errorf("nsuComZeros(%q) tem %d posicoes, esperado %d",
				c.entrada, len(got), tamanhoNSU)
		}
	}
}

// TestDocumentoDoInteressado_EscolhePeloTAMANHO trava :133-152.
//
// A escolha entre CNPJ e CPF nao e por um campo separado: e pelo tamanho
// DEPOIS do preenchimento com zeros, e o preenchimento tem dois alvos
// diferentes conforme o tamanho de entrada.
func TestDocumentoDoInteressado_EscolhePeloTAMANHO(t *testing.T) {
	casos := []struct {
		entrada   string
		esperado  string
		descricao string
	}{
		{"11122233396", "<CPF>11122233396</CPF>", "CPF de 11 digitos"},
		{"1122233396", "<CPF>01122233396</CPF>", "10 digitos completa ate 11"},
		{"1", "<CPF>00000000001</CPF>", "1 digito completa ate 11"},
		{"99999999000191", "<CNPJ>99999999000191</CNPJ>", "CNPJ de 14"},
		{"9999999000191", "<CNPJ>09999999000191</CNPJ>", "13 completa ate 14"},
		{"111222333961", "<CNPJ>00111222333961</CNPJ>", "12 completa ate 14"},
		{"99.999.999/0001-91", "<CNPJ>99999999000191</CNPJ>", "com mascara"},
		// Vazio cai no ramo do CPF, porque nenhum dos dois `if` dispara e o
		// tamanho continua zero. A RFB rejeita, que e a resposta certa para
		// um pedido sem interessado -- mas o XML sai assim.
		{"", "<CPF></CPF>", "vazio sai como CPF vazio"},
	}
	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			if got := documentoDoInteressado(c.entrada); got != c.esperado {
				t.Errorf("= %q, esperado %q", got, c.esperado)
			}
		})
	}
}

// TestGerarXMLDistDFeInt_DuasCamadasEOrdem trava o envelope e a sequencia.
//
// O `nfeDadosMsg` vem do GERADOR da mensagem (:4029), nao do envelope --
// e o elemento do Body e outro, o `nfeDistDFeInteresse`. Confundir as duas
// camadas produz um pedido que a RFB nao entende.
func TestGerarXMLDistDFeInt_DuasCamadasEOrdem(t *testing.T) {
	xml := GerarXMLDistDFeInt(PedidoDistribuicao{
		CUFAutor: 31,
		CNPJCPF:  "99999999000191",
		UltNSU:   "7",
	}, pcn.TaHomologacao, "1.01")

	if !strings.HasPrefix(xml, "<nfeDadosMsg><distDFeInt ") {
		t.Errorf("a mensagem comeca pelo nfeDadosMsg: %s", xml)
	}
	if !strings.HasSuffix(xml, "</distDFeInt></nfeDadosMsg>") {
		t.Errorf("e fecha pelos dois: %s", xml)
	}
	if strings.Contains(xml, "nfeDistDFeInteresse") {
		t.Error("o nfeDistDFeInteresse e o elemento do BODY, nao da mensagem")
	}

	conferirOrdem(t, xml, "<tpAmb>", "<cUFAutor>", "<CNPJ>", "<distNSU>")
	if !strings.Contains(xml, `versao="1.01"`) {
		t.Errorf("versao: %s", xml)
	}
}

// TestGerarXMLDistDFeInt_CUFAutorSoQuandoPOSITIVO trava :128.
func TestGerarXMLDistDFeInt_CUFAutorSoQuandoPOSITIVO(t *testing.T) {
	for _, c := range []struct {
		cUF    int
		temTag bool
	}{
		{31, true},
		{0, false},
		{-1, false},
	} {
		xml := GerarXMLDistDFeInt(PedidoDistribuicao{
			CUFAutor: c.cUF, CNPJCPF: "99999999000191",
		}, pcn.TaHomologacao, "1.01")
		if got := strings.Contains(xml, "<cUFAutor>"); got != c.temTag {
			t.Errorf("cUFAutor %d: tag presente = %v, esperado %v", c.cUF, got, c.temTag)
		}
	}
}
