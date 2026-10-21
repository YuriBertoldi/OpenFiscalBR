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
	"fmt"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Pedido de inutilizacao de faixa de numeracao.
// Porte de TinutNFe (ACBrNFe.Inut.pas) e das validacoes que o
// TNFeInutilizacao faz antes de montar a mensagem.
//
// Inutilizar NAO e cancelar: o pedido declara que uma faixa de numeros
// nunca foi usada e nunca sera. Por isso ele nao tem chave de acesso --
// nao existe documento -- e o identificador e montado dos proprios dados
// do pedido.

// PedidoInutilizacao reune o que a SEFAZ precisa para inutilizar uma faixa.
type PedidoInutilizacao struct {
	// Ano e o ano da faixa. Aceita tanto quatro digitos quanto dois: o
	// gerador reduz para dois, como o original.
	Ano int

	// CNPJ aceita tambem CPF, mas so no Mato Grosso do Sul -- ver
	// gerarDocumentoInutilizacao.
	CNPJ string

	// Modelo e 55 ou 65; Serie, NumeroInicial e NumeroFinal delimitam a
	// faixa.
	Modelo        int
	Serie         int
	NumeroInicial int
	NumeroFinal   int

	// Justificativa tem minimo de quinze caracteres (SetJustificativa,
	// ACBrNFeWebServices.pas:2624).
	Justificativa string
}

// cUFMatoGrossoDoSul e a unica UF que aceita CPF no pedido de inutilizacao
// (ACBrNFe.Inut.pas: `if cUF in [51]`).
const cUFMatoGrossoDoSul = 51

// tamanhoMinimoJustificativa vem de TNFeInutilizacao.SetJustificativa.
const tamanhoMinimoJustificativa = 15

// IDInutilizacao monta o identificador do pedido.
//
// Porte da primeira metade de TinutNFe.GerarXML (ACBrNFe.Inut.pas:124). Sao
// 43 posicoes: "ID" + cUF(2) + ano(2) + CNPJ(14) + modelo(2) + serie(3) +
// nNFIni(9) + nNFFin(9).
//
// O CPF do Mato Grosso do Sul entra aqui com TRES ZEROS a esquerda, para
// ocupar as mesmas catorze posicoes do CNPJ. No CORPO da mensagem ele vai
// sem o preenchimento -- o original recalcula o documento depois de montar
// o ID justamente para desfazer o padding (:144).
func IDInutilizacao(p PedidoInutilizacao, cUF int) string {
	documento := pcn.OnlyCPFCNPJAlphaNum(p.CNPJ)
	if cUF == cUFMatoGrossoDoSul && len(documento) == 11 {
		documento = "000" + documento
	}
	return fmt.Sprintf("ID%02d%s%s%02d%03d%09d%09d",
		cUF, anoDeDoisDigitos(p.Ano), documento,
		p.Modelo, p.Serie, p.NumeroInicial, p.NumeroFinal)
}

// anoDeDoisDigitos devolve as duas ultimas casas do ano com quatro
// posicoes, que e o `Copy(IntToStrZero(ano, 4), 3, 2)` do original.
func anoDeDoisDigitos(ano int) string {
	return fmt.Sprintf("%04d", ano)[2:]
}

// GerarXMLInutilizacao monta o inutNFe.
//
// Porte de TinutNFe.GerarXML somado ao TNFeInutilizacao.DefinirDadosMsg: o
// cUF e o tpAmb sao do COMPONENTE, nao do pedido (ACBrNFeWebServices.pas:
// 2732-2733) -- inutiliza-se na SEFAZ onde se emite.
func GerarXMLInutilizacao(p PedidoInutilizacao, cfg Configuracoes) (string, error) {
	if err := ValidarPedidoInutilizacao(p); err != nil {
		return "", err
	}

	// A ordem abaixo e a do xs:sequence do leiaute, e e a mesma do
	// TinutNFe.GerarXML. Trocar duas linhas de lugar produz um XML que
	// passa em qualquer round-trip e a SEFAZ rejeita.
	inf := pcn.NovoElem("infInut").
		Attr("Id", IDInutilizacao(p, cfg.CodigoUF)).
		Filho(pcn.NovoElem("tpAmb").Texto(cfg.Ambiente.String())).
		Filho(pcn.NovoElem("xServ").Texto("INUTILIZAR")).
		Filho(pcn.NovoElem("cUF").Texto(strconv.Itoa(cfg.CodigoUF))).
		Filho(pcn.NovoElem("ano").Texto(anoReduzido(p.Ano))).
		Filho(gerarDocumentoInutilizacao(p.CNPJ, cfg.CodigoUF)).
		Filho(pcn.NovoElem("mod").Texto(strconv.Itoa(p.Modelo))).
		Filho(pcn.NovoElem("serie").Texto(strconv.Itoa(p.Serie))).
		Filho(pcn.NovoElem("nNFIni").Texto(strconv.Itoa(p.NumeroInicial))).
		Filho(pcn.NovoElem("nNFFin").Texto(strconv.Itoa(p.NumeroFinal))).
		Filho(pcn.NovoElem("xJust").Texto(strings.TrimSpace(p.Justificativa)))

	return pcn.NovoElem("inutNFe").
		Attr("xmlns", namespaceNFe).
		Attr("versao", pcn.FormatarVersaoXML(cfg.VersaoDF.Float())).
		Filho(inf).
		XML(), nil
}

// anoReduzido e o `if ano > 2000 then ano := ano - 2000` do original: o
// corpo leva o ano com DUAS casas, sem zero a esquerda.
//
// O limite e `> 2000`, nao `>= 2000`: o ano 2000 sairia inteiro no corpo e
// como "00" no ID. E uma incoerencia do ACBr sem efeito pratico -- a NF-e
// comecou em 2006 -- e esta replicada de proposito.
func anoReduzido(ano int) string {
	if ano > 2000 {
		ano -= 2000
	}
	return fmt.Sprintf("%d", ano)
}

// gerarDocumentoInutilizacao escolhe entre <CNPJ> e <CPF>.
//
// So o Mato Grosso do Sul aceita CPF, e a escolha e pelo TAMANHO do
// documento informado, nao por um campo separado (ACBrNFe.Inut.pas:141).
func gerarDocumentoInutilizacao(documento string, cUF int) *pcn.Elem {
	documento = pcn.OnlyCPFCNPJAlphaNum(documento)
	if cUF == cUFMatoGrossoDoSul && len(documento) == 11 {
		return pcn.NovoElem("CPF").Texto(documento)
	}
	return pcn.NovoElem("CNPJ").Texto(documento)
}

// ValidarPedidoInutilizacao confere o que o componente confere antes de
// transmitir.
//
// Porte de TNFeInutilizacao.SetJustificativa (ACBrNFeWebServices.pas:2624),
// que e a unica validacao que o ACBr faz no caminho de inutilizacao -- as
// demais ele deixou comentadas no gerador (ACBrNFe.Inut.pas:131-162) e
// portanto a SEFAZ e quem rejeita.
func ValidarPedidoInutilizacao(p PedidoInutilizacao) error {
	justificativa := strings.TrimSpace(p.Justificativa)
	if justificativa == "" {
		return ErrJustificativaObrigatoria
	}
	if len([]rune(justificativa)) < tamanhoMinimoJustificativa {
		return fmt.Errorf("%w: %d caracteres, minimo %d",
			ErrJustificativaCurta, len([]rune(justificativa)), tamanhoMinimoJustificativa)
	}
	return nil
}
