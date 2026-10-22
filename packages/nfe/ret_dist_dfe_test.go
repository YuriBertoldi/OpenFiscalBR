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

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do leitor da Distribuicao DF-e.
//
// Cada item do lote vem GZIP em base64, e os testes comprimem de verdade --
// um docZip montado a mao com texto puro nao exercita o caminho que
// importa, e e justamente ali que o leitor pode falhar em silencio.

// comprimir devolve o conteudo em gzip+base64, como a RFB o envia.
func comprimir(t *testing.T, conteudo string) string {
	t.Helper()
	s, err := dfe.GzipBase64(conteudo)
	if err != nil {
		t.Fatalf("GzipBase64: %v", err)
	}
	return s
}

// montarRetDist monta um retDistDFeInt com os docZip informados.
func montarRetDist(t *testing.T, cStat int, ultNSU, maxNSU string, itens ...string) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(`<retDistDFeInt versao="1.01" xmlns="http://www.portalfiscal.inf.br/nfe">`)
	b.WriteString(`<tpAmb>2</tpAmb><verAplic>1.0</verAplic>`)
	b.WriteString(`<cStat>` + itoa(cStat) + `</cStat>`)
	b.WriteString(`<xMotivo>Documento localizado</xMotivo>`)
	b.WriteString(`<dhResp>2026-01-20T18:45:00-03:00</dhResp>`)
	b.WriteString(`<ultNSU>` + ultNSU + `</ultNSU>`)
	b.WriteString(`<maxNSU>` + maxNSU + `</maxNSU>`)
	if len(itens) > 0 {
		b.WriteString(`<loteDistDFeInt>`)
		for _, it := range itens {
			b.WriteString(it)
		}
		b.WriteString(`</loteDistDFeInt>`)
	}
	b.WriteString(`</retDistDFeInt>`)
	return b.String()
}

// docZipDe monta um <docZip> com o conteudo COMPRIMIDO de verdade.
func docZipDe(t *testing.T, nsu, schema, conteudo string) string {
	t.Helper()
	return `<docZip NSU="` + nsu + `" schema="` + schema + `">` +
		comprimir(t, conteudo) + `</docZip>`
}

const resumoNFe = `<resNFe versao="1.01" xmlns="http://www.portalfiscal.inf.br/nfe">` +
	`<chNFe>` + chaveDeMinas + `</chNFe>` +
	`<CNPJ>99999999000191</CNPJ><xNome>EMPRESA EMITENTE</xNome>` +
	`<IE>1234567890</IE><dhEmi>2026-01-20T10:00:00-03:00</dhEmi>` +
	`<tpNF>1</tpNF><vNF>1500.00</vNF><digVal>abc123</digVal>` +
	`<dhRecbto>2026-01-20T10:05:00-03:00</dhRecbto>` +
	`<nProt>131260000000001</nProt><cSitNFe>1</cSitNFe></resNFe>`

const notaProcessada = `<nfeProc versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
	`<NFe><infNFe Id="NFe` + chaveDeMinas + `" versao="4.00">` +
	`<ide><cUF>31</cUF><dhEmi>2026-01-20T10:00:00-03:00</dhEmi><tpNF>1</tpNF></ide>` +
	`<emit><CNPJ>99999999000191</CNPJ><xNome>EMPRESA EMITENTE</xNome>` +
	`<IE>1234567890</IE></emit>` +
	`<total><ICMSTot><vNF>1500.00</vNF></ICMSTot></total>` +
	`</infNFe></NFe>` +
	`<protNFe versao="4.00"><infProt><digVal>abc123</digVal>` +
	`<dhRecbto>2026-01-20T10:05:00-03:00</dhRecbto>` +
	`<nProt>131260000000001</nProt><cStat>100</cStat></infProt></protNFe>` +
	`</nfeProc>`

// TestLerRetDistDFeInt_DescomprimeCadaDocZip trava o caminho de
// descompressao, que e exclusivo deste servico.
func TestLerRetDistDFeInt_DescomprimeCadaDocZip(t *testing.T) {
	xml := montarRetDist(t, 138, "000000000000002", "000000000000010",
		docZipDe(t, "000000000000001", "resNFe_v1.01", resumoNFe),
		docZipDe(t, "000000000000002", "procNFe_v4.00", notaProcessada))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}

	if !r.Atendida() {
		t.Errorf("cStat %d deveria ser atendido", r.CStat)
	}
	if len(r.DocZip) != 2 {
		t.Fatalf("docZip lidos = %d, esperado 2", len(r.DocZip))
	}

	for i, d := range r.DocZip {
		if d.ErroLeitura != nil {
			t.Errorf("item %d: %v", i, d.ErroLeitura)
		}
		if d.XML == "" {
			t.Errorf("item %d: o XML descomprimido deveria estar preenchido", i)
		}
		if d.InfZip == "" {
			t.Errorf("item %d: o conteudo COMO VEIO tambem e guardado", i)
		}
		if d.XML == d.InfZip {
			t.Errorf("item %d: o XML nao foi descomprimido", i)
		}
	}

	if r.DocZip[0].NSU != "000000000000001" {
		t.Errorf("NSU do primeiro = %q", r.DocZip[0].NSU)
	}
}

// TestLerRetDistDFeInt_ResumoEDocumentoPreenchemOMESMOStruct trava a
// decisao de :628.
//
// O `resNFe` tem os campos prontos; o `procNFe` os tem espalhados por
// quatro grupos aninhados. Os dois alimentam o MESMO ResDFe, e o resultado
// tem de ser indistinguivel -- senao o consumidor precisa saber de onde o
// item veio para saber onde olhar.
func TestLerRetDistDFeInt_ResumoEDocumentoPreenchemOMESMOStruct(t *testing.T) {
	xml := montarRetDist(t, 138, "000000000000002", "000000000000002",
		docZipDe(t, "1", "resNFe_v1.01", resumoNFe),
		docZipDe(t, "2", "procNFe_v4.00", notaProcessada))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}

	resumo, documento := r.DocZip[0].ResDFe, r.DocZip[1].ResDFe

	for _, c := range []struct {
		nome string
		a, b interface{}
	}{
		{"chNFe", resumo.ChDFe, documento.ChDFe},
		{"CNPJ", resumo.CNPJCPF, documento.CNPJCPF},
		{"xNome", resumo.XNome, documento.XNome},
		{"IE", resumo.IE, documento.IE},
		{"vNF", resumo.VNF, documento.VNF},
		{"digVal", resumo.DigVal, documento.DigVal},
		{"nProt", resumo.NProt, documento.NProt},
		{"cSitDFe", resumo.CSitDFe, documento.CSitDFe},
	} {
		if c.a != c.b {
			t.Errorf("%s: resumo = %v, documento = %v -- os dois caminhos"+
				" precisam dar no mesmo", c.nome, c.a, c.b)
		}
	}

	// A data vai separada: dois time.Time do MESMO instante comparam
	// diferente com `==` quando vem de parses distintos, porque o ponteiro
	// de Location difere. E `Equal` que compara instante.
	if !resumo.DhEmi.Equal(documento.DhEmi) {
		t.Errorf("dhEmi: resumo = %v, documento = %v", resumo.DhEmi, documento.DhEmi)
	}

	if resumo.ChDFe != chaveDeMinas {
		t.Errorf("chave = %q", resumo.ChDFe)
	}
	if resumo.VNF != 1500 {
		t.Errorf("vNF = %v", resumo.VNF)
	}
}

// TestLerRetDistDFeInt_ChaveDoDocumentoSaiDoATRIBUTOId trava o
// RightStr(Id, 44) de :662.
//
// No procNFe a chave nao e uma tag: sai do atributo `Id` do infNFe, pelas
// 44 posicoes FINAIS. Remover o literal "NFe" por prefixo parece
// equivalente e nao e -- os tres casos abaixo separam as duas
// implementacoes.
func TestLerRetDistDFeInt_ChaveDoDocumentoSaiDoATRIBUTOId(t *testing.T) {
	casos := []struct {
		id        string
		descricao string
	}{
		{"NFe" + chaveDeMinas, "com o literal NFe"},
		// A RFB nao garante o prefixo.
		{chaveDeMinas, "sem prefixo nenhum"},
		// E um prefixo QUALQUER tambem tem de cair: e o que distingue
		// RightStr de TrimPrefix.
		{"ID" + chaveDeMinas, "com outro prefixo"},
		{"qualquercoisa" + chaveDeMinas, "com prefixo longo"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			nota := strings.Replace(notaProcessada,
				`Id="NFe`+chaveDeMinas+`"`, `Id="`+c.id+`"`, 1)

			xml := montarRetDist(t, 138, "1", "1",
				docZipDe(t, "1", "procNFe_v4.00", nota))

			r, err := LerRetDistDFeInt(xml)
			if err != nil {
				t.Fatal(err)
			}
			if got := r.DocZip[0].ResDFe.ChDFe; got != chaveDeMinas {
				t.Errorf("Id %q: chave = %q, esperada %q", c.id, got, chaveDeMinas)
			}
		})
	}
}

// TestLerRetDistDFeInt_SituacaoSAIDoCStatDoProtocolo trava o `case` de
// :620.
//
// E o unico lugar do porte onde um cStat vira enum de situacao. O `case`
// do Delphi NAO tem `else`: codigo fora dos quatro deixa o campo como
// estava, e esse comportamento tambem e testado.
func TestLerRetDistDFeInt_SituacaoSAIDoCStatDoProtocolo(t *testing.T) {
	casos := []struct {
		cStat     int
		esperado  pcn.SituacaoDFe
		descricao string
	}{
		{100, pcn.SnAutorizado, "100 autorizado"},
		{101, pcn.SnCancelado, "101 cancelado"},
		{110, pcn.SnDenegado, "110 denegado"},
		{132, pcn.SnEncerrado, "132 encerrado"},
		// 150 e "autorizado fora do prazo" -- nao esta no `case`, e o
		// campo fica no zero value.
		{150, pcn.SnAutorizado, "150 nao esta no case: fica no zero value"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			nota := strings.Replace(notaProcessada,
				"<cStat>100</cStat>", "<cStat>"+itoa(c.cStat)+"</cStat>", 1)
			xml := montarRetDist(t, 138, "1", "1",
				docZipDe(t, "1", "procNFe_v4.00", nota))

			r, err := LerRetDistDFeInt(xml)
			if err != nil {
				t.Fatal(err)
			}
			if got := r.DocZip[0].ResDFe.CSitDFe; got != c.esperado {
				t.Errorf("cStat %d: situacao = %v, esperada %v",
					c.cStat, got, c.esperado)
			}
		})
	}
}

// TestLerRetDistDFeInt_DhEmiCaiParaDEmiNa200 trava o retrocesso de :563.
//
// Na versao 2.00 a data de emissao ainda nao tinha hora, e a tag era
// `dEmi`. Sem o retrocesso, toda nota antiga vem com a data zerada.
func TestLerRetDistDFeInt_DhEmiCaiParaDEmiNa200(t *testing.T) {
	antiga := strings.Replace(notaProcessada,
		"<dhEmi>2026-01-20T10:00:00-03:00</dhEmi>", "<dEmi>2026-01-20</dEmi>", 1)

	xml := montarRetDist(t, 138, "1", "1",
		docZipDe(t, "1", "procNFe_v2.00", antiga))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}
	if r.DocZip[0].ResDFe.DhEmi.IsZero() {
		t.Error("a data deveria ter caido para o dEmi da 2.00")
	}
	if got := r.DocZip[0].ResDFe.DhEmi.Format("2006-01-02"); got != "2026-01-20" {
		t.Errorf("dhEmi = %q", got)
	}
}

// TestLerRetDistDFeInt_TipoVemDaRAIZNaoDoAtributo trava a decisao de :817.
//
// O ACBr despacha pelo NOME DA RAIZ do documento descomprimido e ignora o
// atributo `schema` nessa escolha. O teste poe os dois em desacordo: o
// atributo diz `resEvento`, a raiz diz `resNFe`. Quem decide e a raiz.
func TestLerRetDistDFeInt_TipoVemDaRAIZNaoDoAtributo(t *testing.T) {
	xml := montarRetDist(t, 138, "1", "1",
		docZipDe(t, "1", "resEvento_v1.00", resumoNFe))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}
	d := r.DocZip[0]

	// O atributo foi lido e guardado...
	if d.Schema != pcn.SchResEvento {
		t.Errorf("o atributo schema = %v, esperado SchResEvento", d.Schema)
	}
	// ...mas quem preencheu os campos foi a RAIZ.
	if d.ResDFe.ChDFe != chaveDeMinas {
		t.Errorf("o resumo de NF-e deveria ter sido lido: %+v", d.ResDFe)
	}
	if d.ResEvento.ChDFe != "" {
		t.Errorf("o grupo de evento deveria estar vazio: %+v", d.ResEvento)
	}
}

// TestLerRetDistDFeInt_ItemDesconhecidoNaoDerrubaOLote trava a tolerancia.
//
// Um lote de cinquenta documentos nao pode falhar inteiro porque a RFB
// acrescentou um tipo novo. O item fica com o XML cru e os campos vazios,
// e os vizinhos seguem lidos.
func TestLerRetDistDFeInt_ItemDesconhecidoNaoDerrubaOLote(t *testing.T) {
	xml := montarRetDist(t, 138, "3", "3",
		docZipDe(t, "1", "resNFe_v1.01", resumoNFe),
		docZipDe(t, "2", "procAlgoNovo_v1.00",
			`<procAlgoNovo versao="1.00"><campo>x</campo></procAlgoNovo>`),
		docZipDe(t, "3", "resNFe_v1.01", resumoNFe))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatalf("o lote inteiro nao pode falhar: %v", err)
	}
	if len(r.DocZip) != 3 {
		t.Fatalf("docZip lidos = %d, esperado 3", len(r.DocZip))
	}

	desconhecido := r.DocZip[1]
	if desconhecido.Schema != pcn.SchemaDesconhecido {
		t.Errorf("schema = %v, esperado SchemaDesconhecido", desconhecido.Schema)
	}
	if desconhecido.XML == "" {
		t.Error("o XML cru fica guardado mesmo sem leitor para ele")
	}
	if desconhecido.ErroLeitura != nil {
		t.Errorf("raiz desconhecida nao e erro de leitura: %v", desconhecido.ErroLeitura)
	}

	// E os vizinhos continuam lidos.
	for _, i := range []int{0, 2} {
		if r.DocZip[i].ResDFe.ChDFe != chaveDeMinas {
			t.Errorf("item %d perdeu o conteudo: %+v", i, r.DocZip[i].ResDFe)
		}
	}
}

// TestLerRetDistDFeInt_ConteudoCorrompidoMarcaSOOItem e o par do teste
// acima para o caso de gzip invalido.
//
// No ACBr um docZip corrompido cai no `except` do LerXml e descarta o lote
// inteiro. Aqui o item fica marcado e os demais seguem.
func TestLerRetDistDFeInt_ConteudoCorrompidoMarcaSOOItem(t *testing.T) {
	xml := montarRetDist(t, 138, "2", "2",
		`<docZip NSU="1" schema="resNFe_v1.01">isto nao e base64 de gzip</docZip>`,
		docZipDe(t, "2", "resNFe_v1.01", resumoNFe))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatalf("o lote inteiro nao pode falhar: %v", err)
	}
	if len(r.DocZip) != 2 {
		t.Fatalf("docZip lidos = %d, esperado 2", len(r.DocZip))
	}
	if r.DocZip[0].ErroLeitura == nil {
		t.Error("o item corrompido deveria estar marcado")
	}
	if r.DocZip[1].ResDFe.ChDFe != chaveDeMinas {
		t.Error("o item seguinte deveria ter sido lido normalmente")
	}
}

// TestRetDistDFeInt_O137EhSUCESSO trava o criterio de :4120.
//
// "Nenhum documento localizado" e a resposta mais comum de uma varredura
// diaria. Trata-la como falha faz o laco nunca terminar.
func TestRetDistDFeInt_O137EhSUCESSO(t *testing.T) {
	for _, c := range []struct {
		cStat    int
		atendida bool
	}{
		{138, true},
		{137, true},
		{656, false}, // consumo indevido
		{589, false},
		{0, false},
	} {
		r := &RetDistDFeInt{CStat: c.cStat}
		if r.Atendida() != c.atendida {
			t.Errorf("cStat %d: Atendida() = %v, esperado %v",
				c.cStat, r.Atendida(), c.atendida)
		}
	}
}

// TestRetDistDFeInt_VarreduraTerminaNoMaxNSU trava o criterio de parada.
//
// A RFB devolve no maximo cinquenta docZip por chamada, entao "o lote veio
// cheio" nao quer dizer que acabou, e "o lote veio vazio" tambem nao e o
// criterio. O que vale e ultNSU == maxNSU.
func TestRetDistDFeInt_VarreduraTerminaNoMaxNSU(t *testing.T) {
	casos := []struct {
		cStat          int
		ultNSU, maxNSU string
		concluida      bool
		descricao      string
	}{
		{138, "000000000000050", "000000000000120", false,
			"lote cheio mas falta mais"},
		{138, "000000000000120", "000000000000120", true,
			"ultNSU alcancou o maxNSU"},
		{137, "000000000000120", "000000000000120", true,
			"nenhum documento: acabou"},
		{137, "", "", true,
			"nenhum documento e sem NSU: acabou"},
		{138, "", "000000000000120", false,
			"sem ultNSU devolvido: nao da para concluir"},
	}
	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			r := &RetDistDFeInt{CStat: c.cStat, UltNSU: c.ultNSU, MaxNSU: c.maxNSU}
			if got := r.VarreduraConcluida(); got != c.concluida {
				t.Errorf("= %v, esperado %v", got, c.concluida)
			}
		})
	}
}

// TestLerRetDistDFeInt_EventoProcessado trava a leitura do procEventoNFe,
// que preenche um TERCEIRO grupo do item.
func TestLerRetDistDFeInt_EventoProcessado(t *testing.T) {
	const evento = `<procEventoNFe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<evento versao="1.00"><infEvento Id="ID1101113126019999">` +
		`<cOrgao>31</cOrgao><tpAmb>2</tpAmb><CNPJ>99999999000191</CNPJ>` +
		`<chNFe>` + chaveDeMinas + `</chNFe>` +
		`<dhEvento>2026-01-20T18:45:00-03:00</dhEvento>` +
		`<tpEvento>110111</tpEvento><nSeqEvento>1</nSeqEvento>` +
		`<verEvento>1.00</verEvento>` +
		`<detEvento versao="1.00"><descEvento>Cancelamento</descEvento>` +
		`<nProt>131260000000001</nProt><xJust>Erro na emissao</xJust>` +
		`</detEvento></infEvento></evento>` +
		`<retEvento versao="1.00"><infEvento Id="IDRET">` +
		`<tpAmb>2</tpAmb><verAplic>1.0</verAplic><cOrgao>31</cOrgao>` +
		`<cStat>135</cStat><xMotivo>Evento registrado</xMotivo>` +
		`<chNFe>` + chaveDeMinas + `</chNFe>` +
		`<tpEvento>110111</tpEvento><xEvento>Cancelamento</xEvento>` +
		`<nSeqEvento>1</nSeqEvento>` +
		`<dhRegEvento>2026-01-20T18:46:00-03:00</dhRegEvento>` +
		`<nProt>131260000000009</nProt>` +
		`</infEvento></retEvento></procEventoNFe>`

	xml := montarRetDist(t, 138, "1", "1",
		docZipDe(t, "1", "procEventoNFe_v1.00", evento))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}
	p := r.DocZip[0].ProcEvento

	if p.TpEvento != pcn.TeCancelamento {
		t.Errorf("tpEvento = %s", p.TpEvento)
	}
	if p.ChDFe != chaveDeMinas {
		t.Errorf("chNFe = %q", p.ChDFe)
	}
	if p.COrgao != 31 {
		t.Errorf("cOrgao = %d -- o ACBr nao o le aqui; e acrescimo", p.COrgao)
	}
	if p.DetEvento.XJust != "Erro na emissao" {
		t.Errorf("detEvento/xJust = %q", p.DetEvento.XJust)
	}
	if p.RetInfEvento.CStat != 135 {
		t.Errorf("retEvento/cStat = %d", p.RetInfEvento.CStat)
	}
	if p.RetInfEvento.NProt != "131260000000009" {
		t.Errorf("retEvento/nProt = %q", p.RetInfEvento.NProt)
	}
	// Os outros dois grupos ficam vazios.
	if r.DocZip[0].ResDFe.ChDFe != "" || r.DocZip[0].ResEvento.ChDFe != "" {
		t.Error("so um dos tres grupos e preenchido por item")
	}
}

// TestLerRetDistDFeInt_ResumoDeEvento trava o quarto schema, que e o unico
// a preencher o ResEvento.
//
// O resumo de evento NAO e um subconjunto do evento processado: traz
// `xEvento` -- a descricao JA RESOLVIDA pela RFB -- e `dhRecbto`, que nao
// existem no infEvento, e nao traz o retEvento. Quem so testa o
// procEventoNFe deixa esse caminho inteiro sem cobertura.
func TestLerRetDistDFeInt_ResumoDeEvento(t *testing.T) {
	const resumoEvento = `<resEvento versao="1.01" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<cOrgao>31</cOrgao><CNPJ>99999999000191</CNPJ>` +
		`<chNFe>` + chaveDeMinas + `</chNFe>` +
		`<dhEvento>2026-01-20T18:45:00-03:00</dhEvento>` +
		`<tpEvento>110110</tpEvento><nSeqEvento>3</nSeqEvento>` +
		`<xEvento>Carta de Correcao</xEvento>` +
		`<dhRecbto>2026-01-20T18:46:00-03:00</dhRecbto>` +
		`<nProt>131260000000009</nProt></resEvento>`

	xml := montarRetDist(t, 138, "1", "1",
		docZipDe(t, "1", "resEvento_v1.01", resumoEvento))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}
	e := r.DocZip[0].ResEvento

	if e.COrgao != 31 {
		t.Errorf("cOrgao = %d, esperado 31", e.COrgao)
	}
	if e.CNPJCPF != "99999999000191" {
		t.Errorf("CNPJ = %q", e.CNPJCPF)
	}
	if e.ChDFe != chaveDeMinas {
		t.Errorf("chNFe = %q", e.ChDFe)
	}
	if got := e.DhEvento.Format("2006-01-02T15:04:05"); got != "2026-01-20T18:45:00" {
		t.Errorf("dhEvento = %q", got)
	}
	if e.TpEvento != pcn.TeCCe {
		t.Errorf("tpEvento = %s, esperado a CCe", e.TpEvento)
	}
	if e.NSeqEvento != 3 {
		t.Errorf("nSeqEvento = %d, esperado 3", e.NSeqEvento)
	}
	if e.XEvento != "Carta de Correcao" {
		t.Errorf("xEvento = %q", e.XEvento)
	}
	if got := e.DhRecbto.Format("2006-01-02T15:04:05"); got != "2026-01-20T18:46:00" {
		t.Errorf("dhRecbto = %q", got)
	}
	if e.NProt != "131260000000009" {
		t.Errorf("nProt = %q", e.NProt)
	}

	// Os outros dois grupos ficam vazios -- o resEvento nao alimenta o
	// ResDFe nem o ProcEvento.
	if r.DocZip[0].ResDFe.ChDFe != "" || r.DocZip[0].ProcEvento.ChDFe != "" {
		t.Error("so um dos tres grupos e preenchido por item")
	}
}

// TestLerRetDistDFeInt_ProtocoloLidoMesmoSemANFe trava a INDEPENDENCIA dos
// dois blocos de :649-671.
//
// Sao dois `if Assigned` sequenciais, nao um aninhado. Fundi-los -- sair da
// funcao quando o <NFe> falta -- descarta o protocolo inteiro, e com ele a
// SITUACAO: o item volta com CSitDFe no zero value, que e SnAutorizado.
// Uma nota cancelada apareceria como autorizada, e nenhum outro teste
// pegaria isso porque o zero value e um valor valido.
//
// E por isso que o protocolo deste caso tem cStat 101: com 100, o bug
// passaria despercebido.
func TestLerRetDistDFeInt_ProtocoloLidoMesmoSemANFe(t *testing.T) {
	const soProtocolo = `<nfeProc versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<protNFe versao="4.00"><infProt><digVal>abc123</digVal>` +
		`<dhRecbto>2026-01-20T10:05:00-03:00</dhRecbto>` +
		`<nProt>131260000000001</nProt><cStat>101</cStat></infProt></protNFe>` +
		`</nfeProc>`

	xml := montarRetDist(t, 138, "1", "1",
		docZipDe(t, "1", "procNFe_v4.00", soProtocolo))

	r, err := LerRetDistDFeInt(xml)
	if err != nil {
		t.Fatal(err)
	}
	d := r.DocZip[0].ResDFe

	if d.CSitDFe != pcn.SnCancelado {
		t.Errorf("situacao = %v, esperada SnCancelado -- o protocolo foi perdido", d.CSitDFe)
	}
	if d.NProt != "131260000000001" {
		t.Errorf("nProt = %q -- o bloco do protocolo nao foi lido", d.NProt)
	}
	if d.DigVal != "abc123" {
		t.Errorf("digVal = %q", d.DigVal)
	}
	if d.DhRecbto.IsZero() {
		t.Error("dhRecbto deveria ter vindo do protocolo")
	}
	// Sem o <NFe>, os campos da nota ficam vazios -- o que nao impede nada
	// do acima.
	if d.ChDFe != "" || d.XNome != "" {
		t.Errorf("sem o <NFe> os campos da nota ficam vazios: %+v", d)
	}
}

// TestChaveDoDocZip_OlhaOsTresGrupos trava o helper que poupa o consumidor
// de reescrever o mesmo switch.
func TestChaveDoDocZip_OlhaOsTresGrupos(t *testing.T) {
	casos := []struct {
		item      DocZip
		esperado  string
		descricao string
	}{
		{DocZip{ResDFe: ResDFe{ChDFe: chaveDeMinas}}, chaveDeMinas, "do resumo"},
		{DocZip{ResEvento: ResEvento{ChDFe: chaveDeMinas}}, chaveDeMinas,
			"do resumo de evento"},
		{DocZip{ProcEvento: ProcEventoDist{ChDFe: chaveDeMinas}}, chaveDeMinas,
			"do evento processado"},
		{DocZip{ResDFe: ResDFe{ChDFe: "NFe" + chaveDeMinas}}, chaveDeMinas,
			"o literal NFe e removido"},
		{DocZip{}, "", "item sem chave"},
	}
	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			if got := ChaveDoDocZip(c.item); got != c.esperado {
				t.Errorf("= %q, esperado %q", got, c.esperado)
			}
		})
	}
}

// TestParseSchemaDFe_CortaNoUnderscore trava :1156.
func TestParseSchemaDFe_CortaNoUnderscore(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado pcn.SchemaDFe
	}{
		{"resNFe_v1.01", pcn.SchResNFe},
		{"resNFe", pcn.SchResNFe},
		{"schresNFe", pcn.SchResNFe},
		{"procNFe_v4.00", pcn.SchProcNFe},
		{"procEventoNFe_v1.00", pcn.SchProcEventoNFe},
		{"resEvento_v1.00", pcn.SchResEvento},
		{"procCTe_v4.00", pcn.SchProcCTe},
	}
	for _, c := range casos {
		got, err := pcn.ParseSchemaDFe(c.entrada)
		if err != nil {
			t.Errorf("%q: %v", c.entrada, err)
			continue
		}
		if got != c.esperado {
			t.Errorf("%q = %v, esperado %v", c.entrada, got, c.esperado)
		}
	}

	// Desconhecido devolve erro E o sentinela -- o chamador decide se
	// segue. No ACBr ele levanta excecao e derruba o lote.
	got, err := pcn.ParseSchemaDFe("procAlgoNovo_v1.00")
	if err == nil {
		t.Error("schema desconhecido deveria dar erro")
	}
	if got != pcn.SchemaDesconhecido {
		t.Errorf("= %v, esperado SchemaDesconhecido", got)
	}
}

// TestSchemaDFe_EDocumentoDistingueResumoDeCompleto trava a distincao que
// governa o que o consumidor pode fazer com o item.
func TestSchemaDFe_EDocumentoDistingueResumoDeCompleto(t *testing.T) {
	resumos := []pcn.SchemaDFe{pcn.SchResNFe, pcn.SchResEvento,
		pcn.SchResCTe, pcn.SchResMDFe, pcn.SchResBPe}
	for _, s := range resumos {
		if s.EDocumento() {
			t.Errorf("%s e resumo, nao documento completo", s)
		}
	}

	completos := []pcn.SchemaDFe{pcn.SchProcNFe, pcn.SchProcEventoNFe,
		pcn.SchProcCTe, pcn.SchProcCTeOS, pcn.SchProcCTeSimp,
		pcn.SchProcMDFe, pcn.SchProcGTVe}
	for _, s := range completos {
		if !s.EDocumento() {
			t.Errorf("%s e documento completo", s)
		}
	}

	if pcn.SchemaDesconhecido.EDocumento() {
		t.Error("o sentinela nao e documento")
	}
}
