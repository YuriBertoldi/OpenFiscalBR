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
	"io"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitor do XML da NFGas. Porte de TNFGasXmlReader
// (ACBrNFGas.XmlReader.pas), metodo a metodo.
//
// Regras do original preservadas de proposito:
//
//   - node nil sai sem fazer nada ("if not Assigned(ANode) then Exit");
//   - a escolha entre Find (nome qualificado) e FindAnyNs (local name) e a
//     mesma do .pas, campo a campo -- ver os comentarios marcados FIND;
//   - a precisao de cada campo (De2, De4, De6, De8, De10) e a declarada no
//     .pas, e varia para o mesmo nome de tag em grupos diferentes;
//   - o grupo ICMS e achatado, e indSemCST vem do no imposto.
//
// Divergencias deliberadas estao marcadas com DIVERGENCIA.

// LerXML le uma NFGas a partir de um io.Reader.
//
// Aceita as duas formas de raiz: <NFGas> avulsa e <nfgasProc>, que envolve
// a nota e o protocolo de autorizacao.
func LerXML(r io.Reader) (*NFGas, error) {
	doc, err := pcn.Parse(r)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// LerXMLString le uma NFGas a partir de uma string.
func LerXMLString(xml string) (*NFGas, error) {
	return LerBytes([]byte(xml))
}

// LerBytes le uma NFGas a partir de bytes. BOM UTF-8 e declaracao
// <?xml ...?> sao tolerados.
func LerBytes(dados []byte) (*NFGas, error) {
	doc, err := pcn.ParseBytes(dados)
	if err != nil {
		return nil, err
	}
	return lerDocumento(doc)
}

// lerDocumento e o porte de TNFGasXmlReader.LerXml.
//
// DIVERGENCIA: o original levanta excecao em cada validacao; aqui cada uma
// vira um erro sentinela, que o chamador testa com errors.Is.
func lerDocumento(doc *pcn.Document) (*NFGas, error) {
	if doc == nil || doc.Root == nil {
		return nil, ErrXMLVazio
	}

	n := &NFGas{}

	// FIND estrito em protNFGas, NFGas, infNFGas, infNFGasSupl e Signature.
	//
	// DIVERGENCIA (micro): a raiz e comparada pelo LOCAL NAME; o Delphi
	// compara Document.Root.Name, que carrega o prefixo. Um <ns:nfgasProc>
	// entra no ramo de processo aqui e cairia no else la -- onde falharia
	// em seguida no Find('infNFGas'). Na pratica DFe nao prefixa a raiz;
	// aqui fica a forma tolerante, coerente com o separador de lote.
	var nfgasNode *pcn.Node
	if doc.Root.Nome == TagNFGasProc {
		lerProtNFGas(doc.Root.Find(TagProtNFGas), n)
		nfgasNode = doc.Root.Find(TagNFGas)
	} else {
		nfgasNode = doc.Root
	}

	if nfgasNode == nil {
		return nil, ErrNFGasNaoEncontrada
	}

	infNFGasNode := nfgasNode.Find(TagInfNFGas)
	if infNFGasNode == nil {
		return nil, ErrXMLIncorreto
	}

	id, ok := infNFGasNode.AttrOk("Id")
	if !ok {
		return nil, ErrAtributoIDAusente
	}
	n.InfNFGas.ID = id

	versao, ok := infNFGasNode.AttrOk("versao")
	if !ok {
		return nil, ErrAtributoVersaoAusente
	}
	n.InfNFGas.Versao = pcn.StringToFloatDef(versao, 0)

	lerInfNFGas(infNFGasNode, n)
	lerInfNFGasSupl(nfgasNode.Find(TagInfNFGasSupl), &n.InfNFGasSupl)
	pcn.LerSignature(nfgasNode.Find("Signature"), &n.Signature)

	return n, nil
}

// lerProtNFGas le o protocolo de autorizacao. Porte de Ler_ProtNFGas.
func lerProtNFGas(node *pcn.Node, n *NFGas) {
	if node == nil || n == nil {
		return
	}
	// O ACBr le exatamente estes campos. NAO le infProt/@Id, infProt/infFisco
	// nem a assinatura do protocolo.
	pcn.LerInfProt(node, &n.ProcNFGas, "chNFGas")
}

// lerInfNFGas despacha os grupos do documento. Porte de Ler_InfNFGas.
func lerInfNFGas(node *pcn.Node, n *NFGas) {
	if node == nil || n == nil {
		return
	}

	lerIde(node.Find("ide"), n) // FIND estrito, diferente dos demais
	lerEmit(node.FindAnyNs("emit"), &n.Emit)
	lerDest(node.FindAnyNs("dest"), &n.Dest)
	lerInstalacao(node.FindAnyNs("instalacao"), &n.Instalacao)
	lerGSub(node.FindAnyNs("gSub"), &n.GSub)

	n.GVolContrat = nil
	for _, x := range node.FindAllAnyNs("gVolContrat") {
		n.GVolContrat = append(n.GVolContrat, lerGVolContrat(x))
	}

	n.GMed = nil
	for _, x := range node.FindAllAnyNs("gMed") {
		n.GMed = append(n.GMed, lerGMed(x))
	}

	n.Det = nil
	for _, x := range node.FindAllAnyNs("det") {
		n.Det = append(n.Det, lerDet(x))
	}

	lerTotal(node.FindAnyNs("total"), &n.Total)
	rtc.LerPgtoVinc(node.FindAnyNs("pgtoVinc"), &n.PgtoVinc)
	lerGFat(node.FindAnyNs("gFat"), &n.GFat)
	lerGAgencia(node.FindAnyNs("gAgencia"), &n.GAgencia)

	n.AutXML = nil
	for _, x := range node.FindAllAnyNs("autXML") {
		n.AutXML = append(n.AutXML, AutXML{CNPJCPF: pcn.ConteudoCNPJCPF(x)})
	}

	lerInfAdic(node.FindAnyNs("infAdic"), &n.InfAdic)
	// A TAG e gRespTec; o campo e infRespTec.
	lerInfRespTec(node.FindAnyNs("gRespTec"), &n.InfRespTec)
}

// lerIde le o grupo de identificacao. Porte de Ler_Ide.
func lerIde(node *pcn.Node, n *NFGas) {
	if node == nil || n == nil {
		return
	}
	ide := &n.Ide

	ide.CUF = pcn.ConteudoInt(node.FindAnyNs("cUF"))
	ide.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))
	ide.Modelo = pcn.ConteudoInt(node.FindAnyNs("mod"))
	ide.Serie = pcn.ConteudoInt(node.FindAnyNs("serie"))
	ide.NNF = pcn.ConteudoInt(node.FindAnyNs("nNF"))
	ide.CNF = pcn.ConteudoInt(node.FindAnyNs("cNF"))
	ide.CDV = pcn.ConteudoInt(node.FindAnyNs("cDV"))
	ide.DhEmi = pcn.ConteudoDataDef(node.FindAnyNs("dhEmi"))
	ide.TpEmis, _ = pcn.ParseTipoEmissao(pcn.ConteudoStr(node.FindAnyNs("tpEmis")))
	ide.NSiteAutoriz, _ = ParseSiteAutorizador(pcn.ConteudoStr(node.FindAnyNs("nSiteAutoriz")))
	ide.CMunFG = pcn.ConteudoInt(node.FindAnyNs("cMunFG"))
	ide.FinNFGas, _ = ParseFinalidadeNFGas(pcn.ConteudoStr(node.FindAnyNs("finNFGas")))
	ide.TpFat, _ = ParseTpFat(pcn.ConteudoStr(node.FindAnyNs("tpFat")))
	ide.VerProc = pcn.ConteudoStr(node.FindAnyNs("verProc"))
	ide.DhCont = pcn.ConteudoDataDef(node.FindAnyNs("dhCont"))
	ide.XJust = pcn.ConteudoStr(node.FindAnyNs("xJust"))
	ide.TpPagAnt, _ = pcn.ParseTpPagAnt(pcn.ConteudoStr(node.FindAnyNs("tpPagAnt")))

	rtc.LerGCompraGovReduzido(node.FindAnyNs("gCompraGov"), &ide.GCompraGov)
}

// lerEmit le o grupo do emitente. Porte de Ler_Emit.
func lerEmit(node *pcn.Node, e *Emit) {
	if node == nil || e == nil {
		return
	}
	e.CNPJ = pcn.ConteudoCNPJCPF(node)
	e.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))
	e.XNome = pcn.ConteudoStr(node.FindAnyNs("xNome"))
	e.XFant = pcn.ConteudoStr(node.FindAnyNs("xFant"))
	e.ISUFEmit = pcn.ConteudoStr(node.Find("ISUFEmit")) // FIND estrito

	lerEndereco(node.FindAnyNs("enderEmit"), &e.EnderEmit, true)
}

// lerEndereco le um grupo de endereco. Porte de Ler_EmitEnderEmit,
// Ler_DestEnderDest e Ler_gFatEnderCorresp.
//
// comPais distingue os dois casos do original: enderEmit e enderDest leem
// cPais e xPais (campos que o tipo TEndeEmi do XSD nem tem), enderCorresp
// nao le. Replicado.
func lerEndereco(node *pcn.Node, end *Endereco, comPais bool) {
	if node == nil || end == nil {
		return
	}
	end.XLgr = pcn.ConteudoStr(node.FindAnyNs("xLgr"))
	end.Nro = pcn.ConteudoStr(node.FindAnyNs("nro"))
	end.XCpl = pcn.ConteudoStr(node.FindAnyNs("xCpl"))
	end.XBairro = pcn.ConteudoStr(node.FindAnyNs("xBairro"))
	end.CMun = pcn.ConteudoInt(node.FindAnyNs("cMun"))
	end.XMun = pcn.ConteudoStr(node.FindAnyNs("xMun"))
	end.CEP = pcn.ConteudoInt(node.FindAnyNs("CEP"))
	end.UF = pcn.ConteudoStr(node.FindAnyNs("UF"))
	end.Fone = pcn.ConteudoStr(node.FindAnyNs("fone"))
	end.Email = pcn.ConteudoStr(node.FindAnyNs("email"))
	if comPais {
		end.CPais = pcn.ConteudoInt(node.FindAnyNs("cPais"))
		end.XPais = pcn.ConteudoStr(node.FindAnyNs("xPais"))
	}
}

// lerDest le o grupo do destinatario. Porte de Ler_Dest.
//
// A precedencia entre CNPJ/CPF, idOutros e idEstrangeiro e a do original:
// idOutros so e lido se nao houver CNPJ/CPF, e idEstrangeiro so se nao
// houver nenhum dos dois. As duas ultimas tags alimentam o MESMO campo.
//
// indIEDest NAO e lido do XML pelo ACBr -- existe apenas no formato .ini.
func lerDest(node *pcn.Node, d *Dest) {
	if node == nil || d == nil {
		return
	}
	d.XNome = pcn.ConteudoStr(node.FindAnyNs("xNome"))
	d.CNPJCPF = pcn.ConteudoCNPJCPF(node)

	if d.CNPJCPF == "" {
		if v := pcn.ConteudoStr(node.FindAnyNs("idOutros")); v != "" {
			d.IDEstrangeiro = v
			d.TagIDOrigem = "idOutros"
		} else if v := pcn.ConteudoStr(node.FindAnyNs("idEstrangeiro")); v != "" {
			d.IDEstrangeiro = v
			d.TagIDOrigem = "idEstrangeiro"
		}
	}

	d.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))
	d.IM = pcn.ConteudoStr(node.FindAnyNs("IM"))
	d.CNIS = pcn.ConteudoStr(node.FindAnyNs("cNIS"))
	d.NB = pcn.ConteudoStr(node.FindAnyNs("NB"))
	d.XNomeAdicional = pcn.ConteudoStr(node.FindAnyNs("xNomeAdicional"))

	lerEndereco(node.FindAnyNs("enderDest"), &d.EnderDest, true)
}

// lerInstalacao le o grupo da instalacao. Porte de Ler_Instalacao.
func lerInstalacao(node *pcn.Node, i *Instalacao) {
	if node == nil || i == nil {
		return
	}
	i.IDInstalacao = pcn.ConteudoStr(node.FindAnyNs("idInstalacao"))
	i.IDCodCliente = pcn.ConteudoStr(node.FindAnyNs("idCodCliente"))
	i.TpInstalacao, _ = ParseTpInstalacao(pcn.ConteudoStr(node.FindAnyNs("tpInstalacao")))
	i.NContrato = pcn.ConteudoStr(node.FindAnyNs("nContrato"))
	i.TpClasse, _ = ParseTpClasse(pcn.ConteudoStr(node.FindAnyNs("tpClasse")))
	i.XClasse = pcn.ConteudoStr(node.FindAnyNs("xClasse"))
	i.LatGPS = pcn.ConteudoStr(node.FindAnyNs("latGPS"))
	i.LongGPS = pcn.ConteudoStr(node.FindAnyNs("longGPS"))
	i.CodRoteiroLeitura = pcn.ConteudoStr(node.FindAnyNs("codRoteiroLeitura"))
}

// lerGSub le o grupo de substituicao. Porte de Ler_gSub.
func lerGSub(node *pcn.Node, g *GSub) {
	if node == nil || g == nil {
		return
	}
	g.ChNFGas = pcn.ConteudoStr(node.FindAnyNs("chNFGas"))
	g.MotSub, _ = ParseMotSub(pcn.ConteudoStr(node.FindAnyNs("motSub")))
	lerGNF(node.FindAnyNs("gNF"), &g.GNF)
}

// lerGNF le a nota substituida. Porte de Ler_gNF.
//
// CompetEmis e CompetApur vem como AAAAMM e so sao aceitas quando tem 6
// digitos, ano maior que zero e mes entre 1 e 12 -- fora disso ficam
// zeradas, sem erro, como no original.
func lerGNF(node *pcn.Node, g *GNF) {
	if node == nil || g == nil {
		return
	}
	g.CNPJ = pcn.ConteudoStr(node.FindAnyNs("CNPJ"))
	g.Serie = pcn.ConteudoStr(node.FindAnyNs("serie")) // string, nao int
	g.NNF = pcn.ConteudoInt(node.FindAnyNs("nNF"))
	g.CompetEmis = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetEmis")))
	g.CompetApur = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetApur")))
	g.Hash115 = pcn.ConteudoStr(node.FindAnyNs("hash115"))
}

// lerGVolContrat le um volume contratado. Porte de Ler_gVolContrat.
func lerGVolContrat(node *pcn.Node) GVolContrat {
	var v GVolContrat
	if node == nil {
		return v
	}
	v.NContrat = pcn.AtributoInt(node, "nContrat")
	v.TpVolContrat, _ = ParseVolContrat(pcn.ConteudoStr(node.FindAnyNs("tpVolContrat")))
	v.QUnidContrat = pcn.ConteudoDe6(node.FindAnyNs("qUnidContrat"))
	return v
}

// lerGMed le uma medicao. Porte de Ler_gMed.
func lerGMed(node *pcn.Node) GMed {
	var v GMed
	if node == nil {
		return v
	}
	v.NMed = pcn.AtributoInt(node, "nMed")
	v.IDEqp = pcn.ConteudoStr(node.FindAnyNs("idEqp"))
	v.DMedAnt = pcn.ConteudoDataDef(node.FindAnyNs("dMedAnt"))
	v.VMedAnt = pcn.ConteudoDe4(node.FindAnyNs("vMedAnt"))
	v.DMedAtu = pcn.ConteudoDataDef(node.FindAnyNs("dMedAtu"))
	v.VMedAtu = pcn.ConteudoDe4(node.FindAnyNs("vMedAtu"))
	v.TpEqp, _ = ParseTpEqp(pcn.ConteudoStr(node.FindAnyNs("tpEqp")))
	v.TpMedidor, _ = ParseTpMedidor(pcn.ConteudoStr(node.FindAnyNs("tpMedidor")))
	return v
}

// lerDet le um item do documento. Porte de Ler_Det.
//
// nItem, chNFGasAnt e nItemAnt vem em ATRIBUTO, nao em elemento.
func lerDet(node *pcn.Node) Det {
	var d Det
	if node == nil {
		return d
	}
	d.NItem = pcn.AtributoInt(node, "nItem")
	d.ChNFGasAnt = node.Attr("chNFGasAnt")
	d.NItemAnt = pcn.AtributoInt(node, "nItemAnt")

	lerGNormal(node.FindAnyNs("gNormal"), &d.GNormal)
	lerGAgregadora(node.FindAnyNs("gAgregadora"), &d.GAgregadora)
	return d
}

// lerGNormal le o grupo de item normal. Porte de Ler_DetgNormal.
//
// DIVERGENCIA: o original NAO limpa gTarif antes do laco, o que duplicaria
// itens se o mesmo objeto fosse relido. Aqui o slice e reinicializado.
// Numa leitura limpa -- o unico caso em que o original e definido -- o
// resultado e o mesmo.
func lerGNormal(node *pcn.Node, g *GNormal) {
	if node == nil || g == nil {
		return
	}

	g.GTarif = nil
	for _, x := range node.FindAll("gTarif") { // FIND estrito
		g.GTarif = append(g.GTarif, lerGTarif(x))
	}

	lerProd(node.FindAnyNs("prod"), &g.Prod)
	lerImposto(node.FindAnyNs("imposto"), &g.Imposto)
	lerGProcRef(node.FindAnyNs("gProcRef"), &g.GProcRef)
	g.InfAdProd = pcn.ConteudoStr(node.FindAnyNs("infAdProd"))
}

// lerGTarif le uma faixa tarifaria. Porte de Ler_gTarif.
func lerGTarif(node *pcn.Node) GTarif {
	var v GTarif
	if node == nil {
		return v
	}
	v.DIniTarif = pcn.ConteudoDataDef(node.FindAnyNs("dIniTarif"))
	v.DFimTarif = pcn.ConteudoDataDef(node.FindAnyNs("dFimTarif"))
	v.NAto = pcn.ConteudoStr(node.FindAnyNs("nAto"))
	v.AnoAto = pcn.ConteudoInt(node.FindAnyNs("anoAto"))
	v.TpFaixaCons, _ = ParseTpFaixaCons(pcn.ConteudoStr(node.FindAnyNs("tpFaixaCons")))
	v.VTarifAplic = pcn.ConteudoDe8(node.FindAnyNs("vTarifAplic"))
	return v
}

// lerProd le o produto do item. Porte de Ler_Prod.
//
// indDevolucao so e atribuido quando a tag tem conteudo -- do contrario o
// campo fica no valor anterior, como no original.
func lerProd(node *pcn.Node, p *Prod) {
	if node == nil || p == nil {
		return
	}
	p.IndOrigemQtd, _ = ParseIndOrigemQtd(pcn.ConteudoStr(node.FindAnyNs("indOrigemQtd")))

	lerGMedicao(node.FindAnyNs("gMedicao"), &p.GMedicao)

	p.CProd = pcn.ConteudoStr(node.FindAnyNs("cProd"))
	p.XProd = pcn.ConteudoStr(node.FindAnyNs("xProd"))
	p.CClass = pcn.ConteudoStr(node.FindAnyNs("cClass"))
	p.CFOP = pcn.ConteudoInt(node.FindAnyNs("CFOP"))
	p.UMed, _ = ParseUMedItem(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	p.QFaturada = pcn.ConteudoDe4(node.FindAnyNs("qFaturada"))
	p.VItem = pcn.ConteudoDe10(node.FindAnyNs("vItem"))
	p.FatorPCS = pcn.ConteudoDe4(node.FindAnyNs("fatorPCS"))
	p.FatorPTZ = pcn.ConteudoDe4(node.FindAnyNs("fatorPTZ"))
	p.FatorP = pcn.ConteudoDe4(node.FindAnyNs("fatorP"))
	p.FatorT = pcn.ConteudoDe4(node.FindAnyNs("fatorT"))
	p.VProd = pcn.ConteudoDe10(node.FindAnyNs("vProd"))

	if v := pcn.ConteudoStr(node.FindAnyNs("indDevolucao")); v != "" {
		p.IndDevolucao, _ = pcn.ParseIndicadorEx(v)
	}

	rtc.LerGPagAntecipadoProd(node.FindAnyNs("gPagAntecipado"), &p.GPagAntecipado)
}

// lerGMedicao le o grupo de medicao do produto. Porte de Ler_gMedicao.
func lerGMedicao(node *pcn.Node, g *GMedicao) {
	if node == nil || g == nil {
		return
	}
	g.NMed = pcn.ConteudoInt(node.FindAnyNs("nMed"))
	g.NContrat = pcn.ConteudoInt(node.FindAnyNs("nContrat"))
	lerGMedida(node.FindAnyNs("gMedida"), &g.GMedida)
	g.TpMotNaoLeitura, _ = ParseTpMotNaoLeitura(pcn.ConteudoStr(node.FindAnyNs("tpMotNaoLeitura")))
	g.XMotNaoLeitura = pcn.ConteudoStr(node.FindAnyNs("xMotNaoLeitura"))
}

// lerGMedida le a medida registrada. Porte de Ler_gMedida.
func lerGMedida(node *pcn.Node, g *GMedida) {
	if node == nil || g == nil {
		return
	}
	g.UMed, _ = ParseUMed(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	g.VMed = pcn.ConteudoDe4(node.FindAnyNs("vMed"))
}

// lerGAgregadora le o grupo agregador do item. Porte de Ler_DetgAgregadora.
func lerGAgregadora(node *pcn.Node, g *GAgregadora) {
	if node == nil || g == nil {
		return
	}
	g.CClass = pcn.ConteudoStr(node.FindAnyNs("cClass"))
	g.VTotDFe = pcn.ConteudoDe2(node.FindAnyNs("vTotDFe"))
}

// lerImposto le os tributos do item. Porte de Ler_Imposto.
//
// Note que lerICMS recebe o proprio no imposto, e nao um filho: e ele que
// procura o grupo ICMSxx e tambem le indSemCST, que fica um nivel acima.
func lerImposto(node *pcn.Node, i *Imposto) {
	if node == nil || i == nil {
		return
	}
	i.Orig, _ = pcn.ParseOrigemMercadoria(pcn.ConteudoStr(node.FindAnyNs("orig")))

	// IndSemCST e espelhado de lerICMS, que le a tag no no imposto. O tipo
	// e IndicadorEx (zero = nao informado): no ACBr o campo e TIndicador,
	// cujo ordinal ZERO e tiSim -- um imposto lido sem a tag ficava "sem
	// CST" e o proprio ACBr descartaria o grupo ICMS ao regerar
	// (DIVERGENCIA DE TIPO, ver classes.go).
	lerICMS(node, &i.ICMS)
	i.IndSemCST = i.ICMS.IndSemCST
	rtc.LerIBSCBS(node.FindAnyNs("IBSCBS"), &i.IBSCBS)
	lerPIS(node.FindAnyNs("PIS"), &i.PIS)
	lerCOFINS(node.FindAnyNs("COFINS"), &i.COFINS)
	lerRetTrib(node.FindAnyNs("retTrib"), &i.RetTrib)
	lerTxReg(node.FindAnyNs("TxReg"), &i.TxReg)
}

// lerICMS le o grupo de ICMS, ACHATANDO os grupos por CST. Porte de
// Ler_ICMS.
//
// Procura ICMS00, ICMS10, ICMS20, ICMS40, ICMS41, ICMS51, ICMS60, ICMS70 e
// ICMS90, nesta ordem, e usa o primeiro que existir. Dois pontos de
// fidelidade, ambos replicados:
//
//   - ICMS41 e procurado e nao existe no XSD da NFGas;
//   - de ICMS51, os campos pDif, vICMSOp e vICMSDif nao sao lidos.
//
// indSemCST vem do no imposto (ANode), e nao do ICMSxx, e so e atribuido
// quando tem conteudo.
//
// DIVERGENCIA: motDesICMS. O ACBr le a tag com tcInt e atribui o INTEIRO ao
// enum, ou seja, trata o codigo do leiaute como ORDINAL -- o codigo "1"
// (Taxi, ordinal 0) viraria mdiDeficienteFisico, e os codigos 16 e 90
// cairiam fora da faixa do enum. Aqui a conversao e pelo CODIGO, que e o
// que o leiaute define. Ha teste cobrindo.
func lerICMS(node *pcn.Node, icms *ICMS) {
	if node == nil || icms == nil {
		return
	}

	icmsNode := node.PrimeiroDe(gruposICMS...)
	if icmsNode == nil {
		return
	}

	icms.CST, _ = pcn.ParseCSTIcms(pcn.ConteudoStr(icmsNode.FindAnyNs("CST")))
	icms.ModBC, _ = ParseDeterminacaoBaseIcms(pcn.ConteudoStr(icmsNode.FindAnyNs("modBC")))
	icms.PRedBC = pcn.ConteudoDe2(icmsNode.FindAnyNs("pRedBC"))
	icms.VBC = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBC"))
	icms.PICMS = pcn.ConteudoDe2(icmsNode.FindAnyNs("pICMS"))
	icms.VICMS = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMS"))
	icms.VICMSDeson = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMSDeson"))
	icms.MotDesICMS, _ = ParseMotivoDesoneracaoICMS(pcn.ConteudoStr(icmsNode.FindAnyNs("motDesICMS")))
	icms.IndDeduzDeson, _ = pcn.ParseIndicadorEx(pcn.ConteudoStr(icmsNode.FindAnyNs("indDeduzDeson")))
	icms.CBenef = pcn.ConteudoStr(icmsNode.FindAnyNs("cBenef"))
	icms.ModBCST, _ = ParseDeterminacaoBaseIcmsST(pcn.ConteudoStr(icmsNode.FindAnyNs("modBCST")))
	icms.PMVAST = pcn.ConteudoDe2(icmsNode.FindAnyNs("pMVAST"))
	icms.PRedBCST = pcn.ConteudoDe2(icmsNode.FindAnyNs("pRedBCST"))
	icms.VBCST = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCST"))
	icms.PICMSST = pcn.ConteudoDe2(icmsNode.FindAnyNs("pICMSST"))
	icms.VICMSST = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMSST"))
	icms.VBCFCP = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCFCP"))
	icms.PFCPST = pcn.ConteudoDe2(icmsNode.FindAnyNs("pFCPST"))
	icms.VFCPST = pcn.ConteudoDe2(icmsNode.FindAnyNs("vFCPST"))
	icms.VBCFCPST = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCFCPST"))
	icms.VBCSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCSTRet"))
	icms.PICMSSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("pICMSSTRet"))
	icms.VICMSSubstituto = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMSSubstituto"))
	icms.VICMSSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMSSTRet"))
	icms.VBCFCPSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCFCPSTRet"))
	icms.PFCPSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("pFCPSTRet"))
	icms.VFCPSTRet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vFCPSTRet"))
	icms.PRedBCEfet = pcn.ConteudoDe2(icmsNode.FindAnyNs("pRedBCEfet"))
	icms.VBCEfet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vBCEfet"))
	icms.PICMSEfet = pcn.ConteudoDe2(icmsNode.FindAnyNs("pICMSEfet"))
	icms.VICMSEfet = pcn.ConteudoDe2(icmsNode.FindAnyNs("vICMSEfet"))
	icms.PFCP = pcn.ConteudoDe4(icmsNode.FindAnyNs("pFCP")) // De4, nao De2
	icms.VFCP = pcn.ConteudoDe2(icmsNode.FindAnyNs("vFCP"))

	// indSemCST vem do no imposto, um nivel acima do ICMSxx, e so e
	// atribuido quando a tag tem conteudo (ausente = TieNenhum).
	if v := pcn.ConteudoStr(node.FindAnyNs("indSemCST")); v != "" {
		icms.IndSemCST, _ = pcn.ParseIndicadorEx(v)
	}
}

// lerPIS le o grupo de PIS. Porte de Ler_PIS.
func lerPIS(node *pcn.Node, p *PIS) {
	if node == nil || p == nil {
		return
	}
	p.CST, _ = pcn.ParseCSTPis(pcn.ConteudoStr(node.FindAnyNs("CST")))
	p.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	p.PPIS = pcn.ConteudoDe4(node.FindAnyNs("pPIS")) // De4
	p.VPIS = pcn.ConteudoDe2(node.FindAnyNs("vPIS"))
}

// lerCOFINS le o grupo de COFINS. Porte de Ler_COFINS.
func lerCOFINS(node *pcn.Node, c *COFINS) {
	if node == nil || c == nil {
		return
	}
	c.CST, _ = pcn.ParseCSTCofins(pcn.ConteudoStr(node.FindAnyNs("CST")))
	c.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	c.PCOFINS = pcn.ConteudoDe4(node.FindAnyNs("pCOFINS")) // De4
	c.VCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vCOFINS"))
}

// lerRetTrib le os tributos retidos. Porte de Ler_RetTrib.
//
// A tag de COFINS e vRetCofins, nao vRetCOFINS.
// VBCIRRF existe na classe do ACBr e NAO e lido pelo leitor -- replicado.
func lerRetTrib(node *pcn.Node, r *RetTrib) {
	if node == nil || r == nil {
		return
	}
	r.VRetPIS = pcn.ConteudoDe2(node.FindAnyNs("vRetPIS"))
	r.VRetCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vRetCofins"))
	r.VRetCSLL = pcn.ConteudoDe2(node.FindAnyNs("vRetCSLL"))
	r.VIRRF = pcn.ConteudoDe2(node.FindAnyNs("vIRRF"))
}

// lerTxReg le a taxa de regulacao. Porte de Ler_TxReg.
func lerTxReg(node *pcn.Node, t *TxReg) {
	if node == nil || t == nil {
		return
	}
	t.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	t.PTaxa = pcn.ConteudoDe4(node.FindAnyNs("pTaxa")) // De4
	t.VTaxa = pcn.ConteudoDe2(node.FindAnyNs("vTaxa"))
}

// lerGProcRef le o grupo de item referente a processo.
// Porte de Ler_gProcRef.
//
// vItem e vProd sao De8 aqui e De10 em prod -- mesma tag, precisoes
// diferentes, como no original.
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

// lerGProc le um processo referenciado. Porte de Ler_gProc.
func lerGProc(node *pcn.Node) GProc {
	var v GProc
	if node == nil {
		return v
	}
	v.TpProc, _ = ParseTpProc(pcn.ConteudoStr(node.FindAnyNs("tpProc")))
	v.NProcesso = pcn.ConteudoStr(node.FindAnyNs("nProcesso"))
	return v
}

// lerTotal le os totais do documento. Porte de Ler_Total.
//
// Os campos de ICMSTot e de vRetTribTot sao ACHATADOS em Total, como no
// original.
func lerTotal(node *pcn.Node, t *Total) {
	if node == nil || t == nil {
		return
	}
	t.VProd = pcn.ConteudoDe2(node.FindAnyNs("vProd"))

	lerICMSTot(node.FindAnyNs("ICMSTot"), t)
	lerVRetTribTot(node.FindAnyNs("vRetTribTot"), t)

	t.VCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vCOFINS"))
	t.VPIS = pcn.ConteudoDe2(node.FindAnyNs("vPIS"))
	t.VTxReg = pcn.ConteudoDe2(node.FindAnyNs("vTxReg"))
	t.VNF = pcn.ConteudoDe2(node.FindAnyNs("vNF"))

	rtc.LerIBSCBSTot(node.FindAnyNs("IBSCBSTot"), &t.IBSCBSTot)

	t.VTotDFe = pcn.ConteudoDe2(node.FindAnyNs("vTotDFe"))
}

// lerICMSTot achata o grupo ICMSTot em Total. Porte de Ler_ICMSTot.
func lerICMSTot(node *pcn.Node, t *Total) {
	if node == nil || t == nil {
		return
	}
	t.VBC = pcn.ConteudoDe2(node.FindAnyNs("vBC"))
	t.VICMS = pcn.ConteudoDe2(node.FindAnyNs("vICMS"))
	t.VICMSDeson = pcn.ConteudoDe2(node.FindAnyNs("vICMSDeson"))
	t.VFCP = pcn.ConteudoDe2(node.FindAnyNs("vFCP"))
	t.VBCST = pcn.ConteudoDe2(node.FindAnyNs("vBCST"))
	t.VST = pcn.ConteudoDe2(node.FindAnyNs("vST"))
	t.VFCPST = pcn.ConteudoDe2(node.FindAnyNs("vFCPST"))
}

// lerVRetTribTot achata o grupo vRetTribTot em Total.
// Porte de Ler_vRetTribTot. A tag de COFINS e vRetCofins.
func lerVRetTribTot(node *pcn.Node, t *Total) {
	if node == nil || t == nil {
		return
	}
	t.VRetPIS = pcn.ConteudoDe2(node.FindAnyNs("vRetPIS"))
	t.VRetCOFINS = pcn.ConteudoDe2(node.FindAnyNs("vRetCofins"))
	t.VRetCSLL = pcn.ConteudoDe2(node.FindAnyNs("vRetCSLL"))
	t.VIRRF = pcn.ConteudoDe2(node.FindAnyNs("vIRRF"))
}

// lerGFat le o grupo de faturamento. Porte de Ler_gFat.
//
// DIVERGENCIA: CompetFat vem como AAAAMM e o original monta a string
// '01/MM/AAAA' para chamar StrToDate, que depende do formato de data do
// sistema operacional -- numa maquina com locale en-US o mesmo XML e lido
// diferente, ou levanta excecao. Aqui a data e montada direto.
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

	// enderCorresp NAO le cPais/xPais, diferente de enderEmit e enderDest.
	lerEndereco(node.FindAnyNs("enderCorresp"), &g.EnderCorresp, false)
	lerGPIX(node.FindAnyNs("gPIX"), &g.GPIX)

	g.InfAdFat = pcn.ConteudoStr(node.FindAnyNs("infAdFat"))
}

// lerGPIX le o grupo de cobranca por PIX. Porte de Ler_gFatgPIX.
func lerGPIX(node *pcn.Node, g *GPIX) {
	if node == nil || g == nil {
		return
	}
	g.URLQRCodePIX = pcn.ConteudoStr(node.FindAnyNs("urlQRCodePIX"))
}

// lerGAgencia le o grupo da agencia de atendimento. Porte de Ler_gAgencia.
func lerGAgencia(node *pcn.Node, g *GAgencia) {
	if node == nil || g == nil {
		return
	}
	g.NomeAgenciaAtend = pcn.ConteudoStr(node.FindAnyNs("nomeAgenciaAtend"))
	g.EnderAgenciaAtend = pcn.ConteudoStr(node.FindAnyNs("enderAgenciaAtend"))
	g.SitioAgenciaAtend = pcn.ConteudoStr(node.FindAnyNs("sitioAgenciaAtend"))
	g.InfAdReg = pcn.ConteudoStr(node.FindAnyNs("infAdReg"))

	g.GHistCons = nil
	for _, x := range node.FindAll("gHistCons") { // FIND estrito
		g.GHistCons = append(g.GHistCons, lerGHistCons(x))
	}
}

// lerGHistCons le um bloco do historico de consumo.
// Porte de Ler_gHistCons.
func lerGHistCons(node *pcn.Node) GHistCons {
	var v GHistCons
	if node == nil {
		return v
	}
	v.XHistorico = pcn.ConteudoStr(node.FindAnyNs("xHistorico"))
	v.MedMensal = pcn.ConteudoDe4(node.FindAnyNs("medMensal"))

	for _, x := range node.FindAll("gCons") { // FIND estrito
		v.GCons = append(v.GCons, lerGCons(x))
	}
	return v
}

// lerGCons le um mes do historico de consumo. Porte de Ler_gCons.
//
// CompetFat tem a mesma divergencia documentada em lerGFat.
func lerGCons(node *pcn.Node) GCons {
	var v GCons
	if node == nil {
		return v
	}
	v.CompetFat = pcn.ParseCompetenciaDef(pcn.ConteudoStr(node.FindAnyNs("CompetFat")))
	v.UMed, _ = ParseUMed(pcn.ConteudoStr(node.FindAnyNs("uMed")))
	v.QtdDias = pcn.ConteudoInt(node.FindAnyNs("qtdDias"))
	v.MedDiaria = pcn.ConteudoDe4(node.FindAnyNs("medDiaria"))
	v.Consumo = pcn.ConteudoDe4(node.FindAnyNs("consumo"))
	v.VFat = pcn.ConteudoDe4(node.FindAnyNs("vFat"))
	return v
}

// lerInfAdic le as informacoes adicionais. Porte de Ler_InfAdic.
//
// O leiaute admite ate 5 ocorrencias de infCpl, mas o ACBr le apenas a
// primeira -- ha um TODO aberto no proprio fonte. Replicado: o slice recebe
// no maximo um elemento.
func lerInfAdic(node *pcn.Node, i *InfAdic) {
	if node == nil || i == nil {
		return
	}
	i.InfAdFisco = pcn.ConteudoStr(node.FindAnyNs("infAdFisco"))

	i.InfCpl = nil
	if v := pcn.ConteudoStr(node.FindAnyNs("infCpl")); v != "" {
		i.InfCpl = append(i.InfCpl, v)
	}
}

// lerInfRespTec le o responsavel tecnico. Porte de Ler_InfRespTec.
// A tag do grupo e gRespTec -- ver lerInfNFGas.
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

// lerInfNFGasSupl le as informacoes suplementares.
// Porte de Ler_InfNFGasSupl.
//
// A limpeza do CDATA usa StringReplace SEM rfReplaceAll no original, ou
// seja, remove apenas a PRIMEIRA ocorrencia de cada marcacao. Como o
// comportamento e deterministico, foi replicado.
func lerInfNFGasSupl(node *pcn.Node, s *InfNFGasSupl) {
	if node == nil || s == nil {
		return
	}
	s.QrCodNFGas = pcn.RemoverCDATAPrimeiraOcorrencia(
		pcn.ConteudoStr(node.Find("qrCodNFGas")), // FIND estrito
	)
}
