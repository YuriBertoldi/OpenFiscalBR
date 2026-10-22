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

package cte

import "strings"

// Tipos de CONTROLE do componente -- os que descrevem o estado da
// transmissao e a configuracao, nao o conteudo do documento.
//
// A separacao em relacao a conversao.go segue a do packages/nfe: la ficam
// os enums do LEIAUTE, que saem no XML; aqui os que so existem dentro do
// componente.

// VersaoCTe e a versao do leiaute. Porte de TVersaoCTe.
type VersaoCTe int

const (
	Ve200 VersaoCTe = iota // 2.00
	Ve300                  // 3.00
	Ve400                  // 4.00
	// _fimVersaoCTe conta os membros declarados; nao e valor do leiaute.
	_fimVersaoCTe
)

var versaoCTeCodigos = [...]string{"2.00", "3.00", "4.00"}

var versaoCTeValores = [...]float64{2.00, 3.00, 4.00}

func (v VersaoCTe) String() string { return codigoEm(versaoCTeCodigos[:], int(v)) }

// Float retorna a versao como numero, equivalente a VersaoCTeToDbl.
// Fora de faixa devolve 0, como o case-else do Delphi.
func (v VersaoCTe) Float() float64 {
	if int(v) < 0 || int(v) >= len(versaoCTeValores) {
		return 0
	}
	return versaoCTeValores[v]
}

func ParseVersaoCTe(s string) (VersaoCTe, error) {
	i, err := indiceDe(versaoCTeCodigos[:], s, "VersaoCTe")
	return VersaoCTe(i), err
}

// StatusACBrCTe e o estado do componente durante uma operacao.
// Porte de TStatusACBrCTe.
//
// Nao sai no documento: serve ao callback de progresso. Esta aqui porque o
// porte do componente vai precisar dele, e declara-lo junto dos enums do
// leiaute confundiria os dois papeis.
type StatusACBrCTe int

const (
	StCTeIdle            StatusACBrCTe = iota // ocioso
	StCTeStatusServico                        // consultando o status
	StCTeRecepcao                             // transmitindo o lote
	StCTeRetRecepcao                          // consultando o recibo
	StCTeConsulta                             // consultando por chave
	StCTeCancelamento                         // cancelando
	StCTeInutilizacao                         // inutilizando faixa
	StCTeRecibo                               // lendo o recibo
	StCTeCadastro                             // consultando cadastro
	StCTeEmail                                // enviando email
	StCTeCCe                                  // carta de correcao
	StCTeEvento                               // evento generico
	StCTeDistDFeInt                           // distribuicao DF-e
	StCTeEnvioWebService                      // envio bruto
	// _fimStatusACBrCTe conta os membros; nao e valor do leiaute.
	_fimStatusACBrCTe
)

// LayOutCTe e o servico do web service. Porte de TLayOutCTe.
//
// Os dezesseis membros nao sao dezesseis URLs: tres deles -- o QR-Code e as
// duas consultas publicas -- resolvem para endereco de PAGINA, nao de
// servico SOAP. A distincao importa na etapa das URLs.
type LayOutCTe int

const (
	LayCTeRecepcao      LayOutCTe = iota // CTeRecepcao
	LayCTeRetRecepcao                    // CTeRetRecepcao
	LayCTeCancelamento                   // CTeCancelamento
	LayCTeInutilizacao                   // CTeInutilizacao
	LayCTeConsulta                       // CTeConsultaProtocolo
	LayCTeStatusServico                  // CTeStatusServico
	LayCTeCadastro                       // CTeConsultaCadastro
	LayCTeEvento                         // RecepcaoEvento
	LayCTeEventoAN                       // RecepcaoEventoAN
	LayCTeDistDFeInt                     // CTeDistribuicaoDFe
	LayCTeRecepcaoOS                     // CTeRecepcaoOS
	LayCTeRecepcaoSinc                   // CTeRecepcaoSinc
	LayCTeRecepcaoGTVe                   // CTeRecepcaoGTVe
	LayCTeURLQRCode                      // pagina do QR-Code
	LayURLConsultaCTe                    // pagina de consulta publica
	LayCTeRecepcaoSimp                   // CTeRecepcaoSimp
	// _fimLayOutCTe conta os membros; nao e valor do leiaute.
	_fimLayOutCTe
)

// layOutCTeServicos e o nome do servico no `.ini` de enderecos.
//
// A tabela do ACBr (LayOutToServico) tem CATORZE entradas para dezesseis
// membros: o QR-Code e a consulta publica nao tem nome de servico, porque
// nao sao servico -- sao pagina. Ficam vazios aqui pela mesma razao.
var layOutCTeServicos = [...]string{
	"CTeRecepcao", "CTeRetRecepcao", "CTeCancelamento", "CTeInutilizacao",
	"CTeConsultaProtocolo", "CTeStatusServico", "CTeConsultaCadastro",
	"RecepcaoEvento", "RecepcaoEventoAN", "CTeDistribuicaoDFe",
	"CTeRecepcaoOS", "CTeRecepcaoSinc", "CTeRecepcaoGTVe",
	"", "", "CTeRecepcaoSimp",
}

// Servico retorna o nome do servico no `.ini` de enderecos, ou vazio
// quando o membro nao corresponde a um servico SOAP.
func (v LayOutCTe) Servico() string { return codigoEm(layOutCTeServicos[:], int(v)) }

// ParseLayOutCTe converte o nome do servico no enum.
func ParseLayOutCTe(s string) (LayOutCTe, error) {
	for i, nome := range layOutCTeServicos {
		if nome != "" && strings.EqualFold(nome, s) {
			return LayOutCTe(i), nil
		}
	}
	return LayCTeRecepcao, erroEnum("LayOutCTe", s)
}

// TamanhoPapel e o formato do DACTE. Porte de TpcnTamanhoPapel.
//
// E apresentacao, nao leiaute. Declarado porque o componente o carrega na
// configuracao; a impressao em si esta fora do escopo do porte.
type TamanhoPapel int

const (
	TpA4 TamanhoPapel = iota
	TpA4Duas2Vias
	TpA5
	// _fimTamanhoPapel conta os membros; nao e valor do leiaute.
	_fimTamanhoPapel
)

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

// SchemaCTe identifica o schema de um documento ou evento do CT-e.
// Porte de TSchemaCTe.
//
// O nome NAO vem de tabela: o ACBr o deriva do PROPRIO NOME DO MEMBRO,
// cortando o prefixo "sch" (SchemaCTeToStr usa GetEnumName). A
// `TSchemaCTeArrayStrings` que existe no topo da unit tem dezoito entradas
// VAZIAS e e letra morta -- nenhuma funcao a consulta.
//
// Por isso a tabela aqui guarda o nome JA SEM o prefixo, e nao o codigo:
// e ele que aparece no atributo `schema` dos documentos e no nome do `.xsd`.
type SchemaCTe int

const (
	// SchErro e o primeiro membro e o valor de erro do Parse, como no
	// original.
	SchErro SchemaCTe = iota
	SchCTe
	SchCTeOS
	SchGTVe
	SchCancCTe
	SchInutCTe
	SchEventoCTe
	SchCTeSimp
	SchConsReciCTe
	SchConsSitCTe
	SchConsStatServCTe
	SchConsCad
	SchCTeModalAereo
	SchCTeModalAquaviario
	SchCTeModalDutoviario
	SchCTeModalFerroviario
	SchCTeModalRodoviario
	SchCTeMultiModal
	SchEvEPECCTe
	SchEvCancCTe
	SchEvRegMultimodal
	SchEvCCeCTe
	SchDistDFeInt
	SchCTeModalRodoviarioOS
	SchEvPrestDesacordo
	SchEvGTV
	SchEvCECTe
	SchEvCancCECTe
	SchEvCancPrestDesacordo
	SchEvIECTe
	SchEvCancIECTe
	SchEvVincPgto
	SchEvCancVincPgto
	// _fimSchemaCTe conta os membros; nao e valor do leiaute.
	_fimSchemaCTe
)

// schemaCTeNomes guarda o nome COM A CAIXA EXATA do membro Delphi, menos o
// prefixo "sch".
//
// A caixa importa: o ACBr deriva o nome do identificador Pascal, e ha
// membros que comecam em minuscula (`schcancCTe` -> "cancCTe") convivendo
// com membros que comecam em maiuscula (`schCTe` -> "CTe"). Normalizar
// tudo para PascalCase mudaria o atributo `schema` do documento.
var schemaCTeNomes = [...]string{
	"Erro", "CTe", "CTeOS", "GTVe", "cancCTe", "InutCTe",
	"EventoCTe", "CTeSimp", "consReciCTe", "consSitCTe",
	"consStatServCTe", "consCad",
	"cteModalAereo", "cteModalAquaviario", "cteModalDutoviario",
	"cteModalFerroviario", "cteModalRodoviario", "cteMultiModal",
	"evEPECCTe", "evCancCTe", "evRegMultimodal", "evCCeCTe",
	"distDFeInt", "cteModalRodoviarioOS",
	"evPrestDesacordo", "evGTV", "evCECTe", "evCancCECTe",
	"evCancPrestDesacordo", "evIECTe", "evCancIECTe",
	"evVincPgto", "evCancVincPgto",
}

// String retorna o nome do schema, sem o prefixo "sch".
func (v SchemaCTe) String() string { return codigoEm(schemaCTeNomes[:], int(v)) }

// ParseSchemaCTe converte o nome do schema no enum.
//
// Porte de StrToSchemaCTe, que faz duas coisas antes de buscar:
//
//  1. CORTA no primeiro "_". O atributo `schema` dos documentos vem como
//     "cteModalRodoviario_v4.00", e so a parte antes do underscore nomeia o
//     schema;
//  2. ACRESCENTA o prefixo "sch" quando ele nao esta la -- o que torna
//     aceitaveis tanto "CTe" quanto "schCTe".
//
// DIVERGENCIA: o original devolve `TSchemaCTe(-1)` para nome desconhecido e
// ainda assim marca `ok := True` (o GetEnumValue devolve -1 e o codigo nao
// confere). Um schema novo vira um enum fora de faixa que se propaga em
// silencio. Aqui devolve SchErro mais um erro, e quem chama decide -- foi a
// mesma decisao do pcn.ParseSchemaDFe na Distribuicao DF-e da NF-e.
func ParseSchemaCTe(s string) (SchemaCTe, error) {
	nome := s
	if i := strings.Index(nome, "_"); i >= 0 {
		nome = nome[:i]
	}
	nome = strings.TrimPrefix(nome, "sch")

	for i, n := range schemaCTeNomes {
		if strings.EqualFold(n, nome) {
			return SchemaCTe(i), nil
		}
	}
	return SchErro, erroEnum("SchemaCTe", s)
}
