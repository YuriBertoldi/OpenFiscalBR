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
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Carregamento de documentos, avulsos e em lote.
// Porte de ACBrNFGasNotasFiscais.pas.

// NotaFiscal e um documento carregado, com o XML de origem preservado.
// Porte de TNotaFiscal.
type NotaFiscal struct {
	// NFGas sao os dados do documento.
	NFGas *NFGas
	// XMLOriginal e o trecho exato do XML de onde a nota foi lida.
	XMLOriginal string
	// XMLAssinado e o XML gerado e assinado para transmissao.
	// Porte de TNotaFiscal.XMLAssinado; preenchido por Componente.Enviar
	// ou manualmente via GerarXML + Assinar.
	XMLAssinado string
	// NomeArq e o arquivo de origem, quando a leitura veio de disco.
	NomeArq string
	// Alertas acumula avisos nao fatais da leitura.
	Alertas []string
}

// ChaveAcesso devolve a chave de 44 posicoes, sem o literal "NFGas".
// Porte de TNotaFiscal.NumID.
func (n *NotaFiscal) ChaveAcesso() string {
	if n == nil || n.NFGas == nil {
		return ""
	}
	return n.NFGas.ChaveAcesso()
}

// CStat devolve o codigo de situacao do protocolo, ou 0 se o XML nao trazia
// protocolo. Porte de TNotaFiscal.cStat.
func (n *NotaFiscal) CStat() int {
	if n == nil || n.NFGas == nil {
		return 0
	}
	return n.NFGas.ProcNFGas.CStat
}

// Msg devolve o motivo informado no protocolo.
// Porte de TNotaFiscal.Msg.
func (n *NotaFiscal) Msg() string {
	if n == nil || n.NFGas == nil {
		return ""
	}
	return n.NFGas.ProcNFGas.XMotivo
}

// Confirmada informa se o documento foi confirmado pela SEFAZ.
func (n *NotaFiscal) Confirmada() bool {
	return n != nil && n.NFGas.Confirmada()
}

// Processada informa se o documento foi processado pela SEFAZ.
func (n *NotaFiscal) Processada() bool {
	return n != nil && n.NFGas.Processada()
}

// Cancelada informa se o documento foi cancelado.
func (n *NotaFiscal) Cancelada() bool {
	return n != nil && n.NFGas.Cancelada()
}

// Situacao resume o estado do documento numa palavra, para relatorio de
// importacao: "cancelada", "confirmada", "processada" ou "sem protocolo".
func (n *NotaFiscal) Situacao() string {
	switch {
	case n == nil || n.NFGas == nil:
		return ""
	case n.Cancelada():
		return "cancelada"
	case n.Confirmada():
		return "confirmada"
	case n.Processada():
		return "processada"
	case n.CStat() == 0:
		return "sem protocolo"
	default:
		return fmt.Sprintf("cStat %d", n.CStat())
	}
}

// CalcularNomeArquivo devolve o nome padrao do arquivo do documento.
// Porte de TNotaFiscal.CalcularNomeArquivo.
func (n *NotaFiscal) CalcularNomeArquivo() string {
	chave := n.ChaveAcesso()
	if chave == "" {
		return ""
	}
	return chave + "-NFGas.xml"
}

// Assinado informa se o XML de origem carrega assinatura digital.
func (n *NotaFiscal) Assinado() bool {
	return n != nil && n.NFGas != nil && !n.NFGas.Signature.Vazia()
}

// ---------------------------------------------------------------------------
// Leitura de um documento
// ---------------------------------------------------------------------------

// LerNota le um unico documento e devolve a NotaFiscal com o XML de origem
// preservado.
func LerNota(dados []byte) (*NotaFiscal, error) {
	dados = pcn.RemoverBOM(dados)
	n, err := LerBytes(dados)
	if err != nil {
		return nil, err
	}
	return &NotaFiscal{NFGas: n, XMLOriginal: string(dados)}, nil
}

// LerNotaArquivo le um unico documento de um arquivo.
func LerNotaArquivo(caminho string) (*NotaFiscal, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfgas: abrir %q: %w", caminho, err)
	}
	nota, err := LerNota(dados)
	if err != nil {
		return nil, &ErroNFGas{Indice: -1, Arquivo: filepath.Base(caminho), Err: err}
	}
	nota.NomeArq = caminho
	return nota, nil
}

// ---------------------------------------------------------------------------
// Leitura em lote
// ---------------------------------------------------------------------------

// LerLote le todos os documentos presentes em r.
//
// Aceita um arquivo com um documento, varios documentos concatenados
// (<NFGas> e <nfgasProc> misturados, com ou sem prologo <?xml?> entre eles)
// e documentos aninhados dentro de um envelope, como o retorno de consulta.
//
// Devolve as notas que puderam ser lidas SEMPRE, mesmo quando alguma falha:
// numa importacao de milhares de XMLs, uma nota torta nao pode invalidar as
// outras. As falhas vem no erro, do tipo *ErrosLote.
//
// DIVERGENCIA: o ACBr fatia o lote por busca de string ("</NFGas>",
// "</NFGasProc>", "</procNFGas>") com deslocamentos fixos e sensivel a
// caixa. Aqui o corte e feito por um decoder de XML de verdade, o que
// produz o mesmo resultado semantico -- um documento por elemento NFGas --
// sem depender da grafia do fechamento nem da ausencia de espacos.
func LerLote(r io.Reader) ([]*NotaFiscal, error) {
	if r == nil {
		return nil, ErrXMLVazio
	}
	dados, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("nfgas: leitura do lote: %w", err)
	}
	return LerLoteBytes(dados)
}

// LerLoteBytes le todos os documentos presentes nos bytes informados.
func LerLoteBytes(dados []byte) ([]*NotaFiscal, error) {
	dados = pcn.RemoverBOM(dados)
	if len(bytes.TrimSpace(dados)) == 0 {
		return nil, ErrXMLVazio
	}

	trechos, err := separarDocumentos(dados)
	if err != nil {
		return nil, err
	}
	if len(trechos) == 0 {
		return nil, ErrLoteVazio
	}

	notas := make([]*NotaFiscal, 0, len(trechos))
	var falhas []*ErroNFGas

	for i, t := range trechos {
		n, err := LerBytes(t)
		if err != nil {
			falhas = append(falhas, &ErroNFGas{
				Indice: i,
				Chave:  chaveAproximada(t),
				Err:    err,
			})
			continue
		}
		notas = append(notas, &NotaFiscal{NFGas: n, XMLOriginal: string(t)})
	}

	if len(falhas) > 0 {
		return notas, &ErrosLote{Erros: falhas}
	}
	return notas, nil
}

// LerLoteString le todos os documentos presentes na string.
func LerLoteString(xml string) ([]*NotaFiscal, error) {
	return LerLoteBytes([]byte(xml))
}

// LerLoteArquivo le todos os documentos de um arquivo e preenche NomeArq em
// cada nota. Porte de TNotasFiscais.LoadFromFile.
func LerLoteArquivo(caminho string) ([]*NotaFiscal, error) {
	dados, err := os.ReadFile(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfgas: abrir %q: %w", caminho, err)
	}
	notas, errLote := LerLoteBytes(dados)
	for _, n := range notas {
		n.NomeArq = caminho
	}
	if errLote != nil {
		// Acrescenta o nome do arquivo as falhas, que e o que permite
		// localizar o problema numa importacao de diretorio.
		if lote, ok := errLote.(*ErrosLote); ok {
			for _, f := range lote.Erros {
				f.Arquivo = filepath.Base(caminho)
			}
		}
		return notas, errLote
	}
	return notas, nil
}

// LerLoteDiretorio le todos os arquivos .xml de um diretorio, sem descer em
// subpastas. E o caminho de entrada tipico de importacao em lote.
//
// Devolve as notas lidas com sucesso e, se houver falhas, um *ErrosLote com
// uma entrada por documento problematico.
func LerLoteDiretorio(dir string) ([]*NotaFiscal, error) {
	entradas, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("nfgas: listar %q: %w", dir, err)
	}

	var notas []*NotaFiscal
	var falhas []*ErroNFGas

	for _, e := range entradas {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".xml") {
			continue
		}
		caminho := filepath.Join(dir, e.Name())
		lidas, err := LerLoteArquivo(caminho)
		notas = append(notas, lidas...)
		if err != nil {
			if lote, ok := err.(*ErrosLote); ok {
				falhas = append(falhas, lote.Erros...)
				continue
			}
			falhas = append(falhas, &ErroNFGas{Indice: -1, Arquivo: e.Name(), Err: err})
		}
	}

	if len(falhas) > 0 {
		return notas, &ErrosLote{Erros: falhas}
	}
	return notas, nil
}

// ---------------------------------------------------------------------------
// Separacao dos documentos do lote
// ---------------------------------------------------------------------------

// separarDocumentos varre os bytes e devolve o trecho exato de cada
// documento encontrado.
//
// Prefere nfgasProc: ao encontrar um, captura a subarvore inteira (que
// contem a NFGas e o protocolo) e salta para o fim dela, para nao contar a
// NFGas interna duas vezes. Um NFGas fora de nfgasProc e capturado sozinho.
//
// Por varrer em qualquer profundidade, funciona tambem quando os documentos
// vem dentro de um envelope -- retorno de consulta, resposta de web service
// ou um XML agregador qualquer.
func separarDocumentos(dados []byte) ([][]byte, error) {
	dec := xml.NewDecoder(bytes.NewReader(dados))

	var trechos [][]byte
	var anterior int64

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %s", pcn.ErrXMLInvalido, err)
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			anterior = dec.InputOffset()
			continue
		}

		if se.Name.Local != TagNFGasProc && se.Name.Local != TagNFGas {
			anterior = dec.InputOffset()
			continue
		}

		inicio := anterior
		if err := dec.Skip(); err != nil {
			return nil, fmt.Errorf("%w: %s", pcn.ErrXMLInvalido, err)
		}
		fim := dec.InputOffset()

		if inicio >= 0 && fim > inicio && fim <= int64(len(dados)) {
			trechos = append(trechos, dados[inicio:fim])
		}
		anterior = fim
	}

	return trechos, nil
}

// chaveAproximada tenta extrair a chave de acesso de um trecho que nao pode
// ser lido, so para contextualizar o erro. Nao valida nada.
func chaveAproximada(trecho []byte) string {
	const marca = `Id="`
	p := bytes.Index(trecho, []byte(marca))
	if p < 0 {
		return ""
	}
	resto := trecho[p+len(marca):]
	q := bytes.IndexByte(resto, '"')
	if q < 0 {
		return ""
	}
	return pcn.RemoverLiteralChave(string(resto[:q]))
}
