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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Os testes deste arquivo cobrem o gerador do cabecalho, totais, transporte e
// cobranca. O gerador do item (det) entra na etapa 4b.
//
// O teste mais forte e o round-trip gerar -> ler -> gerar: ele pega campo
// gerado que o leitor nao le e campo lido que o gerador nao escreve, os dois
// de uma vez. Os demais travam as condicionais que o round-trip sozinho nao
// discrimina, porque dependem da VERSAO do leiaute.

// TestGerar_RoundTripEstavel e o teste central da etapa.
//
// Gera, le o resultado, gera de novo e compara byte a byte. Uma diferenca
// entre as geracoes so aparece se algum campo tiver sido perdido no caminho
// -- escrito e nao lido, ou lido e nao escrito.
//
// Duas ressalvas, as duas deliberadas:
//
//   - perda FIEL passa no teste. O det, por exemplo, e lido e nao escrito
//     (etapa 4b), e as geracoes concordam em nao emiti-lo;
//   - a comparacao obrigatoria e entre a SEGUNDA e a TERCEIRA geracoes. A
//     primeira passagem NORMALIZA o documento (monta a chave, aplica o
//     literal de homologacao, resolve o CDATA do qrCode), e e dali em diante
//     que o gerador tem de ser ponto fixo. Para as fixtures ja canonicas a
//     primeira tambem e conferida.
func TestGerar_RoundTripEstavel(t *testing.T) {
	fixtures := []struct {
		nome string
		// canonica diz se a PRIMEIRA geracao ja deve bater com a segunda.
		// Falso so para entrada que o proprio ACBr nao reproduz -- ver
		// TestGerar_CDATAAninhadoNaoSobreviveAoRoundTrip.
		canonica bool
	}{
		{"nfe_400_completa.xml", true},
		{"nfe_310_versao_antiga.xml", true},
		{"nfe_200_data_sem_hora.xml", true},
		{"nfce_proc_autorizada.xml", false},
	}

	for _, f := range fixtures {
		t.Run(f.nome, func(t *testing.T) {
			x1 := gerarDe(t, lerFixture(t, f.nome))
			x2 := gerarDe(t, lerXML(t, x1))
			x3 := gerarDe(t, lerXML(t, x2))

			if x2 != x3 {
				t.Errorf("gerador nao e ponto fixo\n--- 2a geracao ---\n%s\n--- 3a ---\n%s",
					primeiraDiferenca(x2, x3), primeiraDiferenca(x3, x2))
			}

			if f.canonica && x1 != x2 {
				t.Errorf("round-trip instavel ja na primeira geracao\n--- 1a ---\n%s\n--- 2a ---\n%s",
					primeiraDiferenca(x1, x2), primeiraDiferenca(x2, x1))
			}
		})
	}
}

// gerarDe gera o XML e falha o teste no erro.
func gerarDe(t *testing.T, n *NFe) string {
	t.Helper()
	x, err := GerarXML(n)
	if err != nil {
		t.Fatalf("GerarXML: %v", err)
	}
	return x
}

// lerXML le o XML e falha o teste no erro.
func lerXML(t *testing.T, xml string) *NFe {
	t.Helper()
	n, err := LerXMLString(xml)
	if err != nil {
		t.Fatalf("LerXMLString: %v", err)
	}
	return n
}

// TestGerar_CDATAAninhadoNaoSobreviveAoRoundTrip documenta a unica entrada
// que o round-trip nao reproduz -- e que o ACBr tambem nao reproduziria.
//
// A fixture da NFC-e traz o qrCode com CDATA DUPLO. O leitor remove so a
// primeira ocorrencia de cada marcador (StringReplace sem rfReplaceAll), e o
// modelo fica com um CDATA interno sobrando. O gerador entao ve um qrCode com
// "?p=" e o grava CRU, marcador e tudo. Na releitura o parser XML consome
// esse CDATA e devolve so a URL -- e dai em diante o documento e estavel.
//
// Nao e defeito do porte: e o comportamento combinado das duas regras do
// ACBr, cada uma ja travada por teste proprio.
func TestGerar_CDATAAninhadoNaoSobreviveAoRoundTrip(t *testing.T) {
	n := lerFixture(t, "nfce_proc_autorizada.xml")

	if !strings.Contains(n.InfNFeSupl.QRCode, "<![CDATA[") {
		t.Fatal("a fixture deveria deixar um CDATA interno no modelo apos a leitura")
	}

	x1 := gerarDe(t, n)
	if !strings.Contains(x1, "<![CDATA[") {
		t.Error("a primeira geracao deveria gravar o CDATA residual cru, porque o " +
			"qrCode contem ?p=")
	}

	// A partir da releitura o marcador some e o documento se estabiliza.
	n2 := lerXML(t, x1)
	if strings.Contains(n2.InfNFeSupl.QRCode, "<![CDATA[") {
		t.Error("o CDATA residual deveria ter sido consumido pelo parser na releitura")
	}
}

// primeiraDiferenca devolve um trecho em volta do primeiro byte divergente,
// para a mensagem de erro caber na tela.
func primeiraDiferenca(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	ini := i - 80
	if ini < 0 {
		ini = 0
	}
	fim := i + 120
	if fim > len(a) {
		fim = len(a)
	}
	return a[ini:fim]
}

// TestGerar_ChaveDeAcessoReproduzAFixture prova que o montador da chave esta
// certo: a chave calculada a partir dos campos tem de bater com a que o XML
// de origem ja trazia, digito verificador incluso.
func TestGerar_ChaveDeAcessoReproduzAFixture(t *testing.T) {
	casos := []struct {
		fixture string
		chave   string
	}{
		{"nfe_400_completa.xml", "NFe35260199999999000191550010000001231123456784"},
		{"nfce_proc_autorizada.xml", "NFe35260199999999000191650010000004561123456792"},
	}

	for _, c := range casos {
		t.Run(c.fixture, func(t *testing.T) {
			n := lerFixture(t, c.fixture)
			if _, err := GerarXML(n); err != nil {
				t.Fatalf("GerarXML: %v", err)
			}
			if n.InfNFe.ID != c.chave {
				t.Errorf("chave = %q, esperado %q", n.InfNFe.ID, c.chave)
			}
			if err := ValidarChaveAcesso(n); err != nil {
				t.Errorf("chave gerada nao valida: %v", err)
			}
		})
	}
}

// TestGerar_PosicaoDoNFrefMudaComAVersao trava a condicional mais sutil do
// gerador.
//
// O MESMO grupo sai em dois lugares diferentes do XML conforme a versao:
// abaixo da 3 ele vem logo depois do cMunFG (antes do tpImp); da 3 em diante,
// so depois do xJust. Gerar na posicao errada produz XML que o XSD recusa, e
// nenhum teste de "o campo esta la" acusa.
func TestGerar_PosicaoDoNFrefMudaComAVersao(t *testing.T) {
	t.Run("versao 4.00: NFref depois do verProc", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		x, err := GerarXML(n)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Index(x, "<NFref>") < strings.Index(x, "<tpImp>") {
			t.Error("NFref saiu ANTES do tpImp numa nota 4.00; a partir da versao 3 " +
				"ele vem depois do verProc/xJust")
		}
	})

	t.Run("versao 2.00: NFref antes do tpImp", func(t *testing.T) {
		n := lerFixture(t, "nfe_200_data_sem_hora.xml")
		// A fixture 2.00 nao tem NFref; acrescenta um para medir a posicao.
		n.Ide.NFref = []NFref{{RefNFe: "35260199999999000191550010000001231123456784"}}

		x, err := GerarXML(n)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Index(x, "<NFref>") > strings.Index(x, "<tpImp>") {
			t.Error("NFref saiu DEPOIS do tpImp numa nota 2.00; abaixo da versao 3 " +
				"ele vem logo depois do cMunFG")
		}
	})
}

// TestGerar_ExportaTemDuasFormasPorVersao trava o grupo que muda de ESQUEMA,
// nao so de campo: da 3.10 em diante sao UFSaidaPais/xLocExporta; abaixo
// dela, UFEmbarq/xLocEmbarq. Nao ha tag em comum entre as duas formas.
func TestGerar_ExportaTemDuasFormasPorVersao(t *testing.T) {
	t.Run("4.00 usa UFSaidaPais", func(t *testing.T) {
		x, err := GerarXML(lerFixture(t, "nfe_400_completa.xml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(x, "<UFSaidaPais>") {
			t.Error("UFSaidaPais ausente numa nota 4.00")
		}
		if strings.Contains(x, "<UFEmbarq>") {
			t.Error("UFEmbarq presente numa nota 4.00: e a forma antiga do grupo")
		}
	})

	t.Run("2.00 usa UFEmbarq", func(t *testing.T) {
		n := lerFixture(t, "nfe_200_data_sem_hora.xml")
		n.Exporta = Exporta{
			UFEmbarq:    "SP",
			XLocEmbarq:  "SANTOS",
			UFSaidaPais: "SP",
			XLocExporta: "SANTOS",
		}
		x, err := GerarXML(n)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(x, "<UFEmbarq>") {
			t.Error("UFEmbarq ausente numa nota 2.00")
		}
		if strings.Contains(x, "<UFSaidaPais>") {
			t.Error("UFSaidaPais presente numa nota 2.00: so existe da 3.10 em diante")
		}
	})
}

// TestGerar_PagamentoMudaDeFormaComAVersao trava a reestruturacao do grupo
// de pagamento na 4.00: um <pag> com varios <detPag> e um <vTroco>, contra
// varios <pag> soltos sem nenhum dos dois.
func TestGerar_PagamentoMudaDeFormaComAVersao(t *testing.T) {
	t.Run("4.00 agrupa em pag/detPag com vTroco", func(t *testing.T) {
		x, err := GerarXML(lerFixture(t, "nfe_400_completa.xml"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(x, "<detPag>") {
			t.Error("detPag ausente numa nota 4.00")
		}
		// vTroco zerado e opcional: o AddNode do ACBr omite. Preenche
		// para conferir que a tag existe e sai no fim do grupo.
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Pag.VTroco = 1.5
		comTroco := gerarDe(t, n)
		if !strings.Contains(comTroco, "<vTroco>1.50</vTroco>") {
			t.Error("vTroco nao saiu com valor preenchido")
		}
		if strings.Contains(x, "<vTroco>") {
			t.Error("vTroco zerado saiu; sendo opcional, o ACBr omite")
		}
		if !strings.Contains(x, "<indPag>") {
			t.Error("indPag ausente no detPag da 4.00")
		}
	})

	t.Run("3.10 emite pag solto, sem detPag nem vTroco", func(t *testing.T) {
		n := lerFixture(t, "nfe_310_versao_antiga.xml")
		// modelo 55 na 3.10 nao emite pag (so NFC-e); muda para 65.
		n.Ide.Modelo = 65
		x, err := GerarXML(n)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(x, "<detPag>") {
			t.Error("detPag presente numa nota 3.10: o envelope so existe da 4.00 em diante")
		}
		if strings.Contains(x, "<vTroco>") {
			t.Error("vTroco presente numa nota 3.10: so existe da 4.00 em diante")
		}
		if !strings.Contains(x, "<tPag>") {
			t.Error("o pagamento nao foi gerado")
		}
	})
}

// TestGerar_EfeitosColateraisNoDocumento trava as mutacoes que o gerador faz
// no modelo -- replicadas do ACBr de proposito, porque o chamador depende
// delas (o ID assinado, o cDV impresso no DANFE).
func TestGerar_EfeitosColateraisNoDocumento(t *testing.T) {
	t.Run("homologacao sobrescreve o nome do destinatario", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.TpAmb = pcn.TaHomologacao
		n.Dest.XNome = "CLIENTE REAL LTDA"

		if _, err := GerarXML(n); err != nil {
			t.Fatal(err)
		}
		if n.Dest.XNome != nomeDestHomologacao {
			t.Errorf("dest.xNome = %q; em homologacao o ACBr sobrescreve com o "+
				"literal exigido pela SEFAZ", n.Dest.XNome)
		}
	})

	t.Run("abaixo da 3.10 indIEDest e forcado a contribuinte", func(t *testing.T) {
		n := lerFixture(t, "nfe_200_data_sem_hora.xml")
		n.Dest.IndIEDest = pcn.InNaoContribuinte

		x, err := GerarXML(n)
		if err != nil {
			t.Fatal(err)
		}
		if n.Dest.IndIEDest != pcn.InContribuinte {
			t.Errorf("indIEDest = %v; abaixo da 3.10 o ACBr forca inContribuinte",
				n.Dest.IndIEDest)
		}
		if strings.Contains(x, "<indIEDest>") {
			t.Error("a tag indIEDest saiu numa nota 2.00; ela so existe da 3.10 em diante")
		}
	})

	t.Run("ID, cDV e cNF sao gravados a partir da chave", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.InfNFe.ID = ""
		n.Ide.CDV = 0

		if _, err := GerarXML(n); err != nil {
			t.Fatal(err)
		}
		if n.InfNFe.ID == "" {
			t.Error("infNFe.ID nao foi gravado")
		}
		if n.Ide.CDV != 4 {
			t.Errorf("cDV = %d, esperado 4", n.Ide.CDV)
		}
		if n.Ide.CNF != 12345678 {
			t.Errorf("cNF = %d, esperado 12345678", n.Ide.CNF)
		}
	})
}

// TestGerar_CNFSentinelaNaoSorteia cobre a ponta do sentinela -2 no gerador.
//
// O leitor grava -2 quando a tag cNF vem zerada. O gerador trata 0 e -1 como
// "sorteie" e qualquer outro negativo como zero -- entao reler e regerar uma
// nota sem cNF produz cNF zero, nao um numero aleatorio. E o que torna o
// round-trip estavel.
func TestGerar_CNFSentinelaNaoSorteia(t *testing.T) {
	n := lerFixture(t, "nfe_310_versao_antiga.xml")
	if n.Ide.CNF != -2 {
		t.Fatalf("a fixture deveria trazer cNF = -2, veio %d", n.Ide.CNF)
	}

	if _, err := GerarXML(n); err != nil {
		t.Fatal(err)
	}
	if n.Ide.CNF != 0 {
		t.Errorf("cNF = %d, esperado 0: o sentinela -2 NAO dispara sorteio", n.Ide.CNF)
	}
}

// TestGerar_CPaisDoEmitenteEhSempreBrasil trava um literal do ACBr que
// parece defeito: o cPais do emitente nao vem do modelo, e a constante
// CODIGO_BRASIL. Replicado -- o emitente de NF-e e sempre brasileiro.
func TestGerar_CPaisDoEmitenteEhSempreBrasil(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Emit.EnderEmit.CPais = 9999

	x, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(x, "<cPais>1058</cPais>") {
		t.Error("o cPais do emitente nao saiu como 1058; o ACBr grava a constante, " +
			"nao o campo do modelo")
	}
	if strings.Contains(x, "<cPais>9999</cPais>") {
		t.Error("o cPais do modelo vazou para o XML do emitente")
	}
}

// TestGerar_GuardsDosGruposOpcionais trava os guards que decidem se um grupo
// inteiro sai.
//
// Cada um deles olha um conjunto de campos, e tirar um campo do conjunto faz
// o grupo sumir de notas em que ele deveria aparecer.
func TestGerar_GuardsDosGruposOpcionais(t *testing.T) {
	t.Run("ISSQNtot some sem nenhum dos cinco valores", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		n.Total.ISSQNTot = ISSQNTot{}
		x, _ := GerarXML(n)
		if strings.Contains(x, "<ISSQNtot>") {
			t.Error("ISSQNtot saiu com todos os valores zerados")
		}

		// basta UM dos cinco para o grupo voltar
		n = lerFixture(t, "nfe_400_completa.xml")
		n.Total.ISSQNTot = ISSQNTot{VISS: 1}
		x, _ = GerarXML(n)
		if !strings.Contains(x, "<ISSQNtot>") {
			t.Error("ISSQNtot sumiu com vISS preenchido")
		}
	})

	t.Run("retTrib some sem nenhum dos sete valores", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		x, _ := GerarXML(n)
		if strings.Contains(x, "<retTrib>") {
			t.Error("retTrib saiu numa nota sem retencao")
		}

		n.Total.RetTrib = RetTrib{VRetPrev: 1}
		x, _ = GerarXML(n)
		if !strings.Contains(x, "<retTrib>") {
			t.Error("retTrib sumiu com vRetPrev preenchido; o guard olha os SETE campos")
		}
	})

	t.Run("retirada e entrega dependem do logradouro", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")
		x, _ := GerarXML(n)
		if !strings.Contains(x, "<retirada>") {
			t.Error("retirada sumiu apesar do logradouro preenchido")
		}

		// O guard e o xLgr, nao o CNPJ.
		n.Retirada.XLgr = ""
		x, _ = GerarXML(n)
		if strings.Contains(x, "<retirada>") {
			t.Error("retirada saiu sem logradouro; o guard do ACBr e o xLgr")
		}
	})

	t.Run("infIntermed so a partir da 4.00", func(t *testing.T) {
		n := lerFixture(t, "nfe_310_versao_antiga.xml")
		n.InfIntermed = InfIntermed{CNPJ: "99999999000191", IDCadIntTran: "X"}
		x, _ := GerarXML(n)
		if strings.Contains(x, "<infIntermed>") {
			t.Error("infIntermed saiu numa nota 3.10; o grupo so existe da 4.00 em diante")
		}
	})
}

// TestGerar_AgropecuarioEhExcludente trava um else-if do ACBr que parece
// dois ifs: havendo defensivo, a guia de transito NAO sai, nem preenchida.
func TestGerar_AgropecuarioEhExcludente(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	// A fixture tem os DOIS preenchidos.
	if len(n.Agropecuario.Defensivo) == 0 || n.Agropecuario.GuiaTransito.TpGuia == TpgNenhum {
		t.Fatal("a fixture deveria trazer defensivo E guiaTransito preenchidos")
	}

	x, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(x, "<defensivo>") {
		t.Error("defensivo nao saiu")
	}
	if strings.Contains(x, "<guiaTransito>") {
		t.Error("guiaTransito saiu junto do defensivo; no ACBr os dois sao " +
			"EXCLUDENTES (else-if), e o defensivo tem precedencia")
	}

	// Sem defensivo, a guia aparece.
	n.Agropecuario.Defensivo = nil
	x, _ = GerarXML(n)
	if !strings.Contains(x, "<guiaTransito>") {
		t.Error("guiaTransito nao saiu na ausencia de defensivo")
	}
}

// TestGerar_QRCodeCDATAEhCondicional trava a regra do infNFeSupl: o ACBr so
// embrulha o qrCode em CDATA quando ele NAO contem "?p=".
func TestGerar_QRCodeCDATAEhCondicional(t *testing.T) {
	n := lerFixture(t, "nfce_proc_autorizada.xml")

	t.Run("com ?p= sai cru", func(t *testing.T) {
		n.InfNFeSupl.QRCode = "https://exemplo.invalid/qr?p=123|2|1"
		x, _ := GerarXML(n)
		if strings.Contains(x, "<qrCode><![CDATA[") {
			t.Error("o qrCode com ?p= foi embrulhado em CDATA; o ACBr so embrulha " +
				"quando a query string nao esta presente")
		}
	})

	t.Run("sem ?p= sai em CDATA", func(t *testing.T) {
		n.InfNFeSupl.QRCode = "texto qualquer"
		x, _ := GerarXML(n)
		if !strings.Contains(x, "<qrCode><![CDATA[texto qualquer]]></qrCode>") {
			t.Error("o qrCode sem ?p= nao foi embrulhado em CDATA")
		}
	})
}

// TestGerar_EnvelopeProcSoComProtocolo confirma que a raiz muda conforme o
// protocolo esteja preenchido.
func TestGerar_EnvelopeProcSoComProtocolo(t *testing.T) {
	n := lerFixture(t, "nfce_proc_autorizada.xml")

	x, err := GerarXML(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(x, "<nfeProc") {
		t.Errorf("a raiz deveria ser nfeProc com protocolo preenchido, veio %.40s", x)
	}
	if !strings.Contains(x, "<protNFe") {
		t.Error("protNFe nao saiu")
	}

	n.ProcNFe.NProt = ""
	x, _ = GerarXML(n)
	if !strings.HasPrefix(x, "<NFe") {
		t.Errorf("sem protocolo a raiz deveria ser NFe, veio %.40s", x)
	}
	if strings.Contains(x, "<protNFe") {
		t.Error("protNFe saiu sem nProt")
	}
}

// TestGerar_IEIsentoNaoEhFiltrada trava a excecao do emitente e do
// transportador: o literal "ISENTO" passa inteiro, qualquer outro valor
// perde a formatacao.
func TestGerar_IEIsentoNaoEhFiltrada(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	n.Emit.IE = "ISENTO"
	x, _ := GerarXML(n)
	if !strings.Contains(x, "<IE>ISENTO</IE>") {
		t.Error("a IE ISENTO do emitente foi filtrada; o literal tem de passar inteiro")
	}

	n.Emit.IE = "111.111.111.111"
	x, _ = GerarXML(n)
	if !strings.Contains(x, "<IE>111111111111</IE>") {
		t.Error("a IE numerica do emitente nao passou pelo OnlyNumber")
	}
}

// TestGerar_CNAESoComIM cobre a NT 2013/005 v1.03, replicada do ACBr: o CNAE
// so pode ser informado junto do IM.
func TestGerar_CNAESoComIM(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")
	n.Emit.CNAE = "4751201"

	n.Emit.IM = ""
	x, _ := GerarXML(n)
	if strings.Contains(x, "<CNAE>") {
		t.Error("CNAE saiu sem IM; a NT 2013/005 so o admite junto do IM")
	}

	n.Emit.IM = "123456"
	x, _ = GerarXML(n)
	if !strings.Contains(x, "<CNAE>4751201</CNAE>") {
		t.Error("CNAE nao saiu com IM preenchido")
	}
}

// TestGerar_DocumentoNiloNaoEntraEmPanico confirma a regra de biblioteca.
func TestGerar_DocumentoNiloNaoEntraEmPanico(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("entrou em panico: %v", r)
		}
	}()

	if _, err := GerarXML(nil); err == nil {
		t.Error("GerarXML(nil) deveria devolver erro")
	}

	// Documento zerado tambem nao pode derrubar.
	vazia := NovaNFe()
	if _, err := GerarXML(&vazia); err != nil {
		t.Logf("nota vazia devolveu erro (aceitavel): %v", err)
	}
}
