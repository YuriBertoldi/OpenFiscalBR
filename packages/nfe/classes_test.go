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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Os testes deste arquivo cobrem UMA coisa: os pontos em que o zero value do
// Go NAO reproduz o objeto recem-criado do ACBr.
//
// Struct de dados pura nao tem teste aqui de proposito -- nao existe regra
// para quebrar. O que existe e o default de construtor, e cada um deles vaza
// direto para o XML quando se perde.

// TestNovoPagItem_IndPagETBandNaoComecamNoZero e o teste mais importante
// deste arquivo.
//
// TpagCollectionItem.Create (Classes.pas:4441) atribui dois campos, e NENHUM
// dos dois e o ordinal zero do seu enum:
//
//	FindPag := ipNenhum                   // ordinal 3, grava ""
//	FtBand  := TpcnBandeiraCartao(-1)     // sentinela, grava ""
//
// O zero value do Go daria IpVista (grava "0") e BcVisa (grava "01"). Sem o
// construtor, toda forma de pagamento sairia do gerador declarando pagamento
// a vista com cartao Visa -- em pagamentos que nao tem nem um nem outro.
//
// E o mesmo defeito que o porte da NFGas encontrou em tiSim: o membro que
// significa "ausente" nao esta no ordinal zero.
func TestNovoPagItem_IndPagETBandNaoComecamNoZero(t *testing.T) {
	p := NovoPagItem()

	if p.IndPag != IpNenhum {
		t.Errorf("IndPag = %v, esperado IpNenhum", p.IndPag)
	}
	if got := p.IndPag.String(); got != "" {
		t.Errorf("IndPag grava %q, esperado vazio", got)
	}

	if p.TBand != BcNenhum {
		t.Errorf("TBand = %v, esperado BcNenhum", p.TBand)
	}
	if got := p.TBand.String(); got != "" {
		t.Errorf("TBand grava %q, esperado vazio", got)
	}

	// A prova de que o construtor e necessario: o zero value grava conteudo.
	var zero PagItem
	if zero.IndPag.String() != "0" || zero.TBand.String() != "01" {
		t.Fatalf("o zero value de PagItem mudou (IndPag=%q TBand=%q); "+
			"se ele passou a gravar vazio, NovoPagItem pode ter virado "+
			"redundante -- confira contra Classes.pas:4441",
			zero.IndPag.String(), zero.TBand.String())
	}
}

// TestNovoEmit_CRTComecaEmRegimeNormal cobre o unico default de
// TEmit.Create (Classes.pas:2702): FCRT := crtRegimeNormal.
//
// O zero value do Go e CrtSimplesNacional, que grava "1" -- declararia o
// emitente no Simples Nacional e mudaria a tributacao esperada da nota.
func TestNovoEmit_CRTComecaEmRegimeNormal(t *testing.T) {
	e := NovoEmit()
	if e.CRT != CrtRegimeNormal {
		t.Errorf("CRT = %v, esperado CrtRegimeNormal", e.CRT)
	}
	if got := e.CRT.String(); got != "3" {
		t.Errorf("CRT grava %q, esperado \"3\"", got)
	}

	var zero Emit
	if zero.CRT == CrtRegimeNormal {
		t.Error("o zero value de Emit passou a ser CrtRegimeNormal; " +
			"NovoEmit so existe porque os dois divergem")
	}
}

// TestNovaNFe_AplicaOsDefaultsDoConstrutorDelphi cobre os tres defaults de
// TNFe.Create (Classes.pas:2435-2439), que saem no XML:
//
//	FEmit.EnderEmit.xPais := 'BRASIL';  cPais := 1058
//	FDest.EnderDest.xPais := 'BRASIL';  cPais := 1058
//
// mais o CRT que o Create herda ao construir o emitente.
func TestNovaNFe_AplicaOsDefaultsDoConstrutorDelphi(t *testing.T) {
	n := NovaNFe()

	if n.Emit.EnderEmit.XPais != "BRASIL" || n.Emit.EnderEmit.CPais != 1058 {
		t.Errorf("enderEmit = %q/%d, esperado BRASIL/1058",
			n.Emit.EnderEmit.XPais, n.Emit.EnderEmit.CPais)
	}
	if n.Dest.EnderDest.XPais != "BRASIL" || n.Dest.EnderDest.CPais != 1058 {
		t.Errorf("enderDest = %q/%d, esperado BRASIL/1058",
			n.Dest.EnderDest.XPais, n.Dest.EnderDest.CPais)
	}
	if n.Emit.CRT != CrtRegimeNormal {
		t.Errorf("CRT = %v, esperado CrtRegimeNormal (herdado de TEmit.Create)", n.Emit.CRT)
	}
}

// TestIde_ZeroValueReproduzOACBr e o inverso dos testes acima: trava a
// AUSENCIA de construtor em Ide.
//
// As properties de TIde declaram `default ipPrazo`, `default tnSaida`,
// `default tiPaisagem` e `default taHomologacao`, mas o `default` do Delphi e
// metadado de streaming e NAO inicializa nada -- TIde.Create (Classes.pas:2597)
// so aloca os filhos. Um TIde recem-criado no ACBr esta, de fato, em PRODUCAO.
//
// Sem este teste, alguem que lesse as properties "corrigiria" o porte criando
// um NovoIde com esses valores, e passaria a emitir em homologacao notas que o
// ACBr emite em producao (e vice-versa).
func TestIde_ZeroValueReproduzOACBr(t *testing.T) {
	var ide Ide

	casos := []struct {
		campo             string
		obtido            string
		propertyDeclarava string
	}{
		{"tpAmb", ide.TpAmb.String(), "taHomologacao (\"2\")"},
		{"tpNF", ide.TpNF.String(), "tnSaida (\"1\")"},
		{"tpImp", ide.TpImp.String(), "tiPaisagem (\"2\")"},
		{"indPag", ide.IndPag.String(), "ipPrazo (\"1\")"},
	}
	esperado := map[string]string{
		"tpAmb":  "1", // taProducao
		"tpNF":   "0", // tnEntrada
		"tpImp":  "0", // tiSemGeracao
		"indPag": "0", // ipVista
	}

	for _, c := range casos {
		if c.obtido != esperado[c.campo] {
			t.Errorf("%s do zero value grava %q, esperado %q; "+
				"a property do ACBr declara %s, mas o Create nao atribui nada",
				c.campo, c.obtido, esperado[c.campo], c.propertyDeclarava)
		}
	}

	// os dois que, por acaso, coincidem com a declaracao da property
	if ide.TpEmis != pcn.TeNormal {
		t.Errorf("tpEmis do zero value = %v, esperado TeNormal", ide.TpEmis)
	}
	if ide.FinNFe != FnNormal {
		t.Errorf("finNFe do zero value = %v, esperado FnNormal", ide.FinNFe)
	}
	if ide.ProcEmi != pcn.PeAplicativoContribuinte {
		t.Errorf("procEmi do zero value = %v, esperado PeAplicativoContribuinte", ide.ProcEmi)
	}
}

// TestGuiaTransito_ZeroValueNaoDeclaraGuia trava a outra ausencia deliberada
// de construtor.
//
// TguiaTransito.Create (Classes.pas:4669) atribui tpgNenhum, e isso SO
// dispensa construtor em Go porque tpgNenhum e o ordinal zero de TtpGuia. A
// afirmacao depende da ordem de um enum declarado em outro arquivo: se alguem
// reordenar TpGuia, GuiaTransito passa a nascer declarando uma guia concreta
// numa nota que nao tem guia, sem nada falhar.
func TestGuiaTransito_ZeroValueNaoDeclaraGuia(t *testing.T) {
	var g GuiaTransito
	if g.TpGuia != TpgNenhum {
		t.Errorf("TpGuia do zero value = %v, esperado TpgNenhum", g.TpGuia)
	}
	if got := g.TpGuia.String(); got != "" {
		t.Errorf("TpGuia do zero value grava %q, esperado vazio", got)
	}
}
