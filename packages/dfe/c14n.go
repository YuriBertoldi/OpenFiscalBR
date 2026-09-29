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

package dfe

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Canonical XML 1.0 sem comentarios (http://www.w3.org/TR/2001/REC-xml-c14n-20010315),
// o algoritmo exigido pelo XMLDSig dos DFe. No ACBr esse trabalho e delegado
// as bibliotecas SSL (libxml2/xmlsec via TDFeSSL); aqui e implementado sobre
// o encoding/xml, cobrindo o subconjunto de XML que ocorre em documento
// fiscal: elementos com ou sem prefixo, atributos SEM prefixo (alem de
// xmlns/xmlns:* e xml:*), texto, CDATA e comentarios (descartados).
// Atributo com outro prefixo devolve erro em vez de canonicalizar errado.

// nsXML e a URI fixa do prefixo reservado xml:.
const nsXML = "http://www.w3.org/XML/1998/namespace"

// CanonicalizarFragmento aplica Canonical XML 1.0 (sem comentarios) a um
// fragmento XML extraido de um documento. nsHerdado e o mapa prefixo->URI
// em escopo no ponto de onde o fragmento saiu (prefixo "" = namespace
// default); para DFe e tipicamente {"": "http://www.portalfiscal.inf.br/nfgas"}.
// removerAssinatura aplica o transform enveloped-signature, omitindo todo
// elemento Signature do namespace XMLDSig.
func CanonicalizarFragmento(frag string, nsHerdado map[string]string, removerAssinatura bool) (string, error) {
	dados := []byte(frag)
	dec := xml.NewDecoder(bytes.NewReader(dados))

	// escopo herdado (nunca renderizado por um ancestral, ja que o fragmento
	// foi destacado do documento)
	base := map[string]string{}
	for p, uri := range nsHerdado {
		if uri != "" {
			base[p] = uri
		}
	}

	var saida strings.Builder
	var pilhaEscopo []map[string]string // in-scope por nivel
	var pilhaRender []map[string]string // ja renderizado por nivel
	var pilhaNomes []string             // qname por nivel
	pilhaEscopo = append(pilhaEscopo, base)
	pilhaRender = append(pilhaRender, map[string]string{})

	var anterior int64
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("dfe: canonicalizacao: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if removerAssinatura && t.Name.Local == "Signature" && t.Name.Space == nsXMLDSig {
				if err := dec.Skip(); err != nil {
					return "", fmt.Errorf("dfe: canonicalizacao: %w", err)
				}
				break
			}

			qname := qualNomeBytes(dados, anterior, t.Name.Local)
			escopoPai := pilhaEscopo[len(pilhaEscopo)-1]
			renderPai := pilhaRender[len(pilhaRender)-1]

			// separa declaracoes de namespace de atributos comuns
			escopo := copiarMapa(escopoPai)
			var attrs []Atributo14n
			for _, a := range t.Attr {
				switch {
				case a.Name.Local == "xmlns" && a.Name.Space == "":
					escopo[""] = a.Value
				case a.Name.Space == "xmlns":
					escopo[a.Name.Local] = a.Value
				case a.Name.Space == "":
					attrs = append(attrs, Atributo14n{URI: "", Nome: a.Name.Local, Valor: a.Value})
				case a.Name.Space == "xml" || a.Name.Space == nsXML:
					attrs = append(attrs, Atributo14n{URI: nsXML, Nome: "xml:" + a.Name.Local, Valor: a.Value})
				default:
					return "", fmt.Errorf("dfe: canonicalizacao: atributo com prefixo nao suportado: %s:%s", a.Name.Space, a.Name.Local)
				}
			}

			// namespaces a renderizar: todo prefixo em escopo cujo valor
			// difere do que o ancestral mais proximo ja rendeu
			render := copiarMapa(renderPai)
			var prefixos []string
			for p, uri := range escopo {
				if renderPai[p] != uri {
					prefixos = append(prefixos, p)
					render[p] = uri
				}
			}
			sort.Strings(prefixos) // "" (default) ordena primeiro

			// atributos ordenados por (URI do namespace, local name)
			sort.Slice(attrs, func(i, j int) bool {
				if attrs[i].URI != attrs[j].URI {
					return attrs[i].URI < attrs[j].URI
				}
				return attrs[i].Nome < attrs[j].Nome
			})

			saida.WriteByte('<')
			saida.WriteString(qname)
			for _, p := range prefixos {
				saida.WriteByte(' ')
				if p == "" {
					saida.WriteString(`xmlns="`)
				} else {
					saida.WriteString("xmlns:" + p + `="`)
				}
				saida.WriteString(escaparAttr14n(escopo[p]))
				saida.WriteByte('"')
			}
			for _, a := range attrs {
				saida.WriteByte(' ')
				saida.WriteString(a.Nome)
				saida.WriteString(`="`)
				saida.WriteString(escaparAttr14n(a.Valor))
				saida.WriteByte('"')
			}
			saida.WriteByte('>')

			pilhaEscopo = append(pilhaEscopo, escopo)
			pilhaRender = append(pilhaRender, render)
			pilhaNomes = append(pilhaNomes, qname)

		case xml.EndElement:
			if len(pilhaNomes) == 0 {
				return "", fmt.Errorf("dfe: canonicalizacao: fechamento sem abertura de <%s>", t.Name.Local)
			}
			saida.WriteString("</")
			saida.WriteString(pilhaNomes[len(pilhaNomes)-1])
			saida.WriteByte('>')
			pilhaNomes = pilhaNomes[:len(pilhaNomes)-1]
			pilhaEscopo = pilhaEscopo[:len(pilhaEscopo)-1]
			pilhaRender = pilhaRender[:len(pilhaRender)-1]

		case xml.CharData:
			// texto fora do elemento raiz do fragmento nao existe em C14N
			// de subtree (so whitespace de formatacao do documento externo)
			if len(pilhaNomes) > 0 {
				saida.WriteString(escaparTexto14n(string(t)))
			}
			// comentarios e process instructions sao descartados (variante
			// "omit comments")
		}

		anterior = dec.InputOffset()
	}

	return saida.String(), nil
}

// Atributo14n e um atributo ja resolvido para a ordenacao do C14N.
type Atributo14n struct {
	URI   string
	Nome  string // como sera escrito (inclui prefixo xml: quando houver)
	Valor string
}

func copiarMapa(m map[string]string) map[string]string {
	c := make(map[string]string, len(m)+2)
	for k, v := range m {
		c[k] = v
	}
	return c
}

// qualNomeBytes recupera o nome do elemento como escrito no XML (com o
// prefixo original) a partir do offset do '<' de abertura. O decoder resolve
// o prefixo para a URI e nao devolve a grafia original.
func qualNomeBytes(dados []byte, inicio int64, local string) string {
	i := int(inicio)
	for i < len(dados) && dados[i] != '<' {
		i++
	}
	i++
	j := i
	for j < len(dados) {
		c := dados[j]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '>' || c == '/' {
			break
		}
		j++
	}
	if j > i {
		return string(dados[i:j])
	}
	return local
}

// escaparTexto14n escapa texto conforme o C14N: & < > e CR.
func escaparTexto14n(s string) string {
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
		case '\r':
			b.WriteString("&#xD;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// escaparAttr14n escapa valor de atributo conforme o C14N: & < " TAB LF CR.
func escaparAttr14n(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '"':
			b.WriteString("&quot;")
		case '\t':
			b.WriteString("&#x9;")
		case '\n':
			b.WriteString("&#xA;")
		case '\r':
			b.WriteString("&#xD;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
