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

// Tabela das 27 UFs para a validacao de Inscricao Estadual.
//
// Porte de ACBrValidador.pas:1166-1458. Mantida na MESMA ORDEM do original,
// para que a conferencia linha a linha seja possivel.
//
// Os defaults, aplicados antes de qualquer UF (ACBrValidador.pas:1153-1161):
// xROT = "E", xMD = 11, xTP = 1, e nenhum segundo digito.
//
// Algumas UFs MUTAM o documento antes de validar (BA e MT completam com
// zeros) ou escolhem a regra pelo TAMANHO (AC, PE, RN, RO, SP, TO) ou pelo
// proprio VALOR (AP e GO comparam o documento como TEXTO contra faixas).
// Por isso a funcao recebe e devolve o documento.

// regraDaUF devolve a regra da UF e o documento eventualmente ajustado.
func regraDaUF(uf, doc string) (regraIE, string) {
	// Defaults do ACBr, validos para a maioria das UFs.
	r := regraIE{xROT: "E", xMD: 11, xTP: 1}

	switch uf {
	case "AC":
		if len(doc) == 9 {
			r.tamanho = 9
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "1", "0")
		} else {
			r.tamanho = 13
			r.xTP = 2
			r.yROT, r.yMD, r.yTP = "E", 11, 1
			r.digitos = digitosIE("DVY", "DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, "1", "0")
		}

	case "AL":
		r.tamanho = 9
		r.xROT = "BD"
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "4", "2")

	case "AP":
		r.tamanho = 9
		r.xROT = "CE"
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "3", "0")

		// Comparacao de TEXTO, nao de numero -- e assim no original, e so
		// funciona porque todos os documentos tem o mesmo tamanho aqui.
		if doc >= "030170010" && doc <= "030190229" {
			r.fatorF = 1
		} else if doc >= "030190230" {
			r.xROT = "E"
		}

	case "AM":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9)

	case "BA":
		if len(doc) < 9 {
			doc = preencherZerosAteIE(doc, 9)
		}
		r.tamanho = 9
		r.xTP, r.yTP, r.yROT = 2, 3, "E"
		r.digitos = digitosIE("DVX", "DVY", d9, d9, d9, d9, d9, d9, d9)

		// O modulo depende do SEGUNDO caractere da inscricao.
		if len(doc) >= 2 && contemByte("0123458", doc[1]) {
			r.xMD, r.yMD = 10, 10
		} else {
			r.xMD, r.yMD = 11, 11
		}

	case "CE":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, "0")

	case "DF":
		r.tamanho = 13
		r.xTP = 2
		r.yROT, r.yMD, r.yTP = "E", 11, 1
		r.digitos = digitosIE("DVY", "DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, "7,8", "0")

	case "ES":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9)

	case "GO":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, "1,2")

		// GO e a unica UF com uma regra de PREFIXO de dois digitos alem do
		// vetor de posicoes (ACBrValidador.pas:1261-1264). O vetor sozinho
		// so restringe o primeiro digito a 1 ou 2; sem esta verificacao os
		// prefixos 12 a 19 passariam.
		if prefixo := prefixoDeDoisDigitos(doc); !prefixoValidoGO(prefixo) {
			r.erroPrefixo = "prefixo deve ser 10, 11 ou de 20 a 29"
		}

		// A faixa 101031050..101199979 troca o substituto do resto 10 --
		// e a unica UF que usa fatorG.
		if doc >= "101031050" && doc <= "101199979" {
			r.fatorG = 1
		}

	case "MA":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "2", "1")

	case "MT":
		if len(doc) == 9 {
			doc = preencherZerosAteIE(doc, 11)
		}
		r.tamanho = 11
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, d9)

	case "MS":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "0,8", "2,5")

	case "MG":
		r.tamanho = 13
		r.xROT, r.xMD, r.xTP = "AE", 10, 10
		r.yROT, r.yMD, r.yTP = "E", 11, 11
		r.digitos = digitosIE("DVY", "DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, d9, d9)

	case "PA":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "5,6,7,8,9", "1,7")

	case "PB":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "6", "1")

	case "PR":
		r.tamanho = 10
		r.xTP = 9
		r.yROT, r.yMD, r.yTP = "E", 11, 8
		r.digitos = digitosIE("DVY", "DVX", d9, d9, d9, d9, d9, d9, d9, d9)

	case "PE":
		if len(doc) == 14 {
			// Antigo CACEPE.
			r.tamanho = 14
			r.xTP = 7
			r.fatorF = 1
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, d9, "1-9", "8", "1")
		} else {
			r.tamanho = 9
			r.xTP = 14
			r.yROT, r.yMD, r.yTP = "E", 11, 7
			r.digitos = digitosIE("DVY", "DVX", d9, d9, d9, d9, d9, d9, d9)
		}

	case "PI":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "9", "1")

	case "RJ":
		r.tamanho = 8
		r.xTP = 8
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "1,7,8,9")

	case "RN":
		if len(doc) == 9 {
			r.tamanho = 9
			r.xROT = "BD"
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "0", "2")
		} else {
			r.tamanho = 10
			r.xROT = "BD"
			r.xTP = 11
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, "0", "2")
		}

	case "RS":
		r.tamanho = 10
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9)

	case "RO":
		r.fatorF = 1
		if len(doc) == 9 {
			r.tamanho = 9
			r.xTP = 4
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, "1-9")
		} else {
			r.tamanho = 14
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9, d9, d9, d9, d9, d9)
		}

	case "RR":
		r.tamanho = 9
		r.xROT, r.xMD, r.xTP = "D", 9, 5
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "4", "2")

	case "SC", "SE":
		r.tamanho = 9
		r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9)

	case "SP":
		r.xROT, r.xTP = "D", 12
		if len(doc) > 0 && doc[0] == 'P' {
			// Inscricao de Produtor Rural: o DV fica no MEIO do documento,
			// nao na ponta, e o "P" ocupa a primeira posicao.
			r.tamanho = 13
			r.digitos = digitosIE(d9, d9, d9, "DVX", d9, d9, d9, d9, d9, d9, d9, d9, "P")
		} else {
			r.tamanho = 12
			r.yROT, r.yMD, r.yTP = "D", 11, 13
			r.digitos = digitosIE("DVY", d9, d9, "DVX", d9, d9, d9, d9, d9, d9, d9, d9)
		}

	case "TO":
		if len(doc) == 11 {
			r.tamanho = 11
			r.xTP = 6
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, "1,2,3,9", "0,9", "9", "2")
		} else {
			// O ACBr deixou a regra de 10 posicoes COMENTADA e usa a de 9.
			r.tamanho = 9
			r.digitos = digitosIE("DVX", d9, d9, d9, d9, d9, d9, d9, d9)
		}
	}

	return r, doc
}

// d9 encurta a referencia a "qualquer algarismo", que aparece mais de 200
// vezes na tabela acima.
const d9 = digito0a9

// digitosIE monta o vetor de 14 posicoes a partir das que a UF declara,
// completando o resto com "" (posicao inexistente).
//
// A ordem dos argumentos e a MESMA do VarArrayOf do original: da direita
// para a esquerda do documento.
func digitosIE(posicoes ...string) [14]string {
	var d [14]string
	for i, p := range posicoes {
		if i >= len(d) {
			break
		}
		d[i] = p
	}
	return d
}

// prefixoDeDoisDigitos devolve os dois primeiros caracteres como numero,
// reproduzindo o `StrToIntDef(fsDocto[1] + fsDocto[2], 0)` do original.
func prefixoDeDoisDigitos(doc string) int {
	if len(doc) < 2 {
		return 0
	}
	if !ehDigito(doc[0]) || !ehDigito(doc[1]) {
		return 0
	}
	return int(doc[0]-'0')*10 + int(doc[1]-'0')
}

// prefixoValidoGO e o conjunto [10, 11, 20..29] do ACBr.
func prefixoValidoGO(prefixo int) bool {
	return prefixo == 10 || prefixo == 11 || (prefixo >= 20 && prefixo <= 29)
}

// contemByte diz se o byte aparece na string.
func contemByte(s string, c byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return true
		}
	}
	return false
}
