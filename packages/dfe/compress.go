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

package dfe

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
)

// GzipBase64 comprime com gzip e codifica em base64 -- o
// EncodeBase64(GZipCompress(...)) que a recepcao sincrona da NFGas (e da
// NF3e/NFCom) aplica ao XML assinado antes de envelopar.
func GzipBase64(s string) (string, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(s)); err != nil {
		return "", fmt.Errorf("dfe: gzip: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("dfe: gzip: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// Base64Gunzip decodifica base64 e descomprime gzip -- o caminho inverso,
// usado quando a SEFAZ devolve conteudo comprimido.
func Base64Gunzip(s string) (string, error) {
	comprimido, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("dfe: base64: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(comprimido))
	if err != nil {
		return "", fmt.Errorf("dfe: gunzip: %w", err)
	}
	defer zr.Close()
	dados, err := io.ReadAll(zr)
	if err != nil {
		return "", fmt.Errorf("dfe: gunzip: %w", err)
	}
	return string(dados), nil
}
