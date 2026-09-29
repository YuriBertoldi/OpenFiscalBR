package sped

import (
	"os"
	"path/filepath"
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
		SitRegular:       "00",
		SitCancelado:     "02",
		SitDenegado:      "04",
		SitFiscalCompl:   "06",
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
	f.Bloco0.Registro0001.IndDad = 0

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

	f.Bloco0.Registro0001.IndDad = 0
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

	f.Bloco0.Registro0001.IndDad = 0

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

	f.Bloco0.Registro0001.IndDad = 0

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

	f.Bloco0.Registro0001.IndDad = 0

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

	f.Bloco0.Registro0001.IndDad = 0

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
// OpenBlocos — IndDad semantics
// ---------------------------------------------------------------------------

func TestOpenBlocos_IndDadDefaults(t *testing.T) {
	reg := NewRegistro0001()
	if reg.IndDad != 1 {
		t.Errorf("IndDad should default to 1 (sem dados), got %d", reg.IndDad)
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
