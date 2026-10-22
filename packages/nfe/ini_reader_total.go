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

// Leitura .ini dos grupos que vem DEPOIS dos itens: totais, transporte,
// cobranca, pagamento e informacoes adicionais.
// Porte das procedures Ler_Total em diante de ACBrNFe.IniReader.pas.
//
// Esta metade do arquivo e onde os ALIAS DE CHAVE em portugues se
// concentram -- `vBC`/`BaseICMS`, `vProd`/`ValorProduto`,
// `modFrete`/`FretePorConta`, `nDup`/`Numero`. Sao o formato que os ERPs
// mais antigos geram, e continuam sendo aceitos.

// lerFloatAlias le a cadeia aninhada de leituras numericas.
//
// Segue a semantica de lerStringAlias, e NAO a de lerInteiroAlias: para na
// primeira chave PRESENTE, mesmo vazia. A diferenca vem do Delphi, onde o
// numero com casas decimais nao tem um ReadFloat encadeado --
//
//	StringToFloatDef(ReadString(s, 'vBC', ReadString(s, 'BaseICMS', '')), 0)
//
// -- e o que encadeia sao os ReadString. Com `vBC=` presente e vazio, o
// ReadString externo devolve "" (a chave existe), o alias nem e consultado,
// e o StringToFloatDef entrega ZERO.
//
// E o oposto do inteiro, onde o encadeamento acontece no DEFAULT do
// ReadInteger e por isso cascateia. As duas formas convivem no mesmo
// arquivo do original.
func lerFloatAlias(ini *pcn.INI, secao string, chaves ...string) float64 {
	const ausente = "\x00"
	for _, chave := range chaves {
		if ini.LerString(secao, chave, ausente) != ausente {
			return ini.LerFloat(secao, chave, 0)
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// Totais
// ---------------------------------------------------------------------------

// lerINITotal le [Total], [ISSQNtot] e [retTrib].
// Porte de Ler_Total, Ler_TotalISSQNTot e Ler_TotalRetTributavel.
//
// As tres procedures do original sao chamadas em sequencia pelo LerIni e
// alimentam partes do mesmo TTotal; aqui ficam numa funcao so, com a mesma
// ordem.
func lerINITotal(ini *pcn.INI, t *Total) {
	const secao = "Total"
	tot := &t.ICMSTot

	tot.VBC = lerFloatAlias(ini, secao, "vBC", "BaseICMS")
	tot.VICMS = lerFloatAlias(ini, secao, "vICMS", "ValorICMS")
	tot.VICMSDeson = ini.LerFloat(secao, "vICMSDeson", 0)
	tot.VFCP = lerFloatAlias(ini, secao, "vFCP", "ValorFCP")
	tot.VBCST = lerFloatAlias(ini, secao, "vBCST", "BaseICMSSubstituicao")
	tot.VST = lerFloatAlias(ini, secao, "vST", "ValorICMSSubstituicao")
	tot.VFCPST = lerFloatAlias(ini, secao, "vFCPST", "ValorFCPST")
	tot.VFCPSTRet = lerFloatAlias(ini, secao, "vFCPSTRet", "ValorFCPSTRet")

	// Monofasia do ICMS -- nao confundir com a monofasia da Reforma
	// (gIBSCBSMono), que vive em outro grupo.
	tot.QBCMono = ini.LerFloat(secao, "qBCMono", 0)
	tot.VICMSMono = ini.LerFloat(secao, "vICMSMono", 0)
	tot.QBCMonoReten = ini.LerFloat(secao, "qBCMonoReten", 0)
	tot.VICMSMonoReten = ini.LerFloat(secao, "vICMSMonoReten", 0)
	tot.QBCMonoRet = ini.LerFloat(secao, "qBCMonoRet", 0)
	tot.VICMSMonoRet = ini.LerFloat(secao, "vICMSMonoRet", 0)

	tot.VProd = lerFloatAlias(ini, secao, "vProd", "ValorProduto")
	tot.VFrete = lerFloatAlias(ini, secao, "vFrete", "ValorFrete")
	tot.VSeg = lerFloatAlias(ini, secao, "vSeg", "ValorSeguro")
	tot.VDesc = lerFloatAlias(ini, secao, "vDesc", "ValorDesconto")
	tot.VII = lerFloatAlias(ini, secao, "vII", "ValorII")
	tot.VIPI = lerFloatAlias(ini, secao, "vIPI", "ValorIPI")
	tot.VIPIDevol = lerFloatAlias(ini, secao, "vIPIDevol", "ValorIPIDevol")
	tot.VPIS = lerFloatAlias(ini, secao, "vPIS", "ValorPIS")
	tot.VCOFINS = lerFloatAlias(ini, secao, "vCOFINS", "ValorCOFINS")
	tot.VOutro = lerFloatAlias(ini, secao, "vOutro", "ValorOutrasDespesas")
	tot.VNF = lerFloatAlias(ini, secao, "vNF", "ValorNota")
	tot.VTotTrib = ini.LerFloat(secao, "vTotTrib", 0)
	tot.VFCPUFDest = ini.LerFloat(secao, "vFCPUFDest", 0)
	tot.VICMSUFDest = ini.LerFloat(secao, "vICMSUFDest", 0)
	tot.VICMSUFRemet = ini.LerFloat(secao, "vICMSUFRemet", 0)

	// vNFTot pertence ao TTotal, nao ao ICMSTot -- e o total da nota com os
	// tributos da Reforma somados.
	t.VNFTot = ini.LerFloat(secao, "vNFTot", 0)

	lerINITotalISSQN(ini, &t.ISSQNTot)
	lerINITotalRetTributavel(ini, &t.RetTrib)

	rtc.LerINIISTot(ini, &t.ISTot)
	rtc.LerINIIBSCBSTot(ini, &t.IBSCBSTot)
}

// lerINITotalISSQN le [ISSQNtot]. Porte de Ler_TotalISSQNTot.
func lerINITotalISSQN(ini *pcn.INI, t *ISSQNTot) {
	const secao = "ISSQNtot"

	t.VServ = lerFloatAlias(ini, secao, "vServ", "ValorServicos")
	t.VBC = lerFloatAlias(ini, secao, "vBC", "ValorBaseISS")
	t.VISS = lerFloatAlias(ini, secao, "vISS", "ValorISSQN")
	t.VPIS = lerFloatAlias(ini, secao, "vPIS", "ValorPISISS")
	// O alias do COFINS do ISSQN e "ValorCONFINSISS" -- com o N trocado de
	// lugar. E assim no original (:1502), e um .ini gerado pelo ACBr traz a
	// chave com o erro de digitacao.
	t.VCOFINS = lerFloatAlias(ini, secao, "vCOFINS", "ValorCONFINSISS")
	t.DCompet = ini.LerData(secao, "dCompet", tempoZero())
	t.VDeducao = ini.LerFloat(secao, "vDeducao", 0)
	t.VOutro = ini.LerFloat(secao, "vOutro", 0)
	t.VDescIncond = ini.LerFloat(secao, "vDescIncond", 0)
	t.VDescCond = ini.LerFloat(secao, "vDescCond", 0)
	t.VISSRet = ini.LerFloat(secao, "vISSRet", 0)

	// Chave PRESENTE E VAZIA cai para "0", e nao para o membro vazio do
	// enum: o original reatribui aValor quando ele vem vazio (:1511-1513).
	regime := ini.LerString(secao, "cRegTrib", "0")
	if regime == "" {
		regime = "0"
	}
	t.CRegTrib, _ = pcn.ParseRegTribISSQN(regime)
}

// lerINITotalRetTributavel le [retTrib]. Porte de Ler_TotalRetTributavel.
func lerINITotalRetTributavel(ini *pcn.INI, r *RetTrib) {
	const secao = "retTrib"
	r.VRetPIS = ini.LerFloat(secao, "vRetPIS", 0)
	r.VRetCOFINS = ini.LerFloat(secao, "vRetCOFINS", 0)
	r.VRetCSLL = ini.LerFloat(secao, "vRetCSLL", 0)
	r.VBCIRRF = ini.LerFloat(secao, "vBCIRRF", 0)
	r.VIRRF = ini.LerFloat(secao, "vIRRF", 0)
	r.VBCRetPrev = ini.LerFloat(secao, "vBCRetPrev", 0)
	r.VRetPrev = ini.LerFloat(secao, "vRetPrev", 0)
}

// ---------------------------------------------------------------------------
// Transporte
// ---------------------------------------------------------------------------

// lerINITransportador le [Transportador] ou [transp], mais os reboques e os
// volumes. Porte de Ler_Transportador.
func lerINITransportador(ini *pcn.INI, t *Transp) {
	secao := secaoAlias(ini, "Transportador", "transp")

	t.ModFrete, _ = ParseModalidadeFrete(primeiroNaoVazio(
		lerStringAlias(ini, secao, "modFrete", "FretePorConta"), "0"))

	t.Transporta.CNPJCPF = ini.LerString(secao, "CNPJCPF", "")
	t.Transporta.XNome = lerStringAlias(ini, secao, "xNome", "NomeRazao")
	t.Transporta.IE = ini.LerString(secao, "IE", "")
	t.Transporta.XEnder = lerStringAlias(ini, secao, "xEnder", "Endereco")
	t.Transporta.XMun = lerStringAlias(ini, secao, "xMun", "Cidade")
	t.Transporta.UF = ini.LerString(secao, "UF", "")

	t.RetTransp.VServ = lerFloatAlias(ini, secao, "vServ", "ValorServico")
	t.RetTransp.VBCRet = lerFloatAlias(ini, secao, "vBCRet", "ValorBase")
	t.RetTransp.PICMSRet = lerFloatAlias(ini, secao, "pICMSRet", "Aliquota")
	t.RetTransp.VICMSRet = lerFloatAlias(ini, secao, "vICMSRet", "Valor")
	t.RetTransp.CFOP = ini.LerString(secao, "CFOP", "")
	t.RetTransp.CMunFG = lerInteiroAlias(ini, secao, 0, "cMunFG", "CidadeCod")

	// As chaves do veiculo NAO seguem o nome da tag: sao `Placa`, `UFPlaca`
	// e `RNTC`, na secao do transporte, e nao num grupo proprio.
	t.VeicTransp.Placa = ini.LerString(secao, "Placa", "")
	t.VeicTransp.UF = ini.LerString(secao, "UFPlaca", "")
	t.VeicTransp.RNTC = ini.LerString(secao, "RNTC", "")

	t.Vagao = ini.LerString(secao, "vagao", "")
	t.Balsa = ini.LerString(secao, "balsa", "")

	lerINIReboques(ini, t)
	lerINIVolumes(ini, t)
}

// lerINIReboques le [ReboqueNNN]. A placa e a sentinela.
func lerINIReboques(ini *pcn.INI, t *Transp) {
	t.Reboque = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("Reboque%03d", i)
		placa := ini.LerString(secao, "placa", "FIM")
		if placa == "FIM" || placa == "" {
			break
		}
		t.Reboque = append(t.Reboque, Reboque{
			Placa: placa,
			UF:    ini.LerString(secao, "UF", ""),
			RNTC:  ini.LerString(secao, "RNTC", ""),
		})
	}
}

// lerINIVolumes le [VolumeNNN] ou [volNNN], com os lacres aninhados.
//
// O criterio de fim aqui NAO e uma sentinela: o laco para quando TODOS os
// seis campos do volume estao vazios (:1419-1424). Um volume que traga so o
// peso, sem especie nem quantidade, continua sendo volume.
func lerINIVolumes(ini *pcn.INI, t *Transp) {
	t.Vol = nil
	for i := 1; ; i++ {
		secao := secaoAlias(ini, fmt.Sprintf("Volume%03d", i), fmt.Sprintf("vol%03d", i))

		qVol := lerStringAlias(ini, secao, "qVol", "Quantidade")
		esp := lerStringAlias(ini, secao, "esp", "Especie")
		marca := ini.LerString(secao, "Marca", "")
		nVol := lerStringAlias(ini, secao, "nVol", "Numeracao")
		pesoL := lerStringAlias(ini, secao, "pesoL", "PesoLiquido")
		pesoB := lerStringAlias(ini, secao, "pesoB", "PesoBruto")

		if qVol == "" && esp == "" && marca == "" && nVol == "" &&
			pesoL == "" && pesoB == "" {
			break
		}

		v := Vol{
			QVol:  lerInteiroAlias(ini, secao, 0, "qVol", "Quantidade"),
			Esp:   esp,
			Marca: marca,
			NVol:  nVol,
			PesoL: lerFloatAlias(ini, secao, "pesoL", "PesoLiquido"),
			PesoB: lerFloatAlias(ini, secao, "pesoB", "PesoBruto"),
		}
		lerINILacres(ini, i, &v)
		t.Vol = append(t.Vol, v)
	}
}

// lerINILacres le [lacresIIIJJJ] ou [LacreIIIJJJ] -- o indice e DUPLO: tres
// digitos do volume seguidos de tres do lacre.
func lerINILacres(ini *pcn.INI, indiceVolume int, v *Vol) {
	for j := 1; ; j++ {
		secao := secaoAlias(ini,
			fmt.Sprintf("lacres%03d%03d", indiceVolume, j),
			fmt.Sprintf("Lacre%03d%03d", indiceVolume, j))

		nLacre := ini.LerString(secao, "nLacre", "FIM")
		if nLacre == "FIM" || nLacre == "" {
			break
		}
		v.Lacres = append(v.Lacres, Lacre{NLacre: nLacre})
	}
}

// ---------------------------------------------------------------------------
// Cobranca e pagamento
// ---------------------------------------------------------------------------

// lerINICobranca le [Fatura]/[fat] e [DuplicataNNN]/[dupNNN].
// Porte de Ler_Fatura e Ler_Duplicata.
func lerINICobranca(ini *pcn.INI, c *Cobr) {
	secao := secaoAlias(ini, "Fatura", "fat")
	c.Fat.NFat = lerStringAlias(ini, secao, "nFat", "Numero")
	c.Fat.VOrig = lerFloatAlias(ini, secao, "vOrig", "ValorOriginal")
	c.Fat.VDesc = lerFloatAlias(ini, secao, "vDesc", "ValorDesconto")
	c.Fat.VLiq = lerFloatAlias(ini, secao, "vLiq", "ValorLiquido")

	c.Dup = nil
	for i := 1; ; i++ {
		secao := secaoAlias(ini, fmt.Sprintf("Duplicata%03d", i), fmt.Sprintf("dup%03d", i))

		numero := lerStringAlias(ini, secao, "nDup", "Numero")
		if numero == "" || numero == "FIM" {
			break
		}
		c.Dup = append(c.Dup, Dup{
			NDup:  numero,
			DVenc: lerDataAlias(ini, secao, "dVenc", "DataVencimento"),
			VDup:  lerFloatAlias(ini, secao, "vDup", "Valor"),
		})
	}
}

// lerINIPagamento le [pagNNN]. Porte de Ler_Pagamento.
//
// Duas particularidades:
//
//   - o vTroco e da COLECAO, nao do item, e so e atribuido quando POSITIVO
//     (:1521-1522). Um `vTroco=0` numa secao nao zera o que outra tenha
//     posto;
//   - o indPag e lido SEM default (:1511): nao informado, fica no membro
//     vazio e a tag nao sai. Por isso o item parte de NovoPagItem, cujo
//     zero value o ACBr nao reproduz.
func lerINIPagamento(ini *pcn.INI, p *Pag) {
	p.DetPag = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("pag%03d", i)

		forma := ini.LerString(secao, "tpag", "FIM")
		if forma == "FIM" || forma == "" {
			break
		}

		item := NovoPagItem()
		item.TPag, _ = ParseFormaPagamento(forma)
		item.XPag = ini.LerString(secao, "xPag", "")
		item.VPag = ini.LerFloat(secao, "vPag", 0)
		item.DPag = ini.LerData(secao, "dPag", tempoZero())
		item.CNPJPag = ini.LerString(secao, "CNPJPag", "")
		item.UFPag = ini.LerString(secao, "UFPag", "")
		item.IndPag, _ = ParseIndicadorPagamento(ini.LerString(secao, "indPag", ""))
		item.TpIntegra, _ = pcn.ParseTpIntegra(ini.LerString(secao, "tpIntegra", ""))
		item.CNPJ = ini.LerString(secao, "CNPJ", "")
		item.TBand, _ = ParseBandeiraCartao(ini.LerString(secao, "tBand", ""))
		item.CAut = ini.LerString(secao, "cAut", "")
		item.CNPJReceb = ini.LerString(secao, "CNPJReceb", "")
		item.IDTermPag = ini.LerString(secao, "idTermPag", "")

		p.DetPag = append(p.DetPag, item)

		if troco := ini.LerFloat(secao, "vTroco", 0); troco > 0 {
			p.VTroco = troco
		}
	}
}

// lerINIInfIntermed le [infIntermed]. Porte de Ler_InfIntermed.
func lerINIInfIntermed(ini *pcn.INI, i *InfIntermed) {
	const secao = "infIntermed"
	i.CNPJ = ini.LerString(secao, "CNPJ", "")
	i.IDCadIntTran = ini.LerString(secao, "idCadIntTran", "")
}

// ---------------------------------------------------------------------------
// Informacoes adicionais
// ---------------------------------------------------------------------------

// lerINIDadosAdicionais le [DadosAdicionais]/[infAdic] e os tres grupos
// repetidos que pendem dele.
// Porte de Ler_DadosAdicionais, Ler_ObsContribuinte, Ler_ObsFisco e
// Ler_ProcessoReferenciado.
func lerINIDadosAdicionais(ini *pcn.INI, a *InfAdic) {
	secao := secaoAlias(ini, "DadosAdicionais", "infAdic")
	a.InfAdFisco = lerStringAlias(ini, secao, "infAdFisco", "Fisco")
	a.InfCpl = lerStringAlias(ini, secao, "infCpl", "Complemento")

	lerINIObsContribuinte(ini, a)
	lerINIObsFisco(ini, a)
	lerINIProcessoReferenciado(ini, a)
}

// lerINIObsContribuinte le [obsContNNN] ou [InfAdicNNN].
//
// O alias da secao aqui e [InfAdic] COM indice -- o mesmo nome da secao de
// dados adicionais, mas numerada. Sao coisas diferentes: [infAdic] traz
// infAdFisco e infCpl, [InfAdic001] traz a primeira observacao.
func lerINIObsContribuinte(ini *pcn.INI, a *InfAdic) {
	a.ObsCont = nil
	for i := 1; ; i++ {
		secao := secaoAlias(ini, fmt.Sprintf("obsCont%03d", i), fmt.Sprintf("InfAdic%03d", i))

		campo := lerStringAlias(ini, secao, "xCampo", "Campo")
		if campo == "" || campo == "FIM" {
			break
		}
		a.ObsCont = append(a.ObsCont, ObsCont{
			XCampo: campo,
			XTexto: lerStringAlias(ini, secao, "xTexto", "Texto"),
		})
	}
}

// lerINIObsFisco le [obsFiscoNNN]. Sem alias de secao, diferente do grupo
// do contribuinte.
func lerINIObsFisco(ini *pcn.INI, a *InfAdic) {
	a.ObsFisco = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("obsFisco%03d", i)

		campo := lerStringAlias(ini, secao, "xCampo", "Campo")
		if campo == "" || campo == "FIM" {
			break
		}
		a.ObsFisco = append(a.ObsFisco, ObsFisco{
			XCampo: campo,
			XTexto: lerStringAlias(ini, secao, "xTexto", "Texto"),
		})
	}
}

// lerINIProcessoReferenciado le [procRefNNN].
func lerINIProcessoReferenciado(ini *pcn.INI, a *InfAdic) {
	a.ProcRef = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("procRef%03d", i)

		proc := ini.LerString(secao, "nProc", "FIM")
		if proc == "FIM" || proc == "" {
			break
		}
		item := ProcRef{NProc: proc}
		item.IndProc, _ = ParseIndicadorProcesso(ini.LerString(secao, "indProc", "0"))
		item.TpAto, _ = ParseTpAto(ini.LerString(secao, "tpAto", ""))
		a.ProcRef = append(a.ProcRef, item)
	}
}

// ---------------------------------------------------------------------------
// Exportacao, compra, cana e os grupos finais
// ---------------------------------------------------------------------------

// lerINIExportacao le [exporta]. Porte de Ler_Exportacao.
//
// O grupo so e lido quando `UFembarq` OU `UFSaidaPais` tem conteudo -- os
// dois sao a mesma informacao em versoes diferentes do leiaute (ate a 3.10
// e a partir dela).
func lerINIExportacao(ini *pcn.INI, e *Exporta) {
	const secao = "exporta"

	uf := lerStringAlias(ini, secao, "UFembarq", "UFSaidaPais")
	if uf == "" || uf == "FIM" {
		return
	}
	e.UFEmbarq = ini.LerString(secao, "UFembarq", "")
	e.XLocEmbarq = ini.LerString(secao, "xLocEmbarq", "")
	e.UFSaidaPais = ini.LerString(secao, "UFSaidaPais", "")
	e.XLocExporta = ini.LerString(secao, "xLocExporta", "")
	e.XLocDespacho = ini.LerString(secao, "xLocDespacho", "")
}

// lerINICompra le [compra]. Porte de Ler_Compra.
//
// Condicionado a QUALQUER um dos tres campos: a nota de compra publica pode
// trazer so o pedido, so o contrato ou so o empenho.
func lerINICompra(ini *pcn.INI, c *Compra) {
	const secao = "compra"
	nEmp := ini.LerString(secao, "xNEmp", "")
	ped := ini.LerString(secao, "xPed", "")
	cont := ini.LerString(secao, "xCont", "")

	if nEmp == "" && ped == "" && cont == "" {
		return
	}
	c.XNEmp, c.XPed, c.XCont = nEmp, ped, cont
}

// lerINICana le [cana], [forDiaNNN] e [deducNNN]. Porte de Ler_Cana,
// Ler_CanaForDia e Ler_CanaDeducao.
func lerINICana(ini *pcn.INI, c *Cana) {
	const secao = "cana"

	c.Safra = ini.LerString(secao, "safra", "")
	c.Ref = ini.LerString(secao, "ref", "")
	c.QTotMes = ini.LerFloat(secao, "qTotMes", 0)
	c.QTotAnt = ini.LerFloat(secao, "qTotAnt", 0)
	c.QTotGer = ini.LerFloat(secao, "qTotGer", 0)
	c.VFor = ini.LerFloat(secao, "vFor", 0)
	c.VTotDed = ini.LerFloat(secao, "vTotDed", 0)
	c.VLiqFor = ini.LerFloat(secao, "vLiqFor", 0)

	c.ForDia = nil
	for i := 1; ; i++ {
		s := fmt.Sprintf("forDia%03d", i)
		dia := ini.LerString(s, "dia", "FIM")
		if dia == "FIM" || dia == "" {
			break
		}
		c.ForDia = append(c.ForDia, ForDia{
			Dia:  ini.LerInteiro(s, "dia", 0),
			Qtde: ini.LerFloat(s, "qtde", 0),
		})
	}

	c.Deduc = nil
	for i := 1; ; i++ {
		s := fmt.Sprintf("deduc%03d", i)
		ded := ini.LerString(s, "xDed", "FIM")
		if ded == "FIM" || ded == "" {
			break
		}
		c.Deduc = append(c.Deduc, Deduc{
			XDed: ded,
			VDed: ini.LerFloat(s, "vDed", 0),
		})
	}
}

// lerINIInfNFeSupl le [infNFeSupl]. Porte de Ler_InfNFeSupl.
//
// O qrCode e a urlChave sao CALCULADOS na emissao; aqui eles entram porque
// o .ini tambem serve para reimprimir uma nota ja autorizada.
func lerINIInfNFeSupl(ini *pcn.INI, s *InfNFeSupl) {
	const secao = "infNFeSupl"
	if !ini.SecaoExiste(secao) {
		return
	}
	s.QRCode = ini.LerString(secao, "qrCode", "")
	s.URLChave = ini.LerString(secao, "urlChave", "")
}

// lerINIInfRespTec le [infRespTec]. Porte de Ler_InfRespTec.
//
// OMISSAO DO ACBr REPLICADA: idCSRT e hashCSRT existem no struct e NAO sao
// lidos do .ini (:1726-1739). Os dois sao calculados na assinatura a partir
// do CSRT configurado, nao informados pelo emissor -- le-los do arquivo
// deixaria um valor obsoleto vencer o calculado.
func lerINIInfRespTec(ini *pcn.INI, r *InfRespTec) {
	const secao = "infRespTec"
	if !ini.SecaoExiste(secao) {
		return
	}
	r.CNPJ = ini.LerString(secao, "CNPJ", "")
	r.XContato = ini.LerString(secao, "xContato", "")
	r.Email = ini.LerString(secao, "email", "")
	r.Fone = ini.LerString(secao, "fone", "")
}

// lerINIAgropecuario le [defensivoNN] e [guiaTransito].
// Porte de Ler_Defensivo e Ler_GuiaTransito.
//
// O defensivo e o unico grupo repetido do .ini da NF-e que indexa com DOIS
// digitos e tem teto declarado: o comentario do original diz "I varia de 01
// a 20". O teto nao e verificado la nem aqui -- quem o aplica e o schema.
func lerINIAgropecuario(ini *pcn.INI, a *Agropecuario) {
	a.Defensivo = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("defensivo%02d", i)
		receituario := ini.LerString(secao, "nReceituario", "FIM")
		if receituario == "FIM" || receituario == "" {
			break
		}
		a.Defensivo = append(a.Defensivo, Defensivo{
			NReceituario: receituario,
			CPFRespTec:   ini.LerString(secao, "CPFRespTec", ""),
		})
	}

	const secao = "guiaTransito"
	if !ini.SecaoExiste(secao) {
		return
	}
	a.GuiaTransito.UFGuia = ini.LerString(secao, "UFGuia", "")
	a.GuiaTransito.TpGuia, _ = ParseTpGuia(ini.LerString(secao, "tpGuia", ""))
	a.GuiaTransito.SerieGuia = ini.LerString(secao, "serieGuia", "")
	// O nGuia cai para "0", e nao para vazio (:1778).
	a.GuiaTransito.NGuia = ini.LerString(secao, "nGuia", "0")
}
