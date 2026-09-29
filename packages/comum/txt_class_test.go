package comum

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// TXTClass — Constructor defaults
// ---------------------------------------------------------------------------

func TestNewTXTClass_Defaults(t *testing.T) {
	txt := NewTXTClass()

	if txt.Delimitador != "|" {
		t.Errorf("Delimitador: got %q, want %q", txt.Delimitador, "|")
	}
	if !txt.TrimString {
		t.Error("TrimString should default to true")
	}
	if !txt.ReplaceDelimitador {
		t.Error("ReplaceDelimitador should default to true")
	}
	if txt.LinhasBuffer != 0 {
		t.Errorf("LinhasBuffer: got %d, want 0", txt.LinhasBuffer)
	}
	if txt.RegistroCount() != 0 {
		t.Errorf("RegistroCount: got %d, want 0", txt.RegistroCount())
	}
}

// ---------------------------------------------------------------------------
// LFillStr — Left-fill string field (SPED convention: |value)
// ---------------------------------------------------------------------------

func TestLFillStr_BasicValue(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("0000", 0, false, '0')
	want := "|0000"
	if got != want {
		t.Errorf("LFillStr basic: got %q, want %q", got, want)
	}
}

func TestLFillStr_LeftPadsWithZeros(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("42", 7, false, '0')
	want := "|0000042"
	if got != want {
		t.Errorf("LFillStr pad: got %q, want %q", got, want)
	}
}

func TestLFillStr_TruncatesWhenLonger(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("ABCDEF", 3, false, '0')
	want := "|ABC"
	if got != want {
		t.Errorf("LFillStr truncate: got %q, want %q", got, want)
	}
}

func TestLFillStr_NuloEmpty(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("", 5, true, '0')
	want := "|"
	if got != want {
		t.Errorf("LFillStr nulo: got %q, want %q", got, want)
	}
}

func TestLFillStr_TrimsSpaces(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("  SP  ", 0, false, '0')
	want := "|SP"
	if got != want {
		t.Errorf("LFillStr trim: got %q, want %q", got, want)
	}
}

func TestLFillStr_ReplacesDelimiter(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillStr("A|B|C", 0, false, '0')
	want := "|ABC"
	if got != want {
		t.Errorf("LFillStr replace delim: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// LFillInt — Integer field (SPED code like COD_MUN with leading zeros)
// ---------------------------------------------------------------------------

func TestLFillInt_CodMun_SevenDigits(t *testing.T) {
	txt := NewTXTClass()
	// IBGE code for Sao Paulo: 3550308
	got := txt.LFillInt(3550308, 7, false, '0')
	want := "|3550308"
	if got != want {
		t.Errorf("LFillInt CodMun: got %q, want %q", got, want)
	}
}

func TestLFillInt_SmallValuePadded(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillInt(1, 3, false, '0')
	want := "|001"
	if got != want {
		t.Errorf("LFillInt pad: got %q, want %q", got, want)
	}
}

func TestLFillInt_NuloZero(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillInt(0, 5, true, '0')
	want := "|"
	if got != want {
		t.Errorf("LFillInt nulo: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// LFillFloat — Monetary/numeric field with decimal places
// SPED uses comma as decimal separator (Brazilian convention)
// ---------------------------------------------------------------------------

func TestLFillFloat_MonetaryValue(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillFloat(1234.56, 0, 2, false, '0', "")
	want := "|1234,56"
	if got != want {
		t.Errorf("LFillFloat monetary: got %q, want %q", got, want)
	}
}

func TestLFillFloat_ZeroWithDecimal(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillFloat(0, 0, 2, false, '0', "")
	want := "|0,00"
	if got != want {
		t.Errorf("LFillFloat zero: got %q, want %q", got, want)
	}
}

func TestLFillFloat_NuloZero(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillFloat(0, 0, 2, true, '0', "")
	want := "|"
	if got != want {
		t.Errorf("LFillFloat nulo: got %q, want %q", got, want)
	}
}

func TestLFillFloat_FourDecimalPlaces(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillFloat(18.5, 0, 4, false, '0', "")
	want := "|18,5000"
	if got != want {
		t.Errorf("LFillFloat 4dp: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// LFillDate — Date field (SPED format: ddmmyyyy)
// ---------------------------------------------------------------------------

func TestLFillDate_SPEDFormat(t *testing.T) {
	txt := NewTXTClass()
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	got := txt.LFillDate(dt, "02012006", false)
	want := "|01092026"
	if got != want {
		t.Errorf("LFillDate: got %q, want %q", got, want)
	}
}

func TestLFillDate_NuloZeroTime(t *testing.T) {
	txt := NewTXTClass()
	got := txt.LFillDate(time.Time{}, "02012006", true)
	want := "|"
	if got != want {
		t.Errorf("LFillDate nulo: got %q, want %q", got, want)
	}
}

func TestLFillDate_DefaultMask(t *testing.T) {
	txt := NewTXTClass()
	dt := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	got := txt.LFillDate(dt, "", false)
	want := "|31122026"
	if got != want {
		t.Errorf("LFillDate default mask: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// RFill — Right-fill (used for some text fields)
// ---------------------------------------------------------------------------

func TestRFill_PadsRight(t *testing.T) {
	txt := NewTXTClass()
	got := txt.RFill("AB", 5, ' ')
	want := "|AB   "
	if got != want {
		t.Errorf("RFill pad: got %q, want %q", got, want)
	}
}

func TestRFill_TruncatesWhenLonger(t *testing.T) {
	txt := NewTXTClass()
	got := txt.RFill("ABCDEF", 3, ' ')
	want := "|ABC"
	if got != want {
		t.Errorf("RFill truncate: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// DFill — Decimal fill (simplified float, no padding, comma separator)
// ---------------------------------------------------------------------------

func TestDFill_TwoDecimalPlaces(t *testing.T) {
	txt := NewTXTClass()
	got := txt.DFill(100.50, 2, false)
	want := "|100,50"
	if got != want {
		t.Errorf("DFill 2dp: got %q, want %q", got, want)
	}
}

func TestDFill_NuloZero(t *testing.T) {
	txt := NewTXTClass()
	got := txt.DFill(0, 2, true)
	want := "|"
	if got != want {
		t.Errorf("DFill nulo: got %q, want %q", got, want)
	}
}

func TestDFill_NonNuloZero(t *testing.T) {
	txt := NewTXTClass()
	got := txt.DFill(0, 2, false)
	want := "|0,00"
	if got != want {
		t.Errorf("DFill non-nulo zero: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Add — Appends line with or without trailing delimiter
// ---------------------------------------------------------------------------

func TestAdd_WithDelimiter(t *testing.T) {
	txt := NewTXTClass()
	txt.Add("|0000|015|0|01092026", true)
	if len(txt.Conteudo) != 1 {
		t.Fatalf("expected 1 line, got %d", len(txt.Conteudo))
	}
	want := "|0000|015|0|01092026|"
	if txt.Conteudo[0] != want {
		t.Errorf("Add with delim: got %q, want %q", txt.Conteudo[0], want)
	}
}

func TestAdd_WithoutDelimiter(t *testing.T) {
	txt := NewTXTClass()
	txt.Add("|0000|015|0|01092026|", false)
	want := "|0000|015|0|01092026|"
	if txt.Conteudo[0] != want {
		t.Errorf("Add without delim: got %q, want %q", txt.Conteudo[0], want)
	}
}

func TestAdd_TrimsInput(t *testing.T) {
	txt := NewTXTClass()
	txt.Add("  |0000|015  ", true)
	want := "|0000|015|"
	if txt.Conteudo[0] != want {
		t.Errorf("Add trim: got %q, want %q", txt.Conteudo[0], want)
	}
}

func TestAdd_IncrementsRegistroCount(t *testing.T) {
	txt := NewTXTClass()
	txt.Add("line1", true)
	txt.Add("line2", true)
	txt.Add("line3", true)
	if txt.RegistroCount() != 3 {
		t.Errorf("RegistroCount: got %d, want 3", txt.RegistroCount())
	}
}

// ---------------------------------------------------------------------------
// WriteBuffer / SaveToFile / LoadFromFile
// ---------------------------------------------------------------------------

func TestWriteBuffer_CreatesFileWithCRLF(t *testing.T) {
	tmpDir := t.TempDir()
	txt := NewTXTClass()
	txt.NomeArquivo = filepath.Join(tmpDir, "test.txt")
	txt.Add("|0000|015|0|", false)
	txt.Add("|0001|0|", false)

	err := txt.WriteBuffer()
	if err != nil {
		t.Fatalf("WriteBuffer error: %v", err)
	}

	data, err := os.ReadFile(txt.NomeArquivo)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)
	// SPED files use CRLF line endings
	if !strings.Contains(content, "\r\n") {
		t.Error("SPED output should use CRLF line endings")
	}

	lines := strings.Split(strings.TrimRight(content, "\r\n"), "\r\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "|0000|015|0|" {
		t.Errorf("line 0: got %q", lines[0])
	}
	if lines[1] != "|0001|0|" {
		t.Errorf("line 1: got %q", lines[1])
	}
}

func TestWriteBuffer_ClearsBuffer(t *testing.T) {
	tmpDir := t.TempDir()
	txt := NewTXTClass()
	txt.NomeArquivo = filepath.Join(tmpDir, "test.txt")
	txt.Add("line1", false)
	_ = txt.WriteBuffer()
	if len(txt.Conteudo) != 0 {
		t.Error("Conteudo should be empty after WriteBuffer")
	}
}

func TestWriteBuffer_EmptyNomeArquivo(t *testing.T) {
	txt := NewTXTClass()
	txt.Add("line1", false)
	err := txt.WriteBuffer()
	if err != ErrArquivoNaoEspecificado {
		t.Errorf("expected ErrArquivoNaoEspecificado, got %v", err)
	}
}

func TestLoadFromFile_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	txt := NewTXTClass()
	txt.NomeArquivo = filepath.Join(tmpDir, "test.txt")
	txt.Add("|0000|015|", false)
	txt.Add("|9999|2|", false)
	_ = txt.SaveToFile()

	txt2 := NewTXTClass()
	txt2.NomeArquivo = txt.NomeArquivo
	err := txt2.LoadFromFile()
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if len(txt2.Conteudo) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(txt2.Conteudo))
	}
	if txt2.Conteudo[0] != "|0000|015|" {
		t.Errorf("line 0: got %q", txt2.Conteudo[0])
	}
}

func TestReset_ClearsContentAndRemovesFile(t *testing.T) {
	tmpDir := t.TempDir()
	txt := NewTXTClass()
	txt.NomeArquivo = filepath.Join(tmpDir, "test.txt")
	txt.Add("line1", false)
	_ = txt.SaveToFile()

	txt.Reset()
	if len(txt.Conteudo) != 0 {
		t.Error("Conteudo should be empty after Reset")
	}
	if txt.RegistroCount() != 0 {
		t.Error("RegistroCount should be 0 after Reset")
	}
	if _, err := os.Stat(txt.NomeArquivo); err == nil {
		t.Error("file should be removed after Reset")
	}
}

// ---------------------------------------------------------------------------
// Auto-flush (LinhasBuffer)
// ---------------------------------------------------------------------------

func TestAutoFlush_TriggersAtThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	txt := NewTXTClass()
	txt.NomeArquivo = filepath.Join(tmpDir, "test.txt")
	txt.LinhasBuffer = 3

	txt.Add("line1", false)
	txt.Add("line2", false)
	// Buffer has 2 lines, not flushed yet
	if _, err := os.Stat(txt.NomeArquivo); err == nil {
		t.Error("file should not exist yet (below threshold)")
	}

	txt.Add("line3", false) // Hits threshold of 3 → auto flush
	if _, err := os.Stat(txt.NomeArquivo); err != nil {
		t.Error("file should exist after auto flush")
	}
	if len(txt.Conteudo) != 0 {
		t.Error("buffer should be cleared after auto flush")
	}
}

// ---------------------------------------------------------------------------
// FormatFloatBR
// ---------------------------------------------------------------------------

func TestFormatFloatBR_CommaDecimalSeparator(t *testing.T) {
	tests := []struct {
		value float64
		mask  string
		want  string
	}{
		{1234.56, "0.00", "1234,56"},
		{0, "0.00", "0,00"},
		{100, "0.0000", "100,0000"},
		{99.999, "0.00", "100,00"}, // rounding
		{0.1, "0.00", "0,10"},
	}
	for _, tt := range tests {
		got := FormatFloatBR(tt.value, tt.mask)
		if got != tt.want {
			t.Errorf("FormatFloatBR(%v, %q) = %q, want %q", tt.value, tt.mask, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// FloatMask
// ---------------------------------------------------------------------------

func TestFloatMask(t *testing.T) {
	tests := []struct {
		decimal int
		useSep  bool
		want    string
	}{
		{2, false, "0.00"},
		{4, true, "#,##0.0000"},
		{0, false, "0"},
	}
	for _, tt := range tests {
		got := FloatMask(tt.decimal, tt.useSep)
		if got != tt.want {
			t.Errorf("FloatMask(%d, %v) = %q, want %q", tt.decimal, tt.useSep, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// ACBrError
// ---------------------------------------------------------------------------

func TestACBrError_ImplementsError(t *testing.T) {
	err := NewACBrError("test error")
	if err.Error() != "test error" {
		t.Errorf("ACBrError.Error() = %q, want %q", err.Error(), "test error")
	}
}

// ---------------------------------------------------------------------------
// Check (panic on failure)
// ---------------------------------------------------------------------------

func TestCheck_DoesNotPanicOnTrue(t *testing.T) {
	txt := NewTXTClass()
	txt.Check(true, "should not panic")
}

func TestCheck_PanicsOnFalse(t *testing.T) {
	txt := NewTXTClass()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		acbrErr, ok := r.(*ACBrError)
		if !ok {
			t.Fatalf("expected *ACBrError, got %T", r)
		}
		if acbrErr.Message != "validation failed" {
			t.Errorf("panic message: got %q, want %q", acbrErr.Message, "validation failed")
		}
	}()
	txt.Check(false, "validation failed")
}

// ---------------------------------------------------------------------------
// SPED pipe-delimited line assembly (integration-like test)
// ---------------------------------------------------------------------------

func TestAssembleSPEDLine_Registro0000(t *testing.T) {
	txt := NewTXTClass()
	dtIni := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	dtFin := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	// Simulate building a Registro 0000 line as the SPED writer does
	linha := txt.LFillStr("0000", 0, false, '0') +
		txt.LFillStr("016", 0, false, '0') + // CodVer 115 → "016"
		txt.LFillInt(0, 1, false, '0') + // CodFin Original
		txt.LFillDate(dtIni, "02012006", false) +
		txt.LFillDate(dtFin, "02012006", false) +
		txt.LFillStr("EMPRESA TESTE LTDA", 0, false, '0') +
		txt.LFillStr("12345678000199", 0, false, '0') + // CNPJ
		txt.LFillStr("", 0, true, '0') + // CPF (empty/nulo)
		txt.LFillStr("SP", 0, false, '0') +
		txt.LFillStr("123456789", 0, false, '0') + // IE
		txt.LFillInt(3550308, 7, false, '0') + // CodMun SP
		txt.LFillStr("", 0, true, '0') + // IM
		txt.LFillStr("", 0, true, '0') + // SUFRAMA
		txt.LFillStr("A", 0, false, '0') + // IndPerfil
		txt.LFillInt(1, 1, false, '0') // IndAtiv Outros

	txt.Add(linha, true)

	want := "|0000|016|0|01092026|30092026|EMPRESA TESTE LTDA|12345678000199||SP|123456789|3550308|||A|1|"
	if txt.Conteudo[0] != want {
		t.Errorf("Registro 0000 line:\ngot:  %q\nwant: %q", txt.Conteudo[0], want)
	}
}
