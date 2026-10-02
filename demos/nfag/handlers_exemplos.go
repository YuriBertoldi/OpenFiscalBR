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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/openfiscalbr/openfiscalbr/packages/nfag"
)

// Rotas do catalogo de exemplos. Sao GET com os parametros na query string,
// e nao POST com corpo JSON, para que o download seja uma ancora href no
// frontend -- sem Blob, sem JavaScript extra.

// prologXML prefixa o arquivo baixado. O XML que trafega para a SEFAZ nao
// leva prolog; o que vira arquivo em disco, sim.
const prologXML = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"

// ExemploResumo e a entrada do catalogo, sem o XML.
type ExemploResumo struct {
	ID           string `json:"id"`
	Nome         string `json:"nome"`
	Descricao    string `json:"descricao"`
	Tipo         string `json:"tipo"`
	AcaoSugerida string `json:"acaoSugerida"`
	Arquivo      string `json:"arquivo"`
	Observacao   string `json:"observacao,omitempty"`
	AceitaItens  bool   `json:"aceitaItens"`
}

// ExemplosResponse e o corpo de GET /api/exemplos.
type ExemplosResponse struct {
	Total    int             `json:"total"`
	Exemplos []ExemploResumo `json:"exemplos"`
}

// ExemploResponse e o corpo de GET /api/exemplos/{id}.
type ExemploResponse struct {
	ExemploResumo
	Chave string `json:"chave,omitempty"`
	XML   string `json:"xml"`
}

func resumoExemplo(ex Exemplo) ExemploResumo {
	return ExemploResumo{
		ID:           ex.ID,
		Nome:         ex.Nome,
		Descricao:    ex.Descricao,
		Tipo:         ex.Tipo,
		AcaoSugerida: ex.AcaoSugerida,
		Arquivo:      ex.Arquivo,
		Observacao:   ex.Observacao,
		AceitaItens:  ex.AceitaItens,
	}
}

// parsearParametros le os parametros opcionais da query string. Query vazia
// devolve o zero value, que e o caminho do "clicar e gerar": XML ficticio
// completo, identico a cada chamada.
func parsearParametros(r *http.Request) (ParametrosExemplo, error) {
	q := r.URL.Query()
	p := ParametrosExemplo{
		CNPJEmit: strings.TrimSpace(q.Get("cnpj")),
		UF:       strings.TrimSpace(q.Get("uf")),
		Variar:   q.Get("variar") == "1" || q.Get("variar") == "true",
	}

	inteiro := func(nome string, destino *int) error {
		v := strings.TrimSpace(q.Get(nome))
		if v == "" {
			return nil
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s: %q nao e um numero inteiro", nome, v)
		}
		*destino = i
		return nil
	}
	for _, c := range []struct {
		nome    string
		destino *int
	}{
		{"serie", &p.Serie}, {"nnf", &p.NNF},
		{"tpamb", &p.TpAmb}, {"itens", &p.Itens},
	} {
		if err := inteiro(c.nome, c.destino); err != nil {
			return ParametrosExemplo{}, err
		}
	}
	// Dominio conferido aqui para que parametro torto vire 400, e nao 500
	// la na frente, quando o builder ja estiver montando documento.
	if err := validarParametros(p); err != nil {
		return ParametrosExemplo{}, err
	}
	return p, nil
}

// gerarExemplo resolve o id, le os parametros e produz o XML. Devolve o
// status HTTP adequado a cada falha.
func gerarExemplo(r *http.Request) (Exemplo, string, int, error) {
	id := r.PathValue("id")
	ex, ok := acharExemplo(id)
	if !ok {
		return Exemplo{}, "", http.StatusNotFound,
			fmt.Errorf("exemplo desconhecido: %q", id)
	}
	p, err := parsearParametros(r)
	if err != nil {
		return Exemplo{}, "", http.StatusBadRequest, err
	}
	xml, err := ex.Gerar(p)
	if err != nil {
		return Exemplo{}, "", http.StatusInternalServerError, err
	}
	return ex, xml, http.StatusOK, nil
}

func handleListarExemplos(w http.ResponseWriter, _ *http.Request) {
	catalogo := catalogoExemplos()
	resp := ExemplosResponse{Total: len(catalogo), Exemplos: make([]ExemploResumo, 0, len(catalogo))}
	for _, ex := range catalogo {
		resp.Exemplos = append(resp.Exemplos, resumoExemplo(ex))
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleExemplo(w http.ResponseWriter, r *http.Request) {
	ex, xml, status, err := gerarExemplo(r)
	if err != nil {
		writeError(w, status, err.Error())
		return
	}
	resp := ExemploResponse{ExemploResumo: resumoExemplo(ex), XML: xml}
	// A chave so existe nos cenarios de documento legivel -- o de erro de
	// leitura nao tem, e o de evento guarda a chave do documento referido.
	if n, err := nfag.LerXMLString(xml); err == nil {
		resp.Chave = n.ChaveAcesso()
	} else if ev, err := nfag.LerEventoString(xml); err == nil {
		resp.Chave = ev.ChaveAcesso()
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleBaixarExemplo(w http.ResponseWriter, r *http.Request) {
	ex, xml, status, err := gerarExemplo(r)
	if err != nil {
		writeError(w, status, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+ex.Arquivo+`"`)
	w.WriteHeader(http.StatusOK)
	// O writer do package nao emite o prolog (ele devolve o documento para
	// ser transmitido, nao gravado). Aqui o XML vira ARQUIVO em disco, e
	// validador de XSD de linha de comando costuma exigi-lo.
	_, _ = io.WriteString(w, prologXML)
	_, _ = io.WriteString(w, xml)
}
