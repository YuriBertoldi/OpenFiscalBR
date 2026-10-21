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
	"time"

	"github.com/openfiscalbr/openfiscalbr/packages/pcn"
	"github.com/openfiscalbr/openfiscalbr/packages/rtc"
)

// Modelo de dados dos eventos da NF-e. Porte de ACBrNFe.EventoClass.pas.
//
// A forma do DetEvento e deliberadamente ACHATADA: um unico struct carrega
// os campos de TODOS os trinta e um eventos, e cada gerador le so os seus.
// E exatamente o que o TDetEvento do ACBr faz (:461), e manter a forma
// importa por tres razoes:
//
//  1. o `.ini` e o JSON do ACBr leem para esse struto unico, e um porte com
//     uma struct por evento nao conseguiria carregar os arquivos existentes;
//  2. varios campos sao compartilhados por eventos sem parentesco nenhum --
//     `nProt` aparece no cancelamento e no pedido de prorrogacao, `verAplic`
//     em dezessete deles;
//  3. a escolha de QUAIS campos saem e do tpEvento, nao do tipo Go. Um
//     struct por evento daria a impressao de que o compilador vigia isso, e
//     ele nao vigia: quem vigia e o `switch` do gerador.

// Destinatario e o grupo dest do EPEC.
// Porte de TDestinatario (:57).
type Destinatario struct {
	UF            string // UF
	CNPJCPF       string // CNPJ ou CPF, escolhidos pelo TAMANHO
	IDEstrangeiro string // idEstrangeiro
	IE            string // IE
}

// ItemPedido e um item do pedido de prorrogacao (ICMS-ST).
// Porte de TitemPedidoCollectionItem (:70).
type ItemPedido struct {
	NumItem  int     // atributo numItem
	QtdeItem float64 // qtdeItem
}

// AutXMLEvento e um CNPJ/CPF autorizado no evento de ator interessado.
//
// NAO se confunde com o AutXML do documento (classes.go): aquele autoriza o
// download da NF-e, este declara o transportador do evento 110150.
type AutXMLEvento struct {
	CNPJCPF string // CNPJ ou CPF
}

// DetPagEvento e um pagamento do evento de conciliacao financeira.
// Porte de TdetPagCollectionItem (:108).
//
// Tem os mesmos nomes de tag do detPag do documento, mas e outro grupo: o
// da conciliacao traz dPag, CNPJPag, UFPag e CNPJReceb, que o do documento
// nao tem.
type DetPagEvento struct {
	IndPag    IndicadorPagamento // indPag
	TPag      FormaPagamento     // tPag
	XPag      string             // xPag
	VPag      float64            // vPag
	DPag      time.Time          // dPag
	CNPJPag   string             // CNPJPag
	UFPag     string             // UFPag
	CNPJIF    string             // CNPJIF
	TBand     BandeiraCartao     // tBand
	CAut      string             // cAut
	CNPJReceb string             // CNPJReceb
	UFReceb   string             // UFReceb
}

// IBSCBSCredPres e o par percentual/valor do credito presumido.
// Porte de TgIBSgCBS_CredPres (:146). O mesmo tipo serve ao gIBSCredPres e
// ao gCBSCredPres -- so o nome do elemento muda.
type IBSCBSCredPres struct {
	PCredPres float64 // pCredPres
	VCredPres float64 // vCredPres
}

// GCredPres e um item do evento de solicitacao de apropriacao de credito
// presumido. Porte de TgCredPresCollectionItem (:156).
//
// O elemento sai como <gCredPresOper>, nao <gCredPres> -- o nome da classe
// Delphi e o da tag divergem.
type GCredPres struct {
	NItem        int            // atributo nItem
	VBCCredPres  float64        // vBCCredPres
	CCredPres    rtc.CCredPres  // cCredPres
	GIBSCredPres IBSCBSCredPres // gIBSCredPres
	GCBSCredPres IBSCBSCredPres // gCBSCredPres
}

// Os sete grupos gControleEstoque.
//
// Todos saem com o MESMO nome de elemento -- <gControleEstoque> --, e cada
// um carrega um par de tags diferente. No Delphi a distincao e feita por
// SOBRECARGA: sete `Gerar_gControleEstoque` que diferem apenas no tipo do
// parametro (:2378, :2390, :2406, :2646, :2724, :2735, :3076).
//
// E a armadilha central deste arquivo. Um porte que escrevesse uma funcao
// generica emitiria <qConsumo> nos sete casos, e os seis errados passariam
// em qualquer round-trip -- o leitor acha a tag pelo nome, e o nome do
// elemento PAI e o mesmo.
type (
	// ControleEstoqueConsumo e o do evento de destinacao para consumo
	// pessoal: qConsumo / uConsumo.
	ControleEstoqueConsumo struct {
		QConsumo float64 // qConsumo
		UConsumo string  // uConsumo
	}

	// ControleEstoqueZFM e o da importacao em ALC/ZFM, e e o unico que NAO
	// nomeia a mercadoria na tag: qtde / unidade.
	ControleEstoqueZFM struct {
		Qtde    float64 // qtde
		Unidade string  // unidade
	}

	// ControleEstoquePerecimento e o do perecimento contratado pelo
	// ADQUIRENTE: qPerecimento / uPerecimento.
	ControleEstoquePerecimento struct {
		QPerecimento float64 // qPerecimento
		UPerecimento string  // uPerecimento
	}

	// ControleEstoquePerecimentoForn e o do perecimento contratado pelo
	// FORNECEDOR. Tem as mesmas duas tags do anterior MAIS vIBS e vCBS --
	// a unica das sete variantes com quatro campos.
	ControleEstoquePerecimentoForn struct {
		QPerecimento float64 // qPerecimento
		UPerecimento string  // uPerecimento
		VIBS         float64 // vIBS
		VCBS         float64 // vCBS
	}

	// ControleEstoqueImobilizacao: qImobilizado / uImobilizado.
	ControleEstoqueImobilizacao struct {
		QImobilizado float64 // qImobilizado
		UImobilizado string  // uImobilizado
	}

	// ControleEstoqueComb: qComb / uComb.
	ControleEstoqueComb struct {
		QComb float64 // qComb
		UComb string  // uComb
	}

	// ControleEstoqueItemNaoFornecido: qNaoFornecida / uNaoFornecida.
	ControleEstoqueItemNaoFornecido struct {
		QNaoFornecida float64 // qNaoFornecida
		UNaoFornecida string  // uNaoFornecida
	}
)

// GConsumo e um item do evento de destinacao para consumo pessoal.
// Porte de TgConsumoCollectionItem (:193).
type GConsumo struct {
	NItem           int                    // atributo nItem
	VIBS            float64                // vIBS
	VCBS            float64                // vCBS
	ControleEstoque ControleEstoqueConsumo // gControleEstoque
	DFeReferenciado rtc.DFeReferenciado    // DFeReferenciado
}

// GConsumoZFM e um item do evento de importacao em ALC/ZFM.
// Porte de TgConsumoZFMCollectionItem (:230).
//
// O elemento sai como <gConsumo>, igual ao do consumo pessoal -- o nome da
// classe Delphi diz ZFM, a tag nao.
type GConsumoZFM struct {
	NItem           int                // atributo nItem
	VIBS            float64            // vIBS
	VCBS            float64            // vCBS
	ControleEstoque ControleEstoqueZFM // gControleEstoque
}

// GPerecimento e um item do evento de perecimento no transporte contratado
// pelo ADQUIRENTE. Porte de TgPerecimentoCollectionItem (:266).
type GPerecimento struct {
	NItem           int                        // atributo nItem
	VIBS            float64                    // vIBS
	VCBS            float64                    // vCBS
	ControleEstoque ControleEstoquePerecimento // gControleEstoque
}

// GPerecimentoForn e o mesmo item para o transporte contratado pelo
// FORNECEDOR. Porte de TgPerecimentoFornCollectionItem (:399).
//
// Sai com o MESMO nome de elemento do anterior -- <gPerecimento> --, e o
// que os distingue e o gControleEstoque de dentro.
type GPerecimentoForn struct {
	NItem           int                            // atributo nItem
	VIBS            float64                        // vIBS
	VCBS            float64                        // vCBS
	ControleEstoque ControleEstoquePerecimentoForn // gControleEstoque
}

// GImobilizacao e um item do evento de imobilizacao.
// Porte de TgImobilizacaoCollectionItem (:302).
type GImobilizacao struct {
	NItem           int                         // atributo nItem
	VIBS            float64                     // vIBS
	VCBS            float64                     // vCBS
	ControleEstoque ControleEstoqueImobilizacao // gControleEstoque
}

// GConsumoComb e um item do evento de apropriacao de credito de
// combustivel. Porte de TgConsumoCombCollectionItem (:338).
type GConsumoComb struct {
	NItem           int                 // atributo nItem
	VIBS            float64             // vIBS
	VCBS            float64             // vCBS
	ControleEstoque ControleEstoqueComb // gControleEstoque
}

// GCredito e um item do evento de apropriacao de credito de bens e
// servicos. Porte de TgCreditoCollectionItem (:365).
//
// E o unico grupo da Reforma SEM gControleEstoque, e os valores tem nome
// proprio: vCredIBS e vCredCBS, nao vIBS e vCBS.
type GCredito struct {
	NItem    int     // atributo nItem
	VCredIBS float64 // vCredIBS
	VCredCBS float64 // vCredCBS
}

// GItemNaoFornecido e um item do evento de fornecimento nao realizado com
// pagamento antecipado. Porte de TgItemNaoFornecidoCollectionItem (:435).
type GItemNaoFornecido struct {
	NItem           int                             // atributo nItem
	VIBS            float64                         // vIBS
	VCBS            float64                         // vCBS
	ControleEstoque ControleEstoqueItemNaoFornecido // gControleEstoque
}

// DetEvento carrega os campos de TODOS os eventos.
// Porte de TDetEvento (:461). Ver a nota no topo do arquivo sobre a forma
// achatada.
type DetEvento struct {
	Versao     string // atributo versao
	DescEvento string // descEvento -- preenchido pelo gerador se vazio

	// Carta de correcao
	XCorrecao string // xCorrecao
	XCondUso  string // xCondUso

	// Cancelamento, cancelamento por substituicao e prorrogacao
	NProt string // nProt
	XJust string // xJust -- tambem na manifestacao do destinatario

	// Comuns a quase todos os eventos recentes
	COrgaoAutor int       // cOrgaoAutor
	TpAutor     TipoAutor // tpAutor
	VerAplic    string    // verAplic

	// EPEC
	DhEmi time.Time    // dhEmi
	TpNF  pcn.TipoNFe  // tpNF
	IE    string       // IE
	Dest  Destinatario // dest
	VNF   float64      // vNF
	VICMS float64      // vICMS
	VST   float64      // vST

	// Prorrogacao de ICMS-ST
	ItemPedido        []ItemPedido // itemPedido
	IDPedidoCancelado string       // idPedidoCancelado

	// Cancelamento por substituicao
	ChNFeRef string // chNFeRef

	// Comprovante de entrega
	DhEntrega         time.Time // dhEntrega
	NDoc              string    // nDoc
	XNome             string    // xNome
	LatGPS            float64   // latGPS
	LongGPS           float64   // longGPS
	HashComprovante   string    // hashComprovante
	DhHashComprovante time.Time // dhHashComprovante

	// Cancelamentos que apontam para outro evento
	NProtEvento string // nProtEvento

	// Ator interessado
	AutXML        []AutXMLEvento // autXML
	TpAutorizacao Autorizacao    // tpAutorizacao

	// Insucesso na entrega
	DhTentativaEntrega     time.Time // dhTentativaEntrega
	NTentativa             int       // nTentativa
	TpMotivo               TpMotivo  // tpMotivo
	XJustMotivo            string    // xJustMotivo
	HashTentativaEntrega   string    // hashTentativaEntrega
	DhHashTentativaEntrega time.Time // dhHashTentativaEntrega

	// UF do emitente, usada para resolver o fuso das datas do insucesso na
	// entrega. NAO sai no XML -- e so entrada do GetUTC (:794 e :817).
	UF string

	// Conciliacao financeira
	DetPag []DetPagEvento // detPag

	// Reforma Tributaria
	TpEventoAut       string              // tpEventoAut -- do cancelamento generico
	GCredPres         []GCredPres         // gCredPresOper
	GConsumo          []GConsumo          // gConsumo
	GConsumoZFM       []GConsumoZFM       // gConsumo (da ALC/ZFM)
	GPerecimento      []GPerecimento      // gPerecimento
	GPerecimentoForn  []GPerecimentoForn  // gPerecimento (do fornecedor)
	GImobilizacao     []GImobilizacao     // gImobilizacao
	GConsumoComb      []GConsumoComb      // gConsumoComb
	GCredito          []GCredito          // gCredito
	GItemNaoFornecido []GItemNaoFornecido // gItemNaoFornecido
	IndAceitacao      IndAceitacao        // indAceitacao
	DPrevEntrega      time.Time           // dPrevEntrega
}

// InfEvento e o grupo infEvento. Porte de TInfEvento (:588).
type InfEvento struct {
	// ID e montado pelo gerador a partir de tpEvento, chNFe e nSeqEvento;
	// preencher aqui nao tem efeito.
	ID string // atributo Id

	// COrgao e o orgao AUTORIZADOR do evento. Zero cai para o cUF da
	// chave (getcOrgao, :705) -- e e assim que a maioria dos eventos
	// funciona, porque o evento vai para a mesma SEFAZ da nota.
	//
	// A excecao e a manifestacao do destinatario, que vai para o Ambiente
	// Nacional e precisa de cOrgao 91.
	COrgao int

	TpAmb      pcn.TipoAmbiente // tpAmb
	CNPJ       string           // CNPJ ou CPF do autor
	ChNFe      string           // chNFe
	DhEvento   time.Time        // dhEvento
	TpEvento   pcn.TipoEvento   // tpEvento
	NSeqEvento int              // nSeqEvento

	// VersaoEvento e a tag <verEvento> do infEvento. NAO se confunde com
	// DetEvento.Versao, que e o ATRIBUTO `versao` do detEvento -- sao dois
	// campos do leiaute, ainda que na pratica carreguem o mesmo valor. Na
	// geracao os dois saem da versao do lote.
	VersaoEvento string

	DetEvento DetEvento // detEvento
}

// EventoNFe e o lote de eventos (envEvento).
// Porte de TEventoNFe (ACBrNFe.EnvEvento.pas).
type EventoNFe struct {
	// IDLote e o idLote. O ACBr o recebe como Int64 e nao o valida.
	IDLote int64

	// Versao e a do leiaute do EVENTO, que nao acompanha a da NF-e.
	//
	// Vale 1.00 para todos, inclusive os da Reforma Tributaria: a SEFAZ
	// publica um unico envEvento_v1.00.xsd. Vazio, o transmissor a
	// preenche com VersaoDaMensagem(ServicoRecepcaoEvento, ...).
	Versao string

	Evento []Evento
}

// Evento e um evento do lote.
type Evento struct {
	InfEvento InfEvento     // infEvento
	Signature pcn.Signature // Signature

	// XMLAssinado guarda o evento como saiu do assinador. Como na nota, a
	// assinatura cobre bytes: regerar invalidaria o digest.
	XMLAssinado string
}

// maximoEventosPorLote e o teto de :2522 do ACBr.
const maximoEventosPorLote = 20

// COrgaoAmbienteNacional e o 91, o orgao da manifestacao do destinatario.
//
// Os quatro eventos de manifestacao nao vao para a SEFAZ de origem: vao
// para o Ambiente Nacional, e o cOrgao tem de dizer isso. Deixar o getcOrgao
// cair para o cUF da chave manda a manifestacao para a UF errada.
const COrgaoAmbienteNacional = 91
