// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-22.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

// Package cte porta o componente ACBrCTe -- Conhecimento de Transporte
// Eletronico.
//
// # Quatro documentos, um componente
//
// A diferenca estrutural em relacao ao packages/nfe e que o CT-e nao e UM
// documento com dois modelos: sao QUATRO leiautes distintos atendidos pela
// mesma arvore de classes.
//
//	modelo  documento                      enum
//	57      CT-e                           MoCTe
//	64      GTV-e (Guia de Transporte)     MoGTVe
//	67      CT-e OS (Outros Servicos)      MoCTeOS
//	57      CT-e Simplificado              MoCTeSimp
//
// O Simplificado COMPARTILHA o codigo 57 com o CT-e normal. O comentario do
// proprio ACBr em ACBrCTe.Conversao.pas:106 explica: "O codigo correto do
// modelo CTe Simplificado e 57, mas foi alterado para 58" -- internamente ha
// um 58 para distingui-los, e o que sai no documento e 57.
//
// A consequencia pratica aparece na conversao: ModeloCTeToStr tem quatro
// entradas e StrToModeloCTe tem TRES. Ler "57" devolve sempre MoCTe, nunca
// MoCTeSimp -- a ida e volta pelo codigo do modelo nao fecha, e nao ha como
// fechar, porque a informacao nao esta no codigo.
//
// # Seis modais
//
// Rodoviario, aereo, aquaviario, ferroviario, dutoviario e multimodal, cada
// um com arvore de classes propria e mutuamente exclusivos. E o analogo do
// grupo `det` da NF-e em peso, mas com seis variantes em vez de uma.
//
// # Origem dos enums: PCNCTe/pcteConversaoCTe.pas
//
// O ACBrCTe declara os MESMOS 37 enums em duas units, e elas divergem:
//
//   - Base/ACBrCTe.Conversao.pas
//   - PCNCTe/pcteConversaoCTe.pas
//
// Vale a segunda, por duas evidencias independentes:
//
//  1. Resolucao do `uses`. As seis units vivas -- XmlWriter, XmlReader,
//     IniWriter, IniReader, Classes e XmlHandler -- listam `pcteConversaoCTe`.
//     A unica mencao a `ACBrCTe.Conversao` esta em XmlWriter.pas:46 e esta
//     COMENTADA.
//  2. A unit Base/ perde membros. O enum TSchemaCTe la nao tem
//     `schevVincPgto` nem `schevCancVincPgto` -- o evento de Vinculo de
//     Pagamento e o cancelamento dele, que sao os dois mais recentes.
//
// Travado por TestSchemaCTe_MembrosExclusivosDaUnitPCN.
package cte
