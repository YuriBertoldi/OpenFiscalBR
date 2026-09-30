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
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitor de XML da NFAg -- porte 1:1 de TNFAgXmlReader
// (ACBrNFAg.XmlReader.pas), metodo a metodo, com a decisao Find/FindAnyNs
// copiada POR CAMPO do original:
//
//   - os grupos do topo (ide, emit, dest, ligacao, gSub, gMed, det, total,
//     gFat...) sao achados com Find (nome qualificado);
//   - DENTRO de det, gFat, gAgencia e gQualiAgua o original troca para
//     FindAnyNs (local name) -- exceto os proprios lacos FindAll('gTarif'),
//     FindAll('gHistCons'), FindAll('gCons') e FindAll('gAnalise'), que
//     usam o nome qualificado.
//
// A leitura e TOLERANTE: nenhuma falha de campo derruba o documento.

// LerXML le um documento NFAg (avulso ou processado) de um io.Reader.
func LerXML(r io.Reader) (*NFAg, error) {
	doc, err := pcn.Parse(r)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// LerXMLString le um documento NFAg de uma string.
func LerXMLString(xml string) (*NFAg, error) {
	return LerBytes([]byte(xml))
}

// LerBytes le um documento NFAg de um []byte.
func LerBytes(dados []byte) (*NFAg, error) {
	doc, err := pcn.ParseBytes(dados)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// LerArquivo le um documento NFAg de um arquivo em disco.
func LerArquivo(caminho string) (*NFAg, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, &ErroNFAg{Indice: -1, Arquivo: caminho, Err: err}
	}
	n, err := LerBytes(dados)
	if err != nil {
		return nil, &ErroNFAg{Indice: -1, Arquivo: caminho, Err: err}
	}
	return n, nil
}

// lerDocumento e o LerXml do TNFAgXmlReader.
//
// O original so reconhece a raiz processada como 'NFAgProc'; o XSD
// (procNFAg_v1.00.xsd) define 'nfagProc'. Aqui as duas grafias sao aceitas
// -- e tambem 'procNFAg', pela mesma tolerancia do porte da NFGas.
func lerDocumento(doc *pcn.Document) (*NFAg, error) {
	root := doc.Root
	n := &NFAg{}

	nfagNode := root
	switch root.Nome {
	case "NFAgProc", "nfagProc", "procNFAg":
		lerProtNFAg(root.FindAnyNs("protNFAg"), n)
		nfagNode = root.FindAnyNs("NFAg")
	}

	if nfagNode == nil {
		return nil, ErrNFAgNaoEncontrada
	}

	infNode := nfagNode.Find("infNFAg")
	if infNode == nil {
		infNode = nfagNode.FindAnyNs("infNFAg")
	}
	if infNode == nil {
		return nil, ErrXMLIncorreto
	}

	id, ok := infNode.AttrOk("Id")
	if !ok {
		return nil, ErrAtributoIDAusente
	}
	n.InfNFAg.ID = id

	versao, ok := infNode.AttrOk("versao")
	if !ok {
		return nil, ErrAtributoVersaoAusente
	}
	n.InfNFAg.Versao = pcn.StringToFloatDef(versao, 0)

	lerInfNFAg(infNode, n)
	lerInfNFAgSupl(nfagNode.FindAnyNs("infNFAgSupl"), &n.InfNFAgSupl)
	pcn.LerSignature(nfagNode.FindAnyNs("Signature"), &n.Signature)

	return n, nil
}

// lerProtNFAg le o protocolo de autorizacao. Porte de Ler_ProtNFAg.
// LerInfProt ja localiza o infProt dentro do protNFAg.
func lerProtNFAg(node *pcn.Node, n *NFAg) {
	if node == nil || n == nil {
		return
	}
	pcn.LerInfProt(node, &n.ProcNFAg, "chNFAg")
}

// lerInfNFAg le o corpo do documento. Porte de Ler_InfNFAg.
func lerInfNFAg(node *pcn.Node, n *NFAg) {
	if node == nil || n == nil {
		return
	}
	lerIde(node.Find("ide"), &n.Ide)
	lerEmit(node.Find("emit"), &n.Emit)
	lerDest(node.Find("dest"), &n.Dest)
	lerLigacao(node.Find("ligacao"), &n.Ligacao)
	lerGSub(node.Find("gSub"), &n.GSub)

	// o original NAO limpa gMed antes do laco (acumulo em reuso de objeto);
	// aqui o slice e sempre reinicializado -- DIVERGENCIA determinada no
	// porte da NFGas, mesmo racional.
	n.GMed = nil
	for _, x := range node.FindAll("gMed") {
		n.GMed = append(n.GMed, lerGMed(x))
	}

	lerGFatConjunto(node.Find("gFatConjunto"), &n.GFatConjunto)

	n.Det = nil // NFAg.Det.Clear do original
	for _, x := range node.FindAll("det") {
		n.Det = append(n.Det, lerDet(x))
	}

	lerTotal(node.Find("total"), &n.Total)
	rtc.LerPgtoVinc(node.Find("pgtoVinc"), &n.PgtoVinc)
	lerGFat(node.Find("gFat"), &n.GFat)
	lerGAgencia(node.Find("gAgencia"), &n.GAgencia)
	lerGQualiAgua(node.Find("gQualiAgua"), &n.GQualiAgua)

	// autXML tambem nao tem Clear no original; slice reinicializado.
	n.AutXML = nil
	for _, x := range node.FindAll("autXML") {
		n.AutXML = append(n.AutXML, AutXML{CNPJCPF: pcn.ConteudoCNPJCPF(x)})
	}

	lerInfAdic(node.Find("infAdic"), &n.InfAdic)
	lerInfPAA(node.Find("infPAA"), &n.InfPAA)
	lerInfRespTec(node.Find("gRespTec"), &n.InfRespTec)
}

// lerIde le a identificacao. Porte de Ler_Ide.
func lerIde(node *pcn.Node, ide *Ide) {
	if node == nil || ide == nil {
		return
	}
	ide.CUF = pcn.ConteudoInt(node.Find("cUF"))
	ide.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.Find("tpAmb")))
	ide.Modelo = pcn.ConteudoInt(node.Find("mod"))
	ide.Serie = pcn.ConteudoInt(node.Find("serie"))
	ide.NNF = pcn.ConteudoInt(node.Find("nNF"))
	ide.CNF = pcn.ConteudoInt(node.Find("cNF"))
	ide.CDV = pcn.ConteudoInt(node.Find("cDV"))
	ide.DhEmi = pcn.ConteudoDataDef(node.Find("dhEmi"))
	ide.TpEmis, _ = pcn.ParseTipoEmissao(pcn.ConteudoStr(node.Find("tpEmis")))
	ide.NSiteAutoriz, _ = ParseSiteAutorizador(pcn.ConteudoStr(node.Find("nSiteAutoriz")))
	ide.CMunFG = pcn.ConteudoInt(node.Find("cMunFG"))
	ide.FinNFAg, _ = ParseFinalidadeNFAg(pcn.ConteudoStr(node.Find("finNFAg")))
	// tpFat NAO e lido do XML pelo ACBr (Ler_Ide nao tem a tag) -- omissao
	// do original REPLICADA; o campo so e populado via .ini.
	ide.VerProc = pcn.ConteudoStr(node.Find("verProc"))
	ide.DhCont = pcn.ConteudoDataDef(node.Find("dhCont"))
	ide.XJust = pcn.ConteudoStr(node.Find("xJust"))
	ide.TpPagAnt, _ = pcn.ParseTpPagAnt(pcn.ConteudoStr(node.Find("tpPagAnt")))

	rtc.LerGCompraGovReduzido(node.Find("gCompraGov"), &ide.GCompraGov)
}

// lerEmit le o emitente. Porte de Ler_Emit.
func lerEmit(node *pcn.Node, e *Emit) {
	if node == nil || e == nil {
		return
	}
	e.CNPJ = pcn.ConteudoCNPJCPF(node)
	e.IE = pcn.ConteudoStr(node.Find("IE"))
	e.XNome = pcn.ConteudoStr(node.Find("xNome"))
	e.XFant = pcn.ConteudoStr(node.Find("xFant"))
	lerEndereco(node.Find("enderEmit"), &e.EnderEmit, false)
	e.ISUFEmit = pcn.ConteudoStr(node.Find("ISUFEmit"))
}

// lerDest le o destinatario. Porte de Ler_Dest.
//
// FIDELIDADE COM RESSALVA: o original NAO tem o guard `if not Assigned`
// deste metodo -- um XML sem <dest> estoura com access violation no Delphi.
// Aqui o guard existe (leitura tolerante); o comportamento com dest presente
// e identico.
func lerDest(node *pcn.Node, d *Dest) {
	if node == nil || d == nil {
		return
	}
	d.XNome = pcn.ConteudoStr(node.Find("xNome"))
	d.CNPJCPF = pcn.ConteudoCNPJCPF(node)
	if d.CNPJCPF == "" {
		d.IDOutros = pcn.ConteudoStr(node.Find("idOutros"))
	}
	d.IE = pcn.ConteudoStr(node.Find("IE"))
	d.IM = pcn.ConteudoStr(node.Find("IM"))
	d.CNIS = pcn.ConteudoStr(node.Find("cNIS"))
	d.NB = pcn.ConteudoStr(node.Find("NB"))
	d.XNomeAdicional = pcn.ConteudoStr(node.Find("xNomeAdicional"))
	lerEndereco(node.Find("enderDest"), &d.EnderDest, false)
}

// lerEndereco cobre enderEmit, enderDest (Find) e enderCorresp (FindAnyNs,
// anyNs=true) -- os campos sao identicos nos tres.
func lerEndereco(node *pcn.Node, e *Endereco, anyNs bool) {
	if node == nil || e == nil {
		return
	}
	busca := node.Find
	if anyNs {
		busca = node.FindAnyNs
	}
	e.XLgr = pcn.ConteudoStr(busca("xLgr"))
	e.Nro = pcn.ConteudoStr(busca("nro"))
	e.XCpl = pcn.ConteudoStr(busca("xCpl"))
	e.XBairro = pcn.ConteudoStr(busca("xBairro"))
	e.CMun = pcn.ConteudoInt(busca("cMun"))
	e.XMun = pcn.ConteudoStr(busca("xMun"))
	e.CEP = pcn.ConteudoInt(busca("CEP"))
	e.UF = pcn.ConteudoStr(busca("UF"))
	e.Fone = pcn.ConteudoStr(busca("fone"))
	e.Email = pcn.ConteudoStr(busca("email"))
}

// lerLigacao le a ligacao. Porte de Ler_Ligacao.
func lerLigacao(node *pcn.Node, l *Ligacao) {
	if node == nil || l == nil {
		return
	}
	l.IDLigacao = pcn.ConteudoStr(node.Find("idLigacao"))
	l.IDCodCliente = pcn.ConteudoStr(node.Find("idCodCliente"))
	l.TpLigacao, _ = ParseTpLigacao(pcn.ConteudoStr(node.Find("tpLigacao")))
	l.LatGPS = pcn.ConteudoStr(node.Find("latGPS"))
	l.LongGPS = pcn.ConteudoStr(node.Find("longGPS"))
	l.CodRoteiroLeitura = pcn.ConteudoStr(node.Find("codRoteiroLeitura"))
}

// lerGSub le o grupo de substituicao. Porte de Ler_gSub.
func lerGSub(node *pcn.Node, g *GSub) {
	if node == nil || g == nil {
		return
	}
	g.ChNFAg = pcn.ConteudoStr(node.Find("chNFAg"))
	g.MotSub, _ = ParseMotSub(pcn.ConteudoStr(node.Find("motSub")))
}

// lerGMed le um medidor. Porte de Ler_gMed.
//
// O original le @nMed com StrToInt (sem Def): atributo ausente aborta a
// nota com exception. Aqui vale 0 -- mesma tolerancia adotada na NFGas
// para @nPag.
func lerGMed(node *pcn.Node) GMed {
	m := GMed{}
	if node == nil {
		return m
	}
	m.NMed, _ = strconv.Atoi(node.Attr("nMed"))
	m.IDMedidor = pcn.ConteudoStr(node.Find("idMedidor"))
	m.DMedAnt = pcn.ConteudoDataDef(node.Find("dMedAnt"))
	m.DMedAtu = pcn.ConteudoDataDef(node.Find("dMedAtu"))
	return m
}

// lerGFatConjunto le a referencia de faturamento em conjunto.
// Porte de Ler_gFatConjunto.
func lerGFatConjunto(node *pcn.Node, g *GFatConjunto) {
	if node == nil || g == nil {
		return
	}
	g.ChNFAgFat = pcn.ConteudoStr(node.Find("chNFAgFat"))
}

// lerDet le um item. Porte de Ler_Det -- corpo achatado.
func lerDet(node *pcn.Node) Det {
	d := Det{}
	if node == nil {
		return d
	}
	d.NItem, _ = strconv.Atoi(node.Attr("nItem"))
	d.ChNFAgAnt = node.Attr("chNFAgAnt")
	d.NItemAnt, _ = strconv.Atoi(node.Attr("nItemAnt"))

	for _, x := range node.FindAll("gTarif") {
		d.GTarif = append(d.GTarif, lerGTarif(x))
	}

	lerProd(node.FindAnyNs("prod"), &d.Prod)
	lerImposto(node.FindAnyNs("imposto"), &d.Imposto)
	lerGProcRef(node.FindAnyNs("gProcRef"), &d.GProcRef)
	d.InfAdProd = pcn.ConteudoStr(node.FindAnyNs("infAdProd"))
	return d
}

// lerGTarif le uma faixa tarifaria. Porte de Ler_gTarif.
func lerGTarif(node *pcn.Node) GTarif {
	t := GTarif{}
	if node == nil {
		return t
	}
	t.DIniTarif = pcn.ConteudoDataDef(node.FindAnyNs("dIniTarif"))
	t.DFimTarif = pcn.ConteudoDataDef(node.FindAnyNs("dFimTarif"))
	t.NAto = pcn.ConteudoStr(node.FindAnyNs("nAto"))
	t.AnoAto = pcn.ConteudoInt(node.FindAnyNs("anoAto"))
	t.TpFaixaCons, _ = ParseTpFaixaCons(pcn.ConteudoStr(node.FindAnyNs("tpFaixaCons")))
	return t
}

// lerProd le o produto. Porte de Ler_Prod. Precisoes POR CAMPO do original:
// qFaturada/fatorPoluicao tcDe4; vItem/vProd tcDe10.
func lerProd(node *pcn.Node, p *Prod) {
	if node == nil || p == nil {
		return
	}
	p.IndOrigemQtd, _ = ParseIndOrigemQtd(pcn.ConteudoStr(node.FindAnyNs("indOrigemQtd")))

	lerGMedicao(node.FindAnyNs("gMedicao"), &p.GMedicao)

	p.CProd = pcn.ConteudoStr(node.FindAnyNs("cProd"))
	p.XProd = pcn.ConteudoStr(node.FindAnyNs("xProd"))
	p.CClass = pcn.ConteudoStr(node.FindAnyNs("cClass"))
	p.TpCategoria, _ = ParseTpCategoria(pcn.ConteudoStr(node.FindAnyNs("tpCategoria")))
	p.XCategoria = pcn.ConteudoStr(node.FindAnyNs("xCategoria"))
	p.QEconomias = pcn.ConteudoStr(node.FindAnyNs("qEconomias"))
	p.UMed, _ = ParseUMedFat(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	p.QFaturada = pcn.ConteudoDe4(node.FindAnyNs("qFaturada"))
	p.VItem = pcn.ConteudoDe10(node.FindAnyNs("vItem"))
	p.FatorPoluicao = pcn.ConteudoDe4(node.FindAnyNs("fatorPoluicao"))
	p.VProd = pcn.ConteudoDe10(node.FindAnyNs("vProd"))

	if v := pcn.ConteudoStr(node.FindAnyNs("indDevolucao")); v != "" {
		p.IndDevolucao, _ = pcn.ParseIndicadorEx(v)
	}

	rtc.LerGPagAntecipadoProd(node.FindAnyNs("gPagAntecipado"), &p.GPagAntecipado)
}

// lerGMedicao le a medicao do produto. Porte de Ler_gMedicao.
func lerGMedicao(node *pcn.Node, g *GMedicao) {
	if node == nil || g == nil {
		return
	}
	g.NMed = pcn.ConteudoInt(node.FindAnyNs("nMed"))

	lerGMedida(node.FindAnyNs("gMedida"), &g.GMedida)

	if v := pcn.ConteudoStr(node.FindAnyNs("tpMotNaoLeitura")); v != "" {
		g.TpMotNaoLeitura, _ = ParseTpMotNaoLeitura(v)
	}
}

// lerGMedida le a medida. Porte de Ler_gMedida (tudo tcDe2).
func lerGMedida(node *pcn.Node, g *GMedida) {
	if node == nil || g == nil {
		return
	}
	g.TpGrMed, _ = ParseTpGrMed(pcn.ConteudoStr(node.FindAnyNs("tpGrMed")))
	g.NUnidConsumo = pcn.ConteudoStr(node.FindAnyNs("nUnidConsumo"))
	g.VUnidConsumo = pcn.ConteudoDe2(node.FindAnyNs("vUnidConsumo"))
	g.UMed, _ = ParseUMedFat(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	g.VMedAnt = pcn.ConteudoDe2(node.FindAnyNs("vMedAnt"))
	g.VMedAtu = pcn.ConteudoDe2(node.FindAnyNs("vMedAtu"))
	g.VConst = pcn.ConteudoDe2(node.FindAnyNs("vConst"))
	g.VMed = pcn.ConteudoDe2(node.FindAnyNs("vMed"))
}

// lerImposto le os tributos do item. Porte de Ler_Imposto -- a NFAg nao
// tem ICMS.
func lerImposto(node *pcn.Node, i *Imposto) {
	if node == nil || i == nil {
		return
	}
	rtc.LerIBSCBS(node.FindAnyNs("IBSCBS"), &i.IBSCBS)
	lerPIS(node.FindAnyNs("PIS"), &i.PIS)
	lerCOFINS(node.FindAnyNs("COFINS"), &i.COFINS)
	lerRetTrib(node.FindAnyNs("retTrib"), &i.RetTrib)
	lerTFS(node.FindAnyNs("TFS"), &i.TFS)
	lerTFU(node.FindAnyNs("TFU"), &i.TFU)
}

// lerPIS le o grupo de PIS. Porte de Ler_PIS (pPIS tcDe4).
func lerPIS(node *pcn.Node, p *PIS) {
	if node == nil || p == nil {
		return
	}
	p.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(node.FindAnyNs("CST")))
	p.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	p.PPIS = pcn.ConteudoDe4(node.FindAnyNs("pPIS"))
	p.VPIS = pcn.ConteudoDe2(node.FindAnyNs("vPIS"))
}

// lerCOFINS le o grupo de COFINS. Porte de Ler_COFINS (pCOFINS tcDe4).
func lerCOFINS(node *pcn.Node, c *COFINS) {
	if node == nil || c == nil {
		return
	}
	c.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(node.FindAnyNs("CST")))
	c.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	c.PCOFINS = pcn.ConteudoDe4(node.FindAnyNs("pCOFINS"))
	c.VCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vCOFINS"))
}

// lerRetTrib le os tributos retidos. Porte de Ler_RetTrib.
// A tag de COFINS retida e vRetCofins; vBCIRRF NAO e lido (omissao do
// original replicada -- o campo so e populado por quem monta a nota).
func lerRetTrib(node *pcn.Node, r *RetTrib) {
	if node == nil || r == nil {
		return
	}
	r.VRetPIS = pcn.ConteudoDe2(node.FindAnyNs("vRetPIS"))
	r.VRetCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vRetCofins"))
	r.VRetCSLL = pcn.ConteudoDe2(node.FindAnyNs("vRetCSLL"))
	r.VIRRF = pcn.ConteudoDe2(node.FindAnyNs("vIRRF"))
}

// lerTFS le a Taxa de Fiscalizacao e Servicos. Porte de Ler_TFS.
func lerTFS(node *pcn.Node, t *TFS) {
	if node == nil || t == nil {
		return
	}
	t.VBCTFS = pcn.ConteudoDe2(node.FindAnyNs("vBCTFS"))
	t.PTFS = pcn.ConteudoDe2(node.FindAnyNs("pTFS"))
	t.VTFS = pcn.ConteudoDe2(node.FindAnyNs("vTFS"))
}

// lerTFU le a Taxa de Fiscalizacao e Uso. Porte de Ler_TFU.
func lerTFU(node *pcn.Node, t *TFU) {
	if node == nil || t == nil {
		return
	}
	t.VBCTFU = pcn.ConteudoDe2(node.FindAnyNs("vBCTFU"))
	t.PTFU = pcn.ConteudoDe2(node.FindAnyNs("pTFU"))
	t.VTFU = pcn.ConteudoDe2(node.FindAnyNs("vTFU"))
}

// lerGProcRef le o grupo de processo referenciado. Porte de Ler_gProcRef
// (vItem/vProd tcDe8, qFaturada tcDe4).
func lerGProcRef(node *pcn.Node, g *GProcRef) {
	if node == nil || g == nil {
		return
	}
	g.VItem = pcn.ConteudoDe8(node.FindAnyNs("vItem"))
	g.QFaturada = pcn.ConteudoDe4(node.FindAnyNs("qFaturada"))
	g.VProd = pcn.ConteudoDe8(node.FindAnyNs("vProd"))

	if v := pcn.ConteudoStr(node.FindAnyNs("indDevolucao")); v != "" {
		g.IndDevolucao, _ = pcn.ParseIndicadorEx(v)
	}

	g.GProc = nil
	for _, x := range node.FindAllAnyNs("gProc") {
		g.GProc = append(g.GProc, lerGProc(x))
	}
}

// lerGProc le um processo. Porte de Ler_gProc.
func lerGProc(node *pcn.Node) GProc {
	p := GProc{}
	if node == nil {
		return p
	}
	p.TpProc, _ = ParseTpProc(pcn.ConteudoStr(node.FindAnyNs("tpProc")))
	p.NProcesso = pcn.ConteudoStr(node.FindAnyNs("nProcesso"))
	return p
}

// lerTotal le os totais. Porte de Ler_Total.
func lerTotal(node *pcn.Node, t *Total) {
	if node == nil || t == nil {
		return
	}
	t.VProd = pcn.ConteudoDe2(node.FindAnyNs("vProd"))

	lerVRetTribTot(node.FindAnyNs("vRetTribTot"), t)

	t.VCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vCOFINS"))
	t.VPIS = pcn.ConteudoDe2(node.FindAnyNs("vPIS"))
	t.VTFS = pcn.ConteudoDe2(node.FindAnyNs("vTFS"))
	t.VTFU = pcn.ConteudoDe2(node.FindAnyNs("vTFU"))
	t.VNF = pcn.ConteudoDe2(node.FindAnyNs("vNF"))

	rtc.LerIBSCBSTot(node.FindAnyNs("IBSCBSTot"), &t.IBSCBSTot)

	t.VTotDFe = pcn.ConteudoDe2(node.FindAnyNs("vTotDFe"))
}

// lerVRetTribTot le os retidos do total. Porte de Ler_vRetTribTot
// (tag vRetCofins).
func lerVRetTribTot(node *pcn.Node, t *Total) {
	if node == nil || t == nil {
		return
	}
	t.VRetPIS = pcn.ConteudoDe2(node.FindAnyNs("vRetPIS"))
	t.VRetCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vRetCofins"))
	t.VRetCSLL = pcn.ConteudoDe2(node.FindAnyNs("vRetCSLL"))
	t.VIRRF = pcn.ConteudoDe2(node.FindAnyNs("vIRRF"))
}

// lerGFat le o faturamento. Porte de Ler_gFat.
//
// DIVERGENCIA (mesma do porte da NFGas): CompetFat vem como AAAAMM e o
// ACBr monta '01/MM/AAAA' com StrToDate, dependente de locale; aqui a data
// e montada direto.
func lerGFat(node *pcn.Node, g *GFat) {
	if node == nil || g == nil {
		return
	}
	g.CompetFat = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetFat")))
	g.DVencFat = pcn.ConteudoDataDef(node.FindAnyNs("dVencFat"))
	g.DApresFat = pcn.ConteudoDataDef(node.FindAnyNs("dApresFat"))
	g.DProxLeitura = pcn.ConteudoDataDef(node.FindAnyNs("dProxLeitura"))
	g.NFat = pcn.ConteudoStr(node.FindAnyNs("nFat"))
	g.CodBarras = pcn.ConteudoStr(node.FindAnyNs("codBarras"))
	g.CodDebAuto = pcn.ConteudoStr(node.FindAnyNs("codDebAuto"))
	g.CodBanco = pcn.ConteudoStr(node.FindAnyNs("codBanco"))
	g.CodAgencia = pcn.ConteudoStr(node.FindAnyNs("codAgencia"))

	lerEndereco(node.FindAnyNs("enderCorresp"), &g.EnderCorresp, true)
	if pix := node.FindAnyNs("gPIX"); pix != nil {
		g.GPIX.URLQRCodePIX = pcn.ConteudoStr(pix.FindAnyNs("urlQRCodePIX"))
	}
}

// lerGAgencia le o grupo da agencia. Porte de Ler_gAgencia.
func lerGAgencia(node *pcn.Node, g *GAgencia) {
	if node == nil || g == nil {
		return
	}
	g.Econ = pcn.ConteudoStr(node.FindAnyNs("econ"))
	g.EconAcumulada = pcn.ConteudoStr(node.FindAnyNs("econAcumulada"))
	g.SPrestador = pcn.ConteudoStr(node.FindAnyNs("sPrestador"))
	g.DEmissSelo = pcn.ConteudoDataDef(node.FindAnyNs("dEmissSelo"))
	g.SRegulador = pcn.ConteudoStr(node.FindAnyNs("sRegulador"))
	g.NAgenciaAtend = pcn.ConteudoStr(node.FindAnyNs("nAgenciaAtend"))
	g.EnderAgenciaAtend = pcn.ConteudoStr(node.FindAnyNs("enderAgenciaAtend"))

	g.GHistCons = nil // gHistCons.Clear do original
	for _, x := range node.FindAll("gHistCons") {
		g.GHistCons = append(g.GHistCons, lerGHistCons(x))
	}
}

// lerGHistCons le um bloco do historico. Porte de Ler_gHistCons.
func lerGHistCons(node *pcn.Node) GHistCons {
	h := GHistCons{}
	if node == nil {
		return h
	}
	h.XHistorico = pcn.ConteudoStr(node.FindAnyNs("xHistorico"))
	h.MedMensal = pcn.ConteudoDe4(node.FindAnyNs("medMensal"))

	for _, x := range node.FindAll("gCons") {
		h.GCons = append(h.GCons, lerGCons(x))
	}
	return h
}

// lerGCons le um mes de consumo. Porte de Ler_gCons -- qtdDias e STRING
// (tcStr) e volFat e o valor faturado.
func lerGCons(node *pcn.Node) GCons {
	c := GCons{}
	if node == nil {
		return c
	}
	c.CompetFat = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetFat")))
	c.UMed, _ = ParseUMedFat(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	c.QtdDias = pcn.ConteudoStr(node.FindAnyNs("qtdDias"))
	c.MedDiaria = pcn.ConteudoDe4(node.FindAnyNs("medDiaria"))
	c.Consumo = pcn.ConteudoDe4(node.FindAnyNs("consumo"))
	c.VolFat = pcn.ConteudoDe2(node.FindAnyNs("volFat"))
	return c
}

// lerGQualiAgua le a qualidade da agua. Porte de Ler_gQualiAgua.
func lerGQualiAgua(node *pcn.Node, g *GQualiAgua) {
	if node == nil || g == nil {
		return
	}
	g.CompetAnalise = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetAnalise")))
	g.Conclusao = pcn.ConteudoStr(node.FindAnyNs("Conclusao"))
	g.CProcesso = pcn.ConteudoStr(node.FindAnyNs("cProcesso"))
	g.SistemaAbast = pcn.ConteudoStr(node.FindAnyNs("SistemaAbast"))

	g.GAnalise = nil // gAnalise.Clear do original
	for _, x := range node.FindAll("gAnalise") {
		g.GAnalise = append(g.GAnalise, lerGAnalise(x))
	}
}

// lerGAnalise le um item analisado. Porte de Ler_gAnalise.
func lerGAnalise(node *pcn.Node) GAnalise {
	a := GAnalise{}
	if node == nil {
		return a
	}
	a.XItemAnalisado = pcn.ConteudoStr(node.FindAnyNs("xItemAnalisado"))
	a.NAmostraMinima = pcn.ConteudoStr(node.FindAnyNs("nAmostraMinima"))
	a.NAmostraAnalisada = pcn.ConteudoStr(node.FindAnyNs("nAmostraAnalisada"))
	a.NAmostraFPadrao = pcn.ConteudoStr(node.FindAnyNs("nAmostraFPadrao"))
	a.NAmostraDPadrao = pcn.ConteudoStr(node.FindAnyNs("nAmostraDPadrao"))
	a.NMediaMensal = pcn.ConteudoStr(node.FindAnyNs("nMediaMensal"))
	a.XValorReferencia = pcn.ConteudoStr(node.FindAnyNs("xValorReferencia"))
	return a
}

// lerInfAdic le as informacoes adicionais. Porte de Ler_InfAdic.
// infCpl: leiaute admite 5 ocorrencias; o ACBr le so a primeira (TODO
// aberto la), replicado aqui com o slice preenchido com uma posicao.
func lerInfAdic(node *pcn.Node, i *InfAdic) {
	if node == nil || i == nil {
		return
	}
	i.InfAdFisco = pcn.ConteudoStr(node.FindAnyNs("infAdFisco"))
	if v := pcn.ConteudoStr(node.FindAnyNs("infCpl")); v != "" {
		i.InfCpl = []string{v}
	}
}

// lerInfPAA le o prestador de apoio ao abastecimento. Porte de Ler_InfPAA.
func lerInfPAA(node *pcn.Node, p *InfPAA) {
	if node == nil || p == nil {
		return
	}
	p.CNPJPAA = pcn.ConteudoStr(node.FindAnyNs("CNPJPAA"))
}

// lerInfRespTec le o responsavel tecnico. Porte de Ler_InfRespTec.
func lerInfRespTec(node *pcn.Node, r *InfRespTec) {
	if node == nil || r == nil {
		return
	}
	r.CNPJ = pcn.ConteudoStr(node.FindAnyNs("CNPJ"))
	r.XContato = pcn.ConteudoStr(node.FindAnyNs("xContato"))
	r.Email = pcn.ConteudoStr(node.FindAnyNs("email"))
	r.Fone = pcn.ConteudoStr(node.FindAnyNs("fone"))
	r.IDCSRT = pcn.ConteudoInt(node.FindAnyNs("idCSRT"))
	r.HashCSRT = pcn.ConteudoStr(node.FindAnyNs("hashCSRT"))
}

// lerInfNFAgSupl le as informacoes suplementares. Porte de Ler_InfNFAgSupl:
// o CDATA e removido so na PRIMEIRA ocorrencia de cada marcador
// (StringReplace sem rfReplaceAll), replicado.
func lerInfNFAgSupl(node *pcn.Node, s *InfNFAgSupl) {
	if node == nil || s == nil {
		return
	}
	v := pcn.ConteudoStr(node.Find("qrCodNFAg"))
	v = strings.Replace(v, "<![CDATA[", "", 1)
	v = strings.Replace(v, "]]>", "", 1)
	s.QrCodNFAg = v
}
