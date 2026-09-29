// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-28.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package sped

import (
	"fmt"
	"time"
)

// intToStrPadded formata um inteiro com zeros a esquerda ate a largura especificada.
func intToStrPadded(v int, width int) string {
	return fmt.Sprintf("%0*d", width, v)
}

// ===========================================================================
// Indicador de Movimento (TACBrIndMov)
// ===========================================================================

// IndMov indica se o bloco possui dados ou nao.
type IndMov int

const (
	IndMovComDados IndMov = iota // 0 - Bloco com dados informados
	IndMovSemDados               // 1 - Bloco sem dados informados
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndMov) String() string {
	switch v {
	case IndMovComDados:
		return "0"
	case IndMovSemDados:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Perfil (TACBrIndPerfil)
// ===========================================================================

// IndPerfil indica o perfil de apresentacao do arquivo fiscal.
type IndPerfil int

const (
	PerfilA      IndPerfil = iota // A
	PerfilB                       // B
	PerfilC                       // C
	PerfilD                       // D
	PerfilNenhum                  // Nenhum
)

// String retorna a letra do perfil ("A", "B", "C", "D" ou "").
func (v IndPerfil) String() string {
	switch v {
	case PerfilA:
		return "A"
	case PerfilB:
		return "B"
	case PerfilC:
		return "C"
	case PerfilD:
		return "D"
	case PerfilNenhum:
		return ""
	default:
		return ""
	}
}

// ===========================================================================
// Atividade (TACBrIndAtiv)
// ===========================================================================

// IndAtiv indica o tipo de atividade do contribuinte.
type IndAtiv int

const (
	AtivIndustrial IndAtiv = iota // 0 - Industrial ou equiparado a industrial
	AtivOutros                    // 1 - Outros
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndAtiv) String() string {
	switch v {
	case AtivIndustrial:
		return "0"
	case AtivOutros:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Versao Leiaute SPED Fiscal (TACBrVersaoLeiauteSPEDFiscal)
// ===========================================================================

// VersaoLeiauteFiscal indica a versao do leiaute do SPED Fiscal.
type VersaoLeiauteFiscal int

const (
	VlVersao100 VersaoLeiauteFiscal = iota // 001 - 2008
	VlVersao101                            // 002 - 2009
	VlVersao102                            // 003 - 2010
	VlVersao103                            // 004 - 2011
	VlVersao104                            // 005 - 2012
	VlVersao105                            // 006 - 2012
	VlVersao106                            // 007 - 2013
	VlVersao107                            // 008 - 2014
	VlVersao108                            // 009 - 2015
	VlVersao109                            // 010 - 2016
	VlVersao110                            // 011 - 2017
	VlVersao111                            // 012 - 2018
	VlVersao112                            // 013 - 2019
	VlVersao113                            // 014 - 2020
	VlVersao114                            // 015 - 2021
	VlVersao115                            // 016 - 2022
	VlVersao116                            // 017 - 2023
	VlVersao117                            // 018 - 2024
	VlVersao118                            // 019 - 2025
	VlVersao119                            // 020 - 2026
)

// String retorna o codigo de 3 digitos ("001", "002", ..., "020").
func (v VersaoLeiauteFiscal) String() string {
	return intToStrPadded(int(v)+1, 3)
}

// ===========================================================================
// Finalidade (TACBrCodFin)
// ===========================================================================

// CodFin indica a finalidade do arquivo (original ou substituto).
type CodFin int

const (
	CodFinOriginal   CodFin = iota // 0 - Original
	CodFinSubstituto               // 1 - Substituto
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v CodFin) String() string {
	switch v {
	case CodFinOriginal:
		return "0"
	case CodFinSubstituto:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Item (TACBrTipoItem)
// ===========================================================================

// TipoItem indica o tipo do item (mercadoria, servico, etc.).
type TipoItem int

const (
	TiMercadoriaRevenda    TipoItem = 0  // 00 - Mercadoria para Revenda
	TiMateriaPrima         TipoItem = 1  // 01 - Materia-Prima
	TiEmbalagem            TipoItem = 2  // 02 - Embalagem
	TiProdutoProcesso      TipoItem = 3  // 03 - Produto em Processo
	TiProdutoAcabado       TipoItem = 4  // 04 - Produto Acabado
	TiSubproduto           TipoItem = 5  // 05 - Subproduto
	TiProdutoIntermediario TipoItem = 6  // 06 - Produto Intermediario
	TiMaterialConsumo      TipoItem = 7  // 07 - Material de Uso e Consumo
	TiAtivoImobilizado     TipoItem = 8  // 08 - Ativo Imobilizado
	TiServicos             TipoItem = 9  // 09 - Servicos
	TiOutrosInsumos        TipoItem = 10 // 10 - Outros Insumos
	TiOutras               TipoItem = 99 // 99 - Outras
)

// String retorna o codigo de 2 digitos ("00", "01", ..., "10", "99").
func (v TipoItem) String() string {
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Tipo Operacao (TACBrIndOper)
// ===========================================================================

// IndOper indica o tipo de operacao (entrada ou saida).
type IndOper int

const (
	IndOperEntrada IndOper = iota // 0 - Entrada
	IndOperSaida                  // 1 - Saida
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndOper) String() string {
	switch v {
	case IndOperEntrada:
		return "0"
	case IndOperSaida:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Emitente (TACBrIndEmit)
// ===========================================================================

// IndEmit indica o emitente do documento fiscal.
type IndEmit int

const (
	IndEmitPropria   IndEmit = iota // 0 - Emissao propria
	IndEmitTerceiros                // 1 - Terceiros
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndEmit) String() string {
	switch v {
	case IndEmitPropria:
		return "0"
	case IndEmitTerceiros:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Pagamento (TACBrIndPgto)
// ===========================================================================

// IndPgto indica a forma de pagamento.
type IndPgto int

const (
	PgtoVista        IndPgto = 0  // 0 - A Vista
	PgtoPrazo        IndPgto = 1  // 1 - A Prazo
	PgtoOutros       IndPgto = 2  // 2 - Outros
	PgtoSemPagamento IndPgto = 9  // 9 - Sem pagamento
	PgtoNenhum       IndPgto = -1 // Nenhum (vazio)
)

// String retorna o valor para o arquivo SPED ("0", "1", "2", "9" ou "").
func (v IndPgto) String() string {
	switch v {
	case PgtoVista:
		return "0"
	case PgtoPrazo:
		return "1"
	case PgtoOutros:
		return "2"
	case PgtoSemPagamento:
		return "9"
	case PgtoNenhum:
		return ""
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// StringEm retorna o codigo de IND_PGTO vigente no periodo de dtIni.
//
// A partir de 2012-07-01 o valor "9" (sem pagamento) deixou de existir e "2"
// (outros) passou a valer; antes disso "2" nao existia. Um valor inaplicavel
// na vigencia sai vazio.
//
// Aqui ha divergencia deliberada em relacao ao ACBr: la os case nao cobrem
// todos os membros e a variavel strIND_PGTO conserva o valor da iteracao
// anterior do laco, emitindo o codigo de outro documento. Devolver vazio e o
// que o Guia Pratico prevê para campo inaplicavel.
//
// Ref.: ACBrEFDBloco_C_Class.pas, WriteRegistroC100.
func (v IndPgto) StringEm(dtIni time.Time) string {
	corte := time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC)

	if dtIni.Before(corte) {
		switch v {
		case PgtoVista:
			return "0"
		case PgtoPrazo:
			return "1"
		case PgtoSemPagamento:
			return "9"
		}
		return "" // PgtoOutros nao existia antes de 07/2012
	}

	switch v {
	case PgtoVista:
		return "0"
	case PgtoPrazo:
		return "1"
	case PgtoOutros:
		return "2"
	}
	return "" // PgtoSemPagamento deixou de existir em 07/2012
}

// ===========================================================================
// Frete (TACBrIndFrt)
// ===========================================================================

// IndFrt indica o tipo de frete.
type IndFrt int

const (
	FrtContaEmitente            IndFrt = 0  // 0 - Por conta do emitente (CIF)
	FrtContaDestinatario        IndFrt = 1  // 1 - Por conta do destinatario (FOB)
	FrtContaTerceiros           IndFrt = 2  // 2 - Por conta de terceiros
	FrtProprioPorContaRemetente IndFrt = 3  // 3 - Proprio por conta do remetente
	FrtProprioContaDestinatario IndFrt = 4  // 4 - Proprio por conta do destinatario
	FrtSemCobranca              IndFrt = 9  // 9 - Sem cobranca de frete
	FrtNenhum                   IndFrt = -1 // Nenhum (vazio)
)

// String retorna o valor para o arquivo SPED ("0"..."4", "9" ou "").
func (v IndFrt) String() string {
	switch v {
	case FrtContaEmitente:
		return "0"
	case FrtContaDestinatario:
		return "1"
	case FrtContaTerceiros:
		return "2"
	case FrtProprioPorContaRemetente:
		return "3"
	case FrtProprioContaDestinatario:
		return "4"
	case FrtSemCobranca:
		return "9"
	case FrtNenhum:
		return ""
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// StringEm retorna o codigo de IND_FRT vigente no periodo de dtIni.
//
// O leiaute mudou duas vezes e os codigos foram remanejados, nao apenas
// acrescentados: o mesmo valor "0" significa "por conta de terceiros" ate 2011
// e "por conta do emitente" de 2012 em diante. Usar String() para periodo
// antigo gera codigo errado sem nenhum erro visivel.
//
// Ref.: ACBrEFDBloco_C_Class.pas, WriteRegistroC100.
func (v IndFrt) StringEm(dtIni time.Time) string {
	if v == FrtNenhum {
		return ""
	}
	if v == FrtSemCobranca {
		return "9"
	}

	corte2012 := time.Date(2012, 1, 1, 0, 0, 0, 0, time.UTC)
	corte2018 := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)

	switch {
	case dtIni.Before(corte2012):
		switch v {
		case FrtContaTerceiros:
			return "0"
		case FrtContaEmitente, FrtProprioPorContaRemetente:
			return "1"
		case FrtContaDestinatario, FrtProprioContaDestinatario:
			return "2"
		}
	case dtIni.Before(corte2018):
		switch v {
		case FrtContaEmitente, FrtProprioPorContaRemetente:
			return "0"
		case FrtContaDestinatario, FrtProprioContaDestinatario:
			return "1"
		case FrtContaTerceiros:
			return "2"
		}
	default:
		return v.String()
	}
	return ""
}

// StringEmD100 retorna o codigo de IND_FRT vigente no periodo de dtIni segundo
// o writer do D100 -- que nao e o mesmo do C100.
//
// O ACBr usa criterios diferentes nos dois blocos e a diferenca nao parece
// intencional, mas replicamos como esta:
//
//   - o corte e 01/07/2012, nao 01/01/2012 como no C100. No primeiro semestre
//     de 2012 o mesmo IND_FRT sai com codigo diferente em C100 e D100;
//   - nao existe a faixa de 2018 em diante, entao o D100 nunca emite os
//     codigos "3" e "4" -- de 07/2012 em diante sempre 0/1/2/9.
//
// Ref.: ACBrEFDBloco_D_Class.pas, WriteRegistroD100.
func (v IndFrt) StringEmD100(dtIni time.Time) string {
	if v == FrtNenhum {
		return ""
	}
	if v == FrtSemCobranca {
		return "9"
	}

	if dtIni.Before(time.Date(2012, 7, 1, 0, 0, 0, 0, time.UTC)) {
		switch v {
		case FrtContaTerceiros:
			return "0"
		case FrtContaEmitente, FrtProprioPorContaRemetente:
			return "1"
		case FrtContaDestinatario, FrtProprioContaDestinatario:
			return "2"
		}
		return ""
	}

	switch v {
	case FrtContaEmitente, FrtProprioPorContaRemetente:
		return "0"
	case FrtContaDestinatario, FrtProprioContaDestinatario:
		return "1"
	case FrtContaTerceiros:
		return "2"
	}
	return ""
}

// ===========================================================================
// Frete Redespacho (TACBrTipoFreteRedespacho)
// ===========================================================================

// TipoFreteRedespacho indica o tipo de frete de redespacho.
type TipoFreteRedespacho int

const (
	FreteRedSemRedespacho     TipoFreteRedespacho = 0  // 0 - Sem redespacho
	FreteRedContaEmitente     TipoFreteRedespacho = 1  // 1 - Por conta do emitente
	FreteRedContaDestinatario TipoFreteRedespacho = 2  // 2 - Por conta do destinatario
	FreteRedOutros            TipoFreteRedespacho = 9  // 9 - Outros
	FreteRedNenhum            TipoFreteRedespacho = -1 // Nenhum (vazio)
)

// String retorna o valor para o arquivo SPED ("0", "1", "2", "9" ou "").
func (v TipoFreteRedespacho) String() string {
	switch v {
	case FreteRedSemRedespacho:
		return "0"
	case FreteRedContaEmitente:
		return "1"
	case FreteRedContaDestinatario:
		return "2"
	case FreteRedOutros:
		return "9"
	case FreteRedNenhum:
		return ""
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Origem Processo (TACBrOrigemProcesso)
// ===========================================================================

// OrigemProcesso indica a origem do processo.
type OrigemProcesso int

const (
	OrigProcSefaz           OrigemProcesso = 0  // 0 - SEFAZ
	OrigProcJusticaFederal  OrigemProcesso = 1  // 1 - Justica Federal
	OrigProcJusticaEstadual OrigemProcesso = 2  // 2 - Justica Estadual
	OrigProcSecexRFB        OrigemProcesso = 3  // 3 - SECEX/RFB
	OrigProcOutros          OrigemProcesso = 9  // 9 - Outros
	OrigProcNenhum          OrigemProcesso = -1 // Nenhum (vazio)
)

// String retorna o valor para o arquivo SPED ("0"..."3", "9" ou "").
func (v OrigemProcesso) String() string {
	switch v {
	case OrigProcSefaz:
		return "0"
	case OrigProcJusticaFederal:
		return "1"
	case OrigProcJusticaEstadual:
		return "2"
	case OrigProcSecexRFB:
		return "3"
	case OrigProcOutros:
		return "9"
	case OrigProcNenhum:
		return ""
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Operacao ST (TACBrIndTipoOperacaoST)
// ===========================================================================

// IndTipoOperacaoST indica o tipo de operacao de substituicao tributaria.
type IndTipoOperacaoST int

const (
	OpSTCombustiveis      IndTipoOperacaoST = iota // 0 - Combustiveis e lubrificantes
	OpSTLeasingVeiculos                            // 1 - Leasing de veiculos ou faturamento direto
	OpSTRecusaRecebimento                          // 2 - Recusa de recebimento
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v IndTipoOperacaoST) String() string {
	switch v {
	case OpSTCombustiveis:
		return "0"
	case OpSTLeasingVeiculos:
		return "1"
	case OpSTRecusaRecebimento:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Documento Arrecadacao (TACBrDoctoArrecada)
// ===========================================================================

// DoctoArrecada indica o tipo de documento de arrecadacao.
type DoctoArrecada int

const (
	DocArrecEstadual DoctoArrecada = iota // 0 - Documento Estadual de Arrecadacao
	DocArrecGNRE                          // 1 - GNRE
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v DoctoArrecada) String() string {
	switch v {
	case DocArrecEstadual:
		return "0"
	case DocArrecGNRE:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Transporte (TACBrTipoTransporte)
// ===========================================================================

// TipoTransporte indica o tipo de transporte.
type TipoTransporte int

const (
	TranspRodoviario      TipoTransporte = 0 // 0 - Rodoviario
	TranspFerroviario     TipoTransporte = 1 // 1 - Ferroviario
	TranspRodoFerroviario TipoTransporte = 2 // 2 - Rodo-Ferroviario
	TranspAquaviario      TipoTransporte = 3 // 3 - Aquaviario
	TranspDutoviario      TipoTransporte = 4 // 4 - Dutoviario
	TranspAereo           TipoTransporte = 5 // 5 - Aereo
	TranspOutros          TipoTransporte = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("0"..."5" ou "9").
func (v TipoTransporte) String() string {
	switch v {
	case TranspRodoviario:
		return "0"
	case TranspFerroviario:
		return "1"
	case TranspRodoFerroviario:
		return "2"
	case TranspAquaviario:
		return "3"
	case TranspDutoviario:
		return "4"
	case TranspAereo:
		return "5"
	case TranspOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Documento Importacao (TACBrDoctoImporta)
// ===========================================================================

// DoctoImporta indica o tipo de documento de importacao.
type DoctoImporta int

const (
	DocImpImportacao   DoctoImporta = iota // 0 - Declaracao de Importacao
	DocImpSimplificada                     // 1 - Declaracao Simplificada de Importacao
	DocImpUnica                            // 2 - DUIMP - Declaracao Unica de Importacao
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v DoctoImporta) String() string {
	switch v {
	case DocImpImportacao:
		return "0"
	case DocImpSimplificada:
		return "1"
	case DocImpUnica:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Titulo (TACBrTipoTitulo)
// ===========================================================================

// TipoTitulo indica o tipo de titulo de credito.
type TipoTitulo int

const (
	TitDuplicata   TipoTitulo = 0  // 00 - Duplicata
	TitCheque      TipoTitulo = 1  // 01 - Cheque
	TitPromissoria TipoTitulo = 2  // 02 - Promissoria
	TitRecibo      TipoTitulo = 3  // 03 - Recibo
	TitOutros      TipoTitulo = 99 // 99 - Outros
)

// String retorna o codigo de 2 digitos ("00", "01", "02", "03" ou "99").
func (v TipoTitulo) String() string {
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Movimento Fisico (TACBrIndMovFisica)
// ===========================================================================

// IndMovFisica indica se houve movimentacao fisica do item/produto.
type IndMovFisica int

const (
	MovFisicaSim IndMovFisica = iota // 0 - Sim
	MovFisicaNao                     // 1 - Nao
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndMovFisica) String() string {
	switch v {
	case MovFisicaSim:
		return "0"
	case MovFisicaNao:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Apuracao IPI (TACBrApuracaoIPI)
// ===========================================================================

// ConhecEmbarque indica o tipo de conhecimento de embarque no registro 1100.
// Os codigos nao sao sequenciais: o layout pula "05" e "15".
type ConhecEmbarque int

const (
	CeAWB ConhecEmbarque = iota
	CeMAWB
	CeHAWB
	CeCOMAT
	CeRExpressas
	CeEtiqRExpressas
	CeHrExpressas
	CeAV7
	CeBL
	CeMBL
	CeHBL
	CeCTR
	CeDSIC
	CeComatBL
	CeRWB
	CeHRWB
	CeTifDta
	CeCP2
	CeNaoIATA
	CeMNaoIATA
	CeHNaoIATA
	CeOutros
)

var conhecEmbarqueCodigos = map[ConhecEmbarque]string{
	CeAWB: "01", CeMAWB: "02", CeHAWB: "03", CeCOMAT: "04",
	CeRExpressas: "06", CeEtiqRExpressas: "07", CeHrExpressas: "08",
	CeAV7: "09", CeBL: "10", CeMBL: "11", CeHBL: "12", CeCTR: "13",
	CeDSIC: "14", CeComatBL: "16", CeRWB: "17", CeHRWB: "18",
	CeTifDta: "19", CeCP2: "20", CeNaoIATA: "91", CeMNaoIATA: "92",
	CeHNaoIATA: "93", CeOutros: "99",
}

// String retorna o codigo de dois digitos do tipo de conhecimento de embarque.
func (v ConhecEmbarque) String() string {
	if c, ok := conhecEmbarqueCodigos[v]; ok {
		return c
	}
	return ""
}

// MovimentoBens indica o tipo de movimentacao do bem ou componente do ativo
// imobilizado, no registro G125 (CIAP).
type MovimentoBens int

const (
	MovBensSaldoInicial      MovimentoBens = iota // SI - Saldo inicial de bens imobilizados
	MovBensImobilizacao                           // IM - Imobilizacao de bem individual
	MovBensImobilizacaoAndam                      // IA - Imobilizacao em andamento, componente
	MovBensConclusaoImob                          // CI - Conclusao de imobilizacao em andamento
	MovBensOriundaCirculante                      // MC - Imobilizacao oriunda do ativo circulante
	MovBensBaixaSaldo                             // BA - Baixa do saldo de ICMS, fim da apropriacao
	MovBensAlienacao                              // AT - Alienacao ou transferencia
	MovBensPerecimento                            // PE - Perecimento, extravio ou deterioracao
	MovBensOutrasSaidas                           // OT - Outras saidas do imobilizado
)

// String retorna a sigla do movimento conforme o layout ("SI", "IM", ...).
func (v MovimentoBens) String() string {
	switch v {
	case MovBensSaldoInicial:
		return "SI"
	case MovBensImobilizacao:
		return "IM"
	case MovBensImobilizacaoAndam:
		return "IA"
	case MovBensConclusaoImob:
		return "CI"
	case MovBensOriundaCirculante:
		return "MC"
	case MovBensBaixaSaldo:
		return "BA"
	case MovBensAlienacao:
		return "AT"
	case MovBensPerecimento:
		return "PE"
	case MovBensOutrasSaidas:
		return "OT"
	default:
		return ""
	}
}

// ApuracaoIPI indica o periodo de apuracao do IPI.
type ApuracaoIPI int

const (
	ApuracaoIPIMensal    ApuracaoIPI = 0  // 0 - Mensal
	ApuracaoIPIDecendial ApuracaoIPI = 1  // 1 - Decendial
	ApuracaoIPINenhum    ApuracaoIPI = -1 // Nenhum (vazio)
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "").
func (v ApuracaoIPI) String() string {
	switch v {
	case ApuracaoIPIMensal:
		return "0"
	case ApuracaoIPIDecendial:
		return "1"
	case ApuracaoIPINenhum:
		// Espaco, nao vazio: e o que o ACBr emite para iaNenhum
		// (ACBrEFDBloco_C_Class.pas, WriteRegistroC170).
		return " "
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Base Medicamento (TACBrTipoBaseMedicamento)
// ===========================================================================

// TipoBaseMedicamento indica o tipo de base de calculo do medicamento.
type TipoBaseMedicamento int

const (
	BaseMedTabeladoSugerido TipoBaseMedicamento = iota // 0 - Tabelado/Sugerido
	BaseMedMargemAgregado                              // 1 - Margem de valor agregado
	BaseMedListaNegativa                               // 2 - Lista Negativa
	BaseMedListaPositiva                               // 3 - Lista Positiva
	BaseMedListaNeutra                                 // 4 - Lista Neutra
)

// String retorna o valor para o arquivo SPED ("0"..."4").
func (v TipoBaseMedicamento) String() string {
	switch v {
	case BaseMedTabeladoSugerido:
		return "0"
	case BaseMedMargemAgregado:
		return "1"
	case BaseMedListaNegativa:
		return "2"
	case BaseMedListaPositiva:
		return "3"
	case BaseMedListaNeutra:
		return "4"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Produto (TACBrTipoProduto)
// ===========================================================================

// TipoProduto indica o tipo de produto (medicamentos).
type TipoProduto int

const (
	ProdSimilar  TipoProduto = iota // 0 - Similar
	ProdGenerico                    // 1 - Generico
	ProdMarca                       // 2 - De Marca
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v TipoProduto) String() string {
	switch v {
	case ProdSimilar:
		return "0"
	case ProdGenerico:
		return "1"
	case ProdMarca:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Arma Fogo (TACBrTipoArmaFogo)
// ===========================================================================

// TipoArmaFogo indica o tipo de arma de fogo.
type TipoArmaFogo int

const (
	ArmaPermitido TipoArmaFogo = iota // 0 - Permitido
	ArmaRestrito                      // 1 - Restrito
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v TipoArmaFogo) String() string {
	switch v {
	case ArmaPermitido:
		return "0"
	case ArmaRestrito:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Veiculo Operacao (TACBrIndVeicOper)
// ===========================================================================

// IndVeicOper indica o tipo de operacao com veiculo.
type IndVeicOper int

const (
	VeicVendaPConcess IndVeicOper = 0 // 0 - Venda para concessionaria
	VeicFaturaDireta  IndVeicOper = 1 // 1 - Faturamento direto
	VeicVendaDireta   IndVeicOper = 2 // 2 - Venda direta
	VeicVendaDConcess IndVeicOper = 3 // 3 - Venda da concessionaria
	VeicVendaOutros   IndVeicOper = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("0"..."3" ou "9").
func (v IndVeicOper) String() string {
	switch v {
	case VeicVendaPConcess:
		return "0"
	case VeicFaturaDireta:
		return "1"
	case VeicVendaDireta:
		return "2"
	case VeicVendaDConcess:
		return "3"
	case VeicVendaOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Receita (TACBrIndRec)
// ===========================================================================

// IndRec indica o tipo de receita.
type IndRec int

const (
	RecPropria  IndRec = iota // 0 - Receita Propria
	RecTerceiro               // 1 - Receita de Terceiro
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndRec) String() string {
	switch v {
	case RecPropria:
		return "0"
	case RecTerceiro:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Veiculo (TACBrTipoVeiculo)
// ===========================================================================

// TipoVeiculo indica o tipo de veiculo aquaviario.
type TipoVeiculo int

const (
	VeicEmbarcacao         TipoVeiculo = iota // 0 - Embarcacao
	VeicEmpuradorRebocador                    // 1 - Empurrador/Rebocador
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v TipoVeiculo) String() string {
	switch v {
	case VeicEmbarcacao:
		return "0"
	case VeicEmpuradorRebocador:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Navegacao (TACBrTipoNavegacao)
// ===========================================================================

// TipoNavegacao indica o tipo de navegacao.
type TipoNavegacao int

const (
	NavInterior  TipoNavegacao = iota // 0 - Interior
	NavCabotagem                      // 1 - Cabotagem
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v TipoNavegacao) String() string {
	switch v {
	case NavInterior:
		return "0"
	case NavCabotagem:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Situacao Documento (TACBrCodSit)
// ===========================================================================

// CodSit indica a situacao do documento fiscal.
type CodSit int

const (
	SitRegular         CodSit = 0 // 00 - Documento regular
	SitExtempRegular   CodSit = 1 // 01 - Escrituracao extemporanea de documento regular
	SitCancelado       CodSit = 2 // 02 - Documento cancelado
	SitCanceladoExtemp CodSit = 3 // 03 - Escrituracao extemporanea de documento cancelado
	SitDenegado        CodSit = 4 // 04 - NF-e ou CT-e denegado
	SitNumInutilizada  CodSit = 5 // 05 - NF-e ou CT-e inutilizada
	SitFiscalCompl     CodSit = 6 // 06 - Documento Fiscal Complementar
	SitExtempCompl     CodSit = 7 // 07 - Escrituracao extemporanea de documento complementar
	SitRegimeEspecNEsp CodSit = 8 // 08 - Documento Fiscal emitido com base em Regime Especial ou Norma Especifica
)

// String retorna o codigo de 2 digitos ("00", "01", ..., "08").
func (v CodSit) String() string {
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Tipo Tarifa (TACBrTipoTarifa)
// ===========================================================================

// TipoTarifa indica o tipo de tarifa aplicada.
type TipoTarifa int

const (
	TarifaExp   TipoTarifa = 0 // 0 - Exp
	TarifaEnc   TipoTarifa = 1 // 1 - Enc
	TarifaCI    TipoTarifa = 2 // 2 - CI
	TarifaOutra TipoTarifa = 9 // 9 - Outra
)

// String retorna o valor para o arquivo SPED ("0", "1", "2" ou "9").
func (v TipoTarifa) String() string {
	switch v {
	case TarifaExp:
		return "0"
	case TarifaEnc:
		return "1"
	case TarifaCI:
		return "2"
	case TarifaOutra:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Natureza Frete (TACBrNaturezaFrete)
// ===========================================================================

// NaturezaFrete indica a natureza do frete contratado.
type NaturezaFrete int

const (
	NatFreteNegociavel    NaturezaFrete = iota // 0 - Negociavel
	NatFreteNaoNegociavel                      // 1 - Nao Negociavel
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v NaturezaFrete) String() string {
	switch v {
	case NatFreteNegociavel:
		return "0"
	case NatFreteNaoNegociavel:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Receita (TACBrIndReceita)
// ===========================================================================

// IndReceita indica o tipo de receita (telecomunicacoes).
type IndReceita int

const (
	ReceitaServicoPrestado       IndReceita = 0 // 0 - Receita propria - servicos prestados
	ReceitaCobrancaDebitos       IndReceita = 1 // 1 - Receita propria - cobranca de debitos
	ReceitaVendaMerc             IndReceita = 2 // 2 - Receita propria - venda de mercadorias
	ReceitaServicoPrePago        IndReceita = 3 // 3 - Receita propria - servico pre-pago
	ReceitaOutrasProprias        IndReceita = 4 // 4 - Outras receitas proprias
	ReceitaTerceiroCoFaturamento IndReceita = 5 // 5 - Receita de terceiros (co-faturamento)
	ReceitaTerceiroOutras        IndReceita = 9 // 9 - Outras receitas de terceiros
)

// String retorna o valor para o arquivo SPED ("0"..."5" ou "9").
func (v IndReceita) String() string {
	switch v {
	case ReceitaServicoPrestado:
		return "0"
	case ReceitaCobrancaDebitos:
		return "1"
	case ReceitaVendaMerc:
		return "2"
	case ReceitaServicoPrePago:
		return "3"
	case ReceitaOutrasProprias:
		return "4"
	case ReceitaTerceiroCoFaturamento:
		return "5"
	case ReceitaTerceiroOutras:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Servico Prestado (TACBrServicoPrestado)
// ===========================================================================

// ServicoPrestado indica o tipo de servico de telecomunicacao prestado.
type ServicoPrestado int

const (
	ServTelefonia        ServicoPrestado = 0 // 0 - Telefonia
	ServComunicacaoDados ServicoPrestado = 1 // 1 - Comunicacao de Dados
	ServTVAssinatura     ServicoPrestado = 2 // 2 - TV por Assinatura
	ServAcessoInternet   ServicoPrestado = 3 // 3 - Acesso a Internet
	ServMultimidia       ServicoPrestado = 4 // 4 - Multimidia
	ServOutros           ServicoPrestado = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("0"..."4" ou "9").
func (v ServicoPrestado) String() string {
	switch v {
	case ServTelefonia:
		return "0"
	case ServComunicacaoDados:
		return "1"
	case ServTVAssinatura:
		return "2"
	case ServAcessoInternet:
		return "3"
	case ServMultimidia:
		return "4"
	case ServOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Movimento ST (TACBrMovimentoST)
// ===========================================================================

// MovimentoST indica se houve movimentacao de substituicao tributaria.
type MovimentoST int

const (
	MovSTSemOperacao MovimentoST = iota // 0 - Sem operacoes com ST
	MovSTComOperacao                    // 1 - Com operacoes de ST
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v MovimentoST) String() string {
	switch v {
	case MovSTSemOperacao:
		return "0"
	case MovSTComOperacao:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Ajuste (TACBrTipoAjuste)
// ===========================================================================

// TipoAjuste indica o tipo de ajuste da apuracao.
type TipoAjuste int

const (
	AjusteDebito  TipoAjuste = iota // 0 - Ajuste a Debito
	AjusteCredito                   // 1 - Ajuste a Credito
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v TipoAjuste) String() string {
	switch v {
	case AjusteDebito:
		return "0"
	case AjusteCredito:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Origem Documento (TACBrOrigemDocto)
// ===========================================================================

// OrigemDocto indica a origem do documento vinculado ao ajuste.
type OrigemDocto int

const (
	OrigDocProcessoJudicial OrigemDocto = 0 // 0 - Processo Judicial
	OrigDocProcessoAdminist OrigemDocto = 1 // 1 - Processo Administrativo
	OrigDocPerDcomp         OrigemDocto = 2 // 2 - PER/DCOMP
	OrigDocDocumentoFiscal  OrigemDocto = 3 // 3 - Documento Fiscal
	OrigDocOutros           OrigemDocto = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("0"..."3" ou "9").
func (v OrigemDocto) String() string {
	switch v {
	case OrigDocProcessoJudicial:
		return "0"
	case OrigDocProcessoAdminist:
		return "1"
	case OrigDocPerDcomp:
		return "2"
	case OrigDocDocumentoFiscal:
		return "3"
	case OrigDocOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Propriedade Item (TACBrIndProp)
// ===========================================================================

// IndProp indica a propriedade/posse do item.
type IndProp int

const (
	PropInformante           IndProp = iota // 0 - Item de propriedade do informante e em seu poder
	PropInformanteNoTerceiro                // 1 - Item de propriedade do informante em poder de terceiros
	PropTerceiroNoInformante                // 2 - Item de propriedade de terceiros em poder do informante
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v IndProp) String() string {
	switch v {
	case PropInformante:
		return "0"
	case PropInformanteNoTerceiro:
		return "1"
	case PropTerceiroNoInformante:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Documento Exportacao (TACBrTipoDocto)
// ===========================================================================

// TipoDoctoExport indica o tipo de documento de exportacao.
type TipoDoctoExport int

const (
	DocExportDeclaracao        TipoDoctoExport = iota // 0 - Declaracao de Exportacao
	DocExportDeclaracaoSimples                        // 1 - Declaracao Simplificada de Exportacao
	DocExportDeclaracaoUnica                          // 2 - Declaracao Unica de Exportacao
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v TipoDoctoExport) String() string {
	switch v {
	case DocExportDeclaracao:
		return "0"
	case DocExportDeclaracaoSimples:
		return "1"
	case DocExportDeclaracaoUnica:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Exportacao (TACBrExportacao)
// ===========================================================================

// Exportacao indica o tipo de exportacao.
type Exportacao int

const (
	ExportDireta   Exportacao = iota // 0 - Exportacao Direta
	ExportIndireta                   // 1 - Exportacao Indireta
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v Exportacao) String() string {
	switch v {
	case ExportDireta:
		return "0"
	case ExportIndireta:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Leiaute K010 (TACBrIndTipoLeiaute)
// ===========================================================================

// IndTipoLeiaute indica o tipo de leiaute do Bloco K.
type IndTipoLeiaute int

const (
	LeiauteSimplificado         IndTipoLeiaute = iota // 0 - Leiaute simplificado
	LeiauteCompleto                                   // 1 - Leiaute completo
	LeiauteRestritoSaldoEstoque                       // 2 - Leiaute restrito a saldo de estoque
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v IndTipoLeiaute) String() string {
	switch v {
	case LeiauteSimplificado:
		return "0"
	case LeiauteCompleto:
		return "1"
	case LeiauteRestritoSaldoEstoque:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Estoque K200 (TACBrIndEstoque)
// ===========================================================================

// IndEstoque indica o tipo de posse/propriedade do estoque.
type IndEstoque int

const (
	EstPropInformantePoder     IndEstoque = iota // 0 - Propriedade do informante e em seu poder
	EstPropInformanteTerceiros                   // 1 - Propriedade do informante e em poder de terceiros
	EstPropTerceirosInformante                   // 2 - Propriedade de terceiros e em poder do informante
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v IndEstoque) String() string {
	switch v {
	case EstPropInformantePoder:
		return "0"
	case EstPropInformanteTerceiros:
		return "1"
	case EstPropTerceirosInformante:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Motivo Inventario (TACBrMotInv)
// ===========================================================================

// MotInv indica o motivo do inventario.
type MotInv int

const (
	MotInvFinalPeriodo         MotInv = iota // 01 - No final no periodo
	MotInvMudancaTributacao                  // 02 - Na mudanca de forma de tributacao da mercadoria
	MotInvBaixaCadastral                     // 03 - Na solicitacao da baixa cadastral, paralisacao temporaria e outras situacoes
	MotInvRegimePagamento                    // 04 - Na alteracao de regime de pagamento condicao de contribuinte
	MotInvDeterminacaoFiscos                 // 05 - Por determinacao dos fiscos
	MotInvControleMercadoriaST               // 06 - Para controle das mercadorias sujeitas ao regime de substituicao tributaria - Loss/Devolucao
)

// String retorna o codigo de 2 digitos ("01", "02", ..., "06").
func (v MotInv) String() string {
	return intToStrPadded(int(v)+1, 2)
}

// ===========================================================================
// Grupo Tensao (TACBrGrupoTensao)
// ===========================================================================

// GrupoTensao indica o grupo de tensao (energia eletrica).
type GrupoTensao int

const (
	GrupoTensaoNenhum           GrupoTensao = iota // (vazio)
	GrupoTensaoA1                                  // 01 - A1 - Alta Tensao (230kV ou mais)
	GrupoTensaoA2                                  // 02 - A2 - Alta Tensao (88 a 138kV)
	GrupoTensaoA3                                  // 03 - A3 - Alta Tensao (69kV)
	GrupoTensaoA3a                                 // 04 - A3a - Alta Tensao (30 a 44kV)
	GrupoTensaoA4                                  // 05 - A4 - Alta Tensao (2,3 a 25kV)
	GrupoTensaoAS                                  // 06 - AS - Alta Tensao Subterraneo
	GrupoTensaoB1                                  // 07 - B1 - Residencial
	GrupoTensaoB1BaixaRenda                        // 08 - B1 - Residencial Baixa Renda
	GrupoTensaoB2Rural                             // 09 - B2 - Rural
	GrupoTensaoB2Cooperativa                       // 10 - B2 - Cooperativa de Eletrificacao Rural
	GrupoTensaoB2ServicoPublico                    // 11 - B2 - Servico Publico de Irrigacao
	GrupoTensaoB3                                  // 12 - B3 - Demais Classes
	GrupoTensaoB4a                                 // 13 - B4a - Iluminacao Publica rede de distribuicao
	GrupoTensaoB4b                                 // 14 - B4b - Iluminacao Publica bulbo da lampada
)

// String retorna o codigo de 2 digitos ("01"..."14") ou "" para nenhum.
func (v GrupoTensao) String() string {
	if v == GrupoTensaoNenhum {
		return ""
	}
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Classe Consumo (TACBrClasseConsumo)
// ===========================================================================

// ClasseConsumo indica a classe de consumo de energia eletrica.
type ClasseConsumo int

const (
	ClasseComercial         ClasseConsumo = 1 // 01 - Comercial
	ClasseConsumoProprio    ClasseConsumo = 2 // 02 - Consumo Proprio
	ClasseIluminacaoPublica ClasseConsumo = 3 // 03 - Iluminacao Publica
	ClasseIndustrial        ClasseConsumo = 4 // 04 - Industrial
	ClassePoderPublico      ClasseConsumo = 5 // 05 - Poder Publico
	ClasseResidencial       ClasseConsumo = 6 // 06 - Residencial
	ClasseRural             ClasseConsumo = 7 // 07 - Rural
	ClasseServicoPublico    ClasseConsumo = 8 // 08 - Servico Publico
)

// String retorna o codigo de 2 digitos ("01"..."08").
func (v ClasseConsumo) String() string {
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Tipo Ligacao (TACBrTpLigacao)
// ===========================================================================

// TpLigacao indica o tipo de ligacao (energia eletrica).
type TpLigacao int

const (
	LigacaoNenhum     TpLigacao = 0 // (vazio)
	LigacaoMonofasico TpLigacao = 1 // 1 - Monofasico
	LigacaoBifasico   TpLigacao = 2 // 2 - Bifasico
	LigacaoTrifasico  TpLigacao = 3 // 3 - Trifasico
)

// String retorna o valor para o arquivo SPED ("1", "2", "3" ou "").
func (v TpLigacao) String() string {
	switch v {
	case LigacaoNenhum:
		return ""
	case LigacaoMonofasico:
		return "1"
	case LigacaoBifasico:
		return "2"
	case LigacaoTrifasico:
		return "3"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Dispositivo Autorizado (TACBrDispositivo)
// ===========================================================================

// Dispositivo indica o tipo de dispositivo autorizado.
type Dispositivo int

const (
	DispFormSeguranca Dispositivo = 0 // 0 - Formulario de Seguranca - AIDF
	DispFSDA          Dispositivo = 1 // 1 - FS-DA - Formulario de Seguranca para Impressao de DANFE
	DispNFe           Dispositivo = 2 // 2 - Nota Fiscal Eletronica
	DispFormContinuo  Dispositivo = 3 // 3 - Formulario Continuo
	DispBlocos        Dispositivo = 4 // 4 - Blocos
	DispJogosSoltos   Dispositivo = 5 // 5 - Jogos Soltos
)

// String retorna o valor para o arquivo SPED ("0"..."5").
func (v Dispositivo) String() string {
	switch v {
	case DispFormSeguranca:
		return "0"
	case DispFSDA:
		return "1"
	case DispNFe:
		return "2"
	case DispFormContinuo:
		return "3"
	case DispBlocos:
		return "4"
	case DispJogosSoltos:
		return "5"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Assinante (TACBrTpAssinante)
// ===========================================================================

// TpAssinante indica o tipo de assinante (telecomunicacoes).
type TpAssinante int

const (
	AssNenhum              TpAssinante = 0 // (vazio)
	AssComercialIndustrial TpAssinante = 1 // 1 - Comercial/Industrial
	AssPoderPublico        TpAssinante = 2 // 2 - Poder Publico
	AssResidencial         TpAssinante = 3 // 3 - Residencial/Pessoa Fisica
	AssPublico             TpAssinante = 4 // 4 - Publico
	AssSemiPublico         TpAssinante = 5 // 5 - Semi-Publico
	AssOutros              TpAssinante = 6 // 6 - Outros
)

// String retorna o valor para o arquivo SPED ("1"..."6" ou "").
func (v TpAssinante) String() string {
	switch v {
	case AssNenhum:
		return ""
	case AssComercialIndustrial:
		return "1"
	case AssPoderPublico:
		return "2"
	case AssResidencial:
		return "3"
	case AssPublico:
		return "4"
	case AssSemiPublico:
		return "5"
	case AssOutros:
		return "6"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Movimento DIFAL (TACBrMovimentoDIFAL)
// ===========================================================================

// MovimentoDIFAL indica se houve movimentacao de DIFAL/FCP.
type MovimentoDIFAL int

const (
	DifalSemOperacao MovimentoDIFAL = iota // 0 - Sem operacoes com DIFAL
	DifalComOperacao                       // 1 - Com operacoes de DIFAL
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v MovimentoDIFAL) String() string {
	switch v {
	case DifalSemOperacao:
		return "0"
	case DifalComOperacao:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Natureza Conta (TACBrNaturezaConta)
// ===========================================================================

// NaturezaConta indica a natureza da conta contabil.
type NaturezaConta int

const (
	NatContaAtivo       NaturezaConta = 1 // 01 - Ativo
	NatContaPassivo     NaturezaConta = 2 // 02 - Passivo
	NatContaLiquido     NaturezaConta = 3 // 03 - Patrimonio Liquido
	NatContaResultado   NaturezaConta = 4 // 04 - Contas de Resultado
	NatContaCompensacao NaturezaConta = 5 // 05 - Contas de Compensacao
	NatContaOutras      NaturezaConta = 9 // 09 - Outras
)

// String retorna o codigo de 2 digitos ("01"..."05" ou "09").
func (v NaturezaConta) String() string {
	return intToStrPadded(int(v), 2)
}

// ===========================================================================
// Indicador CTA (TACBrIndCTA)
// ===========================================================================

// IndCTA indica se a conta contabil e sintetica ou analitica.
type IndCTA int

const (
	CTASintetica IndCTA = iota // S - Sintetica
	CTAAnalitica               // A - Analitica
)

// String retorna "S" (sintetica) ou "A" (analitica).
func (v IndCTA) String() string {
	switch v {
	case CTASintetica:
		return "S"
	case CTAAnalitica:
		return "A"
	default:
		return ""
	}
}

// ===========================================================================
// Medicao (TACBrMedicao)
// ===========================================================================

// Medicao indica o tipo de medicao (energia eletrica/gas).
type Medicao int

const (
	MedicaoAnalogico Medicao = iota // 0 - Analogico
	MedicaoDigital                  // 1 - Digital
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v Medicao) String() string {
	switch v {
	case MedicaoAnalogico:
		return "0"
	case MedicaoDigital:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Residuo (TACBrTipoResiduo)
// ===========================================================================

// TipoResiduo indica o tipo de residuo (usinas de etanol).
type TipoResiduo int

const (
	ResiduoBagacoCana TipoResiduo = 1 // 1 - Bagaco de cana
	ResiduoDDG        TipoResiduo = 2 // 2 - DDG
	ResiduoWDG        TipoResiduo = 3 // 3 - WDG
	ResiduoDDGWDG     TipoResiduo = 4 // 4 - DDG e WDG misturados
)

// String retorna o valor para o arquivo SPED ("1"..."4").
func (v TipoResiduo) String() string {
	switch v {
	case ResiduoBagacoCana:
		return "1"
	case ResiduoDDG:
		return "2"
	case ResiduoWDG:
		return "3"
	case ResiduoDDGWDG:
		return "4"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Finalidade Emissao Doc Eletronico C500 (TACBrFinalidadeEmissaoDocumentoEletronico)
// ===========================================================================

// FinalidadeEmissaoDocEletronico indica a finalidade de emissao do documento eletronico (C500).
type FinalidadeEmissaoDocEletronico int

const (
	FinEmissaoNaoDefinida  FinalidadeEmissaoDocEletronico = 0 // 0 - Nao definida
	FinEmissaoNormal       FinalidadeEmissaoDocEletronico = 1 // 1 - Normal
	FinEmissaoSubstituicao FinalidadeEmissaoDocEletronico = 2 // 2 - Substituicao
	FinEmissaoNormalAjuste FinalidadeEmissaoDocEletronico = 3 // 3 - Normal com ajuste
)

// String retorna o valor para o arquivo SPED ("0"..."3").
func (v FinalidadeEmissaoDocEletronico) String() string {
	switch v {
	case FinEmissaoNaoDefinida:
		return "0"
	case FinEmissaoNormal:
		return "1"
	case FinEmissaoSubstituicao:
		return "2"
	case FinEmissaoNormalAjuste:
		return "3"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Indicador Destinatario Acessante C500
// ===========================================================================

// IndDestinatarioAcessante indica o tipo de destinatario/acessante.
type IndDestinatarioAcessante int

const (
	DestAcessContribuinteICMS   IndDestinatarioAcessante = 1 // 1 - Contribuinte do ICMS
	DestAcessContribuinteIsento IndDestinatarioAcessante = 2 // 2 - Contribuinte Isento de inscricao no CCI
	DestAcessNaoContribuinte    IndDestinatarioAcessante = 9 // 9 - Nao contribuinte
)

// String retorna o valor para o arquivo SPED ("1", "2" ou "9").
func (v IndDestinatarioAcessante) String() string {
	switch v {
	case DestAcessContribuinteICMS:
		return "1"
	case DestAcessContribuinteIsento:
		return "2"
	case DestAcessNaoContribuinte:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Finalidade Fatura Eletronica D700
// ===========================================================================

// FinEmissaoFaturaEletronica indica a finalidade de emissao da fatura eletronica (D700).
type FinEmissaoFaturaEletronica int

const (
	FinFatNaoDefinida  FinEmissaoFaturaEletronica = 0 // 0 - Nao definida
	FinFatNormal       FinEmissaoFaturaEletronica = 1 // 1 - Normal
	FinFatSubstituicao FinEmissaoFaturaEletronica = 2 // 2 - Substituicao
	FinFatAjuste       FinEmissaoFaturaEletronica = 3 // 3 - Ajuste
)

// String retorna o valor para o arquivo SPED ("0"..."3").
func (v FinEmissaoFaturaEletronica) String() string {
	switch v {
	case FinFatNaoDefinida:
		return "0"
	case FinFatNormal:
		return "1"
	case FinFatSubstituicao:
		return "2"
	case FinFatAjuste:
		return "3"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Tipo Faturamento Doc Eletronico
// ===========================================================================

// TipoFaturamentoDocEletronico indica o tipo de faturamento do documento eletronico.
type TipoFaturamentoDocEletronico int

const (
	FatNormal        TipoFaturamentoDocEletronico = 0 // 0 - Faturamento Normal
	FatCentralizado  TipoFaturamentoDocEletronico = 1 // 1 - Faturamento Centralizado
	FatCofaturamento TipoFaturamentoDocEletronico = 2 // 2 - Cofaturamento
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v TipoFaturamentoDocEletronico) String() string {
	switch v {
	case FatNormal:
		return "0"
	case FatCentralizado:
		return "1"
	case FatCofaturamento:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Forma Pagamento (TACBrIndFormaPagto)
// ===========================================================================

// IndFormaPagto indica a forma de pagamento (telecomunicacoes).
type IndFormaPagto int

const (
	FormaPagtoPrePago IndFormaPagto = iota // 0 - Pre-pago
	FormaPagtoPosPago                      // 1 - Pos-pago
)

// String retorna o valor para o arquivo SPED ("0" ou "1").
func (v IndFormaPagto) String() string {
	switch v {
	case FormaPagtoPrePago:
		return "0"
	case FormaPagtoPosPago:
		return "1"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Motivo Ressarcimento
// ===========================================================================

// MotivoRessarcimento indica o motivo do ressarcimento.
type MotivoRessarcimento int

const (
	RessarcNenhum            MotivoRessarcimento = 0 // (vazio)
	RessarcVendaOutraUF      MotivoRessarcimento = 1 // 1 - Venda para outra UF
	RessarcSaidaIsenta       MotivoRessarcimento = 2 // 2 - Saida amparada por isencao ou nao incidencia
	RessarcPerdaDeterioracao MotivoRessarcimento = 3 // 3 - Perda ou deterioracao
	RessarcFurtoRoubo        MotivoRessarcimento = 4 // 4 - Furto ou roubo
	RessarcExportacao        MotivoRessarcimento = 5 // 5 - Exportacao
	RessarcVendaSimplesNac   MotivoRessarcimento = 6 // 6 - Venda interna para Simples Nacional
	RessarcOutros            MotivoRessarcimento = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("1"..."6", "9" ou "").
func (v MotivoRessarcimento) String() string {
	switch v {
	case RessarcNenhum:
		return ""
	case RessarcVendaOutraUF:
		return "1"
	case RessarcSaidaIsenta:
		return "2"
	case RessarcPerdaDeterioracao:
		return "3"
	case RessarcFurtoRoubo:
		return "4"
	case RessarcExportacao:
		return "5"
	case RessarcVendaSimplesNac:
		return "6"
	case RessarcOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Indicador Deducao
// ===========================================================================

// IndicadorDeducao indica o tipo de deducao do ISS.
type IndicadorDeducao int

const (
	DeducaoCompensacaoISS        IndicadorDeducao = 0 // 0 - Compensacao do ISS calculado a maior
	DeducaoBeneficioFiscal       IndicadorDeducao = 1 // 1 - Beneficio fiscal por incentivo a cultura
	DeducaoDecisaoAdministrativa IndicadorDeducao = 2 // 2 - Decisao administrativa ou judicial
	DeducaoOutros                IndicadorDeducao = 9 // 9 - Outros
)

// String retorna o valor para o arquivo SPED ("0", "1", "2" ou "9").
func (v IndicadorDeducao) String() string {
	switch v {
	case DeducaoCompensacaoISS:
		return "0"
	case DeducaoBeneficioFiscal:
		return "1"
	case DeducaoDecisaoAdministrativa:
		return "2"
	case DeducaoOutros:
		return "9"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}

// ===========================================================================
// Indicador Obrigacao
// ===========================================================================

// IndicadorObrigacao indica o tipo de obrigacao do ISS.
type IndicadorObrigacao int

const (
	ObrigISSProprio          IndicadorObrigacao = iota // 0 - ISS proprio
	ObrigISSSubstituto                                 // 1 - ISS substituto (devido pelo tomador)
	ObrigISSUniprofissionais                           // 2 - ISS Uniprofissionais
)

// String retorna o valor para o arquivo SPED ("0", "1" ou "2").
func (v IndicadorObrigacao) String() string {
	switch v {
	case ObrigISSProprio:
		return "0"
	case ObrigISSSubstituto:
		return "1"
	case ObrigISSUniprofissionais:
		return "2"
	default:
		return fmt.Sprintf("%d", int(v))
	}
}
