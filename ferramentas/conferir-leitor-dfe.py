#!/usr/bin/env python3
"""Confere um leitor de DFe portado contra o XmlReader Delphi de origem.

Compara, tag a tag, as duas decisoes que o build, o `go vet` e o teste de
leitura NAO alcancam:

1. a PRECISAO de cada campo (`tcDe2` x `ConteudoDe2`, `tcDe10` x
   `ConteudoDe10`, ...);
2. a escolha entre `Find` (nome qualificado) e `FindAnyNs` (local name).

Por que um script e nao um teste Go: `pcn.ConteudoDec` so usa o numero de
casas quando o node esta marcado com `FloatIsIntString`, flag que o parser
nunca liga -- no XML da SEFAZ o valor ja vem com ponto decimal. Logo
`ConteudoDe2` e `ConteudoDe10` devolvem o MESMO resultado, e nenhum teste de
leitura consegue distinguir os dois. A precisao e um contrato declarativo
(herdado do mesmo mecanismo do ACBr) e so se verifica comparando com o `.pas`.

Uso:

    py ferramentas/conferir-leitor-dfe.py <arquivo.pas> <arquivo.go>

    py ferramentas/conferir-leitor-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.XmlReader.pas" \\
       packages/nfe/xml_reader.go

Saida: uma linha por divergencia e um resumo. Codigo de saida 1 se houver
divergencia, para encaixar em CI.
"""

import collections
import re
import sys

# tcXxx do ACBr -> sufixo de pcn.ConteudoXxx. Fonte da verdade do lado
# Delphi: ObterConteudoTag em ACBrDFe/ACBrXmlBase.pas; do lado Go:
# packages/pcn/tipos_campo.go.
EQUIVALENTES = {
    "tcStr": {"Str"}, "tcEsp": {"Str"}, "tcNumStr": {"Str"}, "tcStrOrig": {"Str"},
    "tcInt": {"Int"}, "tcInt64": {"Int64"},
    "tcDe1": {"De1"}, "tcDe2": {"De2"}, "tcDe3": {"De3"}, "tcDe4": {"De4"},
    "tcDe5": {"De5"}, "tcDe6": {"De6"}, "tcDe7": {"De7"}, "tcDe8": {"De8"},
    "tcDe10": {"De10"},
    # tcDat e tcDatHor caem os dois em ConteudoDataDef: o parser do pcn
    # aceita data e data-hora no mesmo caminho.
    "tcDat": {"Data", "DataDef", "DataHora"},
    "tcDatHor": {"Data", "DataDef", "DataHora"},
    "tcHor": {"Hora"}, "tcBool": {"Bool"}, "tcBoolStr": {"Bool"},
}


def sem_comentarios(fonte):
    """Remove linhas comentadas do Delphi, para nao contar campo desativado."""
    return re.sub(r"^\s*//.*$", "", fonte, flags=re.M)


def ler_pas(caminho):
    """Extrai {tag: {tcXxx}} e {tag: {Estrito|AnyNs}} do XmlReader Delphi."""
    try:
        fonte = open(caminho, encoding="cp1252").read()
    except FileNotFoundError:
        sys.exit(f"fonte Delphi nao encontrado: {caminho}")
    fonte = sem_comentarios(fonte)

    precisao = collections.defaultdict(set)
    for m in re.finditer(
        r"ObterConteudo\(\s*\w+(?:\.\w+)*\.Find(?:AnyNs)?\('([^']+)'\)\s*,\s*(tc\w+)\s*\)",
        fonte,
    ):
        precisao[m.group(1)].add(m.group(2))

    busca = collections.defaultdict(set)
    for m in re.finditer(
        r"\.Childrens\.(Find|FindAnyNs|FindAll|FindAllAnyNs)\('([^']+)'\)", fonte
    ):
        busca[m.group(2)].add("AnyNs" if "AnyNs" in m.group(1) else "Estrito")

    return precisao, busca


def ler_go(caminho):
    """Extrai {tag: {Xxx}} e {tag: {Estrito|AnyNs}} do leitor Go."""
    try:
        fonte = open(caminho, encoding="utf-8").read()
    except FileNotFoundError:
        sys.exit(f"leitor Go nao encontrado: {caminho}")

    precisao = collections.defaultdict(set)
    for m in re.finditer(
        r'pcn\.Conteudo(\w+)\(\s*\w+(?:\.\w+)*\.Find(?:AnyNs)?\("([^"]+)"\)\s*\)', fonte
    ):
        precisao[m.group(2)].add(m.group(1))

    busca = collections.defaultdict(set)
    for m in re.finditer(r'\.(Find|FindAnyNs|FindAll|FindAllAnyNs)\("([^"]+)"\)', fonte):
        busca[m.group(2)].add("AnyNs" if "AnyNs" in m.group(1) else "Estrito")

    return precisao, busca


def main():
    if len(sys.argv) != 3:
        sys.exit(__doc__)

    pas_precisao, pas_busca = ler_pas(sys.argv[1])
    go_precisao, go_busca = ler_go(sys.argv[2])

    falhas = 0

    print("== precisao por campo (tcDeN) ==")
    conferidas = 0
    sem_par = []
    for tag in sorted(pas_precisao):
        esperados = set()
        for tc in pas_precisao[tag]:
            esperados |= EQUIVALENTES.get(tc, {"?"})
        if tag not in go_precisao:
            sem_par.append((tag, sorted(pas_precisao[tag])))
            continue
        conferidas += 1
        if not (go_precisao[tag] & esperados):
            falhas += 1
            print(
                f"  DIVERGE {tag:24} delphi={sorted(pas_precisao[tag])}"
                f" -> esperado {sorted(esperados)}  go={sorted(go_precisao[tag])}"
            )
    print(f"  {conferidas} tags conferidas")
    if sem_par:
        print(
            f"  {len(sem_par)} tag(s) do .pas sem ConteudoXxx(Find()) direto no Go"
            " -- confira uma a uma se foram omitidas de proposito:"
        )
        for tag, tcs in sem_par:
            print(f"    {tag} ({','.join(tcs)})")

    print("\n== Find x FindAnyNs ==")
    comuns = sorted(set(pas_busca) & set(go_busca))
    for tag in comuns:
        if pas_busca[tag] != go_busca[tag]:
            falhas += 1
            print(
                f"  DIVERGE {tag:24} delphi={sorted(pas_busca[tag])}"
                f"  go={sorted(go_busca[tag])}"
            )
    print(f"  {len(comuns)} tags conferidas")

    so_pas = sorted(set(pas_busca) - set(go_busca))
    so_go = sorted(set(go_busca) - set(pas_busca))
    if so_pas:
        print(f"\n  so no .pas ({len(so_pas)}): {so_pas}")
    if so_go:
        print(f"\n  so no .go ({len(so_go)}): {so_go}")

    print()
    if falhas:
        print(f"{falhas} divergencia(s) -- confira cada uma contra o .pas")
        return 1
    print("OK: nenhuma divergencia de precisao nem de Find/FindAnyNs")
    return 0


if __name__ == "__main__":
    sys.exit(main())
