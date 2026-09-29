// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-09-29.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package pcn

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ErrDataInvalida indica conteudo que nao pode ser interpretado como data.
var ErrDataInvalida = errors.New("pcn: data invalida")

// OnlyNumber devolve apenas os digitos da string.
// Porte de OnlyNumber (ACBrUtil.Strings).
func OnlyNumber(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// OnlyAlphaNum devolve apenas letras e digitos da string.
func OnlyAlphaNum(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		}
	}
	return b.String()
}

// StringToFloatDef converte uma string numerica em float64, devolvendo def
// quando a conversao falha. Porte de StringToFloatDef (ACBrUtil.Base).
//
// Divergencia documentada: o original depende do DecimalSeparator do
// sistema. Aqui o separador decimal e sempre o ULTIMO '.' ou ',' da string,
// e os anteriores sao tratados como separador de milhar. Para todo XML de
// DFe -- que usa ponto e nunca separa milhar -- o resultado e identico ao do
// Delphi, e deixa de variar conforme a maquina onde roda.
func StringToFloatDef(s string, def float64) float64 {
	v, err := StringToFloat(s)
	if err != nil {
		return def
	}
	return v
}

// StringToFloat converte uma string numerica em float64.
// Porte de StringToFloat (ACBrUtil.Base).
func StringToFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, strconv.ErrSyntax
	}

	ultPonto := strings.LastIndexByte(s, '.')
	ultVirg := strings.LastIndexByte(s, ',')

	sep := -1
	if ultPonto > ultVirg {
		sep = ultPonto
	} else if ultVirg >= 0 {
		sep = ultVirg
	}

	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '.', ',':
			if i == sep {
				b.WriteByte('.')
			}
			// separadores de milhar sao descartados
		default:
			b.WriteByte(c)
		}
	}
	return strconv.ParseFloat(b.String(), 64)
}

// StringDecimalToFloat converte uma string SEM separador decimal em float64,
// considerando as ultimas casas como decimais: ("10000", 2) vale 100,00 e
// ("123", 2) vale 1,23. Porte de StringDecimalToFloat (ACBrUtil.Base).
func StringDecimalToFloat(s string, casas int) (float64, error) {
	s = strings.TrimSpace(s)
	if casas < 0 {
		casas = 0
	}
	negativo := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	if s == "" {
		return 0, strconv.ErrSyntax
	}
	if len(s) < casas {
		s = strings.Repeat("0", casas-len(s)) + s
	}
	corte := len(s) - casas
	texto := s[:corte]
	if texto == "" {
		texto = "0"
	}
	if casas > 0 {
		texto += "." + s[corte:]
	}
	v, err := strconv.ParseFloat(texto, 64)
	if err != nil {
		return 0, err
	}
	if negativo {
		v = -v
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// Data e hora
// ---------------------------------------------------------------------------

// EncodeDataHora interpreta as formas de data e hora aceitas pelos leitores
// do ACBr. Porte de EncodeDataHora/ParseDataHora (ACBrUtil.DateTime).
//
// Formas reconhecidas:
//
//	AAAAMM                      -> primeiro dia do mes
//	AAAAMMDD  / AAAA-MM-DD
//	DD/MM/AAAA
//	AAAA-MM-DDThh:mm:ss
//	AAAA-MM-DDThh:mm:ss-03:00   (offset preservado)
//	AAAA-MM-DD hh:mm:ss
//	sufixo Z, fracao de segundo e nome de mes em PT/EN tambem sao aceitos
//
// Conteudo vazio devolve o tempo zero sem erro, como no original.
//
// Divergencia documentada: o ACBr DESCARTA o fuso horario (StringToDateTime
// remove o offset antes de converter). Aqui o offset e preservado como
// time.FixedZone. Campo a campo -- Year, Month, Day, Hour, Minute, Second --
// o resultado e identico ao do Delphi; o que se ganha e nao perder a
// informacao de fuso, que num documento fiscal identifica a UF emitente.
func EncodeDataHora(s string) (time.Time, error) {
	texto := strings.ToUpper(strings.TrimSpace(s))
	if texto == "" {
		return time.Time{}, nil
	}
	texto = converteNomeMes(texto)
	texto = strings.ReplaceAll(texto, "Z", "")

	dataStr, horaStr := texto, ""
	if p := strings.IndexAny(texto, "T "); p >= 0 {
		dataStr = texto[:p]
		horaStr = texto[p+1:]
	}
	if len(dataStr) > 10 {
		dataStr = dataStr[:10]
	}

	ano, mes, dia, err := parseData(dataStr)
	if err != nil {
		return time.Time{}, NovoErroLeitura("", s, err)
	}
	if ano == 0 || mes == 0 {
		// Reproduz o "0000/" do original: data zerada devolve tempo zero.
		return time.Time{}, nil
	}

	hora, min, seg, loc, err := parseHora(horaStr)
	if err != nil {
		return time.Time{}, NovoErroLeitura("", s, err)
	}

	return time.Date(ano, time.Month(mes), dia, hora, min, seg, 0, loc), nil
}

// EncodeDataHoraDef e como EncodeDataHora, mas devolve o tempo zero em vez
// de erro. E o comportamento dos leitores do ACBr, que nao conferem.
func EncodeDataHoraDef(s string) time.Time {
	t, err := EncodeDataHora(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ParseCompetencia interpreta uma competencia no formato AAAAMM e devolve o
// primeiro dia do mes. Conteudo vazio devolve o tempo zero sem erro.
func ParseCompetencia(s string) (time.Time, error) {
	texto := strings.TrimSpace(s)
	if texto == "" {
		return time.Time{}, nil
	}
	if len(texto) != 6 {
		return time.Time{}, fmt.Errorf("%w: competencia %q nao tem 6 digitos", ErrDataInvalida, s)
	}
	ano, err := strconv.Atoi(texto[:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: competencia %q", ErrDataInvalida, s)
	}
	mes, err := strconv.Atoi(texto[4:])
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: competencia %q", ErrDataInvalida, s)
	}
	if ano <= 0 || mes < 1 || mes > 12 {
		return time.Time{}, fmt.Errorf("%w: competencia %q fora de faixa", ErrDataInvalida, s)
	}
	return time.Date(ano, time.Month(mes), 1, 0, 0, 0, 0, time.UTC), nil
}

// ParseCompetenciaDef e como ParseCompetencia, mas devolve o tempo zero em
// vez de erro -- o que o ACBr faz ao ler CompetEmis, CompetApur e CompetFat.
func ParseCompetenciaDef(s string) time.Time {
	t, err := ParseCompetencia(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// FormatarCompetencia devolve a competencia no formato AAAAMM. Tempo zero
// devolve string vazia.
func FormatarCompetencia(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("200601")
}

// parseData interpreta a parte de data ja normalizada em maiusculas.
func parseData(s string) (ano, mes, dia int, err error) {
	s = strings.ReplaceAll(s, "-", "/")
	s = strings.ReplaceAll(s, ".", "/")

	if !strings.Contains(s, "/") {
		digitos := OnlyNumber(s)
		switch len(digitos) {
		case 0:
			return 0, 0, 0, nil
		case 6: // AAAAMM
			return atoi(digitos[:4]), atoi(digitos[4:6]), 1, nil
		case 8: // AAAAMMDD
			return atoi(digitos[:4]), atoi(digitos[4:6]), atoi(digitos[6:8]), nil
		default:
			return 0, 0, 0, fmt.Errorf("%w: %q", ErrDataInvalida, s)
		}
	}

	partes := strings.Split(s, "/")
	switch len(partes) {
	case 2: // AAAA/MM ou MM/AAAA
		if len(partes[0]) == 4 {
			return atoi(partes[0]), atoi(partes[1]), 1, nil
		}
		return atoi(partes[1]), atoi(partes[0]), 1, nil
	case 3:
		if len(partes[0]) == 4 { // AAAA/MM/DD
			return atoi(partes[0]), atoi(partes[1]), atoi(partes[2]), nil
		}
		// DD/MM/AAAA
		return atoi(partes[2]), atoi(partes[1]), atoi(partes[0]), nil
	default:
		return 0, 0, 0, fmt.Errorf("%w: %q", ErrDataInvalida, s)
	}
}

// parseHora interpreta a parte de hora, separando o offset de fuso.
func parseHora(s string) (hora, min, seg int, loc *time.Location, err error) {
	loc = time.UTC
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, 0, loc, nil
	}

	tzd := ""
	if p := strings.IndexAny(s, "-+"); p >= 0 {
		tzd = s[p:]
		s = s[:p]
	} else if p := strings.IndexByte(s, ' '); p >= 0 {
		tzd = s[p:]
		s = s[:p]
	}
	if p := strings.IndexByte(s, '.'); p >= 0 {
		s = s[:p]
	}

	partes := strings.Split(strings.TrimSpace(s), ":")
	if len(partes) > 0 && partes[0] != "" {
		hora = atoi(partes[0])
	}
	if len(partes) > 1 {
		min = atoi(partes[1])
	}
	if len(partes) > 2 {
		seg = atoi(partes[2])
	}
	if hora > 23 || min > 59 || seg > 59 {
		return 0, 0, 0, time.UTC, fmt.Errorf("%w: hora %q fora de faixa", ErrDataInvalida, s)
	}

	if l := parseOffset(tzd); l != nil {
		loc = l
	}
	return hora, min, seg, loc, nil
}

// parseOffset converte "-03:00" / "+0200" num time.Location fixo.
func parseOffset(s string) *time.Location {
	s = strings.TrimSpace(s)
	if len(s) < 3 || (s[0] != '-' && s[0] != '+') {
		return nil
	}
	sinal := 1
	if s[0] == '-' {
		sinal = -1
	}
	digitos := OnlyNumber(s[1:])
	if len(digitos) < 2 {
		return nil
	}
	h := atoi(digitos[:2])
	m := 0
	if len(digitos) >= 4 {
		m = atoi(digitos[2:4])
	}
	if h > 14 || m > 59 {
		return nil
	}
	segundos := sinal * (h*3600 + m*60)
	if segundos == 0 {
		return time.UTC
	}
	return time.FixedZone(fmt.Sprintf("%+03d:%02d", sinal*h, m), segundos)
}

// converteNomeMes troca nomes de mes em PT/EN pelo numero correspondente.
// Porte da funcao interna ConverteNomeMes de ParseDataHora.
func converteNomeMes(s string) string {
	if OnlyNumber(s) == OnlyAlphaNum(s) {
		return s
	}
	meses := [12]string{"JAN", "FEV", "MAR", "ABR", "MAI", "JUN", "JUL", "AGO", "SET", "OUT", "NOV", "DEZ"}
	months := [12]string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	for i := 0; i < 12; i++ {
		num := fmt.Sprintf("%02d", i+1)
		s = strings.ReplaceAll(s, meses[i], num)
		// O original so troca o nome em ingles dos meses cuja abreviacao
		// difere do portugues -- os demais ja foram trocados acima.
		switch i + 1 {
		case 2, 4, 5, 8, 9, 10, 12:
			s = strings.ReplaceAll(s, months[i], num)
		}
	}
	return s
}

func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

// NormatizarBoolean devolve "True" quando a string e "true" em qualquer
// caixa, e "False" no resto. Porte de NormatizarBoolean (ACBrXmlBase).
func NormatizarBoolean(s string) string {
	if strings.EqualFold(strings.TrimSpace(s), "true") {
		return "True"
	}
	return "False"
}

// RemoverDeclaracaoXML remove o prologo <?xml ... ?> do inicio do texto.
// Porte de RemoverDeclaracaoXML (ACBrUtil.XMLHTML).
func RemoverDeclaracaoXML(xml string) string {
	s := strings.TrimSpace(xml)
	for strings.HasPrefix(s, "<?xml") {
		p := strings.Index(s, "?>")
		if p < 0 {
			return s
		}
		s = strings.TrimSpace(s[p+2:])
	}
	return s
}

// RemoverCDATA remove TODAS as marcacoes CDATA do texto.
// Porte de RemoverCDATA (ACBrXmlBase), que usa rfReplaceAll.
func RemoverCDATA(s string) string {
	s = strings.ReplaceAll(s, "<![CDATA[", "")
	return strings.ReplaceAll(s, "]]>", "")
}

// RemoverCDATAPrimeiraOcorrencia remove apenas a PRIMEIRA ocorrencia de
// "<![CDATA[" e a primeira de "]]>".
//
// Existe para reproduzir fielmente a leitura de qrCodNFGas, onde o ACBr
// chama StringReplace SEM rfReplaceAll. E deliberado: ver
// ACBrNFGas.XmlReader.pas, Ler_InfNFGasSupl.
func RemoverCDATAPrimeiraOcorrencia(s string) string {
	s = strings.Replace(s, "<![CDATA[", "", 1)
	return strings.Replace(s, "]]>", "", 1)
}
