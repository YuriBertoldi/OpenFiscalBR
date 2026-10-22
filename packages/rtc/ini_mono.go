// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package rtc

import (
	"fmt"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Monofasia (gIBSCBSMono) e Imposto Seletivo (gIS/ISTot) no formato .ini.
// Porte dos grupos correspondentes de ACBrDFe.RTC.IniReader.pas e
// ACBrDFe.RTC.IniWriter.pas.
//
// Ficaram fora do porte original do .ini porque a NFGas nao os emite. A
// NF-e emite: o CST 620 ("tributacao monofasica") e o caminho de
// combustiveis, e e o UNICO CST da NF-e em que o grupo gIBSCBS nao sai e
// esta arvore sai no lugar dele.
//
// A arvore tem quatro ramos, e os quatro repetem a MESMA forma:
//
//	gIBSMonoAdRem      ad rem do IBS       -> padrao, reten, ret, bio
//	gIBSMonoAdValorem  ad valorem do IBS   -> padrao, reten, ret, bio
//	gCBSMonoAdRem      ad rem da CBS       -> padrao, reten, ret, bio
//	gCBSMonoAdValorem  ad valorem da CBS   -> padrao, reten, ret, bio
//
// "Ad rem" cobra por QUANTIDADE (qBCMono x adRem) e "ad valorem" cobra por
// VALOR (vBCMono x pAliq). Cada um tem a sua secao: `gMonoPadraoIBSQtde`
// contra `gMonoPadraoIBSAliq`. Sao secoes distintas com campos distintos --
// misturar as duas e o erro facil aqui.
//
// Duas particularidades do original, preservadas:
//
//   - `gMonoRet` tem secoes diferentes por ramo (`gMonoRetIBSQtde` x
//     `gMonoRetIBSAliq`) mas o MESMO tipo de um campo so. Uma leitura
//     descuidada apontaria os dois ramos para a mesma secao;
//   - `gpBioDiferenca` tem UMA secao por tributo (`gpBioDiferencaIBS`), lida
//     pelos DOIS ramos daquele tributo. O mesmo dado termina em
//     `GIBSMonoAdRem.GpBioDiferenca` e em `GIBSMonoAdValorem.GpBioDiferenca`.
//     E assim no ACBr (Ler_gIBSMonoAdRem e Ler_gIBSMonoAdValorem chamam o
//     mesmo Ler_gpBioDiferencaIBS) e esta replicado.

// secaoMono monta o nome das secoes da monofasia, que indexam pelo primeiro
// indice com tres digitos.
func secaoMono(nome string, idx int) string {
	if idx == SemIndice {
		return nome
	}
	return fmt.Sprintf("%s%03d", nome, idx)
}

// ---------------------------------------------------------------------------
// Leitura
// ---------------------------------------------------------------------------

// LerINIGIBSCBSMono le [gIBSCBSMonoNNN] e os quatro ramos.
// Porte de Ler_gIBSCBSMono.
//
// Diferente dos demais grupos do item, a monofasia usa APENAS o primeiro
// indice -- o ACBr passa so `Idx1` adiante (Ler_IBSCBS:237).
func LerINIGIBSCBSMono(ini *pcn.INI, v *GIBSCBSMono, idx int) {
	if ini == nil || v == nil {
		return
	}
	secao := secaoMono("gIBSCBSMono", idx)
	if !ini.SecaoExiste(secao) {
		return
	}

	v.VTotIBSMonoItem = ini.LerFloat(secao, "vTotIBSMonoItem", 0)
	v.VTotCBSMonoItem = ini.LerFloat(secao, "vTotCBSMonoItem", 0)

	lerINIGIBSMonoAdRem(ini, &v.GIBSMonoAdRem, idx)
	lerINIGIBSMonoAdValorem(ini, &v.GIBSMonoAdValorem, idx)
	lerINIGCBSMonoAdRem(ini, &v.GCBSMonoAdRem, idx)
	lerINIGCBSMonoAdValorem(ini, &v.GCBSMonoAdValorem, idx)
}

// lerINIGIBSMonoAdRem le o ramo ad rem do IBS. Porte de Ler_gIBSMonoAdRem.
func lerINIGIBSMonoAdRem(ini *pcn.INI, v *GIBSMonoAdRem, idx int) {
	if s := secaoMono("gMonoPadraoIBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoPadrao.QBCMono = ini.LerFloat(s, "qBCMono", 0)
		v.GMonoPadrao.AdRemIBS = ini.LerFloat(s, "adRemIBS", 0)
		v.GMonoPadrao.VIBSMono = ini.LerFloat(s, "vIBSMono", 0)
	}
	if s := secaoMono("gMonoRetenIBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoReten.QBCMonoReten = ini.LerFloat(s, "qBCMonoReten", 0)
		v.GMonoReten.AdRemIBSReten = ini.LerFloat(s, "adRemIBSReten", 0)
		v.GMonoReten.VIBSMonoReten = ini.LerFloat(s, "vIBSMonoReten", 0)
	}
	// A secao do gMonoRet carrega o sufixo do RAMO, nao do tributo.
	if s := secaoMono("gMonoRetIBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoRet.VIBSMonoRet = ini.LerFloat(s, "vIBSMonoRet", 0)
	}
	lerINIGpBioDiferencaIBS(ini, &v.GpBioDiferenca, idx)
}

// lerINIGIBSMonoAdValorem le o ramo ad valorem do IBS.
// Porte de Ler_gIBSMonoAdValorem.
func lerINIGIBSMonoAdValorem(ini *pcn.INI, v *GIBSMonoAdValorem, idx int) {
	if s := secaoMono("gMonoPadraoIBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoPadrao.VBCMono = ini.LerFloat(s, "vBCMono", 0)
		v.GMonoPadrao.PAliqMonoUF = ini.LerFloat(s, "pAliqMonoUF", 0)
		v.GMonoPadrao.VIBSMonoUF = ini.LerFloat(s, "vIBSMonoUF", 0)
		v.GMonoPadrao.PAliqMonoMun = ini.LerFloat(s, "pAliqMonoMun", 0)
		v.GMonoPadrao.VIBSMonoMun = ini.LerFloat(s, "vIBSMonoMun", 0)
		v.GMonoPadrao.VIBSMono = ini.LerFloat(s, "vIBSMono", 0)
	}
	if s := secaoMono("gMonoRetenIBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoReten.VBCMonoReten = ini.LerFloat(s, "vBCMonoReten", 0)
		v.GMonoReten.PAliqMonoReten = ini.LerFloat(s, "pAliqMonoReten", 0)
		v.GMonoReten.VIBSMonoReten = ini.LerFloat(s, "vIBSMonoReten", 0)
	}
	if s := secaoMono("gMonoRetIBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoRet.VIBSMonoRet = ini.LerFloat(s, "vIBSMonoRet", 0)
	}
	lerINIGpBioDiferencaIBS(ini, &v.GpBioDiferenca, idx)
}

// lerINIGpBioDiferencaIBS le [gpBioDiferencaIBSNNN].
// Porte de Ler_gpBioDiferencaIBS.
//
// Lido pelos DOIS ramos do IBS, da MESMA secao -- ver o cabecalho do
// arquivo.
func lerINIGpBioDiferencaIBS(ini *pcn.INI, v *GpBioDiferencaIBS, idx int) {
	s := secaoMono("gpBioDiferencaIBS", idx)
	if !ini.SecaoExiste(s) {
		return
	}
	v.QBCBioComb = ini.LerFloat(s, "qBCBioComb", 0)
	v.VIBSDiferenca = ini.LerFloat(s, "vIBSDiferenca", 0)
}

// lerINIGCBSMonoAdRem le o ramo ad rem da CBS. Porte de Ler_gCBSMonoAdRem.
func lerINIGCBSMonoAdRem(ini *pcn.INI, v *GCBSMonoAdRem, idx int) {
	if s := secaoMono("gMonoPadraoCBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoPadrao.QBCMono = ini.LerFloat(s, "qBCMono", 0)
		v.GMonoPadrao.AdRemCBS = ini.LerFloat(s, "adRemCBS", 0)
		v.GMonoPadrao.VCBSMono = ini.LerFloat(s, "vCBSMono", 0)
	}
	if s := secaoMono("gMonoRetenCBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoReten.QBCMonoReten = ini.LerFloat(s, "qBCMonoReten", 0)
		v.GMonoReten.AdRemCBSReten = ini.LerFloat(s, "adRemCBSReten", 0)
		v.GMonoReten.VCBSMonoReten = ini.LerFloat(s, "vCBSMonoReten", 0)
	}
	if s := secaoMono("gMonoRetCBSQtde", idx); ini.SecaoExiste(s) {
		v.GMonoRet.VCBSMonoRet = ini.LerFloat(s, "vCBSMonoRet", 0)
	}
	lerINIGpBioDiferencaCBS(ini, &v.GpBioDiferenca, idx)
}

// lerINIGCBSMonoAdValorem le o ramo ad valorem da CBS.
// Porte de Ler_gCBSMonoAdValorem.
func lerINIGCBSMonoAdValorem(ini *pcn.INI, v *GCBSMonoAdValorem, idx int) {
	if s := secaoMono("gMonoPadraoCBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoPadrao.VBCMono = ini.LerFloat(s, "vBCMono", 0)
		v.GMonoPadrao.PAliqMonoCBS = ini.LerFloat(s, "pAliqMonoCBS", 0)
		v.GMonoPadrao.VCBSMono = ini.LerFloat(s, "vCBSMono", 0)
	}
	if s := secaoMono("gMonoRetenCBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoReten.VBCMonoReten = ini.LerFloat(s, "vBCMonoReten", 0)
		v.GMonoReten.PAliqMonoReten = ini.LerFloat(s, "pAliqMonoReten", 0)
		v.GMonoReten.VCBSMonoReten = ini.LerFloat(s, "vCBSMonoReten", 0)
	}
	if s := secaoMono("gMonoRetCBSAliq", idx); ini.SecaoExiste(s) {
		v.GMonoRet.VCBSMonoRet = ini.LerFloat(s, "vCBSMonoRet", 0)
	}
	lerINIGpBioDiferencaCBS(ini, &v.GpBioDiferenca, idx)
}

// lerINIGpBioDiferencaCBS le [gpBioDiferencaCBSNNN].
// Porte de Ler_gpBioDiferencaCBS.
func lerINIGpBioDiferencaCBS(ini *pcn.INI, v *GpBioDiferencaCBS, idx int) {
	s := secaoMono("gpBioDiferencaCBS", idx)
	if !ini.SecaoExiste(s) {
		return
	}
	v.QBCBioComb = ini.LerFloat(s, "qBCBioComb", 0)
	v.VCBSDiferenca = ini.LerFloat(s, "vCBSDiferenca", 0)
}

// LerINIDFeReferenciado le [DFeReferenciadoNNN], o documento e o item que a
// operacao referencia. Porte de Ler_DFeReferenciado.
//
// Vive no ITEM da NF-e, nao no cabecalho -- cada item pode apontar para um
// item diferente de um documento anterior.
func LerINIDFeReferenciado(ini *pcn.INI, v *DFeReferenciado, idx int) {
	if ini == nil || v == nil {
		return
	}
	secao := secaoMono("DFeReferenciado", idx)
	if !ini.SecaoExiste(secao) {
		return
	}
	v.ChaveAcesso = ini.LerString(secao, "chaveAcesso", "")
	v.NItem = ini.LerInteiro(secao, "nItem", 0)
}

// GravarINIDFeReferenciado grava [DFeReferenciadoNNN].
// Porte de Gerar_DFeReferenciado.
//
// O original grava a secao SEM condicao nenhuma -- todo item de uma NF-e
// produz um [DFeReferenciadoNNN] com `chaveAcesso=` e `nItem=0`. O porte
// acrescenta a guarda de grupo vazio, pela mesma razao do GravarINIISel: um
// item que nao referencia documento algum nao deve deixar uma secao inutil
// no arquivo. O efeito e que o .ini sai MENOR que o do ACBr, nunca
// diferente no conteudo.
func GravarINIDFeReferenciado(ini *pcn.INI, v DFeReferenciado, idx int) {
	if ini == nil || v == (DFeReferenciado{}) {
		return
	}
	secao := secaoMono("DFeReferenciado", idx)
	ini.GravarString(secao, "chaveAcesso", v.ChaveAcesso)
	ini.GravarInteiro(secao, "nItem", v.NItem)
}

// LerINIISel le [ISNNN], o Imposto Seletivo do item. Porte de Ler_ISel.
//
// O CSTIS e lido como STRING, e nao como enum: o ACBr deixou a conversao
// comentada esperando a tabela oficial de CSTs do IS (Ler_ISel:~). O porte
// mantem a string pelo mesmo motivo -- inventar um enum aqui travaria
// codigo que a legislacao ainda nao publicou.
func LerINIISel(ini *pcn.INI, v *GIS, idx int) {
	if ini == nil || v == nil {
		return
	}
	secao := secaoMono("IS", idx)
	if !ini.SecaoExiste(secao) {
		return
	}

	v.CSTIS = ini.LerString(secao, "CSTIS", "")
	v.CClassTribIS = ini.LerString(secao, "cClassTribIS", "")
	v.VBCIS = ini.LerFloat(secao, "vBCIS", 0)
	v.PIS = ini.LerFloat(secao, "pIS", 0)
	v.AdRemIS = ini.LerFloat(secao, "adRemIS", 0)
	v.UTrib = ini.LerString(secao, "uTrib", "")
	v.QTrib = ini.LerFloat(secao, "qTrib", 0)
	v.VIS = ini.LerFloat(secao, "vIS", 0)
}

// LerINIISTot le [ISTot], o total do Imposto Seletivo. Porte de Ler_ISTot.
func LerINIISTot(ini *pcn.INI, v *ISTot) {
	if ini == nil || v == nil {
		return
	}
	if !ini.SecaoExiste("ISTot") {
		return
	}
	v.VIS = ini.LerFloat("ISTot", "vIS", 0)
}

// LerINIGMonoTot le [gMono], o total da monofasia. Porte de Ler_gMonoTot.
func LerINIGMonoTot(ini *pcn.INI, v *GMono) {
	if ini == nil || v == nil {
		return
	}
	if !ini.SecaoExiste("gMono") {
		return
	}
	v.VIBSMono = ini.LerFloat("gMono", "vIBSMono", 0)
	v.VCBSMono = ini.LerFloat("gMono", "vCBSMono", 0)
	v.VIBSMonoReten = ini.LerFloat("gMono", "vIBSMonoReten", 0)
	v.VCBSMonoReten = ini.LerFloat("gMono", "vCBSMonoReten", 0)
	v.VIBSMonoRet = ini.LerFloat("gMono", "vIBSMonoRet", 0)
	v.VCBSMonoRet = ini.LerFloat("gMono", "vCBSMonoRet", 0)
}

// ---------------------------------------------------------------------------
// Escrita
// ---------------------------------------------------------------------------

// GravarINIGIBSCBSMono grava [gIBSCBSMonoNNN] e os quatro ramos.
// Porte de Gerar_gIBSCBSMono.
//
// O original NAO condiciona nada aqui: chamado, grava a secao e os quatro
// ramos inteiros, zeros inclusive. Quem decide se a monofasia sai e o
// `case` do CST em Gerar_IBSCBSNFe -- so o CST 620 chega ate aqui.
func GravarINIGIBSCBSMono(ini *pcn.INI, v GIBSCBSMono, idx int) {
	if ini == nil {
		return
	}
	secao := secaoMono("gIBSCBSMono", idx)
	ini.GravarFloat(secao, "vTotIBSMonoItem", v.VTotIBSMonoItem, 2)
	ini.GravarFloat(secao, "vTotCBSMonoItem", v.VTotCBSMonoItem, 2)

	gravarINIGIBSMonoAdRem(ini, v.GIBSMonoAdRem, idx)
	gravarINIGIBSMonoAdValorem(ini, v.GIBSMonoAdValorem, idx)
	gravarINIGCBSMonoAdRem(ini, v.GCBSMonoAdRem, idx)
	gravarINIGCBSMonoAdValorem(ini, v.GCBSMonoAdValorem, idx)
}

func gravarINIGIBSMonoAdRem(ini *pcn.INI, v GIBSMonoAdRem, idx int) {
	s := secaoMono("gMonoPadraoIBSQtde", idx)
	ini.GravarFloat(s, "qBCMono", v.GMonoPadrao.QBCMono, 4)
	ini.GravarFloat(s, "adRemIBS", v.GMonoPadrao.AdRemIBS, 4)
	ini.GravarFloat(s, "vIBSMono", v.GMonoPadrao.VIBSMono, 2)

	s = secaoMono("gMonoRetenIBSQtde", idx)
	ini.GravarFloat(s, "qBCMonoReten", v.GMonoReten.QBCMonoReten, 4)
	ini.GravarFloat(s, "adRemIBSReten", v.GMonoReten.AdRemIBSReten, 4)
	ini.GravarFloat(s, "vIBSMonoReten", v.GMonoReten.VIBSMonoReten, 2)

	s = secaoMono("gMonoRetIBSQtde", idx)
	ini.GravarFloat(s, "vIBSMonoRet", v.GMonoRet.VIBSMonoRet, 2)

	s = secaoMono("gpBioDiferencaIBS", idx)
	ini.GravarFloat(s, "qBCBioComb", v.GpBioDiferenca.QBCBioComb, 4)
	ini.GravarFloat(s, "vIBSDiferenca", v.GpBioDiferenca.VIBSDiferenca, 2)
}

func gravarINIGIBSMonoAdValorem(ini *pcn.INI, v GIBSMonoAdValorem, idx int) {
	s := secaoMono("gMonoPadraoIBSAliq", idx)
	ini.GravarFloat(s, "vBCMono", v.GMonoPadrao.VBCMono, 2)
	ini.GravarFloat(s, "pAliqMonoUF", v.GMonoPadrao.PAliqMonoUF, 4)
	ini.GravarFloat(s, "vIBSMonoUF", v.GMonoPadrao.VIBSMonoUF, 2)
	ini.GravarFloat(s, "pAliqMonoMun", v.GMonoPadrao.PAliqMonoMun, 4)
	ini.GravarFloat(s, "vIBSMonoMun", v.GMonoPadrao.VIBSMonoMun, 2)
	ini.GravarFloat(s, "vIBSMono", v.GMonoPadrao.VIBSMono, 2)

	s = secaoMono("gMonoRetenIBSAliq", idx)
	ini.GravarFloat(s, "vBCMonoReten", v.GMonoReten.VBCMonoReten, 2)
	ini.GravarFloat(s, "pAliqMonoReten", v.GMonoReten.PAliqMonoReten, 4)
	ini.GravarFloat(s, "vIBSMonoReten", v.GMonoReten.VIBSMonoReten, 2)

	s = secaoMono("gMonoRetIBSAliq", idx)
	ini.GravarFloat(s, "vIBSMonoRet", v.GMonoRet.VIBSMonoRet, 2)

	// O gpBioDiferenca do IBS tem UMA secao, compartilhada pelos dois
	// ramos. Gravado nos dois, como no original: o segundo sobrescreve o
	// primeiro com o mesmo valor quando os dois estao preenchidos.
	s = secaoMono("gpBioDiferencaIBS", idx)
	ini.GravarFloat(s, "qBCBioComb", v.GpBioDiferenca.QBCBioComb, 4)
	ini.GravarFloat(s, "vIBSDiferenca", v.GpBioDiferenca.VIBSDiferenca, 2)
}

func gravarINIGCBSMonoAdRem(ini *pcn.INI, v GCBSMonoAdRem, idx int) {
	s := secaoMono("gMonoPadraoCBSQtde", idx)
	ini.GravarFloat(s, "qBCMono", v.GMonoPadrao.QBCMono, 4)
	ini.GravarFloat(s, "adRemCBS", v.GMonoPadrao.AdRemCBS, 4)
	ini.GravarFloat(s, "vCBSMono", v.GMonoPadrao.VCBSMono, 2)

	s = secaoMono("gMonoRetenCBSQtde", idx)
	ini.GravarFloat(s, "qBCMonoReten", v.GMonoReten.QBCMonoReten, 4)
	ini.GravarFloat(s, "adRemCBSReten", v.GMonoReten.AdRemCBSReten, 4)
	ini.GravarFloat(s, "vCBSMonoReten", v.GMonoReten.VCBSMonoReten, 2)

	s = secaoMono("gMonoRetCBSQtde", idx)
	ini.GravarFloat(s, "vCBSMonoRet", v.GMonoRet.VCBSMonoRet, 2)

	gravarINIGpBioDiferencaCBS(ini, v.GpBioDiferenca, idx)
}

// gravarINIGpBioDiferencaCBS grava [gpBioDiferencaCBSNNN].
//
// DIVERGENCIA: o Gerar_gpBioDiferencaCBS do ACBr grava a chave com o nome do
// IBS (`RTC.IniWriter.pas:896`):
//
//	AINIRec.WriteFloat(sSecao, 'vIBSDiferenca', gpBioDiferenca.vCBSDiferenca);
//
// enquanto o Ler_gpBioDiferencaCBS le `vCBSDiferenca`. E copia-e-cola do
// grupo do IBS logo acima, e o efeito e que o ACBr PERDE o campo na propria
// ida e volta.
//
// Aqui a chave e `vCBSDiferenca` nos dois lados. A consequencia precisa ser
// sabida: um .ini gerado por este porte e lido pelo ACBr devolve o campo
// (o leitor dele procura `vCBSDiferenca`), mas um .ini gerado pelo ACBr e
// lido aqui NAO devolve -- a chave que ele gravou tem outro nome. Corrigir
// o lado que le seria pior: passaria a aceitar a chave errada como se fosse
// contrato.
func gravarINIGpBioDiferencaCBS(ini *pcn.INI, v GpBioDiferencaCBS, idx int) {
	s := secaoMono("gpBioDiferencaCBS", idx)
	ini.GravarFloat(s, "qBCBioComb", v.QBCBioComb, 4)
	ini.GravarFloat(s, "vCBSDiferenca", v.VCBSDiferenca, 2)
}

func gravarINIGCBSMonoAdValorem(ini *pcn.INI, v GCBSMonoAdValorem, idx int) {
	s := secaoMono("gMonoPadraoCBSAliq", idx)
	ini.GravarFloat(s, "vBCMono", v.GMonoPadrao.VBCMono, 2)
	ini.GravarFloat(s, "pAliqMonoCBS", v.GMonoPadrao.PAliqMonoCBS, 4)
	ini.GravarFloat(s, "vCBSMono", v.GMonoPadrao.VCBSMono, 2)

	s = secaoMono("gMonoRetenCBSAliq", idx)
	ini.GravarFloat(s, "vBCMonoReten", v.GMonoReten.VBCMonoReten, 2)
	ini.GravarFloat(s, "pAliqMonoReten", v.GMonoReten.PAliqMonoReten, 4)
	ini.GravarFloat(s, "vCBSMonoReten", v.GMonoReten.VCBSMonoReten, 2)

	s = secaoMono("gMonoRetCBSAliq", idx)
	ini.GravarFloat(s, "vCBSMonoRet", v.GMonoRet.VCBSMonoRet, 2)

	gravarINIGpBioDiferencaCBS(ini, v.GpBioDiferenca, idx)
}

// GravarINIISel grava [ISNNN]. Porte de Gerar_ISel.
//
// O original grava a secao SEM condicao nenhuma, zeros inclusive -- quem
// decide e o chamador. O porte acrescenta a guarda de grupo vazio, pela
// mesma razao dos demais grupos do .ini: um item sem Imposto Seletivo nao
// deve deixar uma secao de oito zeros no arquivo.
func GravarINIISel(ini *pcn.INI, v GIS, idx int) {
	if ini == nil || v == (GIS{}) {
		return
	}
	secao := secaoMono("IS", idx)
	ini.GravarString(secao, "CSTIS", v.CSTIS)
	ini.GravarString(secao, "cClassTribIS", v.CClassTribIS)
	ini.GravarFloat(secao, "vBCIS", v.VBCIS, 2)
	ini.GravarFloat(secao, "pIS", v.PIS, 4)
	ini.GravarFloat(secao, "adRemIS", v.AdRemIS, 4)
	ini.GravarString(secao, "uTrib", v.UTrib)
	ini.GravarFloat(secao, "qTrib", v.QTrib, 4)
	ini.GravarFloat(secao, "vIS", v.VIS, 2)
}

// GravarINIISTot grava [ISTot]. Porte de Gerar_ISTot.
//
// So sai com vIS POSITIVO -- e a condicao do original, e e o que distingue
// este total dos demais.
func GravarINIISTot(ini *pcn.INI, v ISTot) {
	if ini == nil || v.VIS <= 0 {
		return
	}
	ini.GravarFloat("ISTot", "vIS", v.VIS, 2)
}

// GravarINIGMonoTot grava [gMono]. Porte de Gerar_gMonoTot.
func GravarINIGMonoTot(ini *pcn.INI, v GMono) {
	if ini == nil || v == (GMono{}) {
		return
	}
	ini.GravarFloat("gMono", "vIBSMono", v.VIBSMono, 2)
	ini.GravarFloat("gMono", "vCBSMono", v.VCBSMono, 2)
	ini.GravarFloat("gMono", "vIBSMonoReten", v.VIBSMonoReten, 2)
	ini.GravarFloat("gMono", "vCBSMonoReten", v.VCBSMonoReten, 2)
	ini.GravarFloat("gMono", "vIBSMonoRet", v.VIBSMonoRet, 2)
	ini.GravarFloat("gMono", "vCBSMonoRet", v.VCBSMonoRet, 2)
}
