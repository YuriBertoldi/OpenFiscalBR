// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-29.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfgas

import (
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Geracao do XML da NFGas -- porte 1:1 de TNFGasXmlWriter
// (ACBrNFGas.XmlWriter.pas). Cada gerarXxx corresponde a um Gerar_Xxx do
// original, com a mesma ordem de campos e as mesmas condicionais.
//
// Divergencias deliberadas em relacao ao ACBr (todas documentadas tambem em
// .claude/skills/convert/referencias/divergencias-acbr.md):
//
//  1. A raiz do documento processado e `nfgasProc`, como no XSD
//     (nfgasProc_v1.00.xsd). O ACBr gera `NFGasProc`, que nao valida contra
//     o schema e nao e o que o proprio leitor do ACBr procura.
//  2. Enums sao gravados pelo CODIGO do leiaute (String()), nunca pelo
//     ordinal. O ACBr passa o enum cru ao AddNode em finNFGas, indOrigemQtd,
//     tpMotNaoLeitura, tpProc, modBCST e motDesICMS, gravando o ordinal --
//     XML invalido de forma deterministica.
//  3. ICMS70: o ACBr gera o bloco vICMSDeson DUAS vezes (a segunda com
//     cBenef); o XSD admite uma. Geramos apenas o primeiro bloco
//     (vICMSDeson + motDesICMS + indDeduzDeson); cBenef nao e emitido no
//     ICMS70.
//  4. CST de ICMS fora do leiaute: o Delphi faz `Result := nil` e segue
//     anexando (access violation garantida). Aqui o grupo imposto sai nil e
//     os grupos PIS/COFINS/IBSCBS daquele item nao sao gerados.
//  5. Competencia (AAAAMM) com data zero gera tag vazia, nao "189912".
//  6. cNF aleatorio e sorteado com 7 digitos (a posicao 36 da chave e o
//     nSiteAutoriz); o GerarCodigoDFe do ACBr sorteia com 8 e pode produzir
//     chave invalida para a NFGas.
//
// OMISSOES DELIBERADAS: a ListaDeAlertas (wAlerta) e a opcao
// NormatizarMunicipios (consulta ao arquivo de municipios) nao foram
// portadas; a validacao efetiva e o XSD da SEFAZ e regras_negocio.go.

// GerarXML gera o XML da NFGas a partir dos dados do documento.
// Porte de TNFGasXmlWriter.GerarXml.
//
// Como no original, tem efeitos colaterais no documento: monta a chave de
// acesso (sorteando cNF quando zero), e grava InfNFGas.ID, Ide.CDV e
// Ide.CNF. Se ProcNFGas.NProt estiver preenchido, o retorno e o documento
// processado (raiz nfgasProc com protNFGas); senao, a NFGas solta.
// A Signature so e embutida se ja estiver preenchida (o equivalente do
// taSomenteSeAssinada, default do ACBr) -- para assinar, use Assinar.
func GerarXML(n *NFGas) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}

	// cNF: 0 ou -1 sorteia; -2 ou menor vira 0 (GerarChaveAcesso do
	// ACBrDFeUtil).
	if n.Ide.CNF == 0 || n.Ide.CNF == -1 {
		n.Ide.CNF = gerarCodigoDFe(n.Ide.NNF)
	} else if n.Ide.CNF < 0 {
		n.Ide.CNF = 0
	}
	if n.Ide.Modelo == 0 {
		n.Ide.Modelo = ModeloNFGas
	}
	if n.InfNFGas.Versao == 0 {
		n.InfNFGas.Versao = 1.00
	}

	chave, err := MontarChaveAcesso(n)
	if err != nil {
		return "", err
	}
	n.InfNFGas.ID = "NFGas" + chave
	dv, _ := strconv.Atoi(chave[43:])
	n.Ide.CDV = dv
	cnf, _ := strconv.Atoi(chave[36:43])
	n.Ide.CNF = cnf

	w := &xmlWriter{n: n, chave: chave, rtc: rtc.NovoWriter(rtc.ModeloNFGas)}

	nota := pcn.NovoElem("NFGas").Filho(w.gerarInfNFGas())

	if n.InfNFGasSupl.QrCodNFGas != "" {
		nota.Filho(pcn.NovoElem("infNFGasSupl").
			Filho(pcn.NovoElem("qrCodNFGas").
				TextoBruto("<![CDATA[" + n.InfNFGasSupl.QrCodNFGas + "]]>")))
	}

	// so embute assinatura COMPLETA, como o taSomenteSeAssinada exige
	// (Digest + SignatureValue + X509Certificate, nunca um bloco parcial)
	if n.Signature.Assinada() {
		nota.Filho(gerarSignature(&n.Signature))
	}

	if n.ProcNFGas.NProt != "" {
		// DIVERGENCIA 1: raiz nfgasProc conforme o XSD (ACBr: NFGasProc).
		// O namespace e declarado uma vez, na raiz do proc.
		return pcn.NovoElem("nfgasProc").
			Attr("versao", pcn.FormatarVersaoXML(n.InfNFGas.Versao)).
			Attr("xmlns", Namespace).
			Filho(nota).
			Filho(w.gerarProtNFGas()).XML(), nil
	}
	nota.Attr("xmlns", Namespace)
	return nota.XML(), nil
}

// GerarXMLProc gera o XML de nfgasProc; exige protocolo preenchido.
func GerarXMLProc(n *NFGas) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}
	if n.ProcNFGas.NProt == "" {
		return "", ErrProtocoloAusente
	}
	return GerarXML(n)
}

// Assinar assina o XML da NFGas (elemento infNFGas) com o certificado A1,
// inserindo o bloco Signature como ultimo filho de NFGas. Substitui o stub
// da fase de leitura; a assinatura em si vive em packages/dfe.
func Assinar(cert *dfe.Certificado, xmlStr string) (string, error) {
	return dfe.AssinarXML(cert, xmlStr, "infNFGas")
}

// AssinarEvento assina o XML de eventoNFGas (elemento infEvento).
func AssinarEvento(cert *dfe.Certificado, xmlStr string) (string, error) {
	return dfe.AssinarXML(cert, xmlStr, "infEvento")
}

// xmlWriter carrega o estado da geracao de um documento, como o
// TNFGasXmlWriter (FChaveNFGas + o estado herdado do writer da RTC).
type xmlWriter struct {
	n     *NFGas
	chave string
	rtc   *rtc.Writer
}

func (w *xmlWriter) gerarInfNFGas() *pcn.Elem {
	n := w.n
	e := pcn.NovoElem("infNFGas").
		Attr("Id", "NFGas"+w.chave).
		Attr("versao", pcn.FormatarVersaoXML(n.InfNFGas.Versao)).
		Filho(w.gerarIde()).
		Filho(w.gerarEmit()).
		Filho(w.gerarDest())

	if strings.TrimSpace(n.Instalacao.IDInstalacao) != "" {
		e.Filho(w.gerarInstalacao())
	}
	if strings.TrimSpace(n.GSub.ChNFGas) != "" || strings.TrimSpace(n.GSub.GNF.CNPJ) != "" {
		e.Filho(w.gerarGSub())
	}
	for _, v := range n.GVolContrat {
		e.Filho(w.gerarGVolContrat(v))
	}
	for _, m := range n.GMed {
		e.Filho(w.gerarGMed(m))
	}
	for _, d := range n.Det {
		e.Filho(w.gerarDet(d))
	}
	e.Filho(w.gerarTotal(n.Total)).
		Filho(w.rtc.GerarPgtoVinc(n.PgtoVinc)).
		Filho(w.gerarGFat(n.GFat)).
		Filho(w.gerarGAgencia(n.GAgencia))

	for _, a := range n.AutXML {
		e.Filho(pcn.NovoElem("autXML").Filho(nodeCNPJCPF(a.CNPJCPF, true)))
	}

	e.Filho(w.gerarInfAdic(n.InfAdic)).
		Filho(w.gerarRespTec(n.InfRespTec))
	return e
}

func (w *xmlWriter) gerarIde() *pcn.Elem {
	ide := w.n.Ide
	uf := pcn.SiglaUF(ide.CUF)
	e := pcn.NovoElem("ide").
		Filho(pcn.NodeInt("cUF", ide.CUF, 2, true)).
		Filho(pcn.NodeStr("tpAmb", ide.TpAmb.String(), true)).
		Filho(pcn.NodeInt("mod", ide.Modelo, 2, true)).
		Filho(pcn.NodeInt("serie", ide.Serie, 1, true)).
		Filho(pcn.NodeInt("nNF", ide.NNF, 1, true)).
		Filho(pcn.NodeInt("cNF", ide.CNF, 7, true)).
		Filho(pcn.NodeInt("cDV", ide.CDV, 1, true)).
		Filho(pcn.NodeStr("dhEmi", pcn.FormatarDataHoraXML(ide.DhEmi, uf), true)).
		Filho(pcn.NodeStr("tpEmis", ide.TpEmis.String(), true)).
		Filho(pcn.NodeStr("nSiteAutoriz", ide.NSiteAutoriz.String(), true)).
		Filho(pcn.NodeInt("cMunFG", ide.CMunFG, 7, true)).
		// DIVERGENCIA 2: codigo do leiaute; o ACBr grava o ordinal.
		Filho(pcn.NodeStr("finNFGas", ide.FinNFGas.String(), true)).
		Filho(pcn.NodeStr("tpFat", ide.TpFat.String(), true)).
		Filho(pcn.NodeStr("verProc", ide.VerProc, true))

	if !ide.DhCont.IsZero() || strings.TrimSpace(ide.XJust) != "" {
		e.Filho(pcn.NodeStr("dhCont", pcn.FormatarDataHoraXML(ide.DhCont, uf), true)).
			Filho(pcn.NodeStr("xJust", ide.XJust, true))
	}

	e.Filho(w.rtc.GerarGCompraGovReduzido(ide.GCompraGov))

	if ide.TpPagAnt != pcn.TpaNenhum {
		e.Filho(pcn.NodeStr("tpPagAnt", ide.TpPagAnt.String(), true))
	}
	return e
}

func (w *xmlWriter) gerarEmit() *pcn.Elem {
	emit := w.n.Emit
	return pcn.NovoElem("emit").
		Filho(pcn.NodeStr("CNPJ", emit.CNPJ, true)).
		Filho(pcn.NodeStr("IE", pcn.OnlyNumber(emit.IE), true)).
		Filho(pcn.NodeStr("xNome", emit.XNome, true)).
		Filho(pcn.NodeStr("xFant", emit.XFant, false)).
		Filho(w.gerarEndereco("enderEmit", emit.EnderEmit, false)).
		Filho(pcn.NodeStr("ISUFEmit", emit.ISUFEmit, false))
}

func (w *xmlWriter) gerarDest() *pcn.Elem {
	dest := w.n.Dest
	e := pcn.NovoElem("dest").
		Filho(pcn.NodeStr("xNome", dest.XNome, true))

	if strings.TrimSpace(dest.IDOutros()) != "" {
		e.Filho(pcn.NodeStr("idOutros", dest.IDOutros(), true))
	} else {
		e.Filho(nodeCNPJCPF(dest.CNPJCPF, true))
	}

	e.Filho(pcn.NodeStr("IE", dest.IE, false)).
		Filho(pcn.NodeStr("IM", dest.IM, false))

	if strings.TrimSpace(dest.CNIS) != "" {
		e.Filho(pcn.NodeStr("cNIS", dest.CNIS, false))
	} else if strings.TrimSpace(dest.NB) != "" {
		e.Filho(pcn.NodeStr("NB", dest.NB, false))
	}

	return e.Filho(pcn.NodeStr("xNomeAdicional", dest.XNomeAdicional, false)).
		Filho(w.gerarEndereco("enderDest", dest.EnderDest, false))
}

// gerarEndereco cobre enderEmit, enderDest e enderCorresp -- os campos e a
// ordem sao identicos nos tres; so muda a obrigatoriedade do CEP
// (opcional em emit/dest, obrigatorio em enderCorresp).
func (w *xmlWriter) gerarEndereco(tag string, end Endereco, cepObrigatorio bool) *pcn.Elem {
	return pcn.NovoElem(tag).
		Filho(pcn.NodeStr("xLgr", end.XLgr, true)).
		Filho(pcn.NodeStr("nro", end.Nro, true)).
		Filho(pcn.NodeStr("xCpl", end.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", end.XBairro, true)).
		Filho(pcn.NodeInt("cMun", end.CMun, 7, true)).
		Filho(pcn.NodeStr("xMun", end.XMun, true)).
		Filho(pcn.NodeInt("CEP", end.CEP, 8, cepObrigatorio)).
		Filho(pcn.NodeStr("UF", end.UF, true)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(end.Fone), false)).
		Filho(pcn.NodeStr("email", end.Email, false))
}

func (w *xmlWriter) gerarInstalacao() *pcn.Elem {
	inst := w.n.Instalacao
	e := pcn.NovoElem("instalacao").
		Filho(pcn.NodeStr("idInstalacao", inst.IDInstalacao, true)).
		Filho(pcn.NodeStr("idCodCliente", inst.IDCodCliente, false)).
		Filho(pcn.NodeStr("tpInstalacao", inst.TpInstalacao.String(), true)).
		Filho(pcn.NodeStr("nContrato", inst.NContrato, false)).
		Filho(pcn.NodeStr("tpClasse", inst.TpClasse.String(), true)).
		Filho(pcn.NodeStr("xClasse", inst.XClasse, false))

	if strings.TrimSpace(inst.LatGPS) != "" && strings.TrimSpace(inst.LongGPS) != "" {
		e.Filho(pcn.NodeStr("latGPS", inst.LatGPS, true)).
			Filho(pcn.NodeStr("longGPS", inst.LongGPS, true))
	}
	return e.Filho(pcn.NodeStr("codRoteiroLeitura", inst.CodRoteiroLeitura, false))
}

func (w *xmlWriter) gerarGSub() *pcn.Elem {
	gSub := w.n.GSub
	e := pcn.NovoElem("gSub")
	if strings.TrimSpace(gSub.ChNFGas) != "" {
		e.Filho(pcn.NodeStrSemFiltro("chNFGas", gSub.ChNFGas, true))
	} else {
		e.Filho(w.gerarGNF(gSub.GNF))
	}
	return e.Filho(pcn.NodeStr("motSub", gSub.MotSub.String(), true))
}

func (w *xmlWriter) gerarGNF(gNF GNF) *pcn.Elem {
	return pcn.NovoElem("gNF").
		Filho(pcn.NodeStr("CNPJ", gNF.CNPJ, true)).
		Filho(pcn.NodeStr("serie", gNF.Serie, true)).
		Filho(pcn.NodeInt("nNF", gNF.NNF, 1, true)).
		Filho(pcn.NodeStr("CompetEmis", formatarCompetencia(gNF.CompetEmis), true)).
		Filho(pcn.NodeStr("CompetApur", formatarCompetencia(gNF.CompetApur), true)).
		Filho(pcn.NodeStr("hash115", gNF.Hash115, false))
}

func (w *xmlWriter) gerarGVolContrat(v GVolContrat) *pcn.Elem {
	return pcn.NovoElem("gVolContrat").
		Attr("nContrat", pcn.FormatarInteiroZeros(v.NContrat, 2)).
		Filho(pcn.NodeStr("tpVolContrat", v.TpVolContrat.String(), true)).
		Filho(pcn.NodeDec("qUnidContrat", v.QUnidContrat, 6, true))
}

func (w *xmlWriter) gerarGMed(m GMed) *pcn.Elem {
	return pcn.NovoElem("gMed").
		Attr("nMed", pcn.FormatarInteiroZeros(m.NMed, 2)).
		Filho(pcn.NodeStr("idEqp", m.IDEqp, true)).
		Filho(pcn.NodeDat("dMedAnt", m.DMedAnt, true)).
		Filho(pcn.NodeDec("vMedAnt", m.VMedAnt, 4, true)).
		Filho(pcn.NodeDat("dMedAtu", m.DMedAtu, true)).
		Filho(pcn.NodeDec("vMedAtu", m.VMedAtu, 4, true)).
		Filho(pcn.NodeStr("tpEqp", m.TpEqp.String(), true)).
		Filho(pcn.NodeStr("tpMedidor", m.TpMedidor.String(), false))
}

func (w *xmlWriter) gerarDet(d Det) *pcn.Elem {
	e := pcn.NovoElem("det").Attr("nItem", strconv.Itoa(d.NItem))
	if strings.TrimSpace(d.ChNFGasAnt) != "" {
		e.Attr("chNFGasAnt", d.ChNFGasAnt)
	}
	if d.NItemAnt > 0 {
		e.Attr("nItemAnt", strconv.Itoa(d.NItemAnt))
	}
	if strings.TrimSpace(d.GAgregadora.CClass) != "" || d.GAgregadora.VTotDFe > 0 {
		return e.Filho(w.gerarGAgregadora(d.GAgregadora))
	}
	return e.Filho(w.gerarGNormal(d.GNormal))
}

func (w *xmlWriter) gerarGNormal(g GNormal) *pcn.Elem {
	e := pcn.NovoElem("gNormal")
	for _, t := range g.GTarif {
		e.Filho(w.gerarGTarif(t))
	}
	e.Filho(w.gerarProd(g.Prod)).
		Filho(w.gerarImposto(g.Imposto)).
		Filho(w.gerarGProcRef(g.GProcRef))
	if strings.TrimSpace(g.InfAdProd) != "" {
		e.Filho(pcn.NodeStr("infAdProd", g.InfAdProd, false))
	}
	return e
}

func (w *xmlWriter) gerarGTarif(t GTarif) *pcn.Elem {
	return pcn.NovoElem("gTarif").
		Filho(pcn.NodeDat("dIniTarif", t.DIniTarif, true)).
		Filho(pcn.NodeDat("dFimTarif", t.DFimTarif, false)).
		Filho(pcn.NodeStr("nAto", t.NAto, true)).
		Filho(pcn.NodeInt("anoAto", t.AnoAto, 4, true)).
		Filho(pcn.NodeStr("tpFaixaCons", t.TpFaixaCons.String(), true)).
		Filho(pcn.NodeDec("vTarifAplic", t.VTarifAplic, 8, true))
}

func (w *xmlWriter) gerarProd(p Prod) *pcn.Elem {
	e := pcn.NovoElem("prod").
		// DIVERGENCIA 2: codigo do leiaute; o ACBr grava o ordinal.
		Filho(pcn.NodeStr("indOrigemQtd", p.IndOrigemQtd.String(), true))

	// OMISSAO DO ACBr REPLICADA: a guarda exige vMed > 0
	// (ACBrNFGas.XmlWriter.pas:687), mas dentro de gerarGMedicao o
	// tpMotNaoLeitura/xMotNaoLeitura so sai no ramo vMed == 0 -- ou seja,
	// esse ramo e INALCANCAVEL, aqui e no Delphi. Nao "corrigir" sem decidir
	// a divergencia: trocar a guarda faria o Go emitir tag que o ACBr nunca
	// emite. Ver packages/nfgas/README.md e a lista tagsForaDoWriter de
	// demos/nfgas/exemplos_test.go. A NFAg nao tem o problema: la a guarda
	// exige apenas nMed > 0.
	if p.GMedicao.NMed > 0 && p.GMedicao.GMedida.VMed > 0 {
		e.Filho(w.gerarGMedicao(p.GMedicao))
	}

	e.Filho(pcn.NodeStr("cProd", p.CProd, true)).
		Filho(pcn.NodeStr("xProd", p.XProd, true)).
		Filho(pcn.NodeStr("cClass", p.CClass, false)).
		Filho(pcn.NodeInt("CFOP", p.CFOP, 4, false)).
		Filho(pcn.NodeStr("uMed", p.UMed.String(), true)).
		Filho(pcn.NodeDec("qFaturada", p.QFaturada, 4, true)).
		Filho(pcn.NodeDec("vItem", p.VItem, 8, true)).
		Filho(pcn.NodeDec("fatorPCS", p.FatorPCS, 4, false)).
		Filho(pcn.NodeDec("fatorPTZ", p.FatorPTZ, 4, false)).
		Filho(pcn.NodeDec("fatorP", p.FatorP, 4, false)).
		Filho(pcn.NodeDec("fatorT", p.FatorT, 4, false)).
		Filho(pcn.NodeDec("vProd", p.VProd, 8, true))

	if p.IndDevolucao == pcn.TieSim {
		e.Filho(pcn.NodeStr("indDevolucao", "1", true))
	}
	if p.GPagAntecipado.ChDFePagAnt != "" {
		e.Filho(w.rtc.GerarGPagAntecipadoProd(p.GPagAntecipado))
	}
	return e
}

func (w *xmlWriter) gerarGMedicao(g GMedicao) *pcn.Elem {
	e := pcn.NovoElem("gMedicao").
		Filho(pcn.NodeStr("nMed", pcn.FormatarInteiroZeros(g.NMed, 2), true)).
		// OMISSAO DO ACBr REPLICADA: nContrat e formatado com FormatFloat
		// '00' ANTES do AddNode opcional -- "00" nunca e vazio, entao a tag
		// sai sempre, mesmo sem contrato.
		Filho(pcn.NodeStr("nContrat", pcn.FormatarInteiroZeros(g.NContrat, 2), false))

	if g.GMedida.VMed > 0 {
		e.Filho(pcn.NovoElem("gMedida").
			Filho(pcn.NodeStr("uMed", g.GMedida.UMed.String(), true)).
			Filho(pcn.NodeDec("vMed", g.GMedida.VMed, 4, true)))
	} else {
		// DIVERGENCIA 2: codigo do leiaute; o ACBr grava o ordinal.
		e.Filho(pcn.NodeStr("tpMotNaoLeitura", g.TpMotNaoLeitura.String(), true)).
			Filho(pcn.NodeStr("xMotNaoLeitura", g.XMotNaoLeitura, false))
	}
	return e
}

func (w *xmlWriter) gerarImposto(imp Imposto) *pcn.Elem {
	e := pcn.NovoElem("imposto")

	if imp.IndSemCST == pcn.TieSim {
		e.Filho(pcn.NodeStr("indSemCST", "1", false))
	} else {
		e.Filho(pcn.NodeStr("orig", imp.Orig.String(), true))

		switch imp.ICMS.CST {
		case pcn.CST00:
			e.Filho(w.gerarICMS00(imp.ICMS))
		case pcn.CST10:
			e.Filho(w.gerarICMS10(imp.ICMS))
		case pcn.CST20:
			e.Filho(w.gerarICMS20(imp.ICMS))
		case pcn.CST40, pcn.CST41:
			e.Filho(w.gerarICMS40(imp.ICMS))
		case pcn.CST51:
			e.Filho(w.gerarICMS51(imp.ICMS))
		case pcn.CST60:
			e.Filho(w.gerarICMS60(imp.ICMS))
		case pcn.CST70:
			e.Filho(w.gerarICMS70(imp.ICMS))
		case pcn.CST90:
			e.Filho(w.gerarICMS90(imp.ICMS))
		default:
			// DIVERGENCIA 4: no Delphi, CST fora do leiaute faz Result:=nil
			// e os AppendChild seguintes estourariam com access violation.
			// Aqui o imposto inteiro sai nil (o item fica sem imposto).
			e = nil
		}
	}

	return e.Filho(w.rtc.GerarIBSCBS(imp.IBSCBS)).
		Filho(w.gerarPIS(imp.PIS)).
		Filho(w.gerarCOFINS(imp.COFINS)).
		Filho(w.gerarRetTrib(imp.RetTrib)).
		Filho(w.gerarTxReg(imp.TxReg))
}

func (w *xmlWriter) gerarICMS00(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS00").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
		Filho(pcn.NodeDec("pICMS", icms.PICMS, 2, true)).
		Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))
	if icms.PFCP > 0 || icms.VFCP > 0 {
		e.Filho(pcn.NodeDec("pFCP", icms.PFCP, 4, false)).
			Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS10(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS10").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true)).
		Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
		Filho(pcn.NodeDec("pICMSST", icms.PICMSST, 2, true)).
		Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))
	if icms.PFCPST > 0 || icms.VFCPST > 0 {
		e.Filho(pcn.NodeDec("pFCPST", icms.PFCPST, 4, false)).
			Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS20(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS20").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true)).
		Filho(pcn.NodeDec("pRedBC", icms.PRedBC, 2, false)).
		Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
		Filho(pcn.NodeDec("pICMS", icms.PICMS, 2, true)).
		Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))
	if icms.VICMSDeson > 0 {
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("cBenef", icms.CBenef, false))
	}
	if icms.PFCP > 0 || icms.VFCP > 0 {
		e.Filho(pcn.NodeDec("pFCP", icms.PFCP, 4, false)).
			Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS40(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS40").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true))
	if icms.VICMSDeson > 0 {
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("cBenef", icms.CBenef, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS51(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS51").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true))
	if icms.VICMSDeson > 0 {
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("cBenef", icms.CBenef, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS60(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS60").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true))

	if icms.VBCSTRet > 0 || icms.PICMSSTRet > 0 || icms.VICMSSubstituto > 0 ||
		icms.VICMSSTRet > 0 || icms.VBCFCPSTRet > 0 || icms.PFCPSTRet > 0 ||
		icms.VFCPSTRet > 0 {
		e.Filho(pcn.NodeDec("vBCSTRet", icms.VBCSTRet, 2, false)).
			Filho(pcn.NodeDec("pICMSSTRet", icms.PICMSSTRet, 2, false)).
			Filho(pcn.NodeDec("vICMSSubstituto", icms.VICMSSubstituto, 2, false)).
			Filho(pcn.NodeDec("vICMSSTRet", icms.VICMSSTRet, 2, false))
		if icms.VBCFCPSTRet > 0 || icms.PFCPSTRet > 0 || icms.VFCPSTRet > 0 {
			e.Filho(pcn.NodeDec("vBCFCPSTRet", icms.VBCFCPSTRet, 2, false)).
				Filho(pcn.NodeDec("pFCPSTRet", icms.PFCPSTRet, 2, false)).
				Filho(pcn.NodeDec("vFCPSTRet", icms.VFCPSTRet, 2, false))
		}
	} else if icms.PRedBCEfet > 0 || icms.VBCEfet > 0 || icms.PICMSEfet > 0 ||
		icms.VICMSEfet > 0 {
		e.Filho(pcn.NodeDec("pRedBCEfet", icms.PRedBCEfet, 2, true)).
			Filho(pcn.NodeDec("vBCEfet", icms.VBCEfet, 2, true)).
			Filho(pcn.NodeDec("pICMSEfet", icms.PICMSEfet, 2, true)).
			Filho(pcn.NodeDec("vICMSEfet", icms.VICMSEfet, 2, true))
	}

	if icms.VICMSDeson > 0 {
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("cBenef", icms.CBenef, false))
	}
	return e
}

func (w *xmlWriter) gerarICMS70(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS70").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true)).
		Filho(pcn.NodeStr("modBC", icms.ModBC.String(), true)).
		Filho(pcn.NodeDec("pRedBC", icms.PRedBC, 2, true)).
		Filho(pcn.NodeDec("vBC", icms.VBC, 2, true)).
		Filho(pcn.NodeDec("pICMS", icms.PICMS, 2, true)).
		Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, true))

	if icms.VBCFCP > 0 || icms.PFCP > 0 || icms.VFCP > 0 {
		e.Filho(pcn.NodeDec("vBCFCP", icms.VBCFCP, 2, false)).
			Filho(pcn.NodeDec("pFCP", icms.PFCP, 4, false)).
			Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, false))
	}

	// DIVERGENCIA 2: codigo do leiaute; o ACBr passa o enum modBCST cru.
	e.Filho(pcn.NodeStr("modBCST", icms.ModBCST.String(), true))

	if icms.PMVAST > 0 {
		e.Filho(pcn.NodeDec("pMVAST", icms.PMVAST, 2, false))
	}
	if icms.PRedBCST > 0 {
		e.Filho(pcn.NodeDec("pRedBCST", icms.PRedBCST, 2, false))
	}

	e.Filho(pcn.NodeDec("vBCST", icms.VBCST, 2, true)).
		Filho(pcn.NodeDec("pICMSST", icms.PICMSST, 2, true)).
		Filho(pcn.NodeDec("vICMSST", icms.VICMSST, 2, true))

	if icms.VBCFCPST > 0 || icms.PFCPST > 0 || icms.VFCPST > 0 {
		e.Filho(pcn.NodeDec("vBCFCPST", icms.VBCFCPST, 2, false)).
			Filho(pcn.NodeDec("pFCPST", icms.PFCPST, 2, false)).
			Filho(pcn.NodeDec("vFCPST", icms.VFCPST, 2, false))
	}

	if icms.VICMSDeson > 0 {
		// DIVERGENCIA 2 e 3: motDesICMS pelo codigo (o ACBr grava o ordinal
		// via tcInt) e SEM o segundo bloco vICMSDeson/cBenef duplicado que o
		// ACBr emite logo em seguida.
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("motDesICMS", icms.MotDesICMS.String(), false))
		if icms.IndDeduzDeson == pcn.TieSim {
			e.Filho(pcn.NodeStr("indDeduzDeson", icms.IndDeduzDeson.String(), false))
		}
	}
	return e
}

func (w *xmlWriter) gerarICMS90(icms ICMS) *pcn.Elem {
	e := pcn.NovoElem("ICMS90").
		Filho(pcn.NodeStr("CST", icms.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", icms.VBC, 2, false)).
		Filho(pcn.NodeDec("pICMS", icms.PICMS, 2, false)).
		Filho(pcn.NodeDec("vICMS", icms.VICMS, 2, false))
	if icms.VICMSDeson > 0 {
		e.Filho(pcn.NodeDec("vICMSDeson", icms.VICMSDeson, 2, false)).
			Filho(pcn.NodeStr("cBenef", icms.CBenef, false))
	}
	if icms.PFCP > 0 || icms.VFCP > 0 {
		e.Filho(pcn.NodeDec("pFCP", icms.PFCP, 4, false)).
			Filho(pcn.NodeDec("vFCP", icms.VFCP, 2, false))
	}
	return e
}

func (w *xmlWriter) gerarPIS(p PIS) *pcn.Elem {
	if p.VBC == 0 && p.PPIS == 0 && p.VPIS == 0 {
		return nil
	}
	return pcn.NovoElem("PIS").
		Filho(pcn.NodeStr("CST", p.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", p.VBC, 2, true)).
		Filho(pcn.NodeDec("pPIS", p.PPIS, 4, true)).
		Filho(pcn.NodeDec("vPIS", p.VPIS, 2, true))
}

func (w *xmlWriter) gerarCOFINS(c COFINS) *pcn.Elem {
	if c.VBC == 0 && c.PCOFINS == 0 && c.VCOFINS == 0 {
		return nil
	}
	return pcn.NovoElem("COFINS").
		Filho(pcn.NodeStr("CST", c.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", c.VBC, 2, true)).
		Filho(pcn.NodeDec("pCOFINS", c.PCOFINS, 4, true)).
		Filho(pcn.NodeDec("vCOFINS", c.VCOFINS, 2, true))
}

func (w *xmlWriter) gerarRetTrib(r RetTrib) *pcn.Elem {
	if r.VRetPIS == 0 && r.VRetCOFINS == 0 && r.VRetCSLL == 0 && r.VIRRF == 0 {
		return nil
	}
	// a tag de COFINS retida e vRetCofins (grafia do leiaute)
	return pcn.NovoElem("retTrib").
		Filho(pcn.NodeDec("vRetPIS", r.VRetPIS, 2, true)).
		Filho(pcn.NodeDec("vRetCofins", r.VRetCOFINS, 2, true)).
		Filho(pcn.NodeDec("vRetCSLL", r.VRetCSLL, 2, true)).
		Filho(pcn.NodeDec("vIRRF", r.VIRRF, 2, true))
}

func (w *xmlWriter) gerarTxReg(t TxReg) *pcn.Elem {
	if t.VBC == 0 && t.PTaxa == 0 && t.VTaxa == 0 {
		return nil
	}
	return pcn.NovoElem("TxReg").
		Filho(pcn.NodeDec("vBC", t.VBC, 2, true)).
		Filho(pcn.NodeDec("pTaxa", t.PTaxa, 4, true)).
		Filho(pcn.NodeDec("vTaxa", t.VTaxa, 2, true))
}

func (w *xmlWriter) gerarGProcRef(g GProcRef) *pcn.Elem {
	if g.VItem == 0 && g.QFaturada == 0 && g.VProd == 0 &&
		g.IndDevolucao != pcn.TieSim && len(g.GProc) == 0 {
		return nil
	}
	e := pcn.NovoElem("gProcRef").
		Filho(pcn.NodeDec("vItem", g.VItem, 8, true))

	// qFaturada fracionada sai como tcDe4; inteira sai como tcInt.
	// A condicao e Frac() > 0, literal do original: um valor NEGATIVO
	// fracionado (Frac < 0) cai no ramo tcInt, como no Delphi.
	if frac := g.QFaturada - math.Trunc(g.QFaturada); frac > 0 {
		e.Filho(pcn.NodeDec("qFaturada", g.QFaturada, 4, true))
	} else {
		e.Filho(pcn.NodeInt("qFaturada", int(g.QFaturada), 1, true))
	}

	e.Filho(pcn.NodeDec("vProd", g.VProd, 8, true))

	if g.IndDevolucao == pcn.TieSim {
		e.Filho(pcn.NodeStr("indDevolucao", "1", true))
	}
	for _, p := range g.GProc {
		e.Filho(pcn.NovoElem("gProc").
			// DIVERGENCIA 2: codigo do leiaute; o ACBr grava o ordinal.
			Filho(pcn.NodeStr("tpProc", p.TpProc.String(), true)).
			Filho(pcn.NodeStr("nProcesso", p.NProcesso, true)))
	}
	return e
}

func (w *xmlWriter) gerarGAgregadora(g GAgregadora) *pcn.Elem {
	return pcn.NovoElem("gAgregadora").
		Filho(pcn.NodeStr("cClass", g.CClass, true)).
		Filho(pcn.NodeDec("vTotDFe", g.VTotDFe, 2, true))
}

func (w *xmlWriter) gerarTotal(t Total) *pcn.Elem {
	return pcn.NovoElem("total").
		Filho(pcn.NodeDec("vProd", t.VProd, 2, true)).
		Filho(pcn.NovoElem("ICMSTot").
			Filho(pcn.NodeDec("vBC", t.VBC, 2, true)).
			Filho(pcn.NodeDec("vICMS", t.VICMS, 2, true)).
			Filho(pcn.NodeDec("vICMSDeson", t.VICMSDeson, 2, true)).
			Filho(pcn.NodeDec("vFCP", t.VFCP, 2, true)).
			Filho(pcn.NodeDec("vBCST", t.VBCST, 2, true)).
			Filho(pcn.NodeDec("vST", t.VST, 2, true)).
			Filho(pcn.NodeDec("vFCPST", t.VFCPST, 2, true))).
		Filho(pcn.NovoElem("vRetTribTot").
			Filho(pcn.NodeDec("vRetPIS", t.VRetPIS, 2, true)).
			Filho(pcn.NodeDec("vRetCofins", t.VRetCOFINS, 2, true)).
			Filho(pcn.NodeDec("vRetCSLL", t.VRetCSLL, 2, true)).
			Filho(pcn.NodeDec("vIRRF", t.VIRRF, 2, true))).
		Filho(pcn.NodeDec("vCOFINS", t.VCOFINS, 2, true)).
		Filho(pcn.NodeDec("vPIS", t.VPIS, 2, true)).
		Filho(pcn.NodeDec("vTxReg", t.VTxReg, 2, true)).
		Filho(pcn.NodeDec("vNF", t.VNF, 2, true)).
		Filho(w.rtc.GerarIBSCBSTot(t.IBSCBSTot)).
		Filho(pcn.NodeDec("vTotDFe", t.VTotDFe, 2, false))
}

func (w *xmlWriter) gerarGFat(g GFat) *pcn.Elem {
	if g.DVencFat.IsZero() {
		return nil
	}
	e := pcn.NovoElem("gFat").
		Filho(pcn.NodeStr("CompetFat", formatarCompetencia(g.CompetFat), true)).
		Filho(pcn.NodeDat("dVencFat", g.DVencFat, true)).
		Filho(pcn.NodeDat("dApresFat", g.DApresFat, false)).
		Filho(pcn.NodeDat("dProxLeitura", g.DProxLeitura, true)).
		Filho(pcn.NodeStr("nFat", g.NFat, false)).
		Filho(pcn.NodeStr("codBarras", g.CodBarras, true))

	if strings.TrimSpace(g.CodDebAuto) != "" {
		e.Filho(pcn.NodeStr("codDebAuto", g.CodDebAuto, false))
	} else if strings.TrimSpace(g.CodBanco) != "" && strings.TrimSpace(g.CodAgencia) != "" {
		e.Filho(pcn.NodeStr("codBanco", g.CodBanco, false)).
			Filho(pcn.NodeStr("codAgencia", g.CodAgencia, false))
	}

	if strings.TrimSpace(g.EnderCorresp.XLgr) != "" {
		e.Filho(w.gerarEndereco("enderCorresp", g.EnderCorresp, true))
	}
	if strings.TrimSpace(g.GPIX.URLQRCodePIX) != "" {
		e.Filho(pcn.NovoElem("gPIX").
			Filho(pcn.NodeStr("urlQRCodePIX", g.GPIX.URLQRCodePIX, true)))
	}
	return e.Filho(pcn.NodeStr("infAdFat", g.InfAdFat, false))
}

func (w *xmlWriter) gerarGAgencia(g GAgencia) *pcn.Elem {
	if strings.TrimSpace(g.NomeAgenciaAtend) == "" &&
		strings.TrimSpace(g.EnderAgenciaAtend) == "" &&
		strings.TrimSpace(g.SitioAgenciaAtend) == "" &&
		strings.TrimSpace(g.InfAdReg) == "" && len(g.GHistCons) == 0 {
		return nil
	}
	e := pcn.NovoElem("gAgencia").
		Filho(pcn.NodeStr("nomeAgenciaAtend", g.NomeAgenciaAtend, false)).
		Filho(pcn.NodeStr("enderAgenciaAtend", g.EnderAgenciaAtend, false)).
		Filho(pcn.NodeStr("sitioAgenciaAtend", g.SitioAgenciaAtend, false))

	for _, h := range g.GHistCons {
		e.Filho(w.gerarGHistCons(h))
	}
	return e.Filho(pcn.NodeStr("infAdReg", g.InfAdReg, false))
}

func (w *xmlWriter) gerarGHistCons(h GHistCons) *pcn.Elem {
	if strings.TrimSpace(h.XHistorico) == "" && h.MedMensal == 0 && len(h.GCons) == 0 {
		return nil
	}
	e := pcn.NovoElem("gHistCons").
		Filho(pcn.NodeStr("xHistorico", h.XHistorico, true))
	for _, c := range h.GCons {
		e.Filho(pcn.NovoElem("gCons").
			Filho(pcn.NodeStr("CompetFat", formatarCompetencia(c.CompetFat), true)).
			Filho(pcn.NodeStr("uMed", c.UMed.String(), true)).
			Filho(pcn.NodeInt("qtdDias", c.QtdDias, 1, true)).
			Filho(pcn.NodeDec("medDiaria", c.MedDiaria, 4, false)).
			Filho(pcn.NodeDec("consumo", c.Consumo, 4, false)).
			Filho(pcn.NodeDec("vFat", c.VFat, 4, true)))
	}
	return e.Filho(pcn.NodeDec("medMensal", h.MedMensal, 4, true))
}

func (w *xmlWriter) gerarInfAdic(i InfAdic) *pcn.Elem {
	// o elemento infAdic e SEMPRE gerado, mesmo vazio, como no original.
	// infCpl: o ACBr grava um so (TODO aberto la); aqui vai a primeira
	// posicao do slice, espelhando o leitor.
	infCpl := ""
	if len(i.InfCpl) > 0 {
		infCpl = i.InfCpl[0]
	}
	return pcn.NovoElem("infAdic").
		Filho(pcn.NodeStr("infAdFisco", i.InfAdFisco, false)).
		Filho(pcn.NodeStr("infCpl", infCpl, false))
}

func (w *xmlWriter) gerarRespTec(rt InfRespTec) *pcn.Elem {
	if rt.CNPJ == "" {
		return nil
	}
	e := pcn.NovoElem("gRespTec").
		Filho(pcn.NodeStrSemFiltro("CNPJ", pcn.OnlyCPFCNPJAlphaNum(strings.TrimSpace(rt.CNPJ)), true)).
		Filho(pcn.NodeStr("xContato", rt.XContato, true)).
		Filho(pcn.NodeStr("email", rt.Email, true)).
		Filho(pcn.NodeStr("fone", rt.Fone, true))

	// no ACBr idCSRT/CSRT vem da configuracao do componente; aqui o hash ja
	// chega calculado (Componente.PrepararRespTec / dfe.HashCSRT).
	if rt.IDCSRT != 0 && rt.HashCSRT != "" {
		e.Filho(pcn.NodeInt("idCSRT", rt.IDCSRT, 3, true)).
			Filho(pcn.NodeStrSemFiltro("hashCSRT", rt.HashCSRT, true))
	}
	return e
}

func (w *xmlWriter) gerarProtNFGas() *pcn.Elem {
	p := w.n.ProcNFGas
	uf := pcn.SiglaUF(w.n.Ide.CUF)
	// Gerar_ProcNFGas usa AddChild+Content direto: todas as tags saem,
	// mesmo vazias, sem FiltrarTextoXML.
	return pcn.NovoElem("protNFGas").
		Attr("versao", pcn.FormatarVersaoXML(w.n.InfNFGas.Versao)).
		Filho(pcn.NovoElem("infProt").
			Filho(pcn.NovoElem("tpAmb").Texto(p.TpAmb.String())).
			Filho(pcn.NovoElem("verAplic").Texto(p.VerAplic)).
			Filho(pcn.NovoElem("chNFGas").Texto(p.ChDFe)).
			Filho(pcn.NovoElem("dhRecbto").Texto(pcn.FormatarDataHoraXML(p.DhRecbto, uf))).
			Filho(pcn.NovoElem("nProt").Texto(p.NProt)).
			Filho(pcn.NovoElem("digVal").Texto(p.DigVal)).
			Filho(pcn.NovoElem("cStat").Texto(strconv.Itoa(p.CStat))).
			Filho(pcn.NovoElem("xMotivo").Texto(p.XMotivo)))
}

// gerarSignature reembute um bloco Signature ja existente (lido de um XML
// assinado). Porte do template TSignature.GerarXML.
func gerarSignature(s *pcn.Signature) *pcn.Elem {
	const nsDSig = "http://www.w3.org/2000/09/xmldsig#"
	const algC14N = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"
	return pcn.NovoElem("Signature").Attr("xmlns", nsDSig).
		Filho(pcn.NovoElem("SignedInfo").
			Filho(pcn.NovoElem("CanonicalizationMethod").Attr("Algorithm", algC14N)).
			Filho(pcn.NovoElem("SignatureMethod").Attr("Algorithm", nsDSig+"rsa-sha1")).
			Filho(pcn.NovoElem("Reference").Attr("URI", s.URI).
				Filho(pcn.NovoElem("Transforms").
					Filho(pcn.NovoElem("Transform").Attr("Algorithm", nsDSig+"enveloped-signature")).
					Filho(pcn.NovoElem("Transform").Attr("Algorithm", algC14N))).
				Filho(pcn.NovoElem("DigestMethod").Attr("Algorithm", nsDSig+"sha1")).
				Filho(pcn.NovoElem("DigestValue").Texto(s.DigestValue)))).
		Filho(pcn.NovoElem("SignatureValue").Texto(s.SignatureValue)).
		Filho(pcn.NovoElem("KeyInfo").
			Filho(pcn.NovoElem("X509Data").
				Filho(pcn.NovoElem("X509Certificate").Texto(s.X509Certificate))))
}

// ---------------------------------------------------------------------------
// Auxiliares
// ---------------------------------------------------------------------------

// nodeCNPJCPF e o AddNodeCNPJCPF do TACBrXmlWriter: ate 11 posicoes gera
// CPF (completado a 11), senao CNPJ (completado a 14); vazio da preferencia
// a CNPJ -- obrigatorio e vazio sai como 14 zeros, como no original.
func nodeCNPJCPF(valor string, obrigatorio bool) *pcn.Elem {
	valor = pcn.OnlyCPFCNPJAlphaNum(strings.TrimSpace(valor))
	tam := len(valor)

	if tam > 0 && tam <= 11 {
		return pcn.NodeStrSemFiltro("CPF", pcn.PreencherZerosEsquerda(valor, 11), obrigatorio)
	}
	if (obrigatorio || tam > 0) && tam != 14 {
		valor = pcn.PreencherZerosEsquerda(valor, 14)
	}
	return pcn.NodeStrSemFiltro("CNPJ", valor, obrigatorio)
}

// formatarCompetencia formata AAAAMM (FormatDateTime('yyyymm')); data zero
// devolve vazio (DIVERGENCIA 5: o Delphi produziria "189912").
func formatarCompetencia(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("200601")
}

// gerarCodigoDFe sorteia o cNF, rejeitando o proprio nNF e os codigos
// invalidos da NT 2019.001. Porte de GerarCodigoDFe/ValidarCodigoDFe
// (ACBrDFeUtil.pas) com 7 digitos -- DIVERGENCIA 6.
func gerarCodigoDFe(nNF int) int {
	for {
		c := rand.Intn(10000000)
		if validarCodigoDFe(c, nNF) {
			return c
		}
	}
}

var codigosDFeInvalidos7 = map[int]bool{
	0: true, 1111111: true, 2222222: true, 3333333: true, 4444444: true,
	5555555: true, 6666666: true, 7777777: true, 8888888: true, 9999999: true,
	1234567: true, 2345678: true, 3456789: true, 4567890: true, 5678901: true,
	6789012: true, 7890123: true, 8901234: true, 9012345: true, 123456: true,
}

func validarCodigoDFe(codigo, nNF int) bool {
	return codigo != nNF && !codigosDFeInvalidos7[codigo]
}
