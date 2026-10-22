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
	"strconv"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Escrita do formato .ini do ACBr. Porte de TNFeIniWriter
// (ACBrNFe.IniWriter.pas).
//
// O gravador usa SEMPRE o nome LONGO das secoes -- [Identificacao],
// [Emitente], [Destinatario], [Produto], [Transportador], [Fatura],
// [Duplicata], [Volume], [Veiculo], [Medicamento], [Combustivel] --, nunca
// o nome curto do leiaute. O leitor aceita os dois, entao a ida e volta
// fecha; um .ini gerado aqui e lido pelo ACBr sem ajuste.
//
// Tres grupos de condicional governam o que sai, e nenhum deles e "o campo
// esta preenchido":
//
//   - grupo inteiro condicionado a UM campo seu: Avulsa pelo CNPJ, Retirada
//     e Entrega pelo CNPJCPF, Veiculo pelo chassi, Combustivel pelo
//     cProdANP, IPI pelo cEnq, II/PISST/COFINSST/ISSQN pela base de
//     calculo positiva, guiaTransito pelo tpGuia, procNFe pelo nProt;
//   - campos escolhidos pelo CST: PIS e COFINS mudam QUAIS valores saem
//     conforme o CST (01/02 por aliquota, 03 por quantidade, 99 os dois);
//   - campos escolhidos pela VERSAO DO LEIAUTE: o medicamento grava
//     cProdANVISA/xMotivoIsencao na 4.00 e nLote/qLote/dFab/dVal antes
//     dela. E por isso que o gravador precisa da versao da nota.
//
// DIVERGENCIA de formato, pelo mesmo motivo do porte da NFGas: o original
// usa DateTimeToStr e WriteFloat, que dependem da configuracao regional da
// maquina e geram arquivo diferente em cada instalacao. Aqui a data sai em
// ISO e o decimal com ponto. O leitor aceita as duas formas.

const (
	layoutDataHoraINI = "2006-01-02T15:04:05"
	layoutDataINI     = "2006-01-02"
)

// GravarINI serializa a NFe no formato .ini do ACBr.
func GravarINI(n *NFe) (string, error) {
	ini, err := MontarINI(n)
	if err != nil {
		return "", err
	}
	return ini.String(), nil
}

// GravarINIArquivo serializa a NFe e grava no caminho informado.
func GravarINIArquivo(n *NFe, caminho string) error {
	ini, err := MontarINI(n)
	if err != nil {
		return err
	}
	return ini.SalvarArquivo(caminho)
}

// MontarINI devolve a NFe ja montada na estrutura de INI, para quem precisar
// acrescentar secoes proprias antes de serializar.
//
// Como o original (GravarIni, IniWriter.pas:162), recusa nota com chave
// invalida -- o `raise EACBrNFeException` vira erro.
func MontarINI(n *NFe) (*pcn.INI, error) {
	if n == nil {
		return nil, ErrXMLVazio
	}
	if err := pcn.ValidarChaveAcesso(n.InfNFe.ID); err != nil {
		return nil, fmt.Errorf("nfe: NFe inconsistente para gerar INI: %w", err)
	}
	ini := pcn.NovoINI()

	ini.GravarString("infNFe", "ID", n.InfNFe.ID)
	ini.GravarFloat("infNFe", "Versao", n.InfNFe.Versao, 2)

	gravarINIIdentificacao(ini, n.Ide)
	gravarININFReferenciada(ini, n.Ide.NFref)
	gravarINIEmitente(ini, n.Emit)
	gravarINIAvulsa(ini, n.Avulsa)
	gravarINIDestinatario(ini, n.Dest)
	gravarINIRetirada(ini, n.Retirada)
	gravarINIEntrega(ini, n.Entrega)
	gravarINIDetalhamento(ini, n)
	gravarINITotal(ini, n.Total)
	gravarINIAutorizadosXML(ini, n.AutXML)
	gravarINITransportador(ini, n.Transp)
	gravarINICobranca(ini, n.Cobr)
	gravarINIPagamento(ini, n.Pag)
	gravarINIInfIntermediario(ini, n.InfIntermed)
	gravarINIDadosAdicionais(ini, n.InfAdic)
	gravarINIExportacao(ini, n.Exporta)
	gravarINICompra(ini, n.Compra)
	gravarINICana(ini, n.Cana)
	gravarINIInfNFeSupl(ini, n.InfNFeSupl)
	gravarINIInfRespTec(ini, n.InfRespTec)
	gravarINIAgropecuario(ini, n.Agropecuario)

	// Ultimo a ser gerado, como no original.
	gravarINIProcNFe(ini, n.ProcNFe)

	return ini, nil
}

// ---------------------------------------------------------------------------
// Identificacao, emitente e destinatario
// ---------------------------------------------------------------------------

func gravarINIIdentificacao(ini *pcn.INI, ide Ide) {
	const secao = "Identificacao"

	ini.GravarInteiro(secao, "cUF", ide.CUF)
	ini.GravarInteiro(secao, "cNF", ide.CNF)
	ini.GravarString(secao, "natOp", ide.NatOp)
	ini.GravarString(secao, "indPag", ide.IndPag.String())
	ini.GravarInteiro(secao, "Modelo", ide.Modelo)
	ini.GravarInteiro(secao, "Serie", ide.Serie)
	ini.GravarInteiro(secao, "nNF", ide.NNF)
	ini.GravarData(secao, "dhEmi", ide.DEmi, layoutDataHoraINI)
	ini.GravarData(secao, "dhSaiEnt", ide.DSaiEnt, layoutDataHoraINI)
	ini.GravarData(secao, "dPrevEntrega", ide.DPrevEntrega, layoutDataINI)
	ini.GravarString(secao, "tpNF", ide.TpNF.String())
	ini.GravarString(secao, "idDest", ide.IdDest.String())
	ini.GravarInteiro(secao, "cMunFG", ide.CMunFG)
	ini.GravarString(secao, "tpAmb", ide.TpAmb.String())
	ini.GravarString(secao, "tpImp", ide.TpImp.String())
	ini.GravarString(secao, "tpemis", ide.TpEmis.String())
	ini.GravarString(secao, "finNFe", ide.FinNFe.String())
	ini.GravarString(secao, "indFinal", ide.IndFinal.String())
	ini.GravarString(secao, "indPres", ide.IndPres.String())
	ini.GravarString(secao, "indIntermed", ide.IndIntermed.String())
	ini.GravarString(secao, "procEmi", ide.ProcEmi.String())
	ini.GravarString(secao, "verProc", ide.VerProc)
	ini.GravarData(secao, "dhCont", ide.DhCont, layoutDataHoraINI)
	ini.GravarString(secao, "xJust", ide.XJust)

	// Reforma Tributaria
	ini.GravarInteiro(secao, "cMunFGIBS", ide.CMunFGIBS)
	ini.GravarString(secao, "tpNFDebito", ide.TpNFDebito.String())
	ini.GravarString(secao, "tpNFCredito", ide.TpNFCredito.String())
	ini.GravarString(secao, "cIndOp", ide.CIndOp)

	rtc.GravarINIGCompraGov(ini, ide.GCompraGov)
	rtc.GravarINIGPagAntecipado(ini, ide.GPagAntecipado)
}

// gravarININFReferenciada grava [NFRefNNN]. Porte de Gerar_NFReferenciada.
//
// O tipo e escolhido por uma CASCATA de `else if` sobre o que estiver
// preenchido, na ordem refNFe, refNFeSig, RefNF.CNPJ, RefNFP.CNPJCPF,
// refCTe, RefECF.nCOO. Preenchendo mais de um, so o primeiro sai -- e os
// grupos sao mutuamente exclusivos no leiaute, entao isso esta certo.
//
// DIVERGENCIA: o original grava o modelo do ECF na chave `modelo`
// (:308), mas o leitor procura `mod` ou `ModECF` (:293). O modelo se perde
// numa ida e volta pelo ACBr. Aqui a chave gravada e `mod`, que o leitor
// dos dois lados encontra -- `modelo` nao e lido por ninguem.
func gravarININFReferenciada(ini *pcn.INI, refs []NFref) {
	for i, r := range refs {
		secao := fmt.Sprintf("NFRef%03d", i+1)

		switch {
		case trimINI(r.RefNFe) != "":
			ini.GravarString(secao, "Tipo", "NFe")
			ini.GravarString(secao, "refNFe", r.RefNFe)

		case trimINI(r.RefNFeSig) != "":
			ini.GravarString(secao, "Tipo", "NFe")
			ini.GravarString(secao, "refNFeSig", r.RefNFeSig)

		case trimINI(r.RefNF.CNPJ) != "":
			ini.GravarString(secao, "Tipo", "NF")
			ini.GravarInteiro(secao, "cUF", r.RefNF.CUF)
			ini.GravarString(secao, "AAMM", r.RefNF.AAMM)
			ini.GravarString(secao, "CNPJ", r.RefNF.CNPJ)
			ini.GravarInteiro(secao, "Modelo", r.RefNF.Modelo)
			ini.GravarInteiro(secao, "Serie", r.RefNF.Serie)
			ini.GravarInteiro(secao, "nNF", r.RefNF.NNF)

		case trimINI(r.RefNFP.CNPJCPF) != "":
			ini.GravarString(secao, "Tipo", "NFP")
			ini.GravarInteiro(secao, "cUF", r.RefNFP.CUF)
			ini.GravarString(secao, "AAMM", r.RefNFP.AAMM)
			ini.GravarString(secao, "CNPJ", r.RefNFP.CNPJCPF)
			ini.GravarString(secao, "IE", r.RefNFP.IE)
			ini.GravarString(secao, "Modelo", r.RefNFP.Modelo)
			ini.GravarInteiro(secao, "Serie", r.RefNFP.Serie)
			ini.GravarInteiro(secao, "nNF", r.RefNFP.NNF)

		case trimINI(r.RefCTe) != "":
			ini.GravarString(secao, "Tipo", "CTe")
			ini.GravarString(secao, "refCTe", r.RefCTe)

		case trimINI(r.RefECF.NCOO) != "":
			ini.GravarString(secao, "Tipo", "ECF")
			ini.GravarString(secao, "mod", r.RefECF.Modelo.String())
			ini.GravarString(secao, "nECF", r.RefECF.NECF)
			ini.GravarString(secao, "nCOO", r.RefECF.NCOO)
		}
	}
}

func gravarINIEmitente(ini *pcn.INI, e Emit) {
	const secao = "Emitente"

	ini.GravarString(secao, "CNPJCPF", e.CNPJCPF)
	ini.GravarString(secao, "xNome", e.XNome)
	ini.GravarString(secao, "xFant", e.XFant)
	ini.GravarString(secao, "IE", e.IE)
	ini.GravarString(secao, "IEST", e.IEST)
	ini.GravarString(secao, "IM", e.IM)
	ini.GravarString(secao, "CNAE", e.CNAE)
	ini.GravarString(secao, "CRT", e.CRT.String())
	ini.GravarString(secao, "ISUFEmit", e.ISUFEmit)
	ini.GravarString(secao, "xLgr", e.EnderEmit.XLgr)
	ini.GravarString(secao, "nro", e.EnderEmit.Nro)
	ini.GravarString(secao, "xCpl", e.EnderEmit.XCpl)
	ini.GravarString(secao, "xBairro", e.EnderEmit.XBairro)
	ini.GravarInteiro(secao, "cMun", e.EnderEmit.CMun)
	ini.GravarString(secao, "xMun", e.EnderEmit.XMun)
	ini.GravarString(secao, "UF", e.EnderEmit.UF)
	ini.GravarInteiro(secao, "CEP", e.EnderEmit.CEP)
	ini.GravarInteiro(secao, "cPais", e.EnderEmit.CPais)
	ini.GravarString(secao, "xPais", e.EnderEmit.XPais)
	ini.GravarString(secao, "Fone", e.EnderEmit.Fone)
}

func gravarINIAvulsa(ini *pcn.INI, a Avulsa) {
	const secao = "Avulsa"
	if a.CNPJ == "" {
		return
	}
	ini.GravarString(secao, "CNPJ", a.CNPJ)
	ini.GravarString(secao, "xOrgao", a.XOrgao)
	ini.GravarString(secao, "matr", a.Matr)
	ini.GravarString(secao, "xAgente", a.XAgente)
	ini.GravarString(secao, "fone", a.Fone)
	ini.GravarString(secao, "UF", a.UF)
	ini.GravarString(secao, "nDAR", a.NDAR)
	ini.GravarData(secao, "dEmi", a.DEmi, layoutDataINI)
	ini.GravarFloat(secao, "vDAR", a.VDAR, 2)
	ini.GravarString(secao, "repEmi", a.RepEmi)
	ini.GravarData(secao, "dPag", a.DPag, layoutDataINI)
}

func gravarINIDestinatario(ini *pcn.INI, d Dest) {
	const secao = "Destinatario"

	// O idEstrangeiro so sai quando preenchido -- e o unico campo do grupo
	// com condicional propria (:363).
	if d.IDEstrangeiro != "" {
		ini.GravarString(secao, "idEstrangeiro", d.IDEstrangeiro)
	}

	ini.GravarString(secao, "CNPJCPF", d.CNPJCPF)
	ini.GravarString(secao, "xNome", d.XNome)
	ini.GravarString(secao, "indIEDest", d.IndIEDest.String())
	ini.GravarString(secao, "IE", d.IE)
	ini.GravarString(secao, "ISUF", d.ISUF)
	ini.GravarString(secao, "IM", d.IM)
	ini.GravarString(secao, "Email", d.Email)
	ini.GravarString(secao, "xLgr", d.EnderDest.XLgr)
	ini.GravarString(secao, "nro", d.EnderDest.Nro)
	ini.GravarString(secao, "xCpl", d.EnderDest.XCpl)
	ini.GravarString(secao, "xBairro", d.EnderDest.XBairro)
	ini.GravarInteiro(secao, "cMun", d.EnderDest.CMun)
	ini.GravarString(secao, "xMun", d.EnderDest.XMun)
	ini.GravarString(secao, "UF", d.EnderDest.UF)
	ini.GravarInteiro(secao, "CEP", d.EnderDest.CEP)
	ini.GravarInteiro(secao, "cPais", d.EnderDest.CPais)
	ini.GravarString(secao, "xPais", d.EnderDest.XPais)
	ini.GravarString(secao, "Fone", d.EnderDest.Fone)
}

// gravarINIRetirada grava [Retirada]. Porte de Gerar_Retirada.
//
// ACRESCIMO: o original grava OITO campos (:387-399) e o leitor le QUINZE.
// Uma ida e volta pelo ACBr perde xNome, CEP, cPais, xPais, Fone, Email e
// IE do local de retirada. Aqui saem os quinze -- sao campos do leiaute, o
// leitor do proprio ACBr os recupera, e nada no original indica que a
// omissao seja intencional.
//
// A CONDICIONAL, essa sim, esta replicada como no original: o grupo depende
// do CNPJCPF, embora o leitor condicione ao xLgr. Um local de retirada sem
// documento e lido e nao e gravado de volta -- e deterministico, e mudar a
// condicao aqui faria o .ini divergir do que o ACBr produz para a mesma
// nota.
func gravarINIRetirada(ini *pcn.INI, r Retirada) {
	const secao = "Retirada"
	if r.CNPJCPF == "" {
		return
	}
	ini.GravarString(secao, "CNPJCPF", r.CNPJCPF)
	ini.GravarString(secao, "xNome", r.XNome)
	ini.GravarString(secao, "xLgr", r.XLgr)
	ini.GravarString(secao, "nro", r.Nro)
	ini.GravarString(secao, "xCpl", r.XCpl)
	ini.GravarString(secao, "xBairro", r.XBairro)
	ini.GravarInteiro(secao, "cMun", r.CMun)
	ini.GravarString(secao, "xMun", r.XMun)
	ini.GravarString(secao, "UF", r.UF)
	ini.GravarInteiro(secao, "CEP", r.CEP)
	ini.GravarInteiro(secao, "cPais", r.CPais)
	ini.GravarString(secao, "xPais", r.XPais)
	ini.GravarString(secao, "Fone", r.Fone)
	ini.GravarString(secao, "Email", r.Email)
	ini.GravarString(secao, "IE", r.IE)
}

func gravarINIEntrega(ini *pcn.INI, e Entrega) {
	const secao = "Entrega"
	if e.CNPJCPF == "" {
		return
	}
	ini.GravarString(secao, "CNPJCPF", e.CNPJCPF)
	ini.GravarString(secao, "xNome", e.XNome)
	ini.GravarString(secao, "xLgr", e.XLgr)
	ini.GravarString(secao, "nro", e.Nro)
	ini.GravarString(secao, "xCpl", e.XCpl)
	ini.GravarString(secao, "xBairro", e.XBairro)
	ini.GravarInteiro(secao, "cMun", e.CMun)
	ini.GravarString(secao, "xMun", e.XMun)
	ini.GravarString(secao, "UF", e.UF)
	ini.GravarInteiro(secao, "CEP", e.CEP)
	ini.GravarInteiro(secao, "cPais", e.CPais)
	ini.GravarString(secao, "xPais", e.XPais)
	ini.GravarString(secao, "fone", e.Fone)
	ini.GravarString(secao, "email", e.Email)
	ini.GravarString(secao, "IE", e.IE)
}

// ---------------------------------------------------------------------------
// Totais
// ---------------------------------------------------------------------------

// gravarINITotal grava [Total], [ISSQNtot] e [retTrib].
// Porte de Gerar_Total, Gerar_ISSQNtot e Gerar_RetTributo.
func gravarINITotal(ini *pcn.INI, t Total) {
	const secao = "Total"
	v := t.ICMSTot

	ini.GravarFloat(secao, "vBC", v.VBC, 2)
	ini.GravarFloat(secao, "vICMS", v.VICMS, 2)
	ini.GravarFloat(secao, "vICMSDeson", v.VICMSDeson, 2)
	ini.GravarFloat(secao, "vFCP", v.VFCP, 2)
	ini.GravarFloat(secao, "vICMSUFDest", v.VICMSUFDest, 2)
	ini.GravarFloat(secao, "vICMSUFRemet", v.VICMSUFRemet, 2)
	ini.GravarFloat(secao, "vFCPUFDest", v.VFCPUFDest, 2)
	ini.GravarFloat(secao, "vBCST", v.VBCST, 2)
	ini.GravarFloat(secao, "vST", v.VST, 2)
	ini.GravarFloat(secao, "vFCPST", v.VFCPST, 2)
	ini.GravarFloat(secao, "vFCPSTRet", v.VFCPSTRet, 2)
	ini.GravarFloat(secao, "qBCMono", v.QBCMono, 4)
	ini.GravarFloat(secao, "vICMSMono", v.VICMSMono, 2)
	ini.GravarFloat(secao, "qBCMonoReten", v.QBCMonoReten, 4)
	ini.GravarFloat(secao, "vICMSMonoReten", v.VICMSMonoReten, 2)
	ini.GravarFloat(secao, "qBCMonoRet", v.QBCMonoRet, 4)
	ini.GravarFloat(secao, "vICMSMonoRet", v.VICMSMonoRet, 2)
	ini.GravarFloat(secao, "vProd", v.VProd, 2)
	ini.GravarFloat(secao, "vFrete", v.VFrete, 2)
	ini.GravarFloat(secao, "vSeg", v.VSeg, 2)
	ini.GravarFloat(secao, "vDesc", v.VDesc, 2)
	ini.GravarFloat(secao, "vII", v.VII, 2)
	ini.GravarFloat(secao, "vIPI", v.VIPI, 2)
	ini.GravarFloat(secao, "vIPIDevol", v.VIPIDevol, 2)
	ini.GravarFloat(secao, "vPIS", v.VPIS, 2)
	ini.GravarFloat(secao, "vCOFINS", v.VCOFINS, 2)
	ini.GravarFloat(secao, "vOutro", v.VOutro, 2)
	ini.GravarFloat(secao, "vNF", v.VNF, 2)
	ini.GravarFloat(secao, "vTotTrib", v.VTotTrib, 2)
	ini.GravarFloat(secao, "vNFTot", t.VNFTot, 2)

	rtc.GravarINIISTot(ini, t.ISTot)
	rtc.GravarINIIBSCBSTot(ini, t.IBSCBSTot)

	gravarINITotalISSQN(ini, t.ISSQNTot)
	gravarINITotalRetTributo(ini, t.RetTrib)
}

func gravarINITotalISSQN(ini *pcn.INI, t ISSQNTot) {
	const secao = "ISSQNtot"
	ini.GravarFloat(secao, "vServ", t.VServ, 2)
	ini.GravarFloat(secao, "vBC", t.VBC, 2)
	ini.GravarFloat(secao, "vISS", t.VISS, 2)
	ini.GravarFloat(secao, "vPIS", t.VPIS, 2)
	ini.GravarFloat(secao, "vCOFINS", t.VCOFINS, 2)
	ini.GravarData(secao, "dCompet", t.DCompet, layoutDataINI)
	ini.GravarFloat(secao, "vDeducao", t.VDeducao, 2)
	ini.GravarFloat(secao, "vOutro", t.VOutro, 2)
	ini.GravarFloat(secao, "vDescIncond", t.VDescIncond, 2)
	ini.GravarFloat(secao, "vDescCond", t.VDescCond, 2)
	ini.GravarFloat(secao, "vISSRet", t.VISSRet, 2)
	ini.GravarString(secao, "cRegTrib", t.CRegTrib.String())
}

func gravarINITotalRetTributo(ini *pcn.INI, r RetTrib) {
	const secao = "retTrib"
	ini.GravarFloat(secao, "vRetPIS", r.VRetPIS, 2)
	ini.GravarFloat(secao, "vRetCOFINS", r.VRetCOFINS, 2)
	ini.GravarFloat(secao, "vRetCSLL", r.VRetCSLL, 2)
	ini.GravarFloat(secao, "vBCIRRF", r.VBCIRRF, 2)
	ini.GravarFloat(secao, "vIRRF", r.VIRRF, 2)
	ini.GravarFloat(secao, "vBCRetPrev", r.VBCRetPrev, 2)
	ini.GravarFloat(secao, "vRetPrev", r.VRetPrev, 2)
}

// ---------------------------------------------------------------------------
// Autorizados, transporte e cobranca
// ---------------------------------------------------------------------------

// gravarINIAutorizadosXML grava [autXMLNN] -- com DOIS digitos, ao
// contrario da maioria dos grupos. Porte de Gerar_AutorizadosXml.
func gravarINIAutorizadosXML(ini *pcn.INI, autXML []AutXML) {
	for i, a := range autXML {
		ini.GravarString(fmt.Sprintf("autXML%02d", i+1), "CNPJCPF", a.CNPJCPF)
	}
}

func gravarINITransportador(ini *pcn.INI, t Transp) {
	const secao = "Transportador"

	ini.GravarString(secao, "modFrete", t.ModFrete.String())
	ini.GravarString(secao, "CNPJCPF", t.Transporta.CNPJCPF)
	ini.GravarString(secao, "xNome", t.Transporta.XNome)
	ini.GravarString(secao, "IE", t.Transporta.IE)
	ini.GravarString(secao, "xEnder", t.Transporta.XEnder)
	ini.GravarString(secao, "xMun", t.Transporta.XMun)
	ini.GravarString(secao, "UF", t.Transporta.UF)
	ini.GravarFloat(secao, "vServ", t.RetTransp.VServ, 2)
	ini.GravarFloat(secao, "vBCRet", t.RetTransp.VBCRet, 2)
	ini.GravarFloat(secao, "pICMSRet", t.RetTransp.PICMSRet, 2)
	ini.GravarFloat(secao, "vICMSRet", t.RetTransp.VICMSRet, 2)
	ini.GravarString(secao, "CFOP", t.RetTransp.CFOP)
	ini.GravarInteiro(secao, "cMunFG", t.RetTransp.CMunFG)
	ini.GravarString(secao, "Placa", t.VeicTransp.Placa)
	ini.GravarString(secao, "UFPlaca", t.VeicTransp.UF)
	ini.GravarString(secao, "RNTC", t.VeicTransp.RNTC)
	ini.GravarString(secao, "vagao", t.Vagao)
	ini.GravarString(secao, "balsa", t.Balsa)

	for i, r := range t.Reboque {
		s := fmt.Sprintf("Reboque%03d", i+1)
		ini.GravarString(s, "placa", r.Placa)
		ini.GravarString(s, "UF", r.UF)
		ini.GravarString(s, "RNTC", r.RNTC)
	}

	for i, v := range t.Vol {
		s := fmt.Sprintf("Volume%03d", i+1)
		ini.GravarInteiro(s, "qVol", v.QVol)
		ini.GravarString(s, "esp", v.Esp)
		ini.GravarString(s, "marca", v.Marca)
		ini.GravarString(s, "nVol", v.NVol)
		ini.GravarFloat(s, "pesoL", v.PesoL, 3)
		ini.GravarFloat(s, "pesoB", v.PesoB, 3)

		for j, l := range v.Lacres {
			ini.GravarString(fmt.Sprintf("Lacre%03d%03d", i+1, j+1), "nLacre", l.NLacre)
		}
	}
}

func gravarINICobranca(ini *pcn.INI, c Cobr) {
	const secao = "Fatura"
	ini.GravarString(secao, "nFat", c.Fat.NFat)
	ini.GravarFloat(secao, "vOrig", c.Fat.VOrig, 2)
	ini.GravarFloat(secao, "vDesc", c.Fat.VDesc, 2)
	ini.GravarFloat(secao, "vLiq", c.Fat.VLiq, 2)

	for i, d := range c.Dup {
		s := fmt.Sprintf("Duplicata%03d", i+1)
		ini.GravarString(s, "nDup", d.NDup)
		ini.GravarData(s, "dVenc", d.DVenc, layoutDataINI)
		ini.GravarFloat(s, "vDup", d.VDup, 2)
	}
}

// gravarINIPagamento grava [pagNNN]. Porte de Gerar_Pagamento.
//
// O vTroco e da COLECAO, nao do item, e o original o grava FORA do laco
// usando a variavel de secao que sobrou da ULTIMA volta (:1356). Com a
// colecao vazia essa variavel nunca foi atribuida, e o Delphi grava numa
// secao de nome indefinido -- comportamento que o CLAUDE.md manda nao
// replicar.
//
// Aqui o troco vai na ULTIMA secao de pagamento, que e o que o original faz
// no caso definido, e NAO sai quando nao ha pagamento nenhum. O leitor
// aceita o troco em qualquer das secoes e guarda o ultimo positivo.
func gravarINIPagamento(ini *pcn.INI, p Pag) {
	var ultima string
	for i, d := range p.DetPag {
		secao := fmt.Sprintf("pag%03d", i+1)
		ultima = secao

		ini.GravarString(secao, "tPag", d.TPag.String())
		ini.GravarString(secao, "xPag", d.XPag)
		ini.GravarFloat(secao, "vPag", d.VPag, 2)
		ini.GravarData(secao, "dPag", d.DPag, layoutDataINI)
		ini.GravarString(secao, "CNPJPag", d.CNPJPag)
		ini.GravarString(secao, "UFPag", d.UFPag)
		ini.GravarString(secao, "indPag", d.IndPag.String())
		ini.GravarString(secao, "tpIntegra", d.TpIntegra.String())
		ini.GravarString(secao, "CNPJ", d.CNPJ)
		ini.GravarString(secao, "tBand", d.TBand.String())
		ini.GravarString(secao, "cAut", d.CAut)
		ini.GravarString(secao, "CNPJReceb", d.CNPJReceb)
		ini.GravarString(secao, "idTermPag", d.IDTermPag)
	}
	if ultima != "" {
		ini.GravarFloat(ultima, "vTroco", p.VTroco, 2)
	}
}

func gravarINIInfIntermediario(ini *pcn.INI, i InfIntermed) {
	const secao = "infIntermed"
	ini.GravarString(secao, "CNPJ", i.CNPJ)
	ini.GravarString(secao, "idCadIntTran", i.IDCadIntTran)
}

// ---------------------------------------------------------------------------
// Informacoes adicionais e grupos finais
// ---------------------------------------------------------------------------

// gravarINIDadosAdicionais grava [DadosAdicionais] e os tres grupos
// repetidos. Porte de Gerar_DadosAdicionais e das tres procedures
// seguintes.
//
// As observacoes do contribuinte saem em [InfAdicNNN] e as do fisco em
// [ObsFiscoNNN] -- nomes assimetricos, e e assim no original.
func gravarINIDadosAdicionais(ini *pcn.INI, a InfAdic) {
	const secao = "DadosAdicionais"
	ini.GravarString(secao, "infAdFisco", a.InfAdFisco)
	ini.GravarString(secao, "infCpl", a.InfCpl)

	for i, o := range a.ObsCont {
		s := fmt.Sprintf("InfAdic%03d", i+1)
		ini.GravarString(s, "xCampo", o.XCampo)
		ini.GravarString(s, "xTexto", o.XTexto)
	}
	for i, o := range a.ObsFisco {
		s := fmt.Sprintf("ObsFisco%03d", i+1)
		ini.GravarString(s, "xCampo", o.XCampo)
		ini.GravarString(s, "xTexto", o.XTexto)
	}
	for i, p := range a.ProcRef {
		s := fmt.Sprintf("procRef%03d", i+1)
		ini.GravarString(s, "nProc", p.NProc)
		ini.GravarString(s, "indProc", p.IndProc.String())
		ini.GravarString(s, "tpAto", p.TpAto.String())
	}
}

func gravarINIExportacao(ini *pcn.INI, e Exporta) {
	const secao = "Exporta"
	if e.UFEmbarq == "" && e.UFSaidaPais == "" {
		return
	}
	ini.GravarString(secao, "UFembarq", e.UFEmbarq)
	ini.GravarString(secao, "xLocEmbarq", e.XLocEmbarq)
	ini.GravarString(secao, "UFSaidaPais", e.UFSaidaPais)
	ini.GravarString(secao, "xLocExporta", e.XLocExporta)
	ini.GravarString(secao, "xLocDespacho", e.XLocDespacho)
}

// gravarINICompra grava [Compra]. Porte de Gerar_Compra.
//
// CONDICIONAL REPLICADA: o grupo so sai com `xNEmp` preenchido, embora o
// leitor aceite qualquer um dos tres campos (:1641-1643). Uma nota com
// pedido e sem empenho e lida e nao e gravada de volta. Deterministico, e
// mudar a condicao faria o .ini divergir do que o ACBr produz.
func gravarINICompra(ini *pcn.INI, c Compra) {
	const secao = "Compra"
	if c.XNEmp == "" {
		return
	}
	ini.GravarString(secao, "xNEmp", c.XNEmp)
	ini.GravarString(secao, "xPed", c.XPed)
	ini.GravarString(secao, "xCont", c.XCont)
}

func gravarINICana(ini *pcn.INI, c Cana) {
	const secao = "cana"
	ini.GravarString(secao, "safra", c.Safra)
	ini.GravarString(secao, "ref", c.Ref)
	ini.GravarFloat(secao, "qTotMes", c.QTotMes, 10)
	ini.GravarFloat(secao, "qTotAnt", c.QTotAnt, 10)
	ini.GravarFloat(secao, "qTotGer", c.QTotGer, 10)
	ini.GravarFloat(secao, "vFor", c.VFor, 2)
	ini.GravarFloat(secao, "vTotDed", c.VTotDed, 2)
	ini.GravarFloat(secao, "vLiqFor", c.VLiqFor, 2)

	for i, f := range c.ForDia {
		s := fmt.Sprintf("forDia%03d", i+1)
		ini.GravarInteiro(s, "dia", f.Dia)
		ini.GravarFloat(s, "qtde", f.Qtde, 10)
	}
	for i, d := range c.Deduc {
		s := fmt.Sprintf("deduc%03d", i+1)
		ini.GravarString(s, "xDed", d.XDed)
		ini.GravarFloat(s, "vDed", d.VDed, 2)
	}
}

func gravarINIInfNFeSupl(ini *pcn.INI, s InfNFeSupl) {
	const secao = "infNFeSupl"
	ini.GravarString(secao, "qrCode", s.QRCode)
	ini.GravarString(secao, "urlChave", s.URLChave)
}

func gravarINIInfRespTec(ini *pcn.INI, r InfRespTec) {
	const secao = "infRespTec"
	ini.GravarString(secao, "CNPJ", r.CNPJ)
	ini.GravarString(secao, "xContato", r.XContato)
	ini.GravarString(secao, "email", r.Email)
	ini.GravarString(secao, "fone", r.Fone)
}

// gravarINIAgropecuario grava [defensivoNN] e [guiaTransito].
// Porte de Gerar_Defensivo e Gerar_GuiaTransito.
func gravarINIAgropecuario(ini *pcn.INI, a Agropecuario) {
	for i, d := range a.Defensivo {
		s := fmt.Sprintf("defensivo%02d", i+1)
		ini.GravarString(s, "nReceituario", d.NReceituario)
		ini.GravarString(s, "CPFRespTec", d.CPFRespTec)
	}

	// A guia so sai com o tipo preenchido -- e por isso que TpGuia precisa
	// ter o membro "nenhum" no ordinal zero.
	if a.GuiaTransito.TpGuia == TpgNenhum {
		return
	}
	const secao = "guiaTransito"
	ini.GravarString(secao, "UFGuia", a.GuiaTransito.UFGuia)
	ini.GravarString(secao, "tpGuia", a.GuiaTransito.TpGuia.String())
	ini.GravarString(secao, "serieGuia", a.GuiaTransito.SerieGuia)
	ini.GravarString(secao, "nGuia", a.GuiaTransito.NGuia)
}

// gravarINIProcNFe grava [procNFe]. Porte de Gerar_ProcNFe.
//
// So sai com protocolo -- uma nota ainda nao autorizada nao tem o grupo.
func gravarINIProcNFe(ini *pcn.INI, p pcn.ProcDFe) {
	const secao = "procNFe"
	if p.NProt == "" {
		return
	}
	ini.GravarString(secao, "tpAmb", p.TpAmb.String())
	ini.GravarString(secao, "verAplic", p.VerAplic)
	ini.GravarString(secao, "chNFe", p.ChDFe)
	ini.GravarData(secao, "dhRecbto", p.DhRecbto, layoutDataHoraINI)
	ini.GravarString(secao, "nProt", p.NProt)
	ini.GravarString(secao, "digVal", p.DigVal)
	ini.GravarString(secao, "cStat", strconv.Itoa(p.CStat))
	ini.GravarString(secao, "xMotivo", p.XMotivo)
}
