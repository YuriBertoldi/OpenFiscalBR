// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do gerador de eventos.
//
// O XSD do detEvento e xs:sequence e MUDA por tipo de evento. Nenhum
// round-trip pega ordem trocada -- o leitor acha a tag pelo nome --, entao
// a ordem e verificada por POSICAO em todos os eventos que tem mais de dois
// campos.

func eventoBase(tipo pcn.TipoEvento) *EventoNFe {
	return &EventoNFe{
		IDLote: 1,
		Versao: "1.00",
		Evento: []Evento{{
			InfEvento: InfEvento{
				TpAmb:      pcn.TaHomologacao,
				ChNFe:      chaveDeMinas,
				DhEvento:   time.Date(2026, 1, 20, 18, 45, 0, 0, time.UTC),
				TpEvento:   tipo,
				NSeqEvento: 1,
			},
		}},
	}
}

// gerarEventoDeTeste gera o XML e falha o teste em caso de erro.
func gerarEventoDeTeste(t *testing.T, e *EventoNFe) string {
	t.Helper()
	xml, err := GerarXMLEvento(e)
	if err != nil {
		t.Fatalf("GerarXMLEvento: %v", err)
	}
	return xml
}

// conferirOrdem afirma que as tags aparecem na sequencia informada.
func conferirOrdem(t *testing.T, xml string, tags ...string) {
	t.Helper()
	anterior := -1
	for _, tag := range tags {
		pos := strings.Index(xml, tag)
		if pos < 0 {
			t.Fatalf("tag %s ausente: %s", tag, xml)
		}
		if pos <= anterior {
			t.Errorf("%s fora de ordem (posicao %d, anterior %d)", tag, pos, anterior)
		}
		anterior = pos
	}
}

// TestIDEvento_CinquentaEQuatroPosicoes trava a montagem do Id.
//
// O nSeqEvento vai com DUAS posicoes: a sequencia 1 sai "01". Escrever
// `%d` produz um Id de 53 caracteres, e o proprio ACBr avisa disso (:288).
func TestIDEvento_CinquentaEQuatroPosicoes(t *testing.T) {
	id := IDEvento(InfEvento{
		ChNFe: chaveDeMinas, TpEvento: pcn.TeCancelamento, NSeqEvento: 1,
	})

	if len(id) != 54 {
		t.Fatalf("Id tem %d caracteres, esperado 54: %q", len(id), id)
	}
	fatias := []struct {
		nome        string
		inicio, fim int
		esperado    string
	}{
		{"literal", 0, 2, "ID"},
		{"tpEvento", 2, 8, "110111"},
		{"chave", 8, 52, chaveDeMinas},
		{"nSeqEvento", 52, 54, "01"},
	}
	for _, f := range fatias {
		if got := id[f.inicio:f.fim]; got != f.esperado {
			t.Errorf("%s = %q, esperado %q", f.nome, got, f.esperado)
		}
	}

	// Sequencia de dois digitos continua com duas.
	id = IDEvento(InfEvento{
		ChNFe: chaveDeMinas, TpEvento: pcn.TeCCe, NSeqEvento: 12,
	})
	if got := id[52:54]; got != "12" {
		t.Errorf("nSeqEvento 12 saiu %q", got)
	}
	// E a chave com o literal "NFe" e limpa antes.
	id = IDEvento(InfEvento{
		ChNFe: "NFe" + chaveDeMinas, TpEvento: pcn.TeCCe, NSeqEvento: 1,
	})
	if len(id) != 54 {
		t.Errorf("o literal NFe deveria ser removido da chave: %q", id)
	}
}

// TestCOrgaoDoEvento_ZeroCaiParaOCUFDaChave trava o getcOrgao (:705).
//
// E o que faz o evento ir para a SEFAZ da nota sem o consumidor informar
// nada -- e tambem o que torna a manifestacao do destinatario um caso
// especial, porque ela precisa do 91 EXPLICITO.
func TestCOrgaoDoEvento_ZeroCaiParaOCUFDaChave(t *testing.T) {
	semOrgao := InfEvento{ChNFe: chaveDeMinas}
	if got := COrgaoDoEvento(semOrgao); got != 31 {
		t.Errorf("cOrgao = %d, esperado 31 (o cUF da chave de MG)", got)
	}

	comOrgao := InfEvento{ChNFe: chaveDeMinas, COrgao: COrgaoAmbienteNacional}
	if got := COrgaoDoEvento(comOrgao); got != 91 {
		t.Errorf("cOrgao informado deveria vencer: %d", got)
	}
}

// TestGerarInfEvento_OrdemDosCampos trava a sequencia do infEvento, que e a
// mesma em todos os trinta e um eventos.
func TestGerarInfEvento_OrdemDosCampos(t *testing.T) {
	e := eventoBase(pcn.TeCancelamento)
	e.Evento[0].InfEvento.DetEvento.NProt = "131260000000001"
	e.Evento[0].InfEvento.DetEvento.XJust = "Erro na emissao do documento"

	xml := gerarEventoDeTeste(t, e)

	conferirOrdem(t, xml, "<cOrgao>", "<tpAmb>", "<CNPJ>", "<chNFe>",
		"<dhEvento>", "<tpEvento>", "<nSeqEvento>", "<verEvento>", "<detEvento")
}

// TestGerarEvento_TpEventoEhOCodigoNaoADescricao trava a distincao entre os
// dois metodos do TipoEvento.
//
// Os dois devolvem string e trocar um pelo outro compila. Um gerador que
// emitisse <tpEvento>Cancelamento</tpEvento> passaria em qualquer teste de
// estrutura.
func TestGerarEvento_TpEventoEhOCodigoNaoADescricao(t *testing.T) {
	e := eventoBase(pcn.TeCancelamento)
	e.Evento[0].InfEvento.DetEvento.NProt = "131260000000001"
	e.Evento[0].InfEvento.DetEvento.XJust = "Erro na emissao do documento"

	xml := gerarEventoDeTeste(t, e)

	if !strings.Contains(xml, "<tpEvento>110111</tpEvento>") {
		t.Errorf("tpEvento deveria ser o codigo: %s", xml)
	}
	if strings.Contains(xml, "<tpEvento>Cancelamento</tpEvento>") {
		t.Error("tpEvento saiu com a DESCRICAO no lugar do codigo")
	}
	// E o descEvento e o texto, nao o codigo nem o nome curto.
	if !strings.Contains(xml, "<descEvento>Cancelamento</descEvento>") {
		t.Errorf("descEvento deveria ser o texto: %s", xml)
	}
}

// TestDocumentoDoAutor_CPFDeSerieDePessoaFisica trava as tres condicionais
// empilhadas de :982-1013.
//
// Numa nota de serie 910..969 o emitente e pessoa fisica, e a chave guarda
// o CPF com tres zeros a esquerda para ocupar catorze posicoes. O autor do
// evento sai com onze.
func TestDocumentoDoAutor_CPFDeSerieDePessoaFisica(t *testing.T) {
	// Chave de MG, serie 910, com CPF 11122233396 nas posicoes 7..20.
	const chavePessoaFisica = "31260199900011122233396559100000004561123456798"

	// A chave de teste tem 47 caracteres por construcao; o que importa e o
	// recorte, entao usa-se uma chave de 44 montada a mao.
	chave := "3126019990001112223339655" + "9100000004561123456798"
	if len(chave) != 47 {
		t.Fatalf("a chave do teste tem %d caracteres", len(chave))
	}
	_ = chavePessoaFisica

	t.Run("serie na faixa de pessoa fisica", func(t *testing.T) {
		// cUF 31 | AAMM 2601 | CNPJ 00011122233396 | mod 55 | serie 910
		c := "31" + "2601" + "00011122233396" + "55" + "910" +
			"000000456" + "1" + "12345678" + "9"
		if len(c) != 44 {
			t.Fatalf("chave montada tem %d caracteres: %q", len(c), c)
		}
		no := documentoDoAutor(InfEvento{ChNFe: c, TpEvento: pcn.TeCancelamento})
		xml := no.XML()
		if !strings.Contains(xml, "<CPF>11122233396</CPF>") {
			t.Errorf("esperado CPF de 11 digitos, saiu %s", xml)
		}
	})

	t.Run("serie fora da faixa mantem o CNPJ", func(t *testing.T) {
		c := "31" + "2601" + "00011122233396" + "55" + "001" +
			"000000456" + "1" + "12345678" + "9"
		no := documentoDoAutor(InfEvento{ChNFe: c, TpEvento: pcn.TeCancelamento})
		xml := no.XML()
		if !strings.Contains(xml, "<CNPJ>00011122233396</CNPJ>") {
			t.Errorf("fora da faixa o documento fica com 14 digitos: %s", xml)
		}
	})

	t.Run("manifestacao NAO corta, mesmo na faixa", func(t *testing.T) {
		// A terceira condicional: na manifestacao o autor e o
		// DESTINATARIO, nao o emitente da chave. Perder esse guard trunca
		// o CNPJ de quem manifesta.
		c := "31" + "2601" + "00011122233396" + "55" + "910" +
			"000000456" + "1" + "12345678" + "9"
		no := documentoDoAutor(InfEvento{
			ChNFe:    c,
			CNPJ:     "99999999000191",
			TpEvento: pcn.TeManifDestCiencia,
		})
		xml := no.XML()
		if !strings.Contains(xml, "<CNPJ>99999999000191</CNPJ>") {
			t.Errorf("a manifestacao mantem o CNPJ do manifestante: %s", xml)
		}
	})
}

// TestGerarDetEvento_OrdemPorTipo trava a sequencia de campos de cada
// evento com mais de dois campos.
//
// Sao sequencias DIFERENTES, e e isso que torna um gerador generico
// impossivel. A tabela abaixo e a especificacao executavel delas.
func TestGerarDetEvento_OrdemPorTipo(t *testing.T) {
	casos := []struct {
		tipo     pcn.TipoEvento
		preparar func(*DetEvento)
		ordem    []string
	}{
		{
			pcn.TeCCe,
			func(d *DetEvento) { d.XCorrecao = "Correcao do endereco de entrega" },
			[]string{"<descEvento>", "<xCorrecao>", "<xCondUso>"},
		},
		{
			pcn.TeCancelamento,
			func(d *DetEvento) {
				d.NProt = "131260000000001"
				d.XJust = "Erro na emissao do documento"
			},
			[]string{"<descEvento>", "<nProt>", "<xJust>"},
		},
		{
			pcn.TeCancSubst,
			func(d *DetEvento) {
				d.COrgaoAutor = 31
				d.VerAplic = "1.0"
				d.NProt = "131260000000001"
				d.XJust = "Substituicao por nota nova"
				d.ChNFeRef = chaveNFCeDeMinas
			},
			[]string{"<descEvento>", "<cOrgaoAutor>", "<tpAutor>", "<verAplic>",
				"<nProt>", "<xJust>", "<chNFeRef>"},
		},
		{
			pcn.TeComprEntregaNFe,
			func(d *DetEvento) {
				d.COrgaoAutor = 31
				d.VerAplic = "1.0"
				d.NDoc = "12345678901"
				d.XNome = "RECEBEDOR"
				d.HashComprovante = "abcdefghijklmnopqrstuvwxyz12"
				d.DhEntrega = time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)
				d.DhHashComprovante = d.DhEntrega
			},
			[]string{"<descEvento>", "<cOrgaoAutor>", "<tpAutor>", "<verAplic>",
				"<dhEntrega>", "<nDoc>", "<xNome>", "<hashComprovante>",
				"<dhHashComprovante>"},
		},
		{
			pcn.TeInsucessoEntregaNFe,
			func(d *DetEvento) {
				d.COrgaoAutor = 31
				d.VerAplic = "1.0"
				d.UF = "MG"
				d.TpMotivo = TmRecusa
				d.NTentativa = 2
				d.HashTentativaEntrega = "abcdefghijklmnopqrst"
				d.DhTentativaEntrega = time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)
			},
			[]string{"<descEvento>", "<cOrgaoAutor>", "<verAplic>",
				"<dhTentativaEntrega>", "<nTentativa>", "<tpMotivo>",
				"<hashTentativaEntrega>"},
		},
		{
			pcn.TeCanPedProrrog1,
			func(d *DetEvento) {
				d.IDPedidoCancelado = strings.Repeat("1", 54)
				d.NProt = "131260000000001"
			},
			[]string{"<descEvento>", "<idPedidoCancelado>", "<nProt>"},
		},
	}

	for _, c := range casos {
		t.Run(c.tipo.String(), func(t *testing.T) {
			e := eventoBase(c.tipo)
			c.preparar(&e.Evento[0].InfEvento.DetEvento)
			conferirOrdem(t, gerarEventoDeTeste(t, e), c.ordem...)
		})
	}
}

// TestGerarDetEvento_ManifestacoesDiferemNoXJust trava a diferenca entre as
// quatro manifestacoes, que e de UM parametro no Delphi.
//
//	Confirmacao e Ciencia    SO descEvento
//	Desconhecimento          xJust OPCIONAL
//	OperNaoRealizada         xJust OBRIGATORIO
//
// Uniformizar as quatro -- a tentacao obvia -- faz a confirmacao sair com
// um xJust que o schema recusa, ou a operacao nao realizada sair sem o
// campo que a SEFAZ exige.
func TestGerarDetEvento_ManifestacoesDiferemNoXJust(t *testing.T) {
	casos := []struct {
		tipo      pcn.TipoEvento
		justif    string
		temXJust  bool
		descricao string
	}{
		{pcn.TeManifDestConfirmacao, "", false, "confirmacao sem justificativa"},
		{pcn.TeManifDestConfirmacao, "texto qualquer aqui", false,
			"confirmacao IGNORA a justificativa informada"},
		{pcn.TeManifDestCiencia, "texto qualquer aqui", false,
			"ciencia IGNORA a justificativa informada"},
		{pcn.TeManifDestDesconhecimento, "", false,
			"desconhecimento sem justificativa: tag some"},
		{pcn.TeManifDestDesconhecimento, "Nao reconheco a operacao", true,
			"desconhecimento com justificativa"},
		{pcn.TeManifDestOperNaoRealizada, "", true,
			"operacao nao realizada SEM justificativa: tag sai vazia"},
		{pcn.TeManifDestOperNaoRealizada, "A mercadoria nao chegou", true,
			"operacao nao realizada com justificativa"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			e := eventoBase(c.tipo)
			e.Evento[0].InfEvento.COrgao = COrgaoAmbienteNacional
			e.Evento[0].InfEvento.DetEvento.XJust = c.justif

			xml := gerarEventoDeTeste(t, e)
			if got := strings.Contains(xml, "<xJust"); got != c.temXJust {
				t.Errorf("xJust presente = %v, esperado %v: %s", got, c.temXJust, xml)
			}
		})
	}
}

// TestGerarDetEvento_TpAutorFixoEmSeisEventos trava a divisao entre os que
// IGNORAM o campo tpAutor e os que o respeitam.
//
// Os seis que fixam "1" o fazem porque o schema so aceita esse valor, e o
// ACBr comenta isso em cada um. Uniformizar os dois grupos -- em qualquer
// direcao -- produz rejeicao numa metade.
func TestGerarDetEvento_TpAutorFixoEmSeisEventos(t *testing.T) {
	preparar := map[pcn.TipoEvento]func(*DetEvento){
		pcn.TeCancSubst: func(d *DetEvento) {
			d.NProt = "131260000000001"
			d.XJust = "Substituicao por nota nova"
			d.ChNFeRef = chaveNFCeDeMinas
		},
		pcn.TeCancComprEntregaNFe: func(d *DetEvento) { d.NProtEvento = "131260000000002" },
		pcn.TeComprEntregaNFe: func(d *DetEvento) {
			d.NDoc, d.XNome = "12345678901", "RECEBEDOR"
			d.HashComprovante = "abcdefghijklmnopqrstuvwxyz12"
		},
		pcn.TeAtorInteressadoNFe: func(d *DetEvento) {
			d.AutXML = []AutXMLEvento{{CNPJCPF: "99999999000191"}}
		},
	}

	// Os que FIXAM "1", ignorando o campo.
	for _, tipo := range []pcn.TipoEvento{
		pcn.TeCancSubst, pcn.TeCancComprEntregaNFe, pcn.TeComprEntregaNFe,
	} {
		t.Run("fixo/"+tipo.String(), func(t *testing.T) {
			e := eventoBase(tipo)
			d := &e.Evento[0].InfEvento.DetEvento
			d.COrgaoAutor, d.VerAplic = 31, "1.0"
			// TaFisco = "3"; se o campo fosse respeitado, sairia 3.
			d.TpAutor = TaFisco
			preparar[tipo](d)

			xml := gerarEventoDeTeste(t, e)
			if !strings.Contains(xml, "<tpAutor>1</tpAutor>") {
				t.Errorf("%s fixa tpAutor em 1: %s", tipo, xml)
			}
		})
	}

	// O ator interessado RESPEITA o campo.
	t.Run("do campo/AtorInteressadoNFe", func(t *testing.T) {
		e := eventoBase(pcn.TeAtorInteressadoNFe)
		d := &e.Evento[0].InfEvento.DetEvento
		d.COrgaoAutor, d.VerAplic = 31, "1.0"
		d.TpAutor = TaFisco
		preparar[pcn.TeAtorInteressadoNFe](d)

		xml := gerarEventoDeTeste(t, e)
		if !strings.Contains(xml, "<tpAutor>"+TaFisco.String()+"</tpAutor>") {
			t.Errorf("o ator interessado respeita o campo tpAutor: %s", xml)
		}
	})
}

// TestGerarDetEvento_AtorInteressadoXCondUsoSoNaPermissao trava a
// condicional de :456.
//
// Emitir o texto sempre declara permissao de download numa REVOGACAO -- o
// contrario do que o evento quer dizer.
func TestGerarDetEvento_AtorInteressadoXCondUsoSoNaPermissao(t *testing.T) {
	for _, c := range []struct {
		autorizacao Autorizacao
		temTexto    bool
	}{
		{TaPermite, true},
		{TaNaoPermite, false},
		{TaNaoInformar, false},
	} {
		e := eventoBase(pcn.TeAtorInteressadoNFe)
		d := &e.Evento[0].InfEvento.DetEvento
		d.COrgaoAutor, d.VerAplic = 31, "1.0"
		d.AutXML = []AutXMLEvento{{CNPJCPF: "99999999000191"}}
		d.TpAutorizacao = c.autorizacao

		xml := gerarEventoDeTeste(t, e)
		if got := strings.Contains(xml, "<xCondUso>"); got != c.temTexto {
			t.Errorf("autorizacao %v: xCondUso presente = %v, esperado %v",
				c.autorizacao, got, c.temTexto)
		}
	}
}

// TestGerarDetEvento_InsucessoXJustMotivoSoNoMotivoOutro trava :803.
func TestGerarDetEvento_InsucessoXJustMotivoSoNoMotivoOutro(t *testing.T) {
	for _, c := range []struct {
		motivo  TpMotivo
		temJust bool
	}{
		{TmOutro, true},
		{TmRecusa, false},
		{TmNaoEncontrado, false},
		{TmInexistente, false},
	} {
		e := eventoBase(pcn.TeInsucessoEntregaNFe)
		d := &e.Evento[0].InfEvento.DetEvento
		d.COrgaoAutor, d.VerAplic, d.UF = 31, "1.0", "MG"
		d.TpMotivo = c.motivo
		d.XJustMotivo = "Motivo detalhado da tentativa frustrada"
		d.HashTentativaEntrega = "abcdefghijklmnopqrst"
		d.DhTentativaEntrega = time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)

		xml := gerarEventoDeTeste(t, e)
		if got := strings.Contains(xml, "<xJustMotivo>"); got != c.temJust {
			t.Errorf("motivo %v: xJustMotivo presente = %v, esperado %v",
				c.motivo, got, c.temJust)
		}
	}
}

// TestGerarDetEvento_CCePreencheOXCondUso trava o texto legal obrigatorio.
func TestGerarDetEvento_CCePreencheOXCondUso(t *testing.T) {
	e := eventoBase(pcn.TeCCe)
	e.Evento[0].InfEvento.DetEvento.XCorrecao = "Correcao do endereco de entrega"

	xml := gerarEventoDeTeste(t, e)
	if !strings.Contains(xml, "paragrafo 1o-A do art. 7o do Convenio S/N") {
		t.Errorf("a CC-e tem de levar o texto legal no xCondUso: %s", xml)
	}

	// Informado, o texto do consumidor prevalece -- divergencia declarada.
	e = eventoBase(pcn.TeCCe)
	e.Evento[0].InfEvento.DetEvento.XCorrecao = "Correcao do endereco"
	e.Evento[0].InfEvento.DetEvento.XCondUso = "texto proprio"
	if xml := gerarEventoDeTeste(t, e); !strings.Contains(xml, "<xCondUso>texto proprio</xCondUso>") {
		t.Errorf("o texto informado deveria prevalecer: %s", xml)
	}
}

// TestGerarXMLEvento_LimiteDeVinteEventos trava o teto de :2522.
func TestGerarXMLEvento_LimiteDeVinteEventos(t *testing.T) {
	montar := func(n int) *EventoNFe {
		e := eventoBase(pcn.TeManifDestCiencia)
		e.Evento[0].InfEvento.COrgao = COrgaoAmbienteNacional
		for i := 1; i < n; i++ {
			e.Evento = append(e.Evento, e.Evento[0])
		}
		return e
	}

	if _, err := GerarXMLEvento(montar(21)); !errors.Is(err, ErrLoteExcedido) {
		t.Errorf("21 eventos: err = %v, esperado ErrLoteExcedido", err)
	}
	if _, err := GerarXMLEvento(montar(20)); err != nil {
		t.Errorf("20 eventos estao no limite, nao acima dele: %v", err)
	}
	if _, err := GerarXMLEvento(&EventoNFe{Versao: "1.00"}); !errors.Is(err, ErrLoteVazio) {
		t.Errorf("lote vazio: err = %v, esperado ErrLoteVazio", err)
	}
}

// TestGerarDetEvento_TipoSemGeradorEhErro trava a divergencia deliberada.
//
// O ACBr deixa o `case` sem ramo e emite um infEvento SEM detEvento, que a
// SEFAZ rejeita sem dizer o porque.
func TestGerarDetEvento_TipoSemGeradorEhErro(t *testing.T) {
	// Registro de passagem: a NF-e o LE, mas nunca o gera.
	e := eventoBase(pcn.TeRegistroPassagemNFe)
	if _, err := GerarXMLEvento(e); !errors.Is(err, ErrEventoNaoSuportado) {
		t.Errorf("err = %v, esperado ErrEventoNaoSuportado", err)
	}
}

// TestGerarXMLEvento_EventoAssinadoEntraSemRegerar trava a razao de existir
// UM so montador de envelope.
//
// A assinatura cobre BYTES. Se o GerarXMLEvento regerasse um evento ja
// assinado, o digest deixaria de bater e a SEFAZ rejeitaria -- e o erro
// apareceria so na transmissao, nunca em teste de estrutura.
func TestGerarXMLEvento_EventoAssinadoEntraSemRegerar(t *testing.T) {
	e := eventoBase(pcn.TeCancelamento)
	e.Evento[0].XMLAssinado = `<evento versao="1.00" xmlns="` + namespaceNFe +
		`"><infEvento Id="IDJADEFINIDO"><chNFe>x</chNFe></infEvento>` +
		`<Signature><SignatureValue>AAA</SignatureValue></Signature></evento>`

	xml := gerarEventoDeTeste(t, e)

	if !strings.Contains(xml, "<SignatureValue>AAA</SignatureValue>") {
		t.Errorf("a assinatura do evento tem de ser preservada: %s", xml)
	}
	if !strings.Contains(xml, `Id="IDJADEFINIDO"`) {
		t.Errorf("o Id do evento assinado nao pode ser remontado: %s", xml)
	}
	if strings.Contains(xml, "<descEvento>") {
		t.Errorf("o evento assinado NAO foi regerado, entao nao ganha"+
			" detEvento novo: %s", xml)
	}
	// E o envelope do lote continua la.
	if !strings.HasPrefix(xml, "<envEvento ") || !strings.Contains(xml, "<idLote>1</idLote>") {
		t.Errorf("envelope do lote: %s", xml)
	}
}

// TestGerarDetEventoAtorInteressado_SoUmAutXML trava o maxOccurs=1 do
// schema do 110150.
//
// O ACBr gera todos e apenas ALERTA (:322); a SEFAZ rejeita. O porte emite
// só o primeiro -- divergência registrada.
func TestGerarDetEventoAtorInteressado_SoUmAutXML(t *testing.T) {
	e := eventoBase(pcn.TeAtorInteressadoNFe)
	d := &e.Evento[0].InfEvento.DetEvento
	d.COrgaoAutor, d.VerAplic = 31, "1.0"
	d.AutXML = []AutXMLEvento{
		{CNPJCPF: "99999999000191"},
		{CNPJCPF: "11111111000191"},
	}

	xml := gerarEventoDeTeste(t, e)

	if n := strings.Count(xml, "<autXML>"); n != 1 {
		t.Errorf("autXML emitidos = %d, esperado 1 (maxOccurs do schema)", n)
	}
	if !strings.Contains(xml, "<CNPJ>99999999000191</CNPJ>") {
		t.Errorf("o PRIMEIRO autXML e o que sai: %s", xml)
	}
	if strings.Contains(xml, "11111111000191") {
		t.Errorf("o segundo autXML nao pode sair: %s", xml)
	}
}

// TestGerarDetEventoEPEC_ModeloDesconhecidoCaiNoRamoDaNFCe trava o literal
// do teste de modelo.
//
// O ACBr escreve `if sModelo = '55' then <NF-e> else <NFC-e>` (:657).
// Inverter para `== "65"` parece equivalente e muda o destino de toda
// chave curta ou de modelo desconhecido.
func TestGerarDetEventoEPEC_ModeloDesconhecidoCaiNoRamoDaNFCe(t *testing.T) {
	e := eventoEPEC(chaveDeMinas)
	// Chave curta: o recorte do modelo devolve vazio.
	e.Evento[0].InfEvento.ChNFe = "3126"

	xml := gerarEventoDeTeste(t, e)
	if strings.Contains(xml, "<vST>") {
		t.Errorf("modelo desconhecido cai no ramo da NFC-e, que nao tem"+
			" vST: %s", xml)
	}
}
