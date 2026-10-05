/* Badge de tipo de gravação: contínua, movimento e timelapse */

function normalizarTipoSegmento(seg) {
    if (!seg) return 'continua'
    const t = String(seg.tipo || '').trim().toLowerCase()
    if (t === 'movimento') return 'movimento'
    if (t === 'timelapse') return 'timelapse'
    return 'continua'
}

function tipoFromCamera(cam) {
    if (!cam) return 'continua'
    if (cam.grava_timelapse) return 'timelapse'
    if (cam.grava_movimento) return 'movimento'
    return 'continua'
}

function labelTipoGravacao(tipo) {
    if (tipo === 'movimento') return 'Gravação por movimento'
    if (tipo === 'timelapse') return 'Timelapse Inteligente'
    return 'Gravação contínua'
}

function htmlBadgeTipo(tipo, extraClass) {
    const extra = extraClass ? ' ' + extraClass : ''
    let icon, cls, title
    if (tipo === 'movimento') {
        icon  = 'bi-activity'
        cls   = 'cv-tipo-movimento'
        title = 'Gravação por movimento'
    } else if (tipo === 'timelapse') {
        icon  = 'bi-fast-forward-fill'
        cls   = 'cv-tipo-timelapse'
        title = 'Timelapse Inteligente'
    } else {
        icon  = 'bi-record-circle'
        cls   = 'cv-tipo-continua'
        title = 'Gravação contínua'
    }
    return (
        '<span class="cv-tipo-badge ' + cls + extra + '" title="' + title + '" aria-label="' + title + '">' +
        '<i class="bi ' + icon + '"></i></span>'
    )
}

function corTimelineTipo(tipo) {
    if (tipo === 'movimento') return '#f59e0b'
    if (tipo === 'timelapse') return '#7c3aed'
    return '#16a34a'
}

function valorTimelineTipo(seg) {
    const t = normalizarTipoSegmento(seg)
    if (t === 'movimento') return 2
    if (t === 'timelapse') return 1.5
    return 1
}
