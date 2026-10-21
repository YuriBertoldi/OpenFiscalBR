// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-20.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

// Package nfe implementa a Nota Fiscal Eletronica (modelo 55) e a Nota
// Fiscal de Consumidor Eletronica (modelo 65), porte do componente ACBrNFe
// do Projeto ACBr.
//
// # Origem dos enums
//
// O ACBrNFe declara os mesmos 51 enums em DUAS units: Base/ACBrNFe.Conversao.pas
// e PCNNFe/pcnConversaoNFe.pas. As duas divergem, e a que vale e a segunda, por
// tres evidencias independentes:
//
//  1. Todas as units vivas (Base/ACBrNFe.XmlWriter.pas, .XmlReader.pas,
//     .IniWriter.pas, .IniReader.pas e ACBrNFe.Classes.pas) listam
//     pcnConversaoNFe, e o Delphi resolve identificador duplicado pela ULTIMA
//     unit listada. Em ACBrNFe.Classes.pas -- a unica que ainda lista as duas
//     -- pcnConversaoNFe vem depois.
//  2. Os proprios autores do ACBr COMENTARAM ACBrNFe.Conversao em
//     ACBrNFe.ValidarRegrasdeNegocio.pas:45 e em
//     Base/Servicos/ACBrNFe.EventoClass.pas:48, logo abaixo do
//     pcnConversaoNFe ativo.
//  3. Base/ACBrNFe.XmlWriter.pas:546 usa fnDebito, membro que so existe em
//     PCNNFe/pcnConversaoNFe.pas:303.
//
// Base/ACBrNFe.Conversao.pas nao tem fnCredito/fnDebito, nem
// fpPagInstantaneoPIXAutomatico/fpTEFBookTransfer, nem os codigos "23" e "24"
// de FormaPagamento.
//
// Por isso este package porta PCNNFe/pcnConversaoNFe.pas e trata
// Base/ACBrNFe.Conversao.pas como duplicata obsoleta. Registrado em
// .claude/skills/convert/referencias/divergencias-acbr.md.
//
// # Semantica das tabelas de codigo
//
// O ACBr converte enum por duas funcoes de ACBrDFe.Conversao.pas, e as duas
// tem comportamento que nao se adivinha pela assinatura:
//
//   - EnumeradoToStr NAO interrompe o laco: a ULTIMA ocorrencia do membro na
//     lista e a que vale. E o que faz crtRegimeNormal gravar "3" e nao "",
//     apesar de aparecer na posicao 0 da tabela do CRT.
//   - StrToEnumerado interrompe na PRIMEIRA ocorrencia e, quando nada casa,
//     devolve AEnumerados[0] com ok=False -- nao devolve -1.
//
// Cada ParseXxx deste package devolve o mesmo valor de fallback do ACBr junto
// do error, para que o chamador que ignora o erro (padrao do leitor) obtenha o
// mesmo resultado do Delphi.
//
// # Fora de escopo
//
// Helpers de apresentacao do DANFE nao foram portados, por dependerem de
// engine grafica sem equivalente em Go (regra da /convert): os *ToStrTagPosText,
// *ToDescricao, *ToDescStr, indProcToDescrStr, modFreteToDesStr, ArmaTipoStr e
// a familia Veiculos*Str. Pelo mesmo motivo ficaram de fora os enums usados
// unicamente pelo DANFE (TpcnPercentualTributos, TpcnImprimeDescAcrescItem,
// TImprimirUnidQtdeValor e TpcnInformacoesDePagamento).
package nfe
