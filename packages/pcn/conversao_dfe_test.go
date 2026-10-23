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

package pcn

import (
	"errors"
	"testing"
)

// TestSituacaoDFe_QuatroCodigos trava o enum do cSitNFe.
//
// NAO e o cStat: e um codigo proprio de quatro valores. No schema `procNFe`
// ele nem vem pronto -- sai de um `case` sobre o cStat do protocolo
// (RetDistDFeInt.pas:620), e e o packages/nfe quem faz essa derivacao.
func TestSituacaoDFe_QuatroCodigos(t *testing.T) {
	casos := []struct {
		situacao SituacaoDFe
		codigo   string
	}{
		{SnAutorizado, "1"},
		{SnDenegado, "2"},
		{SnCancelado, "3"},
		{SnEncerrado, "4"},
	}
	for _, c := range casos {
		if got := c.situacao.String(); got != c.codigo {
			t.Errorf("%v: codigo = %q, esperado %q", c.situacao, got, c.codigo)
		}
		got, err := ParseSituacaoDFe(c.codigo)
		if err != nil {
			t.Errorf("%q: %v", c.codigo, err)
			continue
		}
		if got != c.situacao {
			t.Errorf("%q voltou como %v", c.codigo, got)
		}
	}

	// A ordem importa: denegado e o 2 e cancelado e o 3, nao o contrario.
	// O `case` do leitor da distribuicao mapeia cStat 101 para CANCELADO e
	// 110 para DENEGADO -- trocar os dois aqui inverteria a situacao de
	// toda nota cancelada.
	if SnDenegado == SnCancelado {
		t.Fatal("denegado e cancelado sao membros distintos")
	}

	// Codigo desconhecido devolve erro E o primeiro membro, como o
	// StrToSituacaoDFe do ACBr -- que tambem nao distingue "autorizado" de
	// "nao reconhecido". E uma armadilha do original: um resumo sem
	// cSitNFe le como autorizado.
	got, err := ParseSituacaoDFe("")
	if err == nil {
		t.Error("codigo vazio deveria dar erro")
	}
	if got != SnAutorizado {
		t.Errorf("o valor de erro e o primeiro membro, veio %v", got)
	}

	// A tabela de codigos e o enum tem de andar juntos. Membro acrescentado
	// sem a entrada correspondente nao quebra build nem vet: o String()
	// cairia no ramo de fora-de-faixa e o Parse nunca devolveria o membro
	// novo -- os dois em silencio.
	if int(_fimSituacaoDFe) != len(situacaoDFeCodigos) {
		t.Fatalf("o enum tem %d membros e a tabela tem %d codigos",
			_fimSituacaoDFe, len(situacaoDFeCodigos))
	}
	if _fimSituacaoDFe != 4 {
		t.Fatalf("SituacaoDFe tem quatro membros no leiaute, declarados %d",
			_fimSituacaoDFe)
	}
}

// TestModal_DoisDigitosEOsSeisMeios trava o enum compartilhado pelo CT-e e
// pelo MDF-e.
//
// O codigo tem DOIS digitos: o schema declara o campo com dois caracteres,
// e emitir "1" em vez de "01" e rejeicao. Como os seis valem de 01 a 06, a
// tentacao de tratar como inteiro e real.
func TestModal_DoisDigitosEOsSeisMeios(t *testing.T) {
	casos := []struct {
		v      Modal
		codigo string
	}{
		{MdRodoviario, "01"},
		{MdAereo, "02"},
		{MdAquaviario, "03"},
		{MdFerroviario, "04"},
		{MdDutoviario, "05"},
		{MdMultimodal, "06"},
	}
	if len(casos) != int(_fimModal) {
		t.Fatalf("o enum tem %d membros e o teste confere %d",
			int(_fimModal), len(casos))
	}
	for _, c := range casos {
		if got := c.v.String(); got != c.codigo {
			t.Errorf("%d = %q, esperado %q", int(c.v), got, c.codigo)
		}
		if v, err := ParseModal(c.codigo); err != nil || v != c.v {
			t.Errorf("Parse(%q) = %v %v", c.codigo, v, err)
		}
	}

	// Um digito so NAO e aceito -- e o engano que o leiaute pune.
	if _, err := ParseModal("1"); !errors.Is(err, ErrEnumInvalido) {
		t.Error(`"1" nao e codigo valido; o leiaute usa "01"`)
	}

	if len(modalCodigos) != int(_fimModal) || len(modalDescricoes) != int(_fimModal) {
		t.Errorf("as tabelas tem %d e %d entradas para %d membros",
			len(modalCodigos), len(modalDescricoes), int(_fimModal))
	}
}

// TestUnidades_LarguraDoCodigoVariaEntreELAS trava a diferenca que mais
// escapa entre os tres enums de unidade.
//
// O UnidMed usa DOIS digitos comecando em "00"; o UnidTransp e o UnidCarga
// usam UM digito comecando em "1". Os tres convivem no mesmo documento, e
// tratar qualquer um como inteiro produz rejeicao.
func TestUnidades_LarguraDoCodigoVariaEntreELAS(t *testing.T) {
	// Dois digitos, base zero.
	if UM3.String() != "00" || UMMBTU.String() != "05" {
		t.Errorf("UnidMed: %q..%q, esperado 00..05", UM3.String(), UMMBTU.String())
	}
	if _, err := ParseUnidMed("0"); !errors.Is(err, ErrEnumInvalido) {
		t.Error(`UnidMed nao aceita "0" -- o leiaute usa "00"`)
	}

	// Um digito, base um.
	if UtRodoTracao.String() != "1" || UtOutros.String() != "7" {
		t.Errorf("UnidTransp: %q..%q, esperado 1..7",
			UtRodoTracao.String(), UtOutros.String())
	}
	if _, err := ParseUnidTransp("0"); !errors.Is(err, ErrEnumInvalido) {
		t.Error("UnidTransp comeca em 1, nao em 0")
	}

	if UcContainer.String() != "1" || UcOutros.String() != "4" {
		t.Errorf("UnidCarga: %q..%q, esperado 1..4",
			UcContainer.String(), UcOutros.String())
	}

	// Ida e volta de cada um, inteiro.
	for i := UM3; i < _fimUnidMed; i++ {
		if v, err := ParseUnidMed(i.String()); err != nil || v != i {
			t.Errorf("UnidMed %d (%q): %v %v", i, i.String(), v, err)
		}
	}
	for i := UtRodoTracao; i < _fimUnidTransp; i++ {
		if v, err := ParseUnidTransp(i.String()); err != nil || v != i {
			t.Errorf("UnidTransp %d (%q): %v %v", i, i.String(), v, err)
		}
	}
	for i := UcContainer; i < _fimUnidCarga; i++ {
		if v, err := ParseUnidCarga(i.String()); err != nil || v != i {
			t.Errorf("UnidCarga %d (%q): %v %v", i, i.String(), v, err)
		}
	}

	// As descricoes sao ABREVIACOES, e duas tentam o porte a escrever por
	// extenso: "UND" e "LT". O DACTE tem campo de largura fixa e o que cabe
	// e a abreviacao -- conferir so o TAMANHO da tabela nao pega a troca.
	descricoes := [...]string{"M3", "KG", "TON", "UND", "LT", "MMBTU"}
	for i := UM3; i < _fimUnidMed; i++ {
		if got := i.Descricao(); got != descricoes[i] {
			t.Errorf("UnidMed(%d).Descricao() = %q, esperado %q",
				int(i), got, descricoes[i])
		}
	}

	// Tabelas alinhadas com os enums.
	casos := []struct {
		nome   string
		tabela int
		fim    int
	}{
		{"UnidMed", len(unidMedCodigos), int(_fimUnidMed)},
		{"UnidMed/descricoes", len(unidMedDescricoes), int(_fimUnidMed)},
		{"UnidTransp", len(unidTranspCodigos), int(_fimUnidTransp)},
		{"UnidCarga", len(unidCargaCodigos), int(_fimUnidCarga)},
	}
	for _, c := range casos {
		if c.tabela != c.fim {
			t.Errorf("%s: tabela %d, enum %d", c.nome, c.tabela, c.fim)
		}
	}
}
