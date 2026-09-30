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

package nfag

// Constantes do documento. Porte de ACBrNFAg.Consts.pas e ACBrNFAg.pas.

const (
	// Namespace do XML da NFAg.
	// Porte de ACBRNFAG_NAMESPACE / NAME_SPACE_NFAG.
	Namespace = "http://www.portalfiscal.inf.br/nfag"

	// ModeloNFAg e o modelo do documento fiscal.
	// Porte de MODELO_NFAg (ACBrNFAg.IniReader.pas).
	ModeloNFAg = 75

	// NomeModeloDFe e o nome do modelo usado para localizar schemas e
	// servicos. Porte de TACBrNFAg.GetNomeModeloDFe.
	NomeModeloDFe = "NFAg"

	// LiteralChave e o prefixo do atributo Id de infNFAg. E MAIUSCULO
	// ("NFAG" + chave): o XSD exige pattern NFAG[0-9]{6}[A-Z0-9]{12}[0-9]{26}
	// (nfagTiposBasico_v1.00.xsd) e o GerarXml do ACBr grava 'NFAG'.
	LiteralChave = "NFAG"

	// TagNFAg e o elemento do documento avulso.
	TagNFAg = "NFAg"
	// TagNFAgProc e o elemento do documento com protocolo. Atencao a
	// grafia: o XSD define nfagProc (minusculo); o ACBr gera e le
	// NFAgProc -- o leitor daqui aceita as duas.
	TagNFAgProc = "nfagProc"
)

// gruposICMS nao existe na NFAg: o imposto do item NAO tem grupo ICMS --
// so IBSCBS, PIS, COFINS, retTrib, TFS e TFU (ver classes.go).

// cStat que indicam documento confirmado, processado ou cancelado.
// Porte de TACBrNFAg.CstatConfirmada / CstatProcessado / CstatCancelada
// (ACBrNFAg.pas:228-253): Confirmada e Processado sao o MESMO conjunto
// {100, 150}, e Cancelada inclui o 135.
var (
	cStatConfirmada = map[int]bool{100: true, 150: true}
	cStatProcessado = map[int]bool{100: true, 150: true}
	cStatCancelada  = map[int]bool{101: true, 135: true, 151: true, 155: true}
)

// cStatComProtocolo lista os cStat em que a consulta de situacao traz o
// grupo protNFAg. Porte de TRetConsSitNFAg.LerXml.
var cStatComProtocolo = map[int]bool{
	100: true, 101: true, 104: true, 150: true, 151: true, 155: true,
}

// cStatRetornoSincronoOK e o cStat do RETORNO da recepcao sincrona que,
// junto com um protocolo processado, indica autorizacao -- na NFAg e 104
// ("Lote processado"), diferente da NFGas (100).
// Porte de TNFAgRecepcao.TratarResposta (ACBrNFAgWebServices.pas).
const cStatRetornoSincronoOK = 104
