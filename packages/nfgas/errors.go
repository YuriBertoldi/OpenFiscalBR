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

package nfgas

import (
	"errors"
	"fmt"
	"strings"
)

// Erros sentinela do package. Use errors.Is para testar.
//
// Nenhuma funcao deste package entra em panico: o ACBr levanta
// EACBrNFGasException em varios pontos da leitura, e aqui tudo vira error.
// Numa importacao em lote, uma nota torta nao pode derrubar o processo.
var (
	// ErrEnumInvalido indica codigo que nao corresponde a nenhum membro do
	// enum. Corresponde aos raise dos StrToXxx do Delphi.
	ErrEnumInvalido = errors.New("nfgas: valor invalido para enum")

	// ErrXMLVazio indica que nada foi fornecido ao leitor.
	// Porte de 'Arquivo XML da NFGas nao carregado.'
	ErrXMLVazio = errors.New("nfgas: XML da NFGas nao carregado")

	// ErrXMLIncorreto indica XML sem o elemento infNFGas.
	// Porte de 'Arquivo xml incorreto.'
	ErrXMLIncorreto = errors.New("nfgas: arquivo XML incorreto, infNFGas nao encontrado")

	// ErrAtributoIDAusente indica infNFGas sem o atributo Id.
	// Porte de 'Nao encontrei o atributo: Id'
	ErrAtributoIDAusente = errors.New("nfgas: atributo Id nao encontrado em infNFGas")

	// ErrAtributoVersaoAusente indica infNFGas sem o atributo versao.
	// Porte de 'Nao encontrei o atributo: versao'
	ErrAtributoVersaoAusente = errors.New("nfgas: atributo versao nao encontrado em infNFGas")

	// ErrNFGasNaoEncontrada indica XML sem o elemento NFGas.
	// Porte de 'NFGas nao encontrada no XML'
	ErrNFGasNaoEncontrada = errors.New("nfgas: NFGas nao encontrada no XML")

	// ErrLoteVazio indica entrada sem nenhum documento reconhecivel.
	ErrLoteVazio = errors.New("nfgas: nenhuma NFGas encontrada no lote")

	// ErrNaoImplementado marca a parte de emissao, que sera portada num
	// segundo momento: geracao de XML, assinatura e web services.
	ErrNaoImplementado = errors.New("nfgas: nao implementado nesta fase (emissao)")

	// ErrProtocoloAusente indica operacao que exige o protocolo de
	// autorizacao (nfgasProc) sem ProcNFGas.NProt preenchido.
	ErrProtocoloAusente = errors.New("nfgas: documento sem protocolo de autorizacao")

	// ErrEventoAusente indica retorno de evento sem a parte enviada, de que
	// o procEventoNFGas nao pode prescindir (o xJust so existe la).
	ErrEventoAusente = errors.New("nfgas: retorno sem o evento enviado")

	// ErrSemURL indica UF sem web service de NFGas definido -- MA e PA
	// constam como SVAN no ACBrNFGasServicos.ini, mas o SVAN nao tem URLs
	// de NFGas publicadas (lacuna herdada do ACBr).
	ErrSemURL = errors.New("nfgas: a UF nao tem URL de web service NFGas definida")

	// ErrCertificadoObrigatorio indica operacao de transmissao sem
	// certificado configurado no Componente.
	ErrCertificadoObrigatorio = errors.New("nfgas: certificado digital nao configurado")

	// ErrDigestDivergente indica protocolo de autorizacao cujo digVal nao
	// corresponde ao DigestValue do XML transmitido (ValidarDigest do
	// TratarResposta).
	ErrDigestDivergente = errors.New("nfgas: digVal do protocolo nao corresponde ao XML transmitido")
)

// ErroNFGas e o erro de leitura de um documento, com o indice dentro do
// lote e a chave de acesso quando ja foi possivel extrai-la.
//
// Numa importacao de milhares de XMLs, saber QUAL documento falhou e o que
// torna o erro acionavel.
type ErroNFGas struct {
	Indice  int    // posicao no lote, base 0; -1 fora de lote
	Chave   string // chave de acesso, quando conhecida
	Arquivo string // nome do arquivo de origem, quando conhecido
	Err     error
}

func (e *ErroNFGas) Error() string {
	var b strings.Builder
	b.WriteString("nfgas: ")
	if e.Arquivo != "" {
		fmt.Fprintf(&b, "arquivo %s: ", e.Arquivo)
	}
	if e.Indice >= 0 {
		fmt.Fprintf(&b, "documento %d: ", e.Indice+1)
	}
	if e.Chave != "" {
		fmt.Fprintf(&b, "chave %s: ", e.Chave)
	}
	b.WriteString(e.Err.Error())
	return b.String()
}

// Unwrap expoe o erro subjacente para errors.Is/errors.As.
func (e *ErroNFGas) Unwrap() error { return e.Err }

// ErrosLote agrega as falhas de uma leitura em lote. As notas que puderam
// ser lidas continuam disponiveis -- uma nota torta nao invalida o lote.
type ErrosLote struct {
	Erros []*ErroNFGas
}

func (e *ErrosLote) Error() string {
	if e == nil || len(e.Erros) == 0 {
		return "nfgas: nenhum erro"
	}
	if len(e.Erros) == 1 {
		return e.Erros[0].Error()
	}
	return fmt.Sprintf("nfgas: %d documentos com erro; primeiro: %s", len(e.Erros), e.Erros[0].Error())
}

// Unwrap devolve os erros individuais, para errors.Is/errors.As.
func (e *ErrosLote) Unwrap() []error {
	if e == nil {
		return nil
	}
	r := make([]error, 0, len(e.Erros))
	for _, x := range e.Erros {
		r = append(r, x)
	}
	return r
}

// ErroRegraNegocio e uma rejeicao das regras de negocio, com o codigo
// oficial da SEFAZ. Porte das mensagens de
// ACBrNFGas.ValidarRegrasdeNegocio.pas.
type ErroRegraNegocio struct {
	Codigo   int    // ex.: 226
	Mensagem string // ex.: "Rejeicao: Codigo da UF do Emitente diverge da UF autorizadora"
}

func (e *ErroRegraNegocio) Error() string {
	return fmt.Sprintf("%d-%s", e.Codigo, e.Mensagem)
}
