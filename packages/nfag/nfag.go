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
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Componente e o equivalente a TACBrNFAg: guarda a configuracao e a lista
// de documentos carregados.
//
// O documento fiscal em si e o tipo NFAg. Para so LER XML, o Componente e
// dispensavel -- use as funcoes de pacote LerXML, LerLote e afins. Ele
// existe para quem quer um ponto unico de configuracao e para receber, na
// fase de emissao, os metodos de web service.
//
// O valor zero e utilizavel.
type Componente struct {
	// Configuracoes reune versao de leiaute, ambiente e UF autorizadora.
	Configuracoes Configuracoes
	// NotasFiscais sao os documentos carregados.
	NotasFiscais []*NotaFiscal
	// Status e o estado interno da ultima operacao.
	Status StatusNFAg
}

// NovoComponente cria um Componente com os defaults do ACBr: leiaute 1.00,
// ambiente de HOMOLOGACAO (default de TConfiguracoesWebServices) e emissao
// normal.
func NovoComponente() *Componente {
	return &Componente{
		Configuracoes: Configuracoes{
			VersaoDF: Ve100,
			Ambiente: pcn.TaHomologacao,
			TpEmis:   pcn.TeNormal,
		},
	}
}

// NomeModelo devolve o nome do modelo usado para localizar schemas e
// servicos. Porte de TACBrNFAg.GetNomeModeloDFe.
func (c *Componente) NomeModelo() string { return NomeModeloDFe }

// NamespaceURI devolve o namespace do XML da NFAg.
// Porte de TACBrNFAg.GetNameSpaceURI.
func (c *Componente) NamespaceURI() string { return Namespace }

// CarregarArquivo le os documentos de um arquivo e ACRESCENTA a lista.
// Porte de TNotasFiscais.LoadFromFile.
func (c *Componente) CarregarArquivo(caminho string) error {
	notas, err := LerLoteArquivo(caminho)
	c.NotasFiscais = append(c.NotasFiscais, notas...)
	return err
}

// CarregarDiretorio le todos os .xml de um diretorio e ACRESCENTA a lista.
func (c *Componente) CarregarDiretorio(dir string) error {
	notas, err := LerLoteDiretorio(dir)
	c.NotasFiscais = append(c.NotasFiscais, notas...)
	return err
}

// CarregarString le os documentos de uma string e ACRESCENTA a lista.
// Porte de TNotasFiscais.LoadFromString.
func (c *Componente) CarregarString(xml string) error {
	notas, err := LerLoteString(xml)
	c.NotasFiscais = append(c.NotasFiscais, notas...)
	return err
}

// CarregarINI le um documento em formato .ini e ACRESCENTA a lista.
// Porte de TNotasFiscais.LoadFromIni.
func (c *Componente) CarregarINI(iniOuCaminho string) error {
	n, err := LerINI(iniOuCaminho, c.Configuracoes)
	if err != nil {
		return err
	}
	c.NotasFiscais = append(c.NotasFiscais, &NotaFiscal{NFAg: n})
	return nil
}

// Limpar descarta os documentos carregados.
func (c *Componente) Limpar() { c.NotasFiscais = nil }

// ValidarRegrasDeNegocio aplica as regras de negocio em todos os
// documentos carregados e devolve as rejeicoes por documento, na mesma
// ordem da lista. Porte de TNotasFiscais.ValidarRegrasdeNegocios.
func (c *Componente) ValidarRegrasDeNegocio() [][]*ErroRegraNegocio {
	r := make([][]*ErroRegraNegocio, 0, len(c.NotasFiscais))
	for _, nota := range c.NotasFiscais {
		r = append(r, ValidarRegrasNegocio(nota.NFAg, c.Configuracoes))
	}
	return r
}

// CStatConfirmada informa se o codigo de situacao indica documento
// confirmado. Porte de TACBrNFAg.CstatConfirmada.
func CStatConfirmada(cStat int) bool { return cStatConfirmada[cStat] }

// CStatProcessado informa se o codigo de situacao indica documento
// processado. Porte de TACBrNFAg.CstatProcessado.
func CStatProcessado(cStat int) bool { return cStatProcessado[cStat] }

// CStatCancelada informa se o codigo de situacao indica documento
// cancelado. Porte de TACBrNFAg.CstatCancelada.
func CStatCancelada(cStat int) bool { return cStatCancelada[cStat] }

// IdentificarSchema descobre o schema de um XML pelo nome do elemento
// raiz. Porte de TACBrNFAg.IdentificaSchema.
func IdentificarSchema(xml string) (SchemaNFAg, error) {
	doc, err := pcn.ParseString(xml)
	if err != nil {
		return SchErroNFAg, err
	}
	switch doc.Root.Nome {
	case TagNFAg, TagNFAgProc:
		return SchNFAg, nil
	case "consStatServNFAg", "retConsStatServNFAg":
		return SchconsStatServNFAg, nil
	case "retNFAg":
		return SchconsSitNFAg, nil // retNFAg e um retConsSitNFAg renomeado; a NFAg nao tem membro schret
	case "consSitNFAg", "retConsSitNFAg":
		return SchconsSitNFAg, nil
	case "eventoNFAg", "retEventoNFAg", "procEventoNFAg":
		return SchEventoNFAg, nil
	case "evCancNFAg":
		return SchCancNFAg, nil
	default:
		return SchErroNFAg, ErrXMLIncorreto
	}
}
