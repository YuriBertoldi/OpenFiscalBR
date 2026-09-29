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
// WriteRegistro implementations for blocks B, C, D, E, G, H, K, 1 and 9.
// ---------------------------------------------------------------------------

// ===========================================================================
// Bloco B - Escrituracao e Apuracao do ISS
// ===========================================================================

// WriteRegistroB001 gera a linha do registro B001.
// Formato: |B001|IND_MOV|
func (b *BlocoB) WriteRegistroB001() {
	if b.RegistroB001 == nil {
		return
	}
	linha := b.LFillStr("B001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroB001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroB990.QtdLinB++

	if b.RegistroB001.IndDad == 0 {
		b.writeRegistroB020()
		b.writeRegistroB350()
		b.writeRegistroB420()
		b.writeRegistroB440()
		b.writeRegistroB460()
		b.writeRegistroB470()
	}
}

func (b *BlocoB) writeRegistroB020() {
	for _, r := range b.RegistroB001.RegistroB020 {
		linha := b.LFillStr("B020", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvNFe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", false) +
			b.LFillStr(r.CodMunServ, 0, false, '0') +
			b.DFill(r.VlContIss, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.VlIssRT, 2, true) +
			b.DFill(r.VlDed, 2, true) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodInfObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB020Count++

		b.writeRegistroB025(r)
	}
}

func (b *BlocoB) writeRegistroB025(parent *RegistroB020) {
	for _, r := range parent.RegistroB025 {
		linha := b.LFillStr("B025", 0, false, '0') +
			b.DFill(r.VlContP, 2, false) +
			b.DFill(r.VlBcIssP, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIssP, 2, false) +
			b.DFill(r.VlIssPRet, 2, true) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB025Count++
	}
}

func (b *BlocoB) writeRegistroB350() {
	for _, r := range b.RegistroB001.RegistroB350 {
		linha := b.LFillStr("B350", 0, false, '0') +
			b.LFillStr(r.CodCtaISS, 0, false, '0') +
			b.LFillStr(r.CodInfObs, 0, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB350Count++
	}
}

func (b *BlocoB) writeRegistroB420() {
	for _, r := range b.RegistroB001.RegistroB420 {
		linha := b.LFillStr("B420", 0, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIssaPag, 2, false) +
			b.DFill(r.VlDed, 2, true) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB420Count++
	}
}

func (b *BlocoB) writeRegistroB440() {
	for _, r := range b.RegistroB001.RegistroB440 {
		linha := b.LFillStr("B440", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillStr(r.CodMunServ, 0, false, '0') +
			b.DFill(r.VlContRT, 2, false) +
			b.DFill(r.VlBcIssRT, 2, false) +
			b.DFill(r.VlIssRT, 2, false)
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB440Count++
	}
}

func (b *BlocoB) writeRegistroB460() {
	for _, r := range b.RegistroB001.RegistroB460 {
		linha := b.LFillStr("B460", 0, false, '0') +
			b.LFillStr(r.IndDed.String(), 0, false, '0') +
			b.DFill(r.VlDed, 2, false) +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB460Count++
	}
}

func (b *BlocoB) writeRegistroB470() {
	for _, r := range b.RegistroB001.RegistroB470 {
		linha := b.LFillStr("B470", 0, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlMatTerc, 2, true) +
			b.DFill(r.VlMatProp, 2, true) +
			b.DFill(r.VlSub, 2, true) +
			b.DFill(r.VlIsntIss, 2, true) +
			b.DFill(r.VlDedBC, 2, true) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.VlBcIssRT, 2, true) +
			b.DFill(r.VlIssRT, 2, true) +
			b.DFill(r.VlDedIss, 2, true) +
			b.DFill(r.VlIss, 2, false)
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB470Count++

		b.writeRegistroB500(r)
	}
}

func (b *BlocoB) writeRegistroB500(parent *RegistroB470) {
	for _, r := range parent.RegistroB500 {
		linha := b.LFillStr("B500", 0, false, '0') +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIssaPag, 2, false) +
			b.DFill(r.VlDed, 2, true) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.IndObrISS.String(), 0, false, '0') +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB500Count++
	}
}

// WriteRegistroB990 gera a linha do registro B990.
// Formato: |B990|QTD_LIN_B|
func (b *BlocoB) WriteRegistroB990() {
	if b.RegistroB990 == nil {
		return
	}
	b.RegistroB990.QtdLinB++
	linha := b.LFillStr("B990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroB990.QtdLinB), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco C - Documentos Fiscais I - Mercadorias (ICMS/IPI)
// ===========================================================================

// WriteRegistroC001 gera a linha do registro C001.
// Formato: |C001|IND_MOV|
func (b *BlocoC) WriteRegistroC001() {
	if b.RegistroC001 == nil {
		return
	}
	linha := b.LFillStr("C001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroC001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroC990.QtdLinC++

	if b.RegistroC001.IndDad == 0 {
		b.writeRegistroC100()
	}
}

func (b *BlocoC) writeRegistroC100() {
	for _, r := range b.RegistroC001.RegistroC100 {
		linha := b.LFillStr("C100", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvNFe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", false) +
			b.LFillDate(r.DtES, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.LFillStr(r.IndPgto.String(), 0, false, '0') +
			b.DFill(r.VlDesc, 2, true) +
			b.DFill(r.VlAbatNT, 2, true) +
			b.DFill(r.VlMerc, 2, false) +
			b.LFillStr(r.IndFrt.String(), 0, false, '0') +
			b.DFill(r.VlFrt, 2, true) +
			b.DFill(r.VlSeg, 2, true) +
			b.DFill(r.VlOutDa, 2, true) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, true) +
			b.DFill(r.VlICMSST, 2, true) +
			b.DFill(r.VlIPI, 2, true) +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.DFill(r.VlPISST, 2, true) +
			b.DFill(r.VlCOFINSST, 2, true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC100Count++

		b.writeRegistroC170(r)
		b.writeRegistroC190(r)
	}
}

func (b *BlocoC) writeRegistroC170(parent *RegistroC100) {
	for _, r := range parent.RegistroC170 {
		linha := b.LFillStr("C170", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.DescrCompl, 0, false, '0') +
			b.DFill(r.Qtd, 5, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, true) +
			b.LFillInt(int64(r.IndMov), 0, false, '0') +
			b.LFillStr(r.CstICMS.String(), 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.LFillStr(r.CodNat, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, true) +
			b.DFill(r.AliqICMS, 2, true) +
			b.DFill(r.VlICMS, 2, true) +
			b.DFill(r.VlBcICMSST, 2, true) +
			b.DFill(r.AliqST, 2, true) +
			b.DFill(r.VlICMSST, 2, true) +
			b.LFillStr(r.IndApur.String(), 0, false, '0') +
			b.LFillStr(r.CstIPI.String(), 0, false, '0') +
			b.LFillStr(r.CodEnq, 0, false, '0') +
			b.DFill(r.VlBcIPI, 2, true) +
			b.DFill(r.AliqIPI, 2, true) +
			b.DFill(r.VlIPI, 2, true) +
			b.LFillStr(r.CstPIS.String(), 0, false, '0') +
			b.DFill(r.VlBcPIS, 2, true) +
			b.DFill(r.AliqPIS, 2, true) +
			b.DFill(r.QtdBcPIS, 4, true) +
			b.DFill(r.AliqPISReais, 4, true) +
			b.DFill(r.VlPIS, 2, true) +
			b.LFillStr(r.CstCOFINS.String(), 0, false, '0') +
			b.DFill(r.VlBcCOFINS, 2, true) +
			b.DFill(r.AliqCOFINS, 2, true) +
			b.DFill(r.QtdBcCOFINS, 4, true) +
			b.DFill(r.AliqCOFINSReais, 4, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.LFillStr(r.CodCta, 0, false, '0') +
			b.DFill(r.VlAbatNT, 2, true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC170Count++
	}
}

func (b *BlocoC) writeRegistroC190(parent *RegistroC100) {
	for _, r := range parent.RegistroC190 {
		linha := b.LFillStr("C190", 0, false, '0') +
			b.LFillStr(r.CstICMS.String(), 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, true) +
			b.DFill(r.VlICMSST, 2, true) +
			b.DFill(r.VlRedBC, 2, true) +
			b.DFill(r.VlIPI, 2, true) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC190Count++
	}
}

// WriteRegistroC990 gera a linha do registro C990.
// Formato: |C990|QTD_LIN_C|
func (b *BlocoC) WriteRegistroC990() {
	if b.RegistroC990 == nil {
		return
	}
	b.RegistroC990.QtdLinC++
	linha := b.LFillStr("C990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroC990.QtdLinC), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco D - Documentos Fiscais II - Servicos (ICMS)
// ===========================================================================

// WriteRegistroD001 gera a linha do registro D001.
// Formato: |D001|IND_MOV|
func (b *BlocoD) WriteRegistroD001() {
	if b.RegistroD001 == nil {
		return
	}
	linha := b.LFillStr("D001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroD001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroD990.QtdLinD++

	if b.RegistroD001.IndDad == 0 {
		b.writeRegistroD100()
	}
}

func (b *BlocoD) writeRegistroD100() {
	for _, r := range b.RegistroD001.RegistroD100 {
		linha := b.LFillStr("D100", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvCTe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", false) +
			b.LFillDate(r.DtAP, "02012006", true) +
			b.LFillStr(r.TpCTe, 0, false, '0') +
			b.LFillStr(r.ChvCTeRef, 0, false, '0') +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, true) +
			b.LFillStr(r.IndFrt.String(), 0, false, '0') +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlNT, 2, true) +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD100Count++

		b.writeRegistroD190(r)
	}
}

func (b *BlocoD) writeRegistroD190(parent *RegistroD100) {
	for _, r := range parent.RegistroD190 {
		linha := b.LFillStr("D190", 0, false, '0') +
			b.LFillStr(r.CstICMS.String(), 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, true) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD190Count++
	}
}

// WriteRegistroD990 gera a linha do registro D990.
// Formato: |D990|QTD_LIN_D|
func (b *BlocoD) WriteRegistroD990() {
	if b.RegistroD990 == nil {
		return
	}
	b.RegistroD990.QtdLinD++
	linha := b.LFillStr("D990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroD990.QtdLinD), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco E - Apuracao do ICMS e do IPI
// ===========================================================================

// WriteRegistroE001 gera a linha do registro E001.
// Formato: |E001|IND_MOV|
func (b *BlocoE) WriteRegistroE001() {
	if b.RegistroE001 == nil {
		return
	}
	linha := b.LFillStr("E001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroE001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroE990.QtdLinE++

	if b.RegistroE001.IndDad == 0 {
		b.writeRegistroE100()
	}
}

func (b *BlocoE) writeRegistroE100() {
	for _, r := range b.RegistroE001.RegistroE100 {
		linha := b.LFillStr("E100", 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", false) +
			b.LFillDate(r.DtFin, "02012006", false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE100Count++

		b.writeRegistroE110(r)
	}
}

func (b *BlocoE) writeRegistroE110(parent *RegistroE100) {
	for _, r := range parent.RegistroE110 {
		if r == nil {
			continue
		}
		linha := b.LFillStr("E110", 0, false, '0') +
			b.DFill(r.VlTotDebitos, 2, false) +
			b.DFill(r.VlAjDebitos, 2, false) +
			b.DFill(r.VlTotAjDebitos, 2, false) +
			b.DFill(r.VlEstornosCred, 2, false) +
			b.DFill(r.VlTotCreditos, 2, false) +
			b.DFill(r.VlAjCreditos, 2, false) +
			b.DFill(r.VlTotAjCreditos, 2, false) +
			b.DFill(r.VlEstornosDeb, 2, false) +
			b.DFill(r.VlSldCredorAnt, 2, false) +
			b.DFill(r.VlSldApurado, 2, false) +
			b.DFill(r.VlTotDed, 2, false) +
			b.DFill(r.VlICMSRecolher, 2, false) +
			b.DFill(r.VlSldCredorTransp, 2, false) +
			b.DFill(r.DebEsp, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE110Count++
	}
}

// WriteRegistroE990 gera a linha do registro E990.
// Formato: |E990|QTD_LIN_E|
func (b *BlocoE) WriteRegistroE990() {
	if b.RegistroE990 == nil {
		return
	}
	b.RegistroE990.QtdLinE++
	linha := b.LFillStr("E990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroE990.QtdLinE), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco G - Controle do Credito de ICMS do Ativo Permanente (CIAP)
// ===========================================================================

// WriteRegistroG001 gera a linha do registro G001.
// Formato: |G001|IND_MOV|
func (b *BlocoG) WriteRegistroG001() {
	if b.RegistroG001 == nil {
		return
	}
	linha := b.LFillStr("G001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroG001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroG990.QtdLinG++
}

// WriteRegistroG990 gera a linha do registro G990.
// Formato: |G990|QTD_LIN_G|
func (b *BlocoG) WriteRegistroG990() {
	if b.RegistroG990 == nil {
		return
	}
	b.RegistroG990.QtdLinG++
	linha := b.LFillStr("G990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroG990.QtdLinG), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco H - Inventario Fisico
// ===========================================================================

// WriteRegistroH001 gera a linha do registro H001.
// Formato: |H001|IND_MOV|
func (b *BlocoH) WriteRegistroH001() {
	if b.RegistroH001 == nil {
		return
	}
	linha := b.LFillStr("H001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroH001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroH990.QtdLinH++

	if b.RegistroH001.IndDad == 0 {
		b.writeRegistroH005()
	}
}

func (b *BlocoH) writeRegistroH005() {
	for _, r := range b.RegistroH001.RegistroH005 {
		linha := b.LFillStr("H005", 0, false, '0') +
			b.LFillDate(r.DtInv, "02012006", false) +
			b.DFill(r.VlInv, 2, false) +
			b.LFillStr(r.MotInv.String(), 0, false, '0')
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH005Count++

		b.writeRegistroH010(r)
	}
}

func (b *BlocoH) writeRegistroH010(parent *RegistroH005) {
	for _, r := range parent.RegistroH010 {
		linha := b.LFillStr("H010", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.DFill(r.VlUnit, 6, false) +
			b.DFill(r.VlItem, 2, false) +
			b.LFillStr(r.IndProp.String(), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0') +
			b.DFill(r.VlItemIR, 2, true)
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH010Count++
	}
}

// WriteRegistroH990 gera a linha do registro H990.
// Formato: |H990|QTD_LIN_H|
func (b *BlocoH) WriteRegistroH990() {
	if b.RegistroH990 == nil {
		return
	}
	b.RegistroH990.QtdLinH++
	linha := b.LFillStr("H990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroH990.QtdLinH), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco K - Controle da Producao e do Estoque
// ===========================================================================

// WriteRegistroK001 gera a linha do registro K001.
// Formato: |K001|IND_MOV|
func (b *BlocoK) WriteRegistroK001() {
	if b.RegistroK001 == nil {
		return
	}
	linha := b.LFillStr("K001", 0, false, '0') +
		b.LFillInt(int64(b.RegistroK001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.RegistroK990.QtdLinK++
}

// WriteRegistroK990 gera a linha do registro K990.
// Formato: |K990|QTD_LIN_K|
func (b *BlocoK) WriteRegistroK990() {
	if b.RegistroK990 == nil {
		return
	}
	b.RegistroK990.QtdLinK++
	linha := b.LFillStr("K990", 0, false, '0') +
		b.LFillInt(int64(b.RegistroK990.QtdLinK), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco 1 - Outras Informacoes
// ===========================================================================

// WriteRegistro1001 gera a linha do registro 1001.
// Formato: |1001|IND_MOV|
func (b *Bloco1) WriteRegistro1001() {
	if b.Registro1001 == nil {
		return
	}
	linha := b.LFillStr("1001", 0, false, '0') +
		b.LFillInt(int64(b.Registro1001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.Registro1990.QtdLin1++
}

// WriteRegistro1990 gera a linha do registro 1990.
// Formato: |1990|QTD_LIN_1|
func (b *Bloco1) WriteRegistro1990() {
	if b.Registro1990 == nil {
		return
	}
	b.Registro1990.QtdLin1++
	linha := b.LFillStr("1990", 0, false, '0') +
		b.LFillInt(int64(b.Registro1990.QtdLin1), 0, false, '0')
	b.Add(linha, true)
}

// ===========================================================================
// Bloco 9 - Controle e Encerramento do Arquivo Digital
// ===========================================================================

// WriteRegistro9001 gera a linha do registro 9001.
// Formato: |9001|IND_MOV|
func (b *Bloco9) WriteRegistro9001() {
	if b.Registro9001 == nil {
		return
	}
	linha := b.LFillStr("9001", 0, false, '0') +
		b.LFillInt(int64(b.Registro9001.IndDad), 0, false, '0')
	b.Add(linha, true)
	b.Registro9990.QtdLin9++
}

// WriteRegistro9900 gera as linhas do registro 9900.
// Formato: |9900|REG_BLC|QTD_REG_BLC|
func (b *Bloco9) WriteRegistro9900() {
	for _, r := range b.Registro9900 {
		linha := b.LFillStr("9900", 0, false, '0') +
			b.LFillStr(r.RegBlc, 0, false, '0') +
			b.LFillInt(int64(r.QtdRegBlc), 0, false, '0')
		b.Add(linha, true)
		b.Registro9990.QtdLin9++
	}
}

// WriteRegistro9990 gera a linha do registro 9990.
// Formato: |9990|QTD_LIN_9|
func (b *Bloco9) WriteRegistro9990() {
	if b.Registro9990 == nil {
		return
	}
	b.Registro9990.QtdLin9++
	linha := b.LFillStr("9990", 0, false, '0') +
		b.LFillInt(int64(b.Registro9990.QtdLin9), 0, false, '0')
	b.Add(linha, true)
}

// WriteRegistro9999 gera a linha do registro 9999.
// Formato: |9999|QTD_LIN|
func (b *Bloco9) WriteRegistro9999() {
	if b.Registro9999 == nil {
		return
	}
	linha := b.LFillStr("9999", 0, false, '0') +
		b.LFillInt(int64(b.Registro9999.QtdLin), 0, false, '0')
	b.Add(linha, true)
}
