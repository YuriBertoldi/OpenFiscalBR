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

package comum

// Component is the base component type, equivalent to TACBrComponent from
// ACBrBase.pas. In Go the component hierarchy is flat; Component serves as
// an embeddable struct that higher-level components (e.g. SPED) can include.
type Component struct{}

// NewComponent creates a new Component instance.
func NewComponent() *Component {
	return &Component{}
}

// ACBrError is the standard error type for the OpenFiscalBR library,
// equivalent to EACBrException in Delphi.
type ACBrError struct {
	Message string
}

// Error implements the error interface.
func (e *ACBrError) Error() string {
	return e.Message
}

// NewACBrError creates a new ACBrError with the given message.
func NewACBrError(msg string) *ACBrError {
	return &ACBrError{Message: msg}
}
