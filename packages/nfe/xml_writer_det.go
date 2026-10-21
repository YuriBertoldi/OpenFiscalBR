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
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Gerador do grupo det (produto + impostos do item).
//
// Porte de GerarDet e dos 24 metodos que ele encadeia, em
// ACBrNFe.XmlWriter.pas:1118-3507. Mora separado do xml_writer.go porque
// sozinho responde por mais da metade do gerador -- o GerarDetImpostoICMS
// tem 1043 linhas no original.
//
// Tres coisas distinguem este grupo do resto do documento:
//
//  1. E o unico lugar do gerador com condicional POR DATA (o "SEM GTIN" de
//     01/12/2018). Todo o resto decide por versao de leiaute.
//  2. O nome da tag de ICMS e CALCULADO a partir do CST/CSOSN -- nao existe
//     tag fixa "ICMS00"; existe um mapa CST -> sufixo em que cinco CSTs
//     distintos caem no mesmo sufixo.
//  3. Ele MUTA o documento em quatro campos (cEAN, cEANTrib, xProd e cEnq),
//     como o ACBr. Ver gerarDetProd.

const (
	// xProdHomologacao e o HOM_XPROD do ACBr: a descricao que SUBSTITUI a do
	// primeiro item quando a nota vai para o ambiente de homologacao.
	xProdHomologacao = "NOTA FISCAL EMITIDA EM AMBIENTE DE HOMOLOGACAO - SEM VALOR FISCAL"

	// semGTIN e o literal que o leiaute 4.00 exige quando o produto nao tem
	// codigo de barras.
	semGTIN = "SEM GTIN"
)

// inicioObrigatoriedadeGTIN e a data a partir da qual o cEAN vazio vira
// "SEM GTIN" (XmlWriter.pas:1163 e :1237).
//
// A comparacao do ACBr e CompareDate, que ignora a hora, e e ESTRITA: uma
// nota emitida EM 01/12/2018 ainda sai com cEAN vazio.
var inicioObrigatoriedadeGTIN = time.Date(2018, 12, 1, 0, 0, 0, 0, time.UTC)

// gerarDet monta um item da nota. Porte de GerarDet (XmlWriter.pas:1118).
//
// O nItem sai como ATRIBUTO do det e vem de Prod.NItem -- nao do indice do
// laco. Quem le o XML preenche esse campo com um contador (ver xml_reader.go),
// entao o round-trip fecha; quem monta a nota na mao precisa numerar os itens.
func (w *xmlWriter) gerarDet(i int) *pcn.Elem {
	d := &w.n.Det[i]

	det := pcn.NovoElem("det").
		Attr("nItem", strconv.Itoa(d.Prod.NItem)).
		Filho(w.gerarDetProd(i)).
		Filho(w.gerarDetImposto(i))

	// impostoDevol so existe com percentual de devolucao (XmlWriter.pas:1130).
	if d.PDevol > 0 {
		det.Filho(w.gerarDetDevol(i))
	}

	det.Filho(pcn.NodeStr("infAdProd", d.InfAdProd, false)).
		Filho(w.gerarDetObsItem(i)).
		Filho(pcn.NodeDec("vItem", d.VItem, 2, false)).
		Filho(w.rtc.GerarDFeReferenciado(d.DFeReferenciado))

	return det
}

// gerarDetProd monta o produto do item. Porte de GerarDetProd
// (XmlWriter.pas:1149).
//
// MUTA o documento em tres campos, como o original:
//
//   - cEAN e cEANTrib vazios viram "SEM GTIN" (so a partir da 4.00 e so
//     para notas emitidas DEPOIS de 01/12/2018);
//   - xProd do PRIMEIRO item vira o aviso de homologacao quando o ambiente e
//     de teste e o documento e NFC-e, ou NF-e com DANFE simplificado.
//
// A mutacao e o comportamento do ACBr e nao da para contorna-la sem mudar a
// saida: o proprio ACBr releva o campo logo depois para decidir se valida o
// GTIN.
func (w *xmlWriter) gerarDetProd(i int) *pcn.Elem {
	n := w.n
	p := &n.Det[i].Prod

	prod := pcn.NovoElem("prod").
		Filho(pcn.NodeStr("cProd", p.CProd, true))

	if w.versao() >= 4 && strings.TrimSpace(p.CEAN) == "" && w.depoisDaObrigatoriedadeGTIN() {
		p.CEAN = semGTIN
	}
	prod.Filho(pcn.NodeStr("cEAN", p.CEAN, true)).
		Filho(pcn.NodeStr("cBarra", p.CBarra, false))

	// O aviso de homologacao so entra no PRIMEIRO item, e so quando o DANFE
	// nao tem espaco para carrega-lo em outro lugar (XmlWriter.pas:1180).
	if p.NItem == 1 && n.Ide.TpAmb == pcn.TaHomologacao &&
		(n.Ide.Modelo == 65 ||
			(n.Ide.Modelo == 55 && n.Ide.TpImp == pcn.TiSimplificadoTipo2)) {
		p.XProd = xProdHomologacao
	}
	prod.Filho(pcn.NodeStr("xProd", p.XProd, true))

	// NCM so e obrigatorio a partir da 2.00 (XmlWriter.pas:1188).
	prod.Filho(pcn.NodeStr("NCM", p.NCM, w.versao() >= 2))

	for _, nve := range p.NVE {
		prod.Filho(pcn.NodeStr("NVE", nve.NVE, false))
	}

	if w.versao() >= 4 {
		// indEscala e CNPJFab sao FILHOS LOGICOS do CEST: sem CEST, nenhum
		// dos tres sai. Fora do if, o indEscala default ("N", de NovoProd)
		// vazaria para todo item sem substituicao tributaria.
		if cest := pcn.OnlyNumber(p.CEST); cest != "" {
			prod.Filho(pcn.NodeStr("CEST", cest, true)).
				Filho(pcn.NodeStr("indEscala", p.IndEscala.String(), false)).
				Filho(pcn.NodeStr("CNPJFab", p.CNPJFab, false))
		}
		prod.Filho(pcn.NodeStr("cBenef", p.CBenef, false))

		for _, c := range p.CredPresumido {
			prod.Filho(pcn.NovoElem("gCred").
				Filho(pcn.NodeStr("cCredPresumido", c.CCredPresumido, true)).
				Filho(w.nodeAliquota("pCredPresumido", c.PCredPresumido, true)).
				Filho(pcn.NodeDec("vCredPresumido", c.VCredPresumido, 2, true)))
		}
	} else {
		// Antes da 4.00 o CEST e opcional e vem sozinho.
		prod.Filho(pcn.NodeStr("CEST", pcn.OnlyNumber(p.CEST), false))
	}

	prod.Filho(pcn.NodeStr("tpCredPresIBSZFM", p.TpCredPresIBSZFM.String(), false)).
		Filho(pcn.NodeStr("EXTIPI", p.EXTIPI, false)).
		// tcEsp: so Trim, sem o filtro de texto -- aqui nao muda nada porque
		// o OnlyNumber ja passou antes.
		Filho(pcn.NodeStrSemFiltro("CFOP", pcn.OnlyNumber(p.CFOP), true)).
		Filho(pcn.NodeStr("uCom", p.UCom, true)).
		Filho(pcn.NodeDec("qCom", p.QCom, 4, true)).
		Filho(w.nodeValorUnitario("vUnCom", p.VUnCom)).
		Filho(pcn.NodeDec("vProd", p.VProd, 2, true))

	if w.versao() >= 4 && strings.TrimSpace(p.CEANTrib) == "" && w.depoisDaObrigatoriedadeGTIN() {
		p.CEANTrib = semGTIN
	}
	prod.Filho(pcn.NodeStr("cEANTrib", p.CEANTrib, true)).
		Filho(pcn.NodeStr("cBarraTrib", p.CBarraTrib, false)).
		Filho(pcn.NodeStr("uTrib", p.UTrib, true)).
		Filho(pcn.NodeDec("qTrib", p.QTrib, 4, true)).
		Filho(w.nodeValorUnitario("vUnTrib", p.VUnTrib)).
		Filho(pcn.NodeDec("vFrete", p.VFrete, 2, false)).
		Filho(pcn.NodeDec("vSeg", p.VSeg, 2, false)).
		Filho(pcn.NodeDec("vDesc", p.VDesc, 2, false)).
		Filho(pcn.NodeDec("vOutro", p.VOutro, 2, false))

	if w.versao() >= 2 {
		prod.Filho(pcn.NodeStr("indTot", p.IndTot.String(), true))
	}

	// indBemMovelUsado so sai quando vale SIM: o "nao" fica implicito
	// (XmlWriter.pas:1273).
	if w.versao() >= 4 && p.IndBemMovelUsado == pcn.TieSim {
		prod.Filho(pcn.NodeStr("indBemMovelUsado", p.IndBemMovelUsado.String(), true))
	}

	for _, di := range p.DI {
		prod.Filho(w.gerarDetProdDI(di))
	}
	for _, de := range p.DetExport {
		prod.Filho(gerarDetProdDetExport(de))
	}

	prod.Filho(pcn.NodeStr("xPed", p.XPed, false)).
		Filho(pcn.NodeStr("nItemPed", pcn.OnlyNumber(p.NItemPed), false)).
		Filho(pcn.NodeStr("nFCI", p.NFCI, false))

	// rastro substituiu os campos de lote do med na 4.00.
	if w.versao() >= 4 {
		for _, r := range p.Rastro {
			prod.Filho(gerarDetProdRastro(r))
		}
	}

	prod.Filho(w.gerarDetProdVeicProd(p.VeicProd))

	for _, m := range p.Med {
		prod.Filho(w.gerarDetProdMed(m))
	}
	for _, a := range p.Arma {
		prod.Filho(gerarDetProdArma(a))
	}

	prod.Filho(w.gerarDetProdComb(p.Comb)).
		Filho(pcn.NodeStr("nRECOPI", p.NRECOPI, false))

	return prod
}

// depoisDaObrigatoriedadeGTIN compara SO A DATA de emissao com 01/12/2018,
// reproduzindo o CompareDate do ACBr.
func (w *xmlWriter) depoisDaObrigatoriedadeGTIN() bool {
	d := w.n.Ide.DEmi
	emissao := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return emissao.After(inicioObrigatoriedadeGTIN)
}

// nodeAliquota grava um percentual com a precisao do FormatoValor4ou2: quatro
// casas a partir da 3.10, duas antes.
func (w *xmlWriter) nodeAliquota(nome string, valor float64, obrigatorio bool) *pcn.Elem {
	return pcn.NodeDec(nome, valor, w.casasValor4ou2, obrigatorio)
}

// nodeValorUnitario grava vUnCom/vUnTrib com a precisao do FormatoValor10ou4:
// dez casas a partir da 2.00, quatro antes. Sempre obrigatorio.
func (w *xmlWriter) nodeValorUnitario(nome string, valor float64) *pcn.Elem {
	return pcn.NodeDec(nome, valor, w.casasValor10ou4, true)
}

// ---------------------------------------------------------------------------
// Produto: grupos por tipo de mercadoria
// ---------------------------------------------------------------------------

// gerarDetProdDI monta uma Declaracao de Importacao. Porte de GerarDetProdDI
// (XmlWriter.pas:1326).
func (w *xmlWriter) gerarDetProdDI(di DI) *pcn.Elem {
	e := pcn.NovoElem("DI").
		Filho(pcn.NodeStr("nDI", di.NDi, true)).
		Filho(pcn.NodeDat("dDI", di.DDi, true)).
		Filho(pcn.NodeStr("xLocDesemb", di.XLocDesemb, true)).
		Filho(pcn.NodeStr("UFDesemb", di.UFDesemb, true)).
		Filho(pcn.NodeDat("dDesemb", di.DDesemb, true))

	if w.versao() >= 3 {
		e.Filho(pcn.NodeStr("tpViaTransp", di.TpViaTransp.String(), true))

		// vAFRMM so e obrigatorio na via maritima -- e la ele sai mesmo
		// zerado (XmlWriter.pas:1360).
		e.Filho(pcn.NodeDec("vAFRMM", di.VAFRMM, 2, di.TpViaTransp == TvMaritima))

		e.Filho(pcn.NodeStr("tpIntermedio", di.TpIntermedio.String(), true)).
			Filho(gerarCNPJCPF(di.CNPJ, false, true)).
			Filho(pcn.NodeStr("UFTerceiro", di.UFTerceiro, false))
	}

	e.Filho(pcn.NodeStr("cExportador", di.CExportador, true))

	for _, a := range di.Adi {
		e.Filho(pcn.NovoElem("adi").
			Filho(pcn.NodeInt("nAdicao", a.NAdicao, 1, false)).
			Filho(pcn.NodeInt("nSeqAdic", a.NSeqAdi, 1, true)).
			Filho(pcn.NodeStr("cFabricante", a.CFabricante, true)).
			Filho(pcn.NodeDec("vDescDI", a.VDescDI, 2, false)).
			Filho(pcn.NodeStr("nDraw", a.NDraw, false)))
	}

	return e
}

// gerarDetProdDetExport monta o detalhamento de exportacao. Porte de
// GerarDetProddetExport (XmlWriter.pas:1446).
//
// O subgrupo exportInd so existe quando nRE ou chNFe estao preenchidos; sem
// eles, o detExport carrega apenas o nDraw.
func gerarDetProdDetExport(d DetExport) *pcn.Elem {
	e := pcn.NovoElem("detExport").
		Filho(pcn.NodeStr("nDraw", d.NDraw, false))

	if d.NRE != "" || d.ChNFe != "" {
		e.Filho(pcn.NovoElem("exportInd").
			Filho(pcn.NodeStr("nRE", d.NRE, true)).
			Filho(pcn.NodeStrSemFiltro("chNFe", pcn.RemoverLiteralChave(d.ChNFe), true)).
			Filho(pcn.NodeDec("qExport", d.QExport, 4, true)))
	}

	return e
}

// gerarDetProdRastro monta o grupo de rastreabilidade. Porte de
// GerarDetProdRastro (XmlWriter.pas:1489).
func gerarDetProdRastro(r Rastro) *pcn.Elem {
	return pcn.NovoElem("rastro").
		Filho(pcn.NodeStr("nLote", r.NLote, true)).
		Filho(pcn.NodeDec("qLote", r.QLote, 3, true)).
		Filho(pcn.NodeDat("dFab", r.DFab, true)).
		Filho(pcn.NodeDat("dVal", r.DVal, true)).
		Filho(pcn.NodeStr("cAgreg", r.CAgreg, false))
}

// gerarDetProdVeicProd monta o grupo de veiculo novo. Porte de
// GerarDetProdVeicProd (XmlWriter.pas:1514).
//
// O CHASSI e a chave do grupo: sem ele o veicProd nao sai, mesmo que os
// demais campos estejam preenchidos.
//
// Duas tags MUDARAM DE NOME na 2.00 e carregam o mesmo campo da classe:
// CM3 virou cilin e CMKG virou CMT.
func (w *xmlWriter) gerarDetProdVeicProd(v VeicProd) *pcn.Elem {
	if strings.TrimSpace(v.Chassi) == "" {
		return nil
	}

	e := pcn.NovoElem("veicProd").
		Filho(pcn.NodeStr("tpOp", v.TpOP.String(), true)).
		Filho(pcn.NodeStr("chassi", v.Chassi, true)).
		Filho(pcn.NodeStr("cCor", v.CCor, true)).
		Filho(pcn.NodeStr("xCor", v.XCor, true)).
		Filho(pcn.NodeStr("pot", v.Pot, true))

	if w.versao() >= 2 {
		e.Filho(pcn.NodeStr("cilin", v.Cilin, true))
	} else {
		e.Filho(pcn.NodeStr("CM3", v.Cilin, true))
	}

	e.Filho(pcn.NodeStr("pesoL", v.PesoL, true)).
		Filho(pcn.NodeStr("pesoB", v.PesoB, true)).
		Filho(pcn.NodeStr("nSerie", v.NSerie, true)).
		Filho(pcn.NodeStr("tpComb", v.TpComb, true)).
		Filho(pcn.NodeStr("nMotor", v.NMotor, true))

	if w.versao() >= 2 {
		e.Filho(pcn.NodeStr("CMT", v.CMT, true))
	} else {
		e.Filho(pcn.NodeStr("CMKG", v.CMT, true))
	}

	e.Filho(pcn.NodeStr("dist", v.Dist, true)).
		Filho(pcn.NodeInt("anoMod", v.AnoMod, 0, true)).
		Filho(pcn.NodeInt("anoFab", v.AnoFab, 0, true)).
		Filho(pcn.NodeStr("tpPint", v.TpPint, true)).
		Filho(pcn.NodeInt("tpVeic", v.TpVeic, 0, true)).
		Filho(pcn.NodeInt("espVeic", v.EspVeic, 0, true)).
		Filho(pcn.NodeStr("VIN", v.VIN, true)).
		Filho(pcn.NodeStr("condVeic", v.CondVeic.String(), true)).
		Filho(pcn.NodeStr("cMod", v.CMod, true))

	if w.versao() >= 2 {
		e.Filho(pcn.NodeStr("cCorDENATRAN", v.CCorDENATRAN, true)).
			Filho(pcn.NodeInt("lota", v.Lota, 1, true)).
			Filho(pcn.NodeInt("tpRest", v.TpRest, 1, true))
	}

	return e
}

// gerarDetProdMed monta o grupo de medicamento. Porte de GerarDetProdMed
// (XmlWriter.pas:1583).
//
// Os dois conjuntos de campos sao EXCLUDENTES por versao: a 4.00 trocou o
// lote (que virou o grupo rastro) pelo codigo ANVISA.
func (w *xmlWriter) gerarDetProdMed(m Med) *pcn.Elem {
	e := pcn.NovoElem("med")

	if w.versao() >= 4 {
		e.Filho(pcn.NodeStr("cProdANVISA", m.CProdANVISA, true)).
			Filho(pcn.NodeStr("xMotivoIsencao", m.XMotivoIsencao, false))
	} else {
		e.Filho(pcn.NodeStr("nLote", m.NLote, true)).
			Filho(pcn.NodeDec("qLote", m.QLote, 3, true)).
			Filho(pcn.NodeDat("dFab", m.DFab, true)).
			Filho(pcn.NodeDat("dVal", m.DVal, true))
	}

	return e.Filho(pcn.NodeDec("vPMC", m.VPMC, 2, true))
}

// gerarDetProdArma monta o grupo de armamento. Porte de GerarDetProdArma
// (XmlWriter.pas:1626).
func gerarDetProdArma(a Arma) *pcn.Elem {
	return pcn.NovoElem("arma").
		Filho(pcn.NodeStr("tpArma", a.TpArma.String(), true)).
		Filho(pcn.NodeStr("nSerie", a.NSerie, true)).
		Filho(pcn.NodeStr("nCano", a.NCano, true)).
		Filho(pcn.NodeStr("descr", a.Descr, true))
}

// cProdANPGLP e o codigo ANP do GLP, o unico que abre o subgrupo de mistura
// (pGLP/pGNn/pGNi/vPart) no leiaute 4.00.
const cProdANPGLP = 210203001

// gerarDetProdComb monta o grupo de combustivel. Porte de GerarDetProdComb
// (XmlWriter.pas:1648).
//
// O cProdANP e a chave do grupo: zero significa "nao e combustivel".
//
// A bifurcacao de baixo e por VERSAO DISFARCADA DE VALOR: o ramo dos quatro
// subgrupos de ICMS so roda antes da 2.00 E com algum daqueles valores
// preenchido. Da 2.00 em diante o leiaute trocou aquilo por UFCons + CIDE +
// encerrante, e o comentario do proprio ACBr ("Manter Disponivel para versao
// 3.0") deixou o codigo morto no lugar.
func (w *xmlWriter) gerarDetProdComb(c Comb) *pcn.Elem {
	if c.CProdANP <= 0 {
		return nil
	}

	e := pcn.NovoElem("comb").
		Filho(pcn.NodeInt("cProdANP", c.CProdANP, 9, true))

	if w.versao() < 4 {
		e.Filho(pcn.NodeDec("pMixGN", c.PMixGN, 4, false))
	} else {
		e.Filho(pcn.NodeStr("descANP", c.DescANP, true))

		if c.CProdANP == cProdANPGLP {
			// Os tres percentuais caem para DUAS casas quando valem 100 --
			// e ai tambem deixam de ser obrigatorios. E o jeito do ACBr de
			// gravar "100.00" em vez de "100.0000" (XmlWriter.pas:1670).
			e.Filho(nodePercentualGLP("pGLP", c.PGLP)).
				Filho(nodePercentualGLP("pGNn", c.PGNn)).
				Filho(nodePercentualGLP("pGNi", c.PGNi)).
				Filho(pcn.NodeDec("vPart", c.VPart, 2, false))
		}
	}

	// tcEsp: so Trim, sem o filtro de texto do tcStr.
	if strings.TrimSpace(c.CODIF) != "" {
		e.Filho(pcn.NodeStrSemFiltro("CODIF", c.CODIF, true))
	}
	if c.QTemp != 0 {
		e.Filho(pcn.NodeDec("qTemp", c.QTemp, 4, true))
	}

	if w.versao() < 2 && (c.ICMS.VBCICMS > 0 || c.ICMS.VICMS > 0 ||
		c.ICMS.VBCICMSST > 0 || c.ICMS.VICMSST > 0 ||
		c.ICMSInter.VBCICMSSTDest > 0 || c.ICMSInter.VICMSSTDest > 0 ||
		c.ICMSCons.VBCICMSSTCons > 0 || c.ICMSCons.VICMSSTCons > 0) {
		e.Filho(gerarDetProdCombCIDE(c.CIDE)).
			Filho(gerarDetProdCombICMS(c.ICMS)).
			Filho(gerarDetProdCombICMSInter(c.ICMSInter)).
			Filho(gerarDetProdCombICMSCons(c.ICMSCons))
		return e
	}

	e.Filho(pcn.NodeStr("UFCons", c.UFCons, true)).
		Filho(gerarDetProdCombCIDE(c.CIDE))

	if c.Encerrante.NBico > 0 {
		e.Filho(pcn.NovoElem("encerrante").
			Filho(pcn.NodeInt("nBico", c.Encerrante.NBico, 1, true)).
			Filho(pcn.NodeInt("nBomba", c.Encerrante.NBomba, 1, false)).
			Filho(pcn.NodeInt("nTanque", c.Encerrante.NTanque, 1, true)).
			Filho(pcn.NodeDec("vEncIni", c.Encerrante.VEncIni, 3, true)).
			Filho(pcn.NodeDec("vEncFin", c.Encerrante.VEncFin, 3, true)))
	}

	e.Filho(pcn.NodeDec("pBio", c.PBio, 4, false))

	for _, o := range c.OrigComb {
		e.Filho(pcn.NovoElem("origComb").
			Filho(pcn.NodeStr("indImport", o.IndImport.String(), true)).
			Filho(pcn.NodeInt("cUFOrig", o.CUFOrig, 2, true)).
			Filho(pcn.NodeDec("pOrig", o.POrig, 4, true)))
	}

	return e
}

// nodePercentualGLP grava pGLP/pGNn/pGNi com a regra do ACBr: 100 sai com
// duas casas e opcional; qualquer outro valor sai com quatro e obrigatorio.
func nodePercentualGLP(nome string, valor float64) *pcn.Elem {
	if valor == 100 {
		return pcn.NodeDec(nome, valor, 2, false)
	}
	return pcn.NodeDec(nome, valor, 4, true)
}

// gerarDetProdCombCIDE monta a CIDE do combustivel. Porte de
// GerarDetProdCombCIDE (XmlWriter.pas:1744).
func gerarDetProdCombCIDE(c CIDE) *pcn.Elem {
	if c.QBCProd <= 0 && c.VAliqProd <= 0 && c.VCIDE <= 0 {
		return nil
	}
	return pcn.NovoElem("CIDE").
		Filho(pcn.NodeDec("qBCProd", c.QBCProd, 4, true)).
		Filho(pcn.NodeDec("vAliqProd", c.VAliqProd, 4, true)).
		Filho(pcn.NodeDec("vCIDE", c.VCIDE, 2, true))
}

// gerarDetProdCombICMS monta o ICMSComb. Porte de GerarDetProdCombICMS
// (XmlWriter.pas:1801). So alcancavel antes da 2.00.
func gerarDetProdCombICMS(c ICMSComb) *pcn.Elem {
	return pcn.NovoElem("ICMSComb").
		Filho(pcn.NodeDec("vBCICMS", c.VBCICMS, 2, true)).
		Filho(pcn.NodeDec("vICMS", c.VICMS, 2, true)).
		Filho(pcn.NodeDec("vBCICMSST", c.VBCICMSST, 2, true)).
		Filho(pcn.NodeDec("vICMSST", c.VICMSST, 2, true))
}

// gerarDetProdCombICMSInter monta o ICMSInter. Porte de
// GerarDetProdCombICMSInter (XmlWriter.pas:1814).
func gerarDetProdCombICMSInter(c ICMSInter) *pcn.Elem {
	if c.VBCICMSSTDest <= 0 && c.VICMSSTDest <= 0 {
		return nil
	}
	return pcn.NovoElem("ICMSInter").
		Filho(pcn.NodeDec("vBCICMSSTDest", c.VBCICMSSTDest, 2, true)).
		Filho(pcn.NodeDec("vICMSSTDest", c.VICMSSTDest, 2, true))
}

// gerarDetProdCombICMSCons monta o ICMSCons. Porte de
// GerarDetProdCombICMSCons (XmlWriter.pas:1828).
func gerarDetProdCombICMSCons(c ICMSCons) *pcn.Elem {
	if c.VBCICMSSTCons <= 0 && c.VICMSSTCons <= 0 && strings.TrimSpace(c.UFCons) == "" {
		return nil
	}
	return pcn.NovoElem("ICMSCons").
		Filho(pcn.NodeDec("vBCICMSSTCons", c.VBCICMSSTCons, 2, true)).
		Filho(pcn.NodeDec("vICMSSTCons", c.VICMSSTCons, 2, true)).
		Filho(pcn.NodeStr("UFCons", c.UFCons, true))
}

// ---------------------------------------------------------------------------
// Impostos do item
// ---------------------------------------------------------------------------

// gerarDetImposto monta o grupo imposto. Porte de GerarDetImposto
// (XmlWriter.pas:1847).
//
// A condicional de fora e a NT 2025.002 (rejeicao 1001): nota com finalidade
// de credito ou debito -- ou com recebimento de pagamento de compra
// governamental -- NAO detalha os impostos antigos no item, porque e um
// documento so de IBS/CBS. Cinco codigos escapam da regra, e a condicional
// interna deixa de fora mais um (pagamento antecipado).
//
// Os dois grupos da Reforma Tributaria ficam FORA de tudo isso: saem sempre.
func (w *xmlWriter) gerarDetImposto(i int) *pcn.Elem {
	n := w.n
	d := &n.Det[i]

	imposto := pcn.NovoElem("imposto").
		Filho(pcn.NodeDec("vTotTrib", d.Imposto.VTotTrib, 2, false))

	if w.detalhaImpostosAntigos() {
		if w.detalhaImpostosDeMercadoria() {
			// ISSQN e ICMS sao EXCLUDENTES: o item e servico ou mercadoria.
			// O criterio muda na 4.00, que passou a aceitar o cListServ
			// sozinho como marca de servico.
			if d.Imposto.ISSQN.CSitTrib != ISSQNcSitTribVazio ||
				(w.versao() > 3 && d.Imposto.ISSQN.CListServ != "") {
				if w.versao() >= 3 {
					imposto.Filho(w.gerarDetImpostoIPI(i))
				}
				imposto.Filho(w.gerarDetImpostoISSQN(i))
			} else {
				imposto.Filho(w.gerarDetImpostoICMS(i)).
					Filho(w.gerarDetImpostoIPI(i)).
					Filho(w.gerarDetImpostoII(i))
			}
		}

		imposto.Filho(w.gerarDetImpostoPIS(i)).
			Filho(w.gerarDetImpostoPISST(i)).
			Filho(w.gerarDetImpostoCOFINS(i)).
			Filho(w.gerarDetImpostoCOFINSST(i))

		// A partilha da EC 87/2015 so sai com percentual de partilha.
		if d.Imposto.ICMSUFDest.PICMSInterPart > 0 {
			imposto.Filho(w.gerarDetImpostoICMSUFDest(i))
		}
	}

	// Reforma Tributaria.
	if d.Imposto.ISel.CSTIS != "" && d.Imposto.ISel.VBCIS > 0 {
		imposto.Filho(w.rtc.GerarISel(d.Imposto.ISel))
	}
	imposto.Filho(w.rtc.GerarIBSCBS(d.Imposto.IBSCBS))

	return imposto
}

// detalhaImpostosAntigos e a condicional externa da NT 2025.002 v1.40
// (XmlWriter.pas:1868): nota de credito/debito puro de IBS/CBS nao detalha
// ICMS, IPI, II, PIS, COFINS nem a partilha no item.
func (w *xmlWriter) detalhaImpostosAntigos() bool {
	ide := w.n.Ide
	naoEhCreditoDebito := ide.FinNFe != FnCredito && ide.FinNFe != FnDebito &&
		ide.GCompraGov.TpOperGov != rtc.TogRecebimentoPag

	return naoEhCreditoDebito ||
		ide.TpNFCredito == TcRetorno ||
		ide.TpNFCredito == TcReducaoValores ||
		ide.TpNFCredito == TcRetornoRecusaParcial ||
		ide.TpNFDebito == TdPerdaEmEstoque
}

// detalhaImpostosDeMercadoria e a condicional interna da mesma NT
// (XmlWriter.pas:1873): o pagamento antecipado detalha PIS/COFINS, mas nao
// ICMS/IPI/II/ISSQN.
func (w *xmlWriter) detalhaImpostosDeMercadoria() bool {
	ide := w.n.Ide
	naoEhCreditoDebito := ide.FinNFe != FnCredito && ide.FinNFe != FnDebito &&
		ide.GCompraGov.TpOperGov != rtc.TogRecebimentoPag

	return naoEhCreditoDebito || ide.TpNFDebito != TdPagamentoAntecipado
}

// sufixoTagICMS mapeia o CST no SUFIXO da tag do grupo de ICMS -- o
// BuscaTag do GerarDetImpostoICMS (XmlWriter.pas:1916).
//
// O mapa NAO e a identidade: cinco CSTs distintos (40, 41 e 50) caem em
// "40"; as partilhas caem em "Part"; o repasse do 41 cai em "ST"; e o
// repasse do 60 DEPENDE DA VERSAO -- era "60" ate a 3.10 e virou "ST" na
// 4.00.
//
// CST sem entrada no mapa devolve vazio, o que gera a tag "ICMS" dentro do
// grupo ICMS -- um "<ICMS><ICMS>" aninhado. E o que o original faz: o case do
// Delphi nao tem else, e o Result de uma string fica vazio.
//
// O caso alcancavel disso e o cstPart20: ele EXISTE no enum
// (ACBrDFe.Conversao.pas:259) e e tratado no case do writer
// (XmlWriter.pas:2201), mas foi esquecido no BuscaTag (:1916-1941).
// Replicado, com teste.
func (w *xmlWriter) sufixoTagICMS(cst pcn.CSTIcms) string {
	switch cst {
	case pcn.CST00:
		return "00"
	case pcn.CST02:
		return "02"
	case pcn.CST10:
		return "10"
	case pcn.CST15:
		return "15"
	case pcn.CST20:
		return "20"
	case pcn.CST30:
		return "30"
	case pcn.CST40, pcn.CST41, pcn.CST50:
		return "40"
	case pcn.CST51:
		return "51"
	case pcn.CST53:
		return "53"
	case pcn.CST60:
		return "60"
	case pcn.CST61:
		return "61"
	case pcn.CST70:
		return "70"
	case pcn.CST80:
		return "80"
	case pcn.CST81:
		return "81"
	case pcn.CST90:
		return "90"
	case pcn.CSTPart10, pcn.CSTPart90:
		return "Part"
	case pcn.CSTRep41:
		return "ST"
	case pcn.CSTRep60:
		if w.versao() < 4 {
			return "60"
		}
		return "ST"
	default:
		return ""
	}
}

// sufixoTagICMSSN mapeia o CSOSN no sufixo da tag ICMSSN -- o
// CSOSNToStrTagPos do ACBr (ACBrDFe.Conversao.pas:1282).
//
// Quatro CSOSN compartilham a tag ICMSSN102 e dois compartilham a ICMSSN202:
// o leiaute reaproveita o mesmo grupo para codigos de mesma estrutura.
func sufixoTagICMSSN(csosn pcn.CSOSNIcms) string {
	switch csosn {
	case pcn.Csosn101:
		return "101"
	case pcn.Csosn102, pcn.Csosn103, pcn.Csosn300, pcn.Csosn400:
		return "102"
	case pcn.Csosn201:
		return "201"
	case pcn.Csosn202, pcn.Csosn203:
		return "202"
	case pcn.Csosn500:
		return "500"
	case pcn.Csosn900:
		return "900"
	default:
		return ""
	}
}

// gerarDetImpostoICMS monta o grupo ICMS do item. Porte de
// GerarDetImpostoICMS (XmlWriter.pas:1910), o maior metodo do writer.
//
// A estrutura tem tres niveis de decisao:
//
//  1. Os quatro CSTs de MONOFASIA (02, 15, 53, 61) desviam antes de tudo --
//     nao olham o regime do emitente, porque monofasia vale para qualquer um.
//  2. Fora deles, o REGIME decide: regime normal e excesso de receita usam
//     CST; Simples Nacional e MEI usam CSOSN.
//  3. Dentro de cada um, o codigo escolhe quais campos saem.
func (w *xmlWriter) gerarDetImpostoICMS(i int) *pcn.Elem {
	icms := w.n.Det[i].Imposto.ICMS
	grupo := pcn.NovoElem("ICMS")

	switch icms.CST {
	case pcn.CST02, pcn.CST15, pcn.CST53, pcn.CST61:
		grupo.Filho(w.gerarICMSMonofasico(icms))
		return grupo
	}

	switch w.n.Emit.CRT {
	case CrtRegimeNormal, CrtSimplesExcessoReceita:
		grupo.Filho(w.gerarICMSRegimeNormal(icms))
	case CrtSimplesNacional, CrtMEI:
		grupo.Filho(w.gerarICMSSimplesNacional(icms))
	}

	return grupo
}

// gerarICMSMonofasico monta os grupos ICMS02, ICMS15, ICMS53 e ICMS61.
func (w *xmlWriter) gerarICMSMonofasico(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS" + w.sufixoTagICMS(icms.CST)).
		Filho(pcn.NodeStr("orig", icms.Orig.String(), true)).
		Filho(pcn.NodeStr("CST", icms.CST.String(), true))

	switch icms.CST {
	case pcn.CST02:
		e.Filho(pcn.NodeDec("qBCMono", icms.QBCMono, 4, true)).
			Filho(pcn.NodeDec("adRemICMS", icms.AdRemICMS, 4, true)).
			Filho(pcn.NodeDec("vICMSMono", icms.VICMSMono, 2, true))

	case pcn.CST15:
		e.Filho(pcn.NodeDec("qBCMono", icms.QBCMono, 4, true)).
			Filho(pcn.NodeDec("adRemICMS", icms.AdRemICMS, 4, true)).
			Filho(pcn.NodeDec("vICMSMono", icms.VICMSMono, 2, true)).
			Filho(pcn.NodeDec("qBCMonoReten", icms.QBCMonoReten, 4, true)).
			Filho(pcn.NodeDec("adRemICMSReten", icms.AdRemICMSReten, 4, true)).
			Filho(pcn.NodeDec("vICMSMonoReten", icms.VICMSMonoReten, 2, true))

		// A reducao da aliquota ad rem so e declarada quando ha base
		// monofasica -- e os dois campos andam juntos.
		if icms.QBCMono != 0 {
			e.Filho(pcn.NodeDec("pRedAdRem", icms.PRedAdRem, 2, true)).
				Filho(pcn.NodeStr("motRedAdRem", icms.MotRedAdRem.String(), true))
		}

	case pcn.CST53:
		e.Filho(pcn.NodeDec("qBCMono", icms.QBCMono, 4, true)).
			Filho(pcn.NodeDec("adRemICMS", icms.AdRemICMS, 4, false)).
			Filho(pcn.NodeDec("vICMSMonoOp", icms.VICMSMonoOp, 2, false))

		// Diferimento integral sai com duas casas; parcial, com quatro.
		casasPDif := 4
		if icms.PDif == 100 {
			casasPDif = 2
		}
		e.Filho(pcn.NodeDec("pDif", icms.PDif, casasPDif, false)).
			Filho(pcn.NodeDec("vICMSMonoDif", icms.VICMSMonoDif, 2, false)).
			Filho(pcn.NodeDec("vICMSMono", icms.VICMSMono, 2, false)).
			Filho(pcn.NodeDec("qBCMonoDif", icms.QBCMonoDif, 4, false)).
			Filho(pcn.NodeDec("adRemICMSDif", icms.AdRemICMSDif, 4, false))

	case pcn.CST61:
		e.Filho(pcn.NodeDec("qBCMonoRet", icms.QBCMonoRet, 4, true)).
			Filho(pcn.NodeDec("adRemICMSRet", icms.AdRemICMSRet, 4, true)).
			Filho(pcn.NodeDec("vICMSMonoRet", icms.VICMSMonoRet, 2, true))
	}

	return e
}

// forcaTagRejeicao906 reproduz o OcorrenciasICMSEfetivo (XmlWriter.pas:1957):
// algumas UFs exigem o grupo de ICMS efetivo mesmo zerado, e so em venda a
// consumidor final.
func (w *xmlWriter) forcaTagRejeicao906() bool {
	return w.n.Ide.IndFinal == CfConsumidorFinal &&
		forcaTagNoAmbiente(w.opcoes.ForcarGerarTagRejeicao906, w.n.Ide.TpAmb)
}

// forcaTagRejeicao938 reproduz o OcorrenciasVICMSSubstituto
// (XmlWriter.pas:1943): algumas UFs exigem o vICMSSubstituto mesmo zerado.
func (w *xmlWriter) forcaTagRejeicao938() bool {
	return forcaTagNoAmbiente(w.opcoes.ForcarGerarTagRejeicao938, w.n.Ide.TpAmb)
}

// forcaTagNoAmbiente resolve um ForcarGeracaoTag contra o ambiente da nota.
func forcaTagNoAmbiente(forcar ForcarGeracaoTag, ambiente pcn.TipoAmbiente) bool {
	switch forcar {
	case FgtSempre:
		return true
	case FgtSomenteProducao:
		return ambiente == pcn.TaProducao
	case FgtSomenteHomologacao:
		return ambiente == pcn.TaHomologacao
	default:
		return false
	}
}

// ufsQueExigemICMS51Zerado sao RJ, SP e PR -- as tres que pedem o grupo do
// CST 51 preenchido ainda que com zeros (XmlWriter.pas:2337).
var ufsQueExigemICMS51Zerado = map[int]bool{33: true, 35: true, 41: true}

// gerarICMSRegimeNormal monta o grupo de ICMS do emitente em regime normal
// (ou Simples com excesso de receita).
func (w *xmlWriter) gerarICMSRegimeNormal(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS" + w.sufixoTagICMS(icms.CST)).
		Filho(pcn.NodeStr("orig", icms.Orig.String(), true)).
		Filho(pcn.NodeStr("CST", icms.CST.String(), true))

	switch icms.CST {
	case pcn.CST00:
		e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
			Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
			Filho(w.nodeAliquota("pICMS", icms.PICMS, true)).
			Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))

		if w.versao() >= 4 && (icms.PFCP > 0 || icms.VFCP > 0) {
			e.Filho(w.nodeAliquota("pFCP", icms.PFCP, true)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, true))
		}

	case pcn.CST10, pcn.CSTPart10:
		// pRedBC, pBCOp e UFST so aparecem na PARTILHA -- o CST 10 puro so
		// os emite quando o usuario preencheu a UF ou o percentual.
		temPartilha := icms.UFST != "" || icms.PBCOp != 0 || icms.CST == pcn.CSTPart10

		e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
			Filho(pcn.NodeDec("vBC", icms.VBC, 2, true))
		if temPartilha {
			e.Filho(pcn.NodeDec("pRedBC", icms.PRedBC, 2, false))
		}
		e.Filho(w.nodeAliquota("pICMS", icms.PICMS, true)).
			Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))

		if w.versao() >= 4 && (icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0) {
			e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, true)).
				Filho(w.nodeAliquota("pFCP", icms.PFCP, true)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, true))
		}

		e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
			Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
			Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
			Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
			Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
			Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

		// pBCOp vem ANTES de UFST, ao contrario do que os IDs do leiaute
		// (N24 = UFST, N25 = pBCOp) sugerem. A ordem do XSD e a do codigo.
		if temPartilha {
			e.Filho(w.nodeAliquota("pBCOp", icms.PBCOp, true)).
				Filho(pcn.NodeStr("UFST", icms.UFST, true))
		}

		if w.versao() >= 4 && icms.VICMSSTDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSSTDeson", icms.VICMSSTDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMSST", icms.MotDesICMSST.String(), true))
		}

		// A desoneracao do ICMS proprio no grupo 10 e exclusiva da partilha.
		if w.versao() >= 4 && icms.VICMSDeson > 0 && icms.CST == pcn.CSTPart10 {
			e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true)).
				Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
		}

	case pcn.CST20, pcn.CSTPart20:
		e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
			Filho(w.nodeAliquota("pRedBC", icms.PRedBC, true)).
			Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
			Filho(w.nodeAliquota("pICMS", icms.PICMS, true)).
			Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))

		if w.versao() >= 4 && (icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0) {
			e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, true)).
				Filho(w.nodeAliquota("pFCP", icms.PFCP, true)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, true))
		}

		// O CST 20 ganhou desoneracao na 3.10; a partilha dele, so na 4.00.
		desoneraCST20 := w.versao() >= 3.10 && icms.CST == pcn.CST20
		desoneraPart20 := w.versao() >= 4 && icms.CST == pcn.CSTPart20
		if (desoneraCST20 || desoneraPart20) && icms.VICMSDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true))
			if w.versao() >= 4 {
				e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
			}
		}

	case pcn.CST30:
		e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
			Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
			Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
			Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
			Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
			Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

		if w.versao() >= 3.10 && icms.VICMSDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true))
			if w.versao() >= 4 {
				e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
			}
		}

	case pcn.CST40, pcn.CST41, pcn.CST50:
		// Ate a 3.00 o grupo da isencao carregava o vICMS; da 3.10 em diante
		// virou o par vICMSDeson/motDesICMS, com IDs DIFERENTES (N13a/N13b,
		// nao N27a/N28 como nos demais grupos).
		if w.versao() >= 3.10 {
			if icms.VICMSDeson > 0 {
				e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
					Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true))
				if w.versao() >= 4 {
					e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
				}
			}
		} else {
			e.Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, false))
			if icms.VICMS > 0 {
				e.Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), false))
			}
		}

	case pcn.CST51:
		// RJ, SP e PR exigem o grupo inteiro mesmo zerado; nas demais UFs os
		// campos sao opcionais e somem quando valem zero.
		exigido := ufsQueExigemICMS51Zerado[w.n.Ide.CUF]

		if icms.ModBC != DbiNenhum {
			e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), exigido))
		}
		e.Filho(w.nodeAliquota("pRedBC", icms.PRedBC, exigido)).
			Filho(pcn.NodeStr("cBenefRBC", icms.CBenefRBC, false)).
			Filho(pcn.NodeDec("vBC", icms.VBC, 2, exigido)).
			Filho(w.nodeAliquota("pICMS", icms.PICMS, exigido)).
			Filho(pcn.NodeDec("vICMSOp", icms.VICMSOp, 2, exigido)).
			Filho(w.nodeAliquota("pDif", icms.PDif, exigido)).
			Filho(pcn.NodeDec("vICMSDif", icms.VICMSDif, 2, exigido)).
			Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, exigido))

		if w.versao() >= 4 && (icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0) {
			e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, true)).
				Filho(w.nodeAliquota("pFCP", icms.PFCP, true)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, true))
		}

		if w.versao() >= 4 && (icms.PFCPDif > 0 || icms.VFCPDif > 0 || icms.VFCPEfet > 0) {
			e.Filho(w.nodeAliquota("pFCPDif", icms.PFCPDif, true)).
				Filho(pcn.NodeDec("vFCPDif", icms.VFCPDif, 2, true)).
				Filho(pcn.NodeDec("vFCPEfet", icms.VFCPEfet, 2, false))
		}

	case pcn.CST60:
		if w.versao() >= 2 {
			// Os quatro campos do ST retido andam juntos: basta um deles
			// estar preenchido para todos saírem.
			if icms.VBCSTRet > 0 || icms.PST > 0 || icms.VICMSSubstituto > 0 ||
				icms.VICMSSTRet > 0 || w.forcaTagRejeicao938() {
				e.Filho(pcn.NodeDec("vBCSTRet", icms.VBCSTRet, 2, true))

				if w.versao() >= 4 {
					e.Filho(w.nodeAliquota("pST", icms.PST, true)).
						Filho(pcn.NodeDec("vICMSSubstituto", icms.VICMSSubstituto, 2,
							w.forcaTagRejeicao938()))
				}

				e.Filho(pcn.NodeDec("vICMSSTRet", icms.VICMSSTRet, 2, true))
			}

			if w.versao() >= 4 {
				if icms.VBCFCPSTRet > 0 || icms.PFCPSTRet > 0 || icms.VFCPSTRet > 0 {
					e.Filho(pcn.NodeDec("vBCFCPSTRet", icms.VBCFCPSTRet, 2, true)).
						Filho(w.nodeAliquota("pFCPSTRet", icms.PFCPSTRet, true)).
						Filho(pcn.NodeDec("vFCPSTRet", icms.VFCPSTRet, 2, true))
				}
				e.Filhos(w.gerarICMSEfetivo(icms)...)
			}
		} else {
			e.Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
				Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))
		}

	case pcn.CST70:
		e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
			Filho(w.nodeAliquota("pRedBC", icms.PRedBC, true)).
			Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
			Filho(w.nodeAliquota("pICMS", icms.PICMS, true)).
			Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))

		if w.versao() >= 4 && (icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0) {
			e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, true)).
				Filho(w.nodeAliquota("pFCP", icms.PFCP, true)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, true))
		}

		e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
			Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
			Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
			Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
			Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
			Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

		// Aqui o corte e >= 3, nao >= 3.10 como nos CSTs 30, 40 e 90.
		if w.versao() >= 3 && icms.VICMSDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true))
			if w.versao() >= 4 {
				e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
			}
		}

		if w.versao() >= 4 && icms.VICMSSTDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSSTDeson", icms.VICMSSTDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMSST", icms.MotDesICMSST.String(), true))
		}

	case pcn.CST90, pcn.CSTPart90:
		// O grupo 90 e o "outros": quatro blocos independentes, cada um com
		// sua propria condicao de existencia. Varios deles ficam de fora na
		// partilha (cstPart90), que so reaproveita a tag.
		soCST90 := w.versao() >= 4 && icms.CST == pcn.CST90

		if icms.VBC > 0 || icms.VICMS > 0 {
			e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
				Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
				Filho(w.nodeAliquota("pRedBC", icms.PRedBC, false)).
				Filho(pcn.NodeStr("cBenefRBC", icms.CBenefRBC, false)).
				Filho(w.nodeAliquota("pICMS", icms.PICMS, true))

			if soCST90 && (icms.VICMSOp > 0 || icms.PDif > 0 || icms.VICMSDif > 0) {
				e.Filho(pcn.NodeDec("vICMSOp", icms.VICMSOp, 2, true)).
					Filho(w.nodeAliquota("pDif", icms.PDif, true)).
					Filho(pcn.NodeDec("vICMSDif", icms.VICMSDif, 2, true))
			}

			e.Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))
		}

		if soCST90 && (icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0) {
			// Ao contrario dos outros grupos, aqui os tres sao OPCIONAIS.
			e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, false)).
				Filho(w.nodeAliquota("pFCP", icms.PFCP, false)).
				Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, false))
		}

		// Este bloco nao tem o recorte por CST: vale tambem para a partilha.
		if w.versao() >= 4 && (icms.PFCPDif > 0 || icms.VFCPDif > 0 || icms.VFCPEfet > 0) {
			e.Filho(w.nodeAliquota("pFCPDif", icms.PFCPDif, true)).
				Filho(pcn.NodeDec("vFCPDif", icms.VFCPDif, 2, true)).
				Filho(pcn.NodeDec("vFCPEfet", icms.VFCPEfet, 2, false))
		}

		if icms.VBCST > 0 || icms.VICMSST > 0 || icms.PICMSST > 0 {
			// O pMVAST deste grupo sai SEMPRE com duas casas -- e o unico
			// percentual do writer que ignora o FormatoValor4ou2.
			e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
				Filho(pcn.NodeDec("pMVAST", icms.PMVAST, 2, false)).
				Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
				Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
				Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
				Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))
		}

		if soCST90 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, false)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, false)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, false))
		}

		if w.versao() >= 3.10 && icms.VICMSDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), true))
			if w.versao() >= 4 {
				e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
			}
		}

		if icms.UFST != "" || icms.PBCOp != 0 || icms.CST == pcn.CSTPart90 {
			e.Filho(w.nodeAliquota("pBCOp", icms.PBCOp, true)).
				Filho(pcn.NodeStr("UFST", icms.UFST, true))
		}

		if soCST90 && icms.VICMSSTDeson > 0 {
			e.Filho(pcn.NodeDec("vICMSSTDeson", icms.VICMSSTDeson, 2, true)).
				Filho(pcn.NodeStr("motDesICMSST", icms.MotDesICMSST.String(), true))
		}

	case pcn.CSTRep41, pcn.CSTRep60:
		// ICMSST -- repasse de ST ja retido anteriormente.
		e.Filho(pcn.NodeDec("vBCSTRet", icms.VBCSTRet, 2, true))

		if w.versao() >= 4 {
			e.Filho(w.nodeAliquota("pST", icms.PST, w.forcaTagRejeicao938())).
				Filho(pcn.NodeDec("vICMSSubstituto", icms.VICMSSubstituto, 2,
					w.forcaTagRejeicao938()))
		}

		e.Filho(pcn.NodeDec("vICMSSTRet", icms.VICMSSTRet, 2, true))

		// Note o vFCPSTRet com ID N27c aqui, contra N27d no grupo 60: sao
		// tags de mesmo nome em posicoes diferentes do leiaute.
		if icms.VBCFCPSTRet > 0 || icms.PFCPSTRet > 0 || icms.VFCPSTRet > 0 {
			e.Filho(pcn.NodeDec("vBCFCPSTRet", icms.VBCFCPSTRet, 2, false)).
				Filho(w.nodeAliquota("pFCPSTRet", icms.PFCPSTRet, false)).
				Filho(pcn.NodeDec("vFCPSTRet", icms.VFCPSTRet, 2, false))
		}

		e.Filho(pcn.NodeDec("vBCSTDest", icms.VBCSTDest, 2, true)).
			Filho(pcn.NodeDec("vICMSSTDest", icms.VICMSSTDest, 2, true)).
			Filhos(w.gerarICMSEfetivo(icms)...)
	}

	return e
}

// gerarICMSEfetivo monta o bloco de ICMS efetivo (pRedBCEfet, vBCEfet,
// pICMSEfet, vICMSEfet), comum aos grupos 60, ST e SN500.
//
// Devolve nil quando nao ha nada a declarar -- e o Filhos(nil...) do
// construtor absorve.
func (w *xmlWriter) gerarICMSEfetivo(icms ICMS) []*pcn.Elem {
	if icms.PRedBCEfet <= 0 && icms.VBCEfet <= 0 && icms.PICMSEfet <= 0 &&
		icms.VICMSEfet <= 0 && !w.forcaTagRejeicao906() {
		return nil
	}
	return []*pcn.Elem{
		w.nodeAliquota("pRedBCEfet", icms.PRedBCEfet, true),
		pcn.NodeDec("vBCEfet", icms.VBCEfet, 2, true),
		w.nodeAliquota("pICMSEfet", icms.PICMSEfet, true),
		pcn.NodeDec("vICMSEfet", icms.VICMSEfet, 2, true),
	}
}

// gerarICMSSimplesNacional monta o grupo ICMSSN do emitente optante pelo
// Simples Nacional ou MEI (XmlWriter.pas:2732).
//
// Diferente do regime normal, aqui a ORIGEM e opcional em cinco dos dez
// CSOSN -- os grupos que nao detalham valor nenhum.
func (w *xmlWriter) gerarICMSSimplesNacional(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMSSN" + sufixoTagICMSSN(icms.CSOSN))

	origemOpcional := icms.CSOSN == pcn.Csosn102 || icms.CSOSN == pcn.Csosn103 ||
		icms.CSOSN == pcn.Csosn300 || icms.CSOSN == pcn.Csosn400 ||
		icms.CSOSN == pcn.Csosn900

	e.Filho(pcn.NodeStr("orig", icms.Orig.String(), !origemOpcional)).
		Filho(pcn.NodeStr("CSOSN", icms.CSOSN.String(), true))

	switch icms.CSOSN {
	case pcn.Csosn101:
		e.Filho(w.nodeAliquota("pCredSN", icms.PCredSN, true)).
			Filho(pcn.NodeDec("vCredICMSSN", icms.VCredICMSSN, 2, true))

	case pcn.Csosn102, pcn.Csosn103, pcn.Csosn300, pcn.Csosn400:
		// orig e CSOSN ja saíram acima: estes grupos nao tem mais campos.

	case pcn.Csosn201:
		e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
			Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
			Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
			Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
			Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
			Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

		e.Filho(w.nodeAliquota("pCredSN", icms.PCredSN, true)).
			Filho(pcn.NodeDec("vCredICMSSN", icms.VCredICMSSN, 2, true))

	case pcn.Csosn202, pcn.Csosn203:
		e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
			Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
			Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
			Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
			Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
			Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

	case pcn.Csosn500:
		if icms.VBCSTRet > 0 || icms.PST > 0 || icms.VICMSSubstituto > 0 ||
			icms.VICMSSTRet > 0 || w.forcaTagRejeicao938() {
			e.Filho(pcn.NodeDec("vBCSTRet", icms.VBCSTRet, 2, true))

			if w.versao() >= 4 {
				e.Filho(w.nodeAliquota("pST", icms.PST, true)).
					Filho(pcn.NodeDec("vICMSSubstituto", icms.VICMSSubstituto, 2,
						w.forcaTagRejeicao938()))
			}

			e.Filho(pcn.NodeDec("vICMSSTRet", icms.VICMSSTRet, 2, true))
		}

		if w.versao() >= 4 {
			if icms.VBCFCPSTRet > 0 || icms.PFCPSTRet > 0 || icms.VFCPSTRet > 0 {
				e.Filho(pcn.NodeDec("vBCFCPSTRet", icms.VBCFCPSTRet, 2, true)).
					Filho(w.nodeAliquota("pFCPSTRet", icms.PFCPSTRet, true)).
					Filho(pcn.NodeDec("vFCPSTRet", icms.VFCPSTRet, 2, true))
			}
			e.Filhos(w.gerarICMSEfetivo(icms)...)
		}

	case pcn.Csosn900:
		if icms.VBC > 0 || icms.VICMS > 0 {
			e.Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
				Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
				Filho(w.nodeAliquota("pRedBC", icms.PRedBC, false)).
				Filho(w.nodeAliquota("pICMS", icms.PICMS, true)).
				Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))
		}

		if icms.VBCST > 0 || icms.VICMSST > 0 {
			e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true)).
				Filho(w.nodeAliquota("pMVAST", icms.PMVAST, false)).
				Filho(w.nodeAliquota("pRedBCST", icms.PRedBCST, false)).
				Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
				Filho(w.nodeAliquota("pICMSST", icms.PICMSST, true)).
				Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))
		}

		if w.versao() >= 4 && (icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0) {
			e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, true)).
				Filho(w.nodeAliquota("pFCPST", icms.PFCPST, true)).
				Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, true))
		}

		if icms.PCredSN > 0 {
			e.Filho(w.nodeAliquota("pCredSN", icms.PCredSN, true)).
				Filho(pcn.NodeDec("vCredICMSSN", icms.VCredICMSSN, 2, true))
		}
	}

	return e
}

// gerarDetImpostoIPI monta o grupo IPI do item. Porte de GerarDetImpostoIPI
// (XmlWriter.pas:2953).
//
// MUTA o documento: cEnq vazio vira "999" (XmlWriter.pas:2990), e e assim
// que o ACBr satisfaz a obrigatoriedade da tag sem recusar o documento.
//
// Tres portoes antes de qualquer coisa:
//
//  1. NFC-e nao tem IPI nenhum;
//  2. CST que nao e 00/49/50/99 so sai quando a opcao manda gerar IPI para
//     nao tributado;
//  3. IPI com TODOS os campos no default -- que e o estado de quem nunca
//     tocou no grupo -- nao sai. Sem isso o zero value de CstIpi ("00")
//     faria todo item declarar IPI tributado.
func (w *xmlWriter) gerarDetImpostoIPI(i int) *pcn.Elem {
	ipi := &w.n.Det[i].Imposto.IPI

	if w.n.Ide.Modelo != 55 {
		return nil
	}

	tributado := ipi.CST == Ipi00 || ipi.CST == Ipi49 || ipi.CST == Ipi50 || ipi.CST == Ipi99

	if w.opcoes.OmitirTagIPIParaNaoTributado && !tributado {
		return nil
	}

	if ipi.CEnq == "" && ipi.CST == Ipi00 && ipi.VBC == 0 && ipi.QUnid == 0 &&
		ipi.VUnid == 0 && ipi.PIPI == 0 && ipi.VIPI == 0 {
		return nil
	}

	e := pcn.NovoElem("IPI")

	// clEnq saiu do leiaute na 4.00.
	if w.versao() < 4 {
		e.Filho(pcn.NodeStr("clEnq", ipi.ClEnq, false))
	}

	e.Filho(pcn.NodeStr("CNPJProd", ipi.CNPJProd, false)).
		Filho(pcn.NodeStr("cSelo", ipi.CSelo, false)).
		Filho(pcn.NodeInt("qSelo", ipi.QSelo, 1, false))

	if ipi.CEnq == "" {
		ipi.CEnq = "999"
	}
	e.Filho(pcn.NodeStr("cEnq", ipi.CEnq, true))

	if !tributado {
		return e.Filho(pcn.NovoElem("IPINT").
			Filho(pcn.NodeStr("CST", ipi.CST.String(), true)))
	}

	trib := pcn.NovoElem("IPITrib").
		Filho(pcn.NodeStr("CST", ipi.CST.String(), true))

	// Base por QUANTIDADE e base por VALOR sao excludentes; a quantidade
	// tem precedencia quando o usuario preencheu as duas.
	if ipi.QUnid+ipi.VUnid > 0 {
		trib.Filho(pcn.NodeDec("qUnid", ipi.QUnid, 4, true)).
			Filho(pcn.NodeDec("vUnid", ipi.VUnid, 4, true))
	} else {
		trib.Filho(pcn.NodeDec("vBC", ipi.VBC, 2, true)).
			Filho(w.nodeAliquota("pIPI", ipi.PIPI, true))
	}
	trib.Filho(pcn.NodeDec("vIPI", ipi.VIPI, 2, true))

	return e.Filho(trib)
}

// gerarDetImpostoII monta o grupo de Imposto de Importacao. Porte de
// GerarDetImpostoII (XmlWriter.pas:3036).
//
// O CFOP de importacao (comecado por 3) abre o grupo sozinho, mesmo com
// todos os valores zerados.
func (w *xmlWriter) gerarDetImpostoII(i int) *pcn.Elem {
	d := w.n.Det[i]
	ii := d.Imposto.II

	if ii.VBC <= 0 && ii.VDespAdu <= 0 && ii.VII <= 0 && ii.VIOF <= 0 &&
		!strings.HasPrefix(d.Prod.CFOP, "3") {
		return nil
	}

	return pcn.NovoElem("II").
		Filho(pcn.NodeDec("vBC", ii.VBC, 2, true)).
		Filho(pcn.NodeDec("vDespAdu", ii.VDespAdu, 2, true)).
		Filho(pcn.NodeDec("vII", ii.VII, 2, true)).
		Filho(pcn.NodeDec("vIOF", ii.VIOF, 2, true))
}

// gerarDetImpostoPIS monta o grupo PIS do item. Porte de GerarDetImpostoPIS
// (XmlWriter.pas:3055).
//
// Na NFC-e o grupo e OPCIONAL: sem valor nenhum e sem um CST que exija
// declaracao, ele some. Na NF-e sai sempre.
func (w *xmlWriter) gerarDetImpostoPIS(i int) *pcn.Elem {
	pis := w.n.Det[i].Imposto.PIS

	if w.n.Ide.Modelo != 55 && pis.VBC == 0 && pis.PPIS == 0 && pis.VPIS == 0 &&
		pis.QBCProd == 0 && pis.VAliqProd == 0 && !pisExigeDeclaracao(pis.CST) {
		return nil
	}

	e := pcn.NovoElem("PIS")
	cst := pcn.NodeStr("CST", pis.CST.String(), true)

	switch {
	case pis.CST == pcn.Pis01 || pis.CST == pcn.Pis02:
		return e.Filho(pcn.NovoElem("PISAliq").
			Filho(cst).
			Filho(pcn.NodeDec("vBC", pis.VBC, 2, true)).
			Filho(w.nodeAliquota("pPIS", pis.PPIS, true)).
			Filho(pcn.NodeDec("vPIS", pis.VPIS, 2, true)))

	case pis.CST == pcn.Pis03:
		return e.Filho(pcn.NovoElem("PISQtde").
			Filho(cst).
			Filho(pcn.NodeDec("qBCProd", pis.QBCProd, 4, true)).
			Filho(pcn.NodeDec("vAliqProd", pis.VAliqProd, 4, true)).
			Filho(pcn.NodeDec("vPIS", pis.VPIS, 2, true)))

	case pisNaoTributado(pis.CST):
		return e.Filho(pcn.NovoElem("PISNT").Filho(cst))

	case pisOutras(pis.CST):
		outr := pcn.NovoElem("PISOutr").Filho(cst)
		if pis.QBCProd+pis.VAliqProd > 0 {
			outr.Filho(pcn.NodeDec("qBCProd", pis.QBCProd, 4, true)).
				Filho(pcn.NodeDec("vAliqProd", pis.VAliqProd, 4, true))
		} else {
			outr.Filho(pcn.NodeDec("vBC", pis.VBC, 2, true)).
				Filho(w.nodeAliquota("pPIS", pis.PPIS, true))
		}
		return e.Filho(outr.Filho(pcn.NodeDec("vPIS", pis.VPIS, 2, true)))
	}

	// CST fora de todas as faixas: o ACBr devolve o grupo PIS vazio.
	return e
}

// pisExigeDeclaracao diz se o CST obriga o grupo PIS a sair mesmo na NFC-e
// sem valores. Sao os nao tributados (04 a 09) e os dois "outros" (49 e 99).
func pisExigeDeclaracao(cst pcn.CSTPis) bool {
	switch cst {
	case pcn.Pis04, pcn.Pis05, pcn.Pis06, pcn.Pis07, pcn.Pis08, pcn.Pis09,
		pcn.Pis49, pcn.Pis99:
		return true
	}
	return false
}

// pisNaoTributado e o conjunto que vai para o grupo PISNT
// (XmlWriter.pas:3094).
func pisNaoTributado(cst pcn.CSTPis) bool {
	switch cst {
	case pcn.Pis04, pcn.Pis05, pcn.Pis06, pcn.Pis07, pcn.Pis08, pcn.Pis09:
		return true
	}
	return false
}

// pisOutras e o conjunto que vai para o grupo PISOutr
// (XmlWriter.pas:3100-3102).
//
// Sao 24 dos 33 CSTs -- todos menos 01 a 09 -- mas a lista esta enumerada
// uma a uma no original e fica assim aqui: depender de "ordinal maior que
// Pis49" transformaria qualquer reordenacao futura do enum em mudanca
// silenciosa de grupo no XML.
func pisOutras(cst pcn.CSTPis) bool {
	switch cst {
	case pcn.Pis49, pcn.Pis50, pcn.Pis51, pcn.Pis52, pcn.Pis53, pcn.Pis54,
		pcn.Pis55, pcn.Pis56, pcn.Pis60, pcn.Pis61, pcn.Pis62, pcn.Pis63,
		pcn.Pis64, pcn.Pis65, pcn.Pis66, pcn.Pis67, pcn.Pis70, pcn.Pis71,
		pcn.Pis72, pcn.Pis73, pcn.Pis74, pcn.Pis75, pcn.Pis98, pcn.Pis99:
		return true
	}
	return false
}

// gerarDetImpostoPISST monta o grupo PISST. Porte de GerarDetImpostoPISST
// (XmlWriter.pas:3137).
//
// ATENCAO: no original os dois ifs internos nao sao excludentes e o segundo
// SOBRESCREVE o Result do primeiro. Quem preencher vBC/pPIS E qBCProd/
// vAliqProd ao mesmo tempo perde o primeiro grupo -- o ACBr emite um alerta
// e segue. Replicado, porque e deterministico.
func (w *xmlWriter) gerarDetImpostoPISST(i int) *pcn.Elem {
	st := w.n.Det[i].Imposto.PISST

	if st.VBC <= 0 && st.PPIS <= 0 && st.QBCProd <= 0 && st.VAliqProd <= 0 &&
		st.VPIS <= 0 {
		return nil
	}

	var e *pcn.Elem

	if st.VBC+st.PPIS > 0 {
		e = pcn.NovoElem("PISST").
			Filho(pcn.NodeDec("vBC", st.VBC, 2, true)).
			Filho(w.nodeAliquota("pPIS", st.PPIS, true)).
			Filho(pcn.NodeDec("vPIS", st.VPIS, 2, true))
		w.anexarIndSomaPISST(e, st)
	}

	if st.QBCProd+st.VAliqProd > 0 {
		e = pcn.NovoElem("PISST").
			Filho(pcn.NodeDec("qBCProd", st.QBCProd, 4, true)).
			Filho(pcn.NodeDec("vAliqProd", st.VAliqProd, 4, true)).
			Filho(pcn.NodeDec("vPIS", st.VPIS, 2, true))
		w.anexarIndSomaPISST(e, st)
	}

	return e
}

// anexarIndSomaPISST acrescenta o indSomaPISST quando a versao permite e o
// indicador nao esta no estado "nenhum".
func (w *xmlWriter) anexarIndSomaPISST(e *pcn.Elem, st PISST) {
	if w.versao() >= 4 && st.IndSomaPISST != IspNenhum {
		e.Filho(pcn.NodeStr("indSomaPISST", st.IndSomaPISST.String(), true))
	}
}

// gerarDetImpostoCOFINS monta o grupo COFINS do item. Porte de
// GerarDetImpostoCOFINS (XmlWriter.pas:3221). Espelha o PIS.
func (w *xmlWriter) gerarDetImpostoCOFINS(i int) *pcn.Elem {
	cof := w.n.Det[i].Imposto.COFINS

	if w.n.Ide.Modelo != 55 && cof.VBC == 0 && cof.PCOFINS == 0 &&
		cof.VCOFINS == 0 && cof.QBCProd == 0 && cof.VAliqProd == 0 &&
		!cofinsExigeDeclaracao(cof.CST) {
		return nil
	}

	e := pcn.NovoElem("COFINS")
	cst := pcn.NodeStr("CST", cof.CST.String(), true)

	switch {
	case cof.CST == pcn.Cof01 || cof.CST == pcn.Cof02:
		return e.Filho(pcn.NovoElem("COFINSAliq").
			Filho(cst).
			Filho(pcn.NodeDec("vBC", cof.VBC, 2, true)).
			Filho(w.nodeAliquota("pCOFINS", cof.PCOFINS, true)).
			Filho(pcn.NodeDec("vCOFINS", cof.VCOFINS, 2, true)))

	case cof.CST == pcn.Cof03:
		return e.Filho(pcn.NovoElem("COFINSQtde").
			Filho(cst).
			Filho(pcn.NodeDec("qBCProd", cof.QBCProd, 4, true)).
			Filho(pcn.NodeDec("vAliqProd", cof.VAliqProd, 4, true)).
			Filho(pcn.NodeDec("vCOFINS", cof.VCOFINS, 2, true)))

	case cofinsNaoTributado(cof.CST):
		return e.Filho(pcn.NovoElem("COFINSNT").Filho(cst))

	case cofinsOutras(cof.CST):
		outr := pcn.NovoElem("COFINSOutr").Filho(cst)
		if cof.QBCProd+cof.VAliqProd > 0 {
			outr.Filho(pcn.NodeDec("qBCProd", cof.QBCProd, 4, true)).
				Filho(pcn.NodeDec("vAliqProd", cof.VAliqProd, 4, true))
		} else {
			outr.Filho(pcn.NodeDec("vBC", cof.VBC, 2, true)).
				Filho(w.nodeAliquota("pCOFINS", cof.PCOFINS, true))
		}
		return e.Filho(outr.Filho(pcn.NodeDec("vCOFINS", cof.VCOFINS, 2, true)))
	}

	return e
}

// cofinsExigeDeclaracao e o par de pisExigeDeclaracao para o COFINS.
func cofinsExigeDeclaracao(cst pcn.CSTCofins) bool {
	switch cst {
	case pcn.Cof04, pcn.Cof05, pcn.Cof06, pcn.Cof07, pcn.Cof08, pcn.Cof09,
		pcn.Cof49, pcn.Cof99:
		return true
	}
	return false
}

// cofinsNaoTributado e o par de pisNaoTributado (XmlWriter.pas:3262).
func cofinsNaoTributado(cst pcn.CSTCofins) bool {
	switch cst {
	case pcn.Cof04, pcn.Cof05, pcn.Cof06, pcn.Cof07, pcn.Cof08, pcn.Cof09:
		return true
	}
	return false
}

// cofinsOutras e o par de pisOutras (XmlWriter.pas:3269-3271).
func cofinsOutras(cst pcn.CSTCofins) bool {
	switch cst {
	case pcn.Cof49, pcn.Cof50, pcn.Cof51, pcn.Cof52, pcn.Cof53, pcn.Cof54,
		pcn.Cof55, pcn.Cof56, pcn.Cof60, pcn.Cof61, pcn.Cof62, pcn.Cof63,
		pcn.Cof64, pcn.Cof65, pcn.Cof66, pcn.Cof67, pcn.Cof70, pcn.Cof71,
		pcn.Cof72, pcn.Cof73, pcn.Cof74, pcn.Cof75, pcn.Cof98, pcn.Cof99:
		return true
	}
	return false
}

// gerarDetImpostoCOFINSST monta o grupo COFINSST. Porte de
// GerarDetImpostoCOFINSST (XmlWriter.pas:3307).
//
// Tem a mesma sobrescrita do PISST: o segundo bloco substitui o primeiro.
func (w *xmlWriter) gerarDetImpostoCOFINSST(i int) *pcn.Elem {
	st := w.n.Det[i].Imposto.COFINSST

	if st.VBC <= 0 && st.PCOFINS <= 0 && st.QBCProd <= 0 && st.VAliqProd <= 0 &&
		st.VCOFINS <= 0 {
		return nil
	}

	var e *pcn.Elem

	if st.VBC+st.PCOFINS > 0 {
		e = pcn.NovoElem("COFINSST").
			Filho(pcn.NodeDec("vBC", st.VBC, 2, true)).
			Filho(w.nodeAliquota("pCOFINS", st.PCOFINS, true)).
			Filho(pcn.NodeDec("vCOFINS", st.VCOFINS, 2, true))
		w.anexarIndSomaCOFINSST(e, st)
	}

	if st.QBCProd+st.VAliqProd > 0 {
		e = pcn.NovoElem("COFINSST").
			Filho(pcn.NodeDec("qBCProd", st.QBCProd, 4, true)).
			Filho(pcn.NodeDec("vAliqProd", st.VAliqProd, 4, true)).
			Filho(pcn.NodeDec("vCOFINS", st.VCOFINS, 2, true))
		w.anexarIndSomaCOFINSST(e, st)
	}

	return e
}

// anexarIndSomaCOFINSST e o par de anexarIndSomaPISST.
func (w *xmlWriter) anexarIndSomaCOFINSST(e *pcn.Elem, st COFINSST) {
	if w.versao() >= 4 && st.IndSomaCOFINSST != IscNenhum {
		e.Filho(pcn.NodeStr("indSomaCOFINSST", st.IndSomaCOFINSST.String(), true))
	}
}

// gerarDetImpostoISSQN monta o grupo ISSQN do item. Porte de
// GerarDetImpostoISSQN (XmlWriter.pas:3355).
//
// O cListServ MUDOU DE FORMATO na 3.10: era de 3 ou 4 posicoes (o codigo
// corrido) e passou a 5, com ponto (ex. "01.01"). O cSitTrib, ao contrario,
// so existe entre a 2.00 e a 3.00 -- fora dessa janela nao e gravado.
func (w *xmlWriter) gerarDetImpostoISSQN(i int) *pcn.Elem {
	iss := w.n.Det[i].Imposto.ISSQN

	if iss.VBC <= 0 && iss.VAliq <= 0 && iss.VISSQN <= 0 && iss.CMunFG <= 0 &&
		iss.CListServ == "" {
		return nil
	}

	e := pcn.NovoElem("ISSQN").
		Filho(pcn.NodeDec("vBC", iss.VBC, 2, true)).
		Filho(w.nodeAliquota("vAliq", iss.VAliq, true)).
		Filho(pcn.NodeDec("vISSQN", iss.VISSQN, 2, true)).
		Filho(pcn.NodeInt("cMunFG", iss.CMunFG, 7, true)).
		Filho(pcn.NodeStr("cListServ", iss.CListServ, true))

	if w.versao() >= 2 && w.versao() < 3 {
		e.Filho(pcn.NodeStr("cSitTrib", iss.CSitTrib.String(), true))
	}

	if w.versao() >= 3.10 {
		e.Filho(pcn.NodeDec("vDeducao", iss.VDeducao, 2, false)).
			Filho(pcn.NodeDec("vOutro", iss.VOutro, 2, false)).
			Filho(pcn.NodeDec("vDescIncond", iss.VDescIncond, 2, false)).
			Filho(pcn.NodeDec("vDescCond", iss.VDescCond, 2, false)).
			// indISSRet esta COMENTADO no ACBr (XmlWriter.pas:3409) e por
			// isso nao e gerado, embora a classe carregue o campo.
			Filho(pcn.NodeDec("vISSRet", iss.VISSRet, 2, false)).
			Filho(pcn.NodeStr("indISS", iss.IndISS.String(), true)).
			Filho(pcn.NodeStr("cServico", iss.CServico, false)).
			Filho(pcn.NodeInt("cMun", iss.CMun, 1, false)).
			Filho(pcn.NodeInt("cPais", iss.CPais, 1, false)).
			Filho(pcn.NodeStr("nProcesso", iss.NProcesso, false)).
			Filho(pcn.NodeStr("indIncentivo", iss.IndIncentivo.String(), true))
	}

	return e
}

// gerarDetObsItem monta as observacoes do item. Porte de GerarDetObsItem
// (XmlWriter.pas:3181).
//
// O criterio de existencia e o xTexto, nao o xCampo -- que vai como ATRIBUTO.
func (w *xmlWriter) gerarDetObsItem(i int) *pcn.Elem {
	d := w.n.Det[i]

	if d.ObsCont.XTexto == "" && d.ObsFisco.XTexto == "" {
		return nil
	}

	e := pcn.NovoElem("obsItem")

	if d.ObsCont.XTexto != "" {
		e.Filho(pcn.NovoElem("obsCont").
			Attr("xCampo", d.ObsCont.XCampo).
			Filho(pcn.NodeStr("xTexto", d.ObsCont.XTexto, true)))
	}
	if d.ObsFisco.XTexto != "" {
		e.Filho(pcn.NovoElem("obsFisco").
			Attr("xCampo", d.ObsFisco.XCampo).
			Filho(pcn.NodeStr("xTexto", d.ObsFisco.XTexto, true)))
	}

	return e
}

// gerarDetDevol monta o imposto devolvido. Porte de GerarDetDevol
// (XmlWriter.pas:3428). Quem decide se ele sai e o gerarDet.
func (w *xmlWriter) gerarDetDevol(i int) *pcn.Elem {
	d := w.n.Det[i]
	return pcn.NovoElem("impostoDevol").
		Filho(pcn.NodeDec("pDevol", d.PDevol, 2, true)).
		Filho(pcn.NovoElem("IPI").
			Filho(pcn.NodeDec("vIPIDevol", d.VIPIDevol, 2, true)))
}

// gerarDetImpostoICMSUFDest monta a partilha do ICMS para a UF de destino
// (EC 87/2015). Porte de GerarDetImpostoICMSUFDest (XmlWriter.pas:3465).
//
// vBCFCPUFDest e pFCPUFDest so existem a partir da 4.00. O ACBr os marca
// como OBRIGATORIOS de propria conta, contrariando a NT 2016.002 v1.50, que
// os tornou opcionais -- a justificativa no codigo e que algumas SEFAZ nao
// implantaram o schema novo. Replicado.
//
// O pICMSInter tem precisao FIXA de duas casas, mesmo nas versoes em que
// todos os demais percentuais vao a quatro.
func (w *xmlWriter) gerarDetImpostoICMSUFDest(i int) *pcn.Elem {
	uf := w.n.Det[i].Imposto.ICMSUFDest

	e := pcn.NovoElem("ICMSUFDest").
		Filho(pcn.NodeDec("vBCUFDest", uf.VBCUFDest, 2, true))

	if w.versao() >= 4 {
		e.Filho(pcn.NodeDec("vBCFCPUFDest", uf.VBCFCPUFDest, 2, true))
	}

	return e.Filho(w.nodeAliquota("pFCPUFDest", uf.PFCPUFDest, true)).
		Filho(w.nodeAliquota("pICMSUFDest", uf.PICMSUFDest, true)).
		Filho(pcn.NodeDec("pICMSInter", uf.PICMSInter, 2, true)).
		Filho(w.nodeAliquota("pICMSInterPart", uf.PICMSInterPart, true)).
		Filho(pcn.NodeDec("vFCPUFDest", uf.VFCPUFDest, 2, true)).
		Filho(pcn.NodeDec("vICMSUFDest", uf.VICMSUFDest, 2, true)).
		Filho(pcn.NodeDec("vICMSUFRemet", uf.VICMSUFRemet, 2, true))
}
