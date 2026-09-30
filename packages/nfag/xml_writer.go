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

package nfag

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

// Geracao do XML da NFAg -- porte 1:1 de TNFAgXmlWriter
// (ACBrNFAg.XmlWriter.pas), gerador a gerador. Diferente da NFGas, o
// XmlWriter da NFAg ja converte todos os enums com XxxToStr -- nao ha a
// divergencia de "ordinal no XML" aqui.
//
// Divergencias deliberadas (mesma politica da NFGas, registradas em
// divergencias-acbr.md):
//
//  1. A raiz do documento processado e `nfagProc`, como no XSD
//     (procNFAg_v1.00.xsd). O ACBr gera e le `NFAgProc`.
//  2. Competencia (AAAAMM) com data zero gera tag vazia, nao "189912".
//  3. cNF aleatorio sorteado com 7 digitos (posicao 36 da chave e o
//     nSiteAutoriz).
//  4. IndDevolucao e pcn.IndicadorEx (zero-value seguro).
//
// OMISSOES DELIBERADAS: ListaDeAlertas/wAlerta, ValidarIE/ValidarMunicipio
// (alertas informativos) e NormatizarMunicipios nao foram portados;
// AjustarTagNro tambem nao (opcao de formatacao do endereco).

// GerarXML gera o XML da NFAg a partir dos dados do documento.
// Porte de TNFAgXmlWriter.GerarXml.
//
// Como no original, tem efeitos colaterais: monta a chave (sorteando cNF
// quando zero) e grava InfNFAg.ID ("NFAG"+chave, literal MAIUSCULO exigido
// pelo XSD), Ide.CDV e Ide.CNF. Se ProcNFAg.NProt estiver preenchido, o
// retorno e o documento processado. A Signature so e embutida quando
// completa (taSomenteSeAssinada); para assinar, use Assinar.
func GerarXML(n *NFAg) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}

	if n.Ide.CNF == 0 || n.Ide.CNF == -1 {
		n.Ide.CNF = gerarCodigoDFe(n.Ide.NNF)
	} else if n.Ide.CNF < 0 {
		n.Ide.CNF = 0
	}
	if n.Ide.Modelo == 0 {
		n.Ide.Modelo = ModeloNFAg
	}
	if n.InfNFAg.Versao == 0 {
		n.InfNFAg.Versao = 1.00
	}

	chave, err := MontarChaveAcesso(n)
	if err != nil {
		return "", err
	}
	n.InfNFAg.ID = LiteralChave + chave
	dv, _ := strconv.Atoi(chave[43:])
	n.Ide.CDV = dv
	cnf, _ := strconv.Atoi(chave[36:43])
	n.Ide.CNF = cnf

	w := &xmlWriter{n: n, chave: chave, rtc: rtc.NovoWriter(rtc.ModeloNFAg)}

	nota := pcn.NovoElem("NFAg").Filho(w.gerarInfNFAg())

	if n.InfNFAgSupl.QrCodNFAg != "" {
		nota.Filho(pcn.NovoElem("infNFAgSupl").
			Filho(pcn.NovoElem("qrCodNFAg").
				TextoBruto("<![CDATA[" + n.InfNFAgSupl.QrCodNFAg + "]]>")))
	}

	if n.Signature.Assinada() {
		nota.Filho(gerarSignature(&n.Signature))
	}

	if n.ProcNFAg.NProt != "" {
		// DIVERGENCIA 1: raiz nfagProc conforme o XSD (ACBr: NFAgProc).
		return pcn.NovoElem("nfagProc").
			Attr("versao", pcn.FormatarVersaoXML(n.InfNFAg.Versao)).
			Attr("xmlns", Namespace).
			Filho(nota).
			Filho(w.gerarProtNFAg()).XML(), nil
	}
	nota.Attr("xmlns", Namespace)
	return nota.XML(), nil
}

// GerarXMLProc gera o XML de nfagProc; exige protocolo preenchido.
func GerarXMLProc(n *NFAg) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}
	if n.ProcNFAg.NProt == "" {
		return "", ErrProtocoloAusente
	}
	return GerarXML(n)
}

// Assinar assina o XML da NFAg (elemento infNFAg) com o certificado A1.
func Assinar(cert *dfe.Certificado, xmlStr string) (string, error) {
	return dfe.AssinarXML(cert, xmlStr, "infNFAg")
}

// AssinarEvento assina o XML de eventoNFAg (elemento infEvento).
func AssinarEvento(cert *dfe.Certificado, xmlStr string) (string, error) {
	return dfe.AssinarXML(cert, xmlStr, "infEvento")
}

// xmlWriter carrega o estado da geracao de um documento (FChaveNFAg + o
// writer da RTC).
type xmlWriter struct {
	n     *NFAg
	chave string
	rtc   *rtc.Writer
}

func (w *xmlWriter) gerarInfNFAg() *pcn.Elem {
	n := w.n
	e := pcn.NovoElem("infNFAg").
		Attr("Id", LiteralChave+w.chave).
		Attr("versao", pcn.FormatarVersaoXML(n.InfNFAg.Versao)).
		Filho(w.gerarIde()).
		Filho(w.gerarEmit()).
		Filho(w.gerarDest()).
		// ligacao e SEMPRE gerada, sem condicao (GerarInfNFAg chama
		// GerarLigacao incondicionalmente)
		Filho(w.gerarLigacao()).
		Filho(w.gerarGSub())

	for _, m := range n.GMed {
		e.Filho(w.gerarGMed(m))
	}

	e.Filho(w.gerarGFatConjunto())

	for _, d := range n.Det {
		e.Filho(w.gerarDet(d))
	}

	e.Filho(w.gerarTotal(n.Total)).
		Filho(w.rtc.GerarPgtoVinc(n.PgtoVinc)).
		Filho(w.gerarGFat(n.GFat)).
		Filho(w.gerarGAgencia(n.GAgencia)).
		Filho(w.gerarGQualiAgua(n.GQualiAgua))

	for _, a := range n.AutXML {
		e.Filho(pcn.NovoElem("autXML").Filho(nodeCNPJCPF(a.CNPJCPF, true)))
	}

	e.Filho(w.gerarInfAdic(n.InfAdic)).
		Filho(w.gerarInfPAA(n.InfPAA)).
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
		Filho(pcn.NodeStr("finNFAg", ide.FinNFAg.String(), true)).
		Filho(pcn.NodeStr("tpFat", ide.TpFat.String(), true)).
		Filho(pcn.NodeStr("verProc", ide.VerProc, true))

	if !ide.DhCont.IsZero() || ide.XJust != "" {
		e.Filho(pcn.NodeStr("dhCont", pcn.FormatarDataHoraXML(ide.DhCont, uf), true)).
			Filho(pcn.NodeStr("xJust", ide.XJust, true))
	}

	e.Filho(w.rtc.GerarGCompraGovReduzido(ide.GCompraGov))

	// diferente da NFGas, aqui NAO ha if em volta: o AddNode e ocorrencia 0
	// e some sozinho quando TpaNenhum ("" e vazio)
	return e.Filho(pcn.NodeStr("tpPagAnt", ide.TpPagAnt.String(), false))
}

func (w *xmlWriter) gerarEmit() *pcn.Elem {
	emit := w.n.Emit
	// IE = 'ISENTO' vai crua; senao OnlyNumber (GerarEmit do original)
	ie := emit.IE
	if ie != "ISENTO" {
		ie = pcn.OnlyNumber(ie)
	}
	return pcn.NovoElem("emit").
		Filho(pcn.NodeStr("CNPJ", emit.CNPJ, true)).
		Filho(pcn.NodeStr("IE", ie, true)).
		Filho(pcn.NodeStr("xNome", emit.XNome, true)).
		Filho(pcn.NodeStr("xFant", emit.XFant, false)).
		Filho(w.gerarEndereco("enderEmit", emit.EnderEmit)).
		Filho(pcn.NodeStr("ISUFEmit", emit.ISUFEmit, false))
}

// homNomeDest e o nome forcado do destinatario em homologacao
// (HOM_NOME_DEST do GerarDest -- o literal diz "NF-E", copiado da NFe).
const homNomeDest = "NF-E EMITIDA EM AMBIENTE DE HOMOLOGACAO - SEM VALOR FISCAL"

func (w *xmlWriter) gerarDest() *pcn.Elem {
	dest := w.n.Dest
	e := pcn.NovoElem("dest")

	// em homologacao o xNome do destinatario e SUBSTITUIDO pelo texto fixo
	if w.n.Ide.TpAmb == pcn.TaProducao {
		e.Filho(pcn.NodeStr("xNome", dest.XNome, true))
	} else {
		e.Filho(pcn.NodeStr("xNome", homNomeDest, true))
	}

	if dest.IDOutros != "" {
		e.Filho(pcn.NodeStr("idOutros", dest.IDOutros, true))
	} else {
		e.Filho(nodeCNPJCPF(dest.CNPJCPF, true))
	}

	// IE so quando preenchida; ISENTO cru, senao OnlyNumber
	if dest.IE != "" {
		if dest.IE == "ISENTO" {
			e.Filho(pcn.NodeStr("IE", dest.IE, true))
		} else if strings.TrimSpace(dest.IE) != "" {
			e.Filho(pcn.NodeStr("IE", pcn.OnlyNumber(dest.IE), true))
		}
	}

	e.Filho(pcn.NodeStr("IM", dest.IM, false))

	if dest.CNIS != "" {
		e.Filho(pcn.NodeStr("cNIS", dest.CNIS, true))
	} else if dest.NB != "" {
		e.Filho(pcn.NodeStr("NB", dest.NB, true))
	}

	return e.Filho(pcn.NodeStr("xNomeAdicional", dest.XNomeAdicional, false)).
		Filho(w.gerarEndereco("enderDest", dest.EnderDest))
}

// gerarEndereco cobre enderEmit, enderDest e enderCorresp: mesmos campos e
// obrigatoriedades nos tres (CEP e obrigatorio em TODOS na NFAg).
func (w *xmlWriter) gerarEndereco(tag string, end Endereco) *pcn.Elem {
	return pcn.NovoElem(tag).
		Filho(pcn.NodeStr("xLgr", end.XLgr, true)).
		Filho(pcn.NodeStr("nro", end.Nro, true)).
		Filho(pcn.NodeStr("xCpl", end.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", end.XBairro, true)).
		Filho(pcn.NodeInt("cMun", end.CMun, 7, true)).
		Filho(pcn.NodeStr("xMun", end.XMun, true)).
		Filho(pcn.NodeInt("CEP", end.CEP, 8, true)).
		Filho(pcn.NodeStr("UF", end.UF, true)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(end.Fone), false)).
		Filho(pcn.NodeStr("email", end.Email, false))
}

func (w *xmlWriter) gerarLigacao() *pcn.Elem {
	l := w.n.Ligacao
	return pcn.NovoElem("ligacao").
		Filho(pcn.NodeStr("idLigacao", l.IDLigacao, true)).
		Filho(pcn.NodeStr("idCodCliente", l.IDCodCliente, false)).
		Filho(pcn.NodeStr("tpLigacao", l.TpLigacao.String(), true)).
		// lat/long sao OBRIGATORIOS e independentes (sem o par condicional
		// da NFGas)
		Filho(pcn.NodeStr("latGPS", l.LatGPS, true)).
		Filho(pcn.NodeStr("longGPS", l.LongGPS, true)).
		Filho(pcn.NodeStr("codRoteiroLeitura", l.CodRoteiroLeitura, false))
}

func (w *xmlWriter) gerarGSub() *pcn.Elem {
	gSub := w.n.GSub
	if gSub.ChNFAg == "" {
		return nil
	}
	return pcn.NovoElem("gSub").
		Filho(pcn.NodeStrSemFiltro("chNFAg", gSub.ChNFAg, true)).
		Filho(pcn.NodeStr("motSub", gSub.MotSub.String(), true))
}

func (w *xmlWriter) gerarGMed(m GMed) *pcn.Elem {
	return pcn.NovoElem("gMed").
		Attr("nMed", pcn.FormatarInteiroZeros(m.NMed, 2)).
		Filho(pcn.NodeStr("idMedidor", m.IDMedidor, true)).
		Filho(pcn.NodeDat("dMedAnt", m.DMedAnt, true)).
		Filho(pcn.NodeDat("dMedAtu", m.DMedAtu, true))
}

func (w *xmlWriter) gerarGFatConjunto() *pcn.Elem {
	g := w.n.GFatConjunto
	if g.ChNFAgFat == "" {
		return nil
	}
	return pcn.NovoElem("gFatConjunto").
		Filho(pcn.NodeStrSemFiltro("chNFAgFat", g.ChNFAgFat, true))
}

func (w *xmlWriter) gerarDet(d Det) *pcn.Elem {
	e := pcn.NovoElem("det").Attr("nItem", strconv.Itoa(d.NItem))
	if d.ChNFAgAnt != "" {
		e.Attr("chNFAgAnt", d.ChNFAgAnt)
	}
	if d.NItemAnt > 0 {
		e.Attr("nItemAnt", strconv.Itoa(d.NItemAnt))
	}

	for _, t := range d.GTarif {
		e.Filho(w.gerarGTarif(t))
	}

	return e.Filho(w.gerarProd(d.Prod)).
		Filho(w.gerarImposto(d.Imposto)).
		Filho(w.gerarGProcRef(d.GProcRef)).
		Filho(pcn.NodeStr("infAdProd", d.InfAdProd, false))
}

func (w *xmlWriter) gerarGTarif(t GTarif) *pcn.Elem {
	return pcn.NovoElem("gTarif").
		Filho(pcn.NodeDat("dIniTarif", t.DIniTarif, true)).
		Filho(pcn.NodeDat("dFimTarif", t.DFimTarif, false)).
		Filho(pcn.NodeStr("nAto", t.NAto, true)).
		Filho(pcn.NodeInt("anoAto", t.AnoAto, 4, true)).
		// ocorrencia 0 no original, mas o codigo do enum nunca e vazio --
		// a tag sai sempre, replicado
		Filho(pcn.NodeStr("tpFaixaCons", t.TpFaixaCons.String(), false))
}

func (w *xmlWriter) gerarProd(p Prod) *pcn.Elem {
	e := pcn.NovoElem("prod").
		Filho(pcn.NodeStr("indOrigemQtd", p.IndOrigemQtd.String(), true))

	// gMedicao apenas com nMed > 0 (sem o "e vMed > 0" da NFGas)
	if p.GMedicao.NMed > 0 {
		e.Filho(w.gerarGMedicao(p.GMedicao))
	}

	e.Filho(pcn.NodeStr("cProd", p.CProd, true)).
		Filho(pcn.NodeStr("xProd", p.XProd, true)).
		// cClass e Integer no ACBr, gravado com tcInt min 7 (PadLeft de
		// zeros); aqui o campo e string (DIVERGENCIA DE TIPO) e o pad
		// reproduz o tcInt
		Filho(pcn.NodeStrSemFiltro("cClass", pcn.PadLeftZeros(pcn.OnlyNumber(p.CClass), 7), true)).
		Filho(pcn.NodeStr("tpCategoria", p.TpCategoria.String(), true)).
		Filho(pcn.NodeStr("xCategoria", p.XCategoria, false)).
		Filho(pcn.NodeStr("qEconomias", p.QEconomias, false)).
		Filho(pcn.NodeStr("uMed", p.UMed.String(), true))

	// qFaturada fracionada sai tcDe4; inteira sai tcInt (Frac() > 0 do
	// original -- no ACBr o campo e Integer e o ramo De4 e codigo morto)
	if frac := p.QFaturada - math.Trunc(p.QFaturada); frac > 0 {
		e.Filho(pcn.NodeDec("qFaturada", p.QFaturada, 4, true))
	} else {
		e.Filho(pcn.NodeInt("qFaturada", int(p.QFaturada), 1, true))
	}

	// vItem/vProd sao gravados com tcDe2 ("pode ter 2 ou 10 casas" no
	// comentario do proprio ACBr) embora o LEITOR leia com tcDe10 --
	// assimetria do original, replicada
	e.Filho(pcn.NodeDec("vItem", p.VItem, 2, true))

	if frac := p.FatorPoluicao - math.Trunc(p.FatorPoluicao); frac > 0 {
		e.Filho(pcn.NodeDec("fatorPoluicao", p.FatorPoluicao, 4, true))
	} else {
		e.Filho(pcn.NodeInt("fatorPoluicao", int(p.FatorPoluicao), 1, true))
	}

	e.Filho(pcn.NodeDec("vProd", p.VProd, 2, true))

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
		Filho(pcn.NodeInt("nMed", g.NMed, 2, true))

	// gMedida quando ha leitura anterior OU atual; senao o motivo da nao
	// leitura (sem o xMotNaoLeitura da NFGas)
	if g.GMedida.VMedAnt > 0 || g.GMedida.VMedAtu > 0 {
		e.Filho(w.gerarGMedida(g.GMedida))
	} else {
		e.Filho(pcn.NodeStr("tpMotNaoLeitura", g.TpMotNaoLeitura.String(), true))
	}
	return e
}

func (w *xmlWriter) gerarGMedida(g GMedida) *pcn.Elem {
	return pcn.NovoElem("gMedida").
		Filho(pcn.NodeStr("tpGrMed", g.TpGrMed.String(), true)).
		Filho(pcn.NodeStr("nUnidConsumo", g.NUnidConsumo, false)).
		Filho(pcn.NodeDec("vUnidConsumo", g.VUnidConsumo, 2, false)).
		Filho(pcn.NodeStr("uMed", g.UMed.String(), true)).
		Filho(pcn.NodeDec("vMedAnt", g.VMedAnt, 2, true)).
		Filho(pcn.NodeDec("vMedAtu", g.VMedAtu, 2, true)).
		Filho(pcn.NodeDec("vConst", g.VConst, 2, true)).
		Filho(pcn.NodeDec("vMed", g.VMed, 2, true))
}

// gerarImposto -- o elemento imposto e SEMPRE gerado (sem condicao no
// original); os filhos e que sao condicionais.
func (w *xmlWriter) gerarImposto(imp Imposto) *pcn.Elem {
	return pcn.NovoElem("imposto").
		Filho(w.rtc.GerarIBSCBS(imp.IBSCBS)).
		Filho(w.gerarPIS(imp.PIS)).
		Filho(w.gerarCOFINS(imp.COFINS)).
		Filho(w.gerarRetTrib(imp.RetTrib)).
		Filho(w.gerarTFS(imp.TFS)).
		Filho(w.gerarTFU(imp.TFU))
}

func (w *xmlWriter) gerarPIS(p PIS) *pcn.Elem {
	if p.VBC == 0 && p.PPIS == 0 && p.VPIS == 0 {
		return nil
	}
	// pPIS e tcDe2 na GERACAO (o leitor le De4) -- assimetria do original
	return pcn.NovoElem("PIS").
		Filho(pcn.NodeStr("CST", p.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", p.VBC, 2, true)).
		Filho(pcn.NodeDec("pPIS", p.PPIS, 2, true)).
		Filho(pcn.NodeDec("vPIS", p.VPIS, 2, true))
}

func (w *xmlWriter) gerarCOFINS(c COFINS) *pcn.Elem {
	if c.VBC == 0 && c.PCOFINS == 0 && c.VCOFINS == 0 {
		return nil
	}
	return pcn.NovoElem("COFINS").
		Filho(pcn.NodeStr("CST", c.CST.String(), true)).
		Filho(pcn.NodeDec("vBC", c.VBC, 2, true)).
		Filho(pcn.NodeDec("pCOFINS", c.PCOFINS, 2, true)).
		Filho(pcn.NodeDec("vCOFINS", c.VCOFINS, 2, true))
}

func (w *xmlWriter) gerarRetTrib(r RetTrib) *pcn.Elem {
	if r.VRetPIS == 0 && r.VRetCOFINS == 0 && r.VRetCSLL == 0 &&
		r.VBCIRRF == 0 && r.VIRRF == 0 {
		return nil
	}
	// DIVERGENCIA: o ACBr grava vRetCOFINS (maiusculo) no item, mas o XSD
	// (nfagTiposBasico_v1.00.xsd:925) e o proprio LEITOR do ACBr usam
	// vRetCofins -- a tag maiuscula e invalida e o valor se perderia na
	// releitura. Grafia corrigida para a do XSD. vBCIRRF e GERADO e nunca
	// LIDO de volta (omissao do leitor do ACBr, replicada).
	return pcn.NovoElem("retTrib").
		Filho(pcn.NodeDec("vRetPIS", r.VRetPIS, 2, true)).
		Filho(pcn.NodeDec("vRetCofins", r.VRetCOFINS, 2, true)).
		Filho(pcn.NodeDec("vRetCSLL", r.VRetCSLL, 2, true)).
		Filho(pcn.NodeDec("vBCIRRF", r.VBCIRRF, 2, true)).
		Filho(pcn.NodeDec("vIRRF", r.VIRRF, 2, true))
}

func (w *xmlWriter) gerarTFS(t TFS) *pcn.Elem {
	if t.VBCTFS == 0 && t.PTFS == 0 && t.VTFS == 0 {
		return nil
	}
	return pcn.NovoElem("TFS").
		Filho(pcn.NodeDec("vBCTFS", t.VBCTFS, 2, true)).
		Filho(pcn.NodeDec("pTFS", t.PTFS, 2, true)).
		Filho(pcn.NodeDec("vTFS", t.VTFS, 2, true))
}

func (w *xmlWriter) gerarTFU(t TFU) *pcn.Elem {
	if t.VBCTFU == 0 && t.PTFU == 0 && t.VTFU == 0 {
		return nil
	}
	return pcn.NovoElem("TFU").
		Filho(pcn.NodeDec("vBCTFU", t.VBCTFU, 2, true)).
		Filho(pcn.NodeDec("pTFU", t.PTFU, 2, true)).
		Filho(pcn.NodeDec("vTFU", t.VTFU, 2, true))
}

func (w *xmlWriter) gerarGProcRef(g GProcRef) *pcn.Elem {
	// na NFAg o grupo so sai com vItem > 0 (condicao unica, diferente da
	// NFGas)
	if g.VItem <= 0 {
		return nil
	}
	// vItem/vProd tcDe2 na geracao ("pode ter 2 ou 6 casas" no comentario
	// do ACBr; o leitor le De8) -- assimetria replicada
	e := pcn.NovoElem("gProcRef").
		Filho(pcn.NodeDec("vItem", g.VItem, 2, true))

	if frac := g.QFaturada - math.Trunc(g.QFaturada); frac > 0 {
		e.Filho(pcn.NodeDec("qFaturada", g.QFaturada, 4, true))
	} else {
		e.Filho(pcn.NodeInt("qFaturada", int(g.QFaturada), 1, true))
	}

	e.Filho(pcn.NodeDec("vProd", g.VProd, 2, true))

	if g.IndDevolucao == pcn.TieSim {
		e.Filho(pcn.NodeStr("indDevolucao", "1", true))
	}
	for _, p := range g.GProc {
		e.Filho(pcn.NovoElem("gProc").
			Filho(pcn.NodeStr("tpProc", p.TpProc.String(), true)).
			Filho(pcn.NodeStr("nProcesso", p.NProcesso, true)))
	}
	return e
}

func (w *xmlWriter) gerarTotal(t Total) *pcn.Elem {
	return pcn.NovoElem("total").
		Filho(pcn.NodeDec("vProd", t.VProd, 2, true)).
		Filho(pcn.NovoElem("vRetTribTot").
			Filho(pcn.NodeDec("vRetPIS", t.VRetPIS, 2, true)).
			Filho(pcn.NodeDec("vRetCofins", t.VRetCOFINS, 2, true)).
			Filho(pcn.NodeDec("vRetCSLL", t.VRetCSLL, 2, true)).
			Filho(pcn.NodeDec("vIRRF", t.VIRRF, 2, true))).
		Filho(pcn.NodeDec("vCOFINS", t.VCOFINS, 2, true)).
		Filho(pcn.NodeDec("vPIS", t.VPIS, 2, true)).
		Filho(pcn.NodeDec("vTFS", t.VTFS, 2, true)).
		Filho(pcn.NodeDec("vTFU", t.VTFU, 2, true)).
		Filho(pcn.NodeDec("vNF", t.VNF, 2, true)).
		Filho(w.rtc.GerarIBSCBSTot(t.IBSCBSTot)).
		Filho(pcn.NodeDec("vTotDFe", t.VTotDFe, 2, false))
}

func (w *xmlWriter) gerarGFat(g GFat) *pcn.Elem {
	// gFat e SEMPRE gerado na NFAg (sem o guard dVencFat da NFGas)
	e := pcn.NovoElem("gFat").
		Filho(pcn.NodeStr("CompetFat", formatarCompetencia(g.CompetFat), true)).
		Filho(pcn.NodeDat("dVencFat", g.DVencFat, true)).
		Filho(pcn.NodeDat("dApresFat", g.DApresFat, false)).
		Filho(pcn.NodeDat("dProxLeitura", g.DProxLeitura, true)).
		Filho(pcn.NodeStr("nFat", g.NFat, false)).
		Filho(pcn.NodeStr("codBarras", g.CodBarras, true))

	if g.CodDebAuto != "" {
		e.Filho(pcn.NodeStr("codDebAuto", g.CodDebAuto, true))
	} else {
		// no else, codBanco e codAgencia sao OBRIGATORIOS (tags vazias se
		// nao informados), sem o "em conjunto" da NFGas
		e.Filho(pcn.NodeStr("codBanco", g.CodBanco, true)).
			Filho(pcn.NodeStr("codAgencia", g.CodAgencia, true))
	}

	if g.EnderCorresp.XLgr != "" {
		e.Filho(w.gerarEndereco("enderCorresp", g.EnderCorresp))
	}
	if g.GPIX.URLQRCodePIX != "" {
		e.Filho(pcn.NovoElem("gPIX").
			Filho(pcn.NodeStr("urlQRCodePIX", g.GPIX.URLQRCodePIX, true)))
	}
	return e
}

func (w *xmlWriter) gerarGAgencia(g GAgencia) *pcn.Elem {
	// gAgencia e SEMPRE gerada na NFAg
	e := pcn.NovoElem("gAgencia").
		Filho(pcn.NodeStr("econ", g.Econ, false)).
		Filho(pcn.NodeStr("econAcumulada", g.EconAcumulada, false)).
		Filho(pcn.NodeStr("sPrestador", g.SPrestador, false)).
		Filho(pcn.NodeDat("dEmissSelo", g.DEmissSelo, false)).
		Filho(pcn.NodeStr("sRegulador", g.SRegulador, false)).
		Filho(pcn.NodeStr("nAgenciaAtend", g.NAgenciaAtend, true)).
		Filho(pcn.NodeStr("enderAgenciaAtend", g.EnderAgenciaAtend, true))

	for _, h := range g.GHistCons {
		e.Filho(w.gerarGHistCons(h))
	}
	return e
}

func (w *xmlWriter) gerarGHistCons(h GHistCons) *pcn.Elem {
	e := pcn.NovoElem("gHistCons").
		Filho(pcn.NodeStr("xHistorico", h.XHistorico, true))
	for _, c := range h.GCons {
		e.Filho(pcn.NovoElem("gCons").
			Filho(pcn.NodeStr("CompetFat", formatarCompetencia(c.CompetFat), true)).
			Filho(pcn.NodeStr("uMed", c.UMed.String(), true)).
			// qtdDias e STRING no leiaute (min 5 no AddNode)
			Filho(pcn.NodeStr("qtdDias", c.QtdDias, true)).
			Filho(pcn.NodeDec("medDiaria", c.MedDiaria, 4, false)).
			// consumo tcDe2 na geracao (leitor De4) -- assimetria replicada
			Filho(pcn.NodeDec("consumo", c.Consumo, 2, false)).
			Filho(pcn.NodeDec("volFat", c.VolFat, 2, true)))
	}
	// medMensal tcDe2 na geracao (leitor De4) -- assimetria replicada
	return e.Filho(pcn.NodeDec("medMensal", h.MedMensal, 2, true))
}

func (w *xmlWriter) gerarGQualiAgua(g GQualiAgua) *pcn.Elem {
	if g.CompetAnalise.IsZero() {
		return nil
	}
	e := pcn.NovoElem("gQualiAgua").
		Filho(pcn.NodeStr("CompetAnalise", formatarCompetencia(g.CompetAnalise), true))

	for _, a := range g.GAnalise {
		e.Filho(pcn.NovoElem("gAnalise").
			Filho(pcn.NodeStr("xItemAnalisado", a.XItemAnalisado, true)).
			Filho(pcn.NodeStr("nAmostraMinima", a.NAmostraMinima, false)).
			Filho(pcn.NodeStr("nAmostraAnalisada", a.NAmostraAnalisada, false)).
			Filho(pcn.NodeStr("nAmostraFPadrao", a.NAmostraFPadrao, false)).
			Filho(pcn.NodeStr("nAmostraDPadrao", a.NAmostraDPadrao, false)).
			Filho(pcn.NodeStr("nMediaMensal", a.NMediaMensal, false)).
			Filho(pcn.NodeStr("xValorReferencia", a.XValorReferencia, false)))
	}

	return e.Filho(pcn.NodeStr("Conclusao", g.Conclusao, false)).
		Filho(pcn.NodeStr("cProcesso", g.CProcesso, false)).
		Filho(pcn.NodeStr("SistemaAbast", g.SistemaAbast, false))
}

func (w *xmlWriter) gerarInfAdic(i InfAdic) *pcn.Elem {
	// diferente da NFGas, o infAdic so sai quando ha conteudo
	infCpl := ""
	if len(i.InfCpl) > 0 {
		infCpl = i.InfCpl[0]
	}
	if strings.TrimSpace(i.InfAdFisco) == "" && strings.TrimSpace(infCpl) == "" {
		return nil
	}
	return pcn.NovoElem("infAdic").
		Filho(pcn.NodeStr("infAdFisco", i.InfAdFisco, false)).
		Filho(pcn.NodeStr("infCpl", infCpl, false))
}

func (w *xmlWriter) gerarInfPAA(p InfPAA) *pcn.Elem {
	if p.CNPJPAA == "" {
		return nil
	}
	return pcn.NovoElem("infPAA").
		Filho(pcn.NodeStrSemFiltro("CNPJPAA", p.CNPJPAA, true))
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

	if rt.IDCSRT != 0 && rt.HashCSRT != "" {
		e.Filho(pcn.NodeInt("idCSRT", rt.IDCSRT, 3, true)).
			Filho(pcn.NodeStrSemFiltro("hashCSRT", rt.HashCSRT, true))
	}
	return e
}

func (w *xmlWriter) gerarProtNFAg() *pcn.Elem {
	p := w.n.ProcNFAg
	uf := pcn.SiglaUF(w.n.Ide.CUF)
	// Gerar_ProtNFAg usa AddChild+Content direto: todas as tags saem,
	// mesmo vazias, sem FiltrarTextoXML.
	return pcn.NovoElem("protNFAg").
		Attr("versao", pcn.FormatarVersaoXML(w.n.InfNFAg.Versao)).
		Filho(pcn.NovoElem("infProt").
			Filho(pcn.NovoElem("tpAmb").Texto(p.TpAmb.String())).
			Filho(pcn.NovoElem("verAplic").Texto(p.VerAplic)).
			Filho(pcn.NovoElem("chNFAg").Texto(p.ChDFe)).
			Filho(pcn.NovoElem("dhRecbto").Texto(pcn.FormatarDataHoraXML(p.DhRecbto, uf))).
			Filho(pcn.NovoElem("nProt").Texto(p.NProt)).
			Filho(pcn.NovoElem("digVal").Texto(p.DigVal)).
			Filho(pcn.NovoElem("cStat").Texto(strconv.Itoa(p.CStat))).
			Filho(pcn.NovoElem("xMotivo").Texto(p.XMotivo)))
}

// gerarSignature reembute um bloco Signature ja existente.
// Porte do template TSignature.GerarXML (identico ao da NFGas).
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

// nodeCNPJCPF e o AddNodeCNPJCPF do TACBrXmlWriter (mesma semantica portada
// na NFGas).
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

// formatarCompetencia formata AAAAMM; data zero devolve vazio
// (DIVERGENCIA 2: o Delphi produziria "189912").
func formatarCompetencia(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("200601")
}

// gerarCodigoDFe sorteia o cNF com 7 digitos (DIVERGENCIA 3), rejeitando o
// proprio nNF e os codigos invalidos. Porte de GerarCodigoDFe/
// ValidarCodigoDFe (ACBrDFeUtil.pas).
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
