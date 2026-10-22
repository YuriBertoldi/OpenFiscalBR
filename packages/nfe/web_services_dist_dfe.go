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

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Transmissao da Distribuicao DF-e. Porte de TDistribuicaoDFe
// (ACBrNFeWebServices.pas:3953).
//
// E o servico mais simples de rotear e o mais sutil de empacotar:
//
//   - o endereco NAO varia por UF. O DefinirURL atribui `UF := 'AN'` como
//     primeira linha e nem consulta a configuracao (:4001) -- o servico e
//     tratado diretamente pela RFB;
//   - a mensagem tem DUAS camadas. O elemento do Body e
//     `nfeDistDFeInteresse` (:3983) e dentro dele vem um `nfeDadosMsg` que
//     o proprio gerador da mensagem emite (:4029). Nenhum outro servico da
//     NF-e faz isso;
//   - NAO ha cabecalho SOAP em versao nenhuma: o Clear zera o
//     FPHeaderElement (:3984), e nem a 3.10 o recebe.

// DistribuicaoDFe consulta os documentos de interesse do contribuinte.
// Porte de TDistribuicaoDFe.
func (c *Componente) DistribuicaoDFe(ctx context.Context, p PedidoDistribuicao) (*RetDistDFeInt, error) {
	url, err := c.urlDaDistribuicao(p.Chave)
	if err != nil {
		return nil, err
	}
	return c.distribuicaoNaURL(ctx, url, p)
}

// urlDaDistribuicao resolve o endereco.
//
// A UF e SEMPRE "AN". O modelo, esse sim, sai da CHAVE quando ela e
// informada (`ModeloDFe(FchNFe)`, :4011) -- e cai para o configurado
// quando nao e, que e o caso da varredura por NSU.
func (c *Componente) urlDaDistribuicao(chave string) (string, error) {
	modelo := ModeloDaChave(chave, c.Config.ModeloDF)
	return URLServico(modelo, ufDaDistribuicao, c.Config.Ambiente,
		ServicoDistribuicaoDFe, c.Config.VersaoDF)
}

// ufDaDistribuicao e o "AN" literal de :4001.
const ufDaDistribuicao = "AN"

// distribuicaoNaURL e a DistribuicaoDFe com o endereco JA RESOLVIDO.
func (c *Componente) distribuicaoNaURL(ctx context.Context, url string,
	p PedidoDistribuicao) (*RetDistDFeInt, error) {

	servico, soapAction := servicoEActionDaDistribuicao()

	versaoMensagem := VersaoDaMensagem(ServicoDistribuicaoDFe, c.Config.VersaoDF)
	dadosMsg := GerarXMLDistDFeInt(p, c.Config.Ambiente, versaoMensagem)

	cliente, err := c.clienteSOAP()
	if err != nil {
		return nil, err
	}
	resposta, err := cliente.Executar(ctx, dfe.Requisicao{
		URL: url, Servico: servico, SoapAction: soapAction,
		// O elemento do Body e o `nfeDistDFeInteresse`; o `nfeDadosMsg`
		// de dentro ja vem do GerarXMLDistDFeInt.
		TagDadosMsg: "nfeDistDFeInteresse",
		DadosMsg:    dadosMsg,
		// Sem cabecalho SOAP, em QUALQUER versao (:3984).
		TagsResultado: []string{"nfeDistDFeInteresseResult", "nfeResultMsg"},
	})
	if err != nil {
		return nil, err
	}
	return LerRetDistDFeInt(resposta)
}

// servicoEActionDaDistribuicao devolve o nome do servico e a SoapAction.
// Porte de TDistribuicaoDFe.DefinirServicoEAction (:4021).
//
// NAO ha ramo por versao aqui, ao contrario de todos os outros servicos da
// NF-e: o nome e um so, e o `.ini` nunca declara override para ele.
func servicoEActionDaDistribuicao() (servico, soapAction string) {
	servico = enderecoWsd + "NFeDistribuicaoDFe"
	return servico, servico + "/nfeDistDFeInteresse"
}

// VarrerDistribuicaoDFe percorre o lote inteiro, do ultNSU informado ate o
// fim, e devolve todos os documentos encontrados.
//
// NAO tem equivalente no ACBr, que expoe apenas a chamada unica. Existe
// porque a varredura e o uso REAL do servico e tem duas armadilhas que
// quem a escreve pela primeira vez cai:
//
//   - a RFB devolve no maximo cinquenta docZip por chamada, entao uma
//     unica chamada quase nunca traz tudo;
//   - o criterio de parada e `ultNSU == maxNSU`, nao "o lote veio vazio" e
//     nem "o cStat foi 137" -- um lote cheio com ultNSU abaixo do maxNSU
//     significa que falta mais.
//
// O `limite` protege contra laco infinito caso a RFB devolva um ultNSU que
// nao avanca. Zero ou negativo usa limiteDeChamadasPadrao.
//
// Bater no limite NAO e erro e nao tem sentinela proprio: o retorno e o
// mesmo de uma varredura completa. Quem precisa distinguir olha o
// RetDistDFeInt devolvido -- `VarreduraConcluida()` falso com erro nil
// significa que o limite foi atingido e ha mais documentos, e o `UltNSU`
// desse retorno e de onde a proxima varredura recomeca.
func (c *Componente) VarrerDistribuicaoDFe(ctx context.Context, p PedidoDistribuicao,
	limite int) ([]DocZip, *RetDistDFeInt, error) {

	return varrer(p, limite, func(pedido PedidoDistribuicao) (*RetDistDFeInt, error) {
		return c.DistribuicaoDFe(ctx, pedido)
	})
}

// varrer e o LACO da varredura, separado de quem faz a chamada.
//
// A separacao existe para que o teste exercite ESTE laco, e nao uma copia
// dele: um criterio de parada reescrito no teste nao prova nada sobre o
// que roda em producao -- foi assim que a comparacao textual de NSU passou
// despercebida na primeira versao.
func varrer(p PedidoDistribuicao, limite int,
	consultar func(PedidoDistribuicao) (*RetDistDFeInt, error)) ([]DocZip, *RetDistDFeInt, error) {

	if limite <= 0 {
		limite = limiteDeChamadasPadrao
	}

	// A varredura ignora NSU e Chave: as duas buscam UM documento, e
	// repetir a chamada devolveria o mesmo item para sempre.
	p.NSU, p.Chave = "", ""

	var todos []DocZip
	var ultimo *RetDistDFeInt

	for i := 0; i < limite; i++ {
		ret, err := consultar(p)
		if err != nil {
			return todos, ultimo, err
		}
		ultimo = ret

		if !ret.Atendida() {
			return todos, ret, nil
		}
		todos = append(todos, ret.DocZip...)

		if ret.VarreduraConcluida() {
			return todos, ret, nil
		}

		// Sem AVANCO do ultNSU a proxima chamada traria o mesmo lote.
		//
		// A comparacao e NUMERICA: a RFB devolve o NSU preenchido com
		// quinze zeros e o consumidor costuma guardar o numero cru, entao
		// "7" e "000000000000007" sao o mesmo NSU. Comparar como texto
		// gasta uma chamada a toa em toda varredura que comece de um NSU
		// nao preenchido.
		if ret.UltNSU == "" || nsuComZeros(ret.UltNSU) <= nsuComZeros(p.UltNSU) {
			return todos, ret, nil
		}
		p.UltNSU = ret.UltNSU
	}
	return todos, ultimo, nil
}

// limiteDeChamadasPadrao e o teto de chamadas de uma varredura.
//
// Cem chamadas cobrem cinco mil documentos, que e mais do que um
// contribuinte acumula entre duas varreduras diarias.
const limiteDeChamadasPadrao = 100

// ChaveDoDocZip devolve a chave de acesso do item, venha ela do resumo, do
// documento completo ou do evento.
//
// Os tres grupos guardam a chave em campos diferentes, e qual deles esta
// preenchido depende do schema. Sem esta funcao, todo consumidor
// reescreveria o mesmo `switch`.
func ChaveDoDocZip(d DocZip) string {
	switch {
	case d.ResDFe.ChDFe != "":
		return pcn.RemoverLiteralChave(d.ResDFe.ChDFe)
	case d.ResEvento.ChDFe != "":
		return pcn.RemoverLiteralChave(d.ResEvento.ChDFe)
	case d.ProcEvento.ChDFe != "":
		return pcn.RemoverLiteralChave(d.ProcEvento.ChDFe)
	}
	return ""
}
