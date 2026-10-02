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

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	// Leitura e validacao
	mux.HandleFunc("POST /api/ler", handleLer)
	mux.HandleFunc("POST /api/ler-lote", handleLerLote)
	mux.HandleFunc("POST /api/validar", handleValidar)
	mux.HandleFunc("POST /api/ler-evento", handleLerEvento)
	mux.HandleFunc("POST /api/ler-consulta", handleLerConsulta)
	mux.HandleFunc("POST /api/ler-ini", handleLerINI)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Emissao -- geracao, assinatura e transmissao a SEFAZ.
	// Certificado A1 via CERT_PATH/CERT_PASS; UF e AMBIENTE via env.
	mux.HandleFunc("POST /api/gerar", handleGerar)
	mux.HandleFunc("POST /api/assinar", handleAssinar)
	mux.HandleFunc("POST /api/transmitir", handleTransmitir)
	mux.HandleFunc("GET /api/status-sefaz", handleStatusSefaz)
	mux.HandleFunc("POST /api/consultar-sefaz", handleConsultarSefaz)
	mux.HandleFunc("POST /api/cancelar", handleCancelar)

	// Exemplos -- XMLs sinteticos para testar a importacao, com dados
	// ficticios completos. Sem parametro na query, o XML sai sempre igual.
	mux.HandleFunc("GET /api/exemplos", handleListarExemplos)
	mux.HandleFunc("GET /api/exemplos/{id}", handleExemplo)
	mux.HandleFunc("GET /api/exemplos/{id}/download", handleBaixarExemplo)

	// Serve frontend static files
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenFiscalBR NFAg Demo rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
