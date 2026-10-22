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

// Testes do catalogo de schemas da Distribuicao DF-e.

// TestSchemaDFe_ListaAlinhadaComOEnum trava o alinhamento entre o enum e a
// tabela de nomes.
//
// Sao duas listas paralelas; um membro acrescentado numa so desloca todos os
// seguintes, e o efeito e o pior possivel: o schema continua existindo, com
// o nome do vizinho.
func TestSchemaDFe_ListaAlinhadaComOEnum(t *testing.T) {
	if len(schemaDFeNomes) != int(_fimSchemaDFe) {
		t.Errorf("nomes = %d, membros do enum = %d",
			len(schemaDFeNomes), _fimSchemaDFe)
	}
	// Dezesseis do ACBr mais o sentinela.
	if _fimSchemaDFe != 17 {
		t.Errorf("o catalogo tem %d membros, esperado 17 (16 do ACBr + o"+
			" sentinela)", _fimSchemaDFe)
	}
	if SchemaDesconhecido.String() != "" {
		t.Errorf("o sentinela nao tem nome de schema: %q",
			SchemaDesconhecido.String())
	}
}

// TestSchemaDFe_NomeEIdaEVolta trava que String e ParseSchemaDFe sao
// inversas para todo membro real.
func TestSchemaDFe_NomeEIdaEVolta(t *testing.T) {
	for s := SchemaDesconhecido + 1; s < _fimSchemaDFe; s++ {
		nome := s.String()
		if nome == "" {
			t.Errorf("membro %d sem nome", s)
			continue
		}
		got, err := ParseSchemaDFe(nome)
		if err != nil {
			t.Errorf("%s: %v", nome, err)
			continue
		}
		if got != s {
			t.Errorf("%s voltou como %s", nome, got)
		}
		// E com a versao colada, que e como a RFB o envia.
		got, err = ParseSchemaDFe(nome + "_v1.01")
		if err != nil || got != s {
			t.Errorf("%s_v1.01 = %s (%v)", nome, got, err)
		}
	}
}
