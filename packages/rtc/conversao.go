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

package rtc

import (
	"errors"
	"fmt"
)

// Enums da Reforma Tributaria. Porte das declaracoes correspondentes de
// ACBrDFe.Conversao.pas.

// ErrEnumInvalido indica valor de string que nao corresponde a nenhum
// membro do enum.
var ErrEnumInvalido = errors.New("rtc: valor invalido para enum")

func erroEnum(enum, valor string) error {
	return fmt.Errorf("%w: %s nao aceita %q", ErrEnumInvalido, enum, valor)
}

// ===========================================================================
// TpEnteGov (TtpEnteGov)
// ===========================================================================

// TpEnteGov identifica o ente governamental da compra.
type TpEnteGov int

const (
	TcgNenhum           TpEnteGov = iota // vazio
	TcgUniao                             // 1
	TcgEstados                           // 2
	TcgDistritoFederal                   // 3
	TcgMunicipios                        // 4
	TcgConsorcioPublico                  // 5
	TcgComiteGestorIBS                   // 6
)

var tpEnteGovCodigos = [...]string{"", "1", "2", "3", "4", "5", "6"}

// String retorna o codigo do leiaute.
func (v TpEnteGov) String() string { return codigo(tpEnteGovCodigos[:], int(v)) }

// ParseTpEnteGov converte o codigo do leiaute no enum. Vazio e valido.
func ParseTpEnteGov(s string) (TpEnteGov, error) {
	i, err := indice(tpEnteGovCodigos[:], s, "TpEnteGov")
	return TpEnteGov(i), err
}

// ===========================================================================
// TpOperGov (TtpOperGov)
// ===========================================================================

// TpOperGov identifica o tipo de operacao governamental.
type TpOperGov int

const (
	TogNenhum                        TpOperGov = iota // vazio
	TogFornecimento                                   // 1
	TogRecebimentoPag                                 // 2
	TogFornecimentoPagRealizado                       // 3
	TogRecebimentoPagFornecPosterior                  // 4
)

var tpOperGovCodigos = [...]string{"", "1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v TpOperGov) String() string { return codigo(tpOperGovCodigos[:], int(v)) }

// ParseTpOperGov converte o codigo do leiaute no enum. Vazio e valido.
func ParseTpOperGov(s string) (TpOperGov, error) {
	i, err := indice(tpOperGovCodigos[:], s, "TpOperGov")
	return TpOperGov(i), err
}

// ===========================================================================
// CSTIBSCBS (TCSTIBSCBS)
// ===========================================================================

// CSTIBSCBS e o codigo de situacao tributaria do IBS e da CBS.
type CSTIBSCBS int

const (
	CSTNenhum CSTIBSCBS = iota // vazio
	CST000                     // 000
	CST010                     // 010
	CST011                     // 011
	CST200                     // 200
	CST220                     // 220
	CST221                     // 221
	CST222                     // 222
	CST400                     // 400
	CST410                     // 410
	CST510                     // 510
	CST515                     // 515
	CST550                     // 550
	CST620                     // 620
	CST800                     // 800
	CST810                     // 810
	CST811                     // 811
	CST820                     // 820
	CST830                     // 830
)

var cstIBSCBSCodigos = [...]string{
	"", "000", "010", "011", "200", "220", "221", "222", "400", "410",
	"510", "515", "550", "620", "800", "810", "811", "820", "830",
}

// String retorna o codigo de tres digitos do leiaute.
func (v CSTIBSCBS) String() string { return codigo(cstIBSCBSCodigos[:], int(v)) }

// ParseCSTIBSCBS converte o codigo do leiaute no enum. Vazio e valido.
func ParseCSTIBSCBS(s string) (CSTIBSCBS, error) {
	i, err := indice(cstIBSCBSCodigos[:], s, "CSTIBSCBS")
	return CSTIBSCBS(i), err
}

// ===========================================================================
// TpALCZFMCBS (TtpALCZFMCBS)
// ===========================================================================

// TpALCZFMCBS identifica o tipo de operacao na Zona Franca de Manaus e
// Areas de Livre Comercio, para fins de CBS.
//
// Atencao: diferente da maioria dos enums da RTC, este NAO tem membro para
// string vazia -- o primeiro membro ja vale "1".
type TpALCZFMCBS int

const (
	TpALCZFMCBSnOpInd TpALCZFMCBS = iota // 1
	TpALCZFMCBSOpInd                     // 2
)

var tpALCZFMCBSCodigos = [...]string{"1", "2"}

// String retorna o codigo do leiaute.
func (v TpALCZFMCBS) String() string { return codigo(tpALCZFMCBSCodigos[:], int(v)) }

// ParseTpALCZFMCBS converte o codigo do leiaute no enum.
func ParseTpALCZFMCBS(s string) (TpALCZFMCBS, error) {
	i, err := indice(tpALCZFMCBSCodigos[:], s, "TpALCZFMCBS")
	return TpALCZFMCBS(i), err
}

// ===========================================================================
// CCredPres (TcCredPres)
// ===========================================================================

// CCredPres e o codigo do credito presumido.
type CCredPres int

const (
	CpNenhum CCredPres = iota // vazio
	Cp01
	Cp02
	Cp03
	Cp04
	Cp05
	Cp06
	Cp07
	Cp08
	Cp09
	Cp10
	Cp11
	Cp12
	Cp13
)

var cCredPresCodigos = [...]string{
	"", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12", "13",
}

// String retorna o codigo de dois digitos do leiaute.
func (v CCredPres) String() string { return codigo(cCredPresCodigos[:], int(v)) }

// ParseCCredPres converte o codigo do leiaute no enum. Vazio e valido.
func ParseCCredPres(s string) (CCredPres, error) {
	i, err := indice(cCredPresCodigos[:], s, "CCredPres")
	return CCredPres(i), err
}

// ===========================================================================
// TpCredPresIBSZFM (TTpCredPresIBSZFM)
// ===========================================================================

// TpCredPresIBSZFM e o tipo de credito presumido do IBS na Zona Franca de
// Manaus.
//
// Atencao ao deslocamento: o membro TcpSemCredito vale "0", nao "1" -- o
// enum tem um membro "nenhum" antes dele, que vale string vazia.
type TpCredPresIBSZFM int

const (
	TcpNenhum                TpCredPresIBSZFM = iota // vazio
	TcpSemCredito                                    // 0
	TcpBensConsumoFinal                              // 1
	TcpBensCapital                                   // 2
	TcpBensIntermediarios                            // 3
	TcpBensInformaticaOutros                         // 4
)

var tpCredPresIBSZFMCodigos = [...]string{"", "0", "1", "2", "3", "4"}

// String retorna o codigo do leiaute.
func (v TpCredPresIBSZFM) String() string { return codigo(tpCredPresIBSZFMCodigos[:], int(v)) }

// ParseTpCredPresIBSZFM converte o codigo do leiaute no enum. Vazio e valido.
func ParseTpCredPresIBSZFM(s string) (TpCredPresIBSZFM, error) {
	i, err := indice(tpCredPresIBSZFMCodigos[:], s, "TpCredPresIBSZFM")
	return TpCredPresIBSZFM(i), err
}

// ---------------------------------------------------------------------------

func codigo(tabela []string, i int) string {
	if i < 0 || i >= len(tabela) {
		return ""
	}
	return tabela[i]
}

func indice(tabela []string, s, enum string) (int, error) {
	for i, c := range tabela {
		if c == s {
			return i, nil
		}
	}
	return 0, erroEnum(enum, s)
}
