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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Cortes por versao de leiaute DENTRO dos grupos de ICMS, e os dois helpers
// de node corrigidos na etapa 4b.
//
// Separados de xml_writer_det_test.go porque sao a camada que o CLAUDE.md
// chama de defeito dominante do port -- a condicional em volta do campo, nao
// o campo. Nenhum deles e alcancado pelo round-trip nem pela auditoria
// mecanica.

// ---------------------------------------------------------------------------
// Cortes por versao do vICMSDeson -- NAO sao uniformes entre os grupos
// ---------------------------------------------------------------------------

// TestGerarDet_CorteDoVICMSDesonVariaPorGrupo trava os quatro cortes
// diferentes que governam a desoneracao do ICMS.
//
// Eles NAO coincidem, e e o tipo de coisa que um porte "arredonda" sem
// perceber: o CST 70 usa Versao >= 3 (XmlWriter.pas:2520), enquanto 30,
// 40/41/50 e 90 usam >= 3.10 (:2285, :2304, :2659) e a partilha do 20 usa
// >= 4 (:2237). vICMSDeson emitido numa versao que nao o tem e rejeicao de
// schema.
func TestGerarDet_CorteDoVICMSDesonVariaPorGrupo(t *testing.T) {
	casos := []struct {
		nome   string
		cst    pcn.CSTIcms
		versao float64
		espera bool
	}{
		// CST 70: entrou na 3.00, nao na 3.10.
		{"CST70 em 2.00", pcn.CST70, 2.00, false},
		{"CST70 em 3.00", pcn.CST70, 3.00, true},

		// CST 30: so na 3.10 -- em 3.00 ainda nao sai.
		{"CST30 em 3.00", pcn.CST30, 3.00, false},
		{"CST30 em 3.10", pcn.CST30, 3.10, true},

		// CST 40/41/50: o par vICMSDeson/motDesICMS substituiu o vICMS solto.
		{"CST40 em 3.00", pcn.CST40, 3.00, false},
		{"CST40 em 3.10", pcn.CST40, 3.10, true},

		// CST 20 x partilha do 20: MESMO bloco de codigo, cortes diferentes.
		{"CST20 em 3.10", pcn.CST20, 3.10, true},
		{"CSTPart20 em 3.10", pcn.CSTPart20, 3.10, false},
		{"CSTPart20 em 4.00", pcn.CSTPart20, 4.00, true},

		// CST 90: 3.10.
		{"CST90 em 3.00", pcn.CST90, 3.00, false},
		{"CST90 em 3.10", pcn.CST90, 3.10, true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			n := lerFixture(t, "nfe_400_completa.xml")
			n.InfNFe.Versao = c.versao
			n.Emit.CRT = CrtRegimeNormal
			n.Det = n.Det[:1]
			n.Det[0].Imposto.ICMS = ICMS{
				CST:        c.cst,
				VBC:        1000,
				VICMS:      180,
				VICMSDeson: 50,
			}

			// ESCOPADO ao grupo imposto de proposito: o ICMSTot tambem tem
			// um vICMSDeson, e ele sai a partir da 3.10. Procurar no
			// documento inteiro faria os casos "espera true" passarem mesmo
			// com o item mudo -- teste que nao consegue falhar.
			x := fragmentoDoGrupo(t, gerarDe(t, n), "imposto")
			if got := strings.Contains(x, "<vICMSDeson>"); got != c.espera {
				t.Errorf("vICMSDeson presente=%v, esperado %v", got, c.espera)
			}
		})
	}
}

// TestGerarDet_IndDeduzDesonSoNa400 trava o corte interno: onde a desoneracao
// sai, o indicador de deducao ainda depende de ser 4.00.
func TestGerarDet_IndDeduzDesonSoNa400(t *testing.T) {
	montar := func(versao float64) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.Versao = versao
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{
			CST:           pcn.CST30,
			VBCST:         1200,
			VICMSST:       216,
			VICMSDeson:    50,
			IndDeduzDeson: pcn.TieSim,
		}
		return fragmentoDoGrupo(t, gerarDe(t, n), "imposto")
	}

	if x := montar(3.10); strings.Contains(x, "<indDeduzDeson>") {
		t.Error("indDeduzDeson saiu numa 3.10; o campo entrou na 4.00")
	}
	if x := montar(4.00); !strings.Contains(x, "<indDeduzDeson>") {
		t.Error("indDeduzDeson ausente numa 4.00 com desoneracao")
	}
}

// ---------------------------------------------------------------------------
// Os blocos independentes do CST 90
// ---------------------------------------------------------------------------

// TestGerarDet_BlocosDoCST90TemRecortesDiferentes trava a assimetria mais
// escorregadia do grupo: quatro blocos do CST 90 exigem Versao >= 4 E
// CST = cst90 (XmlWriter.pas:2567, :2589, :2643, :2682), mas o bloco de FCP
// diferido exige apenas Versao >= 4 (:2605) -- ele sai TAMBEM na partilha.
//
// Colapsar os cinco sob a mesma condicao e o erro natural, e nenhum outro
// teste o detecta.
func TestGerarDet_BlocosDoCST90TemRecortesDiferentes(t *testing.T) {
	montar := func(cst pcn.CSTIcms) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{
			CST: cst,
			// abre o bloco 1 e o sub-bloco do diferimento
			VBC: 1000, VICMS: 180, VICMSOp: 200, PDif: 10, VICMSDif: 20,
			// bloco 2
			VBCFCP: 1000, PFCP: 2, VFCP: 20,
			// bloco 3 -- so >= 4, SEM recorte de CST
			PFCPDif: 1, VFCPDif: 10, VFCPEfet: 5,
			// bloco 5
			VBCST: 1200, VICMSST: 216, VBCFCPST: 1200, PFCPST: 2, VFCPST: 24,
			// bloco 8
			VICMSSTDeson: 30,
		}
		return fragmentoDoGrupo(t, gerarDe(t, n), "imposto")
	}

	soNoCST90 := []string{"vICMSOp", "vBCFCP", "vBCFCPST", "vICMSSTDeson"}

	x90 := montar(pcn.CST90)
	for _, tag := range append(append([]string{}, soNoCST90...), "pFCPDif") {
		if !strings.Contains(x90, "<"+tag+">") {
			t.Errorf("<%s> ausente no CST 90", tag)
		}
	}

	xPart := montar(pcn.CSTPart90)
	for _, tag := range soNoCST90 {
		if strings.Contains(xPart, "<"+tag+">") {
			t.Errorf("<%s> saiu na partilha do 90; o bloco exige CST = cst90", tag)
		}
	}
	if !strings.Contains(xPart, "<pFCPDif>") {
		t.Error("pFCPDif ausente na partilha do 90; esse bloco e o unico dos " +
			"cinco que NAO tem recorte de CST (XmlWriter.pas:2605)")
	}
}

// TestGerarDet_PMVASTDoGrupo90IgnoraOFormatoPorVersao trava a unica aliquota
// do writer que nao segue o FormatoValor4ou2 (XmlWriter.pas:2629).
func TestGerarDet_PMVASTDoGrupo90IgnoraOFormatoPorVersao(t *testing.T) {
	montar := func(cst pcn.CSTIcms) *NFe {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{
			CST: cst, VBC: 1000, VICMS: 180,
			VBCST: 1200, VICMSST: 216, PICMSST: 18, PMVAST: 40,
		}
		return n
	}

	x := fragmentoDoGrupo(t, gerarDe(t, montar(pcn.CST90)), "ICMS90")
	if !strings.Contains(x, "<pMVAST>40.00</pMVAST>") {
		t.Error("pMVAST do grupo 90 deveria sair com DUAS casas numa 4.00; " +
			"todos os outros percentuais usam quatro")
	}

	// Prova de que o contraste e real: no grupo 30 o mesmo campo vai a quatro.
	x = fragmentoDoGrupo(t, gerarDe(t, montar(pcn.CST30)), "ICMS30")
	if !strings.Contains(x, "<pMVAST>40.0000</pMVAST>") {
		t.Error("pMVAST do grupo 30 deveria sair com QUATRO casas numa 4.00")
	}
}

// ---------------------------------------------------------------------------
// Tags que o ACBr deixa sem sufixo
// ---------------------------------------------------------------------------

// TestGerarDet_CSTSemEntradaNoMapaGeraTagAninhada documenta e trava duas
// lacunas REPLICADAS do ACBr.
//
// O cstPart20 existe no enum e e tratado no case do writer
// (XmlWriter.pas:2201), mas foi esquecido no BuscaTag (:1916-1941); o
// csosnVazio idem no CSOSNToStrTagPos (ACBrDFe.Conversao.pas:1282). Nos dois
// casos o case do Delphi nao tem else, o Result fica vazio e sai uma tag sem
// sufixo -- que o proprio leitor nao conhece.
//
// Este teste existe para impedir que alguem "conserte" sem perceber que
// estaria divergindo do original.
func TestGerarDet_CSTSemEntradaNoMapaGeraTagAninhada(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Emit.CRT = CrtRegimeNormal
	n.Det = n.Det[:1]
	n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CSTPart20, VBC: 1000, VICMS: 180}

	if x := gerarDe(t, n); !strings.Contains(x, "<ICMS><ICMS>") {
		t.Error("cstPart20 deveria gerar <ICMS><ICMS>: o BuscaTag do ACBr nao " +
			"tem entrada para ele e devolve sufixo vazio")
	}

	n.Emit.CRT = CrtSimplesNacional
	n.Det[0].Imposto.ICMS = ICMS{CSOSN: pcn.CsosnVazio}
	if x := gerarDe(t, n); !strings.Contains(x, "<ICMSSN>") {
		t.Error("csosnVazio deveria gerar <ICMSSN> sem sufixo, pelo mesmo motivo")
	}
}

// ---------------------------------------------------------------------------
// Helpers de node corrigidos nesta etapa
// ---------------------------------------------------------------------------

// TestGerar_CNPJObrigatorioVazioSaiComQuatorzeZeros trava a condicao
// "obrigatorio OU tamanho > 0" do AddNodeCNPJCPF (ACBrXmlWriter.pas:179).
//
// Era um defeito do porte: eu conferia so o tamanho, entao um CNPJ
// obrigatorio e vazio saia como tag vazia. Nem o round-trip nem a auditoria
// mecanica alcancam -- o leitor devolve "00000000000000" como
// "00000000000000", e as duas geracoes concordam em qualquer das escolhas.
func TestGerar_CNPJObrigatorioVazioSaiComQuatorzeZeros(t *testing.T) {
	// O emit passa obrigatorio=true, o default do helper (XmlWriter.pas:723).
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Emit.CNPJCPF = ""

	if x := gerarDe(t, n); !strings.Contains(x, "<CNPJ>00000000000000</CNPJ>") {
		t.Error("CNPJ obrigatorio e vazio deveria sair com quatorze zeros")
	}

	// Opcional e vazio continua sumindo: a condicao e um OU, nao um E.
	n2 := lerFixture(t, "nfe_400_completa.xml")
	n2.Det = n2.Det[:1]
	n2.Det[0].Prod.DI = []DI{{
		NDi: "1", XLocDesemb: "SANTOS", UFDesemb: "SP", CExportador: "X",
	}}
	di := fragmentoDoGrupo(t, gerarDe(t, n2), "DI")
	if strings.Contains(di, "<CNPJ>") {
		t.Error("o CNPJ do terceiro na DI e OPCIONAL (XmlWriter.pas:1371) e " +
			"deveria sumir quando vazio")
	}
}

// TestGerar_CNPJCPFNaoMaiusculiza trava a escolha de helper de limpeza.
//
// O AddNodeCNPJCPF usa OnlyAlphaNum, que NAO muda a caixa
// (ACBrXmlWriter.pas:161 e ACBrUtil.Strings.pas:1197); so o AddNodeCNPJ usa a
// variante que maiusculiza (:203). A distincao passa a importar com o CNPJ
// alfanumerico de 2026.
func TestGerar_CNPJCPFNaoMaiusculiza(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Dest.CNPJCPF = "99999999ab0272"

	if x := gerarDe(t, n); !strings.Contains(x, "<CNPJ>99999999ab0272</CNPJ>") {
		t.Error("o CNPJ do destinatario foi maiusculizado; o AddNodeCNPJCPF do " +
			"ACBr limpa com OnlyAlphaNum, que preserva a caixa")
	}
}

// TestGerar_HoraMeiaNoiteContaComoVazia trava o criterio do tcHor
// (ACBrXmlWriter.pas:324), que olha SO O RELOGIO -- nao a data que acompanha
// o TDateTime.
//
// Era o outro defeito do porte: eu tratava so o time.Time zerado, entao uma
// hora de meia-noite com data preenchida saia como "00:00:00".
func TestGerar_HoraMeiaNoiteContaComoVazia(t *testing.T) {
	// hSaiEnt so existe abaixo da versao 3.
	n := lerFixture(t, "nfe_200_data_sem_hora.xml")
	n.Ide.DSaiEnt = data(2026, 10, 21)
	n.Ide.HSaiEnt = time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)

	if x := gerarDe(t, n); strings.Contains(x, "<hSaiEnt>") {
		t.Error("hSaiEnt de meia-noite saiu; o criterio de vazio do tcHor olha " +
			"so o relogio, e o campo e opcional")
	}

	n.Ide.HSaiEnt = time.Date(2026, 10, 21, 14, 30, 0, 0, time.UTC)
	if x := gerarDe(t, n); !strings.Contains(x, "<hSaiEnt>14:30:00</hSaiEnt>") {
		t.Error("hSaiEnt com hora real deveria sair")
	}
}
