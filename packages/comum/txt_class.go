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
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// OnErrorFunc is the callback signature for TXTClass errors.
type OnErrorFunc func(msg string)

// TXTClass is the Go equivalent of TACBrTXTClass from ACBrTXTClass.pas.
// It is the core text-file generator used by all SPED components to build
// pipe-delimited register lines and write them to disk.
type TXTClass struct {
	// NomeArquivo is the output file path.
	NomeArquivo string

	// LinhasBuffer controls how many lines are buffered before an
	// automatic flush to disk. Zero means no automatic flush.
	LinhasBuffer int

	// Delimitador is the field delimiter. Default is "|" (pipe) for SPED.
	Delimitador string

	// TrimString controls whether values are trimmed before writing.
	TrimString bool

	// CurMascara holds the current float format mask.
	CurMascara string

	// ReplaceDelimitador, when true, removes occurrences of the
	// delimiter character from field values to avoid corruption.
	ReplaceDelimitador bool

	// Conteudo holds the in-memory lines that have not yet been flushed.
	Conteudo []string

	// OnError is an optional callback invoked when an error occurs.
	OnError OnErrorFunc

	// registroCount tracks lines written (including flushed ones).
	registroCount int
}

// NewTXTClass creates a new TXTClass with sensible defaults matching the
// Delphi constructor.
func NewTXTClass() *TXTClass {
	return &TXTClass{
		Delimitador:        "|",
		TrimString:         true,
		ReplaceDelimitador: true,
		LinhasBuffer:       0,
		Conteudo:           make([]string, 0, 256),
	}
}

// --------------------------------------------------------------------------
// Content management
// --------------------------------------------------------------------------

// Add appends a line to Conteudo. When addDelimiter is true the trailing
// delimiter is appended automatically (SPED convention).
func (t *TXTClass) Add(s string, addDelimiter bool) {
	if t.TrimString {
		s = strings.TrimSpace(s)
	}
	if addDelimiter {
		s += t.Delimitador
	}
	t.Conteudo = append(t.Conteudo, s)
	t.registroCount++

	// Auto-flush when LinhasBuffer is set and threshold is reached.
	if t.LinhasBuffer > 0 && len(t.Conteudo) >= t.LinhasBuffer {
		_ = t.WriteBuffer()
	}
}

// WriteBuffer appends the current Conteudo to the file on disk and clears
// the in-memory buffer. It creates the file if it does not exist.
func (t *TXTClass) WriteBuffer() error {
	if t.NomeArquivo == "" {
		return ErrArquivoNaoEspecificado
	}
	if len(t.Conteudo) == 0 {
		return nil
	}

	f, err := os.OpenFile(t.NomeArquivo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.fireError(err.Error())
		return err
	}
	defer f.Close()

	data := strings.Join(t.Conteudo, "\r\n") + "\r\n"
	if _, err := f.WriteString(data); err != nil {
		t.fireError(err.Error())
		return err
	}

	t.Conteudo = t.Conteudo[:0]
	return nil
}

// SaveToFile flushes the remaining buffer and ensures everything is on disk.
func (t *TXTClass) SaveToFile() error {
	return t.WriteBuffer()
}

// LoadFromFile reads the entire file into Conteudo, splitting on newlines.
func (t *TXTClass) LoadFromFile() error {
	if t.NomeArquivo == "" {
		return ErrArquivoNaoEspecificado
	}

	data, err := os.ReadFile(t.NomeArquivo)
	if err != nil {
		t.fireError(err.Error())
		return err
	}

	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.TrimRight(content, "\n")
	if content == "" {
		t.Conteudo = t.Conteudo[:0]
	} else {
		t.Conteudo = strings.Split(content, "\n")
	}
	t.registroCount = len(t.Conteudo)
	return nil
}

// Reset clears the in-memory content and removes the file from disk (if it
// exists).
func (t *TXTClass) Reset() {
	t.Conteudo = t.Conteudo[:0]
	t.registroCount = 0
	if t.NomeArquivo != "" {
		_ = os.Remove(t.NomeArquivo)
	}
}

// RegistroCount returns the total number of lines added (including those
// already flushed to disk).
func (t *TXTClass) RegistroCount() int {
	return t.registroCount
}

// --------------------------------------------------------------------------
// Field-formatting helpers  (LFill / RFill / DFill)
// --------------------------------------------------------------------------

// LFillStr returns Delimitador + value left-padded with char to size.
// If nulo is true and value is empty, only the delimiter is returned.
func (t *TXTClass) LFillStr(value string, size int, nulo bool, char byte) string {
	if t.TrimString {
		value = strings.TrimSpace(value)
	}
	if t.ReplaceDelimitador && t.Delimitador != "" {
		value = strings.ReplaceAll(value, t.Delimitador, "")
	}

	if nulo && value == "" {
		return t.Delimitador
	}

	if size > 0 && len(value) > size {
		value = value[:size]
	}

	for len(value) < size {
		value = string(char) + value
	}

	return t.Delimitador + value
}

// LFillInt returns Delimitador + the integer value left-padded with char
// to size. If nulo is true and value is 0, only the delimiter is returned.
func (t *TXTClass) LFillInt(value int64, size int, nulo bool, char byte) string {
	if nulo && value == 0 {
		return t.Delimitador
	}
	s := strconv.FormatInt(value, 10)
	return t.LFillStr(s, size, false, char)
}

// LFillFloat returns Delimitador + the float value formatted and
// left-padded. The float is formatted with the given number of decimal
// places. If mascara is not empty it is used as the format mask; otherwise
// a default mask is generated from decimal.
//
// For SPED the decimal separator in the output is a comma, but the
// resulting field uses no thousands separator.
func (t *TXTClass) LFillFloat(value float64, size int, decimal int, nulo bool, char byte, mascara string) string {
	if nulo && value == 0 {
		return t.Delimitador
	}

	if mascara == "" {
		mascara = FloatMask(decimal, false)
	}
	t.CurMascara = mascara

	s := FormatFloatBR(value, mascara)
	return t.LFillStr(s, size, false, char)
}

// LFillDate returns Delimitador + the date formatted with the given mask.
// If nulo is true and value is the zero time, only the delimiter is
// returned. Common SPED mask: "02012006" (ddmmyyyy).
func (t *TXTClass) LFillDate(value time.Time, mask string, nulo bool) string {
	if nulo && value.IsZero() {
		return t.Delimitador
	}
	if mask == "" {
		mask = "02012006" // ddmmyyyy Go reference layout
	}
	s := value.Format(mask)
	return t.Delimitador + s
}

// RFill returns Delimitador + value right-padded with char to size.
func (t *TXTClass) RFill(value string, size int, char byte) string {
	if t.TrimString {
		value = strings.TrimSpace(value)
	}
	if t.ReplaceDelimitador && t.Delimitador != "" {
		value = strings.ReplaceAll(value, t.Delimitador, "")
	}

	if len(value) > size && size > 0 {
		value = value[:size]
	}

	for len(value) < size {
		value = value + string(char)
	}

	return t.Delimitador + value
}

// DFill returns Delimitador + the float value formatted with the given
// number of decimal places. If nulo is true and value is 0, only the
// delimiter is returned.
func (t *TXTClass) DFill(value float64, decimal int, nulo bool) string {
	if nulo && value == 0 {
		return t.Delimitador
	}

	s := formatDecimal(value, decimal)
	return t.Delimitador + s
}

// --------------------------------------------------------------------------
// Validation
// --------------------------------------------------------------------------

// Check validates a condition. If condition is false, the error callback is
// fired with msg and a panic with ACBrError is raised — mirroring the
// Delphi behaviour of raising EACBrException.
func (t *TXTClass) Check(condition bool, msg string) {
	if !condition {
		t.fireError(msg)
		panic(NewACBrError(msg))
	}
}

// --------------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------------

func (t *TXTClass) fireError(msg string) {
	if t.OnError != nil {
		t.OnError(msg)
	}
}

// formatDecimal formats a float64 with exactly decimal places, using comma
// as the decimal separator (Brazilian SPED convention).
func formatDecimal(value float64, decimal int) string {
	// Round to the requested precision to avoid floating-point artefacts.
	shift := math.Pow(10, float64(decimal))
	rounded := math.Round(value*shift) / shift

	s := fmt.Sprintf("%.*f", decimal, rounded)
	s = strings.Replace(s, ".", ",", 1)
	return s
}
