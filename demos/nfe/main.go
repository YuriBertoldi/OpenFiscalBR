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

	// Leitura, validacao e conversao entre os dois formatos de entrada.
	// Nenhuma destas rotas precisa de certificado nem de rede.
	mux.HandleFunc("POST /api/ler", handleLer)
	mux.HandleFunc("POST /api/ler-lote", handleLerLote)
	mux.HandleFunc("POST /api/validar", handleValidar)
	mux.HandleFunc("POST /api/ler-evento", handleLerEvento)
	mux.HandleFunc("POST /api/ler-distribuicao", handleLerDistribuicao)
	mux.HandleFunc("POST /api/ini-para-xml", handleINIParaXML)
	mux.HandleFunc("POST /api/xml-para-ini", handleXMLParaINI)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Emissao -- geracao, assinatura e transmissao a SEFAZ.
	// Certificado A1 via CERT_PATH/CERT_PASS; UF, MODELO e AMBIENTE por env.
	mux.HandleFunc("POST /api/gerar", handleGerar)
	mux.HandleFunc("POST /api/assinar", handleAssinar)
	mux.HandleFunc("POST /api/transmitir", handleTransmitir)
	mux.HandleFunc("GET /api/status-sefaz", handleStatusSefaz)
	mux.HandleFunc("GET /api/consultar/{chave}", handleConsultar)
	mux.HandleFunc("POST /api/cancelar", handleCancelar)
	mux.HandleFunc("POST /api/carta-correcao", handleCartaCorrecao)
	mux.HandleFunc("POST /api/inutilizar", handleInutilizar)
	mux.HandleFunc("POST /api/distribuicao", handleDistribuicao)

	// Exemplos -- documentos sinteticos, com dados ficticios, para quem
	// quiser experimentar sem ter um XML a mao.
	mux.HandleFunc("GET /api/exemplos", handleListarExemplos)
	mux.HandleFunc("GET /api/exemplos/{id}", handleExemplo)

	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "frontend"
	}
	mux.Handle("/", http.FileServer(http.Dir(staticDir)))

	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenFiscalBR NFe Demo rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
