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
	"path/filepath"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/comum"
)

// ---------------------------------------------------------------------------
// SPEDFiscal - Componente principal do SPED Fiscal (EFD-ICMS/IPI)
// Portado de: TACBrSPEDFiscal (ACBrSpedFiscal.pas)
// ---------------------------------------------------------------------------

// SPEDFiscal is the main SPED Fiscal component. It orchestrates the
// generation of the EFD-ICMS/IPI digital tax file by coordinating the
// writing of all 10 blocks in the correct order.
type SPEDFiscal struct {
	comum.TXTClass

	// Block writers — one per SPED block.
	Bloco0 *Bloco0
	Bloco1 *Bloco1
	BlocoB *BlocoB
	BlocoC *BlocoC
	BlocoD *BlocoD
	BlocoE *BlocoE
	BlocoG *BlocoG
	BlocoH *BlocoH
	BlocoK *BlocoK
	Bloco9 *Bloco9

	// Period covered by this SPED file.
	DtIni time.Time
	DtFin time.Time

	// Path is the output directory for the generated file.
	Path string

	// Arquivo is the output file name (without directory).
	Arquivo string

	// Inicializado indicates whether IniciaGeracao was successfully called.
	Inicializado bool

	// OnError is an optional callback for error notifications.
	OnError ErrorFunc
}

// NewSPEDFiscal creates a new SPEDFiscal instance with all block writers
// allocated and sensible defaults configured.
func NewSPEDFiscal() *SPEDFiscal {
	f := &SPEDFiscal{}

	// Main writer defaults.
	f.Delimitador = "|"
	f.TrimString = true
	f.LinhasBuffer = 1000
	f.CurMascara = "#0.00"

	// Create all block instances.
	f.Bloco0 = NewBloco0()
	f.BlocoB = NewBlocoB(f.Bloco0)
	f.BlocoC = NewBlocoC(f.Bloco0)
	f.BlocoD = NewBlocoD(f.Bloco0)
	f.BlocoE = NewBlocoE(f.Bloco0)
	f.BlocoG = NewBlocoG(f.Bloco0)
	f.BlocoH = NewBlocoH(f.Bloco0)
	f.BlocoK = NewBlocoK(f.Bloco0)
	f.Bloco1 = NewBloco1(f.Bloco0)
	f.Bloco9 = NewBloco9()

	return f
}

// ---------------------------------------------------------------------------
// Propagation setters — mirror the Delphi property setters that propagate
// values from the component to every block writer.
// ---------------------------------------------------------------------------

// allBlocoSPEDs returns pointers to the embedded SPED of every block,
// in the canonical block order.
func (f *SPEDFiscal) allBlocoSPEDs() []*SPED {
	return []*SPED{
		&f.Bloco0.SPED,
		&f.BlocoB.SPED,
		&f.BlocoC.SPED,
		&f.BlocoD.SPED,
		&f.BlocoE.SPED,
		&f.BlocoG.SPED,
		&f.BlocoH.SPED,
		&f.BlocoK.SPED,
		&f.Bloco1.SPED,
		&f.Bloco9.SPED,
	}
}

// SetDelimitador propagates the delimiter to all blocks.
func (f *SPEDFiscal) SetDelimitador(v string) {
	f.Delimitador = v
	for _, s := range f.allBlocoSPEDs() {
		s.Delimitador = v
	}
}

// SetTrimString propagates TrimString to all blocks.
func (f *SPEDFiscal) SetTrimString(v bool) {
	f.TrimString = v
	for _, s := range f.allBlocoSPEDs() {
		s.TrimString = v
	}
}

// SetCurMascara propagates the currency mask to all blocks.
func (f *SPEDFiscal) SetCurMascara(v string) {
	f.CurMascara = v
	for _, s := range f.allBlocoSPEDs() {
		s.CurMascara = v
	}
}

// SetReplaceDelimitador propagates ReplaceDelimitador to all blocks.
func (f *SPEDFiscal) SetReplaceDelimitador(v bool) {
	f.ReplaceDelimitador = v
	for _, s := range f.allBlocoSPEDs() {
		s.ReplaceDelimitador = v
	}
}

// SetDtIni sets the initial date on all blocks and on Registro0000.
func (f *SPEDFiscal) SetDtIni(v time.Time) {
	f.DtIni = v
	for _, s := range f.allBlocoSPEDs() {
		s.DtIni = v
	}
	f.Bloco0.Registro0000.DtIni = v
}

// SetDtFin sets the final date on all blocks and on Registro0000.
func (f *SPEDFiscal) SetDtFin(v time.Time) {
	f.DtFin = v
	for _, s := range f.allBlocoSPEDs() {
		s.DtFin = v
	}
	f.Bloco0.Registro0000.DtFin = v
}

// ---------------------------------------------------------------------------
// Generation lifecycle
// ---------------------------------------------------------------------------

// IniciaGeracao validates the dates, prepares the output file, resets all
// blocks and zeroes their line counters. Must be called before writing.
func (f *SPEDFiscal) IniciaGeracao() error {
	// DtIni must be the 1st day of the month.
	if f.DtIni.Day() != 1 {
		return &SPEDFiscalError{"DT_INI deve ser o primeiro dia do mes"}
	}

	// DtFin must be the last day of the month.
	lastDay := lastDayOfMonth(f.DtFin)
	if f.DtFin.Day() != lastDay {
		return &SPEDFiscalError{"DT_FIN deve ser o ultimo dia do mes"}
	}

	// DtFin must be >= DtIni.
	if f.DtFin.Before(f.DtIni) {
		return &SPEDFiscalError{"DT_FIN deve ser maior ou igual a DT_INI"}
	}

	// Build the output file name.
	arquivo := f.Arquivo
	if arquivo == "" {
		arquivo = "SpedFiscal.txt"
	}
	if f.Path != "" {
		f.NomeArquivo = filepath.Join(f.Path, arquivo)
	} else {
		f.NomeArquivo = arquivo
	}

	// Reset the main writer (removes old file, clears buffer).
	f.Reset()

	// Initialize each block.
	f.inicializaBloco(&f.Bloco0.SPED)
	f.inicializaBloco(&f.BlocoB.SPED)
	f.inicializaBloco(&f.BlocoC.SPED)
	f.inicializaBloco(&f.BlocoD.SPED)
	f.inicializaBloco(&f.BlocoE.SPED)
	f.inicializaBloco(&f.BlocoG.SPED)
	f.inicializaBloco(&f.BlocoH.SPED)
	f.inicializaBloco(&f.BlocoK.SPED)
	f.inicializaBloco(&f.Bloco1.SPED)
	f.inicializaBloco(&f.Bloco9.SPED)

	// Zero all line counters.
	f.Bloco0.Registro0990.QtdLin0 = 0
	f.BlocoB.RegistroB990.QtdLinB = 0
	f.BlocoC.RegistroC990.QtdLinC = 0
	f.BlocoD.RegistroD990.QtdLinD = 0
	f.BlocoE.RegistroE990.QtdLinE = 0
	f.BlocoG.RegistroG990.QtdLinG = 0
	f.BlocoH.RegistroH990.QtdLinH = 0
	f.BlocoK.RegistroK990.QtdLinK = 0
	f.Bloco1.Registro1990.QtdLin1 = 0
	f.Bloco9.Registro9990.QtdLin9 = 0

	// Clear Bloco9 register list.
	f.Bloco9.Registro9900 = nil

	f.Inicializado = true
	return nil
}

// CancelaGeracao aborts the current generation, clearing all registers
// and marking the component as not initialized.
func (f *SPEDFiscal) CancelaGeracao() {
	f.LimpaRegistros()
	f.Inicializado = false
}

// LimpaRegistros clears all registers in every block.
func (f *SPEDFiscal) LimpaRegistros() {
	f.Bloco0.LimpaRegistros()
	f.BlocoB.LimpaRegistros()
	f.BlocoC.LimpaRegistros()
	f.BlocoD.LimpaRegistros()
	f.BlocoE.LimpaRegistros()
	f.BlocoG.LimpaRegistros()
	f.BlocoH.LimpaRegistros()
	f.BlocoK.LimpaRegistros()
	f.Bloco1.LimpaRegistros()
	f.Bloco9.LimpaRegistros()
}

// inicializaBloco prepares a single block for writing.
func (f *SPEDFiscal) inicializaBloco(b *SPED) {
	b.NomeArquivo = f.NomeArquivo
	b.LinhasBuffer = f.LinhasBuffer
	b.Gravado = false
	b.Conteudo = nil
	// DT_INI/DT_FIN precisam chegar aos blocos: varios writers decidem o layout
	// por vigencia (IND_FRT, IND_PGTO, Registro 0002). O setter SetDtIni propaga,
	// mas a atribuicao direta em f.DtIni -- usada pelo exemplo do README e pelos
	// testes -- nao passa por ele, e o bloco ficava com a data zerada.
	b.DtIni = f.DtIni
	b.DtFin = f.DtFin
}

// ---------------------------------------------------------------------------
// SaveFileTXT — main entry point to generate the complete SPED Fiscal file.
// ---------------------------------------------------------------------------

// SaveFileTXT generates the full SPED Fiscal TXT file by calling
// IniciaGeracao, writing all blocks in order, and cleaning up.
func (f *SPEDFiscal) SaveFileTXT() error {
	if err := f.IniciaGeracao(); err != nil {
		return err
	}
	defer f.LimpaRegistros()

	if err := f.WriteBloco0(); err != nil {
		return err
	}
	if err := f.WriteBlocoB(); err != nil {
		return err
	}
	if err := f.WriteBlocoC(true); err != nil {
		return err
	}
	if err := f.WriteBlocoD(); err != nil {
		return err
	}
	if err := f.WriteBlocoE(); err != nil {
		return err
	}
	if err := f.WriteBlocoG(); err != nil {
		return err
	}
	if err := f.WriteBlocoH(); err != nil {
		return err
	}
	if err := f.WriteBlocoK(); err != nil {
		return err
	}
	if err := f.WriteBloco1(); err != nil {
		return err
	}
	if err := f.WriteBloco9(); err != nil {
		return err
	}

	return nil
}

// ---------------------------------------------------------------------------
// WriteBloco methods — one per block, following the chain-dependency pattern.
// ---------------------------------------------------------------------------

// WriteBloco0 writes Block 0 (Abertura, Identificacao e Referencias).
func (f *SPEDFiscal) WriteBloco0() error {
	if f.Bloco0.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	f.Bloco0.WriteRegistro0000()
	f.Bloco0.WriteRegistro0001()
	f.Bloco0.WriteRegistro0990()
	f.Bloco0.WriteBuffer()
	f.Bloco0.Conteudo = nil
	f.Bloco0.Gravado = true
	return nil
}

// WriteBlocoB writes Block B (Escrituracao e Apuracao do ISS).
// Only writes if DtIni >= 2019-01-01.
func (f *SPEDFiscal) WriteBlocoB() error {
	if f.BlocoB.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: Bloco0 must be written first.
	if !f.Bloco0.Gravado {
		if err := f.WriteBloco0(); err != nil {
			return err
		}
	}

	// BlocoB only applies from 2019-01-01 onwards. Fora da vigencia o bloco nao e
	// escrito -- nem mesmo B001/B990 (ver ACBrSpedFiscal.pas, WriteBloco_B).
	blocoB_inicio := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.DtIni.Before(blocoB_inicio) {
		f.BlocoB.WriteRegistroB001()
		f.BlocoB.WriteRegistroB990()
		f.BlocoB.WriteBuffer()
	}
	f.BlocoB.Conteudo = nil
	f.BlocoB.Gravado = true
	return nil
}

// WriteBlocoC writes Block C (Documentos Fiscais I - Mercadorias).
// If the block has no data (RegistroC001.IndMov==1), fechaBloco is forced true.
func (f *SPEDFiscal) WriteBlocoC(fechaBloco bool) error {
	if f.BlocoC.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoB must be written first.
	if !f.BlocoB.Gravado {
		if err := f.WriteBlocoB(); err != nil {
			return err
		}
	}

	// If no data, force block closure.
	if f.BlocoC.RegistroC001.IndMov == 1 {
		fechaBloco = true
	}

	f.BlocoC.WriteRegistroC001()
	if fechaBloco {
		f.BlocoC.WriteRegistroC990()
	}
	f.BlocoC.WriteBuffer()
	f.BlocoC.Conteudo = nil
	f.BlocoC.Gravado = fechaBloco
	return nil
}

// WriteBlocoD writes Block D (Documentos Fiscais II - Servicos).
func (f *SPEDFiscal) WriteBlocoD() error {
	if f.BlocoD.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoC must be written first.
	if !f.BlocoC.Gravado {
		if err := f.WriteBlocoC(true); err != nil {
			return err
		}
	}

	f.BlocoD.WriteRegistroD001()
	f.BlocoD.WriteRegistroD990()
	f.BlocoD.WriteBuffer()
	f.BlocoD.Conteudo = nil
	f.BlocoD.Gravado = true
	return nil
}

// WriteBlocoE writes Block E (Apuracao do ICMS e do IPI).
func (f *SPEDFiscal) WriteBlocoE() error {
	if f.BlocoE.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoD must be written first.
	if !f.BlocoD.Gravado {
		if err := f.WriteBlocoD(); err != nil {
			return err
		}
	}

	f.BlocoE.WriteRegistroE001()
	f.BlocoE.WriteRegistroE990()
	f.BlocoE.WriteBuffer()
	f.BlocoE.Conteudo = nil
	f.BlocoE.Gravado = true
	return nil
}

// WriteBlocoG writes Block G (Controle do Credito de ICMS do Ativo Permanente).
// Only writes if DtIni >= 2011-01-01.
func (f *SPEDFiscal) WriteBlocoG() error {
	if f.BlocoG.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoE must be written first.
	if !f.BlocoE.Gravado {
		if err := f.WriteBlocoE(); err != nil {
			return err
		}
	}

	// BlocoG only applies from 2011-01-01 onwards. Fora da vigencia o bloco nao e
	// escrito -- nem mesmo G001/G990 (ver ACBrSpedFiscal.pas, WriteBloco_G).
	blocoG_inicio := time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.DtIni.Before(blocoG_inicio) {
		f.BlocoG.WriteRegistroG001()
		f.BlocoG.WriteRegistroG990()
		f.BlocoG.WriteBuffer()
	}
	f.BlocoG.Conteudo = nil
	f.BlocoG.Gravado = true
	return nil
}

// WriteBlocoH writes Block H (Inventario Fisico).
func (f *SPEDFiscal) WriteBlocoH() error {
	if f.BlocoH.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoG must be written first.
	if !f.BlocoG.Gravado {
		if err := f.WriteBlocoG(); err != nil {
			return err
		}
	}

	f.BlocoH.WriteRegistroH001()
	f.BlocoH.WriteRegistroH990()
	f.BlocoH.WriteBuffer()
	f.BlocoH.Conteudo = nil
	f.BlocoH.Gravado = true
	return nil
}

// WriteBlocoK writes Block K (Controle da Producao e do Estoque).
// Only writes if DtIni >= 2016-01-01.
func (f *SPEDFiscal) WriteBlocoK() error {
	if f.BlocoK.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoH must be written first.
	if !f.BlocoH.Gravado {
		if err := f.WriteBlocoH(); err != nil {
			return err
		}
	}

	// BlocoK only applies from 2016-01-01 onwards. Fora da vigencia o bloco nao e
	// escrito -- nem mesmo K001/K990 (ver ACBrSpedFiscal.pas, WriteBloco_K).
	blocoK_inicio := time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)
	if !f.DtIni.Before(blocoK_inicio) {
		f.BlocoK.WriteRegistroK001()
		f.BlocoK.WriteRegistroK990()
		f.BlocoK.WriteBuffer()
	}
	f.BlocoK.Conteudo = nil
	f.BlocoK.Gravado = true
	return nil
}

// WriteBloco1 writes Block 1 (Outras Informacoes).
func (f *SPEDFiscal) WriteBloco1() error {
	if f.Bloco1.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: BlocoK must be written first.
	if !f.BlocoK.Gravado {
		if err := f.WriteBlocoK(); err != nil {
			return err
		}
	}

	f.Bloco1.WriteRegistro1001()
	f.Bloco1.WriteRegistro1990()
	f.Bloco1.WriteBuffer()
	f.Bloco1.Conteudo = nil
	f.Bloco1.Gravado = true
	return nil
}

// WriteBloco9 writes Block 9 (Controle e Encerramento do Arquivo Digital).
// Per SPED legislation, Block 9 contains:
//   - 9001: Opening
//   - 9900: One entry per register type used, with line counts
//   - 9990: Block 9 line count
//   - 9999: Total file line count
func (f *SPEDFiscal) WriteBloco9() error {
	if f.Bloco9.Gravado {
		return nil
	}
	if !f.Inicializado {
		return &SPEDFiscalError{"IniciaGeracao nao foi executado"}
	}

	// Chain: Bloco1 must be written first.
	if !f.Bloco1.Gravado {
		if err := f.WriteBloco1(); err != nil {
			return err
		}
	}

	// Populate Registro 9900 entries — one per register type with line count.
	// This mirrors Delphi's WriteBloco_9 which enumerates all blocks' counters.
	f.populateRegistro9900()

	// Calculate total lines across all blocks + block 9 itself.
	// Block 9 will have: 9001 + N×9900 + 9990 + 9999 = 2 + N + 1 = N+3 lines
	// where N = number of 9900 entries.
	n9900 := len(f.Bloco9.Registro9900)
	bloco9Lines := n9900 + 3 // 9001 + n×9900 + 9990 + 9999

	// Add 9900 entry for 9900 itself, 9001, 9990, 9999
	f.addRegistro9900("9001", 1)
	f.addRegistro9900("9900", n9900+4) // includes entries for 9001,9900,9990,9999
	f.addRegistro9900("9990", 1)
	f.addRegistro9900("9999", 1)
	bloco9Lines += 4 // the 4 entries we just added

	// Total line count for the whole file.
	// QtdLin0 already includes Registro 0000 (incremented in WriteRegistro0000).
	totalLines := f.Bloco0.Registro0990.QtdLin0 +
		f.BlocoB.RegistroB990.QtdLinB +
		f.BlocoC.RegistroC990.QtdLinC +
		f.BlocoD.RegistroD990.QtdLinD +
		f.BlocoE.RegistroE990.QtdLinE +
		f.BlocoG.RegistroG990.QtdLinG +
		f.BlocoH.RegistroH990.QtdLinH +
		f.BlocoK.RegistroK990.QtdLinK +
		f.Bloco1.Registro1990.QtdLin1 +
		bloco9Lines

	f.Bloco9.Registro9999.QtdLin = totalLines

	f.Bloco9.WriteRegistro9001()
	f.Bloco9.WriteRegistro9900()
	f.Bloco9.WriteRegistro9990()
	f.Bloco9.WriteRegistro9999()
	f.Bloco9.WriteBuffer()
	f.Bloco9.Conteudo = nil
	f.Bloco9.Gravado = true
	return nil
}

// populateRegistro9900 adds 9900 entries for all register types used across
// all blocks. Per SPED Guia Pratico, each register type that appears in the
// file must have a corresponding 9900 entry with its line count.
//
// A contagem vem sempre dos contadores RegistroXXXCount, alimentados pelos
// writers -- nunca de len(slice). Um len() nao alcanca registro neto (o 0175
// pertence a cada 0150, o 0205 a cada 0200), que ficaria contado a menos.
// Mesmo criterio do ACBr original (ACBrSpedFiscal.pas, WriteRegistro9900).
//
// A abertura e o encerramento de cada bloco so entram quando o bloco foi de
// fato escrito (QtdLinX > 0). Bloco fora de vigencia nao gera linha alguma e
// portanto nao pode aparecer no 9900.
func (f *SPEDFiscal) populateRegistro9900() {
	add := f.addRegistro9900
	addIf := func(reg string, qtd int) {
		if qtd > 0 {
			add(reg, qtd)
		}
	}

	// Block 0
	if b := f.Bloco0; b.Registro0990.QtdLin0 > 0 {
		add("0000", 1)
		add("0001", 1)
		addIf("0005", b.Registro0005Count)
		addIf("0015", b.Registro0015Count)
		addIf("0100", b.Registro0100Count)
		addIf("0150", b.Registro0150Count)
		addIf("0175", b.Registro0175Count)
		addIf("0190", b.Registro0190Count)
		addIf("0200", b.Registro0200Count)
		addIf("0205", b.Registro0205Count)
		addIf("0206", b.Registro0206Count)
		addIf("0210", b.Registro0210Count)
		addIf("0220", b.Registro0220Count)
		addIf("0221", b.Registro0221Count)
		addIf("0300", b.Registro0300Count)
		addIf("0305", b.Registro0305Count)
		addIf("0400", b.Registro0400Count)
		addIf("0450", b.Registro0450Count)
		addIf("0460", b.Registro0460Count)
		addIf("0500", b.Registro0500Count)
		addIf("0600", b.Registro0600Count)
		add("0990", 1)
	}

	// Block B
	if b := f.BlocoB; b.RegistroB990.QtdLinB > 0 {
		add("B001", 1)
		addIf("B020", b.RegistroB020Count)
		addIf("B025", b.RegistroB025Count)
		addIf("B030", b.RegistroB030Count)
		// O writer do B350 emite a linha com o literal "B035", por fidelidade ao
		// ACBr; a contagem do 9900 acompanha o que foi de fato escrito.
		addIf("B035", b.RegistroB035Count+b.RegistroB350Count)
		addIf("B420", b.RegistroB420Count)
		addIf("B440", b.RegistroB440Count)
		addIf("B460", b.RegistroB460Count)
		addIf("B470", b.RegistroB470Count)
		addIf("B500", b.RegistroB500Count)
		addIf("B510", b.RegistroB510Count)
		add("B990", 1)
	}

	// Block C
	if b := f.BlocoC; b.RegistroC990.QtdLinC > 0 {
		add("C001", 1)
		addIf("C100", b.RegistroC100Count)
		addIf("C101", b.RegistroC101Count)
		addIf("C110", b.RegistroC110Count)
		addIf("C111", b.RegistroC111Count)
		addIf("C112", b.RegistroC112Count)
		addIf("C113", b.RegistroC113Count)
		addIf("C114", b.RegistroC114Count)
		addIf("C170", b.RegistroC170Count)
		addIf("C190", b.RegistroC190Count)
		addIf("C105", b.RegistroC105Count)
		addIf("C115", b.RegistroC115Count)
		addIf("C116", b.RegistroC116Count)
		addIf("C120", b.RegistroC120Count)
		addIf("C130", b.RegistroC130Count)
		addIf("C140", b.RegistroC140Count)
		addIf("C141", b.RegistroC141Count)
		addIf("C160", b.RegistroC160Count)
		addIf("C165", b.RegistroC165Count)
		addIf("C171", b.RegistroC171Count)
		addIf("C172", b.RegistroC172Count)
		addIf("C173", b.RegistroC173Count)
		addIf("C174", b.RegistroC174Count)
		addIf("C175", b.RegistroC175Count)
		addIf("C176", b.RegistroC176Count)
		addIf("C177", b.RegistroC177Count)
		addIf("C178", b.RegistroC178Count)
		addIf("C179", b.RegistroC179Count)
		addIf("C180", b.RegistroC180Count)
		addIf("C181", b.RegistroC181Count)
		addIf("C185", b.RegistroC185Count)
		addIf("C186", b.RegistroC186Count)
		addIf("C191", b.RegistroC191Count)
		addIf("C195", b.RegistroC195Count)
		addIf("C197", b.RegistroC197Count)
		addIf("C300", b.RegistroC300Count)
		addIf("C310", b.RegistroC310Count)
		addIf("C320", b.RegistroC320Count)
		addIf("C321", b.RegistroC321Count)
		addIf("C330", b.RegistroC330Count)
		addIf("C350", b.RegistroC350Count)
		addIf("C370", b.RegistroC370Count)
		addIf("C380", b.RegistroC380Count)
		addIf("C390", b.RegistroC390Count)
		addIf("C400", b.RegistroC400Count)
		addIf("C405", b.RegistroC405Count)
		addIf("C410", b.RegistroC410Count)
		addIf("C420", b.RegistroC420Count)
		addIf("C425", b.RegistroC425Count)
		addIf("C430", b.RegistroC430Count)
		addIf("C460", b.RegistroC460Count)
		addIf("C465", b.RegistroC465Count)
		addIf("C470", b.RegistroC470Count)
		addIf("C480", b.RegistroC480Count)
		addIf("C490", b.RegistroC490Count)
		addIf("C495", b.RegistroC495Count)
		addIf("C500", b.RegistroC500Count)
		addIf("C510", b.RegistroC510Count)
		addIf("C590", b.RegistroC590Count)
		addIf("C591", b.RegistroC591Count)
		addIf("C595", b.RegistroC595Count)
		addIf("C597", b.RegistroC597Count)
		addIf("C600", b.RegistroC600Count)
		addIf("C601", b.RegistroC601Count)
		addIf("C610", b.RegistroC610Count)
		addIf("C690", b.RegistroC690Count)
		addIf("C700", b.RegistroC700Count)
		addIf("C790", b.RegistroC790Count)
		addIf("C791", b.RegistroC791Count)
		addIf("C800", b.RegistroC800Count)
		addIf("C810", b.RegistroC810Count)
		addIf("C815", b.RegistroC815Count)
		addIf("C850", b.RegistroC850Count)
		addIf("C855", b.RegistroC855Count)
		addIf("C857", b.RegistroC857Count)
		addIf("C860", b.RegistroC860Count)
		addIf("C870", b.RegistroC870Count)
		addIf("C880", b.RegistroC880Count)
		addIf("C890", b.RegistroC890Count)
		addIf("C895", b.RegistroC895Count)
		addIf("C897", b.RegistroC897Count)
		add("C990", 1)
	}

	// Block D
	if b := f.BlocoD; b.RegistroD990.QtdLinD > 0 {
		add("D001", 1)
		addIf("D101", b.RegistroD101Count)
		addIf("D110", b.RegistroD110Count)
		addIf("D120", b.RegistroD120Count)
		addIf("D130", b.RegistroD130Count)
		addIf("D140", b.RegistroD140Count)
		addIf("D150", b.RegistroD150Count)
		addIf("D160", b.RegistroD160Count)
		addIf("D161", b.RegistroD161Count)
		addIf("D162", b.RegistroD162Count)
		addIf("D170", b.RegistroD170Count)
		addIf("D180", b.RegistroD180Count)
		addIf("D190", b.RegistroD190Count)
		addIf("D195", b.RegistroD195Count)
		addIf("D197", b.RegistroD197Count)
		addIf("D300", b.RegistroD300Count)
		addIf("D301", b.RegistroD301Count)
		addIf("D310", b.RegistroD310Count)
		addIf("D350", b.RegistroD350Count)
		addIf("D355", b.RegistroD355Count)
		addIf("D360", b.RegistroD360Count)
		addIf("D365", b.RegistroD365Count)
		addIf("D370", b.RegistroD370Count)
		addIf("D390", b.RegistroD390Count)
		addIf("D400", b.RegistroD400Count)
		addIf("D410", b.RegistroD410Count)
		addIf("D411", b.RegistroD411Count)
		addIf("D420", b.RegistroD420Count)
		addIf("D500", b.RegistroD500Count)
		addIf("D510", b.RegistroD510Count)
		addIf("D530", b.RegistroD530Count)
		addIf("D590", b.RegistroD590Count)
		addIf("D600", b.RegistroD600Count)
		addIf("D610", b.RegistroD610Count)
		addIf("D690", b.RegistroD690Count)
		addIf("D695", b.RegistroD695Count)
		addIf("D696", b.RegistroD696Count)
		addIf("D697", b.RegistroD697Count)
		addIf("D700", b.RegistroD700Count)
		addIf("D730", b.RegistroD730Count)
		addIf("D731", b.RegistroD731Count)
		addIf("D735", b.RegistroD735Count)
		addIf("D737", b.RegistroD737Count)
		addIf("D750", b.RegistroD750Count)
		addIf("D760", b.RegistroD760Count)
		addIf("D761", b.RegistroD761Count)
		add("D990", 1)
	}

	// Block E
	if b := f.BlocoE; b.RegistroE990.QtdLinE > 0 {
		add("E001", 1)
		addIf("E100", b.RegistroE100Count)
		addIf("E110", b.RegistroE110Count)
		addIf("E111", b.RegistroE111Count)
		addIf("E112", b.RegistroE112Count)
		addIf("E113", b.RegistroE113Count)
		addIf("E115", b.RegistroE115Count)
		addIf("E116", b.RegistroE116Count)
		addIf("E200", b.RegistroE200Count)
		addIf("E210", b.RegistroE210Count)
		addIf("E220", b.RegistroE220Count)
		addIf("E230", b.RegistroE230Count)
		addIf("E240", b.RegistroE240Count)
		addIf("E250", b.RegistroE250Count)
		addIf("E300", b.RegistroE300Count)
		addIf("E310", b.RegistroE310Count)
		addIf("E311", b.RegistroE311Count)
		addIf("E312", b.RegistroE312Count)
		addIf("E313", b.RegistroE313Count)
		addIf("E316", b.RegistroE316Count)
		addIf("E500", b.RegistroE500Count)
		addIf("E510", b.RegistroE510Count)
		addIf("E520", b.RegistroE520Count)
		addIf("E530", b.RegistroE530Count)
		addIf("E531", b.RegistroE531Count)
		add("E990", 1)
	}

	// Block G
	if b := f.BlocoG; b.RegistroG990.QtdLinG > 0 {
		add("G001", 1)
		addIf("G110", b.RegistroG110Count)
		addIf("G125", b.RegistroG125Count)
		addIf("G126", b.RegistroG126Count)
		addIf("G130", b.RegistroG130Count)
		addIf("G140", b.RegistroG140Count)
		add("G990", 1)
	}

	// Block H
	if b := f.BlocoH; b.RegistroH990.QtdLinH > 0 {
		add("H001", 1)
		addIf("H005", b.RegistroH005Count)
		addIf("H010", b.RegistroH010Count)
		addIf("H011", b.RegistroH011Count)
		addIf("H020", b.RegistroH020Count)
		addIf("H030", b.RegistroH030Count)
		add("H990", 1)
	}

	// Block K
	if b := f.BlocoK; b.RegistroK990.QtdLinK > 0 {
		add("K001", 1)
		addIf("K010", b.RegistroK010Count)
		addIf("K100", b.RegistroK100Count)
		addIf("K200", b.RegistroK200Count)
		addIf("K210", b.RegistroK210Count)
		addIf("K215", b.RegistroK215Count)
		addIf("K220", b.RegistroK220Count)
		addIf("K230", b.RegistroK230Count)
		addIf("K235", b.RegistroK235Count)
		addIf("K250", b.RegistroK250Count)
		addIf("K255", b.RegistroK255Count)
		addIf("K260", b.RegistroK260Count)
		addIf("K265", b.RegistroK265Count)
		addIf("K270", b.RegistroK270Count)
		addIf("K275", b.RegistroK275Count)
		addIf("K280", b.RegistroK280Count)
		addIf("K290", b.RegistroK290Count)
		addIf("K291", b.RegistroK291Count)
		addIf("K292", b.RegistroK292Count)
		addIf("K300", b.RegistroK300Count)
		addIf("K301", b.RegistroK301Count)
		addIf("K302", b.RegistroK302Count)
		add("K990", 1)
	}

	// Block 1
	if b := f.Bloco1; b.Registro1990.QtdLin1 > 0 {
		add("1001", 1)
		addIf("1010", b.Registro1010Count)
		addIf("1100", b.Registro1100Count)
		addIf("1105", b.Registro1105Count)
		addIf("1110", b.Registro1110Count)
		addIf("1200", b.Registro1200Count)
		addIf("1210", b.Registro1210Count)
		addIf("1250", b.Registro1250Count)
		addIf("1255", b.Registro1255Count)
		addIf("1300", b.Registro1300Count)
		addIf("1310", b.Registro1310Count)
		addIf("1320", b.Registro1320Count)
		addIf("1350", b.Registro1350Count)
		addIf("1360", b.Registro1360Count)
		addIf("1370", b.Registro1370Count)
		addIf("1390", b.Registro1390Count)
		addIf("1391", b.Registro1391Count)
		addIf("1400", b.Registro1400Count)
		addIf("1500", b.Registro1500Count)
		addIf("1510", b.Registro1510Count)
		addIf("1600", b.Registro1600Count)
		addIf("1601", b.Registro1601Count)
		addIf("1700", b.Registro1700Count)
		addIf("1710", b.Registro1710Count)
		addIf("1800", b.Registro1800Count)
		addIf("1900", b.Registro1900Count)
		addIf("1910", b.Registro1910Count)
		addIf("1920", b.Registro1920Count)
		addIf("1921", b.Registro1921Count)
		addIf("1922", b.Registro1922Count)
		addIf("1923", b.Registro1923Count)
		addIf("1925", b.Registro1925Count)
		addIf("1926", b.Registro1926Count)
		addIf("1960", b.Registro1960Count)
		addIf("1970", b.Registro1970Count)
		addIf("1975", b.Registro1975Count)
		addIf("1980", b.Registro1980Count)
		add("1990", 1)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// addRegistro9900 appends a Registro9900 entry to Bloco9, tracking how many
// lines a given register type contributed.
func (f *SPEDFiscal) addRegistro9900(reg string, qtd int) {
	r := &Registro9900{
		RegBlc:    reg,
		QtdRegBlc: qtd,
	}
	f.Bloco9.Registro9900 = append(f.Bloco9.Registro9900, r)
}

// lastDayOfMonth returns the last day number for the month of the given time.
func lastDayOfMonth(t time.Time) int {
	// Move to the 1st of the next month, then subtract one day.
	y, m, _ := t.Date()
	first := time.Date(y, m+1, 1, 0, 0, 0, 0, t.Location())
	last := first.AddDate(0, 0, -1)
	return last.Day()
}
