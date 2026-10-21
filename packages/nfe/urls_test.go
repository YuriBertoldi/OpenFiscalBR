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

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Testes do QR-Code e das URLs publicas.
//
// A tabela de URLs em si e conferida contra o .ini por
// ferramentas/conferir-urls-dfe.py -- repetir aqui os 135 valores so
// provaria que sei copiar. O que estes testes travam e o que o script nao
// alcanca: a MONTAGEM do QR-Code, que e onde moram as condicionais.

// notaParaQRCode monta uma NFC-e com o minimo que o QR-Code consome.
func notaParaQRCode(t *testing.T) *NFe {
	t.Helper()
	n := lerFixture(t, "nfce_proc_autorizada.xml")
	n.Ide.CUF = 35 // SP
	n.Ide.TpAmb = pcn.TaProducao
	n.Ide.Modelo = 65
	n.Signature.DigestValue = "abc123"
	return n
}

// emContingenciaOffline poe a nota em contingencia offline DE VERDADE:
// mexe no tpEmis e REGENERA a chave.
//
// Nao basta mudar o grupo ide -- o QR-Code le o tpEmis da 35a posicao da
// CHAVE, nao do ide, e e proposital: quem confere o documento tem a chave
// em maos, nao o XML.
func emContingenciaOffline(t *testing.T, n *NFe) *NFe {
	t.Helper()
	n.Ide.TpEmis = pcn.TeOffLine
	chave, err := MontarChaveAcesso(n)
	if err != nil {
		t.Fatal(err)
	}
	n.InfNFe.ID = "NFe" + chave
	return n
}

// TestGerarQRCode_VersaoVemDoAjuste trava a primeira decisao do
// GetURLQRCode: a versao pedida na configuracao passa por
// AjustarVersaoQrCode, que a rebaixa conforme a versao do LEIAUTE.
//
// Uma nota 3.10 nunca gera QR-Code versao 3, por mais que a configuracao
// peca -- e o formato muda por inteiro entre as versoes.
func TestGerarQRCode_VersaoVemDoAjuste(t *testing.T) {
	casos := []struct {
		leiaute   float64
		pedida    VersaoQrCode
		marcador  string
		descricao string
	}{
		{4.00, Veqr300, "|3|", "4.00 com versao 3 pedida"},
		{4.00, Veqr200, "|2|", "4.00 com versao 2 pedida"},
		{4.00, Veqr100, "chNFe=", "4.00 com versao 1 pedida"},
		// Abaixo da 4.00 o AjustarVersaoQrCode devolve Veqr000 ("sem
		// QR-Code"), e o case do ACBr cai no ELSE -- que gera a versao 1.
		// Replicado: o ACBr nao deixa de gerar, ele rebaixa ao formato
		// antigo (ACBrNFe.pas:642-648).
		{3.10, Veqr300, "chNFe=", "3.10 com versao 3 pedida cai na versao 1"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n := notaParaQRCode(t)
			n.InfNFe.Versao = c.leiaute

			qr, err := GerarQRCode(n, Configuracoes{
				VersaoQRCode: c.pedida,
				CSC:          "CSCDETESTE",
				IDCSC:        "1",
			}, assinaturaFalsa)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(qr, c.marcador) {
				t.Errorf("QR-Code = %q; esperado conter %q", qr, c.marcador)
			}
		})
	}
}

// assinaturaFalsa substitui o certificado nos testes. Devolve um valor
// constante: o que se testa aqui e QUANDO ela e chamada, nao o que ela
// calcula -- isso e papel do packages/dfe.
func assinaturaFalsa(texto string) (string, error) {
	return "ASSINATURA", nil
}

// TestGerarQRCode_V3SoAcrescentaDadosEmContingencia trava a condicional que
// governa a versao 3: fora da contingencia offline, o QR-Code tem apenas
// chave, versao e ambiente.
//
// Quem decide e o tpEmis lido da CHAVE, nao o do grupo ide -- e a chave que
// o consumidor tem em maos ao conferir o documento.
func TestGerarQRCode_V3SoAcrescentaDadosEmContingencia(t *testing.T) {
	cfg := Configuracoes{VersaoQRCode: Veqr300}

	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00
	n.Ide.TpEmis = pcn.TeNormal

	qr, err := GerarQRCode(n, cfg, assinaturaFalsa)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(qr, "ASSINATURA") {
		t.Error("o QR-Code versao 3 fora da contingencia nao deve ser assinado")
	}
	if partes := strings.Count(qr, "|"); partes != 2 {
		t.Errorf("QR-Code = %q; fora da contingencia ele tem so chave|3|tpAmb", qr)
	}

	// Em contingencia offline (tpEmis 9 na chave) entram mais cinco campos
	// e a assinatura.
	offline := notaParaQRCode(t)
	offline.InfNFe.Versao = 4.00
	emContingenciaOffline(t, offline)

	qr, err = GerarQRCode(offline, cfg, assinaturaFalsa)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(qr, "|ASSINATURA") {
		t.Errorf("QR-Code offline = %q; deveria terminar com a assinatura", qr)
	}
}

// TestGerarQRCode_V3OfflineSemCertificadoEhErro trava o unico caso em que o
// package exige o certificado.
func TestGerarQRCode_V3OfflineSemCertificadoEhErro(t *testing.T) {
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00
	emContingenciaOffline(t, n)

	_, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr300}, nil)
	if !errors.Is(err, ErrCertificadoObrigatorio) {
		t.Errorf("erro = %v; esperado ErrCertificadoObrigatorio", err)
	}

	// Fora da contingencia, a mesma nota dispensa o certificado.
	n.Ide.TpEmis = pcn.TeNormal
	chaveNormal, err := MontarChaveAcesso(n)
	if err != nil {
		t.Fatal(err)
	}
	n.InfNFe.ID = "NFe" + chaveNormal
	if _, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr300}, nil); err != nil {
		t.Errorf("sem contingencia o certificado nao e necessario, mas deu %v", err)
	}
}

// camposDoQRCodeV3 recorta o conteudo depois do "?p=" e o quebra nos
// campos separados por barra vertical.
//
// Conferir o TipoDest por POSICAO, e nao com um Contains, e o que torna
// este teste capaz de falhar: no formato da versao 3 o proprio separador da
// versao e "|3|", e o tpAmb da nota de producao e "|1|" -- um Contains por
// esses textos casa com o campo errado e passa sempre.
//
// Os campos sao: chave, versao, tpAmb, dia, vNF, TipoDest, cDest e a
// assinatura.
func camposDoQRCodeV3(t *testing.T, qr string) []string {
	t.Helper()
	i := strings.Index(qr, "?p=")
	if i < 0 {
		t.Fatalf("QR-Code sem o separador ?p=: %q", qr)
	}
	campos := strings.Split(qr[i+3:], "|")
	if len(campos) != 8 {
		t.Fatalf("QR-Code versao 3 offline tem 8 campos, veio %d: %q", len(campos), qr)
	}
	return campos
}

// TestGerarQRCode_TipoDestinatarioTemPrecedencia trava a ordem do
// TipoDest na versao 3 (ACBrNFe.pas:760-769).
//
// O ACBr marca 3 para estrangeiro e DEPOIS sobrescreve com 2 ou 1 se houver
// CNPJ/CPF de tamanho valido. Um documento com os dois sai como CPF, nao
// como estrangeiro. A ordem inversa pareceria mais natural e esta errada.
func TestGerarQRCode_TipoDestinatarioTemPrecedencia(t *testing.T) {
	const posTipoDest, posDest = 5, 6

	casos := []struct {
		estrangeiro string
		cnpjCPF     string
		tipo        string
		dest        string
		descricao   string
	}{
		{"", "11111111111", "2", "11111111111", "CPF"},
		{"", "99999999000272", "1", "99999999000272", "CNPJ"},
		{"PASSAPORTE123", "", "3", "PASSAPORTE123", "so estrangeiro"},
		{"PASSAPORTE123", "11111111111", "2", "PASSAPORTE123",
			"estrangeiro E CPF: o TIPO vem do CPF, mas o DOCUMENTO continua o estrangeiro"},
		{"", "", "", "", "nenhum dos dois"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n := notaParaQRCode(t)
			n.InfNFe.Versao = 4.00
			emContingenciaOffline(t, n)
			n.Dest.IDEstrangeiro = c.estrangeiro
			n.Dest.CNPJCPF = c.cnpjCPF

			qr, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr300}, assinaturaFalsa)
			if err != nil {
				t.Fatal(err)
			}
			campos := camposDoQRCodeV3(t, qr)

			if campos[posTipoDest] != c.tipo {
				t.Errorf("TipoDest = %q, esperado %q", campos[posTipoDest], c.tipo)
			}
			if campos[posDest] != c.dest {
				t.Errorf("cDest = %q, esperado %q", campos[posDest], c.dest)
			}
		})
	}
}

// TestGerarQRCode_HexMinusculoEmTresUFs trava a excecao de SP, PR e MS
// (ACBrNFe.pas:574 e :718).
//
// O AsciiToHex do ACBr devolve MAIUSCULO; essas tres UFs exigem minusculo, e
// a SEFAZ confere o hash literalmente -- a caixa errada reprova o documento.
func TestGerarQRCode_HexMinusculoEmTresUFs(t *testing.T) {
	cfg := Configuracoes{VersaoQRCode: Veqr200, CSC: "CSC", IDCSC: "1"}

	for _, cUF := range []int{35, 41, 50} {
		n := notaParaQRCode(t)
		n.InfNFe.Versao = 4.00
		n.Ide.CUF = cUF
		emContingenciaOffline(t, n)
		n.Signature.DigestValue = "Zz"

		qr, err := GerarQRCode(n, cfg, nil)
		if err != nil {
			t.Fatal(err)
		}
		// "Zz" em hexadecimal e 5A7A; minusculo fica 5a7a.
		if !strings.Contains(qr, "5a7a") {
			t.Errorf("cUF %d: o digest deveria sair em hexadecimal MINUSCULO: %q", cUF, qr)
		}
	}

	// Numa UF fora das tres, o mesmo digest sai maiusculo.
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00
	n.Ide.CUF = 33 // RJ
	emContingenciaOffline(t, n)
	n.Signature.DigestValue = "Zz"

	qr, err := GerarQRCode(n, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(qr, "5A7A") {
		t.Errorf("no RJ o digest deveria sair MAIUSCULO: %q", qr)
	}
}

// TestGerarQRCode_IDCSCMudaDeFormatoEntreAsVersoes trava uma diferenca que
// passa despercebida: a versao 1 preenche o identificador do CSC com zeros
// ate seis posicoes e a versao 2 nao preenche nada.
func TestGerarQRCode_IDCSCMudaDeFormatoEntreAsVersoes(t *testing.T) {
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00

	v1, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr100, CSC: "X", IDCSC: "7"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(v1, "cIdToken=000007") {
		t.Errorf("versao 1: o idCSC deveria sair com seis posicoes: %q", v1)
	}

	v2, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr200, CSC: "X", IDCSC: "7"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(v2, "000007") {
		t.Errorf("versao 2: o idCSC deveria sair sem zeros a esquerda: %q", v2)
	}
	if !strings.Contains(v2, "|7|") {
		t.Errorf("versao 2: o idCSC 7 deveria aparecer cru: %q", v2)
	}
}

// TestGerarQRCode_CSCVazioEhDerivadoDaChave documenta uma REPLICACAO
// deliberada: sem CSC configurado, o ACBr monta um a partir da propria
// chave (ACBrNFe.pas:680).
//
// O codigo resultante nao valida na SEFAZ. O ACBr faz isso para o DANFE de
// teste nao quebrar, e e deterministico -- por isso foi replicado, e por
// isso existe este teste, que impede alguem de "consertar" sem perceber.
func TestGerarQRCode_CSCVazioEhDerivadoDaChave(t *testing.T) {
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00

	comCSC, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr200, CSC: "SEGREDO", IDCSC: "1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	semCSC, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr200, IDCSC: "1"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if comCSC == semCSC {
		t.Error("o hash deveria mudar com o CSC; sem CSC o ACBr deriva um da chave")
	}

	// A derivacao e deterministica: a mesma chave da sempre o mesmo codigo.
	outraVez, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr200, IDCSC: "1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if semCSC != outraVez {
		t.Error("a derivacao do CSC a partir da chave deveria ser deterministica")
	}
}

// TestAcrescentarSufixoP trava o tratamento da URL que ja traz "?p=".
//
// SC e a UF que expoe o caso: a URL dela termina em "?p=" na propria
// tabela, e acrescentar outro produziria "?p=?p=".
func TestGerarQRCode_SufixoPNaoEhDuplicado(t *testing.T) {
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00
	n.Ide.CUF = 42 // SC

	qr, err := GerarQRCode(n, Configuracoes{VersaoQRCode: Veqr200, CSC: "X", IDCSC: "1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(qr, "?p=") != 1 {
		t.Errorf("a URL de SC ja traz ?p=; o QR-Code nao deve duplica-lo: %q", qr)
	}
}

// TestURLQRCode_ModeloEscolheATabela trava as duas UFs em que a URL do
// modelo 55 difere da do 65.
func TestURLQRCode_ModeloEscolheATabela(t *testing.T) {
	// BA e PR sao as duas; nas outras 25 as tabelas coincidem.
	for _, cUF := range []int{29, 41} { // BA, PR
		nfe, err := URLQRCode(55, cUF, pcn.TaProducao)
		if err != nil {
			t.Fatal(err)
		}
		nfce, err := URLQRCode(65, cUF, pcn.TaProducao)
		if err != nil {
			t.Fatal(err)
		}
		if nfe == nfce {
			t.Errorf("cUF %d: as URLs do modelo 55 e 65 deveriam divergir", cUF)
		}
	}

	// Em SP as duas sao iguais, e a tabela de divergencias nao tem entrada.
	nfe, _ := URLQRCode(55, 35, pcn.TaProducao)
	nfce, _ := URLQRCode(65, 35, pcn.TaProducao)
	if nfe != nfce {
		t.Error("em SP as URLs dos dois modelos deveriam coincidir")
	}
}

// TestURLConsulta_VersaoDoQRCodeEscolheATabela trava o unico servico cuja
// URL muda com a versao do QR-Code.
func TestURLConsulta_VersaoDoQRCodeEscolheATabela(t *testing.T) {
	v1, err := URLConsulta(MoNFCe, 35, pcn.TaProducao, Veqr100)
	if err != nil {
		t.Fatal(err)
	}
	v2, err := URLConsulta(MoNFCe, 35, pcn.TaProducao, Veqr200)
	if err != nil {
		t.Fatal(err)
	}
	v3, err := URLConsulta(MoNFCe, 35, pcn.TaProducao, Veqr300)
	if err != nil {
		t.Fatal(err)
	}
	if v2 != v3 {
		t.Errorf("as versoes 2 e 3 compartilham a URL de consulta: %q x %q", v2, v3)
	}

	// Numa UF em que as duas tabelas divergem, a versao 1 tem URL propria.
	a, _ := URLConsulta(MoNFCe, 33, pcn.TaProducao, Veqr100) // RJ
	b, _ := URLConsulta(MoNFCe, 33, pcn.TaProducao, Veqr200)
	if a == b {
		t.Errorf("no RJ a consulta da versao 1 difere das demais, mas veio %q", a)
	}
	_ = v1

	// O modelo 55 usa a outra tabela, e ela nao varia com a versao.
	nfe1, _ := URLConsulta(MoNFe, 33, pcn.TaProducao, Veqr100)
	nfe3, _ := URLConsulta(MoNFe, 33, pcn.TaProducao, Veqr300)
	if nfe1 != nfe3 {
		t.Error("a consulta do modelo 55 nao varia com a versao do QR-Code")
	}
}

// TestURLQRCode_UFDesconhecidaEhErro trava a diferenca deliberada em relacao
// ao ACBr, que devolve string vazia em silencio.
func TestURLQRCode_UFDesconhecidaEhErro(t *testing.T) {
	if _, err := URLQRCode(65, 99, pcn.TaProducao); !errors.Is(err, ErrSemURL) {
		t.Errorf("erro = %v; esperado ErrSemURL", err)
	}
	if _, err := URLConsulta(MoNFCe, 99, pcn.TaProducao, Veqr200); !errors.Is(err, ErrSemURL) {
		t.Errorf("erro = %v; esperado ErrSemURL", err)
	}
}

// TestTipoEmissaoDaChave trava a extracao da 35a posicao.
func TestTipoEmissaoDaChave(t *testing.T) {
	// cUF(2) AAMM(4) CNPJ(14) mod(2) serie(3) nNF(9) = 34 posicoes antes.
	chave := "35" + "2610" + "99999999000272" + "65" + "001" + "000000123" + "9" + "12345678" + "0"
	if got := TipoEmissaoDaChave(chave); got != 9 {
		t.Errorf("TipoEmissaoDaChave = %d, esperado 9", got)
	}
	if got := TipoEmissaoDaChave("NFe" + chave); got != 9 {
		t.Errorf("com o literal NFe: %d, esperado 9", got)
	}
	if got := TipoEmissaoDaChave("123"); got != 0 {
		t.Errorf("chave curta: %d, esperado 0", got)
	}
}

// TestGerarQRCode_V1MontagemCompleta trava a query string INTEIRA da versao
// 1, parametro a parametro e na ordem.
//
// E a montagem mais longa das tres e a unica com dhEmi e vICMS. O hash e
// calculado sobre a string toda, entao trocar dois parametros de lugar muda
// o cHashQRCode e a SEFAZ recusa -- e nenhum teste de presenca percebe.
func TestGerarQRCode_V1MontagemCompleta(t *testing.T) {
	n := notaParaQRCode(t)
	n.InfNFe.Versao = 4.00
	n.Ide.CUF = 33 // RJ: fora das tres UFs que minusculizam o hexadecimal
	n.Dest.CNPJCPF = "11111111111"
	n.Dest.IDEstrangeiro = ""
	n.Signature.DigestValue = "ABC"
	n.Total.ICMSTot.VNF = 1234.56
	n.Total.ICMSTot.VICMS = 78.90

	qr, err := GerarQRCode(n, Configuracoes{
		VersaoQRCode: Veqr100, CSC: "CSC", IDCSC: "7",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	i := strings.Index(qr, "?")
	if i < 0 {
		t.Fatalf("QR-Code sem query string: %q", qr)
	}
	query := qr[i+1:]

	// A ORDEM e o contrato: cada parametro tem de vir onde o ACBr o coloca.
	ordem := []string{
		"chNFe=", "&nVersao=100&", "tpAmb=", "&cDest=", "&dhEmi=",
		"&vNF=", "&vICMS=", "&digVal=", "&cIdToken=", "&cHashQRCode=",
	}
	pos := -1
	for _, marcador := range ordem {
		j := strings.Index(query, marcador)
		if j < 0 {
			t.Errorf("parametro %q ausente em %q", marcador, query)
			continue
		}
		if j < pos {
			t.Errorf("parametro %q saiu fora de ordem; a sequencia e %v",
				marcador, ordem)
		}
		pos = j
	}

	// Conteudo dos campos que o teste consegue prever.
	esperados := []string{
		"&cDest=11111111111",
		"&vNF=1234.56",
		"&vICMS=78.90",
		"&digVal=414243", // "ABC" em hexadecimal MAIUSCULO
		"&cIdToken=000007",
	}
	for _, e := range esperados {
		if !strings.Contains(query, e) {
			t.Errorf("esperado %q em %q", e, query)
		}
	}

	// O hash fecha a string e tem 40 posicoes (SHA-1 em hexadecimal).
	j := strings.Index(query, "&cHashQRCode=")
	if hash := query[j+len("&cHashQRCode="):]; len(hash) != 40 {
		t.Errorf("cHashQRCode tem %d posicoes, esperado 40: %q", len(hash), hash)
	}
}

// TestGerarQRCode_V1OmiteCDestSemDestinatario trava o unico parametro
// CONDICIONAL da versao 1 (ACBrNFe.pas:688).
func TestGerarQRCode_V1OmiteCDestSemDestinatario(t *testing.T) {
	cfg := Configuracoes{VersaoQRCode: Veqr100, CSC: "CSC", IDCSC: "1"}

	com := notaParaQRCode(t)
	com.InfNFe.Versao = 4.00
	com.Dest.CNPJCPF = "11111111111"
	qr, err := GerarQRCode(com, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(qr, "&cDest=") {
		t.Error("com destinatario, o cDest deveria aparecer")
	}

	sem := notaParaQRCode(t)
	sem.InfNFe.Versao = 4.00
	sem.Dest.CNPJCPF = ""
	sem.Dest.IDEstrangeiro = ""
	qr, err = GerarQRCode(sem, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(qr, "&cDest=") {
		t.Errorf("sem destinatario o cDest nao deve sair: %q", qr)
	}
}

// TestGerarQRCode_V1SeparadorDependeDaURL trava a escolha entre "?" e "&"
// (ACBrNFe.pas:696-699).
//
// A URL de SC ja traz "?p=" e por isso recebe "&"; as demais recebem "?".
func TestGerarQRCode_V1SeparadorDependeDaURL(t *testing.T) {
	cfg := Configuracoes{VersaoQRCode: Veqr100, CSC: "CSC", IDCSC: "1"}

	sc := notaParaQRCode(t)
	sc.InfNFe.Versao = 4.00
	sc.Ide.CUF = 42
	qr, err := GerarQRCode(sc, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(qr, "?p=&chNFe=") {
		t.Errorf("a URL de SC ja tem '?', entao o separador e '&': %q", qr)
	}

	sp := notaParaQRCode(t)
	sp.InfNFe.Versao = 4.00
	qr, err = GerarQRCode(sp, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(qr, "qrcode?chNFe=") {
		t.Errorf("a URL de SP nao tem '?', entao o separador e '?': %q", qr)
	}
}
