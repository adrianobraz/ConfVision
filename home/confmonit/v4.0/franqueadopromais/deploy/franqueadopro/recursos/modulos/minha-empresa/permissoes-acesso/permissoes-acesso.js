var permIdUsuario = ''
var permNomeUsuario = ''

var PERM_AJAX = {
    contentType: 'application/json; charset=utf-8',
    dataType: 'json'
}

$(document).ready(function () {
    function iniciarTela() {
        if (localStorage.getItem('loginMaster') !== 'S') {
            boxMesagemAtencaoPersonalizada('Apenas o usuário master pode configurar permissões.')
            setTimeout(function () {
                window.location.href = '/carregar-menu-gestao'
            }, 2000)
            return
        }

        $(window).on('resize', ajustaTabela)
        ajustaTabela()

        $('#btn-limpar').on('click', permLimparArvore)
        $('#btn-gravar').on('click', permGravar)
        $('#pesquisar-responsavel').on('input', permFiltrarLista)

        permCarregarResponsaveis()
        permMontarArvore({})
    }

    if (localStorage.getItem('loginMaster') === 'S') {
        iniciarTela()
    } else if (typeof fpSyncMaster === 'function') {
        fpSyncMaster(function (master) {
            if (master === 'S') iniciarTela()
            else {
                boxMesagemAtencaoPersonalizada('Apenas o usuário master pode configurar permissões.')
                setTimeout(function () {
                    window.location.href = '/carregar-menu-gestao'
                }, 2000)
            }
        })
    } else {
        boxMesagemAtencaoPersonalizada('Apenas o usuário master pode configurar permissões.')
        setTimeout(function () {
            window.location.href = '/carregar-menu-gestao'
        }, 2000)
    }
})

function permErroMsg(xhr, padrao) {
    var msg = padrao
    try {
        if (xhr && xhr.responseJSON) {
            if (xhr.responseJSON.status) {
                msg = String(xhr.responseJSON.status).replace(/^Erro:\s*/i, '')
            } else if (xhr.responseJSON.message) {
                msg = xhr.responseJSON.message
            }
        }
    } catch (e) { /* ignore */ }
    boxErro(msg)
}

function permExtrairLista(r) {
    if (Array.isArray(r)) return r
    if (r && Array.isArray(r.dados)) return r.dados
    if (r && r.dados && Array.isArray(r.dados.items)) return r.dados.items
    return []
}

function permCarregarResponsaveis() {
    $.ajax($.extend({}, PERM_AJAX, {
        url: '/usuariosCarregarTabela',
        method: 'POST',
        data: JSON.stringify({ idVinculo: localStorage.getItem('idFranqueado') })
    })).fail(function (xhr) {
        permErroMsg(xhr, 'Erro ao carregar responsáveis')
    }).done(function (r) {
        $('#lista-responsaveis').empty()
        if (r.status === 'Vazio') return
        r.dados.forEach(function (u) {
            if (u.master === 'S') return
            $('#lista-responsaveis').append(
                '<tr data-id="' + u.idUsuario + '" data-nome="' + permEscapeAttr(u.nome) + '">' +
                '<td class="text-start">' + permEscapeHtml(u.nome) + '</td>' +
                '<td class="text-start">' + permEscapeHtml(u.nick) + '</td>' +
                '</tr>'
            )
        })
        $('#lista-responsaveis tr').on('click', function () {
            permSelecionarUsuario($(this).attr('data-id'), $(this).attr('data-nome'))
        })
    })
}

function permEscapeHtml(txt) {
    return String(txt || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
}

function permEscapeAttr(txt) {
    return permEscapeHtml(txt)
}

function permFiltrarLista() {
    var termo = ($('#pesquisar-responsavel').val() || '').toLowerCase()
    $('#lista-responsaveis tr').each(function () {
        var txt = $(this).text().toLowerCase()
        $(this).toggle(txt.indexOf(termo) !== -1)
    })
}

function permSelecionarUsuario(id, nome) {
    permIdUsuario = id
    permNomeUsuario = nome
    $('#nome-selecionado').text(nome || '—')
    $('#lista-responsaveis tr').removeClass('fp-row-ativo')
    $('#lista-responsaveis tr[data-id="' + id + '"]').addClass('fp-row-ativo')
    permCarregarPermissoes(id)
}

function permCarregarPermissoes(idUsuario) {
    $.ajax($.extend({}, PERM_AJAX, {
        url: '/permissoesListarByUsuario',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            idUsuario: idUsuario
        })
    })).fail(function (xhr) {
        permErroMsg(xhr, 'Erro ao carregar permissões')
    }).done(function (r) {
        var lista = permExtrairLista(r)
        var map = {}
        lista.forEach(function (p) {
            var k = p.chave_menu || p.chaveMenu
            if (k) map[k] = (p.liberado === 'S' || p.liberado === true)
        })
        permMontarArvore(map)
    })
}

function permSafeId(key) {
    return 'perm-' + String(key).replace(/\./g, '-')
}

function permIconeNivel(depth) {
    if (depth === 0) return 'bi-diagram-3-fill'
    if (depth === 1) return 'bi-folder2-open'
    return 'bi-dot'
}

function permMontarArvore(map) {
    $('#arvore-permissoes').html(permRenderNodes(FP_PERM_CATALOGO, map, 0))
    permBindArvore()
}

function permRenderNodes(nodes, map, depth) {
    var html = ''
    nodes.forEach(function (n) {
        var checked = map[n.key] ? 'checked' : ''
        var safeId = permSafeId(n.key)
        var hasChild = n.children && n.children.length
        html += '<div class="fp-perm-node fp-perm-depth-' + depth + (hasChild ? ' fp-perm-parent' : '') + '" data-key="' + n.key + '">'
        html += '<div class="fp-perm-row">'
        html += '<input type="checkbox" class="perm-chk" id="' + safeId + '" data-key="' + n.key + '" ' + checked + '>'
        html += '<label for="' + safeId + '">'
        html += '<i class="bi ' + permIconeNivel(depth) + ' fp-perm-icon"></i>'
        html += '<span class="fp-perm-label">' + permEscapeHtml(n.label) + '</span>'
        html += '</label>'
        html += '</div>'
        if (hasChild) {
            html += '<div class="fp-perm-children">' + permRenderNodes(n.children, map, depth + 1) + '</div>'
        }
        html += '</div>'
    })
    return html
}

function permBindArvore() {
    $('.perm-chk').off('change').on('change', function () {
        var marcado = $(this).is(':checked')
        var node = $(this).closest('.fp-perm-node')
        node.find('.perm-chk').prop('checked', marcado)
        if (marcado) permMarcarPais(node)
    })
}

function permMarcarPais(node) {
    var parent = node.parent().closest('.fp-perm-node')
    if (!parent.length) return
    parent.find('> .fp-perm-row .perm-chk').prop('checked', true)
    permMarcarPais(parent)
}

function permColetarMarcados() {
    var permissoes = []
    $('.perm-chk:checked').each(function () {
        permissoes.push({
            chave_menu: $(this).data('key'),
            liberado: 'S'
        })
    })
    return permissoes
}

function permLimparArvore() {
    $('.perm-chk').prop('checked', false)
}

function permGravar() {
    if (!permIdUsuario) {
        boxMesagemAtencaoPersonalizada('Selecione um responsável.')
        return
    }
    boxProcessando('Gravando permissões...')
    $.ajax($.extend({}, PERM_AJAX, {
        url: '/permissoesSalvar',
        method: 'POST',
        data: JSON.stringify({
            idFranqueado: localStorage.getItem('idFranqueado'),
            idUsuario: permIdUsuario,
            permissoes: permColetarMarcados()
        })
    })).fail(function (xhr) {
        Swal.close()
        permErroMsg(xhr, 'Erro ao gravar permissões')
    }).done(function () {
        Swal.close()
        boxAteradoSucesso()
    })
}
