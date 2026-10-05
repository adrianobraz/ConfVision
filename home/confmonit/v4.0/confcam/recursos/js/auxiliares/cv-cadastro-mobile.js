/**
 * ConfVision cadastro mobile — lista vs formulário
 * Páginas: Clientes, Dispositivos, Setores
 */
(function () {
    var PATHS = [
        '/CarregarPaginaGerenciarCliente',
        '/carregar-gerenciar-dispositivo',
        '/carregar-gerenciar-setores-alarme'
    ]

    function pathOk() {
        var p = window.location.pathname || ''
        return PATHS.some(function (x) {
            return p === x || p.indexOf(x + '/') === 0
        })
    }

    function ehMobile() {
        return window.matchMedia('(max-width: 991.98px)').matches
    }

    function colunas() {
        var form = document.querySelector('.container-principal .bg-form')
        if (!form) return null
        var row = form.closest('.row')
        if (!row) return null
        var lista = null
        var kids = row.children
        for (var i = 0; i < kids.length; i++) {
            var col = kids[i]
            if (col === form) continue
            if (col.classList && (col.classList.contains('bg-form') === false)) {
                var cn = col.className || ''
                if (cn.indexOf('col') !== -1) lista = col
            }
        }
        return { form: form, lista: lista }
    }

    function mostrarLista() {
        document.body.classList.remove('cv-cadastro-form-mode')
        window.scrollTo(0, 0)
    }

    function mostrarForm() {
        document.body.classList.add('cv-cadastro-form-mode')
        window.scrollTo(0, 0)
        var nome = document.getElementById('nome')
        if (nome && typeof nome.focus === 'function') {
            try { nome.focus() } catch (e) { /* ignore */ }
        }
    }

    function garantirUi(cols) {
        if (!cols || !cols.lista) return

        cols.lista.classList.add('cv-cad-lista')
        cols.form.classList.add('cv-cad-form')

        if (!document.getElementById('cv-cad-bar-lista')) {
            var bar = document.createElement('div')
            bar.id = 'cv-cad-bar-lista'
            bar.className = 'cv-cad-bar-lista'
            bar.innerHTML =
                '<button type="button" id="cv-cad-btn-novo" class="cv-cad-btn-novo">' +
                '<i class="bi bi-plus-lg"></i> Novo' +
                '</button>'
            cols.lista.insertBefore(bar, cols.lista.firstChild)
        }

        if (!document.getElementById('cv-cad-bar-form')) {
            var barForm = document.createElement('div')
            barForm.id = 'cv-cad-bar-form'
            barForm.className = 'cv-cad-bar-form'
            barForm.innerHTML =
                '<button type="button" id="cv-cad-btn-voltar" class="cv-cad-btn-voltar">' +
                '<i class="bi bi-arrow-left"></i> Voltar à lista' +
                '</button>' +
                '<span id="cv-cad-form-titulo" class="cv-cad-form-titulo">Cadastro</span>'
            var formEl = cols.form.querySelector('form') || cols.form
            formEl.insertBefore(barForm, formEl.firstChild)
        }
    }

    function setTituloForm(editando) {
        var el = document.getElementById('cv-cad-form-titulo')
        if (el) el.textContent = editando ? 'Editar' : 'Novo'
    }

    function init() {
        if (!pathOk()) return
        var cols = colunas()
        if (!cols) return

        garantirUi(cols)
        document.body.classList.add('cv-cadastro-split')

        if (ehMobile()) mostrarLista()
        else document.body.classList.remove('cv-cadastro-form-mode')

        document.getElementById('cv-cad-btn-novo').addEventListener('click', function () {
            var limpar = document.getElementById('btn-limpar')
            if (limpar) limpar.click()
            setTituloForm(false)
            mostrarForm()
        })

        document.getElementById('cv-cad-btn-voltar').addEventListener('click', function () {
            var limpar = document.getElementById('btn-limpar')
            if (limpar) limpar.click()
            mostrarLista()
        })

        // Editar na grid → formulário
        document.addEventListener('click', function (e) {
            var btn = e.target.closest('button[tipo="editar"], button[tipo=editar], .btn[tipo="editar"]')
            if (!btn) return
            if (!ehMobile()) return
            setTituloForm(true)
            // Aguarda o AJAX preencher o form
            window.setTimeout(mostrarForm, 80)
            window.setTimeout(mostrarForm, 400)
        })

        // Após limpar/gravar com sucesso as páginas limpam o form —
        // no mobile, se o usuário não estiver abrindo "Novo", volta à lista
        // (Voltar já chama limpar + lista; Novo chama limpar sem voltar)
        var abrindoNovo = false
        var btnNovo = document.getElementById('cv-cad-btn-novo')
        if (btnNovo) {
            btnNovo.addEventListener('click', function () {
                abrindoNovo = true
                window.setTimeout(function () { abrindoNovo = false }, 500)
            }, true)
        }

        var btnLimpar = document.getElementById('btn-limpar')
        if (btnLimpar) {
            btnLimpar.addEventListener('click', function () {
                if (!ehMobile()) return
                if (abrindoNovo) return
                // Se limpar veio do botão Limpar do form (não do Voltar/Novo),
                // mantém no form vazio — só Voltar força lista.
            })
        }

        window.addEventListener('resize', function () {
            if (!ehMobile()) document.body.classList.remove('cv-cadastro-form-mode')
            else if (!document.body.classList.contains('cv-cadastro-form-mode')) mostrarLista()
        })
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function () {
            // Após scripts jQuery das páginas
            window.setTimeout(init, 0)
        })
    } else {
        window.setTimeout(init, 0)
    }
})()
