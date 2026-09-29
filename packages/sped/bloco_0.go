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

package sped

import "time"

// ---------------------------------------------------------------------------
// Registro0000 - Abertura do arquivo digital e identificacao da entidade
// ---------------------------------------------------------------------------

// Registro0000 contem os dados de abertura do arquivo digital e identificacao
// da entidade responsavel pelo envio do SPED Fiscal.
type Registro0000 struct {
	CodVer    VersaoLeiauteFiscal // Codigo da versao do leiaute
	CodFin    CodFin              // Codigo da finalidade do arquivo (0=Original, 1=Substituto)
	DtIni     time.Time           // Data inicial das informacoes contidas no arquivo
	DtFin     time.Time           // Data final das informacoes contidas no arquivo
	Nome      string              // Nome empresarial da entidade
	CNPJ      string              // CNPJ da entidade
	CPF       string              // CPF do responsavel
	UF        string              // Sigla da Unidade da Federacao
	IE        string              // Inscricao Estadual da entidade
	CodMun    int                 // Codigo do municipio do domicilio fiscal (IBGE)
	IM        string              // Inscricao Municipal da entidade
	Suframa   string              // Inscricao na SUFRAMA
	IndPerfil IndPerfil           // Perfil de apresentacao do arquivo fiscal (A, B ou C)
	IndAtiv   IndAtiv             // Indicador de tipo de atividade (0=Industrial, 1=Outros)
}

// ---------------------------------------------------------------------------
// Registro0001 - Abertura do Bloco 0
// ---------------------------------------------------------------------------

// Registro0001 indica a abertura do Bloco 0 e contem os sub-registros de
// cadastro (participantes, itens, unidades, etc.).
type Registro0001 struct {
	OpenBlocos
	Registro0005 *Registro0005   // Dados complementares da entidade
	Registro0015 []*Registro0015 // Dados do contribuinte substituto
	Registro0100 *Registro0100   // Dados do contabilista
	Registro0150 []*Registro0150 // Tabela de cadastro do participante
	Registro0190 []*Registro0190 // Identificacao das unidades de medida
	Registro0200 []*Registro0200 // Tabela de identificacao do item
	Registro0300 []*Registro0300 // Cadastro de bens do ativo imobilizado
	Registro0400 []*Registro0400 // Tabela de natureza da operacao
	Registro0450 []*Registro0450 // Tabela de informacao complementar
	Registro0460 []*Registro0460 // Tabela de observacoes do lancamento fiscal
	Registro0500 []*Registro0500 // Plano de contas contabeis
	Registro0600 []*Registro0600 // Centro de custos
}

// NewRegistro0001 cria um novo Registro0001 com IndDad=1 (sem dados),
// indicando que o bloco inicialmente nao possui movimentacao.
func NewRegistro0001() *Registro0001 {
	return &Registro0001{
		OpenBlocos: OpenBlocos{IndDad: 1},
	}
}

// ---------------------------------------------------------------------------
// Registro0002 - Classificacao do estabelecimento industrial ou equiparado
// ---------------------------------------------------------------------------

// Registro0002 contem a classificacao do estabelecimento industrial ou
// equiparado a industrial, conforme tabela TIPI.
type Registro0002 struct {
	ClasEstabInd string // Classificacao industrial do estabelecimento
}

// ---------------------------------------------------------------------------
// Registro0005 - Dados complementares da entidade
// ---------------------------------------------------------------------------

// Registro0005 contem informacoes complementares da entidade como endereco,
// telefone, fax e e-mail.
type Registro0005 struct {
	Fantasia string // Nome fantasia associado ao nome empresarial
	CEP      string // Codigo de Enderecamento Postal
	Endereco string // Logradouro e endereco do imovel
	Num      string // Numero do imovel
	Compl    string // Dados complementares do endereco
	Bairro   string // Bairro em que o imovel esta situado
	Fone     string // Numero do telefone (DDD+Fone)
	Fax      string // Numero do fax
	Email    string // Endereco do correio eletronico
}

// ---------------------------------------------------------------------------
// Registro0015 - Dados do contribuinte substituto ou responsavel pelo ICMS
// ---------------------------------------------------------------------------

// Registro0015 contem os dados do contribuinte substituto tributario ou
// responsavel pelo recolhimento do ICMS destino (EC 87/15).
type Registro0015 struct {
	UfST string // Sigla da UF do contribuinte substituido
	IeST string // Inscricao Estadual do contribuinte substituto na UF
}

// ---------------------------------------------------------------------------
// Registro0100 - Dados do contabilista
// ---------------------------------------------------------------------------

// Registro0100 contem os dados do contabilista responsavel pela escrituracao
// fiscal digital do contribuinte.
type Registro0100 struct {
	Nome     string // Nome do contabilista
	CPF      string // CPF do contabilista
	CRC      string // Numero de inscricao no Conselho Regional de Contabilidade
	CNPJ     string // CNPJ do escritorio de contabilidade
	CEP      string // Codigo de Enderecamento Postal
	Endereco string // Logradouro e endereco do imovel
	Num      string // Numero do imovel
	Compl    string // Dados complementares do endereco
	Bairro   string // Bairro em que o imovel esta situado
	Fone     string // Numero do telefone (DDD+Fone)
	Fax      string // Numero do fax
	Email    string // Endereco do correio eletronico
	CodMun   int    // Codigo do municipio (IBGE)
}

// ---------------------------------------------------------------------------
// Registro0150 - Tabela de cadastro do participante
// ---------------------------------------------------------------------------

// Registro0150 contem os dados de cadastro dos participantes (clientes,
// fornecedores, transportadores, etc.) referenciados nos documentos fiscais.
type Registro0150 struct {
	CodPart      string          // Codigo de identificacao do participante no arquivo
	Nome         string          // Nome pessoal ou empresarial do participante
	CodPais      string          // Codigo do pais do participante (tabela BACEN)
	CNPJ         string          // CNPJ do participante
	CPF          string          // CPF do participante
	IE           string          // Inscricao Estadual do participante
	CodMun       int             // Codigo do municipio (IBGE)
	Suframa      string          // Numero de inscricao na SUFRAMA
	Endereco     string          // Logradouro e endereco do imovel
	Num          string          // Numero do imovel
	Compl        string          // Dados complementares do endereco
	Bairro       string          // Bairro em que o imovel esta situado
	Registro0175 []*Registro0175 // Alteracoes da tabela de cadastro de participante
}

// ---------------------------------------------------------------------------
// Registro0175 - Alteracao da tabela de cadastro de participante
// ---------------------------------------------------------------------------

// Registro0175 registra as alteracoes efetuadas nos dados de cadastro do
// participante durante o periodo informado.
type Registro0175 struct {
	DtAlt   time.Time // Data de alteracao do cadastro
	NrCampo string    // Numero do campo alterado (campos 03 a 13 do Reg. 0150)
	ContAnt string    // Conteudo anterior do campo
}

// ---------------------------------------------------------------------------
// Registro0190 - Identificacao das unidades de medida
// ---------------------------------------------------------------------------

// Registro0190 contem a identificacao das unidades de medida utilizadas
// no arquivo digital.
type Registro0190 struct {
	Unid  string // Codigo da unidade de medida
	Descr string // Descricao da unidade de medida
}

// ---------------------------------------------------------------------------
// Registro0200 - Tabela de identificacao do item (produtos e servicos)
// ---------------------------------------------------------------------------

// Registro0200 contem a identificacao dos itens (produtos e servicos)
// referenciados nos documentos fiscais e nos registros de inventario.
type Registro0200 struct {
	CodItem      string          // Codigo do item
	DescrItem    string          // Descricao do item
	CodBarra     string          // Representacao alfanumerico do codigo de barra (EAN/GTIN)
	CodAntItem   string          // Codigo anterior do item com referencia a ultima informacao apresentada
	UnidInv      string          // Unidade de medida utilizada na quantificacao de estoques
	TipoItem     TipoItem        // Tipo do item (00=Mercadoria Revenda, 01=Materia Prima, etc.)
	CodNCM       string          // Codigo da Nomenclatura Comum do Mercosul
	ExIPI        string          // Codigo EX conforme Tabela de Incidencia do IPI (TIPI)
	CodGen       string          // Codigo do genero do item (tabela 4.2.1)
	CodLst       string          // Codigo do servico conforme lista do Anexo I da Lei Complementar 116/03
	AliqICMS     float64         // Aliquota de ICMS aplicavel ao item nas operacoes internas
	CEST         string          // Codigo Especificador da Substituicao Tributaria
	Registro0205 []*Registro0205 // Alteracoes do item
	Registro0206 []*Registro0206 // Codigo de produto conforme tabela ANP
	Registro0210 []*Registro0210 // Consumo especifico padronizado
	Registro0220 []*Registro0220 // Fatores de conversao de unidades
	Registro0221 []*Registro0221 // Correlacao entre codigos de itens comercializados
}

// ---------------------------------------------------------------------------
// Registro0205 - Alteracao do item
// ---------------------------------------------------------------------------

// Registro0205 registra as alteracoes efetuadas na descricao ou no codigo
// anterior do item durante o periodo informado.
type Registro0205 struct {
	DescrAntItem string    // Descricao anterior do item
	DtIni        time.Time // Data inicial de utilizacao da descricao do item
	DtFin        time.Time // Data final de utilizacao da descricao do item
	CodAntItem   string    // Codigo anterior do item com referencia a ultima informacao apresentada
}

// ---------------------------------------------------------------------------
// Registro0206 - Codigo de produto conforme tabela ANP (combustiveis)
// ---------------------------------------------------------------------------

// Registro0206 contem o codigo do produto conforme tabela de combustiveis
// da Agencia Nacional do Petroleo (ANP).
type Registro0206 struct {
	CodComb string // Codigo do combustivel conforme tabela ANP
}

// ---------------------------------------------------------------------------
// Registro0210 - Consumo especifico padronizado
// ---------------------------------------------------------------------------

// Registro0210 contem informacoes sobre o consumo especifico padronizado
// para producao de um determinado item (composicao).
type Registro0210 struct {
	CodItemComp string  // Codigo do item componente/insumo (campo 02 do Reg. 0200)
	QtdComp     float64 // Quantidade do item componente/insumo para se produzir uma unidade do item composto
	Perda       float64 // Perda/quebra normal percentual do insumo/componente no processo produtivo
}

// ---------------------------------------------------------------------------
// Registro0220 - Fatores de conversao de unidades
// ---------------------------------------------------------------------------

// Registro0220 contem os fatores de conversao entre as diversas unidades
// de medida utilizadas para o mesmo item.
type Registro0220 struct {
	UnidConv string  // Unidade comercial a ser convertida na unidade de estoque (campo 06 do Reg. 0200)
	FatConv  float64 // Fator de conversao: fator de conversao da unidade comercial para unidade de estoque
	CodBarra string  // Representacao alfanumerico do codigo de barra da unidade comercial do produto
}

// ---------------------------------------------------------------------------
// Registro0221 - Correlacao entre codigos de itens comercializados
// ---------------------------------------------------------------------------

// Registro0221 contem a correlacao entre os codigos de itens comercializados
// e os codigos de itens que os compoem (atomicos).
type Registro0221 struct {
	CodItemAtomico string  // Codigo do item atomico contido no item comercializado
	QtdeContida    float64 // Quantidade do item atomico contida no item comercializado
}

// ---------------------------------------------------------------------------
// Registro0300 - Cadastro de bens ou componentes do ativo imobilizado
// ---------------------------------------------------------------------------

// Registro0300 contem os dados de cadastro dos bens ou componentes do ativo
// imobilizado do contribuinte.
type Registro0300 struct {
	CodIndBem    string        // Codigo individualizado do bem ou componente adotado no controle patrimonial
	IdentMerc    int           // Identificacao do tipo de mercadoria (1=Bem; 2=Componente)
	DescrItem    string        // Descricao do bem ou componente (grupo de ativo imobilizado)
	CodPrnc      string        // Codigo de cadastro do bem principal (nos casos em que o bem e componente)
	CodCta       string        // Codigo da conta analitica de contabilizacao do bem ou componente
	NrParc       float64       // Numero total de parcelas a serem apropriadas conforme a legislacao de cada UF
	Registro0305 *Registro0305 // Informacao sobre utilizacao do bem
}

// ---------------------------------------------------------------------------
// Registro0305 - Informacao sobre utilizacao do bem
// ---------------------------------------------------------------------------

// Registro0305 contem informacoes complementares sobre a utilizacao do bem
// ou componente do ativo imobilizado.
type Registro0305 struct {
	CodCcus  string // Codigo do centro de custo onde o bem esta sendo ou sera utilizado
	Func     string // Descricao sucinta da funcao do bem na atividade do estabelecimento
	VidaUtil int    // Vida util estimada do bem, em numero de meses
}

// ---------------------------------------------------------------------------
// Registro0400 - Tabela de natureza da operacao/prestacao
// ---------------------------------------------------------------------------

// Registro0400 contem a tabela com as naturezas das operacoes/prestacoes
// (CFOP) utilizadas pelo contribuinte.
type Registro0400 struct {
	CodNat   string // Codigo da natureza da operacao/prestacao
	DescrNat string // Descricao da natureza da operacao/prestacao
}

// ---------------------------------------------------------------------------
// Registro0450 - Tabela de informacao complementar do documento fiscal
// ---------------------------------------------------------------------------

// Registro0450 contem a tabela de informacoes complementares vinculadas
// aos documentos fiscais.
type Registro0450 struct {
	CodInf string // Codigo da informacao complementar do documento fiscal
	Txt    string // Texto livre da informacao complementar
}

// ---------------------------------------------------------------------------
// Registro0460 - Tabela de observacoes do lancamento fiscal
// ---------------------------------------------------------------------------

// Registro0460 contem a tabela de observacoes vinculadas aos lancamentos
// fiscais (livros de entrada e saida).
type Registro0460 struct {
	CodObs string // Codigo da observacao do lancamento fiscal
	Txt    string // Texto livre da observacao do lancamento fiscal
}

// ---------------------------------------------------------------------------
// Registro0500 - Plano de contas contabeis
// ---------------------------------------------------------------------------

// Registro0500 contem o plano de contas contabeis utilizado pelo contribuinte,
// mapeando as contas de acordo com a legislacao vigente.
type Registro0500 struct {
	DtAlt    time.Time // Data da inclusao/alteracao
	CodNatCC string    // Codigo da natureza da conta/grupo de contas (tabela 3.1.1)
	IndCta   string    // Indicador do tipo de conta (S=Sintetica; A=Analitica)
	Nivel    string    // Nivel da conta analitica/sintetica
	CodCta   string    // Codigo da conta analitica/sintetica
	NomeCta  string    // Nome da conta analitica/sintetica
}

// ---------------------------------------------------------------------------
// Registro0600 - Centro de custos
// ---------------------------------------------------------------------------

// Registro0600 contem os centros de custos utilizados pelo contribuinte
// na escrituracao contabil e fiscal.
type Registro0600 struct {
	DtAlt   time.Time // Data da inclusao/alteracao
	CodCcus string    // Codigo do centro de custos
	Ccus    string    // Nome do centro de custos
}

// ---------------------------------------------------------------------------
// Registro0990 - Encerramento do Bloco 0
// ---------------------------------------------------------------------------

// Registro0990 indica o encerramento do Bloco 0, contendo a quantidade total
// de linhas do bloco (incluindo os registros de abertura e encerramento).
type Registro0990 struct {
	QtdLin0 int // Quantidade total de linhas do Bloco 0
}
