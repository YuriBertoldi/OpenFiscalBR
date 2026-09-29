"""Gera rascunho do writer Go de um registro SPED a partir do writer ACBr Delphi.

Uso: py ferramentas/gerar-writer.py <bloco> <registro> [<registro>...]

O rascunho SEMPRE precisa de revisao manual. O script sinaliza com // REVISAR
tudo que nao conseguiu resolver sozinho -- em especial condicionais de vigencia
e de versao de leiaute, que sao a principal fonte de erro no port.
"""
import re
import sys
import os
import glob

ACBR = r"C:\Sistemas\Componentes\DelphiXE3\ACBR\Fontes\ACBrTXT\ACBrSPED\ACBrSPEDFiscal"
GO = r"C:\GitHub\OpenFiscalBR\packages\sped"


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


def acbr_expr(corpo):
    for cand in re.findall(r"strLinha\s*:=\s*(.*?);\s*\n", corpo, re.S):
        if "Fill" in cand:
            return cand
    m = re.search(r"Add\(\s*((?:V?LFill|RFill|DFill).*?)\)\s*;", corpo, re.S)
    return m.group(1) if m else ""


def acbr_campos(expr):
    """Devolve [(func, arg, resto_dos_params)] na ordem da linha."""
    out = []
    for m in re.finditer(r"\b(VLFill|LFill|RFill|DFill)\s*\(([^;]*?)\)(?=\s*[+;)])", expr):
        func, dentro = m.group(1), m.group(2)
        # separa o primeiro argumento dos demais, respeitando parenteses
        nivel, corte = 0, len(dentro)
        for i, ch in enumerate(dentro):
            if ch == "(":
                nivel += 1
            elif ch == ")":
                nivel -= 1
            elif ch == "," and nivel == 0:
                corte = i
                break
        arg = dentro[:corte].strip()
        resto = [p.strip() for p in dentro[corte + 1:].split(",") if p.strip()]
        out.append((func, arg, resto))
    return out


def go_struct(reg):
    """Devolve {nome_normalizado: (NomeGo, tipo)} do struct RegistroXXX."""
    for p in glob.glob(os.path.join(GO, "bloco_*.go")):
        m = re.search(r"type Registro%s struct \{(.*?)\n\}" % reg, le(p), re.S)
        if not m:
            continue
        campos = {}
        for linha in m.group(1).split("\n"):
            linha = re.sub(r"//.*", "", linha).strip()
            if not linha:
                continue
            partes = linha.split()
            if len(partes) < 2:
                continue
            nome, tipo = partes[0], partes[1]
            if nome.startswith("Registro"):
                continue
            campos[nome.upper().replace("_", "")] = (nome, tipo)
        return campos
    return {}


BASICOS = {"string", "float64", "int", "int64", "time.Time", "bool"}


def traduz(func, arg, resto, campos):
    """Traduz uma chamada LFill do Delphi para a equivalente em Go."""
    # literal do registro
    if arg.startswith("'"):
        return 'b.LFillStr("%s", 0, false, \'0\')' % arg.strip("'")

    nome_acbr = re.sub(r"^\w+\(|\)$", "", arg).strip()  # tira Integer(...) etc.
    chave = nome_acbr.upper().replace("_", "").split(".")[-1]
    if chave not in campos:
        return "/* REVISAR: campo %s nao encontrado no struct */" % nome_acbr
    nome_go, tipo = campos[chave]

    nulo = "false"
    for p in resto:
        if p.lower() in ("true", "false"):
            nulo = p.lower()

    if tipo == "time.Time":
        return 'b.LFillDate(r.%s, "02012006", true)' % nome_go
    if tipo == "float64":
        dec = "2"
        if func == "DFill" and resto:
            dec = resto[0]
        elif len(resto) >= 2 and resto[1].isdigit():
            dec = resto[1]
        return "b.DFill(r.%s, %s, %s)" % (nome_go, dec, nulo)
    if tipo == "int":
        size = resto[0] if resto and resto[0].isdigit() else "0"
        return "b.LFillInt(int64(r.%s), %s, %s, '0')" % (nome_go, size, nulo)
    if tipo == "string":
        size = resto[0] if resto and resto[0].isdigit() else "0"
        return "b.LFillStr(r.%s, %s, %s, '0')" % (nome_go, size, nulo)
    # enum
    return "b.LFillStr(r.%s.String(), 0, false, '0')" % nome_go


def gera(bloco, reg):
    corpo = acbr_writer(bloco, reg)
    if not corpo:
        return "// REVISAR: writer de %s nao encontrado no ACBr\n" % reg
    expr = acbr_expr(corpo)
    campos = go_struct(reg)
    if not campos:
        return "// REVISAR: struct Registro%s nao encontrado no Go\n" % reg

    chamadas = [traduz(f, a, r, campos) for f, a, r in acbr_campos(expr)]
    if not chamadas:
        return "// REVISAR: nao consegui extrair os campos de %s\n" % reg

    avisos = []
    for marca in ("ifthen", "IfThen", "EncodeDate", "COD_VER"):
        if marca in expr or marca in corpo:
            avisos.append(marca)

    nomes = [re.sub(r"^\w+\(|\)$", "", a).upper().replace("_", "").split(".")[-1]
             for _, a, _ in acbr_campos(expr)]
    layout = "|".join([reg] + [n for n in nomes[1:]])

    out = []
    if avisos:
        out.append("// REVISAR: o writer Delphi tem %s -- confira a condicional."
                   % ", ".join(sorted(set(avisos))))
    out.append("// writeRegistro%s gera as linhas do registro %s." % (reg, reg))
    out.append("// Formato: |%s|" % layout)
    out.append("func (b *Bloco%s) writeRegistro%s(parent *RegistroPAI) { // REVISAR: pai"
               % (bloco.upper(), reg))
    out.append("\tfor _, r := range parent.Registro%s {" % reg)
    out.append("\t\tlinha := " + " +\n\t\t\t".join(chamadas))
    out.append("\t\tb.Add(linha, true)")
    out.append("\t\tb.Registro%s990.QtdLin%s++" % (bloco.upper(), bloco.upper()))
    out.append("\t\tb.Registro%sCount++" % reg)
    out.append("\t}")
    out.append("}")
    return "\n".join(out) + "\n"


if __name__ == "__main__":
    bloco, regs = sys.argv[1], sys.argv[2:]
    for reg in regs:
        print(gera(bloco, reg))
        print()
