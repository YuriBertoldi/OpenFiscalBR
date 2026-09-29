// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-28.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package comum

import (
	"fmt"
	"strings"
)

// FormatFloatBR formats a float64 value according to the given mask.
//
// The mask uses '#' for optional digits and '0' for required digits, with
// '.' as the decimal separator in the mask. The output uses ',' as the
// decimal separator (Brazilian convention) when the mask contains a dot,
// but for SPED files the caller typically passes masks that produce plain
// numeric strings (no thousands separator, dot as decimal).
//
// For SPED output the common pattern is: FormatFloatBR(value, "0.00")
// which yields e.g. "1234,56". The SPED writer later replaces ',' with
// nothing or handles it as needed.
//
// A simpler approach for SPED: format with the given number of decimal
// places and use comma as decimal separator.
func FormatFloatBR(value float64, mask string) string {
	// Count decimal places from mask.
	decimalPlaces := 0
	dotIdx := strings.LastIndex(mask, ".")
	if dotIdx >= 0 {
		decimalPlaces = len(mask) - dotIdx - 1
	}

	formatted := fmt.Sprintf("%.*f", decimalPlaces, value)
	// Replace '.' with ',' for Brazilian convention.
	formatted = strings.Replace(formatted, ".", ",", 1)
	return formatted
}

// FloatMask generates a format mask string with the given number of decimal
// places. If useSeparator is true, the integer part includes a thousands
// separator placeholder.
//
// Examples:
//
//	FloatMask(2, false) => "0.00"
//	FloatMask(4, true)  => "#,##0.0000"
//	FloatMask(0, false) => "0"
func FloatMask(decimal int, useSeparator bool) string {
	var b strings.Builder

	if useSeparator {
		b.WriteString("#,##0")
	} else {
		b.WriteString("0")
	}

	if decimal > 0 {
		b.WriteByte('.')
		for i := 0; i < decimal; i++ {
			b.WriteByte('0')
		}
	}

	return b.String()
}

// ACBrStr returns s unchanged. In the original Delphi code this function
// handled ANSI-to-UTF8 encoding conversion. In Go, strings are already
// UTF-8 so no conversion is needed.
func ACBrStr(s string) string {
	return s
}
