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

// Package pcn provides the XML reading foundation shared by every DFe
// component of OpenFiscalBR (NFe, CTe, MDFe, NF3e, NFCom, NFGas...).
//
// It is a Go port of the ACBr units ACBrXmlDocument.pas, ACBrXmlBase.pas,
// ACBrValidador.pas and the document-key helpers of ACBrDFeUtil.pas.
//
// The central abstraction is a small read-only DOM that mirrors
// TACBrXmlDocument/TACBrXmlNode, so that the imperative readers of each
// component can be ported line by line from their Delphi counterparts:
//
//	doc, err := pcn.ParseString(xml)
//	ide := doc.Root.FindAnyNs("infNFGas").FindAnyNs("ide")
//	cUF := pcn.ConteudoInt(ide.FindAnyNs("cUF"))
//
// Every navigation method is nil-safe: chaining through a missing element
// yields a nil node, and reading a nil node yields the zero value of the
// requested type. That reproduces the tolerant behaviour of the ACBr
// readers without the access violations of the original.
package pcn
