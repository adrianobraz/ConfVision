/* Feedback — icone bug/melhoria/ideia na navbar */
var CmFeedback = (function () {
  var SOFTWARE = 'franqueadopro'
  var URL_ENVIAR = '/feedback/enviar'
  var tipoSel = 'bug'
  var enviando = false

  function garantirUI() {
    if ($('#cmFeedbackBtn').length) return
    var $mount = $('#cmFeedbackMount').first()
    if (!$mount.length) $mount = $('#cmNotifMount').first()
    if (!$mount.length) $mount = $('.fp-navbar-actions').first()
    if (!$mount.length) $mount = $('.amb-navbar-actions').first()
    if (!$mount.length) return

    var btnClass = 'btn fp-icon-btn cm-feedback-btn'
    if ($mount.closest('.amb-navbar-actions').length || $mount.hasClass('amb-navbar-actions')) {
      btnClass = 'btn btn-outline btn-sm cm-feedback-btn'
    }

    var $wrap = $(
      '<span class="cm-feedback-wrap">' +
        '<button type="button" id="cmFeedbackBtn" class="' +
        btnClass +
        '" title="Reportar bug, melhoria ou ideia" aria-label="Feedback">' +
        '<i class="bi bi-bug-fill"></i>' +
        '</button>' +
        '</span>'
    )

    if ($mount.is('#cmNotifMount') || $mount.is('#cmFeedbackMount')) {
      $mount.before($wrap)
    } else {
      $mount.prepend($wrap)
    }

    $('#cmFeedbackBtn').on('click', function (e) {
      e.preventDefault()
      e.stopPropagation()
      abrirModal()
    })
  }

  function garantirModal() {
    if ($('#cmFeedbackModal').length) return
    $('body').append(
      '<div id="cmFeedbackModal" class="cm-feedback-modal cm-feedback-hide" role="dialog" aria-modal="true">' +
        '<div class="cm-feedback-card">' +
        '<h5>Enviar feedback</h5>' +
        '<span class="cm-feedback-label">Tipo</span>' +
        '<div class="cm-feedback-tipos">' +
        '<button type="button" class="cm-feedback-tipo active" data-tipo="bug">Bug</button>' +
        '<button type="button" class="cm-feedback-tipo" data-tipo="melhoria">Melhoria</button>' +
        '<button type="button" class="cm-feedback-tipo" data-tipo="ideia">Ideia</button>' +
        '</div>' +
        '<span class="cm-feedback-label">Descricao</span>' +
        '<textarea id="cmFeedbackDesc" placeholder="Descreva o bug, melhoria ou ideia..." maxlength="4000"></textarea>' +
        '<div class="cm-feedback-foot">' +
        '<button type="button" id="cmFeedbackCancelar">Cancelar</button>' +
        '<button type="button" class="cm-feedback-enviar" id="cmFeedbackEnviar">Enviar</button>' +
        '</div>' +
        '</div>' +
        '</div>'
    )

    $('#cmFeedbackModal').on('click', function (e) {
      if (e.target === this) fecharModal()
    })
    $('#cmFeedbackCancelar').on('click', fecharModal)
    $('#cmFeedbackEnviar').on('click', enviar)
    $(document).on('click', '.cm-feedback-tipo', function () {
      tipoSel = $(this).data('tipo')
      $('.cm-feedback-tipo').removeClass('active')
      $(this).addClass('active')
    })
  }

  function abrirModal() {
    garantirModal()
    tipoSel = 'bug'
    $('.cm-feedback-tipo').removeClass('active')
    $('.cm-feedback-tipo[data-tipo="bug"]').addClass('active')
    $('#cmFeedbackDesc').val('')
    $('#cmFeedbackModal').removeClass('cm-feedback-hide')
    setTimeout(function () {
      $('#cmFeedbackDesc').focus()
    }, 50)
  }

  function fecharModal() {
    $('#cmFeedbackModal').addClass('cm-feedback-hide')
  }

  function enviar() {
    if (enviando) return
    var desc = String($('#cmFeedbackDesc').val() || '').trim()
    if (desc.length < 5) {
      if (typeof Swal !== 'undefined') {
        Swal.fire({ icon: 'warning', title: 'Descricao curta', text: 'Escreva pelo menos 5 caracteres.' })
      } else {
        alert('Escreva pelo menos 5 caracteres.')
      }
      return
    }

    enviando = true
    $('#cmFeedbackEnviar').prop('disabled', true)

    $.ajax({
      url: URL_ENVIAR,
      method: 'POST',
      contentType: 'application/json',
      data: JSON.stringify({
        tipo: tipoSel,
        descricao: desc,
        urlPagina: location.pathname + location.search,
      }),
    })
      .done(function () {
        fecharModal()
        if (typeof Swal !== 'undefined') {
          Swal.fire({ icon: 'success', title: 'Obrigado!', text: 'Recebemos seu feedback.', timer: 2200, showConfirmButton: false })
        } else {
          alert('Obrigado! Recebemos seu feedback.')
        }
      })
      .fail(function () {
        if (typeof Swal !== 'undefined') {
          Swal.fire({ icon: 'error', title: 'Erro', text: 'Nao foi possivel enviar. Tente novamente.' })
        } else {
          alert('Nao foi possivel enviar. Tente novamente.')
        }
      })
      .always(function () {
        enviando = false
        $('#cmFeedbackEnviar').prop('disabled', false)
      })
  }

  function init(opts) {
    opts = opts || {}
    if (opts.software) SOFTWARE = opts.software
    if (opts.url) URL_ENVIAR = opts.url
    garantirUI()
    garantirModal()
  }

  return { init: init }
})()
