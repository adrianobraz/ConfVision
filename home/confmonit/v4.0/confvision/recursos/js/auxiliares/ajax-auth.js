(function () {
    function instalarAjaxAuth() {
        if (typeof $ === 'undefined' || $.ajaxSetup._confmonitAuth) return
        $.ajaxSetup({
            beforeSend: function (xhr) {
                const token = localStorage.getItem('token')
                if (token) {
                    xhr.setRequestHeader('Authorization', 'Bearer ' + token)
                }
            }
        })
        $.ajaxSetup._confmonitAuth = true
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', instalarAjaxAuth)
    } else {
        instalarAjaxAuth()
    }
})()
