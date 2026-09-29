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

FILL = r"\b(?:VLFill|LFill|RFill|DFill)\s*\(\s*([^,()]*(?:\([^()]*\))?[^,()]*)"


def le(path):
    for enc in ("cp1252", "utf-8", "latin-1"):
        try:
            return open(path, encoding=enc).read()
        except Exception:
            pass
    return ""


def acbr_writer(bloco, reg):
    src = le(os.path.join(ACBR, "ACBrEFDBloco_%s_Class.pas" % bloco))
    pat = r"procedure TBloco_%s\.WriteRegistro%s\b.*?(?=\nprocedure |\nfunction )" % (bloco, reg)
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


def extrai_fills(trecho):
    campos = []
    pat = r"b\.(?:LFillStr|LFillInt|LFillDate|LFillFloat|DFill|RFill)\(\s*([^,]+)"
    for arg in re.findall(pat, trecho):
        a = arg.strip()
        a = re.sub(r"^int64\(", "", a).rstrip(")")
        a = a.replace("r.", "").replace("b.", "")
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
    token = r"b\.(?:LFillStr|LFillInt|LFillDate|LFillFloat|DFill|RFill)\([^,]+|\b([a-z][A-Za-z0-9]*)\(b, r\)"
    for m2 in re.finditer(token, expr):
        if m2.group(1):
            campos.extend(extrai_fills(go_helper(m2.group(1))))
        else:
            campos.extend(extrai_fills(m2.group(0) + ","))
    return campos


def main():
    bloco, regs = sys.argv[1], sys.argv[2:]
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
        cond = sorted(set(re.findall(r"ifthen|IfThen|EncodeDate|DT_INI", expr)))
        marca = "OK  " if len(a) == len(g) else "DIF "
        extra = ("  [condicional: %s]" % ",".join(cond)) if cond else ""
        print("%s %s: ACBr=%d | Go=%d%s" % (marca, reg, len(a), len(g), extra))
        if len(a) != len(g):
            print("      ACBr: %s" % a)
            print("      Go  : %s" % g)


if __name__ == "__main__":
    main()
