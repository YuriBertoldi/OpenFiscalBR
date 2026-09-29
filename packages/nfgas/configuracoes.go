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

import "github.com/openfiscalbr/openfiscalbr/packages/pcn"

// Configuracoes reune o que a leitura e a validacao precisam saber sobre o
// contexto. Porte do subconjunto de TConfiguracoesNFGas que a leitura usa.
//
// As configuracoes de certificado, proxy, caminhos de arquivo e web service
// ficam para a fase de emissao.
//
// O valor zero e utilizavel: VersaoDF vale Ve100 (unica versao publicada),
// Ambiente vale TaProducao e TpEmis vale TeNormal -- que sao os zeros dos
// enums. Atencao: o default do componente DELPHI para Ambiente e
// taHomologacao (ACBrDFeConfiguracoes.pas); NovoComponente aplica esse
// default para manter a paridade.
type Configuracoes struct {
	// VersaoDF e a versao do leiaute. Porte de Geral.VersaoDF.
	VersaoDF VersaoNFGas
	// VersaoQRCode e a versao do QR-Code. Porte de Geral.VersaoQRCode.
	VersaoQRCode VersaoQrCode
	// Ambiente e o ambiente esperado. Porte de WebServices.Ambiente.
	// Usado pela regra de negocio 252 e como default de tpAmb no .ini.
	Ambiente pcn.TipoAmbiente
	// TpEmis e a forma de emissao configurada no componente. Porte de
	// Geral.FormaEmissao; usada como default de tpEmis na leitura do .ini
	// -- sem ela, uma instalacao em contingencia leria o .ini como emissao
	// normal.
	TpEmis pcn.TipoEmissao
	// UF e a sigla da UF autorizadora. Porte de WebServices.UF.
	// Usada pela regra de negocio 247.
	UF string
	// CodigoUF e o codigo IBGE da UF autorizadora. Usado pela regra 226.
	// Quando vale 0, e derivado de UF.
	CodigoUF int
}

// CodigoUFEfetivo devolve o codigo IBGE da UF autorizadora, derivando de UF
// quando CodigoUF nao foi informado.
func (c Configuracoes) CodigoUFEfetivo() int {
	if c.CodigoUF != 0 {
		return c.CodigoUF
	}
	return pcn.CodigoUF(c.UF)
}
