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
	"fmt"
	"strings"
)

// Validacao de Inscricao Estadual. Porte de TACBrValidador.ValidarIE
// (ACBrValidador.pas:1101-1580).
//
// Cada UF tem seu proprio algoritmo, mas o ACBr NAO os escreve um a um: ele
// descreve cada UF como uma linha de TABELA e roda um motor unico sobre ela.
// O porte mantem esse arranjo, porque e o que torna a conferencia com o
// original possivel -- 27 funcoes separadas esconderiam a diferenca entre
// elas no meio do codigo repetido.
//
// A tabela descreve, por UF:
//
//   - o tamanho esperado;
//   - o que cada POSICAO aceita (digito, digito verificador, literal, lista
//     ou faixa), sempre listada da DIREITA para a esquerda;
//   - qual linha de pesos usar, qual o modulo, e quais ajustes aplicar a
//     soma antes de tirar o resto;
//   - opcionalmente, um SEGUNDO passo, para as UFs com dois digitos
//     verificadores.

// pesosIE e o cPesos do ACBr: 14 linhas de 14 pesos.
//
// Indexado a partir de 1 nos DOIS eixos, com a linha e a coluna zero
// descartadas, para que os indices batam com os do .pas sem conversao -- a
// tabela e conferida contra o original a olho, e renumerar convidaria erro.
//
// A COLUNA tambem conta da direita para a esquerda: a coluna 1 multiplica o
// ultimo caractere do documento alinhado em 14 posicoes.
var pesosIE = [15][15]int{
	{},
	/* 1 */ {0, 0, 2, 3, 4, 5, 6, 7, 8, 9, 2, 3, 4, 5, 6},
	/* 2 */ {0, 0, 0, 2, 3, 4, 5, 6, 7, 8, 9, 2, 3, 4, 5},
	/* 3 */ {0, 2, 0, 3, 4, 5, 6, 7, 8, 9, 2, 3, 4, 5, 6},
	/* 4 */ {0, 0, 2, 3, 4, 5, 6, 0, 0, 0, 0, 0, 0, 0, 0},
	/* 5 */ {0, 0, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0, 0, 0, 0},
	/* 6 */ {0, 0, 2, 3, 4, 5, 6, 7, 0, 0, 8, 9, 0, 0, 0},
	/* 7 */ {0, 0, 2, 3, 4, 5, 6, 7, 8, 9, 1, 2, 3, 4, 5},
	/* 8 */ {0, 0, 2, 3, 4, 5, 6, 7, 2, 3, 4, 5, 6, 7, 8},
	/* 9 */ {0, 0, 0, 2, 3, 4, 5, 6, 7, 2, 3, 4, 5, 6, 7},
	/* 10 */ {0, 0, 0, 2, 1, 2, 1, 2, 1, 2, 1, 1, 2, 1, 0},
	/* 11 */ {0, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 2, 3, 0},
	/* 12 */ {0, 0, 0, 0, 0, 10, 8, 7, 6, 5, 4, 3, 1, 0, 0},
	/* 13 */ {0, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 2, 3, 0, 0},
	/* 14 */ {0, 0, 0, 2, 3, 4, 5, 6, 7, 8, 3, 4, 5, 6, 7},
}

// digito0a9 e o c0_9 do ACBr: a posicao aceita qualquer algarismo.
const digito0a9 = "0-9"

// regraIE descreve o algoritmo de uma UF.
type regraIE struct {
	// tamanho esperado do documento ja limpo. Zero significa UF sem regra
	// -- o caso do "EX", que passa na validacao de UF e reprova aqui.
	tamanho int

	// xROT, xMD e xTP sao a rotacao, o modulo e a linha de pesos do PRIMEIRO
	// digito verificador. A rotacao e um conjunto de letras: cada uma liga
	// um ajuste (ver aplicarRotacao e calcularDV).
	xROT string
	xMD  int
	xTP  int

	// yROT, yMD e yTP descrevem o SEGUNDO digito verificador. yTP zero
	// significa que a UF so tem um.
	yROT string
	yMD  int
	yTP  int

	// fatorF e fatorG substituem o digito quando o resto da 11 ou 10. Só GO
	// usa fatorG diferente de zero.
	fatorF int
	fatorG int

	// erroPrefixo e preenchido pela UF que recusa o documento pelo PREFIXO,
	// independentemente do digito verificador. So GO tem uma
	// (ACBrValidador.pas:1263).
	//
	// No original a verificacao apenas GRAVA a mensagem e segue: o calculo
	// do digito roda do mesmo jeito, e o documento e reprovado no fim por
	// causa dela. O efeito observavel e o mesmo de reprovar aqui.
	erroPrefixo string

	// digitos descreve as 14 posicoes, da DIREITA para a esquerda:
	// digitos[0] e o ultimo caractere. Cada entrada e uma de cinco formas:
	//
	//	""        a posicao nao existe (o documento e mais curto)
	//	"DVX"     primeiro digito verificador
	//	"DVY"     segundo digito verificador
	//	"0-9"     qualquer algarismo (ou outra faixa, como "1-9")
	//	"1,7,8"   um dos listados
	//	"P"       exatamente esse caractere
	digitos [14]string
}

// ErrIEInvalida e devolvido quando a Inscricao Estadual nao passa.
var ErrIEInvalida = fmt.Errorf("pcn: inscricao estadual invalida")

// ValidarIE confere a Inscricao Estadual contra o algoritmo da UF.
//
// Porte de ACBrValidador.ValidarIE. Devolve nil quando valida.
//
// "ISENTO" e aceito: e o conteudo que o leiaute manda informar quando o
// contribuinte nao tem inscricao.
func ValidarIE(ie, uf string) error {
	doc := limparDocumento(ie)
	uf = strings.ToUpper(strings.TrimSpace(uf))

	// PermiteVazio e False no wrapper ValidarDocumento (ACBrValidador.pas:344).
	if doc == "" {
		return fmt.Errorf("%w: nao informada", ErrIEInvalida)
	}
	if doc == "ISENTO" {
		return nil
	}
	if uf == "" {
		return fmt.Errorf("%w: UF nao informada", ErrIEInvalida)
	}
	if !ufValidaParaIE(uf) {
		return fmt.Errorf("%w: UF %q desconhecida", ErrIEInvalida, uf)
	}

	// So algarismos, com a excecao do "P" inicial da Inscricao de Produtor
	// Rural de SP (ACBrValidador.pas:1146).
	if !apenasDigitos(doc[1:]) || !(ehDigito(doc[0]) || doc[0] == 'P') {
		return fmt.Errorf("%w: caracteres invalidos", ErrIEInvalida)
	}

	regra, doc := regraDaUF(uf, doc)

	if regra.erroPrefixo != "" {
		return fmt.Errorf("%w para %s: %s", ErrIEInvalida, uf, regra.erroPrefixo)
	}

	if regra.tamanho == 0 || len(doc) != regra.tamanho {
		return fmt.Errorf("%w para %s: esperados %d caracteres, informados %d",
			ErrIEInvalida, uf, regra.tamanho, len(doc))
	}

	// O motor trabalha sempre sobre 14 posicoes, completadas a esquerda com
	// ESPACO -- e o espaco e o que marca "posicao inexistente" na tabela.
	alinhado := preencherEspacosEsquerda(doc, 14)

	dvx, dvy, err := conferirPosicoes(alinhado, regra, uf)
	if err != nil {
		return err
	}
	return conferirDigitos(alinhado, regra, dvx, dvy, uf)
}

// conferirPosicoes valida cada posicao contra a tabela e colhe os digitos
// verificadores informados. Porte do laco de ACBrValidador.pas:1473-1506.
func conferirPosicoes(alinhado string, regra regraIE, uf string) (dvx, dvy int, err error) {
	for i := 13; i >= 0; i-- {
		d := alinhado[13-i]
		aceito := regra.digitos[i]

		switch {
		case aceito == "":
			// Posicao fora do documento: so o preenchimento serve.
			if d != ' ' {
				return 0, 0, erroPosicaoIE(uf, i, regra, aceito)
			}

		case aceito == "DVX" || aceito == "DVY" || aceito == digito0a9:
			if !ehDigito(d) {
				return 0, 0, erroPosicaoIE(uf, i, regra, aceito)
			}
			if aceito == "DVX" {
				dvx = int(d - '0')
			} else if aceito == "DVY" {
				dvy = int(d - '0')
			}

		case strings.Contains(aceito, ","):
			if !strings.ContainsRune(aceito, rune(d)) {
				return 0, 0, erroPosicaoIE(uf, i, regra, aceito)
			}

		case strings.Contains(aceito, "-"):
			// Faixa no formato "a-b", conferida caractere a caractere.
			if d < aceito[0] || d > aceito[2] {
				return 0, 0, erroPosicaoIE(uf, i, regra, aceito)
			}

		default:
			if d != aceito[0] {
				return 0, 0, erroPosicaoIE(uf, i, regra, aceito)
			}
		}
	}
	return dvx, dvy, nil
}

// erroPosicaoIE numera a posicao como o ACBr faz na mensagem: contando da
// esquerda do documento REAL, nao das 14 posicoes alinhadas.
func erroPosicaoIE(uf string, i int, regra regraIE, aceito string) error {
	return fmt.Errorf("%w para %s: digito %d deveria ser %q",
		ErrIEInvalida, uf, 14-i-(14-regra.tamanho), aceito)
}

// conferirDigitos roda o calculo do digito verificador, uma ou duas vezes.
// Porte do laco de ACBrValidador.pas:1508-1574.
func conferirDigitos(alinhado string, regra regraIE, dvx, dvy int, uf string) error {
	rot, md, tp := regra.xROT, regra.xMD, regra.xTP
	informado, nome := dvx, "primeiro"

	for tp > 0 {
		calculado := calcularDV(alinhado, rot, md, tp, regra)
		if calculado != informado {
			return fmt.Errorf("%w para %s: %s digito verificador e %d, calculado %d",
				ErrIEInvalida, uf, nome, informado, calculado)
		}
		if nome == "segundo" {
			break
		}
		rot, md, tp = regra.yROT, regra.yMD, regra.yTP
		informado, nome = dvy, "segundo"
	}
	return nil
}

// calcularDV soma os digitos ponderados, aplica a rotacao e tira o resto.
func calcularDV(alinhado, rot string, md, tp int, regra regraIE) int {
	soma, somaDezenas := 0, 0
	for i := 14; i >= 1; i-- {
		d := alinhado[14-i]
		if !ehDigito(d) {
			continue
		}
		m := int(d-'0') * pesosIE[tp][i]
		soma += m
		// Rotacao A: alem do produto, soma a DEZENA dele -- e o que
		// transforma o modulo 10 de MG no "soma dos algarismos".
		if strings.Contains(rot, "A") {
			somaDezenas += m / 10
		}
	}

	switch {
	case strings.Contains(rot, "A"):
		soma += somaDezenas
	case strings.Contains(rot, "B"):
		soma *= 10
	case strings.Contains(rot, "C"):
		soma += 5 + 4*regra.fatorF
	}

	dv := soma % md
	// Rotacao E: o digito e o COMPLEMENTO do resto, nao o resto.
	if strings.Contains(rot, "E") {
		dv = md - dv
	}

	// Resto 10 e 11 nao cabem numa posicao; cada UF decide o que entra no
	// lugar. Só GO usa fatorG diferente de zero.
	switch dv {
	case 10:
		dv = regra.fatorG
	case 11:
		dv = regra.fatorF
	}
	return dv
}

// ufsValidasParaIE e o cUFsValidas do ACBr (ACBrValidador.pas:45).
//
// O "EX" entra na lista de propria conta do ACBr e passa na validacao de UF,
// mas nao tem regra de IE -- cai no erro de tamanho. Replicado.
var ufsValidasParaIE = map[string]bool{
	"AC": true, "AL": true, "AP": true, "AM": true, "BA": true, "CE": true,
	"DF": true, "ES": true, "GO": true, "MA": true, "MT": true, "MS": true,
	"MG": true, "PA": true, "PB": true, "PR": true, "PE": true, "PI": true,
	"RJ": true, "RN": true, "RS": true, "RO": true, "RR": true, "SC": true,
	"SP": true, "SE": true, "TO": true, "EX": true,
}

func ufValidaParaIE(uf string) bool { return ufsValidasParaIE[uf] }

// limparDocumento e o LimpaDocto do ACBr (ACBrValidador.pas:706): descarta
// ponto, barra e hifen, e maiusculiza o resto.
func limparDocumento(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToUpper(s) {
		if r == '.' || r == '/' || r == '-' {
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// ehDigito mora em chave_acesso.go.

func apenasDigitos(s string) bool {
	for i := 0; i < len(s); i++ {
		if !ehDigito(s[i]) {
			return false
		}
	}
	return true
}

// preencherEspacosEsquerda alinha a direita completando com espacos.
func preencherEspacosEsquerda(s string, tamanho int) string {
	if len(s) >= tamanho {
		return s
	}
	return strings.Repeat(" ", tamanho-len(s)) + s
}

// preencherZerosAteIE completa com zeros a esquerda, como o PadLeft que
// algumas UFs aplicam ao documento antes de validar.
func preencherZerosAteIE(s string, tamanho int) string {
	if len(s) >= tamanho {
		return s
	}
	return strings.Repeat("0", tamanho-len(s)) + s
}
