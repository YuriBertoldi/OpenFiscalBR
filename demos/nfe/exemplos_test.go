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

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/nfe"
)

// Testes da demo.
//
// O que eles travam e o CONTRATO ENTRE A DEMO E O PACKAGE: os exemplos são
// montados em Go e serializados na hora, então um campo que mude de
// significado no package passa a produzir documento inválido aqui sem
// quebrar a compilação. Os testes leem de volta o que os exemplos geram.

// TestExemplos_SaoLidosDeVoltaPeloPackage trava a ida e volta de cada
// exemplo do catálogo.
//
// Um exemplo que não volta a ser lido é pior que exemplo nenhum: quem está
// conhecendo a API cola o conteúdo em /api/ler, recebe erro, e conclui que
// a biblioteca é que está quebrada.
func TestExemplos_SaoLidosDeVoltaPeloPackage(t *testing.T) {
	for _, ex := range catalogo() {
		t.Run(ex.ID, func(t *testing.T) {
			conteudo, chave, err := ex.Montar()
			if err != nil {
				t.Fatalf("montar: %v", err)
			}
			if conteudo == "" {
				t.Fatal("conteudo vazio")
			}
			if chave == "" {
				t.Error("o exemplo deveria sair com a chave de acesso montada")
			}

			n, err := lerNota(conteudo)
			if err != nil {
				t.Fatalf("o proprio package nao le o exemplo de volta: %v", err)
			}
			if len(n.Det) == 0 {
				t.Error("nota sem item nenhum")
			}
			if n.Emit.CNPJCPF == "" {
				t.Error("nota sem emitente")
			}
		})
	}
}

// TestExemplos_ChaveDeAcessoFECHA trava o dígito verificador.
//
// A chave é derivada de oito campos da nota. Um exemplo com chave que não
// fecha é recusado pela SEFAZ e, pior, pelo próprio gravador de `.ini`
// deste package — que valida antes de gerar.
//
// O exemplo em `.ini` precisa de um passo a mais, e isso NÃO é defeito: o
// `ID` é uma das oito chaves que o ACBr **grava e não lê**. O leitor de
// `.ini` não restaura a chave; ela é derivada dos campos. Por isso o teste
// monta a chave quando ela vem vazia, em vez de exigir que tenha voltado.
func TestExemplos_ChaveDeAcessoFECHA(t *testing.T) {
	for _, ex := range catalogo() {
		t.Run(ex.ID, func(t *testing.T) {
			conteudo, chaveGerada, err := ex.Montar()
			if err != nil {
				t.Fatalf("montar: %v", err)
			}
			n, err := lerNota(conteudo)
			if err != nil {
				t.Fatalf("ler: %v", err)
			}

			if n.InfNFe.ID == "" {
				remontada, err := nfe.MontarChaveAcesso(n)
				if err != nil {
					t.Fatalf("remontar a chave: %v", err)
				}
				// A chave remontada a partir do que foi lido tem de ser a
				// MESMA que o exemplo gerou. Diferente significa que algum
				// dos oito campos que a compoem se perdeu na ida e volta.
				if remontada != chaveGerada {
					t.Errorf("a chave remontada diverge da gerada:\n  gerada:    %s\n  remontada: %s",
						chaveGerada, remontada)
				}
				n.InfNFe.ID = remontada
			}

			if err := nfe.ValidarChaveAcesso(n); err != nil {
				t.Errorf("a chave do exemplo nao fecha: %v", err)
			}
		})
	}
}

// TestExemplos_NaoLevamDadoReal trava a restrição de segurança do projeto.
//
// O catálogo usa o CNPJ que o próprio ACBr usa nos exemplos dele e marca o
// ambiente como homologação. Um exemplo que chegue em produção, ou com
// documento de cliente, é vazamento — e a revisão de código não pega
// sozinha um CNPJ que parece plausível.
func TestExemplos_NaoLevamDadoReal(t *testing.T) {
	const cnpjDeExemplo = "99999999000191"

	for _, ex := range catalogo() {
		t.Run(ex.ID, func(t *testing.T) {
			conteudo, _, err := ex.Montar()
			if err != nil {
				t.Fatalf("montar: %v", err)
			}
			n, err := lerNota(conteudo)
			if err != nil {
				t.Fatalf("ler: %v", err)
			}
			if n.Emit.CNPJCPF != cnpjDeExemplo {
				t.Errorf("o emitente dos exemplos e sempre o CNPJ ficticio %s, veio %q",
					cnpjDeExemplo, n.Emit.CNPJCPF)
			}
			if n.Ide.TpAmb.String() != "2" {
				t.Errorf("exemplo tem de sair em HOMOLOGACAO (tpAmb=2), veio %q",
					n.Ide.TpAmb.String())
			}
		})
	}
}

// TestExemplos_PassamNaValidacaoDeRegras trava a qualidade do catálogo.
//
// Exemplo que levanta rejeição ensina a coisa errada: quem está conhecendo
// a biblioteca cola o exemplo em `/api/validar`, vê quatro rejeições e
// conclui que o porte é que está torto. E as rejeições aparecem sozinhas
// com o tempo — a 228 recusa nota com mais de trinta dias de atraso, então
// um exemplo com data cravada apodrece.
//
// A validação passa pelo HANDLER, e não pela função do package, de
// propósito: é o handler que monta a chave quando a entrada é `.ini`, e é
// o caminho que o usuário da demo percorre.
func TestExemplos_PassamNaValidacaoDeRegras(t *testing.T) {
	for _, ex := range catalogo() {
		t.Run(ex.ID, func(t *testing.T) {
			conteudo, _, err := ex.Montar()
			if err != nil {
				t.Fatalf("montar: %v", err)
			}

			w := httptest.NewRecorder()
			handleValidar(w, httptest.NewRequest(http.MethodPost, "/api/validar",
				strings.NewReader(conteudo)))

			var resp ValidarResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if !resp.ChaveValida {
				t.Errorf("a chave nao fecha: %s", resp.ErroChave)
			}
			for _, r := range resp.Rejeicoes {
				t.Errorf("rejeicao %d: %s", r.Codigo, r.Mensagem)
			}
			if !resp.Valida {
				t.Error("o exemplo precisa passar limpo na validacao")
			}
		})
	}
}

// TestHandleINIParaXML_ConverteOExemplo trava o caminho que a maioria dos
// ERP usa: o sistema gera o `.ini`, o componente produz o XML.
func TestHandleINIParaXML_ConverteOExemplo(t *testing.T) {
	ini, _, err := exemploINI(notaDeVenda())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/ini-para-xml",
		strings.NewReader(ini))
	w := httptest.NewRecorder()
	handleINIParaXML(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}

	var resp struct {
		Chave string `json:"chave"`
		XML   string `json:"xml"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.XML, "<infNFe") {
		t.Errorf("a resposta deveria trazer o XML da nota: %.200s", resp.XML)
	}
	if resp.Chave == "" {
		t.Error("a resposta deveria trazer a chave")
	}
}

// TestHandleValidar_SeparaChaveDeRegrasDeNegocio trava a distinção entre
// as duas validações, que respondem perguntas diferentes.
//
// A chave confere a coerência interna do documento; as regras conferem o
// conteúdo contra a legislação. Juntá-las numa resposta só faria o
// consumidor não saber o que corrigir.
func TestHandleValidar_SeparaChaveDeRegrasDeNegocio(t *testing.T) {
	xml, _, err := exemploXML(notaDeVenda())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/validar", strings.NewReader(xml))
	w := httptest.NewRecorder()
	handleValidar(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}

	var resp ValidarResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.ChaveValida {
		t.Errorf("a chave do exemplo deveria fechar: %s", resp.ErroChave)
	}
	// As rejeicoes podem existir -- o exemplo nao pretende passar em todas
	// as regras --, mas o campo precisa vir como lista, nunca nulo: o
	// frontend itera sobre ele.
	if resp.Rejeicoes == nil {
		t.Error("rejeicoes deveria ser lista vazia, nunca nula")
	}
}

// TestSepararDocumentos_QuebraPeloFimDaTagRaiz trava o critério de divisão
// do lote.
//
// XML de nota costuma vir numa linha só; dividir por linha em branco
// devolveria um documento gigante e o lote inteiro falharia como um.
func TestSepararDocumentos_QuebraPeloFimDaTagRaiz(t *testing.T) {
	const doc = `<nfeProc versao="4.00"><NFe/></nfeProc>`

	// Dois documentos na MESMA linha, sem separador nenhum.
	partes := separarDocumentos(doc + doc)
	if len(partes) != 2 {
		t.Fatalf("documentos = %d, esperado 2: %q", len(partes), partes)
	}
	for i, p := range partes {
		if p != doc {
			t.Errorf("documento %d = %q", i, p)
		}
	}

	// A raiz e DESCOBERTA: uma nota ainda nao transmitida vem como <NFe>
	// solta, nao como <nfeProc>. Cravar uma das duas faz o lote da outra
	// virar um documento unico e ilegivel -- foi assim que o defeito
	// apareceu, rodando a demo.
	const solta = `<NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe/></NFe>`
	if partes := separarDocumentos(solta + solta); len(partes) != 2 {
		t.Errorf("<NFe> solta tambem e separada, veio %d: %q",
			len(partes), partes)
	}

	// Com prolog na frente, a raiz continua sendo achada.
	const comProlog = `<?xml version="1.0" encoding="UTF-8"?>` + solta
	if partes := separarDocumentos(comProlog); len(partes) != 1 {
		t.Errorf("o prolog nao e a raiz, veio %d", len(partes))
	} else if !strings.HasPrefix(partes[0], "<?xml") {
		t.Errorf("o documento sai inteiro, com o prolog: %q", partes[0])
	}

	// Conteudo sem a tag raiz continua sendo UM documento -- e o caso do
	// .ini, que nao tem raiz nenhuma.
	if partes := separarDocumentos("[ide]\nnatOp=X\n"); len(partes) != 1 {
		t.Errorf("conteudo sem raiz e um documento so, veio %d", len(partes))
	}

	// Vazio nao vira documento.
	if partes := separarDocumentos("   \n  "); len(partes) != 0 {
		t.Errorf("conteudo vazio nao gera documento, veio %d", len(partes))
	}
}

// TestHandleStatus_ListaAsRotasQueExistem trava o catálogo de rotas da
// resposta de status contra o que o main registra de fato.
//
// A lista é documentação servida pela API; uma rota removida do main e
// esquecida ali manda o consumidor para um 404.
func TestHandleStatus_ListaAsRotasQueExistem(t *testing.T) {
	w := httptest.NewRecorder()
	handleStatus(w, httptest.NewRequest(http.MethodGet, "/api/status", nil))

	var resp StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Status != "ok" || resp.Package != "nfe" {
		t.Errorf("status = %+v", resp)
	}
	if len(resp.Modelos) != 2 {
		t.Errorf("a NF-e cobre os modelos 55 e 65: %+v", resp.Modelos)
	}

	// Toda rota anunciada precisa estar no main.go.
	registradas := rotasDoMain(t)
	for _, rota := range resp.Rotas {
		if !registradas[rota] {
			t.Errorf("a rota %q e anunciada no /api/status e NAO esta registrada no main.go", rota)
		}
	}
	// E toda rota registrada precisa ser anunciada.
	anunciadas := map[string]bool{}
	for _, r := range resp.Rotas {
		anunciadas[r] = true
	}
	for rota := range registradas {
		if !anunciadas[rota] {
			t.Errorf("a rota %q esta no main.go e NAO e anunciada no /api/status", rota)
		}
	}
}

// rotasDoMain extrai do main.go as rotas registradas no ServeMux.
//
// Ler o fonte em vez de manter uma segunda lista no teste e deliberado:
// uma lista a mais seria mais um lugar para esquecer de atualizar, que e
// exatamente o defeito que este teste existe para pegar.
func rotasDoMain(t *testing.T) map[string]bool {
	t.Helper()

	fonte, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ler main.go: %v", err)
	}

	padrao := regexp.MustCompile(`mux\.HandleFunc\("([^"]+)"`)
	rotas := map[string]bool{}
	for _, m := range padrao.FindAllStringSubmatch(string(fonte), -1) {
		rotas[m[1]] = true
	}
	if len(rotas) == 0 {
		t.Fatal("nenhuma rota encontrada no main.go -- o padrao de busca quebrou")
	}
	return rotas
}
