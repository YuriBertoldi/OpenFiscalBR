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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/dfe"
	"github.com/openfiscalbr/openfiscalbr/packages/nfe"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Rotas de EMISSAO da demo. A configuracao vem do ambiente:
//
//	CERT_PATH  caminho do .pfx
//	CERT_PASS  senha do .pfx
//	UF         UF do emitente (default SP)
//	MODELO     55 (NF-e) ou 65 (NFC-e); default 55
//	AMBIENTE   "producao" ou "homologacao"; default homologacao
//	CSC/IDCSC  Codigo de Seguranca do Contribuinte, so para NFC-e
//
// NENHUMA rota desta demo transmite em producao por acidente: o default do
// AMBIENTE e homologacao, e e preciso dizer "producao" explicitamente.

var (
	certOnce sync.Once
	certA1   *dfe.Certificado
	certErr  error
)

func certificado() (*dfe.Certificado, error) {
	certOnce.Do(func() {
		caminho := os.Getenv("CERT_PATH")
		if caminho == "" {
			certErr = errors.New(
				"certificado nao configurado: defina CERT_PATH (arquivo .pfx) e CERT_PASS")
			return
		}
		certA1, certErr = dfe.CarregarPFXArquivo(caminho, os.Getenv("CERT_PASS"))
	})
	return certA1, certErr
}

// configuracoes monta a configuracao do componente a partir do ambiente.
//
// E usada tambem pelas rotas que NAO transmitem -- a leitura do .ini e a
// validacao de regras precisam da UF, do ambiente e do modelo para decidir
// defaults e rejeicoes.
func configuracoes() nfe.Configuracoes {
	uf := os.Getenv("UF")
	if uf == "" {
		uf = "SP"
	}

	cfg := nfe.Configuracoes{
		VersaoDF: nfe.Ve400,
		ModeloDF: nfe.MoNFe,
		Ambiente: pcn.TaHomologacao,
		UF:       uf,
		CodigoUF: pcn.CodigoUF(uf),
		TpEmis:   int(pcn.TeNormal),
		CSC:      os.Getenv("CSC"),
		IDCSC:    os.Getenv("IDCSC"),
	}

	if os.Getenv("MODELO") == "65" {
		cfg.ModeloDF = nfe.MoNFCe
	}
	if strings.EqualFold(os.Getenv("AMBIENTE"), "producao") {
		cfg.Ambiente = pcn.TaProducao
	}
	return cfg
}

// componente devolve o componente ja configurado, com certificado quando
// houver. Sem certificado o componente ainda serve as rotas que nao
// transmitem.
func componente() *nfe.Componente {
	c := nfe.NovoComponente(configuracoes())
	if cert, err := certificado(); err == nil {
		c.Config.Certificado = cert
	}
	return c
}

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// GerarResponse devolve o XML gerado (e assinado, quando for o caso).
type GerarResponse struct {
	Chave    string `json:"chave"`
	XML      string `json:"xml"`
	Assinado bool   `json:"assinado"`
	QRCode   string `json:"qrCode,omitempty"`
}

// CancelarRequest e o corpo de /api/cancelar.
type CancelarRequest struct {
	Chave         string `json:"chave"`
	Protocolo     string `json:"protocolo"`
	Justificativa string `json:"justificativa"`
	IDLote        int64  `json:"idLote"`
}

// CartaCorrecaoRequest e o corpo de /api/carta-correcao.
type CartaCorrecaoRequest struct {
	Chave     string `json:"chave"`
	Correcao  string `json:"correcao"`
	Sequencia int    `json:"nSeqEvento"`
	IDLote    int64  `json:"idLote"`
}

// InutilizarRequest e o corpo de /api/inutilizar.
type InutilizarRequest struct {
	CNPJ        string `json:"cnpj"`
	Ano         int    `json:"ano"`
	Modelo      int    `json:"modelo"`
	Serie       int    `json:"serie"`
	NumeroFinal int    `json:"nNFFin"`
	NumeroIni   int    `json:"nNFIni"`
	Justificava string `json:"justificativa"`
}

// DistribuicaoRequest e o corpo de /api/distribuicao.
type DistribuicaoRequest struct {
	CUFAutor int    `json:"cUFAutor"`
	CNPJCPF  string `json:"cnpjCpf"`
	UltNSU   string `json:"ultNSU"`
	NSU      string `json:"nsu"`
	Chave    string `json:"chave"`
	// Varrer percorre o lote inteiro em vez de fazer uma chamada so.
	Varrer bool `json:"varrer"`
	Limite int  `json:"limite"`
}

// ---------------------------------------------------------------------------
// Geracao e assinatura -- nao precisam de rede
// ---------------------------------------------------------------------------

// handleGerar gera o XML da NF-e a partir de um XML ou de um .ini.
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
	xml, err := nfe.GerarXML(n)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, GerarResponse{Chave: n.InfNFe.ID, XML: xml})
}

// handleAssinar gera, assina e -- na NFC-e -- calcula o QR-Code.
//
// O QR-Code vem DEPOIS da assinatura de proposito: a versao 3 do QR-Code
// assina o hash com o certificado, e as versoes 1 e 2 usam o digest do
// documento. Calcular antes produziria um codigo que nao corresponde ao
// XML transmitido.
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
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	nota := &nfe.NotaFiscal{NFe: n}
	assinado, err := componente().PrepararEnvio(nota)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp := GerarResponse{Chave: n.InfNFe.ID, XML: assinado, Assinado: true}

	if n.Ide.Modelo == 65 {
		qr, err := nfe.GerarQRCode(n, configuracoes(), cert.AssinarSHA1Base64)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity,
				"XML assinado, mas o QR-Code falhou: "+err.Error())
			return
		}
		resp.QRCode = qr
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Transmissao
// ---------------------------------------------------------------------------

// handleTransmitir envia a nota para a SEFAZ e devolve o resultado.
func handleTransmitir(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	resultado, err := componente().Enviar(ctx,
		[]*nfe.NotaFiscal{{NFe: n}}, nfe.OpcoesEnvio{Sincrono: true})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resultado)
}

// handleStatusSefaz consulta o status do servico na SEFAZ.
func handleStatusSefaz(w http.ResponseWriter, r *http.Request) {
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}
	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	ret, err := componente().StatusServico(ctx, time.Now())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleConsultar consulta uma nota pela chave de acesso.
func handleConsultar(w http.ResponseWriter, r *http.Request) {
	chave := r.PathValue("chave")
	if chave == "" {
		writeError(w, http.StatusBadRequest, "informe a chave de acesso")
		return
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}
	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	ret, err := componente().Consultar(ctx, chave)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleCancelar cancela uma nota autorizada.
//
// O protocolo e obrigatorio: o cancelamento referencia a AUTORIZACAO, nao
// a nota. Sem ele a SEFAZ rejeita, e exigir aqui evita a viagem.
func handleCancelar(w http.ResponseWriter, r *http.Request) {
	var req CancelarRequest
	if !decodificar(w, r, &req) {
		return
	}
	if req.Chave == "" || req.Protocolo == "" || req.Justificativa == "" {
		writeError(w, http.StatusBadRequest,
			"chave, protocolo e justificativa sao obrigatorios")
		return
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}
	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	ret, err := componente().Cancelamento(ctx, req.Chave, req.Protocolo,
		req.Justificativa, req.IDLote, time.Now())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleCartaCorrecao envia uma carta de correcao.
func handleCartaCorrecao(w http.ResponseWriter, r *http.Request) {
	var req CartaCorrecaoRequest
	if !decodificar(w, r, &req) {
		return
	}
	if req.Chave == "" || req.Correcao == "" {
		writeError(w, http.StatusBadRequest, "chave e correcao sao obrigatorias")
		return
	}
	if req.Sequencia <= 0 {
		req.Sequencia = 1
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}
	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	ret, err := componente().EnviarCartaCorrecao(ctx, req.Chave, req.Correcao,
		req.Sequencia, req.IDLote, time.Now())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleInutilizar inutiliza uma faixa de numeracao.
func handleInutilizar(w http.ResponseWriter, r *http.Request) {
	var req InutilizarRequest
	if !decodificar(w, r, &req) {
		return
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}

	pedido := nfe.PedidoInutilizacao{
		CNPJ:          req.CNPJ,
		Ano:           req.Ano,
		Modelo:        req.Modelo,
		Serie:         req.Serie,
		NumeroInicial: req.NumeroIni,
		NumeroFinal:   req.NumeroFinal,
		Justificativa: req.Justificava,
	}
	// A validacao roda ANTES da transmissao: o pedido tem regras proprias
	// (faixa, ano, justificativa minima) que nao dependem da SEFAZ.
	if err := nfe.ValidarPedidoInutilizacao(pedido); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	ret, err := componente().Inutilizar(ctx, pedido)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ret)
}

// handleDistribuicao consulta os documentos emitidos CONTRA o contribuinte.
//
// Com `varrer`, percorre o lote inteiro em vez de fazer uma chamada so --
// a RFB devolve no maximo cinquenta documentos por chamada, entao a
// chamada unica quase nunca traz tudo.
func handleDistribuicao(w http.ResponseWriter, r *http.Request) {
	var req DistribuicaoRequest
	if !decodificar(w, r, &req) {
		return
	}
	if _, err := certificado(); err != nil {
		writeError(w, http.StatusPreconditionFailed, err.Error())
		return
	}

	pedido := nfe.PedidoDistribuicao{
		CUFAutor: req.CUFAutor,
		CNPJCPF:  req.CNPJCPF,
		UltNSU:   req.UltNSU,
		NSU:      req.NSU,
		Chave:    req.Chave,
	}

	ctx, cancelar := contextoDaRequisicao(r)
	defer cancelar()

	c := componente()
	if !req.Varrer {
		ret, err := c.DistribuicaoDFe(ctx, pedido)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, resumirDistribuicao(ret))
		return
	}

	docs, ret, err := c.VarrerDistribuicaoDFe(ctx, pedido, req.Limite)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	// Na varredura o retorno traz so o ULTIMO lote; os documentos vem da
	// acumulacao. Substituir a lista evita devolver cinquenta de um lote
	// de quinhentos.
	resp := resumirDistribuicao(ret)
	resp.Documentos = resp.Documentos[:0]
	for _, d := range docs {
		item := DocZipResponse{
			NSU: d.NSU, Schema: d.Schema.String(),
			Chave: nfe.ChaveDoDocZip(d), Tipo: tipoDoDocZip(d), XML: d.XML,
		}
		if d.ErroLeitura != nil {
			item.Erro = d.ErroLeitura.Error()
		}
		resp.Documentos = append(resp.Documentos, item)
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// decodificar le o corpo como JSON.
func decodificar(w http.ResponseWriter, r *http.Request, destino any) bool {
	defer r.Body.Close()
	corpo, err := io.ReadAll(io.LimitReader(r.Body, limiteCorpo))
	if err != nil {
		writeError(w, http.StatusBadRequest, "falha ao ler o corpo: "+err.Error())
		return false
	}
	if err := json.Unmarshal(corpo, destino); err != nil {
		writeError(w, http.StatusBadRequest, "JSON invalido: "+err.Error())
		return false
	}
	return true
}

// contextoDaRequisicao devolve o contexto da requisicao com o timeout da
// demo.
//
// Pendura no contexto do REQUEST, e nao em context.Background(): se o
// navegador desistir, a chamada a SEFAZ e cancelada junto em vez de ficar
// segurando conexao aberta ate o timeout.
func contextoDaRequisicao(r *http.Request) (context.Context, context.CancelFunc) {
	segundos := 90
	if v := os.Getenv("TIMEOUT_SEFAZ"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			segundos = n
		}
	}
	return context.WithTimeout(r.Context(), time.Duration(segundos)*time.Second)
}
