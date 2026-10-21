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
	"io"
	"strconv"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitor do XML da NF-e. Porte de TNFeXmlReader
// (Base/ACBrNFe.XmlReader.pas), metodo a metodo e na mesma ordem.
//
// Regras do original preservadas de proposito:
//
//   - node nil sai sem fazer nada ("if not Assigned(ANode) then Exit");
//   - a escolha entre Find (nome qualificado) e FindAnyNs (local name) e a
//     do .pas, campo a campo. O leitor da NFe usa Find em quase tudo; as
//     poucas excecoes estao marcadas com FINDANYNS;
//   - a precisao de cada campo (De2, De3, De4, De10) e a declarada no .pas;
//   - as condicionais POR VERSAO DE LEIAUTE (infNFe.Versao >= 3 e >= 4)
//     governam quais tags sao lidas, e sao a parte mais facil de perder;
//   - grupos achatados (ICMS, PIS, COFINS) sao lidos em SEQUENCIA, nao pelo
//     primeiro achado -- ver lerDetImposto.
//
// Divergencias deliberadas estao marcadas com DIVERGENCIA.
//
// # Sobre DecodeHTMLEntities
//
// O ACBr aplica DecodeHTMLEntities em xNome de dest, retirada, entrega e
// transporta (e em nenhum outro campo). Aqui isso NAO e replicado porque seria
// uma SEGUNDA decodificacao: o encoding/xml do Go ja resolve as referencias de
// entidade ao produzir o CharData. Para XML valido o resultado e identico; a
// unica diferenca aparece em entidade dupla ("&amp;amp;"), que o ACBr reduz a
// "&" e aqui permanece "&amp;" -- que e o conteudo correto.

// LerXML le uma NF-e a partir de um io.Reader.
//
// Aceita as tres formas de raiz: <NFe> avulsa e os envelopes <nfeProc> e
// <procNFe>, que trazem a nota mais o protocolo de autorizacao.
func LerXML(r io.Reader) (*NFe, error) {
	doc, err := pcn.Parse(r)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// LerXMLString le uma NF-e a partir de uma string.
func LerXMLString(xml string) (*NFe, error) {
	return LerBytes([]byte(xml))
}

// LerBytes le uma NF-e a partir de bytes. BOM UTF-8 e declaracao
// <?xml ...?> sao tolerados.
func LerBytes(dados []byte) (*NFe, error) {
	doc, err := pcn.ParseBytes(dados)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// leitor carrega a nota em construcao, como o FNFe do TNFeXmlReader. Existe
// para que as condicionais por versao possam consultar infNFe.Versao sem
// passar o valor por parametro em cada metodo, exatamente como no Delphi.
type leitor struct {
	n *NFe
}

// versao devolve a versao do leiaute lida do atributo de infNFe. E o que
// governa as condicionais do leitor.
func (l *leitor) versao() float64 { return l.n.InfNFe.Versao }

// lerDocumento e o porte de TNFeXmlReader.LerXml.
//
// DIVERGENCIA: o original levanta Exception em cada validacao; aqui cada uma
// vira um erro sentinela, que o chamador testa com errors.Is.
func lerDocumento(doc *pcn.Document) (*NFe, error) {
	if doc == nil || doc.Root == nil {
		return nil, ErrXMLVazio
	}

	n := NovaNFe()
	l := &leitor{n: &n}

	var nfeNode *pcn.Node
	if doc.Root.Nome == "nfeProc" || doc.Root.Nome == "procNFe" {
		l.lerProtNFe(doc.Root.Find("protNFe"))
		nfeNode = doc.Root.Find("NFe")
	} else {
		nfeNode = doc.Root
	}

	if nfeNode == nil {
		// DIVERGENCIA: o .pas:135 envolve todo o corpo num
		// `if NFeNode <> nil`, sem else, e devolve True mesmo assim -- o
		// chamador fica com uma TNFe vazia e sem saber. Aqui o envelope sem
		// nota e erro. Custo: o protocolo ja lido se perde junto.
		return nil, ErrNFeNaoEncontrada
	}

	infNFeNode := nfeNode.Find("infNFe")
	if infNFeNode == nil {
		return nil, ErrXMLIncorreto
	}

	id, ok := infNFeNode.AttrOk("Id")
	if !ok {
		return nil, ErrAtributoIDAusente
	}
	n.InfNFe.ID = id

	versao, ok := infNFeNode.AttrOk("versao")
	if !ok {
		return nil, ErrAtributoVersaoAusente
	}
	n.InfNFe.Versao = pcn.StringToFloatDef(versao, 0)

	l.lerInfNFe(infNFeNode)

	// infNFeSupl e infPAA sao irmaos de infNFe, nao filhos -- pendem do no NFe.
	l.lerInfNFeSupl(nfeNode.Find("infNFeSupl"))
	l.lerInfPAA(nfeNode.Find("infPAA"))

	pcn.LerSignature(nfeNode.Find("Signature"), &n.Signature)

	return &n, nil
}

// lerProtNFe le o protocolo de autorizacao. Porte de LerProtNFe.
func (l *leitor) lerProtNFe(node *pcn.Node) {
	if node == nil {
		return
	}

	// FINDANYNS: infProt e o unico node do leitor procurado por local name.
	aux := node.FindAnyNs("infProt")
	if aux == nil {
		return
	}

	p := &l.n.ProcNFe
	p.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(aux.Find("tpAmb")))
	p.VerAplic = pcn.ConteudoStr(aux.Find("verAplic"))
	p.ChDFe = pcn.ConteudoStr(aux.Find("chNFe"))
	p.DhRecbto = pcn.ConteudoDataDef(aux.Find("dhRecbto"))
	p.NProt = pcn.ConteudoStr(aux.Find("nProt"))
	p.DigVal = pcn.ConteudoStr(aux.Find("digVal"))
	p.CStat = pcn.ConteudoInt(aux.Find("cStat"))
	p.XMotivo = pcn.ConteudoStr(aux.Find("xMotivo"))
	p.CMsg = pcn.ConteudoInt(aux.Find("cMsg"))
	p.XMsg = pcn.ConteudoStr(aux.Find("xMsg"))
}

// lerInfNFe le o corpo da nota. Porte de LerInfNFe.
func (l *leitor) lerInfNFe(node *pcn.Node) {
	if node == nil {
		return
	}

	l.lerIde(node.Find("ide"))
	l.lerEmit(node.Find("emit"))
	l.lerAvulsa(node.Find("avulsa"))
	l.lerDest(node.Find("dest"))
	l.lerRetirada(node.Find("retirada"))
	l.lerEntrega(node.Find("entrega"))

	for _, aut := range node.FindAll("autXML") {
		l.n.AutXML = append(l.n.AutXML, AutXML{CNPJCPF: pcn.ConteudoCNPJCPF(aut)})
	}

	for _, det := range node.FindAll("det") {
		l.lerDet(det)
	}

	l.lerTotal(node.Find("total"))
	l.lerTransp(node.Find("transp"))

	// lerCobr recebe o no infNFe, nao um no cobr: no ACBr ela procura cobr e
	// pag por conta propria, porque os dois sao irmaos.
	l.lerCobr(node)

	l.lerInfIntermed(node.Find("infIntermed"))
	l.lerInfAdic(node.Find("infAdic"))
	l.lerExporta(node.Find("exporta"))
	l.lerCompra(node.Find("compra"))
	l.lerCana(node.Find("cana"))
	l.lerInfRespTec(node.Find("infRespTec"))
	l.lerAgropecuario(node.Find("agropecuario"))
}

// lerIde le a identificacao da nota. Porte de LerIde.
//
// Concentra quatro condicionais por versao de leiaute; perder qualquer uma
// delas faz a nota ser lida com campo vazio sem nenhum erro.
func (l *leitor) lerIde(node *pcn.Node) {
	if node == nil {
		return
	}
	ide := &l.n.Ide

	ide.CUF = pcn.ConteudoInt(node.Find("cUF"))
	ide.CNF = pcn.ConteudoInt(node.Find("cNF"))
	// Sentinela do ACBr: cNF ausente ou zero vira -2, para o gerador saber
	// que precisa sortear um codigo novo em vez de gravar zero.
	if ide.CNF == 0 {
		ide.CNF = -2
	}
	ide.NatOp = pcn.ConteudoStr(node.Find("natOp"))
	ide.IndPag, _ = ParseIndicadorPagamento(pcn.ConteudoStr(node.Find("indPag")))
	ide.Modelo = pcn.ConteudoInt(node.Find("mod"))
	ide.Serie = pcn.ConteudoInt(node.Find("serie"))
	ide.NNF = pcn.ConteudoInt(node.Find("nNF"))

	if l.versao() >= 3 {
		ide.DEmi = pcn.ConteudoDataDef(node.Find("dhEmi"))
		ide.DSaiEnt = pcn.ConteudoDataDef(node.Find("dhSaiEnt"))
	} else {
		ide.DEmi = pcn.ConteudoDataDef(node.Find("dEmi"))
		ide.DSaiEnt = pcn.ConteudoDataDef(node.Find("dSaiEnt"))
		// hSaiEnt so existe ate a 3.10: na 4.00 a hora entrou no dhSaiEnt.
		ide.HSaiEnt = horaComoData(pcn.ConteudoHora(node.Find("hSaiEnt")))
	}

	ide.DPrevEntrega = pcn.ConteudoDataDef(node.Find("dPrevEntrega"))
	ide.TpNF, _ = pcn.ParseTipoNFe(pcn.ConteudoStr(node.Find("tpNF")))

	if l.versao() >= 3 {
		ide.IdDest, _ = ParseDestinoOperacao(pcn.ConteudoStr(node.Find("idDest")))
	}

	ide.CMunFG = pcn.ConteudoInt(node.Find("cMunFG"))
	ide.CMunFGIBS = pcn.ConteudoInt(node.Find("cMunFGIBS"))

	ide.TpImp, _ = pcn.ParseTipoImpressao(pcn.ConteudoStr(node.Find("tpImp")))
	ide.TpEmis, _ = pcn.ParseTipoEmissao(pcn.ConteudoStr(node.Find("tpEmis")))
	ide.CDV = pcn.ConteudoInt(node.Find("cDV"))
	ide.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.Find("tpAmb")))
	ide.FinNFe, _ = ParseFinalidadeNFe(pcn.ConteudoStr(node.Find("finNFe")))

	if l.versao() >= 4 {
		ide.TpNFDebito, _ = ParseTpNFDebito(pcn.ConteudoStr(node.Find("tpNFDebito")))
		ide.TpNFCredito, _ = ParseTpNFCredito(pcn.ConteudoStr(node.Find("tpNFCredito")))
	}

	if l.versao() >= 3 {
		ide.IndFinal, _ = ParseConsumidorFinal(pcn.ConteudoStr(node.Find("indFinal")))
		ide.IndPres, _ = ParsePresencaComprador(pcn.ConteudoStr(node.Find("indPres")))
	}

	if l.versao() >= 4 {
		ide.IndIntermed, _ = ParseIndIntermed(pcn.ConteudoStr(node.Find("indIntermed")))
	}

	ide.ProcEmi, _ = pcn.ParseProcessoEmissao(pcn.ConteudoStr(node.Find("procEmi")))
	ide.VerProc = pcn.ConteudoStr(node.Find("verProc"))
	ide.DhCont = pcn.ConteudoDataDef(node.Find("dhCont"))
	ide.XJust = pcn.ConteudoStr(node.Find("xJust"))
	ide.CIndOp = pcn.ConteudoStr(node.Find("cIndOp"))

	for _, ref := range node.FindAll("NFref") {
		l.lerIdeNFref(ref)
	}

	rtc.LerGCompraGov(node.Find("gCompraGov"), &ide.GCompraGov)
	rtc.LerGPagAntecipado(node.Find("gPagAntecipado"), &ide.GPagAntecipado)
}

// lerIdeNFref le uma nota referenciada. Porte de LerIdeNFref.
func (l *leitor) lerIdeNFref(node *pcn.Node) {
	if node == nil {
		return
	}

	var ref NFref
	ref.RefNFe = pcn.ConteudoStr(node.Find("refNFe"))
	ref.RefNFeSig = pcn.ConteudoStr(node.Find("refNFeSig"))

	if refNode := node.Find("refNF"); refNode != nil {
		ref.RefNF.CUF = pcn.ConteudoInt(refNode.Find("cUF"))
		ref.RefNF.AAMM = pcn.ConteudoStr(refNode.Find("AAMM"))
		ref.RefNF.CNPJ = pcn.ConteudoStr(refNode.Find("CNPJ"))
		// O ACBr escreve StrToIntDef(ObterConteudo(..., tcInt), 55), mas o
		// tcInt ja devolve inteiro: o default 55 e inalcancavel, e tag
		// ausente resulta em 0. Replicado.
		ref.RefNF.Modelo = pcn.ConteudoInt(refNode.Find("mod"))
		ref.RefNF.Serie = pcn.ConteudoInt(refNode.Find("serie"))
		ref.RefNF.NNF = pcn.ConteudoInt(refNode.Find("nNF"))
	}

	if refNode := node.Find("refNFP"); refNode != nil {
		ref.RefNFP.CUF = pcn.ConteudoInt(refNode.Find("cUF"))
		ref.RefNFP.AAMM = pcn.ConteudoStr(refNode.Find("AAMM"))
		ref.RefNFP.CNPJCPF = pcn.ConteudoCNPJCPF(refNode)
		ref.RefNFP.IE = pcn.ConteudoStr(refNode.Find("IE"))
		// DIVERGENCIA (perda de dado no ACBr, corrigida): o .pas:334 le
		// este campo com tcInt e o atribui a uma property STRING
		// (Classes.pas:116), entao o Variant inteiro vira "4" e o zero a
		// esquerda do leiaute se perde. Em refNFP o modelo admite "01" e
		// "04", e o XSD o declara como string. Aqui vale a leitura textual.
		//
		// Contraste com RefNF.Modelo acima, onde a pegadinha equivalente e
		// REPLICADA por nao perder dado.
		ref.RefNFP.Modelo = pcn.ConteudoStr(refNode.Find("mod"))
		ref.RefNFP.Serie = pcn.ConteudoInt(refNode.Find("serie"))
		ref.RefNFP.NNF = pcn.ConteudoInt(refNode.Find("nNF"))
	}

	ref.RefCTe = pcn.ConteudoStr(node.Find("refCTe"))

	if refNode := node.Find("refECF"); refNode != nil {
		ref.RefECF.Modelo, _ = ParseECFModRef(pcn.ConteudoStr(refNode.Find("mod")))
		ref.RefECF.NECF = pcn.ConteudoStr(refNode.Find("nECF"))
		ref.RefECF.NCOO = pcn.ConteudoStr(refNode.Find("nCOO"))
	}

	l.n.Ide.NFref = append(l.n.Ide.NFref, ref)
}

// lerEmit le o emitente. Porte de LerEmit.
func (l *leitor) lerEmit(node *pcn.Node) {
	if node == nil {
		return
	}
	e := &l.n.Emit

	e.CNPJCPF = pcn.ConteudoCNPJCPF(node)
	// O ACBr NAO aplica DecodeHTMLEntities aqui, so no xNome de dest,
	// retirada, entrega e transporta.
	e.XNome = pcn.ConteudoStr(node.Find("xNome"))
	e.XFant = pcn.ConteudoStr(node.Find("xFant"))
	e.IE = pcn.ConteudoStr(node.Find("IE"))
	e.IEST = pcn.ConteudoStr(node.Find("IEST"))
	e.IM = pcn.ConteudoStr(node.Find("IM"))
	e.CNAE = pcn.ConteudoStr(node.Find("CNAE"))
	e.CRT, _ = ParseCRT(pcn.ConteudoStr(node.Find("CRT")))
	e.ISUFEmit = pcn.ConteudoStr(node.Find("ISUFEmit"))

	l.lerEmitEnderEmit(node.Find("enderEmit"))
}

// lerEmitEnderEmit le o endereco do emitente. Porte de LerEmitEnderEmit.
//
// Aplica os defaults de pais do ACBr na LEITURA, nao so na criacao: cPais
// ausente vira 1058 e xPais ausente vira BRASIL. Retirada e Entrega NAO tem
// esse tratamento, de proposito -- e assim no .pas.
func (l *leitor) lerEmitEnderEmit(node *pcn.Node) {
	if node == nil {
		return
	}
	end := &l.n.Emit.EnderEmit

	end.XLgr = pcn.ConteudoStr(node.Find("xLgr"))
	end.Nro = pcn.ConteudoStr(node.Find("nro"))
	end.XCpl = pcn.ConteudoStr(node.Find("xCpl"))
	end.XBairro = pcn.ConteudoStr(node.Find("xBairro"))
	end.CMun = pcn.ConteudoInt(node.Find("cMun"))
	end.XMun = pcn.ConteudoStr(node.Find("xMun"))
	end.UF = pcn.ConteudoStr(node.Find("UF"))
	end.CEP = pcn.ConteudoInt(node.Find("CEP"))
	end.CPais = pcn.ConteudoInt(node.Find("cPais"))

	if end.CPais == 0 {
		end.CPais = codigoPaisPadrao
	}

	end.XPais = pcn.ConteudoStr(node.Find("xPais"))

	if end.XPais == "" {
		end.XPais = paisPadrao
	}

	end.Fone = pcn.ConteudoStr(node.Find("fone"))
}

// lerAgropecuario le o grupo de operacao agropecuaria. Porte de
// LerAgropecuario.
//
// FINDANYNS em todos os campos -- este grupo e a unica parte do leitor, alem
// de infProt, que procura por local name.
func (l *leitor) lerAgropecuario(node *pcn.Node) {
	if node == nil {
		return
	}
	a := &l.n.Agropecuario

	for _, d := range node.FindAll("defensivo") {
		a.Defensivo = append(a.Defensivo, Defensivo{
			NReceituario: pcn.ConteudoStr(d.FindAnyNs("nReceituario")),
			CPFRespTec:   pcn.ConteudoStr(d.FindAnyNs("CPFRespTec")),
		})
	}

	if aux := node.FindAnyNs("guiaTransito"); aux != nil {
		a.GuiaTransito.UFGuia = pcn.ConteudoStr(aux.FindAnyNs("UFGuia"))
		a.GuiaTransito.TpGuia, _ = ParseTpGuia(pcn.ConteudoStr(aux.FindAnyNs("tpGuia")))
		a.GuiaTransito.SerieGuia = pcn.ConteudoStr(aux.FindAnyNs("serieGuia"))
		a.GuiaTransito.NGuia = pcn.ConteudoStr(aux.FindAnyNs("nGuia"))
	}
}

// lerAvulsa le o grupo do fisco emitente. Porte de LerAvulsa.
func (l *leitor) lerAvulsa(node *pcn.Node) {
	if node == nil {
		return
	}
	a := &l.n.Avulsa

	a.CNPJ = pcn.ConteudoStr(node.Find("CNPJ"))
	a.XOrgao = pcn.ConteudoStr(node.Find("xOrgao"))
	a.Matr = pcn.ConteudoStr(node.Find("matr"))
	a.XAgente = pcn.ConteudoStr(node.Find("xAgente"))
	a.Fone = pcn.ConteudoStr(node.Find("fone"))
	a.UF = pcn.ConteudoStr(node.Find("UF"))
	a.NDAR = pcn.ConteudoStr(node.Find("nDAR"))
	a.DEmi = pcn.ConteudoDataDef(node.Find("dEmi"))
	a.VDAR = pcn.ConteudoDe2(node.Find("vDAR"))
	a.RepEmi = pcn.ConteudoStr(node.Find("repEmi"))
	a.DPag = pcn.ConteudoDataDef(node.Find("dPag"))
}

// lerDest le o destinatario. Porte de LerDest.
//
// Atencao ao guard: quando o grupo dest NAO existe, o ACBr marca o
// destinatario como NAO CONTRIBUINTE antes de sair. E o unico lugar do leitor
// em que a ausencia de um node grava um valor.
func (l *leitor) lerDest(node *pcn.Node) {
	d := &l.n.Dest

	if node == nil {
		d.IndIEDest = pcn.InNaoContribuinte
		return
	}

	d.CNPJCPF = pcn.ConteudoCNPJCPF(node)

	if l.versao() >= 3 {
		d.IDEstrangeiro = pcn.ConteudoStr(node.Find("idEstrangeiro"))
	}

	d.XNome = pcn.ConteudoStr(node.Find("xNome"))

	if l.versao() >= 3 {
		d.IndIEDest, _ = pcn.ParseIndIEDest(pcn.ConteudoStr(node.Find("indIEDest")))
	}

	d.IE = pcn.ConteudoStr(node.Find("IE"))
	d.ISUF = pcn.ConteudoStr(node.Find("ISUF"))

	if l.versao() >= 3 {
		d.IM = pcn.ConteudoStr(node.Find("IM"))
	}

	d.Email = pcn.ConteudoStr(node.Find("email"))

	l.lerDestEnderDest(node.Find("enderDest"))
}

// lerDestEnderDest le o endereco do destinatario. Porte de LerDestEnderDest.
func (l *leitor) lerDestEnderDest(node *pcn.Node) {
	if node == nil {
		return
	}
	end := &l.n.Dest.EnderDest

	end.XLgr = pcn.ConteudoStr(node.Find("xLgr"))
	end.Nro = pcn.ConteudoStr(node.Find("nro"))
	end.XCpl = pcn.ConteudoStr(node.Find("xCpl"))
	end.XBairro = pcn.ConteudoStr(node.Find("xBairro"))
	end.CMun = pcn.ConteudoInt(node.Find("cMun"))
	end.XMun = pcn.ConteudoStr(node.Find("xMun"))
	end.UF = pcn.ConteudoStr(node.Find("UF"))
	end.CEP = pcn.ConteudoInt(node.Find("CEP"))
	end.CPais = pcn.ConteudoInt(node.Find("cPais"))

	if end.CPais == 0 {
		end.CPais = codigoPaisPadrao
	}

	end.XPais = pcn.ConteudoStr(node.Find("xPais"))

	if end.XPais == "" {
		end.XPais = paisPadrao
	}

	end.Fone = pcn.ConteudoStr(node.Find("fone"))
}

// lerRetirada le o local de retirada. Porte de LerRetirada.
func (l *leitor) lerRetirada(node *pcn.Node) {
	if node == nil {
		return
	}
	r := &l.n.Retirada

	r.CNPJCPF = pcn.ConteudoCNPJCPF(node)
	r.XNome = pcn.ConteudoStr(node.Find("xNome"))
	r.XLgr = pcn.ConteudoStr(node.Find("xLgr"))
	r.Nro = pcn.ConteudoStr(node.Find("nro"))
	r.XCpl = pcn.ConteudoStr(node.Find("xCpl"))
	r.XBairro = pcn.ConteudoStr(node.Find("xBairro"))
	r.CMun = pcn.ConteudoInt(node.Find("cMun"))
	r.XMun = pcn.ConteudoStr(node.Find("xMun"))
	r.UF = pcn.ConteudoStr(node.Find("UF"))
	r.CEP = pcn.ConteudoInt(node.Find("CEP"))
	r.CPais = pcn.ConteudoInt(node.Find("cPais"))
	r.XPais = pcn.ConteudoStr(node.Find("xPais"))
	r.Fone = pcn.ConteudoStr(node.Find("fone"))
	r.Email = pcn.ConteudoStr(node.Find("email"))
	r.IE = pcn.ConteudoStr(node.Find("IE"))
}

// lerEntrega le o local de entrega. Porte de LerEntrega.
func (l *leitor) lerEntrega(node *pcn.Node) {
	if node == nil {
		return
	}
	e := &l.n.Entrega

	e.CNPJCPF = pcn.ConteudoCNPJCPF(node)
	e.XNome = pcn.ConteudoStr(node.Find("xNome"))
	e.XLgr = pcn.ConteudoStr(node.Find("xLgr"))
	e.Nro = pcn.ConteudoStr(node.Find("nro"))
	e.XCpl = pcn.ConteudoStr(node.Find("xCpl"))
	e.XBairro = pcn.ConteudoStr(node.Find("xBairro"))
	e.CMun = pcn.ConteudoInt(node.Find("cMun"))
	e.XMun = pcn.ConteudoStr(node.Find("xMun"))
	e.UF = pcn.ConteudoStr(node.Find("UF"))
	e.CEP = pcn.ConteudoInt(node.Find("CEP"))
	e.CPais = pcn.ConteudoInt(node.Find("cPais"))
	e.XPais = pcn.ConteudoStr(node.Find("xPais"))
	e.Fone = pcn.ConteudoStr(node.Find("fone"))
	e.Email = pcn.ConteudoStr(node.Find("email"))
	e.IE = pcn.ConteudoStr(node.Find("IE"))
}

// lerDet le um item da nota. Porte de LerDet.
func (l *leitor) lerDet(node *pcn.Node) {
	if node == nil {
		return
	}

	item := NovoDetItem()
	// nItem vem do CONTADOR, nao do atributo @nItem do det -- e assim no
	// ACBr, e por isso uma nota com @nItem fora de ordem e renumerada.
	item.Prod.NItem = len(l.n.Det) + 1
	item.InfAdProd = pcn.ConteudoStr(node.Find("infAdProd"))

	l.lerDetProd(&item, node.Find("prod"))
	l.lerDetImposto(&item, node.Find("imposto"))

	if aux := node.Find("impostoDevol"); aux != nil {
		item.PDevol = pcn.ConteudoDe2(aux.Find("pDevol"))
		if ipi := aux.Find("IPI"); ipi != nil {
			item.VIPIDevol = pcn.ConteudoDe2(ipi.Find("vIPIDevol"))
		}
	}

	l.lerDetObsItem(&item, node.Find("obsItem"))

	item.VItem = pcn.ConteudoDe2(node.Find("vItem"))
	rtc.LerDFeReferenciado(node.Find("DFeReferenciado"), &item.DFeReferenciado)

	l.n.Det = append(l.n.Det, item)
}

// lerDetProd le o produto do item. Porte de LerDetProd.
func (l *leitor) lerDetProd(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}
	p := &item.Prod

	p.CProd = pcn.ConteudoStr(node.Find("cProd"))
	p.CEAN = pcn.ConteudoStr(node.Find("cEAN"))
	p.CBarra = pcn.ConteudoStr(node.Find("cBarra"))
	p.XProd = pcn.ConteudoStr(node.Find("xProd"))
	p.NCM = pcn.ConteudoStr(node.Find("NCM"))
	p.CEST = pcn.ConteudoStr(node.Find("CEST"))

	if l.versao() >= 4 {
		p.IndEscala, _ = ParseIndEscala(pcn.ConteudoStr(node.Find("indEscala")))
		p.CNPJFab = pcn.ConteudoStr(node.Find("CNPJFab"))
		p.CBenef = pcn.ConteudoStr(node.Find("cBenef"))

		// A tag e gCred; o campo, CredPresumido.
		for _, g := range node.FindAll("gCred") {
			p.CredPresumido = append(p.CredPresumido, CredPresumido{
				CCredPresumido: pcn.ConteudoStr(g.FindAnyNs("cCredPresumido")),
				PCredPresumido: pcn.ConteudoDe4(g.FindAnyNs("pCredPresumido")),
				VCredPresumido: pcn.ConteudoDe2(g.FindAnyNs("vCredPresumido")),
			})
		}
	}

	p.TpCredPresIBSZFM, _ = rtc.ParseTpCredPresIBSZFM(pcn.ConteudoStr(node.Find("tpCredPresIBSZFM")))
	p.EXTIPI = pcn.ConteudoStr(node.Find("EXTIPI"))
	p.CFOP = pcn.ConteudoStr(node.Find("CFOP"))
	p.UCom = pcn.ConteudoStr(node.Find("uCom"))
	p.QCom = pcn.ConteudoDe4(node.Find("qCom"))
	p.VUnCom = pcn.ConteudoDe10(node.Find("vUnCom"))
	p.VProd = pcn.ConteudoDe2(node.Find("vProd"))
	p.CEANTrib = pcn.ConteudoStr(node.Find("cEANTrib"))
	p.CBarraTrib = pcn.ConteudoStr(node.Find("cBarraTrib"))
	p.UTrib = pcn.ConteudoStr(node.Find("uTrib"))
	p.QTrib = pcn.ConteudoDe4(node.Find("qTrib"))
	p.VUnTrib = pcn.ConteudoDe10(node.Find("vUnTrib"))
	p.VFrete = pcn.ConteudoDe2(node.Find("vFrete"))
	p.VSeg = pcn.ConteudoDe2(node.Find("vSeg"))
	p.VDesc = pcn.ConteudoDe2(node.Find("vDesc"))
	p.VOutro = pcn.ConteudoDe2(node.Find("vOutro"))
	p.IndTot, _ = ParseIndicadorTotal(pcn.ConteudoStr(node.Find("indTot")))
	p.XPed = pcn.ConteudoStr(node.Find("xPed"))
	p.NItemPed = pcn.ConteudoStr(node.Find("nItemPed"))
	p.NRECOPI = pcn.ConteudoStr(node.Find("nRECOPI"))
	p.NFCI = pcn.ConteudoStr(node.Find("nFCI"))
	p.IndBemMovelUsado, _ = pcn.ParseIndicadorEx(pcn.ConteudoStr(node.Find("indBemMovelUsado")))

	// O conteudo do NVE esta no proprio node, nao num filho.
	for _, nve := range node.FindAll("NVE") {
		p.NVE = append(p.NVE, NVE{NVE: pcn.ConteudoStr(nve)})
	}

	for _, di := range node.FindAll("DI") {
		l.lerDetProdDI(item, di)
	}
	for _, de := range node.FindAll("detExport") {
		l.lerDetProdDetExport(item, de)
	}
	for _, r := range node.FindAll("rastro") {
		l.lerDetProdRastro(item, r)
	}

	l.lerDetProdVeicProd(item, node.Find("veicProd"))

	for _, m := range node.FindAll("med") {
		l.lerDetProdMed(item, m)
	}
	for _, a := range node.FindAll("arma") {
		l.lerDetProdArma(item, a)
	}

	l.lerDetProdComb(item, node.Find("comb"))
}

// lerDetProdDI le uma Declaracao de Importacao. Porte de LerDetProdDI.
func (l *leitor) lerDetProdDI(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	var di DI
	di.NDi = pcn.ConteudoStr(node.Find("nDI"))
	di.DDi = pcn.ConteudoDataDef(node.Find("dDI"))
	di.XLocDesemb = pcn.ConteudoStr(node.Find("xLocDesemb"))
	di.UFDesemb = pcn.ConteudoStr(node.Find("UFDesemb"))
	di.DDesemb = pcn.ConteudoDataDef(node.Find("dDesemb"))

	// tpViaTransp e tpIntermedio sao lidos com tcInt no ACBr e so entao
	// convertidos pelo StrTo... O inteiro vira string decimal no caminho,
	// entao "03" e aceito como "3" -- mais tolerante que a leitura por
	// string. Replicado por enteiroComoCodigo.
	di.TpViaTransp, _ = ParseTipoViaTransp(inteiroComoCodigo(node.Find("tpViaTransp")))
	di.VAFRMM = pcn.ConteudoDe2(node.Find("vAFRMM"))
	di.TpIntermedio, _ = ParseTipoIntermedio(inteiroComoCodigo(node.Find("tpIntermedio")))
	di.CNPJ = pcn.ConteudoCNPJCPF(node)
	di.UFTerceiro = pcn.ConteudoStr(node.Find("UFTerceiro"))
	di.CExportador = pcn.ConteudoStr(node.Find("cExportador"))

	for _, a := range node.FindAll("adi") {
		di.Adi = append(di.Adi, Adi{
			NAdicao: pcn.ConteudoInt(a.Find("nAdicao")),
			// A tag e nSeqAdic; o campo, nSeqAdi.
			NSeqAdi:     pcn.ConteudoInt(a.Find("nSeqAdic")),
			CFabricante: pcn.ConteudoStr(a.Find("cFabricante")),
			VDescDI:     pcn.ConteudoDe2(a.Find("vDescDI")),
			NDraw:       pcn.ConteudoStr(a.Find("nDraw")),
		})
	}

	item.Prod.DI = append(item.Prod.DI, di)
}

// lerDetProdDetExport le o detalhamento de exportacao. Porte de
// LerDetProdDetExport.
func (l *leitor) lerDetProdDetExport(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	var de DetExport
	de.NDraw = pcn.ConteudoStr(node.Find("nDraw"))

	if aux := node.Find("exportInd"); aux != nil {
		de.NRE = pcn.ConteudoStr(aux.Find("nRE"))
		de.ChNFe = pcn.ConteudoStr(aux.Find("chNFe"))
		de.QExport = pcn.ConteudoDe4(aux.Find("qExport"))
	}

	item.Prod.DetExport = append(item.Prod.DetExport, de)
}

// lerDetProdRastro le a rastreabilidade do item. Porte de LerDetProdRastro.
func (l *leitor) lerDetProdRastro(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	item.Prod.Rastro = append(item.Prod.Rastro, Rastro{
		NLote:  pcn.ConteudoStr(node.Find("nLote")),
		QLote:  pcn.ConteudoDe3(node.Find("qLote")),
		DFab:   pcn.ConteudoDataDef(node.Find("dFab")),
		DVal:   pcn.ConteudoDataDef(node.Find("dVal")),
		CAgreg: pcn.ConteudoStr(node.Find("cAgreg")),
	})
}

// lerDetProdVeicProd le o veiculo novo. Porte de LerDetProdVeicProd.
func (l *leitor) lerDetProdVeicProd(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}
	v := &item.Prod.VeicProd

	v.TpOP, _ = ParseTipoOperacao(pcn.ConteudoStr(node.Find("tpOp")))
	v.Chassi = pcn.ConteudoStr(node.Find("chassi"))
	v.CCor = pcn.ConteudoStr(node.Find("cCor"))
	v.XCor = pcn.ConteudoStr(node.Find("xCor"))
	v.Pot = pcn.ConteudoStr(node.Find("pot"))
	v.Cilin = pcn.ConteudoStr(node.Find("cilin"))
	v.PesoL = pcn.ConteudoStr(node.Find("pesoL"))
	v.PesoB = pcn.ConteudoStr(node.Find("pesoB"))
	v.NSerie = pcn.ConteudoStr(node.Find("nSerie"))
	v.TpComb = pcn.ConteudoStr(node.Find("tpComb"))
	v.NMotor = pcn.ConteudoStr(node.Find("nMotor"))
	v.CMT = pcn.ConteudoStr(node.Find("CMT"))
	v.Dist = pcn.ConteudoStr(node.Find("dist"))
	v.AnoMod = pcn.ConteudoInt(node.Find("anoMod"))
	v.AnoFab = pcn.ConteudoInt(node.Find("anoFab"))
	v.TpPint = pcn.ConteudoStr(node.Find("tpPint"))
	v.TpVeic = pcn.ConteudoInt(node.Find("tpVeic"))
	v.EspVeic = pcn.ConteudoInt(node.Find("espVeic"))
	v.VIN = pcn.ConteudoStr(node.Find("VIN"))
	v.CondVeic, _ = ParseCondicaoVeiculo(pcn.ConteudoStr(node.Find("condVeic")))
	v.CMod = pcn.ConteudoStr(node.Find("cMod"))
	v.CCorDENATRAN = pcn.ConteudoStr(node.Find("cCorDENATRAN"))
	v.Lota = pcn.ConteudoInt(node.Find("lota"))
	v.TpRest = pcn.ConteudoInt(node.Find("tpRest"))
}

// lerDetProdMed le o medicamento. Porte de LerDetProdMed.
//
// O leitor NAO condiciona por versao: le as tags das duas eras sem distincao,
// e quem decide o que emitir e o gerador. Replicado.
func (l *leitor) lerDetProdMed(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	item.Prod.Med = append(item.Prod.Med, Med{
		CProdANVISA:    pcn.ConteudoStr(node.Find("cProdANVISA")),
		XMotivoIsencao: pcn.ConteudoStr(node.Find("xMotivoIsencao")),
		NLote:          pcn.ConteudoStr(node.Find("nLote")),
		QLote:          pcn.ConteudoDe3(node.Find("qLote")),
		DFab:           pcn.ConteudoDataDef(node.Find("dFab")),
		DVal:           pcn.ConteudoDataDef(node.Find("dVal")),
		VPMC:           pcn.ConteudoDe2(node.Find("vPMC")),
	})
}

// lerDetProdArma le o armamento. Porte de LerDetProdArma.
func (l *leitor) lerDetProdArma(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	a := Arma{
		NSerie: pcn.ConteudoStr(node.Find("nSerie")),
		NCano:  pcn.ConteudoStr(node.Find("nCano")),
		Descr:  pcn.ConteudoStr(node.Find("descr")),
	}
	a.TpArma, _ = ParseTipoArma(pcn.ConteudoStr(node.Find("tpArma")))

	item.Prod.Arma = append(item.Prod.Arma, a)
}

// lerDetProdComb le o combustivel. Porte de LerDetProdComb.
//
// Tem duas pegadinhas de nome de tag, as duas replicadas:
//
//   - pMixGN e lido da tag qMixGN (nome de campo != nome de tag);
//   - UFcons e lido DUAS vezes no nivel do comb, com grafias diferentes:
//     comb.UFCons vem de <UFCons> e comb.ICMSCons.UFCons vem de <UFcons>.
//     Depois, se o grupo ICMSCons existir, ele sobrescreve o segundo com o
//     proprio <UFCons>.
func (l *leitor) lerDetProdComb(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}
	c := &item.Prod.Comb

	c.CProdANP = pcn.ConteudoInt(node.Find("cProdANP"))
	c.PMixGN = pcn.ConteudoDe4(node.Find("qMixGN"))
	c.DescANP = pcn.ConteudoStr(node.Find("descANP"))
	c.PGLP = pcn.ConteudoDe4(node.Find("pGLP"))
	c.PGNn = pcn.ConteudoDe4(node.Find("pGNn"))
	c.PGNi = pcn.ConteudoDe4(node.Find("pGNi"))
	c.VPart = pcn.ConteudoDe2(node.Find("vPart"))
	c.CODIF = pcn.ConteudoStr(node.Find("CODIF"))
	c.QTemp = pcn.ConteudoDe4(node.Find("qTemp"))
	c.UFCons = pcn.ConteudoStr(node.Find("UFCons"))
	c.ICMSCons.UFCons = pcn.ConteudoStr(node.Find("UFcons"))

	if aux := node.Find("CIDE"); aux != nil {
		c.CIDE.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		c.CIDE.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		c.CIDE.VCIDE = pcn.ConteudoDe2(aux.Find("vCIDE"))
	}

	if aux := node.Find("encerrante"); aux != nil {
		c.Encerrante.NBico = pcn.ConteudoInt(aux.Find("nBico"))
		c.Encerrante.NBomba = pcn.ConteudoInt(aux.Find("nBomba"))
		c.Encerrante.NTanque = pcn.ConteudoInt(aux.Find("nTanque"))
		c.Encerrante.VEncIni = pcn.ConteudoDe3(aux.Find("vEncIni"))
		c.Encerrante.VEncFin = pcn.ConteudoDe3(aux.Find("vEncFin"))
	}

	c.PBio = pcn.ConteudoDe4(node.Find("pBio"))

	for _, o := range node.FindAll("origComb") {
		oc := OrigComb{
			CUFOrig: pcn.ConteudoInt(o.Find("cUFOrig")),
			POrig:   pcn.ConteudoDe4(o.Find("pOrig")),
		}
		oc.IndImport, _ = ParseIndImport(pcn.ConteudoStr(o.Find("indImport")))
		c.OrigComb = append(c.OrigComb, oc)
	}

	if aux := node.Find("ICMSComb"); aux != nil {
		c.ICMS.VBCICMS = pcn.ConteudoDe2(aux.Find("vBCICMS"))
		c.ICMS.VICMS = pcn.ConteudoDe2(aux.Find("vICMS"))
		c.ICMS.VBCICMSST = pcn.ConteudoDe2(aux.Find("vBCICMSST"))
		c.ICMS.VICMSST = pcn.ConteudoDe2(aux.Find("vICMSST"))
	}

	if aux := node.Find("ICMSInter"); aux != nil {
		c.ICMSInter.VBCICMSSTDest = pcn.ConteudoDe2(aux.Find("vBCICMSSTDest"))
		c.ICMSInter.VICMSSTDest = pcn.ConteudoDe2(aux.Find("vICMSSTDest"))
	}

	if aux := node.Find("ICMSCons"); aux != nil {
		c.ICMSCons.VBCICMSSTCons = pcn.ConteudoDe2(aux.Find("vBCICMSSTCons"))
		c.ICMSCons.VICMSSTCons = pcn.ConteudoDe2(aux.Find("vICMSSTCons"))
		c.ICMSCons.UFCons = pcn.ConteudoStr(aux.Find("UFCons"))
	}
}

// gruposICMS e a lista de grupos de ICMS que o leitor percorre, na ORDEM do
// ACBr (LerDetImposto). A ordem e contrato: o leitor chama um por um e cada
// grupo encontrado SOBRESCREVE o anterior, entao vale o ULTIMO achado -- e
// nao o primeiro, como num PrimeiroDe.
//
// Na pratica o leiaute admite um grupo so por item, e a diferenca nao aparece;
// ela existe para XML fora do XSD, e esta replicada.
var gruposICMS = [...]string{
	"ICMS00", "ICMS02", "ICMS10", "ICMS15", "ICMS20", "ICMS30", "ICMS40",
	"ICMS51", "ICMS53", "ICMS60", "ICMS61", "ICMS70", "ICMS90",
	"ICMSSN101", "ICMSSN102", "ICMSSN201", "ICMSSN202", "ICMSSN500", "ICMSSN900",
	"ICMSPart", "ICMSST",
}

// lerDetImposto le os impostos do item. Porte de LerDetImposto.
func (l *leitor) lerDetImposto(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}
	imp := &item.Imposto

	imp.VTotTrib = pcn.ConteudoDe2(node.Find("vTotTrib"))

	icmsNode := node.Find("ICMS")
	for _, grupo := range gruposICMS {
		l.lerICMSGrupo(icmsNode, grupo, &imp.ICMS)
	}

	// Depois de lido o grupo, o ACBr REMAPEIA o CST quando o grupo e
	// ICMSPart ou ICMSST -- sao partilhas que o leiaute escreve com o CST
	// original, e o componente distingue por um membro proprio do enum.
	if icmsNode != nil {
		if icmsNode.FindAnyNs("ICMSPart") != nil {
			switch imp.ICMS.CST {
			case pcn.CST10:
				imp.ICMS.CST = pcn.CSTPart10
			case pcn.CST90:
				imp.ICMS.CST = pcn.CSTPart90
			}
		} else if icmsNode.FindAnyNs("ICMSST") != nil {
			switch imp.ICMS.CST {
			case pcn.CST41:
				imp.ICMS.CST = pcn.CSTRep41
			case pcn.CST60:
				imp.ICMS.CST = pcn.CSTRep60
			}
		}
	}

	if aux := node.Find("ICMSUFDest"); aux != nil {
		u := &imp.ICMSUFDest
		u.VBCUFDest = pcn.ConteudoDe2(aux.Find("vBCUFDest"))
		u.VBCFCPUFDest = pcn.ConteudoDe2(aux.Find("vBCFCPUFDest"))
		u.PFCPUFDest = pcn.ConteudoDe2(aux.Find("pFCPUFDest"))
		u.PICMSUFDest = pcn.ConteudoDe2(aux.Find("pICMSUFDest"))
		u.PICMSInter = pcn.ConteudoDe2(aux.Find("pICMSInter"))
		u.PICMSInterPart = pcn.ConteudoDe2(aux.Find("pICMSInterPart"))
		u.VFCPUFDest = pcn.ConteudoDe2(aux.Find("vFCPUFDest"))
		u.VICMSUFDest = pcn.ConteudoDe2(aux.Find("vICMSUFDest"))
		u.VICMSUFRemet = pcn.ConteudoDe2(aux.Find("vICMSUFRemet"))
	}

	l.lerIPI(imp, node.Find("IPI"))

	if aux := node.Find("II"); aux != nil {
		imp.II.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.II.VDespAdu = pcn.ConteudoDe2(aux.Find("vDespAdu"))
		imp.II.VII = pcn.ConteudoDe2(aux.Find("vII"))
		imp.II.VIOF = pcn.ConteudoDe2(aux.Find("vIOF"))
	}

	l.lerPIS(imp, node.Find("PIS"))

	if aux := node.Find("PISST"); aux != nil {
		s := &imp.PISST
		s.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		s.PPIS = pcn.ConteudoDe2(aux.Find("pPIS"))
		s.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		s.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		s.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
		s.IndSomaPISST, _ = ParseIndSomaPISST(pcn.ConteudoStr(aux.Find("indSomaPISST")))
	}

	l.lerCOFINS(imp, node.Find("COFINS"))

	if aux := node.Find("COFINSST"); aux != nil {
		s := &imp.COFINSST
		s.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		s.PCOFINS = pcn.ConteudoDe2(aux.Find("pCOFINS"))
		s.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		s.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		s.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))
		s.IndSomaCOFINSST, _ = ParseIndSomaCOFINSST(pcn.ConteudoStr(aux.Find("indSomaCOFINSST")))
	}

	l.lerISSQN(imp, node.Find("ISSQN"))

	rtc.LerISel(node.Find("IS"), &imp.ISel)
	rtc.LerIBSCBS(node.Find("IBSCBS"), &imp.IBSCBS)
}

// lerICMSGrupo le UM grupo de ICMS achatado. Porte do procedimento aninhado
// LerICMS de LerDetImposto.
//
// Os campos sao os mesmos para todos os 21 grupos: o ACBr nao distingue quais
// tags cada CST admite, apenas tenta ler todas. Tag ausente deixa o campo no
// zero.
func (l *leitor) lerICMSGrupo(icmsNode *pcn.Node, grupo string, v *ICMS) {
	if icmsNode == nil {
		return
	}
	node := icmsNode.Find(grupo)
	if node == nil {
		return
	}

	v.Orig, _ = pcn.ParseOrigemMercadoria(pcn.ConteudoStr(node.Find("orig")))
	v.CST, _ = pcn.ParseCSTIcms(pcn.ConteudoStr(node.Find("CST")))
	v.CSOSN, _ = pcn.ParseCSOSNIcms(pcn.ConteudoStr(node.Find("CSOSN")))
	v.ModBC, _ = ParseDeterminacaoBaseIcms(pcn.ConteudoStr(node.Find("modBC")))
	v.PRedBC = pcn.ConteudoDe2(node.Find("pRedBC"))
	v.VBC = pcn.ConteudoDe2(node.Find("vBC"))
	v.PICMS = pcn.ConteudoDe2(node.Find("pICMS"))
	v.VICMSOp = pcn.ConteudoDe2(node.Find("vICMSOp"))
	v.PDif = pcn.ConteudoDe4(node.Find("pDif"))
	v.VICMSDif = pcn.ConteudoDe2(node.Find("vICMSDif"))
	v.VICMS = pcn.ConteudoDe2(node.Find("vICMS"))
	v.VBCFCP = pcn.ConteudoDe2(node.Find("vBCFCP"))
	v.PFCP = pcn.ConteudoDe2(node.Find("pFCP"))
	v.VFCP = pcn.ConteudoDe2(node.Find("vFCP"))
	v.ModBCST, _ = ParseDeterminacaoBaseIcmsST(pcn.ConteudoStr(node.Find("modBCST")))
	v.PMVAST = pcn.ConteudoDe2(node.Find("pMVAST"))
	v.PRedBCST = pcn.ConteudoDe2(node.Find("pRedBCST"))
	v.VBCST = pcn.ConteudoDe2(node.Find("vBCST"))
	v.PICMSST = pcn.ConteudoDe2(node.Find("pICMSST"))
	v.VICMSST = pcn.ConteudoDe2(node.Find("vICMSST"))
	v.VBCFCPST = pcn.ConteudoDe2(node.Find("vBCFCPST"))
	v.PFCPST = pcn.ConteudoDe2(node.Find("pFCPST"))
	v.VFCPST = pcn.ConteudoDe2(node.Find("vFCPST"))
	v.UFST = pcn.ConteudoStr(node.Find("UFST"))
	v.PBCOp = pcn.ConteudoDe2(node.Find("pBCOp"))
	v.VBCSTRet = pcn.ConteudoDe2(node.Find("vBCSTRet"))
	v.VICMSSTRet = pcn.ConteudoDe2(node.Find("vICMSSTRet"))
	v.VICMSDeson = pcn.ConteudoDe2(node.Find("vICMSDeson"))
	v.VBCFCPSTRet = pcn.ConteudoDe2(node.Find("vBCFCPSTRet"))
	v.PFCPSTRet = pcn.ConteudoDe2(node.Find("pFCPSTRet"))
	v.VFCPSTRet = pcn.ConteudoDe2(node.Find("vFCPSTRet"))
	v.PST = pcn.ConteudoDe4(node.Find("pST"))
	v.MotDesICMS, _ = ParseMotivoDesoneracaoICMS(pcn.ConteudoStr(node.Find("motDesICMS")))
	v.PCredSN = pcn.ConteudoDe2(node.Find("pCredSN"))
	v.VCredICMSSN = pcn.ConteudoDe2(node.Find("vCredICMSSN"))
	v.VBCSTDest = pcn.ConteudoDe2(node.Find("vBCSTDest"))
	v.VICMSSTDest = pcn.ConteudoDe2(node.Find("vICMSSTDest"))
	v.PRedBCEfet = pcn.ConteudoDe2(node.Find("pRedBCEfet"))
	v.VBCEfet = pcn.ConteudoDe2(node.Find("vBCEfet"))
	v.PICMSEfet = pcn.ConteudoDe2(node.Find("pICMSEfet"))
	v.VICMSEfet = pcn.ConteudoDe2(node.Find("vICMSEfet"))
	v.VICMSSubstituto = pcn.ConteudoDe2(node.Find("vICMSSubstituto"))
	v.VICMSSTDeson = pcn.ConteudoDe2(node.Find("vICMSSTDeson"))
	v.MotDesICMSST, _ = ParseMotivoDesoneracaoICMS(pcn.ConteudoStr(node.Find("motDesICMSST")))
	v.PFCPDif = pcn.ConteudoDe4(node.Find("pFCPDif"))
	v.VFCPDif = pcn.ConteudoDe2(node.Find("vFCPDif"))
	v.VFCPEfet = pcn.ConteudoDe2(node.Find("vFCPEfet"))

	v.AdRemICMS = pcn.ConteudoDe4(node.Find("adRemICMS"))
	v.VICMSMono = pcn.ConteudoDe2(node.Find("vICMSMono"))
	v.AdRemICMSReten = pcn.ConteudoDe4(node.Find("adRemICMSReten"))
	v.VICMSMonoReten = pcn.ConteudoDe2(node.Find("vICMSMonoReten"))
	v.VICMSMonoDif = pcn.ConteudoDe2(node.Find("vICMSMonoDif"))
	v.AdRemICMSRet = pcn.ConteudoDe4(node.Find("adRemICMSRet"))
	v.VICMSMonoRet = pcn.ConteudoDe2(node.Find("vICMSMonoRet"))

	v.QBCMono = pcn.ConteudoDe2(node.Find("qBCMono"))
	v.QBCMonoReten = pcn.ConteudoDe2(node.Find("qBCMonoReten"))
	v.PRedAdRem = pcn.ConteudoDe2(node.Find("pRedAdRem"))

	// motRedAdRem so e lido quando ha reducao: sem o guard, todo item
	// monofasico sairia com motRedAdRem = "1" (transporte coletivo), que e o
	// ordinal zero do enum.
	if v.PRedAdRem != 0 {
		v.MotRedAdRem, _ = ParseMotRedAdRem(pcn.ConteudoStr(node.Find("motRedAdRem")))
	}

	v.QBCMonoRet = pcn.ConteudoDe2(node.Find("qBCMonoRet"))
	v.VICMSMonoOp = pcn.ConteudoDe2(node.Find("vICMSMonoOp"))
	v.QBCMonoDif = pcn.ConteudoDe2(node.Find("qBCMonoDif"))
	v.AdRemICMSDif = pcn.ConteudoDe4(node.Find("adRemICMSDif"))

	// indDeduzDeson nao passa por Parse: o ACBr compara a string crua com
	// "1" e "0" e, em qualquer outro conteudo, deixa TieNenhum.
	v.IndDeduzDeson = pcn.TieNenhum
	switch pcn.ConteudoStr(node.Find("indDeduzDeson")) {
	case "1":
		v.IndDeduzDeson = pcn.TieSim
	case "0":
		v.IndDeduzDeson = pcn.TieNao
	}

	v.CBenefRBC = pcn.ConteudoStr(node.Find("cBenefRBC"))
}

// lerIPI le o IPI do item. Porte do trecho IPI de LerDetImposto.
//
// O CST ganha um default que depende do TIPO DA NOTA, aplicado antes de ler
// os subgrupos: entrada nasce ipi53 (nao tributada na entrada) e saida nasce
// ipi03 (nao tributada na saida). Se nem IPITrib nem IPINT existirem, e esse
// default que fica.
func (l *leitor) lerIPI(imp *Imposto, node *pcn.Node) {
	if node == nil {
		return
	}

	imp.IPI.ClEnq = pcn.ConteudoStr(node.Find("clEnq"))
	imp.IPI.CNPJProd = pcn.ConteudoStr(node.Find("CNPJProd"))
	imp.IPI.CSelo = pcn.ConteudoStr(node.Find("cSelo"))
	imp.IPI.QSelo = pcn.ConteudoInt(node.Find("qSelo"))
	imp.IPI.CEnq = pcn.ConteudoStr(node.Find("cEnq"))

	switch l.n.Ide.TpNF {
	case pcn.TnEntrada:
		imp.IPI.CST = Ipi53
	case pcn.TnSaida:
		imp.IPI.CST = Ipi03
	}

	if aux := node.Find("IPITrib"); aux != nil {
		imp.IPI.CST, _ = ParseCstIpi(pcn.ConteudoStr(aux.Find("CST")))
		imp.IPI.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.IPI.QUnid = pcn.ConteudoDe4(aux.Find("qUnid"))
		imp.IPI.VUnid = pcn.ConteudoDe4(aux.Find("vUnid"))
		imp.IPI.PIPI = pcn.ConteudoDe2(aux.Find("pIPI"))
		imp.IPI.VIPI = pcn.ConteudoDe2(aux.Find("vIPI"))
	}

	// IPINT e lido DEPOIS de IPITrib: se os dois existirem, vale o CST do
	// IPINT. Replicado.
	if aux := node.Find("IPINT"); aux != nil {
		imp.IPI.CST, _ = ParseCstIpi(pcn.ConteudoStr(aux.Find("CST")))
	}
}

// lerPIS le o PIS do item. Porte do trecho PIS de LerDetImposto.
//
// Como no ICMS, os quatro subgrupos sao lidos em sequencia e o ULTIMO
// encontrado prevalece. Cada um le um conjunto DIFERENTE de campos.
func (l *leitor) lerPIS(imp *Imposto, node *pcn.Node) {
	if node == nil {
		return
	}

	if aux := node.Find("PISAliq"); aux != nil {
		imp.PIS.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(aux.Find("CST")))
		imp.PIS.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.PIS.PPIS = pcn.ConteudoDe2(aux.Find("pPIS"))
		imp.PIS.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
	}

	if aux := node.Find("PISQtde"); aux != nil {
		imp.PIS.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(aux.Find("CST")))
		imp.PIS.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		imp.PIS.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		imp.PIS.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
	}

	if aux := node.Find("PISNT"); aux != nil {
		imp.PIS.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(aux.Find("CST")))
	}

	if aux := node.Find("PISOutr"); aux != nil {
		imp.PIS.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(aux.Find("CST")))
		imp.PIS.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.PIS.PPIS = pcn.ConteudoDe2(aux.Find("pPIS"))
		imp.PIS.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		imp.PIS.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		imp.PIS.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
	}
}

// lerCOFINS le a COFINS do item. Porte do trecho COFINS de LerDetImposto,
// simetrico ao do PIS.
func (l *leitor) lerCOFINS(imp *Imposto, node *pcn.Node) {
	if node == nil {
		return
	}

	if aux := node.Find("COFINSAliq"); aux != nil {
		imp.COFINS.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(aux.Find("CST")))
		imp.COFINS.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.COFINS.PCOFINS = pcn.ConteudoDe2(aux.Find("pCOFINS"))
		imp.COFINS.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))
	}

	if aux := node.Find("COFINSQtde"); aux != nil {
		imp.COFINS.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(aux.Find("CST")))
		imp.COFINS.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		imp.COFINS.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		imp.COFINS.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))
	}

	if aux := node.Find("COFINSNT"); aux != nil {
		imp.COFINS.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(aux.Find("CST")))
	}

	if aux := node.Find("COFINSOutr"); aux != nil {
		imp.COFINS.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(aux.Find("CST")))
		imp.COFINS.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		imp.COFINS.PCOFINS = pcn.ConteudoDe2(aux.Find("pCOFINS"))
		imp.COFINS.QBCProd = pcn.ConteudoDe4(aux.Find("qBCProd"))
		imp.COFINS.VAliqProd = pcn.ConteudoDe4(aux.Find("vAliqProd"))
		imp.COFINS.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))
	}
}

// lerISSQN le o ISSQN do item. Porte do trecho ISSQN de LerDetImposto.
//
// DECLARADO E NAO LIDO: indISSRet existe na classe e o ACBr nunca o le do XML
// (so do .ini). Replicado -- o campo fica no zero.
func (l *leitor) lerISSQN(imp *Imposto, node *pcn.Node) {
	if node == nil {
		return
	}
	s := &imp.ISSQN

	s.VBC = pcn.ConteudoDe2(node.Find("vBC"))
	s.VAliq = pcn.ConteudoDe2(node.Find("vAliq"))
	s.VISSQN = pcn.ConteudoDe2(node.Find("vISSQN"))
	s.CMunFG = pcn.ConteudoInt(node.Find("cMunFG"))
	s.CListServ = pcn.ConteudoStr(node.Find("cListServ"))
	s.CSitTrib, _ = ParseISSQNcSitTrib(pcn.ConteudoStr(node.Find("cSitTrib")))
	s.VDeducao = pcn.ConteudoDe2(node.Find("vDeducao"))
	s.VOutro = pcn.ConteudoDe2(node.Find("vOutro"))
	s.VDescIncond = pcn.ConteudoDe2(node.Find("vDescIncond"))
	s.VDescCond = pcn.ConteudoDe2(node.Find("vDescCond"))
	s.VISSRet = pcn.ConteudoDe2(node.Find("vISSRet"))
	s.IndISS, _ = ParseIndISS(pcn.ConteudoStr(node.Find("indISS")))
	s.CServico = pcn.ConteudoStr(node.Find("cServico"))
	s.CMun = pcn.ConteudoInt(node.Find("cMun"))
	s.CPais = pcn.ConteudoInt(node.Find("cPais"))
	s.NProcesso = pcn.ConteudoStr(node.Find("nProcesso"))
	s.IndIncentivo, _ = pcn.ParseIndIncentivo(pcn.ConteudoStr(node.Find("indIncentivo")))
}

// lerDetObsItem le as observacoes do item. Porte de LerDetObsItem.
//
// obsCont e obsFisco do ITEM sao ocorrencia unica (ao contrario das do
// infAdic, que sao lista), e o rotulo vem do atributo xCampo.
func (l *leitor) lerDetObsItem(item *DetItem, node *pcn.Node) {
	if item == nil || node == nil {
		return
	}

	if aux := node.Find("obsCont"); aux != nil {
		item.ObsCont.XCampo = aux.Attr("xCampo")
		item.ObsCont.XTexto = pcn.ConteudoStr(aux.Find("xTexto"))
	}

	if aux := node.Find("obsFisco"); aux != nil {
		item.ObsFisco.XCampo = aux.Attr("xCampo")
		item.ObsFisco.XTexto = pcn.ConteudoStr(aux.Find("xTexto"))
	}
}

// lerTotal le os totais da nota. Porte de LerTotal.
func (l *leitor) lerTotal(node *pcn.Node) {
	if node == nil {
		return
	}
	t := &l.n.Total

	if aux := node.Find("ICMSTot"); aux != nil {
		v := &t.ICMSTot
		v.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		v.VICMS = pcn.ConteudoDe2(aux.Find("vICMS"))
		v.VICMSDeson = pcn.ConteudoDe2(aux.Find("vICMSDeson"))
		v.VFCPUFDest = pcn.ConteudoDe2(aux.Find("vFCPUFDest"))
		v.VICMSUFDest = pcn.ConteudoDe2(aux.Find("vICMSUFDest"))
		v.VICMSUFRemet = pcn.ConteudoDe2(aux.Find("vICMSUFRemet"))
		v.VFCP = pcn.ConteudoDe2(aux.Find("vFCP"))
		v.VBCST = pcn.ConteudoDe2(aux.Find("vBCST"))
		v.VST = pcn.ConteudoDe2(aux.Find("vST"))
		v.VFCPST = pcn.ConteudoDe2(aux.Find("vFCPST"))
		v.VFCPSTRet = pcn.ConteudoDe2(aux.Find("vFCPSTRet"))
		v.QBCMono = pcn.ConteudoDe2(aux.Find("qBCMono"))
		v.VICMSMono = pcn.ConteudoDe2(aux.Find("vICMSMono"))
		v.QBCMonoReten = pcn.ConteudoDe2(aux.Find("qBCMonoReten"))
		v.VICMSMonoReten = pcn.ConteudoDe2(aux.Find("vICMSMonoReten"))
		v.QBCMonoRet = pcn.ConteudoDe2(aux.Find("qBCMonoRet"))
		v.VICMSMonoRet = pcn.ConteudoDe2(aux.Find("vICMSMonoRet"))
		v.VProd = pcn.ConteudoDe2(aux.Find("vProd"))
		v.VFrete = pcn.ConteudoDe2(aux.Find("vFrete"))
		v.VSeg = pcn.ConteudoDe2(aux.Find("vSeg"))
		v.VDesc = pcn.ConteudoDe2(aux.Find("vDesc"))
		v.VII = pcn.ConteudoDe2(aux.Find("vII"))
		v.VIPI = pcn.ConteudoDe2(aux.Find("vIPI"))
		v.VIPIDevol = pcn.ConteudoDe2(aux.Find("vIPIDevol"))
		v.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
		v.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))
		v.VOutro = pcn.ConteudoDe2(aux.Find("vOutro"))
		v.VNF = pcn.ConteudoDe2(aux.Find("vNF"))
		v.VTotTrib = pcn.ConteudoDe2(aux.Find("vTotTrib"))
	}

	if aux := node.Find("ISSQNtot"); aux != nil {
		v := &t.ISSQNTot
		v.VServ = pcn.ConteudoDe2(aux.Find("vServ"))
		v.VBC = pcn.ConteudoDe2(aux.Find("vBC"))
		v.VISS = pcn.ConteudoDe2(aux.Find("vISS"))
		v.VPIS = pcn.ConteudoDe2(aux.Find("vPIS"))
		v.VCOFINS = pcn.ConteudoDe2(aux.Find("vCOFINS"))

		if l.versao() >= 3 {
			v.DCompet = pcn.ConteudoDataDef(aux.Find("dCompet"))
			v.VDeducao = pcn.ConteudoDe2(aux.Find("vDeducao"))
			v.VOutro = pcn.ConteudoDe2(aux.Find("vOutro"))
			v.VDescIncond = pcn.ConteudoDe2(aux.Find("vDescIncond"))
			v.VDescCond = pcn.ConteudoDe2(aux.Find("vDescCond"))
			v.VISSRet = pcn.ConteudoDe2(aux.Find("vISSRet"))

			// cRegTrib ausente vira "0" -- que e RTISSNenhum, o ULTIMO
			// membro do enum. Sem esse tratamento a tag vazia cairia no
			// fallback do Parse, que e o PRIMEIRO membro.
			codigo := pcn.ConteudoStr(aux.Find("cRegTrib"))
			if codigo == "" {
				codigo = "0"
			}
			v.CRegTrib, _ = pcn.ParseRegTribISSQN(codigo)
		}
	}

	if aux := node.Find("retTrib"); aux != nil {
		v := &t.RetTrib
		v.VRetPIS = pcn.ConteudoDe2(aux.Find("vRetPIS"))
		v.VRetCOFINS = pcn.ConteudoDe2(aux.Find("vRetCOFINS"))
		v.VRetCSLL = pcn.ConteudoDe2(aux.Find("vRetCSLL"))
		v.VBCIRRF = pcn.ConteudoDe2(aux.Find("vBCIRRF"))
		v.VIRRF = pcn.ConteudoDe2(aux.Find("vIRRF"))
		v.VBCRetPrev = pcn.ConteudoDe2(aux.Find("vBCRetPrev"))
		v.VRetPrev = pcn.ConteudoDe2(aux.Find("vRetPrev"))
	}

	rtc.LerISTot(node.Find("ISTot"), &t.ISTot)
	rtc.LerIBSCBSTot(node.Find("IBSCBSTot"), &t.IBSCBSTot)

	t.VNFTot = pcn.ConteudoDe2(node.Find("vNFTot"))
}

// lerTransp le o transporte. Porte de LerTransp.
func (l *leitor) lerTransp(node *pcn.Node) {
	if node == nil {
		return
	}
	t := &l.n.Transp

	t.ModFrete, _ = ParseModalidadeFrete(pcn.ConteudoStr(node.Find("modFrete")))
	t.Vagao = pcn.ConteudoStr(node.Find("vagao"))
	t.Balsa = pcn.ConteudoStr(node.Find("balsa"))

	if aux := node.Find("transporta"); aux != nil {
		t.Transporta.CNPJCPF = pcn.ConteudoCNPJCPF(aux)
		t.Transporta.XNome = pcn.ConteudoStr(aux.Find("xNome"))
		t.Transporta.IE = pcn.ConteudoStr(aux.Find("IE"))
		t.Transporta.XEnder = pcn.ConteudoStr(aux.Find("xEnder"))
		t.Transporta.XMun = pcn.ConteudoStr(aux.Find("xMun"))
		t.Transporta.UF = pcn.ConteudoStr(aux.Find("UF"))
	}

	if aux := node.Find("retTransp"); aux != nil {
		t.RetTransp.VServ = pcn.ConteudoDe2(aux.Find("vServ"))
		t.RetTransp.VBCRet = pcn.ConteudoDe2(aux.Find("vBCRet"))
		t.RetTransp.PICMSRet = pcn.ConteudoDe2(aux.Find("pICMSRet"))
		t.RetTransp.VICMSRet = pcn.ConteudoDe2(aux.Find("vICMSRet"))
		t.RetTransp.CFOP = pcn.ConteudoStr(aux.Find("CFOP"))
		t.RetTransp.CMunFG = pcn.ConteudoInt(aux.Find("cMunFG"))
	}

	if aux := node.Find("veicTransp"); aux != nil {
		t.VeicTransp.Placa = pcn.ConteudoStr(aux.Find("placa"))
		t.VeicTransp.UF = pcn.ConteudoStr(aux.Find("UF"))
		t.VeicTransp.RNTC = pcn.ConteudoStr(aux.Find("RNTC"))
	}

	for _, r := range node.FindAll("reboque") {
		t.Reboque = append(t.Reboque, Reboque{
			Placa: pcn.ConteudoStr(r.Find("placa")),
			UF:    pcn.ConteudoStr(r.Find("UF")),
			RNTC:  pcn.ConteudoStr(r.Find("RNTC")),
		})
	}

	for _, v := range node.FindAll("vol") {
		l.lerTranspVol(v)
	}
}

// lerTranspVol le um volume. Porte de LerTranspVol.
func (l *leitor) lerTranspVol(node *pcn.Node) {
	if node == nil {
		return
	}

	vol := Vol{
		QVol:  pcn.ConteudoInt(node.Find("qVol")),
		Esp:   pcn.ConteudoStr(node.Find("esp")),
		Marca: pcn.ConteudoStr(node.Find("marca")),
		NVol:  pcn.ConteudoStr(node.Find("nVol")),
		PesoL: pcn.ConteudoDe3(node.Find("pesoL")),
		PesoB: pcn.ConteudoDe3(node.Find("pesoB")),
	}

	// A tag do lacre e "lacres" (plural) POR OCORRENCIA, nao um grupo
	// "lacres" com filhos "lacre".
	for _, lac := range node.FindAll("lacres") {
		vol.Lacres = append(vol.Lacres, Lacre{NLacre: pcn.ConteudoStr(lac.Find("nLacre"))})
	}

	l.n.Transp.Vol = append(l.n.Transp.Vol, vol)
}

// lerCobr le a cobranca e o pagamento. Porte de LerCobr.
//
// Recebe o no infNFe: cobr e pag sao IRMAOS, e o ACBr procura os dois a
// partir do pai.
//
// DIVERGENCIA (perda de dado no ACBr, corrigida aqui): no original a variavel
// AuxNode e reaproveitada. Ela recebe o no `cobr` no inicio e so e reapontada
// para `pag` quando Versao >= 4. Na versao 3.xx, portanto, o laco de
// pagamentos procura <pag> DENTRO de <cobr> -- onde ele nao existe -- e
// nenhuma forma de pagamento e lida. Abaixo o no de pagamento e resolvido
// separadamente do de cobranca, e a 3.xx le os <pag> do lugar certo.
func (l *leitor) lerCobr(node *pcn.Node) {
	if node == nil {
		return
	}

	// FINDANYNS: cobr e pag sao procurados por local name no ACBr.
	if cobrNode := node.FindAnyNs("cobr"); cobrNode != nil {
		if fat := cobrNode.FindAnyNs("fat"); fat != nil {
			l.n.Cobr.Fat.NFat = pcn.ConteudoStr(fat.Find("nFat"))
			l.n.Cobr.Fat.VOrig = pcn.ConteudoDe2(fat.Find("vOrig"))
			l.n.Cobr.Fat.VDesc = pcn.ConteudoDe2(fat.Find("vDesc"))
			l.n.Cobr.Fat.VLiq = pcn.ConteudoDe2(fat.Find("vLiq"))
		}

		for _, d := range cobrNode.FindAll("dup") {
			l.n.Cobr.Dup = append(l.n.Cobr.Dup, Dup{
				NDup:  pcn.ConteudoStr(d.Find("nDup")),
				DVenc: pcn.ConteudoDataDef(d.Find("dVenc")),
				VDup:  pcn.ConteudoDe2(d.Find("vDup")),
			})
		}
	}

	// Pagamento so existe a partir da versao 3.
	if l.versao() < 3 {
		return
	}

	// Na 4.00 os pagamentos sao <pag><detPag>...; ate a 3.10 sao <pag>
	// repetidos direto no infNFe.
	var pagContainer *pcn.Node
	tagPag := "pag"
	if l.versao() >= 4 {
		pagContainer = node.FindAnyNs("pag")
		if pagContainer == nil {
			return
		}
		l.n.Pag.VTroco = pcn.ConteudoDe2(pagContainer.Find("vTroco"))
		tagPag = "detPag"
	} else {
		pagContainer = node
	}

	for _, p := range pagContainer.FindAllAnyNs(tagPag) {
		item := NovoPagItem()
		item.IndPag, _ = ParseIndicadorPagamento(pcn.ConteudoStr(p.Find("indPag")))
		item.TPag, _ = ParseFormaPagamento(pcn.ConteudoStr(p.Find("tPag")))
		item.XPag = pcn.ConteudoStr(p.Find("xPag"))
		item.VPag = pcn.ConteudoDe2(p.Find("vPag"))
		item.DPag = pcn.ConteudoDataDef(p.Find("dPag"))
		item.CNPJPag = pcn.ConteudoStr(p.Find("CNPJPag"))
		item.UFPag = pcn.ConteudoStr(p.Find("UFPag"))

		if card := p.Find("card"); card != nil {
			item.TpIntegra, _ = pcn.ParseTpIntegra(pcn.ConteudoStr(card.Find("tpIntegra")))
			item.CNPJ = pcn.ConteudoStr(card.Find("CNPJ"))
			item.TBand, _ = ParseBandeiraCartao(pcn.ConteudoStr(card.Find("tBand")))
			item.CAut = pcn.ConteudoStr(card.Find("cAut"))
			item.CNPJReceb = pcn.ConteudoStr(card.Find("CNPJReceb"))
			item.IDTermPag = pcn.ConteudoStr(card.Find("idTermPag"))
		}

		l.n.Pag.DetPag = append(l.n.Pag.DetPag, item)
	}
}

// lerInfAdic le as informacoes adicionais. Porte de LerInfAdic.
func (l *leitor) lerInfAdic(node *pcn.Node) {
	if node == nil {
		return
	}
	a := &l.n.InfAdic

	a.InfAdFisco = pcn.ConteudoStr(node.Find("infAdFisco"))
	a.InfCpl = pcn.ConteudoStr(node.Find("infCpl"))

	// DIVERGENCIA: o .pas:1537 faz
	// `if Attributes.Count > 0 then Items['xCampo'].Content` -- uma obsCont
	// com qualquer atributo que NAO seja xCampo derruba a leitura da nota
	// com access violation. Mesma natureza do forDia@dia de lerCana.
	for _, o := range node.FindAll("obsCont") {
		a.ObsCont = append(a.ObsCont, ObsCont{
			XCampo: o.Attr("xCampo"),
			XTexto: pcn.ConteudoStr(o.Find("xTexto")),
		})
	}

	for _, o := range node.FindAll("obsFisco") {
		a.ObsFisco = append(a.ObsFisco, ObsFisco{
			XCampo: o.Attr("xCampo"),
			XTexto: pcn.ConteudoStr(o.Find("xTexto")),
		})
	}

	for _, p := range node.FindAll("procRef") {
		ref := ProcRef{NProc: pcn.ConteudoStr(p.Find("nProc"))}
		ref.IndProc, _ = ParseIndicadorProcesso(pcn.ConteudoStr(p.Find("indProc")))
		ref.TpAto, _ = ParseTpAto(pcn.ConteudoStr(p.Find("tpAto")))
		a.ProcRef = append(a.ProcRef, ref)
	}
}

// lerInfIntermed le o intermediador. Porte de LerInfIntermed.
func (l *leitor) lerInfIntermed(node *pcn.Node) {
	if node == nil {
		return
	}

	l.n.InfIntermed.CNPJ = pcn.ConteudoStr(node.Find("CNPJ"))
	l.n.InfIntermed.IDCadIntTran = pcn.ConteudoStr(node.Find("idCadIntTran"))
}

// lerExporta le o grupo de exportacao. Porte de LerExporta.
func (l *leitor) lerExporta(node *pcn.Node) {
	if node == nil {
		return
	}
	e := &l.n.Exporta

	// O campo e UFembarq; a tag, UFEmbarq.
	e.UFEmbarq = pcn.ConteudoStr(node.Find("UFEmbarq"))
	e.XLocEmbarq = pcn.ConteudoStr(node.Find("xLocEmbarq"))
	e.UFSaidaPais = pcn.ConteudoStr(node.Find("UFSaidaPais"))
	e.XLocExporta = pcn.ConteudoStr(node.Find("xLocExporta"))
	e.XLocDespacho = pcn.ConteudoStr(node.Find("xLocDespacho"))
}

// lerCompra le o grupo de compras publicas. Porte de LerCompra.
func (l *leitor) lerCompra(node *pcn.Node) {
	if node == nil {
		return
	}

	l.n.Compra.XNEmp = pcn.ConteudoStr(node.Find("xNEmp"))
	l.n.Compra.XPed = pcn.ConteudoStr(node.Find("xPed"))
	l.n.Compra.XCont = pcn.ConteudoStr(node.Find("xCont"))
}

// lerCana le o grupo de aquisicao de cana. Porte de LerCana.
func (l *leitor) lerCana(node *pcn.Node) {
	if node == nil {
		return
	}
	c := &l.n.Cana

	c.Safra = pcn.ConteudoStr(node.Find("safra"))
	c.Ref = pcn.ConteudoStr(node.Find("ref"))
	c.QTotMes = pcn.ConteudoDe10(node.Find("qTotMes"))
	c.QTotAnt = pcn.ConteudoDe10(node.Find("qTotAnt"))
	c.QTotGer = pcn.ConteudoDe10(node.Find("qTotGer"))
	c.VFor = pcn.ConteudoDe2(node.Find("vFor"))
	c.VTotDed = pcn.ConteudoDe2(node.Find("vTotDed"))
	c.VLiqFor = pcn.ConteudoDe2(node.Find("vLiqFor"))

	for _, f := range node.FindAll("forDia") {
		// DIVERGENCIA: o ACBr faz StrToInt(Attributes.Items['dia'].Content)
		// sem Def e sem conferir se o atributo existe -- atributo ausente
		// derruba a leitura da nota inteira com access violation. Aqui vale
		// zero, como em todo atributo ausente do package.
		c.ForDia = append(c.ForDia, ForDia{
			Dia:  pcn.AtributoInt(f, "dia"),
			Qtde: pcn.ConteudoDe10(f.Find("qtde")),
		})
	}

	for _, d := range node.FindAll("deduc") {
		c.Deduc = append(c.Deduc, Deduc{
			XDed: pcn.ConteudoStr(d.Find("xDed")),
			VDed: pcn.ConteudoDe2(d.Find("vDed")),
		})
	}
}

// lerInfRespTec le o responsavel tecnico. Porte de LerInfRespTec.
func (l *leitor) lerInfRespTec(node *pcn.Node) {
	if node == nil {
		return
	}
	r := &l.n.InfRespTec

	r.CNPJ = pcn.ConteudoStr(node.Find("CNPJ"))
	r.XContato = pcn.ConteudoStr(node.Find("xContato"))
	r.Email = pcn.ConteudoStr(node.Find("email"))
	r.Fone = pcn.ConteudoStr(node.Find("fone"))
	r.IDCSRT = pcn.ConteudoInt(node.Find("idCSRT"))
	r.HashCSRT = pcn.ConteudoStr(node.Find("hashCSRT"))
}

// lerInfNFeSupl le o grupo suplementar da NFC-e. Porte de LerInfNFeSupl.
//
// O ACBr remove o CDATA do qrCode com StringReplace SEM rfReplaceAll, entao
// so a PRIMEIRA ocorrencia de cada marcador sai. Deterministico, replicado.
func (l *leitor) lerInfNFeSupl(node *pcn.Node) {
	if node == nil {
		return
	}

	qr := pcn.ConteudoStr(node.Find("qrCode"))
	qr = pcn.RemoverCDATAPrimeiraOcorrencia(qr)
	l.n.InfNFeSupl.QRCode = qr
	l.n.InfNFeSupl.URLChave = pcn.ConteudoStr(node.Find("urlChave"))
}

// lerInfPAA le o grupo do Provedor de Assinatura e Autorizacao. Porte de
// LerInfPAA.
//
// Os dois Exit intermediarios do original sao guards de navegacao: sem
// PAASignature nao ha SignatureValue, e sem RSAKeyValue nao ha Modulus nem
// Exponent.
func (l *leitor) lerInfPAA(node *pcn.Node) {
	if node == nil {
		return
	}
	p := &l.n.InfPAA

	p.CNPJPAA = pcn.ConteudoStr(node.Find("CNPJPAA"))

	sig := node.Find("PAASignature")
	if sig == nil {
		return
	}
	p.SignatureValue = pcn.ConteudoStr(sig.Find("SignatureValue"))

	rsa := sig.Find("RSAKeyValue")
	if rsa == nil {
		return
	}
	p.Modulus = pcn.ConteudoStr(rsa.Find("Modulus"))
	p.Exponent = pcn.ConteudoStr(rsa.Find("Exponent"))
}

// horaComoData converte a hora lida -- uma duracao desde a meia-noite -- num
// time.Time sem data, reproduzindo o TDateTime do Delphi, cuja parte de data
// fica zerada quando so a hora e informada.
//
// So e usada no hSaiEnt das versoes ate a 3.10; da 4.00 em diante a hora de
// saida passou a integrar o dhSaiEnt.
func horaComoData(d time.Duration) time.Time {
	if d == 0 {
		return time.Time{}
	}
	return time.Time{}.Add(d)
}

// inteiroComoCodigo le a tag como inteiro e devolve a representacao decimal,
// reproduzindo o caminho do ACBr quando ele passa ObterConteudo(..., tcInt) a
// uma funcao StrTo... que espera string.
//
// O efeito pratico e tolerancia a zero a esquerda: "03" chega como "3".
func inteiroComoCodigo(n *pcn.Node) string {
	if n == nil {
		return ""
	}
	return strconv.Itoa(pcn.ConteudoInt(n))
}
