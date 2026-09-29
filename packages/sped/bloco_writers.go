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

// ---------------------------------------------------------------------------
// Bloco Writers - Structs that hold register data and provide WriteRegistro
// methods for each SPED Fiscal block. Each Bloco embeds SPED (the base text
// writer) and holds pointers to opening, closing and detail registers, plus
// line counters for every sub-register type.
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Bloco0 - Abertura, Identificacao e Referencias
// ---------------------------------------------------------------------------

// Bloco0 holds register data and counters for Block 0.
type Bloco0 struct {
	SPED
	Registro0000 *Registro0000
	Registro0001 *Registro0001
	Registro0002 *Registro0002
	Registro0990 *Registro0990

	// Counters for sub-registers
	Registro0005Count int
	Registro0015Count int
	Registro0100Count int
	Registro0150Count int
	Registro0175Count int
	Registro0190Count int
	Registro0200Count int
	Registro0205Count int
	Registro0206Count int
	Registro0210Count int
	Registro0220Count int
	Registro0221Count int
	Registro0300Count int
	Registro0305Count int
	Registro0400Count int
	Registro0450Count int
	Registro0460Count int
	Registro0500Count int
	Registro0600Count int

	// Event callbacks
	OnBeforeWriteRegistro0000 WriteRegistroFunc
	OnWriteRegistro0000       WriteRegistroFunc
	OnAfterWriteRegistro0000  WriteRegistroFunc
}

// NewBloco0 creates a new Bloco0 with default settings.
func NewBloco0() *Bloco0 {
	b := &Bloco0{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Registro0000 = &Registro0000{}
	b.Registro0001 = NewRegistro0001()
	b.Registro0002 = &Registro0002{}
	b.Registro0990 = &Registro0990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco 0.
func (b *Bloco0) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.Registro0000 = &Registro0000{}
	b.Registro0001 = NewRegistro0001()
	b.Registro0002 = &Registro0002{}
	b.Registro0990 = &Registro0990{}
	b.Registro0005Count = 0
	b.Registro0015Count = 0
	b.Registro0100Count = 0
	b.Registro0150Count = 0
	b.Registro0175Count = 0
	b.Registro0190Count = 0
	b.Registro0200Count = 0
	b.Registro0205Count = 0
	b.Registro0206Count = 0
	b.Registro0210Count = 0
	b.Registro0220Count = 0
	b.Registro0221Count = 0
	b.Registro0300Count = 0
	b.Registro0305Count = 0
	b.Registro0400Count = 0
	b.Registro0450Count = 0
	b.Registro0460Count = 0
	b.Registro0500Count = 0
	b.Registro0600Count = 0
}

// ---------------------------------------------------------------------------
// BlocoB - Escrituracao e Apuracao do ISS
// ---------------------------------------------------------------------------

// BlocoB holds register data and counters for Block B.
type BlocoB struct {
	SPED
	Bloco0       *Bloco0
	RegistroB001 *RegistroB001
	RegistroB990 *RegistroB990

	// Counters for sub-registers
	RegistroB020Count int
	RegistroB025Count int
	RegistroB030Count int
	RegistroB035Count int
	RegistroB350Count int
	RegistroB420Count int
	RegistroB440Count int
	RegistroB460Count int
	RegistroB470Count int
	RegistroB500Count int
	RegistroB510Count int
}

// NewBlocoB creates a new BlocoB with default settings.
func NewBlocoB(bloco0 *Bloco0) *BlocoB {
	b := &BlocoB{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroB001 = NewRegistroB001()
	b.RegistroB990 = &RegistroB990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco B.
func (b *BlocoB) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroB001 = NewRegistroB001()
	b.RegistroB990 = &RegistroB990{}
	b.RegistroB020Count = 0
	b.RegistroB025Count = 0
	b.RegistroB030Count = 0
	b.RegistroB035Count = 0
	b.RegistroB350Count = 0
	b.RegistroB420Count = 0
	b.RegistroB440Count = 0
	b.RegistroB460Count = 0
	b.RegistroB470Count = 0
	b.RegistroB500Count = 0
	b.RegistroB510Count = 0
}

// ---------------------------------------------------------------------------
// BlocoC - Documentos Fiscais I - Mercadorias (ICMS/IPI)
// ---------------------------------------------------------------------------

// BlocoC holds register data and counters for Block C.
type BlocoC struct {
	SPED
	Bloco0       *Bloco0
	RegistroC001 *RegistroC001
	RegistroC990 *RegistroC990

	// Event callbacks. Sao configuracao do consumidor, nao estado de dados --
	// por isso nao sao zerados em LimpaRegistros.
	OnCheckRegistroC100 CheckRegistroFunc

	// Counters for sub-registers
	RegistroC100Count int
	RegistroC101Count int
	RegistroC105Count int
	RegistroC110Count int
	RegistroC111Count int
	RegistroC112Count int
	RegistroC113Count int
	RegistroC114Count int
	RegistroC120Count int
	RegistroC170Count int
	RegistroC185Count int
	RegistroC186Count int
	RegistroC190Count int
	RegistroC195Count int
	RegistroC197Count int
	RegistroC300Count int
	RegistroC350Count int
	RegistroC400Count int
	RegistroC405Count int
	RegistroC495Count int
	RegistroC500Count int
	RegistroC600Count int
	RegistroC700Count int
	RegistroC800Count int
	RegistroC860Count int
}

// NewBlocoC creates a new BlocoC with default settings.
func NewBlocoC(bloco0 *Bloco0) *BlocoC {
	b := &BlocoC{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroC001 = NewRegistroC001()
	b.RegistroC990 = &RegistroC990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco C.
func (b *BlocoC) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroC001 = NewRegistroC001()
	b.RegistroC990 = &RegistroC990{}
	b.RegistroC100Count = 0
	b.RegistroC101Count = 0
	b.RegistroC105Count = 0
	b.RegistroC110Count = 0
	b.RegistroC111Count = 0
	b.RegistroC112Count = 0
	b.RegistroC113Count = 0
	b.RegistroC114Count = 0
	b.RegistroC120Count = 0
	b.RegistroC170Count = 0
	b.RegistroC185Count = 0
	b.RegistroC186Count = 0
	b.RegistroC190Count = 0
	b.RegistroC195Count = 0
	b.RegistroC197Count = 0
	b.RegistroC300Count = 0
	b.RegistroC350Count = 0
	b.RegistroC400Count = 0
	b.RegistroC405Count = 0
	b.RegistroC495Count = 0
	b.RegistroC500Count = 0
	b.RegistroC600Count = 0
	b.RegistroC700Count = 0
	b.RegistroC800Count = 0
	b.RegistroC860Count = 0
}

// ---------------------------------------------------------------------------
// BlocoD - Documentos Fiscais II - Servicos (ICMS)
// ---------------------------------------------------------------------------

// BlocoD holds register data and counters for Block D.
type BlocoD struct {
	SPED
	Bloco0       *Bloco0
	RegistroD001 *RegistroD001
	RegistroD990 *RegistroD990

	// Counters for sub-registers
	RegistroD100Count int
	RegistroD101Count int
	RegistroD110Count int
	RegistroD120Count int
	RegistroD130Count int
	RegistroD140Count int
	RegistroD150Count int
	RegistroD160Count int
	RegistroD161Count int
	RegistroD162Count int
	RegistroD170Count int
	RegistroD180Count int
	RegistroD190Count int
	RegistroD195Count int
	RegistroD197Count int
	RegistroD300Count int
	RegistroD350Count int
	RegistroD400Count int
	RegistroD500Count int
	RegistroD600Count int
	RegistroD695Count int
	RegistroD696Count int
	RegistroD697Count int
	RegistroD700Count int
	RegistroD730Count int
	RegistroD731Count int
	RegistroD735Count int
	RegistroD737Count int
	RegistroD750Count int
	RegistroD760Count int
	RegistroD761Count int
}

// NewBlocoD creates a new BlocoD with default settings.
func NewBlocoD(bloco0 *Bloco0) *BlocoD {
	b := &BlocoD{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroD001 = NewRegistroD001()
	b.RegistroD990 = &RegistroD990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco D.
func (b *BlocoD) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroD001 = NewRegistroD001()
	b.RegistroD990 = &RegistroD990{}
	b.RegistroD100Count = 0
	b.RegistroD101Count = 0
	b.RegistroD110Count = 0
	b.RegistroD120Count = 0
	b.RegistroD130Count = 0
	b.RegistroD140Count = 0
	b.RegistroD150Count = 0
	b.RegistroD160Count = 0
	b.RegistroD161Count = 0
	b.RegistroD162Count = 0
	b.RegistroD170Count = 0
	b.RegistroD180Count = 0
	b.RegistroD190Count = 0
	b.RegistroD195Count = 0
	b.RegistroD197Count = 0
	b.RegistroD300Count = 0
	b.RegistroD350Count = 0
	b.RegistroD400Count = 0
	b.RegistroD500Count = 0
	b.RegistroD600Count = 0
	b.RegistroD695Count = 0
	b.RegistroD696Count = 0
	b.RegistroD697Count = 0
	b.RegistroD700Count = 0
	b.RegistroD730Count = 0
	b.RegistroD731Count = 0
	b.RegistroD735Count = 0
	b.RegistroD737Count = 0
	b.RegistroD750Count = 0
	b.RegistroD760Count = 0
	b.RegistroD761Count = 0
}

// ---------------------------------------------------------------------------
// BlocoE - Apuracao do ICMS e do IPI
// ---------------------------------------------------------------------------

// BlocoE holds register data and counters for Block E.
type BlocoE struct {
	SPED
	Bloco0       *Bloco0
	RegistroE001 *RegistroE001
	RegistroE990 *RegistroE990

	// Counters for sub-registers
	RegistroE100Count int
	RegistroE110Count int
	RegistroE111Count int
	RegistroE112Count int
	RegistroE113Count int
	RegistroE115Count int
	RegistroE116Count int
	RegistroE200Count int
	RegistroE210Count int
	RegistroE220Count int
	RegistroE230Count int
	RegistroE240Count int
	RegistroE250Count int
	RegistroE300Count int
	RegistroE310Count int
	RegistroE311Count int
	RegistroE312Count int
	RegistroE313Count int
	RegistroE316Count int
	RegistroE500Count int
	RegistroE510Count int
	RegistroE520Count int
	RegistroE530Count int
	RegistroE531Count int
}

// NewBlocoE creates a new BlocoE with default settings.
func NewBlocoE(bloco0 *Bloco0) *BlocoE {
	b := &BlocoE{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroE001 = NewRegistroE001()
	b.RegistroE990 = &RegistroE990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco E.
func (b *BlocoE) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroE001 = NewRegistroE001()
	b.RegistroE990 = &RegistroE990{}
	b.RegistroE100Count = 0
	b.RegistroE110Count = 0
	b.RegistroE111Count = 0
	b.RegistroE112Count = 0
	b.RegistroE113Count = 0
	b.RegistroE115Count = 0
	b.RegistroE116Count = 0
	b.RegistroE200Count = 0
	b.RegistroE210Count = 0
	b.RegistroE220Count = 0
	b.RegistroE230Count = 0
	b.RegistroE240Count = 0
	b.RegistroE250Count = 0
	b.RegistroE300Count = 0
	b.RegistroE310Count = 0
	b.RegistroE311Count = 0
	b.RegistroE312Count = 0
	b.RegistroE313Count = 0
	b.RegistroE316Count = 0
	b.RegistroE500Count = 0
	b.RegistroE510Count = 0
	b.RegistroE520Count = 0
	b.RegistroE530Count = 0
	b.RegistroE531Count = 0
}

// ---------------------------------------------------------------------------
// BlocoG - Controle do Credito de ICMS do Ativo Permanente (CIAP)
// ---------------------------------------------------------------------------

// BlocoG holds register data and counters for Block G.
type BlocoG struct {
	SPED
	Bloco0       *Bloco0
	RegistroG001 *RegistroG001
	RegistroG990 *RegistroG990

	// Counters for sub-registers
	RegistroG110Count int
	RegistroG125Count int
	RegistroG126Count int
	RegistroG130Count int
	RegistroG140Count int
}

// NewBlocoG creates a new BlocoG with default settings.
func NewBlocoG(bloco0 *Bloco0) *BlocoG {
	b := &BlocoG{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroG001 = NewRegistroG001()
	b.RegistroG990 = &RegistroG990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco G.
func (b *BlocoG) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroG001 = NewRegistroG001()
	b.RegistroG990 = &RegistroG990{}
	b.RegistroG110Count = 0
	b.RegistroG125Count = 0
	b.RegistroG126Count = 0
	b.RegistroG130Count = 0
	b.RegistroG140Count = 0
}

// ---------------------------------------------------------------------------
// BlocoH - Inventario Fisico
// ---------------------------------------------------------------------------

// BlocoH holds register data and counters for Block H.
type BlocoH struct {
	SPED
	Bloco0       *Bloco0
	RegistroH001 *RegistroH001
	RegistroH990 *RegistroH990

	// Counters for sub-registers
	RegistroH005Count int
	RegistroH010Count int
	RegistroH011Count int
	RegistroH020Count int
	RegistroH030Count int
}

// NewBlocoH creates a new BlocoH with default settings.
func NewBlocoH(bloco0 *Bloco0) *BlocoH {
	b := &BlocoH{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroH001 = NewRegistroH001()
	b.RegistroH990 = &RegistroH990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco H.
func (b *BlocoH) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroH001 = NewRegistroH001()
	b.RegistroH990 = &RegistroH990{}
	b.RegistroH005Count = 0
	b.RegistroH010Count = 0
	b.RegistroH011Count = 0
	b.RegistroH020Count = 0
	b.RegistroH030Count = 0
}

// ---------------------------------------------------------------------------
// BlocoK - Controle da Producao e do Estoque
// ---------------------------------------------------------------------------

// BlocoK holds register data and counters for Block K.
type BlocoK struct {
	SPED
	Bloco0       *Bloco0
	RegistroK001 *RegistroK001
	RegistroK990 *RegistroK990

	// Counters for sub-registers
	RegistroK010Count int
	RegistroK100Count int
	RegistroK200Count int
	RegistroK210Count int
	RegistroK215Count int
	RegistroK220Count int
	RegistroK230Count int
	RegistroK235Count int
	RegistroK250Count int
	RegistroK255Count int
	RegistroK260Count int
	RegistroK265Count int
	RegistroK270Count int
	RegistroK275Count int
	RegistroK280Count int
	RegistroK290Count int
	RegistroK291Count int
	RegistroK292Count int
	RegistroK300Count int
	RegistroK301Count int
	RegistroK302Count int
}

// NewBlocoK creates a new BlocoK with default settings.
func NewBlocoK(bloco0 *Bloco0) *BlocoK {
	b := &BlocoK{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.RegistroK001 = NewRegistroK001()
	b.RegistroK990 = &RegistroK990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco K.
func (b *BlocoK) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.RegistroK001 = NewRegistroK001()
	b.RegistroK990 = &RegistroK990{}
	b.RegistroK010Count = 0
	b.RegistroK100Count = 0
	b.RegistroK200Count = 0
	b.RegistroK210Count = 0
	b.RegistroK215Count = 0
	b.RegistroK220Count = 0
	b.RegistroK230Count = 0
	b.RegistroK235Count = 0
	b.RegistroK250Count = 0
	b.RegistroK255Count = 0
	b.RegistroK260Count = 0
	b.RegistroK265Count = 0
	b.RegistroK270Count = 0
	b.RegistroK275Count = 0
	b.RegistroK280Count = 0
	b.RegistroK290Count = 0
	b.RegistroK291Count = 0
	b.RegistroK292Count = 0
	b.RegistroK300Count = 0
	b.RegistroK301Count = 0
	b.RegistroK302Count = 0
}

// ---------------------------------------------------------------------------
// Bloco1 - Outras Informacoes
// ---------------------------------------------------------------------------

// Bloco1 holds register data and counters for Block 1.
type Bloco1 struct {
	SPED
	Bloco0       *Bloco0
	Registro1001 *Registro1001
	Registro1990 *Registro1990

	// Counters for sub-registers
	Registro1010Count int
	Registro1100Count int
	Registro1105Count int
	Registro1110Count int
	Registro1200Count int
	Registro1210Count int
	Registro1250Count int
	Registro1255Count int
	Registro1300Count int
	Registro1310Count int
	Registro1320Count int
	Registro1350Count int
	Registro1360Count int
	Registro1370Count int
	Registro1390Count int
	Registro1391Count int
	Registro1400Count int
	Registro1500Count int
	Registro1600Count int
	Registro1601Count int
	Registro1700Count int
	Registro1710Count int
	Registro1800Count int
	Registro1900Count int
	Registro1910Count int
	Registro1920Count int
	Registro1921Count int
	Registro1925Count int
	Registro1926Count int
	Registro1960Count int
	Registro1970Count int
	Registro1975Count int
	Registro1980Count int
}

// NewBloco1 creates a new Bloco1 with default settings.
func NewBloco1(bloco0 *Bloco0) *Bloco1 {
	b := &Bloco1{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Bloco0 = bloco0
	b.Registro1001 = NewRegistro1001()
	b.Registro1990 = &Registro1990{}
	return b
}

// LimpaRegistros clears all registers and resets counters for Bloco 1.
func (b *Bloco1) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.Registro1001 = NewRegistro1001()
	b.Registro1990 = &Registro1990{}
	b.Registro1010Count = 0
	b.Registro1100Count = 0
	b.Registro1105Count = 0
	b.Registro1110Count = 0
	b.Registro1200Count = 0
	b.Registro1210Count = 0
	b.Registro1250Count = 0
	b.Registro1255Count = 0
	b.Registro1300Count = 0
	b.Registro1310Count = 0
	b.Registro1320Count = 0
	b.Registro1350Count = 0
	b.Registro1360Count = 0
	b.Registro1370Count = 0
	b.Registro1390Count = 0
	b.Registro1391Count = 0
	b.Registro1400Count = 0
	b.Registro1500Count = 0
	b.Registro1600Count = 0
	b.Registro1601Count = 0
	b.Registro1700Count = 0
	b.Registro1710Count = 0
	b.Registro1800Count = 0
	b.Registro1900Count = 0
	b.Registro1910Count = 0
	b.Registro1920Count = 0
	b.Registro1921Count = 0
	b.Registro1925Count = 0
	b.Registro1926Count = 0
	b.Registro1960Count = 0
	b.Registro1970Count = 0
	b.Registro1975Count = 0
	b.Registro1980Count = 0
}

// ---------------------------------------------------------------------------
// Bloco9 - Controle e Encerramento do Arquivo Digital
// ---------------------------------------------------------------------------

// Bloco9 holds register data for the control block.
// This block is special: it contains register counts for the entire file
// via Registro9900 entries and the final file-level closing register 9999.
type Bloco9 struct {
	SPED
	Registro9001 *Registro9001
	Registro9900 []*Registro9900
	Registro9990 *Registro9990
	Registro9999 *Registro9999
}

// NewBloco9 creates a new Bloco9 with default settings.
func NewBloco9() *Bloco9 {
	b := &Bloco9{}
	b.Delimitador = "|"
	b.TrimString = true
	b.Registro9001 = NewRegistro9001()
	b.Registro9990 = &Registro9990{}
	b.Registro9999 = &Registro9999{}
	return b
}

// LimpaRegistros clears all registers for Bloco 9.
func (b *Bloco9) LimpaRegistros() {
	b.SPED.LimpaRegistros()
	b.Registro9001 = NewRegistro9001()
	b.Registro9900 = nil
	b.Registro9990 = &Registro9990{}
	b.Registro9999 = &Registro9999{}
}
