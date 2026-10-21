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

// Regras de negocio dos ITENS e dos TOTAIS.
//
// Porte do laco de itens e do bloco de totais de
// ValidarRegrasdeNegocio.pas:695-1200.
//
// As duas coisas vivem juntas porque sao a mesma passagem: o laco que
// confere cada item e o mesmo que ACUMULA os somatorios com que os totais
// serao comparados depois. Separar exigiria percorrer os itens duas vezes e,
// pior, duplicar as regras de quando um valor entra na soma -- que sao a
// parte dificil.

// somatorios guarda o que o laco de itens acumula para as regras de total.
//
// Os nomes seguem os do ACBr (fsvXxx) sem o prefixo, porque cada um
// corresponde a uma tag do grupo ICMSTot e e assim que a conferencia com o
// .pas se faz.
type somatorios struct {
	TotTrib       float64
	BC            float64
	ICMS          float64
	ICMSDeson     float64
	BCST          float64
	ST            float64
	Prod          float64
	Frete         float64
	Seg           float64
	Desc          float64
	II            float64
	IPI           float64
	PIS           float64
	COFINS        float64
	Outro         float64
	Serv          float64
	FCP           float64
	FCPST         float64
	FCPSTRet      float64
	IPIDevol      float64
	PISServico    float64
	COFINSServico float64
	PISST         float64
	COFINSST      float64
	ICMSMonoReten float64

	// FaturamentoDireto muda a FORMULA do total da nota: no faturamento
	// direto de veiculos o ICMS-ST nao entra no valor total.
	FaturamentoDireto bool

	// Importacao desliga a conferencia do total (regra 610).
	Importacao bool

	// UFConsumoDiferente desliga a regra 772.
	UFConsumoDiferente bool
}

// cfopsDeVendaNFCe e a lista fechada de CFOP que a NFC-e aceita
// (ValidarRegrasdeNegocio.pas:719).
//
// O ACBr testa com `pos(CFOP, 'XXXX,5101,...')`, um "contains" de texto: o
// "XXXX" inicial existe so para que um CFOP vazio nao case com o comeco da
// lista. Aqui a comparacao e por membro, que e o que o teste pretendia.
var cfopsDeVendaNFCe = map[string]bool{
	"5101": true, "5102": true, "5103": true, "5104": true, "5115": true,
	"5405": true, "5656": true, "5667": true, "5933": true, "5949": true,
}

// cProdANPGLPRegras e o codigo ANP do GLP, que varias regras tratam a parte.
const cProdANPGLPRegras = 210203001

// cfopsSemRastroMedicamento sao os CFOP excluidos da exigencia de
// rastreabilidade de medicamentos (ValidarRegrasdeNegocio.pas:948).
var cfopsSemRastroMedicamento = map[string]bool{
	"5922": true, "6922": true, "5118": true, "6118": true,
	"5119": true, "6119": true, "5120": true, "6120": true,
}

// regrasDosItens percorre os itens, aplica as regras de cada um e devolve
// os somatorios que as regras de total consomem.
func (c *coletor) regrasDosItens() somatorios {
	n := c.n
	var s somatorios

	for i := range n.Det {
		d := &n.Det[i]
		item := d.Prod.NItem

		// O item e SERVICO quando o NCM e "00" ou quando ha codigo da lista
		// de servicos. A distincao decide em qual somatorio o PIS e o COFINS
		// entram -- e por isso ela vem antes de tudo.
		servico := strings.EqualFold(strings.TrimSpace(d.Prod.NCM), "00") ||
			strings.TrimSpace(d.Imposto.ISSQN.CListServ) != ""

		// 777: mercadoria exige NCM completo; servico nao.
		c.seErroItem(!servico && len(strings.TrimSpace(d.Prod.NCM)) < 8,
			777, item, "Rejeicao: Obrigatoria a informacao do NCM completo")

		if n.Ide.Modelo == 65 {
			c.regrasDoItemNFCe(d, item)
		}

		c.regrasComunsDoItem(d, item)

		if n.InfNFe.Versao >= 4 {
			c.regrasDoItemNa400(d, item)
			if !s.UFConsumoDiferente {
				s.UFConsumoDiferente = strings.TrimSpace(d.Prod.Comb.UFCons) != "" &&
					!strings.EqualFold(d.Prod.Comb.UFCons, n.Emit.EnderEmit.UF)
			}
		}

		acumularItem(&s, d, n, servico)
	}

	// O total de servicos nao vem dos itens: o ACBr o pega direto do grupo
	// ISSQNtot (ValidarRegrasdeNegocio.pas:1018), com o comentario de que a
	// validacao 605 por item nao foi implementada.
	s.Serv = n.Total.ISSQNTot.VServ

	return s
}

// regrasDoItemNFCe sao as proibicoes que valem so no modelo 65
// (ValidarRegrasdeNegocio.pas:708-778).
func (c *coletor) regrasDoItemNFCe(d *DetItem, item int) {
	csosn := d.Imposto.ICMS.CSOSN
	c.seErroItem(csosn == pcn.Csosn101 || csosn == pcn.Csosn201 ||
		csosn == pcn.Csosn202 || csosn == pcn.Csosn203,
		383, item, "Rejeicao: NFC-e Item com CSOSN indevido")

	cst := d.Imposto.ICMS.CST
	c.seErroItem(cst == pcn.CST10 || cst == pcn.CST30 || cst == pcn.CST50 ||
		cst == pcn.CST51 || cst == pcn.CST70,
		766, item, "Rejeicao: NFC-e Item com CST indevido")

	c.seErroItem(!cfopsDeVendaNFCe[pcn.OnlyNumber(d.Prod.CFOP)],
		725, item, "Rejeicao: NFC-e com CFOP invalido")

	c.seErroItem(d.Prod.IndTot == ItNaoSomaTotalNFe,
		774, item, "Rejeicao: NFC-e com indicador de item nao participante do total")

	c.seErroItem(strings.TrimSpace(d.Prod.VeicProd.Chassi) != "",
		736, item, "Rejeicao: NFC-e com grupo de Veiculos novos")
	c.seErroItem(len(d.Prod.Arma) > 0,
		738, item, "Rejeicao: NFC-e com grupo de Armamentos")
	c.seErroItem(strings.TrimSpace(d.Prod.NRECOPI) != "",
		348, item, "Rejeicao: NFC-e com grupo RECOPI")

	c.seErroItem(cst == pcn.CST51, 740, item, "Rejeicao: NFC-e com CST 51-Diferimento")
	c.seErroItem(cst == pcn.CSTPart10 || cst == pcn.CSTPart90,
		741, item, "Rejeicao: NFC-e com Partilha de ICMS entre UF")

	ipi := d.Imposto.IPI
	c.seErroItem(strings.TrimSpace(ipi.CEnq) != "" || diferenteDeZero(ipi.VBC) ||
		diferenteDeZero(ipi.QUnid) || diferenteDeZero(ipi.VUnid) ||
		diferenteDeZero(ipi.PIPI) || diferenteDeZero(ipi.VIPI),
		742, item, "Rejeicao: NFC-e com grupo do IPI")

	ii := d.Imposto.II
	c.seErroItem(maiorQueZero(ii.VBC) || maiorQueZero(ii.VDespAdu) ||
		maiorQueZero(ii.VII) || maiorQueZero(ii.VIOF) ||
		strings.HasPrefix(d.Prod.CFOP, "3"),
		743, item, "Rejeicao: NFC-e com grupo do II")

	pisST := d.Imposto.PISST
	c.seErroItem(maiorQueZero(pisST.VBC) || maiorQueZero(pisST.PPIS) ||
		maiorQueZero(pisST.QBCProd) || maiorQueZero(pisST.VAliqProd) ||
		maiorQueZero(pisST.VPIS),
		746, item, "Rejeicao: NFC-e com grupo do PIS-ST")

	cofinsST := d.Imposto.COFINSST
	c.seErroItem(maiorQueZero(cofinsST.VBC) || maiorQueZero(cofinsST.PCOFINS) ||
		maiorQueZero(cofinsST.QBCProd) || maiorQueZero(cofinsST.VAliqProd) ||
		maiorQueZero(cofinsST.VCOFINS),
		749, item, "Rejeicao: NFC-e com grupo da COFINS-ST")
}

// regrasComunsDoItem valem para os dois modelos
// (ValidarRegrasdeNegocio.pas:835-858).
func (c *coletor) regrasComunsDoItem(d *DetItem, item int) {
	n := c.n
	icms := d.Imposto.ICMS

	// 528: o ICMS tem de ser o produto da base pela aliquota. So nos quatro
	// CST que tem os tres campos, e so em nota de finalidade normal.
	cstComBase := icms.CST == pcn.CST00 || icms.CST == pcn.CST10 ||
		icms.CST == pcn.CST20 || icms.CST == pcn.CST70
	c.seErroItem(cstComBase && n.Ide.FinNFe == FnNormal &&
		pcn.CompararValor(icms.VICMS, icms.VBC*(icms.PICMS/100), pcn.ToleranciaCentavo) != 0,
		528, item, "Rejeicao: Valor do ICMS difere do produto BC e Aliquota")

	// 625: isencao para a Zona Franca exige a inscricao SUFRAMA.
	c.seErroItem(icms.MotDesICMS == MdiSuframa && strings.TrimSpace(n.Dest.ISUF) == "",
		625, item, "Rejeicao: Inscricao SUFRAMA deve ser informada na venda com isencao para ZFM")

	// 530: tributar ISSQN exige Inscricao Municipal do emitente.
	iss := d.Imposto.ISSQN
	temISSQN := maiorQueZero(iss.VBC) || maiorQueZero(iss.VAliq) ||
		maiorQueZero(iss.VISSQN) || maiorQueZero(float64(iss.CMunFG)) ||
		strings.TrimSpace(iss.CListServ) != ""
	c.seErroItem(strings.TrimSpace(n.Emit.IM) == "" && temISSQN,
		530, item, "Rejeicao: Operacao com tributacao de ISSQN sem informar a Inscricao Municipal")

	c.seErroItem(iss.CMunFG > 0 && !pcn.ValidarMunicipio(iss.CMunFG),
		287, item, "Rejeicao: Codigo Municipio do FG - ISSQN: digito invalido")
}

// regrasDoItemNa400 sao as que a 4.00 acrescentou
// (ValidarRegrasdeNegocio.pas:860-962).
func (c *coletor) regrasDoItemNa400(d *DetItem, item int) {
	n := c.n
	p := d.Prod

	// O par cEAN/cEANTrib: cada um, quando informado e diferente de
	// "SEM GTIN", exige digito valido, prefixo valido E o outro preenchido.
	//
	// As regras 883 e 888 (campo sem informacao) estao COMENTADAS no ACBr,
	// com a nota de que so valem em producao a partir de 01/12/2018. Nao
	// portadas, pelo mesmo motivo.
	if strings.TrimSpace(p.CEAN) != "" && !strings.EqualFold(p.CEAN, semGTIN) {
		c.seErroItem(!pcn.ValidarGTIN(p.CEAN), 611, item, "Rejeicao: GTIN (cEAN) invalido")
		c.seErroItem(!pcn.ValidarPrefixoGTIN(p.CEAN), 882, item,
			"Rejeicao: GTIN (cEAN) com prefixo invalido")
		c.seErroItem(strings.TrimSpace(p.CEANTrib) == "" ||
			strings.EqualFold(strings.TrimSpace(p.CEANTrib), semGTIN),
			885, item, "Rejeicao: GTIN informado, mas nao informado o GTIN da unidade tributavel")
	}

	if strings.TrimSpace(p.CEANTrib) != "" && !strings.EqualFold(p.CEANTrib, semGTIN) {
		c.seErroItem(!pcn.ValidarGTIN(p.CEANTrib), 612, item,
			"Rejeicao: GTIN da unidade tributavel (cEANTrib) invalido")
		c.seErroItem(!pcn.ValidarPrefixoGTIN(p.CEANTrib), 884, item,
			"Rejeicao: GTIN da unidade tributavel (cEANTrib) com prefixo invalido")
		c.seErroItem(strings.TrimSpace(p.CEAN) == "" ||
			strings.EqualFold(strings.TrimSpace(p.CEAN), semGTIN),
			886, item, "Rejeicao: GTIN da unidade tributavel informado, mas nao informado o GTIN")
	}

	// 889: esta, ao contrario das comentadas acima, esta ATIVA -- o cEAN
	// vazio e rejeitado. Como o gerador preenche "SEM GTIN" nas notas
	// emitidas depois de 01/12/2018, na pratica ela pega quem monta o
	// documento na mao e nao preenche nada.
	c.seErroItem(strings.TrimSpace(p.CEAN) == "",
		889, item, "Rejeicao: Obrigatoria a informacao do GTIN para o produto")

	// 879 e 489: producao em escala nao relevante exige o CNPJ do fabricante.
	c.seErroItem(p.IndEscala == IeNaoRelevante && strings.TrimSpace(p.CNPJFab) == "",
		879, item, "Rejeicao: Informado item Produzido em Escala NAO Relevante e nao informado CNPJ do Fabricante")
	c.seErroItem(strings.TrimSpace(p.CNPJFab) != "" && pcn.ValidarCNPJ(p.CNPJFab) != nil,
		489, item, "Rejeicao: CNPJFab informado invalido (DV ou zeros)")

	// 854: o GLP e vendido em quilos.
	c.seErroItem(p.Comb.CProdANP == cProdANPGLPRegras && !strings.EqualFold(p.UTrib, "KG"),
		854, item, "Rejeicao: Unidade Tributavel (tag:uTrib) incompativel com produto informado")

	// 877 e 870: coerencia das datas de rastreabilidade.
	for _, r := range p.Rastro {
		c.seErroItem(r.DFab.After(n.Ide.DEmi),
			877, item, "Rejeicao: Data de fabricacao maior que a data de processamento")
		c.seErroItem(r.DVal.Before(r.DFab),
			870, item, "Rejeicao: Data de validade incompativel com data de fabricacao")
	}

	// 873: medicamento exige rastreabilidade, com QUATRO excecoes.
	for _, m := range p.Med {
		if strings.TrimSpace(m.CProdANVISA) == "" || len(p.Rastro) > 0 {
			continue
		}
		excecaoFinalidade := n.Ide.FinNFe == FnDevolucao || n.Ide.FinNFe == FnAjuste ||
			n.Ide.FinNFe == FnComplementar
		excecaoPresenca := n.Ide.IndPres == PcInternet || n.Ide.IndPres == PcTeleatendimento
		excecaoCFOP := cfopsSemRastroMedicamento[p.CFOP]
		excecaoEntrada := n.Ide.TpNF != pcn.TnSaida

		c.seErroItem(!excecaoFinalidade && !excecaoPresenca && !excecaoCFOP && !excecaoEntrada,
			873, item, "Rejeicao: Operacao com medicamentos e nao informado os campos de rastreabilidade")
	}

	// 461 e 855: os percentuais de gas so existem no GLP, e la somam 100.
	temPercentuaisGas := p.Comb.PGLP > 0 || p.Comb.PGNn > 0 || p.Comb.PGNi > 0
	c.seErroItem(p.Comb.CProdANP != cProdANPGLPRegras && temPercentuaisGas,
		461, item, "Rejeicao: Informado campos de percentual de GLP e/ou GLGNn e/ou GLGNi para produto diferente de GLP")
	c.seErroItem(p.Comb.CProdANP == cProdANPGLPRegras &&
		p.Comb.PGLP+p.Comb.PGNn+p.Comb.PGNi != 100,
		855, item, "Rejeicao: Somatorio percentuais de GLP derivado do petroleo, GLGNn e GLGNi diferente de 100")
}

// acumularItem soma o item aos totalizadores.
// Porte de ValidarRegrasdeNegocio.pas:964-1013.
//
// A divisao e em tres niveis, e e ela que torna as regras de total dificeis
// de acertar:
//
//  1. frete, seguro, desconto, outros, II, IPI e a desoneracao entram SEMPRE,
//     mesmo em item que nao compoe o total;
//  2. PIS e COFINS entram em UM DE DOIS somatorios, conforme o item seja
//     mercadoria ou servico;
//  3. o resto so entra quando indTot manda somar -- e, dentro disso, o valor
//     do produto ainda tem a excecao da nota de ajuste.
func acumularItem(s *somatorios, d *DetItem, n *NFe, servico bool) {
	s.TotTrib += d.Imposto.VTotTrib
	s.Frete += d.Prod.VFrete
	s.Seg += d.Prod.VSeg
	s.Outro += d.Prod.VOutro
	s.Desc += d.Prod.VDesc
	s.II += d.Imposto.II.VII
	s.IPI += d.Imposto.IPI.VIPI
	s.IPIDevol += d.VIPIDevol
	s.ICMSDeson += d.Imposto.ICMS.VICMSDeson

	if servico {
		s.PISServico += d.Imposto.PIS.VPIS
		s.COFINSServico += d.Imposto.COFINS.VCOFINS
	} else {
		s.PIS += d.Imposto.PIS.VPIS
		s.COFINS += d.Imposto.COFINS.VCOFINS
	}

	if d.Prod.IndTot == ItSomaTotalNFe {
		s.BC += d.Imposto.ICMS.VBC
		s.ICMS += d.Imposto.ICMS.VICMS
		s.BCST += d.Imposto.ICMS.VBCST
		s.ST += d.Imposto.ICMS.VICMSST
		s.FCP += d.Imposto.ICMS.VFCP
		s.FCPST += d.Imposto.ICMS.VFCPST
		s.FCPSTRet += d.Imposto.ICMS.VFCPSTRet
		s.ICMSMonoReten += d.Imposto.ICMS.VICMSMonoReten

		// O ST de PIS e COFINS so compoe o total quando o indicador manda.
		if d.Imposto.PISST.IndSomaPISST == IspPISSTCompoe {
			s.PISST += d.Imposto.PISST.VPIS
		}
		if d.Imposto.COFINSST.IndSomaCOFINSST == IscCOFINSSTCompoe {
			s.COFINSST += d.Imposto.COFINSST.VCOFINS
		}

		// Servico nao soma em vProd -- EXCETO na nota de ajuste.
		if !servico || n.Ide.FinNFe == FnAjuste {
			s.Prod += d.Prod.VProd
		}
	}

	if d.Prod.VeicProd.TpOP == ToFaturamentoDireto {
		s.FaturamentoDireto = true
	}
	if strings.HasPrefix(d.Prod.CFOP, "3") {
		s.Importacao = true
	}
}

// diferenteDeZero e o ComparaValor(x, 0, tolerancia) <> 0 do ACBr.
func diferenteDeZero(v float64) bool {
	return pcn.CompararValor(v, 0, pcn.ToleranciaMilesimo) != 0
}
