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
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do leitor do retorno dos eventos.

// montarRetEnvEvento monta um retorno com um retEvento do cStat pedido.
func montarRetEnvEvento(cStatLote, cStatEvento int, extra string) string {
	return `<retEnvEvento versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<idLote>1</idLote><tpAmb>2</tpAmb><verAplic>AN_1.0</verAplic>` +
		`<cOrgao>91</cOrgao><cStat>` + itoa(cStatLote) + `</cStat>` +
		`<xMotivo>Lote de evento processado</xMotivo>` +
		`<retEvento versao="1.00"><infEvento Id="ID1101113126019999">` +
		`<tpAmb>2</tpAmb><verAplic>AN_1.0</verAplic><cOrgao>91</cOrgao>` +
		`<cStat>` + itoa(cStatEvento) + `</cStat>` +
		`<xMotivo>Evento registrado e vinculado a NF-e</xMotivo>` +
		`<chNFe>` + chaveDeMinas + `</chNFe>` +
		`<tpEvento>110111</tpEvento><xEvento>Cancelamento</xEvento>` +
		`<nSeqEvento>1</nSeqEvento>` + extra +
		`<dhRegEvento>2026-01-20T18:45:00-03:00</dhRegEvento>` +
		`<nProt>131260000000001</nProt>` +
		`</infEvento></retEvento></retEnvEvento>`
}

// TestLerRetEventoNFe_LeORetornoDoLote trava a leitura do cabecalho e do
// retEvento.
func TestLerRetEventoNFe_LeORetornoDoLote(t *testing.T) {
	r, err := LerRetEventoNFe(montarRetEnvEvento(128, 135, ""))
	if err != nil {
		t.Fatal(err)
	}

	if !r.LoteProcessado() {
		t.Errorf("cStat %d deveria ser lote processado", r.CStat)
	}
	if r.COrgao != 91 {
		t.Errorf("cOrgao = %d, esperado 91", r.COrgao)
	}
	if len(r.RetEvento) != 1 {
		t.Fatalf("retEvento lidos = %d, esperado 1", len(r.RetEvento))
	}

	ret := r.RetEvento[0]
	if ret.TpEvento != pcn.TeCancelamento {
		t.Errorf("tpEvento = %s, esperado Cancelamento", ret.TpEvento)
	}
	if ret.NProt != "131260000000001" {
		t.Errorf("nProt = %q", ret.NProt)
	}
	if ret.DhRegEvento.IsZero() {
		t.Error("dhRegEvento nao foi lido")
	}
	if !strings.HasPrefix(ret.XML, "<infEvento") {
		t.Errorf("XML cru nao e o elemento inteiro: %q", ret.XML)
	}
}

// TestRetInfEvento_O136NaoVinculaANota trava a diferenca entre registrado e
// vinculado.
//
// O 136 e "evento registrado mas NAO vinculado a NF-e": a SEFAZ aceitou o
// evento embora a nota ainda nao exista ali. Quem for montar o
// procEventoNFe para arquivar junto da nota precisa saber disso -- tratar
// os tres codigos como iguais arquiva um processado que nao existe.
func TestRetInfEvento_O136NaoVinculaANota(t *testing.T) {
	casos := []struct {
		cStat      int
		registrado bool
		vinculado  bool
	}{
		{135, true, true},
		{136, true, false},
		{155, true, true},
		{573, false, false},
		{0, false, false},
	}
	for _, c := range casos {
		r := &RetInfEvento{CStat: c.cStat}
		if r.Registrado() != c.registrado {
			t.Errorf("cStat %d: Registrado() = %v, esperado %v",
				c.cStat, r.Registrado(), c.registrado)
		}
		if r.VinculadoANota() != c.vinculado {
			t.Errorf("cStat %d: VinculadoANota() = %v, esperado %v",
				c.cStat, r.VinculadoANota(), c.vinculado)
		}
	}
}

// TestLerRetEventoNFe_CNPJDestCaiParaCPFDest trava a precedencia de :656.
//
// O destinatario de uma manifestacao pode ser pessoa fisica, e nesse caso a
// SEFAZ devolve CPFDest. Ler so o CNPJDest devolve vazio.
func TestLerRetEventoNFe_CNPJDestCaiParaCPFDest(t *testing.T) {
	comCNPJ, err := LerRetEventoNFe(montarRetEnvEvento(128, 135,
		`<CNPJDest>99999999000191</CNPJDest>`))
	if err != nil {
		t.Fatal(err)
	}
	if got := comCNPJ.RetEvento[0].CNPJDest; got != "99999999000191" {
		t.Errorf("CNPJDest = %q", got)
	}

	comCPF, err := LerRetEventoNFe(montarRetEnvEvento(128, 135,
		`<CPFDest>11122233396</CPFDest>`))
	if err != nil {
		t.Fatal(err)
	}
	if got := comCPF.RetEvento[0].CNPJDest; got != "11122233396" {
		t.Errorf("sem CNPJDest o campo deveria trazer o CPFDest, trouxe %q", got)
	}
}

// TestLerRetEventoNFe_ChNFePendTrazAChave trava a correcao de um DEFEITO do
// ACBr.
//
// O Ler_chNFePend (:620) procura uma tag `chNFePend` DENTRO do proprio
// elemento chNFePend, que no leiaute e um `TChNFe` -- tipo simples, com a
// chave como texto. A busca nunca acha nada e o campo sai sempre vazio.
// Pior: o item e gravado em `Items[Idx]` usando o indice do LACO de
// chNFePend, que nao e o indice do retEvento.
//
// Replicar um defeito que so produz vazio seria pior que corrigi-lo: a
// lista de EPEC pendentes e a unica informacao util desse retorno.
func TestLerRetEventoNFe_ChNFePendTrazAChave(t *testing.T) {
	r, err := LerRetEventoNFe(montarRetEnvEvento(128, 135,
		`<chNFePend>`+chaveDeMinas+`</chNFePend>`+
			`<chNFePend>`+chaveNFCeDeMinas+`</chNFePend>`))
	if err != nil {
		t.Fatal(err)
	}

	pend := r.RetEvento[0].ChNFePend
	if len(pend) != 2 {
		t.Fatalf("chaves pendentes lidas = %d, esperado 2", len(pend))
	}
	if pend[0] != chaveDeMinas || pend[1] != chaveNFCeDeMinas {
		t.Errorf("chaves lidas = %v", pend)
	}
	// E elas pertencem ao retEvento certo -- com um so retEvento, qualquer
	// indexacao acerta; o que o teste trava e o CONTEUDO.
	if r.RetEvento[0].ChNFe != chaveDeMinas {
		t.Errorf("o retEvento perdeu a propria chave: %q", r.RetEvento[0].ChNFe)
	}
}

// TestLerRetEventoNFe_CincoFormasDeRaiz trava o LerXml de :728-748.
//
// Ler todas e o que permite reler do disco o que se gravou. A quinta -- o
// NFeDFe -- precisa de ramo proprio porque nela o procEventoNFe aparece
// DUAS vezes aninhado, e a busca rasa nao alcanca o <evento>.
func TestLerRetEventoNFe_CincoFormasDeRaiz(t *testing.T) {
	const infEvento = `<infEvento Id="ID1101113126019999">` +
		`<cOrgao>31</cOrgao><tpAmb>2</tpAmb><CNPJ>99999999000191</CNPJ>` +
		`<chNFe>` + chaveDeMinas + `</chNFe>` +
		`<dhEvento>2026-01-20T18:45:00-03:00</dhEvento>` +
		`<tpEvento>110111</tpEvento><nSeqEvento>1</nSeqEvento>` +
		`<verEvento>1.00</verEvento>` +
		`<detEvento versao="1.00"><descEvento>Cancelamento</descEvento>` +
		`<nProt>131260000000001</nProt><xJust>Erro na emissao</xJust>` +
		`</detEvento></infEvento>`
	const evento = `<evento versao="1.00">` + infEvento + `</evento>`

	casos := []struct {
		nome         string
		xml          string
		temInfEvento bool
		temRetEvento bool
	}{
		{"retEnvEvento", montarRetEnvEvento(128, 135, ""), false, true},
		{"evento solto", `<evento versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
			infEvento + `</evento>`, true, false},
		{"envEvento", `<envEvento versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
			`<idLote>1</idLote>` + evento + `</envEvento>`, true, false},
		{"procEventoNFe", `<procEventoNFe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
			evento + `<retEvento versao="1.00"><infEvento><cStat>135</cStat>` +
			`<chNFe>` + chaveDeMinas + `</chNFe><nProt>131260000000001</nProt>` +
			`</infEvento></retEvento></procEventoNFe>`, true, true},
		// O NFeDFe aninha o procEventoNFe DUAS vezes; sem o ramo proprio o
		// <evento> fica fora do alcance da busca rasa.
		{"NFeDFe", `<NFeDFe xmlns="http://www.portalfiscal.inf.br/nfe">` +
			`<procEventoNFe><procEventoNFe versao="1.00">` + evento +
			`</procEventoNFe></procEventoNFe>` +
			`<retEvento versao="1.00"><infEvento><cStat>135</cStat>` +
			`<chNFe>` + chaveDeMinas + `</chNFe></infEvento></retEvento>` +
			`</NFeDFe>`, true, true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			r, err := LerRetEventoNFe(c.xml)
			if err != nil {
				t.Fatal(err)
			}
			if (r.InfEvento != nil) != c.temInfEvento {
				t.Errorf("infEvento lido = %v, esperado %v",
					r.InfEvento != nil, c.temInfEvento)
			}
			if (len(r.RetEvento) > 0) != c.temRetEvento {
				t.Errorf("retEvento lidos = %d, esperado presenca %v",
					len(r.RetEvento), c.temRetEvento)
			}
			if c.temInfEvento {
				if r.InfEvento.TpEvento != pcn.TeCancelamento {
					t.Errorf("tpEvento = %s", r.InfEvento.TpEvento)
				}
				if r.InfEvento.DetEvento.NProt != "131260000000001" {
					t.Errorf("nProt do detEvento = %q", r.InfEvento.DetEvento.NProt)
				}
			}
		})
	}
}

// TestLerDetEvento_GruposDeMESMONOMEVaoParaColecoesDIFERENTES trava o
// `case` de :508 -- o lado do leitor da mesma armadilha do gerador.
//
// Os dois XML abaixo tem o MESMO elemento <gConsumo>; o que decide em qual
// colecao ele entra e o tpEvento. Um leitor que buscasse so pelo nome
// acertaria metade dos casos, e a metade errada sairia zerada.
func TestLerDetEvento_GruposDeMESMONOMEVaoParaColecoesDIFERENTES(t *testing.T) {
	// O tpEvento entra pelo Codigo() do proprio enum: errar o literal de
	// seis digitos faria o teste exercitar outro ramo do `case` sem avisar,
	// e a correspondencia codigo/enum ja tem teste proprio no pcn.
	montar := func(tipo pcn.TipoEvento, grupo string) string {
		tpEvento := tipo.Codigo()
		return `<evento versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
			`<infEvento Id="X"><cOrgao>31</cOrgao><tpAmb>2</tpAmb>` +
			`<CNPJ>99999999000191</CNPJ><chNFe>` + chaveDeMinas + `</chNFe>` +
			`<dhEvento>2026-01-20T18:45:00-03:00</dhEvento>` +
			`<tpEvento>` + tpEvento + `</tpEvento><nSeqEvento>1</nSeqEvento>` +
			`<verEvento>1.00</verEvento><detEvento versao="1.00">` +
			`<descEvento>x</descEvento><cOrgaoAutor>31</cOrgaoAutor>` +
			`<tpAutor>1</tpAutor><verAplic>1.0</verAplic>` + grupo +
			`</detEvento></infEvento></evento>`
	}

	t.Run("gConsumo do consumo pessoal", func(t *testing.T) {
		r, err := LerRetEventoNFe(montar(pcn.TeDestItemConsPessoal,
			`<gConsumo nItem="1"><vIBS>10.00</vIBS><vCBS>20.00</vCBS>`+
				`<gControleEstoque><qConsumo>3.0000</qConsumo>`+
				`<uConsumo>UN</uConsumo></gControleEstoque>`+
				`<DFeReferenciado><chaveAcesso>`+chaveDeMinas+
				`</chaveAcesso><nItem>7</nItem></DFeReferenciado></gConsumo>`))
		if err != nil {
			t.Fatal(err)
		}
		d := r.InfEvento.DetEvento
		if len(d.GConsumo) != 1 {
			t.Fatalf("GConsumo = %d, esperado 1", len(d.GConsumo))
		}
		if len(d.GConsumoZFM) != 0 {
			t.Errorf("o grupo foi para a colecao da ZFM: %d", len(d.GConsumoZFM))
		}
		g := d.GConsumo[0]
		if g.ControleEstoque.QConsumo != 3 || g.ControleEstoque.UConsumo != "UN" {
			t.Errorf("gControleEstoque = %+v", g.ControleEstoque)
		}
		if g.DFeReferenciado.NItem != 7 {
			t.Errorf("DFeReferenciado = %+v", g.DFeReferenciado)
		}
	})

	t.Run("gConsumo da ALC/ZFM", func(t *testing.T) {
		r, err := LerRetEventoNFe(montar(pcn.TeImporALCZFM,
			`<gConsumo nItem="1"><vIBS>10.00</vIBS><vCBS>20.00</vCBS>`+
				`<gControleEstoque><qtde>3.0000</qtde>`+
				`<unidade>UN</unidade></gControleEstoque></gConsumo>`))
		if err != nil {
			t.Fatal(err)
		}
		d := r.InfEvento.DetEvento
		if len(d.GConsumoZFM) != 1 {
			t.Fatalf("GConsumoZFM = %d, esperado 1", len(d.GConsumoZFM))
		}
		if len(d.GConsumo) != 0 {
			t.Errorf("o grupo foi para a colecao do consumo pessoal: %d",
				len(d.GConsumo))
		}
		g := d.GConsumoZFM[0]
		if g.ControleEstoque.Qtde != 3 || g.ControleEstoque.Unidade != "UN" {
			t.Errorf("gControleEstoque = %+v", g.ControleEstoque)
		}
	})

	t.Run("gPerecimento do fornecedor traz vIBS de DENTRO", func(t *testing.T) {
		r, err := LerRetEventoNFe(montar(pcn.TePerecPerdaRouboFurtoTranspContratFornec,
			`<gPerecimento nItem="1"><vIBS>100.00</vIBS><vCBS>200.00</vCBS>`+
				`<gControleEstoque><qPerecimento>3.0000</qPerecimento>`+
				`<uPerecimento>UN</uPerecimento>`+
				`<vIBS>11.00</vIBS><vCBS>22.00</vCBS>`+
				`</gControleEstoque></gPerecimento>`))
		if err != nil {
			t.Fatal(err)
		}
		d := r.InfEvento.DetEvento
		if len(d.GPerecimentoForn) != 1 {
			t.Fatalf("GPerecimentoForn = %d, esperado 1", len(d.GPerecimentoForn))
		}
		g := d.GPerecimentoForn[0]
		// Os nomes vIBS e vCBS se repetem nos dois niveis com valores
		// diferentes; ler do nivel errado da 100/200 no lugar de 11/22.
		if g.VIBS != 100 || g.VCBS != 200 {
			t.Errorf("valores do grupo pai = %v / %v, esperado 100 / 200", g.VIBS, g.VCBS)
		}
		if g.ControleEstoque.VIBS != 11 || g.ControleEstoque.VCBS != 22 {
			t.Errorf("valores do gControleEstoque = %v / %v, esperado 11 / 22",
				g.ControleEstoque.VIBS, g.ControleEstoque.VCBS)
		}
	})
}

// TestGerarXMLProcEvento_ColaEventoERetorno trava a montagem do documento
// de arquivo.
func TestGerarXMLProcEvento_ColaEventoERetorno(t *testing.T) {
	const assinado = `<evento versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<infEvento Id="ID1"><chNFe>x</chNFe></infEvento>` +
		`<Signature><SignatureValue>AAA</SignatureValue></Signature></evento>`

	ret := RetInfEvento{XML: `<infEvento Id="ID1"><cStat>135</cStat></infEvento>`}
	proc := GerarXMLProcEvento(assinado, ret, "1.00")

	if !strings.HasPrefix(proc, `<procEventoNFe versao="1.00"`) {
		t.Errorf("raiz errada: %s", proc)
	}
	posEvento := strings.Index(proc, "<evento ")
	posRetorno := strings.Index(proc, "<retEvento ")
	if posEvento < 0 || posRetorno < 0 {
		t.Fatalf("faltou um dos dois blocos: %s", proc)
	}
	if posEvento > posRetorno {
		t.Error("o evento tem de vir ANTES do retEvento")
	}
	if !strings.Contains(proc, "<SignatureValue>AAA</SignatureValue>") {
		t.Error("a assinatura do evento tem de ser preservada byte a byte")
	}

	// E o resultado tem de ser relegivel pelo proprio leitor.
	r, err := LerRetEventoNFe(proc)
	if err != nil {
		t.Fatal(err)
	}
	if r.InfEvento == nil {
		t.Error("o procEventoNFe gerado deveria ser relegivel")
	}
	if len(r.RetEvento) != 1 || r.RetEvento[0].CStat != 135 {
		t.Errorf("retEvento relido = %+v", r.RetEvento)
	}
}

// TestLerRetEventoNFe_AceitaDocumentoComPrefixo trava a troca de `Find` por
// `FindAnyNs` em onze tags.
//
// O ACBr usa `Childrens.FindAll(...)` -- comparacao COM o prefixo -- para
// `retEvento`, `itemPedido`, `detPag` e os oito grupos da Reforma, e
// `FindAnyNs` em todas as demais. A inconsistencia e dele: nada no leiaute
// justifica tratar `detPag` diferente de `autXML`.
//
// O porte usa `FindAllAnyNs` nas onze. O teste prova que o documento
// prefixado -- que o ACBr leria pela metade -- vem inteiro.
func TestLerRetEventoNFe_AceitaDocumentoComPrefixo(t *testing.T) {
	const xml = `<nfe:retEnvEvento versao="1.00" xmlns:nfe="http://www.portalfiscal.inf.br/nfe">` +
		`<nfe:idLote>1</nfe:idLote><nfe:tpAmb>2</nfe:tpAmb>` +
		`<nfe:cStat>128</nfe:cStat><nfe:xMotivo>Lote processado</nfe:xMotivo>` +
		`<nfe:retEvento versao="1.00"><nfe:infEvento Id="ID1">` +
		`<nfe:cStat>135</nfe:cStat>` +
		`<nfe:chNFe>` + chaveDeMinas + `</nfe:chNFe>` +
		`<nfe:tpEvento>110111</nfe:tpEvento>` +
		`<nfe:nProt>131260000000001</nfe:nProt>` +
		`</nfe:infEvento></nfe:retEvento></nfe:retEnvEvento>`

	r, err := LerRetEventoNFe(xml)
	if err != nil {
		t.Fatal(err)
	}
	if !r.LoteProcessado() {
		t.Errorf("cStat do lote = %d", r.CStat)
	}
	if len(r.RetEvento) != 1 {
		t.Fatalf("retEvento lidos = %d; o `FindAll` do ACBr com prefixo"+
			" devolveria zero", len(r.RetEvento))
	}
	if r.RetEvento[0].NProt != "131260000000001" {
		t.Errorf("nProt = %q", r.RetEvento[0].NProt)
	}
}

// TestLerDetEvento_GrupoRepetidoComPrefixo e o par do teste acima para os
// grupos que o ACBr le com `FindAll`.
func TestLerDetEvento_GrupoRepetidoComPrefixo(t *testing.T) {
	const xml = `<nfe:evento versao="1.00" xmlns:nfe="http://www.portalfiscal.inf.br/nfe">` +
		`<nfe:infEvento Id="X"><nfe:cOrgao>31</nfe:cOrgao>` +
		`<nfe:tpAmb>2</nfe:tpAmb><nfe:CNPJ>99999999000191</nfe:CNPJ>` +
		`<nfe:chNFe>` + chaveDeMinas + `</nfe:chNFe>` +
		`<nfe:dhEvento>2026-01-20T18:45:00-03:00</nfe:dhEvento>` +
		`<nfe:tpEvento>211150</nfe:tpEvento><nfe:nSeqEvento>1</nfe:nSeqEvento>` +
		`<nfe:verEvento>1.00</nfe:verEvento>` +
		`<nfe:detEvento versao="1.00"><nfe:descEvento>x</nfe:descEvento>` +
		`<nfe:gCredito nItem="1"><nfe:vCredIBS>10.00</nfe:vCredIBS>` +
		`<nfe:vCredCBS>20.00</nfe:vCredCBS></nfe:gCredito>` +
		`</nfe:detEvento></nfe:infEvento></nfe:evento>`

	r, err := LerRetEventoNFe(xml)
	if err != nil {
		t.Fatal(err)
	}
	if r.InfEvento == nil {
		t.Fatal("infEvento nao foi lido")
	}
	g := r.InfEvento.DetEvento.GCredito
	if len(g) != 1 {
		t.Fatalf("gCredito lidos = %d, esperado 1", len(g))
	}
	if g[0].VCredIBS != 10 || g[0].VCredCBS != 20 {
		t.Errorf("valores = %v / %v", g[0].VCredIBS, g[0].VCredCBS)
	}
}
