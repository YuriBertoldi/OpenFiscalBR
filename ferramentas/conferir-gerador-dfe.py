#!/usr/bin/env python3
"""Confere um gerador de DFe portado contra o XmlWriter Delphi de origem.

Compara, tag a tag, as duas decisoes que o build, o `go vet` e o round-trip
NAO alcancam:

1. a PRECISAO de cada campo (`tcDe2` x `NodeDec(..., 2, ...)`, `tcDe4` x
   `NodeDec(..., 4, ...)`, `FormatoValor4ou2` x `nodeAliquota`, ...);
2. a OBRIGATORIEDADE de cada campo -- o `nOcorrencias` do `AddNode`, que vale
   1 (sai mesmo zerado) ou 0 (some quando vazio).

Por que um script e nao um teste Go: o round-trip (gerar -> ler -> gerar) e
cego para as duas coisas. A precisao some porque o leitor converte "0.00" e
"0.0000" no mesmo float, e a obrigatoriedade some porque um campo zerado lido
de volta continua zerado. So a comparacao com o `.pas` distingue.

3. no modo `--ordem`, a SEQUENCIA das tags de um grupo -- o terceiro eixo cego
   ao round-trip, e o mais caro de errar, porque o XSD e `xs:sequence` e campo
   fora de ordem e documento rejeitado pela SEFAZ.

Uso:

    py ferramentas/conferir-gerador-dfe.py <arquivo.pas> <arquivo.go> \\
       [--linhas INICIO:FIM]

    py ferramentas/conferir-gerador-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.XmlWriter.pas" \\
       packages/nfe/xml_writer_det.go --linhas 1118:3507

O recorte por linha serve para comparar um PEDACO do writer Delphi (que e um
arquivo so, de milhares de linhas) com o arquivo Go correspondente. Sem ele a
comparacao mistura grupos e produz ruido.

Modo ordem -- imprime as duas sequencias lado a lado, para conferencia visual:

    py ferramentas/conferir-gerador-dfe.py <pas> <go> \\
       --ordem INICIO:FIM <funcaoGo> [prefixoDoCase]

    py ferramentas/conferir-gerador-dfe.py "$PAS" packages/nfe/xml_writer_det.go \\
       --ordem 2097:2199 gerarICMSRegimeNormal pcn.CST10

O modo ordem e INTERATIVO de proposito: o mapeamento entre faixa do `.pas` e
funcao Go muda a cada etapa do porte e envelheceria guardado num arquivo.

Ele reprova (codigo 1) SO quando a ordem RELATIVA das tags comuns diverge, que
e o que o XSD cobra. Tag presente num lado so nao reprova -- aparece numa lista
a parte, e e esperada em tres casos:

  - sub-writer que o porte INLINOU aparece so do lado Go (confira a POSICAO em
    que ele entra);
  - sub-writer que o porte manteve em funcao propria some dos DOIS lados;
  - campo que o Delphi escreve com dois `AddNode` literais (um por ramo do
    `if`) aparece DUPLICADO so do lado `.pas`.

Saida: uma linha por divergencia e um resumo. Codigo de saida 1 se houver
divergencia, para encaixar em CI.
"""

import collections
import re
import sys

# Tipo do AddNode no Delphi -> rotulo normalizado usado na comparacao.
#
# Os dois "Formato" sao campos do writer cujo valor depende da versao do
# leiaute; no Go viraram helpers com o mesmo papel.
TIPOS_DELPHI = {
    "tcStr": "Str", "tcEsp": "Str", "tcNumStr": "NumStr", "tcStrOrig": "Str",
    "tcInt": "Int", "tcInt64": "Int",
    "tcDe1": "1", "tcDe2": "2", "tcDe3": "3", "tcDe4": "4",
    "tcDe5": "5", "tcDe6": "6", "tcDe7": "7", "tcDe8": "8", "tcDe10": "10",
    "tcDat": "Dat", "tcDatHor": "Dat",
    "tcHor": "Hor", "tcBool": "Bool", "tcBoolStr": "Bool",
    "FormatoValor4ou2": "Aliq",
    "FormatoValor10ou4": "ValorUnit",
    "FormatoAliq4ou2": "AliqLocal",
}

# Construtor de node no Go -> (rotulo do tipo, indice do argumento que carrega
# a precisao, indice do argumento que carrega a obrigatoriedade).
#
# None no indice de precisao significa que o tipo ja define a precisao.
CONSTRUTORES_GO = {
    "pcn.NodeStr": ("Str", None, 2),
    "pcn.NodeStrSemFiltro": ("Str", None, 2),
    "pcn.NodeInt": ("Int", None, 3),
    "pcn.NodeDec": (None, 2, 3),
    "pcn.NodeDat": ("Dat", None, 2),
    "nodeNumStr": ("NumStr", None, 3),
    "nodeIntDeTexto": ("Int", None, 3),
    "nodeHora": ("Hor", None, 2),
    "w.nodeAliquota": ("Aliq", None, 2),
    "w.nodeValorUnitario": ("ValorUnit", None, None),
}

# Helpers do porte que escondem uma decisao de precisao DENTRO deles, e por
# isso nao cabem na tabela acima: declare aqui o conjunto de pares
# (tipo, obrigatoriedade) que o helper e capaz de emitir, copiado do `.pas`.
#
# Acrescentar um helper aqui e assumir a responsabilidade de conferir o corpo
# dele a mao uma vez -- em troca, a tag volta a ser comparada automaticamente.
HELPERS_GO = {
    # nodePercentualGLP: 100 sai tcDe2 opcional; qualquer outro, tcDe4
    # obrigatorio (XmlWriter.pas:1670-1689).
    "nodePercentualGLP": {("2", "nao"), ("4", "sim")},
}


def dividir_argumentos(texto, inicio):
    """Divide os argumentos de uma chamada respeitando parenteses e aspas.

    `inicio` aponta para o '(' da chamada. Devolve (lista_de_args, pos_apos).
    Um splitter ingenuo por virgula quebra nos dois lados: no Delphi por causa
    de `IfThen(FpUsar_tcDe4, 07, 05)` no argumento de tamanho, e no Go por
    causa de chamadas como `w.forcaTagRejeicao938()`.
    """
    args, atual = [], []
    profundidade = 0
    aspas = None
    i = inicio
    while i < len(texto):
        c = texto[i]
        if aspas:
            atual.append(c)
            if c == "\\":
                if i + 1 < len(texto):
                    atual.append(texto[i + 1])
                    i += 2
                    continue
            elif c == aspas:
                aspas = None
        elif c in "'\"":
            aspas = c
            atual.append(c)
        elif c in "([":
            profundidade += 1
            if profundidade > 1:
                atual.append(c)
        elif c in ")]":
            profundidade -= 1
            if profundidade == 0:
                args.append("".join(atual).strip())
                return args, i + 1
            atual.append(c)
        elif c == "," and profundidade == 1:
            args.append("".join(atual).strip())
            atual = []
        else:
            atual.append(c)
        i += 1
    return args, i


def sem_comentarios_pas(fonte):
    """Remove comentarios de linha e de bloco do Delphi.

    Campo comentado nao deve entrar na conferencia -- o ACBr mantem varios
    (indISSRet, RENAVAM, as tags da NT 2016.002) e eles nao saem no XML.
    """
    fonte = re.sub(r"\(\*.*?\*\)", "", fonte, flags=re.S)
    fonte = re.sub(r"\{.*?\}", "", fonte, flags=re.S)
    return re.sub(r"^\s*//.*$", "", fonte, flags=re.M)


def recortar(fonte, faixa):
    """Mantem apenas as linhas da faixa INICIO:FIM (1-based, inclusiva)."""
    if not faixa:
        return fonte
    inicio, fim = (int(p) for p in faixa.split(":"))
    linhas = fonte.splitlines()
    return "\n".join(linhas[inicio - 1:fim])


def normalizar_ocorrencia(bruto):
    """Reduz o nOcorrencias do Delphi a 'sim', 'nao' ou 'calculado'."""
    bruto = bruto.strip()
    if bruto == "1":
        return "sim"
    if bruto == "0":
        return "nao"
    return "calculado"


def normalizar_obrigatorio_go(bruto):
    """Reduz o argumento de obrigatoriedade do Go a 'sim', 'nao' ou
    'calculado'."""
    bruto = bruto.strip()
    if bruto == "true":
        return "sim"
    if bruto == "false":
        return "nao"
    return "calculado"


def ler_pas(caminho, faixa):
    """Extrai {tag: {(tipo, ocorrencia)}} do XmlWriter Delphi."""
    try:
        fonte = open(caminho, encoding="cp1252").read()
    except FileNotFoundError:
        sys.exit(f"fonte Delphi nao encontrado: {caminho}")

    # O recorte vem ANTES da limpeza de comentarios para que os numeros de
    # linha informados pelo usuario batam com o arquivo original.
    fonte = sem_comentarios_pas(recortar(fonte, faixa))

    campos = collections.defaultdict(set)
    for m in re.finditer(r"\bAddNode\s*\(", fonte):
        args, _ = dividir_argumentos(fonte, m.end() - 1)
        if len(args) < 6:
            continue
        tipo_bruto = args[0].strip()
        tag = args[2].strip().strip("'").strip()
        if not tag:
            continue
        tipo = TIPOS_DELPHI.get(tipo_bruto, f"?{tipo_bruto}")
        campos[tag].add((tipo, normalizar_ocorrencia(args[5])))

    return campos


def ler_go(caminho):
    """Extrai {tag: {(tipo, obrigatoriedade)}} do gerador Go."""
    try:
        fonte = open(caminho, encoding="utf-8").read()
    except FileNotFoundError:
        sys.exit(f"gerador Go nao encontrado: {caminho}")

    campos = collections.defaultdict(set)
    for nome, (rotulo, idx_precisao, idx_obrig) in CONSTRUTORES_GO.items():
        for m in re.finditer(re.escape(nome) + r"\s*\(", fonte):
            args, _ = dividir_argumentos(fonte, m.end() - 1)
            if not args or not args[0].startswith('"'):
                continue
            tag = args[0].strip('"')
            tipo = rotulo
            if idx_precisao is not None:
                if len(args) <= idx_precisao:
                    continue
                tipo = args[idx_precisao].strip()
            obrig = "sim"
            if idx_obrig is not None:
                if len(args) <= idx_obrig:
                    continue
                obrig = normalizar_obrigatorio_go(args[idx_obrig])
            campos[tag].add((tipo, obrig))

    for nome, pares in HELPERS_GO.items():
        for m in re.finditer(r"\b" + re.escape(nome) + r"\s*\(", fonte):
            args, _ = dividir_argumentos(fonte, m.end() - 1)
            if not args or not args[0].startswith('"'):
                continue
            campos[args[0].strip('"')] |= pares

    return campos


# Rotulos de precisao que a comparacao sabe interpretar. Qualquer outra coisa
# no argumento de casas do NodeDec e expressao calculada em tempo de execucao.
PRECISOES_CONHECIDAS = {
    "1", "2", "3", "4", "5", "6", "7", "8", "10",
    "Str", "NumStr", "Int", "Dat", "Hor", "Bool", "Aliq", "ValorUnit",
}


# Construtores de node do Go, para a extracao de ordem. Mesma lista das duas
# tabelas acima, achatada.
CONSTRUTORES_ORDEM = sorted(
    (n.rsplit(".", 1)[-1] for n in list(CONSTRUTORES_GO) + list(HELPERS_GO)),
    key=len,
    reverse=True,
)


def ordem_pas(caminho, faixa):
    """Sequencia de tags do trecho indicado do writer Delphi."""
    fonte = sem_comentarios_pas(
        recortar(open(caminho, encoding="cp1252").read(), faixa)
    )
    return re.findall(r"AddNode\(\s*\w+\s*,\s*'[^']*'\s*,\s*'([^']+)'", fonte)


def ordem_go(caminho, funcao, prefixo_case=None):
    """Sequencia de tags de uma funcao Go, opcionalmente de um `case` dela."""
    linhas = open(caminho, encoding="utf-8").read().splitlines()
    try:
        ini = next(
            i for i, l in enumerate(linhas)
            if l.startswith("func ") and funcao + "(" in l
        )
    except StopIteration:
        sys.exit(f"funcao Go nao encontrada: {funcao}")
    fim = next(
        (i for i in range(ini + 1, len(linhas)) if linhas[i].startswith("func ")),
        len(linhas),
    )
    corpo = "\n".join(linhas[ini:fim])

    if prefixo_case:
        partes = re.split(r"^\tcase ", corpo, flags=re.M)
        corpo = next((p for p in partes if p.startswith(prefixo_case)), "")
        if not corpo:
            sys.exit(f"case nao encontrado em {funcao}: {prefixo_case}")

    corpo = re.sub(r"^\s*//.*$", "", corpo, flags=re.M)
    padrao = "|".join(CONSTRUTORES_ORDEM)
    return re.findall(r'(?:' + padrao + r')\(\s*"([^"]+)"', corpo)


def comparar_ordem(pas, go, faixa, funcao, prefixo_case):
    """Imprime as duas sequencias e as diferencas, sem reprovar."""
    seq_pas = ordem_pas(pas, faixa)
    seq_go = ordem_go(go, funcao, prefixo_case)

    print(f"== ordem: .pas {faixa}  x  go {funcao}"
          + (f" / case {prefixo_case}" if prefixo_case else ""))
    print(f"  pas ({len(seq_pas):2}): {' '.join(seq_pas)}")
    print(f"  go  ({len(seq_go):2}): {' '.join(seq_go)}")

    if seq_pas == seq_go:
        print("\n  sequencias identicas")
        return 0

    so_pas = [t for t in seq_pas if t not in seq_go]
    so_go = [t for t in seq_go if t not in seq_pas]
    if so_pas:
        print(f"\n  so no .pas: {so_pas}")
    if so_go:
        print(f"  so no go  : {so_go}")

    # Fora as tags exclusivas de um lado, a ORDEM relativa das comuns tem de
    # bater -- e e isso que o XSD cobra.
    comuns_pas = [t for t in seq_pas if t in seq_go]
    comuns_go = [t for t in seq_go if t in seq_pas]
    # Tag repetida no .pas e o mesmo campo em dois ramos de um `if`; colapsa.
    comuns_pas = [t for i, t in enumerate(comuns_pas) if i == 0 or comuns_pas[i - 1] != t]
    comuns_go = [t for i, t in enumerate(comuns_go) if i == 0 or comuns_go[i - 1] != t]

    if comuns_pas == comuns_go:
        print("\n  ordem relativa das tags comuns: OK")
        return 0

    print("\n  ORDEM RELATIVA DIVERGE nas tags comuns:")
    print(f"    pas: {' '.join(comuns_pas)}")
    print(f"    go : {' '.join(comuns_go)}")
    return 1


def main():
    if "--ordem" in sys.argv:
        i = sys.argv.index("--ordem")
        resto = sys.argv[i + 1:]
        if len(resto) < 2:
            sys.exit(__doc__)
        pas, go = sys.argv[1], sys.argv[2]
        return comparar_ordem(
            pas, go, resto[0], resto[1], resto[2] if len(resto) > 2 else None
        )

    argv = [a for a in sys.argv[1:] if not a.startswith("--")]
    faixa = None
    for a in sys.argv[1:]:
        if a.startswith("--linhas"):
            faixa = a.split("=", 1)[1] if "=" in a else None
    if faixa is None and "--linhas" in sys.argv:
        i = sys.argv.index("--linhas")
        if i + 1 < len(sys.argv):
            faixa = sys.argv[i + 1]
            argv = [a for a in argv if a != faixa]

    if len(argv) != 2:
        sys.exit(__doc__)

    pas = ler_pas(argv[0], faixa)
    go = ler_go(argv[1])

    falhas = 0
    conferidas = 0
    sem_par = []
    indecidiveis = []

    # Uma expressao em vez de literal (no Go) nao e divergencia: e um campo
    # que o porte resolveu com UMA chamada condicional onde o Delphi usou
    # DUAS chamadas literais. A ferramenta nao consegue decidir, entao
    # reporta como pendencia de conferencia manual em vez de reprovar -- uma
    # ferramenta permanentemente vermelha deixa de ser lida.
    print("== tipo e obrigatoriedade por campo ==")
    for tag in sorted(pas):
        if tag not in go:
            sem_par.append((tag, sorted(pas[tag])))
            continue
        conferidas += 1

        tipos_pas = {t for t, _ in pas[tag]}
        tipos_go = {t for t, _ in go[tag]}
        if tipos_pas != tipos_go:
            if tipos_go - PRECISOES_CONHECIDAS:
                indecidiveis.append(
                    f"tipo de {tag}: delphi={sorted(tipos_pas)}"
                    f" go={sorted(tipos_go)} (expressao)"
                )
            else:
                falhas += 1
                print(
                    f"  TIPO     {tag:20} delphi={sorted(tipos_pas)}"
                    f"  go={sorted(tipos_go)}"
                )

        ocorr_pas = {o for _, o in pas[tag]}
        ocorr_go = {o for _, o in go[tag]}
        if ocorr_pas != ocorr_go:
            if "calculado" in ocorr_go:
                indecidiveis.append(
                    f"ocorrencia de {tag}: delphi={sorted(ocorr_pas)}"
                    f" go={sorted(ocorr_go)}"
                )
            else:
                falhas += 1
                print(
                    f"  OCORRE   {tag:20} delphi={sorted(ocorr_pas)}"
                    f"  go={sorted(ocorr_go)}"
                )

    print(f"  {conferidas} tags conferidas")

    if indecidiveis:
        print(
            f"\n  {len(indecidiveis)} campo(s) decididos por expressao no Go"
            " -- confira o corpo da condicional contra o .pas:"
        )
        for linha in indecidiveis:
            print(f"    {linha}")

    if sem_par:
        print(
            f"\n  {len(sem_par)} tag(s) do .pas sem construtor de node no Go"
            " -- confira uma a uma se foram omitidas de proposito:"
        )
        for tag, pares in sem_par:
            print(f"    {tag} ({pares})")

    so_go = sorted(set(go) - set(pas))
    if so_go:
        print(f"\n  {len(so_go)} tag(s) so no Go: {so_go}")

    print()
    if falhas:
        print(f"{falhas} divergencia(s) -- confira cada uma contra o .pas")
        return 1
    print("OK: nenhuma divergencia de tipo nem de obrigatoriedade")
    return 0


if __name__ == "__main__":
    sys.exit(main())
