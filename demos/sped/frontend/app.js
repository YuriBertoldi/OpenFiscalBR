document.addEventListener("DOMContentLoaded", function () {
  var form = document.getElementById("spedForm");
  var resultado = document.getElementById("resultado");
  var mensagem = document.getElementById("mensagem");
  var conteudo = document.getElementById("conteudo");
  var nomeArquivo = document.getElementById("nomeArquivo");
  var btnGerar = document.getElementById("btnGerar");
  var btnDownload = document.getElementById("btnDownload");

  // Set default dates (first and last day of current month)
  var now = new Date();
  var dtIni = document.getElementById("dt_ini");
  var dtFin = document.getElementById("dt_fin");
  var year = now.getFullYear();
  var month = String(now.getMonth() + 1).padStart(2, "0");
  dtIni.value = year + "-" + month + "-01";
  var lastDay = new Date(year, now.getMonth() + 1, 0).getDate();
  dtFin.value = year + "-" + month + "-" + String(lastDay).padStart(2, "0");

  var lastContent = "";
  var lastFilename = "";

  form.addEventListener("submit", function (e) {
    e.preventDefault();

    var payload = {
      cod_ver: parseInt(document.getElementById("cod_ver").value, 10),
      cod_fin: parseInt(document.getElementById("cod_fin").value, 10),
      dt_ini: document.getElementById("dt_ini").value,
      dt_fin: document.getElementById("dt_fin").value,
      nome: document.getElementById("nome").value,
      cnpj: document.getElementById("cnpj").value,
      cpf: document.getElementById("cpf").value,
      uf: document.getElementById("uf").value,
      ie: document.getElementById("ie").value,
      cod_mun: parseInt(document.getElementById("cod_mun").value, 10) || 0,
      im: document.getElementById("im").value,
      suframa: document.getElementById("suframa").value,
      ind_perfil: parseInt(document.getElementById("ind_perfil").value, 10),
      ind_ativ: parseInt(document.getElementById("ind_ativ").value, 10),
      fantasia: document.getElementById("fantasia").value,
      cep: document.getElementById("cep").value,
      endereco: document.getElementById("endereco").value,
      num: document.getElementById("num").value,
      compl: document.getElementById("compl").value,
      bairro: document.getElementById("bairro").value,
      fone: document.getElementById("fone").value,
      email: document.getElementById("email").value,
    };

    btnGerar.disabled = true;
    btnGerar.textContent = "Gerando...";

    fetch("/api/gerar", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then(function (res) {
        return res.json();
      })
      .then(function (data) {
        resultado.classList.remove("hidden");

        if (data.sucesso) {
          mensagem.className = "sucesso";
          mensagem.textContent = data.mensagem;
          conteudo.textContent = data.conteudo;
          nomeArquivo.textContent = data.arquivo;
          lastContent = data.conteudo;
          lastFilename = data.arquivo;
          document.getElementById("conteudoWrapper").style.display = "block";
        } else {
          mensagem.className = "erro";
          mensagem.textContent = data.mensagem;
          document.getElementById("conteudoWrapper").style.display = "none";
        }
      })
      .catch(function (err) {
        resultado.classList.remove("hidden");
        mensagem.className = "erro";
        mensagem.textContent = "Erro de conexao: " + err.message;
        document.getElementById("conteudoWrapper").style.display = "none";
      })
      .finally(function () {
        btnGerar.disabled = false;
        btnGerar.textContent = "Gerar Arquivo SPED";
      });
  });

  btnDownload.addEventListener("click", function () {
    if (!lastContent) return;
    var blob = new Blob([lastContent], { type: "text/plain;charset=utf-8" });
    var url = URL.createObjectURL(blob);
    var a = document.createElement("a");
    a.href = url;
    a.download = lastFilename || "sped.txt";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  });
});
