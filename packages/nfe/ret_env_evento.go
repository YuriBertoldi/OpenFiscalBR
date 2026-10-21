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
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Leitor do retorno dos eventos. Porte de TRetEventoNFe
// (ACBrNFe.RetEnvEvento.pas).
//
// Ele atende QUATRO formas de raiz, e nao uma (LerXml, :728-748):
//
//	retEnvEvento    a resposta da SEFAZ ao lote
//	procEventoNFe   o evento processado, como se arquiva
//	evento          o evento solto, antes de transmitir
//	envEvento       o lote enviado
//
// Ler as quatro com o mesmo codigo e o que permite reler do disco o que se
// gravou -- e e por isso que o leitor nao exige a presenca de nenhuma
// delas.
//
// DIVERGENCIA em onze tags: o ACBr usa `Childrens.FindAll(...)` -- que
// compara o nome COM o prefixo de namespace -- para `retEvento`,
// `itemPedido`, `detPag` e os oito grupos da Reforma, e `FindAnyNs` em
// todas as demais. O porte usa `FindAllAnyNs` nas onze.
//
// A troca e uniformemente mais PERMISSIVA: aceita o documento prefixado que
// o ACBr ignoraria, e nunca recusa o que ele aceita. A inconsistencia e do
// original -- nada no leiaute justifica tratar `detPag` diferente de
// `autXML` --, e replicar uma inconsistencia acidental nao protege ninguem.
// Travado em TestLerRetEventoNFe_AceitaDocumentoComPrefixo.

// RetInfEvento e o resultado do processamento de UM evento.
// Porte de TRetInfEvento (ACBrNFe.EventoClass.pas:643).
type RetInfEvento struct {
	ID          string           // atributo Id
	TpAmb       pcn.TipoAmbiente // tpAmb
	VerAplic    string           // verAplic
	COrgao      int              // cOrgao
	CStat       int              // cStat
	XMotivo     string           // xMotivo
	ChNFe       string           // chNFe
	TpEvento    pcn.TipoEvento   // tpEvento
	XEvento     string           // xEvento
	NSeqEvento  int              // nSeqEvento
	CNPJDest    string           // CNPJDest, ou CPFDest quando aquele falta
	EmailDest   string           // emailDest
	COrgaoAutor int              // cOrgaoAutor
	DhRegEvento time.Time        // dhRegEvento
	NProt       string           // nProt

	// ChNFePend sao as chaves de EPEC pendentes de autorizacao. So o
	// retorno do EPEC as traz.
	ChNFePend []string

	// XML e o <infEvento> do retorno COMO VEIO. E dele que sai o
	// procEventoNFe (GerarXMLProcEvento).
	XML string
}

// RetEventoNFe e o retorno do lote de eventos.
type RetEventoNFe struct {
	Versao   string           // atributo versao
	IDLote   int64            // idLote
	TpAmb    pcn.TipoAmbiente // tpAmb
	VerAplic string           // verAplic
	COrgao   int              // cOrgao
	CStat    int              // cStat
	XMotivo  string           // xMotivo

	RetEvento []RetInfEvento // retEvento

	// InfEvento e o evento ENVIADO, preenchido quando a raiz o contem --
	// no procEventoNFe, no envEvento e no evento solto.
	InfEvento *InfEvento
}

// Os cStat que indicam evento REGISTRADO.
//
// Porte da condicao de TACBrNFe.EnviarEvento/Cancelamento
// (ACBrNFeWebServices.pas). Sao tres, e o 136 e o que mais escapa: ele diz
// "evento registrado mas NAO vinculado a NF-e" -- a SEFAZ aceitou o evento
// embora a nota ainda nao exista ali. E sucesso para o registro e nao
// e para quem espera o vinculo.
const (
	cStatEventoRegistrado           = 135
	cStatEventoRegistradoSemVinculo = 136
	cStatEventoRegistradoVinculado  = 155
)

// Registrado informa se o evento foi aceito pela SEFAZ.
func (r *RetInfEvento) Registrado() bool {
	switch r.CStat {
	case cStatEventoRegistrado, cStatEventoRegistradoSemVinculo,
		cStatEventoRegistradoVinculado:
		return true
	}
	return false
}

// VinculadoANota informa se o evento ficou AMARRADO a NF-e.
//
// O 136 fica de fora de proposito: ele e registro sem vinculo. Quem for
// montar o procEventoNFe para arquivar junto da nota precisa desta
// resposta, nao da anterior.
func (r *RetInfEvento) VinculadoANota() bool {
	return r.CStat == cStatEventoRegistrado || r.CStat == cStatEventoRegistradoVinculado
}

// LoteProcessado informa se a SEFAZ processou o lote de eventos.
// O 128 e "Lote de evento processado".
func (r *RetEventoNFe) LoteProcessado() bool {
	return r != nil && r.CStat == cStatLoteDeEventoProcessado
}

// cStatLoteDeEventoProcessado e o 128 do leiaute.
const cStatLoteDeEventoProcessado = 128

// RetornoDaChave devolve o retorno do evento daquela chave, ou nil.
func (r *RetEventoNFe) RetornoDaChave(chave string) *RetInfEvento {
	if r == nil {
		return nil
	}
	chave = pcn.RemoverLiteralChave(chave)
	for i := range r.RetEvento {
		if pcn.RemoverLiteralChave(r.RetEvento[i].ChNFe) == chave {
			return &r.RetEvento[i]
		}
	}
	return nil
}

// LerRetEventoNFe le qualquer uma das formas de raiz.
//
// Duas observacoes sobre o desenho, que diverge do LerXml do ACBr (:728):
//
//   - o original compara o nome da raiz contra quatro literais e sai calado
//     quando nenhum casa. Aqui o caminho comum e GENERICO -- procura um
//     <evento> e, em paralelo, um conjunto de <retEvento> --, porque as duas
//     buscas sao mutuamente inofensivas: um retEnvEvento nao tem <evento>
//     filho, e um <evento> solto nao tem <retEvento>. Isso faz o leitor
//     aceitar envelopes que a SEFAZ acrescente sem que o porte saiba;
//   - o `NFeDFe` continua precisando de ramo proprio, porque nele o
//     procEventoNFe aparece DUAS vezes aninhado (:736-741) e a busca rasa
//     nao alcanca o <evento>.
func LerRetEventoNFe(xml string) (*RetEventoNFe, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return nil, err
	}
	raiz := doc.Root
	r := &RetEventoNFe{Versao: raiz.Attr("versao")}

	if nomeLocal(raiz) == "NFeDFe" {
		if proc := raiz.FindAnyNs("procEventoNFe"); proc != nil {
			if interno := proc.FindAnyNs("procEventoNFe"); interno != nil {
				r.Versao = interno.Attr("versao")
				if ev := interno.FindAnyNs("evento"); ev != nil {
					r.InfEvento = lerInfEvento(ev.FindAnyNs("infEvento"))
				}
			}
		}
		lerRetEvento(raiz, r)
		return r, nil
	}

	if ev := raiz.FindAnyNs("evento"); ev != nil {
		r.InfEvento = lerInfEvento(ev.FindAnyNs("infEvento"))
	} else if nomeLocal(raiz) == "evento" {
		r.InfEvento = lerInfEvento(raiz.FindAnyNs("infEvento"))
	}
	lerRetEvento(raiz, r)

	return r, nil
}

// nomeLocal devolve o nome do elemento sem o prefixo de namespace.
func nomeLocal(n *pcn.Node) string {
	if n == nil {
		return ""
	}
	nome := n.Nome
	if i := strings.LastIndex(nome, ":"); i >= 0 {
		nome = nome[i+1:]
	}
	return nome
}

// lerRetEvento preenche o cabecalho do retorno e a lista de retEvento.
// Porte de Ler_RetEvento (:683).
func lerRetEvento(node *pcn.Node, r *RetEventoNFe) {
	// O idLote tem quinze digitos no leiaute e o ACBr o le com tcInt64
	// (:697). Ler como int estoura em build de 32 bits.
	r.IDLote = pcn.ConteudoInt64(node.FindAnyNs("idLote"))
	r.VerAplic = pcn.ConteudoStr(node.FindAnyNs("verAplic"))
	r.COrgao = pcn.ConteudoInt(node.FindAnyNs("cOrgao"))
	r.CStat = pcn.ConteudoInt(node.FindAnyNs("cStat"))
	r.XMotivo = pcn.ConteudoStr(node.FindAnyNs("xMotivo"))

	// O tpAmb so e atribuido quando VEM preenchido (:694). Num
	// procEventoNFe o retEvento traz o seu e o evento traz o dele; ler sem
	// o guard sobrescreveria com vazio.
	if v := pcn.ConteudoStr(node.FindAnyNs("tpAmb")); v != "" {
		r.TpAmb, _ = pcn.ParseTipoAmbiente(v)
	}

	for _, ret := range node.FindAllAnyNs("retEvento") {
		if inf := ret.FindAnyNs("infEvento"); inf != nil {
			r.RetEvento = append(r.RetEvento, lerRetInfEvento(inf))
		}
	}
}

// lerRetInfEvento le um <infEvento> de dentro de um <retEvento>.
// Porte de Ler_InfEventos (:631).
func lerRetInfEvento(inf *pcn.Node) RetInfEvento {
	r := RetInfEvento{
		XML:         inf.OuterXML(),
		ID:          inf.Attr("Id"),
		VerAplic:    pcn.ConteudoStr(inf.FindAnyNs("verAplic")),
		COrgao:      pcn.ConteudoInt(inf.FindAnyNs("cOrgao")),
		CStat:       pcn.ConteudoInt(inf.FindAnyNs("cStat")),
		XMotivo:     pcn.ConteudoStr(inf.FindAnyNs("xMotivo")),
		ChNFe:       pcn.ConteudoStr(inf.FindAnyNs("chNFe")),
		XEvento:     pcn.ConteudoStr(inf.FindAnyNs("xEvento")),
		NSeqEvento:  pcn.ConteudoInt(inf.FindAnyNs("nSeqEvento")),
		EmailDest:   pcn.ConteudoStr(inf.FindAnyNs("emailDest")),
		COrgaoAutor: pcn.ConteudoInt(inf.FindAnyNs("cOrgaoAutor")),
		DhRegEvento: pcn.ConteudoDataDef(inf.FindAnyNs("dhRegEvento")),
		NProt:       pcn.ConteudoStr(inf.FindAnyNs("nProt")),
	}
	r.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(inf.FindAnyNs("tpAmb")))
	r.TpEvento, _ = pcn.ParseTipoEventoNFe(pcn.ConteudoStr(inf.FindAnyNs("tpEvento")))

	// CNPJDest cai para CPFDest quando aquele falta (:656). O destinatario
	// de uma manifestacao pode ser pessoa fisica, e o campo do retorno muda
	// de nome -- nao de significado.
	r.CNPJDest = pcn.ConteudoStr(inf.FindAnyNs("CNPJDest"))
	if r.CNPJDest == "" {
		r.CNPJDest = pcn.ConteudoStr(inf.FindAnyNs("CPFDest"))
	}

	for _, p := range inf.FindAllAnyNs("chNFePend") {
		if chave := pcn.ConteudoStr(p); chave != "" {
			r.ChNFePend = append(r.ChNFePend, chave)
		}
	}
	return r
}

// lerInfEvento le o <infEvento> do evento ENVIADO.
// Porte de Ler_InfEvento (:601).
func lerInfEvento(node *pcn.Node) *InfEvento {
	if node == nil {
		return nil
	}
	inf := &InfEvento{
		ID:         node.Attr("Id"),
		COrgao:     pcn.ConteudoInt(node.FindAnyNs("cOrgao")),
		CNPJ:       pcn.ConteudoCNPJCPFAnyNs(node),
		ChNFe:      pcn.ConteudoStr(node.FindAnyNs("chNFe")),
		DhEvento:   pcn.ConteudoDataDef(node.FindAnyNs("dhEvento")),
		NSeqEvento: pcn.ConteudoInt(node.FindAnyNs("nSeqEvento")),
	}
	inf.TpAmb, _ = pcn.ParseTipoAmbiente(pcn.ConteudoStr(node.FindAnyNs("tpAmb")))
	inf.TpEvento, _ = pcn.ParseTipoEventoNFe(pcn.ConteudoStr(node.FindAnyNs("tpEvento")))
	inf.VersaoEvento = pcn.ConteudoStr(node.FindAnyNs("verEvento"))
	inf.DetEvento.Versao = detEventoVersao(node)

	lerDetEvento(node.FindAnyNs("detEvento"), inf)
	return inf
}

// lerDetEvento le o detEvento.
// Porte de Ler_DetEvento (:441).
//
// A forma reproduz a do original e a do struct: os campos comuns sao lidos
// SEM olhar o tpEvento -- cada tag que existir vira campo --, e so os
// grupos repetidos passam por um `switch`. Ler tudo incondicionalmente
// funciona porque os nomes de tag nao colidem entre eventos... exceto nos
// grupos, que e exatamente onde o switch entra.
func lerDetEvento(node *pcn.Node, inf *InfEvento) {
	if node == nil {
		return
	}
	d := &inf.DetEvento

	d.DescEvento = pcn.ConteudoStr(node.FindAnyNs("descEvento"))
	d.XCorrecao = pcn.ConteudoStr(node.FindAnyNs("xCorrecao"))
	d.XCondUso = pcn.ConteudoStr(node.FindAnyNs("xCondUso"))
	d.NProt = pcn.ConteudoStr(node.FindAnyNs("nProt"))
	d.XJust = pcn.ConteudoStr(node.FindAnyNs("xJust"))
	d.COrgaoAutor = pcn.ConteudoInt(node.FindAnyNs("cOrgaoAutor"))
	d.VerAplic = pcn.ConteudoStr(node.FindAnyNs("verAplic"))
	d.DhEmi = pcn.ConteudoDataDef(node.FindAnyNs("dhEmi"))
	d.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))
	d.DhEntrega = pcn.ConteudoDataDef(node.FindAnyNs("dhEntrega"))
	d.NDoc = pcn.ConteudoStr(node.FindAnyNs("nDoc"))
	d.XNome = pcn.ConteudoStr(node.FindAnyNs("xNome"))
	d.LatGPS = pcn.ConteudoDe6(node.FindAnyNs("latGPS"))
	d.LongGPS = pcn.ConteudoDe6(node.FindAnyNs("longGPS"))
	d.HashComprovante = pcn.ConteudoStr(node.FindAnyNs("hashComprovante"))
	d.DhHashComprovante = pcn.ConteudoDataDef(node.FindAnyNs("dhHashComprovante"))
	d.NProtEvento = pcn.ConteudoStr(node.FindAnyNs("nProtEvento"))
	d.IDPedidoCancelado = pcn.ConteudoStr(node.FindAnyNs("idPedidoCancelado"))
	d.DhTentativaEntrega = pcn.ConteudoDataDef(node.FindAnyNs("dhTentativaEntrega"))
	d.NTentativa = pcn.ConteudoInt(node.FindAnyNs("nTentativa"))
	d.XJustMotivo = pcn.ConteudoStr(node.FindAnyNs("xJustMotivo"))
	d.HashTentativaEntrega = pcn.ConteudoStr(node.FindAnyNs("hashTentativaEntrega"))
	d.DhHashTentativaEntrega = pcn.ConteudoDataDef(node.FindAnyNs("dhHashTentativaEntrega"))

	// No EPEC de NF-e os tres valores vivem DENTRO do dest; no de NFC-e,
	// fora dele e sem o vST (Gerar_Evento_EPEC, :650-660). Le-los aqui e um
	// ACRESCIMO ao ACBr, que so os le dentro do Ler_Dest (:195) e por isso
	// perde os dois da NFC-e no round-trip. O lerDestEvento sobrescreve
	// depois quando o grupo existir, entao a precedencia continua sendo a
	// do dest.
	d.VNF = pcn.ConteudoDe2(node.FindAnyNs("vNF"))
	d.VICMS = pcn.ConteudoDe2(node.FindAnyNs("vICMS"))
	d.VST = pcn.ConteudoDe2(node.FindAnyNs("vST"))

	// Os tres enums abaixo so sao atribuidos quando a tag VEM preenchida
	// (:458, :465, :486). Sem o guard, o zero value do enum passaria a
	// significar "tpAutor = 1" num evento que nao tem tpAutor nenhum.
	if v := pcn.ConteudoStr(node.FindAnyNs("tpAutor")); v != "" {
		d.TpAutor, _ = ParseTipoAutor(v)
	}
	if v := pcn.ConteudoStr(node.FindAnyNs("tpNF")); v != "" {
		d.TpNF, _ = pcn.ParseTipoNFe(v)
	}
	if v := pcn.ConteudoStr(node.FindAnyNs("tpAutorizacao")); v != "" {
		d.TpAutorizacao, _ = ParseAutorizacao(v)
	}
	// O tpMotivo e lido SEM guard no ACBr (:492) -- replicado.
	d.TpMotivo, _ = ParseTpMotivo(pcn.ConteudoStr(node.FindAnyNs("tpMotivo")))

	lerDestEvento(node.FindAnyNs("dest"), d)

	for _, a := range node.FindAllAnyNs("autXML") {
		d.AutXML = append(d.AutXML, AutXMLEvento{
			CNPJCPF: pcn.ConteudoCNPJCPFAnyNs(a),
		})
	}

	for _, it := range node.FindAllAnyNs("itemPedido") {
		d.ItemPedido = append(d.ItemPedido, ItemPedido{
			// O numItem e ATRIBUTO na geracao (:1100) e o ACBr o le como
			// TAG (:679). A leitura cobre as duas formas; emitir so a tag
			// faria o round-trip passar e a SEFAZ rejeitar.
			NumItem:  nItemDoGrupoPedido(it),
			QtdeItem: pcn.ConteudoDe4(it.FindAnyNs("qtdeItem")),
		})
	}

	for _, p := range node.FindAllAnyNs("detPag") {
		d.DetPag = append(d.DetPag, lerDetPagEvento(p))
	}

	lerDetEventoGrupos(node, inf)
}

// nItemDoGrupoPedido le o numItem do itemPedido.
//
// O ACBr GRAVA numItem como ATRIBUTO (:1100) e o LE como TAG (:679) -- o
// round-trip dele perde o campo. Aqui o atributo vence e a tag e o
// retrocesso, para que o leitor aceite tanto o que este porte gera quanto
// o que o ACBr gerou um dia.
func nItemDoGrupoPedido(n *pcn.Node) int {
	if v := n.Attr("numItem"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return pcn.ConteudoInt(n.FindAnyNs("numItem"))
}

// lerDestEvento le o grupo dest do EPEC. Porte de Ler_Dest (:195).
func lerDestEvento(node *pcn.Node, d *DetEvento) {
	if node == nil {
		return
	}
	d.Dest.UF = pcn.ConteudoStr(node.FindAnyNs("UF"))
	d.Dest.CNPJCPF = pcn.ConteudoCNPJCPFAnyNs(node)
	d.Dest.IDEstrangeiro = pcn.ConteudoStr(node.FindAnyNs("idEstrangeiro"))
	d.Dest.IE = pcn.ConteudoStr(node.FindAnyNs("IE"))

	// Na NF-e os tres valores vivem DENTRO do dest (Gerar_DestNFe, :372);
	// na NFC-e, fora dele. O lerDetEvento ja os leu do nivel de cima, e
	// estas atribuicoes SOBRESCREVEM -- entao o dest vence quando existe,
	// que e o que o ACBr faz.
	d.VNF = pcn.ConteudoDe2(node.FindAnyNs("vNF"))
	d.VICMS = pcn.ConteudoDe2(node.FindAnyNs("vICMS"))
	d.VST = pcn.ConteudoDe2(node.FindAnyNs("vST"))
}

// lerDetPagEvento le um <detPag> da conciliacao financeira.
// Porte de Ler_detPag (:221).
func lerDetPagEvento(node *pcn.Node) DetPagEvento {
	p := DetPagEvento{
		XPag:      pcn.ConteudoStr(node.FindAnyNs("xPag")),
		VPag:      pcn.ConteudoDe2(node.FindAnyNs("vPag")),
		DPag:      pcn.ConteudoDataDef(node.FindAnyNs("dPag")),
		CNPJPag:   pcn.ConteudoStr(node.FindAnyNs("CNPJPag")),
		UFPag:     pcn.ConteudoStr(node.FindAnyNs("UFPag")),
		CNPJIF:    pcn.ConteudoStr(node.FindAnyNs("CNPJIF")),
		CAut:      pcn.ConteudoStr(node.FindAnyNs("cAut")),
		CNPJReceb: pcn.ConteudoStr(node.FindAnyNs("CNPJReceb")),
		UFReceb:   pcn.ConteudoStr(node.FindAnyNs("UFReceb")),
	}
	p.IndPag, _ = ParseIndicadorPagamento(pcn.ConteudoStr(node.FindAnyNs("indPag")))
	p.TPag, _ = ParseFormaPagamento(pcn.ConteudoStr(node.FindAnyNs("tPag")))
	p.TBand, _ = ParseBandeiraCartao(pcn.ConteudoStr(node.FindAnyNs("tBand")))
	return p
}

// detEventoVersao le o atributo `versao` do proprio detEvento.
//
// E um campo DIFERENTE do <verEvento> do infEvento, embora na pratica os
// dois carreguem o mesmo valor. O ACBr nem le este; guardar os dois separa
// o que o leiaute separa.
func detEventoVersao(infEvento *pcn.Node) string {
	if det := infEvento.FindAnyNs("detEvento"); det != nil {
		return det.Attr("versao")
	}
	return ""
}
