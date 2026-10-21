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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Regras de TOTAIS e de PAGAMENTO.
// Porte de ValidarRegrasdeNegocio.pas:1020-1200.

// regrasDosTotais compara cada tag do ICMSTot com o somatorio dos itens.
func (c *coletor) regrasDosTotais(s somatorios) {
	n := c.n
	t := n.Total.ICMSTot

	// 772: operacao interestadual com a mesma UF nos dois lados. A regra so
	// vale quando NENHUM item declarou UF de consumo diferente -- o
	// combustivel consumido em outra UF e caso legitimo.
	if !s.UFConsumoDiferente {
		c.seErro(n.Ide.IdDest == DoInterestadual &&
			strings.EqualFold(n.Dest.EnderDest.UF, n.Emit.EnderEmit.UF) &&
			!strings.EqualFold(n.Dest.CNPJCPF, n.Emit.CNPJCPF) &&
			strings.EqualFold(n.Entrega.UF, n.Emit.EnderEmit.UF),
			772, "Rejeicao: Operacao Interestadual e UF de destino igual a UF do emitente")
	}

	// Os dezessete confrontos diretos entre total e somatorio.
	confrontos := []struct {
		codigo    int
		total     float64
		somatorio float64
		mensagem  string
	}{
		{531, t.VBC, s.BC, "Rejeicao: Total da BC ICMS difere do somatorio dos itens"},
		{532, t.VICMS, s.ICMS, "Rejeicao: Total do ICMS difere do somatorio dos itens"},
		{795, t.VICMSDeson, s.ICMSDeson, "Rejeicao: Total do ICMS desonerado difere do somatorio dos itens"},
		{533, t.VBCST, s.BCST, "Rejeicao: Total da BC ICMS-ST difere do somatorio dos itens"},
		{534, t.VST, s.ST, "Rejeicao: Total do ICMS-ST difere do somatorio dos itens"},
		{564, t.VProd, s.Prod, "Rejeicao: Total do Produto / Servico difere do somatorio dos itens"},
		{535, t.VFrete, s.Frete, "Rejeicao: Total do Frete difere do somatorio dos itens"},
		{536, t.VSeg, s.Seg, "Rejeicao: Total do Seguro difere do somatorio dos itens"},
		{537, t.VDesc, s.Desc, "Rejeicao: Total do Desconto difere do somatorio dos itens"},
		{601, t.VII, s.II, "Rejeicao: Total do II difere do somatorio dos itens"},
		{538, t.VIPI, s.IPI, "Rejeicao: Total do IPI difere do somatorio dos itens"},
		{602, t.VPIS, s.PIS, "Rejeicao: Total do PIS difere do somatorio dos itens sujeitos ao ICMS"},
		{603, t.VCOFINS, s.COFINS, "Rejeicao: Total da COFINS difere do somatorio dos itens sujeitos ao ICMS"},
		{604, t.VOutro, s.Outro, "Rejeicao: Total do vOutro difere do somatorio dos itens"},
		{608, n.Total.ISSQNTot.VPIS, s.PISServico, "Rejeicao: Total do PIS difere do somatorio dos itens sujeitos ao ISSQN"},
		{609, n.Total.ISSQNTot.VCOFINS, s.COFINSServico, "Rejeicao: Total da COFINS difere do somatorio dos itens sujeitos ao ISSQN"},
		{861, t.VFCP, s.FCP, "Rejeicao: Total do FCP difere do somatorio dos itens"},
	}
	for _, cf := range confrontos {
		c.seErro(pcn.CompararValor(cf.total, cf.somatorio, pcn.ToleranciaMilesimo) != 0,
			cf.codigo, cf.mensagem)
	}

	// Tres totais so sao conferidos na NF-e.
	if n.Ide.Modelo == 55 {
		c.seErro(pcn.CompararValor(t.VFCPST, s.FCPST, pcn.ToleranciaMilesimo) != 0,
			862, "Rejeicao: Total do FCP ST difere do somatorio dos itens")
		c.seErro(pcn.CompararValor(t.VFCPSTRet, s.FCPSTRet, pcn.ToleranciaMilesimo) != 0,
			859, "Rejeicao: Total do FCP retido anteriormente por Substituicao Tributaria difere do somatorio dos itens")
		c.seErro(pcn.CompararValor(t.VIPIDevol, s.IPIDevol, pcn.ToleranciaMilesimo) != 0,
			863, "Rejeicao: Total do IPI devolvido difere do somatorio dos itens")
	}

	c.regraDoTotalDaNota(s)

	c.seErro(pcn.CompararValor(t.VTotTrib, s.TotTrib, pcn.ToleranciaMilesimo) != 0,
		685, "Rejeicao: Total do Valor Aproximado dos Tributos difere do somatorio dos itens")
}

// regraDoTotalDaNota e a 610, a mais complicada do conjunto
// (ValidarRegrasdeNegocio.pas:1030-1127).
//
// Tres coisas a tornam diferente dos outros confrontos:
//
//  1. a FORMULA muda com o faturamento direto de veiculos, que tira o ICMS-ST
//     e varios outros componentes da conta;
//  2. nota de IMPORTACAO nao e conferida -- qualquer item com CFOP iniciado
//     por 3 desliga a regra inteira;
//  3. a tolerancia e DUPLA: primeiro o milesimo e, se falhar, uma segunda
//     chance com 0.009 aceitando o total COM ou SEM a desoneracao do ICMS.
//     E o jeito do ACBr de conviver com as duas interpretacoes que as UFs
//     deram ao vICMSDeson.
func (c *coletor) regraDoTotalDaNota(s somatorios) {
	if s.Importacao {
		return
	}

	var esperado float64
	if s.FaturamentoDireto {
		esperado = (s.Prod + s.Frete + s.Seg + s.Outro + s.II + s.IPI +
			s.Serv + s.PISST + s.COFINSST) - (s.Desc + s.ICMSDeson)
	} else {
		esperado = (s.Prod + s.ST + s.Frete + s.Seg + s.Outro + s.II + s.IPI +
			s.Serv + s.FCPST + s.ICMSMonoReten + s.IPIDevol + s.PISST +
			s.COFINSST) - (s.Desc + s.ICMSDeson)
	}

	declarado := c.n.Total.ICMSTot.VNF
	if pcn.CompararValor(declarado, esperado, pcn.ToleranciaMilesimo) == 0 {
		return
	}

	const toleranciaAmpliada = 0.009
	semDesoneracao := pcn.CompararValor(declarado, esperado, toleranciaAmpliada) != 0
	comDesoneracao := pcn.CompararValor(declarado, esperado+s.ICMSDeson, toleranciaAmpliada) != 0

	c.seErro(semDesoneracao && comDesoneracao,
		610, "Rejeicao: Total da NF difere do somatorio dos Valores compoe o valor Total da NF.")
}

// regrasDePagamento confere o grupo pag contra o total da nota
// (ValidarRegrasdeNegocio.pas:1133-1200).
//
// O corte por versao e por modelo deixa TRES caminhos distintos, e so um
// deles roda em cada documento.
func (c *coletor) regrasDePagamento(s somatorios) {
	n := c.n

	totalPago := 0.0
	for _, p := range n.Pag.DetPag {
		totalPago += p.VPag
	}

	// Caminho 1: NFC-e abaixo da 4.00 exige igualdade exata com o total.
	if n.Ide.Modelo == 65 && n.InfNFe.Versao < 4 {
		c.seErro(pcn.CompararValor(n.Total.ICMSTot.VNF, totalPago, pcn.ToleranciaMilesimo) != 0,
			767, "Rejeicao: NFC-e com somatorio dos pagamentos diferente do total da Nota Fiscal")
		return
	}

	if n.InfNFe.Versao < 4 {
		return
	}

	// Caminho 2 e 3: da 4.00 em diante, a finalidade decide.
	switch n.Ide.FinNFe {
	case FnNormal, FnComplementar:
		// A regra 767 nesta faixa foi REMOVIDA pela NT 2016.002 v1.10 e
		// esta comentada no ACBr. Nao portada.

		if n.Ide.Modelo == 65 {
			// 899: na NFC-e nao existe venda sem pagamento. O ACBr para no
			// primeiro item que viola.
			for _, p := range n.Pag.DetPag {
				if p.TPag == FpSemPagamento {
					c.erro(899, "Rejeicao: Informado incorretamente o campo meio de pagamento")
					break
				}
			}

			c.seErro(pcn.CompararValor(totalPago, n.Total.ICMSTot.VNF, pcn.ToleranciaMilesimo) < 0,
				865, "Rejeicao: Total dos pagamentos menor que o total da nota")
		}

		// 866 e 869: a aritmetica do troco, que vale para os DOIS modelos.
		semTroco := pcn.CompararValor(n.Pag.VTroco, 0, pcn.ToleranciaMilesimo) == 0
		pagouMais := pcn.CompararValor(totalPago, n.Total.ICMSTot.VNF, pcn.ToleranciaMilesimo) > 0
		c.seErro(semTroco && pagouMais,
			866, "Rejeicao: Ausencia de troco quando o valor dos pagamentos informados for maior que o total da nota")

		temTroco := pcn.CompararValor(n.Pag.VTroco, 0, pcn.ToleranciaMilesimo) > 0
		trocoErrado := pcn.CompararValor(n.Total.ICMSTot.VNF, totalPago-n.Pag.VTroco,
			pcn.ToleranciaMilesimo) != 0
		c.seErro(temTroco && trocoErrado, 869, "Rejeicao: Valor do troco incorreto")

	case FnDevolucao:
		// 871: devolucao nao tem pagamento. Aqui o ACBr NAO para no primeiro
		// -- acrescenta uma rejeicao por item que violar.
		for _, p := range n.Pag.DetPag {
			c.seErro(p.TPag != FpSemPagamento,
				871, "Rejeicao: O campo Meio de Pagamento deve ser preenchido com a opcao Sem Pagamento")
		}
	}
}
