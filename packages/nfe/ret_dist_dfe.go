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
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitor do retorno da Distribuicao DF-e. Porte de TRetDistDFeInt
// (ACBrDFeComum.RetDistDFeInt.pas).
//
// E o unico retorno da NF-e em que cada item vem COMPRIMIDO: o conteudo do
// <docZip> e gzip em base64, e dentro dele ha um documento XML inteiro --
// um resumo, uma nota processada ou um evento processado.
//
// Duas decisoes do original governam a leitura:
//
//  1. o tipo de cada item e decidido pelo NOME DA RAIZ do documento
//     descomprimido (:817-855), e NAO pelo atributo `schema` do docZip. Os
//     dois concordam na pratica, mas e a raiz que manda;
//  2. a leitura e TOLERANTE: um item que nao case com nenhuma das quatro
//     raizes conhecidas fica com o XML cru e os campos vazios, sem erro.
//     Preservado -- um lote de cinquenta documentos nao pode falhar inteiro
//     porque a RFB acrescentou um tipo novo.

// ResDFe e o resumo de um documento. Porte de TresDFe (:176).
//
// Vem preenchido tanto do schema `resNFe` quanto do `procNFe`: no primeiro
// os campos sao lidos diretamente, no segundo sao extraidos de dentro da
// nota (ide, emit, total e protNFe).
type ResDFe struct {
	ChDFe    string          // chNFe
	CNPJCPF  string          // CNPJ, ou CPF quando aquele falta
	XNome    string          // xNome
	IE       string          // IE
	DhEmi    time.Time       // dhEmi
	TpNF     pcn.TipoNFe     // tpNF
	VNF      float64         // vNF
	DigVal   string          // digVal
	DhRecbto time.Time       // dhRecbto
	NProt    string          // nProt
	CSitDFe  pcn.SituacaoDFe // cSitNFe
}

// ResEvento e o resumo de um evento. Porte de TresEvento (:203).
type ResEvento struct {
	COrgao     int            // cOrgao
	CNPJCPF    string         // CNPJ, ou CPF quando aquele falta
	ChDFe      string         // chNFe
	DhEvento   time.Time      // dhEvento
	TpEvento   pcn.TipoEvento // tpEvento
	NSeqEvento int            // nSeqEvento
	XEvento    string         // xEvento
	DhRecbto   time.Time      // dhRecbto
	NProt      string         // nProt
}

// ProcEventoDetEvento e o detEvento de um evento processado que chega pela
// distribuicao. Porte de TprocEvento_DetEvento (:109).
//
// E um subconjunto do DetEvento da etapa de eventos: a distribuicao le so
// os campos comuns, nao os grupos repetidos.
type ProcEventoDetEvento struct {
	Versao     string // atributo versao
	DescEvento string // descEvento
	NProt      string // nProt
	XJust      string // xJust
	XCorrecao  string // xCorrecao
	TpAutor    int    // tpAutor -- lido como INTEIRO aqui, nao como enum
	VerAplic   string // verAplic
}

// ProcEventoRetInfEvento e o retEvento de um evento processado.
// Porte de TprocEvento_RetInfEvento (:141).
type ProcEventoRetInfEvento struct {
	ID          string           // atributo Id
	TpAmb       pcn.TipoAmbiente // tpAmb
	VerAplic    string           // verAplic
	COrgao      int              // cOrgao
	CStat       int              // cStat
	XMotivo     string           // xMotivo
	ChDFe       string           // chNFe
	TpEvento    pcn.TipoEvento   // tpEvento
	XEvento     string           // xEvento
	NSeqEvento  int              // nSeqEvento
	CNPJDest    string           // CNPJDest
	DhRegEvento time.Time        // dhRegEvento
	NProt       string           // nProt
}

// ProcEventoDist e o evento processado que chega pela distribuicao.
// Porte de TprocEvento (:226).
type ProcEventoDist struct {
	ID         string           // atributo Id do infEvento
	COrgao     int              // cOrgao
	TpAmb      pcn.TipoAmbiente // tpAmb
	CNPJ       string           // CNPJ, ou CPF quando aquele falta
	ChDFe      string           // chNFe
	DhEvento   time.Time        // dhEvento
	TpEvento   pcn.TipoEvento   // tpEvento
	NSeqEvento int              // nSeqEvento
	VerEvento  string           // verEvento

	DetEvento    ProcEventoDetEvento    // detEvento
	RetInfEvento ProcEventoRetInfEvento // retEvento/infEvento
}

// DocZip e um item do lote. Porte de TdocZipCollectionItem (:258).
type DocZip struct {
	// NSU e o numero sequencial do documento na base da RFB. E o que o
	// consumidor guarda para a proxima varredura.
	NSU string

	// Schema vem do ATRIBUTO do docZip. O tipo efetivo, porem, e decidido
	// pela raiz do documento descomprimido -- ver lerDocumentoDecodificado.
	Schema pcn.SchemaDFe

	// InfZip e o conteudo COMO VEIO: gzip em base64.
	InfZip string

	// XML e o documento JA DESCOMPRIMIDO. E ele que se arquiva.
	XML string

	// Os tres grupos abaixo sao preenchidos conforme o tipo do documento;
	// num item so, apenas um deles tem conteudo.
	ResDFe     ResDFe
	ResEvento  ResEvento
	ProcEvento ProcEventoDist

	// ErroLeitura guarda o problema do item, quando houve.
	//
	// NAO existe no ACBr, onde um docZip corrompido derruba a leitura do
	// lote inteiro (o LerXml tem um `except` que devolve False e descarta
	// tudo). Aqui o item problematico fica marcado e os demais seguem.
	ErroLeitura error
}

// RetDistDFeInt e o retorno da Distribuicao DF-e.
// Porte de TRetDistDFeInt (:300).
type RetDistDFeInt struct {
	Versao   string           // atributo versao
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	CStat    int              // cStat
	XMotivo  string           // xMotivo
	DhResp   time.Time        // dhResp

	// UltNSU e o ultimo NSU DEVOLVIDO neste lote; MaxNSU e o maior que
	// existe na base da RFB. A varredura termina quando os dois se
	// igualam -- comparar o UltNSU com o que foi PEDIDO nao serve, porque
	// a RFB devolve no maximo cinquenta documentos por chamada.
	UltNSU string
	MaxNSU string

	DocZip []DocZip
}

const (
	// cStatDistribuicaoComDocumento e o 138: "Documento localizado".
	cStatDistribuicaoComDocumento = 138
	// cStatDistribuicaoSemDocumento e o 137: "Nenhum documento localizado".
	cStatDistribuicaoSemDocumento = 137
)

// Atendida informa se a consulta foi processada.
//
// Porte do resultado de TDistribuicaoDFe.TratarResposta
// (ACBrNFeWebServices.pas:4120): 137 OU 138. O 137 -- "nenhum documento
// localizado" -- CONTA como sucesso: a consulta funcionou e a resposta e
// que nao ha nada novo. Numa varredura esse e o caso mais comum, e trata-lo
// como falha faz o laco nunca terminar.
func (r *RetDistDFeInt) Atendida() bool {
	return r != nil &&
		(r.CStat == cStatDistribuicaoComDocumento ||
			r.CStat == cStatDistribuicaoSemDocumento)
}

// VarreduraConcluida informa se nao ha mais documentos a buscar.
//
// A RFB devolve no maximo cinquenta docZip por chamada, entao o fim da
// varredura e UltNSU == MaxNSU -- e nao "o lote veio vazio". Um lote cheio
// com UltNSU abaixo do MaxNSU significa que falta mais.
func (r *RetDistDFeInt) VarreduraConcluida() bool {
	if r == nil {
		return false
	}
	// A comparacao e NUMERICA pelo mesmo motivo do laco da varredura: os
	// dois campos vem preenchidos com quinze zeros, mas nada garante isso
	// num retorno gravado em disco e relido.
	return r.CStat == cStatDistribuicaoSemDocumento ||
		(r.UltNSU != "" && nsuComZeros(r.UltNSU) == nsuComZeros(r.MaxNSU))
}

// LerRetDistDFeInt le o XML de retDistDFeInt, descomprimindo cada docZip.
func LerRetDistDFeInt(xml string) (*RetDistDFeInt, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	node := doc.Root

	r := &RetDistDFeInt{
		Versao:   node.Attr("versao"),
		VerAplic: pcn.ConteudoStr(node.FindAnyNs("verAplic")),
		CStat:    pcn.ConteudoInt(node.FindAnyNs("cStat")),
		XMotivo:  pcn.ConteudoStr(node.FindAnyNs("xMotivo")),
		DhResp:   pcn.ConteudoDataDef(node.FindAnyNs("dhResp")),
		UltNSU:   pcn.ConteudoStr(node.FindAnyNs("ultNSU")),
		MaxNSU:   pcn.ConteudoStr(node.FindAnyNs("maxNSU")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))

	lote := node.FindAnyNs("loteDistDFeInt")
	if lote == nil {
		return r, nil
	}

	for _, z := range lote.FindAllAnyNs("docZip") {
		r.DocZip = append(r.DocZip, lerDocZip(z))
	}
	return r, nil
}

// lerDocZip descomprime um item e le o documento de dentro.
func lerDocZip(z *pcn.Node) DocZip {
	d := DocZip{
		NSU:    z.Attr("NSU"),
		InfZip: strings.TrimSpace(z.Conteudo()),
	}
	// O schema desconhecido NAO interrompe: o tipo efetivo sai da raiz do
	// documento, e o atributo e so uma confirmacao.
	d.Schema, _ = pcn.ParseSchemaDFe(z.Attr("schema"))

	documento, err := dfe.Base64Gunzip(d.InfZip)
	if err != nil {
		d.ErroLeitura = err
		return d
	}
	// ACRESCIMO: o ACBr so atribui o XML DENTRO de cada Ler* (:513, :537,
	// :634, :791), entao raiz desconhecida fica com XML vazio e o
	// TratarResposta pula o item inteiro (:4072). Aqui o XML e guardado
	// sempre -- e o que permite arquivar o documento de um tipo que o
	// porte ainda nao sabe ler.
	d.XML = documento

	lerDocumentoDecodificado(documento, &d)
	return d
}

// lerDocumentoDecodificado despacha pela RAIZ do documento descomprimido.
//
// Porte de LerDocumentoDecodificado (:813-858). O ACBr compara o nome da
// raiz contra quatro literais montados a partir do tipo de DF-e --
// "res"+NFe, "resEvento", minusculo(NFe)+"Proc" e "procEvento"+NFe --, e
// IGNORA o atributo `schema` do docZip nessa decisao.
//
// Documento que nao case com nenhuma das quatro fica com o XML cru e os
// campos vazios, sem erro. E o comportamento do original, e e o certo: a
// RFB acrescenta tipo de documento sem aviso, e perder um item e melhor
// que perder o lote.
func lerDocumentoDecodificado(xml string, d *DocZip) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		d.ErroLeitura = err
		return
	}
	raiz := doc.Root

	switch nomeLocal(raiz) {
	case "resNFe":
		lerResumoNFe(raiz, &d.ResDFe)
	case "resEvento":
		lerResumoEvento(raiz, &d.ResEvento)
	case "nfeProc":
		lerNotaProcessada(raiz, &d.ResDFe)
	case "procEventoNFe":
		lerEventoProcessado(raiz, &d.ProcEvento)
	}
}

// lerResumoNFe le o schema resNFe. Porte de LerResumo (:507).
func lerResumoNFe(node *pcn.Node, r *ResDFe) {
	r.ChDFe = pcn.ConteudoStr(node.FindAnyNs("chNFe"))
	r.CNPJCPF = pcn.ConteudoCNPJCPFAnyNs(node)
	r.XNome = pcn.ConteudoStr(node.FindAnyNs("xNome"))
	r.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))
	r.DhEmi = pcn.ConteudoDataDef(node.FindAnyNs("dhEmi"))
	r.VNF = pcn.ConteudoDe2(node.FindAnyNs("vNF"))
	r.DigVal = pcn.ConteudoStr(node.FindAnyNs("digVal"))
	r.DhRecbto = pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto"))
	r.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))
	r.TpNF, _ = pcn.ParseTipoNFe(pcn.ConteudoStr(node.FindAnyNs("tpNF")))
	// A tag e `cSitNFe` -- o nome do CAMPO e cSitDFe, generico, mas a tag
	// carrega o sufixo do documento.
	r.CSitDFe, _ = pcn.ParseSituacaoDFe(pcn.ConteudoStr(node.FindAnyNs("cSitNFe")))
}

// lerResumoEvento le o schema resEvento. Porte de LerResumoEvento (:531).
func lerResumoEvento(node *pcn.Node, r *ResEvento) {
	r.ChDFe = pcn.ConteudoStr(node.FindAnyNs("chNFe"))
	r.CNPJCPF = pcn.ConteudoCNPJCPFAnyNs(node)
	r.COrgao = pcn.ConteudoInt(node.FindAnyNs("cOrgao"))
	r.DhEvento = pcn.ConteudoDataDef(node.FindAnyNs("dhEvento"))
	r.NSeqEvento = pcn.ConteudoInt(node.FindAnyNs("nSeqEvento"))
	r.XEvento = pcn.ConteudoStr(node.FindAnyNs("xEvento"))
	r.DhRecbto = pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto"))
	r.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))
	r.TpEvento, _ = pcn.ParseTipoEventoNFe(pcn.ConteudoStr(node.FindAnyNs("tpEvento")))
}

// lerNotaProcessada extrai o resumo de dentro de um nfeProc.
// Porte de LerDocumento (:628).
//
// O mesmo struct ResDFe recebe o resumo e o extrato da nota completa --
// mas por caminhos diferentes: aqui os campos vem de quatro grupos
// aninhados, e a chave sai do ATRIBUTO Id do infNFe, pelas 44 posicoes
// finais.
func lerNotaProcessada(node *pcn.Node, r *ResDFe) {
	// Os dois blocos sao INDEPENDENTES, como os dois `if Assigned`
	// sequenciais de :649-671. Fundi-los -- sair da funcao quando o <NFe>
	// falta -- perderia o protocolo inteiro: digVal, dhRecbto, nProt E a
	// situacao. O item voltaria com CSitDFe no zero value, que e
	// SnAutorizado: uma nota cancelada apareceria como autorizada.
	if nota := node.FindAnyNs("NFe"); nota != nil {
		if inf := nota.FindAnyNs("infNFe"); inf != nil {
			r.ChDFe = ultimos44(inf.Attr("Id"))
			lerGrupoIDEDistribuicao(inf.FindAnyNs("ide"), r)
			lerGrupoEmitDistribuicao(inf.FindAnyNs("emit"), r)
			lerGrupoTotalDistribuicao(inf.FindAnyNs("total"), r)
		}
	}

	if prot := node.FindAnyNs("protNFe"); prot != nil {
		lerGrupoInfProtDistribuicao(prot.FindAnyNs("infProt"), r)
	}
}

// ultimos44 devolve as ultimas 44 posicoes, que e o RightStr(Id, 44) do
// original -- a forma de tirar o literal "NFe" do atributo sem supor que
// ele esteja la.
func ultimos44(s string) string {
	if len(s) <= 44 {
		return s
	}
	return s[len(s)-44:]
}

// lerGrupoIDEDistribuicao le o ide. Porte de LerGrupo_ide (:554).
//
// O retrocesso para `dEmi` e o que permite ler uma nota 2.00, onde a data
// de emissao ainda nao tinha hora.
func lerGrupoIDEDistribuicao(node *pcn.Node, r *ResDFe) {
	if node == nil {
		return
	}
	r.DhEmi = pcn.ConteudoDataDef(node.FindAnyNs("dhEmi"))
	if r.DhEmi.IsZero() {
		r.DhEmi = pcn.ConteudoDataDef(node.FindAnyNs("dEmi"))
	}
	r.TpNF, _ = pcn.ParseTipoNFe(pcn.ConteudoStr(node.FindAnyNs("tpNF")))
}

// lerGrupoEmitDistribuicao le o emit. Porte de LerGrupo_emit (:570).
func lerGrupoEmitDistribuicao(node *pcn.Node, r *ResDFe) {
	if node == nil {
		return
	}
	r.CNPJCPF = pcn.ConteudoCNPJCPFAnyNs(node)
	r.XNome = pcn.ConteudoStr(node.FindAnyNs("xNome"))
	r.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))
}

// lerGrupoTotalDistribuicao le o vNF de dentro do ICMSTot.
// Porte de LerGrupo_total (:584).
func lerGrupoTotalDistribuicao(node *pcn.Node, r *ResDFe) {
	if node == nil {
		return
	}
	if tot := node.FindAnyNs("ICMSTot"); tot != nil {
		r.VNF = pcn.ConteudoDe2(tot.FindAnyNs("vNF"))
	}
}

// lerGrupoInfProtDistribuicao le o protocolo e DERIVA a situacao.
// Porte de LerGrupo_infProt (:608).
//
// A situacao nao vem pronta: sai de um `case` sobre o cStat do protocolo, e
// e o unico lugar do porte onde um cStat vira enum de situacao. Codigo fora
// dos quatro listados deixa o campo como estava -- o `case` do Delphi nao
// tem `else`.
func lerGrupoInfProtDistribuicao(node *pcn.Node, r *ResDFe) {
	if node == nil {
		return
	}
	r.DigVal = pcn.ConteudoStr(node.FindAnyNs("digVal"))
	r.DhRecbto = pcn.ConteudoDataDef(node.FindAnyNs("dhRecbto"))
	r.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))

	switch pcn.ConteudoInt(node.FindAnyNs("cStat")) {
	case 100:
		r.CSitDFe = pcn.SnAutorizado
	case 101:
		r.CSitDFe = pcn.SnCancelado
	case 110:
		r.CSitDFe = pcn.SnDenegado
	case 132:
		r.CSitDFe = pcn.SnEncerrado
	}
}

// lerEventoProcessado le um procEventoNFe. Porte de LerEvento (:781) mais
// LerGrupo_detEvento (:734) e LerGrupo_retEvento (:752).
func lerEventoProcessado(node *pcn.Node, p *ProcEventoDist) {
	// O elemento interno e <evento>; o ACBr aceita tambem <eventoNFe>
	// (:786), e a tolerancia nao custa nada.
	evento := node.FindAnyNs("evento")
	if evento == nil {
		evento = node.FindAnyNs("eventoNFe")
	}
	if evento == nil {
		return
	}

	if inf := evento.FindAnyNs("infEvento"); inf != nil {
		p.ID = inf.Attr("Id")
		p.CNPJ = pcn.ConteudoCNPJCPFAnyNs(inf)
		p.ChDFe = pcn.ConteudoStr(inf.FindAnyNs("chNFe"))
		p.DhEvento = pcn.ConteudoDataDef(inf.FindAnyNs("dhEvento"))
		p.NSeqEvento = pcn.ConteudoInt(inf.FindAnyNs("nSeqEvento"))
		p.VerEvento = pcn.ConteudoStr(inf.FindAnyNs("verEvento"))
		p.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
		p.TpEvento, _ = pcn.ParseTipoEventoNFe(pcn.ConteudoStr(inf.FindAnyNs("tpEvento")))

		// O cOrgao NAO e lido pelo ACBr aqui, embora o campo exista no
		// struct (:781-812). Acrescimo: ele esta no XML e a omissao do
		// original parece descuido, nao decisao.
		p.COrgao = pcn.ConteudoInt(inf.FindAnyNs("cOrgao"))

		// ACRESCIMO: o ACBr procura o detEvento como filho DIRETO da raiz
		// procEventoNFe (:850), e o FindAnyNs do TACBrXmlNodeList so
		// percorre filhos diretos. No leiaute o caminho real e
		// procEventoNFe/evento/infEvento/detEvento -- ou seja, o original
		// NUNCA preenche este grupo para a NF-e. O porte o le do nivel
		// certo.
		lerDetEventoDistribuicao(inf.FindAnyNs("detEvento"), &p.DetEvento)
	}

	// E mantem a busca na raiz como retrocesso, que e onde o ACBr a faz --
	// para o dia em que algum envelope traga o grupo nesse nivel.
	if p.DetEvento.DescEvento == "" {
		lerDetEventoDistribuicao(node.FindAnyNs("detEvento"), &p.DetEvento)
	}

	ret := node.FindAnyNs("retEvento")
	if ret == nil {
		ret = node.FindAnyNs("retEventoNFe")
	}
	lerRetInfEventoDistribuicao(ret, &p.RetInfEvento)
}

// lerDetEventoDistribuicao le o detEvento. Porte de LerGrupo_detEvento
// (:734), sem os grupos de CT-e e sem os itens averbados -- os dois
// pertencem ao CT-e, nao a NF-e.
func lerDetEventoDistribuicao(node *pcn.Node, d *ProcEventoDetEvento) {
	if node == nil {
		return
	}
	d.Versao = node.Attr("versao")
	d.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))
	d.XJust = pcn.ConteudoStr(node.FindAnyNs("xJust"))
	d.XCorrecao = pcn.ConteudoStr(node.FindAnyNs("xCorrecao"))
	d.DescEvento = pcn.ConteudoStr(node.FindAnyNs("descEvento"))
	d.TpAutor = pcn.ConteudoInt(node.FindAnyNs("tpAutor"))
	d.VerAplic = pcn.ConteudoStr(node.FindAnyNs("verAplic"))
}

// lerRetInfEventoDistribuicao le o retEvento/infEvento.
// Porte de LerGrupo_retEvento (:752).
func lerRetInfEventoDistribuicao(node *pcn.Node, r *ProcEventoRetInfEvento) {
	if node == nil {
		return
	}
	inf := node.FindAnyNs("infEvento")
	if inf == nil {
		return
	}
	r.ID = inf.Attr("Id")
	r.VerAplic = pcn.ConteudoStr(inf.FindAnyNs("verAplic"))
	r.COrgao = pcn.ConteudoInt(inf.FindAnyNs("cOrgao"))
	r.CStat = pcn.ConteudoInt(inf.FindAnyNs("cStat"))
	r.XMotivo = pcn.ConteudoStr(inf.FindAnyNs("xMotivo"))
	r.ChDFe = pcn.ConteudoStr(inf.FindAnyNs("chNFe"))
	r.XEvento = pcn.ConteudoStr(inf.FindAnyNs("xEvento"))
	r.NSeqEvento = pcn.ConteudoInt(inf.FindAnyNs("nSeqEvento"))
	r.CNPJDest = pcn.ConteudoStr(inf.FindAnyNs("CNPJDest"))
	r.DhRegEvento = pcn.ConteudoDataDef(inf.FindAnyNs("dhRegEvento"))
	r.NProt = pcn.ConteudoStr(inf.FindAnyNs("nProt"))
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
	r.TpEvento, _ = pcn.ParseTipoEventoNFe(pcn.ConteudoStr(inf.FindAnyNs("tpEvento")))
}
