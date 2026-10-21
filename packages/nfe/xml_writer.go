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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Gerador do XML da NF-e. Porte de TNFeXmlWriter
// (Base/ACBrNFe.XmlWriter.pas), grupo a grupo e na mesma ordem.
//
// # O que governa a saida
//
// A versao do leiaute (infNFe.Versao) decide, sozinha, dezenas de campos e
// ate a POSICAO de um grupo inteiro: o NFref sai antes do tpImp abaixo da
// versao 3 e depois do xJust a partir dela. O modelo (55 x 65) decide outros.
// Cada condicional esta marcada com a linha do .pas.
//
// # Efeitos colaterais, replicados de proposito
//
// Como no original, GerarXML MUTA o documento: monta a chave de acesso
// (sorteando cNF quando zero) e grava infNFe.ID, ide.cDV e ide.cNF. O
// GerarDest ainda sobrescreve dest.xNome em homologacao e forca indIEDest
// abaixo da versao 3.10.
//
// # Omissoes deliberadas
//
// A ListaDeAlertas (wAlerta) nao foi portada: no ACBr ela so acumula avisos
// e nao impede a geracao, e a validacao efetiva e o XSD da SEFAZ mais o
// regras_negocio.go. Pelo mesmo motivo ficaram de fora ValidarIE e
// ValidarISUF, que so rodam sob a opcao ValidarInscricoes (default False).
//
// NormatizarMunicipios e caso diferente: ela MUDA a saida, preenchendo cMun a
// partir do xMun e vice-versa. O que torna a omissao segura e o default ser
// False no ACBr (XmlWriter.pas:215) -- liga-la exigiria a tabela de
// municipios em disco, que o package nao carrega.

// Opcoes reproduz as TNFeXmlWriterOptions que mudam a SAIDA do gerador.
//
// As opcoes do ACBr que so controlam alerta ou normalizacao de municipio nao
// foram portadas -- ver "Omissoes deliberadas" acima.
type Opcoes struct {
	// FatCamposOpcionais DESLIGA a obrigatoriedade de nFat, vOrig, vDesc e
	// vLiq no grupo fat.
	//
	// A opcao esta invertida em relacao ao ACBr de proposito: la o campo se
	// chama CamposFatObrigatorios e o Create o liga por padrao
	// (XmlWriter.pas:220). Em Go o zero value de um bool e false, entao
	// manter o mesmo nome faria Opcoes{} gerar DIFERENTE do ACBr. Invertido,
	// o zero value reproduz o default do original.
	//
	// So tem efeito a partir da versao 4.00 (XmlWriter.pas:3657).
	FatCamposOpcionais bool

	// AjustarTagNro troca um "nro" vazio por "S/N", como o
	// ExecutarAjusteTagNro do ACBr.
	AjustarTagNro bool

	// IDCSRT e CSRT habilitam o grupo de hash do responsavel tecnico. O
	// hashCSRT so e emitido quando os dois estao preenchidos
	// (XmlWriter.pas:4153).
	IDCSRT int
	CSRT   string
}

// xmlWriter carrega o estado que o TNFeXmlWriter mantem entre os grupos.
type xmlWriter struct {
	n      *NFe
	opcoes Opcoes
	chave  string
	rtc    *rtc.Writer

	// casasValor4ou2 e casasValor10ou4 sao o FormatoValor4ou2 e o
	// FormatoValor10ou4 do ACBr (XmlWriter.pas:299-309): a precisao de
	// alguns campos do ITEM depende da versao do leiaute. Ficam aqui porque
	// quem os define e o GerarXml; quem os consome e o gerador do det.
	casasValor4ou2  int
	casasValor10ou4 int
}

// versao devolve a versao do leiaute, que governa quase toda condicional.
func (w *xmlWriter) versao() float64 { return w.n.InfNFe.Versao }

// ehNFe diz se o documento e NF-e (modelo 55); o contrario e NFC-e (65).
func (w *xmlWriter) ehNFe() bool { return w.n.Ide.Modelo == 55 }

// GerarXML gera o XML da NF-e a partir dos dados do documento, com as opcoes
// padrao. Porte de TNFeXmlWriter.GerarXml.
func GerarXML(n *NFe) (string, error) {
	return GerarXMLComOpcoes(n, Opcoes{})
}

// GerarXMLComOpcoes gera o XML da NF-e com as opcoes informadas.
//
// MUTA o documento, como o original: monta a chave de acesso e grava
// infNFe.ID, ide.cDV e ide.cNF. Quando procNFe.nProt esta preenchido, o
// retorno e o documento processado (raiz nfeProc com protNFe); senao, a NF-e
// solta.
func GerarXMLComOpcoes(n *NFe, opcoes Opcoes) (string, error) {
	if n == nil {
		return "", ErrDocumentoNulo
	}

	// MontarChaveAcesso normaliza o cNF (sorteia ou zera) antes de montar.
	chave, err := MontarChaveAcesso(n)
	if err != nil {
		return "", err
	}
	n.InfNFe.ID = "NFe" + chave

	// O digito vem do ULTIMO caractere, como no ExtrairDigitoChaveAcesso do
	// ACBr -- nao de um offset fixo, que silenciaria uma chave malformada.
	dv, err := strconv.Atoi(chave[len(chave)-1:])
	if err != nil {
		return "", fmt.Errorf("nfe: digito verificador invalido na chave %q: %w", chave, err)
	}
	n.Ide.CDV = dv

	w := &xmlWriter{n: n, opcoes: opcoes, chave: chave}

	switch n.Ide.Modelo {
	case 65:
		w.rtc = rtc.NovoWriter(rtc.ModeloNFCe)
	default:
		w.rtc = rtc.NovoWriter(rtc.ModeloNFe)
	}

	// Precisao por versao (XmlWriter.pas:299-309).
	if w.versao() >= 3.10 {
		w.casasValor4ou2 = 4
	} else {
		w.casasValor4ou2 = 2
	}
	if w.versao() >= 2 {
		w.casasValor10ou4 = 10
	} else {
		w.casasValor10ou4 = 4
	}

	nota := pcn.NovoElem("NFe").
		Attr("xmlns", namespaceNFe).
		Filho(w.gerarInfNFe())

	w.anexarInfNFeSupl(nota)

	// So embute assinatura COMPLETA, como o taSomenteSeAssinada exige
	// (default do ACBr): Digest + SignatureValue + X509Certificate.
	//
	// O URI e REESCRITO a partir da chave recem-montada, como no .pas:373.
	// Sem isso, uma nota cujo cNF foi re-sorteado ficaria com o Reference
	// apontando para a chave antiga -- assinatura invalida em silencio.
	if n.Signature.Assinada() {
		n.Signature.URI = "#NFe" + pcn.RemoverLiteralChave(n.InfNFe.ID)
		nota.Filho(gerarSignature(&n.Signature))
	}

	if n.ProcNFe.NProt != "" {
		proc := pcn.NovoElem("nfeProc").
			Attr("xmlns", namespaceNFe).
			Attr("versao", pcn.FormatarVersaoXML(w.versao())).
			Filho(nota).
			Filho(w.gerarProtNFe())
		return proc.XML(), nil
	}

	return nota.XML(), nil
}

// namespaceNFe e o namespace do portal fiscal, igual em todos os DFe.
const namespaceNFe = "http://www.portalfiscal.inf.br/nfe"

// anexarInfNFeSupl acrescenta o grupo suplementar da NFC-e (QR-Code).
//
// O CDATA e CONDICIONAL: o ACBr so envolve o qrCode quando ele NAO contem
// "?p=" (XmlWriter.pas:345). Um QR-Code ja montado com a query string sai
// cru; um texto qualquer sai embrulhado.
func (w *xmlWriter) anexarInfNFeSupl(nota *pcn.Elem) {
	qr := w.n.InfNFeSupl.QRCode
	if qr == "" {
		return
	}

	if !strings.Contains(qr, "?p=") {
		qr = "<![CDATA[" + qr + "]]>"
	}

	supl := pcn.NovoElem("infNFeSupl").
		Filho(pcn.NovoElem("qrCode").TextoBruto(qr))

	if w.versao() >= 4 {
		supl.Filho(pcn.NodeStr("urlChave", w.n.InfNFeSupl.URLChave, true))
	}
	nota.Filho(supl)
}

// gerarInfNFe monta o corpo da nota. Porte de GerarInfNFe.
func (w *xmlWriter) gerarInfNFe() *pcn.Elem {
	n := w.n

	inf := pcn.NovoElem("infNFe").
		Attr("Id", n.InfNFe.ID).
		Attr("versao", pcn.FormatarVersaoXML(w.versao())).
		Filho(w.gerarIde()).
		Filho(w.gerarEmit()).
		Filho(w.gerarAvulsa())

	// dest sai sempre na NF-e; na NFC-e, so com identificacao do
	// destinatario ou operacao para o exterior (XmlWriter.pas:398).
	if n.Dest.CNPJCPF != "" || n.Dest.IDEstrangeiro != "" ||
		n.Dest.EnderDest.UF == UFExterior || n.Ide.Modelo != 65 {
		inf.Filho(w.gerarDest())
	}

	inf.Filho(w.gerarRetirada()).Filho(w.gerarEntrega())

	for _, a := range n.AutXML {
		inf.Filho(gerarAutXML(a))
	}

	// O grupo det entra na etapa 4b.
	for i := range n.Det {
		inf.Filho(w.gerarDet(i))
	}

	inf.Filho(w.gerarTotal()).
		Filho(w.gerarTransp()).
		Filho(w.gerarCobr())

	// pag so existe na NFC-e a partir da 3, e em qualquer modelo a partir
	// da 4 (XmlWriter.pas:421).
	if (w.versao() >= 3 && n.Ide.Modelo != 55) || w.versao() >= 4 {
		for _, p := range w.gerarPag() {
			inf.Filho(p)
		}
	}

	inf.Filho(w.gerarInfIntermed()).
		Filho(w.gerarInfAdic()).
		Filho(w.gerarExporta()).
		Filho(w.gerarCompra()).
		Filho(w.gerarCana())

	if w.versao() >= 4 {
		inf.Filho(w.gerarInfRespTec())
	}

	inf.Filho(w.gerarAgropecuario())

	return inf
}

// gerarIde monta a identificacao. Porte de GerarIde.
//
// Atencao a POSICAO do NFref: abaixo da versao 3 ele sai logo depois do
// cMunFG; a partir dela, so depois do xJust. E o mesmo grupo em dois lugares
// diferentes do XML, decidido pela versao (XmlWriter.pas:523 e :591).
func (w *xmlWriter) gerarIde() *pcn.Elem {
	n := w.n
	ide := pcn.NovoElem("ide")

	ide.Filho(pcn.NodeInt("cUF", n.Ide.CUF, 2, true)).
		Filho(pcn.NodeInt("cNF", n.Ide.CNF, 8, true)).
		Filho(pcn.NodeStr("natOp", n.Ide.NatOp, true))

	// indPag saiu do leiaute na 4.00.
	if w.versao() < 4 {
		ide.Filho(pcn.NodeStr("indPag", n.Ide.IndPag.String(), true))
	}

	ide.Filho(pcn.NodeInt("mod", n.Ide.Modelo, 2, true)).
		Filho(pcn.NodeInt("serie", n.Ide.Serie, 1, true)).
		Filho(pcn.NodeInt("nNF", n.Ide.NNF, 1, true))

	if w.versao() >= 3 {
		ide.Filho(pcn.NodeStr("dhEmi",
			pcn.FormatarDataHoraXML(n.Ide.DEmi, pcn.SiglaUF(n.Ide.CUF)), true))

		// dhSaiEnt so na NF-e, e so com data informada.
		if n.Ide.Modelo == 55 && !n.Ide.DSaiEnt.IsZero() {
			ide.Filho(pcn.NodeStr("dhSaiEnt",
				pcn.FormatarDataHoraXML(n.Ide.DSaiEnt, pcn.SiglaUF(n.Ide.CUF)), false))
		}
	} else {
		ide.Filho(pcn.NodeDat("dEmi", n.Ide.DEmi, true)).
			Filho(pcn.NodeDat("dSaiEnt", n.Ide.DSaiEnt, false))
		// hSaiEnt depende da DATA de saida, nao da hora.
		if !n.Ide.DSaiEnt.IsZero() {
			ide.Filho(pcn.NodeStr("hSaiEnt", formatarHora(n.Ide.HSaiEnt), false))
		}
	}

	ide.Filho(pcn.NodeDat("dPrevEntrega", n.Ide.DPrevEntrega, false)).
		Filho(pcn.NodeStr("tpNF", n.Ide.TpNF.String(), true))

	if w.versao() >= 3 {
		ide.Filho(pcn.NodeStr("idDest", n.Ide.IdDest.String(), true))
	}

	ide.Filho(pcn.NodeInt("cMunFG", n.Ide.CMunFG, 7, true))

	if n.Ide.CMunFGIBS > 0 {
		ide.Filho(pcn.NodeInt("cMunFGIBS", n.Ide.CMunFGIBS, 7, false))
	}

	// PRIMEIRA posicao possivel do NFref: so abaixo da versao 3.
	if w.versao() < 3 {
		for i := range n.Ide.NFref {
			ide.Filho(w.gerarNFref(i))
		}
	}

	ide.Filho(pcn.NodeStr("tpImp", n.Ide.TpImp.String(), true)).
		Filho(pcn.NodeStr("tpEmis", n.Ide.TpEmis.String(), true)).
		Filho(pcn.NodeInt("cDV", n.Ide.CDV, 1, true)).
		Filho(pcn.NodeStr("tpAmb", n.Ide.TpAmb.String(), true)).
		Filho(pcn.NodeStr("finNFe", n.Ide.FinNFe.String(), true))

	// Reforma Tributaria: cada tipo so sai com a finalidade correspondente.
	if w.versao() >= 4 {
		if n.Ide.FinNFe == FnDebito {
			ide.Filho(pcn.NodeStr("tpNFDebito", n.Ide.TpNFDebito.String(), false))
		}
		if n.Ide.FinNFe == FnCredito {
			ide.Filho(pcn.NodeStr("tpNFCredito", n.Ide.TpNFCredito.String(), false))
		}
	}

	if w.versao() >= 3 {
		ide.Filho(pcn.NodeStr("indFinal", n.Ide.IndFinal.String(), true)).
			Filho(pcn.NodeStr("indPres", n.Ide.IndPres.String(), true))
	}

	if w.versao() >= 4 {
		ide.Filho(pcn.NodeStr("indIntermed", n.Ide.IndIntermed.String(), false)).
			Filho(nodeNumStr("cIndOp", n.Ide.CIndOp, 6, false))
	}

	ide.Filho(pcn.NodeStr("procEmi", n.Ide.ProcEmi.String(), true)).
		Filho(pcn.NodeStr("verProc", n.Ide.VerProc, true))

	// O par dhCont/xJust sai junto, e basta UM dos dois estar preenchido.
	if !n.Ide.DhCont.IsZero() || n.Ide.XJust != "" {
		if w.versao() >= 3 {
			ide.Filho(pcn.NodeStr("dhCont",
				pcn.FormatarDataHoraXML(n.Ide.DhCont, pcn.SiglaUF(n.Ide.CUF)), true))
		} else {
			// Abaixo da 3 o campo tem 19 posicoes: sem fuso.
			ide.Filho(pcn.NodeStr("dhCont", formatarDataHoraSemFuso(n.Ide.DhCont), true))
		}
		ide.Filho(pcn.NodeStr("xJust", n.Ide.XJust, true))
	}

	// SEGUNDA posicao possivel do NFref: da versao 3 em diante.
	if w.versao() >= 3 {
		for i := range n.Ide.NFref {
			ide.Filho(w.gerarNFref(i))
		}
	}

	ide.Filho(w.rtc.GerarGCompraGov(n.Ide.GCompraGov)).
		Filho(w.rtc.GerarGPagAntecipado(n.Ide.GPagAntecipado))

	return ide
}

// gerarNFref monta uma nota referenciada. Porte de GerarIdeNFref.
//
// Os cinco subgrupos sao mutuamente exclusivos no leiaute, mas o ACBr nao os
// trata como tal: avalia os cinco guards e anexa todos os que passarem.
func (w *xmlWriter) gerarNFref(i int) *pcn.Elem {
	ref := w.n.Ide.NFref[i]
	e := pcn.NovoElem("NFref")

	// refNFe e refNFeSig compartilham a posicao: vale o primeiro preenchido.
	if ref.RefNFe != "" {
		e.Filho(pcn.NodeStr("refNFe", pcn.RemoverLiteralChave(ref.RefNFe), true))
	} else if ref.RefNFeSig != "" {
		e.Filho(pcn.NodeStr("refNFeSig", pcn.RemoverLiteralChave(ref.RefNFeSig), true))
	}

	if ref.RefNF.NNF > 0 {
		e.Filho(pcn.NovoElem("refNF").
			Filho(pcn.NodeInt("cUF", ref.RefNF.CUF, 2, true)).
			Filho(pcn.NodeStr("AAMM", ref.RefNF.AAMM, true)).
			Filho(pcn.NodeStr("CNPJ", pcn.OnlyCPFCNPJAlphaNum(ref.RefNF.CNPJ), true)).
			Filho(pcn.NodeInt("mod", ref.RefNF.Modelo, 2, true)).
			Filho(pcn.NodeInt("serie", ref.RefNF.Serie, 1, true)).
			Filho(pcn.NodeInt("nNF", ref.RefNF.NNF, 1, true)))
	}

	if ref.RefNFP.NNF > 0 {
		e.Filho(pcn.NovoElem("refNFP").
			Filho(pcn.NodeInt("cUF", ref.RefNFP.CUF, 2, true)).
			Filho(pcn.NodeStr("AAMM", ref.RefNFP.AAMM, true)).
			Filho(gerarCNPJCPF(ref.RefNFP.CNPJCPF, true, true)).
			Filho(pcn.NodeStr("IE", ref.RefNFP.IE, true)).
			// O campo e string na classe, mas o ACBr o grava com tcInt
			// (XmlWriter.pas:693): "1" vira "01". Normalizado.
			Filho(nodeIntDeTexto("mod", ref.RefNFP.Modelo, 2, true)).
			Filho(pcn.NodeInt("serie", ref.RefNFP.Serie, 1, true)).
			Filho(pcn.NodeInt("nNF", ref.RefNFP.NNF, 1, true)))
	}

	if ref.RefCTe != "" {
		e.Filho(pcn.NodeStr("refCTe", pcn.RemoverLiteralChave(ref.RefCTe), true))
	}

	if ref.RefECF.NCOO != "" {
		// nECF e nCOO sao string na classe e o ACBr os grava com tcInt,
		// que aplica zero a esquerda: "1" vira "001" e "123" vira "000123".
		e.Filho(pcn.NovoElem("refECF").
			Filho(pcn.NodeStr("mod", ref.RefECF.Modelo.String(), true)).
			Filho(nodeIntDeTexto("nECF", ref.RefECF.NECF, 3, true)).
			Filho(nodeIntDeTexto("nCOO", ref.RefECF.NCOO, 6, true)))
	}

	return e
}

// gerarEmit monta o emitente. Porte de GerarEmit.
func (w *xmlWriter) gerarEmit() *pcn.Elem {
	e := w.n.Emit
	emit := pcn.NovoElem("emit").
		Filho(gerarCNPJCPF(e.CNPJCPF, true, true)).
		Filho(pcn.NodeStr("xNome", e.XNome, true)).
		Filho(pcn.NodeStr("xFant", e.XFant, false)).
		Filho(w.gerarEnderEmit())

	// "ISENTO" e literal e passa inteiro; qualquer outro valor perde a
	// formatacao.
	if e.IE == "ISENTO" {
		emit.Filho(pcn.NodeStr("IE", e.IE, true))
	} else {
		emit.Filho(pcn.NodeStr("IE", pcn.OnlyNumber(e.IE), true))
	}

	emit.Filho(pcn.NodeStr("IEST", e.IEST, false)).
		Filho(pcn.NodeStr("IM", e.IM, false))

	// NT 2013/005 v1.03: o CNAE so pode ser informado junto do IM.
	if e.IM != "" {
		emit.Filho(pcn.NodeStr("CNAE", e.CNAE, false))
	}

	if w.versao() >= 2 {
		emit.Filho(pcn.NodeStr("CRT", e.CRT.String(), true))
	}

	emit.Filho(pcn.NodeStr("ISUFEmit", e.ISUFEmit, false))

	return emit
}

// gerarEnderEmit monta o endereco do emitente. Porte de GerarEmitEnderEmit.
//
// O cPais sai SEMPRE como 1058: o ACBr grava a constante CODIGO_BRASIL e
// ignora o campo do modelo (XmlWriter.pas:789). O emitente de NF-e e sempre
// brasileiro, entao a constante nao perde dado -- replicado.
func (w *xmlWriter) gerarEnderEmit() *pcn.Elem {
	end := w.n.Emit.EnderEmit
	uf, mun, cMun := ajustarMunicipioUF(codigoPaisPadrao, end.UF, end.XMun, end.CMun)

	return pcn.NovoElem("enderEmit").
		Filho(pcn.NodeStr("xLgr", end.XLgr, true)).
		Filho(pcn.NodeStr("nro", w.ajustarNro(end.Nro), true)).
		Filho(pcn.NodeStr("xCpl", end.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", end.XBairro, true)).
		Filho(pcn.NodeInt("cMun", cMun, 1, true)).
		Filho(pcn.NodeStr("xMun", mun, true)).
		Filho(pcn.NodeStr("UF", uf, true)).
		Filho(pcn.NodeInt("CEP", end.CEP, 8, true)).
		Filho(pcn.NodeInt("cPais", codigoPaisPadrao, 4, false)).
		Filho(pcn.NodeStr("xPais", end.XPais, false)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(end.Fone), false))
}

// gerarAvulsa monta o grupo do fisco emitente. Porte de GerarAvulsa.
func (w *xmlWriter) gerarAvulsa() *pcn.Elem {
	a := w.n.Avulsa
	if strings.TrimSpace(a.CNPJ) == "" {
		return nil
	}

	return pcn.NovoElem("avulsa").
		Filho(pcn.NodeStr("CNPJ", a.CNPJ, true)).
		Filho(pcn.NodeStr("xOrgao", a.XOrgao, true)).
		Filho(pcn.NodeStr("matr", a.Matr, true)).
		Filho(pcn.NodeStr("xAgente", a.XAgente, true)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(a.Fone), false)).
		Filho(pcn.NodeStr("UF", a.UF, true)).
		Filho(pcn.NodeStr("nDAR", a.NDAR, false)).
		Filho(pcn.NodeDat("dEmi", a.DEmi, false)).
		Filho(pcn.NodeDec("vDAR", a.VDAR, 2, false)).
		Filho(pcn.NodeStr("repEmi", a.RepEmi, true)).
		Filho(pcn.NodeDat("dPag", a.DPag, false))
}

// nomeDestHomologacao e o literal que a SEFAZ exige no destinatario de uma
// nota emitida em homologacao.
const nomeDestHomologacao = "NF-E EMITIDA EM AMBIENTE DE HOMOLOGACAO - SEM VALOR FISCAL"

// gerarDest monta o destinatario. Porte de GerarDest.
//
// MUTA o documento em dois pontos, como o original:
//
//   - em homologacao, dest.xNome vira o literal exigido pela SEFAZ
//     (XmlWriter.pas:854);
//   - abaixo da versao 3.10, indIEDest e forcado a inContribuinte
//     (XmlWriter.pas:870).
func (w *xmlWriter) gerarDest() *pcn.Elem {
	n := w.n
	dest := pcn.NovoElem("dest")

	// Identificacao: estrangeiro quando declarado ou quando o pais nao e o
	// Brasil; CNPJ/CPF no resto.
	if w.versao() >= 3 &&
		(n.Dest.IDEstrangeiro != "" ||
			(n.Dest.EnderDest.CPais != 0 && n.Dest.EnderDest.CPais != codigoPaisPadrao)) {
		dest.Filho(pcn.NodeStr("idEstrangeiro", n.Dest.IDEstrangeiro, true))
	} else {
		dest.Filho(gerarCNPJCPF(n.Dest.CNPJCPF, w.ehNFe(), true))
	}

	if n.Ide.TpAmb == pcn.TaHomologacao {
		n.Dest.XNome = nomeDestHomologacao
	}
	dest.Filho(pcn.NodeStr("xNome", n.Dest.XNome, w.ehNFe()))

	// Na NF-e o endereco e obrigatorio; na NFC-e so sai se houver logradouro.
	if w.ehNFe() || n.Dest.EnderDest.XLgr != "" {
		dest.Filho(w.gerarEnderDest())
	}

	if w.versao() >= 3.10 {
		dest.Filho(pcn.NodeStr("indIEDest", n.Dest.IndIEDest.String(), true))
	} else {
		n.Dest.IndIEDest = pcn.InContribuinte
	}

	// A IE so sai quando o destinatario NAO e isento.
	if n.Dest.IndIEDest != pcn.InIsento {
		if n.Dest.IE != "" || w.versao() < 3 {
			switch {
			case n.Dest.IE == "":
				dest.Filho(pcn.NodeStr("IE", "", true))
			case n.Dest.IE == "ISENTO":
				dest.Filho(pcn.NodeStr("IE", n.Dest.IE, true))
			case strings.TrimSpace(n.Dest.IE) != "" || n.Ide.Modelo != 65:
				dest.Filho(pcn.NodeStr("IE", pcn.OnlyNumber(n.Dest.IE), true))
			}
		}
	}

	dest.Filho(pcn.NodeStr("ISUF", n.Dest.ISUF, false)).
		Filho(pcn.NodeStr("IM", n.Dest.IM, false)).
		Filho(pcn.NodeStr("email", n.Dest.Email, false))

	return dest
}

// gerarEnderDest monta o endereco do destinatario. Porte de
// GerarDestEnderDest.
func (w *xmlWriter) gerarEnderDest() *pcn.Elem {
	end := w.n.Dest.EnderDest
	uf, mun, cMun := ajustarMunicipioUF(end.CPais, end.UF, end.XMun, end.CMun)

	return pcn.NovoElem("enderDest").
		Filho(pcn.NodeStr("xLgr", end.XLgr, true)).
		Filho(pcn.NodeStr("nro", w.ajustarNro(end.Nro), true)).
		Filho(pcn.NodeStr("xCpl", end.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", end.XBairro, true)).
		Filho(pcn.NodeInt("cMun", cMun, 1, true)).
		Filho(pcn.NodeStr("xMun", mun, true)).
		Filho(pcn.NodeStr("UF", uf, true)).
		Filho(pcn.NodeInt("CEP", end.CEP, 8, false)).
		Filho(pcn.NodeInt("cPais", end.CPais, 2, false)).
		Filho(pcn.NodeStr("xPais", end.XPais, false)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(end.Fone), false))
}

// gerarRetirada monta o local de retirada. Porte de GerarRetirada.
// O guard e o LOGRADOURO, nao o CNPJ.
func (w *xmlWriter) gerarRetirada() *pcn.Elem {
	r := w.n.Retirada
	if strings.TrimSpace(r.XLgr) == "" {
		return nil
	}
	uf, mun, cMun := ajustarMunicipioUF(r.CPais, r.UF, r.XMun, r.CMun)

	return pcn.NovoElem("retirada").
		Filho(gerarCNPJCPF(r.CNPJCPF, true, false)).
		Filho(pcn.NodeStr("xNome", r.XNome, false)).
		Filho(pcn.NodeStr("xLgr", r.XLgr, true)).
		Filho(pcn.NodeStr("nro", w.ajustarNro(r.Nro), true)).
		Filho(pcn.NodeStr("xCpl", r.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", r.XBairro, true)).
		Filho(pcn.NodeInt("cMun", cMun, 1, true)).
		Filho(pcn.NodeStr("xMun", mun, true)).
		Filho(pcn.NodeStr("UF", uf, true)).
		Filho(pcn.NodeInt("CEP", r.CEP, 8, false)).
		Filho(pcn.NodeInt("cPais", r.CPais, 1, false)).
		Filho(pcn.NodeStr("xPais", r.XPais, false)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(r.Fone), false)).
		Filho(pcn.NodeStr("email", r.Email, false)).
		Filho(pcn.NodeStr("IE", pcn.OnlyNumber(r.IE), false))
}

// gerarEntrega monta o local de entrega. Porte de GerarEntrega.
func (w *xmlWriter) gerarEntrega() *pcn.Elem {
	e := w.n.Entrega
	if strings.TrimSpace(e.XLgr) == "" {
		return nil
	}
	uf, mun, cMun := ajustarMunicipioUF(e.CPais, e.UF, e.XMun, e.CMun)

	return pcn.NovoElem("entrega").
		Filho(gerarCNPJCPF(e.CNPJCPF, true, false)).
		Filho(pcn.NodeStr("xNome", e.XNome, false)).
		Filho(pcn.NodeStr("xLgr", e.XLgr, true)).
		Filho(pcn.NodeStr("nro", w.ajustarNro(e.Nro), true)).
		Filho(pcn.NodeStr("xCpl", e.XCpl, false)).
		Filho(pcn.NodeStr("xBairro", e.XBairro, true)).
		Filho(pcn.NodeInt("cMun", cMun, 1, true)).
		Filho(pcn.NodeStr("xMun", mun, true)).
		Filho(pcn.NodeStr("UF", uf, true)).
		Filho(pcn.NodeInt("CEP", e.CEP, 8, false)).
		Filho(pcn.NodeInt("cPais", e.CPais, 1, false)).
		Filho(pcn.NodeStr("xPais", e.XPais, false)).
		Filho(pcn.NodeStr("fone", pcn.OnlyNumber(e.Fone), false)).
		Filho(pcn.NodeStr("email", e.Email, false)).
		Filho(pcn.NodeStr("IE", pcn.OnlyNumber(e.IE), false))
}

// gerarAutXML monta um autorizado a baixar o XML. Porte de GerarautXML.
func gerarAutXML(a AutXML) *pcn.Elem {
	return pcn.NovoElem("autXML").Filho(gerarCNPJCPF(a.CNPJCPF, true, true))
}

// gerarTotal monta os totais. Porte de GerarTotal.
func (w *xmlWriter) gerarTotal() *pcn.Elem {
	return pcn.NovoElem("total").
		Filho(w.gerarICMSTot()).
		Filho(w.gerarISSQNTot()).
		Filho(w.gerarRetTrib()).
		Filho(w.rtc.GerarISTot(w.n.Total.ISTot)).
		Filho(w.rtc.GerarIBSCBSTot(w.n.Total.IBSCBSTot)).
		Filho(pcn.NodeDec("vNFTot", w.n.Total.VNFTot, 2, false))
}

// gerarICMSTot monta o total de ICMS. Porte de GerarTotalICMSTotal.
//
// Sai SEMPRE, sem guard -- e o unico subgrupo de total assim.
func (w *xmlWriter) gerarICMSTot() *pcn.Elem {
	v := w.n.Total.ICMSTot
	e := pcn.NovoElem("ICMSTot").
		Filho(pcn.NodeDec("vBC", v.VBC, 2, true)).
		Filho(pcn.NodeDec("vICMS", v.VICMS, 2, true))

	if w.versao() >= 3.10 {
		e.Filho(pcn.NodeDec("vICMSDeson", v.VICMSDeson, 2, true)).
			Filho(pcn.NodeDec("vFCPUFDest", v.VFCPUFDest, 2, false)).
			Filho(pcn.NodeDec("vICMSUFDest", v.VICMSUFDest, 2, false)).
			Filho(pcn.NodeDec("vICMSUFRemet", v.VICMSUFRemet, 2, false))
	}

	if w.versao() >= 4 {
		e.Filho(pcn.NodeDec("vFCP", v.VFCP, 2, true))
	}

	e.Filho(pcn.NodeDec("vBCST", v.VBCST, 2, true)).
		Filho(pcn.NodeDec("vST", v.VST, 2, true))

	if w.versao() >= 4 {
		e.Filho(pcn.NodeDec("vFCPST", v.VFCPST, 2, true)).
			Filho(pcn.NodeDec("vFCPSTRet", v.VFCPSTRet, 2, true)).
			Filho(pcn.NodeDec("qBCMono", v.QBCMono, 2, false)).
			Filho(pcn.NodeDec("vICMSMono", v.VICMSMono, 2, false)).
			Filho(pcn.NodeDec("qBCMonoReten", v.QBCMonoReten, 2, false)).
			Filho(pcn.NodeDec("vICMSMonoReten", v.VICMSMonoReten, 2, false)).
			Filho(pcn.NodeDec("qBCMonoRet", v.QBCMonoRet, 2, false)).
			Filho(pcn.NodeDec("vICMSMonoRet", v.VICMSMonoRet, 2, false))
	}

	e.Filho(pcn.NodeDec("vProd", v.VProd, 2, true)).
		Filho(pcn.NodeDec("vFrete", v.VFrete, 2, true)).
		Filho(pcn.NodeDec("vSeg", v.VSeg, 2, true)).
		Filho(pcn.NodeDec("vDesc", v.VDesc, 2, true)).
		Filho(pcn.NodeDec("vII", v.VII, 2, true)).
		Filho(pcn.NodeDec("vIPI", v.VIPI, 2, true))

	if w.versao() >= 4 {
		e.Filho(pcn.NodeDec("vIPIDevol", v.VIPIDevol, 2, true))
	}

	e.Filho(pcn.NodeDec("vPIS", v.VPIS, 2, true)).
		Filho(pcn.NodeDec("vCOFINS", v.VCOFINS, 2, true)).
		Filho(pcn.NodeDec("vOutro", v.VOutro, 2, true)).
		Filho(pcn.NodeDec("vNF", v.VNF, 2, true)).
		Filho(pcn.NodeDec("vTotTrib", v.VTotTrib, 2, false))

	return e
}

// gerarISSQNTot monta o total de ISSQN. Porte de GerarTotalISSQNtot.
//
// O guard olha CINCO valores; nenhum deles e o dCompet, que e obrigatorio
// dentro do grupo mas nao o dispara.
func (w *xmlWriter) gerarISSQNTot() *pcn.Elem {
	v := w.n.Total.ISSQNTot
	if v.VServ <= 0 && v.VBC <= 0 && v.VISS <= 0 && v.VPIS <= 0 && v.VCOFINS <= 0 {
		return nil
	}

	e := pcn.NovoElem("ISSQNtot").
		Filho(pcn.NodeDec("vServ", v.VServ, 2, false)).
		Filho(pcn.NodeDec("vBC", v.VBC, 2, false)).
		Filho(pcn.NodeDec("vISS", v.VISS, 2, false)).
		Filho(pcn.NodeDec("vPIS", v.VPIS, 2, false)).
		Filho(pcn.NodeDec("vCOFINS", v.VCOFINS, 2, false))

	if w.versao() >= 3.10 {
		e.Filho(pcn.NodeDat("dCompet", v.DCompet, true)).
			Filho(pcn.NodeDec("vDeducao", v.VDeducao, 2, false)).
			Filho(pcn.NodeDec("vOutro", v.VOutro, 2, false)).
			Filho(pcn.NodeDec("vDescIncond", v.VDescIncond, 2, false)).
			Filho(pcn.NodeDec("vDescCond", v.VDescCond, 2, false)).
			Filho(pcn.NodeDec("vISSRet", v.VISSRet, 2, false))

		// RTISSNenhum vale "0" no leiaute, mas o ACBr o usa como "ausente"
		// e omite a tag.
		if v.CRegTrib != pcn.RTISSNenhum {
			e.Filho(pcn.NodeStr("cRegTrib", v.CRegTrib.String(), false))
		}
	}

	return e
}

// gerarRetTrib monta o total de tributos retidos. Porte de
// GerarTotalretTrib. Guard sobre os SETE valores.
func (w *xmlWriter) gerarRetTrib() *pcn.Elem {
	v := w.n.Total.RetTrib
	if v.VRetPIS <= 0 && v.VRetCOFINS <= 0 && v.VRetCSLL <= 0 &&
		v.VBCIRRF <= 0 && v.VIRRF <= 0 && v.VBCRetPrev <= 0 && v.VRetPrev <= 0 {
		return nil
	}

	return pcn.NovoElem("retTrib").
		Filho(pcn.NodeDec("vRetPIS", v.VRetPIS, 2, false)).
		Filho(pcn.NodeDec("vRetCOFINS", v.VRetCOFINS, 2, false)).
		Filho(pcn.NodeDec("vRetCSLL", v.VRetCSLL, 2, false)).
		Filho(pcn.NodeDec("vBCIRRF", v.VBCIRRF, 2, false)).
		Filho(pcn.NodeDec("vIRRF", v.VIRRF, 2, false)).
		Filho(pcn.NodeDec("vBCRetPrev", v.VBCRetPrev, 2, false)).
		Filho(pcn.NodeDec("vRetPrev", v.VRetPrev, 2, false))
}

// gerarTransp monta o transporte. Porte de GerarTransp.
// Sai sempre: modFrete e obrigatorio ate numa nota sem transporte.
func (w *xmlWriter) gerarTransp() *pcn.Elem {
	t := w.n.Transp
	e := pcn.NovoElem("transp").
		Filho(pcn.NodeStr("modFrete", t.ModFrete.String(), true)).
		Filho(w.gerarTransporta()).
		Filho(w.gerarRetTransp()).
		Filho(w.gerarVeicTransp())

	for _, r := range t.Reboque {
		e.Filho(pcn.NovoElem("reboque").
			Filho(pcn.NodeStr("placa", r.Placa, true)).
			Filho(pcn.NodeStr("UF", r.UF, false)).
			Filho(pcn.NodeStr("RNTC", r.RNTC, false)))
	}

	e.Filho(pcn.NodeStr("vagao", t.Vagao, false)).
		Filho(pcn.NodeStr("balsa", t.Balsa, false))

	for _, v := range t.Vol {
		e.Filho(gerarVol(v))
	}

	return e
}

// gerarTransporta monta o transportador. Porte de GerarTranspTransporta.
func (w *xmlWriter) gerarTransporta() *pcn.Elem {
	t := w.n.Transp.Transporta
	if strings.TrimSpace(t.CNPJCPF) == "" && strings.TrimSpace(t.XNome) == "" &&
		strings.TrimSpace(t.IE) == "" && strings.TrimSpace(t.XEnder) == "" &&
		strings.TrimSpace(t.XMun) == "" && strings.TrimSpace(t.UF) == "" {
		return nil
	}

	e := pcn.NovoElem("transporta")
	if strings.TrimSpace(t.CNPJCPF) != "" {
		e.Filho(gerarCNPJCPF(t.CNPJCPF, true, true))
	}
	e.Filho(pcn.NodeStr("xNome", t.XNome, false))

	if strings.TrimSpace(t.IE) == "ISENTO" {
		e.Filho(pcn.NodeStr("IE", t.IE, false))
	} else {
		e.Filho(pcn.NodeStr("IE", pcn.OnlyNumber(t.IE), false))
	}

	e.Filho(pcn.NodeStr("xEnder", t.XEnder, false)).
		Filho(pcn.NodeStr("xMun", t.XMun, false))

	if strings.TrimSpace(t.UF) != "" {
		e.Filho(pcn.NodeStr("UF", t.UF, false))
	}

	return e
}

// gerarRetTransp monta a retencao do ICMS do transporte. Porte de
// GerarTranspRetTransp.
func (w *xmlWriter) gerarRetTransp() *pcn.Elem {
	r := w.n.Transp.RetTransp
	if r.VServ <= 0 && r.VBCRet <= 0 && r.PICMSRet <= 0 && r.VICMSRet <= 0 &&
		strings.TrimSpace(r.CFOP) == "" && r.CMunFG <= 0 {
		return nil
	}

	return pcn.NovoElem("retTransp").
		Filho(pcn.NodeDec("vServ", r.VServ, 2, true)).
		Filho(pcn.NodeDec("vBCRet", r.VBCRet, 2, true)).
		Filho(pcn.NodeDec("pICMSRet", r.PICMSRet, 2, true)).
		Filho(pcn.NodeDec("vICMSRet", r.VICMSRet, 2, true)).
		Filho(pcn.NodeStr("CFOP", pcn.OnlyNumber(r.CFOP), true)).
		// O ACBr grava este cMunFG com tcStr, nao tcInt como os demais:
		// sem zero a esquerda. Replicado.
		Filho(pcn.NodeStr("cMunFG", strconv.Itoa(r.CMunFG), true))
}

// gerarVeicTransp monta o veiculo do transporte. Porte de
// GerarTranspVeicTransp.
func (w *xmlWriter) gerarVeicTransp() *pcn.Elem {
	v := w.n.Transp.VeicTransp
	if strings.TrimSpace(v.Placa) == "" && strings.TrimSpace(v.UF) == "" &&
		strings.TrimSpace(v.RNTC) == "" {
		return nil
	}

	return pcn.NovoElem("veicTransp").
		Filho(pcn.NodeStr("placa", v.Placa, true)).
		Filho(pcn.NodeStr("UF", v.UF, false)).
		Filho(pcn.NodeStr("RNTC", v.RNTC, false))
}

// gerarVol monta um volume. Porte de GerarTranspVol.
//
// A tag do lacre e "lacres" POR OCORRENCIA, nao um grupo com filhos "lacre".
func gerarVol(v Vol) *pcn.Elem {
	e := pcn.NovoElem("vol").
		Filho(pcn.NodeInt("qVol", v.QVol, 1, true)).
		Filho(pcn.NodeStr("esp", v.Esp, false)).
		Filho(pcn.NodeStr("marca", v.Marca, false)).
		Filho(pcn.NodeStr("nVol", v.NVol, false)).
		Filho(pcn.NodeDec("pesoL", v.PesoL, 3, false)).
		Filho(pcn.NodeDec("pesoB", v.PesoB, 3, false))

	for _, l := range v.Lacres {
		e.Filho(pcn.NovoElem("lacres").Filho(pcn.NodeStr("nLacre", l.NLacre, true)))
	}

	return e
}

// gerarCobr monta a cobranca. Porte de GerarCobr.
func (w *xmlWriter) gerarCobr() *pcn.Elem {
	c := w.n.Cobr
	if strings.TrimSpace(c.Fat.NFat) == "" && c.Fat.VOrig <= 0 &&
		c.Fat.VDesc <= 0 && c.Fat.VLiq <= 0 && len(c.Dup) == 0 {
		return nil
	}

	e := pcn.NovoElem("cobr").Filho(w.gerarFat())

	for _, d := range c.Dup {
		dup := pcn.NovoElem("dup")
		// nDup e dVenc so sao obrigatorios a partir da 4.00.
		obrig := w.versao() >= 4
		dup.Filho(pcn.NodeStr("nDup", d.NDup, obrig)).
			Filho(pcn.NodeDat("dVenc", d.DVenc, obrig)).
			Filho(pcn.NodeDec("vDup", d.VDup, 2, true))
		e.Filho(dup)
	}

	return e
}

// gerarFat monta a fatura. Porte de GerarCobrFat.
//
// A obrigatoriedade dos quatro campos depende da OPCAO CamposFatObrigatorios
// combinada com a versao -- e a unica opcao do componente que muda a forma de
// uma tag.
func (w *xmlWriter) gerarFat() *pcn.Elem {
	f := w.n.Cobr.Fat
	if strings.TrimSpace(f.NFat) == "" && f.VOrig <= 0 && f.VDesc <= 0 && f.VLiq <= 0 {
		return nil
	}

	obrig := !w.opcoes.FatCamposOpcionais && w.versao() >= 4

	return pcn.NovoElem("fat").
		Filho(pcn.NodeStr("nFat", f.NFat, obrig)).
		Filho(pcn.NodeDec("vOrig", f.VOrig, 2, obrig)).
		Filho(pcn.NodeDec("vDesc", f.VDesc, 2, obrig)).
		Filho(pcn.NodeDec("vLiq", f.VLiq, 2, obrig))
}

// gerarPag monta os pagamentos. Porte de Gerarpag.
//
// A FORMA do grupo muda com a versao: a partir da 4.00 e um unico <pag> com
// varios <detPag> e um <vTroco> no fim; abaixo dela sao varios <pag> soltos,
// sem vTroco e sem indPag.
func (w *xmlWriter) gerarPag() []*pcn.Elem {
	n := w.n

	if w.versao() >= 4 {
		pag := pcn.NovoElem("pag")
		for _, p := range n.Pag.DetPag {
			det := pcn.NovoElem("detPag").
				Filho(pcn.NodeStr("indPag", p.IndPag.String(), false))
			w.preencherPagamento(det, p)
			pag.Filho(det)
		}
		pag.Filho(pcn.NodeDec("vTroco", n.Pag.VTroco, 2, false))
		return []*pcn.Elem{pag}
	}

	var pags []*pcn.Elem
	for _, p := range n.Pag.DetPag {
		e := pcn.NovoElem("pag")
		w.preencherPagamento(e, p)
		pags = append(pags, e)
	}
	return pags
}

// preencherPagamento acrescenta os campos comuns de uma forma de pagamento.
func (w *xmlWriter) preencherPagamento(e *pcn.Elem, p PagItem) {
	e.Filho(pcn.NodeStr("tPag", p.TPag.String(), true)).
		Filho(pcn.NodeStr("xPag", p.XPag, false)).
		Filho(pcn.NodeDec("vPag", p.VPag, 2, true)).
		Filho(pcn.NodeDat("dPag", p.DPag, false))

	// O par CNPJPag/UFPag sai junto, e basta UM dos dois estar preenchido.
	if p.CNPJPag != "" || p.UFPag != "" {
		e.Filho(pcn.NodeStr("CNPJPag", p.CNPJPag, true)).
			Filho(pcn.NodeStr("UFPag", p.UFPag, true))
	}

	// O grupo card e disparado pelo tpIntegra, nao pela bandeira nem pelo
	// CNPJ da credenciadora.
	if p.TpIntegra != pcn.TiNaoInformado {
		e.Filho(pcn.NovoElem("card").
			Filho(pcn.NodeStr("tpIntegra", p.TpIntegra.String(), true)).
			Filho(pcn.NodeStr("CNPJ", p.CNPJ, false)).
			Filho(pcn.NodeStr("tBand", p.TBand.String(), false)).
			Filho(pcn.NodeStr("cAut", p.CAut, false)).
			Filho(pcn.NodeStr("CNPJReceb", p.CNPJReceb, false)).
			Filho(pcn.NodeStr("idTermPag", p.IDTermPag, false)))
	}
}

// gerarInfIntermed monta o intermediador. Porte de GerarInfIntermed.
func (w *xmlWriter) gerarInfIntermed() *pcn.Elem {
	if w.versao() < 4 {
		return nil
	}
	i := w.n.InfIntermed
	if strings.TrimSpace(i.CNPJ)+strings.TrimSpace(i.IDCadIntTran) == "" {
		return nil
	}

	return pcn.NovoElem("infIntermed").
		Filho(pcn.NodeStr("CNPJ", i.CNPJ, true)).
		Filho(pcn.NodeStr("idCadIntTran", i.IDCadIntTran, true))
}

// gerarInfAdic monta as informacoes adicionais. Porte de GerarInfAdic.
func (w *xmlWriter) gerarInfAdic() *pcn.Elem {
	a := w.n.InfAdic
	if strings.TrimSpace(a.InfAdFisco) == "" && strings.TrimSpace(a.InfCpl) == "" &&
		len(a.ObsCont) == 0 && len(a.ObsFisco) == 0 && len(a.ProcRef) == 0 {
		return nil
	}

	e := pcn.NovoElem("infAdic").
		Filho(pcn.NodeStr("infAdFisco", a.InfAdFisco, false)).
		Filho(pcn.NodeStr("infCpl", a.InfCpl, false))

	for _, o := range a.ObsCont {
		e.Filho(pcn.NovoElem("obsCont").
			Attr("xCampo", o.XCampo).
			Filho(pcn.NodeStr("xTexto", o.XTexto, true)))
	}
	for _, o := range a.ObsFisco {
		e.Filho(pcn.NovoElem("obsFisco").
			Attr("xCampo", o.XCampo).
			Filho(pcn.NodeStr("xTexto", o.XTexto, true)))
	}

	for _, p := range a.ProcRef {
		ref := pcn.NovoElem("procRef").
			Filho(pcn.NodeStr("nProc", p.NProc, true)).
			Filho(pcn.NodeStr("indProc", p.IndProc.String(), true))
		// tpAto so sai quando informado: TaNenhum grava vazio.
		if p.TpAto != TaNenhum {
			ref.Filho(pcn.NodeStr("tpAto", p.TpAto.String(), false))
		}
		e.Filho(ref)
	}

	return e
}

// gerarExporta monta o grupo de exportacao. Porte de GerarExporta.
//
// O grupo tem DUAS FORMAS COMPLETAMENTE DIFERENTES, decididas pela versao:
// da 3.10 em diante sao UFSaidaPais/xLocExporta/xLocDespacho; abaixo dela,
// UFEmbarq/xLocEmbarq. Nao ha campo em comum, e ate o guard muda de campo.
func (w *xmlWriter) gerarExporta() *pcn.Elem {
	ex := w.n.Exporta

	if w.versao() >= 3.10 {
		if strings.TrimSpace(ex.UFSaidaPais)+strings.TrimSpace(ex.XLocExporta) == "" {
			return nil
		}
		return pcn.NovoElem("exporta").
			Filho(pcn.NodeStr("UFSaidaPais", ex.UFSaidaPais, true)).
			Filho(pcn.NodeStr("xLocExporta", ex.XLocExporta, true)).
			Filho(pcn.NodeStr("xLocDespacho", ex.XLocDespacho, false))
	}

	if strings.TrimSpace(ex.UFEmbarq)+strings.TrimSpace(ex.XLocEmbarq) == "" {
		return nil
	}
	return pcn.NovoElem("exporta").
		Filho(pcn.NodeStr("UFEmbarq", ex.UFEmbarq, true)).
		Filho(pcn.NodeStr("xLocEmbarq", ex.XLocEmbarq, true))
}

// gerarCompra monta o grupo de compras publicas. Porte de GerarCompra.
func (w *xmlWriter) gerarCompra() *pcn.Elem {
	c := w.n.Compra
	if strings.TrimSpace(c.XNEmp)+strings.TrimSpace(c.XPed)+strings.TrimSpace(c.XCont) == "" {
		return nil
	}

	return pcn.NovoElem("compra").
		Filho(pcn.NodeStr("xNEmp", c.XNEmp, false)).
		Filho(pcn.NodeStr("xPed", c.XPed, false)).
		Filho(pcn.NodeStr("xCont", c.XCont, false))
}

// gerarCana monta o grupo de aquisicao de cana. Porte de GerarCana.
func (w *xmlWriter) gerarCana() *pcn.Elem {
	c := w.n.Cana
	if strings.TrimSpace(c.Safra) == "" && strings.TrimSpace(c.Ref) == "" &&
		len(c.ForDia) == 0 && len(c.Deduc) == 0 {
		return nil
	}

	e := pcn.NovoElem("cana").
		Filho(pcn.NodeStr("safra", c.Safra, false)).
		Filho(pcn.NodeStr("ref", c.Ref, false))

	for _, f := range c.ForDia {
		e.Filho(pcn.NovoElem("forDia").
			Attr("dia", strconv.Itoa(f.Dia)).
			Filho(pcn.NodeDec("qtde", f.Qtde, 10, true)))
	}

	e.Filho(pcn.NodeDec("qTotMes", c.QTotMes, 10, true)).
		Filho(pcn.NodeDec("qTotAnt", c.QTotAnt, 10, true)).
		Filho(pcn.NodeDec("qTotGer", c.QTotGer, 10, true))

	for _, d := range c.Deduc {
		e.Filho(pcn.NovoElem("deduc").
			Filho(pcn.NodeStr("xDed", d.XDed, true)).
			Filho(pcn.NodeDec("vDed", d.VDed, 2, true)))
	}

	e.Filho(pcn.NodeDec("vFor", c.VFor, 2, true)).
		Filho(pcn.NodeDec("vTotDed", c.VTotDed, 2, true)).
		Filho(pcn.NodeDec("vLiqFor", c.VLiqFor, 2, true))

	return e
}

// gerarInfRespTec monta o responsavel tecnico. Porte de GerarinfRespTec.
//
// O hashCSRT so sai quando idCSRT e CSRT estao nas OPCOES -- o hash e
// calculado na hora, nao lido do modelo.
func (w *xmlWriter) gerarInfRespTec() *pcn.Elem {
	r := w.n.InfRespTec
	if r.CNPJ == "" {
		return nil
	}

	e := pcn.NovoElem("infRespTec").
		Filho(pcn.NodeStr("CNPJ", pcn.OnlyCPFCNPJAlphaNum(r.CNPJ), true)).
		Filho(pcn.NodeStr("xContato", r.XContato, true)).
		Filho(pcn.NodeStr("email", r.Email, true)).
		Filho(pcn.NodeStr("fone", r.Fone, true))

	if w.opcoes.IDCSRT != 0 && w.opcoes.CSRT != "" {
		e.Filho(pcn.NodeInt("idCSRT", w.opcoes.IDCSRT, 2, true)).
			Filho(pcn.NodeStr("hashCSRT", hashCSRT(w.opcoes.CSRT, w.chave), true))
	}

	return e
}

// gerarAgropecuario monta o grupo agropecuario. Porte de GerarAgropecuario.
//
// Os dois subgrupos sao EXCLUDENTES no ACBr: havendo defensivo, a guia de
// transito nao sai, nem que esteja preenchida. E um else-if, nao dois ifs.
func (w *xmlWriter) gerarAgropecuario() *pcn.Elem {
	a := w.n.Agropecuario

	if len(a.Defensivo) > 0 {
		e := pcn.NovoElem("agropecuario")
		for _, d := range a.Defensivo {
			e.Filho(pcn.NovoElem("defensivo").
				Filho(pcn.NodeStr("nReceituario", d.NReceituario, true)).
				Filho(pcn.NodeStr("CPFRespTec", d.CPFRespTec, true)))
		}
		return e
	}

	if a.GuiaTransito.TpGuia != TpgNenhum {
		return pcn.NovoElem("agropecuario").
			// ORDEM: tpGuia vem ANTES de UFGuia. O XSD e xs:sequence
			// (leiauteNFe_v4.00.xsd:6539) e o .pas segue essa ordem
			// (XmlWriter.pas:1088-1092), apesar de os IDs do leiaute
			// sugerirem o contrario (ZF05 = UFGuia, ZF06 = tpGuia).
			Filho(pcn.NovoElem("guiaTransito").
				Filho(pcn.NodeStr("tpGuia", a.GuiaTransito.TpGuia.String(), true)).
				Filho(pcn.NodeStr("UFGuia", a.GuiaTransito.UFGuia, true)).
				Filho(pcn.NodeStr("serieGuia", a.GuiaTransito.SerieGuia, false)).
				Filho(pcn.NodeStr("nGuia", a.GuiaTransito.NGuia, true)))
	}

	return nil
}

// gerarProtNFe monta o protocolo de autorizacao. Porte de GerarProtNFe.
func (w *xmlWriter) gerarProtNFe() *pcn.Elem {
	p := w.n.ProcNFe

	dhRecbto := pcn.FormatarDataHoraXML(p.DhRecbto, pcn.SiglaUF(w.n.Ide.CUF))
	if w.versao() < 3.10 {
		dhRecbto = formatarDataHoraSemFuso(p.DhRecbto)
	}

	return pcn.NovoElem("protNFe").
		Attr("versao", pcn.FormatarVersaoXML(w.versao())).
		// O .pas NAO usa AddNode aqui: escreve Content direto, sem
		// FiltrarTextoXML. O xMotivo tem de voltar igual ao que a SEFAZ
		// devolveu, acento incluso.
		Filho(pcn.NovoElem("infProt").
			Filho(pcn.NodeStrSemFiltro("tpAmb", p.TpAmb.String(), true)).
			Filho(pcn.NodeStrSemFiltro("verAplic", p.VerAplic, true)).
			Filho(pcn.NodeStrSemFiltro("chNFe", p.ChDFe, true)).
			Filho(pcn.NodeStrSemFiltro("dhRecbto", dhRecbto, true)).
			Filho(pcn.NodeStrSemFiltro("nProt", p.NProt, true)).
			Filho(pcn.NodeStrSemFiltro("digVal", p.DigVal, true)).
			Filho(pcn.NodeStrSemFiltro("cStat", strconv.Itoa(p.CStat), true)).
			Filho(pcn.NodeStrSemFiltro("xMotivo", p.XMotivo, true)))
}

// ajustarNro aplica o ExecutarAjusteTagNro do ACBr: com a opcao ligada, um
// numero vazio vira "S/N".
func (w *xmlWriter) ajustarNro(nro string) string {
	if w.opcoes.AjustarTagNro && strings.TrimSpace(nro) == "" {
		return "S/N"
	}
	return nro
}

// ajustarMunicipioUF troca municipio e UF pelos literais de exterior quando o
// pais nao e o Brasil. Porte de AjustarMunicipioUF (XmlWriter.pas:247).
//
// cPais zero e tratado como Brasil, como no original.
func ajustarMunicipioUF(cPais int, uf, xMun string, cMun int) (string, string, int) {
	if cPais == 0 {
		cPais = codigoPaisPadrao
	}
	if cPais == codigoPaisPadrao {
		return uf, xMun, cMun
	}
	return UFExterior, XMunExterior, CMunExterior
}

// ---------------------------------------------------------------------------
// Auxiliares
// ---------------------------------------------------------------------------

// gerarCNPJCPF e o AddNodeCNPJCPF do TACBrXmlWriter: ate 11 posicoes gera
// CPF (completado a 11), senao CNPJ (completado a 14). Vazio da preferencia
// a CNPJ.
// O .pas VARIA os dois ultimos argumentos por chamada: dest passa
// obrigatorio := IsNFe (:849 e :852) e retirada/entrega passam
// (True, False) (:955, :1005).
func gerarCNPJCPF(valor string, obrigatorio, preencheZeros bool) *pcn.Elem {
	valor = pcn.OnlyCPFCNPJAlphaNum(strings.TrimSpace(valor))
	tam := len(valor)

	if tam > 0 && tam <= 11 {
		if preencheZeros {
			valor = pcn.PreencherZerosEsquerda(valor, 11)
		}
		return pcn.NodeStrSemFiltro("CPF", valor, obrigatorio)
	}
	if preencheZeros && tam > 0 && tam != 14 {
		valor = pcn.PreencherZerosEsquerda(valor, 14)
	}
	return pcn.NodeStrSemFiltro("CNPJ", valor, obrigatorio)
}

// gerarSignature monta o bloco XMLDSig a partir de uma assinatura ja
// calculada. Assinar de verdade e papel de dfe.AssinarXML.
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

// formatarHora formata a hora de saida no padrao hh:mm:ss do leiaute.
// So e usada abaixo da versao 3, unica faixa em que hSaiEnt existe.
func formatarHora(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("15:04:05")
}

// formatarDataHoraSemFuso formata data e hora com 19 posicoes, sem o fuso.
// E o DateTimeTodh do ACBr, usado no dhCont e no dhRecbto abaixo da 3.10.
func formatarDataHoraSemFuso(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05")
}

// hashCSRT calcula o hash do Codigo de Seguranca do Responsavel Tecnico
// sobre a chave de acesso. Delega ao packages/dfe, que ja o implementa.
func hashCSRT(csrt, chave string) string {
	return dfe.HashCSRT(csrt, chave)
}

// gerarDet monta um item da nota.
//
// STUB da etapa 4a: devolve nil, e o Filho(nil) do construtor de elementos
// absorve sem gerar nada. O gerador do det -- produto, dez grupos de imposto
// e os grupos por tipo de mercadoria -- entra na etapa 4b, que preenche este
// metodo sem mexer na estrutura acima.
func (w *xmlWriter) gerarDet(_ int) *pcn.Elem {
	return nil
}

// nodeNumStr reproduz o tcNumStr do AddNode (ACBrXmlWriter.pas:282-289):
// texto puro, completado com zeros a esquerda ate o tamanho minimo.
//
// Diferente do tcInt, NAO converte para numero -- o conteudo pode ter letra.
// Vazio continua sumindo quando o campo e opcional.
func nodeNumStr(nome, valor string, minimo int, obrigatorio bool) *pcn.Elem {
	valor = strings.TrimSpace(valor)
	if valor != "" && len(valor) < minimo {
		valor = pcn.PreencherZerosEsquerda(valor, minimo)
	}
	return pcn.NodeStr(nome, valor, obrigatorio)
}

// nodeIntDeTexto grava com a semantica do tcInt um campo que a classe Delphi
// declara como string: converte para numero e completa com zeros.
//
// O efeito que importa e a NORMALIZACAO: "0012" num campo de 3 posicoes sai
// "012", nao "0012". Usado em refNFP/mod, refECF/nECF e refECF/nCOO, os tres
// campos string que o ACBr grava com tcInt.
func nodeIntDeTexto(nome, valor string, minimo int, obrigatorio bool) *pcn.Elem {
	n, err := strconv.Atoi(pcn.OnlyNumber(strings.TrimSpace(valor)))
	if err != nil {
		// Conteudo nao numerico: sem o que converter, vale o original.
		return pcn.NodeStr(nome, valor, obrigatorio)
	}
	return pcn.NodeInt(nome, n, minimo, obrigatorio)
}
