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
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do ROTEAMENTO dos eventos.
//
// E a condicional mais rica do porte da NF-e: eventos do MESMO documento
// vao para SEFAZ diferentes conforme o tipo, e errar o endereco nao da erro
// de rede -- da rejeicao, com uma mensagem que nao menciona o endereco.
//
// Os testes afirmam sobre a UF ESCOLHIDA, e nao sobre a URL: a URL vem da
// tabela, que ja tem auditoria propria, e afirmar sobre ela amarraria o
// teste a um endereco que a SEFAZ muda.

// chaveDeMinas e uma chave de NF-e emitida em MG (cUF 31, modelo 55).
const chaveDeMinas = "31260199999999000191550010000004561123456791"

// chaveNFCeDeMinas e a mesma coisa no modelo 65.
const chaveNFCeDeMinas = "31260199999999000191650010000004561123456798"

// chaveNFCeDeSaoPaulo e uma NFC-e emitida em SP (cUF 35) -- a unica UF com
// servico de EPEC para o modelo 65.
const chaveNFCeDeSaoPaulo = "35260199999999000191650010000004561123456794"

func ufEscolhidaDoEvento(t *testing.T, cfg Configuracoes, inf InfEvento) string {
	t.Helper()
	destino, err := NovoComponente(cfg).enderecoDoEvento(inf)
	if err != nil {
		t.Fatalf("enderecoDoEvento: %v", err)
	}
	return destino.UF
}

// TestEnderecoDoEvento_EventoLocalVaiParaAUFDaChave trava a lista fechada
// de :3109.
//
// Sao SETE tipos que a SEFAZ do estado trata. Todo o resto vai para o
// Ambiente Nacional, e e a maioria -- inclusive a manifestacao do
// destinatario, que e o evento mais comum depois do cancelamento.
func TestEnderecoDoEvento_EventoLocalVaiParaAUFDaChave(t *testing.T) {
	cfg := configDeTransmissao() // SP; a chave e de MG, de proposito

	locais := []pcn.TipoEvento{
		pcn.TeCCe, pcn.TeCancelamento, pcn.TeCancSubst,
		pcn.TePedProrrog1, pcn.TePedProrrog2,
		pcn.TeCanPedProrrog1, pcn.TeCanPedProrrog2,
	}
	for _, tipo := range locais {
		inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: tipo}
		if got := ufEscolhidaDoEvento(t, cfg, inf); got != "MG" {
			t.Errorf("%s: UF = %q, esperada MG (a da CHAVE, nao a do componente)",
				tipo, got)
		}
	}
}

// TestEnderecoDoEvento_ManifestacaoVaiParaOAmbienteNacional trava o desvio
// que atinge mais eventos.
func TestEnderecoDoEvento_ManifestacaoVaiParaOAmbienteNacional(t *testing.T) {
	cfg := configDeTransmissao()

	manifestacoes := []pcn.TipoEvento{
		pcn.TeManifDestConfirmacao, pcn.TeManifDestCiencia,
		pcn.TeManifDestDesconhecimento, pcn.TeManifDestOperNaoRealizada,
	}
	for _, tipo := range manifestacoes {
		inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: tipo}
		if got := ufEscolhidaDoEvento(t, cfg, inf); got != "AN" {
			t.Errorf("%s: UF = %q, esperada AN", tipo, got)
		}
	}

	// E o comprovante de entrega, que tambem nao e local.
	inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: pcn.TeComprEntregaNFe}
	if got := ufEscolhidaDoEvento(t, cfg, inf); got != "AN" {
		t.Errorf("comprovante de entrega: UF = %q, esperada AN", got)
	}
}

// TestEnderecoDoEvento_InsucessoEntregaVoltaParaOSVRS trava o desvio de
// :3114, que acontece DEPOIS do desvio para o AN.
//
// O insucesso na entrega nao e evento local, entao a primeira regra o manda
// para o AN; a segunda o traz de volta para o SVRS. A ordem das duas no
// Delphi e o que decide -- invertida, o evento ficaria no AN.
func TestEnderecoDoEvento_InsucessoEntregaVoltaParaOSVRS(t *testing.T) {
	cfg := configDeTransmissao()

	for _, tipo := range []pcn.TipoEvento{
		pcn.TeInsucessoEntregaNFe, pcn.TeCancInsucessoEntregaNFe,
	} {
		inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: tipo}
		got := ufEscolhidaDoEvento(t, cfg, inf)
		if got != "SVRS" {
			t.Errorf("%s: UF = %q, esperada SVRS", tipo, got)
		}
		if got == "AN" {
			t.Errorf("%s ficou no Ambiente Nacional: o desvio para o SVRS"+
				" tem de vir DEPOIS", tipo)
		}
	}
}

// TestEnderecoDoEvento_ConciliacaoFinanceiraOlhaOModeloDaCONFIGURACAO trava
// o unico desvio que nao usa o modelo da chave (:3124).
//
// Os dois casos usam a MESMA chave -- de NF-e, de MG. O que muda e o
// ModeloDF do componente. Com o modelo da chave, os dois dariam SVRS.
func TestEnderecoDoEvento_ConciliacaoFinanceiraOlhaOModeloDaCONFIGURACAO(t *testing.T) {
	for _, tipo := range []pcn.TipoEvento{
		pcn.TeConcFinanceira, pcn.TeCancConcFinanceira,
	} {
		t.Run(tipo.String(), func(t *testing.T) {
			inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: tipo}

			comNFe := configDeTransmissao() // ModeloDF = MoNFe
			if got := ufEscolhidaDoEvento(t, comNFe, inf); got != "SVRS" {
				t.Errorf("componente NF-e: UF = %q, esperada SVRS", got)
			}

			comNFCe := configDeTransmissao()
			comNFCe.ModeloDF = MoNFCe
			// A chave continua sendo a de NF-e: o desvio olha a
			// CONFIGURACAO.
			if got := ufEscolhidaDoEvento(t, comNFCe, inf); got != "MG" {
				t.Errorf("componente NFC-e: UF = %q, esperada MG (a da chave)", got)
			}
		})
	}
}

// TestEnderecoDoEvento_EventosDaReformaVaoParaOSVRS trava os quinze tipos
// de :3133.
//
// A lista NAO e "o que a Reforma criou": o teAtualizacaoDataPrevisaoEntrega
// esta nela e nao e tributario. Derivar o grupo de um criterio semantico
// deixaria esse de fora.
func TestEnderecoDoEvento_EventosDaReformaVaoParaOSVRS(t *testing.T) {
	cfg := configDeTransmissao()

	daReforma := []pcn.TipoEvento{
		pcn.TeCancGenerico, pcn.TePagIntegLibCredPresAdq, pcn.TeImporALCZFM,
		pcn.TePerecPerdaRouboFurtoTranspContratFornec,
		pcn.TeFornecNaoRealizPagAntec, pcn.TeSolicApropCredPres,
		pcn.TeDestItemConsPessoal,
		pcn.TePerecPerdaRouboFurtoTranspContratAqu,
		pcn.TeAceiteDebitoApuracaoNotaCredito, pcn.TeImobilizacaoItem,
		pcn.TeSolicApropCredCombustivel, pcn.TeSolicApropCredBensServicos,
		pcn.TeManifPedTransfCredIBSSucessao,
		pcn.TeManifPedTransfCredCBSSucessao,
		pcn.TeAtualizacaoDataPrevisaoEntrega,
	}
	if len(daReforma) != 15 {
		t.Fatalf("a lista do ACBr tem 15 tipos, o teste cobre %d", len(daReforma))
	}
	for _, tipo := range daReforma {
		inf := InfEvento{ChNFe: chaveDeMinas, TpEvento: tipo}
		if got := ufEscolhidaDoEvento(t, cfg, inf); got != "SVRS" {
			t.Errorf("%s: UF = %q, esperada SVRS", tipo, got)
		}
	}
}

// TestEnderecoDoEvento_EPECdeNFCeSoTemServicoProprioEmSP trava a condicao
// TRIPLA de :3101.
//
// Tipo EPEC, componente em SP E modelo NFC-e. Afrouxar qualquer um manda o
// evento para o RecepcaoEvento comum -- ou, pior, procura um EventoEPEC que
// so Sao Paulo publica.
func TestEnderecoDoEvento_EPECdeNFCeSoTemServicoProprioEmSP(t *testing.T) {
	casos := []struct {
		cUF       int
		modelo    ModeloDF
		tipo      pcn.TipoEvento
		servico   Servico
		descricao string
	}{
		{35, MoNFCe, pcn.TeEPECNFe, ServicoEventoEPEC, "SP + NFC-e + EPEC"},
		{35, MoNFe, pcn.TeEPECNFe, ServicoRecepcaoEvento, "SP + NF-e: servico comum"},
		// MG + NFC-e fica de fora: fora de SP o EPEC de NFC-e vai para o
		// RecepcaoEvento do Ambiente Nacional, que NAO publica endereco
		// para o modelo 65 -- o ACBr levanta excecao no mesmo ponto. O
		// caso esta coberto em TestEnderecoDoEvento_EPECdeNFCeForaDeSPNaoTemEndereco.
		{35, MoNFCe, pcn.TeCCe, ServicoRecepcaoEvento, "SP + NFC-e mas outro evento"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			cfg := configDeTransmissao()
			cfg.CodigoUF, cfg.ModeloDF = c.cUF, c.modelo
			if c.cUF == 31 {
				cfg.UF = "MG"
			}

			// A UF da BUSCA sai da CHAVE, nao da configuracao: para
			// exercitar o EPEC de SP a chave tambem tem de ser de SP.
			chave := chaveDeMinas
			switch {
			case c.cUF == 35 && c.modelo == MoNFCe:
				chave = chaveNFCeDeSaoPaulo
			case c.modelo == MoNFCe:
				chave = chaveNFCeDeMinas
			}

			destino, err := NovoComponente(cfg).enderecoDoEvento(
				InfEvento{ChNFe: chave, TpEvento: c.tipo})
			if err != nil {
				t.Fatal(err)
			}
			if destino.Servico != c.servico {
				t.Errorf("servico = %q, esperado %q", destino.Servico, c.servico)
			}
		})
	}
}

// TestEnderecoDoEvento_ContingenciaSVCSoAlcancaOEventoLOCAL trava a ordem
// das atribuicoes.
//
// O SVC e escolhido PRIMEIRO (:3096) e os desvios posteriores o
// sobrescrevem. O resultado e que a contingencia muda o destino de um
// cancelamento, mas nao o de uma manifestacao -- que continua indo ao
// Ambiente Nacional.
func TestEnderecoDoEvento_ContingenciaSVCSoAlcancaOEventoLOCAL(t *testing.T) {
	cfg := configDeTransmissao()
	cfg.FormaEmissao = pcn.TeSVCRS

	local := InfEvento{ChNFe: chaveDeMinas, TpEvento: pcn.TeCancelamento}
	if got := ufEscolhidaDoEvento(t, cfg, local); got != "SVC-RS" {
		t.Errorf("cancelamento em contingencia: UF = %q, esperada SVC-RS", got)
	}

	nacional := InfEvento{ChNFe: chaveDeMinas, TpEvento: pcn.TeManifDestCiencia}
	if got := ufEscolhidaDoEvento(t, cfg, nacional); got != "AN" {
		t.Errorf("manifestacao em contingencia: UF = %q; o desvio para o AN"+
			" vem DEPOIS do SVC e vence", got)
	}
}

// TestEnderecoDoEvento_ModeloSaiDaCHAVE trava :3089.
//
// O componente esta configurado para NF-e; a chave e de NFC-e. O endereco
// tem de ser procurado na tabela da NFC-e.
func TestEnderecoDoEvento_ModeloSaiDaCHAVE(t *testing.T) {
	cfg := configDeTransmissao() // MoNFe

	destino, err := NovoComponente(cfg).enderecoDoEvento(
		InfEvento{ChNFe: chaveNFCeDeMinas, TpEvento: pcn.TeCancelamento})
	if err != nil {
		t.Fatal(err)
	}
	if destino.Modelo != MoNFCe {
		t.Errorf("modelo = %v, esperado MoNFCe", destino.Modelo)
	}

	comoNFe, err := URLServico(MoNFe, "MG", cfg.Ambiente,
		ServicoRecepcaoEvento, cfg.VersaoDF)
	if err != nil {
		t.Fatal(err)
	}
	if destino.URL == comoNFe {
		t.Errorf("o evento de NFC-e foi para o endereco da NF-e: %q", destino.URL)
	}
}

// TestServicoEActionDoEvento trava o par servico/action por versao.
func TestServicoEActionDoEvento(t *testing.T) {
	casos := []struct {
		versao  VersaoDF
		servico string
		acao    string
	}{
		{Ve400, "NFeRecepcaoEvento4", "NFeRecepcaoEvento4/nfeRecepcaoEvento"},
		{Ve310, "RecepcaoEvento", "RecepcaoEvento"},
		{Ve200, "RecepcaoEvento", "RecepcaoEvento"},
	}
	for _, c := range casos {
		cfg := configDeTransmissao()
		cfg.VersaoDF = c.versao
		servico, acao := NovoComponente(cfg).servicoEActionDoEvento()
		if servico != enderecoWsd+c.servico || acao != enderecoWsd+c.acao {
			t.Errorf("versao %v: %q / %q, esperado %q / %q", c.versao,
				servico, acao, enderecoWsd+c.servico, enderecoWsd+c.acao)
		}
	}
}

// TestEventoVaiNaVersao100MesmoNumComponente400 trava a versao do leiaute
// do EVENTO, que nao acompanha a da NF-e.
//
// A SEFAZ publica um unico envEvento_v1.00.xsd. Emitir "4.00" porque o
// componente esta na 4.00 e rejeicao de schema -- e o erro natural, porque
// todos os outros servicos acompanham a configuracao.
func TestEventoVaiNaVersao100MesmoNumComponente400(t *testing.T) {
	cfg := configDeTransmissao() // Ve400
	if got := VersaoDaMensagem(ServicoRecepcaoEvento, cfg.VersaoDF); got != "1.00" {
		t.Errorf("versao do envEvento = %q, esperada 1.00", got)
	}
	if got := VersaoDaMensagem(ServicoEventoEPEC, cfg.VersaoDF); got != "1.00" {
		t.Errorf("versao do EventoEPEC = %q, esperada 1.00", got)
	}
	// E o contraste: a autorizacao acompanha.
	if got := VersaoDaMensagem(ServicoAutorizacao, cfg.VersaoDF); got != "4.00" {
		t.Errorf("versao da autorizacao = %q, esperada 4.00", got)
	}
}

// TestEnderecoDoEvento_EPECdeNFCeForaDeSPNaoTemEndereco trava a outra
// metade da excecao de Sao Paulo.
//
// Fora de SP o EPEC de NFC-e cai no RecepcaoEvento comum, que por nao ser
// evento local vai para o Ambiente Nacional -- e o `.ini` nao publica
// secao AN para o modelo 65. O resultado e ErrSemURL, e e CORRETO: as
// demais UFs nao oferecem EPEC de NFC-e porque permitem contingencia
// off-line (comentario de :3101).
//
// O teste existe para que esse erro seja uma decisao registrada, e nao uma
// surpresa que alguem "conserte" acrescentando uma entrada inventada na
// tabela.
func TestEnderecoDoEvento_EPECdeNFCeForaDeSPNaoTemEndereco(t *testing.T) {
	cfg := configDeTransmissao()
	cfg.UF, cfg.CodigoUF = "MG", 31
	cfg.ModeloDF = MoNFCe

	_, err := NovoComponente(cfg).enderecoDoEvento(
		InfEvento{ChNFe: chaveNFCeDeMinas, TpEvento: pcn.TeEPECNFe})
	if !errors.Is(err, ErrSemURL) {
		t.Errorf("err = %v, esperado ErrSemURL", err)
	}

	// Em SP o mesmo evento tem endereco.
	cfg.UF, cfg.CodigoUF = "SP", 35
	destino, err := NovoComponente(cfg).enderecoDoEvento(
		InfEvento{ChNFe: chaveNFCeDeSaoPaulo, TpEvento: pcn.TeEPECNFe})
	if err != nil {
		t.Fatalf("em SP o EPEC de NFC-e existe: %v", err)
	}
	if destino.Servico != ServicoEventoEPEC {
		t.Errorf("servico = %q, esperado EventoEPEC", destino.Servico)
	}
}
