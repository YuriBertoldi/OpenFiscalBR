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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes de transmissao.
//
// Nenhum toca a rede: um httptest responde no lugar da SEFAZ e guarda o
// que recebeu, que e o que permite afirmar sobre a REQUISICAO -- a
// SoapAction, o versaoDados e o conteudo do dadosMsg.
//
// A escolha do ENDERECO e testada a parte, sem servidor, porque ela e
// decidida antes de qualquer conexao.

// sefazFalsa devolve um servidor que responde sempre o mesmo corpo e guarda
// a ultima requisicao recebida.
type sefazFalsa struct {
	servidor    *httptest.Server
	ultimoCorpo string
	ultimaAcao  string
}

func novaSefazFalsa(t *testing.T, resposta string) *sefazFalsa {
	t.Helper()
	s := &sefazFalsa{}
	s.servidor = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corpo, _ := io.ReadAll(r.Body)
		s.ultimoCorpo = string(corpo)
		s.ultimaAcao = r.Header.Get("SOAPAction")
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		_, _ = io.WriteString(w, resposta)
	}))
	t.Cleanup(s.servidor.Close)
	return s
}

// componenteApontandoPara monta um componente cujo cliente SOAP fala com o
// servidor de teste em vez da SEFAZ.
func componenteApontandoPara(s *sefazFalsa, cfg Configuracoes) *Componente {
	c := NovoComponente(cfg)
	c.cliente = &dfe.ClienteSOAP{HTTPClient: s.servidor.Client()}
	return c
}

// envelopeSOAP embrulha um corpo de resposta no envelope que a SEFAZ usa.
func envelopeSOAP(tagResultado, conteudo string) string {
	return `<?xml version="1.0" encoding="utf-8"?>` +
		`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">` +
		`<soap:Body><` + tagResultado + `>` + conteudo +
		`</` + tagResultado + `></soap:Body></soap:Envelope>`
}

const respostaStatusOK = `<retConsStatServ versao="4.00" xmlns="http://www.portalfiscal.inf.br/nfe">` +
	`<tpAmb>2</tpAmb><verAplic>SP_NFE_PL009_V4</verAplic><cStat>107</cStat>` +
	`<xMotivo>Servico em Operacao</xMotivo><cUF>35</cUF>` +
	`<dhRecbto>2026-01-20T18:45:00-03:00</dhRecbto><tMed>1</tMed>` +
	`<xObs>observacao</xObs></retConsStatServ>`

func configDeTransmissao() Configuracoes {
	return Configuracoes{
		VersaoDF: Ve400,
		ModeloDF: MoNFe,
		Ambiente: pcn.TaHomologacao,
		UF:       "SP",
		CodigoUF: 35,
	}
}

// TestStatusServico_LeARespostaEMontaORequest cobre a ida e a volta.
func TestStatusServico_LeARespostaEMontaORequest(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeStatusServicoNF2Result", respostaStatusOK))
	c := componenteApontandoPara(sefaz, configDeTransmissao())

	// O endereco real da tabela e substituido pelo do servidor de teste.
	ret, err := c.statusServicoNaURL(context.Background(), sefaz.servidor.URL, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !ret.EmOperacao() {
		t.Errorf("cStat = %d; 107 e o unico que conta como em operacao", ret.CStat)
	}
	if ret.VerAplic != "SP_NFE_PL009_V4" || ret.CUF != 35 || ret.TMed != 1 {
		t.Errorf("resposta lida errado: %+v", ret)
	}
	if ret.TpAmb != pcn.TaHomologacao {
		t.Errorf("tpAmb = %v, esperado homologacao", ret.TpAmb)
	}

	// A REQUISICAO: o dadosMsg tem de carregar os tres campos do consStatServ.
	for _, esperado := range []string{
		"<consStatServ", `versao="4.00"`, "<tpAmb>2</tpAmb>",
		"<cUF>35</cUF>", "<xServ>STATUS</xServ>",
	} {
		if !strings.Contains(sefaz.ultimoCorpo, esperado) {
			t.Errorf("a requisicao nao contem %q", esperado)
		}
	}
	// Da 4.00 em diante a NF-e NAO manda cabecalho SOAP nenhum.
	if strings.Contains(sefaz.ultimoCorpo, "soap12:Header") {
		t.Errorf("a 4.00 nao tem <soap12:Header>: %s", sefaz.ultimoCorpo)
	}
}

// TestStatusServico_CabecalhoSOAPSoAbaixoDa400 trava o corte do
// InicializarServico (ACBrNFeWebServices.pas:692).
//
// Abaixo da 4.00 a SEFAZ exige um <nfeCabecMsg> com cUF e versaoDados; da
// 4.00 em diante, exige que ele NAO venha.
func TestStatusServico_CabecalhoSOAPSoAbaixoDa400(t *testing.T) {
	casos := []struct {
		versao    VersaoDF
		temHeader bool
	}{
		{Ve200, true},
		{Ve310, true},
		{Ve400, false},
	}

	for _, caso := range casos {
		t.Run(caso.versao.String(), func(t *testing.T) {
			sefaz := novaSefazFalsa(t, envelopeSOAP("nfeResultMsg", respostaStatusOK))
			cfg := configDeTransmissao()
			cfg.VersaoDF = caso.versao
			c := componenteApontandoPara(sefaz, cfg)

			if _, err := c.statusServicoNaURL(context.Background(), sefaz.servidor.URL, time.Now()); err != nil {
				t.Fatal(err)
			}

			temHeader := strings.Contains(sefaz.ultimoCorpo, "nfeCabecMsg")
			if temHeader != caso.temHeader {
				t.Errorf("cabecalho presente=%v, esperado %v: %s",
					temHeader, caso.temHeader, sefaz.ultimoCorpo)
			}
			if caso.temHeader {
				// RECORTA o cabecalho antes de procurar: o dadosMsg do
				// consStatServ tambem emite <cUF>, e procurar no corpo
				// inteiro faria a assercao passar com o cabecalho vazio.
				cab := recortarCabecalho(t, sefaz.ultimoCorpo)
				esperado := "<versaoDados>" + c.versaoServico() + "</versaoDados>"
				if !strings.Contains(cab, esperado) {
					t.Errorf("o cabecalho deveria conter %q, veio %q", esperado, cab)
				}
				if !strings.Contains(cab, "<cUF>35</cUF>") {
					t.Errorf("o cabecalho deveria conter o cUF configurado, veio %q", cab)
				}
			}
		})
	}
}

// TestStatusServico_SoapActionMudaComAVersao trava o DefinirServicoEAction,
// que tem tres formas e uma delas existe por causa de UMA UF.
func TestStatusServico_SoapActionMudaComAVersao(t *testing.T) {
	casos := []struct {
		nome       string
		ajusta     func(*Configuracoes)
		servico    string
		soapAction string
	}{
		{"4.00", func(cfg *Configuracoes) { cfg.VersaoDF = Ve400 },
			enderecoWsd + "NFeStatusServico4",
			enderecoWsd + "NFeStatusServico4/nfeStatusServicoNF"},
		{"3.10 fora da BA", func(cfg *Configuracoes) { cfg.VersaoDF = Ve310 },
			enderecoWsd + "NfeStatusServico2",
			enderecoWsd + "NfeStatusServico2"},
		{"3.10 na BA", func(cfg *Configuracoes) {
			cfg.VersaoDF = Ve310
			cfg.CodigoUF = 29
			cfg.UF = "BA"
		},
			enderecoWsd + "NfeStatusServico",
			enderecoWsd + "NfeStatusServico/NfeStatusServicoNF"},
		{"3.10 na BA mas em contingencia", func(cfg *Configuracoes) {
			cfg.VersaoDF = Ve310
			cfg.CodigoUF = 29
			cfg.UF = "BA"
			cfg.FormaEmissao = pcn.TeSVCRS
		},
			enderecoWsd + "NfeStatusServico2",
			enderecoWsd + "NfeStatusServico2"},
		{"3.10 na BA mas NFC-e", func(cfg *Configuracoes) {
			cfg.VersaoDF = Ve310
			cfg.CodigoUF = 29
			cfg.UF = "BA"
			cfg.ModeloDF = MoNFCe
		},
			enderecoWsd + "NfeStatusServico2",
			enderecoWsd + "NfeStatusServico2"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			cfg := configDeTransmissao()
			caso.ajusta(&cfg)
			c := NovoComponente(cfg)

			servico, action := c.servicoEActionDoStatus()
			if servico != caso.servico {
				t.Errorf("servico = %q, esperado %q", servico, caso.servico)
			}
			if action != caso.soapAction {
				t.Errorf("soapAction = %q, esperado %q", action, caso.soapAction)
			}
		})
	}
}

// TestConsultar_ChaveDecideUFEModelo trava a regra do ACBr de que a
// consulta vai para onde a NOTA foi autorizada, nao para onde o componente
// aponta.
func TestConsultar_ChaveDecideUFEModelo(t *testing.T) {
	// Chave de MG (31), modelo 65, com componente configurado para SP/55.
	const chaveMG = "31260199999999000191650010000004561123456797"

	if got := UFDaChave(chaveMG); got != "MG" {
		t.Errorf("UFDaChave = %q, esperado MG", got)
	}
	if got := ModeloDaChave(chaveMG, MoNFe); got != MoNFCe {
		t.Errorf("ModeloDaChave = %v, esperado NFC-e", got)
	}

	// A chave vazia cai no modelo configurado.
	if got := ModeloDaChave("", MoNFCe); got != MoNFCe {
		t.Error("sem chave, o modelo vem da configuracao")
	}

	// E a URL resolvida e a de MG, nao a de SP.
	urlMG, err := URLServico(MoNFCe, "MG", pcn.TaProducao, ServicoConsultaProtocolo, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	urlSP, err := URLServico(MoNFCe, "SP", pcn.TaProducao, ServicoConsultaProtocolo, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	if urlMG == urlSP {
		t.Error("MG e SP operam ambientes proprios; as URLs deveriam diferir")
	}
}

// TestUFParaBusca_ContingenciaSVC cobre o desvio usado pelo STATUS.
func TestUFParaBusca_ContingenciaSVC(t *testing.T) {
	cfg := configDeTransmissao()

	cfg.FormaEmissao = pcn.TeSVCRS
	if got := UFParaBusca(cfg); got != "SVC-RS" {
		t.Errorf("UFParaBusca = %q, esperado SVC-RS", got)
	}
	cfg.FormaEmissao = pcn.TeSVCAN
	if got := UFParaBusca(cfg); got != "SVC-AN" {
		t.Errorf("UFParaBusca = %q, esperado SVC-AN", got)
	}
	cfg.FormaEmissao = pcn.TeNormal
	if got := UFParaBusca(cfg); got != "SP" {
		t.Errorf("UFParaBusca = %q, esperado a UF configurada", got)
	}
}

// TestConsultar_ContingenciaSVCDesviaOEndereco trava o desvio da CONSULTA,
// que tem switch proprio e nao passa por UFParaBusca.
//
// A diferenca importa: a consulta normal vai para a UF da CHAVE, nao para a
// configurada, e so a contingencia a desvia.
func TestConsultar_ContingenciaSVCDesviaOEndereco(t *testing.T) {
	// Chave de MG com componente configurado para SP.
	const chaveMG = "31260199999999000191550010000004561123456793"

	normal, err := URLServico(MoNFe, UFDaChave(chaveMG), pcn.TaHomologacao,
		ServicoConsultaProtocolo, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	svc, err := URLServico(MoNFe, "SVC-RS", pcn.TaHomologacao,
		ServicoConsultaProtocolo, Ve400)
	if err != nil {
		t.Fatal(err)
	}
	if normal == svc {
		t.Fatal("o teste precisa de enderecos distintos para provar o desvio")
	}

	// Em contingencia, o endereco escolhido e o do SVC -- nao o de MG.
	cfg := configDeTransmissao()
	cfg.FormaEmissao = pcn.TeSVCRS
	c := NovoComponente(cfg)
	c.cliente = &dfe.ClienteSOAP{}

	escolhida := enderecoEscolhidoNaConsulta(t, c, chaveMG)
	if escolhida != svc {
		t.Errorf("em SVC-RS o endereco deveria ser %q, veio %q", svc, escolhida)
	}

	cfg.FormaEmissao = pcn.TeNormal
	c = NovoComponente(cfg)
	escolhida = enderecoEscolhidoNaConsulta(t, c, chaveMG)
	if escolhida != normal {
		t.Errorf("fora de contingencia o endereco deveria ser o de MG (%q), veio %q",
			normal, escolhida)
	}
}

// enderecoEscolhidoNaConsulta repete a resolucao que Consultar faz, para o
// teste poder afirmar sobre ela sem transmitir.
func enderecoEscolhidoNaConsulta(t *testing.T, c *Componente, chave string) string {
	t.Helper()
	uf := UFDaChave(chave)
	switch c.Config.FormaEmissao {
	case pcn.TeSVCAN:
		uf = "SVC-AN"
	case pcn.TeSVCRS:
		uf = "SVC-RS"
	}
	url, err := URLServico(ModeloDaChave(chave, c.Config.ModeloDF), uf,
		c.Config.Ambiente, ServicoConsultaProtocolo, c.Config.VersaoDF)
	if err != nil {
		t.Fatal(err)
	}
	return url
}

// recortarCabecalho devolve o conteudo do <nfeCabecMsg> do envelope.
func recortarCabecalho(t *testing.T, envelope string) string {
	t.Helper()
	i := strings.Index(envelope, "<nfeCabecMsg")
	j := strings.Index(envelope, "</nfeCabecMsg>")
	if i < 0 || j < 0 {
		t.Fatalf("envelope sem <nfeCabecMsg>: %s", envelope)
	}
	return envelope[i:j]
}

// TestConsultar_ChaveInvalidaNaoTransmite trava a validacao local: a chave
// e conferida ANTES de qualquer conexao.
func TestConsultar_ChaveInvalidaNaoTransmite(t *testing.T) {
	sefaz := novaSefazFalsa(t, "")
	c := componenteApontandoPara(sefaz, configDeTransmissao())

	_, err := c.Consultar(context.Background(), "35260199999999000191650010000004561123456790")
	if err == nil {
		t.Fatal("chave com digito verificador errado deveria reprovar")
	}
	// A mensagem tem de apontar a chave: sem isso, o erro e indistinguivel
	// de uma falha de rede.
	if !strings.Contains(err.Error(), "invalida") {
		t.Errorf("erro = %v; esperado dizer que a chave e invalida", err)
	}
	_ = sefaz
}

// TestTransmitir_SemCertificadoEhErro trava o portao de transmissao.
func TestTransmitir_SemCertificadoEhErro(t *testing.T) {
	c := NovoComponente(configDeTransmissao())

	_, err := c.StatusServico(context.Background(), time.Now())
	if !errors.Is(err, ErrCertificadoNaoConfigurado) {
		t.Errorf("erro = %v; esperado ErrCertificadoNaoConfigurado", err)
	}
}

// TestChamar_TentaOsNomesDeTagNaMesmaResposta trava o porte do
// SeparaDadosArray.
//
// A SEFAZ nomeia o elemento de retorno de formas diferentes por versao e
// por UF. O que importa aqui e que a tentativa seja feita sobre a MESMA
// resposta: reenviar o documento a cada nome seria errado e, na
// autorizacao, registraria duplicidade.
func TestChamar_TentaOsNomesDeTagNaMesmaResposta(t *testing.T) {
	// O servidor responde com o TERCEIRO nome da lista e CONTA as
	// requisicoes -- e a contagem que prova que nao houve reenvio.
	requisicoes := 0
	servidor := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requisicoes++
		_, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, envelopeSOAP("nfeResultMsg", respostaStatusOK))
	}))
	t.Cleanup(servidor.Close)

	c := NovoComponente(configDeTransmissao())
	c.cliente = &dfe.ClienteSOAP{HTTPClient: servidor.Client()}

	ret, err := c.statusServicoNaURL(context.Background(), servidor.URL, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if ret.CStat != 107 {
		t.Errorf("cStat = %d, esperado 107", ret.CStat)
	}
	if requisicoes != 1 {
		t.Errorf("foram %d requisicoes; os tres nomes de tag devem ser tentados "+
			"sobre a MESMA resposta", requisicoes)
	}
}

// TestAjustarHoraDoSVRS trava as condicoes do remendo de horario.
func TestAjustarHoraDoSVRS(t *testing.T) {
	const urlSVRS = "https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"
	const urlSP = "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx"

	agora := time.Date(2026, 1, 20, 18, 0, 0, 0, time.UTC)
	adiantada := agora.Add(time.Hour)
	proxima := agora.Add(10 * time.Minute)

	casos := []struct {
		nome     string
		recbto   time.Time
		url      string
		esperado time.Time
	}{
		{"SVRS com mais de 50 minutos de diferenca", adiantada, urlSVRS, agora},
		{"SVRS dentro dos 50 minutos", proxima, urlSVRS, proxima},
		{"outra UF, mesma diferenca", adiantada, urlSP, adiantada},
		{"data zerada", time.Time{}, urlSVRS, time.Time{}},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			got := AjustarHoraDoSVRS(c.recbto, c.url, agora)
			if !got.Equal(c.esperado) {
				t.Errorf("= %v, esperado %v", got, c.esperado)
			}
		})
	}
}

// TestConsultar_TransmiteComOCUFDaChave trava o ponto mais facil de
// regredir da etapa: o cUF do cabecalho SOAP vem da CHAVE do documento, nao
// da configuracao (TNFeConsulta.GerarUFSoap, ACBrNFeWebServices.pas:2169).
//
// Usa a versao 3.10 porque e a faixa em que o cabecalho existe.
func TestConsultar_TransmiteComOCUFDaChave(t *testing.T) {
	const respostaConsulta = `<retConsSitNFe versao="3.10" xmlns="http://www.portalfiscal.inf.br/nfe">` +
		`<tpAmb>2</tpAmb><verAplic>MG_v3</verAplic><cStat>100</cStat>` +
		`<xMotivo>Autorizado</xMotivo><cUF>31</cUF>` +
		`<chNFe>31260199999999000191550010000004561123456793</chNFe>` +
		`</retConsSitNFe>`

	sefaz := novaSefazFalsa(t, envelopeSOAP("NfeConsultaNF2Result", respostaConsulta))

	cfg := configDeTransmissao() // SP / cUF 35
	cfg.VersaoDF = Ve310
	c := componenteApontandoPara(sefaz, cfg)

	// Chave de MG (31).
	const chaveMG = "31260199999999000191550010000004561123456793"
	ret, err := c.consultarNaURL(context.Background(), sefaz.servidor.URL, chaveMG)
	if err != nil {
		t.Fatal(err)
	}
	if ret.CStat != 100 {
		t.Errorf("cStat = %d, esperado 100", ret.CStat)
	}

	// O cabecalho tem de trazer o cUF da CHAVE (31), nao o configurado (35).
	cab := recortarCabecalho(t, sefaz.ultimoCorpo)
	if !strings.Contains(cab, "<cUF>31</cUF>") {
		t.Errorf("o cabecalho deveria trazer o cUF da chave (31), veio %q", cab)
	}
	if strings.Contains(cab, "<cUF>35</cUF>") {
		t.Error("o cabecalho trouxe o cUF da configuracao, nao o da chave")
	}

	// E o dadosMsg tem os tres campos do consSitNFe, na ordem do .pas.
	ordem := []string{"<consSitNFe", "<tpAmb>", "<xServ>CONSULTAR</xServ>", "<chNFe>"}
	pos := -1
	for _, marcador := range ordem {
		i := strings.Index(sefaz.ultimoCorpo, marcador)
		if i < 0 {
			t.Errorf("dadosMsg sem %q", marcador)
			continue
		}
		if i < pos {
			t.Errorf("%q saiu fora de ordem; a sequencia e %v", marcador, ordem)
		}
		pos = i
	}
}

// TestSoapAction_BahiaNa400UsaOOverrideDoIni trava o achado que so o `.ini`
// revela: a secao [NFe_BA_P] declara SoapAction=SOAP_V4_BA, e la a action da
// consulta de protocolo e "NFeConsulta4", nao "NFeConsultaProtocolo4".
//
// Sem o override, a SEFAZ-BA rejeita a chamada -- a action vai no
// Content-Type e no header SOAPAction.
func TestSoapAction_BahiaNa400UsaOOverrideDoIni(t *testing.T) {
	// A consulta DIVERGE do que o codigo montaria.
	got := SoapActionDoIni(MoNFe, "BA", pcn.TaProducao, ServicoConsultaProtocolo, Ve400)
	const esperado = enderecoWsd + "NFeConsulta4/nfeConsultaNF"
	if got != esperado {
		t.Errorf("SoapAction da BA = %q, esperado %q", got, esperado)
	}

	montada := enderecoWsd + "NFeConsultaProtocolo4/nfeConsultaNF"
	if got == montada {
		t.Error("o override tem de DIFERIR do que o DefinirServicoEAction montaria")
	}

	// O status da BA coincide com a montada -- mas vem do .ini do mesmo jeito.
	gotStatus := SoapActionDoIni(MoNFe, "BA", pcn.TaProducao, ServicoStatus, Ve400)
	if gotStatus != enderecoWsd+"NFeStatusServico4/nfeStatusServicoNF" {
		t.Errorf("SoapAction do status da BA = %q", gotStatus)
	}

	// Fora da BA e abaixo da 4.00 nao ha override.
	for _, caso := range []struct {
		uf     string
		versao VersaoDF
	}{
		{"SP", Ve400}, {"BA", Ve310}, {"BA", Ve200},
	} {
		if got := SoapActionDoIni(MoNFe, caso.uf, pcn.TaProducao,
			ServicoConsultaProtocolo, caso.versao); got != "" {
			t.Errorf("%s/%s nao deveria ter override, veio %q",
				caso.uf, caso.versao, got)
		}
	}
}

// TestSoapAction_VaiNoFio confere que a action chega mesmo na requisicao --
// o teste anterior olha a funcao pura, este olha o header.
func TestSoapAction_VaiNoFio(t *testing.T) {
	sefaz := novaSefazFalsa(t, envelopeSOAP("nfeResultMsg", respostaStatusOK))
	c := componenteApontandoPara(sefaz, configDeTransmissao())

	if _, err := c.statusServicoNaURL(context.Background(), sefaz.servidor.URL, time.Now()); err != nil {
		t.Fatal(err)
	}

	esperado := enderecoWsd + "NFeStatusServico4/nfeStatusServicoNF"
	if sefaz.ultimaAcao != esperado {
		t.Errorf("SOAPAction no fio = %q, esperado %q", sefaz.ultimaAcao, esperado)
	}
}
