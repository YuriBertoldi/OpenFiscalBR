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

	"github.com/openfiscalbr/openfiscalbr/packages/nfgas"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Catalogo de XMLs de exemplo da demo. Cada cenario e montado preenchendo as
// structs do package e chamando o writer REAL (nfgas.GerarXML e companhia) --
// nunca um XML estatico embutido. XML de arquivo nao exercita as divergencias
// deliberadas documentadas em packages/nfgas/xml_writer.go e apodrece em
// silencio quando o writer muda.
//
// Os dados sao ficticios e FIXOS: com ParametrosExemplo zerado, o mesmo
// cenario gera sempre o mesmo XML, byte a byte. E isso que permite ao usuario
// so clicar e gerar, e aos testes compararem o resultado.

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
			Nome:         "Transmissao - NFGas pronta para envio",
			Descricao:    "NFGas avulsa, sem assinatura e sem protocolo, como vai para a SEFAZ.",
			Tipo:         "documento",
			AcaoSugerida: "gerar",
			Arquivo:      "nfgas-transmissao.xml",
			Gerar:        exemploTransmissao,
		},
		{
			ID:           "autorizada",
			Nome:         "Autorizada - nfgasProc com protocolo",
			Descricao:    "Documento + protNFGas (cStat 100), que e o que o emitente arquiva e o importador recebe.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfgas-autorizada.xml",
			Observacao:   "protocolo e assinatura sao ficticios: serve para importar, nao para transmitir",
			Gerar:        exemploAutorizada,
		},
		{
			ID:           "cancelamento",
			Nome:         "Cancelamento - evento avulso",
			Descricao:    "eventoNFGas com evCancNFGas (tpEvento 110111), amarrado a chave do cenario de transmissao.",
			Tipo:         "evento",
			AcaoSugerida: "ler-evento",
			Arquivo:      "nfgas-evento-cancelamento.xml",
			Gerar:        exemploCancelamento,
		},
		{
			ID:           "cancelamento-proc",
			Nome:         "Cancelamento - procEvento com retorno",
			Descricao:    "procEventoNFGas: evento enviado + retEventoNFGas homologado (cStat 135).",
			Tipo:         "evento",
			AcaoSugerida: "ler-evento",
			Arquivo:      "nfgas-proc-evento-cancelamento.xml",
			Gerar:        exemploCancelamentoProc,
		},
		{
			ID:           "completo",
			Nome:         "Completo - todas as informacoes de tag",
			Descricao:    "Todo grupo opcional do leiaute preenchido: gCompraGov, gSub, gVolContrat, gMed, gTarif, gMedicao, IBSCBS, retTrib, TxReg, gProcRef, pgtoVinc, gFat, gAgencia, autXML, infAdic, gRespTec, protocolo e assinatura.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfgas-completa.xml",
			Observacao:   "nenhum documento tem TODAS as tags do writer ao mesmo tempo -- os grupos de ICMS se excluem por CST; aqui estao as de um documento real maximamente preenchido",
			Gerar:        exemploCompleto,
		},
		{
			ID:           "erro-leitura",
			Nome:         "Com erro - quebra na leitura",
			Descricao:    "XML sem o atributo versao em infNFGas: o leitor recusa com ErrAtributoVersaoAusente (HTTP 422).",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfgas-erro-leitura.xml",
			Observacao:   "erro proposital para testar o tratamento de documento invalido na importacao",
			Gerar:        exemploErroLeitura,
		},
		{
			ID:           "erro-regras",
			Nome:         "Com erro - reprova na validacao",
			Descricao:    "Le normalmente, mas viola as regras 226 (cMun x cUF), 227 (concatenacao da chave) e 247 (sigla da UF), e tem CNPJ de emitente com digito invalido.",
			Tipo:         "documento",
			AcaoSugerida: "validar",
			Arquivo:      "nfgas-erro-regras.xml",
			Observacao:   "o digito verificador da chave fica INTEGRO de proposito: chave valida com concatenacao divergente e o caso que mais escapa de um importador",
			Gerar:        exemploErroRegras,
		},
		{
			ID:           "multi-itens",
			Nome:         "Varios itens",
			Descricao:    "Cinco itens com CST de ICMS diferentes (00, 20, 40, 60, 90) e total coerente com a soma.",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfgas-varios-itens.xml",
			AceitaItens:  true,
			Gerar:        exemploMultiItens,
		},
		{
			ID:           "multi-cfop",
			Nome:         "Varios CFOP",
			Descricao:    "Quatro itens com CFOP distintos: 5253 (dentro do estado), 5257 (zona franca), 6253 (interestadual) e 5949 (outra saida).",
			Tipo:         "documento",
			AcaoSugerida: "ler",
			Arquivo:      "nfgas-varios-cfop.xml",
			Gerar:        exemploMultiCFOP,
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
func aplicarParametros(n *nfgas.NFGas, p ParametrosExemplo) error {
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

// notaBase monta a NFGas ficticia comum a todos os cenarios: emitente,
// destinatario, instalacao, um item com ICMS/PIS/COFINS e os totais. Quem
// precisa de mais grupos acrescenta por cima.
func notaBase() *nfgas.NFGas {
	n := &nfgas.NFGas{
		InfNFGas: nfgas.InfNFGas{Versao: 1.00},
		Ide: nfgas.Ide{
			CUF: 35, TpAmb: pcn.TaHomologacao, Modelo: 76, Serie: 1, NNF: 1,
			CNF: 1, DhEmi: dataBase, TpEmis: pcn.TeNormal,
			NSiteAutoriz: nfgas.Sa0, CMunFG: 3550308,
			FinNFGas: nfgas.FnNormal, TpFat: nfgas.TfNormal,
			VerProc: "OpenFiscalBR Demo 1.0",
		},
		Emit: nfgas.Emit{
			CNPJ:  "11222333000181",
			IE:    "111222333444",
			XNome: "Distribuidora de Gas Exemplo LTDA",
			XFant: "GasExemplo",
			EnderEmit: nfgas.Endereco{
				XLgr: "Rua das Tubulacoes", Nro: "1000", XBairro: "Industrial",
				CMun: 3550308, XMun: "Sao Paulo", CEP: 1310100, UF: "SP",
				Fone: "1130001000", Email: "fiscal@gasexemplo.com.br",
			},
		},
		Dest: nfgas.Dest{
			XNome:   "Consumidor Exemplo",
			CNPJCPF: "52998224725",
			IE:      "ISENTO",
			EnderDest: nfgas.Endereco{
				XLgr: "Avenida Central", Nro: "200", XCpl: "Apto 101",
				XBairro: "Centro", CMun: 3550308, XMun: "Sao Paulo",
				CEP: 1001000, UF: "SP", Fone: "1199990000",
				Email: "consumidor@exemplo.com.br",
			},
		},
		Instalacao: nfgas.Instalacao{
			IDInstalacao: "INST-000123", IDCodCliente: "CLI-987",
			TpInstalacao: nfgas.TiCativo, NContrato: "CT-2026-001",
			TpClasse: nfgas.TcResidencial, XClasse: "Residencial",
		},
		Det: []nfgas.Det{itemGas(1, 5253, pcn.CST00, 250.4444, 3.12)},
		GFat: nfgas.GFat{
			CompetFat: dataBase,
			DVencFat:  dataBase.AddDate(0, 0, 26),
			NFat:      "FAT-2026-03-0001",
			CodBarras: "84670000009487200011222333000181000000000001",
		},
	}
	recalcularTotal(n)
	return n
}

// itemGas monta um item de gas canalizado com ICMS do CST informado.
func itemGas(nItem, cfop int, cst pcn.CSTIcms, qtd, tarifa float64) nfgas.Det {
	vProd := arredondar(qtd*tarifa, 2)
	imposto := nfgas.Imposto{
		Orig: pcn.OeNacional,
		PIS: nfgas.PIS{
			CST: pcn.Pis01, VBC: vProd, PPIS: 1.65,
			VPIS: arredondar(vProd*0.0165, 2),
		},
		COFINS: nfgas.COFINS{
			CST: pcn.Cof01, VBC: vProd, PCOFINS: 7.60,
			VCOFINS: arredondar(vProd*0.076, 2),
		},
	}
	imposto.ICMS = icmsPorCST(cst, vProd)

	return nfgas.Det{
		NItem: nItem,
		GNormal: nfgas.GNormal{
			GTarif: []nfgas.GTarif{{
				DIniTarif:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				DFimTarif:   time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
				NAto:        "ATO-123",
				AnoAto:      2026,
				TpFaixaCons: nfgas.TfFixa,
				VTarifAplic: tarifa,
			}},
			Prod: nfgas.Prod{
				IndOrigemQtd: nfgas.IoMedido,
				CProd:        fmt.Sprintf("GAS%03d", nItem),
				XProd:        "Gas Natural Canalizado",
				CClass:       "0100101",
				CFOP:         cfop,
				UMed:         nfgas.Umim3,
				QFaturada:    qtd,
				VItem:        tarifa,
				VProd:        vProd,
			},
			Imposto: imposto,
		},
	}
}

// icmsPorCST monta o grupo de ICMS adequado ao CST. Os grupos ICMS00..ICMS90
// se excluem entre si -- um item so pode ter um deles --, e e por isso que
// documento nenhum exercita todas as tags de ICMS do writer de uma vez.
func icmsPorCST(cst pcn.CSTIcms, vProd float64) nfgas.ICMS {
	icms := nfgas.ICMS{CST: cst}
	switch cst {
	case pcn.CST00:
		icms.ModBC = nfgas.DbiValorOperacao
		icms.VBC = vProd
		icms.PICMS = 18
		icms.VICMS = arredondar(vProd*0.18, 2)
	case pcn.CST20:
		icms.ModBC = nfgas.DbiValorOperacao
		icms.PRedBC = 33.33
		icms.VBC = arredondar(vProd*0.6667, 2)
		icms.PICMS = 18
		icms.VICMS = arredondar(vProd*0.6667*0.18, 2)
		icms.VBCFCP = icms.VBC
		icms.PFCP = 2
		icms.VFCP = arredondar(icms.VBC*0.02, 2)
	case pcn.CST40:
		icms.VICMSDeson = arredondar(vProd*0.18, 2)
		icms.MotDesICMS = nfgas.MdiOutros
		icms.IndDeduzDeson = pcn.TieSim
	case pcn.CST60:
		icms.VBCSTRet = vProd
		icms.PICMSSTRet = 18
		icms.VICMSSTRet = arredondar(vProd*0.18, 2)
		icms.PRedBCEfet = 10
		icms.VBCEfet = arredondar(vProd*0.9, 2)
		icms.PICMSEfet = 18
		icms.VICMSEfet = arredondar(vProd*0.9*0.18, 2)
	case pcn.CST70:
		// Tributada com reducao e cobranca por ST, mais desoneracao parcial.
		icms.ModBC = nfgas.DbiValorOperacao
		icms.PRedBC = 20
		icms.VBC = arredondar(vProd*0.8, 2)
		icms.PICMS = 18
		icms.VICMS = arredondar(vProd*0.8*0.18, 2)
		icms.VBCFCP = icms.VBC
		icms.PFCP = 2
		icms.VFCP = arredondar(icms.VBC*0.02, 2)
		icms.ModBCST = nfgas.DbisMargemValorAgregado
		icms.PMVAST = 40
		icms.PRedBCST = 10
		icms.VBCST = arredondar(vProd*1.4*0.9, 2)
		icms.PICMSST = 18
		icms.VICMSST = arredondar(vProd*1.4*0.9*0.18, 2)
		icms.VBCFCPST = icms.VBCST
		icms.PFCPST = 2
		icms.VFCPST = arredondar(icms.VBCST*0.02, 2)
		icms.VICMSDeson = arredondar(vProd*0.05, 2)
		icms.MotDesICMS = nfgas.MdiOutros
		icms.IndDeduzDeson = pcn.TieSim
	case pcn.CST90:
		icms.ModBC = nfgas.DbiValorOperacao
		icms.VBC = vProd
		icms.PICMS = 12
		icms.VICMS = arredondar(vProd*0.12, 2)
	}
	return icms
}

// recalcularTotal soma os itens no grupo total, para que o documento gerado
// nunca saia com total divergente da soma -- o cenario de erro de regras e
// quem introduz divergencia, e de proposito.
//
// O ACBr NAO calcula totais: vNF so e lido, escrito e copiado
// (ACBrNFGas.Classes.pas:1735), nunca derivado -- preencher o grupo total e
// responsabilidade do emitente. Como estes exemplos existem para ser
// importados e conferidos, o vNF e composto aqui pelas parcelas que acrescem
// ao valor da nota (ST, FCP, FCPST, taxa de regulacao e o total dos itens
// agregadores), descontada a desoneracao quando indDeduzDeson manda deduzir.
// vTotDFe acompanha o vNF, como na fixture do package.
func recalcularTotal(n *nfgas.NFGas) {
	var t nfgas.Total
	var agregado float64
	for i := range n.Det {
		agregado += n.Det[i].GAgregadora.VTotDFe

		p := n.Det[i].GNormal.Prod
		imp := n.Det[i].GNormal.Imposto
		t.VProd += p.VProd
		t.VBC += imp.ICMS.VBC
		t.VICMS += imp.ICMS.VICMS
		t.VICMSDeson += imp.ICMS.VICMSDeson
		t.VFCP += imp.ICMS.VFCP
		t.VBCST += imp.ICMS.VBCST
		t.VST += imp.ICMS.VICMSST
		t.VFCPST += imp.ICMS.VFCPST
		t.VPIS += imp.PIS.VPIS
		t.VCOFINS += imp.COFINS.VCOFINS
		t.VTxReg += imp.TxReg.VTaxa
		t.VRetPIS += imp.RetTrib.VRetPIS
		t.VRetCOFINS += imp.RetTrib.VRetCOFINS
		t.VRetCSLL += imp.RetTrib.VRetCSLL
		t.VIRRF += imp.RetTrib.VIRRF
	}
	t.VProd = arredondar(t.VProd, 2)

	// Desoneracao so abate o total quando o item mandou deduzir.
	var deson float64
	for i := range n.Det {
		icms := n.Det[i].GNormal.Imposto.ICMS
		if icms.IndDeduzDeson == pcn.TieSim {
			deson += icms.VICMSDeson
		}
	}

	t.VNF = arredondar(t.VProd+t.VST+t.VFCP+t.VFCPST+t.VTxReg+agregado-deson, 2)
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
	return nfgas.GerarXML(n)
}

func exemploAutorizada(p ParametrosExemplo) (string, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	// A chave do protocolo so existe depois que o writer a monta; gerar duas
	// vezes e o caminho mais honesto -- a primeira passada calcula a chave.
	if _, err := nfgas.GerarXML(n); err != nil {
		return "", err
	}
	aplicarProtocolo(n)
	assinarFicticio(n)
	return nfgas.GerarXMLProc(n)
}

func exemploCancelamento(p ParametrosExemplo) (string, error) {
	ev, err := montarEventoCancelamento(p)
	if err != nil {
		return "", err
	}
	return nfgas.GerarXMLEvento(ev)
}

func exemploCancelamentoProc(p ParametrosExemplo) (string, error) {
	ev, err := montarEventoCancelamento(p)
	if err != nil {
		return "", err
	}
	ret := &nfgas.RetEventoNFGas{
		Versao:    "1.00",
		TemEvento: true,
		Evento:    *ev,
		RetInfEvento: nfgas.RetInfEvento{
			TpAmb:       ev.InfEvento.TpAmb,
			VerAplic:    "OpenFiscalBR_DEMO_1.0",
			COrgao:      ev.InfEvento.COrgaoEfetivo(),
			CStat:       135,
			XMotivo:     "Evento registrado e vinculado a NFGas",
			ChNFGas:     ev.InfEvento.ChNFGas,
			TpEvento:    nfgas.TeCancelamento,
			XEvento:     "Cancelamento",
			NSeqEvento:  ev.InfEvento.NSeqEvento,
			CNPJDest:    "99888777000166",
			EmailDest:   "consumidor@exemplo.com.br",
			COrgaoAutor: ev.InfEvento.COrgaoEfetivo(),
			DhRegEvento: dataBase.AddDate(0, 0, 1).Add(5 * time.Minute),
			NProt:       "135260000000099",
		},
	}
	return nfgas.GerarXMLProcEvento(ret)
}

// montarEventoCancelamento amarra o evento a chave do cenario de transmissao
// -- evento apontando para chave que o usuario nao tem nao serve de teste.
func montarEventoCancelamento(p ParametrosExemplo) (*nfgas.EventoNFGas, error) {
	n := notaBase()
	if err := aplicarParametros(n, p); err != nil {
		return nil, err
	}
	if _, err := nfgas.GerarXML(n); err != nil {
		return nil, err
	}
	return &nfgas.EventoNFGas{
		Versao: "1.00",
		InfEvento: nfgas.InfEvento{
			COrgao:     n.Ide.CUF,
			TpAmb:      n.Ide.TpAmb,
			CNPJ:       n.Emit.CNPJ,
			ChNFGas:    n.ChaveAcesso(),
			DhEvento:   dataBase.AddDate(0, 0, 1),
			TpEvento:   nfgas.TeCancelamento,
			NSeqEvento: 1,
			DetEvento: nfgas.DetEvento{
				NProt: "135260000000001",
				XJust: "Erro na medicao do consumo do periodo faturado",
			},
		},
	}, nil
}

func exemploErroLeitura(p ParametrosExemplo) (string, error) {
	xml, err := exemploTransmissao(p)
	if err != nil {
		return "", err
	}
	// Remover o atributo versao de infNFGas e o erro mais limpo que se pode
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

	xml, err := nfgas.GerarXML(n)
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

	// Os CST giram para que o documento exercite grupos de ICMS diferentes:
	// ICMS00, ICMS20, ICMS40, ICMS60 e ICMS90 num documento so.
	csts := []pcn.CSTIcms{pcn.CST00, pcn.CST20, pcn.CST40, pcn.CST60, pcn.CST90}

	n := notaBase()
	n.Det = nil
	for i := 0; i < qtd; i++ {
		item := itemGas(i+1, 5253, csts[i%len(csts)], float64(100+i*25), 3.12+float64(i)*0.1)
		item.GNormal.Prod.XProd = fmt.Sprintf("Gas Natural Canalizado - faixa %d", i+1)
		n.Det = append(n.Det, item)
	}
	recalcularTotal(n)

	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	return nfgas.GerarXML(n)
}

func exemploMultiCFOP(p ParametrosExemplo) (string, error) {
	cfops := []struct {
		CFOP  int
		XProd string
	}{
		{5253, "Gas canalizado - consumidor do mesmo estado"},
		{5257, "Gas canalizado - consumidor em zona franca"},
		{6253, "Gas canalizado - consumidor de outro estado"},
		{5949, "Taxa de disponibilidade"},
	}

	n := notaBase()
	n.Det = nil
	for i, c := range cfops {
		item := itemGas(i+1, c.CFOP, pcn.CST00, float64(50+i*20), 3.12)
		item.GNormal.Prod.XProd = c.XProd
		n.Det = append(n.Det, item)
	}
	recalcularTotal(n)

	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}
	return nfgas.GerarXML(n)
}

// ---------------------------------------------------------------------------
// Cenario completo
// ---------------------------------------------------------------------------

// exemploCompleto preenche TODO grupo opcional do leiaute que pode conviver
// num mesmo documento. O que nao cabe aqui sao os grupos de ICMS alternativos
// (um item so tem um CST) -- para esses, veja o cenario multi-itens.
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
			{RefDFeAnt: "35260111222333000181760010000000001100000003"},
			{RefDFeAnt: "35260111222333000181760010000000002100000002"},
		},
	}

	n.Emit.ISUFEmit = "ISUF123"
	n.Emit.EnderEmit.CPais = 1058
	n.Emit.EnderEmit.XPais = "BRASIL"
	n.Dest.IM = "IM123"
	// cNIS exclui NB no writer (xml_writer.go:250): preencher os dois faria
	// o NB sumir em silencio.
	n.Dest.CNIS = "12345678901"
	n.Dest.XNomeAdicional = "Apto 101"
	n.Dest.EnderDest.CPais = 1058
	n.Dest.EnderDest.XPais = "BRASIL"

	n.Instalacao.LatGPS = "-23.550520"
	n.Instalacao.LongGPS = "-46.633308"
	n.Instalacao.CodRoteiroLeitura = "ROT-A12"

	// gSub: chNFGas e gNF sao MUTUAMENTE EXCLUSIVOS no writer
	// (xml_writer.go:297). Usamos o gNF, que carrega seis tags contra uma.
	n.GSub = nfgas.GSub{
		GNF: nfgas.GNF{
			CNPJ: "11222333000181", Serie: "001", NNF: 5,
			CompetEmis: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			CompetApur: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Hash115:    "HASH115EXEMPLO",
		},
		MotSub: nfgas.MsErroLeitura,
	}

	n.GVolContrat = []nfgas.GVolContrat{
		{NContrat: 1, TpVolContrat: nfgas.VcVolumeContratado, QUnidContrat: 1500.123456},
		{NContrat: 2, TpVolContrat: nfgas.VcDemandaMinima, QUnidContrat: 250.5},
	}

	n.GMed = []nfgas.GMed{{
		NMed: 1, IDEqp: "MED-0001",
		DMedAnt: dataBase.AddDate(0, -1, 0), VMedAnt: 1000.1234,
		DMedAtu: dataBase, VMedAtu: 1250.5678,
		TpEqp: nfgas.TeMedidor, TpMedidor: nfgas.TmDiafragma,
	}}

	// Item 1: o item cheio -- medicao, segunda faixa tarifaria, pagamento
	// antecipado, IBS/CBS, tributos retidos, taxa de regulacao, processo
	// referenciado e informacao adicional do produto.
	//
	// O CST 70 nao e escolha estetica: ICMS70 e o grupo mais rico do writer
	// (xml_writer.go:558) -- e o unico que emite modBC, vBCFCP, modBCST,
	// pMVAST, pRedBCST, o bloco de ST inteiro, motDesICMS e indDeduzDeson.
	det := itemGas(1, 5253, pcn.CST70, 250.4444, 3.12)
	prod := &det.GNormal.Prod
	// Sem tpMotNaoLeitura/xMotNaoLeitura: elas so sairiam no ramo sem
	// leitura, que e INALCANCAVEL neste writer e no proprio ACBr -- ver a
	// entrada correspondente em tagsForaDoWriter (exemplos_test.go).
	prod.GMedicao = nfgas.GMedicao{
		NMed: 1, NContrat: 1,
		GMedida: nfgas.GMedida{UMed: nfgas.Umm3, VMed: 250.4444},
	}
	prod.FatorPCS = 1.01
	prod.FatorPTZ = 0.99
	prod.FatorP = 1.0
	prod.FatorT = 1.0
	prod.IndDevolucao = pcn.TieNao
	prod.GPagAntecipado = rtc.GPagAntecipadoProd{
		ChDFePagAnt: "35260111222333000181760010000000001100000003",
		NItemPagAnt: 1,
	}
	det.GNormal.GTarif = append(det.GNormal.GTarif, nfgas.GTarif{
		DIniTarif:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		NAto:        "ATO-124",
		AnoAto:      2026,
		TpFaixaCons: nfgas.TfMedia,
		VTarifAplic: 4.0,
	})

	imp := &det.GNormal.Imposto
	imp.RetTrib = nfgas.RetTrib{
		VRetPIS: 1.11, VRetCOFINS: 2.22, VRetCSLL: 3.33,
		VBCIRRF: 100.00, VIRRF: 4.44,
	}
	imp.TxReg = nfgas.TxReg{VBC: prod.VProd, PTaxa: 0.5, VTaxa: arredondar(prod.VProd*0.005, 2)}
	imp.IBSCBS = ibscbsExemplo(prod.VProd)

	det.GNormal.GProcRef = nfgas.GProcRef{
		VItem: 1.23456789, QFaturada: 10.5, VProd: 12.96,
		IndDevolucao: pcn.TieSim,
		GProc: []nfgas.GProc{
			{TpProc: nfgas.TpJusticaEstadual, NProcesso: "0001234-55.2026.8.26.0100"},
			{TpProc: nfgas.TpProcon, NProcesso: "PROCON-2026-9"},
		},
	}
	det.GNormal.InfAdProd = "Consumo referente a marco/2026"

	// Item 2: referencia item de documento anterior (atributos chNFGasAnt e
	// nItemAnt do det) e usa a unidade "Unidade". Leva o CST 40 porque so os
	// grupos ICMS20/40/51 emitem cBenef -- no ICMS70 ele e suprimido
	// (divergencia 3 do writer).
	det2 := itemGas(2, 5949, pcn.CST40, 1, 50)
	det2.ChNFGasAnt = "35260211222333000181760010000000001100000005"
	det2.NItemAnt = 1
	det2.GNormal.GTarif = nil
	det2.GNormal.Prod.IndOrigemQtd = nfgas.IoCalculada
	det2.GNormal.Prod.CProd = "SERV001"
	det2.GNormal.Prod.XProd = "Taxa de disponibilidade"
	det2.GNormal.Prod.CClass = "0900101"
	det2.GNormal.Prod.UMed = nfgas.UmiUnidade
	det2.GNormal.Imposto.ICMS.CBenef = "SP000001"

	// Item 3: agregador. gNormal e gAgregadora sao MUTUAMENTE EXCLUSIVOS no
	// writer (xml_writer.go:342): gAgregadora preenchida descarta o gNormal
	// inteiro daquele item. Por isso o agregador vai num item proprio, em
	// vez de conviver com o item cheio.
	det3 := nfgas.Det{
		NItem:       3,
		GAgregadora: nfgas.GAgregadora{CClass: "0100101", VTotDFe: 100},
	}

	// Item 4: item sem CST de ICMS. indSemCST substitui orig e o grupo de
	// ICMS inteiro (xml_writer.go:426), entao precisa de item proprio.
	det4 := itemGas(4, 5253, pcn.CSTVazio, 5, 2)
	det4.GNormal.GTarif = nil
	det4.GNormal.Prod.CProd = "SERV002"
	det4.GNormal.Prod.XProd = "Item nao tributado pelo ICMS"
	det4.GNormal.Imposto.IndSemCST = pcn.TieSim

	// Itens 5 e 6 existem para exercitar os grupos ICMS00 e ICMS20, que os
	// CST dos itens anteriores excluem. Um item so pode ter um grupo.
	det5 := itemGas(5, 5253, pcn.CST00, 20, 3.12)
	det5.GNormal.GTarif = nil
	det5.GNormal.Prod.XProd = "Gas canalizado - tributacao integral"
	det6 := itemGas(6, 5253, pcn.CST20, 15, 3.12)
	det6.GNormal.GTarif = nil
	det6.GNormal.Prod.XProd = "Gas canalizado - base de calculo reduzida"
	det6.GNormal.Imposto.ICMS.VICMSDeson = 1.5
	det6.GNormal.Imposto.ICMS.CBenef = "SP000002"

	n.Det = []nfgas.Det{det, det2, det3, det4, det5, det6}
	recalcularTotal(n)
	n.Total.IBSCBSTot = ibscbsTotExemplo(n.Total.VProd)

	n.PgtoVinc = rtc.PgtoVinc{Pgto: []rtc.Pgto{
		{NPag: 1, IDTransacao: "TX-0001", TpMeioPgto: "03",
			CNPJReceb: "11222333000181", CNPJBasePSP: "11222333"},
		{NPag: 2, IDTransacao: "TX-0002", TpMeioPgto: "17"},
	}}

	n.GFat.DApresFat = dataBase.AddDate(0, 0, 5)
	n.GFat.DProxLeitura = dataBase.AddDate(0, 1, 0)
	// codDebAuto exclui codBanco+codAgencia (xml_writer.go:740); ficamos com
	// o par, que cobre duas tags.
	n.GFat.CodBanco = "001"
	n.GFat.CodAgencia = "1234"
	n.GFat.InfAdFat = "Pagamento ate o vencimento"
	n.GFat.GPIX = nfgas.GPIX{URLQRCodePIX: "https://pix.exemplo.com.br/qr/0001"}
	n.GFat.EnderCorresp = nfgas.Endereco{
		XLgr: "Rua da Correspondencia", Nro: "50", XCpl: "Sala 3",
		XBairro: "Jardins", CMun: 3550308, XMun: "Sao Paulo",
		CEP: 1400000, UF: "SP", Fone: "1155554444",
		Email: "correspondencia@exemplo.com.br",
	}

	n.GAgencia = nfgas.GAgencia{
		NomeAgenciaAtend:  "Agencia Central",
		EnderAgenciaAtend: "Rua do Atendimento, 10",
		SitioAgenciaAtend: "https://www.gasexemplo.com.br",
		InfAdReg:          "Regulado pela ARSESP",
		GHistCons: []nfgas.GHistCons{{
			XHistorico: "Consumo dos ultimos meses",
			MedMensal:  245.0,
			GCons: []nfgas.GCons{
				{CompetFat: dataBase.AddDate(0, -1, 0), UMed: nfgas.Umm3,
					QtdDias: 28, MedDiaria: 8.9286, Consumo: 250, VFat: 900},
				{CompetFat: dataBase.AddDate(0, -2, 0), UMed: nfgas.Umm3,
					QtdDias: 31, MedDiaria: 7.7419, Consumo: 240, VFat: 870},
			},
		}},
	}

	n.AutXML = []nfgas.AutXML{
		{CNPJCPF: "99888777000166"},
		{CNPJCPF: "52998224725"},
	}

	n.InfAdic = nfgas.InfAdic{
		InfAdFisco: "Documento de exemplo gerado pela demo do OpenFiscalBR",
		InfCpl: []string{
			"Informacao complementar 1",
			"Informacao complementar 2",
		},
	}

	n.InfRespTec = nfgas.InfRespTec{
		CNPJ: "99888777000166", XContato: "Responsavel Tecnico",
		Email: "ti@gasexemplo.com.br", Fone: "1140004000",
		IDCSRT: 1, HashCSRT: "SGVsbG9Xb3JsZEhhc2hDU1JU",
	}

	if err := aplicarParametros(n, p); err != nil {
		return "", err
	}

	// Primeira passada para o writer montar a chave; so entao da para
	// preencher o QR-Code e o protocolo, que dependem dela.
	if _, err := nfgas.GerarXML(n); err != nil {
		return "", err
	}
	n.InfNFGasSupl.QrCodNFGas = fmt.Sprintf(
		"https://www.nfgas.fazenda.sp.gov.br/consulta?chNFGas=%s&tpAmb=%s",
		n.ChaveAcesso(), n.Ide.TpAmb.String())
	aplicarProtocolo(n)
	assinarFicticio(n)

	return nfgas.GerarXML(n)
}

// ibscbsExemplo monta o grupo de IBS/CBS do item com todos os subgrupos que
// convivem: diferimento, devolucao, reducao, ZFM, credito presumido, estorno,
// transferencia e ajuste de competencia.
func ibscbsExemplo(vProd float64) rtc.IBSCBS {
	return rtc.IBSCBS{
		// CST 000 e obrigatorio para que o gIBSCBS saia: no modelo NFGas o
		// writer do rtc so gera o grupo quando o CST e 000
		// (packages/rtc/xml_writer.go:263).
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
func aplicarProtocolo(n *nfgas.NFGas) {
	n.ProcNFGas = pcn.ProcDFe{
		TpAmb:    n.Ide.TpAmb,
		VerAplic: "OpenFiscalBR_DEMO_1.0",
		ChDFe:    n.ChaveAcesso(),
		DhRecbto: dataBase.Add(2 * time.Minute),
		NProt:    "135260000000001",
		DigVal:   digestFicticio,
		CStat:    100,
		XMotivo:  "Autorizado o uso da NFGas",
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
func assinarFicticio(n *nfgas.NFGas) {
	n.Signature = pcn.Signature{
		URI:             "#" + n.InfNFGas.ID,
		DigestValue:     digestFicticio,
		SignatureValue:  assinaturaFalsa,
		X509Certificate: certificadoeFalso,
	}
}
