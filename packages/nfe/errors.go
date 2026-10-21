// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import "errors"

// Erros do package nfe.
//
// Seguem a convencao oficial do projeto (CLAUDE.md, "Convencao de erros"):
// sentinelas testaveis com errors.Is, nada de panic em biblioteca. Cada
// `raise Exception` do TNFeXmlReader vira um destes.
var (
	// ErrXMLVazio indica entrada sem documento. Porte de
	// "Arquivo xml da Nfe nao carregado".
	ErrXMLVazio = errors.New("nfe: XML da NFe nao carregado")

	// ErrXMLIncorreto indica XML sem o grupo infNFe. Porte de
	// "Arquivo xml incorreto".
	ErrXMLIncorreto = errors.New("nfe: XML incorreto, grupo infNFe nao encontrado")

	// ErrNFeNaoEncontrada indica envelope sem o elemento NFe.
	ErrNFeNaoEncontrada = errors.New("nfe: elemento NFe nao encontrado")

	// ErrAtributoIDAusente indica infNFe sem o atributo Id. Porte de
	// "Nao encontrei o atributo: Id".
	ErrAtributoIDAusente = errors.New("nfe: atributo Id ausente em infNFe")

	// ErrAtributoVersaoAusente indica infNFe sem o atributo versao. Porte de
	// "Nao encontrei o atributo: versao".
	ErrAtributoVersaoAusente = errors.New("nfe: atributo versao ausente em infNFe")

	// ErrNaoImplementado marca fase do porte ainda nao entregue.
	ErrNaoImplementado = errors.New("nfe: nao implementado")
)
