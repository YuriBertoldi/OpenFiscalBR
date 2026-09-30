// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-30.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

const (
	chaveTeste     = "35260311222333000181750010000000011100000016"
	chaveProcTeste = "35260311222333000181750010000000021100000099"
	chaveAntTeste  = "35260211222333000181750010000000011100000055"
)

func carregarFixture(t *testing.T, nome string) []byte {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join("testdata", nome))
	if err != nil {
		t.Fatalf("abrir fixture %s: %v", nome, err)
	}
	return dados
}

func lerCompleta(t *testing.T) *NFAg {
	t.Helper()
	n, err := LerBytes(carregarFixture(t, "nfag_completa.xml"))
	if err != nil {
		t.Fatalf("LerBytes: %v", err)
	}
	return n
}

func testContext() context.Context { return context.Background() }

// ---------------------------------------------------------------------------
// Enums -- codigo exato do leiaute
// ---------------------------------------------------------------------------

func TestEnums_CodigoDoLeiaute(t *testing.T) {
	casos := []struct {
		nome     string
		got      string
		esperado string
	}{
		{"Ve100", Ve100.String(), "1.00"},
		{"Sa0", Sa0.String(), "0"},
		{"Sa9", Sa9.String(), "9"},
		{"FnNormal", FnNormal.String(), "0"},
		{"FnSubstituicao", FnSubstituicao.String(), "3"},
		{"TfNormal", TfNormal.String(), "1"},
		{"TfConjunto", TfConjunto.String(), "3"},
		{"TlAgua", TlAgua.String(), "1"},
		{"TlAguaEsgoto", TlAguaEsgoto.String(), "3"},
		{"MsErroLeitura", MsErroLeitura.String(), "01"},
		{"MsErroTributacao", MsErroTributacao.String(), "05"},
		{"IoMedia", IoMedia.String(), "1"},
		{"IoSemQuantidade", IoSemQuantidade.String(), "6"},
		{"TgmAguaTratada", TgmAguaTratada.String(), "01"},
		{"TgmOutros", TgmOutros.String(), "99"},
		// UMedFat SALTA de "2" para "5" (nao ha 3 e 4)
		{"UmM3", UmM3.String(), "1"},
		{"UmLitros", UmLitros.String(), "2"},
		{"UmUnidade", UmUnidade.String(), "5"},
		{"UmTon", UmTon.String(), "6"},
		// TpCategoria tem SALTOS (nao ha 03 nem 05)
		{"TcComercial", TcComercial.String(), "01"},
		{"TcIndustrial", TcIndustrial.String(), "04"},
		{"TcResidencial", TcResidencial.String(), "06"},
		{"TcOutros", TcOutros.String(), "99"},
		// TpProc da NFAg comeca em "0" (adm ESTADUAL)
		{"TpProcAdmEstadual", TpProcAdmEstadual.String(), "0"},
		{"TpProcon", TpProcon.String(), "5"},
		{"TmConsumidor", TmConsumidor.String(), "1"},
		{"TeCancelamento", TeCancelamento.String(), "110111"},
		{"TeAutorizadoSubstituicao", TeAutorizadoSubstituicao.String(), "240140"},
		{"TeAutorizadoAjuste", TeAutorizadoAjuste.String(), "240150"},
		{"TeLiberacaoPrazoCancelado", TeLiberacaoPrazoCancelado.String(), "240170"},
	}
	for _, c := range casos {
		if c.got != c.esperado {
			t.Errorf("%s.String() = %q, esperado %q", c.nome, c.got, c.esperado)
		}
	}
}

func TestEnums_ParseRoundTrip(t *testing.T) {
	if v, err := ParseUMedFat("5"); err != nil || v != UmUnidade {
		t.Errorf("ParseUMedFat(5) = %v, %v", v, err)
	}
	if v, err := ParseTpCategoria("04"); err != nil || v != TcIndustrial {
		t.Errorf("ParseTpCategoria(04) = %v, %v", v, err)
	}
	if _, err := ParseTpCategoria("03"); err == nil {
		t.Error("ParseTpCategoria(03) deveria falhar: o codigo nao existe na tabela")
	}
	if v, err := ParseTipoEvento("240170"); err != nil || v != TeLiberacaoPrazoCancelado {
		t.Errorf("ParseTipoEvento(240170) = %v, %v", v, err)
	}
}

// ---------------------------------------------------------------------------
// Leitura da fixture completa -- campo a campo
// ---------------------------------------------------------------------------

func TestLerXML_Ide(t *testing.T) {
	n := lerCompleta(t)
	ide := n.Ide

	if n.InfNFAg.ID != "NFAG"+chaveTeste {
		t.Errorf("ID = %q", n.InfNFAg.ID)
	}
	if n.InfNFAg.Versao != 1.00 {
		t.Errorf("versao = %v", n.InfNFAg.Versao)
	}
	if ide.CUF != 35 || ide.Modelo != 75 || ide.Serie != 1 || ide.NNF != 1 ||
		ide.CNF != 1 || ide.CDV != 6 || ide.CMunFG != 3550308 {
		t.Errorf("ide numericos: %+v", ide)
	}
	if ide.TpAmb != pcn.TaHomologacao || ide.TpEmis != pcn.TeNormal ||
		ide.NSiteAutoriz != Sa1 || ide.FinNFAg != FnNormal {
		t.Errorf("ide enums: %+v", ide)
	}
	// tpFat NAO e lido do XML (omissao do ACBr replicada): fica no zero
	if ide.TpFat != TfNormal {
		t.Errorf("tpFat deveria ficar no zero-value TfNormal, veio %v", ide.TpFat)
	}
	if ide.DhEmi.IsZero() || ide.DhEmi.Format("2006-01-02T15:04:05") != "2026-03-15T10:30:00" {
		t.Errorf("dhEmi = %v", ide.DhEmi)
	}
	if ide.TpPagAnt.String() != "2" {
		t.Errorf("tpPagAnt = %v", ide.TpPagAnt)
	}
	if ide.GCompraGov.PRedutor != 12.3456 || ide.GCompraGov.TpEnteGov.String() != "2" {
		t.Errorf("gCompraGov: %+v", ide.GCompraGov)
	}
}

func TestLerXML_EmitDestLigacao(t *testing.T) {
	n := lerCompleta(t)

	if n.Emit.CNPJ != "11222333000181" || n.Emit.XNome != "Saneamento Teste LTDA" ||
		n.Emit.ISUFEmit != "ISUF123" || n.Emit.EnderEmit.CEP != 1310100 {
		t.Errorf("emit: %+v", n.Emit)
	}
	if n.Dest.CNPJCPF != "52998224725" || n.Dest.IE != "ISENTO" ||
		n.Dest.CNIS != "12345678901" || n.Dest.NB != "1234567890" {
		t.Errorf("dest: %+v", n.Dest)
	}
	// idOutros so e lido quando CNPJCPF vazio -- aqui tem CPF, fica vazio
	if n.Dest.IDOutros != "" {
		t.Errorf("idOutros deveria ser descartado com CPF presente: %q", n.Dest.IDOutros)
	}
	if n.Ligacao.IDLigacao != "LIG-000123" || n.Ligacao.TpLigacao != TlAguaEsgoto ||
		n.Ligacao.LatGPS != "-23.550520" || n.Ligacao.CodRoteiroLeitura != "ROT-A12" {
		t.Errorf("ligacao: %+v", n.Ligacao)
	}
}

func TestLerXML_GruposDoTopo(t *testing.T) {
	n := lerCompleta(t)

	if n.GSub.ChNFAg != chaveAntTeste || n.GSub.MotSub != MsErroLeitura {
		t.Errorf("gSub: %+v", n.GSub)
	}
	if len(n.GMed) != 2 || n.GMed[0].NMed != 1 || n.GMed[0].IDMedidor != "HIDRO-0001" ||
		n.GMed[1].NMed != 2 {
		t.Errorf("gMed: %+v", n.GMed)
	}
	if n.GMed[0].DMedAnt.Format("2006-01-02") != "2026-02-15" {
		t.Errorf("dMedAnt = %v", n.GMed[0].DMedAnt)
	}
	if n.GFatConjunto.ChNFAgFat != chaveProcTeste {
		t.Errorf("gFatConjunto: %+v", n.GFatConjunto)
	}
}

func TestLerXML_DetProd(t *testing.T) {
	n := lerCompleta(t)
	if len(n.Det) != 2 {
		t.Fatalf("det = %d itens", len(n.Det))
	}
	d := n.Det[0]
	p := d.Prod

	if d.NItem != 1 || len(d.GTarif) != 2 {
		t.Fatalf("det1: nItem=%d gTarif=%d", d.NItem, len(d.GTarif))
	}
	if d.GTarif[0].NAto != "0123" || d.GTarif[0].TpFaixaCons != TfcMinimo ||
		!d.GTarif[1].DFimTarif.IsZero() {
		t.Errorf("gTarif: %+v", d.GTarif)
	}
	if p.IndOrigemQtd != IoMedido || p.CProd != "AGUA001" {
		t.Errorf("prod: %+v", p)
	}
	// cClass e STRING e preserva o zero a esquerda (divergencia de tipo)
	if p.CClass != "0100101" {
		t.Errorf("cClass = %q, o zero a esquerda se perdeu", p.CClass)
	}
	if p.TpCategoria != TcResidencial || p.XCategoria != "Residencial padrao" ||
		p.QEconomias != "2" || p.UMed != UmM3 {
		t.Errorf("categorias: %+v", p)
	}
	// qFaturada e float64 lido com tcDe4 (divergencia de tipo: o ACBr
	// arredondaria para Integer)
	if p.QFaturada != 250.4478 {
		t.Errorf("qFaturada = %v, esperado 250.4478", p.QFaturada)
	}
	// vItem/vProd tcDe10
	if p.VItem != 3.1234567891 {
		t.Errorf("vItem = %v, esperado 10 casas", p.VItem)
	}
	if p.VProd != 782.1234567891 {
		t.Errorf("vProd = %v, esperado 10 casas", p.VProd)
	}
	if p.FatorPoluicao != 1.05 {
		t.Errorf("fatorPoluicao = %v", p.FatorPoluicao)
	}
	if p.IndDevolucao != pcn.TieSim {
		t.Errorf("indDevolucao = %v", p.IndDevolucao)
	}
	if p.GPagAntecipado.ChDFePagAnt != chaveAntTeste || p.GPagAntecipado.NItemPagAnt != 1 {
		t.Errorf("gPagAntecipado: %+v", p.GPagAntecipado)
	}

	// det2: atributos de item anterior + prod sem indDevolucao
	d2 := n.Det[1]
	if d2.ChNFAgAnt != chaveAntTeste || d2.NItemAnt != 1 {
		t.Errorf("det2 atributos: %+v", d2)
	}
	if d2.Prod.IndDevolucao != pcn.TieNenhum {
		t.Errorf("det2 indDevolucao ausente deveria ser TieNenhum: %v", d2.Prod.IndDevolucao)
	}
}

func TestLerXML_GMedicaoEGMedida(t *testing.T) {
	n := lerCompleta(t)
	g := n.Det[0].Prod.GMedicao

	if g.NMed != 1 {
		t.Errorf("nMed = %d", g.NMed)
	}
	m := g.GMedida
	if m.TpGrMed != TgmAguaTratada || m.NUnidConsumo != "UC-9988" || m.UMed != UmM3 {
		t.Errorf("gMedida: %+v", m)
	}
	// tudo tcDe2 no gMedida
	if m.VUnidConsumo != 3.75 || m.VMedAnt != 1000.12 || m.VMedAtu != 1250.56 ||
		m.VConst != 1.00 || m.VMed != 250.44 {
		t.Errorf("gMedida valores: %+v", m)
	}
}

func TestLerXML_Imposto(t *testing.T) {
	n := lerCompleta(t)
	imp := n.Det[0].Imposto

	if imp.IBSCBS.CST.String() != "000" || imp.IBSCBS.CClassTrib != "000001" {
		t.Errorf("IBSCBS: %+v", imp.IBSCBS)
	}
	if imp.IBSCBS.GIBSCBS.VBC != 782.12 || imp.IBSCBS.GIBSCBS.GIBSUF.GDif.PDif != 50 {
		t.Errorf("gIBSCBS: %+v", imp.IBSCBS.GIBSCBS)
	}
	if imp.PIS.CST.String() != "01" || imp.PIS.PPIS != 1.65 || imp.PIS.VPIS != 12.90 {
		t.Errorf("PIS: %+v", imp.PIS)
	}
	if imp.COFINS.PCOFINS != 7.60 || imp.COFINS.VCOFINS != 59.44 {
		t.Errorf("COFINS: %+v", imp.COFINS)
	}
	// vRetCofins e a grafia da tag; vBCIRRF nunca e lido
	if imp.RetTrib.VRetCOFINS != 2.22 || imp.RetTrib.VIRRF != 4.44 {
		t.Errorf("retTrib: %+v", imp.RetTrib)
	}
	if imp.RetTrib.VBCIRRF != 0 {
		t.Errorf("vBCIRRF nunca e lido do XML (omissao do ACBr): %v", imp.RetTrib.VBCIRRF)
	}
	if imp.TFS.VBCTFS != 782.12 || imp.TFS.PTFS != 0.50 || imp.TFS.VTFS != 3.91 {
		t.Errorf("TFS: %+v", imp.TFS)
	}
	if imp.TFU.VBCTFU != 782.12 || imp.TFU.PTFU != 0.25 || imp.TFU.VTFU != 1.96 {
		t.Errorf("TFU: %+v", imp.TFU)
	}
}

func TestLerXML_GProcRef(t *testing.T) {
	n := lerCompleta(t)
	g := n.Det[0].GProcRef

	// vItem/vProd tcDe8; qFaturada tcDe4
	if g.VItem != 2.96296285 || g.VProd != 31.11111003 || g.QFaturada != 10.5 {
		t.Errorf("gProcRef: %+v", g)
	}
	if g.IndDevolucao != pcn.TieSim {
		t.Errorf("indDevolucao = %v", g.IndDevolucao)
	}
	if len(g.GProc) != 2 || g.GProc[0].TpProc != TpProcAdmEstadual ||
		g.GProc[1].TpProc != TpJusticaEstadual {
		t.Errorf("gProc: %+v", g.GProc)
	}
}

func TestLerXML_TotalEGFat(t *testing.T) {
	n := lerCompleta(t)
	tot := n.Total

	if tot.VProd != 832.12 || tot.VNF != 910.33 || tot.VTotDFe != 910.33 {
		t.Errorf("total: %+v", tot)
	}
	if tot.VTFS != 3.91 || tot.VTFU != 1.96 {
		t.Errorf("taxas do total: %+v", tot)
	}
	if tot.VRetCOFINS != 2.22 || tot.VIRRF != 4.44 {
		t.Errorf("vRetTribTot achatado: %+v", tot)
	}
	if tot.IBSCBSTot.VBCIBSCBS != 782.12 || tot.IBSCBSTot.GIBS.VIBS != 69.22 {
		t.Errorf("IBSCBSTot: %+v", tot.IBSCBSTot)
	}

	g := n.GFat
	if g.CompetFat.Format("200601") != "202603" {
		t.Errorf("CompetFat = %v", g.CompetFat)
	}
	if g.DVencFat.Format("2006-01-02") != "2026-04-10" || g.NFat != "FAT-2026-03-0001" ||
		g.CodDebAuto != "DEB-001" {
		t.Errorf("gFat: %+v", g)
	}
	if g.EnderCorresp.XLgr != "Rua da Correspondencia" || g.EnderCorresp.CEP != 1415000 {
		t.Errorf("enderCorresp: %+v", g.EnderCorresp)
	}
	if g.GPIX.URLQRCodePIX == "" {
		t.Error("gPIX nao lido")
	}
}

func TestLerXML_GAgenciaEQualiAgua(t *testing.T) {
	n := lerCompleta(t)
	a := n.GAgencia

	if a.Econ != "12345" || a.SPrestador != "SELO-P1" || a.SRegulador != "ARSESP" ||
		a.NAgenciaAtend != "AG-CENTRO-01" {
		t.Errorf("gAgencia: %+v", a)
	}
	if a.DEmissSelo.Format("2006-01-02") != "2026-01-10" {
		t.Errorf("dEmissSelo = %v", a.DEmissSelo)
	}
	if len(a.GHistCons) != 1 || a.GHistCons[0].MedMensal != 250.0 ||
		len(a.GHistCons[0].GCons) != 2 {
		t.Fatalf("gHistCons: %+v", a.GHistCons)
	}
	c := a.GHistCons[0].GCons[0]
	// qtdDias e STRING (tcStr) e preserva zeros
	if c.QtdDias != "00030" {
		t.Errorf("qtdDias = %q", c.QtdDias)
	}
	if c.MedDiaria != 8.3312 || c.Consumo != 249.936 || c.VolFat != 250.00 {
		t.Errorf("gCons: %+v", c)
	}
	if c.CompetFat.Format("200601") != "202601" {
		t.Errorf("CompetFat do gCons = %v", c.CompetFat)
	}

	q := n.GQualiAgua
	if q.CompetAnalise.Format("200601") != "202602" || q.Conclusao != "Agua propria para consumo" ||
		q.SistemaAbast != "Sistema Cantareira" {
		t.Errorf("gQualiAgua: %+v", q)
	}
	if len(q.GAnalise) != 2 || q.GAnalise[0].XItemAnalisado != "Cloro residual livre" ||
		q.GAnalise[0].NAmostraMinima != "0120" || q.GAnalise[1].XItemAnalisado != "Turbidez" {
		t.Errorf("gAnalise: %+v", q.GAnalise)
	}
}

func TestLerXML_GruposFinais(t *testing.T) {
	n := lerCompleta(t)

	if len(n.AutXML) != 2 || n.AutXML[0].CNPJCPF != "11222333000181" ||
		n.AutXML[1].CNPJCPF != "52998224725" {
		t.Errorf("autXML: %+v", n.AutXML)
	}
	if n.InfAdic.InfAdFisco != "Informacao ao fisco" || len(n.InfAdic.InfCpl) != 1 {
		t.Errorf("infAdic: %+v", n.InfAdic)
	}
	if n.InfPAA.CNPJPAA != "10880919000160" {
		t.Errorf("infPAA: %+v", n.InfPAA)
	}
	if n.InfRespTec.CNPJ != "10880919000160" || n.InfRespTec.IDCSRT != 1 ||
		n.InfRespTec.HashCSRT == "" {
		t.Errorf("gRespTec: %+v", n.InfRespTec)
	}
	// CDATA removido (primeira ocorrencia de cada marcador)
	if !strings.HasPrefix(n.InfNFAgSupl.QrCodNFAg, "https://") ||
		strings.Contains(n.InfNFAgSupl.QrCodNFAg, "CDATA") {
		t.Errorf("qrCodNFAg = %q", n.InfNFAgSupl.QrCodNFAg)
	}
}

// ---------------------------------------------------------------------------
// Documento processado e lote
// ---------------------------------------------------------------------------

func TestLerXML_Proc(t *testing.T) {
	n, err := LerBytes(carregarFixture(t, "nfag_proc.xml"))
	if err != nil {
		t.Fatalf("LerBytes: %v", err)
	}
	if n.ProcNFAg.NProt != "335260000000101" || n.ProcNFAg.CStat != 100 ||
		n.ProcNFAg.ChDFe != chaveProcTeste {
		t.Errorf("protNFAg: %+v", n.ProcNFAg)
	}
	if !n.Confirmada() || !n.Processada() || n.Cancelada() {
		t.Errorf("situacao: confirmada=%v processada=%v cancelada=%v",
			n.Confirmada(), n.Processada(), n.Cancelada())
	}
	// idOutros lido porque nao ha CNPJ/CPF
	if n.Dest.IDOutros != "MATRICULA-778899" {
		t.Errorf("idOutros = %q", n.Dest.IDOutros)
	}
	// imposto vazio (<imposto/>) nao derruba a leitura
	if len(n.Det) != 1 {
		t.Errorf("det: %d", len(n.Det))
	}
}

func TestLerLote_Misto(t *testing.T) {
	notas, err := LerLoteBytes(carregarFixture(t, "lote_misto.xml"))
	if len(notas) != 2 {
		t.Fatalf("lote deveria ter 2 legiveis, veio %d (err=%v)", len(notas), err)
	}
	var lote *ErrosLote
	if !errors.As(err, &lote) || len(lote.Erros) != 1 {
		t.Fatalf("deveria haver 1 falha no lote, veio %v", err)
	}
	if notas[0].ChaveAcesso() != chaveTeste || notas[1].ChaveAcesso() != chaveProcTeste {
		t.Errorf("chaves: %q, %q", notas[0].ChaveAcesso(), notas[1].ChaveAcesso())
	}
	if notas[1].NFAg.ProcNFAg.NProt == "" {
		t.Error("nota processada do lote perdeu o protocolo")
	}
}

// ---------------------------------------------------------------------------
// Evento e retorno de consulta
// ---------------------------------------------------------------------------

func TestLerEvento_ProcCancelamento(t *testing.T) {
	ev, err := LerEventoBytes(carregarFixture(t, "proc_evento_cancelamento.xml"))
	if err != nil {
		t.Fatalf("LerEventoBytes: %v", err)
	}
	if ev.RetInfEvento.CStat != 135 || !ev.RetInfEvento.Registrado() {
		t.Errorf("retEvento: %+v", ev.RetInfEvento)
	}
	if ev.RetInfEvento.TpEvento != TeCancelamento || !ev.Cancelamento() {
		t.Errorf("tipo: %v", ev.RetInfEvento.TpEvento)
	}
	if ev.Evento.InfEvento.DetEvento.XJust != "Erro de medicao identificado apos emissao" {
		t.Errorf("justificativa: %q", ev.Evento.InfEvento.DetEvento.XJust)
	}
	if ev.ChaveAcesso() != chaveProcTeste {
		t.Errorf("chave: %q", ev.ChaveAcesso())
	}
}

func TestLerRetConsSit(t *testing.T) {
	ret, err := LerRetConsSitBytes(carregarFixture(t, "ret_cons_sit.xml"))
	if err != nil {
		t.Fatalf("LerRetConsSitBytes: %v", err)
	}
	if ret.CStat != 100 || ret.ChNFAg != chaveProcTeste || !ret.Autorizada() {
		t.Errorf("retConsSit: %+v", ret)
	}
	if ret.ProtNFAg.NProt != "335260000000101" || ret.ProtNFAg.DigVal != "abc123=" {
		t.Errorf("protocolo: %+v", ret.ProtNFAg)
	}
}

// ---------------------------------------------------------------------------
// INI -- leitura e round-trip (com os dois bugs do ACBr replicados)
// ---------------------------------------------------------------------------

func TestLerINI_Completa(t *testing.T) {
	n, err := LerINI(string(carregarFixture(t, "nfag_acbr.ini")), Configuracoes{VersaoDF: Ve100})
	if err != nil {
		t.Fatalf("LerINI: %v", err)
	}
	// a versao vem da secao [infNFGas] -- bug do ACBr replicado; a fixture
	// usa exatamente essa secao, como um .ini legado do ACBr teria de usar
	if n.InfNFAg.Versao != 1.00 {
		t.Errorf("versao = %v", n.InfNFAg.Versao)
	}
	if n.Ide.CUF != 35 || n.Ide.Modelo != 75 || n.Ide.TpFat != TfNormal {
		t.Errorf("ide: %+v", n.Ide)
	}
	if n.Ligacao.IDLigacao != "LIG-000123" || n.Ligacao.TpLigacao != TlAguaEsgoto {
		t.Errorf("ligacao: %+v", n.Ligacao)
	}
	if len(n.Det) != 1 {
		t.Fatalf("det: %d", len(n.Det))
	}
	d := n.Det[0]
	if d.Prod.CClass != "0100101" || d.Prod.QFaturada != 250.4478 ||
		d.Prod.VItem != 3.1234567891 {
		t.Errorf("prod do ini: %+v", d.Prod)
	}
	if d.Prod.GMedicao.GMedida.VMed != 250.44 || d.Prod.GMedicao.GMedida.TpGrMed != TgmAguaTratada {
		t.Errorf("gMedicao/gMedida: %+v", d.Prod.GMedicao)
	}
	if len(d.GTarif) != 1 || d.GTarif[0].NAto != "0123" {
		t.Errorf("gTarif: %+v", d.GTarif)
	}
	if d.Imposto.TFS.VTFS != 3.91 {
		t.Errorf("TFS do ini: %+v", d.Imposto.TFS)
	}
	// TFU001 existe na fixture (montada a mao) e E lida -- o que o proprio
	// ACBr nunca produz, porque grava TFU na secao TFS (bug 2)
	if d.Imposto.TFU.VTFU != 1.96 {
		t.Errorf("TFU do ini: %+v", d.Imposto.TFU)
	}
	if len(n.GAgencia.GHistCons) != 1 || len(n.GAgencia.GHistCons[0].GCons) != 1 {
		t.Errorf("gHistCons do ini: %+v", n.GAgencia.GHistCons)
	}
	if n.GQualiAgua.Conclusao == "" || len(n.GQualiAgua.GAnalise) != 1 {
		t.Errorf("gQualiAgua do ini: %+v", n.GQualiAgua)
	}
	if n.InfPAA.CNPJPAA != "10880919000160" {
		t.Errorf("infPAA do ini: %+v", n.InfPAA)
	}
}

func TestINI_RoundTripComBugsDoACBr(t *testing.T) {
	original, err := LerINI(string(carregarFixture(t, "nfag_acbr.ini")), Configuracoes{VersaoDF: Ve100})
	if err != nil {
		t.Fatal(err)
	}

	// o IniReader (como o do ACBr) NAO le o ID; GravarINI recusa chave
	// invalida (raise do IniWriter.pas:148-149), entao o chamador monta o
	// ID como o GerarXML faria
	original.InfNFAg.ID = LiteralChave + chaveTeste

	gravado, err := GravarINI(original)
	if err != nil {
		t.Fatalf("GravarINI: %v", err)
	}

	// BUG 1 replicado: a versao e gravada em [infNFAg]...
	if !strings.Contains(gravado, "[infNFAg]") {
		t.Error("versao deveria ser gravada na secao [infNFAg]")
	}
	// BUG 2 replicado: o TFU e gravado na secao do TFS
	if !strings.Contains(gravado, "vTFU=") {
		t.Fatal("TFU nao foi gravado")
	}
	if strings.Contains(gravado, "[TFU001]") {
		t.Error("TFU deveria ser gravado na secao [TFS001] (bug do ACBr replicado)")
	}

	relido, err := LerINI(gravado, Configuracoes{VersaoDF: Ve100})
	if err != nil {
		t.Fatalf("reler ini gravado: %v", err)
	}
	// ...e o reader le de [infNFGas]: a versao cai no default (perda do
	// proprio ACBr, replicada)
	if relido.InfNFAg.Versao != Ve100.Float() {
		t.Errorf("versao relida = %v", relido.InfNFAg.Versao)
	}
	// TFU gravado nunca e relido (secao errada)
	if relido.Det[0].Imposto.TFU.VTFU != 0 {
		t.Errorf("TFU relido deveria ser 0 (bug replicado): %+v", relido.Det[0].Imposto.TFU)
	}
	// o resto sobrevive ao round-trip
	if relido.Det[0].Prod.QFaturada != original.Det[0].Prod.QFaturada ||
		relido.Det[0].Imposto.TFS != original.Det[0].Imposto.TFS ||
		relido.GAgencia.GHistCons[0].GCons[0].QtdDias != "00030" {
		t.Errorf("round-trip perdeu dados: %+v", relido.Det[0])
	}
}

func TestGravarINI_ChaveInvalidaERecusada(t *testing.T) {
	// GravarIni do original raise "Chave Invalida" (IniWriter.pas:148-149)
	if _, err := GravarINI(&NFAg{}); !errors.Is(err, pcn.ErrChaveInvalida) {
		t.Errorf("GravarINI sem chave deveria falhar com ErrChaveInvalida, veio: %v", err)
	}
}

func TestINI_RoundTripDaNotaCompleta(t *testing.T) {
	// nota vinda do XML: ID valido, dhEmi com hora, gPagAntecipado no item
	original := lerCompleta(t)

	gravado, err := GravarINI(original)
	if err != nil {
		t.Fatalf("GravarINI: %v", err)
	}
	relido, err := LerINI(gravado, Configuracoes{VersaoDF: Ve100, Ambiente: original.Ide.TpAmb})
	if err != nil {
		t.Fatalf("reler ini gravado: %v", err)
	}

	// a HORA de dhEmi sobrevive (DateTimeToIni do ACBr grava data e hora;
	// gravar so a data zeraria a emissao no XML regerado)
	const layout = "2006-01-02T15:04:05"
	if relido.Ide.DhEmi.Format(layout) != original.Ide.DhEmi.Format(layout) {
		t.Errorf("dhEmi apos round-trip = %v, original %v", relido.Ide.DhEmi, original.Ide.DhEmi)
	}
	// DIVERGENCIA gPagAntecipado: o ACBr grava 0-based e le 1-based (perde o
	// grupo); aqui o round-trip preserva
	if relido.Det[0].Prod.GPagAntecipado != original.Det[0].Prod.GPagAntecipado {
		t.Errorf("gPagAntecipado apos round-trip = %+v, original %+v",
			relido.Det[0].Prod.GPagAntecipado, original.Det[0].Prod.GPagAntecipado)
	}
}

func TestGravarINI_GuardasDeSecao(t *testing.T) {
	n := lerCompleta(t)

	// Gerar_Ligacao: sem idLigacao a secao [Ligacao] nao sai
	n.Ligacao = Ligacao{IDCodCliente: "999"}
	// Gerar_gMedicao so pula com nMed<=0 E vMed=0: nMed=0 com vMed>0 SAI
	n.Det[0].Prod.GMedicao.NMed = 0
	n.Det[0].Prod.GMedicao.GMedida.VMed = 12.5

	gravado, err := GravarINI(n)
	if err != nil {
		t.Fatalf("GravarINI: %v", err)
	}
	if strings.Contains(gravado, "[Ligacao]") {
		t.Error("[Ligacao] sem idLigacao nao deveria ser gravada (IniWriter.pas:269-270)")
	}
	if !strings.Contains(gravado, "[gMedicao001]") {
		t.Error("gMedicao com vMed>0 deveria ser gravada mesmo com nMed=0 (IniWriter.pas:363-364)")
	}

	// nMed=0 e vMed=0: agora sim a secao nao sai
	n.Det[0].Prod.GMedicao = GMedicao{}
	gravado, err = GravarINI(n)
	if err != nil {
		t.Fatalf("GravarINI: %v", err)
	}
	if strings.Contains(gravado, "[gMedicao001]") {
		t.Error("gMedicao zerada nao deveria ser gravada")
	}
}

func TestRetInfEventoRegistrado(t *testing.T) {
	// conjunto de aceite do ACBrNFAgWebServices.pas:1573: [135, 136, 155]
	for _, cStat := range []int{135, 136, 155} {
		r := &RetInfEvento{CStat: cStat}
		if !r.Registrado() {
			t.Errorf("cStat %d deveria ser evento registrado", cStat)
		}
	}
	for _, cStat := range []int{0, 100, 128, 573} {
		r := &RetInfEvento{CStat: cStat}
		if r.Registrado() {
			t.Errorf("cStat %d nao deveria ser evento registrado", cStat)
		}
	}
}

// ---------------------------------------------------------------------------
// Chave de acesso e regras de negocio
// ---------------------------------------------------------------------------

func TestMontarChaveAcesso(t *testing.T) {
	n := lerCompleta(t)
	chave, err := MontarChaveAcesso(n)
	if err != nil {
		t.Fatal(err)
	}
	if chave != chaveTeste {
		t.Errorf("chave montada = %q, esperado %q", chave, chaveTeste)
	}
	if !ValidarConcatChave(n) {
		t.Error("ValidarConcatChave deveria confirmar a fixture")
	}
	if err := ValidarChaveAcesso(n); err != nil {
		t.Errorf("ValidarChaveAcesso: %v", err)
	}
}

func TestValidarRegrasNegocio(t *testing.T) {
	n := lerCompleta(t)
	cfg := Configuracoes{VersaoDF: Ve100, Ambiente: pcn.TaHomologacao, UF: "SP", CodigoUF: 35}
	if rejeicoes := ValidarRegrasNegocio(n, cfg); len(rejeicoes) != 0 {
		t.Errorf("fixture valida nao deveria ter rejeicoes: %v", rejeicoes)
	}

	// ambiente divergente dispara a regra 252
	cfg.Ambiente = pcn.TaProducao
	if rejeicoes := ValidarRegrasNegocio(n, cfg); len(rejeicoes) == 0 {
		t.Error("ambiente divergente deveria rejeitar (regra 252)")
	}
}

func TestCStat_Classificacao(t *testing.T) {
	// Tabela de TACBrNFAg (ACBrNFAg.pas:228-253), identica a da NFGas.
	casos := []struct {
		cStat                             int
		confirmada, processada, cancelada bool
	}{
		{100, true, true, false},
		{150, true, true, false},
		{101, false, false, true},
		{135, false, false, true},
		{151, false, false, true},
		{155, false, false, true},
		{104, false, false, false},
		{999, false, false, false},
	}
	for _, c := range casos {
		if got := CStatConfirmada(c.cStat); got != c.confirmada {
			t.Errorf("CStatConfirmada(%d) = %v", c.cStat, got)
		}
		if got := CStatProcessado(c.cStat); got != c.processada {
			t.Errorf("CStatProcessado(%d) = %v", c.cStat, got)
		}
		if got := CStatCancelada(c.cStat); got != c.cancelada {
			t.Errorf("CStatCancelada(%d) = %v", c.cStat, got)
		}
	}
}

func TestIdentificarSchema(t *testing.T) {
	casos := map[string]SchemaNFAg{
		`<NFAg xmlns="x"><infNFAg/></NFAg>`:          SchNFAg,
		`<nfagProc versao="1.00"><NFAg/></nfagProc>`: SchNFAg,
		`<consStatServNFAg/>`:                        SchconsStatServNFAg,
		`<retConsSitNFAg/>`:                          SchconsSitNFAg,
		`<retNFAg/>`:                                 SchconsSitNFAg,
		`<eventoNFAg/>`:                              SchEventoNFAg,
		`<evCancNFAg/>`:                              SchCancNFAg,
	}
	for xml, esperado := range casos {
		got, err := IdentificarSchema(xml)
		if err != nil || got != esperado {
			t.Errorf("IdentificarSchema(%q) = %v, %v (esperado %v)", xml, got, err, esperado)
		}
	}
}

// ---------------------------------------------------------------------------
// Componente
// ---------------------------------------------------------------------------

func TestComponente_CargaEContratos(t *testing.T) {
	c := NovoComponente()
	if c.Configuracoes.Ambiente != pcn.TaHomologacao {
		t.Errorf("default do ambiente deveria ser homologacao: %v", c.Configuracoes.Ambiente)
	}
	if err := c.CarregarString(string(carregarFixture(t, "nfag_completa.xml"))); err != nil {
		t.Fatalf("CarregarString: %v", err)
	}
	if len(c.NotasFiscais) != 1 || c.NotasFiscais[0].ChaveAcesso() != chaveTeste {
		t.Errorf("notas carregadas: %d", len(c.NotasFiscais))
	}

	// transmissao sem certificado falha cedo, sem rede
	if _, err := c.Consultar(testContext(), chaveTeste); !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Errorf("Consultar sem certificado = %v", err)
	}
	// MA e PA apontam para o SVAN, sem URL publicada (lacuna do ACBr)
	if _, err := c.URLConsultaNFAg(21); !errors.Is(err, ErrSemURL) {
		t.Errorf("URLConsultaNFAg(MA) deveria dar ErrSemURL, veio %v", err)
	}
	if url, err := c.URLConsultaNFAg(35); err != nil ||
		url != "https://dfe-portal.svrs.rs.gov.br/nfag/Consulta" {
		t.Errorf("URLConsultaNFAg(SP) = %q, %v", url, err)
	}
}
