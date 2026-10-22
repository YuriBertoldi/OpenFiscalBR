# -*- coding: utf-8 -*-
"""Confere o porte do formato .ini de um DFe contra o par IniReader/IniWriter
do ACBr.

O .ini do ACBr nao tem schema publicado: o contrato e o proprio codigo. Isso
torna o porte dele especialmente sujeito a erro silencioso -- uma chave com
nome trocado nao quebra build, nao quebra vet, e so aparece quando o cliente
reclama que o campo sumiu.

Este script extrai as CHAVES que o Delphi le e grava, faz o mesmo com o Go, e
compara os dois conjuntos nos dois sentidos.

O que ele confere
-----------------
1. chaves lidas: toda chave do Ler_* existe no leitor Go, e vice-versa;
2. chaves gravadas: o mesmo para o Gerar_*;
3. SIMETRIA entre leitor e gravador do PROPRIO ACBr -- chave que ele grava e
   nao le (ou le e nao grava) e defeito do original, e o porte precisa
   decidir o que fazer com cada uma em vez de herdar o defeito sem perceber.

O que ele NAO confere
---------------------
- nome e largura de SECAO. O Delphi monta o nome em variavel
  (`sSecao := 'NVE' + IntToStrZero(Idx,3) + ...`) e o Go com Sprintf; comparar
  isso exigiria interpretar os dois. A largura do indice e justamente onde
  mais se erra, entao ela precisa de teste, nao de script;
- a CONDICIONAL em volta do grupo. Uma chave pode estar presente nos dois
  lados e sair em situacoes diferentes;
- o valor e a precisao.

Ou seja: saida limpa aqui NAO quer dizer porte correto. Quer dizer que
nenhuma chave foi esquecida, que e o erro mais comum e o mais chato de achar
a mao.

Uso:

    py ferramentas/conferir-ini-dfe.py <IniReader.pas> <IniWriter.pas> \\
       <leitor1.go> [<leitor2.go>...] -- <gravador1.go> [<gravador2.go>...]

    py ferramentas/conferir-ini-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.IniReader.pas" \\
       "$ACBR_FONTES/ACBrDFe/ACBrNFe/Base/ACBrNFe.IniWriter.pas" \\
       packages/nfe/ini_reader.go packages/nfe/ini_reader_total.go \\
       packages/nfe/ini_reader_det.go \\
       -- \\
       packages/nfe/ini_writer.go packages/nfe/ini_writer_det.go

Divergencia INTENCIONAL -- chave que o porte grava de proposito diferente do
original -- se declara com `--esperado`, separando por virgula:

    py ferramentas/conferir-ini-dfe.py ... --esperado mod

A lista nao esconde nada: ela aparece na saida, e qualquer chave FORA dela
continua sendo acusada. O que se ganha e um comando que passa em CI enquanto
as decisoes do porte ficam visiveis no proprio comando.

Codigo de saida 1 se houver divergencia nao declarada, para encaixar em CI.
"""
import io
import re
import sys

# ---------------------------------------------------------------------------
# Extracao do Delphi
# ---------------------------------------------------------------------------

# ReadString(sSecao, 'chave', ...) -- o segundo argumento e a chave.
#
# A busca e CASO-INSENSITIVA de proposito: o ACBr tem pelo menos uma chamada
# escrita `Writestring` com s minusculo (Gerar_ImpostoISSQN), e um regex
# sensivel a caixa a perderia -- aparecendo como "chave que o porte grava e o
# original nao".
LER_PAS = re.compile(r"Read(?:String|Integer|Float|DateTime|Bool)\s*\(", re.I)
GRAVAR_PAS = re.compile(r"Write(?:String|Integer|Float|DateTime|Bool)\s*\(", re.I)

# ---------------------------------------------------------------------------
# Extracao do Go
# ---------------------------------------------------------------------------

LER_GO = re.compile(r'\.Ler(?:String|Inteiro|Float|Data|Bool)\s*\(')
GRAVAR_GO = re.compile(r'\.Gravar(?:String|Inteiro|Float|Data|Bool)\s*\(')
# lerStringAlias(ini, secao, "c1", "c2", ...) -- varias chaves numa chamada.
ALIAS_GO = re.compile(r'ler(?:String|Inteiro|Float|Data)Alias\s*\(([^)]*)\)')


def segundo_argumento(texto, inicio, aspas):
    """Devolve o SEGUNDO argumento da chamada que abre em `inicio`.

    Nao da para usar `[^,]+` para pular o primeiro argumento: ele costuma ser
    uma chamada com virgulas dentro -- `fmt.Sprintf("NVE%03d%03d", i, j)` no
    Go, `'Produto' + IntToStrZero(Idx+1, 3)` no Delphi. Comparar com um regex
    ingenuo esconde justamente os grupos indexados, que sao a parte mais
    sujeita a erro do formato.

    Por isso o parser conta parenteses e so aceita a virgula de nivel zero.
    """
    nivel = 0
    i = inicio
    while i < len(texto):
        c = texto[i]
        if c == '(':
            nivel += 1
        elif c == ')':
            nivel -= 1
            if nivel < 0:
                return None
        elif c == ',' and nivel == 0:
            # Achou o fim do primeiro argumento; o proximo literal entre
            # aspas e a chave.
            resto = texto[i + 1:]
            m = re.match(r'\s*' + aspas + r'([^' + aspas + r']*)' + aspas, resto)
            return m.group(1) if m else None
        elif c in ("'", '"'):
            # Pula literal, para que virgula dentro de string nao conte.
            fim = texto.find(c, i + 1)
            if fim < 0:
                return None
            i = fim
        i += 1
    return None


def chaves_por_chamada(texto, padrao, aspas):
    """Extrai a chave de cada chamada que `padrao` encontra."""
    achadas = set()
    for m in padrao.finditer(texto):
        chave = segundo_argumento(texto, m.end(), aspas)
        if chave is not None:
            achadas.add(chave)
    return achadas


def ler_arquivo(caminho):
    """Le o arquivo tentando UTF-8 e caindo para CP-1252.

    Os fontes do ACBr sao ANSI; os do porte, UTF-8.
    """
    dados = io.open(caminho, 'rb').read()
    for codec in ('utf-8', 'cp1252'):
        try:
            return dados.decode(codec)
        except UnicodeDecodeError:
            continue
    return dados.decode('latin-1')


def chaves_pas(caminho, padrao):
    return chaves_por_chamada(ler_arquivo(caminho), padrao, "'")


def chaves_go(caminhos, padrao, com_alias):
    achadas = set()
    for caminho in caminhos:
        texto = ler_arquivo(caminho)
        achadas |= chaves_por_chamada(texto, padrao, '"')
        if com_alias:
            for args in ALIAS_GO.findall(texto):
                achadas |= set(re.findall(r'"([^"]*)"', args))
    return achadas


def normalizar(chaves):
    """Descarta o vazio e compara sem distinguir caixa.

    O TMemIniFile do Delphi nao distingue caixa, e o pcn.INI tambem nao --
    `nRe` e `nRE` sao a mesma chave. Comparar com caixa produziria uma lista
    de falsos positivos que esconderia os achados reais.
    """
    return {c.lower(): c for c in chaves if c.strip()}


def comparar(titulo, do_pas, do_go, ignorar):
    """Compara os dois conjuntos e devolve a quantidade de divergencias."""
    pas, go = normalizar(do_pas), normalizar(do_go)

    so_pas = sorted(pas[k] for k in set(pas) - set(go) - ignorar)
    so_go = sorted(go[k] for k in set(go) - set(pas) - ignorar)

    print('== %s ==' % titulo)
    print('  %d chaves no .pas, %d no .go, %d em comum'
          % (len(pas), len(go), len(set(pas) & set(go))))

    if so_pas:
        print('  SO NO .pas (%d): %s' % (len(so_pas), so_pas))
    if so_go:
        print('  SO NO .go  (%d): %s' % (len(so_go), so_go))
    if not so_pas and not so_go:
        print('  OK: nenhuma chave esquecida')
    print()
    return len(so_pas) + len(so_go)


def conferir_simetria(do_pas_ler, do_pas_gravar):
    """Aponta as chaves que o PROPRIO ACBr le e nao grava, ou o contrario.

    Nao e divergencia do porte -- e defeito do original --, mas cada uma
    dessas e uma decisao que o porte tem de tomar conscientemente, e e
    exatamente nelas que o round-trip perde dado.
    """
    ler, gravar = normalizar(do_pas_ler), normalizar(do_pas_gravar)

    so_le = sorted(ler[k] for k in set(ler) - set(gravar))
    so_grava = sorted(gravar[k] for k in set(gravar) - set(ler))

    print('== assimetria DENTRO do proprio ACBr ==')
    if so_le:
        print('  le e NAO grava (%d): %s' % (len(so_le), so_le))
    if so_grava:
        print('  grava e NAO le (%d): %s' % (len(so_grava), so_grava))
    if not so_le and not so_grava:
        print('  o leitor e o gravador do ACBr usam as mesmas chaves')
    print()
    print('  Cada uma acima e um ponto onde a ida e volta pelo .ini do ACBr')
    print('  perde informacao. Confira se o porte decidiu o que fazer com')
    print('  ela -- replicar ou corrigir -- e se a decisao esta registrada.')
    print()


def main(argv):
    esperado = set()
    if '--esperado' in argv:
        i = argv.index('--esperado')
        if i + 1 >= len(argv):
            print(__doc__)
            return 2
        esperado = {c.strip().lower() for c in argv[i + 1].split(',') if c.strip()}
        argv = argv[:i] + argv[i + 2:]

    if '--' not in argv:
        print(__doc__)
        return 2

    corte = argv.index('--')
    if corte < 3:
        print(__doc__)
        return 2

    pas_reader, pas_writer = argv[0], argv[1]
    go_leitores = argv[2:corte]
    go_gravadores = argv[corte + 1:]

    if not go_leitores or not go_gravadores:
        print(__doc__)
        return 2

    pas_ler = chaves_pas(pas_reader, LER_PAS)
    pas_gravar = chaves_pas(pas_writer, GRAVAR_PAS)
    go_ler = chaves_go(go_leitores, LER_GO, com_alias=True)
    go_gravar = chaves_go(go_gravadores, GRAVAR_GO, com_alias=False)

    if esperado:
        print('Divergencias declaradas como intencionais: %s'
              % sorted(esperado))
        print()

    divergencias = 0
    divergencias += comparar('chaves LIDAS', pas_ler, go_ler, esperado)
    divergencias += comparar('chaves GRAVADAS', pas_gravar, go_gravar, esperado)
    conferir_simetria(pas_ler, pas_gravar)

    if divergencias:
        print('TOTAL: %d chave(s) divergente(s) NAO declarada(s)' % divergencias)
        print()
        print('Antes de "corrigir": confira se a chave nao vem de um grupo')
        print('compartilhado (packages/rtc) ou se o nome nao e montado em')
        print('variavel no Delphi. O script so enxerga literal.')
        return 1

    print('OK: nenhuma chave esquecida nos dois sentidos')
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
