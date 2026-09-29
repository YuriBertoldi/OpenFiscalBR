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

// Constantes do documento. Porte de ACBrNFGas.Consts.pas e
// ACBrNFGas.pas.

const (
	// Namespace do XML da NFGas.
	// Porte de ACBRNFGAS_NAMESPACE / NAME_SPACE_NFGAS.
	Namespace = "http://www.portalfiscal.inf.br/nfgas"

	// ModeloNFGas e o modelo do documento fiscal.
	// Porte de MODELO_NFGAS (ACBrNFGas.IniReader.pas).
	ModeloNFGas = 76

	// NomeModeloDFe e o nome do modelo usado para localizar schemas e
	// servicos. Porte de TACBrNFGas.GetNomeModeloDFe.
	NomeModeloDFe = "NFGas"

	// LiteralChave e o prefixo do atributo Id de infNFGas.
	LiteralChave = "NFGas"

	// TamanhoChaveAcesso e o numero de posicoes da chave.
	TamanhoChaveAcesso = 44
)

// Elementos raiz aceitos pelo leitor.
const (
	// TagNFGas e o elemento do documento avulso.
	TagNFGas = "NFGas"
	// TagNFGasProc e o elemento do documento com protocolo. Atencao a
	// caixa: o XSD declara "nfgasProc" com n minusculo.
	TagNFGasProc = "nfgasProc"
	// TagProtNFGas e o elemento do protocolo dentro de nfgasProc.
	TagProtNFGas = "protNFGas"
	// TagInfNFGas e o elemento que carrega os dados do documento.
	TagInfNFGas = "infNFGas"
	// TagInfNFGasSupl carrega o QR-Code.
	TagInfNFGasSupl = "infNFGasSupl"
)

// gruposICMS e a ordem EXATA em que o leitor procura o grupo de ICMS no
// elemento imposto. O primeiro que existir alimenta a struct ICMS inteira.
//
// Duas observacoes sobre fidelidade, ambas deliberadas:
//
//   - ICMS41 e procurado pelo ACBr e NAO existe no XSD da NFGas;
//   - ICMS30, ICMS50 e ICMS90 parcial nao sao procurados.
//
// Porte de Ler_ICMS (ACBrNFGas.XmlReader.pas).
var gruposICMS = []string{
	"ICMS00", "ICMS10", "ICMS20", "ICMS40", "ICMS41",
	"ICMS51", "ICMS60", "ICMS70", "ICMS90",
}

// cStat que indicam documento confirmado, processado ou cancelado.
// Porte de TACBrNFGas.CstatConfirmada / CstatProcessado / CstatCancelada
// (ACBrNFGas.pas:230-256). Na NFGas, Confirmada e Processado sao o MESMO
// conjunto {100, 150} -- diferente do NFe, que acrescenta 110/301/302 -- e
// Cancelada inclui o 135 (evento registrado e vinculado).
var (
	cStatConfirmada = map[int]bool{100: true, 150: true}
	cStatProcessado = map[int]bool{100: true, 150: true}
	cStatCancelada  = map[int]bool{101: true, 135: true, 151: true, 155: true}
)

// cStatComProtocolo lista os cStat em que a consulta de situacao traz o
// grupo protNFGas. Porte de TRetConsSitNFGas.LerXml.
var cStatComProtocolo = map[int]bool{
	100: true, 101: true, 104: true, 150: true, 151: true, 155: true,
}
