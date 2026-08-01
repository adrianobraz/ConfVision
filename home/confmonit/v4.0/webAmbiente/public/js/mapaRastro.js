// Rastro de disparos / trajetória preditiva no mapa (Centro Operacional)
var MapaRastro = (function () {
  var STORAGE_KEY = "co_rastro_disparos";
  var MAX_SESSAO = 80;
  var MAX_JANELA_MS = 15 * 60 * 1000;
  var PRED_DIST_MAX = 28;

  var estadoPorMapa = {};
  var modoAnalise = false;
  var analiseTimer = null;
  var analiseIdx = 0;
  var hooks = {
    onAlerta: null,
    onChange: null,
  };
  var ctx = { idCliente: "", idFranqueado: "", idOperador: "" };
  var pendentesSalvar = {};
  var saveTimer = null;
  var servidorOk = true;

  function agora() {
    return Date.now();
  }

  function lerSessao() {
    try {
      var raw = sessionStorage.getItem(STORAGE_KEY);
      var arr = raw ? JSON.parse(raw) : [];
      return Array.isArray(arr) ? arr : [];
    } catch (e) {
      return [];
    }
  }

  function gravarSessao(lista) {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(lista.slice(-MAX_SESSAO)));
    } catch (e) {}
  }

  function estadoMapa(mapaId) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId) return null;
    if (!estadoPorMapa[mapaId]) {
      estadoPorMapa[mapaId] = {
        disparos: [],
        preditoId: null,
        direcao: "",
        processoId: null,
        ultimoAlerta: "",
      };
    }
    return estadoPorMapa[mapaId];
  }

  function resolverSetor(mapaId, idSetor, setores) {
    idSetor = String(idSetor || "").trim();
    if (!idSetor) return null;
    var lista = setores || [];
    for (var i = 0; i < lista.length; i++) {
      var p = typeof mapaNormalizarSetor === "function" ? mapaNormalizarSetor(lista[i]) : lista[i];
      if (String(p.idSetor || "").trim() === idSetor) {
        return {
          idSetor: p.idSetor,
          label: p.label || p.idSetor,
          posX: Number(p.posX) || 0,
          posY: Number(p.posY) || 0,
        };
      }
    }
    return null;
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
    return isNaN(n) ? Date.now() : n;
  }

  function chaveDisparo(mapaId, item) {
    return (
      "rastro:" +
      mapaId +
      ":" +
      (item.idProcesso || "") +
      ":" +
      (item.idSetor || "") +
      ":" +
      (item.ts || "")
    );
  }

  function agendarSalvar(mapaId) {
    if (!servidorOk) return;
    mapaId = parseInt(mapaId, 10);
    if (!mapaId) return;
    pendentesSalvar[mapaId] = true;
    clearTimeout(saveTimer);
    saveTimer = setTimeout(flushSalvar, 700);
  }

  function flushSalvar() {
    var ids = Object.keys(pendentesSalvar);
    if (!ids.length) return;
    pendentesSalvar = {};
    ids.forEach(function (mid) {
      var mapaId = parseInt(mid, 10);
      var st = estadoMapa(mapaId);
      if (!st || !st.disparos.length) return;
      var pontos = st.disparos.map(function (d, idx) {
        return {
          chave: chaveDisparo(mapaId, d),
          mapa_ambiente_id: mapaId,
          idCliente: ctx.idCliente,
          idFranqueado: ctx.idFranqueado,
          idProcesso: d.idProcesso || st.processoId || "",
          idSetor: d.idSetor,
          labelSetor: d.label || "",
          posX: d.posX,
          posY: d.posY,
          sequencia: idx + 1,
          predito_id_setor: st.preditoId || "",
          direcao: st.direcao || "",
          evento_ts: tsISO(d.ts),
        };
      });
      $.ajax({
        url: "/centroOperacional/rastro/salvar",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({ pontos: pontos }),
      }).fail(function () {
        servidorOk = false;
      });
    });
  }

  function mergeServidor(mapaId, rows) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId || !rows || !rows.length) return;
    var st = estadoMapa(mapaId);
    rows.forEach(function (row) {
      var idSetor = row.idSetor || row.IdSetor;
      if (!idSetor) return;
      var ts = parseTsServidor(row.evento_ts || row.EventoTs);
      var dup = st.disparos.some(function (d) {
        return String(d.idSetor) === String(idSetor) && Math.abs(d.ts - ts) < 1500;
      });
      if (dup) return;
      st.disparos.push({
        idSetor: idSetor,
        label: row.labelSetor || row.LabelSetor || idSetor,
        posX: Number(row.posX || row.PosX) || 0,
        posY: Number(row.posY || row.PosY) || 0,
        mapaId: mapaId,
        idProcesso: row.idProcesso || row.IdProcesso || "",
        ts: ts,
      });
      if (row.predito_id_setor || row.PreditoIdSetor) {
        st.preditoId = row.predito_id_setor || row.PreditoIdSetor;
      }
      if (row.direcao || row.Direcao) {
        st.direcao = row.direcao || row.Direcao;
      }
    });
    st.disparos.sort(function (a, b) {
      return a.ts - b.ts;
    });
    sincronizarSessao();
  }

  function carregarServidor(mapaId, cb) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId) {
      if (cb) cb();
      return;
    }
    $.ajax({
      url: "/centroOperacional/rastro/listar",
      method: "POST",
      contentType: "application/json",
      data: JSON.stringify({
        mapa_ambiente_id: mapaId,
        idCliente: ctx.idCliente || undefined,
        limite: MAX_SESSAO,
      }),
    })
      .done(function (r) {
        var lista = r.dados || r.items || (Array.isArray(r) ? r : []);
        if (r.status !== "Vazio" && lista && lista.length) {
          mergeServidor(mapaId, lista);
          desenhar(mapaId);
        }
        if (cb) cb();
      })
      .fail(function () {
        if (cb) cb();
      });
  }

  function direcaoTexto(dx, dy) {
    if (Math.abs(dx) < 0.5 && Math.abs(dy) < 0.5) return "";
    var ang = (Math.atan2(dy, dx) * 180) / Math.PI;
    if (ang < 0) ang += 360;
    if (ang >= 337.5 || ang < 22.5) return "Leste";
    if (ang < 67.5) return "Sudeste";
    if (ang < 112.5) return "Sul";
    if (ang < 157.5) return "Sudoeste";
    if (ang < 202.5) return "Oeste";
    if (ang < 247.5) return "Noroeste";
    if (ang < 292.5) return "Norte";
    return "Nordeste";
  }

  function dist(a, b) {
    var dx = a.posX - b.posX;
    var dy = a.posY - b.posY;
    return Math.sqrt(dx * dx + dy * dy);
  }

  function preverProximo(mapaId, setores) {
    var st = estadoMapa(mapaId);
    if (!st || st.disparos.length < 2) {
      st.preditoId = null;
      st.direcao = "";
      return null;
    }

    var a = st.disparos[st.disparos.length - 2];
    var b = st.disparos[st.disparos.length - 1];
    var dx = b.posX - a.posX;
    var dy = b.posY - a.posY;
    st.direcao = direcaoTexto(dx, dy);

    if (st.disparos.length < 3 && dist(a, b) < 1) {
      st.preditoId = null;
      return null;
    }

    var alvo = {
      posX: b.posX + dx,
      posY: b.posY + dy,
    };

    var usados = {};
    st.disparos.forEach(function (d) {
      usados[String(d.idSetor)] = true;
    });

    var melhor = null;
    var melhorD = Infinity;
    (setores || []).forEach(function (raw) {
      var p = typeof mapaNormalizarSetor === "function" ? mapaNormalizarSetor(raw) : raw;
      var id = String(p.idSetor || "");
      if (!id || usados[id]) return;
      var cand = { idSetor: id, posX: Number(p.posX) || 0, posY: Number(p.posY) || 0, label: p.label || id };
      var d = dist(cand, alvo);
      var alinhado =
        dx * (cand.posX - b.posX) + dy * (cand.posY - b.posY) > 0 || dist(a, b) < 2;
      if (!alinhado) return;
      if (d < melhorD && d <= PRED_DIST_MAX) {
        melhorD = d;
        melhor = cand;
      }
    });

    st.preditoId = melhor ? melhor.idSetor : null;
    return melhor
      ? {
          setor: melhor,
          direcao: st.direcao,
          mensagem:
            "Invasor em deslocamento na direção " +
            (st.direcao || "desconhecida") +
            (melhor.label ? " · próxima zona provável: " + melhor.label : ""),
        }
      : st.direcao
      ? {
          setor: null,
          direcao: st.direcao,
          mensagem: "Invasor em deslocamento na direção " + st.direcao,
        }
      : null;
  }

  function emitirAlerta(mapaId, pred) {
    var st = estadoMapa(mapaId);
    if (!pred || !pred.mensagem) return;
    if (st.ultimoAlerta === pred.mensagem) return;
    st.ultimoAlerta = pred.mensagem;
    if (typeof hooks.onAlerta === "function") hooks.onAlerta(mapaId, pred);
  }

  function limparAntigos(st) {
    var lim = agora() - MAX_JANELA_MS;
    st.disparos = st.disparos.filter(function (d) {
      return d.ts >= lim;
    });
  }

  function registrarDisparo(opts) {
    var mapaId = parseInt(opts.mapaId, 10);
    var idSetor = String(opts.idSetor || "").trim();
    if (!mapaId || !idSetor) return null;

    var setor = resolverSetor(mapaId, idSetor, opts.setores);
    if (!setor && (opts.posX != null || opts.posY != null)) {
      setor = {
        idSetor: idSetor,
        label: opts.label || idSetor,
        posX: Number(opts.posX) || 0,
        posY: Number(opts.posY) || 0,
      };
    }
    if (!setor) return null;

    var st = estadoMapa(mapaId);
    limparAntigos(st);

    var ultimo = st.disparos[st.disparos.length - 1];
    if (ultimo && String(ultimo.idSetor) === idSetor && agora() - ultimo.ts < 8000) {
      ultimo.ts = agora();
      ultimo.idProcesso = opts.idProcesso || ultimo.idProcesso;
      sincronizarSessao();
      return ultimo;
    }

    var item = {
      idSetor: setor.idSetor,
      label: setor.label,
      posX: setor.posX,
      posY: setor.posY,
      mapaId: mapaId,
      idProcesso: opts.idProcesso || st.processoId || "",
      ts: agora(),
    };
    st.disparos.push(item);
    if (opts.idProcesso) st.processoId = opts.idProcesso;

    var pred = preverProximo(mapaId, opts.setores);
    emitirAlerta(mapaId, pred);
    sincronizarSessao();
    desenhar(mapaId);
    agendarSalvar(mapaId);
    if (typeof hooks.onChange === "function") hooks.onChange(mapaId, st);
    return item;
  }

  function sincronizarSessao() {
    var todos = [];
    Object.keys(estadoPorMapa).forEach(function (mid) {
      estadoPorMapa[mid].disparos.forEach(function (d) {
        todos.push(d);
      });
    });
    todos.sort(function (a, b) {
      return a.ts - b.ts;
    });
    gravarSessao(todos);
  }

  function carregarSessao() {
    var lista = lerSessao();
    var lim = agora() - MAX_JANELA_MS;
    lista.forEach(function (d) {
      if (!d || !d.mapaId || !d.idSetor || d.ts < lim) return;
      var st = estadoMapa(d.mapaId);
      if (st.disparos.some(function (x) { return x.idSetor === d.idSetor && Math.abs(x.ts - d.ts) < 1000; })) {
        return;
      }
      st.disparos.push(d);
    });
    Object.keys(estadoPorMapa).forEach(function (mid) {
      estadoPorMapa[mid].disparos.sort(function (a, b) { return a.ts - b.ts; });
    });
  }

  function appendSvgEl(g, tag) {
    return document.createElementNS("http://www.w3.org/2000/svg", tag);
  }

  function desenharChevron(g, cx, cy, ux, uy, size, fill, opacity) {
    var px = -uy;
    var py = ux;
    var wing = size * 0.46;
    var tipX = cx + ux * size * 0.52;
    var tipY = cy + uy * size * 0.52;
    var backX = cx - ux * size * 0.48;
    var backY = cy - uy * size * 0.48;
    var d =
      "M" +
      (backX - px * wing) +
      "," +
      (backY - py * wing) +
      " L" +
      tipX +
      "," +
      tipY +
      " L" +
      (backX + px * wing) +
      "," +
      (backY + py * wing) +
      " Z";
    var path = appendSvgEl(g, "path");
    path.setAttribute("d", d);
    path.setAttribute("fill", fill);
    path.setAttribute("opacity", String(opacity));
    path.setAttribute("class", "mapa-rastro-chevron");
    g.appendChild(path);
  }

  function desenharTraco(g, x1, y1, x2, y2, stroke, width, opacity) {
    var line = appendSvgEl(g, "line");
    line.setAttribute("x1", x1);
    line.setAttribute("y1", y1);
    line.setAttribute("x2", x2);
    line.setAttribute("y2", y2);
    line.setAttribute("stroke", stroke);
    line.setAttribute("stroke-width", String(width));
    line.setAttribute("stroke-linecap", "round");
    line.setAttribute("opacity", String(opacity));
    line.setAttribute("class", "mapa-rastro-traco");
    g.appendChild(line);
  }

  // Padrão normal: >>>>>>>  |  calor/análise: -->>>>>>>--
  function desenharSegmentoTrilha(g, x1, y1, x2, y2, opts) {
    opts = opts || {};
    var tipo = opts.tipo || "normal";
    var cor = opts.cor || "#ef4444";
    var opacity = opts.opacity != null ? opts.opacity : 1;
    var dx = x2 - x1;
    var dy = y2 - y1;
    var len = Math.sqrt(dx * dx + dy * dy) || 1;
    if (len < 6) return;
    var ux = dx / len;
    var uy = dy / len;
    var dashLen = 7;
    var arrowStep = 10;
    var arrowSize = 6.5;
    var pos = 0;

    function ponto(dist) {
      return { x: x1 + ux * dist, y: y1 + uy * dist };
    }

    function traco(distA, distB) {
      if (distB - distA < 1) return;
      var a = ponto(distA);
      var b = ponto(distB);
      desenharTraco(g, a.x, a.y, b.x, b.y, cor, 1.6, opacity);
    }

    function seta(dist) {
      var c = ponto(dist);
      desenharChevron(g, c.x, c.y, ux, uy, arrowSize, cor, opacity);
    }

    if (tipo === "calor") {
      traco(pos, pos + dashLen);
      pos += dashLen;
      traco(pos, pos + dashLen);
      pos += dashLen;
      while (pos + arrowStep + dashLen * 2 <= len) {
        seta(pos + arrowStep * 0.5);
        pos += arrowStep;
      }
      if (len - pos >= dashLen * 2) {
        traco(len - dashLen * 2, len - dashLen);
        traco(len - dashLen, len);
      } else if (len - pos >= dashLen) {
        traco(len - dashLen, len);
      }
      return;
    }

    if (len < arrowStep * 1.5) {
      seta(len * 0.5);
      return;
    }

    while (pos + arrowStep <= len) {
      seta(pos + arrowStep * 0.5);
      pos += arrowStep;
    }
  }

  function pontosSegmentoTrilha(a, b, offset) {
    var dx = b.x - a.x;
    var dy = b.y - a.y;
    var len = Math.sqrt(dx * dx + dy * dy) || 1;
    var insetInicio = Math.min(18, len * 0.22);
    var insetFim = Math.min(22, len * 0.32);
    if (insetInicio + insetFim >= len - 4) {
      insetInicio = Math.max(6, len * 0.18);
      insetFim = Math.max(8, len * 0.28);
    }
    var ux = dx / len;
    var uy = dy / len;
    var ox = -uy * offset;
    var oy = ux * offset;
    return {
      x1: a.x + ux * insetInicio + ox,
      y1: a.y + uy * insetInicio + oy,
      x2: b.x - ux * insetFim + ox,
      y2: b.y - uy * insetFim + oy,
    };
  }

  function garantirOverlay(mapaId) {
    var $container = $("#coPane" + mapaId + " .mapa-container");
    if (!$container.length) return $();
    var $ov = $container.children(".mapa-rastro-overlay");
    if (!$ov.length) {
      $ov = $(
        '<svg class="mapa-rastro-overlay" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">' +
          '<g class="mapa-rastro-linhas"></g>' +
          '<g class="mapa-rastro-marcadores"></g>' +
          "</svg>"
      );
      $container.append($ov);
    }
    return $ov;
  }

  function pctParaPx(posX, posY, $container, $img) {
    if (typeof MapaLayout === "undefined") return null;
    var area = MapaLayout.obterArea($container, $img);
    if (!area) return null;
    return {
      x: area.offsetX + (posX / 100) * area.width,
      y: area.offsetY + (posY / 100) * area.height,
    };
  }

  function limparClassesPontos(mapaId) {
    $("#coPane" + mapaId + " .mapa-ponto-monitor")
      .removeClass("mapa-ponto-rastro-antigo mapa-ponto-rastro-atual mapa-ponto-rastro-predito");
  }

  function aplicarClassesPontos(mapaId, st, ateIdx) {
    limparClassesPontos(mapaId);
    if (!st || !st.disparos.length) return;
    var fim = ateIdx == null ? st.disparos.length - 1 : ateIdx;
    for (var i = 0; i <= fim; i++) {
      var d = st.disparos[i];
      var $p = $('#coPane' + mapaId + ' .mapa-ponto-monitor[data-id-setor="' + d.idSetor + '"]');
      if (!$p.length) continue;
      if (i === fim) $p.addClass("mapa-ponto-rastro-atual");
      else $p.addClass("mapa-ponto-rastro-antigo");
    }
    if (!modoAnalise && st.preditoId) {
      $('#coPane' + mapaId + ' .mapa-ponto-monitor[data-id-setor="' + st.preditoId + '"]')
        .addClass("mapa-ponto-rastro-predito");
    }
  }

  function desenhar(mapaId, limiteIdx) {
    mapaId = parseInt(mapaId, 10);
    var st = estadoMapa(mapaId);
    if (!st) return;

    var $container = $("#coPane" + mapaId + " .mapa-container");
    var $img = $("#coPane" + mapaId + " .mapa-imagem");
    if (!$container.length) return;

    var $ov = garantirOverlay(mapaId);
    var $linhas = $ov.find(".mapa-rastro-linhas");
    var $marks = $ov.find(".mapa-rastro-marcadores");
    $linhas.empty();
    $marks.empty();

    var pontos = st.disparos;
    var fim = limiteIdx == null ? pontos.length - 1 : Math.min(limiteIdx, pontos.length - 1);
    if (fim < 0) {
      limparClassesPontos(mapaId);
      return;
    }

    var coords = [];
    for (var i = 0; i <= fim; i++) {
      var px = pctParaPx(pontos[i].posX, pontos[i].posY, $container, $img);
      if (px) coords.push({ d: pontos[i], px: px, idx: i });
    }

    var estiloTrilha = modoAnalise ? "calor" : "normal";

    for (var j = 1; j < coords.length; j++) {
      var seg = pontosSegmentoTrilha(coords[j - 1].px, coords[j].px, 6);
      var op = 0.45 + (j / coords.length) * 0.5;
      desenharSegmentoTrilha($linhas[0], seg.x1, seg.y1, seg.x2, seg.y2, {
        tipo: estiloTrilha,
        cor: "#ef4444",
        opacity: op,
      });
    }

    if (!modoAnalise && st.preditoId && limiteIdx == null) {
      var predSetor = resolverSetor(mapaId, st.preditoId, getSetoresMapa(mapaId));
      var ult = coords[coords.length - 1];
      if (predSetor && ult) {
        var pPred = pctParaPx(predSetor.posX, predSetor.posY, $container, $img);
        if (pPred) {
          var segPred = pontosSegmentoTrilha(ult.px, pPred, 6);
          desenharSegmentoTrilha($linhas[0], segPred.x1, segPred.y1, segPred.x2, segPred.y2, {
            tipo: "calor",
            cor: "#f97316",
            opacity: 0.9,
          });
        }
      }
    }

    coords.forEach(function (c, n) {
      var circ = document.createElementNS("http://www.w3.org/2000/svg", "circle");
      circ.setAttribute("cx", c.px.x);
      circ.setAttribute("cy", c.px.y);
      circ.setAttribute("r", n === coords.length - 1 ? 7 : 5);
      circ.setAttribute(
        "class",
        "mapa-rastro-dot" + (n === coords.length - 1 ? " mapa-rastro-dot-atual" : " mapa-rastro-dot-antigo")
      );
      $marks[0].appendChild(circ);

      var txt = document.createElementNS("http://www.w3.org/2000/svg", "text");
      txt.setAttribute("x", c.px.x);
      txt.setAttribute("y", c.px.y - 10);
      txt.setAttribute("class", "mapa-rastro-seq");
      txt.setAttribute("text-anchor", "middle");
      txt.textContent = String(n + 1);
      $marks[0].appendChild(txt);
    });

    aplicarClassesPontos(mapaId, st, fim);
  }

  function getSetoresMapa(mapaId) {
    if (typeof estadoCO === "undefined") return [];
    var mon = estadoCO.monitores && estadoCO.monitores[mapaId];
    if (mon && typeof mon.getSetores === "function") return mon.getSetores() || [];
    return [];
  }

  function observarStatus(mapaId, statusMap, setores) {
    var st = estadoMapa(mapaId);
    if (!st) return;
    if (!st._statusAnterior) st._statusAnterior = {};
    var prev = st._statusAnterior;
    Object.keys(statusMap || {}).forEach(function (idSetor) {
      var agoraSt = statusMap[idSetor];
      var antes = prev[idSetor] || "normal";
      if (agoraSt === "alarme" && antes !== "alarme") {
        registrarDisparo({
          mapaId: mapaId,
          idSetor: idSetor,
          setores: setores || getSetoresMapa(mapaId),
        });
      }
      prev[idSetor] = agoraSt;
    });
    desenhar(mapaId);
  }

  function sincronizarProcesso(mapaId, idProcesso, eventos, setores) {
    mapaId = parseInt(mapaId, 10);
    if (!mapaId || !eventos || !eventos.length) return;
    var st = estadoMapa(mapaId);
    st.processoId = idProcesso || st.processoId;

    eventos.forEach(function (ev) {
      var idSetor = String(ev.idSetor || "").trim();
      if (!idSetor) return;
      var grupo = String(ev.grupo || "").toUpperCase();
      if (grupo && grupo.indexOf("ALARME") < 0 && grupo.indexOf("PANICO") < 0 && grupo.indexOf("EMERGENCIA") < 0) {
        return;
      }
      registrarDisparo({
        mapaId: mapaId,
        idSetor: idSetor,
        label: ev.descZona || ev.descricao || idSetor,
        idProcesso: idProcesso,
        setores: setores || getSetoresMapa(mapaId),
      });
    });
  }

  function limparMapa(mapaId) {
    var st = estadoMapa(mapaId);
    if (!st) return;
    st.disparos = [];
    st.preditoId = null;
    st.direcao = "";
    st.ultimoAlerta = "";
    st.processoId = "";
    st._statusAnterior = {};
    limparClassesPontos(mapaId);
    var $ov = $("#coPane" + mapaId + " .mapa-rastro-overlay");
    $ov.find(".mapa-rastro-linhas, .mapa-rastro-marcadores").empty();
    sincronizarSessao();
    if (typeof hooks.onChange === "function") hooks.onChange(mapaId, st);
  }

  function limparTudo() {
    Object.keys(estadoPorMapa).forEach(function (mid) {
      limparMapa(mid);
    });
    try {
      sessionStorage.removeItem(STORAGE_KEY);
    } catch (e) {}
  }

  function getTrail(mapaId) {
    var st = estadoMapa(mapaId);
    return st ? st.disparos.slice() : [];
  }

  function getPredicao(mapaId) {
    var st = estadoMapa(mapaId);
    if (!st) return null;
    return {
      preditoId: st.preditoId,
      direcao: st.direcao,
      mensagem: st.ultimoAlerta,
    };
  }

  function setModoAnalise(ativo) {
    modoAnalise = !!ativo;
    if (!modoAnalise) {
      pararAnalise();
      if (typeof estadoCO !== "undefined" && estadoCO.mapaAtivoId) {
        desenhar(estadoCO.mapaAtivoId);
      }
    }
    $("body").toggleClass("co-modo-analise-calor", modoAnalise);
    return modoAnalise;
  }

  function isModoAnalise() {
    return modoAnalise;
  }

  function pararAnalise() {
    if (analiseTimer) {
      clearInterval(analiseTimer);
      analiseTimer = null;
    }
  }

  function reproduzirAnalise(mapaId, opts) {
    mapaId = parseInt(mapaId, 10) || (typeof estadoCO !== "undefined" && estadoCO.mapaAtivoId);
    var st = estadoMapa(mapaId);
    if (!st || !st.disparos.length) return;
    setModoAnalise(true);
    pararAnalise();
    analiseIdx = 0;
    var passo = (opts && opts.intervalo) || 900;
    desenhar(mapaId, 0);
    if (typeof hooks.onChange === "function") {
      hooks.onChange(mapaId, st, { analiseIdx: 0, total: st.disparos.length });
    }
    analiseTimer = setInterval(function () {
      analiseIdx++;
      if (analiseIdx >= st.disparos.length) {
        pararAnalise();
        desenhar(mapaId);
        if (typeof hooks.onChange === "function") {
          hooks.onChange(mapaId, st, { analiseIdx: st.disparos.length - 1, total: st.disparos.length, fim: true });
        }
        return;
      }
      desenhar(mapaId, analiseIdx);
      if (typeof hooks.onChange === "function") {
        hooks.onChange(mapaId, st, { analiseIdx: analiseIdx, total: st.disparos.length });
      }
    }, passo);
  }

  function irParaPasso(mapaId, idx) {
    var st = estadoMapa(mapaId);
    if (!st || !st.disparos.length) return;
    pararAnalise();
    analiseIdx = Math.max(0, Math.min(idx, st.disparos.length - 1));
    desenhar(mapaId, analiseIdx);
  }

  function init(options) {
    hooks.onAlerta = options && options.onAlerta;
    hooks.onChange = options && options.onChange;
    configurarContexto(options);
    carregarSessao();
  }

  return {
    init: init,
    configurarContexto: configurarContexto,
    carregarServidor: carregarServidor,
    registrarDisparo: registrarDisparo,
    observarStatus: observarStatus,
    sincronizarProcesso: sincronizarProcesso,
    desenhar: desenhar,
    limparMapa: limparMapa,
    limparTudo: limparTudo,
    getTrail: getTrail,
    getPredicao: getPredicao,
    setModoAnalise: setModoAnalise,
    isModoAnalise: isModoAnalise,
    reproduzirAnalise: reproduzirAnalise,
    irParaPasso: irParaPasso,
    pararAnalise: pararAnalise,
  };
})();
