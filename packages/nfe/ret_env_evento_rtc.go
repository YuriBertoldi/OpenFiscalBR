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

// Leitura dos grupos repetidos do detEvento.
//
// E aqui que o `case` por tpEvento do Ler_DetEvento (:508) entra, e a
// razao e a mesma do gerador: nomes de elemento que COLIDEM.
//
//	<gConsumo>      vem do consumo pessoal E da importacao em ALC/ZFM
//	<gPerecimento>  vem do transporte do adquirente E do fornecedor
//
// Sem o tpEvento nao ha como saber em qual colecao guardar, e os dois pares
// tem gControleEstoque com tags diferentes. Um leitor que buscasse so pelo
// nome do elemento acertaria metade dos casos -- e a metade errada sairia
// zerada, nao ausente, o que e pior de diagnosticar.

// lerDetEventoGrupos preenche os grupos repetidos conforme o tipo.
// Porte do `case` de Ler_DetEvento (:508-598).
func lerDetEventoGrupos(node *pcn.Node, inf *InfEvento) {
	d := &inf.DetEvento

	switch inf.TpEvento {
	case pcn.TeCancGenerico:
		d.TpEventoAut = pcn.ConteudoStr(node.FindAnyNs("tpEventoAut"))

	case pcn.TeCancSubst:
		d.ChNFeRef = pcn.ConteudoStr(node.FindAnyNs("chNFeRef"))

	case pcn.TeAtualizacaoDataPrevisaoEntrega:
		d.DPrevEntrega = pcn.ConteudoDataDef(node.FindAnyNs("dPrevEntrega"))

	case pcn.TeAceiteDebitoApuracaoNotaCredito,
		pcn.TeManifPedTransfCredIBSSucessao,
		pcn.TeManifPedTransfCredCBSSucessao:
		d.IndAceitacao, _ = ParseIndAceitacao(
			pcn.ConteudoStr(node.FindAnyNs("indAceitacao")))

	case pcn.TeDestItemConsPessoal:
		for i, g := range node.FindAllAnyNs("gConsumo") {
			d.GConsumo = append(d.GConsumo, GConsumo{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoqueConsumo{
					QConsumo: controleEstoqueDecimal(g, "qConsumo"),
					UConsumo: controleEstoqueTexto(g, "uConsumo"),
				},
				DFeReferenciado: lerDFeReferenciadoEvento(g.FindAnyNs("DFeReferenciado")),
			})
		}

	case pcn.TeImporALCZFM:
		// MESMO nome de elemento do caso anterior, colecao e
		// gControleEstoque diferentes.
		for i, g := range node.FindAllAnyNs("gConsumo") {
			d.GConsumoZFM = append(d.GConsumoZFM, GConsumoZFM{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoqueZFM{
					Qtde:    controleEstoqueDecimal(g, "qtde"),
					Unidade: controleEstoqueTexto(g, "unidade"),
				},
			})
		}

	case pcn.TePerecPerdaRouboFurtoTranspContratAqu:
		for i, g := range node.FindAllAnyNs("gPerecimento") {
			d.GPerecimento = append(d.GPerecimento, GPerecimento{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoquePerecimento{
					QPerecimento: controleEstoqueDecimal(g, "qPerecimento"),
					UPerecimento: controleEstoqueTexto(g, "uPerecimento"),
				},
			})
		}

	case pcn.TePerecPerdaRouboFurtoTranspContratFornec:
		// MESMO nome de elemento do caso anterior; o gControleEstoque
		// daqui tem vIBS e vCBS PROPRIOS, que repetem os nomes do grupo
		// pai e por isso precisam ser lidos de dentro dele.
		for i, g := range node.FindAllAnyNs("gPerecimento") {
			ce := g.FindAnyNs("gControleEstoque")
			d.GPerecimentoForn = append(d.GPerecimentoForn, GPerecimentoForn{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoquePerecimentoForn{
					QPerecimento: controleEstoqueDecimal(g, "qPerecimento"),
					UPerecimento: controleEstoqueTexto(g, "uPerecimento"),
					VIBS:         pcn.ConteudoDe2(ce.FindAnyNs("vIBS")),
					VCBS:         pcn.ConteudoDe2(ce.FindAnyNs("vCBS")),
				},
			})
		}

	case pcn.TeFornecNaoRealizPagAntec:
		for i, g := range node.FindAllAnyNs("gItemNaoFornecido") {
			d.GItemNaoFornecido = append(d.GItemNaoFornecido, GItemNaoFornecido{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoqueItemNaoFornecido{
					QNaoFornecida: controleEstoqueDecimal(g, "qNaoFornecida"),
					UNaoFornecida: controleEstoqueTexto(g, "uNaoFornecida"),
				},
			})
		}

	case pcn.TeImobilizacaoItem:
		for i, g := range node.FindAllAnyNs("gImobilizacao") {
			d.GImobilizacao = append(d.GImobilizacao, GImobilizacao{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoqueImobilizacao{
					QImobilizado: controleEstoqueDecimal(g, "qImobilizado"),
					UImobilizado: controleEstoqueTexto(g, "uImobilizado"),
				},
			})
		}

	case pcn.TeSolicApropCredCombustivel:
		for i, g := range node.FindAllAnyNs("gConsumoComb") {
			d.GConsumoComb = append(d.GConsumoComb, GConsumoComb{
				NItem: nItemDoGrupo(g, i),
				VIBS:  pcn.ConteudoDe2(g.FindAnyNs("vIBS")),
				VCBS:  pcn.ConteudoDe2(g.FindAnyNs("vCBS")),
				ControleEstoque: ControleEstoqueComb{
					QComb: controleEstoqueDecimal(g, "qComb"),
					UComb: controleEstoqueTexto(g, "uComb"),
				},
			})
		}

	case pcn.TeSolicApropCredBensServicos:
		for i, g := range node.FindAllAnyNs("gCredito") {
			d.GCredito = append(d.GCredito, GCredito{
				NItem:    nItemDoGrupo(g, i),
				VCredIBS: pcn.ConteudoDe2(g.FindAnyNs("vCredIBS")),
				VCredCBS: pcn.ConteudoDe2(g.FindAnyNs("vCredCBS")),
			})
		}

	case pcn.TeSolicApropCredPres:
		// O elemento e <gCredPresOper>, nao <gCredPres>.
		for i, g := range node.FindAllAnyNs("gCredPresOper") {
			item := GCredPres{
				NItem:        nItemDoGrupo(g, i),
				VBCCredPres:  pcn.ConteudoDe2(g.FindAnyNs("vBCCredPres")),
				GIBSCredPres: lerCredPresPorTributo(g.FindAnyNs("gIBSCredPres")),
				GCBSCredPres: lerCredPresPorTributo(g.FindAnyNs("gCBSCredPres")),
			}
			item.CCredPres, _ = rtc.ParseCCredPres(pcn.ConteudoStr(g.FindAnyNs("cCredPres")))
			d.GCredPres = append(d.GCredPres, item)
		}
	}
}

// nItemDoGrupo le o atributo nItem.
//
// Faltando o atributo, vale a POSICAO comecando em 1 -- que e o que o ACBr
// passa como `i+1` ao ler os grupos (:522 e seguintes). Ele nao usa o
// atributo na leitura, embora o grave na geracao; aqui o atributo vence
// quando existe, porque e o que a SEFAZ devolve.
func nItemDoGrupo(g *pcn.Node, indice int) int {
	if v := g.Attr("nItem"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return indice + 1
}

// controleEstoqueDecimal le um campo numerico de dentro do
// <gControleEstoque>.
//
// As quantidades sao tcDe4 nos sete grupos.
func controleEstoqueDecimal(g *pcn.Node, tag string) float64 {
	ce := g.FindAnyNs("gControleEstoque")
	if ce == nil {
		return 0
	}
	return pcn.ConteudoDe4(ce.FindAnyNs(tag))
}

// controleEstoqueTexto le um campo de texto de dentro do
// <gControleEstoque>.
func controleEstoqueTexto(g *pcn.Node, tag string) string {
	ce := g.FindAnyNs("gControleEstoque")
	if ce == nil {
		return ""
	}
	return pcn.ConteudoStr(ce.FindAnyNs(tag))
}

// lerCredPresPorTributo le gIBSCredPres ou gCBSCredPres.
func lerCredPresPorTributo(node *pcn.Node) IBSCBSCredPres {
	if node == nil {
		return IBSCBSCredPres{}
	}
	return IBSCBSCredPres{
		PCredPres: pcn.ConteudoDe2(node.FindAnyNs("pCredPres")),
		VCredPres: pcn.ConteudoDe2(node.FindAnyNs("vCredPres")),
	}
}

// lerDFeReferenciadoEvento le o <DFeReferenciado> do gConsumo.
func lerDFeReferenciadoEvento(node *pcn.Node) rtc.DFeReferenciado {
	if node == nil {
		return rtc.DFeReferenciado{}
	}
	return rtc.DFeReferenciado{
		ChaveAcesso: pcn.ConteudoStr(node.FindAnyNs("chaveAcesso")),
		NItem:       pcn.ConteudoInt(node.FindAnyNs("nItem")),
	}
}
