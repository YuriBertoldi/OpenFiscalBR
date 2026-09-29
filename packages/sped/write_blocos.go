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
	"strings"
	"time"
)

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
		b.writeRegistroB030()
		b.writeRegistroB350()
		b.writeRegistroB420()
		b.writeRegistroB440()
		b.writeRegistroB460()
		b.writeRegistroB470()
		b.writeRegistroB500()
	}
}

// writeRegistroB020 gera as linhas do registro B020.
// Formato: |B020|IND_OPER|IND_EMIT|COD_PART|COD_MOD|COD_SIT|SER|NUM_DOC|CHV_NFE|DT_DOC|
//
//	COD_MUN_SERV|VL_CONT|VL_MAT_TERC|VL_SUB|VL_ISNT_ISS|VL_DED_BC|VL_BC_ISS|
//	VL_BC_ISS_RT|VL_ISS_RT|VL_ISS|COD_INF_OBS|
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
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlMatTerc, 2, false) +
			b.DFill(r.VlSub, 2, false) +
			b.DFill(r.VlIsntIss, 2, false) +
			b.DFill(r.VlDedBC, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.VlBcIssRT, 2, false) +
			b.DFill(r.VlIssRT, 2, false) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodInfObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB020Count++

		b.writeRegistroB025(r)
	}
}

// writeRegistroB025 gera as linhas do registro B025.
// Formato: |B025|VL_CONT_P|VL_BC_ISS_P|ALIQ_ISS|VL_ISS_P|VL_ISNT_ISS_P|COD_SERV|
func (b *BlocoB) writeRegistroB025(parent *RegistroB020) {
	for _, r := range parent.RegistroB025 {
		linha := b.LFillStr("B025", 0, false, '0') +
			b.DFill(r.VlContP, 2, false) +
			b.DFill(r.VlBcIssP, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIssP, 2, false) +
			b.DFill(r.VlIsntIssP, 2, false) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB025Count++
	}
}

// writeRegistroB030 gera as linhas do registro B030.
// Formato: |B030|COD_MOD|SER|NUM_DOC_INI|NUM_DOC_FIN|DT_DOC|QTD_CANC|VL_CONT|
//
//	VL_ISNT_ISS|VL_BC_ISS|VL_ISS|COD_INF_OBS|
func (b *BlocoB) writeRegistroB030() {
	for _, r := range b.RegistroB001.RegistroB030 {
		linha := b.LFillStr("B030", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDocIni, 0, false, '0') +
			b.LFillStr(r.NumDocFin, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", false) +
			b.LFillInt(int64(r.QtdCanc), 0, false, '0') +
			b.DFill(r.VlCont, 2, true) +
			b.DFill(r.VlIsntIss, 2, true) +
			b.DFill(r.VlBcIss, 2, true) +
			b.DFill(r.VlIss, 2, true) +
			b.LFillStr(r.CodInfObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB030Count++

		b.writeRegistroB035(r)
	}
}

// writeRegistroB035 gera as linhas do registro B035.
// Formato: |B035|VL_CONT_P|VL_BC_ISS_P|ALIQ_ISS|VL_ISS_P|VL_ISNT_ISS_P|COD_SERV|
func (b *BlocoB) writeRegistroB035(parent *RegistroB030) {
	for _, r := range parent.RegistroB035 {
		linha := b.LFillStr("B035", 0, false, '0') +
			b.DFill(r.VlContP, 2, false) +
			b.DFill(r.VlBcIssP, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIssP, 2, false) +
			b.DFill(r.VlIsntIssP, 2, false) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB035Count++
	}
}

// writeRegistroB350 gera as linhas do registro B350.
// Formato: |B035|COD_CTD|CTA_ISS|CTA_COSIF|QTD_OCOR|COD_SERV|VL_CONT|VL_BC_ISS|
//
//	ALIQ_ISS|VL_ISS|COD_INF_OBS|
//
// O literal emitido e "B035", nao "B350": e o que o ACBr faz
// (ACBrEFDBloco_B_Class.pas, WriteRegistroB350), mantido por fidelidade ao
// original. Aparentemente e um erro de copia no Delphi, ja que o registro B035
// tem outro conjunto de campos.
func (b *BlocoB) writeRegistroB350() {
	for _, r := range b.RegistroB001.RegistroB350 {
		linha := b.LFillStr("B035", 0, false, '0') +
			b.LFillStr(r.CodCtd, 0, false, '0') +
			b.LFillStr(r.CtaIss, 0, false, '0') +
			b.LFillStr(r.CtaCosif, 8, false, '0') +
			b.LFillInt(int64(r.QtdOcor), 0, false, '0') +
			b.LFillStr(r.CodServ, 4, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodInfObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB350Count++
	}
}

// writeRegistroB420 gera as linhas do registro B420.
// Formato: |B420|VL_CONT|VL_BC_ISS|ALIQ_ISS|VL_ISNT_ISS|VL_ISS|COD_SERV|
func (b *BlocoB) writeRegistroB420() {
	for _, r := range b.RegistroB001.RegistroB420 {
		linha := b.LFillStr("B420", 0, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.AliqIss, 2, false) +
			b.DFill(r.VlIsntIss, 2, false) +
			b.DFill(r.VlIss, 2, false) +
			b.LFillStr(r.CodServ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB420Count++
	}
}

// writeRegistroB440 gera as linhas do registro B440.
// Formato: |B440|IND_OPER|COD_PART|VL_CONT_RT|VL_BC_ISS_RT|VL_ISS_RT|
func (b *BlocoB) writeRegistroB440() {
	for _, r := range b.RegistroB001.RegistroB440 {
		linha := b.LFillStr("B440", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.DFill(r.VlContRT, 2, false) +
			b.DFill(r.VlBcIssRT, 2, false) +
			b.DFill(r.VlIssRT, 2, false)
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB440Count++
	}
}

// writeRegistroB460 gera as linhas do registro B460.
// Formato: |B460|IND_DED|VL_DED|NUM_PROC|IND_PROC|PROC|COD_INF_OBS|IND_OBR|
func (b *BlocoB) writeRegistroB460() {
	for _, r := range b.RegistroB001.RegistroB460 {
		linha := b.LFillStr("B460", 0, false, '0') +
			b.LFillStr(r.IndDed.String(), 0, false, '0') +
			b.DFill(r.VlDed, 2, false) +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.CodInfObs, 0, false, '0') +
			b.LFillStr(r.IndObr.String(), 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB460Count++
	}
}

// writeRegistroB470 gera as linhas do registro B470.
// Formato: |B470|VL_CONT|VL_MAT_TERC|VL_MAT_PROP|VL_SUB|VL_ISNT|VL_DED_BC|
//
//	VL_BC_ISS|VL_BC_ISS_RT|VL_ISS|VL_ISS_RT|VL_DED|VL_ISS_REC|VL_ISS_ST|
//	VL_ISS_REC_UNI|
func (b *BlocoB) writeRegistroB470() {
	for _, r := range b.RegistroB001.RegistroB470 {
		linha := b.LFillStr("B470", 0, false, '0') +
			b.DFill(r.VlCont, 2, false) +
			b.DFill(r.VlMatTerc, 2, false) +
			b.DFill(r.VlMatProp, 2, false) +
			b.DFill(r.VlSub, 2, false) +
			b.DFill(r.VlIsnt, 2, false) +
			b.DFill(r.VlDedBC, 2, false) +
			b.DFill(r.VlBcIss, 2, false) +
			b.DFill(r.VlBcIssRT, 2, false) +
			b.DFill(r.VlIss, 2, false) +
			b.DFill(r.VlIssRT, 2, false) +
			b.DFill(r.VlDed, 2, false) +
			b.DFill(r.VlIssRec, 2, false) +
			b.DFill(r.VlIssST, 2, false) +
			b.DFill(r.VlIssRecUni, 2, false)
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB470Count++
	}
}

// writeRegistroB500 gera as linhas do registro B500.
// Formato: |B500|VL_REC|QTD_PROF|VL_OR|
func (b *BlocoB) writeRegistroB500() {
	for _, r := range b.RegistroB001.RegistroB500 {
		linha := b.LFillStr("B500", 0, false, '0') +
			b.DFill(r.VlRec, 2, false) +
			b.LFillInt(int64(r.QtdProf), 0, false, '0') +
			b.DFill(r.VlOR, 2, false)
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB500Count++

		b.writeRegistroB510(r)
	}
}

// writeRegistroB510 gera as linhas do registro B510.
// Formato: |B510|IND_PROF|IND_ESC|IND_SOC|CPF|NOME|
func (b *BlocoB) writeRegistroB510(parent *RegistroB500) {
	for _, r := range parent.RegistroB510 {
		linha := b.LFillStr("B510", 0, false, '0') +
			b.LFillStr(r.IndProf, 0, false, '0') +
			b.LFillStr(r.IndEsc, 0, false, '0') +
			b.LFillStr(r.IndSoc, 0, false, '0') +
			b.LFillStr(r.CPF, 0, false, '0') +
			b.LFillStr(r.Nome, 0, false, '0')
		b.Add(linha, true)
		b.RegistroB990.QtdLinB++
		b.RegistroB510Count++
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

// writeRegistroC100 gera as linhas do registro C100.
// Formato: |C100|IND_OPER|IND_EMIT|COD_PART|COD_MOD|COD_SIT|SER|NUM_DOC|CHV_NFE|DT_DOC|DT_E_S|
//
//	VL_DOC|IND_PGTO|VL_DESC|VL_ABAT_NT|VL_MERC|IND_FRT|VL_FRT|VL_SEG|VL_OUT_DA|
//	VL_BC_ICMS|VL_ICMS|VL_BC_ICMS_ST|VL_ICMS_ST|VL_IPI|VL_PIS|VL_COFINS|VL_PIS_ST|VL_COFINS_ST|
//
// Ref.: ACBrEFDBloco_C_Class.pas, TBloco_C.WriteRegistroC100.
func (b *BlocoC) writeRegistroC100() {
	for _, r := range b.RegistroC001.RegistroC100 {
		// O consumidor pode vetar o registro pelo callback. Ao contrario do
		// ACBr, a flag e reavaliada a cada item -- la ela persiste entre
		// iteracoes e um veto derruba todos os documentos seguintes.
		if b.OnCheckRegistroC100 != nil {
			abortar := false
			b.OnCheckRegistroC100(r, &abortar)
			if abortar {
				continue
			}
		}

		// NF-e (modelo 55) exige chave de acesso, salvo numeracao inutilizada.
		b.Checkf(!(r.CodMod == "55" && r.CodSit != SitNumInutilizada && strings.TrimSpace(r.ChvNFe) == ""),
			"(C-C100) Nota: %s / Serie: %s / Emitida em: %s / Modelo: %s / chave de acesso ausente",
			r.NumDoc, r.Ser, r.DtDoc.Format("02/01/2006"), r.CodMod)

		// Documento cancelado (02/03), denegado (04) ou inutilizado (05): datas
		// e indicadores sao zerados e os valores saem vazios em vez de 0,00.
		cancelada := r.CodSit == SitCancelado || r.CodSit == SitCanceladoExtemp ||
			r.CodSit == SitDenegado || r.CodSit == SitNumInutilizada

		dtDoc, dtES := r.DtDoc, r.DtES
		indFrt, indPgto := r.IndFrt, r.IndPgto
		if cancelada {
			dtDoc, dtES = time.Time{}, time.Time{}
			indFrt, indPgto = FrtNenhum, PgtoNenhum
		}

		// NFC-e (modelo 65) nao informa ST nem IPI, mesmo que preenchidos.
		nulosST := cancelada || r.CodMod == "65"

		linha := b.LFillStr("C100", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvNFe, 0, false, '0') +
			b.LFillDate(dtDoc, "02012006", true) +
			b.LFillDate(dtES, "02012006", true) +
			b.DFill(r.VlDoc, 2, cancelada) +
			b.LFillStr(indPgto.StringEm(b.DtIni), 0, false, '0') +
			b.DFill(r.VlDesc, 2, cancelada) +
			b.DFill(r.VlAbatNT, 2, cancelada) +
			b.DFill(r.VlMerc, 2, cancelada) +
			b.LFillStr(indFrt.StringEm(b.DtIni), 0, false, '0') +
			b.DFill(r.VlFrt, 2, cancelada) +
			b.DFill(r.VlSeg, 2, cancelada) +
			b.DFill(r.VlOutDa, 2, cancelada) +
			b.DFill(r.VlBcICMS, 2, cancelada) +
			b.DFill(r.VlICMS, 2, cancelada) +
			b.DFill(r.VlBcICMSST, 2, nulosST) +
			b.DFill(r.VlICMSST, 2, nulosST) +
			b.DFill(r.VlIPI, 2, nulosST) +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.DFill(r.VlPISST, 2, true) +
			b.DFill(r.VlCOFINSST, 2, true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC100Count++

		b.writeRegistroC101(r)
		b.writeRegistroC110(r)
		b.writeRegistroC170(r)
		b.writeRegistroC190(r)
	}
}

// writeRegistroC101 gera as linhas do registro C101 (partilha do ICMS
// interestadual a consumidor final, EC 87/2015).
// Formato: |C101|VL_FCP_UF_DEST|VL_ICMS_UF_DEST|VL_ICMS_UF_REM|
func (b *BlocoC) writeRegistroC101(parent *RegistroC100) {
	for _, r := range parent.RegistroC101 {
		linha := b.LFillStr("C101", 0, false, '0') +
			b.DFill(r.VlFcpUFDest, 2, false) +
			b.DFill(r.VlICMSUFDest, 2, false) +
			b.DFill(r.VlICMSUFRem, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC101Count++
	}
}

// writeRegistroC110 gera as linhas do registro C110 (informacao complementar
// do documento fiscal).
// Formato: |C110|COD_INF|TXT_COMPL|
func (b *BlocoC) writeRegistroC110(parent *RegistroC100) {
	for _, r := range parent.RegistroC110 {
		linha := b.LFillStr("C110", 0, false, '0') +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC110Count++

		b.writeRegistroC111(r)
		b.writeRegistroC112(r)
		b.writeRegistroC113(r)
		b.writeRegistroC114(r)
	}
}

// writeRegistroC111 gera as linhas do registro C111 (processo referenciado).
// Formato: |C111|NUM_PROC|IND_PROC|
func (b *BlocoC) writeRegistroC111(parent *RegistroC110) {
	for _, r := range parent.RegistroC111 {
		linha := b.LFillStr("C111", 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC111Count++
	}
}

// writeRegistroC112 gera as linhas do registro C112 (documento de arrecadacao
// referenciado).
// Formato: |C112|COD_DA|UF|NUM_DA|COD_AUT|VL_DA|DT_VCTO|DT_PGTO|
func (b *BlocoC) writeRegistroC112(parent *RegistroC110) {
	for _, r := range parent.RegistroC112 {
		linha := b.LFillStr("C112", 0, false, '0') +
			b.LFillInt(int64(r.CodDa), 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0') +
			b.LFillStr(r.NumDa, 0, false, '0') +
			b.LFillStr(r.CodAut, 0, false, '0') +
			b.DFill(r.VlDa, 2, false) +
			b.LFillDate(r.DtVcto, "02012006", true) +
			b.LFillDate(r.DtPgto, "02012006", true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC112Count++
	}
}

// writeRegistroC113 gera as linhas do registro C113 (documento fiscal
// referenciado).
// Formato: |C113|IND_OPER|IND_EMIT|COD_PART|COD_MOD|SER|SUB|NUM_DOC|DT_DOC|CHV_DOCe|
//
// CHV_DOCe so existe no leiaute a partir da versao 110.
func (b *BlocoC) writeRegistroC113(parent *RegistroC110) {
	for _, r := range parent.RegistroC113 {
		linha := b.LFillStr("C113", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			chvDocE113(b, r)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC113Count++
	}
}

// chvDocE113 devolve o campo CHV_DOCe do C113, que so existe a partir da
// versao 110 do leiaute.
// Ref.: ACBrEFDBloco_C_Class.pas, WriteRegistroC113.
func chvDocE113(b *BlocoC, r *RegistroC113) string {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil ||
		b.Bloco0.Registro0000.CodVer < VlVersao110 {
		return ""
	}
	return b.LFillStr(r.ChvDocE, 0, false, '0')
}

// writeRegistroC114 gera as linhas do registro C114 (cupom fiscal
// referenciado).
// Formato: |C114|COD_MOD|ECF_FAB|ECF_CX|NUM_DOC|DT_DOC|
func (b *BlocoC) writeRegistroC114(parent *RegistroC110) {
	for _, r := range parent.RegistroC114 {
		linha := b.LFillStr("C114", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.EcfFab, 0, false, '0') +
			b.LFillStr(r.EcfCx, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC114Count++
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
			vlAbatNT170(b, r)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC170Count++
	}
}

// vlAbatNT170 devolve o campo VL_ABAT_NT do C170, que so entrou no leiaute a
// partir da versao 112. Antes disso o campo nao e emitido -- nem o delimitador.
// Ref.: ACBrEFDBloco_C_Class.pas, WriteRegistroC170.
func vlAbatNT170(b *BlocoC, r *RegistroC170) string {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil ||
		b.Bloco0.Registro0000.CodVer <= VlVersao111 {
		return ""
	}
	return b.DFill(r.VlAbatNT, 2, true)
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
			b.LFillStr(r.CodCta, 0, false, '0') +
			municipiosD100(b, r)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD100Count++

		b.writeRegistroD190(r)
	}
}

// municipiosD100 devolve os campos COD_MUN_ORIG e COD_MUN_DEST do D100, que
// entraram no leiaute em 01/2018. Antes disso o registro termina em COD_CTA.
// Ref.: ACBrEFDBloco_D_Class.pas, WriteRegistroD100.
func municipiosD100(b *BlocoD, r *RegistroD100) string {
	inicio := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	if b.DtIni.Before(inicio) {
		return ""
	}
	return b.LFillStr(r.CodMunOrig, 0, false, '0') +
		b.LFillStr(r.CodMunDest, 0, false, '0')
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

		b.writeRegistroE111(r)
		b.writeRegistroE116(r)
	}
}

// writeRegistroE111 gera as linhas do registro E111 (ajuste/beneficio/incentivo
// da apuracao do ICMS).
// Formato: |E111|COD_AJ_APUR|DESCR_COMPL_AJ|VL_AJ_APUR|
func (b *BlocoE) writeRegistroE111(parent *RegistroE110) {
	for _, r := range parent.RegistroE111 {
		linha := b.LFillStr("E111", 0, false, '0') +
			b.LFillStr(r.CodAjApur, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.DFill(r.VlAjApur, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE111Count++
	}
}

// writeRegistroE116 gera as linhas do registro E116 (obrigacoes do ICMS a
// recolher).
// Formato: |E116|COD_OR|VL_OR|DT_VCTO|COD_REC|NUM_PROC|IND_PROC|PROC|TXT_COMPL|MES_REF|
//
// O registro nao existe na versao 101 do leiaute; MES_REF so entra a partir da
// versao 103. Ref.: ACBrEFDBloco_E_Class.pas, WriteRegistroE116.
func (b *BlocoE) writeRegistroE116(parent *RegistroE110) {
	ver := VlVersao100
	if b.Bloco0 != nil && b.Bloco0.Registro0000 != nil {
		ver = b.Bloco0.Registro0000.CodVer
	}
	if ver == VlVersao101 {
		return
	}

	for _, r := range parent.RegistroE116 {
		linha := b.LFillStr("E116", 0, false, '0') +
			b.LFillStr(r.CodOR, 0, false, '0') +
			b.DFill(r.VlOR, 2, false) +
			b.LFillDate(r.DtVcto, "02012006", true) +
			b.LFillStr(r.CodRec, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		if ver >= VlVersao103 {
			linha += b.LFillStr(r.MesRef, 0, false, '0')
		}
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE116Count++
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

	if b.RegistroG001.IndDad == 0 {
		b.writeRegistroG110()
	}
}

// versaoG devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoG) versaoG() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// writeRegistroG110 gera as linhas do registro G110 (ICMS do periodo, CIAP).
// Formato ate a versao 102: |G110|DT_INI|DT_FIN|MODO_CIAP|SALDO_IN_ICMS|SALDO_FN_ICMS|
//
//	SOM_PARC|VL_TRIB_EXP|VL_TOTAL|IND_PER_SAI|ICMS_APROP|SOM_ICMS_OC|
//
// A partir da versao 103, MODO_CIAP e SALDO_FN_ICMS deixam de existir.
// Ref.: ACBrEFDBloco_G_Class.pas, WriteRegistroG110.
func (b *BlocoG) writeRegistroG110() {
	for _, r := range b.RegistroG001.RegistroG110 {
		linha := b.LFillStr("G110", 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		if b.versaoG() == VlVersao102 {
			linha += b.LFillStr(r.ModoCiap, 0, false, '0') +
				b.DFill(r.SaldoInICMS, 2, false) +
				b.DFill(r.SaldoFnICMS, 2, false)
		} else {
			linha += b.DFill(r.SaldoInICMS, 2, false)
		}
		linha += b.DFill(r.SomParc, 2, false) +
			b.DFill(r.VlTribExp, 2, false) +
			b.DFill(r.VlTotal, 2, false) +
			b.DFill(r.IndPerSai, 8, false) +
			b.DFill(r.ICMSAprop, 2, false) +
			b.DFill(r.SomICMSOC, 2, false)
		b.Add(linha, true)
		b.RegistroG990.QtdLinG++
		b.RegistroG110Count++

		b.writeRegistroG125(r)
	}
}

// writeRegistroG125 gera as linhas do registro G125 (movimentacao do bem).
// Formato: |G125|COD_IND_BEM|DT_MOV|TIPO_MOV|VL_IMOB_ICMS_OP|VL_IMOB_ICMS_ST|
//
//	VL_IMOB_ICMS_FRT|VL_IMOB_ICMS_DIF|NUM_PARC|VL_PARC_PASS|VL_PARC_APROP|
//
// VL_PARC_APROP so e emitido na versao 102; a partir da 103 o campo sai do
// layout. Ref.: ACBrEFDBloco_G_Class.pas, WriteRegistroG125.
func (b *BlocoG) writeRegistroG125(parent *RegistroG110) {
	for _, r := range parent.RegistroG125 {
		linha := b.LFillStr("G125", 0, false, '0') +
			b.LFillStr(r.CodIndBem, 0, false, '0') +
			b.LFillDate(r.DtMov, "02012006", true) +
			b.LFillStr(r.TipoMov.String(), 0, false, '0') +
			b.DFill(r.VlImobICMSOp, 2, false) +
			b.DFill(r.VlImobICMSST, 2, false) +
			b.DFill(r.VlImobICMSFrt, 2, false) +
			b.DFill(r.VlImobICMSDif, 2, false) +
			b.LFillStr(r.NumParc, 3, false, '0') +
			b.DFill(r.VlParcPass, 2, false)
		if b.versaoG() == VlVersao102 {
			linha += b.DFill(r.VlParcApr, 2, false)
		}
		b.Add(linha, true)
		b.RegistroG990.QtdLinG++
		b.RegistroG125Count++

		b.writeRegistroG126(r)
		b.writeRegistroG130(r)
	}
}

// writeRegistroG126 gera as linhas do registro G126 (outros creditos CIAP).
// Formato: |G126|DT_INI|DT_FIN|NUM_PARC|VL_PARC_PASS|VL_TRIB_OC|VL_TOTAL|
//
//	IND_PER_SAI|VL_PARC_APROP|
func (b *BlocoG) writeRegistroG126(parent *RegistroG125) {
	for _, r := range parent.RegistroG126 {
		linha := b.LFillStr("G126", 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true) +
			b.LFillStr(r.NumParc, 3, false, '0') +
			b.DFill(r.VlParcPass, 2, false) +
			b.DFill(r.VlTribOC, 2, false) +
			b.DFill(r.VlTotal, 2, false) +
			b.DFill(r.IndPerSai, 8, false) +
			b.DFill(r.VlParcApr, 2, false)
		b.Add(linha, true)
		b.RegistroG990.QtdLinG++
		b.RegistroG126Count++
	}
}

// writeRegistroG130 gera as linhas do registro G130 (documento fiscal do bem).
// Formato: |G130|IND_EMIT|COD_PART|COD_MOD|SERIE|NUM_DOC|CHV_NFE_CTE|DT_DOC|NUM_DA|
//
// NUM_DA so existe a partir da versao 113.
// Ref.: ACBrEFDBloco_G_Class.pas, WriteRegistroG130.
func (b *BlocoG) writeRegistroG130(parent *RegistroG125) {
	for _, r := range parent.RegistroG130 {
		linha := b.LFillStr("G130", 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Serie, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvNFeCTe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true)
		if b.versaoG() >= VlVersao113 {
			linha += b.LFillStr(r.NumDA, 0, false, '0')
		}
		b.Add(linha, true)
		b.RegistroG990.QtdLinG++
		b.RegistroG130Count++

		b.writeRegistroG140(r)
	}
}

// writeRegistroG140 gera as linhas do registro G140 (item do documento fiscal).
// Formato: |G140|NUM_ITEM|COD_ITEM|QTDE|UNID|VL_ICMS_OP_APLICADO|
//
//	VL_ICMS_ST_APLICADO|VL_ICMS_FRT_APLICADO|VL_ICMS_DIF_APLICADO|
//
// Ate a versao 112 o registro tem apenas NUM_ITEM e COD_ITEM.
// Ref.: ACBrEFDBloco_G_Class.pas, WriteRegistroG140.
func (b *BlocoG) writeRegistroG140(parent *RegistroG130) {
	for _, r := range parent.RegistroG140 {
		b.Checkf(b.itemExisteNo0200(r.CodItem),
			"(G-G140) ITENS: O codigo do item %q nao existe no registro 0200!", r.CodItem)

		linha := b.LFillStr("G140", 0, false, '0') +
			b.LFillStr(r.NumItem, 3, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0')
		if b.versaoG() >= VlVersao113 {
			linha += b.DFill(r.Qtde, 2, false) +
				b.LFillStr(r.Unid, 0, false, '0') +
				b.DFill(r.VlICMSOpAplicado, 2, false) +
				b.DFill(r.VlICMSSTAplicado, 2, false) +
				b.DFill(r.VlICMSFrtAplicado, 2, false) +
				b.DFill(r.VlICMSDifAplicado, 2, false)
		}
		b.Add(linha, true)
		b.RegistroG990.QtdLinG++
		b.RegistroG140Count++
	}
}

// itemExisteNo0200 informa se o codigo do item foi declarado na tabela de itens
// do Bloco 0. Espelha o LocalizaRegistro do ACBr.
func (b *BlocoG) itemExisteNo0200(codItem string) bool {
	if b.Bloco0 == nil || b.Bloco0.Registro0001 == nil {
		return true // sem tabela carregada, nao ha como validar
	}
	for _, it := range b.Bloco0.Registro0001.Registro0200 {
		if it.CodItem == codItem {
			return true
		}
	}
	return false
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
			vlItemIR010(b, r)
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH010Count++

		b.writeRegistroH011(r)
		b.writeRegistroH020(r)
		b.writeRegistroH030(r)
	}
}

// inventarioApos072012 informa se o periodo permite os registros complementares
// do inventario. Ref.: ACBrEFDBloco_H_Class.pas.
func (b *BlocoH) inventarioApos072012() bool {
	return !b.DtIni.Before(time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC))
}

// versaoH devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoH) versaoH() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// writeRegistroH011 gera as linhas do registro H011 (proprietario do estoque,
// quando diferente do informante).
// Formato: |H011|CNPJ|
//
// So existe para periodo a partir de 07/2012.
func (b *BlocoH) writeRegistroH011(parent *RegistroH010) {
	if !b.inventarioApos072012() {
		return
	}
	for _, r := range parent.RegistroH011 {
		linha := b.LFillStr("H011", 0, false, '0') +
			b.LFillStr(r.CNPJ, 0, false, '0')
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH011Count++
	}
}

// writeRegistroH020 gera as linhas do registro H020 (informacao complementar
// do inventario).
// Formato: |H020|CST_ICMS|BC_ICMS|VL_ICMS|
//
// So existe a partir da versao 104 do leiaute e para periodo a partir de 07/2012.
func (b *BlocoH) writeRegistroH020(parent *RegistroH010) {
	if b.versaoH() < VlVersao104 || !b.inventarioApos072012() {
		return
	}
	for _, r := range parent.RegistroH020 {
		linha := b.LFillStr("H020", 0, false, '0') +
			b.LFillStr(r.CstICMS.String(), 0, false, '0') +
			b.DFill(r.BcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH020Count++
	}
}

// writeRegistroH030 gera as linhas do registro H030 (informacoes complementares
// do inventario das mercadorias sujeitas ao regime de substituicao tributaria).
// Formato: |H030|VL_ICMS_OP|VL_BC_ICMS_ST|VL_ICMS_ST|VL_FCP|
//
// Os quatro valores usam SEIS casas decimais, nao duas.
// So existe a partir da versao 104 do leiaute e para periodo a partir de 07/2012.
func (b *BlocoH) writeRegistroH030(parent *RegistroH010) {
	if b.versaoH() < VlVersao104 || !b.inventarioApos072012() {
		return
	}
	for _, r := range parent.RegistroH030 {
		linha := b.LFillStr("H030", 0, false, '0') +
			b.DFill(r.VlICMSOp, 6, false) +
			b.DFill(r.VlBcICMSST, 6, false) +
			b.DFill(r.VlICMSST, 6, false) +
			b.DFill(r.VlFCP, 6, false)
		b.Add(linha, true)
		b.RegistroH990.QtdLinH++
		b.RegistroH030Count++
	}
}

// vlItemIR010 devolve o campo VL_ITEM_IR do H010, inserido no leiaute apenas
// a partir de 01/2015. Antes disso o campo nao e emitido.
// Ref.: ACBrEFDBloco_H_Class.pas, WriteRegistroH010.
func vlItemIR010(b *BlocoH, r *RegistroH010) string {
	inicio := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	if b.DtIni.Before(inicio) {
		return ""
	}
	return b.DFill(r.VlItemIR, 2, true)
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

	if b.RegistroK001.IndDad == 0 {
		b.writeRegistroK010()
		b.writeRegistroK100()
	}
}

// versaoK devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoK) versaoK() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// leiauteRestrito replica ValidacaoVersaoeLeiaute do ACBr: a partir da versao
// 117, quem declara leiaute diferente do completo no K010 deixa de emitir uma
// parte dos registros. Quando devolve true, o registro nao e escrito.
// Ref.: ACBrEFDBloco_K_Class.pas, ValidacaoVersaoeLeiaute.
func (b *BlocoK) leiauteRestrito() bool {
	tipo := LeiauteCompleto
	if b.RegistroK001 != nil && b.RegistroK001.RegistroK010 != nil {
		tipo = b.RegistroK001.RegistroK010.IndTipoLeiaute
	}
	return b.versaoK() > VlVersao116 && tipo != LeiauteCompleto
}

// decQtdK devolve o numero de casas decimais das quantidades do Bloco K:
// tres ate a versao 111, seis a partir da 112.
func (b *BlocoK) decQtdK() int {
	if b.versaoK() < VlVersao112 {
		return 3
	}
	return 6
}

// writeRegistroK010 gera a linha do registro K010 (tipo de leiaute).
// Formato: |K010|IND_TIPO_LEIAUTE|
//
// So existe a partir da versao 116 do leiaute. E registro unico, nao lista.
func (b *BlocoK) writeRegistroK010() {
	if b.versaoK() < VlVersao116 || b.RegistroK001.RegistroK010 == nil {
		return
	}
	r := b.RegistroK001.RegistroK010
	linha := b.LFillStr("K010", 0, false, '0') +
		b.LFillInt(int64(r.IndTipoLeiaute), 0, false, '0')
	b.Add(linha, true)
	b.RegistroK990.QtdLinK++
	b.RegistroK010Count++
}

// writeRegistroK100 gera as linhas do registro K100 (periodo de apuracao).
// Formato: |K100|DT_INI|DT_FIN|
func (b *BlocoK) writeRegistroK100() {
	for _, r := range b.RegistroK001.RegistroK100 {
		b.Checkf(!r.DtIni.Before(b.DtIni) && !r.DtIni.After(b.DtFin),
			"A data inicial esta fora do periodo do EFD!")
		b.Checkf(!r.DtFin.Before(b.DtIni) && !r.DtFin.After(b.DtFin),
			"A data final esta fora do periodo do EFD!")

		linha := b.LFillStr("K100", 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK100Count++

		b.writeRegistroK200(r)
		b.writeRegistroK210(r)
		b.writeRegistroK220(r)
		b.writeRegistroK230(r)
		b.writeRegistroK250(r)
		b.writeRegistroK260(r)
		b.writeRegistroK270(r)
		b.writeRegistroK280(r)
		b.writeRegistroK290(r)
		b.writeRegistroK300(r)
	}
}

// writeRegistroK200 gera as linhas do registro K200 (estoque escriturado).
// Formato: |K200|DT_EST|COD_ITEM|QTD|IND_EST|COD_PART|
//
// QTD usa tres casas fixas, nao a regra por versao.
func (b *BlocoK) writeRegistroK200(parent *RegistroK100) {
	for _, r := range parent.RegistroK200 {
		b.Checkf(r.DtEst.Equal(parent.DtFin),
			"A data do estoque deve ser igual a data final do periodo de apuracao - campo DT_FIN do Registro K100")
		b.Checkf(!((r.IndEst == EstPropInformanteTerceiros || r.IndEst == EstPropTerceirosInformante) &&
			strings.TrimSpace(r.CodPart) == ""),
			"O campo COD_PART sera obrigatorio conforme informacao do campo IND_EST")

		linha := b.LFillStr("K200", 0, false, '0') +
			b.LFillDate(r.DtEst, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillInt(int64(r.IndEst), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK200Count++
	}
}

// writeRegistroK210 gera as linhas do registro K210 (desmontagem, item de origem).
// Formato: |K210|DT_INI_OS|DT_FIN_OS|COD_DOC_OS|COD_ITEM_ORI|QTD_ORI|
func (b *BlocoK) writeRegistroK210(parent *RegistroK100) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK210 {
		temDoc := r.CodDocOS != ""
		b.Checkf(!((temDoc || !r.DtFinOS.IsZero()) && r.DtIniOS.IsZero()),
			"O campo DT_INI_OS sera obrigatorio quando informado o campo COD_DOC_OS ou DT_FIN_OS")
		b.Checkf(!(!r.DtIniOS.IsZero() && !temDoc),
			"O campo COD_DOC_OS sera obrigatorio quando informado o campo DT_INI_OS.")
		b.Checkf(!(temDoc && r.DtIniOS.After(parent.DtFin)),
			"O campo DT_INI_OS deve ser menor ou igual a DT_FIN do registro K100.")
		b.Checkf(!(temDoc && (r.DtFinOS.Before(parent.DtIni) || r.DtFinOS.After(parent.DtFin))),
			"O campo DT_FIN_OS deve estar compreendido no periodo de apuracao do Registro K100")
		b.Checkf(!r.DtIniOS.After(r.DtFinOS),
			"O campo DT_INI_OS nao pode ser maior do que o campo DT_FIN_OS")

		linha := b.LFillStr("K210", 0, false, '0') +
			b.LFillDate(r.DtIniOS, "02012006", true) +
			b.LFillDate(r.DtFinOS, "02012006", true) +
			b.LFillStr(r.CodDocOS, 0, false, '0') +
			b.LFillStr(r.CodItemOri, 0, false, '0') +
			b.DFill(r.QtdOri, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK210Count++

		b.writeRegistroK215(r)
	}
}

// writeRegistroK215 gera as linhas do registro K215 (desmontagem, item de destino).
// Formato: |K215|COD_ITEM_DES|QTD_DES|
func (b *BlocoK) writeRegistroK215(parent *RegistroK210) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK215 {
		linha := b.LFillStr("K215", 0, false, '0') +
			b.LFillStr(r.CodItemDes, 0, false, '0') +
			b.DFill(r.QtdDes, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK215Count++
	}
}

// writeRegistroK220 gera as linhas do registro K220 (outras movimentacoes internas).
// Formato: |K220|DT_MOV|COD_ITEM_ORI|COD_ITEM_DEST|QTD|QTD_DEST|
//
// QTD_DEST so e emitido quando o periodo do arquivo comeca em 2018 ou depois --
// a data avaliada e a do arquivo, nao a do K100 nem a do K220.
func (b *BlocoK) writeRegistroK220(parent *RegistroK100) {
	for _, r := range parent.RegistroK220 {
		b.Checkf(!(r.DtMov.Before(parent.DtIni) || r.DtMov.After(parent.DtFin)),
			"A data deve estar compreendida no periodo informado nos campos DT_INI e DT_FIN do Registro K100")

		linha := b.LFillStr("K220", 0, false, '0') +
			b.LFillDate(r.DtMov, "02012006", true) +
			b.LFillStr(r.CodItemOri, 0, false, '0') +
			b.LFillStr(r.CodItemDest, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		if !b.DtIni.Before(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
			linha += b.DFill(r.QtdDest, b.decQtdK(), false)
		}
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK220Count++
	}
}

// writeRegistroK230 gera as linhas do registro K230 (itens produzidos).
// Formato: |K230|DT_INI_OP|DT_FIN_OP|COD_DOC_OP|COD_ITEM|QTD_ENC|
func (b *BlocoK) writeRegistroK230(parent *RegistroK100) {
	for _, r := range parent.RegistroK230 {
		temDoc := r.CodDocOP != ""
		b.Checkf(!((temDoc || !r.DtFinOP.IsZero()) && r.DtIniOP.IsZero()),
			"O campo DT_INI_OS sera obrigatorio conforme informacao do campo COD_DOC_OP ou DT_FIN_OP")
		b.Checkf(!((!r.DtIniOP.IsZero() || !r.DtFinOP.IsZero()) && !temDoc),
			"O campo COD_DOC_OP sera obrigatorio conforme informacao do campo DT_INI_OS ou DT_FIN_OS")

		linha := b.LFillStr("K230", 0, false, '0') +
			b.LFillDate(r.DtIniOP, "02012006", true) +
			b.LFillDate(r.DtFinOP, "02012006", true) +
			b.LFillStr(r.CodDocOP, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.QtdEnc, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK230Count++

		b.writeRegistroK235(r)
	}
}

// writeRegistroK235 gera as linhas do registro K235 (insumos consumidos).
// Formato: |K235|DT_SAIDA|COD_ITEM|QTD|COD_INS_SUBST|
func (b *BlocoK) writeRegistroK235(parent *RegistroK230) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK235 {
		b.Checkf(!r.DtSaida.Before(parent.DtIniOP),
			"A data de saida deve ser igual ou posterior ao inicio da producao, informado em DT_INI_OP do Registro K230")

		linha := b.LFillStr("K235", 0, false, '0') +
			b.LFillDate(r.DtSaida, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false) +
			b.LFillStr(r.CodInsSubst, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK235Count++
	}
}

// writeRegistroK250 gera as linhas do registro K250 (indus. por terceiros, produzidos).
// Formato: |K250|DT_PROD|COD_ITEM|QTD|
func (b *BlocoK) writeRegistroK250(parent *RegistroK100) {
	for _, r := range parent.RegistroK250 {
		b.Checkf(!(r.DtProd.Before(parent.DtIni) || r.DtProd.After(parent.DtFin)),
			"A data deve estar compreendida no periodo informado nos campos DT_INI e DT_FIN do Registro K100")

		linha := b.LFillStr("K250", 0, false, '0') +
			b.LFillDate(r.DtProd, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK250Count++

		b.writeRegistroK255(r)
	}
}

// writeRegistroK255 gera as linhas do registro K255 (indus. em terceiros, insumos).
// Formato: |K255|DT_CONS|COD_ITEM|QTD|COD_INS_SUBST|
func (b *BlocoK) writeRegistroK255(parent *RegistroK250) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK255 {
		linha := b.LFillStr("K255", 0, false, '0') +
			b.LFillDate(r.DtCons, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false) +
			b.LFillStr(r.CodInsSubst, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK255Count++
	}
}

// writeRegistroK260 gera as linhas do registro K260 (reprocessamento/reparo).
// Formato: |K260|COD_OP_OS|COD_ITEM|DT_SAIDA|QTD_SAIDA|DT_RET|QTD_RET|
func (b *BlocoK) writeRegistroK260(parent *RegistroK100) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK260 {
		b.Checkf(!r.DtSaida.After(parent.DtFin),
			"A data de saida deve ser menor ou igual a DT_FIN do Registro K100")

		linha := b.LFillStr("K260", 0, false, '0') +
			b.LFillStr(r.CodOpOS, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillDate(r.DtSaida, "02012006", true) +
			b.DFill(r.QtdSaida, b.decQtdK(), false) +
			b.LFillDate(r.DtRet, "02012006", true) +
			b.DFill(r.QtdRet, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK260Count++

		b.writeRegistroK265(r)
	}
}

// writeRegistroK265 gera as linhas do registro K265 (consumidas/retornadas).
// Formato: |K265|COD_ITEM|QTD_CONS|QTD_RET|
func (b *BlocoK) writeRegistroK265(parent *RegistroK260) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK265 {
		linha := b.LFillStr("K265", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.QtdCons, b.decQtdK(), false) +
			b.DFill(r.QtdRet, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK265Count++
	}
}

// writeRegistroK270 gera as linhas do registro K270 (correcao de apontamento).
// Formato: |K270|DT_INI_AP|DT_FIN_AP|COD_OP_OS|COD_ITEM|QTD_COR_POS|QTD_COR_NEG|ORIGEM|
//
// As quantidades usam nulo = (valor <= 0). Como o helper so suprime o valor
// exatamente zero, um valor negativo continua sendo emitido -- e o que o ACBr faz.
func (b *BlocoK) writeRegistroK270(parent *RegistroK100) {
	for _, r := range parent.RegistroK270 {
		linha := b.LFillStr("K270", 0, false, '0') +
			b.LFillDate(r.DtIniAP, "02012006", true) +
			b.LFillDate(r.DtFinAP, "02012006", true) +
			b.LFillStr(r.CodOpOS, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.QtdCorPos, b.decQtdK(), r.QtdCorPos <= 0) +
			b.DFill(r.QtdCorNeg, b.decQtdK(), r.QtdCorNeg <= 0) +
			b.LFillStr(r.Origem, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK270Count++

		b.writeRegistroK275(r)
	}
}

// writeRegistroK275 gera as linhas do registro K275 (correcao, insumos).
// Formato: |K275|COD_ITEM|QTD_COR_POS|QTD_COR_NEG|COD_INS_SUBST|
func (b *BlocoK) writeRegistroK275(parent *RegistroK270) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK275 {
		linha := b.LFillStr("K275", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.QtdCorPos, b.decQtdK(), r.QtdCorPos <= 0) +
			b.DFill(r.QtdCorNeg, b.decQtdK(), r.QtdCorNeg <= 0) +
			b.LFillStr(r.CodInsSubst, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK275Count++
	}
}

// writeRegistroK280 gera as linhas do registro K280 (correcao do estoque).
// Formato: |K280|DT_EST|COD_ITEM|QTD_COR_POS|QTD_COR_NEG|IND_EST|COD_PART|
//
// As quantidades usam tres casas fixas, nao a regra por versao.
func (b *BlocoK) writeRegistroK280(parent *RegistroK100) {
	for _, r := range parent.RegistroK280 {
		b.Checkf(r.DtEst.Before(parent.DtIni),
			"A data do estoque que esta sendo corrigido deve ser anterior a data informada no campo DT_INI do Registro K100")

		linha := b.LFillStr("K280", 0, false, '0') +
			b.LFillDate(r.DtEst, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.QtdCorPos, 3, r.QtdCorPos <= 0) +
			b.DFill(r.QtdCorNeg, 3, r.QtdCorNeg <= 0) +
			b.LFillInt(int64(r.IndEst), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK280Count++
	}
}

// writeRegistroK290 gera as linhas do registro K290 (producao conjunta, ordem).
// Formato: |K290|DT_INI_OP|DT_FIN_OP|COD_DOC_OP|
func (b *BlocoK) writeRegistroK290(parent *RegistroK100) {
	for _, r := range parent.RegistroK290 {
		temDoc := r.CodDocOP != ""
		b.Checkf(!((temDoc || !r.DtFinOP.IsZero()) && r.DtIniOP.IsZero()),
			"O campo DT_INI_OP sera obrigatorio conforme informacao do campo COD_DOC_OP ou DT_FIN_OP")
		b.Checkf(!((!r.DtIniOP.IsZero() || !r.DtFinOP.IsZero()) && !temDoc),
			"O campo COD_DOC_OP sera obrigatorio conforme informacao do campo DT_INI_OP ou DT_FIN_OP")

		linha := b.LFillStr("K290", 0, false, '0') +
			b.LFillDate(r.DtIniOP, "02012006", true) +
			b.LFillDate(r.DtFinOP, "02012006", true) +
			b.LFillStr(r.CodDocOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK290Count++

		b.writeRegistroK291(r)
		b.writeRegistroK292(r)
	}
}

// writeRegistroK291 gera as linhas do registro K291 (producao conjunta, produzidos).
// Formato: |K291|COD_ITEM|QTD|
func (b *BlocoK) writeRegistroK291(parent *RegistroK290) {
	for _, r := range parent.RegistroK291 {
		b.Checkf(r.Qtd > 0,
			"Nao e admitida quantidade negativa, valor deve ser maior que zero, informado em QTD do Registro K291")

		linha := b.LFillStr("K291", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK291Count++
	}
}

// writeRegistroK292 gera as linhas do registro K292 (producao conjunta, insumos).
// Formato: |K292|COD_ITEM|QTD|
func (b *BlocoK) writeRegistroK292(parent *RegistroK290) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK292 {
		b.Checkf(r.Qtd > 0,
			"Nao e admitida quantidade negativa, valor deve ser maior que zero, informado em QTD do Registro K292")

		linha := b.LFillStr("K292", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK292Count++
	}
}

// writeRegistroK300 gera as linhas do registro K300 (prod. conjunta em terceiros).
// Formato: |K300|DT_PROD|
func (b *BlocoK) writeRegistroK300(parent *RegistroK100) {
	for _, r := range parent.RegistroK300 {
		b.Checkf(!(r.DtProd.After(parent.DtFin) || r.DtProd.Before(parent.DtIni)),
			"A data deve estar compreendida no periodo informado nos campos DT_INI e DT_FIN do Registro K100")

		linha := b.LFillStr("K300", 0, false, '0') +
			b.LFillDate(r.DtProd, "02012006", true)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK300Count++

		b.writeRegistroK301(r)
		b.writeRegistroK302(r)
	}
}

// writeRegistroK301 gera as linhas do registro K301.
// Formato: |K301|COD_ITEM|QTD|
func (b *BlocoK) writeRegistroK301(parent *RegistroK300) {
	for _, r := range parent.RegistroK301 {
		b.Checkf(r.Qtd > 0,
			"Nao e admitida quantidade negativa, valor deve ser maior que zero, informado em QTD do Registro K301")

		linha := b.LFillStr("K301", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK301Count++
	}
}

// writeRegistroK302 gera as linhas do registro K302.
// Formato: |K302|COD_ITEM|QTD|
func (b *BlocoK) writeRegistroK302(parent *RegistroK300) {
	if b.leiauteRestrito() {
		return
	}
	for _, r := range parent.RegistroK302 {
		b.Checkf(r.Qtd > 0,
			"Nao e admitida quantidade negativa, valor deve ser maior que zero, informado em QTD do Registro K302")

		linha := b.LFillStr("K302", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, b.decQtdK(), false)
		b.Add(linha, true)
		b.RegistroK990.QtdLinK++
		b.RegistroK302Count++
	}
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

	if b.Registro1001.IndDad == 0 {
		b.writeRegistro1010()
	}
}

// writeRegistro1010 gera as linhas do registro 1010 (obrigatoriedade de
// registros do Bloco 1).
// Formato: |1010|IND_EXP|IND_CCRF|IND_COMB|IND_USINA|IND_VA|IND_EE|IND_CART|
//
//	IND_FORM|IND_AER|IND_GIAF1|IND_GIAF3|IND_GIAF4|IND_REST_RESSARC_COMPL_ICMS|
//
// IND_GIAF1/3/4 so existem a partir da versao 112 do leiaute, e
// IND_REST_RESSARC_COMPL_ICMS a partir da 113.
// Ref.: ACBrEFDBloco_1_Class.pas, WriteRegistro1010.
func (b *Bloco1) writeRegistro1010() {
	ver := VlVersao100
	if b.Bloco0 != nil && b.Bloco0.Registro0000 != nil {
		ver = b.Bloco0.Registro0000.CodVer
	}

	for _, r := range b.Registro1001.Registro1010 {
		linha := b.LFillStr("1010", 0, false, '0') +
			b.LFillStr(r.IndExp, 0, false, '0') +
			b.LFillStr(r.IndCCRF, 0, false, '0') +
			b.LFillStr(r.IndComb, 0, false, '0') +
			b.LFillStr(r.IndUsina, 0, false, '0') +
			b.LFillStr(r.IndVA, 0, false, '0') +
			b.LFillStr(r.IndEE, 0, false, '0') +
			b.LFillStr(r.IndCart, 0, false, '0') +
			b.LFillStr(r.IndForm, 0, false, '0') +
			b.LFillStr(r.IndAer, 0, false, '0')
		if ver >= VlVersao112 {
			linha += b.LFillStr(r.IndGIAF1, 0, false, '0') +
				b.LFillStr(r.IndGIAF3, 0, false, '0') +
				b.LFillStr(r.IndGIAF4, 0, false, '0')
		}
		if ver >= VlVersao113 {
			linha += b.LFillStr(r.IndRestRessarcComplICMS, 0, false, '0')
		}
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1010Count++
	}
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
