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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes da transmissao da Distribuicao DF-e.

// TestUrlDaDistribuicao_SempreNoAmbienteNacional trava :4001.
//
// A primeira linha do DefinirURL e `UF := 'AN'`, e a configuracao nem e
// consultada. E o unico servico da NF-e sem variacao por UF -- e o unico
// em que a contingencia SVC nao muda nada.
func TestUrlDaDistribuicao_SempreNoAmbienteNacional(t *testing.T) {
	naAN, err := URLServico(MoNFe, "AN", pcn.TaHomologacao,
		ServicoDistribuicaoDFe, Ve400)
	if err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		ajustar   func(*Configuracoes)
		descricao string
	}{
		{func(c *Configuracoes) {}, "configuracao normal"},
		{func(c *Configuracoes) { c.UF, c.CodigoUF = "MG", 31 }, "outra UF"},
		{func(c *Configuracoes) { c.FormaEmissao = pcn.TeSVCAN }, "contingencia SVC-AN"},
		{func(c *Configuracoes) { c.FormaEmissao = pcn.TeSVCRS }, "contingencia SVC-RS"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			cfg := configDeTransmissao()
			c.ajustar(&cfg)

			url, err := NovoComponente(cfg).urlDaDistribuicao("")
			if err != nil {
				t.Fatal(err)
			}
			if url != naAN {
				t.Errorf("url = %q, esperada a do AN (%q)", url, naAN)
			}
		})
	}
}

// TestUrlDaDistribuicao_ModeloSaiDaChaveQuandoHa trava :4011.
//
// A UF e fixa, mas o modelo nao: sai da CHAVE quando ela e informada, e cai
// para o configurado na varredura por NSU, que nao tem chave.
func TestUrlDaDistribuicao_ModeloSaiDaChaveQuandoHa(t *testing.T) {
	cfg := configDeTransmissao() // MoNFe

	// Sem chave: vale o modelo da configuracao, e o AN publica o servico.
	if _, err := NovoComponente(cfg).urlDaDistribuicao(""); err != nil {
		t.Fatalf("varredura sem chave: %v", err)
	}

	// Com chave de NFC-e o modelo vira 65 -- e o `.ini` NAO publica
	// NFeDistribuicaoDFe sob NFCe/AN. O erro e correto: a distribuicao e
	// servico de NF-e.
	_, err := NovoComponente(cfg).urlDaDistribuicao(chaveNFCeDeMinas)
	if err == nil {
		t.Error("com chave de NFC-e o modelo vira 65, que o AN nao atende")
	}
}

// TestDistribuicaoNaURL_DuasCamadasNoEnvelope trava o empacotamento, que e
// unico neste servico.
//
// O elemento do Body e `nfeDistDFeInteresse` (:3983) e DENTRO dele vem um
// `nfeDadosMsg` que o proprio gerador emite (:4029). Nenhum outro servico
// da NF-e tem as duas camadas.
func TestDistribuicaoNaURL_DuasCamadasNoEnvelope(t *testing.T) {
	resposta := montarRetDist(t, 137, "000000000000000", "000000000000000")
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeDistDFeInteresseResult", resposta))

	c := componenteApontandoPara(sefaz, configDeTransmissao())

	ret, err := c.distribuicaoNaURL(context.Background(), sefaz.servidor.URL,
		PedidoDistribuicao{CNPJCPF: "99999999000191", UltNSU: "0"})
	if err != nil {
		t.Fatal(err)
	}
	if !ret.Atendida() {
		t.Errorf("cStat %d", ret.CStat)
	}

	corpo := sefaz.ultimoCorpo
	if !strings.Contains(corpo, "<nfeDistDFeInteresse ") {
		t.Errorf("o elemento do Body e o nfeDistDFeInteresse: %s", corpo)
	}
	if !strings.Contains(corpo, "<nfeDadosMsg><distDFeInt ") {
		t.Errorf("e dentro dele vem o nfeDadosMsg: %s", corpo)
	}
	conferirOrdem(t, corpo, "<nfeDistDFeInteresse ", "<nfeDadosMsg>", "<distDFeInt ")

	// Sem cabecalho SOAP, em qualquer versao (:3984).
	if strings.Contains(corpo, "nfeCabecMsg") || strings.Contains(corpo, "Header") {
		t.Errorf("a distribuicao nao manda cabecalho SOAP: %s", corpo)
	}

	if !strings.Contains(sefaz.ultimaAcao, "NFeDistribuicaoDFe/nfeDistDFeInteresse") {
		t.Errorf("SoapAction = %q", sefaz.ultimaAcao)
	}
}

// TestDistribuicaoNaURL_MensagemVaiNaVersao101 trava a versao fixa.
//
// O `.ini` so publica a chave `_1.01` e o schema mais alto e o
// distDFeInt_v1.01.xsd -- entao a mensagem sai em 1.01 mesmo num
// componente 4.00.
func TestDistribuicaoNaURL_MensagemVaiNaVersao101(t *testing.T) {
	if got := VersaoDaMensagem(ServicoDistribuicaoDFe, Ve400); got != "1.01" {
		t.Errorf("versao da mensagem = %q, esperada 1.01", got)
	}

	resposta := montarRetDist(t, 137, "0", "0")
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeDistDFeInteresseResult", resposta))
	c := componenteApontandoPara(sefaz, configDeTransmissao()) // Ve400

	if _, err := c.distribuicaoNaURL(context.Background(), sefaz.servidor.URL,
		PedidoDistribuicao{CNPJCPF: "99999999000191"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sefaz.ultimoCorpo, `<distDFeInt xmlns="`+namespaceNFe+`" versao="1.01">`) {
		t.Errorf("o distDFeInt deveria sair na versao 1.01: %s", sefaz.ultimoCorpo)
	}
}

// sefazComLotes responde uma sequencia de corpos, um por chamada, e guarda
// os pedidos recebidos.
type sefazComLotes struct {
	servidor  *httptest.Server
	respostas []string
	pedidos   []string
	chamada   int
}

func novaSefazComLotes(t *testing.T, respostas ...string) *sefazComLotes {
	t.Helper()
	s := &sefazComLotes{respostas: respostas}
	s.servidor = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		s.pedidos = append(s.pedidos, string(corpo))

		i := s.chamada
		if i >= len(s.respostas) {
			i = len(s.respostas) - 1
		}
		s.chamada++
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = io.WriteString(w, s.respostas[i])
	}))
	t.Cleanup(s.servidor.Close)
	return s
}

// TestVarrerDistribuicaoDFe_ContinuaAteOMaxNSU trava o criterio de parada
// da varredura.
//
// A RFB devolve no maximo cinquenta docZip por chamada. Parar no primeiro
// lote -- que e o que uma chamada unica faz -- deixa documentos para tras
// sem nenhum sinal de erro.
func TestVarrerDistribuicaoDFe_ContinuaAteOMaxNSU(t *testing.T) {
	primeiro := envelopeSOAP("nfeDistDFeInteresseResult",
		montarRetDist(t, 138, "000000000000002", "000000000000004",
			docZipDe(t, "1", "resNFe_v1.01", resumoNFe),
			docZipDe(t, "2", "resNFe_v1.01", resumoNFe)))
	segundo := envelopeSOAP("nfeDistDFeInteresseResult",
		montarRetDist(t, 138, "000000000000004", "000000000000004",
			docZipDe(t, "3", "resNFe_v1.01", resumoNFe),
			docZipDe(t, "4", "resNFe_v1.01", resumoNFe)))

	sefaz := novaSefazComLotes(t, primeiro, segundo)
	c := NovoComponente(configDeTransmissao())
	c.cliente = &dfe.ClienteSOAP{HTTPClient: sefaz.servidor.Client()}

	todos, ultimo := varrerNaURL(t, c, sefaz.servidor.URL,
		PedidoDistribuicao{CNPJCPF: "99999999000191"})

	if len(todos) != 4 {
		t.Errorf("documentos coletados = %d, esperado 4 (dois lotes)", len(todos))
	}
	if sefaz.chamada != 2 {
		t.Errorf("chamadas = %d, esperado 2", sefaz.chamada)
	}
	if !ultimo.VarreduraConcluida() {
		t.Error("a varredura deveria ter terminado no maxNSU")
	}

	// A segunda chamada tem de pedir a partir do ultNSU devolvido pela
	// primeira -- senao o laco repete o mesmo lote para sempre.
	if len(sefaz.pedidos) != 2 {
		t.Fatalf("pedidos = %d", len(sefaz.pedidos))
	}
	if !strings.Contains(sefaz.pedidos[0], "<ultNSU>000000000000000</ultNSU>") {
		t.Errorf("o primeiro pedido parte do zero: %s", sefaz.pedidos[0])
	}
	if !strings.Contains(sefaz.pedidos[1], "<ultNSU>000000000000002</ultNSU>") {
		t.Errorf("o segundo pedido parte do ultNSU devolvido: %s", sefaz.pedidos[1])
	}
}

// TestVarrerDistribuicaoDFe_ParaQuandoOUltNSUNaoAvanca trava a protecao
// contra laco infinito.
//
// Uma RFB que devolvesse sempre o mesmo ultNSU faria a varredura rodar
// para sempre. O laco para, e devolve o que coletou.
func TestVarrerDistribuicaoDFe_ParaQuandoOUltNSUNaoAvanca(t *testing.T) {
	travado := envelopeSOAP("nfeDistDFeInteresseResult",
		montarRetDist(t, 138, "000000000000007", "000000000000099",
			docZipDe(t, "7", "resNFe_v1.01", resumoNFe)))

	sefaz := novaSefazComLotes(t, travado)
	c := NovoComponente(configDeTransmissao())
	c.cliente = &dfe.ClienteSOAP{HTTPClient: sefaz.servidor.Client()}

	todos, _ := varrerNaURL(t, c, sefaz.servidor.URL,
		PedidoDistribuicao{CNPJCPF: "99999999000191", UltNSU: "7"})

	if sefaz.chamada != 1 {
		t.Errorf("chamadas = %d; o ultNSU nao avancou, entao para na"+
			" primeira", sefaz.chamada)
	}
	if len(todos) != 1 {
		t.Errorf("documentos = %d, esperado 1", len(todos))
	}
}

// TestVarrerDistribuicaoDFe_IgnoraNSUeChave trava a limpeza do pedido.
//
// NSU e Chave buscam UM documento; repetir a chamada com eles devolveria o
// mesmo item para sempre. A varredura os descarta.
func TestVarrerDistribuicaoDFe_IgnoraNSUeChave(t *testing.T) {
	resposta := envelopeSOAP("nfeDistDFeInteresseResult",
		montarRetDist(t, 137, "000000000000000", "000000000000000"))

	sefaz := novaSefazComLotes(t, resposta)
	c := NovoComponente(configDeTransmissao())
	c.cliente = &dfe.ClienteSOAP{HTTPClient: sefaz.servidor.Client()}

	varrerNaURL(t, c, sefaz.servidor.URL, PedidoDistribuicao{
		CNPJCPF: "99999999000191",
		NSU:     "42",
		Chave:   chaveDeMinas,
	})

	if len(sefaz.pedidos) != 1 {
		t.Fatalf("pedidos = %d", len(sefaz.pedidos))
	}
	if strings.Contains(sefaz.pedidos[0], "<consNSU>") ||
		strings.Contains(sefaz.pedidos[0], "<consChNFe>") {
		t.Errorf("a varredura descarta NSU e Chave: %s", sefaz.pedidos[0])
	}
	if !strings.Contains(sefaz.pedidos[0], "<distNSU>") {
		t.Errorf("e consulta por ultNSU: %s", sefaz.pedidos[0])
	}
}

// varrerNaURL exercita o LACO REAL da varredura apontando para um
// endereco fixo.
//
// A funcao `varrer` recebe quem faz a chamada, justamente para que o teste
// nao precise reescrever o criterio de parada. Uma copia do laco aqui nao
// provaria nada sobre o que roda em producao.
func varrerNaURL(t *testing.T, c *Componente, url string,
	p PedidoDistribuicao) ([]DocZip, *RetDistDFeInt) {

	t.Helper()
	todos, ultimo, err := varrer(p, 10, func(pedido PedidoDistribuicao) (*RetDistDFeInt, error) {
		return c.distribuicaoNaURL(context.Background(), url, pedido)
	})
	if err != nil {
		t.Fatal(err)
	}
	return todos, ultimo
}
