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

import "time"

// codNatCCValidos lista os codigos aceitos em COD_NAT_CC do registro 0500
// (natureza da conta ou grupo de contas).
// Ref.: ACBrEFDBloco_0_Class.pas, WriteRegistro0500.
var codNatCCValidos = map[string]bool{
	"01": true, "02": true, "03": true, "04": true,
	"05": true, "09": true, "10": true, "99": true,
}

// codPaisBrasil e o codigo BACEN do Brasil na tabela de paises. Participante
// com qualquer outro codigo e do exterior e tem tratamento proprio no 0150.
// O ACBr aceita as duas grafias, com e sem zero a esquerda.
func participanteDoExterior(codPais string) bool {
	return codPais != "01058" && codPais != "1058"
}

// codMun0150 devolve o campo COD_MUN do registro 0150. Participante do
// exterior nao tem municipio do IBGE: o layout exige o literal 9999999.
// Ref.: ACBrEFDBloco_0_Class.pas, WriteRegistro0150.
func codMun0150(b *Bloco0, r *Registro0150) string {
	if participanteDoExterior(r.CodPais) {
		return b.LFillStr("9999999", 0, false, '0')
	}
	return b.LFillInt(int64(r.CodMun), 7, false, '0')
}

// cest0200 devolve o campo CEST do registro 0200. O campo so passou a existir
// no layout em 01/01/2017; antes disso nao e emitido -- nem o delimitador.
// Ref.: ACBrEFDBloco_0_Class.pas, WriteRegistro0200.
func cest0200(b *Bloco0, r *Registro0200) string {
	inicioCEST := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
	if b.DtIni.Before(inicioCEST) {
		return ""
	}
	return b.LFillStr(r.CEST, 0, false, '0')
}

// codBarra0220 devolve o campo COD_BARRA do registro 0220, que so existe a
// partir da versao 115 do leiaute. Ate a 114 o registro tem tres campos.
// Ref.: ACBrEFDBloco_0_Class.pas, WriteRegistro0220.
func codBarra0220(b *Bloco0, r *Registro0220) string {
	if b.Registro0000 != nil && b.Registro0000.CodVer <= VlVersao114 {
		return ""
	}
	return b.LFillStr(r.CodBarra, 0, false, '0')
}

// ---------------------------------------------------------------------------
// WriteRegistro0000 - Abertura do arquivo digital e identificacao da entidade
// ---------------------------------------------------------------------------

// WriteRegistro0000 gera a linha do registro 0000 no arquivo SPED Fiscal.
// Formato: |0000|COD_VER|COD_FIN|DT_INI|DT_FIN|NOME|CNPJ|CPF|UF|IE|COD_MUN|IM|SUFRAMA|IND_PERFIL|IND_ATIV|
func (b *Bloco0) WriteRegistro0000() {
	if b.Registro0000 == nil {
		return
	}

	// Fire before event
	if b.OnBeforeWriteRegistro0000 != nil {
		var linha string
		b.OnBeforeWriteRegistro0000(&linha)
		if linha != "" {
			b.Add(linha, true)
		}
	}

	r := b.Registro0000
	linha := b.LFillStr("0000", 0, false, '0') +
		b.LFillStr(r.CodVer.String(), 0, false, '0') +
		b.LFillInt(int64(r.CodFin), 1, false, '0') +
		b.LFillDate(r.DtIni, "02012006", false) +
		b.LFillDate(r.DtFin, "02012006", false) +
		b.LFillStr(r.Nome, 0, false, '0') +
		b.LFillStr(r.CNPJ, 0, false, '0') +
		b.LFillStr(r.CPF, 0, false, '0') +
		b.LFillStr(r.UF, 0, false, '0') +
		b.LFillStr(r.IE, 0, false, '0') +
		b.LFillInt(int64(r.CodMun), 7, false, '0') +
		b.LFillStr(r.IM, 0, false, '0') +
		b.LFillStr(r.Suframa, 0, false, '0') +
		b.LFillStr(r.IndPerfil.String(), 0, false, '0') +
		b.LFillInt(int64(r.IndAtiv), 1, false, '0')

	// Fire write event
	if b.OnWriteRegistro0000 != nil {
		b.OnWriteRegistro0000(&linha)
	}
	b.Add(linha, true)
	b.Registro0990.QtdLin0++

	// Fire after event
	if b.OnAfterWriteRegistro0000 != nil {
		var linhaAfter string
		b.OnAfterWriteRegistro0000(&linhaAfter)
		if linhaAfter != "" {
			b.Add(linhaAfter, true)
		}
	}
}

// ---------------------------------------------------------------------------
// WriteRegistro0001 - Abertura do Bloco 0
// ---------------------------------------------------------------------------

// WriteRegistro0001 gera a linha do registro 0001 e dispara a escrita de
// todos os sub-registros do Bloco 0 quando IndMov == 0 (com dados).
// Formato: |0001|IND_MOV|
func (b *Bloco0) WriteRegistro0001() {
	if b.Registro0001 == nil {
		return
	}

	linha := b.LFillStr("0001", 0, false, '0') +
		b.LFillInt(int64(b.Registro0001.IndMov), 0, false, '0')
	b.Add(linha, true)

	// Registro0002 se industrial e periodo >= 2020
	if b.Registro0002 != nil && !b.DtIni.IsZero() &&
		b.DtIni.Year() >= 2020 && b.Registro0000 != nil && b.Registro0000.IndAtiv == AtivIndustrial {
		linha = b.LFillStr("0002", 0, false, '0') +
			b.LFillStr(b.Registro0002.ClasEstabInd, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
	}

	if b.Registro0001.IndMov == 0 { // com dados
		b.writeRegistro0005()
		b.writeRegistro0015()
		b.writeRegistro0100()
		b.writeRegistro0150()
		b.writeRegistro0190()
		b.writeRegistro0200()
		b.writeRegistro0300()
		b.writeRegistro0400()
		b.writeRegistro0450()
		b.writeRegistro0460()
		b.writeRegistro0500()
		b.writeRegistro0600()
	}
	b.Registro0990.QtdLin0++
}

// ---------------------------------------------------------------------------
// writeRegistro0005 - Dados complementares da entidade
// ---------------------------------------------------------------------------

// writeRegistro0005 gera a linha do registro 0005.
// Formato: |0005|FANTASIA|CEP|ENDERECO|NUM|COMPL|BAIRRO|FONE|FAX|EMAIL|
func (b *Bloco0) writeRegistro0005() {
	r := b.Registro0001.Registro0005
	if r == nil {
		return
	}
	linha := b.LFillStr("0005", 0, false, '0') +
		b.LFillStr(r.Fantasia, 0, false, '0') +
		b.LFillStr(r.CEP, 8, false, '0') +
		b.LFillStr(r.Endereco, 0, false, '0') +
		b.LFillStr(r.Num, 0, false, '0') +
		b.LFillStr(r.Compl, 0, false, '0') +
		b.LFillStr(r.Bairro, 0, false, '0') +
		b.LFillStr(r.Fone, 0, false, '0') +
		b.LFillStr(r.Fax, 0, false, '0') +
		b.LFillStr(r.Email, 0, false, '0')
	b.Add(linha, true)
	b.Registro0990.QtdLin0++
	b.Registro0005Count++
}

// ---------------------------------------------------------------------------
// writeRegistro0015 - Dados do contribuinte substituto
// ---------------------------------------------------------------------------

// writeRegistro0015 itera sobre Registro0001.Registro0015 e gera as linhas.
// Formato: |0015|UF_ST|IE_ST|
func (b *Bloco0) writeRegistro0015() {
	for _, r := range b.Registro0001.Registro0015 {
		linha := b.LFillStr("0015", 0, false, '0') +
			b.LFillStr(r.UfST, 0, false, '0') +
			b.LFillStr(r.IeST, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0015Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0100 - Dados do contabilista
// ---------------------------------------------------------------------------

// writeRegistro0100 gera a linha do registro 0100.
// Formato: |0100|NOME|CPF|CRC|CNPJ|CEP|ENDERECO|NUM|COMPL|BAIRRO|FONE|FAX|EMAIL|COD_MUN|
func (b *Bloco0) writeRegistro0100() {
	r := b.Registro0001.Registro0100
	if r == nil {
		return
	}
	linha := b.LFillStr("0100", 0, false, '0') +
		b.LFillStr(r.Nome, 0, false, '0') +
		b.LFillStr(r.CPF, 0, false, '0') +
		b.LFillStr(r.CRC, 0, false, '0') +
		b.LFillStr(r.CNPJ, 0, false, '0') +
		b.LFillStr(r.CEP, 8, false, '0') +
		b.LFillStr(r.Endereco, 0, false, '0') +
		b.LFillStr(r.Num, 0, false, '0') +
		b.LFillStr(r.Compl, 0, false, '0') +
		b.LFillStr(r.Bairro, 0, false, '0') +
		b.LFillStr(r.Fone, 0, false, '0') +
		b.LFillStr(r.Fax, 0, false, '0') +
		b.LFillStr(r.Email, 0, false, '0') +
		b.LFillInt(int64(r.CodMun), 7, false, '0')
	b.Add(linha, true)
	b.Registro0990.QtdLin0++
	b.Registro0100Count++
}

// ---------------------------------------------------------------------------
// writeRegistro0150 - Tabela de cadastro do participante
// ---------------------------------------------------------------------------

// writeRegistro0150 itera sobre Registro0001.Registro0150 e gera as linhas.
// Para cada registro, tambem chama writeRegistro0175.
// Formato: |0150|COD_PART|NOME|COD_PAIS|CNPJ|CPF|IE|COD_MUN|SUFRAMA|ENDERECO|NUM|COMPL|BAIRRO|
func (b *Bloco0) writeRegistro0150() {
	for _, r := range b.Registro0001.Registro0150 {
		linha := b.LFillStr("0150", 0, false, '0') +
			b.LFillStr(r.CodPart, 0, false, '0') +
			b.LFillStr(r.Nome, 0, false, '0') +
			b.LFillStr(r.CodPais, 0, false, '0') +
			b.LFillStr(r.CNPJ, 0, false, '0') +
			b.LFillStr(r.CPF, 0, false, '0') +
			b.LFillStr(r.IE, 0, false, '0') +
			codMun0150(b, r) +
			b.LFillStr(r.Suframa, 0, false, '0') +
			b.LFillStr(r.Endereco, 0, false, '0') +
			b.LFillStr(r.Num, 0, false, '0') +
			b.LFillStr(r.Compl, 0, false, '0') +
			b.LFillStr(r.Bairro, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0150Count++

		b.writeRegistro0175(r)
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0175 - Alteracao da tabela de cadastro de participante
// ---------------------------------------------------------------------------

// writeRegistro0175 itera sobre os sub-registros 0175 de um Registro0150.
// Formato: |0175|DT_ALT|NR_CAMPO|CONT_ANT|
func (b *Bloco0) writeRegistro0175(parent *Registro0150) {
	for _, r := range parent.Registro0175 {
		linha := b.LFillStr("0175", 0, false, '0') +
			b.LFillDate(r.DtAlt, "02012006", false) +
			b.LFillStr(r.NrCampo, 0, false, '0') +
			b.LFillStr(r.ContAnt, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0175Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0190 - Identificacao das unidades de medida
// ---------------------------------------------------------------------------

// writeRegistro0190 itera sobre Registro0001.Registro0190 e gera as linhas.
// Formato: |0190|UNID|DESCR|
func (b *Bloco0) writeRegistro0190() {
	for _, r := range b.Registro0001.Registro0190 {
		linha := b.LFillStr("0190", 0, false, '0') +
			b.LFillStr(r.Unid, 0, false, '0') +
			b.LFillStr(r.Descr, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0190Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0200 - Tabela de identificacao do item
// ---------------------------------------------------------------------------

// writeRegistro0200 itera sobre Registro0001.Registro0200 e gera as linhas.
// Para cada registro, tambem chama writeRegistro0205, 0206, 0210, 0220, 0221.
// Formato: |0200|COD_ITEM|DESCR_ITEM|COD_BARRA|COD_ANT_ITEM|UNID_INV|TIPO_ITEM|COD_NCM|EX_IPI|COD_GEN|COD_LST|ALIQ_ICMS|CEST|
func (b *Bloco0) writeRegistro0200() {
	for _, r := range b.Registro0001.Registro0200 {
		linha := b.LFillStr("0200", 0, false, '0') +
			b.LFillStr(r.CodItem, 0, false, '0') +
			b.LFillStr(r.DescrItem, 0, false, '0') +
			b.LFillStr(r.CodBarra, 0, false, '0') +
			b.LFillStr(r.CodAntItem, 0, false, '0') +
			b.LFillStr(r.UnidInv, 0, false, '0') +
			b.LFillStr(r.TipoItem.String(), 0, false, '0') +
			b.LFillStr(r.CodNCM, 0, false, '0') +
			b.LFillStr(r.ExIPI, 0, false, '0') +
			b.LFillStr(r.CodGen, 0, false, '0') +
			b.LFillStr(r.CodLst, 0, false, '0') +
			b.DFill(r.AliqICMS, 2, true) +
			cest0200(b, r)
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0200Count++

		b.writeRegistro0205(r)
		b.writeRegistro0206(r)
		b.writeRegistro0210(r)
		b.writeRegistro0220(r)
		b.writeRegistro0221(r)
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0205 - Alteracao do item
// ---------------------------------------------------------------------------

// writeRegistro0205 itera sobre os sub-registros 0205 de um Registro0200.
// Formato: |0205|DESCR_ANT_ITEM|DT_INI|DT_FIN|COD_ANT_ITEM|
func (b *Bloco0) writeRegistro0205(parent *Registro0200) {
	for _, r := range parent.Registro0205 {
		linha := b.LFillStr("0205", 0, false, '0') +
			b.LFillStr(r.DescrAntItem, 0, false, '0') +
			b.LFillDate(r.DtIni, "02012006", false) +
			b.LFillDate(r.DtFin, "02012006", false) +
			b.LFillStr(r.CodAntItem, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0205Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0206 - Codigo de produto conforme tabela ANP
// ---------------------------------------------------------------------------

// writeRegistro0206 itera sobre os sub-registros 0206 de um Registro0200.
// Formato: |0206|COD_COMB|
func (b *Bloco0) writeRegistro0206(parent *Registro0200) {
	for _, r := range parent.Registro0206 {
		linha := b.LFillStr("0206", 0, false, '0') +
			b.LFillStr(r.CodComb, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0206Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0210 - Consumo especifico padronizado
// ---------------------------------------------------------------------------

// writeRegistro0210 itera sobre os sub-registros 0210 de um Registro0200.
// Formato: |0210|COD_ITEM_COMP|QTD_COMP|PERDA|
func (b *Bloco0) writeRegistro0210(parent *Registro0200) {
	for _, r := range parent.Registro0210 {
		linha := b.LFillStr("0210", 0, false, '0') +
			b.LFillStr(r.CodItemComp, 0, false, '0') +
			b.DFill(r.QtdComp, 2, false) +
			b.DFill(r.Perda, 2, false)
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0210Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0220 - Fatores de conversao de unidades
// ---------------------------------------------------------------------------

// writeRegistro0220 itera sobre os sub-registros 0220 de um Registro0200.
// Formato: |0220|UNID_CONV|FAT_CONV|COD_BARRA|
//
// COD_BARRA so existe no layout a partir da versao 115; ate a 114 o registro
// tem apenas tres campos. Ref.: ACBrEFDBloco_0_Class.pas, WriteRegistro0220.
func (b *Bloco0) writeRegistro0220(parent *Registro0200) {
	for _, r := range parent.Registro0220 {
		linha := b.LFillStr("0220", 0, false, '0') +
			b.LFillStr(r.UnidConv, 0, false, '0') +
			b.DFill(r.FatConv, 6, false) +
			codBarra0220(b, r)
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0220Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0221 - Correlacao entre codigos de itens comercializados
// ---------------------------------------------------------------------------

// writeRegistro0221 itera sobre os sub-registros 0221 de um Registro0200.
// Formato: |0221|COD_ITEM_ATOMICO|QTDE_CONTIDA|
func (b *Bloco0) writeRegistro0221(parent *Registro0200) {
	for _, r := range parent.Registro0221 {
		linha := b.LFillStr("0221", 0, false, '0') +
			b.LFillStr(r.CodItemAtomico, 0, false, '0') +
			b.DFill(r.QtdeContida, 2, false)
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0221Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0300 - Cadastro de bens do ativo imobilizado
// ---------------------------------------------------------------------------

// writeRegistro0300 itera sobre Registro0001.Registro0300 e gera as linhas.
// Para cada registro, tambem chama writeRegistro0305.
// Formato: |0300|COD_IND_BEM|IDENT_MERC|DESCR_ITEM|COD_PRNC|COD_CTA|NR_PARC|
func (b *Bloco0) writeRegistro0300() {
	for _, r := range b.Registro0001.Registro0300 {
		linha := b.LFillStr("0300", 0, false, '0') +
			b.LFillStr(r.CodIndBem, 0, false, '0') +
			b.LFillInt(int64(r.IdentMerc), 1, false, '0') +
			b.LFillStr(r.DescrItem, 0, false, '0') +
			b.LFillStr(r.CodPrnc, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0') +
			b.DFill(r.NrParc, 0, false)
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0300Count++

		b.writeRegistro0305(r)
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0305 - Informacao sobre utilizacao do bem
// ---------------------------------------------------------------------------

// writeRegistro0305 gera a linha do sub-registro 0305 de um Registro0300.
// Formato: |0305|COD_CCUS|FUNC|VIDA_UTIL|
func (b *Bloco0) writeRegistro0305(parent *Registro0300) {
	r := parent.Registro0305
	if r == nil {
		return
	}
	linha := b.LFillStr("0305", 0, false, '0') +
		b.LFillStr(r.CodCcus, 0, false, '0') +
		b.LFillStr(r.Func, 0, false, '0') +
		b.LFillInt(int64(r.VidaUtil), 0, false, '0')
	b.Add(linha, true)
	b.Registro0990.QtdLin0++
	b.Registro0305Count++
}

// ---------------------------------------------------------------------------
// writeRegistro0400 - Tabela de natureza da operacao
// ---------------------------------------------------------------------------

// writeRegistro0400 itera sobre Registro0001.Registro0400 e gera as linhas.
// Formato: |0400|COD_NAT|DESCR_NAT|
func (b *Bloco0) writeRegistro0400() {
	for _, r := range b.Registro0001.Registro0400 {
		linha := b.LFillStr("0400", 0, false, '0') +
			b.LFillStr(r.CodNat, 0, false, '0') +
			b.LFillStr(r.DescrNat, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0400Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0450 - Tabela de informacao complementar
// ---------------------------------------------------------------------------

// writeRegistro0450 itera sobre Registro0001.Registro0450 e gera as linhas.
// Formato: |0450|COD_INF|TXT|
func (b *Bloco0) writeRegistro0450() {
	for _, r := range b.Registro0001.Registro0450 {
		linha := b.LFillStr("0450", 0, false, '0') +
			b.LFillStr(r.CodInf, 0, false, '0') +
			b.LFillStr(r.Txt, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0450Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0460 - Tabela de observacoes do lancamento fiscal
// ---------------------------------------------------------------------------

// writeRegistro0460 itera sobre Registro0001.Registro0460 e gera as linhas.
// Formato: |0460|COD_OBS|TXT|
func (b *Bloco0) writeRegistro0460() {
	for _, r := range b.Registro0001.Registro0460 {
		linha := b.LFillStr("0460", 0, false, '0') +
			b.LFillStr(r.CodObs, 0, false, '0') +
			b.LFillStr(r.Txt, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0460Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0500 - Plano de contas contabeis
// ---------------------------------------------------------------------------

// writeRegistro0500 itera sobre Registro0001.Registro0500 e gera as linhas.
// Formato: |0500|DT_ALT|COD_NAT_CC|IND_CTA|NIVEL|COD_CTA|NOME_CTA|
func (b *Bloco0) writeRegistro0500() {
	for _, r := range b.Registro0001.Registro0500 {
		b.Checkf(codNatCCValidos[r.CodNatCC],
			"(0-0500) O codigo da natureza da conta/grupo de contas %q digitado e invalido!",
			r.CodNatCC)
		b.Checkf(r.IndCta == "S" || r.IndCta == "A",
			"(0-0500) O indicador %q do tipo de conta deve ser informado S ou A!",
			r.IndCta)

		linha := b.LFillStr("0500", 0, false, '0') +
			b.LFillDate(r.DtAlt, "02012006", false) +
			b.LFillStr(r.CodNatCC, 0, false, '0') +
			b.LFillStr(r.IndCta, 0, false, '0') +
			b.LFillStr(r.Nivel, 0, false, '0') +
			b.LFillStr(r.CodCta, 0, false, '0') +
			b.LFillStr(r.NomeCta, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0500Count++
	}
}

// ---------------------------------------------------------------------------
// writeRegistro0600 - Centro de custos
// ---------------------------------------------------------------------------

// writeRegistro0600 itera sobre Registro0001.Registro0600 e gera as linhas.
// Formato: |0600|DT_ALT|COD_CCUS|CCUS|
func (b *Bloco0) writeRegistro0600() {
	for _, r := range b.Registro0001.Registro0600 {
		linha := b.LFillStr("0600", 0, false, '0') +
			b.LFillDate(r.DtAlt, "02012006", false) +
			b.LFillStr(r.CodCcus, 0, false, '0') +
			b.LFillStr(r.Ccus, 0, false, '0')
		b.Add(linha, true)
		b.Registro0990.QtdLin0++
		b.Registro0600Count++
	}
}

// ---------------------------------------------------------------------------
// WriteRegistro0990 - Encerramento do Bloco 0
// ---------------------------------------------------------------------------

// WriteRegistro0990 gera a linha do registro 0990 (encerramento do Bloco 0).
// Formato: |0990|QTD_LIN_0|
func (b *Bloco0) WriteRegistro0990() {
	if b.Registro0990 == nil {
		return
	}
	b.Registro0990.QtdLin0++ // inclui esta propria linha de encerramento
	linha := b.LFillStr("0990", 0, false, '0') +
		b.LFillInt(int64(b.Registro0990.QtdLin0), 0, false, '0')
	b.Add(linha, true)
}
