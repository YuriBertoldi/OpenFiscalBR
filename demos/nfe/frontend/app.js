// OpenFiscalBR - Demo NF-e
// Derivado do Projeto ACBr (http://projetoacbr.com.br), LGPL 2.1+.
//
// Sem framework e sem CDN: o projeto inteiro esta em stdlib pura, e a demo
// nao e lugar para mudar isso.

'use strict';

// acoes mapeia cada aba para o metodo e o caminho que ela chama.
//
// A tabela existe para que acrescentar uma rota seja uma linha aqui e um
// botao no HTML -- e nao um `if` novo no meio do envio.
const acoes = {
  'ler':              { metodo: 'POST', caminho: '/api/ler' },
  'ler-lote':         { metodo: 'POST', caminho: '/api/ler-lote' },
  'validar':          { metodo: 'POST', caminho: '/api/validar' },
  'ini-para-xml':     { metodo: 'POST', caminho: '/api/ini-para-xml' },
  'xml-para-ini':     { metodo: 'POST', caminho: '/api/xml-para-ini' },
  'ler-evento':       { metodo: 'POST', caminho: '/api/ler-evento' },
  'ler-distribuicao': { metodo: 'POST', caminho: '/api/ler-distribuicao' },
  'gerar':            { metodo: 'POST', caminho: '/api/gerar' },
  'assinar':          { metodo: 'POST', caminho: '/api/assinar' },
  'transmitir':       { metodo: 'POST', caminho: '/api/transmitir' },
};

let acaoAtual = 'ler';

const elConteudo    = document.getElementById('conteudo');
const elResposta    = document.getElementById('resposta');
const elStatusHttp  = document.getElementById('statusHttp');
const elTempo       = document.getElementById('tempo');
const elAcaoAtual   = document.getElementById('acaoAtual');
const elExemplo     = document.getElementById('exemplo');
const elDescricao   = document.getElementById('exemploDescricao');
const elCarregar    = document.getElementById('carregarExemplo');

// ---------------------------------------------------------------------------
// Abas
// ---------------------------------------------------------------------------

document.querySelectorAll('.aba').forEach((aba) => {
  aba.addEventListener('click', () => {
    document.querySelectorAll('.aba').forEach((a) => a.classList.remove('ativa'));
    aba.classList.add('ativa');
    acaoAtual = aba.dataset.acao;
    const a = acoes[acaoAtual];
    elAcaoAtual.textContent = `${a.metodo} ${a.caminho}`;
  });
});

// ---------------------------------------------------------------------------
// Execucao
// ---------------------------------------------------------------------------

document.getElementById('enviar').addEventListener('click', () => {
  const conteudo = elConteudo.value.trim();
  if (!conteudo) {
    mostrarErro('Cole um documento ou carregue um exemplo antes de executar.');
    return;
  }
  const a = acoes[acaoAtual];
  chamar(a.metodo, a.caminho, { conteudo });
});

document.getElementById('limpar').addEventListener('click', () => {
  elConteudo.value = '';
  elResposta.textContent = 'Nenhuma requisição ainda.';
  elStatusHttp.textContent = '';
  elTempo.textContent = '';
  elDescricao.hidden = true;
});

// Botoes que chamam uma rota fixa, sem corpo.
document.querySelectorAll('[data-direto]').forEach((botao) => {
  botao.addEventListener('click', () => {
    const [metodo, caminho] = botao.dataset.direto.split(' ');
    chamar(metodo, caminho, null);
  });
});

document.getElementById('consultar').addEventListener('click', () => {
  const chave = document.getElementById('chaveConsulta').value.trim();
  if (chave.length !== 44) {
    mostrarErro('A chave de acesso tem 44 dígitos.');
    return;
  }
  chamar('GET', `/api/consultar/${encodeURIComponent(chave)}`, null);
});

// chamar executa a requisicao e faz o pretty-print da resposta.
//
// Mede o tempo porque as rotas que falam com a SEFAZ demoram em ordem de
// grandeza diferente das locais, e ver isso ajuda a entender o que e lento.
async function chamar(metodo, caminho, corpo) {
  elResposta.textContent = 'Executando…';
  elStatusHttp.textContent = '';
  elTempo.textContent = '';

  const inicio = performance.now();
  try {
    const opcoes = { method: metodo };
    if (corpo !== null) {
      opcoes.headers = { 'Content-Type': 'application/json' };
      opcoes.body = JSON.stringify(corpo);
    }

    const resposta = await fetch(caminho, opcoes);
    const texto = await resposta.text();

    elStatusHttp.textContent = `HTTP ${resposta.status}`;
    elStatusHttp.className = `pill ${resposta.ok ? 'ok' : 'erro'}`;
    elTempo.textContent = `${Math.round(performance.now() - inicio)} ms`;

    // A API sempre devolve JSON, inclusive no erro. Se nao for JSON, o
    // texto cru e mais util que uma mensagem generica.
    try {
      elResposta.textContent = JSON.stringify(JSON.parse(texto), null, 2);
    } catch {
      elResposta.textContent = texto;
    }
  } catch (e) {
    elStatusHttp.textContent = 'falha';
    elStatusHttp.className = 'pill erro';
    elResposta.textContent = `Não foi possível chamar a API: ${e.message}`;
  }
}

function mostrarErro(mensagem) {
  elStatusHttp.textContent = '';
  elTempo.textContent = '';
  elResposta.textContent = mensagem;
}

// ---------------------------------------------------------------------------
// Exemplos
// ---------------------------------------------------------------------------

let catalogo = [];

async function carregarCatalogo() {
  try {
    const resposta = await fetch('/api/exemplos');
    const dados = await resposta.json();
    catalogo = dados.exemplos || [];

    elExemplo.innerHTML = '';
    if (catalogo.length === 0) {
      elExemplo.innerHTML = '<option value="">Nenhum exemplo</option>';
      return;
    }
    for (const ex of catalogo) {
      const opcao = document.createElement('option');
      opcao.value = ex.id;
      opcao.textContent = ex.nome;
      elExemplo.appendChild(opcao);
    }
    elCarregar.disabled = false;
    descrever();
  } catch {
    elExemplo.innerHTML = '<option value="">Falha ao carregar</option>';
  }
}

function descrever() {
  const ex = catalogo.find((e) => e.id === elExemplo.value);
  if (!ex) {
    elDescricao.hidden = true;
    return;
  }
  elDescricao.textContent = ex.descricao;
  elDescricao.hidden = false;
}

elExemplo.addEventListener('change', descrever);

elCarregar.addEventListener('click', async () => {
  const id = elExemplo.value;
  if (!id) return;

  elCarregar.disabled = true;
  try {
    const resposta = await fetch(`/api/exemplos/${encodeURIComponent(id)}`);
    const dados = await resposta.json();
    if (!resposta.ok) {
      mostrarErro(dados.erro || 'Falha ao carregar o exemplo.');
      return;
    }
    elConteudo.value = dados.conteudo;

    // Leva o usuario direto para a aba que faz sentido com o exemplo --
    // um .ini na aba "Ler XML" daria erro e pareceria defeito da API.
    const aba = document.querySelector(`.aba[data-acao="${dados.acaoSugerida}"]`);
    if (aba) aba.click();
  } finally {
    elCarregar.disabled = false;
  }
});

carregarCatalogo();
