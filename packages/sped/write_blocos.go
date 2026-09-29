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
		b.LFillInt(int64(b.RegistroB001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroB990.QtdLinB++

	if b.RegistroB001.IndMov == 0 {
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
		b.LFillInt(int64(b.RegistroC001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroC990.QtdLinC++

	if b.RegistroC001.IndMov == 0 {
		b.writeRegistroC100()
		b.writeRegistroC300()
		b.writeRegistroC350()
		b.writeRegistroC400()
		b.writeRegistroC495()
		b.writeRegistroC500()
		b.writeRegistroC600()
		b.writeRegistroC700()
		b.writeRegistroC800()
		b.writeRegistroC860()
	}
}

// versaoC devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoC) versaoC() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// ufDoInformante devolve a UF declarada no registro 0000.
func (b *BlocoC) ufDoInformante() string {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return ""
	}
	return b.Bloco0.Registro0000.UF
}

// ressarcimentoConv monta o bloco de campos de ressarcimento compartilhado por
// C330, C380, C430 e C480. Os quatro ultimos campos sao opcionais: nao
// informados saem vazios, diferente de zero.
// Os parametros repetem o nome do campo do struct de proposito: a auditoria de
// campos (ferramentas/comparar-campos.py) expande o helper e compara os nomes
// posicao a posicao contra o writer do ACBr. Abreviar aqui cega a auditoria.
func (b *BlocoC) ressarcimentoConv(codMotRestCompl string, quantConv float64, unid string,
	vlUnitConv, vlUnitICMSNaOperacaoConv, vlUnitICMSOpConv, vlUnitICMSOpEstoqueConv,
	vlUnitICMSSTEstoqueConv, vlUnitFcpICMSSTEstoqueConv float64,
	vlUnitICMSSTConvRest, vlUnitFcpSTConvRest,
	vlUnitICMSSTConvCompl, vlUnitFcpSTConvCompl *float64) string {
	return b.LFillStr(codMotRestCompl, 0, false, '0') +
		b.LFillFloat(quantConv, 0, 6, false, '0', "") +
		b.LFillStr(unid, 0, false, '0') +
		b.LFillFloat(vlUnitConv, 0, 6, false, '0', "") +
		b.LFillFloat(vlUnitICMSNaOperacaoConv, 0, 6, false, '0', "") +
		b.LFillFloat(vlUnitICMSOpConv, 0, 6, false, '0', "") +
		b.LFillFloat(vlUnitICMSOpEstoqueConv, 0, 6, false, '0', "") +
		b.LFillFloat(vlUnitICMSSTEstoqueConv, 0, 6, false, '0', "") +
		b.LFillFloat(vlUnitFcpICMSSTEstoqueConv, 0, 6, false, '0', "") +
		b.VDFill(vlUnitICMSSTConvRest, 6) +
		b.VDFill(vlUnitFcpSTConvRest, 6) +
		b.VDFill(vlUnitICMSSTConvCompl, 6) +
		b.VDFill(vlUnitFcpSTConvCompl, 6)
}

// writeRegistroC105 gera as linhas do registro C105.
// Formato: |C105|OPER|UF|
func (b *BlocoC) writeRegistroC105(parent *RegistroC100) {
	for _, r := range parent.RegistroC105 {
		linha := b.LFillStr("C105", 0, false, '0') +
			b.LFillInt(int64(r.Oper), 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC105Count++
	}
}

// writeRegistroC115 gera as linhas do registro C115.
func (b *BlocoC) writeRegistroC115(parent *RegistroC110) {
	for _, r := range parent.RegistroC115 {
		linha := b.LFillStr("C115", 0, false, '0') +
			b.LFillInt(int64(r.IndCarga), 0, false, '0') +
			b.LFillStr(r.CNPJCol, 0, false, '0') +
			b.LFillStr(r.IECol, 0, false, '0') +
			b.LFillStr(r.CPFCol, 0, false, '0') +
			b.LFillStr(r.CodMunCol, 0, false, '0') +
			b.LFillStr(r.CNPJEntg, 0, false, '0') +
			b.LFillStr(r.IEEntg, 0, false, '0') +
			b.LFillStr(r.CPFEntg, 0, false, '0') +
			b.LFillStr(r.CodMunEntg, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC115Count++
	}
}

// writeRegistroC116 gera as linhas do registro C116.
func (b *BlocoC) writeRegistroC116(parent *RegistroC110) {
	for _, r := range parent.RegistroC116 {
		linha := b.LFillStr("C116", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.NrSat, 0, false, '0') +
			b.LFillStr(r.ChvCFe, 0, false, '0') +
			b.LFillStr(r.NumCFe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC116Count++
	}
}

// writeRegistroC120 gera as linhas do registro C120.
func (b *BlocoC) writeRegistroC120(parent *RegistroC100) {
	for _, r := range parent.RegistroC120 {
		linha := b.LFillStr("C120", 0, false, '0') +
			b.LFillInt(int64(r.CodDocImp), 0, false, '0') +
			b.LFillStr(r.NumDocImp, 0, false, '0') +
			b.DFill(r.PisImp, 2, true) +
			b.DFill(r.CofinsImp, 2, true) +
			b.LFillStr(r.NumACDraw, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC120Count++
	}
}

// writeRegistroC130 gera as linhas do registro C130.
//
// Os sete valores usam LFill com tamanho 2 e ZERO casas decimais, como no ACBr.
func (b *BlocoC) writeRegistroC130(parent *RegistroC100) {
	for _, r := range parent.RegistroC130 {
		linha := b.LFillStr("C130", 0, false, '0') +
			b.LFillFloat(r.VlServNT, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlBcISSQN, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlISSQN, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlBcIRRF, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlIRRF, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlBcPrev, 2, 0, false, '0', "") +
			b.LFillFloat(r.VlPrev, 2, 0, false, '0', "")
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC130Count++
	}
}

// writeRegistroC140 gera as linhas do registro C140 (fatura).
func (b *BlocoC) writeRegistroC140(parent *RegistroC100) {
	for _, r := range parent.RegistroC140 {
		linha := b.LFillStr("C140", 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.IndTit.String(), 0, false, '0') +
			b.LFillStr(r.DescTit, 0, false, '0') +
			b.LFillStr(r.NumTit, 0, false, '0') +
			b.LFillInt(int64(r.QtdParc), 2, false, '0') +
			b.DFill(r.VlTit, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC140Count++

		b.writeRegistroC141(r)
	}
}

// writeRegistroC141 gera as linhas do registro C141.
func (b *BlocoC) writeRegistroC141(parent *RegistroC140) {
	for _, r := range parent.RegistroC141 {
		linha := b.LFillStr("C141", 0, false, '0') +
			b.LFillStr(r.NumParc, 0, false, '0') +
			b.LFillDate(r.DtVcto, "02012006", true) +
			b.DFill(r.VlParc, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC141Count++
	}
}

// writeRegistroC160 gera as linhas do registro C160.
func (b *BlocoC) writeRegistroC160(parent *RegistroC100) {
	for _, r := range parent.RegistroC160 {
		linha := b.LFillStr("C160", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.LFillInt(int64(r.QtdVol), 0, false, '0') +
			b.DFill(r.PesoBrt, 2, false) +
			b.DFill(r.PesoLiq, 2, false) +
			b.LFillStr(r.UFID, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC160Count++
	}
}

// writeRegistroC165 gera as linhas do registro C165.
func (b *BlocoC) writeRegistroC165(parent *RegistroC100) {
	for _, r := range parent.RegistroC165 {
		linha := b.LFillStr("C165", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.LFillStr(r.CodAut, 0, false, '0') +
			b.LFillStr(r.NrPasse, 0, false, '0') +
			b.LFillStr(r.Hora, 0, false, '0') +
			b.LFillStr(r.Temper, 0, false, '0') +
			b.LFillInt(int64(r.QtdVol), 0, false, '0') +
			b.DFill(r.PesoBrt, 2, false) +
			b.DFill(r.PesoLiq, 2, false) +
			b.LFillStr(r.NomMot, 0, false, '0') +
			b.LFillStr(r.CPF, 0, false, '0') +
			b.LFillStr(r.UFID, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC165Count++
	}
}

// writeRegistroC171 gera as linhas do registro C171.
func (b *BlocoC) writeRegistroC171(parent *RegistroC170) {
	for _, r := range parent.RegistroC171 {
		linha := b.LFillStr("C171", 0, false, '0') +
			b.LFillStr(r.NumTanque, 3, false, '0') +
			b.DFill(r.Qtde, 3, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC171Count++
	}
}

// writeRegistroC172 gera as linhas do registro C172.
func (b *BlocoC) writeRegistroC172(parent *RegistroC170) {
	for _, r := range parent.RegistroC172 {
		linha := b.LFillStr("C172", 0, false, '0') +
			b.DFill(r.VlBcISSQN, 2, false) +
			b.DFill(r.AliqISSQN, 2, false) +
			b.DFill(r.VlISSQN, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC172Count++
	}
}

// writeRegistroC173 gera as linhas do registro C173.
func (b *BlocoC) writeRegistroC173(parent *RegistroC170) {
	for _, r := range parent.RegistroC173 {
		linha := b.LFillStr("C173", 0, false, '0') +
			b.LFillStr(r.LoteMed, 0, false, '0') +
			b.DFill(r.QtdItem, 3, false) +
			b.LFillDate(r.DtFab, "02012006", true) +
			b.LFillDate(r.DtVal, "02012006", true) +
			b.LFillInt(int64(r.IndMed), 0, false, '0') +
			b.LFillInt(int64(r.TpProd), 0, false, '0') +
			b.DFill(r.VlTabMax, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC173Count++
	}
}

// writeRegistroC174 gera as linhas do registro C174.
func (b *BlocoC) writeRegistroC174(parent *RegistroC170) {
	for _, r := range parent.RegistroC174 {
		linha := b.LFillStr("C174", 0, false, '0') +
			b.LFillInt(int64(r.IndArm), 0, false, '0') +
			b.LFillStr(r.NumArm, 0, false, '0') +
			b.LFillStr(r.DescrCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC174Count++
	}
}

// writeRegistroC175 gera as linhas do registro C175.
func (b *BlocoC) writeRegistroC175(parent *RegistroC170) {
	for _, r := range parent.RegistroC175 {
		linha := b.LFillStr("C175", 0, false, '0') +
			b.LFillInt(indVeicOperInt(r.IndVeicOper), 0, false, '0') +
			b.LFillStr(r.CNPJ, 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0') +
			b.LFillStr(r.ChassiVeic, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC175Count++
	}
}

// indVeicOperInt devolve IND_VEIC_OPER como inteiro, com 9 no default.
func indVeicOperInt(v IndVeicOper) int64 {
	switch v {
	case VeicVendaPConcess:
		return 0
	case VeicFaturaDireta:
		return 1
	case VeicVendaDireta:
		return 2
	case VeicVendaDConcess:
		return 3
	}
	return 9
}

// writeRegistroC176 gera as linhas do registro C176 (ressarcimento de ICMS ST).
//
// O bloco a partir de CHAVE_NFE_ULT_E so existe de 2017 em diante, e
// VL_UNIT_RES_FCP_ST so acima da versao 111.
func (b *BlocoC) writeRegistroC176(parent *RegistroC170) {
	for _, r := range parent.RegistroC176 {
		linha := b.LFillStr("C176", 0, false, '0') +
			b.LFillStr(r.CodModUltE, 0, false, '0') +
			b.LFillStr(r.NumDocUltE, 0, false, '0') +
			b.LFillStr(r.SerUltE, 0, false, '0') +
			b.LFillDate(r.DtUltE, "02012006", true) +
			b.LFillStr(r.CodPartUltE, 0, false, '0') +
			b.DFill(r.QuantUltE, 3, false) +
			b.DFill(r.VlUnitUltE, 3, false) +
			b.DFill(r.VlUnitBcST, 3, false)
		if !b.DtIni.Before(time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)) {
			linha += b.LFillStr(r.ChaveNfeUltE, 0, false, '0') +
				b.LFillStr(r.NumItemUltE, 0, false, '0') +
				b.DFill(r.VlUnitBcICMSUltE, 2, true) +
				b.DFill(r.AliqICMSUltE, 2, true) +
				b.DFill(r.VlUnitLimiteBcICMSUltE, 2, true) +
				b.DFill(r.VlUnitICMSUltE, 3, true) +
				b.DFill(r.AliqSTUltE, 2, true) +
				b.DFill(r.VlUnitRes, 3, true) +
				b.LFillStr(r.CodRespRet, 0, false, '0') +
				b.LFillStr(r.CodMotRes.String(), 0, false, '0') +
				b.LFillStr(r.ChaveNfeRet, 0, false, '0') +
				b.LFillStr(r.CodPartNfeRet, 0, false, '0') +
				b.LFillStr(r.SerNfeRet, 0, false, '0') +
				b.LFillStr(r.NumNfeRet, 0, false, '0') +
				b.LFillStr(r.ItemNfeRet, 0, false, '0') +
				b.LFillStr(r.CodDA, 0, false, '0') +
				b.LFillStr(r.NumDA, 0, false, '0')
			if b.versaoC() > VlVersao111 {
				linha += b.LFillFloat(r.VlUnitResFcpST, 0, 3, false, '0', "")
			}
		}
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC176Count++
	}
}

// writeRegistroC177 gera as linhas do registro C177.
//
// Ate a versao 111 o registro traz COD_SELO_IPI e QT_SELO_IPI; da 112 em diante,
// apenas COD_INF_ITEM. So e emitido para estabelecimento em Pernambuco.
func (b *BlocoC) writeRegistroC177(parent *RegistroC170) {
	for _, r := range parent.RegistroC177 {
		linha := b.LFillStr("C177", 0, false, '0')
		if b.versaoC() < VlVersao112 {
			linha += b.LFillStr(r.CodSeloIPI, 0, false, '0') +
				b.LFillFloat(r.QtSeloIPI, 0, 0, false, '0', "")
		} else {
			linha += b.LFillStr(r.CodInfItem, 0, false, '0')
		}
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC177Count++
	}
}

// writeRegistroC178 gera as linhas do registro C178.
func (b *BlocoC) writeRegistroC178(parent *RegistroC170) {
	for _, r := range parent.RegistroC178 {
		linha := b.LFillStr("C178", 0, false, '0') +
			b.LFillStr(r.ClEnq, 0, false, '0') +
			b.DFill(r.VlUnid, 2, false) +
			b.DFill(r.QuantPad, 3, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC178Count++
	}
}

// writeRegistroC179 gera as linhas do registro C179.
func (b *BlocoC) writeRegistroC179(parent *RegistroC170) {
	for _, r := range parent.RegistroC179 {
		linha := b.LFillStr("C179", 0, false, '0') +
			b.DFill(r.BcSTOrigDest, 2, false) +
			b.DFill(r.ICMSSTRep, 2, false) +
			b.DFill(r.ICMSSTCompl, 2, false) +
			b.DFill(r.BcRet, 2, false) +
			b.DFill(r.ICMSRet, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC179Count++
	}
}

// writeRegistroC180 gera as linhas do registro C180.
func (b *BlocoC) writeRegistroC180(parent *RegistroC170) {
	for _, r := range parent.RegistroC180 {
		linha := b.LFillStr("C180", 0, false, '0') +
			b.LFillStr(r.CodRespRet, 0, false, '0') +
			b.DFill(r.QuantConv, 6, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlUnitConv, 6, false) +
			b.DFill(r.VlUnitICMSOpConv, 6, false) +
			b.DFill(r.VlUnitBcICMSSTConv, 6, false) +
			b.DFill(r.VlUnitICMSSTConv, 6, false) +
			b.DFill(r.VlUnitFcpSTConv, 6, false) +
			b.LFillStr(r.CodDA, 0, false, '0') +
			b.LFillStr(r.NumDA, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC180Count++
	}
}

// writeRegistroC181 gera as linhas do registro C181.
func (b *BlocoC) writeRegistroC181(parent *RegistroC170) {
	for _, r := range parent.RegistroC181 {
		linha := b.LFillStr("C181", 0, false, '0') +
			b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
			b.LFillFloat(r.QuantConv, 0, 6, false, '0', "") +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.LFillStr(r.CodModSaida, 0, false, '0') +
			b.LFillStr(r.SerieSaida, 0, false, '0') +
			b.LFillStr(r.EcfFabSaida, 0, false, '0') +
			b.LFillStr(r.NumDocSaida, 0, false, '0') +
			b.LFillStr(r.ChvDfeSaida, 0, false, '0') +
			b.LFillDate(r.DtDocSaida, "02012006", true) +
			b.LFillStr(r.NumItemSaida, 0, false, '0') +
			b.DFill(r.VlUnitConvSaida, 6, true) +
			b.VDFill(r.VlUnitICMSOpEstoqueConvSaida, 6) +
			b.VDFill(r.VlUnitICMSSTEstoqueConvSaida, 6) +
			b.VDFill(r.VlUnitFcpICMSSTEstoqueConvSaida, 6) +
			b.VDFill(r.VlUnitICMSNaOperacaoConvSaida, 6) +
			b.VDFill(r.VlUnitICMSOpConvSaida, 6) +
			b.VDFill(r.VlUnitICMSSTConvRest, 6) +
			b.VDFill(r.VlUnitFcpSTConvRest, 6) +
			b.VDFill(r.VlUnitICMSSTConvCompl, 6) +
			b.VDFill(r.VlUnitFcpSTConvCompl, 6)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC181Count++
	}
}

// writeRegistroC185 gera as linhas do registro C185.
func (b *BlocoC) writeRegistroC185(parent *RegistroC100) {
	for _, r := range parent.RegistroC185 {
		linha := b.LFillStr("C185", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
			b.DFill(r.QuantConv, 6, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlUnitConv, 6, false) +
			b.VDFill(r.VlUnitICMSNaOperacaoConv, 6) +
			b.VDFill(r.VlUnitICMSOpConv, 6) +
			b.VDFill(r.VlUnitICMSOpEstoqueConv, 6) +
			b.VDFill(r.VlUnitICMSSTEstoqueConv, 6) +
			b.VDFill(r.VlUnitFcpICMSSTEstoqueConv, 6) +
			b.VDFill(r.VlUnitICMSSTConvRest, 6) +
			b.VDFill(r.VlUnitFcpSTConvRest, 6) +
			b.VDFill(r.VlUnitICMSSTConvCompl, 6) +
			b.VDFill(r.VlUnitFcpSTConvCompl, 6)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC185Count++
	}
}

// writeRegistroC186 gera as linhas do registro C186.
func (b *BlocoC) writeRegistroC186(parent *RegistroC100) {
	for _, r := range parent.RegistroC186 {
		linha := b.LFillStr("C186", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
			b.DFill(r.QuantConv, 6, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.LFillStr(r.CodModEntrada, 0, false, '0') +
			b.LFillStr(r.SerieEntrada, 0, false, '0') +
			b.LFillStr(r.NumDocEntrada, 0, false, '0') +
			b.LFillStr(r.ChvDfeEntrada, 0, false, '0') +
			b.LFillDate(r.DtDocEntrada, "02012006", true) +
			b.LFillStr(r.NumItemEntrada, 0, false, '0') +
			b.DFill(r.VlUnitConvEntrada, 6, false) +
			b.DFill(r.VlUnitICMSOpConvEntrada, 6, false) +
			b.DFill(r.VlUnitBcICMSSTConvEntrada, 6, false) +
			b.DFill(r.VlUnitICMSSTConvEntrada, 6, false) +
			b.DFill(r.VlUnitFcpSTConvEntrada, 6, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC186Count++
	}
}

// writeRegistroC191 gera as linhas do registro C191.
// So existe acima da versao 111.
func (b *BlocoC) writeRegistroC191(parent *RegistroC190) {
	if b.versaoC() <= VlVersao111 {
		return
	}
	for _, r := range parent.RegistroC191 {
		linha := b.LFillStr("C191", 0, false, '0') +
			b.DFill(r.VlFcpOp, 2, false) +
			b.DFill(r.VlFcpST, 2, false) +
			b.DFill(r.VlFcpRet, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC191Count++
	}
}

// writeRegistroC300 gera as linhas do registro C300 (resumo diario de NFVC).
func (b *BlocoC) writeRegistroC300() {
	for _, r := range b.RegistroC001.RegistroC300 {
		linha := b.LFillStr("C300", 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDocIni, 0, false, '0') +
			b.LFillStr(r.NumDocFin, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC300Count++

		b.writeRegistroC310(r)
		b.writeRegistroC320(r)
	}
}

// writeRegistroC310 gera as linhas do registro C310.
func (b *BlocoC) writeRegistroC310(parent *RegistroC300) {
	for _, r := range parent.RegistroC310 {
		linha := b.LFillStr("C310", 0, false, '0') +
			b.LFillStr(r.NumDocCanc, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC310Count++
	}
}

// writeRegistroC320 gera as linhas do registro C320.
func (b *BlocoC) writeRegistroC320(parent *RegistroC300) {
	for _, r := range parent.RegistroC320 {
		linha := b.LFillStr("C320", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC320Count++

		b.writeRegistroC321(r)
	}
}

// writeRegistroC321 gera as linhas do registro C321.
func (b *BlocoC) writeRegistroC321(parent *RegistroC320) {
	for _, r := range parent.RegistroC321 {
		linha := b.LFillStr("C321", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC321Count++

		b.writeRegistroC330(r)
	}
}

// writeRegistroC330 gera as linhas do registro C330.
func (b *BlocoC) writeRegistroC330(parent *RegistroC321) {
	for _, r := range parent.RegistroC330 {
		linha := b.LFillStr("C330", 0, false, '0') +
			b.ressarcimentoConv(r.CodMotRestCompl, r.QuantConv, r.Unid,
				r.VlUnitConv, r.VlUnitICMSNaOperacaoConv, r.VlUnitICMSOpConv,
				r.VlUnitICMSOpEstoqueConv, r.VlUnitICMSSTEstoqueConv,
				r.VlUnitFcpICMSSTEstoqueConv, r.VlUnitICMSSTConvRest,
				r.VlUnitFcpSTConvRest, r.VlUnitICMSSTConvCompl, r.VlUnitFcpSTConvCompl)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC330Count++
	}
}

// writeRegistroC350 gera as linhas do registro C350 (NF de venda a consumidor).
func (b *BlocoC) writeRegistroC350() {
	for _, r := range b.RegistroC001.RegistroC350 {
		linha := b.LFillStr("C350", 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.SubSer, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CNPJCPF, 0, false, '0') +
			b.DFill(r.VlMerc, 2, false) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC350Count++

		b.writeRegistroC370(r)
		b.writeRegistroC390(r)
	}
}

// writeRegistroC370 gera as linhas do registro C370.
// VL_ITEM e VL_DESC usam tres casas decimais neste registro.
func (b *BlocoC) writeRegistroC370(parent *RegistroC350) {
	for _, r := range parent.RegistroC370 {
		linha := b.LFillStr("C370", 0, false, '0') +
			b.LFillStr(r.NumItem, 3, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.LFillFloat(r.VlItem, 0, 3, false, '0', "") +
			b.LFillFloat(r.VlDesc, 0, 3, false, '0', "")
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC370Count++

		b.writeRegistroC380(r)
	}
}

// writeRegistroC380 gera as linhas do registro C380.
func (b *BlocoC) writeRegistroC380(parent *RegistroC370) {
	for _, r := range parent.RegistroC380 {
		linha := b.LFillStr("C380", 0, false, '0') +
			b.ressarcimentoConv(r.CodMotRestCompl, r.QuantConv, r.Unid,
				r.VlUnitConv, r.VlUnitICMSNaOperacaoConv, r.VlUnitICMSOpConv,
				r.VlUnitICMSOpEstoqueConv, r.VlUnitICMSSTEstoqueConv,
				r.VlUnitFcpICMSSTEstoqueConv, r.VlUnitICMSSTConvRest,
				r.VlUnitFcpSTConvRest, r.VlUnitICMSSTConvCompl, r.VlUnitFcpSTConvCompl) +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC380Count++
	}
}

// writeRegistroC390 gera as linhas do registro C390.
// Diferente do C320 e do C490, ALIQ_ICMS aqui nao tem tamanho fixo.
func (b *BlocoC) writeRegistroC390(parent *RegistroC350) {
	for _, r := range parent.RegistroC390 {
		linha := b.LFillStr("C390", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC390Count++
	}
}

// writeRegistroC400 gera as linhas do registro C400 (equipamento ECF).
func (b *BlocoC) writeRegistroC400() {
	for _, r := range b.RegistroC001.RegistroC400 {
		linha := b.LFillStr("C400", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.EcfMod, 0, false, '0') +
			b.LFillStr(r.EcfFab, 0, false, '0') +
			b.LFillStr(r.EcfCx, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC400Count++

		b.writeRegistroC405(r)
	}
}

// writeRegistroC405 gera as linhas do registro C405 (reducao Z).
// NUM_COO_FIN passou de 6 para 9 digitos em 10/2013.
func (b *BlocoC) writeRegistroC405(parent *RegistroC400) {
	tam := 6
	if !b.DtIni.Before(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC)) {
		tam = 9
	}
	for _, r := range parent.RegistroC405 {
		linha := b.LFillStr("C405", 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillInt(int64(r.Cro), 3, false, '0') +
			b.LFillInt(int64(r.Crz), 6, false, '0') +
			b.LFillInt(int64(r.NumCooFin), tam, false, '0') +
			b.DFill(r.GtFin, 2, false) +
			b.DFill(r.VlBrt, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC405Count++

		b.writeRegistroC410(r)
		b.writeRegistroC420(r)
		b.writeRegistroC460(r)
		b.writeRegistroC490(r)
	}
}

// writeRegistroC410 gera as linhas do registro C410.
func (b *BlocoC) writeRegistroC410(parent *RegistroC405) {
	for _, r := range parent.RegistroC410 {
		linha := b.LFillStr("C410", 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC410Count++
	}
}

// writeRegistroC420 gera as linhas do registro C420.
func (b *BlocoC) writeRegistroC420(parent *RegistroC405) {
	for _, r := range parent.RegistroC420 {
		linha := b.LFillStr("C420", 0, false, '0') +
			b.LFillStr(r.CodTotPar, 0, false, '0') +
			b.DFill(r.VlrAcumTot, 2, false) +
			b.LFillInt(int64(r.NrTot), 2, true, '0') +
			b.LFillStr(r.DescrNrTot, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC420Count++

		b.writeRegistroC425(r)
	}
}

// writeRegistroC425 gera as linhas do registro C425.
func (b *BlocoC) writeRegistroC425(parent *RegistroC420) {
	for _, r := range parent.RegistroC425 {
		b.Checkf(b.itemExisteNo0200C(r.CodItem),
			"(C-C425) ITENS: O codigo do item %q nao existe no registro 0200!", r.CodItem)

		linha := b.LFillStr("C425", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC425Count++

		b.writeRegistroC430(r)
	}
}

// itemExisteNo0200C informa se o codigo do item foi declarado na tabela de
// itens do Bloco 0.
func (b *BlocoC) itemExisteNo0200C(codItem string) bool {
	if b.Bloco0 == nil || b.Bloco0.Registro0001 == nil {
		return true
	}
	for _, it := range b.Bloco0.Registro0001.Registro0200 {
		if it.CodItem == codItem {
			return true
		}
	}
	return false
}

// writeRegistroC430 gera as linhas do registro C430.
func (b *BlocoC) writeRegistroC430(parent *RegistroC425) {
	for _, r := range parent.RegistroC430 {
		linha := b.LFillStr("C430", 0, false, '0') +
			b.ressarcimentoConv(r.CodMotRestCompl, r.QuantConv, r.Unid,
				r.VlUnitConv, r.VlUnitICMSNaOperacaoConv, r.VlUnitICMSOpConv,
				r.VlUnitICMSOpEstoqueConv, r.VlUnitICMSSTEstoqueConv,
				r.VlUnitFcpICMSSTEstoqueConv, r.VlUnitICMSSTConvRest,
				r.VlUnitFcpSTConvRest, r.VlUnitICMSSTConvCompl, r.VlUnitFcpSTConvCompl) +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC430Count++
	}
}

// writeRegistroC460 gera as linhas do registro C460.
// NUM_DOC passou de 6 para 9 digitos em 10/2013. Os filhos saem na ordem
// C470 e depois C465, como no ACBr.
func (b *BlocoC) writeRegistroC460(parent *RegistroC405) {
	tam := 6
	if !b.DtIni.Before(time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC)) {
		tam = 9
	}
	for _, r := range parent.RegistroC460 {
		linha := b.LFillStr("C460", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.NumDoc, tam, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, true) +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.LFillStr(r.CPFCNPJ, 0, false, '0') +
			b.LFillStr(r.NomAdq, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC460Count++

		b.writeRegistroC470(r)
		b.writeRegistroC465(r)
	}
}

// writeRegistroC465 gera as linhas do registro C465.
func (b *BlocoC) writeRegistroC465(parent *RegistroC460) {
	for _, r := range parent.RegistroC465 {
		linha := b.LFillStr("C465", 0, false, '0') +
			b.LFillStr(r.ChvCFe, 0, false, '0') +
			b.LFillStr(r.NumCCF, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC465Count++
	}
}

// writeRegistroC470 gera as linhas do registro C470.
func (b *BlocoC) writeRegistroC470(parent *RegistroC460) {
	for _, r := range parent.RegistroC470 {
		linha := b.LFillStr("C470", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.DFill(r.QtdCanc, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC470Count++

		b.writeRegistroC480(r)
	}
}

// writeRegistroC480 gera as linhas do registro C480.
func (b *BlocoC) writeRegistroC480(parent *RegistroC470) {
	for _, r := range parent.RegistroC480 {
		linha := b.LFillStr("C480", 0, false, '0') +
			b.ressarcimentoConv(r.CodMotRestCompl, r.QuantConv, r.Unid,
				r.VlUnitConv, r.VlUnitICMSNaOperacaoConv, r.VlUnitICMSOpConv,
				r.VlUnitICMSOpEstoqueConv, r.VlUnitICMSSTEstoqueConv,
				r.VlUnitFcpICMSSTEstoqueConv, r.VlUnitICMSSTConvRest,
				r.VlUnitFcpSTConvRest, r.VlUnitICMSSTConvCompl, r.VlUnitFcpSTConvCompl) +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC480Count++
	}
}

// writeRegistroC490 gera as linhas do registro C490.
func (b *BlocoC) writeRegistroC490(parent *RegistroC405) {
	for _, r := range parent.RegistroC490 {
		linha := b.LFillStr("C490", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC490Count++
	}
}

// writeRegistroC495 gera as linhas do registro C495.
func (b *BlocoC) writeRegistroC495() {
	if len(b.RegistroC001.RegistroC495) > 0 &&
		!b.DtIni.Before(time.Date(2014, 1, 1, 0, 0, 0, 0, time.UTC)) &&
		b.ufDoInformante() == "BA" {
		b.Checkf(false,
			"A partir de 01/01/2014, os contribuintes situados na Bahia obrigados a este registro devem apresentar o registro C425.")
	}
	for _, r := range b.RegistroC001.RegistroC495 {
		linha := b.LFillStr("C495", 0, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.DFill(r.QtdCanc, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlCanc, 2, false) +
			b.DFill(r.VlAcmo, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlIsen, 2, false) +
			b.DFill(r.VlNT, 2, false) +
			b.DFill(r.VlICMSST, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC495Count++
	}
}

// writeRegistroC500 gera as linhas do registro C500 (energia, agua, gas).
//
// O leiaute cresceu duas vezes: ate a versao 112 vai ate COD_GRUPO_TENSAO;
// das 113/114 vai ate COD_CTA; da 115 em diante inclui os campos de documento
// referenciado e de energia injetada.
func (b *BlocoC) writeRegistroC500() {
	ver := b.versaoC()
	for _, r := range b.RegistroC001.RegistroC500 {
		b.Checkf(r.CodMod == "06" || r.CodMod == "28" || r.CodMod == "29" || r.CodMod == "66",
			"Registro C500 : O codigo do modelo %q nao esta na lista de valores validos [06, 28, 29, 66]!", r.CodMod)

		fin := ""
		if r.CodMod == "66" {
			fin = r.FinDOCe.String()
		}
		indDest := int64(0)
		if r.IndOper != IndOperEntrada {
			indDest = int64(r.IndDest)
		}

		linha := b.LFillStr("C500", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.CodCons, 0, false, '0') +
			b.LFillStr(r.NumDoc, 9, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillDate(r.DtES, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlForn, 2, false) +
			b.DFill(r.VlServNT, 2, false) +
			b.DFill(r.VlTerc, 2, false) +
			b.DFill(r.VlDa, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.LFillInt(tpLigacaoIntC(r.TpLigacao), 0, true, '0') +
			b.LFillStr(r.CodGrupoTensao.String(), 0, false, '0')
		if ver >= VlVersao113 {
			linha += b.LFillStr(r.ChvDOCe, 0, false, '0') +
				b.LFillStr(fin, 0, false, '0') +
				b.LFillStr(r.ChvDOCeRef, 0, false, '0') +
				b.LFillInt(indDest, 0, true, '0') +
				b.LFillStr(r.CodMunDest, 0, false, '0') +
				b.LFillStr(r.CodCta, 0, false, '0')
		}
		if ver >= VlVersao115 {
			linha += b.LFillStr(r.CodModDocRef, 0, false, '0') +
				b.LFillStr(r.HashDocRef, 0, false, '0') +
				b.LFillStr(r.SerDocRef, 0, false, '0') +
				b.LFillStr(r.NumDocRef, 0, false, '0') +
				b.LFillStr(r.MesDocRef, 0, false, '0') +
				b.VLFill(r.EnerInjet, 0, 2) +
				b.VLFill(r.OutrasDed, 0, 2)
		}
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC500Count++

		b.writeRegistroC510(r)
		b.writeRegistroC590(r)
		b.writeRegistroC595(r)
	}
}

// tpLigacaoIntC devolve TP_LIGACAO como inteiro; 0 para valor nao mapeado.
func tpLigacaoIntC(v TpLigacao) int64 {
	switch v {
	case LigacaoMonofasico:
		return 1
	case LigacaoBifasico:
		return 2
	case LigacaoTrifasico:
		return 3
	}
	return 0
}

// writeRegistroC510 gera as linhas do registro C510.
func (b *BlocoC) writeRegistroC510(parent *RegistroC500) {
	for _, r := range parent.RegistroC510 {
		linha := b.LFillStr("C510", 0, false, '0') +
			b.LFillStr(r.NumItem, 3, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.CodClass, 4, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.AliqST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.LFillInt(int64(r.IndRec), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC510Count++
	}
}

// writeRegistroC590 gera as linhas do registro C590.
func (b *BlocoC) writeRegistroC590(parent *RegistroC500) {
	for _, r := range parent.RegistroC590 {
		linha := b.LFillStr("C590", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC590Count++

		b.writeRegistroC591(r)
	}
}

// writeRegistroC591 gera as linhas do registro C591.
func (b *BlocoC) writeRegistroC591(parent *RegistroC590) {
	for _, r := range parent.RegistroC591 {
		linha := b.LFillStr("C591", 0, false, '0') +
			b.DFill(r.VlFcpOp, 2, false) +
			b.DFill(r.VlFcpST, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC591Count++
	}
}

// writeRegistroC595 gera as linhas do registro C595.
func (b *BlocoC) writeRegistroC595(parent *RegistroC500) {
	for _, r := range parent.RegistroC595 {
		linha := b.LFillStr("C595", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC595Count++

		b.writeRegistroC597(r)
	}
}

// writeRegistroC597 gera as linhas do registro C597.
func (b *BlocoC) writeRegistroC597(parent *RegistroC595) {
	for _, r := range parent.RegistroC597 {
		linha := b.LFillStr("C597", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC597Count++
	}
}

// writeRegistroC600 gera as linhas do registro C600 (consolidacao diaria).
func (b *BlocoC) writeRegistroC600() {
	for _, r := range b.RegistroC001.RegistroC600 {
		linha := b.LFillStr("C600", 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.CodMun, 7, false, '0') +
			b.LFillStr(r.Ser, 4, false, '0') +
			b.LFillStr(r.Sub, 3, false, '0') +
			b.LFillStr(r.CodCons, 2, false, '0') +
			b.LFillInt(int64(r.QtdCons), 0, false, '0') +
			b.LFillInt(int64(r.QtdCanc), 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.LFillInt(int64(r.Cons), 0, false, '0') +
			b.DFill(r.VlForn, 2, false) +
			b.DFill(r.VlServNT, 2, false) +
			b.DFill(r.VlTerc, 2, false) +
			b.DFill(r.VlDa, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC600Count++

		b.writeRegistroC601(r)
		b.writeRegistroC610(r)
		b.writeRegistroC690(r)
	}
}

// writeRegistroC601 gera as linhas do registro C601.
func (b *BlocoC) writeRegistroC601(parent *RegistroC600) {
	for _, r := range parent.RegistroC601 {
		linha := b.LFillStr("C601", 0, false, '0') +
			b.LFillStr(r.NumDocCanc, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC601Count++
	}
}

// writeRegistroC610 gera as linhas do registro C610.
// QTD usa tamanho 3 com duas casas decimais, como no ACBr.
func (b *BlocoC) writeRegistroC610(parent *RegistroC600) {
	for _, r := range parent.RegistroC610 {
		linha := b.LFillStr("C610", 0, false, '0') +
			b.LFillStr(r.CodClass, 4, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillFloat(r.Qtd, 3, 2, false, '0', "") +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 0, 3, false, '0', "") +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC610Count++
	}
}

// writeRegistroC690 gera as linhas do registro C690.
//
// O ACBr emite VL_RED_BC duas vezes e nunca emite COD_OBS. O comportamento e
// reproduzido por fidelidade ao original.
func (b *BlocoC) writeRegistroC690(parent *RegistroC600) {
	for _, r := range parent.RegistroC690 {
		linha := b.LFillStr("C690", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC690Count++
	}
}

// writeRegistroC700 gera as linhas do registro C700.
func (b *BlocoC) writeRegistroC700() {
	for _, r := range b.RegistroC001.RegistroC700 {
		linha := b.LFillStr("C700", 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.Ser, 4, false, '0') +
			b.LFillInt(int64(r.NroOrdIni), 9, false, '0') +
			b.LFillInt(int64(r.NroOrdFin), 9, false, '0') +
			b.LFillDate(r.DtDocIni, "02012006", true) +
			b.LFillDate(r.DtDocFin, "02012006", true) +
			b.LFillStr(r.NomMest, 0, false, '0') +
			b.LFillStr(r.ChvCodDig, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC700Count++

		b.writeRegistroC790(r)
	}
}

// writeRegistroC790 gera as linhas do registro C790.
func (b *BlocoC) writeRegistroC790(parent *RegistroC700) {
	for _, r := range parent.RegistroC790 {
		linha := b.LFillStr("C790", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC790Count++

		b.writeRegistroC791(r)
	}
}

// writeRegistroC791 gera as linhas do registro C791.
func (b *BlocoC) writeRegistroC791(parent *RegistroC790) {
	for _, r := range parent.RegistroC791 {
		linha := b.LFillStr("C791", 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0') +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC791Count++
	}
}

// writeRegistroC800 gera as linhas do registro C800 (CF-e-SAT).
//
// Documento cancelado, denegado ou inutilizado zera PIS/COFINS e faz os demais
// valores aceitarem vazio.
func (b *BlocoC) writeRegistroC800() {
	for _, r := range b.RegistroC001.RegistroC800 {
		cancelada := r.CodSit == SitCancelado || r.CodSit == SitCanceladoExtemp ||
			r.CodSit == SitDenegado || r.CodSit == SitNumInutilizada

		pis, cofins, pisST, cofinsST := r.VlPIS, r.VlCOFINS, r.VlPISST, r.VlCOFINSST
		if cancelada {
			pis, cofins, pisST, cofinsST = nil, nil, nil, nil
		}

		linha := b.LFillStr("C800", 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.NumCFe, 6, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", cancelada) +
			b.DFill(r.VlCFe, 2, cancelada) +
			b.VLFill(pis, 0, 2) +
			b.VLFill(cofins, 0, 2) +
			b.LFillStr(r.CNPJCPF, 0, false, '0') +
			b.LFillStr(r.NrSat, 9, false, '0') +
			b.LFillStr(r.ChvCFe, 0, false, '0') +
			b.DFill(r.VlDesc, 2, cancelada) +
			b.DFill(r.VlMerc, 2, cancelada) +
			b.DFill(r.VlOutDa, 2, cancelada) +
			b.DFill(r.VlICMS, 2, cancelada) +
			b.VLFill(pisST, 0, 2) +
			b.VLFill(cofinsST, 0, 2)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC800Count++

		b.writeRegistroC810(r)
		b.writeRegistroC850(r)
		b.writeRegistroC855(r)
	}
}

// writeRegistroC810 gera as linhas do registro C810.
func (b *BlocoC) writeRegistroC810(parent *RegistroC800) {
	for _, r := range parent.RegistroC810 {
		linha := b.LFillStr("C810", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillFloat(r.Qtd, 0, 5, false, '0', "") +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC810Count++

		b.writeRegistroC815(r)
	}
}

// writeRegistroC815 gera a linha do registro C815 (unico por C810).
//
// Os cinco campos apos VL_UNIT_CONV usam VLFill com TAMANHO 6 e duas casas
// decimais, enquanto os quatro ultimos usam VDFill com SEIS casas. A assimetria
// e do ACBr e esta reproduzida aqui.
func (b *BlocoC) writeRegistroC815(parent *RegistroC810) {
	r := parent.RegistroC815
	if r == nil {
		return
	}
	linha := b.LFillStr("C815", 0, false, '0') +
		b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
		b.LFillFloat(r.QuantConv, 0, 6, false, '0', "") +
		b.LFillStr(r.Unid, 0, false, '0') +
		b.LFillFloat(r.VlUnitConv, 0, 6, false, '0', "") +
		b.VLFill(r.VlUnitICMSNaOperacaoConv, 6, 2) +
		b.VLFill(r.VlUnitICMSOpConv, 6, 2) +
		b.VLFill(r.VlUnitICMSOpEstoqueConv, 6, 2) +
		b.VLFill(r.VlUnitICMSSTEstoqueConv, 6, 2) +
		b.VLFill(r.VlUnitFcpICMSSTEstoqueConv, 6, 2) +
		b.VDFill(r.VlUnitICMSSTConvRest, 6) +
		b.VDFill(r.VlUnitFcpSTConvRest, 6) +
		b.VDFill(r.VlUnitICMSSTConvCompl, 6) +
		b.VDFill(r.VlUnitFcpSTConvCompl, 6)
	b.Add(linha, true)
	b.RegistroC990.QtdLinC++
	b.RegistroC815Count++
}

// writeRegistroC850 gera as linhas do registro C850.
func (b *BlocoC) writeRegistroC850(parent *RegistroC800) {
	for _, r := range parent.RegistroC850 {
		linha := b.LFillStr("C850", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC850Count++
	}
}

// writeRegistroC855 gera as linhas do registro C855.
// So existe a partir da versao 116.
func (b *BlocoC) writeRegistroC855(parent *RegistroC800) {
	if b.versaoC() < VlVersao116 {
		return
	}
	for _, r := range parent.RegistroC855 {
		linha := b.LFillStr("C855", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC855Count++

		b.writeRegistroC857(r)
	}
}

// writeRegistroC857 gera as linhas do registro C857.
func (b *BlocoC) writeRegistroC857(parent *RegistroC855) {
	for _, r := range parent.RegistroC857 {
		linha := b.LFillStr("C857", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC857Count++
	}
}

// writeRegistroC860 gera as linhas do registro C860 (equipamento SAT).
func (b *BlocoC) writeRegistroC860() {
	for _, r := range b.RegistroC001.RegistroC860 {
		linha := b.LFillStr("C860", 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.NrSat, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.DocIni, 0, false, '0') +
			b.LFillStr(r.DocFin, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC860Count++

		b.writeRegistroC870(r)
		b.writeRegistroC890(r)
		b.writeRegistroC895(r)
	}
}

// writeRegistroC870 gera as linhas do registro C870.
func (b *BlocoC) writeRegistroC870(parent *RegistroC860) {
	for _, r := range parent.RegistroC870 {
		linha := b.LFillStr("C870", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillFloat(r.Qtd, 0, 5, false, '0', "") +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC870Count++

		b.writeRegistroC880(r)
	}
}

// writeRegistroC880 gera a linha do registro C880 (unico por C870).
// Mesma assimetria VLFill/VDFill do C815.
func (b *BlocoC) writeRegistroC880(parent *RegistroC870) {
	r := parent.RegistroC880
	if r == nil {
		return
	}
	linha := b.LFillStr("C880", 0, false, '0') +
		b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
		b.LFillFloat(r.QuantConv, 0, 6, false, '0', "") +
		b.LFillStr(r.Unid, 0, false, '0') +
		b.LFillFloat(r.VlUnitConv, 0, 6, false, '0', "") +
		b.VLFill(r.VlUnitICMSNaOperacaoConv, 6, 2) +
		b.VLFill(r.VlUnitICMSOpConv, 6, 2) +
		b.VLFill(r.VlUnitICMSOpEstoqueConv, 6, 2) +
		b.VLFill(r.VlUnitICMSSTEstoqueConv, 6, 2) +
		b.VLFill(r.VlUnitFcpICMSSTEstoqueConv, 6, 2) +
		b.VDFill(r.VlUnitICMSSTConvRest, 6) +
		b.VDFill(r.VlUnitFcpSTConvRest, 6) +
		b.VDFill(r.VlUnitICMSSTConvCompl, 6) +
		b.VDFill(r.VlUnitFcpSTConvCompl, 6)
	b.Add(linha, true)
	b.RegistroC990.QtdLinC++
	b.RegistroC880Count++
}

// writeRegistroC890 gera as linhas do registro C890.
func (b *BlocoC) writeRegistroC890(parent *RegistroC860) {
	for _, r := range parent.RegistroC890 {
		linha := b.LFillStr("C890", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC890Count++
	}
}

// writeRegistroC895 gera as linhas do registro C895.
// So existe a partir da versao 116.
func (b *BlocoC) writeRegistroC895(parent *RegistroC860) {
	if b.versaoC() < VlVersao116 {
		return
	}
	for _, r := range parent.RegistroC895 {
		linha := b.LFillStr("C895", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC895Count++

		b.writeRegistroC897(r)
	}
}

// writeRegistroC897 gera as linhas do registro C897.
func (b *BlocoC) writeRegistroC897(parent *RegistroC895) {
	for _, r := range parent.RegistroC897 {
		linha := b.LFillStr("C897", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC897Count++
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
		b.writeRegistroC105(r)
		b.writeRegistroC110(r)
		b.writeRegistroC120(r)
		b.writeRegistroC130(r)
		b.writeRegistroC140(r)
		b.writeRegistroC160(r)
		b.writeRegistroC165(r)
		b.writeRegistroC170(r)
		b.writeRegistroC185(r)
		b.writeRegistroC186(r)
		b.writeRegistroC190(r)
		b.writeRegistroC195(r)
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
		b.writeRegistroC115(r)
		b.writeRegistroC116(r)
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

		b.writeRegistroC171(r)
		b.writeRegistroC172(r)
		b.writeRegistroC173(r)
		b.writeRegistroC174(r)
		b.writeRegistroC175(r)
		b.writeRegistroC176(r)
		// C177 so se aplica a Pernambuco.
		if b.ufDoInformante() == "PE" {
			b.writeRegistroC177(r)
		}
		b.writeRegistroC178(r)
		b.writeRegistroC179(r)
		b.writeRegistroC180(r)
		b.writeRegistroC181(r)
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

		b.writeRegistroC191(r)
	}
}

// writeRegistroC195 gera as linhas do registro C195.
// Formato: |C195|COD_OBS|TXT_COMPL|
func (b *BlocoC) writeRegistroC195(parent *RegistroC100) {
	for _, r := range parent.RegistroC195 {
		linha := b.LFillStr("C195", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC195Count++

		b.writeRegistroC197(r)
	}
}

// writeRegistroC197 gera as linhas do registro C197.
func (b *BlocoC) writeRegistroC197(parent *RegistroC195) {
	for _, r := range parent.RegistroC197 {
		linha := b.LFillStr("C197", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroC990.QtdLinC++
		b.RegistroC197Count++
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
		b.LFillInt(int64(b.RegistroD001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroD990.QtdLinD++

	if b.RegistroD001.IndMov == 0 {
		b.writeRegistroD100()
		b.writeRegistroD300()
		b.writeRegistroD350()
		b.writeRegistroD400()
		b.writeRegistroD500()
		b.writeRegistroD600()
		b.writeRegistroD695()
		b.writeRegistroD700()
		b.writeRegistroD750()
	}
}

// versaoD devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoD) versaoD() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// writeRegistroD101 gera as linhas do registro D101 (DIFAL do CT-e).
// Formato: |D101|VL_FCP_UF_DEST|VL_ICMS_UF_DEST|VL_ICMS_UF_REM|
func (b *BlocoD) writeRegistroD101(parent *RegistroD100) {
	for _, r := range parent.RegistroD101 {
		linha := b.LFillStr("D101", 0, false, '0') +
			b.DFill(r.VlFcpUFDest, 2, false) +
			b.DFill(r.VlICMSUFDest, 2, false) +
			b.DFill(r.VlICMSUFRem, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD101Count++
	}
}

// writeRegistroD110 gera as linhas do registro D110.
// Formato: |D110|NUN_ITEM|COD_ITEM|VL_SERV|VL_OUT|
func (b *BlocoD) writeRegistroD110(parent *RegistroD100) {
	for _, r := range parent.RegistroD110 {
		linha := b.LFillStr("D110", 0, false, '0') +
			b.LFillInt(int64(r.NunItem), 3, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlOut, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD110Count++

		b.writeRegistroD120(r)
	}
}

// writeRegistroD120 gera as linhas do registro D120.
// Formato: |D120|COD_MUN_ORIG|COD_MUN_DEST|VEIC_ID|UF_ID|
func (b *BlocoD) writeRegistroD120(parent *RegistroD110) {
	for _, r := range parent.RegistroD120 {
		linha := b.LFillStr("D120", 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.LFillStr(r.UFID, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD120Count++
	}
}

// writeRegistroD130 gera as linhas do registro D130 (conhecimento rodoviario).
func (b *BlocoD) writeRegistroD130(parent *RegistroD100) {
	for _, r := range parent.RegistroD130 {
		linha := b.LFillStr("D130", 0, false, '0') +
			b.LFillStr(r.CodPartConsg, 0, false, '0') +
			b.LFillStr(r.CodPartRed, 0, false, '0') +
			b.LFillStr(r.IndFrtRed.String(), 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.DFill(r.VlLiqFrt, 2, false) +
			b.DFill(r.VlSecCat, 2, false) +
			b.DFill(r.VlDesp, 2, false) +
			b.DFill(r.VlPedg, 2, false) +
			b.DFill(r.VlOut, 2, false) +
			b.DFill(r.VlFrt, 2, false) +
			b.LFillStr(r.UFID, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD130Count++
	}
}

// writeRegistroD140 gera as linhas do registro D140 (conhecimento aquaviario).
func (b *BlocoD) writeRegistroD140(parent *RegistroD100) {
	for _, r := range parent.RegistroD140 {
		linha := b.LFillStr("D140", 0, false, '0') +
			b.LFillStr(r.CodPartConsg, 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillInt(int64(r.IndVeic), 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.LFillInt(int64(r.IndNav), 0, false, '0') +
			b.LFillStr(r.Viagem, 0, false, '0') +
			b.DFill(r.VlFrtLiq, 2, false) +
			b.DFill(r.VlDespPort, 2, false) +
			b.DFill(r.VlDespCarDesc, 2, false) +
			b.DFill(r.VlOut, 2, false) +
			b.DFill(r.VlFrtBrt, 2, false) +
			b.DFill(r.VlFrtMM, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD140Count++
	}
}

// writeRegistroD150 gera as linhas do registro D150 (conhecimento aereo).
func (b *BlocoD) writeRegistroD150(parent *RegistroD100) {
	for _, r := range parent.RegistroD150 {
		linha := b.LFillStr("D150", 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillStr(r.VeicID, 0, false, '0') +
			b.LFillStr(r.Viagem, 0, false, '0') +
			b.LFillInt(int64(r.IndTFA), 0, false, '0') +
			b.DFill(r.VlPesoTx, 2, false) +
			b.DFill(r.VlTxTerr, 2, false) +
			b.DFill(r.VlTxRed, 2, false) +
			b.DFill(r.VlOut, 2, false) +
			b.DFill(r.VlTxAdv, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD150Count++
	}
}

// writeRegistroD160 gera as linhas do registro D160 (carga transportada).
func (b *BlocoD) writeRegistroD160(parent *RegistroD100) {
	for _, r := range parent.RegistroD160 {
		linha := b.LFillStr("D160", 0, false, '0') +
			b.LFillStr(r.Despacho, 0, false, '0') +
			b.LFillStr(r.CNPJCPFRem, 0, false, '0') +
			b.LFillStr(r.IERem, 0, false, '0') +
			b.LFillStr(r.CodMunOri, 0, false, '0') +
			b.LFillStr(r.CNPJCPFDest, 0, false, '0') +
			b.LFillStr(r.IEDest, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD160Count++

		b.writeRegistroD161(r)
		b.writeRegistroD162(r)
	}
}

// writeRegistroD161 gera as linhas do registro D161 (local de coleta e entrega).
func (b *BlocoD) writeRegistroD161(parent *RegistroD160) {
	for _, r := range parent.RegistroD161 {
		linha := b.LFillStr("D161", 0, false, '0') +
			b.LFillInt(int64(r.IndCarga), 0, false, '0') +
			b.LFillStr(r.CNPJCol, 0, false, '0') +
			b.LFillStr(r.IECol, 0, false, '0') +
			b.LFillStr(r.CodMunCol, 0, false, '0') +
			b.LFillStr(r.CNPJEntg, 0, false, '0') +
			b.LFillStr(r.IEEntg, 0, false, '0') +
			b.LFillStr(r.CodMunEntg, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD161Count++
	}
}

// writeRegistroD162 gera as linhas do registro D162 (documentos fiscais).
func (b *BlocoD) writeRegistroD162(parent *RegistroD160) {
	for _, r := range parent.RegistroD162 {
		linha := b.LFillStr("D162", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlMerc, 2, false) +
			b.LFillInt(int64(r.QtdVol), 0, false, '0') +
			b.DFill(r.PesoBrt, 2, false) +
			b.DFill(r.PesoLiq, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD162Count++
	}
}

// writeRegistroD170 gera as linhas do registro D170 (conhecimento multimodal).
//
// O ACBr emite VL_FRT duas vezes e nunca emite VEIC_ID. O comportamento e
// reproduzido por fidelidade ao original.
func (b *BlocoD) writeRegistroD170(parent *RegistroD100) {
	for _, r := range parent.RegistroD170 {
		linha := b.LFillStr("D170", 0, false, '0') +
			b.LFillStr(r.CodPartConsg, 0, false, '0') +
			b.LFillStr(r.CodPartRed, 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillStr(r.Otm, 0, false, '0') +
			b.LFillInt(int64(r.IndNatFrt), 0, false, '0') +
			b.DFill(r.VlLiqFrt, 2, false) +
			b.DFill(r.VlGris, 2, false) +
			b.DFill(r.VlPdg, 2, false) +
			b.DFill(r.VlOut, 2, false) +
			b.DFill(r.VlFrt, 2, false) +
			b.DFill(r.VlFrt, 2, false) +
			b.LFillStr(r.UFID, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD170Count++
	}
}

// writeRegistroD180 gera as linhas do registro D180 (modais).
func (b *BlocoD) writeRegistroD180(parent *RegistroD100) {
	for _, r := range parent.RegistroD180 {
		linha := b.LFillStr("D180", 0, false, '0') +
			b.LFillStr(r.NumSeq, 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CNPJEmit, 0, false, '0') +
			b.LFillStr(r.UFEmit, 0, false, '0') +
			b.LFillStr(r.IEEmit, 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.LFillStr(r.CNPJCPFTom, 0, false, '0') +
			b.LFillStr(r.UFTom, 0, false, '0') +
			b.LFillStr(r.IETom, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD180Count++
	}
}

// writeRegistroD300 gera as linhas do registro D300 (bilhetes consolidados).
func (b *BlocoD) writeRegistroD300() {
	for _, r := range b.RegistroD001.RegistroD300 {
		linha := b.LFillStr("D300", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDocIni, 0, false, '0') +
			b.LFillStr(r.NumDocFin, 0, false, '0') +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlSeg, 2, false) +
			b.DFill(r.VlOutDesp, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD300Count++

		b.writeRegistroD301(r)
		b.writeRegistroD310(r)
	}
}

// writeRegistroD301 gera as linhas do registro D301.
// Formato: |D301|NUM_DOC_CANC|
func (b *BlocoD) writeRegistroD301(parent *RegistroD300) {
	for _, r := range parent.RegistroD301 {
		linha := b.LFillStr("D301", 0, false, '0') +
			b.LFillStr(r.NumDocCanc, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD301Count++
	}
}

// writeRegistroD310 gera as linhas do registro D310.
// Formato: |D310|COD_MUN_ORIG|VL_SERV|VL_BC_ICMS|VL_ICMS|
func (b *BlocoD) writeRegistroD310(parent *RegistroD300) {
	for _, r := range parent.RegistroD310 {
		linha := b.LFillStr("D310", 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD310Count++
	}
}

// writeRegistroD350 gera as linhas do registro D350 (equipamento ECF).
// Formato: |D350|COD_MOD|ECF_MOD|ECF_FAB|ECF_CX|
func (b *BlocoD) writeRegistroD350() {
	for _, r := range b.RegistroD001.RegistroD350 {
		linha := b.LFillStr("D350", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.EcfMod, 0, false, '0') +
			b.LFillStr(r.EcfFab, 0, false, '0') +
			b.LFillStr(r.EcfCx, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD350Count++

		b.writeRegistroD355(r)
	}
}

// writeRegistroD355 gera as linhas do registro D355 (reducao Z).
// Formato: |D355|DT_DOC|CRO|CRZ|NUM_COO_FIN|GT_FIN|VL_BRT|
func (b *BlocoD) writeRegistroD355(parent *RegistroD350) {
	for _, r := range parent.RegistroD355 {
		linha := b.LFillStr("D355", 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillInt(int64(r.Cro), 3, false, '0') +
			b.LFillInt(int64(r.Crz), 6, false, '0') +
			b.LFillInt(int64(r.NumCooFin), 6, false, '0') +
			b.DFill(r.GtFin, 2, false) +
			b.DFill(r.VlBrt, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD355Count++

		b.writeRegistroD360(r)
		b.writeRegistroD365(r)
		b.writeRegistroD390(r)
	}
}

// writeRegistroD360 gera as linhas do registro D360.
// Formato: |D360|VL_PIS|VL_COFINS|
func (b *BlocoD) writeRegistroD360(parent *RegistroD355) {
	for _, r := range parent.RegistroD360 {
		linha := b.LFillStr("D360", 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD360Count++
	}
}

// writeRegistroD365 gera as linhas do registro D365.
// Formato: |D365|COD_TOT_PAR|VLR_ACUM_TOT|NR_TOT|DESCR_NR_TOT|
func (b *BlocoD) writeRegistroD365(parent *RegistroD355) {
	for _, r := range parent.RegistroD365 {
		linha := b.LFillStr("D365", 0, false, '0') +
			b.LFillStr(r.CodTotPar, 0, false, '0') +
			b.DFill(r.VlrAcumTot, 2, false) +
			b.LFillStr(r.NrTot, 0, false, '0') +
			b.LFillStr(r.DescrNrTot, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD365Count++

		b.writeRegistroD370(r)
	}
}

// writeRegistroD370 gera as linhas do registro D370.
// Formato: |D370|COD_MUN_ORIG|VL_SERV|QTD_BILH|VL_BC_ICMS|VL_ICMS|
func (b *BlocoD) writeRegistroD370(parent *RegistroD365) {
	for _, r := range parent.RegistroD370 {
		linha := b.LFillStr("D370", 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.DFill(r.VlServ, 2, false) +
			b.LFillInt(int64(r.QtdBilh), 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD370Count++
	}
}

// writeRegistroD390 gera as linhas do registro D390.
func (b *BlocoD) writeRegistroD390(parent *RegistroD355) {
	for _, r := range parent.RegistroD390 {
		linha := b.LFillStr("D390", 0, false, '0') +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcISSQN, 2, false) +
			b.DFill(r.AliqISSQN, 2, false) +
			b.DFill(r.VlISSQN, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD390Count++
	}
}

// writeRegistroD400 gera as linhas do registro D400 (resumo do movimento diario).
func (b *BlocoD) writeRegistroD400() {
	for _, r := range b.RegistroD001.RegistroD400 {
		linha := b.LFillStr("D400", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD400Count++

		b.writeRegistroD410(r)
		b.writeRegistroD420(r)
	}
}

// writeRegistroD410 gera as linhas do registro D410 (documentos informados).
func (b *BlocoD) writeRegistroD410(parent *RegistroD400) {
	for _, r := range parent.RegistroD410 {
		linha := b.LFillStr("D410", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDocIni, 0, false, '0') +
			b.LFillStr(r.NumDocFin, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CstICMS, 0, false, '0') +
			b.LFillStr(r.CFOP, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD410Count++

		b.writeRegistroD411(r)
	}
}

// writeRegistroD411 gera as linhas do registro D411.
// Formato: |D411|NUM_DOC_CANC|
func (b *BlocoD) writeRegistroD411(parent *RegistroD410) {
	for _, r := range parent.RegistroD411 {
		linha := b.LFillStr("D411", 0, false, '0') +
			b.LFillStr(r.NumDocCanc, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD411Count++
	}
}

// writeRegistroD420 gera as linhas do registro D420.
// Formato: |D420|COD_MUN_ORIG|VL_SERV|VL_BC_ICMS|VL_ICMS|
func (b *BlocoD) writeRegistroD420(parent *RegistroD400) {
	for _, r := range parent.RegistroD420 {
		linha := b.LFillStr("D420", 0, false, '0') +
			b.LFillStr(r.CodMunOrig, 0, false, '0') +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD420Count++
	}
}

// writeRegistroD500 gera as linhas do registro D500 (comunicacao/telecomunicacao).
//
// Documento cancelado, denegado ou inutilizado faz varios campos aceitarem
// vazio; VL_PIS, VL_COFINS e TP_ASSINANTE suprimem zero sempre.
func (b *BlocoD) writeRegistroD500() {
	for _, r := range b.RegistroD001.RegistroD500 {
		nulo := r.CodSit == SitCancelado || r.CodSit == SitCanceladoExtemp ||
			r.CodSit == SitDenegado || r.CodSit == SitNumInutilizada

		linha := b.LFillStr("D500", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, nulo, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, nulo, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillDate(r.DtAP, "02012006", true) +
			b.DFill(r.VlDoc, 2, nulo) +
			b.DFill(r.VlDesc, 2, nulo) +
			b.DFill(r.VlServ, 2, nulo) +
			b.DFill(r.VlServNT, 2, nulo) +
			b.DFill(r.VlTerc, 2, nulo) +
			b.DFill(r.VlDa, 2, nulo) +
			b.DFill(r.VlBcICMS, 2, nulo) +
			b.DFill(r.VlICMS, 2, nulo) +
			b.LFillStr(r.CodInf, 0, nulo, '0') +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.LFillStr(r.CodCta, 0, nulo, '0') +
			b.LFillInt(tpAssinanteInt(r.TpAssinante), 0, true, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD500Count++

		b.writeRegistroD510(r)
		b.writeRegistroD530(r)
		b.writeRegistroD590(r)
	}
}

// tpAssinanteInt devolve TP_ASSINANTE como inteiro; o ACBr usa 0 para qualquer
// valor nao mapeado, inclusive "nenhum".
func tpAssinanteInt(v TpAssinante) int64 {
	switch v {
	case AssComercialIndustrial:
		return 1
	case AssPoderPublico:
		return 2
	case AssResidencial:
		return 3
	case AssPublico:
		return 4
	case AssSemiPublico:
		return 5
	case AssOutros:
		return 6
	}
	return 0
}

// writeRegistroD510 gera as linhas do registro D510 (itens do D500).
func (b *BlocoD) writeRegistroD510(parent *RegistroD500) {
	for _, r := range parent.RegistroD510 {
		linha := b.LFillStr("D510", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.CodClass, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSUF, 2, false) +
			b.DFill(r.VlICMSUF, 2, false) +
			b.LFillInt(int64(r.IndRec), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD510Count++
	}
}

// writeRegistroD530 gera as linhas do registro D530 (terminal faturado).
func (b *BlocoD) writeRegistroD530(parent *RegistroD500) {
	for _, r := range parent.RegistroD530 {
		linha := b.LFillStr("D530", 0, false, '0') +
			b.LFillInt(int64(r.IndServ), 0, false, '0') +
			b.LFillDate(r.DtIniServ, "02012006", true) +
			b.LFillDate(r.DtFinServ, "02012006", true) +
			b.LFillStr(r.PerFiscal, 0, false, '0') +
			b.LFillStr(r.CodArea, 0, false, '0') +
			b.LFillStr(r.Terminal, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD530Count++
	}
}

// writeRegistroD590 gera as linhas do registro D590 (analitico do D500).
func (b *BlocoD) writeRegistroD590(parent *RegistroD500) {
	for _, r := range parent.RegistroD590 {
		linha := b.LFillStr("D590", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSUF, 2, false) +
			b.DFill(r.VlICMSUF, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD590Count++
	}
}

// writeRegistroD600 gera as linhas do registro D600 (consolidacao de servicos).
func (b *BlocoD) writeRegistroD600() {
	for _, r := range b.RegistroD001.RegistroD600 {
		linha := b.LFillStr("D600", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodMun, 7, false, '0') +
			b.LFillStr(r.Ser, 4, false, '0') +
			b.LFillInt(int64(r.Sub), 3, false, '0') +
			b.LFillInt(int64(r.CodCons), 2, false, '0') +
			b.LFillInt(int64(r.QtdCons), 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlServNT, 2, false) +
			b.DFill(r.VlTerc, 2, false) +
			b.DFill(r.VlDa, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD600Count++

		b.writeRegistroD610(r)
		b.writeRegistroD690(r)
	}
}

// writeRegistroD610 gera as linhas do registro D610 (itens do D600).
func (b *BlocoD) writeRegistroD610(parent *RegistroD600) {
	for _, r := range parent.RegistroD610 {
		linha := b.LFillStr("D610", 0, false, '0') +
			b.LFillInt(int64(r.CodClass), 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD610Count++
	}
}

// writeRegistroD690 gera as linhas do registro D690 (analitico do D600).
func (b *BlocoD) writeRegistroD690(parent *RegistroD600) {
	for _, r := range parent.RegistroD690 {
		linha := b.LFillStr("D690", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSUF, 2, false) +
			b.DFill(r.VlICMSUF, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD690Count++
	}
}

// writeRegistroD695 gera as linhas do registro D695 (consolidacao de documentos).
func (b *BlocoD) writeRegistroD695() {
	for _, r := range b.RegistroD001.RegistroD695 {
		linha := b.LFillStr("D695", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillInt(int64(r.NroOrdIni), 9, false, '0') +
			b.LFillInt(int64(r.NroOrdFin), 9, false, '0') +
			b.LFillDate(r.DtDocIni, "02012006", true) +
			b.LFillDate(r.DtDocFin, "02012006", true) +
			b.LFillStr(r.NomMest, 0, false, '0') +
			b.LFillStr(r.ChvCodDig, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD695Count++

		b.writeRegistroD696(r)
	}
}

// writeRegistroD696 gera as linhas do registro D696.
func (b *BlocoD) writeRegistroD696(parent *RegistroD695) {
	for _, r := range parent.RegistroD696 {
		linha := b.LFillStr("D696", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSUF, 2, false) +
			b.DFill(r.VlICMSUF, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD696Count++

		b.writeRegistroD697(r)
	}
}

// writeRegistroD697 gera as linhas do registro D697.
// Formato: |D697|UF|VL_BC_ICMS|VL_ICMS|
func (b *BlocoD) writeRegistroD697(parent *RegistroD696) {
	for _, r := range parent.RegistroD697 {
		linha := b.LFillStr("D697", 0, false, '0') +
			b.LFillStr(r.UF, 2, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD697Count++
	}
}

// writeRegistroD700 gera as linhas do registro D700 (nota fatura de comunicacao).
//
// So existe a partir da versao 116. A partir da 118, FIN_DOCe e TIP_FAT ganham
// tamanho 1 com zero a esquerda, e o campo DED entra no fim.
func (b *BlocoD) writeRegistroD700() {
	if b.versaoD() < VlVersao116 {
		return
	}
	for _, r := range b.RegistroD001.RegistroD700 {
		b.Checkf(r.CodMod == "62",
			"Registro D700: O codigo do modelo %q nao esta na lista de valores validos [62]!", r.CodMod)
		cod := r.CodSit.String()
		b.Checkf(cod == "00" || cod == "08",
			"Registro D700: O codigo da situacao do documento fiscal %q nao esta na lista de valores validos [00, 08]!", cod)

		linha := b.LFillStr("D700", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(cod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 9, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillDate(r.DtES, "02012006", true) +
			b.DFill(r.VlDoc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlServ, 2, false) +
			b.DFill(r.VlServNT, 2, false) +
			b.DFill(r.VlTerc, 2, false) +
			b.DFill(r.VlDa, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.DFill(r.VlPIS, 2, true) +
			b.DFill(r.VlCOFINS, 2, true) +
			b.LFillStr(r.ChvDOCe, 0, false, '0')
		if b.versaoD() < VlVersao118 {
			linha += b.LFillStr(r.FinDOCe.String(), 0, false, '0') +
				b.LFillStr(r.TipFat.String(), 0, false, '0')
		} else {
			linha += b.LFillStr(r.FinDOCe.String(), 1, false, '0') +
				b.LFillStr(r.TipFat.String(), 1, false, '0')
		}
		linha += b.LFillStr(r.CodModDocRef, 0, false, '0') +
			b.LFillStr(r.ChvDOCeRef, 0, false, '0') +
			b.LFillStr(r.HashDocRef, 0, false, '0') +
			b.LFillStr(r.SerDocRef, 0, false, '0') +
			b.LFillStr(r.NumDocRef, 0, false, '0') +
			b.LFillStr(r.MesDocRef, 0, false, '0') +
			b.LFillStr(r.CodMunDest, 0, false, '0')
		if b.versaoD() >= VlVersao118 {
			linha += b.DFill(r.Ded, 2, true)
		}
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD700Count++

		b.writeRegistroD730(r)
		b.writeRegistroD735(r)
	}
}

// writeRegistroD730 gera as linhas do registro D730.
func (b *BlocoD) writeRegistroD730(parent *RegistroD700) {
	for _, r := range parent.RegistroD730 {
		linha := b.LFillStr("D730", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD730Count++

		b.writeRegistroD731(r)
	}
}

// writeRegistroD731 gera as linhas do registro D731.
// Formato: |D731|VL_FCP_OP|
func (b *BlocoD) writeRegistroD731(parent *RegistroD730) {
	for _, r := range parent.RegistroD731 {
		linha := b.LFillStr("D731", 0, false, '0') +
			b.DFill(r.VlFcpOp, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD731Count++
	}
}

// writeRegistroD735 gera as linhas do registro D735.
// Formato: |D735|COD_OBS|TXT_COMPL|
func (b *BlocoD) writeRegistroD735(parent *RegistroD700) {
	for _, r := range parent.RegistroD735 {
		linha := b.LFillStr("D735", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD735Count++

		b.writeRegistroD737(r)
	}
}

// writeRegistroD737 gera as linhas do registro D737.
func (b *BlocoD) writeRegistroD737(parent *RegistroD735) {
	for _, r := range parent.RegistroD737 {
		linha := b.LFillStr("D737", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD737Count++
	}
}

// writeRegistroD750 gera as linhas do registro D750.
//
// So existe a partir da versao 116. O ACBr monta a linha mas nunca a grava --
// apenas incrementa a contagem e escreve os filhos. O comportamento e
// reproduzido por fidelidade ao original.
func (b *BlocoD) writeRegistroD750() {
	if b.versaoD() < VlVersao116 {
		return
	}
	for _, r := range b.RegistroD001.RegistroD750 {
		// A linha e montada e descartada, como no ACBr.
		b.RegistroD990.QtdLinD++
		b.RegistroD750Count++

		b.writeRegistroD760(r)
	}
}

// writeRegistroD760 gera as linhas do registro D760.
func (b *BlocoD) writeRegistroD760(parent *RegistroD750) {
	for _, r := range parent.RegistroD760 {
		linha := b.LFillStr("D760", 0, false, '0') +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlOpr, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlRedBC, 2, false) +
			b.LFillStr(r.CodObs, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD760Count++

		b.writeRegistroD761(r)
	}
}

// writeRegistroD761 gera as linhas do registro D761.
// Formato: |D761|VL_FCP_OP|
func (b *BlocoD) writeRegistroD761(parent *RegistroD760) {
	for _, r := range parent.RegistroD761 {
		linha := b.LFillStr("D761", 0, false, '0') +
			b.DFill(r.VlFcpOp, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD761Count++
	}
}

func (b *BlocoD) writeRegistroD100() {
	for _, r := range b.RegistroD001.RegistroD100 {
		// Cancelado (02/03), denegado (04) ou inutilizado (05): os valores saem
		// vazios em vez de 0,00 (booConsiderarComoValorNulo no ACBr).
		cancelado := r.CodSit == SitCancelado || r.CodSit == SitCanceladoExtemp ||
			r.CodSit == SitDenegado || r.CodSit == SitNumInutilizada

		// So o inutilizado zera a chave do CT-e
		// (booConsiderarComoValorNuloParaInutilizado no ACBr).
		chvCTe := r.ChvCTe
		if r.CodSit == SitNumInutilizada {
			chvCTe = ""
		}

		linha := b.LFillStr("D100", 0, false, '0') +
			b.LFillInt(int64(r.IndOper), 0, false, '0') +
			b.LFillInt(int64(r.IndEmit), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(chvCTe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", false) +
			b.LFillDate(r.DtAP, "02012006", true) +
			b.LFillStr(r.TpCTe, 0, false, '0') +
			b.LFillStr(r.ChvCTeRef, 0, false, '0') +
			b.DFill(r.VlDoc, 2, cancelado) +
			b.DFill(r.VlDesc, 2, cancelado) +
			b.LFillStr(r.IndFrt.StringEmD100(b.DtIni), 0, false, '0') +
			b.DFill(r.VlServ, 2, cancelado) +
			b.DFill(r.VlBcICMS, 2, cancelado) +
			b.DFill(r.VlICMS, 2, cancelado) +
			b.DFill(r.VlNT, 2, cancelado) +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0') +
			municipiosD100(b, r)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD100Count++

		b.writeRegistroD101(r)
		b.writeRegistroD110(r)
		b.writeRegistroD130(r)
		b.writeRegistroD140(r)
		b.writeRegistroD150(r)
		b.writeRegistroD160(r)
		b.writeRegistroD170(r)
		b.writeRegistroD180(r)
		b.writeRegistroD190(r)
		b.writeRegistroD195(r)
	}
}

// writeRegistroD195 gera as linhas do registro D195 (observacoes do lancamento).
// Formato: |D195|COD_OBS|TXT_COMPL|
func (b *BlocoD) writeRegistroD195(parent *RegistroD100) {
	for _, r := range parent.RegistroD195 {
		linha := b.LFillStr("D195", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD195Count++

		b.writeRegistroD197(r)
	}
}

// writeRegistroD197 gera as linhas do registro D197.
func (b *BlocoD) writeRegistroD197(parent *RegistroD195) {
	for _, r := range parent.RegistroD197 {
		linha := b.LFillStr("D197", 0, false, '0') +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.LFillFloat(r.AliqICMS, 6, 2, false, '0', "") +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlOutros, 2, false)
		b.Add(linha, true)
		b.RegistroD990.QtdLinD++
		b.RegistroD197Count++
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
		b.LFillInt(int64(b.RegistroE001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroE990.QtdLinE++

	if b.RegistroE001.IndMov == 0 {
		b.writeRegistroE100()
		b.writeRegistroE200()
		// O ramo do DIFAL so existe a partir da versao 109 do leiaute.
		if b.versaoE() >= VlVersao109 {
			b.writeRegistroE300()
		}
		b.writeRegistroE500()
	}
}

// versaoE devolve a versao do leiaute declarada no registro 0000.
func (b *BlocoE) versaoE() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// indProcInt devolve IND_PROC como inteiro, no formato usado pelos registros
// E230, E312 e E530: o ACBr cai em 9 para qualquer valor nao mapeado,
// inclusive "nenhum" -- diferente da forma string usada em E112/E116/E250/E316,
// que devolve vazio.
func indProcInt(v OrigemProcesso) int64 {
	switch v {
	case OrigProcSefaz:
		return 0
	case OrigProcJusticaFederal:
		return 1
	case OrigProcJusticaEstadual:
		return 2
	case OrigProcSecexRFB:
		return 3
	}
	return 9
}

// writeRegistroE112 gera as linhas do registro E112.
// Formato: |E112|NUM_DA|NUM_PROC|IND_PROC|PROC|TXT_COMPL|
func (b *BlocoE) writeRegistroE112(parent *RegistroE111) {
	for _, r := range parent.RegistroE112 {
		linha := b.LFillStr("E112", 0, false, '0') +
			b.LFillStr(r.NumDA, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE112Count++
	}
}

// writeRegistroE113 gera as linhas do registro E113.
// Formato ate a versao 102: |E113|COD_PART|COD_MOD|SER|SUB|NUM_DOC|DT_DOC|CHV_NFE|COD_ITEM|VL_AJ_ITEM|
// Das demais versoes: CHV_NFE vai para o fim e so sai a partir de 2017.
func (b *BlocoE) writeRegistroE113(parent *RegistroE111) {
	for _, r := range parent.RegistroE113 {
		linha := b.LFillStr("E113", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true)
		if b.versaoE() == VlVersao102 {
			linha += b.LFillStr(r.ChvNFe, 0, false, '0') +
				b.LFillStr(r.CodItem, 0, false, '0') +
				b.DFill(r.VlAjItem, 2, false)
		} else {
			linha += b.LFillStr(r.CodItem, 0, false, '0') +
				b.DFill(r.VlAjItem, 2, false)
			if !b.DtIni.Before(time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)) {
				linha += b.LFillStr(r.ChvNFe, 0, false, '0')
			}
		}
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE113Count++
	}
}

// writeRegistroE115 gera as linhas do registro E115.
// Formato: |E115|COD_INF_ADIC|VL_INF_ADIC|DESCR_COMPL_AJ|
func (b *BlocoE) writeRegistroE115(parent *RegistroE110) {
	for _, r := range parent.RegistroE115 {
		linha := b.LFillStr("E115", 0, false, '0') +
			b.LFillStr(r.CodInfAdic, 0, false, '0') +
			b.DFill(r.VlInfAdic, 2, false) +
			b.LFillStr(r.DescrComplAj, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE115Count++
	}
}

// writeRegistroE200 gera as linhas do registro E200 (apuracao do ICMS ST).
// Formato: |E200|UF|DT_INI|DT_FIN|
func (b *BlocoE) writeRegistroE200() {
	for _, r := range b.RegistroE001.RegistroE200 {
		linha := b.LFillStr("E200", 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE200Count++

		b.writeRegistroE210(r)
	}
}

// writeRegistroE210 gera as linhas do registro E210.
// Formato: |E210|IND_MOV_ST|VL_SLD_CRED_ANT_ST|VL_DEVOL_ST|VL_RESSARC_ST|
//
//	VL_OUT_CRED_ST|VL_AJ_CREDITOS_ST|VL_RETENCAO_ST|VL_OUT_DEB_ST|
//	VL_AJ_DEBITOS_ST|VL_SLD_DEV_ANT_ST|VL_DEDUCOES_ST|VL_ICMS_RECOL_ST|
//	VL_SLD_CRED_ST_TRANSPORTAR|DEB_ESP_ST|
func (b *BlocoE) writeRegistroE210(parent *RegistroE200) {
	for _, r := range parent.RegistroE210 {
		linha := b.LFillStr("E210", 0, false, '0') +
			b.LFillInt(int64(r.IndMovST), 0, false, '0') +
			b.DFill(r.VlSldCredAntST, 2, false) +
			b.DFill(r.VlDevolST, 2, false) +
			b.DFill(r.VlRessarcST, 2, false) +
			b.DFill(r.VlOutCredST, 2, false) +
			b.DFill(r.VlAjCreditosST, 2, false) +
			b.DFill(r.VlRetencaoST, 2, false) +
			b.DFill(r.VlOutDebST, 2, false) +
			b.DFill(r.VlAjDebitosST, 2, false) +
			b.DFill(r.VlSldDevAntST, 2, false) +
			b.DFill(r.VlDeducoesST, 2, false) +
			b.DFill(r.VlICMSRecolST, 2, false) +
			b.DFill(r.VlSldCredSTTransportar, 2, false) +
			b.DFill(r.DebEspST, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE210Count++

		b.writeRegistroE220(r)
		b.writeRegistroE250(r)
	}
}

// writeRegistroE220 gera as linhas do registro E220.
// Formato: |E220|COD_AJ_APUR|DESCR_COMPL_AJ|VL_AJ_APUR|
func (b *BlocoE) writeRegistroE220(parent *RegistroE210) {
	for _, r := range parent.RegistroE220 {
		linha := b.LFillStr("E220", 0, false, '0') +
			b.LFillStr(r.CodAjApur, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.DFill(r.VlAjApur, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE220Count++

		b.writeRegistroE230(r)
		b.writeRegistroE240(r)
	}
}

// writeRegistroE230 gera as linhas do registro E230.
// Formato: |E230|NUM_DA|NUM_PROC|IND_PROC|PROC|TXT_COMPL|
//
// IND_PROC sai como inteiro, com 9 para qualquer valor nao mapeado.
func (b *BlocoE) writeRegistroE230(parent *RegistroE220) {
	for _, r := range parent.RegistroE230 {
		linha := b.LFillStr("E230", 0, false, '0') +
			b.LFillStr(r.NumDA, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillInt(indProcInt(r.IndProc), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE230Count++
	}
}

// writeRegistroE240 gera as linhas do registro E240.
// Formato: |E240|COD_PART|COD_MOD|SER|SUB|NUM_DOC|DT_DOC|COD_ITEM|VL_AJ_ITEM|CHV_NFE|
//
// CHV_NFE so entra a partir da versao 104.
func (b *BlocoE) writeRegistroE240(parent *RegistroE220) {
	for _, r := range parent.RegistroE240 {
		linha := b.LFillStr("E240", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlAjItem, 2, false)
		if b.versaoE() > VlVersao103 {
			linha += b.LFillStr(r.ChvNFe, 0, false, '0')
		}
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE240Count++
	}
}

// writeRegistroE250 gera as linhas do registro E250.
// Formato: |E250|COD_OR|VL_OR|DT_VCTO|COD_REC|NUM_PROC|IND_PROC|PROC|TXT_COMPL|MES_REF|
//
// MES_REF so entra a partir da versao 103. Nas versoes 100 e 101 o ACBr nao
// emite linha alguma, mas ainda assim conta o registro -- comportamento
// reproduzido aqui por fidelidade.
func (b *BlocoE) writeRegistroE250(parent *RegistroE210) {
	ver := b.versaoE()
	for _, r := range parent.RegistroE250 {
		if ver == VlVersao102 || ver >= VlVersao103 {
			linha := b.LFillStr("E250", 0, false, '0') +
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
		}
		b.RegistroE990.QtdLinE++
		b.RegistroE250Count++
	}
}

// writeRegistroE300 gera as linhas do registro E300 (apuracao do DIFAL).
// Formato: |E300|UF|DT_INI|DT_FIN|
func (b *BlocoE) writeRegistroE300() {
	for _, r := range b.RegistroE001.RegistroE300 {
		linha := b.LFillStr("E300", 0, false, '0') +
			b.LFillStr(r.UF, 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE300Count++

		b.writeRegistroE310(r)
	}
}

// writeRegistroE310 gera as linhas do registro E310.
//
// Ate 31/12/2016 o registro tem 14 campos, com VL_TOT_DEB_FCP e VL_TOT_CRED_FCP
// intercalados e os campos legados VL_RECOL e VL_SLD_CRED_TRANSPORTAR. A partir
// de 01/01/2017 passa a 22 campos, com o bloco do FCP separado.
// Ref.: ACBrEFDBloco_E_Class.pas, WriteRegistroE310.
func (b *BlocoE) writeRegistroE310(parent *RegistroE300) {
	antigo := !b.DtFin.After(time.Date(2016, 12, 31, 0, 0, 0, 0, time.UTC))
	for _, r := range parent.RegistroE310 {
		linha := b.LFillStr("E310", 0, false, '0') +
			b.LFillInt(int64(r.IndMovDIFAL), 0, false, '0') +
			b.DFill(r.VlSldCredAntDIFAL, 2, false) +
			b.DFill(r.VlTotDebitosDIFAL, 2, false) +
			b.DFill(r.VlOutDebDIFAL, 2, false)
		if antigo {
			linha += b.DFill(r.VlTotDebFCP, 2, false) +
				b.DFill(r.VlTotCreditosDIFAL, 2, false) +
				b.DFill(r.VlTotCredFCP, 2, false) +
				b.DFill(r.VlOutCredDIFAL, 2, false) +
				b.DFill(r.VlSldDevAntDIFAL, 2, false) +
				b.DFill(r.VlDeducoesDIFAL, 2, false) +
				b.DFill(r.VlRecol, 2, false) +
				b.DFill(r.VlSldCredTransportar, 2, false) +
				b.DFill(r.DebEspDIFAL, 2, false)
		} else {
			linha += b.DFill(r.VlTotCreditosDIFAL, 2, false) +
				b.DFill(r.VlOutCredDIFAL, 2, false) +
				b.DFill(r.VlSldDevAntDIFAL, 2, false) +
				b.DFill(r.VlDeducoesDIFAL, 2, false) +
				b.DFill(r.VlRecolDIFAL, 2, false) +
				b.DFill(r.VlSldCredTranspDIFAL, 2, false) +
				b.DFill(r.DebEspDIFAL, 2, false) +
				b.DFill(r.VlSldCredAntFCP, 2, false) +
				b.DFill(r.VlTotDebFCP, 2, false) +
				b.DFill(r.VlOutDebFCP, 2, false) +
				b.DFill(r.VlTotCredFCP, 2, false) +
				b.DFill(r.VlOutCredFCP, 2, false) +
				b.DFill(r.VlSldDevAntFCP, 2, false) +
				b.DFill(r.VlDeducoesFCP, 2, false) +
				b.DFill(r.VlRecolFCP, 2, false) +
				b.DFill(r.VlSldCredTranspFCP, 2, false) +
				b.DFill(r.DebEspFCP, 2, false)
		}
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE310Count++

		b.writeRegistroE311(r)
		b.writeRegistroE316(r)
	}
}

// writeRegistroE311 gera as linhas do registro E311.
// Formato: |E311|COD_AJ_APUR|DESCR_COMPL_AJ|VL_AJ_APUR|
func (b *BlocoE) writeRegistroE311(parent *RegistroE310) {
	for _, r := range parent.RegistroE311 {
		linha := b.LFillStr("E311", 0, false, '0') +
			b.LFillStr(r.CodAjApur, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.DFill(r.VlAjApur, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE311Count++

		b.writeRegistroE312(r)
		b.writeRegistroE313(r)
	}
}

// writeRegistroE312 gera as linhas do registro E312.
// Formato: |E312|NUM_DA|NUM_PROC|IND_PROC|PROC|TXT_COMPL|
func (b *BlocoE) writeRegistroE312(parent *RegistroE311) {
	for _, r := range parent.RegistroE312 {
		linha := b.LFillStr("E312", 0, false, '0') +
			b.LFillStr(r.NumDA, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillInt(indProcInt(r.IndProc), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE312Count++
	}
}

// writeRegistroE313 gera as linhas do registro E313.
// Formato: |E313|COD_PART|COD_MOD|SER|SUB|NUM_DOC|CHV_DOCe|DT_DOC|COD_ITEM|VL_AJ_ITEM|
//
// Diferente do E113 e do E240, aqui a chave vem ANTES da data.
func (b *BlocoE) writeRegistroE313(parent *RegistroE311) {
	for _, r := range parent.RegistroE313 {
		linha := b.LFillStr("E313", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvDOCe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlAjItem, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE313Count++
	}
}

// writeRegistroE316 gera as linhas do registro E316.
// Formato: |E316|COD_OR|VL_OR|DT_VCTO|COD_REC|NUM_PROC|IND_PROC|PROC|TXT_COMPL|MES_REF|
func (b *BlocoE) writeRegistroE316(parent *RegistroE310) {
	for _, r := range parent.RegistroE316 {
		linha := b.LFillStr("E316", 0, false, '0') +
			b.LFillStr(r.CodOR, 0, false, '0') +
			b.DFill(r.VlOR, 2, false) +
			b.LFillDate(r.DtVcto, "02012006", true) +
			b.LFillStr(r.CodRec, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc.String(), 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0') +
			b.LFillStr(r.MesRef, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE316Count++
	}
}

// writeRegistroE500 gera as linhas do registro E500 (apuracao do IPI).
// Formato: |E500|IND_APUR|DT_INI|DT_FIN|
//
// IND_APUR sai como ordinal do enum, nao pelo String().
func (b *BlocoE) writeRegistroE500() {
	for _, r := range b.RegistroE001.RegistroE500 {
		ordinal := int64(r.IndApur)
		if r.IndApur == ApuracaoIPINenhum {
			ordinal = 2 // iaNenhum e o terceiro membro do enum Delphi
		}
		linha := b.LFillStr("E500", 0, false, '0') +
			b.LFillInt(ordinal, 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE500Count++

		b.writeRegistroE510(r)
		b.writeRegistroE520(r)
	}
}

// writeRegistroE510 gera as linhas do registro E510.
// Formato: |E510|CFOP|CST_IPI|VL_CONT_IPI|VL_BC_IPI|VL_IPI|
func (b *BlocoE) writeRegistroE510(parent *RegistroE500) {
	for _, r := range parent.RegistroE510 {
		linha := b.LFillStr("E510", 0, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.LFillStr(r.CstIPI, 0, false, '0') +
			b.DFill(r.VlContIPI, 2, false) +
			b.DFill(r.VlBcIPI, 2, false) +
			b.DFill(r.VlIPI, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE510Count++
	}
}

// writeRegistroE520 gera as linhas do registro E520.
// Formato: |E520|VL_SD_ANT_IPI|VL_DEB_IPI|VL_CRED_IPI|VL_OD_IPI|VL_OC_IPI|VL_SC_IPI|VL_SD_IPI|
func (b *BlocoE) writeRegistroE520(parent *RegistroE500) {
	for _, r := range parent.RegistroE520 {
		linha := b.LFillStr("E520", 0, false, '0') +
			b.DFill(r.VlSdAntIPI, 2, false) +
			b.DFill(r.VlDebIPI, 2, false) +
			b.DFill(r.VlCredIPI, 2, false) +
			b.DFill(r.VlOdIPI, 2, false) +
			b.DFill(r.VlOcIPI, 2, false) +
			b.DFill(r.VlScIPI, 2, false) +
			b.DFill(r.VlSdIPI, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE520Count++

		b.writeRegistroE530(r)
	}
}

// writeRegistroE530 gera as linhas do registro E530.
// Formato: |E530|IND_AJ|VL_AJ|COD_AJ|IND_DOC|NUM_DOC|DESCR_AJ|
func (b *BlocoE) writeRegistroE530(parent *RegistroE520) {
	for _, r := range parent.RegistroE530 {
		linha := b.LFillStr("E530", 0, false, '0') +
			b.LFillInt(int64(r.IndAj), 0, false, '0') +
			b.DFill(r.VlAj, 2, false) +
			b.LFillStr(r.CodAj, 0, false, '0') +
			b.LFillInt(indDocInt(r.IndDoc), 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.DescrAj, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE530Count++

		b.writeRegistroE531(r)
	}
}

// indDocInt devolve IND_DOC do E530 como inteiro, com 9 para valor nao mapeado.
func indDocInt(v OrigemDocto) int64 {
	switch v {
	case OrigDocProcessoJudicial:
		return 0
	case OrigDocProcessoAdminist:
		return 1
	case OrigDocPerDcomp:
		return 2
	case OrigDocDocumentoFiscal:
		return 3
	}
	return 9
}

// writeRegistroE531 gera as linhas do registro E531.
// Formato: |E531|COD_PART|COD_MOD|SER|SUB|NUM_DOC|DT_DOC|COD_ITEM|VL_AJ_ITEM|CHV_NFE|
func (b *BlocoE) writeRegistroE531(parent *RegistroE530) {
	for _, r := range parent.RegistroE531 {
		linha := b.LFillStr("E531", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlAjItem, 2, false) +
			b.LFillStr(r.ChvNFe, 0, false, '0')
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE531Count++
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
			b.DFill(r.VlSldCredorTransportar, 2, false) +
			b.DFill(r.DebEsp, 2, false)
		b.Add(linha, true)
		b.RegistroE990.QtdLinE++
		b.RegistroE110Count++

		b.writeRegistroE111(r)
		b.writeRegistroE115(r)
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

		b.writeRegistroE112(r)
		b.writeRegistroE113(r)
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
		b.LFillInt(int64(b.RegistroG001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroG990.QtdLinG++

	if b.RegistroG001.IndMov == 0 {
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
	ver := b.versaoG()
	for _, r := range b.RegistroG001.RegistroG110 {
		// Nas versoes 100 e 101 o ACBr nao cobre nenhum dos dois ramos e a
		// linha simplesmente nao sai -- mas o contador incrementa do mesmo
		// jeito, porque no Delphi ele fica fora do if. Mesmo caso do E250.
		if ver == VlVersao102 || ver >= VlVersao103 {
			linha := b.LFillStr("G110", 0, false, '0') +
				b.LFillDate(r.DtIni, "02012006", true) +
				b.LFillDate(r.DtFin, "02012006", true)
			if ver == VlVersao102 {
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
		}
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
			linha += b.DFill(r.VlParcAprop, 2, false)
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
			b.DFill(r.VlParcAprop, 2, false)
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
		b.LFillInt(int64(b.RegistroH001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroH990.QtdLinH++

	if b.RegistroH001.IndMov == 0 {
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
		b.LFillInt(int64(b.RegistroK001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.RegistroK990.QtdLinK++

	if b.RegistroK001.IndMov == 0 {
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
		b.LFillInt(int64(b.Registro1001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.Registro1990.QtdLin1++

	if b.Registro1001.IndMov == 0 {
		b.writeRegistro1010()
		b.writeRegistro1100()
		b.writeRegistro1200()
		b.writeRegistro1250()
		b.writeRegistro1300()
		b.writeRegistro1350()
		b.writeRegistro1390()
		b.writeRegistro1400()
		b.writeRegistro1500()
		b.writeRegistro1600()
		b.writeRegistro1601()
		b.writeRegistro1700()
		b.writeRegistro1800()
		b.writeRegistro1900()
		// GIAF so se aplica a Pernambuco, a partir da versao 112.
		if b.giafAplicavel() {
			b.writeRegistro1960()
			b.writeRegistro1970()
			b.writeRegistro1980()
		}
	}
}

// versao1 devolve a versao do leiaute declarada no registro 0000.
func (b *Bloco1) versao1() VersaoLeiauteFiscal {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return VlVersao100
	}
	return b.Bloco0.Registro0000.CodVer
}

// giafAplicavel informa se os registros 1960/1970/1980 (GIAF) devem sair:
// so a partir da versao 112 e apenas para estabelecimento em Pernambuco.
func (b *Bloco1) giafAplicavel() bool {
	if b.Bloco0 == nil || b.Bloco0.Registro0000 == nil {
		return false
	}
	return b.versao1() >= VlVersao112 && b.Bloco0.Registro0000.UF == "PE"
}

// writeRegistro1100 gera as linhas do registro 1100 (informacoes sobre exportacao).
// Formato: |1100|IND_DOC|NRO_DE|DT_DE|NAT_EXP|NRO_RE|DT_RE|CHC_EMB|DT_CHC|DT_AVB|TP_CHC|PAIS|
func (b *Bloco1) writeRegistro1100() {
	for _, r := range b.Registro1001.Registro1100 {
		linha := b.LFillStr("1100", 0, false, '0') +
			b.LFillInt(int64(r.IndDoc), 0, false, '0') +
			b.LFillStr(r.NroDE, 0, false, '0') +
			b.LFillDate(r.DtDE, "02012006", true) +
			b.LFillInt(int64(r.NatExp), 0, false, '0') +
			b.LFillStr(r.NroRE, 0, false, '0') +
			b.LFillDate(r.DtRE, "02012006", true) +
			b.LFillStr(r.ChcEmb, 0, false, '0') +
			b.LFillDate(r.DtChc, "02012006", true) +
			b.LFillDate(r.DtAvb, "02012006", true) +
			b.LFillStr(r.TpChc.String(), 0, false, '0') +
			b.LFillStr(r.Pais, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1100Count++

		b.writeRegistro1105(r)
	}
}

// writeRegistro1105 gera as linhas do registro 1105.
// Formato: |1105|COD_MOD|SERIE|NUM_DOC|CHV_NFE|DT_DOC|COD_ITEM|
func (b *Bloco1) writeRegistro1105(parent *Registro1100) {
	for _, r := range parent.Registro1105 {
		linha := b.LFillStr("1105", 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Serie, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillStr(r.ChvNFe, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1105Count++

		b.writeRegistro1110(r)
	}
}

// writeRegistro1110 gera as linhas do registro 1110.
// Formato: |1110|COD_PART|COD_MOD|SER|NUM_DOC|DT_DOC|CHV_NFE|NR_MEMO|QTD|UNID|
func (b *Bloco1) writeRegistro1110(parent *Registro1105) {
	for _, r := range parent.Registro1110 {
		linha := b.LFillStr("1110", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.ChvNFe, 0, false, '0') +
			b.LFillStr(r.NrMemo, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1110Count++
	}
}

// writeRegistro1200 gera as linhas do registro 1200 (controle de creditos fiscais).
// Formato: |1200|COD_AJ_APUR|SLD_CRED|CRED_APR|CRED_RECEB|CRED_UTIL|SLD_CRED_FIM|
func (b *Bloco1) writeRegistro1200() {
	for _, r := range b.Registro1001.Registro1200 {
		linha := b.LFillStr("1200", 0, false, '0') +
			b.LFillStr(r.CodAjApur, 0, false, '0') +
			b.DFill(r.SldCred, 2, false) +
			b.DFill(r.CredApr, 2, false) +
			b.DFill(r.CredReceb, 2, false) +
			b.DFill(r.CredUtil, 2, false) +
			b.DFill(r.SldCredFim, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1200Count++

		b.writeRegistro1210(r)
	}
}

// writeRegistro1210 gera as linhas do registro 1210.
// Formato: |1210|TIPO_UTIL|NR_DOC|VL_CRED_UTIL|CHV_DOCe|
func (b *Bloco1) writeRegistro1210(parent *Registro1200) {
	for _, r := range parent.Registro1210 {
		linha := b.LFillStr("1210", 0, false, '0') +
			b.LFillStr(r.TipoUtil, 0, false, '0') +
			b.LFillStr(r.NrDoc, 0, false, '0') +
			b.DFill(r.VlCredUtil, 2, false) +
			b.LFillStr(r.ChvDOCe, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1210Count++
	}
}

// writeRegistro1250 gera as linhas do registro 1250.
// Formato: |1250|VL_CREDITO_ICMS_OP|VL_ICMS_ST_REST|VL_FCP_ST_REST|VL_ICMS_ST_COMPL|VL_FCP_ST_COMPL|
func (b *Bloco1) writeRegistro1250() {
	for _, r := range b.Registro1001.Registro1250 {
		linha := b.LFillStr("1250", 0, false, '0') +
			b.DFill(r.VlCreditoICMSOp, 2, false) +
			b.DFill(r.VlICMSSTRest, 2, false) +
			b.DFill(r.VlFCPSTRest, 2, false) +
			b.DFill(r.VlICMSSTCompl, 2, false) +
			b.DFill(r.VlFCPSTCompl, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1250Count++

		b.writeRegistro1255(r)
	}
}

// writeRegistro1255 gera as linhas do registro 1255.
// Formato: |1255|COD_MOT_REST_COMPL|VL_CREDITO_ICMS_OP_MOT|VL_ICMS_ST_REST_MOT|
//
//	VL_FCP_ST_REST_MOT|VL_ICMS_ST_COMPL_MOT|VL_FCP_ST_COMPL_MOT|
func (b *Bloco1) writeRegistro1255(parent *Registro1250) {
	for _, r := range parent.Registro1255 {
		linha := b.LFillStr("1255", 0, false, '0') +
			b.LFillStr(r.CodMotRestCompl, 0, false, '0') +
			b.DFill(r.VlCreditoICMSOpMot, 2, false) +
			b.DFill(r.VlICMSSTRestMot, 2, false) +
			b.DFill(r.VlFCPSTRestMot, 2, false) +
			b.DFill(r.VlICMSSTComplMot, 2, false) +
			b.DFill(r.VlFCPSTComplMot, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1255Count++
	}
}

// writeRegistro1300 gera as linhas do registro 1300 (combustiveis).
// Formato: |1300|COD_ITEM|DT_FECH|ESTQ_ABERT|VOL_ENTR|VOL_DISP|VOL_SAIDAS|
//
//	ESTQ_ESCR|VAL_AJ_PERDA|VAL_AJ_GANHO|FECH_FISICO|
func (b *Bloco1) writeRegistro1300() {
	for _, r := range b.Registro1001.Registro1300 {
		linha := b.LFillStr("1300", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillDate(r.DtFech, "02012006", true) +
			b.DFill(r.EstqAbert, 3, false) +
			b.DFill(r.VolEntr, 3, false) +
			b.DFill(r.VolDisp, 3, false) +
			b.DFill(r.VolSaidas, 3, false) +
			b.DFill(r.EstqEscr, 3, false) +
			b.DFill(r.ValAjPerda, 3, false) +
			b.DFill(r.ValAjGanho, 3, false) +
			b.DFill(r.FechFisico, 3, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1300Count++

		b.writeRegistro1310(r)
	}
}

// writeRegistro1310 gera as linhas do registro 1310 (por tanque).
// Formato: |1310|NUM_TANQUE|ESTQ_ABERT|VOL_ENTR|VOL_DISP|VOL_SAIDAS|ESTQ_ESCR|
//
//	VAL_AJ_PERDA|VAL_AJ_GANHO|FECH_FISICO|CAP_TANQUE|
//
// CAP_TANQUE so entra a partir de 01/2026, e a data avaliada e a do arquivo.
func (b *Bloco1) writeRegistro1310(parent *Registro1300) {
	for _, r := range parent.Registro1310 {
		linha := b.LFillStr("1310", 0, false, '0') +
			b.LFillStr(r.NumTanque, 0, false, '0') +
			b.DFill(r.EstqAbert, 3, false) +
			b.DFill(r.VolEntr, 3, false) +
			b.DFill(r.VolDisp, 3, false) +
			b.DFill(r.VolSaidas, 3, false) +
			b.DFill(r.EstqEscr, 3, false) +
			b.DFill(r.ValAjPerda, 3, false) +
			b.DFill(r.ValAjGanho, 3, false) +
			b.DFill(r.FechFisico, 3, false)
		if !b.DtIni.Before(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
			linha += b.LFillInt(int64(r.CapTanque), 6, false, '0')
		}
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1310Count++

		b.writeRegistro1320(r)
	}
}

// writeRegistro1320 gera as linhas do registro 1320 (volume de vendas).
// Formato: |1320|NUM_BICO|NR_INTERV|MOT_INTERV|NOM_INTERV|CNPJ_INTERV|CPF_INTERV|
//
//	VAL_FECHA|VAL_ABERT|VOL_AFERI|VOL_VENDAS|
func (b *Bloco1) writeRegistro1320(parent *Registro1310) {
	for _, r := range parent.Registro1320 {
		linha := b.LFillStr("1320", 0, false, '0') +
			b.LFillStr(r.NumBico, 0, false, '0') +
			b.LFillStr(r.NrInterv, 0, false, '0') +
			b.LFillStr(r.MotInterv, 0, false, '0') +
			b.LFillStr(r.NomInterv, 0, false, '0') +
			b.LFillStr(r.CnpjInterv, 0, false, '0') +
			b.LFillStr(r.CpfInterv, 0, false, '0') +
			b.DFill(r.ValFecha, 3, false) +
			b.DFill(r.ValAbert, 3, false) +
			b.DFill(r.VolAferi, 3, false) +
			b.DFill(r.VolVendas, 3, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1320Count++
	}
}

// writeRegistro1350 gera as linhas do registro 1350 (bombas).
// Formato: |1350|SERIE|FABRICANTE|MODELO|TIPO_MEDICAO|
func (b *Bloco1) writeRegistro1350() {
	for _, r := range b.Registro1001.Registro1350 {
		linha := b.LFillStr("1350", 0, false, '0') +
			b.LFillStr(r.Serie, 0, false, '0') +
			b.LFillStr(r.Fabricante, 0, false, '0') +
			b.LFillStr(r.Modelo, 0, false, '0') +
			b.LFillInt(int64(r.TipoMedicao), 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1350Count++

		b.writeRegistro1360(r)
		b.writeRegistro1370(r)
	}
}

// writeRegistro1360 gera as linhas do registro 1360 (lacres das bombas).
// Formato: |1360|NUM_LACRE|DT_APLICACAO|
func (b *Bloco1) writeRegistro1360(parent *Registro1350) {
	for _, r := range parent.Registro1360 {
		linha := b.LFillStr("1360", 0, false, '0') +
			b.LFillStr(r.NumLacre, 0, false, '0') +
			b.LFillDate(r.DtAplicacao, "02012006", true)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1360Count++
	}
}

// writeRegistro1370 gera as linhas do registro 1370 (bicos da bomba).
// Formato: |1370|NUM_BICO|COD_ITEM|NUM_TANQUE|
func (b *Bloco1) writeRegistro1370(parent *Registro1350) {
	for _, r := range parent.Registro1370 {
		linha := b.LFillStr("1370", 0, false, '0') +
			b.LFillStr(r.NumBico, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.NumTanque, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1370Count++
	}
}

// writeRegistro1390 gera as linhas do registro 1390 (producao de usina).
// Formato: |1390|COD_PROD|
func (b *Bloco1) writeRegistro1390() {
	for _, r := range b.Registro1001.Registro1390 {
		linha := b.LFillStr("1390", 0, false, '0') +
			b.LFillStr(r.CodProd, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1390Count++

		b.writeRegistro1391(r)
	}
}

// writeRegistro1391 gera as linhas do registro 1391 (producao diaria da usina).
//
// A partir da versao 113 entram COD_ITEM, TP_RESIDUO e QTD_RESIDUO; a partir da
// 117 entram tambem QTD_RESIDUO_DDG, QTD_RESIDUO_WDG e QTD_RESIDUO_CANA.
//
// TP_RESIDUO e declarado Integer no ACBr e emitido por LFill sem tamanho, o que
// cai no overload de data e o formata como ddmmaaaa. O comportamento e
// reproduzido aqui por fidelidade ao original.
func (b *Bloco1) writeRegistro1391(parent *Registro1390) {
	ver := b.versao1()
	for _, r := range parent.Registro1391 {
		linha := b.LFillStr("1391", 0, false, '0') +
			b.LFillDate(r.DtRegistro, "02012006", true) +
			b.DFill(r.QtdMoid, 2, false) +
			b.DFill(r.EstqIni, 2, false) +
			b.DFill(r.QtdProduz, 2, false) +
			b.DFill(r.EntAnidHid, 2, false) +
			b.DFill(r.OutrEntr, 2, false) +
			b.DFill(r.Perda, 2, false) +
			b.DFill(r.Cons, 2, false) +
			b.DFill(r.SaiAniHid, 2, false) +
			b.DFill(r.Saidas, 2, false) +
			b.DFill(r.EstqFin, 2, false) +
			b.DFill(r.EstqIniMel, 2, false) +
			b.DFill(r.ProdDiaMel, 2, false) +
			b.DFill(r.UtilMel, 2, false) +
			b.DFill(r.ProdAlcMel, 2, false) +
			b.LFillStr(r.Obs, 0, false, '0')
		if ver >= VlVersao113 {
			linha += b.LFillStr(r.CodItem, 0, false, '0') +
				b.LFillDate(tpResiduoComoData(r.TpResiduo), "02012006", true) +
				b.DFill(r.QtdResiduo, 2, false)
		}
		if ver >= VlVersao117 {
			linha += b.DFill(r.QtdResiduoDDG, 2, false) +
				b.DFill(r.QtdResiduoWDG, 2, false) +
				b.DFill(r.QtdResiduoCana, 2, false)
		}
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1391Count++
	}
}

// tpResiduoComoData reproduz o overload que o ACBr acaba escolhendo para
// TP_RESIDUO: o valor inteiro e interpretado como TDateTime e formatado como
// data. Zero devolve data zerada, que sai como campo vazio.
func tpResiduoComoData(v int) time.Time {
	if v == 0 {
		return time.Time{}
	}
	// TDateTime conta dias a partir de 30/12/1899.
	return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, v)
}

// writeRegistro1400 gera as linhas do registro 1400 (valores agregados).
// Formato: |1400|COD_ITEM_IPM|MUN|VALOR|
//
// Quando COD_ITEM_IPM esta vazio, o ACBr usa COD_ITEM no lugar.
func (b *Bloco1) writeRegistro1400() {
	for _, r := range b.Registro1001.Registro1400 {
		// Mesmo nome da local do ACBr (vCodItem), para a auditoria de campos
		// casar o nome posicao a posicao.
		codItem := strings.TrimSpace(r.CodItemIPM)
		if codItem == "" {
			codItem = strings.TrimSpace(r.CodItem)
		}
		linha := b.LFillStr("1400", 0, false, '0') +
			b.LFillStr(codItem, 0, false, '0') +
			b.LFillStr(strings.TrimSpace(r.Mun), 0, false, '0') +
			b.DFill(r.Valor, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1400Count++
	}
}

// writeRegistro1500 gera as linhas do registro 1500 (energia eletrica interestadual).
//
// O ACBr emite VL_DESC duas vezes e nunca emite VL_DOC. O comportamento e
// reproduzido por fidelidade ao original.
func (b *Bloco1) writeRegistro1500() {
	for _, r := range b.Registro1001.Registro1500 {
		linha := b.LFillStr("1500", 0, false, '0') +
			b.LFillStr(r.IndOper, 0, false, '0') +
			b.LFillStr(r.IndEmit, 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.CodSit.String(), 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.CodCons.String(), 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillDate(r.DtES, "02012006", true) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.DFill(r.VlForn, 2, false) +
			b.DFill(r.VlServNT, 2, false) +
			b.DFill(r.VlTerc, 2, false) +
			b.DFill(r.VlDa, 2, false) +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillInt(tpLigacaoInt(r.TpLigacao), 0, false, '0') +
			b.LFillStr(r.CodGrupoTensao.String(), 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1500Count++

		b.writeRegistro1510(r)
	}
}

// tpLigacaoInt devolve TP_LIGACAO como inteiro; o ACBr usa 1 como padrao para
// qualquer valor nao mapeado, inclusive "nenhum".
func tpLigacaoInt(v TpLigacao) int64 {
	switch v {
	case LigacaoMonofasico:
		return 1
	case LigacaoBifasico:
		return 2
	case LigacaoTrifasico:
		return 3
	}
	return 1
}

// writeRegistro1510 gera as linhas do registro 1510 (itens do 1500).
func (b *Bloco1) writeRegistro1510(parent *Registro1500) {
	for _, r := range parent.Registro1510 {
		linha := b.LFillStr("1510", 0, false, '0') +
			b.LFillStr(r.NumItem, 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.CodClass, 0, false, '0') +
			b.DFill(r.Qtd, 3, false) +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.DFill(r.VlItem, 2, false) +
			b.DFill(r.VlDesc, 2, false) +
			b.LFillStr(r.CstICMS, 3, false, '0') +
			b.LFillStr(r.CFOP, 4, false, '0') +
			b.DFill(r.VlBcICMS, 2, false) +
			b.DFill(r.AliqICMS, 2, false) +
			b.DFill(r.VlICMS, 2, false) +
			b.DFill(r.VlBcICMSST, 2, false) +
			b.DFill(r.AliqST, 2, false) +
			b.DFill(r.VlICMSST, 2, false) +
			b.LFillInt(int64(r.IndRec), 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.DFill(r.VlPIS, 2, false) +
			b.DFill(r.VlCOFINS, 2, false) +
			b.LFillStr(r.CodCta, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1510Count++
	}
}

// writeRegistro1600 gera as linhas do registro 1600.
// Formato: |1600|COD_PART|TOT_CREDITO|TOT_DEBITO|
func (b *Bloco1) writeRegistro1600() {
	for _, r := range b.Registro1001.Registro1600 {
		linha := b.LFillStr("1600", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.DFill(r.TotCredito, 2, false) +
			b.DFill(r.TotDebito, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1600Count++
	}
}

// writeRegistro1601 gera as linhas do registro 1601.
// Formato: |1601|COD_PART_IP|COD_PART_IT|TOT_VS|TOT_ISS|TOT_OUTROS|
func (b *Bloco1) writeRegistro1601() {
	for _, r := range b.Registro1001.Registro1601 {
		linha := b.LFillStr("1601", 0, false, '0') +
			b.LFillStr(r.CodPartIP, 0, false, '0') +
			b.LFillStr(r.CodPartIT, 0, false, '0') +
			b.DFill(r.TotVS, 2, false) +
			b.DFill(r.TotISS, 2, false) +
			b.DFill(r.TotOutros, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1601Count++
	}
}

// writeRegistro1700 gera as linhas do registro 1700 (documentos fiscais utilizados).
// Formato: |1700|COD_DISP|COD_MOD|SER|SUB|NUM_DOC_INI|NUM_DOC_FIN|NUM_AUT|
func (b *Bloco1) writeRegistro1700() {
	for _, r := range b.Registro1001.Registro1700 {
		b.Checkf(ehNumerico(r.NumDocIni),
			"(1-1700) Documento Fiscal: Numeracao incorreta %q para Documento Inicial", r.NumDocIni)
		b.Checkf(ehNumerico(r.NumDocFin),
			"(1-1700) Documento Fiscal: Numeracao incorreta %q para Documento Final", r.NumDocFin)
		b.Checkf(ehNumerico(r.NumAut),
			"(1-1700) Documento Fiscal: Numeracao incorreta %q para Numero Autorizacao", r.NumAut)

		linha := b.LFillStr("1700", 0, false, '0') +
			b.LFillStr(r.CodDisp.String(), 2, false, '0') +
			b.LFillStr(r.CodMod, 2, false, '0') +
			b.LFillStr(r.Ser, 4, false, '0') +
			b.LFillStr(r.Sub, 3, false, '0') +
			b.LFillStr(r.NumDocIni, 12, false, '0') +
			b.LFillStr(r.NumDocFin, 12, false, '0') +
			b.LFillStr(r.NumAut, 60, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1700Count++

		b.writeRegistro1710(r)
	}
}

// ehNumerico replica StrIsNumber do ACBr: verdadeiro quando a string, apos
// remover espacos, contem apenas digitos. String vazia conta como numerica.
func ehNumerico(s string) bool {
	for _, c := range strings.TrimSpace(s) {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// writeRegistro1710 gera as linhas do registro 1710.
// Formato: |1710|NUM_DOC_INI|NUM_DOC_FIN|
func (b *Bloco1) writeRegistro1710(parent *Registro1700) {
	for _, r := range parent.Registro1710 {
		b.Checkf(ehNumerico(r.NumDocIni),
			"(1-1710) Documento Fiscal Cancelado: Numeracao incorreta %q para Documento Inicial", r.NumDocIni)
		b.Checkf(ehNumerico(r.NumDocFin),
			"(1-1710) Documento Fiscal Cancelado: Numeracao incorreta %q para Documento Final", r.NumDocFin)

		linha := b.LFillStr("1710", 0, false, '0') +
			b.LFillStr(r.NumDocIni, 12, false, '0') +
			b.LFillStr(r.NumDocFin, 12, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1710Count++
	}
}

// writeRegistro1800 gera as linhas do registro 1800 (credito presumido de transporte).
// Formato: |1800|VL_CARGA|VL_PASS|VL_FAT|IND_RAT|VL_ICMS_ANT|VL_BC_ICMS|
//
//	VL_ICMS_APUR|VL_BC_ICMS_APUR|VL_DIF|
//
// E o unico registro do bloco em que todos os valores suprimem o zero.
func (b *Bloco1) writeRegistro1800() {
	for _, r := range b.Registro1001.Registro1800 {
		linha := b.LFillStr("1800", 0, false, '0') +
			b.DFill(r.VlCarga, 2, true) +
			b.DFill(r.VlPass, 2, true) +
			b.DFill(r.VlFat, 2, true) +
			b.LFillFloat(r.IndRat, 6, 2, true, '0', "") +
			b.DFill(r.VlICMSAnt, 2, true) +
			b.DFill(r.VlBcICMS, 2, true) +
			b.DFill(r.VlICMSApur, 2, true) +
			b.DFill(r.VlBcICMSApur, 2, true) +
			b.DFill(r.VlDif, 2, true)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1800Count++
	}
}

// writeRegistro1900 gera as linhas do registro 1900 (sub-apuracao do ICMS).
// Formato: |1900|IND_APUR_ICMS|DESCR_COMPL_OUT_APUR|
func (b *Bloco1) writeRegistro1900() {
	for _, r := range b.Registro1001.Registro1900 {
		linha := b.LFillStr("1900", 0, false, '0') +
			b.LFillStr(r.IndApurICMS, 1, false, '0') +
			b.LFillStr(r.DescrComplOutApur, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1900Count++

		b.writeRegistro1910(r)
	}
}

// writeRegistro1910 gera as linhas do registro 1910.
// Formato: |1910|DT_INI|DT_FIN|
func (b *Bloco1) writeRegistro1910(parent *Registro1900) {
	for _, r := range parent.Registro1910 {
		linha := b.LFillStr("1910", 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", true) +
			b.LFillDate(r.DtFin, "02012006", true)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1910Count++

		b.writeRegistro1920(r)
	}
}

// writeRegistro1920 gera as linhas do registro 1920 (sub-apuracao do ICMS).
func (b *Bloco1) writeRegistro1920(parent *Registro1910) {
	for _, r := range parent.Registro1920 {
		linha := b.LFillStr("1920", 0, false, '0') +
			b.DFill(r.VlTotTransfDebitosOA, 2, false) +
			b.DFill(r.VlTotAjDebitosOA, 2, false) +
			b.DFill(r.VlEstornosCredOA, 2, false) +
			b.DFill(r.VlTotTransfCreditosOA, 2, false) +
			b.DFill(r.VlTotAjCreditosOA, 2, false) +
			b.DFill(r.VlEstornosDebOA, 2, false) +
			b.DFill(r.VlSldCredorAntOA, 2, false) +
			b.DFill(r.VlSldApuradoOA, 2, false) +
			b.DFill(r.VlTotDed, 2, false) +
			b.DFill(r.VlICMSRecolherOA, 2, false) +
			b.DFill(r.VlSldCredorTranspOA, 2, false) +
			b.DFill(r.DebEspOA, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1920Count++

		b.writeRegistro1921(r)
		b.writeRegistro1925(r)
		b.writeRegistro1926(r)
	}
}

// writeRegistro1921 gera as linhas do registro 1921.
// Formato: |1921|COD_AJ_APUR|DESCR_COMPL_AJ|VL_AJ_APUR|
func (b *Bloco1) writeRegistro1921(parent *Registro1920) {
	for _, r := range parent.Registro1921 {
		linha := b.LFillStr("1921", 0, false, '0') +
			b.LFillStr(r.CodAjApur, 0, false, '0') +
			b.LFillStr(r.DescrComplAj, 0, false, '0') +
			b.DFill(r.VlAjApur, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1921Count++

		b.writeRegistro1922(r)
		b.writeRegistro1923(r)
	}
}

// writeRegistro1922 gera as linhas do registro 1922.
// Formato: |1922|NUM_DA|NUM_PROC|IND_PROC|PROC|TXT_COMPL|
func (b *Bloco1) writeRegistro1922(parent *Registro1921) {
	for _, r := range parent.Registro1922 {
		linha := b.LFillStr("1922", 0, false, '0') +
			b.LFillStr(r.NumDA, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc, 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1922Count++
	}
}

// writeRegistro1923 gera as linhas do registro 1923.
// Formato: |1923|COD_PART|COD_MOD|SER|SUB|NUM_DOC|DT_DOC|COD_ITEM|VL_AJ_ITEM|CHV_DOCe|
func (b *Bloco1) writeRegistro1923(parent *Registro1921) {
	for _, r := range parent.Registro1923 {
		linha := b.LFillStr("1923", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.CodMod, 0, false, '0') +
			b.LFillStr(r.Ser, 0, false, '0') +
			b.LFillStr(r.Sub, 0, false, '0') +
			b.LFillStr(r.NumDoc, 0, false, '0') +
			b.LFillDate(r.DtDoc, "02012006", true) +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.DFill(r.VlAjItem, 2, false) +
			b.LFillStr(r.ChvDOCe, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1923Count++
	}
}

// writeRegistro1925 gera as linhas do registro 1925.
// Formato: |1925|COD_INF_ADIC|VL_INF_ADIC|DESCR_COMPL_AJ|
func (b *Bloco1) writeRegistro1925(parent *Registro1920) {
	for _, r := range parent.Registro1925 {
		linha := b.LFillStr("1925", 0, false, '0') +
			b.LFillStr(r.CodInfAdic, 0, false, '0') +
			b.DFill(r.VlInfAdic, 2, false) +
			b.LFillStr(r.DescrComplAj, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1925Count++
	}
}

// writeRegistro1926 gera as linhas do registro 1926.
// Formato: |1926|COD_OR|VL_OR|DT_VCTO|COD_REC|NUM_PROC|IND_PROC|PROC|TXT_COMPL|MES_REF|
func (b *Bloco1) writeRegistro1926(parent *Registro1920) {
	for _, r := range parent.Registro1926 {
		linha := b.LFillStr("1926", 0, false, '0') +
			b.LFillStr(r.CodOR, 0, false, '0') +
			b.DFill(r.VlOR, 2, false) +
			b.LFillDate(r.DtVcto, "02012006", true) +
			b.LFillStr(r.CodRec, 0, false, '0') +
			b.LFillStr(r.NumProc, 0, false, '0') +
			b.LFillStr(r.IndProc, 0, false, '0') +
			b.LFillStr(r.Proc, 0, false, '0') +
			b.LFillStr(r.TxtCompl, 0, false, '0') +
			b.LFillStr(r.MesRef, 0, false, '0')
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1926Count++
	}
}

// writeRegistro1960 gera as linhas do registro 1960 (GIAF1).
// Formato: |1960|IND_AP|G1_01|...|G1_11|
func (b *Bloco1) writeRegistro1960() {
	for _, r := range b.Registro1001.Registro1960 {
		linha := b.LFillStr("1960", 0, false, '0') +
			b.LFillStr(r.IndAp, 0, false, '0') +
			b.DFill(r.G1_01, 2, false) + b.DFill(r.G1_02, 2, false) +
			b.DFill(r.G1_03, 2, false) + b.DFill(r.G1_04, 2, false) +
			b.DFill(r.G1_05, 2, false) + b.DFill(r.G1_06, 2, false) +
			b.DFill(r.G1_07, 2, false) + b.DFill(r.G1_08, 2, false) +
			b.DFill(r.G1_09, 2, false) + b.DFill(r.G1_10, 2, false) +
			b.DFill(r.G1_11, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1960Count++
	}
}

// writeRegistro1970 gera as linhas do registro 1970 (GIAF3).
// Formato: |1970|IND_AP|G3_01|...|G3_07|G3_T|G3_08|G3_09|
func (b *Bloco1) writeRegistro1970() {
	for _, r := range b.Registro1001.Registro1970 {
		linha := b.LFillStr("1970", 0, false, '0') +
			b.LFillStr(r.IndAp, 0, false, '0') +
			b.DFill(r.G3_01, 2, false) + b.DFill(r.G3_02, 2, false) +
			b.DFill(r.G3_03, 2, false) + b.DFill(r.G3_04, 2, false) +
			b.DFill(r.G3_05, 2, false) + b.DFill(r.G3_06, 2, false) +
			b.DFill(r.G3_07, 2, false) + b.DFill(r.G3_T, 2, false) +
			b.DFill(r.G3_08, 2, false) + b.DFill(r.G3_09, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1970Count++

		b.writeRegistro1975(r)
	}
}

// writeRegistro1975 gera as linhas do registro 1975 (GIAF3 por aliquota).
// Formato: |1975|ALIQ_IMP_BASE|G3_10|G3_11|G3_12|
func (b *Bloco1) writeRegistro1975(parent *Registro1970) {
	for _, r := range parent.Registro1975 {
		linha := b.LFillStr("1975", 0, false, '0') +
			b.DFill(r.AliqImpBase, 2, false) +
			b.DFill(r.G3_10, 2, false) +
			b.DFill(r.G3_11, 2, false) +
			b.DFill(r.G3_12, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1975Count++
	}
}

// writeRegistro1980 gera as linhas do registro 1980 (GIAF4).
// Formato: |1980|IND_AP|G4_01|...|G4_12|
func (b *Bloco1) writeRegistro1980() {
	for _, r := range b.Registro1001.Registro1980 {
		linha := b.LFillStr("1980", 0, false, '0') +
			b.LFillStr(r.IndAp, 0, false, '0') +
			b.DFill(r.G4_01, 2, false) + b.DFill(r.G4_02, 2, false) +
			b.DFill(r.G4_03, 2, false) + b.DFill(r.G4_04, 2, false) +
			b.DFill(r.G4_05, 2, false) + b.DFill(r.G4_06, 2, false) +
			b.DFill(r.G4_07, 2, false) + b.DFill(r.G4_08, 2, false) +
			b.DFill(r.G4_09, 2, false) + b.DFill(r.G4_10, 2, false) +
			b.DFill(r.G4_11, 2, false) + b.DFill(r.G4_12, 2, false)
		b.Add(linha, true)
		b.Registro1990.QtdLin1++
		b.Registro1980Count++
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
		b.LFillInt(int64(b.Registro9001.IndMov), 0, false, '0')
	b.Add(linha, true)
	b.Registro9990.QtdLin9++
}

// WriteRegistro9900 gera as linhas do registro 9900.
// Formato: |9900|REG_BLC|QTD_REG_BLC|
//
// A contagem do bloco soma, alem das proprias linhas 9900, mais duas: a do
// 9990 e a do 9999, que ainda nao foram escritas neste ponto. E o que o ACBr
// faz (QTD_LIN_9 := QTD_LIN_9 + Registro9900.Count + 2).
func (b *Bloco9) WriteRegistro9900() {
	for _, r := range b.Registro9900 {
		linha := b.LFillStr("9900", 0, false, '0') +
			b.LFillStr(r.RegBlc, 0, false, '0') +
			b.LFillInt(int64(r.QtdRegBlc), 0, false, '0')
		b.Add(linha, true)
	}
	b.Registro9990.QtdLin9 += len(b.Registro9900) + 2
}

// WriteRegistro9990 gera a linha do registro 9990.
// Formato: |9990|QTD_LIN_9|
//
// Nao incrementa o contador: a propria linha e a do 9999 ja foram somadas em
// WriteRegistro9900.
func (b *Bloco9) WriteRegistro9990() {
	if b.Registro9990 == nil {
		return
	}
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
