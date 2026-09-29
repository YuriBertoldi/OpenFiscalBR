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

// Geracao do XML da NFGas -- NAO IMPLEMENTADA nesta fase.
//
// As assinaturas abaixo correspondem a TNFGasXmlWriter
// (ACBrNFGas.XmlWriter.pas) e existem para que o consumidor ja escreva
// contra a API definitiva. Todas devolvem ErrNaoImplementado.
//
// O que falta portar, em ordem de dependencia:
//
//  1. montagem dos elementos de infNFGas, na ordem exata do XSD;
//  2. validacao de campo com as descricoes DSC_* de ACBrNFGas.Consts.pas;
//  3. calculo do digito verificador e montagem do atributo Id;
//  4. assinatura digital XMLDSig (certificado A1/A3);
//  5. montagem do QR-Code (GetURLQRCode / AjustarVersaoQRCode).
//
// Os itens 1 a 3 ja tem toda a informacao necessaria neste package: o
// modelo de dados esta completo, MontarChaveAcesso monta a chave e os
// enums devolvem o codigo exato do leiaute.

// GerarXML gera o XML da NFGas a partir dos dados do documento.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func GerarXML(n *NFGas) (string, error) {
	return "", ErrNaoImplementado
}

// GerarXMLProc gera o XML de nfgasProc, juntando o documento ao protocolo.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func GerarXMLProc(n *NFGas) (string, error) {
	return "", ErrNaoImplementado
}

// GerarXMLEvento gera o XML de eventoNFGas.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func GerarXMLEvento(e *EventoNFGas) (string, error) {
	return "", ErrNaoImplementado
}

// Assinar assina digitalmente o XML do documento.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado. Depende de certificado
// A1/A3 e de canonicalizacao XML, que ainda nao foram portados.
func Assinar(n *NFGas, xml string) (string, error) {
	return "", ErrNaoImplementado
}

// GerarQRCode monta a URL do QR-Code do DANFGas.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado. Depende da tabela de URLs
// por UF, que vive em ACBrNFGasServicos.ini.
func GerarQRCode(n *NFGas, cfg Configuracoes) (string, error) {
	return "", ErrNaoImplementado
}
