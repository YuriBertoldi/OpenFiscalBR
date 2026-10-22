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
import os
import re
import sys

# Constante Go -> nome da chave no `.ini`. Acrescente aqui o servico novo
# quando a etapa que o transmite entrar.
SERVICOS = {
    "ServicoStatus": "NfeStatusServico",
    "ServicoConsultaProtocolo": "NfeConsultaProtocolo",
    "ServicoAutorizacao": "NfeAutorizacao",
    "ServicoRetAutorizacao": "NfeRetAutorizacao",
    "ServicoRecepcaoLote": "NfeRecepcao",
    "ServicoRetRecepcaoLote": "NfeRetRecepcao",
    "ServicoInutilizacao": "NfeInutilizacao",
    "ServicoConsultaCadastro": "NfeConsultaCadastro",
    "ServicoAdministrarCSC": "AdministrarCSCNFCe",
    "ServicoRecepcaoEvento": "RecepcaoEvento",
    "ServicoEventoEPEC": "EventoEPEC",
    "ServicoDistribuicaoDFe": "NFeDistribuicaoDFe",
}

# Campo Go -> versao de leiaute que ele atende.
VERSOES = {"V200": 2.00, "V310": 3.10, "V400": 4.00}

# Nome do servico no `.ini` -> prefixo do arquivo de schema, porte de
# LayOutToSchema (ACBrNFe.Conversao.pas:1193). E este mapa que diz onde
# procurar o `.xsd` que pode resgatar a versao pedida.
SCHEMA_DO_SERVICO = {
    "NfeStatusServico": "consStatServ",
    "NfeConsultaProtocolo": "consSitNFe",
    "NfeAutorizacao": "nfe",
    "NfeRetAutorizacao": "retEnviNFe",
    "NfeRecepcao": "nfe",
    "NfeRetRecepcao": "consReciNFe",
    "NfeInutilizacao": "inutNFe",
    "NfeConsultaCadastro": "consCad",
    "AdministrarCSCNFCe": "admCscNFCe",
    "RecepcaoEvento": "envEvento",
    "EventoEPEC": "eventoEPEC",
    "NFeDistribuicaoDFe": "distDFeInt",
}


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


def versao_da_chave(dados, sessao, servico, versao):
    """Versao da chave que o LerServicoChaveDeParams acharia, ou None."""
    kv = dados.get(sessao)
    if kv is None:
        return None

    base = (servico + "_").upper()
    achada = 0.0
    if any(k.upper() == base + "%.2f" % versao and kv[k] for k in kv):
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
                    achada = v
    return achada or None


def versao_resolvida(dados, modelo, uf, amb, servico, versao):
    """Segue o `Usar=` ate achar a chave, e devolve a versao dela."""
    sessao = f"{modelo}_{uf}_{amb}"
    vistas = []
    while True:
        v = versao_da_chave(dados, sessao, servico, versao)
        if v:
            return v
        if sessao in vistas:
            return None
        vistas.append(sessao)
        usar = dados.get(sessao, {}).get("Usar", "")
        if not usar:
            return None
        sessao = usar


def versoes_de_schema(pasta, prefixo):
    """Versoes para as quais existe `<prefixo>_vN.NN.xsd` na pasta."""
    if not pasta or not os.path.isdir(pasta):
        return None
    achadas = set()
    alvo = prefixo.lower() + "_v"
    for nome in os.listdir(pasta):
        n = nome.lower()
        if n.startswith(alvo) and n.endswith(".xsd"):
            try:
                achadas.add(float(n[len(alvo):-len(".xsd")]))
            except ValueError:
                continue
    return achadas


def ler_versoes_fixas_go(caminho):
    """Extrai o mapa versaoFixaPorServico do Go."""
    fonte = re.sub(r"^\s*//.*$", "", io.open(caminho, encoding="utf-8").read(), flags=re.M)
    bloco = re.search(r"var versaoFixaPorServico = .*?\n\}", fonte, re.S)
    if not bloco:
        return {}
    return dict(re.findall(r'(Servico\w+):\s*"([\d.]+)"', bloco.group(0)))


def main():
    if len(sys.argv) not in (3, 4):
        sys.exit(__doc__)

    pasta_schemas = sys.argv[3] if len(sys.argv) == 4 else ""
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

    # 3) A premissa do modelo: UF que mistura definicao local com delegacao.
    #
    # Misturar nao e, por si, um defeito: a entrada da tabela leva o nome do
    # autorizador, mas carrega os valores de UMA UF -- a primeira do grupo.
    # Enquanto o grupo tiver so aquela UF, os dois coincidem e a tabela esta
    # certa (e a verificacao 2, que confere UF a UF, prova isso).
    #
    # Vira defeito quando o grupo tem mais de uma UF: as demais herdariam um
    # endereco que o `.ini` nunca lhes daria. Por isso a severidade depende
    # do tamanho do grupo -- grupo de um e AVISO, grupo maior e falha.
    print("\n== UF que mistura definicao local com delegacao ==")
    grupos = collections.defaultdict(set)
    for modelo, mapa in mapas.items():
        for uf, aut in mapa.items():
            grupos[(modelo, aut)].add(uf)

    avisos, quebras = [], []
    for modelo, mapa in mapas.items():
        for uf, aut in mapa.items():
            for amb in ("P", "H"):
                secoes = set()
                for nome in SERVICOS.values():
                    for versao in VERSOES.values():
                        s, _ = secao_que_define(dados, modelo, uf, amb, nome, versao)
                        if s:
                            secoes.add(s)
                if len(secoes) <= 1:
                    continue
                companhia = sorted(grupos[(modelo, aut)] - {uf})
                texto = f"{modelo}/{uf}/{amb} ({aut}): {sorted(secoes)}"
                if companhia:
                    quebras.append(f"{texto}; divide o autorizador com {companhia}")
                else:
                    avisos.append(texto)

    for a in avisos:
        print(f"  AVISO    {a}")
    if avisos:
        print("  sozinha no autorizador: a entrada da tabela e so dela, segue"
              " correta. Se outra UF entrar nesse autorizador, vira falha.")
    if quebras:
        falhas += 1
        print("  QUEBROU: UF que mistura E divide o autorizador com outra")
        for q in quebras[:5]:
            print(f"    {q}")
        print("  as demais UFs do grupo herdariam endereco que o .ini nao"
              " lhes da; a tabela precisa de entrada propria para esta UF")
    elif not avisos:
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

    # 5) A versao da MENSAGEM: o `.ini` caiu de versao e o schema nao resgatou?
    print("\n== versao da mensagem (chave do .ini + schema) ==")
    fixas_go = ler_versoes_fixas_go(sys.argv[2])
    if not pasta_schemas:
        print("  pasta de schemas nao informada; so os casos de queda do .ini")

    proprias = {}
    for modelo, mapa in mapas.items():
        for uf in mapa:
            for const, nome in SERVICOS.items():
                for versao in sorted(VERSOES.values()):
                    achada = versao_resolvida(dados, modelo, uf, "P", nome, versao)
                    if achada is None or abs(achada - versao) < 1e-9:
                        continue
                    schemas = versoes_de_schema(pasta_schemas,
                                                SCHEMA_DO_SERVICO.get(nome, ""))
                    if schemas is None:
                        proprias.setdefault((const, achada), set()).add("?")
                        continue
                    # O AchaArquivoSchema pega o maior schema <= pedida.
                    candidatos = [v for v in schemas if v <= versao]
                    if candidatos and abs(max(candidatos) - versao) < 1e-9:
                        continue  # o schema resgatou a versao pedida
                    efetiva = max(candidatos) if candidatos else achada
                    proprias.setdefault((const, efetiva), set()).add(f"{modelo}/{uf}")

    if not proprias:
        print("  OK: toda mensagem sai na versao do leiaute configurado")
    for (const, efetiva), ufs in sorted(proprias.items(), key=lambda x: x[0]):
        rotulo = "%.2f" % efetiva
        declarada = fixas_go.get(const)
        exemplo = sorted(ufs)[0]
        if declarada == rotulo:
            print(f"  OK       {const} sai em {rotulo} e o Go declara isso")
        elif "?" in ufs:
            print(f"  AVISO    {const}: o .ini cai para {rotulo}; informe a pasta"
                  f" de schemas para decidir (ex.: {exemplo})")
        else:
            falhas += 1
            print(f"  FALTA    {const} sai em {rotulo}, mas versaoFixaPorServico"
                  f" diz {declarada or 'nada'} (ex.: {exemplo})")

    print()
    if falhas:
        print(f"{falhas} divergencia(s) -- confira cada uma contra o .ini")
        return 1
    print("OK: tabela de endpoints identica a resolucao do .ini")
    return 0


if __name__ == "__main__":
    sys.exit(main())
