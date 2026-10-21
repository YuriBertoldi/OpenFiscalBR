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
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Testes dos eventos da Reforma Tributaria.
//
// O risco aqui nao e campo faltando: e NOME REPETIDO. Tres elementos saem
// com o mesmo nome em contextos diferentes, e o `gControleEstoque` tem sete
// formas. Um gerador generico emitiria o par de tags errado em seis dos
// sete casos, e nenhum round-trip pegaria -- o elemento PAI tem o mesmo
// nome, e o leitor acha a tag por nome.

// TestGControleEstoque_SeteFormasComTagsDIFERENTES e o teste central deste
// arquivo.
//
// Cada evento emite <gControleEstoque> com um par de tags PROPRIO. A tabela
// abaixo e a especificacao executavel das sete sobrecargas do Delphi, e a
// segunda metade de cada caso -- as tags que NAO podem aparecer -- e o que
// pega um gerador copiado do vizinho.
func TestGControleEstoque_SeteFormasComTagsDIFERENTES(t *testing.T) {
	casos := []struct {
		tipo      pcn.TipoEvento
		preparar  func(*DetEvento)
		presentes []string
		ausentes  []string
	}{
		{
			pcn.TeDestItemConsPessoal,
			func(d *DetEvento) {
				d.GConsumo = []GConsumo{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoqueConsumo{QConsumo: 3, UConsumo: "UN"},
					DFeReferenciado: rtc.DFeReferenciado{ChaveAcesso: chaveDeMinas, NItem: 1},
				}}
			},
			[]string{"<gConsumo ", "<qConsumo>", "<uConsumo>", "<DFeReferenciado>"},
			[]string{"<qtde>", "<qPerecimento>", "<qImobilizado>", "<qComb>",
				"<qNaoFornecida>"},
		},
		{
			pcn.TeImporALCZFM,
			func(d *DetEvento) {
				d.GConsumoZFM = []GConsumoZFM{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoqueZFM{Qtde: 3, Unidade: "UN"},
				}}
			},
			// MESMO nome de elemento do caso anterior, tags diferentes.
			[]string{"<gConsumo ", "<qtde>", "<unidade>"},
			[]string{"<qConsumo>", "<uConsumo>", "<DFeReferenciado>"},
		},
		{
			pcn.TePerecPerdaRouboFurtoTranspContratAqu,
			func(d *DetEvento) {
				d.GPerecimento = []GPerecimento{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoquePerecimento{
						QPerecimento: 3, UPerecimento: "UN"},
				}}
			},
			[]string{"<gPerecimento ", "<qPerecimento>", "<uPerecimento>"},
			[]string{"<qConsumo>", "<qtde>"},
		},
		{
			pcn.TePerecPerdaRouboFurtoTranspContratFornec,
			func(d *DetEvento) {
				d.GPerecimentoForn = []GPerecimentoForn{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoquePerecimentoForn{
						QPerecimento: 3, UPerecimento: "UN", VIBS: 1, VCBS: 2},
				}}
			},
			// MESMO nome de elemento do anterior; o gControleEstoque tem
			// DOIS campos a mais.
			[]string{"<gPerecimento ", "<qPerecimento>", "<uPerecimento>"},
			[]string{"<qConsumo>", "<qtde>"},
		},
		{
			pcn.TeImobilizacaoItem,
			func(d *DetEvento) {
				d.GImobilizacao = []GImobilizacao{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoqueImobilizacao{
						QImobilizado: 3, UImobilizado: "UN"},
				}}
			},
			[]string{"<gImobilizacao ", "<qImobilizado>", "<uImobilizado>"},
			[]string{"<qConsumo>", "<qPerecimento>"},
		},
		{
			pcn.TeSolicApropCredCombustivel,
			func(d *DetEvento) {
				d.GConsumoComb = []GConsumoComb{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoqueComb{QComb: 3, UComb: "LT"},
				}}
			},
			[]string{"<gConsumoComb ", "<qComb>", "<uComb>"},
			[]string{"<qConsumo>", "<qtde>"},
		},
		{
			pcn.TeFornecNaoRealizPagAntec,
			func(d *DetEvento) {
				d.GItemNaoFornecido = []GItemNaoFornecido{{
					NItem: 1, VIBS: 10, VCBS: 20,
					ControleEstoque: ControleEstoqueItemNaoFornecido{
						QNaoFornecida: 3, UNaoFornecida: "UN"},
				}}
			},
			[]string{"<gItemNaoFornecido ", "<qNaoFornecida>", "<uNaoFornecida>"},
			[]string{"<qConsumo>", "<qtde>", "<qPerecimento>"},
		},
	}

	if len(casos) != 7 {
		t.Fatalf("o Delphi tem 7 sobrecargas de Gerar_gControleEstoque,"+
			" o teste cobre %d", len(casos))
	}

	for _, c := range casos {
		t.Run(c.tipo.String(), func(t *testing.T) {
			e := eventoBase(c.tipo)
			d := &e.Evento[0].InfEvento.DetEvento
			d.COrgaoAutor, d.VerAplic = 31, "1.0"
			c.preparar(d)

			xml := gerarEventoDeTeste(t, e)

			if !strings.Contains(xml, "<gControleEstoque>") {
				t.Fatalf("faltou o gControleEstoque: %s", xml)
			}
			for _, tag := range c.presentes {
				if !strings.Contains(xml, tag) {
					t.Errorf("faltou %s: %s", tag, xml)
				}
			}
			for _, tag := range c.ausentes {
				if strings.Contains(xml, tag) {
					t.Errorf("%s NAO pertence a este evento: %s", tag, xml)
				}
			}
		})
	}
}

// TestGPerecimentoForn_ControleEstoqueTemVIBSeVCBSPROPRIOS trava a unica
// variante com quatro campos (:2390).
//
// Os nomes vIBS e vCBS se REPETEM: existem no grupo pai e dentro do
// gControleEstoque, com valores diferentes. Copiar o gerador do adquirente
// perde os dois de dentro, e nenhum teste de presenca de tag percebe --
// as tags continuam no documento, so que no nivel de cima.
func TestGPerecimentoForn_ControleEstoqueTemVIBSeVCBSPROPRIOS(t *testing.T) {
	e := eventoBase(pcn.TePerecPerdaRouboFurtoTranspContratFornec)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.GPerecimentoForn = []GPerecimentoForn{{
		NItem: 1,
		VIBS:  100.00, // do grupo PAI
		VCBS:  200.00,
		ControleEstoque: ControleEstoquePerecimentoForn{
			QPerecimento: 3, UPerecimento: "UN",
			VIBS: 11.00, // de DENTRO do gControleEstoque
			VCBS: 22.00,
		},
	}}

	xml := gerarEventoDeTeste(t, e)

	ce := recortarEntre(t, xml, "<gControleEstoque>", "</gControleEstoque>")
	if !strings.Contains(ce, "<vIBS>11.00</vIBS>") {
		t.Errorf("o vIBS de DENTRO do gControleEstoque sumiu: %s", ce)
	}
	if !strings.Contains(ce, "<vCBS>22.00</vCBS>") {
		t.Errorf("o vCBS de DENTRO do gControleEstoque sumiu: %s", ce)
	}

	// E os do grupo pai continuam la, com os outros valores.
	if !strings.Contains(xml, "<vIBS>100.00</vIBS>") {
		t.Errorf("o vIBS do grupo pai sumiu: %s", xml)
	}

	// O contraste: o perecimento do ADQUIRENTE nao tem os de dentro.
	e = eventoBase(pcn.TePerecPerdaRouboFurtoTranspContratAqu)
	d = &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.GPerecimento = []GPerecimento{{
		NItem: 1, VIBS: 100, VCBS: 200,
		ControleEstoque: ControleEstoquePerecimento{QPerecimento: 3, UPerecimento: "UN"},
	}}
	ce = recortarEntre(t, gerarEventoDeTeste(t, e), "<gControleEstoque>", "</gControleEstoque>")
	if strings.Contains(ce, "<vIBS>") {
		t.Errorf("o perecimento do adquirente NAO tem vIBS dentro do"+
			" gControleEstoque: %s", ce)
	}
}

// recortarEntre devolve o trecho entre dois delimitadores.
func recortarEntre(t *testing.T, s, inicio, fim string) string {
	t.Helper()
	i := strings.Index(s, inicio)
	if i < 0 {
		t.Fatalf("delimitador %q ausente", inicio)
	}
	j := strings.Index(s[i:], fim)
	if j < 0 {
		t.Fatalf("delimitador %q ausente", fim)
	}
	return s[i : i+j+len(fim)]
}

// TestGCredPres_ElementoSeChamaGCredPresOper trava a divergencia entre o
// nome da classe Delphi e o da tag (:2543).
func TestGCredPres_ElementoSeChamaGCredPresOper(t *testing.T) {
	e := eventoBase(pcn.TeSolicApropCredPres)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.GCredPres = []GCredPres{{
		NItem: 1, VBCCredPres: 1000, CCredPres: rtc.Cp02,
		GIBSCredPres: IBSCBSCredPres{PCredPres: 1.5, VCredPres: 15},
		GCBSCredPres: IBSCBSCredPres{PCredPres: 2.5, VCredPres: 25},
	}}

	xml := gerarEventoDeTeste(t, e)

	if !strings.Contains(xml, "<gCredPresOper ") {
		t.Errorf("o elemento e <gCredPresOper>: %s", xml)
	}
	if strings.Contains(xml, "<gCredPres ") || strings.Contains(xml, "<gCredPres>") {
		t.Errorf("<gCredPres> e o nome da CLASSE, nao o da tag: %s", xml)
	}

	// Os dois subgrupos tem as MESMAS tags e nomes de elemento diferentes.
	conferirOrdem(t, xml, "<vBCCredPres>", "<cCredPres>", "<gIBSCredPres>",
		"<gCBSCredPres>")

	ibs := recortarEntre(t, xml, "<gIBSCredPres>", "</gIBSCredPres>")
	if !strings.Contains(ibs, "<vCredPres>15.00</vCredPres>") {
		t.Errorf("gIBSCredPres: %s", ibs)
	}
	cbs := recortarEntre(t, xml, "<gCBSCredPres>", "</gCBSCredPres>")
	if !strings.Contains(cbs, "<vCredPres>25.00</vCredPres>") {
		t.Errorf("gCBSCredPres: %s", cbs)
	}
}

// TestGCredito_ValoresTemNOMEPROPRIO trava o unico grupo da Reforma sem
// gControleEstoque e com nomes de valor diferentes (:2670).
//
// Copiar o molde dos vizinhos emitiria vIBS/vCBS no lugar de
// vCredIBS/vCredCBS -- e a tag estaria presente, so que com o nome errado.
func TestGCredito_ValoresTemNOMEPROPRIO(t *testing.T) {
	e := eventoBase(pcn.TeSolicApropCredBensServicos)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.GCredito = []GCredito{{NItem: 1, VCredIBS: 10, VCredCBS: 20}}

	xml := gerarEventoDeTeste(t, e)

	if !strings.Contains(xml, "<vCredIBS>10.00</vCredIBS>") {
		t.Errorf("faltou vCredIBS: %s", xml)
	}
	if !strings.Contains(xml, "<vCredCBS>20.00</vCredCBS>") {
		t.Errorf("faltou vCredCBS: %s", xml)
	}
	if strings.Contains(xml, "<vIBS>") || strings.Contains(xml, "<vCBS>") {
		t.Errorf("este grupo NAO usa vIBS/vCBS: %s", xml)
	}
	if strings.Contains(xml, "<gControleEstoque>") {
		t.Errorf("o gCredito nao tem gControleEstoque: %s", xml)
	}
}

// TestGerarDetEvento_CancGenericoNaoTemTpAutor trava a excecao de :2464.
//
// E o unico evento da Reforma sem tpAutor. Passa-lo pelo cabecalho comum --
// que e o que um porte uniformizado faria -- acrescenta um campo que o
// schema nao tem.
func TestGerarDetEvento_CancGenericoNaoTemTpAutor(t *testing.T) {
	e := eventoBase(pcn.TeCancGenerico)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.TpEventoAut = "110111"
	d.NProtEvento = "131260000000002"

	xml := gerarEventoDeTeste(t, e)

	if strings.Contains(xml, "<tpAutor>") {
		t.Errorf("o cancelamento generico nao tem tpAutor: %s", xml)
	}
	conferirOrdem(t, xml, "<descEvento>", "<cOrgaoAutor>", "<verAplic>",
		"<tpEventoAut>", "<nProtEvento>")

	// O contraste: os demais da Reforma TEM tpAutor.
	e = eventoBase(pcn.TeImobilizacaoItem)
	d = &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.GImobilizacao = []GImobilizacao{{NItem: 1}}
	if xml := gerarEventoDeTeste(t, e); !strings.Contains(xml, "<tpAutor>") {
		t.Errorf("a imobilizacao tem tpAutor: %s", xml)
	}
}

// TestGerarDetEvento_IndQuitacaoEhFixo trava o literal "1" de :2503.
//
// Nao existe campo para ele no documento; o valor e do schema.
func TestGerarDetEvento_IndQuitacaoEhFixo(t *testing.T) {
	e := eventoBase(pcn.TePagIntegLibCredPresAdq)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"

	xml := gerarEventoDeTeste(t, e)
	if !strings.Contains(xml, "<indQuitacao>1</indQuitacao>") {
		t.Errorf("indQuitacao e fixo em 1: %s", xml)
	}
}

// TestEventoLevaAcentos_ListaDoACBr trava a lista de :947, que NAO e
// derivavel de nenhum criterio semantico.
//
// O ACBr desliga o RetirarAcentos so para esses eventos, com o comentario
// "Nao funciona sem acentos". Dois detalhes desmentem o atalho "e da
// Reforma":
//
//   - o cancelamento generico E da Reforma e NAO esta na lista;
//   - o ator interessado NAO e da Reforma e esta.
//
// E a lista nem sequer coincide com "a descricao tem acento": a do ator
// interessado -- "Ator interessado na NF-e" -- e ASCII puro. A implicacao
// vale numa direcao so, e o teste afirma exatamente essa.
func TestEventoLevaAcentos_ListaDoACBr(t *testing.T) {
	naLista := []pcn.TipoEvento{
		pcn.TeAtorInteressadoNFe, pcn.TeCancConcFinanceira,
		pcn.TeSolicApropCredPres, pcn.TeDestItemConsPessoal,
		pcn.TeImobilizacaoItem,
	}
	for _, tipo := range naLista {
		if !EventoLevaAcentos(tipo) {
			t.Errorf("%s esta na lista do ACBr", tipo)
		}
	}

	foraDaLista := []pcn.TipoEvento{
		pcn.TeCCe, pcn.TeCancelamento, pcn.TeCancGenerico,
		pcn.TeManifDestCiencia, pcn.TeConcFinanceira,
	}
	for _, tipo := range foraDaLista {
		if EventoLevaAcentos(tipo) {
			t.Errorf("%s NAO esta na lista do ACBr", tipo)
		}
	}

	// O ator interessado e o caso que desmente "lista = tem acento".
	if temAcento(DescricaoDoTipoEvento(pcn.TeAtorInteressadoNFe)) {
		t.Error("a descricao do ator interessado e ASCII; se isso mudar," +
			" o comentario da lista precisa mudar junto")
	}
}

// TestDescricaoDoEvento_AcentoImplicaEstarNaLista trava a direcao que de
// fato vale: toda descricao com acento PRECISA estar na lista, senao o
// RetirarAcentos a mutilaria.
//
// O teste varre o catalogo inteiro -- noventa e seis tipos --, e e assim
// que um evento novo com descricao acentuada e esquecido da lista aparece.
func TestDescricaoDoEvento_AcentoImplicaEstarNaLista(t *testing.T) {
	conferidos := 0
	for tipo := pcn.TeNaoMapeado; tipo < 96; tipo++ {
		desc := DescricaoDoTipoEvento(tipo)
		if desc == "" || !temAcento(desc) {
			continue
		}
		conferidos++
		if !EventoLevaAcentos(tipo) {
			t.Errorf("%s tem descricao acentuada (%q) e NAO esta na lista"+
				" de acentos -- o texto sairia mutilado", tipo, desc)
		}
	}
	if conferidos == 0 {
		t.Fatal("nenhuma descricao acentuada foi encontrada; o teste nao" +
			" exercita nada")
	}
}

// temAcento diz se a string tem algum caractere fora do ASCII.
func temAcento(s string) bool {
	for _, r := range s {
		if r > 127 {
			return true
		}
	}
	return false
}

// TestDescEvento_SaiComACENTOSNoXML trava a regra que o ACBr liga em
// :947-957 e que o filtro padrao do XML apagaria.
//
// Este e o unico teste que olha o XML GERADO. Afirmar sobre
// EventoLevaAcentos ou sobre o mapa de descricoes nao basta: os dois podem
// estar certos e o acento morrer no pcn.NodeStr, que translitera para
// ASCII. Foi exatamente o que aconteceu na primeira versao desta etapa.
func TestDescEvento_SaiComACENTOSNoXML(t *testing.T) {
	casos := []struct {
		tipo     pcn.TipoEvento
		preparar func(*DetEvento)
		trecho   string
	}{
		{pcn.TeImporALCZFM, func(d *DetEvento) {},
			"Importação em ALC/ZFM não convertida em isenção"},
		{pcn.TeSolicApropCredPres, func(d *DetEvento) {},
			"Solicitação de Apropriação de crédito presumido"},
		{pcn.TeImobilizacaoItem, func(d *DetEvento) {},
			"Imobilização de Item"},
		{pcn.TeCancConcFinanceira, func(d *DetEvento) { d.NProtEvento = "1" },
			"Cancelamento Conciliação Financeira"},
	}

	for _, c := range casos {
		t.Run(c.tipo.String(), func(t *testing.T) {
			e := eventoBase(c.tipo)
			d := &e.Evento[0].InfEvento.DetEvento
			d.COrgaoAutor, d.VerAplic = 31, "1.0"
			c.preparar(d)

			xml := gerarEventoDeTeste(t, e)
			esperado := "<descEvento>" + c.trecho + "</descEvento>"
			if !strings.Contains(xml, esperado) {
				t.Errorf("esperado %s\nno XML: %s", esperado, xml)
			}
		})
	}
}

// TestDescEvento_ForaDaListaContinuaSemAcento trava a outra metade: os
// eventos classicos passam pelo filtro padrao, que e o que o ACBr faz.
func TestDescEvento_ForaDaListaContinuaSemAcento(t *testing.T) {
	e := eventoBase(pcn.TeCancGenerico)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.TpEventoAut, d.NProtEvento = "110111", "1"

	xml := gerarEventoDeTeste(t, e)
	if !strings.Contains(xml, "<descEvento>Cancelamento de Evento</descEvento>") {
		t.Errorf("descEvento do cancelamento generico: %s", xml)
	}

	// E um evento acentuado NAO contamina o seguinte: a decisao e por
	// evento, diferente do ACBr, onde a flag e global e nao e restaurada.
	lote := eventoBase(pcn.TeImobilizacaoItem)
	lote.Evento[0].InfEvento.DetEvento.COrgaoAutor = 31
	lote.Evento[0].InfEvento.DetEvento.VerAplic = "1.0"
	segundo := eventoBase(pcn.TeCancelamento).Evento[0]
	segundo.InfEvento.DetEvento.NProt = "131260000000001"
	segundo.InfEvento.DetEvento.XJust = "Erro na emissao do documento"
	lote.Evento = append(lote.Evento, segundo)

	xml = gerarEventoDeTeste(t, lote)
	if !strings.Contains(xml, "<descEvento>Imobilização de Item</descEvento>") {
		t.Errorf("o primeiro evento perdeu os acentos: %s", xml)
	}
	if !strings.Contains(xml, "<descEvento>Cancelamento</descEvento>") {
		t.Errorf("o segundo evento deveria seguir sem acento: %s", xml)
	}
}
