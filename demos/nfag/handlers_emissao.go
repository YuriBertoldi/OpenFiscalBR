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
	"os"
	"strings"
	"sync"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/nfag"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Rotas de EMISSAO da demo. O certificado A1 vem do ambiente:
//
//	CERT_PATH  caminho do .pfx (ou CERT_PEM/CERT_KEY para par PEM)
//	CERT_PASS  senha do .pfx
//	UF         UF autorizadora (default SP)
//	AMBIENTE   "producao" ou "homologacao" (default homologacao)

var (
	certOnce sync.Once
	certA1   *dfe.Certificado
	certErr  error
)

func certificado() (*dfe.Certificado, error) {
	certOnce.Do(func() {
		caminho := os.Getenv("CERT_PATH")
		if caminho == "" {
			certErr = errors.New("certificado nao configurado: defina CERT_PATH (arquivo .pfx) e CERT_PASS")
			return
		}
		certA1, certErr = dfe.CarregarPFXArquivo(caminho, os.Getenv("CERT_PASS"))
	})
	return certA1, certErr
}

func componente() *nfag.Componente {
	c := nfag.NovoComponente()
	c.Configuracoes.UF = os.Getenv("UF")
	if c.Configuracoes.UF == "" {
		c.Configuracoes.UF = "SP"
	}
	if strings.EqualFold(os.Getenv("AMBIENTE"), "producao") {
		c.Configuracoes.Ambiente = pcn.TaProducao
	}
	if cert, err := certificado(); err == nil {
		c.Configuracoes.Certificado = cert
	}
	return c
}

// GerarResponse devolve o XML gerado (e assinado, quando for o caso).
type GerarResponse struct {
	Chave     string `json:"chave"`
	XML       string `json:"xml"`
	Assinado  bool   `json:"assinado"`
	QRCodeURL string `json:"qrCodeUrl,omitempty"`
}

// CancelarRequest e o corpo de /api/cancelar.
type CancelarRequest struct {
	Chave         string `json:"chave"`
	Protocolo     string `json:"protocolo"`
	Justificativa string `json:"justificativa"`
}

// ConsultarRequest e o corpo de /api/consultar-sefaz.
type ConsultarRequest struct {
	Chave string `json:"chave"`
}

// lerNota aceita XML da NFAg ou .ini no formato do ACBr e devolve a nota.
func lerNota(conteudo string) (*nfag.NFAg, error) {
	if strings.HasPrefix(strings.TrimSpace(conteudo), "<") {
		return nfag.LerXMLString(conteudo)
	}
	return nfag.LerINI(conteudo, nfag.Configuracoes{VersaoDF: nfag.Ve100})
}

// handleGerar gera o XML da NFAg a partir de um XML ou .ini.
func handleGerar(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	xml, err := nfag.GerarXML(n)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	resp := GerarResponse{Chave: n.ChaveAcesso(), XML: xml}
	if url, err := componente().GerarQRCode(n); err == nil {
		resp.QRCodeURL = url
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleAssinar gera (se preciso) e assina o XML com o certificado do
// ambiente, devolvendo o XML pronto para transmissao.
func handleAssinar(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	cert, err := certificado()
	if err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}

	xml := strings.TrimSpace(conteudo)
	var chave string
	if !strings.HasPrefix(xml, "<") || !strings.Contains(xml, "<NFAg") {
		n, err := lerNota(conteudo)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if xml, err = nfag.GerarXML(n); err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		chave = n.ChaveAcesso()
	}

	assinado, err := nfag.Assinar(cert, xml)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, GerarResponse{Chave: chave, XML: assinado, Assinado: true})
}

// handleTransmitir envia a nota a SEFAZ (recepcao sincrona de 1 NFAg).
func handleTransmitir(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	c := componente()
	if c.Configuracoes.Certificado == nil {
		writeError(w, http.StatusPreconditionFailed, "certificado nao configurado: defina CERT_PATH e CERT_PASS")
		return
	}
	nota := &nfag.NotaFiscal{NFAg: n}
	ret, err := c.Enviar(r.Context(), nota)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"retorno":     ret,
		"autorizada":  ret.Autorizada(),
		"xmlAssinado": nota.XMLAssinado,
	})
}

// handleStatusSefaz consulta a disponibilidade do web service.
func handleStatusSefaz(w http.ResponseWriter, r *http.Request) {
	c := componente()
	if c.Configuracoes.Certificado == nil {
		writeError(w, http.StatusPreconditionFailed, "certificado nao configurado: defina CERT_PATH e CERT_PASS")
		return
	}
	ret, err := c.StatusServico(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleConsultarSefaz consulta a situacao de uma chave na SEFAZ.
func handleConsultarSefaz(w http.ResponseWriter, r *http.Request) {
	var req ConsultarRequest
	if !lerJSON(w, r, &req) {
		return
	}
	c := componente()
	if c.Configuracoes.Certificado == nil {
		writeError(w, http.StatusPreconditionFailed, "certificado nao configurado: defina CERT_PATH e CERT_PASS")
		return
	}
	ret, err := c.Consultar(r.Context(), req.Chave)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleCancelar envia o evento de cancelamento.
func handleCancelar(w http.ResponseWriter, r *http.Request) {
	var req CancelarRequest
	if !lerJSON(w, r, &req) {
		return
	}
	c := componente()
	if c.Configuracoes.Certificado == nil {
		writeError(w, http.StatusPreconditionFailed, "certificado nao configurado: defina CERT_PATH e CERT_PASS")
		return
	}
	ret, err := c.Cancelamento(r.Context(), req.Chave, req.Protocolo, req.Justificativa)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

func lerJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()
	corpo, err := io.ReadAll(io.LimitReader(r.Body, limiteCorpo))
	if err != nil || len(corpo) == 0 {
		writeError(w, http.StatusBadRequest, "corpo vazio ou ilegivel")
		return false
	}
	if err := json.Unmarshal(corpo, v); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return false
	}
	return true
}
