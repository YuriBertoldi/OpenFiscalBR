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
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Leitura do formato .ini do ACBr. Porte de TNFeIniReader
// (ACBrNFe.IniReader.pas).
//
// O .ini e a forma como a maioria dos ERP alimenta o ACBr: mais facil de
// gerar que XML e sem namespace. O formato nao tem schema publicado -- o
// contrato e o proprio leitor --, e tres convencoes dele escapam de quem o
// porta pela primeira vez:
//
//  1. NOME DE SECAO COM ALIAS. As secoes principais aceitam dois nomes, um
//     deles em portugues por extenso, e o leitor escolhe pelo que EXISTIR
//     no arquivo: [Identificacao] ou [ide], [Emitente] ou [emit],
//     [Destinatario] ou [dest], [Produto] ou [detNNN]. Suportar so um dos
//     dois recusa metade dos arquivos em circulacao;
//
//  2. NOME DE CHAVE COM ALIAS. Dentro da secao, muitas chaves tem um
//     sinonimo historico lido como PADRAO da primeira: `cNF`/`Codigo`,
//     `natOp`/`NaturezaOperacao`, `mod`/`Modelo`, `nNF`/`Numero`,
//     `dhEmi`/`dEmi`/`Emissao`, `xLgr`/`Logradouro`, `cMun`/`CidadeCod`.
//     A cadeia e aninhada, e a ORDEM importa;
//
//  3. GRUPO REPETIDO TERMINA POR SENTINELA. Nao ha contador: o leitor tenta
//     [NFref001], [NFref002]... e para quando a chave sentinela daquele
//     grupo nao existe. O valor "FIM" e o default que sinaliza o fim, e
//     string vazia tambem encerra.
//
// Os indices de secao sao base 1 e tem largura fixa, que varia por grupo:
// tres digitos na maioria ([detNNN], [NFrefNNN]), dois em alguns
// ([gPagAntecipadoNN]), e ha grupo que aceita as DUAS larguras ([autXML]).

// LerINI le uma NFe a partir do conteudo de um arquivo .ini.
//
// O parametro aceita tanto o caminho de um arquivo quanto o proprio
// conteudo, como LerIniArquivoOuString do ACBr.
func LerINI(iniOuCaminho string, cfg Configuracoes) (*NFe, error) {
	ini, err := pcn.LerINIArquivoOuString(iniOuCaminho)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIReader le uma NFe a partir de um io.Reader com conteudo .ini.
func LerINIReader(r io.Reader, cfg Configuracoes) (*NFe, error) {
	ini, err := pcn.LerINI(r)
	if err != nil {
		return nil, err
	}
	return lerINIDocumento(ini, cfg)
}

// LerINIArquivo le uma NFe de um arquivo .ini.
func LerINIArquivo(caminho string, cfg Configuracoes) (*NFe, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return nil, fmt.Errorf("nfe: abrir %q: %w", caminho, err)
	}
	defer f.Close()
	return LerINIReader(f, cfg)
}

// lerINIDocumento e o porte de TNFeIniReader.LerIni.
//
// A ORDEM das chamadas e a do original e NAO e livre: o emitente grava
// Ide.CUF e pode completar Ide.CMunFG, entao precisa rodar depois da
// identificacao. Ver lerINIEmitente.
func lerINIDocumento(ini *pcn.INI, cfg Configuracoes) (*NFe, error) {
	if ini == nil {
		return nil, ErrXMLVazio
	}
	n := NovaNFe()

	n.InfNFe.Versao = ini.LerFloat("infNFe", "versao", cfg.VersaoDF.Float())

	lerINIIdentificacao(ini, &n.Ide, cfg)
	lerININFReferencia(ini, &n.Ide)
	lerINIEmitente(ini, &n)
	lerINIAvulsa(ini, &n.Avulsa)
	lerINIDestinatario(ini, &n.Dest)
	lerINIRetirada(ini, &n.Retirada)
	lerINIEntrega(ini, &n.Entrega)
	lerINIAutorizadosXML(ini, &n)
	lerINIProdutos(ini, &n)
	lerINITotal(ini, &n.Total)
	lerINITransportador(ini, &n.Transp)
	lerINICobranca(ini, &n.Cobr)
	lerINIPagamento(ini, &n.Pag)
	lerINIInfIntermed(ini, &n.InfIntermed)
	lerINIDadosAdicionais(ini, &n.InfAdic)
	lerINIExportacao(ini, &n.Exporta)
	lerINICompra(ini, &n.Compra)
	lerINICana(ini, &n.Cana)
	lerINIInfNFeSupl(ini, &n.InfNFeSupl)
	lerINIInfRespTec(ini, &n.InfRespTec)
	lerINIAgropecuario(ini, &n.Agropecuario)

	return &n, nil
}

// secaoAlias devolve o primeiro nome de secao que EXISTE no arquivo, ou o
// ultimo da lista quando nenhum existe.
//
// Porte do `IfThen(AINIRec.SectionExists('Identificacao'), 'Identificacao',
// 'ide')` que abre quase toda procedure do original. Devolver o ultimo no
// caso de nenhuma existir preserva o comportamento: as leituras seguintes
// caem nos defaults, sem erro.
func secaoAlias(ini *pcn.INI, nomes ...string) string {
	for _, nome := range nomes {
		if ini.SecaoExiste(nome) {
			return nome
		}
	}
	return nomes[len(nomes)-1]
}

// lerStringAlias le a primeira chave preenchida da lista.
//
// Porte da cadeia aninhada `ReadString(s,'cNF', ReadString(s,'Codigo',”))`.
// A forma aninhada do Delphi significa: use a primeira; faltando ela, a
// segunda; e assim por diante. "Faltando" ali e a chave NAO EXISTIR -- uma
// chave presente e vazia vence as seguintes, e isso esta preservado.
func lerStringAlias(ini *pcn.INI, secao string, chaves ...string) string {
	for _, chave := range chaves {
		if v := ini.LerString(secao, chave, "\x00"); v != "\x00" {
			return v
		}
	}
	return ""
}

// lerInteiroAlias le a cadeia aninhada de ReadInteger.
//
// Tem semantica DIFERENTE de lerStringAlias, e a diferenca vem do Delphi:
// o TCustomIniFile.ReadInteger e
//
//	StrToIntDef(ReadString(Secao, Chave, ''), Default)
//
// e numa cadeia `ReadInteger(a, ReadInteger(b, 0))` o `Default` do externo e
// o VALOR do interno. Chave presente e VAZIA faz o StrToIntDef falhar e cair
// para o default -- ou seja, para o alias. O mesmo vale para conteudo nao
// numerico.
//
// Por isso a avaliacao aqui e de TRAS PARA A FRENTE: o resultado comeca no
// padrao e cada chave que CONVERTE o sobrescreve, de modo que a primeira da
// lista vence quando converte, e cede para a seguinte quando nao.
//
// Um `[ide]` com `cNF=` vazio e `Codigo=12345` devolve 12345, nao zero.
func lerInteiroAlias(ini *pcn.INI, secao string, padrao int, chaves ...string) int {
	valor := padrao
	for i := len(chaves) - 1; i >= 0; i-- {
		v := strings.TrimSpace(ini.LerString(secao, chaves[i], ""))
		if n, err := strconv.Atoi(v); err == nil {
			valor = n
		}
	}
	return valor
}

// tempoZero e a data ausente -- o `StringToDateTime('0')` do original.
func tempoZero() time.Time { return time.Time{} }

// lerDataAlias le a primeira chave de data preenchida da lista.
//
// A cadeia `dhEmi` -> `dEmi` -> `Emissao` existe porque a tag mudou de nome
// na 4.00 (ganhou hora) e o nome em portugues e mais antigo que os dois.
func lerDataAlias(ini *pcn.INI, secao string, chaves ...string) time.Time {
	for _, chave := range chaves {
		if v := ini.LerData(secao, chave, tempoZero()); !v.IsZero() {
			return v
		}
	}
	return tempoZero()
}

// ---------------------------------------------------------------------------
// Identificacao
// ---------------------------------------------------------------------------

// lerINIIdentificacao le [Identificacao] ou [ide].
// Porte de Ler_Identificacao.
//
// O modelo cai para 65 (NFC-e) quando ausente, e NAO para 55 -- e o default
// do original (:205).
func lerINIIdentificacao(ini *pcn.INI, ide *Ide, cfg Configuracoes) {
	secao := secaoAlias(ini, "Identificacao", "ide")

	ide.CNF = lerInteiroAlias(ini, secao, 0, "cNF", "Codigo")
	ide.NatOp = lerStringAlias(ini, secao, "natOp", "NaturezaOperacao")
	ide.IndPag, _ = ParseIndicadorPagamento(primeiroNaoVazio(
		lerStringAlias(ini, secao, "indPag", "FormaPag"), "0"))
	ide.Modelo = lerInteiroAlias(ini, secao, 65, "mod", "Modelo")
	ide.Serie = ini.LerInteiro(secao, "Serie", 1)
	ide.NNF = lerInteiroAlias(ini, secao, 0, "nNF", "Numero")
	ide.DEmi = lerDataAlias(ini, secao, "dhEmi", "dEmi", "Emissao")
	ide.DSaiEnt = lerDataAlias(ini, secao, "dhSaiEnt", "dSaiEnt", "Saida")
	ide.HSaiEnt = ini.LerData(secao, "hSaiEnt", tempoZero())
	ide.DPrevEntrega = ini.LerData(secao, "dPrevEntrega", tempoZero())
	ide.TpNF, _ = pcn.ParseTipoNFe(primeiroNaoVazio(
		lerStringAlias(ini, secao, "tpNF", "Tipo"), "1"))
	ide.IdDest, _ = ParseDestinoOperacao(ini.LerString(secao, "idDest", "1"))
	ide.TpImp, _ = pcn.ParseTipoImpressao(ini.LerString(secao, "tpImp", "1"))
	ide.TpEmis, _ = pcn.ParseTipoEmissao(ini.LerString(secao, "tpEmis",
		strconv.Itoa(cfg.TpEmis)))
	ide.TpAmb, _ = pcn.ParseTipoAmbiente(ini.LerString(secao, "tpAmb",
		cfg.Ambiente.String()))
	ide.FinNFe, _ = ParseFinalidadeNFe(primeiroNaoVazio(
		lerStringAlias(ini, secao, "finNFe", "Finalidade"), "0"))
	ide.IndFinal, _ = ParseConsumidorFinal(ini.LerString(secao, "indFinal", "0"))
	ide.IndPres, _ = ParsePresencaComprador(ini.LerString(secao, "indPres", "0"))
	ide.IndIntermed, _ = ParseIndIntermed(ini.LerString(secao, "indIntermed", ""))
	ide.ProcEmi, _ = pcn.ParseProcessoEmissao(ini.LerString(secao, "procEmi", "0"))
	ide.VerProc = ini.LerString(secao, "verProc", "ACBrNFe")
	ide.DhCont = ini.LerData(secao, "dhCont", tempoZero())
	ide.XJust = ini.LerString(secao, "xJust", "")
	ide.CMunFG = ini.LerInteiro(secao, "cMunFG", 0)
	ide.CIndOp = ini.LerString(secao, "cIndOp", "")

	// Reforma Tributaria
	ide.CMunFGIBS = ini.LerInteiro(secao, "cMunFGIBS", 0)
	ide.TpNFDebito, _ = ParseTpNFDebito(ini.LerString(secao, "tpNFDebito", ""))
	ide.TpNFCredito, _ = ParseTpNFCredito(ini.LerString(secao, "tpNFCredito", ""))

	rtc.LerINIGCompraGov(ini, &ide.GCompraGov)
	rtc.LerINIGPagAntecipado(ini, &ide.GPagAntecipado)
}

// primeiroNaoVazio devolve o primeiro valor nao vazio.
//
// Existe porque o ACBr passa o DEFAULT para dentro do ReadString ('1', '0'),
// e o alias ja consumiu essa posicao. Sem isso, um alias ausente devolveria
// string vazia e o Parse cairia no membro errado.
func primeiroNaoVazio(valores ...string) string {
	for _, v := range valores {
		if v != "" {
			return v
		}
	}
	return ""
}

// lerININFReferencia le [NFrefNNN]. Porte de Ler_NFReferencia.
//
// E a unica secao do .ini cujo TIPO e inferido quando nao declarado, e a
// ordem da inferencia e contrato: refNFe/refNFeSig -> NFE, refCTe -> CTE,
// nECF -> ECF, IE -> NFP, CNPJ -> NF. Sem nenhuma dessas, o laco TERMINA --
// nao e "ignore e siga".
//
// A ordem importa porque as chaves se sobrepoem: uma NF de produtor rural
// ([NFP]) tambem tem CNPJ, e se o teste de CNPJ viesse antes do de IE toda
// NFP seria lida como NF comum.
func lerININFReferencia(ini *pcn.INI, ide *Ide) {
	ide.NFref = nil
	for i := 1; ; i++ {
		secao := fmt.Sprintf("NFref%03d", i)

		tipo := strings.ToUpper(ini.LerString(secao, "Tipo", "NFe"))
		if declarado := ini.LerString(secao, "Tipo", "FIM"); declarado == "FIM" || declarado == "" {
			var ok bool
			if tipo, ok = tipoNFrefInferido(ini, secao); !ok {
				break
			}
		}

		var ref NFref
		switch tipo {
		// O "SAT" cai no mesmo ramo do "NFE": o cupom do SAT e referenciado
		// pela chave de acesso dele, no mesmo campo (:265).
		case "NFE", "SAT":
			ref.RefNFe = ini.LerString(secao, "refNFe", "")
			ref.RefNFeSig = ini.LerString(secao, "refNFeSig", "")
		case "NF":
			ref.RefNF.CUF = ini.LerInteiro(secao, "cUF", 0)
			ref.RefNF.AAMM = ini.LerString(secao, "AAMM", "")
			ref.RefNF.CNPJ = ini.LerString(secao, "CNPJ", "")
			ref.RefNF.Modelo = lerInteiroAlias(ini, secao, 0, "mod", "Modelo")
			ref.RefNF.Serie = ini.LerInteiro(secao, "Serie", 0)
			ref.RefNF.NNF = ini.LerInteiro(secao, "nNF", 0)
		case "NFP":
			ref.RefNFP.CUF = ini.LerInteiro(secao, "cUF", 0)
			ref.RefNFP.AAMM = ini.LerString(secao, "AAMM", "")
			ref.RefNFP.CNPJCPF = lerStringAlias(ini, secao, "CNPJ", "CPF", "CNPJCPF")
			ref.RefNFP.IE = ini.LerString(secao, "IE", "")
			// O modelo da NFP e STRING, ao contrario do da NF (ver RefNFP
			// em classes.go).
			ref.RefNFP.Modelo = lerStringAlias(ini, secao, "mod", "Modelo")
			ref.RefNFP.Serie = ini.LerInteiro(secao, "Serie", 0)
			ref.RefNFP.NNF = ini.LerInteiro(secao, "nNF", 0)
		case "CTE":
			ref.RefCTe = ini.LerString(secao, "refCTe", "")
		case "ECF":
			ref.RefECF.Modelo, _ = ParseECFModRef(lerStringAlias(ini, secao, "mod", "ModECF"))
			ref.RefECF.NECF = ini.LerString(secao, "nECF", "")
			ref.RefECF.NCOO = ini.LerString(secao, "nCOO", "")
		}

		// Tipo declarado mas desconhecido entra como item VAZIO, e o laco
		// segue -- e o que o `with NFref.New do` do original faz, porque o
		// item e criado ANTES do encadeamento de `if` (:264).
		ide.NFref = append(ide.NFref, ref)
	}
}

// tipoNFrefInferido descobre o tipo pela chave presente, quando a chave
// `Tipo` nao foi declarada. Devolve false quando nenhuma casa, o que
// ENCERRA o laco.
func tipoNFrefInferido(ini *pcn.INI, secao string) (string, bool) {
	switch {
	case ini.LerString(secao, "refNFe", "") != "",
		ini.LerString(secao, "refNFeSig", "") != "":
		return "NFE", true
	case ini.LerString(secao, "refCTe", "") != "":
		return "CTE", true
	case ini.LerString(secao, "nECF", "") != "":
		return "ECF", true
	case ini.LerString(secao, "IE", "") != "":
		return "NFP", true
	case ini.LerString(secao, "CNPJ", "") != "":
		return "NF", true
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Emitente, avulsa e destinatario
// ---------------------------------------------------------------------------

// lerINIEmitente le [Emitente] ou [emit]. Porte de Ler_Emitente.
//
// Recebe a NOTA, e nao so o emitente, porque o original escreve em DOIS
// grupos: alem do emit, ele completa a identificacao (:309-310).
//
//   - Ide.CUF cai para o codigo da UF do emitente quando a chave `cUF` nao
//     esta no arquivo. Sem isso, toda nota sem cUF explicito sairia com
//     cUF zero e chave de acesso invalida;
//   - Ide.CMunFG so e preenchido aqui se ainda estiver ZERO, caindo para o
//     cMun do emitente. Por isso a identificacao tem de ser lida ANTES.
func lerINIEmitente(ini *pcn.INI, n *NFe) {
	secao := secaoAlias(ini, "Emitente", "emit")
	e := &n.Emit

	e.CNPJCPF = lerStringAlias(ini, secao, "CNPJ", "CNPJCPF", "CPF")
	e.XNome = lerStringAlias(ini, secao, "xNome", "Razao")
	e.XFant = lerStringAlias(ini, secao, "xFant", "Fantasia")
	e.IE = ini.LerString(secao, "IE", "")
	e.IEST = ini.LerString(secao, "IEST", "")
	e.IM = ini.LerString(secao, "IM", "")
	e.CNAE = ini.LerString(secao, "CNAE", "")
	e.CRT, _ = ParseCRT(ini.LerString(secao, "CRT", "3"))
	e.ISUFEmit = ini.LerString(secao, "ISUFEmit", "")

	end := &e.EnderEmit
	end.XLgr = lerStringAlias(ini, secao, "xLgr", "Logradouro")

	// nro e xCpl so sao ATRIBUIDOS se algum dos dois aliases estiver
	// preenchido (:299-303). A diferenca aparece no construtor: NovaNFe nao
	// poe nada nesses dois, mas a condicao existe no original e esta
	// preservada -- um .ini com `nro=` vazio nao apaga o que ja houvesse.
	if v := lerStringAlias(ini, secao, "nro", "Numero"); v != "" {
		end.Nro = v
	}
	if v := lerStringAlias(ini, secao, "xCpl", "Complemento"); v != "" {
		end.XCpl = v
	}

	end.XBairro = lerStringAlias(ini, secao, "xBairro", "Bairro")
	end.CMun = lerInteiroAlias(ini, secao, 0, "cMun", "CidadeCod")
	end.XMun = lerStringAlias(ini, secao, "xMun", "Cidade")
	end.UF = ini.LerString(secao, "UF", "")
	end.CEP = ini.LerInteiro(secao, "CEP", 0)
	end.CPais = lerInteiroAlias(ini, secao, 1058, "cPais", "PaisCod")
	end.XPais = primeiroNaoVazio(lerStringAlias(ini, secao, "xPais", "Pais"), "BRASIL")
	end.Fone = ini.LerString(secao, "Fone", "")

	n.Ide.CUF = ini.LerInteiro(secao, "cUF", pcn.CodigoUF(end.UF))

	if n.Ide.CMunFG == 0 {
		n.Ide.CMunFG = lerInteiroAlias(ini, secao, end.CMun, "cMunFG", "CidadeCod")
	}
}

// lerINIAvulsa le [Avulsa]. Porte de Ler_Avulsa.
//
// O grupo inteiro e condicionado ao CNPJ estar preenchido -- e o fisco que
// emite a nota avulsa, e sem o CNPJ dele o grupo nao existe.
func lerINIAvulsa(ini *pcn.INI, a *Avulsa) {
	const secao = "Avulsa"
	if ini.LerString(secao, "CNPJ", "") == "" {
		return
	}
	a.CNPJ = ini.LerString(secao, "CNPJ", "")
	a.XOrgao = ini.LerString(secao, "xOrgao", "")
	a.Matr = ini.LerString(secao, "matr", "")
	a.XAgente = ini.LerString(secao, "xAgente", "")
	a.Fone = ini.LerString(secao, "fone", "")
	a.UF = ini.LerString(secao, "UF", "")
	a.NDAR = ini.LerString(secao, "nDAR", "")
	a.DEmi = ini.LerData(secao, "dEmi", tempoZero())
	a.VDAR = ini.LerFloat(secao, "vDAR", 0)
	a.RepEmi = ini.LerString(secao, "repEmi", "")
	a.DPag = ini.LerData(secao, "dPag", tempoZero())
}

// lerINIDestinatario le [Destinatario] ou [dest]. Porte de Ler_Destinatario.
//
// Diferente do emitente, o destinatario NAO tem default de cPais/xPais no
// construtor do grupo -- mas o .ini traz o mesmo 1058/BRASIL, porque e o
// default da leitura.
func lerINIDestinatario(ini *pcn.INI, d *Dest) {
	secao := secaoAlias(ini, "Destinatario", "dest")

	d.IDEstrangeiro = ini.LerString(secao, "idEstrangeiro", "")
	d.CNPJCPF = lerStringAlias(ini, secao, "CNPJ", "CNPJCPF", "CPF")
	d.XNome = lerStringAlias(ini, secao, "xNome", "NomeRazao")
	d.IndIEDest, _ = pcn.ParseIndIEDest(ini.LerString(secao, "indIEDest", "1"))
	d.IE = ini.LerString(secao, "IE", "")
	d.IM = ini.LerString(secao, "IM", "")
	d.ISUF = ini.LerString(secao, "ISUF", "")
	d.Email = ini.LerString(secao, "Email", "")

	end := &d.EnderDest
	end.XLgr = lerStringAlias(ini, secao, "xLgr", "Logradouro")
	if v := lerStringAlias(ini, secao, "nro", "Numero"); v != "" {
		end.Nro = v
	}
	if v := lerStringAlias(ini, secao, "xCpl", "Complemento"); v != "" {
		end.XCpl = v
	}
	end.XBairro = lerStringAlias(ini, secao, "xBairro", "Bairro")
	end.CMun = lerInteiroAlias(ini, secao, 0, "cMun", "CidadeCod")
	end.XMun = lerStringAlias(ini, secao, "xMun", "Cidade")
	end.UF = ini.LerString(secao, "UF", "")
	end.CEP = ini.LerInteiro(secao, "CEP", 0)
	end.CPais = lerInteiroAlias(ini, secao, 1058, "cPais", "PaisCod")
	end.XPais = primeiroNaoVazio(lerStringAlias(ini, secao, "xPais", "Pais"), "BRASIL")
	end.Fone = ini.LerString(secao, "Fone", "")
}

// lerINIRetirada le [Retirada]. Porte de Ler_Retirada.
//
// Condicionado ao LOGRADOURO, nao ao CNPJ: o local de retirada pode ser um
// endereco sem documento associado.
func lerINIRetirada(ini *pcn.INI, r *Retirada) {
	const secao = "Retirada"
	logradouro := ini.LerString(secao, "xLgr", "")
	if logradouro == "" {
		return
	}
	r.CNPJCPF = lerStringAlias(ini, secao, "CNPJ", "CPF", "CNPJCPF")
	r.XNome = ini.LerString(secao, "xNome", "")
	r.XLgr = logradouro
	r.Nro = ini.LerString(secao, "nro", "")
	r.XCpl = ini.LerString(secao, "xCpl", "")
	r.XBairro = ini.LerString(secao, "xBairro", "")
	r.CMun = ini.LerInteiro(secao, "cMun", 0)
	r.XMun = ini.LerString(secao, "xMun", "")
	r.UF = ini.LerString(secao, "UF", "")
	r.CEP = ini.LerInteiro(secao, "CEP", 0)
	r.CPais = lerInteiroAlias(ini, secao, 1058, "cPais", "PaisCod")
	r.XPais = primeiroNaoVazio(lerStringAlias(ini, secao, "xPais", "Pais"), "BRASIL")
	r.Fone = ini.LerString(secao, "Fone", "")
	r.Email = ini.LerString(secao, "Email", "")
	r.IE = ini.LerString(secao, "IE", "")
}

// lerINIEntrega le [Entrega]. Porte de Ler_Entrega.
func lerINIEntrega(ini *pcn.INI, e *Entrega) {
	const secao = "Entrega"
	logradouro := ini.LerString(secao, "xLgr", "")
	if logradouro == "" {
		return
	}
	e.CNPJCPF = lerStringAlias(ini, secao, "CNPJ", "CPF", "CNPJCPF")
	e.XNome = ini.LerString(secao, "xNome", "")
	e.XLgr = logradouro
	e.Nro = ini.LerString(secao, "nro", "")
	e.XCpl = ini.LerString(secao, "xCpl", "")
	e.XBairro = ini.LerString(secao, "xBairro", "")
	e.CMun = ini.LerInteiro(secao, "cMun", 0)
	e.XMun = ini.LerString(secao, "xMun", "")
	e.UF = ini.LerString(secao, "UF", "")
	e.CEP = ini.LerInteiro(secao, "CEP", 0)
	e.CPais = lerInteiroAlias(ini, secao, 1058, "cPais", "PaisCod")
	e.XPais = primeiroNaoVazio(lerStringAlias(ini, secao, "xPais", "Pais"), "BRASIL")
	e.Fone = ini.LerString(secao, "Fone", "")
	e.Email = ini.LerString(secao, "Email", "")
	e.IE = ini.LerString(secao, "IE", "")
}

// lerINIAutorizadosXML le [autXMLNNN]. Porte de Ler_AutorizadosXML.
//
// E o unico grupo que tenta DUAS larguras de indice: primeiro tres digitos
// ([autXML001]) e, nao achando, dois ([autXML01]). O ACBr escreve com tres,
// mas arquivos antigos usam dois.
//
// O documento passa por OnlyCPFCNPJAlphaNum antes do teste de fim, o que
// significa que uma secao com CNPJ so de pontuacao encerra o laco.
func lerINIAutorizadosXML(ini *pcn.INI, n *NFe) {
	n.AutXML = nil
	for i := 1; ; i++ {
		doc := documentoAutorizado(ini, fmt.Sprintf("autXML%03d", i))
		if doc == "" {
			doc = documentoAutorizado(ini, fmt.Sprintf("autXML%02d", i))
		}
		if doc == "" {
			break
		}
		n.AutXML = append(n.AutXML, AutXML{CNPJCPF: doc})
	}
}

// documentoAutorizado le o CNPJ/CPF de uma secao de autXML, ja filtrado.
// Devolve string vazia quando a secao nao tem documento.
func documentoAutorizado(ini *pcn.INI, secao string) string {
	v := lerStringAlias(ini, secao, "CNPJ", "CPF", "CNPJCPF")
	if v == "" || v == "FIM" {
		return ""
	}
	return pcn.OnlyCPFCNPJAlphaNum(v)
}
