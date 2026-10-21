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
	"errors"
	"strconv"
	"testing"
)

// Testes do validador de Inscricao Estadual.
//
// Os vetores vem da PROPRIA SUITE DE TESTES DO ACBR
// (Testes/FPCUnit/ACBrDiversos/acbrvalidadortest.pas), com as mascaras
// originais preservadas -- elas exercitam de quebra o LimpaDocto.
//
// Usar os vetores do original, e nao exemplos gerados aqui, e o que torna
// estes testes capazes de reprovar o porte: uma IE calculada pelo proprio
// motor validaria contra si mesma mesmo com o algoritmo errado.

// iesValidas sao as inscricoes que o ACBr declara VALIDAS, por UF.
var iesValidas = []struct{ uf, ie string }{
	{"AC", "01.004.823/001-12"}, {"AC", "013456784"},
	{"AL", "240123450"},
	{"AM", "123123127"},
	{"AP", "030123459"}, {"AP", "030170011"}, {"AP", "030190225"}, {"AP", "030190231"},
	{"BA", "123456-63"}, {"BA", "173456-13"},
	{"CE", "023456787"},
	{"DF", "0734567890103"},
	{"ES", "123123127"},
	{"GO", "10.987.654-7"},
	{"MA", "120000385"},
	{"MG", "062.307.904/0081"},
	{"MS", "28.312.312-5"},
	{"MT", "0013000001-9"},
	{"PA", "15999999-5"},
	{"PB", "16000001-7"},
	{"PE", "0321418-40"}, {"PE", "18.1.001.0000004-9"},
	{"PI", "192345672"},
	{"PR", "123.45678-50"},
	{"RJ", "12.123.12-4"},
	{"RN", "20.040.040-1"}, {"RN", "20.0.040.040-0"},
	{"RO", "101.62521-3"}, {"RO", "0000000062521-3"},
	{"RR", "24008266-8"},
	{"RS", "224/3658792"},
	{"SC", "251.040.852"},
	{"SE", "27123456-3"},
	{"SP", "110.042.490.114"}, {"SP", "P-01100424.3/002"},
	{"TO", "01.022.783-0"}, {"TO", "29.01.022783-6"},
}

// iesInvalidas sao as que o ACBr declara INVALIDAS. A maioria difere da
// valida so no digito verificador -- e esse o par que prova discriminacao.
var iesInvalidas = []struct{ uf, ie string }{
	{"AC", "99999"}, {"AC", "01.004.823/001-99"},
	{"AL", "240123456"},
	{"AM", "123123123"},
	{"AP", "123456789"},
	{"BA", "123456-78"},
	{"CE", "123456789"},
	{"DF", "12345678901"}, {"DF", "1234567890123"},
	{"ES", "123123123"},
	{"GO", "12.312.312-3"},
	{"MA", "123123123"},
	{"MG", "123.123.123/9999"},
	{"MS", "123123123"},
	{"MT", "1234567890-1"},
	{"PA", "15999999-9"},
	{"PB", "06000001-9"},
	{"PE", "0321418-99"}, {"PE", "18.1.001.0000004-0"},
	{"PI", "012345678"},
	{"PR", "123.45678-99"},
	{"RJ", "12.123.12-9"},
	{"RN", "20.040.040-9"}, {"RN", "20.0.040.040-9"},
	{"RO", "101.12345-6"}, {"RO", "1234567890521-3"},
	{"RR", "12345678-8"},
	{"RS", "224/1234567"},
	{"SC", "123.123.123"},
	{"SE", "12312312-3"},
	{"SP", "123.123.123.123"}, {"SP", "P-12345678.9/002"},
	{"TO", "12.123.123-9"}, {"TO", "12.34.567890-6"},
}

func TestValidarIE_AceitaAsValidasDoACBr(t *testing.T) {
	for _, c := range iesValidas {
		if err := ValidarIE(c.ie, c.uf); err != nil {
			t.Errorf("%s: ValidarIE(%q) = %v, esperado valida", c.uf, c.ie, err)
		}
	}
}

func TestValidarIE_ReprovaAsInvalidasDoACBr(t *testing.T) {
	for _, c := range iesInvalidas {
		if err := ValidarIE(c.ie, c.uf); err == nil {
			t.Errorf("%s: ValidarIE(%q) aceitou, esperado reprovar", c.uf, c.ie)
		}
	}
}

// TestValidarIE_DigitoTrocadoReprova vai alem dos vetores do ACBr: pega cada
// IE valida e troca o ULTIMO digito por todos os outros nove.
//
// Cobre o modo de falha que nenhum vetor positivo alcanca -- um motor que
// aceitasse tudo passaria nos dois testes acima menos nos negativos, e um
// que errasse so em uma UF poderia escapar se o vetor negativo daquela UF
// falhasse por outro motivo (tamanho, prefixo).
func TestValidarIE_DigitoTrocadoReprova(t *testing.T) {
	for _, c := range iesValidas {
		// A inscricao de produtor rural de SP tem o DV no MEIO; trocar o
		// ultimo digito dela nao mexe em verificador nenhum.
		if c.uf == "SP" && limparDocumento(c.ie)[0] == 'P' {
			continue
		}

		limpa := limparDocumento(c.ie)
		pos := len(limpa) - 1
		for d := byte('0'); d <= '9'; d++ {
			if d == limpa[pos] {
				continue
			}
			alterada := limpa[:pos] + string(d)
			if ValidarIE(alterada, c.uf) == nil {
				t.Errorf("%s: ValidarIE(%q) aceitou; e %q com o ultimo digito trocado",
					c.uf, alterada, limpa)
			}
		}
	}
}

func TestValidarIE_IsentoEhValida(t *testing.T) {
	for _, s := range []string{"ISENTO", "isento", "  Isento  "} {
		if err := ValidarIE(s, "SP"); err != nil {
			t.Errorf("ValidarIE(%q) = %v, esperado nil", s, err)
		}
	}
}

// TestValidarIE_LimpaMascara trava o LimpaDocto: ponto, barra e hifen saem
// antes da validacao, e so eles.
func TestValidarIE_LimpaMascara(t *testing.T) {
	// Espaco NO MEIO nao e removido -- o LimpaDocto so faz Trim nas pontas.
	if err := ValidarIE("110 042490114", "SP"); err == nil {
		t.Error("IE com espaco no meio deveria reprovar: o LimpaDocto do ACBr " +
			"so remove ponto, barra e hifen")
	}
	// Com e sem mascara tem de dar o mesmo resultado.
	if err := ValidarIE("110042490114", "SP"); err != nil {
		t.Errorf("a mesma IE sem mascara reprovou: %v", err)
	}
}

func TestValidarIE_EntradasInvalidas(t *testing.T) {
	casos := []struct{ ie, uf, porque string }{
		{"", "SP", "vazia (PermiteVazio e False no wrapper)"},
		{"110042490114", "", "UF nao informada"},
		{"110042490114", "ZZ", "UF desconhecida"},
		{"11004249011A", "SP", "letra fora da primeira posicao"},
		{"123456789", "EX", "EX passa na validacao de UF mas nao tem regra de IE"},
	}
	for _, c := range casos {
		err := ValidarIE(c.ie, c.uf)
		if err == nil {
			t.Errorf("ValidarIE(%q, %q) aceitou; esperado reprovar: %s",
				c.ie, c.uf, c.porque)
			continue
		}
		if !errors.Is(err, ErrIEInvalida) {
			t.Errorf("ValidarIE(%q, %q) devolveu %v; esperado envolver ErrIEInvalida",
				c.ie, c.uf, err)
		}
	}
}

// TestValidarIE_BAEscolheModuloPeloSegundoDigito trava a unica UF cujo
// MODULO depende do conteudo do documento (ACBrValidador.pas:1224).
//
// A ORDEM importa e e facil de inverter: o ACBr completa com zeros ate nove
// posicoes ANTES de olhar o segundo caractere. Os dois vetores validos da BA
// no ACBr tem oito digitos e, depois do preenchimento, caem os DOIS no
// modulo 10 -- entao eles nao cobrem o outro ramo, e por isso o teste
// exercita regraDaUF direto, com documentos ja de nove posicoes.
func TestValidarIE_BAEscolheModuloPeloSegundoDigito(t *testing.T) {
	casos := []struct {
		doc    string
		modulo int
		porque string
	}{
		{"012345663", 10, "segundo caractere 1, na lista 0123458"},
		{"672345678", 11, "segundo caractere 7, fora da lista"},
		{"082345678", 10, "segundo caractere 8, ultimo da lista"},
		{"092345678", 11, "segundo caractere 9, fora da lista"},
		// Oito digitos: o preenchimento com zero entra ANTES do teste, e o
		// segundo caractere passa a ser o PRIMEIRO do documento original.
		{"72345678", 11, "completado para 072345678, segundo caractere 7"},
	}
	for _, c := range casos {
		r, _ := regraDaUF("BA", c.doc)
		if r.xMD != c.modulo || r.yMD != c.modulo {
			t.Errorf("BA %q: modulo %d/%d, esperado %d (%s)",
				c.doc, r.xMD, r.yMD, c.modulo, c.porque)
		}
	}
}

// TestValidarIE_RegraEscolhidaPeloTamanho trava as seis UFs com duas regras.
func TestValidarIE_RegraEscolhidaPeloTamanho(t *testing.T) {
	casos := []struct {
		uf      string
		doc     string
		tamanho int
	}{
		{"AC", "123456789", 9},
		{"AC", "1234567890123", 13},
		{"PE", "123456789", 9},
		{"PE", "12345678901234", 14},
		{"RN", "123456789", 9},
		{"RN", "1234567890", 10},
		{"RO", "123456789", 9},
		{"RO", "12345678901234", 14},
		{"TO", "12345678901", 11},
		{"TO", "123456789", 9},
		{"SP", "123456789012", 12},
		{"SP", "P123456789012", 13},
	}
	for _, c := range casos {
		r, _ := regraDaUF(c.uf, c.doc)
		if r.tamanho != c.tamanho {
			t.Errorf("regraDaUF(%s, %d chars) tamanho = %d, esperado %d",
				c.uf, len(c.doc), r.tamanho, c.tamanho)
		}
	}
}

// TestValidarIE_DocumentoCompletadoComZeros trava as duas UFs que MUTAM o
// documento antes de validar (ACBrValidador.pas:1216 e :1279).
func TestValidarIE_DocumentoCompletadoComZeros(t *testing.T) {
	if _, doc := regraDaUF("BA", "12345606"); doc != "012345606" {
		t.Errorf("BA com 8 digitos: doc = %q, esperado completado ate 9", doc)
	}
	if _, doc := regraDaUF("MT", "130000019"); doc != "00130000019" {
		t.Errorf("MT com 9 digitos: doc = %q, esperado completado ate 11", doc)
	}

	// Efeito visivel: a mesma IE da MT passa com 9 ou com 11 digitos.
	if err := ValidarIE("130000019", "MT"); err != nil {
		t.Errorf("MT com 9 digitos reprovou: %v", err)
	}
}

// TestValidarIE_APMudaDeRegraPorFaixa trava a unica UF que escolhe a
// ROTACAO pelo valor do documento (ACBrValidador.pas:1201).
//
// Os quatro vetores validos da AP no ACBr existem para cobrir os tres
// ramos: abaixo da faixa, dentro dela (fatorF = 1) e acima (rotacao E).
func TestValidarIE_APMudaDeRegraPorFaixa(t *testing.T) {
	abaixo, _ := regraDaUF("AP", "030123459")
	dentro, _ := regraDaUF("AP", "030170011")
	acima, _ := regraDaUF("AP", "030190231")

	if abaixo.fatorF != 0 || abaixo.xROT != "CE" {
		t.Errorf("AP abaixo da faixa: fatorF=%d rot=%q, esperado 0 e CE",
			abaixo.fatorF, abaixo.xROT)
	}
	if dentro.fatorF != 1 || dentro.xROT != "CE" {
		t.Errorf("AP dentro da faixa: fatorF=%d rot=%q, esperado 1 e CE",
			dentro.fatorF, dentro.xROT)
	}
	if acima.xROT != "E" {
		t.Errorf("AP acima da faixa: rot=%q, esperado E", acima.xROT)
	}
}

// TestValidarIE_GOUsaFatorGNaFaixa trava a unica UF que substitui o resto 10
// por algo diferente de zero (ACBrValidador.pas:1266).
func TestValidarIE_GOUsaFatorGNaFaixa(t *testing.T) {
	dentro, _ := regraDaUF("GO", "101031050")
	fora, _ := regraDaUF("GO", "101031049")

	if dentro.fatorG != 1 {
		t.Error("GO dentro da faixa 101031050..101199979 deveria usar fatorG = 1")
	}
	if fora.fatorG != 0 {
		t.Error("GO fora da faixa deveria manter fatorG = 0")
	}
}

// TestPesosIE_TabelaConfereComOOriginal reproduz a matriz do .pas por
// transcricao INDEPENDENTE, linha a linha.
//
// Existe porque um erro de digitacao num peso nao aparece necessariamente
// nos vetores do ACBr: a linha errada pode ser a de uma UF cujo vetor
// negativo reprova por outro motivo. Aqui a matriz e escrita de novo, a
// partir do original, e comparada celula a celula.
func TestPesosIE_TabelaConfereComOOriginal(t *testing.T) {
	esperado := map[int]string{
		1:  "0 2 3 4 5 6 7 8 9 2 3 4 5 6",
		2:  "0 0 2 3 4 5 6 7 8 9 2 3 4 5",
		3:  "2 0 3 4 5 6 7 8 9 2 3 4 5 6",
		4:  "0 2 3 4 5 6 0 0 0 0 0 0 0 0",
		5:  "0 8 7 6 5 4 3 2 1 0 0 0 0 0",
		6:  "0 2 3 4 5 6 7 0 0 8 9 0 0 0",
		7:  "0 2 3 4 5 6 7 8 9 1 2 3 4 5",
		8:  "0 2 3 4 5 6 7 2 3 4 5 6 7 8",
		9:  "0 0 2 3 4 5 6 7 2 3 4 5 6 7",
		10: "0 0 2 1 2 1 2 1 2 1 1 2 1 0",
		11: "0 2 3 4 5 6 7 8 9 10 11 2 3 0",
		12: "0 0 0 0 10 8 7 6 5 4 3 1 0 0",
		13: "0 2 3 4 5 6 7 8 9 10 2 3 0 0",
		14: "0 0 2 3 4 5 6 7 8 3 4 5 6 7",
	}

	for linha, texto := range esperado {
		campos := camposDe(texto)
		if len(campos) != 14 {
			t.Fatalf("linha %d do teste tem %d pesos, esperado 14", linha, len(campos))
		}
		for col := 1; col <= 14; col++ {
			querido, err := strconv.Atoi(campos[col-1])
			if err != nil {
				t.Fatal(err)
			}
			if pesosIE[linha][col] != querido {
				t.Errorf("pesosIE[%d][%d] = %d, esperado %d",
					linha, col, pesosIE[linha][col], querido)
			}
		}
	}
}

// camposDe quebra em espacos, sem strings.Fields, para deixar claro que a
// transcricao acima e literal.
func camposDe(s string) []string {
	var campos []string
	atual := ""
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			if atual != "" {
				campos = append(campos, atual)
				atual = ""
			}
			continue
		}
		atual += string(s[i])
	}
	if atual != "" {
		campos = append(campos, atual)
	}
	return campos
}

// TestValidarIE_GOTemRegraDePrefixo trava a unica UF com regra de PREFIXO
// alem do vetor de posicoes (ACBrValidador.pas:1261-1264).
//
// O vetor sozinho so restringe o primeiro digito a 1 ou 2; sem a regra, os
// prefixos 12 a 19 passam. Os vetores do proprio ACBr NAO cobrem o caso --
// o negativo de GO ("12.312.312-3") reprova pelo digito verificador, e a
// ausencia da regra fica invisivel.
func TestValidarIE_GOTemRegraDePrefixo(t *testing.T) {
	// Prefixo fora de [10, 11, 20..29], com digito verificador VALIDO: a
	// reprovacao so pode vir do prefixo.
	const prefixoInvalido = "150000006"
	r, _ := regraDaUF("GO", prefixoInvalido)
	if r.erroPrefixo == "" {
		t.Fatalf("o caso de teste %q precisa ter prefixo recusado", prefixoInvalido)
	}
	if err := ValidarIE(prefixoInvalido, "GO"); err == nil {
		t.Errorf("ValidarIE(%q, GO) aceitou; o prefixo 15 esta fora da faixa",
			prefixoInvalido)
	}

	validos := []int{10, 11, 20, 25, 29}
	for _, p := range validos {
		if !prefixoValidoGO(p) {
			t.Errorf("o prefixo %d deveria ser aceito", p)
		}
	}
	for _, p := range []int{0, 9, 12, 15, 19, 30, 99} {
		if prefixoValidoGO(p) {
			t.Errorf("o prefixo %d deveria ser recusado", p)
		}
	}

	// A IE valida do ACBr tem prefixo 10 e continua passando.
	if err := ValidarIE("10.987.654-7", "GO"); err != nil {
		t.Errorf("a IE valida de GO do ACBr reprovou: %v", err)
	}
}
