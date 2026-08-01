/* Notificacoes administrativas ConfMonit (sino + modal no login) */
var CmNotificacoes = (function () {
  var SOFTWARE = 'franqueadopro'
  var BASE = {
    minhas: '/notificacoes/minhas',
    contagem: '/notificacoes/contagem',
    visto: '/notificacoes/marcar-visto',
    lido: '/notificacoes/marcar-lido',
  }
  var lista = []
  var aberto = false
  var modalFila = []
  var modalIdx = 0

  function esc(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
  }

  function post(url, body) {
    return $.ajax({
      url: url,
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify(body || {}),
    })
  }

  function garantirUI() {
    if ($('#cmNotifBtn').length) return
    var $actions = $('#cmNotifMount').first()
    if (!$actions.length) $actions = $('.cv-header-user, .cv-dvr-bar-user').first()
    if (!$actions.length) $actions = $('.fp-navbar-actions').first()
    if (!$actions.length) $actions = $('.amb-navbar-actions').first()
    if (!$actions.length) $actions = $('.navbar .collapse').first()
    if (!$actions.length) return

    var btnClass = 'btn cm-notif-btn'
    if ($('.cv-header, .cv-dvr-bar').length) {
      btnClass += ' cv-btn-ghost'
    } else {
      btnClass += ' fp-icon-btn'
    }

    var $wrap = $(
      '<div class="cm-notif-wrap dropdown">' +
        '<button type="button" id="cmNotifBtn" class="' +
        btnClass +
        '" title="Notificacoes" aria-label="Notificacoes">' +
        '<i class="bi bi-bell-fill"></i>' +
        '<span id="cmNotifBadge" class="cm-notif-badge cm-notif-hide">0</span>' +
        '</button>' +
        '<div id="cmNotifDrop" class="cm-notif-drop cm-notif-hide">' +
        '<div class="cm-notif-drop-head">Notificacoes</div>' +
        '<div id="cmNotifLista" class="cm-notif-lista"></div>' +
        '</div>' +
        '</div>'
    )
    if ($actions.is('#cmNotifMount') || $actions.hasClass('cv-notif-mount')) {
      $actions.append($wrap)
    } else if ($actions.hasClass('cv-header-user') || $actions.hasClass('cv-dvr-bar-user')) {
      $actions.prepend($wrap)
    } else {
      $actions.prepend($wrap)
    }

    $('#cmNotifBtn').on('click', function (e) {
      e.preventDefault()
      e.stopPropagation()
      aberto = !aberto
      $('#cmNotifDrop').toggleClass('cm-notif-hide', !aberto)
      if (aberto) renderLista()
    })
    $(document).off('click.cmNotif').on('click.cmNotif', function (e) {
      if (!aberto) return
      if ($(e.target).closest('.cm-notif-wrap').length) return
      aberto = false
      $('#cmNotifDrop').addClass('cm-notif-hide')
    })
    $('#cmNotifDrop').on('click', function (e) {
      e.stopPropagation()
    })
  }

  function setBadge(n) {
    var $b = $('#cmNotifBadge')
    if (!$b.length) return
    if (n > 0) {
      $b.text(n > 99 ? '99+' : String(n)).removeClass('cm-notif-hide')
    } else {
      $b.addClass('cm-notif-hide')
    }
  }

  function renderLista() {
    var $el = $('#cmNotifLista')
    if (!$el.length) return
    if (!lista.length) {
      $el.html('<div class="cm-notif-vazio">Nenhuma notificacao</div>')
      return
    }
    var html = ''
    lista.forEach(function (n) {
      var naoVisto = !n.vistoEm
      html +=
        '<button type="button" class="cm-notif-item' +
        (naoVisto ? ' cm-notif-item-nova' : '') +
        '" data-id="' +
        esc(n.idNotificacao) +
        '">' +
        '<div class="cm-notif-item-titulo">' +
        esc(n.titulo) +
        '</div>' +
        '<div class="cm-notif-item-meta">' +
        esc(n.tipo || '') +
        (n.dataCadastro ? ' · ' + esc(n.dataCadastro) : '') +
        '</div>' +
        '</button>'
    })
    $el.html(html)
    $el.find('.cm-notif-item').on('click', function () {
      var id = $(this).data('id')
      var item = lista.find(function (x) {
        return x.idNotificacao === id
      })
      aberto = false
      $('#cmNotifDrop').addClass('cm-notif-hide')
      if (item) abrirModal([item], 0, true)
    })
  }

  function garantirModal() {
    if ($('#cmNotifModal').length) return
    $('body').append(
      '<div id="cmNotifModal" class="cm-notif-modal cm-notif-hide" role="dialog" aria-modal="true">' +
        '<div class="cm-notif-modal-card">' +
        '<div class="cm-notif-modal-head">' +
        '<span id="cmNotifModalTipo" class="cm-notif-tipo"></span>' +
        '<button type="button" id="cmNotifModalFechar" class="cm-notif-modal-x">&times;</button>' +
        '</div>' +
        '<h5 id="cmNotifModalTitulo"></h5>' +
        '<div id="cmNotifModalCorpo" class="cm-notif-modal-corpo"></div>' +
        '<div class="cm-notif-modal-foot">' +
        '<span id="cmNotifModalPos" class="cm-notif-modal-pos"></span>' +
        '<button type="button" id="cmNotifModalOk" class="btn btn-primary btn-sm">Entendi</button>' +
        '</div>' +
        '</div>' +
        '</div>'
    )
    $('#cmNotifModalFechar, #cmNotifModalOk').on('click', function () {
      avancarModal()
    })
  }

  function abrirModal(itens, idx, marcar) {
    garantirModal()
    modalFila = itens || []
    modalIdx = idx || 0
    if (!modalFila.length) {
      $('#cmNotifModal').addClass('cm-notif-hide')
      return
    }
    mostrarModalItem(marcar !== false)
  }

  function mostrarModalItem(marcar) {
    var n = modalFila[modalIdx]
    if (!n) {
      $('#cmNotifModal').addClass('cm-notif-hide')
      return
    }
    $('#cmNotifModalTipo').text((n.tipo || 'info').toUpperCase()).attr('data-tipo', n.tipo || 'info')
    $('#cmNotifModalTitulo').text(n.titulo || '')
    $('#cmNotifModalCorpo').text(n.corpo || '')
    $('#cmNotifModalPos').text(modalFila.length > 1 ? modalIdx + 1 + ' / ' + modalFila.length : '')
    $('#cmNotifModal').removeClass('cm-notif-hide')
    if (marcar && !n.vistoEm) {
      marcarVisto(n.idNotificacao)
      n.vistoEm = new Date().toISOString()
    }
  }

  function avancarModal() {
    modalIdx++
    if (modalIdx >= modalFila.length) {
      $('#cmNotifModal').addClass('cm-notif-hide')
      atualizarBadge()
      renderLista()
      return
    }
    mostrarModalItem(true)
  }

  function marcarVisto(id) {
    return post(BASE.visto, { idNotificacao: id }).fail(function () {})
  }

  function atualizarBadge() {
    var n = lista.filter(function (x) {
      return !x.vistoEm
    }).length
    setBadge(n)
  }

  function carregar(opts) {
    opts = opts || {}
    garantirUI()
    garantirModal()
    post(BASE.minhas, {})
      .done(function (r) {
        lista = r && r.status === 'OK' && Array.isArray(r.dados) ? r.dados : []
        atualizarBadge()
        if (opts.popup !== false) {
          var novas = lista.filter(function (x) {
            return !x.vistoEm
          })
          if (novas.length) abrirModal(novas, 0, true)
        }
      })
      .fail(function () {
        lista = []
        setBadge(0)
      })
  }

  function ehTelaInicial() {
    var path = (location.pathname || '').replace(/\/+$/, '') || '/'
    return (
      path === '/carregar-menu-principal' ||
      path === '/carregar-menu-confvision' ||
      path === '/home'
    )
  }

  function init(opts) {
    opts = opts || {}
    if (opts.software) SOFTWARE = opts.software
    if (opts.urls) {
      BASE.minhas = opts.urls.minhas || BASE.minhas
      BASE.contagem = opts.urls.contagem || BASE.contagem
      BASE.visto = opts.urls.visto || BASE.visto
      BASE.lido = opts.urls.lido || BASE.lido
    }
    $(function () {
      // Popup automatico so na tela inicial; nas demais so sino/badge
      var popup = opts.popup
      if (popup === undefined) popup = ehTelaInicial()
      carregar({ popup: !!popup })
    })
  }

  return { init: init, carregar: carregar }
})()
