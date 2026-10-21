#!/usr/bin/env python3
"""Confere as regras de negocio portadas contra o ValidarRegrasdeNegocio do ACBr.

As regras de negocio sao 155 rejeicoes avulsas, cada uma com um codigo
oficial da SEFAZ. O risco nao e errar uma delas -- e ESQUECER uma: regra que
some nao quebra build, nao quebra teste e so aparece quando a SEFAZ recusa o
documento em producao.

Este script extrai os codigos dos dois lados e compara os conjuntos.

Do lado Delphi, separa as regras ATIVAS das COMENTADAS. O ACBr comenta regra
que a SEFAZ revogou ou que ainda nao entrou em vigor, e mante-las comentadas
e informacao: o porte nao deve implementa-las, mas quem le precisa saber que
elas existem.

Uso:

    py ferramentas/conferir-regras-dfe.py <ValidarRegrasdeNegocio.pas> <arquivo.go>...

    py ferramentas/conferir-regras-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.ValidarRegrasdeNegocio.pas" \\
       packages/nfe/regras_negocio*.go

Saida: os dois conjuntos e a diferenca. Codigo de saida 1 se faltar regra.
"""

import io
import re
import sys


def sem_comentarios(fonte):
    """Remove comentarios Delphi de bloco e de linha."""
    fonte = re.sub(r"\(\*.*?\*\)", "", fonte, flags=re.S)
    # Chaves aninhadas nao ocorrem no arquivo; o padrao simples basta.
    fonte = re.sub(r"\{[^{}]*\}", "", fonte, flags=re.S)
    return re.sub(r"^\s*//.*$", "", fonte, flags=re.M)


def ler_pas(caminho):
    """Devolve (ativas, comentadas) com os codigos de rejeicao do Delphi."""
    try:
        fonte = io.open(caminho, encoding="cp1252").read()
    except FileNotFoundError:
        sys.exit(f"fonte Delphi nao encontrado: {caminho}")

    padrao = r"AdicionaErro\(\s*'(\d+)-"
    todas = {int(m) for m in re.findall(padrao, fonte)}
    ativas = {int(m) for m in re.findall(padrao, sem_comentarios(fonte))}
    return sorted(ativas), sorted(todas - ativas)


def ler_go(caminhos):
    """Devolve os codigos que o porte Go e capaz de emitir.

    Reconhece um codigo por POSICAO SINTATICA, nao por contexto: um literal
    de tres digitos cercado por abre-parentese, abre-chave ou virgula de um
    lado e virgula ou fecha-parentese do outro.

    Isso cobre as quatro formas que o porte usa -- c.seErro(cond, 253, msg),
    c.seErroItem(cond, 777, item, msg), a tabela de confrontos {531, ...} e
    os codigos passados como parametro para uma regra generica -- sem casar
    com numeros de conta, como o 100 de uma divisao por cem.
    """
    codigos = set()
    for caminho in caminhos:
        try:
            fonte = io.open(caminho, encoding="utf-8").read()
        except FileNotFoundError:
            sys.exit(f"porte Go nao encontrado: {caminho}")
        fonte = re.sub(r"^\s*//.*$", "", fonte, flags=re.M)
        codigos |= {int(m) for m in re.findall(r"(?<=[,{(])\s*(\d{3})\s*(?=[,)])", fonte)}
    return sorted(codigos)


def main():
    if len(sys.argv) < 3:
        sys.exit(__doc__)

    ativas, comentadas = ler_pas(sys.argv[1])
    portadas = ler_go(sys.argv[2:])

    print(f"regras ATIVAS no .pas : {len(ativas)}")
    print(f"regras COMENTADAS     : {len(comentadas)}")
    if comentadas:
        print(f"  {comentadas}")
        print("  (o porte NAO deve implementa-las; confira se o motivo esta"
              " registrado num comentario do Go)")
    print(f"regras no porte Go    : {len(portadas)}")

    faltando = [r for r in ativas if r not in portadas]
    sobrando = [r for r in portadas if r not in ativas]

    if faltando:
        print(f"\nFALTANDO no Go ({len(faltando)}): {faltando}")
    if sobrando:
        # Pode ser codigo de rejeicao de outra origem ou falso positivo da
        # extracao -- nao reprova, mas precisa ser lido.
        print(f"\nSO no Go ({len(sobrando)}): {sobrando}")
        print("  (confira um a um: ou e falso positivo da extracao, ou e"
              " regra que o porte inventou)")

    print()
    if faltando:
        print(f"{len(faltando)} regra(s) do .pas sem contraparte no porte")
        return 1
    print("OK: todas as regras ativas do .pas tem contraparte no porte")
    return 0


if __name__ == "__main__":
    sys.exit(main())
