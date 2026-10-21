#!/usr/bin/env python3
"""Confere a tabela de URLs portada contra o <Componente>Servicos.ini do ACBr.

A tabela de URLs e DADO, nao logica: 108 entradas transcritas a mao de um
.ini de 1388 linhas. Nenhum teste Go consegue dizer se uma delas foi digitada
errada -- um teste so pode repetir o valor que esta no Go, o que nao prova
nada. A unica fonte de verdade e o .ini.

Este script reimplementa a resolucao do ACBr sobre o .ini e compara com o que
esta no arquivo Go, entrada a entrada.

A resolucao tem tres etapas (ACBrDFe.pas:456-607):

1. chave versionada exata (`URL-QRCode_2.00`);
2. senao, a maior versao <= a pedida entre as chaves com o mesmo prefixo;
3. senao, a chave SEM versao (`URL-QRCode`).

E, se ainda assim nao achou nada, a secao pode redirecionar para outra pela
chave `Usar=`, com protecao contra ciclo.

Uso:

    py ferramentas/conferir-urls-dfe.py <servicos.ini> <urls_tabela.go>

    py ferramentas/conferir-urls-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/ACBrNFeServicos.ini" \\
       packages/nfe/urls_tabela.go

Saida: uma linha por divergencia e um resumo. Codigo de saida 1 se houver
divergencia, para encaixar em CI.
"""

import collections
import io
import re
import sys

# UFs na ordem em que a tabela Go as declara -- a mesma do IBGE por codigo.
UFS = [
    "AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA", "MT", "MS",
    "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN", "RS", "RO", "RR", "SC",
    "SP", "SE", "TO",
]

# Mapa dos nomes Go para (modelo do .ini, servico, versao do QR-Code pedida).
#
# A versao importa so no URL-ConsultaNFCe, unico servico do .ini cuja URL
# muda com ela. Nos demais, 3.0 resolve para o mesmo valor que 1.0 e 2.0 --
# o proprio script confere isso e avisa se deixar de ser verdade.
TABELAS = {
    "urlsQRCode": ("NFCe", "URL-QRCode", 3.0),
    "urlsQRCodeModelo55": ("NFe", "URL-QRCode", 3.0),
    "urlsConsultaNFe": ("NFe", "URL-ConsultaNFe", 3.0),
    "urlsConsultaNFCe": ("NFCe", "URL-ConsultaNFCe", 2.0),
    "urlsConsultaNFCeQRCodeV1": ("NFCe", "URL-ConsultaNFCe", 1.0),
}

# Tabelas que declaram SO as divergencias em relacao a outra -- a ausencia de
# uma UF nelas e intencional, nao falta de transcricao.
PARCIAIS = {"urlsQRCodeModelo55": "urlsQRCode"}


def carregar_ini(caminho):
    """Le o .ini preservando a ordem e sem interpretar nada."""
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

    # Busca exata, pela mesma versao.
    for k in kv:
        if k.upper() == chave and kv[k]:
            achada = versao

    # Senao, a maior versao que nao ultrapasse a pedida.
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

    # Nenhuma versionada serviu: cai na chave sem versao.
    for k in kv:
        if k.upper() == servico.upper():
            return kv[k]
    return ""


def ler_url(dados, modelo, uf, amb, servico, versao):
    """Porte de LerServicoDeParams + o laco de redirecionamento `Usar=`."""
    sessao = f"{modelo}_{uf}_{amb}"
    vistas = []
    url = chave_de_params(dados, sessao, servico, versao)
    while not url:
        if sessao:
            if sessao in vistas:
                sys.exit(f"redirecionamento circular na secao {sessao}")
            vistas.append(sessao)
        usar = dados.get(sessao, {}).get("Usar", "")
        if not usar:
            break
        sessao = usar
        url = chave_de_params(dados, sessao, servico, versao)
    return url


def ler_go(caminho):
    """Extrai {tabela: {UF: (producao, homologacao)}} do arquivo Go."""
    try:
        fonte = io.open(caminho, encoding="utf-8").read()
    except FileNotFoundError:
        sys.exit(f"tabela Go nao encontrada: {caminho}")

    fonte = re.sub(r"^\s*//.*$", "", fonte, flags=re.M)
    tabelas = {}
    for m in re.finditer(
        r"var\s+(\w+)\s*=\s*map\[string\]urlPorAmbiente\{(.*?)\n\}", fonte, re.S
    ):
        entradas = {}
        for e in re.finditer(r'"([A-Z]{2})":\s*\{"([^"]*)",\s*"([^"]*)"\}', m.group(2)):
            entradas[e.group(1)] = (e.group(2), e.group(3))
        tabelas[m.group(1)] = entradas
    return tabelas


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)

    dados = carregar_ini(sys.argv[1])
    go = ler_go(sys.argv[2])

    falhas = 0
    conferidas = 0

    print("== URLs por UF e ambiente ==")
    for nome, (modelo, servico, versao) in TABELAS.items():
        if nome not in go:
            print(f"  AUSENTE  tabela {nome} nao existe no Go")
            falhas += 1
            continue

        parcial_de = PARCIAIS.get(nome)
        for uf in UFS:
            esperado = (
                ler_url(dados, modelo, uf, "P", servico, versao),
                ler_url(dados, modelo, uf, "H", servico, versao),
            )

            if uf not in go[nome]:
                # Numa tabela de divergencias, a ausencia so e correta se o
                # valor for mesmo igual ao da tabela de referencia.
                if parcial_de:
                    base = go.get(parcial_de, {}).get(uf)
                    if base is None:
                        continue
                    if tuple(base) != esperado:
                        falhas += 1
                        print(
                            f"  FALTA    {nome}[{uf}] deveria divergir de"
                            f" {parcial_de}: ini={esperado} base={tuple(base)}"
                        )
                    else:
                        conferidas += 1
                    continue
                falhas += 1
                print(f"  FALTA    {nome}[{uf}] ausente no Go; ini={esperado}")
                continue

            conferidas += 1
            if tuple(go[nome][uf]) != esperado:
                falhas += 1
                print(f"  DIVERGE  {nome}[{uf}]")
                print(f"             ini={esperado}")
                print(f"             go ={tuple(go[nome][uf])}")

        # Numa tabela de divergencias, UF declarada que NAO diverge e ruido.
        if parcial_de:
            for uf in go[nome]:
                base = go.get(parcial_de, {}).get(uf)
                if base is not None and tuple(base) == tuple(go[nome][uf]):
                    falhas += 1
                    print(
                        f"  REDUNDA  {nome}[{uf}] repete {parcial_de}[{uf}];"
                        " a tabela so deve declarar divergencia"
                    )

    print(f"  {conferidas} entradas conferidas")

    print("\n== invariantes assumidas pelo porte ==")
    invariantes = [
        ("URL-QRCode nao varia com a versao do QR-Code",
         [("NFe", "URL-QRCode"), ("NFCe", "URL-QRCode")], (1.0, 2.0, 3.0)),
        ("URL-ConsultaNFe nao varia com a versao do QR-Code",
         [("NFe", "URL-ConsultaNFe")], (1.0, 2.0, 3.0)),
        ("URL-ConsultaNFCe e igual nas versoes 2 e 3",
         [("NFCe", "URL-ConsultaNFCe")], (2.0, 3.0)),
    ]
    for descricao, alvos, versoes in invariantes:
        quebras = []
        for modelo, servico in alvos:
            for uf in UFS:
                for amb in ("P", "H"):
                    vals = {ler_url(dados, modelo, uf, amb, servico, v) for v in versoes}
                    if len(vals) > 1:
                        quebras.append(f"{modelo}/{uf}/{amb}: {sorted(vals)}")
        if quebras:
            falhas += 1
            print(f"  QUEBROU  {descricao}")
            for q in quebras[:5]:
                print(f"             {q}")
        else:
            print(f"  OK       {descricao}")

    print()
    if falhas:
        print(f"{falhas} divergencia(s) -- confira cada uma contra o .ini")
        return 1
    print("OK: tabela de URLs identica a resolucao do .ini")
    return 0


if __name__ == "__main__":
    sys.exit(main())
