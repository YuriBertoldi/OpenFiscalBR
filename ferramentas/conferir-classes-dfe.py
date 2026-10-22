# -*- coding: utf-8 -*-
"""Confere as structs de um package DFe contra as classes do ACBr Delphi.

    py ferramentas/conferir-classes-dfe.py <Classes.pas> <arquivo.go>... \\
       [--par TClasseDelphi=StructGo] [--ignorar StructGo]

O casamento e por NOME: a classe `TInfCarga` do Delphi casa com a struct
`InfCarga` do Go. Sufixos `Collection` e `CollectionItem` sao descartados,
porque no Go uma colecao vira slice e nao tem tipo proprio.

O casamento de CAMPO usa o comentario de tag. Toda struct portada aqui
escreve a tag do leiaute ao lado do campo:

    CMunIni  int     // cMunIni
    Modelo   int     // mod
    CNPJCPF  string  // CNPJ ou CPF

Quando o comentario e um identificador unico, ele e a tag e vale mais que o
nome do campo -- e assim que `Modelo` casa com a property `mod` do Delphi.
Quando e prosa ("CNPJ ou CPF", "seletor"), cai de volta no nome do campo.
Texto apos `--` e descartado.

Comparacao SEM CAIXA, porque o Delphi nao distingue e o leiaute mistura
`UFIni` com `cMunIni` sem criterio.

O TIPO tambem e conferido, em tres formas:

    TIPO DIVERGE      escalar contra escalar (Currency x string)
    ARIDADE DIVERGE   colecao do Delphi contra campo unico no Go
    ALVO DIVERGE      duas structs portadas apontando para alvos diferentes

O terceiro e o que importa: `Det.InfNFe []InfNF` no lugar de `[]InfNFe` casa
pelo nome do campo e passaria limpo. E a "hierarquia trocada".

LIMITES -- o que este script NAO pega:

  - ORDEM dos campos. Para struct isso nao importa; para writer importa, e
    quem confere ordem e o conferir-gerador-dfe.py.
  - NOME de enum. TpcteTipoCTe e TipoCTe nao casam por nome; para enum, so
    a aridade e conferida.
  - ALIAS. `TObsFiscoCollection = TObsContCollection` nao e classe e nao
    aparece em lista nenhuma. Alias so se acha lendo o .pas.
  - grupo inteiro esquecido. Se a struct inteira nao existe no Go, aparece
    como classe sem par -- mas so se o nome bater pela regra acima.

Como toda ferramenta daqui, ela propria precisa ser auditada: um achado do
script e hipotese, nao veredito. Confira no .pas antes de mudar codigo.
"""
import io
import os
import re
import sys

SUFIXOS = ("CollectionItem", "Collection")


def le(path):
    for enc in ("utf-8", "cp1252", "latin-1"):
        try:
            return io.open(path, encoding=enc).read()
        except (UnicodeDecodeError, UnicodeError):
            pass
    return io.open(path, encoding="latin-1", errors="replace").read()


def tira_comentarios_pascal(src):
    src = re.sub(r"//[^\n]*", "", src)
    src = re.sub(r"\{.*?\}", "", src, flags=re.S)
    src = re.sub(r"\(\*.*?\*\)", "", src, flags=re.S)
    return src


def classes_delphi(path, pares, pares_tipo):
    """Devolve {chave: (Nome, [propriedades])} de um Classes.pas.

    `pares` casa pelo nome ORIGINAL da classe, antes de normalizar. Isso
    existe porque a normalizacao pode COLIDIR: `TInfCteComp` e
    `TInfCteCompCollectionItem` viram ambas `InfCteComp`, e uma apagava a
    outra. Declarar --par TInfCteCompCollectionItem=InfCteCompItem separa as
    duas.
    """
    src = tira_comentarios_pascal(le(path))
    # A indentacao da declaracao VARIA: o TCTe tem UM espaco a frente, nao
    # dois, e foi o que fez a raiz passar despercebida na primeira varredura.
    decls = list(re.finditer(r"^ *(T\w+) = class\b", src, re.M))
    out, colisoes, brutos = {}, [], {}
    for i, m in enumerate(decls):
        fim = decls[i + 1].start() if i + 1 < len(decls) else len(src)
        corpo = src[m.start():fim]
        props = []
        vistos = set()
        for p in re.finditer(
            r"^\s*property\s+(\w+)\s*:\s*([\w.]+)", corpo, re.M
        ):
            nome = p.group(1)
            if nome.lower() not in vistos:
                vistos.add(nome.lower())
                props.append((nome, tipo_delphi(p.group(2), pares_tipo)))
        bruto = m.group(1)
        if bruto.lower() in pares:
            out[pares[bruto.lower()].lower()] = (bruto, props)
            continue
        chave = normaliza(bruto).lower()
        # TXxxCollectionItem e TXxxCollection normalizam para o MESMO nome, e
        # a Collection vem depois no arquivo. Deixar a segunda sobrescrever
        # zerava a lista de campos do item inteiro -- a Collection so tem
        # `Items[Index]`, que nao e property nomeada -- e o relatorio acusava
        # a struct Go inteira como sobra. Vence quem tem mais properties.
        if chave in out:
            # Colisao. Resolver em SILENCIO era um falso negativo da classe
            # exata que esta ferramenta existe para pegar: sem o --par,
            # TInfCteCompCollectionItem (1 property, empate com TInfCteComp)
            # sumia das DUAS listas -- nao aparecia como classe sem struct,
            # nao gerava divergencia, nada. Grupo inteiro nao portado passava
            # batido.
            #
            # Mas a colisao entre TXxxCollectionItem e TXxxCollection e o
            # caso NORMAL -- acontece em toda colecao, 21 vezes so no CT-e --
            # e avisar dela afogaria o aviso util. Suspeita e qualquer OUTRA
            # combinacao: classe simples contra item, por exemplo.
            antigo, antigas = brutos[chave], out[chave][1]
            if {forma(antigo), forma(bruto)} != {"item", "collection"}:
                vence = bruto if len(props) > len(antigas) else antigo
                colisoes.append((chave, antigo, bruto, vence))
        if chave not in out or len(props) > len(out[chave][1]):
            out[chave] = (normaliza(bruto), props)
            brutos[chave] = bruto
    return out, colisoes


# Tipo normalizado: ("num"|"int"|"str"|"time"|"bool"|"nome", eh_slice).
#
# Comparar tipo era o furo declarado da ferramenta -- e o pior caso nao era
# escalar trocado, era ALVO trocado: `Det.InfNFe []InfNF` no lugar de
# `[]InfNFe` casa pelo nome do campo e nao gerava divergencia nenhuma. Isso
# e "hierarquia trocada", que o CLAUDE.md manda conferir.
ESCALAR_DELPHI = {
    "currency": "num", "double": "num", "real": "num", "extended": "num",
    "single": "num",
    "integer": "int", "int64": "int", "cardinal": "int", "word": "int",
    "byte": "int", "longint": "int", "smallint": "int",
    "string": "str", "ansistring": "str", "widestring": "str",
    "tdatetime": "time",
    "boolean": "bool",
}

ESCALAR_GO = {
    "float64": "num", "float32": "num",
    "int": "int", "int64": "int", "int32": "int",
    "string": "str",
    "time.time": "time",
    "bool": "bool",
}


def tipo_delphi(bruto, pares):
    baixo = bruto.lower()
    if baixo in ESCALAR_DELPHI:
        return (ESCALAR_DELPHI[baixo], False)
    eh_slice = baixo.endswith("collection")
    # O --par vale aqui tambem: sem isso o tipo TInfCteCompCollection
    # normaliza para `infctecomp` e acusa ALVO DIVERGE contra o
    # InfCteCompItem que o proprio --par pediu.
    if baixo in pares:
        return (pares[baixo].lower(), eh_slice)
    return (normaliza(bruto).lower(), eh_slice)


def tipo_go(bruto):
    eh_slice = bruto.startswith("[]")
    nu = bruto[2:] if eh_slice else bruto
    baixo = nu.lower()
    if baixo in ESCALAR_GO:
        return (ESCALAR_GO[baixo], eh_slice)
    # `pcn.Modal` e `rtc.IBSCBS` respondem pelo nome sem o package.
    return (baixo.split(".")[-1], eh_slice)


def forma(nome):
    """item, collection ou simples."""
    if nome.endswith("CollectionItem"):
        return "item"
    if nome.endswith("Collection"):
        return "collection"
    return "simples"


def normaliza(nome):
    """TInfCargaCollectionItem -> InfCarga."""
    if nome.startswith("T"):
        nome = nome[1:]
    for s in SUFIXOS:
        if nome.endswith(s) and len(nome) > len(s):
            nome = nome[: -len(s)]
    return nome


def tag_do_comentario(comentario, campo):
    """A tag do leiaute, ou o nome do campo quando o comentario e prosa."""
    if not comentario:
        return campo
    texto = comentario.split("--")[0].strip()
    # "xCampo (atributo)" continua dizendo a tag; o parentese e nota.
    texto = re.sub(r"\(.*?\)", "", texto).strip()
    # Um identificador unico e tag; qualquer outra coisa e explicacao.
    if re.match(r"^[A-Za-z_]\w*$", texto):
        return texto
    return campo


def structs_go(paths):
    """Devolve {Nome: [tags]} das structs dos .go informados."""
    out = {}
    for path in paths:
        src = le(path)
        # O `\{\}?` apanha tambem `type Rodo struct{}` numa linha so. Sem
        # ele, uma struct declarada vazia some do relatorio em vez de
        # aparecer como classe por portar -- que e justamente o que ela e.
        for m in re.finditer(
            r"^type (\w+) struct ?\{(?:\}|(.*?)^\})", src, re.M | re.S
        ):
            campos = []
            for linha in (m.group(2) or "").split("\n"):
                linha = linha.strip()
                if not linha or linha.startswith("//"):
                    continue
                # A struct tag e opcional. Sem admiti-la, o dia em que a
                # etapa do gerador acrescentar `xml:"nFat"` a ferramenta
                # deixa de enxergar o campo -- e no sentido "sobra" isso e
                # falso NEGATIVO: campo Go que nao existe no Delphi para de
                # ser reportado.
                c = re.match(
                    r"^([A-Z]\w*)\s+(\S+)(?:\s+`[^`]*`)?\s*(?://\s*(.*))?$",
                    linha,
                )
                if c:
                    # O par e (nome do campo, tag). Casa se QUALQUER um dos
                    # dois bater com a property: a tag do leiaute e o nome da
                    # property as vezes divergem -- `mod` no XML, `modelo` na
                    # classe -- e exigir que a tag casasse acusava Ide, InfNF
                    # e RefNF sem haver campo faltando em nenhuma.
                    campos.append((
                        c.group(1),
                        tag_do_comentario(c.group(3), c.group(1)),
                        tipo_go(c.group(2)),
                    ))
            out[m.group(1)] = campos
    return out


def main():
    args = [a for a in sys.argv[1:]]
    pares, pares_tipo, ignorar = {}, {}, set()
    resto = []
    i = 0
    while i < len(args):
        if args[i] == "--par":
            d, _, g = args[i + 1].partition("=")
            pares[d.lower()] = g
            pares_tipo[d.lower()] = g
            # Um --par sobre o CollectionItem vale tambem para a Collection
            # do mesmo nome base QUANDO o alvo e um TIPO de campo: a property
            # diz `infCteComp10: TInfCteCompCollection`, e sem isso a
            # comparacao acusava um ALVO DIVERGE que nao existe.
            #
            # So no mapa de TIPO. No de CLASSE, deixar a Collection reclamar
            # a mesma chave a fazia sobrescrever o CollectionItem -- ela nao
            # tem property nomeada nenhuma, e a struct Go inteira voltava a
            # aparecer como sobra.
            if d.endswith("CollectionItem"):
                pares_tipo[d[: -len("Item")].lower()] = g
            i += 2
        elif args[i] == "--ignorar":
            ignorar.add(args[i + 1])
            i += 2
        else:
            resto.append(args[i])
            i += 1

    if len(resto) < 2:
        sys.exit(__doc__)

    pas, gos = resto[0], resto[1:]
    if not os.path.exists(pas):
        sys.exit("nao encontrei %s" % pas)

    delphi, colisoes = classes_delphi(pas, pares, pares_tipo)
    go = structs_go(gos)

    # O casamento de NOME tambem e sem caixa: o ACBr escreve `TdetGTV` e
    # `TautXMLCollectionItem` com inicial minuscula, e o Go exporta com
    # maiuscula obrigatoriamente. Comparar com caixa reportava 25 structs
    # como sobra e 35 classes como falta, todas falsas.
    go_por_chave = {n.lower(): n for n in go}
    ignorar_chave = {n.lower() for n in ignorar}

    casados, sem_par_delphi, sem_par_go = [], [], []
    usados = set()
    for alvo in sorted(delphi):
        nome, props = delphi[alvo]
        if alvo in ignorar_chave:
            continue
        if alvo in go_por_chave:
            g = go_por_chave[alvo]
            casados.append((nome, g, props, go[g]))
            usados.add(alvo)
        elif props:  # classe sem property nenhuma nao e grupo de dados
            sem_par_delphi.append(nome)
    for chave in sorted(go_por_chave):
        if chave not in usados and chave not in ignorar_chave:
            sem_par_go.append(go_por_chave[chave])

    # Nomes de struct conhecidos: so para eles o ALVO de um campo pode ser
    # comparado. Enum do Delphi e do Go nao casam por nome (TpcteTipoCTe x
    # TipoCTe), entao nesses so a aridade -- slice ou nao -- e conferivel.
    conhecidos = set(go_por_chave)

    divergencias = 0
    print("== Campos divergentes ==")
    repetidas = []
    for nome, g, props, campos in casados:
        pset = {p.lower() for p, _ in props}
        # Um campo Go responde pelo seu nome E pela sua tag.
        gset = set()
        por_chave = {}
        for campo, tag, tipo in campos:
            gset.add(campo.lower())
            gset.add(tag.lower())
            por_chave[campo.lower()] = tipo
            por_chave.setdefault(tag.lower(), tipo)
        faltam = [p for p, _ in props if p.lower() not in gset]
        sobram = [
            c for c, t, _ in campos
            if c.lower() not in pset and t.lower() not in pset
        ]

        for prop, (dtipo, dslice) in props:
            gtipo = por_chave.get(prop.lower())
            if gtipo is None:
                continue
            gnome, gslice = gtipo
            if dslice != gslice:
                divergencias += 1
                print("\n%s.%s -> ARIDADE DIVERGE: Delphi %s, Go %s" % (
                    g, prop,
                    "colecao" if dslice else "unico",
                    "slice" if gslice else "unico"))
            elif dtipo in ("num", "int", "str", "time", "bool") or \
                    gnome in ("num", "int", "str", "time", "bool"):
                if dtipo != gnome:
                    divergencias += 1
                    print("\n%s.%s -> TIPO DIVERGE: Delphi %s, Go %s"
                          % (g, prop, dtipo, gnome))
            elif dtipo in conhecidos and gnome != dtipo:
                # Os dois sao struct portada e apontam para alvos
                # diferentes -- e a "hierarquia trocada".
                divergencias += 1
                print("\n%s.%s -> ALVO DIVERGE: Delphi %s, Go %s"
                      % (g, prop, dtipo, gnome))

        if faltam or sobram:
            divergencias += 1
            print("\n%s -> %s" % (nome, g))
            if faltam:
                print("   SO NO DELPHI: %s" % ", ".join(faltam))
            if sobram:
                print("   SO NO GO    : %s" % ", ".join(sobram))

        tags = [t for _, t, _ in campos]
        dup = sorted({t for t in tags if tags.count(t) > 1})
        if dup:
            repetidas.append((g, dup))

    if not divergencias:
        print("(nenhuma)")

    # Tag repetida nao e erro por si: o ACBr tem grupos em que duas
    # propriedades saem com a MESMA tag, uma por versao de leiaute. Mas e
    # sempre algo a conferir, entao sai como aviso separado.
    print("\n== Avisos: dois campos com a mesma tag ==")
    if repetidas:
        for g, dup in repetidas:
            print("   %s: %s" % (g, ", ".join(dup)))
    else:
        print("(nenhum)")

    print("\n== Avisos: nomes de classe que colidiram ==")
    if colisoes:
        for chave, antigo, novo, vence in colisoes:
            print("   %s e %s -> ambos viram %s; venceu %s"
                  % (antigo, novo, chave, vence))
        print("   (silencie com --par <ClasseDelphi>=<StructGo>)")
    else:
        print("(nenhum)")

    print("\n== Classes do Delphi sem struct no Go ==")
    print(", ".join(sem_par_delphi) if sem_par_delphi else "(nenhuma)")

    print("\n== Structs do Go sem classe no Delphi ==")
    print(", ".join(sem_par_go) if sem_par_go else "(nenhuma)")

    print(
        "\n%d classes casadas, %d com divergencia de campo."
        % (len(casados), divergencias)
    )


if __name__ == "__main__":
    main()
