"""Mensagens amigáveis para troubleshooting de publicação RTMP no MediaMTX."""

from __future__ import annotations

# motivo_codigo -> (titulo PT, dica PT, gravidade)
MOTIVOS = {
    "eof_sem_publish": (
        "Conectou, mas não publicou o stream",
        "A câmera abriu a porta 1935 e fechou sem enviar vídeo. "
        "Confira a URL (WIFI sem / no fim; DVR com /), path cam/{hash12} e codec H.264.",
        "error",
    ),
    "path_barra_final": (
        "Path inválido: barra no final",
        "MediaMTX rejeita barra no fim do path. WIFI: sem /; DVR: / só no campo do aparelho.",
        "error",
    ),
    "auth_falhou": (
        "Autenticação falhou",
        "Chave RTMP inválida. Use a URL cam/{hash} gerada no ConfVision (sem ?pass=).",
        "error",
    ),
    "path_invalido": (
        "Nome de path inválido",
        "Use o path cam/{hash12} exatamente como no cadastro ConfVision.",
        "error",
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
