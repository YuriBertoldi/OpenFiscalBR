// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import (
	"fmt"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Escrita .ini dos ITENS. Porte de Gerar_Detalhamento e das procedures que
// ela chama.
//
// Aqui estao as condicionais que mais se perdem num porte, porque nenhuma
// delas aparece em build, vet ou teste de round-trip:
//
//   - o MEDICAMENTO troca de campos conforme a VERSAO DO LEIAUTE. Na 4.00
//     saem cProdANVISA e xMotivoIsencao; antes dela, nLote, qLote, dFab e
//     dVal. Gravar os quatro sempre produz .ini que o validador recusa;
//   - PIS e COFINS escolhem os campos pelo CST. O 01 e o 02 saem por
//     aliquota, o 03 por quantidade, o 99 pelos dois. Qualquer outro CST
//     grava SO o CST -- e o `case` do Delphi nao tem `else`;
//   - o IPI depende do `cEnq`, nao do CST. Um item com IPI tributado e sem
//     codigo de enquadramento nao grava o grupo;
//   - o item so entra em [impostoDevolNNN] com pDevol POSITIVO, mas o
//     pDevol e o vIPIDevol tambem saem na secao do produto, sempre.

// trimINI e o `trim` do Delphi aplicado aos testes de campo preenchido.
func trimINI(s string) string { return strings.TrimSpace(s) }

// gravarINIDetalhamento grava [ProdutoNNN] e tudo que pende do item.
// Porte de Gerar_Detalhamento.
//
// Recebe a NOTA, e nao so os itens, porque o medicamento precisa da versao
// do leiaute -- que vive em InfNFe, fora do item.
func gravarINIDetalhamento(ini *pcn.INI, n *NFe) {
	for i, d := range n.Det {
		secao := fmt.Sprintf("Produto%03d", i+1)

		gravarINIProduto(ini, secao, i, d.Prod)

		ini.GravarString(secao, "infAdProd", d.InfAdProd)
		// Os dois saem AQUI e, quando positivos, tambem na secao propria.
		ini.GravarFloat(secao, "pDevol", d.PDevol, 2)
		ini.GravarFloat(secao, "vIPIDevol", d.VIPIDevol, 2)
		ini.GravarFloat(secao, "vTotTrib", d.Imposto.VTotTrib, 2)
		ini.GravarFloat(secao, "vItem", d.VItem, 2)

		gravarINIProdutoNVE(ini, i, d.Prod.NVE)
		gravarINIProdutoRastro(ini, i, d.Prod.Rastro)
		gravarINIProdutoDI(ini, i, d.Prod.DI)
		gravarINIProdutoDetExportacao(ini, i, d.Prod.DetExport)

		if d.PDevol > 0 {
			s := fmt.Sprintf("impostoDevol%03d", i+1)
			ini.GravarFloat(s, "pDevol", d.PDevol, 2)
			ini.GravarFloat(s, "vIPIDevol", d.VIPIDevol, 2)
		}

		gravarINIProdutoVeiculo(ini, i, d.Prod.VeicProd)
		gravarINIProdutoMedicamento(ini, i, d.Prod.Med, n.InfNFe.Versao)
		gravarINIProdutoArma(ini, i, d.Prod.Arma)
		gravarINIProdutoCombustivel(ini, i, d.Prod.Comb)
		gravarINIProdutoCreditoPresumido(ini, i, d.Prod.CredPresumido)

		gravarINIProdutoImposto(ini, i, d.Imposto, n.Ide)

		gravarINIObsItem(ini, fmt.Sprintf("obsContItem%03d", i+1), d.ObsCont)
		gravarINIObsItem(ini, fmt.Sprintf("obsFiscoItem%03d", i+1), d.ObsFisco)

		rtc.GravarINIDFeReferenciado(ini, d.DFeReferenciado, i+1)
	}
}

func gravarINIProduto(ini *pcn.INI, secao string, idx int, p Prod) {
	ini.GravarInteiro(secao, "nItem", p.NItem)
	ini.GravarString(secao, "cProd", p.CProd)
	ini.GravarString(secao, "cEAN", p.CEAN)
	ini.GravarString(secao, "cBarra", p.CBarra)
	ini.GravarString(secao, "xProd", p.XProd)
	ini.GravarString(secao, "NCM", p.NCM)
	ini.GravarString(secao, "CEST", p.CEST)
	ini.GravarString(secao, "indEscala", p.IndEscala.String())
	ini.GravarString(secao, "CNPJFab", p.CNPJFab)
	ini.GravarString(secao, "cBenef", p.CBenef)
	ini.GravarString(secao, "tpCredPresIBSZFM", p.TpCredPresIBSZFM.String())
	ini.GravarString(secao, "EXTIPI", p.EXTIPI)
	ini.GravarString(secao, "CFOP", p.CFOP)
	ini.GravarString(secao, "uCom", p.UCom)
	ini.GravarFloat(secao, "qCom", p.QCom, 4)
	ini.GravarFloat(secao, "vUnCom", p.VUnCom, 10)
	ini.GravarFloat(secao, "vProd", p.VProd, 2)
	ini.GravarString(secao, "cEANTrib", p.CEANTrib)
	ini.GravarString(secao, "cBarraTrib", p.CBarraTrib)
	ini.GravarString(secao, "uTrib", p.UTrib)
	ini.GravarFloat(secao, "qTrib", p.QTrib, 4)
	ini.GravarFloat(secao, "vUnTrib", p.VUnTrib, 10)
	ini.GravarFloat(secao, "vFrete", p.VFrete, 2)
	ini.GravarFloat(secao, "vSeg", p.VSeg, 2)
	ini.GravarFloat(secao, "vDesc", p.VDesc, 2)
	ini.GravarFloat(secao, "vOutro", p.VOutro, 2)
	ini.GravarString(secao, "IndTot", p.IndTot.String())
	ini.GravarString(secao, "xPed", p.XPed)
	ini.GravarString(secao, "nItemPed", p.NItemPed)
	ini.GravarString(secao, "nFCI", p.NFCI)
	ini.GravarString(secao, "nRECOPI", p.NRECOPI)

	// Reforma Tributaria
	ini.GravarString(secao, "indBemMovelUsado", p.IndBemMovelUsado.String())
}

// gravarINIObsItem grava [obsContItemNNN] ou [obsFiscoItemNNN].
// So sai com TEXTO -- o campo nao e o xCampo.
func gravarINIObsItem(ini *pcn.INI, secao string, o ObsItem) {
	if o.XTexto == "" {
		return
	}
	ini.GravarString(secao, "xCampo", o.XCampo)
	ini.GravarString(secao, "xTexto", o.XTexto)
}

// ---------------------------------------------------------------------------
// Subgrupos do produto
// ---------------------------------------------------------------------------

// gravarINIProdutoNVE grava [NVENNNJJJ].
//
// O laco do original PARA no primeiro item vazio em vez de pula-lo
// (`else Break`, :545): um NVE vazio no meio da lista descarta os
// seguintes. Replicado -- e deterministico, e o leitor tem o mesmo
// criterio de parada, entao os dois lados concordam.
func gravarINIProdutoNVE(ini *pcn.INI, idx int, nves []NVE) {
	for j, v := range nves {
		if v.NVE == "" {
			break
		}
		ini.GravarString(fmt.Sprintf("NVE%03d%03d", idx+1, j+1), "NVE", v.NVE)
	}
}

// gravarINIProdutoRastro grava [RastroNNNJJJ]. Mesmo `else Break` do NVE.
func gravarINIProdutoRastro(ini *pcn.INI, idx int, rastros []Rastro) {
	for j, r := range rastros {
		if r.NLote == "" {
			break
		}
		s := fmt.Sprintf("Rastro%03d%03d", idx+1, j+1)
		ini.GravarString(s, "nLote", r.NLote)
		ini.GravarFloat(s, "qLote", r.QLote, 3)
		ini.GravarData(s, "dFab", r.DFab, layoutDataINI)
		ini.GravarData(s, "dVal", r.DVal, layoutDataINI)
		ini.GravarString(s, "cAgreg", r.CAgreg)
	}
}

// gravarINIProdutoDI grava [DINNNJJJ] e [LADINNNJJJKKK].
// Porte de Gerar_DI.
//
// Dois campos saem CONDICIONADOS ao enum que os acompanha, e nao ao
// proprio valor:
//
//   - vAFRMM so com tpViaTransp = maritima -- o adicional ao frete da
//     marinha mercante nao existe em outra via;
//   - CNPJ e UFTerceiro so quando tpIntermedio NAO e conta propria -- o
//     terceiro so existe quando ha intermediacao.
func gravarINIProdutoDI(ini *pcn.INI, idx int, dis []DI) {
	for j, d := range dis {
		if d.NDi == "" {
			break
		}
		s := fmt.Sprintf("DI%03d%03d", idx+1, j+1)
		ini.GravarString(s, "nDi", d.NDi)
		ini.GravarData(s, "dDi", d.DDi, layoutDataINI)
		ini.GravarString(s, "xLocDesemb", d.XLocDesemb)
		ini.GravarString(s, "UFDesemb", d.UFDesemb)
		ini.GravarData(s, "dDesemb", d.DDesemb, layoutDataINI)
		ini.GravarString(s, "cExportador", d.CExportador)

		if via := d.TpViaTransp.String(); via != "" {
			ini.GravarString(s, "tpViaTransp", via)
			if d.TpViaTransp == TvMaritima {
				ini.GravarFloat(s, "vAFRMM", d.VAFRMM, 2)
			}
		}
		if inter := d.TpIntermedio.String(); inter != "" {
			ini.GravarString(s, "tpIntermedio", inter)
			if d.TpIntermedio != TiContaPropria {
				ini.GravarString(s, "CNPJ", d.CNPJ)
				ini.GravarString(s, "UFTerceiro", d.UFTerceiro)
			}
		}

		for k, a := range d.Adi {
			sa := fmt.Sprintf("LADI%03d%03d%03d", idx+1, j+1, k+1)
			ini.GravarInteiro(sa, "nAdicao", a.NAdicao)
			ini.GravarInteiro(sa, "nSeqAdi", a.NSeqAdi)
			ini.GravarString(sa, "cFabricante", a.CFabricante)
			ini.GravarFloat(sa, "vDescDI", a.VDescDI, 2)
			ini.GravarString(sa, "nDraw", a.NDraw)
		}
	}
}

// gravarINIProdutoDetExportacao grava [detExportNNNJJJ].
//
// CONDICIONAL REPLICADA: o item so sai com `nDraw` preenchido (:634),
// embora o leitor aceite nRE ou chNFe tambem. Uma exportacao sem
// drawback e lida e nao e gravada de volta. Deterministico; mudar a
// condicao faria o .ini divergir do que o ACBr produz.
//
// Diferente do NVE e do rastro, aqui o item vazio e PULADO, nao encerra o
// laco -- o original nao tem `else Break` nesta procedure.
func gravarINIProdutoDetExportacao(ini *pcn.INI, idx int, dets []DetExport) {
	for j, d := range dets {
		if d.NDraw == "" {
			continue
		}
		s := fmt.Sprintf("detExport%03d%03d", idx+1, j+1)
		ini.GravarString(s, "nDraw", d.NDraw)
		ini.GravarString(s, "nRe", d.NRE)
		ini.GravarString(s, "chNFe", d.ChNFe)
		ini.GravarFloat(s, "qExport", d.QExport, 2)
	}
}

func gravarINIProdutoCreditoPresumido(ini *pcn.INI, idx int, creds []CredPresumido) {
	for j, c := range creds {
		// O indice do credito tem UM digito so -- ver lerINIProdutoCredPresumido.
		s := fmt.Sprintf("gCred%03d%d", idx+1, j+1)
		ini.GravarString(s, "cCredPresumido", c.CCredPresumido)
		ini.GravarFloat(s, "pCredPresumido", c.PCredPresumido, 4)
		ini.GravarFloat(s, "vCredPresumido", c.VCredPresumido, 2)
	}
}

// ---------------------------------------------------------------------------
// Mercadorias reguladas
// ---------------------------------------------------------------------------

func gravarINIProdutoVeiculo(ini *pcn.INI, idx int, v VeicProd) {
	if v.Chassi == "" {
		return
	}
	s := fmt.Sprintf("Veiculo%03d", idx+1)
	ini.GravarString(s, "tpOP", v.TpOP.String())
	ini.GravarString(s, "Chassi", v.Chassi)
	ini.GravarString(s, "cCor", v.CCor)
	ini.GravarString(s, "xCor", v.XCor)
	ini.GravarString(s, "pot", v.Pot)
	ini.GravarString(s, "Cilin", v.Cilin)
	ini.GravarString(s, "pesoL", v.PesoL)
	ini.GravarString(s, "pesoB", v.PesoB)
	ini.GravarString(s, "nSerie", v.NSerie)
	ini.GravarString(s, "tpComb", v.TpComb)
	ini.GravarString(s, "nMotor", v.NMotor)
	ini.GravarString(s, "CMT", v.CMT)
	ini.GravarString(s, "dist", v.Dist)
	ini.GravarInteiro(s, "anoMod", v.AnoMod)
	ini.GravarInteiro(s, "anoFab", v.AnoFab)
	ini.GravarString(s, "tpPint", v.TpPint)
	ini.GravarInteiro(s, "tpVeic", v.TpVeic)
	ini.GravarInteiro(s, "espVeic", v.EspVeic)
	ini.GravarString(s, "VIN", v.VIN)
	ini.GravarString(s, "condVeic", v.CondVeic.String())
	ini.GravarString(s, "cMod", v.CMod)
	ini.GravarString(s, "cCorDENATRAN", v.CCorDENATRAN)
	ini.GravarInteiro(s, "lota", v.Lota)
	ini.GravarInteiro(s, "tpRest", v.TpRest)
}

// gravarINIProdutoMedicamento grava [MedicamentoNNNJJJ].
// Porte de Gerar_Medicamento.
//
// CONDICIONAL POR VERSAO DE LEIAUTE -- o tipo de condicional que o
// CLAUDE.md marca como o defeito dominante do porte, e o unico lugar do
// .ini da NF-e onde ela aparece.
//
// A 4.00 trocou o grupo inteiro: saiu o controle de lote (nLote, qLote,
// dFab, dVal), que passou para o grupo `rastro`, e entraram o codigo ANVISA
// e o motivo de isencao. O vPMC sai nas duas.
func gravarINIProdutoMedicamento(ini *pcn.INI, idx int, meds []Med, versao float64) {
	for j, m := range meds {
		s := fmt.Sprintf("Medicamento%03d%03d", idx+1, j+1)

		if versao >= 4 {
			ini.GravarString(s, "cProdANVISA", m.CProdANVISA)
			ini.GravarString(s, "xMotivoIsencao", m.XMotivoIsencao)
		} else {
			ini.GravarString(s, "nLote", m.NLote)
			ini.GravarFloat(s, "qLote", m.QLote, 3)
			ini.GravarData(s, "dFab", m.DFab, layoutDataINI)
			ini.GravarData(s, "dVal", m.DVal, layoutDataINI)
		}
		ini.GravarFloat(s, "vPMC", m.VPMC, 2)
	}
}

func gravarINIProdutoArma(ini *pcn.INI, idx int, armas []Arma) {
	for j, a := range armas {
		s := fmt.Sprintf("Arma%03d%03d", idx+1, j+1)
		ini.GravarString(s, "tpArma", a.TpArma.String())
		ini.GravarString(s, "nSerie", a.NSerie)
		ini.GravarString(s, "nCano", a.NCano)
		ini.GravarString(s, "descr", a.Descr)
	}
}

// gravarINIProdutoCombustivel grava [CombustivelNNN] e os seis subgrupos.
//
// O grupo principal depende do cProdANP POSITIVO; os subgrupos origComb,
// CIDE, encerrante e ICMSComb saem INDEPENDENTEMENTE dele, porque no
// original sao procedures separadas chamadas pelo Gerar_Detalhamento. Aqui
// ficam juntas, mas as condicoes sao as mesmas.
func gravarINIProdutoCombustivel(ini *pcn.INI, idx int, c Comb) {
	if c.CProdANP > 0 {
		s := fmt.Sprintf("Combustivel%03d", idx+1)
		ini.GravarInteiro(s, "cProdANP", c.CProdANP)
		ini.GravarFloat(s, "pMixGN", c.PMixGN, 4)
		ini.GravarString(s, "descANP", c.DescANP)
		ini.GravarFloat(s, "pGLP", c.PGLP, 4)
		ini.GravarFloat(s, "pGNn", c.PGNn, 4)
		ini.GravarFloat(s, "pGNi", c.PGNi, 4)
		ini.GravarFloat(s, "vPart", c.VPart, 2)
		ini.GravarString(s, "CODIF", c.CODIF)
		ini.GravarFloat(s, "qTemp", c.QTemp, 4)
		ini.GravarString(s, "UFCons", c.UFCons)
		ini.GravarFloat(s, "pBio", c.PBio, 4)
	}

	for j, o := range c.OrigComb {
		// Indice da origem com DOIS digitos.
		s := fmt.Sprintf("origComb%03d%02d", idx+1, j+1)
		ini.GravarString(s, "indImport", o.IndImport.String())
		ini.GravarInteiro(s, "cUFOrig", o.CUFOrig)
		ini.GravarFloat(s, "pOrig", o.POrig, 4)
	}

	cide := fmt.Sprintf("CIDE%03d", idx+1)
	ini.GravarFloat(cide, "qBCprod", c.CIDE.QBCProd, 4)
	ini.GravarFloat(cide, "vAliqProd", c.CIDE.VAliqProd, 4)
	ini.GravarFloat(cide, "vCIDE", c.CIDE.VCIDE, 2)

	enc := fmt.Sprintf("encerrante%03d", idx+1)
	ini.GravarInteiro(enc, "nBico", c.Encerrante.NBico)
	ini.GravarInteiro(enc, "nBomba", c.Encerrante.NBomba)
	ini.GravarInteiro(enc, "nTanque", c.Encerrante.NTanque)
	ini.GravarFloat(enc, "vEncIni", c.Encerrante.VEncIni, 3)
	ini.GravarFloat(enc, "vEncFin", c.Encerrante.VEncFin, 3)

	icms := fmt.Sprintf("ICMSComb%03d", idx+1)
	ini.GravarFloat(icms, "vBCICMS", c.ICMS.VBCICMS, 2)
	ini.GravarFloat(icms, "vICMS", c.ICMS.VICMS, 2)
	ini.GravarFloat(icms, "vBCICMSST", c.ICMS.VBCICMSST, 2)
	ini.GravarFloat(icms, "vICMSST", c.ICMS.VICMSST, 2)

	if c.ICMSInter.VBCICMSSTDest > 0 {
		s := fmt.Sprintf("ICMSInter%03d", idx+1)
		ini.GravarFloat(s, "vBCICMSSTDest", c.ICMSInter.VBCICMSSTDest, 2)
		ini.GravarFloat(s, "vICMSSTDest", c.ICMSInter.VICMSSTDest, 2)
	}
	if c.ICMSCons.VBCICMSSTCons > 0 {
		s := fmt.Sprintf("ICMSCons%03d", idx+1)
		ini.GravarFloat(s, "vBCICMSSTCons", c.ICMSCons.VBCICMSSTCons, 2)
		ini.GravarFloat(s, "vICMSSTCons", c.ICMSCons.VICMSSTCons, 2)
		ini.GravarString(s, "UFCons", c.ICMSCons.UFCons)
	}
}

// ---------------------------------------------------------------------------
// Impostos do item
// ---------------------------------------------------------------------------

// gravarINIProdutoImposto grava os onze grupos de imposto do item.
// Porte de Gerar_Imposto.
//
// Recebe a IDE porque o grupo IBSCBS depende do modelo do documento e do
// tpNFDebito -- ver rtc.OpcoesIBSCBS.
func gravarINIProdutoImposto(ini *pcn.INI, idx int, imp Imposto, ide Ide) {
	gravarINIImpostoICMS(ini, idx, imp.ICMS)
	gravarINIImpostoICMSUFDest(ini, idx, imp.ICMSUFDest)
	gravarINIImpostoIPI(ini, idx, imp.IPI)
	gravarINIImpostoII(ini, idx, imp.II)
	gravarINIImpostoPIS(ini, idx, imp.PIS)
	gravarINIImpostoPISST(ini, idx, imp.PISST)
	gravarINIImpostoCOFINS(ini, idx, imp.COFINS)
	gravarINIImpostoCOFINSST(ini, idx, imp.COFINSST)
	gravarINIImpostoISSQN(ini, idx, imp.ISSQN)

	// O Imposto Seletivo so sai com CST preenchido -- que aqui e STRING,
	// porque a tabela oficial de CSTs do IS ainda nao foi publicada.
	if imp.ISel.CSTIS != "" {
		rtc.GravarINIISel(ini, imp.ISel, idx+1)
	}
	rtc.GravarINIIBSCBS(ini, imp.IBSCBS, idx+1, rtc.SemIndice, opcoesIBSCBSDaNota(ide))
}

// opcoesIBSCBSDaNota traduz a identificacao da nota no que o gravador do
// grupo IBSCBS precisa saber.
//
// O modelo sai do `mod` do documento, nao da configuracao do componente: o
// .ini pode ser gerado a partir de uma nota ja montada, sem componente
// configurado. O tpNFDebito e convertido pelo ORDINAL -- os dois enums
// listam os mesmos codigos na mesma ordem, o que esta travado por
// TestTpNFDebito_OrdinaisAlinhadosComRTC.
func opcoesIBSCBSDaNota(ide Ide) rtc.OpcoesIBSCBS {
	modelo := rtc.ModeloNFe
	if ide.Modelo == 65 {
		modelo = rtc.ModeloNFCe
	}
	return rtc.OpcoesIBSCBS{
		Modelo:     modelo,
		TpNFDebito: rtc.TpNFDebito(ide.TpNFDebito),
	}
}

func gravarINIImpostoICMS(ini *pcn.INI, idx int, v ICMS) {
	s := fmt.Sprintf("ICMS%03d", idx+1)

	ini.GravarString(s, "orig", v.Orig.String())
	ini.GravarString(s, "CST", v.CST.String())
	ini.GravarString(s, "CSOSN", v.CSOSN.String())
	ini.GravarString(s, "modBC", v.ModBC.String())
	ini.GravarFloat(s, "pRedBC", v.PRedBC, 4)
	ini.GravarString(s, "cBenefRBC", v.CBenefRBC)
	ini.GravarFloat(s, "vBC", v.VBC, 2)
	ini.GravarFloat(s, "pICMS", v.PICMS, 4)
	ini.GravarFloat(s, "vICMS", v.VICMS, 2)
	ini.GravarFloat(s, "vBCFCP", v.VBCFCP, 2)
	ini.GravarFloat(s, "pFCP", v.PFCP, 4)
	ini.GravarFloat(s, "vFCP", v.VFCP, 2)
	ini.GravarString(s, "modBCST", v.ModBCST.String())
	ini.GravarFloat(s, "pMVAST", v.PMVAST, 4)
	ini.GravarFloat(s, "pRedBCST", v.PRedBCST, 4)
	ini.GravarFloat(s, "vBCST", v.VBCST, 2)
	ini.GravarFloat(s, "pICMSST", v.PICMSST, 4)
	ini.GravarFloat(s, "vICMSST", v.VICMSST, 2)
	ini.GravarFloat(s, "vBCFCPST", v.VBCFCPST, 2)
	ini.GravarFloat(s, "pFCPST", v.PFCPST, 4)
	ini.GravarFloat(s, "vFCPST", v.VFCPST, 2)
	ini.GravarString(s, "UFST", v.UFST)
	ini.GravarFloat(s, "pBCOp", v.PBCOp, 4)
	ini.GravarFloat(s, "vBCSTRet", v.VBCSTRet, 2)
	ini.GravarFloat(s, "pST", v.PST, 4)
	ini.GravarFloat(s, "vICMSSTRet", v.VICMSSTRet, 2)
	ini.GravarFloat(s, "vBCFCPSTRet", v.VBCFCPSTRet, 2)
	ini.GravarFloat(s, "pFCPSTRet", v.PFCPSTRet, 4)
	ini.GravarFloat(s, "vFCPSTRet", v.VFCPSTRet, 2)
	ini.GravarString(s, "motDesICMS", v.MotDesICMS.String())
	ini.GravarFloat(s, "pCredSN", v.PCredSN, 4)
	ini.GravarFloat(s, "vCredICMSSN", v.VCredICMSSN, 2)
	ini.GravarFloat(s, "vBCSTDest", v.VBCSTDest, 2)
	ini.GravarFloat(s, "vICMSSTDest", v.VICMSSTDest, 2)
	ini.GravarFloat(s, "vICMSDeson", v.VICMSDeson, 2)
	ini.GravarFloat(s, "vICMSOp", v.VICMSOp, 2)
	ini.GravarFloat(s, "pDif", v.PDif, 4)
	ini.GravarFloat(s, "vICMSDif", v.VICMSDif, 2)
	ini.GravarFloat(s, "pRedBCEfet", v.PRedBCEfet, 4)
	ini.GravarFloat(s, "vBCEfet", v.VBCEfet, 2)
	ini.GravarFloat(s, "pICMSEfet", v.PICMSEfet, 4)
	ini.GravarFloat(s, "vICMSEfet", v.VICMSEfet, 2)
	ini.GravarFloat(s, "vICMSSubstituto", v.VICMSSubstituto, 2)
	ini.GravarFloat(s, "vICMSSTDeson", v.VICMSSTDeson, 2)
	ini.GravarString(s, "motDesICMSST", v.MotDesICMSST.String())
	ini.GravarFloat(s, "pFCPDif", v.PFCPDif, 4)
	ini.GravarFloat(s, "vFCPDif", v.VFCPDif, 2)
	ini.GravarFloat(s, "vFCPEfet", v.VFCPEfet, 2)
	ini.GravarFloat(s, "adRemICMS", v.AdRemICMS, 4)
	ini.GravarFloat(s, "vICMSMono", v.VICMSMono, 2)
	ini.GravarFloat(s, "adRemICMSReten", v.AdRemICMSReten, 4)
	ini.GravarFloat(s, "vICMSMonoReten", v.VICMSMonoReten, 2)
	ini.GravarFloat(s, "vICMSMonoDif", v.VICMSMonoDif, 2)
	ini.GravarFloat(s, "adRemICMSRet", v.AdRemICMSRet, 4)
	ini.GravarFloat(s, "vICMSMonoRet", v.VICMSMonoRet, 2)
	ini.GravarFloat(s, "qBCMono", v.QBCMono, 4)
	ini.GravarFloat(s, "qBCMonoReten", v.QBCMonoReten, 4)
	ini.GravarFloat(s, "pRedAdRem", v.PRedAdRem, 4)
	ini.GravarString(s, "motRedAdRem", v.MotRedAdRem.String())
	ini.GravarFloat(s, "qBCMonoRet", v.QBCMonoRet, 4)
	ini.GravarFloat(s, "vICMSMonoOp", v.VICMSMonoOp, 2)
	ini.GravarString(s, "indDeduzDeson", v.IndDeduzDeson.String())
}

// gravarINIImpostoICMSUFDest grava [ICMSUFDESTNNN] -- com DEST em
// maiusculas, ao contrario do leitor, que procura `ICMSUFDest`. A
// comparacao de secao ignora caixa nos dois lados, entao a ida e volta
// fecha; a grafia esta preservada para que o arquivo saia identico ao do
// ACBr.
func gravarINIImpostoICMSUFDest(ini *pcn.INI, idx int, v ICMSUFDest) {
	s := fmt.Sprintf("ICMSUFDEST%03d", idx+1)
	ini.GravarFloat(s, "vBCUFDest", v.VBCUFDest, 2)
	ini.GravarFloat(s, "vBCFCPUFDest", v.VBCFCPUFDest, 2)
	ini.GravarFloat(s, "pICMSUFDest", v.PICMSUFDest, 4)
	ini.GravarFloat(s, "pICMSInter", v.PICMSInter, 4)
	ini.GravarFloat(s, "pICMSInterPart", v.PICMSInterPart, 4)
	ini.GravarFloat(s, "vICMSUFDest", v.VICMSUFDest, 2)
	ini.GravarFloat(s, "vICMSUFRemet", v.VICMSUFRemet, 2)
	ini.GravarFloat(s, "pFCPUFDest", v.PFCPUFDest, 4)
	ini.GravarFloat(s, "vFCPUFDest", v.VFCPUFDest, 2)
}

// gravarINIImpostoIPI grava [IPINNN].
//
// Condicionado ao `cEnq`, e NAO ao CST: o codigo de enquadramento e
// obrigatorio em todo IPI do leiaute, entao a ausencia dele significa que
// o grupo nao foi preenchido.
func gravarINIImpostoIPI(ini *pcn.INI, idx int, v IPI) {
	if v.CEnq == "" {
		return
	}
	s := fmt.Sprintf("IPI%03d", idx+1)
	ini.GravarString(s, "CST", v.CST.String())
	ini.GravarString(s, "cEnq", v.CEnq)
	ini.GravarString(s, "clEnq", v.ClEnq)
	ini.GravarString(s, "CNPJProd", v.CNPJProd)
	ini.GravarString(s, "cSelo", v.CSelo)
	ini.GravarInteiro(s, "qSelo", v.QSelo)
	ini.GravarFloat(s, "vBC", v.VBC, 2)
	ini.GravarFloat(s, "qUnid", v.QUnid, 4)
	ini.GravarFloat(s, "vUnid", v.VUnid, 4)
	ini.GravarFloat(s, "pIPI", v.PIPI, 4)
	ini.GravarFloat(s, "vIPI", v.VIPI, 2)
}

func gravarINIImpostoII(ini *pcn.INI, idx int, v II) {
	if v.VBC <= 0 {
		return
	}
	s := fmt.Sprintf("II%03d", idx+1)
	ini.GravarFloat(s, "vBc", v.VBC, 2)
	ini.GravarFloat(s, "vDespAdu", v.VDespAdu, 2)
	ini.GravarFloat(s, "vII", v.VII, 2)
	ini.GravarFloat(s, "vIOF", v.VIOF, 2)
}

// gravarINIImpostoPIS grava [PISNNN]. Porte de Gerar_ImpostoPIS.
//
// Os campos saem CONFORME O CST, e o `case` do Delphi nao tem `else`:
//
//	01, 02  tributacao por ALIQUOTA  -> vBC, pPIS, vPIS
//	03      tributacao por QUANTIDADE -> qBCProd, vAliqProd, vPIS
//	99      outras operacoes          -> os cinco
//
// Qualquer outro CST grava SO o CST. Gravar os cinco sempre produziria um
// .ini que o validador da SEFAZ recusa -- vBC e qBCProd sao excludentes.
func gravarINIImpostoPIS(ini *pcn.INI, idx int, v PIS) {
	s := fmt.Sprintf("PIS%03d", idx+1)
	ini.GravarString(s, "CST", v.CST.String())

	switch v.CST {
	case pcn.Pis01, pcn.Pis02:
		ini.GravarFloat(s, "vBC", v.VBC, 2)
		ini.GravarFloat(s, "pPIS", v.PPIS, 4)
		ini.GravarFloat(s, "vPIS", v.VPIS, 2)
	case pcn.Pis03:
		ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
		ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
		ini.GravarFloat(s, "vPIS", v.VPIS, 2)
	case pcn.Pis99:
		ini.GravarFloat(s, "vBC", v.VBC, 2)
		ini.GravarFloat(s, "pPIS", v.PPIS, 4)
		ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
		ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
		ini.GravarFloat(s, "vPIS", v.VPIS, 2)
	}
}

func gravarINIImpostoPISST(ini *pcn.INI, idx int, v PISST) {
	if v.VBC <= 0 {
		return
	}
	s := fmt.Sprintf("PISST%03d", idx+1)
	ini.GravarFloat(s, "vBc", v.VBC, 2)
	ini.GravarFloat(s, "pPis", v.PPIS, 4)
	ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
	ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
	ini.GravarFloat(s, "vPIS", v.VPIS, 2)
	ini.GravarString(s, "indSomaPISST", v.IndSomaPISST.String())
}

// gravarINIImpostoCOFINS grava [COFINSNNN], com o mesmo `case` por CST do
// PIS -- ver gravarINIImpostoPIS.
func gravarINIImpostoCOFINS(ini *pcn.INI, idx int, v COFINS) {
	s := fmt.Sprintf("COFINS%03d", idx+1)
	ini.GravarString(s, "CST", v.CST.String())

	switch v.CST {
	case pcn.Cof01, pcn.Cof02:
		ini.GravarFloat(s, "vBC", v.VBC, 2)
		ini.GravarFloat(s, "pCOFINS", v.PCOFINS, 4)
		ini.GravarFloat(s, "vCOFINS", v.VCOFINS, 2)
	case pcn.Cof03:
		ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
		ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
		ini.GravarFloat(s, "vCOFINS", v.VCOFINS, 2)
	case pcn.Cof99:
		ini.GravarFloat(s, "vBC", v.VBC, 2)
		ini.GravarFloat(s, "pCOFINS", v.PCOFINS, 4)
		ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
		ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
		ini.GravarFloat(s, "vCOFINS", v.VCOFINS, 2)
	}
}

func gravarINIImpostoCOFINSST(ini *pcn.INI, idx int, v COFINSST) {
	if v.VBC <= 0 {
		return
	}
	s := fmt.Sprintf("COFINSST%03d", idx+1)
	ini.GravarFloat(s, "vBC", v.VBC, 2)
	ini.GravarFloat(s, "pCOFINS", v.PCOFINS, 4)
	ini.GravarFloat(s, "qBCProd", v.QBCProd, 4)
	ini.GravarFloat(s, "vAliqProd", v.VAliqProd, 4)
	ini.GravarFloat(s, "vCOFINS", v.VCOFINS, 2)
	ini.GravarString(s, "indSomaCOFINSST", v.IndSomaCOFINSST.String())
}

func gravarINIImpostoISSQN(ini *pcn.INI, idx int, v ISSQN) {
	if v.VBC <= 0 {
		return
	}
	s := fmt.Sprintf("ISSQN%03d", idx+1)
	ini.GravarFloat(s, "vBC", v.VBC, 2)
	ini.GravarFloat(s, "vAliq", v.VAliq, 4)
	ini.GravarFloat(s, "vISSQN", v.VISSQN, 2)
	ini.GravarInteiro(s, "cMunFG", v.CMunFG)
	ini.GravarString(s, "cListServ", v.CListServ)
	ini.GravarString(s, "cSitTrib", v.CSitTrib.String())
	ini.GravarFloat(s, "vDeducao", v.VDeducao, 2)
	ini.GravarFloat(s, "vOutro", v.VOutro, 2)
	ini.GravarFloat(s, "vDescIncond", v.VDescIncond, 2)
	ini.GravarFloat(s, "vDescCond", v.VDescCond, 2)
	ini.GravarFloat(s, "vISSRet", v.VISSRet, 2)
	ini.GravarString(s, "indISS", v.IndISS.String())
	ini.GravarString(s, "cServico", v.CServico)
	ini.GravarInteiro(s, "cMun", v.CMun)
	ini.GravarInteiro(s, "cPais", v.CPais)
	ini.GravarString(s, "nProcesso", v.NProcesso)
	ini.GravarString(s, "indIncentivo", v.IndIncentivo.String())
}
