/* Paginacao por scroll: 100 registros por vez */
var FP_PAGE_SIZE = 100

function fpPagExtrairDados(r) {
    if (!r) return []
    if (Array.isArray(r)) return r
    if (Array.isArray(r.dados)) return r.dados
    return []
}

function fpPagMeta(r) {
    return {
        total: (r && r.total != null) ? r.total : null,
        hasMore: !!(r && r.hasMore),
        limit: (r && r.limit != null) ? r.limit : FP_PAGE_SIZE,
        offset: (r && r.offset != null) ? r.offset : 0
    }
}

/**
 * Scroll infinito para tabelas de relatorio.
 * opcoes: {
 *   url, method, getPayload(offset),
 *   tbodySel, renderRows(dados, primeira),
 *   onEmpty(), onTotal(total), onDone(primeira),
 *   scrollTarget (elemento ou window; padrao window),
 *   showLoading (bool, padrao true na 1a pagina)
 * }
 */
function fpScrollPaginacao(opcoes) {
    var state = {
        offset: 0,
        total: null,
        hasMore: true,
        loading: false,
        ativo: true
    }

    var scrollEl = opcoes.scrollTarget || window

    function desvincularScroll() {
        $(scrollEl).off('scroll.fpPag')
    }

    function vincularScroll() {
        desvincularScroll()
        $(scrollEl).on('scroll.fpPag', function () {
            if (!state.ativo || state.loading || !state.hasMore) return
            var el = scrollEl === window ? document.documentElement : scrollEl
            var scrollTop = scrollEl === window ? window.scrollY : el.scrollTop
            var clientH = scrollEl === window ? window.innerHeight : el.clientHeight
            var scrollH = el.scrollHeight
            if (scrollTop + clientH >= scrollH - 180) {
                carregar(false)
            }
        })
    }

    function carregar(primeira) {
        if (!state.ativo || state.loading) return
        if (!primeira && !state.hasMore) return

        state.loading = true
        if (primeira) {
            state.offset = 0
            state.hasMore = true
            state.total = null
            $(opcoes.tbodySel).empty()
            if (opcoes.showLoading !== false) boxProcessando()
        } else if (opcoes.onLoadingMore) {
            opcoes.onLoadingMore()
        }

        var payload = opcoes.getPayload(state.offset) || {}
        payload.limit = FP_PAGE_SIZE
        payload.offset = state.offset

        $.ajax({
            url: opcoes.url,
            method: opcoes.method || 'POST',
            contentType: 'application/json; charset=utf-8',
            dataType: 'json',
            data: JSON.stringify(payload)
        }).fail(function (e) {
            state.loading = false
            if (primeira) boxFechar()
            if (opcoes.onFail) opcoes.onFail(e)
            else console.log(e)
        }).done(function (r) {
            state.loading = false
            if (primeira) boxFechar()

            if (r && r.status === 'Vazio') {
                state.hasMore = false
                if (primeira && opcoes.onEmpty) opcoes.onEmpty()
                if (opcoes.onDone) opcoes.onDone(primeira, r)
                return
            }

            var dados = fpPagExtrairDados(r)
            var meta = fpPagMeta(r)

            if (meta.total != null) state.total = meta.total
            if (dados.length > 0 && opcoes.renderRows) {
                opcoes.renderRows(dados, primeira)
            } else if (primeira && dados.length === 0 && opcoes.onEmpty) {
                opcoes.onEmpty()
            }

            state.offset += dados.length
            if (meta.hasMore) {
                state.hasMore = true
            } else if (state.total != null) {
                state.hasMore = state.offset < state.total
            } else {
                state.hasMore = dados.length >= FP_PAGE_SIZE
            }

            if (opcoes.onTotal) opcoes.onTotal(state.total != null ? state.total : state.offset)
            if (opcoes.onDone) opcoes.onDone(primeira, r)
        })
    }

    vincularScroll()

    return {
        reset: function () { carregar(true) },
        carregar: carregar,
        destroy: function () {
            state.ativo = false
            desvincularScroll()
        },
        getOffset: function () { return state.offset },
        getTotal: function () { return state.total }
    }
}
