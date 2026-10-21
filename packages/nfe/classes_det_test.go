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
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Como em classes_test.go, os testes aqui cobrem so os pontos em que o zero
// value do Go poderia divergir do objeto recem-criado do ACBr -- nos dois
// sentidos: construtor que precisa existir, e construtor que NAO deve existir.

// TestNovoProd_IndEscalaNaoComecaNoZero cobre o unico default de
// TProd.Create (Classes.pas:2747): FindEscala := ieNenhum.
//
// ieNenhum e o ordinal DOIS de (ieRelevante, ieNaoRelevante, ieNenhum). Sem o
// construtor, todo item nasceria IeRelevante e gravaria indEscala="S" -- campo
// que o leiaute so admite junto do CEST, e que declara producao em escala
// relevante numa nota que nao disse isso.
func TestNovoProd_IndEscalaNaoComecaNoZero(t *testing.T) {
	p := NovoProd()
	if p.IndEscala != IeNenhum {
		t.Errorf("IndEscala = %v, esperado IeNenhum", p.IndEscala)
	}
	if got := p.IndEscala.String(); got != "" {
		t.Errorf("IndEscala grava %q, esperado vazio", got)
	}

	// A prova de que o construtor e necessario.
	var zero Prod
	if zero.IndEscala.String() != "S" {
		t.Fatalf("o zero value de Prod.IndEscala grava %q, nao \"S\"; "+
			"se IndEscala mudou de ordem, NovoProd precisa ser reconferido "+
			"contra Classes.pas:2747", zero.IndEscala.String())
	}

	// Os outros dois defaults de TProd coincidem com o zero, e e isso que
	// dispensa atribui-los no construtor.
	if p.IndTot != ItSomaTotalNFe {
		t.Errorf("IndTot = %v, esperado ItSomaTotalNFe (o item compoe o total)", p.IndTot)
	}
	if p.IndBemMovelUsado != pcn.TieNenhum {
		t.Errorf("IndBemMovelUsado = %v, esperado TieNenhum", p.IndBemMovelUsado)
	}
}

// TestNovoDetItem_HerdaODefaultDoProd confirma a cadeia de construtores:
// TDetCollectionItem.Create constroi o TProd, entao o item tambem nasce com
// indEscala vazio.
//
// Quem montar DetItem{} por literal nao recebe isso -- e o motivo de o
// doc-comment do campo NFe.Det apontar para NovoDetItem.
func TestNovoDetItem_HerdaODefaultDoProd(t *testing.T) {
	d := NovoDetItem()
	if d.Prod.IndEscala != IeNenhum {
		t.Errorf("Prod.IndEscala = %v, esperado IeNenhum", d.Prod.IndEscala)
	}
}

// TestICMS_ZeroValueReproduzOACBr e o inverso: trava a AUSENCIA de construtor
// em ICMS.
//
// TICMS.Create (Classes.pas:4170) so atribui FindDeduzDeson := tieNenhum, que
// ja e o ordinal zero. Mas as properties declaram quatro defaults --
// `default oeNacional`, `default cst00`, `default dbiMargemValorAgregado` e
// `default dbisPrecoTabelado` -- e `default` nao inicializa nada.
//
// O perigoso e o CST: quem lesse `default cst00` criaria um construtor
// atribuindo CST00, e todo item sem ICMS passaria a declarar CST "00"
// (tributado integralmente) em vez de nao declarar CST nenhum.
func TestICMS_ZeroValueReproduzOACBr(t *testing.T) {
	var icms ICMS

	if icms.CST != pcn.CSTVazio {
		t.Errorf("CST do zero value = %v, esperado CSTVazio; a property declara "+
			"`default cst00`, mas TICMS.Create nao atribui nada e o Delphi "+
			"tambem tem cstVazio como primeiro membro", icms.CST)
	}
	if got := icms.CST.String(); got != "" {
		t.Errorf("CST do zero value grava %q, esperado vazio", got)
	}
	if icms.CSOSN != pcn.CsosnVazio {
		t.Errorf("CSOSN do zero value = %v, esperado CsosnVazio", icms.CSOSN)
	}

	// Estes tres coincidem com o declarado na property, por acaso.
	if icms.Orig != pcn.OeNacional {
		t.Errorf("orig do zero value = %v, esperado OeNacional", icms.Orig)
	}
	if icms.ModBC != DbiMargemValorAgregado {
		t.Errorf("modBC do zero value = %v, esperado DbiMargemValorAgregado", icms.ModBC)
	}
	if icms.ModBCST != DbisPrecoTabelado {
		t.Errorf("modBCST do zero value = %v, esperado DbisPrecoTabelado", icms.ModBCST)
	}

	// E este e o unico que o Create de fato atribui -- coincidindo com o zero.
	if icms.IndDeduzDeson != pcn.TieNenhum {
		t.Errorf("indDeduzDeson do zero value = %v, esperado TieNenhum", icms.IndDeduzDeson)
	}
}

// TestICMS_MotDesICMSNasceTaxi documenta e trava uma heranca incomoda do ACBr.
//
// FmotDesICMS nao e inicializado por TICMS.Create e nao tem `default` na
// property, entao nasce no ordinal zero -- que em MotivoDesoneracaoICMS e
// mdiTaxi, codigo "1". Um ICMS recem-criado no ACBr ja declara "desonerado
// por ser taxi".
//
// Nao e defeito do porte: Delphi e Go concordam, e cabe ao gerador so emitir
// motDesICMS quando o CST admite. O teste existe para que ninguem "conserte"
// isso para MdiNenhum -- o que divergiria do ACBr em silencio.
func TestICMS_MotDesICMSNasceTaxi(t *testing.T) {
	var icms ICMS
	if icms.MotDesICMS != MdiTaxi {
		t.Errorf("motDesICMS do zero value = %v, esperado MdiTaxi (fidelidade ao ACBr)",
			icms.MotDesICMS)
	}
	if got := icms.MotDesICMS.String(); got != "1" {
		t.Errorf("motDesICMS do zero value grava %q, esperado \"1\"", got)
	}
	if icms.MotDesICMSST != MdiTaxi {
		t.Errorf("motDesICMSST do zero value = %v, esperado MdiTaxi", icms.MotDesICMSST)
	}
}

// TestImposto_DefaultsDeCSTCoincidemComOZero trava os quatro grupos de
// imposto cujo `default` de property coincide com o ordinal zero -- e que por
// isso NAO ganharam construtor.
//
// Nenhum dos quatro grava vazio: IPI nasce "00" (entrada com recuperacao de
// credito), PIS e COFINS nascem "01" (operacao tributavel). So o ISSQN nasce
// vazio. Se um desses enums for reordenado, o default de todo item muda sem
// que nada mais falhe.
//
// O mais exposto e o CstIpi, cuja ordem de declaracao ja nao segue a sequencia
// numerica (00, 49, 50, 99, 01...) e por isso parece candidata a "arrumacao".
func TestImposto_DefaultsDeCSTCoincidemComOZero(t *testing.T) {
	var imp Imposto

	casos := []struct {
		campo    string
		obtido   string
		esperado string
		property string
	}{
		{"IPI.CST", imp.IPI.CST.String(), "00", "default ipi00 (Classes.pas:1212)"},
		{"PIS.CST", imp.PIS.CST.String(), "01", "default pis01 (Classes.pas:1248)"},
		{"COFINS.CST", imp.COFINS.CST.String(), "01", "default cof01 (Classes.pas:1289)"},
		{"ISSQN.CSitTrib", imp.ISSQN.CSitTrib.String(), "", "default ISSQNcSitTribVazio (Classes.pas:1347)"},
	}
	for _, c := range casos {
		if c.obtido != c.esperado {
			t.Errorf("%s do zero value grava %q, esperado %q; a property declara %s",
				c.campo, c.obtido, c.esperado, c.property)
		}
	}
}

// TestArma_TpArmaZeroEhUsoPermitido trava a terceira afirmacao de "o default
// da property coincide com o ordinal zero, entao nao precisa de construtor".
//
// Mesmo criterio de TestGuiaTransito_ZeroValueNaoDeclaraGuia: a afirmacao
// depende da ordem de um enum de outro arquivo.
func TestArma_TpArmaZeroEhUsoPermitido(t *testing.T) {
	var a Arma
	if a.TpArma != TaUsoPermitido {
		t.Errorf("tpArma do zero value = %v, esperado TaUsoPermitido", a.TpArma)
	}
	if got := a.TpArma.String(); got != "0" {
		t.Errorf("tpArma do zero value grava %q, esperado \"0\"", got)
	}
}
