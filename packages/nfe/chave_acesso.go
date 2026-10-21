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

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Chave de acesso da NF-e. Porte de GerarChaveAcesso e GerarCodigoDFe
// (ACBrDFeUtil.pas), que o TNFeXmlWriter.GerarXml chama antes de montar o
// documento.
//
// Mora aqui, e nao em regras_negocio.go como na NFGas, porque o gerador
// depende dela estruturalmente: GerarXml monta a chave e so entao grava
// infNFe.ID, ide.cDV e ide.cNF.

// Layout das 44 posicoes da chave da NF-e:
//
//	cUF(2) AAMM(4) CNPJ(14) mod(2) serie(3) nNF(9) tpEmis(1) cNF(8) cDV(1)
//
// Diferente da NFGas, que gasta a posicao 36 com o nSiteAutoriz e por isso
// tem cNF de 7 digitos, a NF-e usa os 8 digitos cheios.
const digitosCNF = 8

// MontarChaveAcesso monta a chave de 44 posicoes a partir dos campos do
// documento, ja com o digito verificador.
//
// Como o GerarChaveAcesso do ACBr, NORMALIZA o cNF antes de montar: 0 ou -1
// sorteiam um codigo novo e qualquer outro negativo vira zero. O cNF
// normalizado e gravado de volta no documento, porque a chave e o ide
// precisam concordar.
//
// O CNPJ usado e o do emitente, EXCETO quando a nota e avulsa: nesse caso o
// ACBr usa o CNPJ do fisco emitente (GerarXml, XmlWriter.pas:311-314).
func MontarChaveAcesso(n *NFe) (string, error) {
	if n == nil {
		return "", ErrDocumentoNulo
	}

	normalizarCNF(n)

	documento := n.Emit.CNPJCPF
	if strings.TrimSpace(n.Avulsa.CNPJ) != "" {
		documento = n.Avulsa.CNPJ
	}

	base := pcn.FormatarInteiroZeros(n.Ide.CUF, 2) +
		pcn.FormatarInteiroZeros(n.Ide.DEmi.Year(), 4)[2:4] +
		pcn.FormatarInteiroZeros(int(n.Ide.DEmi.Month()), 2) +
		pcn.PreencherZerosEsquerda(pcn.OnlyCPFCNPJAlphaNum(documento), 14) +
		pcn.FormatarInteiroZeros(n.Ide.Modelo, 2) +
		pcn.FormatarInteiroZeros(n.Ide.Serie, 3) +
		pcn.FormatarInteiroZeros(n.Ide.NNF, 9) +
		n.Ide.TpEmis.String() +
		pcn.FormatarInteiroZeros(n.Ide.CNF, digitosCNF)

	dv, err := pcn.DigitoChaveAcesso(base)
	if err != nil {
		return "", err
	}
	return base + strconv.Itoa(dv), nil
}

// ValidarChaveAcesso confere tamanho, digito verificador, codigo de UF e
// AAMM da chave gravada em infNFe.ID.
func ValidarChaveAcesso(n *NFe) error {
	if n == nil {
		return ErrDocumentoNulo
	}
	return pcn.ValidarChaveAcesso(n.InfNFe.ID)
}

// normalizarCNF aplica a regra do GerarChaveAcesso: 0 e -1 sorteiam, e
// qualquer outro negativo vira zero.
//
// O -2 e o sentinela que o LEITOR grava quando a tag cNF vem zerada, entao
// reler e regerar uma nota sem cNF produz zero -- nao um sorteio. E o que
// torna o round-trip estavel.
func normalizarCNF(n *NFe) {
	switch {
	case n.Ide.CNF == 0 || n.Ide.CNF == -1:
		n.Ide.CNF = gerarCodigoDFe(n.Ide.NNF)
	case n.Ide.CNF < 0:
		n.Ide.CNF = 0
	}
}

// gerarCodigoDFe sorteia o cNF, evitando os codigos que a SEFAZ recusa.
// Porte de GerarCodigoDFe (ACBrDFeUtil.pas).
func gerarCodigoDFe(nNF int) int {
	for {
		c := rand.Intn(100000000) // 8 digitos
		if validarCodigoDFe(c, nNF) {
			return c
		}
	}
}

// codigosDFeInvalidos e a lista de codigos que a SEFAZ rejeita como cNF:
// zero, repeticoes de um mesmo digito e sequencias.
var codigosDFeInvalidos = map[int]bool{
	0: true, 11111111: true, 22222222: true, 33333333: true, 44444444: true,
	55555555: true, 66666666: true, 77777777: true, 88888888: true,
	99999999: true, 12345678: true, 23456789: true, 34567890: true,
	45678901: true, 56789012: true, 67890123: true, 78901234: true,
	89012345: true, 90123456: true, 1234567: true,
}

// validarCodigoDFe recusa o codigo igual ao numero da nota e os da lista de
// invalidos -- as duas regras do ACBr.
func validarCodigoDFe(codigo, nNF int) bool {
	return codigo != nNF && !codigosDFeInvalidos[codigo]
}
