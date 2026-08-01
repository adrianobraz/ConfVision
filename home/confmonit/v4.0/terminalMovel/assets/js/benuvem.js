function benuvem_start() {
    benuvem_gerarToken(ret => {
        benuvem_logout(ret)
    })
}

function benuvem_gerarToken(next) {
    var formdata = new FormData();
    formdata.append("email", sessionStorage.getItem('login_benuvemLogin'))
    formdata.append("password", sessionStorage.getItem('login_benuvemSenha'));

    var requestOptions = {
        method: 'POST',
        body: formdata,
        redirect: 'follow'
    };

    fetch("https://app.benuvem.com.br/api/v1/auth/login", requestOptions)
        .then(response => response.json())
        .then(result => next(result.access_token))
        .catch(error => console.log('error', error));
}

function benuvem_atualizarToken(next) {
    var requestOptions = {
        method: 'POST',
        redirect: 'follow'
    };

    fetch("https://app.benuvem.com.br/api/v1/auth/refresh", requestOptions)
        .then(response => response.json())
        .then(result => next(result))
        .catch(error => console.log('error', error));
}

function benuvem_logout(param) {
    var requestOptions = {
        method: 'POST',
        redirect: 'follow',
        headers: {
            'Authorization': `Bearer ${param}`
        }
    };

    fetch("https://app.benuvem.com.br/api/v1/auth/logout", requestOptions)
        .then(response => response.json())
        .then(result => {
            //console.log(result)
        })
        .catch(error => console.log('error', error));
}

let benuvem_janela
function benuvem_visualizar(codFranq, particao, conta, setor, data) {
    // console.log(codFranq, particao, conta, setor, data)
    benuvem_gerarToken(token => {
        var formdata = new FormData();
        formdata.append("client_code", conta);
        formdata.append("partition", particao);
        formdata.append("company_code", codFranq);
        formdata.append("channels[]", setor);
        if (data != undefined) {formdata.append("date", data);}
        formdata.append("show_records_popup", false)

        var requestOptions = {
            method: 'POST',
            body: formdata,
            redirect: 'follow',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        };

        fetch("https://app.benuvem.com.br/api/v1/cameras/show", requestOptions)
            .then(response => response.json())
            .then(result => {
                benuvem_logout(token)

                if (result.success = true) {
                    if (benuvem_janela != undefined) { benuvem_janela.close() }
                    
                    benuvem_janela = window.open( result.url, '_blank', `location=no, menubar=no, titlebar=no, height=${screen.height}, width=${screen.width}, top=0, left=0`)
                } else {
                    msgErro("erro ao acessar a camera")
                }
            })
            .catch(error => console.log('error', error));
    })

}

function benuvem_finalizarEvento(codFranq, particao, conta, setor) {
    benuvem_gerarToken(token => {
        var formdata = new FormData();
        formdata.append("client_code", conta);
        formdata.append("partition", particao);
        formdata.append("company_code", codFranq);
        formdata.append("channels[]", setor);

        var requestOptions = {
            method: 'POST',
            body: formdata,
            redirect: 'follow',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        };

        fetch("https://app.benuvem.com.br/api/v1/events/close", requestOptions)
            .then(response => response.json())
            .then(result => {
                benuvem_logout(token)

                if (result.success = true) {
                   console.log("Evento benuvem finalizado com sucesso")
                } else {
                    msgErro("erro ao finalizar evento benuvem")
                }
            })
            .catch(error => console.log('error', error));
    })

}