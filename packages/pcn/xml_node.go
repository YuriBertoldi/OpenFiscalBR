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
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Atributo e um atributo de elemento XML. Porte de TACBrXmlAttribute.
type Atributo struct {
	Nome  string // local name
	Space string // URI do namespace, vazio quando sem prefixo
	Valor string
}

// Node e um elemento do documento XML. Porte de TACBrXmlNode.
//
// Todos os metodos de navegacao sao seguros em receptor nil: chamar
// FindAnyNs sobre um node inexistente devolve nil em vez de estourar, o que
// permite encadear a busca do mesmo jeito que o Delphi faz com
// Childrens.FindAnyNs(...).Childrens.FindAnyNs(...) -- so que sem a access
// violation que o original produz quando o caminho nao existe.
type Node struct {
	// Nome e o local name do elemento, sem prefixo. Corresponde a
	// TACBrXmlNode.LocalName, comparado por FindAnyNs/FindAllAnyNs.
	Nome string
	// NomeQualificado e o nome como escrito no XML, com prefixo se houver
	// ("Signature", "ds:Signature"). Corresponde a TACBrXmlNode.Name,
	// comparado por Find/FindAll.
	NomeQualificado string
	// Space e a URI do namespace resolvido do elemento.
	Space string

	Atributos []Atributo
	Filhos    []*Node
	Texto     string // conteudo textual direto do elemento, sem Trim

	// FloatIsIntString reproduz TACBrXmlNode.FloatIsIntString: quando true,
	// os tipos TcDe1..TcDe10 interpretam o conteudo como inteiro com casas
	// decimais implicitas ("1234" com TcDe2 vale 12,34). Falso por padrao,
	// que e o caso de todo XML de DFe da SEFAZ.
	FloatIsIntString bool

	doc    *Document
	inicio int64 // offset do '<' de abertura no XML de origem
	fim    int64 // offset logo apos o '>' de fechamento
	pai    *Node
}

// Document e um documento XML ja parseado. Porte de TACBrXmlDocument.
type Document struct {
	Root *Node

	fonte []byte
}

// Parse le um documento XML de r.
func Parse(r io.Reader) (*Document, error) {
	if r == nil {
		return nil, ErrXMLVazio
	}
	dados, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("pcn: leitura do XML: %w", err)
	}
	return ParseBytes(dados)
}

// ParseString le um documento XML a partir de uma string.
func ParseString(s string) (*Document, error) {
	return ParseBytes([]byte(s))
}

// ParseBytes le um documento XML a partir de bytes. O BOM UTF-8 e a
// declaracao <?xml ...?> sao tolerados.
func ParseBytes(dados []byte) (*Document, error) {
	dados = RemoverBOM(dados)
	if len(bytes.TrimSpace(dados)) == 0 {
		return nil, ErrXMLVazio
	}

	doc := &Document{fonte: dados}
	dec := xml.NewDecoder(bytes.NewReader(dados))

	var pilha []*Node
	// Offset do fim do token anterior. Como os tokens do decoder sao
	// contiguos, ele e exatamente o offset do '<' do proximo StartElement.
	var anterior int64

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrXMLInvalido, err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			n := &Node{
				Nome:   t.Name.Local,
				Space:  t.Name.Space,
				doc:    doc,
				inicio: anterior,
			}
			n.NomeQualificado = nomeQualificado(dados, anterior, t.Name.Local)
			for _, a := range t.Attr {
				// O decoder do Go entrega as declaracoes de namespace como
				// atributos comuns; o ACBr nao as expoe, entao ficam de fora.
				if a.Name.Local == "xmlns" || a.Name.Space == "xmlns" {
					continue
				}
				n.Atributos = append(n.Atributos, Atributo{
					Nome:  a.Name.Local,
					Space: a.Name.Space,
					Valor: a.Value,
				})
			}
			if len(pilha) == 0 {
				if doc.Root != nil {
					return nil, fmt.Errorf("%w: mais de um elemento raiz", ErrXMLInvalido)
				}
				doc.Root = n
			} else {
				pai := pilha[len(pilha)-1]
				n.pai = pai
				pai.Filhos = append(pai.Filhos, n)
			}
			pilha = append(pilha, n)

		case xml.EndElement:
			if len(pilha) == 0 {
				return nil, fmt.Errorf("%w: fechamento sem abertura de <%s>", ErrXMLInvalido, t.Name.Local)
			}
			pilha[len(pilha)-1].fim = dec.InputOffset()
			pilha = pilha[:len(pilha)-1]

		case xml.CharData:
			if len(pilha) > 0 {
				pilha[len(pilha)-1].Texto += string(t)
			}
		}

		anterior = dec.InputOffset()
	}

	if doc.Root == nil {
		return nil, ErrXMLVazio
	}
	return doc, nil
}

// nomeQualificado recupera o nome do elemento como escrito no XML (com
// prefixo, se houver) a partir do offset de abertura. O decoder do Go ja
// resolveu o prefixo para a URI e nao o devolve, mas Find/FindAll do ACBr
// comparam exatamente esse nome -- por isso ele e relido da fonte.
func nomeQualificado(fonte []byte, inicio int64, localName string) string {
	if inicio < 0 || inicio >= int64(len(fonte)) {
		return localName
	}
	i := int(inicio)
	if fonte[i] != '<' {
		j := bytes.IndexByte(fonte[i:], '<')
		if j < 0 {
			return localName
		}
		i += j
	}
	i++
	ini := i
	for i < len(fonte) {
		switch fonte[i] {
		case ' ', '\t', '\r', '\n', '/', '>':
			return string(fonte[ini:i])
		}
		i++
	}
	return localName
}

// XML devolve o documento de origem, sem BOM.
func (d *Document) XML() string {
	if d == nil {
		return ""
	}
	return string(d.fonte)
}

// ---------------------------------------------------------------------------
// Navegacao
// ---------------------------------------------------------------------------

// Find procura o primeiro filho direto cujo nome qualificado seja igual ao
// informado. Porte de TACBrXmlNodeList.Find, que compara TACBrXmlNode.Name
// -- ou seja, o nome COM prefixo. Buscar "Signature" acha
// <Signature xmlns="...xmldsig#"> mas nao acha <ds:Signature>.
func (n *Node) Find(nome string) *Node {
	if n == nil {
		return nil
	}
	for _, f := range n.Filhos {
		if f.NomeQualificado == nome {
			return f
		}
	}
	return nil
}

// FindAnyNs procura o primeiro filho direto pelo local name, ignorando
// prefixo e namespace. Porte de TACBrXmlNodeList.FindAnyNs, o metodo usado
// pela esmagadora maioria dos leitores do ACBr.
func (n *Node) FindAnyNs(nome string) *Node {
	if n == nil {
		return nil
	}
	for _, f := range n.Filhos {
		if f.Nome == nome {
			return f
		}
	}
	return nil
}

// FindAll devolve todos os filhos diretos cujo nome qualificado seja igual
// ao informado. Porte de TACBrXmlNodeList.FindAll.
func (n *Node) FindAll(nome string) []*Node {
	if n == nil {
		return nil
	}
	var r []*Node
	for _, f := range n.Filhos {
		if f.NomeQualificado == nome {
			r = append(r, f)
		}
	}
	return r
}

// FindAllAnyNs devolve todos os filhos diretos pelo local name, ignorando
// prefixo e namespace. Porte de TACBrXmlNodeList.FindAllAnyNs.
func (n *Node) FindAllAnyNs(nome string) []*Node {
	if n == nil {
		return nil
	}
	var r []*Node
	for _, f := range n.Filhos {
		if f.Nome == nome {
			r = append(r, f)
		}
	}
	return r
}

// PrimeiroDe devolve o primeiro filho cujo local name esteja na lista, na
// ordem em que os nomes foram informados -- e nao na ordem do documento.
//
// E o padrao de grupo achatado por CST: o leitor procura ICMS00, depois
// ICMS10, depois ICMS20... e usa o primeiro que existir.
func (n *Node) PrimeiroDe(nomes ...string) *Node {
	if n == nil {
		return nil
	}
	for _, nome := range nomes {
		if f := n.FindAnyNs(nome); f != nil {
			return f
		}
	}
	return nil
}

// Pai devolve o elemento pai, ou nil na raiz.
func (n *Node) Pai() *Node {
	if n == nil {
		return nil
	}
	return n.pai
}

// Existe informa se o node foi encontrado. Equivale a n != nil, mas le
// melhor em condicional encadeada.
func (n *Node) Existe() bool { return n != nil }

// Attr devolve o valor do atributo, ja com Trim. Devolve vazio quando o node
// ou o atributo nao existem. Porte de ObterConteudoTag(TACBrXmlAttribute).
func (n *Node) Attr(nome string) string {
	v, _ := n.AttrOk(nome)
	return v
}

// AttrOk devolve o valor do atributo e se ele existe.
func (n *Node) AttrOk(nome string) (string, bool) {
	if n == nil {
		return "", false
	}
	for _, a := range n.Atributos {
		if a.Nome == nome {
			return strings.TrimSpace(a.Valor), true
		}
	}
	return "", false
}

// Conteudo devolve o texto do elemento ja com Trim, ou vazio se o node nao
// existe.
func (n *Node) Conteudo() string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.Texto)
}

// OuterXML devolve o trecho exato do XML de origem correspondente a este
// elemento, incluindo as proprias tags. Porte de TACBrXmlNode.OuterXml.
//
// Diferente de uma reserializacao, o retorno e byte a byte igual ao que
// entrou -- o que importa para reaproveitar o trecho (protNFGas,
// procEventoNFGas) em outra leitura ou na conferencia de assinatura.
func (n *Node) OuterXML() string {
	if n == nil || n.doc == nil {
		return ""
	}
	if n.inicio < 0 || n.fim <= n.inicio || n.fim > int64(len(n.doc.fonte)) {
		return ""
	}
	return string(n.doc.fonte[n.inicio:n.fim])
}

// Caminho devolve o caminho do elemento a partir da raiz, no formato
// "NFGas/infNFGas/ide/dhEmi". Usado para contextualizar erros de leitura.
func (n *Node) Caminho() string {
	if n == nil {
		return ""
	}
	var partes []string
	for atual := n; atual != nil; atual = atual.pai {
		partes = append(partes, atual.Nome)
	}
	for i, j := 0, len(partes)-1; i < j; i, j = i+1, j-1 {
		partes[i], partes[j] = partes[j], partes[i]
	}
	return strings.Join(partes, "/")
}

// bomUTF8 e a marca de ordem de byte UTF-8, escrita em escape hexadecimal
// para nao aparecer como BOM literal no proprio fonte Go.
const bomUTF8 = "\xef\xbb\xbf"

// RemoverBOM remove o BOM UTF-8 do inicio dos dados, se houver.
// Porte de RemoverUTF8Bom (ACBrUtil.Strings).
func RemoverBOM(dados []byte) []byte {
	if bytes.HasPrefix(dados, []byte(bomUTF8)) {
		return dados[len(bomUTF8):]
	}
	return dados
}
