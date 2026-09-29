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

package sped

import "time"

// ---------------------------------------------------------------------------
// Bloco K - Controle da Producao e do Estoque
// Portado de: ACBrEFDBlocos.pas / ACBrSpedFiscal
// ---------------------------------------------------------------------------

// RegistroK001 - Abertura do Bloco K
type RegistroK001 struct {
	OpenBlocos
	RegistroK010 []*RegistroK010
	RegistroK100 []*RegistroK100
}

// NewRegistroK001 cria um novo RegistroK001 com IndDad=1 (sem dados).
func NewRegistroK001() *RegistroK001 {
	return &RegistroK001{OpenBlocos: OpenBlocos{IndDad: 1}}
}

// RegistroK010 - Informacao sobre o Tipo de Leiaute (K010)
type RegistroK010 struct {
	IndTpLeiaute IndTipoLeiaute
}

// RegistroK100 - Periodo de Apuracao do ICMS/IPI
type RegistroK100 struct {
	DtIni        time.Time
	DtFin        time.Time
	RegistroK200 []*RegistroK200
	RegistroK210 []*RegistroK210
	RegistroK220 []*RegistroK220
	RegistroK230 []*RegistroK230
	RegistroK250 []*RegistroK250
	RegistroK260 []*RegistroK260
	RegistroK270 []*RegistroK270
	RegistroK280 []*RegistroK280
	RegistroK290 []*RegistroK290
	RegistroK300 []*RegistroK300
}

// RegistroK200 - Estoque Escriturado
type RegistroK200 struct {
	DtEst   time.Time
	CodItem string
	Qtd     float64
	IndEst  IndEstoque
	CodPart string
}

// RegistroK210 - Desmontagem de Mercadorias - Item de Origem
type RegistroK210 struct {
	DtIniOS      time.Time
	DtFinOS      time.Time
	CodDocOS     string
	CodItemOri   string
	Qtd          float64
	RegistroK215 []*RegistroK215
}

// RegistroK215 - Desmontagem de Mercadorias - Item de Destino
type RegistroK215 struct {
	CodItemDest string
	Qtd         float64
}

// RegistroK220 - Outras Movimentacoes Internas entre Mercadorias
type RegistroK220 struct {
	DtMov       time.Time
	CodItemOri  string
	CodItemDest string
	Qtd         float64
}

// RegistroK230 - Itens Produzidos
type RegistroK230 struct {
	DtIniOP      time.Time
	DtFinOP      time.Time
	CodDocOP     string
	CodItem      string
	Qtd          float64
	RegistroK235 []*RegistroK235
}

// RegistroK235 - Insumos Consumidos
type RegistroK235 struct {
	DtSaida time.Time
	CodItem string
	Qtd     float64
	CodInsProd string
}

// RegistroK250 - Industrializacao Efetuada por Terceiros - Itens Produzidos
type RegistroK250 struct {
	DtProd       time.Time
	CodItem      string
	Qtd          float64
	RegistroK255 []*RegistroK255
}

// RegistroK255 - Industrializacao em Terceiros - Insumos Consumidos
type RegistroK255 struct {
	DtCons time.Time
	CodItem string
	Qtd     float64
	CodInsProd string
}

// RegistroK260 - Reprocessamento/Reparo de Produto/Insumo
type RegistroK260 struct {
	CodOP        string
	CodItem      string
	DtSaida      time.Time
	Qtd          float64
	DtRet        time.Time
	QtdRet       float64
	RegistroK265 []*RegistroK265
}

// RegistroK265 - Reprocessamento/Reparo - Mercadorias Consumidas e/ou Retornadas
type RegistroK265 struct {
	CodItem string
	Qtd     float64
	QtdRet  float64
}

// RegistroK270 - Correcao de Apontamento dos Registros K210, K220, K230,
// K250, K260
type RegistroK270 struct {
	DtIniAP      time.Time
	DtFinAP      time.Time
	CodOP        string
	CodItem      string
	Qtd          float64
	QtdCorr      float64
	RegistroK275 []*RegistroK275
}

// RegistroK275 - Correcao de Apontamento e Retorno de Insumos dos Registros
// K215, K220, K235, K255, K265
type RegistroK275 struct {
	CodItem string
	Qtd     float64
	QtdCorr float64
}

// RegistroK280 - Correcao de Apontamento - Estoque Escriturado
type RegistroK280 struct {
	DtEst   time.Time
	CodItem string
	Qtd     float64
	QtdCorr float64
	IndEst  IndEstoque
	CodPart string
}

// RegistroK290 - Producao Conjunta - Ordem de Producao
type RegistroK290 struct {
	DtIniOP      time.Time
	DtFinOP      time.Time
	CodDocOP     string
	RegistroK291 []*RegistroK291
	RegistroK292 []*RegistroK292
}

// RegistroK291 - Producao Conjunta - Itens Produzidos
type RegistroK291 struct {
	CodItem string
	Qtd     float64
}

// RegistroK292 - Producao Conjunta - Insumos Consumidos
type RegistroK292 struct {
	CodItem    string
	Qtd        float64
	CodInsProd string
}

// RegistroK300 - Producao Conjunta - Industrializacao Efetuada por Terceiros
type RegistroK300 struct {
	DtProd       time.Time
	RegistroK301 []*RegistroK301
	RegistroK302 []*RegistroK302
}

// RegistroK301 - Producao Conjunta - Industrializacao Efetuada por Terceiros -
// Itens Produzidos
type RegistroK301 struct {
	CodItem string
	Qtd     float64
}

// RegistroK302 - Producao Conjunta - Industrializacao Efetuada por Terceiros -
// Insumos Consumidos
type RegistroK302 struct {
	CodItem    string
	Qtd        float64
	CodInsProd string
}

// RegistroK990 - Encerramento do Bloco K
type RegistroK990 struct {
	QtdLinK int
}
