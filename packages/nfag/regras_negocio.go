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

import (
	"strconv"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Regras de negocio da NFAg.
// Porte de ACBrNFAg.ValidarRegrasdeNegocio.pas.
//
// O ACBr declara 13 familias de validacao e implementa apenas quatro
// regras, todas em ValidacoesGerais: 226, 227, 247 e 252. As outras doze
// funcoes estao vazias no fonte, com o comentario "Implementar a validacao
// de todas as regras", e as regras 253, 415, 416, 417, 418, 419 e 421 estao
// comentadas. O porte reproduz exatamente esse estado -- acrescentar regra
// que o ACBr nao aplica faria o Go rejeitar documento que o Delphi aceita.

// ValidarRegrasNegocio aplica as regras de negocio sobre o documento e
// devolve as rejeicoes encontradas. Fatia vazia significa documento valido.
// Porte de TNFAgValidarRegras.Validar.
func ValidarRegrasNegocio(n *NFAg, cfg Configuracoes) []*ErroRegraNegocio {
	if n == nil {
		return nil
	}

	var erros []*ErroRegraNegocio

	if e := validarRegra226(n, cfg); e != nil {
		erros = append(erros, e)
	}
	if e := validarRegra227(n); e != nil {
		erros = append(erros, e)
	}
	if e := validarRegra247(n, cfg); e != nil {
		erros = append(erros, e)
	}
	if e := validarRegra252(n, cfg); e != nil {
		erros = append(erros, e)
	}

	// As familias abaixo existem no ACBr e estao VAZIAS no fonte. Ficam
	// registradas aqui para que a lacuna seja visivel a quem for implementar,
	// em vez de parecer que a validacao esta completa:
	//
	//   ValidacoesDoEmitente, ValidacoesDoDestinatario, ValidacoesDosItens,
	//   ValidacoesDosTributos, ValidacoesDeComprasGovernamentais,
	//   ValidacoesDosTotais, ValidacoesDaFatura, ValidacoesDaAgencia,
	//   ValidacoesDosAutorizadosAoXML, ValidacoesDoQRCode,
	//   ValidacoesDoResponsavelTecnico, ValidacoesAdicionais

	return erros
}

// validarRegra226 confere se os dois primeiros digitos do codigo IBGE do
// municipio do emitente batem com o codigo da UF autorizadora.
// Porte de ValidarRegra226.
func validarRegra226(n *NFAg, cfg Configuracoes) *ErroRegraNegocio {
	cod := cfg.CodigoUFEfetivo()
	if cod == 0 {
		// Sem UF autorizadora configurada nao ha o que comparar. O ACBr
		// compararia contra "0" e rejeitaria sempre; aqui a regra fica em
		// silencio, que e o comportamento util para quem so importa XML.
		return nil
	}
	mun := strconv.Itoa(n.Emit.EnderEmit.CMun)
	if len(mun) < 2 || mun[:2] != strconv.Itoa(cod) {
		return &ErroRegraNegocio{
			Codigo:   226,
			Mensagem: "Rejeicao: Codigo da UF do Emitente diverge da UF autorizadora",
		}
	}
	return nil
}

// validarRegra227 confere a chave de acesso contra a concatenacao dos
// campos correspondentes. Porte de ValidarRegra227.
func validarRegra227(n *NFAg) *ErroRegraNegocio {
	if ValidarConcatChave(n) {
		return nil
	}
	return &ErroRegraNegocio{
		Codigo:   227,
		Mensagem: "Rejeicao: Chave de Acesso do Campo Id difere da concatenacao dos campos correspondentes",
	}
}

// validarRegra247 confere a sigla da UF do emitente contra a do web
// service. Porte de ValidarRegra247.
func validarRegra247(n *NFAg, cfg Configuracoes) *ErroRegraNegocio {
	if cfg.UF == "" {
		// DIVERGENCIA: o ACBr compara mesmo com a UF vazia e rejeitaria
		// todo documento. Sem UF autorizadora configurada -- o caso de quem
		// so importa XML -- a regra fica em silencio, como a 226.
		return nil
	}
	if n.Emit.EnderEmit.UF != cfg.UF {
		return &ErroRegraNegocio{
			Codigo:   247,
			Mensagem: "Rejeicao: Sigla da UF do Emitente difere da UF do Web Service",
		}
	}
	return nil
}

// validarRegra252 confere o ambiente do documento contra o configurado.
// Porte de ValidarRegra252.
//
// A mensagem do ACBr fala em "MDF-e" -- texto herdado de outro componente.
// Aqui ela foi corrigida para NFAg, por ser texto de exibicao e nao
// contrato: o codigo 252 e o que identifica a rejeicao.
func validarRegra252(n *NFAg, cfg Configuracoes) *ErroRegraNegocio {
	if n.Ide.TpAmb != cfg.Ambiente {
		return &ErroRegraNegocio{
			Codigo:   252,
			Mensagem: "Rejeicao: Tipo do ambiente da NFAg difere do ambiente do Web Service",
		}
	}
	return nil
}

// ValidarConcatChave confere se a chave de acesso do atributo Id bate com a
// concatenacao dos campos do documento.
// Porte de TNFAgValidarRegras.ValidarConcatChave.
//
// O layout da chave da NFAg NAO e o classico da NF-e: a posicao 36 e o
// site autorizador, e cNF ocupa as posicoes 37 a 43.
//
//	pos  1- 2 (2)  cUF
//	pos  3- 4 (2)  ano  de dhEmi, dois digitos
//	pos  5- 6 (2)  mes  de dhEmi
//	pos  7-20 (14) CNPJ do emitente, com zeros a esquerda
//	pos 21-22 (2)  modelo (75)
//	pos 23-25 (3)  serie
//	pos 26-34 (9)  nNF
//	pos 35    (1)  tpEmis
//	pos 36    (1)  nSiteAutoriz
//	pos 37-43 (7)  cNF
//	pos 44    (1)  digito verificador -- NAO conferido por esta funcao
func ValidarConcatChave(n *NFAg) bool {
	if n == nil {
		return false
	}
	chave := pcn.RemoverLiteralChave(n.InfNFAg.ID)
	if len(chave) < 43 {
		return false
	}

	ano := n.Ide.DhEmi.Year()
	mes := int(n.Ide.DhEmi.Month())

	if chave[0:2] != pcn.FormatarInteiroZeros(n.Ide.CUF, 2) {
		return false
	}
	if chave[2:4] != pcn.FormatarInteiroZeros(ano, 4)[2:4] {
		return false
	}
	if chave[4:6] != pcn.FormatarInteiroZeros(mes, 2) {
		return false
	}
	if chave[6:20] != pcn.PreencherZerosEsquerda(pcn.OnlyCPFCNPJAlphaNum(n.Emit.CNPJ), 14) {
		return false
	}
	if chave[20:22] != pcn.FormatarInteiroZeros(n.Ide.Modelo, 2) {
		return false
	}
	if chave[22:25] != pcn.FormatarInteiroZeros(n.Ide.Serie, 3) {
		return false
	}
	if chave[25:34] != pcn.FormatarInteiroZeros(n.Ide.NNF, 9) {
		return false
	}
	if chave[34:35] != n.Ide.TpEmis.String() {
		return false
	}
	if chave[35:36] != n.Ide.NSiteAutoriz.String() {
		return false
	}
	if chave[36:43] != pcn.FormatarInteiroZeros(n.Ide.CNF, 7) {
		return false
	}
	return true
}

// MontarChaveAcesso monta a chave de 44 posicoes a partir dos campos do
// documento, ja com o digito verificador.
//
// Nao tem equivalente no ACBr -- la a chave e montada dentro do gerador de
// XML. Existe aqui para conferir a chave de um documento importado sem
// depender da emissao.
func MontarChaveAcesso(n *NFAg) (string, error) {
	if n == nil {
		return "", ErrXMLVazio
	}

	base := pcn.FormatarInteiroZeros(n.Ide.CUF, 2) +
		pcn.FormatarInteiroZeros(n.Ide.DhEmi.Year(), 4)[2:4] +
		pcn.FormatarInteiroZeros(int(n.Ide.DhEmi.Month()), 2) +
		pcn.PreencherZerosEsquerda(pcn.OnlyCPFCNPJAlphaNum(n.Emit.CNPJ), 14) +
		pcn.FormatarInteiroZeros(n.Ide.Modelo, 2) +
		pcn.FormatarInteiroZeros(n.Ide.Serie, 3) +
		pcn.FormatarInteiroZeros(n.Ide.NNF, 9) +
		n.Ide.TpEmis.String() +
		n.Ide.NSiteAutoriz.String() +
		pcn.FormatarInteiroZeros(n.Ide.CNF, 7)

	dv, err := pcn.DigitoChaveAcesso(base)
	if err != nil {
		return "", err
	}
	return base + strconv.Itoa(dv), nil
}

// ValidarChaveAcesso confere tamanho, digito verificador, codigo de UF e
// AAMM da chave do documento.
func ValidarChaveAcesso(n *NFAg) error {
	if n == nil {
		return ErrXMLVazio
	}
	return pcn.ValidarChaveAcesso(n.InfNFAg.ID)
}
