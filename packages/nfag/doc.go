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

// Package nfag implements the NFAg -- Nota Fiscal de Fornecimento de Agua
// Canalizada Eletronica, modelo 75, leiaute 1.00.
//
// It is a Go port of the ACBr component ACBrNFAg, covering both READING
// (XML single/batch, events, consultation results, the ACBr .ini format)
// and ISSUING (XML generation, XMLDSig signature via packages/dfe, and the
// SEFAZ SOAP web services: synchronous reception, consultation, service
// status and cancellation event).
//
// Uso tipico de importacao em lote:
//
//	notas, err := nfag.LerLoteArquivo("lote.xml")
//	for _, n := range notas {
//	    fmt.Println(n.ChaveAcesso(), n.NFAg.Total.VNF, n.Situacao())
//	}
//
// Os nomes dos campos sao os do leiaute da SEFAZ (cUF, nNF, vBC, qFaturada)
// e nao devem ser traduzidos: eles mapeiam 1:1 para as tags XML.
//
// O documento fiscal e o tipo NFAg; o componente equivalente a TACBrNFAg
// e o tipo Componente.
package nfag
