"""Compara a sequencia de campos de um writer do ACBr Delphi com a do writer Go.

Uso: py comparar.py <bloco> <reg> [<reg>...]
     py comparar.py C 100 170 190
"""
import re
import sys
import os
import glob

ACBR = r"C:\Sistemas\Componentes\DelphiXE3\ACBR\Fontes\ACBrTXT\ACBrSPED\ACBrSPEDFiscal"
GO = r"C:\GitHub\OpenFiscalBR\packages\sped"

FILL = r"\b(?:VLFill|VDFill|LFill|RFill|DFill)\s*\(\s*([^,()]*(?:\([^()]*\))?[^,()]*)"


def le(path):
    for enc in ("cp1252", "utf-8", "latin-1"):
        try:
            return open(path, encoding=enc).read()
        except Exception:
            pass
    return ""


def tira_comentarios(src):
    """Remove comentarios Pascal (// ..., { ... }, (* ... *)).

    Sem isso, uma chamada LFill comentada e contada como campo -- foi o que
    aconteceu com o CNPJ_CPF do C800, que aparecia duas vezes.
    """
    src = re.sub(r"//[^\n]*", "", src)
    src = re.sub(r"\{.*?\}", "", src, flags=re.S)
    src = re.sub(r"\(\*.*?\*\)", "", src, flags=re.S)
    return src


def acbr_writer(bloco, reg):
    src = tira_comentarios(le(os.path.join(ACBR, "ACBrEFDBloco_%s_Class.pas" % bloco)))
    # o "\nend." fecha a unit: sem ele o ultimo writer de cada arquivo (um por
    # bloco, dez no total) aparecia como "nao encontrado no ACBr"
    pat = (r"procedure TBloco_%s\.WriteRegistro%s\b.*?(?=\nprocedure |\nfunction |\nend\.)"
           % (bloco, reg))
    m = re.search(pat, src, re.S | re.I)
    return m.group(0) if m else ""


def acbr_campos(corpo):
    """Devolve (campos, expressao). Cobre 'strLinha := LFill...' e 'Add( LFill... )'."""
    # Pode haver varias atribuicoes a strLinha (inicializacao, callbacks).
    # A que interessa e a que monta a linha, ou seja, a que contem LFill.
    expr = ""
    for cand in re.findall(r"strLinha\s*:=\s*(.*?);\s*\n", corpo, re.S):
        if "Fill" in cand:
            expr = cand
            break
    if not expr:
        m = re.search(r"Add\(\s*(LFill.*?)\)\s*;", corpo, re.S)
        if not m:
            return [], ""
        expr = m.group(1)
    campos = []
    for arg in re.findall(FILL, expr):
        a = arg.strip().strip("'").strip()
        if a:
            campos.append(a)
    return campos, expr


def go_writer(reg):
    for p in glob.glob(os.path.join(GO, "write_*.go")):
        src = le(p)
        pat = r"func \(b \*Bloco[0-9A-Za-z]+\) [Ww]riteRegistro%s\(.*?\n\}" % reg
        m = re.search(pat, src, re.S)
        if m:
            return m.group(0)
    return ""


def go_helper(nome):
    """Corpo de uma funcao auxiliar do package (campos condicionais extraidos)."""
    for p in glob.glob(os.path.join(GO, "*.go")):
        m = re.search(r"func %s\(.*?\n\}" % re.escape(nome), le(p), re.S)
        if m:
            return m.group(0)
    return ""


def go_metodo(nome):
    """Corpo de um metodo auxiliar do tipo func (b *BlocoX) nome(...)."""
    pat = r"func \(b \*Bloco[0-9A-Za-z]+\) %s\(.*?\n\}" % re.escape(nome)
    for p in glob.glob(os.path.join(GO, "*.go")):
        m = re.search(pat, le(p), re.S)
        if m:
            return m.group(0)
    return ""


def extrai_fills(trecho):
    campos = []
    pat = r"b\.(?:LFillStr|LFillInt|LFillDate|LFillFloat|VLFill|VDFill|DFill|RFill)\(\s*([^,]+)"
    for arg in re.findall(pat, trecho):
        a = arg.strip()
        a = re.sub(r"^int64\(", "", a).rstrip(")")
        # so o prefixo do receiver -- um replace solto quebra nomes que contem
        # "r." no meio, como CodVer.String() virando CodVeString()
        a = re.sub(r"^[rb]\.", "", a)
        a = re.sub(r"\.String\(\)$|\.StringEm\(DtIni\)$", "", a)
        campos.append(a.strip().strip('"'))
    return campos


def go_campos(corpo):
    m = re.search(r"linha :?=(.*?)\n\s*b\.Add", corpo, re.S)
    if not m:
        return []
    expr = m.group(1)
    campos = []
    # Percorre a expressao na ordem, expandindo chamadas a helpers locais
    # (campos condicionais extraidos para funcao propria).
    fills = "LFillStr|LFillInt|LFillDate|LFillFloat|VLFill|VDFill|DFill|RFill"
    token = (r"b\.(?:" + fills + r")\([^,]+"
             r"|\bb\.([a-z][A-Za-z0-9]*)\("
             r"|\b([a-z][A-Za-z0-9]*)\(b, r\)")
    for m2 in re.finditer(token, expr):
        nome = m2.group(1) or m2.group(2)
        if nome:
            campos.extend(extrai_fills(go_metodo(nome) or go_helper(nome)))
        else:
            campos.extend(extrai_fills(m2.group(0) + ","))
    return campos


# Variaveis locais que o writer Go usa no lugar do campo do struct. Sem isso a
# comparacao de nomes acusaria falso positivo em todo registro que trata
# documento cancelado, monta codigo por vigencia ou calcula tamanho.
ALIAS_GO = {
    "cod": "COD_SIT", "fin": "FIN_DOCe", "inddest": "IND_DEST",
    "pis": "VL_PIS", "cofins": "VL_COFINS",
    "pisst": "VL_PIS_ST", "cofinsst": "VL_COFINS_ST",
    "dtdoc": "DT_DOC", "dtes": "DT_E_S",
    "indfrt": "IND_FRT", "indpgto": "IND_PGTO",
    "ordinal": "IND_APUR", "tam": "NUM_DOC",
    # o ACBr batiza a local com nome proprio em vez do nome do campo
    "chvcte": "ChaveEletronicaCTe",
}


def normaliza(nome):
    """Reduz um argumento dos dois lados a uma chave comparavel.

    COD_PART / CodPart            -> CODPART
    VL_BC_ICMS_ST / VlBcICMSST    -> VLBCICMSST
    Integer(IND_OPER) / IndOper   -> INDOPER
    strCOD_SIT / CodSit.String()  -> CODSIT
    UmReg0220.UNID_CONV           -> UNIDCONV
    """
    n = nome.strip().strip("'").strip('"')
    # sufixo de metodo do lado Go: CodSit.String( , IndFrt.StringEmD100(DtIni
    # (vem truncado porque a extracao corta na primeira virgula)
    n = re.sub(r"\.String\w*\(.*$", "", n)
    # conteudo do wrapper: Integer(X), CodVerToStr(X), int64(r.X)
    while "(" in n:
        dentro = n[n.index("(") + 1:]
        dentro = dentro.rsplit(")", 1)[0] if ")" in dentro else dentro
        if not dentro.strip():
            break
        n = dentro
    n = n.split(".")[-1]                      # r.CodPart / UmReg.CAMPO
    n = re.sub(r"^(str|int|v)(?=[A-Z_])", "", n)  # strCOD_SIT, intIND_PROC, vFin_DOCe
    chave = n.upper().replace("_", "")
    return ALIAS_GO.get(n.lower(), chave).upper().replace("_", "")


def exibe(nome):
    """Devolve o argumento apresentavel na tabela do modo --doc.

    A extracao corta na primeira virgula, entao chamadas ficam com o parentese
    aberto (`CodVer.String(`). Aqui so fechamos o que ficou pendente, para a
    tabela nao publicar um nome truncado.
    """
    n = nome.strip()
    faltam = n.count("(") - n.count(")")
    return n + ")" * faltam if faltam > 0 else n


def compara_nomes(a, g):
    """Devolve a lista de (posicao, nome_acbr, nome_go) que nao batem."""
    fora = []
    for i, (x, y) in enumerate(zip(a, g)):
        if normaliza(x) != normaliza(y):
            fora.append((i, x, y))
    return fora


def main():
    argv = [x for x in sys.argv[1:] if x != "--doc"]
    doc = "--doc" in sys.argv
    bloco, regs = argv[0], argv[1:]
    for reg in regs:
        corpo_acbr = acbr_writer(bloco, reg)
        if not corpo_acbr:
            print("??   %s: writer nao encontrado no ACBr" % reg)
            continue
        a, expr = acbr_campos(corpo_acbr)
        g = go_campos(go_writer(reg))
        if not g:
            print("--   %s: sem writer no Go (ACBr tem %d campos)" % (reg, len(a)))
            continue

        if doc:
            marca = "" if len(a) == len(g) else " (ramos por versao, veja a nota do topo)"
            print("### %s%s" % (reg, marca))
            print("| # | ACBr | Go |")
            print("|---|------|-----|")
            for i in range(max(len(a), len(g))):
                print("| %d | %s | %s |" % (i,
                                            exibe(a[i]) if i < len(a) else "—",
                                            exibe(g[i]) if i < len(g) else "—"))
            print("")
            continue

        cond = sorted(set(re.findall(r"ifthen|IfThen|EncodeDate|DT_INI", expr)))
        extra = ("  [condicional: %s]" % ",".join(cond)) if cond else ""
        if len(a) != len(g):
            print("DIF  %s: ACBr=%d | Go=%d%s" % (reg, len(a), len(g), extra))
            print("      ACBr: %s" % a)
            print("      Go  : %s" % g)
            continue
        fora = compara_nomes(a, g)
        if fora:
            print("NOME %s: %d campos, %d com nome divergente%s"
                  % (reg, len(a), len(fora), extra))
            for i, x, y in fora:
                print("      campo %2d: ACBr %-28s | Go %s" % (i, x, y))
        else:
            print("OK   %s: ACBr=%d | Go=%d%s" % (reg, len(a), len(g), extra))


if __name__ == "__main__":
    main()
