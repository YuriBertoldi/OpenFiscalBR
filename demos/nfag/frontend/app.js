// OpenFiscalBR - Demo NFAg (leitura e importacao em lote)
// Derivado do Projeto ACBr (http://projetoacbr.com.br) - LGPL v2.1+

(function () {
  "use strict";

  let acao = "ler";

  const abas = document.querySelectorAll(".aba");
  const conteudo = document.getElementById("conteudo");
  const arquivo = document.getElementById("arquivo");
  const processar = document.getElementById("processar");
  const erro = document.getElementById("erro");
  const resumo = document.getElementById("resumo");
  const resultado = document.getElementById("resultado");

  abas.forEach(function (aba) {
    aba.addEventListener("click", function () {
      abas.forEach(function (a) { a.classList.remove("ativa"); });
      aba.classList.add("ativa");
      acao = aba.dataset.acao;
    });
  });

  arquivo.addEventListener("change", function () {
    const f = arquivo.files[0];
    if (!f) return;
    f.text().then(function (texto) { conteudo.value = texto; });
  });

  processar.addEventListener("click", function () {
    erro.hidden = true;
    resumo.hidden = true;

    const texto = conteudo.value.trim();
    if (!texto) {
      mostrarErro("Cole um documento ou abra um arquivo antes de processar.");
      return;
    }

    processar.disabled = true;
    resultado.textContent = "Processando…";

    fetch("/api/" + acao, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ conteudo: texto })
    })
      .then(function (resp) {
        return resp.json().then(function (dados) {
          return { ok: resp.ok, dados: dados };
        });
      })
      .then(function (r) {
        if (!r.ok) {
          mostrarErro(r.dados.erro || "Falha ao processar o documento.");
          resultado.textContent = "";
          return;
        }
        mostrarResumo(r.dados);
        resultado.textContent = JSON.stringify(r.dados, null, 2);
      })
      .catch(function (e) {
        mostrarErro("Erro de rede: " + e.message);
        resultado.textContent = "";
      })
      .finally(function () {
        processar.disabled = false;
      });
  });

  function mostrarErro(msg) {
    erro.textContent = msg;
    erro.hidden = false;
  }

  function mostrarResumo(dados) {
    let html = "";

    if (dados.resumo) {
      html = cartaoNota(dados.resumo);
    } else if (dados.resumos) {
      html = "<p class='destaque'>" + dados.lidas + " de " + dados.total +
        " documento(s) lido(s) com sucesso.</p>";
      dados.resumos.forEach(function (r) { html += cartaoNota(r); });
      (dados.falhas || []).forEach(function (f) {
        html += "<p class='falha'>" + escapar(f) + "</p>";
      });
    } else if (dados.rejeicoes !== undefined) {
      const okChave = dados.chaveValida ? "válida" : "INVÁLIDA — " + escapar(dados.erroChave || "");
      const okConcat = dados.concatConfere ? "confere" : "NÃO confere com os campos";
      html = "<p><strong>Chave:</strong> " + escapar(dados.chave) + "</p>" +
        "<p><strong>Dígito verificador:</strong> " + okChave + "</p>" +
        "<p><strong>Concatenação (regra 227):</strong> " + okConcat + "</p>";
      if (dados.rejeicoes.length === 0) {
        html += "<p class='destaque'>Nenhuma rejeição das regras de negócio.</p>";
      } else {
        dados.rejeicoes.forEach(function (r) {
          html += "<p class='falha'>" + escapar(r) + "</p>";
        });
      }
    } else if (dados.tipoEvento !== undefined) {
      html = "<p><strong>Evento:</strong> " + escapar(dados.tipoEvento) +
        " — " + escapar(dados.descricao) + "</p>" +
        "<p><strong>Chave:</strong> " + escapar(dados.chave) + "</p>" +
        "<p><strong>cStat:</strong> " + dados.cStat + " — " + escapar(dados.xMotivo) + "</p>";
      if (dados.justificativa) {
        html += "<p><strong>Justificativa:</strong> " + escapar(dados.justificativa) + "</p>";
      }
    }

    if (html) {
      resumo.innerHTML = html;
      resumo.hidden = false;
    }
  }

  function cartaoNota(r) {
    return "<div class='cartao'>" +
      "<p><strong>NF " + r.nNF + "</strong> · série " + r.serie +
      " · <span class='situacao " + classeSituacao(r.situacao) + "'>" +
      escapar(r.situacao) + "</span></p>" +
      "<p class='chave'>" + escapar(r.chave) + "</p>" +
      "<p>" + escapar(r.emitente) + " (" + escapar(r.cnpj) + ")</p>" +
      "<p>" + r.itens + " item(ns) · vNF R$ " + Number(r.vNF).toFixed(2).replace(".", ",") +
      (r.dhEmi ? " · emitida em " + escapar(r.dhEmi) : "") + "</p>" +
      "</div>";
  }

  function classeSituacao(s) {
    if (s === "cancelada") return "cancelada";
    if (s === "confirmada" || s === "processada") return "confirmada";
    return "neutra";
  }

  function escapar(s) {
    const div = document.createElement("div");
    div.textContent = s == null ? "" : String(s);
    return div.innerHTML;
  }
})();
