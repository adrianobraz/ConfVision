function go() {
    document.getElementById("nvoipconectarServ").click()
}

function inicia() {
    setTimeout(function() {
        go()
    }, 1000);
    document.getElementById('ua').style.display = "none";
}
inicia();
$(".nvoipmove").draggable({
    containment: "window"
});
var seconds;
var nvoiptemporizador = true;
var nwp_time;
var stopT;
var horasAtual;

function nvoiptempo() {
    horasAtual = new Date().getTime();
    horasAtual = (horasAtual - seconds) / 1000;
    segundos = horasAtual;
    segundo = segundos % 60;
    minutos = segundos / 60;
    minuto = minutos % 60;
    hora = minutos / 60;
    if (segundo < 10) {
        segundo = "0" + parseInt(segundo)
    } else {
        segundo = parseInt(segundo)
    }
    if (minuto < 10) {
        minuto = "0" + parseInt(minuto)
    } else {
        minuto = parseInt(minuto)
    }
    if (hora < 1) {
        hms = minuto + ":" + segundo
    } else {
        hms = Math.abs(hora) + ":" + minuto + ":" + segundo
    }
    document.getElementById('nvoiptempo').innerHTML = hms
}

function abreTeclado() {
    document.getElementById('nvoipbotoes2').hidden = true;
    document.getElementById('tecladoLig').hidden = false;
    document.getElementById('nvoipfechaTecla').hidden = false;
    document.getElementById('teclado').hidden = true;
}

function nvoipfechaTeclado() {
    document.getElementById('tecladoLig').hidden = true;
    document.getElementById('nvoipbotoes2').hidden = false;
    document.getElementById('nvoipfechaTecla').hidden = true;
    document.getElementById('teclado').hidden = false;
}

function nvoipdigita(numero) {
    var aux;
    aux = document.getElementById('nvoipua-uri').value;
    aux = aux + numero;
    document.getElementById('nvoipua-uri').value = aux
}

function minimiza() {
    document.getElementById('ua').style.display = "none";
    document.getElementById('nvoipdesminimiza').style.display = "block";
}

function nvoipdesminimiza() {
    document.getElementById('ua').style.display = "block";
    document.getElementById('nvoipdesminimiza').style.display = "none";
}

function dtmf(numero) {
    document.getElementById('inputDTMF').value = numero;
    document.getElementById('enviadorDTMF').click()
}

function rodarScript() {
    nwp_time = setInterval("nvoiptempo()", 1000)
}

function reiniciarScript() {
    seconds = new Date().getTime();
    document.getElementById('nvoiptempo').innerHTML = '00:00';
    rodarScript()
}

function pararScript() {
    clearInterval(nwp_time);
    clearInterval(stopT)
}

function apagaInput() {
    var aux = document.getElementById('nvoipua-uri').value;
    document.getElementById('nvoipua-uri').value = aux.substring(0, aux.length - 1)
}
var ponto = '.';
var timeout = false;
var velocidade = 500;

function animaPonto() {
    timeout = setTimeout('animaPonto()', velocidade);
    document.getElementById('nvoipalvo').innerHTML = ponto;
    if (ponto == '...') {
        ponto = '.'
    } else {
        ponto += '.'
    }
}
animaPonto();
var nvoiptempofechar;

function fecharLigacao() {
    nvoipfecharWebPhone.click()
}

function rodarScriptLigacao() {
    nvoiptempofechar = setInterval("fecharLigacao()", 3000)
}

function pararScriptLigacao() {
    clearInterval(nvoiptempofechar)
}
audio = document.getElementById('audio');

function playSound() {
    audio.play()
}

function stopSound() {
    audio.pause()
}
audio2 = document.getElementById('audio2');

function playSound2() {
    audio2.play()
}

function stopSound2() {
    audio2.pause()
}

var elements = {
    configForm: document.getElementById('nvoipconfig-form'),
    uaStatus: document.getElementById('nvoipua-status'),
    registerButton: document.getElementById('nvoipua-register'),
    newSessionForm: document.getElementById('nvoipnew-session-form'),
    inviteButton: document.getElementById('nvoipua-invite-submit'),
    messageButton: document.getElementById('nvoipua-message-submit'),
    uaURI: document.getElementById('nvoipua-uri'),
    sessionList: document.getElementById('nvoipsession-list'),
    sessionTemplate: document.getElementById('nvoipsession-template'),
};

config.uri = 'sip:' + config.authorizationUser + '@54.233.253.44';
config.register = true;
config.registerOptions = {
    expires: 600
};
config.usePreloadedRoute = true;
config.userAgentString = 'Webphone Nvoip v0.15.11';
config.displayName = config.authorizationUser;
config.sessionDescriptionHandlerFactoryOptions = {
    constraints: {
            audio: true,
            video: false
        }
    };
config.transportOptions = {
    wsServers: [{scheme:"WSS", sipUri:"sip:app.nvoip.com.br:7443;transport=wss;lr", wsUri: 'wss://app.nvoip.com.br:7443', weight: 10}],
    traceSip: true
};
config.hackWssInTransport = true;
config.hackAllowUnregisteredOptionTags = true;
config.contactTransport = "wss";

var ua;
var sessionUIs = {};
elements.configForm.addEventListener('submit', function(e) {
    var form, i, l, name, value;
    e.preventDefault();
    form = elements.configForm;
    elements.uaStatus.innerHTML = 'Conectando...';
    ua = new SIP.UA(config);
    ua.on('connected', function() {
        elements.uaStatus.innerHTML = ' <img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/minimizar.png" style="height:6px; margin: auto; margin-left: 8px; cursor: pointer" onclick="minimiza();"> <img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/webphone.png" style="width:100px; display:flex; margin:auto"> <div class="conectando"></div>'
    });
    ua.on('registered', function() {
        elements.registerButton.innerHTML = 'Unregister';
        elements.uaStatus.innerHTML = '<img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/minimizar.png" style="height:6px; margin: auto; margin-left: 8px; cursor: pointer" onclick="minimiza();"> <img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/webphone.png" style="width:100px; display:flex; margin:auto"> <div class="nvoipconectado"></div>'
    });
    ua.on('unregistered', function() {
        elements.registerButton.innerHTML = 'Register';
        elements.uaStatus.innerHTML = '<img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/minimizar.png" style="height:6px; margin: auto; margin-left: 8px; cursor: pointer" onclick="minimiza();"> <img src="https://nvoipcom.s3-sa-east-1.amazonaws.com/public/webphone/v4/img/webphone.png" style="width:100px; display:flex; margin:auto"> <div class="conectando"></div>'
    });
    ua.on('invite', function(session) {
        nvoipdesminimiza();
    if(session.remoteIdentity.uri.normal.user == sessionStorage.getItem("userRamal")) {
                nvoipdesminimiza();
                session.accept();
    }
        createNewSessionUI(session.remoteIdentity.uri, session);
    });
    ua.on('message', function(message) {
        if (!sessionUIs[message.remoteIdentity.uri]) {
            createNewSessionUI(message.remoteIdentity.uri, null, message)
        }
    });
    document.body.className = 'started'
}, false);
elements.registerButton.addEventListener('click', function() {
    if (!ua) return;
    if (ua.isRegistered()) {
        ua.unregister()
    } else {
        ua.register()
    }
}, false);

function inviteSubmit(e) {
    e.preventDefault();
    e.stopPropagation();
    var video = false;
    var uri = elements.uaURI.value;
    elements.uaURI.value = '';
    if (!uri) return;
    var session = ua.invite(uri, {
        mediaConstraints: {
            audio: true,
            video: false
        }
    });
    var ui = createNewSessionUI(uri, session)
}
elements.inviteButton.addEventListener('click', inviteSubmit, false);
elements.newSessionForm.addEventListener('submit', inviteSubmit, false);
elements.messageButton.addEventListener('click', function(e) {
    e.preventDefault();
    e.stopPropagation();
    var uri = elements.uaURI.value;
    elements.uaURI.value = '';
    var ui = createNewSessionUI(uri)
}, false);


function createNewSessionUI(uri, session, message) {
    var numeroAtual = uri;
    var tpl = elements.sessionTemplate;
    var node = tpl.cloneNode(true);
    var sessionUI = {};
    var messageNode;
    var flagado;
    uri = session ? session.remoteIdentity.uri : SIP.Utils.normalizeTarget(uri, ua.configuration.hostport_params);
    var displayName = (session && session.remoteIdentity.displayName) || uri.user;
    if (!uri) {
        return
    }
    sessionUI.session = session;
    sessionUI.node = node;
    sessionUI.displayName = node.querySelector('.nvoipdisplay-name');
    sessionUI.uri = node.querySelector('.nvoipuri');
    sessionUI.green = node.querySelector('.green');
    if (session.remoteIdentity.uri.normal.user == sessionStorage.getItem("userRamal")) {
        sessionUI.green.style.display = 'none';
    }
    sessionUI.red = node.querySelector('.red');
    sessionUI.nvoipbotoes2 = node.querySelector('.nvoipbotoes2');
    sessionUI.nvoipespacamento1 = node.querySelector('.nvoipespacamento1');
    sessionUI.nvoipefetuandoLig = node.querySelector('.nvoipefetuandoLig');
    sessionUI.nvoipfecharWebPhone = node.querySelector('.nvoipfecharWebPhone');
    sessionUI.mute = node.querySelector('.mute');
    sessionUI.teclado = node.querySelector('.teclado');
    sessionUI.nvoipdev2 = node.querySelector('.nvoipdev2');
    sessionUI.nvoipespacoLig = node.querySelector('.nvoipespacoLig');
    sessionUI.nvoiprecebendoLig = node.querySelector('.nvoiprecebendoLig');
    sessionUI.unmute = node.querySelector('.unmute');
    sessionUI.dtmf = node.querySelector('.dtmf');
    sessionUI.dtmfInput = node.querySelector('.dtmf input[type="text"]');
    sessionUI.video = node.querySelector('video');
    sessionUIs[uri] = sessionUI;
    node.classList.remove('nvoiptemplate');
    sessionUI.displayName.textContent = displayName || uri.user;
    sessionUI.uri.textContent = '<' + uri + '>';
    sessionUI.green.addEventListener('click', function() {
        var session = sessionUI.session;
        /*** AQUI É ONDE DA HIDDEN NO BOTÃO VERDE */
        sessionUI.green.hidden = true;
        sessionUI.teclado.hidden = false;
        sessionUI.unmute.hidden = false;
        sessionUI.nvoipdev2.hidden = true;
        if (!session) {
            session = sessionUI.session = ua.invite(uri, {
                mediaConstraints: {
                    audio: true,
                    video: false
                }
            });
            playSound2();
            setUpListeners(session)
        } else if (session.accept && !session.startTime) {
            stopSound();
            document.getElementById("nvoipdesminimiza").className = "nvoipdesminimiza";
            session.accept({
                mediaConstraints: {
                    audio: true,
                    video: false
                }
            })
        } else if (session.hasAnswer != null && session.hasAnswer == false) {
            flagado = true
        }
    }, false);
    sessionUI.red.addEventListener('click', function() {
        var session = sessionUI.session;
        stopSound2();
        if (!session) {
            return
        } else if (session.startTime) {
            session.bye()
        } else if (session.reject) {
            stopSound();
            document.getElementById("nvoipdesminimiza").className = "nvoipdesminimiza";
            elements.sessionList.removeChild(node);
            session.reject();
            pararScript()
        } else if (session.cancel) {
            elements.sessionList.removeChild(node);
            pararScript();
            session.cancel()
        }
    }, false);
    sessionUI.mute.addEventListener('click', function() {
        var session = sessionUI.session;
        
        if (!session) {
            sessionUI.mute.hidden = true;
            sessionUI.unmute.hidden = false;
            var pc = session.sessionDescriptionHandler.peerConnection
            pc.getSenders().forEach((stream) => {
                stream.track.enabled = false
            })
            return
        } else {
            var pc = session.sessionDescriptionHandler.peerConnection
            pc.getSenders().forEach((stream) => {
                stream.track.enabled = false
            })
            sessionUI.mute.hidden = true;
            sessionUI.unmute.hidden = false
        }
    }, false);
    sessionUI.unmute.addEventListener('click', function() {
        var session = sessionUI.session;
        if (!session) {
            sessionUI.mute.hidden = false;
            sessionUI.unmute.hidden = true;
            var pc = session.sessionDescriptionHandler.peerConnection
            pc.getSenders().forEach((stream) => {
                stream.track.enabled = true
            })
            return
        } else {
            var pc = session.sessionDescriptionHandler.peerConnection
            pc.getSenders().forEach((stream) => {
                stream.track.enabled = true
            })
            sessionUI.mute.hidden = false;
            sessionUI.unmute.hidden = true
        }
    }, false);
    sessionUI.nvoipfecharWebPhone.addEventListener('click', function() {
        elements.sessionList.removeChild(node);
        var session = sessionUI.session;
        pararScript();
        if (!session) {
            elements.sessionList.removeChild(node);
            return
        } else if (session.startTime) {
            session.bye();
            elements.sessionList.removeChild(node)
        } else if (session.reject) {
            session.reject();
            elements.sessionList.removeChild(node)
        } else if (session.cancel) {
            if (flagado == true) {
                var session2 = ua.invite(numeroAtual, {
                    mediaConstraints: {
                        audio: true,
                        video: false
                    }
                });
                var ui = createNewSessionUI(numeroAtual, session2)
            }
        }
    }, false);
    sessionUI.dtmf.addEventListener('submit', function(e) {
        e.preventDefault();
        var value = sessionUI.dtmfInput.value;
        if (value === '' || !session) return;
        sessionUI.dtmfInput.value = '';
        if (['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '#'].indexOf(value) > -1) {
            session.dtmf(value)
        }
    });
    if (session && !session.accept) {
        sessionUI.green.disabled = true;
        sessionUI.teclado.hidden = true;
        sessionUI.unmute.hidden = true;
        sessionUI.green.innerHTML = '...';
        sessionUI.red.innerHTML = 'Cancel';
        sessionUI.nvoipefetuandoLig.hidden = false;
        playSound2()
    } else if (!session) {
        sessionUI.red.disabled = true;
        sessionUI.green.innerHTML = 'Invite';
        sessionUI.red.innerHTML = '...'
    } else {
        sessionUI.green.innerHTML = 'Accept';
        sessionUI.red.innerHTML = 'Reject'
    }
    sessionUI.dtmfInput.disabled = true;

    function setUpListeners(session) {
        sessionUI.red.disabled = false;
        if (session.accept) {
            sessionUI.green.disabled = false;
            sessionUI.mute.hidden = true;
            sessionUI.unmute.hidden = true;
            sessionUI.teclado.hidden = true;
            sessionUI.green.hidden = false;
            sessionUI.green.innerHTML = 'Accept';
            sessionUI.red.innerHTML = 'Reject';
            sessionUI.nvoiprecebendoLig.hidden = false;
            playSound();
            document.getElementById("nvoipdesminimiza").className = "primeira teste nvoipdesminimiza"
        } else {
            sessionUI.green.innerHMTL = '...';
            sessionUI.red.innerHTML = 'Cancel'
        }
        session.on('accepted', function() {
            stopSound();
            document.getElementById("nvoipdesminimiza").className = "nvoipdesminimiza";
            reiniciarScript();
            stopSound2();
            
            sessionUI.green.disabled = true;
            sessionUI.green.innerHTML = '...';
            sessionUI.teclado.hidden = false;
            sessionUI.unmute.hidden = false;
            sessionUI.nvoipdev2.hidden = true;
            sessionUI.red.innerHTML = 'Bye';
            sessionUI.dtmfInput.disabled = false;
            sessionUI.video.className = 'on';
            sessionUI.nvoipbotoes2.hidden = false;
            sessionUI.nvoipespacamento1.hidden = true;
            
            var element = sessionUI.video;
            var pc = session.sessionDescriptionHandler.peerConnection;
            var stream = new MediaStream();

            pc.getReceivers().forEach(function(receiver){
                var track = receiver.track;
                if (track) {
                    stream.addTrack(track);
                }
            });

            if (typeof element.srcObject !== 'undefined') {
                element.srcObject = stream
            } else if (typeof element.mozSrcObject !== 'undefined') {
                element.mozSrcObject = stream
            } else if (typeof element.src !== 'undefined') {
                element.src = URL.createObjectURL(stream)
            } else {
                console.log('Error attaching stream to element.')
            }

            element.autoplay = true;
            sessionUI.nvoipbotoes2.hidden = false;
            sessionUI.nvoipefetuandoLig.hidden = true;
            sessionUI.nvoiprecebendoLig.hidden = true
        });
        session.on('bye', function() {
            pararScript();
            sessionUI.green.disabled = false;
            sessionUI.red.disabled = true;
            sessionUI.dtmfInput.disable = true;
            sessionUI.green.innerHTML = 'Invite';
            sessionUI.red.innerHTML = '...';
            sessionUI.video.className = '';
            elements.sessionList.removeChild(node);
            delete sessionUI.session
        });
        session.on('failed', function(data) {
            if (data.method == "CANCEL") {
                sessionUI.nvoipespacoLig.hidden = true
            } else {
                sessionUI.nvoipespacoLig.hidden = false
            }
            stopSound2();
            sessionUI.green.disabled = false;
            sessionUI.red.disabled = true;
            sessionUI.dtmfInput.disable = true;
            sessionUI.green.innerHTML = 'Invite';
            sessionUI.red.innerHTML = '...';
            sessionUI.video.className = '';
            sessionUI.green.hidden = false;
            if (data.headers.Reason[0]) {
                var resposta = data.headers.Reason[0].raw;
                resposta = resposta.split("=");
                var numeroErro = resposta[1].split(";");
                numeroErro = numeroErro[0];
                sessionUI.nvoipefetuandoLig.hidden = true;
                sessionUI.nvoiprecebendoLig.hidden = true;
                sessionUI.nvoipespacoLig.hidden = false;
                stopSound();
                document.getElementById("nvoipdesminimiza").className = "nvoipdesminimiza";
                switch (numeroErro) {
                    case "487":
                        numeroErro = "Cancelada";
                        break;
                    case "16":
                        numeroErro = "Completada";
                        break;
                    case "31":
                        numeroErro = "Falha no destino";
                        break;
                    case "65":
                        numeroErro = "Falha no destino";
                        break;
                    case "88":
                        numeroErro = "Falha no destino";
                        break;
                    case "96":
                        numeroErro = "Falha no destino";
                        break;
                    case "38":
                        numeroErro = "Falha no destino";
                        break;
                    case "34":
                        numeroErro = "Falha no destino";
                        break;
                    case "41":
                        numeroErro = "Falha no destino";
                        break;
                    case "102":
                        numeroErro = "Fora de Area";
                        break;
                    case "44":
                        numeroErro = "Falha no destino";
                        break;
                    case "79":
                        numeroErro = "Falha no destino";
                        break;
                    case "604":
                        numeroErro = "Falha no destino";
                        break;
                    case "57":
                        numeroErro = "Falha no destino";
                        break;
                    case "111":
                        numeroErro = "Falha no destino";
                        break;
                    case "27":
                        numeroErro = "Falha no destino";
                        break;
                    case "31":
                        numeroErro = "Falha no destino";
                        break;
                    case "31":
                        numeroErro = "Falha no destino";
                        break;
                    case "127":
                        numeroErro = "Falha no destino";
                        break;
                    case "63":
                        numeroErro = "Falha no destino";
                        break;
                    case "20":
                        numeroErro = "Falha no destino";
                        break;
                    case "42":
                        numeroErro = "Limite ligacoes";
                        break;
                    case "28":
                        numeroErro = "Numero Invalido";
                        break;
                    case "3":
                        numeroErro = "Numero Invalido";
                        break;
                    case "31":
                        numeroErro = "Numero Invalido";
                        break;
                    case "31":
                        numeroErro = "Numero Invalido";
                        break;
                    case "22":
                        numeroErro = "Numero n�o existe";
                        break;
                    case "1":
                        numeroErro = "Numero n�o existe";
                        break;
                    case "17":
                        numeroErro = "Ocupado";
                        break;
                    case "21":
                        numeroErro = "Rejeitada";
                        break;
                    case "602":
                        numeroErro = "Sem resposta";
                        break;
                    case "19":
                        numeroErro = "Sem resposta";
                        break;
                    case "18":
                        numeroErro = "Sem resposta";
                        break;
                    case "31":
                        numeroErro = "Saldo insuficiente";
                        break;
                    case "606":
                        numeroErro = "Usuario Offline";
                        break;
                    default:
                        numeroErro = "Falha Desconhecida"
                }
                document.getElementById("nvoipstatusLig").innerHTML = numeroErro
            }
            pararScript();
            setTimeout(function() {
                sessionUI.nvoipfecharWebPhone.click()
            }, 3000)
        });
        session.on('refer', function(target) {
            session.bye();
            pararScript();
            createNewSessionUI(target, ua.invite(target, {
                mediaConstraints: {
                    audio: true,
                    video: false
                }
            }))
        })
    }
    if (session) {
        setUpListeners(session)
    }
    elements.sessionList.appendChild(node)
};