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
	"context"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Transmissao dos eventos. Porte de TNFeEnvEvento
// (ACBrNFeWebServices.pas:3030).
//
// O que torna este servico diferente de todos os outros da NF-e e o
// ROTEAMENTO: eventos do mesmo documento vao para SEFAZ DIFERENTES,
// conforme o tipo. Nenhum outro servico faz isso.
//
//	CCe, cancelamento, cancelamento por substituicao e prorrogacao
//	                        -> SEFAZ da UF da chave (ou SVC em contingencia)
//	manifestacao do destinatario e demais eventos "nao locais"
//	                        -> Ambiente Nacional (UF "AN")
//	insucesso na entrega e seu cancelamento
//	                        -> SVRS, sempre
//	conciliacao financeira  -> SVRS na NF-e; UF da chave na NFC-e
//	eventos da Reforma      -> SVRS, sempre
//	EPEC de NFC-e em SP     -> servico PROPRIO (EventoEPEC)
//
// Mandar um evento para a SEFAZ errada nao da erro de rede: da rejeicao,
// e a mensagem nao diz que o endereco era o problema.

// EnviarEvento transmite um lote de eventos.
// Porte de TNFeEnvEvento.
//
// Assina cada evento que ainda nao tenha XMLAssinado.
func (c *Componente) EnviarEvento(ctx context.Context, lote *EventoNFe) (*RetEventoNFe, error) {
	if lote == nil || len(lote.Evento) == 0 {
		return nil, ErrLoteVazio
	}
	if lote.Versao == "" {
		lote.Versao = VersaoDaMensagem(ServicoRecepcaoEvento, c.Config.VersaoDF)
	}

	destino, err := c.enderecoDoEvento(lote.Evento[0].InfEvento)
	if err != nil {
		return nil, err
	}
	return c.enviarEventoNaURL(ctx, destino, lote)
}

// DestinoEvento e o endereco e o servico escolhidos para um evento.
type DestinoEvento struct {
	URL string

	// Modelo sai da CHAVE do documento (ModeloDFe, :3089).
	Modelo ModeloDF

	// UF e a unidade sob a qual o endereco foi procurado -- que raramente
	// e a do emitente. Guardada para que a busca da SoapAction use a MESMA
	// secao do `.ini`.
	UF string

	// Servico distingue o RecepcaoEvento comum do EventoEPEC de Sao Paulo.
	Servico Servico
}

// enderecoDoEvento resolve para onde o evento vai.
// Porte de TNFeEnvEvento.DefinirURL (:3076-3160).
func (c *Componente) enderecoDoEvento(inf InfEvento) (DestinoEvento, error) {
	chave := pcn.RemoverLiteralChave(inf.ChNFe)
	d := DestinoEvento{
		Modelo:  ModeloDaChave(chave, c.Config.ModeloDF),
		Servico: ServicoRecepcaoEvento,
	}

	// A contingencia SVC vale como PONTO DE PARTIDA e pode ser sobrescrita
	// pelos desvios abaixo -- a ordem das atribuicoes no Delphi e o que
	// decide isso (:3096 antes de :3114).
	switch c.Config.FormaEmissao {
	case pcn.TeSVCAN:
		d.UF = "SVC-AN"
	case pcn.TeSVCRS:
		d.UF = "SVC-RS"
	default:
		d.UF = UFDaChave(chave)
	}

	// O EPEC de NFC-e tem servico proprio, e SO em Sao Paulo: as demais
	// UFs nao o oferecem porque permitem contingencia off-line (:3101).
	// Repare que as duas condicoes olham a CONFIGURACAO, nao a chave.
	if inf.TpEvento == pcn.TeEPECNFe && c.Config.CodigoUF == cUFSaoPaulo &&
		c.Config.ModeloDF == MoNFCe {
		d.Servico = ServicoEventoEPEC
		url, err := URLServico(d.Modelo, d.UF, c.Config.Ambiente, d.Servico, c.Config.VersaoDF)
		if err != nil {
			return DestinoEvento{}, err
		}
		d.URL = url
		return d, nil
	}

	// Tudo que NAO e evento local vai para o Ambiente Nacional. A lista de
	// locais e fechada e curta (:3109).
	if !eventoLocal(inf.TpEvento) {
		d.UF = "AN"
	}

	// E tres grupos saem do Ambiente Nacional de volta para uma SEFAZ
	// virtual (:3114-3143). Sao desvios POSTERIORES ao AN, e por isso
	// vencem.
	switch {
	case inf.TpEvento == pcn.TeInsucessoEntregaNFe ||
		inf.TpEvento == pcn.TeCancInsucessoEntregaNFe:
		d.UF = "SVRS"

	case inf.TpEvento == pcn.TeConcFinanceira ||
		inf.TpEvento == pcn.TeCancConcFinanceira:
		// O unico desvio que olha o modelo da CONFIGURACAO, e nao o da
		// chave: NF-e vai para o SVRS, NFC-e volta para a UF da chave.
		if c.Config.ModeloDF == MoNFe {
			d.UF = "SVRS"
		} else {
			d.UF = UFDaChave(chave)
		}

	case eventoDaReformaTributaria(inf.TpEvento):
		d.UF = "SVRS"
	}

	url, err := URLServico(d.Modelo, d.UF, c.Config.Ambiente, d.Servico, c.Config.VersaoDF)
	if err != nil {
		return DestinoEvento{}, err
	}
	d.URL = url
	return d, nil
}

// cUFSaoPaulo e o 35 da excecao do EPEC de NFC-e.
const cUFSaoPaulo = 35

// eventoLocal diz se o evento e tratado pela SEFAZ da UF do documento.
//
// Porte da lista de :3109. Sao sete tipos, e todos os demais -- inclusive
// a manifestacao do destinatario, que e o caso mais comum -- vao para o
// Ambiente Nacional.
//
// O comentario do ACBr explica a regra: "somente os eventos de Cancelamento
// e CCe serao tratados pela SEFAZ do estado, os outros eventos como
// manifestacao de destinatarios serao tratados diretamente pela RFB".
func eventoLocal(t pcn.TipoEvento) bool {
	switch t {
	case pcn.TeCCe, pcn.TeCancelamento, pcn.TeCancSubst,
		pcn.TePedProrrog1, pcn.TePedProrrog2,
		pcn.TeCanPedProrrog1, pcn.TeCanPedProrrog2:
		return true
	}
	return false
}

// eventoDaReformaTributaria diz se o evento pertence ao grupo que o ACBr
// roteia para o SVRS em :3133.
//
// A lista NAO coincide com "eventos que a Reforma criou": o
// teAtualizacaoDataPrevisaoEntrega esta nela, e ele nao e tributario.
// Deriva-la de um criterio semantico erraria.
func eventoDaReformaTributaria(t pcn.TipoEvento) bool {
	switch t {
	case pcn.TeCancGenerico, pcn.TePagIntegLibCredPresAdq, pcn.TeImporALCZFM,
		pcn.TePerecPerdaRouboFurtoTranspContratFornec,
		pcn.TeFornecNaoRealizPagAntec, pcn.TeSolicApropCredPres,
		pcn.TeDestItemConsPessoal,
		pcn.TePerecPerdaRouboFurtoTranspContratAqu,
		pcn.TeAceiteDebitoApuracaoNotaCredito, pcn.TeImobilizacaoItem,
		pcn.TeSolicApropCredCombustivel, pcn.TeSolicApropCredBensServicos,
		pcn.TeManifPedTransfCredIBSSucessao,
		pcn.TeManifPedTransfCredCBSSucessao,
		pcn.TeAtualizacaoDataPrevisaoEntrega:
		return true
	}
	return false
}

// enviarEventoNaURL e o EnviarEvento com o endereco JA RESOLVIDO.
func (c *Componente) enviarEventoNaURL(ctx context.Context, destino DestinoEvento,
	lote *EventoNFe) (*RetEventoNFe, error) {

	if c.Config.Certificado == nil {
		return nil, ErrCertificadoNaoConfigurado
	}

	dadosMsg, err := c.montarEnvEvento(lote)
	if err != nil {
		return nil, err
	}

	servico, soapAction := c.servicoEActionDoEvento()
	soapAction = comOverrideDoIni(soapAction,
		SoapActionDoIni(destino.Modelo, destino.UF, c.Config.Ambiente,
			destino.Servico, c.Config.VersaoDF),
		c.Config.VersaoDF)

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: destino.URL, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  "nfeDadosMsg",
		DadosMsg:     dadosMsg,
		TagCabecalho: tagCabecalhoNFe,
		// O cUF do cabecalho e o do ORGAO do evento, que pode ser 91 na
		// manifestacao -- nao o da configuracao.
		Cabecalho: c.cabecalhoComVersao(COrgaoDoEvento(lote.Evento[0].InfEvento),
			lote.Versao),
		TagsResultado: []string{"nfeRecepcaoEventoResult", "nfeResultMsg"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetEventoNFe(resposta)
}

// montarEnvEvento assina os eventos do lote e devolve o envelope.
//
// A assinatura e por EVENTO, nao por lote: cada <evento> leva a sua
// Signature sobre o proprio infEvento. Depois de assinados, o envelope sai
// do MESMO GerarXMLEvento que o consumidor chamaria -- manter dois
// montadores foi a primeira versao deste arquivo, e os dois ja divergiam
// no tratamento do idLote.
func (c *Componente) montarEnvEvento(lote *EventoNFe) (string, error) {
	for i := range lote.Evento {
		if lote.Evento[i].XMLAssinado != "" {
			continue
		}
		if _, err := c.AssinarEvento(&lote.Evento[i], lote.Versao); err != nil {
			return "", err
		}
	}
	return GerarXMLEvento(lote)
}

// AssinarEvento gera e assina UM evento, gravando em ev.XMLAssinado.
//
// O elemento assinado e o `infEvento`, nao o `evento` -- como na nota, onde
// e o `infNFe`.
func (c *Componente) AssinarEvento(ev *Evento, versao string) (string, error) {
	if c.Config.Certificado == nil {
		return "", ErrCertificadoNaoConfigurado
	}
	if versao == "" {
		versao = VersaoDaMensagem(ServicoRecepcaoEvento, c.Config.VersaoDF)
	}

	no, err := gerarEvento(ev, versao)
	if err != nil {
		return "", err
	}

	assinado, err := dfe.AssinarXML(c.Config.Certificado, no.XML(), "infEvento")
	if err != nil {
		return "", err
	}
	ev.XMLAssinado = assinado

	if doc, err := pcn.ParseString(assinado); err == nil {
		pcn.LerSignature(doc.Root.FindAnyNs("Signature"), &ev.Signature)
	}
	return assinado, nil
}

// recortarEvento extrai <evento ...>...</evento> do XML assinado, como o
// recortarNFe faz com a nota.
func recortarEvento(xml string) string {
	inicio := strings.Index(xml, "<evento")
	fim := strings.LastIndex(xml, "</evento>")
	if inicio < 0 || fim < 0 || fim < inicio {
		return xml
	}
	return xml[inicio : fim+len("</evento>")]
}

// servicoEActionDoEvento devolve o nome do servico SOAP e a SoapAction.
// Porte de TNFeEnvEvento.DefinirServicoEAction (:3163).
func (c *Componente) servicoEActionDoEvento() (servico, soapAction string) {
	if c.Config.VersaoDF >= Ve400 {
		servico = enderecoWsd + "NFeRecepcaoEvento4"
		return servico, servico + "/nfeRecepcaoEvento"
	}
	servico = enderecoWsd + "RecepcaoEvento"
	return servico, servico
}

// GerarXMLProcEvento monta o procEventoNFe -- o evento processado, que e o
// que se arquiva.
//
// Porte do envelope que o ACBr monta apos o registro: cola o evento
// ASSINADO com o retEvento recebido, os dois como vieram.
func GerarXMLProcEvento(eventoAssinado string, ret RetInfEvento, versao string) string {
	return `<procEventoNFe versao="` + versao + `" xmlns="` + namespaceNFe + `">` +
		recortarEvento(pcn.RemoverDeclaracaoXML(eventoAssinado)) +
		`<retEvento versao="` + versao + `">` +
		pcn.RemoverDeclaracaoXML(ret.XML) +
		`</retEvento></procEventoNFe>`
}

// Cancelamento monta e envia o evento de cancelamento.
// Porte de TACBrNFe.Cancelamento (ACBrNFe.pas:423).
//
// Duas divergencias deliberadas:
//
//   - NAO consulta a situacao da nota antes para descobrir o protocolo.
//     Obte-lo e do chamador, que normalmente ja o tem do retorno da
//     autorizacao -- mesma decisao tomada no packages/nfag;
//   - o nSeqEvento sai 1, nao 0. O TACBrNFe.Cancelamento deixa o campo em
//     zero (ACBrNFe.pas:812) e o Id termina em "00", o que a SEFAZ nao
//     aceita: a sequencia comeca em 1. E correcao de defeito deterministico
//     do ACBr, nao mudanca de contrato.
func (c *Componente) Cancelamento(ctx context.Context, chave, protocolo, justificativa string,
	idLote int64, agora time.Time) (*RetEventoNFe, error) {

	return c.EnviarEvento(ctx, &EventoNFe{
		IDLote: idLote,
		Evento: []Evento{{
			InfEvento: InfEvento{
				TpAmb:      c.Config.Ambiente,
				ChNFe:      pcn.RemoverLiteralChave(strings.TrimSpace(chave)),
				DhEvento:   agora,
				TpEvento:   pcn.TeCancelamento,
				NSeqEvento: 1,
				DetEvento: DetEvento{
					NProt: protocolo,
					XJust: justificativa,
				},
			},
		}},
	})
}

// EnviarCartaCorrecao monta e envia uma CC-e.
// Porte de TACBrNFe.EnviarCartaCorrecao (ACBrNFe.pas:472).
//
// O xCondUso e preenchido pelo gerador com o texto legal.
func (c *Componente) EnviarCartaCorrecao(ctx context.Context, chave, correcao string,
	sequencia int, idLote int64, agora time.Time) (*RetEventoNFe, error) {

	return c.EnviarEvento(ctx, &EventoNFe{
		IDLote: idLote,
		Evento: []Evento{{
			InfEvento: InfEvento{
				TpAmb:      c.Config.Ambiente,
				ChNFe:      pcn.RemoverLiteralChave(strings.TrimSpace(chave)),
				DhEvento:   agora,
				TpEvento:   pcn.TeCCe,
				NSeqEvento: sequencia,
				DetEvento:  DetEvento{XCorrecao: correcao},
			},
		}},
	})
}

// ManifestarDestinatario monta e envia uma das quatro manifestacoes.
//
// O `cnpjDoManifestante` e OBRIGATORIO e nao pode sair da chave: quem
// manifesta e o DESTINATARIO, e a chave carrega o documento do EMITENTE.
// Omiti-lo faria o documentoDoAutor cair na chave e declarar o emitente
// como autor da manifestacao.
//
// O cOrgao e forcado a 91: a manifestacao e processada pelo Ambiente
// Nacional, e deixar o COrgaoDoEvento cair para o cUF da chave a mandaria
// para a UF errada.
func (c *Componente) ManifestarDestinatario(ctx context.Context, chave string,
	tipo pcn.TipoEvento, cnpjDoManifestante, justificativa string, idLote int64,
	agora time.Time) (*RetEventoNFe, error) {

	if !ehManifestacaoDestinatario(tipo) {
		return nil, ErrEventoNaoSuportado
	}

	return c.EnviarEvento(ctx, &EventoNFe{
		IDLote: idLote,
		Evento: []Evento{{
			InfEvento: InfEvento{
				COrgao:     COrgaoAmbienteNacional,
				TpAmb:      c.Config.Ambiente,
				CNPJ:       cnpjDoManifestante,
				ChNFe:      pcn.RemoverLiteralChave(strings.TrimSpace(chave)),
				DhEvento:   agora,
				TpEvento:   tipo,
				NSeqEvento: 1,
				DetEvento:  DetEvento{XJust: justificativa},
			},
		}},
	})
}
