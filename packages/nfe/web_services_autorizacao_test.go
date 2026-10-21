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
	"errors"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes dos servicos de escrita.
//
// A maioria afirma sobre a REQUISICAO, nao sobre a resposta: o que um porte
// erra aqui e a mensagem enviada e o endereco escolhido, e os dois sao
// decididos antes de a SEFAZ dizer qualquer coisa.

// notaParaEnvio monta uma nota ja com XMLAssinado, para exercitar o envio
// sem precisar de certificado.
//
// O XML e falso de proposito: o que o Enviar faz com ele e recortar o
// trecho <NFe>..</NFe>, e nenhum passo do caminho testado o interpreta.
func notaParaEnvio(chave string, cUF, modelo int, versao float64, ambiente pcn.TipoAmbiente) *NotaFiscal {
	n := NovaNFe()
	n.InfNFe.ID = "NFe" + chave
	n.InfNFe.Versao = versao
	n.Ide.CUF = cUF
	n.Ide.Modelo = modelo
	n.Ide.TpAmb = ambiente

	return &NotaFiscal{
		NFe: &n,
		XMLAssinado: `<?xml version="1.0"?><NFe xmlns="http://www.portalfiscal.inf.br/nfe">` +
			`<infNFe Id="NFe` + chave + `"><ide><cUF>` + itoa(cUF) + `</cUF></ide></infNFe>` +
			`<Signature><SignatureValue>AAA</SignatureValue></Signature></NFe>`,
	}
}

// TestEhAutorizacao_ExcecaoDoAmazonas trava a excecao por UF.
//
// Porte de ACBrNFe.pas:313. A regra "versao >= 3.10" cobre quinze dos
// dezesseis casos; o decimo sexto -- NFC-e 3.10 no Amazonas -- continua no
// servico antigo, e e o unico que um porte apressado erra.
func TestEhAutorizacao_ExcecaoDoAmazonas(t *testing.T) {
	casos := []struct {
		versao    VersaoDF
		modelo    ModeloDF
		cUF       int
		esperado  bool
		descricao string
	}{
		{Ve200, MoNFe, 35, false, "NF-e 2.00 usa o lote antigo"},
		{Ve310, MoNFe, 35, true, "NF-e 3.10 ja e autorizacao"},
		{Ve400, MoNFe, 35, true, "NF-e 4.00"},
		{Ve310, MoNFCe, 35, true, "NFC-e 3.10 fora do AM e autorizacao"},
		{Ve310, MoNFCe, 13, false, "NFC-e 3.10 no AM continua no lote antigo"},
		{Ve400, MoNFCe, 13, true, "na 4.00 o AM ja migrou"},
		{Ve200, MoNFCe, 13, false, "NFC-e 2.00 no AM"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			if got := EhAutorizacao(c.versao, c.modelo, c.cUF); got != c.esperado {
				t.Errorf("EhAutorizacao(%v, %v, %d) = %v, esperado %v",
					c.versao, c.modelo, c.cUF, got, c.esperado)
			}
		})
	}
}

// TestEnviar_VersaoVemDoDocumentoENaoDaConfiguracao trava a decisao de
// InicializarServico (:1026).
//
// O componente esta configurado para 3.10 e a nota e 4.00: o endereco, o
// servico e o cabecalho tem de seguir a NOTA. Um porte que leia a
// configuracao manda o lote 4.00 para o endereco da 3.10.
func TestEnviar_VersaoVemDoDocumentoENaoDaConfiguracao(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeAutorizacaoLoteResult",
		`<retEnviNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">`+
			`<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>35</cUF>`+
			`<infRec><nRec>351000000000001</nRec><tMed>1</tMed></infRec></retEnviNFe>`))

	cfg := configDeTransmissao()
	cfg.VersaoDF = Ve310 // o COMPONENTE esta na 3.10
	c := componenteApontandoPara(sefaz, cfg)

	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao) // a NOTA e 4.00

	destino := DestinoEnvio{URL: sefaz.servidor.URL, Versao: Ve400,
		Modelo: MoNFe, UF: "SP", Autorizacao: true}
	if _, err := c.enviarNaURL(context.Background(), destino,
		[]*NotaFiscal{nota}, OpcoesEnvio{Lote: "1"}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sefaz.ultimoCorpo, `versao="4.00"`) {
		t.Errorf("o enviNFe deveria sair com versao 4.00: %s", sefaz.ultimoCorpo)
	}
	// Na 4.00 nao ha <soap12:Header> nenhum (:1030).
	if strings.Contains(sefaz.ultimoCorpo, "nfeCabecMsg") {
		t.Error("lote 4.00 nao leva cabecalho, mesmo com o componente na 3.10")
	}
	if !strings.Contains(sefaz.ultimaAcao, "NFeAutorizacao4/nfeAutorizacaoLote") {
		t.Errorf("SoapAction = %q, esperada a da 4.00", sefaz.ultimaAcao)
	}
}

// TestEnviar_VersaoDadosDoCabecalhoVemDoDocumento trava a outra metade da
// mesma decisao.
//
// O teste anterior usa a 4.00, onde NAO ha cabecalho -- entao ele nao
// consegue afirmar sobre o versaoDados. Aqui o componente esta na 2.00 e a
// nota na 3.10: o cabecalho existe, e o numero dentro dele tem de ser o do
// DOCUMENTO.
func TestEnviar_VersaoDadosDoCabecalhoVemDoDocumento(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeAutorizacaoLoteResult",
		`<retEnviNFe versao="3.10" xmlns="http://www.portalfiscal.inf.br/nfe">`+
			`<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>35</cUF>`+
			`<infRec><nRec>351000000000001</nRec><tMed>1</tMed></infRec></retEnviNFe>`))

	cfg := configDeTransmissao() // SP, cUF 35
	cfg.VersaoDF = Ve200         // o COMPONENTE esta na 2.00
	c := componenteApontandoPara(sefaz, cfg)

	// A NOTA e 3.10 e de Minas Gerais -- os dois eixos divergem da
	// configuracao de proposito.
	nota := notaParaEnvio(chaveLote1, 31, 55, 3.10, pcn.TaHomologacao)

	destino := DestinoEnvio{URL: sefaz.servidor.URL, Versao: Ve310,
		Modelo: MoNFe, UF: "MG", Autorizacao: true}
	if _, err := c.enviarNaURL(context.Background(), destino,
		[]*NotaFiscal{nota}, OpcoesEnvio{Lote: "1"}); err != nil {
		t.Fatal(err)
	}

	cab := recortarCabecalho(t, sefaz.ultimoCorpo)
	if !strings.Contains(cab, "<versaoDados>3.10</versaoDados>") {
		t.Errorf("versaoDados = %q; deveria seguir a versao do DOCUMENTO", cab)
	}

	// O cUF, ao contrario da versao, NAO vem do documento: TNFeRecepcao nao
	// sobrescreve GerarUFSoap, e o do TDFeWebService le a configuracao
	// (ACBrDFeWebService.pas:349). Os dois eixos do cabecalho tem origens
	// diferentes, e e isso que o teste trava.
	if !strings.Contains(cab, "<cUF>35</cUF>") {
		t.Errorf("o cUF do cabecalho vem da CONFIGURACAO (35), nao do documento (31): %q", cab)
	}
}

// TestMontarEnviNFe_IndSincTemCondicaoComposta trava a condicao de :1146.
//
// Sao tres alternativas em OU, e a mais facil de perder e a primeira: o
// servico ser o de autorizacao. Um lote 2.00 de NF-e -- unico caso em que
// as tres falham -- sai SEM indSinc, e e a unica configuracao em que a
// presenca da tag causaria rejeicao de schema.
func TestMontarEnviNFe_IndSincTemCondicaoComposta(t *testing.T) {
	casos := []struct {
		modelo      ModeloDF
		versao      VersaoDF
		autorizacao bool
		temIndSinc  bool
		descricao   string
	}{
		{MoNFe, Ve200, false, false, "NF-e 2.00 no lote antigo: sem indSinc"},
		{MoNFe, Ve200, true, true, "autorizacao sempre traz indSinc"},
		{MoNFe, Ve310, false, true, "3.10 traz indSinc mesmo fora da autorizacao"},
		{MoNFCe, Ve200, false, true, "NFC-e traz indSinc em qualquer versao"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			cfg := configDeTransmissao()
			cfg.ModeloDF = c.modelo
			componente := NovoComponente(cfg)

			nota := notaParaEnvio(chaveLote1, 35, 55, c.versao.Float(), pcn.TaHomologacao)
			xml, err := componente.montarEnviNFe([]*NotaFiscal{nota},
				OpcoesEnvio{Lote: "1"}, c.versao, c.autorizacao)
			if err != nil {
				t.Fatal(err)
			}

			if got := strings.Contains(xml, "<indSinc>"); got != c.temIndSinc {
				t.Errorf("indSinc presente = %v, esperado %v: %s", got, c.temIndSinc, xml)
			}
		})
	}
}

// TestMontarEnviNFe_RecortaONFeDoXMLAssinado trava o recorte de :1152.
//
// O XML assinado pode vir envelopado num nfeProc -- uma nota recarregada do
// disco vem assim. Enviar o envelope inteiro faz a SEFAZ rejeitar o lote
// por schema.
func TestMontarEnviNFe_RecortaONFeDoXMLAssinado(t *testing.T) {
	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	nota.XMLAssinado = `<nfeProc versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		nota.XMLAssinado[strings.Index(nota.XMLAssinado, "<NFe"):] +
		`<protNFe><infProt><nProt>1</nProt></infProt></protNFe></nfeProc>`

	componente := NovoComponente(configDeTransmissao())
	xml, err := componente.montarEnviNFe([]*NotaFiscal{nota},
		OpcoesEnvio{Lote: "1"}, Ve400, true)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(xml, "nfeProc") {
		t.Errorf("o envelope nfeProc nao pode ir no lote: %s", xml)
	}
	if strings.Contains(xml, "protNFe") {
		t.Errorf("o protocolo nao pode ir no lote: %s", xml)
	}
	if !strings.Contains(xml, "<SignatureValue>AAA</SignatureValue>") {
		t.Error("a assinatura da nota tem de ser preservada")
	}
}

// TestMontarEnviNFe_LimiteDeTamanhoEhMedidoDepoisDaCompressao trava a
// ordem das duas operacoes em :1156-1161.
//
// O ACBr comprime e SO ENTAO mede. Medir antes recusaria um lote que
// caberia -- e e a inversao natural, porque a validacao parece pertencer a
// montagem.
func TestMontarEnviNFe_LimiteDeTamanhoEhMedidoDepoisDaCompressao(t *testing.T) {
	// Uma nota com um bloco grande e MUITO repetitivo: o gzip a reduz a uma
	// fracao, e so o lote comprimido cabe no limite.
	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	nota.XMLAssinado = `<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe Id="NFe` +
		chaveLote1 + `"><infAdic><infCpl>` + strings.Repeat("A", 900*1024) +
		`</infCpl></infAdic></infNFe></NFe>`

	componente := NovoComponente(configDeTransmissao())

	_, err := componente.montarEnviNFe([]*NotaFiscal{nota},
		OpcoesEnvio{Lote: "1"}, Ve400, true)
	if !errors.Is(err, ErrLoteExcedido) {
		t.Fatalf("sem compressao o lote excede o limite: err = %v", err)
	}

	comprimido, err := componente.montarEnviNFe([]*NotaFiscal{nota},
		OpcoesEnvio{Lote: "1", Zipado: true}, Ve400, true)
	if err != nil {
		t.Fatalf("comprimido o mesmo lote cabe: %v", err)
	}
	if len(comprimido) >= tamanhoMaximoLote {
		t.Fatalf("o lote comprimido deveria caber: %d bytes", len(comprimido))
	}
}

// TestMontarEnviNFe_ZipadoPrefixaADeclaracaoXML trava o
// CUTF8DeclaracaoXML de :1156.
//
// O ACBr comprime a declaracao JUNTO com a mensagem. Sem ela a SEFAZ nao
// decodifica o lote -- e o teste so pega isso descomprimindo.
func TestMontarEnviNFe_ZipadoPrefixaADeclaracaoXML(t *testing.T) {
	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	componente := NovoComponente(configDeTransmissao())

	comprimido, err := componente.montarEnviNFe([]*NotaFiscal{nota},
		OpcoesEnvio{Lote: "1", Zipado: true}, Ve400, true)
	if err != nil {
		t.Fatal(err)
	}

	original, err := descomprimirLote(comprimido)
	if err != nil {
		t.Fatalf("o conteudo deveria ser gzip em base64: %v", err)
	}
	if !strings.HasPrefix(original, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Errorf("o conteudo comprimido tem de comecar pela declaracao XML: %.60q", original)
	}
	if !strings.Contains(original, "<enviNFe") {
		t.Errorf("o conteudo comprimido deveria ser o enviNFe: %.80q", original)
	}
}

// TestServicoEActionDaAutorizacao_SufixoDoZipMudaComAVersao trava :1113.
//
// Sao dois sufixos diferentes para a mesma opcao, e a escolha e por versao:
// "ZIP" da 4.00 em diante, "LoteZip" antes. Usar um so manda o lote para um
// metodo que a SEFAZ nao expoe.
func TestServicoEActionDaAutorizacao_SufixoDoZipMudaComAVersao(t *testing.T) {
	casos := []struct {
		versao      VersaoDF
		autorizacao bool
		zipado      bool
		servico     string
		acao        string
		descricao   string
	}{
		{Ve400, true, false, "NFeAutorizacao4", "NFeAutorizacao4/nfeAutorizacaoLote", "4.00 sem zip"},
		{Ve400, true, true, "NFeAutorizacao4", "NFeAutorizacao4/nfeAutorizacaoLoteZIP", "4.00 com zip"},
		{Ve310, true, false, "NfeAutorizacao", "NfeAutorizacao", "3.10 sem zip"},
		{Ve310, true, true, "NfeAutorizacao", "NfeAutorizacaoLoteZip", "3.10 com zip"},
		{Ve200, false, false, "NfeRecepcao2", "NfeRecepcao2", "lote antigo"},
		{Ve200, false, true, "NfeRecepcao2", "NfeRecepcao2LoteZip", "lote antigo com zip"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			servico, acao := servicoEActionDaAutorizacao(c.versao, c.autorizacao, c.zipado)
			if servico != enderecoWsd+c.servico {
				t.Errorf("servico = %q, esperado %q", servico, enderecoWsd+c.servico)
			}
			if acao != enderecoWsd+c.acao {
				t.Errorf("SoapAction = %q, esperada %q", acao, enderecoWsd+c.acao)
			}
		})
	}
}

// TestServicoEActionDaRetAutorizacao trava o par da consulta do recibo
// (:1553).
func TestServicoEActionDaRetAutorizacao(t *testing.T) {
	casos := []struct {
		versao      VersaoDF
		autorizacao bool
		servico     string
		acao        string
	}{
		{Ve400, true, "NFeRetAutorizacao4", "NFeRetAutorizacao4/nfeRetAutorizacaoLote"},
		{Ve310, true, "NfeRetAutorizacao", "NfeRetAutorizacao"},
		{Ve200, false, "NfeRetRecepcao2", "NfeRetRecepcao2"},
	}

	for _, c := range casos {
		servico, acao := servicoEActionDaRetAutorizacao(c.versao, c.autorizacao)
		if servico != enderecoWsd+c.servico || acao != enderecoWsd+c.acao {
			t.Errorf("versao %v autorizacao %v: servico %q acao %q, esperado %q / %q",
				c.versao, c.autorizacao, servico, acao,
				enderecoWsd+c.servico, enderecoWsd+c.acao)
		}
	}
}

// TestTratarRespostaDoEnvio_RamoSincronoTemCondicaoComposta trava :1191.
//
// A condicao e `(modelo 65 OU versao >= 3.10) E Sincrono`. O caso que
// revela um porte incompleto e o ultimo: NF-e 2.00 pedida como sincrona
// ainda volta como RECIBO, porque o servico antigo nao sabe responder
// sincrono.
func TestTratarRespostaDoEnvio_RamoSincronoTemCondicaoComposta(t *testing.T) {
	respostaRecibo := `<retEnviNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>35</cUF>` +
		`<infRec><nRec>351000000000001</nRec><tMed>1</tMed></infRec></retEnviNFe>`

	casos := []struct {
		modelo    ModeloDF
		versao    VersaoDF
		sincrono  bool
		ehSync    bool
		descricao string
	}{
		{MoNFe, Ve400, true, true, "NF-e 4.00 sincrona"},
		{MoNFe, Ve400, false, false, "NF-e 4.00 assincrona"},
		{MoNFCe, Ve200, true, true, "NFC-e sincrona em qualquer versao"},
		{MoNFe, Ve200, true, false, "NF-e 2.00 sincrona ainda volta como recibo"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			cfg := configDeTransmissao()
			cfg.ModeloDF = c.modelo
			componente := NovoComponente(cfg)

			res, err := componente.tratarRespostaDoEnvio(respostaRecibo, nil, c.sincrono, DestinoEnvio{Versao: c.versao, Modelo: MoNFe})
			if err != nil {
				t.Fatal(err)
			}
			if res.Sincrono != c.ehSync {
				t.Errorf("Sincrono = %v, esperado %v", res.Sincrono, c.ehSync)
			}
			if c.ehSync && res.Processamento == nil {
				t.Error("o ramo sincrono devolve Processamento")
			}
			if !c.ehSync && res.Recibo == nil {
				t.Error("o ramo assincrono devolve Recibo")
			}
		})
	}
}

// TestTratarRespostaDoEnvio_RenomeiaORetornoSincrono trava o StringReplace
// de :1193-1200.
//
// A SEFAZ responde ao envio sincrono com retEnviNFe, mas o leitor que o
// ACBr usa e o do retConsSitNFe. Sem a troca de nome o leitor acha a raiz
// errada e devolve um retorno vazio -- sem erro algum.
func TestTratarRespostaDoEnvio_RenomeiaORetornoSincrono(t *testing.T) {
	resposta := `<retEnviNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><verAplic>SP_v4</verAplic><cStat>104</cStat>` +
		`<xMotivo>Lote processado</xMotivo><cUF>35</cUF>` +
		`<protNFe versao="4.00"><infProt><tpAmb>2</tpAmb>` +
		`<chNFe>` + chaveLote1 + `</chNFe><nProt>135260000000001</nProt>` +
		`<digVal>abc</digVal><cStat>100</cStat>` +
		`<xMotivo>Autorizado o uso da NF-e</xMotivo></infProt></protNFe>` +
		`</retEnviNFe>`

	componente := NovoComponente(configDeTransmissao())
	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)

	res, err := componente.tratarRespostaDoEnvio(resposta, []*NotaFiscal{nota}, true, DestinoEnvio{Versao: Ve400, Modelo: MoNFe})
	if err != nil {
		t.Fatal(err)
	}
	if res.Processamento == nil {
		t.Fatal("o ramo sincrono devolve Processamento")
	}
	if res.Processamento.CStat != 104 {
		t.Errorf("cStat = %d, esperado 104 -- o retEnviNFe nao foi renomeado",
			res.Processamento.CStat)
	}
	if res.Processamento.ProtNFe == nil {
		t.Fatal("o protocolo deveria ter sido lido")
	}
	if res.Processamento.ProtNFe.NProt != "135260000000001" {
		t.Errorf("nProt = %q", res.Processamento.ProtNFe.NProt)
	}

	// A renomeacao nao muda o parse -- nenhum dos dois leitores confere o
	// nome da raiz --, mas muda o que o consumidor arquiva.
	if strings.Contains(res.XMLRetorno, "retEnviNFe") {
		t.Errorf("o XML arquivado ainda traz retEnviNFe: %.80q", res.XMLRetorno)
	}
	if !strings.Contains(res.XMLRetorno, "<retConsSitNFe") {
		t.Errorf("o XML arquivado deveria ser um retConsSitNFe: %.80q", res.XMLRetorno)
	}
	if !strings.Contains(res.XMLRetorno, "</retConsSitNFe>") {
		t.Errorf("o fechamento tambem tem de ser renomeado: %.80q", res.XMLRetorno)
	}
}

// TestTratarRespostaDoEnvio_RenomeiaTambemORetConsReci trava o SEGUNDO
// StringReplace de :1196: ha UF que responde o envio sincrono com
// retConsReciNFe, e nao com retEnviNFe.
func TestTratarRespostaDoEnvio_RenomeiaTambemORetConsReci(t *testing.T) {
	resposta := `<retConsReciNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><cStat>104</cStat><xMotivo>Lote processado</xMotivo>` +
		`<cUF>35</cUF></retConsReciNFe>`

	componente := NovoComponente(configDeTransmissao())
	res, err := componente.tratarRespostaDoEnvio(resposta, nil, true, DestinoEnvio{Versao: Ve400, Modelo: MoNFe})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(res.XMLRetorno, "retConsReciNFe") {
		t.Errorf("o retConsReciNFe tambem tem de ser renomeado: %.80q", res.XMLRetorno)
	}
	if !strings.Contains(res.XMLRetorno, "<retConsSitNFe") {
		t.Errorf("XML arquivado = %.80q", res.XMLRetorno)
	}
}

// TestResultadoEnvio_AssincronoArquivaORetornoComoVeio trava que o ramo
// assincrono NAO renomeia nada: la o documento ja e um retEnviNFe, e
// renomea-lo produziria um arquivo que nenhum schema aceita.
func TestResultadoEnvio_AssincronoArquivaORetornoComoVeio(t *testing.T) {
	resposta := `<retEnviNFe versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>35</cUF>` +
		`<infRec><nRec>351000000000001</nRec><tMed>1</tMed></infRec></retEnviNFe>`

	componente := NovoComponente(configDeTransmissao())
	res, err := componente.tratarRespostaDoEnvio(resposta, nil, false, DestinoEnvio{Versao: Ve400, Modelo: MoNFe})
	if err != nil {
		t.Fatal(err)
	}
	if res.XMLRetorno != resposta {
		t.Errorf("o ramo assincrono arquiva a resposta intacta: %q", res.XMLRetorno)
	}
}

// TestAnexarProtocoloSincrono_CondicaoDupla trava :1233.
//
// O protocolo so e anexado quando o cStat EXTERNO e 104 E o do PROTOCOLO
// esta em CStatProcessado. Olhar so um dos dois anexa protocolo de
// rejeicao, e a nota sai com um nfeProc que a SEFAZ nao reconhece.
func TestAnexarProtocoloSincrono_CondicaoDupla(t *testing.T) {
	casos := []struct {
		cStatLote int
		cStatProt int
		anexa     bool
		descricao string
	}{
		{104, 100, true, "lote processado e nota autorizada"},
		{104, 110, false, "lote processado mas nota DENEGADA"},
		{104, 539, false, "lote processado mas nota rejeitada"},
		{103, 100, false, "lote so recebido: protocolo ainda nao vale"},
		{104, 120, true, "autorizacao por DPEC tambem anexa"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
			ret := &RetConsSitNFe{
				CStat: c.cStatLote,
				ProtNFe: &ProtNFe{
					ChNFe: chaveLote1,
					NProt: "135260000000001",
					CStat: c.cStatProt,
				},
			}

			anexarProtocoloSincrono(ret, []*NotaFiscal{nota})

			if anexou := nota.NFe.ProcNFe.NProt != ""; anexou != c.anexa {
				t.Errorf("protocolo anexado = %v, esperado %v", anexou, c.anexa)
			}
		})
	}
}

// TestAnexarProtocoloSincrono_CasaPelaChave trava a comparacao de :1245.
//
// O protocolo vai para a nota cuja chave bate, nao para a primeira do
// lote. Com uma nota so o erro nao aparece -- por isso o teste usa duas.
func TestAnexarProtocoloSincrono_CasaPelaChave(t *testing.T) {
	primeira := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	segunda := notaParaEnvio(chaveLote2, 35, 55, 4.00, pcn.TaHomologacao)

	ret := &RetConsSitNFe{
		CStat: 104,
		ProtNFe: &ProtNFe{
			ChNFe: chaveLote2, // o protocolo e da SEGUNDA
			NProt: "135260000000002",
			CStat: 100,
		},
	}

	anexarProtocoloSincrono(ret, []*NotaFiscal{primeira, segunda})

	if primeira.NFe.ProcNFe.NProt != "" {
		t.Errorf("a primeira nota nao deveria receber protocolo: %q",
			primeira.NFe.ProcNFe.NProt)
	}
	if segunda.NFe.ProcNFe.NProt != "135260000000002" {
		t.Errorf("a segunda nota deveria receber o protocolo, recebeu %q",
			segunda.NFe.ProcNFe.NProt)
	}
}

// TestEnviar_RecusaLoteAcimaDoLimite trava os dois tetos de :1013 e :1021.
//
// Sao limites DIFERENTES por modelo, e o da NFC-e vem de uma nota tecnica
// recente: uma nota por chamada.
func TestEnviar_RecusaLoteAcimaDoLimite(t *testing.T) {
	ctx := context.Background()

	t.Run("NF-e acima de 50", func(t *testing.T) {
		c := NovoComponente(configDeTransmissao())
		var notas []*NotaFiscal
		for i := 0; i < 51; i++ {
			notas = append(notas, notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao))
		}
		if _, err := c.Enviar(ctx, notas, OpcoesEnvio{Lote: "1"}); !errors.Is(err, ErrLoteExcedido) {
			t.Errorf("err = %v, esperado ErrLoteExcedido", err)
		}
	})

	t.Run("NFC-e acima de 1", func(t *testing.T) {
		cfg := configDeTransmissao()
		cfg.ModeloDF = MoNFCe
		c := NovoComponente(cfg)
		notas := []*NotaFiscal{
			notaParaEnvio(chaveLote1, 35, 65, 4.00, pcn.TaHomologacao),
			notaParaEnvio(chaveLote2, 35, 65, 4.00, pcn.TaHomologacao),
		}
		if _, err := c.Enviar(ctx, notas, OpcoesEnvio{Lote: "1"}); !errors.Is(err, ErrLoteExcedido) {
			t.Errorf("err = %v, esperado ErrLoteExcedido", err)
		}
	})

	t.Run("NF-e com 50 passa do limite", func(t *testing.T) {
		c := NovoComponente(configDeTransmissao())
		var notas []*NotaFiscal
		for i := 0; i < 50; i++ {
			notas = append(notas, notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao))
		}
		// Sem certificado a chamada para adiante, no clienteSOAP: o que
		// importa e NAO ser ErrLoteExcedido.
		if _, err := c.Enviar(ctx, notas, OpcoesEnvio{Lote: "1"}); errors.Is(err, ErrLoteExcedido) {
			t.Error("cinquenta notas estao no limite, nao acima dele")
		}
	})
}

// TestEnviar_AmbienteDaNotaTemDeBaterComOConfigurado trava o
// CErroAmbienteDiferente de :1053.
//
// Uma nota de homologacao transmitida para producao e rejeicao certa. O
// componente antecipa o erro, e e a unica validacao que ele faz sobre o
// conteudo do documento antes de transmitir.
func TestEnviar_AmbienteDaNotaTemDeBaterComOConfigurado(t *testing.T) {
	cfg := configDeTransmissao() // homologacao
	c := NovoComponente(cfg)

	nota := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaProducao)

	_, err := c.Enviar(context.Background(), []*NotaFiscal{nota}, OpcoesEnvio{Lote: "1"})
	if !errors.Is(err, ErrAmbienteDiferente) {
		t.Errorf("err = %v, esperado ErrAmbienteDiferente", err)
	}
}

// TestEnviar_ContingenciaSVCDesviaOEndereco trava :1066-1073.
//
// Em SVC a nota passa a ser autorizada pelo ambiente virtual, nao pela
// SEFAZ de origem. Sem o desvio o lote vai para a SEFAZ que esta fora do ar
// -- que e justamente a razao de estar em contingencia.
func TestEnviar_ContingenciaSVCDesviaOEndereco(t *testing.T) {
	semContingencia, err := URLServico(MoNFe, ufDeAutorizacao(35, pcn.TeNormal),
		pcn.TaHomologacao, ServicoAutorizacao, Ve400)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		forma pcn.TipoEmissao
		uf    string
	}{
		{pcn.TeSVCAN, "SVC-AN"},
		{pcn.TeSVCRS, "SVC-RS"},
	} {
		if got := ufDeAutorizacao(35, c.forma); got != c.uf {
			t.Errorf("forma %v: UF = %q, esperada %q", c.forma, got, c.uf)
		}
		url, err := URLServico(MoNFe, ufDeAutorizacao(35, c.forma),
			pcn.TaHomologacao, ServicoAutorizacao, Ve400)
		if err != nil {
			t.Fatal(err)
		}
		if url == semContingencia {
			t.Errorf("forma %v deveria mudar o endereco, continuou em %q", c.forma, url)
		}
	}
}

// TestConsultarRecibo_MensagemEEndereco trava o consReciNFe e o par
// servico/action da consulta de recibo.
func TestConsultarRecibo_MensagemEEndereco(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeRetAutorizacaoLoteResult",
		montarRetConsReci(104, chaveLote1)))

	c := componenteApontandoPara(sefaz, configDeTransmissao())

	lote := ReciboDoLote{Recibo: "351000000000001", Modelo: MoNFe, Versao: Ve400, CUF: 35}
	ret, err := c.consultarReciboNaURL(context.Background(), sefaz.servidor.URL,
		lote, "SP", true)
	if err != nil {
		t.Fatal(err)
	}
	if !ret.LoteProcessado() {
		t.Error("cStat 104 e lote processado")
	}

	if !strings.Contains(sefaz.ultimoCorpo, "<consReciNFe") {
		t.Errorf("a mensagem deveria ser um consReciNFe: %s", sefaz.ultimoCorpo)
	}
	if !strings.Contains(sefaz.ultimoCorpo, "<nRec>351000000000001</nRec>") {
		t.Errorf("o recibo deveria ir na mensagem: %s", sefaz.ultimoCorpo)
	}
	if !strings.Contains(sefaz.ultimaAcao, "NFeRetAutorizacao4/nfeRetAutorizacaoLote") {
		t.Errorf("SoapAction = %q", sefaz.ultimaAcao)
	}
}

// TestConsultarRecibo_ReciboVazioNaoTransmite trava a guarda de entrada.
func TestConsultarRecibo_ReciboVazioNaoTransmite(t *testing.T) {
	c := NovoComponente(configDeTransmissao())
	lote := ReciboDoLote{Recibo: "   ", Modelo: MoNFe, Versao: Ve400, CUF: 35}
	if _, err := c.ConsultarRecibo(context.Background(), lote); !errors.Is(err, ErrReciboVazio) {
		t.Errorf("err = %v, esperado ErrReciboVazio", err)
	}
}

// TestInutilizar_UsaAUFDaConfiguracaoMesmoEmContingencia trava a
// DIVERGENCIA entre a inutilizacao e a autorizacao.
//
// O DefinirURL da inutilizacao (:2672) usa `WebServices.UF` cru, sem o case
// de FormaEmissao que a autorizacao tem. Padronizar os dois -- a
// "correcao" que um revisor faria -- mandaria o pedido de inutilizacao
// para o ambiente virtual, que nao o atende.
func TestInutilizar_UsaAUFDaConfiguracaoMesmoEmContingencia(t *testing.T) {
	cfg := configDeTransmissao() // SP
	cfg.FormaEmissao = pcn.TeSVCAN

	daConfiguracao, err := URLServico(MoNFe, cfg.UF, cfg.Ambiente,
		ServicoInutilizacao, cfg.VersaoDF)
	if err != nil {
		t.Fatal(err)
	}
	daContingencia, err := URLServico(MoNFe, UFParaBusca(cfg), cfg.Ambiente,
		ServicoInutilizacao, cfg.VersaoDF)
	if err != nil {
		t.Fatal(err)
	}
	if daConfiguracao == daContingencia {
		t.Fatal("o caso de teste exige que os dois enderecos difiram")
	}

	escolhido, err := NovoComponente(cfg).urlDaInutilizacao(MoNFe)
	if err != nil {
		t.Fatal(err)
	}
	if escolhido != daConfiguracao {
		t.Errorf("a inutilizacao escolheu %q; deveria usar a UF da configuracao (%q)",
			escolhido, daConfiguracao)
	}
}

// TestEnviar_EInutilizar_EscolhemUFsDiferentesEmContingencia poe as duas
// decisoes lado a lado: e o contraste que torna a divergencia visivel.
func TestEnviar_EInutilizar_EscolhemUFsDiferentesEmContingencia(t *testing.T) {
	cfg := configDeTransmissao() // SP, cUF 35
	cfg.FormaEmissao = pcn.TeSVCRS

	if got := ufDeAutorizacao(35, cfg.FormaEmissao); got != "SVC-RS" {
		t.Errorf("a autorizacao desvia para SVC-RS, foi para %q", got)
	}

	escolhida, err := NovoComponente(cfg).urlDaInutilizacao(MoNFe)
	if err != nil {
		t.Fatal(err)
	}
	naSVC, err := URLServico(MoNFe, "SVC-RS", cfg.Ambiente,
		ServicoInutilizacao, cfg.VersaoDF)
	if err != nil {
		t.Fatal(err)
	}
	if escolhida == naSVC {
		t.Error("a inutilizacao NAO desvia para SVC; ela ficou com o endereco da SVC-RS")
	}
}

// TestServicoEActionDaInutilizacao_ExcecaoDaBahia trava :2685.
//
// A Bahia tem notacao propria de servicos na 3.10, e a condicao tem QUATRO
// fatores: UF, modelo, versao e forma de emissao. Afrouxar qualquer um
// manda outras UFs para um endereco que so a Bahia expoe.
func TestServicoEActionDaInutilizacao_ExcecaoDaBahia(t *testing.T) {
	casos := []struct {
		cUF       int
		modelo    ModeloDF
		versao    VersaoDF
		forma     pcn.TipoEmissao
		servico   string
		acao      string
		descricao string
	}{
		{29, MoNFe, Ve310, pcn.TeNormal, "NfeInutilizacao",
			"NfeInutilizacao/NfeInutilizacao", "BA 3.10 normal: notacao propria"},
		{29, MoNFe, Ve400, pcn.TeNormal, "NFeInutilizacao4",
			"NFeInutilizacao4/nfeInutilizacaoNF", "BA na 4.00 volta ao padrao"},
		{29, MoNFCe, Ve310, pcn.TeNormal, "NfeInutilizacao2",
			"NfeInutilizacao2", "BA 3.10 mas NFC-e: padrao"},
		{29, MoNFe, Ve310, pcn.TeSVCAN, "NfeInutilizacao2",
			"NfeInutilizacao2", "BA 3.10 em contingencia: padrao"},
		{35, MoNFe, Ve310, pcn.TeNormal, "NfeInutilizacao2",
			"NfeInutilizacao2", "SP 3.10: padrao"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			cfg := configDeTransmissao()
			cfg.CodigoUF, cfg.VersaoDF, cfg.FormaEmissao = c.cUF, c.versao, c.forma
			componente := NovoComponente(cfg)

			servico, acao := componente.servicoEActionDaInutilizacao(c.modelo)
			if servico != enderecoWsd+c.servico {
				t.Errorf("servico = %q, esperado %q", servico, enderecoWsd+c.servico)
			}
			if acao != enderecoWsd+c.acao {
				t.Errorf("SoapAction = %q, esperada %q", acao, enderecoWsd+c.acao)
			}
		})
	}
}

// TestComOverrideDoIni_SoVenceNaVersao400 trava o guard `EstaVazio` que
// existe apenas no ramo da 4.00 de todo DefinirServicoEAction.
//
// Abaixo da 4.00 o codigo atribui a SoapAction sem guard, descartando o que
// veio do `.ini`. Aplicar o override em qualquer versao e o porte natural, e
// esta errado.
func TestComOverrideDoIni_SoVenceNaVersao400(t *testing.T) {
	const doCodigo = "http://exemplo/do-codigo"
	const doIni = "http://exemplo/do-ini"

	for _, c := range []struct {
		versao   VersaoDF
		esperado string
	}{
		{Ve400, doIni},
		{Ve310, doCodigo},
		{Ve200, doCodigo},
	} {
		if got := comOverrideDoIni(doCodigo, doIni, c.versao); got != c.esperado {
			t.Errorf("versao %v: %q, esperado %q", c.versao, got, c.esperado)
		}
	}

	// Sem override declarado, o valor do codigo permanece em qualquer versao.
	if got := comOverrideDoIni(doCodigo, "", Ve400); got != doCodigo {
		t.Errorf("sem override o valor do codigo permanece: %q", got)
	}
}

// TestInutilizar_SoapActionDaBahiaNa400VemDoIni trava a ponta a ponta do
// override: o `.ini` da Bahia declara NFeInutilizacao4 com um caminho que
// coincide com o do codigo, mas a tabela e quem manda.
func TestInutilizar_SoapActionDaBahiaNa400VemDoIni(t *testing.T) {
	doIni := SoapActionDoIni(MoNFe, "BA", pcn.TaProducao, ServicoInutilizacao, Ve400)
	if doIni == "" {
		t.Fatal("o .ini da Bahia declara NfeInutilizacao_4.00; a tabela perdeu a entrada")
	}

	// E na 3.10 nao ha override nenhum -- a tabela so tem entradas _4.00.
	if a := SoapActionDoIni(MoNFe, "BA", pcn.TaProducao, ServicoInutilizacao, Ve310); a != "" {
		t.Errorf("nao deveria haver override de 3.10 para a BA: %q", a)
	}
}

// TestConsultarCadastro_MatoGrossoEmbrulhaAMensagem trava :2958 e :3019.
//
// O Mato Grosso, da 4.00 em diante, espera a mensagem dentro de um
// <nfeDadosMsg> EXTRA e com o elemento do Body renomeado para
// <consultaCadastro>. Sao duas mudancas na MESMA condicao, e o teste afirma
// sobre o ANINHAMENTO -- contar ocorrencias de nfeDadosMsg nao distingue os
// dois casos, porque o total e um nos dois.
func TestConsultarCadastro_MatoGrossoEmbrulhaAMensagem(t *testing.T) {
	resposta := `<retConsCad versao="2.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<infCons><verAplic>MT_v4</verAplic><cStat>111</cStat>` +
		`<xMotivo>Consulta cadastro com uma ocorrencia</xMotivo>` +
		`<UF>MT</UF><IE>13000000</IE><cUF>51</cUF>` +
		`<infCad><IE>13000000</IE><CNPJ>99999999000191</CNPJ><UF>MT</UF>` +
		`<cSit>1</cSit><xNome>EMPRESA TESTE</xNome></infCad></infCons></retConsCad>`

	casos := []struct {
		uf        string
		versao    VersaoDF
		aninhado  bool
		descricao string
	}{
		{"MT", Ve400, true, "MT na 4.00: camada extra"},
		{"MT", Ve310, false, "MT na 3.10: sem camada extra"},
		{"SP", Ve400, false, "SP na 4.00: sem camada extra"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			sefaz := novaSefazFalsa(t, envelopeSOAP("consultaCadastro4Result", resposta))
			cfg := configDeTransmissao()
			cfg.VersaoDF = c.versao
			componente := componenteApontandoPara(sefaz, cfg)

			if _, err := componente.consultarCadastroNaURL(context.Background(),
				sefaz.servidor.URL, c.uf, "13000000", "", ""); err != nil {
				t.Fatal(err)
			}
			corpo := sefaz.ultimoCorpo

			// Com a camada extra o <nfeDadosMsg> interno vem SEM atributo
			// nenhum e colado no <ConsCad>; sem ela o <nfeDadosMsg> e o
			// proprio elemento do Body e carrega o xmlns do servico.
			temCamadaExtra := strings.Contains(corpo, "<nfeDadosMsg><ConsCad")
			if temCamadaExtra != c.aninhado {
				t.Errorf("camada extra = %v, esperado %v: %s", temCamadaExtra, c.aninhado, corpo)
			}

			temElementoProprio := strings.Contains(corpo, "<consultaCadastro ")
			if temElementoProprio != c.aninhado {
				t.Errorf("elemento <consultaCadastro> do Body = %v, esperado %v: %s",
					temElementoProprio, c.aninhado, corpo)
			}
		})
	}
}

// TestGerarXMLConsCad_PrecedenciaDosDocumentos trava a escolha exclusiva do
// TConsCad.GerarXML.
//
// So UM documento sai no XML, e a ordem e IE, CNPJ, CPF. Emitir mais de um
// e rejeicao de schema; emitir o errado consulta outro contribuinte.
func TestGerarXMLConsCad_PrecedenciaDosDocumentos(t *testing.T) {
	casos := []struct {
		ie, cnpj, cpf string
		esperado      string
		ausentes      []string
	}{
		{"13000000", "99999999000191", "11122233396", "<IE>13000000</IE>",
			[]string{"<CNPJ>", "<CPF>"}},
		{"", "99999999000191", "11122233396", "<CNPJ>99999999000191</CNPJ>",
			[]string{"<IE>", "<CPF>"}},
		{"", "", "11122233396", "<CPF>11122233396</CPF>",
			[]string{"<IE>", "<CNPJ>"}},
	}

	for _, c := range casos {
		xml := GerarXMLConsCad("SP", c.ie, c.cnpj, c.cpf, "2.00")
		if !strings.Contains(xml, c.esperado) {
			t.Errorf("esperado %s em %s", c.esperado, xml)
		}
		for _, ausente := range c.ausentes {
			if strings.Contains(xml, ausente) {
				t.Errorf("%s nao deveria aparecer em %s", ausente, xml)
			}
		}
	}
}

// TestLerRetConsCad_PreencheZerosAEsquerdaNoDocumento trava :210-219.
//
// Ha web service que devolve o CNPJ sem os zeros a esquerda. Sem o
// preenchimento, quem comparar o retorno com o cadastro proprio conclui que
// sao contribuintes diferentes.
func TestLerRetConsCad_PreencheZerosAEsquerdaNoDocumento(t *testing.T) {
	xml := `<retConsCad versao="2.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<infCons><CNPJ>9999999000191</CNPJ><cStat>111</cStat><UF>SP</UF>` +
		`<infCad><CNPJ>9999999000191</CNPJ><CPF>1122233396</CPF>` +
		`<IE>111111111111</IE><UF>SP</UF><cSit>1</cSit>` +
		`<ender><xLgr>RUA TESTE</xLgr><nro>100</nro><cMun>3550308</cMun>` +
		`<xMun>SAO PAULO</xMun><CEP>01001000</CEP></ender>` +
		`</infCad></infCons></retConsCad>`

	r, err := LerRetConsCad(xml)
	if err != nil {
		t.Fatal(err)
	}

	if r.CNPJ != "09999999000191" {
		t.Errorf("CNPJ do infCons = %q, esperado 09999999000191", r.CNPJ)
	}
	if len(r.InfCad) != 1 {
		t.Fatalf("cadastros lidos = %d, esperado 1", len(r.InfCad))
	}
	if r.InfCad[0].CNPJ != "09999999000191" {
		t.Errorf("CNPJ do infCad = %q, esperado 09999999000191", r.InfCad[0].CNPJ)
	}
	if r.InfCad[0].CPF != "01122233396" {
		t.Errorf("CPF do infCad = %q, esperado 01122233396", r.InfCad[0].CPF)
	}
	if !r.InfCad[0].Habilitado() {
		t.Error("cSit 1 e habilitado")
	}
	// O endereco e achatado no item, como no ACBr.
	if r.InfCad[0].XMun != "SAO PAULO" || r.InfCad[0].CEP != "01001000" {
		t.Errorf("endereco nao foi lido: %+v", r.InfCad[0])
	}
}

// TestRetConsCad_EncontradoSoNosDoisCodigos trava :2991.
func TestRetConsCad_EncontradoSoNosDoisCodigos(t *testing.T) {
	for _, c := range []struct {
		cStat      int
		encontrado bool
	}{
		{111, true},
		{112, true},
		{110, false},
		{259, false},
		{0, false},
	} {
		r := &RetConsCad{CStat: c.cStat}
		if r.Encontrado() != c.encontrado {
			t.Errorf("cStat %d: Encontrado() = %v, esperado %v",
				c.cStat, r.Encontrado(), c.encontrado)
		}
	}
}

// TestConsultarCadastro_SemDocumentoNaoTransmite trava a guarda de entrada.
//
// Divergencia deliberada: o ACBr montaria um infCons com <CPF></CPF> vazio e
// gastaria uma chamada de rede que nunca pode dar certo.
func TestConsultarCadastro_SemDocumentoNaoTransmite(t *testing.T) {
	c := NovoComponente(configDeTransmissao())
	_, err := c.ConsultarCadastro(context.Background(), "SP", "", "", "")
	if !errors.Is(err, ErrDocumentoConsultaAusente) {
		t.Errorf("err = %v, esperado ErrDocumentoConsultaAusente", err)
	}
}

// TestGerarXMLAdmCSC_DadosCscSoNaRevogacao trava :88-95.
//
// O grupo dadosCsc so e emitido quando indOp e revogacao. Emiti-lo sempre
// faz a consulta carregar um CSC que ela nao deveria mandar -- e o codigo e
// segredo.
func TestGerarXMLAdmCSC_DadosCscSoNaRevogacao(t *testing.T) {
	casos := []struct {
		indOp     IndOperacao
		temDados  bool
		descricao string
	}{
		{IoConsultaCSC, false, "consulta"},
		{IoNovoCSC, false, "requisicao de novo"},
		{IoRevogaCSC, true, "revogacao"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			p := PedidoAdmCSC{
				IndOp:     c.indOp,
				RaizCNPJ:  "99999999",
				IDCSC:     7,
				CodigoCSC: "SEGREDO",
			}
			xml := GerarXMLAdmCSC(p, pcn.TaHomologacao, "1.00")

			if got := strings.Contains(xml, "<dadosCsc>"); got != c.temDados {
				t.Errorf("dadosCsc presente = %v, esperado %v: %s", got, c.temDados, xml)
			}
			if c.temDados && !strings.Contains(xml, "<idCsc>000007</idCsc>") {
				t.Errorf("o idCsc vai com seis posicoes zero-preenchidas: %s", xml)
			}
			if !c.temDados && strings.Contains(xml, "SEGREDO") {
				t.Errorf("o codigo do CSC nao pode vazar fora da revogacao: %s", xml)
			}
		})
	}
}

// TestRetAdmCSCNFCe_O151ContaComoAtendido trava a faixa 150..153 de :3925.
//
// O 151 e "nao ha CSC ativo": a consulta FUNCIONOU, e a resposta e que nao
// ha nada. Lendo so a descricao, o porte natural o trata como falha.
func TestRetAdmCSCNFCe_O151ContaComoAtendido(t *testing.T) {
	for _, c := range []struct {
		cStat    int
		atendido bool
	}{
		{150, true},
		{151, true},
		{152, true},
		{153, true},
		{149, false},
		{154, false},
		{217, false},
	} {
		r := &RetAdmCSCNFCe{CStat: c.cStat}
		if r.Atendido() != c.atendido {
			t.Errorf("cStat %d: Atendido() = %v, esperado %v",
				c.cStat, r.Atendido(), c.atendido)
		}
	}
}

// TestAdministrarCSC_MandaCabecalhoAteNa400 trava :3940.
//
// E o unico servico da NF-e que forca o <soap12:Header> em qualquer versao.
// Reaproveitar o cabecalhoSOAP -- que o suprime da 4.00 em diante -- faria
// o servico responder erro de cabecalho ausente.
func TestAdministrarCSC_MandaCabecalhoAteNa400(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("admCscNFCeResult",
		`<retAdmCscNFCe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">`+
			`<tpAmb>2</tpAmb><indOp>1</indOp><cStat>150</cStat>`+
			`<xMotivo>CSC ativos</xMotivo>`+
			`<dadosCsc><idCsc>000001</idCsc><codigoCsc>ABC</codigoCsc></dadosCsc>`+
			`<dadosCsc><idCsc>000002</idCsc><codigoCsc>DEF</codigoCsc></dadosCsc>`+
			`</retAdmCscNFCe>`))

	cfg := configDeTransmissao()
	cfg.VersaoDF = Ve400
	cfg.ModeloDF = MoNFCe
	c := componenteApontandoPara(sefaz, cfg)

	ret, err := c.administrarCSCNaURL(context.Background(), sefaz.servidor.URL,
		PedidoAdmCSC{IndOp: IoConsultaCSC, RaizCNPJ: "99999999"})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(sefaz.ultimoCorpo, "nfeCabecMsg") {
		t.Errorf("o AdministrarCSC manda cabecalho em qualquer versao: %s", sefaz.ultimoCorpo)
	}
	// O versaoDados acompanha a versao do SERVICO (1.00), nao a do leiaute
	// -- ver TestAdministrarCSC_MensagemVaiNaVersao100.
	cab := recortarCabecalho(t, sefaz.ultimoCorpo)
	if !strings.Contains(cab, "<versaoDados>1.00</versaoDados>") {
		t.Errorf("versaoDados do cabecalho = %q", cab)
	}

	if len(ret.DadosCSC) != 2 {
		t.Fatalf("CSC lidos = %d, esperado 2", len(ret.DadosCSC))
	}
	if ret.DadosCSC[0].CodigoCSC == ret.DadosCSC[1].CodigoCSC {
		t.Error("os dois CSC deveriam ser distintos")
	}
	if !ret.Atendido() {
		t.Error("cStat 150 e atendido")
	}
}

// descomprimirLote desfaz o base64+gzip do lote zipado.
func descomprimirLote(s string) (string, error) {
	return dfe.Base64Gunzip(s)
}

// TestAdministrarCSC_ModeloVemDaConfiguracao trava a escolha do modelo na
// busca do endereco.
//
// O CSC so existe na NFC-e, mas TAdministrarCSCNFCe nao sobrescreve
// DefinirURL: o modelo da busca e o da CONFIGURACAO (:713). Fixar o 65 no
// porte seria uma gentileza que o original nao faz, e esconderia do usuario
// que o componente esta configurado para o documento errado.
//
// A UF e Goias porque so AM e GO publicam o servico -- o `.ini` declara
// AdministrarCSCNFCe em quatro linhas, duas por UF.
func TestAdministrarCSC_ModeloVemDaConfiguracao(t *testing.T) {
	cfg := configDeTransmissao()
	cfg.UF, cfg.CodigoUF = "GO", 52
	cfg.ModeloDF = MoNFe // modelo ERRADO para o servico

	_, err := NovoComponente(cfg).AdministrarCSC(context.Background(),
		PedidoAdmCSC{IndOp: IoConsultaCSC, RaizCNPJ: "99999999"})
	if !errors.Is(err, ErrSemURL) {
		t.Errorf("err = %v; o modelo 55 nao atende o AdministrarCSCNFCe", err)
	}

	// Com o modelo certo o endereco existe.
	cfg.ModeloDF = MoNFCe
	if _, err := URLServico(cfg.ModeloDF, UFParaBusca(cfg), cfg.Ambiente,
		ServicoAdministrarCSC, cfg.VersaoDF); err != nil {
		t.Errorf("sob a NFC-e o endereco deveria existir: %v", err)
	}
}

// TestVersaoDaMensagem_ServicosComVersaoPROPRIA trava os dois servicos cuja
// mensagem NAO acompanha a versao do leiaute.
//
// A resolucao no ACBr tem dois estagios: a chave do `.ini` e, depois, o
// SCHEMA instalado, que vence quando existe (ACBrDFe.pas:580-590). Para
// consStatServ, consSitNFe e inutNFe o schema da 4.00 existe e restaura a
// versao pedida; para consCad ele so existe em 2.00, e para admCscNFCe nao
// existe schema nenhum. Emitir "4.00" nesses dois e rejeicao certa, porque a
// SEFAZ valida contra o XSD do SERVICO.
func TestVersaoDaMensagem_ServicosComVersaoPROPRIA(t *testing.T) {
	casos := []struct {
		servico  Servico
		esperado string
	}{
		{ServicoConsultaCadastro, "2.00"},
		{ServicoAdministrarCSC, "1.00"},
		{ServicoStatus, "4.00"},
		{ServicoConsultaProtocolo, "4.00"},
		{ServicoInutilizacao, "4.00"},
		{ServicoAutorizacao, "4.00"},
	}
	for _, c := range casos {
		if got := VersaoDaMensagem(c.servico, Ve400); got != c.esperado {
			t.Errorf("%s com componente na 4.00 = %q, esperado %q",
				c.servico, got, c.esperado)
		}
	}

	// Os dois fixos nao mudam nem com o componente na 3.10.
	if got := VersaoDaMensagem(ServicoConsultaCadastro, Ve310); got != "2.00" {
		t.Errorf("consCad na 3.10 = %q, esperado 2.00", got)
	}
	// Os demais acompanham.
	if got := VersaoDaMensagem(ServicoStatus, Ve310); got != "3.10" {
		t.Errorf("status na 3.10 = %q, esperado 3.10", got)
	}
}

// TestConsultarCadastro_MensagemVaiNaVersao200 fecha a ponta a ponta: o
// ConsCad que sobe no fio tem de dizer 2.00, mesmo num componente 4.00.
func TestConsultarCadastro_MensagemVaiNaVersao200(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("consultaCadastro4Result",
		`<retConsCad versao="2.00" xmlns="http://www.portalfiscal.inf.br/nfe">`+
			`<infCons><cStat>111</cStat><UF>SP</UF></infCons></retConsCad>`))

	cfg := configDeTransmissao() // 4.00
	c := componenteApontandoPara(sefaz, cfg)

	if _, err := c.consultarCadastroNaURL(context.Background(),
		sefaz.servidor.URL, "SP", "111111111111", "", ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sefaz.ultimoCorpo, `<ConsCad xmlns="http://www.portalfiscal.inf.br/nfe" versao="2.00">`) {
		t.Errorf("o ConsCad deveria sair na versao 2.00: %s", sefaz.ultimoCorpo)
	}
}

// TestAdministrarCSC_MensagemVaiNaVersao100 e o par do teste acima.
func TestAdministrarCSC_MensagemVaiNaVersao100(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("admCscNFCeResult",
		`<retAdmCscNFCe versao="1.00" xmlns="http://www.portalfiscal.inf.br/nfe">`+
			`<tpAmb>2</tpAmb><indOp>1</indOp><cStat>151</cStat>`+
			`<xMotivo>Nao ha CSC ativo</xMotivo></retAdmCscNFCe>`))

	cfg := configDeTransmissao() // 4.00
	cfg.ModeloDF = MoNFCe
	c := componenteApontandoPara(sefaz, cfg)

	if _, err := c.administrarCSCNaURL(context.Background(), sefaz.servidor.URL,
		PedidoAdmCSC{IndOp: IoConsultaCSC, RaizCNPJ: "99999999"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sefaz.ultimoCorpo, `versao="1.00"`) {
		t.Errorf("o admCscNFCe deveria sair na versao 1.00: %s", sefaz.ultimoCorpo)
	}
	// E o versaoDados do cabecalho acompanha a mensagem, nao o leiaute.
	cab := recortarCabecalho(t, sefaz.ultimoCorpo)
	if !strings.Contains(cab, "<versaoDados>1.00</versaoDados>") {
		t.Errorf("versaoDados do cabecalho = %q, esperado 1.00", cab)
	}
}
