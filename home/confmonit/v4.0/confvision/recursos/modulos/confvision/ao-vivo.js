$(document).ready(function () {
    confVisionAuthGuard()
    confVisionCarregarCabecalho()
    carregarCamerasLive()

    $('#sel-camera-live').on('change', function () {
        const id = $(this).val()
        if (id) {
            ConfVisionLivePlayer.start({
                video: '#player-hls',
                status: '#cv-live-status',
                cameraId: id
            })
        } else {
            ConfVisionLivePlayer.stop()
        }
    })
})

function carregarCamerasLive() {
    const preId = ($('#cfg-camera-id').val() || '').trim()

    $.get('/api/cameras?' + ConfVisionUrls.camerasQueryString())
        .done(function (r) {
            const lista = (r.dados || r || []).filter(confVisionCameraApareceAoVivo)
            const sel = $('#sel-camera-live').empty().append('<option value="">Selecione...</option>')
            lista.forEach(function (cam) {
                sel.append('<option value="' + cam.id + '">' + cam.id + ' — ' + (cam.nome || 'Câmera') + '</option>')
            })
            if (preId && sel.find('option[value="' + preId + '"]').length) {
                sel.val(preId)
                ConfVisionLivePlayer.start({
                    video: '#player-hls',
                    status: '#cv-live-status',
                    cameraId: preId
                })
            }
        })
}
