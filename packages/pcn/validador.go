// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-29.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package pcn

import (
	"fmt"
	"strings"
)

// Porte das validacoes de documento de ACBrValidador.pas.

// OnlyCPFCNPJAlphaNum devolve apenas os caracteres validos de um CPF/CNPJ,
// preservando letras maiusculas -- necessario desde o CNPJ alfanumerico.
// Porte de OnlyCPFCNPJAlphaNum (ACBrValidador).
func OnlyCPFCNPJAlphaNum(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToUpper(s) {
		switch {
		case r >= '0' && r <= '9', r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ValidarCPF confere os dois digitos verificadores de um CPF.
// Devolve ErrDocumentoInvalido quando o documento nao passa.
func ValidarCPF(cpf string) error {
	s := OnlyNumber(cpf)
	if len(s) != 11 {
		return fmt.Errorf("%w: CPF %q nao tem 11 digitos", ErrDocumentoInvalido, cpf)
	}
	if todosIguais(s) {
		return fmt.Errorf("%w: CPF %q tem todos os digitos iguais", ErrDocumentoInvalido, cpf)
	}

	d1 := digitoModulo11(s[:9], 10)
	d2 := digitoModulo11(s[:10], 11)
	if int(s[9]-'0') != d1 || int(s[10]-'0') != d2 {
		return fmt.Errorf("%w: CPF %q com digito verificador incorreto", ErrDocumentoInvalido, cpf)
	}
	return nil
}

// ValidarCNPJ confere os dois digitos verificadores de um CNPJ.
//
// Aceita o CNPJ alfanumerico: cada caractere da base vale o codigo ASCII
// menos 48 -- digitos valem 0 a 9 e letras de 'A' a 'Z' valem 17 a 42. Os
// dois digitos verificadores continuam sendo sempre numericos. E a mesma
// conta que o ACBr usa em GerarDigito da chave de acesso.
func ValidarCNPJ(cnpj string) error {
	s := OnlyCPFCNPJAlphaNum(cnpj)
	if len(s) != 14 {
		return fmt.Errorf("%w: CNPJ %q nao tem 14 caracteres", ErrDocumentoInvalido, cnpj)
	}
	if todosIguais(s) {
		return fmt.Errorf("%w: CNPJ %q tem todos os caracteres iguais", ErrDocumentoInvalido, cnpj)
	}
	if s[12] < '0' || s[12] > '9' || s[13] < '0' || s[13] > '9' {
		return fmt.Errorf("%w: CNPJ %q com digito verificador nao numerico", ErrDocumentoInvalido, cnpj)
	}

	d1 := digitoCNPJ(s[:12])
	d2 := digitoCNPJ(s[:13])
	if int(s[12]-'0') != d1 || int(s[13]-'0') != d2 {
		return fmt.Errorf("%w: CNPJ %q com digito verificador incorreto", ErrDocumentoInvalido, cnpj)
	}
	return nil
}

// ValidarCNPJouCPF escolhe a validacao pelo tamanho do documento: 11
// caracteres valem CPF, 14 valem CNPJ. Documento vazio nao e erro -- muitos
// campos de DFe sao opcionais.
func ValidarCNPJouCPF(doc string) error {
	s := OnlyCPFCNPJAlphaNum(doc)
	switch len(s) {
	case 0:
		return nil
	case 11:
		return ValidarCPF(s)
	case 14:
		return ValidarCNPJ(s)
	default:
		return fmt.Errorf("%w: documento %q nao tem 11 nem 14 caracteres", ErrDocumentoInvalido, doc)
	}
}

// digitoModulo11 calcula o digito verificador de CPF: pesos decrescentes a
// partir de pesoInicial, resto 0 ou 1 resulta em digito 0.
func digitoModulo11(base string, pesoInicial int) int {
	soma := 0
	peso := pesoInicial
	for i := 0; i < len(base); i++ {
		soma += int(base[i]-'0') * peso
		peso--
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

// digitoCNPJ calcula o digito verificador de CNPJ com os pesos ciclicos
// 2..9, da direita para a esquerda.
func digitoCNPJ(base string) int {
	soma := 0
	peso := 2
	for i := len(base) - 1; i >= 0; i-- {
		soma += valorCaractereCNPJ(base[i]) * peso
		peso++
		if peso > 9 {
			peso = 2
		}
	}
	resto := soma % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

// valorCaractereCNPJ devolve o valor numerico de um caractere do CNPJ:
// codigo ASCII menos 48, conforme a regra do CNPJ alfanumerico.
func valorCaractereCNPJ(c byte) int {
	return int(c) - 48
}

func todosIguais(s string) bool {
	if s == "" {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

// ValidarCodigoUF confere se o codigo IBGE de UF existe.
// Porte de ValidarCodigoUF (ACBrDFeUtil).
func ValidarCodigoUF(cUF int) bool {
	_, ok := siglaPorCodigoUF[cUF]
	return ok
}

// SiglaUF devolve a sigla da UF a partir do codigo IBGE, ou vazio.
func SiglaUF(cUF int) string {
	return siglaPorCodigoUF[cUF]
}

// CodigoUF devolve o codigo IBGE a partir da sigla da UF, ou 0.
func CodigoUF(sigla string) int {
	sigla = strings.ToUpper(strings.TrimSpace(sigla))
	for cod, s := range siglaPorCodigoUF {
		if s == sigla {
			return cod
		}
	}
	return 0
}

var siglaPorCodigoUF = map[int]string{
	11: "RO", 12: "AC", 13: "AM", 14: "RR", 15: "PA", 16: "AP", 17: "TO",
	21: "MA", 22: "PI", 23: "CE", 24: "RN", 25: "PB", 26: "PE", 27: "AL", 28: "SE", 29: "BA",
	31: "MG", 32: "ES", 33: "RJ", 35: "SP",
	41: "PR", 42: "SC", 43: "RS",
	50: "MS", 51: "MT", 52: "GO", 53: "DF",
	// 91 e o codigo do "Ambiente Nacional" usado por alguns servicos.
	91: "AN",
}
