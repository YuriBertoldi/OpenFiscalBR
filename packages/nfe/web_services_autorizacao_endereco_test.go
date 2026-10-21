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
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes da DERIVACAO que antecede o envio.
//
// Os testes de enviarNaURL nao alcancam isto: la o endereco ja chega
// resolvido. O que estes exercitam e o enderecoDoEnvio, que e onde um
// refactor quebraria calado -- o lote iria para um endereco valido, so que
// o errado, e voltaria rejeitado sem erro de programa.

// TestEnderecoDoEnvio_VersaoVemDaNota trava a decisao de :1026.
func TestEnderecoDoEnvio_VersaoVemDaNota(t *testing.T) {
	cfg := configDeTransmissao()
	cfg.VersaoDF = Ve310 // o COMPONENTE esta na 3.10
	c := NovoComponente(cfg)

	destino, err := c.enderecoDoEnvio([]*NotaFiscal{
		notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)})
	if err != nil {
		t.Fatal(err)
	}
	if destino.Versao != Ve400 {
		t.Errorf("versao = %v, esperado Ve400", destino.Versao)
	}

	na310, err := URLServico(MoNFe, "SP", cfg.Ambiente, ServicoAutorizacao, Ve310)
	if err != nil {
		t.Fatal(err)
	}
	if destino.URL == na310 {
		t.Errorf("o endereco deveria ser o da 4.00, saiu o da 3.10: %q", destino.URL)
	}
}

// TestEnderecoDoEnvio_UFVemDaNota trava :1047 -- o cUF do documento, nao o
// da configuracao, decide a SEFAZ que recebe o lote.
func TestEnderecoDoEnvio_UFVemDaNota(t *testing.T) {
	cfg := configDeTransmissao() // SP
	c := NovoComponente(cfg)

	// Minas Gerais (cUF 31) opera o proprio ambiente.
	destino, err := c.enderecoDoEnvio([]*NotaFiscal{
		notaParaEnvio(chaveLote1, 31, 55, 4.00, pcn.TaHomologacao)})
	if err != nil {
		t.Fatal(err)
	}
	if destino.UF != "MG" {
		t.Errorf("UF = %q, esperada MG", destino.UF)
	}

	emSP, err := URLServico(MoNFe, "SP", cfg.Ambiente, ServicoAutorizacao, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	if destino.URL == emSP {
		t.Errorf("o lote de MG foi para o endereco de SP: %q", destino.URL)
	}
}

// TestEnderecoDoEnvio_ModeloVemDaNota trava o StrToModeloDF de :1044.
func TestEnderecoDoEnvio_ModeloVemDaNota(t *testing.T) {
	cfg := configDeTransmissao() // MoNFe
	c := NovoComponente(cfg)

	destino, err := c.enderecoDoEnvio([]*NotaFiscal{
		notaParaEnvio(chaveLote1, 35, 65, 4.00, pcn.TaHomologacao)})
	if err != nil {
		t.Fatal(err)
	}
	if destino.Modelo != MoNFCe {
		t.Errorf("modelo = %v, esperado MoNFCe", destino.Modelo)
	}

	comoNFe, err := URLServico(MoNFe, "SP", cfg.Ambiente, ServicoAutorizacao, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	if destino.URL == comoNFe {
		t.Errorf("a NFC-e foi para o endereco da NF-e: %q", destino.URL)
	}
}

// TestEnderecoDoEnvio_ExcecaoDoAmazonasChegaAoENDERECO fecha o laco do
// EhAutorizacao.
//
// A funcao pura ja esta travada; o que falta provar e que o enderecoDoEnvio
// a CONSULTA. Sem isso, a exceção do Amazonas existiria e ninguem a usaria.
func TestEnderecoDoEnvio_ExcecaoDoAmazonasChegaAoENDERECO(t *testing.T) {
	cfg := configDeTransmissao()
	cfg.ModeloDF = MoNFCe
	cfg.VersaoDF = Ve310
	c := NovoComponente(cfg)

	destino, err := c.enderecoDoEnvio([]*NotaFiscal{
		notaParaEnvio(chaveLote1, 13, 65, 3.10, pcn.TaHomologacao)})
	if err != nil {
		t.Fatal(err)
	}
	if destino.Autorizacao {
		t.Fatal("a NFC-e 3.10 do AM usa NfeRecepcao, nao NfeAutorizacao")
	}

	doLoteAntigo, err := URLServico(MoNFCe, "AM", cfg.Ambiente, ServicoRecepcaoLote, Ve310)
	if err != nil {
		t.Fatal(err)
	}
	if destino.URL != doLoteAntigo {
		t.Errorf("endereco = %q, esperado o do NfeRecepcao (%q)", destino.URL, doLoteAntigo)
	}

	// E o endereco do servico NOVO existe no AM -- ou seja, a escolha foi
	// uma decisao, nao a unica opcao disponivel.
	comAutorizacao, err := URLServico(MoNFCe, "AM", cfg.Ambiente, ServicoAutorizacao, Ve310)
	if err != nil {
		t.Fatalf("o caso de teste exige que o AM tambem publique o NfeAutorizacao: %v", err)
	}
	if destino.URL == comAutorizacao {
		t.Error("o endereco escolhido foi o do NfeAutorizacao")
	}

	// A mesma NFC-e 3.10 fora do AM vai pelo par novo.
	foraDoAM, err := c.enderecoDoEnvio([]*NotaFiscal{
		notaParaEnvio(chaveLote1, 35, 65, 3.10, pcn.TaHomologacao)})
	if err != nil {
		t.Fatal(err)
	}
	if !foraDoAM.Autorizacao {
		t.Error("so o AM ficou no par antigo")
	}
}

// TestEnviarNaURL_OverrideDeSoapActionSaiDaSecaoDoDocumento trava o eixo
// da busca do override.
//
// No ACBr a URL, o FPServico e a FPSoapAction voltam de UMA chamada a
// LerServicoDeParams (:1079-1087) -- logo, da mesma secao do `.ini`, a do
// DOCUMENTO. Buscar o override pela UF da CONFIGURACAO pega a secao errada
// quando a nota e de outra UF, e sem erro nenhum.
//
// O teste precisa INJETAR um override, e a razao e honesta: a Bahia e a
// unica UF com `SoapAction=` no `.ini`, e a action dela para a autorizacao
// e exatamente a mesma string que o codigo monta. Com os dados reais os
// dois eixos produzem o mesmo resultado, e nenhuma afirmacao sobre eles
// pode falhar. A entrada ficticia abaixo torna a diferenca observavel; ela
// vive so durante o teste.
func TestEnviarNaURL_OverrideDeSoapActionSaiDaSecaoDoDocumento(t *testing.T) {
	const acaoFicticia = "http://exemplo/secao-do-documento/nfeAutorizacaoLote"

	soapActionPorAutorizador["NFe/MG"] = map[Servico]endpointPorVersao{
		ServicoAutorizacao: {
			V400: urlPorAmbiente{acaoFicticia, acaoFicticia},
		},
	}
	t.Cleanup(func() { delete(soapActionPorAutorizador, "NFe/MG") })

	resposta := "<retEnviNFe versao=\"4.00\" xmlns=\"http://www.portalfiscal.inf.br/nfe\">" +
		"<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>31</cUF>" +
		"<infRec><nRec>311000000000001</nRec><tMed>1</tMed></infRec></retEnviNFe>"
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeAutorizacaoLoteResult", resposta))

	// Componente em SP, nota de MINAS -- a UF que acabou de ganhar o
	// override. Buscar por SP nao acharia nada.
	cfg := configDeTransmissao()
	c := componenteApontandoPara(sefaz, cfg)

	if SoapActionDoIni(MoNFe, cfg.UF, cfg.Ambiente, ServicoAutorizacao, Ve400) != "" {
		t.Fatal("o caso de teste exige que SP nao tenha override")
	}

	destino := DestinoEnvio{URL: sefaz.servidor.URL, Versao: Ve400,
		Modelo: MoNFe, UF: "MG", Autorizacao: true}
	if _, err := c.enviarNaURL(context.Background(), destino,
		[]*NotaFiscal{notaParaEnvio(chaveLote1, 31, 55, 4.00, pcn.TaHomologacao)},
		OpcoesEnvio{Lote: "1"}); err != nil {
		t.Fatal(err)
	}

	if sefaz.ultimaAcao != acaoFicticia {
		t.Errorf("SoapAction = %q; deveria vir da secao de MG (%q)",
			sefaz.ultimaAcao, acaoFicticia)
	}
}

// TestAnexarProtocolos_AnexaMesmoEmRejeicao trava a assimetria entre os
// dois ramos.
//
// No sincrono o ACBr exige cStat 104 do lote E CstatProcessado do protocolo
// (:1233); no assincrono ele grava o procNFe SEMPRE que a chave bate
// (:1667-1679), e o `[100,120,150]` ali governa so a montagem do nfeProc.
// Uniformizar os dois perde o motivo da rejeicao -- que e justamente o que
// o consumidor precisa mostrar ao usuario.
func TestAnexarProtocolos_AnexaMesmoEmRejeicao(t *testing.T) {
	ret := &RetConsReciNFe{
		CStat: 104,
		ProtNFe: []ProtDFe{
			{ChNFe: chaveLote1, NProt: "1", CStat: 100, XMotivo: "Autorizado o uso da NF-e"},
			{ChNFe: chaveLote2, NProt: "", CStat: 539, XMotivo: "Rejeicao: duplicidade de NF-e"},
		},
	}
	autorizada := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	rejeitada := notaParaEnvio(chaveLote2, 35, 55, 4.00, pcn.TaHomologacao)

	AnexarProtocolos(ret, []*NotaFiscal{autorizada, rejeitada})

	if autorizada.NFe.ProcNFe.NProt != "1" {
		t.Errorf("a nota autorizada deveria receber o protocolo, recebeu %q",
			autorizada.NFe.ProcNFe.NProt)
	}
	if rejeitada.NFe.ProcNFe.CStat != 539 {
		t.Errorf("a nota REJEITADA tambem recebe o procNFe, com o cStat da rejeicao: %d",
			rejeitada.NFe.ProcNFe.CStat)
	}
	if rejeitada.NFe.ProcNFe.XMotivo == "" {
		t.Error("o xMotivo da rejeicao e o que explica o porque ao usuario")
	}
	// Sem nProt, o GerarXML nao monta o nfeProc -- que e o outro lado da
	// regra do original.
	if rejeitada.NFe.ProcNFe.NProt != "" {
		t.Errorf("a rejeitada nao tem nProt: %q", rejeitada.NFe.ProcNFe.NProt)
	}
}

// TestAnexarProtocolos_CasaPelaChave trava a correspondencia, que no ACBr e
// um laco duplo sobre chaves (:1649) -- nunca por posicao.
func TestAnexarProtocolos_CasaPelaChave(t *testing.T) {
	// Protocolos em ordem INVERTIDA em relacao as notas.
	ret := &RetConsReciNFe{
		CStat: 104,
		ProtNFe: []ProtDFe{
			{ChNFe: chaveLote2, NProt: "segunda", CStat: 100},
			{ChNFe: chaveLote1, NProt: "primeira", CStat: 100},
		},
	}
	primeira := notaParaEnvio(chaveLote1, 35, 55, 4.00, pcn.TaHomologacao)
	segunda := notaParaEnvio(chaveLote2, 35, 55, 4.00, pcn.TaHomologacao)

	AnexarProtocolos(ret, []*NotaFiscal{primeira, segunda})

	if primeira.NFe.ProcNFe.NProt != "primeira" {
		t.Errorf("primeira nota recebeu %q", primeira.NFe.ProcNFe.NProt)
	}
	if segunda.NFe.ProcNFe.NProt != "segunda" {
		t.Errorf("segunda nota recebeu %q", segunda.NFe.ProcNFe.NProt)
	}
}

// TestResultadoEnvio_AssincronoDevolveOLotePronto trava a conveniencia que
// fecha o fluxo: os quatro dados da consulta seguinte saem prontos do
// envio, sem o consumidor ter de guardar versao, modelo e cUF por fora.
func TestResultadoEnvio_AssincronoDevolveOLotePronto(t *testing.T) {
	resposta := "<retEnviNFe versao=\"3.10\" xmlns=\"http://www.portalfiscal.inf.br/nfe\">" +
		"<tpAmb>2</tpAmb><cStat>103</cStat><xMotivo>ok</xMotivo><cUF>31</cUF>" +
		"<infRec><nRec>311000000000007</nRec><tMed>1</tMed></infRec></retEnviNFe>"

	componente := NovoComponente(configDeTransmissao()) // SP, 4.00
	nota := notaParaEnvio(chaveLote1, 31, 55, 3.10, pcn.TaHomologacao)

	res, err := componente.tratarRespostaDoEnvio(resposta, []*NotaFiscal{nota}, false,
		DestinoEnvio{Versao: Ve310, Modelo: MoNFe, UF: "MG"})
	if err != nil {
		t.Fatal(err)
	}

	esperado := ReciboDoLote{Recibo: "311000000000007", Modelo: MoNFe, Versao: Ve310, CUF: 31}
	if res.Lote != esperado {
		t.Errorf("Lote = %+v, esperado %+v", res.Lote, esperado)
	}
}
