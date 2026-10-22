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

package nfe

import (
	"fmt"
	"testing"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Testes do formato .ini.
//
// O que estes testes travam NAO e "o campo foi lido": e a CONDICIONAL em
// volta dele. Um leitor de .ini que ignore os aliases de secao recusa
// metade dos arquivos em circulacao e passa em qualquer round-trip; um
// gravador que ignore o `case` por CST produz arquivo que a SEFAZ recusa e
// que a ida e volta nao percebe.

// chaveValidaINI e a chaveDeMinas com o digito verificador CORRIGIDO.
//
// A constante compartilhada termina em 1 e nao fecha o DV -- serve aos
// testes que so a carregam como texto. O gravador do .ini valida a chave
// antes de gerar, como o original faz, entao aqui ela precisa fechar.
const chaveValidaINI = "31260199999999000191550010000004561123456790"

// cfgINI e a configuracao minima usada nos testes de leitura.
func cfgINI() Configuracoes {
	return Configuracoes{
		VersaoDF: Ve400,
		Ambiente: pcn.TaHomologacao,
		TpEmis:   1,
	}
}

// ---------------------------------------------------------------------------
// Aliases de secao e de chave
// ---------------------------------------------------------------------------

// TestLerINI_AceitaNomeCurtoELongoDaSecao trava o alias de secao.
//
// O ACBr escolhe o nome pela SECAO QUE EXISTE no arquivo, nao por
// configuracao. Suportar so [ide] recusa todo .ini gerado pelo proprio
// ACBr, que grava [Identificacao]; suportar so [Identificacao] recusa os
// gerados a partir do leiaute.
func TestLerINI_AceitaNomeCurtoELongoDaSecao(t *testing.T) {
	casos := []struct {
		descricao string
		texto     string
	}{
		{"nome longo", "[Identificacao]\nnatOp=VENDA\nnNF=77\n" +
			"[Emitente]\nxNome=ACME\nUF=MG\n" +
			"[Destinatario]\nxNome=CLIENTE\n"},
		{"nome curto", "[ide]\nnatOp=VENDA\nnNF=77\n" +
			"[emit]\nxNome=ACME\nUF=MG\n" +
			"[dest]\nxNome=CLIENTE\n"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n, err := LerINI(c.texto, cfgINI())
			if err != nil {
				t.Fatal(err)
			}
			if n.Ide.NatOp != "VENDA" || n.Ide.NNF != 77 {
				t.Errorf("ide = %+v", n.Ide)
			}
			if n.Emit.XNome != "ACME" {
				t.Errorf("emit = %+v", n.Emit)
			}
			if n.Dest.XNome != "CLIENTE" {
				t.Errorf("dest = %+v", n.Dest)
			}
		})
	}
}

// TestLerINI_AceitaChaveEmPortuguesPorExtenso trava o alias de CHAVE, que e
// uma cadeia e tem ORDEM.
//
// A forma `ReadString(s,'cNF', ReadString(s,'Codigo',”))` quer dizer: use
// `cNF`; faltando ele, use `Codigo`. Inverter a ordem faz o nome antigo
// vencer o do leiaute num arquivo que traga os dois.
func TestLerINI_AceitaChaveEmPortuguesPorExtenso(t *testing.T) {
	texto := "[ide]\nCodigo=123\nNaturezaOperacao=REMESSA\nNumero=9\n" +
		"Modelo=55\nTipo=0\nFinalidade=1\n" +
		"[emit]\nRazao=EMPRESA\nLogradouro=RUA A\nCidadeCod=3106200\n" +
		"Cidade=BH\nUF=MG\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Ide.CNF != 123 {
		t.Errorf("Codigo -> cNF = %d", n.Ide.CNF)
	}
	if n.Ide.NatOp != "REMESSA" {
		t.Errorf("NaturezaOperacao -> natOp = %q", n.Ide.NatOp)
	}
	if n.Ide.NNF != 9 {
		t.Errorf("Numero -> nNF = %d", n.Ide.NNF)
	}
	if n.Ide.Modelo != 55 {
		t.Errorf("Modelo -> mod = %d", n.Ide.Modelo)
	}
	if n.Ide.TpNF != pcn.TnEntrada {
		t.Errorf("Tipo -> tpNF = %v", n.Ide.TpNF)
	}
	if n.Emit.XNome != "EMPRESA" {
		t.Errorf("Razao -> xNome = %q", n.Emit.XNome)
	}
	if n.Emit.EnderEmit.XLgr != "RUA A" {
		t.Errorf("Logradouro -> xLgr = %q", n.Emit.EnderEmit.XLgr)
	}
	if n.Emit.EnderEmit.CMun != 3106200 {
		t.Errorf("CidadeCod -> cMun = %d", n.Emit.EnderEmit.CMun)
	}

	// A chave do leiaute VENCE a chave por extenso quando as duas estao
	// presentes -- e a ordem da cadeia aninhada.
	comAsDuas := "[ide]\ncNF=1\nCodigo=2\n"
	n2, err := LerINI(comAsDuas, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Ide.CNF != 1 {
		t.Errorf("com as duas chaves, `cNF` vence; veio %d", n2.Ide.CNF)
	}
}

// TestLerINI_ModeloPadraoEhNFCe trava um default que surpreende.
//
// Faltando o `mod`, a nota sai como modelo 65 (NFC-e), nao 55. Quem supuser
// 55 gera nota de modelo errado a partir de todo .ini que omita a chave.
func TestLerINI_ModeloPadraoEhNFCe(t *testing.T) {
	n, err := LerINI("[ide]\nnatOp=VENDA\n", cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Ide.Modelo != 65 {
		t.Errorf("modelo padrao = %d, esperado 65 (NFC-e)", n.Ide.Modelo)
	}
}

// ---------------------------------------------------------------------------
// Efeitos cruzados entre grupos
// ---------------------------------------------------------------------------

// TestLerINI_EmitenteCompletaCUFeCMunFG trava os dois efeitos do emitente
// SOBRE a identificacao -- a razao de a ordem das chamadas nao ser livre.
func TestLerINI_EmitenteCompletaCUFeCMunFG(t *testing.T) {
	// Sem cUF e sem cMunFG: os dois vem do emitente.
	n, err := LerINI("[ide]\nnatOp=VENDA\n[emit]\nUF=MG\ncMun=3106200\n", cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Ide.CUF != 31 {
		t.Errorf("cUF deveria cair para o codigo da UF do emitente, veio %d", n.Ide.CUF)
	}
	if n.Ide.CMunFG != 3106200 {
		t.Errorf("cMunFG deveria cair para o cMun do emitente, veio %d", n.Ide.CMunFG)
	}

	// Com cMunFG na identificacao: o do emitente NAO sobrescreve.
	n2, err := LerINI("[ide]\ncMunFG=3550308\n[emit]\nUF=MG\ncMun=3106200\n", cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Ide.CMunFG != 3550308 {
		t.Errorf("o cMunFG da identificacao vence; veio %d", n2.Ide.CMunFG)
	}

	// Com cUF explicito no emitente: vence o derivado da UF.
	n3, err := LerINI("[ide]\nnatOp=X\n[emit]\nUF=MG\ncUF=99\n", cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n3.Ide.CUF != 99 {
		t.Errorf("cUF explicito deveria vencer, veio %d", n3.Ide.CUF)
	}
}

// ---------------------------------------------------------------------------
// Nota referenciada: o tipo inferido
// ---------------------------------------------------------------------------

// TestLerINI_TipoDaReferenciadaEhInferidoPelaChavePresente trava a cascata
// de inferencia, que e contrato e tem ORDEM.
//
// A ordem importa porque as chaves se sobrepoem: a NF de produtor rural
// tambem tem CNPJ, e se o teste de CNPJ viesse antes do de IE toda NFP
// seria lida como NF comum.
func TestLerINI_TipoDaReferenciadaEhInferidoPelaChavePresente(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[NFref001]\nrefNFe=" + chaveDeMinas + "\n" +
		"[NFref002]\nrefCTe=" + chaveDeMinas + "\n" +
		"[NFref003]\nnECF=001\nnCOO=123456\nmod=2B\n" +
		"[NFref004]\ncUF=31\nIE=1234567890\nCNPJ=99999999000191\nAAMM=2601\n" +
		"[NFref005]\ncUF=31\nCNPJ=99999999000191\nAAMM=2601\nmod=55\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Ide.NFref) != 5 {
		t.Fatalf("referenciadas lidas = %d, esperado 5", len(n.Ide.NFref))
	}

	if n.Ide.NFref[0].RefNFe != chaveDeMinas {
		t.Errorf("refNFe = %+v", n.Ide.NFref[0])
	}
	if n.Ide.NFref[1].RefCTe != chaveDeMinas {
		t.Errorf("refCTe = %+v", n.Ide.NFref[1])
	}
	if n.Ide.NFref[2].RefECF.NCOO != "123456" {
		t.Errorf("ECF = %+v", n.Ide.NFref[2])
	}

	// Com IE presente, a quarta e NFP -- e NAO NF, embora tenha CNPJ.
	if n.Ide.NFref[3].RefNFP.IE != "1234567890" {
		t.Errorf("com IE a referenciada e NFP, veio %+v", n.Ide.NFref[3])
	}
	if n.Ide.NFref[3].RefNF.CNPJ != "" {
		t.Error("a NFP nao pode preencher tambem o grupo RefNF")
	}

	// Sem IE, a quinta e NF comum.
	if n.Ide.NFref[4].RefNF.CNPJ != "99999999000191" {
		t.Errorf("sem IE a referenciada e NF, veio %+v", n.Ide.NFref[4])
	}
}

// ---------------------------------------------------------------------------
// Grupos condicionados
// ---------------------------------------------------------------------------

// TestLerINI_RetiradaEEntregaDependemDoLOGRADOURO trava uma condicional que
// nao e sobre o documento.
//
// O local de retirada pode nao ter CNPJ -- e um endereco. O ACBr condiciona
// a LEITURA ao xLgr (:405 e :431); condicionar ao CNPJCPF descartaria o
// grupo de toda nota que o informe sem documento.
func TestLerINI_RetiradaEEntregaDependemDoLOGRADOURO(t *testing.T) {
	semLogradouro := "[ide]\nnatOp=X\n[Retirada]\nCNPJ=99999999000191\nxMun=BH\n" +
		"[Entrega]\nCNPJ=99999999000191\nxMun=BH\n"
	n, err := LerINI(semLogradouro, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Retirada.CNPJCPF != "" || n.Entrega.CNPJCPF != "" {
		t.Error("sem xLgr os dois grupos ficam vazios, mesmo com CNPJ")
	}

	comLogradouro := "[ide]\nnatOp=X\n[Retirada]\nxLgr=RUA A\nxMun=BH\n" +
		"[Entrega]\nxLgr=RUA B\nIE=ISENTO\n"
	n2, err := LerINI(comLogradouro, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Retirada.XLgr != "RUA A" || n2.Retirada.XMun != "BH" {
		t.Errorf("retirada = %+v", n2.Retirada)
	}
	if n2.Entrega.XLgr != "RUA B" || n2.Entrega.IE != "ISENTO" {
		t.Errorf("entrega = %+v", n2.Entrega)
	}
}

// TestLerINI_AutXMLAceitaDuasLargurasDeIndice trava o unico grupo do
// formato que tenta DUAS larguras de secao.
//
// O ACBr grava com dois digitos ([autXML01]) e le tentando tres primeiro.
// Suportar so uma das larguras perde o grupo em metade dos arquivos.
func TestLerINI_AutXMLAceitaDuasLargurasDeIndice(t *testing.T) {
	casos := []struct {
		descricao string
		texto     string
	}{
		{"tres digitos", "[ide]\nnatOp=X\n[autXML001]\nCNPJ=99999999000191\n" +
			"[autXML002]\nCPF=11111111111\n"},
		{"dois digitos", "[ide]\nnatOp=X\n[autXML01]\nCNPJ=99999999000191\n" +
			"[autXML02]\nCPF=11111111111\n"},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n, err := LerINI(c.texto, cfgINI())
			if err != nil {
				t.Fatal(err)
			}
			if len(n.AutXML) != 2 {
				t.Fatalf("autXML lidos = %d, esperado 2", len(n.AutXML))
			}
			if n.AutXML[0].CNPJCPF != "99999999000191" || n.AutXML[1].CNPJCPF != "11111111111" {
				t.Errorf("autXML = %+v", n.AutXML)
			}
		})
	}
}

// TestLerINI_VolumeTerminaQuandoTODOSOsCamposFaltam trava um criterio de
// parada que nao e sentinela.
//
// Um volume que traga so o peso continua sendo volume. Usar uma unica
// chave como sentinela -- `qVol`, por exemplo -- descartaria esse caso e
// todos os seguintes.
func TestLerINI_VolumeTerminaQuandoTODOSOsCamposFaltam(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[vol001]\nqVol=2\nesp=CAIXA\n" +
		"[vol002]\npesoB=15.5\n" +
		"[vol003]\nMarca=ACME\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Transp.Vol) != 3 {
		t.Fatalf("volumes lidos = %d, esperado 3", len(n.Transp.Vol))
	}
	if n.Transp.Vol[1].PesoB != 15.5 {
		t.Errorf("o volume com so o peso tambem e volume: %+v", n.Transp.Vol[1])
	}
	if n.Transp.Vol[2].Marca != "ACME" {
		t.Errorf("o volume com so a marca tambem e volume: %+v", n.Transp.Vol[2])
	}
}

// TestLerINI_LacreTemIndiceDUPLO trava a largura do indice aninhado.
//
// A secao do lacre carrega TRES digitos do volume mais TRES do lacre. Errar
// a largura nao da erro -- a secao simplesmente nao e encontrada e a nota
// sai sem lacre.
func TestLerINI_LacreTemIndiceDUPLO(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[vol001]\nqVol=1\nesp=CAIXA\n" +
		"[Lacre001001]\nnLacre=AAA\n[Lacre001002]\nnLacre=BBB\n" +
		"[vol002]\nqVol=1\nesp=PALETE\n" +
		"[lacres002001]\nnLacre=CCC\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Transp.Vol) != 2 {
		t.Fatalf("volumes = %d", len(n.Transp.Vol))
	}
	if len(n.Transp.Vol[0].Lacres) != 2 {
		t.Fatalf("lacres do primeiro volume = %+v", n.Transp.Vol[0].Lacres)
	}
	if n.Transp.Vol[0].Lacres[1].NLacre != "BBB" {
		t.Errorf("lacre = %+v", n.Transp.Vol[0].Lacres)
	}
	// O segundo volume usa a grafia alternativa da secao.
	if len(n.Transp.Vol[1].Lacres) != 1 || n.Transp.Vol[1].Lacres[0].NLacre != "CCC" {
		t.Errorf("lacres do segundo volume = %+v", n.Transp.Vol[1].Lacres)
	}
}

// ---------------------------------------------------------------------------
// Itens
// ---------------------------------------------------------------------------

// TestLerINI_ItemComLarguraDeIndiceDiferentePorSubgrupo trava as tres
// larguras que convivem no mesmo item.
//
// gCred usa UM digito, origComb usa DOIS e os demais usam TRES. Uniformizar
// qualquer um deles faz o subgrupo sumir sem erro.
func TestLerINI_ItemComLarguraDeIndiceDiferentePorSubgrupo(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=PRODUTO\nuCom=UN\nqCom=2\nvUnCom=10\nvProd=20\n" +
		"[NVE001001]\nNVE=AB1234\n" +
		"[gCred0011]\ncCredPresumido=01\npCredPresumido=5\nvCredPresumido=1\n" +
		"[Combustivel001]\ncProdANP=210203001\ndescANP=GASOLINA\n" +
		"[origComb00101]\ncUFOrig=31\npOrig=100\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Det) != 1 {
		t.Fatalf("itens = %d", len(n.Det))
	}
	p := n.Det[0].Prod

	if len(p.NVE) != 1 || p.NVE[0].NVE != "AB1234" {
		t.Errorf("NVE (3+3 digitos) = %+v", p.NVE)
	}
	if len(p.CredPresumido) != 1 || p.CredPresumido[0].CCredPresumido != "01" {
		t.Errorf("gCred (3+1 digitos) = %+v", p.CredPresumido)
	}
	if len(p.Comb.OrigComb) != 1 || p.Comb.OrigComb[0].CUFOrig != 31 {
		t.Errorf("origComb (3+2 digitos) = %+v", p.Comb.OrigComb)
	}
}

// TestLerINI_ItemSoTerminaQuandoAS_DUAS_ChavesFaltam trava a sentinela do
// item, que distingue chave AUSENTE de chave VAZIA.
//
// A linha do original e
//
//	sProdID := ReadString(s, 'Codigo', ReadString(s, 'cProd', 'FIM'))
//	if sProdID = 'FIM' then break
//
// Faltando as duas chaves da "FIM" e o laco acaba -- e e assim que ele
// acaba, porque secao inexistente cai nesse caso. Mas um `cProd=` VAZIO
// devolve string vazia, que NAO e "FIM": o item entra na nota. Tratar os
// dois casos igual ou descarta um item valido ou faz o laco criar itens
// vazios ate o teto.
func TestLerINI_ItemSoTerminaQuandoAS_DUAS_ChavesFaltam(t *testing.T) {
	// O segundo item tem a chave PRESENTE e vazia: entra na nota, e o
	// terceiro continua sendo lido.
	texto := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=PRIMEIRO\n" +
		"[det002]\ncProd=\nxProd=SEGUNDO\n" +
		"[det003]\ncProd=P3\nxProd=TERCEIRO\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Det) != 3 {
		t.Fatalf("itens = %d, esperado 3: codigo vazio NAO encerra o laco", len(n.Det))
	}
	if n.Det[1].Prod.XProd != "SEGUNDO" {
		t.Errorf("o item de codigo vazio entra na nota: %+v", n.Det[1].Prod)
	}

	// Uma secao que existe mas nao tem NENHUMA das duas chaves encerra,
	// porque a leitura mais interna cai no default "FIM".
	semAsDuas := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=PRIMEIRO\n" +
		"[det002]\nxProd=SEM CODIGO\n" +
		"[det003]\ncProd=P3\nxProd=TERCEIRO\n"

	n2, err := LerINI(semAsDuas, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n2.Det) != 1 {
		t.Fatalf("itens = %d, esperado 1: sem as duas chaves o laco encerra", len(n2.Det))
	}

	// E o literal "FIM" encerra, que e como os arquivos antigos marcam o
	// fim da lista.
	comFIM := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=PRIMEIRO\n" +
		"[det002]\ncProd=FIM\n"

	n3, err := LerINI(comFIM, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if len(n3.Det) != 1 {
		t.Errorf("itens = %d, esperado 1: o literal FIM encerra", len(n3.Det))
	}
}

// TestLerINI_UnidadeTributavelCaiParaAComercial trava tres defaults
// encadeados que dependem da ORDEM de leitura.
//
// uTrib, qTrib e vUnTrib caem para uCom, qCom e vUnCom. Ler os tributaveis
// ANTES dos comerciais deixaria os tres zerados, e a nota sairia sem
// unidade tributavel -- rejeicao certa.
func TestLerINI_UnidadeTributavelCaiParaAComercial(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=P\nuCom=CX\nqCom=3\nvUnCom=25.5\nvProd=76.5\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	p := n.Det[0].Prod
	if p.UTrib != "CX" || p.QTrib != 3 || p.VUnTrib != 25.5 {
		t.Errorf("os tributaveis deveriam cair para os comerciais: %+v", p)
	}

	// Informados, vencem o default.
	texto2 := texto + "uTrib=UN\nqTrib=36\nvUnTrib=2.125\n"
	n2, err := LerINI(texto2, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	p2 := n2.Det[0].Prod
	if p2.UTrib != "UN" || p2.QTrib != 36 || p2.VUnTrib != 2.125 {
		t.Errorf("os tributaveis informados vencem: %+v", p2)
	}
}

// TestLerINI_ICMSAceitaCSTouCSOSN trava a sentinela dupla do grupo.
//
// O emitente do Simples Nacional nao tem CST, tem CSOSN. Testar so um dos
// dois deixa metade das notas sem o grupo de ICMS inteiro.
func TestLerINI_ICMSAceitaCSTouCSOSN(t *testing.T) {
	base := "[ide]\nnatOp=X\n[det001]\ncProd=P1\nxProd=P\n"

	comCST := base + "[ICMS001]\norig=0\nCST=00\nvBC=100\npICMS=18\nvICMS=18\n"
	n, err := LerINI(comCST, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Det[0].Imposto.ICMS.VBC != 100 {
		t.Errorf("ICMS por CST = %+v", n.Det[0].Imposto.ICMS)
	}

	comCSOSN := base + "[ICMS001]\norig=0\nCSOSN=101\npCredSN=2.5\nvCredICMSSN=2.5\n"
	n2, err := LerINI(comCSOSN, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Det[0].Imposto.ICMS.VCredICMSSN != 2.5 {
		t.Errorf("ICMS por CSOSN = %+v", n2.Det[0].Imposto.ICMS)
	}

	semNenhum := base + "[ICMS001]\norig=0\nvBC=100\n"
	n3, err := LerINI(semNenhum, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n3.Det[0].Imposto.ICMS.VBC != 0 {
		t.Error("sem CST e sem CSOSN o grupo inteiro nao e lido")
	}
}

// TestLerINI_PISSTUsaAConcatenacaoComoSentinela trava a sentinela mais
// exotica do formato.
//
// O original concatena duas leituras com defaults "F" e "IM" e compara com
// "FIM": faltando as duas chaves, da exatamente "FIM" e o grupo nao e
// lido; havendo qualquer uma, o grupo e lido. Qualquer leitura que exija
// uma chave especifica perde metade dos casos -- o grupo aceita base de
// calculo OU quantidade.
func TestLerINI_PISSTUsaAConcatenacaoComoSentinela(t *testing.T) {
	base := "[ide]\nnatOp=X\n[det001]\ncProd=P1\nxProd=P\n"

	casos := []struct {
		descricao string
		secao     string
		esperaLer bool
	}{
		{"por base de calculo", "[PISST001]\nvBC=100\npPis=1.65\nvPIS=1.65\n", true},
		{"por quantidade", "[PISST001]\nqBCProd=10\nvAliqProd=0.5\nvPIS=5\n", true},
		{"pelo nome por extenso", "[PISST001]\nValorBase=100\nAliquotaPerc=1.65\nValorPISST=1.65\n", true},
		{"vazio", "[PISST001]\nindSomaPISST=0\n", false},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n, err := LerINI(base+c.secao, cfgINI())
			if err != nil {
				t.Fatal(err)
			}
			lido := n.Det[0].Imposto.PISST != (PISST{})
			if lido != c.esperaLer {
				t.Errorf("lido = %v, esperado %v: %+v",
					lido, c.esperaLer, n.Det[0].Imposto.PISST)
			}
		})
	}
}

// TestLerINI_ISSQNExigeBaseDeCalculoPOSITIVA trava a unica condicional de
// VALOR do item -- as demais sao de presenca.
func TestLerINI_ISSQNExigeBaseDeCalculoPOSITIVA(t *testing.T) {
	base := "[ide]\nnatOp=X\n[det001]\ncProd=P1\nxProd=P\n"

	zerada := base + "[ISSQN001]\nvBC=0\nvAliq=5\nvISSQN=0\ncListServ=0101\n"
	n, err := LerINI(zerada, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Det[0].Imposto.ISSQN.CListServ != "" {
		t.Errorf("base zerada nao le o grupo: %+v", n.Det[0].Imposto.ISSQN)
	}

	positiva := base + "[ISSQN001]\nvBC=200\nvAliq=5\nvISSQN=10\ncListServ=0101\n"
	n2, err := LerINI(positiva, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Det[0].Imposto.ISSQN.VBC != 200 || n2.Det[0].Imposto.ISSQN.CListServ != "0101" {
		t.Errorf("ISSQN = %+v", n2.Det[0].Imposto.ISSQN)
	}
}

// TestGravarINI_CamposDaDIDependemDoENUM_QueOsAcompanha trava duas
// condicionais que nao olham o proprio campo.
//
// O `vAFRMM` e o adicional ao frete da marinha mercante: so existe na via
// MARITIMA. O CNPJ e a UF do terceiro so existem quando ha intermediacao --
// em conta propria nao ha terceiro. Os dois campos podem estar preenchidos
// na struct e ainda assim nao devem sair.
//
// Nenhum round-trip pega isso: o campo volta com o valor que foi gravado, e
// gravar a mais so aparece quando a SEFAZ recusa.
func TestGravarINI_CamposDaDIDependemDoENUM_QueOsAcompanha(t *testing.T) {
	base := DI{
		NDi: "DI001", XLocDesemb: "SANTOS", UFDesemb: "SP",
		VAFRMM: 150.25,
		CNPJ:   "55555555000191", UFTerceiro: "SP",
	}

	t.Run("vAFRMM so na via maritima", func(t *testing.T) {
		maritima := base
		maritima.TpViaTransp = TvMaritima

		ini := pcn.NovoINI()
		gravarINIProdutoDI(ini, 0, []DI{maritima})
		if ini.LerFloat("DI001001", "vAFRMM", 0) != 150.25 {
			t.Error("na via maritima o vAFRMM sai")
		}

		aerea := base
		aerea.TpViaTransp = TvAerea

		ini = pcn.NovoINI()
		gravarINIProdutoDI(ini, 0, []DI{aerea})
		if ini.LerString("DI001001", "vAFRMM", "AUSENTE") != "AUSENTE" {
			t.Error("fora da via maritima o vAFRMM NAO sai, por mais preenchido que esteja")
		}
		if ini.LerString("DI001001", "tpViaTransp", "") == "" {
			t.Error("a via em si continua saindo")
		}
	})

	t.Run("terceiro so fora de conta propria", func(t *testing.T) {
		porOrdem := base
		porOrdem.TpIntermedio = TiContaOrdem

		ini := pcn.NovoINI()
		gravarINIProdutoDI(ini, 0, []DI{porOrdem})
		if ini.LerString("DI001001", "CNPJ", "") != "55555555000191" {
			t.Error("havendo intermediacao o CNPJ do terceiro sai")
		}

		propria := base
		propria.TpIntermedio = TiContaPropria

		ini = pcn.NovoINI()
		gravarINIProdutoDI(ini, 0, []DI{propria})
		if ini.LerString("DI001001", "CNPJ", "AUSENTE") != "AUSENTE" {
			t.Error("em conta propria nao ha terceiro: o CNPJ NAO sai")
		}
		if ini.LerString("DI001001", "UFTerceiro", "AUSENTE") != "AUSENTE" {
			t.Error("em conta propria a UF do terceiro tambem nao sai")
		}
	})
}

// TestLerINI_ImpostoDevolSOBRESCREVE_ASecaoDoItem trava a precedencia entre
// duas secoes que carregam os MESMOS dois campos.
//
// O `pDevol` e o `vIPIDevol` aparecem na secao do produto E em
// [impostoDevolNNN], e o segundo vence (:552-558). O gravador emite valores
// iguais nos dois lugares, entao o round-trip nao distingue -- so um .ini
// com valores DIFERENTES mostra quem manda.
func TestLerINI_ImpostoDevolSOBRESCREVE_ASecaoDoItem(t *testing.T) {
	texto := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=P\npDevol=10\nvIPIDevol=1\n" +
		"[impostoDevol001]\npDevol=55\nvIPIDevol=9.9\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n.Det[0].PDevol != 55 || n.Det[0].VIPIDevol != 9.9 {
		t.Errorf("a secao impostoDevol vence a do item: pDevol=%v vIPIDevol=%v",
			n.Det[0].PDevol, n.Det[0].VIPIDevol)
	}

	// Sem a secao propria, valem os da secao do item.
	semSecao := "[ide]\nnatOp=X\n" +
		"[det001]\ncProd=P1\nxProd=P\npDevol=10\nvIPIDevol=1\n"

	n2, err := LerINI(semSecao, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Det[0].PDevol != 10 || n2.Det[0].VIPIDevol != 1 {
		t.Errorf("sem a secao propria valem os do item: %v/%v",
			n2.Det[0].PDevol, n2.Det[0].VIPIDevol)
	}
}

// TestLerINI_InteiroCascateiaSobreChaveVaziaEOFloatNAO trava uma diferenca
// entre os dois tipos que vem do Delphi e nao da conveniencia.
//
// O ReadInteger do TCustomIniFile e
//
//	StrToIntDef(ReadString(Secao, Chave, ''), Default)
//
// e numa cadeia `ReadInteger(a, ReadInteger(b, 0))` o Default do externo e o
// VALOR do interno: chave presente e VAZIA faz o StrToIntDef falhar e cair
// para o alias.
//
// Com decimal nao ha ReadFloat encadeado -- o que encadeia sao os ReadString
// dentro de um StringToFloatDef --, e a cadeia de ReadString PARA na
// primeira chave presente. Chave presente e vazia entrega ZERO, sem
// consultar o alias.
//
// Arquivo de ERP antigo que emite a chave do leiaute vazia e a em portugues
// preenchida e exatamente o caso que os aliases existem para atender.
func TestLerINI_InteiroCascateiaSobreChaveVaziaEOFloatNAO(t *testing.T) {
	texto := "[ide]\ncNF=\nCodigo=12345\nnatOp=X\n" +
		"[Total]\nvProd=\nValorProduto=999\n"

	n, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}

	if n.Ide.CNF != 12345 {
		t.Errorf("inteiro: `cNF=` vazio cede para `Codigo`; veio %d", n.Ide.CNF)
	}
	if n.Total.ICMSTot.VProd != 0 {
		t.Errorf("decimal: `vProd=` vazio entrega ZERO e NAO consulta o alias; veio %v",
			n.Total.ICMSTot.VProd)
	}

	// Com a chave do leiaute AUSENTE, os dois caem para o alias.
	ausentes := "[ide]\nCodigo=12345\nnatOp=X\n" +
		"[Total]\nValorProduto=999\n"

	n2, err := LerINI(ausentes, cfgINI())
	if err != nil {
		t.Fatal(err)
	}
	if n2.Ide.CNF != 12345 || n2.Total.ICMSTot.VProd != 999 {
		t.Errorf("com a chave ausente os dois usam o alias: cNF=%d vProd=%v",
			n2.Ide.CNF, n2.Total.ICMSTot.VProd)
	}
}

// ---------------------------------------------------------------------------
// Escrita
// ---------------------------------------------------------------------------

// notaDeTeste monta uma nota com todos os grupos que o .ini cobre.
func notaDeTeste(t *testing.T) *NFe {
	t.Helper()

	n := NovaNFe()
	n.InfNFe.ID = "NFe" + chaveValidaINI
	n.InfNFe.Versao = 4.00

	n.Ide = Ide{
		CUF: 31, CNF: 12345678, NatOp: "VENDA", Modelo: 55, Serie: 1, NNF: 456,
		DEmi:    time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC),
		DSaiEnt: time.Date(2026, 1, 20, 14, 30, 0, 0, time.UTC),
		TpNF:    pcn.TnSaida, IdDest: DoInterestadual, CMunFG: 3106200,
		TpImp: pcn.TiRetrato, TpEmis: pcn.TeNormal, TpAmb: pcn.TaHomologacao,
		FinNFe: FnNormal, IndFinal: CfNao, IndPres: PcPresencial,
		ProcEmi: pcn.PeAplicativoContribuinte, VerProc: "OpenFiscalBR",
		NFref: []NFref{
			{RefNFe: chaveDeMinas},
			{RefNF: RefNF{CUF: 31, AAMM: "2601", CNPJ: "99999999000191", Modelo: 55, Serie: 1, NNF: 10}},
			{RefECF: RefECF{Modelo: ECFModRef2D, NECF: "001", NCOO: "123456"}},
		},
	}

	n.Emit = NovoEmit()
	n.Emit.CNPJCPF = "99999999000191"
	n.Emit.XNome = "EMPRESA EMITENTE"
	n.Emit.IE = "1234567890"
	n.Emit.EnderEmit = EnderEmit{
		XLgr: "RUA DAS FLORES", Nro: "100", XBairro: "CENTRO",
		CMun: 3106200, XMun: "BELO HORIZONTE", UF: "MG", CEP: 30000000,
		CPais: 1058, XPais: "BRASIL", Fone: "3133334444",
	}

	n.Dest.CNPJCPF = "11111111000191"
	n.Dest.XNome = "CLIENTE"
	n.Dest.IndIEDest = pcn.InContribuinte
	n.Dest.IE = "9876543210"
	n.Dest.EnderDest = EnderDest{
		XLgr: "AV PAULISTA", Nro: "2000", XBairro: "BELA VISTA",
		CMun: 3550308, XMun: "SAO PAULO", UF: "SP", CEP: 1310000,
		CPais: 1058, XPais: "BRASIL",
	}

	n.Entrega.CNPJCPF = "11111111000191"
	n.Entrega.XLgr = "RUA DA ENTREGA"
	n.Entrega.XMun = "SAO PAULO"
	n.Entrega.UF = "SP"
	n.Entrega.IE = "ISENTO"

	n.AutXML = []AutXML{{CNPJCPF: "22222222000191"}}

	item := NovoDetItem()
	item.Prod.NItem = 1
	item.Prod.CProd = "P001"
	item.Prod.CEAN = "7891234567890"
	item.Prod.XProd = "PRODUTO DE TESTE"
	item.Prod.NCM = "22021000"
	item.Prod.CFOP = "6102"
	item.Prod.UCom = "UN"
	item.Prod.QCom = 10
	item.Prod.VUnCom = 25.5
	item.Prod.VProd = 255
	item.Prod.UTrib = "UN"
	item.Prod.QTrib = 10
	item.Prod.VUnTrib = 25.5
	item.Prod.IndTot = ItSomaTotalNFe
	item.Prod.NVE = []NVE{{NVE: "AB1234"}}
	item.Prod.CredPresumido = []CredPresumido{
		{CCredPresumido: "01", PCredPresumido: 5, VCredPresumido: 12.75},
	}
	item.Prod.Rastro = []Rastro{{
		NLote: "L2026A", QLote: 10,
		DFab:   time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		DVal:   time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC),
		CAgreg: "AG1",
	}}
	item.Prod.Arma = []Arma{
		{TpArma: TaUsoPermitido, NSerie: "SR123", NCano: "CN456", Descr: "PISTOLA"},
	}
	// O detExport sai com `nRe` e `chNFe` VAZIOS quando so ha drawback --
	// que e o caso comum do grupo e o que quebrava a sentinela do leitor.
	item.Prod.DetExport = []DetExport{{NDraw: "DRAW1", QExport: 5}}
	item.Prod.DI = []DI{{
		NDi: "DI001", XLocDesemb: "SANTOS", UFDesemb: "SP",
		DDi:     time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
		DDesemb: time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		// tpViaTransp maritima libera o vAFRMM; tpIntermedio fora de conta
		// propria libera CNPJ e UFTerceiro.
		TpViaTransp: TvMaritima, VAFRMM: 150.25,
		TpIntermedio: TiContaOrdem, CNPJ: "55555555000191", UFTerceiro: "SP",
		CExportador: "EXP1",
		Adi: []Adi{
			{NAdicao: 1, NSeqAdi: 1, CFabricante: "FAB1", VDescDI: 10, NDraw: "D1"},
			{NAdicao: 2, NSeqAdi: 2, CFabricante: "FAB2", VDescDI: 20},
		},
	}}
	item.Prod.Comb = Comb{
		CProdANP: 210203001, DescANP: "GASOLINA C", PGLP: 0, UFCons: "MG",
		CODIF: "COD1", QTemp: 1000.5, PBio: 27.5,
		OrigComb:   []OrigComb{{IndImport: IiNacional, CUFOrig: 31, POrig: 100}},
		CIDE:       CIDE{QBCProd: 1000, VAliqProd: 0.1, VCIDE: 100},
		Encerrante: Encerrante{NBico: 1, NBomba: 2, NTanque: 3, VEncIni: 100.5, VEncFin: 200.75},
		ICMS:       ICMSComb{VBCICMS: 500, VICMS: 90, VBCICMSST: 600, VICMSST: 108},
		ICMSInter:  ICMSInter{VBCICMSSTDest: 300, VICMSSTDest: 54},
		ICMSCons:   ICMSCons{VBCICMSSTCons: 200, VICMSSTCons: 36, UFCons: "SP"},
	}
	// pDevol POSITIVO: o grupo [impostoDevolNNN] passa a sair, e o leitor o
	// faz SOBRESCREVER o que veio da secao do item.
	item.PDevol = 50
	item.VIPIDevol = 7.5
	item.ObsCont = ObsItem{XCampo: "OBS1", XTexto: "TEXTO DO ITEM"}
	item.ObsFisco = ObsItem{XCampo: "FIS1", XTexto: "TEXTO DO FISCO"}
	item.Imposto.ICMS = ICMS{
		Orig: pcn.OeNacional, CST: pcn.CST00, ModBC: DbiValorOperacao,
		VBC: 255, PICMS: 18, VICMS: 45.9,
	}
	item.Imposto.PIS = PIS{CST: pcn.Pis01, VBC: 255, PPIS: 1.65, VPIS: 4.21}
	item.Imposto.COFINS = COFINS{CST: pcn.Cof01, VBC: 255, PCOFINS: 7.6, VCOFINS: 19.38}
	n.Det = append(n.Det, item)

	n.Total.ICMSTot = ICMSTot{
		VBC: 255, VICMS: 45.9, VProd: 255, VNF: 255,
		VPIS: 4.21, VCOFINS: 19.38,
	}

	n.Transp.ModFrete = MfContaEmitente
	n.Transp.Transporta.CNPJCPF = "33333333000191"
	n.Transp.Transporta.XNome = "TRANSPORTADORA"
	n.Transp.Vol = []Vol{{
		QVol: 2, Esp: "CAIXA", Marca: "ACME", NVol: "001",
		PesoL: 10.5, PesoB: 11,
		Lacres: []Lacre{{NLacre: "L1"}, {NLacre: "L2"}},
	}}
	n.Transp.Reboque = []Reboque{{Placa: "ABC1234", UF: "MG", RNTC: "999"}}

	n.Cobr.Fat = Fat{NFat: "F001", VOrig: 255, VLiq: 255}
	n.Cobr.Dup = []Dup{
		{NDup: "001", DVenc: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), VDup: 255},
	}

	pag := NovoPagItem()
	pag.TPag = FpDinheiro
	pag.VPag = 255
	n.Pag.DetPag = append(n.Pag.DetPag, pag)
	n.Pag.VTroco = 5

	n.InfAdic.InfCpl = "OBSERVACAO"
	n.InfAdic.ObsCont = []ObsCont{{XCampo: "C1", XTexto: "T1"}}
	n.InfAdic.ObsFisco = []ObsFisco{{XCampo: "F1", XTexto: "T2"}}
	n.InfAdic.ProcRef = []ProcRef{{NProc: "PROC1", IndProc: IpSEFAZ}}

	n.InfRespTec = InfRespTec{
		CNPJ: "44444444000191", XContato: "SUPORTE",
		Email: "suporte@exemplo.com.br", Fone: "3133335555",
	}

	return &n
}

// TestGravarINI_UsaONomeLONGODaSecao trava a convencao do gravador.
//
// O ACBr grava [Identificacao], nao [ide]. O leitor aceita os dois, mas o
// arquivo precisa sair igual ao do original -- e ha ERP que procura a
// secao pelo nome literal.
func TestGravarINI_UsaONomeLONGODaSecao(t *testing.T) {
	ini, err := MontarINI(notaDeTeste(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, secao := range []string{
		"Identificacao", "Emitente", "Destinatario", "Produto001",
		"Transportador", "Fatura", "Duplicata001", "Volume001",
	} {
		if !ini.SecaoExiste(secao) {
			t.Errorf("a secao %q deveria existir", secao)
		}
	}
}

// TestINI_RoundTrip grava a nota, rele e compara.
//
// Nao substitui os testes de condicional -- um gravador e um leitor com o
// MESMO engano passam neste teste --, mas pega perda de campo e troca de
// indice.
func TestINI_RoundTrip(t *testing.T) {
	original := notaDeTeste(t)

	texto, err := GravarINI(original)
	if err != nil {
		t.Fatal(err)
	}
	relido, err := LerINI(texto, cfgINI())
	if err != nil {
		t.Fatal(err)
	}

	if relido.Ide.NatOp != original.Ide.NatOp || relido.Ide.NNF != original.Ide.NNF ||
		relido.Ide.Modelo != original.Ide.Modelo || relido.Ide.Serie != original.Ide.Serie {
		t.Errorf("ide: %+v", relido.Ide)
	}
	if !relido.Ide.DEmi.Equal(original.Ide.DEmi) {
		t.Errorf("dhEmi = %v, esperado %v", relido.Ide.DEmi, original.Ide.DEmi)
	}
	if relido.Ide.CUF != original.Ide.CUF || relido.Ide.CMunFG != original.Ide.CMunFG {
		t.Errorf("cUF/cMunFG = %d/%d", relido.Ide.CUF, relido.Ide.CMunFG)
	}
	if len(relido.Ide.NFref) != len(original.Ide.NFref) {
		t.Errorf("NFref: %d vs %d", len(relido.Ide.NFref), len(original.Ide.NFref))
	}

	if relido.Emit.CNPJCPF != original.Emit.CNPJCPF ||
		relido.Emit.EnderEmit != original.Emit.EnderEmit {
		t.Errorf("emit: %+v", relido.Emit)
	}
	if relido.Dest.EnderDest != original.Dest.EnderDest {
		t.Errorf("dest: %+v", relido.Dest.EnderDest)
	}
	if relido.Entrega.IE != original.Entrega.IE {
		t.Errorf("entrega: %+v", relido.Entrega)
	}
	if len(relido.AutXML) != 1 || relido.AutXML[0] != original.AutXML[0] {
		t.Errorf("autXML: %+v", relido.AutXML)
	}

	if len(relido.Det) != 1 {
		t.Fatalf("itens: %d", len(relido.Det))
	}
	po, pr := original.Det[0].Prod, relido.Det[0].Prod
	if pr.CProd != po.CProd || pr.XProd != po.XProd || pr.QCom != po.QCom ||
		pr.VUnCom != po.VUnCom || pr.VProd != po.VProd {
		t.Errorf("prod: %+v", pr)
	}
	if len(pr.NVE) != 1 || pr.NVE[0] != po.NVE[0] {
		t.Errorf("NVE: %+v", pr.NVE)
	}
	if len(pr.CredPresumido) != 1 || pr.CredPresumido[0] != po.CredPresumido[0] {
		t.Errorf("gCred: %+v", pr.CredPresumido)
	}
	if len(pr.Rastro) != 1 || pr.Rastro[0].NLote != po.Rastro[0].NLote ||
		!pr.Rastro[0].DVal.Equal(po.Rastro[0].DVal) {
		t.Errorf("rastro: %+v", pr.Rastro)
	}
	if len(pr.Arma) != 1 || pr.Arma[0] != po.Arma[0] {
		t.Errorf("arma: %+v", pr.Arma)
	}

	// O detExport com `nRe` e `chNFe` vazios tem de sobreviver: e o caso
	// mais comum do grupo e o que o gravador produz por padrao.
	if len(pr.DetExport) != 1 {
		t.Fatalf("detExport: %d itens -- o grupo com so o drawback se perdeu", len(pr.DetExport))
	}
	if pr.DetExport[0].NDraw != po.DetExport[0].NDraw ||
		pr.DetExport[0].QExport != po.DetExport[0].QExport {
		t.Errorf("detExport: %+v", pr.DetExport[0])
	}

	if len(pr.DI) != 1 {
		t.Fatalf("DI: %d itens", len(pr.DI))
	}
	di, dio := pr.DI[0], po.DI[0]
	if di.NDi != dio.NDi || di.XLocDesemb != dio.XLocDesemb || di.UFDesemb != dio.UFDesemb {
		t.Errorf("DI: %+v", di)
	}
	if di.TpViaTransp != dio.TpViaTransp || di.VAFRMM != dio.VAFRMM {
		t.Errorf("DI/tpViaTransp: o vAFRMM so sai na via maritima: %+v", di)
	}
	if di.TpIntermedio != dio.TpIntermedio || di.CNPJ != dio.CNPJ || di.UFTerceiro != dio.UFTerceiro {
		t.Errorf("DI/tpIntermedio: CNPJ e UFTerceiro so saem fora de conta propria: %+v", di)
	}
	// [LADINNNJJJKKK] -- indice TRIPLO. Errar a largura nao da erro, so
	// esvazia o grupo.
	if len(di.Adi) != 2 {
		t.Fatalf("adicoes da DI: %d, esperado 2 -- confira o indice 3+3+3", len(di.Adi))
	}
	if di.Adi[1].NAdicao != 2 || di.Adi[1].CFabricante != "FAB2" {
		t.Errorf("segunda adicao: %+v", di.Adi[1])
	}

	cb, cbo := pr.Comb, po.Comb
	if cb.CProdANP != cbo.CProdANP || cb.DescANP != cbo.DescANP || cb.PBio != cbo.PBio {
		t.Errorf("combustivel: %+v", cb)
	}
	if len(cb.OrigComb) != 1 || cb.OrigComb[0] != cbo.OrigComb[0] {
		t.Errorf("origComb: %+v", cb.OrigComb)
	}
	if cb.CIDE != cbo.CIDE {
		t.Errorf("CIDE: %+v", cb.CIDE)
	}
	if cb.Encerrante != cbo.Encerrante {
		t.Errorf("encerrante: %+v", cb.Encerrante)
	}
	if cb.ICMS != cbo.ICMS || cb.ICMSInter != cbo.ICMSInter || cb.ICMSCons != cbo.ICMSCons {
		t.Errorf("ICMS do combustivel: %+v / %+v / %+v", cb.ICMS, cb.ICMSInter, cb.ICMSCons)
	}

	if relido.Det[0].PDevol != original.Det[0].PDevol ||
		relido.Det[0].VIPIDevol != original.Det[0].VIPIDevol {
		t.Errorf("impostoDevol: %v/%v", relido.Det[0].PDevol, relido.Det[0].VIPIDevol)
	}
	if relido.Det[0].ObsCont != original.Det[0].ObsCont ||
		relido.Det[0].ObsFisco != original.Det[0].ObsFisco {
		t.Errorf("observacoes do item: %+v / %+v",
			relido.Det[0].ObsCont, relido.Det[0].ObsFisco)
	}

	io, ir := original.Det[0].Imposto, relido.Det[0].Imposto
	if ir.ICMS.CST != io.ICMS.CST || ir.ICMS.VBC != io.ICMS.VBC ||
		ir.ICMS.VICMS != io.ICMS.VICMS {
		t.Errorf("ICMS: %+v", ir.ICMS)
	}
	if ir.PIS != io.PIS {
		t.Errorf("PIS: %+v vs %+v", ir.PIS, io.PIS)
	}
	if ir.COFINS.VCOFINS != io.COFINS.VCOFINS {
		t.Errorf("COFINS: %+v", ir.COFINS)
	}

	if relido.Total.ICMSTot != original.Total.ICMSTot {
		t.Errorf("total: %+v", relido.Total.ICMSTot)
	}

	if relido.Transp.ModFrete != original.Transp.ModFrete ||
		relido.Transp.Transporta != original.Transp.Transporta {
		t.Errorf("transp: %+v", relido.Transp)
	}
	if len(relido.Transp.Vol) != 1 {
		t.Fatalf("volumes: %d", len(relido.Transp.Vol))
	}
	if len(relido.Transp.Vol[0].Lacres) != 2 {
		t.Errorf("lacres: %+v", relido.Transp.Vol[0].Lacres)
	}
	if len(relido.Transp.Reboque) != 1 || relido.Transp.Reboque[0] != original.Transp.Reboque[0] {
		t.Errorf("reboque: %+v", relido.Transp.Reboque)
	}

	if relido.Cobr.Fat != original.Cobr.Fat {
		t.Errorf("fatura: %+v", relido.Cobr.Fat)
	}
	if len(relido.Cobr.Dup) != 1 || relido.Cobr.Dup[0].NDup != "001" ||
		!relido.Cobr.Dup[0].DVenc.Equal(original.Cobr.Dup[0].DVenc) {
		t.Errorf("duplicata: %+v", relido.Cobr.Dup)
	}

	if len(relido.Pag.DetPag) != 1 || relido.Pag.DetPag[0].VPag != 255 {
		t.Errorf("pagamento: %+v", relido.Pag)
	}
	if relido.Pag.VTroco != 5 {
		t.Errorf("vTroco = %v -- e da COLECAO, e sai na ultima secao", relido.Pag.VTroco)
	}

	if relido.InfAdic.InfCpl != original.InfAdic.InfCpl ||
		len(relido.InfAdic.ObsCont) != 1 || len(relido.InfAdic.ObsFisco) != 1 ||
		len(relido.InfAdic.ProcRef) != 1 {
		t.Errorf("infAdic: %+v", relido.InfAdic)
	}
	if relido.InfRespTec != original.InfRespTec {
		t.Errorf("infRespTec: %+v", relido.InfRespTec)
	}
}

// TestGravarINI_MedicamentoTrocaDeCamposPorVERSAO trava a unica condicional
// por versao de leiaute do .ini da NF-e.
//
// A 4.00 tirou o controle de lote do medicamento -- ele passou para o grupo
// `rastro` -- e pos o codigo ANVISA no lugar. Gravar os dois conjuntos
// sempre produz arquivo que o validador recusa, e nenhum round-trip
// perceberia.
func TestGravarINI_MedicamentoTrocaDeCamposPorVERSAO(t *testing.T) {
	med := Med{
		CProdANVISA: "123456789", XMotivoIsencao: "ISENTO",
		NLote: "LOTE1", QLote: 10,
		DFab: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		DVal: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		VPMC: 99.9,
	}

	casos := []struct {
		versao    float64
		presentes []string
		ausentes  []string
	}{
		{4.00, []string{"cProdANVISA", "xMotivoIsencao"}, []string{"nLote", "qLote", "dFab", "dVal"}},
		{3.10, []string{"nLote", "qLote", "dFab", "dVal"}, []string{"cProdANVISA", "xMotivoIsencao"}},
	}

	for _, c := range casos {
		t.Run(fmt.Sprintf("versao_%.2f", c.versao), func(t *testing.T) {
			ini := pcn.NovoINI()
			gravarINIProdutoMedicamento(ini, 0, []Med{med}, c.versao)

			const secao = "Medicamento001001"
			for _, chave := range c.presentes {
				if ini.LerString(secao, chave, "") == "" {
					t.Errorf("na %.2f a chave %q deveria sair", c.versao, chave)
				}
			}
			for _, chave := range c.ausentes {
				if ini.LerString(secao, chave, "AUSENTE") != "AUSENTE" {
					t.Errorf("na %.2f a chave %q NAO deveria sair", c.versao, chave)
				}
			}
			// O vPMC sai nas duas.
			if ini.LerFloat(secao, "vPMC", 0) != 99.9 {
				t.Error("o vPMC sai em qualquer versao")
			}
		})
	}
}

// TestGravarINI_PISeCOFINSEscolhemOsCamposPeloCST trava o `case` sem
// `else` dos dois grupos.
//
// vBC e qBCProd sao EXCLUDENTES no leiaute: a tributacao e por aliquota ou
// por quantidade. Um gravador que emita os dois sempre produz .ini que a
// SEFAZ recusa, e o round-trip nao acusa porque o leitor aceita os dois.
func TestGravarINI_PISeCOFINSEscolhemOsCamposPeloCST(t *testing.T) {
	pis := PIS{VBC: 100, PPIS: 1.65, QBCProd: 10, VAliqProd: 0.5, VPIS: 5}

	casos := []struct {
		cst       pcn.CSTPis
		presentes []string
		ausentes  []string
	}{
		{pcn.Pis01, []string{"vBC", "pPIS"}, []string{"qBCProd", "vAliqProd"}},
		{pcn.Pis02, []string{"vBC", "pPIS"}, []string{"qBCProd", "vAliqProd"}},
		{pcn.Pis03, []string{"qBCProd", "vAliqProd"}, []string{"vBC", "pPIS"}},
		{pcn.Pis99, []string{"vBC", "pPIS", "qBCProd", "vAliqProd"}, nil},
		// CST fora dos quatro: so o CST sai. O `case` do Delphi nao tem
		// `else`, e replicar isso e o ponto do teste.
		{pcn.Pis04, nil, []string{"vBC", "pPIS", "qBCProd", "vAliqProd", "vPIS"}},
	}

	for _, c := range casos {
		t.Run(c.cst.String(), func(t *testing.T) {
			ini := pcn.NovoINI()
			p := pis
			p.CST = c.cst
			gravarINIImpostoPIS(ini, 0, p)

			const secao = "PIS001"
			if ini.LerString(secao, "CST", "") != c.cst.String() {
				t.Error("o CST sai sempre")
			}
			for _, chave := range c.presentes {
				if ini.LerString(secao, chave, "") == "" {
					t.Errorf("CST %s: a chave %q deveria sair", c.cst, chave)
				}
			}
			for _, chave := range c.ausentes {
				if ini.LerString(secao, chave, "AUSENTE") != "AUSENTE" {
					t.Errorf("CST %s: a chave %q NAO deveria sair", c.cst, chave)
				}
			}
		})
	}
}

// TestGravarINI_IPIDependeDoEnquadramentoNaoDoCST trava uma condicional
// facil de trocar: o grupo de IPI nao sai sem `cEnq`, mesmo com CST e
// valores preenchidos.
func TestGravarINI_IPIDependeDoEnquadramentoNaoDoCST(t *testing.T) {
	ipi := IPI{CST: Ipi00, VBC: 100, PIPI: 10, VIPI: 10}

	ini := pcn.NovoINI()
	gravarINIImpostoIPI(ini, 0, ipi)
	if ini.SecaoExiste("IPI001") {
		t.Error("sem cEnq o grupo de IPI nao sai, por mais preenchido que esteja")
	}

	ipi.CEnq = "999"
	ini = pcn.NovoINI()
	gravarINIImpostoIPI(ini, 0, ipi)
	if !ini.SecaoExiste("IPI001") {
		t.Fatal("com cEnq o grupo sai")
	}
	if ini.LerFloat("IPI001", "vIPI", 0) != 10 {
		t.Error("o grupo sai completo")
	}
}

// TestGravarINI_TrocoVaiNaULTIMASecaoDePagamento trava a posicao do vTroco,
// que e da COLECAO e nao do item.
//
// No original a gravacao fica FORA do laco e usa a variavel de secao que
// sobrou da ultima volta. Com a lista vazia essa variavel nunca foi
// atribuida -- aqui o troco simplesmente nao sai, em vez de ir para uma
// secao de nome indefinido.
func TestGravarINI_TrocoVaiNaULTIMASecaoDePagamento(t *testing.T) {
	p := Pag{VTroco: 7.5}
	for i := 0; i < 3; i++ {
		item := NovoPagItem()
		item.TPag = FpDinheiro
		item.VPag = 10
		p.DetPag = append(p.DetPag, item)
	}

	ini := pcn.NovoINI()
	gravarINIPagamento(ini, p)

	if ini.LerString("pag001", "vTroco", "AUSENTE") != "AUSENTE" {
		t.Error("o troco nao sai na primeira secao")
	}
	if ini.LerFloat("pag003", "vTroco", 0) != 7.5 {
		t.Error("o troco sai na ULTIMA secao de pagamento")
	}

	// Sem pagamento nenhum, o troco nao vai para lugar nenhum.
	ini = pcn.NovoINI()
	gravarINIPagamento(ini, Pag{VTroco: 7.5})
	for _, s := range ini.Secoes() {
		t.Errorf("sem pagamento nao deveria haver secao, veio %q", s)
	}
}

// TestGravarINI_GruposCondicionadosAUmCampo trava, num teste so, as sete
// condicionais de grupo inteiro do gravador.
//
// Todas somem em silencio quando quebradas: o grupo simplesmente deixa de
// sair, e nenhum build, vet ou round-trip de nota preenchida percebe.
func TestGravarINI_GruposCondicionadosAUmCampo(t *testing.T) {
	casos := []struct {
		descricao string
		secao     string
		preparar  func(*NFe)
	}{
		{"Avulsa sem CNPJ", "Avulsa", func(n *NFe) {
			n.Avulsa = Avulsa{XOrgao: "SEFAZ", NDAR: "1"}
		}},
		{"Entrega sem CNPJCPF", "Entrega", func(n *NFe) {
			n.Entrega = Entrega{XLgr: "RUA X", UF: "MG"}
		}},
		{"Veiculo sem chassi", "Veiculo001", func(n *NFe) {
			n.Det[0].Prod.VeicProd = VeicProd{CCor: "1", XCor: "PRETO"}
		}},
		{"Combustivel sem cProdANP", "Combustivel001", func(n *NFe) {
			n.Det[0].Prod.Comb = Comb{DescANP: "GASOLINA"}
		}},
		{"II com base zerada", "II001", func(n *NFe) {
			n.Det[0].Imposto.II = II{VII: 10}
		}},
		{"ISSQN com base zerada", "ISSQN001", func(n *NFe) {
			n.Det[0].Imposto.ISSQN = ISSQN{CListServ: "0101", VISSQN: 5}
		}},
		{"Compra sem xNEmp", "Compra", func(n *NFe) {
			n.Compra = Compra{XPed: "PED1", XCont: "CONT1"}
		}},
		{"guiaTransito sem tipo", "guiaTransito", func(n *NFe) {
			n.Agropecuario.GuiaTransito = GuiaTransito{UFGuia: "MG", NGuia: "1"}
		}},
		{"procNFe sem protocolo", "procNFe", func(n *NFe) {
			n.ProcNFe = pcn.ProcDFe{ChDFe: chaveDeMinas, CStat: 100}
		}},
	}

	for _, c := range casos {
		t.Run(c.descricao, func(t *testing.T) {
			n := notaDeTeste(t)
			c.preparar(n)

			ini, err := MontarINI(n)
			if err != nil {
				t.Fatal(err)
			}
			if ini.SecaoExiste(c.secao) {
				t.Errorf("a secao %q nao deveria sair: falta o campo que a condiciona",
					c.secao)
			}
		})
	}
}

// TestGravarINI_ChaveInvalidaEhErro trava a recusa do original.
func TestGravarINI_ChaveInvalidaEhErro(t *testing.T) {
	n := notaDeTeste(t)
	n.InfNFe.ID = "NFe123"

	if _, err := GravarINI(n); err == nil {
		t.Error("chave invalida deveria impedir a geracao do .ini")
	}
}

// TestINI_IBSCBSDoItemUsaOModeloDaNOTA trava a ligacao entre o .ini da NF-e
// e a tabela de GrupoDoCST do rtc.
//
// O gravador do IBSCBS precisa saber o modelo do documento, e esse modelo
// sai do `mod` da NOTA -- nao da configuracao do componente, que pode nem
// existir quando o .ini e gerado a partir de uma nota ja montada.
func TestINI_IBSCBSDoItemUsaOModeloDaNOTA(t *testing.T) {
	casos := []struct {
		modelo   int
		cst      rtc.CSTIBSCBS
		esperada string
		proibida string
	}{
		// O CST 510 existe na NF-e e NAO na NFC-e.
		{55, rtc.CST510, "gIBSCBS001", ""},
		{65, rtc.CST510, "", "gIBSCBS001"},
		// O 620 leva a monofasia nos dois.
		{55, rtc.CST620, "gIBSCBSMono001", "gIBSCBS001"},
		{65, rtc.CST620, "gIBSCBSMono001", "gIBSCBS001"},
	}

	for _, c := range casos {
		t.Run(c.cst.String(), func(t *testing.T) {
			n := notaDeTeste(t)
			n.Ide.Modelo = c.modelo
			n.Det[0].Imposto.IBSCBS = rtc.IBSCBS{
				CST: c.cst, CClassTrib: "000001",
				GIBSCBS:     rtc.GIBSCBS{VBC: 100, VIBS: 10},
				GIBSCBSMono: rtc.GIBSCBSMono{VTotIBSMonoItem: 5},
			}

			ini, err := MontarINI(n)
			if err != nil {
				t.Fatal(err)
			}
			if c.esperada != "" && !ini.SecaoExiste(c.esperada) {
				t.Errorf("modelo %d, CST %s: %q deveria sair",
					c.modelo, c.cst, c.esperada)
			}
			if c.proibida != "" && ini.SecaoExiste(c.proibida) {
				t.Errorf("modelo %d, CST %s: %q NAO deveria sair",
					c.modelo, c.cst, c.proibida)
			}
		})
	}
}
