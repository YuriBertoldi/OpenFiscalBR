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

package pcn

import (
	"math"
	"sort"
	"testing"
)

// TestValidarMunicipio_DigitoVerificador usa codigos REAIS do IBGE, com o
// digito conferido a mao pela regra de pesos 1212120.
func TestValidarMunicipio_DigitoVerificador(t *testing.T) {
	validos := []struct {
		codigo int
		nome   string
	}{
		{3550308, "Sao Paulo"},
		{3304557, "Rio de Janeiro"},
		{3106200, "Belo Horizonte"},
		{4106902, "Curitiba"},
		{2927408, "Salvador"},
		{5300108, "Brasilia"},
		{1302603, "Manaus"},
		{2304400, "Fortaleza"},
	}
	for _, c := range validos {
		if !ValidarMunicipio(c.codigo) {
			t.Errorf("ValidarMunicipio(%d) = false, esperado true (%s)", c.codigo, c.nome)
		}
	}

	// O digito trocado tem de reprovar -- senao a funcao aceitaria tudo.
	for _, c := range validos {
		for d := 0; d <= 9; d++ {
			alterado := c.codigo/10*10 + d
			if alterado == c.codigo {
				continue
			}
			if ValidarMunicipio(alterado) {
				t.Errorf("ValidarMunicipio(%d) = true; e %s com o digito trocado",
					alterado, c.nome)
			}
		}
	}
}

// TestValidarMunicipio_Excecoes trava as duas saidas antecipadas: o codigo
// do exterior e a lista dos nove municipios que NAO fecham o digito.
func TestValidarMunicipio_Excecoes(t *testing.T) {
	if !ValidarMunicipio(9999999) {
		t.Error("o codigo 9999999 do exterior deveria ser aceito sem conta")
	}

	// Os nove da lista sao aceitos apesar do digito nao fechar.
	for codigo := range municipiosSemDigito {
		if !ValidarMunicipio(codigo) {
			t.Errorf("ValidarMunicipio(%d) = false; esta na lista de excecao", codigo)
		}
	}

	// Mas um VIZINHO do mesmo prefixo de seis digitos e recusado: a regra
	// casa pelo prefixo e so entao exige o codigo exato.
	if ValidarMunicipio(2201910) {
		t.Error("2201910 compartilha o prefixo 220191 com uma excecao, mas nao " +
			"esta na lista; deveria reprovar")
	}
}

func TestValidarMunicipio_FormatoInvalido(t *testing.T) {
	casos := []struct {
		codigo int
		porque string
	}{
		{355030, "seis digitos"},
		{35503081, "oito digitos"},
		{9950308, "codigo de UF inexistente"},
		{3500008, "municipio zerado dentro da UF"},
	}
	for _, c := range casos {
		if ValidarMunicipio(c.codigo) {
			t.Errorf("ValidarMunicipio(%d) = true; esperado reprovar: %s",
				c.codigo, c.porque)
		}
	}
}

// TestValidarGTIN cobre os quatro tamanhos aceitos. O digito de cada codigo
// foi conferido a mao pela regra 3-1 do GS1.
func TestValidarGTIN(t *testing.T) {
	validos := []struct{ gtin, nota string }{
		{"12345670", "GTIN-8"},
		{"614141000036", "GTIN-12"},
		{"7891910000197", "GTIN-13 real"},
		{"07891910000197", "GTIN-14 (o mesmo com zero a esquerda)"},
	}
	for _, c := range validos {
		if !ValidarGTIN(c.gtin) {
			t.Errorf("ValidarGTIN(%q) = false, esperado true (%s)", c.gtin, c.nota)
		}
	}

	for _, c := range validos {
		for d := byte('0'); d <= '9'; d++ {
			if d == c.gtin[len(c.gtin)-1] {
				continue
			}
			alterado := c.gtin[:len(c.gtin)-1] + string(d)
			if ValidarGTIN(alterado) {
				t.Errorf("ValidarGTIN(%q) = true; e %q com o digito trocado",
					alterado, c.gtin)
			}
		}
	}
}

func TestValidarGTIN_TamanhoENaoNumerico(t *testing.T) {
	casos := []struct{ gtin, porque string }{
		{"", "vazio"},
		{"1234567", "sete digitos"},
		{"123456789", "nove digitos"},
		{"1234567890", "dez digitos"},
		{"123456789012345", "quinze digitos"},
		{"789191000019X", "caractere nao numerico"},
	}
	for _, c := range casos {
		if ValidarGTIN(c.gtin) {
			t.Errorf("ValidarGTIN(%q) = true; esperado reprovar: %s", c.gtin, c.porque)
		}
	}
}

// TestValidarPrefixoGTIN trava o que o ValidarGTIN NAO pega: codigo com o
// digito certo mas prefixo fora das faixas da GS1.
func TestValidarPrefixoGTIN(t *testing.T) {
	// 789 e a faixa do Brasil; o codigo tem digito valido e prefixo valido.
	if !ValidarPrefixoGTIN("7891910000197") {
		t.Error("GTIN brasileiro (prefixo 789) deveria passar")
	}

	// Prefixo 975 cai na LACUNA entre as faixas 960-969 e 977-977, mas o
	// digito verificador fecha -- e exatamente esse o caso que separa as
	// duas funcoes.
	const prefixoOrfao = "9751910000192"
	if !ValidarGTIN(prefixoOrfao) {
		t.Fatalf("o caso de teste %q precisa ter digito VALIDO para provar que "+
			"a reprovacao vem do prefixo", prefixoOrfao)
	}
	if ValidarPrefixoGTIN(prefixoOrfao) {
		t.Error("prefixo 975 nao pertence a nenhuma faixa GS1 e deveria reprovar")
	}
}

// TestFaixasPrefixoGTIN_NaoSeSobrepoem confere a integridade da tabela
// transcrita: 127 faixas, nenhuma invertida e nenhuma sobreposta.
//
// A tabela do ACBr NAO esta em ordem crescente -- 490-499 vem antes de
// 460-469 --, e tudo bem: a busca e varredura linear. Por isso o teste
// ordena uma copia antes de conferir sobreposicao, em vez de exigir ordem.
//
// Um erro de digitacao que invertesse inicio e fim, ou que fizesse duas
// faixas colidirem, passaria despercebido nos testes de codigo individual.
func TestFaixasPrefixoGTIN_NaoSeSobrepoem(t *testing.T) {
	if len(faixasPrefixoGTIN) != 127 {
		t.Errorf("a tabela tem %d faixas, esperado 127", len(faixasPrefixoGTIN))
	}

	ordenadas := append([]faixaPrefixoGTIN(nil), faixasPrefixoGTIN...)
	sort.Slice(ordenadas, func(i, j int) bool {
		return ordenadas[i].Inicio < ordenadas[j].Inicio
	})

	anterior := -1
	for i, f := range ordenadas {
		if f.Inicio > f.Fim {
			t.Errorf("faixa %d: inicio %d maior que fim %d", i, f.Inicio, f.Fim)
		}
		if f.Inicio <= anterior {
			t.Errorf("faixa que comeca em %d colide com a anterior, que termina em %d",
				f.Inicio, anterior)
		}
		anterior = f.Fim
	}
}

// TestValidarISUF confere a Inscricao SUFRAMA.
func TestValidarISUF(t *testing.T) {
	// Digito conferido a mao: pesos 9..1 sobre os NOVE digitos -- o proprio
	// verificador entra na soma.
	const valida = "100016754"
	if !ValidarISUF(valida) {
		t.Errorf("ValidarISUF(%q) = false, esperado true", valida)
	}

	for d := byte('0'); d <= '9'; d++ {
		if d == valida[8] {
			continue
		}
		alterada := valida[:8] + string(d)
		if ValidarISUF(alterada) {
			t.Errorf("ValidarISUF(%q) = true; e %q com o digito trocado",
				alterada, valida)
		}
	}

	if ValidarISUF("1000167") {
		t.Error("ISUF com menos de nove digitos deveria reprovar")
	}
}

// TestCompararValor trava a tolerancia, que e o que permite a soma de
// centenas de itens em ponto flutuante bater com o total declarado.
func TestCompararValor(t *testing.T) {
	casos := []struct {
		a, b, tolerancia float64
		esperado         int
		porque           string
	}{
		{10.00, 10.00, ToleranciaMilesimo, 0, "iguais"},
		{10.0009, 10.00, ToleranciaMilesimo, 0, "dentro do milesimo"},
		{10.002, 10.00, ToleranciaMilesimo, 1, "fora do milesimo, maior"},
		{9.998, 10.00, ToleranciaMilesimo, -1, "fora do milesimo, menor"},
		{10.009, 10.00, ToleranciaCentavo, 0, "dentro do centavo"},
		{10.02, 10.00, ToleranciaCentavo, 1, "fora do centavo"},
		// O limite e inclusivo: exatamente a tolerancia ainda conta como igual.
		{10.001, 10.00, ToleranciaMilesimo, 0, "exatamente a tolerancia"},
	}
	for _, c := range casos {
		if got := CompararValor(c.a, c.b, c.tolerancia); got != c.esperado {
			t.Errorf("CompararValor(%v, %v, %v) = %d, esperado %d (%s)",
				c.a, c.b, c.tolerancia, got, c.esperado, c.porque)
		}
	}
}

// TestCompararValor_ArredondaADiferenca trava o SimpleRoundTo(-9) do
// original (ACBrUtil.Math.pas:236).
//
// Sem ele, uma diferenca que vale EXATAMENTE a tolerancia pode sair do
// ponto flutuante como 0.0010000000000331966 e reprovar um documento que o
// ACBr aceitaria. Como esta funcao e o motor dos dezessete confrontos de
// total, da regra 610 e da aritmetica do troco, o efeito seria uma rejeicao
// falsa em nota valida.
func TestCompararValor_ArredondaADiferenca(t *testing.T) {
	// Esta soma NAO da exatamente 1234.567 em float64.
	soma := 0.0
	for i := 0; i < 1000; i++ {
		soma += 1.234567
	}
	total := 1234.567

	if math.Abs(soma-total) == 0 {
		t.Skip("a soma fechou exata nesta plataforma; o caso nao e exercitado")
	}
	if CompararValor(soma, total, ToleranciaMilesimo) != 0 {
		t.Errorf("soma acumulada %.17f contra total %.2f deveria ser igual "+
			"dentro do milesimo; diferenca bruta %.17f", soma, total, soma-total)
	}

	// E o arredondamento NAO pode engolir diferenca real.
	if CompararValor(10.002, 10.000, ToleranciaMilesimo) != 1 {
		t.Error("diferenca de 2 milesimos esta FORA da tolerancia de 1 milesimo")
	}
}
