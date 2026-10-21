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

// Testes de ORDEM e de condicional fina do gerador.
//
// Existem porque o round-trip e CEGO PARA ORDEM: o leitor acha tag por nome,
// entao gerar -> ler -> gerar fica estavel mesmo com os campos fora de
// sequencia. O XSD, porem, e xs:sequence -- campo fora de ordem e documento
// rejeitado pela SEFAZ.
//
// Foi exatamente assim que a inversao de tpGuia/UFGuia passou por todos os
// outros testes desta etapa e so apareceu na conferencia com o .pas.

// ordemDasTags confere que as tags aparecem NA ORDEM informada, e nao apenas
// que existem.
func ordemDasTags(t *testing.T, xml string, tags ...string) {
	t.Helper()
	anterior := -1
	anteriorNome := ""
	for _, tag := range tags {
		i := posicaoDaTag(xml, tag)
		if i < 0 {
			t.Errorf("tag <%s> ausente", tag)
			return
		}
		if i < anterior {
			t.Errorf("<%s> saiu ANTES de <%s>; a sequencia esperada e %v",
				tag, anteriorNome, tags)
			return
		}
		anterior, anteriorNome = i, tag
	}
}

// posicaoDaTag devolve o indice da abertura da tag, aceitando tanto a forma
// simples (<forDia>) quanto a com atributo (<forDia dia="1">).
func posicaoDaTag(xml, tag string) int {
	if i := strings.Index(xml, "<"+tag+">"); i >= 0 {
		return i
	}
	return strings.Index(xml, "<"+tag+" ")
}

// TestGerar_OrdemDosCamposSegueOXSD trava a sequencia dos grupos.
func TestGerar_OrdemDosCamposSegueOXSD(t *testing.T) {
	x := gerarDe(t, lerFixture(t, "nfe_400_completa.xml"))

	t.Run("guiaTransito comeca por tpGuia", func(t *testing.T) {
		// Os IDs do leiaute sugerem o contrario (ZF05 = UFGuia, ZF06 =
		// tpGuia), mas o XSD (leiauteNFe_v4.00.xsd:6539) e o .pas
		// (XmlWriter.pas:1088-1092) mandam tpGuia primeiro.
		semDefensivo := lerFixture(t, "nfe_400_completa.xml")
		semDefensivo.Agropecuario.Defensivo = nil
		ordemDasTags(t, gerarDe(t, semDefensivo),
			"guiaTransito", "tpGuia", "UFGuia", "serieGuia", "nGuia")
	})

	t.Run("ide", func(t *testing.T) {
		ordemDasTags(t, x, "cUF", "cNF", "natOp", "mod", "serie", "nNF",
			"dhEmi", "dhSaiEnt", "tpNF", "idDest", "cMunFG", "tpImp",
			"tpEmis", "cDV", "tpAmb", "finNFe", "indFinal", "indPres",
			"procEmi", "verProc", "NFref")
	})

	t.Run("emit e enderEmit", func(t *testing.T) {
		ordemDasTags(t, x, "emit", "xNome", "xFant", "enderEmit", "xLgr",
			"nro", "xBairro", "cMun", "xMun", "UF", "CEP", "cPais", "xPais")
	})

	t.Run("ICMSTot", func(t *testing.T) {
		ordemDasTags(t, x, "ICMSTot", "vBC", "vICMS", "vICMSDeson", "vFCP",
			"vBCST", "vST", "vFCPST", "vFCPSTRet", "vProd", "vFrete", "vSeg",
			"vDesc", "vII", "vIPI", "vIPIDevol", "vPIS", "vCOFINS", "vOutro",
			"vNF")
	})

	t.Run("transp", func(t *testing.T) {
		ordemDasTags(t, x, "transp", "modFrete", "transporta", "veicTransp",
			"reboque", "vol")
	})

	t.Run("cana", func(t *testing.T) {
		ordemDasTags(t, x, "cana", "safra", "ref", "forDia", "qTotMes",
			"qTotAnt", "qTotGer", "vFor", "vTotDed", "vLiqFor")
	})

	t.Run("infNFe: grupos de primeiro nivel", func(t *testing.T) {
		ordemDasTags(t, x, "ide", "emit", "dest", "retirada", "autXML",
			"total", "transp", "cobr", "pag", "infAdic", "exporta", "compra",
			"cana", "infRespTec", "agropecuario")
	})
}

// TestGerar_GatesInternosDoICMSTot trava os dois cortes de versao DENTRO do
// grupo de totais -- um em 3.10 e outro em 4.00.
func TestGerar_GatesInternosDoICMSTot(t *testing.T) {
	t.Run("2.00 nao tem vICMSDeson nem vFCP", func(t *testing.T) {
		x := gerarDe(t, lerFixture(t, "nfe_200_data_sem_hora.xml"))
		if strings.Contains(x, "<vICMSDeson>") {
			t.Error("vICMSDeson saiu numa 2.00; o campo entrou na 3.10")
		}
		if strings.Contains(x, "<vFCP>") {
			t.Error("vFCP saiu numa 2.00; o campo entrou na 4.00")
		}
	})

	t.Run("3.10 tem vICMSDeson mas nao vFCP", func(t *testing.T) {
		x := gerarDe(t, lerFixture(t, "nfe_310_versao_antiga.xml"))
		if !strings.Contains(x, "<vICMSDeson>") {
			t.Error("vICMSDeson ausente numa 3.10")
		}
		if strings.Contains(x, "<vFCP>") {
			t.Error("vFCP saiu numa 3.10; o campo so entra na 4.00")
		}
		if strings.Contains(x, "<vIPIDevol>") {
			t.Error("vIPIDevol saiu numa 3.10; o campo so entra na 4.00")
		}
	})

	t.Run("4.00 tem os dois", func(t *testing.T) {
		x := gerarDe(t, lerFixture(t, "nfe_400_completa.xml"))
		for _, tag := range []string{"vICMSDeson", "vFCP", "vFCPST", "vIPIDevol"} {
			if !strings.Contains(x, "<"+tag+">") {
				t.Errorf("%s ausente numa 4.00", tag)
			}
		}
	})
}

// TestGerar_FatCamposOpcionaisReproduzODefaultDoACBr cobre a opcao invertida.
//
// O Create do ACBr LIGA CamposFatObrigatorios (XmlWriter.pas:220). Como o
// zero value de um bool em Go e false, a opcao foi invertida no porte para
// que Opcoes{} reproduza o default do original.
func TestGerar_FatCamposOpcionaisReproduzODefaultDoACBr(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Cobr.Fat.VDesc = 0

	// Default: campos obrigatorios, entao o vDesc zerado sai como "0.00"
	// -- uma vez no fat e uma no ICMSTot.
	if got := strings.Count(gerarDe(t, n), "<vDesc>0.00</vDesc>"); got != 2 {
		t.Errorf("com as opcoes padrao ha %d <vDesc>0.00</vDesc>, esperado 2 "+
			"(fat + ICMSTot); o Create do ACBr liga CamposFatObrigatorios", got)
	}

	// Com a opcao, o campo do fat vira opcional e some.
	x, err := GerarXMLComOpcoes(n, Opcoes{FatCamposOpcionais: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(x, "<vDesc>0.00</vDesc>"); got != 1 {
		t.Errorf("com FatCamposOpcionais ha %d <vDesc>0.00</vDesc>, esperado 1 "+
			"(so o do ICMSTot)", got)
	}
}

// TestGerar_SignatureURIVemDaChaveRecemMontada trava o efeito colateral do
// .pas:373.
//
// Sem ele, uma nota cujo cNF foi re-sorteado ficaria com o Reference
// apontando para a chave antiga -- assinatura invalida sem nenhum erro.
func TestGerar_SignatureURIVemDaChaveRecemMontada(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Signature = pcn.Signature{
		URI:             "#NFeCHAVEANTIGA",
		DigestValue:     "ZGlnZXN0",
		SignatureValue:  "YXNzaW5hdHVyYQ==",
		X509Certificate: "Y2VydA==",
	}

	x := gerarDe(t, n)

	esperado := "#NFe" + pcn.RemoverLiteralChave(n.InfNFe.ID)
	if n.Signature.URI != esperado {
		t.Errorf("signature.URI = %q, esperado %q", n.Signature.URI, esperado)
	}
	if !strings.Contains(x, `URI="`+esperado+`"`) {
		t.Error("o Reference do XML nao aponta para a chave gerada")
	}
	if strings.Contains(x, "CHAVEANTIGA") {
		t.Error("o URI antigo vazou para o XML")
	}
}

// TestGerar_CIndOpGanhaZeroAEsquerda cobre o unico campo tcNumStr do
// cabecalho: o AddNode do ACBr completa ate o tamanho minimo.
func TestGerar_CIndOpGanhaZeroAEsquerda(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Ide.CIndOp = "123"

	if x := gerarDe(t, n); !strings.Contains(x, "<cIndOp>000123</cIndOp>") {
		t.Error("cIndOp nao foi completado ate 6 posicoes; o tcNumStr do ACBr " +
			"aplica PadLeft com zeros")
	}

	// Vazio continua sumindo: o campo e opcional.
	n.Ide.CIndOp = ""
	if x := gerarDe(t, n); strings.Contains(x, "<cIndOp>") {
		t.Error("cIndOp vazio saiu; sendo opcional, deveria sumir")
	}
}

// TestGerar_AvulsaUsaOCNPJDoFiscoNaChave trava a regra do .pas:311-314: numa
// nota avulsa a chave de acesso usa o CNPJ do fisco emitente, nao o do
// contribuinte.
func TestGerar_AvulsaUsaOCNPJDoFiscoNaChave(t *testing.T) {
	semAvulsa := lerFixture(t, "nfe_400_completa.xml")
	xSem := gerarDe(t, semAvulsa)
	chaveSemAvulsa := semAvulsa.InfNFe.ID

	comAvulsa := lerFixture(t, "nfe_400_completa.xml")
	comAvulsa.Avulsa = Avulsa{
		CNPJ:    "99999999000969",
		XOrgao:  "SEFAZ",
		Matr:    "123",
		XAgente: "AGENTE DE TESTE",
		UF:      "SP",
		RepEmi:  "REP",
	}
	xCom := gerarDe(t, comAvulsa)

	if comAvulsa.InfNFe.ID == chaveSemAvulsa {
		t.Error("a chave nao mudou com a nota avulsa; ela deveria usar o CNPJ " +
			"do fisco emitente, nao o do contribuinte")
	}
	if !strings.Contains(comAvulsa.InfNFe.ID, "99999999000969") {
		t.Errorf("o CNPJ da avulsa nao chegou a chave: %s", comAvulsa.InfNFe.ID)
	}
	if strings.Contains(xSem, "<avulsa>") {
		t.Error("o grupo avulsa saiu sem CNPJ do fisco")
	}
	if !strings.Contains(xCom, "<avulsa>") {
		t.Error("o grupo avulsa nao saiu")
	}
}

// TestGerar_CNPJCPFEscolhidoPeloTamanho trava o AddNodeCNPJCPF: ate 11
// posicoes gera CPF; acima, CNPJ.
func TestGerar_CNPJCPFEscolhidoPeloTamanho(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	n.Dest.CNPJCPF = "11111111111"
	if x := gerarDe(t, n); !strings.Contains(x, "<CPF>11111111111</CPF>") {
		t.Error("documento de 11 digitos no destinatario deveria sair como CPF")
	}

	n.Dest.CNPJCPF = "99999999000272"
	if x := gerarDe(t, n); !strings.Contains(x, "<CNPJ>99999999000272</CNPJ>") {
		t.Error("documento de 14 digitos no destinatario deveria sair como CNPJ")
	}
}

// TestGerar_DhSaiEntSoNaNFeComData cobre as DUAS condicionais que governam a
// data de saida: o modelo e a presenca da data.
func TestGerar_DhSaiEntSoNaNFeComData(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	if x := gerarDe(t, n); !strings.Contains(x, "<dhSaiEnt>") {
		t.Error("dhSaiEnt ausente numa NF-e com data de saida")
	}

	// Modelo 65 nao emite dhSaiEnt, mesmo com a data preenchida.
	n.Ide.Modelo = 65
	if x := gerarDe(t, n); strings.Contains(x, "<dhSaiEnt>") {
		t.Error("dhSaiEnt saiu numa NFC-e; o campo so existe no modelo 55")
	}

	// Sem data, some tambem na NF-e.
	n2 := lerFixture(t, "nfe_400_completa.xml")
	n2.Ide.DSaiEnt = time.Time{}
	if x := gerarDe(t, n2); strings.Contains(x, "<dhSaiEnt>") {
		t.Error("dhSaiEnt saiu sem data de saida")
	}
}

// TestGerar_NormalizacaoDosCamposDeReferencia cobre os campos que a classe
// declara como string mas o ACBr grava com tcInt ou AddNodeCNPJ -- ou seja,
// com normalizacao embutida.
func TestGerar_NormalizacaoDosCamposDeReferencia(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	// CNPJ com mascara tem de sair so com digitos (AddNodeCNPJ).
	n.Ide.NFref[0].RefNF.CNPJ = "99.999.999/0002-72"
	// refNFP/mod e string na classe, mas gravado com tcInt: "4" vira "04".
	n.Ide.NFref[1].RefNFP.Modelo = "4"
	// refECF/nECF e nCOO idem: "0012" num campo de 3 vira "012".
	n.Ide.NFref[2].RefECF.NECF = "0012"
	n.Ide.NFref[2].RefECF.NCOO = "123456"
	n.InfRespTec.CNPJ = "99.999.999/0008-58"

	x := gerarDe(t, n)

	casos := []struct{ tag, esperado string }{
		{"CNPJ", "<CNPJ>99999999000272</CNPJ>"},
		{"refNFP/mod", "<mod>04</mod>"},
		{"nECF", "<nECF>012</nECF>"},
		{"nCOO", "<nCOO>123456</nCOO>"},
		{"infRespTec/CNPJ", "<CNPJ>99999999000858</CNPJ>"},
	}
	for _, c := range casos {
		if !strings.Contains(x, c.esperado) {
			t.Errorf("%s: esperado %s no XML", c.tag, c.esperado)
		}
	}
}
