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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
)

// Regras de negocio da NF-e e da NFC-e.
//
// Porte de TNFeValidarRegras (ACBrNFe.ValidarRegrasdeNegocio.pas), as 155
// rejeicoes ATIVAS que a SEFAZ aplicaria e que o ACBr antecipa localmente.
//
// Valem tres observacoes que mudam como o codigo abaixo se le:
//
//  1. Elas rodam sobre o documento PRONTO, depois do gerador. Varias
//     comparam o que esta no XML com o que esta na CONFIGURACAO -- o
//     ambiente da nota contra o do web service, a UF do emitente contra a
//     UF autorizadora. Sem a configuracao, metade delas nao existe.
//  2. O "agora" entra por parametro, nao de time.Now(). Oito regras comparam
//     datas com o momento da transmissao, e uma validacao que muda de
//     resultado conforme o relogio nao e testavel.
//  3. O ACBr acumula os erros numa STRING e devolve um booleano. Aqui cada
//     rejeicao e um valor, com o codigo separado da mensagem, porque e isso
//     que o consumidor precisa para decidir o que fazer com cada uma.
//
// # Regras que o ACBr mantem COMENTADAS
//
// Quatorze rejeicoes estao no arquivo mas desativadas, e por isso NAO foram
// portadas. A lista existe para que a proxima leitura nao as confunda com
// esquecimento -- e porque algumas voltarao:
//
//	502           concatenacao da chave; o proprio ValidarRegra512 e um
//	              procedimento de corpo vazio
//	508, 529, 600 CST/CSOSN incompativel com o tipo de contribuinte
//	571           tipo de emissao diferente de 3 para contingencia SCAN,
//	              com a nota "Nao existe mais essa regra de validacao"
//	629, 630      valor do produto contra quantidade vezes valor unitario
//	719, 720      identificacao do destinatario e idEstrangeiro no exterior
//	806           ICMS-ST sem CEST
//	856, 858      valor de partida e grupo ICMS60 do GLP
//	883, 888      cEAN e cEANTrib sem informacao -- so valeriam em producao
//	              a partir de 01/12/2018, e o ACBr nunca as reativou
//
// ferramentas/conferir-regras-dfe.py compara os dois conjuntos e reprova se
// uma regra ATIVA ficar sem contraparte aqui.

// ErroRegraNegocio e uma rejeicao das regras de negocio, com o codigo
// oficial da SEFAZ.
type ErroRegraNegocio struct {
	// Codigo e o numero da rejeicao, ex.: 226.
	Codigo int

	// Mensagem e o texto da rejeicao, sem o codigo.
	Mensagem string

	// Item e o nItem a que a rejeicao se refere, ou zero quando a regra vale
	// para o documento inteiro. O ACBr costura isso no texto; aqui e campo,
	// para quem consome poder destacar a linha na tela.
	Item int
}

func (e *ErroRegraNegocio) Error() string {
	if e.Item > 0 {
		return fmt.Sprintf("%d-%s [nItem: %d]", e.Codigo, e.Mensagem, e.Item)
	}
	return fmt.Sprintf("%d-%s", e.Codigo, e.Mensagem)
}

// coletor acumula as rejeicoes enquanto as regras rodam.
type coletor struct {
	n     *NFe
	cfg   Configuracoes
	agora time.Time
	erros []*ErroRegraNegocio
}

func (c *coletor) erro(codigo int, mensagem string) {
	c.erros = append(c.erros, &ErroRegraNegocio{Codigo: codigo, Mensagem: mensagem})
}

func (c *coletor) erroItem(codigo, item int, mensagem string) {
	c.erros = append(c.erros, &ErroRegraNegocio{Codigo: codigo, Mensagem: mensagem, Item: item})
}

// seErro acrescenta a rejeicao quando a condicao e verdadeira. Encurta as
// 155 regras para uma linha cada, que e o que permite le-las lado a lado
// com o .pas.
func (c *coletor) seErro(condicao bool, codigo int, mensagem string) {
	if condicao {
		c.erro(codigo, mensagem)
	}
}

func (c *coletor) seErroItem(condicao bool, codigo, item int, mensagem string) {
	if condicao {
		c.erroItem(codigo, item, mensagem)
	}
}

// ValidarRegrasNegocio roda as regras de negocio sobre o documento.
//
// Porte de TNFeValidarRegras.Validar(Agora). Devolve a lista de rejeicoes;
// vazia significa documento aprovado.
//
// `agora` e o instante que a SEFAZ usaria como referencia -- passe o horario
// da transmissao.
func ValidarRegrasNegocio(n *NFe, cfg Configuracoes, agora time.Time) []*ErroRegraNegocio {
	if n == nil {
		// Codigo zero porque nao e rejeicao da SEFAZ: e erro de uso da
		// biblioteca. A mensagem vem do sentinela do package, para o texto
		// nao divergir do que GerarXML e GerarQRCode devolvem.
		return []*ErroRegraNegocio{{Mensagem: ErrDocumentoNulo.Error()}}
	}

	c := &coletor{n: n, cfg: cfg, agora: agora}

	c.regrasDeAmbosModelos()
	if n.Ide.Modelo == 55 {
		c.regrasSomenteNFe()
	}
	if n.Ide.Modelo == 65 {
		c.regrasSomenteNFCe()
	}

	c.regrasDoCabecalho()
	c.regrasDoEmitenteEDestinatario()
	c.regrasDeLocalETransporte()
	totais := c.regrasDosItens()
	c.regrasDosTotais(totais)
	c.regrasDePagamento(totais)

	return c.erros
}

// ---------------------------------------------------------------------------
// Regras nomeadas, comuns aos dois modelos
// ---------------------------------------------------------------------------

// regrasDeAmbosModelos e o RegrasValidasParaAmbosModelos
// (ValidarRegrasdeNegocio.pas:128).
func (c *coletor) regrasDeAmbosModelos() {
	n := c.n

	// 226: o codigo de UF embutido no municipio do emitente tem de bater com
	// a UF autorizadora CONFIGURADA, nao com a UF declarada no endereco.
	c.seErro(
		!strings.EqualFold(primeirosDoisDigitos(n.Emit.EnderEmit.CMun), strconv.Itoa(c.cfg.CodigoUF)),
		226, "Rejeicao: Codigo da UF do Emitente diverge da UF autorizadora")

	// 228: emissao atrasada mais de 30 dias.
	c.seErro(c.agora.Sub(n.Ide.DEmi) > 30*24*time.Hour,
		228, "Rejeicao: Data de Emissao muito atrasada")

	// O ValidarRegra512 do ACBr e um procedimento de corpo VAZIO: a unica
	// rejeicao que ele teria (502, concatenacao da chave) esta comentada
	// (ValidarRegrasdeNegocio.pas:1277). Nao portado.

	// 701: versao de leiaute abaixo da 3.10 nao e mais aceita.
	c.seErro(n.InfNFe.Versao < 3.10, 701, "Rejeicao: Versao invalida")

	// 703: emissao no futuro, com cinco minutos de folga.
	c.seErro(n.Ide.DEmi.After(c.agora.Add(5*time.Minute)),
		703, "Rejeicao: Data-Hora de Emissao posterior ao horario de recebimento")

	// 897: o codigo numerico nao pode ser igual ao numero da nota nem estar
	// na lista de codigos recusados.
	c.seErro(!validarCodigoDFe(n.Ide.CNF, n.Ide.NNF),
		897, "Rejeicao: Codigo numerico em formato invalido")
}

// ---------------------------------------------------------------------------
// Regras exclusivas da NF-e (modelo 55)
// ---------------------------------------------------------------------------

// regrasSomenteNFe e o RegrasValidasSomenteParaNFe
// (ValidarRegrasdeNegocio.pas:138).
func (c *coletor) regrasSomenteNFe() {
	n := c.n

	// 321: devolucao exige nota referenciada -- a menos que algum item
	// carregue DFe referenciado da Reforma Tributaria.
	if !algumItemTemDFeReferenciado(n) {
		c.seErro(n.Ide.FinNFe == FnDevolucao && len(n.Ide.NFref) == 0,
			321, "Rejeicao: NF-e devolucao nao possui NF referenciada")
	}

	// 504, 505 e 506: coerencia da data de saida.
	c.seErro(n.Ide.DSaiEnt.Sub(c.agora) > 30*24*time.Hour,
		504, "Rejeicao: Data de Entrada/Saida posterior ao permitido")
	c.seErro(!n.Ide.DSaiEnt.IsZero() && c.agora.Sub(n.Ide.DSaiEnt) > 30*24*time.Hour,
		505, "Rejeicao: Data de Entrada/Saida anterior ao permitido")
	c.seErro(!n.Ide.DSaiEnt.IsZero() && n.Ide.DSaiEnt.Before(n.Ide.DEmi),
		506, "Rejeicao: Data de Saida menor que a Data de Emissao")

	// 710: o DANFE da NFC-e nao serve para NF-e.
	c.seErro(n.Ide.TpImp == pcn.TiNFCe || n.Ide.TpImp == pcn.TiMsgEletronica,
		710, "Rejeicao: NF-e com formato de DANFE invalido")

	// 711: contingencia offline so com DANFE simplificado.
	c.seErro(n.Ide.TpEmis == pcn.TeOffLine && n.Ide.TpImp != pcn.TiSimplificadoTipo2,
		711, "Rejeicao: NF-e com contingencia off-line em operacao nao permitida")

	// 254 e 255: a nota complementar referencia uma, e exatamente uma, nota.
	complementar := n.Ide.FinNFe == FnComplementar
	c.seErro(complementar && len(n.Ide.NFref) == 0,
		254, "Rejeicao: NF-e complementar nao possui NF referenciada")
	c.seErro(complementar && len(n.Ide.NFref) > 1,
		255, "Rejeicao: NF-e complementar possui mais de uma NF referenciada")

	if complementar && len(n.Ide.NFref) == 1 {
		ref := n.Ide.NFref[0]

		// 269: o emitente da complementar tem de ser o da referenciada.
		cnpjDiverge := (ref.RefNF.CNPJ != "" && !strings.EqualFold(ref.RefNF.CNPJ, n.Emit.CNPJCPF)) ||
			(ref.RefNFP.CNPJCPF != "" && !strings.EqualFold(ref.RefNFP.CNPJCPF, n.Emit.CNPJCPF))
		c.seErro(cnpjDiverge,
			269, "Rejeicao: CNPJ Emitente da NF Complementar difere do CNPJ da NF Referenciada")

		// 678: idem para a UF. O teste e pelo NUMERO da nota, nao pela
		// presenca do grupo -- e o jeito do ACBr de saber qual dos dois
		// grupos de referencia foi preenchido.
		ufEmit := pcn.CodigoUF(n.Emit.EnderEmit.UF)
		ufDiverge := (ref.RefNF.NNF > 0 && ref.RefNF.CUF != ufEmit) ||
			(ref.RefNFP.NNF > 0 && ref.RefNFP.CUF != ufEmit)
		c.seErro(ufDiverge,
			678, "Rejeicao: NF referenciada com UF diferente da NF-e complementar")
	}

	// 794: o indicativo de entrega a domicilio e da NFC-e.
	c.seErro(n.Ide.IndPres == PcEntregaDomicilio && n.Ide.TpImp != pcn.TiSimplificadoTipo2,
		794, "Rejeicao: NF-e com indicativo de NFC-e com entrega a domicilio")

	// 237: CPF do destinatario.
	cpfDest := strings.TrimSpace(pcn.OnlyCPFCNPJAlphaNum(n.Dest.CNPJCPF))
	c.seErro(cpfDest != "" && len(cpfDest) <= 11 && pcn.ValidarCPF(n.Dest.CNPJCPF) != nil,
		237, "Rejeicao: CPF do destinatario invalido")

	// 721, 723: coerencia entre o destino da operacao e a identificacao.
	c.seErro(n.Ide.IdDest == DoInterestadual && strings.TrimSpace(n.Dest.CNPJCPF) == "",
		721, "Rejeicao: Operacao interestadual deve informar CNPJ ou CPF")
	c.seErro(n.Ide.IdDest == DoInterna &&
		strings.TrimSpace(n.Dest.IDEstrangeiro) != "" &&
		n.Ide.IndFinal != CfConsumidorFinal,
		723, "Rejeicao: Operacao interna com idEstrangeiro informado deve ser para consumidor final")

	// 724, 726: identificacao minima do destinatario.
	c.seErro(strings.TrimSpace(n.Dest.XNome) == "",
		724, "Rejeicao: NF-e sem o nome do destinatario")
	c.seErro(strings.TrimSpace(n.Dest.EnderDest.XLgr) == "",
		726, "Rejeicao: NF-e sem a informacao de endereco do destinatario")

	// 509: fora do exterior, o municipio do destinatario tem de existir.
	c.seErro(!strings.EqualFold(n.Dest.EnderDest.UF, UFExterior) &&
		!pcn.ValidarMunicipio(n.Dest.EnderDest.CMun),
		509, `Rejeicao: Informado codigo de municipio diferente de "9999999" para operacao com o exterior`)

	// 727, 771, 773: coerencia entre o destino e a UF.
	c.seErro(n.Ide.IdDest == DoExterior && !strings.EqualFold(n.Dest.EnderDest.UF, UFExterior),
		727, "Rejeicao: Operacao com Exterior e UF diferente de EX")
	c.seErro(n.Ide.IdDest == DoInterestadual && strings.EqualFold(n.Dest.EnderDest.UF, UFExterior),
		771, "Rejeicao: Operacao Interestadual e UF de destino com EX")
	c.seErro(n.Ide.IdDest == DoInterna &&
		!strings.EqualFold(n.Dest.EnderDest.UF, n.Emit.EnderEmit.UF) &&
		n.Ide.IndPres != PcPresencial,
		773, "Rejeicao: Operacao Interna e UF de destino difere da UF do emitente - nao presencial")

	// 790: exportacao para contribuinte de ICMS nao existe.
	c.seErro(n.Ide.IdDest == DoExterior && n.Dest.IndIEDest != pcn.InNaoContribuinte,
		790, "Rejeicao: Operacao com Exterior para destinatario Contribuinte de ICMS")

	// 768 e 769: o grupo de pagamento entrou na 4.00 para a NF-e.
	if n.InfNFe.Versao < 4 {
		c.seErro(len(n.Pag.DetPag) > 0,
			768, "Rejeicao: NF-e nao deve possuir o grupo de Formas de Pagamento")
	} else {
		c.seErro(len(n.Pag.DetPag) == 0,
			769, "Rejeicao: NF-e deve possuir o grupo de Formas de Pagamento")
	}

	if n.InfNFe.Versao >= 4 {
		c.regrasDaCobrancaNa400()
	}
}

// regrasDaCobrancaNa400 agrupa o que a 4.00 acrescentou a NF-e
// (ValidarRegrasdeNegocio.pas:256-315).
func (c *coletor) regrasDaCobrancaNa400() {
	n := c.n

	// 864: venda fora do estabelecimento referencia a nota de remessa.
	c.seErro(n.Ide.IndPres == PcPresencialForaEstabelecimento && len(n.Ide.NFref) == 0,
		864, "Rejeicao: NF-e com indicativo de Operacao presencial, fora do estabelecimento e nao informada NF referenciada")

	// 868: operacao interestadual nao declara veiculo nem reboque.
	temVeiculo := strings.TrimSpace(n.Transp.VeicTransp.Placa) != "" ||
		strings.TrimSpace(n.Transp.VeicTransp.UF) != "" ||
		strings.TrimSpace(n.Transp.VeicTransp.RNTC) != ""
	c.seErro(n.Ide.IdDest == DoInterestadual && (temVeiculo || len(n.Transp.Reboque) > 0),
		868, "Rejeicao: Grupos Veiculo Transporte e Reboque nao devem ser informados")

	if n.Ide.FinNFe != FnNormal && n.Ide.FinNFe != FnComplementar {
		return
	}

	// 895 e 902: aritmetica da fatura.
	c.seErro(pcn.CompararValor(n.Cobr.Fat.VDesc, n.Cobr.Fat.VOrig, pcn.ToleranciaMilesimo) > 0,
		895, "Rejeicao: Valor do Desconto da Fatura maior que Valor Original da Fatura")
	c.seErro(pcn.CompararValor(n.Cobr.Fat.VLiq, n.Cobr.Fat.VOrig-n.Cobr.Fat.VDesc, pcn.ToleranciaMilesimo) != 0,
		902, "Rejeicao: Valor Liquido da Fatura difere do Valor Original menos o Valor do Desconto")

	// As parcelas sao conferidas em sequencia: cada vencimento tem de ser
	// maior que o da parcela anterior, e a primeira compara com a emissao.
	somaDuplicatas := 0.0
	ultimoVencimento := apenasData(n.Ide.DEmi)
	for _, d := range n.Cobr.Dup {
		somaDuplicatas += d.VDup

		c.seErro(strings.TrimSpace(d.NDup) == "",
			857, "Rejeicao: Numero da parcela invalido ou nao informado")
		c.seErro(d.DVenc.Before(apenasData(n.Ide.DEmi)),
			894, "Rejeicao: Data de vencimento da parcela nao informada ou menor que Data de Emissao")
		c.seErro(d.DVenc.Before(ultimoVencimento),
			867, "Rejeicao: Data de vencimento da parcela nao informada ou menor que a Data de vencimento da parcela anterior")

		ultimoVencimento = d.DVenc
	}

	// 851: a soma das parcelas nao pode ficar ABAIXO do liquido da fatura.
	// Note que sao duas comparacoes em OU, e as duas sao "menor que" -- uma
	// soma MAIOR que a fatura nao e rejeitada por esta regra.
	if len(n.Cobr.Dup) > 0 {
		abaixoDoLiquido := pcn.CompararValor(n.Cobr.Fat.VLiq, 0, pcn.ToleranciaMilesimo) > 0 &&
			pcn.CompararValor(somaDuplicatas, n.Cobr.Fat.VLiq, pcn.ToleranciaMilesimo) < 0
		abaixoDoOriginal := pcn.CompararValor(somaDuplicatas,
			n.Cobr.Fat.VOrig-n.Cobr.Fat.VDesc, pcn.ToleranciaMilesimo) < 0
		c.seErro(abaixoDoLiquido || abaixoDoOriginal,
			851, "Rejeicao: Soma do valor das parcelas difere do Valor Liquido da Fatura")
	}
}

// ---------------------------------------------------------------------------
// Regras exclusivas da NFC-e (modelo 65)
// ---------------------------------------------------------------------------

// regrasSomenteNFCe e o RegrasValidasSomenteParaNFCe
// (ValidarRegrasdeNegocio.pas:318).
//
// Quase todas sao PROIBICOES: a NFC-e e um documento de venda presencial a
// consumidor final, e tudo que pressupoe outra coisa -- frete, transporte,
// cobranca a prazo, referencia a outro documento -- e rejeitado.
func (c *coletor) regrasSomenteNFCe() {
	n := c.n

	// 789: destinatario contribuinte de ICMS nao compra por NFC-e.
	c.seErro(n.Dest.IndIEDest != pcn.InNaoContribuinte,
		789, "Rejeicao: NFC-e para destinatario contribuinte de ICMS")

	// 704: a NFC-e tem dez minutos de tolerancia de atraso, nao trinta dias.
	emissaoNormal := n.Ide.TpEmis == pcn.TeNormal || n.Ide.TpEmis == pcn.TeSCAN ||
		n.Ide.TpEmis == pcn.TeSVCAN || n.Ide.TpEmis == pcn.TeSVCRS
	c.seErro(n.Ide.DEmi.Before(c.agora.Add(-10*time.Minute)) && emissaoNormal,
		704, "Rejeicao: NFC-e com Data-Hora de emissao atrasada")

	c.seErro(!n.Ide.DSaiEnt.IsZero(), 705, "Rejeicao: NFC-e com data de entrada/saida")
	c.seErro(n.Ide.TpNF == pcn.TnEntrada, 706, "Rejeicao: NFC-e para operacao de entrada")
	c.seErro(n.Ide.IdDest != DoInterna, 707, "NFC-e para operacao interestadual ou com o exterior")

	c.seErro(n.Ide.TpImp != pcn.TiNFCe && n.Ide.TpImp != pcn.TiMsgEletronica,
		709, "Rejeicao: NFC-e com formato de DANFE invalido")
	c.seErro(n.Ide.TpEmis == pcn.TeSCAN, 782, "Rejeicao: NFC-e nao e autorizada pelo SCAN")
	c.seErro(n.Ide.TpEmis == pcn.TeSVCAN || n.Ide.TpEmis == pcn.TeSVCRS,
		783, "Rejeicao: NFC-e nao e autorizada pela SVC")

	c.seErro(n.Ide.FinNFe != FnNormal, 715, "Rejeicao: NFC-e com finalidade invalida")
	c.seErro(n.Ide.IndFinal == CfNao, 716, "Rejeicao: NFC-e em operacao nao destinada a consumidor final")
	c.seErro(n.Ide.IndPres != PcPresencial && n.Ide.IndPres != PcEntregaDomicilio,
		717, "Rejeicao: NFC-e em operacao nao presencial")

	// 785: a lista de UFs que proibem entrega a domicilio esta VAZIA no
	// ACBr -- o proprio codigo pergunta "Qual estado nao permite entrega a
	// domicilio?" e compara contra o literal "XX"
	// (ValidarRegrasdeNegocio.pas:368). A regra nunca dispara. Replicada
	// como esta, sem lista, para o dia em que a UF aparecer.
	c.seErro(n.Ide.IndPres == PcEntregaDomicilio && ufProibeEntregaDomicilio(n.Emit.EnderEmit.UF),
		785, "Rejeicao: NFC-e com entrega a domicilio nao permitida pela UF")

	c.seErro(len(n.Ide.NFref) > 0, 708, "Rejeicao: NFC-e nao pode referenciar documento fiscal")
	c.seErro(strings.TrimSpace(n.Emit.IEST) != "",
		718, "Rejeicao: NFC-e nao deve informar IE de Substituto Tributario")

	c.seErro(n.Ide.IndPres == PcEntregaDomicilio &&
		strings.TrimSpace(n.Entrega.XLgr) == "" &&
		strings.TrimSpace(n.Dest.EnderDest.XLgr) == "",
		787, "Rejeicao: NFC-e de entrega a domicilio sem a identificacao do destinatario")

	c.seErro(strings.TrimSpace(n.Dest.IE) != "", 729, "Rejeicao: NFC-e com informacao da IE do destinatario")
	c.seErro(strings.TrimSpace(n.Dest.ISUF) != "", 730, "Rejeicao: NFC-e com Inscricao Suframa")

	c.regrasDeTransporteDaNFCe()

	// 760: cobranca a prazo nao existe na NFC-e.
	temCobranca := strings.TrimSpace(n.Cobr.Fat.NFat) != "" ||
		maiorQueZero(n.Cobr.Fat.VOrig) || maiorQueZero(n.Cobr.Fat.VDesc) ||
		maiorQueZero(n.Cobr.Fat.VLiq) || len(n.Cobr.Dup) > 0
	c.seErro(temCobranca, 760, "Rejeicao: NFC-e com dados de cobranca (Fatura, Duplicata)")

	c.seErro(len(n.Pag.DetPag) == 0, 769, "Rejeicao: NFC-e deve possuir o grupo de Formas de Pagamento")

	c.seErro(strings.TrimSpace(n.Compra.XNEmp)+strings.TrimSpace(n.Compra.XPed)+
		strings.TrimSpace(n.Compra.XCont) != "",
		762, "Rejeicao: NFC-e com dados de compras (Empenho, Pedido, Contrato)")

	c.seErro(strings.TrimSpace(n.Cana.Safra) != "" || strings.TrimSpace(n.Cana.Ref) != "" ||
		len(n.Cana.ForDia) > 0 || len(n.Cana.Deduc) > 0,
		763, "Rejeicao: NFC-e com dados de aquisicao de Cana")
}

// regrasDeTransporteDaNFCe agrupa as sete proibicoes de transporte
// (ValidarRegrasdeNegocio.pas:394-440).
func (c *coletor) regrasDeTransporteDaNFCe() {
	n := c.n
	entregaDomicilio := n.Ide.IndPres == PcEntregaDomicilio

	c.seErro(n.Transp.ModFrete != MfSemFrete && !entregaDomicilio,
		753, "Rejeicao: NFC-e com Frete")

	transportadorPreenchido := strings.TrimSpace(n.Transp.Transporta.CNPJCPF) != "" ||
		strings.TrimSpace(n.Transp.Transporta.XNome) != "" ||
		strings.TrimSpace(n.Transp.Transporta.IE) != "" ||
		strings.TrimSpace(n.Transp.Transporta.XEnder) != "" ||
		strings.TrimSpace(n.Transp.Transporta.XMun) != "" ||
		strings.TrimSpace(n.Transp.Transporta.UF) != ""
	c.seErro(!entregaDomicilio && transportadorPreenchido,
		754, "Rejeicao: NFC-e com dados do Transportador")

	// 786: e o espelho da anterior -- na entrega a domicilio o
	// transportador passa a ser OBRIGATORIO.
	c.seErro(entregaDomicilio &&
		(strings.TrimSpace(n.Transp.Transporta.CNPJCPF) == "" ||
			strings.TrimSpace(n.Transp.Transporta.XNome) == ""),
		786, "Rejeicao: NFC-e de entrega a domicilio sem dados do Transportador")

	r := n.Transp.RetTransp
	c.seErro(maiorQueZero(r.VServ) || maiorQueZero(r.VBCRet) || maiorQueZero(r.PICMSRet) ||
		maiorQueZero(r.VICMSRet) || strings.TrimSpace(r.CFOP) != "" ||
		maiorQueZero(float64(r.CMunFG)),
		755, "Rejeicao: NFC-e com dados de Retencao do ICMS no Transporte")

	c.seErro(strings.TrimSpace(n.Transp.VeicTransp.Placa) != "" ||
		strings.TrimSpace(n.Transp.VeicTransp.UF) != "" ||
		strings.TrimSpace(n.Transp.VeicTransp.RNTC) != "",
		756, "Rejeicao: NFC-e com dados do veiculo de Transporte")

	c.seErro(len(n.Transp.Reboque) > 0,
		757, "Rejeicao: NFC-e com dados de Reboque do veiculo de Transporte")
	c.seErro(strings.TrimSpace(n.Transp.Vagao) != "",
		758, "Rejeicao: NFC-e com dados do Vagao de Transporte")
	c.seErro(strings.TrimSpace(n.Transp.Balsa) != "",
		759, "Rejeicao: NFC-e com dados da Balsa de Transporte")
}

// ---------------------------------------------------------------------------
// Regras do cabecalho, validas para os dois modelos
// ---------------------------------------------------------------------------

// regrasDoCabecalho e o primeiro trecho de Validar
// (ValidarRegrasdeNegocio.pas:498-555).
func (c *coletor) regrasDoCabecalho() {
	n := c.n

	// 253: digito verificador da chave.
	c.seErro(pcn.ValidarChaveAcesso(n.InfNFe.ID) != nil,
		253, "Rejeicao: Digito Verificador da chave de acesso composta invalida")

	// 270 e 271: municipio do fato gerador.
	c.seErro(!pcn.ValidarMunicipio(n.Ide.CMunFG),
		270, "Rejeicao: Codigo Municipio do Fato Gerador: digito invalido")
	c.seErro(pcn.CodigoUF(n.Emit.EnderEmit.UF) != inteiroOuZero(primeirosDoisDigitos(n.Ide.CMunFG)),
		271, "Rejeicao: Codigo Municipio do Fato Gerador: difere da UF do emitente")

	// 570 e 713: coerencia entre a contingencia da nota e a configurada.
	contingenciaSVC := n.Ide.TpEmis == pcn.TeSCAN || n.Ide.TpEmis == pcn.TeSVCAN ||
		n.Ide.TpEmis == pcn.TeSVCRS
	c.seErro(contingenciaSVC && c.cfg.TpEmis == 1,
		570, "Rejeicao: Tipo de Emissao 3, 6 ou 7 so e valido nas contingencias SCAN/SVC")
	c.seErro((c.cfg.TpEmis == 6 || c.cfg.TpEmis == 7) &&
		!(n.Ide.TpEmis == pcn.TeSVCAN || n.Ide.TpEmis == pcn.TeSVCRS),
		713, "Rejeicao: Tipo de Emissao diferente de 6 ou 7 para contingencia da SVC acessada")

	// 252: o ambiente da nota contra o do web service.
	c.seErro(n.Ide.TpAmb != c.cfg.Ambiente,
		252, "Rejeicao: Ambiente informado diverge do Ambiente de recebimento "+
			"(Tipo do ambiente da NF-e difere do ambiente do Web Service)")

	// 370: nota avulsa nao entra em contingencia.
	avulsa := n.Ide.ProcEmi == pcn.PeAvulsaFisco || n.Ide.ProcEmi == pcn.PeAvulsaContribuinte
	c.seErro(avulsa && n.Ide.TpEmis != pcn.TeNormal,
		370, "Rejeicao: Nota Fiscal Avulsa com tipo de emissao invalido")

	// 556, 557, 558, 569: a justificativa de contingencia e seu espelho.
	c.seErro(n.Ide.TpEmis == pcn.TeNormal && (n.Ide.XJust != "" || !n.Ide.DhCont.IsZero()),
		556, "Justificativa de entrada em contingencia nao deve ser informada para tipo de emissao normal")
	emContingencia := n.Ide.TpEmis == pcn.TeContingencia || n.Ide.TpEmis == pcn.TeFSDA ||
		n.Ide.TpEmis == pcn.TeOffLine
	c.seErro(emContingencia && strings.TrimSpace(n.Ide.XJust) == "",
		557, "A Justificativa de entrada em contingencia deve ser informada")
	// CompareDate compara so a DATA, nao a hora.
	c.seErro(apenasData(n.Ide.DhCont).After(apenasData(c.agora)),
		558, "Rejeicao: Data de entrada em contingencia posterior a data de recebimento")
	c.seErro(!n.Ide.DhCont.IsZero() && c.agora.Sub(n.Ide.DhCont) > 30*24*time.Hour,
		569, "Rejeicao: Data de entrada em contingencia muito atrasada")
}

// regrasDoEmitenteEDestinatario e o trecho de ValidarRegrasdeNegocio.pas:557-663.
func (c *coletor) regrasDoEmitenteEDestinatario() {
	n := c.n

	c.seErro(pcn.ValidarCNPJouCPF(n.Emit.CNPJCPF) != nil,
		207, "Rejeicao: CNPJ do emitente invalido")
	c.seErro(!pcn.ValidarMunicipio(n.Emit.EnderEmit.CMun),
		272, "Rejeicao: Codigo Municipio do Emitente: digito invalido")
	c.seErro(pcn.CodigoUF(n.Emit.EnderEmit.UF) != inteiroOuZero(primeirosDoisDigitos(n.Emit.EnderEmit.CMun)),
		273, "Rejeicao: Codigo Municipio do Emitente: difere da UF do emitente")

	c.seErro(strings.TrimSpace(n.Emit.IE) == "", 229, "Rejeicao: IE do emitente nao informada")
	c.seErro(pcn.ValidarIE(n.Emit.IE, n.Emit.EnderEmit.UF) != nil,
		209, "Rejeicao: IE do emitente invalida")

	// 208: CNPJ do destinatario, so quando o documento tem 14 posicoes ou mais.
	docDest := strings.TrimSpace(pcn.OnlyCPFCNPJAlphaNum(n.Dest.CNPJCPF))
	c.seErro(len(docDest) >= 14 && pcn.ValidarCNPJ(n.Dest.CNPJCPF) != nil,
		208, "Rejeicao: CNPJ do destinatario invalido")

	// 325 e 323: os autorizados a baixar o XML.
	for _, a := range n.AutXML {
		doc := strings.TrimSpace(pcn.OnlyCPFCNPJAlphaNum(a.CNPJCPF))
		c.seErro(len(doc) <= 11 && pcn.ValidarCPF(a.CNPJCPF) != nil,
			325, "Rejeicao: CPF autorizado para download invalido")
		c.seErro(len(doc) > 11 && pcn.ValidarCNPJ(a.CNPJCPF) != nil,
			323, "Rejeicao: CNPJ autorizado para download invalido")
	}
}

// regrasDeLocalETransporte cobre retirada, entrega e transportador
// (ValidarRegrasdeNegocio.pas:584-650).
func (c *coletor) regrasDeLocalETransporte() {
	n := c.n

	c.regrasDeLocal(n.Retirada.UF, n.Retirada.XMun, n.Retirada.CMun,
		"Retirada", 513, 276, 277)
	c.regrasDeLocal(n.Entrega.UF, n.Entrega.XMun, n.Entrega.CMun,
		"Entrega", 515, 278, 279)

	t := n.Transp.Transporta
	doc := strings.TrimSpace(pcn.OnlyCPFCNPJAlphaNum(t.CNPJCPF))
	c.seErro(strings.TrimSpace(t.CNPJCPF) != "" && len(doc) >= 14 && pcn.ValidarCNPJ(t.CNPJCPF) != nil,
		542, "Rejeicao: CNPJ do Transportador invalido")
	c.seErro(strings.TrimSpace(t.CNPJCPF) != "" && len(doc) <= 11 && pcn.ValidarCPF(t.CNPJCPF) != nil,
		543, "Rejeicao: CPF do Transportador invalido")
	c.seErro(strings.TrimSpace(t.IE) != "" && strings.TrimSpace(t.UF) == "",
		559, "Rejeicao: UF do Transportador nao informada")
	c.seErro(strings.TrimSpace(t.IE) != "" && pcn.ValidarIE(t.IE, t.UF) != nil,
		544, "Rejeicao: IE do Transportador invalida")
}

// regrasDeLocal confere um local de retirada ou de entrega. As tres regras
// sao identicas nos dois grupos, so mudam os codigos de rejeicao.
func (c *coletor) regrasDeLocal(uf, xMun string, cMun int, nome string, codEX, codDigito, codUF int) {
	// ASSIMETRIA DO ACBr, replicada: a regra 513 compara a UF de retirada
	// com SameText (sem caixa) e a 515 compara a de entrega com "="
	// (ValidarRegrasdeNegocio.pas:585 x :608). Uma entrega com "ex" em
	// minusculas dispara a 515 no original e nao dispararia aqui se as duas
	// usassem EqualFold.
	exterior := strings.EqualFold(uf, UFExterior)
	if codEX == 515 {
		exterior = uf == UFExterior
	}

	c.seErro(exterior && cMun != CMunExterior, codEX,
		fmt.Sprintf(`Rejeicao: Codigo Municipio do Local de %s deve ser 9999999 para UF %s = "EX"`,
			nome, strings.ToLower(nome)))

	// O digito so e conferido quando o municipio foi NOMEADO -- sem xMun o
	// ACBr considera que o grupo nao foi preenchido.
	c.seErro(!exterior && strings.TrimSpace(xMun) != "" && !pcn.ValidarMunicipio(cMun),
		codDigito, fmt.Sprintf("Rejeicao: Codigo Municipio do Local de %s: digito invalido", nome))

	if strings.TrimSpace(uf) != "" && cMun > 0 {
		codigoUF := pcn.CodigoUF(uf)
		if exterior {
			codigoUF = 99
		}
		c.seErro(codigoUF != inteiroOuZero(primeirosDoisDigitos(cMun)), codUF,
			fmt.Sprintf("Rejeicao: Codigo Municipio do Local de %s: difere da UF do Local de %s",
				nome, nome))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// primeirosDoisDigitos devolve os dois primeiros algarismos de um numero,
// como o `copy(IntToStr(x), 1, 2)` que o ACBr usa para extrair o codigo de
// UF de dentro do codigo de municipio.
func primeirosDoisDigitos(valor int) string {
	s := strconv.Itoa(valor)
	if len(s) < 2 {
		return s
	}
	return s[:2]
}

// apenasData zera a hora, reproduzindo o DateOf e o CompareDate do Delphi.
func apenasData(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// maiorQueZero compara com a tolerancia de milesimo que as regras usam.
func maiorQueZero(v float64) bool {
	return pcn.CompararValor(v, 0, pcn.ToleranciaMilesimo) > 0
}

// algumItemTemDFeReferenciado reproduz o laco da regra 321
// (ValidarRegrasdeNegocio.pas:1237): basta UM item com chave referenciada
// para a regra nao se aplicar.
func algumItemTemDFeReferenciado(n *NFe) bool {
	for _, d := range n.Det {
		if strings.TrimSpace(d.DFeReferenciado.ChaveAcesso) != "" {
			return true
		}
	}
	return false
}

// ufProibeEntregaDomicilio e a lista vazia do ACBr. Ver a regra 785.
func ufProibeEntregaDomicilio(uf string) bool {
	return false
}
