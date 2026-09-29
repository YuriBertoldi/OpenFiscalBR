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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/sped"
)

// GerarRequest represents the JSON body for POST /api/gerar.
type GerarRequest struct {
	// Registro 0000
	CodVer    int    `json:"cod_ver"`
	CodFin    int    `json:"cod_fin"`
	DtIni     string `json:"dt_ini"` // yyyy-MM-dd
	DtFin     string `json:"dt_fin"` // yyyy-MM-dd
	Nome      string `json:"nome"`
	CNPJ      string `json:"cnpj"`
	CPF       string `json:"cpf"`
	UF        string `json:"uf"`
	IE        string `json:"ie"`
	CodMun    int    `json:"cod_mun"`
	IM        string `json:"im"`
	Suframa   string `json:"suframa"`
	IndPerfil int    `json:"ind_perfil"` // 0=A, 1=B, 2=C
	IndAtiv   int    `json:"ind_ativ"`   // 0=Industrial, 1=Outros

	// Registro 0005 (optional)
	Fantasia string `json:"fantasia"`
	CEP      string `json:"cep"`
	Endereco string `json:"endereco"`
	Num      string `json:"num"`
	Compl    string `json:"compl"`
	Bairro   string `json:"bairro"`
	Fone     string `json:"fone"`
	Email    string `json:"email"`
}

// GerarResponse is the JSON response from POST /api/gerar.
type GerarResponse struct {
	Sucesso  bool   `json:"sucesso"`
	Mensagem string `json:"mensagem"`
	Arquivo  string `json:"arquivo,omitempty"`
	Conteudo string `json:"conteudo,omitempty"`
}

// StatusResponse is the JSON response from GET /api/status.
type StatusResponse struct {
	Status string `json:"status"`
	Versao string `json:"versao"`
	Modulo string `json:"modulo"`
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	resp := StatusResponse{
		Status: "ok",
		Versao: "1.0.0",
		Modulo: "SPED Fiscal (EFD-ICMS/IPI)",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleGerar(w http.ResponseWriter, r *http.Request) {
	var req GerarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return
	}

	dtIni, err := time.Parse("2006-01-02", req.DtIni)
	if err != nil {
		writeError(w, http.StatusBadRequest, "dt_ini invalida (use yyyy-MM-dd): "+err.Error())
		return
	}
	dtFin, err := time.Parse("2006-01-02", req.DtFin)
	if err != nil {
		writeError(w, http.StatusBadRequest, "dt_fin invalida (use yyyy-MM-dd): "+err.Error())
		return
	}

	// Create SPEDFiscal instance
	fiscal := sped.NewSPEDFiscal()

	// Output to temp dir
	tmpDir := os.TempDir()
	fiscal.Path = tmpDir
	fiscal.Arquivo = fmt.Sprintf("SPED_%s_%s.txt", req.CNPJ, dtIni.Format("012006"))
	fiscal.DtIni = dtIni
	fiscal.DtFin = dtFin

	// Populate Registro0000
	reg0000 := fiscal.Bloco0.Registro0000
	reg0000.CodVer = sped.VersaoLeiauteFiscal(req.CodVer)
	reg0000.CodFin = sped.CodFin(req.CodFin)
	reg0000.DtIni = dtIni
	reg0000.DtFin = dtFin
	reg0000.Nome = req.Nome
	reg0000.CNPJ = req.CNPJ
	reg0000.CPF = req.CPF
	reg0000.UF = req.UF
	reg0000.IE = req.IE
	reg0000.CodMun = req.CodMun
	reg0000.IM = req.IM
	reg0000.Suframa = req.Suframa
	reg0000.IndPerfil = sped.IndPerfil(req.IndPerfil)
	reg0000.IndAtiv = sped.IndAtiv(req.IndAtiv)

	// Populate Registro0005 if provided
	if req.Fantasia != "" || req.CEP != "" || req.Endereco != "" {
		fiscal.Bloco0.Registro0001.Registro0005 = &sped.Registro0005{
			Fantasia: req.Fantasia,
			CEP:      req.CEP,
			Endereco: req.Endereco,
			Num:      req.Num,
			Compl:    req.Compl,
			Bairro:   req.Bairro,
			Fone:     req.Fone,
			Email:    req.Email,
		}
	}

	// Mark block 0 as having data
	fiscal.Bloco0.Registro0001.IndDad = 0

	// Generate file
	if err := fiscal.SaveFileTXT(); err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao gerar arquivo: "+err.Error())
		return
	}

	// Read generated file content
	outPath := filepath.Join(tmpDir, fiscal.Arquivo)
	content, err := os.ReadFile(outPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Erro ao ler arquivo gerado: "+err.Error())
		return
	}

	// Clean up temp file
	defer os.Remove(outPath)

	resp := GerarResponse{
		Sucesso:  true,
		Mensagem: "Arquivo SPED Fiscal gerado com sucesso",
		Arquivo:  fiscal.Arquivo,
		Conteudo: string(content),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(GerarResponse{
		Sucesso:  false,
		Mensagem: msg,
	})
}
