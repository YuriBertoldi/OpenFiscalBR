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

import "testing"

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
