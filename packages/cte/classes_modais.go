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

package cte

// Os SEIS MODAIS do CT-e, mais a variante RodoOS do CT-e OS.
//
// ESTE ARQUIVO AINDA NAO ESTA PORTADO. Os sete tipos sao declarados vazios
// aqui para que o classes.go e a raiz CTe fiquem COMPLETOS e o package
// compile; a etapa 2b os preenche a partir de ACBrCTe.Classes.pas:1548-2131.
//
// A escolha e deliberada. A alternativa seria omitir os campos `rodo`,
// `aereo` e companhia do InfCTeNorm nesta etapa e acrescenta-los na 2b --
// mas isso faria a raiz portada divergir do TCTe do ACBr por uma etapa
// inteira, e um struct raiz incompleto e exatamente o tipo de coisa que
// passa despercebida: nao quebra build nem teste, so produz documento sem
// modal nenhum.
//
// Enquanto estiverem vazios, preencher qualquer um deles nao tem efeito --
// nao ha campo para preencher. O gerador ainda nao existe, entao nao ha
// risco de emitir documento mutilado.
//
// Linhas de origem em ACBrCTe.Classes.pas:
//
//	TRodo        1651    TFerrov      2067
//	TRodoOS      1727    TDuto        2092
//	TAereo       1821    TMultimodal  2113
//	TAquav       1941
//
// O TPeriCollectionItem (:1788) cai no meio dessa faixa, e o tipo ja esta
// portado no classes.go porque tem DOIS donos -- veja o doc-comment do Peri
// la. O que falta e o CAMPO dentro do Aereo, anotado abaixo.
//
// A 2b tambem precisa do alias TLacRodoCollection = TLacreCollection
// (:1630): o `lacRodo` do modal rodoviario e `[]Lacre`, como os outros tres
// aliases de lacre ja documentados no README.

// Rodo e o modal rodoviario. Porte de TRodo -- etapa 2b.
type Rodo struct{}

// RodoOS e o modal rodoviario do CT-e OS. Porte de TRodoOS -- etapa 2b.
//
// E tipo PROPRIO, nao variante do Rodo: o CT-e OS troca o grupo de veiculos
// e o de vale-pedagio.
type RodoOS struct{}

// Aereo e o modal aereo. Porte de TAereo -- etapa 2b.
//
// ATENCAO ao portar: alem das sete properties proprias, o TAereo tem
// `peri: TPeriCollection` (Classes.pas:1829), que da ve300 em diante e o
// UNICO lugar de onde o produto perigoso sai -- o infCTeNorm.peri so vale
// na ve200. Esquecer o campo produz CT-e aereo 4.00 sem produto perigoso
// nenhum, sem quebrar build, vet ou teste. O tipo Peri ja existe, no
// classes.go.
type Aereo struct{}

// Aquav e o modal aquaviario. Porte de TAquav -- etapa 2b.
type Aquav struct{}

// Ferrov e o modal ferroviario. Porte de TFerrov -- etapa 2b.
type Ferrov struct{}

// Duto e o modal dutoviario. Porte de TDuto -- etapa 2b.
type Duto struct{}

// Multimodal e o modal multimodal. Porte de TMultimodal -- etapa 2b.
type Multimodal struct{}
