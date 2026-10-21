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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Gerador do envEvento. Porte de TEventoNFe (ACBrNFe.EnvEvento.pas).
//
// A estrutura externa e fixa e pequena; o que varia e o <detEvento>, e e
// ali que mora todo o risco do porte:
//
//   - a sequencia de campos do detEvento MUDA por tipo de evento, e o XSD e
//     xs:sequence. Nenhum round-trip pega ordem trocada;
//   - varios eventos emitem `tpAutor` com o valor FIXO "1", ignorando o
//     campo do documento, porque o schema so aceita esse valor. Outros
//     emitem o campo. Ler o Delphi e a unica forma de saber qual e qual;
//   - as datas de dentro do detEvento usam fusos de ORIGENS diferentes: o
//     comprovante de entrega resolve pelo cOrgaoAutor, o insucesso na
//     entrega pela UF do emitente.

// GerarXMLEvento monta o envEvento com o lote de eventos.
//
// MUTA os eventos, como o original: grava infEvento.ID em cada um
// (Gerar_Evento, :284).
//
// Evento que ja tenha XMLAssinado entra COMO ESTA, recortado do <evento>
// ao </evento>. Regera-lo invalidaria o digest, e e por isso que esta e a
// unica funcao que monta o envelope -- o transmissor assina evento a
// evento e chama esta mesma funcao, em vez de remontar o envelope por
// fora.
func GerarXMLEvento(e *EventoNFe) (string, error) {
	if e == nil || len(e.Evento) == 0 {
		return "", ErrLoteVazio
	}
	if len(e.Evento) > maximoEventosPorLote {
		return "", fmt.Errorf("%w: %d eventos no lote, maximo %d",
			ErrLoteExcedido, len(e.Evento), maximoEventosPorLote)
	}

	var b strings.Builder
	b.WriteString(`<envEvento xmlns="` + namespaceNFe + `" versao="` +
		pcn.EscaparAtributoXML(e.Versao) + `">`)
	b.WriteString(pcn.NodeInt64("idLote", e.IDLote, 1, true).XML())

	for i := range e.Evento {
		if assinado := e.Evento[i].XMLAssinado; assinado != "" {
			b.WriteString(recortarEvento(assinado))
			continue
		}
		no, err := gerarEvento(&e.Evento[i], e.Versao)
		if err != nil {
			return "", err
		}
		b.WriteString(no.XML())
	}
	b.WriteString("</envEvento>")
	return b.String(), nil
}

// gerarEvento monta um <evento>, montando antes o Id.
func gerarEvento(ev *Evento, versao string) (*pcn.Elem, error) {
	ev.InfEvento.ID = IDEvento(ev.InfEvento)

	inf, err := gerarInfEvento(ev.InfEvento, versao)
	if err != nil {
		return nil, err
	}

	no := pcn.NovoElem("evento").
		Attr("xmlns", namespaceNFe).
		Attr("versao", versao).
		Filho(inf)

	// A assinatura so entra quando esta completa, como na nota.
	if ev.Signature.Assinada() {
		ev.Signature.URI = "#" + ev.InfEvento.ID
		no.Filho(gerarSignature(&ev.Signature))
	}
	return no, nil
}

// IDEvento monta o identificador do evento.
//
// Porte de Gerar_Evento (:284): "ID" + tpEvento(6) + chave(44) +
// nSeqEvento com DUAS posicoes -- cinquenta e quatro caracteres.
//
// O nSeqEvento vai com `%.2d`, nao com `%d`: a sequencia 1 sai "01". Um
// porte que a escreva sem preenchimento produz um Id de 53 posicoes, e e o
// proprio ACBr que avisa disso (:288).
func IDEvento(inf InfEvento) string {
	return "ID" + inf.TpEvento.Codigo() +
		pcn.RemoverLiteralChave(inf.ChNFe) +
		fmt.Sprintf("%02d", inf.NSeqEvento)
}

// COrgaoDoEvento devolve o orgao autorizador efetivo.
//
// Porte de TInfEvento.getcOrgao (:705): zero cai para o cUF da CHAVE. E o
// que faz o evento ir para a mesma SEFAZ da nota sem que o consumidor
// precise informar nada -- e tambem o que torna a manifestacao do
// destinatario um caso especial, porque ela precisa do 91 explicito.
func COrgaoDoEvento(inf InfEvento) int {
	if inf.COrgao != 0 {
		return inf.COrgao
	}
	chave := pcn.RemoverLiteralChave(inf.ChNFe)
	if len(chave) < 2 {
		return 0
	}
	cUF, err := strconv.Atoi(chave[:2])
	if err != nil {
		return 0
	}
	return cUF
}

// gerarInfEvento monta o grupo infEvento.
// Porte de Gerar_InfEvento (:942).
func gerarInfEvento(inf InfEvento, versao string) (*pcn.Elem, error) {
	cOrgao := COrgaoDoEvento(inf)

	no := pcn.NovoElem("infEvento").
		Attr("Id", inf.ID).
		Filho(pcn.NodeInt("cOrgao", cOrgao, 1, true)).
		Filho(pcn.NodeStr("tpAmb", inf.TpAmb.String(), true)).
		Filho(documentoDoAutor(inf)).
		Filho(pcn.NodeStr("chNFe", pcn.RemoverLiteralChave(inf.ChNFe), true)).
		Filho(pcn.NodeStr("dhEvento",
			pcn.FormatarDataHoraXML(inf.DhEvento, pcn.SiglaUF(cOrgao)), true)).
		Filho(pcn.NodeStr("tpEvento", inf.TpEvento.Codigo(), true)).
		Filho(pcn.NodeInt("nSeqEvento", inf.NSeqEvento, 1, true)).
		Filho(pcn.NodeStr("verEvento", versao, true))

	det, err := gerarDetEvento(inf, versao)
	if err != nil {
		return nil, err
	}
	no.Filho(det)
	return no, nil
}

// serieDePessoaFisicaInicio e serieDePessoaFisicaFim delimitam a faixa de
// serie reservada a emitente pessoa fisica (:988).
const (
	serieDePessoaFisicaInicio = 910
	serieDePessoaFisicaFim    = 969
)

// documentoDoAutor escolhe entre <CNPJ> e <CPF> para o autor do evento.
//
// Porte de :982-1013, que tem tres decisoes empilhadas:
//
//  1. CNPJ vazio cai para o documento extraido da CHAVE;
//  2. um documento de catorze digitos numa nota de SERIE 910..969 e na
//     verdade um CPF com tres zeros na frente, porque a chave reserva
//     catorze posicoes -- entao os tres primeiros sao descartados;
//  3. a regra 2 NAO se aplica a manifestacao do destinatario, porque ali o
//     autor e o destinatario e nao o emitente da chave.
//
// Perder a terceira condicional faz toda manifestacao sobre nota de pessoa
// fisica sair com o CNPJ do destinatario truncado.
func documentoDoAutor(inf InfEvento) *pcn.Elem {
	chave := pcn.RemoverLiteralChave(inf.ChNFe)

	doc := pcn.OnlyCPFCNPJAlphaNum(inf.CNPJ)
	if doc == "" {
		doc = documentoDaChave(chave)
	}

	if len(doc) == 14 && !ehManifestacaoDestinatario(inf.TpEvento) {
		if serie := serieDaChave(chave); serie >= serieDePessoaFisicaInicio &&
			serie <= serieDePessoaFisicaFim {
			doc = doc[3:]
		}
	}

	if len(doc) == 14 {
		return pcn.NodeStr("CNPJ", doc, true)
	}
	return pcn.NodeStr("CPF", doc, true)
}

// ehManifestacaoDestinatario reproduz o intervalo de enum
// `teManifDestConfirmacao..teManifDestOperNaoRealizada` de :990.
//
// No Delphi isso e um range do enum, e os quatro membros sao contiguos por
// acaso da ordem de declaracao. Aqui a lista e explicita: um membro novo
// inserido no meio do catalogo mudaria o range em silencio.
func ehManifestacaoDestinatario(t pcn.TipoEvento) bool {
	switch t {
	case pcn.TeManifDestConfirmacao, pcn.TeManifDestCiencia,
		pcn.TeManifDestDesconhecimento, pcn.TeManifDestOperNaoRealizada:
		return true
	}
	return false
}

// documentoDaChave devolve as catorze posicoes do CNPJ/CPF da chave.
func documentoDaChave(chave string) string {
	if len(chave) < 20 {
		return ""
	}
	return chave[6:20]
}

// serieDaChave devolve a serie gravada na chave.
func serieDaChave(chave string) int {
	if len(chave) < 25 {
		return 0
	}
	serie, err := strconv.Atoi(chave[22:25])
	if err != nil {
		return 0
	}
	return serie
}

// gerarDetEvento despacha para o gerador do tipo de evento.
// Porte do `case` de :1021-1089.
//
// Um tipo sem gerador NAO produz detEvento no ACBr -- o case simplesmente
// nao casa e o infEvento sai sem o grupo. Aqui o caso vira erro: um evento
// sem detEvento e rejeicao certa, e falhar cedo diz ao consumidor o que
// faltou.
func gerarDetEvento(inf InfEvento, versao string) (*pcn.Elem, error) {
	d := inf.DetEvento
	desc := descricaoDoEvento(inf)

	// O descEvento de dezesseis tipos sai COM acentos. Ver nodeDescEvento.
	acentos := EventoLevaAcentos(inf.TpEvento)

	det := pcn.NovoElem("detEvento").Attr("versao", versao)

	switch inf.TpEvento {
	case pcn.TeCCe:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("xCorrecao", d.XCorrecao, true)).
			Filho(pcn.NodeStr("xCondUso", condicaoDeUsoDaCCe(d.XCondUso), true))

	case pcn.TeCancelamento:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("nProt", d.NProt, true)).
			Filho(pcn.NodeStr("xJust", d.XJust, true))

	case pcn.TeCancSubst:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
			Filho(pcn.NodeStr("tpAutor", tpAutorFixoDoSchema, true)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
			Filho(pcn.NodeStr("nProt", d.NProt, true)).
			Filho(pcn.NodeStr("xJust", d.XJust, true)).
			Filho(pcn.NodeStr("chNFeRef", pcn.RemoverLiteralChave(d.ChNFeRef), true))

	case pcn.TeManifDestConfirmacao, pcn.TeManifDestCiencia:
		// Os dois SO tem descEvento. Acrescentar xJust aqui -- a tentacao
		// de uniformizar as quatro manifestacoes -- e rejeicao de schema.
		det.Filho(nodeDescEvento(desc, acentos))

	case pcn.TeManifDestDesconhecimento:
		// xJust OPCIONAL (:721, obrigatoriedade 0).
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("xJust", d.XJust, false))

	case pcn.TeManifDestOperNaoRealizada:
		// xJust OBRIGATORIO (:733, obrigatoriedade 1). A diferenca entre
		// este e o anterior e de UM parametro no Delphi.
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("xJust", d.XJust, true))

	case pcn.TeEPECNFe:
		gerarDetEventoEPEC(det, inf, desc, acentos)

	case pcn.TePedProrrog1, pcn.TePedProrrog2:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("nProt", d.NProt, true))
		for _, item := range d.ItemPedido {
			det.Filho(pcn.NovoElem("itemPedido").
				Attr("numItem", strconv.Itoa(item.NumItem)).
				Filho(pcn.NodeDec("qtdeItem", item.QtdeItem, 2, true)))
		}

	case pcn.TeCanPedProrrog1, pcn.TeCanPedProrrog2:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("idPedidoCancelado", d.IDPedidoCancelado, true)).
			Filho(pcn.NodeStr("nProt", d.NProt, true))

	case pcn.TeComprEntregaNFe:
		gerarDetEventoComprovanteEntrega(det, d, desc, acentos)

	case pcn.TeCancComprEntregaNFe:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
			Filho(pcn.NodeStr("tpAutor", tpAutorFixoDoSchema, true)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
			Filho(pcn.NodeStr("nProtEvento", d.NProtEvento, true))

	case pcn.TeAtorInteressadoNFe:
		gerarDetEventoAtorInteressado(det, d, desc, acentos)

	case pcn.TeInsucessoEntregaNFe:
		gerarDetEventoInsucessoEntrega(det, d, desc, acentos)

	case pcn.TeCancInsucessoEntregaNFe:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
			Filho(pcn.NodeStr("nProtEvento", d.NProtEvento, true))

	case pcn.TeConcFinanceira:
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true))
		for _, p := range d.DetPag {
			det.Filho(gerarDetPagEvento(p))
		}

	case pcn.TeCancConcFinanceira:
		// Sem cOrgaoAutor e sem tpAutor, ao contrario dos outros
		// cancelamentos de evento (:926).
		det.Filho(nodeDescEvento(desc, acentos)).
			Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
			Filho(pcn.NodeStr("nProtEvento", d.NProtEvento, true))

	default:
		if gerarDetEventoReformaTributaria(det, inf, desc, acentos) {
			return det, nil
		}
		return nil, fmt.Errorf("%w: evento %s (%s) nao tem gerador de detEvento",
			ErrEventoNaoSuportado, inf.TpEvento.Codigo(), inf.TpEvento)
	}

	return det, nil
}

// tpAutorFixoDoSchema e o "1" que QUATRO eventos emitem IGNORANDO o campo
// tpAutor do documento: cancelamento por substituicao, comprovante de
// entrega, seu cancelamento e o EPEC.
//
// O ACBr comenta o porque em cada um (:490, :541, :586, :633): "Conforme o
// Schema o unico valor aceito e 1". Os demais eventos passam pelo
// TipoAutorToStr e respeitam o campo. Uniformizar os dois grupos e o erro
// natural, e ele so aparece numa rejeicao da SEFAZ.
const tpAutorFixoDoSchema = "1"

// condicaoDeUsoCCe e o texto legal obrigatorio da carta de correcao.
//
// O ACBr o grava num setter (setxCondUso, ACBrNFe.EventoClass.pas), que
// SOBRESCREVE o que o consumidor tenha posto. Aqui a substituicao acontece
// na geracao, e so quando o campo vem vazio -- ver condicaoDeUsoDaCCe.
const condicaoDeUsoCCe = "A Carta de Correcao e disciplinada pelo paragrafo " +
	"1o-A do art. 7o do Convenio S/N, de 15 de dezembro de 1970 e pode ser " +
	"utilizada para regularizacao de erro ocorrido na emissao de documento " +
	"fiscal, desde que o erro nao esteja relacionado com: I - as variaveis " +
	"que determinam o valor do imposto tais como: base de calculo, aliquota, " +
	"diferenca de preco, quantidade, valor da operacao ou da prestacao; II - " +
	"a correcao de dados cadastrais que implique mudanca do remetente ou do " +
	"destinatario; III - a data de emissao ou de saida."

// condicaoDeUsoDaCCe devolve o texto do xCondUso.
//
// Vazio recebe o texto legal. Preenchido, o do consumidor PREVALECE -- uma
// divergencia deliberada: o setter do ACBr sobrescreve sempre, e isso
// impediria corrigir o texto se a legislacao mudar sem que o componente
// seja atualizado.
func condicaoDeUsoDaCCe(informado string) string {
	if strings.TrimSpace(informado) == "" {
		return condicaoDeUsoCCe
	}
	return informado
}

// gerarDetEventoEPEC monta o detEvento do EPEC.
// Porte de Gerar_Evento_EPEC (:617).
//
// O corte decisivo e o MODELO, que sai da chave e nao da configuracao: na
// NF-e o grupo dest carrega vNF, vICMS e vST; na NFC-e o dest e opcional,
// vNF e vICMS ficam FORA dele e o vST nao existe (:650-660).
func gerarDetEventoEPEC(det *pcn.Elem, inf InfEvento, desc string, acentos bool) {
	d := inf.DetEvento
	chave := pcn.RemoverLiteralChave(inf.ChNFe)

	det.Filho(nodeDescEvento(desc, acentos)).
		Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
		Filho(pcn.NodeStr("tpAutor", tpAutorFixoDoSchema, true)).
		Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
		Filho(pcn.NodeStr("dhEmi",
			pcn.FormatarDataHoraXML(d.DhEmi, pcn.SiglaUF(d.COrgaoAutor)), true)).
		Filho(pcn.NodeStr("tpNF", d.TpNF.String(), true)).
		Filho(pcn.NodeStr("IE", d.IE, true))

	// O teste e pelo "55", como no original (:657) -- e nao pelo "65". A
	// diferenca so aparece com chave curta ou de modelo desconhecido, e nos
	// dois casos o ACBr cai no ramo da NFC-e. Inverter o literal parece
	// equivalente e nao e.
	if modeloDaChaveCrua(chave) == "55" {
		det.Filho(gerarDestEventoNFe(d))
		return
	}

	if dest := gerarDestEventoNFCe(d.Dest); dest != nil {
		det.Filho(dest)
	}
	det.Filho(pcn.NodeDec("vNF", d.VNF, 2, true)).
		Filho(pcn.NodeDec("vICMS", d.VICMS, 2, true))
}

// modeloDaChaveCrua devolve as duas posicoes do modelo na chave ja limpa.
func modeloDaChaveCrua(chave string) string {
	if len(chave) < 22 {
		return ""
	}
	return chave[20:22]
}

// gerarDestEventoNFe monta o <dest> do EPEC de NF-e.
// Porte de Gerar_DestNFe (:372).
func gerarDestEventoNFe(d DetEvento) *pcn.Elem {
	dest := pcn.NovoElem("dest").
		Filho(pcn.NodeStr("UF", d.Dest.UF, true)).
		Filho(documentoDoDestinatario(d.Dest))

	// IE opcional aqui; no dest da NFC-e ela nem existe.
	dest.Filho(pcn.NodeStr("IE", d.Dest.IE, false)).
		Filho(pcn.NodeDec("vNF", d.VNF, 2, true)).
		Filho(pcn.NodeDec("vICMS", d.VICMS, 2, true)).
		Filho(pcn.NodeDec("vST", d.VST, 2, true))
	return dest
}

// gerarDestEventoNFCe monta o <dest> do EPEC de NFC-e, que e OPCIONAL.
// Porte de Gerar_DestNFCe (:326).
//
// O grupo inteiro so sai quando ha UF E (documento OU idEstrangeiro) --
// condicao composta que, perdida, produz um <dest> vazio no lugar de
// nenhum.
func gerarDestEventoNFCe(dest Destinatario) *pcn.Elem {
	if dest.UF == "" || (dest.CNPJCPF == "" && dest.IDEstrangeiro == "") {
		return nil
	}
	return pcn.NovoElem("dest").
		Filho(pcn.NodeStr("UF", dest.UF, true)).
		Filho(documentoDoDestinatario(dest))
}

// documentoDoDestinatario escolhe entre CNPJ, CPF e idEstrangeiro.
//
// A condicao e a mesma nos dois grupos dest (:342 e :386): o estrangeiro
// vale quando ha idEstrangeiro OU quando a UF e "EX".
func documentoDoDestinatario(dest Destinatario) *pcn.Elem {
	if dest.IDEstrangeiro != "" || dest.UF == "EX" {
		return pcn.NodeStr("idEstrangeiro", dest.IDEstrangeiro, true)
	}
	doc := pcn.OnlyCPFCNPJAlphaNum(dest.CNPJCPF)
	if len(doc) == 14 {
		return pcn.NodeStr("CNPJ", doc, true)
	}
	return pcn.NodeStr("CPF", doc, true)
}

// gerarDetEventoComprovanteEntrega monta o detEvento do 110130.
// Porte de Gerar_Evento_ComprEntrega (:574).
//
// As duas datas resolvem o fuso pelo cOrgaoAutor, nao pela UF do emitente
// -- diferente do insucesso na entrega, que usa a UF (:794).
func gerarDetEventoComprovanteEntrega(det *pcn.Elem, d DetEvento, desc string, acentos bool) {
	uf := pcn.SiglaUF(d.COrgaoAutor)

	det.Filho(nodeDescEvento(desc, acentos)).
		Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
		Filho(pcn.NodeStr("tpAutor", tpAutorFixoDoSchema, true)).
		Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
		Filho(pcn.NodeStr("dhEntrega",
			pcn.FormatarDataHoraXML(d.DhEntrega, uf), true)).
		Filho(pcn.NodeStr("nDoc", d.NDoc, true)).
		Filho(pcn.NodeStr("xNome", d.XNome, true)).
		Filho(pcn.NodeDec("latGPS", d.LatGPS, 6, false)).
		Filho(pcn.NodeDec("longGPS", d.LongGPS, 6, false)).
		Filho(pcn.NodeStr("hashComprovante", d.HashComprovante, true)).
		Filho(pcn.NodeStr("dhHashComprovante",
			pcn.FormatarDataHoraXML(d.DhHashComprovante, uf), true))
}

// gerarDetEventoAtorInteressado monta o detEvento do 110150.
// Porte de Gerar_Evento_AtorInteressado (:426).
//
// Este evento emite tpAutor do CAMPO, nao o "1" fixo -- e um dos que
// respeitam o documento.
func gerarDetEventoAtorInteressado(det *pcn.Elem, d DetEvento, desc string, acentos bool) {
	det.Filho(nodeDescEvento(desc, acentos)).
		Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
		Filho(pcn.NodeStr("tpAutor", d.TpAutor.String(), true)).
		Filho(pcn.NodeStr("verAplic", d.VerAplic, true))

	// O schema do 110150 declara autXML com maxOccurs=1, e o ACBr alerta
	// quando ha mais de um (:322) -- mas gera todos, e a SEFAZ rejeita.
	// Aqui so o PRIMEIRO sai: emitir um documento que se sabe invalido nao
	// ajuda ninguem, e o alerta do ACBr e silencioso na pratica.
	if len(d.AutXML) > 0 {
		det.Filho(pcn.NovoElem("autXML").
			Filho(gerarCNPJCPF(d.AutXML[0].CNPJCPF, true, false)))
	}

	det.Filho(pcn.NodeStr("tpAutorizacao", d.TpAutorizacao.String(), false))

	// O xCondUso so sai quando a autorizacao PERMITE (:456). Emiti-lo
	// sempre declara permissao numa revogacao.
	if d.TpAutorizacao == TaPermite {
		det.Filho(pcn.NodeStr("xCondUso", condicaoDeUsoAtorInteressado, true))
	}
}

// condicaoDeUsoAtorInteressado e o texto fixo do evento 110150 (:457).
const condicaoDeUsoAtorInteressado = "O emitente ou destinatario da NF-e, " +
	"declara que permite o transportador declarado no campo CNPJ/CPF deste " +
	"evento a autorizar os transportadores subcontratados ou redespachados " +
	"a terem acesso ao download da NF-e"

// gerarDetEventoInsucessoEntrega monta o detEvento do 110192.
// Porte de Gerar_Evento_InsucessoEntrega (:778).
//
// Duas particularidades: as datas resolvem o fuso pela UF do campo
// detEvento.UF, e o xJustMotivo so sai quando o motivo e "outro" (:803).
func gerarDetEventoInsucessoEntrega(det *pcn.Elem, d DetEvento, desc string, acentos bool) {
	det.Filho(nodeDescEvento(desc, acentos)).
		Filho(pcn.NodeInt("cOrgaoAutor", d.COrgaoAutor, 1, true)).
		Filho(pcn.NodeStr("verAplic", d.VerAplic, true)).
		Filho(pcn.NodeStr("dhTentativaEntrega",
			pcn.FormatarDataHoraXML(d.DhTentativaEntrega, d.UF), true)).
		Filho(pcn.NodeInt("nTentativa", d.NTentativa, 3, false)).
		Filho(pcn.NodeStr("tpMotivo", d.TpMotivo.String(), true))

	if d.TpMotivo == TmOutro {
		det.Filho(pcn.NodeStr("xJustMotivo", d.XJustMotivo, false))
	}

	det.Filho(pcn.NodeDec("latGPS", d.LatGPS, 6, false)).
		Filho(pcn.NodeDec("longGPS", d.LongGPS, 6, false)).
		Filho(pcn.NodeStr("hashTentativaEntrega", d.HashTentativaEntrega, true)).
		Filho(pcn.NodeStr("dhHashTentativaEntrega",
			pcn.FormatarDataHoraXML(d.DhHashTentativaEntrega, d.UF), false))
}

// gerarDetPagEvento monta um <detPag> da conciliacao financeira.
// Porte de Gerar_DetalhePagamento (:866).
//
// Os dois blocos finais sao CONDICIONAIS e independentes: o do pagador sai
// quando ha CNPJPag ou UFPag; o do recebedor, quando ha CNPJReceb ou
// UFReceb. Emitir os dois sempre produz tags vazias que o schema recusa.
func gerarDetPagEvento(p DetPagEvento) *pcn.Elem {
	no := pcn.NovoElem("detPag").
		Filho(pcn.NodeStr("indPag", p.IndPag.String(), false)).
		Filho(pcn.NodeStr("tPag", p.TPag.String(), true)).
		Filho(pcn.NodeStr("xPag", p.XPag, false)).
		Filho(pcn.NodeDec("vPag", p.VPag, 2, true)).
		Filho(pcn.NodeDat("dPag", p.DPag, true))

	if p.CNPJPag != "" || p.UFPag != "" {
		no.Filho(pcn.NodeStr("CNPJPag", p.CNPJPag, true)).
			Filho(pcn.NodeStr("UFPag", p.UFPag, true)).
			Filho(pcn.NodeStr("CNPJIF", p.CNPJIF, false)).
			Filho(pcn.NodeStr("tBand", p.TBand.String(), false)).
			Filho(pcn.NodeStr("cAut", p.CAut, false))
	}

	if p.CNPJReceb != "" || p.UFReceb != "" {
		no.Filho(pcn.NodeStr("CNPJReceb", p.CNPJReceb, true)).
			Filho(pcn.NodeStr("UFReceb", p.UFReceb, true))
	}
	return no
}

// nodeDescEvento monta o <descEvento> escolhendo o filtro de texto.
//
// Porte do `FOpcoes.RetirarAcentos := False` de Gerar_InfEvento (:947-957),
// que o ACBr liga para dezesseis tipos com o comentario "Nao funciona sem
// acentos". O filtro padrao do TACBrXmlWriter -- e o do pcn.NodeStr --
// translitera acentos para ASCII, e o texto resultante nao e o que a SEFAZ
// espera nesses eventos.
//
// Uma DIVERGENCIA deliberada: no ACBr a flag e GLOBAL e nao e restaurada
// depois de ligada, entao o primeiro evento acentuado de um lote contamina
// todos os seguintes. Isso depende da ordem de iteracao, e o CLAUDE.md manda
// nao replicar esse tipo de comportamento -- aqui a decisao e por EVENTO.
func nodeDescEvento(desc string, comAcentos bool) *pcn.Elem {
	if comAcentos {
		return pcn.NodeStrComAcentos("descEvento", desc, true)
	}
	return pcn.NodeStr("descEvento", desc, true)
}
