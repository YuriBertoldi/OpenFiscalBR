// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package main

import (
	"math"
	"net/http"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/nfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Catalogo de exemplos da demo.
//
// TODOS OS DADOS SAO FICTICIOS. O CNPJ 99.999.999/0001-91 e o que o proprio
// ACBr usa nos exemplos dele, e a inscricao estadual, o endereco e os
// produtos foram inventados. Nenhum certificado, CNPJ real ou dado de
// cliente entra aqui.
//
// Os exemplos sao construidos em Go e serializados na hora, em vez de
// ficarem como texto fixo: assim eles nao envelhecem em relacao ao package
// -- se um campo mudar de nome, isto deixa de compilar em vez de passar a
// gerar XML invalido em silencio.

// Exemplo e uma entrada do catalogo.
type Exemplo struct {
	ID           string
	Nome         string
	Descricao    string
	Formato      string // "xml" ou "ini"
	AcaoSugerida string
	Montar       func() (conteudo string, chave string, err error)
}

// ExemploResumo e a entrada do catalogo, sem o conteudo.
type ExemploResumo struct {
	ID           string `json:"id"`
	Nome         string `json:"nome"`
	Descricao    string `json:"descricao"`
	Formato      string `json:"formato"`
	AcaoSugerida string `json:"acaoSugerida"`
}

// ExemplosResponse e o corpo de GET /api/exemplos.
type ExemplosResponse struct {
	Total    int             `json:"total"`
	Exemplos []ExemploResumo `json:"exemplos"`
	Aviso    string          `json:"aviso"`
}

// ExemploResponse e o corpo de GET /api/exemplos/{id}.
type ExemploResponse struct {
	ExemploResumo
	Chave    string `json:"chave,omitempty"`
	Conteudo string `json:"conteudo"`
}

const avisoDadosFicticios = "Todos os dados sao ficticios e servem apenas " +
	"para experimentar a API. Nao transmita a SEFAZ."

func catalogo() []Exemplo {
	return []Exemplo{
		{
			ID:           "nfe-venda",
			Nome:         "NF-e de venda (modelo 55)",
			Descricao:    "Operacao interestadual com dois itens, ICMS tributado e duplicata.",
			Formato:      "xml",
			AcaoSugerida: "ler",
			Montar:       func() (string, string, error) { return exemploXML(notaDeVenda()) },
		},
		{
			ID:           "nfce-consumidor",
			Nome:         "NFC-e ao consumidor (modelo 65)",
			Descricao:    "Venda presencial com pagamento em dinheiro e troco.",
			Formato:      "xml",
			AcaoSugerida: "ler",
			Montar:       func() (string, string, error) { return exemploXML(notaDeConsumidor()) },
		},
		{
			ID:   "nfe-ini",
			Nome: "NF-e no formato .ini do ACBr",
			Descricao: "A mesma nota de venda, no formato que a maioria dos ERP gera. " +
				"Use em /api/ini-para-xml para ver a conversao.",
			Formato:      "ini",
			AcaoSugerida: "ini-para-xml",
			Montar:       func() (string, string, error) { return exemploINI(notaDeVenda()) },
		},
	}
}

// ---------------------------------------------------------------------------
// Montagem dos documentos
// ---------------------------------------------------------------------------

// baseDaNota monta o que as duas notas de exemplo tem em comum.
func baseDaNota() nfe.NFe {
	n := nfe.NovaNFe()
	n.InfNFe.Versao = 4.00

	n.Ide = nfe.Ide{
		CUF: 35, CNF: 54071382, NatOp: "VENDA DE MERCADORIA",
		Modelo: 55, Serie: 1, NNF: 1001,
		// A emissao e RELATIVA a agora, e nao uma data fixa: a regra 228
		// recusa nota com mais de trinta dias de atraso, entao um exemplo
		// com data cravada passa a levantar rejeicao sozinho com o tempo.
		DEmi:    time.Now().Add(-time.Hour).Truncate(time.Second),
		TpNF:    pcn.TnSaida,
		CMunFG:  3550308,
		TpImp:   pcn.TiRetrato,
		TpEmis:  pcn.TeNormal,
		TpAmb:   pcn.TaHomologacao,
		FinNFe:  nfe.FnNormal,
		ProcEmi: pcn.PeAplicativoContribuinte,
		VerProc: "OpenFiscalBR-demo",
	}

	n.Emit = nfe.NovoEmit()
	n.Emit.CNPJCPF = "99999999000191"
	n.Emit.XNome = "EMPRESA EXEMPLO LTDA"
	n.Emit.XFant = "EXEMPLO"
	// IE de SP com digito verificador valido -- a regra 209 confere.
	n.Emit.IE = "110042490114"
	n.Emit.CRT = nfe.CrtRegimeNormal
	n.Emit.EnderEmit = nfe.EnderEmit{
		XLgr: "RUA DOS EXEMPLOS", Nro: "100", XBairro: "CENTRO",
		CMun: 3550308, XMun: "SAO PAULO", UF: "SP", CEP: 1001000,
		CPais: 1058, XPais: "BRASIL", Fone: "1133334444",
	}

	return n
}

// notaDeVenda monta uma NF-e interestadual com dois itens.
func notaDeVenda() nfe.NFe {
	n := baseDaNota()

	n.Ide.IdDest = nfe.DoInterestadual
	n.Ide.IndFinal = nfe.CfNao
	n.Ide.IndPres = nfe.PcNao

	n.Dest.CNPJCPF = "11111111000191"
	n.Dest.XNome = "CLIENTE EXEMPLO SA"
	n.Dest.IndIEDest = pcn.InContribuinte
	n.Dest.IE = "222222222"
	n.Dest.EnderDest = nfe.EnderDest{
		XLgr: "AVENIDA DAS NACOES", Nro: "2000", XBairro: "SAVASSI",
		CMun: 3106200, XMun: "BELO HORIZONTE", UF: "MG", CEP: 30140000,
		CPais: 1058, XPais: "BRASIL",
	}

	n.Det = []nfe.DetItem{
		itemDeVenda(1, "P001", "CAFE TORRADO E MOIDO 500G", "09012100",
			"6102", "PC", 100, 18.50),
		itemDeVenda(2, "P002", "ACUCAR REFINADO 1KG", "17019900",
			"6102", "PC", 50, 4.20),
	}
	n.Total.ICMSTot = totalizar(n.Det)

	n.Transp.ModFrete = nfe.MfContaEmitente
	n.Transp.Vol = []nfe.Vol{{QVol: 3, Esp: "CAIXA", PesoL: 60, PesoB: 63}}

	// A cobranca acompanha o total: duas parcelas iguais, vencendo em
	// trinta e sessenta dias.
	total := n.Total.ICMSTot.VNF
	parcela := arredondar(total / 2)
	n.Cobr.Fat = nfe.Fat{NFat: "F1001", VOrig: total, VLiq: total}
	n.Cobr.Dup = []nfe.Dup{
		{NDup: "001", DVenc: n.Ide.DEmi.AddDate(0, 0, 30), VDup: parcela},
		{NDup: "002", DVenc: n.Ide.DEmi.AddDate(0, 0, 60), VDup: arredondar(total - parcela)},
	}

	pag := nfe.NovoPagItem()
	pag.TPag = nfe.FpBoletoBancario
	pag.VPag = total
	n.Pag.DetPag = []nfe.PagItem{pag}

	n.InfAdic.InfCpl = "Documento de exemplo com dados ficticios."
	return n
}

// notaDeConsumidor monta uma NFC-e de venda presencial.
func notaDeConsumidor() nfe.NFe {
	n := baseDaNota()

	n.Ide.Modelo = 65
	n.Ide.NNF = 2002
	n.Ide.IdDest = nfe.DoInterna
	n.Ide.IndFinal = nfe.CfConsumidorFinal
	n.Ide.IndPres = nfe.PcPresencial
	n.Ide.TpImp = pcn.TiNFCe

	// A NFC-e tem DEZ MINUTOS de tolerancia de atraso, nao trinta dias
	// como a NF-e (regra 704). O exemplo emite agora mesmo.
	n.Ide.DEmi = time.Now().Truncate(time.Second)

	// E nao admite frete (regra 753). O zero value de ModalidadeFrete e
	// "por conta do emitente", que a NFC-e recusa -- aqui precisa ser
	// explicito.
	n.Transp.ModFrete = nfe.MfSemFrete

	// A NFC-e ao consumidor nao identificado nao leva destinatario.
	n.Det = []nfe.DetItem{
		itemDeVenda(1, "P010", "AGUA MINERAL 500ML", "22011000", "5102", "UN", 2, 3.50),
	}
	n.Total.ICMSTot = totalizar(n.Det)

	// Pagamento em dinheiro com troco: o vPag cobre o total e o troco e a
	// diferenca. O troco e da COLECAO de pagamentos, nao do item.
	total := n.Total.ICMSTot.VNF
	pag := nfe.NovoPagItem()
	pag.TPag = nfe.FpDinheiro
	pag.VPag = total
	n.Pag.DetPag = []nfe.PagItem{pag}
	n.Pag.VTroco = 0

	return n
}

// itemDeVenda monta um item tributado pelo ICMS, com PIS e COFINS.
func itemDeVenda(nItem int, codigo, descricao, ncm, cfop, unidade string,
	quantidade, valorUnitario float64) nfe.DetItem {

	total := quantidade * valorUnitario

	d := nfe.NovoDetItem()
	d.Prod.NItem = nItem
	d.Prod.CProd = codigo
	// A regra 889 exige o GTIN. Produto sem codigo de barras declara o
	// literal "SEM GTIN" -- deixar vazio e rejeicao, nao omissao.
	//
	// No caminho do XML o gerador preenche isso sozinho; no do .ini, nao:
	// o leitor so ATRIBUI o cEAN quando ha conteudo. Preencher aqui faz os
	// dois caminhos saírem iguais.
	d.Prod.CEAN = "SEM GTIN"
	d.Prod.CEANTrib = "SEM GTIN"
	d.Prod.XProd = descricao
	d.Prod.NCM = ncm
	d.Prod.CFOP = cfop
	d.Prod.UCom = unidade
	d.Prod.QCom = quantidade
	d.Prod.VUnCom = valorUnitario
	d.Prod.VProd = total
	d.Prod.UTrib = unidade
	d.Prod.QTrib = quantidade
	d.Prod.VUnTrib = valorUnitario
	d.Prod.IndTot = nfe.ItSomaTotalNFe

	d.Imposto.ICMS = nfe.ICMS{
		Orig: pcn.OeNacional, CST: pcn.CST00,
		ModBC: nfe.DbiValorOperacao,
		VBC:   total, PICMS: 12, VICMS: arredondar(total * 0.12),
	}
	d.Imposto.PIS = nfe.PIS{
		CST: pcn.Pis01, VBC: total, PPIS: 1.65, VPIS: arredondar(total * 0.0165),
	}
	d.Imposto.COFINS = nfe.COFINS{
		CST: pcn.Cof01, VBC: total, PCOFINS: 7.6, VCOFINS: arredondar(total * 0.076),
	}
	return d
}

// totalizar soma os itens e monta o ICMSTot.
//
// Os totais sao CALCULADOS, e nao escritos a mao: a regra 602 e as irmas
// dela conferem cada total contra o somatorio dos itens, e um exemplo com
// total digitado passa a levantar rejeicao assim que alguem mexe num item.
func totalizar(itens []nfe.DetItem) nfe.ICMSTot {
	var t nfe.ICMSTot
	for _, d := range itens {
		t.VProd += d.Prod.VProd
		t.VBC += d.Imposto.ICMS.VBC
		t.VICMS += d.Imposto.ICMS.VICMS
		t.VPIS += d.Imposto.PIS.VPIS
		t.VCOFINS += d.Imposto.COFINS.VCOFINS
	}
	t.VProd = arredondar(t.VProd)
	t.VBC = arredondar(t.VBC)
	t.VICMS = arredondar(t.VICMS)
	t.VPIS = arredondar(t.VPIS)
	t.VCOFINS = arredondar(t.VCOFINS)
	t.VNF = t.VProd
	return t
}

// arredondar devolve o valor com duas casas.
//
// Precisa existir porque o XML sai com duas casas e as regras de negocio
// comparam o total contra o somatorio DOS VALORES GRAVADOS: somar em
// float64 e nao arredondar deixa uma diferenca de centavos que a SEFAZ
// recusa.
func arredondar(v float64) float64 {
	return math.Round(v*100) / 100
}

// exemploXML gera o XML da nota, com a chave ja calculada.
func exemploXML(n nfe.NFe) (string, string, error) {
	xml, err := nfe.GerarXML(&n)
	if err != nil {
		return "", "", err
	}
	return xml, n.InfNFe.ID, nil
}

// exemploINI gera o .ini da nota.
//
// A chave precisa ser montada e ATRIBUIDA antes: o gravador do .ini recusa
// nota com chave invalida, como o original faz, e o MontarChaveAcesso
// DEVOLVE a chave sem grava-la -- quem a monta decide onde ela vai. No
// caminho do XML isso acontece dentro do GerarXML; aqui e explicito.
func exemploINI(n nfe.NFe) (string, string, error) {
	chave, err := nfe.MontarChaveAcesso(&n)
	if err != nil {
		return "", "", err
	}
	n.InfNFe.ID = chave

	texto, err := nfe.GravarINI(&n)
	if err != nil {
		return "", "", err
	}
	return texto, n.InfNFe.ID, nil
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func handleListarExemplos(w http.ResponseWriter, _ *http.Request) {
	lista := catalogo()
	resp := ExemplosResponse{Total: len(lista), Aviso: avisoDadosFicticios}
	for _, ex := range lista {
		resp.Exemplos = append(resp.Exemplos, ExemploResumo{
			ID: ex.ID, Nome: ex.Nome, Descricao: ex.Descricao,
			Formato: ex.Formato, AcaoSugerida: ex.AcaoSugerida,
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleExemplo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	for _, ex := range catalogo() {
		if ex.ID != id {
			continue
		}
		conteudo, chave, err := ex.Montar()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, ExemploResponse{
			ExemploResumo: ExemploResumo{
				ID: ex.ID, Nome: ex.Nome, Descricao: ex.Descricao,
				Formato: ex.Formato, AcaoSugerida: ex.AcaoSugerida,
			},
			Chave:    chave,
			Conteudo: conteudo,
		})
		return
	}
	writeError(w, http.StatusNotFound, "exemplo nao encontrado: "+id)
}
