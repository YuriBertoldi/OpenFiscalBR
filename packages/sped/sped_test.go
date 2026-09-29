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

package sped

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Enum String() tests — validate SPED output codes per legislation
// ---------------------------------------------------------------------------

func TestVersaoLeiauteFiscal_String(t *testing.T) {
	tests := []struct {
		v    VersaoLeiauteFiscal
		want string
	}{
		{VlVersao100, "001"}, // Versao 100 → codigo "001"
		{VlVersao115, "016"}, // Versao 115 → codigo "016"
		{VlVersao119, "020"}, // Versao 119 → codigo "020"
	}
	for _, tt := range tests {
		got := tt.v.String()
		if got != tt.want {
			t.Errorf("VersaoLeiauteFiscal(%d).String() = %q, want %q", int(tt.v), got, tt.want)
		}
	}
}

func TestCodFin_String(t *testing.T) {
	if CodFinOriginal.String() != "0" {
		t.Errorf("CodFinOriginal = %q, want %q", CodFinOriginal.String(), "0")
	}
	if CodFinSubstituto.String() != "1" {
		t.Errorf("CodFinSubstituto = %q, want %q", CodFinSubstituto.String(), "1")
	}
}

func TestIndPerfil_String(t *testing.T) {
	tests := map[IndPerfil]string{
		PerfilA: "A", PerfilB: "B", PerfilC: "C",
	}
	for v, want := range tests {
		if v.String() != want {
			t.Errorf("IndPerfil(%d) = %q, want %q", int(v), v.String(), want)
		}
	}
}

func TestIndAtiv_String(t *testing.T) {
	if AtivIndustrial.String() != "0" {
		t.Error("AtivIndustrial should be 0")
	}
	if AtivOutros.String() != "1" {
		t.Error("AtivOutros should be 1")
	}
}

func TestCodSit_String_TwoDigits(t *testing.T) {
	tests := map[CodSit]string{
		SitRegular:     "00",
		SitCancelado:   "02",
		SitDenegado:    "04",
		SitFiscalCompl: "06",
	}
	for v, want := range tests {
		if v.String() != want {
			t.Errorf("CodSit(%d) = %q, want %q", int(v), v.String(), want)
		}
	}
}

func TestTipoItem_String_TwoDigits(t *testing.T) {
	if TiMercadoriaRevenda.String() != "00" {
		t.Errorf("TiMercadoriaRevenda = %q, want 00", TiMercadoriaRevenda.String())
	}
	if TiOutras.String() != "99" {
		t.Errorf("TiOutras = %q, want 99", TiOutras.String())
	}
}

func TestIndPgto_String(t *testing.T) {
	if PgtoVista.String() != "0" {
		t.Error("PgtoVista should be 0")
	}
	if PgtoSemPagamento.String() != "9" {
		t.Error("PgtoSemPagamento should be 9")
	}
	if PgtoNenhum.String() != "" {
		t.Error("PgtoNenhum should be empty")
	}
}

func TestIndFrt_String(t *testing.T) {
	if FrtContaEmitente.String() != "0" {
		t.Error("FrtContaEmitente should be 0")
	}
	if FrtSemCobranca.String() != "9" {
		t.Error("FrtSemCobranca should be 9")
	}
	if FrtNenhum.String() != "" {
		t.Error("FrtNenhum should be empty")
	}
}

func TestIndCTA_String(t *testing.T) {
	if CTASintetica.String() != "S" {
		t.Error("CTASintetica should be S")
	}
	if CTAAnalitica.String() != "A" {
		t.Error("CTAAnalitica should be A")
	}
}

func TestMotInv_String_TwoDigits(t *testing.T) {
	if MotInvFinalPeriodo.String() != "01" {
		t.Errorf("MotInvFinalPeriodo = %q, want 01", MotInvFinalPeriodo.String())
	}
	if MotInvControleMercadoriaST.String() != "06" {
		t.Errorf("MotInvControleMercadoriaST = %q, want 06", MotInvControleMercadoriaST.String())
	}
}

func TestNaturezaConta_String(t *testing.T) {
	if NatContaAtivo.String() != "01" {
		t.Errorf("NatContaAtivo = %q, want 01", NatContaAtivo.String())
	}
	if NatContaOutras.String() != "09" {
		t.Errorf("NatContaOutras = %q, want 09", NatContaOutras.String())
	}
}

// ---------------------------------------------------------------------------
// CST ICMS tests — validate specific codes per Tabela B do Anexo do RICMS
// ---------------------------------------------------------------------------

func TestCstIcms_String(t *testing.T) {
	tests := []struct {
		cst  CstIcms
		want string
	}{
		{CstIcmsNenhum, ""},
		{CstIcmsTributadaIntegralmente, "000"},
		{CstIcmsTributadaComCobracaPorST, "010"},
		{CstIcmsComReducao, "020"},
		{CstIcmsIsenta, "040"},
		{CstIcmsNaoTributada, "041"},
		{CstIcmsSuspensao, "050"},
		{CstIcmsCobradoAnteriormentePorST, "060"},
		{CstIcmsOutros, "090"},
		{CstIcmsSimplesNacionalTributadaComPermissaoCredito, "101"},
		{CstIcmsSimplesNacionalOutros, "900"},
	}
	for _, tt := range tests {
		got := tt.cst.String()
		if got != tt.want {
			t.Errorf("CstIcms(%d).String() = %q, want %q", int(tt.cst), got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// CST PIS/COFINS tests
// ---------------------------------------------------------------------------

func TestCstPis_String(t *testing.T) {
	tests := []struct {
		cst  CstPis
		want string
	}{
		{CstPisValorAliquotaNormal, "01"},
		{CstPisAliquotaZero, "06"},
		{CstPisOutrasOperacoesSaida, "49"},
		{CstPisOperCredExcRecTribMercInt, "50"},
		{CstPisOutrasOperacoes, "99"},
		{CstPisNenhum, ""},
	}
	for _, tt := range tests {
		got := tt.cst.String()
		if got != tt.want {
			t.Errorf("CstPis(%d).String() = %q, want %q", int(tt.cst), got, tt.want)
		}
	}
}

func TestCstCofins_String(t *testing.T) {
	if CstCofinsValorAliquotaNormal.String() != "01" {
		t.Errorf("CstCofinsValorAliquotaNormal = %q, want 01", CstCofinsValorAliquotaNormal.String())
	}
	if CstCofinsOutrasOperacoes.String() != "99" {
		t.Errorf("CstCofinsOutrasOperacoes = %q, want 99", CstCofinsOutrasOperacoes.String())
	}
}

func TestCstIpi_String(t *testing.T) {
	tests := []struct {
		cst  CstIpi
		want string
	}{
		{CstIpiEntradaRecuperacaoCredito, "00"},
		{CstIpiSaidaTributada, "50"},
		{CstIpiOutrasSaidas, "99"},
		{CstIpiVazio, ""},
	}
	for _, tt := range tests {
		got := tt.cst.String()
		if got != tt.want {
			t.Errorf("CstIpi(%d).String() = %q, want %q", int(tt.cst), got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — Constructor and initialization
// ---------------------------------------------------------------------------

func TestNewSPEDFiscal_DefaultsSet(t *testing.T) {
	f := NewSPEDFiscal()
	if f.Delimitador != "|" {
		t.Errorf("Delimitador = %q, want %q", f.Delimitador, "|")
	}
	if f.CurMascara != "#0.00" {
		t.Errorf("CurMascara = %q, want %q", f.CurMascara, "#0.00")
	}
	if f.Bloco0 == nil || f.BlocoB == nil || f.BlocoC == nil || f.BlocoD == nil ||
		f.BlocoE == nil || f.BlocoG == nil || f.BlocoH == nil || f.BlocoK == nil ||
		f.Bloco1 == nil || f.Bloco9 == nil {
		t.Fatal("all block pointers must be non-nil")
	}
}

func TestNewSPEDFiscal_Registro0000_Allocated(t *testing.T) {
	f := NewSPEDFiscal()
	if f.Bloco0.Registro0000 == nil {
		t.Fatal("Registro0000 should be allocated")
	}
}

func TestNewSPEDFiscal_BlocksHaveClosingRegisters(t *testing.T) {
	f := NewSPEDFiscal()
	if f.Bloco0.Registro0990 == nil {
		t.Error("Bloco0.Registro0990 should be allocated")
	}
	if f.BlocoC.RegistroC990 == nil {
		t.Error("BlocoC.RegistroC990 should be allocated")
	}
	if f.Bloco9.Registro9990 == nil {
		t.Error("Bloco9.Registro9990 should be allocated")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — SetDelimitador/SetDtIni propagation
// ---------------------------------------------------------------------------

func TestSetDelimitador_PropagatesAllBlocks(t *testing.T) {
	f := NewSPEDFiscal()
	f.SetDelimitador(";")
	for _, s := range f.allBlocoSPEDs() {
		if s.Delimitador != ";" {
			t.Error("delimiter not propagated to all blocks")
		}
	}
}

func TestSetDtIni_PropagatesAndSetsRegistro0000(t *testing.T) {
	f := NewSPEDFiscal()
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.SetDtIni(dt)

	if f.DtIni != dt {
		t.Error("DtIni not set on SPEDFiscal")
	}
	if f.Bloco0.Registro0000.DtIni != dt {
		t.Error("DtIni not set on Registro0000")
	}
	if f.Bloco0.DtIni != dt {
		t.Error("DtIni not set on Bloco0.SPED")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — IniciaGeracao validations (per legislation)
// ---------------------------------------------------------------------------

func TestIniciaGeracao_DtIniMustBeFirstDay(t *testing.T) {
	f := NewSPEDFiscal()
	f.DtIni = time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC) // NOT 1st
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	err := f.IniciaGeracao()
	if err == nil {
		t.Fatal("expected error: DT_INI must be first day of month")
	}
	if !strings.Contains(err.Error(), "primeiro dia") {
		t.Errorf("error msg = %q, want mention of primeiro dia", err.Error())
	}
}

func TestIniciaGeracao_DtFinMustBeLastDay(t *testing.T) {
	f := NewSPEDFiscal()
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC) // NOT 30
	err := f.IniciaGeracao()
	if err == nil {
		t.Fatal("expected error: DT_FIN must be last day of month")
	}
	if !strings.Contains(err.Error(), "ultimo dia") {
		t.Errorf("error msg = %q, want mention of ultimo dia", err.Error())
	}
}

func TestIniciaGeracao_DtFinBeforeDtIni(t *testing.T) {
	f := NewSPEDFiscal()
	f.DtIni = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	err := f.IniciaGeracao()
	if err == nil {
		t.Fatal("expected error: DT_FIN before DT_INI")
	}
}

func TestIniciaGeracao_February29(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	// 2024 is a leap year
	f.DtIni = time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	err := f.IniciaGeracao()
	if err != nil {
		t.Fatalf("Feb 29 on leap year should be valid: %v", err)
	}
}

func TestIniciaGeracao_DefaultFilename(t *testing.T) {
	f := NewSPEDFiscal()
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	f.Path = t.TempDir()
	err := f.IniciaGeracao()
	if err != nil {
		t.Fatalf("IniciaGeracao error: %v", err)
	}
	if !strings.Contains(f.NomeArquivo, "SpedFiscal.txt") {
		t.Errorf("default filename should be SpedFiscal.txt, got %q", f.NomeArquivo)
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — SaveFileTXT end-to-end (minimal file structure)
// ---------------------------------------------------------------------------

func TestSaveFileTXT_MinimalFile(t *testing.T) {
	f := NewSPEDFiscal()
	tmpDir := t.TempDir()
	f.Path = tmpDir
	f.Arquivo = "sped_test.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	// Fill Registro0000 (mandatory)
	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "EMPRESA TESTE LTDA"
	r.CNPJ = "12345678000199"
	r.UF = "SP"
	r.IE = "123456789"
	r.CodMun = 3550308
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	// Mark block 0 as having data
	f.Bloco0.Registro0001.IndMov = 0

	err := f.SaveFileTXT()
	if err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}

	// Read output
	outPath := filepath.Join(tmpDir, "sped_test.txt")
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)
	lines := strings.Split(strings.TrimRight(content, "\r\n"), "\r\n")

	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}

	// --- Validate line structure per SPED Fiscal Guia Pratico ---

	// Line must start and end with pipe delimiter
	for i, line := range lines {
		if !strings.HasPrefix(line, "|") {
			t.Errorf("line %d: must start with |: %q", i, line)
		}
		if !strings.HasSuffix(line, "|") {
			t.Errorf("line %d: must end with |: %q", i, line)
		}
	}

	// First line must be Registro 0000
	if !strings.HasPrefix(lines[0], "|0000|") {
		t.Errorf("first line must be |0000|..., got %q", lines[0])
	}

	// Validate Registro 0000 fields
	fields0000 := strings.Split(lines[0], "|")
	// fields0000[0] is empty (before first |), [1]="0000", [2]=CodVer, etc.
	if len(fields0000) < 16 {
		t.Fatalf("Registro 0000: expected 16+ fields, got %d", len(fields0000))
	}
	if fields0000[1] != "0000" {
		t.Errorf("REG = %q, want 0000", fields0000[1])
	}
	if fields0000[2] != "016" { // VlVersao115 → "016"
		t.Errorf("COD_VER = %q, want 016", fields0000[2])
	}
	if fields0000[3] != "0" { // CodFinOriginal
		t.Errorf("COD_FIN = %q, want 0", fields0000[3])
	}
	if fields0000[4] != "01092026" { // DT_INI ddmmyyyy
		t.Errorf("DT_INI = %q, want 01092026", fields0000[4])
	}
	if fields0000[5] != "30092026" { // DT_FIN
		t.Errorf("DT_FIN = %q, want 30092026", fields0000[5])
	}
	if fields0000[6] != "EMPRESA TESTE LTDA" {
		t.Errorf("NOME = %q", fields0000[6])
	}
	if fields0000[7] != "12345678000199" {
		t.Errorf("CNPJ = %q", fields0000[7])
	}
	if fields0000[9] != "SP" {
		t.Errorf("UF = %q, want SP", fields0000[9])
	}
	if fields0000[11] != "3550308" {
		t.Errorf("COD_MUN = %q, want 3550308", fields0000[11])
	}
	if fields0000[14] != "A" { // IndPerfil
		t.Errorf("IND_PERFIL = %q, want A", fields0000[14])
	}
	if fields0000[15] != "1" { // IndAtiv Outros
		t.Errorf("IND_ATIV = %q, want 1", fields0000[15])
	}

	// Last line must be Registro 9999 (encerramento)
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "|9999|") {
		t.Errorf("last line must be |9999|..., got %q", lastLine)
	}

	// Must contain all mandatory block opening/closing registers
	mandatoryRegisters := []string{
		"|0000|", "|0001|", "|0990|",
		"|B001|", "|B990|",
		"|C001|", "|C990|",
		"|D001|", "|D990|",
		"|E001|", "|E990|",
		"|G001|", "|G990|",
		"|H001|", "|H990|",
		"|K001|", "|K990|",
		"|1001|", "|1990|",
		"|9001|", "|9990|", "|9999|",
	}
	for _, reg := range mandatoryRegisters {
		if !strings.Contains(content, reg) {
			t.Errorf("missing mandatory register %s in output", reg)
		}
	}

	// Must contain 9900 registers (block count summary)
	if !strings.Contains(content, "|9900|") {
		t.Error("missing Registro 9900 (block count summary)")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — Registro 0005 included in output when populated
// ---------------------------------------------------------------------------

func TestSaveFileTXT_WithRegistro0005(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_0005.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "TESTE LTDA"
	r.CNPJ = "11222333000181"
	r.UF = "RJ"
	r.IE = "987654321"
	r.CodMun = 3304557
	r.IndPerfil = PerfilB
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0
	f.Bloco0.Registro0001.Registro0005 = &Registro0005{
		Fantasia: "TESTE FANTASIA",
		CEP:      "20040020",
		Endereco: "AV RIO BRANCO",
		Num:      "100",
		Compl:    "SALA 1001",
		Bairro:   "CENTRO",
		Fone:     "2133334444",
		Email:    "teste@teste.com.br",
	}

	err := f.SaveFileTXT()
	if err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	content := string(data)

	if !strings.Contains(content, "|0005|") {
		t.Error("output should contain Registro 0005")
	}
	if !strings.Contains(content, "TESTE FANTASIA") {
		t.Error("output should contain fantasia name")
	}
	if !strings.Contains(content, "20040020") {
		t.Error("output should contain CEP")
	}
	if !strings.Contains(content, "teste@teste.com.br") {
		t.Error("output should contain email")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — Block order validation (legislative requirement)
// SPED files must write blocks in this exact order: 0, B, C, D, E, G, H, K, 1, 9
// ---------------------------------------------------------------------------

func TestSaveFileTXT_BlockOrder(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_order.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "ORDEM TESTE"
	r.CNPJ = "99999999000199"
	r.UF = "MG"
	r.IE = "111222333"
	r.CodMun = 3106200
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0

	err := f.SaveFileTXT()
	if err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	content := string(data)

	// Verify block order by finding positions of opening registers
	blockOpens := []string{"|0001|", "|B001|", "|C001|", "|D001|", "|E001|", "|G001|", "|H001|", "|K001|", "|1001|", "|9001|"}
	prevIdx := -1
	for _, reg := range blockOpens {
		idx := strings.Index(content, reg)
		if idx < 0 {
			t.Errorf("missing block opening register %s", reg)
			continue
		}
		if idx <= prevIdx {
			t.Errorf("block %s appears before previous block (position %d <= %d)", reg, idx, prevIdx)
		}
		prevIdx = idx
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — Registro 9999 line count (QTD_LIN)
// Per legislation, 9999 must contain the total number of lines in the file.
// ---------------------------------------------------------------------------

func TestSaveFileTXT_Registro9999_TotalLines(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_9999.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "CONTAGEM TESTE"
	r.CNPJ = "88888888000188"
	r.UF = "PR"
	r.IE = "444555666"
	r.CodMun = 4106902
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0

	err := f.SaveFileTXT()
	if err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	content := string(data)
	lines := strings.Split(strings.TrimRight(content, "\r\n"), "\r\n")
	totalLines := len(lines)

	// Find 9999 line
	lastLine := lines[totalLines-1]
	if !strings.HasPrefix(lastLine, "|9999|") {
		t.Fatalf("last line should be |9999|, got %q", lastLine)
	}

	// Parse QTD_LIN from |9999|QTD_LIN|
	fields := strings.Split(lastLine, "|")
	qtdLin := fields[2]
	wantQtd := strings.TrimSpace(qtdLin)

	// Verify the count matches
	expectedStr := ""
	for _, c := range wantQtd {
		if c >= '0' && c <= '9' {
			expectedStr += string(c)
		}
	}

	// The QTD_LIN value should equal the total line count in the file
	var parsedQtd int
	for _, c := range expectedStr {
		parsedQtd = parsedQtd*10 + int(c-'0')
	}
	if parsedQtd != totalLines {
		t.Errorf("Registro 9999 QTD_LIN = %d, but file has %d lines", parsedQtd, totalLines)
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — Pipe-delimited format validation
// Per Guia Pratico EFD-ICMS/IPI: every field is delimited by |
// ---------------------------------------------------------------------------

func TestSaveFileTXT_AllLinesPipeDelimited(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_pipe.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "PIPE TESTE"
	r.CNPJ = "77777777000177"
	r.UF = "RS"
	r.IE = "777888999"
	r.CodMun = 4314902
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0

	err := f.SaveFileTXT()
	if err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), "\r\n")

	for i, line := range lines {
		if line == "" {
			continue
		}
		if line[0] != '|' {
			t.Errorf("line %d: first char must be |, got %q", i, string(line[0]))
		}
		if line[len(line)-1] != '|' {
			t.Errorf("line %d: last char must be |, got %q", i, string(line[len(line)-1]))
		}
		// Must have at least 2 pipes (|REG|...|)
		pipeCount := strings.Count(line, "|")
		if pipeCount < 3 {
			t.Errorf("line %d: expected at least 3 pipes, got %d: %q", i, pipeCount, line)
		}
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — CRLF line endings (per SPED specification)
// ---------------------------------------------------------------------------

func TestSaveFileTXT_CRLFLineEndings(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_crlf.txt"
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "CRLF TESTE"
	r.CNPJ = "66666666000166"
	r.UF = "SC"
	r.IE = "666777888"
	r.CodMun = 4205407
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0

	_ = f.SaveFileTXT()

	data, _ := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	content := string(data)
	if !strings.Contains(content, "\r\n") {
		t.Error("SPED file must use CRLF line endings")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscal — LimpaRegistros / CancelaGeracao
// ---------------------------------------------------------------------------

func TestCancelaGeracao_ResetsInitialized(t *testing.T) {
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	_ = f.IniciaGeracao()
	if !f.Inicializado {
		t.Fatal("should be initialized after IniciaGeracao")
	}
	f.CancelaGeracao()
	if f.Inicializado {
		t.Error("should not be initialized after CancelaGeracao")
	}
}

// ---------------------------------------------------------------------------
// OpenBlocos — IndMov semantics
// ---------------------------------------------------------------------------

func TestOpenBlocos_IndDadDefaults(t *testing.T) {
	reg := NewRegistro0001()
	if reg.IndMov != 1 {
		t.Errorf("IndMov should default to 1 (sem dados), got %d", reg.IndMov)
	}
}

// ---------------------------------------------------------------------------
// WriteRegistro0000 — generates correct pipe-delimited line
// ---------------------------------------------------------------------------

func TestWriteRegistro0000_OutputFormat(t *testing.T) {
	f := NewSPEDFiscal()
	f.DtIni = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao100
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "ABC LTDA"
	r.CNPJ = "00000000000100"
	r.UF = "SP"
	r.IE = "999"
	r.CodMun = 3550308
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivIndustrial

	f.Bloco0.WriteRegistro0000()

	if len(f.Bloco0.Conteudo) != 1 {
		t.Fatalf("expected 1 line, got %d", len(f.Bloco0.Conteudo))
	}

	line := f.Bloco0.Conteudo[0]
	fields := strings.Split(line, "|")

	// fields[1] = "0000", fields[2] = CodVer "001"
	if fields[1] != "0000" {
		t.Errorf("REG = %q", fields[1])
	}
	if fields[2] != "001" { // VlVersao100 → "001"
		t.Errorf("COD_VER = %q, want 001", fields[2])
	}
	if fields[3] != "0" {
		t.Errorf("COD_FIN = %q, want 0", fields[3])
	}
	if fields[4] != "01012026" {
		t.Errorf("DT_INI = %q, want 01012026", fields[4])
	}
	if fields[5] != "31012026" {
		t.Errorf("DT_FIN = %q, want 31012026", fields[5])
	}
}

// ---------------------------------------------------------------------------
// IndMov — Block movement indicator
// ---------------------------------------------------------------------------

func TestIndMov_String(t *testing.T) {
	if IndMovComDados.String() != "0" {
		t.Error("IndMovComDados should be 0")
	}
	if IndMovSemDados.String() != "1" {
		t.Error("IndMovSemDados should be 1")
	}
}

// ---------------------------------------------------------------------------
// SPEDFiscalError
// ---------------------------------------------------------------------------

func TestSPEDFiscalError_ImplementsError(t *testing.T) {
	err := &SPEDFiscalError{"test msg"}
	if err.Error() != "SPED Fiscal: test msg" {
		t.Errorf("Error() = %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Vigencia dos blocos B, G e K
//
// O ACBr omite o bloco inteiro fora da vigencia -- nem a abertura nem o
// encerramento sao emitidos (ACBrSpedFiscal.pas, WriteBloco_B/_G/_K).
// Blocos: B a partir de 2019-01-01, G de 2011-01-01, K de 2016-01-01.
// ---------------------------------------------------------------------------

func newSPEDFiscalParaPeriodo(t *testing.T, ano int, mes time.Month, ultimoDia int) *SPEDFiscal {
	t.Helper()
	f := NewSPEDFiscal()
	f.Path = t.TempDir()
	f.Arquivo = "sped_vigencia.txt"
	f.DtIni = time.Date(ano, mes, 1, 0, 0, 0, 0, time.UTC)
	f.DtFin = time.Date(ano, mes, ultimoDia, 0, 0, 0, 0, time.UTC)

	r := f.Bloco0.Registro0000
	r.CodVer = VlVersao115
	r.CodFin = CodFinOriginal
	r.DtIni = f.DtIni
	r.DtFin = f.DtFin
	r.Nome = "EMPRESA VIGENCIA LTDA"
	r.CNPJ = "11222333000181"
	r.UF = "SP"
	r.IE = "123456789"
	r.CodMun = 3550308
	r.IndPerfil = PerfilA
	r.IndAtiv = AtivOutros

	f.Bloco0.Registro0001.IndMov = 0
	return f
}

func TestSaveFileTXT_VigenciaBlocosBGK(t *testing.T) {
	casos := []struct {
		nome             string
		ano              int
		mes              time.Month
		ultimoDia        int
		temB, temG, temK bool
	}{
		{"2010 anterior a todos", 2010, time.March, 31, false, false, false},
		{"2012 so bloco G", 2012, time.March, 31, false, true, false},
		{"2017 blocos G e K", 2017, time.March, 31, false, true, true},
		{"2019 todos", 2019, time.March, 31, true, true, true},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			f := newSPEDFiscalParaPeriodo(t, c.ano, c.mes, c.ultimoDia)
			if err := f.SaveFileTXT(); err != nil {
				t.Fatalf("SaveFileTXT error: %v", err)
			}
			data, err := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
			if err != nil {
				t.Fatalf("ReadFile error: %v", err)
			}
			content := string(data)

			checar := func(bloco string, esperado bool) {
				abertura := "|" + bloco + "001|"
				encerramento := "|" + bloco + "990|"
				if got := strings.Contains(content, abertura); got != esperado {
					t.Errorf("%s presente = %v, want %v", abertura, got, esperado)
				}
				if got := strings.Contains(content, encerramento); got != esperado {
					t.Errorf("%s presente = %v, want %v", encerramento, got, esperado)
				}
				// Bloco fora de vigencia tambem nao pode aparecer no 9900.
				if !esperado && strings.Contains(content, "|9900|"+bloco+"001|") {
					t.Errorf("9900 nao deve listar %s001 fora da vigencia", bloco)
				}
			}
			checar("B", c.temB)
			checar("G", c.temG)
			checar("K", c.temK)
		})
	}
}

// ---------------------------------------------------------------------------
// Registro 9900 conta registros de dados a partir dos contadores
// ---------------------------------------------------------------------------

func TestPopulateRegistro9900_ContaRegistrosDeDados(t *testing.T) {
	f := newSPEDFiscalParaPeriodo(t, 2026, time.September, 30)

	// Dois 0150, sendo que o primeiro tem tres 0175 (registro neto).
	// len(slice) nao alcanca netos -- so o contador conta certo.
	p1 := &Registro0150{CodPart: "P1", Nome: "PARTICIPANTE 1", CodPais: "01058"}
	p1.Registro0175 = []*Registro0175{
		{DtAlt: f.DtIni, NrCampo: "03", ContAnt: "A"},
		{DtAlt: f.DtIni, NrCampo: "04", ContAnt: "B"},
		{DtAlt: f.DtIni, NrCampo: "05", ContAnt: "C"},
	}
	p2 := &Registro0150{CodPart: "P2", Nome: "PARTICIPANTE 2", CodPais: "01058"}
	f.Bloco0.Registro0001.Registro0150 = []*Registro0150{p1, p2}

	if err := f.SaveFileTXT(); err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "|9900|0150|2|") {
		t.Error("9900 deve registrar 2 ocorrencias de 0150")
	}
	if !strings.Contains(content, "|9900|0175|3|") {
		t.Error("9900 deve registrar 3 ocorrencias de 0175 (registro neto)")
	}
}

// ---------------------------------------------------------------------------
// Propagacao de DT_INI/DT_FIN aos blocos
//
// Varios writers decidem o layout por vigencia lendo b.DtIni. Antes, so o
// setter SetDtIni propagava; a atribuicao direta em f.DtIni -- que e o que o
// exemplo do README faz -- deixava os blocos com data zerada, e o Registro
// 0002 nunca era emitido.
// ---------------------------------------------------------------------------

func TestSaveFileTXT_Registro0002_ComAtribuicaoDiretaDeDtIni(t *testing.T) {
	f := newSPEDFiscalParaPeriodo(t, 2026, time.September, 30)
	f.Bloco0.Registro0000.IndAtiv = AtivIndustrial
	f.Bloco0.Registro0002 = &Registro0002{ClasEstabInd: "01"}

	if err := f.SaveFileTXT(); err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if f.Bloco0.DtIni.IsZero() {
		t.Error("Bloco0.DtIni nao deveria estar zerado apos IniciaGeracao")
	}
	if !strings.Contains(string(data), "|0002|01|") {
		t.Error("output deveria conter o Registro 0002")
	}
}

// ---------------------------------------------------------------------------
// IND_FRT e IND_PGTO por vigencia
// ---------------------------------------------------------------------------

func TestIndFrt_StringEm(t *testing.T) {
	ate2011 := time.Date(2011, 6, 1, 0, 0, 0, 0, time.UTC)
	ate2017 := time.Date(2015, 6, 1, 0, 0, 0, 0, time.UTC)
	de2018 := time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC)

	casos := []struct {
		v       IndFrt
		ate2011 string
		ate2017 string
		novo    string
	}{
		{FrtContaEmitente, "1", "0", "0"},
		{FrtContaDestinatario, "2", "1", "1"},
		{FrtContaTerceiros, "0", "2", "2"},
		{FrtProprioPorContaRemetente, "1", "0", "3"},
		{FrtProprioContaDestinatario, "2", "1", "4"},
		{FrtSemCobranca, "9", "9", "9"},
		{FrtNenhum, "", "", ""},
	}

	for _, c := range casos {
		if got := c.v.StringEm(ate2011); got != c.ate2011 {
			t.Errorf("IND_FRT(%d) ate 2011 = %q, want %q", int(c.v), got, c.ate2011)
		}
		if got := c.v.StringEm(ate2017); got != c.ate2017 {
			t.Errorf("IND_FRT(%d) 2012-2017 = %q, want %q", int(c.v), got, c.ate2017)
		}
		if got := c.v.StringEm(de2018); got != c.novo {
			t.Errorf("IND_FRT(%d) de 2018 = %q, want %q", int(c.v), got, c.novo)
		}
	}
}

func TestIndPgto_StringEm(t *testing.T) {
	antes := time.Date(2012, 6, 1, 0, 0, 0, 0, time.UTC)
	depois := time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC)

	casos := []struct {
		v      IndPgto
		antes  string
		depois string
	}{
		{PgtoVista, "0", "0"},
		{PgtoPrazo, "1", "1"},
		{PgtoSemPagamento, "9", ""}, // deixou de existir em 07/2012
		{PgtoOutros, "", "2"},       // nao existia antes de 07/2012
		{PgtoNenhum, "", ""},
	}

	for _, c := range casos {
		if got := c.v.StringEm(antes); got != c.antes {
			t.Errorf("IND_PGTO(%d) antes de 07/2012 = %q, want %q", int(c.v), got, c.antes)
		}
		if got := c.v.StringEm(depois); got != c.depois {
			t.Errorf("IND_PGTO(%d) de 07/2012 = %q, want %q", int(c.v), got, c.depois)
		}
	}
}

// ---------------------------------------------------------------------------
// Registro C100 - fidelidade ao writer do ACBr
// ---------------------------------------------------------------------------

// blocoCParaTeste devolve um BlocoC pronto com os C100 informados, no periodo dado.
func blocoCParaTeste(t *testing.T, dtIni time.Time, regs ...*RegistroC100) *BlocoC {
	t.Helper()
	f := NewSPEDFiscal()
	b := f.BlocoC
	b.DtIni = dtIni
	b.RegistroC001.IndMov = 0
	b.RegistroC001.RegistroC100 = regs
	return b
}

func c100Padrao() *RegistroC100 {
	return &RegistroC100{
		CodPart: "P1",
		CodMod:  "55",
		CodSit:  SitRegular,
		Ser:     "1",
		NumDoc:  "000001",
		ChvNFe:  "35260111222333000181550010000000011000000017",
		DtDoc:   time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
		DtES:    time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC),
		VlDoc:   1000,
		IndPgto: PgtoVista,
		VlMerc:  1000,
		IndFrt:  FrtContaEmitente,
	}
}

func TestWriteRegistroC100_ValoresObrigatoriosZerados(t *testing.T) {
	b := blocoCParaTeste(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), c100Padrao())
	b.writeRegistroC100()

	if len(b.Conteudo) != 1 {
		t.Fatalf("esperava 1 linha, obtive %d", len(b.Conteudo))
	}
	campos := strings.Split(b.Conteudo[0], "|")

	// Documento normal: campo monetario obrigatorio sai 0,00, nunca vazio.
	// Indice = posicao no layout + 1, pois campos[1] e o REG.
	obrigatorios := map[int]string{
		14: "VL_DESC",
		15: "VL_ABAT_NT",
		18: "VL_FRT",
		19: "VL_SEG",
		20: "VL_OUT_DA",
		23: "VL_BC_ICMS_ST",
		24: "VL_ICMS_ST",
		25: "VL_IPI",
	}
	for idx, nome := range obrigatorios {
		if campos[idx] != "0,00" {
			t.Errorf("%s = %q, want 0,00", nome, campos[idx])
		}
	}
	// PIS/COFINS sao sempre nulos no C100.
	for _, idx := range []int{26, 27, 28, 29} {
		if campos[idx] != "" {
			t.Errorf("campo %d (PIS/COFINS) = %q, want vazio", idx, campos[idx])
		}
	}
}

// O parametro nulo dos helpers so esvazia valor ZERO -- nao apaga valor
// preenchido. Entao o efeito de um documento cancelado e duplo e os testes
// abaixo cobrem os dois lados: (a) datas e indicadores sao zerados na origem,
// saindo vazios sempre; (b) os campos monetarios passam a aceitar vazio quando
// valem zero, onde num documento normal sairiam 0,00.
func TestWriteRegistroC100_DocumentoCancelado(t *testing.T) {
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("datas e indicadores zerados na origem", func(t *testing.T) {
		r := c100Padrao() // DtDoc, DtES, IndPgto e IndFrt preenchidos
		r.CodSit = SitCancelado
		b := blocoCParaTeste(t, dt, r)
		b.writeRegistroC100()

		campos := strings.Split(b.Conteudo[0], "|")
		if campos[6] != "02" {
			t.Errorf("COD_SIT = %q, want 02", campos[6])
		}
		for idx, nome := range map[int]string{
			10: "DT_DOC", 11: "DT_E_S", 13: "IND_PGTO", 17: "IND_FRT",
		} {
			if campos[idx] != "" {
				t.Errorf("%s = %q, want vazio em documento cancelado", nome, campos[idx])
			}
		}
	})

	t.Run("monetarios zerados saem vazios", func(t *testing.T) {
		r := c100Padrao()
		r.CodSit = SitCancelado
		r.VlDoc, r.VlMerc = 0, 0
		b := blocoCParaTeste(t, dt, r)
		b.writeRegistroC100()

		campos := strings.Split(b.Conteudo[0], "|")
		for idx, nome := range map[int]string{
			12: "VL_DOC", 16: "VL_MERC", 21: "VL_BC_ICMS", 22: "VL_ICMS",
		} {
			if campos[idx] != "" {
				t.Errorf("%s = %q, want vazio em documento cancelado", nome, campos[idx])
			}
		}
	})

	t.Run("documento regular emite 0,00 nos mesmos campos", func(t *testing.T) {
		r := c100Padrao()
		r.VlDoc, r.VlMerc = 0, 0
		b := blocoCParaTeste(t, dt, r)
		b.writeRegistroC100()

		campos := strings.Split(b.Conteudo[0], "|")
		for idx, nome := range map[int]string{
			12: "VL_DOC", 16: "VL_MERC", 21: "VL_BC_ICMS", 22: "VL_ICMS",
		} {
			if campos[idx] != "0,00" {
				t.Errorf("%s = %q, want 0,00 em documento regular", nome, campos[idx])
			}
		}
	})
}

func TestWriteRegistroC100_NFCe65(t *testing.T) {
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	stEIpi := map[int]string{23: "VL_BC_ICMS_ST", 24: "VL_ICMS_ST", 25: "VL_IPI"}

	t.Run("modelo 65 nao informa ST nem IPI", func(t *testing.T) {
		r := c100Padrao()
		r.CodMod = "65"
		r.ChvNFe = ""
		b := blocoCParaTeste(t, dt, r)
		b.writeRegistroC100()

		campos := strings.Split(b.Conteudo[0], "|")
		for idx, nome := range stEIpi {
			if campos[idx] != "" {
				t.Errorf("%s = %q, want vazio em NFC-e (modelo 65)", nome, campos[idx])
			}
		}
		// VL_MERC continua obrigatorio no modelo 65.
		if campos[16] != "1000,00" {
			t.Errorf("VL_MERC = %q, want 1000,00", campos[16])
		}
	})

	t.Run("modelo 55 emite 0,00 nos mesmos campos", func(t *testing.T) {
		b := blocoCParaTeste(t, dt, c100Padrao())
		b.writeRegistroC100()

		campos := strings.Split(b.Conteudo[0], "|")
		for idx, nome := range stEIpi {
			if campos[idx] != "0,00" {
				t.Errorf("%s = %q, want 0,00 em NF-e (modelo 55)", nome, campos[idx])
			}
		}
	})
}

func TestWriteRegistroC100_ChaveObrigatoriaModelo55(t *testing.T) {
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("sem chave entra em panico", func(t *testing.T) {
		r := c100Padrao()
		r.ChvNFe = "   "
		b := blocoCParaTeste(t, dt, r)
		defer func() {
			if recover() == nil {
				t.Error("esperava panico para NF-e modelo 55 sem chave de acesso")
			}
		}()
		b.writeRegistroC100()
	})

	t.Run("numeracao inutilizada dispensa chave", func(t *testing.T) {
		r := c100Padrao()
		r.ChvNFe = ""
		r.CodSit = SitNumInutilizada
		b := blocoCParaTeste(t, dt, r)
		defer func() {
			if rec := recover(); rec != nil {
				t.Errorf("nao deveria entrar em panico: %v", rec)
			}
		}()
		b.writeRegistroC100()
	})
}

func TestWriteRegistroC100_CallbackPodeVetarRegistro(t *testing.T) {
	b := blocoCParaTeste(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		c100Padrao(), c100Padrao())

	chamadas := 0
	b.OnCheckRegistroC100 = func(registro any, abortar *bool) {
		chamadas++
		*abortar = chamadas == 1 // veta apenas o primeiro
	}
	b.writeRegistroC100()

	if chamadas != 2 {
		t.Errorf("callback chamado %d vezes, want 2", chamadas)
	}
	// O veto do primeiro nao pode derrubar o segundo.
	if len(b.Conteudo) != 1 {
		t.Errorf("esperava 1 linha escrita, obtive %d", len(b.Conteudo))
	}
	if b.RegistroC100Count != 1 {
		t.Errorf("RegistroC100Count = %d, want 1", b.RegistroC100Count)
	}
}

// ---------------------------------------------------------------------------
// Campos condicionais por vigencia ou versao de leiaute
//
// Todos estes existiam no writer Delphi e tinham sido perdidos no port: o
// campo era emitido incondicionalmente, gerando linha com campo a mais (ou a
// menos) para o periodo declarado no 0000.
// ---------------------------------------------------------------------------

// bloco0ParaTeste devolve um Bloco0 pronto, no periodo e versao informados.
func bloco0ParaTeste(t *testing.T, dtIni time.Time, ver VersaoLeiauteFiscal) *Bloco0 {
	t.Helper()
	f := NewSPEDFiscal()
	b := f.Bloco0
	b.DtIni = dtIni
	b.Registro0000.CodVer = ver
	b.Registro0001.IndMov = 0
	return b
}

func TestWriteRegistro0150_ParticipanteDoExterior(t *testing.T) {
	casos := []struct {
		nome    string
		codPais string
		codMun  int
		want    string
	}{
		{"nacional usa o municipio do IBGE", "01058", 3550308, "3550308"},
		{"exterior usa o literal 9999999", "02321", 0, "9999999"},
		{"brasil sem zero a esquerda ainda e nacional", "1058", 3550308, "3550308"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			b := bloco0ParaTeste(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), VlVersao115)
			b.Registro0001.Registro0150 = []*Registro0150{
				{CodPart: "P1", Nome: "PARTICIPANTE", CodPais: c.codPais, CodMun: c.codMun},
			}
			b.writeRegistro0150()

			campos := strings.Split(b.Conteudo[0], "|")
			if campos[8] != c.want {
				t.Errorf("COD_MUN = %q, want %q", campos[8], c.want)
			}
		})
	}
}

func TestWriteRegistro0200_CESTApenasApos2017(t *testing.T) {
	item := func() []*Registro0200 {
		return []*Registro0200{{CodItem: "I1", DescrItem: "ITEM", UnidInv: "UN", CEST: "0100100"}}
	}

	t.Run("antes de 2017 o campo nao existe", func(t *testing.T) {
		b := bloco0ParaTeste(t, time.Date(2016, 12, 1, 0, 0, 0, 0, time.UTC), VlVersao109)
		b.Registro0001.Registro0200 = item()
		b.writeRegistro0200()
		if n := strings.Count(b.Conteudo[0], "|"); n != 13 {
			t.Errorf("delimitadores = %d, want 13 (sem CEST)", n)
		}
		if strings.Contains(b.Conteudo[0], "0100100") {
			t.Error("CEST nao deveria ser emitido antes de 2017")
		}
	})

	t.Run("de 2017 em diante o campo e emitido", func(t *testing.T) {
		b := bloco0ParaTeste(t, time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC), VlVersao110)
		b.Registro0001.Registro0200 = item()
		b.writeRegistro0200()
		if !strings.Contains(b.Conteudo[0], "|0100100|") {
			t.Errorf("CEST ausente: %q", b.Conteudo[0])
		}
	})
}

func TestWriteRegistro0220_CodBarraApenasAposVersao114(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) string {
		b := bloco0ParaTeste(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), ver)
		pai := &Registro0200{CodItem: "I1"}
		pai.Registro0220 = []*Registro0220{{UnidConv: "CX", FatConv: 12, CodBarra: "789"}}
		b.writeRegistro0220(pai)
		return b.Conteudo[0]
	}

	if l := monta(VlVersao114); strings.Contains(l, "789") {
		t.Errorf("COD_BARRA nao deveria sair na versao 114: %q", l)
	}
	if l := monta(VlVersao115); !strings.Contains(l, "|789|") {
		t.Errorf("COD_BARRA deveria sair na versao 115: %q", l)
	}
}

func TestWriteRegistroC170_VlAbatNTApenasAposVersao111(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) string {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.BlocoC
		b.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		pai := &RegistroC100{CodMod: "55"}
		pai.RegistroC170 = []*RegistroC170{{NumItem: "1", CodItem: "I1", VlAbatNT: 7}}
		b.writeRegistroC170(pai)
		return b.Conteudo[0]
	}

	if l := monta(VlVersao111); strings.Contains(l, "7,00") {
		t.Errorf("VL_ABAT_NT nao deveria sair na versao 111: %q", l)
	}
	if l := monta(VlVersao112); !strings.Contains(l, "|7,00|") {
		t.Errorf("VL_ABAT_NT deveria sair na versao 112: %q", l)
	}
}

func TestWriteRegistroD100_MunicipiosApenas2018(t *testing.T) {
	monta := func(ano int) string {
		f := NewSPEDFiscal()
		b := f.BlocoD
		b.DtIni = time.Date(ano, 3, 1, 0, 0, 0, 0, time.UTC)
		b.RegistroD001.IndMov = 0
		b.RegistroD001.RegistroD100 = []*RegistroD100{
			{CodMod: "57", NumDoc: "1", CodMunOrig: "3550308", CodMunDest: "3304557"},
		}
		b.writeRegistroD100()
		return b.Conteudo[0]
	}

	if l := monta(2017); strings.Contains(l, "3550308") {
		t.Errorf("COD_MUN_ORIG nao deveria sair antes de 2018: %q", l)
	}
	l := monta(2018)
	if !strings.Contains(l, "|3550308|3304557|") {
		t.Errorf("municipios deveriam sair de 2018 em diante: %q", l)
	}
}

func TestWriteRegistroH010_VlItemIRApenas2015(t *testing.T) {
	monta := func(ano int) string {
		f := NewSPEDFiscal()
		b := f.BlocoH
		b.DtIni = time.Date(ano, 3, 1, 0, 0, 0, 0, time.UTC)
		pai := &RegistroH005{}
		pai.RegistroH010 = []*RegistroH010{{CodItem: "I1", Unid: "UN", VlItemIR: 5}}
		b.writeRegistroH010(pai)
		return b.Conteudo[0]
	}

	if l := monta(2014); strings.Contains(l, "5,00") {
		t.Errorf("VL_ITEM_IR nao deveria sair antes de 2015: %q", l)
	}
	if l := monta(2015); !strings.Contains(l, "|5,00|") {
		t.Errorf("VL_ITEM_IR deveria sair de 2015 em diante: %q", l)
	}
}

// ---------------------------------------------------------------------------
// Bloco B (ISS) - fidelidade ao ACBr
// ---------------------------------------------------------------------------

func blocoBParaTeste(t *testing.T) *BlocoB {
	t.Helper()
	f := NewSPEDFiscal()
	b := f.BlocoB
	b.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	b.RegistroB001.IndMov = 0
	return b
}

func TestWriteRegistroB020_LayoutCompleto(t *testing.T) {
	b := blocoBParaTeste(t)
	r := &RegistroB020{
		CodPart: "P1", CodMod: "55", CodSit: SitRegular, Ser: "1", NumDoc: "10",
		DtDoc: b.DtIni, CodMunServ: "5300108", VlCont: 100, VlBcIss: 100, VlIss: 5,
	}
	r.RegistroB025 = []*RegistroB025{{VlContP: 100, VlBcIssP: 100, AliqIss: 5, VlIssP: 5, CodServ: "0101"}}
	b.RegistroB001.RegistroB020 = []*RegistroB020{r}
	b.writeRegistroB020()

	// REG + 20 campos = 21 itens, logo 22 delimitadores.
	// Conforme ACBrEFDBloco_B_Class.pas, WriteRegistroB020.
	if n := strings.Count(b.Conteudo[0], "|"); n != 22 {
		t.Errorf("B020: delimitadores = %d, want 22\n%q", n, b.Conteudo[0])
	}
	// REG + 6 campos = 7 itens, logo 8 delimitadores.
	if n := strings.Count(b.Conteudo[1], "|"); n != 8 {
		t.Errorf("B025: delimitadores = %d, want 8\n%q", n, b.Conteudo[1])
	}
}

// O ACBr emite a linha do B350 com o literal "B035" (WriteRegistroB350 em
// ACBrEFDBloco_B_Class.pas). Mantido por fidelidade ao original -- este teste
// existe para que a escolha seja explicita e nao se perca numa refatoracao.
func TestWriteRegistroB350_EmiteLiteralB035ComoNoACBr(t *testing.T) {
	b := blocoBParaTeste(t)
	b.RegistroB001.RegistroB350 = []*RegistroB350{
		{CodCtd: "C1", CtaIss: "CT", CtaCosif: "12345678", QtdOcor: 3, CodServ: "0101", VlCont: 50},
	}
	b.writeRegistroB350()

	if !strings.HasPrefix(b.Conteudo[0], "|B035|") {
		t.Errorf("B350 deveria emitir o literal B035 (como o ACBr): %q", b.Conteudo[0])
	}
	// REG + 10 campos = 11 itens, logo 12 delimitadores.
	if n := strings.Count(b.Conteudo[0], "|"); n != 12 {
		t.Errorf("delimitadores = %d, want 12\n%q", n, b.Conteudo[0])
	}
}

func TestWriteRegistroB030_ComFilhoB035(t *testing.T) {
	b := blocoBParaTeste(t)
	r := &RegistroB030{
		CodMod: "3B", Ser: "1", NumDocIni: "1", NumDocFin: "50",
		DtDoc: b.DtIni, QtdCanc: 2, VlCont: 300,
	}
	r.RegistroB035 = []*RegistroB035{{VlContP: 300, VlBcIssP: 300, AliqIss: 5, VlIssP: 15, CodServ: "0101"}}
	b.RegistroB001.RegistroB030 = []*RegistroB030{r}
	b.writeRegistroB030()

	if len(b.Conteudo) != 2 {
		t.Fatalf("esperava B030 + B035, obtive %d linhas", len(b.Conteudo))
	}
	if !strings.HasPrefix(b.Conteudo[1], "|B035|") {
		t.Errorf("filho deveria ser B035: %q", b.Conteudo[1])
	}
}

// No ACBr o B500 e filho do B001, nao do B470, e tem apenas VL_REC, QTD_PROF
// e VL_OR (sociedade uniprofissional).
func TestWriteRegistroB500_FilhoDoB001ComFilhoB510(t *testing.T) {
	b := blocoBParaTeste(t)
	r := &RegistroB500{VlRec: 1000, QtdProf: 3, VlOR: 90}
	r.RegistroB510 = []*RegistroB510{
		{IndProf: "1", IndEsc: "0", IndSoc: "1", CPF: "11122233344", Nome: "FULANO"},
	}
	b.RegistroB001.RegistroB500 = []*RegistroB500{r}
	b.writeRegistroB500()

	if b.Conteudo[0] != "|B500|1000,00|3|90,00|" {
		t.Errorf("B500 = %q", b.Conteudo[0])
	}
	if b.Conteudo[1] != "|B510|1|0|1|11122233344|FULANO|" {
		t.Errorf("B510 = %q", b.Conteudo[1])
	}
}

// ---------------------------------------------------------------------------
// Registros acrescentados: familia C110, C101, E111, E116 e 1010
// ---------------------------------------------------------------------------

func TestWriteRegistroC101_DIFAL(t *testing.T) {
	f := NewSPEDFiscal()
	b := f.BlocoC
	b.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	pai := &RegistroC100{CodMod: "55"}
	pai.RegistroC101 = []*RegistroC101{{VlFcpUFDest: 10, VlICMSUFDest: 20, VlICMSUFRem: 5}}
	b.writeRegistroC101(pai)

	if b.Conteudo[0] != "|C101|10,00|20,00|5,00|" {
		t.Errorf("C101 = %q", b.Conteudo[0])
	}
}

func TestWriteRegistroC110_ComFilhos(t *testing.T) {
	f := NewSPEDFiscal()
	f.Bloco0.Registro0000.CodVer = VlVersao115
	b := f.BlocoC
	b.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	pai := &RegistroC100{CodMod: "55"}
	c110 := &RegistroC110{CodInf: "I01", TxtCompl: "OBSERVACAO"}
	c110.RegistroC111 = []*RegistroC111{{NumProc: "P1", IndProc: OrigProcSefaz}}
	c110.RegistroC112 = []*RegistroC112{{UF: "SP", NumDa: "DA1", VlDa: 100}}
	c110.RegistroC113 = []*RegistroC113{{CodPart: "P1", CodMod: "55", NumDoc: "9", ChvDocE: "CHV"}}
	c110.RegistroC114 = []*RegistroC114{{CodMod: "2D", EcfFab: "FAB", EcfCx: "1", NumDoc: "5"}}
	pai.RegistroC110 = []*RegistroC110{c110}
	b.writeRegistroC110(pai)

	esperado := []string{"|C110|", "|C111|", "|C112|", "|C113|", "|C114|"}
	if len(b.Conteudo) != len(esperado) {
		t.Fatalf("esperava %d linhas, obtive %d", len(esperado), len(b.Conteudo))
	}
	for i, pref := range esperado {
		if !strings.HasPrefix(b.Conteudo[i], pref) {
			t.Errorf("linha %d = %q, want prefixo %s", i, b.Conteudo[i], pref)
		}
	}
}

func TestWriteRegistroC113_ChaveApenasAposVersao110(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) string {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.BlocoC
		pai := &RegistroC110{}
		pai.RegistroC113 = []*RegistroC113{{CodMod: "55", NumDoc: "9", ChvDocE: "CHAVE123"}}
		b.writeRegistroC113(pai)
		return b.Conteudo[0]
	}

	if l := monta(VlVersao109); strings.Contains(l, "CHAVE123") {
		t.Errorf("CHV_DOCe nao deveria sair antes da versao 110: %q", l)
	}
	if l := monta(VlVersao110); !strings.Contains(l, "|CHAVE123|") {
		t.Errorf("CHV_DOCe deveria sair a partir da versao 110: %q", l)
	}
}

func TestWriteRegistroE111_Ajuste(t *testing.T) {
	f := NewSPEDFiscal()
	b := f.BlocoE
	pai := &RegistroE110{}
	pai.RegistroE111 = []*RegistroE111{{CodAjApur: "SP000001", DescrComplAj: "AJUSTE", VlAjApur: 150}}
	b.writeRegistroE111(pai)

	if b.Conteudo[0] != "|E111|SP000001|AJUSTE|150,00|" {
		t.Errorf("E111 = %q", b.Conteudo[0])
	}
}

func TestWriteRegistroE116_PorVersaoDeLeiaute(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) []string {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.BlocoE
		pai := &RegistroE110{}
		pai.RegistroE116 = []*RegistroE116{
			{CodOR: "000", VlOR: 500, CodRec: "REC", MesRef: "092026"},
		}
		b.writeRegistroE116(pai)
		return b.Conteudo
	}

	if l := monta(VlVersao101); len(l) != 0 {
		t.Errorf("E116 nao deve ser gerado na versao 101: %v", l)
	}
	if l := monta(VlVersao102); len(l) != 1 || strings.Contains(l[0], "092026") {
		t.Errorf("versao 102 nao deve conter MES_REF: %v", l)
	}
	if l := monta(VlVersao103); len(l) != 1 || !strings.Contains(l[0], "|092026|") {
		t.Errorf("versao 103 deve conter MES_REF: %v", l)
	}
}

func TestWriteRegistro1010_PorVersaoDeLeiaute(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) string {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.Bloco1
		b.Registro1001.IndMov = 0
		b.Registro1001.Registro1010 = []*Registro1010{{
			IndExp: "N", IndCCRF: "N", IndComb: "N", IndUsina: "N", IndVA: "N",
			IndEE: "N", IndCart: "N", IndForm: "N", IndAer: "N",
			IndGIAF1: "G1", IndGIAF3: "G3", IndGIAF4: "G4",
			IndRestRessarcComplICMS: "R",
		}}
		b.writeRegistro1010()
		return b.Conteudo[0]
	}

	if l := monta(VlVersao111); strings.Contains(l, "G1") || strings.Contains(l, "|R|") {
		t.Errorf("versao 111 nao deve ter GIAF nem REST_RESSARC: %q", l)
	}
	if l := monta(VlVersao112); !strings.Contains(l, "|G1|G3|G4|") || strings.Contains(l, "|R|") {
		t.Errorf("versao 112 deve ter GIAF e nao REST_RESSARC: %q", l)
	}
	if l := monta(VlVersao113); !strings.Contains(l, "|G1|G3|G4|R|") {
		t.Errorf("versao 113 deve ter GIAF e REST_RESSARC: %q", l)
	}
}

// ---------------------------------------------------------------------------
// Blocos G (CIAP) e H (inventario) - registros de dados
// ---------------------------------------------------------------------------

func TestWriteRegistroG110_PorVersaoDeLeiaute(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) (linhas []string, qtd int) {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.BlocoG
		b.RegistroG001.IndMov = 0
		b.RegistroG001.RegistroG110 = []*RegistroG110{{
			ModoCiap: "A", SaldoInICMS: 100, SaldoFnICMS: 50, SomParc: 10,
			VlTribExp: 1, VlTotal: 200, IndPerSai: 0.5, ICMSAprop: 5, SomICMSOC: 2,
		}}
		b.writeRegistroG110()
		return b.Conteudo, b.RegistroG990.QtdLinG
	}
	primeira := func(ver VersaoLeiauteFiscal) string {
		l, _ := monta(ver)
		return l[0]
	}

	// Nas versoes 100/101 o ACBr nao cobre nenhum dos dois ramos: a linha nao
	// sai, mas o contador do G990 incrementa assim mesmo.
	if l, qtd := monta(VlVersao101); len(l) != 0 || qtd != 1 {
		t.Errorf("versao 101 deveria contar sem emitir linha: linhas=%v qtd=%d", l, qtd)
	}
	// Ate a 102 o registro tem MODO_CIAP e SALDO_FN_ICMS.
	if l := primeira(VlVersao102); !strings.Contains(l, "|A|100,00|50,00|") {
		t.Errorf("versao 102: %q", l)
	}
	// Da 103 em diante os dois campos somem.
	l := primeira(VlVersao103)
	if strings.Contains(l, "|A|") {
		t.Errorf("MODO_CIAP nao deveria sair da versao 103 em diante: %q", l)
	}
	if !strings.Contains(l, "|100,00|10,00|") {
		t.Errorf("versao 103 deveria ir de SALDO_IN_ICMS direto a SOM_PARC: %q", l)
	}
}

func TestWriteRegistroG140_ValidaItemNo0200(t *testing.T) {
	monta := func(codItemNo0200 string) (panicou bool) {
		defer func() { panicou = recover() != nil }()
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = VlVersao115
		f.Bloco0.Registro0001.Registro0200 = []*Registro0200{{CodItem: codItemNo0200}}
		b := f.BlocoG
		pai := &RegistroG130{}
		pai.RegistroG140 = []*RegistroG140{{NumItem: "1", CodItem: "ITEM1", Qtde: 2}}
		b.writeRegistroG140(pai)
		return
	}

	if monta("ITEM1") {
		t.Error("item declarado no 0200 nao deveria entrar em panico")
	}
	if !monta("OUTRO") {
		t.Error("item ausente do 0200 deveria entrar em panico")
	}
}

func TestWriteRegistroH030_SeisCasasDecimais(t *testing.T) {
	f := NewSPEDFiscal()
	f.Bloco0.Registro0000.CodVer = VlVersao115
	b := f.BlocoH
	b.DtIni = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	pai := &RegistroH010{}
	pai.RegistroH030 = []*RegistroH030{{VlICMSOp: 1.5, VlBcICMSST: 2, VlICMSST: 3, VlFCP: 4}}
	b.writeRegistroH030(pai)

	if b.Conteudo[0] != "|H030|1,500000|2,000000|3,000000|4,000000|" {
		t.Errorf("H030 = %q", b.Conteudo[0])
	}
}

func TestWriteRegistroH020_H030_ExigemVersaoEPeriodo(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal, ano int) int {
		f := NewSPEDFiscal()
		f.Bloco0.Registro0000.CodVer = ver
		b := f.BlocoH
		b.DtIni = time.Date(ano, 9, 1, 0, 0, 0, 0, time.UTC)
		pai := &RegistroH010{}
		pai.RegistroH020 = []*RegistroH020{{BcICMS: 10, VlICMS: 1}}
		pai.RegistroH030 = []*RegistroH030{{VlICMSOp: 1}}
		b.writeRegistroH020(pai)
		b.writeRegistroH030(pai)
		return len(b.Conteudo)
	}

	if n := monta(VlVersao103, 2026); n != 0 {
		t.Errorf("versao anterior a 104 nao deveria emitir H020/H030, emitiu %d", n)
	}
	if n := monta(VlVersao115, 2011); n != 0 {
		t.Errorf("periodo anterior a 07/2012 nao deveria emitir H020/H030, emitiu %d", n)
	}
	if n := monta(VlVersao115, 2026); n != 2 {
		t.Errorf("esperava H020 e H030, obtive %d linhas", n)
	}
}

// ---------------------------------------------------------------------------
// Bloco K (producao e estoque)
// ---------------------------------------------------------------------------

func blocoKParaTeste(t *testing.T, ver VersaoLeiauteFiscal, ano int) (*BlocoK, *RegistroK100) {
	t.Helper()
	f := NewSPEDFiscal()
	f.Bloco0.Registro0000.CodVer = ver
	b := f.BlocoK
	b.DtIni = time.Date(ano, 9, 1, 0, 0, 0, 0, time.UTC)
	b.DtFin = time.Date(ano, 9, 30, 0, 0, 0, 0, time.UTC)
	b.RegistroK001.IndMov = 0
	k100 := &RegistroK100{DtIni: b.DtIni, DtFin: b.DtFin}
	b.RegistroK001.RegistroK100 = []*RegistroK100{k100}
	return b, k100
}

// As quantidades do Bloco K passaram de 3 para 6 casas decimais na versao 112.
func TestBlocoK_DecimaisDaQuantidadePorVersao(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal) string {
		b, k100 := blocoKParaTeste(t, ver, 2026)
		k100.RegistroK230 = []*RegistroK230{{CodDocOP: "OP1", DtIniOP: b.DtIni, DtFinOP: b.DtFin, CodItem: "I1", QtdEnc: 2}}
		b.writeRegistroK230(k100)
		return b.Conteudo[0]
	}

	if l := monta(VlVersao111); !strings.Contains(l, "|2,000|") {
		t.Errorf("versao 111 deveria usar 3 casas: %q", l)
	}
	if l := monta(VlVersao112); !strings.Contains(l, "|2,000000|") {
		t.Errorf("versao 112 deveria usar 6 casas: %q", l)
	}
}

// K200 e K280 usam tres casas fixas, independentemente da versao.
func TestWriteRegistroK200_TresCasasFixas(t *testing.T) {
	b, k100 := blocoKParaTeste(t, VlVersao115, 2026)
	k100.RegistroK200 = []*RegistroK200{{DtEst: k100.DtFin, CodItem: "I1", Qtd: 5}}
	b.writeRegistroK200(k100)

	if !strings.Contains(b.Conteudo[0], "|5,000|") {
		t.Errorf("K200 deveria usar 3 casas fixas: %q", b.Conteudo[0])
	}
}

// A partir da versao 117, quem declara leiaute nao completo no K010 deixa de
// emitir parte dos registros.
func TestBlocoK_LeiauteRestritoSuprimeRegistros(t *testing.T) {
	monta := func(ver VersaoLeiauteFiscal, tipo IndTipoLeiaute) int {
		b, k100 := blocoKParaTeste(t, ver, 2026)
		b.RegistroK001.RegistroK010 = &RegistroK010{IndTipoLeiaute: tipo}
		k100.RegistroK260 = []*RegistroK260{{CodOpOS: "OP1", CodItem: "I1", DtSaida: b.DtIni}}
		b.writeRegistroK260(k100)
		return len(b.Conteudo)
	}

	if n := monta(VlVersao117, LeiauteSimplificado); n != 0 {
		t.Errorf("leiaute simplificado na 117 nao deveria emitir K260, emitiu %d", n)
	}
	if n := monta(VlVersao117, LeiauteCompleto); n != 1 {
		t.Errorf("leiaute completo deveria emitir K260, emitiu %d", n)
	}
	if n := monta(VlVersao116, LeiauteSimplificado); n != 1 {
		t.Errorf("antes da 117 a restricao nao se aplica, emitiu %d", n)
	}
}

// QTD_DEST do K220 depende do periodo do ARQUIVO, nao da data do registro.
func TestWriteRegistroK220_QtdDestApenas2018(t *testing.T) {
	monta := func(ano int) string {
		b, k100 := blocoKParaTeste(t, VlVersao115, ano)
		k100.RegistroK220 = []*RegistroK220{{
			DtMov: b.DtIni, CodItemOri: "A", CodItemDest: "B", Qtd: 1, QtdDest: 7,
		}}
		b.writeRegistroK220(k100)
		return b.Conteudo[0]
	}

	if l := monta(2017); strings.Contains(l, "7,000000") {
		t.Errorf("QTD_DEST nao deveria sair antes de 2018: %q", l)
	}
	if l := monta(2018); !strings.Contains(l, "7,000000") {
		t.Errorf("QTD_DEST deveria sair de 2018 em diante: %q", l)
	}
}

func TestWriteRegistroK291_RejeitaQuantidadeNaoPositiva(t *testing.T) {
	monta := func(qtd float64) (panicou bool) {
		defer func() { panicou = recover() != nil }()
		b, k100 := blocoKParaTeste(t, VlVersao115, 2026)
		pai := &RegistroK290{CodDocOP: "OP1", DtIniOP: b.DtIni, DtFinOP: b.DtFin}
		pai.RegistroK291 = []*RegistroK291{{CodItem: "I1", Qtd: qtd}}
		k100.RegistroK290 = []*RegistroK290{pai}
		b.writeRegistroK291(pai)
		return
	}

	if monta(1) {
		t.Error("quantidade positiva nao deveria entrar em panico")
	}
	if !monta(0) {
		t.Error("quantidade zero deveria entrar em panico")
	}
}

// O 9990 declara a quantidade de linhas do Bloco 9, o que inclui a propria
// linha do 9990 e a do 9999. O port contava as duas a menos -- defeito que a
// auditoria campo a campo nao pega, porque e valor calculado e nao layout.
func TestSaveFileTXT_ContagensDoBloco9(t *testing.T) {
	f := newSPEDFiscalParaPeriodo(t, 2026, time.September, 30)
	f.Bloco0.Registro0001.Registro0190 = []*Registro0190{{Unid: "UN", Descr: "UNIDADE"}}

	if err := f.SaveFileTXT(); err != nil {
		t.Fatalf("SaveFileTXT error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(f.Path, f.Arquivo))
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	linhas := strings.Split(strings.TrimRight(string(data), "\r\n"), "\r\n")

	// 9999 = total de linhas do arquivo
	if got, want := linhas[len(linhas)-1], "|9999|"+strconv.Itoa(len(linhas))+"|"; got != want {
		t.Errorf("9999 = %q, want %q", got, want)
	}

	// 9990 = total de linhas do Bloco 9 (9001 + 9900s + 9990 + 9999)
	doBloco9, declarado := 0, ""
	for _, l := range linhas {
		reg := strings.Split(l, "|")[1]
		if strings.HasPrefix(reg, "9") {
			doBloco9++
		}
		if reg == "9990" {
			declarado = strings.Split(l, "|")[2]
		}
	}
	if declarado != strconv.Itoa(doBloco9) {
		t.Errorf("9990 declara %s, mas o Bloco 9 tem %d linhas", declarado, doBloco9)
	}
}

func TestWriteRegistro0500_ValidaNaturezaEIndicador(t *testing.T) {
	monta := func(codNat, indCta string) (panicou bool) {
		defer func() { panicou = recover() != nil }()
		b := bloco0ParaTeste(t, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), VlVersao115)
		b.Registro0001.Registro0500 = []*Registro0500{
			{CodNatCC: codNat, IndCta: indCta, CodCta: "1", NomeCta: "CAIXA"},
		}
		b.writeRegistro0500()
		return
	}

	if monta("01", "S") {
		t.Error("combinacao valida nao deveria entrar em panico")
	}
	if !monta("07", "S") {
		t.Error("COD_NAT_CC invalido deveria entrar em panico")
	}
	if !monta("01", "X") {
		t.Error("IND_CTA invalido deveria entrar em panico")
	}
}

// Os tres registros abaixo tinham writer pronto e ninguem os chamava: o
// consumidor preenchia e a linha nunca saia no arquivo. O ACBr chama E112 e
// E113 de dentro do laco do E111, e o E115 a partir do E110.
func TestBlocoE_E112_E113_E115_SaemNoArquivo(t *testing.T) {
	f := NewSPEDFiscal()
	f.Bloco0.Registro0000.CodVer = VlVersao103
	b := f.BlocoE

	e111 := &RegistroE111{CodAjApur: "SP000001", VlAjApur: 10}
	e111.RegistroE112 = []*RegistroE112{{NumDA: "DA1", NumProc: "P1"}}
	e111.RegistroE113 = []*RegistroE113{{CodPart: "PART1", CodItem: "IT1", VlAjItem: 3}}

	pai := &RegistroE110{}
	pai.RegistroE111 = []*RegistroE111{e111}
	pai.RegistroE115 = []*RegistroE115{{CodInfAdic: "SP90", VlInfAdic: 7}}

	b.writeRegistroE111(pai)
	b.writeRegistroE115(pai)

	achou := func(reg string) bool {
		for _, l := range b.Conteudo {
			if strings.HasPrefix(l, "|"+reg+"|") {
				return true
			}
		}
		return false
	}
	for _, reg := range []string{"E111", "E112", "E113", "E115"} {
		if !achou(reg) {
			t.Errorf("%s nao saiu no arquivo: %v", reg, b.Conteudo)
		}
	}
	// O E112/E113 saem logo apos a linha do E111 a que pertencem.
	if b.Conteudo[0][:6] != "|E111|" || b.Conteudo[1][:6] != "|E112|" {
		t.Errorf("ordem errada: %v", b.Conteudo)
	}
}

// No D100 o ACBr zera a chave do CT-e so quando o documento e inutilizado (05),
// e zera os valores em 02/03/04/05.
func TestWriteRegistroD100_InutilizadoZeraChaveECancelaValores(t *testing.T) {
	const chave = "35260112345678901234550010000000011000000017"
	// O parametro nulo do LFill/DFill so embranquece valor ZERO -- nunca
	// sobrescreve valor preenchido. Por isso o teste usa zeros: e a unica
	// situacao em que o efeito de booConsiderarComoValorNulo e observavel.
	monta := func(sit CodSit) string {
		f := NewSPEDFiscal()
		f.DtIni = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		b := f.BlocoD
		b.DtIni = f.DtIni
		b.RegistroD001.RegistroD100 = []*RegistroD100{{
			CodPart: "P1", CodMod: "57", CodSit: sit, NumDoc: "1", ChvCTe: chave,
		}}
		b.writeRegistroD100()
		return b.Conteudo[0]
	}

	regular := monta(SitRegular)
	if !strings.Contains(regular, chave) {
		t.Errorf("documento regular deveria manter a chave: %q", regular)
	}
	if !strings.Contains(regular, "|0,00|") {
		t.Errorf("documento regular deveria emitir 0,00 nos valores zerados: %q", regular)
	}

	// Cancelado: valores zerados saem vazios; a chave continua.
	canc := monta(SitCancelado)
	if strings.Contains(canc, "|0,00|") {
		t.Errorf("cancelado deveria embranquecer os valores zerados: %q", canc)
	}
	if !strings.Contains(canc, chave) {
		t.Errorf("cancelado nao zera a chave, so o inutilizado: %q", canc)
	}

	// Inutilizado: alem dos valores, a chave tambem sai vazia.
	inut := monta(SitNumInutilizada)
	if strings.Contains(inut, chave) {
		t.Errorf("inutilizado deveria zerar a chave: %q", inut)
	}
	if strings.Contains(inut, "|0,00|") {
		t.Errorf("inutilizado deveria embranquecer os valores zerados: %q", inut)
	}
}

// O D100 remapeia IND_FRT com corte em 01/07/2012, diferente do C100
// (01/01/2012 e 01/01/2018). Replicado como esta no ACBr.
func TestIndFrt_StringEmD100_DifereDoC100(t *testing.T) {
	maio2012 := time.Date(2012, 5, 1, 0, 0, 0, 0, time.UTC)
	// Em maio/2012 o C100 ja usa a tabela nova e o D100 ainda usa a antiga.
	if got := FrtContaEmitente.StringEm(maio2012); got != "0" {
		t.Errorf("C100 em 05/2012: esperado 0, veio %q", got)
	}
	if got := FrtContaEmitente.StringEmD100(maio2012); got != "1" {
		t.Errorf("D100 em 05/2012: esperado 1, veio %q", got)
	}

	// De 2018 em diante o C100 passa a emitir 3 e 4; o D100 nunca emite.
	jan2018 := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := FrtProprioPorContaRemetente.StringEm(jan2018); got != "3" {
		t.Errorf("C100 em 2018: esperado 3, veio %q", got)
	}
	if got := FrtProprioPorContaRemetente.StringEmD100(jan2018); got != "0" {
		t.Errorf("D100 em 2018: esperado 0, veio %q", got)
	}
}
