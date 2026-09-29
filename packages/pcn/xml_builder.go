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
	"strconv"
	"strings"
	"time"
)

// Construtor de XML para a GERACAO de documentos fiscais.
// Porte do papel de TACBrXmlDocument/TACBrXmlNode na escrita
// (CreateElement/AppendChild/SetAttribute) mais os formatadores de campo do
// TACBrXmlWriter.AddNode.
//
// A saida e sempre COMPACTA (sem identacao nem quebras) -- e assim que o
// ACBr transmite, e e o que mantem o digest da assinatura estavel.

// Elem e um elemento XML em construcao.
type Elem struct {
	Nome   string
	attrs  []Atributo
	filhos []*Elem
	texto  string
	// bruto desliga o escape do texto -- usado quando o conteudo ja carrega
	// marcacao propria, como o CDATA do qrCodNFGas.
	bruto bool
}

// NovoElem cria um elemento com o nome informado.
func NovoElem(nome string) *Elem { return &Elem{Nome: nome} }

// Attr acrescenta um atributo, preservando a ordem de insercao.
// Devolve o proprio elemento para encadear.
func (e *Elem) Attr(nome, valor string) *Elem {
	if e == nil {
		return e
	}
	e.attrs = append(e.attrs, Atributo{Nome: nome, Valor: valor})
	return e
}

// Filho acrescenta um elemento filho. Filho nil e ignorado -- e o
// equivalente do AppendChild(nil) que o ACBr faz quando o gerador do grupo
// devolve nil por o grupo estar vazio.
func (e *Elem) Filho(filho *Elem) *Elem {
	if e == nil || filho == nil {
		return e
	}
	e.filhos = append(e.filhos, filho)
	return e
}

// Filhos acrescenta varios filhos, ignorando os nil.
func (e *Elem) Filhos(filhos ...*Elem) *Elem {
	for _, f := range filhos {
		e.Filho(f)
	}
	return e
}

// Texto define o conteudo textual do elemento (escapado na serializacao).
func (e *Elem) Texto(s string) *Elem {
	if e == nil {
		return e
	}
	e.texto = s
	return e
}

// TextoBruto define conteudo que NAO sera escapado (ex.: CDATA literal).
func (e *Elem) TextoBruto(s string) *Elem {
	if e == nil {
		return e
	}
	e.texto = s
	e.bruto = true
	return e
}

// TemFilhos informa se algum filho foi acrescentado.
func (e *Elem) TemFilhos() bool { return e != nil && len(e.filhos) > 0 }

// XML serializa o elemento de forma compacta.
func (e *Elem) XML() string {
	var b strings.Builder
	e.escrever(&b)
	return b.String()
}

func (e *Elem) escrever(b *strings.Builder) {
	if e == nil {
		return
	}
	b.WriteByte('<')
	b.WriteString(e.Nome)
	for _, a := range e.attrs {
		b.WriteByte(' ')
		b.WriteString(a.Nome)
		b.WriteString(`="`)
		b.WriteString(EscaparAtributoXML(a.Valor))
		b.WriteByte('"')
	}
	if len(e.filhos) == 0 && e.texto == "" {
		b.WriteString("/>")
		return
	}
	b.WriteByte('>')
	if e.texto != "" {
		if e.bruto {
			b.WriteString(e.texto)
		} else {
			b.WriteString(EscaparTextoXML(e.texto))
		}
	}
	for _, f := range e.filhos {
		f.escrever(b)
	}
	b.WriteString("</")
	b.WriteString(e.Nome)
	b.WriteByte('>')
}

// EscaparTextoXML escapa conteudo de elemento: & < >.
func EscaparTextoXML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// EscaparAtributoXML escapa valor de atributo: & < > " '.
func EscaparAtributoXML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&#39;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Formatadores de campo (lado da geracao)
// ---------------------------------------------------------------------------

// FormatarDecimalXML formata um float com o numero FIXO de casas decimais e
// ponto como separador. Porte do AddNode tcDeN, que usa
// FloatToString(valor, '.', FloatMask(N)).
func FormatarDecimalXML(v float64, casas int) string {
	return strconv.FormatFloat(v, 'f', casas, 64)
}

// FormatarVersaoXML formata o atributo versao com duas casas ("1.00").
// Porte de FloatToString(Versao, '.', '#0.00').
func FormatarVersaoXML(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// FormatarDataXML devolve AAAA-MM-DD; tempo zero devolve vazio.
// Porte do AddNode tcDat.
func FormatarDataXML(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatarDataHoraXML devolve AAAA-MM-DDTHH:MM:SS acrescido do fuso:
// o do proprio time.Time quando presente, ou o fuso oficial da UF quando o
// valor esta em UTC "de parede". Porte de DateTimeTodh + GetUTC(UF).
func FormatarDataHoraXML(t time.Time, uf string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02T15:04:05") + fusoDe(t, uf)
}

func fusoDe(t time.Time, uf string) string {
	if _, off := t.Zone(); off != 0 {
		sinal := "+"
		if off < 0 {
			sinal = "-"
			off = -off
		}
		return sinal + zero2(off/3600) + ":" + zero2((off%3600)/60)
	}
	return OffsetUF(uf)
}

func zero2(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}

// OffsetUF devolve o fuso horario oficial da UF no formato "-03:00".
// Porte de GetUTCUF (ACBrUtil.DateTime), sem o horario de verao -- abolido
// pelo Decreto 9.772/2019 e ja tratado como inexistente para datas atuais.
func OffsetUF(uf string) string {
	switch strings.ToUpper(strings.TrimSpace(uf)) {
	case "AC":
		return "-05:00"
	case "AM", "RR", "RO", "MT", "MS":
		return "-04:00"
	default:
		// Brasilia. UFs vazias e os codigos 90/91 tambem caem aqui,
		// como no original.
		return "-03:00"
	}
}

// FiltrarTextoXML aplica o tratamento que o TACBrXmlWriter da a todo campo
// texto antes de gravar, com as opcoes DEFAULT do ACBr: remove acentos,
// colapsa espacos duplicados e troca quebras de linha por ";".
// Porte de FiltrarTextoXML (ACBrXmlBase) com RetirarEspacos=True,
// RetirarAcentos=True e QuebraLinha=';'.
func FiltrarTextoXML(s string) string {
	s = TiraAcentos(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	s = strings.ReplaceAll(s, "\r\n", ";")
	s = strings.ReplaceAll(s, "\r", ";")
	s = strings.ReplaceAll(s, "\n", ";")
	return strings.TrimSpace(s)
}

// TiraAcentos translitera caracteres acentuados para ASCII.
// Porte de TiraAcentos (ACBrUtil.Strings).
func TiraAcentos(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if sub, ok := semAcento[r]; ok {
			b.WriteString(sub)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var semAcento = map[rune]string{
	'á': "a", 'à': "a", 'ã': "a", 'â': "a", 'ä': "a",
	'é': "e", 'è': "e", 'ê': "e", 'ë': "e",
	'í': "i", 'ì': "i", 'î': "i", 'ï': "i",
	'ó': "o", 'ò': "o", 'õ': "o", 'ô': "o", 'ö': "o",
	'ú': "u", 'ù': "u", 'û': "u", 'ü': "u",
	'ç': "c", 'ñ': "n", 'ý': "y",
	'Á': "A", 'À': "A", 'Ã': "A", 'Â': "A", 'Ä': "A",
	'É': "E", 'È': "E", 'Ê': "E", 'Ë': "E",
	'Í': "I", 'Ì': "I", 'Î': "I", 'Ï': "I",
	'Ó': "O", 'Ò': "O", 'Õ': "O", 'Ô': "O", 'Ö': "O",
	'Ú': "U", 'Ù': "U", 'Û': "U", 'Ü': "U",
	'Ç': "C", 'Ñ': "N", 'Ý': "Y",
	'ª': "a", 'º': "o", '°': "o",
}

// PadLeftZeros preenche com zeros a esquerda ate o tamanho minimo -- o
// PadLeft(TamMin, '0') que o AddNode aplica a tcInt e tcNumStr.
func PadLeftZeros(s string, minimo int) string {
	for len(s) < minimo {
		s = "0" + s
	}
	return s
}

// ---------------------------------------------------------------------------
// Equivalentes do TACBrXmlWriter.AddNode por tipo de campo
// ---------------------------------------------------------------------------
//
// Semantica portada do AddNode (ACBrXmlWriter.pas):
//   - obrigatorio (ocorrencias=1) com valor vazio  -> gera a TAG VAZIA;
//   - opcional (ocorrencias=0) com valor vazio     -> devolve nil (nao gera);
//   - caso contrario gera a tag com o conteudo formatado.
// "Vazio" e: string vazia apos Trim (tcStr), zero (tcInt/tcDeN, so quando
// opcional) e data zero (tcDat).
//
// OMISSAO DELIBERADA: a ListaDeAlertas do ACBr (wAlerta de min/max/vazio)
// nao foi portada -- os alertas do original sao informativos e nao impedem a
// geracao; a validacao efetiva e a do XSD da SEFAZ e das regras de negocio.

// NodeStr e o AddNode tcStr: Trim + FiltrarTextoXML (acentos, espacos
// duplicados, quebras de linha).
func NodeStr(nome, valor string, obrigatorio bool) *Elem {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		if obrigatorio {
			return NovoElem(nome)
		}
		return nil
	}
	return NovoElem(nome).Texto(FiltrarTextoXML(valor))
}

// NodeStrSemFiltro e o AddNode tcStr com ParseTextoXML=False: o conteudo vai
// como esta (apenas escapado na serializacao). Usado para campos que nao
// podem sofrer o filtro, como chaves e URLs.
func NodeStrSemFiltro(nome, valor string, obrigatorio bool) *Elem {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		if obrigatorio {
			return NovoElem(nome)
		}
		return nil
	}
	return NovoElem(nome).Texto(valor)
}

// NodeInt e o AddNode tcInt: zero e vazio quando opcional; PadLeft de zeros
// ate o tamanho minimo.
func NodeInt(nome string, valor int, minimo int, obrigatorio bool) *Elem {
	if valor == 0 && !obrigatorio {
		return nil
	}
	return NovoElem(nome).Texto(PadLeftZeros(strconv.Itoa(valor), minimo))
}

// NodeDec e o AddNode tcDeN: numero de casas FIXO, ponto decimal; zero e
// vazio quando opcional, e obrigatorio gera "0.00...".
func NodeDec(nome string, valor float64, casas int, obrigatorio bool) *Elem {
	if valor == 0 && !obrigatorio {
		return nil
	}
	return NovoElem(nome).Texto(FormatarDecimalXML(valor, casas))
}

// NodeDat e o AddNode tcDat: AAAA-MM-DD; tempo zero e vazio.
func NodeDat(nome string, t time.Time, obrigatorio bool) *Elem {
	if t.IsZero() {
		if obrigatorio {
			return NovoElem(nome)
		}
		return nil
	}
	return NovoElem(nome).Texto(FormatarDataXML(t))
}
