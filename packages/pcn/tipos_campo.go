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

package pcn

import (
	"strconv"
	"time"
)

// ---------------------------------------------------------------------------
// Leitura de conteudo de tag
//
// Porte de ObterConteudoTag (ACBrXmlBase.pas). No Delphi ha uma unica funcao
// que recebe um TACBrTipoCampo e devolve variant; aqui cada tipo vira uma
// funcao tipada, porque em Go o variant custaria uma asercao de tipo em todo
// ponto de uso -- e um ponto de falha em runtime onde hoje ha erro de
// compilacao.
//
// A correspondencia com os tipos do Delphi e:
//
//	tcStr, tcEsp, tcStrOrig, tcNumStr  -> ConteudoStr
//	tcInt                              -> ConteudoInt
//	tcInt64                            -> ConteudoInt64
//	tcDe1..tcDe8, tcDe10               -> ConteudoDec / ConteudoDe2 ... De10
//	tcDat, tcDatHor                    -> ConteudoData / ConteudoDataHora
//	tcHor                              -> ConteudoHora
//	tcBool, tcBoolStr                  -> ConteudoBool
//
// Node nil devolve sempre o valor zero do tipo, sem erro -- e o que permite
// portar os leitores do ACBr sem um "if Assigned" por campo.
// ---------------------------------------------------------------------------

// ConteudoStr devolve o texto da tag, ja com Trim.
// Corresponde a tcStr / tcEsp / tcStrOrig / tcNumStr.
func ConteudoStr(n *Node) string {
	return n.Conteudo()
}

// ConteudoInt devolve o conteudo da tag como inteiro.
// Corresponde a tcInt.
//
// Atencao: como no original, o conteudo passa por OnlyNumber ANTES da
// conversao -- "35 ", "3-5" e "n35" todos valem 35. Nao e strconv.Atoi.
// Conteudo vazio vale 0.
func ConteudoInt(n *Node) int {
	s := OnlyNumber(n.Conteudo())
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// ConteudoInt64 devolve o conteudo da tag como int64, com a mesma regra de
// OnlyNumber de ConteudoInt. Corresponde a tcInt64.
func ConteudoInt64(n *Node) int64 {
	s := OnlyNumber(n.Conteudo())
	if s == "" {
		return 0
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return v
}

// ConteudoDec devolve o conteudo da tag como float64 com o numero de casas
// decimais informado. Corresponde a tcDe1..tcDe8 e tcDe10.
//
// O parametro casas so tem efeito quando o node esta marcado com
// FloatIsIntString -- ai o conteudo e lido como inteiro com decimais
// implicitas ("1234" com 2 casas vale 12,34). No XML de DFe da SEFAZ o valor
// ja vem com ponto decimal e o parametro e apenas documentacao da precisao
// declarada no leiaute, exatamente como no Delphi.
func ConteudoDec(n *Node, casas int) float64 {
	s := n.Conteudo()
	if n != nil && n.FloatIsIntString {
		v, err := StringDecimalToFloat(s, casas)
		if err != nil {
			return 0
		}
		return v
	}
	return StringToFloatDef(s, 0)
}

// Atalhos por precisao, para o leitor ficar com a mesma cara do .pas.
func ConteudoDe1(n *Node) float64  { return ConteudoDec(n, 1) }
func ConteudoDe2(n *Node) float64  { return ConteudoDec(n, 2) }
func ConteudoDe3(n *Node) float64  { return ConteudoDec(n, 3) }
func ConteudoDe4(n *Node) float64  { return ConteudoDec(n, 4) }
func ConteudoDe5(n *Node) float64  { return ConteudoDec(n, 5) }
func ConteudoDe6(n *Node) float64  { return ConteudoDec(n, 6) }
func ConteudoDe7(n *Node) float64  { return ConteudoDec(n, 7) }
func ConteudoDe8(n *Node) float64  { return ConteudoDec(n, 8) }
func ConteudoDe10(n *Node) float64 { return ConteudoDec(n, 10) }

// ConteudoData devolve o conteudo da tag como data.
// Corresponde a tcDat. Tag ausente ou vazia devolve o tempo zero.
func ConteudoData(n *Node) (time.Time, error) {
	t, err := EncodeDataHora(n.Conteudo())
	if err != nil {
		return time.Time{}, NovoErroLeitura(n.Caminho(), n.Conteudo(), ErrDataInvalida)
	}
	return t, nil
}

// ConteudoDataHora devolve o conteudo da tag como data e hora.
// Corresponde a tcDatHor. Tag ausente ou vazia devolve o tempo zero.
func ConteudoDataHora(n *Node) (time.Time, error) {
	return ConteudoData(n)
}

// ConteudoDataDef e como ConteudoData, mas devolve o tempo zero em vez de
// erro -- o comportamento dos leitores do ACBr, que nao conferem.
func ConteudoDataDef(n *Node) time.Time {
	t, _ := ConteudoData(n)
	return t
}

// ConteudoHora devolve o conteudo da tag como duracao desde a meia-noite.
// Corresponde a tcHor, que le hh:mm:ss.
func ConteudoHora(n *Node) time.Duration {
	s := n.Conteudo()
	if len(s) < 5 {
		return 0
	}
	h := atoi(s[0:2])
	m := atoi(s[3:5])
	seg := 0
	if len(s) >= 8 {
		seg = atoi(s[6:8])
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(seg)*time.Second
}

// ConteudoBool devolve o conteudo da tag como booleano.
// Corresponde a tcBool / tcBoolStr: so "true" (em qualquer caixa) e
// verdadeiro; qualquer outro conteudo, inclusive vazio, e falso.
func ConteudoBool(n *Node) bool {
	return NormatizarBoolean(n.Conteudo()) == "True"
}

// ConteudoCNPJCPF devolve o CNPJ do node e, se vazio, o CPF.
// Porte de ObterConteudoTagCNPJCPF (ACBrXmlBase.pas).
//
// O original usa Find (nome qualificado); aqui tambem, para manter a
// fidelidade -- num XML com prefixo declarado o ACBr nao acharia o campo.
func ConteudoCNPJCPF(n *Node) string {
	r := ConteudoStr(n.Find("CNPJ"))
	if r == "" {
		r = ConteudoStr(n.Find("CPF"))
	}
	return r
}

// ConteudoCNPJCPFAnyNs e a variante tolerante a prefixo de namespace.
// Nao tem equivalente no ACBr -- existe para quem importa XML de origem
// heterogenea, onde o emitente usa prefixo.
func ConteudoCNPJCPFAnyNs(n *Node) string {
	r := ConteudoStr(n.FindAnyNs("CNPJ"))
	if r == "" {
		r = ConteudoStr(n.FindAnyNs("CPF"))
	}
	return r
}

// AtributoInt devolve o atributo como inteiro, com a mesma regra de
// OnlyNumber de ConteudoInt.
func AtributoInt(n *Node, nome string) int {
	s := OnlyNumber(n.Attr(nome))
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}
