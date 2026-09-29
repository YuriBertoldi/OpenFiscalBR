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

// Package nfgas implements the NFGas -- Nota Fiscal de Fornecimento de Gas
// Canalizado Eletronica, modelo 76, leiaute 1.00.
//
// It is a Go port of the ACBr component ACBrNFGas. This phase covers
// READING only: XML (single document and batch), events, consultation
// results and the ACBr .ini exchange format. Issuing a document -- XML
// generation, digital signature and web services -- is declared but not
// implemented yet; those methods return ErrNaoImplementado.
//
// Uso tipico de importacao em lote:
//
//	notas, err := nfgas.LerLoteArquivo("lote.xml")
//	for _, n := range notas {
//	    fmt.Println(n.ChaveAcesso(), n.NFGas.Total.VNF, n.Situacao())
//	}
//
// Os nomes dos campos sao os do leiaute da SEFAZ (cUF, nNF, vBC, qFaturada)
// e nao devem ser traduzidos: eles mapeiam 1:1 para as tags XML.
//
// O documento fiscal e o tipo NFGas; o componente equivalente a TACBrNFGas
// e o tipo Componente.
package nfgas
