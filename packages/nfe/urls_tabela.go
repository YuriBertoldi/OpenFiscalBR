// OpenFiscalBR - Automacao Fiscal Brasileira em Go
// Derivado do Projeto ACBr (http://projetoacbr.com.br)
// Copyright (c) 2004-2026 Projeto ACBr - Daniel Simoes de Almeida
//
// Este arquivo e parte do OpenFiscalBR, derivado do Projeto ACBr.
// Portado de Delphi para Go em 2026-10-21.
//
// Esta biblioteca e software livre; voce pode redistribui-la sob os
// termos da Licenca Publica Geral Menor GNU (LGPL v2.1+).
// Veja LICENSE.TXT para detalhes.

package nfe

// Tabela de URLs publicas da NF-e e da NFC-e.
//
// Porte dos dados de ACBrNFeServicos.ini (secoes NFe_<UF>_<P|H> e
// NFCe_<UF>_<P|H>), limitado as duas familias que o QR-Code e o urlChave
// consomem: URL-QRCode e URL-Consulta. Os endpoints dos web services entram
// junto com quem os transmite.
//
// O .ini resolve uma consulta em tres passos (ACBrDFe.pas:456-607): chave
// versionada exata, senao a maior versao <= a pedida, senao a chave sem
// versao; e, se ainda assim nada, segue o redirecionamento "Usar=" para
// outra secao. Aqui as tres etapas ja estao RESOLVIDAS nos valores abaixo --
// o redirecionamento e detalhe de codificacao do .ini, nao regra fiscal.
//
// ferramentas/conferir-urls-dfe.py reimplementa a resolucao do ACBr sobre o
// .ini e compara com estas tabelas, entrada a entrada.

// urlsQRCode e a URL base do QR-Code por UF, para o modelo 65 (NFC-e).
// Extraida de ACBrNFeServicos.ini, chave URL-QRCode das secoes NFCe_<UF>_<P|H>,
// ja com os redirecionamentos "Usar=" resolvidos.
//
// A URL NAO varia com a versao do QR-Code: as tres versoes resolvem para o
// mesmo valor nas 54 combinacoes de UF e ambiente. Conferido por
// ferramentas/conferir-urls-dfe.py.
//
// Atencao ao SC, que ja traz "?p=" na propria URL -- e por isso que o ACBr
// testa a presenca do sufixo antes de acrescenta-lo.
var urlsQRCode = map[string]urlPorAmbiente{
	"AC": {"http://www.sefaznet.ac.gov.br/nfce/qrcode", "http://www.hml.sefaznet.ac.gov.br/nfce/qrcode"},
	"AL": {"http://nfce.sefaz.al.gov.br/QRCode/consultarNFCe.jsp", "http://nfce.sefaz.al.gov.br/QRCode/consultarNFCe.jsp"},
	"AP": {"https://www.sefaz.ap.gov.br/nfce/nfcep.php", "https://www.sefaz.ap.gov.br/nfcehml/nfce.php"},
	"AM": {"http://sistemas.sefaz.am.gov.br/nfceweb/consultarNFCe.jsp", "http://homnfce.sefaz.am.gov.br/nfceweb/consultarNFCe.jsp"},
	"BA": {"http://nfe.sefaz.ba.gov.br/servicos/nfce/qrcode.aspx", "http://hnfe.sefaz.ba.gov.br/servicos/nfce/qrcode.aspx"},
	"CE": {"http://nfce.sefaz.ce.gov.br/pages/ShowNFCe.html", "http://nfceh.sefaz.ce.gov.br/pages/ShowNFCe.html"},
	"DF": {"http://www.fazenda.df.gov.br/nfce/qrcode", "http://dec.fazenda.df.gov.br/ConsultarNFCe.aspx"},
	"ES": {"http://app.sefaz.es.gov.br/ConsultaNFCe/qrcode.aspx", "http://homologacao.sefaz.es.gov.br/ConsultaNFCe/qrcode.aspx"},
	"GO": {"https://nfeweb.sefaz.go.gov.br/nfeweb/sites/nfce/danfeNFCe", "https://nfewebhomolog.sefaz.go.gov.br/nfeweb/sites/nfce/danfeNFCe"},
	"MA": {"http://www.nfce.sefaz.ma.gov.br/portal/consultarNFCe.jsp", "http://www.hom.nfce.sefaz.ma.gov.br/portal/consultarNFCe.jsp"},
	"MT": {"http://www.sefaz.mt.gov.br/nfce/consultanfce", "http://homologacao.sefaz.mt.gov.br/nfce/consultanfce"},
	"MS": {"http://www.dfe.ms.gov.br/nfce/qrcode", "http://www.dfe.ms.gov.br/nfce/qrcode"},
	"MG": {"https://portalsped.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml", "https://portalsped.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml"},
	"PA": {"https://appnfc.sefa.pa.gov.br/portal/view/consultas/nfce/nfceForm.seam", "https://appnfc.sefa.pa.gov.br/portal-homologacao/view/consultas/nfce/nfceForm.seam"},
	"PB": {"http://www.sefaz.pb.gov.br/nfce", "http://www.sefaz.pb.gov.br/nfcehom"},
	"PR": {"http://www.fazenda.pr.gov.br/nfce/qrcode/", "http://www.fazenda.pr.gov.br/nfce/qrcode/"},
	"PE": {"http://nfce.sefaz.pe.gov.br/nfce-web/consultarNFCe", "http://nfcehomolog.sefaz.pe.gov.br/nfce-web/consultarNFCe"},
	"PI": {"http://www.sefaz.pi.gov.br/nfce/qrcode", "http://www.sefaz.pi.gov.br/nfce/qrcode"},
	"RJ": {"https://consultadfe.fazenda.rj.gov.br/consultaNFCe/QRCode", "https://consultadfe.fazenda.rj.gov.br/consultaNFCe/QRCode"},
	"RN": {"https://nfce.sefaz.rn.gov.br/consultarNFCe.aspx", "https://hom.nfce.sefaz.rn.gov.br/consultarNFCe.aspx"},
	"RS": {"https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx", "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"},
	"RO": {"http://www.nfce.sefin.ro.gov.br/consultanfce/consulta.jsp", "http://www.nfce.sefin.ro.gov.br/consultanfce/consulta.jsp"},
	"RR": {"https://www.sefaz.rr.gov.br/nfce/servlet/qrcode", "http://200.174.88.103:8080/nfce/servlet/qrcode"},
	"SC": {"https://sat.sef.sc.gov.br/nfce/consulta?p=", "https://hom.sat.sef.sc.gov.br/nfce/consulta?p="},
	"SP": {"https://www.nfce.fazenda.sp.gov.br/qrcode", "https://www.homologacao.nfce.fazenda.sp.gov.br/qrcode"},
	"SE": {"http://www.nfce.se.gov.br/portal/consultarNFCe.jsp", "http://www.hom.nfe.se.gov.br/portal/consultarNFCe.jsp"},
	"TO": {"http://www.sefaz.to.gov.br/nfce/qrcode", "http://homologacao.sefaz.to.gov.br/nfce/qrcode"},
}

// urlsQRCodeModelo55 sao as DUAS UFs em que a URL do QR-Code do modelo 55
// difere da do 65. Nas outras 25 as duas tabelas do .ini coincidem, entao so
// a divergencia e declarada aqui -- um segundo mapa completo esconderia o
// fato de que a diferenca existe em dois lugares so.
var urlsQRCodeModelo55 = map[string]urlPorAmbiente{
	"BA": {"http://nfe.sefaz.ba.gov.br/servicos/nfce/modulos/geral/NFCEC_consulta_chave_acesso.aspx", "http://hnfe.sefaz.ba.gov.br/servicos/nfce/modulos/geral/NFCEC_consulta_chave_acesso.aspx"},
	"PR": {"http://www.fazenda.pr.gov.br/nfce/qrcode", "http://www.fazenda.pr.gov.br/nfce/qrcode"},
}

// urlsConsultaNFe e a URL do portal de consulta publica do modelo 55
// (chave URL-ConsultaNFe). Nao varia com a versao do QR-Code.
var urlsConsultaNFe = map[string]urlPorAmbiente{
	"AC": {"www.sefaznet.ac.gov.br/nfce/consulta", "www.sefaznet.ac.gov.br/nfce/consulta"},
	"AL": {"www.sefaz.al.gov.br/nfce/consulta", "www.sefaz.al.gov.br/nfce/consulta"},
	"AP": {"www.sefaz.ap.gov.br/nfce/consulta", "www.sefaz.ap.gov.br/nfce/consulta"},
	"AM": {"www.sefaz.am.gov.br/nfce/consulta", "www.sefaz.am.gov.br/nfce/consulta"},
	"BA": {"http://www.sefaz.ba.gov.br/nfce/consulta", "http://hinternet.sefaz.ba.gov.br/nfce/consulta"},
	"CE": {"www.sefaz.ce.gov.br/nfce/consulta", "www.sefaz.ce.gov.br/nfce/consulta"},
	"DF": {"www.fazenda.df.gov.br/nfce/consulta", "www.fazenda.df.gov.br/nfce/consulta"},
	"ES": {"www.sefaz.es.gov.br/nfce/consulta", "www.sefaz.es.gov.br/nfce/consulta"},
	"GO": {"http://www.sefaz.go.gov.br/nfce/consulta", "http://www.nfce.go.gov.br/post/ver/214413/consulta-nfc-e-homologacao"},
	"MA": {"www.sefaz.ma.gov.br/nfce/consulta", "www.sefaz.ma.gov.br/nfce/consulta"},
	"MT": {"http://www.sefaz.mt.gov.br/nfce/consultanfce", "http://homologacao.sefaz.mt.gov.br/nfce/consultanfce"},
	"MS": {"http://www.dfe.ms.gov.br/nfce/consulta", "http://www.dfe.ms.gov.br/nfce/consulta"},
	"MG": {"https://portalsped.fazenda.mg.gov.br/portalnfce", "https://hportalsped.fazenda.mg.gov.br/portalnfce"},
	"PA": {"www.sefa.pa.gov.br/nfce/consulta", "www.sefa.pa.gov.br/nfce/consulta"},
	"PB": {"www.sefaz.pb.gov.br/nfce/consulta", "www.receita.pb.gov.br/nfcehom"},
	"PR": {"http://www.fazenda.pr.gov.br/nfce/consulta", "http://www.fazenda.pr.gov.br/nfce/consulta"},
	"PE": {"nfce.sefaz.pe.gov.br/nfce/consulta", "nfce.sefaz.pe.gov.br/nfce/consulta"},
	"PI": {"www.sefaz.pi.gov.br/nfce/consulta", "www.sefaz.pi.gov.br/nfce/consulta"},
	"RJ": {"www.fazenda.rj.gov.br/nfce/consulta", "www.fazenda.rj.gov.br/nfce/consulta"},
	"RN": {"https://nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx", "https://hom.nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx"},
	"RS": {"www.sefaz.rs.gov.br/nfce/consulta", "www.sefaz.rs.gov.br/nfce/consulta"},
	"RO": {"www.sefin.ro.gov.br/nfce/consulta", "www.sefin.ro.gov.br/nfce/consulta"},
	"RR": {"www.sefaz.rr.gov.br/nfce/consulta", "www.sefaz.rr.gov.br/nfce/consulta"},
	"SC": {"https://sat.sef.sc.gov.br/nfce/consulta", "https://hom.sat.sef.sc.gov.br/nfce/consulta"},
	"SP": {"https://www.nfe.fazenda.sp.gov.br/consultanfe/consulta/publica/consultanfe.aspx", "https://www.homologacao.nfe.fazenda.sp.gov.br/ConsultaNFe/consulta/publica/ConsultarNFe.aspx"},
	"SE": {"http://www.nfce.se.gov.br/nfce/consulta", "http://www.hom.nfe.se.gov.br/nfce/consulta"},
	"TO": {"www.sefaz.to.gov.br/nfce/consulta", "http://homologacao.sefaz.to.gov.br/nfce/consulta.jsf"},
}

// urlsConsultaNFCe e a URL de consulta do modelo 65 para QR-Code versao 2 ou
// 3 (chave URL-ConsultaNFCe_2.00, com queda para a sem versao).
var urlsConsultaNFCe = map[string]urlPorAmbiente{
	"AC": {"www.sefaznet.ac.gov.br/nfce/consulta", "www.sefaznet.ac.gov.br/nfce/consulta"},
	"AL": {"www.sefaz.al.gov.br/nfce/consulta", "www.sefaz.al.gov.br/nfce/consulta"},
	"AP": {"www.sefaz.ap.gov.br/nfce/consulta", "www.sefaz.ap.gov.br/nfce/consulta"},
	"AM": {"www.sefaz.am.gov.br/nfce/consulta", "www.sefaz.am.gov.br/nfce/consulta"},
	"BA": {"http://www.sefaz.ba.gov.br/nfce/consulta", "http://hinternet.sefaz.ba.gov.br/nfce/consulta"},
	"CE": {"www.sefaz.ce.gov.br/nfce/consulta", "www.sefaz.ce.gov.br/nfce/consulta"},
	"DF": {"www.fazenda.df.gov.br/nfce/consulta", "www.fazenda.df.gov.br/nfce/consulta"},
	"ES": {"www.sefaz.es.gov.br/nfce/consulta", "www.sefaz.es.gov.br/nfce/consulta"},
	"GO": {"http://www.sefaz.go.gov.br/nfce/consulta", "http://www.nfce.go.gov.br/post/ver/214413/consulta-nfc-e-homologacao"},
	"MA": {"www.sefaz.ma.gov.br/nfce/consulta", "www.sefaz.ma.gov.br/nfce/consulta"},
	"MT": {"http://www.sefaz.mt.gov.br/nfce/consultanfce", "http://homologacao.sefaz.mt.gov.br/nfce/consultanfce"},
	"MS": {"http://www.dfe.ms.gov.br/nfce/consulta", "http://www.dfe.ms.gov.br/nfce/consulta"},
	"MG": {"https://portalsped.fazenda.mg.gov.br/portalnfce", "https://hportalsped.fazenda.mg.gov.br/portalnfce"},
	"PA": {"www.sefa.pa.gov.br/nfce/consulta", "www.sefa.pa.gov.br/nfce/consulta"},
	"PB": {"www.sefaz.pb.gov.br/nfce/consulta", "www.receita.pb.gov.br/nfcehom"},
	"PR": {"http://www.fazenda.pr.gov.br/nfce/consulta", "http://www.fazenda.pr.gov.br/nfce/consulta"},
	"PE": {"nfce.sefaz.pe.gov.br/nfce/consulta", "nfce.sefaz.pe.gov.br/nfce/consulta"},
	"PI": {"www.sefaz.pi.gov.br/nfce/consulta", "www.sefaz.pi.gov.br/nfce/consulta"},
	"RJ": {"www.fazenda.rj.gov.br/nfce/consulta", "www.fazenda.rj.gov.br/nfce/consulta"},
	"RN": {"https://nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx", "https://hom.nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx"},
	"RS": {"www.sefaz.rs.gov.br/nfce/consulta", "www.sefaz.rs.gov.br/nfce/consulta"},
	"RO": {"www.sefin.ro.gov.br/nfce/consulta", "www.sefin.ro.gov.br/nfce/consulta"},
	"RR": {"www.sefaz.rr.gov.br/nfce/consulta", "www.sefaz.rr.gov.br/nfce/consulta"},
	"SC": {"https://sat.sef.sc.gov.br/nfce/consulta", "https://hom.sat.sef.sc.gov.br/nfce/consulta"},
	"SP": {"https://www.nfce.fazenda.sp.gov.br/consulta", "https://www.homologacao.nfce.fazenda.sp.gov.br/consulta"},
	"SE": {"http://www.nfce.se.gov.br/nfce/consulta", "http://www.hom.nfe.se.gov.br/nfce/consulta"},
	"TO": {"www.sefaz.to.gov.br/nfce/consulta", "http://homologacao.sefaz.to.gov.br/nfce/consulta.jsf"},
}

// urlsConsultaNFCeQRCodeV1 e a URL de consulta do modelo 65 para QR-Code
// versao 1 (chave URL-ConsultaNFCe_1.00, com queda para a sem versao).
//
// Existe porque este e o UNICO servico do .ini cuja URL muda com a versao
// pedida: 44 das 54 combinacoes de UF e ambiente divergem da tabela acima.
// MG nao tem valor para a versao 1 e fica com string vazia, como no original.
var urlsConsultaNFCeQRCodeV1 = map[string]urlPorAmbiente{
	"AC": {"http://www.sefaznet.ac.gov.br/nfce/", "http://hml.sefaznet.ac.gov.br/nfce/"},
	"AL": {"http://nfce.sefaz.al.gov.br/consultaNFCe.htm", "http://nfce.sefaz.al.gov.br/consultaNFCe.htm"},
	"AP": {"https://www.sefaz.ap.gov.br/sate/seg/SEGf_AcessarFuncao.jsp?cdFuncao=FIS_1261", "https://www.sefaz.ap.gov.br/sate1/seg/SEGf_AcessarFuncao.jsp?cdFuncao=FIS_1261"},
	"AM": {"http://sistemas.sefaz.am.gov.br/nfceweb/formConsulta.do", "http://homnfce.sefaz.am.gov.br/nfceweb/formConsulta.do"},
	"BA": {"www.sefaz.ba.gov.br/nfce/consulta", "http://hinternet.sefaz.ba.gov.br/nfce/consulta"},
	"CE": {"http://nfce.sefaz.ce.gov.br/pages/consultaChaveAcesso.jsf", "http://nfceh.sefaz.ce.gov.br/pages/consultaChaveAcesso.jsf"},
	"DF": {"http://dec.fazenda.df.gov.br/nfce", "http://dec.fazenda.df.gov.br/nfce"},
	"ES": {"http://app.sefaz.es.gov.br/ConsultaNFCe", "http://app.sefaz.es.gov.br/ConsultaNFCe"},
	"GO": {"http://www.nfce.go.gov.br/post/ver/214344/consulta-nfce", "http://www.nfce.go.gov.br/post/ver/214413/consulta-nfc-e-homologacao"},
	"MA": {"http://www.nfce.sefaz.ma.gov.br/portal/consultaNFe.do", "http://www.hom.nfce.sefaz.ma.gov.br/portal/consultaNFe.do"},
	"MT": {"http://www.sefaz.mt.gov.br/nfce/consultanfce", "http://homologacao.sefaz.mt.gov.br/nfce/consultanfce"},
	"MS": {"http://www.dfe.ms.gov.br/nfce", "http://www.dfe.ms.gov.br/nfce"},
	"MG": {"", ""},
	"PA": {"https://appnfc.sefa.pa.gov.br/portal/view/consultas/nfce/consultanfce.seam", "https://appnfc.sefa.pa.gov.br/portal-homologacao/view/consultas/nfce/consultanfce.seam"},
	"PB": {"http://www.receita.pb.gov.br/nfce", "http://www.receita.pb.gov.br/nfcehom"},
	"PR": {"http://www.fazenda.pr.gov.br", "http://www.fazenda.pr.gov.br"},
	"PE": {"http://nfce.sefaz.pe.gov.br/nfce-web/consultarNFCe", "http://nfcehomolog.sefaz.pe.gov.br/nfce-web/consultarNFCe"},
	"PI": {"http://webas.sefaz.pi.gov.br/nfceweb/consultarNFCe.jsf", "http://webas.sefaz.pi.gov.br/nfceweb-homologacao/consultarNFCe.jsf"},
	"RJ": {"http://nfce.fazenda.rj.gov.br/consulta", "http://nfce.fazenda.rj.gov.br/consulta"},
	"RN": {"https://nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx", "https://hom.nfce.sefaz.rn.gov.br/portalDFE/NFCe/ConsultaNFCe.aspx"},
	"RS": {"https://www.sefaz.rs.gov.br/NFE/NFE-NFC.aspx", "https://www.sefaz.rs.gov.br/NFE/NFE-NFC.aspx"},
	"RO": {"http://www.nfce.sefin.ro.gov.br", "http://www.nfce.sefin.ro.gov.br/consultaAmbHomologacao.jsp"},
	"RR": {"https://www.sefaz.rr.gov.br/nfce/servlet/wp_consulta_nfce", "http://200.174.88.103:8080/nfce/servlet/wp_consulta_nfce"},
	"SC": {"https://sat.sef.sc.gov.br/nfce/consulta", "https://hom.sat.sef.sc.gov.br/nfce/consulta"},
	"SP": {"https://www.nfce.fazenda.sp.gov.br/consulta", "https://www.homologacao.nfce.fazenda.sp.gov.br/consulta"},
	"SE": {"http://www.nfce.se.gov.br/portal/consultarNFCe.jsp", "http://www.hom.nfe.se.gov.br/portal/consultarNFCe.jsp"},
	"TO": {"http://apps.sefaz.to.gov.br/portal-nfce/consultarNFCe.jsf", "http://apps.sefaz.to.gov.br/portal-nfce-homologacao/consultarNFCe.jsf"},
}
