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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/nfgas"
	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// idsObrigatorios sao os cenarios que as duas demos (nfgas e nfag) tem que
// expor com o mesmo id, tipo e aba sugerida. O extra multi-cfop so existe na
// NFGas: a NFAg nao tem CFOP no leiaute.
var idsObrigatorios = map[string]struct{ Tipo, Acao string }{
	"transmissao":       {"documento", "gerar"},
	"autorizada":        {"documento", "ler"},
	"cancelamento":      {"evento", "ler-evento"},
	"cancelamento-proc": {"evento", "ler-evento"},
	"completo":          {"documento", "ler"},
	"erro-leitura":      {"documento", "ler"},
	"erro-regras":       {"documento", "validar"},
	"multi-itens":       {"documento", "ler"},
}

// acoesValidas e o conjunto de data-acao das abas do frontend/index.html.
var acoesValidas = map[string]bool{
	"ler": true, "ler-lote": true, "validar": true, "ler-evento": true,
	"ler-consulta": true, "ler-ini": true, "gerar": true, "assinar": true,
	"transmitir": true,
}

func exemploPorID(t *testing.T, id string) Exemplo {
	t.Helper()
	ex, ok := acharExemplo(id)
	if !ok {
		t.Fatalf("exemplo %q nao existe no catalogo", id)
	}
	return ex
}

// ---------------------------------------------------------------------------
// Catalogo
// ---------------------------------------------------------------------------

func TestExemplos_MetadadosCoerentes(t *testing.T) {
	vistos := map[string]bool{}
	for _, ex := range catalogoExemplos() {
		if ex.ID == "" || vistos[ex.ID] {
			t.Fatalf("id vazio ou repetido: %q", ex.ID)
		}
		vistos[ex.ID] = true

		if ex.Nome == "" || ex.Descricao == "" {
			t.Errorf("%s: nome e descricao sao o que o usuario le antes de gerar", ex.ID)
		}
		if !strings.HasSuffix(ex.Arquivo, ".xml") {
			t.Errorf("%s: arquivo %q deveria terminar em .xml", ex.ID, ex.Arquivo)
		}
		if ex.Tipo != "documento" && ex.Tipo != "evento" {
			t.Errorf("%s: tipo %q fora de {documento, evento}", ex.ID, ex.Tipo)
		}
		if !acoesValidas[ex.AcaoSugerida] {
			t.Errorf("%s: acaoSugerida %q nao corresponde a nenhuma aba do index.html",
				ex.ID, ex.AcaoSugerida)
		}
		if ex.Gerar == nil {
			t.Errorf("%s: sem builder", ex.ID)
		}
	}
}

func TestCatalogoTemOsCenariosObrigatorios(t *testing.T) {
	presentes := map[string]Exemplo{}
	for _, ex := range catalogoExemplos() {
		presentes[ex.ID] = ex
	}
	for id, esperado := range idsObrigatorios {
		ex, ok := presentes[id]
		if !ok {
			t.Errorf("cenario obrigatorio ausente: %q", id)
			continue
		}
		if ex.Tipo != esperado.Tipo || ex.AcaoSugerida != esperado.Acao {
			t.Errorf("%s: tipo/acao = %q/%q, esperado %q/%q",
				id, ex.Tipo, ex.AcaoSugerida, esperado.Tipo, esperado.Acao)
		}
	}
	// Extra desta demo: a NFAg nao tem equivalente porque o leiaute dela nao
	// tem CFOP (ausencia de fato, nao lacuna do porte).
	if _, ok := presentes["multi-cfop"]; !ok {
		t.Error("cenario multi-cfop ausente na demo da NFGas")
	}
}

// ---------------------------------------------------------------------------
// Geracao
// ---------------------------------------------------------------------------

func TestExemplos_GeramXMLLegivel(t *testing.T) {
	for _, ex := range catalogoExemplos() {
		t.Run(ex.ID, func(t *testing.T) {
			xml, err := ex.Gerar(ParametrosExemplo{})
			if err != nil {
				t.Fatalf("gerar: %v", err)
			}
			if strings.HasPrefix(ex.ID, "erro") {
				return // tem erro de proposito; conferidos em teste proprio
			}
			if ex.Tipo == "evento" {
				if _, err := nfgas.LerEventoString(xml); err != nil {
					t.Fatalf("reler evento: %v\n%s", err, xml)
				}
				return
			}
			if _, err := nfgas.LerXMLString(xml); err != nil {
				t.Fatalf("reler documento: %v\n%s", err, xml)
			}
		})
	}
}

func TestExemplos_Deterministicos(t *testing.T) {
	// Sem parametros, o mesmo cenario tem que sair identico byte a byte --
	// e isso que sustenta "clicar e gerar" como caminho testavel. Pega cNF
	// nao fixado (o writer sorteia quando zero) e time.Now() esquecido.
	for _, ex := range catalogoExemplos() {
		t.Run(ex.ID, func(t *testing.T) {
			a, err := ex.Gerar(ParametrosExemplo{})
			if err != nil {
				t.Fatal(err)
			}
			b, err := ex.Gerar(ParametrosExemplo{})
			if err != nil {
				t.Fatal(err)
			}
			if a != b {
				t.Fatalf("exemplo %s nao e deterministico", ex.ID)
			}
		})
	}
}

func TestExemplos_VariarMudaAChave(t *testing.T) {
	ex := exemploPorID(t, "transmissao")
	a, err := ex.Gerar(ParametrosExemplo{Variar: true})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ex.Gerar(ParametrosExemplo{Variar: true})
	if err != nil {
		t.Fatal(err)
	}
	na, err := nfgas.LerXMLString(a)
	if err != nil {
		t.Fatal(err)
	}
	nb, err := nfgas.LerXMLString(b)
	if err != nil {
		t.Fatal(err)
	}
	if na.ChaveAcesso() == nb.ChaveAcesso() {
		t.Fatalf("variar=1 deveria mudar a chave, as duas vieram %s", na.ChaveAcesso())
	}
}

func TestExemplos_ParametrosSobrescrevem(t *testing.T) {
	ex := exemploPorID(t, "transmissao")
	p := ParametrosExemplo{
		CNPJEmit: "99888777000166", UF: "MG", Serie: 7, NNF: 4321, TpAmb: 1,
	}
	xml, err := ex.Gerar(p)
	if err != nil {
		t.Fatal(err)
	}
	n, err := nfgas.LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if n.Emit.CNPJ != "99888777000166" {
		t.Errorf("CNPJ = %q", n.Emit.CNPJ)
	}
	if n.Ide.CUF != 31 || n.Emit.EnderEmit.UF != "MG" {
		t.Errorf("UF = %d/%q, esperado 31/MG", n.Ide.CUF, n.Emit.EnderEmit.UF)
	}
	if n.Ide.Serie != 7 || n.Ide.NNF != 4321 {
		t.Errorf("serie/nNF = %d/%d", n.Ide.Serie, n.Ide.NNF)
	}
	if n.Ide.TpAmb != pcn.TaProducao {
		t.Errorf("tpAmb = %v", n.Ide.TpAmb)
	}

	// A chave tem que ter sido RECALCULADA com os novos campos -- e por isso
	// que os parametros sao aplicados antes do writer, nao depois.
	if !nfgas.ValidarConcatChave(n) {
		t.Errorf("chave %q nao confere com os campos sobrescritos", n.ChaveAcesso())
	}
	if err := nfgas.ValidarChaveAcesso(n); err != nil {
		t.Errorf("chave invalida apos sobrescrita: %v", err)
	}

	// Trocar a UF nao pode introduzir rejeicao: cUF, UF e cMun andam juntos.
	cfg := nfgas.Configuracoes{
		VersaoDF: nfgas.Ve100, Ambiente: n.Ide.TpAmb,
		UF: pcn.SiglaUF(n.Ide.CUF), CodigoUF: n.Ide.CUF,
	}
	if rej := nfgas.ValidarRegrasNegocio(n, cfg); len(rej) > 0 {
		t.Errorf("sobrescrever a UF nao deveria gerar rejeicao, veio %v", rej)
	}
}

func TestParsearParametros_Invalido(t *testing.T) {
	for _, q := range []string{"?nnf=abc", "?serie=x", "?itens=1.5"} {
		r := httptest.NewRequest(http.MethodGet, "/api/exemplos/transmissao"+q, nil)
		if _, err := parsearParametros(r); err == nil {
			t.Errorf("query %q deveria ser recusada", q)
		}
	}
	// UF e tpAmb sao validados na aplicacao, onde esta o dominio deles.
	ex := exemploPorID(t, "transmissao")
	if _, err := ex.Gerar(ParametrosExemplo{UF: "ZZ"}); err == nil {
		t.Error("UF inexistente deveria ser recusada")
	}
	if _, err := ex.Gerar(ParametrosExemplo{TpAmb: 3}); err == nil {
		t.Error("tpAmb 3 deveria ser recusado")
	}
	if _, err := ex.Gerar(ParametrosExemplo{CNPJEmit: "123"}); err == nil {
		t.Error("CNPJ curto deveria ser recusado")
	}
}

// ---------------------------------------------------------------------------
// Cenarios especificos
// ---------------------------------------------------------------------------

func TestExemploErroLeitura_FalhaNaLeitura(t *testing.T) {
	xml, err := exemploPorID(t, "erro-leitura").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = nfgas.LerXMLString(xml)
	if !errors.Is(err, nfgas.ErrAtributoVersaoAusente) {
		t.Fatalf("esperado ErrAtributoVersaoAusente, veio %v", err)
	}
}

func TestExemploErroRegras_RejeicoesEsperadas(t *testing.T) {
	xml, err := exemploPorID(t, "erro-regras").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := nfgas.LerXMLString(xml)
	if err != nil {
		t.Fatalf("o cenario de regras tem que LER bem: %v", err)
	}

	// cfg montada como em handleValidar (handlers.go): vem do proprio
	// documento. Por isso a regra 252 (ambiente) e inalcancavel por esta
	// rota -- ela compararia TpAmb consigo mesmo. Nao esta prometida no
	// cenario justamente por isso.
	cfg := nfgas.Configuracoes{
		VersaoDF: nfgas.Ve100, Ambiente: n.Ide.TpAmb,
		UF: pcn.SiglaUF(n.Ide.CUF), CodigoUF: n.Ide.CUF,
	}
	codigos := map[int]bool{}
	for _, rej := range nfgas.ValidarRegrasNegocio(n, cfg) {
		codigos[rej.Codigo] = true
	}
	for _, c := range []int{226, 227, 247} {
		if !codigos[c] {
			t.Errorf("regra %d deveria ter sido acusada; veio %v", c, codigos)
		}
	}
	if codigos[252] {
		t.Error("252 nao deveria aparecer: handleValidar deriva o ambiente do documento")
	}

	if nfgas.ValidarConcatChave(n) {
		t.Error("a concatenacao da chave deveria divergir (nNF trocado apos a geracao)")
	}
	// O digito verificador fica integro de proposito: chave valida com
	// concatenacao divergente e o caso que mais escapa de um importador.
	if err := nfgas.ValidarChaveAcesso(n); err != nil {
		t.Errorf("o dV da chave deveria continuar valido: %v", err)
	}
	if err := pcn.ValidarCNPJouCPF(n.Emit.CNPJ); err == nil {
		t.Error("o CNPJ do emitente deveria ter digito invalido")
	}
}

func TestExemploMultiItens_TotalCoerente(t *testing.T) {
	xml, err := exemploPorID(t, "multi-itens").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := nfgas.LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Det) != 5 {
		t.Fatalf("itens = %d, esperado 5", len(n.Det))
	}
	var soma float64
	csts := map[string]bool{}
	for i, d := range n.Det {
		if d.NItem != i+1 {
			t.Errorf("nItem fora de ordem: %d na posicao %d", d.NItem, i+1)
		}
		soma += d.GNormal.Prod.VProd
		csts[d.GNormal.Imposto.ICMS.CST.String()] = true
	}
	if dif := soma - n.Total.VProd; dif > 0.01 || dif < -0.01 {
		t.Errorf("vProd do total = %.2f, soma dos itens = %.2f", n.Total.VProd, soma)
	}
	if len(csts) < 3 {
		t.Errorf("esperados varios CST de ICMS, vieram %v", csts)
	}

	// O parametro itens muda a quantidade.
	xml, err = exemploPorID(t, "multi-itens").Gerar(ParametrosExemplo{Itens: 3})
	if err != nil {
		t.Fatal(err)
	}
	if n, err = nfgas.LerXMLString(xml); err != nil {
		t.Fatal(err)
	}
	if len(n.Det) != 3 {
		t.Fatalf("itens=3 gerou %d itens", len(n.Det))
	}
}

func TestExemploMultiCFOP_CFOPsDistintos(t *testing.T) {
	xml, err := exemploPorID(t, "multi-cfop").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := nfgas.LerXMLString(xml)
	if err != nil {
		t.Fatal(err)
	}
	cfops := map[int]bool{}
	for _, d := range n.Det {
		if d.GNormal.Prod.CFOP == 0 {
			t.Error("item sem CFOP")
		}
		cfops[d.GNormal.Prod.CFOP] = true
	}
	if len(cfops) < 3 {
		t.Fatalf("esperados ao menos 3 CFOP distintos, vieram %v", cfops)
	}
}

func TestExemploCancelamento_AmarraNaChave(t *testing.T) {
	// Evento apontando para chave que o usuario nao tem nao serve de teste
	// de importacao: os dois cenarios precisam falar do mesmo documento.
	doc, err := exemploPorID(t, "transmissao").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	n, err := nfgas.LerXMLString(doc)
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"cancelamento", "cancelamento-proc"} {
		xml, err := exemploPorID(t, id).Gerar(ParametrosExemplo{})
		if err != nil {
			t.Fatal(err)
		}
		ev, err := nfgas.LerEventoString(xml)
		if err != nil {
			t.Fatalf("%s: %v", id, err)
		}
		if ev.ChaveAcesso() != n.ChaveAcesso() {
			t.Errorf("%s aponta para %s, documento e %s", id, ev.ChaveAcesso(), n.ChaveAcesso())
		}
		if !ev.Cancelamento() || ev.Justificativa() == "" {
			t.Errorf("%s: cancelamento=%v xJust=%q", id, ev.Cancelamento(), ev.Justificativa())
		}
	}

	// So o procEvento tem retorno homologado.
	xml, _ := exemploPorID(t, "cancelamento-proc").Gerar(ParametrosExemplo{})
	ev, err := nfgas.LerEventoString(xml)
	if err != nil {
		t.Fatal(err)
	}
	if !ev.RetInfEvento.Registrado() {
		t.Errorf("procEvento deveria vir registrado, cStat = %d", ev.RetInfEvento.CStat)
	}
}

// ---------------------------------------------------------------------------
// Cobertura de tags do cenario "completo"
// ---------------------------------------------------------------------------

var reTag = regexp.MustCompile(`<([a-zA-Z][\w]*)[\s/>]`)

// tagsDoXML devolve o conjunto de nomes de tag presentes no XML.
func tagsDoXML(xml string) map[string]bool {
	tags := map[string]bool{}
	for _, m := range reTag.FindAllStringSubmatch(xml, -1) {
		tags[m[1]] = true
	}
	return tags
}

func tagsDoArquivo(t *testing.T, caminho string) map[string]bool {
	t.Helper()
	dados, err := os.ReadFile(caminho)
	if err != nil {
		t.Fatalf("ler fixture %s: %v", caminho, err)
	}
	return tagsDoXML(string(dados))
}

// tagsForaDoWriter sao tags que a fixture tem e NENHUM documento gerado pelo
// writer pode conter. As fixtures existem para exercitar o LEITOR, que aceita
// combinacoes que o gerador nunca produz; por isso a lista nao e vazia.
//
// Lista fechada e revisada: tag nova aqui = o writer perdeu capacidade de
// emitir algo; tag que sai = a cobertura melhorou. Nos dois casos a mudanca
// tem que passar por revisao, nunca ser atualizada no automatico.
var tagsForaDoWriter = []string{
	// Existem na classe do ACBr e sao lidas de enderEmit/enderDest, mas nao
	// existem no TEndeEmi do XSD -- packages/nfgas/classes.go:83-86.
	"cPais", "xPais",

	// gerarProtNFGas (xml_writer.go:825) nao emite cMsg/xMsg, embora o
	// pcn.ProcDFe tenha os campos e o leitor os leia.
	"cMsg", "xMsg",

	// RAMO MORTO, e o defeito e do ACBr, nao do porte: gerarProd so chama
	// gerarGMedicao quando nMed > 0 E gMedida.vMed > 0 (xml_writer.go:377),
	// mas dentro de gerarGMedicao o tpMotNaoLeitura so sai no ramo
	// vMed == 0 (xml_writer.go:415). A mesma guarda esta em
	// ACBrNFGas.XmlWriter.pas:687, antes do Gerar_det_prod_gMedicao, entao
	// nem o Delphi consegue emitir essas duas tags.
	"tpMotNaoLeitura", "xMotNaoLeitura",
}

// tagsExclusivasDaFixture sao tags que o writer SABE emitir, mas nao junto
// com as escolhas deste exemplo -- o leiaute as poe em alternativa com outra
// tag que o exemplo preferiu por cobrir mais campos.
var tagsExclusivasDaFixture = []string{
	// NB so sai quando cNIS esta vazio (xml_writer.go:250).
	"NB",
	// codDebAuto exclui o par codBanco+codAgencia (xml_writer.go:740).
	"codDebAuto",
}

func TestExemploCompleto_CobreTagsDaFixture(t *testing.T) {
	xml, err := exemploPorID(t, "completo").Gerar(ParametrosExemplo{})
	if err != nil {
		t.Fatal(err)
	}
	geradas := tagsDoXML(xml)

	base := filepath.Join("..", "..", "packages", "nfgas", "testdata")
	alvo := map[string]bool{}
	for _, fixture := range []string{"nfgas_completa.xml", "nfgas_proc.xml"} {
		for tag := range tagsDoArquivo(t, filepath.Join(base, fixture)) {
			alvo[tag] = true
		}
	}
	for _, tag := range tagsForaDoWriter {
		delete(alvo, tag)
	}
	for _, tag := range tagsExclusivasDaFixture {
		delete(alvo, tag)
	}

	var faltando []string
	for tag := range alvo {
		if !geradas[tag] {
			faltando = append(faltando, tag)
		}
	}
	sort.Strings(faltando)
	if len(faltando) > 0 {
		t.Fatalf("o exemplo 'completo' nao emite %d tag(s) que a fixture tem: %v",
			len(faltando), faltando)
	}

	// As duas listas de excecao tem que continuar sendo excecao de verdade:
	// se o exemplo passar a emitir uma delas, a lista e que esta errada.
	for _, tag := range append(append([]string{}, tagsForaDoWriter...), tagsExclusivasDaFixture...) {
		if geradas[tag] {
			t.Errorf("tag %q esta na lista de inalcancaveis mas foi emitida; "+
				"atualize a lista", tag)
		}
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func servidorExemplos() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/exemplos", handleListarExemplos)
	mux.HandleFunc("GET /api/exemplos/{id}", handleExemplo)
	mux.HandleFunc("GET /api/exemplos/{id}/download", handleBaixarExemplo)
	return mux
}

func TestHandlersExemplos(t *testing.T) {
	mux := servidorExemplos()

	t.Run("lista", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exemplos", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var resp ExemplosResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Total != len(catalogoExemplos()) || len(resp.Exemplos) != resp.Total {
			t.Fatalf("total = %d, exemplos = %d", resp.Total, len(resp.Exemplos))
		}
	})

	t.Run("item sem query gera completo", func(t *testing.T) {
		// O caminho do "clicar e gerar": nenhum parametro, XML pronto.
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exemplos/completo", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d: %s", w.Code, w.Body)
		}
		var resp ExemploResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp.XML == "" || resp.Chave == "" {
			t.Fatalf("resposta sem xml ou sem chave: %+v", resp.ExemploResumo)
		}
	})

	t.Run("id inexistente", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exemplos/naoexiste", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, esperado 404", w.Code)
		}
		var corpo map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &corpo); err != nil || corpo["erro"] == "" {
			t.Fatalf("corpo = %s", w.Body)
		}
	})

	t.Run("parametro invalido", func(t *testing.T) {
		for _, q := range []string{"?uf=ZZ", "?nnf=abc", "?tpamb=3"} {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exemplos/transmissao"+q, nil))
			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: status = %d, esperado 400", q, w.Code)
			}
		}
	})

	t.Run("download", func(t *testing.T) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exemplos/transmissao/download", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); ct != "application/xml; charset=utf-8" {
			t.Errorf("Content-Type = %q", ct)
		}
		if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "nfgas-transmissao.xml") {
			t.Errorf("Content-Disposition = %q", cd)
		}
		esperado, err := exemploPorID(t, "transmissao").Gerar(ParametrosExemplo{})
		if err != nil {
			t.Fatal(err)
		}
		// O arquivo baixado leva o prolog; o XML que vai para a SEFAZ, nao.
		if w.Body.String() != prologXML+esperado {
			t.Error("o corpo do download diverge do XML do builder")
		}
		if !strings.HasPrefix(w.Body.String(), `<?xml version="1.0" encoding="UTF-8"?>`) {
			t.Error("o arquivo baixado deveria comecar pelo prolog XML")
		}
	})
}
