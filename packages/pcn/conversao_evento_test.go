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

// Testes do catalogo de eventos.
//
// Noventa e seis membros nao cabem num teste caso a caso que valha alguma
// coisa. O que estes travam sao as PROPRIEDADES do catalogo -- alinhamento
// das tres listas, codigos de seis digitos, e a ambiguidade que justifica
// nao haver um parse global -- mais os codigos que um porte erra.

// TestTipoEvento_TresListasAlinhadas trava o alinhamento.
//
// O enum, os codigos e as descricoes sao tres arrays paralelas no Delphi, e
// um membro acrescentado numa so desloca TODOS os seguintes. O efeito e o
// pior possivel: o evento continua existindo, com o codigo do vizinho.
func TestTipoEvento_TresListasAlinhadas(t *testing.T) {
	if len(tipoEventoCodigos) != int(_fimTipoEvento) {
		t.Errorf("codigos = %d, membros do enum = %d",
			len(tipoEventoCodigos), _fimTipoEvento)
	}
	if len(tipoEventoDescricoes) != int(_fimTipoEvento) {
		t.Errorf("descricoes = %d, membros do enum = %d",
			len(tipoEventoDescricoes), _fimTipoEvento)
	}
	if _fimTipoEvento != 96 {
		t.Errorf("o catalogo do ACBr tem 96 membros, este tem %d", _fimTipoEvento)
	}
}

// TestTipoEvento_TodoCodigoTemSeisDigitos trava o formato.
//
// O tpEvento do leiaute e `TEvento` com exatamente seis posicoes. O unico
// que escapa e o sentinela TeNaoMapeado, que nao e valor de leiaute --
// e justamente por isso ele entra no teste, para que ninguem o "conserte".
func TestTipoEvento_TodoCodigoTemSeisDigitos(t *testing.T) {
	if TeNaoMapeado.Codigo() != "-99999" {
		t.Errorf("o sentinela deveria ser -99999, e %q", TeNaoMapeado.Codigo())
	}

	for i := TeNaoMapeado + 1; i < _fimTipoEvento; i++ {
		c := i.Codigo()
		if len(c) != 6 {
			t.Errorf("%s: codigo %q tem %d digitos, esperado 6", i, c, len(c))
			continue
		}
		for _, d := range c {
			if d < '0' || d > '9' {
				t.Errorf("%s: codigo %q nao e numerico", i, c)
				break
			}
		}
	}
}

// TestTipoEvento_CodigoEDescricaoNaoSeConfundem trava a distincao entre os
// dois metodos.
//
// Os dois devolvem string, e trocar um pelo outro compila. Um gerador que
// emitisse `<tpEvento>Cancelamento</tpEvento>` passaria em qualquer teste
// de estrutura.
func TestTipoEvento_CodigoEDescricaoNaoSeConfundem(t *testing.T) {
	if TeCancelamento.Codigo() != "110111" {
		t.Errorf("Codigo() = %q, esperado 110111", TeCancelamento.Codigo())
	}
	if TeCancelamento.String() != "Cancelamento" {
		t.Errorf("String() = %q, esperado Cancelamento", TeCancelamento.String())
	}
	if TeCCe.Codigo() != "110110" || TeCCe.String() != "CCe" {
		t.Errorf("CCe: %q / %q", TeCCe.Codigo(), TeCCe.String())
	}
}

// TestTipoEvento_CodigosDaNFeQueUmPorteErra trava os codigos cuja ordem no
// enum NAO acompanha a ordem numerica -- que e onde um deslocamento de uma
// casa passa despercebido.
func TestTipoEvento_CodigosDaNFeQueUmPorteErra(t *testing.T) {
	casos := []struct {
		evento TipoEvento
		codigo string
	}{
		{TeCCe, "110110"},
		{TeCancelamento, "110111"},
		{TeCancSubst, "110112"},
		{TeEPECNFe, "110140"},
		{TeManifDestConfirmacao, "210200"},
		{TeManifDestCiencia, "210210"},
		{TeManifDestDesconhecimento, "210220"},
		{TeManifDestOperNaoRealizada, "210240"},
		{TePedProrrog1, "111500"},
		{TePedProrrog2, "111501"},
		{TeCanPedProrrog1, "111502"},
		{TeCanPedProrrog2, "111503"},
		{TeComprEntregaNFe, "110130"},
		{TeCancComprEntregaNFe, "110131"},
		{TeAtorInteressadoNFe, "110150"},
		{TeInsucessoEntregaNFe, "110192"},
		{TeCancInsucessoEntregaNFe, "110193"},
		{TeConcFinanceira, "110750"},
		{TeCancConcFinanceira, "110751"},
		{TeCancGenerico, "110001"},
		{TeAtualizacaoDataPrevisaoEntrega, "112150"},
	}
	for _, c := range casos {
		if got := c.evento.Codigo(); got != c.codigo {
			t.Errorf("%s: codigo = %q, esperado %q", c.evento, got, c.codigo)
		}
	}
}

// TestTipoEvento_OCatalogoTemCodigoAMBIGUO trava a razao de nao existir um
// parse global.
//
// Sete codigos se repetem entre documentos. Se alguem "limpar" as
// duplicatas, este teste cai -- e a limpeza teria quebrado o CTe ou o MDFe,
// que e onde o outro significado vive.
func TestTipoEvento_OCatalogoTemCodigoAMBIGUO(t *testing.T) {
	pares := []struct {
		a, b   TipoEvento
		codigo string
	}{
		{TeEncerramento, TeCancSubst, "110112"},
		{TeNaoEmbarque, TeInclusaoDFe, "110115"},
		{TeAlteracaoPoltrona, TePagamentoOperacao, "110116"},
		{TeExcessoBagagem, TeConfirmaServMDFe, "110117"},
		{TeCTeSubstituicao, TeAutorizadoSubstituicao, "240140"},
		{TeCTeAnulacao, TeAutorizadoAjuste, "240150"},
		{TeLiberacaoPrazoCanc, TeLiberacaoPrazoCancelado, "240170"},
	}
	for _, p := range pares {
		if p.a.Codigo() != p.codigo || p.b.Codigo() != p.codigo {
			t.Errorf("%s e %s deveriam ambos valer %q; valem %q e %q",
				p.a, p.b, p.codigo, p.a.Codigo(), p.b.Codigo())
		}
		if p.a == p.b {
			t.Errorf("%s e %s sao o mesmo membro", p.a, p.b)
		}
	}
}

// TestParseTipoEventoNFe_DesempataOAmbiguoPelaTabelaDaNFe trava a resolucao
// do 110112.
//
// No catalogo ele e Encerramento (MDFe) e CancSubst (NF-e). A tabela da
// NF-e diz qual dos dois vale aqui -- e e por isso que o parse e por
// documento.
func TestParseTipoEventoNFe_DesempataOAmbiguoPelaTabelaDaNFe(t *testing.T) {
	got, err := ParseTipoEventoNFe("110112")
	if err != nil {
		t.Fatal(err)
	}
	if got != TeCancSubst {
		t.Errorf("110112 na NF-e = %s, esperado CancelamentoPorSubstituicao", got)
	}
	if got == TeEncerramento {
		t.Error("110112 nao e Encerramento na NF-e; esse e o do MDFe")
	}
}

// TestParseTipoEventoNFe_LeMaisDoQueGera trava a largura da tabela.
//
// A NF-e so GERA trinta e um eventos, mas a tabela de leitura tem cinquenta
// -- ela precisa nomear registro de passagem, averbacao de exportacao e
// eventos de MDFe e CTe que chegam amarrados a chave da nota. Estreitar a
// tabela para "o que a NF-e gera" e o atalho que deixa a consulta devolver
// evento sem nome.
func TestParseTipoEventoNFe_LeMaisDoQueGera(t *testing.T) {
	soLeitura := []struct {
		codigo string
		evento TipoEvento
	}{
		{"610600", TeRegistroCTe},
		{"790700", TeAverbacaoExportacao},
		{"990900", TeVistoriaSuframa},
		{"990910", TeConfInternalizacao},
		{"610500", TeRegistroPassagemNFe},
		{"610610", TeMDFeAutorizado},
		{"510630", TeRegistroPassagemMDFe},
	}
	for _, c := range soLeitura {
		got, err := ParseTipoEventoNFe(c.codigo)
		if err != nil {
			t.Errorf("%s deveria ser legivel pela NF-e: %v", c.codigo, err)
			continue
		}
		if got != c.evento {
			t.Errorf("%s = %s, esperado %s", c.codigo, got, c.evento)
		}
	}

	if len(tipoEventoNFePorCodigo) != 50 {
		t.Errorf("a tabela da NF-e tem %d codigos, esperado 50 (51 do ACBr"+
			" menos o sentinela)", len(tipoEventoNFePorCodigo))
	}
}

// TestParseTipoEventoNFe_CodigoDesconhecidoEhErro trava a divergencia
// deliberada: o ACBr devolve teNaoMapeado com ok=False e quase ninguem
// confere o ok; aqui o caso vira erro.
// O "-99999" entra na lista de proposito: e o unico codigo que o ACBr
// considera VALIDO (StrToTpEventoNFe o inclui e devolve ok=True) e que aqui
// vira erro. Aceita-lo criaria um caminho em que evento sem tipo passa como
// se tivesse um.
func TestParseTipoEventoNFe_CodigoDesconhecidoEhErro(t *testing.T) {
	for _, s := range []string{"110160", "999999", "", "abc", "-99999"} {
		got, err := ParseTipoEventoNFe(s)
		if err == nil {
			t.Errorf("%q deveria dar erro, devolveu %s", s, got)
		}
		if got != TeNaoMapeado {
			t.Errorf("%q: o valor de erro e TeNaoMapeado, veio %s", s, got)
		}
	}
}
