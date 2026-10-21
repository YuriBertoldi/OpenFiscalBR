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

	// ErrDocumentoNulo indica chamada de geracao sem documento. O sentinela
	// do leitor (ErrXMLVazio) nao serve: ali o caso e XML ausente, aqui e
	// *NFe nil.
	ErrDocumentoNulo = errors.New("nfe: documento nao informado")

	// ErrNaoImplementado marca fase do porte ainda nao entregue.
	ErrNaoImplementado = errors.New("nfe: nao implementado")

	// ErrSemURL indica que a tabela de URLs nao tem entrada para a UF, o
	// modelo ou o servico pedido. No ACBr o caso devolve string vazia em
	// silencio, e o QR-Code sai sem o endereco da SEFAZ -- aqui vira erro.
	ErrSemURL = errors.New("nfe: URL nao encontrada para a UF")

	// ErrCertificadoObrigatorio indica QR-Code versao 3 em contingencia
	// offline sem funcao de assinatura. A versao 3 fecha com assinatura RSA,
	// nao com hash, e sem o certificado nao ha o que gerar.
	ErrCertificadoObrigatorio = errors.New("nfe: QR-Code versao 3 offline exige o certificado")

	// ErrCertificadoNaoConfigurado indica tentativa de transmitir sem
	// certificado. Os web services da NF-e exigem autenticacao mutua: sem
	// ele nao ha conexao, nem para consultar.
	ErrCertificadoNaoConfigurado = errors.New("nfe: certificado nao configurado")

	// ErrLoteVazio indica envio sem nota nenhuma.
	ErrLoteVazio = errors.New("nfe: lote sem notas")

	// ErrLoteExcedido indica lote acima do permitido -- em quantidade de
	// notas ou em tamanho da mensagem. Porte dos tres GerarException de
	// TNFeRecepcao (ACBrNFeWebServices.pas:1013, :1021 e :1160).
	ErrLoteExcedido = errors.New("nfe: lote excede o limite")

	// ErrAmbienteDiferente indica nota emitida num ambiente e transmitida
	// para outro. Porte do CErroAmbienteDiferente (:1053): e rejeicao certa
	// na SEFAZ, e o componente a antecipa.
	ErrAmbienteDiferente = errors.New("nfe: ambiente da nota difere do ambiente do web service")

	// ErrReciboVazio indica consulta de recibo sem o numero.
	ErrReciboVazio = errors.New("nfe: recibo nao informado")

	// ErrJustificativaObrigatoria indica inutilizacao sem justificativa.
	// Porte do primeiro GerarException de SetJustificativa (:2630).
	ErrJustificativaObrigatoria = errors.New("nfe: justificativa obrigatoria na inutilizacao")

	// ErrJustificativaCurta indica justificativa abaixo do minimo legal de
	// quinze caracteres (:2634).
	ErrJustificativaCurta = errors.New("nfe: justificativa abaixo do minimo de 15 caracteres")

	// ErrModeloInvalido indica `mod` que nao e 55 nem 65. Porte do raise de
	// TNFeInutilizacao.DefinirURL (:2668).
	ErrModeloInvalido = errors.New("nfe: modelo invalido")

	// ErrDocumentoConsultaAusente indica consulta de cadastro sem IE, CNPJ
	// nem CPF. O ACBr nao confere e manda um infCons com <CPF></CPF> vazio;
	// aqui o caso vira erro, porque a chamada nunca pode dar certo.
	ErrDocumentoConsultaAusente = errors.New("nfe: informe IE, CNPJ ou CPF na consulta de cadastro")
)
