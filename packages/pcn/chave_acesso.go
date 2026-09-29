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
	"strconv"
	"strings"
)

// Porte dos utilitarios de chave de acesso de ACBrDFeUtil.pas.

// pesoChaveAcesso e a mascara de pesos do digito verificador da chave.
// Manual de Integracao do Contribuinte v2.02a, pagina 70.
const pesoChaveAcesso = "4329876543298765432987654329876543298765432"

// RemoverLiteralChave remove o literal que precede a chave de acesso
// ("NFGas", "NFe", "CTe", "ID"...), devolvendo a chave a partir do primeiro
// par de caracteres numericos consecutivos.
// Porte de RemoverLiteralChave (ACBrDFeUtil).
func RemoverLiteralChave(chave string) string {
	for i := 0; i < len(chave)-1; i++ {
		if ehDigito(chave[i]) && ehDigito(chave[i+1]) {
			return chave[i:]
		}
	}
	return ""
}

// DigitoChaveAcesso calcula o digito verificador das 43 primeiras posicoes
// da chave de acesso. Porte de GerarDigito (ACBrDFeUtil).
//
// O valor de cada caractere e o codigo ASCII menos 48, o que faz a conta
// funcionar tanto para a chave inteiramente numerica quanto para a que
// carrega CNPJ alfanumerico.
func DigitoChaveAcesso(chave43 string) (int, error) {
	c := RemoverLiteralChave(chave43)
	if c == "" {
		c = chave43
	}
	if len(c) > 43 {
		c = c[:43]
	}
	if len(c) != 43 {
		return 0, fmt.Errorf("%w: base do digito tem %d caracteres, esperado 43", ErrChaveInvalida, len(c))
	}

	c = strings.ToUpper(c)
	soma := 0
	for i := 0; i < 43; i++ {
		peso := int(pesoChaveAcesso[i] - '0')
		soma += (int(c[i]) - 48) * peso
	}

	resto := soma % 11
	if resto < 2 {
		return 0, nil
	}
	return 11 - resto, nil
}

// ValidarChaveAcesso confere o tamanho, o digito verificador, o codigo da UF
// e o AAMM de uma chave de acesso de 44 posicoes.
// Porte de ValidarChave (ACBrDFeUtil).
func ValidarChaveAcesso(chave string) error {
	c := RemoverLiteralChave(chave)
	if c == "" {
		c = chave
	}
	if len(c) != 44 {
		return fmt.Errorf("%w: chave tem %d caracteres, esperado 44", ErrChaveInvalida, len(c))
	}

	dv, err := DigitoChaveAcesso(c[:43])
	if err != nil {
		return err
	}
	informado, err := strconv.Atoi(c[43:])
	if err != nil {
		return fmt.Errorf("%w: digito verificador %q nao e numerico", ErrChaveInvalida, c[43:])
	}
	if dv != informado {
		return fmt.Errorf("%w: digito verificador %d, esperado %d", ErrChaveInvalida, informado, dv)
	}

	cUF, err := strconv.Atoi(c[0:2])
	if err != nil || !ValidarCodigoUF(cUF) {
		return fmt.Errorf("%w: codigo de UF %q invalido", ErrChaveInvalida, c[0:2])
	}
	if err := ValidarAAMM(c[2:6]); err != nil {
		return err
	}
	return nil
}

// ValidarAAMM confere o trecho de ano e mes de uma chave de acesso.
// Porte de ValidarAAMM (ACBrDFeUtil).
func ValidarAAMM(aamm string) error {
	if len(aamm) != 4 {
		return fmt.Errorf("%w: AAMM %q nao tem 4 digitos", ErrChaveInvalida, aamm)
	}
	ano, err := strconv.Atoi(aamm[:2])
	if err != nil {
		return fmt.Errorf("%w: ano %q nao e numerico", ErrChaveInvalida, aamm[:2])
	}
	mes, err := strconv.Atoi(aamm[2:])
	if err != nil {
		return fmt.Errorf("%w: mes %q nao e numerico", ErrChaveInvalida, aamm[2:])
	}
	if ano < 0 || ano > 99 || mes < 1 || mes > 12 {
		return fmt.Errorf("%w: AAMM %q fora de faixa", ErrChaveInvalida, aamm)
	}
	return nil
}

// FormatarInteiroZeros devolve o inteiro com zeros a esquerda no tamanho
// informado. Porte de IntToStrZero (ACBrUtil), usado na montagem da chave.
func FormatarInteiroZeros(valor, tamanho int) string {
	return fmt.Sprintf("%0*d", tamanho, valor)
}

// PreencherZerosEsquerda completa a string com zeros a esquerda ate o
// tamanho informado. Se ja for maior, devolve as ultimas posicoes.
func PreencherZerosEsquerda(s string, tamanho int) string {
	if len(s) >= tamanho {
		return s[len(s)-tamanho:]
	}
	return strings.Repeat("0", tamanho-len(s)) + s
}

func ehDigito(c byte) bool { return c >= '0' && c <= '9' }
