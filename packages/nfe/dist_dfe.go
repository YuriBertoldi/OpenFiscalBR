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
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Pedido da Distribuicao DF-e. Porte de TDistDFeInt
// (ACBrDFeComum.DistDFeInt.pas).
//
// E o servico pelo qual o contribuinte descobre os documentos emitidos
// CONTRA ele -- notas em que figura como destinatario, e eventos dessas
// notas. Nenhum outro servico da NF-e faz isso: todos os demais partem de
// um documento que o proprio emissor ja tem.
//
// A consulta tem TRES formas mutuamente exclusivas, e a precedencia entre
// elas e o que mais escapa num porte:
//
//	NSU      consulta UM documento pelo numero sequencial. Vence as outras
//	Chave    consulta UM documento pela chave de acesso
//	ultNSU   varre a partir do ultimo NSU conhecido. E o caso comum
//
// Perder a precedencia faz um pedido com NSU e ultNSU preenchidos varrer o
// lote inteiro no lugar de buscar o documento pedido.

// PedidoDistribuicao reune o que a RFB precisa para devolver o lote.
type PedidoDistribuicao struct {
	// CUFAutor e a UF do autor da consulta. Zero nao sai no XML.
	CUFAutor int

	// CNPJCPF e o documento do interessado. O gerador o completa com zeros
	// a esquerda e escolhe a tag pelo TAMANHO resultante.
	CNPJCPF string

	// UltNSU e o ultimo NSU que o consumidor ja processou. A RFB devolve o
	// lote a partir dele -- e e por isso que guardar o maxNSU da resposta
	// anterior e obrigatorio para nao reprocessar.
	UltNSU string

	// NSU pede UM documento especifico. Vence o UltNSU e a Chave.
	NSU string

	// Chave pede UM documento pela chave de acesso. Vence o UltNSU.
	Chave string
}

// tamanhoNSU e o preenchimento com zeros que a RFB exige nos dois campos
// de NSU (Poem_Zeros(..., 15)).
const tamanhoNSU = 15

// GerarXMLDistDFeInt monta o distDFeInt.
//
// Porte de TDistDFeInt.GerarXML. O resultado JA VEM envolvido no
// <nfeDadosMsg>, que e o `tagGrupoMsg` com que o TDistribuicaoDFe constroi
// o objeto (ACBrNFeWebServices.pas:4029) -- e que NAO e o elemento do Body
// do envelope, esse e o `nfeDistDFeInteresse`. Sao duas camadas, e so uma
// delas vem daqui.
func GerarXMLDistDFeInt(p PedidoDistribuicao, ambiente pcn.TipoAmbiente, versao string) string {
	var b strings.Builder
	b.WriteString("<nfeDadosMsg>")
	b.WriteString(`<distDFeInt xmlns="` + namespaceNFe + `" versao="` + versao + `">`)
	b.WriteString("<tpAmb>" + ambiente.String() + "</tpAmb>")

	// O cUFAutor so sai quando positivo (:128).
	if p.CUFAutor > 0 {
		b.WriteString("<cUFAutor>" + strconv.Itoa(p.CUFAutor) + "</cUFAutor>")
	}

	b.WriteString(documentoDoInteressado(p.CNPJCPF))
	b.WriteString(consultaDaDistribuicao(p))
	b.WriteString("</distDFeInt></nfeDadosMsg>")
	return b.String()
}

// documentoDoInteressado escolhe entre <CNPJ> e <CPF>.
//
// Porte de :133-152. A escolha NAO e por um campo separado: e pelo tamanho
// do documento DEPOIS do preenchimento com zeros.
//
//	ate 11 digitos  -> completa ate 11 e sai como <CPF>
//	12 ou 13        -> completa ate 14 e sai como <CNPJ>
//	14              -> sai como <CNPJ>
//
// Documento VAZIO cai no ramo do CPF e sai como <CPF></CPF>, porque o
// tamanho continua zero e nenhum dos dois `if` dispara. E o que o ACBr
// faz, e esta replicado -- a RFB rejeita, que e a resposta correta para um
// pedido sem interessado.
func documentoDoInteressado(documento string) string {
	documento = pcn.OnlyCPFCNPJAlphaNum(strings.TrimSpace(documento))

	switch tam := len(documento); {
	case tam > 0 && tam <= 11:
		documento = pcn.PreencherZerosEsquerda(documento, 11)
	case tam > 0 && tam != 14:
		documento = pcn.PreencherZerosEsquerda(documento, 14)
	}

	if len(documento) == 14 {
		return "<CNPJ>" + documento + "</CNPJ>"
	}
	return "<CPF>" + documento + "</CPF>"
}

// consultaDaDistribuicao monta o grupo que escolhe o que buscar.
//
// Porte de :154-171, onde a precedencia esta aninhada em dois `if`: o NSU
// vence tudo; faltando ele, a Chave vence o ultNSU.
func consultaDaDistribuicao(p PedidoDistribuicao) string {
	if nsu := strings.TrimSpace(p.NSU); nsu != "" {
		return "<consNSU><NSU>" + nsuComZeros(nsu) + "</NSU></consNSU>"
	}
	if chave := strings.TrimSpace(p.Chave); chave != "" {
		return "<consChNFe><chNFe>" + pcn.RemoverLiteralChave(chave) +
			"</chNFe></consChNFe>"
	}
	return "<distNSU><ultNSU>" + nsuComZeros(p.UltNSU) + "</ultNSU></distNSU>"
}

// nsuComZeros formata o NSU com quinze posicoes.
//
// O ACBr converte para inteiro antes de preencher (`Poem_Zeros(StrToIntDef
// (x, 0), 15)`), e por isso um NSU nao numerico vira ZERO em vez de ser
// recusado -- o que, numa varredura, faz o lote recomecar do inicio.
// Replicado: a conversao e deterministica e o consumidor controla a
// entrada.
func nsuComZeros(nsu string) string {
	n, err := strconv.ParseInt(strings.TrimSpace(nsu), 10, 64)
	if err != nil {
		n = 0
	}
	return pcn.PreencherZerosEsquerda(strconv.FormatInt(n, 10), tamanhoNSU)
}
