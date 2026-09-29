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

// Package rtc implements the Reforma Tributaria sobre o Consumo groups
// (IBS, CBS, Imposto Seletivo, compra governamental, pagamento vinculado)
// shared by every Brazilian electronic fiscal document.
//
// It is a Go port of ACBrDFe.RTC.Classes.pas and ACBrDFe.RTC.XmlReader.pas.
// The same tree is consumed by NFe, CTe, MDFe, BPe, NF3e, NFCom and NFGas,
// which is why it lives in its own package instead of being duplicated in
// each component.
//
// Os nomes dos campos sao os do leiaute (vBC, pIBSUF, vCredPresCondSus) e
// nao devem ser traduzidos: eles mapeiam 1:1 para as tags XML da SEFAZ.
//
// Todos os grupos sao structs por valor, nao ponteiros: no Delphi os
// construtores sempre instanciam os subgrupos, entao eles nunca sao nil. O
// valor zero de qualquer struct daqui ja e utilizavel.
package rtc
