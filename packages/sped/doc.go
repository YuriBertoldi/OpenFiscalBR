// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-28.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

// Package sped implementa o SPED (Sistema Publico de Escrituracao Digital)
// para geracao de arquivos texto no formato exigido pela Receita Federal.
//
// Sub-componentes suportados:
//   - SPED Fiscal (EFD-ICMS/IPI)
//   - SPED Contabil (ECD)
//   - SPED ECF (Escrituracao Contabil Fiscal)
//   - SPED PIS/COFINS (EFD-Contribuicoes)
package sped
