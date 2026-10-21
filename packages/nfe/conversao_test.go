// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-20.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import (
	"errors"
	"testing"
)

// enumSobTeste descreve um enum pela sua superficie generica, para que a
// invariante codigo<->ordinal possa ser verificada sem repetir as tabelas de
// codigo (o que so testaria a implementacao contra ela mesma).
//
// A quantidade de membros vem do sentinela _fimXxx do bloco const, e nao de um
// numero escrito a mao aqui. E isso que faz membro novo entrar sob teste
// sozinho: quem acrescenta uma constante empurra o sentinela, e o teste passa
// a cobrir o membro novo sem ninguem lembrar de atualizar o teste.
type enumSobTeste struct {
	nome string
	// tabela de codigos do enum, indexada pelo ordinal.
	tabela []string
	// fim e o valor do sentinela _fimXxx, ou seja, a quantidade de membros
	// DECLARADOS no bloco const. O sentinela -1, quando existe, fica de fora
	// (nao participa do iota) e tem teste proprio.
	fim    int
	codigo func(ordinal int) string
	parse  func(string) (int, error)
}

func todosOsEnums() []enumSobTeste {
	return []enumSobTeste{
		{"IndicadorPagamento", indicadorPagamentoCodigos[:], int(_fimIndicadorPagamento),
			func(o int) string { return IndicadorPagamento(o).String() },
			func(s string) (int, error) { v, e := ParseIndicadorPagamento(s); return int(v), e }},
		{"DeterminacaoBaseIcms", determinacaoBaseIcmsCodigos[:], int(_fimDeterminacaoBaseIcms),
			func(o int) string { return DeterminacaoBaseIcms(o).String() },
			func(s string) (int, error) { v, e := ParseDeterminacaoBaseIcms(s); return int(v), e }},
		{"DeterminacaoBaseIcmsST", determinacaoBaseIcmsSTCodigos[:], int(_fimDeterminacaoBaseIcmsST),
			func(o int) string { return DeterminacaoBaseIcmsST(o).String() },
			func(s string) (int, error) { v, e := ParseDeterminacaoBaseIcmsST(s); return int(v), e }},
		{"MotivoDesoneracaoICMS", motivoDesoneracaoICMSCodigos[:], int(_fimMotivoDesoneracaoICMS),
			func(o int) string { return MotivoDesoneracaoICMS(o).String() },
			func(s string) (int, error) { v, e := ParseMotivoDesoneracaoICMS(s); return int(v), e }},
		{"CstIpi", cstIpiCodigos[:], int(_fimCstIpi),
			func(o int) string { return CstIpi(o).String() },
			func(s string) (int, error) { v, e := ParseCstIpi(s); return int(v), e }},
		{"IndicadorProcesso", indicadorProcessoCodigos[:], int(_fimIndicadorProcesso),
			func(o int) string { return IndicadorProcesso(o).String() },
			func(s string) (int, error) { v, e := ParseIndicadorProcesso(s); return int(v), e }},
		{"CRT", crtCodigos[:], int(_fimCRT),
			func(o int) string { return CRT(o).String() },
			func(s string) (int, error) { v, e := ParseCRT(s); return int(v), e }},
		{"IndicadorTotal", indicadorTotalCodigos[:], int(_fimIndicadorTotal),
			func(o int) string { return IndicadorTotal(o).String() },
			func(s string) (int, error) { v, e := ParseIndicadorTotal(s); return int(v), e }},
		{"ECFModRef", ecfModRefCodigos[:], int(_fimECFModRef),
			func(o int) string { return ECFModRef(o).String() },
			func(s string) (int, error) { v, e := ParseECFModRef(s); return int(v), e }},
		{"ISSQNcSitTrib", issqncSitTribCodigos[:], int(_fimISSQNcSitTrib),
			func(o int) string { return ISSQNcSitTrib(o).String() },
			func(s string) (int, error) { v, e := ParseISSQNcSitTrib(s); return int(v), e }},
		{"DestinoOperacao", destinoOperacaoCodigos[:], int(_fimDestinoOperacao),
			func(o int) string { return DestinoOperacao(o).String() },
			func(s string) (int, error) { v, e := ParseDestinoOperacao(s); return int(v), e }},
		{"ConsumidorFinal", consumidorFinalCodigos[:], int(_fimConsumidorFinal),
			func(o int) string { return ConsumidorFinal(o).String() },
			func(s string) (int, error) { v, e := ParseConsumidorFinal(s); return int(v), e }},
		{"PresencaComprador", presencaCompradorCodigos[:], int(_fimPresencaComprador),
			func(o int) string { return PresencaComprador(o).String() },
			func(s string) (int, error) { v, e := ParsePresencaComprador(s); return int(v), e }},
		{"FormaPagamento", formaPagamentoCodigos[:], int(_fimFormaPagamento),
			func(o int) string { return FormaPagamento(o).String() },
			func(s string) (int, error) { v, e := ParseFormaPagamento(s); return int(v), e }},
		{"BandeiraCartao", bandeiraCartaoCodigos[:], int(_fimBandeiraCartao),
			func(o int) string { return BandeiraCartao(o).String() },
			func(s string) (int, error) { v, e := ParseBandeiraCartao(s); return int(v), e }},
		{"TipoViaTransp", tipoViaTranspCodigos[:], int(_fimTipoViaTransp),
			func(o int) string { return TipoViaTransp(o).String() },
			func(s string) (int, error) { v, e := ParseTipoViaTransp(s); return int(v), e }},
		{"TipoIntermedio", tipoIntermedioCodigos[:], int(_fimTipoIntermedio),
			func(o int) string { return TipoIntermedio(o).String() },
			func(s string) (int, error) { v, e := ParseTipoIntermedio(s); return int(v), e }},
		{"IndISSRet", indISSRetCodigos[:], int(_fimIndISSRet),
			func(o int) string { return IndISSRet(o).String() },
			func(s string) (int, error) { v, e := ParseIndISSRet(s); return int(v), e }},
		{"IndISS", indISSCodigos[:], int(_fimIndISS),
			func(o int) string { return IndISS(o).String() },
			func(s string) (int, error) { v, e := ParseIndISS(s); return int(v), e }},
		{"TipoAutor", tipoAutorCodigos[:], int(_fimTipoAutor),
			func(o int) string { return TipoAutor(o).String() },
			func(s string) (int, error) { v, e := ParseTipoAutor(s); return int(v), e }},
		{"IndOperacao", indOperacaoCodigos[:], int(_fimIndOperacao),
			func(o int) string { return IndOperacao(o).String() },
			func(s string) (int, error) { v, e := ParseIndOperacao(s); return int(v), e }},
		{"IndSomaPISST", indSomaPISSTCodigos[:], int(_fimIndSomaPISST),
			func(o int) string { return IndSomaPISST(o).String() },
			func(s string) (int, error) { v, e := ParseIndSomaPISST(s); return int(v), e }},
		{"IndSomaCOFINSST", indSomaCOFINSSTCodigos[:], int(_fimIndSomaCOFINSST),
			func(o int) string { return IndSomaCOFINSST(o).String() },
			func(s string) (int, error) { v, e := ParseIndSomaCOFINSST(s); return int(v), e }},
		{"FinalidadeNFe", finalidadeNFeCodigos[:], int(_fimFinalidadeNFe),
			func(o int) string { return FinalidadeNFe(o).String() },
			func(s string) (int, error) { v, e := ParseFinalidadeNFe(s); return int(v), e }},
		{"IndicadorNFe", indicadorNFeCodigos[:], int(_fimIndicadorNFe),
			func(o int) string { return IndicadorNFe(o).String() },
			func(s string) (int, error) { v, e := ParseIndicadorNFe(s); return int(v), e }},
		{"TipoOperacao", tipoOperacaoCodigos[:], int(_fimTipoOperacao),
			func(o int) string { return TipoOperacao(o).String() },
			func(s string) (int, error) { v, e := ParseTipoOperacao(s); return int(v), e }},
		{"CondicaoVeiculo", condicaoVeiculoCodigos[:], int(_fimCondicaoVeiculo),
			func(o int) string { return CondicaoVeiculo(o).String() },
			func(s string) (int, error) { v, e := ParseCondicaoVeiculo(s); return int(v), e }},
		{"TipoArma", tipoArmaCodigos[:], int(_fimTipoArma),
			func(o int) string { return TipoArma(o).String() },
			func(s string) (int, error) { v, e := ParseTipoArma(s); return int(v), e }},
		{"IndEscala", indEscalaCodigos[:], int(_fimIndEscala),
			func(o int) string { return IndEscala(o).String() },
			func(s string) (int, error) { v, e := ParseIndEscala(s); return int(v), e }},
		{"ModalidadeFrete", modalidadeFreteCodigos[:], int(_fimModalidadeFrete),
			func(o int) string { return ModalidadeFrete(o).String() },
			func(s string) (int, error) { v, e := ParseModalidadeFrete(s); return int(v), e }},
		{"Autorizacao", autorizacaoCodigos[:], int(_fimAutorizacao),
			func(o int) string { return Autorizacao(o).String() },
			func(s string) (int, error) { v, e := ParseAutorizacao(s); return int(v), e }},
		{"IndIntermed", indIntermedCodigos[:], int(_fimIndIntermed),
			func(o int) string { return IndIntermed(o).String() },
			func(s string) (int, error) { v, e := ParseIndIntermed(s); return int(v), e }},
		{"TpAto", tpAtoCodigos[:], int(_fimTpAto),
			func(o int) string { return TpAto(o).String() },
			func(s string) (int, error) { v, e := ParseTpAto(s); return int(v), e }},
		{"IndImport", indImportCodigos[:], int(_fimIndImport),
			func(o int) string { return IndImport(o).String() },
			func(s string) (int, error) { v, e := ParseIndImport(s); return int(v), e }},
		{"MotRedAdRem", motRedAdRemCodigos[:], int(_fimMotRedAdRem),
			func(o int) string { return MotRedAdRem(o).String() },
			func(s string) (int, error) { v, e := ParseMotRedAdRem(s); return int(v), e }},
		{"TpMotivo", tpMotivoCodigos[:], int(_fimTpMotivo),
			func(o int) string { return TpMotivo(o).String() },
			func(s string) (int, error) { v, e := ParseTpMotivo(s); return int(v), e }},
		{"TpGuia", tpGuiaCodigos[:], int(_fimTpGuia),
			func(o int) string { return TpGuia(o).String() },
			func(s string) (int, error) { v, e := ParseTpGuia(s); return int(v), e }},
		{"TpNFDebito", tpNFDebitoCodigos[:], int(_fimTpNFDebito),
			func(o int) string { return TpNFDebito(o).String() },
			func(s string) (int, error) { v, e := ParseTpNFDebito(s); return int(v), e }},
		{"TpNFCredito", tpNFCreditoCodigos[:], int(_fimTpNFCredito),
			func(o int) string { return TpNFCredito(o).String() },
			func(s string) (int, error) { v, e := ParseTpNFCredito(s); return int(v), e }},
		{"CSTIS", cstISCodigos[:], int(_fimCSTIS),
			func(o int) string { return CSTIS(o).String() },
			func(s string) (int, error) { v, e := ParseCSTIS(s); return int(v), e }},
		{"VersaoDF", versaoDFCodigos[:], int(_fimVersaoDF),
			func(o int) string { return VersaoDF(o).String() },
			func(s string) (int, error) { v, e := ParseVersaoDF(s); return int(v), e }},
		{"VersaoQrCode", versaoQrCodeCodigos[:], int(_fimVersaoQrCode),
			func(o int) string { return VersaoQrCode(o).String() },
			func(s string) (int, error) { v, e := ParseVersaoQrCode(s); return int(v), e }},
		{"ModeloDF", modeloDFCodigos[:], int(_fimModeloDF),
			func(o int) string { return ModeloDF(o).String() },
			func(s string) (int, error) { v, e := ParseModeloDF(s); return int(v), e }},
	}
}

// TestEnums_TabelaCobreTodoMembroDeclarado pega a tabela mais curta (ou mais
// longa) que o bloco const.
//
// E o defeito mais barato de cometer: acrescentar uma constante ao enum e
// esquecer o codigo na tabela. O membro novo passa a gravar string VAZIA em
// silencio -- build, vet e ate o teste de round-trip abaixo continuam verdes,
// porque o laco deles tambem acabaria antes do membro novo. O sentinela
// _fimXxx e o que fecha esse buraco: ele acompanha o bloco const, nao a
// tabela.
func TestEnums_TabelaCobreTodoMembroDeclarado(t *testing.T) {
	for _, e := range todosOsEnums() {
		if len(e.tabela) != e.fim {
			t.Errorf("%s: %d membros declarados no const, %d codigos na tabela",
				e.nome, e.fim, len(e.tabela))
		}
	}
}

// TestEnums_CodigoDeCadaOrdinalVoltaAoMesmoOrdinal trava a invariante que
// sustenta todo o resto do porte: a tabela de codigos nao repete codigo, e
// cada codigo gravado volta ao mesmo ordinal na leitura.
//
// O que ESTE teste pega: codigo duplicado (dois membros gravam o mesmo e a
// leitura perde um) e desalinhamento entre String() e Parse.
//
// O que ele NAO pega, por construir a expectativa a partir da propria tabela:
// tabela curta (coberto por TestEnums_TabelaCobreTodoMembroDeclarado) e dois
// codigos TROCADOS de posicao -- o round-trip fecha sobre a tabela corrompida.
// A transposicao so se pega comparando com o .pas, e e o papel de
// TestEnums_TabelaLiteralDosNaoSequenciais.
func TestEnums_CodigoDeCadaOrdinalVoltaAoMesmoOrdinal(t *testing.T) {
	for _, e := range todosOsEnums() {
		t.Run(e.nome, func(t *testing.T) {
			vistos := make(map[string]int, e.fim)
			for ordinal := 0; ordinal < e.fim; ordinal++ {
				codigo := e.codigo(ordinal)

				if anterior, repetido := vistos[codigo]; repetido {
					t.Errorf("ordinais %d e %d gravam o mesmo codigo %q",
						anterior, ordinal, codigo)
				}
				vistos[codigo] = ordinal

				volta, err := e.parse(codigo)
				if err != nil {
					t.Errorf("ordinal %d grava %q, que o Parse recusa: %v",
						ordinal, codigo, err)
					continue
				}
				if volta != ordinal {
					t.Errorf("ordinal %d grava %q, que o Parse devolve como %d",
						ordinal, codigo, volta)
				}
			}
		})
	}
}

// TestEnums_ParseRecusaCodigoInexistente confirma que o erro e reconhecivel
// por errors.Is, como manda a convencao de erros do projeto -- sem isso o
// chamador nao tem como distinguir "codigo novo do leiaute" de qualquer
// outra falha.
func TestEnums_ParseRecusaCodigoInexistente(t *testing.T) {
	for _, e := range todosOsEnums() {
		t.Run(e.nome, func(t *testing.T) {
			if _, err := e.parse("codigo-que-nao-existe"); !errors.Is(err, ErrEnumInvalido) {
				t.Errorf("esperava ErrEnumInvalido, obtive %v", err)
			}
		})
	}
}

// TestCRT_RegimeNormalGravaTresEVazioEhLidoComoRegimeNormal cobre a
// assimetria do CRT, que vem de crtRegimeNormal aparecer DUAS vezes na tabela
// do ACBr (pcnConversaoNFe.pas:776).
//
// EnumeradoToStr nao interrompe o laco e fica com a ULTIMA ocorrencia, entao
// a gravacao e "3". StrToEnumerado interrompe na PRIMEIRA, entao a leitura de
// "" devolve crtRegimeNormal sem erro. Quem "simplificasse" a tabela para uma
// ocorrencia so quebraria uma das duas pontas.
func TestCRT_RegimeNormalGravaTresEVazioEhLidoComoRegimeNormal(t *testing.T) {
	if got := CrtRegimeNormal.String(); got != "3" {
		t.Errorf("CrtRegimeNormal.String() = %q, esperado \"3\"", got)
	}

	v, err := ParseCRT("")
	if err != nil {
		t.Errorf("ParseCRT(\"\") devolveu erro %v; no ACBr a string vazia e valida", err)
	}
	if v != CrtRegimeNormal {
		t.Errorf("ParseCRT(\"\") = %v, esperado CrtRegimeNormal", v)
	}

	// Fallback de falha: StrToEnumerado devolve AEnumerados[0], que nesta
	// tabela e crtRegimeNormal -- nao o ordinal zero (crtSimplesNacional).
	if v, _ := ParseCRT("7"); v != CrtRegimeNormal {
		t.Errorf("ParseCRT(invalido) = %v, esperado CrtRegimeNormal", v)
	}
}

// TestIndicadorTotal_ValorZeroCompoeOTotal trava o default perigoso: o ACBr
// declara itSomaTotalNFe primeiro, entao um item recem-criado grava "1".
// Inverter a ordem dos membros faria todo item nascer fora do total da nota.
func TestIndicadorTotal_ValorZeroCompoeOTotal(t *testing.T) {
	var zeroValue IndicadorTotal
	if zeroValue != ItSomaTotalNFe {
		t.Fatalf("valor zero de IndicadorTotal = %v, esperado ItSomaTotalNFe", zeroValue)
	}
	if got := zeroValue.String(); got != "1" {
		t.Errorf("valor zero grava %q, esperado \"1\"", got)
	}
	if got := ItNaoSomaTotalNFe.String(); got != "0" {
		t.Errorf("ItNaoSomaTotalNFe grava %q, esperado \"0\"", got)
	}

	// Fallback de falha e o primeiro da tabela de mapeamento do ACBr.
	if v, _ := ParseIndicadorTotal("7"); v != ItNaoSomaTotalNFe {
		t.Errorf("ParseIndicadorTotal(invalido) = %v, esperado ItNaoSomaTotalNFe", v)
	}
}

// TestEnums_CodigosForaDeOrdemNumerica trava os enums em que a ordem de
// declaracao do ACBr nao acompanha a ordem dos codigos. Sao os candidatos
// naturais a "arrumacao" por quem olha so a lista de constantes -- e arrumar
// qualquer um deles troca o significado do campo no XML.
func TestEnums_CodigosForaDeOrdemNumerica(t *testing.T) {
	casos := []struct {
		nome     string
		obtido   string
		esperado string
	}{
		// TipoAutor: taOutros ("9") e declarado ANTES de taEmpresaSucessora
		// ("8") -- pcnConversaoNFe.pas:1097.
		{"TaOutros", TaOutros.String(), "9"},
		{"TaEmpresaSucessora", TaEmpresaSucessora.String(), "8"},

		// CstIpi: a declaracao comeca em 00, 49, 50, 99 e so entao retoma
		// em 01.
		{"Ipi49 (ordinal 1)", Ipi49.String(), "49"},
		{"Ipi99 (ordinal 3)", Ipi99.String(), "99"},
		{"Ipi01 (ordinal 4)", Ipi01.String(), "01"},

		// MotivoDesoneracaoICMS: os codigos saltam de 12 para 16 e 90.
		{"MdiOrgaoFomento", MdiOrgaoFomento.String(), "12"},
		{"MdiOlimpiadaRio2016", MdiOlimpiadaRio2016.String(), "16"},
		{"MdiSolicitadoFisco", MdiSolicitadoFisco.String(), "90"},

		// TipoOperacao: toOutros vale "0" e e o ULTIMO declarado.
		{"ToOutros", ToOutros.String(), "0"},

		// FormaPagamento: a sequencia salta de 19 para 90/98/99 e so entao
		// volta para 20.
		{"FpProgramaFidelidade", FpProgramaFidelidade.String(), "19"},
		{"FpSemPagamento", FpSemPagamento.String(), "90"},
		{"FpOutro", FpOutro.String(), "99"},
		{"FpPagamentoInstantaneoEstatico", FpPagamentoInstantaneoEstatico.String(), "20"},
		{"FpPagamentoPosterior", FpPagamentoPosterior.String(), "91"},

		// IndISSRet: o "nao" vale "2", nao "0" como no indicador comum.
		{"IirNao", IirNao.String(), "2"},

		// CSTIS: tres digitos, zeros a esquerda significativos.
		{"Cstis000", Cstis000.String(), "000"},
	}

	for _, c := range casos {
		if c.obtido != c.esperado {
			t.Errorf("%s grava %q, esperado %q", c.nome, c.obtido, c.esperado)
		}
	}
}

// TestEnums_TabelaLiteralDosNaoSequenciais confere a tabela inteira dos enums
// cuja ordem de declaracao NAO acompanha a ordem numerica dos codigos.
//
// A lista abaixo e uma transcricao independente de pcnConversaoNFe.pas, e nao
// uma copia da tabela do codigo. E de proposito: so uma segunda transcricao
// pega dois codigos TROCADOS de posicao, que e o unico defeito desta familia
// que sobrevive ao round-trip (ele fecha certinho sobre a tabela corrompida).
//
// Nos enums estritamente sequenciais ("1","2","3"...) a transposicao e visivel
// a olho nu e nao justifica a duplicacao.
func TestEnums_TabelaLiteralDosNaoSequenciais(t *testing.T) {
	casos := []struct {
		nome   string
		fonte  string
		codigo func(int) string
		// esperado transcrito do .pas, na ordem de DECLARACAO dos membros
		esperado []string
	}{
		{"CstIpi", "CSTIPIToStr, pcnConversaoNFe.pas:716",
			func(o int) string { return CstIpi(o).String() },
			[]string{"00", "49", "50", "99", "01", "02", "03", "04", "05",
				"51", "52", "53", "54", "55"}},

		{"FormaPagamento", "FormaPagamentoToStr, pcnConversaoNFe.pas:872",
			func(o int) string { return FormaPagamento(o).String() },
			[]string{"01", "02", "03", "04", "05", "10", "11", "12", "13",
				"14", "15", "16", "17", "18", "19", "90", "98", "99", "20",
				"21", "22", "91", "23", "24"}},

		{"BandeiraCartao", "BandeiraCartaoToStr, pcnConversaoNFe.pas:973",
			func(o int) string { return BandeiraCartao(o).String() },
			[]string{"01", "02", "03", "04", "05", "06", "07", "08", "09",
				"10", "11", "12", "13", "14", "15", "16", "17", "18", "19",
				"20", "21", "22", "23", "24", "25", "26", "27", "99"}},

		{"MotivoDesoneracaoICMS", "motDesICMSToStr, pcnConversaoNFe.pas:644",
			func(o int) string { return MotivoDesoneracaoICMS(o).String() },
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11",
				"12", "16", "90"}},

		{"TipoAutor", "TipoAutorToStr, pcnConversaoNFe.pas:1095",
			func(o int) string { return TipoAutor(o).String() },
			[]string{"1", "2", "3", "5", "6", "9", "8"}},

		{"TipoOperacao", "tpOPToStr, pcnConversaoNFe.pas:1376",
			func(o int) string { return TipoOperacao(o).String() },
			[]string{"1", "2", "3", "0"}},

		{"TipoViaTransp", "TipoViaTranspToStr, pcnConversaoNFe.pas:1003",
			func(o int) string { return TipoViaTransp(o).String() },
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11",
				"12", "13"}},

		{"IndicadorTotal", "indTotToStr, pcnConversaoNFe.pas:797",
			func(o int) string { return IndicadorTotal(o).String() },
			[]string{"1", "0"}},

		{"CRT", "CRTToStr, pcnConversaoNFe.pas:774 (ultima ocorrencia vence)",
			func(o int) string { return CRT(o).String() },
			[]string{"1", "2", "3", "4"}},

		{"TpAto", "tpAtoToStr, pcnConversaoNFe.pas:1672",
			func(o int) string { return TpAto(o).String() },
			[]string{"", "08", "10", "12", "14", "15"}},

		{"ModalidadeFrete", "modFreteToStr, pcnConversaoNFe.pas:1583",
			func(o int) string { return ModalidadeFrete(o).String() },
			[]string{"0", "1", "2", "3", "4", "9"}},

		{"IndicadorProcesso", "indProcToStr, pcnConversaoNFe.pas:749",
			func(o int) string { return IndicadorProcesso(o).String() },
			[]string{"0", "1", "2", "3", "4", "9"}},

		{"PresencaComprador", "PresencaCompradorToStr, pcnConversaoNFe.pas:858",
			func(o int) string { return PresencaComprador(o).String() },
			[]string{"0", "1", "2", "3", "4", "5", "9"}},

		{"CSTIS", "TCSTISArrayStrings, pcnConversaoNFe.pas:443",
			func(o int) string { return CSTIS(o).String() },
			[]string{"", "000", "100", "200", "300", "400", "500", "600"}},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			for ordinal, esperado := range c.esperado {
				if got := c.codigo(ordinal); got != esperado {
					t.Errorf("ordinal %d grava %q, mas %s diz %q",
						ordinal, got, c.fonte, esperado)
				}
			}
			// um ordinal alem do ultimo transcrito precisa gravar vazio,
			// senao a tabela do codigo tem membro que o .pas nao tem
			if got := c.codigo(len(c.esperado)); got != "" {
				t.Errorf("ordinal %d grava %q, mas %s so vai ate o ordinal %d",
					len(c.esperado), got, c.fonte, len(c.esperado)-1)
			}
		})
	}
}

// TestFormaPagamento_MembrosExclusivosDaUnitPCN e a prova executavel de qual
// unit de conversao vale.
//
// fpPagInstantaneoPIXAutomatico ("23") e fpTEFBookTransfer ("24") existem
// APENAS em PCNNFe/pcnConversaoNFe.pas:156. Base/ACBrNFe.Conversao.pas para em
// fpPagamentoPosterior. Se alguem reportar este package a unit Base/, estes
// dois codigos somem e o teste acusa.
func TestFormaPagamento_MembrosExclusivosDaUnitPCN(t *testing.T) {
	if got := FpPagInstantaneoPIXAutomatico.String(); got != "23" {
		t.Errorf("FpPagInstantaneoPIXAutomatico = %q, esperado \"23\"", got)
	}
	if got := FpTEFBookTransfer.String(); got != "24" {
		t.Errorf("FpTEFBookTransfer = %q, esperado \"24\"", got)
	}
}

// TestFinalidadeNFe_CreditoEDebitoExistem cobre o outro par exclusivo da
// unit PCN. Sao os membros que Base/ACBrNFe.XmlWriter.pas:546 usa, e por isso
// a prova de que o gerador NAO compila contra Base/ACBrNFe.Conversao.pas.
func TestFinalidadeNFe_CreditoEDebitoExistem(t *testing.T) {
	if got := FnCredito.String(); got != "5" {
		t.Errorf("FnCredito = %q, esperado \"5\"", got)
	}
	if got := FnDebito.String(); got != "6" {
		t.Errorf("FnDebito = %q, esperado \"6\"", got)
	}
}

// TestEnums_SentinelaNegativoGravaVazio cobre os tres enums em que o ACBr usa
// TXxx(-1) como "nao informado". O membro nao pertence a faixa declarada, e
// tanto a gravacao quanto a leitura da string vazia precisam reconhece-lo.
func TestEnums_SentinelaNegativoGravaVazio(t *testing.T) {
	if got := DbisNenhum.String(); got != "" {
		t.Errorf("DbisNenhum grava %q, esperado vazio", got)
	}
	if v, err := ParseDeterminacaoBaseIcmsST(""); err != nil || v != DbisNenhum {
		t.Errorf("ParseDeterminacaoBaseIcmsST(\"\") = (%v, %v), esperado (DbisNenhum, nil)", v, err)
	}

	if got := MdiNenhum.String(); got != "" {
		t.Errorf("MdiNenhum grava %q, esperado vazio", got)
	}
	if v, err := ParseMotivoDesoneracaoICMS(""); err != nil || v != MdiNenhum {
		t.Errorf("ParseMotivoDesoneracaoICMS(\"\") = (%v, %v), esperado (MdiNenhum, nil)", v, err)
	}

	if got := BcNenhum.String(); got != "" {
		t.Errorf("BcNenhum grava %q, esperado vazio", got)
	}
	if v, err := ParseBandeiraCartao(""); err != nil || v != BcNenhum {
		t.Errorf("ParseBandeiraCartao(\"\") = (%v, %v), esperado (BcNenhum, nil)", v, err)
	}
}

// TestParse_SensibilidadeACaixaSegueAFuncaoDeOrigem trava a diferenca entre
// os dois conversores do ACBr: StrToEnumerado compara com AnsiSameText
// (ignora caixa) e os lacos proprios de TtpGuia, TtpNFDebito, TtpNFCredito e
// TCSTIS comparam com "=" (nao ignoram).
//
// Uniformizar os dois lados parece limpeza e muda o que o leitor aceita.
func TestParse_SensibilidadeACaixaSegueAFuncaoDeOrigem(t *testing.T) {
	// StrToEnumerado: caixa ignorada.
	if v, err := ParseIndEscala("s"); err != nil || v != IeRelevante {
		t.Errorf("ParseIndEscala(%q) = (%v, %v), esperado (IeRelevante, nil)", "s", v, err)
	}
	if v, err := ParseISSQNcSitTrib("n"); err != nil || v != ISSQNcSitTribNormal {
		t.Errorf("ParseISSQNcSitTrib(%q) = (%v, %v), esperado (ISSQNcSitTribNormal, nil)", "n", v, err)
	}
	if v, err := ParseECFModRef("2b"); err != nil || v != ECFModRef2B {
		t.Errorf("ParseECFModRef(%q) = (%v, %v), esperado (ECFModRef2B, nil)", "2b", v, err)
	}

	// Laco proprio: caixa respeitada. Os codigos destes enums sao
	// numericos, entao o efeito so aparece em entrada torta -- mas a
	// assimetria e do original e esta registrada aqui de proposito.
	if _, err := ParseCSTIS("00O"); err == nil {
		t.Errorf("ParseCSTIS(%q) deveria falhar: a comparacao do ACBr e exata", "00O")
	}
}
