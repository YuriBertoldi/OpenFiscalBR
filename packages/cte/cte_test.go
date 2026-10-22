// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package cte

import (
	"errors"
	"testing"
)

// Testes dos enums do CT-e.
//
// O que eles travam nao e "o enum existe": e o CODIGO que sai no XML e o
// ORDINAL que o decide. Um membro inserido no meio de qualquer lista
// desloca todos os seguintes, e nada -- nem build, nem vet, nem round-trip
// ingenuo -- acusa: os codigos continuam saindo, trocados.

// TestEnums_CodigoDoLeiaute confere o codigo exato de cada enum.
//
// Os casos escolhidos sao as BORDAS: primeiro membro, ultimo membro, e
// todo lugar onde a tabela faz algo que a contagem nao prevê.
func TestEnums_CodigoDoLeiaute(t *testing.T) {
	casos := []struct {
		nome     string
		got      string
		esperado string
	}{
		{"MoCTe", MoCTe.String(), "57"},
		{"MoGTVe", MoGTVe.String(), "64"},
		{"MoCTeOS", MoCTeOS.String(), "67"},
		{"MoCTeSimp", MoCTeSimp.String(), "57"},

		{"TcNormal", TcNormal.String(), "0"},
		{"TcSubstCTeSimpl", TcSubstCTeSimpl.String(), "6"},

		{"TsMultimodal", TsMultimodal.String(), "4"},
		{"TsTranspPessoas", TsTranspPessoas.String(), "6"},
		{"TsGTV", TsGTV.String(), "9"},

		{"RtSim", RtSim.String(), "0"},
		{"RtNao", RtNao.String(), "1"},
		{"LtNao", LtNao.String(), "0"},
		{"LtSim", LtSim.String(), "1"},

		{"TdDeclaracao", TdDeclaracao.String(), "00"},
		{"TdNFCe", TdNFCe.String(), "65"},
		{"TdDocOutros", TdDocOutros.String(), "99"},

		{"DaCTRC", DaCTRC.String(), "00"},
		{"DaBL", DaBL.String(), "13"},
		{"DaAnteriorOutros", DaAnteriorOutros.String(), "99"},

		{"TdNaoInformado", TdNaoInformado.String(), "-1"},
		{"ThHoraNaoInformado", ThHoraNaoInformado.String(), "-1"},

		{"DrNorte", DrNorte.String(), "N"},
		{"DrOeste", DrOeste.String(), "O"},
		{"TpProprio", TpProprio.String(), "P"},
		{"TpTerceiro", TpTerceiro.String(), "T"},

		{"TfNenhum", TfNenhum.String(), ""},
		{"TfContinuo", TfContinuo.String(), "2"},
		{"CrtNenhum", CrtNenhum.String(), ""},
		{"CrtSimplesNacionalMEI", CrtSimplesNacionalMEI.String(), "4"},
		{"TcdNenhum", TcdNenhum.String(), ""},
		{"TcdOleoduto", TcdOleoduto.String(), "3"},
		{"TcNenhum", TcNenhum.String(), ""},
		{"TcCapacidadeSaida", TcCapacidadeSaida.String(), "3"},

		{"Ve200", Ve200.String(), "2.00"},
		{"Ve400", Ve400.String(), "4.00"},
	}
	for _, c := range casos {
		if c.got != c.esperado {
			t.Errorf("%s = %q, esperado %q", c.nome, c.got, c.esperado)
		}
	}
}

// TestTipoServico_PulaOCodigoCINCO trava a descontinuidade da tabela.
//
// O leiaute vai de "4" para "6" -- nao existe servico de codigo "5".
// Preencher a lacuna deslocaria TsTranspPessoas, TsTranspValores,
// TsExcessoBagagem e TsGTV de uma posicao cada, e todos sairiam com o
// codigo do vizinho.
func TestTipoServico_PulaOCodigoCINCO(t *testing.T) {
	if _, err := ParseTipoServico("5"); !errors.Is(err, ErrEnumInvalido) {
		t.Error(`o codigo "5" nao existe em TipoServico`)
	}

	// E o membro seguinte ao multimodal e o "6", nao o "5".
	if TsTranspPessoas.String() != "6" {
		t.Errorf("TsTranspPessoas = %q, esperado 6", TsTranspPessoas.String())
	}
	if int(TsTranspPessoas)-int(TsMultimodal) != 1 {
		t.Error("os dois membros continuam CONTIGUOS no ordinal, so o codigo pula")
	}
}

// TestRetira_SIM_EhZero trava um ordinal contraintuitivo.
//
// O campo `retira` responde "o recebedor retira?" com 0=Sim e 1=Nao. Quem
// supuser a ordem natural (Nao primeiro) inverte a resposta de toda
// prestacao sem erro de compilacao -- os dois valem "0" e "1".
func TestRetira_SIM_EhZero(t *testing.T) {
	if int(RtSim) != 0 || int(RtNao) != 1 {
		t.Errorf("RtSim=%d RtNao=%d; o leiaute tem Sim=0 e Nao=1",
			int(RtSim), int(RtNao))
	}
	// A Lotacao, que convive no mesmo leiaute, e o contrario.
	if int(LtNao) != 0 || int(LtSim) != 1 {
		t.Errorf("LtNao=%d LtSim=%d; aqui o Nao vem primeiro",
			int(LtNao), int(LtSim))
	}
}

// TestModeloCTe_SimplificadoCompartilhaO57 trava a assimetria do modelo.
//
// MoCTe e MoCTeSimp saem como "57". A leitura NAO tem como distinguir os
// dois, e devolve sempre MoCTe -- a informacao que os separa nao esta no
// codigo. Fazer o Parse devolver MoCTeSimp para "57" quebraria todo
// documento CT-e normal.
func TestModeloCTe_SimplificadoCompartilhaO57(t *testing.T) {
	if MoCTe.String() != MoCTeSimp.String() {
		t.Fatal("os dois modelos saem com o MESMO codigo")
	}

	v, err := ParseModeloCTe("57")
	if err != nil {
		t.Fatal(err)
	}
	if v != MoCTe {
		t.Errorf(`Parse("57") = %d, esperado MoCTe -- a leitura nao distingue`, v)
	}

	// Os outros dois fecham a ida e volta.
	for _, m := range []ModeloCTe{MoGTVe, MoCTeOS} {
		v, err := ParseModeloCTe(m.String())
		if err != nil || v != m {
			t.Errorf("%q: %v %v", m.String(), v, err)
		}
	}
}

// TestTpInfManu_DuasTabelasQueNaoSeFalam trava o defeito do original.
//
// O ACBr grava com a tabela V2 ("1") e le com a do leiaute ("01"): um
// documento gerado por ele e relido por ele PERDE o campo. E seis membros
// -- as Packing Instructions da IATA -- nao tem codigo na tabela V2, logo
// nao tem como ser emitidos por ela.
//
// As duas formas existem no mundo, e as duas ficam expostas. Este teste
// garante que nenhuma se perca e que a V2 nao finja cobrir o que nao cobre.
func TestTpInfManu_DuasTabelasQueNaoSeFalam(t *testing.T) {
	// A mesma informacao, dois codigos de largura diferente.
	if ImCEEAV.String() != "01" || ImCEEAV.StringV2() != "1" {
		t.Errorf("ImCEEAV: leiaute=%q legado=%q", ImCEEAV.String(), ImCEEAV.StringV2())
	}
	if ImOUTRO.String() != "99" || ImOUTRO.StringV2() != "99" {
		t.Errorf("ImOUTRO: leiaute=%q legado=%q", ImOUTRO.String(), ImOUTRO.StringV2())
	}

	// Os seis membros que a tabela legada NAO alcanca.
	semV2 := []struct {
		nome string
		v    TpInfManu
		cod  string
	}{
		{"ImPI965", ImPI965, "10"}, {"ImPI966", ImPI966, "11"},
		{"ImPI967", ImPI967, "12"}, {"ImPI968", ImPI968, "13"},
		{"ImPI969", ImPI969, "14"}, {"ImPI970", ImPI970, "15"},
	}
	for _, c := range semV2 {
		if c.v.String() != c.cod {
			t.Errorf("%s: codigo do leiaute = %q, esperado %q",
				c.nome, c.v.String(), c.cod)
		}
		if c.v.StringV2() != "" {
			t.Errorf("%s: a tabela legada NAO o alcanca; devolveu %q",
				c.nome, c.v.StringV2())
		}
	}

	// Ida e volta dentro de CADA tabela.
	for _, c := range semV2 {
		if v, err := ParseTpInfManu(c.cod); err != nil || v != c.v {
			t.Errorf("Parse(%q) = %v %v", c.cod, v, err)
		}
	}
	if v, err := ParseTpInfManuV2("9"); err != nil || v != ImAAGA {
		t.Errorf(`ParseV2("9") = %v %v, esperado ImAAGA`, v, err)
	}

	// E o cruzamento FALHA, que e o ponto: "1" nao e codigo do leiaute e
	// "01" nao e codigo da tabela legada.
	if _, err := ParseTpInfManu("1"); !errors.Is(err, ErrEnumInvalido) {
		t.Error(`"1" nao e codigo do leiaute -- la os codigos tem dois digitos`)
	}
	if _, err := ParseTpInfManuV2("01"); !errors.Is(err, ErrEnumInvalido) {
		t.Error(`"01" nao e codigo da tabela legada`)
	}

	// String vazia nao pode casar com as posicoes vazias da tabela legada.
	if _, err := ParseTpInfManuV2(""); !errors.Is(err, ErrEnumInvalido) {
		t.Error("string vazia casaria com ImPI965 se a busca nao pulasse os vazios")
	}
}

// TestCRT_NAO_EhOMesmoDoPackageNFe trava um deslocamento fácil de herdar.
//
// No packages/nfe o primeiro membro de CRT e CrtSimplesNacional, codigo
// "1". Aqui ha um CrtNenhum de codigo VAZIO na frente, e todos os demais
// ficam deslocados em um. Copiar o enum de la declararia o regime errado em
// toda prestacao.
func TestCRT_NAO_EhOMesmoDoPackageNFe(t *testing.T) {
	if int(CrtNenhum) != 0 || CrtNenhum.String() != "" {
		t.Errorf("CrtNenhum e o ordinal 0 e sai vazio; veio %d/%q",
			int(CrtNenhum), CrtNenhum.String())
	}
	if int(CrtSimplesNacional) != 1 || CrtSimplesNacional.String() != "1" {
		t.Errorf("CrtSimplesNacional e o ordinal 1 e sai \"1\"; veio %d/%q",
			int(CrtSimplesNacional), CrtSimplesNacional.String())
	}
	if int(CrtRegimeNormal) != 3 {
		t.Errorf("CrtRegimeNormal e o ordinal 3 aqui, veio %d", int(CrtRegimeNormal))
	}
}

// TestEnums_IdaEVolta percorre cada enum inteiro.
//
// Pega o caso em que String() e Parse usam tabelas diferentes -- que foi
// exatamente o defeito encontrado no TpInfManu do original.
func TestEnums_IdaEVolta(t *testing.T) {
	t.Run("TipoCTe", func(t *testing.T) {
		for i := TcNormal; i < _fimTipoCTe; i++ {
			if v, err := ParseTipoCTe(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("TipoServico", func(t *testing.T) {
		for i := TsNormal; i < _fimTipoServico; i++ {
			if v, err := ParseTipoServico(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("TipoDocumentoAnterior", func(t *testing.T) {
		for i := DaCTRC; i < _fimTipoDocumentoAnterior; i++ {
			if v, err := ParseTipoDocumentoAnterior(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("TpInfManu", func(t *testing.T) {
		for i := ImCEEAV; i < _fimTpInfManu; i++ {
			if v, err := ParseTpInfManu(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("CRT", func(t *testing.T) {
		for i := CrtNenhum; i < _fimCRT; i++ {
			if v, err := ParseCRT(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("ClassDuto", func(t *testing.T) {
		for i := TcdNenhum; i < _fimClassDuto; i++ {
			if v, err := ParseClassDuto(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
	t.Run("TipoContratacao", func(t *testing.T) {
		for i := TcNenhum; i < _fimTipoContratacao; i++ {
			if v, err := ParseTipoContratacao(i.String()); err != nil || v != i {
				t.Errorf("%d (%q): %v %v", i, i.String(), v, err)
			}
		}
	})
}

// TestEnums_TabelaAlinhadaComOEnum trava a contagem de cada tabela contra a
// sentinela do enum.
//
// Membro acrescentado sem a entrada correspondente nao quebra build nem
// vet: o String() cairia no ramo de fora-de-faixa e o Parse nunca
// devolveria o membro novo -- os dois em silencio.
func TestEnums_TabelaAlinhadaComOEnum(t *testing.T) {
	casos := []struct {
		nome    string
		tabela  int
		fimEnum int
	}{
		{"ModeloCTe", len(modeloCTeCodigos), int(_fimModeloCTe)},
		{"TipoCTe", len(tipoCTeCodigos), int(_fimTipoCTe)},
		{"TipoCTe/descricoes", len(tipoCTeDescricoes), int(_fimTipoCTe)},
		{"TipoServico", len(tipoServicoCodigos), int(_fimTipoServico)},
		{"TipoServico/descricoes", len(tipoServicoDescricoes), int(_fimTipoServico)},
		{"FormaPagamento", len(formaPagamentoCodigos), int(_fimFormaPagamento)},
		{"FormaPagamento/descricoes", len(formaPagamentoDescricoes), int(_fimFormaPagamento)},
		{"Tomador", len(tomadorCodigos), int(_fimTomador)},
		{"Tomador/descricoes", len(tomadorDescricoes), int(_fimTomador)},
		{"Retira", len(retiraCodigos), int(_fimRetira)},
		{"Lotacao", len(lotacaoCodigos), int(_fimLotacao)},
		{"ModeloNF", len(modeloNFCodigos), int(_fimModeloNF)},
		{"TipoDocumento", len(tipoDocumentoCodigos), int(_fimTipoDocumento)},
		{"TipoDocumentoAnterior", len(tipoDocumentoAnteriorCodigos), int(_fimTipoDocumentoAnterior)},
		{"IndNegociavel", len(indNegociavelCodigos), int(_fimIndNegociavel)},
		{"TipoDataPeriodo", len(tipoDataPeriodoCodigos), int(_fimTipoDataPeriodo)},
		{"TipoHorarioIntervalo", len(tipoHorarioIntervaloCodigos), int(_fimTipoHorarioIntervalo)},
		{"RspPagPedagio", len(rspPagPedagioCodigos), int(_fimRspPagPedagio)},
		{"TipoDispositivo", len(tipoDispositivoCodigos), int(_fimTipoDispositivo)},
		{"TipoPropriedade", len(tipoPropriedadeCodigos), int(_fimTipoPropriedade)},
		{"TipoVeiculo", len(tipoVeiculoCodigos), int(_fimTipoVeiculo)},
		{"TipoFretamento", len(tipoFretamentoCodigos), int(_fimTipoFretamento)},
		{"TipoTrafego", len(tipoTrafegoCodigos), int(_fimTipoTrafego)},
		{"TrafegoMutuo", len(trafegoMutuoCodigos), int(_fimTrafegoMutuo)},
		{"TipoDirecao", len(tipoDirecaoCodigos), int(_fimTipoDirecao)},
		{"ClassDuto", len(classDutoCodigos), int(_fimClassDuto)},
		{"TipoContratacao", len(tipoContratacaoCodigos), int(_fimTipoContratacao)},
		{"TpInfManu", len(tpInfManuCodigos), int(_fimTpInfManu)},
		{"TpInfManu/V2", len(tpInfManuCodigosV2), int(_fimTpInfManu)},
		{"UniMed", len(uniMedCodigos), int(_fimUniMed)},
		{"RspSeg", len(rspSegCodigos), int(_fimRspSeg)},
		{"RspSeg/descricoes", len(rspSegDescricoes), int(_fimRspSeg)},
		{"Especie", len(especieCodigos), int(_fimEspecie)},
		{"TpNumerario", len(tpNumerarioCodigos), int(_fimTpNumerario)},
		{"TpComp", len(tpCompCodigos), int(_fimTpComp)},
		{"TpPrest", len(tpPrestCodigos), int(_fimTpPrest)},
		{"CRT", len(crtCodigos), int(_fimCRT)},
		{"TpMotivo", len(tpMotivoCodigos), int(_fimTpMotivo)},
		{"TpMotivo/descricoes", len(tpMotivoDescricoes), int(_fimTpMotivo)},
		{"VersaoCTe", len(versaoCTeCodigos), int(_fimVersaoCTe)},
		{"VersaoCTe/valores", len(versaoCTeValores), int(_fimVersaoCTe)},
		{"LayOutCTe", len(layOutCTeServicos), int(_fimLayOutCTe)},
		{"SchemaCTe", len(schemaCTeNomes), int(_fimSchemaCTe)},
	}
	for _, c := range casos {
		if c.tabela != c.fimEnum {
			t.Errorf("%s: a tabela tem %d entradas e o enum %d membros",
				c.nome, c.tabela, c.fimEnum)
		}
	}
}

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

// TestSchemaCTe_MembrosExclusivosDaUnitPCN trava a escolha de origem.
//
// Tomar Base/ACBrCTe.Conversao.pas por fonte da verdade perderia estes
// dois membros -- o evento de Vinculo de Pagamento e o cancelamento dele --
// e nada acusaria: os demais 31 continuariam funcionando.
func TestSchemaCTe_MembrosExclusivosDaUnitPCN(t *testing.T) {
	exclusivos := []struct {
		v    SchemaCTe
		nome string
	}{
		{SchEvVincPgto, "evVincPgto"},
		{SchEvCancVincPgto, "evCancVincPgto"},
	}
	for _, c := range exclusivos {
		if c.v.String() != c.nome {
			t.Errorf("%s = %q", c.nome, c.v.String())
		}
		if v, err := ParseSchemaCTe(c.nome); err != nil || v != c.v {
			t.Errorf("Parse(%q) = %v %v", c.nome, v, err)
		}
	}
}

// TestSchemaCTe_NomeVemDoMEMBRO_ComACaixaEXATA trava os 33 nomes.
//
// O ACBr deriva o nome do identificador Pascal com GetEnumName e corta o
// prefixo "sch". A caixa vem do identificador, e o resultado e uma lista
// MISTA: sete nomes comecam em maiuscula e vinte e seis em minuscula.
//
// A lista inteira e conferida, e nao uma amostra: este nome sai no atributo
// `schema` do documento e no nome do arquivo `.xsd`. Uma passagem
// "organizando" tudo para PascalCase mudaria vinte e seis deles de uma vez,
// e o Parse nao acusaria -- ele compara sem caixa.
func TestSchemaCTe_NomeVemDoMEMBRO_ComACaixaEXATA(t *testing.T) {
	esperados := []struct {
		v    SchemaCTe
		nome string
	}{
		{SchErro, "Erro"},
		{SchCTe, "CTe"},
		{SchCTeOS, "CTeOS"},
		{SchGTVe, "GTVe"},
		{SchCancCTe, "cancCTe"},
		{SchInutCTe, "InutCTe"},
		{SchEventoCTe, "EventoCTe"},
		{SchCTeSimp, "CTeSimp"},
		{SchConsReciCTe, "consReciCTe"},
		{SchConsSitCTe, "consSitCTe"},
		{SchConsStatServCTe, "consStatServCTe"},
		{SchConsCad, "consCad"},
		{SchCTeModalAereo, "cteModalAereo"},
		{SchCTeModalAquaviario, "cteModalAquaviario"},
		{SchCTeModalDutoviario, "cteModalDutoviario"},
		{SchCTeModalFerroviario, "cteModalFerroviario"},
		{SchCTeModalRodoviario, "cteModalRodoviario"},
		{SchCTeMultiModal, "cteMultiModal"},
		{SchEvEPECCTe, "evEPECCTe"},
		{SchEvCancCTe, "evCancCTe"},
		{SchEvRegMultimodal, "evRegMultimodal"},
		{SchEvCCeCTe, "evCCeCTe"},
		{SchDistDFeInt, "distDFeInt"},
		{SchCTeModalRodoviarioOS, "cteModalRodoviarioOS"},
		{SchEvPrestDesacordo, "evPrestDesacordo"},
		{SchEvGTV, "evGTV"},
		{SchEvCECTe, "evCECTe"},
		{SchEvCancCECTe, "evCancCECTe"},
		{SchEvCancPrestDesacordo, "evCancPrestDesacordo"},
		{SchEvIECTe, "evIECTe"},
		{SchEvCancIECTe, "evCancIECTe"},
		{SchEvVincPgto, "evVincPgto"},
		{SchEvCancVincPgto, "evCancVincPgto"},
	}

	if len(esperados) != int(_fimSchemaCTe) {
		t.Fatalf("o enum tem %d membros e o teste confere %d",
			int(_fimSchemaCTe), len(esperados))
	}

	maiusculas := 0
	for _, c := range esperados {
		if got := c.v.String(); got != c.nome {
			t.Errorf("ordinal %d = %q, esperado %q", int(c.v), got, c.nome)
		}
		if c.nome != "" && c.nome[0] >= 'A' && c.nome[0] <= 'Z' {
			maiusculas++
		}
	}

	// A contagem trava a MISTURA. Normalizar tudo para um lado so --
	// o engano provavel -- mudaria este numero mesmo que alguem ajustasse
	// a tabela de esperados junto.
	if maiusculas != 7 {
		t.Errorf("sete nomes comecam em maiuscula e vinte e seis em minuscula; "+
			"contei %d maiusculos", maiusculas)
	}
}

// TestParseSchemaCTe_CortaNoUnderscoreEAceitaOPrefixo trava as duas
// normalizacoes que o Parse faz antes de buscar.
//
// O atributo `schema` dos documentos vem com a versao grudada
// ("cteModalRodoviario_v4.00"); sem o corte, nenhum schema seria
// reconhecido num documento real.
func TestParseSchemaCTe_CortaNoUnderscoreEAceitaOPrefixo(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado SchemaCTe
	}{
		{"cteModalRodoviario_v4.00", SchCTeModalRodoviario},
		{"cteModalRodoviario", SchCTeModalRodoviario},
		{"schcteModalRodoviario", SchCTeModalRodoviario},
		{"CTe_v4.00", SchCTe},
		{"evCancCTe_v4.00", SchEvCancCTe},
	}
	for _, c := range casos {
		v, err := ParseSchemaCTe(c.entrada)
		if err != nil {
			t.Errorf("%q: %v", c.entrada, err)
			continue
		}
		if v != c.esperado {
			t.Errorf("%q = %v, esperado %v", c.entrada, v, c.esperado)
		}
	}

	// Schema desconhecido devolve erro E o primeiro membro.
	//
	// DIVERGENCIA do original, que devolve TSchemaCTe(-1) e ainda marca
	// ok=True -- um enum fora de faixa que se propaga em silencio.
	v, err := ParseSchemaCTe("schNaoExiste")
	if !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("schema desconhecido deveria dar erro, veio %v", err)
	}
	if v != SchErro {
		t.Errorf("o valor de erro e SchErro, veio %v", v)
	}
}

// ---------------------------------------------------------------------------
// LayOut
// ---------------------------------------------------------------------------

// TestLayOutCTe_PaginasNaoTemNomeDeServico trava a distincao entre servico
// SOAP e pagina.
//
// Tres membros do enum resolvem para endereco de pagina -- o QR-Code e as
// duas consultas publicas. Dar nome de servico a eles faria a tabela de
// enderecos procurar uma chave que o `.ini` nao tem.
func TestLayOutCTe_PaginasNaoTemNomeDeServico(t *testing.T) {
	if LayCTeURLQRCode.Servico() != "" || LayURLConsultaCTe.Servico() != "" {
		t.Error("QR-Code e consulta publica sao PAGINA, nao servico SOAP")
	}
	if LayCTeRecepcao.Servico() != "CTeRecepcao" {
		t.Errorf("LayCTeRecepcao = %q", LayCTeRecepcao.Servico())
	}
	if LayCTeRecepcaoSimp.Servico() != "CTeRecepcaoSimp" {
		t.Errorf("LayCTeRecepcaoSimp = %q", LayCTeRecepcaoSimp.Servico())
	}

	// A busca por nome nao pode casar com os vazios.
	if _, err := ParseLayOutCTe(""); !errors.Is(err, ErrEnumInvalido) {
		t.Error("nome vazio casaria com o QR-Code se a busca nao pulasse os vazios")
	}

	// Ida e volta dos que SAO servico.
	for i := LayCTeRecepcao; i < _fimLayOutCTe; i++ {
		nome := i.Servico()
		if nome == "" {
			continue
		}
		if v, err := ParseLayOutCTe(nome); err != nil || v != i {
			t.Errorf("%q: %v %v", nome, v, err)
		}
	}
}

// TestVersaoCTe_FloatAcompanhaOCodigo trava as duas representacoes.
func TestVersaoCTe_FloatAcompanhaOCodigo(t *testing.T) {
	casos := []struct {
		v      VersaoCTe
		codigo string
		valor  float64
	}{
		{Ve200, "2.00", 2.00},
		{Ve300, "3.00", 3.00},
		{Ve400, "4.00", 4.00},
	}
	for _, c := range casos {
		if c.v.String() != c.codigo || c.v.Float() != c.valor {
			t.Errorf("%d: %q/%v, esperado %q/%v",
				c.v, c.v.String(), c.v.Float(), c.codigo, c.valor)
		}
	}
	// Fora de faixa devolve zero, como o case-else do Delphi.
	if VersaoCTe(99).Float() != 0 {
		t.Error("versao fora de faixa devolve 0")
	}
}

// TestEnums_ParseInvalido confere a convencao de erro do projeto.
func TestEnums_ParseInvalido(t *testing.T) {
	if _, err := ParseTipoCTe("99"); !errors.Is(err, ErrEnumInvalido) {
		t.Errorf("erro = %v, esperado ErrEnumInvalido", err)
	}
	if _, err := ParseModeloCTe("55"); !errors.Is(err, ErrEnumInvalido) {
		t.Error("o modelo 55 e da NF-e, nao do CT-e")
	}
}
