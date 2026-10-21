#!/usr/bin/env python3
"""Confere os geradores de EVENTO do porte contra o Delphi de origem.

Os eventos sao o lugar do porte onde mais nome se repete. Tres elementos
distintos saem com o mesmo nome de tag, e o `gControleEstoque` tem SETE
formas, cada uma com um par de tags proprio. Nada disso aparece num
round-trip: o leitor acha a tag pelo nome, e o nome do elemento PAI e o
mesmo nos dois lados da troca.

O que o script confere, por evento:

  1. o CONJUNTO de tags emitidas -- faltando ou sobrando;
  2. a ORDEM delas, que e xs:sequence no XSD;
  3. o literal fixo, quando o Delphi passa uma constante no lugar do campo
     (o `tpAutor := '1'` de seis eventos e o `indQuitacao := '1'`).

Ele SEGUE os helpers dos dois lados -- os `Gerar_gXxx` do Delphi e as
funcoes locais do Go --, porque e justamente nos grupos repetidos que mora
o risco. Um script que parasse no corpo do evento acusaria "sobra" em
metade dos casos e nao conferiria nada do que importa.

Limites, que importam tanto quanto o que ele faz:

  - ele le os dois lados por EXPRESSAO REGULAR, nao por parser. Uma tag
    montada por concatenacao, dos dois lados, escapa;
  - ele nao sabe nada sobre obrigatoriedade nem sobre tipo de campo. Um
    `tcDe2` portado como `tcDe4` passa aqui -- o conferir-gerador-dfe.py e
    quem cuida disso;
  - a comparacao de ORDEM so olha as tags que os dois lados tem. Uma tag
    faltando ja foi acusada antes, e compara-la de novo na ordem so
    produziria ruido.

Uso:

    py ferramentas/conferir-eventos-dfe.py <ACBrNFe.EnvEvento.pas> <dir Go>

    py ferramentas/conferir-eventos-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/Servicos/ACBrNFe.EnvEvento.pas" \\
       packages/nfe

Saida: uma linha por divergencia e um resumo. Codigo de saida 1 se houver
divergencia, para encaixar em CI.
"""

import io
import os
import re
import sys

# Funcao Delphi -> constante do TipoEvento que o `case` do Go usa.
# Acrescente a entrada quando um evento novo ganhar gerador; o script acusa
# o que estiver so de um lado.
EQUIVALENTES = {
    "Gerar_Evento_CCe": "TeCCe",
    "Gerar_Evento_Cancelamento": "TeCancelamento",
    "Gerar_Evento_CancSubstituicao": "TeCancSubst",
    "Gerar_Evento_ManifDestCiencia": "TeManifDestCiencia",
    "Gerar_Evento_ManifDestConfirmacao": "TeManifDestConfirmacao",
    "Gerar_Evento_ManifDestDesconhecimento": "TeManifDestDesconhecimento",
    "Gerar_Evento_ManifDestOperNaoRealizada": "TeManifDestOperNaoRealizada",
    "Gerar_Evento_ComprEntrega": "TeComprEntregaNFe",
    "Gerar_Evento_CancComprEntrega": "TeCancComprEntregaNFe",
    "Gerar_Evento_AtorInteressado": "TeAtorInteressadoNFe",
    "Gerar_Evento_InsucessoEntrega": "TeInsucessoEntregaNFe",
    "Gerar_Evento_CancInsucessoEntrega": "TeCancInsucessoEntregaNFe",
    "Gerar_Evento_ConciliacaoFinanceira": "TeConcFinanceira",
    "Gerar_Evento_CancConciliacaoFinanceira": "TeCancConcFinanceira",
    "Gerar_Evento_PedProrrogacao": "TePedProrrog1",
    "Gerar_Evento_CancPedProrrogacao": "TeCanPedProrrog1",
    "Gerar_Evento_CancelamentoGenerico": "TeCancGenerico",
    "Gerar_Evento_PagIntegLibCredPresAdq": "TePagIntegLibCredPresAdq",
    "Gerar_Evento_ImportALCZFM": "TeImporALCZFM",
    "Gerar_Evento_PerecPerdaRouboTranspContrForn":
        "TePerecPerdaRouboFurtoTranspContratFornec",
    "Gerar_Evento_PerecPerdaRouboTranspContrAdq":
        "TePerecPerdaRouboFurtoTranspContratAqu",
    "Gerar_Evento_FornecNaoRealizPagAntec": "TeFornecNaoRealizPagAntec",
    "Gerar_Evento_AtualizacaoDataPrevisaoEntrega":
        "TeAtualizacaoDataPrevisaoEntrega",
    "Gerar_Evento_SolicApropCredPres": "TeSolicApropCredPres",
    "Gerar_Evento_DestItemConsPessoal": "TeDestItemConsPessoal",
    "Gerar_Evento_AceiteDebitoApuracaoNotaCredito":
        "TeAceiteDebitoApuracaoNotaCredito",
    "Gerar_Evento_ImobilizacaoItem": "TeImobilizacaoItem",
    "Gerar_Evento_SolicApropCredCombustivel": "TeSolicApropCredCombustivel",
    "Gerar_Evento_SolicApropCredBensServicos": "TeSolicApropCredBensServicos",
    "Gerar_Evento_ManifestacaoPedidoTransfCreditoIBSCBSOperacoesSucessao":
        "TeManifPedTransfCredIBSSucessao",
    "Gerar_Evento_EPEC": "TeEPECNFe",
}

# Evento -> tipo Delphi do parametro de Gerar_gControleEstoque.
#
# O Delphi tem SETE sobrecargas com o mesmo nome, distinguidas so pelo TIPO
# do parametro -- e nenhuma analise por expressao regular resolve isso. A
# tabela abaixo declara a correspondencia, e a verificacao 3 confere as
# tags de cada uma contra o Go. E o unico lugar do script onde a ligacao
# entre os dois lados e afirmada a mao, e e deliberado: o resto o script
# descobre sozinho.
CONTROLE_ESTOQUE = {
    "TeDestItemConsPessoal": "TgControleEstoque",
    "TeImporALCZFM": "TgControleEstoqueZFM",
    "TePerecPerdaRouboFurtoTranspContratAqu": "TgControleEstoquePerecimento",
    "TePerecPerdaRouboFurtoTranspContratFornec":
        "TgControleEstoquePerecimentoForn",
    "TeImobilizacaoItem": "TgControleEstoqueImobilizacao",
    "TeSolicApropCredCombustivel": "TgControleEstoqueComb",
    "TeFornecNaoRealizPagAntec": "TgControleEstoqueItemNaoFornecido",
}

# Tags que o porte emite por um caminho condicional que o script nao
# reproduz. Cada entrada e uma decisao registrada, nao um atalho.
TAGS_IGNORADAS = {
    # O EPEC escolhe entre Gerar_DestNFe e Gerar_DestNFCe conforme o modelo
    # da chave, e os dois emitem o grupo `dest` com conteudo diferente. O
    # script veria a uniao dos dois como "sobra". A diferenca esta travada
    # em TestEnderecoDoEvento e nos testes do gerador.
    "Gerar_Evento_EPEC": {"dest", "UF", "CNPJ", "CPF", "idEstrangeiro",
                          "vNF", "vICMS", "vST"},
}


def ler(caminho):
    if not os.path.exists(caminho):
        sys.exit("arquivo nao encontrado: " + caminho)
    for enc in ("cp1252", "utf-8"):
        try:
            return io.open(caminho, encoding=enc).read()
        except UnicodeDecodeError:
            continue
    sys.exit("nao consegui decodificar " + caminho)


def corpo_da_funcao(fonte, nome):
    """Devolve o corpo de `function TEventoNFe.<nome>` ate o `end;` final."""
    m = re.search(r"^function TEventoNFe\." + re.escape(nome) + r"\b",
                  fonte, re.M)
    if not m:
        return None
    fim = fonte.find("\nend;", m.end())
    return fonte[m.end():fim if fim > 0 else len(fonte)]


def tags_do_delphi(fonte, nome, vistos=None):
    """Tags emitidas por `nome`, na ordem, seguindo os helpers que chama."""
    if vistos is None:
        vistos = set()
    if nome in vistos:
        return []
    vistos.add(nome)

    corpo = corpo_da_funcao(fonte, nome)
    if corpo is None:
        return []

    saida = []
    for linha in corpo.splitlines():
        if linha.lstrip().startswith("//"):
            continue
        for m in re.finditer(
                r"AddNode\s*\(\s*tc\w+\s*,\s*'[^']*'\s*,\s*'(\w+)'", linha):
            saida.append(m.group(1))
        for m in re.finditer(r"CreateElement\('(\w+)'\)", linha):
            saida.append(m.group(1))
        for m in re.finditer(r"\b(Gerar_\w+)\s*\(", linha):
            saida.extend(tags_do_delphi(fonte, m.group(1), vistos))
    # O detEvento e o envelope, nao um campo.
    return [t for t in saida if t != "detEvento"]


def literais_do_delphi(fonte, nome, vistos=None):
    """Tags cujo valor e um LITERAL no Delphi, e nao um campo."""
    if vistos is None:
        vistos = set()
    if nome in vistos:
        return {}
    vistos.add(nome)

    corpo = corpo_da_funcao(fonte, nome)
    if corpo is None:
        return {}

    saida = {}
    for m in re.finditer(
            r"AddNode\s*\(\s*tc\w+\s*,\s*'[^']*'\s*,\s*'(\w+)'\s*,"
            r"\s*\d+\s*,\s*\d+\s*,\s*\d+\s*,\s*'([^']*)'\s*\)", corpo):
        saida[m.group(1)] = m.group(2)
    for m in re.finditer(r"\b(Gerar_\w+)\s*\(", corpo):
        saida.update(literais_do_delphi(fonte, m.group(1), vistos))
    return saida


def corpo_do_case_go(fonte, constante):
    """Devolve o trecho do `case ... pcn.<constante> ...:` ate o proximo.

    A constante pode NAO ser a primeira do `case`: a confirmacao e a
    ciencia do destinatario dividem um ramo so, e os dois PedProrrog
    tambem. Ancorar em `case pcn.X` perderia a segunda de cada par.
    """
    for m in re.finditer(r"^\tcase ([^\n]*?):$", fonte, re.M):
        if re.search(r"\bpcn\." + re.escape(constante) + r"\b", m.group(1)):
            resto = fonte[m.end():]
            prox = re.search(r"^\t(?:case |default:)", resto, re.M)
            return resto[:prox.start()] if prox else resto
    return None


def funcoes_go(fonte):
    """Mapa nome -> corpo de toda funcao do fonte Go."""
    saida = {}
    for m in re.finditer(r"^func (?:\([^)]*\) )?(\w+)\(", fonte, re.M):
        fim = re.search(r"^}", fonte[m.end():], re.M)
        if fim:
            saida[m.group(1)] = fonte[m.end():m.end() + fim.start()]
    return saida


def tags_do_go(trecho, funcoes, vistos=None):
    """Tags emitidas pelo trecho, seguindo as funcoes locais que chama.

    Duas sutilezas do lado Go:

      - o conjunto de visitados e indexado por (funcao, argumento). O
        gerarCredPresPorTributo e chamado DUAS vezes com literais
        diferentes, e colapsar as duas perderia o gCBSCredPres;
      - quando a funcao chamada monta o elemento com nome PARAMETRIZADO
        (`pcn.NovoElem(nome)`), o nome real e o literal do ponto de
        chamada.
    """
    if vistos is None:
        vistos = set()
    tags = []
    for linha in trecho.splitlines():
        if linha.lstrip().startswith("//"):
            continue
        for m in re.finditer(r'pcn\.Node\w+\(\s*"(\w+)"', linha):
            tags.append(m.group(1))
        for m in re.finditer(r'pcn\.NovoElem\(\s*"(\w+)"', linha):
            tags.append(m.group(1))
        for m in re.finditer(r'\b([a-z]\w*)\(\s*(?:"(\w+)")?', linha):
            nome, literal = m.group(1), m.group(2) or ""
            if nome not in funcoes or (nome, literal) in vistos:
                continue
            vistos.add((nome, literal))
            corpo = funcoes[nome]
            if literal and re.search(r"pcn\.NovoElem\(\s*[a-z]\w*\)", corpo):
                tags.append(literal)
            tags.extend(tags_do_go(corpo, funcoes, vistos))
    return colapsar_repetidas([t for t in tags if t != "detEvento"])


def colapsar_repetidas(tags):
    """Funde repeticoes IMEDIATAS do mesmo nome de tag.

    Um helper com dois ramos mutuamente exclusivos -- o nodeDescEvento, que
    escolhe entre o filtro com e sem acentos -- aparece aqui como a mesma
    tag duas vezes seguidas, porque o script ve o codigo e nao a execucao.

    A fusao e segura porque nenhum evento emite a MESMA tag duas vezes em
    sequencia: os casos de nome repetido do leiaute (vIBS/vCBS do
    gPerecimento do fornecedor, pCredPres/vCredPres dos dois subgrupos de
    credito presumido) sempre tem outra tag no meio.
    """
    saida = []
    for t in tags:
        if not saida or saida[-1] != t:
            saida.append(t)
    return saida


def literais_do_go(trecho, funcoes, constantes, vistos=None):
    """Tags cujo valor e uma constante no Go, seguindo os helpers."""
    if vistos is None:
        vistos = set()
    saida = {}
    for linha in trecho.splitlines():
        if linha.lstrip().startswith("//"):
            continue
        m = re.search(r'pcn\.NodeStr\(\s*"(\w+)"\s*,\s*(\w+)\s*,', linha)
        if m and m.group(2) in constantes:
            saida[m.group(1)] = constantes[m.group(2)]
        m = re.search(r'pcn\.NodeStr\(\s*"(\w+)"\s*,\s*"([^"]*)"\s*,', linha)
        if m:
            saida[m.group(1)] = m.group(2)
        for m in re.finditer(r"\b([a-z]\w*)\(", linha):
            nome = m.group(1)
            if nome in funcoes and nome not in vistos:
                vistos.add(nome)
                saida.update(literais_do_go(funcoes[nome], funcoes,
                                            constantes, vistos))
    return saida


def sobrecargas_do_controle_estoque(fonte):
    """Mapa tipo do parametro -> tags, das SETE sobrecargas do Delphi."""
    saida = {}
    for m in re.finditer(
            r"^function TEventoNFe\.Gerar_gControleEstoque\s*\(?\s*\n?\s*"
            r"gControleEstoque:\s*(\w+)\s*\)", fonte, re.M):
        fim = fonte.find("\nend;", m.end())
        corpo = fonte[m.end():fim if fim > 0 else len(fonte)]
        tags = re.findall(
            r"AddNode\s*\(\s*tc\w+\s*,\s*'[^']*'\s*,\s*'(\w+)'", corpo)
        saida[m.group(1)] = tags
    return saida


def controle_estoque_do_go(trecho, funcoes, vistos=None):
    """Tags emitidas DENTRO do pcn.NovoElem("gControleEstoque") do trecho.

    O recorte e por PROFUNDIDADE de parenteses, nao por delimitador de
    texto: o bloco e um encadeamento de `.Filho(...)` e procurar o
    primeiro `))` pararia na primeira tag.
    """
    if vistos is None:
        vistos = set()

    marcador = 'pcn.NovoElem("gControleEstoque")'
    i = trecho.find(marcador)
    if i >= 0:
        # Comeca a contar na abertura do NovoElem e segue ate ela fechar.
        j = i + len(marcador)
        profundidade = 0
        for k in range(j, len(trecho)):
            if trecho[k] == "(":
                profundidade += 1
            elif trecho[k] == ")":
                if profundidade == 0:
                    break
                profundidade -= 1
        return re.findall(r'pcn\.Node\w+\(\s*"(\w+)"', trecho[j:k])

    for linha in trecho.splitlines():
        if linha.lstrip().startswith("//"):
            continue
        for m in re.finditer(r"\b([a-z]\w*)\(", linha):
            nome = m.group(1)
            if nome in funcoes and nome not in vistos:
                vistos.add(nome)
                achado = controle_estoque_do_go(funcoes[nome], funcoes, vistos)
                if achado:
                    return achado
    return []


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)

    delphi = ler(sys.argv[1])

    diretorio = sys.argv[2]
    go = ""
    for nome in ("evento_writer.go", "evento_writer_rtc.go"):
        caminho = os.path.join(diretorio, nome)
        if not os.path.exists(caminho):
            sys.exit("gerador Go nao encontrado: " + caminho)
        go += io.open(caminho, encoding="utf-8").read()

    falhas = 0
    conferidos = 0
    funcoes = funcoes_go(go)
    constantes = dict(re.findall(r'^const (\w+) = "([^"]*)"', go, re.M))
    sobrecargas = sobrecargas_do_controle_estoque(delphi)

    print("== conjunto e ordem das tags do detEvento ==")
    for funcao, constante in sorted(EQUIVALENTES.items()):
        if corpo_da_funcao(delphi, funcao) is None:
            falhas += 1
            print("  SUMIU    %s nao existe mais no Delphi;"
                  " a tabela EQUIVALENTES precisa de revisao" % funcao)
            continue

        trecho = corpo_do_case_go(go, constante)
        if trecho is None:
            falhas += 1
            print("  FALTA    o Go nao tem `case pcn.%s`" % constante)
            continue

        conferidos += 1
        ignoradas = set(TAGS_IGNORADAS.get(funcao, set()))

        # As tags de DENTRO do gControleEstoque saem desta verificacao: o
        # Delphi tem sete sobrecargas de mesmo nome e nenhuma regex as
        # distingue. Elas tem verificacao propria, a de numero 3.
        for tags in sobrecargas.values():
            ignoradas.update(tags)

        esperadas = [t for t in tags_do_delphi(delphi, funcao)
                     if t not in ignoradas]
        obtidas = [t for t in tags_do_go(trecho, funcoes)
                   if t not in ignoradas]

        faltando = [t for t in esperadas if t not in obtidas]
        sobrando = [t for t in obtidas if t not in esperadas]
        if faltando:
            falhas += 1
            print("  FALTA    %s: tag(s) %s" % (constante, faltando))
        if sobrando:
            falhas += 1
            print("  SOBRA    %s: tag(s) %s que o Delphi nao emite"
                  % (constante, sobrando))

        comuns_delphi = [t for t in esperadas if t in obtidas]
        comuns_go = [t for t in obtidas if t in esperadas]
        if comuns_delphi != comuns_go:
            falhas += 1
            print("  ORDEM    %s:" % constante)
            print("             delphi=%s" % comuns_delphi)
            print("             go    =%s" % comuns_go)

    print("  %d evento(s) conferido(s)" % conferidos)

    print("\n== literais fixos (campo ignorado pelo schema) ==")
    achados = 0
    for funcao, constante in sorted(EQUIVALENTES.items()):
        trecho = corpo_do_case_go(go, constante)
        if trecho is None or corpo_da_funcao(delphi, funcao) is None:
            continue
        esperados = literais_do_delphi(delphi, funcao)
        obtidos = literais_do_go(trecho, funcoes, constantes)
        for tag, valor in sorted(esperados.items()):
            achados += 1
            if obtidos.get(tag) != valor:
                falhas += 1
                print("  DIVERGE  %s/%s: delphi=%r, go=%r"
                      % (constante, tag, valor, obtidos.get(tag)))
    if achados:
        print("  %d literal(is) conferido(s)" % achados)
    else:
        print("  nenhum literal fixo encontrado")

    print("\n== as sete formas do gControleEstoque ==")
    if len(sobrecargas) != 7:
        falhas += 1
        print("  ACHEI %d sobrecarga(s) de Gerar_gControleEstoque no Delphi,"
              " esperadas 7 -- a extracao precisa de revisao"
              % len(sobrecargas))
    for constante, tipo in sorted(CONTROLE_ESTOQUE.items()):
        esperadas = sobrecargas.get(tipo)
        if esperadas is None:
            falhas += 1
            print("  SUMIU    o Delphi nao tem mais a sobrecarga %s" % tipo)
            continue
        trecho = corpo_do_case_go(go, constante)
        if trecho is None:
            falhas += 1
            print("  FALTA    o Go nao tem `case pcn.%s`" % constante)
            continue
        obtidas = controle_estoque_do_go(trecho, funcoes)
        if obtidas != esperadas:
            falhas += 1
            print("  DIVERGE  %s (%s):" % (constante, tipo))
            print("             delphi=%s" % esperadas)
            print("             go    =%s" % obtidas)
    print("  %d forma(s) conferida(s)" % len(CONTROLE_ESTOQUE))

    print()
    if falhas:
        print("%d divergencia(s) -- confira cada uma contra o .pas" % falhas)
        return 1
    print("OK: geradores de evento identicos ao Delphi")
    return 0


if __name__ == "__main__":
    sys.exit(main())
