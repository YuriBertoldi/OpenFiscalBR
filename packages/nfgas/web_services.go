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

import "context"

// Comunicacao com os web services da SEFAZ -- NAO IMPLEMENTADA nesta fase.
//
// As assinaturas correspondem aos metodos de TACBrNFGas
// (ACBrNFGas.pas / ACBrNFGasWebServices.pas) e existem para que o
// consumidor ja escreva contra a API definitiva. Todas devolvem
// ErrNaoImplementado.
//
// O que falta portar:
//
//   - certificado digital A1/A3 e TLS mutuo;
//   - envelope SOAP por servico;
//   - tabela de URLs por UF e ambiente (ACBrNFGasServicos.ini);
//   - leitura dos retornos de recepcao e de status de servico.
//
// A leitura dos retornos de CONSULTA e de EVENTO ja esta pronta:
// LerRetConsSit e LerEvento funcionam sobre um XML de retorno guardado em
// disco, sem precisar de rede.
//
// Todos os metodos recebem context.Context porque a implementacao futura
// fara chamada de rede -- o contrato ja nasce cancelavel.

// Enviar transmite um lote de documentos a SEFAZ.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func (c *Componente) Enviar(ctx context.Context, lote string, notas []*NotaFiscal) error {
	return ErrNaoImplementado
}

// Consultar consulta a situacao de um documento pela chave de acesso.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado. Para ler um retorno de
// consulta ja obtido, use LerRetConsSit.
func (c *Componente) Consultar(ctx context.Context, chave string) (*RetConsSitNFGas, error) {
	return nil, ErrNaoImplementado
}

// StatusServico consulta a disponibilidade do servico da SEFAZ.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func (c *Componente) StatusServico(ctx context.Context) error {
	return ErrNaoImplementado
}

// Cancelamento envia o evento de cancelamento de um documento.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado.
func (c *Componente) Cancelamento(ctx context.Context, chave, protocolo, justificativa string) (*RetEventoNFGas, error) {
	return nil, ErrNaoImplementado
}

// EnviarEvento transmite um evento a SEFAZ.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado. Para ler um evento ja
// obtido, use LerEvento.
func (c *Componente) EnviarEvento(ctx context.Context, e *EventoNFGas) (*RetEventoNFGas, error) {
	return nil, ErrNaoImplementado
}

// URLConsultaNFGas devolve a URL publica de consulta do documento.
//
// NAO IMPLEMENTADO: devolve ErrNaoImplementado. Depende da tabela de URLs
// por UF de ACBrNFGasServicos.ini.
func (c *Componente) URLConsultaNFGas(cUF int) (string, error) {
	return "", ErrNaoImplementado
}
