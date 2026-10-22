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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/nfe"
)

const limiteCorpo = 32 << 20 // 32 MiB por requisicao

// ---------------------------------------------------------------------------
// DTOs
// ---------------------------------------------------------------------------

// LerRequest e o corpo das rotas que recebem um documento como texto.
type LerRequest struct {
	// Conteudo e o XML ou o .ini colado como texto.
	Conteudo string `json:"conteudo"`
}

// NotaResumo e o resumo de uma nota, para listagem.
type NotaResumo struct {
	Chave        string  `json:"chave"`
	Modelo       int     `json:"modelo"`
	Serie        int     `json:"serie"`
	NNF          int     `json:"nNF"`
	DhEmi        string  `json:"dhEmi"`
	Emitente     string  `json:"emitente"`
	CNPJEmitente string  `json:"cnpjEmitente"`
	Destinatario string  `json:"destinatario"`
	UF           string  `json:"uf"`
	VNF          float64 `json:"vNF"`
	Itens        int     `json:"itens"`
	Protocolo    string  `json:"protocolo,omitempty"`
}

// LerResponse devolve a nota completa mais o resumo.
type LerResponse struct {
	Resumo NotaResumo `json:"resumo"`
	NFe    *nfe.NFe   `json:"nfe"`
}

// LoteResponse devolve o resultado da importacao em lote.
type LoteResponse struct {
	Total   int          `json:"total"`
	Lidas   int          `json:"lidas"`
	Falhas  []string     `json:"falhas,omitempty"`
	Resumos []NotaResumo `json:"resumos"`
}

// RejeicaoDTO e uma rejeicao da validacao de regras de negocio.
type RejeicaoDTO struct {
	Codigo   int    `json:"codigo"`
	Mensagem string `json:"mensagem"`
	Item     int    `json:"item,omitempty"`
}

// ValidarResponse devolve o resultado da validacao.
//
// A chave e validada SEPARADAMENTE das regras de negocio porque sao duas
// coisas diferentes: a chave confere a coerencia interna do documento
// (UF, data, CNPJ, modelo, serie, numero e digito), as regras conferem o
// conteudo contra a legislacao.
type ValidarResponse struct {
	Chave       string        `json:"chave"`
	ChaveValida bool          `json:"chaveValida"`
	ErroChave   string        `json:"erroChave,omitempty"`
	Rejeicoes   []RejeicaoDTO `json:"rejeicoes"`
	Valida      bool          `json:"valida"`
}

// EventoResponse devolve o retorno de evento lido.
type EventoResponse struct {
	Versao   string               `json:"versao"`
	CStat    int                  `json:"cStat"`
	XMotivo  string               `json:"xMotivo"`
	IDLote   int64                `json:"idLote,omitempty"`
	Eventos  []EventoItemResponse `json:"eventos"`
	TodosOKs bool                 `json:"todosRegistrados"`
}

// EventoItemResponse e um evento dentro do retorno do lote.
type EventoItemResponse struct {
	Chave      string `json:"chave"`
	TipoEvento string `json:"tipoEvento"`
	Descricao  string `json:"descricao"`
	Sequencia  int    `json:"nSeqEvento"`
	CStat      int    `json:"cStat"`
	XMotivo    string `json:"xMotivo"`
	NProt      string `json:"nProt,omitempty"`
	Registrado bool   `json:"registrado"`
}

// DistribuicaoResponse resume o retorno da Distribuicao DF-e.
type DistribuicaoResponse struct {
	CStat      int              `json:"cStat"`
	XMotivo    string           `json:"xMotivo"`
	UltNSU     string           `json:"ultNSU"`
	MaxNSU     string           `json:"maxNSU"`
	Atendida   bool             `json:"atendida"`
	Concluida  bool             `json:"varreduraConcluida"`
	Documentos []DocZipResponse `json:"documentos"`
}

// DocZipResponse e um item do lote da distribuicao.
type DocZipResponse struct {
	NSU    string `json:"nsu"`
	Schema string `json:"schema"`
	Chave  string `json:"chave,omitempty"`
	Tipo   string `json:"tipo"`
	Erro   string `json:"erro,omitempty"`
	XML    string `json:"xml,omitempty"`
}

// StatusResponse informa que a API esta de pe.
type StatusResponse struct {
	Status  string   `json:"status"`
	Package string   `json:"package"`
	Modelos []int    `json:"modelos"`
	Leiaute string   `json:"leiaute"`
	Rotas   []string `json:"rotas"`
}

// ---------------------------------------------------------------------------
// Handlers de leitura
// ---------------------------------------------------------------------------

func handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, StatusResponse{
		Status:  "ok",
		Package: "nfe",
		Modelos: []int{55, 65},
		Leiaute: nfe.Ve400.String(),
		Rotas: []string{
			"POST /api/ler", "POST /api/ler-lote", "POST /api/validar",
			"POST /api/ler-evento", "POST /api/ler-distribuicao",
			"POST /api/ini-para-xml", "POST /api/xml-para-ini",
			"GET /api/status",
			"POST /api/gerar", "POST /api/assinar", "POST /api/transmitir",
			"GET /api/status-sefaz", "GET /api/consultar/{chave}",
			"POST /api/cancelar", "POST /api/carta-correcao",
			"POST /api/inutilizar", "POST /api/distribuicao",
			"GET /api/exemplos", "GET /api/exemplos/{id}",
		},
	})
}

// handleLer le uma NF-e em XML ou .ini e devolve a nota estruturada.
func handleLer(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, LerResponse{Resumo: resumir(n), NFe: n})
}

// handleLerLote le varios documentos de uma vez.
//
// O criterio de separacao e o fim da tag raiz, e nao uma linha em branco:
// XML pode vir numa linha so. Um documento ilegivel entra em `falhas` e os
// demais seguem -- e a mesma decisao do leitor de lote do package.
func handleLerLote(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}

	partes := separarDocumentos(conteudo)
	resp := LoteResponse{Total: len(partes)}

	for i, p := range partes {
		n, err := lerNota(p)
		if err != nil {
			resp.Falhas = append(resp.Falhas,
				"documento "+itoa(i+1)+": "+err.Error())
			continue
		}
		resp.Lidas++
		resp.Resumos = append(resp.Resumos, resumir(n))
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleValidar confere a chave de acesso e as regras de negocio.
func handleValidar(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := lerNota(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// O `.ini` NAO carrega a chave: o `infNFe/ID` e uma das oito chaves que
	// o ACBr grava e nao le, porque ela e DERIVADA dos campos. Monta-la
	// aqui faz a validacao responder a mesma coisa para os dois formatos de
	// entrada -- sem isso, todo .ini levantaria a rejeicao 253 por falta de
	// uma chave que nunca esteve la.
	if n.InfNFe.ID == "" {
		if chave, err := nfe.MontarChaveAcesso(n); err == nil {
			n.InfNFe.ID = chave
		}
	}

	resp := ValidarResponse{Chave: n.InfNFe.ID, Rejeicoes: []RejeicaoDTO{}}
	if err := nfe.ValidarChaveAcesso(n); err != nil {
		resp.ErroChave = err.Error()
	} else {
		resp.ChaveValida = true
	}

	for _, e := range nfe.ValidarRegrasNegocio(n, configuracoes(), time.Now()) {
		resp.Rejeicoes = append(resp.Rejeicoes, RejeicaoDTO{
			Codigo: e.Codigo, Mensagem: e.Mensagem, Item: e.Item,
		})
	}
	resp.Valida = resp.ChaveValida && len(resp.Rejeicoes) == 0

	writeJSON(w, http.StatusOK, resp)
}

// handleLerEvento le o retorno de um lote de eventos (retEnvEvento).
func handleLerEvento(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	ret, err := nfe.LerRetEventoNFe(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	resp := EventoResponse{
		Versao: ret.Versao, CStat: ret.CStat, XMotivo: ret.XMotivo,
		IDLote: ret.IDLote, Eventos: []EventoItemResponse{}, TodosOKs: true,
	}
	for _, e := range ret.RetEvento {
		registrado := nfe.CStatProcessado(e.CStat)
		if !registrado {
			resp.TodosOKs = false
		}
		resp.Eventos = append(resp.Eventos, EventoItemResponse{
			Chave:      e.ChNFe,
			TipoEvento: e.TpEvento.String(),
			Descricao:  nfe.DescricaoDoTipoEvento(e.TpEvento),
			Sequencia:  e.NSeqEvento,
			CStat:      e.CStat,
			XMotivo:    e.XMotivo,
			NProt:      e.NProt,
			Registrado: registrado,
		})
	}
	if len(resp.Eventos) == 0 {
		resp.TodosOKs = false
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleLerDistribuicao le um retDistDFeInt, descomprimindo cada docZip.
func handleLerDistribuicao(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	ret, err := nfe.LerRetDistDFeInt(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resumirDistribuicao(ret))
}

// ---------------------------------------------------------------------------
// Conversao entre os dois formatos de entrada
// ---------------------------------------------------------------------------

// handleINIParaXML converte o .ini do ACBr em XML da NF-e.
//
// E o caminho que a maioria dos ERP usa: o sistema gera o .ini, o
// componente produz o XML. A demo expoe os dois lados para que de para
// conferir a ida e a volta sem certificado.
func handleINIParaXML(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := nfe.LerINI(conteudo, configuracoes())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	xml, err := nfe.GerarXML(n)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chave": n.InfNFe.ID,
		"xml":   xml,
	})
}

// handleXMLParaINI converte o XML da NF-e no .ini do ACBr.
func handleXMLParaINI(w http.ResponseWriter, r *http.Request) {
	conteudo, ok := lerConteudo(w, r)
	if !ok {
		return
	}
	n, err := nfe.LerXMLString(conteudo)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	texto, err := nfe.GravarINI(n)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"chave": n.InfNFe.ID,
		"ini":   texto,
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// lerNota aceita XML da NF-e ou .ini no formato do ACBr.
//
// O criterio e o primeiro caractere nao branco: `<` e XML, qualquer outra
// coisa e .ini. E o mesmo criterio do LerIniArquivoOuString do original.
func lerNota(conteudo string) (*nfe.NFe, error) {
	if strings.HasPrefix(strings.TrimSpace(conteudo), "<") {
		return nfe.LerXMLString(conteudo)
	}
	return nfe.LerINI(conteudo, configuracoes())
}

// separarDocumentos quebra um texto com varios XML em documentos.
//
// Divide pelo FIM da tag raiz, e nao por linha em branco: um XML de nota
// costuma vir numa linha so, e dividir por linha devolveria um documento
// gigante que falharia como um.
//
// A tag raiz e DESCOBERTA, nao cravada. Uma nota autorizada vem como
// `<nfeProc>` e uma ainda nao transmitida vem como `<NFe>` solta -- cravar
// uma das duas faz o lote da outra virar um documento unico e ilegivel.
func separarDocumentos(conteudo string) []string {
	raiz := tagRaiz(conteudo)
	if raiz == "" {
		if s := strings.TrimSpace(conteudo); s != "" {
			return []string{s}
		}
		return nil
	}

	fim := "</" + raiz + ">"
	var docs []string
	resto := conteudo
	for {
		i := strings.Index(resto, fim)
		if i < 0 {
			break
		}
		corte := i + len(fim)
		if doc := strings.TrimSpace(resto[:corte]); doc != "" {
			docs = append(docs, doc)
		}
		resto = resto[corte:]
	}
	if s := strings.TrimSpace(resto); s != "" {
		docs = append(docs, s)
	}
	return docs
}

// tagRaiz devolve o nome do primeiro elemento do texto, pulando o prolog e
// os comentarios. Devolve vazio quando nao ha XML -- o caso do `.ini`.
func tagRaiz(conteudo string) string {
	resto := conteudo
	for {
		i := strings.Index(resto, "<")
		if i < 0 {
			return ""
		}
		resto = resto[i+1:]

		switch {
		case strings.HasPrefix(resto, "?"):
			// Prolog: <?xml ... ?>
			if j := strings.Index(resto, "?>"); j >= 0 {
				resto = resto[j+2:]
				continue
			}
			return ""
		case strings.HasPrefix(resto, "!"):
			// Comentario ou DOCTYPE.
			if j := strings.Index(resto, ">"); j >= 0 {
				resto = resto[j+1:]
				continue
			}
			return ""
		}

		fim := strings.IndexAny(resto, " \t\r\n>/")
		if fim <= 0 {
			return ""
		}
		return resto[:fim]
	}
}

// resumir monta o resumo de uma nota para listagem.
func resumir(n *nfe.NFe) NotaResumo {
	r := NotaResumo{
		Chave:        n.InfNFe.ID,
		Modelo:       n.Ide.Modelo,
		Serie:        n.Ide.Serie,
		NNF:          n.Ide.NNF,
		Emitente:     n.Emit.XNome,
		CNPJEmitente: n.Emit.CNPJCPF,
		Destinatario: n.Dest.XNome,
		UF:           n.Emit.EnderEmit.UF,
		VNF:          n.Total.ICMSTot.VNF,
		Itens:        len(n.Det),
		Protocolo:    n.ProcNFe.NProt,
	}
	if !n.Ide.DEmi.IsZero() {
		r.DhEmi = n.Ide.DEmi.Format("2006-01-02T15:04:05-07:00")
	}
	return r
}

// resumirDistribuicao monta a resposta da Distribuicao DF-e.
func resumirDistribuicao(ret *nfe.RetDistDFeInt) DistribuicaoResponse {
	resp := DistribuicaoResponse{
		CStat: ret.CStat, XMotivo: ret.XMotivo,
		UltNSU: ret.UltNSU, MaxNSU: ret.MaxNSU,
		Atendida:   ret.Atendida(),
		Concluida:  ret.VarreduraConcluida(),
		Documentos: []DocZipResponse{},
	}
	for _, d := range ret.DocZip {
		item := DocZipResponse{
			NSU:    d.NSU,
			Schema: d.Schema.String(),
			Chave:  nfe.ChaveDoDocZip(d),
			Tipo:   tipoDoDocZip(d),
			XML:    d.XML,
		}
		if d.ErroLeitura != nil {
			item.Erro = d.ErroLeitura.Error()
		}
		resp.Documentos = append(resp.Documentos, item)
	}
	return resp
}

// tipoDoDocZip diz, em uma palavra, o que o item e.
//
// O schema do atributo nao serve para isso: quem decide o tipo e a RAIZ do
// documento descomprimido, e e ela que preenche um dos tres grupos.
func tipoDoDocZip(d nfe.DocZip) string {
	switch {
	case d.ResDFe.ChDFe != "" && d.ResDFe.NProt != "":
		return "nota processada"
	case d.ResDFe.ChDFe != "":
		return "resumo de nota"
	case d.ResEvento.ChDFe != "":
		return "resumo de evento"
	case d.ProcEvento.ChDFe != "":
		return "evento processado"
	}
	return "desconhecido"
}

// lerConteudo extrai o conteudo da requisicao. Aceita JSON
// {"conteudo": "..."} e tambem o corpo cru, para facilitar o curl.
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

// itoa evita importar strconv so para numerar mensagem de falha.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d [20]byte
	i := len(d)
	for n > 0 {
		i--
		d[i] = byte('0' + n%10)
		n /= 10
	}
	return string(d[i:])
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"erro": msg})
}
