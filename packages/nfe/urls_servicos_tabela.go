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
// Resultado: quinze autorizadores no modelo 55 e doze no 65, contra
// cinquenta e oito combinacoes de UF e ambiente.
//
// O redirecionamento, porem, NAO e de secao inteira: o ACBr resolve servico
// a servico, tentando a secao local e so caindo no `Usar=` quando aquela
// chave esta vazia (ACBrDFe.pas:550-571). Uma UF pode portanto definir um
// servico localmente e delegar todos os demais -- e uma faz: o MA, que no
// modelo 55 delega ao SVAN mas mantem o proprio NfeConsultaCadastro_2.00.
//
// A tabela continua valida porque o MA e a UNICA UF do grupo SVAN: a
// entrada "NFe/SVAN" e, na pratica, so dele. O dia em que outra UF entrar
// nesse autorizador, ela herdaria o cadastro do MA, que o `.ini` nunca lhe
// daria. O ferramentas/conferir-endpoints-dfe.py vigia exatamente isso:
// hoje emite AVISO para o MA e passa a FALHAR se o grupo crescer.

// Servico identifica um web service da SEFAZ. O valor e o nome da chave no
// `.ini`, que e tambem o que o LerServicoDeParams procura.
type Servico string

// Servicos portados ate aqui. Os demais entram com quem os transmite.
const (
	ServicoStatus            Servico = "NfeStatusServico"
	ServicoConsultaProtocolo Servico = "NfeConsultaProtocolo"
	ServicoAutorizacao       Servico = "NfeAutorizacao"
	ServicoRetAutorizacao    Servico = "NfeRetAutorizacao"
	ServicoInutilizacao      Servico = "NfeInutilizacao"
	ServicoConsultaCadastro  Servico = "NfeConsultaCadastro"
	ServicoAdministrarCSC    Servico = "AdministrarCSCNFCe"

	// Os dois do lote ASSINCRONO, que a 3.10 aposentou em favor do par
	// Autorizacao/RetAutorizacao. Permanecem porque a tabela os tem e
	// porque a 2.00 ainda e leitura valida.
	ServicoRecepcaoLote    Servico = "NfeRecepcao"
	ServicoRetRecepcaoLote Servico = "NfeRetRecepcao"

	// ServicoRecepcaoEvento atende TODOS os eventos -- carta de correcao,
	// cancelamento, manifestacao do destinatario e os da Reforma. No ACBr
	// tres LayOut diferentes (LayNFeCCe, LayNFeEvento e LayNFeEventoAN)
	// mapeiam para este mesmo nome de servico; o que muda entre eles e o
	// SCHEMA e, sobretudo, a UF sob a qual o endereco e procurado.
	ServicoRecepcaoEvento Servico = "RecepcaoEvento"

	// ServicoEventoEPEC atende UM caso so: o EPEC de NFC-e em Sao Paulo,
	// a unica UF que o oferece -- as demais usam a contingencia off-line
	// (ACBrNFeWebServices.pas:3101).
	ServicoEventoEPEC Servico = "EventoEPEC"

	// ServicoDistribuicaoDFe e a Distribuicao DF-e, tratada DIRETAMENTE
	// pela RFB: o DefinirURL forca a UF "AN" e nem olha a configuracao
	// (ACBrNFeWebServices.pas:4001). E o unico servico da NF-e sem
	// variacao por UF.
	ServicoDistribuicaoDFe Servico = "NFeDistribuicaoDFe"
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
	"SVRS":   "SVRS",
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
	"SVRS":   "SVRS",
}

// endpointsPorAutorizador e indexado por "<modelo>/<autorizador>", porque o
// mesmo autorizador publica enderecos DIFERENTES para a NF-e e para a
// NFC-e -- o SVRS, por exemplo, tem um conjunto para cada.
var endpointsPorAutorizador = map[string]map[Servico]endpointPorVersao{
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeAutorizacao", "https://homnfe.sefaz.am.gov.br/services2/services/NfeAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeAutorizacao4", "https://homnfe.sefaz.am.gov.br/services2/services/NfeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao4", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao4"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRecepcao2"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetRecepcao2"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeInutilizacao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeInutilizacao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeInutilizacao2", "https://homnfe.sefaz.am.gov.br/services2/services/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/NfeInutilizacao4", "https://homnfe.sefaz.am.gov.br/services2/services/NfeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/cadconsultacadastro2", "https://homnfe.sefaz.am.gov.br/services2/services/cadconsultacadastro2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/cadconsultacadastro2", "https://homnfe.sefaz.am.gov.br/services2/services/cadconsultacadastro2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/CadConsultaCadastro4", "https://homnfe.sefaz.am.gov.br/services2/services/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/RecepcaoEvento", "https://homnfe.sefaz.am.gov.br/services2/services/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/RecepcaoEvento", "https://homnfe.sefaz.am.gov.br/services2/services/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfe.sefaz.am.gov.br/services2/services/RecepcaoEvento4", "https://homnfe.sefaz.am.gov.br/services2/services/RecepcaoEvento4"},
		},
	},
	"NFe/AN": {
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://www.nfe.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.nfe.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V310: urlPorAmbiente{"https://www.nfe.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.nfe.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V400: urlPorAmbiente{"https://www.nfe.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx", "https://hom1.nfe.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx"},
		},
		ServicoDistribuicaoDFe: {
			V200: urlPorAmbiente{"https://www1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx", "https://hom1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx"},
			V310: urlPorAmbiente{"https://www1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx", "https://hom1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx"},
			V400: urlPorAmbiente{"https://www1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx", "https://hom1.nfe.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NfeAutorizacao/NfeAutorizacao.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NfeAutorizacao/NfeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeAutorizacao4/NFeAutorizacao4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeAutorizacao4/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NfeRetAutorizacao/NfeRetAutorizacao.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NfeRetAutorizacao/NfeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRecepcao2.asmx"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/NfeRetRecepcao2.asmx"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/nfeinutilizacao2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/nfeinutilizacao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NfeInutilizacao/NfeInutilizacao.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NfeInutilizacao/NfeInutilizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeInutilizacao4/NFeInutilizacao4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeInutilizacao4/NFeInutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/CadConsultaCadastro2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/CadConsultaCadastro2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/nfenw/CadConsultaCadastro2.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/nfenw/CadConsultaCadastro2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/CadConsultaCadastro4/CadConsultaCadastro4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/CadConsultaCadastro4/CadConsultaCadastro4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/sre/recepcaoevento.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/sre/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/sre/recepcaoevento.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/sre/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ba.gov.br/webservices/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx", "https://hnfe.sefaz.ba.gov.br/webservices/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeAutorizacao?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeAutorizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeAutorizacao4?wsdl"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetAutorizacao?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeRetAutorizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeRetAutorizacao4?wsdl"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeInutilizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeInutilizacao4?wsdl"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/CadConsultaCadastro4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/CadConsultaCadastro4?wsdl"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeRecepcaoEvento4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeRecepcaoEvento4?wsdl"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeAutorizacao", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeAutorizacao4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRetAutorizacao", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRetAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeRetAutorizacao4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeRetAutorizacao4"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRecepcao2"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeRetRecepcao2"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeInutilizacao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeInutilizacao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NfeInutilizacao2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeInutilizacao4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/cadconsultacadastro2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/cadconsultacadastro2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/cadconsultacadastro2", "https://hnfe.fazenda.mg.gov.br/nfe2/services/cadconsultacadastro2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/CadConsultaCadastro4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/RecepcaoEvento", "https://hnfe.fazenda.mg.gov.br/nfe2/services/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/RecepcaoEvento", "https://hnfe.fazenda.mg.gov.br/nfe2/services/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfe.fazenda.mg.gov.br/nfe2/services/NFeRecepcaoEvento4", "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeRecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeAutorizacao", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeAutorizacao4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRetAutorizacao", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRetAutorizacao"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeRetAutorizacao4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeRetAutorizacao4"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRecepcao2"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRetRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRetRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRetRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRetRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeRetRecepcao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeRetRecepcao2"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeInutilizacao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeInutilizacao2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/NfeInutilizacao2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeInutilizacao4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/CadConsultaCadastro2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/CadConsultaCadastro2"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/CadConsultaCadastro2", "https://homologacao.nfe.ms.gov.br/homologacao/services2/CadConsultaCadastro2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4", "https://hom.nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/RecepcaoEvento", "https://homologacao.nfe.ms.gov.br/homologacao/services2/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/RecepcaoEvento", "https://homologacao.nfe.ms.gov.br/homologacao/services2/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4", "https://hom.nfe.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetAutorizacao4"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRecepcao2?wsdl"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeRetRecepcao2?wsdl"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/NfeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4?wsdl", "https://homologacao.sefaz.mt.gov.br/nfews/v2/services/RecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeAutorizacao?wsdl", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeAutorizacao4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeAutorizacao4?wsdl"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRetAutorizacao?wsdl", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRetAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeRetAutorizacao4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeRetAutorizacao4?wsdl"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRecepcao2"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeRetRecepcao2"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeInutilizacao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeInutilizacao2"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NfeInutilizacao2", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeInutilizacao4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeInutilizacao4?wsdl"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/CadConsultaCadastro2", ""},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/CadConsultaCadastro2", ""},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/CadConsultaCadastro4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/CadConsultaCadastro4?wsdl"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/RecepcaoEvento", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/RecepcaoEvento", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfe.sefaz.pe.gov.br/nfe-service/services/NFeRecepcaoEvento4", "https://nfehomolog.sefaz.pe.gov.br/nfe-service/services/NFeRecepcaoEvento4?wsdl"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeAutorizacao3?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeAutorizacao3?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeRetAutorizacao3?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeRetAutorizacao3?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeInutilizacao3?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeInutilizacao3?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeInutilizacao4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeInutilizacao4?wsdl"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/CadConsultaCadastro2?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/CadConsultaCadastro2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/CadConsultaCadastro2?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/CadConsultaCadastro2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/CadConsultaCadastro4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/CadConsultaCadastro4?wsdl"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeRecepcaoEvento?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeRecepcaoEvento?wsdl"},
			V310: urlPorAmbiente{"https://nfe.fazenda.pr.gov.br/nfe/NFeRecepcaoEvento?wsdl", "https://homologacao.nfe.fazenda.pr.gov.br/nfe/NFeRecepcaoEvento?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefa.pr.gov.br/nfe/NFeRecepcaoEvento4?wsdl", "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeRecepcaoEvento4?wsdl"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeInutilizacao/NfeInutilizacao2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/NfeInutilizacao/NfeInutilizacao2.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://cad.sefazrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx", "https://cad.sefazrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx"},
			V310: urlPorAmbiente{"https://cad.sefazrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx", "https://cad.sefazrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx"},
			V400: urlPorAmbiente{"https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx", "https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfe.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx", "https://nfe-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeautorizacao.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeautorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nferetautorizacao.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferetautorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRecepcao2.asmx"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferetrecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRetRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferetrecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRetRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nferetrecepcao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/NfeRetRecepcao2.asmx"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/nfeweb/services/nfeinutilizacao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/nfeweb/services/nfeinutilizacao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeinutilizacao2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/RecepcaoEvento.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/RecepcaoEvento.asmx"},
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/RecepcaoEvento.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/RecepcaoEvento.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeAutorizacao/NfeAutorizacao.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeAutorizacao/NfeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRetAutorizacao/NfeRetAutorizacao.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRetAutorizacao/NfeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeInutilizacao2/NfeInutilizacao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeInutilizacao2/NfeInutilizacao2.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NfeInutilizacao2/NfeInutilizacao2.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NfeInutilizacao2/NfeInutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeInutilizacao4/NFeInutilizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeInutilizacao4/NFeInutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl", "https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl"},
			V310: urlPorAmbiente{"https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl", "https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl"},
			V400: urlPorAmbiente{"https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl", "https://sistemas.sefaz.ma.gov.br/wscadastro/CadConsultaCadastro2?wsdl"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.sefazvirtual.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V310: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.sefazvirtual.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeAutorizacao/NfeAutorizacao.asmx", "https://hom.svc.fazenda.gov.br/NfeAutorizacao/NfeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeAutorizacao4/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRetAutorizacao/NfeRetAutorizacao.asmx", "https://hom.svc.fazenda.gov.br/NfeRetAutorizacao/NfeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRecepcao2/NfeRecepcao2.asmx"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
			V400: urlPorAmbiente{"https://www.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx", "https://hom.svc.fazenda.gov.br/NfeRetRecepcao2/NfeRetRecepcao2.asmx"},
		},
		ServicoInutilizacao: {
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeInutilizacao4/NFeInutilizacao4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeInutilizacao4/NFeInutilizacao4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://www.svc.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.svc.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V310: urlPorAmbiente{"https://www.svc.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx", "https://hom.svc.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx"},
			V400: urlPorAmbiente{"https://www.sefazvirtual.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx", "https://hom.sefazvirtual.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx", "https://cad-homologacao.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx"},
			V310: urlPorAmbiente{"https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx", "https://cad-homologacao.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro2.asmx"},
			V400: urlPorAmbiente{"https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx", "https://cad-homologacao.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx", "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx"},
		},
	},
	"NFCe/AM": {
		ServicoStatus: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeStatusServico2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeStatusServico2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeStatusServico4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeConsulta2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeConsulta2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeConsulta4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeConsulta4"},
		},
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeAutorizacao", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeAutorizacao"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeAutorizacao4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRetAutorizacao", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRetAutorizacao"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRetAutorizacao4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRetAutorizacao4"},
		},
		ServicoRecepcaoLote: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRecepcao2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRecepcao2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRecepcao2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRecepcao2"},
		},
		ServicoRetRecepcaoLote: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRetRecepcao2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRetRecepcao2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeRetRecepcao2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeRetRecepcao2"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeInutilizacao2", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/NfeInutilizacao4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/NfeInutilizacao4"},
		},
		ServicoAdministrarCSC: {
			V200: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/CscNFCe", "https://homnfce.sefaz.am.gov.br/nfce-services/services/CscNFCe"},
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/CscNFCe", "https://homnfce.sefaz.am.gov.br/nfce-services/services/CscNFCe"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/CscNFCe", "https://homnfce.sefaz.am.gov.br/nfce-services/services/CscNFCe"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/RecepcaoEvento", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/RecepcaoEvento", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfce.sefaz.am.gov.br/nfce-services/services/RecepcaoEvento4", "https://homnfce.sefaz.am.gov.br/nfce-services-nac/services/RecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeAutorizacao?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeAutorizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeAutorizacao4?wsdl"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetAutorizacao?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeRetAutorizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeRetAutorizacao4?wsdl"},
		},
		ServicoRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRecepcao2?wsdl"},
		},
		ServicoRetRecepcaoLote: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeRetRecepcao2?wsdl"},
		},
		ServicoInutilizacao: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/NfeInutilizacao2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeInutilizacao4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeInutilizacao4?wsdl"},
		},
		ServicoConsultaCadastro: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CadConsultaCadastro2?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/CadConsultaCadastro4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/CadConsultaCadastro4?wsdl"},
		},
		ServicoAdministrarCSC: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/CscNFCe?wsdl"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl"},
			V310: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/v2/RecepcaoEvento?wsdl"},
			V400: urlPorAmbiente{"https://nfe.sefaz.go.gov.br/nfe/services/NFeRecepcaoEvento4?wsdl", "https://homolog.sefaz.go.gov.br/nfe/services/NFeRecepcaoEvento4?wsdl"},
		},
	},
	"NFCe/MG": {
		ServicoStatus: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeStatusServico4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeStatusServico4"},
		},
		ServicoConsultaProtocolo: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeConsultaProtocolo4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeConsultaProtocolo4"},
		},
		ServicoAutorizacao: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeAutorizacao4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeRetAutorizacao4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeRetAutorizacao4"},
		},
		ServicoInutilizacao: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeInutilizacao4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeInutilizacao4"},
		},
		ServicoRecepcaoEvento: {
			V400: urlPorAmbiente{"https://nfce.fazenda.mg.gov.br/nfce/services/NFeRecepcaoEvento4", "https://hnfce.fazenda.mg.gov.br/nfce/services/NFeRecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/NfeAutorizacao", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/NfeAutorizacao"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeAutorizacao4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/NfeRetAutorizacao", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/NfeRetAutorizacao"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeRetAutorizacao4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeRetAutorizacao4"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/NfeInutilizacao2", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/NfeInutilizacao2"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeInutilizacao4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V310: urlPorAmbiente{"https://nfe.fazenda.ms.gov.br/producao/services2/CadConsultaCadastro2", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/CadConsultaCadastro2"},
			V400: urlPorAmbiente{"https://nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4", "https://hom.nfe.sefaz.ms.gov.br/ws/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/RecepcaoEvento", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/RecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfce.fazenda.ms.gov.br/producao/services2/RecepcaoEvento", "https://homologacao.nfce.fazenda.ms.gov.br/homologacao/services2/RecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfce.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4", "https://hom.nfce.sefaz.ms.gov.br/ws/NFeRecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeAutorizacao?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeAutorizacao4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeRetAutorizacao?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeRetAutorizacao?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeRetAutorizacao4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeRetAutorizacao4"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeInutilizacao2?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeInutilizacao2?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/NfeInutilizacao4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/NfeInutilizacao4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento?wsdl"},
			V310: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento?wsdl", "https://homologacao.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento?wsdl"},
			V400: urlPorAmbiente{"https://nfce.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento4", "https://homologacao.sefaz.mt.gov.br/nfcews/services/RecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeAutorizacao3", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeAutorizacao3"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeAutorizacao4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeAutorizacao4"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeRetAutorizacao3", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeRetAutorizacao3"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeRetAutorizacao4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeRetAutorizacao4"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeInutilizacao3", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeInutilizacao3"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeInutilizacao4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeInutilizacao4"},
		},
		ServicoConsultaCadastro: {
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/CadConsultaCadastro4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/CadConsultaCadastro4"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeRecepcaoEvento", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeRecepcaoEvento"},
			V310: urlPorAmbiente{"https://nfce.fazenda.pr.gov.br/nfce/NFeRecepcaoEvento", "https://homologacao.nfce.fazenda.pr.gov.br/nfce/NFeRecepcaoEvento"},
			V400: urlPorAmbiente{"https://nfce.sefa.pr.gov.br/nfce/NFeRecepcaoEvento4", "https://homologacao.nfce.sefa.pr.gov.br/nfce/NFeRecepcaoEvento4"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfce.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx", "https://nfce-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/nfeautorizacao.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/nfeautorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeAutorizacao4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/nferetautorizacao.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/nferetautorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeRetAutorizacao4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/nfeinutilizacao2.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeInutilizacao4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeInutilizacao4.asmx"},
		},
		ServicoConsultaCadastro: {
			V310: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro2.asmx"},
			V400: urlPorAmbiente{"https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx", "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/recepcaoevento.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/recepcaoevento.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfce.fazenda.sp.gov.br/ws/NFeRecepcaoEvento4.asmx", "https://homologacao.nfce.fazenda.sp.gov.br/ws/NFeRecepcaoEvento4.asmx"},
		},
		ServicoEventoEPEC: {
			V200: urlPorAmbiente{"https://nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx", "https://homologacao.nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx"},
			V310: urlPorAmbiente{"https://nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx", "https://homologacao.nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx"},
			V400: urlPorAmbiente{"https://nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx", "https://homologacao.nfce.epec.fazenda.sp.gov.br/EPECws/RecepcaoEPEC.asmx"},
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
		ServicoAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx"},
		},
		ServicoRetAutorizacao: {
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx"},
		},
		ServicoInutilizacao: {
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao2.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx"},
		},
		ServicoRecepcaoEvento: {
			V200: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V310: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento.asmx"},
			V400: urlPorAmbiente{"https://nfce.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx", "https://nfce-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx"},
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
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF"},
		},
		ServicoConsultaProtocolo: {
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsulta4/nfeConsultaNF", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsulta4/nfeConsultaNF"},
		},
		ServicoAutorizacao: {
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote"},
		},
		ServicoRetAutorizacao: {
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote"},
		},
		ServicoInutilizacao: {
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF"},
		},
		ServicoRecepcaoEvento: {
			V400: urlPorAmbiente{"http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4/nfeRecepcaoEventoNF", "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4/nfeRecepcaoEventoNF"},
		},
	},
}

// versaoFixaPorServico lista os servicos cuja versao de MENSAGEM nao
// acompanha a versao do leiaute da nota.
//
// Como ela e decidida no ACBr: o LerServicoDeParams devolve a versao da
// CHAVE que achou no `.ini` e, logo depois, tenta substitui-la pela versao
// do SCHEMA instalado (ACBrDFe.pas:580-590). Quando o schema da versao
// pedida existe, ele vence e desfaz a queda -- e o caso de consStatServ,
// consSitNFe, inutNFe, nfe e retEnviNFe, que a SEFAZ publica em 3.10 e
// 4.00. Sobram dois servicos:
//
//	consCad        existe SO em 2.00 (consCad_v2.00.xsd); nao houve versao nova
//	admCscNFCe     nao tem schema nenhum, entao vale a chave do `.ini`: 1.00
//	envEvento      existe SO em 1.00, mesmo onde o `.ini` publica a URL da
//	               4.00 -- os eventos tem leiaute PROPRIO, que nao acompanha
//	               o da NF-e. Vale tambem para os da Reforma Tributaria
//	eventoEPEC     idem, 1.00
//	distDFeInt     1.01. Aqui existem os dois schemas (1.00 e 1.01), e o
//	               AchaArquivoSchema pegaria o maior <= a versao pedida --
//	               que para 4.00 e o 1.01. O `.ini` tambem so publica a
//	               chave `_1.01`, entao os dois caminhos concordam
//
// Para esses dois, emitir a versao do leiaute da nota e rejeicao certa: a
// SEFAZ valida a mensagem contra o XSD do servico, nao contra o da NF-e.
//
// Se a SEFAZ publicar um consCad 4.00, a entrada sai daqui -- confira a
// pasta de schemas do ACBr antes de mexer.
var versaoFixaPorServico = map[Servico]string{
	ServicoConsultaCadastro: "2.00",
	ServicoAdministrarCSC:   "1.00",
	ServicoRecepcaoEvento:   "1.00",
	ServicoEventoEPEC:       "1.00",
	ServicoDistribuicaoDFe:  "1.01",
}

// VersaoDaMensagem devolve a versao que vai no atributo `versao` da
// mensagem e no <versaoDados> do cabecalho SOAP.
//
// Para quase todo servico e a versao configurada; para os dois de
// versaoFixaPorServico, e a do proprio servico.
func VersaoDaMensagem(servico Servico, configurada VersaoDF) string {
	if fixa, ok := versaoFixaPorServico[servico]; ok {
		return fixa
	}
	return pcn.FormatarVersaoXML(configurada.Float())
}
