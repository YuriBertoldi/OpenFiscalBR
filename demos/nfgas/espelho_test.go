// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-02.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// As demos nfgas e nfag sao gemeas: o MECANISMO (rotas, handlers, JS, HTML,
// CSS) tem que ser identico a menos dos nomes do modelo, e so o CONTEUDO
// FISCAL (exemplos.go, que monta as structs de cada leiaute) diverge.
//
// Esse grau de simetria e um ativo -- corrigir um bug de UI em um lado e
// esquecer o outro e o jeito mais facil de perde-lo. Este teste existe num
// lado so; roda sobre os dois.

// arquivosEspelhados sao os de mecanismo, que precisam bater byte a byte
// depois da normalizacao de nomes.
var arquivosEspelhados = []string{
	"handlers.go",
	"handlers_emissao.go",
	"handlers_exemplos.go",
	"main.go",
	"frontend/app.js",
	"frontend/index.html",
	"frontend/style.css",
}

// linhasToleradas sao divergencias legitimas de conteudo, nao de mecanismo:
// cada demo descreve o proprio documento fiscal. Sao casadas por prefixo
// depois de TrimSpace.
var linhasToleradas = []string{
	"<p>Nota Fiscal de Fornecimento de", // subtitulo: gas canalizado x agua canalizada
}

// espacosInternos colapsa sequencias de espaco que nao estao na indentacao.
// O gofmt alinha campos de struct em coluna, e a largura da coluna depende do
// comprimento do nome ("NFGas" tem 5 letras, "NFAg" tem 4) -- alinhamento
// diferente ali e consequencia do nome, nao divergencia de mecanismo.
var espacosInternos = regexp.MustCompile(`(\S) {2,}`)

// normalizar troca os nomes proprios de cada modelo por marcadores neutros,
// para que a comparacao enxergue so a estrutura.
func normalizar(s string) string {
	s = espacosInternos.ReplaceAllString(s, "$1 ")
	r := strings.NewReplacer(
		// ordem importa: do mais especifico para o mais generico
		"modelo 76", "§modelo", "modelo 75", "§modelo",
		"ModeloNFGas", "§Modelo", "ModeloNFAg", "§Modelo",
		"NFGAS", "§UPPER", "NFAG", "§UPPER",
		"NFGas", "§Pascal", "NFAg", "§Pascal",
		"nfgas", "§lower", "nfag", "§lower",
	)
	return r.Replace(s)
}

func tolerada(linha string) bool {
	t := strings.TrimSpace(linha)
	for _, p := range linhasToleradas {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func TestDemoEspelhaNFAg(t *testing.T) {
	for _, nome := range arquivosEspelhados {
		t.Run(nome, func(t *testing.T) {
			aqui, err := os.ReadFile(filepath.FromSlash(nome))
			if err != nil {
				t.Fatalf("ler %s: %v", nome, err)
			}
			la, err := os.ReadFile(filepath.Join("..", "nfag", filepath.FromSlash(nome)))
			if err != nil {
				t.Fatalf("ler o espelho de %s: %v", nome, err)
			}

			linhasA := strings.Split(normalizar(string(aqui)), "\n")
			linhasB := strings.Split(normalizar(string(la)), "\n")

			if len(linhasA) != len(linhasB) {
				t.Errorf("%s tem %d linhas aqui e %d na demo nfag -- as gemeas divergiram",
					nome, len(linhasA), len(linhasB))
			}

			n := min(len(linhasA), len(linhasB))
			for i := 0; i < n; i++ {
				a := strings.TrimRight(linhasA[i], "\r")
				b := strings.TrimRight(linhasB[i], "\r")
				if a == b || tolerada(a) {
					continue
				}
				t.Fatalf("%s divergiu na linha %d:\n  nfgas: %s\n   nfag: %s",
					nome, i+1, a, b)
			}
		})
	}
}

// TestCatalogosTemOsMesmosObrigatorios confere que a lista de cenarios
// obrigatorios usada pelos dois exemplos_test.go nao saiu do lugar. A lista
// em si vive em exemplos_test.go de cada demo, e e identica por construcao;
// aqui garantimos que a daqui cobre o que foi acordado.
func TestListaDeObrigatoriosEstavel(t *testing.T) {
	esperados := []string{
		"transmissao", "autorizada", "cancelamento", "cancelamento-proc",
		"completo", "erro-leitura", "erro-regras", "multi-itens",
	}
	if len(idsObrigatorios) != len(esperados) {
		t.Fatalf("idsObrigatorios tem %d entradas, esperadas %d", len(idsObrigatorios), len(esperados))
	}
	for _, id := range esperados {
		if _, ok := idsObrigatorios[id]; !ok {
			t.Errorf("%q sumiu de idsObrigatorios", id)
		}
	}
}
