// Linha do Tempo da Ocorrência (caixa-preta) — Centro Operacional
var CoTimeline = (function () {
  var STORAGE_KEY = "co_timeline_caixa_preta";
  var MAX_ITENS = 400;
  var porMapa = {};
  var syncChaves = {};
  var ctx = { idCliente: "", idFranqueado: "", idOperador: "" };
  var pendentesSalvar = {};
  var saveTimer = null;
  var servidorOk = true;

  function nomeOperador() {
    return (
      sessionStorage.getItem("login_userNick") ||
      sessionStorage.getItem("login_userNome") ||
      "Operador"
    );
  }

  function lerStore() {
    try {
      var raw = sessionStorage.getItem(STORAGE_KEY);
      var obj = raw ? JSON.parse(raw) : {};
      return obj && typeof obj === "object" ? obj : {};
    } catch (e) {
      return {};
    }
  }

  function gravarStore() {
    try {
      var out = {};
      Object.keys(porMapa).forEach(function (mid) {
        out[mid] = (porMapa[mid] || []).slice(-MAX_ITENS);
      });
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(out));
    } catch (e) {}
  }

  function carregarStore() {
    var store = lerStore();
    Object.keys(store).forEach(function (mid) {
      porMapa[mid] = Array.isArray(store[mid]) ? store[mid] : [];
    });
  }

  function listaMapa(mapaId) {
    mapaId = String(parseInt(mapaId, 10) || "");
    if (!mapaId) return [];
    if (!porMapa[mapaId]) porMapa[mapaId] = [];
    return porMapa[mapaId];
  }

  function parseHora(h) {
    if (!h) return Date.now();
    if (typeof h === "number") return h;
    var s = String(h).trim();
    // 02/01/2006 15:04:05
    var m = s.match(/^(\d{2})\/(\d{2})\/(\d{4})\s+(\d{2}):(\d{2}):(\d{2})/);
    if (m) {
      return new Date(+m[3], +m[2] - 1, +m[1], +m[4], +m[5], +m[6]).getTime();
    }
    // 15:04:05
    var t = s.match(/^(\d{2}):(\d{2}):(\d{2})$/);
    if (t) {
      var d = new Date();
      d.setHours(+t[1], +t[2], +t[3], 0);
      return d.getTime();
    }
    var n = Date.parse(s);
    return isNaN(n) ? Date.now() : n;
  }

  function formatHora(ts) {
    var d = new Date(ts);
    function z(n) {
      return n < 10 ? "0" + n : String(n);
    }
    return z(d.getHours()) + ":" + z(d.getMinutes()) + ":" + z(d.getSeconds());
  }

  function formatDataHora(ts) {
    var d = new Date(ts);
    function z(n) {
      return n < 10 ? "0" + n : String(n);
    }
    return (
      z(d.getDate()) +
      "/" +
      z(d.getMonth() + 1) +
      "/" +
      d.getFullYear() +
      " " +
      formatHora(ts)
    );
  }

  function normalizarGrupo(g) {
    return String(g || "")
      .toUpperCase()
      .replace(/Ã/g, "A")
      .replace(/É/g, "E")
      .replace(/Ê/g, "E")
      .replace(/Í/g, "I")
      .replace(/Ó/g, "O");
  }

  function humanizarEventoContact(ev, ctx) {
    var grupo = normalizarGrupo(ev.grupo);
    var desc = String(ev.descricao || "").trim();
    var zonaNome = String(ev.descZona || "").trim() || "Zona " + String(ev.zonaUser || "").trim();
    var zonaNum = String(ev.zonaUser || "").replace(/^0+/, "") || String(ev.zonaUser || "");
    var qtd = parseInt(ev.quantidade, 10) || 1;

    if (grupo.indexOf("ARME") >= 0 && grupo.indexOf("DESARME") < 0) {
      if (/noite|noturno/i.test(desc)) {
        return "Central Armada em Modo Noturno.";
      }
      return desc ? "Central Armada — " + desc + "." : "Central Armada.";
    }
    if (grupo.indexOf("DESARME") >= 0) {
      return desc ? "Central Desarmada — " + desc + "." : "Central Desarmada.";
    }
    if (grupo.indexOf("ALARME") >= 0 || grupo.indexOf("PANICO") >= 0 || grupo.indexOf("EMERGENCIA") >= 0) {
      var sufixo = "Aguardando verificação.";
      if (ctx && ctx.deslocamento) {
        sufixo = "Invasor em deslocamento" + (ctx.direcao ? " na direção " + ctx.direcao : "") + ".";
      } else if (qtd > 1) {
        sufixo = "Repetição do disparo (" + qtd + "x).";
      }
      return (
        "Disparo de Alarme: " +
        zonaNome +
        (zonaNum ? " (Zona " + zonaNum + ")" : "") +
        " — " +
        sufixo
      );
    }
    if (grupo.indexOf("FALHA") >= 0 || grupo.indexOf("TROUBLE") >= 0) {
      return (
        "Falha reportada: " +
        (desc || zonaNome) +
        (zonaNum ? " (Zona " + zonaNum + ")" : "") +
        "."
      );
    }
    if (grupo.indexOf("RESTAURA") >= 0 || /^R/i.test(String(ev.codigoFull || ""))) {
      return "Restauro: " + (desc || zonaNome) + ".";
    }
    if (desc) {
      return desc.replace(/\.*$/, "") + ".";
    }
    return "Evento Contact ID " + (ev.codigoFull || ev.codigo || "—") + ".";
  }

  function tipoDoEvento(ev) {
    var g = normalizarGrupo(ev.grupo);
    if (g.indexOf("ALARME") >= 0 || g.indexOf("PANICO") >= 0 || g.indexOf("EMERGENCIA") >= 0) return "alarme";
    if (g.indexOf("ARME") >= 0 || g.indexOf("DESARME") >= 0 || g.indexOf("CONTROLE") >= 0) return "controle";
    if (g.indexOf("FALHA") >= 0) return "falha";
    return "info";
  }

  function configurarContexto(opts) {
    opts = opts || {};
    ctx.idCliente = String(opts.idCliente || "").trim();
    ctx.idFranqueado = String(opts.idFranqueado || "").trim();
    ctx.idOperador = String(opts.idOperador || "").trim();
  }

  function tsISO(ts) {
    try {
      return new Date(ts || Date.now()).toISOString();
    } catch (e) {
      return new Date().toISOString();
    }
  }

  function parseTsServidor(v) {
    if (!v) return Date.now();
    if (typeof v === "number") return v;
    var n = Date.parse(String(v));
    return isNaN(n) ? parseHora(v) : n;
  }

  function itemParaApi(item, mapaId) {
    mapaId = parseInt(mapaId, 10);
    return {
      chave: item.chave,
      mapa_ambiente_id: mapaId,
      idCliente: ctx.idCliente,
      idFranqueado: ctx.idFranqueado,
      idProcesso: item.idProcesso || "",
      idSetor: item.idSetor || "",
      idDispositivo: item.idDispositivo || "",
      idOperador: ctx.idOperador,
      nomeOperador: item.operador || nomeOperador(),
      tipo: item.tipo || "info",
      origem: item.origem || "operador",
      codigo: item.codigo || "",
      zonaUser: item.zonaUser || "",
      particao: item.particao || "",
      texto: item.texto || "",
      evento_ts: tsISO(item.ts),
    };
  }

  function agendarSalvar(mapaId) {
    if (!servidorOk) return;
    mapaId = String(parseInt(mapaId, 10) || "");
    if (!mapaId) return;
    pendentesSalvar[mapaId] = true;
    clearTimeout(saveTimer);
    saveTimer = setTimeout(flushSalvar, 700);
  }

  function flushSalvar() {
    var ids = Object.keys(pendentesSalvar);
    if (!ids.length) return;
    pendentesSalvar = {};
    ids.forEach(function (mapaId) {
      var lista = listaMapa(mapaId);
      if (!lista.length) return;
      var eventos = lista.map(function (it) {
        return itemParaApi(it, mapaId);
      });
      $.ajax({
        url: "/centroOperacional/timeline/salvar",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ eventos: eventos }),
      }).fail(function () {
        servidorOk = false;
      });
    });
  }

  function mergeServidor(mapaId, rows) {
    mapaId = String(parseInt(mapaId, 10) || "");
    if (!mapaId || !rows || !rows.length) return;
    rows.forEach(function (row) {
      var chave = row.chave || row.Chave;
      if (!chave) return;
      syncChaves[chave] = true;
      var ts = parseTsServidor(row.evento_ts || row.EventoTs);
      upsertItemLocal(mapaId, {
        id: chave,
        chave: chave,
        ts: ts,
        hora: formatHora(ts),
        dataHora: formatDataHora(ts),
        texto: row.texto || row.Texto || "",
        tipo: row.tipo || row.Tipo || "info",
        origem: row.origem || row.Origem || "servidor",
        idProcesso: row.idProcesso || row.IdProcesso || "",
        idSetor: row.idSetor || row.IdSetor || "",
        codigo: row.codigo || row.Codigo || "",
        zonaUser: row.zonaUser || row.ZonaUser || "",
        operador: row.nomeOperador || row.NomeOperador || "",
      });
    });
  }

  function upsertItemLocal(mapaId, item) {
    var lista = listaMapa(mapaId);
    var chave = item.chave || item.id;
    if (!chave) return null;
    for (var i = 0; i < lista.length; i++) {
      if (lista[i].chave === chave) {
        lista[i] = Object.assign({}, lista[i], item);
        gravarStore();
        return lista[i];
      }
    }
    lista.push(item);
    lista.sort(function (a, b) {
      return a.ts - b.ts;
    });
    if (lista.length > MAX_ITENS) {
      porMapa[String(parseInt(mapaId, 10))] = lista.slice(-MAX_ITENS);
    }
    gravarStore();
    return item;
  }

  function carregarServidor(mapaId, cb) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId) {
      if (cb) cb();
      return;
    }
    $.ajax({
      url: "/centroOperacional/timeline/listar",
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify({
        mapa_ambiente_id: mapaId,
        idCliente: ctx.idCliente || undefined,
        limite: MAX_ITENS,
      }),
    })
      .done(function (r) {
        var lista = r.dados || r.items || (Array.isArray(r) ? r : []);
        if (r.status !== "Vazio" && lista && lista.length) {
          mergeServidor(mapaId, lista);
        }
        if (cb) cb();
      })
      .fail(function () {
        if (cb) cb();
      });
  }

  function upsertItem(mapaId, item) {
    var salvo = upsertItemLocal(mapaId, item);
    agendarSalvar(mapaId);
    return salvo;
  }

  function adicionarAcaoOperador(opts) {
    var mapaId = parseInt(opts.mapaId, 10);
    if (!mapaId) return null;
    var ts = opts.ts || Date.now();
    var op = opts.operador || nomeOperador();
    var texto = opts.texto || "";
    var tipo = opts.tipo || "operador";
    var chave =
      opts.chave ||
      "op:" + tipo + ":" + mapaId + ":" + ts + ":" + String(texto).slice(0, 40);

    return upsertItem(mapaId, {
      id: chave,
      chave: chave,
      ts: ts,
      hora: formatHora(ts),
      dataHora: formatDataHora(ts),
      texto: texto,
      tipo: tipo,
      origem: "operador",
      idProcesso: opts.idProcesso || "",
      operador: op,
    });
  }

  function registrarCamera(opts) {
    var op = nomeOperador();
    var cam = String(opts.setorNome || opts.label || opts.nomeSetor || "câmera").trim();
    return adicionarAcaoOperador({
      mapaId: opts.mapaId,
      idProcesso: opts.idProcesso,
      tipo: "camera",
      chave: "cam:" + (opts.idProcesso || "") + ":" + (opts.idSetor || cam) + ":" + Math.floor(Date.now() / 5000),
      texto: 'Operador ' + op + ' abriu a câmera "' + cam + '".',
    });
  }

  function registrarFinalizacao(opts) {
    var desc = String(opts.descricao || "").trim();
    var texto = "Ocorrência finalizada";
    if (desc) texto += " (" + desc + ")";
    else texto += " (atendimento concluído)";
    texto += ".";
    return adicionarAcaoOperador({
      mapaId: opts.mapaId,
      idProcesso: opts.idProcesso,
      tipo: "finalizado",
      chave: "fin:" + (opts.idProcesso || "") + ":" + Date.now(),
      texto: texto,
    });
  }

  function registrarArme(opts) {
    var op = nomeOperador();
    var armado = !!opts.armado;
    return adicionarAcaoOperador({
      mapaId: opts.mapaId,
      tipo: "controle",
      chave: "arme:" + opts.mapaId + ":" + Math.floor(Date.now() / 3000),
      texto: armado
        ? "Operador " + op + " armou a central pelo Centro Operacional."
        : "Operador " + op + " desarmou a central pelo Centro Operacional.",
    });
  }

  function registrarRastroDisparo(opts) {
    opts = opts || {};
    var mapaId = parseInt(opts.mapaId, 10);
    if (!mapaId) return null;
    var zona = String(opts.label || opts.idSetor || "zona").trim();
    var ts = opts.ts || Date.now();
    var chave =
      opts.chave ||
      "rastro:disp:" +
        mapaId +
        ":" +
        String(opts.idSetor || zona) +
        ":" +
        Math.floor(ts / 5000);

    return upsertItem(mapaId, {
      id: chave,
      chave: chave,
      ts: ts,
      hora: formatHora(ts),
      dataHora: formatDataHora(ts),
      texto: 'Rastro: disparo registrado em "' + zona + '".',
      tipo: "rastro",
      origem: "rastro",
      idProcesso: opts.idProcesso || "",
      idSetor: opts.idSetor || "",
      zonaUser: opts.zonaUser || "",
    });
  }

  function registrarRastroAlerta(opts) {
    opts = opts || {};
    var mapaId = parseInt(opts.mapaId, 10);
    var mensagem = String(opts.mensagem || "").trim();
    if (!mapaId || !mensagem) return null;
    var ts = opts.ts || Date.now();
    var zonaProvavel =
      (opts.setor && (opts.setor.label || opts.setor.idSetor)) ||
      opts.proximaZona ||
      "";
    var chave =
      opts.chave ||
      "rastro:alerta:" +
        mapaId +
        ":" +
        String(opts.direcao || "") +
        ":" +
        String(zonaProvavel) +
        ":" +
        Math.floor(ts / 8000);

    return upsertItem(mapaId, {
      id: chave,
      chave: chave,
      ts: ts,
      hora: formatHora(ts),
      dataHora: formatDataHora(ts),
      texto: "Rastreamento: " + mensagem.replace(/\.*$/, "") + ".",
      tipo: "rastro",
      origem: "rastro",
      idProcesso: opts.idProcesso || "",
      direcao: opts.direcao || "",
      idSetor: (opts.setor && opts.setor.idSetor) || opts.idSetor || "",
    });
  }

  function sincronizarEventosProcesso(mapaId, idProcesso, eventos, ctx) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId || !eventos || !eventos.length) return;

    var alarmeCount = 0;
    eventos.forEach(function (ev) {
      var g = normalizarGrupo(ev.grupo);
      var isAlarme = g.indexOf("ALARME") >= 0 || g.indexOf("PANICO") >= 0 || g.indexOf("EMERGENCIA") >= 0;
      if (isAlarme) alarmeCount++;
      var desloc = isAlarme && alarmeCount >= 2;
      var chave =
        "cid:" +
        (idProcesso || "") +
        ":" +
        String(ev.codigoFull || ev.codigo || "") +
        ":" +
        String(ev.zonaUser || "") +
        ":" +
        String(ev.particao || "") +
        ":" +
        String(ev.hora || "");

      if (syncChaves[chave]) {
        return;
      }
      syncChaves[chave] = true;

      var ts = parseHora(ev.hora);
      upsertItem(mapaId, {
        id: chave,
        chave: chave,
        ts: ts,
        hora: formatHora(ts),
        dataHora: formatDataHora(ts),
        texto: humanizarEventoContact(ev, {
          deslocamento: desloc || (ctx && ctx.deslocamento),
          direcao: (ctx && ctx.direcao) || "",
        }),
        tipo: tipoDoEvento(ev),
        origem: "contact_id",
        idProcesso: idProcesso || "",
        codigo: ev.codigoFull || ev.codigo || "",
        zonaUser: ev.zonaUser || "",
        idSetor: ev.idSetor || "",
        quantidade: ev.quantidade || 1,
      });
    });
  }

  function sincronizarProcessoResumo(mapaId, proc, ctx) {
    if (!proc || !proc.idProcesso) return;
    var chaveBase = "proc:" + proc.idProcesso + ":" + (proc.quantidade || 0) + ":" + (proc.idSetor || "") + ":" + (proc.dataHora || "");
    if (syncChaves[chaveBase]) return;
    syncChaves[chaveBase] = true;

    var fakeEv = {
      grupo: proc.grupo,
      descricao: proc.descricaoGrupo || proc.codigo,
      descZona: proc.nomeSetor,
      zonaUser: proc.zonaUser,
      codigo: proc.codigo,
      codigoFull: proc.codigo,
      quantidade: proc.quantidade,
      hora: proc.dataHora || proc.dataCriacao,
    };
    var ts = parseHora(fakeEv.hora);
    upsertItem(mapaId, {
      id: chaveBase,
      chave: chaveBase,
      ts: ts,
      hora: formatHora(ts),
      dataHora: formatDataHora(ts),
      texto: humanizarEventoContact(fakeEv, ctx || {}),
      tipo: tipoDoEvento(fakeEv),
      origem: "processo",
      idProcesso: proc.idProcesso,
      codigo: proc.codigo || "",
      zonaUser: proc.zonaUser || "",
      idSetor: proc.idSetor || "",
    });
  }

  function obter(mapaId) {
    return listaMapa(mapaId).slice().sort(function (a, b) {
      return a.ts - b.ts;
    });
  }

  function limparMapa(mapaId) {
    mapaId = String(parseInt(mapaId, 10) || "");
    if (!mapaId) return;
    porMapa[mapaId] = [];
    Object.keys(syncChaves).forEach(function (k) {
      if (k.indexOf(":" + mapaId + ":") >= 0) delete syncChaves[k];
    });
    gravarStore();
  }

  function exportarTexto(mapaId) {
    var itens = obter(mapaId);
    if (!itens.length) return "Sem eventos na linha do tempo.";
    return itens
      .map(function (it) {
        return it.dataHora + " — " + it.texto;
      })
      .join("\n");
  }

  carregarStore();

  return {
    configurarContexto: configurarContexto,
    carregarServidor: carregarServidor,
    obter: obter,
    sincronizarEventosProcesso: sincronizarEventosProcesso,
    sincronizarProcessoResumo: sincronizarProcessoResumo,
    registrarCamera: registrarCamera,
    registrarFinalizacao: registrarFinalizacao,
    registrarArme: registrarArme,
    registrarRastroDisparo: registrarRastroDisparo,
    registrarRastroAlerta: registrarRastroAlerta,
    adicionarAcaoOperador: adicionarAcaoOperador,
    limparMapa: limparMapa,
    exportarTexto: exportarTexto,
    formatHora: formatHora,
    humanizarEventoContact: humanizarEventoContact,
  };
})();
