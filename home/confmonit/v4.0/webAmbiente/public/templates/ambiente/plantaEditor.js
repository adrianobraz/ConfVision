/* Editor de planta baixa — canvas 1920×1080, export PNG para upload */
var PlantaEditor = (function () {
  var W = 1920;
  var H = 1080;
  var GRID = 20;
  var WALL_W = 14;
  var DOOR_W = 90;
  var WIN_W = 110;

  var canvas, ctx;
  var objects = [];
  var tool = "selecionar";
  var selectedId = null;
  var history = [];
  var drag = null;
  var drawState = null;
  var onChange = null;
  var onHint = null;

  function hint(msg) {
    if (onHint) onHint(msg);
  }

  function snap(v) {
    return Math.round(v / GRID) * GRID;
  }

  function snapPt(x, y) {
    return { x: snap(x), y: snap(y) };
  }

  function uid() {
    return "o" + Date.now().toString(36) + Math.random().toString(36).slice(2, 7);
  }

  function pushHistory() {
    history.push(JSON.stringify(objects));
    if (history.length > 40) history.shift();
  }

  function undo() {
    if (history.length === 0) return;
    objects = JSON.parse(history.pop());
    selectedId = null;
    redraw();
    notifyChange();
  }

  function notifyChange() {
    if (onChange) onChange(hasContent());
  }

  function hasContent() {
    return objects.length > 0;
  }

  function canvasPt(e) {
    var rect = canvas.getBoundingClientRect();
    var sx = W / rect.width;
    var sy = H / rect.height;
    return {
      x: snap((e.clientX - rect.left) * sx),
      y: snap((e.clientY - rect.top) * sy),
    };
  }

  function distSeg(px, py, x1, y1, x2, y2) {
    var dx = x2 - x1;
    var dy = y2 - y1;
    var len2 = dx * dx + dy * dy;
    if (len2 === 0) return Math.hypot(px - x1, py - y1);
    var t = Math.max(0, Math.min(1, ((px - x1) * dx + (py - y1) * dy) / len2));
    var qx = x1 + t * dx;
    var qy = y1 + t * dy;
    return { d: Math.hypot(px - qx, py - qy), t: t, qx: qx, qy: qy, ang: Math.atan2(dy, dx) };
  }

  function snapWallEnd(x1, y1, x2, y2) {
    var dx = Math.abs(x2 - x1);
    var dy = Math.abs(y2 - y1);
    if (dx > dy) return { x: snap(x2), y: y1 };
    return { x: x1, y: snap(y2) };
  }

  function nearestWall(px, py, maxD) {
    var best = null;
    objects.forEach(function (o) {
      if (o.type !== "wall") return;
      var r = distSeg(px, py, o.x1, o.y1, o.x2, o.y2);
      var lim = maxD + WALL_W / 2;
      if (r.d <= lim && (!best || r.d < best.d)) {
        best = { wall: o, d: r.d, qx: r.qx, qy: r.qy, ang: r.ang };
      }
    });
    return best;
  }

  function hitObject(px, py) {
    for (var i = objects.length - 1; i >= 0; i--) {
      var o = objects[i];
      if (o.type === "wall") {
        var r = distSeg(px, py, o.x1, o.y1, o.x2, o.y2);
        if (r.d <= WALL_W + 6) return o;
      } else if (o.type === "text") {
        ctx.save();
        ctx.font = "600 " + o.fontSize + "px sans-serif";
        var tw = ctx.measureText(o.label).width;
        ctx.restore();
        if (px >= o.x - 4 && px <= o.x + tw + 4 && py >= o.y - o.fontSize && py <= o.y + 4) return o;
      } else if (o.type === "door" || o.type === "window") {
        if (Math.hypot(px - o.cx, py - o.cy) <= (o.type === "door" ? DOOR_W : WIN_W) * 0.6) return o;
      } else if (o.type === "external") {
        if (pointInPoly(px, py, o.points)) return o;
      }
    }
    return null;
  }

  function pointInPoly(x, y, pts) {
    var inside = false;
    for (var i = 0, j = pts.length - 1; i < pts.length; j = i++) {
      var xi = pts[i].x, yi = pts[i].y;
      var xj = pts[j].x, yj = pts[j].y;
      if (yi > y !== yj > y && x < ((xj - xi) * (y - yi)) / (yj - yi) + xi) inside = !inside;
    }
    return inside;
  }

  function drawGrid() {
    ctx.strokeStyle = "#e8e8ec";
    ctx.lineWidth = 1;
    for (var x = 0; x <= W; x += GRID) {
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, H);
      ctx.stroke();
    }
    for (var y = 0; y <= H; y += GRID) {
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(W, y);
      ctx.stroke();
    }
  }

  function drawWall(w, highlight) {
    ctx.strokeStyle = highlight ? "#2563eb" : "#1a1a2e";
    ctx.lineWidth = WALL_W;
    ctx.lineCap = "square";
    ctx.beginPath();
    ctx.moveTo(w.x1, w.y1);
    ctx.lineTo(w.x2, w.y2);
    ctx.stroke();
  }

  function drawExternal(o, highlight) {
    if (o.points.length < 3) return;
    ctx.beginPath();
    ctx.moveTo(o.points[0].x, o.points[0].y);
    for (var i = 1; i < o.points.length; i++) ctx.lineTo(o.points[i].x, o.points[i].y);
    ctx.closePath();
    ctx.fillStyle = highlight ? "rgba(96, 165, 250, 0.55)" : "rgba(147, 197, 253, 0.5)";
    ctx.fill();
    ctx.strokeStyle = highlight ? "#2563eb" : "#60a5fa";
    ctx.lineWidth = 2;
    ctx.stroke();
  }

  function drawDoor(o, highlight) {
    var hw = DOOR_W / 2;
    var ang = o.ang;
    var cx = o.cx;
    var cy = o.cy;
    var cos = Math.cos(ang);
    var sin = Math.sin(ang);
    var x1 = cx - cos * hw;
    var y1 = cy - sin * hw;
    var x2 = cx + cos * hw;
    var y2 = cy + sin * hw;

    ctx.strokeStyle = highlight ? "#2563eb" : "#1a1a2e";
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.stroke();

    var swing = o.swing || 1;
    var perp = ang + (Math.PI / 2) * swing;
    ctx.beginPath();
    ctx.arc(x1, y1, DOOR_W, ang, ang + (Math.PI / 2) * swing, swing < 0);
    ctx.strokeStyle = highlight ? "#2563eb" : "#64748b";
    ctx.lineWidth = 1.5;
    ctx.stroke();
  }

  function drawWindow(o, highlight) {
    var hw = WIN_W / 2;
    var ang = o.ang;
    var cos = Math.cos(ang);
    var sin = Math.sin(ang);
    var x1 = o.cx - cos * hw;
    var y1 = o.cy - sin * hw;
    var x2 = o.cx + cos * hw;
    var y2 = o.cy + sin * hw;

    ctx.strokeStyle = "#ffffff";
    ctx.lineWidth = WALL_W + 6;
    ctx.lineCap = "butt";
    ctx.beginPath();
    ctx.moveTo(x1, y1);
    ctx.lineTo(x2, y2);
    ctx.stroke();

    ctx.strokeStyle = highlight ? "#2563eb" : "#0284c7";
    ctx.lineWidth = 3;
    var px = -sin;
    var py = cos;
    var gap = 6;
    for (var side = -1; side <= 1; side += 2) {
      var ox = o.cx + px * gap * side;
      var oy = o.cy + py * gap * side;
      ctx.beginPath();
      ctx.moveTo(ox - cos * hw, oy - sin * hw);
      ctx.lineTo(ox + cos * hw, oy + sin * hw);
      ctx.stroke();
    }
  }

  function drawExternalPreview(state) {
    if (!state || !state.points.length) return;

    var placed = state.points;
    var pts = placed.slice();
    if (state.cur) pts.push(state.cur);

    if (pts.length >= 2) {
      ctx.beginPath();
      ctx.moveTo(pts[0].x, pts[0].y);
      for (var i = 1; i < pts.length; i++) {
        ctx.lineTo(pts[i].x, pts[i].y);
      }
      ctx.strokeStyle = "#3b82f6";
      ctx.lineWidth = 3;
      ctx.lineCap = "round";
      ctx.lineJoin = "round";
      ctx.setLineDash([]);
      ctx.stroke();
    }

    if (placed.length >= 3) {
      ctx.beginPath();
      ctx.moveTo(placed[0].x, placed[0].y);
      for (var j = 1; j < placed.length; j++) {
        ctx.lineTo(placed[j].x, placed[j].y);
      }
      ctx.closePath();
      ctx.fillStyle = "rgba(147, 197, 253, 0.4)";
      ctx.fill();
    }

    if (placed.length >= 3 && state.cur) {
      var first = placed[0];
      var perto = Math.hypot(state.cur.x - first.x, state.cur.y - first.y) < GRID * 2;
      ctx.beginPath();
      ctx.moveTo(state.cur.x, state.cur.y);
      ctx.lineTo(first.x, first.y);
      ctx.strokeStyle = perto ? "#2563eb" : "rgba(59, 130, 246, 0.55)";
      ctx.lineWidth = perto ? 3 : 2;
      ctx.setLineDash(perto ? [] : [6, 5]);
      ctx.stroke();
      ctx.setLineDash([]);
    }

    placed.forEach(function (p, idx) {
      ctx.beginPath();
      ctx.arc(p.x, p.y, idx === 0 ? 10 : 6, 0, Math.PI * 2);
      ctx.fillStyle = idx === 0 ? "#2563eb" : "#64748b";
      ctx.fill();
      ctx.strokeStyle = "#ffffff";
      ctx.lineWidth = 2;
      ctx.stroke();
    });

    if (state.cur) {
      ctx.beginPath();
      ctx.arc(state.cur.x, state.cur.y, 5, 0, Math.PI * 2);
      ctx.fillStyle = "rgba(59, 130, 246, 0.75)";
      ctx.fill();
    }
  }

  function finishExternalArea() {
    if (!drawState || drawState.type !== "externa" || drawState.points.length < 3) return false;
    addObject({ id: uid(), type: "external", points: drawState.points.slice() });
    drawState = null;
    hint("Área externa criada.");
    return true;
  }
  function drawText(o, highlight) {
    ctx.font = "600 " + o.fontSize + "px system-ui, sans-serif";
    ctx.fillStyle = highlight ? "#2563eb" : "#1e293b";
    ctx.textBaseline = "bottom";
    ctx.fillText(o.label, o.x, o.y);
  }

  function redraw() {
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, W, H);
    drawGrid();

    objects.forEach(function (o) {
      if (o.type === "external") drawExternal(o, o.id === selectedId);
    });
    objects.forEach(function (o) {
      if (o.type === "wall") drawWall(o, o.id === selectedId);
    });
    objects.forEach(function (o) {
      if (o.type === "door") drawDoor(o, o.id === selectedId);
      if (o.type === "window") drawWindow(o, o.id === selectedId);
    });
    objects.forEach(function (o) {
      if (o.type === "text") drawText(o, o.id === selectedId);
    });

    if (drawState && drawState.type === "wall") {
      var end = snapWallEnd(drawState.x1, drawState.y1, drawState.x2, drawState.y2);
      drawWall({ x1: drawState.x1, y1: drawState.y1, x2: end.x, y2: end.y }, true);
    }
    if (drawState && drawState.type === "retangulo") {
      drawRectPreview(drawState.x1, drawState.y1, drawState.x2, drawState.y2, true);
    }
    if (drawState && drawState.type === "externa") {
      drawExternalPreview(drawState);
    }
  }

  function addObjectsBatch(list) {
    if (!list.length) return;
    pushHistory();
    list.forEach(function (o) {
      objects.push(o);
    });
    redraw();
    notifyChange();
  }

  function addObject(o) {
    addObjectsBatch([o]);
  }

  function rectBounds(x1, y1, x2, y2) {
    return {
      left: Math.min(x1, x2),
      right: Math.max(x1, x2),
      top: Math.min(y1, y2),
      bottom: Math.max(y1, y2),
    };
  }

  function addRectangleWalls(x1, y1, x2, y2) {
    var b = rectBounds(x1, y1, x2, y2);
    if (b.right - b.left < GRID * 2 || b.bottom - b.top < GRID * 2) return false;
    addObjectsBatch([
      { id: uid(), type: "wall", x1: b.left, y1: b.top, x2: b.right, y2: b.top },
      { id: uid(), type: "wall", x1: b.right, y1: b.top, x2: b.right, y2: b.bottom },
      { id: uid(), type: "wall", x1: b.right, y1: b.bottom, x2: b.left, y2: b.bottom },
      { id: uid(), type: "wall", x1: b.left, y1: b.bottom, x2: b.left, y2: b.top },
    ]);
    return true;
  }

  function drawRectPreview(x1, y1, x2, y2, highlight) {
    var b = rectBounds(snap(x1), snap(y1), snap(x2), snap(y2));
    var walls = [
      { x1: b.left, y1: b.top, x2: b.right, y2: b.top },
      { x1: b.right, y1: b.top, x2: b.right, y2: b.bottom },
      { x1: b.right, y1: b.bottom, x2: b.left, y2: b.bottom },
      { x1: b.left, y1: b.bottom, x2: b.left, y2: b.top },
    ];
    walls.forEach(function (w) {
      drawWall(w, highlight);
    });
  }

  function removeSelected() {
    if (!selectedId) return;
    pushHistory();
    objects = objects.filter(function (o) {
      return o.id !== selectedId;
    });
    selectedId = null;
    redraw();
    notifyChange();
  }

  function onMouseDown(e) {
    var p = canvasPt(e);

    if (tool === "selecionar") {
      var hit = hitObject(p.x, p.y);
      selectedId = hit ? hit.id : null;
      if (hit) {
        drag = { id: hit.id, ox: p.x, oy: p.y, start: JSON.parse(JSON.stringify(hit)) };
        pushHistory();
      }
      redraw();
      return;
    }

    if (tool === "parede") {
      drawState = { type: "wall", x1: p.x, y1: p.y, x2: p.x, y2: p.y };
      return;
    }

    if (tool === "retangulo") {
      drawState = { type: "retangulo", x1: p.x, y1: p.y, x2: p.x, y2: p.y };
      return;
    }

    if (tool === "externa") {
      if (e.detail >= 2) return;

      if (!drawState || drawState.type !== "externa") {
        drawState = { type: "externa", points: [p], cur: p };
      } else {
        var first = drawState.points[0];
        if (drawState.points.length >= 3 && Math.hypot(p.x - first.x, p.y - first.y) < GRID * 2) {
          finishExternalArea();
        } else {
          drawState.points.push(p);
          drawState.cur = p;
        }
      }
      redraw();
      return;
    }

    if (tool === "porta") {
      var nw = nearestWall(p.x, p.y, 40);
      if (!nw) {
        hint("Clique em cima de uma parede para colocar a porta.");
        return;
      }
      addObject({
        id: uid(),
        type: "door",
        cx: nw.qx,
        cy: nw.qy,
        ang: nw.ang,
        swing: 1,
      });
      hint("Porta adicionada.");
      return;
    }

    if (tool === "janela") {
      var nw2 = nearestWall(p.x, p.y, 40);
      if (!nw2) {
        hint("Clique em cima de uma parede para colocar a janela.");
        return;
      }
      addObject({
        id: uid(),
        type: "window",
        cx: nw2.qx,
        cy: nw2.qy,
        ang: nw2.ang,
      });
      hint("Janela adicionada.");
      return;
    }

    if (tool === "texto") {
      var label = window.prompt("Nome do ambiente:", "QUARTO");
      if (!label || !label.trim()) return;
      addObject({
        id: uid(),
        type: "text",
        x: p.x,
        y: p.y,
        label: label.trim().toUpperCase(),
        fontSize: 32,
      });
    }
  }

  function onMouseMove(e) {
    var p = canvasPt(e);

    if (drag) {
      var o = objects.find(function (x) {
        return x.id === drag.id;
      });
      if (!o) return;
      var dx = p.x - drag.ox;
      var dy = p.y - drag.oy;
      var s = drag.start;
      if (o.type === "wall") {
        o.x1 = s.x1 + dx;
        o.y1 = s.y1 + dy;
        o.x2 = s.x2 + dx;
        o.y2 = s.y2 + dy;
      } else if (o.type === "text") {
        o.x = s.x + dx;
        o.y = s.y + dy;
      } else if (o.type === "door" || o.type === "window") {
        o.cx = s.cx + dx;
        o.cy = s.cy + dy;
      } else if (o.type === "external") {
        o.points = s.points.map(function (pt) {
          return { x: pt.x + dx, y: pt.y + dy };
        });
      }
      redraw();
      return;
    }

    if (drawState && drawState.type === "wall") {
      var end = snapWallEnd(drawState.x1, drawState.y1, p.x, p.y);
      drawState.x2 = end.x;
      drawState.y2 = end.y;
      redraw();
    }
    if (drawState && drawState.type === "retangulo") {
      drawState.x2 = snap(p.x);
      drawState.y2 = snap(p.y);
      redraw();
    }
    if (drawState && drawState.type === "externa") {
      drawState.cur = p;
      redraw();
    }
  }

  function onMouseUp(e) {
    if (drag) {
      drag = null;
      notifyChange();
      return;
    }

    if (drawState && drawState.type === "retangulo") {
      var pr = canvasPt(e);
      if (!addRectangleWalls(drawState.x1, drawState.y1, snap(pr.x), snap(pr.y))) {
        redraw();
      }
      drawState = null;
      return;
    }

    if (drawState && drawState.type === "wall") {
      var p = canvasPt(e);
      var end = snapWallEnd(drawState.x1, drawState.y1, p.x, p.y);
      var len = Math.hypot(end.x - drawState.x1, end.y - drawState.y1);
      if (len >= GRID) {
        addObject({
          id: uid(),
          type: "wall",
          x1: drawState.x1,
          y1: drawState.y1,
          x2: end.x,
          y2: end.y,
        });
      } else {
        redraw();
      }
      drawState = null;
    }
  }

  function onDblClick(e) {
    if (tool !== "externa" || !drawState || drawState.type !== "externa") return;
    e.preventDefault();
    if (drawState.points.length >= 3) {
      finishExternalArea();
      redraw();
    } else {
      hint("Coloque pelo menos 3 pontos antes de fechar a área.");
    }
  }

  function onKeyDown(e) {
    if (e.key === "Delete" || e.key === "Backspace") {
      if (document.activeElement && document.activeElement.tagName === "INPUT") return;
      e.preventDefault();
      removeSelected();
    }
    if (e.key === "Escape") {
      drawState = null;
      selectedId = null;
      drag = null;
      redraw();
    }
  }

  function init(selector, opts) {
    canvas = document.querySelector(selector);
    if (!canvas) return;
    ctx = canvas.getContext("2d");
    canvas.width = W;
    canvas.height = H;
    onChange = opts && opts.onChange;
    onHint = opts && opts.onHint;

    canvas.addEventListener("mousedown", onMouseDown);
    canvas.addEventListener("mousemove", onMouseMove);
    canvas.addEventListener("mouseup", onMouseUp);
    canvas.addEventListener("mouseleave", function () {
      if (drag) {
        drag = null;
        notifyChange();
      }
      if (drawState && (drawState.type === "wall" || drawState.type === "retangulo")) {
        drawState = null;
        redraw();
      } else if (drawState && drawState.type === "externa") {
        drawState.cur = null;
        redraw();
      }
    });
    canvas.addEventListener("dblclick", onDblClick);
    document.addEventListener("keydown", onKeyDown);

    redraw();
  }

  function setTool(t) {
    tool = t;
    drawState = null;
    selectedId = null;
    drag = null;
    redraw();
  }

  function clear() {
    if (objects.length === 0) return;
    pushHistory();
    objects = [];
    selectedId = null;
    drawState = null;
    redraw();
    notifyChange();
  }

  function exportBlob(cb) {
    selectedId = null;
    redraw();
    canvas.toBlob(
      function (blob) {
        if (cb) cb(blob);
      },
      "image/png",
      1
    );
  }

  function destroy() {
    document.removeEventListener("keydown", onKeyDown);
  }

  return {
    init: init,
    setTool: setTool,
    undo: undo,
    clear: clear,
    removeSelected: removeSelected,
    exportBlob: exportBlob,
    hasContent: hasContent,
    destroy: destroy,
  };
})();
