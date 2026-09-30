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

import "fmt"

// Tipos estruturais do componente (status, versao, schema, layout). Porte
// de ACBrNFAg.Conversao.pas.

// ===========================================================================
// StatusNFAg (TStatusNFAg)
// ===========================================================================

// StatusNFAg e o estado interno do componente durante uma operacao.
// Nao tem representacao no leiaute.
type StatusNFAg int

const (
	StNFAgIdle StatusNFAg = iota
	StNFAgStatusServico
	StNFAgRecepcao
	StNFAgRetRecepcao
	StNFAgConsulta
	StNFAgEvento
	StNFAgEnvioWebService
	StNFAgEmail
)

// String retorna o nome do estado, para log e diagnostico.
func (v StatusNFAg) String() string {
	nomes := [...]string{
		"Idle", "StatusServico", "Recepcao", "RetRecepcao",
		"Consulta", "Evento", "EnvioWebService", "Email",
	}
	if int(v) < 0 || int(v) >= len(nomes) {
		return "Desconhecido"
	}
	return nomes[v]
}

// ===========================================================================
// VersaoNFAg (TVersaoNFAg)
// ===========================================================================

// VersaoNFAg e a versao do leiaute. O layout da NFAg tem uma unica versao
// publicada, e por isso NAO existe nenhum campo do modelo condicionado a
// versao.
type VersaoNFAg int

const (
	Ve100 VersaoNFAg = iota // 1.00
)

// String retorna a versao no formato do leiaute ("1.00").
func (v VersaoNFAg) String() string {
	if v == Ve100 {
		return "1.00"
	}
	return ""
}

// Float retorna a versao como numero, para o atributo versao do XML.
func (v VersaoNFAg) Float() float64 {
	if v == Ve100 {
		return 1.00
	}
	return 0
}

// ParseVersaoNFAg converte "1.00" no enum.
func ParseVersaoNFAg(s string) (VersaoNFAg, error) {
	if s == "1.00" {
		return Ve100, nil
	}
	return Ve100, erroEnum("VersaoNFAg", s)
}

// ParseVersaoNFAgFloat converte 1.00 no enum.
func ParseVersaoNFAgFloat(d float64) (VersaoNFAg, error) {
	if d == 1.00 {
		return Ve100, nil
	}
	return Ve100, fmt.Errorf("%w: VersaoNFAg nao aceita %0.2f", ErrEnumInvalido, d)
}

// ===========================================================================
// SchemaNFAg (TSchemaNFAg)
// ===========================================================================

// SchemaNFAg identifica o schema XSD de um documento ou servico.
// A ordem dos membros e a de TSchemaNFAg, que difere da do NFGas
// (schNFAg vem DEPOIS de schconsStatServNFAg, e nao ha membro retNFAg).
type SchemaNFAg int

const (
	SchErroNFAg SchemaNFAg = iota
	SchconsStatServNFAg
	SchNFAg
	SchconsSitNFAg
	SchEventoNFAg
	SchCancNFAg
)

// nomes dos membros sem o prefixo "sch" (GetEnumName + Copy do Delphi).
var schemaNFAgNomes = [...]string{
	"Erro", "consStatServNFAg", "NFAg", "consSitNFAg", "EventoNFAg", "CancNFAg",
}

// String retorna o nome do schema sem o prefixo "sch".
// Porte de SchemaNFAgToStr.
func (v SchemaNFAg) String() string {
	if int(v) < 0 || int(v) >= len(schemaNFAgNomes) {
		return ""
	}
	return schemaNFAgNomes[v]
}

// schemaEventoNomes reproduz TSchemaArrayStrings: so o membro de
// cancelamento tem string. Porte de SchemaEventoToStr.
var schemaEventoNomes = [...]string{"", "", "", "", "", "evCancNFAg"}

// SchemaEvento retorna o nome do schema de evento, ou vazio quando o membro
// nao corresponde a um evento.
func (v SchemaNFAg) SchemaEvento() string {
	if int(v) < 0 || int(v) >= len(schemaEventoNomes) {
		return ""
	}
	return schemaEventoNomes[v]
}

// ===========================================================================
// LayOutNFAg (TLayOut)
// ===========================================================================

// LayOutNFAg identifica o servico do web service.
type LayOutNFAg int

const (
	LayNFAgStatusServico LayOutNFAg = iota
	LayNFAgRecepcao
	LayNFAgRetRecepcao
	LayNFAgConsulta
	LayNFAgEvento
	LayNFAgQRCode
	LayNFAgURLConsulta
)

var layOutNFAgNomes = [...]string{
	"NFAgStatusServico", "NFAgRecepcao", "NFAgRetRecepcao",
	"NFAgConsulta", "NFAgRecepcaoEvento", "URL-QRCode", "URL-ConsultaNFAg",
}

// String retorna o nome do servico como usado no ACBrNFAgServicos.ini.
// Porte de LayOutToServico.
func (v LayOutNFAg) String() string {
	if int(v) < 0 || int(v) >= len(layOutNFAgNomes) {
		return ""
	}
	return layOutNFAgNomes[v]
}

// ParseLayOutNFAg converte o nome do servico no enum.
// Porte de ServicoToLayOut.
func ParseLayOutNFAg(s string) (LayOutNFAg, error) {
	for i, nome := range layOutNFAgNomes {
		if nome == s {
			return LayOutNFAg(i), nil
		}
	}
	return LayNFAgStatusServico, erroEnum("LayOutNFAg", s)
}

// Schema retorna o schema associado ao servico. Porte de LayOutToSchema.
func (v LayOutNFAg) Schema() SchemaNFAg {
	switch v {
	case LayNFAgStatusServico:
		return SchconsStatServNFAg
	case LayNFAgRecepcao:
		return SchNFAg
	case LayNFAgConsulta:
		return SchconsSitNFAg
	case LayNFAgEvento:
		return SchEventoNFAg
	}
	return SchErroNFAg
}

// ===========================================================================
// VersaoQrCode (TVersaoQrCode)
// ===========================================================================

// VersaoQrCode e a versao do QR-Code do DANFAg.
type VersaoQrCode int

const (
	Veqr000 VersaoQrCode = iota // 0
	Veqr100                     // 1
	Veqr200                     // 2
)

// String retorna o codigo da versao ("0", "1", "2").
func (v VersaoQrCode) String() string {
	nomes := [...]string{"0", "1", "2"}
	if int(v) < 0 || int(v) >= len(nomes) {
		return ""
	}
	return nomes[v]
}

// Float retorna a versao como numero.
func (v VersaoQrCode) Float() float64 { return float64(v) }

// ParseVersaoQrCode converte o codigo no enum.
func ParseVersaoQrCode(s string) (VersaoQrCode, error) {
	switch s {
	case "0":
		return Veqr000, nil
	case "1":
		return Veqr100, nil
	case "2":
		return Veqr200, nil
	}
	return Veqr000, erroEnum("VersaoQrCode", s)
}
