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

package pcn

import "strings"

// ===========================================================================
// SchemaDFe (TSchemaDFe) -- atributo schema do docZip
// ===========================================================================

// SchemaDFe diz que tipo de documento vem dentro de um <docZip> da
// Distribuicao DF-e. Porte de TSchemaDFe (ACBrDFe.Conversao.pas:223).
//
// Vive no pcn por ser compartilhado: o mesmo lote de distribuicao mistura
// documentos de NF-e, CT-e, MDF-e, BP-e e GTV-e, e o leitor de qualquer um
// deles precisa saber nomear o que encontrou.
type SchemaDFe int

const (
	// SchemaDesconhecido NAO existe no ACBr. Ver ParseSchemaDFe para a
	// razao de ele existir aqui.
	SchemaDesconhecido SchemaDFe = iota

	SchResNFe         // resNFe -- resumo da NF-e
	SchResEvento      // resEvento -- resumo de evento
	SchProcNFe        // procNFe -- NF-e processada (nfeProc)
	SchProcEventoNFe  // procEventoNFe -- evento processado
	SchResCTe         // resCTe
	SchProcCTe        // procCTe
	SchProcCTeOS      // procCTeOS
	SchProcEventoCTe  // procEventoCTe
	SchResMDFe        // resMDFe
	SchProcMDFe       // procMDFe
	SchProcEventoMDFe // procEventoMDFe
	SchResBPe         // resBPe
	SchProcBPe        // procBPe
	SchProcEventoBPe  // procEventoBPe
	SchProcGTVe       // procGTVe
	SchProcCTeSimp    // procCTeSimp

	// _fimSchemaDFe conta os membros declarados; nao e valor do leiaute.
	_fimSchemaDFe
)

var schemaDFeNomes = [...]string{
	"", // o desconhecido nao tem nome de schema
	"resNFe", "resEvento", "procNFe", "procEventoNFe",
	"resCTe", "procCTe", "procCTeOS", "procEventoCTe",
	"resMDFe", "procMDFe", "procEventoMDFe",
	"resBPe", "procBPe", "procEventoBPe",
	"procGTVe", "procCTeSimp",
}

// String devolve o nome do schema sem o sufixo de versao.
// Porte de SchemaDFeToStr, que remove o prefixo "sch" do nome do enum.
func (v SchemaDFe) String() string {
	if int(v) < 0 || int(v) >= len(schemaDFeNomes) {
		return ""
	}
	return schemaDFeNomes[v]
}

// EDocumento informa se o schema traz o documento COMPLETO, e nao o resumo.
//
// A distincao governa o que o consumidor pode fazer com o item: do resumo
// saem chave, emitente e valor; do documento completo sai o XML que se
// arquiva.
func (v SchemaDFe) EDocumento() bool {
	switch v {
	case SchProcNFe, SchProcEventoNFe, SchProcCTe, SchProcCTeOS,
		SchProcEventoCTe, SchProcMDFe, SchProcEventoMDFe, SchProcBPe,
		SchProcEventoBPe, SchProcGTVe, SchProcCTeSimp:
		return true
	}
	return false
}

// ParseSchemaDFe converte o atributo `schema` do docZip no enum.
//
// O atributo vem com a versao colada -- "resNFe_v1.01" --, e o ACBr corta
// no primeiro "_" antes de resolver (ACBrDFe.Conversao.pas:1156).
//
// DIVERGENCIA: o original levanta excecao quando o nome nao esta no enum.
// Num lote de cinquenta documentos isso derruba a leitura inteira porque um
// deles trouxe um schema novo -- e a SEFAZ acrescenta schema sem avisar.
// Aqui o caso devolve SchemaDesconhecido junto com o erro, para que o
// chamador possa seguir com os quarenta e nove restantes.
func ParseSchemaDFe(s string) (SchemaDFe, error) {
	nome := s
	if i := strings.Index(nome, "_"); i >= 0 {
		nome = nome[:i]
	}
	nome = strings.TrimPrefix(nome, "sch")

	for i, n := range schemaDFeNomes {
		if n != "" && n == nome {
			return SchemaDFe(i), nil
		}
	}
	return SchemaDesconhecido, erroEnum("SchemaDFe", s)
}
