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

package pcn

import (
	"errors"
	"fmt"
)

// Erros sentinela do package. Use errors.Is para testar.
var (
	// ErrXMLVazio indica que nada foi fornecido ao leitor.
	ErrXMLVazio = errors.New("pcn: XML vazio")
	// ErrXMLInvalido indica XML mal formado.
	ErrXMLInvalido = errors.New("pcn: XML invalido")
	// ErrNoAusente indica um elemento obrigatorio que nao existe no documento.
	ErrNoAusente = errors.New("pcn: no ausente")
	// ErrAtributoAusente indica um atributo obrigatorio que nao existe no elemento.
	ErrAtributoAusente = errors.New("pcn: atributo ausente")
	// ErrChaveInvalida indica chave de acesso fora do formato de 44 digitos.
	ErrChaveInvalida = errors.New("pcn: chave de acesso invalida")
	// ErrDocumentoInvalido indica CNPJ/CPF que nao passa no digito verificador.
	ErrDocumentoInvalido = errors.New("pcn: documento invalido")
	// ErrSecaoAusente indica secao inexistente em arquivo INI.
	ErrSecaoAusente = errors.New("pcn: secao ausente")
)

// ErroLeitura carrega o caminho da tag em que a leitura falhou. Numa
// importacao em lote, saber qual campo de qual documento falhou e o que
// torna o erro util.
type ErroLeitura struct {
	Caminho string // ex.: "infNFGas/ide/dhEmi"
	Valor   string // conteudo bruto que provocou a falha
	Err     error
}

func (e *ErroLeitura) Error() string {
	if e.Valor == "" {
		return fmt.Sprintf("pcn: erro ao ler %s: %v", e.Caminho, e.Err)
	}
	return fmt.Sprintf("pcn: erro ao ler %s (valor %q): %v", e.Caminho, e.Valor, e.Err)
}

// Unwrap expoe o erro subjacente para errors.Is/errors.As.
func (e *ErroLeitura) Unwrap() error { return e.Err }

// NovoErroLeitura monta um ErroLeitura para o caminho e valor informados.
func NovoErroLeitura(caminho, valor string, err error) *ErroLeitura {
	return &ErroLeitura{Caminho: caminho, Valor: valor, Err: err}
}
