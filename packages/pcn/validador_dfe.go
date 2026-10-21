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
	"strconv"
	"strings"
)

// Validadores que as REGRAS DE NEGOCIO dos DFe consomem. Porte das funcoes
// homonimas de ACBrDFeUtil.pas e ACBrValidador.pas.
//
// Devolvem bool, nao error, porque e assim que as regras os usam: o texto da
// rejeicao vem da regra, nao do validador. A excecao e ValidarIE, que mora
// em validador_ie.go e tem erro proprio por causa das 27 variantes.

// municipiosSemDigito sao os codigos do IBGE que NAO fecham o digito
// verificador e ainda assim existem. Porte da constante NAO_VALIDAR
// (ACBrDFeUtil.pas:897).
//
// A regra do ACBr olha os SEIS primeiros digitos: se baterem com algum da
// lista, o codigo so e aceito se for exatamente um deles.
var municipiosSemDigito = map[int]bool{
	2201919: true, 2202251: true, 2201988: true, 2611533: true,
	3117836: true, 3152131: true, 4305871: true, 5203939: true,
	5203962: true,
}

// pesosMunicipio e o PESO da validacao do codigo do IBGE.
const pesosMunicipio = "1212120"

// municipioExterior e o codigo reservado para operacao com o exterior.
const municipioExterior = 9999999

// ValidarMunicipio confere o digito verificador do codigo do IBGE.
// Porte de ValidarMunicipio (ACBrDFeUtil.pas:893).
//
// O 9999999 do exterior e aceito sem conta nenhuma, e os nove municipios da
// lista de excecao tambem -- eles nao fecham o digito e existem assim mesmo.
func ValidarMunicipio(municipio int) bool {
	if municipio == municipioExterior {
		return true
	}

	codigo := strconv.Itoa(municipio)

	// A excecao casa pelos SEIS primeiros digitos, nao pelo codigo inteiro.
	if len(codigo) >= 6 {
		prefixo, err := strconv.Atoi(codigo[:6])
		if err == nil {
			for excecao := range municipiosSemDigito {
				if excecao/10 == prefixo {
					return municipiosSemDigito[municipio]
				}
			}
		}
	}

	if len(codigo) != 7 {
		return false
	}
	cUF, err := strconv.Atoi(codigo[:2])
	if err != nil || !ValidarCodigoUF(cUF) {
		return false
	}
	// Codigo de municipio zerado dentro da UF nao existe.
	if codigo[2:6] == "0000" {
		return false
	}

	soma := 0
	for i := 0; i < 7; i++ {
		valor := int(codigo[i]-'0') * int(pesosMunicipio[i]-'0')
		// Produto de dois digitos entra pela SOMA dos seus algarismos.
		if valor > 9 {
			soma += valor/10 + valor%10
		} else {
			soma += valor
		}
	}

	digito := (10 - soma%10) % 10
	return digito == int(codigo[6]-'0')
}

// ValidarGTIN confere o digito verificador de um GTIN-8, 12, 13 ou 14.
// Porte de TACBrValidador.ValidarGTIN (ACBrValidador.pas:1924).
func ValidarGTIN(numero string) bool {
	numero = strings.TrimSpace(numero)
	if numero == "" || !apenasDigitos(numero) {
		return false
	}
	switch len(numero) {
	case 8, 12, 13, 14:
	default:
		return false
	}

	codigo := numero[:len(numero)-1]
	return digitoGTIN(codigo) == int(numero[len(numero)-1]-'0')
}

// digitoGTIN calcula o verificador. O codigo (SEM o digito) e completado
// com zeros a esquerda ate TREZE posicoes antes da conta -- sem isso os
// pesos caem nas posicoes erradas para os GTIN mais curtos.
func digitoGTIN(codigo string) int {
	codigo = PreencherZerosEsquerda(codigo, 13)

	soma := 0
	for i := 0; i < len(codigo); i++ {
		// Peso 3 nas posicoes IMPARES contadas a partir de 1.
		peso := 1
		if (i+1)%2 == 1 {
			peso = 3
		}
		soma += int(codigo[i]-'0') * peso
	}

	return int(math.Ceil(float64(soma)/10))*10 - soma
}

// ValidarPrefixoGTIN confere o digito E o prefixo GS1 do codigo.
// Porte de TACBrValidador.ValidarPrefixoGTIN (ACBrValidador.pas:1690).
//
// Um GTIN pode ter o digito certo e um prefixo que nao pertence a nenhuma
// organizacao GS1 -- e o caso que esta funcao pega e o ValidarGTIN nao.
func ValidarPrefixoGTIN(numero string) bool {
	if !ValidarGTIN(numero) {
		return false
	}

	normalizado := PreencherZerosEsquerda(strings.TrimSpace(numero), 14)

	// Num GTIN-8 os seis primeiros digitos sao zero e o prefixo fica mais a
	// direita; nos demais ele comeca na segunda posicao.
	inicio := 1
	if normalizado[:6] == "000000" {
		inicio = 6
	}
	prefixo, err := strconv.Atoi(normalizado[inicio : inicio+3])
	if err != nil {
		return false
	}

	for _, f := range faixasPrefixoGTIN {
		if prefixo >= f.Inicio && prefixo <= f.Fim {
			return true
		}
	}
	return false
}

// ValidarISUF confere a Inscricao SUFRAMA.
// Porte de ValidarISUF (ACBrDFeUtil.pas:950).
//
// O laco do original vai de 1 a NOVE com peso (10-i), ou seja, o PROPRIO
// digito verificador entra na soma, com peso 1. Nao e engano de leitura: e
// o que faz o resto fechar.
//
// DIVERGENCIA: o ACBr confere o tamanho sobre OnlyNumber(ISUF) mas indexa a
// string CRUA, entao uma inscricao com mascara faria ele somar pontos e
// levantar excecao. Aqui as duas coisas usam o valor limpo.
func ValidarISUF(isuf string) bool {
	isuf = OnlyNumber(isuf)
	if len(isuf) < 9 {
		return false
	}

	soma := 0
	for i := 0; i < 9; i++ {
		soma += int(isuf[i]-'0') * (9 - i)
	}

	digito := 11 - soma%11
	if digito > 9 {
		digito = 0
	}
	return digito == int(isuf[8]-'0')
}

// ToleranciaCentavo e ToleranciaMilesimo sao as duas margens com que as
// regras de negocio comparam valores. Porte de TOLERANCIA_COMPARACAO_01 e
// TOLERANCIA_COMPARACAO_001 (ACBrNFe.ValidarRegrasdeNegocio.pas:112).
const (
	ToleranciaCentavo  = 0.01
	ToleranciaMilesimo = 0.001
)

// CompararValor devolve -1, 0 ou 1 comparando dois valores com tolerancia.
// Porte de ComparaValor (ACBrUtil.Math.pas:229).
//
// A tolerancia existe porque somar centenas de itens em ponto flutuante nao
// da exatamente o total declarado, e a SEFAZ aceita a diferenca.
//
// A diferenca e ARREDONDADA A NOVE CASAS antes de ser comparada
// (SimpleRoundTo(..., -9) no original). Sem isso, uma diferenca que vale
// exatamente a tolerancia pode sair como 0.0010000000000331966 e reprovar
// um documento que o ACBr aceitaria -- e este e o motor dos dezessete
// confrontos de total das regras de negocio, da 610, da 851 e do troco.
func CompararValor(a, b, tolerancia float64) int {
	diferenca := math.Round(math.Abs(a-b)*1e9) / 1e9
	if diferenca <= tolerancia {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}
