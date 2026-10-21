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

// Testes do gerador do det.
//
// O escopo aqui e o que as outras tres linhas de defesa NAO alcancam:
//
//   - o round-trip (gerar -> ler -> gerar) e cego para ORDEM, porque o leitor
//     acha tag por nome, e cego para MUTACAO, porque um campo ja alterado na
//     primeira geracao e relido igual na segunda;
//   - a auditoria mecanica (ferramentas/conferir-gerador-dfe.py) confere tipo
//     e obrigatoriedade de cada tag, mas nao as CONDICIONAIS em volta delas;
//   - o build e o vet nao veem nada disso.
//
// Sobra para estes testes exatamente a camada em que o CLAUDE.md diz morar o
// defeito dominante do port: a condicional.

// ---------------------------------------------------------------------------
// Ordem dos campos (XSD e xs:sequence)
// ---------------------------------------------------------------------------

// TestGerarDet_OrdemDosCamposSegueOXSD trava a sequencia dentro do item.
//
// O prod tem 30 campos em sequencia fixa e varios deles tem nome repetido em
// outros grupos do documento -- por isso cada subteste recorta o grupo antes
// de conferir.
func TestGerarDet_OrdemDosCamposSegueOXSD(t *testing.T) {
	x := gerarDe(t, lerFixture(t, "nfe_400_completa.xml"))
	det := fragmentoDoGrupo(t, x, "det")

	t.Run("prod", func(t *testing.T) {
		ordemDasTags(t, fragmentoDoGrupo(t, det, "prod"),
			"cProd", "cEAN", "xProd", "NCM", "CFOP", "uCom", "qCom",
			"vUnCom", "vProd", "cEANTrib", "uTrib", "qTrib", "vUnTrib",
			"indTot", "DI", "rastro")
	})

	t.Run("DI: tpViaTransp antes de cExportador", func(t *testing.T) {
		// O cExportador (I24) vem DEPOIS do bloco que a 3.00 acrescentou
		// (I23a a I23e), e o adi fecha o grupo.
		ordemDasTags(t, fragmentoDoGrupo(t, det, "DI"),
			"nDI", "dDI", "xLocDesemb", "UFDesemb", "dDesemb", "tpViaTransp",
			"tpIntermedio", "cExportador", "adi", "nAdicao", "nSeqAdic")
	})

	t.Run("imposto: ICMS antes de PIS e COFINS", func(t *testing.T) {
		ordemDasTags(t, fragmentoDoGrupo(t, det, "imposto"),
			"vTotTrib", "ICMS", "PIS", "COFINS")
	})

	t.Run("ICMS00", func(t *testing.T) {
		ordemDasTags(t, fragmentoDoGrupo(t, det, "ICMS00"),
			"orig", "CST", "modBC", "vBC", "pICMS", "vICMS")
	})

	// Os tres grupos de maior contagem de campos. Nao sao exercitados pelo
	// round-trip de forma util -- ele e cego para ordem -- e um deles
	// (veicProd) ainda troca duas tags de NOME por versao.
	t.Run("veicProd", func(t *testing.T) {
		ordemDasTags(t, fragmentoDoGrupo(t, x, "veicProd"),
			"tpOp", "chassi", "cCor", "xCor", "pot", "cilin", "pesoL", "pesoB",
			"nSerie", "tpComb", "nMotor", "CMT", "dist", "anoMod", "anoFab",
			"tpPint", "tpVeic", "espVeic", "VIN", "condVeic", "cMod",
			"cCorDENATRAN", "lota", "tpRest")
	})

	t.Run("ISSQN", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ISSQN = ISSQN{
			VBC: 100, VAliq: 5, VISSQN: 5, CMunFG: 3550308, CListServ: "01.01",
			VDeducao: 1, VOutro: 2, VDescIncond: 3, VDescCond: 4, VISSRet: 5,
			CServico: "S1", CMun: 3550308, CPais: 1058, NProcesso: "P1",
		}
		ordemDasTags(t, fragmentoDoGrupo(t, gerarDe(t, n), "ISSQN"),
			"vBC", "vAliq", "vISSQN", "cMunFG", "cListServ", "vDeducao",
			"vOutro", "vDescIncond", "vDescCond", "vISSRet", "indISS",
			"cServico", "cMun", "cPais", "nProcesso", "indIncentivo")
	})

	t.Run("ICMSUFDest", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMSUFDest = ICMSUFDest{
			VBCUFDest: 1000, VBCFCPUFDest: 1000, PFCPUFDest: 2,
			PICMSUFDest: 18, PICMSInter: 12, PICMSInterPart: 100,
			VFCPUFDest: 20, VICMSUFDest: 60, VICMSUFRemet: 0.01,
		}
		ordemDasTags(t, fragmentoDoGrupo(t, gerarDe(t, n), "ICMSUFDest"),
			"vBCUFDest", "vBCFCPUFDest", "pFCPUFDest", "pICMSUFDest",
			"pICMSInter", "pICMSInterPart", "vFCPUFDest", "vICMSUFDest",
			"vICMSUFRemet")
	})
}

// TestGerarDet_VeicProdTrocaDuasTagsDeNomeNa200 trava o par de renomeacoes
// do grupo de veiculo (XmlWriter.pas:1530 e :1546): CM3 virou cilin e CMKG
// virou CMT, carregando o MESMO campo da classe.
//
// Emitir o nome novo numa 1.10 -- ou o antigo numa 4.00 -- e documento
// rejeitado, e o round-trip nao ve: o leitor acha as duas tags por nome.
func TestGerarDet_VeicProdTrocaDuasTagsDeNomeNa200(t *testing.T) {
	montar := func(versao float64) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.Versao = versao
		return gerarDe(t, n)
	}

	x4 := montar(4.00)
	if !strings.Contains(x4, "<cilin>") || !strings.Contains(x4, "<CMT>") {
		t.Error("cilin/CMT ausentes numa 4.00")
	}
	if strings.Contains(x4, "<CM3>") || strings.Contains(x4, "<CMKG>") {
		t.Error("CM3/CMKG sairam numa 4.00; os nomes antigos acabaram na 2.00")
	}

	x1 := montar(1.10)
	if !strings.Contains(x1, "<CM3>") || !strings.Contains(x1, "<CMKG>") {
		t.Error("CM3/CMKG ausentes abaixo da 2.00")
	}
	if strings.Contains(x1, "<cilin>") || strings.Contains(x1, "<CMT>") {
		t.Error("cilin/CMT sairam abaixo da 2.00")
	}
}

// TestGerarDet_PBCOpVemAntesDeUFST trava a inversao mais facil de cometer do
// grupo de partilha.
//
// Os IDs do leiaute sugerem o contrario (N24 = UFST, N25 = pBCOp), mas o
// codigo do ACBr grava pBCOp primeiro (XmlWriter.pas:2167-2171) e o XSD
// confirma. Foi exatamente esse tipo de inversao que passou por todos os
// testes da etapa 4a no grupo guiaTransito.
func TestGerarDet_PBCOpVemAntesDeUFST(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	// O terceiro item da fixture e o ICMSPart.
	x := gerarDe(t, n)

	ordemDasTags(t, fragmentoDoGrupo(t, x, "ICMSPart"), "pBCOp", "UFST")
}

// ---------------------------------------------------------------------------
// A unica condicional POR DATA do gerador
// ---------------------------------------------------------------------------

// TestGerarDet_SemGTINDependeDaVersaoEDaData cobre as tres entradas da unica
// regra do gerador que decide por data (XmlWriter.pas:1163 e :1237).
//
// O round-trip nao alcanca: assim que "SEM GTIN" e escrito, ele e relido como
// conteudo normal e as geracoes seguintes concordam.
func TestGerarDet_SemGTINDependeDaVersaoEDaData(t *testing.T) {
	casos := []struct {
		nome    string
		versao  float64
		emissao time.Time
		espera  bool
	}{
		{"4.00 depois do corte", 4.00, data(2018, 12, 2), true},
		{"4.00 no dia do corte", 4.00, data(2018, 12, 1), false},
		{"4.00 antes do corte", 4.00, data(2018, 11, 30), false},
		{"3.10 depois do corte", 3.10, data(2019, 6, 1), false},
	}

	// A fixture ja traz "SEM GTIN" escrito em todos os itens, entao o caso
	// precisa ser montado aqui: o criterio da regra e o cEAN VAZIO.
	montar := func(versao float64, emissao time.Time, ean string) *NFe {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.Versao = versao
		n.Ide.DEmi = emissao
		for i := range n.Det {
			n.Det[i].Prod.CEAN = ean
			n.Det[i].Prod.CEANTrib = ean
		}
		return n
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			x := gerarDe(t, montar(c.versao, c.emissao, ""))
			tem := strings.Contains(x, "<cEAN>SEM GTIN</cEAN>")
			temTrib := strings.Contains(x, "<cEANTrib>SEM GTIN</cEANTrib>")

			if tem != c.espera || temTrib != c.espera {
				t.Errorf("cEAN=%v cEANTrib=%v, esperado %v para versao %.2f "+
					"emitida em %s", tem, temTrib, c.espera, c.versao,
					c.emissao.Format("2006-01-02"))
			}

			// Com cEAN preenchido o literal nunca entra.
			x = gerarDe(t, montar(c.versao, c.emissao, "7891234567895"))
			if strings.Contains(x, "SEM GTIN") {
				t.Error("SEM GTIN substituiu um cEAN preenchido")
			}
		})
	}
}

// data monta uma data sem hora, para os cortes por vigencia.
func data(ano int, mes time.Month, dia int) time.Time {
	return time.Date(ano, mes, dia, 0, 0, 0, 0, time.UTC)
}

// ---------------------------------------------------------------------------
// Mutacoes que o gerador faz no documento
// ---------------------------------------------------------------------------

// TestGerarDet_XProdDeHomologacaoSoNoPrimeiroItem cobre as tres condicoes
// que somadas decidem a substituicao da descricao (XmlWriter.pas:1180).
//
// Errar para MENOS faz a SEFAZ rejeitar a NFC-e de homologacao; errar para
// MAIS manda uma nota de producao com o aviso de "sem valor fiscal" impresso
// no DANFE. Os dois lados importam, entao o teste cobre os dois.
func TestGerarDet_XProdDeHomologacaoSoNoPrimeiroItem(t *testing.T) {
	casos := []struct {
		nome   string
		ajusta func(*NFe)
		espera bool
	}{
		{"NFC-e em homologacao", func(n *NFe) {
			n.Ide.Modelo = 65
			n.Ide.TpAmb = pcn.TaHomologacao
		}, true},
		{"NFC-e em producao", func(n *NFe) {
			n.Ide.Modelo = 65
			n.Ide.TpAmb = pcn.TaProducao
		}, false},
		{"NF-e em homologacao com DANFE normal", func(n *NFe) {
			n.Ide.Modelo = 55
			n.Ide.TpAmb = pcn.TaHomologacao
			n.Ide.TpImp = pcn.TiRetrato
		}, false},
		{"NF-e em homologacao com DANFE simplificado", func(n *NFe) {
			n.Ide.Modelo = 55
			n.Ide.TpAmb = pcn.TaHomologacao
			n.Ide.TpImp = pcn.TiSimplificadoTipo2
		}, true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			n := lerFixture(t, "nfe_400_completa.xml")
			c.ajusta(n)
			x := gerarDe(t, n)

			if got := strings.Contains(x, xProdHomologacao); got != c.espera {
				t.Errorf("aviso de homologacao presente=%v, esperado %v", got, c.espera)
			}

			// Mesmo quando entra, entra UMA vez so -- no item 1.
			if c.espera {
				if got := strings.Count(x, xProdHomologacao); got != 1 {
					t.Errorf("o aviso saiu %d vezes; so o primeiro item deve "+
						"receber a substituicao", got)
				}
				if n.Det[1].Prod.XProd == xProdHomologacao {
					t.Error("o segundo item teve a descricao substituida")
				}
			}
		})
	}
}

// TestGerarDet_CEnqVazioViraNoveNoveNove trava a mutacao do IPI
// (XmlWriter.pas:2990): a tag e obrigatoria e o ACBr preenche sozinho em vez
// de recusar o documento.
func TestGerarDet_CEnqVazioViraNoveNoveNove(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Det[0].Imposto.IPI.CST = Ipi00
	n.Det[0].Imposto.IPI.VBC = 100
	n.Det[0].Imposto.IPI.PIPI = 5
	n.Det[0].Imposto.IPI.VIPI = 5
	n.Det[0].Imposto.IPI.CEnq = ""

	if x := gerarDe(t, n); !strings.Contains(x, "<cEnq>999</cEnq>") {
		t.Error("cEnq vazio nao virou 999")
	}
	if n.Det[0].Imposto.IPI.CEnq != "999" {
		t.Errorf("a mutacao nao chegou ao documento: cEnq = %q",
			n.Det[0].Imposto.IPI.CEnq)
	}
}

// ---------------------------------------------------------------------------
// Portoes do IPI
// ---------------------------------------------------------------------------

// TestGerarDet_IPITemTresPortoes cobre as tres saidas antecipadas do
// GerarDetImpostoIPI (XmlWriter.pas:2960, :2968 e :2972).
//
// O terceiro e o mais importante: sem ele, o zero value de CstIpi -- que e
// Ipi00, nao vazio -- faria TODO item declarar IPI tributado com base zero.
func TestGerarDet_IPITemTresPortoes(t *testing.T) {
	t.Run("NFC-e nunca tem IPI", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.Modelo = 65
		n.Det[0].Imposto.IPI.CST = Ipi00
		n.Det[0].Imposto.IPI.VBC = 100
		n.Det[0].Imposto.IPI.VIPI = 5

		if strings.Contains(gerarDe(t, n), "<IPI>") {
			t.Error("grupo IPI saiu numa NFC-e")
		}
	})

	t.Run("IPI intocado nao sai", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		for i := range n.Det {
			n.Det[i].Imposto.IPI = IPI{}
		}

		if strings.Contains(gerarDe(t, n), "<IPI>") {
			t.Error("grupo IPI saiu com todos os campos no default; o zero " +
				"value de CstIpi e Ipi00, e sem o portao todo item declararia " +
				"IPI tributado")
		}
	})

	t.Run("nao tributado obedece a opcao", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Det[0].Imposto.IPI.CST = Ipi51
		n.Det[0].Imposto.IPI.CEnq = "999"

		// Default do ACBr: GerarTagIPIparaNaoTributado ligado, entao sai.
		if !strings.Contains(gerarDe(t, n), "<IPINT>") {
			t.Error("IPINT ausente com as opcoes padrao; o Create do ACBr liga " +
				"GerarTagIPIparaNaoTributado")
		}

		x, err := GerarXMLComOpcoes(n, Opcoes{OmitirTagIPIParaNaoTributado: true})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(x, "<IPINT>") {
			t.Error("IPINT saiu com OmitirTagIPIParaNaoTributado")
		}
	})

	t.Run("base por quantidade exclui base por valor", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		ipi := &n.Det[0].Imposto.IPI
		ipi.CST, ipi.CEnq = Ipi00, "999"
		ipi.VBC, ipi.PIPI = 100, 5
		ipi.QUnid, ipi.VUnid = 10, 2
		ipi.VIPI = 20

		x := fragmentoDoGrupo(t, gerarDe(t, n), "IPITrib")
		if !strings.Contains(x, "<qUnid>") || strings.Contains(x, "<vBC>") {
			t.Error("com qUnid/vUnid preenchidos o IPITrib deve sair por " +
				"quantidade, sem vBC nem pIPI")
		}
	})
}

// ---------------------------------------------------------------------------
// Escolha da tag de ICMS
// ---------------------------------------------------------------------------

// TestGerarDet_NomeDaTagDeICMSVemDoCST trava o BuscaTag (XmlWriter.pas:1916),
// que NAO e a identidade: CSTs distintos compartilham tag, e um deles muda de
// tag conforme a versao.
func TestGerarDet_NomeDaTagDeICMSVemDoCST(t *testing.T) {
	casos := []struct {
		cst    pcn.CSTIcms
		versao float64
		tag    string
	}{
		{pcn.CST00, 4.00, "ICMS00"},
		// Os tres CSTs de isencao caem na MESMA tag.
		{pcn.CST40, 4.00, "ICMS40"},
		{pcn.CST41, 4.00, "ICMS40"},
		{pcn.CST50, 4.00, "ICMS40"},
		// As duas partilhas tambem.
		{pcn.CSTPart10, 4.00, "ICMSPart"},
		{pcn.CSTPart90, 4.00, "ICMSPart"},
		// O repasse do 41 e sempre ST.
		{pcn.CSTRep41, 4.00, "ICMSST"},
		// O do 60 MUDOU DE TAG na 4.00.
		{pcn.CSTRep60, 3.10, "ICMS60"},
		{pcn.CSTRep60, 4.00, "ICMSST"},
	}

	for _, c := range casos {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.Versao = c.versao
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{CST: c.cst}

		x := gerarDe(t, n)
		if !strings.Contains(x, "<"+c.tag+">") {
			t.Errorf("CST %v na versao %.2f: esperado <%s>, nao encontrado",
				c.cst.CodigoEntrada(), c.versao, c.tag)
		}
	}
}

// TestGerarDet_CSOSNDecideATagDoSimples trava o CSOSNToStrTagPos: quatro
// CSOSN compartilham a tag ICMSSN102 e dois, a ICMSSN202.
func TestGerarDet_CSOSNDecideATagDoSimples(t *testing.T) {
	casos := []struct {
		csosn pcn.CSOSNIcms
		tag   string
	}{
		{pcn.Csosn101, "ICMSSN101"},
		{pcn.Csosn102, "ICMSSN102"},
		{pcn.Csosn103, "ICMSSN102"},
		{pcn.Csosn300, "ICMSSN102"},
		{pcn.Csosn400, "ICMSSN102"},
		{pcn.Csosn201, "ICMSSN201"},
		{pcn.Csosn202, "ICMSSN202"},
		{pcn.Csosn203, "ICMSSN202"},
		{pcn.Csosn500, "ICMSSN500"},
		{pcn.Csosn900, "ICMSSN900"},
	}

	for _, c := range casos {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Emit.CRT = CrtSimplesNacional
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{CSOSN: c.csosn}

		if x := gerarDe(t, n); !strings.Contains(x, "<"+c.tag+">") {
			t.Errorf("CSOSN %s: esperado <%s>, nao encontrado", c.csosn, c.tag)
		}
	}
}

// TestGerarDet_RegimeDoEmitenteEscolheEntreCSTeCSOSN trava o segundo nivel do
// GerarDetImpostoICMS: o CSOSN so e usado por optante do Simples e MEI.
func TestGerarDet_RegimeDoEmitenteEscolheEntreCSTeCSOSN(t *testing.T) {
	casos := []struct {
		crt     CRT
		simples bool
	}{
		{CrtRegimeNormal, false},
		{CrtSimplesExcessoReceita, false},
		{CrtSimplesNacional, true},
		{CrtMEI, true},
	}

	for _, c := range casos {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Emit.CRT = c.crt
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CST00, CSOSN: pcn.Csosn102}

		x := gerarDe(t, n)
		if got := strings.Contains(x, "<ICMSSN102>"); got != c.simples {
			t.Errorf("CRT %s: grupo do Simples presente=%v, esperado %v",
				c.crt, got, c.simples)
		}
	}
}

// TestGerarDet_MonofasiaIgnoraORegime trava o primeiro nivel: os quatro CSTs
// de monofasia desviam ANTES do case de regime, entao valem tambem para
// emitente do Simples (XmlWriter.pas:1974).
func TestGerarDet_MonofasiaIgnoraORegime(t *testing.T) {
	for _, crt := range []CRT{CrtRegimeNormal, CrtSimplesNacional, CrtMEI} {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Emit.CRT = crt
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CST02, QBCMono: 10, VICMSMono: 5}

		x := gerarDe(t, n)
		if !strings.Contains(x, "<ICMS02>") {
			t.Errorf("CRT %s: o CST 02 deveria gerar ICMS02 em qualquer regime", crt)
		}
		if strings.Contains(x, "<ICMSSN") {
			t.Errorf("CRT %s: o CST 02 nao deve cair no grupo do Simples", crt)
		}
	}
}

// ---------------------------------------------------------------------------
// Condicionais finas dentro dos grupos
// ---------------------------------------------------------------------------

// TestGerarDet_ICMS51ExigeGrupoZeradoEmTresUFs trava a lista de UFs do
// XmlWriter.pas:2337. Nas demais, campo zerado some.
func TestGerarDet_ICMS51ExigeGrupoZeradoEmTresUFs(t *testing.T) {
	casos := []struct {
		cUF    int
		nome   string
		espera bool
	}{
		{33, "RJ", true},
		{35, "SP", true},
		{41, "PR", true},
		{31, "MG", false},
		{43, "RS", false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			n := lerFixture(t, "nfe_400_completa.xml")
			n.Ide.CUF = c.cUF
			n.Emit.CRT = CrtRegimeNormal
			n.Det = n.Det[:1]
			n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CST51}

			grupo := fragmentoDoGrupo(t, gerarDe(t, n), "ICMS51")
			if got := strings.Contains(grupo, "<vBC>0.00</vBC>"); got != c.espera {
				t.Errorf("vBC zerado presente=%v, esperado %v", got, c.espera)
			}
		})
	}
}

// TestGerarDet_CESTCarregaIndEscalaECNPJFab trava o aninhamento do
// XmlWriter.pas:1199: sem CEST, nenhum dos tres sai.
//
// Importa porque IndEscala tem default NAO ZERO (ieNenhum, ordinal 2). Fora
// do if, todo item sem substituicao tributaria declararia indEscala.
func TestGerarDet_CESTCarregaIndEscalaECNPJFab(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	for i := range n.Det {
		n.Det[i].Prod.CEST = ""
		n.Det[i].Prod.IndEscala = IeRelevante
		n.Det[i].Prod.CNPJFab = "99999999000272"
	}

	x := gerarDe(t, n)
	for _, tag := range []string{"CEST", "indEscala", "CNPJFab"} {
		if strings.Contains(x, "<"+tag+">") {
			t.Errorf("<%s> saiu sem CEST; os tres sao filhos logicos dele", tag)
		}
	}

	n.Det[0].Prod.CEST = "0100100"
	x = gerarDe(t, n)
	prod := fragmentoDoGrupo(t, x, "prod")
	for _, tag := range []string{"CEST", "indEscala", "CNPJFab"} {
		if !strings.Contains(prod, "<"+tag+">") {
			t.Errorf("<%s> ausente com CEST preenchido", tag)
		}
	}
}

// TestGerarDet_MedTrocaDeCamposNa400 trava a exclusao mutua dos dois
// conjuntos (XmlWriter.pas:1593-1611). Emitir os dois juntos e documento
// rejeitado.
func TestGerarDet_MedTrocaDeCamposNa400(t *testing.T) {
	montar := func(versao float64) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.Versao = versao
		n.Det = n.Det[:1]
		n.Det[0].Prod.Med = []Med{{
			CProdANVISA: "1234567890123",
			NLote:       "LOTE1",
			QLote:       10,
			DFab:        data(2026, 1, 1),
			DVal:        data(2027, 1, 1),
			VPMC:        99.90,
		}}
		return fragmentoDoGrupo(t, gerarDe(t, n), "med")
	}

	x4 := montar(4.00)
	if !strings.Contains(x4, "<cProdANVISA>") {
		t.Error("cProdANVISA ausente na 4.00")
	}
	if strings.Contains(x4, "<nLote>") {
		t.Error("nLote saiu na 4.00; na 4.00 o lote virou o grupo rastro")
	}

	x3 := montar(3.10)
	if strings.Contains(x3, "<cProdANVISA>") {
		t.Error("cProdANVISA saiu numa 3.10; o campo entrou na 4.00")
	}
	if !strings.Contains(x3, "<nLote>") {
		t.Error("nLote ausente numa 3.10")
	}
}

// TestGerarDet_IIAbrePorCFOPDeImportacao trava a quarta condicao do
// XmlWriter.pas:3041, a unica que nao olha valor.
func TestGerarDet_IIAbrePorCFOPDeImportacao(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Det = n.Det[:1]
	n.Det[0].Imposto.II = II{}
	n.Det[0].Prod.CFOP = "5102"
	if strings.Contains(gerarDe(t, n), "<II>") {
		t.Error("grupo II saiu zerado com CFOP interno")
	}

	n.Det[0].Prod.CFOP = "3102"
	if !strings.Contains(gerarDe(t, n), "<II>") {
		t.Error("grupo II ausente com CFOP de importacao; o CFOP iniciado por " +
			"3 abre o grupo sozinho")
	}
}

// TestGerarDet_PISOpcionalSoNaNFCe trava a assimetria do XmlWriter.pas:3060:
// na NF-e o grupo sai sempre; na NFC-e, so com valor ou com CST que exija.
func TestGerarDet_PISOpcionalSoNaNFCe(t *testing.T) {
	montar := func(modelo int, cst pcn.CSTPis) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.Modelo = modelo
		n.Det = n.Det[:1]
		n.Det[0].Imposto.PIS = PIS{CST: cst}
		n.Det[0].Imposto.COFINS = COFINS{CST: pcn.Cof01}
		return gerarDe(t, n)
	}

	if !strings.Contains(montar(55, pcn.Pis01), "<PIS>") {
		t.Error("grupo PIS ausente numa NF-e sem valores; na NF-e ele sai sempre")
	}
	if strings.Contains(montar(65, pcn.Pis01), "<PIS>") {
		t.Error("grupo PIS saiu numa NFC-e sem valores e com CST que nao exige")
	}
	if !strings.Contains(montar(65, pcn.Pis04), "<PIS>") {
		t.Error("grupo PIS ausente numa NFC-e com CST 04; os nao tributados " +
			"exigem declaracao mesmo sem valor")
	}
}

// TestGerarDet_NTDeCreditoDebitoSuprimeOsImpostosAntigos cobre a NT 2025.002
// (XmlWriter.pas:1868-1875), que tem duas condicionais encaixadas e cinco
// codigos de excecao.
func TestGerarDet_NTDeCreditoDebitoSuprimeOsImpostosAntigos(t *testing.T) {
	montar := func(ajusta func(*NFe)) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		ajusta(n)
		return fragmentoDoGrupo(t, gerarDe(t, n), "imposto")
	}

	t.Run("finalidade normal detalha tudo", func(t *testing.T) {
		x := montar(func(n *NFe) { n.Ide.FinNFe = FnNormal })
		if !strings.Contains(x, "<ICMS>") {
			t.Error("ICMS ausente numa nota de finalidade normal")
		}
	})

	t.Run("credito puro nao detalha nada", func(t *testing.T) {
		x := montar(func(n *NFe) {
			n.Ide.FinNFe = FnCredito
			n.Ide.TpNFCredito = TcNenhum
		})
		for _, tag := range []string{"ICMS", "PIS", "COFINS"} {
			if strings.Contains(x, "<"+tag+">") {
				t.Errorf("<%s> saiu numa nota so de credito de IBS/CBS", tag)
			}
		}
	})

	t.Run("credito por retorno volta a detalhar", func(t *testing.T) {
		x := montar(func(n *NFe) {
			n.Ide.FinNFe = FnCredito
			n.Ide.TpNFCredito = TcRetorno
		})
		if !strings.Contains(x, "<ICMS>") {
			t.Error("ICMS ausente num retorno por recusa; tpNFCredito 03 e " +
				"excecao da NT 2025.002")
		}
	})

	t.Run("perda em estoque volta a detalhar", func(t *testing.T) {
		x := montar(func(n *NFe) {
			n.Ide.FinNFe = FnDebito
			n.Ide.TpNFDebito = TdPerdaEmEstoque
			n.Ide.TpNFCredito = TcNenhum
		})
		if !strings.Contains(x, "<ICMS>") {
			t.Error("ICMS ausente numa nota de perda em estoque; tpNFDebito 07 " +
				"e a unica excecao de debito da lista externa")
		}
	})

	t.Run("pagamento antecipado nao detalha nada", func(t *testing.T) {
		// O comentario do ACBr logo acima da condicional (XmlWriter.pas:1862)
		// lista tpNFDebito=06 entre as EXCECOES, mas o codigo nao o colocou na
		// lista externa -- so tdPerdaEmEstoque esta la. Resultado: a nota de
		// pagamento antecipado nao detalha nem PIS/COFINS.
		//
		// Replicado por fidelidade; registrado em divergencias-acbr.md.
		x := montar(func(n *NFe) {
			n.Ide.FinNFe = FnDebito
			n.Ide.TpNFDebito = TdPagamentoAntecipado
			n.Ide.TpNFCredito = TcNenhum
		})
		for _, tag := range []string{"ICMS", "PIS", "COFINS"} {
			if strings.Contains(x, "<"+tag+">") {
				t.Errorf("<%s> saiu num pagamento antecipado; a lista externa "+
					"do ACBr so excetua tdPerdaEmEstoque", tag)
			}
		}
	})

	t.Run("a condicional interna separa PIS de ICMS", func(t *testing.T) {
		// Unica combinacao que alcanca a condicional interna: a nota passa
		// pela lista externa por um tpNFCredito de excecao E carrega
		// tpNFDebito = pagamento antecipado.
		x := montar(func(n *NFe) {
			n.Ide.FinNFe = FnCredito
			n.Ide.TpNFCredito = TcRetorno
			n.Ide.TpNFDebito = TdPagamentoAntecipado
		})
		if strings.Contains(x, "<ICMS>") {
			t.Error("ICMS saiu; a condicional interna deixa de fora ICMS, IPI, " +
				"II e ISSQN quando tpNFDebito e pagamento antecipado")
		}
		if !strings.Contains(x, "<PIS>") {
			t.Error("PIS ausente; ele fica FORA da condicional interna")
		}
	})
}

// TestGerarDet_ISSQNEICMSSaoExcludentes trava a bifurcacao do
// XmlWriter.pas:1877: item de servico nao declara ICMS nem II.
func TestGerarDet_ISSQNEICMSSaoExcludentes(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Det = n.Det[:1]
	n.Det[0].Imposto.ISSQN = ISSQN{
		VBC: 100, VAliq: 5, VISSQN: 5, CMunFG: 3550308, CListServ: "01.01",
	}

	x := fragmentoDoGrupo(t, gerarDe(t, n), "imposto")
	if !strings.Contains(x, "<ISSQN>") {
		t.Error("grupo ISSQN ausente num item de servico")
	}
	if strings.Contains(x, "<ICMS>") {
		t.Error("ICMS saiu junto com ISSQN; os dois sao excludentes")
	}
}

// TestGerarDet_ICMSUFDestExigePercentualDePartilha trava o guard do
// XmlWriter.pas:1898.
func TestGerarDet_ICMSUFDestExigePercentualDePartilha(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Det = n.Det[:1]
	n.Det[0].Imposto.ICMSUFDest = ICMSUFDest{VBCUFDest: 1000, VICMSUFDest: 80}

	if strings.Contains(gerarDe(t, n), "<ICMSUFDest>") {
		t.Error("ICMSUFDest saiu sem pICMSInterPart")
	}

	n.Det[0].Imposto.ICMSUFDest.PICMSInterPart = 100
	if !strings.Contains(gerarDe(t, n), "<ICMSUFDest>") {
		t.Error("ICMSUFDest ausente com pICMSInterPart preenchido")
	}
}

// ---------------------------------------------------------------------------
// Opcoes de forcar tag
// ---------------------------------------------------------------------------

// TestGerarDet_ForcarTagRejeicao938ObedeceAoAmbiente cobre o
// OcorrenciasVICMSSubstituto (XmlWriter.pas:1943): a opcao tem quatro
// estados, dois deles dependentes do ambiente da nota.
func TestGerarDet_ForcarTagRejeicao938ObedeceAoAmbiente(t *testing.T) {
	casos := []struct {
		forcar   ForcarGeracaoTag
		ambiente pcn.TipoAmbiente
		espera   bool
	}{
		{FgtNunca, pcn.TaProducao, false},
		{FgtNunca, pcn.TaHomologacao, false},
		{FgtSempre, pcn.TaProducao, true},
		{FgtSempre, pcn.TaHomologacao, true},
		{FgtSomenteProducao, pcn.TaProducao, true},
		{FgtSomenteProducao, pcn.TaHomologacao, false},
		{FgtSomenteHomologacao, pcn.TaProducao, false},
		{FgtSomenteHomologacao, pcn.TaHomologacao, true},
	}

	for _, c := range casos {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.TpAmb = c.ambiente
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		// CST 60 com os valores do ST retido zerados: so a opcao abre o grupo.
		n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CST60}

		x, err := GerarXMLComOpcoes(n, Opcoes{ForcarGerarTagRejeicao938: c.forcar})
		if err != nil {
			t.Fatal(err)
		}
		if got := strings.Contains(x, "<vICMSSubstituto>"); got != c.espera {
			t.Errorf("forcar=%d ambiente=%s: vICMSSubstituto presente=%v, "+
				"esperado %v", c.forcar, c.ambiente, got, c.espera)
		}
	}
}

// TestGerarDet_ForcarTagRejeicao906ExigeConsumidorFinal trava a condicao
// EXTRA do OcorrenciasICMSEfetivo (XmlWriter.pas:1959), que o 938 nao tem.
func TestGerarDet_ForcarTagRejeicao906ExigeConsumidorFinal(t *testing.T) {
	montar := func(indFinal ConsumidorFinal) string {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.IndFinal = indFinal
		n.Emit.CRT = CrtRegimeNormal
		n.Det = n.Det[:1]
		n.Det[0].Imposto.ICMS = ICMS{CST: pcn.CST60}

		x, err := GerarXMLComOpcoes(n, Opcoes{ForcarGerarTagRejeicao906: FgtSempre})
		if err != nil {
			t.Fatal(err)
		}
		return x
	}

	if !strings.Contains(montar(CfConsumidorFinal), "<vICMSEfet>") {
		t.Error("vICMSEfet ausente numa venda a consumidor final com a opcao " +
			"ligada")
	}
	if strings.Contains(montar(CfNao), "<vICMSEfet>") {
		t.Error("vICMSEfet saiu numa venda que nao e a consumidor final; a " +
			"opcao 906 so vale nesse caso")
	}
}

// ---------------------------------------------------------------------------
// Divergencias deliberadas do ACBr
// ---------------------------------------------------------------------------

// TestGerarDet_PISSTSobrescreveQuandoAsDuasBasesSaoInformadas documenta uma
// divergencia REPLICADA de proposito.
//
// No XmlWriter.pas:3150-3177 os dois ifs nao sao excludentes e o segundo
// sobrescreve o Result do primeiro: preenchendo vBC/pPIS E qBCProd/vAliqProd,
// o primeiro grupo e perdido. O ACBr emite um alerta e segue. Replicado
// porque e deterministico -- e este teste e o que impede alguem de "consertar"
// sem perceber que esta divergindo do original.
func TestGerarDet_PISSTSobrescreveQuandoAsDuasBasesSaoInformadas(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Det = n.Det[:1]
	n.Det[0].Imposto.PISST = PISST{
		VBC: 100, PPIS: 1.65,
		QBCProd: 10, VAliqProd: 0.50,
		VPIS: 5,
	}

	x := fragmentoDoGrupo(t, gerarDe(t, n), "PISST")
	if strings.Contains(x, "<vBC>") {
		t.Error("o grupo por VALOR sobreviveu; no ACBr o bloco por quantidade " +
			"sobrescreve o Result e so ele sai")
	}
	if !strings.Contains(x, "<qBCProd>") {
		t.Error("o grupo por quantidade deveria ser o que sai")
	}
}
