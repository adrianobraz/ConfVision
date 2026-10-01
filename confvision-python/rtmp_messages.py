"""Mensagens amigáveis para troubleshooting de publicação RTMP no MediaMTX."""

from __future__ import annotations

# motivo_codigo -> (titulo PT, dica PT, gravidade)
MOTIVOS = {
    "eof_sem_publish": (
        "Conectou, mas não publicou o stream",
        "A câmera abriu a porta 1935 e fechou sem enviar vídeo. "
        "Confira a URL (Wifi sem / no fim; Cabeado IP com /), path cam/{hash12} e codec H.264.",
        "error",
    ),
    "path_barra_final": (
        "Path inválido: barra no final",
        "MediaMTX rejeita barra no fim do path. Wifi: sem /; Cabeado IP: / só no campo do aparelho.",
        "error",
    ),
    "auth_falhou": (
        "Autenticação falhou",
        "Chave RTMP inválida. Use a URL cam/{hash} gerada no ConfVision (sem ?pass=).",
        "error",
    ),
    "auth_guard_indisponivel": (
        "Guard RTMP indisponível no boot",
        "MediaMTX tentou autenticar antes do Guard na :8100. Reinicie confvision ou aguarde; "
        "não é falha de chave da câmera.",
        "info",
    ),
    "path_invalido": (
        "Nome de path inválido",
        "Use o path cam/{hash12} exatamente como no cadastro ConfVision.",
        "error",
    ),
    "stream_pausado_sistema": (
        "Publicação RTMP pausada pelo sistema",
        "O analítico foi pausado por falha de stream (404/NAL/retry). Corrija o encode ou o path RTMP "
        "e use Reativar stream no painel ConfVision antes de publicar de novo.",
        "warn",
    ),
    "camera_inativa": (
        "Câmera inativa no cadastro",
        "Ative a câmera no painel ou use plano online se for sob demanda.",
        "info",
    ),
    "camera_bloqueada": (
        "Câmera bloqueada (RTMP)",
        "Desbloqueie em Câmeras ou contate o suporte.",
        "warn",
    ),
    "terminated": (
        "Conexão encerrada pelo servidor",
        "Reinício do MediaMTX ou encerramento administrativo. Normal após deploy.",
        "info",
    ),
    "muxer_destroyed": (
        "Stream offline (publisher caiu)",
        "O publicador do path desconectou. Verifique rede/encode da câmera.",
        "warn",
    ),
    "hls_dts_nao_monotono": (
        "HLS: timestamps (DTS) inconsistentes",
        "O encode RTMP envia DTS fora de ordem. HLS pode falhar ou travar. "
        "No DVR: H.264 baseline, GOP fixo, desligue VBR agressivo; evite só-áudio intercalado.",
        "error",
    ),
    "rtmp_video_track_nao_configurado": (
        "Câmera com problema — video track não configurado",
        "O ConfVision pausa o analítico automaticamente quando detecta "
        "'video track not set up' no MediaMTX. Corrija H.264 + AAC no encoder/DVR "
        "e use Reativar stream no painel.",
        "error",
    ),
    "rtmp_so_audio_sem_video": (
        "Stream só áudio (sem frames de vídeo)",
        "O publisher não envia vídeo H.264. Analítico Rust/RTSP pode falhar. "
        "Ative sub-stream de vídeo no NVR/DVR.",
        "error",
    ),
    "ip_banido": (
        "IP bloqueado temporariamente",
        "Muitas falhas RTMP anteriores. Use IPs banidos no painel para desbanir ou aguarde o TTL. "
        "O equipamento continua tentando reconectar — corrija URL/path no DVR.",
        "info",
    ),
    "closed_outro": (
        "Conexão RTMP fechada com erro",
        "Veja o motivo técnico. Em geral é URL, codec ou rede no cliente.",
        "warn",
    ),
    "desconhecido": (
        "Evento de conexão com falha",
        "Consulte o log bruto do MediaMTX para detalhes.",
        "warn",
    ),
}


def classificar_motivo(raw: str) -> str:
    texto = (raw or "").strip().lower()
    if "connection refused" in texto and ":8100" in texto:
        return "auth_guard_indisponivel"
    if "can't end with a slash" in texto or "cant end with a slash" in texto:
        return "path_barra_final"
    if "authentication failed" in texto or "invalid credentials" in texto:
        return "auth_falhou"
    if "invalid path name" in texto:
        return "path_invalido"
    if texto in ("eof", "closed: eof") or texto.endswith(": eof") or "closed: eof" in texto:
        return "eof_sem_publish"
    if "terminated" in texto:
        return "terminated"
    if "muxer" in texto and "destroyed" in texto:
        return "muxer_destroyed"
    if texto.startswith("closed:") or "closed:" in texto:
        if "eof" in texto:
            return "eof_sem_publish"
        return "closed_outro"
    return "desconhecido"


def mensagem_amigavel(codigo: str) -> tuple[str, str, str]:
    return MOTIVOS.get(codigo, MOTIVOS["desconhecido"])
