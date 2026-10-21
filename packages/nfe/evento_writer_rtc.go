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
	"strconv"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Geradores dos quatorze eventos da Reforma Tributaria.
//
// Eles seguem um molde comum -- descEvento, cOrgaoAutor, tpAutor, verAplic
// e depois um grupo repetido --, e e justamente por isso que o porte e
// arriscado: a diferenca entre eles esta no grupo repetido, e varios
// grupos saem com o MESMO nome de elemento carregando tags diferentes.
//
// Tres coincidencias de nome que um porte achatado perde:
//
//	gConsumo          vem de GConsumo (consumo pessoal) E de GConsumoZFM
//	gPerecimento      vem de GPerecimento E de GPerecimentoForn
//	gControleEstoque  tem SETE formas, uma por grupo pai
//
// Em todos os casos o leitor acha a tag pelo nome e nao percebe a troca: o
// documento sai aceito pelo parse e rejeitado pelo schema.

// gerarDetEventoReformaTributaria preenche o detEvento dos eventos da
// Reforma e devolve true quando o tipo e um deles.
//
// Devolver bool, e nao um erro, mantem o despacho em um lugar so: quem
// chama ja sabe transformar "nao e meu" em erro.
func gerarDetEventoReformaTributaria(det *pcn.Elem, inf InfEvento, desc string, acentos bool) bool {
	d := inf.DetEvento

	switch inf.TpEvento {
	case pcn.TeCancGenerico:
		// Unico da Reforma SEM tpAutor (:2464).
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
			Filho(pcn.NodeStr("tpEventoAut", d.TpEventoAut, true)).
			Filho(pcn.NodeStr("nProtEvento", d.NProtEvento, true))

	case pcn.TePagIntegLibCredPresAdq:
		// O indQuitacao e FIXO "1" no ACBr (:2503) -- nao ha campo no
		// documento para ele.
		cabecalhoReforma(det, d, desc, acentos).
			Filho(pcn.NodeStr("indQuitacao", indQuitacaoFixoDoSchema, true))

	case pcn.TeImporALCZFM:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GConsumoZFM {
			det.Filho(gerarGConsumoZFM(g))
		}

	case pcn.TePerecPerdaRouboFurtoTranspContratFornec:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GPerecimentoForn {
			det.Filho(gerarGPerecimentoForn(g))
		}

	case pcn.TeFornecNaoRealizPagAntec:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GItemNaoFornecido {
			det.Filho(gerarGItemNaoFornecido(g))
		}

	case pcn.TeAtualizacaoDataPrevisaoEntrega:
		cabecalhoReforma(det, d, desc, acentos).
			Filho(pcn.NodeDat("dPrevEntrega", d.DPrevEntrega, true))

	case pcn.TeSolicApropCredPres:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GCredPres {
			det.Filho(gerarGCredPres(g))
		}

	case pcn.TeDestItemConsPessoal:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GConsumo {
			det.Filho(gerarGConsumo(g))
		}

	case pcn.TePerecPerdaRouboFurtoTranspContratAqu:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GPerecimento {
			det.Filho(gerarGPerecimento(g))
		}

	case pcn.TeAceiteDebitoApuracaoNotaCredito:
		cabecalhoReforma(det, d, desc, acentos).
			Filho(pcn.NodeStr("indAceitacao", d.IndAceitacao.String(), true))

	case pcn.TeImobilizacaoItem:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GImobilizacao {
			det.Filho(gerarGImobilizacao(g))
		}

	case pcn.TeSolicApropCredCombustivel:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GConsumoComb {
			det.Filho(gerarGConsumoComb(g))
		}

	case pcn.TeSolicApropCredBensServicos:
		cabecalhoReforma(det, d, desc, acentos)
		for _, g := range d.GCredito {
			det.Filho(gerarGCredito(g))
		}

	case pcn.TeManifPedTransfCredIBSSucessao, pcn.TeManifPedTransfCredCBSSucessao:
		cabecalhoReforma(det, d, desc, acentos).
			Filho(pcn.NodeStr("indAceitacao", d.IndAceitacao.String(), true))

	default:
		return false
	}
	return true
}

// indQuitacaoFixoDoSchema e o "1" literal de :2503. Como o tpAutor fixo dos
// eventos classicos, ele ignora qualquer campo do documento -- so que aqui
// nem campo existe.
const indQuitacaoFixoDoSchema = "1"

// cabecalhoReforma emite os quatro campos que abrem quase todo evento da
// Reforma e devolve o proprio detEvento, para encadear.
//
// Diferente dos eventos classicos, aqui o tpAutor vem do CAMPO -- todos os
// geradores da Reforma usam TipoAutorToStr, nenhum fixa "1".
func cabecalhoReforma(det *pcn.Elem, d DetEvento, desc string, acentos bool) *pcn.Elem {
	return det.
		Filho(nodeDescEvento(desc, acentos)).
		Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
		Filho(pcn.NodeStr("tpAutor", d.TpAutor.String(), true)).
		Filho(pcn.NodeStr("verAplic", d.VerAplic, true))
}

// gerarGCredPres monta um <gCredPresOper>.
// Porte de Gerar_gCredPres (:2537).
//
// O nome do elemento NAO e "gCredPres", apesar do nome da classe Delphi.
func gerarGCredPres(g GCredPres) *pcn.Elem {
	return pcn.NovoElem("gCredPresOper").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vBCCredPres", g.VBCCredPres, 2, true)).
		Filho(pcn.NodeStr("cCredPres", g.CCredPres.String(), true)).
		Filho(gerarCredPresPorTributo("gIBSCredPres", g.GIBSCredPres)).
		Filho(gerarCredPresPorTributo("gCBSCredPres", g.GCBSCredPres))
}

// gerarCredPresPorTributo monta gIBSCredPres ou gCBSCredPres, que tem as
// MESMAS duas tags e diferem so no nome do elemento (:2565 e :2576).
func gerarCredPresPorTributo(nome string, g IBSCBSCredPres) *pcn.Elem {
	return pcn.NovoElem(nome).
		Filho(pcn.NodeDec("pCredPres", g.PCredPres, 2, true)).
		Filho(pcn.NodeDec("vCredPres", g.VCredPres, 2, true))
}

// gerarGConsumo monta o <gConsumo> do evento de destinacao para consumo
// pessoal. Porte de Gerar_gConsumo (:2618).
//
// E o unico grupo da Reforma que carrega DFeReferenciado.
func gerarGConsumo(g GConsumo) *pcn.Elem {
	return pcn.NovoElem("gConsumo").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qConsumo", g.ControleEstoque.QConsumo, 4, true)).
			Filho(pcn.NodeStr("uConsumo", g.ControleEstoque.UConsumo, true))).
		Filho(gerarDFeReferenciadoEvento(g.DFeReferenciado))
}

// gerarGConsumoZFM monta o <gConsumo> da importacao em ALC/ZFM.
// Porte de Gerar_gConsumoZFM (:2902).
//
// Mesmo nome de elemento do anterior, gControleEstoque DIFERENTE: aqui as
// tags sao qtde e unidade, sem nomear a mercadoria.
func gerarGConsumoZFM(g GConsumoZFM) *pcn.Elem {
	return pcn.NovoElem("gConsumo").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qtde", g.ControleEstoque.Qtde, 4, true)).
			Filho(pcn.NodeStr("unidade", g.ControleEstoque.Unidade, true)))
}

// gerarGPerecimento monta o <gPerecimento> do transporte contratado pelo
// ADQUIRENTE. Porte de Gerar_gPerecimento (:3024).
func gerarGPerecimento(g GPerecimento) *pcn.Elem {
	return pcn.NovoElem("gPerecimento").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qPerecimento", g.ControleEstoque.QPerecimento, 4, true)).
			Filho(pcn.NodeStr("uPerecimento", g.ControleEstoque.UPerecimento, true)))
}

// gerarGPerecimentoForn monta o <gPerecimento> do transporte contratado
// pelo FORNECEDOR. Porte de Gerar_gPerecimentoForn (:3050).
//
// Mesmo nome de elemento do anterior; o gControleEstoque traz DOIS campos a
// mais, vIBS e vCBS, repetindo os nomes que o grupo pai ja usa (:2390).
// Copiar o gerador do adquirente perde os dois em silencio.
func gerarGPerecimentoForn(g GPerecimentoForn) *pcn.Elem {
	return pcn.NovoElem("gPerecimento").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qPerecimento", g.ControleEstoque.QPerecimento, 4, true)).
			Filho(pcn.NodeStr("uPerecimento", g.ControleEstoque.UPerecimento, true)).
			Filho(pcn.NodeDec("vIBS", g.ControleEstoque.VIBS, 2, true)).
			Filho(pcn.NodeDec("vCBS", g.ControleEstoque.VCBS, 2, true)))
}

// gerarGImobilizacao monta o <gImobilizacao>.
// Porte de Gerar_gImobilizacao (:2746).
func gerarGImobilizacao(g GImobilizacao) *pcn.Elem {
	return pcn.NovoElem("gImobilizacao").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qImobilizado", g.ControleEstoque.QImobilizado, 4, true)).
			Filho(pcn.NodeStr("uImobilizado", g.ControleEstoque.UImobilizado, true)))
}

// gerarGConsumoComb monta o <gConsumoComb>.
// Porte de Gerar_gConsumoComb (:2877).
func gerarGConsumoComb(g GConsumoComb) *pcn.Elem {
	return pcn.NovoElem("gConsumoComb").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qComb", g.ControleEstoque.QComb, 4, true)).
			Filho(pcn.NodeStr("uComb", g.ControleEstoque.UComb, true)))
}

// gerarGCredito monta o <gCredito>.
// Porte de Gerar_gCredito (:2670).
//
// Unico grupo da Reforma SEM gControleEstoque, e os valores tem nome
// proprio: vCredIBS e vCredCBS. Copiar o molde dos vizinhos emitiria vIBS e
// vCBS.
func gerarGCredito(g GCredito) *pcn.Elem {
	return pcn.NovoElem("gCredito").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vCredIBS", g.VCredIBS, 2, true)).
		Filho(pcn.NodeDec("vCredCBS", g.VCredCBS, 2, true))
}

// gerarGItemNaoFornecido monta o <gItemNaoFornecido>.
// Porte de Gerar_gItemNaoFornecido (:2772).
func gerarGItemNaoFornecido(g GItemNaoFornecido) *pcn.Elem {
	return pcn.NovoElem("gItemNaoFornecido").
		Attr("nItem", strconv.Itoa(g.NItem)).
		Filho(pcn.NodeDec("vIBS", g.VIBS, 2, true)).
		Filho(pcn.NodeDec("vCBS", g.VCBS, 2, true)).
		Filho(pcn.NovoElem("gControleEstoque").
			Filho(pcn.NodeDec("qNaoFornecida", g.ControleEstoque.QNaoFornecida, 4, true)).
			Filho(pcn.NodeStr("uNaoFornecida", g.ControleEstoque.UNaoFornecida, true)))
}

// gerarDFeReferenciadoEvento monta o <DFeReferenciado> do gConsumo.
// Porte de Gerar_DFeReferenciado (:2658).
//
// Repare na caixa: "DFeReferenciado" comeca com maiuscula, ao contrario de
// todas as outras tags de grupo do leiaute.
func gerarDFeReferenciadoEvento(d rtc.DFeReferenciado) *pcn.Elem {
	return pcn.NovoElem("DFeReferenciado").
		Filho(pcn.NodeStr("chaveAcesso", pcn.RemoverLiteralChave(d.ChaveAcesso), true)).
		Filho(pcn.NodeInt("nItem", d.NItem, 1, true))
}
