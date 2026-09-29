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
// Bloco 9 - Controle e Encerramento do Arquivo Digital
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// Registro9001 - Abertura do Bloco 9
type Registro9001 struct {
	OpenBlocos
	Registro9900 []*Registro9900
}

// NewRegistro9001 cria um novo Registro9001 com IndDad=1 (sem dados).
func NewRegistro9001() *Registro9001 {
	return &Registro9001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// Registro9900 - Registros do Arquivo
type Registro9900 struct {
	RegBlc   string
	QtdRegBlc int
}

// Registro9990 - Encerramento do Bloco 9
type Registro9990 struct {
	QtdLin9 int
}

// Registro9999 - Encerramento do Arquivo Digital
type Registro9999 struct {
	QtdLin int
}
