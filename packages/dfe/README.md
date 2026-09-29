# packages/dfe — Infraestrutura de transmissão de DFe (Layer 2)

Porte dos papéis de `ACBrDFe.pas`, `ACBrDFeSSL.pas` e `ACBrDFeWebService.pas`: certificado
digital A1, assinatura XMLDSig e cliente SOAP 1.2 com TLS mútuo. É a base de emissão de
qualquer DFe (NFGas hoje; NFe/CTe/MDFe quando forem portados).

> Este é o **único** package autorizado a importar `net/http` — como CLIENTE dos web services
> da SEFAZ (ver seção "Fronteira entre package e demo" do `CLAUDE.md`). Servidor, handler e
> JSON continuam proibidos aqui.

## Arquivos

| Arquivo | Conteúdo | Origem Delphi |
|---|---|---|
| `certificado.go` | `Certificado` (PFX via go-pkcs12, PEM via stdlib), `TLSCertificate`, `CNPJ()` da extensão ICP-Brasil | `ACBrDFeSSL.pas` (CarregarCertificado, CertCNPJ...) |
| `c14n.go` | `CanonicalizarFragmento` — Canonical XML 1.0 sem comentários, com transform enveloped-signature | delegado a libxml2/xmlsec no ACBr |
| `assinatura.go` | `AssinarXML`, `VerificarAssinatura`, `HashCSRT`, `AssinarSHA1Base64` | `TDFeSSL.Assinar` + template `TSignature.GerarXML` |
| `soap.go` | `ClienteSOAP`, `MontarEnvelope` (soap12), `ExtrairResultadoSOAP` (Fault + `*ResultMsg` caso-insensitivo) | `TDFeWebService.Executar`/`EnviarDados` |
| `compress.go` | `GzipBase64` / `Base64Gunzip` | `EncodeBase64(GZipCompress(...))` |
| `errors.go` | sentinelas `errors.Is` + `ErroTransmissao` com URL/ação/status HTTP | — |

## Uso

```go
cert, err := dfe.CarregarPFXArquivo("emitente.pfx", "senha")
assinado, err := dfe.AssinarXML(cert, xmlNFGas, "infNFGas")
err = dfe.VerificarAssinatura(assinado)

cliente := dfe.NovoClienteSOAP(cert, 0) // timeout 0 = TimeoutPadrao (90s)
resultado, err := cliente.Chamar(ctx, url, servico, soapAction,
    "nfgasDadosMsg", "nfgasResultMsg", dadosMsg)
```

## Decisões e limites

- **RSA-SHA1 e C14N 1.0** são exigência dos leiautes da SEFAZ, não escolha do porte.
- **go-pkcs12** (software.sslmate.com/src/go-pkcs12) no lugar do `golang.org/x/crypto/pkcs12`
  indicado no CLAUDE.md: o pacote do x/crypto está congelado e não decodifica PFX modernos da
  ICP-Brasil (PBES2/AES-256). Divergência documentada no doc.go.
- A canonicalização cobre o subconjunto de XML que ocorre em DFe: elementos com/sem prefixo,
  atributos sem prefixo (além de `xmlns`/`xml:`), texto e CDATA. Atributo com outro prefixo
  devolve erro em vez de canonicalizar errado. Whitespace entre elementos é preservado (a
  serialização trabalha sobre os bytes originais), então verificação de XML identado funciona.
- `VerificarAssinatura` confere digest e RSA contra o certificado do KeyInfo; **não** valida a
  cadeia ICP-Brasil (exige repositório de ACs — a cargo do consumidor).

## Testes

`go test ./packages/dfe/ -v` — sem dependência externa e sem rede: certificado RSA gerado em
teste, servidor SOAP via `httptest`. Cobrem C14N (ordenação de atributos, namespaces herdados,
escapes, remoção de Signature), assinar→verificar (round-trip, adulteração detectada,
reassinatura), PFX/PEM, gzip+base64 e o envelope/extração SOAP byte a byte.
