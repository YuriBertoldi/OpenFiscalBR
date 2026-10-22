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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitura .ini dos ITENS da nota -- produto, mercadorias reguladas e os onze
// grupos de imposto. Porte de Ler_Produto e das procedures que ela chama.
//
// E a metade mais densa do formato, e o que a torna densa nao e a quantidade
// de campos: e a INDEXACAO. Um subgrupo do item pode indexar de tres
// maneiras, e a largura nao e uniforme:
//
//	[detNNN]            item, 3 digitos
//	[NVENNNJJJ]         subgrupo do item, 3 + 3
//	[gCredNNNJ]         subgrupo do item, 3 + 1   <- um digito so
//	[origCombNNNJJ]     subgrupo do item, 3 + 2   <- dois digitos
//	[adiNNNJJJKKK]      sub-subgrupo, 3 + 3 + 3
//
// Errar a largura nao da erro: a secao simplesmente nao e encontrada, o
// grupo sai vazio e a nota e transmitida sem ele.
//
// A outra armadilha e a SENTINELA de cada grupo, que quase nunca e a
// primeira chave e nem sempre e a mesma coisa:
//
//	det        `Codigo` ou `cProd`, e so "FIM" encerra (vazio NAO encerra)
//	ICMS       `CST`, caindo para `CSOSN`
//	PISST      a CONCATENACAO de duas chaves com defaults "F" e "IM"
//	ISSQN      a base de calculo, e ainda tem de ser POSITIVA
//	detExport  `nRE`, caindo para `chNFe`, caindo para `nDraw`

// lerINIProdutos le [ProdutoNNN] ou [detNNN] e tudo que pende do item.
// Porte de Ler_Produto.
//
// A sentinela do item e peculiar de DUAS maneiras, e as duas vem da mesma
// linha do original (:491):
//
//		sProdID := ReadString(sSecao, 'Codigo', ReadString(sSecao, 'cProd', 'FIM'))
//		if sProdID = 'FIM' then break
//
//	 1. so o literal "FIM" encerra. Um `Codigo=` VAZIO NAO encerra -- falta o
//	    `Length(...) <= 0` que os demais grupos tem. Um item sem codigo entra
//	    na nota e e a validacao que o recusa;
//	 2. o "FIM" so aparece quando as DUAS chaves faltam, porque e o default
//	    da leitura mais interna. Secao inexistente cai nesse caso, e e assim
//	    que o laco termina.
//
// Por isso a cadeia e escrita aqui a mao, e nao com lerStringAlias: aquele
// devolve "" tanto para chave ausente quanto para chave vazia, e os dois
// casos decidem coisas diferentes neste laco.
func lerINIProdutos(ini *pcn.INI, n *NFe) {
	const ausente = "\x00"

	n.Det = nil
	for i := 1; ; i++ {
		secao := secaoAlias(ini, fmt.Sprintf("Produto%03d", i), fmt.Sprintf("det%03d", i))

		codigo := ini.LerString(secao, "Codigo", ausente)
		if codigo == ausente {
			codigo = ini.LerString(secao, "cProd", "FIM")
		}
		if codigo == "FIM" {
			break
		}

		d := NovoDetItem()
		d.Prod.NItem = i
		d.InfAdProd = ini.LerString(secao, "infAdProd", "")

		lerINIProduto(ini, secao, i, &d.Prod)

		d.PDevol = ini.LerFloat(secao, "pDevol", 0)
		d.VIPIDevol = ini.LerFloat(secao, "vIPIDevol", 0)
		d.Imposto.VTotTrib = ini.LerFloat(secao, "vTotTrib", 0)
		d.VItem = ini.LerFloat(secao, "vItem", 0)

		// [impostoDevolNNN] SOBRESCREVE o pDevol/vIPIDevol lidos da secao
		// do item (:552-558). E o grupo proprio do leiaute; a leitura na
		// secao do item e atalho historico.
		lerINIImpostoDevolucao(ini, i, &d)

		lerINIProdutoVeiculo(ini, i, &d.Prod.VeicProd)
		lerINIProdutoMedicamento(ini, i, &d.Prod)
		lerINIProdutoArma(ini, i, &d.Prod)
		lerINIProdutoCombustivel(ini, i, &d.Prod.Comb)
		lerINIProdutoImposto(ini, i, &d.Imposto)

		d.ObsCont = lerINIObsItem(ini, fmt.Sprintf("obsContItem%03d", i))
		d.ObsFisco = lerINIObsItem(ini, fmt.Sprintf("obsFiscoItem%03d", i))

		rtc.LerINIDFeReferenciado(ini, &d.DFeReferenciado, i)

		n.Det = append(n.Det, d)
	}
}

// lerINIProduto le os campos do grupo prod e os subgrupos indexados.
func lerINIProduto(ini *pcn.INI, secao string, idx int, p *Prod) {
	p.CProd = lerStringAlias(ini, secao, "cProd", "Codigo")

	// cEAN e cEANTrib so sao ATRIBUIDOS quando ha conteudo (:500 e :518).
	// Nao e detalhe: o XSD recusa <cEAN></cEAN> vazio, e o gerador so emite
	// a tag se o campo tiver valor.
	if v := lerStringAlias(ini, secao, "cEAN", "EAN"); v != "" {
		p.CEAN = v
	}
	if v := ini.LerString(secao, "cEANTrib", ""); v != "" {
		p.CEANTrib = v
	}

	p.CBarra = ini.LerString(secao, "cBarra", "")
	p.XProd = lerStringAlias(ini, secao, "xProd", "Descricao")
	p.NCM = ini.LerString(secao, "NCM", "")
	p.CEST = ini.LerString(secao, "CEST", "")
	p.IndEscala, _ = ParseIndEscala(ini.LerString(secao, "indEscala", ""))
	p.CNPJFab = ini.LerString(secao, "CNPJFab", "")
	p.CBenef = ini.LerString(secao, "cBenef", "")
	p.TpCredPresIBSZFM, _ = rtc.ParseTpCredPresIBSZFM(
		ini.LerString(secao, "tpCredPresIBSZFM", ""))
	p.EXTIPI = ini.LerString(secao, "EXTIPI", "")
	p.CFOP = ini.LerString(secao, "CFOP", "")
	p.UCom = lerStringAlias(ini, secao, "uCom", "Unidade")
	p.QCom = lerFloatAlias(ini, secao, "qCom", "Quantidade")
	p.VUnCom = lerFloatAlias(ini, secao, "vUnCom", "ValorUnitario")
	p.VProd = lerFloatAlias(ini, secao, "vProd", "ValorTotal")

	p.CBarraTrib = ini.LerString(secao, "cBarraTrib", "")

	// A unidade e a quantidade TRIBUTAVEIS caem para as COMERCIAIS
	// (:521-523). E a razao de as tres leituras virem depois de uCom/qCom/
	// vUnCom nesta ordem -- inverter deixaria o default zerado.
	p.UTrib = ini.LerString(secao, "uTrib", p.UCom)
	p.QTrib = ini.LerFloat(secao, "qTrib", p.QCom)
	p.VUnTrib = ini.LerFloat(secao, "vUnTrib", p.VUnCom)

	p.VFrete = ini.LerFloat(secao, "vFrete", 0)
	p.VSeg = ini.LerFloat(secao, "vSeg", 0)
	p.VDesc = lerFloatAlias(ini, secao, "vDesc", "ValorDesconto")
	p.VOutro = ini.LerFloat(secao, "vOutro", 0)
	p.IndTot, _ = ParseIndicadorTotal(ini.LerString(secao, "indTot", "1"))
	p.XPed = ini.LerString(secao, "xPed", "")
	p.NItemPed = ini.LerString(secao, "nItemPed", "")
	p.NFCI = ini.LerString(secao, "nFCI", "")
	p.NRECOPI = ini.LerString(secao, "nRECOPI", "")

	p.IndBemMovelUsado, _ = pcn.ParseIndicadorEx(
		ini.LerString(secao, "indBemMovelUsado", ""))

	lerINIProdutoCredPresumido(ini, idx, p)
	lerINIProdutoNVE(ini, idx, p)
	lerINIProdutoRastro(ini, idx, p)
	lerINIProdutoDI(ini, idx, p)
	lerINIProdutoDetExport(ini, idx, p)
}

// lerINIImpostoDevolucao le [impostoDevolNNN]. Porte de :552-558.
func lerINIImpostoDevolucao(ini *pcn.INI, idx int, d *DetItem) {
	secao := fmt.Sprintf("impostoDevol%03d", idx)
	v := ini.LerString(secao, "pDevol", "FIM")
	if v == "FIM" || v == "" {
		return
	}
	d.PDevol = ini.LerFloat(secao, "pDevol", 0)
	d.VIPIDevol = ini.LerFloat(secao, "vIPIDevol", 0)
}

// lerINIObsItem le [obsContItemNNN] ou [obsFiscoItemNNN].
// Porte de Ler_ProdutoObsContItem e Ler_ProdutoObsFiscoItem.
//
// Diferente dos homonimos do cabecalho, estes NAO sao repetidos: cada item
// tem no maximo uma observacao de cada tipo.
func lerINIObsItem(ini *pcn.INI, secao string) ObsItem {
	return ObsItem{
		XCampo: ini.LerString(secao, "xCampo", ""),
		XTexto: ini.LerString(secao, "xTexto", ""),
	}
}

// ---------------------------------------------------------------------------
// Subgrupos do produto
// ---------------------------------------------------------------------------

// lerINIProdutoCredPresumido le [gCredNNNJ] -- indice do item com TRES
// digitos e do credito com UM. Porte de Ler_ProdutogCred.
//
// O digito unico limita o grupo a nove ocorrencias por item. E o que o
// original faz (IntToStrZero(J,1), :584), e nao ha indicio de ser engano: o
// leiaute nao preve mais de um credito presumido por item.
func lerINIProdutoCredPresumido(ini *pcn.INI, idx int, p *Prod) {
	p.CredPresumido = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("gCred%03d%d", idx, j)
		codigo := ini.LerString(secao, "cCredPresumido", "")
		if codigo == "" {
			break
		}
		p.CredPresumido = append(p.CredPresumido, CredPresumido{
			CCredPresumido: codigo,
			PCredPresumido: ini.LerFloat(secao, "pCredPresumido", 0),
			VCredPresumido: ini.LerFloat(secao, "vCredPresumido", 0),
		})
	}
}

// lerINIProdutoNVE le [NVENNNJJJ]. Porte de Ler_ProdutoNVE.
func lerINIProdutoNVE(ini *pcn.INI, idx int, p *Prod) {
	p.NVE = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("NVE%03d%03d", idx, j)
		v := ini.LerString(secao, "NVE", "")
		if v == "" {
			break
		}
		p.NVE = append(p.NVE, NVE{NVE: v})
	}
}

// lerINIProdutoRastro le [rastroNNNJJJ]. Porte de Ler_ProdutoRastro.
func lerINIProdutoRastro(ini *pcn.INI, idx int, p *Prod) {
	p.Rastro = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("rastro%03d%03d", idx, j)
		lote := ini.LerString(secao, "nLote", "")
		if lote == "" {
			break
		}
		p.Rastro = append(p.Rastro, Rastro{
			NLote:  lote,
			QLote:  ini.LerFloat(secao, "qLote", 0),
			DFab:   ini.LerData(secao, "dFab", tempoZero()),
			DVal:   ini.LerData(secao, "dVal", tempoZero()),
			CAgreg: ini.LerString(secao, "cAgreg", ""),
		})
	}
}

// lerINIProdutoDI le [DINNNJJJ] e as adicoes de cada uma.
// Porte de Ler_ProdutoDI.
func lerINIProdutoDI(ini *pcn.INI, idx int, p *Prod) {
	p.DI = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("DI%03d%03d", idx, j)

		numero := lerStringAlias(ini, secao, "NumeroDI", "nDi")
		if numero == "" {
			break
		}

		d := DI{
			NDi:         numero,
			DDi:         lerDataAlias(ini, secao, "dDi", "DataRegistroDI"),
			XLocDesemb:  lerStringAlias(ini, secao, "xLocDesemb", "LocalDesembaraco"),
			UFDesemb:    lerStringAlias(ini, secao, "UFDesemb", "UFDesembaraco"),
			DDesemb:     lerDataAlias(ini, secao, "dDesemb", "DataDesembaraco"),
			VAFRMM:      ini.LerFloat(secao, "vAFRMM", 0),
			CNPJ:        ini.LerString(secao, "CNPJ", ""),
			UFTerceiro:  ini.LerString(secao, "UFTerceiro", ""),
			CExportador: lerStringAlias(ini, secao, "cExportador", "CodigoExportador"),
		}
		d.TpViaTransp, _ = ParseTipoViaTransp(ini.LerString(secao, "tpViaTransp", ""))
		d.TpIntermedio, _ = ParseTipoIntermedio(ini.LerString(secao, "tpIntermedio", ""))

		lerINIProdutoAdicao(ini, idx, j, &d)
		p.DI = append(p.DI, d)
	}
}

// lerINIProdutoAdicao le [LADINNNJJJKKK] ou [adiNNNJJJKKK] -- indice TRIPLO.
// Porte de Ler_ProdutoADI.
//
// O nSeqAdi cai para o PROPRIO indice do laco quando ausente (:709). E o
// unico lugar do .ini em que um campo do documento assume o numero de ordem
// da secao como valor.
func lerINIProdutoAdicao(ini *pcn.INI, idxItem, idxDI int, d *DI) {
	for k := 1; ; k++ {
		secao := secaoAlias(ini,
			fmt.Sprintf("LADI%03d%03d%03d", idxItem, idxDI, k),
			fmt.Sprintf("adi%03d%03d%03d", idxItem, idxDI, k))

		adicao := lerStringAlias(ini, secao, "nAdicao", "NumeroAdicao")
		if adicao == "" || adicao == "FIM" {
			break
		}
		d.Adi = append(d.Adi, Adi{
			NAdicao:     lerInteiroAlias(ini, secao, 0, "nAdicao", "NumeroAdicao"),
			NSeqAdi:     ini.LerInteiro(secao, "nSeqAdi", k),
			CFabricante: lerStringAlias(ini, secao, "cFabricante", "CodigoFabricante"),
			VDescDI:     lerFloatAlias(ini, secao, "vDescDI", "DescontoADI"),
			NDraw:       ini.LerString(secao, "nDraw", ""),
		})
	}
}

// lerINIProdutoDetExport le [detExportNNNJJJ].
// Porte de Ler_ProdutoDetExport.
//
// A sentinela e uma CASCATA de tres chaves: `nRE`, caindo para `chNFe`,
// caindo para `nDraw`. Os tres grupos de exportacao do leiaute sao
// mutuamente exclusivos, e o item pode trazer qualquer um deles.
//
// NAO da para usar lerStringAlias aqui. Aquele porta a cadeia ANINHADA
// (`ReadString(a, ReadString(b, ”))`), que para na primeira chave
// PRESENTE, mesmo vazia. Esta aqui e uma cascata de `if` (:726-736), e cada
// degrau testa TAMBEM o comprimento:
//
//	sFim := ReadString('nRE','FIM');
//	if (sFim = 'FIM') or (Length(sFim) <= 0) then sFim := ReadString('chNFe','FIM');
//	if (sFim = 'FIM') or (Length(sFim) <= 0) then ...
//
// A diferenca nao e teorica: o PROPRIO gravador emite `nRe=` e `chNFe=`
// vazios sempre que o `nDraw` esta preenchido, que e o caso mais comum do
// grupo. Com a cadeia aninhada, a sentinela parava no `nRE` vazio e o item
// era descartado -- um .ini gerado aqui deixava de ser lido aqui.
func lerINIProdutoDetExport(ini *pcn.INI, idx int, p *Prod) {
	p.DetExport = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("detExport%03d%03d", idx, j)

		marca := ini.LerString(secao, "nRE", "FIM")
		if marca == "FIM" || marca == "" {
			marca = ini.LerString(secao, "chNFe", "FIM")
		}
		if marca == "FIM" || marca == "" {
			marca = ini.LerString(secao, "nDraw", "FIM")
		}
		if marca == "FIM" || marca == "" {
			break
		}
		p.DetExport = append(p.DetExport, DetExport{
			NDraw:   ini.LerString(secao, "nDraw", ""),
			NRE:     ini.LerString(secao, "nRE", ""),
			ChNFe:   ini.LerString(secao, "chNFe", ""),
			QExport: ini.LerFloat(secao, "qExport", 0),
		})
	}
}

// ---------------------------------------------------------------------------
// Mercadorias reguladas
// ---------------------------------------------------------------------------

// lerINIProdutoVeiculo le [VeiculoNNN] ou [veicProdNNN].
// Porte de Ler_ProdutoVeiculo. A sentinela e o CHASSI.
func lerINIProdutoVeiculo(ini *pcn.INI, idx int, v *VeicProd) {
	secao := secaoAlias(ini, fmt.Sprintf("Veiculo%03d", idx), fmt.Sprintf("veicProd%03d", idx))

	chassi := ini.LerString(secao, "Chassi", "FIM")
	if chassi == "FIM" || chassi == "" {
		return
	}

	v.Chassi = chassi
	v.TpOP, _ = ParseTipoOperacao(ini.LerString(secao, "tpOP", "0"))
	v.CCor = ini.LerString(secao, "cCor", "")
	v.XCor = ini.LerString(secao, "xCor", "")
	v.Pot = ini.LerString(secao, "pot", "")
	// Os aliases `CM3` e `CMKG` sao as unidades por extenso -- cilindrada em
	// centimetros cubicos e capacidade maxima de tracao em quilos.
	v.Cilin = lerStringAlias(ini, secao, "Cilin", "CM3")
	v.PesoL = ini.LerString(secao, "pesoL", "")
	v.PesoB = ini.LerString(secao, "pesoB", "")
	v.NSerie = ini.LerString(secao, "nSerie", "")
	v.TpComb = ini.LerString(secao, "tpComb", "")
	v.NMotor = ini.LerString(secao, "nMotor", "")
	v.CMT = lerStringAlias(ini, secao, "CMT", "CMKG")
	v.Dist = ini.LerString(secao, "dist", "")
	v.AnoMod = ini.LerInteiro(secao, "anoMod", 0)
	v.AnoFab = ini.LerInteiro(secao, "anoFab", 0)
	v.TpPint = ini.LerString(secao, "tpPint", "")
	v.TpVeic = ini.LerInteiro(secao, "tpVeic", 0)
	v.EspVeic = ini.LerInteiro(secao, "espVeic", 0)
	v.VIN = ini.LerString(secao, "VIN", "")
	v.CondVeic, _ = ParseCondicaoVeiculo(ini.LerString(secao, "condVeic", "1"))
	v.CMod = ini.LerString(secao, "cMod", "")
	v.CCorDENATRAN = ini.LerString(secao, "cCorDENATRAN", "")
	v.Lota = ini.LerInteiro(secao, "lota", 0)
	v.TpRest = ini.LerInteiro(secao, "tpRest", 0)
}

// lerINIProdutoMedicamento le [MedicamentoNNNJJJ] ou [medNNNJJJ].
// Porte de Ler_ProdutoMedicamento. A sentinela e o cProdANVISA.
func lerINIProdutoMedicamento(ini *pcn.INI, idx int, p *Prod) {
	p.Med = nil
	for j := 1; ; j++ {
		secao := secaoAlias(ini,
			fmt.Sprintf("Medicamento%03d%03d", idx, j),
			fmt.Sprintf("med%03d%03d", idx, j))

		anvisa := ini.LerString(secao, "cProdANVISA", "FIM")
		if anvisa == "FIM" || anvisa == "" {
			break
		}
		p.Med = append(p.Med, Med{
			CProdANVISA:    anvisa,
			NLote:          ini.LerString(secao, "nLote", ""),
			XMotivoIsencao: ini.LerString(secao, "xMotivoIsencao", ""),
			QLote:          ini.LerFloat(secao, "qLote", 0),
			DFab:           ini.LerData(secao, "dFab", tempoZero()),
			DVal:           ini.LerData(secao, "dVal", tempoZero()),
			VPMC:           ini.LerFloat(secao, "vPMC", 0),
		})
	}
}

// lerINIProdutoArma le [ArmaNNNJJJ]. Porte de Ler_ProdutoArma.
func lerINIProdutoArma(ini *pcn.INI, idx int, p *Prod) {
	p.Arma = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("Arma%03d%03d", idx, j)

		serie := ini.LerString(secao, "nSerie", "FIM")
		if serie == "FIM" || serie == "" {
			break
		}
		item := Arma{NSerie: serie}
		item.TpArma, _ = ParseTipoArma(ini.LerString(secao, "tpArma", "0"))
		item.NCano = ini.LerString(secao, "nCano", "")
		item.Descr = ini.LerString(secao, "descr", "")
		p.Arma = append(p.Arma, item)
	}
}

// lerINIProdutoCombustivel le [CombustivelNNN] ou [combNNN] e os seis
// subgrupos. Porte de Ler_ProdutoCombustivel.
//
// Os subgrupos CIDE, encerrante, ICMSComb, ICMSInter e ICMSCons tem secao
// PROPRIA no .ini -- nao ficam dentro da secao do combustivel --, e cada um
// indexa so pelo item.
func lerINIProdutoCombustivel(ini *pcn.INI, idx int, c *Comb) {
	secao := secaoAlias(ini,
		fmt.Sprintf("Combustivel%03d", idx), fmt.Sprintf("comb%03d", idx))

	anp := ini.LerString(secao, "cProdANP", "FIM")
	if anp == "FIM" || anp == "" {
		return
	}

	c.CProdANP = ini.LerInteiro(secao, "cProdANP", 0)
	c.PMixGN = ini.LerFloat(secao, "pMixGN", 0)
	c.DescANP = ini.LerString(secao, "descANP", "")
	c.PGLP = ini.LerFloat(secao, "pGLP", 0)
	c.PGNn = ini.LerFloat(secao, "pGNn", 0)
	c.PGNi = ini.LerFloat(secao, "pGNi", 0)
	c.VPart = ini.LerFloat(secao, "vPart", 0)
	c.CODIF = ini.LerString(secao, "CODIF", "")
	c.QTemp = ini.LerFloat(secao, "qTemp", 0)
	c.UFCons = ini.LerString(secao, "UFCons", "")
	c.PBio = ini.LerFloat(secao, "pBio", 0)

	lerINICombustivelOrigem(ini, idx, c)

	cide := fmt.Sprintf("CIDE%03d", idx)
	c.CIDE.QBCProd = ini.LerFloat(cide, "qBCprod", 0)
	c.CIDE.VAliqProd = ini.LerFloat(cide, "vAliqProd", 0)
	c.CIDE.VCIDE = ini.LerFloat(cide, "vCIDE", 0)

	enc := fmt.Sprintf("encerrante%03d", idx)
	c.Encerrante.NBico = ini.LerInteiro(enc, "nBico", 0)
	c.Encerrante.NBomba = ini.LerInteiro(enc, "nBomba", 0)
	c.Encerrante.NTanque = ini.LerInteiro(enc, "nTanque", 0)
	c.Encerrante.VEncIni = ini.LerFloat(enc, "vEncIni", 0)
	c.Encerrante.VEncFin = ini.LerFloat(enc, "vEncFin", 0)

	icms := fmt.Sprintf("ICMSComb%03d", idx)
	c.ICMS.VBCICMS = ini.LerFloat(icms, "vBCICMS", 0)
	c.ICMS.VICMS = ini.LerFloat(icms, "vICMS", 0)
	c.ICMS.VBCICMSST = ini.LerFloat(icms, "vBCICMSST", 0)
	c.ICMS.VICMSST = ini.LerFloat(icms, "vICMSST", 0)

	// Os dois ultimos sao CONDICIONADOS a propria base de calculo, e os
	// unicos subgrupos do combustivel que podem nao existir.
	inter := fmt.Sprintf("ICMSInter%03d", idx)
	if v := ini.LerString(inter, "vBCICMSSTDest", "FIM"); v != "FIM" && v != "" {
		c.ICMSInter.VBCICMSSTDest = ini.LerFloat(inter, "vBCICMSSTDest", 0)
		c.ICMSInter.VICMSSTDest = ini.LerFloat(inter, "vICMSSTDest", 0)
	}

	cons := fmt.Sprintf("ICMSCons%03d", idx)
	if v := ini.LerString(cons, "vBCICMSSTCons", "FIM"); v != "FIM" && v != "" {
		c.ICMSCons.VBCICMSSTCons = ini.LerFloat(cons, "vBCICMSSTCons", 0)
		c.ICMSCons.VICMSSTCons = ini.LerFloat(cons, "vICMSSTCons", 0)
		c.ICMSCons.UFCons = ini.LerString(cons, "UFCons", "")
	}
}

// lerINICombustivelOrigem le [origCombNNNJJ] -- indice do item com TRES
// digitos e da origem com DOIS. Porte de Ler_ProdutoCombustivelOrigemComb.
func lerINICombustivelOrigem(ini *pcn.INI, idx int, c *Comb) {
	c.OrigComb = nil
	for j := 1; ; j++ {
		secao := fmt.Sprintf("origComb%03d%02d", idx, j)

		uf := ini.LerString(secao, "cUFOrig", "FIM")
		if uf == "FIM" || uf == "" {
			break
		}
		item := OrigComb{
			CUFOrig: ini.LerInteiro(secao, "cUFOrig", 0),
			POrig:   ini.LerFloat(secao, "pOrig", 0),
		}
		item.IndImport, _ = ParseIndImport(ini.LerString(secao, "indImport", "0"))
		c.OrigComb = append(c.OrigComb, item)
	}
}

// ---------------------------------------------------------------------------
// Impostos do item
// ---------------------------------------------------------------------------

// lerINIProdutoImposto le os onze grupos de imposto do item.
// Porte de Ler_ProdutoImposto.
func lerINIProdutoImposto(ini *pcn.INI, idx int, imp *Imposto) {
	lerINIImpostoICMS(ini, idx, &imp.ICMS)
	lerINIImpostoICMSUFDest(ini, idx, &imp.ICMSUFDest)
	lerINIImpostoIPI(ini, idx, &imp.IPI)
	lerINIImpostoII(ini, idx, &imp.II)
	lerINIImpostoPIS(ini, idx, &imp.PIS)
	lerINIImpostoPISST(ini, idx, &imp.PISST)
	lerINIImpostoCOFINS(ini, idx, &imp.COFINS)
	lerINIImpostoCOFINSST(ini, idx, &imp.COFINSST)
	lerINIImpostoISSQN(ini, idx, &imp.ISSQN)

	rtc.LerINIISel(ini, &imp.ISel, idx)
	rtc.LerINIIBSCBS(ini, &imp.IBSCBS, idx, rtc.SemIndice)
}

// lerINIImpostoICMS le [ICMSNNN]. Porte de Ler_ImpostoICMS.
//
// A sentinela e o `CST`, caindo para o `CSOSN`: o item de um emitente do
// Simples Nacional nao tem CST, tem CSOSN. Testar so um dos dois deixa
// metade das notas sem ICMS.
//
// Note que o CST tem default "00" na leitura do campo (:1001) mas "FIM" na
// sentinela: um .ini com `CSOSN=101` e sem CST entra no grupo e sai com
// CST=00 preenchido. E o que o original faz.
func lerINIImpostoICMS(ini *pcn.INI, idx int, v *ICMS) {
	secao := fmt.Sprintf("ICMS%03d", idx)

	marca := ini.LerString(secao, "CST", "FIM")
	if marca == "FIM" || marca == "" {
		marca = ini.LerString(secao, "CSOSN", "FIM")
	}
	if marca == "FIM" || marca == "" {
		return
	}

	v.Orig, _ = pcn.ParseOrigemMercadoria(primeiroNaoVazio(
		lerStringAlias(ini, secao, "orig", "Origem"), "0"))
	v.CST, _ = pcn.ParseCSTIcms(ini.LerString(secao, "CST", "00"))
	v.CSOSN, _ = pcn.ParseCSOSNIcms(ini.LerString(secao, "CSOSN", ""))
	v.ModBC, _ = ParseDeterminacaoBaseIcms(primeiroNaoVazio(
		lerStringAlias(ini, secao, "modBC", "Modalidade"), "0"))
	v.PRedBC = lerFloatAlias(ini, secao, "pRedBC", "PercentualReducao")
	v.CBenefRBC = ini.LerString(secao, "cBenefRBC", "")
	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.PICMS = lerFloatAlias(ini, secao, "pICMS", "Aliquota")
	v.VICMS = lerFloatAlias(ini, secao, "vICMS", "Valor")
	v.VBCFCP = lerFloatAlias(ini, secao, "vBCFCP", "ValorBaseFCP")
	v.PFCP = lerFloatAlias(ini, secao, "pFCP", "PercentualFCP")
	v.VFCP = lerFloatAlias(ini, secao, "vFCP", "ValorFCP")
	v.ModBCST, _ = ParseDeterminacaoBaseIcmsST(primeiroNaoVazio(
		lerStringAlias(ini, secao, "modBCST", "ModalidadeST"), "0"))
	v.PMVAST = lerFloatAlias(ini, secao, "pMVAST", "PercentualMargemST")
	v.PRedBCST = lerFloatAlias(ini, secao, "pRedBCST", "PercentualReducaoST")
	v.VBCST = lerFloatAlias(ini, secao, "vBCST", "ValorBaseST")
	v.PICMSST = lerFloatAlias(ini, secao, "pICMSST", "AliquotaST")
	v.VICMSST = lerFloatAlias(ini, secao, "vICMSST", "ValorST")
	v.VBCFCPST = lerFloatAlias(ini, secao, "vBCFCPST", "ValorBaseFCPST")
	v.PFCPST = lerFloatAlias(ini, secao, "pFCPST", "PercentualFCPST")
	v.VFCPST = lerFloatAlias(ini, secao, "vFCPST", "ValorFCPST")
	v.UFST = ini.LerString(secao, "UFST", "")
	v.PBCOp = ini.LerFloat(secao, "pBCOp", 0)
	v.VBCSTRet = ini.LerFloat(secao, "vBCSTRet", 0)
	v.PST = ini.LerFloat(secao, "pST", 0)
	v.VICMSSTRet = ini.LerFloat(secao, "vICMSSTRet", 0)
	// O alias do vBCFCPSTRet e "ValorBaseFCPSTRes" -- com RES no fim, nao
	// RET. E assim no original (:1026), e um .ini antigo traz a chave com o
	// erro de digitacao.
	v.VBCFCPSTRet = lerFloatAlias(ini, secao, "vBCFCPSTRet", "ValorBaseFCPSTRes")
	v.PFCPSTRet = lerFloatAlias(ini, secao, "pFCPSTRet", "PercentualFCPSTRet")
	v.VFCPSTRet = lerFloatAlias(ini, secao, "vFCPSTRet", "ValorFCPSTRet")
	v.MotDesICMS, _ = ParseMotivoDesoneracaoICMS(ini.LerString(secao, "motDesICMS", "0"))
	v.PCredSN = ini.LerFloat(secao, "pCredSN", 0)
	v.VCredICMSSN = ini.LerFloat(secao, "vCredICMSSN", 0)
	v.VBCSTDest = ini.LerFloat(secao, "vBCSTDest", 0)
	v.VICMSSTDest = ini.LerFloat(secao, "vICMSSTDest", 0)
	v.VICMSDeson = ini.LerFloat(secao, "vICMSDeson", 0)
	v.VICMSOp = ini.LerFloat(secao, "vICMSOp", 0)
	v.PDif = ini.LerFloat(secao, "pDif", 0)
	v.VICMSDif = ini.LerFloat(secao, "vICMSDif", 0)

	v.PRedBCEfet = ini.LerFloat(secao, "pRedBCEfet", 0)
	v.VBCEfet = ini.LerFloat(secao, "vBCEfet", 0)
	v.PICMSEfet = ini.LerFloat(secao, "pICMSEfet", 0)
	v.VICMSEfet = ini.LerFloat(secao, "vICMSEfet", 0)
	v.VICMSSubstituto = ini.LerFloat(secao, "vICMSSubstituto", 0)
	v.VICMSSTDeson = ini.LerFloat(secao, "vICMSSTDeson", 0)
	v.MotDesICMSST, _ = ParseMotivoDesoneracaoICMS(ini.LerString(secao, "motDesICMSST", "0"))
	v.PFCPDif = ini.LerFloat(secao, "pFCPDif", 0)
	v.VFCPDif = ini.LerFloat(secao, "vFCPDif", 0)
	v.VFCPEfet = ini.LerFloat(secao, "vFCPEfet", 0)

	// Monofasia do ICMS (CST 02, 15, 53 e 61).
	v.AdRemICMS = ini.LerFloat(secao, "adRemICMS", 0)
	v.VICMSMono = ini.LerFloat(secao, "vICMSMono", 0)
	v.AdRemICMSReten = ini.LerFloat(secao, "adRemICMSReten", 0)
	v.VICMSMonoReten = ini.LerFloat(secao, "vICMSMonoReten", 0)
	v.VICMSMonoDif = ini.LerFloat(secao, "vICMSMonoDif", 0)
	v.AdRemICMSRet = ini.LerFloat(secao, "adRemICMSRet", 0)
	v.VICMSMonoRet = ini.LerFloat(secao, "vICMSMonoRet", 0)
	v.QBCMono = ini.LerFloat(secao, "qBCMono", 0)
	v.QBCMonoReten = ini.LerFloat(secao, "qBCMonoReten", 0)
	v.PRedAdRem = ini.LerFloat(secao, "pRedAdRem", 0)
	v.MotRedAdRem, _ = ParseMotRedAdRem(ini.LerString(secao, "motRedAdRem", "0"))
	v.QBCMonoRet = ini.LerFloat(secao, "qBCMonoRet", 0)
	v.VICMSMonoOp = ini.LerFloat(secao, "vICMSMonoOp", 0)
	v.IndDeduzDeson, _ = pcn.ParseIndicadorEx(ini.LerString(secao, "indDeduzDeson", ""))
}

// lerINIImpostoICMSUFDest le [ICMSUFDestNNN].
// Porte de Ler_ImpostoICMSUFDest. A sentinela e a base de calculo.
func lerINIImpostoICMSUFDest(ini *pcn.INI, idx int, v *ICMSUFDest) {
	secao := fmt.Sprintf("ICMSUFDest%03d", idx)

	base := ini.LerString(secao, "vBCUFDest", "FIM")
	if base == "FIM" || base == "" {
		return
	}
	v.VBCUFDest = ini.LerFloat(secao, "vBCUFDest", 0)
	v.VBCFCPUFDest = ini.LerFloat(secao, "vBCFCPUFDest", 0)
	v.PFCPUFDest = ini.LerFloat(secao, "pFCPUFDest", 0)
	v.PICMSUFDest = ini.LerFloat(secao, "pICMSUFDest", 0)
	v.PICMSInter = ini.LerFloat(secao, "pICMSInter", 0)
	v.PICMSInterPart = ini.LerFloat(secao, "pICMSInterPart", 0)
	v.VFCPUFDest = ini.LerFloat(secao, "vFCPUFDest", 0)
	v.VICMSUFDest = ini.LerFloat(secao, "vICMSUFDest", 0)
	v.VICMSUFRemet = ini.LerFloat(secao, "vICMSUFRemet", 0)
}

// lerINIImpostoIPI le [IPINNN]. Porte de Ler_ImpostoIPI.
//
// Tem uma condicao A MAIS que os demais: alem de o CST existir, ele precisa
// CONVERTER (`if OK` em :1109). CST invalido deixa o grupo inteiro vazio --
// nem os campos que nao dependem dele sao lidos.
func lerINIImpostoIPI(ini *pcn.INI, idx int, v *IPI) {
	secao := fmt.Sprintf("IPI%03d", idx)

	marca := ini.LerString(secao, "CST", "FIM")
	if marca == "FIM" || marca == "" {
		return
	}

	cst, err := ParseCstIpi(ini.LerString(secao, "CST", ""))
	if err != nil {
		return
	}
	v.CST = cst

	v.ClEnq = lerStringAlias(ini, secao, "clEnq", "ClasseEnquadramento")
	v.CNPJProd = lerStringAlias(ini, secao, "CNPJProd", "CNPJProdutor")
	v.CSelo = lerStringAlias(ini, secao, "cSelo", "CodigoSeloIPI")
	v.QSelo = lerInteiroAlias(ini, secao, 0, "qSelo", "QuantidadeSelos")
	v.CEnq = lerStringAlias(ini, secao, "cEnq", "CodigoEnquadramento")
	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.QUnid = lerFloatAlias(ini, secao, "qUnid", "Quantidade")
	v.VUnid = lerFloatAlias(ini, secao, "vUnid", "ValorUnidade")
	v.PIPI = lerFloatAlias(ini, secao, "pIPI", "Aliquota")
	v.VIPI = lerFloatAlias(ini, secao, "vIPI", "Valor")
}

// lerINIImpostoII le [IINNN]. Porte de Ler_ImpostoII.
func lerINIImpostoII(ini *pcn.INI, idx int, v *II) {
	secao := fmt.Sprintf("II%03d", idx)

	base := lerStringAlias(ini, secao, "vBC", "ValorBase")
	if base == "" || base == "FIM" {
		return
	}
	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.VDespAdu = lerFloatAlias(ini, secao, "vDespAdu", "ValorDespAduaneiras")
	v.VII = lerFloatAlias(ini, secao, "vII", "ValorII")
	v.VIOF = lerFloatAlias(ini, secao, "vIOF", "ValorIOF")
}

// lerINIImpostoPIS le [PISNNN]. Porte de Ler_ImpostoPIS.
func lerINIImpostoPIS(ini *pcn.INI, idx int, v *PIS) {
	secao := fmt.Sprintf("PIS%03d", idx)

	marca := ini.LerString(secao, "CST", "FIM")
	if marca == "FIM" || marca == "" {
		return
	}
	v.CST, _ = pcn.ParseCSTPis(ini.LerString(secao, "CST", ""))
	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.PPIS = lerFloatAlias(ini, secao, "pPIS", "Aliquota")
	v.QBCProd = lerFloatAlias(ini, secao, "qBCProd", "Quantidade")
	v.VAliqProd = lerFloatAlias(ini, secao, "vAliqProd", "ValorAliquota")
	v.VPIS = lerFloatAlias(ini, secao, "vPIS", "Valor")
}

// lerINIImpostoPISST le [PISSTNNN]. Porte de Ler_ImpostoPISST.
//
// A sentinela e a mais exotica do formato: o original CONCATENA duas
// leituras cujos defaults sao "F" e "IM" e compara o resultado com "FIM"
// (:1163-1165). Faltando as duas chaves, a concatenacao da exatamente
// "FIM" e o grupo nao e lido; havendo qualquer conteudo, nao da.
//
// A ordem dos dois degraus tambem e contrato: primeiro os nomes por
// extenso, depois os do leiaute.
func lerINIImpostoPISST(ini *pcn.INI, idx int, v *PISST) {
	secao := fmt.Sprintf("PISST%03d", idx)

	// A concatenacao esta escrita como no original (:1163-1165) de
	// proposito. Descreve-la em prosa perde um caso: chave PRESENTE E
	// VAZIA com a outra AUSENTE da "" + "IM" = "IM", que nao e "FIM" e tem
	// comprimento maior que zero -- o grupo E lido. Com as duas presentes e
	// vazias da "", comprimento zero, e NAO e lido.
	marca := ini.LerString(secao, "ValorBase", "F") + ini.LerString(secao, "Quantidade", "IM")
	if marca == "FIM" {
		marca = ini.LerString(secao, "vBC", "F") + ini.LerString(secao, "qBCProd", "IM")
	}
	if marca == "FIM" || marca == "" {
		return
	}

	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.PPIS = lerFloatAlias(ini, secao, "pPis", "AliquotaPerc")
	v.QBCProd = lerFloatAlias(ini, secao, "qBCProd", "Quantidade")
	v.VAliqProd = lerFloatAlias(ini, secao, "vAliqProd", "AliquotaValor")
	v.VPIS = lerFloatAlias(ini, secao, "vPIS", "ValorPISST")
	v.IndSomaPISST, _ = ParseIndSomaPISST(ini.LerString(secao, "indSomaPISST", ""))
}

// lerINIImpostoCOFINS le [COFINSNNN]. Porte de Ler_ImpostoCOFINS.
func lerINIImpostoCOFINS(ini *pcn.INI, idx int, v *COFINS) {
	secao := fmt.Sprintf("COFINS%03d", idx)

	marca := ini.LerString(secao, "CST", "FIM")
	if marca == "FIM" || marca == "" {
		return
	}
	v.CST, _ = pcn.ParseCSTCofins(ini.LerString(secao, "CST", ""))
	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.PCOFINS = lerFloatAlias(ini, secao, "pCOFINS", "Aliquota")
	v.QBCProd = lerFloatAlias(ini, secao, "qBCProd", "Quantidade")
	v.VAliqProd = lerFloatAlias(ini, secao, "vAliqProd", "ValorAliquota")
	v.VCOFINS = lerFloatAlias(ini, secao, "vCOFINS", "Valor")
}

// lerINIImpostoCOFINSST le [COFINSSTNNN]. Porte de Ler_ImpostoCOFINSST.
// A sentinela e a mesma concatenacao do PISST -- ver o comentario la.
func lerINIImpostoCOFINSST(ini *pcn.INI, idx int, v *COFINSST) {
	secao := fmt.Sprintf("COFINSST%03d", idx)

	// A concatenacao esta escrita como no original (:1163-1165) de
	// proposito. Descreve-la em prosa perde um caso: chave PRESENTE E
	// VAZIA com a outra AUSENTE da "" + "IM" = "IM", que nao e "FIM" e tem
	// comprimento maior que zero -- o grupo E lido. Com as duas presentes e
	// vazias da "", comprimento zero, e NAO e lido.
	marca := ini.LerString(secao, "ValorBase", "F") + ini.LerString(secao, "Quantidade", "IM")
	if marca == "FIM" {
		marca = ini.LerString(secao, "vBC", "F") + ini.LerString(secao, "qBCProd", "IM")
	}
	if marca == "FIM" || marca == "" {
		return
	}

	v.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBase")
	v.PCOFINS = lerFloatAlias(ini, secao, "pCOFINS", "AliquotaPerc")
	v.QBCProd = lerFloatAlias(ini, secao, "qBCProd", "Quantidade")
	v.VAliqProd = lerFloatAlias(ini, secao, "vAliqProd", "AliquotaValor")
	v.VCOFINS = lerFloatAlias(ini, secao, "vCOFINS", "ValorCOFINSST")
	v.IndSomaCOFINSST, _ = ParseIndSomaCOFINSST(ini.LerString(secao, "indSomaCOFINSST", ""))
}

// lerINIImpostoISSQN le [ISSQNNNN]. Porte de Ler_ImpostoISSQN.
//
// Unico grupo do item com condicao de VALOR, e nao so de presenca: a base
// de calculo precisa ser POSITIVA (:1240). Um servico com vBC zero nao
// entra no .ini -- e tambem nao deveria entrar na nota.
func lerINIImpostoISSQN(ini *pcn.INI, idx int, v *ISSQN) {
	secao := fmt.Sprintf("ISSQN%03d", idx)

	base := lerStringAlias(ini, secao, "ValorBase", "vBC")
	if base == "" || base == "FIM" {
		return
	}
	if vbc := ini.LerFloat(secao, "ValorBase", ini.LerFloat(secao, "vBC", 0)); vbc <= 0 {
		return
	} else {
		v.VBC = vbc
	}

	v.VAliq = lerFloatAlias(ini, secao, "vAliq", "Aliquota")
	v.VISSQN = lerFloatAlias(ini, secao, "vISSQN", "ValorISSQN")
	v.CMunFG = lerInteiroAlias(ini, secao, 0, "cMunFG", "MunicipioFatoGerador")
	v.CListServ = lerStringAlias(ini, secao, "cListServ", "CodigoServico")
	v.CSitTrib, _ = ParseISSQNcSitTrib(ini.LerString(secao, "cSitTrib", ""))
	v.VDeducao = lerFloatAlias(ini, secao, "vDeducao", "ValorDeducao")
	v.VOutro = lerFloatAlias(ini, secao, "vOutro", "ValorOutro")
	v.VDescIncond = lerFloatAlias(ini, secao, "vDescIncond", "ValorDescontoIncondicional")
	v.VDescCond = lerFloatAlias(ini, secao, "vDescCond", "vDescontoCondicional")
	v.VISSRet = lerFloatAlias(ini, secao, "vISSRet", "ValorISSRetido")
	v.IndISS, _ = ParseIndISS(ini.LerString(secao, "indISS", ""))
	v.CServico = ini.LerString(secao, "cServico", "")
	v.CMun = ini.LerInteiro(secao, "cMun", 0)
	v.CPais = ini.LerInteiro(secao, "cPais", 1058)
	v.NProcesso = ini.LerString(secao, "nProcesso", "")
	v.IndIncentivo, _ = pcn.ParseIndIncentivo(ini.LerString(secao, "indIncentivo", ""))
}
