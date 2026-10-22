# -*- coding: utf-8 -*-
"""Confere os enums portados de um DFe contra a unit de conversao do ACBr.

Enum e o lugar onde o porte erra barato e caro ao mesmo tempo: trocar um
codigo ou deslocar um ordinal nao quebra build, nao quebra `vet`, e o
documento sai com o valor do vizinho. A SEFAZ aceita -- o campo e valido --
e o erro so aparece na fiscalizacao.

O script compara, enum a enum:

1. a QUANTIDADE de membros;
2. a SEQUENCIA de codigos, posicao a posicao.

Ele le o Delphi de duas formas, porque o ACBr usa duas:

  - `XxxArrayStrings: array[TXxx] of string = ('a', 'b');`
  - `StrToEnumerado(ok, s, ['a','b'], [memA, memB])` dentro da funcao

Quando as duas existem e DIVERGEM, vale a da funcao -- e ela que o codigo
chama. O script avisa quando isso acontece: no CT-e, a ArrayStrings de
TSchemaCTe tem dezoito entradas vazias e e letra morta.

O que ele NAO confere
---------------------
- o NOME do membro. Renomear nao muda o documento;
- a semantica. Dois codigos trocados entre si passam se as duas tabelas
  estiverem trocadas do mesmo jeito -- e por isso que os cortes
  deliberados existem;
- enum cujo codigo e derivado do nome do membro (TSchemaCTe do CT-e, por
  exemplo) -- la nao ha tabela a comparar.

Uso:

    py ferramentas/conferir-enums-dfe.py <conversao.pas> <pacote.go> \\
       [--par TXxxDelphi=NomeGo]...

    py ferramentas/conferir-enums-dfe.py \\
       "$ACBR_FONTES/ACBrDFe/ACBrCTe/PCNCTe/pcteConversaoCTe.pas" \\
       packages/cte/conversao.go packages/cte/types.go \\
       --par TpcteTipoCTe=tipoCTeCodigos \\
       --par TpcteTipoServico=tipoServicoCodigos

O par e necessario porque o nome da tabela Go nao se deduz do nome do tipo
Delphi: `TpcteTipoCTe` virou `tipoCTeCodigos`. Declarar o par explicitamente
e melhor que adivinhar -- adivinhacao errada vira "enum nao encontrado", que
e facil de ignorar.

Codigo de saida 1 se houver divergencia, para encaixar em CI.
"""
import io
import re
import sys


def ler(caminho):
    dados = io.open(caminho, 'rb').read()
    for codec in ('utf-8', 'cp1252'):
        try:
            return dados.decode(codec)
        except UnicodeDecodeError:
            continue
    return dados.decode('latin-1')


def tabelas_delphi(texto):
    """Devolve {tipo: [codigos]} lendo as duas formas do ACBr."""
    achadas = {}

    # Forma 1: XxxArrayStrings: array[TTipo] of string = ('a', 'b');
    for m in re.finditer(
            r'(\w+)\s*:\s*array\[(\w+)\]\s*of\s*string\s*=\s*\((.*?)\);',
            texto, re.S):
        vals = [v.strip().strip("'")
                for v in m.group(3).replace('\n', ' ').split(',') if v.strip()]
        achadas.setdefault(m.group(2), {})['array:' + m.group(1)] = vals

    # Forma 2: a lista literal dentro de StrToEnumerado/EnumeradoToStr.
    #
    # Casa a PRIMEIRA lista entre colchetes depois da chamada: e a dos
    # codigos; a segunda e a dos membros.
    impl = texto[texto.index('\nimplementation'):] if '\nimplementation' in texto else texto
    for bloco in re.split(r'\n(?=function )', impl):
        m = re.match(r'function (\w+)[^;]*?:\s*(\w+);', bloco, re.S)
        if not m or 'Enumerado' not in bloco:
            continue
        tipo = m.group(2)
        if tipo.lower() in ('string', 'boolean', 'integer', 'double'):
            # E um XxxToStr: o tipo de retorno nao diz o enum. Pega do
            # parametro.
            p = re.search(r'\(\s*const\s+t\s*:\s*(\w+)', bloco)
            if not p:
                continue
            tipo = p.group(1)
        lista = re.search(r"\[([^\]]*'[^\]]*)\]", bloco)
        if not lista:
            continue
        vals = [v.strip().strip("'")
                for v in lista.group(1).replace('\n', ' ').split(',') if v.strip()]
        achadas.setdefault(tipo, {})['func:' + m.group(1)] = vals

    return achadas


def tabelas_go(textos):
    """Devolve {nome: [codigos]} das tabelas `var xxx = [...]string{...}`."""
    achadas = {}
    for texto in textos:
        for m in re.finditer(
                r'var (\w+) = \[\.\.\.\]string\{(.*?)\}', texto, re.S):
            vals = [v.strip().strip('"')
                    for v in m.group(2).replace('\n', ' ').split(',') if v.strip()]
            achadas[m.group(1)] = vals
    return achadas


def escolher_funcao(candidatas):
    """Escolhe a funcao que devolve o CODIGO, nao a descricao.

    O ACBr costuma manter duas para o mesmo enum: uma com o codigo e outra
    com o texto de impressao -- tpMotivoToStr e tpMotivoToDesc,
    TpServPagToStr e TpServToStrText. Escolher por ordem alfabetica pega a
    descricao em metade dos casos e reporta divergencia inventada.
    """
    def prioridade(nome):
        n = nome.lower()
        if 'desc' in n or 'text' in n:
            return 2  # descricao: so se nao houver outra
        if n.startswith('func:strto'):
            return 0  # o leitor e a fonte mais confiavel do codigo
        return 1

    return sorted(candidatas.items(), key=lambda kv: (prioridade(kv[0]), kv[0]))[0]


def main(argv):
    pares = []
    resto = []
    i = 0
    while i < len(argv):
        if argv[i] == '--par' and i + 1 < len(argv):
            if '=' not in argv[i + 1]:
                print('par mal formado: %s' % argv[i + 1])
                return 2
            pares.append(tuple(argv[i + 1].split('=', 1)))
            i += 2
            continue
        resto.append(argv[i])
        i += 1

    if len(resto) < 2 or not pares:
        print(__doc__)
        return 2

    delphi = tabelas_delphi(ler(resto[0]))
    go = tabelas_go([ler(c) for c in resto[1:]])

    divergencias = 0
    avisos = 0

    for tipo, nomeGo in pares:
        print('== %s -> %s ==' % (tipo, nomeGo))

        if tipo not in delphi:
            print('  FALHA: tipo nao encontrado no .pas')
            divergencias += 1
            continue
        if nomeGo not in go:
            print('  FALHA: tabela nao encontrada no .go')
            divergencias += 1
            continue

        fontes = delphi[tipo]
        # A da FUNCAO vence a ArrayStrings quando as duas existem.
        daFuncao = {k: v for k, v in fontes.items() if k.startswith('func:')}
        doArray = {k: v for k, v in fontes.items() if k.startswith('array:')}

        if daFuncao:
            chave, esperado = escolher_funcao(daFuncao)
        else:
            chave, esperado = sorted(doArray.items())[0]

        if daFuncao and doArray:
            for k, v in doArray.items():
                # Array de DESCRICAO nunca bate com funcao de CODIGO, e o
                # aviso seria ruido. Comparar os dois exigiria casar
                # descricao com descricao, que nao muda documento.
                if 'desc' in k.lower():
                    continue
                if v != esperado:
                    print('  AVISO: %s diverge de %s -- vale a funcao' % (k, chave))
                    print('         array: %s' % v)
                    print('         func:  %s' % esperado)
                    avisos += 1

        obtido = go[nomeGo]
        if len(obtido) != len(esperado):
            print('  FALHA: %d codigos no .go, %d no .pas (%s)'
                  % (len(obtido), len(esperado), chave))
            divergencias += 1
            continue

        erros = [(i, a, b) for i, (a, b) in enumerate(zip(esperado, obtido)) if a != b]
        if erros:
            for i, a, b in erros:
                print('  FALHA: posicao %d -- .pas %r, .go %r' % (i, a, b))
            divergencias += len(erros)
        else:
            print('  OK: %d codigos identicos (%s)' % (len(obtido), chave))

    print()
    if avisos:
        print('%d aviso(s) de tabela divergente DENTRO do ACBr.' % avisos)
        print('Nao sao erro do porte, mas cada um e uma decisao a tomar.')
        print()
    if divergencias:
        print('TOTAL: %d divergencia(s)' % divergencias)
        return 1
    print('OK: todos os enums conferidos batem')
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
