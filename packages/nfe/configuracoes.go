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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Configuracoes reune o que o COMPONENTE sabe e o documento nao.
//
// Porte parcial de TConfiguracoesNFe.Geral e .WebServices: por ora os
// campos que o QR-Code, as regras de negocio e os web services consomem.
// Os de arquivo (caminhos, salvar XML) nao entram: gravar em disco e papel
// do consumidor da biblioteca, nao dela.
//
// A distincao importa: tudo que esta aqui e configuracao do EMISSOR, nao do
// documento. O ACBr compara os dois em varias regras -- o ambiente da nota
// contra o ambiente do web service, por exemplo -- e e dessa comparacao que
// nascem algumas rejeicoes.
type Configuracoes struct {
	// VersaoDF e a versao do leiaute que o componente esta configurado para
	// emitir. NAO e a mesma coisa que InfNFe.Versao, que e a do documento.
	VersaoDF VersaoDF

	// VersaoQRCode e a versao pedida do QR-Code. O valor efetivo passa por
	// AjustarVersaoQrCode, que a rebaixa conforme a versao do leiaute.
	VersaoQRCode VersaoQrCode

	// ModeloDF decide, no urlChave, se a URL de consulta vem da tabela da
	// NF-e ou da NFC-e. E o modelo do COMPONENTE, nao o do documento -- o
	// ACBr le Configuracoes.Geral.ModeloDF ali (ACBrNFe.pas:545).
	ModeloDF ModeloDF

	// Ambiente e o ambiente do web service. A regra 252 rejeita a nota cujo
	// tpAmb difere deste.
	Ambiente pcn.TipoAmbiente

	// TpEmis e a forma de emissao configurada, no formato NUMERICO que o
	// validador do ACBr usa (TNFeValidarRegras.tpEmis: Integer). Nao e o
	// enum do documento: as regras 570 e 713 comparam o enum da nota com
	// este inteiro.
	TpEmis int

	// UF e CodigoUF sao a unidade federada do emissor, usadas pelas regras
	// que conferem coerencia entre a nota e o ambiente de autorizacao.
	UF       string
	CodigoUF int

	// CSC e IDCSC sao o Codigo de Seguranca do Contribuinte e seu
	// identificador, usados no hash do QR-Code das versoes 1 e 2. A versao 3
	// nao os usa -- ela assina com o certificado.
	//
	// IDCSC e string porque o ACBr o guarda assim e o converte de formas
	// DIFERENTES conforme a versao do QR-Code: zero-padded em seis posicoes
	// na versao 1, sem padding na 2.
	CSC   string
	IDCSC string

	// FormaEmissao e a contingencia configurada. Ela decide SOB QUAL
	// UNIDADE o endereco do web service e procurado: em SVC o documento
	// passa a ser autorizado pelo ambiente virtual, nao pela SEFAZ de
	// origem (ACBrNFe.pas:517-523).
	//
	// Nao se confunde com o Ide.TpEmis do documento, embora os dois devam
	// concordar -- e as regras 570 e 713 existem justamente para conferir
	// isso.
	FormaEmissao pcn.TipoEmissao

	// Certificado assina e autentica a conexao TLS com a SEFAZ. Nulo
	// impede qualquer transmissao: os web services da NF-e exigem
	// autenticacao mutua.
	Certificado *dfe.Certificado

	// Timeout limita cada chamada. Zero usa o padrao do cliente SOAP.
	Timeout time.Duration
}
