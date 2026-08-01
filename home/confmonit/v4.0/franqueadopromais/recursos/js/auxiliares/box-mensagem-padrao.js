/**
 * Extrai mensagem amigável de erro Xano/proxy (remove "erro xano pro HTTP 500: {...}").
 * Aceita xhr jQuery, string ou objeto.
 */
function fpExtrairMsgErro(input) {
    var raw = ''
    if (input == null) return 'Não foi possível concluir a operação.'
    if (typeof input === 'string') {
        raw = input
    } else if (input.responseJSON && (input.responseJSON.message || input.responseJSON.status || input.responseJSON.erro)) {
        var rj = input.responseJSON
        raw = rj.message || rj.status || rj.erro || ''
    } else if (input.responseText) {
        raw = input.responseText
    } else if (input.message) {
        raw = input.message
    } else if (input.status) {
        raw = String(input.status)
    } else {
        raw = String(input)
    }

    raw = String(raw || '').trim()
    if (!raw) return 'Não foi possível concluir a operação.'

    try {
        var j = JSON.parse(raw)
        if (j && typeof j === 'object') {
            if (j.message) return String(j.message).trim()
            if (j.erro) raw = String(j.erro)
            else if (j.status) raw = String(j.status)
        }
    } catch (e) { /* não é JSON puro */ }

    raw = raw.replace(/^Erro:\s*/i, '')
    raw = raw.replace(/^erro\s*xano\s*pro:\s*/i, '')
    raw = raw.replace(/^xano\s*pro\s*HTTP\s*\d+:\s*/i, '')

    var m = raw.match(/"message"\s*:\s*"((?:\\.|[^"\\])*)"/)
    if (m && m[1]) {
        try {
            return JSON.parse('"' + m[1] + '"')
        } catch (e2) {
            return m[1].replace(/\\"/g, '"')
        }
    }

    raw = raw.trim()
    return raw || 'Não foi possível concluir a operação.'
}

/** Popup padrão (SweetAlert2) — substitui alert(). tipo: sucesso|erro|aviso|info */
function fpMsg(mensagem, tipo) {
    tipo = tipo || 'info'
    var icon = 'info'
    var title = 'Informação'
    if (tipo === 'erro' || tipo === 'error') {
        icon = 'error'
        title = 'Atenção'
    } else if (tipo === 'sucesso' || tipo === 'success' || tipo === 'ok') {
        icon = 'success'
        title = 'Sucesso'
    } else if (tipo === 'aviso' || tipo === 'warning' || tipo === 'atencao') {
        icon = 'warning'
        title = 'Atenção'
    }

    if (typeof Swal === 'undefined') {
        window.alert(String(mensagem || ''))
        return
    }

    Swal.fire({
        position: 'center',
        icon: icon,
        title: title,
        text: String(mensagem || ''),
        confirmButtonColor: '#0d6efd',
        confirmButtonText: 'OK',
        customClass: {
            popup: 'fp-swal-popup'
        }
    })
}

function fpMsgErro(xhrOuTexto) {
    fpMsg(fpExtrairMsgErro(xhrOuTexto), 'erro')
}

function fpMsgSucesso(mensagem) {
    fpMsg(mensagem, 'sucesso')
}

function fpMsgAviso(mensagem) {
    fpMsg(mensagem, 'aviso')
}

/** Confirmação (substitui confirm()). Retorna Promise<boolean>. */
function fpConfirmar(mensagem, opts) {
    opts = opts || {}
    var texto = String(mensagem || '')
    if (typeof Swal === 'undefined') {
        return Promise.resolve(window.confirm(texto))
    }
    return Swal.fire({
        position: 'center',
        icon: opts.icon || 'question',
        title: opts.title || 'Confirmar',
        html: texto.replace(/\n/g, '<br>'),
        showCancelButton: true,
        confirmButtonColor: '#0d6efd',
        cancelButtonColor: '#6c757d',
        confirmButtonText: opts.confirmText || 'Sim',
        cancelButtonText: opts.cancelText || 'Cancelar',
        customClass: {
            popup: 'fp-swal-popup'
        }
    }).then(function (r) {
        return !!(r && r.isConfirmed)
    })
}

/**
 * Select em popup (substitui prompt numérico).
 * opcoes: [{ value, label }] — value retornado no callback.
 */
function fpEscolherOpcao(titulo, opcoes, onEscolha) {
    opcoes = opcoes || []
    if (!opcoes.length) {
        fpMsgAviso('Nenhuma opção disponível.')
        return
    }
    if (typeof Swal === 'undefined') {
        var lista = opcoes.map(function (o, i) {
            return (i + 1) + ') ' + o.label
        }).join('\n')
        var escolha = window.prompt((titulo || 'Escolha') + '\n\n' + lista + '\n\nDigite o número:', '1')
        if (escolha == null) return
        var idx = parseInt(escolha, 10) - 1
        if (isNaN(idx) || idx < 0 || idx >= opcoes.length) {
            fpMsgAviso('Opção inválida.')
            return
        }
        onEscolha(opcoes[idx].value)
        return
    }
    var inputOptions = {}
    opcoes.forEach(function (o, i) {
        inputOptions[String(i)] = o.label
    })
    Swal.fire({
        position: 'center',
        title: titulo || 'Escolha',
        input: 'select',
        inputOptions: inputOptions,
        inputValue: '0',
        showCancelButton: true,
        confirmButtonColor: '#0d6efd',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Continuar',
        cancelButtonText: 'Cancelar',
        customClass: {
            popup: 'fp-swal-popup'
        }
    }).then(function (r) {
        if (!r.isConfirmed) return
        var i = parseInt(r.value, 10)
        if (isNaN(i) || i < 0 || i >= opcoes.length) {
            fpMsgAviso('Opção inválida.')
            return
        }
        onEscolha(opcoes[i].value)
    })
}

function boxSenhaAlterada(texto) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Senhar resetada com sucesso',
        text: `Aterada para: ${texto}`,
        confirmButtonColor: '#3085d6',
    })
}

function boxInseridoSucesso(idAtribuido) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Inserido com Sucesso',
        text: `ID atribuido: ${idAtribuido}`,
        confirmButtonColor: '#3085d6',
    })
}

function boxConfirmarExcluir(item, url, id) {
    Swal.fire({
        //position: 'top',
        title: `Quer Realmente apagar ${item}?`,
        text: "Não poderar reverter essa ação!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#3085d6',
        cancelButtonColor: '#d33',
        confirmButtonText: 'Sim, pode apagar!'
    }).then((result) => {
        if (result.isConfirmed) {
            $.ajax({
                start: boxProcessando(),
                url: url,
                method: 'POST',
                data: JSON.stringify(
                    {
                        id: id
                    }
                )
            }).fail(function (erro) {
                boxErro(e)
            }).done(function (r) {
                    boxDeletadoSucesso()
            })
        }
    })
}

function boxDeletadoSucesso() {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Apagado com sucesso',
        //showConfirmButton: false,
        confirmButtonColor: '#3085d6',
        timer: 3000
    })
}

function boxAteradoSucesso() {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Alterado com Sucesso',
        confirmButtonColor: '#3085d6',
        timer: 3000
    })
}

function boxEmCostrucao(link) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Em Cosntrução',
        text: `Em breve novidades`,
        confirmButtonColor: '#3085d6',
        didClose: () => {
            window.location.href = `/${link}`;
        }
    })
}

function boxMesagemAtencaoPersonalizada(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
    })
}

function boxProcessando(texto = "") {
    const txt = (texto == "") ? 'Processando...' : texto
    Swal.fire({
        //position: 'top-end',
        //icon: 'success',
        title: txt,
        showConfirmButton: false,
        didOpen: () => {
            Swal.showLoading()
        }
    })
}

function boxErro(erro) {
    fpMsgErro(erro)
}

function boxMensagemAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 2000
    })
}

function boxSucessoAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'success',
        title: 'Sucesso !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 2000
    })
}

function boxAdvertenciaAuto(mensagem) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 3000,
    })
}

function boxAdvertenciaCampoAuto(mensagem, idCampo) {
    Swal.fire({
        position: 'top',
        icon: 'warning',
        title: 'Atenção !!!',
        text: mensagem,
        confirmButtonColor: '#3085d6',
        timer: 3000,
        didClose: () => {
            $(idCampo).focus()
        }

    })
}

function boxFechar() {
    Swal.close()
}