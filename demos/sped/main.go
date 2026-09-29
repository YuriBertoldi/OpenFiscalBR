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

	// API routes
	mux.HandleFunc("POST /api/gerar", handleGerar)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Serve frontend static files
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenFiscalBR SPED Demo rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
