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

// Package dfe provides the transmission infrastructure shared by every
// electronic fiscal document: digital certificate loading (A1 PFX/PEM),
// XMLDSig enveloped signature with Canonical XML 1.0, and the SOAP 1.2
// client with mutual TLS used by the SEFAZ web services.
//
// It is a Go port of the corresponding roles of ACBrDFe.pas,
// ACBrDFeSSL.pas and ACBrDFeWebService.pas (Layer 2 da arquitetura).
//
// Este package E o unico em packages/ autorizado a importar net/http: ele e
// CLIENTE dos web services da SEFAZ -- transmissao e a funcao fiscal em si,
// nao exposicao de transporte. A regra de fronteira do CLAUDE.md (packages/
// nao servem HTTP nem falam JSON) continua valendo: aqui nao ha servidor,
// handler nem JSON.
package dfe
