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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Os testes deste arquivo cobrem o que a comparacao com
// Base/ACBrNFe.XmlReader.pas revelou e que build, vet e um teste ingenuo de
// "leu alguma coisa" NAO pegam: condicional por versao de leiaute, default
// aplicado na ausencia de um grupo, nome de tag diferente do nome do campo,
// precisao por campo e grupo achatado.

func lerFixture(t *testing.T, nome string) *NFe {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join("testdata", nome))
	if err != nil {
		t.Fatalf("lendo fixture %s: %v", nome, err)
	}
	n, err := LerBytes(dados)
	if err != nil {
		t.Fatalf("LerBytes(%s): %v", nome, err)
	}
	return n
}

// TestLer_CondicionaisPorVersaoDeLeiaute e o teste central desta etapa.
//
// O leitor da NF-e tem oito condicionais em `infNFe.Versao`, e cada uma
// decide QUAIS TAGS sao lidas. Perder uma nao quebra build nem vet: a nota
// simplesmente chega com o campo vazio, e so a conferencia com o .pas acusa.
func TestLer_CondicionaisPorVersaoDeLeiaute(t *testing.T) {
	t.Run("3.10 e 4.00 leem dhEmi e dhSaiEnt", func(t *testing.T) {
		n := lerFixture(t, "nfe_400_completa.xml")

		if n.Ide.DEmi.IsZero() {
			t.Error("dhEmi nao foi lido; na 4.00 a data vem de dhEmi, nao de dEmi")
		}
		if h, m := n.Ide.DEmi.Hour(), n.Ide.DEmi.Minute(); h != 10 || m != 30 {
			t.Errorf("dhEmi = %dh%02d, esperado 10h30 (a hora faz parte do campo na 4.00)", h, m)
		}
		if !n.Ide.HSaiEnt.IsZero() {
			t.Error("hSaiEnt foi preenchido numa nota 4.00; a tag so existe ate a 3.10")
		}
	})

	t.Run("2.00 le dEmi, dSaiEnt e hSaiEnt", func(t *testing.T) {
		n := lerFixture(t, "nfe_200_data_sem_hora.xml")

		if n.Ide.DEmi.IsZero() {
			t.Fatal("dEmi nao foi lido; abaixo da versao 3 a data vem de dEmi")
		}
		if n.Ide.DEmi.Hour() != 0 {
			t.Errorf("dEmi veio com hora %dh; na 2.00 a data nao carrega hora",
				n.Ide.DEmi.Hour())
		}
		if n.Ide.HSaiEnt.IsZero() {
			t.Error("hSaiEnt nao foi lido; e tag propria e so existe abaixo da versao 3")
		}
		if h, m := n.Ide.HSaiEnt.Hour(), n.Ide.HSaiEnt.Minute(); h != 14 || m != 30 {
			t.Errorf("hSaiEnt = %dh%02d, esperado 14h30", h, m)
		}
	})

	t.Run("tags exclusivas da versao 3 sao ignoradas na 2.00", func(t *testing.T) {
		n := lerFixture(t, "nfe_200_data_sem_hora.xml")

		// A fixture declara idDest=3, indFinal=1 e indPres=5, e nenhum
		// deles deve ser lido abaixo da versao 3.
		if n.Ide.IdDest != DoInterna {
			t.Errorf("idDest = %v numa nota 2.00; so e lido a partir da versao 3",
				n.Ide.IdDest)
		}
		if n.Ide.IndFinal != CfNao {
			t.Errorf("indFinal = %v numa nota 2.00; so e lido a partir da versao 3",
				n.Ide.IndFinal)
		}
		if n.Ide.IndPres != PcNao {
			t.Errorf("indPres = %v numa nota 2.00; so e lido a partir da versao 3",
				n.Ide.IndPres)
		}
		// dest existe, mas indIEDest e IM so a partir da versao 3.
		if n.Dest.IM != "" {
			t.Errorf("IM = %q numa nota 2.00; so e lido a partir da versao 3", n.Dest.IM)
		}
		if n.Total.ISSQNTot.VDeducao != 0 {
			t.Errorf("vDeducao = %v numa nota 2.00; o grupo estendido do "+
				"ISSQNtot so e lido a partir da versao 3", n.Total.ISSQNTot.VDeducao)
		}
	})

	t.Run("tags exclusivas da 4.00 sao ignoradas na 3.10", func(t *testing.T) {
		n := lerFixture(t, "nfe_310_versao_antiga.xml")

		// A fixture TEM <indIntermed>, mas o leitor so o le a partir da 4.00.
		if n.Ide.IndIntermed != IiSemOperacao {
			t.Errorf("indIntermed = %v numa nota 3.10; a tag esta no XML mas so "+
				"deve ser lida a partir da 4.00", n.Ide.IndIntermed)
		}

		// Idem para indEscala e cBenef do produto.
		p := n.Det[0].Prod
		if p.IndEscala != IeNenhum {
			t.Errorf("indEscala = %v numa nota 3.10; so e lido a partir da 4.00", p.IndEscala)
		}
		if p.CBenef != "" {
			t.Errorf("cBenef = %q numa nota 3.10; so e lido a partir da 4.00", p.CBenef)
		}
	})

	t.Run("4.00 le os campos novos do produto", func(t *testing.T) {
		p := lerFixture(t, "nfe_400_completa.xml").Det[0].Prod

		if p.IndEscala != IeNaoRelevante {
			t.Errorf("indEscala = %v, esperado IeNaoRelevante", p.IndEscala)
		}
		if p.CBenef != "SP000001" {
			t.Errorf("cBenef = %q, esperado SP000001", p.CBenef)
		}
		if len(p.CredPresumido) != 1 {
			t.Fatalf("gCred: %d grupos, esperado 1", len(p.CredPresumido))
		}
		if p.CredPresumido[0].PCredPresumido != 1.5 {
			t.Errorf("pCredPresumido = %v, esperado 1.5", p.CredPresumido[0].PCredPresumido)
		}
	})

	t.Run("ISSQNtot estendido e lido de 3 em diante", func(t *testing.T) {
		n := lerFixture(t, "nfe_310_versao_antiga.xml")
		if n.Total.ISSQNTot.VDeducao != 5 {
			t.Errorf("vDeducao = %v, esperado 5; o grupo estendido e lido de 3 em diante",
				n.Total.ISSQNTot.VDeducao)
		}
		if n.Total.ISSQNTot.CRegTrib != pcn.RTISSMicroempresaMunicipal {
			t.Errorf("cRegTrib = %v, esperado RTISSMicroempresaMunicipal",
				n.Total.ISSQNTot.CRegTrib)
		}
	})
}

// TestLer_PagamentoDaVersao310 cobre a DIVERGENCIA deliberada do porte.
//
// No ACBr a variavel AuxNode e reaproveitada: recebe o no <cobr> e so e
// reapontada para <pag> quando Versao >= 4. Na 3.xx, portanto, o laco procura
// <pag> DENTRO de <cobr>, onde ele nao existe, e nenhum pagamento e lido.
// Aqui o no de pagamento e resolvido separadamente, e a 3.10 le do lugar certo.
func TestLer_PagamentoDaVersao310(t *testing.T) {
	n := lerFixture(t, "nfe_310_versao_antiga.xml")

	if len(n.Pag.DetPag) != 1 {
		t.Fatalf("pagamentos lidos: %d, esperado 1. O ACBr le ZERO aqui "+
			"(AuxNode reaproveitado); esta e a divergencia registrada", len(n.Pag.DetPag))
	}
	if got := n.Pag.DetPag[0].TPag; got != FpDinheiro {
		t.Errorf("tPag = %v, esperado FpDinheiro", got)
	}
	if got := n.Pag.DetPag[0].VPag; got != 100 {
		t.Errorf("vPag = %v, esperado 100", got)
	}
}

// TestLer_CNFZeradoViraMenosDois cobre o sentinela do ACBr: cNF ausente ou
// zero vira -2, para o gerador saber que precisa sortear um codigo.
func TestLer_CNFZeradoViraMenosDois(t *testing.T) {
	n := lerFixture(t, "nfe_310_versao_antiga.xml")
	if n.Ide.CNF != -2 {
		t.Errorf("cNF = %d, esperado -2 (sentinela do ACBr para codigo ausente)", n.Ide.CNF)
	}
}

// TestLer_DestAusenteMarcaNaoContribuinte cobre o unico guard do leitor que
// GRAVA um valor quando o node nao existe.
func TestLer_DestAusenteMarcaNaoContribuinte(t *testing.T) {
	n := lerFixture(t, "nfe_310_versao_antiga.xml") // fixture sem <dest>

	if n.Dest.IndIEDest != pcn.InNaoContribuinte {
		t.Errorf("indIEDest = %v com dest ausente, esperado InNaoContribuinte",
			n.Dest.IndIEDest)
	}
}

// TestLer_DefaultDePaisSoNoEmitenteEDestinatario trava uma assimetria do
// leitor: cPais/xPais ganham default em enderEmit e enderDest, e NAO em
// retirada nem entrega.
func TestLer_DefaultDePaisSoNoEmitenteEDestinatario(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	// A fixture nao declara cPais/xPais em lugar nenhum.
	if n.Emit.EnderEmit.CPais != 1058 || n.Emit.EnderEmit.XPais != "BRASIL" {
		t.Errorf("enderEmit = %d/%q, esperado 1058/BRASIL",
			n.Emit.EnderEmit.CPais, n.Emit.EnderEmit.XPais)
	}
	if n.Dest.EnderDest.CPais != 1058 || n.Dest.EnderDest.XPais != "BRASIL" {
		t.Errorf("enderDest = %d/%q, esperado 1058/BRASIL",
			n.Dest.EnderDest.CPais, n.Dest.EnderDest.XPais)
	}

	// Retirada NAO tem o tratamento -- e assim no .pas.
	if n.Retirada.CPais != 0 || n.Retirada.XPais != "" {
		t.Errorf("retirada = %d/%q, esperado 0/vazio: o ACBr nao aplica "+
			"default de pais aqui", n.Retirada.CPais, n.Retirada.XPais)
	}
}

// TestLer_NItemVemDoContadorNaoDoAtributo trava a numeracao dos itens.
//
// A fixture declara <det nItem="99"> no primeiro item de proposito. O leitor
// do ACBr ignora o atributo e numera pela posicao.
func TestLer_NItemVemDoContadorNaoDoAtributo(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	if len(n.Det) != 4 {
		t.Fatalf("itens lidos: %d, esperado 4", len(n.Det))
	}
	if n.Det[0].Prod.NItem != 1 {
		t.Errorf("nItem do primeiro item = %d, esperado 1; o XML diz 99 e o "+
			"ACBr ignora o atributo", n.Det[0].Prod.NItem)
	}
	if n.Det[1].Prod.NItem != 2 {
		t.Errorf("nItem do segundo item = %d, esperado 2", n.Det[1].Prod.NItem)
	}
}

// TestLer_PrecisaoPorCampo confere que os valores com muitas casas chegam
// inteiros ao modelo.
//
// ATENCAO ao que este teste NAO faz: ele nao distingue ConteudoDe2 de
// ConteudoDe10. O parametro de casas do pcn.ConteudoDec so tem efeito quando
// o node esta marcado com FloatIsIntString, que o parser nunca liga -- no XML
// da SEFAZ o valor ja vem com ponto decimal. A precisao declarada campo a
// campo e CONTRATO (herdado do mesmo mecanismo do ACBr), nao conversao, e
// trocar De10 por De2 aqui nao mudaria nenhum resultado.
//
// Quem trava a precisao de verdade e ferramentas/conferir-leitor-dfe.py, que
// compara (tag, tcDeN) do .pas com (tag, ConteudoDeN) do .go.
func TestLer_PrecisaoPorCampo(t *testing.T) {
	p := lerFixture(t, "nfe_400_completa.xml").Det[0].Prod

	if p.VUnCom != 333.3333333333 {
		t.Errorf("vUnCom = %v, esperado 333.3333333333 (tcDe10)", p.VUnCom)
	}
	if p.VUnTrib != 333.3333333333 {
		t.Errorf("vUnTrib = %v, esperado 333.3333333333 (tcDe10)", p.VUnTrib)
	}
	if p.QCom != 3 {
		t.Errorf("qCom = %v, esperado 3 (tcDe4)", p.QCom)
	}
	// pesoL/pesoB do volume sao tcDe3.
	vol := lerFixture(t, "nfe_400_completa.xml").Transp.Vol
	if len(vol) != 1 || vol[0].PesoL != 10 {
		t.Errorf("pesoL do volume = %v, esperado 10 (tcDe3)", vol)
	}
}

// TestLer_CombTemNomeDeTagDiferenteDoCampo cobre as duas pegadinhas do grupo
// combustivel.
//
//   - pMixGN e lido da tag qMixGN;
//   - UFCons e lido de <UFCons> no nivel do comb, enquanto ICMSCons.UFCons
//     vem de <UFcons> (minusculo) no MESMO nivel.
func TestLer_CombTemNomeDeTagDiferenteDoCampo(t *testing.T) {
	c := lerFixture(t, "nfe_400_completa.xml").Det[1].Prod.Comb

	if c.PMixGN != 27 {
		t.Errorf("pMixGN = %v, esperado 27; o campo e lido da tag qMixGN", c.PMixGN)
	}
	if c.UFCons != "SP" {
		t.Errorf("comb.UFCons = %q, esperado SP (tag <UFCons>)", c.UFCons)
	}
	// O trio completo: comb.UFCons vem de <UFCons>; ICMSCons.UFCons vem
	// primeiro de <UFcons> minusculo no nivel do comb ("MG") e e DEPOIS
	// sobrescrito pelo <UFCons> de dentro do grupo ICMSCons ("RJ"). Se a
	// ordem das duas leituras for invertida, sobra "MG" e o teste cai.
	if c.ICMSCons.UFCons != "RJ" {
		t.Errorf("ICMSCons.UFCons = %q, esperado RJ: o <UFCons> de dentro do "+
			"grupo ICMSCons sobrescreve o <UFcons> minusculo lido no nivel "+
			"do comb", c.ICMSCons.UFCons)
	}
	if c.ICMSCons.VICMSSTCons != 9 {
		t.Errorf("ICMSCons.vICMSSTCons = %v, esperado 9", c.ICMSCons.VICMSSTCons)
	}
	if c.CProdANP != 320102001 {
		t.Errorf("cProdANP = %d, esperado 320102001", c.CProdANP)
	}
	if len(c.OrigComb) != 1 || c.OrigComb[0].CUFOrig != 35 {
		t.Errorf("origComb nao foi lido: %+v", c.OrigComb)
	}
}

// TestLer_ICMSAchatadoEIndDeduzDeson cobre o grupo achatado por CST.
//
// O leitor percorre 21 nomes de grupo em sequencia e despeja o encontrado
// numa struct unica. Tambem trava o indDeduzDeson, que NAO passa por Parse:
// o ACBr compara a string crua com "1" e "0".
func TestLer_ICMSAchatadoEIndDeduzDeson(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	icms0 := n.Det[0].Imposto.ICMS
	if icms0.CST != pcn.CST00 {
		t.Errorf("item 1: CST = %v, esperado CST00", icms0.CST)
	}
	if icms0.VICMS != 180 {
		t.Errorf("item 1: vICMS = %v, esperado 180", icms0.VICMS)
	}

	// Item 2 usa ICMS61 -- grupo diferente, mesma struct.
	icms1 := n.Det[1].Imposto.ICMS
	if icms1.CST != pcn.CST61 {
		t.Errorf("item 2: CST = %v, esperado CST61", icms1.CST)
	}
	if icms1.VICMSMonoRet != 122 {
		t.Errorf("item 2: vICMSMonoRet = %v, esperado 122", icms1.VICMSMonoRet)
	}
	if icms1.IndDeduzDeson != pcn.TieNao {
		t.Errorf("item 2: indDeduzDeson = %v, esperado TieNao; o XML diz \"0\"",
			icms1.IndDeduzDeson)
	}

	// Simples Nacional: o CSOSN sai no mesmo campo achatado.
	sn := lerFixture(t, "nfce_proc_autorizada.xml").Det[0].Imposto.ICMS
	if sn.CSOSN != pcn.Csosn102 {
		t.Errorf("NFC-e: CSOSN = %v, esperado Csosn102", sn.CSOSN)
	}
}

// TestLer_RemapeamentoDeCSTDaPartilhaEDoRepasse cobre a logica mais sutil do
// lerDetImposto.
//
// ICMSPart e ICMSST nao tem CST proprio no leiaute: a partilha de ICMS sai
// com CST "10" ou "90" e o repasse de ST sai com "41" ou "60". Depois de ler
// o grupo, o ACBr REMAPEIA o enum para um membro proprio
// (cstPart10/cstPart90/cstRep41/cstRep60), que e o que permite ao gerador
// escolher a tag certa na volta.
//
// Perder o remapeamento nao quebra a leitura: a nota chega com CST 10 e, na
// regeracao, o grupo sai como ICMS10 em vez de ICMSPart.
func TestLer_RemapeamentoDeCSTDaPartilhaEDoRepasse(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	// Item 3: <ICMSPart> com <CST>10</CST>
	if got := n.Det[2].Imposto.ICMS.CST; got != pcn.CSTPart10 {
		t.Errorf("item 3: CST = %v, esperado CSTPart10. O XML diz 10 dentro de "+
			"<ICMSPart>, e o remapeamento do .pas:1023 nao aconteceu", got)
	}
	if n.Det[2].Imposto.ICMS.PBCOp != 50 {
		t.Errorf("item 3: pBCOp = %v, esperado 50", n.Det[2].Imposto.ICMS.PBCOp)
	}

	// Item 4: <ICMSST> com <CST>60</CST>
	if got := n.Det[3].Imposto.ICMS.CST; got != pcn.CSTRep60 {
		t.Errorf("item 4: CST = %v, esperado CSTRep60. O XML diz 60 dentro de "+
			"<ICMSST>, e o remapeamento do .pas:1030 nao aconteceu", got)
	}
	if n.Det[3].Imposto.ICMS.VICMSSTRet != 3.6 {
		t.Errorf("item 4: vICMSSTRet = %v, esperado 3.6", n.Det[3].Imposto.ICMS.VICMSSTRet)
	}
	// COFINSST so aparece neste item: confirma que o grupo e lido.
	if n.Det[3].Imposto.COFINSST.VCOFINS != 1.52 {
		t.Errorf("item 4: COFINSST.vCOFINS = %v, esperado 1.52",
			n.Det[3].Imposto.COFINSST.VCOFINS)
	}
}

// TestLer_VeicProdMantemCamposNumericosComoString trava a decisao de tipo do
// grupo de veiculo novo.
//
// pot, cilin, pesoL, pesoB, CMT e dist sao string no ACBr e no XSD (texto com
// pattern), e o zero a esquerda e significativo: "085" de potencia e "0950"
// de peso perderiam o zero se virassem numero.
func TestLer_VeicProdMantemCamposNumericosComoString(t *testing.T) {
	v := lerFixture(t, "nfe_400_completa.xml").Det[2].Prod.VeicProd

	casos := []struct{ campo, obtido, esperado string }{
		{"pot", v.Pot, "085"},
		{"cilin", v.Cilin, "1600"},
		{"pesoL", v.PesoL, "0950"},
		{"pesoB", v.PesoB, "1200"},
		{"CMT", v.CMT, "1500"},
		{"cCorDENATRAN", v.CCorDENATRAN, "04"},
	}
	for _, c := range casos {
		if c.obtido != c.esperado {
			t.Errorf("%s = %q, esperado %q (o zero a esquerda e significativo)",
				c.campo, c.obtido, c.esperado)
		}
	}

	// tpOp e lido da tag tpOp (minuscula no "p"), nao tpOP.
	if v.TpOP != ToVendaConcessionaria {
		t.Errorf("tpOp = %v, esperado ToVendaConcessionaria", v.TpOP)
	}
	if v.CondVeic != CvAcabado {
		t.Errorf("condVeic = %v, esperado CvAcabado", v.CondVeic)
	}
}

// TestLer_ExportaUFEmbarq cobre o quarto caso de tag != nome do campo: o
// campo e UFembarq e a tag e UFEmbarq.
func TestLer_ExportaUFEmbarq(t *testing.T) {
	e := lerFixture(t, "nfe_400_completa.xml").Exporta

	if e.UFEmbarq != "SP" {
		t.Errorf("UFEmbarq = %q, esperado SP; a tag e <UFEmbarq> e o campo do "+
			"ACBr se chama UFembarq", e.UFEmbarq)
	}
	if e.XLocEmbarq != "SANTOS" {
		t.Errorf("xLocEmbarq = %q, esperado SANTOS", e.XLocEmbarq)
	}
	if e.UFSaidaPais != "SP" || e.XLocExporta != "SANTOS" {
		t.Errorf("grupo da 3.10 nao foi lido: %+v", e)
	}
}

// TestLer_MotRedAdRemSoComReducao trava o guard do .pas.
//
// motRedAdRem so e lido quando pRedAdRem != 0. Sem o guard, todo item
// monofasico sairia com motRedAdRem no ordinal zero, que grava "1"
// (transporte coletivo) -- um motivo de reducao que a nota nao declarou.
func TestLer_MotRedAdRemSoComReducao(t *testing.T) {
	// O item 2 da fixture e monofasico, declara <motRedAdRem>9</motRedAdRem>
	// e NAO declara pRedAdRem. Com o guard, a tag nao e lida.
	icms := lerFixture(t, "nfe_400_completa.xml").Det[1].Imposto.ICMS

	if icms.PRedAdRem != 0 {
		t.Fatalf("a fixture deveria ter pRedAdRem zerado, veio %v", icms.PRedAdRem)
	}
	if icms.MotRedAdRem == MotOutros {
		t.Errorf("motRedAdRem = MotOutros: a tag <motRedAdRem>9</motRedAdRem> foi " +
			"LIDA apesar de pRedAdRem estar zerado. O guard do .pas:964 " +
			"(pRedAdRem <> 0) foi perdido")
	}
	if icms.MotRedAdRem != MotTranspColetivo {
		t.Errorf("motRedAdRem = %v, esperado o zero do enum", icms.MotRedAdRem)
	}
}

// TestLer_IPICSTGanhaDefaultPeloTipoDaNota cobre o default condicional do
// IPI: antes de ler os subgrupos, o ACBr marca o CST conforme tpNF --
// entrada vira ipi53 e saida vira ipi03.
func TestLer_IPICSTGanhaDefaultPeloTipoDaNota(t *testing.T) {
	// Fixture 4.00: tpNF = 1 (saida) e o item tem IPITrib com CST 50.
	n := lerFixture(t, "nfe_400_completa.xml")
	if n.Ide.TpNF != pcn.TnSaida {
		t.Fatalf("a fixture deveria ser de saida, veio %v", n.Ide.TpNF)
	}
	if got := n.Det[0].Imposto.IPI.CST; got != Ipi50 {
		t.Errorf("IPI.CST = %v, esperado Ipi50 (veio do IPITrib, sobrescrevendo "+
			"o default de saida)", got)
	}

	// O item 2 nao tem grupo IPI nenhum: o default nao chega a ser aplicado,
	// porque o ACBr so entra no bloco quando <IPI> existe.
	if got := n.Det[1].Imposto.IPI.CST; got != Ipi00 {
		t.Errorf("IPI.CST do item sem grupo IPI = %v, esperado o zero do enum "+
			"(Ipi00): o default por tpNF so e aplicado dentro do bloco <IPI>", got)
	}
}

// TestLer_EnvelopeProcEProtocolo cobre a leitura do nfeProc.
func TestLer_EnvelopeProcEProtocolo(t *testing.T) {
	n := lerFixture(t, "nfce_proc_autorizada.xml")

	if n.ProcNFe.CStat != 100 {
		t.Errorf("cStat = %d, esperado 100", n.ProcNFe.CStat)
	}
	if n.ProcNFe.NProt != "135260000000001" {
		t.Errorf("nProt = %q", n.ProcNFe.NProt)
	}
	if n.ProcNFe.ChDFe != "35260199999999000191650010000004561123456792" {
		t.Errorf("chNFe = %q", n.ProcNFe.ChDFe)
	}
	if n.ProcNFe.TpAmb != pcn.TaHomologacao {
		t.Errorf("tpAmb = %v, esperado TaHomologacao", n.ProcNFe.TpAmb)
	}
	// A nota dentro do envelope tem de ser lida tambem.
	if n.Ide.NNF != 456 {
		t.Errorf("nNF = %d, esperado 456", n.Ide.NNF)
	}
}

// TestLer_QRCodeRemoveSoAPrimeiraOcorrenciaDoCDATA trava um comportamento
// deterministico do ACBr que parece defeito.
//
// O StringReplace do original nao usa rfReplaceAll, entao de um CDATA duplo
// sai so o marcador externo. A fixture tem CDATA duplo justamente para isso.
func TestLer_QRCodeRemoveSoAPrimeiraOcorrenciaDoCDATA(t *testing.T) {
	n := lerFixture(t, "nfce_proc_autorizada.xml")

	qr := n.InfNFeSupl.QRCode
	if qr == "" {
		t.Fatal("qrCode nao foi lido")
	}
	if !strings.Contains(qr, "<![CDATA[") {
		t.Errorf("qrCode = %q; o marcador interno deveria PERMANECER, porque o "+
			"ACBr so remove a primeira ocorrencia", qr)
	}
	if strings.Count(qr, "<![CDATA[") != 1 {
		t.Errorf("qrCode tem %d marcadores de abertura, esperado 1",
			strings.Count(qr, "<![CDATA["))
	}
	if n.InfNFeSupl.URLChave == "" {
		t.Error("urlChave nao foi lido")
	}
}

// TestLer_ForDiaSemAtributoDiaNaoDerrubaALeitura cobre a divergencia do
// grupo cana.
//
// O ACBr faz StrToInt(Attributes.Items['dia'].Content) sem conferir se o
// atributo existe: um forDia sem @dia derruba a leitura da nota inteira. Aqui
// vale zero, como em todo atributo ausente do package.
func TestLer_ForDiaSemAtributoDiaNaoDerrubaALeitura(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	if len(n.Cana.ForDia) != 2 {
		t.Fatalf("forDia lidos: %d, esperado 2", len(n.Cana.ForDia))
	}
	if n.Cana.ForDia[0].Dia != 1 {
		t.Errorf("primeiro forDia: dia = %d, esperado 1", n.Cana.ForDia[0].Dia)
	}
	if n.Cana.ForDia[1].Dia != 0 {
		t.Errorf("segundo forDia (sem @dia): dia = %d, esperado 0", n.Cana.ForDia[1].Dia)
	}
	if n.Cana.ForDia[0].Qtde != 100 {
		t.Errorf("qtde = %v, esperado 100 (tcDe10)", n.Cana.ForDia[0].Qtde)
	}
}

// TestLer_ISSQNIndISSRetNaoEhLido trava uma omissao do ACBr.
//
// O campo existe na classe e o leitor de XML nunca o le -- so o leitor de
// .ini o preenche. Replicado: quem "consertar" isso diverge do original.
func TestLer_ISSQNIndISSRetNaoEhLido(t *testing.T) {
	// Este NAO e um teste de comportamento: e um registro executavel da
	// omissao, que confere a GRAFIA do fonte. Quebra se o leitor passar a
	// usar constante ou variavel no lugar do literal, e nao alcanca leitura
	// feita por outro caminho.
	const tag = "indISSRet"
	fonte, err := os.ReadFile("xml_reader.go")
	if err != nil {
		t.Fatalf("lendo xml_reader.go: %v", err)
	}
	if strings.Contains(string(fonte), `Find("`+tag+`")`) {
		t.Errorf("o leitor passou a ler %s; o ACBr NAO le essa tag do XML "+
			"(so do .ini). Se a leitura foi acrescentada de proposito, "+
			"registre como divergencia", tag)
	}
}

// TestLer_GruposComTagPluralEAtributo cobre dois pontos faceis de errar na
// navegacao: lacres (plural POR ocorrencia, nao um grupo) e o rotulo das
// observacoes, que vem de atributo.
func TestLer_GruposComTagPluralEAtributo(t *testing.T) {
	n := lerFixture(t, "nfe_400_completa.xml")

	vol := n.Transp.Vol
	if len(vol) != 1 {
		t.Fatalf("volumes: %d, esperado 1", len(vol))
	}
	if len(vol[0].Lacres) != 2 {
		t.Errorf("lacres: %d, esperado 2; a tag e <lacres> repetida, nao um "+
			"grupo <lacres> com filhos", len(vol[0].Lacres))
	}

	if got := n.Det[0].ObsCont.XCampo; got != "Origem" {
		t.Errorf("obsItem/obsCont/@xCampo = %q, esperado Origem", got)
	}
	if got := n.Det[0].ObsCont.XTexto; got != "IMPORTADO" {
		t.Errorf("obsItem/obsCont/xTexto = %q", got)
	}
	if len(n.InfAdic.ObsCont) != 1 || n.InfAdic.ObsCont[0].XCampo != "Pedido" {
		t.Errorf("infAdic/obsCont nao foi lido: %+v", n.InfAdic.ObsCont)
	}
}

// TestLer_NFrefAchatadoLeOsTresTipos confirma que os subgrupos mutuamente
// exclusivos de NFref caem cada um no seu campo.
func TestLer_NFrefAchatadoLeOsTresTipos(t *testing.T) {
	refs := lerFixture(t, "nfe_400_completa.xml").Ide.NFref

	if len(refs) != 3 {
		t.Fatalf("NFref lidos: %d, esperado 3", len(refs))
	}
	if refs[0].RefNF.CNPJ != "99999999000272" || refs[0].RefNF.Modelo != 1 {
		t.Errorf("refNF: %+v", refs[0].RefNF)
	}
	// Em refNFP o modelo e STRING, e o zero a esquerda importa.
	if refs[1].RefNFP.Modelo != "04" {
		t.Errorf("refNFP.mod = %q, esperado \"04\" (string, com zero a esquerda)",
			refs[1].RefNFP.Modelo)
	}
	if refs[1].RefNFP.CNPJCPF != "11111111111" {
		t.Errorf("refNFP CPF = %q", refs[1].RefNFP.CNPJCPF)
	}
	if refs[2].RefECF.Modelo != ECFModRef2B {
		t.Errorf("refECF.mod = %v, esperado ECFModRef2B", refs[2].RefECF.Modelo)
	}
}

// TestLer_TpViaTranspAceitaZeroAEsquerda cobre o caminho tcInt do .pas.
//
// O ACBr le tpViaTransp com tcInt e so entao converte para o enum, o que
// torna "01" equivalente a "1". Ler como string puro recusaria "01" e
// deixaria o campo no primeiro membro do enum.
func TestLer_TpViaTranspAceitaZeroAEsquerda(t *testing.T) {
	di := lerFixture(t, "nfe_400_completa.xml").Det[0].Prod.DI

	if len(di) != 1 {
		t.Fatalf("DI lidas: %d, esperado 1", len(di))
	}
	if di[0].TpViaTransp != TvMaritima {
		t.Errorf("tpViaTransp = %v, esperado TvMaritima; o XML diz \"01\" e o "+
			"caminho tcInt do ACBr o normaliza para \"1\"", di[0].TpViaTransp)
	}
	if di[0].Adi[0].NSeqAdi != 1 {
		t.Errorf("nSeqAdi = %d, esperado 1; a tag e nSeqAdic", di[0].Adi[0].NSeqAdi)
	}
}

// TestLer_ErrosSaoSentinelas confirma a convencao de erros do projeto: nada
// de panic, e cada falha e reconhecivel por errors.Is.
func TestLer_ErrosSaoSentinelas(t *testing.T) {
	casos := []struct {
		nome     string
		xml      string
		esperado error
	}{
		{"sem infNFe", `<NFe><outro/></NFe>`, ErrXMLIncorreto},
		{"sem Id", `<NFe><infNFe versao="4.00"/></NFe>`, ErrAtributoIDAusente},
		{"sem versao", `<NFe><infNFe Id="NFe1"/></NFe>`, ErrAtributoVersaoAusente},
		{"proc sem NFe", `<nfeProc><protNFe/></nfeProc>`, ErrNFeNaoEncontrada},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			_, err := LerXMLString(c.xml)
			if !errors.Is(err, c.esperado) {
				t.Errorf("erro = %v, esperado %v", err, c.esperado)
			}
		})
	}
}

// TestLer_EntradaTortaNaoEntraEmPanico cobre a regra de biblioteca do
// projeto: documento invalido vira error, nunca panic -- numa importacao em
// lote uma nota torta nao pode derrubar o processo.
func TestLer_EntradaTortaNaoEntraEmPanico(t *testing.T) {
	entradas := []string{
		"",
		"nao e xml",
		"<NFe>",
		"<NFe><infNFe Id=\"x\" versao=\"abc\"><ide><cUF>nao-numero</cUF></ide></infNFe></NFe>",
		"<nfeProc><NFe><infNFe Id=\"x\" versao=\"4.00\"><det><prod><qCom>x</qCom></prod></det></infNFe></NFe></nfeProc>",
		"\xef\xbb\xbf<NFe><infNFe Id=\"x\" versao=\"4.00\"/></NFe>",
	}

	for i, entrada := range entradas {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("entrada %d entrou em panico: %v", i, r)
				}
			}()
			_, _ = LerXMLString(entrada)
		}()
	}
}
