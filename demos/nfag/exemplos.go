// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-02.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/nfag"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Catalogo de XMLs de exemplo da demo. Cada cenario e montado preenchendo as
// structs do package e chamando o writer REAL (nfag.GerarXML e companhia) --
// nunca um XML estatico embutido. XML de arquivo nao exercita as condicionais
// do writer e apodrece em silencio quando ele muda.
//
// Os dados sao ficticios e FIXOS: com ParametrosExemplo zerado, o mesmo
// cenario gera sempre o mesmo XML, byte a byte. E isso que permite ao usuario
// so clicar e gerar, e aos testes compararem o resultado.
//
// NAO existe cenario "varios CFOP" aqui, e nao e esquecimento: a NFAg nao tem
// CFOP no leiaute -- o campo nao existe em Prod (packages/nfag/classes.go) e
// nao aparece em lugar nenhum do package. Quem precisa de CFOP esta na NFGas.

// ParametrosExemplo sobrescreve os dados ficticios do cenario. Campo vazio
// (ou zero) mantem o ficticio -- o catalogo gera documento completo sem que o
// usuario preencha nada.
type ParametrosExemplo struct {
	CNPJEmit string // cnpj  -- CNPJ do emitente
	UF       string // uf    -- move cUF, UF do emitente e cMun juntos
	Serie    int    // serie
	NNF      int    // nnf
	TpAmb    int    // tpamb -- 1 producao, 2 homologacao
	Itens    int    // itens -- so nos cenarios que aceitam
	Variar   bool   // variar -- sorteia nNF, cNF, serie e dhEmi
}

// Exemplo e um cenario do catalogo.
type Exemplo struct {
	ID           string
	Nome         string
	Descricao    string
	Tipo         string // "documento" | "evento"
	AcaoSugerida string // data-acao da aba que melhor demonstra o cenario
	Arquivo      string // nome sugerido para download
	Observacao   string // ressalva mostrada junto da descricao
	AceitaItens  bool   // o campo "itens" do formulario vale neste cenario
	Gerar        func(p ParametrosExemplo) (string, error)
}

// catalogoExemplos devolve os cenarios na ordem em que aparecem na tela.
func catalogoExemplos() []Exemplo {
	return []Exemplo{
		{
			ID:           "transmissao",
			Nome:         "Transmissao - NFAg pronta para envio",
			Descricao:    "NFAg avulsa, sem assinatura e sem protocolo, como vai para a SEFAZ.",
			Tipo:         "documento",
			AcaoSugerida: "gerar",
			Arquivo:      "nfag-transmissao.xml",
			Gerar:        exemploTransmissao,
		},
		{
			ID:           "autorizada",
			Nome:         "Autorizada - nfagProc com protocolo",
			Descricao:    "Documento + protNFAg (cStat 100), que e o que o emitente arquiva e o importador recebe.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfag-autorizada.xml",
			Observacao:   "protocolo e assinatura sao ficticios: serve para importar, nao para transmitir",
			Gerar:        exemploAutorizada,
		},
		{
			ID:           "cancelamento",
			Nome:         "Cancelamento - evento avulso",
			Descricao:    "eventoNFAg com evCancNFAg (tpEvento 110111), amarrado a chave do cenario de transmissao.",
			Tipo:         "evento",
			AcaoSugerida: "ler-evento",
			Arquivo:      "nfag-evento-cancelamento.xml",
			Gerar:        exemploCancelamento,
		},
		{
			ID:           "cancelamento-proc",
			Nome:         "Cancelamento - procEvento com retorno",
			Descricao:    "procEventoNFAg: evento enviado + retEventoNFAg homologado (cStat 135).",
			Tipo:         "evento",
			AcaoSugerida: "ler-evento",
			Arquivo:      "nfag-proc-evento-cancelamento.xml",
			Gerar:        exemploCancelamentoProc,
		},
		{
			ID:           "completo",
			Nome:         "Completo - todas as informacoes de tag",
			Descricao:    "Todo grupo opcional do leiaute preenchido: gCompraGov, gSub, gFatConjunto, gMed, gTarif, gMedicao, IBSCBS, retTrib, TFS, TFU, gProcRef, pgtoVinc, gFat, gAgencia, gQualiAgua, autXML, infAdic, infPAA, gRespTec, protocolo e assinatura.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfag-completa.xml",
			Observacao:   "nenhum documento tem TODAS as tags do writer ao mesmo tempo -- varios grupos se excluem entre si; aqui estao as de um documento real maximamente preenchido",
			Gerar:        exemploCompleto,
		},
		{
			ID:           "erro-leitura",
			Nome:         "Com erro - quebra na leitura",
			Descricao:    "XML sem o atributo versao em infNFAg: o leitor recusa com ErrAtributoVersaoAusente (HTTP 422).",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfag-erro-leitura.xml",
			Observacao:   "erro proposital para testar o tratamento de documento invalido na importacao",
			Gerar:        exemploErroLeitura,
		},
		{
			ID:           "erro-regras",
			Nome:         "Com erro - reprova na validacao",
			Descricao:    "Le normalmente, mas viola as regras 226 (cMun x cUF), 227 (concatenacao da chave) e 247 (sigla da UF), e tem CNPJ de emitente com digito invalido.",
			Tipo:         "documento",
			AcaoSugerida: "validar",
			Arquivo:      "nfag-erro-regras.xml",
			Observacao:   "o digito verificador da chave fica INTEGRO de proposito: chave valida com concatenacao divergente e o caso que mais escapa de um importador",
			Gerar:        exemploErroRegras,
		},
		{
			ID:           "multi-itens",
			Nome:         "Varios itens",
			Descricao:    "Cinco itens de categorias diferentes (residencial, comercial, industrial, social, rural), com total coerente com a soma.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfag-varios-itens.xml",
			AceitaItens:  true,
			Gerar:        exemploMultiItens,
		},
	}
}

// acharExemplo localiza um cenario pelo ID.
func acharExemplo(id string) (Exemplo, bool) {
	for _, ex := range catalogoExemplos() {
		if ex.ID == id {
			return ex, true
		}
	}
	return Exemplo{}, false
}

// ---------------------------------------------------------------------------
// Parametros
// ---------------------------------------------------------------------------

// municipioPorUF e a capital de cada UF, com o codigo IBGE. A regra 226 so
// confere os dois primeiros digitos do codigo contra o codigo da UF, mas usar
// municipio real mantem o exemplo critivel.
var municipioPorUF = map[string]struct {
	Codigo int
	Nome   string
}{
	"AC": {1200401, "Rio Branco"}, "AL": {2704302, "Maceio"},
	"AM": {1302603, "Manaus"}, "AP": {1600303, "Macapa"},
	"BA": {2927408, "Salvador"}, "CE": {2304400, "Fortaleza"},
	"DF": {5300108, "Brasilia"}, "ES": {3205309, "Vitoria"},
	"GO": {5208707, "Goiania"}, "MA": {2111300, "Sao Luis"},
	"MG": {3106200, "Belo Horizonte"}, "MS": {5002704, "Campo Grande"},
	"MT": {5103403, "Cuiaba"}, "PA": {1501402, "Belem"},
	"PB": {2507507, "Joao Pessoa"}, "PE": {2611606, "Recife"},
	"PI": {2211001, "Teresina"}, "PR": {4106902, "Curitiba"},
	"RJ": {3304557, "Rio de Janeiro"}, "RN": {2408102, "Natal"},
	"RO": {1100205, "Porto Velho"}, "RR": {1400100, "Boa Vista"},
	"RS": {4314902, "Porto Alegre"}, "SC": {4205407, "Florianopolis"},
	"SE": {2800308, "Aracaju"}, "SP": {3550308, "Sao Paulo"},
	"TO": {1721000, "Palmas"},
}

// validarParametros confere o dominio de cada campo informado. Fica separado
// da aplicacao para que a camada HTTP possa recusar com 400 antes de montar
// documento nenhum.
func validarParametros(p ParametrosExemplo) error {
	if p.CNPJEmit != "" && len(pcn.OnlyCPFCNPJAlphaNum(p.CNPJEmit)) != 14 {
		return fmt.Errorf("cnpj: %q nao tem 14 posicoes", p.CNPJEmit)
	}
	if p.UF != "" {
		if _, ok := municipioPorUF[strings.ToUpper(strings.TrimSpace(p.UF))]; !ok {
			return fmt.Errorf("uf: %q nao e uma UF valida", p.UF)
		}
	}
	if p.TpAmb != 0 && p.TpAmb != 1 && p.TpAmb != 2 {
		return fmt.Errorf("tpamb: %d nao existe (1 producao, 2 homologacao)", p.TpAmb)
	}
	if p.Serie < 0 || p.Serie > 889 {
		return fmt.Errorf("serie: %d fora da faixa 0..889", p.Serie)
	}
	if p.NNF < 0 {
		return fmt.Errorf("nnf: %d nao pode ser negativo", p.NNF)
	}
	if p.Itens < 0 || p.Itens > 990 {
		return fmt.Errorf("itens: %d fora da faixa 1..990 do leiaute", p.Itens)
	}
	return nil
}

// aplicarParametros sobrescreve no documento o que o usuario informou. Roda
// SEMPRE antes de GerarXML: o writer recalcula o Id e o dV a partir dos
// campos, entao a chave sai coerente com o que foi trocado.
func aplicarParametros(n *nfag.NFAg, p ParametrosExemplo) error {
	if err := validarParametros(p); err != nil {
		return err
	}

	if p.CNPJEmit != "" {
		n.Emit.CNPJ = pcn.OnlyCPFCNPJAlphaNum(p.CNPJEmit)
	}

	if p.UF != "" {
		uf := strings.ToUpper(strings.TrimSpace(p.UF))
		mun := municipioPorUF[uf]
		// Os tres campos andam JUNTOS: trocar so o cUF faria o documento
		// reprovar nas regras 226 e 247 sem o usuario entender por que.
		n.Ide.CUF = pcn.CodigoUF(uf)
		n.Ide.CMunFG = mun.Codigo
		n.Emit.EnderEmit.UF = uf
		n.Emit.EnderEmit.CMun = mun.Codigo
		n.Emit.EnderEmit.XMun = mun.Nome
		n.Dest.EnderDest.UF = uf
		n.Dest.EnderDest.CMun = mun.Codigo
		n.Dest.EnderDest.XMun = mun.Nome
	}

	switch p.TpAmb {
	case 1:
		n.Ide.TpAmb = pcn.TaProducao
	case 2:
		n.Ide.TpAmb = pcn.TaHomologacao
	}

	if p.Serie != 0 {
		n.Ide.Serie = p.Serie
	}
	if p.NNF != 0 {
		n.Ide.NNF = p.NNF
	}

	if p.Variar {
		// cNF zerado devolve o sorteio ao writer (gerarCodigoDFe).
		n.Ide.CNF = 0
		n.Ide.NNF = 1 + rand.Intn(999999)
		n.Ide.Serie = 1 + rand.Intn(10)
		n.Ide.DhEmi = time.Now()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Base ficticia
// ---------------------------------------------------------------------------

// dataBase e a data de emissao fixa dos exemplos. Fixa-la (e fixar cNF) e o
// que torna o XML gerado identico a cada chamada.
var dataBase = time.Date(2026, 3, 15, 10, 30, 0, 0, time.FixedZone("BRT", -3*3600))

// notaBase monta a NFAg ficticia comum a todos os cenarios: emitente,
// destinatario, ligacao, um item com PIS/COFINS e os totais. Quem precisa de
// mais grupos acrescenta por cima.
func notaBase() *nfag.NFAg {
	n := &nfag.NFAg{
		InfNFAg: nfag.InfNFAg{Versao: 1.00},
		Ide: nfag.Ide{
			CUF: 35, TpAmb: pcn.TaHomologacao, Modelo: 75, Serie: 1, NNF: 1,
			CNF: 1, DhEmi: dataBase, TpEmis: pcn.TeNormal,
			NSiteAutoriz: nfag.Sa0, CMunFG: 3550308,
			FinNFAg: nfag.FnNormal, TpFat: nfag.TfNormal,
			VerProc: "OpenFiscalBR Demo 1.0",
		},
		Emit: nfag.Emit{
			CNPJ:  "11222333000181",
			IE:    "111222333444",
			XNome: "Companhia de Saneamento Exemplo SA",
			XFant: "SaneExemplo",
			EnderEmit: nfag.Endereco{
				XLgr: "Rua dos Reservatorios", Nro: "1000", XBairro: "Industrial",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1310100, UF: "SP",
				Fone: "1130001000", Email: "fiscal@saneexemplo.com.br",
			},
		},
		Dest: nfag.Dest{
			XNome:   "Consumidor Exemplo",
			CNPJCPF: "52998224725",
			IE:      "ISENTO",
			EnderDest: nfag.Endereco{
				XLgr: "Avenida Central", Nro: "200", XCpl: "Apto 101",
				XBairro: "Centro", CMun: 3550308, XMun: "Sao Paulo",
				CEP: 1001000, UF: "SP", Fone: "1199990000",
				Email: "consumidor@exemplo.com.br",
			},
		},
		Ligacao: nfag.Ligacao{
			IDLigacao:    "LIG-000123",
			IDCodCliente: "CLI-987",
			TpLigacao:    nfag.TlAguaEsgoto,
		},
		Det: []nfag.Det{itemAgua(1, nfag.TcResidencial, "Agua tratada", 25.5, 4.80)},
		GFat: nfag.GFat{
			CompetFat: dataBase,
			DVencFat:  dataBase.AddDate(0, 0, 26),
			NFat:      "FAT-2026-03-0001",
			CodBarras: "84670000009487200011222333000181000000000001",
		},
	}
	recalcularTotal(n)
	return n
}

// itemAgua monta um item de fornecimento de agua da categoria informada.
// A descricao serve tanto para xProd quanto para xCategoria, que no leiaute
// e a categoria por extenso.
func itemAgua(nItem int, categoria nfag.TpCategoria, descricao string, qtd, tarifa float64) nfag.Det {
	vProd := arredondar(qtd*tarifa, 2)
	return nfag.Det{
		NItem: nItem,
		GTarif: []nfag.GTarif{{
			DIniTarif:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			DFimTarif:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			NAto:        "ATO-123",
			AnoAto:      2026,
			TpFaixaCons: nfag.TfcMinimo,
		}},
		Prod: nfag.Prod{
			IndOrigemQtd: nfag.IoMedido,
			CProd:        fmt.Sprintf("AGUA%03d", nItem),
			XProd:        descricao,
			CClass:       "0100101",
			TpCategoria:  categoria,
			XCategoria:   descricao,
			QEconomias:   "1",
			UMed:         nfag.UmM3,
			QFaturada:    qtd,
			VItem:        tarifa,
			VProd:        vProd,
		},
		Imposto: nfag.Imposto{
			PIS: nfag.PIS{
				CST: pcn.Pis01, VBC: vProd, PPIS: 1.65,
				VPIS: arredondar(vProd*0.0165, 2),
			},
			COFINS: nfag.COFINS{
				CST: pcn.Cof01, VBC: vProd, PCOFINS: 7.60,
				VCOFINS: arredondar(vProd*0.076, 2),
			},
		},
	}
}

// recalcularTotal soma os itens no grupo total, para que o documento gerado
// nunca saia com total divergente da soma -- o cenario de erro de regras e
// quem introduz divergencia, e de proposito.
//
// O ACBr NAO calcula totais: vNF so e lido, escrito e copiado, nunca
// derivado -- preencher o grupo total e responsabilidade do emitente. Como
// estes exemplos existem para ser importados e conferidos, o vNF e composto
// aqui pelas parcelas que acrescem ao valor da nota (as duas taxas de
// fiscalizacao). vTotDFe acompanha o vNF, como na fixture do package.
func recalcularTotal(n *nfag.NFAg) {
	var t nfag.Total
	for i := range n.Det {
		imp := n.Det[i].Imposto
		t.VProd += n.Det[i].Prod.VProd
		t.VPIS += imp.PIS.VPIS
		t.VCOFINS += imp.COFINS.VCOFINS
		t.VTFS += imp.TFS.VTFS
		t.VTFU += imp.TFU.VTFU
		t.VRetPIS += imp.RetTrib.VRetPIS
		t.VRetCOFINS += imp.RetTrib.VRetCOFINS
		t.VRetCSLL += imp.RetTrib.VRetCSLL
		t.VIRRF += imp.RetTrib.VIRRF
	}
	t.VProd = arredondar(t.VProd, 2)
	t.VNF = arredondar(t.VProd+t.VTFS+t.VTFU, 2)
	t.VTotDFe = t.VNF
	n.Total = t
}

func arredondar(v float64, casas int) float64 {
	f := 1.0
	for i := 0; i < casas; i++ {
		f *= 10
	}
	return float64(int64(v*f+0.5)) / f
}

// ---------------------------------------------------------------------------
// Builders dos cenarios
// ---------------------------------------------------------------------------

func exemploTransmissao(p ParametrosExemplo) (string, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	return nfag.GerarXML(n)
}

func exemploAutorizada(p ParametrosExemplo) (string, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	// A chave do protocolo so existe depois que o writer a monta; gerar duas
	// vezes e o caminho mais honesto -- a primeira passada calcula a chave.
	if _, err := nfag.GerarXML(n); err != nil {
		return "", err
	}
	aplicarProtocolo(n)
	assinarFicticio(n)
	return nfag.GerarXMLProc(n)
}

func exemploCancelamento(p ParametrosExemplo) (string, error) {
	ev, err := montarEventoCancelamento(p)
	if err != nil {
		return "", err
	}
	return nfag.GerarXMLEvento(ev)
}

func exemploCancelamentoProc(p ParametrosExemplo) (string, error) {
	ev, err := montarEventoCancelamento(p)
	if err != nil {
		return "", err
	}
	ret := &nfag.RetEventoNFAg{
		Versao:    "1.00",
		TemEvento: true,
		Evento:    *ev,
		RetInfEvento: nfag.RetInfEvento{
			TpAmb:       ev.InfEvento.TpAmb,
			VerAplic:    "OpenFiscalBR_DEMO_1.0",
			COrgao:      ev.InfEvento.COrgaoEfetivo(),
			CStat:       135,
			XMotivo:     "Evento registrado e vinculado a NFAg",
			ChNFAg:      ev.InfEvento.ChNFAg,
			TpEvento:    nfag.TeCancelamento,
			XEvento:     "Cancelamento",
			NSeqEvento:  ev.InfEvento.NSeqEvento,
			CNPJDest:    "99888777000166",
			EmailDest:   "consumidor@exemplo.com.br",
			COrgaoAutor: ev.InfEvento.COrgaoEfetivo(),
			DhRegEvento: dataBase.AddDate(0, 0, 1).Add(5 * time.Minute),
			NProt:       "135260000000099",
		},
	}
	return nfag.GerarXMLProcEvento(ret)
}

// montarEventoCancelamento amarra o evento a chave do cenario de transmissao
// -- evento apontando para chave que o usuario nao tem nao serve de teste.
func montarEventoCancelamento(p ParametrosExemplo) (*nfag.EventoNFAg, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return nil, err
	}
	if _, err := nfag.GerarXML(n); err != nil {
		return nil, err
	}
	return &nfag.EventoNFAg{
		Versao: "1.00",
		InfEvento: nfag.InfEvento{
			COrgao:     n.Ide.CUF,
			TpAmb:      n.Ide.TpAmb,
			CNPJ:       n.Emit.CNPJ,
			ChNFAg:     n.ChaveAcesso(),
			DhEvento:   dataBase.AddDate(0, 0, 1),
			TpEvento:   nfag.TeCancelamento,
			NSeqEvento: 1,
			DetEvento: nfag.DetEvento{
				NProt: "135260000000001",
				XJust: "Erro na leitura do hidrometro no periodo faturado",
			},
		},
	}, nil
}

func exemploErroLeitura(p ParametrosExemplo) (string, error) {
	xml, err := exemploTransmissao(p)
	if err != nil {
		return "", err
	}
	// Remover o atributo versao de infNFAg e o erro mais limpo que se pode
	// fabricar: cai num sentinela nomeado (ErrAtributoVersaoAusente) em vez
	// de erro de sintaxe, e atinge igualmente ler, validar, gerar e
	// transmitir, que passam todos por LerXMLString.
	const alvo = ` versao="1.00"`
	if !strings.Contains(xml, alvo) {
		return "", fmt.Errorf("exemplo erro-leitura: %q nao encontrado no XML gerado "+
			"(o writer mudou de formato; ajuste o builder)", alvo)
	}
	return strings.Replace(xml, alvo, "", 1), nil
}

func exemploErroRegras(p ParametrosExemplo) (string, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}

	// Regra 226: municipio do emitente de outra UF (BH com cUF de SP).
	n.Emit.EnderEmit.CMun = 3106200
	n.Emit.EnderEmit.XMun = "Belo Horizonte"
	// Regra 247: sigla da UF do emitente diferente da UF autorizadora.
	n.Emit.EnderEmit.UF = "MG"
	// CNPJ com digito verificador errado.
	n.Emit.CNPJ = "11222333000199"

	xml, err := nfag.GerarXML(n)
	if err != nil {
		return "", err
	}

	// Regra 227 (concatenacao) tem que ser quebrada DEPOIS de gerar: mexer
	// no campo e regerar faria o writer recalcular a chave, e o documento
	// voltaria a ser coerente. O digito verificador fica integro de
	// proposito -- chave valida com concatenacao divergente e justamente o
	// caso que escapa de quem so confere o dV.
	alvo := fmt.Sprintf("<nNF>%d</nNF>", n.Ide.NNF)
	if !strings.Contains(xml, alvo) {
		return "", fmt.Errorf("exemplo erro-regras: %q nao encontrado no XML gerado "+
			"(o writer mudou de formato; ajuste o builder)", alvo)
	}
	return strings.Replace(xml, alvo, "<nNF>999</nNF>", 1), nil
}

func exemploMultiItens(p ParametrosExemplo) (string, error) {
	if err := validarParametros(p); err != nil {
		return "", err
	}
	qtd := p.Itens
	if qtd <= 0 {
		qtd = 5
	}

	categorias := []struct {
		Cat  nfag.TpCategoria
		Desc string
	}{
		{nfag.TcResidencial, "Agua tratada - residencial"},
		{nfag.TcComercial, "Agua tratada - comercial"},
		{nfag.TcIndustrial, "Agua tratada - industrial"},
		{nfag.TcSocial, "Agua tratada - tarifa social"},
		{nfag.TcRural, "Agua tratada - rural"},
	}

	n := notaBase()
	n.Det = nil
	for i := 0; i < qtd; i++ {
		c := categorias[i%len(categorias)]
		n.Det = append(n.Det, itemAgua(i+1, c.Cat, c.Desc, float64(10+i*5), 4.80+float64(i)*0.2))
	}
	recalcularTotal(n)

	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	return nfag.GerarXML(n)
}

// ---------------------------------------------------------------------------
// Cenario completo
// ---------------------------------------------------------------------------

// exemploCompleto preenche TODO grupo opcional do leiaute que pode conviver
// num mesmo documento.
func exemploCompleto(p ParametrosExemplo) (string, error) {
	n := notaBase()

	n.Ide.DhCont = dataBase.Add(-2 * time.Hour)
	n.Ide.XJust = "Contingencia por indisponibilidade do autorizador"
	n.Ide.TpPagAnt = pcn.TpaPagServicoContinuado
	n.Ide.GCompraGov = rtc.GCompraGovReduzido{
		TpEnteGov: rtc.TcgEstados,
		PRedutor:  12.3456,
		TpOperGov: rtc.TogFornecimento,
		RefDFe: []rtc.RefDFeAnt{
			{RefDFeAnt: "35260111222333000181750010000000001100000003"},
			{RefDFeAnt: "35260111222333000181750010000000002100000002"},
		},
	}

	n.Emit.ISUFEmit = "ISUF123"
	n.Dest.IM = "IM123"
	n.Dest.CNIS = "12345678901"
	n.Dest.XNomeAdicional = "Apto 101"

	n.Ligacao.LatGPS = "-23.550520"
	n.Ligacao.LongGPS = "-46.633308"
	n.Ligacao.CodRoteiroLeitura = "ROT-A12"

	n.GSub = nfag.GSub{
		ChNFAg: "35260211222333000181750010000000001100000005",
		MotSub: nfag.MsErroLeitura,
	}
	n.GFatConjunto = nfag.GFatConjunto{
		ChNFAgFat: "35260211222333000181750010000000009100000001",
	}

	n.GMed = []nfag.GMed{{
		NMed: 1, IDMedidor: "HID-0001",
		DMedAnt: dataBase.AddDate(0, -1, 0),
		DMedAtu: dataBase,
	}}

	// Item 1: o item cheio -- medicao com leitura, segunda faixa tarifaria,
	// pagamento antecipado, IBS/CBS, tributos retidos, as duas taxas de
	// fiscalizacao, processo referenciado e informacao adicional.
	det := itemAgua(1, nfag.TcResidencial, "Agua tratada", 25.5, 4.80)
	prod := &det.Prod
	prod.GMedicao = nfag.GMedicao{
		NMed: 1,
		GMedida: nfag.GMedida{
			TpGrMed:      nfag.TgmAguaTratada,
			NUnidConsumo: "1",
			VUnidConsumo: 1,
			UMed:         nfag.UmM3,
			VMedAnt:      1000.1234,
			VMedAtu:      1025.6234,
			VConst:       1,
			VMed:         25.5,
		},
	}
	prod.FatorPoluicao = 0.8
	prod.IndDevolucao = pcn.TieNao
	prod.GPagAntecipado = rtc.GPagAntecipadoProd{
		ChDFePagAnt: "35260111222333000181750010000000001100000003",
		NItemPagAnt: 1,
	}
	det.GTarif = append(det.GTarif, nfag.GTarif{
		DIniTarif:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NAto:        "ATO-124",
		AnoAto:      2026,
		TpFaixaCons: nfag.TfcMedio,
	})

	vProd := prod.VProd
	det.Imposto.RetTrib = nfag.RetTrib{
		VRetPIS: 1.11, VRetCOFINS: 2.22, VRetCSLL: 3.33,
		VBCIRRF: 100.00, VIRRF: 4.44,
	}
	det.Imposto.TFS = nfag.TFS{VBCTFS: vProd, PTFS: 0.5, VTFS: arredondar(vProd*0.005, 2)}
	det.Imposto.TFU = nfag.TFU{VBCTFU: vProd, PTFU: 0.3, VTFU: arredondar(vProd*0.003, 2)}
	det.Imposto.IBSCBS = ibscbsExemplo(vProd)

	det.GProcRef = nfag.GProcRef{
		VItem: 1.23456789, QFaturada: 10.5, VProd: 12.96,
		IndDevolucao: pcn.TieSim,
		GProc: []nfag.GProc{
			{TpProc: nfag.TpJusticaEstadual, NProcesso: "0001234-55.2026.8.26.0100"},
			{TpProc: nfag.TpProcon, NProcesso: "PROCON-2026-9"},
		},
	}
	det.InfAdProd = "Consumo referente a marco/2026"

	// Item 2: referencia item de documento anterior (atributos chNFAgAnt e
	// nItemAnt) e leva a medicao SEM leitura, unico caminho para o
	// tpMotNaoLeitura (xml_writer.go:405). Na NFAg esse ramo E alcancavel:
	// a guarda do gMedicao exige so nMed > 0, diferente da NFGas.
	det2 := itemAgua(2, nfag.TcServicoPublico, "Coleta de esgoto", 10, 3.20)
	det2.ChNFAgAnt = "35260211222333000181750010000000001100000005"
	det2.NItemAnt = 1
	det2.GTarif = nil
	// Sem leitura no medidor, o volume de esgoto e calculado a partir do
	// consumo de agua -- indOrigemQtd 4, nao 6 (que e "sem quantidade" e
	// nao combina com qFaturada preenchida).
	det2.Prod.IndOrigemQtd = nfag.IoCalculada
	det2.Prod.CProd = "ESG001"
	det2.Prod.GMedicao = nfag.GMedicao{
		NMed:            1,
		TpMotNaoLeitura: nfag.TmDistribuidora,
	}

	n.Det = []nfag.Det{det, det2}
	recalcularTotal(n)
	n.Total.IBSCBSTot = ibscbsTotExemplo(n.Total.VProd)

	n.PgtoVinc = rtc.PgtoVinc{Pgto: []rtc.Pgto{
		{NPag: 1, IDTransacao: "TX-0001", TpMeioPgto: "03",
			CNPJReceb: "11222333000181", CNPJBasePSP: "11222333"},
		{NPag: 2, IDTransacao: "TX-0002", TpMeioPgto: "17"},
	}}

	n.GFat.DApresFat = dataBase.AddDate(0, 0, 5)
	n.GFat.DProxLeitura = dataBase.AddDate(0, 1, 0)
	// codDebAuto exclui codBanco+codAgencia (xml_writer.go:555); ficamos com
	// o par, que cobre duas tags.
	n.GFat.CodBanco = "001"
	n.GFat.CodAgencia = "1234"
	n.GFat.GPIX = nfag.GPIX{URLQRCodePIX: "https://pix.exemplo.com.br/qr/0001"}
	n.GFat.EnderCorresp = nfag.Endereco{
		XLgr: "Rua da Correspondencia", Nro: "50", XCpl: "Sala 3",
		XBairro: "Jardins", CMun: 3550308, XMun: "Sao Paulo",
		CEP: 1400000, UF: "SP", Fone: "1155554444",
		Email: "correspondencia@exemplo.com.br",
	}

	n.GAgencia = nfag.GAgencia{
		Econ:              "1",
		EconAcumulada:     "12",
		SPrestador:        "SELO-PRESTADOR-001",
		DEmissSelo:        dataBase.AddDate(0, -3, 0),
		SRegulador:        "SELO-REGULADOR-001",
		NAgenciaAtend:     "0800 123 4567",
		EnderAgenciaAtend: "Rua do Atendimento, 10",
		GHistCons: []nfag.GHistCons{{
			XHistorico: "Consumo dos ultimos meses",
			MedMensal:  24.5,
			GCons: []nfag.GCons{
				{CompetFat: dataBase.AddDate(0, -1, 0), UMed: nfag.UmM3,
					QtdDias: "28", MedDiaria: 0.8929, Consumo: 25, VolFat: 25},
				{CompetFat: dataBase.AddDate(0, -2, 0), UMed: nfag.UmM3,
					QtdDias: "31", MedDiaria: 0.7742, Consumo: 24, VolFat: 24},
			},
		}},
	}

	n.GQualiAgua = nfag.GQualiAgua{
		CompetAnalise: dataBase,
		GAnalise: []nfag.GAnalise{
			{
				XItemAnalisado: "Turbidez", NAmostraMinima: "10",
				NAmostraAnalisada: "12", NAmostraFPadrao: "0",
				NAmostraDPadrao: "12", NMediaMensal: "0.35",
				XValorReferencia: "ate 5,0 uT",
			},
			{
				XItemAnalisado: "Cloro residual livre", NAmostraMinima: "10",
				NAmostraAnalisada: "12", NAmostraFPadrao: "0",
				NAmostraDPadrao: "12", NMediaMensal: "1.20",
				XValorReferencia: "0,2 a 5,0 mg/L",
			},
		},
		Conclusao:    "Agua distribuida dentro dos padroes de potabilidade",
		CProcesso:    "PROC-QUALI-2026-03",
		SistemaAbast: "Sistema Produtor Exemplo",
	}

	n.AutXML = []nfag.AutXML{
		{CNPJCPF: "99888777000166"},
		{CNPJCPF: "52998224725"},
	}

	n.InfAdic = nfag.InfAdic{
		InfAdFisco: "Documento de exemplo gerado pela demo do OpenFiscalBR",
		InfCpl: []string{
			"Informacao complementar 1",
			"Informacao complementar 2",
		},
	}

	n.InfPAA = nfag.InfPAA{CNPJPAA: "99888777000166"}

	n.InfRespTec = nfag.InfRespTec{
		CNPJ: "99888777000166", XContato: "Responsavel Tecnico",
		Email: "ti@saneexemplo.com.br", Fone: "1140004000",
		IDCSRT: 1, HashCSRT: "SGVsbG9Xb3JsZEhhc2hDU1JU",
	}

	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}

	// Primeira passada para o writer montar a chave; so entao da para
	// preencher o QR-Code e o protocolo, que dependem dela.
	if _, err := nfag.GerarXML(n); err != nil {
		return "", err
	}
	n.InfNFAgSupl.QrCodNFAg = fmt.Sprintf(
		"https://www.nfag.fazenda.sp.gov.br/consulta?chNFAg=%s&tpAmb=%s",
		n.ChaveAcesso(), n.Ide.TpAmb.String())
	aplicarProtocolo(n)
	assinarFicticio(n)

	return nfag.GerarXML(n)
}

// ibscbsExemplo monta o grupo de IBS/CBS do item com todos os subgrupos que
// convivem: diferimento, devolucao, reducao, ZFM, estorno, transferencia e
// ajuste de competencia.
func ibscbsExemplo(vProd float64) rtc.IBSCBS {
	return rtc.IBSCBS{
		// CST 000 e obrigatorio para que o gIBSCBS saia: no modelo NFAg o
		// writer do rtc so gera o grupo quando o CST e 000
		// (packages/rtc/xml_writer.go:224).
		CST:        rtc.CST000,
		CClassTrib: "000001",
		// indDoacao so e emitido quando TieSim (rtc/xml_writer.go:144).
		IndDoacao: pcn.TieSim,
		GIBSCBS: rtc.GIBSCBS{
			VBC:  vProd,
			VIBS: arredondar(vProd*0.085, 2),
			GIBSUF: rtc.GIBSUFValores{
				PIBSUF:   0.1,
				GDif:     rtc.GDif{PDif: 1.0, VDif: 1.0},
				GDevTrib: rtc.GDevTrib{PDevTrib: 2.0, VDevTrib: 2.0},
				GRed:     rtc.GRed{PRedAliq: 3.0, PAliqEfet: 4.0},
				VIBSUF:   arredondar(vProd*0.001, 2),
			},
			GIBSMun: rtc.GIBSMunValores{
				PIBSMun:  0.2,
				GDif:     rtc.GDif{PDif: 1.0, VDif: 3.0},
				GDevTrib: rtc.GDevTrib{PDevTrib: 2.0, VDevTrib: 4.0},
				GRed:     rtc.GRed{PRedAliq: 3.0, PAliqEfet: 4.0},
				VIBSMun:  arredondar(vProd*0.002, 2),
			},
			GCBS: rtc.GCBSValores{
				PCBS:     0.9,
				GDif:     rtc.GDif{PDif: 1.0, VDif: 5.0},
				GDevTrib: rtc.GDevTrib{PDevTrib: 2.0, VDevTrib: 6.0},
				GRed:     rtc.GRed{PRedAliq: 3.0, PAliqEfet: 4.0},
				GALCZFMCBS: rtc.GALCZFMCBS{
					TpALCZFMCBS:     rtc.TpALCZFMCBSnOpInd,
					NProcSuframa:    "SUF-001",
					PAliqEfetRegCBS: 1.5,
					VTribRegCBS:     11.73,
				},
				VCBS: arredondar(vProd*0.009, 2),
			},
		},
		GTransfCred:   rtc.GTransfCred{VIBS: 1.0, VCBS: 2.0},
		GAjusteCompet: rtc.GAjusteCompet{CompetApur: dataBase, VIBS: 3.0, VCBS: 4.0},
		GEstornoCred:  rtc.GEstornoCred{VIBSEstCred: 5.0, VCBSEstCred: 6.0},
	}
}

// ibscbsTotExemplo monta o total de IBS/CBS do documento.
func ibscbsTotExemplo(vProd float64) rtc.IBSCBSTot {
	return rtc.IBSCBSTot{
		VBCIBSCBS: vProd,
		GIBS: rtc.GIBS{
			GIBSUFTot:        rtc.GIBSUFTot{VDif: 1.0, VDevTrib: 2.0, VIBSUF: arredondar(vProd*0.001, 2)},
			GIBSMunTot:       rtc.GIBSMunTot{VDif: 3.0, VDevTrib: 4.0, VIBSMun: arredondar(vProd*0.002, 2)},
			VIBS:             arredondar(vProd*0.085, 2),
			VCredPres:        1.0,
			VCredPresCondSus: 2.0,
		},
		GCBS: rtc.GCBS{
			VDif: 5.0, VDevTrib: 6.0, VCBS: arredondar(vProd*0.009, 2),
			VCredPres: 7.0, VCredPresCondSus: 8.0,
		},
		GMono: rtc.GMono{
			VIBSMono: 9.0, VCBSMono: 10.0,
			VIBSMonoReten: 11.0, VCBSMonoReten: 12.0,
			VIBSMonoRet: 13.0, VCBSMonoRet: 14.0,
		},
		GEstornoCred: rtc.GEstornoCred{VIBSEstCred: 15.0, VCBSEstCred: 16.0},
	}
}

// ---------------------------------------------------------------------------
// Protocolo e assinatura ficticios
// ---------------------------------------------------------------------------

// aplicarProtocolo preenche um protocolo de autorizacao ficticio. Exige que a
// chave ja tenha sido montada pelo writer.
func aplicarProtocolo(n *nfag.NFAg) {
	n.ProcNFAg = pcn.ProcDFe{
		TpAmb:    n.Ide.TpAmb,
		VerAplic: "OpenFiscalBR_DEMO_1.0",
		ChDFe:    n.ChaveAcesso(),
		DhRecbto: dataBase.Add(2 * time.Minute),
		NProt:    "135260000000001",
		DigVal:   digestFicticio,
		CStat:    100,
		XMotivo:  "Autorizado o uso da NFAg",
	}
}

// Valores de assinatura obviamente falsos: o documento serve para testar
// IMPORTACAO, nunca para transmitir. Quem for transmitir usa o cenario
// "transmissao" e assina com certificado de verdade via /api/assinar.
const (
	digestFicticio    = "RElHRVNUVkFMVUVERUVYRU1QTE8="
	assinaturaFalsa   = "QVNTSU5BVFVSQS1ERS1FWEVNUExPLU5BTy1WQUxJREE="
	certificadoeFalso = "Q0VSVElGSUNBRE8tREUtRVhFTVBMTy1OQU8tVkFMSURP"
)

// assinarFicticio preenche o bloco Signature com valores falsos, para que o
// XML gerado exercite a estrutura do XMLDSig (SignedInfo, Reference,
// Transforms, KeyInfo, X509Data) sem precisar de certificado.
func assinarFicticio(n *nfag.NFAg) {
	n.Signature = pcn.Signature{
		URI:             "#" + n.InfNFAg.ID,
		DigestValue:     digestFicticio,
		SignatureValue:  assinaturaFalsa,
		X509Certificate: certificadoeFalso,
	}
}
