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
	"fmt"
	"strings"
)

// Tipos de controle do componente. Porte das declaracoes correspondentes
// de ACBrNFGas.Conversao.pas.

// ===========================================================================
// StatusNFGas (TStatusNFGas)
// ===========================================================================

// StatusNFGas e o estado interno do componente durante uma operacao.
// Nao tem representacao no leiaute.
type StatusNFGas int

const (
	StNFGasIdle StatusNFGas = iota
	StNFGasStatusServico
	StNFGasRecepcao
	StNFGasRetRecepcao
	StNFGasConsulta
	StNFGasEnvioWebService
	StNFGasEmail
	StNFGasEvento
)

// String retorna o nome do estado, para log e diagnostico.
func (v StatusNFGas) String() string {
	nomes := [...]string{
		"Idle", "StatusServico", "Recepcao", "RetRecepcao",
		"Consulta", "EnvioWebService", "Email", "Evento",
	}
	if int(v) < 0 || int(v) >= len(nomes) {
		return "Desconhecido"
	}
	return nomes[v]
}

// ===========================================================================
// VersaoNFGas (TVersaoNFGas)
// ===========================================================================

// VersaoNFGas e a versao do leiaute. O layout da NFGas tem uma unica versao
// publicada, e por isso NAO existe nenhum campo do modelo condicionado a
// versao -- diferente do que acontece na NFe e no SPED.
type VersaoNFGas int

const (
	Ve100 VersaoNFGas = iota // 1.00
)

// String retorna a versao no formato do leiaute ("1.00").
func (v VersaoNFGas) String() string {
	if v == Ve100 {
		return "1.00"
	}
	return ""
}

// Float retorna a versao como numero, para o atributo versao do XML.
func (v VersaoNFGas) Float() float64 {
	if v == Ve100 {
		return 1.00
	}
	return 0
}

// ParseVersaoNFGas converte "1.00" no enum.
func ParseVersaoNFGas(s string) (VersaoNFGas, error) {
	if s == "1.00" {
		return Ve100, nil
	}
	return Ve100, erroEnum("VersaoNFGas", s)
}

// ParseVersaoNFGasFloat converte 1.00 no enum.
func ParseVersaoNFGasFloat(d float64) (VersaoNFGas, error) {
	if d == 1.00 {
		return Ve100, nil
	}
	return Ve100, fmt.Errorf("%w: VersaoNFGas nao aceita %0.2f", ErrEnumInvalido, d)
}

// ===========================================================================
// SchemaNFGas (TSchemaNFGas)
// ===========================================================================

// SchemaNFGas identifica o schema XSD de um documento ou servico.
type SchemaNFGas int

const (
	SchErroNFGas SchemaNFGas = iota
	SchNFGas
	SchconsStatServNFGas
	SchretNFGas
	SchconsSitNFGas
	SchEventoNFGas
	SchevCancNFGas
)

// nomes dos membros sem o prefixo "sch". O Delphi obtem isso via
// GetEnumName + Copy(Result, 4, ...); aqui a tabela e explicita.
var schemaNFGasNomes = [...]string{
	"ErroNFGas", "NFGas", "consStatServNFGas", "retNFGas",
	"consSitNFGas", "EventoNFGas", "evCancNFGas",
}

// String retorna o nome do schema sem o prefixo "sch".
// Porte de SchemaNFGasToStr.
func (v SchemaNFGas) String() string {
	if int(v) < 0 || int(v) >= len(schemaNFGasNomes) {
		return ""
	}
	return schemaNFGasNomes[v]
}

// schemaEventoNomes reproduz TSchemaNFGasArrayStrings: so o membro de
// cancelamento tem string; os demais sao vazios. Usado por SchemaEvento.
var schemaEventoNomes = [...]string{"", "", "", "", "", "", "evCancNFGas"}

// SchemaEvento retorna o nome do schema de evento, ou vazio quando o membro
// nao corresponde a um evento. Porte de SchemaEventoToStr.
func (v SchemaNFGas) SchemaEvento() string {
	if int(v) < 0 || int(v) >= len(schemaEventoNomes) {
		return ""
	}
	return schemaEventoNomes[v]
}

// ParseSchemaNFGas converte o nome do schema no enum.
// Porte de StrToSchemaNFGas: corta em "_" se houver, aceita o nome com ou
// sem o prefixo "sch".
func ParseSchemaNFGas(s string) (SchemaNFGas, error) {
	nome := s
	if p := strings.Index(nome, "_"); p >= 0 {
		nome = nome[:p]
	}
	nome = strings.TrimPrefix(nome, "sch")

	for i, n := range schemaNFGasNomes {
		if n == nome {
			return SchemaNFGas(i), nil
		}
	}
	return SchErroNFGas, erroEnum("SchemaNFGas", s)
}

// ===========================================================================
// LayOutNFGas (TLayOutNFGas)
// ===========================================================================

// LayOutNFGas identifica o servico de web service.
type LayOutNFGas int

const (
	LayNFGasStatusServico LayOutNFGas = iota // NFGasStatusServico
	LayNFGasRecepcao                         // NFGasRecepcao
	LayNFGasConsulta                         // NFGasConsulta
	LayNFGasRetRecepcao                      // NFGasRetRecepcao
	LayNFGasEvento                           // NFGasRecepcaoEvento
	LayNFGasURLQRCode                        // URL-QRCode
	LayURLConsultaNFGas                      // URL-ConsultaNFGas
)

var layOutNFGasNomes = [...]string{
	"NFGasStatusServico", "NFGasRecepcao", "NFGasConsulta", "NFGasRetRecepcao",
	"NFGasRecepcaoEvento", "URL-QRCode", "URL-ConsultaNFGas",
}

// String retorna o nome do servico. Porte de LayOutNFGasToServico.
func (v LayOutNFGas) String() string {
	if int(v) < 0 || int(v) >= len(layOutNFGasNomes) {
		return ""
	}
	return layOutNFGasNomes[v]
}

// Schema devolve o schema correspondente ao servico.
// Porte de LayOutNFGasToSchema -- os servicos de URL caem em SchErroNFGas,
// que e o "else" do case original.
func (v LayOutNFGas) Schema() SchemaNFGas {
	switch v {
	case LayNFGasStatusServico:
		return SchconsStatServNFGas
	case LayNFGasRecepcao:
		return SchNFGas
	case LayNFGasConsulta:
		return SchconsSitNFGas
	case LayNFGasRetRecepcao:
		return SchretNFGas
	case LayNFGasEvento:
		return SchEventoNFGas
	default:
		return SchErroNFGas
	}
}

// ParseLayOutNFGas converte o nome do servico no enum.
// Porte de ServicoToLayOutNFGas.
func ParseLayOutNFGas(s string) (LayOutNFGas, error) {
	for i, n := range layOutNFGasNomes {
		if n == s {
			return LayOutNFGas(i), nil
		}
	}
	return LayNFGasStatusServico, erroEnum("LayOutNFGas", s)
}

// ===========================================================================
// VersaoQrCode (TVersaoQrCode)
// ===========================================================================

// VersaoQrCode e a versao do QR-Code impresso no DANFGas.
type VersaoQrCode int

const (
	Veqr000 VersaoQrCode = iota // 0
	Veqr100                     // 1
	Veqr200                     // 2
)

var versaoQrCodeCodigos = [...]string{"0", "1", "2"}

// String retorna o codigo da versao do QR-Code.
func (v VersaoQrCode) String() string {
	if int(v) < 0 || int(v) >= len(versaoQrCodeCodigos) {
		return ""
	}
	return versaoQrCodeCodigos[v]
}

// Float retorna a versao do QR-Code como numero.
func (v VersaoQrCode) Float() float64 { return float64(v) }

// ParseVersaoQrCode converte o codigo no enum.
func ParseVersaoQrCode(s string) (VersaoQrCode, error) {
	for i, c := range versaoQrCodeCodigos {
		if c == s {
			return VersaoQrCode(i), nil
		}
	}
	return Veqr000, erroEnum("VersaoQrCode", s)
}
