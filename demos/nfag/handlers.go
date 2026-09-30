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
	"errors"
	"io"
	"net/http"

	"github.com/openfiscalbr/openfiscalbr/packages/nfag"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

const limiteCorpo = 32 << 20 // 32 MiB por requisicao

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// LerRequest e o corpo de /api/ler, /api/ler-lote, /api/ler-evento,
// /api/ler-consulta e /api/ler-ini.
type LerRequest struct {
	// Conteudo e o XML (ou .ini) colado como texto.
	Conteudo string `json:"conteudo"`
}

// NotaResumo e o resumo de uma nota lida, para a listagem do lote.
type NotaResumo struct {
	Chave    string  `json:"chave"`
	NNF      int     `json:"nNF"`
	Serie    int     `json:"serie"`
	DhEmi    string  `json:"dhEmi"`
	Emitente string  `json:"emitente"`
	CNPJ     string  `json:"cnpj"`
	VNF      float64 `json:"vNF"`
	CStat    int     `json:"cStat"`
	Situacao string  `json:"situacao"`
	Itens    int     `json:"itens"`
}

// LerResponse devolve a nota completa mais o resumo.
type LerResponse struct {
	Resumo NotaResumo   `json:"resumo"`
	NFAg  *nfag.NFAg `json:"nfag"`
}

// LoteResponse devolve o resultado da importacao em lote.
type LoteResponse struct {
	Total   int          `json:"total"`
	Lidas   int          `json:"lidas"`
	Falhas  []string     `json:"falhas,omitempty"`
	Resumos []NotaResumo `json:"resumos"`
}

// ValidarResponse devolve o resultado da validacao de um XML.
type ValidarResponse struct {
	Chave          string   `json:"chave"`
	ChaveValida    bool     `json:"chaveValida"`
	ErroChave      string   `json:"erroChave,omitempty"`
	ConcatConfere  bool     `json:"concatConfere"`
	Rejeicoes      []string `json:"rejeicoes"`
	CNPJEmitenteOK bool     `json:"cnpjEmitenteOk"`
	ErroCNPJ       string   `json:"erroCnpj,omitempty"`
}

// EventoResponse devolve o evento lido.
type EventoResponse struct {
	Chave         string `json:"chave"`
	TipoEvento    string `json:"tipoEvento"`
	Descricao     string `json:"descricao"`
	CStat         int    `json:"cStat"`
	XMotivo       string `json:"xMotivo"`
	NProt         string `json:"nProt"`
	Registrado    bool   `json:"registrado"`
	Cancelamento  bool   `json:"cancelamento"`
	Justificativa string `json:"justificativa,omitempty"`
}

// StatusResponse informa que a API esta de pe.
type StatusResponse struct {
	Status  string   `json:"status"`
	Package string   `json:"package"`
	Modelo  int      `json:"modelo"`
	Leiaute string   `json:"leiaute"`
	Rotas   []string `json:"rotas"`
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, StatusResponse{
		Status:  "ok",
		Package: "nfag",
		Modelo:  nfag.ModeloNFAg,
		Leiaute: nfag.Ve100.String(),
		Rotas: []string{
			"POST /api/ler", "POST /api/ler-lote", "POST /api/validar",
			"POST /api/ler-evento", "POST /api/ler-consulta", "POST /api/ler-ini",
			"GET /api/status",
			"POST /api/gerar", "POST /api/assinar", "POST /api/transmitir",
			"GET /api/status-sefaz", "POST /api/consultar-sefaz", "POST /api/cancelar",
		},
	})
}

func handleLer(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := nfag.LerXMLString(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, LerResponse{Resumo: resumo(n), NFAg: n})
}

func handleLerLote(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}

	notas, err := nfag.LerLoteString(conteudo)
	resp := LoteResponse{Lidas: len(notas)}

	for _, nota := range notas {
		resp.Resumos = append(resp.Resumos, resumo(nota.NFAg))
	}
	resp.Total = len(notas)

	if err != nil {
		var lote *nfag.ErrosLote
		if errors.As(err, &lote) {
			// Nota torta nao derruba o lote: as lidas voltam, as falhas
			// viram texto.
			resp.Total += len(lote.Erros)
			for _, f := range lote.Erros {
				resp.Falhas = append(resp.Falhas, f.Error())
			}
		} else {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func handleValidar(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := nfag.LerXMLString(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	cfg := nfag.Configuracoes{
		VersaoDF: nfag.Ve100,
		Ambiente: n.Ide.TpAmb,
		UF:       pcn.SiglaUF(n.Ide.CUF),
		CodigoUF: n.Ide.CUF,
	}

	resp := ValidarResponse{
		Chave:         n.ChaveAcesso(),
		ConcatConfere: nfag.ValidarConcatChave(n),
		Rejeicoes:     []string{},
	}
	if err := nfag.ValidarChaveAcesso(n); err != nil {
		resp.ErroChave = err.Error()
	} else {
		resp.ChaveValida = true
	}
	if err := pcn.ValidarCNPJouCPF(n.Emit.CNPJ); err != nil {
		resp.ErroCNPJ = err.Error()
	} else {
		resp.CNPJEmitenteOK = true
	}
	for _, rej := range nfag.ValidarRegrasNegocio(n, cfg) {
		resp.Rejeicoes = append(resp.Rejeicoes, rej.Error())
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleLerEvento(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	ev, err := nfag.LerEventoString(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	tipo := ev.RetInfEvento.TpEvento
	if ev.TemEvento && ev.RetInfEvento.CStat == 0 {
		tipo = ev.Evento.InfEvento.TpEvento
	}

	writeJSON(w, http.StatusOK, EventoResponse{
		Chave:         ev.ChaveAcesso(),
		TipoEvento:    tipo.String(),
		Descricao:     tipo.Descricao(),
		CStat:         ev.RetInfEvento.CStat,
		XMotivo:       ev.RetInfEvento.XMotivo,
		NProt:         ev.RetInfEvento.NProt,
		Registrado:    ev.RetInfEvento.Registrado(),
		Cancelamento:  ev.Cancelamento(),
		Justificativa: ev.Justificativa(),
	})
}

func handleLerConsulta(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	ret, err := nfag.LerRetConsSitString(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

func handleLerINI(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := nfag.LerINI(conteudo, nfag.Configuracoes{VersaoDF: nfag.Ve100})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, LerResponse{Resumo: resumo(n), NFAg: n})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func resumo(n *nfag.NFAg) NotaResumo {
	r := NotaResumo{
		Chave:    n.ChaveAcesso(),
		NNF:      n.Ide.NNF,
		Serie:    n.Ide.Serie,
		Emitente: n.Emit.XNome,
		CNPJ:     n.Emit.CNPJ,
		VNF:      n.Total.VNF,
		CStat:    n.ProcNFAg.CStat,
		Itens:    len(n.Det),
	}
	if !n.Ide.DhEmi.IsZero() {
		r.DhEmi = n.Ide.DhEmi.Format("2006-01-02T15:04:05-07:00")
	}
	nota := nfag.NotaFiscal{NFAg: n}
	r.Situacao = nota.Situacao()
	return r
}

// lerConteudo extrai o conteudo da requisicao. Aceita JSON {"conteudo": "..."}
// e tambem o corpo cru (text/plain ou application/xml), para facilitar curl.
func lerConteudo(w http.ResponseWriter, r *http.Request) (string, bool) {
	defer r.Body.Close()
	corpo, err := io.ReadAll(io.LimitReader(r.Body, limiteCorpo))
	if err != nil {
		writeError(w, http.StatusBadRequest, "falha ao ler o corpo: "+err.Error())
		return "", false
	}
	if len(corpo) == 0 {
		writeError(w, http.StatusBadRequest, "corpo vazio")
		return "", false
	}

	var req LerRequest
	if json.Unmarshal(corpo, &req) == nil && req.Conteudo != "" {
		return req.Conteudo, true
	}
	return string(corpo), true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"erro": msg})
}
