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
	"errors"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes da inutilizacao de faixa.
//
// O ID e o que mais escapa num porte: sao quarenta e tres posicoes FIXAS
// -- o literal "ID" mais quarenta e um digitos --, e um unico campo com
// largura errada desloca todos os seguintes sem mudar o tamanho total. Um
// teste que so conte caracteres passa.

func pedidoInutilizacaoBase() PedidoInutilizacao {
	return PedidoInutilizacao{
		Ano:           2026,
		CNPJ:          "99999999000191",
		Modelo:        55,
		Serie:         1,
		NumeroInicial: 10,
		NumeroFinal:   25,
		Justificativa: "Numeracao inutilizada por falha no sistema emissor",
	}
}

func configInutilizacao() Configuracoes {
	return Configuracoes{
		VersaoDF: Ve400,
		ModeloDF: MoNFe,
		Ambiente: pcn.TaHomologacao,
		UF:       "SP",
		CodigoUF: 35,
	}
}

// TestIDInutilizacao_LarguraDeCadaCampo trava a posicao de cada campo no ID.
//
// Porte de ACBrNFe.Inut.pas:124. A conferencia e por FATIA, nao pelo texto
// inteiro: assim a falha aponta qual campo deslocou.
func TestIDInutilizacao_LarguraDeCadaCampo(t *testing.T) {
	id := IDInutilizacao(pedidoInutilizacaoBase(), 35)

	if len(id) != 43 {
		t.Fatalf("ID tem %d caracteres, esperado 43 (\"ID\" + 41 digitos): %q", len(id), id)
	}

	fatias := []struct {
		nome     string
		inicio   int
		fim      int
		esperado string
	}{
		{"literal", 0, 2, "ID"},
		{"cUF", 2, 4, "35"},
		{"ano", 4, 6, "26"},
		{"CNPJ", 6, 20, "99999999000191"},
		{"modelo", 20, 22, "55"},
		{"serie", 22, 25, "001"},
		{"nNFIni", 25, 34, "000000010"},
		{"nNFFin", 34, 43, "000000025"},
	}
	for _, f := range fatias {
		if got := id[f.inicio:f.fim]; got != f.esperado {
			t.Errorf("%s = %q, esperado %q (ID completo: %q)", f.nome, got, f.esperado, id)
		}
	}
}

// TestIDInutilizacao_AnoDeDoisDigitosNoID trava a reducao do ano.
//
// O ID leva SEMPRE duas posicoes, venha o ano com quatro digitos ou com
// dois. Quem porta `Copy(IntToStrZero(ano,4),3,2)` como `ano mod 100`
// acerta; quem o porta como `ano - 2000` erra para o ano ja reduzido.
func TestIDInutilizacao_AnoDeDoisDigitosNoID(t *testing.T) {
	for _, ano := range []int{2026, 26} {
		p := pedidoInutilizacaoBase()
		p.Ano = ano
		if got := IDInutilizacao(p, 35)[4:6]; got != "26" {
			t.Errorf("ano %d produziu %q no ID, esperado \"26\"", ano, got)
		}
	}
}

// TestGerarXMLInutilizacao_AnoNoCorpoVaiSemZeroAEsquerda trava o
// `if ano > 2000 then ano := ano - 2000` do corpo.
//
// O ID leva "26"; o corpo leva "26" tambem, mas por um caminho diferente --
// e para ano abaixo de dez os dois DIVERGEM: o ID traz "09" e o corpo "9".
func TestGerarXMLInutilizacao_AnoNoCorpoVaiSemZeroAEsquerda(t *testing.T) {
	p := pedidoInutilizacaoBase()
	p.Ano = 2009

	xml, err := GerarXMLInutilizacao(p, configInutilizacao())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(xml, "<ano>9</ano>") {
		t.Errorf("o corpo deveria trazer <ano>9</ano>, sem zero a esquerda: %s", xml)
	}
	if got := IDInutilizacao(p, 35)[4:6]; got != "09" {
		t.Errorf("o ID deveria trazer \"09\": %q", got)
	}
}

// TestInutilizacao_MatoGrossoDoSulAceitaCPF trava a excecao da UF 51, que
// tem DUAS metades contraditorias.
//
// No ID o CPF vai preenchido com tres zeros, para ocupar as catorze
// posicoes do CNPJ; no CORPO ele vai limpo, dentro de <CPF>. Portar so uma
// das metades produz ou um ID de quarenta posicoes ou um CPF de catorze
// digitos -- as duas coisas sao rejeicao certa.
func TestInutilizacao_MatoGrossoDoSulAceitaCPF(t *testing.T) {
	p := pedidoInutilizacaoBase()
	p.CNPJ = "11122233396" // CPF de onze digitos

	cfg := configInutilizacao()
	cfg.UF, cfg.CodigoUF = "MS", 51

	id := IDInutilizacao(p, 51)
	if got := id[6:20]; got != "00011122233396" {
		t.Errorf("no ID o CPF deveria vir com 000 a esquerda, veio %q", got)
	}
	if len(id) != 43 {
		t.Errorf("o ID com CPF deveria ter as mesmas 43 posicoes, tem %d", len(id))
	}

	xml, err := GerarXMLInutilizacao(p, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<CPF>11122233396</CPF>") {
		t.Errorf("o corpo deveria trazer <CPF> sem preenchimento: %s", xml)
	}
	if strings.Contains(xml, "<CNPJ>") {
		t.Error("o corpo nao deveria trazer <CNPJ> quando o documento e CPF no MS")
	}
}

// TestInutilizacao_CPFSoValeNoMatoGrossoDoSul trava a outra metade da
// regra: em qualquer outra UF o mesmo documento de onze digitos sai como
// <CNPJ>, porque a condicao olha a UF antes do tamanho.
func TestInutilizacao_CPFSoValeNoMatoGrossoDoSul(t *testing.T) {
	p := pedidoInutilizacaoBase()
	p.CNPJ = "11122233396"

	xml, err := GerarXMLInutilizacao(p, configInutilizacao()) // SP, cUF 35
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<CNPJ>11122233396</CNPJ>") {
		t.Errorf("fora do MS o documento sai sempre em <CNPJ>: %s", xml)
	}
	if id := IDInutilizacao(p, 35); len(id) != 40 {
		// 43 menos os tres zeros que o MS acrescentaria. O ID fica CURTO, e
		// e isso que a SEFAZ rejeita -- o teste trava que nao ha
		// preenchimento fora do MS.
		t.Errorf("fora do MS nao ha preenchimento: ID tem %d caracteres, esperado 40", len(id))
	}
}

// TestGerarXMLInutilizacao_OrdemDosCampos trava a sequencia do infInut.
//
// O XSD e xs:sequence, e nenhum round-trip pega ordem trocada: o leitor
// acha a tag pelo nome.
func TestGerarXMLInutilizacao_OrdemDosCampos(t *testing.T) {
	xml, err := GerarXMLInutilizacao(pedidoInutilizacaoBase(), configInutilizacao())
	if err != nil {
		t.Fatal(err)
	}

	ordem := []string{"<tpAmb>", "<xServ>", "<cUF>", "<ano>", "<CNPJ>",
		"<mod>", "<serie>", "<nNFIni>", "<nNFFin>", "<xJust>"}
	anterior := -1
	for _, tag := range ordem {
		pos := strings.Index(xml, tag)
		if pos < 0 {
			t.Fatalf("tag %s ausente: %s", tag, xml)
		}
		if pos <= anterior {
			t.Errorf("%s aparece fora de ordem (posicao %d, anterior %d)", tag, pos, anterior)
		}
		anterior = pos
	}
}

// TestValidarPedidoInutilizacao_Justificativa trava o minimo de quinze
// caracteres de SetJustificativa (ACBrNFeWebServices.pas:2634).
//
// O limite e medido no texto JA APARADO: espacos em volta nao contam.
func TestValidarPedidoInutilizacao_Justificativa(t *testing.T) {
	casos := []struct {
		justificativa string
		erro          error
		descricao     string
	}{
		{"", ErrJustificativaObrigatoria, "vazia"},
		{"      ", ErrJustificativaObrigatoria, "so espacos"},
		{"quatorze letra", ErrJustificativaCurta, "14 caracteres"},
		{"   quatorze letra   ", ErrJustificativaCurta, "14 apos aparar"},
		{"quinze caracter", nil, "exatamente 15"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			p := pedidoInutilizacaoBase()
			p.Justificativa = c.justificativa
			err := ValidarPedidoInutilizacao(p)
			if !errors.Is(err, c.erro) {
				t.Errorf("erro = %v, esperado %v", err, c.erro)
			}
		})
	}
}

// TestGerarXMLInutilizacao_JustificativaVaiAparada trava que o XML leva o
// texto aparado, e nao o original -- o SetJustificativa guarda o Trim
// (:2628) e o DefinirDadosMsg usa o campo guardado.
func TestGerarXMLInutilizacao_JustificativaVaiAparada(t *testing.T) {
	p := pedidoInutilizacaoBase()
	p.Justificativa = "  Faixa nao utilizada pelo contribuinte  "

	xml, err := GerarXMLInutilizacao(p, configInutilizacao())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(xml, "<xJust>Faixa nao utilizada pelo contribuinte</xJust>") {
		t.Errorf("a justificativa deveria ir aparada: %s", xml)
	}
}

// TestRetInutNFe_ProcInutNFeSaiEmDoisCodigos trava a divergencia entre
// "deu certo" e "vale arquivar".
//
// O ACBr devolve sucesso so no 102, mas gera o ProcInutNFe tambem no 563
// (ACBrNFeWebServices.pas:2765 e :2777). Unificar os dois criterios -- o
// porte natural -- deixaria de arquivar a inutilizacao que a SEFAZ ja tem.
func TestRetInutNFe_ProcInutNFeSaiEmDoisCodigos(t *testing.T) {
	casos := []struct {
		cStat        int
		homologada   bool
		temDocumento bool
	}{
		{102, true, true},
		{563, false, true},
		{241, false, false},
		{0, false, false},
	}

	for _, c := range casos {
		r := &RetInutNFe{CStat: c.cStat}
		if r.Homologada() != c.homologada {
			t.Errorf("cStat %d: Homologada() = %v, esperado %v",
				c.cStat, r.Homologada(), c.homologada)
		}
		if r.TemProcInutNFe() != c.temDocumento {
			t.Errorf("cStat %d: TemProcInutNFe() = %v, esperado %v",
				c.cStat, r.TemProcInutNFe(), c.temDocumento)
		}
	}
}

// TestLerRetInutNFe_AceitaOEnvelopeDoProc trava as duas formas de raiz que
// o ACBr aceita (ACBrNFe.RetInut.pas:124-133).
//
// A segunda e o arquivo de inutilizacao relido do disco, cuja raiz e
// ProcInutNFe e nao retInutNFe.
func TestLerRetInutNFe_AceitaOEnvelopeDoProc(t *testing.T) {
	retorno := `<retInutNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<infInut Id="ID3526999999990001915500100000001000000025">` +
		`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic><cStat>102</cStat>` +
		`<xMotivo>Inutilizacao de numero homologado</xMotivo><cUF>35</cUF>` +
		`<ano>26</ano><CNPJ>99999999000191</CNPJ><mod>55</mod><serie>1</serie>` +
		`<nNFIni>10</nNFIni><nNFFin>25</nNFFin>` +
		`<dhRecbto>2026-01-20T18:45:00-03:00</dhRecbto>` +
		`<nProt>135260000000009</nProt></infInut></retInutNFe>`

	envelopado := `<ProcInutNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<inutNFe><infInut Id="X"><xJust>pedido</xJust></infInut></inutNFe>` +
		retorno + `</ProcInutNFe>`

	for _, caso := range []struct {
		nome string
		xml  string
	}{
		{"raiz e o proprio retInutNFe", retorno},
		{"raiz e o ProcInutNFe", envelopado},
	} {
		t.Run(caso.nome, func(t *testing.T) {
			r, err := LerRetInutNFe(caso.xml)
			if err != nil {
				t.Fatal(err)
			}
			if r.CStat != 102 {
				t.Errorf("cStat = %d, esperado 102", r.CStat)
			}
			if r.NProt != "135260000000009" {
				t.Errorf("nProt = %q", r.NProt)
			}
			if r.NNFIni != 10 || r.NNFFin != 25 {
				t.Errorf("faixa lida = %d..%d, esperado 10..25", r.NNFIni, r.NNFFin)
			}
		})
	}
}

// TestLerRetInutNFe_NoEnvelopeOInfInutDoPedidoNaoVence trava a precedencia
// do laco de :126-129: quando ha retInutNFe, o infInut lido e o DELE, nao
// o primeiro infInut do documento -- que, no ProcInutNFe, e o do PEDIDO.
//
// Sem a precedencia o leitor devolveria a justificativa enviada no lugar do
// protocolo recebido.
func TestLerRetInutNFe_NoEnvelopeOInfInutDoPedidoNaoVence(t *testing.T) {
	envelopado := `<ProcInutNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<inutNFe><infInut Id="IDPEDIDO"><cStat>0</cStat>` +
		`<xJust>justificativa do pedido</xJust></infInut></inutNFe>` +
		`<retInutNFe versao="4.00"><infInut Id="IDRETORNO">` +
		`<cStat>102</cStat><nProt>135260000000009</nProt>` +
		`</infInut></retInutNFe></ProcInutNFe>`

	r, err := LerRetInutNFe(envelopado)
	if err != nil {
		t.Fatal(err)
	}
	if r.ID != "IDRETORNO" {
		t.Errorf("Id = %q, esperado IDRETORNO -- o infInut do retorno vence o do pedido", r.ID)
	}
	if r.CStat != 102 {
		t.Errorf("cStat = %d, esperado 102", r.CStat)
	}
}

// TestGerarXMLProcInutNFe_ColaPedidoERetorno trava a montagem do documento
// de arquivo: pedido ASSINADO primeiro, retorno depois.
//
// A ordem importa porque o XSD do ProcInutNFe e sequence, e o pedido tem de
// entrar como veio -- regerar invalidaria a assinatura.
func TestGerarXMLProcInutNFe_ColaPedidoERetorno(t *testing.T) {
	pedido := `<inutNFe versao="4.00"><infInut Id="ID1"><xJust>j</xJust></infInut>` +
		`<Signature><SignatureValue>AAA</SignatureValue></Signature></inutNFe>`
	retorno := `<retInutNFe versao="4.00"><infInut><cStat>102</cStat></infInut></retInutNFe>`

	proc := GerarXMLProcInutNFe(pedido, retorno, "4.00")

	if !strings.HasPrefix(proc, `<ProcInutNFe versao="4.00"`) {
		t.Errorf("raiz errada: %s", proc)
	}
	posPedido := strings.Index(proc, "<inutNFe")
	posRetorno := strings.Index(proc, "<retInutNFe")
	if posPedido < 0 || posRetorno < 0 {
		t.Fatalf("faltou um dos dois blocos: %s", proc)
	}
	if posPedido > posRetorno {
		t.Error("o pedido tem de vir ANTES do retorno")
	}
	if !strings.Contains(proc, "<SignatureValue>AAA</SignatureValue>") {
		t.Error("a assinatura do pedido tem de ser preservada byte a byte")
	}
}
