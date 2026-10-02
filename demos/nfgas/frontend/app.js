// OpenFiscalBR - Demo NFGas (leitura e importacao em lote)
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

  // -------------------------------------------------------------------------
  // Catálogo de exemplos
  // -------------------------------------------------------------------------

  const seletor = document.getElementById("exemplo");
  const gerarExemplo = document.getElementById("gerarExemplo");
  const gerarComDados = document.getElementById("gerarComDados");
  const baixarExemplo = document.getElementById("baixarExemplo");
  const variarDados = document.getElementById("variarDados");
  const exemploDescricao = document.getElementById("exemploDescricao");
  const campoItens = document.getElementById("campoItens");

  const catalogo = {};

  const campos = {
    cnpj: document.getElementById("pCnpj"),
    uf: document.getElementById("pUf"),
    serie: document.getElementById("pSerie"),
    nnf: document.getElementById("pNnf"),
    tpamb: document.getElementById("pTpAmb"),
    itens: document.getElementById("pItens")
  };

  fetch("/api/exemplos")
    .then(function (r) { return r.json(); })
    .then(function (dados) {
      seletor.innerHTML = "";
      const grupos = {
        documento: document.createElement("optgroup"),
        evento: document.createElement("optgroup")
      };
      grupos.documento.label = "Documento";
      grupos.evento.label = "Evento";

      dados.exemplos.forEach(function (ex) {
        catalogo[ex.id] = ex;
        const op = document.createElement("option");
        op.value = ex.id;
        op.textContent = ex.nome;
        (grupos[ex.tipo] || grupos.documento).appendChild(op);
      });
      Object.keys(grupos).forEach(function (k) {
        if (grupos[k].childElementCount) seletor.appendChild(grupos[k]);
      });
      atualizarCamposDoExemplo();
      // Só agora há um id para gerar: antes disso o clique não faria nada,
      // e botão que não responde parece defeito.
      gerarExemplo.disabled = false;
      gerarComDados.disabled = false;
    })
    .catch(function () {
      seletor.innerHTML = "<option value=''>(catálogo indisponível)</option>";
      atualizarCamposDoExemplo();
    });

  seletor.addEventListener("change", atualizarCamposDoExemplo);

  function atualizarCamposDoExemplo() {
    const ex = catalogo[seletor.value];
    campoItens.hidden = !(ex && ex.aceitaItens);
    baixarExemplo.hidden = true;
    if (ex) {
      exemploDescricao.textContent = ex.descricao;
      exemploDescricao.hidden = false;
    } else {
      // Sem exemplo selecionado (catálogo indisponível), a descrição do
      // anterior ficaria na tela descrevendo coisa nenhuma.
      exemploDescricao.hidden = true;
    }
  }

  // Monta a query só com o que o usuário preencheu. Vazia = dados fictícios
  // fixos, que é o caminho do "clicar e gerar".
  function queryDosParametros(usarFormulario) {
    const q = new URLSearchParams();
    if (usarFormulario) {
      Object.keys(campos).forEach(function (nome) {
        const el = campos[nome];
        if (!el || el.closest("[hidden]")) return;
        const v = (el.value || "").trim();
        if (v) q.set(nome, v);
      });
    }
    if (variarDados.checked) q.set("variar", "1");
    const s = q.toString();
    return s ? "?" + s : "";
  }

  function carregarExemplo(usarFormulario) {
    const id = seletor.value;
    if (!id) return;

    erro.hidden = true;
    const query = queryDosParametros(usarFormulario);

    fetch("/api/exemplos/" + encodeURIComponent(id) + query)
      .then(function (resp) {
        return resp.json().then(function (dados) {
          return { ok: resp.ok, dados: dados };
        });
      })
      .then(function (r) {
        if (!r.ok) {
          mostrarErro(r.dados.erro || "Não foi possível gerar o exemplo.");
          return;
        }
        conteudo.value = r.dados.xml;

        let texto = r.dados.descricao;
        if (r.dados.chave) texto += " · chave " + r.dados.chave;
        if (r.dados.observacao) texto += " — " + r.dados.observacao;
        exemploDescricao.textContent = texto;
        exemploDescricao.hidden = false;

        baixarExemplo.href = "/api/exemplos/" + encodeURIComponent(id) + "/download" + query;
        baixarExemplo.hidden = false;

        // Seleciona a aba sugerida pelo mesmo caminho do clique humano, para
        // não duplicar o controle de estado da barra de abas.
        const alvo = document.querySelector('.aba[data-acao="' + r.dados.acaoSugerida + '"]');
        if (alvo) alvo.click();
      })
      .catch(function (e) {
        mostrarErro("Erro de rede: " + e.message);
      });
  }

  gerarExemplo.addEventListener("click", function () { carregarExemplo(false); });
  gerarComDados.addEventListener("click", function () { carregarExemplo(true); });

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
