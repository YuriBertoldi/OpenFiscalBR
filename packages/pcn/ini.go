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

package pcn

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// INI e um arquivo .ini em memoria, equivalente ao TMemIniFile que os
// componentes ACBr usam para trocar documentos em formato texto.
//
// Implementado aqui, e nao com uma biblioteca de terceiros, porque o
// OpenFiscalBR nao tem go.sum e tanto a demo quanto a skill de validacao
// assumem zero dependencia externa.
//
// Como no Delphi, nomes de secao e de chave sao comparados sem distinguir
// maiusculas de minusculas, e a ordem de insercao e preservada na escrita.
type INI struct {
	secoes []*secaoINI
	indice map[string]*secaoINI
}

type secaoINI struct {
	nome   string
	chaves []parINI
	indice map[string]int
}

type parINI struct {
	chave string
	valor string
}

// NovoINI cria um arquivo INI vazio.
func NovoINI() *INI {
	return &INI{indice: make(map[string]*secaoINI)}
}

// LerINI carrega um arquivo INI de r.
func LerINI(r io.Reader) (*INI, error) {
	if r == nil {
		return nil, ErrXMLVazio
	}
	ini := NovoINI()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	atual := ""
	for sc.Scan() {
		linha := strings.TrimSpace(strings.TrimPrefix(sc.Text(), bomUTF8))
		if linha == "" || strings.HasPrefix(linha, ";") || strings.HasPrefix(linha, "#") {
			continue
		}
		if strings.HasPrefix(linha, "[") && strings.HasSuffix(linha, "]") {
			atual = strings.TrimSpace(linha[1 : len(linha)-1])
			ini.criarSecao(atual)
			continue
		}
		p := strings.Index(linha, "=")
		if p < 0 || atual == "" {
			continue
		}
		chave := strings.TrimSpace(linha[:p])
		valor := strings.TrimSpace(linha[p+1:])
		ini.GravarString(atual, chave, valor)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("pcn: leitura do INI: %w", err)
	}
	return ini, nil
}

// LerINIString carrega um arquivo INI a partir de uma string.
func LerINIString(s string) (*INI, error) {
	return LerINI(strings.NewReader(s))
}

// LerINIArquivoOuString decide se o parametro e um caminho de arquivo ou o
// proprio conteudo do INI, e carrega de acordo.
// Porte de LerIniArquivoOuString (ACBrUtil).
//
// O criterio e o mesmo do original: se nao houver quebra de linha e o
// arquivo existir no disco, trata como caminho.
func LerINIArquivoOuString(s string) (*INI, error) {
	if !strings.ContainsAny(s, "\r\n") {
		if info, err := os.Stat(s); err == nil && !info.IsDir() {
			f, err := os.Open(s)
			if err != nil {
				return nil, fmt.Errorf("pcn: abrir INI %q: %w", s, err)
			}
			defer f.Close()
			return LerINI(f)
		}
	}
	return LerINIString(s)
}

func (i *INI) criarSecao(nome string) *secaoINI {
	if i.indice == nil {
		i.indice = make(map[string]*secaoINI)
	}
	chave := strings.ToLower(nome)
	if s, ok := i.indice[chave]; ok {
		return s
	}
	s := &secaoINI{nome: nome, indice: make(map[string]int)}
	i.secoes = append(i.secoes, s)
	i.indice[chave] = s
	return s
}

func (i *INI) secao(nome string) *secaoINI {
	if i == nil || i.indice == nil {
		return nil
	}
	return i.indice[strings.ToLower(nome)]
}

// SecaoExiste informa se a secao existe. Porte de SectionExists.
//
// E a condicao que o ACBr usa para tornar um grupo inteiro opcional --
// varios leitores fazem "if not SectionExists then Exit".
func (i *INI) SecaoExiste(nome string) bool {
	return i.secao(nome) != nil
}

// Secoes devolve os nomes das secoes na ordem em que aparecem.
func (i *INI) Secoes() []string {
	if i == nil {
		return nil
	}
	r := make([]string, 0, len(i.secoes))
	for _, s := range i.secoes {
		r = append(r, s.nome)
	}
	return r
}

// Chaves devolve os nomes das chaves da secao, na ordem de insercao.
func (i *INI) Chaves(secao string) []string {
	s := i.secao(secao)
	if s == nil {
		return nil
	}
	r := make([]string, 0, len(s.chaves))
	for _, p := range s.chaves {
		r = append(r, p.chave)
	}
	return r
}

// ---------------------------------------------------------------------------
// Leitura
// ---------------------------------------------------------------------------

// LerString devolve o valor da chave, ou padrao se secao ou chave nao
// existirem. Porte de ReadString.
func (i *INI) LerString(secao, chave, padrao string) string {
	s := i.secao(secao)
	if s == nil {
		return padrao
	}
	pos, ok := s.indice[strings.ToLower(chave)]
	if !ok {
		return padrao
	}
	return s.chaves[pos].valor
}

// LerInteiro devolve o valor da chave como inteiro. Porte de ReadInteger.
func (i *INI) LerInteiro(secao, chave string, padrao int) int {
	v := strings.TrimSpace(i.LerString(secao, chave, ""))
	if v == "" {
		return padrao
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return padrao
	}
	return n
}

// LerFloat devolve o valor da chave como float64, aceitando virgula ou
// ponto como separador decimal. Porte de ReadFloat/StringToFloatDef.
func (i *INI) LerFloat(secao, chave string, padrao float64) float64 {
	v := strings.TrimSpace(i.LerString(secao, chave, ""))
	if v == "" {
		return padrao
	}
	return StringToFloatDef(v, padrao)
}

// LerData devolve o valor da chave como data/hora, aceitando as mesmas
// formas de EncodeDataHora.
func (i *INI) LerData(secao, chave string, padrao time.Time) time.Time {
	v := strings.TrimSpace(i.LerString(secao, chave, ""))
	if v == "" {
		return padrao
	}
	t, err := EncodeDataHora(v)
	if err != nil || t.IsZero() {
		return padrao
	}
	return t
}

// LerBool devolve o valor da chave como booleano. Aceita "1", "true", "sim"
// e "s" como verdadeiro, em qualquer caixa.
func (i *INI) LerBool(secao, chave string, padrao bool) bool {
	v := strings.ToLower(strings.TrimSpace(i.LerString(secao, chave, "")))
	switch v {
	case "":
		return padrao
	case "1", "true", "sim", "s", "t":
		return true
	case "0", "false", "nao", "n", "f":
		return false
	default:
		return padrao
	}
}

// ---------------------------------------------------------------------------
// Escrita
// ---------------------------------------------------------------------------

// GravarString grava o valor na chave, criando a secao se preciso.
// Porte de WriteString.
func (i *INI) GravarString(secao, chave, valor string) {
	s := i.criarSecao(secao)
	k := strings.ToLower(chave)
	if pos, ok := s.indice[k]; ok {
		s.chaves[pos].valor = valor
		return
	}
	s.indice[k] = len(s.chaves)
	s.chaves = append(s.chaves, parINI{chave: chave, valor: valor})
}

// GravarInteiro grava um inteiro. Porte de WriteInteger.
func (i *INI) GravarInteiro(secao, chave string, valor int) {
	i.GravarString(secao, chave, strconv.Itoa(valor))
}

// GravarFloat grava um float com ponto como separador decimal e sem
// separador de milhar, no numero de casas informado.
//
// Divergencia documentada: WriteFloat do Delphi usa o separador decimal do
// sistema, o que faz o mesmo documento sair diferente em cada maquina.
// Aqui o separador e sempre ponto, e a leitura aceita os dois.
func (i *INI) GravarFloat(secao, chave string, valor float64, casas int) {
	i.GravarString(secao, chave, strconv.FormatFloat(valor, 'f', casas, 64))
}

// GravarData grava uma data no formato informado. Data zerada nao e gravada.
func (i *INI) GravarData(secao, chave string, valor time.Time, layout string) {
	if valor.IsZero() {
		return
	}
	i.GravarString(secao, chave, valor.Format(layout))
}

// GravarBool grava um booleano como "1" ou "0".
func (i *INI) GravarBool(secao, chave string, valor bool) {
	if valor {
		i.GravarString(secao, chave, "1")
		return
	}
	i.GravarString(secao, chave, "0")
}

// String serializa o INI no formato texto, com as secoes e chaves na ordem
// de insercao e terminador de linha CRLF -- o mesmo que o TMemIniFile
// produz no Windows, onde os arquivos sao consumidos.
func (i *INI) String() string {
	if i == nil {
		return ""
	}
	var b strings.Builder
	for _, s := range i.secoes {
		b.WriteString("[" + s.nome + "]\r\n")
		for _, p := range s.chaves {
			b.WriteString(p.chave + "=" + p.valor + "\r\n")
		}
		b.WriteString("\r\n")
	}
	return b.String()
}

// SalvarArquivo grava o INI no caminho informado.
func (i *INI) SalvarArquivo(caminho string) error {
	if caminho == "" {
		return fmt.Errorf("pcn: caminho do INI nao informado")
	}
	if err := os.WriteFile(caminho, []byte(i.String()), 0o644); err != nil {
		return fmt.Errorf("pcn: gravar INI %q: %w", caminho, err)
	}
	return nil
}
