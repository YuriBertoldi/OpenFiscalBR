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

import "github.com/openfiscalbr/openfiscalbr/packages/pcn"

// Tabela de endpoints dos web services da SEFAZ, por AUTORIZADOR.
//
// Porte dos dados de ACBrNFeServicos.ini. A forma e diferente da tabela de
// URLs publicas (urls_tabela.go) por uma razao estrutural, nao de economia:
//
// Nem toda UF opera o proprio ambiente de autorizacao. A maioria delega --
// dezesseis delas para o SVRS no modelo 55, dezenove no 65 -- e o `.ini`
// expressa isso com um redirecionamento `Usar=` de secao para secao. Quem
// decide o endereco, portanto, nao e a UF do emitente: e o AUTORIZADOR que
// atende aquela UF naquele modelo.
//
// Conferido: nas duas familias de servico portadas aqui, o redirecionamento
// e sempre de SECAO INTEIRA -- nao existe UF que defina um servico
// localmente e delegue outro. Se isso mudar num servico futuro, o
// ferramentas/conferir-urls-dfe.py acusa.
//
// Resultado: quinze autorizadores no modelo 55 e doze no 65, contra
// cinquenta e oito combinacoes de UF e ambiente.

// Servico identifica um web service da SEFAZ. O valor e o nome da chave no
// `.ini`, que e tambem o que o LerServicoDeParams procura.
type Servico string

// Servicos portados ate aqui. Os demais entram com quem os transmite.
const (
	ServicoStatus            Servico = "NfeStatusServico"
	ServicoConsultaProtocolo Servico = "NfeConsultaProtocolo"
)

// endpointPorVersao guarda as URLs de um servico por versao de leiaute.
//
// Diferente das URLs publicas do QR-Code, os endpoints VARIAM com a versao:
// a SEFAZ publica um servico novo a cada leiaute e mantem os antigos no ar.
// Versao sem entrada significa que o servico nao existia naquele leiaute.
type endpointPorVersao struct {
	V200 urlPorAmbiente
	V310 urlPorAmbiente
	V400 urlPorAmbiente
}

// para devolve a URL da versao de leiaute pedida.
//
// A queda para a versao anterior ja esta MATERIALIZADA nos valores: quando
// o `.ini` nao tem a versao pedida, a tabela foi gerada com a maior versao
// inferior disponivel. Uma queda em tempo de execucao aqui seria silenciosa
// e mandaria o documento para o endereco errado.
//
// O que NAO foi materializado e a outra metade do mecanismo: no ACBr a
// versao achada tambem volta pelo parametro e vira o `versaoDados` do
// cabecalho (ACBrDFe.pas:505). O porte mantem sempre a versao configurada
// porque o AchaArquivoSchema desfaz aquela queda quando o `.xsd` da versao
// pedida existe (:581-590) -- e existe na instalacao padrao. Dispensar a
// dimensao e deliberado; voltar atras exige o inventario de schemas.
func (e endpointPorVersao) para(versao VersaoDF, ambiente pcn.TipoAmbiente) string {
	switch {
	case versao >= Ve400:
		return e.V400.para(ambiente)
	case versao >= Ve310:
		return e.V310.para(ambiente)
	default:
		return e.V200.para(ambiente)
	}
}

// autorizadorNFe e autorizadorNFCe dizem qual ambiente atende cada UF.
//
// SVC-AN, SVC-RS e AN nao sao UFs: sao os ambientes de contingencia e o
// Ambiente Nacional, que entram na busca no lugar da UF quando a forma de
// emissao configurada manda.
var autorizadorNFe = map[string]string{
	"AC":     "SVRS",
	"AL":     "SVRS",
	"AP":     "SVRS",
	"AM":     "AM",
	"BA":     "BA",
	"CE":     "SVRS",
	"DF":     "SVRS",
	"ES":     "SVRS",
	"GO":     "GO",
	"MA":     "SVAN",
	"MT":     "MT",
	"MS":     "MS",
	"MG":     "MG",
	"PA":     "SVRS",
	"PB":     "SVRS",
	"PR":     "PR",
	"PE":     "PE",
	"PI":     "SVRS",
	"RJ":     "SVRS",
	"RN":     "SVRS",
	"RS":     "RS",
	"RO":     "SVRS",
	"RR":     "SVRS",
	"SC":     "SVRS",
	"SP":     "SP",
	"SE":     "SVRS",
	"TO":     "SVRS",
	"SVC-AN": "SVC-AN",
	"SVC-RS": "SVC-RS",
	"AN":     "AN",
}

var autorizadorNFCe = map[string]string{
	"AC":     "SVRS",
	"AL":     "SVRS",
	"AP":     "SVRS",
	"AM":     "AM",
	"BA":     "SVRS",
	"CE":     "SVRS",
	"DF":     "SVRS",
	"ES":     "SVRS",
	"GO":     "GO",
	"MA":     "SVRS",
	"MT":     "MT",
	"MS":     "MS",
	"MG":     "MG",
	"PA":     "SVRS",
	"PB":     "SVRS",
	"PR":     "PR",
	"PE":     "SVRS",
	"PI":     "SVRS",
	"RJ":     "SVRS",
	"RN":     "SVRS",
	"RS":     "RS",
	"RO":     "SVRS",
	"RR":     "SVRS",
	"SC":     "SVRS",
	"SP":     "SP",
	"SE":     "SVRS",
	"TO":     "SVRS",
	"SVC-AN": "SVC-AN",
	"SVC-RS": "SVC-RS",
	"AN":     "AN",
}

// endpointsPorAutorizador e indexado por "<modelo>/<autorizador>", porque o
// mesmo autorizador publica enderecos DIFERENTES para a NF-e e para a
// NFC-e -- o SVRS, por exemplo, tem um conjunto para cada.
var endpointsPorAutorizador = map[string]map[Servico]endpointPorVersao{
	// --- modelo NFe ---
	"NFe/AM": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeStatusServico2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeStatusServico2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeStatusServico2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeStatusServico4", "https://homnfe.sefaz.am.gov.br/services2/services/NfeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeConsulta2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeConsulta2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeConsulta2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeConsulta4", "https://homnfe.sefaz.am.gov.br/services2/services/NfeConsulta4"},
		},
	},
	"NFe/BA": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeStatusServico2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeStatusServico2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NfeStatusServico/NfeStatusServico.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NfeStatusServico/NfeStatusServico.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeStatusServico4/NFeStatusServico4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeStatusServico4/NFeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/nfeconsulta2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/nfeconsulta2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NfeConsulta/NfeConsulta.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NfeConsulta/NfeConsulta.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx"},
		},
	},
	"NFe/GO": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeStatusServico4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeStatusServico4?wsdl"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeConsultaProtocolo4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeConsultaProtocolo4?wsdl"},
		},
	},
	"NFe/MG": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeStatus2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeStatusServico2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeStatus2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeStatusServico4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeConsulta2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeConsulta2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeConsulta2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeConsultaProtocolo4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeConsultaProtocolo4"},
		},
	},
	"NFe/MS": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeStatusServico2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeStatusServico2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeStatusServico2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeStatusServico4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeConsulta2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeConsulta2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeConsulta2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4"},
		},
	},
	"NFe/MT": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeConsulta4"},
		},
	},
	"NFe/PE": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeStatusServico2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeStatusServico2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeStatusServico2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeStatusServico4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeStatusServico4?wsdl"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeConsulta2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeConsulta2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeConsulta2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeConsultaProtocolo4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeConsultaProtocolo4?wsdl"},
		},
	},
	"NFe/PR": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeStatusServico3?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeStatusServico3?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeConsulta3?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeConsulta3?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeConsultaProtocolo4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeConsultaProtocolo4?wsdl"},
		},
	},
	"NFe/RS": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx"},
		},
	},
	"NFe/SP": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nfestatusservico2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/nfestatusservico2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfestatusservico2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nfeconsulta2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/nfeconsulta2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeconsulta2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeconsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx"},
		},
	},
	"NFe/SVAN": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx"},
		},
	},
	"NFe/SVC-AN": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx", "https://hom.svc.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx"},
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx", "https://hom.svc.fazenda.gov.br/NfeStatusServico2/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx", "https://hom.svc.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx"},
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx", "https://hom.svc.fazenda.gov.br/NfeConsulta2/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx"},
		},
	},
	"NFe/SVC-RS": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx"},
		},
	},
	"NFe/SVRS": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx"},
		},
	},
	// --- modelo NFCe ---
	"NFCe/AM": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeStatusServico2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeStatusServico4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeConsulta2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeConsulta4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeConsulta4"},
		},
	},
	"NFCe/GO": {
		ServicoStatus: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeStatusServico2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeStatusServico4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeStatusServico4?wsdl"},
		},
		ServicoConsultaProtocolo: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeConsulta2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeConsultaProtocolo4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeConsultaProtocolo4?wsdl"},
		},
	},
	"NFCe/MG": {
		ServicoStatus: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeStatusServico4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeConsultaProtocolo4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeConsultaProtocolo4"},
		},
	},
	"NFCe/MS": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/NfeStatusServico2", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeStatusServico4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/NfeConsulta2", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeConsultaProtocolo4"},
		},
	},
	"NFCe/MT": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeStatusServico2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeStatusServico2?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeStatusServico4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeConsulta2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeConsulta2?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeConsulta4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeConsulta4"},
		},
	},
	"NFCe/PR": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeStatusServico3", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeStatusServico3"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeStatusServico4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeConsulta3", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeConsulta3"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeConsultaProtocolo4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeConsultaProtocolo4"},
		},
	},
	"NFCe/RS": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx"},
		},
	},
	"NFCe/SP": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/nfestatusservico2.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/nfestatusservico2.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeStatusServico4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/nfeconsulta2.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/nfeconsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeConsultaProtocolo4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeConsultaProtocolo4.asmx"},
		},
	},
	"NFCe/SVRS": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico2.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta2.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx"},
		},
	},
}

// soapActionPorAutorizador sao os overrides de SoapAction que o `.ini`
// declara pela chave `SoapAction=`, apontando para uma secao de actions.
//
// Uma secao inteira do `.ini` usa isso: a Bahia, pela chave `SoapAction=`
// de [NFe_BA_P] e [NFe_BA_H] apontando para [SOAP_V4_BA]. E la a action da
// consulta de protocolo usa "NFeConsulta4", nao o "NFeConsultaProtocolo4"
// que o DefinirServicoEAction montaria.
//
// Por que isso existe: no ACBr a ordem e DefinirURL -> DefinirServicoEAction
// (ACBrDFeWebService.pas:260-264), e o segundo so preenche quando o primeiro
// deixou vazio -- e o `EstaVazio(FPServico)` de ACBrNFeWebServices.pas:2123.
// O endereco do envelope (FPServico) NAO muda, porque nenhuma secao usa a
// chave `WSDL=`; so a action muda.
//
// A action do status da BA coincide com a montada, e por isso nao precisa
// estar aqui -- mas esta, para a tabela refletir o `.ini` e nao a
// coincidencia.
var soapActionPorAutorizador = map[string]map[Servico]endpointPorVersao{
	"NFe/BA": {
		ServicoStatus: {
			V400: urlPorAmbiente{
				enderecoWsd + "NFeStatusServico4/nfeStatusServicoNF",
				enderecoWsd + "NFeStatusServico4/nfeStatusServicoNF",
			},
		},
		ServicoConsultaProtocolo: {
			V400: urlPorAmbiente{
				enderecoWsd + "NFeConsulta4/nfeConsultaNF",
				enderecoWsd + "NFeConsulta4/nfeConsultaNF",
			},
		},
	},
}
