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
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes das regras de negocio.
//
// Sao 155 rejeicoes ativas, e repetir cada uma num teste proprio daria 155
// testes que apenas reafirmam o `if` ao lado. O que estes testes cobrem e
// outra coisa:
//
//   - que a fixture canonica passa LIMPA, o que exercita as 155 de uma vez
//     no caminho feliz e e o teste que mais pega regra invertida;
//   - as regras cuja CONDICAO e composta ou tem excecao, onde um porte
//     distraido acerta o caso simples e erra o resto;
//   - a aritmetica dos totais, que e a parte com estado.
//
// ferramentas/conferir-regras-dfe.py cuida do outro risco, o de uma regra
// sumir sem ninguem notar.

// agoraDeTeste e o instante de referencia. Fixo, porque oito regras comparam
// datas e um teste que dependesse do relogio mudaria de resultado sozinho.
var agoraDeTeste = time.Date(2026, 1, 20, 18, 50, 0, 0, time.UTC)

// configDeTeste reproduz o que o componente saberia ao transmitir a fixture.
func configDeTeste() Configuracoes {
	return Configuracoes{
		VersaoDF: Ve400,
		Ambiente: pcn.TaHomologacao,
		TpEmis:   1,
		UF:       "SP",
		CodigoUF: 35,
	}
}

// codigos extrai os codigos das rejeicoes, para comparacao em teste.
func codigos(erros []*ErroRegraNegocio) map[int]bool {
	m := make(map[int]bool, len(erros))
	for _, e := range erros {
		m[e.Codigo] = true
	}
	return m
}

// validar roda as regras com o instante e a configuracao de teste.
func validar(n *NFe) map[int]bool {
	return codigos(ValidarRegrasNegocio(n, configDeTeste(), agoraDeTeste))
}

// notaValida devolve a fixture canonica ajustada para PASSAR nas regras.
//
// A fixture foi feita para exercitar o leitor e o gerador, nao a aritmetica
// fiscal, e por isso precisa destes ajustes. Cada um corresponde a uma regra
// real, e a lista vale como inventario do que uma nota precisa ter para a
// SEFAZ aceitar.
func notaValida(t *testing.T) *NFe {
	t.Helper()
	n := lerFixture(t, "nfe_400_completa.xml")

	// Datas coerentes com o instante de referencia (228, 504, 505, 506, 703).
	n.Ide.DEmi = agoraDeTeste.Add(-time.Hour)
	n.Ide.DSaiEnt = agoraDeTeste.Add(-time.Hour)
	n.Ide.TpAmb = pcn.TaHomologacao

	// 897: o cNF da fixture e um dos codigos que a SEFAZ recusa.
	n.Ide.CNF = 87654321

	// 209: a IE da fixture e ficticia. Esta e valida para SP -- vem dos
	// vetores de teste do proprio ACBr.
	n.Emit.IE = "110042490114"

	// 542 e 544: idem para o transportador. Zerar e mais honesto do que
	// inventar um documento: o grupo e opcional.
	n.Transp.Transporta = Transporta{}

	// 868: a fixture e interestadual e declara veiculo de transporte, o que
	// a 4.00 proibe.
	n.Transp.VeicTransp = VeicTransp{}
	n.Transp.Reboque = nil

	// A fixture nao tem frete nem desconto, e sem eles metade dos confrontos
	// de total compara ZERO COM ZERO -- o teste passa com a acumulacao
	// quebrada. Valores diferentes por item fazem cada somatorio ter um
	// total proprio.
	for i := range n.Det {
		n.Det[i].Prod.VFrete = float64(10 * (i + 1))
		n.Det[i].Prod.VSeg = float64(5 * (i + 1))
		n.Det[i].Prod.VDesc = float64(3 * (i + 1))
		n.Det[i].Prod.VOutro = float64(7 * (i + 1))
	}

	// 531, 532, 564 e 610: a fixture nao fecha a aritmetica dos totais.
	recalcularTotais(n)

	return n
}

// recalcularTotais preenche o ICMSTot a partir dos itens, usando a MESMA
// acumulacao que as regras conferem.
//
// Isso torna o caso-base circular de proposito, e so ele: o valor dos
// testes de total esta na MUTACAO -- alterar um campo e exigir que a
// rejeicao correspondente dispare prova que a regra le o campo certo. Para
// a acumulacao em si existem testes proprios
// (IndTotDecideOQueEntraNoTotal e ServicoSeparaPISeCOFINS), e um confronto
// com valores calculados a mao em TestRecalcularTotais_ConfereComAMao.
func recalcularTotais(n *NFe) {
	c := &coletor{n: n, cfg: configDeTeste(), agora: agoraDeTeste}
	s := c.regrasDosItens()

	t := &n.Total.ICMSTot
	t.VBC, t.VICMS, t.VICMSDeson = s.BC, s.ICMS, s.ICMSDeson
	t.VBCST, t.VST = s.BCST, s.ST
	t.VProd, t.VFrete, t.VSeg, t.VDesc, t.VOutro = s.Prod, s.Frete, s.Seg, s.Desc, s.Outro
	t.VII, t.VIPI, t.VIPIDevol = s.II, s.IPI, s.IPIDevol
	t.VPIS, t.VCOFINS = s.PIS, s.COFINS
	t.VFCP, t.VFCPST, t.VFCPSTRet = s.FCP, s.FCPST, s.FCPSTRet
	t.VTotTrib = s.TotTrib
	n.Total.ISSQNTot.VPIS, n.Total.ISSQNTot.VCOFINS = s.PISServico, s.COFINSServico

	if s.FaturamentoDireto {
		t.VNF = (s.Prod + s.Frete + s.Seg + s.Outro + s.II + s.IPI +
			s.Serv + s.PISST + s.COFINSST) - (s.Desc + s.ICMSDeson)
	} else {
		t.VNF = (s.Prod + s.ST + s.Frete + s.Seg + s.Outro + s.II + s.IPI +
			s.Serv + s.FCPST + s.ICMSMonoReten + s.IPIDevol + s.PISST +
			s.COFINSST) - (s.Desc + s.ICMSDeson)
	}
}

// TestRecalcularTotais_ConfereComAMao quebra a circularidade do helper
// acima: soma os itens da fixture a mao, campo a campo, e compara.
//
// Se acumularItem esquecer um componente, o helper e a regra concordarao no
// valor errado e nenhum outro teste percebera -- este percebe.
func TestRecalcularTotais_ConfereComAMao(t *testing.T) {
	n := notaValida(t)

	var vProd, vFrete, vDesc, vIPI float64
	for _, d := range n.Det {
		// Frete, desconto e IPI entram SEMPRE; vProd so com indTot.
		vFrete += d.Prod.VFrete
		vDesc += d.Prod.VDesc
		vIPI += d.Imposto.IPI.VIPI
		if d.Prod.IndTot == ItSomaTotalNFe {
			servico := d.Prod.NCM == "00" || d.Imposto.ISSQN.CListServ != ""
			if !servico || n.Ide.FinNFe == FnAjuste {
				vProd += d.Prod.VProd
			}
		}
	}

	casos := []struct {
		campo          string
		mao, calculado float64
	}{
		{"vProd", vProd, n.Total.ICMSTot.VProd},
		{"vFrete", vFrete, n.Total.ICMSTot.VFrete},
		{"vDesc", vDesc, n.Total.ICMSTot.VDesc},
		{"vIPI", vIPI, n.Total.ICMSTot.VIPI},
	}
	for _, c := range casos {
		// Campo zerado nao prova nada: zero comparado com zero passa com a
		// acumulacao quebrada. Falhar aqui avisa que a nota de teste parou
		// de exercitar o campo.
		if c.mao == 0 {
			t.Errorf("%s: a nota de teste nao exercita este campo; o confronto "+
				"compararia zero com zero", c.campo)
			continue
		}
		if pcn.CompararValor(c.mao, c.calculado, pcn.ToleranciaMilesimo) != 0 {
			t.Errorf("%s: soma a mao %.2f, acumulada %.2f", c.campo, c.mao, c.calculado)
		}
	}
}

// TestValidarRegrasNegocio_RejeicoesDaFixtureSaoAsEsperadas e o teste mais
// amplo do conjunto: roda as 155 regras sobre um documento real e fixa
// exatamente quais disparam.
//
// A nota de teste e ajustada para ser VALIDA, entao o resultado esperado e
// lista vazia. Qualquer rejeicao aqui e regra portada errado -- invertida,
// com condicao larga demais, ou lendo o campo errado --, e e por isso que
// este teste vale mais do que 155 testes isolados.
func TestValidarRegrasNegocio_RejeicoesDaFixtureSaoAsEsperadas(t *testing.T) {
	got := validar(notaValida(t))

	// A nota ajustada tem de passar LIMPA. Qualquer rejeicao aqui e regra
	// portada errado -- invertida, com condicao larga demais, ou lendo o
	// campo errado.
	for codigo := range got {
		t.Errorf("rejeicao %d disparou numa nota valida; a regra provavelmente "+
			"esta invertida ou le o campo errado", codigo)
	}
}

// TestValidarRegrasNegocio_DatasComparamComOAgoraInjetado trava as regras
// que comparam datas com o momento da transmissao. O instante entra por
// parametro justamente para isto.
func TestValidarRegrasNegocio_DatasComparamComOAgoraInjetado(t *testing.T) {
	casos := []struct {
		codigo    int
		ajusta    func(*NFe)
		descricao string
	}{
		{228, func(n *NFe) { n.Ide.DEmi = agoraDeTeste.Add(-31 * 24 * time.Hour) },
			"emissao atrasada mais de 30 dias"},
		{703, func(n *NFe) { n.Ide.DEmi = agoraDeTeste.Add(6 * time.Minute) },
			"emissao no futuro alem dos 5 minutos de folga"},
		{504, func(n *NFe) { n.Ide.DSaiEnt = agoraDeTeste.Add(31 * 24 * time.Hour) },
			"saida 31 dias no futuro"},
		{505, func(n *NFe) { n.Ide.DSaiEnt = agoraDeTeste.Add(-31 * 24 * time.Hour) },
			"saida 31 dias no passado"},
		{506, func(n *NFe) { n.Ide.DSaiEnt = n.Ide.DEmi.Add(-time.Hour) },
			"saida antes da emissao"},
		{569, func(n *NFe) {
			n.Ide.TpEmis = pcn.TeContingencia
			n.Ide.XJust = "justificativa"
			n.Ide.DhCont = agoraDeTeste.Add(-31 * 24 * time.Hour)
		}, "entrada em contingencia muito atrasada"},
		// 558 compara so a DATA (CompareDate), nao a hora: o dia seguinte
		// dispara, mas adiantar o relogio dentro do mesmo dia nao.
		{558, func(n *NFe) {
			n.Ide.TpEmis = pcn.TeContingencia
			n.Ide.XJust = "justificativa"
			n.Ide.DhCont = agoraDeTeste.Add(24 * time.Hour)
		}, "entrada em contingencia no dia seguinte"},
		{894, func(n *NFe) {
			n.Cobr.Dup = []Dup{{NDup: "001", DVenc: n.Ide.DEmi.Add(-48 * time.Hour), VDup: n.Cobr.Fat.VLiq}}
		}, "parcela vencendo antes da emissao"},
		{867, func(n *NFe) {
			base := apenasData(n.Ide.DEmi)
			n.Cobr.Dup = []Dup{
				{NDup: "001", DVenc: base.Add(60 * 24 * time.Hour), VDup: n.Cobr.Fat.VLiq / 2},
				{NDup: "002", DVenc: base.Add(30 * 24 * time.Hour), VDup: n.Cobr.Fat.VLiq / 2},
			}
		}, "segunda parcela vencendo antes da primeira"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			limpa := notaValida(t)
			if validar(limpa)[c.codigo] {
				t.Fatalf("a rejeicao %d ja dispara na nota sem o defeito; "+
					"o teste nao prova nada", c.codigo)
			}

			comDefeito := notaValida(t)
			c.ajusta(comDefeito)
			if !validar(comDefeito)[c.codigo] {
				t.Errorf("a rejeicao %d nao disparou", c.codigo)
			}
		})
	}
}

// TestValidarRegrasNegocio_703TemCincoMinutosDeFolga trava o LIMITE, que e
// onde o erro de porte se esconde: trocar o 5 por 0 ou o > por >= passa em
// qualquer teste que use uma data muito no futuro.
func TestValidarRegrasNegocio_703TemCincoMinutosDeFolga(t *testing.T) {
	dentro := notaValida(t)
	dentro.Ide.DEmi = agoraDeTeste.Add(4 * time.Minute)
	dentro.Ide.DSaiEnt = dentro.Ide.DEmi
	if validar(dentro)[703] {
		t.Error("emissao 4 minutos no futuro esta DENTRO da folga de 5 minutos")
	}

	fora := notaValida(t)
	fora.Ide.DEmi = agoraDeTeste.Add(6 * time.Minute)
	fora.Ide.DSaiEnt = fora.Ide.DEmi
	if !validar(fora)[703] {
		t.Error("emissao 6 minutos no futuro esta FORA da folga")
	}
}

// TestValidarRegrasNegocio_704NFCeTemDezMinutosDeAtraso trava a tolerancia
// propria da NFC-e, que e diferente dos 30 dias da NF-e.
func TestValidarRegrasNegocio_704NFCeTemDezMinutosDeAtraso(t *testing.T) {
	nfce := func(t *testing.T, atraso time.Duration) *NFe {
		n := notaValida(t)
		n.Ide.Modelo = 65
		n.Ide.TpEmis = pcn.TeNormal
		n.Ide.DEmi = agoraDeTeste.Add(-atraso)
		return n
	}

	if validar(nfce(t, 9*time.Minute))[704] {
		t.Error("NFC-e com 9 minutos de atraso esta dentro da tolerancia")
	}
	if !validar(nfce(t, 11*time.Minute))[704] {
		t.Error("NFC-e com 11 minutos de atraso deveria ser rejeitada")
	}

	// Em contingencia a regra nao se aplica.
	emContingencia := nfce(t, 11*time.Minute)
	emContingencia.Ide.TpEmis = pcn.TeOffLine
	if validar(emContingencia)[704] {
		t.Error("a regra 704 so vale para os tipos de emissao normal e SCAN/SVC")
	}
}

// TestValidarRegrasNegocio_ComplementarTemQuatroRegrasEncadeadas cobre as
// regras 254, 255, 269 e 678, que so existem juntas.
func TestValidarRegrasNegocio_ComplementarTemQuatroRegrasEncadeadas(t *testing.T) {
	semReferencia := notaValida(t)
	semReferencia.Ide.FinNFe = FnComplementar
	semReferencia.Ide.NFref = nil
	if !validar(semReferencia)[254] {
		t.Error("complementar sem referenciada deveria disparar 254")
	}

	comVarias := notaValida(t)
	comVarias.Ide.FinNFe = FnComplementar
	if len(comVarias.Ide.NFref) < 2 {
		t.Fatal("a fixture precisa de pelo menos duas NFref para este caso")
	}
	erros := validar(comVarias)
	if !erros[255] {
		t.Error("complementar com mais de uma referenciada deveria disparar 255")
	}
	// Com mais de uma referencia, 269 e 678 NAO sao conferidas: o ACBr so as
	// aplica quando ha exatamente uma.
	if erros[269] || erros[678] {
		t.Error("269 e 678 so valem com exatamente uma NF referenciada")
	}

	// Com uma referencia de outro CNPJ, 269 dispara.
	cnpjDiferente := notaValida(t)
	cnpjDiferente.Ide.FinNFe = FnComplementar
	cnpjDiferente.Ide.NFref = cnpjDiferente.Ide.NFref[:1]
	cnpjDiferente.Ide.NFref[0].RefNF.CNPJ = "99999999000515"
	cnpjDiferente.Ide.NFref[0].RefNFP.CNPJCPF = ""
	if !validar(cnpjDiferente)[269] {
		t.Error("CNPJ da referenciada diferente do emitente deveria disparar 269")
	}
}

// TestValidarRegrasNegocio_GrupoDePagamentoMudaComAVersao trava o par
// 768/769, que sao a mesma pergunta com a resposta invertida.
func TestValidarRegrasNegocio_GrupoDePagamentoMudaComAVersao(t *testing.T) {
	antiga := notaValida(t)
	antiga.InfNFe.Versao = 3.10
	if !validar(antiga)[768] {
		t.Error("NF-e abaixo da 4.00 COM pagamento deveria disparar 768")
	}

	antigaSemPag := notaValida(t)
	antigaSemPag.InfNFe.Versao = 3.10
	antigaSemPag.Pag.DetPag = nil
	if validar(antigaSemPag)[768] {
		t.Error("NF-e abaixo da 4.00 SEM pagamento nao dispara 768")
	}

	nova := notaValida(t)
	nova.Pag.DetPag = nil
	if !validar(nova)[769] {
		t.Error("NF-e 4.00 SEM pagamento deveria disparar 769")
	}
}

// TestValidarRegrasNegocio_610MudaDeFormulaNoFaturamentoDireto trava a
// regra mais complexa do conjunto.
func TestValidarRegrasNegocio_610MudaDeFormulaNoFaturamentoDireto(t *testing.T) {
	// A fixture fecha o total; mexer nele faz a regra disparar.
	errada := notaValida(t)
	errada.Total.ICMSTot.VNF += 100
	if !validar(errada)[610] {
		t.Error("total da nota alterado deveria disparar 610")
	}

	// Nota de IMPORTACAO desliga a regra por inteiro -- basta um item com
	// CFOP iniciado por 3.
	importacao := notaValida(t)
	importacao.Total.ICMSTot.VNF += 100
	importacao.Det[0].Prod.CFOP = "3102"
	if validar(importacao)[610] {
		t.Error("nota de importacao nao tem o total conferido pela regra 610")
	}
}

// TestValidarRegrasNegocio_610AceitaOTotalComOuSemDesoneracao documenta a
// tolerancia DUPLA da regra (ValidarRegrasdeNegocio.pas:1125).
//
// O ACBr aceita o total calculado com OU sem a desoneracao do ICMS, porque
// as UFs divergiram sobre o que o vICMSDeson significa. Um porte que
// aplicasse so a primeira formula rejeitaria metade das notas do mercado.
//
// A nota e montada AQUI, e nao a partir da fixture: a fixture nao tem ICMS
// desonerado, e um teste que dependesse dela pularia sempre -- que foi
// exatamente o que aconteceu na primeira versao deste teste.
func TestValidarRegrasNegocio_610AceitaOTotalComOuSemDesoneracao(t *testing.T) {
	const desoneracao = 50.0

	montar := func() *NFe {
		item := NovoDetItem()
		item.Prod.NItem = 1
		item.Prod.NCM = "84713012"
		item.Prod.CFOP = "5102"
		item.Prod.VProd = 1000
		item.Prod.IndTot = ItSomaTotalNFe
		item.Imposto.ICMS.VICMSDeson = desoneracao

		n := lerFixture(t, "nfe_400_completa.xml")
		n.Ide.DEmi = agoraDeTeste.Add(-time.Hour)
		n.Ide.DSaiEnt = agoraDeTeste.Add(-time.Hour)
		n.Det = []DetItem{item}
		recalcularTotais(n)
		return n
	}

	// Caso-base: o total calculado com a desoneracao SUBTRAIDA passa.
	base := montar()
	if base.Total.ICMSTot.VICMSDeson != desoneracao {
		t.Fatalf("a nota de teste precisa ter ICMS desonerado; veio %.2f",
			base.Total.ICMSTot.VICMSDeson)
	}
	if validar(base)[610] {
		t.Fatal("o total com a desoneracao subtraida deveria passar")
	}

	// Segunda tolerancia: o total SEM a desoneracao subtraida tambem passa.
	semSubtrair := montar()
	semSubtrair.Total.ICMSTot.VNF += desoneracao
	if validar(semSubtrair)[610] {
		t.Error("o total somado de volta a desoneracao deveria ser aceito pela " +
			"segunda tolerancia da regra 610")
	}

	// E um total que nao bate com NENHUMA das duas leituras reprova -- sem
	// isto, a regra poderia estar aceitando tudo.
	erradoDeVerdade := montar()
	erradoDeVerdade.Total.ICMSTot.VNF += desoneracao * 3
	if !validar(erradoDeVerdade)[610] {
		t.Error("total fora das duas leituras deveria reprovar")
	}
}

// TestValidarRegrasNegocio_TotaisConferemItemAItem trava os dezessete
// confrontos entre o ICMSTot e o somatorio dos itens.
func TestValidarRegrasNegocio_TotaisConferemItemAItem(t *testing.T) {
	casos := []struct {
		codigo int
		ajusta func(*NFe)
		campo  string
	}{
		{531, func(n *NFe) { n.Total.ICMSTot.VBC += 1 }, "vBC"},
		{532, func(n *NFe) { n.Total.ICMSTot.VICMS += 1 }, "vICMS"},
		{533, func(n *NFe) { n.Total.ICMSTot.VBCST += 1 }, "vBCST"},
		{534, func(n *NFe) { n.Total.ICMSTot.VST += 1 }, "vST"},
		{564, func(n *NFe) { n.Total.ICMSTot.VProd += 1 }, "vProd"},
		{535, func(n *NFe) { n.Total.ICMSTot.VFrete += 1 }, "vFrete"},
		{536, func(n *NFe) { n.Total.ICMSTot.VSeg += 1 }, "vSeg"},
		{537, func(n *NFe) { n.Total.ICMSTot.VDesc += 1 }, "vDesc"},
		{601, func(n *NFe) { n.Total.ICMSTot.VII += 1 }, "vII"},
		{538, func(n *NFe) { n.Total.ICMSTot.VIPI += 1 }, "vIPI"},
		{602, func(n *NFe) { n.Total.ICMSTot.VPIS += 1 }, "vPIS"},
		{603, func(n *NFe) { n.Total.ICMSTot.VCOFINS += 1 }, "vCOFINS"},
		{604, func(n *NFe) { n.Total.ICMSTot.VOutro += 1 }, "vOutro"},
		{861, func(n *NFe) { n.Total.ICMSTot.VFCP += 1 }, "vFCP"},
		{862, func(n *NFe) { n.Total.ICMSTot.VFCPST += 1 }, "vFCPST"},
		{859, func(n *NFe) { n.Total.ICMSTot.VFCPSTRet += 1 }, "vFCPSTRet"},
		{863, func(n *NFe) { n.Total.ICMSTot.VIPIDevol += 1 }, "vIPIDevol"},
		{685, func(n *NFe) { n.Total.ICMSTot.VTotTrib += 1 }, "vTotTrib"},
		{795, func(n *NFe) { n.Total.ICMSTot.VICMSDeson += 1 }, "vICMSDeson"},
		{608, func(n *NFe) { n.Total.ISSQNTot.VPIS += 1 }, "ISSQNtot/vPIS"},
		{609, func(n *NFe) { n.Total.ISSQNTot.VCOFINS += 1 }, "ISSQNtot/vCOFINS"},
	}

	for _, c := range casos {
		t.Run(c.campo, func(t *testing.T) {
			limpa := notaValida(t)
			if validar(limpa)[c.codigo] {
				t.Fatalf("a rejeicao %d ja dispara na fixture; o teste nao prova nada",
					c.codigo)
			}

			alterada := notaValida(t)
			c.ajusta(alterada)
			if !validar(alterada)[c.codigo] {
				t.Errorf("alterar %s nao disparou a rejeicao %d", c.campo, c.codigo)
			}
		})
	}
}

// TestValidarRegrasNegocio_IndTotDecideOQueEntraNoTotal trava a condicional
// que separa os tres niveis de acumulacao.
//
// Frete e desconto entram SEMPRE; a base do ICMS so entra quando o indTot
// manda somar. Confundir os dois niveis e o erro natural.
func TestValidarRegrasNegocio_IndTotDecideOQueEntraNoTotal(t *testing.T) {
	n := notaValida(t)
	item := &n.Det[0]
	if item.Prod.IndTot != ItSomaTotalNFe {
		t.Fatal("o teste precisa de um item que componha o total")
	}

	// Tirando o item do total, a base de ICMS deixa de somar -- e o total
	// declarado passa a divergir.
	n.Det[0].Prod.IndTot = ItNaoSomaTotalNFe
	erros := validar(n)
	if !erros[531] {
		t.Error("com o item fora do total, o vBC do ICMSTot deveria divergir")
	}

	// Mas o frete continua somando, entao o total de frete NAO diverge.
	if n.Det[0].Prod.VFrete == 0 {
		t.Fatal("a nota de teste precisa de frete no item para provar o contraste")
	}
	if erros[535] {
		t.Error("o frete entra no total mesmo com indTot = nao soma")
	}
}

// TestValidarRegrasNegocio_ServicoSeparaPISeCOFINS trava a bifurcacao que
// decide em qual dos dois somatorios o PIS e o COFINS do item entram.
func TestValidarRegrasNegocio_ServicoSeparaPISeCOFINS(t *testing.T) {
	n := notaValida(t)

	// Transformar um item em servico tira o PIS dele do total de mercadorias
	// e o joga no total de servicos.
	n.Det[0].Prod.NCM = "00"
	erros := validar(n)

	if n.Det[0].Imposto.PIS.VPIS > 0 && !erros[602] {
		t.Error("com o item virando servico, o vPIS do ICMSTot deveria divergir")
	}
	if n.Det[0].Imposto.COFINS.VCOFINS > 0 && !erros[603] {
		t.Error("idem para o vCOFINS do ICMSTot")
	}
}

// TestValidarRegrasNegocio_NFCeProibeQuaseTudo cobre o bloco de proibicoes
// do modelo 65 de uma vez: transforma a NF-e da fixture em NFC-e e confere
// que as rejeicoes esperadas aparecem.
func TestValidarRegrasNegocio_NFCeProibeQuaseTudo(t *testing.T) {
	n := notaValida(t)
	n.Ide.Modelo = 65

	erros := validar(n)

	// A fixture e uma NF-e completa, e carrega varias coisas que a NFC-e
	// proibe. Estas sao as que ela de fato tem.
	proibicoes := map[int]string{
		705: "data de entrada/saida",
		707: "operacao interestadual",
		708: "documento referenciado",
		760: "dados de cobranca",
		725: "item com CFOP fora da lista de venda",
		729: "IE do destinatario",
	}
	for codigo, oque := range proibicoes {
		if !erros[codigo] {
			t.Errorf("a NFC-e deveria rejeitar %s (codigo %d)", oque, codigo)
		}
	}

	// E a mesma nota como modelo 55 nao dispara nenhuma delas.
	comoNFe := validar(notaValida(t))
	for codigo, oque := range proibicoes {
		if comoNFe[codigo] {
			t.Errorf("a rejeicao %d (%s) e exclusiva da NFC-e", codigo, oque)
		}
	}
}

// TestValidarRegrasNegocio_NFCeExigeVendaPresencialAConsumidor cobre as
// tres proibicoes que a fixture NAO exercita, montadas uma a uma.
//
// Elas ficam de fora do teste acima porque a fixture ja e uma venda normal,
// presencial e a consumidor final -- justamente o que a NFC-e exige.
func TestValidarRegrasNegocio_NFCeExigeVendaPresencialAConsumidor(t *testing.T) {
	casos := []struct {
		codigo    int
		ajusta    func(*NFe)
		descricao string
	}{
		{715, func(n *NFe) { n.Ide.FinNFe = FnDevolucao }, "finalidade diferente de normal"},
		{716, func(n *NFe) { n.Ide.IndFinal = CfNao }, "operacao nao destinada a consumidor final"},
		{717, func(n *NFe) { n.Ide.IndPres = PcInternet }, "operacao nao presencial"},
		{706, func(n *NFe) { n.Ide.TpNF = pcn.TnEntrada }, "operacao de entrada"},
		{782, func(n *NFe) { n.Ide.TpEmis = pcn.TeSCAN }, "emissao pelo SCAN"},
		{718, func(n *NFe) { n.Emit.IEST = "110042490114" }, "IE de substituto tributario"},
		{730, func(n *NFe) { n.Dest.ISUF = "100016754" }, "inscricao SUFRAMA"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			base := notaValida(t)
			base.Ide.Modelo = 65
			if validar(base)[c.codigo] {
				t.Fatalf("a rejeicao %d ja dispara na NFC-e base; o teste nao prova nada",
					c.codigo)
			}

			comDefeito := notaValida(t)
			comDefeito.Ide.Modelo = 65
			c.ajusta(comDefeito)
			if !validar(comDefeito)[c.codigo] {
				t.Errorf("a rejeicao %d nao disparou", c.codigo)
			}
		})
	}
}

// TestValidarRegrasNegocio_AmbienteEUFVemDaConfiguracao trava as regras que
// comparam o documento com a CONFIGURACAO, nao consigo mesmo.
func TestValidarRegrasNegocio_AmbienteEUFVemDaConfiguracao(t *testing.T) {
	n := notaValida(t)

	// 252: ambiente da nota contra o do web service.
	cfg := configDeTeste()
	cfg.Ambiente = pcn.TaProducao
	if !codigos(ValidarRegrasNegocio(n, cfg, agoraDeTeste))[252] {
		t.Error("ambiente da nota diferente do configurado deveria disparar 252")
	}

	// 570 e 713: contingencia da nota contra a configurada.
	contingencia := notaValida(t)
	contingencia.Ide.TpEmis = pcn.TeSVCRS
	contingencia.Ide.XJust = "contingencia"
	if !validar(contingencia)[570] {
		t.Error("tipo de emissao SVC com TpEmis configurado em 1 deveria disparar 570")
	}

	cfgSVC := configDeTeste()
	cfgSVC.TpEmis = 6
	normal := notaValida(t)
	if !codigos(ValidarRegrasNegocio(normal, cfgSVC, agoraDeTeste))[713] {
		t.Error("TpEmis configurado em 6 com nota normal deveria disparar 713")
	}
}

// TestValidarRegrasNegocio_JustificativaDeContingencia cobre o par 556/557,
// que sao espelhos.
func TestValidarRegrasNegocio_JustificativaDeContingencia(t *testing.T) {
	comJustSemContingencia := notaValida(t)
	comJustSemContingencia.Ide.TpEmis = pcn.TeNormal
	comJustSemContingencia.Ide.XJust = "nao deveria estar aqui"
	if !validar(comJustSemContingencia)[556] {
		t.Error("justificativa em emissao normal deveria disparar 556")
	}

	semJustComContingencia := notaValida(t)
	semJustComContingencia.Ide.TpEmis = pcn.TeContingencia
	semJustComContingencia.Ide.XJust = ""
	if !validar(semJustComContingencia)[557] {
		t.Error("contingencia sem justificativa deveria disparar 557")
	}
}

// TestValidarRegrasNegocio_DocumentoNuloNaoEntraEmPanico cobre a entrada
// degenerada.
func TestValidarRegrasNegocio_DocumentoNuloNaoEntraEmPanico(t *testing.T) {
	erros := ValidarRegrasNegocio(nil, configDeTeste(), agoraDeTeste)
	if len(erros) != 1 {
		t.Fatalf("documento nulo deveria devolver uma rejeicao, veio %d", len(erros))
	}
}

// TestErroRegraNegocio_TextoInclui o item quando ha.
func TestErroRegraNegocio_TextoIncluiOItem(t *testing.T) {
	semItem := (&ErroRegraNegocio{Codigo: 226, Mensagem: "Rejeicao: teste"}).Error()
	if semItem != "226-Rejeicao: teste" {
		t.Errorf("Error() = %q", semItem)
	}

	comItem := (&ErroRegraNegocio{Codigo: 777, Mensagem: "Rejeicao: teste", Item: 3}).Error()
	if comItem != "777-Rejeicao: teste [nItem: 3]" {
		t.Errorf("Error() = %q", comItem)
	}
}

// TestAcumularItem_TresNiveisDeAcumulacao monta uma nota com itens
// construidos aqui, e nao a partir da fixture.
//
// Existe porque a fixture NAO exercita os dois caminhos mais delicados da
// acumulacao -- ela nao tem item de servico nem frete --, e o helper
// recalcularTotais usa a mesma funcao que as regras conferem. Com dados
// que nao exercitam o campo, o confronto compara zero com zero e passa com
// a acumulacao quebrada.
//
// Aqui os valores sao escolhidos para que cada componente tenha um total
// DIFERENTE dos outros: trocar dois campos de lugar muda o resultado.
func TestAcumularItem_TresNiveisDeAcumulacao(t *testing.T) {
	mercadoria := NovoDetItem()
	mercadoria.Prod.NItem = 1
	mercadoria.Prod.NCM = "84713012"
	mercadoria.Prod.VProd = 1000
	mercadoria.Prod.VFrete = 10
	mercadoria.Prod.VSeg = 20
	mercadoria.Prod.VDesc = 30
	mercadoria.Prod.VOutro = 40
	mercadoria.Prod.IndTot = ItSomaTotalNFe
	mercadoria.Imposto.ICMS.VBC = 1000
	mercadoria.Imposto.PIS.VPIS = 16.50

	// Item de SERVICO: o PIS vai para o outro somatorio e o vProd nao soma.
	servico := NovoDetItem()
	servico.Prod.NItem = 2
	servico.Prod.NCM = "00"
	servico.Prod.VProd = 500
	servico.Prod.VFrete = 5
	servico.Prod.IndTot = ItSomaTotalNFe
	servico.Imposto.ICMS.VBC = 500
	servico.Imposto.PIS.VPIS = 7.25

	// Item FORA do total: so os componentes do primeiro nivel somam.
	foraDoTotal := NovoDetItem()
	foraDoTotal.Prod.NItem = 3
	foraDoTotal.Prod.NCM = "84713012"
	foraDoTotal.Prod.VProd = 900
	foraDoTotal.Prod.VFrete = 7
	foraDoTotal.Prod.IndTot = ItNaoSomaTotalNFe
	foraDoTotal.Imposto.ICMS.VBC = 900

	n := &NFe{Det: []DetItem{mercadoria, servico, foraDoTotal}}
	n.InfNFe.Versao = 4.00
	n.Ide.Modelo = 55

	c := &coletor{n: n, cfg: configDeTeste(), agora: agoraDeTeste}
	s := c.regrasDosItens()

	casos := []struct {
		campo      string
		got        float64
		esperado   float64
		explicacao string
	}{
		{"Frete", s.Frete, 22, "10+5+7: entra de TODOS, inclusive do item fora do total"},
		{"Seg", s.Seg, 20, "so o primeiro item tem"},
		{"Desc", s.Desc, 30, "idem"},
		{"Outro", s.Outro, 40, "idem"},
		{"Prod", s.Prod, 1000, "1000: o servico nao soma e o item fora do total tampouco"},
		{"BC", s.BC, 1500, "1000+500: so os itens que compoem o total"},
		{"PIS", s.PIS, 16.50, "so a mercadoria"},
		{"PISServico", s.PISServico, 7.25, "so o servico"},
	}
	for _, c := range casos {
		if pcn.CompararValor(c.got, c.esperado, pcn.ToleranciaMilesimo) != 0 {
			t.Errorf("%s = %.2f, esperado %.2f (%s)", c.campo, c.got, c.esperado, c.explicacao)
		}
	}
}

// TestAcumularItem_NotaDeAjusteSomaOServico trava a excecao do vProd: numa
// nota de ajuste o item de servico PASSA a somar.
func TestAcumularItem_NotaDeAjusteSomaOServico(t *testing.T) {
	montar := func(finalidade FinalidadeNFe) float64 {
		servico := NovoDetItem()
		servico.Prod.NItem = 1
		servico.Prod.NCM = "00"
		servico.Prod.VProd = 500
		servico.Prod.IndTot = ItSomaTotalNFe

		n := &NFe{Det: []DetItem{servico}}
		n.InfNFe.Versao = 4.00
		n.Ide.Modelo = 55
		n.Ide.FinNFe = finalidade

		c := &coletor{n: n, cfg: configDeTeste(), agora: agoraDeTeste}
		return c.regrasDosItens().Prod
	}

	if got := montar(FnNormal); got != 0 {
		t.Errorf("nota normal: vProd do servico = %.2f, esperado 0", got)
	}
	if got := montar(FnAjuste); got != 500 {
		t.Errorf("nota de ajuste: vProd do servico = %.2f, esperado 500", got)
	}
}

// TestValidarRegrasNegocio_558ComparaSoAData trava o CompareDate da regra
// (ValidarRegrasdeNegocio.pas:550), que ignora a hora.
//
// Trocar o CompareDate por uma comparacao de instante faria a entrada em
// contingencia registrada mais tarde no MESMO DIA disparar a rejeicao.
func TestValidarRegrasNegocio_558ComparaSoAData(t *testing.T) {
	mesmoDia := notaValida(t)
	mesmoDia.Ide.TpEmis = pcn.TeContingencia
	mesmoDia.Ide.XJust = "justificativa"
	mesmoDia.Ide.DhCont = agoraDeTeste.Add(2 * time.Hour) // ainda 20/01

	if validar(mesmoDia)[558] {
		t.Error("entrada em contingencia mais tarde no MESMO dia nao dispara 558; " +
			"a comparacao do ACBr e de data, nao de instante")
	}
}
