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
	"fmt"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Transmissao dos servicos que ESCREVEM na SEFAZ: autorizacao, consulta do
// recibo, inutilizacao, consulta de cadastro e administracao do CSC.
//
// Porte de TNFeRecepcao, TNFeRetRecepcao, TNFeRecibo, TNFeInutilizacao,
// TNFeConsultaCadastro e TAdministrarCSCNFCe (ACBrNFeWebServices.pas).
//
// Tres decisoes governam este arquivo, e as tres sao invisiveis em build e
// em teste de estrutura:
//
//  1. a versao do leiaute usada para resolver endereco e servico vem do
//     DOCUMENTO quando ha documento, nao da configuracao
//     (TNFeRecepcao.InicializarServico, :1026);
//  2. o par de servicos usado -- Autorizacao/RetAutorizacao contra
//     Recepcao/RetRecepcao -- depende de EhAutorizacao, que nao e so uma
//     comparacao de versao: tem uma excecao por UF;
//  3. cada servico escolhe a UF de busca por um criterio PROPRIO. A
//     autorizacao usa a UF do documento (com desvio para SVC); a
//     inutilizacao usa a da configuracao, sem desvio nenhum; a consulta de
//     cadastro usa a UF CONSULTADA.

// NotaFiscal e uma NF-e com o XML assinado que sera transmitido.
//
// O XML assinado e guardado como texto, e nao regerado na hora do envio,
// porque regerar invalidaria o digest: a assinatura cobre os bytes do
// infNFe como eles sairam do assinador.
type NotaFiscal struct {
	NFe         *NFe
	XMLAssinado string
}

// EhAutorizacao diz se o documento usa o par de servicos NOVO
// (NfeAutorizacao/NfeRetAutorizacao) em vez do antigo lote assincrono.
//
// Porte de TACBrNFe.EhAutorizacao (ACBrNFe.pas:313). A regra base e a
// versao, mas a NFC-e da 3.10 tem uma excecao: o Amazonas permaneceu no
// servico antigo. Portar so a comparacao de versao manda a NFC-e amazonense
// para um endereco que nao existe.
func EhAutorizacao(versao VersaoDF, modelo ModeloDF, cUF int) bool {
	if modelo == MoNFCe && versao == Ve310 {
		return cUF != cUFAmazonas
	}
	return versao >= Ve310
}

// cUFAmazonas e o 13 da excecao do EhAutorizacao.
const cUFAmazonas = 13

// OpcoesEnvio sao os tres eixos do TACBrNFe.Enviar que mudam a mensagem.
type OpcoesEnvio struct {
	// Lote e o idLote. O ACBr o recebe como string e nao o valida.
	Lote string

	// Sincrono pede a resposta com o protocolo na propria chamada, em vez
	// do recibo. Para NFC-e ele e FORCADO a true (NT 2023/002).
	Sincrono bool

	// Zipado comprime o lote com gzip e base64. Muda a SoapAction e o
	// elemento do Body.
	Zipado bool
}

// ResultadoEnvio e a resposta do envio, que tem DUAS formas.
//
// No modo assincrono a SEFAZ devolve um recibo e nada mais; no sincrono ela
// devolve o proprio resultado do processamento. Sao documentos diferentes,
// com leitores diferentes -- por isso dois campos, e nao um so tipo com
// metade dos campos vazios.
type ResultadoEnvio struct {
	// Sincrono repete o modo efetivamente usado, que pode diferir do
	// pedido: NFC-e e sempre sincrona.
	Sincrono bool

	// Recibo vem preenchido no modo ASSINCRONO; Lote traz os quatro dados
	// que a consulta seguinte precisa, ja montados.
	Recibo *RetEnviNFe
	Lote   ReciboDoLote

	// Processamento vem preenchido no modo SINCRONO, ja com o protocolo.
	Processamento *RetConsSitNFe

	// XMLRetorno e a resposta da SEFAZ como ela sera arquivada.
	//
	// No modo sincrono ela sai com o elemento JA RENOMEADO para
	// retConsSitNFe, que e o que o ACBr guarda em
	// FNFeRetornoSincrono.XmlRetorno (:1202). A renomeacao nao afeta o
	// parse -- nem aqui nem no original, porque nenhum dos dois leitores
	// confere o nome da raiz --, mas afeta o que o consumidor arquiva, e e
	// por isso que o campo existe: sem ele a renomeacao seria codigo que
	// ninguem consegue observar.
	XMLRetorno string
}

// maximoNFePorLote e o teto do lote assincrono de NF-e
// (ACBrNFeWebServices.pas:1019).
const maximoNFePorLote = 50

// tamanhoMaximoLote e o limite de 500 KB da mensagem de envio (:1159).
const tamanhoMaximoLote = 500 * 1024

// PrepararEnvio gera e assina o XML de uma nota, gravando em
// nota.XMLAssinado.
//
// A nota e gerada SOLTA, sem o envelope nfeProc: o protocolo so existe
// depois da autorizacao, e emiti-lo aqui faria a SEFAZ rejeitar o lote.
func (c *Componente) PrepararEnvio(nota *NotaFiscal) (string, error) {
	if nota == nil || nota.NFe == nil {
		return "", ErrDocumentoNulo
	}
	if c.Config.Certificado == nil {
		return "", ErrCertificadoNaoConfigurado
	}

	// Zera o protocolo durante a geracao para que o GerarXML devolva a NFe
	// solta, e o restaura em seguida -- uma nota reenviada apos consulta
	// pode ter o protocolo carregado.
	protocolo := nota.NFe.ProcNFe
	nota.NFe.ProcNFe = pcn.ProcDFe{}
	xml, err := GerarXML(nota.NFe)
	nota.NFe.ProcNFe = protocolo
	if err != nil {
		return "", err
	}

	assinado, err := dfe.AssinarXML(c.Config.Certificado, xml, "infNFe")
	if err != nil {
		return "", err
	}
	nota.XMLAssinado = assinado

	// Repovoa n.Signature com a assinatura recem-criada: e o que permite ao
	// GerarXML reembutir o bloco Signature quando montar o nfeProc depois
	// da autorizacao.
	if doc, err := pcn.ParseString(assinado); err == nil {
		pcn.LerSignature(doc.Root.FindAnyNs("Signature"), &nota.NFe.Signature)
	}
	return assinado, nil
}

// Enviar transmite um lote de notas para autorizacao.
// Porte de TNFeRecepcao.
func (c *Componente) Enviar(ctx context.Context, notas []*NotaFiscal, opcoes OpcoesEnvio) (*ResultadoEnvio, error) {
	if len(notas) == 0 {
		return nil, ErrLoteVazio
	}

	// A NT 2023/002 eliminou o lote de NFC-e: uma nota por chamada, e
	// sempre sincrona (ACBrNFeWebServices.pas:1008-1016).
	if c.Config.ModeloDF == MoNFCe {
		opcoes.Sincrono = true
		if len(notas) > 1 {
			return nil, fmt.Errorf("%w: %d NFC-e no lote, maximo 1 (NT 2023/002)",
				ErrLoteExcedido, len(notas))
		}
	} else if len(notas) > maximoNFePorLote {
		return nil, fmt.Errorf("%w: %d NF-e no lote, maximo %d",
			ErrLoteExcedido, len(notas), maximoNFePorLote)
	}

	for _, nota := range notas {
		if nota == nil || nota.NFe == nil {
			return nil, ErrDocumentoNulo
		}
		if nota.XMLAssinado == "" {
			if _, err := c.PrepararEnvio(nota); err != nil {
				return nil, err
			}
		}
	}

	destino, err := c.enderecoDoEnvio(notas)
	if err != nil {
		return nil, err
	}
	return c.enviarNaURL(ctx, destino, notas, opcoes)
}

// DestinoEnvio reune as quatro decisoes que o envio toma ANTES de abrir
// conexao: para onde vai, em que versao, sob que modelo e por qual par de
// servicos.
//
// Existe como tipo, e nao como quatro retornos, porque as quatro andam
// juntas por todo o caminho -- a SoapAction precisa do modelo e da UF, o
// enviNFe precisa da versao, o nome do servico precisa do par. Separa-las
// foi o que deixou o override de SoapAction ser buscado numa secao do
// `.ini` diferente daquela de onde a URL saiu.
type DestinoEnvio struct {
	// URL e o endereco do web service.
	URL string

	// Versao e a do DOCUMENTO, nao a da configuracao (:1026).
	Versao VersaoDF

	// Modelo e UF sao os usados na BUSCA: o modelo sai do `Ide.mod` e a UF
	// do `Ide.cUF`, com desvio para SVC (:1047 e :1066-1073).
	Modelo ModeloDF
	UF     string

	// Autorizacao diz qual par de servicos atende: true para
	// NfeAutorizacao/NfeRetAutorizacao, false para o lote assincrono antigo.
	Autorizacao bool
}

// enderecoDoEnvio resolve o destino a partir da PRIMEIRA nota do lote.
//
// Porte de TNFeRecepcao.InicializarServico (:1026) somado ao DefinirURL
// (:1037-1087). A primeira nota manda porque o lote e homogeneo: o ACBr le
// `Items[0]` e nao confere as demais.
//
// Esta separada de Enviar porque e a decisao mais facil de quebrar em
// silencio -- um lote mandado para o endereco errado volta rejeitado, nao
// com erro de programa.
func (c *Componente) enderecoDoEnvio(notas []*NotaFiscal) (DestinoEnvio, error) {
	primeira := notas[0].NFe

	if primeira.Ide.TpAmb != c.Config.Ambiente {
		return DestinoEnvio{}, ErrAmbienteDiferente
	}

	d := DestinoEnvio{
		Versao: ParseVersaoDFFloat(primeira.InfNFe.Versao),
		Modelo: modeloDoDocumento(primeira),
		UF:     ufDeAutorizacao(primeira.Ide.CUF, c.Config.FormaEmissao),
	}
	d.Autorizacao = EhAutorizacao(d.Versao, d.Modelo, primeira.Ide.CUF)

	servico := ServicoRecepcaoLote
	if d.Autorizacao {
		servico = ServicoAutorizacao
	}

	url, err := URLServico(d.Modelo, d.UF, c.Config.Ambiente, servico, d.Versao)
	if err != nil {
		return DestinoEnvio{}, err
	}
	d.URL = url
	return d, nil
}

// enviarNaURL e o Enviar com o endereco JA RESOLVIDO e o lote JA validado.
// Ver a nota em statusServicoNaURL.
func (c *Componente) enviarNaURL(ctx context.Context, destino DestinoEnvio,
	notas []*NotaFiscal, opcoes OpcoesEnvio) (*ResultadoEnvio, error) {

	versao := destino.Versao

	dadosMsg, err := c.montarEnviNFe(notas, opcoes, versao, destino.Autorizacao)
	if err != nil {
		return nil, err
	}

	servico, soapAction := servicoEActionDaAutorizacao(versao, destino.Autorizacao, opcoes.Zipado)
	if destino.Autorizacao {
		// O override sai da MESMA secao do `.ini` de onde saiu a URL -- no
		// ACBr os tres (URL, FPServico e FPSoapAction) voltam de uma unica
		// chamada a LerServicoDeParams (:1079-1087). Procurar pela UF da
		// configuracao pegaria o override de outra secao quando a nota for
		// de UF diferente, sem erro nenhum.
		soapAction = comOverrideDoIni(soapAction,
			SoapActionDoIni(destino.Modelo, destino.UF,
				c.Config.Ambiente, ServicoAutorizacao, versao),
			versao)
	}

	// O elemento do Body muda quando o lote vai comprimido (:1120).
	tagDadosMsg := "nfeDadosMsg"
	if opcoes.Zipado {
		tagDadosMsg = "nfeDadosMsgZip"
	}

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: destino.URL, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  tagDadosMsg,
		DadosMsg:     dadosMsg,
		TagCabecalho: tagCabecalhoNFe,
		// O cUF do cabecalho vem da CONFIGURACAO, nao do documento:
		// TNFeRecepcao nao sobrescreve GerarUFSoap, e o do TDFeWebService
		// le Configuracoes.WebServices.UFCodigo (ACBrDFeWebService.pas:349).
		// A URL, essa sim, saiu do Ide.cUF -- os dois eixos sao separados,
		// e unifica-los e a simplificacao que quebra o SVC.
		Cabecalho: c.cabecalhoDaVersao(versao, c.Config.CodigoUF),
		TagsResultado: []string{"nfeAutorizacaoLoteResult", "nfeAutorizacaoResult",
			"nfeAutorizacaoLoteZipResult", "nfeResultMsg", "nfeResultMsgZip",
			"nfeRecepcaoLote2Result"},
	})
	if err != nil {
		return nil, err
	}

	return c.tratarRespostaDoEnvio(resposta, notas, opcoes.Sincrono, destino)
}

// tratarRespostaDoEnvio escolhe o leitor conforme o modo da chamada.
//
// Porte de TNFeRecepcao.TratarResposta (:1175). A condicao do ramo sincrono
// e COMPOSTA: modelo 65 OU versao >= 3.10, E Sincrono -- um lote 2.00 de
// NF-e marcado como sincrono ainda volta como recibo (:1191).
func (c *Componente) tratarRespostaDoEnvio(resposta string, notas []*NotaFiscal,
	sincrono bool, destino DestinoEnvio) (*ResultadoEnvio, error) {

	if !((c.Config.ModeloDF == MoNFCe || destino.Versao >= Ve310) && sincrono) {
		recibo, err := LerRetEnviNFe(resposta)
		if err != nil {
			return nil, err
		}
		return &ResultadoEnvio{
			Sincrono:   false,
			Recibo:     recibo,
			XMLRetorno: resposta,
			Lote: ReciboDoLote{
				Recibo: recibo.InfRec.NRec,
				Modelo: destino.Modelo,
				Versao: destino.Versao,
				CUF:    cUFDoLote(notas, c.Config.CodigoUF),
			},
		}, nil
	}

	// A SEFAZ responde ao envio sincrono com retEnviNFe ou retConsReciNFe,
	// conforme a UF; o ACBr renomeia o elemento para retConsSitNFe e usa o
	// leitor da consulta de situacao (:1193-1200). Renomear e mais barato
	// que manter tres leitores com os mesmos campos.
	renomeada := renomearParaRetConsSit(resposta)
	ret, err := LerRetConsSitNFe(renomeada)
	if err != nil {
		return nil, err
	}

	anexarProtocoloSincrono(ret, notas)
	return &ResultadoEnvio{Sincrono: true, Processamento: ret, XMLRetorno: renomeada}, nil
}

// anexarProtocoloSincrono grava o protocolo na nota correspondente.
//
// Porte de :1233-1261. A condicao e DUPLA: o cStat externo tem de ser 104
// (lote processado) E o do protocolo tem de estar em CStatProcessado. Olhar
// so um dos dois anexaria protocolo de rejeicao.
func anexarProtocoloSincrono(ret *RetConsSitNFe, notas []*NotaFiscal) {
	if ret.CStat != cStatLoteProcessado || ret.ProtNFe == nil ||
		!CStatProcessado(ret.ProtNFe.CStat) {
		return
	}
	chave := pcn.RemoverLiteralChave(ret.ProtNFe.ChNFe)
	for _, nota := range notas {
		if pcn.RemoverLiteralChave(nota.NFe.InfNFe.ID) != chave {
			continue
		}
		nota.NFe.ProcNFe = pcn.ProcDFe{
			TpAmb:    ret.TpAmb,
			VerAplic: ret.VerAplic,
			ChDFe:    ret.ProtNFe.ChNFe,
			DhRecbto: ret.ProtNFe.DhRecbto,
			NProt:    ret.ProtNFe.NProt,
			DigVal:   ret.ProtNFe.DigVal,
			CStat:    ret.ProtNFe.CStat,
			XMotivo:  ret.ProtNFe.XMotivo,
			CMsg:     ret.ProtNFe.CMsg,
			XMsg:     ret.ProtNFe.XMsg,
		}
		return
	}
}

// renomearParaRetConsSit troca o nome do elemento de retorno pelo do
// documento que o leitor da consulta de situacao arquiva -- o equivalente
// dos dois StringReplace de :1193-1200.
//
// O ACBr tenta retEnviNFe primeiro e retConsReciNFe depois; so UM dos dois
// nomes aparece numa resposta, e por isso a funcao para no primeiro que
// encontrar.
//
// O `rfIgnoreCase` do original NAO foi replicado, de proposito: o teste que
// o ACBr faz antes de trocar (`pos('retEnviNFe', ...)`) e sensivel a caixa,
// entao o ignore-case so teria efeito numa resposta que misturasse caixas
// no MESMO documento -- abertura numa, fechamento noutra. Nenhum XML valido
// faz isso. Replicar daria a impressao de cobrir um caso que nao existe.
func renomearParaRetConsSit(xml string) string {
	for _, nome := range []string{"retEnviNFe", "retConsReciNFe"} {
		if !strings.Contains(xml, nome) {
			continue
		}
		return strings.ReplaceAll(xml, nome, "retConsSitNFe")
	}
	return xml
}

// montarEnviNFe monta a mensagem enviNFe.
//
// Porte de TNFeRecepcao.DefinirDadosMsg (:1140). Duas sutilezas:
//
//   - o indSinc so existe a partir da 3.10, ou na NFC-e, ou quando o
//     servico e o de autorizacao (:1146);
//   - cada nota entra como o RECORTE <NFe>...</NFe> do XML ASSINADO, nao
//     como regeracao: a assinatura cobre bytes.
func (c *Componente) montarEnviNFe(notas []*NotaFiscal, opcoes OpcoesEnvio,
	versao VersaoDF, autorizacao bool) (string, error) {

	var b strings.Builder
	b.WriteString(`<enviNFe xmlns="` + namespaceNFe + `" versao="` +
		pcn.FormatarVersaoXML(versao.Float()) + `">`)
	b.WriteString("<idLote>" + pcn.EscaparTextoXML(opcoes.Lote) + "</idLote>")

	if autorizacao || c.Config.ModeloDF == MoNFCe || versao >= Ve310 {
		indSinc := "0"
		if opcoes.Sincrono {
			indSinc = "1"
		}
		b.WriteString("<indSinc>" + indSinc + "</indSinc>")
	}

	for _, nota := range notas {
		b.WriteString(recortarNFe(nota.XMLAssinado))
	}
	b.WriteString("</enviNFe>")

	dadosMsg := b.String()
	if opcoes.Zipado {
		comprimido, err := dfe.GzipBase64(declaracaoUTF8 + dadosMsg)
		if err != nil {
			return "", err
		}
		dadosMsg = comprimido
	}

	// O limite e medido DEPOIS da compressao, porque e o que sobe na rede
	// (:1159).
	if len(dadosMsg) > tamanhoMaximoLote {
		return "", fmt.Errorf("%w: %d Kbytes, maximo %d Kbytes",
			ErrLoteExcedido, len(dadosMsg)/1024, tamanhoMaximoLote/1024)
	}
	return dadosMsg, nil
}

// declaracaoUTF8 e o CUTF8DeclaracaoXML que o ACBr prefixa ANTES de
// comprimir (:1156). Sem ele a SEFAZ nao decodifica o lote zipado.
const declaracaoUTF8 = `<?xml version="1.0" encoding="UTF-8"?>`

// recortarNFe extrai <NFe ...>...</NFe> do XML assinado.
//
// Porte do RetornarConteudoEntre('<NFe', '</NFe>') de :1152, que descarta
// um eventual envelope nfeProc ao redor.
func recortarNFe(xml string) string {
	inicio := strings.Index(xml, "<NFe")
	fim := strings.LastIndex(xml, "</NFe>")
	if inicio < 0 || fim < 0 || fim < inicio {
		return xml
	}
	return xml[inicio : fim+len("</NFe>")]
}

// servicoEActionDaAutorizacao devolve o nome do servico SOAP e a SoapAction
// do envio.
//
// Porte de TNFeRecepcao.DefinirServicoEAction (:1089). O sufixo do zipado
// MUDA com a versao: "ZIP" da 4.00 em diante, "LoteZip" antes (:1113).
func servicoEActionDaAutorizacao(versao VersaoDF, autorizacao, zipado bool) (servico, soapAction string) {
	switch {
	case !autorizacao:
		servico = enderecoWsd + "NfeRecepcao2"
		soapAction = servico
	case versao >= Ve400:
		servico = enderecoWsd + "NFeAutorizacao4"
		soapAction = servico + "/nfeAutorizacaoLote"
	default:
		servico = enderecoWsd + "NfeAutorizacao"
		soapAction = servico
	}

	if zipado {
		if versao >= Ve400 {
			soapAction += "ZIP"
		} else {
			soapAction += "LoteZip"
		}
	}
	return servico, soapAction
}

// ReciboDoLote identifica o lote cujo resultado se quer consultar.
//
// Os quatro campos existem porque, no ACBr, os quatro saem do ESTADO do
// componente -- o recibo que o envio guardou e as notas que continuam
// carregadas (TNFeRecibo.DefinirURL, :1888-1899). Esta biblioteca nao
// guarda estado entre chamadas, entao eles entram explicitos; o
// ResultadoEnvio ja devolve um preenchido.
type ReciboDoLote struct {
	// Recibo e o nRec devolvido pelo envio.
	Recibo string

	// Modelo, Versao e CUF sao os do DOCUMENTO enviado, nao os da
	// configuracao: o original os tira de `Items[0]` e so cai na
	// configuracao quando nao ha nota carregada.
	Modelo ModeloDF
	Versao VersaoDF
	CUF    int
}

// ConsultarRecibo consulta o resultado do lote pelo recibo.
//
// Porte de TNFeRecibo -- a variante SEM a espera do TNFeRetRecepcao. O
// laco de tentativas do original (:1475-1497) nao entra aqui: repetir uma
// chamada de rede e politica do consumidor, e o context ja da a ele o
// controle de cancelamento. Quem quiser o comportamento do ACBr chama em
// laco enquanto LoteEmProcessamento() for verdadeiro.
func (c *Componente) ConsultarRecibo(ctx context.Context, lote ReciboDoLote) (*RetConsReciNFe, error) {
	lote.Recibo = strings.TrimSpace(lote.Recibo)
	if lote.Recibo == "" {
		return nil, ErrReciboVazio
	}
	if lote.Modelo == 0 {
		lote.Modelo = c.Config.ModeloDF
	}

	autorizacao := EhAutorizacao(lote.Versao, lote.Modelo, lote.CUF)

	servicoTabela := ServicoRetRecepcaoLote
	if autorizacao {
		servicoTabela = ServicoRetAutorizacao
	}

	uf := ufDeAutorizacao(lote.CUF, c.Config.FormaEmissao)
	url, err := URLServico(lote.Modelo, uf, c.Config.Ambiente, servicoTabela, lote.Versao)
	if err != nil {
		return nil, err
	}
	return c.consultarReciboNaURL(ctx, url, lote, uf, autorizacao)
}

// consultarReciboNaURL e o ConsultarRecibo com o endereco JA RESOLVIDO.
func (c *Componente) consultarReciboNaURL(ctx context.Context, url string,
	lote ReciboDoLote, uf string, autorizacao bool) (*RetConsReciNFe, error) {

	versao := lote.Versao

	servico, soapAction := servicoEActionDaRetAutorizacao(versao, autorizacao)
	if autorizacao {
		// Mesma secao do `.ini` de onde saiu a URL -- ver a nota em
		// enviarNaURL.
		soapAction = comOverrideDoIni(soapAction,
			SoapActionDoIni(lote.Modelo, uf,
				c.Config.Ambiente, ServicoRetAutorizacao, versao),
			versao)
	}

	dadosMsg := pcn.NovoElem("consReciNFe").
		Attr("xmlns", namespaceNFe).
		Attr("versao", pcn.FormatarVersaoXML(versao.Float())).
		Filho(pcn.NovoElem("tpAmb").Texto(c.Config.Ambiente.String())).
		Filho(pcn.NovoElem("nRec").Texto(lote.Recibo)).
		XML()

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  "nfeDadosMsg",
		DadosMsg:     dadosMsg,
		TagCabecalho: tagCabecalhoNFe,
		// Como no envio, o cUF do cabecalho e o da configuracao: nem
		// TNFeRetRecepcao nem TNFeRecibo sobrescrevem GerarUFSoap.
		Cabecalho: c.cabecalhoDaVersao(versao, c.Config.CodigoUF),
		TagsResultado: []string{"nfeRetAutorizacaoResult",
			"nfeRetAutorizacaoLoteResult", "nfeResultMsg",
			"nfeRetRecepcao2Result"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetConsReciNFe(resposta)
}

// AnexarProtocolos grava nas notas os protocolos vindos da consulta do
// recibo.
//
// Porte de TNFeRetRecepcao.TratarRespostaFinal (:1645-1680). A regra NAO e
// a mesma do ramo sincrono, e a diferenca e do proprio ACBr:
//
//	sincrono    anexa so quando o cStat do LOTE e 104 E o do protocolo
//	            esta em CstatProcessado (:1233)
//	assincrono  anexa SEMPRE que a chave bater, qualquer que seja o cStat
//	            do protocolo; o `[100,120,150]` ali governa so a montagem
//	            do nfeProc (:1683)
//
// Uniformizar os dois e o que quem reimplementa por fora costuma fazer, e
// com isso perde o motivo da rejeicao: numa nota rejeitada o procNFe guarda
// o cStat e o xMotivo que explicam o porque.
//
// A correspondencia e pela CHAVE, nunca pela posicao.
func AnexarProtocolos(ret *RetConsReciNFe, notas []*NotaFiscal) {
	if ret == nil {
		return
	}
	for i := range ret.ProtNFe {
		prot := &ret.ProtNFe[i]
		chave := pcn.RemoverLiteralChave(prot.ChNFe)
		for _, nota := range notas {
			if nota == nil || nota.NFe == nil {
				continue
			}
			if pcn.RemoverLiteralChave(nota.NFe.InfNFe.ID) != chave {
				continue
			}
			nota.NFe.ProcNFe = pcn.ProcDFe{
				TpAmb:    prot.TpAmb,
				VerAplic: prot.VerAplic,
				ChDFe:    prot.ChNFe,
				DhRecbto: prot.DhRecbto,
				NProt:    prot.NProt,
				DigVal:   prot.DigVal,
				CStat:    prot.CStat,
				XMotivo:  prot.XMotivo,
				CMsg:     prot.CMsg,
				XMsg:     prot.XMsg,
			}
			break
		}
	}
}

// servicoEActionDaRetAutorizacao e o par do servicoEActionDaAutorizacao.
// Porte de TNFeRetRecepcao.DefinirServicoEAction (:1553).
func servicoEActionDaRetAutorizacao(versao VersaoDF, autorizacao bool) (servico, soapAction string) {
	switch {
	case !autorizacao:
		servico = enderecoWsd + "NfeRetRecepcao2"
		return servico, servico
	case versao >= Ve400:
		servico = enderecoWsd + "NFeRetAutorizacao4"
		return servico, servico + "/nfeRetAutorizacaoLote"
	default:
		servico = enderecoWsd + "NfeRetAutorizacao"
		return servico, servico
	}
}

// Inutilizar transmite um pedido de inutilizacao de faixa.
// Porte de TNFeInutilizacao.
//
// A resolucao do endereco aqui NAO passa pelo desvio de contingencia: o
// DefinirURL usa `Configuracoes.WebServices.UF` cru (:2672), sem o case de
// FormaEmissao que a autorizacao tem. Inutilizar numero e ato da SEFAZ de
// origem, nao do ambiente virtual -- e a divergencia esta registrada.
func (c *Componente) Inutilizar(ctx context.Context, p PedidoInutilizacao) (*RetInutNFe, error) {
	if err := ValidarPedidoInutilizacao(p); err != nil {
		return nil, err
	}
	modelo, err := modeloDoPedido(p.Modelo)
	if err != nil {
		return nil, err
	}

	url, err := c.urlDaInutilizacao(modelo)
	if err != nil {
		return nil, err
	}
	return c.inutilizarNaURL(ctx, url, p, modelo)
}

// urlDaInutilizacao resolve o endereco do servico de inutilizacao.
//
// Existe separada para deixar a divergencia VISIVEL e testavel: a UF e a da
// CONFIGURACAO, e so ela. O DefinirURL da inutilizacao (:2672) nao tem o
// case de FormaEmissao que o da autorizacao tem, entao em contingencia SVC
// o pedido continua indo para a SEFAZ de origem.
func (c *Componente) urlDaInutilizacao(modelo ModeloDF) (string, error) {
	return URLServico(modelo, c.Config.UF, c.Config.Ambiente,
		ServicoInutilizacao, c.Config.VersaoDF)
}

// inutilizarNaURL e o Inutilizar com o endereco JA RESOLVIDO.
//
// Devolve tambem o pedido assinado, para quem quiser montar o ProcInutNFe:
// e por isso que o XML assinado nao fica so na pilha.
func (c *Componente) inutilizarNaURL(ctx context.Context, url string,
	p PedidoInutilizacao, modelo ModeloDF) (*RetInutNFe, error) {

	if c.Config.Certificado == nil {
		return nil, ErrCertificadoNaoConfigurado
	}

	xml, err := GerarXMLInutilizacao(p, c.Config)
	if err != nil {
		return nil, err
	}
	assinado, err := dfe.AssinarXML(c.Config.Certificado, xml, "infInut")
	if err != nil {
		return nil, err
	}

	servico, soapAction := c.servicoEActionDaInutilizacao(modelo)
	soapAction = comOverrideDoIni(soapAction,
		SoapActionDoIni(modelo, c.Config.UF, c.Config.Ambiente,
			ServicoInutilizacao, c.Config.VersaoDF),
		c.Config.VersaoDF)

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  "nfeDadosMsg",
		DadosMsg:     assinado,
		TagCabecalho: tagCabecalhoNFe,
		Cabecalho:    c.cabecalhoSOAP(c.Config.CodigoUF),
		TagsResultado: []string{"nfeInutilizacaoNF2Result",
			"nfeInutilizacaoNFResult", "nfeResultMsg"},
	})
	if err != nil {
		return nil, err
	}

	ret, err := LerRetInutNFe(resposta)
	if err != nil {
		return nil, err
	}
	if ret.TemProcInutNFe() {
		ret.XMLProcInutNFe = GerarXMLProcInutNFe(assinado, resposta,
			pcn.FormatarVersaoXML(c.Config.VersaoDF.Float()))
	}
	return ret, nil
}

// servicoEActionDaInutilizacao devolve o nome do servico e a SoapAction.
// Porte de TNFeInutilizacao.DefinirServicoEAction (:2685).
//
// A excecao da Bahia aqui tem um detalhe que as outras nao tem: a
// SoapAction e atribuida FORA do guard de EstaVazio, ou seja, ela
// sobrescreve o que viesse do `.ini` (:2707). O guard protege so o
// FPServico.
func (c *Componente) servicoEActionDaInutilizacao(modelo ModeloDF) (servico, soapAction string) {
	if c.Config.VersaoDF >= Ve400 {
		servico = enderecoWsd + "NFeInutilizacao4"
		return servico, servico + "/nfeInutilizacaoNF"
	}

	if c.Config.CodigoUF == cUFBahia && modelo == MoNFe &&
		c.Config.VersaoDF == Ve310 && c.Config.FormaEmissao == pcn.TeNormal {
		servico = enderecoWsd + "NfeInutilizacao"
		return servico, servico + "/NfeInutilizacao"
	}

	servico = enderecoWsd + "NfeInutilizacao2"
	return servico, servico
}

// cUFBahia e o 29 das excecoes de notacao de servico.
const cUFBahia = 29

// ConsultarCadastro consulta o cadastro de um contribuinte.
// Porte de TNFeConsultaCadastro.
//
// `uf` e a UF CONSULTADA, nao a do emissor: vazio cai para a da
// configuracao (:2932). O documento segue a precedencia IE > CNPJ > CPF.
func (c *Componente) ConsultarCadastro(ctx context.Context, uf, ie, cnpj, cpf string) (*RetConsCad, error) {
	uf = strings.ToUpper(strings.TrimSpace(uf))
	if uf == "" {
		uf = c.Config.UF
	}
	if ie == "" && cnpj == "" && cpf == "" {
		return nil, ErrDocumentoConsultaAusente
	}

	url, err := URLServico(c.Config.ModeloDF, uf, c.Config.Ambiente,
		ServicoConsultaCadastro, c.Config.VersaoDF)
	if err != nil {
		return nil, err
	}
	return c.consultarCadastroNaURL(ctx, url, uf, ie, cnpj, cpf)
}

// consultarCadastroNaURL e o ConsultarCadastro com o endereco JA RESOLVIDO.
func (c *Componente) consultarCadastroNaURL(ctx context.Context, url, uf, ie, cnpj, cpf string) (*RetConsCad, error) {
	servico, soapAction := c.servicoEActionDaConsultaCadastro()
	soapAction = comOverrideDoIni(soapAction,
		SoapActionDoIni(c.Config.ModeloDF, uf, c.Config.Ambiente,
			ServicoConsultaCadastro, c.Config.VersaoDF),
		c.Config.VersaoDF)

	// A versao da mensagem e a do SERVICO, nao a do leiaute da nota: o
	// consCad so existe em 2.00. Ver versaoFixaPorServico.
	versaoMensagem := VersaoDaMensagem(ServicoConsultaCadastro, c.Config.VersaoDF)
	dadosMsg := GerarXMLConsCad(uf, ie, cnpj, cpf, versaoMensagem)

	// O Mato Grosso embrulha a mensagem num <nfeDadosMsg> EXTRA e troca o
	// nome do elemento do Body (:2958 e :3019). Nao e erro de leitura do
	// `.ini`: o servico de la espera a dupla camada.
	tagDadosMsg := "nfeDadosMsg"
	if c.Config.VersaoDF >= Ve400 && uf == "MT" {
		dadosMsg = "<nfeDadosMsg>" + dadosMsg + "</nfeDadosMsg>"
		tagDadosMsg = "consultaCadastro"
	}

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  tagDadosMsg,
		DadosMsg:     dadosMsg,
		TagCabecalho: tagCabecalhoNFe,
		// O cUF do cabecalho e o da UF CONSULTADA (GerarUFSoap, :3014), e o
		// versaoDados acompanha a versao do servico.
		Cabecalho: c.cabecalhoComVersao(pcn.CodigoUF(uf), versaoMensagem),
		TagsResultado: []string{"consultaCadastroResult",
			"consultaCadastro2Result", "nfeResultMsg", "consultaCadastro4Result"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetConsCad(resposta)
}

// servicoEActionDaConsultaCadastro e o porte de :2898.
//
// Repare que o nome do servico comeca com "Cad", nao com "Nfe": e o unico
// servico da NF-e com esse prefixo, e e por isso que o override de
// SoapAction da Bahia nao o alcanca -- a secao [SOAP_V4_BA] do `.ini`
// declara "CadConsultaCadastro_4.00", mas a busca usa a chave
// "NfeConsultaCadastro_4.00".
func (c *Componente) servicoEActionDaConsultaCadastro() (servico, soapAction string) {
	if c.Config.VersaoDF >= Ve400 {
		servico = enderecoWsd + "CadConsultaCadastro4"
		return servico, servico + "/consultaCadastro"
	}
	servico = enderecoWsd + "CadConsultaCadastro2"
	return servico, servico
}

// AdministrarCSC consulta, requisita ou revoga o CSC da NFC-e.
// Porte de TAdministrarCSCNFCe.
//
// E o unico servico que manda o <soap12:Header> em QUALQUER versao: o
// InicializarServico forca FPHeaderElement mesmo na 4.00 (:3940).
//
// O modelo usado na busca e o da CONFIGURACAO, e nao o 65 fixo:
// TAdministrarCSCNFCe nao sobrescreve DefinirURL, entao cai no do
// TNFeWebService (:713), que chama LerServicoDeParams com
// GetNomeModeloDFe. Um componente configurado para NF-e recebe ErrSemURL
// -- no ACBr ele receberia o raise equivalente, porque o `.ini` so declara
// AdministrarCSCNFCe nas secoes NFCe_*.
func (c *Componente) AdministrarCSC(ctx context.Context, p PedidoAdmCSC) (*RetAdmCSCNFCe, error) {
	url, err := URLServico(c.Config.ModeloDF, UFParaBusca(c.Config), c.Config.Ambiente,
		ServicoAdministrarCSC, c.Config.VersaoDF)
	if err != nil {
		return nil, err
	}
	return c.administrarCSCNaURL(ctx, url, p)
}

// administrarCSCNaURL e o AdministrarCSC com o endereco JA RESOLVIDO.
func (c *Componente) administrarCSCNaURL(ctx context.Context, url string, p PedidoAdmCSC) (*RetAdmCSCNFCe, error) {
	servico := enderecoWsd + "CscNFCe"
	soapAction := servico + "/admCscNFCe"

	// O admCscNFCe nao tem schema publicado, entao a versao e a da chave do
	// `.ini`: 1.00, em qualquer leiaute. Ver versaoFixaPorServico.
	versaoMensagem := VersaoDaMensagem(ServicoAdministrarCSC, c.Config.VersaoDF)
	dadosMsg := GerarXMLAdmCSC(p, c.Config.Ambiente, versaoMensagem)

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		TagDadosMsg:  "nfeDadosMsg",
		DadosMsg:     dadosMsg,
		TagCabecalho: tagCabecalhoNFe,
		// Sem o cabecalhoSOAP: este servico manda o header SEMPRE.
		Cabecalho:     c.conteudoDoCabecalho(c.Config.CodigoUF, versaoMensagem),
		TagsResultado: []string{"admCscNFCeResult"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetAdmCSCNFCe(resposta)
}

// cabecalhoDaVersao e o cabecalhoSOAP com a versao vinda de fora.
//
// Existe porque na autorizacao a versao sai do DOCUMENTO, e o cabecalho tem
// de concordar com ela: um lote 4.00 nao leva header, mesmo que o
// componente esteja configurado para 3.10 (:1030).
func (c *Componente) cabecalhoDaVersao(versao VersaoDF, cUF int) string {
	if versao >= Ve400 {
		return ""
	}
	return c.conteudoDoCabecalho(cUF, pcn.FormatarVersaoXML(versao.Float()))
}

// cabecalhoComVersao e o cabecalhoSOAP com o versaoDados vindo de fora.
//
// A supressao continua governada pela versao CONFIGURADA, que e o que o
// InicializarServico olha (:692) -- nao pela versao da mensagem. Na consulta
// de cadastro as duas divergem: um componente na 4.00 nao manda cabecalho,
// embora a mensagem seja 2.00.
func (c *Componente) cabecalhoComVersao(cUF int, versaoDados string) string {
	if c.Config.VersaoDF >= Ve400 {
		return ""
	}
	return c.conteudoDoCabecalho(cUF, versaoDados)
}

// conteudoDoCabecalho monta cUF e versaoDados sem decidir se o header sai.
func (c *Componente) conteudoDoCabecalho(cUF int, versaoDados string) string {
	return "<cUF>" + strconv.Itoa(cUF) + "</cUF>" +
		"<versaoDados>" + versaoDados + "</versaoDados>"
}

// ufDeAutorizacao devolve a UF sob a qual o endereco deve ser procurado na
// autorizacao e na consulta do recibo.
//
// A UF vem do DOCUMENTO (Ide.cUF), nao da configuracao, e so a contingencia
// SVC a substitui (ACBrNFeWebServices.pas:1066-1073).
func ufDeAutorizacao(cUF int, forma pcn.TipoEmissao) string {
	switch forma {
	case pcn.TeSVCAN:
		return "SVC-AN"
	case pcn.TeSVCRS:
		return "SVC-RS"
	default:
		return pcn.SiglaUF(cUF)
	}
}

// modeloDoDocumento devolve o ModeloDF correspondente ao mod do documento.
func modeloDoDocumento(n *NFe) ModeloDF {
	if n.Ide.Modelo == 65 {
		return MoNFCe
	}
	return MoNFe
}

// modeloDoPedido converte o `mod` numerico do pedido de inutilizacao.
//
// Porte do StrToModeloDF de :2666, que o original acompanha de um raise
// quando o modelo nao e reconhecido.
func modeloDoPedido(modelo int) (ModeloDF, error) {
	switch modelo {
	case 55:
		return MoNFe, nil
	case 65:
		return MoNFCe, nil
	}
	return MoNFe, fmt.Errorf("%w: %d", ErrModeloInvalido, modelo)
}

// cUFDoLote devolve o cUF da primeira nota, caindo para o configurado
// quando o lote nao tem documento -- que e o caminho do teste que chama
// tratarRespostaDoEnvio sem notas.
func cUFDoLote(notas []*NotaFiscal, padrao int) int {
	if len(notas) > 0 && notas[0] != nil && notas[0].NFe != nil {
		return notas[0].NFe.Ide.CUF
	}
	return padrao
}
