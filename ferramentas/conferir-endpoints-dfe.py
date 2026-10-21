#!/usr/bin/env python3
"""Confere a tabela de ENDPOINTS dos web services contra o Servicos.ini.

Enquanto conferir-urls-dfe.py cuida das URLs publicas (QR-Code, consulta do
consumidor), este cuida dos enderecos de TRANSMISSAO -- os que recebem o
documento.

Eles tem uma forma diferente, e a diferenca e estrutural:

Nem toda UF opera o proprio ambiente de autorizacao. A maioria delega, e o
`.ini` expressa isso com um redirecionamento `Usar=` de secao para secao.
Quem decide o endereco, portanto, nao e a UF do emitente: e o AUTORIZADOR
que atende aquela UF naquele modelo. O porte guarda as duas coisas
separadas -- um mapa UF -> autorizador e uma tabela de enderecos por
autorizador --, e este script confere as duas contra o `.ini`.

Confere tambem a premissa que permite essa separacao: a de que o
redirecionamento e sempre de SECAO INTEIRA. Se alguma UF passar a definir um
servico localmente e delegar outro, o modelo quebra e o script acusa.

Uso:

    py ferramentas/conferir-endpoints-dfe.py <servicos.ini> <tabela.go>

    py ferramentas/conferir-endpoints-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/ACBrNFeServicos.ini" \\
       packages/nfe/urls_servicos_tabela.go

Saida: uma linha por divergencia e um resumo. Codigo de saida 1 se houver
divergencia, para encaixar em CI.
"""

import collections
import io
import re
import sys

# Constante Go -> nome da chave no `.ini`. Acrescente aqui o servico novo
# quando a etapa que o transmite entrar.
SERVICOS = {
    "ServicoStatus": "NfeStatusServico",
    "ServicoConsultaProtocolo": "NfeConsultaProtocolo",
}

# Campo Go -> versao de leiaute que ele atende.
VERSOES = {"V200": 2.00, "V310": 3.10, "V400": 4.00}


def carregar_ini(caminho):
    try:
        linhas = io.open(caminho, encoding="cp1252").read().splitlines()
    except FileNotFoundError:
        sys.exit(f".ini do ACBr nao encontrado: {caminho}")

    sec, dados = None, collections.OrderedDict()
    for l in linhas:
        l = l.strip()
        if l.startswith("[") and l.endswith("]"):
            sec = l[1:-1]
            dados.setdefault(sec, {})
        elif sec and "=" in l and not l.startswith(";"):
            k, v = l.split("=", 1)
            dados[sec][k.strip()] = v.strip()
    return dados


def chave_de_params(dados, sessao, servico, versao):
    """Porte de LerServicoChaveDeParams (ACBrDFe.pas:456)."""
    kv = dados.get(sessao)
    if kv is None:
        return ""

    base = (servico + "_").upper()
    achada = 0.0
    chave = base + "%.2f" % versao

    for k in kv:
        if k.upper() == chave and kv[k]:
            achada = versao

    if achada == 0.0:
        for k in kv:
            K = k.upper()
            if K.startswith(base):
                try:
                    v = float(K[len(base):])
                except ValueError:
                    continue
                if achada < v <= versao:
                    achada, chave = v, K

    if achada > 0:
        for k in kv:
            if k.upper() == chave:
                return kv[k]
        return ""

    for k in kv:
        if k.upper() == servico.upper():
            return kv[k]
    return ""


def secao_que_define(dados, modelo, uf, amb, servico, versao):
    """Em qual secao a URL deste servico foi de fato encontrada."""
    sessao = f"{modelo}_{uf}_{amb}"
    vistas = []
    url = chave_de_params(dados, sessao, servico, versao)
    while not url:
        if sessao in vistas:
            return None, ""
        vistas.append(sessao)
        usar = dados.get(sessao, {}).get("Usar", "")
        if not usar:
            return None, ""
        sessao = usar
        url = chave_de_params(dados, sessao, servico, versao)
    return sessao, url


def action_do_ini(dados, modelo, uf, amb, servico, versao):
    """Resolve o override de SoapAction declarado pela chave `SoapAction=`.

    Porte de LerServicoDeParams (ACBrDFe.pas:546): a sessao pode apontar,
    pela chave `SoapAction=`, para OUTRA sessao que traz as actions por
    servico e versao. O que vem dali VENCE o que o codigo montaria.
    """
    sessao = f"{modelo}_{uf}_{amb}"
    vistas = []
    while sessao in dados:
        destino = dados[sessao].get("SoapAction", "")
        if destino:
            return chave_de_params(dados, destino, servico, versao)
        usar = dados[sessao].get("Usar", "")
        if not usar or sessao in vistas:
            break
        vistas.append(sessao)
        sessao = usar
    return ""


def ler_actions_go(caminho):
    """Extrai {(modelo, autorizador, servico, versao): (producao, homologacao)}
    da tabela de overrides de SoapAction."""
    fonte = re.sub(r"^\s*//.*$", "", io.open(caminho, encoding="utf-8").read(), flags=re.M)
    bloco = re.search(r"var soapActionPorAutorizador = .*?\n\}\n", fonte, re.S)
    if not bloco:
        return {}

    saida = {}
    chave = servico = None
    for linha in bloco.group(0).splitlines():
        m = re.match(r'\s*"(\w+)/([\w-]+)": \{', linha)
        if m:
            chave = (m.group(1), m.group(2))
            continue
        m = re.match(r"\s*(Servico\w+): \{", linha)
        if m:
            servico = m.group(1)
            continue
        m = re.match(r"\s*(V\d{3}): urlPorAmbiente\{", linha)
        if m and chave and servico:
            saida[(chave[0], chave[1], servico, m.group(1))] = "declarado"
    return saida


def ler_go(caminho):
    """Extrai os mapas UF->autorizador e a tabela de enderecos."""
    try:
        fonte = io.open(caminho, encoding="utf-8").read()
    except FileNotFoundError:
        sys.exit(f"tabela Go nao encontrada: {caminho}")
    fonte = re.sub(r"^\s*//.*$", "", fonte, flags=re.M)

    mapas = {}
    for m in re.finditer(
        r"var autorizador(NFe|NFCe) = map\[string\]string\{(.*?)\n\}", fonte, re.S
    ):
        mapas[m.group(1)] = dict(re.findall(r'"([\w-]+)":\s*"([\w-]+)"', m.group(2)))

    tabela = {}
    bloco = re.search(r"var endpointsPorAutorizador = .*?\n\}\n", fonte, re.S)
    if bloco:
        chave = servico = None
        for linha in bloco.group(0).splitlines():
            m = re.match(r'\s*"(\w+)/([\w-]+)": \{', linha)
            if m:
                chave = (m.group(1), m.group(2))
                continue
            m = re.match(r"\s*(Servico\w+): \{", linha)
            if m:
                servico = m.group(1)
                continue
            m = re.match(r'\s*(V\d{3}): urlPorAmbiente\{"([^"]*)", "([^"]*)"\}', linha)
            if m and chave and servico:
                tabela[(chave[0], chave[1], servico, m.group(1))] = (
                    m.group(2), m.group(3))

    return mapas, tabela


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)

    dados = carregar_ini(sys.argv[1])
    mapas, tabela = ler_go(sys.argv[2])

    if not mapas:
        sys.exit("nenhum mapa autorizadorNFe/autorizadorNFCe encontrado no Go")

    falhas = 0

    # 1) O mapa UF -> autorizador bate com o destino do `Usar=`?
    print("== mapa UF -> autorizador ==")
    conferidos = 0
    for modelo, mapa in mapas.items():
        for uf, declarado in mapa.items():
            sessao = f"{modelo}_{uf}_P"
            vistas = []
            while sessao in dados:
                usar = dados[sessao].get("Usar", "")
                if not usar or sessao in vistas:
                    break
                vistas.append(sessao)
                sessao = usar
            real = sessao
            for pref in ("NFCe_", "NFe_"):
                if real.startswith(pref):
                    real = real[len(pref):]
            real = real[:-2] if real.endswith(("_P", "_H")) else real

            conferidos += 1
            if real != declarado:
                falhas += 1
                print(f"  DIVERGE  {modelo}/{uf}: Go diz {declarado}, "
                      f"o .ini resolve para {real}")
    print(f"  {conferidos} UFs conferidas")

    # 2) Os enderecos por autorizador batem, UF a UF?
    print("\n== enderecos por autorizador ==")
    conferidos = 0
    for modelo, mapa in mapas.items():
        for uf, aut in mapa.items():
            for const, nome in SERVICOS.items():
                for rot, versao in VERSOES.items():
                    _, p = secao_que_define(dados, modelo, uf, "P", nome, versao)
                    _, h = secao_que_define(dados, modelo, uf, "H", nome, versao)
                    esperado = (p, h)
                    got = tabela.get((modelo, aut, const, rot), ("", ""))
                    if not any(esperado) and not any(got):
                        continue
                    conferidos += 1
                    if got != esperado:
                        falhas += 1
                        print(f"  DIVERGE  {modelo}/{uf} ({aut}) {const} {rot}")
                        print(f"             ini={esperado}")
                        print(f"             go ={got}")
    print(f"  {conferidos} combinacoes conferidas")

    # 3) A premissa do modelo: o `Usar=` e sempre de SECAO INTEIRA.
    print("\n== premissa: redirecionamento e de secao inteira ==")
    quebras = []
    for modelo in mapas:
        for uf in mapas[modelo]:
            for amb in ("P", "H"):
                secoes = set()
                for nome in SERVICOS.values():
                    for versao in VERSOES.values():
                        s, _ = secao_que_define(dados, modelo, uf, amb, nome, versao)
                        if s:
                            secoes.add(s)
                if len(secoes) > 1:
                    quebras.append(f"{modelo}/{uf}/{amb}: {sorted(secoes)}")
    if quebras:
        falhas += 1
        print("  QUEBROU: ha UF que define um servico localmente e delega outro")
        for q in quebras[:5]:
            print(f"    {q}")
        print("  o modelo por autorizador NAO serve mais; a tabela precisa"
              " voltar a ser por UF")
    else:
        print("  OK: nenhuma UF mistura definicao local com delegacao")

    # 4) Os overrides de SoapAction declarados no `.ini` estao no Go?
    print("\n== overrides de SoapAction (chave SoapAction= do .ini) ==")
    declarados = ler_actions_go(sys.argv[2])
    achados = 0
    for modelo, mapa in mapas.items():
        for uf, aut in mapa.items():
            for const, nome in SERVICOS.items():
                for rot, versao in VERSOES.items():
                    a = action_do_ini(dados, modelo, uf, "P", nome, versao)
                    if not a:
                        continue
                    achados += 1
                    if (modelo, aut, const, rot) not in declarados:
                        falhas += 1
                        print(f"  FALTA    {modelo}/{uf} ({aut}) {const} {rot}:"
                              f" o .ini declara {a}")
    if achados:
        print(f"  {achados} override(s) no .ini, todos declarados no Go"
              if not falhas else f"  {achados} override(s) no .ini")
    else:
        print("  nenhum override no .ini para os servicos conferidos")

    print()
    if falhas:
        print(f"{falhas} divergencia(s) -- confira cada uma contra o .ini")
        return 1
    print("OK: tabela de endpoints identica a resolucao do .ini")
    return 0


if __name__ == "__main__":
    sys.exit(main())
